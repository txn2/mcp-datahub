package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL queries for context documents (DataHub 1.4.x+).
const (
	// GetDocumentQuery retrieves a single document by URN.
	GetDocumentQuery = `
query getDocument($urn: String!) {
  document(urn: $urn) {
    urn
    type
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
        urn
      }
      relatedDocuments {
        urn
      }
      parentDocument {
        urn
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
  }
}
`

	// SearchDocumentsQuery searches for documents with filtering.
	SearchDocumentsQuery = `
query searchDocuments($input: SearchDocumentsInput!) {
  searchDocuments(input: $input) {
    start
    count
    total
    searchResults {
      document {
        urn
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
            urn
          }
        }
        settings {
          showInGlobalContext
        }
      }
    }
  }
}
`

	// GetRelatedDocumentsQuery retrieves documents linked to an entity.
	GetRelatedDocumentsQuery = `
query getRelatedDocuments($urn: String!, $input: RelatedDocumentsInput!) {
  entity(urn: $urn) {
    ... on Dataset {
      relatedDocuments(input: $input) {
        total
        documents {
          urn
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
    }
    ... on GlossaryTerm {
      relatedDocuments(input: $input) {
        total
        documents {
          urn
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
    }
    ... on GlossaryNode {
      relatedDocuments(input: $input) {
        total
        documents {
          urn
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
    }
    ... on Container {
      relatedDocuments(input: $input) {
        total
        documents {
          urn
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
    }
  }
}
`

	// CreateDocumentMutation creates a new document.
	CreateDocumentMutation = `
mutation createDocument($input: CreateDocumentInput!) {
  createDocument(input: $input)
}
`

	// UpdateDocumentContentsMutation updates document content.
	UpdateDocumentContentsMutation = `
mutation updateDocumentContents($input: UpdateDocumentContentsInput!) {
  updateDocumentContents(input: $input)
}
`

	// UpdateDocumentRelatedEntitiesMutation updates related entities.
	UpdateDocumentRelatedEntitiesMutation = `
mutation updateDocumentRelatedEntities($input: UpdateDocumentRelatedEntitiesInput!) {
  updateDocumentRelatedEntities(input: $input)
}
`

	// UpdateDocumentStatusMutation updates document status.
	UpdateDocumentStatusMutation = `
mutation updateDocumentStatus($input: UpdateDocumentStatusInput!) {
  updateDocumentStatus(input: $input)
}
`

	// UpdateDocumentSettingsMutation updates document settings.
	UpdateDocumentSettingsMutation = `
mutation updateDocumentSettings($input: UpdateDocumentSettingsInput!) {
  updateDocumentSettings(input: $input)
}
`

	// DeleteDocumentMutation deletes a document.
	DeleteDocumentMutation = `
mutation deleteDocument($urn: String!) {
  deleteDocument(urn: $urn)
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

// SearchDocuments searches for documents.
func (c *Client) SearchDocuments(ctx context.Context, query string, opts ...SearchOption) (*types.DocumentSearchResult, error) {
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

	variables := map[string]any{"input": input}

	var response struct {
		SearchDocuments struct {
			Start         int `json:"start"`
			Count         int `json:"count"`
			Total         int `json:"total"`
			SearchResults []struct {
				Document documentResponse `json:"document"`
			} `json:"searchResults"`
		} `json:"searchDocuments"`
	}

	if err := c.Execute(ctx, SearchDocumentsQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("SearchDocuments(%q): %w", query, err)
	}

	result := &types.DocumentSearchResult{
		Total: response.SearchDocuments.Total,
	}

	for _, sr := range response.SearchDocuments.SearchResults {
		result.Documents = append(result.Documents, *parseDocumentResponse(&sr.Document))
	}

	return result, nil
}

// GetRelatedDocuments retrieves documents linked to an entity.
func (c *Client) GetRelatedDocuments(ctx context.Context, urn string) ([]types.Document, error) {
	variables := map[string]any{
		"urn":   urn,
		"input": map[string]any{"start": 0, "count": 100},
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

	var docs []types.Document
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

// CreateDocument creates a new context document.
func (c *Client) CreateDocument(ctx context.Context, input types.CreateDocumentInput) (string, error) {
	gqlInput := map[string]any{
		"title": input.Title,
		"contents": map[string]any{
			"text": input.Content,
		},
	}

	if input.SubType != "" {
		gqlInput["subType"] = input.SubType
	}
	if input.State != "" {
		gqlInput["state"] = input.State
	}
	if input.ShowInGlobalContext != nil {
		gqlInput["settings"] = map[string]any{
			"showInGlobalContext": *input.ShowInGlobalContext,
		}
	}
	if len(input.RelatedAssets) > 0 {
		assets := make([]map[string]any, len(input.RelatedAssets))
		for i, urn := range input.RelatedAssets {
			assets[i] = map[string]any{"urn": urn}
		}
		gqlInput["relatedAssets"] = assets
	}

	variables := map[string]any{"input": gqlInput}

	var response struct {
		CreateDocument string `json:"createDocument"`
	}

	if err := c.Execute(ctx, CreateDocumentMutation, variables, &response); err != nil {
		return "", fmt.Errorf("CreateDocument: %w", err)
	}

	return response.CreateDocument, nil
}

// UpdateDocumentContents updates a document's title and/or content.
func (c *Client) UpdateDocumentContents(ctx context.Context, urn string, title, content string) error {
	gqlInput := map[string]any{"urn": urn}
	if title != "" {
		gqlInput["title"] = title
	}
	if content != "" {
		gqlInput["contents"] = map[string]any{"text": content}
	}

	variables := map[string]any{"input": gqlInput}

	var response struct {
		UpdateDocumentContents bool `json:"updateDocumentContents"`
	}

	if err := c.Execute(ctx, UpdateDocumentContentsMutation, variables, &response); err != nil {
		return fmt.Errorf("UpdateDocumentContents(%s): %w", urn, err)
	}

	return nil
}

// UpdateDocumentRelatedEntities updates the related entities for a document.
func (c *Client) UpdateDocumentRelatedEntities(ctx context.Context, urn string, assetURNs []string) error {
	assets := make([]map[string]any, len(assetURNs))
	for i, assetURN := range assetURNs {
		assets[i] = map[string]any{"urn": assetURN}
	}

	variables := map[string]any{
		"input": map[string]any{
			"urn":           urn,
			"relatedAssets": assets,
		},
	}

	var response struct {
		UpdateDocumentRelatedEntities bool `json:"updateDocumentRelatedEntities"`
	}

	if err := c.Execute(ctx, UpdateDocumentRelatedEntitiesMutation, variables, &response); err != nil {
		return fmt.Errorf("UpdateDocumentRelatedEntities(%s): %w", urn, err)
	}

	return nil
}

// UpdateDocumentStatus updates a document's publication status.
func (c *Client) UpdateDocumentStatus(ctx context.Context, urn, state string) error {
	variables := map[string]any{
		"input": map[string]any{
			"urn":   urn,
			"state": state,
		},
	}

	var response struct {
		UpdateDocumentStatus bool `json:"updateDocumentStatus"`
	}

	if err := c.Execute(ctx, UpdateDocumentStatusMutation, variables, &response); err != nil {
		return fmt.Errorf("UpdateDocumentStatus(%s): %w", urn, err)
	}

	return nil
}

// UpdateDocumentSettings updates a document's settings.
func (c *Client) UpdateDocumentSettings(ctx context.Context, urn string, showInGlobalContext bool) error {
	variables := map[string]any{
		"input": map[string]any{
			"urn":                 urn,
			"showInGlobalContext": showInGlobalContext,
		},
	}

	var response struct {
		UpdateDocumentSettings bool `json:"updateDocumentSettings"`
	}

	if err := c.Execute(ctx, UpdateDocumentSettingsMutation, variables, &response); err != nil {
		return fmt.Errorf("UpdateDocumentSettings(%s): %w", urn, err)
	}

	return nil
}

// DeleteDocument deletes a document.
func (c *Client) DeleteDocument(ctx context.Context, urn string) error {
	variables := map[string]any{"urn": urn}

	var response struct {
		DeleteDocument bool `json:"deleteDocument"`
	}

	if err := c.Execute(ctx, DeleteDocumentMutation, variables, &response); err != nil {
		return fmt.Errorf("DeleteDocument(%s): %w", urn, err)
	}

	return nil
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
			URN string `json:"urn"`
		} `json:"relatedAssets"`
		RelatedDocuments []struct {
			URN string `json:"urn"`
		} `json:"relatedDocuments"`
		ParentDocument *struct {
			URN string `json:"urn"`
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
		doc.RelatedAssets = append(doc.RelatedAssets, types.DocumentRelatedAsset{URN: a.URN})
	}

	for _, rd := range d.Info.RelatedDocuments {
		doc.RelatedDocuments = append(doc.RelatedDocuments, types.DocumentRelatedDocument{URN: rd.URN})
	}

	if d.Info.ParentDocument != nil {
		doc.ParentDocument = &types.DocumentParent{URN: d.Info.ParentDocument.URN}
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

	// Parse tags
	for _, t := range d.Tags.Tags {
		doc.Tags = append(doc.Tags, types.Tag{
			URN:         t.Tag.URN,
			Name:        t.Tag.Name,
			Description: t.Tag.Description,
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
