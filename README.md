![txn2/mcp-datahub](./docs/images/txn2_mcp_datahub_banner.png)

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

**Claude Desktop (Easiest)** - Download the `.mcpb` bundle from [releases](https://github.com/txn2/mcp-datahub/releases) and double-click to install:
- macOS Apple Silicon: `mcp-datahub_X.X.X_darwin_arm64.mcpb`
- macOS Intel: `mcp-datahub_X.X.X_darwin_amd64.mcpb`
- Windows: `mcp-datahub_X.X.X_windows_amd64.mcpb`

**Other Installation Methods:**
```bash
# Homebrew (macOS)
brew install txn2/tap/mcp-datahub

# Go install
go install github.com/txn2/mcp-datahub/cmd/mcp-datahub@latest
```

**Manual Claude Desktop Configuration** (if not using MCPB):
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

#### Multi-Server Configuration

Connect to multiple DataHub instances simultaneously:

```bash
# Primary server
export DATAHUB_URL=https://prod.datahub.example.com/api/graphql
export DATAHUB_TOKEN=prod-token
export DATAHUB_CONNECTION_NAME=prod

# Additional servers (JSON)
export DATAHUB_ADDITIONAL_SERVERS='{"staging":{"url":"https://staging.datahub.example.com/api/graphql","token":"staging-token"}}'
```

Use `datahub_list_connections` to discover available connections, then pass the `connection` parameter to any tool.

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

### Bidirectional Integration with QueryProvider

The library supports bidirectional context injection. While mcp-trino can pull semantic context from DataHub, mcp-datahub can receive query execution context back from a query engine:

```go
import (
    datahubTools "github.com/txn2/mcp-datahub/pkg/tools"
    "github.com/txn2/mcp-datahub/pkg/integration"
)

// QueryProvider enables query engines to inject context into DataHub tools
type myQueryProvider struct {
    trinoClient *trino.Client
}

func (p *myQueryProvider) Name() string { return "trino" }

func (p *myQueryProvider) ResolveTable(ctx context.Context, urn string) (*integration.TableIdentifier, error) {
    // Map DataHub URN to Trino table (catalog.schema.table)
    return &integration.TableIdentifier{
        Catalog: "hive", Schema: "production", Table: "users",
    }, nil
}

func (p *myQueryProvider) GetTableAvailability(ctx context.Context, urn string) (*integration.TableAvailability, error) {
    // Check if table is queryable
    return &integration.TableAvailability{Available: true}, nil
}

func (p *myQueryProvider) GetQueryExamples(ctx context.Context, urn string) ([]integration.QueryExample, error) {
    // Return sample queries for this entity
    return []integration.QueryExample{
        {Name: "sample", SQL: "SELECT * FROM hive.production.users LIMIT 10"},
    }, nil
}

// Wire it up
toolkit := datahubTools.NewToolkit(datahubClient, config,
    datahubTools.WithQueryProvider(&myQueryProvider{trinoClient: trino}),
)
```

When a QueryProvider is configured, tool responses are enriched:
- **Search results**: Include `query_context` with table availability
- **Entity details**: Include `query_table`, `query_examples`, `query_availability`
- **Schema**: Include `query_table` for immediate SQL usage
- **Lineage**: Include `execution_context` mapping URNs to tables

### Integration Middleware

Enterprise features like access control and audit logging are enabled through middleware adapters:

```go
import (
    datahubTools "github.com/txn2/mcp-datahub/pkg/tools"
    "github.com/txn2/mcp-datahub/pkg/integration"
)

// Access control - filter entities by user permissions
type myAccessFilter struct{}
func (f *myAccessFilter) CanAccess(ctx context.Context, urn string) (bool, error) { /* ... */ }
func (f *myAccessFilter) FilterURNs(ctx context.Context, urns []string) ([]string, error) { /* ... */ }

// Audit logging - track all tool invocations
type myAuditLogger struct{}
func (l *myAuditLogger) LogToolCall(ctx context.Context, tool string, params map[string]any, userID string) error { /* ... */ }

// Wire up with multiple integration options
toolkit := datahubTools.NewToolkit(datahubClient, config,
    datahubTools.WithAccessFilter(&myAccessFilter{}),
    datahubTools.WithAuditLogger(&myAuditLogger{}, func(ctx context.Context) string {
        return ctx.Value("user_id").(string)
    }),
    datahubTools.WithURNResolver(&myURNResolver{}),      // Map external IDs to URNs
    datahubTools.WithMetadataEnricher(&myEnricher{}),    // Add custom metadata
)
```

See the [library documentation](https://mcp-datahub.txn2.com/library/) for complete integration patterns.

## Available Tools

| Tool | Description |
|------|-------------|
| `datahub_search` | Search for datasets, dashboards, pipelines by query and entity type |
| `datahub_get_entity` | Get entity metadata by URN (description, owners, tags, domain) |
| `datahub_get_schema` | Get dataset schema with field types and descriptions |
| `datahub_get_lineage` | Get upstream/downstream data lineage |
| `datahub_get_column_lineage` | Get fine-grained column-level lineage mappings |
| `datahub_get_queries` | Get SQL queries associated with a dataset |
| `datahub_get_glossary_term` | Get glossary term definition and properties |
| `datahub_list_tags` | List available tags in the catalog |
| `datahub_list_domains` | List data domains |
| `datahub_list_data_products` | List data products |
| `datahub_get_data_product` | Get data product details (owners, domain, properties) |
| `datahub_list_connections` | List configured DataHub server connections (multi-server mode) |

See the [tools reference](https://mcp-datahub.txn2.com/server/tools/) for detailed documentation.

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DATAHUB_URL` | DataHub GraphQL API URL | (required) |
| `DATAHUB_TOKEN` | API token | (required) |
| `DATAHUB_TIMEOUT` | Request timeout (seconds) | `30` |
| `DATAHUB_DEFAULT_LIMIT` | Default search limit | `10` |
| `DATAHUB_MAX_LIMIT` | Maximum limit | `100` |
| `DATAHUB_CONNECTION_NAME` | Display name for primary connection | `datahub` |
| `DATAHUB_ADDITIONAL_SERVERS` | JSON map of additional servers | (optional) |

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
