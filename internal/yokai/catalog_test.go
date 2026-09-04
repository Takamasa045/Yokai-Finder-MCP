package yokai

import "testing"

func TestEmbeddedCatalogSizes(t *testing.T) {
	if got := len(Index()); got < 160 {
		t.Fatalf("index too small: %d", got)
	}
	if got := len(Profiles()); got < 83 {
		t.Fatalf("profiles too small: %d", got)
	}
	if len(ExtraAliases()) < 50 {
		t.Fatalf("aliases too small: %d", len(ExtraAliases()))
	}
}
