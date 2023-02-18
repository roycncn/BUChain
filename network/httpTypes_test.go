package network

import (
	"encoding/json"
	"fmt"
	"github.com/roycncn/BUChain/blockchain"
	"testing"
)

func TestBlockJson(t *testing.T) {

	NGB := blockchain.NewGenesisBlock()

	req := ReqPostBlock{Block: NGB}
	newData, _ := json.Marshal(req)
	fmt.Println(newData)

}
