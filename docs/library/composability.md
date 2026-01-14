# Composability

mcp-datahub is designed for composition with other MCP tool libraries.

## Island Architecture

Each txn2 library (mcp-trino, mcp-datahub, mcp-nifi, mcp-s3) is an **island**:

- No knowledge of each other
- No shared dependencies
- Can be used independently or together

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  mcp-trino      │     │  mcp-datahub    │     │  mcp-nifi       │
│ (no imports)    │     │ (no imports)    │     │ (no imports)    │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌────────────────────────┐
                    │    Your Custom MCP     │
                    │  Imports all, wires    │
                    └────────────────────────┘
```

## Basic Composition

```go
package main

import (
    "github.com/modelcontextprotocol/go-sdk/mcp"

    datahubclient "github.com/txn2/mcp-datahub/pkg/client"
    datahubtools "github.com/txn2/mcp-datahub/pkg/tools"

    trinoclient "github.com/txn2/mcp-trino/pkg/client"
    trinotools "github.com/txn2/mcp-trino/pkg/tools"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "unified-data-server",
        Version: "1.0.0",
    }, nil)

    // DataHub tools
    datahub, _ := datahubclient.NewFromEnv()
    datahubtools.NewToolkit(datahub).RegisterAll(server)

    // Trino tools
    trino, _ := trinoclient.NewFromEnv()
    trinotools.NewToolkit(trino).RegisterAll(server)

    // Run combined server
    server.Run(ctx, &mcp.StdioTransport{})
}
```

## Adding Middleware

Add cross-cutting concerns like logging or access control:

```go
// Create toolkit with middleware
toolkit := tools.NewToolkit(datahubClient,
    tools.WithMiddleware(loggingMiddleware),
    tools.WithMiddleware(accessControlMiddleware),
)

// Logging middleware
func loggingMiddleware(next tools.Handler) tools.Handler {
    return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        log.Printf("Tool called: %s", req.Name)
        return next(ctx, req)
    }
}
```

## Per-Tool Middleware

Apply middleware to specific tools:

```go
toolkit := tools.NewToolkit(datahubClient,
    tools.WithToolMiddleware(tools.ToolSearch, rateLimiter),
    tools.WithToolMiddleware(tools.ToolGetLineage, cacheMiddleware),
)
```

## Selective Registration

Register only the tools you need:

```go
// DataHub: only search and entity tools
datahubToolkit.Register(server,
    tools.ToolSearch,
    tools.ToolGetEntity,
)

// Trino: only query tool
trinoToolkit.Register(server,
    trinotools.ToolQuery,
)
```

## Adding Custom Tools

Extend with your own domain-specific tools:

```go
// Register DataHub tools
datahubtools.NewToolkit(datahub).RegisterAll(server)

// Add your custom tools
mcp.AddTool(server, &mcp.Tool{
    Name:        "company_data_dictionary",
    Description: "Get company-specific data dictionary",
}, yourHandler)
```
