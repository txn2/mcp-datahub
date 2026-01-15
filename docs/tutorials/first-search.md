# Tutorial: Your First DataHub Search

Learn how to search your DataHub catalog using an AI assistant.

**Prerequisites**:

- mcp-datahub installed ([Installation Guide](../server/installation.md))
- Environment configured with DATAHUB_URL and DATAHUB_TOKEN
- Claude Desktop, Claude Code, or another MCP client

## What You Will Learn

- How to perform basic searches
- How to filter by entity type
- How to explore search results
- How to get detailed entity information

## Step 1: Verify Your Connection

Start your MCP client and verify mcp-datahub is connected.

Ask your AI assistant:

> "List the available DataHub connections"

You should see a response showing your configured connection:

```json
{
  "connections": [
    {
      "name": "datahub",
      "url": "https://your-datahub.example.com",
      "is_default": true
    }
  ]
}
```

If you see an error, check your configuration in the [Troubleshooting Guide](../support/troubleshooting.md).

## Step 2: Basic Search

Now perform your first search. Ask:

> "Search DataHub for customer"

The AI will use the `datahub_search` tool and return results like:

```json
{
  "entities": [
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
      "type": "DATASET",
      "name": "customers",
      "platform": "snowflake",
      "description": "Customer master data including contact info and preferences"
    },
    {
      "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customer_orders,PROD)",
      "type": "DATASET",
      "name": "customer_orders",
      "platform": "snowflake"
    }
  ],
  "total": 15
}
```

**Understanding the Results**

- `urn`: The unique identifier for this entity in DataHub
- `type`: The kind of entity (DATASET, DASHBOARD, etc.)
- `name`: The human-readable name
- `platform`: The data platform (Snowflake, BigQuery, etc.)
- `description`: Business description if available
- `total`: Total matching results (may exceed returned count)

## Step 3: Filter by Entity Type

Narrow your search to specific entity types. Ask:

> "Search for customer dashboards in DataHub"

The AI will add an entity type filter:

```json
{
  "entities": [
    {
      "urn": "urn:li:dashboard:(looker,customer_360)",
      "type": "DASHBOARD",
      "name": "Customer 360 Dashboard",
      "platform": "looker",
      "description": "Unified view of customer metrics"
    }
  ],
  "total": 3
}
```

**Available Entity Types**

- `DATASET`: Tables, views, files
- `DASHBOARD`: BI dashboards
- `CHART`: Individual visualizations
- `DATA_FLOW`: Pipelines and workflows
- `DATA_JOB`: Individual pipeline tasks

## Step 4: Get Entity Details

Pick an entity from your search results and ask for details. Use the URN from the results:

> "Get details for the customers dataset in DataHub"

Or be specific with the URN:

> "Get the DataHub entity urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)"

You will see detailed metadata:

```json
{
  "urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
  "name": "customers",
  "description": "Customer master data including contact info and preferences",
  "platform": "snowflake",
  "owners": [
    {
      "name": "Data Team",
      "type": "DATAOWNER"
    }
  ],
  "tags": ["pii", "customer-data"],
  "glossaryTerms": ["Customer", "PII"],
  "domain": "Sales"
}
```

## Step 5: Explore the Schema

For datasets, you can explore the schema. Ask:

> "What fields are in the customers table?"

The AI uses `datahub_get_schema`:

```json
{
  "fields": [
    {
      "fieldPath": "customer_id",
      "type": "NUMBER",
      "description": "Unique customer identifier",
      "nullable": false
    },
    {
      "fieldPath": "email",
      "type": "STRING",
      "description": "Customer email address",
      "nullable": true,
      "glossaryTerms": ["PII", "Email"]
    },
    {
      "fieldPath": "created_at",
      "type": "TIMESTAMP",
      "description": "Account creation timestamp"
    }
  ]
}
```

## Step 6: Paginate Results

For searches with many results, use pagination. Ask:

> "Search for datasets in DataHub, show results 11-20"

The AI will use offset and limit parameters to paginate.

## Practice Exercises

Try these searches on your own:

1. Search for all datasets from a specific platform (e.g., "BigQuery datasets")
2. Find dashboards related to "revenue" or "sales"
3. Search for glossary terms containing "customer"
4. Get the schema for a dataset you found

## What You Learned

- Basic search with `datahub_search`
- Filtering by entity type
- Getting entity details with `datahub_get_entity`
- Exploring schemas with `datahub_get_schema`
- Understanding URNs and search results

## Next Steps

- [Exploring Data Lineage](exploring-lineage.md): Trace data dependencies
- [Available Tools Reference](../server/tools.md): All tool documentation
- [Understanding URNs](../concepts/urns.md): Deep dive into URN structure
