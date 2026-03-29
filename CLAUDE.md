# CLAUDE.md

This file provides guidance to Claude Code when working with this project.

## Project Overview

**mcp-datahub** is a composable Go library for building custom MCP (Model Context Protocol) servers that integrate DataHub metadata capabilities. It is part of the txn2 MCP toolkit ecosystem alongside [mcp-trino](https://github.com/txn2/mcp-trino).

### Project Positioning

**This is NOT a replacement for the official DataHub MCP server.** This library serves a different purpose:

1. **Primary Use**: A composable building block for creating unified MCP servers that combine multiple data tools (DataHub + Trino + dbt + custom tools) into a single AI assistant interface.

2. **Secondary Use**: The included `mcp-datahub` binary is a reference implementation that demonstrates library capabilities—functional as a standalone DataHub MCP server.

### Key Design Principles

- **Composable**: Designed to work alongside other MCP toolkits (mcp-trino, etc.)
- **Library-first**: The CLI binary is a reference implementation, not the primary deliverable
- **Extensible**: Middleware system for adding auth, audit logging, rate limiting
- **Island architecture**: No dependencies on other txn2 libraries
- **Generic**: No domain-specific logic; suitable for any DataHub deployment

## Architecture

```
pkg/
├── client/      # DataHub GraphQL client (standalone, no MCP dependency)
├── multiserver/ # Multi-server configuration and connection management
├── types/       # Domain types (entities, schema, lineage, etc.)
├── tools/       # MCP toolkit (composable tool registration)
├── integration/ # Extension interfaces for custom integrations
└── extensions/  # Optional middleware: logging, metrics, error hints, config files

internal/
└── server/      # Reference implementation server setup

cmd/
└── mcp-datahub/ # Reference implementation CLI binary

mcpb/
├── manifest.json # MCPB bundle manifest for Claude Desktop
└── build.sh      # Build script for .mcpb bundles
```

### Composition Pattern

The core pattern enables combining multiple toolkits:

```go
import (
    datahubTools "github.com/txn2/mcp-datahub/pkg/tools"
    trinoTools "github.com/txn2/mcp-trino/pkg/tools"
)

// Add DataHub tools to server
datahubTools.NewToolkit(datahubClient, config).RegisterAll(server)

// Add Trino tools to same server
trinoTools.NewToolkit(trinoClient, config).RegisterAll(server)

// AI assistants now have unified access to both
```

## Code Standards

1. **Idiomatic Go**: All code must follow idiomatic Go patterns. Use gofmt, follow Effective Go guidelines.

2. **Test Coverage**: Project must maintain >80% unit test coverage. Use table-driven tests.

3. **Testing Definition**: When asked to "test", run the full CI suite:
   - Unit tests with race detection: `go test -race ./...`
   - Linting: `golangci-lint run`
   - Security scanning: `gosec ./...` and `govulncheck ./...`
   - Cyclomatic complexity: `gocyclo -over 15 .` (must have no output)
   - All CI checks must pass locally

4. **Human Review Required**: A human must review every line before commit.

5. **Go Report Card**: MUST maintain 100% across ALL categories including:
   - go vet
   - gofmt
   - gocyclo (no functions with complexity > 15)
   - ineffassign
   - license
   - misspell

## Building and Running

```bash
# Build
go build -o mcp-datahub ./cmd/mcp-datahub

# Run reference implementation
export DATAHUB_URL=https://datahub.example.com
export DATAHUB_TOKEN=your_token
./mcp-datahub
```

## Available Tools (12 total: 9 read + 3 write)

Each tool has a `Title` (human-readable display name shown in MCP clients like Claude Desktop),
an `OutputSchema` (JSON Schema describing the response structure), and `Annotations`.
All are customizable via the three-tier priority pattern.

### Read Tools

| Tool | Title | Description |
|------|-------|-------------|
| `datahub_search` | Search Catalog | Search entities by query and type |
| `datahub_get_entity` | Get Entity | Get entity metadata by URN |
| `datahub_get_schema` | Get Schema | Get dataset schema |
| `datahub_get_lineage` | Get Lineage | Get upstream/downstream lineage; set `level=column` for column-level lineage |
| `datahub_get_queries` | Get Queries | Get associated SQL queries |
| `datahub_browse` | Browse Catalog | Browse tags, domains, or data products (set `what` parameter) |
| `datahub_get_glossary_term` | Get Glossary Term | Get glossary term details |
| `datahub_get_data_product` | Get Data Product | Get data product details |
| `datahub_list_connections` | List Connections | List configured DataHub server connections |

### Write Tools (require `WriteEnabled: true`)

3 CRUD tools using the discriminator pattern (`what` parameter) to cover 35 operations:

| Tool | Title | Operations | Description |
|------|-------|------------|-------------|
| `datahub_create` | Create Entity | 10 | Create tags, domains, glossary terms, data products, documents, applications, queries, incidents, structured properties, data contracts |
| `datahub_update` | Update Entity | 17 | Update descriptions, tags, glossary terms, links, owners, domains, structured properties, incidents, queries, document contents/status/related entities/sub-type, data contracts |
| `datahub_delete` | Delete Entity | 8 | Delete queries, tags, domains, glossary entities, data products, applications, documents, structured properties |

## Description Overrides

Tool descriptions can be customized at three levels of priority:

1. **Per-registration** (highest): `toolkit.RegisterWith(server, tools.ToolSearch, tools.WithDescription("custom"))`
2. **Toolkit-level**: `tools.NewToolkit(client, cfg, tools.WithDescriptions(map[tools.ToolName]string{...}))`
3. **Default**: Built-in descriptions from `pkg/tools/descriptions.go`

Descriptions can also be set via config file (`toolkit.descriptions` section) or the `Descriptions` field in `server.Options`.

## Title Overrides

Tool display names (shown in MCP clients) follow the same three-level priority:

1. **Per-registration** (highest): `toolkit.RegisterWith(server, tools.ToolSearch, tools.WithTitle("My Search"))`
2. **Toolkit-level**: `tools.NewToolkit(client, cfg, tools.WithTitles(map[tools.ToolName]string{...}))`
3. **Default**: Built-in titles from `pkg/tools/titles.go`

## OutputSchema Overrides

Tool output schemas (JSON Schema describing the response) follow the same three-level priority:

1. **Per-registration** (highest): `toolkit.RegisterWith(server, tools.ToolSearch, tools.WithOutputSchema(schema))`
2. **Toolkit-level**: `tools.NewToolkit(client, cfg, tools.WithOutputSchemas(map[tools.ToolName]any{...}))`
3. **Default**: Built-in schemas from `pkg/tools/output_schemas.go`

## Annotation Overrides

MCP tool annotations (behavior hints per the MCP specification) follow the same three-level priority:

1. **Per-registration** (highest): `toolkit.RegisterWith(server, tools.ToolSearch, tools.WithAnnotation(&mcp.ToolAnnotations{...}))`
2. **Toolkit-level**: `tools.NewToolkit(client, cfg, tools.WithAnnotations(map[tools.ToolName]*mcp.ToolAnnotations{...}))`
3. **Default**: Built-in annotations from `pkg/tools/annotations.go`

Default annotations for all 12 tools:

- **Read tools** (9): `ReadOnlyHint: true`, `IdempotentHint: true`, `OpenWorldHint: true`
- **datahub_create**: `DestructiveHint: false`, `IdempotentHint: false`, `OpenWorldHint: true`
- **datahub_update**: `DestructiveHint: false`, `IdempotentHint: true`, `OpenWorldHint: true`
- **datahub_delete**: `DestructiveHint: true`, `IdempotentHint: true`, `OpenWorldHint: true`

`OpenWorldHint: true` is correct because all tools communicate with an external DataHub instance.

## Extensions Package (`pkg/extensions/`)

Optional middleware and config file support. All extensions are opt-in.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_DATAHUB_EXT_LOGGING` | Enable structured logging of tool calls | `false` |
| `MCP_DATAHUB_EXT_METRICS` | Enable metrics collection | `false` |
| `MCP_DATAHUB_EXT_METADATA` | Enable metadata enrichment on results | `false` |
| `MCP_DATAHUB_EXT_ERRORS` | Enable error hint enrichment | `true` |

### Available Middleware

- **LoggingMiddleware**: Logs tool invocations and results to an `io.Writer`
- **MetricsMiddleware**: Collects call count, error count, and duration via `MetricsCollector` interface
- **ErrorHintMiddleware**: Appends helpful hints to error results (e.g., "Use datahub_search to find entities")
- **MetadataMiddleware**: Appends execution metadata footer (tool name, duration, timestamp) to results

### Config File Support

Load configuration from YAML or JSON files with `extensions.FromFile()` or `extensions.LoadConfig()`:

```yaml
datahub:
  url: https://datahub.example.com
  token: "${DATAHUB_TOKEN}"
  timeout: "30s"
  write_enabled: true

toolkit:
  default_limit: 20
  descriptions:
    datahub_search: "Custom search description for your org"

extensions:
  logging: true
  errors: true
```

Environment variables override file values for sensitive fields (`DATAHUB_URL`, `DATAHUB_TOKEN`, etc.).
Token values support `$VAR` / `${VAR}` expansion.

## Multi-Server Configuration

The reference implementation supports connecting to multiple DataHub instances:

```bash
# Primary server
export DATAHUB_URL=https://prod.datahub.example.com/api/graphql
export DATAHUB_TOKEN=prod-token
export DATAHUB_CONNECTION_NAME=prod

# Enable write operations (default: false)
export DATAHUB_WRITE_ENABLED=true

# Additional servers (JSON) - write_enabled is per-connection (nil=inherit)
export DATAHUB_ADDITIONAL_SERVERS='{"staging":{"url":"https://staging.datahub.example.com/api/graphql","token":"staging-token","write_enabled":false}}'
```

All tools accept an optional `connection` parameter to target a specific server.

## DataHub API Compatibility

**Schema source: `testdata/datahub-schema/` (synced from [datahub-project/datahub](https://github.com/datahub-project/datahub))**

**Current schema version: v1.5.0.1. Minimum supported: v1.3.x. Full feature set: v1.4.x+.**

### Version Compatibility Matrix

| DataHub Version | Features Available | Schema Validated |
|---|---|---|
| 1.3.x+ (minimum) | All read tools, all write operations except documents (tags, domains, glossary, data products, queries, owners, links, descriptions, incidents, applications, structured properties incl. delete, data contracts) | No (pre-dates schema sync) |
| 1.4.x+ (full) | + Documents (create/update/delete), semantic search | Yes (v1.4.0.3) |
| 1.5.x+ (current) | + Batch data product operations | Yes (v1.5.0.1) |

### Schema Validation

All GraphQL queries are validated against the upstream DataHub schema files:

```bash
make schema-sync               # Download schema for pinned version
make schema-check              # Validate all queries against schema
DATAHUB_VERSION=v1.5.0.1 make schema-sync  # Target a specific version
```

Schema files are checked into `testdata/datahub-schema/` and `make schema-check` runs as part of `make verify`. When adding or modifying GraphQL queries, **always check the upstream `.graphql` files first** — never guess field names or type structures.

### Graceful Degradation

The client handles variations across DataHub versions gracefully:
- Uses search fallback when `listDataProducts` query unavailable
- Returns empty results (not errors) when usage stats not configured
- Returns empty results (not errors) when incidents or structured properties are unavailable
- Parses properties from different response structures

## Verification (AI-Verified Development)

Run the full verification suite before every commit:
```
make verify
```

Individual checks (all must pass):
```
make lint            # golangci-lint (43 linters) + go vet
make test            # go test -race -shuffle=on ./...
make coverage        # Coverage report (threshold: 80%)
make patch-coverage  # Coverage of changed lines only (threshold: 80%)
make security        # gosec + govulncheck
make schema-check    # Validate GraphQL queries against upstream schema
make mutation        # gremlins (threshold: 60%)
make deadcode        # deadcode (unreachable functions)
make build-check     # go build + go mod verify
```

Performance diagnostics (not part of verify, use when investigating):
```
make bench           # Run benchmarks with memory allocation reporting
make profile         # Generate CPU and memory profiles for pprof
```

## Code Quality Thresholds

- Test coverage: >=80%
- Mutation score: >=60%
- Cyclomatic complexity: <=15 per function
- Cognitive complexity: <=15 per function
- Function length: <=80 lines, <=50 statements
- Function arguments: <=5
- Function return values: <=3

## Go Code Standards (AI-Verified)

1. **Error handling**: Always wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
2. **Naming**: Follow Go conventions. MixedCaps, not underscores. Acronyms are all-caps (HTTP, URL, ID).
3. **Interfaces**: Accept interfaces, return structs. Define interfaces at the consumer, not the provider.
4. **Context**: First parameter when needed. Never store in structs.
5. **Concurrency**: Use channels for communication, mutexes for state. Always run tests with `-race`.
6. **Dependencies**: Use `internal/` for code that shouldn't be imported. Minimize third-party dependencies.
7. **Testing**: Table-driven tests. Property-based tests for pure functions.

## AI-Specific Rules

1. **No tautological tests**: tests must encode expected outputs, not reimplement logic
2. **No hallucinated imports**: verify every dependency exists in the Go module ecosystem
3. **Human review required**: all code requires human review before merge
4. **Acceptance criteria first**: do not write code without Given/When/Then criteria
5. **Explain non-obvious decisions**: comment WHY, not WHAT
6. **No vaporware**: every package must be imported by non-test code
