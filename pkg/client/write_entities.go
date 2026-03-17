package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL mutations for entity creation.
const (
	createTagMutation = `mutation createTag($input: CreateTagInput!) {
		createTag(input: $input)
	}`

	createDomainMutation = `mutation createDomain($input: CreateDomainInput!) {
		createDomain(input: $input)
	}`

	createGlossaryTermMutation = `mutation createGlossaryTerm($input: CreateGlossaryEntityInput!) {
		createGlossaryTerm(input: $input)
	}`

	createDataProductMutation = `mutation createDataProduct($input: CreateDataProductInput!) {
		createDataProduct(input: $input)
	}`

	createDocumentMutation = `mutation createDocument($input: CreateDocumentInput!) {
		createDocument(input: $input)
	}`

	createApplicationMutation = `mutation createApplication($input: CreateApplicationInput!) {
		createApplication(input: $input)
	}`

	createStructuredPropertyMutation = `mutation createStructuredProperty($input: CreateStructuredPropertyInput!) {
		createStructuredProperty(input: $input) {
			urn
		}
	}`

	upsertDataContractMutation = `mutation upsertDataContract($input: UpsertDataContractInput!) {
		upsertDataContract(input: $input) {
			urn
		}
	}`
)

// CreateTag creates a new tag entity in DataHub.
func (c *Client) CreateTag(ctx context.Context, name, description string) (string, error) {
	input := map[string]any{
		"name":        name,
		"description": description,
	}
	var resp struct {
		CreateTag string `json:"createTag"`
	}
	if err := c.Execute(ctx, createTagMutation, map[string]any{"input": input}, &resp); err != nil {
		return "", fmt.Errorf("CreateTag: %w", err)
	}
	return resp.CreateTag, nil
}

// CreateDomain creates a new domain entity in DataHub.
func (c *Client) CreateDomain(ctx context.Context, name, description string) (string, error) {
	input := map[string]any{
		"name":        name,
		"description": description,
	}
	var resp struct {
		CreateDomain string `json:"createDomain"`
	}
	if err := c.Execute(ctx, createDomainMutation, map[string]any{"input": input}, &resp); err != nil {
		return "", fmt.Errorf("CreateDomain: %w", err)
	}
	return resp.CreateDomain, nil
}

// CreateGlossaryTerm creates a new glossary term entity in DataHub.
func (c *Client) CreateGlossaryTerm(ctx context.Context, name, description, parentNode string) (string, error) {
	input := map[string]any{
		"name":        name,
		"description": description,
	}
	if parentNode != "" {
		input["parentNode"] = parentNode
	}
	var resp struct {
		CreateGlossaryTerm string `json:"createGlossaryTerm"`
	}
	if err := c.Execute(ctx, createGlossaryTermMutation, map[string]any{"input": input}, &resp); err != nil {
		return "", fmt.Errorf("CreateGlossaryTerm: %w", err)
	}
	return resp.CreateGlossaryTerm, nil
}

// CreateDataProduct creates a new data product entity in DataHub.
func (c *Client) CreateDataProduct(ctx context.Context, name, description, domainURN string) (string, error) {
	if domainURN == "" {
		return "", fmt.Errorf("CreateDataProduct: domainURN is required")
	}
	input := map[string]any{
		"name":        name,
		"description": description,
		"domainUrn":   domainURN,
	}
	var resp struct {
		CreateDataProduct string `json:"createDataProduct"`
	}
	if err := c.Execute(ctx, createDataProductMutation, map[string]any{"input": input}, &resp); err != nil {
		return "", fmt.Errorf("CreateDataProduct: %w", err)
	}
	return resp.CreateDataProduct, nil
}

// CreateDocument creates a new context document entity in DataHub.
func (c *Client) CreateDocument(ctx context.Context, input types.CreateDocumentInput) (string, error) {
	if input.Title == "" {
		return "", fmt.Errorf("CreateDocument: title is required")
	}
	gqlInput := map[string]any{
		"title":   input.Title,
		"content": input.Content,
	}
	if input.Status != "" {
		gqlInput["status"] = input.Status
	}
	if input.SubType != "" {
		gqlInput["subType"] = input.SubType
	}
	if len(input.RelatedAssetURNs) > 0 {
		gqlInput["relatedAssets"] = input.RelatedAssetURNs
	}
	if input.GlobalContext {
		gqlInput["settings"] = map[string]any{"showInGlobalContext": true}
	}

	var resp struct {
		CreateDocument string `json:"createDocument"`
	}
	if err := c.Execute(ctx, createDocumentMutation, map[string]any{"input": gqlInput}, &resp); err != nil {
		return "", fmt.Errorf("CreateDocument: %w", err)
	}
	return resp.CreateDocument, nil
}

// CreateApplication creates a new application entity in DataHub.
func (c *Client) CreateApplication(ctx context.Context, name, description string) (string, error) {
	input := map[string]any{
		"name":        name,
		"description": description,
	}
	var resp struct {
		CreateApplication string `json:"createApplication"`
	}
	if err := c.Execute(ctx, createApplicationMutation, map[string]any{"input": input}, &resp); err != nil {
		return "", fmt.Errorf("CreateApplication: %w", err)
	}
	return resp.CreateApplication, nil
}

// CreateStructuredProperty creates a new structured property definition in DataHub.
func (c *Client) CreateStructuredProperty(ctx context.Context, input types.CreateStructuredPropertyInput) (string, error) {
	if input.QualifiedName == "" {
		return "", fmt.Errorf("CreateStructuredProperty: qualifiedName is required")
	}
	if input.ValueType == "" {
		return "", fmt.Errorf("CreateStructuredProperty: valueType is required")
	}

	gqlInput := map[string]any{
		"qualifiedName": input.QualifiedName,
		"valueType":     input.ValueType,
	}
	if input.DisplayName != "" {
		gqlInput["displayName"] = input.DisplayName
	}
	if input.Description != "" {
		gqlInput["description"] = input.Description
	}
	if input.Cardinality != "" {
		gqlInput["cardinality"] = input.Cardinality
	}
	if len(input.EntityTypes) > 0 {
		gqlInput["entityTypes"] = input.EntityTypes
	}
	if len(input.AllowedValues) > 0 {
		av := make([]map[string]any, len(input.AllowedValues))
		for i, v := range input.AllowedValues {
			av[i] = map[string]any{
				"value":       map[string]any{"stringValue": v.Value},
				"description": v.Description,
			}
		}
		gqlInput["allowedValues"] = av
	}

	var resp struct {
		CreateStructuredProperty struct {
			URN string `json:"urn"`
		} `json:"createStructuredProperty"`
	}
	if err := c.Execute(ctx, createStructuredPropertyMutation, map[string]any{"input": gqlInput}, &resp); err != nil {
		return "", fmt.Errorf("CreateStructuredProperty: %w", err)
	}
	return resp.CreateStructuredProperty.URN, nil
}

// UpsertDataContract creates or updates a data contract for a dataset.
func (c *Client) UpsertDataContract(ctx context.Context, input types.UpsertDataContractInput) (string, error) {
	if input.DatasetURN == "" {
		return "", fmt.Errorf("UpsertDataContract: datasetURN is required")
	}

	gqlInput := map[string]any{
		"entityUrn": input.DatasetURN,
	}

	if len(input.SchemaAssertionURNs) > 0 {
		schema := make([]map[string]string, len(input.SchemaAssertionURNs))
		for i, urn := range input.SchemaAssertionURNs {
			schema[i] = map[string]string{"assertionUrn": urn}
		}
		gqlInput["schema"] = schema
	}
	if len(input.FreshnessAssertionURNs) > 0 {
		freshness := make([]map[string]string, len(input.FreshnessAssertionURNs))
		for i, urn := range input.FreshnessAssertionURNs {
			freshness[i] = map[string]string{"assertionUrn": urn}
		}
		gqlInput["freshness"] = freshness
	}
	if len(input.DataQualityAssertionURNs) > 0 {
		dq := make([]map[string]string, len(input.DataQualityAssertionURNs))
		for i, urn := range input.DataQualityAssertionURNs {
			dq[i] = map[string]string{"assertionUrn": urn}
		}
		gqlInput["dataQuality"] = dq
	}

	var resp struct {
		UpsertDataContract struct {
			URN string `json:"urn"`
		} `json:"upsertDataContract"`
	}
	if err := c.Execute(ctx, upsertDataContractMutation, map[string]any{"input": gqlInput}, &resp); err != nil {
		return "", fmt.Errorf("UpsertDataContract: %w", err)
	}
	return resp.UpsertDataContract.URN, nil
}
