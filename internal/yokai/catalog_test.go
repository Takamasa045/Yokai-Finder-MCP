package yokai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedCatalogSizes(t *testing.T) {
	if got := len(Index()); got < 160 {
		t.Fatalf("index too small: %d", got)
	}
	if got := len(Profiles()); got < 98 {
		t.Fatalf("profiles too small: %d", got)
	}
	if len(ExtraAliases()) < 50 {
		t.Fatalf("aliases too small: %d", len(ExtraAliases()))
	}
}

func TestValidateLoadedCatalogRejectsBadData(t *testing.T) {
	good := Index()[0]
	good.Tags = []string{"a", "b", "c"}
	err := validateLoadedCatalog(
		[]IndexEntry{good},
		[]Profile{{
			Name: "Ghost", NativeName: "幽霊X", Category: "死霊",
			Summary: "s", SummaryJA: "概要", FunFact: "f", FunFactJA: "豆",
			SearchQuery: "幽霊", Legends: []string{"lore"},
		}},
		map[string]string{"phantom": "nobody"},
	)
	if err == nil {
		t.Fatal("expected validation errors")
	}
	msg := err.Error()
	for _, want := range []string{"missing from the index", "unknown yokai"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error should mention %q, got %s", want, msg)
		}
	}

	if err := validateLoadedCatalog(nil, nil, nil); err == nil {
		t.Fatal("empty catalog should fail")
	}

	split := good
	split.Name = "Kappa"
	split.NativeName = "河童"
	err = validateLoadedCatalog(
		[]IndexEntry{split},
		[]Profile{{
			Name: "Kappa", NativeName: "河童X", Category: split.Category,
			Summary: "s", SummaryJA: "概要", FunFact: "f", FunFactJA: "豆",
			SearchQuery: "河童", Legends: []string{"lore"},
		}},
		map[string]string{},
	)
	if err == nil || !strings.Contains(err.Error(), "missing from the index") && !strings.Contains(err.Error(), "different index") {
		t.Fatalf("split-brain profile should fail, got %v", err)
	}
}

func TestDocsMentionLiveCatalogCounts(t *testing.T) {
	root := filepath.Join("..", "..")
	info := CatalogOverview()
	files := []string{"README.md", "mcp.json", "server.json"}
	needles := []string{
		fmt.Sprintf("%d", info.IndexCount),
		fmt.Sprintf("%d", info.ProfileCount),
	}
	for _, name := range files {
		body, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(body)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				t.Errorf("%s should mention catalog count %s", name, needle)
			}
		}
	}
}
