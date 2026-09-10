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

func TestCatalogArtifactDetectsResidue(t *testing.T) {
	cases := map[string]bool{
		"Do not merge 大蟇 into a neighbour card.":  true,
		"Keep 化け蟹 as the faced shell.":            true,
		"Encyclopedia body":                       true,
		"folk plate":                              true,
		"The workplace is 巨大な蟇.":                  true,
		"井戸は別項。":                                  true,
		"化け蟹の、巨大側。":                               true,
		"Challenges travellers to sumo bouts":     false,
		"A municipal treaty written in cucumbers": false,
	}
	for text, want := range cases {
		got := CatalogArtifact(text) != ""
		if got != want {
			t.Errorf("%q artifact=%v want %v (hit %q)", text, got, want, CatalogArtifact(text))
		}
	}
}

func TestValidateLoadedCatalogRejectsArtifacts(t *testing.T) {
	good := Index()[0]
	good.Tags = []string{"a", "b", "c"}
	good.FamousRank = 3
	good.Name = "Okiseru"
	good.NativeName = "大煙管"
	err := validateLoadedCatalog(
		[]IndexEntry{good},
		[]Profile{{
			Name: "Okiseru", NativeName: "大煙管", Category: good.Category,
			Summary: "s", SummaryJA: "巨大な煙管の付喪神。", FunFact: "f", FunFactJA: "豆知識",
			SearchQuery: "大煙管",
			Legends:     []string{"Do not merge 大煙管 into a neighbour card."},
			Sources:     []string{"大煙管 / encyclopedia plate"},
		}},
		map[string]string{},
	)
	if err == nil {
		t.Fatal("expected residue to fail validation")
	}
	if !strings.Contains(err.Error(), "generation residue") {
		t.Fatalf("expected residue error, got %v", err)
	}
}

func TestLiveCatalogHasNoGenerationResidue(t *testing.T) {
	for _, e := range Index() {
		if hit := indexArtifact(e); hit != "" {
			t.Errorf("%s index residue %q", e.NativeName, hit)
		}
	}
	for _, p := range Profiles() {
		if hit := profileArtifact(p); hit != "" {
			t.Errorf("%s profile residue %q", p.NativeName, hit)
		}
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
