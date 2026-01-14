# Available Tools

mcp-datahub provides the following MCP tools for interacting with DataHub.

## datahub_search

Search for datasets, dashboards, pipelines, and other assets.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query string |
| `entity_type` | string | No | Filter by entity type (DATASET, DASHBOARD, etc.) |
| `limit` | integer | No | Maximum results (default: 10) |
| `offset` | integer | No | Pagination offset |

**Example:**
```
Search for "customer" datasets
```

## datahub_get_entity

Get detailed metadata for an entity by URN.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | DataHub URN of the entity |

**Example:**
```
Get entity urn:li:dataset:(urn:li:dataPlatform:snowflake,db.schema.customers,PROD)
```

## datahub_get_schema

Get schema fields for a dataset with descriptions.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Dataset URN |

**Returns:** Schema metadata including field names, types, and descriptions.

## datahub_get_lineage

Get upstream and downstream lineage for an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `direction` | string | No | UPSTREAM, DOWNSTREAM, or BOTH (default: BOTH) |
| `depth` | integer | No | Maximum traversal depth (default: 3) |

## datahub_get_queries

Get SQL queries associated with a dataset.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Dataset URN |

**Returns:** List of SQL queries that reference this dataset.

## datahub_get_glossary_term

Get glossary term definition and related assets.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Glossary term URN |

## datahub_list_tags

List available tags in the catalog.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filter` | string | No | Filter tags by name pattern |

## datahub_list_domains

List data domains in the organization.

**Parameters:** None

**Returns:** List of all domains with their descriptions and entity counts.

## datahub_list_data_products

List all data products in the catalog.

**Parameters:** None

**Returns:** List of data products with their metadata.

## datahub_get_data_product

Get detailed information about a data product.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Data product URN |
