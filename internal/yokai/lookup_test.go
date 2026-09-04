package yokai

import "testing"

func TestNormalizeQueryKanaAndHyphen(t *testing.T) {
	if NormalizeQuery("カッパ") != NormalizeQuery("かっぱ") {
		t.Fatalf("katakana and hiragana should fold: %q vs %q", NormalizeQuery("カッパ"), NormalizeQuery("かっぱ"))
	}
	if NormalizeQuery("Yuki-onna") != NormalizeQuery("yuki onna") {
		t.Fatalf("hyphen/space should fold")
	}
}

func TestLookupIndexAliases(t *testing.T) {
	for _, q := range []string{"河童", "Kappa", "kappa", "カッパ", "かっぱ", "かわっぱ"} {
		e, ok := LookupIndex(q)
		if !ok {
			t.Fatalf("LookupIndex(%q) missed", q)
		}
		if e.NativeName != "河童" {
			t.Fatalf("LookupIndex(%q) = %s, want 河童", q, e.NativeName)
		}
	}
}

func TestLookupProfileViaAlias(t *testing.T) {
	p, ok := LookupProfile("カッパ")
	if !ok {
		t.Fatal("expected profile for カッパ")
	}
	if p.Name != "Kappa" {
		t.Fatalf("got %s", p.Name)
	}
}

func TestCanonicalCategoryUnifiesEnglishAndJapanese(t *testing.T) {
	if CanonicalCategory("water") != "水系" {
		t.Fatalf("water -> %s", CanonicalCategory("water"))
	}
	if CanonicalCategory("水系") != "水系" {
		t.Fatalf("水系 -> %s", CanonicalCategory("水系"))
	}
	if CanonicalCategory("Water spirit") != "水系" {
		t.Fatalf("Water spirit -> %s", CanonicalCategory("Water spirit"))
	}
}

func TestLookupMiss(t *testing.T) {
	if _, ok := LookupIndex("zzzz-nope"); ok {
		t.Fatal("expected miss")
	}
	if _, ok := LookupProfile(""); ok {
		t.Fatal("empty should miss")
	}
}

func TestCanonicalRegion(t *testing.T) {
	if CanonicalRegion("tohoku") != "東北" {
		t.Fatalf("tohoku -> %s", CanonicalRegion("tohoku"))
	}
	if CanonicalRegion("Kumamoto") != "熊本" {
		t.Fatalf("Kumamoto -> %s", CanonicalRegion("Kumamoto"))
	}
}

func TestFilterIndexTagToneRank(t *testing.T) {
	yes := true
	hits := FilterIndexOpts(IndexFilter{
		Tag:           "水",
		Tone:          "playful",
		FamousRankMax: 2,
		HasProfile:    &yes,
	})
	if len(hits) == 0 {
		t.Fatal("expected filtered hits")
	}
	for _, e := range hits {
		if e.Tone != "playful" {
			t.Fatalf("tone %s", e.Tone)
		}
		if e.FamousRank > 2 {
			t.Fatalf("rank %d", e.FamousRank)
		}
		if !e.HasCuratedProfile() {
			t.Fatalf("%s should have profile", e.Name)
		}
	}
}

func TestRelatedUnknown(t *testing.T) {
	origin, neighbours, _, err := Related("存在しないxyz", 3)
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	if origin.Name != "" || len(neighbours) != 0 {
		t.Fatalf("expected empty, got %+v %d", origin, len(neighbours))
	}
}

func TestLevenshteinEmpty(t *testing.T) {
	if levenshtein("", "ab") != 2 {
		t.Fatalf("empty a")
	}
	if levenshtein("ab", "") != 2 {
		t.Fatalf("empty b")
	}
}

func TestSuggestNamesForTypo(t *testing.T) {
	hits := SuggestNames("Kapa", 5)
	if len(hits) == 0 {
		t.Fatal("expected suggestions for Kapa")
	}
	found := false
	for _, h := range hits {
		if h.NativeName == "河童" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 河童 among suggestions, got %#v", namesOf(hits))
	}
}
