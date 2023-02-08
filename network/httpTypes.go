package network

import "github.com/roycncn/BUChain/blockchain"

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
