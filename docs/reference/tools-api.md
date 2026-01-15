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

---

## Client Package

The `client` package provides the DataHub GraphQL client.

### NewFromEnv

Creates a client from environment variables.

```go
func NewFromEnv() (*Client, error)
```

**Environment Variables:**

| Variable | Required | Description |
|----------|----------|-------------|
| `DATAHUB_URL` | Yes | DataHub GraphQL endpoint URL |
| `DATAHUB_TOKEN` | Yes | Authentication token |
| `DATAHUB_CONNECTION_NAME` | No | Named connection identifier |

### New

Creates a client with explicit configuration.

```go
func New(cfg Config) (*Client, error)
```

**Config Fields:**

```go
type Config struct {
    URL              string        // DataHub GraphQL endpoint
    Token            string        // Authentication token
    ConnectionName   string        // Optional connection name
    Timeout          time.Duration // Request timeout (default: 30s)
    MaxRetries       int           // Max retry attempts (default: 3)
    RetryBackoff     time.Duration // Initial backoff duration (default: 1s)
}
```

### Client Methods

| Method | Description |
|--------|-------------|
| `Search(ctx, query, entityType, limit, offset)` | Search for entities |
| `GetEntity(ctx, urn)` | Get entity by URN |
| `GetSchema(ctx, urn)` | Get dataset schema |
| `GetLineage(ctx, urn, direction, depth)` | Get entity lineage |
| `GetQueries(ctx, urn)` | Get associated queries |
| `GetGlossaryTerm(ctx, urn)` | Get glossary term details |
| `ListTags(ctx, filter)` | List tags |
| `ListDomains(ctx)` | List domains |
| `ListDataProducts(ctx)` | List data products |
| `GetDataProduct(ctx, urn)` | Get data product details |
| `Close()` | Close the client |

---

## Types Package

The `types` package contains domain types returned by the client and tools.

### Entity

Represents a DataHub entity.

```go
type Entity struct {
    URN         string            `json:"urn"`
    Type        string            `json:"type"`
    Name        string            `json:"name"`
    Platform    string            `json:"platform,omitempty"`
    Description string            `json:"description,omitempty"`
    Owners      []Owner           `json:"owners,omitempty"`
    Tags        []Tag             `json:"tags,omitempty"`
    Terms       []GlossaryTerm    `json:"glossaryTerms,omitempty"`
    Domain      *Domain           `json:"domain,omitempty"`
    Properties  map[string]any    `json:"properties,omitempty"`
}
```

### SchemaField

Represents a field in a dataset schema.

```go
type SchemaField struct {
    FieldPath     string         `json:"fieldPath"`
    Type          string         `json:"type"`
    NativeType    string         `json:"nativeType,omitempty"`
    Description   string         `json:"description,omitempty"`
    Nullable      bool           `json:"nullable"`
    IsPrimaryKey  bool           `json:"isPrimaryKey,omitempty"`
    GlossaryTerms []GlossaryTerm `json:"glossaryTerms,omitempty"`
}
```

### LineageResult

Represents lineage query results.

```go
type LineageResult struct {
    URN        string          `json:"urn"`
    Upstream   []LineageEntity `json:"upstream,omitempty"`
    Downstream []LineageEntity `json:"downstream,omitempty"`
}

type LineageEntity struct {
    URN      string `json:"urn"`
    Name     string `json:"name"`
    Type     string `json:"type"`
    Platform string `json:"platform,omitempty"`
    Degree   int    `json:"degree"`
}
```

### Owner

Represents an entity owner.

```go
type Owner struct {
    URN  string `json:"urn"`
    Name string `json:"name"`
    Type string `json:"type"`
}
```

### Tag

Represents a tag.

```go
type Tag struct {
    URN         string `json:"urn"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
}
```

### Domain

Represents a data domain.

```go
type Domain struct {
    URN         string `json:"urn"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    EntityCount int    `json:"entityCount,omitempty"`
}
```

### DataProduct

Represents a data product.

```go
type DataProduct struct {
    URN         string            `json:"urn"`
    Name        string            `json:"name"`
    Description string            `json:"description,omitempty"`
    Domain      *Domain           `json:"domain,omitempty"`
    Owners      []Owner           `json:"owners,omitempty"`
    Assets      []Entity          `json:"assets,omitempty"`
    Properties  map[string]any    `json:"properties,omitempty"`
}
```

---

## Thread Safety

All components in mcp-datahub are designed for concurrent use:

### Client Thread Safety

The `Client` is safe for concurrent use by multiple goroutines:

- Uses connection pooling with proper synchronization
- HTTP client is shared across requests
- No shared mutable state between requests

```go
// Safe: multiple goroutines using same client
client, _ := datahubclient.NewFromEnv()
defer client.Close()

var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        result, _ := client.Search(ctx, "customer", "", 10, 0)
        // Process result
    }()
}
wg.Wait()
```

### Toolkit Thread Safety

The `Toolkit` handles concurrent tool calls:

- Tool handlers are stateless
- Middleware must be either stateless or properly synchronized
- Per-request state passed through context

```go
// Safe: concurrent tool registration and execution
toolkit := tools.NewToolkit(client)
toolkit.RegisterAll(server)
// Server handles concurrent requests automatically
```

### Middleware Thread Safety Requirements

When implementing custom middleware:

| Guideline | Description |
|-----------|-------------|
| Avoid shared state | Do not store request-specific data in middleware structs |
| Use context | Pass request-scoped data via context.Context |
| Synchronize if needed | Use sync.Mutex for shared counters or caches |
| Prefer immutable | Design middleware to be stateless when possible |

```go
// Thread-safe rate limiter example
type RateLimiter struct {
    mu       sync.Mutex
    requests map[string]int
}

func (r *RateLimiter) Before(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // Safe access to shared state
    return ctx, nil
}
```

---

## Performance Characteristics

### Request Latency

Typical latency ranges for tool operations:

| Tool | Typical Latency | Factors |
|------|-----------------|---------|
| `datahub_search` | 50-200ms | Query complexity, result count |
| `datahub_get_entity` | 20-100ms | Entity type, aspect count |
| `datahub_get_schema` | 30-150ms | Field count |
| `datahub_get_lineage` | 100-500ms | Depth, graph size |
| `datahub_list_*` | 50-200ms | Result count |

### Connection Pooling

The client uses HTTP connection pooling:

```go
// Default transport settings
Transport: &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
```

### Retry Behavior

Failed requests are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 1 second |
| 3 | 2 seconds |
| 4 | 4 seconds |

Retries only occur for:

- Network timeouts
- HTTP 500, 502, 503, 504 errors
- Connection refused errors

Not retried:

- HTTP 400, 401, 403, 404 errors
- Context cancellation
- Invalid request errors

---

## Memory Considerations

### Response Size Limits

The client limits response sizes to prevent memory issues:

| Limit | Default | Description |
|-------|---------|-------------|
| Max response body | 10MB | Maximum GraphQL response size |
| Max entities | 1000 | Maximum search results per request |
| Max lineage depth | 5 | Maximum traversal depth |

### Streaming Large Results

For large result sets, use pagination:

```go
// Paginate search results
offset := 0
limit := 100
for {
    result, err := toolkit.Search(ctx, SearchInput{
        Query:  "customer",
        Limit:  limit,
        Offset: offset,
    })
    if err != nil {
        break
    }
    if len(result.Entities) == 0 {
        break
    }
    // Process batch
    offset += limit
}
```

### Memory-Efficient Patterns

| Pattern | Description |
|---------|-------------|
| Process in batches | Use pagination for large result sets |
| Close clients | Call `Close()` when done to release resources |
| Limit lineage depth | Use depth=2 or 3 for most use cases |
| Filter by entity type | Reduce result count with type filters |

---

## Error Handling

### Error Types

The library uses typed errors for specific conditions:

```go
import "github.com/txn2/mcp-datahub/pkg/client"

// Check error types
switch {
case errors.Is(err, client.ErrUnauthorized):
    // Handle auth error
case errors.Is(err, client.ErrNotFound):
    // Handle not found
case errors.Is(err, client.ErrRateLimited):
    // Handle rate limiting
case errors.Is(err, client.ErrTimeout):
    // Handle timeout
default:
    // Handle other errors
}
```

### Error Wrapping

All errors include context for debugging:

```go
// Errors include operation context
// Example: "search failed: graphql error: unauthorized"
```

### Tool Error Responses

Tools return structured error responses:

```go
// Error response format
{
    "error": true,
    "message": "Entity not found: urn:li:dataset:...",
    "code": "NOT_FOUND"
}
```

---

## Context Usage

### Standard Context Values

The library recognizes these context values:

| Key | Type | Description |
|-----|------|-------------|
| `auth_token` | string | Authentication token for requests |
| `user_id` | string | User identifier for audit logging |
| `tenant_id` | string | Tenant identifier for multi-tenancy |
| `request_id` | string | Request correlation ID |

### Setting Context Values

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "user_id", "user@example.com")
ctx = context.WithValue(ctx, "request_id", uuid.New().String())

result, err := toolkit.Search(ctx, input)
```

### Context Cancellation

All operations respect context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := client.Search(ctx, "query", "", 10, 0)
if errors.Is(err, context.DeadlineExceeded) {
    // Handle timeout
}
```

---

## Related Topics

- [Architecture](../library/architecture.md): System design and component diagrams
- [Composability](../library/composability.md): Combining toolkits and middleware patterns
- [Testing Guide](../guides/testing.md): Testing strategies for integrations
