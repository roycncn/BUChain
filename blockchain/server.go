package blockchain

import (
	"container/heap"
	"encoding/hex"
	"github.com/docker/docker/pkg/pubsub"
	"github.com/patrickmn/go-cache"
	"github.com/roycncn/BUChain/config"
	"github.com/roycncn/BUChain/tx"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

const (
	SYNCING = 1
	OUTSYNC = -1
	IDLE    = 0
)

type blockServer struct {
	quitCh  chan struct{}
	wg      sync.WaitGroup
	cfg     *config.Config
	memPool TXPriorityQueue

	ChainCache         *cache.Cache
	SyncBlockPipe      *pubsub.Publisher
	BroadcastBlockPipe *pubsub.Publisher
	NewBlockCommitPipe *pubsub.Publisher //For Mempool
	NewTXPipe          *pubsub.Publisher //For Mempool
	MempoolSyncPipe    *pubsub.Publisher //For Mempool

	mempoolLck *sync.Mutex
}

func NewBlockServer(cfg *config.Config, pipeSet *PipeSet, cacheSet *CacheSet) *blockServer {

	return &blockServer{
		cfg:                cfg,
		quitCh:             make(chan struct{}),
		ChainCache:         cacheSet.ChainCache,
		SyncBlockPipe:      pipeSet.SyncBlockPipe,
		BroadcastBlockPipe: pipeSet.BroadcastBlockPipe,
		memPool:            make(TXPriorityQueue, 0),
	}
}

func (s blockServer) Start() {
	log.Info("blockServer Start")
	genesisBlock := NewGenesisBlock()
	s.ChainCache.Set("CURR_HEIGHT", genesisBlock.Index, cache.NoExpiration)
	s.ChainCache.Set("HEIGHT_"+strconv.FormatInt(genesisBlock.Index, 10), genesisBlock, cache.NoExpiration)
	heap.Init(&s.memPool)
	s.wg.Add(3)
	go s.doRunMemPool()
	go s.doSyncBlock()
	go s.doLocalMining()

}

func (s blockServer) Stop() {
	log.Infof("Block Server shutting down...\n")
	close(s.quitCh)
	s.wg.Wait()
}

func (s blockServer) doLocalMining() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quitCh:
			return
		default:
			prevBlock := s.GetCurrBlock()
			newBlock := NewBlock("TEST"+string(s.cfg.RestPort), prevBlock, s.GetDifficulty(), s.ChainCache)
			if newBlock != nil && newBlock.isValidNextBlock(s.GetCurrBlock()) {
				s.ChainCache.Set("CURR_HEIGHT", newBlock.Index, cache.NoExpiration)
				s.ChainCache.Set("HEIGHT_"+strconv.FormatInt(newBlock.Index, 10), newBlock, cache.NoExpiration)
				s.BroadcastBlockPipe.Publish(newBlock)
				s.NewBlockCommitPipe.Publish(newBlock)
				log.Infof("New Block %v, %v Added to chain by Local", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
			}

		}
	}

}
func (s blockServer) doRunMemPool() {
	defer s.wg.Done()

	mempoolSyncPipe := s.MempoolSyncPipe.Subscribe()
	newBlockCommitPipe := s.NewBlockCommitPipe.Subscribe()
	newTXPipe := s.NewTXPipe.Subscribe()

	defer s.MempoolSyncPipe.Evict(mempoolSyncPipe)
	defer s.NewBlockCommitPipe.Evict(newBlockCommitPipe)
	defer s.NewTXPipe.Evict(newTXPipe)

	for {
		select {
		case <-s.quitCh:
			return
		case msg := <-mempoolSyncPipe:
			s.mempoolLck.Lock()
			m := msg.(TXPriorityQueue)
			s.memPool = m
			//delete invalid
			s.mempoolLck.Unlock()
		case msg := <-newBlockCommitPipe:
			m := msg.(*Block)
			//If TX in block then remove
			s.mempoolLck.Lock()
			for _, TX := range m.Transcations {
				for _, txInPool := range s.memPool {
					if txInPool.tx.Id == TX.Id {
						heap.Remove(&s.memPool, txInPool.index)
					}
				}
			}
			s.mempoolLck.Unlock()
		case msg := <-newTXPipe:
			s.mempoolLck.Lock()
			m := msg.(*tx.Transcation)
			ptx := &PooledTX{
				timestamp: time.Now().UnixNano(),
				tx:        m,
			}
			heap.Push(&s.memPool, ptx)
			s.mempoolLck.Unlock()
		}
	}

}

func (s blockServer) doSyncBlock() {
	defer s.wg.Done()
	syncBlockPipe := s.SyncBlockPipe.Subscribe()
	defer s.SyncBlockPipe.Evict(syncBlockPipe)

	for {
		select {
		case <-s.quitCh:
			return
		case msg := <-syncBlockPipe:
			newBlock := msg.(*Block)
			log.Infof("New Block %v arrive from others. Hash %v ", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
			curr_block := s.GetCurrBlock()
			if newBlock.Index-curr_block.Index <= 0 {
				log.Infof("Same Height Block or early Height %v , hash %v", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
			} else if newBlock.isValidNextBlock(curr_block) {
				//STOP MINING
				s.ChainCache.Set("MINING_STATUS", 0, cache.NoExpiration)
				s.ChainCache.Set("CURR_HEIGHT", newBlock.Index, cache.NoExpiration)
				s.ChainCache.Set("HEIGHT_"+strconv.FormatInt(newBlock.Index, 10), newBlock, cache.NoExpiration)
				s.NewBlockCommitPipe.Publish(newBlock)
				log.Infof("New Block %v, %v Added to chain from ohters", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
			} else if newBlock.Index-curr_block.Index > 1 {
				log.Infof("New Block %v too high ignore by now, hash %v", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
			} else {
				//Same Block or Early blocks arrived. just ignore.
				log.Infof("Strange New Block %v ignore, hash %v", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
			}
		}
	}
}

func (s blockServer) GetCurrBlock() *Block {
	var height int64 = -1
	block := &Block{}
	if x, found := s.ChainCache.Get("CURR_HEIGHT"); found {
		height = x.(int64)
	}
	key := "HEIGHT_" + strconv.FormatInt(height, 10)
	if x, found := s.ChainCache.Get(key); found {
		block = x.(*Block)
	}
	return block
}

func (s blockServer) GetBlockByHeight(height int64) *Block {
	block := &Block{}
	if x, found := s.ChainCache.Get("HEIGHT_" + strconv.FormatInt(height, 10)); found {
		block = x.(*Block)
	}
	return block
}

func (s blockServer) GetDifficulty() int {
	latestBlock := s.GetCurrBlock()
	if latestBlock.Index%10 == 0 && latestBlock.Index != 0 {
		return s.GetAdjustedDifficulty()
	} else {
		return latestBlock.Difficulty
	}

}

func (s blockServer) GetAdjustedDifficulty() int {
	latestBlock := s.GetCurrBlock()
	lastAdjustBlock := s.GetBlockByHeight(latestBlock.Index - 9)
	timeExpected := int64(30)
	timeTaken := latestBlock.Timestamp - lastAdjustBlock.Timestamp
	if timeTaken < timeExpected/2 {
		log.Infof("Increase Difficulty to %v !", lastAdjustBlock.Difficulty+1)
		return lastAdjustBlock.Difficulty + 1
	} else if timeTaken > timeExpected*2 {
		log.Infof("Decrease Difficulty to %v !", lastAdjustBlock.Difficulty-1)
		return lastAdjustBlock.Difficulty - 1
	} else {
		log.Infof("Maintain Difficulty at %v !", lastAdjustBlock.Difficulty)
		return lastAdjustBlock.Difficulty
	}
}
