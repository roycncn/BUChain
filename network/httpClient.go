package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
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

	s.wg.Add(2)
	go s.doBroadcastBlock()
	go s.doGetChain()

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

func (s *HTTPClient) doGetChain() {
	defer s.wg.Done()
	getChainPipe := s.pipeSet.GetChainPipe.Subscribe()
	defer s.pipeSet.GetChainPipe.Evict(getChainPipe)

	for {
		select {
		case <-s.quitCh:
			return
		case <-getChainPipe:
			log.Infof("Asking for blocks")
			for _, peer := range s.peers {

				requestURL := fmt.Sprintf("http://127.0.0.1:%v/chain", peer)
				req, _ := http.NewRequest(http.MethodGet, requestURL, nil)
				client := &http.Client{Timeout: 50 * time.Millisecond}
				resp, err := client.Do(req)
				if err != nil {
					log.Errorf("GET CHAIN REQ broadcast to peer %v result :%v", peer, err.Error())
				} else {
					decoder := json.NewDecoder(resp.Body)
					var getBlockResp *RespGetChain
					err := decoder.Decode(&getBlockResp)
					if err != nil {
						log.Errorf("GET CHAIN RESUT error, result :%v", peer, err.Error())
					} else {
						log.Infof("Chain Addr before repace %v", s.cacheSet.ChainCache)
						newCache := cache.New(5*time.Minute, 10*time.Minute)
						s.cacheSet.UXTOCache = cache.New(5*time.Minute, 10*time.Minute)
						for i, j := range getBlockResp.Chain {
							if strings.HasPrefix(i, "CURR_HEIGHT") {
								newCache.Set("CURR_HEIGHT", int64(j.Object.(float64)), cache.NoExpiration)
							} else if strings.HasPrefix(i, "HEIGHT_") {
								block := &blockchain.Block{}
								tmp, _ := json.Marshal(j.Object)
								json.Unmarshal(tmp, block)
								s.pipeSet.NewBlockCommitPipe.Publish(block)
								newCache.Set(i, block, cache.NoExpiration)
							} else if strings.HasPrefix(i, "MINING_STATUS") {
								newCache.Set(i, int(j.Object.(float64)), cache.NoExpiration)
							}
						}

						s.cacheSet.ChainCache = newCache
						log.Infof("Repacing Chain %v", s.cacheSet.ChainCache)
					}
				}

			}

		}
	}
}
