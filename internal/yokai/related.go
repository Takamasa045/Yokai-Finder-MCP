package yokai

import (
	"sort"
	"strings"
)

// Related returns yokai that share tags, category, or tone with the named entry.
func Related(name string, limit int) (IndexEntry, []IndexEntry, [][]string, error) {
	if limit <= 0 {
		limit = 6
	}
	if limit > 20 {
		limit = 20
	}
	origin, ok := LookupIndex(name)
	if !ok {
		if profile, found := LookupProfile(name); found {
			origin, ok = LookupIndex(profile.Name)
		}
	}
	if !ok {
		return IndexEntry{}, nil, nil, nil
	}

	type scored struct {
		entry  IndexEntry
		score  int
		shared []string
	}
	var hits []scored
	for _, candidate := range yokaiIndex {
		if candidate.Name == origin.Name {
			continue
		}
		shared := sharedTokens(origin, candidate)
		score := len(shared) * 8
		if candidate.Category == origin.Category {
			score += 12
			shared = appendUnique(shared, "category:"+candidate.Category)
		}
		if candidate.Tone == origin.Tone && candidate.Tone != "" {
			score += 6
			shared = appendUnique(shared, "tone:"+candidate.Tone)
		}
		if absInt(candidate.FamousRank-origin.FamousRank) <= 1 {
			score += 2
		}
		if score <= 0 {
			continue
		}
		hits = append(hits, scored{entry: candidate, score: score, shared: shared})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].entry.FamousRank < hits[j].entry.FamousRank
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	out := make([]IndexEntry, len(hits))
	shared := make([][]string, len(hits))
	for i, h := range hits {
		out[i] = copyIndexEntry(h.entry)
		shared[i] = h.shared
	}
	return origin, out, shared, nil
}

func sharedTokens(a, b IndexEntry) []string {
	var shared []string
	seen := map[string]struct{}{}
	for _, tag := range a.Tags {
		seen[strings.ToLower(tag)] = struct{}{}
	}
	for _, tag := range b.Tags {
		if _, ok := seen[strings.ToLower(tag)]; ok {
			shared = append(shared, tag)
		}
	}
	return shared
}

func appendUnique(list []string, value string) []string {
	for _, existing := range list {
		if existing == value {
			return list
		}
	}
	return append(list, value)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
