package client

import (
	"context"
	"fmt"
)

// GraphQL mutations for entity types where REST aspect writes are not supported.
// Domain, glossaryTerm, and glossaryNode do not register globalTags, glossaryTerms,
// or editable description aspects in the REST API. DataHub's GraphQL mutations
// handle the entity-type routing internally.

const (
	// AddTagMutation adds a tag to any entity via GraphQL.
	// Signature: addTag(input: TagAssociationInput!): Boolean.
	AddTagMutation = `
mutation addTag($input: TagAssociationInput!) {
  addTag(input: $input)
}
`

	// RemoveTagMutation removes a tag from any entity via GraphQL.
	// Signature: removeTag(input: TagAssociationInput!): Boolean.
	RemoveTagMutation = `
mutation removeTag($input: TagAssociationInput!) {
  removeTag(input: $input)
}
`

	// AddTermMutation adds a glossary term to any entity via GraphQL.
	// Signature: addTerm(input: TermAssociationInput!): Boolean.
	AddTermMutation = `
mutation addTerm($input: TermAssociationInput!) {
  addTerm(input: $input)
}
`

	// RemoveTermMutation removes a glossary term from any entity via GraphQL.
	// Signature: removeTerm(input: TermAssociationInput!): Boolean.
	RemoveTermMutation = `
mutation removeTerm($input: TermAssociationInput!) {
  removeTerm(input: $input)
}
`

	// UpdateDescriptionMutation updates the description of any entity via GraphQL.
	// Signature: updateDescription(input: DescriptionUpdateInput!): Boolean.
	UpdateDescriptionMutation = `
mutation updateDescription($input: DescriptionUpdateInput!) {
  updateDescription(input: $input)
}
`
)

// graphQLWriteTypes lists entity types that require GraphQL mutations for write
// operations because the REST API does not support their aspects (globalTags,
// glossaryTerms, editable description).
var graphQLWriteTypes = map[string]bool{
	entityTypeDomain:       true,
	entityTypeGlossaryTerm: true,
	entityTypeGlossaryNode: true,
}

// addTagGraphQL adds a tag to an entity using the GraphQL addTag mutation.
func (c *Client) addTagGraphQL(ctx context.Context, urn, tagURN string) error {
	variables := map[string]any{
		"input": map[string]any{
			"tagUrn":      tagURN,
			"resourceUrn": urn,
		},
	}

	var response struct {
		AddTag bool `json:"addTag"`
	}

	if err := c.Execute(ctx, AddTagMutation, variables, &response); err != nil {
		return fmt.Errorf("addTagGraphQL: %w", err)
	}

	return nil
}

// removeTagGraphQL removes a tag from an entity using the GraphQL removeTag mutation.
func (c *Client) removeTagGraphQL(ctx context.Context, urn, tagURN string) error {
	variables := map[string]any{
		"input": map[string]any{
			"tagUrn":      tagURN,
			"resourceUrn": urn,
		},
	}

	var response struct {
		RemoveTag bool `json:"removeTag"`
	}

	if err := c.Execute(ctx, RemoveTagMutation, variables, &response); err != nil {
		return fmt.Errorf("removeTagGraphQL: %w", err)
	}

	return nil
}

// addTermGraphQL adds a glossary term to an entity using the GraphQL addTerm mutation.
func (c *Client) addTermGraphQL(ctx context.Context, urn, termURN string) error {
	variables := map[string]any{
		"input": map[string]any{
			"termUrn":     termURN,
			"resourceUrn": urn,
		},
	}

	var response struct {
		AddTerm bool `json:"addTerm"`
	}

	if err := c.Execute(ctx, AddTermMutation, variables, &response); err != nil {
		return fmt.Errorf("addTermGraphQL: %w", err)
	}

	return nil
}

// removeTermGraphQL removes a glossary term from an entity using the GraphQL removeTerm mutation.
func (c *Client) removeTermGraphQL(ctx context.Context, urn, termURN string) error {
	variables := map[string]any{
		"input": map[string]any{
			"termUrn":     termURN,
			"resourceUrn": urn,
		},
	}

	var response struct {
		RemoveTerm bool `json:"removeTerm"`
	}

	if err := c.Execute(ctx, RemoveTermMutation, variables, &response); err != nil {
		return fmt.Errorf("removeTermGraphQL: %w", err)
	}

	return nil
}

// updateDescriptionGraphQL updates an entity's description using the GraphQL mutation.
func (c *Client) updateDescriptionGraphQL(ctx context.Context, urn, description string) error {
	variables := map[string]any{
		"input": map[string]any{
			"description": description,
			"resourceUrn": urn,
		},
	}

	var response struct {
		UpdateDescription bool `json:"updateDescription"`
	}

	if err := c.Execute(ctx, UpdateDescriptionMutation, variables, &response); err != nil {
		return fmt.Errorf("updateDescriptionGraphQL: %w", err)
	}

	return nil
}
