package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/ndl"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
)

const sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:os="http://a9.com/-/spec/opensearchrss/1.0/" version="2.0">
  <channel>
    <title>Sample Yokai Results</title>
    <os:totalResults>12</os:totalResults>
    <item>
      <title>河童大全</title>
      <link>https://example.org/kappa1</link>
      <description><![CDATA[All about the kappa's watery deals.]]></description>
      <dc:creator>Folklore Society</dc:creator>
      <dc:publisher>Example Press</dc:publisher>
      <dc:date>1999</dc:date>
      <dcterms:issued>1999</dcterms:issued>
      <dc:subject>妖怪</dc:subject>
      <dc:identifier>ISBN:1234567890</dc:identifier>
    </item>
    <item>
      <title>川と河童の博物誌</title>
      <link>https://example.org/kappa2</link>
      <description><![CDATA[Exploring river spirits and their legends.]]></description>
      <dc:creator>Mizuno Haru</dc:creator>
      <dc:publisher>River Books</dc:publisher>
      <dc:date>2008</dc:date>
      <dcterms:issued>2008</dcterms:issued>
      <dc:subject>伝承</dc:subject>
      <dc:identifier>ISBN:0987654321</dc:identifier>
    </item>
  </channel>
</rss>`

func TestHandlerYokaiOfTheDayByName(t *testing.T) {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, sampleRSS)
	}))
	defer ts.Close()

	client := ndl.NewClient().WithHTTPClient(ts.Client()).WithBaseURL(ts.URL)
	c := cache.NewCache(time.Minute, 16)
	h := New(client, c)

	result, err := h.YokaiOfTheDay(context.Background(), types.YokaiOfTheDayParams{
		Name:  "Kappa",
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("YokaiOfTheDay returned error: %v", err)
	}

	if result.Profile.Name != "Kappa" {
		t.Fatalf("unexpected profile name: %s", result.Profile.Name)
	}
	if result.Profile.NativeName != "河童" {
		t.Fatalf("unexpected profile native name: %s", result.Profile.NativeName)
	}
	if result.TotalBooks != 12 {
		t.Fatalf("expected total books to be 12, got %d", result.TotalBooks)
	}
	if len(result.RecommendedBooks) != 2 {
		t.Fatalf("expected 2 recommended books, got %d", len(result.RecommendedBooks))
	}
	if !strings.Contains(result.StoryPrompt, "Kappa") {
		t.Fatalf("expected story prompt to mention Kappa, got %q", result.StoryPrompt)
	}
	if len(result.Notes) != 0 {
		t.Fatalf("expected no notes, got %v", result.Notes)
	}
	if result.Query == "" {
		t.Fatalf("expected a populated query")
	}
}

func TestHandlerYokaiOfTheDayFallback(t *testing.T) {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer ts.Close()

	client := ndl.NewClient().WithHTTPClient(ts.Client()).WithBaseURL(ts.URL)
	c := cache.NewCache(time.Minute, 16)
	h := New(client, c)

	result, err := h.YokaiOfTheDay(context.Background(), types.YokaiOfTheDayParams{
		Category: "space pirates", // intentionally unmatched
		Seed:     42,
		Limit:    1,
	})
	if err != nil {
		t.Fatalf("YokaiOfTheDay returned error: %v", err)
	}

	if result.Profile.Name == "" {
		t.Fatalf("expected a fallback profile, got empty name")
	}
	if len(result.Notes) == 0 {
		t.Fatalf("expected notes describing the fallback behaviour")
	}
	// Should include both filter fallback and search failure notes.
	var sawFilterNote, sawSearchNote bool
	for _, note := range result.Notes {
		if strings.Contains(note, "surprise pick") {
			sawFilterNote = true
		}
		if strings.Contains(note, "NDL search unavailable") {
			sawSearchNote = true
		}
	}
	if !sawFilterNote || !sawSearchNote {
		t.Fatalf("expected both filter and search notes, got %v", result.Notes)
	}
	if len(result.RecommendedBooks) != 0 {
		t.Fatalf("expected no recommended books when search fails, got %d", len(result.RecommendedBooks))
	}
	if result.TotalBooks != 0 {
		t.Fatalf("expected total books to be zero on failure, got %d", result.TotalBooks)
	}
}
