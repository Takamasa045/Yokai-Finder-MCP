package yokai

import (
	"strings"
	"testing"
)

func TestIndexSize(t *testing.T) {
	idx := Index()
	if got := len(idx); got < 160 {
		t.Fatalf("expected at least 160 indexed yokai, got %d", got)
	}
}

func TestIndexHasRequiredFields(t *testing.T) {
	seenName := make(map[string]bool)
	seenNative := make(map[string]bool)
	validTones := map[string]bool{
		"gentle": true, "comic": true, "horror": true,
		"solemn": true, "tragic": true, "mysterious": true, "playful": true,
	}

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
		if len(e.Tags) < 3 || len(e.Tags) > 8 {
			t.Errorf("%s Tags count %d want 3–8: %v", e.Name, len(e.Tags), e.Tags)
		}
		for _, tag := range e.Tags {
			if strings.TrimSpace(tag) == "" {
				t.Errorf("%s has empty tag", e.Name)
			}
		}
		if !validTones[e.Tone] {
			t.Errorf("%s invalid Tone %q", e.Name, e.Tone)
		}
		if e.FamousRank < 1 || e.FamousRank > 5 {
			t.Errorf("%s FamousRank %d want 1–5", e.Name, e.FamousRank)
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

func TestFilterIndexTohokuIncludesNamahage(t *testing.T) {
	hits := FilterIndex("", "", "tohoku")
	found := false
	for _, e := range hits {
		if e.NativeName == "なまはげ" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("tohoku should include なまはげ, got %d hits", len(hits))
	}
}

func TestFilterIndexEnglishCategory(t *testing.T) {
	ja := FilterIndex("", "水系", "")
	en := FilterIndex("", "water", "")
	if len(ja) == 0 || len(en) == 0 {
		t.Fatalf("expected 水系 and water to match, got ja=%d en=%d", len(ja), len(en))
	}
	if len(ja) != len(en) {
		t.Fatalf("水系 (%d) and water (%d) should return the same roster", len(ja), len(en))
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

	// Tags / Tone should also match term
	horrorTag := FilterIndex("怖い", "", "")
	if len(horrorTag) == 0 {
		t.Fatalf("expected tag 怖い matches via FilterIndex")
	}
	toneHit := FilterIndex("horror", "", "")
	if len(toneHit) == 0 {
		t.Fatalf("expected tone horror matches via FilterIndex")
	}

	none := FilterIndex("zzzz-not-a-yokai", "", "")
	if len(none) != 0 {
		t.Fatalf("expected no matches for nonsense term, got %d", len(none))
	}
}

func TestIndexDefensiveCopy(t *testing.T) {
	a := Index()
	a[0].Name = "MUTATED"
	if len(a[0].Tags) > 0 {
		a[0].Tags[0] = "MUTATED_TAG"
	}
	b := Index()
	if b[0].Name == "MUTATED" {
		t.Fatalf("Index() should return a defensive copy")
	}
	if len(b[0].Tags) > 0 && b[0].Tags[0] == "MUTATED_TAG" {
		t.Fatalf("Index() should deep-copy Tags")
	}
}

func TestFindIndexByName(t *testing.T) {
	e, ok := FindIndexByName("河童")
	if !ok {
		t.Fatalf("expected FindIndexByName(河童)")
	}
	if e.Name != "Kappa" || e.NativeName != "河童" {
		t.Fatalf("unexpected entry: %+v", e)
	}

	e2, ok := FindIndexByName("Kappa")
	if !ok {
		t.Fatalf("expected FindIndexByName(Kappa)")
	}
	if e2.NativeName != "河童" {
		t.Fatalf("Kappa should map to 河童, got %s", e2.NativeName)
	}

	e3, ok := FindIndexByName("kappa")
	if !ok {
		t.Fatalf("expected case-insensitive English match for kappa")
	}
	if e3.Name != "Kappa" {
		t.Fatalf("got %s", e3.Name)
	}

	if _, ok := FindIndexByName(""); ok {
		t.Fatalf("empty name should not match")
	}
	if _, ok := FindIndexByName("not-a-real-yokai-xyz"); ok {
		t.Fatalf("unknown name should not match")
	}
}

func TestSuggestWater(t *testing.T) {
	results := Suggest(SuggestQuery{Vibe: "水", Limit: 8})
	if len(results) == 0 {
		t.Fatalf("Suggest(水) returned nothing")
	}
	if len(results) > 8 {
		t.Fatalf("limit exceeded: %d", len(results))
	}
	waterish := 0
	for _, e := range results {
		pool := strings.ToLower(e.Category + e.Region + e.BlurbJA + strings.Join(e.Tags, " "))
		if strings.Contains(pool, "水") || strings.Contains(pool, "海") || strings.Contains(pool, "川") ||
			strings.Contains(e.Category, "水系") || strings.Contains(e.Category, "海") {
			waterish++
		}
	}
	if waterish == 0 {
		t.Fatalf("Suggest(水) expected water-related entries, got: names=%v", namesOf(results))
	}
}

func TestSuggestHorror(t *testing.T) {
	results := Suggest(SuggestQuery{Vibe: "怖い", Limit: 8})
	if len(results) == 0 {
		t.Fatalf("Suggest(怖い) returned nothing")
	}
	horrorish := 0
	for _, e := range results {
		if e.Tone == "horror" || e.Tone == "tragic" || e.Tone == "mysterious" {
			horrorish++
			continue
		}
		for _, tag := range e.Tags {
			if strings.Contains(tag, "怖") || tag == "現代" || tag == "死霊" {
				horrorish++
				break
			}
		}
	}
	if horrorish == 0 {
		t.Fatalf("Suggest(怖い) expected horror-ish entries, got: %v tones", tonesOf(results))
	}
}

func TestSuggestEmptyQueryNoPanic(t *testing.T) {
	results := Suggest(SuggestQuery{})
	if len(results) == 0 {
		t.Fatalf("empty Suggest should return well-known set, got 0")
	}
	if len(results) > 6 {
		t.Fatalf("default limit is 6, got %d", len(results))
	}
	// Should prefer famous ranks
	for _, e := range results {
		if e.FamousRank > 3 {
			t.Errorf("unfiltered Suggest should prefer famous, got rank %d for %s", e.FamousRank, e.Name)
		}
	}
}

func TestSuggestLimitAndSeed(t *testing.T) {
	a := Suggest(SuggestQuery{Vibe: "水", Limit: 4, Seed: 0})
	if len(a) > 4 {
		t.Fatalf("limit 4 exceeded")
	}
	b := Suggest(SuggestQuery{Vibe: "水", Limit: 4, Seed: 42})
	if len(b) > 4 {
		t.Fatalf("seeded limit exceeded")
	}
	// Max clamp
	big := Suggest(SuggestQuery{Limit: 100})
	if len(big) > 20 {
		t.Fatalf("max limit 20, got %d", len(big))
	}
}

func namesOf(entries []IndexEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}
	return out
}

func tonesOf(entries []IndexEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name + ":" + e.Tone
	}
	return out
}
