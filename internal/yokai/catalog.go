package yokai

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
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
	if err := validateLoadedCatalog(index, profiles, aliases); err != nil {
		return err
	}
	yokaiIndex = index
	curatedProfiles = profiles
	extraAliases = aliases
	return nil
}

func validateLoadedCatalog(index []IndexEntry, profiles []Profile, aliases map[string]string) error {
	if len(index) == 0 || len(profiles) == 0 {
		return fmt.Errorf("catalog is empty (index=%d profiles=%d)", len(index), len(profiles))
	}

	var errs []string
	knownCats := canonicalCategorySet()
	seenName := map[string]string{}
	seenNative := map[string]string{}
	seenNorm := map[string]string{}
	indexByName := map[string]IndexEntry{}

	note := func(format string, args ...any) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	for _, e := range index {
		label := strings.TrimSpace(e.NativeName)
		if label == "" {
			label = strings.TrimSpace(e.Name)
		}
		if strings.TrimSpace(e.Name) == "" {
			note("index entry missing name")
		}
		if strings.TrimSpace(e.NativeName) == "" {
			note("%s missing nativeName", e.Name)
		}
		if strings.TrimSpace(e.Category) == "" {
			note("%s missing category", label)
		} else if _, ok := knownCats[e.Category]; !ok {
			note("%s unknown category %q", label, e.Category)
		}
		if strings.TrimSpace(e.BlurbJA) == "" {
			note("%s missing blurbJa", label)
		}
		if len(e.Tags) < 3 || len(e.Tags) > 8 {
			note("%s tags count %d want 3–8", label, len(e.Tags))
		}
		if _, ok := ValidTones[e.Tone]; !ok {
			note("%s invalid tone %q", label, e.Tone)
		}
		if e.FamousRank < 1 || e.FamousRank > 5 {
			note("%s famousRank %d want 1–5", label, e.FamousRank)
		}
		if prev, ok := seenName[e.Name]; ok {
			note("duplicate name %q (also %s)", e.Name, prev)
		}
		seenName[e.Name] = label
		if prev, ok := seenNative[e.NativeName]; ok {
			note("duplicate nativeName %q (also %s)", e.NativeName, prev)
		}
		seenNative[e.NativeName] = label
		indexByName[NormalizeQuery(e.Name)] = e
		indexByName[NormalizeQuery(e.NativeName)] = e

		for _, key := range append([]string{e.Name, e.NativeName}, e.Aliases...) {
			norm := NormalizeQuery(key)
			if norm == "" {
				continue
			}
			if prev, ok := seenNorm[norm]; ok && prev != e.NativeName {
				note("normalized name %q maps to both %s and %s", norm, prev, e.NativeName)
			}
			seenNorm[norm] = e.NativeName
		}
	}

	seenProfile := map[string]string{}
	for _, p := range profiles {
		label := strings.TrimSpace(p.NativeName)
		if label == "" {
			label = strings.TrimSpace(p.Name)
		}
		if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.NativeName) == "" {
			note("profile missing name fields: %+v", p.Name)
		}
		if strings.TrimSpace(p.Summary) == "" || strings.TrimSpace(p.SummaryJA) == "" {
			note("%s missing summary / summaryJa", label)
		}
		if strings.TrimSpace(p.FunFact) == "" || strings.TrimSpace(p.FunFactJA) == "" {
			note("%s missing funFact / funFactJa", label)
		}
		if strings.TrimSpace(p.SearchQuery) == "" {
			note("%s missing searchQuery", label)
		}
		if len(p.Legends) == 0 {
			note("%s missing legends", label)
		}
		if p.Category != "" {
			if _, ok := knownCats[p.Category]; !ok {
				note("%s unknown profile category %q", label, p.Category)
			}
		}
		for _, key := range []string{p.Name, p.NativeName} {
			norm := NormalizeQuery(key)
			if norm == "" {
				continue
			}
			if prev, ok := seenProfile[norm]; ok && prev != p.NativeName {
				note("duplicate profile %q (also %s)", key, prev)
			}
			seenProfile[norm] = p.NativeName
		}
		en, byEn := indexByName[NormalizeQuery(p.Name)]
		ja, byJa := indexByName[NormalizeQuery(p.NativeName)]
		if !byEn || !byJa {
			note("profile %s (%s) is missing from the index", p.Name, p.NativeName)
		} else if en.NativeName != ja.NativeName || en.Name != ja.Name {
			note("profile %s (%s) maps to different index rows", p.Name, p.NativeName)
		}
	}

	for alias, target := range aliases {
		if NormalizeQuery(alias) == "" || NormalizeQuery(target) == "" {
			note("blank alias mapping %q -> %q", alias, target)
			continue
		}
		if _, ok := indexByName[NormalizeQuery(target)]; !ok {
			note("alias %q points at unknown yokai %q", alias, target)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("catalog: %s", strings.Join(errs, "; "))
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
