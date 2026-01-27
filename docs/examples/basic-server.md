# Example: Basic Server

Minimal MCP server with DataHub tools.

## Complete Code

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
        Name:    "basic-datahub-server",
        Version: "1.0.0",
    }, nil)

    // Create DataHub client from environment
    datahubClient, err := client.New(client.Config{
        URL:   os.Getenv("DATAHUB_URL"),
        Token: os.Getenv("DATAHUB_TOKEN"),
    })
    if err != nil {
        log.Fatalf("Failed to create DataHub client: %v", err)
    }
    defer datahubClient.Close()

    // Create toolkit and register all tools
    toolkit := tools.NewToolkit(datahubClient)
    toolkit.RegisterAll(server)

    log.Println("Starting basic DataHub MCP server...")

    // Run with stdio transport
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

## Configuration

Set these environment variables:

```bash
export DATAHUB_URL=https://your-datahub.example.com
export DATAHUB_TOKEN=your_personal_access_token
```

## Build and Run

```bash
# Initialize module
go mod init basic-server
go mod tidy

# Build
go build -o basic-server

# Run
./basic-server
```

## Claude Desktop Configuration

```json
{
  "mcpServers": {
    "datahub": {
      "command": "/path/to/basic-server",
      "env": {
        "DATAHUB_URL": "https://your-datahub.example.com",
        "DATAHUB_TOKEN": "your_token"
      }
    }
  }
}
```

## Available Tools

This example registers all 12 DataHub tools:

- `datahub_search`
- `datahub_get_entity`
- `datahub_get_schema`
- `datahub_get_lineage`
- `datahub_get_column_lineage`
- `datahub_get_queries`
- `datahub_get_glossary_term`
- `datahub_list_tags`
- `datahub_list_domains`
- `datahub_list_data_products`
- `datahub_get_data_product`
- `datahub_list_connections`

## Selective Registration

To register only specific tools:

```go
toolkit := tools.NewToolkit(datahubClient)
toolkit.Register(server,
    tools.ToolSearch,
    tools.ToolGetEntity,
    tools.ToolGetSchema,
)
```

## With Custom Configuration

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{
    DefaultLimit:    20,
    MaxLimit:        100,
    MaxLineageDepth: 5,
})
```

## Next Steps

- [With Authentication](with-authentication.md): Add JWT authentication
- [Combined Trino](combined-trino.md): Add Trino query execution
