package yokai

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// NormalizeQuery folds whitespace, case, hyphens, and kana so lookups can
// match カッパ / かっぱ / Kappa / kappa.
func NormalizeQuery(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '-' || r == '_' || r == ' ' || r == '\u3000' || r == '\'':
			continue
		case unicode.Is(unicode.Mn, r):
			continue
		default:
			b.WriteRune(toHiragana(unicode.ToLower(r)))
		}
	}
	return b.String()
}

func toHiragana(r rune) rune {
	if r >= 0x30A1 && r <= 0x30F6 {
		return r - 0x60
	}
	return r
}

func isMostlyASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return utf8.RuneCountInString(s) > 0
}
