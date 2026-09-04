package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
)

// Cache is a simple in-memory cache for search results.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]*CacheItem
	ttl     time.Duration
	maxSize int
	stopCh  chan struct{}
	stop    sync.Once
}

// CacheItem represents a cached search result.
type CacheItem struct {
	Result    *types.YokaiSearchResult
	ExpiresAt time.Time
	CreatedAt time.Time
}

// NewCache creates a new cache with the specified TTL and max size.
func NewCache(ttl time.Duration, maxSize int) *Cache {
	if maxSize <= 0 {
		maxSize = 1
	}
	c := &Cache{
		items:   make(map[string]*CacheItem),
		ttl:     ttl,
		maxSize: maxSize,
		stopCh:  make(chan struct{}),
	}

	if ttl > 0 {
		go c.cleanupExpired()
	}

	return c
}

// Stop ends the background cleanup goroutine. It is safe to call more than once.
func (c *Cache) Stop() {
	if c == nil {
		return
	}
	c.stop.Do(func() {
		if c.stopCh != nil {
			close(c.stopCh)
		}
	})
}

// Get retrieves a cached result if it exists and is not expired.
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

	return cloneResult(item.Result), true
}

// Set stores a search result in the cache.
func (c *Cache) Set(params types.YokaiSearchParams, result *types.YokaiSearchResult) {
	if c == nil || c.ttl <= 0 || result == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	key := c.generateKey(params)
	now := time.Now()
	c.items[key] = &CacheItem{
		Result:    cloneResult(result),
		ExpiresAt: now.Add(c.ttl),
		CreatedAt: now,
	}
}

// Clear removes all items from the cache.
func (c *Cache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

// Size returns the current number of items in the cache.
func (c *Cache) Size() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *Cache) generateKey(params types.YokaiSearchParams) string {
	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Sprintf("%s:%s:%s:%d", params.Name, params.Region, params.Category, params.Limit)
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}
	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
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
}

func cloneResult(result *types.YokaiSearchResult) *types.YokaiSearchResult {
	if result == nil {
		return nil
	}
	out := *result
	if result.Results == nil {
		return &out
	}
	out.Results = make([]types.YokaiBook, len(result.Results))
	copy(out.Results, result.Results)
	for i := range out.Results {
		out.Results[i].Subjects = cloneStrings(result.Results[i].Subjects)
		out.Results[i].CoverImageCandidates = cloneStrings(result.Results[i].CoverImageCandidates)
	}
	return &out
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
