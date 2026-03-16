package client

import (
	"context"
	"fmt"
)

// GraphQL mutations for entity deletion.
const (
	deleteTagMutation = `mutation deleteTag($urn: String!) {
		deleteTag(urn: $urn)
	}`

	deleteDomainMutation = `mutation deleteDomain($urn: String!) {
		deleteDomain(urn: $urn)
	}`

	deleteGlossaryEntityMutation = `mutation deleteGlossaryEntity($urn: String!) {
		deleteGlossaryEntity(urn: $urn)
	}`

	deleteDataProductMutation = `mutation deleteDataProduct($urn: String!) {
		deleteDataProduct(urn: $urn)
	}`

	deleteApplicationMutation = `mutation deleteApplication($urn: String!) {
		deleteApplication(urn: $urn)
	}`

	deleteStructuredPropertyMutation = `mutation deleteStructuredProperty($input: DeleteStructuredPropertyInput!) {
		deleteStructuredProperty(input: $input)
	}`
)

// DeleteTag deletes a tag entity.
func (c *Client) DeleteTag(ctx context.Context, urn string) error {
	var resp struct {
		DeleteTag bool `json:"deleteTag"`
	}
	if err := c.Execute(ctx, deleteTagMutation, map[string]any{"urn": urn}, &resp); err != nil {
		return fmt.Errorf("DeleteTag: %w", err)
	}
	return nil
}

// DeleteDomain deletes a domain entity.
func (c *Client) DeleteDomain(ctx context.Context, urn string) error {
	var resp struct {
		DeleteDomain bool `json:"deleteDomain"`
	}
	if err := c.Execute(ctx, deleteDomainMutation, map[string]any{"urn": urn}, &resp); err != nil {
		return fmt.Errorf("DeleteDomain: %w", err)
	}
	return nil
}

// DeleteGlossaryEntity deletes a glossary term or node.
func (c *Client) DeleteGlossaryEntity(ctx context.Context, urn string) error {
	var resp struct {
		DeleteGlossaryEntity bool `json:"deleteGlossaryEntity"`
	}
	if err := c.Execute(ctx, deleteGlossaryEntityMutation, map[string]any{"urn": urn}, &resp); err != nil {
		return fmt.Errorf("DeleteGlossaryEntity: %w", err)
	}
	return nil
}

// DeleteDataProduct deletes a data product entity.
func (c *Client) DeleteDataProduct(ctx context.Context, urn string) error {
	var resp struct {
		DeleteDataProduct bool `json:"deleteDataProduct"`
	}
	if err := c.Execute(ctx, deleteDataProductMutation, map[string]any{"urn": urn}, &resp); err != nil {
		return fmt.Errorf("DeleteDataProduct: %w", err)
	}
	return nil
}

// DeleteApplication deletes an application entity.
func (c *Client) DeleteApplication(ctx context.Context, urn string) error {
	var resp struct {
		DeleteApplication bool `json:"deleteApplication"`
	}
	if err := c.Execute(ctx, deleteApplicationMutation, map[string]any{"urn": urn}, &resp); err != nil {
		return fmt.Errorf("DeleteApplication: %w", err)
	}
	return nil
}

// DeleteStructuredProperty deletes a structured property definition.
func (c *Client) DeleteStructuredProperty(ctx context.Context, urn string) error {
	input := map[string]any{"urn": urn}
	var resp struct {
		DeleteStructuredProperty bool `json:"deleteStructuredProperty"`
	}
	if err := c.Execute(ctx, deleteStructuredPropertyMutation, map[string]any{"input": input}, &resp); err != nil {
		return fmt.Errorf("DeleteStructuredProperty: %w", err)
	}
	return nil
}
