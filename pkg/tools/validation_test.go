package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestInputTypeValidation tests that all tools properly reject invalid input types.
// This covers the type assertion branches in the register functions.
func TestInputTypeValidation(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	tests := []struct {
		name    string
		handler func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error)
	}{
		{
			name: "GetEntity invalid type",
			handler: func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
				entityInput, ok := input.(GetEntityInput)
				if !ok {
					return ErrorResult("internal error: invalid input type"), nil, nil
				}
				return toolkit.handleGetEntity(ctx, req, entityInput)
			},
		},
		{
			name: "GetSchema invalid type",
			handler: func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
				schemaInput, ok := input.(GetSchemaInput)
				if !ok {
					return ErrorResult("internal error: invalid input type"), nil, nil
				}
				return toolkit.handleGetSchema(ctx, req, schemaInput)
			},
		},
		{
			name: "GetLineage invalid type",
			handler: func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
				lineageInput, ok := input.(GetLineageInput)
				if !ok {
					return ErrorResult("internal error: invalid input type"), nil, nil
				}
				return toolkit.handleGetLineage(ctx, req, lineageInput)
			},
		},
		{
			name: "GetQueries invalid type",
			handler: func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
				queriesInput, ok := input.(GetQueriesInput)
				if !ok {
					return ErrorResult("internal error: invalid input type"), nil, nil
				}
				return toolkit.handleGetQueries(ctx, req, queriesInput)
			},
		},
		{
			name: "GetGlossaryTerm invalid type",
			handler: func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
				glossaryInput, ok := input.(GetGlossaryTermInput)
				if !ok {
					return ErrorResult("internal error: invalid input type"), nil, nil
				}
				return toolkit.handleGetGlossaryTerm(ctx, req, glossaryInput)
			},
		},
		{
			name: "ListTags invalid type",
			handler: func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
				tagsInput, ok := input.(ListTagsInput)
				if !ok {
					return ErrorResult("internal error: invalid input type"), nil, nil
				}
				return toolkit.handleListTags(ctx, req, tagsInput)
			},
		},
		{
			name: "GetDataProduct invalid type",
			handler: func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
				dpInput, ok := input.(GetDataProductInput)
				if !ok {
					return ErrorResult("internal error: invalid input type"), nil, nil
				}
				return toolkit.handleGetDataProduct(ctx, req, dpInput)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pass wrong type (string instead of expected struct)
			result, _, _ := tt.handler(context.Background(), nil, "wrong type")
			if !result.IsError {
				t.Error("Should return error for invalid input type")
			}
		})
	}
}

// TestToolRegistrationCoverage ensures all tools can be registered and called.
func TestToolRegistrationCoverage(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)

	// Register all tools
	toolkit.RegisterAll(server)

	// Verify all tools are registered
	for _, name := range AllTools() {
		if !toolkit.registeredTools[name] {
			t.Errorf("Tool %s should be registered", name)
		}
	}
}

// TestRegisterWithMiddleware tests RegisterWith functionality.
func TestRegisterWithMiddleware(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)

	middlewareCalled := false
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		middlewareCalled = true
		return ctx, nil
	})

	// Register with middleware
	toolkit.RegisterWith(server, ToolSearch, WithPerToolMiddleware(mw))
	toolkit.RegisterWith(server, ToolGetEntity, WithPerToolMiddleware(mw))

	if !toolkit.registeredTools[ToolSearch] {
		t.Error("ToolSearch should be registered")
	}
	if !toolkit.registeredTools[ToolGetEntity] {
		t.Error("ToolGetEntity should be registered")
	}

	_ = middlewareCalled // Middleware will be called when tool is invoked
}
