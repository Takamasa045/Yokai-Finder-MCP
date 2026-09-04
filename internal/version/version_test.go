package version

import "testing"

func TestVersionIsSemverish(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
	if Version[0] < '0' || Version[0] > '9' {
		t.Fatalf("Version should start with a digit, got %q", Version)
	}
}
