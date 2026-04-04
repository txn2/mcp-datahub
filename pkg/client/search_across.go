package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// SearchAcrossEntitiesQuery searches across multiple entity types with advanced filtering.
// Uses the searchAcrossEntities GraphQL endpoint which supports orFilters and multi-type search.
const SearchAcrossEntitiesQuery = `
query searchAcrossEntities($input: SearchAcrossEntitiesInput!) {
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
        ... on DataFlow {
          flowId
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
        ... on Tag {
          properties {
            name
            description
          }
        }
        ... on Document {
          subType
          info {
            title
            contents {
              text
            }
            status {
              state
            }
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

// SearchAcrossEntities searches across entity types with optional advanced filtering.
// Supports multi-type search and field-level filters (e.g., fieldPaths, fieldTags, platform).
// Available in DataHub 1.3.x+.
func (c *Client) SearchAcrossEntities(ctx context.Context, query string, opts ...SearchOption) (*types.SearchResult, error) {
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
		"query": query,
		"start": options.offset,
		"count": options.limit,
	}

	if len(options.types) > 0 {
		input["types"] = options.types
	} else if options.entityType != "" {
		input["types"] = []string{options.entityType}
	}

	if len(options.orFilters) > 0 {
		input["orFilters"] = buildOrFilters(options.orFilters)
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

	if err := c.Execute(ctx, SearchAcrossEntitiesQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("SearchAcrossEntities(%q): %w", query, err)
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

// buildOrFilters converts SearchFilter slices into the GraphQL orFilters structure.
// All filters are AND'd together in a single filter group.
func buildOrFilters(filters []SearchFilter) []map[string]any {
	andConditions := make([]map[string]any, 0, len(filters))

	for _, f := range filters {
		condition := map[string]any{
			"field":  f.Field,
			"values": f.Values,
		}

		if f.Condition != "" {
			condition["condition"] = f.Condition
		}

		if f.Negated {
			condition["negated"] = true
		}

		andConditions = append(andConditions, condition)
	}

	return []map[string]any{
		{"and": andConditions},
	}
}
