package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/yourname/yokai-finder-mcp/internal/cache"
	"github.com/yourname/yokai-finder-mcp/internal/ndl"
	"github.com/yourname/yokai-finder-mcp/pkg/types"
)

// Handler coordinates Yokai search requests against the NDL API with caching.
type Handler struct {
	ndlClient *ndl.Client
	cache     *cache.Cache
}

// New creates a handler with the provided dependencies.
func New(ndlClient *ndl.Client, cache *cache.Cache) *Handler {
	return &Handler{
		ndlClient: ndlClient,
		cache:     cache,
	}
}

// SearchYokai finds yokai related literature via the NDL OpenSearch API.
func (h *Handler) SearchYokai(ctx context.Context, params types.YokaiSearchParams) (*types.YokaiSearchResult, error) {
	if h == nil || h.ndlClient == nil {
		return nil, errors.New("handler not initialised")
	}

	cleaned := normaliseParams(params)

	if h.cache != nil {
		if result, ok := h.cache.Get(cleaned); ok {
			return result, nil
		}
	}

	result, err := h.ndlClient.SearchYokaiBooks(ctx, cleaned)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("ndl returned no data")
	}

	if h.cache != nil {
		h.cache.Set(cleaned, result)
	}
	return result, nil
}

func normaliseParams(p types.YokaiSearchParams) types.YokaiSearchParams {
	p.Name = strings.TrimSpace(p.Name)
	p.Region = strings.TrimSpace(p.Region)
	p.Category = strings.TrimSpace(p.Category)

	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 50 {
		p.Limit = 50
	}
	return p
}
