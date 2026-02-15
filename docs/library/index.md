# Go Library Overview

mcp-datahub is designed as a **composable Go library**. Import it into your own MCP server to add DataHub capabilities alongside other tools.

## Why Use as a Library?

- **Custom Authentication**: Add your own auth layer (OAuth, API keys, etc.)
- **Tenant Isolation**: Filter results by tenant/organization
- **Audit Logging**: Log all tool invocations for compliance
- **Tool Composition**: Combine with Trino, NiFi, S3, and other MCP tools
- **Custom Middleware**: Add rate limiting, caching, or transformations

## Packages

### pkg/client

The DataHub GraphQL client:

```go
import "github.com/txn2/mcp-datahub/pkg/client"

c, err := client.New(client.Config{
    URL:   "https://datahub.company.com",
    Token: os.Getenv("DATAHUB_TOKEN"),
})

result, err := c.Search(ctx, "customers", client.WithLimit(20))
```

### pkg/tools

The MCP toolkit for registering tools:

```go
import "github.com/txn2/mcp-datahub/pkg/tools"

toolkit := tools.NewToolkit(datahubClient)
toolkit.RegisterAll(server)
```

### pkg/types

Domain types for DataHub entities:

```go
import "github.com/txn2/mcp-datahub/pkg/types"

var entity *types.Entity
var schema *types.SchemaMetadata
var lineage *types.LineageResult
```

### pkg/integration

Extension interfaces for enterprise integrations (access control, audit logging, query providers):

```go
import "github.com/txn2/mcp-datahub/pkg/integration"
```

### pkg/extensions

Built-in middleware for logging, metrics, error hints, and metadata enrichment. Also provides YAML/JSON config file support:

```go
import "github.com/txn2/mcp-datahub/pkg/extensions"

// Load from environment variables
cfg := extensions.FromEnv()
opts := extensions.BuildToolkitOptions(cfg)
toolkit := tools.NewToolkit(datahubClient, toolsCfg, opts...)

// Or load from a config file
serverCfg, _ := extensions.LoadConfig("config.yaml")
```

### pkg/multiserver

Multi-server configuration and connection management for connecting to multiple DataHub instances:

```go
import "github.com/txn2/mcp-datahub/pkg/multiserver"
```

## Next Steps

- [Quick Start Guide](quickstart.md)
- [Architecture](architecture.md)
- [Composability](composability.md)
