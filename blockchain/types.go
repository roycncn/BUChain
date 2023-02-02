package blockchain

import (
	"github.com/docker/docker/pkg/pubsub"
	"github.com/patrickmn/go-cache"
	"time"
)

type PipeSet struct {
	SyncBlockPipe *pubsub.Publisher
}

type CacheSet struct {
	ChainCache *cache.Cache
}

func NewPipeSet() *PipeSet {
	return &PipeSet{
		SyncBlockPipe: pubsub.NewPublisher(100*time.Millisecond, 100),
	}
}

func NewCacheSet() *CacheSet {
	return &CacheSet{
		ChainCache: cache.New(5*time.Minute, 10*time.Minute), //缓存数据

	}
}
