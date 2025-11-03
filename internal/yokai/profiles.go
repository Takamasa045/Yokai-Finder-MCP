package yokai

import (
	"math/rand"
	"strings"
	"time"
)

// Profile represents a curated yokai entry with lore and creative hooks.
type Profile struct {
	Name          string
	NativeName    string
	Region        string
	Category      string
	Summary       string
	Legends       []string
	Traits        []string
	Motifs        []string
	FunFact       string
	SearchQuery   string
	CreativeHooks []string
}

var curatedProfiles = []Profile{
	{
		Name:        "Kappa",
		NativeName:  "河童",
		Region:      "River valleys across Japan",
		Category:    "Water spirit",
		Summary:     "Mischievous, turtle-backed river dwellers that bargain with humans in exchange for cucumbers or favors.",
		Legends:     []string{"Challenges travellers to sumo bouts by the water's edge", "Known to return politeness with politeness, sometimes to their own downfall"},
		Traits:      []string{"Shell-backed", "Webbed hands", "Cucumber obsession"},
		Motifs:      []string{"Water bargains", "Etiquette", "Harvest rituals"},
		FunFact:     "If you bow deeply to a kappa, it will bow back and spill the life-giving water from its sara (head dish).",
		SearchQuery: "河童 妖怪 民話",
		CreativeHooks: []string{
			"Imagine a riverside festival game inspired by a kappa's playful pranks.",
			"Design a collaborative truce between local farmers and a curious kappa.",
		},
	},
	{
		Name:        "Tengu",
		NativeName:  "天狗",
		Region:      "Mountain shrines and deep cedar forests",
		Category:    "Mountain sentinel",
		Summary:     "Avian-nosed martial spirits who guard sacred mountains and challenge arrogant wanderers.",
		Legends:     []string{"Teaches swordplay to worthy disciples", "Kidnaps boastful priests to humble them"},
		Traits:      []string{"Command over wind gusts", "Master swordsmen", "Prideful guardians"},
		Motifs:      []string{"Discipline", "Humility", "Mountain austerity"},
		FunFact:     "Some medieval stories claim tengu invented the fan that can summon storms with a single flap.",
		SearchQuery: "天狗 修験道 伝承",
		CreativeHooks: []string{
			"Stage a duel where a tengu tests a traveller's resolve atop a precarious bridge.",
			"Create a secret tengu dojo hidden behind masked cedar trees.",
		},
	},
	{
		Name:        "Yuki-onna",
		NativeName:  "雪女",
		Region:      "Snowbound provinces of northern Honshu",
		Category:    "Snow phantom",
		Summary:     "Ethereal women cloaked in frost who appear during blizzards, balancing mercy and chill judgment.",
		Legends:     []string{"Spare kind-hearted woodcutters who share warmth", "Vanish at sunrise, leaving only drifting snowflakes"},
		Traits:      []string{"Breath colder than night", "Gliding footsteps", "Moonlit presence"},
		Motifs:      []string{"Impermanence", "Winter trials", "Frozen bargains"},
		FunFact:     "Some tales describe the Yuki-onna carrying a baby made of ice to test a traveller's compassion.",
		SearchQuery: "雪女 逸話 民俗学",
		CreativeHooks: []string{
			"Compose a scene where a Yuki-onna negotiates a fragile truce with spring itself.",
			"Illustrate snowflake sigils that mark the path to a Yuki-onna's hidden glade.",
		},
	},
	{
		Name:        "Kitsune",
		NativeName:  "九尾の狐",
		Region:      "Shrines and mist-laden plains",
		Category:    "Mystic fox",
		Summary:     "Shape-shifting fox spirits credited with trickery, wisdom, and occasionally saving entire provinces.",
		Legends:     []string{"Transforms into a dazzling performer to expose corruption", "Serves as messengers to the deity Inari"},
		Traits:      []string{"Illusory flames", "Multiple tails", "Silver-tongued"},
		Motifs:      []string{"Transformation", "Hidden intentions", "Sacred rice granaries"},
		FunFact:     "Folklore says each tail a kitsune grows grants a century of experience and sharpens its magic.",
		SearchQuery: "九尾狐 伝説 稲荷",
		CreativeHooks: []string{
			"Imagine a kitsune producing illusions to orchestrate a moonlit festival.",
			"Plot a political thriller where a kitsune advisor balances trickery with loyalty.",
		},
	},
	{
		Name:        "Bakeneko",
		NativeName:  "化け猫",
		Region:      "Castle towns and tea houses across Japan",
		Category:    "Shape-shifting beast",
		Summary:     "Once-domestic cats who, after long years, gain the uncanny ability to mimic human life and speech.",
		Legends:     []string{"Disguises itself as an innkeeper to run a midnight salon", "Dances with a towel wrapped on its head in the glow of lantern light"},
		Traits:      []string{"Forked tail", "Human-like speech", "Illusory shadows"},
		Motifs:      []string{"Transformation", "Domestic secrets", "Nighttime revelry"},
		FunFact:     "People once believed a cat's tail should be shortened to prevent it from becoming a bakeneko.",
		SearchQuery: "化け猫 怪談 江戸",
		CreativeHooks: []string{
			"Script a late-night tea ceremony hosted by a refined bakeneko.",
			"Invent a jazz club run by century-old cats trading gossip for secrets.",
		},
	},
	{
		Name:        "Nurikabe",
		NativeName:  "ぬりかべ",
		Region:      "Winding roads of Kyushu",
		Category:    "Trickster wall",
		Summary:     "Invisible wall yokai that halts travellers, forcing them to rethink their route or wait until dawn.",
		Legends:     []string{"Blocks the path of impatient samurai, yet lets lost children pass", "Said to fade if tapped with a stick near the ground"},
		Traits:      []string{"Immovable presence", "Invisible bulk", "Silence"},
		Motifs:      []string{"Patience", "Detours", "Hidden thresholds"},
		FunFact:     "Some Edo-period maps warn, tongue-in-cheek, of nurikabe hotspots to explain mysteriously long journeys.",
		SearchQuery: "ぬりかべ 伝説 道中",
		CreativeHooks: []string{
			"Design an escape-room puzzle where the walls are mischievous nurikabe spirits.",
			"Write a travelogue that becomes an allegory about embracing detours.",
		},
	},
	{
		Name:        "Rokurokubi",
		NativeName:  "ろくろ首",
		Region:      "Edo-period inns and waystations",
		Category:    "Nightly wanderer",
		Summary:     "By day they appear human; by night their necks stretch impossibly long, roaming halls in search of stories.",
		Legends:     []string{"Eavesdrops on sleeping guests to devour their whispered secrets", "Falls for a traveller who notices her dual life"},
		Traits:      []string{"Extendable neck", "Hypnotic gaze", "Nocturnal curiosity"},
		Motifs:      []string{"Hidden identities", "Nightly vigilance", "Story gathering"},
		FunFact:     "Old tales describe rokurokubi sending their heads to scout ahead while the body continues household chores.",
		SearchQuery: "ろくろ首 怪異 怪談",
		CreativeHooks: []string{
			"Compose a mystery solved by a rokurokubi who hears clues in the rafters.",
			"Storyboard a comic where elongated necks become a whimsical superpower.",
		},
	},
	{
		Name:        "Umi-bozu",
		NativeName:  "海坊主",
		Region:      "Storm-swept coasts of western Japan",
		Category:    "Sea apparition",
		Summary:     "Enigmatic, towering figures that rise from turbulent seas, tipping ships unless appeased with lantern light.",
		Legends:     []string{"Silences entire crews before crushing their masts", "Seeks offerings of oil to calm the waves"},
		Traits:      []string{"Towering silhouette", "Ink-black skin", "Tempest caller"},
		Motifs:      []string{"Respect for the sea", "Offerings", "Sailors' superstition"},
		FunFact:     "Fisherfolk once feared whispering the umi-bozu's name, believing it summoned waves tall enough to swallow lighthouses.",
		SearchQuery: "海坊主 海難 民話",
		CreativeHooks: []string{
			"Plot a maritime negotiation between lighthouse keepers and a lonely umi-bozu.",
			"Create ambient sea shanties that keep the umi-bozu lulled to sleep.",
		},
	},
	{
		Name:        "Gashadokuro",
		NativeName:  "餓者髑髏",
		Region:      "Battlefields and famine-struck villages",
		Category:    "Colossal skeleton",
		Summary:     "Gigantic skeleton spirits formed from the restless dead, stalking the countryside until pacified with ritual bells.",
		Legends:     []string{"Materializes at midnight to devour lone travellers", "Announced by a ringing in the ears before it strikes"},
		Traits:      []string{"Skull taller than treetops", "Bone-rattling footsteps", "Relentless hunger"},
		Motifs:      []string{"Reckoning", "Remembrance", "Famine tales"},
		FunFact:     "Some modern retellings convert the gashadokuro into gentle guardians once their bones are respectfully enshrined.",
		SearchQuery: "餓者髑髏 怪物 伝説",
		CreativeHooks: []string{
			"Write a requiem that persuades a gashadokuro to become a village protector.",
			"Design glowing talismans that disrupt the skeleton's eerie ear-ringing omen.",
		},
	},
	{
		Name:        "Amabie",
		NativeName:  "アマビエ",
		Region:      "Kumamoto's coastal waters",
		Category:    "Prophetic being",
		Summary:     "A mermaid-like oracle who surfaces to warn of plagues and promises good harvests if its image is shared.",
		Legends:     []string{"Emerges shining from the sea to deliver prophecies", "Inspires villagers to sketch its likeness for protection"},
		Traits:      []string{"Beaked face", "Glowing scales", "Future sight"},
		Motifs:      []string{"Hope", "Community art", "Healing"},
		FunFact:     "During modern outbreaks, Japanese artists revived the tradition of drawing Amabie as a charm against illness.",
		SearchQuery: "アマビエ 予言 疫病",
		CreativeHooks: []string{
			"Curate a community mural project led by Amabie fans across eras.",
			"Imagine Amabie livestreaming predictions to lighthouse keepers.",
		},
	},
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

// RandomProfile picks a profile from the provided slice using the seed for determinism.
func RandomProfile(seed int64, candidates []Profile) Profile {
	list := candidates
	if len(list) == 0 {
		list = curatedProfiles
	}
	if len(list) == 0 {
		return Profile{}
	}

	if seed == 0 {
		seed = time.Now().UnixNano()
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
