package blockchain

import (
	"container/heap"
	"github.com/docker/docker/pkg/pubsub"
	"github.com/patrickmn/go-cache"
	"time"
)

type PipeSet struct {
	SyncBlockPipe      *pubsub.Publisher
	BroadcastBlockPipe *pubsub.Publisher
	BroadcastTxPipe    *pubsub.Publisher
	GetChainPipe       *pubsub.Publisher
	NewBlockCommitPipe *pubsub.Publisher //For Mempool
	NewTXPipe          *pubsub.Publisher //For Mempool

}

type CacheSet struct {
	ChainCache *cache.Cache
	UTXOCache  *cache.Cache
}

func NewPipeSet() *PipeSet {
	return &PipeSet{
		GetChainPipe:       pubsub.NewPublisher(100*time.Millisecond, 100),
		SyncBlockPipe:      pubsub.NewPublisher(100*time.Millisecond, 100),
		BroadcastBlockPipe: pubsub.NewPublisher(100*time.Millisecond, 100),
		BroadcastTxPipe:    pubsub.NewPublisher(100*time.Millisecond, 100),
		NewBlockCommitPipe: pubsub.NewPublisher(100*time.Millisecond, 100),
		NewTXPipe:          pubsub.NewPublisher(100*time.Millisecond, 100),
	}
}

func NewCacheSet() *CacheSet {
	return &CacheSet{
		ChainCache: cache.New(5*time.Minute, 10*time.Minute), //缓存数据
		UTXOCache:  cache.New(5*time.Minute, 10*time.Minute), //缓存数据

	}
}

type PooledTX struct {
	timestamp int64
	tx        *Transaction
	index     int
}

type TXPriorityQueue []*PooledTX

func (pq TXPriorityQueue) Len() int { return len(pq) }

func (pq TXPriorityQueue) Less(i, j int) bool {
	return pq[i].timestamp < pq[j].timestamp
}

func (pq TXPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *TXPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PooledTX)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *TXPriorityQueue) Peek() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	return item
}

func (pq *TXPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *TXPriorityQueue) update(item *PooledTX, tx *Transaction, timestamp int64) {
	item.tx = tx
	item.timestamp = timestamp
	heap.Fix(pq, item.index)
}
