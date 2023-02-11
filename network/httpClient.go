package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/roycncn/BUChain/blockchain"
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HTTPClient struct {
	wg     sync.WaitGroup
	cfg    *config.Config
	quitCh chan struct{}

	peers    []string
	pipeSet  *blockchain.PipeSet
	cacheSet *blockchain.CacheSet
}

func NewHTTPClient(cfg *config.Config, pipeSet *blockchain.PipeSet, cacheSet *blockchain.CacheSet) *HTTPClient {

	return &HTTPClient{
		cfg:      cfg,
		quitCh:   make(chan struct{}),
		pipeSet:  pipeSet,
		cacheSet: cacheSet,
		peers:    strings.Split(cfg.PeersList, ","),
	}
}

func (s *HTTPClient) Start() {
	log.Info("HTTP Client Start")

	s.wg.Add(1)
	go s.doBroadcastBlock()

}

func (s *HTTPClient) Stop() {
	log.Info("HTTP Client Stop")
	close(s.quitCh)
	s.wg.Wait()
}

func (s *HTTPClient) doBroadcastBlock() {
	defer s.wg.Done()
	broadcastBlockPipe := s.pipeSet.BroadcastBlockPipe.Subscribe()
	defer s.pipeSet.BroadcastBlockPipe.Evict(broadcastBlockPipe)

	for {
		select {
		case <-s.quitCh:
			return
		case msg := <-broadcastBlockPipe:
			newBlock := msg.(*blockchain.Block)
			log.Infof("New Block %v broadcast", newBlock.Index)
			for _, peer := range s.peers {
				reqBody, _ := json.Marshal(ReqPostBlock{Block: newBlock})
				bodyReader := bytes.NewReader(reqBody)

				requestURL := fmt.Sprintf("http://127.0.0.1:%v/block", peer)
				req, _ := http.NewRequest(http.MethodPost, requestURL, bodyReader)
				client := &http.Client{Timeout: 50 * time.Millisecond}
				resp, err := client.Do(req)
				if err != nil {
					log.Infof("New Block %v broadcast to peer %v Err :%v", newBlock.Index, peer, err.Error())
				} else {
					log.Infof("New Block %v broadcast to peer %v result :%v", newBlock.Index, peer, resp.StatusCode)
				}

			}

		}
	}
}
