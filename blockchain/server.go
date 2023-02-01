package blockchain

import (
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
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
}

func NewBlockServer(cfg *config.Config, pipeSet *PipeSet) *blockServer {

	return &blockServer{
		quitCh: make(chan struct{}),
	}
}

func (s blockServer) Start() {
	log.Info("blockServer Start")

	s.wg.Add(1)
	go s.doWorker()

}

func (s blockServer) doWorker() {
	defer s.wg.Done()
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-s.quitCh:
			return
		case <-t.C:
			log.Infof("Block Server is Working")

			t.Reset(time.Second * 2)
		}
	}
}

func (s blockServer) Stop() {
	log.Infof("Block Server shutting down...\n")
	close(s.quitCh)
	s.wg.Wait()
}
