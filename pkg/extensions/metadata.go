package extensions

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/tools"
)

// MetadataMiddleware appends execution metadata to tool results.
type MetadataMiddleware struct{}

// NewMetadataMiddleware creates a metadata enrichment middleware.
func NewMetadataMiddleware() *MetadataMiddleware {
	return &MetadataMiddleware{}
}

// Before is a no-op for metadata enrichment.
func (m *MetadataMiddleware) Before(ctx context.Context, _ *tools.ToolContext) (context.Context, error) {
	return ctx, nil
}

// After appends execution metadata to successful results.
func (m *MetadataMiddleware) After(
	_ context.Context,
	tc *tools.ToolContext,
	result *mcp.CallToolResult,
	_ error,
) (*mcp.CallToolResult, error) {
	if result == nil || result.IsError {
		return result, nil
	}

	duration := time.Since(tc.StartTime)
	footer := fmt.Sprintf("\n\n---\ntool: %s | duration: %s | timestamp: %s",
		tc.ToolName,
		duration.Round(time.Millisecond),
		tc.StartTime.UTC().Format(time.RFC3339),
	)

	return appendTextToResult(result, footer), nil
}
