package handler

import (
	"context"
	"errors"
	"fmt"
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
