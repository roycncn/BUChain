package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Index      int64
	Hash       [32]byte
	PrevHash   [32]byte
	Timestamp  int64
	Data       string
	Difficulty int
	Nonce      int64
}

func NewBlock(data string, prevBlock *Block, difficulty int) *Block {
	block := &Block{}
	block.Index = prevBlock.Index + 1
	block.PrevHash = prevBlock.Hash
	block.Timestamp = time.Now().Unix()
	block.Data = data
	block.Difficulty = difficulty

	block.MineBlock()
	return block
}

func NewGenesisBlock() *Block {
	b := &Block{}
	b.Index = 0
	b.PrevHash = [32]byte{}
	b.Timestamp = time.Now().Unix()
	b.Data = "ROOT"
	b.Nonce = 0
	b.Difficulty = 10
	//b.Hash = sha256.Sum256([]byte(strconv.FormatInt(b.Index, 10) + hex.EncodeToString(b.PrevHash[:]) + strconv.FormatInt(b.Timestamp, 10) + b.Data + string(b.Difficulty) + strconv.FormatInt(b.Nonce, 10)))

	b.MineBlock()
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

func (b *Block) MineBlock() {
	b.Nonce = 0
	for {
		b.Hash = sha256.Sum256([]byte(strconv.FormatInt(b.Index, 10) + hex.EncodeToString(b.PrevHash[:]) + strconv.FormatInt(b.Timestamp, 10) + b.Data + strconv.FormatInt(b.Nonce, 10)))
		if b.hashMatchDifficulty() {
			break
		}
		b.Nonce += 1
	}
	return
}

func (b *Block) isValidNextBlock(prev *Block) bool {
	if prev.Index+1 != b.Index {
		log.Infof("Wrong Block Index!")
		return false
	}
	if prev.Hash != b.PrevHash {
		log.Infof("Wrong Block prevhash!")
		return false
	}

	hash := sha256.Sum256([]byte(strconv.FormatInt(b.Index, 10) + hex.EncodeToString(b.PrevHash[:]) + strconv.FormatInt(b.Timestamp, 10) + b.Data + strconv.FormatInt(b.Nonce, 10)))
	if b.Hash != hash {
		log.Infof("Wrong Block Hash!")
		return false
	}

	return true

}

func (b *Block) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Index      int64
		Hash       [32]byte
		PrevHash   [32]byte
		Timestamp  int64
		Data       string
		Difficulty int
		Nonce      int64
	}{
		Index:      b.Index,
		Hash:       b.Hash,
		PrevHash:   b.PrevHash,
		Timestamp:  b.Timestamp,
		Data:       b.Data,
		Difficulty: b.Difficulty,
		Nonce:      b.Nonce,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}
