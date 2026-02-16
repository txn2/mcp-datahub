package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
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
		Name:        string(ToolGetLineage),
		Description: t.getDescription(ToolGetLineage, cfg),
		Annotations: t.getAnnotations(ToolGetLineage, cfg),
		Icons:       t.getIcons(ToolGetLineage, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetLineageInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetLineage(ctx context.Context, _ *mcp.CallToolRequest, input GetLineageInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
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

	// Build response - include execution context if provider configured
	if t.queryProvider != nil {
		response := map[string]any{
			"lineage": lineage,
		}

		// Collect all URNs from lineage
		urns := collectLineageURNs(lineage)

		// Get execution context for lineage bridge
		if len(urns) > 0 {
			if execCtx, execErr := t.queryProvider.GetExecutionContext(ctx, urns); execErr == nil && execCtx != nil {
				response["execution_context"] = execCtx
			}
		}

		jsonResult, jsonErr := JSONResult(response)
		if jsonErr != nil {
			return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
		}
		return jsonResult, nil, nil
	}

	// No query provider - return lineage only
	jsonResult, err := JSONResult(lineage)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
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
