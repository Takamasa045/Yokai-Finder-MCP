package ndl

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

func SearchFulltext(ctx context.Context, httpc *http.Client, base, query string, contentOnly bool, from, size int) ([]types.FulltextHit, error) {
	if base == "" {
		return nil, errors.New("not_configured: set NDL_LAB_BASE env")
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 20 * time.Second}
	}
	// TODO: base + "/book/search" に相当するパラメータで実装
	// ここでは空配列を返す（APIキー不要の公開エンドポイント前提）
	return []types.FulltextHit{}, nil
}
