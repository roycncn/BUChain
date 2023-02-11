package network

import (
	"encoding/json"
	"fmt"
	"github.com/roycncn/BUChain/blockchain"
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

const RESP_SUCCESS = "Success"
const RESP_FAILED = "Failed"

func byte32(s []byte) (a *[32]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}

type HTTPServer struct {
	mux    *http.ServeMux
	wg     sync.WaitGroup
	cfg    *config.Config
	quitCh chan struct{}
}

func NewHTTPServer(cfg *config.Config, pipeSet *blockchain.PipeSet, cacheSet *blockchain.CacheSet) *HTTPServer {

	mux := http.NewServeMux()
	mux.Handle("/", &indexHandler{cacheSet: cacheSet})
	mux.Handle("/block", &blockHandler{cacheSet: cacheSet, pipeSet: pipeSet})
	return &HTTPServer{
		mux:    mux,
		cfg:    cfg,
		quitCh: make(chan struct{}),
	}
}

func (s *HTTPServer) Start() {
	log.Info("HTTP Server start : %v", s.cfg.RestPort)

	go http.ListenAndServe(":"+s.cfg.RestPort, s.mux)

}

func (s *HTTPServer) Stop() {
	log.Info("HTTP Server Stop")
	close(s.quitCh)
	s.wg.Wait()
}

type indexHandler struct {
	cacheSet *blockchain.CacheSet
}

func (h *indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		height, _ := h.cacheSet.ChainCache.Get("CURR_HEIGHT")
		data := &Resp{Result: RESP_SUCCESS, Msg: fmt.Sprintf("%v : Server is working at Height %v", time.Now().String(), height)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Errorf("Error happened in JSON marshal. Err: %s", err)
		}
	default:
		data := &Resp{Result: RESP_FAILED, Msg: "Not Allowed"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Errorf("Error happened in JSON marshal. Err: %s", err)
		}

	}

}

type blockHandler struct {
	cacheSet *blockchain.CacheSet
	pipeSet  *blockchain.PipeSet
}

func (h *blockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var newblockReq *ReqPostBlock
		err := decoder.Decode(&newblockReq)
		if err != nil {
			data := &Resp{Result: RESP_FAILED, Msg: err.Error()}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf("Error happened in Block Unmarshal. Err: %s", err)
			err := json.NewEncoder(w).Encode(data)
			if err != nil {
				return
			}
		} else {
			//Publish New Block to Chain without
			h.pipeSet.SyncBlockPipe.Publish(newblockReq.Block)
			data := &Resp{Result: RESP_SUCCESS, Msg: fmt.Sprintf("Block %v Received", newblockReq.Block.Index)}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		}
	case "GET":
		index, _ := strconv.ParseInt(r.URL.Query().Get("index"), 10, 64)
		curr_index, _ := h.cacheSet.ChainCache.Get("CURR_HEIGHT")
		curr_index_i64 := curr_index.(int64)
		if index > curr_index_i64 {
			//Ask block that we don't have, Failed
			data := &Resp{Result: RESP_FAILED, Msg: fmt.Sprintf("We don't have such block yet, we are at %v", curr_index_i64)}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		} else {

			x, _ := h.cacheSet.ChainCache.Get("HEIGHT_" + strconv.FormatInt(curr_index_i64, 10))
			block := x.(*blockchain.Block)
			data := &RespGetBlock{Result: RESP_SUCCESS, Block: block}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)

		}

	default:
		data := &Resp{Result: RESP_FAILED, Msg: "Not Allowed"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Errorf("Error happened in JSON marshal. Err: %s", err)
		}

	}
}
