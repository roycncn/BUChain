package BUChain

import (
	"github.com/roycncn/BUChain/blockchain"
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func Start(cfg *config.Config) error {

	log.Info("BU Chain Server Starting")
	//PipeSet contain a set of pub/sub pipe for internal go routine communication
	pipeSet := blockchain.NewPipeSet()
	//CacheSet contain a set of global routine safe cache for context
	cacheSet := blockchain.NewCacheSet()

	blockServer := blockchain.NewBlockServer(cfg, pipeSet, cacheSet)

	blockServer.Start() //non block start up

	waitExit()

	blockServer.Stop()

	log.Infof("Main Server Stopped")
	return nil
}

func waitExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	i := <-c
	log.Infof("Main Server Received interrupt[%v], shutting down...\n", i)
}
