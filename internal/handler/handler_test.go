package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestHandlerYokaiOfTheDayDailyNote(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, sampleRSS)
	}))
	defer ts.Close()

	client := ndl.NewClient().WithHTTPClient(ts.Client()).WithBaseURL(ts.URL)
	h := New(client, cache.NewCache(time.Minute, 16))

	first, err := h.YokaiOfTheDay(context.Background(), types.YokaiOfTheDayParams{})
	if err != nil {
		t.Fatalf("YokaiOfTheDay returned error: %v", err)
	}

	var sawDailyNote bool
	for _, note := range first.Notes {
		if strings.Contains(note, "Daily pick") {
			sawDailyNote = true
		}
	}
	if !sawDailyNote {
		t.Fatalf("expected a daily-pick note, got %v", first.Notes)
	}

	second, err := h.YokaiOfTheDay(context.Background(), types.YokaiOfTheDayParams{})
	if err != nil {
		t.Fatalf("YokaiOfTheDay returned error: %v", err)
	}
	if first.Profile.Name != second.Profile.Name {
		t.Fatalf("expected the same daily yokai on repeat calls, got %s vs %s", first.Profile.Name, second.Profile.Name)
	}
}

func TestListCuratedYokaiFiltersAndShape(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	result, err := h.ListCuratedYokai(context.Background(), types.CuratedYokaiParams{
		Category:             "water",
		IncludeLegends:       true,
		IncludeCreativeHooks: true,
		Limit:                3,
	})
	if err != nil {
		t.Fatalf("ListCuratedYokai returned error: %v", err)
	}
	if result.Total == 0 {
		t.Fatalf("expected filtered results")
	}
	if result.Returned == 0 || len(result.Profiles) == 0 {
		t.Fatalf("expected at least one profile, got %+v", result)
	}

	for _, profile := range result.Profiles {
		if !strings.Contains(strings.ToLower(profile.Category), "water") {
			t.Fatalf("expected category to include 'water', got %q", profile.Category)
		}
		if len(profile.Legends) == 0 {
			t.Fatalf("expected legends to be included when requested")
		}
		if profile.Traits != nil {
			t.Fatalf("did not expect traits without IncludeTraits flag")
		}
		if profile.CreativeHooks == nil {
			t.Fatalf("expected creative hooks when requested")
		}
	}
	if len(result.Notes) != 0 {
		t.Fatalf("expected no truncation notes when limit exceeds matches, got %v", result.Notes)
	}
}

func TestListCuratedYokaiDeterministicShuffle(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	params := types.CuratedYokaiParams{
		Seed:  99,
		Limit: 5,
	}

	first, err := h.ListCuratedYokai(context.Background(), params)
	if err != nil {
		t.Fatalf("ListCuratedYokai returned error: %v", err)
	}

	second, err := h.ListCuratedYokai(context.Background(), params)
	if err != nil {
		t.Fatalf("second run returned error: %v", err)
	}

	if !reflect.DeepEqual(first.Profiles, second.Profiles) {
		t.Fatalf("expected deterministic ordering with seed, got\nfirst:  %+v\nsecond: %+v", first.Profiles, second.Profiles)
	}

	params.Seed = 0
	sorted, err := h.ListCuratedYokai(context.Background(), params)
	if err != nil {
		t.Fatalf("expected sorted run to succeed: %v", err)
	}
	for i := 1; i < len(sorted.Profiles); i++ {
		if strings.ToLower(sorted.Profiles[i-1].Name) > strings.ToLower(sorted.Profiles[i].Name) {
			t.Fatalf("expected alphabetical order when no seed is provided")
		}
	}
}

func TestListYokaiIndex(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	all, err := h.ListYokai(context.Background(), types.YokaiIndexParams{})
	if err != nil {
		t.Fatalf("ListYokai returned error: %v", err)
	}
	if all.Total != 111 {
		t.Fatalf("expected 111 indexed yokai, got total=%d", all.Total)
	}
	if all.Returned != 111 || len(all.Items) != 111 {
		t.Fatalf("expected full roster returned by default, got returned=%d items=%d", all.Returned, len(all.Items))
	}

	var sawCurated bool
	for _, item := range all.Items {
		if item.NativeName == "河童" {
			if !item.HasProfile {
				t.Fatalf("河童 should have hasProfile=true")
			}
			if item.BlurbJA == "" {
				t.Fatalf("expected blurbJa for 河童")
			}
			sawCurated = true
		}
	}
	if !sawCurated {
		t.Fatalf("expected 河童 in index")
	}

	filtered, err := h.ListYokai(context.Background(), types.YokaiIndexParams{
		Category: "付喪神",
		Limit:    5,
	})
	if err != nil {
		t.Fatalf("filtered ListYokai error: %v", err)
	}
	if filtered.Total == 0 {
		t.Fatalf("expected 付喪神 matches")
	}
	if filtered.Returned > 5 {
		t.Fatalf("expected limit 5, got returned=%d", filtered.Returned)
	}
	for _, item := range filtered.Items {
		if !strings.Contains(item.Category, "付喪神") {
			t.Fatalf("unexpected category %q", item.Category)
		}
	}

	none, err := h.ListYokai(context.Background(), types.YokaiIndexParams{Term: "zzzz-nope"})
	if err != nil {
		t.Fatalf("empty ListYokai error: %v", err)
	}
	if none.Total != 0 || len(none.Items) != 0 {
		t.Fatalf("expected empty result, got %+v", none)
	}
	if len(none.Notes) == 0 {
		t.Fatalf("expected a note when nothing matches")
	}
}
