package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL query for semantic/fulltext search (DataHub 1.3.x+).
const (
	// SemanticSearchQuery performs fulltext search across all entity types.
	// Uses searchAcrossEntities with searchFlags.fulltext=true for broader
	// natural language matching than the type-scoped search() query.
	SemanticSearchQuery = `
query semanticSearch($input: SearchAcrossEntitiesInput!) {
  searchAcrossEntities(input: $input) {
    start
    count
    total
    searchResults {
      entity {
        urn
        type
        ... on Dataset {
          name
          description
          platform {
            name
          }
          ownership {
            owners {
              owner {
                ... on CorpUser {
                  urn
                  username
                }
                ... on CorpGroup {
                  urn
                  name
                }
              }
              type
            }
          }
          tags {
            tags {
              tag {
                urn
                name
                description
              }
            }
          }
          domain {
            domain {
              urn
              properties {
                name
                description
              }
            }
          }
        }
        ... on Dashboard {
          dashboardId
          info {
            name
            description
          }
          platform {
            name
          }
        }
        ... on DataProduct {
          properties {
            name
            description
          }
        }
        ... on GlossaryTerm {
          properties {
            name
            description
          }
        }
      }
      matchedFields {
        name
        value
      }
    }
  }
}
`
)

// SemanticSearch performs fulltext search across entities.
// Uses searchAcrossEntities with fulltext mode for broader natural language
// matching. Returns an error (not empty results) when not supported because
// the caller explicitly chose semantic mode and should know it's unavailable.
func (c *Client) SemanticSearch(ctx context.Context, query string, opts ...SearchOption) (*types.SearchResult, error) {
	options := &searchOptions{
		limit:  c.config.DefaultLimit,
		offset: 0,
	}
	for _, opt := range opts {
		opt(options)
	}

	if options.limit > c.config.MaxLimit {
		options.limit = c.config.MaxLimit
	}

	input := map[string]any{
		"query":       query,
		"start":       options.offset,
		"count":       options.limit,
		"searchFlags": map[string]any{"fulltext": true},
	}

	if options.entityType != "" {
		input["types"] = []string{options.entityType}
	}

	variables := map[string]any{
		"input": input,
	}

	var response struct {
		SearchAcrossEntities struct {
			Start         int                `json:"start"`
			Count         int                `json:"count"`
			Total         int                `json:"total"`
			SearchResults []searchResultItem `json:"searchResults"`
		} `json:"searchAcrossEntities"`
	}

	if err := c.Execute(ctx, SemanticSearchQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("SemanticSearch(%q): %w", query, err)
	}

	result := &types.SearchResult{
		Total:  response.SearchAcrossEntities.Total,
		Offset: response.SearchAcrossEntities.Start,
		Limit:  response.SearchAcrossEntities.Count,
	}

	for _, sr := range response.SearchAcrossEntities.SearchResults {
		result.Entities = append(result.Entities, parseSearchResult(sr))
	}

	return result, nil
}
