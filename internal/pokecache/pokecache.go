package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val []byte
}

type Cache struct {
	entries map[string]cacheEntry
	mu *sync.Mutex
	timeout time.Duration
}

func NewCache(timeout time.Duration) *Cache{
	c := Cache{
		entries: make(map[string]cacheEntry),
		mu: &sync.Mutex{},
		timeout: timeout,
	}
	ticker := time.NewTicker(timeout)
	go func(){
		for range ticker.C {
			c.reapLoop()
		}
	}()
	return &c
}

func (c *Cache) Add(key string, val []byte){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val: val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool){
	if entiti, ok := c.entries[key]; ok {
		return entiti.val, true
	}
	return nil, false
}

func (c *Cache) reapLoop(){
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, entry := range c.entries {
		if time.Since(entry.createdAt) > c.timeout {
			delete(c.entries, key)
		}
	}
}