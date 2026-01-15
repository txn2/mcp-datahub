# Available Tools

mcp-datahub provides the following MCP tools for interacting with DataHub.

## Multi-Server Support

All tools accept an optional `connection` parameter to target a specific DataHub server in multi-server environments. Use `datahub_list_connections` to discover available connections.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `connection` | string | No | Named connection to use (see `datahub_list_connections`) |

---

## datahub_list_connections

List all configured DataHub server connections. Use this to discover available connections before querying specific servers.

**Parameters:** None

**Returns:** List of connections with name, URL, and default status.

**Example Response:**
```json
{
  "connections": [
    {"name": "prod", "url": "https://prod.datahub.example.com", "is_default": true},
    {"name": "staging", "url": "https://staging.datahub.example.com", "is_default": false}
  ],
  "count": 2
}
```

---

## datahub_search

Search for datasets, dashboards, pipelines, and other assets.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query string |
| `entity_type` | string | No | Filter by entity type (DATASET, DASHBOARD, etc.) |
| `limit` | integer | No | Maximum results (default: 10) |
| `offset` | integer | No | Pagination offset |
| `connection` | string | No | Named connection to use |

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
| `connection` | string | No | Named connection to use |

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
| `connection` | string | No | Named connection to use |

**Returns:** Schema metadata including field names, types, and descriptions.

## datahub_get_lineage

Get upstream and downstream lineage for an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `direction` | string | No | UPSTREAM, DOWNSTREAM, or BOTH (default: BOTH) |
| `depth` | integer | No | Maximum traversal depth (default: 3) |
| `connection` | string | No | Named connection to use |

## datahub_get_queries

Get SQL queries associated with a dataset.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Dataset URN |
| `connection` | string | No | Named connection to use |

**Returns:** List of SQL queries that reference this dataset.

## datahub_get_glossary_term

Get glossary term definition and related assets.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Glossary term URN |
| `connection` | string | No | Named connection to use |

## datahub_list_tags

List available tags in the catalog.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filter` | string | No | Filter tags by name pattern |
| `connection` | string | No | Named connection to use |

## datahub_list_domains

List data domains in the organization.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `connection` | string | No | Named connection to use |

**Returns:** List of all domains with their descriptions and entity counts.

## datahub_list_data_products

List all data products in the catalog.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `connection` | string | No | Named connection to use |

**Returns:** List of data products with their metadata.

## datahub_get_data_product

Get detailed information about a data product.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Data product URN |
| `connection` | string | No | Named connection to use |
