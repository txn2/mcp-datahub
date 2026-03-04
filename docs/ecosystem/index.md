---
hide:
  - navigation
  - title
---

# MCP Data Platform Ecosystem

mcp-datahub is part of a broader suite of open-source MCP servers designed to work together as a composable data platform. Each component can run standalone or be combined to give AI assistants unified access to storage, query engines, and metadata catalogs.

---

## [mcp-data-platform](https://github.com/txn2/mcp-data-platform/)

The orchestration layer that ties the ecosystem together. mcp-data-platform provides a single MCP server that coordinates access across S3 storage, Trino query engines, and DataHub metadata catalogs. Rather than configuring each MCP server independently, mcp-data-platform presents a unified interface where AI assistants can discover datasets through the catalog, query them through Trino, and access the underlying files in S3, all from one connection. It handles connection routing, credential management, and cross-service context so that assistants can work with data end-to-end without switching between tools.

## [mcp-s3](https://github.com/txn2/mcp-s3/)

An MCP server for [Amazon S3](https://aws.amazon.com/s3/), providing AI assistants with direct access to object storage. mcp-s3 lets assistants list buckets, browse prefixes, read objects, and generate presigned URLs for temporary access. When paired with mcp-datahub and mcp-trino, it provides the raw file access layer: assistants can discover datasets through the catalog, query structured data through Trino, and retrieve or inspect the underlying files in S3. It supports multi-server configurations for accessing storage across accounts and regions.

## [mcp-trino](https://github.com/txn2/mcp-trino/)

An MCP server for [Trino](https://trino.io/), the distributed SQL query engine. mcp-trino enables AI assistants to run read-only SQL queries across any data source that Trino connects to, including data lakes, warehouses, and relational databases. Assistants can list catalogs and schemas, describe tables, explain query plans, and execute analytical queries with configurable timeouts and row limits. Combined with mcp-datahub for discovery and mcp-s3 for raw file access, mcp-trino completes the platform by providing the structured query interface that turns raw data into answers.
