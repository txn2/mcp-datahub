package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// SearchDocumentsQuery searches context documents via searchAcrossEntities scoped
// to the DOCUMENT entity type. Unlike the generic SearchAcrossEntitiesQuery, it
// returns the full document selection set (the same fields as GetDocumentQuery)
// so callers can discover and filter documents without a follow-up GetDocument
// lookup. Available in DataHub 1.4.x+ (documents are a 1.4.x feature).
const SearchDocumentsQuery = `
query searchDocuments($input: SearchAcrossEntitiesInput!) {
  searchAcrossEntities(input: $input) {
    start
    count
    total
    searchResults {
      entity {
        urn
        type
        ... on Document {` + documentSelectionFields + `        }
      }
    }
  }
}
`

// SearchDocuments searches context documents by relevance and returns full
// document metadata. A query of "*" lists all documents; a text query ranks by
// relevance. Use SearchOption values (WithLimit, WithOffset, WithSearchFilters)
// to page or filter results.
//
// Unlike the point-lookup accessors (GetDocument, GetRelatedDocuments), this
// method discovers documents without a known URN. Each result carries the same
// metadata as GetDocument (URN, title, sub-type, related-asset URNs,
// showInGlobalContext, ownership, tags, glossary terms, and domain) so callers
// can filter to globally-visible documents.
//
// Available in DataHub 1.4.x+.
func (c *Client) SearchDocuments(ctx context.Context, query string, opts ...SearchOption) ([]types.Document, error) {
	input, _ := c.buildBaseSearchInput(query, opts)
	input["types"] = []string{EntityTypeDocument}

	variables := map[string]any{
		"input": input,
	}

	var response struct {
		SearchAcrossEntities struct {
			SearchResults []struct {
				Entity documentResponse `json:"entity"`
			} `json:"searchResults"`
		} `json:"searchAcrossEntities"`
	}

	if err := c.Execute(ctx, SearchDocumentsQuery, variables, &response); err != nil {
		return nil, fmt.Errorf("SearchDocuments(%q): %w", query, err)
	}

	docs := make([]types.Document, 0, len(response.SearchAcrossEntities.SearchResults))
	for i := range response.SearchAcrossEntities.SearchResults {
		docs = append(docs, *parseDocumentResponse(&response.SearchAcrossEntities.SearchResults[i].Entity))
	}

	return docs, nil
}
