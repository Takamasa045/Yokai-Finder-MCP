package ndl

import (
	"fmt"
	"strings"
)

// NDLのサムネイルは資料により可否やパスが異なる場合あり。
// ISBNがある場合の代表的パターンと、openBDフォールバックを返却。

type CoverURLs struct {
	Candidates []string `json:"candidates"`
}

// NormalizeISBN13 strips hyphens/spaces and converts ISBN-10 input to
// ISBN-13. It returns "" when the value cannot be read as an ISBN.
func sanitizeJPNo(raw string) string {
	var digits strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		if r < '0' || r > '9' {
			return ""
		}
		digits.WriteRune(r)
	}
	s := digits.String()
	if len(s) < 8 || len(s) > 10 {
		return ""
	}
	return s
}

func NormalizeISBN13(raw string) string {
	var cleaned strings.Builder
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			cleaned.WriteRune(r)
		case r == 'x' || r == 'X':
			cleaned.WriteByte('X')
		case r == '-' || r == ' ':
			// separators are ignored
		default:
			return ""
		}
	}

	s := cleaned.String()
	switch len(s) {
	case 13:
		if strings.ContainsRune(s, 'X') {
			return ""
		}
		return s
	case 10:
		// The ISBN-10 check digit (possibly X) is dropped and recomputed
		// for the 978-prefixed EAN body.
		body := s[:9]
		if strings.ContainsRune(body, 'X') {
			return ""
		}
		body = "978" + body
		return body + string(ean13CheckDigit(body))
	}
	return ""
}

// ean13CheckDigit computes the trailing check digit for a 12-digit EAN body.
func ean13CheckDigit(body12 string) byte {
	sum := 0
	for i, r := range body12 {
		digit := int(r - '0')
		if i%2 == 1 {
			digit *= 3
		}
		sum += digit
	}
	return byte('0' + (10-sum%10)%10)
}

// BuildCoverURLs returns candidate cover-image URLs for an ISBN (10 or 13,
// hyphens allowed) and/or a JP number (全国書誌番号).
func BuildCoverURLs(isbn, jpno string) CoverURLs {
	urls := []string{}
	if isbn13 := NormalizeISBN13(isbn); isbn13 != "" {
		urls = append(urls,
			fmt.Sprintf("https://ndlsearch.ndl.go.jp/thumbnail/%s.jpg", isbn13),
			// openBD フォールバック
			fmt.Sprintf("https://cover.openbd.jp/%s.jpg", isbn13),
			// 旧NDLサーチ（リダイレクトされる場合あり）
			fmt.Sprintf("https://iss.ndl.go.jp/thumbnail/%s", isbn13),
		)
	}
	if jpno = sanitizeJPNo(jpno); jpno != "" {
		urls = append(urls,
			fmt.Sprintf("https://iss.ndl.go.jp/thumbnail/%s", jpno),
		)
	}
	return CoverURLs{Candidates: urls}
}
