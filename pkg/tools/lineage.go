package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// GetLineageInput is the input for the get_lineage tool.
type GetLineageInput struct {
	URN       string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	Direction string `json:"direction,omitempty" jsonschema_description:"Lineage direction: UPSTREAM or DOWNSTREAM (default: DOWNSTREAM)"`
	Depth     int    `json:"depth,omitempty" jsonschema_description:"Maximum depth of lineage traversal (default: 1, max: 5)"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerGetLineageTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		lineageInput, ok := input.(GetLineageInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetLineage(ctx, req, lineageInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetLineage, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolGetLineage),
		Description:  t.getDescription(ToolGetLineage, cfg),
		Annotations:  t.getAnnotations(ToolGetLineage, cfg),
		Icons:        t.getIcons(ToolGetLineage, cfg),
		Title:        t.getTitle(ToolGetLineage, cfg),
		OutputSchema: t.getOutputSchema(ToolGetLineage, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetLineageInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetLineage(ctx context.Context, _ *mcp.CallToolRequest, input GetLineageInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	var opts []client.LineageOption
	if input.Direction != "" {
		opts = append(opts, client.WithDirection(input.Direction))
	}
	if input.Depth > 0 {
		opts = append(opts, client.WithDepth(input.Depth))
	}

	lineage, err := datahubClient.GetLineage(ctx, input.URN, opts...)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	if t.queryProvider != nil {
		return t.enrichLineageWithQueryContext(ctx, lineage)
	}

	return formatJSONResult(lineage)
}

// enrichLineageWithQueryContext flattens lineage fields to top level and appends
// query execution context at the same level (matches OutputSchema).
// LineageEdge.Properties is map[string]any, so json.Marshal can fail for pathological
// values (e.g. channels). Unmarshal of the resulting JSON into map[string]any
// is always safe and its error is intentionally ignored (check-blank: false).
func (t *Toolkit) enrichLineageWithQueryContext(ctx context.Context, lineage *types.LineageResult) (*mcp.CallToolResult, any, error) {
	lineageJSON, err := json.Marshal(lineage)
	if err != nil {
		return ErrorResult("failed to marshal lineage: " + err.Error()), nil, nil
	}
	response := map[string]any{}
	_ = json.Unmarshal(lineageJSON, &response)

	urns := collectLineageURNs(lineage)
	if len(urns) > 0 {
		if execCtx, execErr := t.queryProvider.GetExecutionContext(ctx, urns); execErr == nil && execCtx != nil {
			response["execution_context"] = execCtx
		}
	}

	return formatJSONResult(response)
}

// collectLineageURNs extracts all URNs from a lineage result.
func collectLineageURNs(lineage any) []string {
	var urns []string

	// Use type switch to handle different lineage result structures
	switch v := lineage.(type) {
	case map[string]any:
		// Check for start URN
		if start, ok := v["start"].(string); ok {
			urns = append(urns, start)
		}
		// Check for nodes array
		if nodes, ok := v["nodes"].([]any); ok {
			for _, n := range nodes {
				if node, ok := n.(map[string]any); ok {
					if urn, ok := node["urn"].(string); ok {
						urns = append(urns, urn)
					}
				}
			}
		}
	default:
		// For typed structs, we need reflection or specific type handling
		// For now, handle via JSON roundtrip
		data, err := json.Marshal(lineage)
		if err != nil {
			return urns
		}
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return urns
		}
		return collectLineageURNs(m)
	}

	return urns
}
