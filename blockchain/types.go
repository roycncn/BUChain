package blockchain

import (
	"github.com/docker/docker/pkg/pubsub"
	"github.com/patrickmn/go-cache"
	"time"
)

type PipeSet struct {
	MsgPeerBrocastPipe     *pubsub.Publisher
	NewBlockBrocastPipe    *pubsub.Publisher
	NewReceiptsBrocastPipe *pubsub.Publisher
	NewTxsBrocastPipe      *pubsub.Publisher
}

type CacheSet struct {
	UniCache    *cache.Cache
	WorkerCache *cache.Cache
	ReqCache    *cache.Cache
	RetryCache  *cache.Cache
}

func NewPipeSet() *PipeSet {
	return &PipeSet{
		MsgPeerBrocastPipe:     pubsub.NewPublisher(100*time.Millisecond, 100),
		NewBlockBrocastPipe:    pubsub.NewPublisher(100*time.Millisecond, 100),
		NewReceiptsBrocastPipe: pubsub.NewPublisher(100*time.Millisecond, 100),
		NewTxsBrocastPipe:      pubsub.NewPublisher(200*time.Millisecond, 1000),
	}
}

func NewCacheSet() *CacheSet {
	return &CacheSet{
		UniCache:    cache.New(5*time.Minute, 10*time.Minute), //缓存数据
		ReqCache:    cache.New(20*time.Second, 1*time.Minute), //跟踪发出的请求id,获取bodies时需要用
		WorkerCache: cache.New(5*time.Minute, 5*time.Minute),  //跟踪最近五分钟中继过tx的节点
		RetryCache:  cache.New(5*time.Minute, 10*time.Minute), //跟踪需要重试的节点,0需要重连
	}
}
