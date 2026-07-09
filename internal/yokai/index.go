package yokai

import "strings"

// IndexEntry is a lightweight yokai roster row for browsing.
// Full lore lives in curated Profile entries; books come from NDL search.
type IndexEntry struct {
	Name       string
	NativeName string
	Category   string
	Region     string
	BlurbJA    string
}

// Index returns a defensive copy of the yokai name index (111 entries).
func Index() []IndexEntry {
	out := make([]IndexEntry, len(yokaiIndex))
	copy(out, yokaiIndex)
	return out
}

// FilterIndex returns index entries matching term / category / region hints.
// Matching is case-insensitive substring search across name fields and blurb.
func FilterIndex(term, category, region string) []IndexEntry {
	term = strings.TrimSpace(strings.ToLower(term))
	category = strings.TrimSpace(strings.ToLower(category))
	region = strings.TrimSpace(strings.ToLower(region))

	var filtered []IndexEntry
	for _, entry := range yokaiIndex {
		if category != "" && !strings.Contains(strings.ToLower(entry.Category), category) {
			continue
		}
		if region != "" && !strings.Contains(strings.ToLower(entry.Region), region) {
			continue
		}
		if term != "" && !indexMatchesTerm(entry, term) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

// HasCuratedProfile reports whether a full bilingual Profile exists for this entry.
func (e IndexEntry) HasCuratedProfile() bool {
	if _, ok := FindByName(e.Name); ok {
		return true
	}
	if e.NativeName != "" {
		if _, ok := FindByName(e.NativeName); ok {
			return true
		}
	}
	return false
}

func indexMatchesTerm(entry IndexEntry, term string) bool {
	fields := []string{
		entry.Name,
		entry.NativeName,
		entry.Category,
		entry.Region,
		entry.BlurbJA,
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), term) {
			return true
		}
	}
	return false
}
