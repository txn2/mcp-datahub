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
└── integration/ # Extension interfaces for custom integrations

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

## Available Tools (11 total)

| Tool | Description |
|------|-------------|
| `datahub_search` | Search entities by query and type |
| `datahub_get_entity` | Get entity metadata by URN |
| `datahub_get_schema` | Get dataset schema |
| `datahub_get_lineage` | Get upstream/downstream lineage |
| `datahub_get_queries` | Get associated SQL queries |
| `datahub_get_glossary_term` | Get glossary term details |
| `datahub_list_tags` | List available tags |
| `datahub_list_domains` | List data domains |
| `datahub_list_data_products` | List data products |
| `datahub_get_data_product` | Get data product details |
| `datahub_list_connections` | List configured DataHub server connections |

## Multi-Server Configuration

The reference implementation supports connecting to multiple DataHub instances:

```bash
# Primary server
export DATAHUB_URL=https://prod.datahub.example.com/api/graphql
export DATAHUB_TOKEN=prod-token
export DATAHUB_CONNECTION_NAME=prod

# Additional servers (JSON)
export DATAHUB_ADDITIONAL_SERVERS='{"staging":{"url":"https://staging.datahub.example.com/api/graphql","token":"staging-token"}}'
```

All tools accept an optional `connection` parameter to target a specific server.

## DataHub API Compatibility

The client handles variations across DataHub versions gracefully:
- Uses search fallback when `listDataProducts` query unavailable
- Returns empty results (not errors) when usage stats not configured
- Parses properties from different response structures

When adding new queries, test against actual DataHub instances as GraphQL schemas vary between versions.
