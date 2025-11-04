package ndl

import "fmt"

// NDLのサムネイルは資料により可否やパスが異なる場合あり。
// ISBNがある場合の代表的パターンと、openBDフォールバックを返却。

type CoverURLs struct {
	Candidates []string `json:"candidates"`
}

func BuildCoverURLs(isbn13, jpno string) CoverURLs {
	urls := []string{}
	if isbn13 != "" {
		urls = append(urls,
			fmt.Sprintf("https://ndlsearch.ndl.go.jp/thumbnail/%s", isbn13),
			fmt.Sprintf("https://iss.ndl.go.jp/thumbnail/%s", isbn13),
			// openBD フォールバック
			fmt.Sprintf("https://cover.openbd.jp/%s.jpg", isbn13),
		)
	}
	if jpno != "" {
		urls = append(urls,
			fmt.Sprintf("https://iss.ndl.go.jp/thumbnail/%s", jpno),
		)
	}
	return CoverURLs{Candidates: urls}
}
