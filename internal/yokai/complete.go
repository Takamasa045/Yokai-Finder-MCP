package yokai

import "strings"

const maxCompletions = 20

// CompleteNames returns native names (and English names when the prefix is
// ASCII) that start with the given prefix, for MCP completion/complete.
func CompleteNames(prefix string, limit int) []string {
	if limit <= 0 || limit > maxCompletions {
		limit = maxCompletions
	}
	prefix = strings.TrimSpace(prefix)
	norm := NormalizeQuery(prefix)
	seen := map[string]struct{}{}
	var out []string

	add := func(value string) {
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	for _, entry := range yokaiIndex {
		if len(out) >= limit {
			break
		}
		if prefix == "" {
			add(entry.NativeName)
			continue
		}
		if strings.HasPrefix(entry.NativeName, prefix) || (norm != "" && strings.HasPrefix(NormalizeQuery(entry.NativeName), norm)) {
			add(entry.NativeName)
			continue
		}
		if strings.HasPrefix(strings.ToLower(entry.Name), strings.ToLower(prefix)) {
			add(entry.NativeName)
			continue
		}
		for _, alias := range entry.Aliases {
			if strings.HasPrefix(alias, prefix) || (norm != "" && strings.HasPrefix(NormalizeQuery(alias), norm)) {
				add(entry.NativeName)
				break
			}
		}
	}

	if prefix == "" {
		return out
	}
	// Second pass: extraAliases keys
	if len(out) < limit {
		for alias, target := range extraAliases {
			if len(out) >= limit {
				break
			}
			if strings.HasPrefix(alias, prefix) || (norm != "" && strings.HasPrefix(NormalizeQuery(alias), norm)) {
				if entry, ok := LookupIndex(target); ok {
					add(entry.NativeName)
				}
			}
		}
	}
	return out
}

// CatalogInfo is a compact overview for the yokai://catalog resource.
type CatalogInfo struct {
	IndexCount   int            `json:"indexCount"`
	ProfileCount int            `json:"profileCount"`
	AliasCount   int            `json:"aliasCount"`
	Categories   map[string]int `json:"categories"`
	Tones        map[string]int `json:"tones"`
	FamousRank   map[int]int    `json:"famousRank"`
	ProfileShare float64        `json:"profileShare"`
}

func CatalogOverview() CatalogInfo {
	info := CatalogInfo{
		IndexCount:   len(yokaiIndex),
		ProfileCount: len(curatedProfiles),
		AliasCount:   len(extraAliases),
		Categories:   map[string]int{},
		Tones:        map[string]int{},
		FamousRank:   map[int]int{},
	}
	for _, e := range yokaiIndex {
		info.Categories[e.Category]++
		info.Tones[e.Tone]++
		info.FamousRank[e.FamousRank]++
	}
	if info.IndexCount > 0 {
		info.ProfileShare = float64(info.ProfileCount) / float64(info.IndexCount)
	}
	return info
}
