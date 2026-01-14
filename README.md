# txn2/mcp-datahub

[![GitHub license](https://img.shields.io/github/license/txn2/mcp-datahub.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/txn2/mcp-datahub.svg)](https://pkg.go.dev/github.com/txn2/mcp-datahub)
[![Go Report Card](https://goreportcard.com/badge/github.com/txn2/mcp-datahub)](https://goreportcard.com/report/github.com/txn2/mcp-datahub)
[![codecov](https://codecov.io/gh/txn2/mcp-datahub/branch/main/graph/badge.svg)](https://codecov.io/gh/txn2/mcp-datahub)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/txn2/mcp-datahub/badge)](https://scorecard.dev/viewer/?uri=github.com/txn2/mcp-datahub)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

A **composable Go library** for building custom MCP servers that integrate [DataHub](https://datahubproject.io/) metadata capabilities. Part of the txn2 MCP toolkit ecosystem.

## Why This Library?

**This is not a replacement for the official DataHub MCP server.** Instead, mcp-datahub is designed as a building block for organizations that need to:

- **Combine multiple data tools** into a unified MCP server (DataHub + Trino + dbt + custom tools)
- **Add enterprise features** like multi-tenant isolation, custom auth, or audit logging
- **Build domain-specific assistants** that understand your particular data stack
- **Extend functionality** with middleware, hooks, and custom integrations

The included CLI binary (`mcp-datahub`) serves as a **reference implementation** demonstrating the library's capabilities—it's fully functional as a standalone DataHub MCP server, but the real power comes from composition.

## Composable Architecture

### Combining Multiple Toolkits

Build a unified MCP server that gives AI assistants access to your entire data stack:

```go
package main

import (
    "context"
    "log"

    "github.com/modelcontextprotocol/go-sdk/mcp"

    // DataHub for metadata and lineage
    datahubClient "github.com/txn2/mcp-datahub/pkg/client"
    datahubTools "github.com/txn2/mcp-datahub/pkg/tools"

    // Trino for query execution
    trinoClient "github.com/txn2/mcp-trino/pkg/client"
    trinoTools "github.com/txn2/mcp-trino/pkg/tools"
)

func main() {
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "my-data-platform",
        Version: "1.0.0",
    }, nil)

    // Add DataHub tools (search, lineage, schema, glossary)
    dh, _ := datahubClient.NewFromEnv()
    defer dh.Close()
    datahubTools.NewToolkit(dh, datahubTools.Config{}).RegisterAll(server)

    // Add Trino tools (query execution, catalog browsing)
    tr, _ := trinoClient.NewFromEnv()
    defer tr.Close()
    trinoTools.NewToolkit(tr, trinoTools.Config{}).RegisterAll(server)

    // Now AI assistants can:
    // - Search DataHub for tables → Get schema → Query via Trino
    // - Explore lineage → Understand data flow → Run validation queries
    // - Look up glossary terms → Find related datasets → Analyze data

    server.Run(context.Background(), &mcp.StdioTransport{})
}
```

### Selective Tool Registration

Register only the tools you need:

```go
toolkit := datahubTools.NewToolkit(client, config)

// Register specific tools
toolkit.Register(server,
    datahubTools.ToolSearch,
    datahubTools.ToolGetSchema,
    datahubTools.ToolGetLineage,
)

// Or register all
toolkit.RegisterAll(server)
```

### Middleware for Enterprise Features

Add cross-cutting concerns without modifying tool implementations:

```go
// Audit logging middleware
auditMiddleware := &AuditMiddleware{logger: auditLog}

// Rate limiting middleware
rateLimiter := &RateLimitMiddleware{limiter: limiter}

toolkit := datahubTools.NewToolkit(client, config,
    datahubTools.WithMiddleware(auditMiddleware),
    datahubTools.WithMiddleware(rateLimiter),
    datahubTools.WithToolMiddleware(datahubTools.ToolSearch, searchSpecificMiddleware),
)
```

## Available Tools

| Tool | Description |
|------|-------------|
| `datahub_search` | Search for datasets, dashboards, pipelines by query and entity type |
| `datahub_get_entity` | Get entity metadata by URN (description, owners, tags, domain) |
| `datahub_get_schema` | Get dataset schema with field types and descriptions |
| `datahub_get_lineage` | Get upstream/downstream data lineage |
| `datahub_get_queries` | Get SQL queries associated with a dataset |
| `datahub_get_glossary_term` | Get glossary term definition and properties |
| `datahub_list_tags` | List available tags in the catalog |
| `datahub_list_domains` | List data domains |
| `datahub_list_data_products` | List data products |
| `datahub_get_data_product` | Get data product details (owners, domain, properties) |

## Installation

### As a Library (Recommended)

```bash
go get github.com/txn2/mcp-datahub
```

### Reference Implementation Binary

For standalone use or testing:

```bash
# Homebrew (macOS)
brew install txn2/tap/mcp-datahub

# Go install
go install github.com/txn2/mcp-datahub/cmd/mcp-datahub@latest
```

### Claude Desktop Configuration

```json
{
  "mcpServers": {
    "datahub": {
      "command": "/opt/homebrew/bin/mcp-datahub",
      "env": {
        "DATAHUB_URL": "https://datahub.example.com",
        "DATAHUB_TOKEN": "your_token"
      }
    }
  }
}
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DATAHUB_URL` | DataHub GMS URL | (required) |
| `DATAHUB_TOKEN` | API token | (required) |
| `DATAHUB_TIMEOUT` | Request timeout (seconds) | `30` |
| `DATAHUB_DEFAULT_LIMIT` | Default search limit | `10` |
| `DATAHUB_MAX_LIMIT` | Maximum limit | `100` |

## Library Usage Examples

### Basic Usage

```go
import (
    "github.com/txn2/mcp-datahub/pkg/client"
    "github.com/txn2/mcp-datahub/pkg/tools"
)

// Create client from environment
datahubClient, err := client.NewFromEnv()
if err != nil {
    log.Fatal(err)
}
defer datahubClient.Close()

// Create toolkit and register with your MCP server
toolkit := tools.NewToolkit(datahubClient, tools.Config{})
toolkit.RegisterAll(yourMCPServer)
```

### Direct Client Usage

Use the client directly for custom integrations:

```go
import "github.com/txn2/mcp-datahub/pkg/client"

c, _ := client.New(client.Config{
    URL:   "https://datahub.example.com",
    Token: "your_token",
})

// Search for datasets
results, _ := c.Search(ctx, "customer", client.WithEntityType("DATASET"), client.WithLimit(10))

// Get entity details
entity, _ := c.GetEntity(ctx, "urn:li:dataset:(urn:li:dataPlatform:postgres,mydb.users,PROD)")

// Get schema
schema, _ := c.GetSchema(ctx, datasetURN)

// Get lineage
lineage, _ := c.GetLineage(ctx, datasetURN, client.WithDirection(client.LineageDirectionUpstream))
```

### Custom MCP Server with Multiple Toolkits

See [examples/combined-server](examples/combined-server) for a complete example combining DataHub with other data tools.

## Related Projects

- [txn2/mcp-trino](https://github.com/txn2/mcp-trino) - Composable MCP toolkit for Trino query execution
- [DataHub](https://datahubproject.io/) - The open-source metadata platform

## Development

```bash
make build     # Build binary
make test      # Run tests with race detection
make lint      # Run golangci-lint
make security  # Run gosec and govulncheck
make coverage  # Generate coverage report
make verify    # Run tidy, lint, and test
make help      # Show all targets
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[Apache License 2.0](LICENSE)

---

Open source by [Craig Johnston](https://twitter.com/cjimti)
