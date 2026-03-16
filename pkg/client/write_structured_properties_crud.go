package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL mutation for updating structured property definitions.
const updateStructuredPropertyMutation = `mutation updateStructuredProperty($input: UpdateStructuredPropertyInput!) {
	updateStructuredProperty(input: $input) {
		urn
	}
}`

// UpdateStructuredProperty updates a structured property definition.
func (c *Client) UpdateStructuredProperty(ctx context.Context, urn string, input types.UpdateStructuredPropertyInput) error {
	gqlInput := map[string]any{
		"urn": urn,
	}
	if input.DisplayName != "" {
		gqlInput["displayName"] = input.DisplayName
	}
	if input.Description != "" {
		gqlInput["description"] = input.Description
	}
	if len(input.NewAllowedValues) > 0 {
		values := make([]map[string]any, len(input.NewAllowedValues))
		for i, v := range input.NewAllowedValues {
			values[i] = map[string]any{
				"value":       map[string]any{"stringValue": v.Value},
				"description": v.Description,
			}
		}
		gqlInput["newAllowedValues"] = values
	}

	variables := map[string]any{"input": gqlInput}
	var resp struct {
		UpdateStructuredProperty struct {
			URN string `json:"urn"`
		} `json:"updateStructuredProperty"`
	}
	if err := c.Execute(ctx, updateStructuredPropertyMutation, variables, &resp); err != nil {
		return fmt.Errorf("UpdateStructuredProperty: %w", err)
	}
	return nil
}
