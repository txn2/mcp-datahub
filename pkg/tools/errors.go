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

// errTargetConflict returns an error when target_urn and value both carry a
// URN for an association operation but disagree, rather than silently
// preferring one.
func errTargetConflict(targetURN, value string) error {
	return fmt.Errorf("target_urn (%s) and value (%s) differ; set only target_urn", targetURN, value)
}

// errDescriptionConflict returns an error when value and description both carry
// description text for a description update but disagree, rather than silently
// preferring one.
func errDescriptionConflict(value, description string) error {
	return fmt.Errorf("value (%s) and description (%s) differ; set only value", value, description)
}

// JSONResult creates a JSON result for tool responses.
func JSONResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return TextResult(string(data)), nil
}
