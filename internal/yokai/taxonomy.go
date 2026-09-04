package yokai

import "strings"

// Canonical category keys used by the index. English and poetic profile
// labels map onto these so list_yokai and list_curated_yokai share a vocabulary.
var categoryAliases = map[string]string{
	"水系": "水系", "water": "水系", "waterspirit": "水系", "river": "水系", "kappa": "水系",
	"海": "海", "sea": "海", "ocean": "海", "seaapparition": "海", "shipghosts": "海", "merfolk": "海",
	"山系": "山系", "mountain": "山系", "mountainsentinel": "山系", "mountaincrone": "山系", "treespirit": "山系",
	"付喪神": "付喪神", "tsukumogami": "付喪神", "tool": "付喪神", "toolspirits": "付喪神", "tsukumogamilantern": "付喪神", "oneleggedumbrella": "付喪神",
	"狐狸": "狐狸", "fox": "狐狸", "kitsune": "狐狸", "tanuki": "狐狸", "mysticfox": "狐狸", "shapeshiftingtanuki": "狐狸", "ninetailedcourtesan": "狐狸",
	"変化": "変化", "shapeshift": "変化", "shapeshiftingbeast": "変化", "forktailedcat": "変化",
	"屋敷": "屋敷", "house": "屋敷", "housespirit": "屋敷", "filthlicker": "屋敷", "eyesintheshoji": "屋敷",
	"死霊": "死霊", "ghost": "死霊", "spirit": "死霊", "birthspirit": "死霊", "colossalskeleton": "死霊",
	"気象": "気象", "weather": "気象", "snow": "気象", "snowphantom": "気象", "windbeast": "気象", "thunderbeast": "気象",
	"鬼": "鬼", "oni": "鬼", "demon": "鬼", "horneddemon": "鬼", "demonlord": "鬼", "onilieutenant": "鬼", "contrarianimp": "鬼",
	"現代伝承": "現代伝承", "urban": "現代伝承", "urbanlegend": "現代伝承", "modern": "現代伝承", "slitmouthedwoman": "現代伝承",
	"古典": "古典", "classic": "古典", "classical": "古典", "earthspider": "古典", "serpentofobsession": "古典", "bridgeprincess": "古典", "bluelanternspectre": "古典",
	"異形": "異形", "strange": "異形", "chimeraomen": "異形", "facelessphantom": "異形", "slipperyoverlord": "異形", "lookovermonk": "異形",
	"道中": "道中", "road": "道中", "tricksterwall": "道中", "nightlywanderer": "道中", "wheelmonk": "道中",
	"町": "町", "town": "町", "tofuboy": "町",
	"田畑": "田畑", "field": "田畑", "muddyfieldwraith": "田畑",
	"霊火": "霊火", "fire": "霊火",
	"憑きもの": "憑きもの", "possession": "憑きもの", "dogspiritpossession": "憑きもの",
	"予言": "予言", "prophecy": "予言", "propheticbeing": "予言",
}

// ValidTones are the allowed index tone values.
var ValidTones = map[string]struct{}{
	"gentle": {}, "comic": {}, "horror": {},
	"solemn": {}, "tragic": {}, "mysterious": {}, "playful": {},
}

func canonicalCategorySet() map[string]struct{} {
	out := make(map[string]struct{}, len(categoryAliases))
	for _, canon := range categoryAliases {
		out[canon] = struct{}{}
	}
	return out
}

var regionAliases = map[string]string{
	"東北": "東北", "tohoku": "東北",
	"京都": "京都", "kyoto": "京都",
	"九州": "九州", "kyushu": "九州",
	"熊本": "熊本", "kumamoto": "熊本",
	"江戸": "江戸", "edo": "江戸", "tokyo": "江戸",
	"全国": "全国", "japan": "全国",
	"海": "海", "sea": "海", "coast": "海",
	"山": "山", "mountain": "山",
}

// CanonicalCategory maps a free-text category hint onto an index key.
// Unknown values are returned as a normalized string for substring matching.
func CanonicalCategory(hint string) string {
	key := NormalizeQuery(hint)
	if key == "" {
		return ""
	}
	if canon, ok := categoryAliases[key]; ok {
		return canon
	}
	if canon, ok := categoryAliases[strings.ToLower(strings.TrimSpace(hint))]; ok {
		return canon
	}
	return key
}

func CanonicalRegion(hint string) string {
	key := NormalizeQuery(hint)
	if key == "" {
		return ""
	}
	if canon, ok := regionAliases[key]; ok {
		return canon
	}
	return strings.ToLower(strings.TrimSpace(hint))
}

func categoryMatches(value, hint string) bool {
	if strings.TrimSpace(hint) == "" {
		return true
	}
	canon := CanonicalCategory(hint)
	if canon != "" && (value == canon || CanonicalCategory(value) == canon) {
		return true
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(strings.TrimSpace(hint))) ||
		strings.Contains(NormalizeQuery(value), NormalizeQuery(hint))
}

var prefectureToRegion = map[string]string{
	"青森": "東北", "岩手": "東北", "宮城": "東北", "秋田": "東北", "山形": "東北", "福島": "東北",
	"茨城": "関東", "栃木": "関東", "群馬": "関東", "埼玉": "関東", "千葉": "関東", "東京": "関東", "江戸": "関東", "神奈川": "関東",
	"新潟": "北陸", "富山": "北陸", "石川": "北陸", "福井": "北陸",
	"山梨": "甲信", "長野": "甲信",
	"岐阜": "東海", "静岡": "東海", "愛知": "東海", "三重": "東海",
	"滋賀": "関西", "京都": "関西", "大阪": "関西", "兵庫": "関西", "奈良": "関西", "和歌山": "関西",
	"鳥取": "中国", "島根": "中国", "岡山": "中国", "広島": "中国", "山口": "中国",
	"徳島": "四国", "香川": "四国", "愛媛": "四国", "高知": "四国",
	"福岡": "九州", "佐賀": "九州", "長崎": "九州", "熊本": "九州", "大分": "九州", "宮崎": "九州", "鹿児島": "九州", "沖縄": "九州",
}

func regionMatches(value, hint string) bool {
	if strings.TrimSpace(hint) == "" {
		return true
	}
	canon := CanonicalRegion(hint)
	lower := strings.ToLower(value)
	if canon != "" && (strings.Contains(value, canon) || strings.Contains(lower, strings.ToLower(canon))) {
		return true
	}
	for pref, region := range prefectureToRegion {
		if strings.Contains(value, pref) && (canon == region || strings.Contains(strings.ToLower(hint), strings.ToLower(region))) {
			return true
		}
	}
	return strings.Contains(lower, strings.ToLower(strings.TrimSpace(hint))) ||
		strings.Contains(NormalizeQuery(value), NormalizeQuery(hint))
}
