package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListConnectionsInput defines the input for the datahub_list_connections tool.
// This tool has no required parameters.
type ListConnectionsInput struct{}

// ListConnectionsOutput defines the output of the datahub_list_connections tool.
type ListConnectionsOutput struct {
	Connections []ConnectionInfoOutput `json:"connections"`
	Count       int                    `json:"count"`
}

// ConnectionInfoOutput provides information about a single connection.
type ConnectionInfoOutput struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	IsDefault bool   `json:"is_default"`
}

// registerListConnectionsTool adds the datahub_list_connections tool to the server.
func (t *Toolkit) registerListConnectionsTool(server *mcp.Server, cfg *toolConfig) {
	// Create the base handler
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		return t.handleListConnections(ctx, req)
	}

	// Wrap with middleware if configured
	wrappedHandler := t.wrapHandler(ToolListConnections, baseHandler, cfg)

	// Register with MCP
	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolListConnections),
		Description: t.getDescription(ToolListConnections, cfg),
		Annotations: t.getAnnotations(ToolListConnections, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListConnectionsInput) (*mcp.CallToolResult, *ListConnectionsOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*ListConnectionsOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) handleListConnections(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, any, error) {
	infos := t.ConnectionInfos()

	output := ListConnectionsOutput{
		Connections: make([]ConnectionInfoOutput, len(infos)),
		Count:       len(infos),
	}

	for i, info := range infos {
		output.Connections[i] = ConnectionInfoOutput{
			Name:      info.Name,
			URL:       info.URL,
			IsDefault: info.IsDefault,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return ErrorResult("Failed to marshal connection info"), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, &output, nil
}
