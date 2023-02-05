package blockchain

import (
	"github.com/docker/docker/pkg/pubsub"
	"github.com/patrickmn/go-cache"
	"github.com/roycncn/BUChain/config"
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
	quitCh chan struct{}
	wg     sync.WaitGroup

	ChainCache    *cache.Cache
	SyncBlockPipe *pubsub.Publisher
}

func NewBlockServer(cfg *config.Config, pipeSet *PipeSet, cacheSet *CacheSet) *blockServer {

	return &blockServer{
		quitCh:        make(chan struct{}),
		ChainCache:    cacheSet.ChainCache,
		SyncBlockPipe: pipeSet.SyncBlockPipe,
	}
}

func (s blockServer) Start() {
	log.Info("blockServer Start")
	genesisBlock := NewGenesisBlock()
	s.ChainCache.Set("CURR_HEIGHT", genesisBlock.index, cache.NoExpiration)
	s.ChainCache.Set("HEIGHT_"+strconv.FormatInt(genesisBlock.index, 10), genesisBlock, cache.NoExpiration)

	s.wg.Add(2)
	go s.doSyncBlock()
	go s.doAddBlockbyTimer()

}

func (s blockServer) Stop() {
	log.Infof("Block Server shutting down...\n")
	close(s.quitCh)
	s.wg.Wait()
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
			newblock := msg.(*Block)
			log.Infof("New Block %v arrive", newblock.index)
			newblock.isValidNextBlock(s.GetCurrBlock())
			s.ChainCache.Set("CURR_HEIGHT", newblock.index, cache.NoExpiration)
			s.ChainCache.Set("HEIGHT_"+strconv.FormatInt(newblock.index, 10), newblock, cache.NoExpiration)
		}
	}
}

func (s blockServer) doAddBlockbyTimer() {
	defer s.wg.Done()
	t := time.NewTimer(0)
	defer t.Stop()

	for {
		select {
		case <-s.quitCh:
			return
		case <-t.C:
			t.Reset(time.Second * 5)
			prevBlock := s.GetCurrBlock()
			newBlock := NewBlock("TEST", prevBlock)
			log.Infof("New Block %v produced", newBlock.index)
			s.SyncBlockPipe.Publish(newBlock)

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
