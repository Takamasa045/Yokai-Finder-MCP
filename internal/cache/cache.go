package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

// Cache is a simple in-memory cache for search results
type Cache struct {
	mu      sync.RWMutex
	items   map[string]*CacheItem
	ttl     time.Duration
	maxSize int
}

// CacheItem represents a cached search result
type CacheItem struct {
	Result    *types.YokaiSearchResult
	ExpiresAt time.Time
}

// NewCache creates a new cache with the specified TTL and max size
func NewCache(ttl time.Duration, maxSize int) *Cache {
	c := &Cache{
		items:   make(map[string]*CacheItem),
		ttl:     ttl,
		maxSize: maxSize,
	}

	if ttl > 0 {
		go c.cleanupExpired()
	}

	return c
}

// Get retrieves a cached result if it exists and is not expired
func (c *Cache) Get(params types.YokaiSearchParams) (*types.YokaiSearchResult, bool) {
	if c == nil || c.ttl <= 0 {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateKey(params)
	item, exists := c.items[key]

	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Result, true
}

// Set stores a search result in the cache
func (c *Cache) Set(params types.YokaiSearchParams, result *types.YokaiSearchResult) {
	if c == nil || c.ttl <= 0 || result == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if cache is full
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	key := c.generateKey(params)
	c.items[key] = &CacheItem{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
}

// Size returns the current number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// generateKey creates a cache key from search parameters
func (c *Cache) generateKey(params types.YokaiSearchParams) string {
	data, err := json.Marshal(params)
	if err != nil {
		// Fallback to simple string concatenation
		return fmt.Sprintf("%s:%s:%s:%d", params.Name, params.Region, params.Category, params.Limit)
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// evictOldest removes the oldest item from the cache
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// cleanupExpired periodically removes expired items
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
