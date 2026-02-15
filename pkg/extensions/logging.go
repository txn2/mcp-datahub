package extensions

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/tools"
)

// LoggingMiddleware logs tool invocations and results.
type LoggingMiddleware struct {
	output io.Writer
}

// NewLoggingMiddleware creates a logging middleware that writes to the given writer.
func NewLoggingMiddleware(output io.Writer) *LoggingMiddleware {
	return &LoggingMiddleware{output: output}
}

// Before logs the tool invocation.
func (m *LoggingMiddleware) Before(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
	connection := tc.GetString("connection")
	if connection == "" {
		connection = "(default)"
	}
	_, _ = fmt.Fprintf(m.output, "[mcp-datahub] tool=%s connection=%s\n",
		tc.ToolName, connection)
	return ctx, nil
}

// After logs the tool result.
func (m *LoggingMiddleware) After(
	_ context.Context,
	tc *tools.ToolContext,
	result *mcp.CallToolResult,
	handlerErr error,
) (*mcp.CallToolResult, error) {
	duration := time.Since(tc.StartTime)

	switch {
	case handlerErr != nil:
		_, _ = fmt.Fprintf(m.output, "[mcp-datahub] tool=%s duration=%s error=%s\n",
			tc.ToolName, duration.Round(time.Millisecond), handlerErr.Error())
	case result != nil && result.IsError:
		_, _ = fmt.Fprintf(m.output, "[mcp-datahub] tool=%s duration=%s status=error\n",
			tc.ToolName, duration.Round(time.Millisecond))
	default:
		_, _ = fmt.Fprintf(m.output, "[mcp-datahub] tool=%s duration=%s status=ok\n",
			tc.ToolName, duration.Round(time.Millisecond))
	}

	return result, nil
}
