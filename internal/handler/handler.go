package handler

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/ndl"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/yokai"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
)

// Handler coordinates Yokai search requests against the NDL API with caching.
type Handler struct {
	ndlClient *ndl.Client
	cache     *cache.Cache
}

// New creates a handler with the provided dependencies.
func New(ndlClient *ndl.Client, cache *cache.Cache) *Handler {
	return &Handler{
		ndlClient: ndlClient,
		cache:     cache,
	}
}

// SearchYokai finds yokai related literature via the NDL OpenSearch API.
func (h *Handler) SearchYokai(ctx context.Context, params types.YokaiSearchParams) (*types.YokaiSearchResult, error) {
	if h == nil || h.ndlClient == nil {
		return nil, errors.New("handler not initialised")
	}

	cleaned := normaliseParams(params)

	if h.cache != nil {
		if result, ok := h.cache.Get(cleaned); ok {
			return result, nil
		}
	}

	result, err := h.ndlClient.SearchYokaiBooks(ctx, cleaned)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("ndl returned no data")
	}

	if h.cache != nil {
		h.cache.Set(cleaned, result)
	}
	return result, nil
}

// YokaiOfTheDay returns a curated yokai profile with optional book recommendations.
func (h *Handler) YokaiOfTheDay(ctx context.Context, params types.YokaiOfTheDayParams) (*types.YokaiOfTheDayResult, error) {
	if h == nil || h.ndlClient == nil {
		return nil, errors.New("handler not initialised")
	}

	cleaned := normaliseHighlightParams(params)

	profile, notes, err := selectCuratedProfile(cleaned)
	if err != nil {
		return nil, err
	}

	queryTerm := strings.TrimSpace(profile.SearchQuery)
	if queryTerm == "" {
		queryTerm = strings.TrimSpace(profile.NativeName)
	}
	if queryTerm == "" {
		queryTerm = profile.Name
	}

	searchParams := types.YokaiSearchParams{
		Name:     queryTerm,
		Region:   cleaned.Region,
		Category: cleaned.Category,
		Limit:    cleaned.Limit,
	}

	var (
		query      string
		totalBooks int
		recs       []types.YokaiBook
	)

	searchResult, searchErr := h.SearchYokai(ctx, searchParams)
	if searchErr != nil {
		notes = append(notes, fmt.Sprintf("NDL search unavailable (%v); showing curated lore only.", searchErr))
		query = queryTerm
	} else if searchResult != nil {
		query = searchResult.Query
		totalBooks = searchResult.Total
		recs = takeBooks(searchResult.Results, cleaned.Limit)
	}

	result := &types.YokaiOfTheDayResult{
		Profile:          convertProfile(profile),
		Query:            query,
		TotalBooks:       totalBooks,
		RecommendedBooks: recs,
		StoryPrompt:      buildStoryPrompt(profile, recs),
		Notes:            notes,
	}

	if result.Query == "" {
		result.Query = queryTerm
	}

	return result, nil
}

// ListYokai returns the lightweight yokai roster for discovery.
func (h *Handler) ListYokai(_ context.Context, params types.YokaiIndexParams) (*types.YokaiIndexResult, error) {
	cleaned := normaliseIndexParams(params)

	matches := yokai.FilterIndexOpts(yokai.IndexFilter{
		Term:          cleaned.Term,
		Category:      cleaned.Category,
		Region:        cleaned.Region,
		Tag:           cleaned.Tag,
		Tone:          cleaned.Tone,
		FamousRankMin: cleaned.FamousRankMin,
		FamousRankMax: cleaned.FamousRankMax,
		HasProfile:    cleaned.HasProfile,
	})
	if len(matches) == 0 {
		return &types.YokaiIndexResult{
			Query:    cleaned.Term,
			Total:    0,
			Returned: 0,
			Items:    nil,
			Notes:    []string{"No yokai in the index matched the provided filters."},
		}, nil
	}

	limited := matches
	notes := []string{}
	if cleaned.Limit > 0 && len(matches) > cleaned.Limit {
		limited = matches[:cleaned.Limit]
		notes = append(notes, fmt.Sprintf("Showing first %d of %d indexed yokai.", cleaned.Limit, len(matches)))
	}

	items := make([]types.YokaiIndexItem, 0, len(limited))
	for _, entry := range limited {
		items = append(items, convertIndexEntry(entry))
	}

	return &types.YokaiIndexResult{
		Query:    cleaned.Term,
		Total:    len(matches),
		Returned: len(items),
		Items:    items,
		Notes:    notes,
	}, nil
}

// SuggestYokai recommends yokai for vague vibe/theme/setting queries.
func (h *Handler) SuggestYokai(_ context.Context, params types.SuggestYokaiParams) (*types.SuggestYokaiResult, error) {
	cleaned := normaliseSuggestParams(params)

	matches := yokai.Suggest(yokai.SuggestQuery{
		Vibe:     cleaned.Vibe,
		Theme:    cleaned.Theme,
		Setting:  cleaned.Setting,
		Audience: cleaned.Audience,
		Term:     cleaned.Term,
		Limit:    cleaned.Limit,
		Seed:     cleaned.Seed,
	})

	query := buildSuggestQueryLabel(cleaned)
	if len(matches) == 0 {
		return &types.SuggestYokaiResult{
			Query:    query,
			Total:    0,
			Returned: 0,
			Items:    nil,
			Notes: []string{
				"条件に合う妖怪が見つかりませんでした。vibe/theme/setting を緩めるか list_yokai・search_yokai_books をお試しください。",
			},
		}, nil
	}

	items := make([]types.SuggestYokaiItem, 0, len(matches))
	for _, entry := range matches {
		items = append(items, types.SuggestYokaiItem{
			Name:         entry.Name,
			NativeName:   entry.NativeName,
			Category:     entry.Category,
			Region:       entry.Region,
			BlurbJA:      entry.BlurbJA,
			Tags:         cloneStrings(entry.Tags),
			Tone:         entry.Tone,
			FamousRank:   entry.FamousRank,
			HasProfile:   entry.HasCuratedProfile(),
			WhySuggested: buildWhySuggested(entry, cleaned),
		})
	}

	return &types.SuggestYokaiResult{
		Query:    query,
		Total:    len(items),
		Returned: len(items),
		Items:    items,
		Notes:    nil,
	}, nil
}

// GetYokai looks up one yokai by Japanese or English name (profile first, then index).
func (h *Handler) GetYokai(_ context.Context, params types.GetYokaiParams) (*types.GetYokaiResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	if profile, ok := yokai.FindByName(name); ok {
		converted := convertProfile(profile)
		return &types.GetYokaiResult{
			Found:   true,
			Source:  "profile",
			Profile: &converted,
			Notes:   nil,
		}, nil
	}

	if entry, ok := yokai.FindIndexByName(name); ok {
		indexItem := convertIndexEntry(entry)
		notes := []string{
			"索引カードのみです。詳細な伝承・創作フックは未整備の可能性があります。",
			"深掘りするなら search_yokai_books で関連書籍を探すか、related_yokai で近い図鑑エントリを参照してください。",
		}
		return &types.GetYokaiResult{
			Found:  true,
			Source: "index",
			Index:  &indexItem,
			Notes:  notes,
		}, nil
	}

	suggestions := yokai.SuggestNames(name, 5)
	items := make([]types.YokaiIndexItem, 0, len(suggestions))
	for _, entry := range suggestions {
		items = append(items, convertIndexEntry(entry))
	}
	notes := []string{
		fmt.Sprintf("「%s」に一致する妖怪が見つかりませんでした。", name),
		"list_yokai で一覧を見る、suggest_yokai で雰囲気から探す、search_yokai_books で文献検索をお試しください。",
	}
	if len(items) > 0 {
		notes = append([]string{"もしかして、次の候補ですか？"}, notes...)
	}
	return &types.GetYokaiResult{
		Found:       false,
		Source:      "",
		Suggestions: items,
		Notes:       notes,
	}, nil
}

// ListCuratedYokai returns curated profiles filtered and shaped for quick browsing.
func (h *Handler) ListCuratedYokai(_ context.Context, params types.CuratedYokaiParams) (*types.CuratedYokaiResult, error) {
	cleaned := normaliseCuratedParams(params)

	profiles := yokai.Profiles()
	matches := filterCuratedProfiles(profiles, cleaned)

	if len(matches) == 0 {
		return &types.CuratedYokaiResult{
			Query:    cleaned.Term,
			Total:    0,
			Returned: 0,
			Profiles: nil,
			Notes:    []string{"No curated yokai matched the provided filters."},
		}, nil
	}

	ordered := orderCuratedProfiles(matches, cleaned.Seed)

	limited := ordered
	notes := []string{}
	if cleaned.Limit > 0 && len(ordered) > cleaned.Limit {
		limited = ordered[:cleaned.Limit]
		notes = append(notes, fmt.Sprintf("Showing first %d of %d curated yokai.", cleaned.Limit, len(ordered)))
	}

	resultProfiles := make([]types.CuratedYokaiProfile, 0, len(limited))
	for _, profile := range limited {
		resultProfiles = append(resultProfiles, convertCuratedProfile(profile, cleaned))
	}

	return &types.CuratedYokaiResult{
		Query:    cleaned.Term,
		Total:    len(ordered),
		Returned: len(resultProfiles),
		Profiles: resultProfiles,
		Notes:    notes,
	}, nil
}

func normaliseParams(p types.YokaiSearchParams) types.YokaiSearchParams {
	p.Name = strings.TrimSpace(p.Name)
	p.Region = strings.TrimSpace(p.Region)
	p.Category = strings.TrimSpace(p.Category)

	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 50 {
		p.Limit = 50
	}
	return p
}

func normaliseHighlightParams(p types.YokaiOfTheDayParams) types.YokaiOfTheDayParams {
	p.Name = strings.TrimSpace(p.Name)
	p.Category = strings.TrimSpace(p.Category)
	p.Region = strings.TrimSpace(p.Region)

	if p.Limit <= 0 {
		p.Limit = 5
	}
	if p.Limit > 10 {
		p.Limit = 10
	}
	return p
}

func normaliseCuratedParams(p types.CuratedYokaiParams) types.CuratedYokaiParams {
	p.Term = strings.TrimSpace(p.Term)
	p.Category = strings.TrimSpace(p.Category)
	p.Region = strings.TrimSpace(p.Region)

	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 200 {
		p.Limit = 200
	}
	return p
}

func normaliseIndexParams(p types.YokaiIndexParams) types.YokaiIndexParams {
	p.Term = strings.TrimSpace(p.Term)
	p.Category = strings.TrimSpace(p.Category)
	p.Region = strings.TrimSpace(p.Region)
	p.Tag = strings.TrimSpace(p.Tag)
	p.Tone = strings.TrimSpace(p.Tone)

	if p.Limit <= 0 {
		p.Limit = 200
	}
	if p.Limit > 200 {
		p.Limit = 200
	}
	return p
}

func normaliseSuggestParams(p types.SuggestYokaiParams) types.SuggestYokaiParams {
	p.Vibe = strings.TrimSpace(p.Vibe)
	p.Theme = strings.TrimSpace(p.Theme)
	p.Setting = strings.TrimSpace(p.Setting)
	p.Audience = strings.TrimSpace(p.Audience)
	p.Term = strings.TrimSpace(p.Term)

	if p.Limit <= 0 {
		p.Limit = 6
	}
	if p.Limit > 20 {
		p.Limit = 20
	}
	return p
}

func convertIndexEntry(entry yokai.IndexEntry) types.YokaiIndexItem {
	return types.YokaiIndexItem{
		Name:       entry.Name,
		NativeName: entry.NativeName,
		Category:   entry.Category,
		Region:     entry.Region,
		BlurbJA:    entry.BlurbJA,
		Tags:       cloneStrings(entry.Tags),
		Tone:       entry.Tone,
		FamousRank: entry.FamousRank,
		HasProfile: entry.HasCuratedProfile(),
	}
}

func buildSuggestQueryLabel(p types.SuggestYokaiParams) string {
	parts := make([]string, 0, 5)
	if p.Vibe != "" {
		parts = append(parts, "vibe="+p.Vibe)
	}
	if p.Theme != "" {
		parts = append(parts, "theme="+p.Theme)
	}
	if p.Setting != "" {
		parts = append(parts, "setting="+p.Setting)
	}
	if p.Audience != "" {
		parts = append(parts, "audience="+p.Audience)
	}
	if p.Term != "" {
		parts = append(parts, "term="+p.Term)
	}
	return strings.Join(parts, " ")
}

func buildWhySuggested(entry yokai.IndexEntry, p types.SuggestYokaiParams) string {
	reasons := make([]string, 0, 4)

	if p.Vibe != "" && fieldMentions(entry, p.Vibe) {
		reasons = append(reasons, fmt.Sprintf("雰囲気「%s」に合いそう", p.Vibe))
	}
	if p.Theme != "" && fieldMentions(entry, p.Theme) {
		reasons = append(reasons, fmt.Sprintf("テーマ「%s」と関連", p.Theme))
	}
	if p.Setting != "" && fieldMentions(entry, p.Setting) {
		reasons = append(reasons, fmt.Sprintf("舞台「%s」に近い", p.Setting))
	}
	if p.Term != "" && fieldMentions(entry, p.Term) {
		reasons = append(reasons, fmt.Sprintf("キーワード「%s」にヒット", p.Term))
	}
	if p.Audience != "" {
		reasons = append(reasons, fmt.Sprintf("想定読者「%s」向けの候補", p.Audience))
	}

	// Tag overlap is a strong discovery signal even without query field hits.
	if len(reasons) == 0 && len(entry.Tags) > 0 {
		reasons = append(reasons, fmt.Sprintf("タグ: %s", strings.Join(entry.Tags, "・")))
	}
	if len(reasons) == 0 {
		if entry.Tone != "" {
			return fmt.Sprintf("トーン「%s」の代表的な候補", entry.Tone)
		}
		if entry.Category != "" {
			return fmt.Sprintf("カテゴリ「%s」からのおすすめ", entry.Category)
		}
		return "索引からのおすすめ"
	}
	return strings.Join(reasons, "／")
}

func fieldMentions(entry yokai.IndexEntry, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	if needle == "" {
		return false
	}
	fields := []string{
		entry.Name,
		entry.NativeName,
		entry.Category,
		entry.Region,
		entry.BlurbJA,
		entry.Tone,
		strings.Join(entry.Tags, " "),
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), needle) {
			return true
		}
	}
	return false
}

func selectCuratedProfile(params types.YokaiOfTheDayParams) (yokai.Profile, []string, error) {
	var notes []string

	if params.Name != "" {
		if profile, ok := yokai.FindByName(params.Name); ok {
			if params.Category != "" && !strings.Contains(strings.ToLower(profile.Category), strings.ToLower(params.Category)) {
				notes = append(notes, "Requested category did not match the curated profile; showing the requested yokai anyway.")
			}
			if params.Region != "" && !strings.Contains(strings.ToLower(profile.Region), strings.ToLower(params.Region)) {
				notes = append(notes, "Requested region did not match the curated profile; showing the requested yokai anyway.")
			}
			return profile, notes, nil
		}
		return yokai.Profile{}, nil, fmt.Errorf("no curated data available for yokai %q", params.Name)
	}

	candidates := yokai.Filter(params.Category, params.Region)
	if len(candidates) == 0 {
		notes = append(notes, "No curated yokai matched the provided filters; offering a surprise pick instead.")
		candidates = yokai.Profiles()
	}

	if params.Seed == 0 {
		notes = append(notes, fmt.Sprintf("Daily pick for %s JST: the same yokai appears all day. Pass a seed (or name) for a different one.", time.Now().In(yokai.JST).Format("2006-01-02")))
	}

	profile := yokai.RandomProfile(params.Seed, candidates)
	return profile, notes, nil
}

func convertProfile(profile yokai.Profile) types.YokaiProfile {
	return types.YokaiProfile{
		Name:          profile.Name,
		NativeName:    profile.NativeName,
		Region:        profile.Region,
		Category:      profile.Category,
		CategoryEN:    profile.CategoryEN,
		Summary:       profile.Summary,
		SummaryJA:     profile.SummaryJA,
		Legends:       cloneStrings(profile.Legends),
		Traits:        cloneStrings(profile.Traits),
		Motifs:        cloneStrings(profile.Motifs),
		FunFact:       profile.FunFact,
		FunFactJA:     profile.FunFactJA,
		CreativeHooks: cloneStrings(profile.CreativeHooks),
		Sources:       cloneStrings(profile.Sources),
	}
}

func takeBooks(books []types.YokaiBook, limit int) []types.YokaiBook {
	if len(books) == 0 || limit <= 0 {
		return nil
	}
	if len(books) > limit {
		books = books[:limit]
	}
	out := make([]types.YokaiBook, len(books))
	copy(out, books)
	return out
}

func buildStoryPrompt(profile yokai.Profile, books []types.YokaiBook) string {
	displayName := profile.Name
	if profile.NativeName != "" {
		displayName = fmt.Sprintf("%s (%s)", profile.Name, profile.NativeName)
	}

	region := profile.Region
	if region == "" {
		region = "Japan"
	}

	var motif string
	if len(profile.Motifs) > 0 {
		motif = profile.Motifs[0]
	} else if len(profile.Traits) > 0 {
		motif = profile.Traits[0]
	} else {
		motif = "their folklore presence"
	}

	var bookTitle string
	for _, book := range books {
		if strings.TrimSpace(book.Title) != "" {
			bookTitle = book.Title
			break
		}
	}

	var hook string
	if len(profile.CreativeHooks) > 0 {
		hook = profile.CreativeHooks[0]
	}

	var builder strings.Builder
	builder.WriteString("Craft a scene featuring ")
	builder.WriteString(displayName)
	builder.WriteString(" in ")
	builder.WriteString(region)
	builder.WriteString(", highlighting ")
	builder.WriteString(motif)
	builder.WriteString(".")

	if bookTitle != "" {
		builder.WriteString(" Let the tone draw inspiration from the book \"")
		builder.WriteString(bookTitle)
		builder.WriteString("\".")
	}

	if hook != "" {
		builder.WriteString(" Bonus idea: ")
		builder.WriteString(hook)
	}

	return builder.String()
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func filterCuratedProfiles(profiles []yokai.Profile, params types.CuratedYokaiParams) []yokai.Profile {
	var filtered []yokai.Profile
	term := strings.ToLower(params.Term)

	allowed := map[string]struct{}{}
	useAllow := params.Category != "" || params.Region != ""
	if useAllow {
		for _, profile := range yokai.Filter(params.Category, params.Region) {
			allowed[profile.Name] = struct{}{}
		}
	}

	for _, profile := range profiles {
		if useAllow {
			if _, ok := allowed[profile.Name]; !ok {
				continue
			}
		}
		if term != "" && !profileMatchesTerm(profile, term) {
			continue
		}
		filtered = append(filtered, profile)
	}
	return filtered
}

func profileMatchesTerm(profile yokai.Profile, term string) bool {
	fields := []string{
		profile.Name,
		profile.NativeName,
		profile.Region,
		profile.Category,
		profile.Summary,
		profile.SummaryJA,
		profile.FunFactJA,
		strings.Join(profile.Legends, " "),
		strings.Join(profile.Traits, " "),
		strings.Join(profile.Motifs, " "),
		strings.Join(profile.CreativeHooks, " "),
	}

	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), term) {
			return true
		}
	}
	return false
}

func orderCuratedProfiles(profiles []yokai.Profile, seed int64) []yokai.Profile {
	if len(profiles) <= 1 {
		return profiles
	}

	ordered := make([]yokai.Profile, len(profiles))
	copy(ordered, profiles)

	if seed != 0 {
		r := rand.New(rand.NewPCG(uint64(seed), uint64(^seed)))
		r.Shuffle(len(ordered), func(i, j int) {
			ordered[i], ordered[j] = ordered[j], ordered[i]
		})
		return ordered
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return strings.ToLower(ordered[i].Name) < strings.ToLower(ordered[j].Name)
	})
	return ordered
}

func convertCuratedProfile(profile yokai.Profile, params types.CuratedYokaiParams) types.CuratedYokaiProfile {
	result := types.CuratedYokaiProfile{
		Name:       profile.Name,
		NativeName: profile.NativeName,
		Region:     profile.Region,
		Category:   profile.Category,
		CategoryEN: profile.CategoryEN,
		Summary:    profile.Summary,
		SummaryJA:  profile.SummaryJA,
	}

	if params.IncludeLegends {
		result.Legends = cloneStrings(profile.Legends)
	}
	if params.IncludeTraits {
		result.Traits = cloneStrings(profile.Traits)
	}
	if params.IncludeMotifs {
		result.Motifs = cloneStrings(profile.Motifs)
	}
	if params.IncludeCreativeHooks {
		result.CreativeHooks = cloneStrings(profile.CreativeHooks)
	}

	return result
}

// RelatedYokai returns roster neighbours that share tags, category, or tone.
func (h *Handler) RelatedYokai(_ context.Context, params types.RelatedYokaiParams) (*types.RelatedYokaiResult, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	origin, matches, shared, _ := yokai.Related(name, params.Limit)
	if origin.Name == "" {
		return &types.RelatedYokaiResult{
			Name:  name,
			Notes: []string{fmt.Sprintf("「%s」に一致する妖怪が見つかりませんでした。", name)},
		}, nil
	}
	items := make([]types.RelatedYokaiItem, 0, len(matches))
	for i, entry := range matches {
		item := types.RelatedYokaiItem{
			YokaiIndexItem: convertIndexEntry(entry),
			Shared:         cloneStrings(shared[i]),
			Score:          len(shared[i]),
		}
		items = append(items, item)
	}
	return &types.RelatedYokaiResult{
		Name:     origin.NativeName,
		Total:    len(items),
		Returned: len(items),
		Items:    items,
	}, nil
}

// CompareYokai places two yokai side by side.
func (h *Handler) CompareYokai(ctx context.Context, params types.CompareYokaiParams) (*types.CompareYokaiResult, error) {
	leftName := strings.TrimSpace(params.Left)
	rightName := strings.TrimSpace(params.Right)
	if leftName == "" || rightName == "" {
		return nil, errors.New("left and right names are required")
	}

	left, err := h.GetYokai(ctx, types.GetYokaiParams{Name: leftName})
	if err != nil {
		return nil, err
	}
	right, err := h.GetYokai(ctx, types.GetYokaiParams{Name: rightName})
	if err != nil {
		return nil, err
	}

	result := &types.CompareYokaiResult{
		Left:  compareSide(left),
		Right: compareSide(right),
	}
	if !left.Found || !right.Found {
		result.Notes = append(result.Notes, "片方または両方の名前が索引にありません。")
		return result, nil
	}

	leftTags := sideTags(left)
	rightTags := sideTags(right)
	leftSet := map[string]struct{}{}
	for _, tag := range leftTags {
		leftSet[tag] = struct{}{}
	}
	rightSet := map[string]struct{}{}
	for _, tag := range rightTags {
		rightSet[tag] = struct{}{}
	}
	for _, tag := range leftTags {
		if _, ok := rightSet[tag]; ok {
			result.Shared = append(result.Shared, tag)
		} else {
			result.Contrast = append(result.Contrast, "left:"+tag)
		}
	}
	for _, tag := range rightTags {
		if _, ok := leftSet[tag]; !ok {
			result.Contrast = append(result.Contrast, "right:"+tag)
		}
	}
	return result, nil
}

func compareSide(result *types.GetYokaiResult) types.CompareYokaiSide {
	return types.CompareYokaiSide{
		Found:   result.Found,
		Source:  result.Source,
		Profile: result.Profile,
		Index:   result.Index,
	}
}

func sideTags(result *types.GetYokaiResult) []string {
	name := ""
	native := ""
	if result.Index != nil {
		name = result.Index.Name
		native = result.Index.NativeName
	} else if result.Profile != nil {
		name = result.Profile.Name
		native = result.Profile.NativeName
	}
	entry, ok := yokai.FindIndexByName(name)
	if !ok {
		entry, ok = yokai.FindIndexByName(native)
	}
	if !ok {
		return nil
	}
	tags := append([]string{}, entry.Tags...)
	if entry.Category != "" {
		tags = append(tags, "category:"+entry.Category)
	}
	if entry.Tone != "" {
		tags = append(tags, "tone:"+entry.Tone)
	}
	return tags
}
