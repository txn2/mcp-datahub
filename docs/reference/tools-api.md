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
