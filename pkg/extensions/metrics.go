package extensions

import (
	"context"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/tools"
)

// MetricsCollector defines the interface for collecting tool metrics.
type MetricsCollector interface {
	// RecordCall records a tool call with its duration and success status.
	RecordCall(toolName string, duration time.Duration, success bool)
}

// MetricsMiddleware collects metrics for tool calls.
type MetricsMiddleware struct {
	collector MetricsCollector
}

// NewMetricsMiddleware creates a metrics middleware with the given collector.
func NewMetricsMiddleware(collector MetricsCollector) *MetricsMiddleware {
	return &MetricsMiddleware{collector: collector}
}

// Before is a no-op for metrics (timing starts from ToolContext.StartTime).
func (m *MetricsMiddleware) Before(ctx context.Context, _ *tools.ToolContext) (context.Context, error) {
	return ctx, nil
}

// After records the tool call metrics.
func (m *MetricsMiddleware) After(
	_ context.Context,
	tc *tools.ToolContext,
	result *mcp.CallToolResult,
	handlerErr error,
) (*mcp.CallToolResult, error) {
	duration := time.Since(tc.StartTime)
	success := handlerErr == nil && (result == nil || !result.IsError)
	m.collector.RecordCall(string(tc.ToolName), duration, success)
	return result, nil
}

// ToolMetrics holds aggregated metrics for a single tool.
type ToolMetrics struct {
	Calls      int64
	Errors     int64
	TotalNanos int64
}

// InMemoryCollector is a thread-safe in-memory metrics collector.
type InMemoryCollector struct {
	mu      sync.Mutex
	metrics map[string]*ToolMetrics
}

// NewInMemoryCollector creates a new in-memory metrics collector.
func NewInMemoryCollector() *InMemoryCollector {
	return &InMemoryCollector{
		metrics: make(map[string]*ToolMetrics),
	}
}

// RecordCall records a tool call metric.
func (c *InMemoryCollector) RecordCall(toolName string, duration time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.metrics[toolName]
	if !ok {
		m = &ToolMetrics{}
		c.metrics[toolName] = m
	}

	m.Calls++
	m.TotalNanos += duration.Nanoseconds()
	if !success {
		m.Errors++
	}
}

// GetMetrics returns a snapshot of metrics for a tool.
// Returns nil if no metrics have been recorded for the tool.
func (c *InMemoryCollector) GetMetrics(toolName string) *ToolMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.metrics[toolName]
	if !ok {
		return nil
	}

	// Return a copy
	return &ToolMetrics{
		Calls:      m.Calls,
		Errors:     m.Errors,
		TotalNanos: m.TotalNanos,
	}
}

// Reset clears all collected metrics.
func (c *InMemoryCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = make(map[string]*ToolMetrics)
}
