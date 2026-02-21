package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// CreateQueryInput holds the parameters for creating a new Query entity.
type CreateQueryInput struct {
	// Name is an optional human-readable name for the query.
	Name string

	// Description is an optional description of the query.
	Description string

	// Statement is the SQL query text (required).
	Statement string

	// Language is the query language (default: "SQL").
	Language string

	// DatasetURNs are optional dataset URNs to associate with the query.
	DatasetURNs []string
}

// UpdateQueryInput holds the parameters for updating an existing Query entity.
type UpdateQueryInput struct {
	// URN is the query URN (required).
	URN string

	// Name is an optional updated name. Empty string means no change.
	Name string

	// Description is an optional updated description. Empty string means no change.
	Description string

	// Statement is an optional updated SQL text. Empty string means no change.
	Statement string

	// Language is an optional updated language. Empty string means no change.
	Language string

	// DatasetURNs are optional updated dataset associations. Nil means no change.
	DatasetURNs []string
}

// createQueryResponse is the GraphQL response shape for createQuery.
type createQueryResponse struct {
	CreateQuery queryEntityResponse `json:"createQuery"`
}

// updateQueryResponse is the GraphQL response shape for updateQuery.
type updateQueryResponse struct {
	UpdateQuery queryEntityResponse `json:"updateQuery"`
}

// deleteQueryResponse is the GraphQL response shape for deleteQuery.
type deleteQueryResponse struct {
	DeleteQuery bool `json:"deleteQuery"`
}

// queryEntityResponse is the nested query entity returned by mutations.
type queryEntityResponse struct {
	URN        string                 `json:"urn"`
	Properties *queryPropertiesRaw    `json:"properties"`
	Subjects   *querySubjectsRawOuter `json:"subjects"`
}

// queryPropertiesRaw maps the GraphQL QueryProperties type.
type queryPropertiesRaw struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Source      string            `json:"source"`
	Statement   queryStatementRaw `json:"statement"`
	Created     *auditStampGQL    `json:"created"`
}

// queryStatementRaw maps the QueryStatement type.
type queryStatementRaw struct {
	Value    string `json:"value"`
	Language string `json:"language"`
}

// auditStampGQL maps a GraphQL AuditStamp with time as int64.
type auditStampGQL struct {
	Time  int64  `json:"time"`
	Actor string `json:"actor"`
}

// querySubjectsRawOuter maps the QuerySubjects type.
type querySubjectsRawOuter struct {
	Datasets []queryDatasetRef `json:"datasets"`
}

// queryDatasetRef maps QuerySubjects.datasets[].dataset.
type queryDatasetRef struct {
	Dataset struct {
		URN string `json:"urn"`
	} `json:"dataset"`
}

// CreateQuery creates a new Query entity in DataHub.
func (c *Client) CreateQuery(ctx context.Context, input CreateQueryInput) (*types.Query, error) {
	if input.Statement == "" {
		return nil, fmt.Errorf("CreateQuery: statement is required")
	}

	lang := input.Language
	if lang == "" {
		lang = "SQL"
	}

	gqlInput := map[string]any{
		"properties": map[string]any{
			"name":        input.Name,
			"description": input.Description,
			"statement": map[string]any{
				"value":    input.Statement,
				"language": lang,
			},
		},
	}

	if len(input.DatasetURNs) > 0 {
		datasets := make([]map[string]string, len(input.DatasetURNs))
		for i, urn := range input.DatasetURNs {
			datasets[i] = map[string]string{"datasetUrn": urn}
		}
		gqlInput["subjects"] = map[string]any{
			"datasets": datasets,
		}
	}

	variables := map[string]any{"input": gqlInput}

	var resp createQueryResponse
	if err := c.Execute(ctx, CreateQueryMutation, variables, &resp); err != nil {
		return nil, fmt.Errorf("CreateQuery: %w", err)
	}

	return toQuery(&resp.CreateQuery), nil
}

// UpdateQuery updates an existing Query entity in DataHub.
func (c *Client) UpdateQuery(ctx context.Context, input UpdateQueryInput) (*types.Query, error) {
	if input.URN == "" {
		return nil, fmt.Errorf("UpdateQuery: urn is required")
	}

	properties := map[string]any{}
	hasProperties := false

	if input.Name != "" {
		properties["name"] = input.Name
		hasProperties = true
	}
	if input.Description != "" {
		properties["description"] = input.Description
		hasProperties = true
	}
	if input.Statement != "" {
		lang := input.Language
		if lang == "" {
			lang = "SQL"
		}
		properties["statement"] = map[string]any{
			"value":    input.Statement,
			"language": lang,
		}
		hasProperties = true
	}

	gqlInput := map[string]any{}
	if hasProperties {
		gqlInput["properties"] = properties
	}

	if input.DatasetURNs != nil {
		datasets := make([]map[string]string, len(input.DatasetURNs))
		for i, urn := range input.DatasetURNs {
			datasets[i] = map[string]string{"datasetUrn": urn}
		}
		gqlInput["subjects"] = map[string]any{
			"datasets": datasets,
		}
	}

	variables := map[string]any{
		"urn":   input.URN,
		"input": gqlInput,
	}

	var resp updateQueryResponse
	if err := c.Execute(ctx, UpdateQueryMutation, variables, &resp); err != nil {
		return nil, fmt.Errorf("UpdateQuery: %w", err)
	}

	return toQuery(&resp.UpdateQuery), nil
}

// DeleteQuery deletes a Query entity from DataHub.
func (c *Client) DeleteQuery(ctx context.Context, urn string) error {
	if urn == "" {
		return fmt.Errorf("DeleteQuery: urn is required")
	}

	variables := map[string]any{"urn": urn}

	var resp deleteQueryResponse
	if err := c.Execute(ctx, DeleteQueryMutation, variables, &resp); err != nil {
		return fmt.Errorf("DeleteQuery: %w", err)
	}

	return nil
}

// toQuery converts a GraphQL query entity response to a types.Query.
func toQuery(r *queryEntityResponse) *types.Query {
	q := &types.Query{URN: r.URN}

	if r.Properties != nil {
		q.Name = r.Properties.Name
		q.Description = r.Properties.Description
		q.Source = r.Properties.Source
		q.Statement = r.Properties.Statement.Value

		if r.Properties.Created != nil {
			q.CreatedBy = r.Properties.Created.Actor
			q.Created = r.Properties.Created.Time
		}
	}

	return q
}
