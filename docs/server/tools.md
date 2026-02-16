# Available Tools

mcp-datahub provides 19 MCP tools for interacting with DataHub (12 read + 7 write).

## Tool Annotations

All tools include [MCP tool annotations](https://modelcontextprotocol.io/specification/2025-03-26/server/tools#annotations) that describe their behavior to AI clients:

| Hint | Read Tools | Write Tools | Description |
|------|:----------:|:-----------:|-------------|
| `ReadOnlyHint` | `true` | `false` | Whether the tool only reads data |
| `DestructiveHint` | _(default)_ | `false` | Whether the tool may destructively update |
| `IdempotentHint` | `true` | `true` | Whether repeated calls produce the same result |
| `OpenWorldHint` | `false` | `false` | Whether the tool interacts with external entities |

These annotations help MCP clients make informed decisions about tool invocation (e.g., auto-approving read-only tools). Library users can override annotations per-tool or per-toolkit; see the [Tools API Reference](../reference/tools-api.md#withannotations).

## Multi-Server Support

All tools accept an optional `connection` parameter to target a specific DataHub server in multi-server environments. Use `datahub_list_connections` to discover available connections.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `connection` | string | No | Named connection to use (see `datahub_list_connections`) |

---

## datahub_list_connections

List all configured DataHub server connections.

**Parameters:** None

**Example Response:**

```json
{
  "connections": [
    {
      "name": "prod",
      "url": "https://prod.datahub.example.com",
      "is_default": true
    },
    {
      "name": "staging",
      "url": "https://staging.datahub.example.com",
      "is_default": false
    }
  ],
  "count": 2
}
```

**Use Cases:**

- Discover available connections before querying
- Verify multi-server configuration
- Check which connection is the default

---

## datahub_search

Search for datasets, dashboards, pipelines, and other assets.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query string |
| `entity_type` | string | No | Filter by entity type (DATASET, DASHBOARD, etc.) |
| `limit` | integer | No | Maximum results (default: 10, max: 100) |
| `offset` | integer | No | Pagination offset (default: 0) |
| `connection` | string | No | Named connection to use |

**Example Request:**

```json
{
  "query": "customer",
  "entity_type": "DATASET",
  "limit": 5
}
```

**Example Response:**

```json
{
  "entities": [
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
      "type": "DATASET",
      "name": "customers",
      "platform": "snowflake",
      "description": "Customer master data including contact information",
      "owners": ["Data Team"],
      "tags": ["pii", "customer-data"],
      "domain": "Sales"
    },
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customer_orders,PROD)",
      "type": "DATASET",
      "name": "customer_orders",
      "platform": "snowflake",
      "description": "Order history by customer"
    }
  ],
  "total": 42,
  "offset": 0,
  "limit": 5
}
```

**Common Use Cases:**

- Find datasets by name or description
- Search within a specific domain
- Discover dashboards related to a topic
- Find entities by tag

**Entity Type Values:**

| Value | Description |
|-------|-------------|
| `DATASET` | Tables, views, files |
| `DASHBOARD` | BI dashboards |
| `CHART` | Individual visualizations |
| `DATA_FLOW` | Pipelines |
| `DATA_JOB` | Pipeline tasks |
| `GLOSSARY_TERM` | Glossary terms |
| `DOMAIN` | Domains |
| `DATA_PRODUCT` | Data products |

---

## datahub_get_entity

Get detailed metadata for an entity by URN.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | DataHub URN of the entity |
| `connection` | string | No | Named connection to use |

**Example Request:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)"
}
```

**Example Response:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
  "type": "DATASET",
  "name": "customers",
  "qualifiedName": "prod.sales.customers",
  "description": "Customer master data including contact information and preferences",
  "platform": "snowflake",
  "owners": [
    {
      "urn": "urn:li:corpuser:jane.smith@company.com",
      "name": "Jane Smith",
      "type": "DATAOWNER"
    }
  ],
  "tags": [
    {"name": "pii", "urn": "urn:li:tag:pii"},
    {"name": "customer-data", "urn": "urn:li:tag:customer-data"}
  ],
  "glossaryTerms": [
    {"name": "Customer", "urn": "urn:li:glossaryTerm:Customer"},
    {"name": "PII", "urn": "urn:li:glossaryTerm:Classification.PII"}
  ],
  "domain": {
    "name": "Sales",
    "urn": "urn:li:domain:sales"
  },
  "created": "2023-06-15T10:30:00Z",
  "lastModified": "2024-01-10T14:22:00Z",
  "properties": {
    "customProperties": {
      "retention_days": "365",
      "data_classification": "confidential"
    }
  }
}
```

**Common Use Cases:**

- Get full details about a search result
- Find owners for a dataset
- Check tags and glossary terms
- Get custom properties

---

## datahub_get_schema

Get schema fields for a dataset with descriptions.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Dataset URN |
| `connection` | string | No | Named connection to use |

**Example Request:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)"
}
```

**Example Response:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
  "name": "customers",
  "fields": [
    {
      "fieldPath": "customer_id",
      "type": "NUMBER",
      "nativeType": "INT64",
      "description": "Unique customer identifier",
      "nullable": false,
      "isPrimaryKey": true
    },
    {
      "fieldPath": "email",
      "type": "STRING",
      "nativeType": "VARCHAR(255)",
      "description": "Customer email address",
      "nullable": true,
      "glossaryTerms": [
        {"name": "PII", "urn": "urn:li:glossaryTerm:Classification.PII"},
        {"name": "Email", "urn": "urn:li:glossaryTerm:ContactInfo.Email"}
      ]
    },
    {
      "fieldPath": "created_at",
      "type": "TIMESTAMP",
      "nativeType": "TIMESTAMP_NTZ",
      "description": "Account creation timestamp",
      "nullable": false
    },
    {
      "fieldPath": "address.street",
      "type": "STRING",
      "nativeType": "VARCHAR(500)",
      "description": "Street address",
      "nullable": true
    },
    {
      "fieldPath": "address.city",
      "type": "STRING",
      "nativeType": "VARCHAR(100)",
      "description": "City name",
      "nullable": true
    }
  ],
  "primaryKeys": ["customer_id"],
  "foreignKeys": []
}
```

**Field Properties:**

| Property | Description |
|----------|-------------|
| `fieldPath` | Full path including nested fields |
| `type` | Normalized type (STRING, NUMBER, etc.) |
| `nativeType` | Platform-specific type |
| `description` | Field description |
| `nullable` | Whether field can be null |
| `isPrimaryKey` | Whether field is a primary key |
| `glossaryTerms` | Associated glossary terms |

---

## datahub_get_lineage

Get upstream and downstream lineage for an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `direction` | string | No | UPSTREAM, DOWNSTREAM, or BOTH (default: BOTH) |
| `depth` | integer | No | Maximum traversal depth (default: 3, max: 5) |
| `connection` | string | No | Named connection to use |

**Example Request:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.analytics.customer_metrics,PROD)",
  "direction": "BOTH",
  "depth": 2
}
```

**Example Response:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.analytics.customer_metrics,PROD)",
  "upstream": [
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
      "name": "customers",
      "type": "DATASET",
      "platform": "snowflake",
      "degree": 1
    },
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.orders,PROD)",
      "name": "orders",
      "type": "DATASET",
      "platform": "snowflake",
      "degree": 1
    },
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.raw.customer_events,PROD)",
      "name": "customer_events",
      "type": "DATASET",
      "platform": "snowflake",
      "degree": 2
    }
  ],
  "downstream": [
    {
      "urn": "urn:li:dashboard:(looker,customer_360)",
      "name": "Customer 360 Dashboard",
      "type": "DASHBOARD",
      "platform": "looker",
      "degree": 1
    },
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.ml.churn_features,PROD)",
      "name": "churn_features",
      "type": "DATASET",
      "platform": "snowflake",
      "degree": 1
    }
  ]
}
```

**Common Use Cases:**

- Impact analysis before schema changes
- Root cause analysis for data issues
- Understanding data flow
- Discovering related datasets

---

## datahub_get_column_lineage

Get fine-grained column-level lineage mappings for a dataset.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Dataset URN |
| `connection` | string | No | Named connection to use |

**Example Request:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.analytics.customer_metrics,PROD)"
}
```

**Example Response:**

```json
{
  "dataset_urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.analytics.customer_metrics,PROD)",
  "mappings": [
    {
      "downstream_column": "customer_id",
      "upstream_dataset": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
      "upstream_column": "id",
      "transform": "IDENTITY"
    },
    {
      "downstream_column": "total_orders",
      "upstream_dataset": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.orders,PROD)",
      "upstream_column": "order_count",
      "transform": "AGGREGATE",
      "confidence_score": 0.95
    },
    {
      "downstream_column": "last_order_date",
      "upstream_dataset": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.orders,PROD)",
      "upstream_column": "order_date",
      "transform": "AGGREGATE"
    }
  ]
}
```

**Mapping Properties:**

| Property | Description |
|----------|-------------|
| `downstream_column` | Column name in the target dataset |
| `upstream_dataset` | URN of the source dataset |
| `upstream_column` | Column name in the source dataset |
| `transform` | Transformation type (IDENTITY, AGGREGATE, etc.) |
| `query` | Optional SQL query that defines the transformation |
| `confidence_score` | Optional confidence score (0-1) for inferred lineage |

**Common Use Cases:**

- Fine-grained impact analysis for column changes
- Understanding column-level data transformations
- Tracing data from source to derived columns
- Data quality root cause analysis at column level

---

## datahub_get_queries

Get SQL queries associated with a dataset.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Dataset URN |
| `connection` | string | No | Named connection to use |

**Example Response:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
  "queries": [
    {
      "query": "SELECT customer_id, email, created_at FROM prod.sales.customers WHERE created_at > DATEADD(day, -30, CURRENT_DATE())",
      "createdAt": "2024-01-10T09:15:00Z",
      "user": "analyst@company.com"
    },
    {
      "query": "SELECT COUNT(*) as total_customers FROM prod.sales.customers",
      "createdAt": "2024-01-09T14:30:00Z",
      "user": "dashboard_service"
    }
  ],
  "count": 2
}
```

---

## datahub_get_glossary_term

Get glossary term definition and related assets.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Glossary term URN |
| `connection` | string | No | Named connection to use |

**Example Request:**

```json
{
  "urn": "urn:li:glossaryTerm:Classification.PII"
}
```

**Example Response:**

```json
{
  "urn": "urn:li:glossaryTerm:Classification.PII",
  "name": "PII",
  "description": "Personally Identifiable Information - data that can identify an individual",
  "definition": "PII includes names, email addresses, phone numbers, social security numbers, and other data that can be used to identify a specific person.",
  "termSource": "INTERNAL",
  "parentNode": {
    "name": "Classification",
    "urn": "urn:li:glossaryNode:Classification"
  },
  "relatedTerms": [
    {"name": "Sensitive Data", "urn": "urn:li:glossaryTerm:Classification.Sensitive"},
    {"name": "PHI", "urn": "urn:li:glossaryTerm:Classification.PHI"}
  ],
  "owners": [
    {"name": "Data Governance Team", "type": "DATAOWNER"}
  ]
}
```

---

## datahub_list_tags

List available tags in the catalog.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filter` | string | No | Filter tags by name pattern |
| `connection` | string | No | Named connection to use |

**Example Response:**

```json
{
  "tags": [
    {"name": "pii", "urn": "urn:li:tag:pii", "description": "Contains personally identifiable information"},
    {"name": "deprecated", "urn": "urn:li:tag:deprecated", "description": "This asset is deprecated"},
    {"name": "certified", "urn": "urn:li:tag:certified", "description": "Quality certified dataset"},
    {"name": "sensitive", "urn": "urn:li:tag:sensitive", "description": "Contains sensitive data"}
  ],
  "count": 4
}
```

---

## datahub_list_domains

List data domains in the organization.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `connection` | string | No | Named connection to use |

**Example Response:**

```json
{
  "domains": [
    {
      "urn": "urn:li:domain:sales",
      "name": "Sales",
      "description": "Sales and revenue data",
      "entityCount": 45
    },
    {
      "urn": "urn:li:domain:marketing",
      "name": "Marketing",
      "description": "Marketing campaigns and analytics",
      "entityCount": 32
    },
    {
      "urn": "urn:li:domain:finance",
      "name": "Finance",
      "description": "Financial reporting and accounting",
      "entityCount": 28
    }
  ],
  "count": 3
}
```

---

## datahub_list_data_products

List all data products in the catalog.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `connection` | string | No | Named connection to use |

**Example Response:**

```json
{
  "dataProducts": [
    {
      "urn": "urn:li:dataProduct:customer-360",
      "name": "Customer 360",
      "description": "Unified view of customer data across all touchpoints",
      "domain": "Sales"
    },
    {
      "urn": "urn:li:dataProduct:revenue-analytics",
      "name": "Revenue Analytics",
      "description": "Revenue metrics and forecasting data",
      "domain": "Finance"
    }
  ],
  "count": 2
}
```

---

## datahub_get_data_product

Get detailed information about a data product.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Data product URN |
| `connection` | string | No | Named connection to use |

**Example Response:**

```json
{
  "urn": "urn:li:dataProduct:customer-360",
  "name": "Customer 360",
  "description": "Unified view of customer data across all touchpoints",
  "domain": {
    "name": "Sales",
    "urn": "urn:li:domain:sales"
  },
  "owners": [
    {"name": "Customer Data Team", "type": "DATAOWNER"}
  ],
  "assets": [
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.customer360.profile,PROD)",
      "name": "profile",
      "type": "DATASET"
    },
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.customer360.interactions,PROD)",
      "name": "interactions",
      "type": "DATASET"
    }
  ],
  "properties": {
    "sla": "99.9%",
    "refresh_frequency": "hourly"
  }
}
```

---

## Write Tools

Write tools require `DATAHUB_WRITE_ENABLED=true` to be set. They use DataHub's REST API (`POST /aspects?action=ingestProposal`) with read-modify-write semantics for array aspects (tags, terms, links).

---

### datahub_update_description

Update the description of an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `description` | string | Yes | New description text |
| `connection` | string | No | Named connection to use |

---

### datahub_add_tag

Add a tag to an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `tag_urn` | string | Yes | Tag URN to add |
| `connection` | string | No | Named connection to use |

---

### datahub_remove_tag

Remove a tag from an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `tag_urn` | string | Yes | Tag URN to remove |
| `connection` | string | No | Named connection to use |

---

### datahub_add_glossary_term

Add a glossary term to an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `term_urn` | string | Yes | Glossary term URN to add |
| `connection` | string | No | Named connection to use |

---

### datahub_remove_glossary_term

Remove a glossary term from an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `term_urn` | string | Yes | Glossary term URN to remove |
| `connection` | string | No | Named connection to use |

---

### datahub_add_link

Add a link to an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `link_url` | string | Yes | URL to add |
| `link_label` | string | Yes | Display label for the link |
| `connection` | string | No | Named connection to use |

---

### datahub_remove_link

Remove a link from an entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `link_url` | string | Yes | URL to remove |
| `connection` | string | No | Named connection to use |

---

## Error Responses

All tools may return error responses:

```json
{
  "error": true,
  "message": "Entity not found: urn:li:dataset:..."
}
```

**Common Errors:**

| Error | Cause | Solution |
|-------|-------|----------|
| `unauthorized` | Invalid or expired token | Generate new token |
| `entity not found` | URN does not exist | Verify URN is correct |
| `connection refused` | Cannot reach DataHub | Check DATAHUB_URL |
| `rate limit exceeded` | Too many requests | Reduce request rate |
| `invalid parameter` | Bad parameter value | Check parameter format |
