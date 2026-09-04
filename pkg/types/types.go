package types

// Yokai search types
type YokaiSearchParams struct {
	Name     string `json:"name,omitempty"`
	Region   string `json:"region,omitempty"`
	Category string `json:"category,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type YokaiBook struct {
	Title                string   `json:"title"`
	Author               string   `json:"author,omitempty"`
	Publisher            string   `json:"publisher,omitempty"`
	PublishDate          string   `json:"publishDate,omitempty"`
	Description          string   `json:"description,omitempty"`
	URL                  string   `json:"url,omitempty"`
	ISBN                 string   `json:"isbn,omitempty"`
	Subjects             []string `json:"subjects,omitempty"`
	CoverImageCandidates []string `json:"coverImageCandidates,omitempty"`
}

type YokaiSearchResult struct {
	Query   string      `json:"query"`
	Total   int         `json:"total"`
	Results []YokaiBook `json:"results"`
}

// YokaiProfile captures curated lore for a yokai entry.
type YokaiProfile struct {
	Name          string   `json:"name"`
	NativeName    string   `json:"nativeName,omitempty"`
	Region        string   `json:"region,omitempty"`
	Category      string   `json:"category,omitempty"`
	Summary       string   `json:"summary"`
	SummaryJA     string   `json:"summaryJa,omitempty"`
	Legends       []string `json:"legends,omitempty"`
	Traits        []string `json:"traits,omitempty"`
	Motifs        []string `json:"motifs,omitempty"`
	FunFact       string   `json:"funFact,omitempty"`
	FunFactJA     string   `json:"funFactJa,omitempty"`
	CreativeHooks []string `json:"creativeHooks,omitempty"`
	Sources       []string `json:"sources,omitempty"`
}

// YokaiOfTheDayParams controls how the highlight tool selects a yokai.
type YokaiOfTheDayParams struct {
	Name     string `json:"name,omitempty"`
	Category string `json:"category,omitempty"`
	Region   string `json:"region,omitempty"`
	Seed     int64  `json:"seed,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// YokaiOfTheDayResult combines curated lore with recommended reading.
type YokaiOfTheDayResult struct {
	Profile          YokaiProfile `json:"profile"`
	Query            string       `json:"query"`
	TotalBooks       int          `json:"totalBooks"`
	RecommendedBooks []YokaiBook  `json:"recommendedBooks,omitempty"`
	StoryPrompt      string       `json:"storyPrompt,omitempty"`
	Notes            []string     `json:"notes,omitempty"`
}

// CuratedYokaiParams controls how curated yokai metadata is listed.
type CuratedYokaiParams struct {
	Term                 string `json:"term,omitempty"`
	Category             string `json:"category,omitempty"`
	Region               string `json:"region,omitempty"`
	Seed                 int64  `json:"seed,omitempty"`
	Limit                int    `json:"limit,omitempty"`
	IncludeLegends       bool   `json:"includeLegends,omitempty"`
	IncludeTraits        bool   `json:"includeTraits,omitempty"`
	IncludeMotifs        bool   `json:"includeMotifs,omitempty"`
	IncludeCreativeHooks bool   `json:"includeCreativeHooks,omitempty"`
}

// CuratedYokaiProfile surfaces curated lore in a lightweight shape for listings.
type CuratedYokaiProfile struct {
	Name          string   `json:"name"`
	NativeName    string   `json:"nativeName,omitempty"`
	Region        string   `json:"region,omitempty"`
	Category      string   `json:"category,omitempty"`
	Summary       string   `json:"summary,omitempty"`
	SummaryJA     string   `json:"summaryJa,omitempty"`
	Legends       []string `json:"legends,omitempty"`
	Traits        []string `json:"traits,omitempty"`
	Motifs        []string `json:"motifs,omitempty"`
	CreativeHooks []string `json:"creativeHooks,omitempty"`
}

// CuratedYokaiResult returns curated profiles plus helpful context.
type CuratedYokaiResult struct {
	Query    string                `json:"query,omitempty"`
	Total    int                   `json:"total"`
	Returned int                   `json:"returned"`
	Profiles []CuratedYokaiProfile `json:"profiles"`
	Notes    []string              `json:"notes,omitempty"`
}

// YokaiIndexParams controls lightweight roster browsing via list_yokai.
type YokaiIndexParams struct {
	Term          string `json:"term,omitempty"`
	Category      string `json:"category,omitempty"`
	Region        string `json:"region,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Tone          string `json:"tone,omitempty"`
	FamousRankMin int    `json:"famousRankMin,omitempty"`
	FamousRankMax int    `json:"famousRankMax,omitempty"`
	HasProfile    *bool  `json:"hasProfile,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

// YokaiIndexItem is a compact roster row for discovering yokai names.
type YokaiIndexItem struct {
	Name       string   `json:"name"`
	NativeName string   `json:"nativeName,omitempty"`
	Category   string   `json:"category,omitempty"`
	Region     string   `json:"region,omitempty"`
	BlurbJA    string   `json:"blurbJa,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Tone       string   `json:"tone,omitempty"`
	FamousRank int      `json:"famousRank,omitempty"`
	HasProfile bool     `json:"hasProfile"`
}

// YokaiIndexResult returns the filtered yokai roster.
type YokaiIndexResult struct {
	Query    string           `json:"query,omitempty"`
	Total    int              `json:"total"`
	Returned int              `json:"returned"`
	Items    []YokaiIndexItem `json:"items"`
	Notes    []string         `json:"notes,omitempty"`
}

// SuggestYokaiParams controls vibe/theme-based yokai discovery.
type SuggestYokaiParams struct {
	Vibe     string `json:"vibe,omitempty"`
	Theme    string `json:"theme,omitempty"`
	Setting  string `json:"setting,omitempty"`
	Audience string `json:"audience,omitempty"`
	Term     string `json:"term,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Seed     int64  `json:"seed,omitempty"`
}

// SuggestYokaiItem is one discovery candidate with a brief rationale.
type SuggestYokaiItem struct {
	Name         string   `json:"name"`
	NativeName   string   `json:"nativeName,omitempty"`
	Category     string   `json:"category,omitempty"`
	Region       string   `json:"region,omitempty"`
	BlurbJA      string   `json:"blurbJa,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Tone         string   `json:"tone,omitempty"`
	FamousRank   int      `json:"famousRank,omitempty"`
	HasProfile   bool     `json:"hasProfile"`
	WhySuggested string   `json:"whySuggested,omitempty"`
}

// SuggestYokaiResult returns discovery candidates for vague queries.
type SuggestYokaiResult struct {
	Query    string             `json:"query,omitempty"`
	Total    int                `json:"total"`
	Returned int                `json:"returned"`
	Items    []SuggestYokaiItem `json:"items"`
	Notes    []string           `json:"notes,omitempty"`
}

// GetYokaiParams looks up a single yokai by Japanese or English name.
type GetYokaiParams struct {
	Name string `json:"name"`
}

// GetYokaiResult returns either a full curated profile, an index card, or not-found.
type GetYokaiResult struct {
	Found       bool             `json:"found"`
	Source      string           `json:"source,omitempty"` // "profile" | "index" | ""
	Profile     *YokaiProfile    `json:"profile,omitempty"`
	Index       *YokaiIndexItem  `json:"index,omitempty"`
	Suggestions []YokaiIndexItem `json:"suggestions,omitempty"`
	Notes       []string         `json:"notes,omitempty"`
}

// RelatedYokaiParams finds nearby yokai by tags/category/tone.
type RelatedYokaiParams struct {
	Name  string `json:"name"`
	Limit int    `json:"limit,omitempty"`
}

// RelatedYokaiItem is a neighbour in the roster.
type RelatedYokaiItem struct {
	YokaiIndexItem
	Score  int      `json:"score"`
	Shared []string `json:"shared,omitempty"`
}

// RelatedYokaiResult lists similar yokai.
type RelatedYokaiResult struct {
	Name     string             `json:"name"`
	Total    int                `json:"total"`
	Returned int                `json:"returned"`
	Items    []RelatedYokaiItem `json:"items"`
	Notes    []string           `json:"notes,omitempty"`
}

// CompareYokaiParams compares two yokai by name.
type CompareYokaiParams struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

// CompareYokaiSide is one side of a comparison.
type CompareYokaiSide struct {
	Found   bool            `json:"found"`
	Source  string          `json:"source,omitempty"`
	Profile *YokaiProfile   `json:"profile,omitempty"`
	Index   *YokaiIndexItem `json:"index,omitempty"`
}

// CompareYokaiResult places two yokai side by side.
type CompareYokaiResult struct {
	Left     CompareYokaiSide `json:"left"`
	Right    CompareYokaiSide `json:"right"`
	Shared   []string         `json:"shared,omitempty"`
	Contrast []string         `json:"contrast,omitempty"`
	Notes    []string         `json:"notes,omitempty"`
}
