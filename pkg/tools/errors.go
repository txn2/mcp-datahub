package tools

import (
	"encoding/json"
	"fmt"
	"strings"

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

// errRequired returns an error indicating a required parameter is missing.
func errRequired(param string) error {
	return fmt.Errorf("%s parameter is required", param)
}

// errInvalidAction returns an error indicating the action value is not valid for the operation.
func errInvalidAction(got string, valid ...string) error {
	return fmt.Errorf("action must be '%s', got '%s'", strings.Join(valid, "' or '"), got)
}

// JSONResult creates a JSON result for tool responses.
func JSONResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return TextResult(string(data)), nil
}
