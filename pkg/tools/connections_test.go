package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
)

func TestHandleListConnections_SingleClient(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, err := toolkit.handleListConnections(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("result should not be an error")
	}

	// Parse the result
	var output ListConnectionsOutput
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &output); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if output.Count != 1 {
		t.Errorf("expected 1 connection, got %d", output.Count)
	}
	if len(output.Connections) != 1 {
		t.Errorf("expected 1 connection info, got %d", len(output.Connections))
	}
	if !output.Connections[0].IsDefault {
		t.Error("single connection should be default")
	}
	if output.Connections[0].Name != "default" {
		t.Errorf("expected name 'default', got %q", output.Connections[0].Name)
	}
}

func TestHandleListConnections_MultiServer(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]multiserver.ConnectionConfig{
			"staging": {
				URL:   "https://staging.datahub.example.com",
				Token: "staging-token",
			},
			"dev": {
				URL: "https://dev.datahub.example.com",
			},
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	result, _, err := toolkit.handleListConnections(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("result should not be an error")
	}

	// Parse the result
	var output ListConnectionsOutput
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &output); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if output.Count != 3 {
		t.Errorf("expected 3 connections, got %d", output.Count)
	}
	if len(output.Connections) != 3 {
		t.Errorf("expected 3 connection infos, got %d", len(output.Connections))
	}

	// Verify default connection
	var foundDefault bool
	for _, conn := range output.Connections {
		if conn.IsDefault {
			foundDefault = true
			if conn.Name != "prod" {
				t.Errorf("expected default name 'prod', got %q", conn.Name)
			}
			if conn.URL != "https://prod.datahub.example.com" {
				t.Errorf("expected default URL 'https://prod.datahub.example.com', got %q", conn.URL)
			}
		}
	}
	if !foundDefault {
		t.Error("no default connection found")
	}
}

func TestRegisterListConnectionsTool(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)

	// Register the tool via the public method which tracks registration
	toolkit.Register(server, ToolListConnections)

	// Verify registration tracking
	if !toolkit.registeredTools[ToolListConnections] {
		t.Error("ToolListConnections should be registered")
	}
}

func TestRegisterListConnectionsTool_WithMiddleware(t *testing.T) {
	mock := &mockClient{}

	middlewareCalled := false
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		middlewareCalled = true
		return ctx, nil
	})

	toolkit := NewToolkit(mock, DefaultConfig(), WithMiddleware(mw))

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)

	// Register the tool
	toolkit.registerListConnectionsTool(server, nil)

	// Create wrapped handler and call it to verify middleware works
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		return toolkit.handleListConnections(ctx, req)
	}
	wrapped := toolkit.wrapHandler(ToolListConnections, baseHandler, nil)

	_, _, _ = wrapped(context.Background(), nil, ListConnectionsInput{})

	if !middlewareCalled {
		t.Error("middleware should have been called")
	}
}

func TestConnectionInfoOutput_Fields(t *testing.T) {
	output := ConnectionInfoOutput{
		Name:      "prod",
		URL:       "https://prod.datahub.example.com",
		IsDefault: true,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["name"] != "prod" {
		t.Errorf("expected name 'prod', got %v", parsed["name"])
	}
	if parsed["url"] != "https://prod.datahub.example.com" {
		t.Errorf("expected url 'https://prod.datahub.example.com', got %v", parsed["url"])
	}
	if parsed["is_default"] != true {
		t.Errorf("expected is_default true, got %v", parsed["is_default"])
	}
}

func TestListConnectionsOutput_JSONFormat(t *testing.T) {
	output := ListConnectionsOutput{
		Connections: []ConnectionInfoOutput{
			{Name: "prod", URL: "https://prod.example.com", IsDefault: true},
			{Name: "staging", URL: "https://staging.example.com", IsDefault: false},
		},
		Count: 2,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed ListConnectionsOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Count != 2 {
		t.Errorf("expected count 2, got %d", parsed.Count)
	}
	if len(parsed.Connections) != 2 {
		t.Errorf("expected 2 connections, got %d", len(parsed.Connections))
	}
}

func TestHandleListConnections_EmptyManager(t *testing.T) {
	// Test with manager that has only default connection
	cfg := multiserver.Config{
		Default: "datahub",
		Primary: client.Config{
			URL:   "https://datahub.example.com",
			Token: "token",
		},
		Connections: map[string]multiserver.ConnectionConfig{},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	result, _, err := toolkit.handleListConnections(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("result should not be an error")
	}

	var output ListConnectionsOutput
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if output.Count != 1 {
		t.Errorf("expected 1 connection, got %d", output.Count)
	}
}

func TestListConnectionsInput_Empty(t *testing.T) {
	// Verify the input struct has no required fields
	input := ListConnectionsInput{}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if string(data) != "{}" {
		t.Errorf("expected empty JSON object, got %s", string(data))
	}
}

func TestHandleListConnections_MultiServerWithInheritance(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:             "https://prod.datahub.example.com",
			Token:           "prod-token",
			Timeout:         30 * time.Second,
			DefaultLimit:    10,
			MaxLimit:        100,
			MaxLineageDepth: 5,
		},
		Connections: map[string]multiserver.ConnectionConfig{
			"staging": {
				URL: "https://staging.datahub.example.com",
				// Token inherits from primary
			},
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	result, _, err := toolkit.handleListConnections(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output ListConnectionsOutput
	text := result.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Should have 2 connections
	if output.Count != 2 {
		t.Errorf("expected 2 connections, got %d", output.Count)
	}

	// Verify staging connection info
	var stagingFound bool
	for _, conn := range output.Connections {
		if conn.Name == "staging" {
			stagingFound = true
			if conn.URL != "https://staging.datahub.example.com" {
				t.Errorf("expected staging URL 'https://staging.datahub.example.com', got %q", conn.URL)
			}
			if conn.IsDefault {
				t.Error("staging should not be default")
			}
		}
	}
	if !stagingFound {
		t.Error("staging connection not found in output")
	}
}
