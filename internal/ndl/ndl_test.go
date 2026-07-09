package ndl

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
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
      <title>妖怪退治入門</title>
      <link>https://example.com/book</link>
      <description><![CDATA[<p>妖怪の基礎知識を紹介する解説書。</p>]]></description>
      <dc:creator>山田 太郎</dc:creator>
      <dc:publisher>民俗学出版</dc:publisher>
      <dcterms:issued>2020</dcterms:issued>
      <dc:subject>妖怪</dc:subject>
      <dc:identifier xsi:type="dcndl:ISBN">978-4-0000-0000-0</dc:identifier>
    </item>
  </channel>
</rss>`

func TestSearchYokaiBooks(t *testing.T) {
	t.Helper()

	var requestCount int
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(sampleRSS))
	}))
	defer mockServer.Close()

	client := NewClient().WithBaseURL(mockServer.URL).WithHTTPClient(mockServer.Client())

	params := types.YokaiSearchParams{Name: "天狗", Limit: 5}
	result, err := client.SearchYokaiBooks(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}

	book := result.Results[0]
	t.Logf("parsed book: %+v", book)
	if book.Title != "妖怪退治入門" {
		t.Errorf("unexpected title: %s", book.Title)
	}
	if book.Author != "山田 太郎" {
		t.Errorf("unexpected author: %s", book.Author)
	}
	if book.Publisher != "民俗学出版" {
		t.Errorf("unexpected publisher: %s", book.Publisher)
	}
	if book.PublishDate != "2020" {
		t.Errorf("unexpected publish date: %s", book.PublishDate)
	}
	if book.ISBN != "978-4-0000-0000-0" {
		t.Errorf("unexpected isbn: %s", book.ISBN)
	}
	if book.Description == "" {
		t.Errorf("expected description to be populated")
	}

	if requestCount != 1 {
		t.Fatalf("expected 1 request, got %d", requestCount)
	}
}

func TestBuildQuery(t *testing.T) {
	cases := []struct {
		name   string
		params types.YokaiSearchParams
		want   string
	}{
		{
			name:   "defaults to yokai",
			params: types.YokaiSearchParams{},
			want:   "妖怪",
		},
		{
			name:   "joins fields",
			params: types.YokaiSearchParams{Name: "天狗", Region: "高尾山", Category: "伝説"},
			want:   "天狗 高尾山 伝説",
		},
		{
			name:   "trims whitespace",
			params: types.YokaiSearchParams{Name: "  河童  ", Region: "  東北 "},
			want:   "河童 東北",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildQuery(tc.params); got != tc.want {
				t.Fatalf("buildQuery() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSanitizeDescription(t *testing.T) {
	raw := "<p>妖怪を <strong>分かりやすく</strong> 紹介します。</p>"
	got := sanitizeDescription(raw)
	if got != "妖怪を 分かりやすく 紹介します。" {
		t.Fatalf("unexpected sanitized description: %q", got)
	}
}

func TestExtractISBN(t *testing.T) {
	ids := []xmlIdentifier{
		{Type: "dcndl:TRCMARCNO", Value: "22102386"},
		{Type: "dcndl:ISBN", Value: "978-4-0000-0000-0"},
	}

	if isbn := extractISBN(ids); isbn != "978-4-0000-0000-0" {
		t.Fatalf("unexpected isbn: %q", isbn)
	}
}

func TestSearchYokaiBooksLive(t *testing.T) {
	t.Skip("integration test for manual debugging")

	client := NewClient()
	params := types.YokaiSearchParams{Name: "妖怪", Limit: 1}

	rawURL := client.buildAPIURL(buildQuery(params), 1)
	resp, err := http.Get(rawURL)
	if err != nil {
		t.Fatalf("direct http get failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	t.Logf("raw body snippet: %s", string(body[:min(len(body), 600)]))

	result, err := client.SearchYokaiBooks(context.Background(), params)
	if err != nil {
		t.Fatalf("live query failed: %v", err)
	}
	t.Logf("live result: %+v", result)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestParseSample(t *testing.T) {
	result, err := parseRSS([]byte(sampleRSS))
	if err != nil {
		t.Fatalf("parseRSS returned error: %v", err)
	}
	if len(result.Results) == 0 {
		t.Fatalf("expected results")
	}
	book := result.Results[0]
	t.Logf("sample book: %+v", book)
}
