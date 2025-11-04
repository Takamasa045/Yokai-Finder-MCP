package ndl

import (
	"context"
	"errors"
	"net/http"

	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

func FindIllustrations(ctx context.Context, httpc *http.Client, base, keyword string, limit int) ([]types.IllustrationHit, error) {
	if base == "" {
		return nil, errors.New("not_configured: set NDL_LAB_BASE env")
	}
	// TODO: base + "/illustration/search" 相当を実装。bbox_pct 等を詰めて返却
	return []types.IllustrationHit{}, nil
}
