package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entries  map[string]cacheEntry
	mutex    sync.Mutex
	interval time.Duration
}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
	}

	cache.reapLoop()

	return cache
}

func (c *Cache) Add(key string, val []byte) {
	// Lock the mutex to ensure thread safety
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a new cache entry with the current time
	entry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

	// Add the entry to the map
	c.entries[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Try to get the entry from the map
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	return entry.val, true
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)

	// This goroutine will run in the background
	go func() {
		for {
			<-ticker.C
			c.reap()
		}
	}()
}

func (c *Cache) reap() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	// Check each entry
	for key, entry := range c.entries {
		// If entry is older than interval, remove it
		if now.Sub(entry.createdAt) > c.interval {
			delete(c.entries, key)
		}
	}
}
