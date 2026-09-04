package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/cache"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
	"github.com/Takamasa045/Yokai-Finder-MCP/internal/ndl"
	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := cache.NewCache(5*time.Minute, 50)
	defer c.Stop()
	h := handler.New(ndl.NewClient(), c)
	result, err := h.SearchYokai(ctx, types.YokaiSearchParams{Name: "河童", Limit: 2})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Query=%s Total=%d\n", result.Query, result.Total)
	for i, book := range result.Results {
		fmt.Printf("%d: %s (%s)\n", i+1, book.Title, book.PublishDate)
	}
}
