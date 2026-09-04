package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/yokai"
)

func main() {
	outDir := "internal/yokai/data"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}

	index := yokai.Index()
	profiles := yokai.Profiles()
	writeJSON(filepath.Join(outDir, "index.json"), index)
	writeJSON(filepath.Join(outDir, "profiles.json"), profiles)
	writeJSON(filepath.Join(outDir, "aliases.json"), yokai.ExtraAliases())
	fmt.Printf("wrote %d index rows, %d profiles to %s\n", len(index), len(profiles), outDir)
}

func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "export-catalog: %v\n", err)
	os.Exit(1)
}
