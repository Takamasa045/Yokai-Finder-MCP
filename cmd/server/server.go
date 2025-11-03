package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourname/yokai-finder-mcp/internal/cache"
	"github.com/yourname/yokai-finder-mcp/internal/handler"
	"github.com/yourname/yokai-finder-mcp/internal/ndl"
	"github.com/yourname/yokai-finder-mcp/pkg/types"
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

	return server
}
