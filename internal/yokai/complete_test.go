package yokai

import "testing"

func TestCompleteNamesKappa(t *testing.T) {
	hits := CompleteNames("河", 10)
	found := false
	for _, h := range hits {
		if h == "河童" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 河童 in completions for 河, got %v", hits)
	}

	en := CompleteNames("Kap", 10)
	found = false
	for _, h := range en {
		if h == "河童" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 河童 for Kap, got %v", en)
	}
}

func TestCompleteNamesAlias(t *testing.T) {
	hits := CompleteNames("カッパ", 5)
	found := false
	for _, h := range hits {
		if h == "河童" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 河童 for カッパ, got %v", hits)
	}
}

func TestCatalogOverview(t *testing.T) {
	info := CatalogOverview()
	if info.IndexCount < 160 || info.ProfileCount < 98 {
		t.Fatalf("unexpected catalog size %+v", info)
	}
	if info.Categories["水系"] == 0 {
		t.Fatalf("expected 水系 category count, got %v", info.Categories)
	}
}

func TestProfileCategoryIsCanonicalJapanese(t *testing.T) {
	p, ok := FindByName("河童")
	if !ok {
		t.Fatal("missing kappa")
	}
	if p.Category != "水系" {
		t.Fatalf("category = %q, want 水系", p.Category)
	}
	if p.CategoryEN == "" {
		t.Fatal("expected categoryEn")
	}
}
