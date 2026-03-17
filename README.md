[![txn2/mcp-datahub](docs/images/MCP-datahub-logo-banner.svg)](https://mcp-datahub.txn2.com)

[![GitHub license](https://img.shields.io/github/license/txn2/mcp-datahub.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/txn2/mcp-datahub.svg)](https://pkg.go.dev/github.com/txn2/mcp-datahub)
[![Go Report Card](https://goreportcard.com/badge/github.com/txn2/mcp-datahub)](https://goreportcard.com/report/github.com/txn2/mcp-datahub)
[![codecov](https://codecov.io/gh/txn2/mcp-datahub/branch/main/graph/badge.svg)](https://codecov.io/gh/txn2/mcp-datahub)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/txn2/mcp-datahub/badge)](https://scorecard.dev/viewer/?uri=github.com/txn2/mcp-datahub)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

An MCP server and composable Go library that connects AI assistants to [DataHub](https://datahubproject.io/) metadata catalogs. Search datasets, explore schemas, trace lineage, and access glossary terms and domains.

**[mcp-datahub.txn2.com](https://mcp-datahub.txn2.com)** | **[Installation](https://mcp-datahub.txn2.com/server/installation/)** | **[Library Docs](https://mcp-datahub.txn2.com/library/)**

## MCP Data Platform Ecosystem

mcp-datahub is part of a broader suite of open-source MCP servers designed to work together as a composable data platform. Each component can run standalone or be combined to give AI assistants unified access to storage, query engines, and metadata catalogs.

- [txn2/mcp-data-platform](https://github.com/txn2/mcp-data-platform/)
- [txn2/mcp-s3](https://github.com/txn2/mcp-s3/)
- [txn2/mcp-trino](https://github.com/txn2/mcp-trino/)

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

#### Customizing Tool Descriptions

Override tool descriptions to match your deployment:

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{},
    tools.WithDescriptions(map[tools.ToolName]string{
        tools.ToolSearch: "Search our internal data catalog for datasets and dashboards",
    }),
)
```

#### Customizing Tool Annotations

Override [MCP tool annotations](https://modelcontextprotocol.io/specification/2025-03-26/server/tools#annotations) (behavior hints for AI clients):

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{},
    tools.WithAnnotations(map[tools.ToolName]*mcp.ToolAnnotations{
        tools.ToolSearch: {ReadOnlyHint: true, OpenWorldHint: boolPtr(true)},
    }),
)
```

All 12 tools ship with default annotations: read tools are marked `ReadOnlyHint: true`; `datahub_create` is non-destructive and non-idempotent; `datahub_update` is non-destructive and idempotent; `datahub_delete` is destructive and idempotent.

#### Extensions (Logging, Metrics, Error Hints)

Enable optional middleware via the extensions package:

```go
import "github.com/txn2/mcp-datahub/pkg/extensions"

// Load from environment variables (MCP_DATAHUB_EXT_*)
cfg := extensions.FromEnv()
opts := extensions.BuildToolkitOptions(cfg)
toolkit := tools.NewToolkit(datahubClient, toolsCfg, opts...)

// Or load from a YAML/JSON config file
serverCfg, _ := extensions.LoadConfig("config.yaml")
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

### Read Tools (always available)

| Tool | Description |
|------|-------------|
| `datahub_search` | Search for datasets, dashboards, pipelines by query and entity type |
| `datahub_get_entity` | Get entity metadata by URN (description, owners, tags, domain) |
| `datahub_get_schema` | Get dataset schema with field types and descriptions |
| `datahub_get_lineage` | Get upstream/downstream lineage (supports `level=column` for column-level) |
| `datahub_get_queries` | Get SQL queries associated with a dataset |
| `datahub_browse` | Browse catalog: list tags, domains, or data products |
| `datahub_get_glossary_term` | Get glossary term definition and properties |
| `datahub_get_data_product` | Get data product details (owners, domain, properties) |
| `datahub_list_connections` | List configured DataHub server connections (multi-server mode) |

### Write Tools (require `DATAHUB_WRITE_ENABLED=true`)

3 CRUD tools using the `what` discriminator pattern — 35 operations total:

| Tool | Operations | Description |
|------|------------|-------------|
| `datahub_create` | 10 | Create tags, domains, glossary terms, data products, documents, applications, queries, incidents, structured properties, data contracts |
| `datahub_update` | 17 | Update descriptions, tags, glossary terms, links, owners, domains, structured properties, incidents, queries, documents, data contracts |
| `datahub_delete` | 8 | Delete queries, tags, domains, glossary entities, data products, applications, documents, structured properties |

Write tools are disabled by default for safety.

### DataHub Version Compatibility

**Minimum: DataHub 1.3.x. Full feature set: DataHub 1.4.x.**

| DataHub Version | Features |
|---|---|
| 1.3.x+ (minimum) | All read tools, core write operations (tags, domains, glossary, data products, queries, owners, links, descriptions, incidents, structured properties, data contracts) |
| 1.4.x+ (full) | + Documents (create/update/delete), applications (create/delete), `updateIncident`, `deleteStructuredProperty` |

The client gracefully handles version differences — read queries return empty results (not errors) when a feature is unavailable on older versions.

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
| `DATAHUB_WRITE_ENABLED` | Enable write operations (`true` or `1`) | `false` |
| `DATAHUB_DEBUG` | Enable debug logging (`1` or `true`) | `false` |

### Extensions

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_DATAHUB_EXT_LOGGING` | Enable structured logging of tool calls | `false` |
| `MCP_DATAHUB_EXT_METRICS` | Enable metrics collection | `false` |
| `MCP_DATAHUB_EXT_METADATA` | Enable metadata enrichment on results | `false` |
| `MCP_DATAHUB_EXT_ERRORS` | Enable error hint enrichment | `true` |

### Config File

As an alternative to environment variables, configure via YAML or JSON:

```yaml
datahub:
  url: https://datahub.example.com
  token: "${DATAHUB_TOKEN}"
  timeout: "30s"
  write_enabled: true

toolkit:
  default_limit: 20
  descriptions:
    datahub_search: "Custom search description for your deployment"

extensions:
  logging: true
  errors: true
```

Load with `extensions.LoadConfig("config.yaml")`. Environment variables override file values for sensitive fields. Token values support `$VAR` / `${VAR}` expansion.

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
