package yokai

import (
	"math/rand"
	"strings"
	"time"
)

// Profile represents a curated yokai entry with lore and creative hooks.
// The curated entries themselves live in profiles_data.go.
type Profile struct {
	Name          string
	NativeName    string
	Region        string
	Category      string
	Summary       string
	SummaryJA     string
	Legends       []string
	Traits        []string
	Motifs        []string
	FunFact       string
	FunFactJA     string
	SearchQuery   string
	CreativeHooks []string
}

// Profiles returns a defensive copy of the curated profile list.
func Profiles() []Profile {
	out := make([]Profile, len(curatedProfiles))
	copy(out, curatedProfiles)
	return out
}

// FindByName looks up a profile by English or native name (case-insensitive for English).
func FindByName(name string) (Profile, bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return Profile{}, false
	}

	for _, profile := range curatedProfiles {
		if strings.EqualFold(profile.Name, trimmed) || profile.NativeName == trimmed {
			return profile, true
		}
	}
	return Profile{}, false
}

// Filter returns profiles matching the provided category and region hints.
func Filter(category, region string) []Profile {
	category = strings.TrimSpace(strings.ToLower(category))
	region = strings.TrimSpace(strings.ToLower(region))

	var filtered []Profile
	for _, profile := range curatedProfiles {
		if matchesHint(profile.Category, category) && matchesHint(profile.Region, region) {
			filtered = append(filtered, profile)
		}
	}
	return filtered
}

// DailySeed derives a deterministic seed from the calendar date so the
// "yokai of the day" stays the same until midnight in the given location.
func DailySeed(t time.Time) int64 {
	return int64(t.Year()*10000 + int(t.Month())*100 + t.Day())
}

// RandomProfile picks a profile from the provided slice using the seed for
// determinism. A zero seed falls back to today's DailySeed, making the pick
// stable for the whole day.
func RandomProfile(seed int64, candidates []Profile) Profile {
	list := candidates
	if len(list) == 0 {
		list = curatedProfiles
	}
	if len(list) == 0 {
		return Profile{}
	}

	if seed == 0 {
		seed = DailySeed(time.Now())
	}
	r := rand.New(rand.NewSource(seed))
	return list[r.Intn(len(list))]
}

func matchesHint(value string, hint string) bool {
	if hint == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), hint)
}
