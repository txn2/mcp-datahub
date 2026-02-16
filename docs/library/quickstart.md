# Quick Start

Get started with mcp-datahub as a Go library in minutes.

## Installation

```bash
go get github.com/txn2/mcp-datahub
```

## Basic Usage

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/txn2/mcp-datahub/pkg/client"
    "github.com/txn2/mcp-datahub/pkg/tools"
)

func main() {
    // Create MCP server
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "my-data-server",
        Version: "1.0.0",
    }, nil)

    // Create DataHub client
    datahubClient, err := client.New(client.Config{
        URL:   os.Getenv("DATAHUB_URL"),
        Token: os.Getenv("DATAHUB_TOKEN"),
    })
    if err != nil {
        log.Fatal(err)
    }
    defer datahubClient.Close()

    // Register all DataHub tools
    toolkit := tools.NewToolkit(datahubClient)
    toolkit.RegisterAll(server)

    // Run server
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatal(err)
    }
}
```

## Register Specific Tools

Only register the tools you need:

```go
toolkit := tools.NewToolkit(datahubClient)
toolkit.Register(server,
    tools.ToolSearch,
    tools.ToolGetEntity,
    tools.ToolGetSchema,
)
```

## Using the Client Directly

```go
// Search for datasets
result, err := datahubClient.Search(ctx, "customers",
    client.WithEntityType("DATASET"),
    client.WithLimit(20),
)

// Get entity details
entity, err := datahubClient.GetEntity(ctx, "urn:li:dataset:...")

// Get schema
schema, err := datahubClient.GetSchema(ctx, "urn:li:dataset:...")

// Get lineage
lineage, err := datahubClient.GetLineage(ctx, "urn:li:dataset:...",
    client.WithDirection("UPSTREAM"),
    client.WithDepth(2),
)
```

## With Custom Configuration

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{
    DefaultLimit:    20,
    MaxLimit:        50,
    MaxLineageDepth: 3,
})
```

## With Description Overrides

Customize tool descriptions to match your deployment:

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{},
    tools.WithDescriptions(map[tools.ToolName]string{
        tools.ToolSearch: "Search our internal data catalog for datasets and dashboards",
    }),
)
```

## With Annotation Overrides

Customize MCP tool annotations (behavior hints for AI clients):

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{},
    tools.WithAnnotations(map[tools.ToolName]*mcp.ToolAnnotations{
        tools.ToolSearch: {ReadOnlyHint: true, OpenWorldHint: boolPtr(true)},
    }),
)
```

All tools ship with sensible defaults (read tools marked read-only, write tools marked non-destructive and idempotent). See the [Tools API Reference](../reference/tools-api.md#withannotations) for the full annotation API.

## With Extensions

Enable built-in middleware for logging, metrics, and error hints:

```go
import "github.com/txn2/mcp-datahub/pkg/extensions"

// Load from environment variables (MCP_DATAHUB_EXT_*)
cfg := extensions.FromEnv()
opts := extensions.BuildToolkitOptions(cfg)
toolkit := tools.NewToolkit(datahubClient, toolsCfg, opts...)
```

Or load everything from a config file:

```go
serverCfg, _ := extensions.LoadConfig("config.yaml")
clientCfg := serverCfg.ClientConfig()
toolsCfg := serverCfg.ToolsConfig()
extOpts := extensions.BuildToolkitOptions(serverCfg.ExtConfig())

datahubClient, _ := client.New(clientCfg)
toolkit := tools.NewToolkit(datahubClient, toolsCfg, extOpts...)
```

## Next Steps

- [Architecture](architecture.md)
- [Composability](composability.md)
- [Tools API Reference](../reference/tools-api.md)
