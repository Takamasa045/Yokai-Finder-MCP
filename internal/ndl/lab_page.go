package ndl

import (
	"context"
	"errors"
	"net/http"
)

func SearchInBook(ctx context.Context, httpc *http.Client, base, pid, q string, from, size int) ([]int, error) {
	if base == "" {
		return nil, errors.New("not_configured: set NDL_LAB_BASE env")
	}
	// TODO: base + "/page/search" に相当する呼び出しを実装し、ヒットした page 番号配列を返す
	return []int{}, nil
}
