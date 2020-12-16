package hzycache

import (
	"sync"

	"hzycache/lru"
)

// cache is a wrapper around an *lru.Cache that adds synchronization,
// makes values always be ByteView, and counts the size of all keys and
// values.
type cache struct {
	mu         sync.RWMutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return ByteView{}, false
}
