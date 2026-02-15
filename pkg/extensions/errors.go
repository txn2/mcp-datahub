package extensions

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/tools"
)

// errorHint maps an error substring to a helpful hint message.
type errorHint struct {
	substring string
	hint      string
}

// errorHints defines DataHub-specific error patterns and their hints.
var errorHints = []errorHint{
	{
		substring: "entity not found",
		hint:      "Hint: Use datahub_search to find entities.",
	},
	{
		substring: "connection error",
		hint:      "Hint: Use datahub_list_connections to see available connections.",
	},
	{
		substring: "access denied",
		hint:      "Hint: Check your DataHub token and permissions.",
	},
	{
		substring: "unauthorized",
		hint:      "Hint: Check your DataHub token and permissions.",
	},
	{
		substring: "write operations are not enabled",
		hint:      "Hint: Set DATAHUB_WRITE_ENABLED=true to enable writes.",
	},
	{
		substring: "urn parameter is required",
		hint:      "Hint: Use datahub_search to find entity URNs.",
	},
	{
		substring: "invalid urn",
		hint:      "Hint: URNs follow the format urn:li:dataset:(platform,name,env).",
	},
}

// ErrorHintMiddleware enriches error results with helpful hints.
type ErrorHintMiddleware struct{}

// NewErrorHintMiddleware creates an error hint middleware.
func NewErrorHintMiddleware() *ErrorHintMiddleware {
	return &ErrorHintMiddleware{}
}

// Before is a no-op for error hints.
func (m *ErrorHintMiddleware) Before(ctx context.Context, _ *tools.ToolContext) (context.Context, error) {
	return ctx, nil
}

// After appends helpful hints to error results.
func (m *ErrorHintMiddleware) After(
	_ context.Context,
	_ *tools.ToolContext,
	result *mcp.CallToolResult,
	_ error,
) (*mcp.CallToolResult, error) {
	if result == nil || !result.IsError {
		return result, nil
	}

	// Extract error text from the result
	errText := extractResultText(result)
	if errText == "" {
		return result, nil
	}

	// Find matching hint
	lowerErr := strings.ToLower(errText)
	for _, h := range errorHints {
		if strings.Contains(lowerErr, h.substring) {
			return appendTextToResult(result, "\n\n"+h.hint), nil
		}
	}

	return result, nil
}

// extractResultText extracts the first text content from a result.
func extractResultText(result *mcp.CallToolResult) string {
	for _, content := range result.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

// appendTextToResult appends text to the first text content in a result.
func appendTextToResult(result *mcp.CallToolResult, text string) *mcp.CallToolResult {
	for i, content := range result.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			result.Content[i] = &mcp.TextContent{
				Text: tc.Text + text,
			}
			return result
		}
	}
	return result
}
