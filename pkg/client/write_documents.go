package client

import (
	"context"
	"fmt"
)

// GraphQL mutations for document management.
const (
	updateDocumentContentsMutation = `mutation updateDocumentContents($input: UpdateDocumentContentsInput!) {
		updateDocumentContents(input: $input)
	}`

	updateDocumentStatusMutation = `mutation updateDocumentStatus($input: UpdateDocumentStatusInput!) {
		updateDocumentStatus(input: $input)
	}`

	updateDocumentRelatedEntitiesMutation = `mutation updateDocumentRelatedEntities($input: UpdateDocumentRelatedEntitiesInput!) {
		updateDocumentRelatedEntities(input: $input)
	}`

	updateDocumentSubTypeMutation = `mutation updateDocumentSubType($input: UpdateDocumentSubTypeInput!) {
		updateDocumentSubType(input: $input)
	}`

	deleteDocumentMutation = `mutation deleteDocument($urn: String!) {
		deleteDocument(urn: $urn)
	}`
)

// UpdateDocumentContents updates the title and text of a document.
func (c *Client) UpdateDocumentContents(ctx context.Context, urn, title, text string) error {
	input := map[string]any{"urn": urn}
	if title != "" {
		input["title"] = title
	}
	if text != "" {
		input["contents"] = map[string]any{"text": text}
	}
	variables := map[string]any{"input": input}
	var resp struct {
		UpdateDocumentContents bool `json:"updateDocumentContents"`
	}
	if err := c.Execute(ctx, updateDocumentContentsMutation, variables, &resp); err != nil {
		return fmt.Errorf("UpdateDocumentContents: %w", err)
	}
	return nil
}

// UpdateDocumentStatus updates the publication status of a document.
func (c *Client) UpdateDocumentStatus(ctx context.Context, urn, status string) error {
	variables := map[string]any{
		"input": map[string]any{"urn": urn, "state": status},
	}
	var resp struct {
		UpdateDocumentStatus bool `json:"updateDocumentStatus"`
	}
	if err := c.Execute(ctx, updateDocumentStatusMutation, variables, &resp); err != nil {
		return fmt.Errorf("UpdateDocumentStatus: %w", err)
	}
	return nil
}

// UpdateDocumentRelatedEntities updates entities related to a document.
func (c *Client) UpdateDocumentRelatedEntities(ctx context.Context, urn string, entityURNs []string) error {
	variables := map[string]any{
		"input": map[string]any{"urn": urn, "relatedAssets": entityURNs},
	}
	var resp struct {
		UpdateDocumentRelatedEntities bool `json:"updateDocumentRelatedEntities"`
	}
	if err := c.Execute(ctx, updateDocumentRelatedEntitiesMutation, variables, &resp); err != nil {
		return fmt.Errorf("UpdateDocumentRelatedEntities: %w", err)
	}
	return nil
}

// UpdateDocumentSubType updates the sub-type of a document.
func (c *Client) UpdateDocumentSubType(ctx context.Context, urn, subType string) error {
	variables := map[string]any{
		"input": map[string]any{"urn": urn, "subType": subType},
	}
	var resp struct {
		UpdateDocumentSubType bool `json:"updateDocumentSubType"`
	}
	if err := c.Execute(ctx, updateDocumentSubTypeMutation, variables, &resp); err != nil {
		return fmt.Errorf("UpdateDocumentSubType: %w", err)
	}
	return nil
}

// DeleteDocument deletes a document entity.
func (c *Client) DeleteDocument(ctx context.Context, urn string) error {
	variables := map[string]any{"urn": urn}
	var resp struct {
		DeleteDocument bool `json:"deleteDocument"`
	}
	if err := c.Execute(ctx, deleteDocumentMutation, variables, &resp); err != nil {
		return fmt.Errorf("DeleteDocument: %w", err)
	}
	return nil
}
