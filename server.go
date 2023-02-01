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
	pipeSet := blockchain.NewPipeSet()
	_ = blockchain.NewCacheSet()

	blockServer := blockchain.NewBlockServer(cfg, pipeSet)

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
