package blockchain

import (
	"github.com/docker/docker/pkg/pubsub"
	"github.com/patrickmn/go-cache"
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
	"sync"
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

	s.wg.Add(1)
	go s.doSyncBlock()

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
			newblock := msg.(Block)
			log.Infof("New Block %v arrive", newblock.index)
			newblock.isValidNextBlock(s.GetCurrBlock())
			
		}
	}
}

func (s blockServer) GetCurrBlock() *Block {
	height := 0
	block := &Block{}
	if x, found := s.ChainCache.Get("CURR_HEIGHT"); found {
		height = x.(int)
	}

	if x, found := s.ChainCache.Get("HEIGHT_" + string(height)); found {
		block = x.(*Block)
	}
	return block
}

func (s blockServer) GetBlockByHeight(height int) *Block {
	block := &Block{}
	if x, found := s.ChainCache.Get("HEIGHT_" + string(height)); found {
		block = x.(*Block)
	}
	return block
}
