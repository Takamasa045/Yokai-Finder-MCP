package ndl

import (
	"context"
	"net/http"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/pkg/types"
)

func verifyCoverCandidates(ctx context.Context, books []types.YokaiBook, client *http.Client) {
	if client == nil {
		return
	}
	headClient := *client
	if headClient.Timeout == 0 || headClient.Timeout > 3*time.Second {
		headClient.Timeout = 3 * time.Second
	}
	for i := range books {
		books[i].CoverImageCandidates = liveCoverURLs(ctx, &headClient, books[i].CoverImageCandidates)
	}
}

func liveCoverURLs(ctx context.Context, client *http.Client, urls []string) []string {
	if len(urls) == 0 {
		return nil
	}
	out := make([]string, 0, len(urls))
	for _, raw := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, raw, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", userAgent())
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			out = append(out, raw)
		}
	}
	return out
}
