package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/yourname/yokai-finder-mcp/internal/ndl"
)

type Handler struct {
	HC *http.Client
}

func New() *Handler {
	return &Handler{HC: &http.Client{Timeout: 25 * time.Second}}
}

// tools/list で返す定義（最簡素）
func (h *Handler) Tools() []map[string]any {
	return []map[string]any{
		{
			"name":        "search_bibliography",
			"description": "NDLサーチ SRU（CQL）で書誌を検索",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":      map[string]any{"type": "string"},
					"region":    map[string]any{"type": "string"},
					"category":  map[string]any{"type": "string"},
					"ndc":       map[string]any{"type": "string"},
					"year_from": map[string]any{"type": "integer"},
					"year_to":   map[string]any{"type": "integer"},
					"limit":     map[string]any{"type": "integer"},
				},
			},
		},
		{
			"name":        "get_cover_thumbnail",
			"description": "ISBN/JP番号から書影URL候補を返す",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{
				"isbn13": map[string]any{"type": "string"},
				"jpec":   map[string]any{"type": "string"},
			}},
		},
		{
			"name":        "crop_iiif_region",
			"description": "IIIFのpct座標で切り出しURLを合成",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{
				"iiif_base": map[string]any{"type": "string"},
				"bbox_pct":  map[string]any{"type": "string"},
			}},
		},
		{"name": "search_fulltext", "description": "NDLラボ Book API 横断検索（要 base 設定）", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"query": map[string]any{"type": "string"}, "content_only": map[string]any{"type": "boolean"}, "from": map[string]any{"type": "integer"}, "size": map[string]any{"type": "integer"}}}},
		{"name": "search_in_book", "description": "NDLラボ Page API 資料内検索（要 base 設定）", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"pid": map[string]any{"type": "string"}, "q": map[string]any{"type": "string"}, "from": map[string]any{"type": "integer"}, "size": map[string]any{"type": "integer"}}}},
		{"name": "find_illustrations", "description": "NDLラボ Illustration API 類似画像（要 base 設定）", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"keyword": map[string]any{"type": "string"}, "limit": map[string]any{"type": "integer"}}}},
		{"name": "resolve_ndl_authority", "description": "NDLSH 典拠（SPARQL、要 endpoint 設定）", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"term": map[string]any{"type": "string"}, "id": map[string]any{"type": "string"}}}},
	}
}

// tools/call で呼ばれる実体
func (h *Handler) Call(ctx context.Context, name string, args json.RawMessage) (any, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	switch name {
	case "search_bibliography":
		var p struct {
			Name      string
			Region    string
			Category  string
			Ndc       string
			Year_from int
			Year_to   int
			Limit     int
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		base := os.Getenv("NDL_SRU_BASE")
		return ndl.SearchSRU(ctx, h.HC, base, p.Name, p.Region, p.Category, p.Ndc, p.Year_from, p.Year_to, max1(p.Limit, 10))

	case "get_cover_thumbnail":
		var p struct {
			Isbn13 string
			Jpec   string
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		return ndl.BuildCoverURLs(p.Isbn13, p.Jpec), nil

	case "crop_iiif_region":
		var p struct {
			IIIFBase string
			Bbox_pct string
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		return map[string]string{"url": ndl.BuildIIIFCropURL(p.IIIFBase, p.Bbox_pct)}, nil

	case "search_fulltext":
		var p struct {
			Query        string
			Content_only bool
			From         int
			Size         int
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		base := os.Getenv("NDL_LAB_BASE")
		hits, err := ndl.SearchFulltext(ctx, h.HC, base, p.Query, p.Content_only, p.From, max1(p.Size, 20))
		return hits, err

	case "search_in_book":
		var p struct {
			PID  string
			Q    string
			From int
			Size int
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		base := os.Getenv("NDL_LAB_BASE")
		pages, err := ndl.SearchInBook(ctx, h.HC, base, p.PID, p.Q, p.From, max1(p.Size, 20))
		return map[string]any{"pages": pages}, err

	case "find_illustrations":
		var p struct {
			Keyword string
			Limit   int
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		base := os.Getenv("NDL_LAB_BASE")
		return ndl.FindIllustrations(ctx, h.HC, base, p.Keyword, max1(p.Limit, 5))

	case "resolve_ndl_authority":
		var p struct {
			Term string
			Id   string
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		ep := os.Getenv("NDLA_SPARQL_ENDPOINT")
		return ndl.ResolveAuthority(ctx, h.HC, ep, p.Term, p.Id)
	}
	return nil, errors.New("unknown tool: " + name)
}

func max1(v, d int) int {
	if v <= 0 {
		return d
	}
	return v
}
