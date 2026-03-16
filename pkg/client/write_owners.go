package client

import (
	"context"
	"fmt"
)

// GraphQL mutations for ownership management.
const (
	addOwnerMutation = `mutation addOwner($input: AddOwnerInput!) {
		addOwner(input: $input)
	}`

	removeOwnerMutation = `mutation removeOwner($input: RemoveOwnerInput!) {
		removeOwner(input: $input)
	}`
)

// AddOwner adds an owner to an entity.
func (c *Client) AddOwner(ctx context.Context, urn, ownerURN, ownershipType string) error {
	if ownershipType == "" {
		ownershipType = "TECHNICAL_OWNER"
	}
	input := map[string]any{
		"ownerUrn":         ownerURN,
		"ownerEntityType":  "CORP_USER",
		"ownershipTypeUrn": "urn:li:ownershipType:" + ownershipType,
		"resourceUrn":      urn,
	}
	var resp struct {
		AddOwner bool `json:"addOwner"`
	}
	if err := c.Execute(ctx, addOwnerMutation, map[string]any{"input": input}, &resp); err != nil {
		return fmt.Errorf("AddOwner: %w", err)
	}
	return nil
}

// RemoveOwner removes an owner from an entity.
func (c *Client) RemoveOwner(ctx context.Context, urn, ownerURN string) error {
	input := map[string]any{
		"ownerUrn":    ownerURN,
		"resourceUrn": urn,
	}
	var resp struct {
		RemoveOwner bool `json:"removeOwner"`
	}
	if err := c.Execute(ctx, removeOwnerMutation, map[string]any{"input": input}, &resp); err != nil {
		return fmt.Errorf("RemoveOwner: %w", err)
	}
	return nil
}
