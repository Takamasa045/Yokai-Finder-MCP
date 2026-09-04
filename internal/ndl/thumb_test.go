package ndl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
)

func TestNormalizeISBN13(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"hyphenated isbn13", "978-4-04-883992-1", "9784048839921"},
		{"plain isbn13", "9784048839921", "9784048839921"},
		{"isbn10 converts to isbn13", "4048839926", "9784048839921"},
		{"hyphenated isbn10", "4-04-883992-6", "9784048839921"},
		{"isbn10 with X check digit", "404883992X", "9784048839921"},
		{"empty", "", ""},
		{"garbage", "not-an-isbn", ""},
		{"too short", "12345", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeISBN13(tc.in); got != tc.want {
				t.Fatalf("NormalizeISBN13(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildCoverURLs(t *testing.T) {
	covers := BuildCoverURLs("978-4-04-883992-1", "")
	if len(covers.Candidates) == 0 {
		t.Fatalf("expected candidates for a valid ISBN")
	}
	for _, u := range covers.Candidates {
		if !strings.Contains(u, "9784048839921") {
			t.Fatalf("candidate %q should embed the normalized ISBN", u)
		}
	}

	empty := BuildCoverURLs("", "")
	if len(empty.Candidates) != 0 {
		t.Fatalf("expected no candidates without identifiers")
	}

	jpno := BuildCoverURLs("", "22222222")
	if len(jpno.Candidates) == 0 {
		t.Fatalf("expected JP number fallback candidates")
	}

	junk := BuildCoverURLs("", "../../../etc/passwd")
	if len(junk.Candidates) != 0 {
		t.Fatalf("expected invalid jpno to be rejected")
	}
}

func TestVerifyCoverCandidates(t *testing.T) {
	var sawHead bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && r.URL.Path == "/ok.jpg" {
			sawHead = true
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer ts.Close()

	books := []types.YokaiBook{{
		CoverImageCandidates: []string{ts.URL + "/ok.jpg", ts.URL + "/missing.jpg"},
	}}
	verifyCoverCandidates(context.Background(), books, ts.Client())
	if !sawHead {
		t.Fatal("expected HEAD probe")
	}
	if len(books[0].CoverImageCandidates) != 1 || !strings.HasSuffix(books[0].CoverImageCandidates[0], "/ok.jpg") {
		t.Fatalf("unexpected candidates %v", books[0].CoverImageCandidates)
	}
}
