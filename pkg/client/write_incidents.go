package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL mutation for updating incidents.
const updateIncidentMutation = `mutation updateIncident($urn: String!, $input: UpdateIncidentInput!) {
	updateIncident(urn: $urn, input: $input)
}`

// UpdateIncident updates an existing incident.
func (c *Client) UpdateIncident(ctx context.Context, urn string, input types.UpdateIncidentInput) error {
	gqlInput := map[string]any{}
	if input.Title != "" {
		gqlInput["title"] = input.Title
	}
	if input.Description != "" {
		gqlInput["description"] = input.Description
	}
	if input.Type != "" {
		gqlInput["type"] = input.Type
	}
	if input.CustomType != "" {
		gqlInput["customType"] = input.CustomType
	}
	if input.Priority != "" {
		gqlInput["priority"] = input.Priority
	}

	variables := map[string]any{"urn": urn, "input": gqlInput}
	var resp struct {
		UpdateIncident bool `json:"updateIncident"`
	}
	if err := c.Execute(ctx, updateIncidentMutation, variables, &resp); err != nil {
		return fmt.Errorf("UpdateIncident: %w", err)
	}
	return nil
}
