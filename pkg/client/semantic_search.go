package client

import (
	"context"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// SemanticSearchQuery is an alias for SearchAcrossEntitiesQuery.
// SemanticSearch uses the same GraphQL query but adds searchFlags.fulltext=true.
const SemanticSearchQuery = SearchAcrossEntitiesQuery

// semanticSearchFlags are the extra flags passed to searchAcrossEntities for fulltext mode.
var semanticSearchFlags = map[string]any{
	"searchFlags": map[string]any{"fulltext": true},
}

// SemanticSearch performs fulltext search across entities.
// Uses searchAcrossEntities with fulltext mode for broader natural language
// matching. Returns an error (not empty results) when not supported because
// the caller explicitly chose semantic mode and should know it's unavailable.
func (c *Client) SemanticSearch(ctx context.Context, query string, opts ...SearchOption) (*types.SearchResult, error) {
	return c.doSearchAcrossEntities(ctx, "SemanticSearch", query, semanticSearchFlags, opts)
}
