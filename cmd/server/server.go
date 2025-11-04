package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/ndl"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName       = "yokai-finder-mcp"
	defaultVersion   = "0.1.0"
	defaultCacheTTL  = 5 * time.Minute
	defaultCacheSize = 256
)

type searchArgs struct {
	Name     string `json:"name,omitempty" description:"Yokai name or keyword to search for"`
	Region   string `json:"region,omitempty" description:"Region or place associated with the yokai"`
	Category string `json:"category,omitempty" description:"Yokai category or theme"`
	Limit    int    `json:"limit,omitempty" description:"Maximum number of books to return (default 10, max 50)"`
}

type yokaiOfTheDayArgs struct {
	Name     string `json:"name,omitempty" description:"Exact yokai to highlight"`
	Category string `json:"category,omitempty" description:"Filter curated yokai by category hint"`
	Region   string `json:"region,omitempty" description:"Filter curated yokai by region hint"`
	Seed     int64  `json:"seed,omitempty" description:"Deterministic selection seed"`
	Limit    int    `json:"limit,omitempty" description:"Maximum number of book recommendations (default 5, max 10)"`
}

type listCuratedArgs struct {
	Term                 string `json:"term,omitempty" description:"Keyword to match name, lore, or motifs"`
	Category             string `json:"category,omitempty" description:"Filter curated yokai by category hint"`
	Region               string `json:"region,omitempty" description:"Filter curated yokai by region hint"`
	Seed                 int64  `json:"seed,omitempty" description:"Shuffle results deterministically when provided"`
	Limit                int    `json:"limit,omitempty" description:"Maximum number of curated entries to return (default 10, max 50)"`
	IncludeLegends       bool   `json:"includeLegends,omitempty" description:"Include folkloric legend snippets"`
	IncludeTraits        bool   `json:"includeTraits,omitempty" description:"Include notable traits"`
	IncludeMotifs        bool   `json:"includeMotifs,omitempty" description:"Include thematic motifs"`
	IncludeCreativeHooks bool   `json:"includeCreativeHooks,omitempty" description:"Include creative hook suggestions"`
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
		Description: "Search the NDL OpenSearch API for yokai related literature",
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
		Description: "Surface a curated yokai profile with lore, creative hooks, and recommended reading",
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
		Name:        "list_curated_yokai",
		Description: "Browse curated yokai profiles with optional lore snippets and filters",
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

	return server
}
