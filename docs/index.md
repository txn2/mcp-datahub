---
hide:
  - toc
---

# txn2/mcp-datahub

An MCP server that connects AI assistants to DataHub metadata catalogs. Search datasets, explore schemas, understand lineage, and access business context like glossary terms and domains.

Unlike other MCP servers, mcp-datahub is designed as a composable Go library. Import it into your own MCP server to add DataHub capabilities with custom authentication, tenant isolation, and audit logging. The standalone server works out of the box; the library lets you build exactly what your organization needs.

[Get Started](server/installation.md){ .md-button .md-button--primary }
[View on GitHub](https://github.com/txn2/mcp-datahub){ .md-button }

---

## Two Ways to Use

<div class="grid cards" markdown>

-   :material-server:{ .lg .middle } **Use the Server**

    ---

    Connect Claude, Cursor, or any MCP client to DataHub with secure defaults.

    - Search across all assets
    - Schema exploration
    - Lineage visualization

    [:octicons-arrow-right-24: Install in 5 minutes](server/installation.md)

-   :material-code-braces:{ .lg .middle } **Build Custom MCP**

    ---

    Import the Go library for enterprise servers with auth, tenancy, and compliance.

    - OAuth, API keys, SSO
    - Row-level tenant isolation
    - SOC2 / HIPAA audit logs

    [:octicons-arrow-right-24: View library docs](library/index.md)

</div>

---

## Core Capabilities

<div class="grid cards" markdown>

-   :material-puzzle:{ .lg .middle } **Composable Architecture**

    ---

    Import as a Go library to build custom MCP servers with authentication,
    tenant isolation, and audit logging without forking.

    [:octicons-arrow-right-24: Library docs](library/index.md)

-   :material-database-search:{ .lg .middle } **Metadata Catalog**

    ---

    Access business descriptions, ownership, tags, domains, glossary terms,
    and data quality information from your DataHub instance.

    [:octicons-arrow-right-24: Tools reference](server/tools.md)

-   :material-graph:{ .lg .middle } **Lineage Exploration**

    ---

    Understand upstream and downstream dependencies for datasets,
    dashboards, and pipelines with configurable depth.

    [:octicons-arrow-right-24: Configuration](server/configuration.md)

-   :material-shield-check:{ .lg .middle } **Secure Defaults**

    ---

    Token-based authentication, read-only operations, and SLSA Level 3
    provenance for production deployments.

    [:octicons-arrow-right-24: Security reference](reference/security.md)

</div>

---

## Available Tools

| Tool | Description |
|------|-------------|
| `datahub_search` | Search across all DataHub assets |
| `datahub_get_entity` | Get entity metadata by URN |
| `datahub_get_schema` | Get dataset schema with field details |
| `datahub_get_lineage` | Explore upstream/downstream dependencies |
| `datahub_get_column_lineage` | Get fine-grained column-level lineage |
| `datahub_get_queries` | Get SQL queries associated with a dataset |
| `datahub_get_glossary_term` | Get term definition and relationships |
| `datahub_list_tags` | List available tags in the catalog |
| `datahub_list_domains` | List organizational domains |
| `datahub_list_data_products` | List data products in catalog |
| `datahub_get_data_product` | Get data product details and assets |
| `datahub_list_connections` | List configured server connections |

---

## Related Projects

Pair mcp-datahub with [txn2/mcp-trino](https://mcp-trino.txn2.com) for a complete data stack. mcp-trino queries your Trino data warehouse and can use DataHub as a semantic layer to enrich query results with business context.

---

## Works With

Claude Desktop · Claude Code · Cursor · Windsurf · Any MCP Client
