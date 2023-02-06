package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	index      int64
	hash       [32]byte
	prevHash   [32]byte
	timestamp  int64
	data       string
	difficulty int
	nonce      int64
}

func NewBlock(data string, prevBlock *Block, difficulty int) *Block {
	block := &Block{}
	block.index = prevBlock.index + 1
	block.prevHash = prevBlock.hash
	block.timestamp = time.Now().Unix()
	block.data = data
	block.difficulty = difficulty

	block.mineBlock()
	return block
}

func NewGenesisBlock() *Block {
	b := &Block{}
	b.index = 0
	b.prevHash = [32]byte{}
	b.timestamp = time.Now().Unix()
	b.data = "ROOT"
	b.nonce = 0
	b.difficulty = 5
	//b.hash = sha256.Sum256([]byte(strconv.FormatInt(b.index, 10) + hex.EncodeToString(b.prevHash[:]) + strconv.FormatInt(b.timestamp, 10) + b.data + string(b.difficulty) + strconv.FormatInt(b.nonce, 10)))

	b.mineBlock()
	return b
}

func (b *Block) hashMatchDifficulty() bool {

	s := hex.EncodeToString(b.hash[:])
	ok := strings.HasPrefix(s, strings.Repeat("0", b.difficulty))
	if ok {
		return true
	} else {
		return false
	}

}

func (b *Block) mineBlock() {
	b.nonce = 0
	for {
		b.hash = sha256.Sum256([]byte(strconv.FormatInt(b.index, 10) + hex.EncodeToString(b.prevHash[:]) + strconv.FormatInt(b.timestamp, 10) + b.data + strconv.FormatInt(b.nonce, 10)))
		if b.hashMatchDifficulty() {
			break
		}
		b.nonce += 1
	}
	return
}

func (b *Block) isValidNextBlock(prev *Block) bool {
	if prev.index+1 != b.index {
		log.Infof("Wrong Block Index!")
		return false
	}
	if prev.hash != b.prevHash {
		log.Infof("Wrong Block prevhash!")
		return false
	}

	hash := sha256.Sum256([]byte(strconv.FormatInt(b.index, 10) + hex.EncodeToString(b.prevHash[:]) + strconv.FormatInt(b.timestamp, 10) + b.data + strconv.FormatInt(b.nonce, 10)))
	if b.hash != hash {
		log.Infof("Wrong Block hash!")
		return false
	}

	return true

}
