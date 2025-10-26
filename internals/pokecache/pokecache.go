package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	val       []byte
	createdAt time.Time
}

type Cache struct {
	entries map[string]cacheEntry
	mu      sync.RWMutex
}

func NewCache(interval time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		mu:      sync.RWMutex{},
	}
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		val:       val,
		createdAt: time.Now(),
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}
	return entry.val, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for range ticker.C {
		for k, v := range c.entries {
			if time.Since(v.createdAt) > interval {
				c.Delete(k)
			}
		}
	}
}

func (c *Cache) StartReapLoop(interval time.Duration) {
	go c.reapLoop(interval)
}
