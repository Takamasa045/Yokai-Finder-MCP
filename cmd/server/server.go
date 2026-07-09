package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/ndl"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName       = "yokai-finder-mcp"
	defaultVersion   = "0.3.0"
	defaultCacheTTL  = 5 * time.Minute
	defaultCacheSize = 256
)

type searchArgs struct {
	Name     string `json:"name,omitempty" jsonschema:"Yokai name or keyword to search for"`
	Region   string `json:"region,omitempty" jsonschema:"Region or place associated with the yokai"`
	Category string `json:"category,omitempty" jsonschema:"Yokai category or theme"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum number of books to return (default 10, max 50)"`
}

type yokaiOfTheDayArgs struct {
	Name     string `json:"name,omitempty" jsonschema:"Exact yokai to highlight"`
	Category string `json:"category,omitempty" jsonschema:"Filter curated yokai by category hint"`
	Region   string `json:"region,omitempty" jsonschema:"Filter curated yokai by region hint"`
	Seed     int64  `json:"seed,omitempty" jsonschema:"Deterministic selection seed"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum number of book recommendations (default 5, max 10)"`
}

type coverThumbnailArgs struct {
	ISBN string `json:"isbn,omitempty" jsonschema:"ISBN-10 or ISBN-13 (hyphens/spaces allowed)"`
	JPNo string `json:"jpno,omitempty" jsonschema:"JP number (全国書誌番号) used as a fallback identifier"`
}

type listCuratedArgs struct {
	Term                 string `json:"term,omitempty" jsonschema:"Keyword to match name, lore, or motifs"`
	Category             string `json:"category,omitempty" jsonschema:"Filter curated yokai by category hint"`
	Region               string `json:"region,omitempty" jsonschema:"Filter curated yokai by region hint"`
	Seed                 int64  `json:"seed,omitempty" jsonschema:"Shuffle results deterministically when provided"`
	Limit                int    `json:"limit,omitempty" jsonschema:"Maximum number of curated entries to return (default 10, max 50)"`
	IncludeLegends       bool   `json:"includeLegends,omitempty" jsonschema:"Include folkloric legend snippets"`
	IncludeTraits        bool   `json:"includeTraits,omitempty" jsonschema:"Include notable traits"`
	IncludeMotifs        bool   `json:"includeMotifs,omitempty" jsonschema:"Include thematic motifs"`
	IncludeCreativeHooks bool   `json:"includeCreativeHooks,omitempty" jsonschema:"Include creative hook suggestions"`
}

type listYokaiArgs struct {
	Term     string `json:"term,omitempty" jsonschema:"Keyword to match name, category, region, or Japanese blurb"`
	Category string `json:"category,omitempty" jsonschema:"Category hint (e.g. 水系, 付喪神, 狐狸)"`
	Region   string `json:"region,omitempty" jsonschema:"Region hint (e.g. 東北, 九州, 海)"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum entries to return (default 200, max 200)"`
}

type suggestYokaiArgs struct {
	Vibe     string `json:"vibe,omitempty" jsonschema:"Mood or feeling (e.g. 怖い, かわいい, 不気味, ほのぼの)"`
	Theme    string `json:"theme,omitempty" jsonschema:"Theme or motif (e.g. 水, 狐, 付喪神, 呪い)"`
	Setting  string `json:"setting,omitempty" jsonschema:"Setting or place vibe (e.g. 山, 川, 夜の町, 学校)"`
	Audience string `json:"audience,omitempty" jsonschema:"Intended audience or use (e.g. 子ども, 創作向け, ホラー)"`
	Term     string `json:"term,omitempty" jsonschema:"Free-form keyword when vibe/theme are hard to pin down"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum suggestions to return (default 6, max 20)"`
	Seed     int64  `json:"seed,omitempty" jsonschema:"Deterministic shuffle seed for alternate suggestions"`
}

type getYokaiArgs struct {
	Name string `json:"name" jsonschema:"Japanese or English yokai name to look up (e.g. 河童, Kappa, 八岐大蛇)"`
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func run(ctx context.Context) error {
	h := handler.New(ndl.NewClient(), cache.NewCache(defaultCacheTTL, defaultCacheSize))

	server := newServer(h)
	return server.Run(ctx, &mcp.StdioTransport{})
}

func newServer(h *handler.Handler) *mcp.Server {
	version := os.Getenv("YOKAI_FINDER_VERSION")
	if version == "" {
		version = defaultVersion
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_yokai_books",
		Description: "国立国会図書館（NDL）で妖怪関連の書籍を検索する。妖怪名・地域・カテゴリで絞り込め、ISBNがある書籍には書影URL候補（coverImageCandidates）も付く。Search the National Diet Library for yokai-related books; results include cover-image URL candidates when an ISBN is available.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, *types.YokaiSearchResult, error) {
		params := types.YokaiSearchParams{
			Name:     args.Name,
			Region:   args.Region,
			Category: args.Category,
			Limit:    args.Limit,
		}

		result, err := h.SearchYokai(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "yokai_of_the_day",
		Description: "「今日の妖怪」を紹介する。引数なしなら日替わりで1体を選び、伝承・特徴・創作フック・おすすめ書籍・ストーリープロンプトをまとめて返す。特定の妖怪は name、別の妖怪を引きたいときは seed を指定。Daily featured yokai with lore, creative hooks, recommended reading, and a story prompt; deterministic per calendar day unless a seed or name is given.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args yokaiOfTheDayArgs) (*mcp.CallToolResult, *types.YokaiOfTheDayResult, error) {
		params := types.YokaiOfTheDayParams{
			Name:     args.Name,
			Category: args.Category,
			Region:   args.Region,
			Seed:     args.Seed,
			Limit:    args.Limit,
		}

		result, err := h.YokaiOfTheDay(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_yokai",
		Description: "妖怪索引（160体超）をざっくり一覧・検索する。名前・一言紹介・カテゴリ・タグの軽量リスト。雰囲気や曖昧な希望（「怖い妖怪」「水のやつ」など）なら suggest_yokai を使う。気になった名前は get_yokai で詳細、search_yokai_books で本を、hasProfile=true なら list_curated_yokai / yokai_of_the_day で深掘り。Browse the yokai name index (compact blurbs); use suggest_yokai for vague discovery.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args listYokaiArgs) (*mcp.CallToolResult, *types.YokaiIndexResult, error) {
		params := types.YokaiIndexParams{
			Term:     args.Term,
			Category: args.Category,
			Region:   args.Region,
			Limit:    args.Limit,
		}

		result, err := h.ListYokai(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "suggest_yokai",
		Description: "曖昧な条件から妖怪を提案する。例: 「怖い妖怪」「水のやつ」「創作向け」「かわいいが不気味」。vibe/theme/setting/audience や自由語 term で候補と短い理由（whySuggested）を返す。名前が分かったら get_yokai、網羅一覧は list_yokai。Suggest yokai from vague queries (scary, water-y, for fiction); returns short rationales.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args suggestYokaiArgs) (*mcp.CallToolResult, *types.SuggestYokaiResult, error) {
		params := types.SuggestYokaiParams{
			Vibe:     args.Vibe,
			Theme:    args.Theme,
			Setting:  args.Setting,
			Audience: args.Audience,
			Term:     args.Term,
			Limit:    args.Limit,
			Seed:     args.Seed,
		}

		result, err := h.SuggestYokai(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_yokai",
		Description: "妖怪を日本語名または英語名で1体取得する。キュレーション済みなら詳細プロフィール（source=profile）、索引のみなら軽いカード（source=index）。見つからないときは list_yokai / suggest_yokai / search_yokai_books を案内。Look up one yokai by Japanese or English name (full profile or index card).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args getYokaiArgs) (*mcp.CallToolResult, *types.GetYokaiResult, error) {
		params := types.GetYokaiParams{
			Name: args.Name,
		}

		result, err := h.GetYokai(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_curated_yokai",
		Description: "キュレーション済み妖怪図鑑（詳細プロフィール50体）を一覧・検索する。伝承・特徴・創作フック付き。ざっくり顔ぶれを知りたいときは list_yokai を先に使う。Browse the deep curated encyclopedia (50 full bilingual profiles); use list_yokai first for a broad roster.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args listCuratedArgs) (*mcp.CallToolResult, *types.CuratedYokaiResult, error) {
		params := types.CuratedYokaiParams{
			Term:                 args.Term,
			Category:             args.Category,
			Region:               args.Region,
			Seed:                 args.Seed,
			Limit:                args.Limit,
			IncludeLegends:       args.IncludeLegends,
			IncludeTraits:        args.IncludeTraits,
			IncludeMotifs:        args.IncludeMotifs,
			IncludeCreativeHooks: args.IncludeCreativeHooks,
		}

		result, err := h.ListCuratedYokai(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_cover_thumbnail",
		Description: "ISBN または全国書誌番号（JP番号）から書影（表紙画像）のURL候補を返す。候補は上から順に試すこと（資料により提供有無が異なる）。Build cover-image URL candidates from an ISBN (10 or 13) or JP number; try candidates in order.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args coverThumbnailArgs) (*mcp.CallToolResult, *ndl.CoverURLs, error) {
		isbn := strings.TrimSpace(args.ISBN)
		jpno := strings.TrimSpace(args.JPNo)
		if isbn == "" && jpno == "" {
			return nil, nil, errors.New("provide at least one of isbn or jpno")
		}
		if isbn != "" && ndl.NormalizeISBN13(isbn) == "" {
			return nil, nil, errors.New("isbn could not be parsed: use ISBN-10 or ISBN-13, hyphens allowed")
		}

		covers := ndl.BuildCoverURLs(isbn, jpno)
		return nil, &covers, nil
	})

	return server
}
