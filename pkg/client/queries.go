package client

// GraphQL query templates for DataHub operations.
const (
	// SearchQuery searches for entities.
	SearchQuery = `
query search($input: SearchInput!) {
  search(input: $input) {
    start
    count
    total
    searchResults {
      entity {
        urn
        type
        ... on Dataset {
          name
          description
          platform {
            name
          }
          ownership {
            owners {
              owner {
                ... on CorpUser {
                  urn
                  username
                }
                ... on CorpGroup {
                  urn
                  name
                }
              }
              type
            }
          }
          tags {
            tags {
              tag {
                urn
                name
                description
              }
            }
          }
          domain {
            domain {
              urn
              properties {
                name
                description
              }
            }
          }
        }
        ... on Dashboard {
          dashboardId
          info {
            name
            description
          }
          platform {
            name
          }
        }
        ... on DataFlow {
          flowId
          info {
            name
            description
          }
          platform {
            name
          }
        }
        ... on DataProduct {
          properties {
            name
            description
          }
        }
        ... on GlossaryTerm {
          properties {
            name
            description
          }
        }
        ... on Tag {
          properties {
            name
            description
          }
        }
      }
      matchedFields {
        name
        value
      }
    }
  }
}
`

	// GetEntityQuery retrieves a single entity by URN.
	GetEntityQuery = `
query getEntity($urn: String!) {
  entity(urn: $urn) {
    urn
    type
    ... on Dataset {
      name
      description
      platform {
        name
      }
      ownership {
        owners {
          owner {
            ... on CorpUser {
              urn
              username
              info {
                displayName
                email
              }
            }
            ... on CorpGroup {
              urn
              name
            }
          }
          type
        }
      }
      tags {
        tags {
          tag {
            urn
            name
            description
          }
        }
      }
      glossaryTerms {
        terms {
          term {
            urn
            properties {
              name
              description
            }
          }
        }
      }
      domain {
        domain {
          urn
          properties {
            name
            description
          }
        }
      }
      deprecation {
        deprecated
        note
        actor
        decommissionTime
      }
      properties {
        name
        description
        customProperties {
          key
          value
        }
      }
      subTypes {
        typeNames
      }
    }
    ... on Dashboard {
      dashboardId
      info {
        name
        description
        externalUrl
      }
      platform {
        name
      }
      ownership {
        owners {
          owner {
            ... on CorpUser {
              urn
              username
            }
          }
          type
        }
      }
    }
  }
}
`

	// GetSchemaQuery retrieves schema for a dataset.
	GetSchemaQuery = `
query getSchema($urn: String!) {
  dataset(urn: $urn) {
    schemaMetadata {
      name
      platformSchema {
        ... on TableSchema {
          schema
        }
      }
      version
      hash
      fields {
        fieldPath
        type
        nativeDataType
        description
        nullable
        isPartOfKey
        tags {
          tags {
            tag {
              urn
              name
            }
          }
        }
        glossaryTerms {
          terms {
            term {
              urn
              name
            }
          }
        }
      }
      primaryKeys
      foreignKeys {
        name
        sourceFields {
          fieldPath
        }
        foreignDataset {
          urn
        }
        foreignFields {
          fieldPath
        }
      }
    }
  }
}
`

	// GetLineageQuery retrieves lineage for an entity.
	// Note: maxHops parameter was removed from DataHub's SearchAcrossLineageInput.
	// Depth filtering should be done client-side using the returned degree values.
	GetLineageQuery = `
query getLineage($urn: String!, $direction: LineageDirection!) {
  searchAcrossLineage(
    input: {
      urn: $urn
      direction: $direction
    }
  ) {
    searchResults {
      entity {
        urn
        type
        ... on Dataset {
          name
          platform {
            name
          }
          description
        }
        ... on DataJob {
          jobId
          info {
            name
          }
          dataFlow {
            urn
            flowId
          }
        }
      }
      degree
      paths {
        path {
          urn
        }
      }
    }
  }
}
`

	// GetQueriesQuery retrieves queries for a dataset.
	GetQueriesQuery = `
query getQueries($urn: String!) {
  dataset(urn: $urn) {
    usageStats {
      buckets {
        bucket
        duration
        metrics {
          topSqlQueries
        }
      }
    }
  }
}
`

	// GetGlossaryTermQuery retrieves a glossary term.
	GetGlossaryTermQuery = `
query getGlossaryTerm($urn: String!) {
  glossaryTerm(urn: $urn) {
    urn
    name
    hierarchicalName
    properties {
      name
      description
      customProperties {
        key
        value
      }
    }
    parentNodes {
      nodes {
        urn
        properties {
          name
        }
      }
    }
    ownership {
      owners {
        owner {
          ... on CorpUser {
            urn
            username
          }
        }
        type
      }
    }
  }
}
`

	// ListTagsQuery lists all tags.
	ListTagsQuery = `
query listTags($input: SearchInput!) {
  search(input: $input) {
    total
    searchResults {
      entity {
        ... on Tag {
          urn
          name
          description
          properties {
            name
            description
          }
        }
      }
    }
  }
}
`

	// ListDomainsQuery lists all domains.
	ListDomainsQuery = `
query listDomains {
  listDomains(input: {start: 0, count: 100}) {
    total
    domains {
      urn
      properties {
        name
        description
      }
      ownership {
        owners {
          owner {
            ... on CorpUser {
              urn
              username
            }
          }
          type
        }
      }
      entities(input: {start: 0, count: 0}) {
        total
      }
    }
  }
}
`

	// PingQuery is a simple query to test connectivity.
	PingQuery = `
query ping {
  __typename
}
`

	// ListDataProductsQuery lists all data products.
	ListDataProductsQuery = `
query listDataProducts {
  listDataProducts(input: {start: 0, count: 100}) {
    total
    dataProducts {
      urn
      properties {
        name
        description
        customProperties {
          key
          value
        }
      }
      domain {
        domain {
          urn
          properties {
            name
          }
        }
      }
      ownership {
        owners {
          owner {
            ... on CorpUser {
              urn
              username
            }
            ... on CorpGroup {
              urn
              name
            }
          }
          type
        }
      }
    }
  }
}
`

	// GetDataProductQuery retrieves a single data product by URN.
	GetDataProductQuery = `
query getDataProduct($urn: String!) {
  dataProduct(urn: $urn) {
    urn
    properties {
      name
      description
      customProperties {
        key
        value
      }
    }
    domain {
      domain {
        urn
        properties {
          name
          description
        }
      }
    }
    ownership {
      owners {
        owner {
          ... on CorpUser {
            urn
            username
            info {
              displayName
              email
            }
          }
          ... on CorpGroup {
            urn
            name
          }
        }
        type
      }
    }
  }
}
`

	// GetColumnLineageQuery retrieves fine-grained column-level lineage for a dataset.
	GetColumnLineageQuery = `
query getColumnLineage($urn: String!) {
  dataset(urn: $urn) {
    fineGrainedLineages {
      upstreams {
        path
        dataset
      }
      downstreams {
        path
      }
      transformOperation
      confidenceScore
      query
    }
  }
}
`

	// BatchGetSchemasQuery retrieves schemas for multiple datasets by URN.
	BatchGetSchemasQuery = `
query batchGetSchemas($urns: [String!]!) {
  entities(urns: $urns) {
    ... on Dataset {
      urn
      schemaMetadata {
        name
        platformSchema {
          ... on TableSchema {
            schema
          }
        }
        version
        hash
        fields {
          fieldPath
          type
          nativeDataType
          description
          nullable
          isPartOfKey
          tags {
            tags {
              tag {
                urn
                name
              }
            }
          }
          glossaryTerms {
            terms {
              term {
                urn
                name
              }
            }
          }
        }
        primaryKeys
        foreignKeys {
          name
          sourceFields {
            fieldPath
          }
          foreignDataset {
            urn
          }
          foreignFields {
            fieldPath
          }
        }
      }
    }
  }
}
`
)
