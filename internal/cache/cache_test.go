package cache

import (
	"testing"
	"time"

	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache(time.Minute, 10)
	params := types.YokaiSearchParams{Name: "天狗", Limit: 5}
	expected := &types.YokaiSearchResult{Total: 1}

	c.Set(params, expected)
	got, ok := c.Get(params)
	if !ok {
		t.Fatalf("expected cache hit")
	}
	if got != expected {
		t.Fatalf("cache returned unexpected pointer")
	}
}

func TestCacheExpiry(t *testing.T) {
	c := NewCache(10*time.Millisecond, 10)
	params := types.YokaiSearchParams{Name: "雪女"}
	c.Set(params, &types.YokaiSearchResult{Total: 2})

	time.Sleep(30 * time.Millisecond)

	if _, ok := c.Get(params); ok {
		t.Fatalf("expected cache miss after expiry")
	}
}

func TestCacheDisabledWhenTTLZero(t *testing.T) {
	c := NewCache(0, 10)
	params := types.YokaiSearchParams{Name: "座敷童子"}
	c.Set(params, &types.YokaiSearchResult{Total: 3})

	if _, ok := c.Get(params); ok {
		t.Fatalf("expected cache miss when ttl is zero")
	}
}
