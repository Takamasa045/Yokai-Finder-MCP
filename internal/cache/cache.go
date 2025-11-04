package cache

import (
	"sync"
	"time"
)

type entry struct {
	v       any
	expires time.Time
}

type TTLCache struct {
	mu   sync.Mutex
	data map[string]entry
	ttl  time.Duration
	maxN int
}

func New(ttl time.Duration, maxN int) *TTLCache {
	return &TTLCache{data: make(map[string]entry), ttl: ttl, maxN: maxN}
}

func (c *TTLCache) Get(k string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.data[k]
	if !ok || time.Now().After(e.expires) {
		if ok {
			delete(c.data, k)
		}
		return nil, false
	}
	return e.v, true
}

func (c *TTLCache) Set(k string, v any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.data) >= c.maxN {
		// 超雑に1件掃除
		for kk := range c.data {
			delete(c.data, kk)
			break
		}
	}
	c.data[k] = entry{v: v, expires: time.Now().Add(c.ttl)}
}
