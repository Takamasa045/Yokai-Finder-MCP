package yokai

import (
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

type nameIndex struct {
	once        sync.Once
	exact       map[string]int // normalized query -> index offset
	profileBy   map[string]int // normalized query -> profile offset
	profileSet  map[string]struct{}
	displayKeys []string
}

var catalogIndex nameIndex

func (n *nameIndex) build() {
	n.once.Do(func() {
		n.exact = make(map[string]int, len(yokaiIndex)*4)
		n.profileBy = make(map[string]int, len(curatedProfiles)*4)
		n.profileSet = make(map[string]struct{}, len(curatedProfiles)*2)
		n.displayKeys = make([]string, 0, len(yokaiIndex))

		for i, e := range yokaiIndex {
			n.displayKeys = append(n.displayKeys, e.Name)
			n.addIndexKeys(i, e.Name, e.NativeName)
			for _, alias := range e.Aliases {
				n.addIndexKeys(i, alias)
			}
		}
		for alias, target := range extraAliases {
			if i, ok := n.lookupIndex(target); ok {
				n.addIndexKeys(i, alias)
			}
		}
		for i, p := range curatedProfiles {
			n.profileSet[NormalizeQuery(p.Name)] = struct{}{}
			n.profileSet[NormalizeQuery(p.NativeName)] = struct{}{}
			n.addProfileKeys(i, p.Name, p.NativeName)
		}
	})
}

func (n *nameIndex) addIndexKeys(i int, keys ...string) {
	for _, key := range keys {
		norm := NormalizeQuery(key)
		if norm == "" {
			continue
		}
		if _, exists := n.exact[norm]; !exists {
			n.exact[norm] = i
		}
	}
}

func (n *nameIndex) addProfileKeys(i int, keys ...string) {
	for _, key := range keys {
		norm := NormalizeQuery(key)
		if norm == "" {
			continue
		}
		if _, exists := n.profileBy[norm]; !exists {
			n.profileBy[norm] = i
		}
	}
}

func (n *nameIndex) lookupIndex(name string) (int, bool) {
	norm := NormalizeQuery(name)
	if norm == "" {
		return 0, false
	}
	if i, ok := n.exact[norm]; ok {
		return i, true
	}
	// extraAliases values may not yet be in exact during build
	for i, e := range yokaiIndex {
		if NormalizeQuery(e.Name) == norm || NormalizeQuery(e.NativeName) == norm {
			return i, true
		}
	}
	return 0, false
}

// LookupIndex finds an index entry by English name, native name, or alias.
func LookupIndex(name string) (IndexEntry, bool) {
	catalogIndex.build()
	norm := NormalizeQuery(name)
	if norm == "" {
		return IndexEntry{}, false
	}
	if i, ok := catalogIndex.exact[norm]; ok {
		return copyIndexEntry(yokaiIndex[i]), true
	}
	return IndexEntry{}, false
}

// LookupProfile finds a curated profile by English name, native name, or alias.
func LookupProfile(name string) (Profile, bool) {
	catalogIndex.build()
	norm := NormalizeQuery(name)
	if norm == "" {
		return Profile{}, false
	}
	if i, ok := catalogIndex.profileBy[norm]; ok {
		return curatedProfiles[i], true
	}
	if entry, ok := LookupIndex(name); ok {
		if i, ok := catalogIndex.profileBy[NormalizeQuery(entry.Name)]; ok {
			return curatedProfiles[i], true
		}
		if i, ok := catalogIndex.profileBy[NormalizeQuery(entry.NativeName)]; ok {
			return curatedProfiles[i], true
		}
	}
	return Profile{}, false
}

// SuggestNames returns close matches when an exact lookup fails.
func SuggestNames(name string, limit int) []IndexEntry {
	catalogIndex.build()
	if limit <= 0 {
		limit = 5
	}
	norm := NormalizeQuery(name)
	if norm == "" {
		return nil
	}

	type scored struct {
		entry IndexEntry
		score int
	}
	var hits []scored
	for _, e := range yokaiIndex {
		pool := NormalizeQuery(e.Name + e.NativeName + strings.Join(e.Aliases, ""))
		score := 0
		if strings.Contains(pool, norm) || strings.Contains(norm, NormalizeQuery(e.NativeName)) {
			score = 40
		}
		if isMostlyASCII(name) {
			en := NormalizeQuery(e.Name)
			if d := levenshtein(norm, en); d <= 2 && utf8.RuneCountInString(norm) >= 3 {
				score += 30 - d*8
			}
		}
		for _, alias := range e.Aliases {
			if NormalizeQuery(alias) == norm {
				score += 50
			}
		}
		if score > 0 {
			hits = append(hits, scored{entry: e, score: score + scoreFamous(e)})
		}
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
	for i, h := range hits {
		out[i] = copyIndexEntry(h.entry)
	}
	return out
}

func hasProfileName(name, native string) bool {
	catalogIndex.build()
	if _, ok := catalogIndex.profileSet[NormalizeQuery(name)]; ok {
		return true
	}
	_, ok := catalogIndex.profileSet[NormalizeQuery(native)]
	return ok
}

func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = minInt(del, ins, sub)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

func minInt(values ...int) int {
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
