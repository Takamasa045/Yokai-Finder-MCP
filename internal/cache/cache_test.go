package cache

import (
	"testing"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache(time.Minute, 10)
	t.Cleanup(c.Stop)

	params := types.YokaiSearchParams{Name: "天狗", Limit: 5}
	expected := &types.YokaiSearchResult{
		Total:   1,
		Results: []types.YokaiBook{{Title: "天狗大全", Subjects: []string{"妖怪"}}},
	}

	c.Set(params, expected)
	got, ok := c.Get(params)
	if !ok {
		t.Fatalf("expected cache hit")
	}
	if got.Total != expected.Total || got.Results[0].Title != "天狗大全" {
		t.Fatalf("cache returned unexpected result: %+v", got)
	}
	got.Results[0].Title = "MUTATED"
	expected.Results[0].Subjects[0] = "MUTATED"
	again, ok := c.Get(params)
	if !ok {
		t.Fatalf("expected cache hit after mutation")
	}
	if again.Results[0].Title != "天狗大全" {
		t.Fatalf("cache should return a defensive copy, got %q", again.Results[0].Title)
	}
	if again.Results[0].Subjects[0] != "妖怪" {
		t.Fatalf("cache should deep-copy subjects, got %q", again.Results[0].Subjects[0])
	}
}

func TestCacheStopIdempotent(t *testing.T) {
	c := NewCache(time.Minute, 4)
	c.Stop()
	c.Stop()
}

func TestCacheExpiry(t *testing.T) {
	c := NewCache(10*time.Millisecond, 10)
	t.Cleanup(c.Stop)
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

func TestCacheEvictsOldestWhenFull(t *testing.T) {
	c := NewCache(time.Minute, 2)
	t.Cleanup(c.Stop)

	first := types.YokaiSearchParams{Name: "河童"}
	second := types.YokaiSearchParams{Name: "天狗"}
	third := types.YokaiSearchParams{Name: "鵺"}

	c.Set(first, &types.YokaiSearchResult{Total: 1})
	time.Sleep(2 * time.Millisecond)
	c.Set(second, &types.YokaiSearchResult{Total: 2})
	time.Sleep(2 * time.Millisecond)
	c.Set(third, &types.YokaiSearchResult{Total: 3})

	if c.Size() != 2 {
		t.Fatalf("expected size capped at 2, got %d", c.Size())
	}
	if _, ok := c.Get(first); ok {
		t.Fatalf("expected the oldest entry to be evicted")
	}
	if _, ok := c.Get(third); !ok {
		t.Fatalf("expected the newest entry to survive eviction")
	}
}

func TestCacheClearAndSize(t *testing.T) {
	c := NewCache(time.Minute, 10)
	t.Cleanup(c.Stop)
	c.Set(types.YokaiSearchParams{Name: "化け猫"}, &types.YokaiSearchResult{Total: 1})
	c.Set(types.YokaiSearchParams{Name: "雪女"}, &types.YokaiSearchResult{Total: 2})

	if c.Size() != 2 {
		t.Fatalf("expected size 2, got %d", c.Size())
	}

	c.Clear()

	if c.Size() != 0 {
		t.Fatalf("expected empty cache after Clear, got %d", c.Size())
	}
}
