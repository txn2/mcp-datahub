# Frequently Asked Questions

## General

### What is mcp-datahub?

mcp-datahub is an MCP (Model Context Protocol) server that connects AI assistants like Claude to DataHub metadata catalogs. It allows AI assistants to search datasets, explore schemas, trace lineage, and access business context.

### What is the difference between this and the official DataHub MCP server?

This project serves a different purpose:

| Feature | mcp-datahub | Official DataHub MCP |
|---------|-------------|---------------------|
| Primary use | Composable Go library | Standalone server |
| Custom auth | Full middleware support | Limited |
| Multi-tool composition | Designed for it | Not supported |
| Tenant isolation | Built-in hooks | Not available |
| Audit logging | Interface provided | Not available |

Use mcp-datahub when you need to build enterprise MCP servers with custom authentication, combine multiple data tools, or add audit logging for compliance.

### Can I use this in production?

Yes. mcp-datahub is designed for production use with:

- SLSA Level 3 provenance for supply chain security
- Cosign-signed releases
- Read-only operations (no mutations to DataHub)
- Token-based authentication
- Configurable rate limits

### Which DataHub versions are supported?

mcp-datahub is tested against DataHub 0.12.x and later. The GraphQL API is generally stable, but some features may vary between versions. The client handles version differences gracefully by returning empty results rather than errors for unsupported features.

### Which Go versions are supported?

Go 1.24 or later is required for building from source or using as a library.

## Technical

### How do I handle large result sets?

Use pagination with the `limit` and `offset` parameters:

```
datahub_search query="customer" limit=20 offset=0    # First 20
datahub_search query="customer" limit=20 offset=20   # Next 20
```

You can also configure default and maximum limits:

```bash
export DATAHUB_DEFAULT_LIMIT=20
export DATAHUB_MAX_LIMIT=50
```

### What is the recommended lineage depth?

The default depth of 3 works well for most use cases. Deeper lineage traversal (4-5) can be slow for entities with many dependencies. Start with 2-3 and increase if needed:

```bash
export DATAHUB_MAX_LINEAGE_DEPTH=5
```

### How do URNs work?

DataHub uses URNs (Uniform Resource Names) as unique identifiers. A dataset URN looks like:

```
urn:li:dataset:(urn:li:dataPlatform:snowflake,mydb.schema.table,PROD)
```

See [Understanding URNs](concepts/urns.md) for a complete explanation.

### What entity types can I search for?

Common entity types include:

- `DATASET`: Tables, views, files
- `DASHBOARD`: BI dashboards
- `CHART`: Individual charts/visualizations
- `DATA_FLOW`: Pipelines/workflows
- `DATA_JOB`: Individual pipeline tasks
- `GLOSSARY_TERM`: Business glossary terms
- `DOMAIN`: Organizational domains
- `DATA_PRODUCT`: Data products

### How do I connect to multiple DataHub instances?

Configure additional servers via environment variable:

```bash
export DATAHUB_URL=https://prod.datahub.example.com
export DATAHUB_TOKEN=prod-token
export DATAHUB_CONNECTION_NAME=prod

export DATAHUB_ADDITIONAL_SERVERS='{
  "staging": {
    "url": "https://staging.datahub.example.com",
    "token": "staging-token"
  }
}'
```

Then use the `connection` parameter to target a specific server:

```
datahub_search query="customer" connection="staging"
```

## Library Usage

### Can I use just the client without the MCP tools?

Yes. The client package is independent:

```go
import "github.com/txn2/mcp-datahub/pkg/client"

c, _ := client.New(client.Config{
    URL:   "https://datahub.example.com",
    Token: os.Getenv("DATAHUB_TOKEN"),
})

result, _ := c.Search(ctx, "customers")
```

### How do I add custom tools alongside DataHub tools?

Register DataHub tools first, then add your own:

```go
// DataHub tools
toolkit := tools.NewToolkit(datahubClient)
toolkit.RegisterAll(server)

// Your custom tools
server.AddTool(&mcp.Tool{
    Name:        "my_custom_tool",
    Description: "Does something custom",
}, myHandler)
```

### How do I filter which tools are registered?

Use selective registration:

```go
toolkit := tools.NewToolkit(datahubClient)
toolkit.Register(server,
    tools.ToolSearch,
    tools.ToolGetEntity,
    tools.ToolGetSchema,
)
```

### Can I modify tool responses?

Yes, use middleware:

```go
toolkit := tools.NewToolkit(datahubClient,
    tools.WithMiddleware(tools.AfterFunc(func(ctx context.Context, tc *tools.ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
        // Modify result here
        return result, err
    })),
)
```

### How do I add authentication to the library?

Implement the `AccessFilter` interface and use `WithAccessFilter`:

```go
type myAuth struct {
    authService AuthService
}

func (a *myAuth) CanAccess(ctx context.Context, urn string) (bool, error) {
    userID := ctx.Value("user_id").(string)
    return a.authService.CheckPermission(userID, urn)
}

func (a *myAuth) FilterURNs(ctx context.Context, urns []string) ([]string, error) {
    // Filter to only accessible URNs
}

toolkit := tools.NewToolkit(client,
    tools.WithAccessFilter(&myAuth{authService}),
)
```

## Troubleshooting

### Why am I getting "unauthorized" errors?

1. Your token may be expired. Generate a new one from DataHub Settings.
2. The token may not have sufficient permissions.
3. Check that DATAHUB_TOKEN is set correctly in your environment.

### Why is lineage empty for my dataset?

1. Lineage may not be ingested for this data platform.
2. The dataset may genuinely have no upstream/downstream dependencies.
3. Check the DataHub UI to verify lineage exists.

### Why are searches slow?

1. Reduce the `limit` parameter.
2. Add entity type filters to narrow results.
3. Increase the timeout: `DATAHUB_TIMEOUT=60`
4. Check DataHub server performance.

### Where can I get more help?

1. Check the [Troubleshooting Guide](support/troubleshooting.md)
2. Search [GitHub Issues](https://github.com/txn2/mcp-datahub/issues)
3. Open a new issue with version info and error messages
