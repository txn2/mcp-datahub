# Tools API Reference

Complete API reference for the tools package.

## Toolkit

### NewToolkit

Creates a new toolkit instance.

```go
func NewToolkit(client DataHubClient, opts ...Option) *Toolkit
```

**Parameters:**

- `client`: A DataHub client implementing the `DataHubClient` interface
- `opts`: Optional configuration options

**Example:**

```go
toolkit := tools.NewToolkit(datahubClient)
```

### RegisterAll

Registers all available tools with the MCP server.

```go
func (t *Toolkit) RegisterAll(server *mcp.Server)
```

### Register

Registers specific tools with the MCP server.

```go
func (t *Toolkit) Register(server *mcp.Server, names ...ToolName)
```

**Example:**

```go
toolkit.Register(server, tools.ToolSearch, tools.ToolGetEntity)
```

### RegisterWith

Registers a tool with per-tool options.

```go
func (t *Toolkit) RegisterWith(server *mcp.Server, name ToolName, opts ...PerToolOption)
```

## Options

### WithMiddleware

Adds global middleware to all tools.

```go
func WithMiddleware(mw ToolMiddleware) Option
```

### WithToolMiddleware

Adds middleware to a specific tool.

```go
func WithToolMiddleware(name ToolName, mw ToolMiddleware) Option
```

### WithQueryProvider

Injects a query execution context provider for bidirectional integration.

```go
func WithQueryProvider(p integration.QueryProvider) Option
```

When configured, enriches tool responses with query execution context (table resolution, availability, examples).

### WithURNResolver

Maps external IDs to DataHub URNs before tool execution.

```go
func WithURNResolver(r integration.URNResolver) Option
```

### WithAccessFilter

Adds access control filtering before and after tool execution.

```go
func WithAccessFilter(f integration.AccessFilter) Option
```

### WithAuditLogger

Logs all tool invocations for audit purposes.

```go
func WithAuditLogger(l integration.AuditLogger, getUserID func(context.Context) string) Option
```

### WithMetadataEnricher

Adds custom metadata to entity responses.

```go
func WithMetadataEnricher(e integration.MetadataEnricher) Option
```

## Tool Names

Available tool name constants:

```go
const (
    ToolSearch           ToolName = "datahub_search"
    ToolGetEntity        ToolName = "datahub_get_entity"
    ToolGetSchema        ToolName = "datahub_get_schema"
    ToolGetLineage       ToolName = "datahub_get_lineage"
    ToolGetQueries       ToolName = "datahub_get_queries"
    ToolGetGlossaryTerm  ToolName = "datahub_get_glossary_term"
    ToolListTags         ToolName = "datahub_list_tags"
    ToolListDomains      ToolName = "datahub_list_domains"
    ToolListDataProducts ToolName = "datahub_list_data_products"
    ToolGetDataProduct   ToolName = "datahub_get_data_product"
)
```

## Middleware

### ToolMiddleware Interface

```go
type ToolMiddleware interface {
    Before(ctx context.Context, tc *ToolContext) (context.Context, error)
    After(ctx context.Context, tc *ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error)
}
```

### BeforeFunc

Creates middleware that runs before tool execution.

```go
func BeforeFunc(fn func(ctx context.Context, tc *ToolContext) (context.Context, error)) ToolMiddleware
```

### AfterFunc

Creates middleware that runs after tool execution.

```go
func AfterFunc(fn func(ctx context.Context, tc *ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error)) ToolMiddleware
```

## Helper Functions

### TextResult

Creates a text result.

```go
func TextResult(text string) *mcp.CallToolResult
```

### JSONResult

Creates a JSON result.

```go
func JSONResult(v any) (*mcp.CallToolResult, error)
```

### ErrorResult

Creates an error result.

```go
func ErrorResult(msg string) *mcp.CallToolResult
```

## Integration Package

The `integration` package provides interfaces for enterprise integration.

### QueryProvider Interface

Enables query engines to inject execution context into DataHub tools.

```go
type QueryProvider interface {
    Name() string
    ResolveTable(ctx context.Context, urn string) (*TableIdentifier, error)
    GetTableAvailability(ctx context.Context, urn string) (*TableAvailability, error)
    GetQueryExamples(ctx context.Context, urn string) ([]QueryExample, error)
    GetExecutionContext(ctx context.Context, urns []string) (*ExecutionContext, error)
    Close() error
}
```

### TableIdentifier

Represents a fully-qualified table reference.

```go
type TableIdentifier struct {
    Connection string `json:"connection,omitempty"`  // Named connection (optional)
    Catalog    string `json:"catalog"`               // Catalog/database name
    Schema     string `json:"schema"`                // Schema name
    Table      string `json:"table"`                 // Table name
}

func (t TableIdentifier) String() string  // Returns "catalog.schema.table" or "conn:catalog.schema.table"
```

### TableAvailability

Indicates whether a DataHub entity is queryable.

```go
type TableAvailability struct {
    Available   bool             `json:"available"`
    Table       *TableIdentifier `json:"table,omitempty"`
    Connection  string           `json:"connection,omitempty"`
    Error       string           `json:"error,omitempty"`
    LastChecked time.Time        `json:"last_checked,omitempty"`
    RowCount    *int64           `json:"row_count,omitempty"`
    LastUpdated *time.Time       `json:"last_updated,omitempty"`
}
```

### QueryExample

Represents a sample SQL query for a DataHub entity.

```go
type QueryExample struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    SQL         string `json:"sql"`
    Category    string `json:"category,omitempty"`   // "sample", "aggregation", "join", etc.
    Source      string `json:"source,omitempty"`     // "generated", "history", "template"
}
```

### ExecutionContext

Provides query execution context for lineage bridging.

```go
type ExecutionContext struct {
    Tables      map[string]*TableIdentifier `json:"tables,omitempty"`
    Connections []string                    `json:"connections,omitempty"`
    Queries     []ExecutionQuery            `json:"queries,omitempty"`
    Source      string                      `json:"source,omitempty"`
}
```

### URNResolver Interface

Maps external IDs to DataHub URNs.

```go
type URNResolver interface {
    ResolveToDataHubURN(ctx context.Context, externalID string) (string, error)
}
```

### AccessFilter Interface

Controls access to entities.

```go
type AccessFilter interface {
    CanAccess(ctx context.Context, urn string) (bool, error)
    FilterURNs(ctx context.Context, urns []string) ([]string, error)
}
```

### AuditLogger Interface

Logs tool invocations.

```go
type AuditLogger interface {
    LogToolCall(ctx context.Context, tool string, params map[string]any, userID string) error
}
```

### MetadataEnricher Interface

Adds custom metadata to entity responses.

```go
type MetadataEnricher interface {
    EnrichEntity(ctx context.Context, urn string, data map[string]any) (map[string]any, error)
}
```

### NoOpQueryProvider

A default no-op implementation of QueryProvider.

```go
var _ QueryProvider = (*NoOpQueryProvider)(nil)

type NoOpQueryProvider struct{}
```

### QueryProviderFunc

Function-based QueryProvider implementation for simple cases.

```go
type QueryProviderFunc struct {
    NameFn                 func() string
    ResolveTableFn         func(context.Context, string) (*TableIdentifier, error)
    GetTableAvailabilityFn func(context.Context, string) (*TableAvailability, error)
    GetQueryExamplesFn     func(context.Context, string) ([]QueryExample, error)
    GetExecutionContextFn  func(context.Context, []string) (*ExecutionContext, error)
    CloseFn                func() error
}
```
