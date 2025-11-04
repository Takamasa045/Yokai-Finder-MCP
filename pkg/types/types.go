package types

// 書誌1件
type BibliographyItem struct {
	Title       string   `json:"title"`
	Creators    []string `json:"creators,omitempty"`
	Subjects    []string `json:"subjects,omitempty"`
	Identifiers []string `json:"identifiers,omitempty"`
	Date        string   `json:"date,omitempty"`
	NDC         []string `json:"ndc,omitempty"`
	URL         string   `json:"url,omitempty"`
}

// Fulltext 横断検索ヒット
type FulltextHit struct {
	PID     string `json:"pid"`
	PageNo  int    `json:"pageNo,omitempty"`
	Snippet string `json:"snippet,omitempty"`
	Title   string `json:"title,omitempty"`
}

// 挿絵候補
type IllustrationHit struct {
	PID     string `json:"pid"`
	PageNo  int    `json:"pageNo"`
	BBoxPct string `json:"bbox_pct"` // pct:x,y,w,h
	IIIFURL string `json:"iiif_url"`
	Caption string `json:"caption,omitempty"`
}

// NDLSH 典拠
type AuthorityResolution struct {
	Term      string   `json:"term"`
	PrefLabel string   `json:"prefLabel"`
	AltLabels []string `json:"altLabels,omitempty"`
	Broader   []string `json:"broader,omitempty"`
	Narrower  []string `json:"narrower,omitempty"`
	Related   []string `json:"related,omitempty"`
}
