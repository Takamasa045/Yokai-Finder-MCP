package yokai

import "testing"

func TestRelatedKappa(t *testing.T) {
	origin, neighbours, shared, err := Related("カッパ", 5)
	if err != nil {
		t.Fatalf("Related: %v", err)
	}
	if origin.NativeName != "河童" {
		t.Fatalf("origin = %s", origin.NativeName)
	}
	if len(neighbours) == 0 {
		t.Fatal("expected neighbours")
	}
	if len(shared) != len(neighbours) {
		t.Fatalf("shared rows %d != neighbours %d", len(shared), len(neighbours))
	}
}
