package network

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/roycncn/BUChain/blockchain"
	"github.com/roycncn/BUChain/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
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
	mux.Handle("/chain", &chainHandler{cacheSet: cacheSet, pipeSet: pipeSet})
	mux.Handle("/tx", &txHandler{cacheSet: cacheSet, pipeSet: pipeSet})
	//User Side
	mux.Handle("/wallet", &walletHandler{cacheSet: cacheSet, pipeSet: pipeSet})
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

type chainHandler struct {
	cacheSet *blockchain.CacheSet
	pipeSet  *blockchain.PipeSet
}

func (h *chainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		x := h.cacheSet.ChainCache.Items()
		data := &RespGetChain{Result: RESP_SUCCESS, Chain: x}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)

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

type walletHandler struct {
	cacheSet *blockchain.CacheSet
	pipeSet  *blockchain.PipeSet
}

func (h *walletHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var postWallet *ReqPostWallet
		err := decoder.Decode(&postWallet)
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
			var data *Resp
			balance := make(map[string]int)
			record := make(map[string][]string)
			x := h.cacheSet.UXTOCache.Items()
			for i, j := range x {
				str := j.Object.(string)
				acct := strings.Split(str, "-")
				temp, _ := strconv.Atoi(acct[1])
				balance[acct[0]] += temp
				record[postWallet.FromAddr] = append(record[postWallet.FromAddr], i+", Amount"+acct[1])

			}
			//toAddrByte, _ := hex.DecodeString(postWallet.ToAddr)
			privKeyBytes, _ := hex.DecodeString(postWallet.Private)
			priv := secp256k1.PrivKeyFromBytes(privKeyBytes)
			if hex.EncodeToString(priv.PubKey().SerializeCompressed()) != postWallet.FromAddr {
				data = &Resp{
					Result: RESP_FAILED,
					Msg:    fmt.Sprintf("Given Wrong PK for Address: %v ", postWallet.FromAddr)}
			} else if balance[postWallet.FromAddr] < postWallet.Amount {
				data = &Resp{
					Result: RESP_FAILED,
					Msg: fmt.Sprintf("Address: %v Balance %d. Not enought to pay Amount %d",
						postWallet.FromAddr, balance[postWallet.FromAddr], postWallet.Amount)}
			} else {
				txi, txo, _ := blockchain.GenerateUXTO(postWallet.FromAddr, postWallet.ToAddr, postWallet.Amount, h.cacheSet.UXTOCache)
				Transaction := &blockchain.Transaction{
					Id:    "",
					TxIns: txi,
					TxOut: txo,
				}

				Transaction.Id = Transaction.CalcTxID()
				blockchain.CheckAndSignTxIn(priv, Transaction, h.cacheSet.UXTOCache)
				//BroadCast TX to peers
				h.pipeSet.BroadcastTxPipe.Publish(Transaction)
				h.pipeSet.NewTXPipe.Publish(Transaction)
				data = &Resp{
					Result: RESP_SUCCESS,
					Msg:    Transaction.Id}

			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		}
	case "GET":
		addr := r.URL.Query().Get("addr")
		balance := make(map[string]int)
		record := make(map[string][]string)
		x := h.cacheSet.UXTOCache.Items()
		for i, j := range x {
			str := j.Object.(string)
			acct := strings.Split(str, "-")
			temp, _ := strconv.Atoi(acct[1])
			balance[acct[0]] += temp
			record[acct[0]] = append(record[acct[0]], i+", Amount"+acct[1])

		}

		data := &Resp{Result: RESP_SUCCESS, Msg: fmt.Sprintf("Address: %v Balance %d. Useable vin %v", addr, balance[addr], record[addr])}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)

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

type txHandler struct {
	cacheSet *blockchain.CacheSet
	pipeSet  *blockchain.PipeSet
}

func (h *txHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var postTX *ReqPostTX
		err := decoder.Decode(&postTX)
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
			h.pipeSet.NewTXPipe.Publish(postTX.Tx)
			data := &Resp{Result: RESP_SUCCESS, Msg: "OK"}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		}
	case "GET":
		txId := r.URL.Query().Get("txId")
		x := h.cacheSet.ChainCache.Items()
		height := ""
		var txResult *blockchain.Transaction
		for i, j := range x {
			if strings.HasPrefix(i, "HEIGHT_") {
				block := j.Object.(*blockchain.Block)
				for _, txTmp := range block.Transactions {
					if txTmp.Id == txId {
						height = i
						txResult = txTmp
					}
				}
			}
		}

		if txResult != nil {

			data := &RespGetTX{Result: RESP_SUCCESS, Msg: fmt.Sprintf("Tx @ %v", height), Tx: txResult}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		} else {
			data := &Resp{Result: RESP_FAILED, Msg: fmt.Sprintf("Can't find txId %v ", txId)}
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
