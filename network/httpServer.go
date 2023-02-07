package network

import (
	"encoding/json"
	"fmt"
	"github.com/roycncn/BUChain/blockchain"
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
	"net/http"
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

type Resp struct {
	Result string `json:"result"`
	Msg    string `json:"msg"`
}

type HTTPServer struct {
	mux    *http.ServeMux
	wg     sync.WaitGroup
	quitCh chan struct{}
}

func NewHTTPServer(cfg *config.Config, pipeSet *blockchain.PipeSet, cacheSet *blockchain.CacheSet) *HTTPServer {

	mux := http.NewServeMux()
	mux.Handle("/", &indexHandler{cacheSet: cacheSet})

	return &HTTPServer{
		mux:    mux,
		quitCh: make(chan struct{}),
	}
}

func (s *HTTPServer) Start() {
	log.Info("HTTP Server start")

	go http.ListenAndServe(":8001", s.mux)

}

func (s *HTTPServer) Stop() {
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
