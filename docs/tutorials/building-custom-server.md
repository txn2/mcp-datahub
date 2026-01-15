# Tutorial: Building a Custom MCP Server

Learn how to create your own MCP server using the mcp-datahub library.

**Prerequisites**:

- Go 1.24 or later installed
- Basic Go programming knowledge
- A DataHub instance with token

## What You Will Learn

- How to set up a Go project with mcp-datahub
- How to configure the DataHub client
- How to register tools with an MCP server
- How to run and test your custom server

## Why Build a Custom Server?

The standalone `mcp-datahub` binary works for basic use cases. A custom server lets you:

- Add authentication and authorization
- Combine DataHub with other tools (Trino, S3, etc.)
- Add custom middleware for logging and metrics
- Implement tenant isolation
- Add company-specific tools

## Step 1: Create Your Project

Create a new Go project:

```bash
mkdir my-data-server
cd my-data-server
go mod init my-data-server
```

Add the required dependencies:

```bash
go get github.com/txn2/mcp-datahub
go get github.com/modelcontextprotocol/go-sdk
```

## Step 2: Write the Basic Server

Create `main.go`:

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
        log.Fatalf("Failed to create DataHub client: %v", err)
    }
    defer datahubClient.Close()

    // Create toolkit and register all tools
    toolkit := tools.NewToolkit(datahubClient)
    toolkit.RegisterAll(server)

    log.Println("Starting my-data-server...")

    // Run server with stdio transport
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

## Step 3: Build and Test

Build your server:

```bash
go build -o my-data-server
```

Test it manually:

```bash
export DATAHUB_URL=https://your-datahub.example.com
export DATAHUB_TOKEN=your_token

./my-data-server
```

The server starts and waits for MCP messages on stdin.

## Step 4: Configure Claude Desktop

Add your server to Claude Desktop. Edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "my-data-server": {
      "command": "/path/to/my-data-server",
      "env": {
        "DATAHUB_URL": "https://your-datahub.example.com",
        "DATAHUB_TOKEN": "your_token"
      }
    }
  }
}
```

Restart Claude Desktop and test:

> "Search for datasets in DataHub"

## Step 5: Add Custom Configuration

Customize the toolkit with configuration options:

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{
    DefaultLimit:    20,   // Default search results
    MaxLimit:        50,   // Maximum allowed
    MaxLineageDepth: 3,    // Maximum lineage depth
})
```

## Step 6: Register Specific Tools

Instead of registering all tools, select only what you need:

```go
toolkit := tools.NewToolkit(datahubClient)

// Only register search and schema tools
toolkit.Register(server,
    tools.ToolSearch,
    tools.ToolGetEntity,
    tools.ToolGetSchema,
)
```

Available tool constants:

| Constant | Tool Name |
|----------|-----------|
| `tools.ToolSearch` | `datahub_search` |
| `tools.ToolGetEntity` | `datahub_get_entity` |
| `tools.ToolGetSchema` | `datahub_get_schema` |
| `tools.ToolGetLineage` | `datahub_get_lineage` |
| `tools.ToolGetQueries` | `datahub_get_queries` |
| `tools.ToolGetGlossaryTerm` | `datahub_get_glossary_term` |
| `tools.ToolListTags` | `datahub_list_tags` |
| `tools.ToolListDomains` | `datahub_list_domains` |
| `tools.ToolListDataProducts` | `datahub_list_data_products` |
| `tools.ToolGetDataProduct` | `datahub_get_data_product` |

## Step 7: Add Logging Middleware

Add middleware to log all tool calls:

```go
toolkit := tools.NewToolkit(datahubClient,
    tools.WithMiddleware(tools.BeforeFunc(
        func(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
            log.Printf("Tool called: %s", tc.Name)
            return ctx, nil
        },
    )),
)
```

## Step 8: Add a Custom Tool

Extend your server with custom tools:

```go
// Register DataHub tools
toolkit := tools.NewToolkit(datahubClient)
toolkit.RegisterAll(server)

// Add a custom tool
server.AddTool(&mcp.Tool{
    Name:        "company_data_dictionary",
    Description: "Get company-specific data dictionary entries",
    InputSchema: mcp.ToolInputSchema{
        Type: "object",
        Properties: map[string]any{
            "term": map[string]any{
                "type":        "string",
                "description": "Term to look up",
            },
        },
        Required: []string{"term"},
    },
}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    term := req.Arguments["term"].(string)

    // Your custom logic here
    definition := lookupInCompanyDictionary(term)

    return tools.TextResult(definition), nil
})
```

## Step 9: Combine with Other Toolkits

Add Trino tools alongside DataHub:

```go
import (
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
    dhClient, _ := datahubclient.New(datahubclient.Config{
        URL:   os.Getenv("DATAHUB_URL"),
        Token: os.Getenv("DATAHUB_TOKEN"),
    })
    datahubtools.NewToolkit(dhClient).RegisterAll(server)

    // Trino tools
    trClient, _ := trinoclient.New(trinoclient.Config{
        Host: os.Getenv("TRINO_HOST"),
        User: os.Getenv("TRINO_USER"),
    })
    trinotools.NewToolkit(trClient).RegisterAll(server)

    // Now AI assistants can:
    // - Search DataHub for tables
    // - Get schema information
    // - Query data via Trino

    server.Run(context.Background(), &mcp.StdioTransport{})
}
```

## Complete Example

Here is a complete server with logging and configuration:

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/txn2/mcp-datahub/pkg/client"
    "github.com/txn2/mcp-datahub/pkg/tools"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "my-data-server",
        Version: "1.0.0",
    }, nil)

    datahubClient, err := client.New(client.Config{
        URL:     os.Getenv("DATAHUB_URL"),
        Token:   os.Getenv("DATAHUB_TOKEN"),
        Timeout: 30 * time.Second,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer datahubClient.Close()

    toolkit := tools.NewToolkit(datahubClient,
        tools.Config{
            DefaultLimit:    20,
            MaxLimit:        100,
            MaxLineageDepth: 5,
        },
        tools.WithMiddleware(loggingMiddleware()),
    )
    toolkit.RegisterAll(server)

    log.Println("Server starting...")
    if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}

func loggingMiddleware() tools.ToolMiddleware {
    return tools.BeforeFunc(func(ctx context.Context, tc *tools.ToolContext) (context.Context, error) {
        log.Printf("[%s] Tool: %s", time.Now().Format(time.RFC3339), tc.Name)
        return ctx, nil
    })
}
```

## What You Learned

- Setting up a Go project with mcp-datahub
- Creating and configuring the DataHub client
- Registering all or specific tools
- Adding logging middleware
- Combining multiple toolkits
- Adding custom tools

## Next Steps

- [Adding Authentication](adding-authentication.md): Secure your server
- [Composability Guide](../library/composability.md): Advanced patterns
- [Middleware Documentation](../guides/custom-middleware.md): Custom middleware
