package main

import (
	"sync"

	"github.com/hey-kong/mayflycache/lru"
)

// SafeCache is for concurrency control of LRU cache.
type SafeCache struct {
	maxBytes int64
	mu       sync.Mutex
	lru      *lru.LRUCache
}

// Get locks and unlocks when the it exits to ensure concurrency security.
func (c *SafeCache) Get(key string) (value Chunk, done bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		value, done = v.(Chunk), ok
	}
	return
}

// Set locks and unlocks when the it exits to ensure concurrency security.
func (c *SafeCache) Set(key string, value Chunk) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.NewLRUCache(c.maxBytes, nil)
	}
	c.lru.Set(key, value)
}
