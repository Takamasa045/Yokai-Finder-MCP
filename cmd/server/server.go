package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/ndl"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/version"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/yokai"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName        = "yokai-finder-mcp"
	defaultCacheTTL   = 5 * time.Minute
	defaultCacheSize  = 256
	agentInstructions = `Yokai Finder helps users discover Japanese yokai, read folklore, and find related books.

Tool choice:
- Vague mood/theme ("scary", "water-y", "for kids") → suggest_yokai
- A specific name (河童, Kappa, カッパ) → get_yokai
- Browse/filter the roster → list_yokai (category 水系/water, tag, tone, famousRank all work)
- Featured daily pick → yokai_of_the_day
- Books from the National Diet Library → search_yokai_books
- Related or side-by-side entries → related_yokai / compare_yokai
- list_curated_yokai is deprecated; use list_yokai with hasProfile=true

Prefer Japanese names in answers. Category values are Japanese keys (水系, 山系). hasProfile=true means a full bilingual encyclopedia card is available via get_yokai.`
)

type searchArgs struct {
	Name         string `json:"name,omitempty" jsonschema:"Yokai name or keyword to search for"`
	Region       string `json:"region,omitempty" jsonschema:"Region or place associated with the yokai"`
	Category     string `json:"category,omitempty" jsonschema:"Yokai category or theme"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Maximum number of books to return (default 10, max 50)"`
	VerifyCovers bool   `json:"verifyCovers,omitempty" jsonschema:"If true, HEAD-check cover URLs and drop dead links (slower)"`
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
	Term          string `json:"term,omitempty" jsonschema:"Keyword to match name, category, region, or Japanese blurb"`
	Category      string `json:"category,omitempty" jsonschema:"Category hint (e.g. 水系, water, 付喪神, 狐狸)"`
	Region        string `json:"region,omitempty" jsonschema:"Region hint (e.g. 東北, tohoku, 九州, 海)"`
	Tag           string `json:"tag,omitempty" jsonschema:"Tag filter (e.g. 怖い, かわいい, 入門)"`
	Tone          string `json:"tone,omitempty" jsonschema:"Tone filter: gentle, comic, horror, solemn, tragic, mysterious, playful"`
	FamousRankMin int    `json:"famousRankMin,omitempty" jsonschema:"Minimum famousRank (1=iconic, 5=obscure)"`
	FamousRankMax int    `json:"famousRankMax,omitempty" jsonschema:"Maximum famousRank"`
	HasProfile    *bool  `json:"hasProfile,omitempty" jsonschema:"If true, only yokai with full encyclopedia cards"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum entries to return (default 200, max 200)"`
}

type relatedYokaiArgs struct {
	Name  string `json:"name" jsonschema:"Japanese or English yokai name"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum neighbours to return (default 6, max 20)"`
}

type compareYokaiArgs struct {
	Left  string `json:"left" jsonschema:"First yokai name"`
	Right string `json:"right" jsonschema:"Second yokai name"`
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
	listen := flag.String("http", "", "optional streamable HTTP address (example: 127.0.0.1:8080). Empty means stdio.")
	flag.Parse()

	if err := run(context.Background(), *listen); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func run(ctx context.Context, listen string) error {
	c := cache.NewCache(defaultCacheTTL, defaultCacheSize)
	defer c.Stop()
	h := handler.New(ndl.NewClient(), c)

	server := newServer(h)
	if listen != "" {
		return serveHTTP(ctx, listen, server)
	}
	return server.Run(ctx, &mcp.StdioTransport{})
}

func newServer(h *handler.Handler) *mcp.Server {
	ver := os.Getenv("YOKAI_FINDER_VERSION")
	if ver == "" {
		ver = version.Version
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: ver,
		Title:   "Yokai Finder",
	}, &mcp.ServerOptions{
		Instructions:      agentInstructions,
		CompletionHandler: completeYokaiNames,
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_yokai_books",
		Description: "国立国会図書館（NDL）で妖怪関連の書籍を検索する。妖怪名・地域・カテゴリで絞り込め、ISBNがある書籍には書影URL候補（coverImageCandidates）も付く。Search the National Diet Library for yokai-related books; results include cover-image URL candidates when an ISBN is available.",
		Annotations: openWorldTool(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, *types.YokaiSearchResult, error) {
		params := types.YokaiSearchParams{
			Name:         args.Name,
			Region:       args.Region,
			Category:     args.Category,
			Limit:        args.Limit,
			VerifyCovers: args.VerifyCovers,
		}

		result, err := h.SearchYokai(ctx, params)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "yokai_of_the_day",
		Description: "「今日の妖怪」を紹介する。引数なしなら日替わり（JST）で1体を選び、伝承・特徴・創作フック・おすすめ書籍・ストーリープロンプトをまとめて返す。特定の妖怪は name、別の妖怪を引きたいときは seed を指定。Daily featured yokai (JST) with lore, creative hooks, recommended reading, and a story prompt.",
		Annotations: openWorldTool(),
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
		Description: "妖怪索引（160体超）を一覧・検索する。category は 水系 でも water でも可。tag / tone / famousRank / hasProfile で絞り込み。雰囲気だけの質問は suggest_yokai。Browse the yokai roster; Japanese and English category hints both work.",
		Annotations: localTool(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args listYokaiArgs) (*mcp.CallToolResult, *types.YokaiIndexResult, error) {
		params := types.YokaiIndexParams{
			Term:          args.Term,
			Category:      args.Category,
			Region:        args.Region,
			Tag:           args.Tag,
			Tone:          args.Tone,
			FamousRankMin: args.FamousRankMin,
			FamousRankMax: args.FamousRankMax,
			HasProfile:    args.HasProfile,
			Limit:         args.Limit,
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
		Annotations: localTool(),
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
		Description: "妖怪を日本語名・英語名・別名（カッパ、かわっぱ など）で1体取得する。図鑑があれば source=profile。見つからないときは suggestions（もしかして）を返す。Look up one yokai by Japanese, English, or alias.",
		Annotations: localTool(),
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
		Description: "非推奨。詳細図鑑の一覧は list_yokai に hasProfile=true を付けてください。Deprecated: use list_yokai with hasProfile=true. Still returns bilingual encyclopedia cards for compatibility.",
		Annotations: localTool(),
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
		Annotations: localTool(),
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "related_yokai",
		Description: "指定した妖怪に近い索引エントリを返す。タグ・カテゴリ・トーンの重なりで近傍を探す。Find neighbouring yokai by shared tags, category, and tone.",
		Annotations: localTool(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args relatedYokaiArgs) (*mcp.CallToolResult, *types.RelatedYokaiResult, error) {
		result, err := h.RelatedYokai(ctx, types.RelatedYokaiParams{Name: args.Name, Limit: args.Limit})
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "compare_yokai",
		Description: "2体の妖怪を並べて共通点・相違点を返す。Compare two yokai side by side.",
		Annotations: localTool(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args compareYokaiArgs) (*mcp.CallToolResult, *types.CompareYokaiResult, error) {
		result, err := h.CompareYokai(ctx, types.CompareYokaiParams{Left: args.Left, Right: args.Right})
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	registerResources(server, h)
	registerPrompts(server)

	return server
}

func localTool() *mcp.ToolAnnotations {
	open := false
	return &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: &open}
}

func openWorldTool() *mcp.ToolAnnotations {
	open := true
	return &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: &open}
}

func completeYokaiNames(_ context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	prefix := ""
	if req != nil && req.Params.Argument.Name != "" {
		prefix = req.Params.Argument.Value
	}
	values := yokai.CompleteNames(prefix, 20)
	return &mcp.CompleteResult{
		Completion: mcp.CompletionResultDetails{
			Values:  values,
			Total:   len(values),
			HasMore: false,
		},
	}, nil
}
