package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// documentSelectionFields is the shared GraphQL selection set for a Document
// entity (everything after urn/type). Reused by GetDocumentQuery and
// SearchDocumentsQuery so single-document reads and document search return
// identical metadata from a single source of truth.
const documentSelectionFields = `
    subType
    info {
      title
      contents {
        text
      }
      source {
        sourceType
        externalUrl
      }
      status {
        state
      }
      created {
        time
      }
      lastModified {
        time
      }
      relatedAssets {
        asset {
          urn
        }
      }
      relatedDocuments {
        document {
          urn
        }
      }
      parentDocument {
        document {
          urn
        }
      }
    }
    settings {
      showInGlobalContext
    }
    ownership {
      owners {
        owner {
          ... on CorpUser {
            urn
            username
            info {
              displayName
              email
            }
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
          properties {
            name
            description
          }
        }
      }
    }
    glossaryTerms {
      terms {
        term {
          urn
          properties {
            name
            description
          }
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
`

// GraphQL queries for context documents (DataHub 1.4.x+).
const (
	// GetDocumentQuery retrieves a single document by URN.
	GetDocumentQuery = `
query getDocument($urn: String!) {
  document(urn: $urn) {
    urn
    type` + documentSelectionFields + `  }
}
`

	// GetRelatedDocumentsQuery retrieves documents linked to an entity.
	// The relatedDocuments field is repeated per entity type because GraphQL
	// inline fragments require each concrete type to declare its own fields —
	// there is no shared interface to query against. Only entity types that
	// support relatedDocuments in DataHub are included: Dataset, GlossaryTerm,
	// GlossaryNode, and Container.
	GetRelatedDocumentsQuery = `
query getRelatedDocuments($urn: String!, $input: RelatedDocumentsInput!) {
  entity(urn: $urn) {
    ... on Dataset {
      relatedDocuments(input: $input) {
        total
        documents {
          urn` + documentSelectionFields + `        }
      }
    }
    ... on GlossaryTerm {
      relatedDocuments(input: $input) {
        total
        documents {
          urn` + documentSelectionFields + `        }
      }
    }
    ... on GlossaryNode {
      relatedDocuments(input: $input) {
        total
        documents {
          urn` + documentSelectionFields + `        }
      }
    }
    ... on Container {
      relatedDocuments(input: $input) {
        total
        documents {
          urn` + documentSelectionFields + `        }
      }
    }
  }
}
`
)

// GetDocument retrieves a document by URN.
func (c *Client) GetDocument(ctx context.Context, urn string) (*types.Document, error) {
	variables := map[string]any{"urn": urn}

	var response struct {
		Document documentResponse `json:"document"`
	}

	if err := c.Execute(ctx, GetDocumentQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetDocument(%s): %w", urn, err)
	}

	if response.Document.URN == "" {
		return nil, fmt.Errorf("GetDocument(%s): %w", urn, ErrNotFound)
	}

	return parseDocumentResponse(&response.Document), nil
}

// GetRelatedDocuments retrieves documents linked to an entity.
func (c *Client) GetRelatedDocuments(ctx context.Context, urn string) ([]types.Document, error) {
	variables := map[string]any{
		"urn":   urn,
		"input": map[string]any{"start": 0, "count": c.config.MaxLimit},
	}

	var response struct {
		Entity struct {
			RelatedDocuments *relatedDocumentsResult `json:"relatedDocuments"`
		} `json:"entity"`
	}

	if err := c.Execute(ctx, GetRelatedDocumentsQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetRelatedDocuments(%s): %w", urn, err)
	}

	if response.Entity.RelatedDocuments == nil {
		return nil, nil
	}

	docs := make([]types.Document, 0, len(response.Entity.RelatedDocuments.Documents))
	for _, d := range response.Entity.RelatedDocuments.Documents {
		docs = append(docs, *parseDocumentResponse(&d))
	}

	return docs, nil
}

// relatedDocumentsResult matches the GraphQL RelatedDocumentsResult type.
type relatedDocumentsResult struct {
	Total     int                `json:"total"`
	Documents []documentResponse `json:"documents"`
}

// documentResponse is the internal GraphQL response shape for a document.
type documentResponse struct {
	URN     string `json:"urn"`
	Type    string `json:"type"`
	SubType string `json:"subType"`
	Info    struct {
		Title    string `json:"title"`
		Contents struct {
			Text string `json:"text"`
		} `json:"contents"`
		Source *struct {
			SourceType  string `json:"sourceType"`
			ExternalURL string `json:"externalUrl"`
		} `json:"source"`
		Status struct {
			State string `json:"state"`
		} `json:"status"`
		Created struct {
			Time int64 `json:"time"`
		} `json:"created"`
		LastModified struct {
			Time int64 `json:"time"`
		} `json:"lastModified"`
		RelatedAssets []struct {
			Asset struct {
				URN string `json:"urn"`
			} `json:"asset"`
		} `json:"relatedAssets"`
		RelatedDocuments []struct {
			Document struct {
				URN string `json:"urn"`
			} `json:"document"`
		} `json:"relatedDocuments"`
		ParentDocument *struct {
			Document struct {
				URN string `json:"urn"`
			} `json:"document"`
		} `json:"parentDocument"`
	} `json:"info"`
	Settings *struct {
		ShowInGlobalContext bool `json:"showInGlobalContext"`
	} `json:"settings"`
	Ownership struct {
		Owners []struct {
			Owner struct {
				URN      string `json:"urn"`
				Username string `json:"username"`
				Name     string `json:"name"`
				Info     struct {
					DisplayName string `json:"displayName"`
					Email       string `json:"email"`
				} `json:"info"`
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
				Properties  struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"properties"`
			} `json:"tag"`
		} `json:"tags"`
	} `json:"tags"`
	GlossaryTerms struct {
		Terms []struct {
			Term struct {
				URN        string `json:"urn"`
				Properties struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"properties"`
			} `json:"term"`
		} `json:"terms"`
	} `json:"glossaryTerms"`
	Domain struct {
		Domain struct {
			URN        string `json:"urn"`
			Properties struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"properties"`
		} `json:"domain"`
	} `json:"domain"`
}

// parseDocumentResponse converts the GraphQL response into a types.Document.
func parseDocumentResponse(d *documentResponse) *types.Document {
	doc := &types.Document{
		URN:     d.URN,
		Title:   d.Info.Title,
		Content: d.Info.Contents.Text,
		Status:  d.Info.Status.State,
		SubType: d.SubType,
		Created: d.Info.Created.Time,
	}

	if d.Info.LastModified.Time > 0 {
		doc.LastModified = d.Info.LastModified.Time
	}

	if d.Info.Source != nil {
		doc.Source = &types.DocumentSource{
			SourceType:  d.Info.Source.SourceType,
			ExternalURL: d.Info.Source.ExternalURL,
		}
	}

	if d.Settings != nil {
		doc.Settings = &types.DocumentSettings{
			ShowInGlobalContext: d.Settings.ShowInGlobalContext,
		}
	}

	for _, a := range d.Info.RelatedAssets {
		doc.RelatedAssets = append(doc.RelatedAssets, types.DocumentRelatedAsset{URN: a.Asset.URN})
	}

	for _, rd := range d.Info.RelatedDocuments {
		doc.RelatedDocuments = append(doc.RelatedDocuments, types.DocumentRelatedDocument{URN: rd.Document.URN})
	}

	if d.Info.ParentDocument != nil {
		doc.ParentDocument = &types.DocumentParent{URN: d.Info.ParentDocument.Document.URN}
	}

	// Parse ownership
	for _, o := range d.Ownership.Owners {
		name := firstNonEmpty(o.Owner.Info.DisplayName, o.Owner.Name, o.Owner.Username)
		doc.Owners = append(doc.Owners, types.Owner{
			URN:   o.Owner.URN,
			Name:  name,
			Email: o.Owner.Info.Email,
			Type:  types.OwnershipType(o.Type),
		})
	}

	// Parse tags. Prefer properties.name/description over the deprecated
	// key-derived top-level fields (UUIDs for tags created without an explicit id).
	for _, t := range d.Tags.Tags {
		doc.Tags = append(doc.Tags, types.Tag{
			URN:         t.Tag.URN,
			Name:        firstNonEmpty(t.Tag.Properties.Name, t.Tag.Name),
			Description: firstNonEmpty(t.Tag.Properties.Description, t.Tag.Description),
		})
	}

	// Parse glossary terms
	for _, gt := range d.GlossaryTerms.Terms {
		doc.GlossaryTerms = append(doc.GlossaryTerms, types.GlossaryTerm{
			URN:         gt.Term.URN,
			Name:        gt.Term.Properties.Name,
			Description: gt.Term.Properties.Description,
		})
	}

	// Parse domain
	if d.Domain.Domain.URN != "" {
		doc.Domain = &types.Domain{
			URN:         d.Domain.Domain.URN,
			Name:        d.Domain.Domain.Properties.Name,
			Description: d.Domain.Domain.Properties.Description,
		}
	}

	return doc
}
