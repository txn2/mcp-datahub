package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// mockClient implements DataHubClient for testing.
type mockClient struct {
	searchFunc                            func(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error)
	getEntityFunc                         func(ctx context.Context, urn string) (*types.Entity, error)
	getSchemaFunc                         func(ctx context.Context, urn string) (*types.SchemaMetadata, error)
	getSchemasFunc                        func(ctx context.Context, urns []string) (map[string]*types.SchemaMetadata, error)
	getLineageFunc                        func(ctx context.Context, urn string, opts ...client.LineageOption) (*types.LineageResult, error)
	getColumnLineageFunc                  func(ctx context.Context, urn string) (*types.ColumnLineage, error)
	getQueriesFunc                        func(ctx context.Context, urn string) (*types.QueryList, error)
	getGlossaryTermFunc                   func(ctx context.Context, urn string) (*types.GlossaryTerm, error)
	listTagsFunc                          func(ctx context.Context, filter string) ([]types.Tag, error)
	listDomainsFunc                       func(ctx context.Context) ([]types.Domain, error)
	listDataProductsFunc                  func(ctx context.Context) ([]types.DataProduct, error)
	getDataProductFunc                    func(ctx context.Context, urn string) (*types.DataProduct, error)
	pingFunc                              func(ctx context.Context) error
	updateDescriptionFunc                 func(ctx context.Context, urn, description string) error
	addTagFunc                            func(ctx context.Context, urn, tagURN string) error
	removeTagFunc                         func(ctx context.Context, urn, tagURN string) error
	addGlossaryTermFunc                   func(ctx context.Context, urn, termURN string) error
	removeGlossaryTermFunc                func(ctx context.Context, urn, termURN string) error
	addLinkFunc                           func(ctx context.Context, urn, linkURL, description string) error
	removeLinkFunc                        func(ctx context.Context, urn, linkURL string) error
	getStructuredPropertiesFunc           func(ctx context.Context, urn string) ([]types.StructuredPropertyValue, error)
	listStructuredPropertyDefinitionsFunc func(ctx context.Context) ([]types.StructuredPropertyDefinition, error)
	upsertStructuredPropertiesFunc        func(ctx context.Context, urn string, properties []types.StructuredPropertyInput) error
	removeStructuredPropertiesFunc        func(ctx context.Context, urn string, propertyURNs []string) error
	getIncidentsFunc                      func(ctx context.Context, urn string) (*types.IncidentResult, error)
	raiseIncidentFunc                     func(ctx context.Context, input types.RaiseIncidentInput) (string, error)
	updateIncidentStatusFunc              func(ctx context.Context, incidentURN, state, message string) error
	resolveIncidentFunc                   func(ctx context.Context, incidentURN, message string) error
	getDataContractFunc                   func(ctx context.Context, datasetURN string) (*types.DataContract, error)
	semanticSearchFunc                    func(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error)
	// Context documents (DataHub 1.4.x+)
	getDocumentFunc      func(ctx context.Context, urn string) (*types.Document, error)
	getRelatedDocsFunc   func(ctx context.Context, urn string) ([]types.Document, error)
	getContextDocsFunc   func(ctx context.Context, urn string) ([]types.ContextDocument, error)
	upsertContextDocFunc func(ctx context.Context, entityURN string,
		doc types.ContextDocumentInput) (*types.ContextDocument, error)
	deleteContextDocFunc func(ctx context.Context, documentID string) error

	// New CRUD methods
	updateColumnDescriptionFunc  func(ctx context.Context, urn, fieldPath, description string) error
	createQueryFunc              func(ctx context.Context, input client.CreateQueryInput) (*types.Query, error)
	updateQueryFunc              func(ctx context.Context, input client.UpdateQueryInput) (*types.Query, error)
	deleteQueryFunc              func(ctx context.Context, urn string) error
	createTagFunc                func(ctx context.Context, name, description string) (string, error)
	createDomainFunc             func(ctx context.Context, name, description string) (string, error)
	createGlossaryTermFunc       func(ctx context.Context, name, description, parentNode string) (string, error)
	createDataProductFunc        func(ctx context.Context, name, description, domainURN string) (string, error)
	createDocumentFunc           func(ctx context.Context, input types.CreateDocumentInput) (string, error)
	createApplicationFunc        func(ctx context.Context, name, description string) (string, error)
	createStructuredPropertyFunc func(ctx context.Context, input types.CreateStructuredPropertyInput) (string, error)
	upsertDataContractFunc       func(ctx context.Context, input types.UpsertDataContractInput) (string, error)
	addOwnerFunc                 func(ctx context.Context, urn, ownerURN, ownershipType string) error
	removeOwnerFunc              func(ctx context.Context, urn, ownerURN string) error
	setDomainFunc                func(ctx context.Context, entityURN, domainURN string) error
	unsetDomainFunc              func(ctx context.Context, entityURN string) error
	updateIncidentFunc           func(ctx context.Context, urn string, input types.UpdateIncidentInput) error
	updateStructuredPropertyFunc func(ctx context.Context, urn string, input types.UpdateStructuredPropertyInput) error
	updateDocumentContentsFunc   func(ctx context.Context, urn, title, text string) error
	updateDocumentStatusFunc     func(ctx context.Context, urn, status string) error
	updateDocRelatedFunc         func(ctx context.Context, urn string, entityURNs []string) error
	updateDocSubTypeFunc         func(ctx context.Context, urn, subType string) error
	deleteTagFunc                func(ctx context.Context, urn string) error
	deleteDomainFunc             func(ctx context.Context, urn string) error
	deleteGlossaryEntityFunc     func(ctx context.Context, urn string) error
	deleteDataProductFunc        func(ctx context.Context, urn string) error
	deleteApplicationFunc        func(ctx context.Context, urn string) error
	deleteDocumentFunc           func(ctx context.Context, urn string) error
	deleteStructuredPropertyFunc func(ctx context.Context, urn string) error
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

func (m *mockClient) GetSchemas(ctx context.Context, urns []string) (map[string]*types.SchemaMetadata, error) {
	if m.getSchemasFunc != nil {
		return m.getSchemasFunc(ctx, urns)
	}
	return make(map[string]*types.SchemaMetadata), nil
}

func (m *mockClient) GetLineage(ctx context.Context, urn string, opts ...client.LineageOption) (*types.LineageResult, error) {
	if m.getLineageFunc != nil {
		return m.getLineageFunc(ctx, urn, opts...)
	}
	return &types.LineageResult{Start: urn}, nil
}

func (m *mockClient) GetColumnLineage(ctx context.Context, urn string) (*types.ColumnLineage, error) {
	if m.getColumnLineageFunc != nil {
		return m.getColumnLineageFunc(ctx, urn)
	}
	return &types.ColumnLineage{DatasetURN: urn}, nil
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

func (m *mockClient) UpdateDescription(ctx context.Context, urn, description string) error {
	if m.updateDescriptionFunc != nil {
		return m.updateDescriptionFunc(ctx, urn, description)
	}
	return nil
}

func (m *mockClient) AddTag(ctx context.Context, urn, tagURN string) error {
	if m.addTagFunc != nil {
		return m.addTagFunc(ctx, urn, tagURN)
	}
	return nil
}

func (m *mockClient) RemoveTag(ctx context.Context, urn, tagURN string) error {
	if m.removeTagFunc != nil {
		return m.removeTagFunc(ctx, urn, tagURN)
	}
	return nil
}

func (m *mockClient) AddGlossaryTerm(ctx context.Context, urn, termURN string) error {
	if m.addGlossaryTermFunc != nil {
		return m.addGlossaryTermFunc(ctx, urn, termURN)
	}
	return nil
}

func (m *mockClient) RemoveGlossaryTerm(ctx context.Context, urn, termURN string) error {
	if m.removeGlossaryTermFunc != nil {
		return m.removeGlossaryTermFunc(ctx, urn, termURN)
	}
	return nil
}

func (m *mockClient) AddLink(ctx context.Context, urn, linkURL, description string) error {
	if m.addLinkFunc != nil {
		return m.addLinkFunc(ctx, urn, linkURL, description)
	}
	return nil
}

func (m *mockClient) RemoveLink(ctx context.Context, urn, linkURL string) error {
	if m.removeLinkFunc != nil {
		return m.removeLinkFunc(ctx, urn, linkURL)
	}
	return nil
}

func (m *mockClient) GetStructuredProperties(ctx context.Context, urn string) ([]types.StructuredPropertyValue, error) {
	if m.getStructuredPropertiesFunc != nil {
		return m.getStructuredPropertiesFunc(ctx, urn)
	}
	return nil, nil
}

func (m *mockClient) ListStructuredPropertyDefinitions(ctx context.Context) ([]types.StructuredPropertyDefinition, error) {
	if m.listStructuredPropertyDefinitionsFunc != nil {
		return m.listStructuredPropertyDefinitionsFunc(ctx)
	}
	return nil, nil
}

func (m *mockClient) UpsertStructuredProperties(ctx context.Context, urn string, properties []types.StructuredPropertyInput) error {
	if m.upsertStructuredPropertiesFunc != nil {
		return m.upsertStructuredPropertiesFunc(ctx, urn, properties)
	}
	return nil
}

func (m *mockClient) RemoveStructuredProperties(ctx context.Context, urn string, propertyURNs []string) error {
	if m.removeStructuredPropertiesFunc != nil {
		return m.removeStructuredPropertiesFunc(ctx, urn, propertyURNs)
	}
	return nil
}

func (m *mockClient) GetIncidents(ctx context.Context, urn string) (*types.IncidentResult, error) {
	if m.getIncidentsFunc != nil {
		return m.getIncidentsFunc(ctx, urn)
	}
	return &types.IncidentResult{}, nil
}

func (m *mockClient) RaiseIncident(ctx context.Context, input types.RaiseIncidentInput) (string, error) {
	if m.raiseIncidentFunc != nil {
		return m.raiseIncidentFunc(ctx, input)
	}
	return "", nil
}

func (m *mockClient) UpdateIncidentStatus(ctx context.Context, incidentURN, state, message string) error {
	if m.updateIncidentStatusFunc != nil {
		return m.updateIncidentStatusFunc(ctx, incidentURN, state, message)
	}
	return nil
}

func (m *mockClient) ResolveIncident(ctx context.Context, incidentURN, message string) error {
	if m.resolveIncidentFunc != nil {
		return m.resolveIncidentFunc(ctx, incidentURN, message)
	}
	return nil
}

func (m *mockClient) GetDataContract(ctx context.Context, datasetURN string) (*types.DataContract, error) {
	if m.getDataContractFunc != nil {
		return m.getDataContractFunc(ctx, datasetURN)
	}
	return nil, nil
}

func (m *mockClient) SemanticSearch(ctx context.Context, query string, opts ...client.SearchOption) (*types.SearchResult, error) {
	if m.semanticSearchFunc != nil {
		return m.semanticSearchFunc(ctx, query, opts...)
	}
	return &types.SearchResult{}, nil
}

func (m *mockClient) GetDocument(ctx context.Context, urn string) (*types.Document, error) {
	if m.getDocumentFunc != nil {
		return m.getDocumentFunc(ctx, urn)
	}
	return &types.Document{URN: urn}, nil
}

func (m *mockClient) GetRelatedDocuments(ctx context.Context, urn string) ([]types.Document, error) {
	if m.getRelatedDocsFunc != nil {
		return m.getRelatedDocsFunc(ctx, urn)
	}
	return nil, nil
}

func (m *mockClient) GetContextDocuments(ctx context.Context, urn string) ([]types.ContextDocument, error) {
	if m.getContextDocsFunc != nil {
		return m.getContextDocsFunc(ctx, urn)
	}
	return nil, nil
}

func (m *mockClient) UpsertContextDocument(
	ctx context.Context, entityURN string, doc types.ContextDocumentInput,
) (*types.ContextDocument, error) {
	if m.upsertContextDocFunc != nil {
		return m.upsertContextDocFunc(ctx, entityURN, doc)
	}
	return &types.ContextDocument{ID: "new", Title: doc.Title}, nil
}

func (m *mockClient) DeleteContextDocument(ctx context.Context, documentID string) error {
	if m.deleteContextDocFunc != nil {
		return m.deleteContextDocFunc(ctx, documentID)
	}
	return nil
}

func (m *mockClient) UpdateColumnDescription(ctx context.Context, urn, fieldPath, description string) error {
	if m.updateColumnDescriptionFunc != nil {
		return m.updateColumnDescriptionFunc(ctx, urn, fieldPath, description)
	}
	return nil
}

func (m *mockClient) CreateQuery(ctx context.Context, input client.CreateQueryInput) (*types.Query, error) {
	if m.createQueryFunc != nil {
		return m.createQueryFunc(ctx, input)
	}
	return &types.Query{URN: "urn:li:query:new"}, nil
}

func (m *mockClient) UpdateQuery(ctx context.Context, input client.UpdateQueryInput) (*types.Query, error) {
	if m.updateQueryFunc != nil {
		return m.updateQueryFunc(ctx, input)
	}
	return &types.Query{URN: input.URN}, nil
}

func (m *mockClient) DeleteQuery(ctx context.Context, urn string) error {
	if m.deleteQueryFunc != nil {
		return m.deleteQueryFunc(ctx, urn)
	}
	return nil
}

func (m *mockClient) CreateTag(ctx context.Context, name, description string) (string, error) {
	if m.createTagFunc != nil {
		return m.createTagFunc(ctx, name, description)
	}
	return "urn:li:tag:" + name, nil
}

func (m *mockClient) CreateDomain(ctx context.Context, name, description string) (string, error) {
	if m.createDomainFunc != nil {
		return m.createDomainFunc(ctx, name, description)
	}
	return "urn:li:domain:" + name, nil
}

func (m *mockClient) CreateGlossaryTerm(ctx context.Context, name, description, parentNode string) (string, error) {
	if m.createGlossaryTermFunc != nil {
		return m.createGlossaryTermFunc(ctx, name, description, parentNode)
	}
	return "urn:li:glossaryTerm:" + name, nil
}

func (m *mockClient) CreateDataProduct(ctx context.Context, name, description, domainURN string) (string, error) {
	if m.createDataProductFunc != nil {
		return m.createDataProductFunc(ctx, name, description, domainURN)
	}
	return "urn:li:dataProduct:" + name, nil
}

func (m *mockClient) CreateDocument(ctx context.Context, input types.CreateDocumentInput) (string, error) {
	if m.createDocumentFunc != nil {
		return m.createDocumentFunc(ctx, input)
	}
	return "urn:li:document:new", nil
}

func (m *mockClient) CreateApplication(ctx context.Context, name, description string) (string, error) {
	if m.createApplicationFunc != nil {
		return m.createApplicationFunc(ctx, name, description)
	}
	return "urn:li:application:" + name, nil
}

func (m *mockClient) CreateStructuredProperty(ctx context.Context, input types.CreateStructuredPropertyInput) (string, error) {
	if m.createStructuredPropertyFunc != nil {
		return m.createStructuredPropertyFunc(ctx, input)
	}
	return "urn:li:structuredProperty:" + input.QualifiedName, nil
}

func (m *mockClient) UpsertDataContract(ctx context.Context, input types.UpsertDataContractInput) (string, error) {
	if m.upsertDataContractFunc != nil {
		return m.upsertDataContractFunc(ctx, input)
	}
	return "urn:li:dataContract:new", nil
}

func (m *mockClient) AddOwner(ctx context.Context, urn, ownerURN, ownershipType string) error {
	if m.addOwnerFunc != nil {
		return m.addOwnerFunc(ctx, urn, ownerURN, ownershipType)
	}
	return nil
}

func (m *mockClient) RemoveOwner(ctx context.Context, urn, ownerURN string) error {
	if m.removeOwnerFunc != nil {
		return m.removeOwnerFunc(ctx, urn, ownerURN)
	}
	return nil
}

func (m *mockClient) SetDomain(ctx context.Context, entityURN, domainURN string) error {
	if m.setDomainFunc != nil {
		return m.setDomainFunc(ctx, entityURN, domainURN)
	}
	return nil
}

func (m *mockClient) UnsetDomain(ctx context.Context, entityURN string) error {
	if m.unsetDomainFunc != nil {
		return m.unsetDomainFunc(ctx, entityURN)
	}
	return nil
}

func (m *mockClient) UpdateIncident(ctx context.Context, urn string, input types.UpdateIncidentInput) error {
	if m.updateIncidentFunc != nil {
		return m.updateIncidentFunc(ctx, urn, input)
	}
	return nil
}

func (m *mockClient) UpdateStructuredProperty(ctx context.Context, urn string, input types.UpdateStructuredPropertyInput) error {
	if m.updateStructuredPropertyFunc != nil {
		return m.updateStructuredPropertyFunc(ctx, urn, input)
	}
	return nil
}

func (m *mockClient) UpdateDocumentContents(ctx context.Context, urn, title, text string) error {
	if m.updateDocumentContentsFunc != nil {
		return m.updateDocumentContentsFunc(ctx, urn, title, text)
	}
	return nil
}

func (m *mockClient) UpdateDocumentStatus(ctx context.Context, urn, status string) error {
	if m.updateDocumentStatusFunc != nil {
		return m.updateDocumentStatusFunc(ctx, urn, status)
	}
	return nil
}

func (m *mockClient) UpdateDocumentRelatedEntities(ctx context.Context, urn string, entityURNs []string) error {
	if m.updateDocRelatedFunc != nil {
		return m.updateDocRelatedFunc(ctx, urn, entityURNs)
	}
	return nil
}

func (m *mockClient) UpdateDocumentSubType(ctx context.Context, urn, subType string) error {
	if m.updateDocSubTypeFunc != nil {
		return m.updateDocSubTypeFunc(ctx, urn, subType)
	}
	return nil
}

func (m *mockClient) DeleteTag(ctx context.Context, urn string) error {
	if m.deleteTagFunc != nil {
		return m.deleteTagFunc(ctx, urn)
	}
	return nil
}

func (m *mockClient) DeleteDomain(ctx context.Context, urn string) error {
	if m.deleteDomainFunc != nil {
		return m.deleteDomainFunc(ctx, urn)
	}
	return nil
}

func (m *mockClient) DeleteGlossaryEntity(ctx context.Context, urn string) error {
	if m.deleteGlossaryEntityFunc != nil {
		return m.deleteGlossaryEntityFunc(ctx, urn)
	}
	return nil
}

func (m *mockClient) DeleteDataProduct(ctx context.Context, urn string) error {
	if m.deleteDataProductFunc != nil {
		return m.deleteDataProductFunc(ctx, urn)
	}
	return nil
}

func (m *mockClient) DeleteApplication(ctx context.Context, urn string) error {
	if m.deleteApplicationFunc != nil {
		return m.deleteApplicationFunc(ctx, urn)
	}
	return nil
}

func (m *mockClient) DeleteDocument(ctx context.Context, urn string) error {
	if m.deleteDocumentFunc != nil {
		return m.deleteDocumentFunc(ctx, urn)
	}
	return nil
}

func (m *mockClient) DeleteStructuredProperty(ctx context.Context, urn string) error {
	if m.deleteStructuredPropertyFunc != nil {
		return m.deleteStructuredPropertyFunc(ctx, urn)
	}
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

	// Verify write tools are NOT registered when WriteEnabled is false
	for _, name := range WriteTools() {
		if toolkit.registeredTools[name] {
			t.Errorf("RegisterAll() should NOT register write tool %s when WriteEnabled is false", name)
		}
	}
}

func TestToolkitRegisterAll_WriteEnabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.RegisterAll(server)

	// Verify all read tools are registered
	for _, name := range AllTools() {
		if !toolkit.registeredTools[name] {
			t.Errorf("RegisterAll() should register read tool %s", name)
		}
	}

	// Verify all write tools are registered
	for _, name := range WriteTools() {
		if !toolkit.registeredTools[name] {
			t.Errorf("RegisterAll() should register write tool %s when WriteEnabled is true", name)
		}
	}
}

func TestToolkitGetWriteClient_Disabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig()) // WriteEnabled = false

	_, err := toolkit.getWriteClient("")
	if err == nil {
		t.Fatal("expected error when write is disabled")
	}
	if !errors.Is(err, client.ErrWriteDisabled) {
		t.Errorf("expected ErrWriteDisabled, got: %v", err)
	}
}

func TestToolkitGetWriteClient_Enabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	c, err := toolkit.getWriteClient("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c != mock {
		t.Error("expected mock client")
	}
}

func TestToolkitGetWriteClient_PerConnectionDisabled(t *testing.T) {
	writeDisabled := false
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]multiserver.ConnectionConfig{
			"staging": {
				URL:          "https://staging.datahub.example.com",
				Token:        "staging-token",
				WriteEnabled: &writeDisabled,
			},
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	// Global write enabled, but staging has write_enabled=false
	toolkit := NewToolkitWithManager(mgr, Config{WriteEnabled: true})

	// Default connection should work
	_, err := toolkit.getWriteClient("")
	if err != nil {
		t.Fatalf("expected write to succeed on default connection: %v", err)
	}

	// Staging connection should be blocked
	_, err = toolkit.getWriteClient("staging")
	if err == nil {
		t.Fatal("expected error when writing to connection with write_enabled=false")
	}
	if !errors.Is(err, client.ErrWriteDisabled) {
		t.Errorf("expected ErrWriteDisabled, got: %v", err)
	}
}

func TestToolkitGetWriteClient_PerConnectionEnabled(t *testing.T) {
	writeEnabled := true
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]multiserver.ConnectionConfig{
			"staging": {
				URL:          "https://staging.datahub.example.com",
				Token:        "staging-token",
				WriteEnabled: &writeEnabled,
			},
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	// Global write enabled, staging also explicitly enabled
	toolkit := NewToolkitWithManager(mgr, Config{WriteEnabled: true})

	// Both should work
	_, err := toolkit.getWriteClient("")
	if err != nil {
		t.Fatalf("expected write to succeed on default: %v", err)
	}
	_, err = toolkit.getWriteClient("staging")
	if err != nil {
		t.Fatalf("expected write to succeed on staging: %v", err)
	}
}

func TestToolkitGetWriteClient_PerConnectionOverridesGlobal(t *testing.T) {
	writeEnabled := true
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]multiserver.ConnectionConfig{
			"staging": {
				URL:          "https://staging.datahub.example.com",
				Token:        "staging-token",
				WriteEnabled: &writeEnabled,
			},
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	// Global write DISABLED, but staging has explicit write_enabled=true
	toolkit := NewToolkitWithManager(mgr, Config{WriteEnabled: false})

	// Default connection should be blocked (inherits global false)
	_, err := toolkit.getWriteClient("")
	if err == nil {
		t.Fatal("expected error for default connection with global write disabled")
	}
	if !errors.Is(err, client.ErrWriteDisabled) {
		t.Errorf("expected ErrWriteDisabled, got: %v", err)
	}

	// Staging should work (explicit true overrides global false)
	_, err = toolkit.getWriteClient("staging")
	if err != nil {
		t.Fatalf("expected write to succeed on staging with explicit write_enabled=true: %v", err)
	}
}

func TestWriteTools(t *testing.T) {
	wt := WriteTools()
	if len(wt) != 3 {
		t.Errorf("expected 3 write tools, got %d", len(wt))
	}

	expected := map[ToolName]bool{
		ToolCreate: true,
		ToolUpdate: true,
		ToolDelete: true,
	}
	for _, name := range wt {
		if !expected[name] {
			t.Errorf("unexpected write tool: %s", name)
		}
	}
}

func TestAllToolsUnchanged(t *testing.T) {
	at := AllTools()
	if len(at) != 9 {
		t.Errorf("AllTools() should return 9 tools, got %d", len(at))
	}

	// Verify no write tools in AllTools
	writeSet := make(map[ToolName]bool)
	for _, name := range WriteTools() {
		writeSet[name] = true
	}
	for _, name := range at {
		if writeSet[name] {
			t.Errorf("AllTools() should not contain write tool %s", name)
		}
	}
}

func TestEnrichEntityWith14xFeatures(t *testing.T) {
	mock := &mockClient{}
	mock.getEntityFunc = func(_ context.Context, _ string) (*types.Entity, error) {
		return &types.Entity{
			URN:  "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)",
			Type: "DATASET",
			Name: "users",
		}, nil
	}
	mock.getStructuredPropertiesFunc = func(_ context.Context, _ string) ([]types.StructuredPropertyValue, error) {
		return []types.StructuredPropertyValue{
			{PropertyURN: "urn:li:structuredProperty:retention", Values: []any{float64(30)}},
		}, nil
	}
	mock.getIncidentsFunc = func(_ context.Context, _ string) (*types.IncidentResult, error) {
		return &types.IncidentResult{
			Total: 1,
			Incidents: []types.Incident{
				{URN: "urn:li:incident:1", Type: "OPERATIONAL", Title: "Pipeline down", State: "ACTIVE"},
			},
		}, nil
	}
	mock.getDataContractFunc = func(_ context.Context, _ string) (*types.DataContract, error) {
		return &types.DataContract{Status: "PASSING"}, nil
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	entity := &types.Entity{
		URN:  "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)",
		Type: "DATASET",
		Name: "users",
	}

	toolkit.enrichEntityWith14xFeatures(t.Context(), mock, entity)

	if len(entity.StructuredProperties) != 1 {
		t.Errorf("StructuredProperties count = %d, want 1", len(entity.StructuredProperties))
	}
	if entity.StructuredProperties[0].PropertyURN != "urn:li:structuredProperty:retention" {
		t.Errorf("StructuredProperties[0].PropertyURN = %q", entity.StructuredProperties[0].PropertyURN)
	}
	if entity.ActiveIncidents == nil || entity.ActiveIncidents.Total != 1 {
		t.Errorf("ActiveIncidents = %+v, want total=1", entity.ActiveIncidents)
	}
	if entity.DataContract == nil || entity.DataContract.Status != "PASSING" {
		t.Errorf("DataContract = %+v, want status=PASSING", entity.DataContract)
	}
}

func TestEnrichEntityWith14xFeatures_NonDataset_SkipsContract(t *testing.T) {
	mock := &mockClient{}
	mock.getStructuredPropertiesFunc = func(_ context.Context, _ string) ([]types.StructuredPropertyValue, error) {
		return nil, nil
	}
	mock.getIncidentsFunc = func(_ context.Context, _ string) (*types.IncidentResult, error) {
		return &types.IncidentResult{}, nil
	}
	mock.getDataContractFunc = func(_ context.Context, _ string) (*types.DataContract, error) {
		t.Error("GetDataContract should not be called for non-dataset entities")
		return nil, nil
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	entity := &types.Entity{
		URN:  "urn:li:dashboard:(urn:li:dataPlatform:looker,dashboard1)",
		Type: "DASHBOARD",
		Name: "dashboard1",
	}

	toolkit.enrichEntityWith14xFeatures(t.Context(), mock, entity)

	if entity.DataContract != nil {
		t.Error("DataContract should be nil for non-dataset entities")
	}
}

func TestSearchModeRouting(t *testing.T) {
	tests := []struct {
		name           string
		mode           string
		wantSemantic   bool
		wantKeyword    bool
		wantErr        bool
		wantErrContain string
	}{
		{
			name:        "default mode uses keyword",
			mode:        "",
			wantKeyword: true,
		},
		{
			name:        "keyword mode",
			mode:        "keyword",
			wantKeyword: true,
		},
		{
			name:         "semantic mode",
			mode:         "semantic",
			wantSemantic: true,
		},
		{
			name:           "invalid mode",
			mode:           "vector",
			wantErr:        true,
			wantErrContain: "invalid mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywordCalled := false
			semanticCalled := false

			mock := &mockClient{}
			mock.searchFunc = func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
				keywordCalled = true
				return &types.SearchResult{}, nil
			}
			mock.semanticSearchFunc = func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
				semanticCalled = true
				return &types.SearchResult{}, nil
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			input := SearchInput{Query: "test", Mode: tt.mode}
			result, _, _ := toolkit.handleSearch(t.Context(), nil, input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error result")
				}
				// Check error message content
				if tt.wantErrContain != "" {
					for _, c := range result.Content {
						if tc, ok := c.(*mcp.TextContent); ok {
							if !strings.Contains(tc.Text, tt.wantErrContain) {
								t.Errorf("error %q should contain %q", tc.Text, tt.wantErrContain)
							}
						}
					}
				}
				return
			}

			if tt.wantKeyword && !keywordCalled {
				t.Error("expected keyword Search to be called")
			}
			if tt.wantSemantic && !semanticCalled {
				t.Error("expected SemanticSearch to be called")
			}
			if tt.wantKeyword && semanticCalled {
				t.Error("SemanticSearch should not be called in keyword mode")
			}
			if tt.wantSemantic && keywordCalled {
				t.Error("keyword Search should not be called in semantic mode")
			}
		})
	}
}
