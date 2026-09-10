package yokai

import "strings"

// Phrases left behind by bulk catalog generation. A hit means the field is
// prompt residue, not folklore.
var catalogArtifactNeedles = []string{
	"do not merge",
	"as its own plate",
	"encyclopedia plate",
	"encyclopedia body",
	"folk plate",
	"the workplace is",
	"neighbour card",
	"neighbor card",
	"well card",
	"faced shell",
	"別項",
	"巨大側",
	"図鑑に残る",
}

// CatalogArtifact reports the first generation-residue needle in s.
func CatalogArtifact(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	for _, needle := range catalogArtifactNeedles {
		if strings.Contains(lower, needle) || strings.Contains(trimmed, needle) {
			return needle
		}
	}
	if strings.HasPrefix(lower, "keep ") && strings.Contains(lower, " as ") {
		return "keep ... as"
	}
	return ""
}

func profileArtifact(p Profile) string {
	fields := []string{
		p.Name, p.NativeName, p.Region, p.Category, p.CategoryEN,
		p.Summary, p.SummaryJA, p.FunFact, p.FunFactJA, p.SearchQuery,
	}
	fields = append(fields, p.Legends...)
	fields = append(fields, p.LegendsJA...)
	fields = append(fields, p.Traits...)
	fields = append(fields, p.TraitsJA...)
	fields = append(fields, p.Motifs...)
	fields = append(fields, p.MotifsJA...)
	fields = append(fields, p.CreativeHooks...)
	fields = append(fields, p.CreativeHooksJA...)
	fields = append(fields, p.Sources...)
	for _, field := range fields {
		if hit := CatalogArtifact(field); hit != "" {
			return hit
		}
	}
	return ""
}

func indexArtifact(e IndexEntry) string {
	fields := []string{e.Name, e.NativeName, e.Category, e.Region, e.BlurbJA, e.Tone}
	fields = append(fields, e.Tags...)
	fields = append(fields, e.Aliases...)
	for _, field := range fields {
		if hit := CatalogArtifact(field); hit != "" {
			return hit
		}
	}
	return ""
}
