package yokai

import (
	"strings"
	"testing"
)

func TestIndexSize(t *testing.T) {
	idx := Index()
	if got := len(idx); got != 111 {
		t.Fatalf("expected 111 indexed yokai, got %d", got)
	}
}

func TestIndexHasRequiredFields(t *testing.T) {
	seenName := make(map[string]bool)
	seenNative := make(map[string]bool)

	for _, e := range Index() {
		if strings.TrimSpace(e.Name) == "" {
			t.Errorf("index entry missing Name")
		}
		if strings.TrimSpace(e.NativeName) == "" {
			t.Errorf("%s missing NativeName", e.Name)
		}
		if strings.TrimSpace(e.Category) == "" {
			t.Errorf("%s missing Category", e.Name)
		}
		if strings.TrimSpace(e.BlurbJA) == "" {
			t.Errorf("%s missing BlurbJA", e.Name)
		}
		if seenName[e.Name] {
			t.Errorf("duplicate Name %q", e.Name)
		}
		seenName[e.Name] = true
		if seenNative[e.NativeName] {
			t.Errorf("duplicate NativeName %q", e.NativeName)
		}
		seenNative[e.NativeName] = true
	}
}

func TestIndexIncludesCuratedRoster(t *testing.T) {
	for _, p := range Profiles() {
		found := false
		for _, e := range Index() {
			if e.Name == p.Name || e.NativeName == p.NativeName {
				if !e.HasCuratedProfile() {
					t.Errorf("%s should report HasCuratedProfile", e.Name)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("curated profile %s (%s) missing from index", p.Name, p.NativeName)
		}
	}
}

func TestFilterIndex(t *testing.T) {
	water := FilterIndex("", "水系", "")
	if len(water) == 0 {
		t.Fatalf("expected 水系 category matches")
	}
	for _, e := range water {
		if !strings.Contains(e.Category, "水系") {
			t.Errorf("unexpected category %q for %s", e.Category, e.Name)
		}
	}

	kappa := FilterIndex("河童", "", "")
	if len(kappa) == 0 {
		t.Fatalf("expected term 河童 to match")
	}
	if kappa[0].NativeName != "河童" && kappa[0].Name != "Kappa" {
		// allow blurb matches too, but first should prefer direct name hits in list order
		found := false
		for _, e := range kappa {
			if e.NativeName == "河童" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected 河童 in term filter results")
		}
	}

	none := FilterIndex("zzzz-not-a-yokai", "", "")
	if len(none) != 0 {
		t.Fatalf("expected no matches for nonsense term, got %d", len(none))
	}
}

func TestIndexDefensiveCopy(t *testing.T) {
	a := Index()
	a[0].Name = "MUTATED"
	b := Index()
	if b[0].Name == "MUTATED" {
		t.Fatalf("Index() should return a defensive copy")
	}
}
