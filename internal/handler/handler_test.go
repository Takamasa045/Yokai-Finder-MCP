package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourname/yokai-finder-mcp/internal/cache"
	"github.com/yourname/yokai-finder-mcp/internal/ndl"
	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss xmlns:dc="http://purl.org/dc/elements/1.1/"
     xmlns:dcterms="http://purl.org/dc/terms/"
     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
     xmlns:dcndl="http://ndl.go.jp/dcndl/terms/"
     xmlns:openSearch="http://a9.com/-/spec/opensearchrss/1.0/">
  <channel>
    <openSearch:totalResults>1</openSearch:totalResults>
    <item>
      <title>河童と水神信仰</title>
      <link>https://example.com/kappa</link>
      <description><![CDATA[<p>河童伝説にまつわるフィールドワーク報告。</p>]]></description>
      <dc:creator>佐藤 花子</dc:creator>
      <dc:publisher>水辺文化社</dc:publisher>
      <dcterms:issued>2018</dcterms:issued>
      <dc:subject>河童</dc:subject>
      <dc:identifier xsi:type="dcndl:ISBN">978-4-1111-1111-1</dc:identifier>
    </item>
  </channel>
</rss>`

func TestSearchYokaiUsesCache(t *testing.T) {
	var callCount int32

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(sampleRSS))
	}))
	defer mockServer.Close()

	ndlClient := ndl.NewClient().WithBaseURL(mockServer.URL).WithHTTPClient(mockServer.Client())
	cache := cache.NewCache(time.Minute, 10)
	h := New(ndlClient, cache)

	params := types.YokaiSearchParams{Name: "河童", Limit: 3}

	ctx := context.Background()
	result1, err := h.SearchYokai(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result1 == nil {
		t.Fatalf("expected result on first call")
	}

	result2, err := h.SearchYokai(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result2 == nil {
		t.Fatalf("expected result on second call")
	}

	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatalf("expected API to be called once, got %d", callCount)
	}
}

func TestNormaliseParams(t *testing.T) {
	params := types.YokaiSearchParams{Name: "  狐  ", Region: "  北海道", Category: "伝承  ", Limit: 200}
	got := normaliseParams(params)
	if got.Name != "狐" {
		t.Errorf("expected trimmed name, got %q", got.Name)
	}
	if got.Region != "北海道" {
		t.Errorf("expected trimmed region, got %q", got.Region)
	}
	if got.Category != "伝承" {
		t.Errorf("expected trimmed category, got %q", got.Category)
	}
	if got.Limit != 50 {
		t.Errorf("expected limit clamp to 50, got %d", got.Limit)
	}
}

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss xmlns:dc="http://purl.org/dc/elements/1.1/"
     xmlns:dcterms="http://purl.org/dc/terms/"
     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
     xmlns:dcndl="http://ndl.go.jp/dcndl/terms/">
  <channel>
    <item>
      <title>妖怪退治入門</title>
      <dc:creator>山田 太郎</dc:creator>
      <dc:identifier xsi:type="dcndl:ISBN">978-4-0000-0000-0</dc:identifier>
    </item>
  </channel>
</rss>`
