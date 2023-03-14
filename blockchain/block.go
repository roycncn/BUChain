package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Index        int64
	Hash         [32]byte
	PrevHash     [32]byte
	Timestamp    int64
	Transactions []*Transaction
	Difficulty   int
	Nonce        int64
}

func NewBlockWithoutControl(txs []*Transaction, prevBlock *Block, difficulty int) *Block {
	block := &Block{}
	block.Index = prevBlock.Index + 1
	block.PrevHash = prevBlock.Hash
	block.Timestamp = time.Now().Unix()
	block.Transactions = txs
	block.Difficulty = difficulty
	block.Transactions = txs

	hasher := sha256.New()
	for _, tx := range block.Transactions {
		hasher.Write([]byte(tx.Id))
	}
	txHash := hex.EncodeToString(hasher.Sum(nil))

	block.MineBlock(txHash)
	return block
}

func NewBlock(txs []*Transaction, prevBlock *Block, difficulty int, minecache *cache.Cache) *Block {
	block := &Block{}
	block.Index = prevBlock.Index + 1
	block.PrevHash = prevBlock.Hash
	block.Timestamp = time.Now().Unix()
	block.Transactions = txs
	block.Difficulty = difficulty
	block.Transactions = txs

	hasher := sha256.New()
	for _, tx := range block.Transactions {
		hasher.Write([]byte(tx.Id))
	}
	txHash := hex.EncodeToString(hasher.Sum(nil))

	if block.MineBlockWithControl(minecache, txHash) {
		return block
	} else {
		return nil
	}

}
func NewGenesisBlockCalculate() *Block {
	b := &Block{}
	b.Index = 0
	b.PrevHash = [32]byte{}
	b.Timestamp = time.Now().Unix()
	b.Nonce = 0
	b.Difficulty = 20
	hasher := sha256.New()
	for _, tx := range b.Transactions {
		hasher.Write([]byte(tx.Id))
	}
	txHash := hex.EncodeToString(hasher.Sum(nil))
	b.MineBlock(txHash)
	return b
}

func NewGenesisBlock() *Block {
	b := &Block{}
	b.Index = 0
	b.PrevHash = [32]byte{}
	b.Timestamp = 1675996673
	b.Nonce = 1884219
	b.Difficulty = 20

	hasher := sha256.New()
	for _, tx := range b.Transactions {
		hasher.Write([]byte(tx.Id))
	}
	txHash := hex.EncodeToString(hasher.Sum(nil))
	b.Hash = sha256.Sum256([]byte(strconv.FormatInt(b.Index, 10) + hex.EncodeToString(b.PrevHash[:]) + strconv.FormatInt(b.Timestamp, 10) + txHash + strconv.FormatInt(b.Nonce, 10)))
	return b
}

func (b *Block) hashMatchDifficulty() bool {
	tmp := ""
	for _, n := range b.Hash[:] {
		s := fmt.Sprintf("%08b", n)
		tmp += s
	}
	ok := strings.HasPrefix(tmp, strings.Repeat("0", b.Difficulty))
	if ok {
		return true
	} else {
		return false
	}

}

func (b *Block) MineBlock(txHash string) {
	b.Nonce = 0

	for {

		b.Hash = sha256.Sum256([]byte(strconv.FormatInt(b.Index, 10) + hex.EncodeToString(b.PrevHash[:]) + strconv.FormatInt(b.Timestamp, 10) + txHash + strconv.FormatInt(b.Nonce, 10)))
		if b.hashMatchDifficulty() {
			break
		}
		b.Nonce += 1
	}
	return
}

func (b *Block) MineBlockWithControl(minecache *cache.Cache, txHash string) bool {
	b.Nonce = 0
	minecache.Set("MINING_STATUS", 1, cache.NoExpiration)
	for {
		mineStatus, _ := minecache.Get("MINING_STATUS")
		if mineStatus != 1 {
			return false
		}
		b.Hash = sha256.Sum256([]byte(strconv.FormatInt(b.Index, 10) + hex.EncodeToString(b.PrevHash[:]) + strconv.FormatInt(b.Timestamp, 10) + txHash + strconv.FormatInt(b.Nonce, 10)))
		if b.hashMatchDifficulty() {
			return true
		}
		b.Nonce += 1
	}

}

func (b *Block) isValidNextBlock(prev *Block) bool {
	if prev.Index+1 != b.Index {
		log.Infof("Wrong Block Index: Curr index %v hash %v; prev index %v hash %v",
			b.Index, hex.EncodeToString(b.Hash[:]), prev.Index, hex.EncodeToString(prev.Hash[:]))
		return false
	}
	if prev.Hash != b.PrevHash {
		log.Infof("Wrong Block prevhash: Curr index %v hash %v; prev index %v hash %v",
			b.Index, hex.EncodeToString(b.Hash[:]), prev.Index, hex.EncodeToString(prev.Hash[:]))
		return false
	}
	hasher := sha256.New()
	for _, tx := range b.Transactions {
		hasher.Write([]byte(tx.Id))
	}
	txHash := hex.EncodeToString(hasher.Sum(nil))

	hash := sha256.Sum256([]byte(strconv.FormatInt(b.Index, 10) + hex.EncodeToString(b.PrevHash[:]) + strconv.FormatInt(b.Timestamp, 10) + txHash + strconv.FormatInt(b.Nonce, 10)))
	if b.Hash != hash {
		log.Infof("Wrong Block Hash!: Curr index %v hash %v; prev index %v hash %v",
			b.Index, hex.EncodeToString(b.Hash[:]), prev.Index, hex.EncodeToString(prev.Hash[:]))
		return false
	}

	return true

}
