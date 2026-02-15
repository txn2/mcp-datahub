package extensions

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/tools"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.EnableLogging {
		t.Error("EnableLogging should be false by default")
	}
	if cfg.EnableMetrics {
		t.Error("EnableMetrics should be false by default")
	}
	if cfg.EnableMetadata {
		t.Error("EnableMetadata should be false by default")
	}
	if !cfg.EnableErrorHelp {
		t.Error("EnableErrorHelp should be true by default")
	}
}

func TestFromEnv(t *testing.T) {
	envVars := map[string]string{
		"MCP_DATAHUB_EXT_LOGGING":  "true",
		"MCP_DATAHUB_EXT_METRICS":  "1",
		"MCP_DATAHUB_EXT_METADATA": "yes",
		"MCP_DATAHUB_EXT_ERRORS":   "false",
	}

	// Save and set
	originals := make(map[string]string)
	for k, v := range envVars {
		originals[k] = os.Getenv(k)
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("failed to set %s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for k, v := range originals {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	})

	cfg := FromEnv()

	if !cfg.EnableLogging {
		t.Error("EnableLogging should be true")
	}
	if !cfg.EnableMetrics {
		t.Error("EnableMetrics should be true")
	}
	if !cfg.EnableMetadata {
		t.Error("EnableMetadata should be true")
	}
	if cfg.EnableErrorHelp {
		t.Error("EnableErrorHelp should be false")
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{" true ", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"", false},
		{"anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseBool(tt.input); got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildToolkitOptions(t *testing.T) {
	// All disabled should return empty
	cfg := Config{}
	opts := BuildToolkitOptions(cfg)
	if len(opts) != 0 {
		t.Errorf("BuildToolkitOptions() with all disabled = %d opts, want 0", len(opts))
	}

	// All enabled should return 4 options
	cfg = Config{
		EnableLogging:   true,
		EnableMetrics:   true,
		EnableMetadata:  true,
		EnableErrorHelp: true,
	}
	opts = BuildToolkitOptions(cfg)
	if len(opts) != 4 {
		t.Errorf("BuildToolkitOptions() with all enabled = %d opts, want 4", len(opts))
	}
}

func TestBuildToolkitOptions_SubsetEnabled(t *testing.T) {
	cfg := Config{
		EnableLogging: true,
		EnableMetrics: true,
	}
	opts := BuildToolkitOptions(cfg)
	if len(opts) != 2 {
		t.Errorf("BuildToolkitOptions() with 2 enabled = %d opts, want 2", len(opts))
	}
}

// Logging middleware tests

func TestLoggingMiddleware_Before(t *testing.T) {
	var buf bytes.Buffer
	mw := NewLoggingMiddleware(&buf)

	tc := tools.NewToolContext(tools.ToolSearch, nil)
	ctx, err := mw.Before(context.Background(), tc)
	if err != nil {
		t.Fatalf("Before() error: %v", err)
	}
	if ctx == nil {
		t.Fatal("Before() returned nil context")
	}

	output := buf.String()
	if !strings.Contains(output, "tool=datahub_search") {
		t.Errorf("Before() output missing tool name: %s", output)
	}
}

func TestLoggingMiddleware_After(t *testing.T) {
	var buf bytes.Buffer
	mw := NewLoggingMiddleware(&buf)

	tc := tools.NewToolContext(tools.ToolSearch, nil)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "test"}},
	}

	got, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}
	if got != result {
		t.Error("After() should return the original result")
	}

	output := buf.String()
	if !strings.Contains(output, "status=ok") {
		t.Errorf("After() output missing status: %s", output)
	}
}

func TestLoggingMiddleware_AfterError(t *testing.T) {
	var buf bytes.Buffer
	mw := NewLoggingMiddleware(&buf)

	tc := tools.NewToolContext(tools.ToolSearch, nil)

	_, err := mw.After(context.Background(), tc, nil, errors.New("test error"))
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "error=test error") {
		t.Errorf("After() output missing error: %s", output)
	}
}

func TestLoggingMiddleware_AfterErrorResult(t *testing.T) {
	var buf bytes.Buffer
	mw := NewLoggingMiddleware(&buf)

	tc := tools.NewToolContext(tools.ToolSearch, nil)
	result := &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: "failed"}},
	}

	_, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "status=error") {
		t.Errorf("After() output missing error status: %s", output)
	}
}

// Metrics middleware tests

func TestMetricsMiddleware(t *testing.T) {
	collector := NewInMemoryCollector()
	mw := NewMetricsMiddleware(collector)

	tc := tools.NewToolContext(tools.ToolSearch, nil)

	// Before should be no-op
	ctx, err := mw.Before(context.Background(), tc)
	if err != nil {
		t.Fatalf("Before() error: %v", err)
	}

	// Simulate some elapsed time
	time.Sleep(5 * time.Millisecond)

	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
	}
	_, err = mw.After(ctx, tc, result, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}

	metrics := collector.GetMetrics("datahub_search")
	if metrics == nil {
		t.Fatal("GetMetrics() returned nil")
	}
	if metrics.Calls != 1 {
		t.Errorf("Calls = %d, want 1", metrics.Calls)
	}
	if metrics.Errors != 0 {
		t.Errorf("Errors = %d, want 0", metrics.Errors)
	}
}

func TestMetricsMiddleware_Error(t *testing.T) {
	collector := NewInMemoryCollector()
	mw := NewMetricsMiddleware(collector)

	tc := tools.NewToolContext(tools.ToolSearch, nil)
	result := &mcp.CallToolResult{IsError: true}

	_, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}

	metrics := collector.GetMetrics("datahub_search")
	if metrics.Errors != 1 {
		t.Errorf("Errors = %d, want 1", metrics.Errors)
	}
}

func TestInMemoryCollector_Reset(t *testing.T) {
	collector := NewInMemoryCollector()
	collector.RecordCall("test_tool", time.Second, true)

	if collector.GetMetrics("test_tool") == nil {
		t.Fatal("expected metrics before reset")
	}

	collector.Reset()

	if collector.GetMetrics("test_tool") != nil {
		t.Error("expected nil metrics after reset")
	}
}

func TestInMemoryCollector_NilForUnknown(t *testing.T) {
	collector := NewInMemoryCollector()
	if collector.GetMetrics("unknown") != nil {
		t.Error("expected nil for unknown tool")
	}
}

// Error hint middleware tests

func TestErrorHintMiddleware_NoError(t *testing.T) {
	mw := NewErrorHintMiddleware()

	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "success"}},
	}

	got, err := mw.After(context.Background(), nil, result, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}

	text := extractResultText(got)
	if text != "success" {
		t.Errorf("should not modify non-error result, got %q", text)
	}
}

func TestErrorHintMiddleware_NilResult(t *testing.T) {
	mw := NewErrorHintMiddleware()

	got, err := mw.After(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}
	if got != nil {
		t.Error("should return nil for nil result")
	}
}

func TestErrorHintMiddleware_Hints(t *testing.T) {
	mw := NewErrorHintMiddleware()

	tests := []struct {
		name     string
		errMsg   string
		wantHint string
	}{
		{
			name:     "entity not found",
			errMsg:   "Entity not found: urn:li:dataset:foo",
			wantHint: "datahub_search",
		},
		{
			name:     "connection error",
			errMsg:   "Connection error: dial tcp timeout",
			wantHint: "datahub_list_connections",
		},
		{
			name:     "unauthorized",
			errMsg:   "Unauthorized: invalid token",
			wantHint: "DataHub token",
		},
		{
			name:     "write disabled",
			errMsg:   "write operations are not enabled",
			wantHint: "DATAHUB_WRITE_ENABLED",
		},
		{
			name:     "urn required",
			errMsg:   "urn parameter is required",
			wantHint: "datahub_search",
		},
		{
			name:     "invalid urn",
			errMsg:   "invalid urn format",
			wantHint: "urn:li:dataset:",
		},
		{
			name:     "no matching hint",
			errMsg:   "some random error",
			wantHint: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: tt.errMsg}},
			}

			got, err := mw.After(context.Background(), nil, result, nil)
			if err != nil {
				t.Fatalf("After() error: %v", err)
			}

			text := extractResultText(got)
			if tt.wantHint == "" {
				if text != tt.errMsg {
					t.Errorf("should not modify unmatched error, got %q", text)
				}
			} else {
				if !strings.Contains(text, tt.wantHint) {
					t.Errorf("hint should contain %q, got %q", tt.wantHint, text)
				}
			}
		})
	}
}

// Metadata middleware tests

func TestMetadataMiddleware_Success(t *testing.T) {
	mw := NewMetadataMiddleware()

	tc := tools.NewToolContext(tools.ToolSearch, nil)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "data"}},
	}

	got, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}

	text := extractResultText(got)
	if !strings.Contains(text, "tool: datahub_search") {
		t.Errorf("metadata should contain tool name, got %q", text)
	}
	if !strings.Contains(text, "duration:") {
		t.Errorf("metadata should contain duration, got %q", text)
	}
	if !strings.Contains(text, "timestamp:") {
		t.Errorf("metadata should contain timestamp, got %q", text)
	}
}

func TestMetadataMiddleware_ErrorResult(t *testing.T) {
	mw := NewMetadataMiddleware()

	tc := tools.NewToolContext(tools.ToolSearch, nil)
	result := &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: "error"}},
	}

	got, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}

	text := extractResultText(got)
	if text != "error" {
		t.Errorf("should not modify error result, got %q", text)
	}
}

func TestMetadataMiddleware_NilResult(t *testing.T) {
	mw := NewMetadataMiddleware()

	tc := tools.NewToolContext(tools.ToolSearch, nil)

	got, err := mw.After(context.Background(), tc, nil, nil)
	if err != nil {
		t.Fatalf("After() error: %v", err)
	}
	if got != nil {
		t.Error("should return nil for nil result")
	}
}

// Interface compliance tests

func TestLoggingMiddleware_ImplementsToolMiddleware(t *testing.T) {
	var _ tools.ToolMiddleware = (*LoggingMiddleware)(nil)
}

func TestMetricsMiddleware_ImplementsToolMiddleware(t *testing.T) {
	var _ tools.ToolMiddleware = (*MetricsMiddleware)(nil)
}

func TestErrorHintMiddleware_ImplementsToolMiddleware(t *testing.T) {
	var _ tools.ToolMiddleware = (*ErrorHintMiddleware)(nil)
}

func TestMetadataMiddleware_ImplementsToolMiddleware(t *testing.T) {
	var _ tools.ToolMiddleware = (*MetadataMiddleware)(nil)
}

func TestInMemoryCollector_ImplementsMetricsCollector(t *testing.T) {
	var _ MetricsCollector = (*InMemoryCollector)(nil)
}

// Helper tests

func TestExtractResultText_Empty(t *testing.T) {
	result := &mcp.CallToolResult{Content: []mcp.Content{}}
	if got := extractResultText(result); got != "" {
		t.Errorf("extractResultText() = %q, want empty", got)
	}
}

func TestAppendTextToResult_NoTextContent(t *testing.T) {
	result := &mcp.CallToolResult{Content: []mcp.Content{}}
	got := appendTextToResult(result, "extra")
	if len(got.Content) != 0 {
		t.Error("should not add content to empty result")
	}
}
