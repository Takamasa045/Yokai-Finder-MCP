package version

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionIsSemverish(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
	if Version[0] < '0' || Version[0] > '9' {
		t.Fatalf("Version should start with a digit, got %q", Version)
	}
}

func TestPublishedFilesPinTheSameVersion(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, name := range []string{"README.md", "mcp.json", "server.json"} {
		body, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(body), Version) {
			t.Errorf("%s should mention version %s", name, Version)
		}
	}
}
