package BUChain

import (
	"github.com/roycncn/BUChain/blockchain"
	"github.com/roycncn/BUChain/config"
	"github.com/roycncn/BUChain/network"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

//PipeSet contain a set of pub/sub pipe for internal go routine communication
var pipeSet = blockchain.NewPipeSet()

//CacheSet contain a set of global routine safe cache for context
var cacheSet = blockchain.NewCacheSet()

func Start(cfg *config.Config) error {

	log.Info("BU Chain Server Starting")

	blockServer := blockchain.NewBlockServer(cfg, pipeSet, cacheSet)
	httpServer := network.NewHTTPServer(cfg, pipeSet, cacheSet)
	httpClient := network.NewHTTPClient(cfg, pipeSet, cacheSet)
	httpServer.Start()
	httpClient.Start()
	blockServer.Start() //non block start up

	waitExit()

	blockServer.Stop()
	httpClient.Stop()
	httpServer.Stop()
	log.Infof("Main Server Stopped")
	return nil
}

func waitExit() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	i := <-c
	log.Infof("Main Server Received interrupt[%v], shutting down...\n", i)
}
