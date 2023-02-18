package network

import (
	"github.com/patrickmn/go-cache"
	"github.com/roycncn/BUChain/blockchain"
)

type Resp struct {
	Result string `json:"result"`
	Msg    string `json:"msg"`
}

type ReqPostBlock struct {
	Block *blockchain.Block `json:"block"`
}

type RespGetBlock struct {
	Result string            `json:"result"`
	Block  *blockchain.Block `json:"block"`
}

type RespGetChain struct {
	Result string                `json:"result"`
	Chain  map[string]cache.Item `json:"chain"`
}

type ReqGetWallet struct {
	Addr string `json:"Addr"`
}

type ReqPostWallet struct {
	FromAddr string `json:"FromAddr"`
	ToAddr   string `json:"ToAddr"`
	Amount   int    `json:"Amount"`
	Private  string `json:"PrivKey"`
}
