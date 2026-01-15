# txn2/mcp-datahub

[![GitHub license](https://img.shields.io/github/license/txn2/mcp-datahub.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/txn2/mcp-datahub.svg)](https://pkg.go.dev/github.com/txn2/mcp-datahub)
[![Go Report Card](https://goreportcard.com/badge/github.com/txn2/mcp-datahub)](https://goreportcard.com/report/github.com/txn2/mcp-datahub)
[![codecov](https://codecov.io/gh/txn2/mcp-datahub/branch/main/graph/badge.svg)](https://codecov.io/gh/txn2/mcp-datahub)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/txn2/mcp-datahub/badge)](https://scorecard.dev/viewer/?uri=github.com/txn2/mcp-datahub)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

An MCP server and composable Go library that connects AI assistants to [DataHub](https://datahubproject.io/) metadata catalogs. Search datasets, explore schemas, trace lineage, and access glossary terms and domains.

**[Documentation](https://mcp-datahub.txn2.com)** | **[Installation](https://mcp-datahub.txn2.com/server/installation/)** | **[Library Docs](https://mcp-datahub.txn2.com/library/)**

## Two Ways to Use

### 1. Standalone MCP Server

Install and connect to Claude Desktop, Cursor, or any MCP client:

```bash
# Homebrew (macOS)
brew install txn2/tap/mcp-datahub

# Go install
go install github.com/txn2/mcp-datahub/cmd/mcp-datahub@latest
```

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

### 2. Composable Go Library

Import into your own MCP server for custom authentication, tenant isolation, and audit logging:

```go
import (
    "github.com/txn2/mcp-datahub/pkg/client"
    "github.com/txn2/mcp-datahub/pkg/tools"
)

// Create client and register tools with your MCP server
datahubClient, _ := client.NewFromEnv()
defer datahubClient.Close()

toolkit := tools.NewToolkit(datahubClient, tools.Config{})
toolkit.RegisterAll(yourMCPServer)
```

See the [library documentation](https://mcp-datahub.txn2.com/library/) for middleware, selective tool registration, and enterprise patterns.

## Combining with mcp-trino

Build a unified data platform MCP server by combining DataHub metadata with Trino query execution:

```go
import (
    datahubClient "github.com/txn2/mcp-datahub/pkg/client"
    datahubTools "github.com/txn2/mcp-datahub/pkg/tools"
    trinoClient "github.com/txn2/mcp-trino/pkg/client"
    trinoTools "github.com/txn2/mcp-trino/pkg/tools"
)

// Add DataHub tools (search, lineage, schema, glossary)
dh, _ := datahubClient.NewFromEnv()
datahubTools.NewToolkit(dh, datahubTools.Config{}).RegisterAll(server)

// Add Trino tools (query execution, catalog browsing)
tr, _ := trinoClient.NewFromEnv()
trinoTools.NewToolkit(tr, trinoTools.Config{}).RegisterAll(server)

// AI assistants can now:
// - Search DataHub for tables -> Get schema -> Query via Trino
// - Explore lineage -> Understand data flow -> Run validation queries
```

See [txn2/mcp-trino](https://github.com/txn2/mcp-trino) for the companion library.

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

See the [tools reference](https://mcp-datahub.txn2.com/server/tools/) for detailed documentation.

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DATAHUB_URL` | DataHub GMS URL | (required) |
| `DATAHUB_TOKEN` | API token | (required) |
| `DATAHUB_TIMEOUT` | Request timeout (seconds) | `30` |
| `DATAHUB_DEFAULT_LIMIT` | Default search limit | `10` |
| `DATAHUB_MAX_LIMIT` | Maximum limit | `100` |

See [configuration reference](https://mcp-datahub.txn2.com/server/configuration/) for all options.

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

## Related Projects

- [txn2/mcp-trino](https://github.com/txn2/mcp-trino) ([docs](https://mcp-trino.txn2.com)) - Composable MCP toolkit for Trino query execution
- [DataHub](https://datahubproject.io/) - The open-source metadata platform

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[Apache License 2.0](LICENSE)

---

Open source by [Craig Johnston](https://twitter.com/cjimti), sponsored by [Deasil Works, Inc.](https://deasil.works/)
