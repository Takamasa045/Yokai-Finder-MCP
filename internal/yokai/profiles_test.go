package yokai

import "testing"

func TestFindByName(t *testing.T) {
	profile, ok := FindByName("kappa")
	if !ok {
		t.Fatalf("expected to find kappa profile")
	}
	if profile.Name != "Kappa" {
		t.Fatalf("unexpected name: %s", profile.Name)
	}
	if profile.NativeName != "河童" {
		t.Fatalf("unexpected native name: %s", profile.NativeName)
	}

	profile, ok = FindByName("天狗")
	if !ok || profile.Name != "Tengu" {
		t.Fatalf("expected to resolve native name lookup for 天狗")
	}
}

func TestFilter(t *testing.T) {
	water := Filter("water", "")
	if len(water) == 0 {
		t.Fatalf("expected water category to return results")
	}

	foundKappa := false
	for _, p := range water {
		if p.Name == "Kappa" {
			foundKappa = true
			break
		}
	}
	if !foundKappa {
		t.Fatalf("filtered list should include Kappa")
	}

	region := Filter("", "kumamoto")
	if len(region) == 0 {
		t.Fatalf("expected region filter to return results")
	}
	if region[0].Name != "Amabie" {
		t.Fatalf("expected Kumamoto region to prioritise Amabie, got %s", region[0].Name)
	}
}

func TestRandomProfileDeterministic(t *testing.T) {
	candidates := Filter("water", "")
	if len(candidates) == 0 {
		t.Fatalf("expected candidates for water category")
	}

	first := RandomProfile(99, candidates)
	second := RandomProfile(99, candidates)

	if first.Name != second.Name {
		t.Fatalf("expected deterministic random selection, got %s vs %s", first.Name, second.Name)
	}
}
