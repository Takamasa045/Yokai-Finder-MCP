package handler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"

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
	if p.Limit > 50 {
		p.Limit = 50
	}
	return p
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

	profile := yokai.RandomProfile(params.Seed, candidates)
	return profile, notes, nil
}

func convertProfile(profile yokai.Profile) types.YokaiProfile {
	return types.YokaiProfile{
		Name:          profile.Name,
		NativeName:    profile.NativeName,
		Region:        profile.Region,
		Category:      profile.Category,
		Summary:       profile.Summary,
		Legends:       cloneStrings(profile.Legends),
		Traits:        cloneStrings(profile.Traits),
		Motifs:        cloneStrings(profile.Motifs),
		FunFact:       profile.FunFact,
		CreativeHooks: cloneStrings(profile.CreativeHooks),
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
	category := strings.ToLower(params.Category)
	region := strings.ToLower(params.Region)

	for _, profile := range profiles {
		if category != "" && !strings.Contains(strings.ToLower(profile.Category), category) {
			continue
		}
		if region != "" && !strings.Contains(strings.ToLower(profile.Region), region) {
			continue
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
		r := rand.New(rand.NewSource(seed))
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
		Summary:    profile.Summary,
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
