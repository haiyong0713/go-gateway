package model

import (
	"encoding/json"
	"sync"

	"github.com/golang/groupcache/lru"
)

type LRUCache struct {
	bucketSize int64
	bucket     []*cache
}

type cache struct {
	mu         sync.RWMutex
	groupcache *lru.Cache
}

func New(bucketSize int64, maxEntries int) *LRUCache {
	res := &LRUCache{
		bucketSize: bucketSize,
		bucket:     make([]*cache, 0, bucketSize),
	}
	for i := int64(0); i < bucketSize; i++ {
		res.bucket = append(res.bucket, &cache{
			groupcache: lru.New(maxEntries),
		})
	}
	return res
}

func (c *LRUCache) hitBucket(key int64) int64 {
	return key % c.bucketSize
}

func (c *LRUCache) Add(key int64, value json.RawMessage) {
	bucket := c.bucket[c.hitBucket(key)]
	if bucket == nil {
		return
	}
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	bucket.groupcache.Add(key, value)
}

func (c *LRUCache) Get(key int64) (value json.RawMessage, ok bool) {
	var res interface{}
	bucket := c.bucket[c.hitBucket(key)]
	if bucket == nil {
		return
	}
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	if res, ok = bucket.groupcache.Get(key); !ok {
		return
	}
	if value, ok = res.(json.RawMessage); ok {
		return
	}
	return
}
