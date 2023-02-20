package blockchain

import (
	"container/heap"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestTxPQ(t *testing.T) {
	var pooltxs []*PooledTX

	for i := 1; i < 10; i++ {
		tx := &Transcation{
			Id:    "123",
			TxIns: nil,
			TxOut: nil,
		}
		pooltx := &PooledTX{
			timestamp: int64(i),
			tx:        tx,
		}
		pooltxs = append(pooltxs, pooltx)
		time.Sleep(1)
	}
	rand.Shuffle(len(pooltxs), func(i, j int) {
		pooltxs[i], pooltxs[j] = pooltxs[j], pooltxs[i]
	})

	pq := make(TXPriorityQueue, 0)

	heap.Init(&pq)
	for _, pooltx := range pooltxs {
		heap.Push(&pq, pooltx)
	}

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*PooledTX)
		fmt.Printf("%.2d:%s ", item.timestamp, item.tx.Id)
	}

	println("OK")
}
