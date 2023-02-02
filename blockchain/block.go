package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Block struct {
	index     int64
	hash      [32]byte
	prevHash  [32]byte
	timestamp int64
	data      string
}

func NewBlock(data string, prevBlock Block) *Block {
	block := &Block{}
	block.index = prevBlock.index + 1
	block.prevHash = prevBlock.hash
	block.timestamp = time.Now().Unix()
	block.data = data
	hash := sha256.Sum256([]byte(strconv.FormatInt(block.index, 10) + hex.EncodeToString(block.prevHash[:]) + strconv.FormatInt(block.timestamp, 10) + block.data))
	block.hash = hash
	return block
}

func NewGenesisBlock() *Block {
	block := &Block{}
	block.index = 0
	block.prevHash = [32]byte{}
	block.timestamp = time.Now().Unix()
	block.data = "ROOT"
	hash := sha256.Sum256([]byte(strconv.FormatInt(block.index, 10) + hex.EncodeToString(block.prevHash[:]) + strconv.FormatInt(block.timestamp, 10) + block.data))
	block.hash = hash
	return block
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

	hash := sha256.Sum256([]byte(strconv.FormatInt(b.index, 10) + hex.EncodeToString(b.prevHash[:]) + strconv.FormatInt(b.timestamp, 10) + b.data))
	if b.hash != hash {
		log.Infof("Wrong Block hash!")
		return false
	}

	return true

}
