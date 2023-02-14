package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/roycncn/BUChain/tx"
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
	fmt.Println(NGB.Nonce)
	fmt.Println(NGB.Hash)

}

func TestTime(t *testing.T) {
	a := time.Now().Unix()
	time.Sleep(5 * time.Second)
	b := time.Now().Unix()

	println(a)
	println(b)

}

func TestGenPriKey(t *testing.T) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return
	}

	fmt.Println(hex.EncodeToString(key.Serialize()))
	fmt.Println(hex.EncodeToString(key.PubKey().SerializeCompressed()))

	privKeyBytes, _ := hex.DecodeString(hex.EncodeToString(key.Serialize()))
	priv := secp256k1.PrivKeyFromBytes(privKeyBytes)
	fmt.Println(hex.EncodeToString(priv.Serialize()))
	fmt.Println(hex.EncodeToString(priv.PubKey().SerializeCompressed()))
}

func TestCoinbase(t *testing.T) {
	key, _ := secp256k1.GeneratePrivateKey()
	tx := tx.GetCoinbaseTX(50, key.PubKey())
	x, _ := json.Marshal(tx)
	fmt.Println(x)
}
