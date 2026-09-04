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
		if strings.Contains(note, "Daily pick") && strings.Contains(note, "JST") {
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

	foundKappa := false
	for _, profile := range result.Profiles {
		if profile.Name == "Kappa" {
			foundKappa = true
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
	if !foundKappa {
		t.Fatalf("water filter should include Kappa")
	}
	if result.Returned > 3 {
		t.Fatalf("expected limit 3, got returned=%d", result.Returned)
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
	if all.Total < 160 {
		t.Fatalf("expected at least 160 indexed yokai, got total=%d", all.Total)
	}
	if all.Returned != all.Total || len(all.Items) != all.Total {
		t.Fatalf("expected full roster returned by default, got returned=%d items=%d total=%d", all.Returned, len(all.Items), all.Total)
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

func TestSuggestYokaiWaterOrDefault(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	water, err := h.SuggestYokai(context.Background(), types.SuggestYokaiParams{
		Vibe:  "水",
		Limit: 6,
	})
	if err != nil {
		t.Fatalf("SuggestYokai(水) error: %v", err)
	}
	if water.Returned == 0 || len(water.Items) == 0 {
		t.Fatalf("expected suggestions for vibe 水, got %+v", water)
	}
	if water.Returned > 6 {
		t.Fatalf("expected limit 6, got returned=%d", water.Returned)
	}
	for _, item := range water.Items {
		if item.Name == "" {
			t.Fatalf("expected non-empty name in suggestion: %+v", item)
		}
		if item.WhySuggested == "" {
			t.Fatalf("expected WhySuggested for %s", item.Name)
		}
	}

	defaults, err := h.SuggestYokai(context.Background(), types.SuggestYokaiParams{})
	if err != nil {
		t.Fatalf("SuggestYokai(empty) error: %v", err)
	}
	if defaults.Returned == 0 || len(defaults.Items) == 0 {
		t.Fatalf("expected well-known default suggestions, got %+v", defaults)
	}
	if defaults.Returned > 6 {
		t.Fatalf("expected default limit 6, got returned=%d", defaults.Returned)
	}
}

func TestGetYokaiProfileSource(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	result, err := h.GetYokai(context.Background(), types.GetYokaiParams{Name: "河童"})
	if err != nil {
		t.Fatalf("GetYokai(河童) error: %v", err)
	}
	if !result.Found {
		t.Fatalf("expected Found=true for 河童")
	}
	if result.Source != "profile" {
		t.Fatalf("expected source=profile, got %q", result.Source)
	}
	if result.Profile == nil {
		t.Fatalf("expected Profile to be set")
	}
	if result.Profile.NativeName != "河童" && result.Profile.Name != "Kappa" {
		t.Fatalf("unexpected profile: %+v", result.Profile)
	}
	if result.Index == nil {
		t.Fatalf("expected Index card (tags/tone) alongside profile")
	}
	if result.Index.NativeName != "河童" {
		t.Fatalf("unexpected index native name: %+v", result.Index)
	}
	if len(result.Index.Tags) == 0 {
		t.Fatalf("expected tags on index card")
	}
}

func TestGetYokaiUnknown(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	result, err := h.GetYokai(context.Background(), types.GetYokaiParams{Name: "存在しない妖怪xyz"})
	if err != nil {
		t.Fatalf("GetYokai(unknown) error: %v", err)
	}
	if result.Found {
		t.Fatalf("expected Found=false, got %+v", result)
	}
	if result.Source != "" {
		t.Fatalf("expected empty source, got %q", result.Source)
	}
	if result.Profile != nil || result.Index != nil {
		t.Fatalf("expected no profile/index payload")
	}
	if len(result.Notes) == 0 {
		t.Fatalf("expected notes suggesting alternatives")
	}
}

func TestGetYokaiIndexOnly(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	// わいら is expected to be index-only (no curated profile).
	result, err := h.GetYokai(context.Background(), types.GetYokaiParams{Name: "わいら"})
	if err != nil {
		t.Fatalf("GetYokai(わいら) error: %v", err)
	}
	if !result.Found {
		t.Fatalf("expected Found=true for index entry わいら")
	}
	if result.Source != "index" {
		t.Fatalf("expected source=index, got %q (profile may have been added)", result.Source)
	}
	if result.Index == nil {
		t.Fatalf("expected Index card")
	}
	if result.Index.NativeName != "わいら" && result.Index.Name != "Waira" {
		t.Fatalf("unexpected index item: %+v", result.Index)
	}
	if result.Profile != nil {
		t.Fatalf("did not expect Profile for index-only entry")
	}
	if len(result.Notes) == 0 {
		t.Fatalf("expected notes about limited lore / next steps")
	}
}

func TestGetYokaiAliasAndSuggestions(t *testing.T) {
	h := New(nil, nil)

	kappa, err := h.GetYokai(context.Background(), types.GetYokaiParams{Name: "カッパ"})
	if err != nil {
		t.Fatalf("GetYokai(カッパ) error: %v", err)
	}
	if !kappa.Found || kappa.Source != "profile" || kappa.Profile == nil || kappa.Profile.NativeName != "河童" {
		t.Fatalf("expected 河童 profile via カッパ, got %+v", kappa)
	}

	unknown, err := h.GetYokai(context.Background(), types.GetYokaiParams{Name: "Kapa"})
	if err != nil {
		t.Fatalf("GetYokai(Kapa) error: %v", err)
	}
	if unknown.Found {
		t.Fatalf("Kapa should not be an exact hit")
	}
	if len(unknown.Suggestions) == 0 {
		t.Fatalf("expected did-you-mean suggestions")
	}
}

func TestListYokaiEnglishCategoryAndTag(t *testing.T) {
	h := New(nil, nil)

	water, err := h.ListYokai(context.Background(), types.YokaiIndexParams{Category: "water", Limit: 50})
	if err != nil {
		t.Fatalf("ListYokai(water) error: %v", err)
	}
	if water.Total == 0 {
		t.Fatalf("expected water/水系 matches")
	}
	foundKappa := false
	for _, item := range water.Items {
		if item.NativeName == "河童" {
			foundKappa = true
		}
	}
	if !foundKappa {
		t.Fatalf("water category should include 河童")
	}

	horror := true
	tagged, err := h.ListYokai(context.Background(), types.YokaiIndexParams{
		Tag:        "怖い",
		Tone:       "horror",
		HasProfile: &horror,
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("tagged ListYokai error: %v", err)
	}
	if tagged.Returned == 0 {
		t.Fatalf("expected horror+怖い+profile matches")
	}
	for _, item := range tagged.Items {
		if item.Tone != "horror" {
			t.Fatalf("expected horror tone, got %s", item.Tone)
		}
		if !item.HasProfile {
			t.Fatalf("expected hasProfile filter to hold for %s", item.Name)
		}
	}
}

func TestRelatedAndCompare(t *testing.T) {
	h := New(nil, nil)

	related, err := h.RelatedYokai(context.Background(), types.RelatedYokaiParams{Name: "河童", Limit: 5})
	if err != nil {
		t.Fatalf("RelatedYokai error: %v", err)
	}
	if related.Returned == 0 {
		t.Fatalf("expected neighbours for 河童")
	}

	cmp, err := h.CompareYokai(context.Background(), types.CompareYokaiParams{Left: "河童", Right: "天狗"})
	if err != nil {
		t.Fatalf("CompareYokai error: %v", err)
	}
	if !cmp.Left.Found || !cmp.Right.Found {
		t.Fatalf("expected both sides found: %+v", cmp)
	}
	if cmp.Left.Profile == nil || cmp.Right.Profile == nil {
		t.Fatalf("expected profiles on both sides")
	}
	joined := strings.Join(cmp.Shared, " ")
	if !strings.Contains(joined, "入門") && !strings.Contains(joined, "古典") {
		t.Fatalf("expected shared folklore tags, got %v", cmp.Shared)
	}
}

func TestGetYokaiEmptyName(t *testing.T) {
	t.Helper()

	h := New(nil, nil)

	_, err := h.GetYokai(context.Background(), types.GetYokaiParams{Name: "  "})
	if err == nil {
		t.Fatalf("expected error for empty name")
	}
}
