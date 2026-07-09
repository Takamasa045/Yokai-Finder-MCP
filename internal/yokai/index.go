package yokai

import (
	"math/rand"
	"sort"
	"strings"
)

// IndexEntry is a lightweight yokai roster row for browsing.
// Full lore lives in curated Profile entries; books come from NDL search.
type IndexEntry struct {
	Name       string
	NativeName string
	Category   string
	Region     string
	BlurbJA    string
	Tags       []string // e.g. 水 家 付喪神 入門 怖い かわいい 恋 予言 古典 現代
	Tone       string   // one of: gentle, comic, horror, solemn, tragic, mysterious, playful
	FamousRank int      // 1=iconic/well-known ... 5=obscure; default 3 if unsure
}

// SuggestQuery drives tag/vibe-based discovery for users who don't know names.
type SuggestQuery struct {
	Vibe     string // free text: 怖い, かわいい, 水, 夜道, etc.
	Theme    string // same matching pool
	Setting  string
	Audience string // 子ども, 創作, 入門, 学術 — mapped loosely to tags/famousRank
	Term     string // optional extra keyword
	Limit    int    // default 6, max 20
	Seed     int64  // 0 = stable by famous rank; non-zero = shuffle among top candidates
}

// Index returns a defensive copy of the yokai name index.
func Index() []IndexEntry {
	out := make([]IndexEntry, len(yokaiIndex))
	for i, e := range yokaiIndex {
		out[i] = copyIndexEntry(e)
	}
	return out
}

// FilterIndex returns index entries matching term / category / region hints.
// Matching is case-insensitive substring search across name fields, blurb, tags, and tone.
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
		filtered = append(filtered, copyIndexEntry(entry))
	}
	return filtered
}

// FindIndexByName looks up an index entry by English Name (case-insensitive)
// or exact NativeName, mirroring FindByName for profiles.
func FindIndexByName(name string) (IndexEntry, bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return IndexEntry{}, false
	}
	for _, entry := range yokaiIndex {
		if strings.EqualFold(entry.Name, trimmed) || entry.NativeName == trimmed {
			return copyIndexEntry(entry), true
		}
	}
	return IndexEntry{}, false
}

// Suggest ranks and returns index entries matching vibe/theme/setting/audience.
// With no filters it returns a diverse well-known set (famous rank 1–2 first).
func Suggest(query SuggestQuery) []IndexEntry {
	limit := query.Limit
	if limit <= 0 {
		limit = 6
	}
	if limit > 20 {
		limit = 20
	}

	vibe := normalizeQuery(query.Vibe)
	theme := normalizeQuery(query.Theme)
	setting := normalizeQuery(query.Setting)
	audience := normalizeQuery(query.Audience)
	term := normalizeQuery(query.Term)

	hasFilter := vibe != "" || theme != "" || setting != "" || audience != "" || term != ""

	type scored struct {
		entry IndexEntry
		score int
	}
	candidates := make([]scored, 0, len(yokaiIndex))

	for _, entry := range yokaiIndex {
		s := scoreEntry(entry, vibe, theme, setting, audience, term, hasFilter)
		if !hasFilter {
			// Unfiltered: only well-known first tier, mild mix of rank 3
			if entry.FamousRank > 3 {
				continue
			}
			candidates = append(candidates, scored{entry: entry, score: s})
			continue
		}
		// With filters: keep entries that scored above baseline
		if s > 0 {
			candidates = append(candidates, scored{entry: entry, score: s})
		}
	}

	// Fallback: if filters yielded nothing, return well-known set
	if len(candidates) == 0 {
		for _, entry := range yokaiIndex {
			if entry.FamousRank <= 2 {
				candidates = append(candidates, scored{
					entry: entry,
					score: scoreFamous(entry) + scoreAudience(entry, audience),
				})
			}
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		if candidates[i].entry.FamousRank != candidates[j].entry.FamousRank {
			return candidates[i].entry.FamousRank < candidates[j].entry.FamousRank
		}
		return candidates[i].entry.Name < candidates[j].entry.Name
	})

	// Optionally mix lesser-known among top pool when seed != 0
	poolSize := limit * 3
	if poolSize > len(candidates) {
		poolSize = len(candidates)
	}
	if poolSize == 0 {
		return nil
	}
	pool := candidates[:poolSize]

	if query.Seed != 0 && poolSize > 1 {
		r := rand.New(rand.NewSource(query.Seed))
		// Weighted shuffle: keep score influence but randomize among top
		r.Shuffle(len(pool), func(i, j int) {
			pool[i], pool[j] = pool[j], pool[i]
		})
		// Re-sort lightly so higher scores still tend to rise, then pick
		sort.SliceStable(pool, func(i, j int) bool {
			// After shuffle, only reorder when score gap is large
			if pool[i].score-pool[j].score > 15 {
				return true
			}
			if pool[j].score-pool[i].score > 15 {
				return false
			}
			return false // keep shuffled order for close scores
		})
	}

	n := limit
	if n > len(pool) {
		n = len(pool)
	}
	out := make([]IndexEntry, n)
	for i := 0; i < n; i++ {
		out[i] = copyIndexEntry(pool[i].entry)
	}
	return out
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

func copyIndexEntry(e IndexEntry) IndexEntry {
	out := e
	if len(e.Tags) > 0 {
		out.Tags = make([]string, len(e.Tags))
		copy(out.Tags, e.Tags)
	}
	return out
}

func normalizeQuery(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func indexMatchesTerm(entry IndexEntry, term string) bool {
	fields := []string{
		entry.Name,
		entry.NativeName,
		entry.Category,
		entry.Region,
		entry.BlurbJA,
		entry.Tone,
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), term) {
			return true
		}
	}
	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), term) {
			return true
		}
	}
	return false
}

func scoreEntry(entry IndexEntry, vibe, theme, setting, audience, term string, hasFilter bool) int {
	score := scoreFamous(entry)

	if vibe != "" {
		score += scoreFieldHit(entry, vibe, 30)
		score += scoreToneForVibe(entry, vibe)
	}
	if theme != "" {
		score += scoreFieldHit(entry, theme, 28)
	}
	if setting != "" {
		score += scoreFieldHit(entry, setting, 26)
		// Setting often maps to region/category
		if strings.Contains(strings.ToLower(entry.Region), setting) {
			score += 12
		}
		if strings.Contains(strings.ToLower(entry.Category), setting) {
			score += 10
		}
	}
	if term != "" {
		score += scoreFieldHit(entry, term, 22)
	}

	score += scoreAudience(entry, audience)

	if hasFilter {
		// Require at least one soft match from free-text fields when filters present;
		// audience alone can still surface via scoreAudience + famous baseline.
		hits := 0
		for _, q := range []string{vibe, theme, setting, term} {
			if q == "" {
				continue
			}
			if scoreFieldHit(entry, q, 1) > 0 || scoreToneForVibe(entry, q) > 0 {
				hits++
			}
		}
		if vibe == "" && theme == "" && setting == "" && term == "" {
			// audience-only query: keep if audience scoring applied or famous
			if scoreAudience(entry, audience) == 0 && entry.FamousRank > 3 {
				return 0
			}
			return score
		}
		if hits == 0 {
			// Allow audience-only boost path if audience set and entry fits
			if audience != "" && scoreAudience(entry, audience) > 0 {
				return score
			}
			return 0
		}
	}
	return score
}

func scoreFamous(entry IndexEntry) int {
	// Prefer lower FamousRank (more famous)
	switch entry.FamousRank {
	case 1:
		return 20
	case 2:
		return 14
	case 3:
		return 8
	case 4:
		return 3
	default:
		return 0
	}
}

func scoreFieldHit(entry IndexEntry, q string, boost int) int {
	if q == "" {
		return 0
	}
	score := 0
	// Strong: tags
	for _, tag := range entry.Tags {
		tl := strings.ToLower(tag)
		if tl == q {
			score += boost + 8
		} else if strings.Contains(tl, q) || strings.Contains(q, tl) {
			score += boost
		}
	}
	// Category / tone / names / blurb
	if strings.Contains(strings.ToLower(entry.Category), q) {
		score += boost
	}
	if strings.Contains(strings.ToLower(entry.Tone), q) {
		score += boost / 2
	}
	if strings.Contains(strings.ToLower(entry.Name), q) {
		score += boost / 2
	}
	if strings.Contains(strings.ToLower(entry.NativeName), q) {
		score += boost
	}
	if strings.Contains(strings.ToLower(entry.BlurbJA), q) {
		score += boost / 2
	}
	if strings.Contains(strings.ToLower(entry.Region), q) {
		score += boost / 3
	}
	// Synonym / related Japanese keyword expansions
	score += scoreSynonyms(entry, q, boost)
	return score
}

func scoreSynonyms(entry IndexEntry, q string, boost int) int {
	// Map common free-text vibes to related tags/tones/categories
	related := synonymMap[q]
	if len(related) == 0 {
		return 0
	}
	score := 0
	pool := entryMatchPool(entry)
	for _, r := range related {
		if strings.Contains(pool, r) {
			score += boost / 2
		}
	}
	return score
}

func entryMatchPool(entry IndexEntry) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(entry.Category))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(entry.Region))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(entry.BlurbJA))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(entry.Tone))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(entry.Name))
	b.WriteByte(' ')
	b.WriteString(strings.ToLower(entry.NativeName))
	for _, tag := range entry.Tags {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(tag))
	}
	return b.String()
}

// synonymMap expands Japanese free-text queries to related tokens present in tags/fields.
var synonymMap = map[string][]string{
	"水":   {"水", "水系", "海", "川", "沼", "河童", "雨"},
	"海":   {"海", "水", "船", "波", "海岸"},
	"川":   {"水", "水系", "川", "河童"},
	"怖い":  {"怖い", "horror", "恐怖", "死霊", "怨霊", "都市伝説", "現代"},
	"恐怖":  {"怖い", "horror", "死霊", "怨霊"},
	"かわいい": {"かわいい", "gentle", "comic", "playful", "入門", "童子"},
	"可愛い": {"かわいい", "gentle", "comic", "playful"},
	"夜道":  {"夜道", "道中", "夜", "街道"},
	"家":   {"家", "屋敷", "座敷", "廃屋", "障子"},
	"屋敷":  {"家", "屋敷", "座敷"},
	"山":   {"山", "山系", "森", "峠"},
	"森":   {"山", "森", "木", "木霊"},
	"恋":   {"恋", "恋情", "嫉妬", "tragic", "女"},
	"予言":  {"予言", "疫病", "瑞獣"},
	"古典":  {"古典", "神話", "絵巻", "平家"},
	"現代":  {"現代", "都市伝説", "現代伝承", "horror"},
	"付喪神": {"付喪神", "道具", "器物"},
	"狐":   {"狐", "狐狸", "稲荷"},
	"狸":   {"狸", "狐狸"},
	"猫":   {"猫", "化け猫", "猫又"},
	"鬼":   {"鬼", "鬼婆", "大江山"},
	"火":   {"火", "霊火", "炎", "雷"},
	"雪":   {"雪", "氷", "北国", "気象"},
	"子ども": {"かわいい", "入門", "童子", "gentle", "playful", "comic"},
	"子供":  {"かわいい", "入門", "童子", "gentle", "playful"},
	"入門":  {"入門", "有名", "かわいい"},
	"創作":  {"創作", "変化", "異形"},
	"学術":  {"古典", "神話", "絵巻", "solemn"},
	"夜":   {"夜", "夜道", "闇"},
	"学校":  {"学校", "現代", "都市伝説"},
	"田畑":  {"田畑", "農村", "田"},
}

func scoreToneForVibe(entry IndexEntry, vibe string) int {
	switch {
	case strings.Contains(vibe, "怖") || strings.Contains(vibe, "恐") || vibe == "horror":
		switch entry.Tone {
		case "horror":
			return 25
		case "tragic", "mysterious":
			return 12
		case "gentle", "comic", "playful":
			return -8
		}
	case strings.Contains(vibe, "かわい") || strings.Contains(vibe, "可愛") || strings.Contains(vibe, "癒し"):
		switch entry.Tone {
		case "gentle", "comic", "playful":
			return 25
		case "horror":
			return -15
		}
	case strings.Contains(vibe, "悲") || strings.Contains(vibe, "切な") || strings.Contains(vibe, "恋"):
		switch entry.Tone {
		case "tragic":
			return 22
		case "solemn", "mysterious":
			return 8
		}
	case strings.Contains(vibe, "おかし") || strings.Contains(vibe, "面白") || strings.Contains(vibe, "笑"):
		switch entry.Tone {
		case "comic", "playful":
			return 22
		}
	case strings.Contains(vibe, "神秘") || strings.Contains(vibe, "謎"):
		if entry.Tone == "mysterious" {
			return 20
		}
	}
	return 0
}

func scoreAudience(entry IndexEntry, audience string) int {
	if audience == "" {
		return 0
	}
	score := 0
	hasProfile := entry.HasCuratedProfile()
	tags := strings.Join(entry.Tags, " ")

	switch {
	case strings.Contains(audience, "入門") || strings.Contains(audience, "子ども") || strings.Contains(audience, "子供"):
		if entry.FamousRank <= 2 {
			score += 18
		} else if entry.FamousRank >= 4 {
			score -= 10
		}
		switch entry.Tone {
		case "gentle", "comic", "playful":
			score += 14
		case "horror":
			// deprioritize pure horror for kids/beginners unless other filters want it
			score -= 12
		}
		if strings.Contains(tags, "入門") || strings.Contains(tags, "かわいい") {
			score += 10
		}
	case strings.Contains(audience, "創作"):
		if hasProfile {
			score += 10
		}
		// mild variety boost for mid-tier
		if entry.FamousRank >= 2 && entry.FamousRank <= 4 {
			score += 6
		}
		if strings.Contains(tags, "創作") || entry.Tone == "mysterious" || entry.Tone == "tragic" {
			score += 6
		}
	case strings.Contains(audience, "学術"):
		if strings.Contains(tags, "古典") || strings.Contains(entry.Category, "古典") {
			score += 14
		}
		if entry.Tone == "solemn" || entry.Tone == "mysterious" {
			score += 8
		}
		if entry.FamousRank >= 3 {
			score += 4 // obscure scholarly interest
		}
	}
	return score
}
