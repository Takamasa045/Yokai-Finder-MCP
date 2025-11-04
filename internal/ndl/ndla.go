package ndl

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

// 将来的に SPARQL を叩く実装用の雛形。
// 未設定時は not_configured エラー。

func ResolveAuthority(ctx context.Context, httpc *http.Client, endpoint, term, id string) (*types.AuthorityResolution, error) {
	if endpoint == "" {
		return nil, errors.New("not_configured: set NDLA_SPARQL_ENDPOINT env")
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 15 * time.Second}
	}
	// TODO: term/id を使った SPARQL クエリを実装
	// ひとまずダミーで prefLabel=term を返す
	return &types.AuthorityResolution{Term: term, PrefLabel: term}, nil
}
