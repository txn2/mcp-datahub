package tools

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrorResult creates an error result for tool responses.
func ErrorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
		IsError: true,
	}
}

// TextResult creates a text result for tool responses.
func TextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// JSONResult creates a JSON result for tool responses.
func JSONResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return TextResult(string(data)), nil
}
