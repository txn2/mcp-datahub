# Available Tools

mcp-datahub provides 12 MCP tools for interacting with DataHub (9 read + 3 write), covering 35 write operations through the CRUD discriminator pattern.

## DataHub Version Compatibility

**Minimum: DataHub 1.3.x. Full feature set: DataHub 1.4.x.**

All read tools and most write operations work with DataHub 1.3.x+. Only document operations require DataHub 1.4.x+:

| Operation | Tool | Requires |
|-----------|------|----------|
| `what=document` | `datahub_create` | DataHub 1.4.x+ |
| `what=document_contents` | `datahub_update` | DataHub 1.4.x+ |
| `what=document_status` | `datahub_update` | DataHub 1.4.x+ |
| `what=document_related_entities` | `datahub_update` | DataHub 1.4.x+ |
| `what=document_sub_type` | `datahub_update` | DataHub 1.4.x+ |
| `what=document` | `datahub_delete` | DataHub 1.4.x+ |

The client gracefully handles version differences for read queries — returning empty results (not errors) when a feature is unavailable.

## Tool Annotations

All tools include [MCP tool annotations](https://modelcontextprotocol.io/specification/2025-03-26/server/tools#annotations) that describe their behavior to AI clients:

| Hint | Read Tools | `datahub_create` | `datahub_update` | `datahub_delete` | Description |
|------|:----------:|:----------------:|:----------------:|:----------------:|-------------|
| `ReadOnlyHint` | `true` | `false` | `false` | `false` | Whether the tool only reads data |
| `DestructiveHint` | _(default)_ | `false` | `false` | `true` | Whether the tool may destructively update |
| `IdempotentHint` | `true` | `false` | `true` | `true` | Whether repeated calls produce the same result |
| `OpenWorldHint` | `true` | `true` | `true` | `true` | Whether the tool interacts with external entities |

`OpenWorldHint` is `true` for all tools because every tool communicates with an external DataHub instance.

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

Get upstream and downstream lineage for an entity. Supports both dataset-level and column-level lineage via the `level` parameter.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urn` | string | Yes | Entity URN |
| `level` | string | No | Granularity: `dataset` or `column` (default: `dataset`) |
| `direction` | string | No | UPSTREAM, DOWNSTREAM, or BOTH (default: BOTH, dataset level only) |
| `depth` | integer | No | Maximum traversal depth (default: 3, max: 5, dataset level only) |
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

### Column-Level Lineage (level=column)

When `level=column` is specified, returns fine-grained column-level lineage mappings instead of dataset-level lineage. The `direction` and `depth` parameters are ignored for column-level lineage.

**Example Request:**

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.analytics.customer_metrics,PROD)",
  "level": "column"
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

## datahub_browse

Browse the catalog to list tags, domains, or data products.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `what` | string | Yes | What to browse: `tags`, `domains`, or `data_products` |
| `filter` | string | No | Optional filter string (tags only) |
| `connection` | string | No | Named connection to use |

**Example Request (tags):**

```json
{
  "what": "tags"
}
```

**Example Response (tags):**

```json
{
  "tags": [
    {"name": "pii", "urn": "urn:li:tag:pii", "description": "Contains personally identifiable information"},
    {"name": "deprecated", "urn": "urn:li:tag:deprecated", "description": "This asset is deprecated"},
    {"name": "certified", "urn": "urn:li:tag:certified", "description": "Quality certified dataset"}
  ]
}
```

**Example Request (domains):**

```json
{
  "what": "domains"
}
```

**Example Response (domains):**

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
    }
  ]
}
```

**Example Request (data_products):**

```json
{
  "what": "data_products"
}
```

**Example Response (data_products):**

```json
{
  "data_products": [
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
  ]
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

Write tools require `DATAHUB_WRITE_ENABLED=true`. They use the CRUD discriminator pattern — 3 tools covering 35 operations via the `what` parameter.

---

### datahub_create

Create a new entity or resource. Returns the URN of the created entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `what` | string | Yes | Entity type: `tag`, `domain`, `glossary_term`, `data_product`, `document`, `application`, `query`, `incident`, `structured_property`, `data_contract` |
| `name` | string | Varies | Name or title (required for most types) |
| `description` | string | No | Description or content |
| `parent_node` | string | No | Parent glossary node URN (glossary_term) |
| `domain_urn` | string | No | Domain URN (data_product, required) |
| `value` | string | No | SQL statement (query) |
| `language` | string | No | Query language, default SQL (query) |
| `dataset_urns` | string[] | No | Associated dataset URNs (query, data_contract) |
| `entity_urns` | string[] | No | Affected entity URNs (incident) |
| `incident_type` | string | No | Incident type (incident) |
| `priority` | string | No | Priority: LOW, MEDIUM, HIGH, CRITICAL (incident) |
| `qualified_name` | string | No | Fully qualified name (structured_property) |
| `value_type` | string | No | Value type (structured_property) |
| `connection` | string | No | Named connection to use |

---

### datahub_update

Update metadata on an existing entity.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `what` | string | Yes | What to update (see table below) |
| `action` | string | No | `add`, `remove`, or `set` (default: `set`) |
| `urn` | string | Yes | Entity URN |
| `value` | string | No | New value (description, status, label, message) |
| `target_urn` | string | No | Target URN for add/remove (tag, glossary term, owner, domain) |
| `url` | string | No | URL for link operations |
| `field_path` | string | No | Schema field path (column_description) |
| `name` | string | No | Updated name (query, incident, structured_property) |
| `description` | string | No | Updated description |
| `title` | string | No | Document title (document_contents) |
| `text` | string | No | Document text (document_contents) |
| `entity_urns` | string[] | No | Related entity URNs (document_related_entities) |
| `connection` | string | No | Named connection to use |

**`what` values and required `action`:**

| what | action | Description |
|------|--------|-------------|
| `description` | set | Set entity description |
| `column_description` | set | Set schema field description |
| `tag` | add/remove | Add or remove a tag |
| `glossary_term` | add/remove | Add or remove a glossary term |
| `link` | add/remove | Add or remove a link |
| `owner` | add/remove | Add or remove an owner |
| `domain` | set/remove | Set or remove domain assignment |
| `structured_properties` | set/remove | Set or remove structured property values |
| `structured_property` | set | Update a structured property definition |
| `incident_status` | set | Resolve an incident |
| `incident` | set | Update incident details |
| `query` | set | Update query properties |
| `document_contents` | set | Update document title/text |
| `document_status` | set | Update document status |
| `document_related_entities` | set | Update document related entities |
| `document_sub_type` | set | Update document sub-type |
| `data_contract` | set | Upsert a data contract |

---

### datahub_delete

Delete an entity or resource. This is destructive and cannot be undone.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `what` | string | Yes | Entity type: `query`, `tag`, `domain`, `glossary_entity`, `data_product`, `application`, `document`, `structured_property` |
| `urn` | string | Yes | URN of the entity to delete |
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
