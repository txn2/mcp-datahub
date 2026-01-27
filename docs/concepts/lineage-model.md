# Lineage Model

How DataHub tracks data flow and dependencies.

## What is Data Lineage?

Data lineage tracks the origin, movement, and transformation of data through systems. It answers:

- Where does this data come from?
- What systems use this data?
- How was this data transformed?
- What is the impact of changing this data?

## Lineage Directions

```mermaid
flowchart LR
    subgraph Upstream
        A[Source A]
        B[Source B]
    end

    A --> C[Your Dataset]
    B --> C

    subgraph Downstream
        D[Dashboard]
        E[ML Model]
        F[Report]
    end

    C --> D
    C --> E
    C --> F
```

### Upstream Lineage

Entities that feed data INTO the target entity.

- Raw data sources
- Staging tables
- Parent datasets in transformations

### Downstream Lineage

Entities that consume data FROM the target entity.

- Dashboards and reports
- Derived datasets
- ML models
- Other consumers

## Lineage Depth

Lineage can be traversed to different depths:

```mermaid
flowchart TB
    subgraph Depth 3
        L3A[Raw Events]
        L3B[External API]
    end

    subgraph Depth 2
        L2A[Staging Events]
        L2B[Staging Users]
    end

    subgraph Depth 1
        L1A[User Activity]
        L1B[User Profiles]
    end

    subgraph Target
        T[Analytics Dataset]
    end

    L3A --> L2A
    L3B --> L2B
    L2A --> L1A
    L2B --> L1B
    L1A --> T
    L1B --> T
```

| Depth | Description |
|-------|-------------|
| 1 | Direct dependencies only |
| 2 | Two hops away |
| 3 | Three hops away |
| N | N hops away |

Higher depth reveals more context but increases query time.

## Lineage Types

### Column-Level Lineage

Tracks which columns derive from which source columns.

```mermaid
flowchart LR
    subgraph Source Table
        s1[customer_id]
        s2[first_name]
        s3[last_name]
    end

    subgraph Target Table
        t1[id]
        t2[full_name]
    end

    s1 --> t1
    s2 --> t2
    s3 --> t2
```

This shows that:

- `id` comes from `customer_id`
- `full_name` is derived from `first_name` and `last_name`

### Table-Level Lineage

Tracks which tables are used to create other tables.

```mermaid
flowchart LR
    customers[customers] --> customer_orders[customer_orders]
    orders[orders] --> customer_orders
    customer_orders --> revenue_report[revenue_report]
```

### Cross-Platform Lineage

Tracks data across different systems.

```mermaid
flowchart LR
    subgraph Postgres
        pg[raw_events]
    end

    subgraph Snowflake
        sf_stg[stg_events]
        sf_mart[mart_events]
    end

    subgraph Looker
        dash[Events Dashboard]
    end

    pg --> sf_stg --> sf_mart --> dash
```

## Lineage Sources

DataHub collects lineage from multiple sources:

| Source | Method |
|--------|--------|
| SQL parsing | Parse CREATE TABLE AS, INSERT INTO |
| dbt | dbt manifest and catalog files |
| Airflow | Task dependencies |
| Spark | Spark lineage events |
| Great Expectations | Validation dependencies |
| Manual | UI or API annotations |

## Querying Lineage

### Table-Level Lineage Query

```
datahub_get_lineage urn="urn:li:dataset:..." direction="BOTH" depth=2
```

### Column-Level Lineage Query

```
datahub_get_column_lineage urn="urn:li:dataset:..."
```

Column-level lineage returns fine-grained mappings:

```json
{
  "dataset_urn": "urn:li:dataset:(platform,db.schema.target,PROD)",
  "mappings": [
    {
      "downstream_column": "id",
      "upstream_dataset": "urn:li:dataset:(platform,db.schema.source,PROD)",
      "upstream_column": "customer_id",
      "transform": "IDENTITY"
    },
    {
      "downstream_column": "full_name",
      "upstream_dataset": "urn:li:dataset:(platform,db.schema.source,PROD)",
      "upstream_column": "first_name",
      "transform": "TRANSFORM",
      "confidence_score": 0.9
    }
  ]
}
```

**Transform Types:**

| Transform | Description |
|-----------|-------------|
| IDENTITY | Column copied directly |
| AGGREGATE | Column derived from aggregation |
| TRANSFORM | Column has been computed or transformed |

### Response Structure

```json
{
  "urn": "urn:li:dataset:(platform,db.schema.target,PROD)",
  "upstream": [
    {
      "urn": "urn:li:dataset:(platform,db.schema.source1,PROD)",
      "name": "source1",
      "type": "DATASET",
      "platform": "snowflake",
      "degree": 1
    },
    {
      "urn": "urn:li:dataset:(platform,db.schema.source2,PROD)",
      "name": "source2",
      "type": "DATASET",
      "platform": "snowflake",
      "degree": 1
    }
  ],
  "downstream": [
    {
      "urn": "urn:li:dashboard:(looker,dashboard_id)",
      "name": "Sales Dashboard",
      "type": "DASHBOARD",
      "platform": "looker",
      "degree": 1
    }
  ]
}
```

### Understanding Degree

The `degree` field indicates distance from the target:

- Degree 1: Direct dependency
- Degree 2: One hop away
- Degree 3: Two hops away

## Use Cases

### Impact Analysis

Before changing a table, understand what depends on it:

```mermaid
flowchart TB
    change[Proposed Change]
    target[customers table]
    d1[Dashboard A]
    d2[Dashboard B]
    d3[ML Model]
    d4[Report Generator]

    change --> target
    target --> d1
    target --> d2
    target --> d3
    target --> d4

    style change fill:#ff9999
    style d1 fill:#ffcc99
    style d2 fill:#ffcc99
    style d3 fill:#ffcc99
    style d4 fill:#ffcc99
```

Questions to answer:

- What dashboards will break?
- What ML models need retraining?
- What reports need updating?

### Root Cause Analysis

When data quality issues occur, trace back to the source:

```mermaid
flowchart TB
    problem[Bad Data in Report]
    l1[mart_revenue]
    l2[stg_orders]
    l3[raw_orders]
    root[Source System Bug]

    problem --> l1
    l1 --> l2
    l2 --> l3
    l3 --> root

    style problem fill:#ff9999
    style root fill:#ff9999
```

### Compliance and Governance

Track sensitive data through the pipeline:

```mermaid
flowchart LR
    pii[PII Source]
    transform[Transform]
    anonymized[Anonymized Data]
    reporting[Reporting]

    pii -->|Contains SSN| transform
    transform -->|SSN Masked| anonymized
    anonymized --> reporting

    style pii fill:#ffcc99
```

## Lineage Limitations

### Not Always Complete

- Some transformations are opaque (stored procedures)
- Manual data movements are not tracked
- Some platforms do not emit lineage events

### Point-in-Time

- Lineage represents current state
- Historical lineage may not be available
- Schema changes can break lineage

### Performance Considerations

- Deep lineage queries can be slow
- Large lineage graphs consume memory
- Consider caching for repeated queries

## Best Practices

### Start with Depth 1

Begin with direct dependencies, then expand:

```
datahub_get_lineage urn="..." depth=1
```

### Filter by Direction

If you only need upstream or downstream:

```
datahub_get_lineage urn="..." direction="UPSTREAM"
```

### Use for Specific Questions

Instead of fetching everything, query for specific needs:

- Impact analysis: downstream depth 2-3
- Root cause: upstream depth 3-5
- Direct dependencies: depth 1

## Related Topics

- [Tutorial: Exploring Lineage](../tutorials/exploring-lineage.md): Hands-on lineage exploration
- [Entity Types](entity-types.md): What entities have lineage
- [Tools Reference](../server/tools.md): Lineage tool parameters
