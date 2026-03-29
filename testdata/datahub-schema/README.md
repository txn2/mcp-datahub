# DataHub GraphQL Schema Files

This directory contains GraphQL schema files downloaded from the upstream
[DataHub repository](https://github.com/datahub-project/datahub). These
files are the **source of truth** for validating that our GraphQL queries
match the actual DataHub API.

## Current Version

See `VERSION` file. Updated by `make schema-sync`.

## Usage

```bash
# Sync schema files for the current pinned version
make schema-sync

# Sync schema files for a specific version
DATAHUB_VERSION=v1.5.0.1 make schema-sync

# Validate all queries against the schema
make schema-check
```

## How It Works

1. `sync.sh` downloads `.graphql` files from the DataHub repo at the
   specified git tag
2. Files are checked into this repo so CI runs without network access
3. `TestGraphQLQueriesMatchSchema` in `pkg/client/` validates every
   query/mutation constant against these schema files
4. `make schema-check` is part of `make verify`

## When to Update

- When targeting a new DataHub version
- When adding new queries that reference new types
- When DataHub releases a version with schema changes

## Source

Files are from:
```
https://github.com/datahub-project/datahub/tree/<version>/datahub-graphql-core/src/main/resources/
```
