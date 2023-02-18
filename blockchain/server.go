package blockchain

import (
	"container/heap"
	"encoding/hex"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
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
	quitCh             chan struct{}
	wg                 sync.WaitGroup
	cfg                *config.Config
	memPool            *TXPriorityQueue
	priv               *secp256k1.PrivateKey
	Pubkey             *secp256k1.PublicKey
	cacheSet           *CacheSet
	SyncBlockPipe      *pubsub.Publisher
	BroadcastBlockPipe *pubsub.Publisher
	GetChainPipe       *pubsub.Publisher
	NewBlockCommitPipe *pubsub.Publisher //For Mempool
	NewTXPipe          *pubsub.Publisher //For Mempool
	MempoolSyncPipe    *pubsub.Publisher //For Mempool

	mempoolLck *sync.Mutex
}

func NewBlockServer(cfg *config.Config, pipeSet *PipeSet, cacheSet *CacheSet) *blockServer {
	privByte, _ := hex.DecodeString(cfg.PrivateKey)
	priv := secp256k1.PrivKeyFromBytes(privByte)
	pubkey := priv.PubKey()
	pq := make(TXPriorityQueue, 0)
	return &blockServer{
		cfg:                cfg,
		quitCh:             make(chan struct{}),
		cacheSet:           cacheSet,
		SyncBlockPipe:      pipeSet.SyncBlockPipe,
		BroadcastBlockPipe: pipeSet.BroadcastBlockPipe,
		NewBlockCommitPipe: pipeSet.NewBlockCommitPipe,
		NewTXPipe:          pipeSet.NewTXPipe,
		MempoolSyncPipe:    pipeSet.MempoolSyncPipe,
		GetChainPipe:       pipeSet.GetChainPipe,
		priv:               priv,
		Pubkey:             pubkey,
		memPool:            &pq,
		mempoolLck:         new(sync.Mutex),
	}
}

func (s blockServer) Start() {
	log.Info("blockServer Start")
	genesisBlock := NewGenesisBlock()
	s.cacheSet.ChainCache.Set("CURR_HEIGHT", genesisBlock.Index, cache.NoExpiration)
	s.cacheSet.ChainCache.Set("HEIGHT_"+strconv.FormatInt(genesisBlock.Index, 10), genesisBlock, cache.NoExpiration)
	heap.Init(s.memPool)
	s.wg.Add(5)
	go s.doRunMemPool()
	go s.doSyncBlock()
	go s.doLocalMining()
	go s.doSyncUXTO()
	go s.doTimerTest()
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
			txs := make([]*tx.Transcation, 0)
			transcation := tx.GetCoinbaseTX(50, s.Pubkey, int(prevBlock.Index+1))
			txs = append(txs, transcation)
			s.mempoolLck.Lock()
			log.Infof("Current mempool length %d", s.memPool.Len())
			for i := 0; i < 10 && s.memPool.Len() > 0; i++ {
				pooledtx := s.memPool.Pop().(*PooledTX)
				txs = append(txs, pooledtx.tx)
			}
			s.mempoolLck.Unlock()
			newBlock := NewBlock(txs, prevBlock, s.GetDifficulty(), s.cacheSet.ChainCache)
			if newBlock != nil && newBlock.isValidNextBlock(s.GetCurrBlock()) {
				s.cacheSet.ChainCache.Set("CURR_HEIGHT", newBlock.Index, cache.NoExpiration)
				s.cacheSet.ChainCache.Set("HEIGHT_"+strconv.FormatInt(newBlock.Index, 10), newBlock, cache.NoExpiration)
				s.BroadcastBlockPipe.Publish(newBlock)
				s.NewBlockCommitPipe.Publish(newBlock)
				log.Infof("New Block %v, %v Added to chain by Local ,cache addr %v", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]), s.cacheSet.ChainCache)
			}

		}
	}

}

func (s blockServer) doSyncUXTO() {
	defer s.wg.Done()
	newBlockCommitPipe := s.NewBlockCommitPipe.Subscribe()
	defer s.NewBlockCommitPipe.Evict(newBlockCommitPipe)

	for {
		select {
		case <-s.quitCh:
			return
		case msg := <-newBlockCommitPipe:
			m := msg.(*Block)
			for _, TX := range m.Transcations {

				for _, txin := range TX.TxIns {
					s.cacheSet.UXTOCache.Delete(txin.TxOutId + "-" + strconv.Itoa(txin.TxOutIndex))
				}

				for i, txout := range TX.TxOut {
					s.cacheSet.UXTOCache.Set(TX.Id+"-"+strconv.Itoa(i),
						hex.EncodeToString(txout.Address)+"-"+strconv.Itoa(txout.Amount), cache.NoExpiration)
				}

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

out:
	for {
		select {
		case <-s.quitCh:
			return
		case msg := <-mempoolSyncPipe:
			s.mempoolLck.Lock()
			m := msg.(*TXPriorityQueue)
			s.memPool = m
			//delete invalid
			s.mempoolLck.Unlock()
		case msg := <-newBlockCommitPipe:
			m := msg.(*Block)
			//If TX in block then remove
			s.mempoolLck.Lock()
			for _, TX := range m.Transcations {
				for _, txInPool := range *s.memPool {
					if txInPool.tx.Id == TX.Id {
						heap.Remove(s.memPool, txInPool.index)
					}
				}
			}
			s.mempoolLck.Unlock()
		case msg := <-newTXPipe:

			m := msg.(*tx.Transcation)
			check, err := tx.CheckUXTOandCheckSign(m, s.cacheSet.UXTOCache)
			if !check {
				log.Errorf("TX check failed %v", err)
				break
			}

			s.mempoolLck.Lock()
			for _, txInPool := range *s.memPool {
				if txInPool.tx.Id == m.Id {
					s.mempoolLck.Unlock()
					//If the newTX already exist, then we don't add in the mempool
					log.Infof("SKIP TX %v ", m.Id)
					break out
				}
			}

			ptx := &PooledTX{
				timestamp: time.Now().UnixNano(),
				tx:        m,
			}

			heap.Push(s.memPool, ptx)
			s.mempoolLck.Unlock()
			log.Infof("A TX %v added to memPOOL @ %v", ptx.tx.Id, ptx.timestamp)
		}
	}

}

func (s blockServer) doTimerTest() {
	defer s.wg.Done()
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-s.quitCh:
			return
		case <-t.C:
			//log.Infof("Timer IS WORKING")
			//balance := make(map[string]int)
			//x := s.cacheSet.UXTOCache.Items()
			//for i, j := range x {
			//fmt.Println(i, j.Object.(string))
			/*			str := j.Object.(string)
						acct := strings.Split(str, "-")

						temp, _ := strconv.Atoi(acct[1])
						balance[acct[0]] += temp*/
			/*
				txout := strings.Split(i, "-")
				temp2, _ := strconv.Atoi(txout[1])

				pubkeybyte, _ := hex.DecodeString("02e90c589d434fe3b70f11bea48bf54922c46646a82d4418d3d9ed16f258a7f88b")
				recvpubkey, _ := secp256k1.ParsePubKey(pubkeybyte)

				if balance[acct[0]] > 150 {
					txIns := []*tx.TxIn{{
						TxOutId:    txout[0],
						TxOutIndex: temp2,
						Sig:        nil,
					}}

					txOuts := []*tx.TxOut{{
						Address: recvpubkey,
						Amount:  50,
					}}
					transcation := &tx.Transcation{
						Id:    "",
						TxIns: txIns,
						TxOut: txOuts,
					}

					transcation.Id = transcation.CalcTxID()
					tx.CheckAndSignTxIn(s.priv, transcation, s.UXTOCache)
					s.NewTXPipe.Publish(transcation)
				}*/
		}
		/*			for x, y := range balance {
					fmt.Println(x, y)
				}*/
		t.Reset(time.Second * 5)
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
			} else if newBlock.Index-curr_block.Index > 1 {
				log.Infof("New Block %v too high, trigger replacechain, hash %v", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
				s.GetChainPipe.Publish("GO")
				time.Sleep(2 * time.Second)
				log.Infof("Current chain cache %v", s.cacheSet.ChainCache)

			} else if newBlock.isValidNextBlock(curr_block) {
				//STOP MINING
				s.cacheSet.ChainCache.Set("MINING_STATUS", 0, cache.NoExpiration)
				s.cacheSet.ChainCache.Set("CURR_HEIGHT", newBlock.Index, cache.NoExpiration)
				s.cacheSet.ChainCache.Set("HEIGHT_"+strconv.FormatInt(newBlock.Index, 10), newBlock, cache.NoExpiration)
				s.NewBlockCommitPipe.Publish(newBlock)
				log.Infof("New Block %v, %v Added to chain from ohters", newBlock.Index, hex.EncodeToString(newBlock.Hash[:]))
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
	if x, found := s.cacheSet.ChainCache.Get("CURR_HEIGHT"); found {
		height = x.(int64)
	}
	key := "HEIGHT_" + strconv.FormatInt(height, 10)
	if x, found := s.cacheSet.ChainCache.Get(key); found {
		block = x.(*Block)
	}
	return block
}

func (s blockServer) GetBlockByHeight(height int64) *Block {
	block := &Block{}
	if x, found := s.cacheSet.ChainCache.Get("HEIGHT_" + strconv.FormatInt(height, 10)); found {
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
