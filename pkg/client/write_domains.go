package client

import (
	"context"
	"fmt"
)

// GraphQL mutations for domain assignment.
const (
	setDomainMutation = `mutation setDomain($entityUrn: String!, $domainUrn: String!) {
		setDomain(entityUrn: $entityUrn, domainUrn: $domainUrn)
	}`

	unsetDomainMutation = `mutation unsetDomain($entityUrn: String!) {
		unsetDomain(entityUrn: $entityUrn)
	}`
)

// SetDomain assigns a domain to an entity.
func (c *Client) SetDomain(ctx context.Context, entityURN, domainURN string) error {
	variables := map[string]any{
		"entityUrn": entityURN,
		"domainUrn": domainURN,
	}
	var resp struct {
		SetDomain bool `json:"setDomain"`
	}
	if err := c.Execute(ctx, setDomainMutation, variables, &resp); err != nil {
		return fmt.Errorf("SetDomain: %w", err)
	}
	return nil
}

// UnsetDomain removes the domain from an entity.
func (c *Client) UnsetDomain(ctx context.Context, entityURN string) error {
	variables := map[string]any{
		"entityUrn": entityURN,
	}
	var resp struct {
		UnsetDomain bool `json:"unsetDomain"`
	}
	if err := c.Execute(ctx, unsetDomainMutation, variables, &resp); err != nil {
		return fmt.Errorf("UnsetDomain: %w", err)
	}
	return nil
}
