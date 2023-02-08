package blockchain

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewBlock(t *testing.T) {
	NGB := NewGenesisBlock()
	fmt.Println(NGB.Hash)

}

func TestCount(t *testing.T) {

	count := strings.Count("000000123456", "0")
	fmt.Println(count)

}

func TestMineGenesisBlock(t *testing.T) {
	NGB := NewGenesisBlock()
	NGB.MineBlock()
	fmt.Println(NGB.Nonce)
	fmt.Println(hex.EncodeToString(NGB.Hash[:]))

}

func TestTime(t *testing.T) {
	a := time.Now().Unix()
	time.Sleep(5 * time.Second)
	b := time.Now().Unix()

	println(a)
	println(b)

}
