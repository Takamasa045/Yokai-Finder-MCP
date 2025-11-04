package ndl

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

const defaultSRUBase = "https://ndlsearch.ndl.go.jp/api/sru"

// CQLの素朴な組み立て（タイトル/件名/簡易地域/カテゴリ/NDC/年）
func buildCQL(name, region, category, ndc string, yearFrom, yearTo int) string {
	var parts []string
	if name != "" {
		parts = append(parts, fmt.Sprintf("(title any \"%s\" or subject any \"%s\")", name, name))
	}
	if region != "" {
		parts = append(parts, fmt.Sprintf("(title any \"%s\" or subject any \"%s\")", region, region))
	}
	if category != "" {
		parts = append(parts, fmt.Sprintf("subject any \"%s\"", category))
	}
	if ndc != "" {
		parts = append(parts, fmt.Sprintf("ndc any \"%s\"", ndc))
	}
	// 年はSRUの実際のインデックス仕様に合わせ要調整。ここでは素朴にsubject/タイトル側に混ぜる簡易対応。
	if yearFrom > 0 && yearTo > 0 {
		parts = append(parts, fmt.Sprintf("(date any \"%d\" or date any \"%d\")", yearFrom, yearTo))
	}
	if len(parts) == 0 {
		parts = append(parts, "(title any \"妖怪\")")
	}
	return strings.Join(parts, " and ")
}

func SearchSRU(ctx context.Context, httpc *http.Client, base string, name, region, category, ndc string, yearFrom, yearTo, limit int) ([]types.BibliographyItem, error) {
	if base == "" {
		base = defaultSRUBase
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 20 * time.Second}
	}

	cql := buildCQL(name, region, category, ndc, yearFrom, yearTo)

	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("operation", "searchRetrieve")
	q.Set("version", "1.2")
	q.Set("recordSchema", "dcndl_simple")
	q.Set("maximumRecords", fmt.Sprintf("%d", limit))
	q.Set("query", cql)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	// 最低限のパーサ（dc要素を拾う）
	type dc struct {
		Title   []string `xml:"recordData/*/*[local-name()='title']"`
		Creator []string `xml:"recordData/*/*[local-name()='creator']"`
		Subject []string `xml:"recordData/*/*[local-name()='subject']"`
		Date    []string `xml:"recordData/*/*[local-name()='date']"`
		ID      []string `xml:"recordData/*/*[local-name()='identifier']"`
	}
	type sruResp struct {
		Records []dc `xml:"records\nrecord"`
	}

	var sr sruResp
	if err := xml.Unmarshal(b, &sr); err != nil {
		return nil, fmt.Errorf("sru xml parse: %w", err)
	}

	out := make([]types.BibliographyItem, 0, len(sr.Records))
	for _, r := range sr.Records {
		it := types.BibliographyItem{}
		if len(r.Title) > 0 {
			it.Title = strings.TrimSpace(r.Title[0])
		}
		it.Creators = r.Creator
		it.Subjects = r.Subject
		it.Identifiers = r.ID
		if len(r.Date) > 0 {
			it.Date = r.Date[0]
		}
		out = append(out, it)
	}
	return out, nil
}
