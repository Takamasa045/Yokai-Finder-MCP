package yokai

import (
	"math/rand"
	"time"
)

// Profile represents a curated yokai entry with lore and creative hooks.
// The curated entries themselves live in data/profiles.json.
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
	Sources       []string
}

// curatedProfiles is populated from the embedded catalog at init.
var curatedProfiles []Profile

// Profiles returns a defensive copy of the curated profile list.
func Profiles() []Profile {
	out := make([]Profile, len(curatedProfiles))
	copy(out, curatedProfiles)
	return out
}

// FindByName looks up a profile by English or native name (case-insensitive for English).
func FindByName(name string) (Profile, bool) {
	return LookupProfile(name)
}

// Filter returns profiles matching the provided category and region hints.
func Filter(category, region string) []Profile {
	var filtered []Profile
	for _, profile := range curatedProfiles {
		catHit := categoryMatches(profile.Category, category)
		regOK := regionMatches(profile.Region, region)
		if entry, ok := LookupIndex(profile.Name); ok {
			catHit = catHit || categoryMatches(entry.Category, category)
			regOK = regOK || regionMatches(entry.Region, region)
		}
		if !catHit || !regOK {
			continue
		}
		filtered = append(filtered, profile)
	}
	return filtered
}

// JST is Japan Standard Time, used for the daily yokai pick.
var JST = time.FixedZone("JST", 9*60*60)

// DailySeed derives a deterministic seed from the calendar date in JST so the
// "yokai of the day" stays the same until midnight in Japan, regardless of
// the host timezone.
func DailySeed(t time.Time) int64 {
	t = t.In(JST)
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
