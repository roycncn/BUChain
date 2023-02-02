package blockchain

import (
	"fmt"
	"testing"
)

func TestNewBlock(t *testing.T) {
	NGB := NewGenesisBlock()
	fmt.Println(NGB.hash)

}
