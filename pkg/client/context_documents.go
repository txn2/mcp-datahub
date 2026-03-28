package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// contextDocFields is the shared field selection for context document queries.
// Extracted as a const to avoid duplicating the same fields across inline
// fragments for each entity type.
const contextDocFields = `{
          urn
          subType
          info {
            title
            contents { text }
            created { time }
            lastModified { time }
          }
          ownership {
            owners {
              owner {
                ... on CorpUser { urn username }
              }
            }
          }
        }`

// GetContextDocumentsQuery retrieves context documents linked to an entity.
// Uses inline fragments because the relatedDocuments field must be declared
// per concrete type — there is no shared GraphQL interface.
//
// Covers Dataset, GlossaryTerm, GlossaryNode, and Container. Other entity
// types (Dashboard, Chart, DataFlow, DataJob, DataProduct, etc.) are not
// included — if they gain relatedDocuments support, add fragments here.
const GetContextDocumentsQuery = `
query getContextDocuments($urn: String!, $input: RelatedDocumentsInput!) {
  entity(urn: $urn) {
    ... on Dataset {
      relatedDocuments(input: $input) {
        documents ` + contextDocFields + `
      }
    }
    ... on GlossaryTerm {
      relatedDocuments(input: $input) {
        documents ` + contextDocFields + `
      }
    }
    ... on GlossaryNode {
      relatedDocuments(input: $input) {
        documents ` + contextDocFields + `
      }
    }
    ... on Container {
      relatedDocuments(input: $input) {
        documents ` + contextDocFields + `
      }
    }
  }
}
`

// contextDocResponse matches the document response shape for context document queries.
type contextDocResponse struct {
	URN     string `json:"urn"`
	SubType string `json:"subType"`
	Info    struct {
		Title    string `json:"title"`
		Contents struct {
			Text string `json:"text"`
		} `json:"contents"`
		Created struct {
			Time int64 `json:"time"`
		} `json:"created"`
		LastModified struct {
			Time int64 `json:"time"`
		} `json:"lastModified"`
	} `json:"info"`
	Ownership struct {
		Owners []struct {
			Owner struct {
				URN      string `json:"urn"`
				Username string `json:"username"`
			} `json:"owner"`
		} `json:"owners"`
	} `json:"ownership"`
}

// GetContextDocuments retrieves context documents linked to an entity.
// Results are capped at MaxLimit; no pagination is performed. If an entity
// has more context documents than MaxLimit, excess documents are truncated.
func (c *Client) GetContextDocuments(ctx context.Context, urn string) ([]types.ContextDocument, error) {
	variables := map[string]any{
		"urn":   urn,
		"input": map[string]any{"start": 0, "count": c.config.MaxLimit},
	}

	var response struct {
		Entity struct {
			RelatedDocuments *struct {
				Documents []contextDocResponse `json:"documents"`
			} `json:"relatedDocuments"`
		} `json:"entity"`
	}

	if err := c.Execute(ctx, GetContextDocumentsQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("GetContextDocuments(%s): %w", urn, err)
	}

	if response.Entity.RelatedDocuments == nil {
		return nil, nil
	}

	docs := make([]types.ContextDocument, 0, len(response.Entity.RelatedDocuments.Documents))
	for i := range response.Entity.RelatedDocuments.Documents {
		docs = append(docs, toContextDocument(&response.Entity.RelatedDocuments.Documents[i]))
	}

	return docs, nil
}

// UpsertContextDocument creates or updates a context document on an entity.
// If doc.ID is empty, creates a new document linked to entityURN.
// If doc.ID is set, updates the existing document.
func (c *Client) UpsertContextDocument(
	ctx context.Context, entityURN string, doc types.ContextDocumentInput,
) (*types.ContextDocument, error) {
	if doc.Title == "" {
		return nil, fmt.Errorf("UpsertContextDocument: title is required")
	}

	if doc.ID == "" {
		return c.createContextDocument(ctx, entityURN, doc)
	}
	return c.updateContextDocument(ctx, doc)
}

// createContextDocument creates a new document linked to an entity.
func (c *Client) createContextDocument(
	ctx context.Context, entityURN string, doc types.ContextDocumentInput,
) (*types.ContextDocument, error) {
	input := types.CreateDocumentInput{
		Title:            doc.Title,
		Content:          doc.Content,
		SubType:          doc.Category,
		RelatedAssetURNs: []string{entityURN},
		Status:           "PUBLISHED",
	}

	urn, err := c.CreateDocument(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("UpsertContextDocument(create): %w", err)
	}

	return c.fetchContextDocument(ctx, urn)
}

// updateContextDocument updates an existing document's contents and optional subType.
func (c *Client) updateContextDocument(
	ctx context.Context, doc types.ContextDocumentInput,
) (*types.ContextDocument, error) {
	urn := BuildDocumentURN(doc.ID)

	if err := c.UpdateDocumentContents(ctx, urn, doc.Title, doc.Content); err != nil {
		return nil, fmt.Errorf("UpsertContextDocument(update contents): %w", err)
	}

	if doc.Category != "" {
		if err := c.UpdateDocumentSubType(ctx, urn, doc.Category); err != nil {
			return nil, fmt.Errorf("UpsertContextDocument(update category): %w", err)
		}
	}

	return c.fetchContextDocument(ctx, urn)
}

// fetchContextDocument retrieves a document by URN and converts it.
func (c *Client) fetchContextDocument(ctx context.Context, urn string) (*types.ContextDocument, error) {
	full, err := c.GetDocument(ctx, urn)
	if err != nil {
		return nil, fmt.Errorf("fetchContextDocument(%s): %w", urn, err)
	}
	return documentToContextDocument(full), nil
}

// DeleteContextDocument removes a context document by its ID.
func (c *Client) DeleteContextDocument(ctx context.Context, documentID string) error {
	urn := BuildDocumentURN(documentID)
	if err := c.DeleteDocument(ctx, urn); err != nil {
		return fmt.Errorf("DeleteContextDocument(%s): %w", documentID, err)
	}
	return nil
}

// toContextDocument converts a contextDocResponse to a ContextDocument.
func toContextDocument(d *contextDocResponse) types.ContextDocument {
	doc := types.ContextDocument{
		ID:          documentIDFromURN(d.URN),
		Title:       d.Info.Title,
		Content:     d.Info.Contents.Text,
		ContentType: "text/markdown",
		Category:    d.SubType,
		CreatedAt:   d.Info.Created.Time,
		UpdatedAt:   d.Info.LastModified.Time,
	}

	if len(d.Ownership.Owners) > 0 {
		owner := d.Ownership.Owners[0].Owner
		doc.Author = &types.ContextDocumentAuthor{
			URN:      owner.URN,
			Username: owner.Username,
		}
	}

	return doc
}

// documentToContextDocument converts a full Document to a ContextDocument.
func documentToContextDocument(d *types.Document) *types.ContextDocument {
	doc := &types.ContextDocument{
		ID:          documentIDFromURN(d.URN),
		Title:       d.Title,
		Content:     d.Content,
		ContentType: "text/markdown",
		Category:    d.SubType,
		CreatedAt:   d.Created,
		UpdatedAt:   d.LastModified,
	}

	if len(d.Owners) > 0 {
		doc.Author = &types.ContextDocumentAuthor{
			URN:      d.Owners[0].URN,
			Username: usernameFromOwnerURN(d.Owners[0].URN),
		}
	}

	return doc
}

// usernameFromOwnerURN extracts the username from a corpuser URN.
// Owner.Name is unsuitable because it may contain a display name
// (e.g., "Alice Smith") rather than the login (e.g., "alice").
func usernameFromOwnerURN(urn string) string {
	return strings.TrimPrefix(urn, "urn:li:corpuser:")
}

// documentIDFromURN extracts the document ID from a document URN.
func documentIDFromURN(urn string) string {
	return strings.TrimPrefix(urn, "urn:li:document:")
}
