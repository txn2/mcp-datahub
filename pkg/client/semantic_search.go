package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL query for semantic search (DataHub 1.4.x + OpenSearch 2.19.3+).
const (
	// SemanticSearchQuery performs natural language semantic search across entities.
	SemanticSearchQuery = `
query semanticSearch($input: SemanticSearchInput!) {
  semanticSearchAcrossEntities(input: $input) {
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

// SemanticSearch performs natural language semantic search across entities.
// Requires DataHub 1.4.x with OpenSearch 2.19.3+. Returns an error (not empty
// results) when not supported because the caller explicitly chose semantic mode
// and should know it's unavailable.
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
		"query": query,
		"start": options.offset,
		"count": options.limit,
	}

	if options.entityType != "" {
		input["types"] = []string{options.entityType}
	}

	variables := map[string]any{
		"input": input,
	}

	var response struct {
		SemanticSearchAcrossEntities struct {
			Start         int `json:"start"`
			Count         int `json:"count"`
			Total         int `json:"total"`
			SearchResults []struct {
				Entity struct {
					URN         string `json:"urn"`
					Type        string `json:"type"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Platform    struct {
						Name string `json:"name"`
					} `json:"platform"`
					Properties struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"properties"`
					DashboardID string `json:"dashboardId"`
					Info        struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"info"`
					Ownership struct {
						Owners []struct {
							Owner struct {
								URN      string `json:"urn"`
								Username string `json:"username"`
								Name     string `json:"name"`
							} `json:"owner"`
							Type string `json:"type"`
						} `json:"owners"`
					} `json:"ownership"`
					Tags struct {
						Tags []struct {
							Tag struct {
								URN         string `json:"urn"`
								Name        string `json:"name"`
								Description string `json:"description"`
							} `json:"tag"`
						} `json:"tags"`
					} `json:"tags"`
					Domain struct {
						Domain struct {
							URN        string `json:"urn"`
							Properties struct {
								Name        string `json:"name"`
								Description string `json:"description"`
							} `json:"properties"`
						} `json:"domain"`
					} `json:"domain"`
				} `json:"entity"`
				MatchedFields []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"matchedFields"`
			} `json:"searchResults"`
		} `json:"semanticSearchAcrossEntities"`
	}

	if err := c.Execute(ctx, SemanticSearchQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("SemanticSearch(%q): %w", query, err)
	}

	result := &types.SearchResult{
		Total:  response.SemanticSearchAcrossEntities.Total,
		Offset: response.SemanticSearchAcrossEntities.Start,
		Limit:  response.SemanticSearchAcrossEntities.Count,
	}

	for _, sr := range response.SemanticSearchAcrossEntities.SearchResults {
		name := sr.Entity.Name
		description := sr.Entity.Description
		if sr.Entity.Properties.Name != "" {
			name = sr.Entity.Properties.Name
		}
		if sr.Entity.Properties.Description != "" {
			description = sr.Entity.Properties.Description
		}
		if sr.Entity.Info.Name != "" {
			name = sr.Entity.Info.Name
		}
		if sr.Entity.Info.Description != "" {
			description = sr.Entity.Info.Description
		}

		entity := types.SearchEntity{
			URN:         sr.Entity.URN,
			Type:        sr.Entity.Type,
			Name:        name,
			Description: description,
			Platform:    sr.Entity.Platform.Name,
		}

		for _, o := range sr.Entity.Ownership.Owners {
			ownerName := o.Owner.Username
			if o.Owner.Name != "" {
				ownerName = o.Owner.Name
			}
			entity.Owners = append(entity.Owners, types.Owner{
				URN:  o.Owner.URN,
				Name: ownerName,
				Type: types.OwnershipType(o.Type),
			})
		}

		for _, t := range sr.Entity.Tags.Tags {
			entity.Tags = append(entity.Tags, types.Tag{
				URN:         t.Tag.URN,
				Name:        t.Tag.Name,
				Description: t.Tag.Description,
			})
		}

		if sr.Entity.Domain.Domain.URN != "" {
			entity.Domain = &types.Domain{
				URN:         sr.Entity.Domain.Domain.URN,
				Name:        sr.Entity.Domain.Domain.Properties.Name,
				Description: sr.Entity.Domain.Domain.Properties.Description,
			}
		}

		for _, mf := range sr.MatchedFields {
			entity.MatchedFields = append(entity.MatchedFields, types.MatchedField{
				Name:  mf.Name,
				Value: mf.Value,
			})
		}

		result.Entities = append(result.Entities, entity)
	}

	return result, nil
}
