package blockchain

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func TestNewBlock(t *testing.T) {
	NGB := NewGenesisBlock()
	fmt.Println(NGB.hash)

}

func TestCount(t *testing.T) {

	count := strings.Count("000000123456", "0")
	fmt.Println(count)

}

func TestMineGenesisBlock(t *testing.T) {
	NGB := NewGenesisBlock()
	NGB.mineBlock()
	fmt.Println(NGB.nonce)
	fmt.Println(hex.EncodeToString(NGB.hash[:]))

}
