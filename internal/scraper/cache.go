package scraper

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/keircn/karu/internal/config"
)

type CacheEntry struct {
	Data      any
	ExpiresAt time.Time
}

type Cache struct {
	entries map[string]CacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	cache := &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}

	go cache.cleanup()
	return cache
}
func (c *Cache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(c.entries, key)
		return nil, false
	}

	return entry.Data, true
}

func (c *Cache) Set(key string, data any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) cleanup() {
	cleanupInterval := min(c.ttl, 5*time.Minute)

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mutex.Unlock()
	}
}
func generateCacheKey(query string, vars map[string]any) string {
	data := fmt.Sprintf("%s:%v", query, vars)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

var (
	searchCache  *Cache
	episodeCache *Cache
	cacheOnce    sync.Once
)

func initCaches() {
	cacheOnce.Do(func() {
		cfg, _ := config.Load()

		searchTTL := time.Duration(cfg.CacheTTL) * time.Minute
		if searchTTL <= 0 {
			searchTTL = 15 * time.Minute
		}

		episodeTTL := time.Duration(cfg.CacheTTL*2) * time.Minute
		if episodeTTL <= 0 {
			episodeTTL = 30 * time.Minute
		}

		searchCache = NewCache(searchTTL)
		episodeCache = NewCache(episodeTTL)
	})
}
