package yokai

import "strings"

// RegionFacets is a structured reading of a free-text region label.
type RegionFacets struct {
	Prefectures []string
	Areas       []string
	Habitats    []string
}

var categoryHabitats = map[string][]string{
	"水系":   {"川", "沼", "水辺"},
	"海":    {"海", "海岸"},
	"山系":   {"山", "森"},
	"屋敷":   {"家", "屋敷"},
	"町":    {"町"},
	"道中":   {"道", "峠"},
	"田畑":   {"田", "畑"},
	"霊火":   {"夜"},
	"現代伝承": {"都市", "学校"},
	"気象":   {"空"},
	"鬼":    {"山"},
	"狐狸":   {"里"},
	"付喪神":  {"家"},
}

// Well-known entries whose display region omits a prefecture.
var yokaiPrefectures = map[string][]string{
	"なまはげ":      {"秋田"},
	"座敷童子":      {"岩手"},
	"アマビエ":      {"熊本"},
	"鵺":         {"京都"},
	"キジムナー":     {"沖縄"},
	"置行堀":       {"東京"},
	"清姫":        {"和歌山"},
	"安達ヶ原":      {"福島"},
	"酒呑童子":      {"京都"},
	"おしら様":      {"岩手"},
	"ひょうすべ":     {"佐賀"},
	"金長狸":       {"愛媛"},
	"団三郎狸":      {"新潟"},
	"隠神刑部":      {"香川"},
	"芝右衛門狸":     {"兵庫"},
	"殺生石":       {"栃木"},
	"夜泣石":       {"静岡"},
	"ガタロ":       {"福岡"},
	"アイヌカイセイ":   {"北海道"},
	"アツゥイコロエカシ": {"北海道"},
}

var placePrefecture = map[string]string{
	"北海道": "北海道", "アイヌ": "北海道", "蝦夷": "北海道",
	"青森": "青森", "津軽": "青森", "陸奥": "青森",
	"岩手": "岩手", "遠野": "岩手", "南部": "岩手",
	"宮城": "宮城", "仙台": "宮城",
	"秋田": "秋田",
	"山形": "山形",
	"福島": "福島", "会津": "福島", "安達": "福島",
	"茨城": "茨城",
	"栃木": "栃木", "那須": "栃木",
	"群馬": "群馬",
	"埼玉": "埼玉",
	"千葉": "千葉",
	"東京": "東京", "江戸": "東京", "本所": "東京", "池袋": "東京",
	"神奈川": "神奈川", "鎌倉": "神奈川",
	"新潟": "新潟", "越後": "新潟", "佐渡": "新潟",
	"富山": "富山",
	"石川": "石川", "加賀": "石川",
	"福井": "福井",
	"山梨": "山梨", "甲斐": "山梨",
	"長野": "長野", "信濃": "長野",
	"岐阜": "岐阜",
	"静岡": "静岡", "駿河": "静岡",
	"愛知": "愛知", "尾張": "愛知", "三河": "愛知",
	"三重": "三重", "伊勢": "三重",
	"滋賀": "滋賀", "近江": "滋賀",
	"京都": "京都", "大江山": "京都", "丹波": "京都",
	"大阪": "大阪", "摂津": "大阪", "河内": "大阪", "和泉": "大阪",
	"兵庫": "兵庫", "播磨": "兵庫", "淡路": "兵庫",
	"奈良": "奈良", "吉野": "奈良",
	"和歌山": "和歌山", "紀伊": "和歌山", "高野": "和歌山", "道成寺": "和歌山",
	"鳥取": "鳥取", "因幡": "鳥取", "伯耆": "鳥取",
	"島根": "島根", "出雲": "島根",
	"岡山": "岡山", "備前": "岡山",
	"広島": "広島", "安芸": "広島",
	"山口": "山口", "長門": "山口",
	"徳島": "徳島", "阿波": "徳島",
	"香川": "香川", "讃岐": "香川",
	"愛媛": "愛媛", "伊予": "愛媛",
	"高知": "高知", "土佐": "高知",
	"福岡": "福岡", "筑前": "福岡", "筑後": "福岡",
	"佐賀": "佐賀", "肥前": "佐賀",
	"長崎": "長崎",
	"熊本": "熊本", "肥後": "熊本",
	"大分": "大分", "豊後": "大分",
	"宮崎": "宮崎", "日向": "宮崎",
	"鹿児島": "鹿児島", "薩摩": "鹿児島",
	"沖縄": "沖縄", "琉球": "沖縄", "八重山": "沖縄", "石垣": "沖縄", "宮古": "沖縄",
}

// RegionFacetsFor derives prefecture / area / habitat tokens for filtering.
func RegionFacetsFor(entry IndexEntry) RegionFacets {
	prefs := append([]string{}, yokaiPrefectures[entry.NativeName]...)
	text := entry.Region + " " + entry.BlurbJA
	for place, pref := range placePrefecture {
		if strings.Contains(text, place) {
			prefs = appendUnique(prefs, pref)
		}
	}

	areas := make([]string, 0, 4)
	for _, pref := range prefs {
		if area, ok := prefectureToRegion[pref]; ok {
			areas = appendUnique(areas, area)
		}
	}
	for label, area := range map[string]string{
		"東北": "東北", "関東": "関東", "北陸": "北陸", "甲信": "甲信",
		"東海": "東海", "関西": "関西", "中国": "中国", "四国": "四国",
		"九州": "九州", "北海道": "北海道", "沖縄": "沖縄", "全国": "全国",
	} {
		if strings.Contains(entry.Region, label) {
			areas = appendUnique(areas, area)
		}
	}

	habitats := append([]string{}, categoryHabitats[entry.Category]...)
	for _, tag := range entry.Tags {
		if pref, ok := placePrefecture[tag]; ok {
			prefs = appendUnique(prefs, pref)
		}
		switch tag {
		case "川", "水", "沼", "海", "山", "森", "家", "屋敷", "道", "学校", "田", "畑":
			habitats = appendUnique(habitats, tag)
		}
	}

	return RegionFacets{
		Prefectures: prefs,
		Areas:       areas,
		Habitats:    habitats,
	}
}

func samePlace(a, b string) bool {
	a = CanonicalRegion(a)
	b = CanonicalRegion(b)
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	edoTokyo := func(s string) bool {
		return s == "江戸" || s == "東京"
	}
	return edoTokyo(a) && edoTokyo(b)
}

func (e IndexEntry) MatchesRegion(hint string) bool {
	if strings.TrimSpace(hint) == "" {
		return true
	}
	if regionMatches(e.Region, hint) {
		return true
	}
	facets := RegionFacetsFor(e)
	for _, pool := range [][]string{facets.Prefectures, facets.Areas, facets.Habitats} {
		for _, value := range pool {
			if samePlace(value, hint) {
				return true
			}
		}
	}
	return false
}
