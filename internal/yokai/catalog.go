package yokai

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed data/index.json data/profiles.json data/aliases.json
var catalogFS embed.FS

func init() {
	if err := loadEmbeddedCatalog(); err != nil {
		panic(fmt.Sprintf("load yokai catalog: %v", err))
	}
}

func loadEmbeddedCatalog() error {
	index, err := readJSON[[]IndexEntry]("data/index.json")
	if err != nil {
		return err
	}
	profiles, err := readJSON[[]Profile]("data/profiles.json")
	if err != nil {
		return err
	}
	aliases, err := readJSON[map[string]string]("data/aliases.json")
	if err != nil {
		return err
	}
	if len(index) == 0 || len(profiles) == 0 {
		return fmt.Errorf("catalog is empty (index=%d profiles=%d)", len(index), len(profiles))
	}
	yokaiIndex = index
	curatedProfiles = profiles
	extraAliases = aliases
	return validateAliasTargets()
}

func validateAliasTargets() error {
	for alias, target := range extraAliases {
		if NormalizeQuery(alias) == "" || NormalizeQuery(target) == "" {
			return fmt.Errorf("blank alias mapping %q -> %q", alias, target)
		}
		found := false
		want := NormalizeQuery(target)
		for _, entry := range yokaiIndex {
			if NormalizeQuery(entry.Name) == want || NormalizeQuery(entry.NativeName) == want {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("alias %q points at unknown yokai %q", alias, target)
		}
	}
	return nil
}

func readJSON[T any](name string) (T, error) {
	var out T
	raw, err := catalogFS.ReadFile(name)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}
