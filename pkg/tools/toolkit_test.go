package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// mockClient implements DataHubClient for testing.
type mockClient struct {
	searchFunc           func(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error)
	getEntityFunc        func(ctx context.Context, urn string) (*types.Entity, error)
	getSchemaFunc        func(ctx context.Context, urn string) (*types.SchemaMetadata, error)
	getLineageFunc       func(ctx context.Context, urn string, opts ...client.LineageOption) (*types.LineageResult, error)
	getQueriesFunc       func(ctx context.Context, urn string) (*types.QueryList, error)
	getGlossaryTermFunc  func(ctx context.Context, urn string) (*types.GlossaryTerm, error)
	listTagsFunc         func(ctx context.Context, filter string) ([]types.Tag, error)
	listDomainsFunc      func(ctx context.Context) ([]types.Domain, error)
	listDataProductsFunc func(ctx context.Context) ([]types.DataProduct, error)
	getDataProductFunc   func(ctx context.Context, urn string) (*types.DataProduct, error)
	pingFunc             func(ctx context.Context) error
}

func (m *mockClient) Search(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, query, opts...)
	}
	return &types.SearchResult{}, nil
}

func (m *mockClient) GetEntity(ctx context.Context, urn string) (*types.Entity, error) {
	if m.getEntityFunc != nil {
		return m.getEntityFunc(ctx, urn)
	}
	return &types.Entity{URN: urn}, nil
}

func (m *mockClient) GetSchema(ctx context.Context, urn string) (*types.SchemaMetadata, error) {
	if m.getSchemaFunc != nil {
		return m.getSchemaFunc(ctx, urn)
	}
	return &types.SchemaMetadata{}, nil
}

func (m *mockClient) GetLineage(ctx context.Context, urn string, opts ...client.LineageOption) (*types.LineageResult, error) {
	if m.getLineageFunc != nil {
		return m.getLineageFunc(ctx, urn, opts...)
	}
	return &types.LineageResult{Start: urn}, nil
}

func (m *mockClient) GetQueries(ctx context.Context, urn string) (*types.QueryList, error) {
	if m.getQueriesFunc != nil {
		return m.getQueriesFunc(ctx, urn)
	}
	return &types.QueryList{}, nil
}

func (m *mockClient) GetGlossaryTerm(ctx context.Context, urn string) (*types.GlossaryTerm, error) {
	if m.getGlossaryTermFunc != nil {
		return m.getGlossaryTermFunc(ctx, urn)
	}
	return &types.GlossaryTerm{URN: urn}, nil
}

func (m *mockClient) ListTags(ctx context.Context, filter string) ([]types.Tag, error) {
	if m.listTagsFunc != nil {
		return m.listTagsFunc(ctx, filter)
	}
	return []types.Tag{}, nil
}

func (m *mockClient) ListDomains(ctx context.Context) ([]types.Domain, error) {
	if m.listDomainsFunc != nil {
		return m.listDomainsFunc(ctx)
	}
	return []types.Domain{}, nil
}

func (m *mockClient) ListDataProducts(ctx context.Context) ([]types.DataProduct, error) {
	if m.listDataProductsFunc != nil {
		return m.listDataProductsFunc(ctx)
	}
	return []types.DataProduct{}, nil
}

func (m *mockClient) GetDataProduct(ctx context.Context, urn string) (*types.DataProduct, error) {
	if m.getDataProductFunc != nil {
		return m.getDataProductFunc(ctx, urn)
	}
	return &types.DataProduct{URN: urn}, nil
}

func (m *mockClient) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func (m *mockClient) Close() error {
	return nil
}

func TestNewToolkit(t *testing.T) {
	mock := &mockClient{}
	cfg := DefaultConfig()

	toolkit := NewToolkit(mock, cfg)

	if toolkit == nil {
		t.Fatal("NewToolkit() returned nil")
	}

	if toolkit.Client() != mock {
		t.Error("Client() should return the mock client")
	}

	config := toolkit.Config()
	if config.DefaultLimit != 10 {
		t.Errorf("Config() DefaultLimit = %d, want 10", config.DefaultLimit)
	}
}

func TestNewToolkitWithOptions(t *testing.T) {
	mock := &mockClient{}
	cfg := DefaultConfig()

	middlewareCalled := false
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		middlewareCalled = true
		return ctx, nil
	})

	toolkit := NewToolkit(mock, cfg,
		WithMiddleware(mw),
		WithToolMiddleware(ToolSearch, mw),
	)

	if !toolkit.HasMiddleware() {
		t.Error("HasMiddleware() should return true")
	}

	// The middleware would be called when tools are invoked
	_ = middlewareCalled // Just verifying the setup worked
}

func TestToolkitRegisterAll(t *testing.T) {
	mock := &mockClient{}
	cfg := DefaultConfig()
	toolkit := NewToolkit(mock, cfg)

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.RegisterAll(server)

	// Verify all tools are registered by checking internal tracking
	for _, name := range AllTools() {
		if !toolkit.registeredTools[name] {
			t.Errorf("RegisterAll() should register %s", name)
		}
	}
}

func TestToolkitRegister(t *testing.T) {
	mock := &mockClient{}
	cfg := DefaultConfig()
	toolkit := NewToolkit(mock, cfg)

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolSearch, ToolGetEntity)

	if !toolkit.registeredTools[ToolSearch] {
		t.Error("Register() should register ToolSearch")
	}
	if !toolkit.registeredTools[ToolGetEntity] {
		t.Error("Register() should register ToolGetEntity")
	}
	if toolkit.registeredTools[ToolGetSchema] {
		t.Error("Register() should not register ToolGetSchema")
	}
}

func TestToolkitRegisterDuplicate(t *testing.T) {
	mock := &mockClient{}
	cfg := DefaultConfig()
	toolkit := NewToolkit(mock, cfg)

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)

	// Register same tool twice - should not panic
	toolkit.Register(server, ToolSearch)
	toolkit.Register(server, ToolSearch)

	// Should only be registered once (internal tracking)
	if !toolkit.registeredTools[ToolSearch] {
		t.Error("Register() should register ToolSearch")
	}
}

func TestToolkitRegisterWith(t *testing.T) {
	mock := &mockClient{}
	cfg := DefaultConfig()
	toolkit := NewToolkit(mock, cfg)

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)

	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, nil
	})

	toolkit.RegisterWith(server, ToolSearch, WithPerToolMiddleware(mw))

	if !toolkit.registeredTools[ToolSearch] {
		t.Error("RegisterWith() should register ToolSearch")
	}
}

func TestToolkitHasMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*Toolkit)
		wantResult bool
	}{
		{
			name:       "no middleware",
			setupFunc:  func(_ *Toolkit) {},
			wantResult: false,
		},
		{
			name: "global middleware",
			setupFunc: func(t *Toolkit) {
				t.middlewares = append(t.middlewares, BeforeFunc(nil))
			},
			wantResult: true,
		},
		{
			name: "tool-specific middleware",
			setupFunc: func(t *Toolkit) {
				t.toolMiddlewares[ToolSearch] = append(t.toolMiddlewares[ToolSearch], BeforeFunc(nil))
			},
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{}
			toolkit := NewToolkit(mock, DefaultConfig())
			tt.setupFunc(toolkit)

			if toolkit.HasMiddleware() != tt.wantResult {
				t.Errorf("HasMiddleware() = %v, want %v", toolkit.HasMiddleware(), tt.wantResult)
			}
		})
	}
}

func TestToolkitMiddlewareExecution(t *testing.T) {
	mock := &mockClient{}
	mock.searchFunc = func(_ context.Context, query string, _ ...client.SearchOption) (*types.SearchResult, error) {
		return &types.SearchResult{
			Total: 1,
			Entities: []types.SearchEntity{
				{URN: "urn:li:dataset:test", Name: "test"},
			},
		}, nil
	}

	beforeCalled := false
	afterCalled := false

	beforeMW := BeforeFunc(func(ctx context.Context, tc *ToolContext) (context.Context, error) {
		beforeCalled = true
		tc.Set("before", true)
		return ctx, nil
	})

	afterMW := AfterFunc(func(_ context.Context, tc *ToolContext, result *mcp.CallToolResult, _ error) (*mcp.CallToolResult, error) {
		afterCalled = true
		if _, ok := tc.Get("before"); !ok {
			return ErrorResult("before not set"), nil
		}
		return result, nil
	})

	toolkit := NewToolkit(mock, DefaultConfig(),
		WithMiddleware(beforeMW),
		WithMiddleware(afterMW),
	)

	// Test with a wrapped handler
	handler := func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		return TextResult("test"), nil, nil
	}

	wrapped := toolkit.wrapHandler(ToolSearch, handler, nil)
	result, _, _ := wrapped(context.Background(), nil, SearchInput{Query: "test"})

	if !beforeCalled {
		t.Error("Before middleware should be called")
	}
	if !afterCalled {
		t.Error("After middleware should be called")
	}
	if result.IsError {
		t.Error("Result should not be an error")
	}
}

func TestToolkitMiddlewareError(t *testing.T) {
	mock := &mockClient{}

	expectedErr := errors.New("middleware error")
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, expectedErr
	})

	toolkit := NewToolkit(mock, DefaultConfig(), WithMiddleware(mw))

	handler := func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		return TextResult("test"), nil, nil
	}

	wrapped := toolkit.wrapHandler(ToolSearch, handler, nil)
	result, _, _ := wrapped(context.Background(), nil, SearchInput{Query: "test"})

	if !result.IsError {
		t.Error("Result should be an error when middleware fails")
	}
}

func TestToolkitNoMiddleware(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	handlerCalled := false
	handler := func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		handlerCalled = true
		return TextResult("test"), nil, nil
	}

	wrapped := toolkit.wrapHandler(ToolSearch, handler, nil)

	// With no middleware, the handler should be returned unchanged
	_, _, _ = wrapped(context.Background(), nil, SearchInput{Query: "test"})

	if !handlerCalled {
		t.Error("Handler should be called")
	}
}

func TestNewToolkitWithManager(t *testing.T) {
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
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	if toolkit == nil {
		t.Fatal("NewToolkitWithManager() returned nil")
	}
	if !toolkit.HasManager() {
		t.Error("HasManager() should return true")
	}
	if toolkit.Manager() != mgr {
		t.Error("Manager() should return the manager")
	}
	if toolkit.Client() != nil {
		t.Error("Client() should return nil in manager mode")
	}
}

func TestNewToolkitWithManager_WithOptions(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	middlewareCalled := false
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		middlewareCalled = true
		return ctx, nil
	})

	toolkit := NewToolkitWithManager(mgr, DefaultConfig(), WithMiddleware(mw))

	if !toolkit.HasMiddleware() {
		t.Error("HasMiddleware() should return true")
	}

	// Verify middleware setup
	_ = middlewareCalled
}

func TestToolkitHasManager(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() *Toolkit
		wantResult bool
	}{
		{
			name: "single client mode",
			setupFunc: func() *Toolkit {
				return NewToolkit(&mockClient{}, DefaultConfig())
			},
			wantResult: false,
		},
		{
			name: "manager mode",
			setupFunc: func() *Toolkit {
				cfg := multiserver.Config{
					Default: "default",
					Primary: client.Config{URL: "https://localhost", Token: "token"},
				}
				mgr := multiserver.NewManager(cfg)
				return NewToolkitWithManager(mgr, DefaultConfig())
			},
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolkit := tt.setupFunc()
			if toolkit.HasManager() != tt.wantResult {
				t.Errorf("HasManager() = %v, want %v", toolkit.HasManager(), tt.wantResult)
			}
		})
	}
}

func TestToolkitManager(t *testing.T) {
	// Test nil manager in single client mode
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())
	if toolkit.Manager() != nil {
		t.Error("Manager() should return nil in single client mode")
	}

	// Test non-nil manager in manager mode
	cfg := multiserver.Config{
		Default: "default",
		Primary: client.Config{URL: "https://localhost", Token: "token"},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit = NewToolkitWithManager(mgr, DefaultConfig())
	if toolkit.Manager() != mgr {
		t.Error("Manager() should return the manager")
	}
}

func TestToolkitConnectionInfos_SingleClient(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	infos := toolkit.ConnectionInfos()

	if len(infos) != 1 {
		t.Errorf("expected 1 connection info, got %d", len(infos))
	}
	if infos[0].Name != "default" {
		t.Errorf("expected name 'default', got %q", infos[0].Name)
	}
	if !infos[0].IsDefault {
		t.Error("single connection should be default")
	}
	if infos[0].URL != "configured via single client" {
		t.Errorf("unexpected URL: %q", infos[0].URL)
	}
}

func TestToolkitConnectionInfos_MultiServer(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]multiserver.ConnectionConfig{
			"staging": {URL: "https://staging.datahub.example.com"},
			"dev":     {URL: "https://dev.datahub.example.com"},
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())
	infos := toolkit.ConnectionInfos()

	if len(infos) != 3 {
		t.Errorf("expected 3 connection infos, got %d", len(infos))
	}

	// Verify default connection
	var foundDefault bool
	for _, info := range infos {
		if info.IsDefault {
			foundDefault = true
			if info.Name != "prod" {
				t.Errorf("expected default name 'prod', got %q", info.Name)
			}
		}
	}
	if !foundDefault {
		t.Error("no default connection found")
	}
}

func TestToolkitConnectionCount_SingleClient(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	if toolkit.ConnectionCount() != 1 {
		t.Errorf("expected 1, got %d", toolkit.ConnectionCount())
	}
}

func TestToolkitConnectionCount_MultiServer(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]multiserver.ConnectionConfig{
			"staging": {URL: "https://staging.datahub.example.com"},
			"dev":     {URL: "https://dev.datahub.example.com"},
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	if toolkit.ConnectionCount() != 3 {
		t.Errorf("expected 3, got %d", toolkit.ConnectionCount())
	}
}

func TestToolkitGetClient_SingleClientMode(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	// Empty connection name should return the single client
	c, err := toolkit.getClient("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != mock {
		t.Error("expected mock client")
	}

	// Any connection name should still return the single client
	c, err = toolkit.getClient("anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != mock {
		t.Error("expected mock client even with connection name")
	}
}

func TestToolkitGetClient_SingleClientMode_NoClient(t *testing.T) {
	// Create toolkit without client
	toolkit := &Toolkit{
		config:          DefaultConfig(),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	_, err := toolkit.getClient("")
	if err == nil {
		t.Error("expected error when no client configured")
	}
	if err.Error() != "no client configured" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestToolkitGetClient_MultiServerMode(t *testing.T) {
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
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	// Empty connection name returns default client
	c1, err := toolkit.getClient("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c1 == nil {
		t.Error("expected non-nil client")
	}

	// Explicit connection name
	c2, err := toolkit.getClient("staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c2 == nil {
		t.Error("expected non-nil client")
	}

	// Clients should be different
	if c1 == c2 {
		t.Error("expected different clients for different connections")
	}
}

func TestToolkitGetClient_MultiServerMode_UnknownConnection(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	_, err := toolkit.getClient("unknown")
	if err == nil {
		t.Error("expected error for unknown connection")
	}
}

func TestToolkitRegisterAll_MultiServer(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.RegisterAll(server)

	// Verify all tools are registered
	for _, name := range AllTools() {
		if !toolkit.registeredTools[name] {
			t.Errorf("RegisterAll() should register %s", name)
		}
	}

	// Verify ToolListConnections is included
	if !toolkit.registeredTools[ToolListConnections] {
		t.Error("RegisterAll() should register ToolListConnections")
	}
}
