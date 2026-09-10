package yokai

import (
	"strings"
	"testing"
	"time"
)

func TestPriorityRankThreeHaveProfiles(t *testing.T) {
	for _, name := range []string{
		"山彦", "磯撫", "大百足", "火車", "枕返し", "天井嘗", "一つ目入道",
		"牛鬼", "疫病神", "件", "アマビコ", "白澤", "骨女", "累", "管狐",
		"二口女", "手の目", "山童", "鉄鼠", "ぬっぺふほふ", "赤いマント", "きさらぎ駅",
		"貉", "お歯黒べったり", "八百比丘尼", "烏天狗", "白蔵主", "七人みさき", "逆柱",
		"大天狗", "白蛇", "安達ヶ原", "団三郎狸", "海座頭", "わいら",
		"鳳凰", "天逆毎", "夜雀", "隠れ里", "金神",
		"殺生石", "百鬼夜行", "人柱", "龍", "前鬼", "後鬼",
		"雨女", "雪ん子",
		"神隠し", "平将門", "菅原道真", "柳の精",
		"式神",
		"ツチノコ", "すねこすり", "キジムナー", "こっくりさん", "さざえ鬼",
		"長壁姫", "小袖の手", "物の怪", "屋島の禿", "百目",
		"白うねり", "籠釣瓶", "首切れ馬", "蛤女房", "お菊虫",
		"般若", "野寺坊", "以津真天", "大鯰", "置行堀",
	} {
		if _, ok := FindByName(name); !ok {
			t.Errorf("expected profile for %s", name)
		}
	}
}

func TestFamousRankOneTwoHaveProfiles(t *testing.T) {
	for _, e := range Index() {
		if e.FamousRank <= 2 && !e.HasCuratedProfile() {
			t.Errorf("%s (%s) famousRank %d is missing a profile", e.Name, e.NativeName, e.FamousRank)
		}
	}
}

func TestFindByName(t *testing.T) {
	profile, ok := FindByName("kappa")
	if !ok {
		t.Fatalf("expected to find kappa profile")
	}
	if profile.Name != "Kappa" {
		t.Fatalf("unexpected name: %s", profile.Name)
	}
	if profile.NativeName != "河童" {
		t.Fatalf("unexpected native name: %s", profile.NativeName)
	}

	profile, ok = FindByName("天狗")
	if !ok || profile.Name != "Tengu" {
		t.Fatalf("expected to resolve native name lookup for 天狗")
	}
}

func TestFilter(t *testing.T) {
	water := Filter("water", "")
	if len(water) == 0 {
		t.Fatalf("expected water category to return results")
	}

	foundKappa := false
	for _, p := range water {
		if p.Name == "Kappa" {
			foundKappa = true
			break
		}
	}
	if !foundKappa {
		t.Fatalf("filtered list should include Kappa")
	}

	region := Filter("", "kumamoto")
	if len(region) == 0 {
		t.Fatalf("expected region filter to return results")
	}
	if region[0].Name != "Amabie" {
		t.Fatalf("expected Kumamoto region to prioritise Amabie, got %s", region[0].Name)
	}
}

func TestProfilesHaveJapaneseLore(t *testing.T) {
	for _, p := range Profiles() {
		if strings.TrimSpace(p.SummaryJA) == "" {
			t.Errorf("%s is missing SummaryJA", p.Name)
		}
		if strings.TrimSpace(p.FunFactJA) == "" {
			t.Errorf("%s is missing FunFactJA", p.Name)
		}
	}
}

func TestFamousRankOneHasJapaneseLegends(t *testing.T) {
	for _, e := range Index() {
		if e.FamousRank != 1 {
			continue
		}
		p, ok := FindByName(e.NativeName)
		if !ok {
			t.Errorf("%s missing profile", e.NativeName)
			continue
		}
		if len(p.LegendsJA) < 3 {
			t.Errorf("%s rank-1 legendsJa %d want >= 3", e.NativeName, len(p.LegendsJA))
		}
		if len([]rune(strings.TrimSpace(p.SummaryJA))) < 40 {
			t.Errorf("%s rank-1 summaryJa too short (%d runes)", e.NativeName, len([]rune(p.SummaryJA)))
		}
		for _, line := range p.LegendsJA {
			if hit := CatalogArtifact(line); hit != "" {
				t.Errorf("%s legendsJa residue %q in %q", e.NativeName, hit, line)
			}
		}
	}
}

func TestProfilesHaveMinimumThickness(t *testing.T) {
	for _, p := range Profiles() {
		label := p.NativeName
		if len(p.Legends) < 3 {
			t.Errorf("%s legends %d want >= 3", label, len(p.Legends))
		}
		if len(p.Traits) < 4 {
			t.Errorf("%s traits %d want >= 4", label, len(p.Traits))
		}
		if len(p.Motifs) < 4 {
			t.Errorf("%s motifs %d want >= 4", label, len(p.Motifs))
		}
		if len(p.CreativeHooks) < 3 {
			t.Errorf("%s creativeHooks %d want >= 3", label, len(p.CreativeHooks))
		}
		if len(p.Sources) == 0 {
			t.Errorf("%s missing sources", label)
		}
		for _, m := range p.Motifs {
			if m == "index" {
				t.Errorf("%s still has filler motif %q", label, m)
			}
		}
	}
}

func TestCuratedRosterIncludesNewYokai(t *testing.T) {
	if got := len(Profiles()); got < 98 {
		t.Fatalf("expected at least 98 curated yokai, got %d", got)
	}

	for _, name := range []string{"座敷童子", "鵺", "鎌鼬", "一反木綿", "天邪鬼", "のっぺらぼう", "酒呑童子", "玉藻前", "口裂け女", "付喪神"} {
		if _, ok := FindByName(name); !ok {
			t.Errorf("expected curated profile for %s", name)
		}
	}
}

func TestDailySeed(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	seed := DailySeed(time.Date(2026, 7, 9, 23, 59, 0, 0, jst))
	if seed != 20260709 {
		t.Fatalf("expected daily seed 20260709, got %d", seed)
	}

	// 15:00 UTC on 9 July is already 00:00 JST on 10 July.
	utc := time.Date(2026, 7, 9, 15, 0, 0, 0, time.UTC)
	if got := DailySeed(utc); got != 20260710 {
		t.Fatalf("expected UTC evening to roll to next JST day, got %d", got)
	}

	morning := DailySeed(time.Date(2026, 7, 9, 0, 0, 1, 0, jst))
	if morning != seed {
		t.Fatalf("expected the same seed all day, got %d vs %d", morning, seed)
	}

	nextDay := DailySeed(time.Date(2026, 7, 10, 0, 0, 0, 0, jst))
	if nextDay == seed {
		t.Fatalf("expected a different seed on the next day")
	}
}

func TestRandomProfileZeroSeedIsStableWithinDay(t *testing.T) {
	first := RandomProfile(0, nil)
	second := RandomProfile(0, nil)
	if first.Name != second.Name {
		t.Fatalf("expected zero seed to yield today's stable pick, got %s vs %s", first.Name, second.Name)
	}

	expected := RandomProfile(DailySeed(time.Now()), nil)
	if first.Name != expected.Name {
		t.Fatalf("expected zero seed to match DailySeed(now) pick, got %s vs %s", first.Name, expected.Name)
	}
}

func TestRandomProfileDeterministic(t *testing.T) {
	candidates := Filter("water", "")
	if len(candidates) == 0 {
		t.Fatalf("expected candidates for water category")
	}

	first := RandomProfile(99, candidates)
	second := RandomProfile(99, candidates)

	if first.Name != second.Name {
		t.Fatalf("expected deterministic random selection, got %s vs %s", first.Name, second.Name)
	}
}
