package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL queries and mutations for structured properties (DataHub 1.4.x+).
const (
	// GetStructuredPropertiesQuery reads structured property values assigned to an entity.
	GetStructuredPropertiesQuery = `
query getStructuredProperties($urn: String!) {
  entity(urn: $urn) {
    ... on Dataset {
      structuredProperties {
        properties {
          structuredProperty {
            urn
            definition {
              qualifiedName
              displayName
              description
              valueType {
                info {
                  type
                }
              }
              cardinality
              entityTypes {
                info {
                  type
                }
              }
              allowedValues {
                value {
                  ... on StringValue {
                    stringValue
                  }
                  ... on NumberValue {
                    numberValue
                  }
                }
                description
              }
            }
          }
          values {
            ... on StringValue {
              stringValue
            }
            ... on NumberValue {
              numberValue
            }
          }
        }
      }
    }
  }
}
`

	// ListStructuredPropertyDefinitionsQuery searches for all structured property definitions.
	ListStructuredPropertyDefinitionsQuery = `
query listStructuredPropertyDefinitions($input: SearchInput!) {
  search(input: $input) {
    total
    searchResults {
      entity {
        ... on StructuredPropertyEntity {
          urn
          definition {
            qualifiedName
            displayName
            description
            valueType {
              info {
                type
              }
            }
            cardinality
            entityTypes {
              info {
                type
              }
            }
            allowedValues {
              value {
                ... on StringValue {
                  stringValue
                }
                ... on NumberValue {
                  numberValue
                }
              }
              description
            }
          }
        }
      }
    }
  }
}
`

	// UpsertStructuredPropertiesMutation sets or updates structured property values on an entity.
	UpsertStructuredPropertiesMutation = `
mutation upsertStructuredProperties($input: UpsertStructuredPropertiesInput!) {
  upsertStructuredProperties(input: $input)
}
`

	// RemoveStructuredPropertiesMutation removes structured property values from an entity.
	RemoveStructuredPropertiesMutation = `
mutation removeStructuredProperties($input: RemoveStructuredPropertiesInput!) {
  removeStructuredProperties(input: $input)
}
`
)

// StructuredPropertyInput represents a structured property value to set on an entity.
type StructuredPropertyInput struct {
	// PropertyURN is the URN of the structured property definition.
	PropertyURN string

	// Values holds the value(s) to assign. Elements should be strings or numbers.
	Values []any
}

// GetStructuredProperties retrieves structured property values assigned to an entity.
// Returns empty results (not an error) when structured properties are not available,
// which is common on DataHub versions before 1.4.x.
func (c *Client) GetStructuredProperties(ctx context.Context, urn string) ([]types.StructuredPropertyValue, error) {
	variables := map[string]any{
		"urn": urn,
	}

	var response struct {
		Entity struct {
			StructuredProperties *struct {
				Properties []structuredPropertyEntry `json:"properties"`
			} `json:"structuredProperties"`
		} `json:"entity"`
	}

	if err := c.Execute(ctx, GetStructuredPropertiesQuery, variables, &response); err != nil {
		// Return empty result when structured properties are not supported (DataHub < 1.4.x)
		return nil, nil
	}

	if response.Entity.StructuredProperties == nil {
		return nil, nil
	}

	results := make([]types.StructuredPropertyValue, 0, len(response.Entity.StructuredProperties.Properties))
	for _, prop := range response.Entity.StructuredProperties.Properties {
		results = append(results, prop.toValue())
	}

	return results, nil
}

// ListStructuredPropertyDefinitions retrieves all structured property definitions.
// Returns empty results (not an error) when structured properties are not available,
// which is common on DataHub versions before 1.4.x.
func (c *Client) ListStructuredPropertyDefinitions(ctx context.Context) ([]types.StructuredPropertyDefinition, error) {
	variables := map[string]any{
		"input": map[string]any{
			"type":  "STRUCTURED_PROPERTY",
			"query": "*",
			"start": 0,
			"count": c.config.MaxLimit,
		},
	}

	var response struct {
		Search struct {
			Total         int `json:"total"`
			SearchResults []struct {
				Entity struct {
					URN        string                      `json:"urn"`
					Definition *structuredPropertyDefEntry `json:"definition"`
				} `json:"entity"`
			} `json:"searchResults"`
		} `json:"search"`
	}

	if err := c.Execute(ctx, ListStructuredPropertyDefinitionsQuery, variables, &response); err != nil {
		// Return empty result when STRUCTURED_PROPERTY type is not supported (DataHub < 1.4.x)
		return nil, nil
	}

	results := make([]types.StructuredPropertyDefinition, 0, len(response.Search.SearchResults))
	for _, sr := range response.Search.SearchResults {
		def := parseDefinition(sr.Entity.URN, sr.Entity.Definition)
		results = append(results, def)
	}

	return results, nil
}

// UpsertStructuredProperties sets or updates structured property values on an entity.
func (c *Client) UpsertStructuredProperties(ctx context.Context, urn string, properties []StructuredPropertyInput) error {
	propInputs := make([]map[string]any, 0, len(properties))
	for _, p := range properties {
		propInputs = append(propInputs, map[string]any{
			"propertyUrn": p.PropertyURN,
			"values":      p.Values,
		})
	}

	variables := map[string]any{
		"input": map[string]any{
			"assetUrn":                      urn,
			"structuredPropertyInputParams": propInputs,
		},
	}

	var response struct {
		UpsertStructuredProperties bool `json:"upsertStructuredProperties"`
	}

	if err := c.Execute(ctx, UpsertStructuredPropertiesMutation, variables, &response); err != nil {
		return fmt.Errorf("UpsertStructuredProperties: %w", err)
	}

	return nil
}

// RemoveStructuredProperties removes structured properties from an entity.
func (c *Client) RemoveStructuredProperties(ctx context.Context, urn string, propertyURNs []string) error {
	variables := map[string]any{
		"input": map[string]any{
			"assetUrn":     urn,
			"propertyUrns": propertyURNs,
		},
	}

	var response struct {
		RemoveStructuredProperties bool `json:"removeStructuredProperties"`
	}

	if err := c.Execute(ctx, RemoveStructuredPropertiesMutation, variables, &response); err != nil {
		return fmt.Errorf("RemoveStructuredProperties: %w", err)
	}

	return nil
}

// structuredPropertyEntry represents a single property+values pair from the GraphQL response.
type structuredPropertyEntry struct {
	StructuredProperty struct {
		URN        string                      `json:"urn"`
		Definition *structuredPropertyDefEntry `json:"definition"`
	} `json:"structuredProperty"`
	Values []propertyValueEntry `json:"values"`
}

// structuredPropertyDefEntry represents the definition portion of a GraphQL response.
type structuredPropertyDefEntry struct {
	QualifiedName string `json:"qualifiedName"`
	DisplayName   string `json:"displayName"`
	Description   string `json:"description"`
	ValueType     struct {
		Info struct {
			Type string `json:"type"`
		} `json:"info"`
	} `json:"valueType"`
	Cardinality string `json:"cardinality"`
	EntityTypes []struct {
		Info struct {
			Type string `json:"type"`
		} `json:"info"`
	} `json:"entityTypes"`
	AllowedValues []struct {
		Value       propertyValueEntry `json:"value"`
		Description string             `json:"description"`
	} `json:"allowedValues"`
}

// propertyValueEntry represents a union value from GraphQL (StringValue | NumberValue).
type propertyValueEntry struct {
	StringValue *string  `json:"stringValue"`
	NumberValue *float64 `json:"numberValue"`
}

// toAny converts the union value to a Go value.
func (v propertyValueEntry) toAny() any {
	if v.StringValue != nil {
		return *v.StringValue
	}
	if v.NumberValue != nil {
		return *v.NumberValue
	}
	return nil
}

// toValue converts a GraphQL response entry to the domain type.
func (e structuredPropertyEntry) toValue() types.StructuredPropertyValue {
	result := types.StructuredPropertyValue{
		PropertyURN: e.StructuredProperty.URN,
	}

	if e.StructuredProperty.Definition != nil {
		def := parseDefinition(e.StructuredProperty.URN, e.StructuredProperty.Definition)
		result.Definition = &def
	}

	for _, v := range e.Values {
		if val := v.toAny(); val != nil {
			result.Values = append(result.Values, val)
		}
	}

	return result
}

// parseDefinition converts a GraphQL definition entry to the domain type.
func parseDefinition(urn string, entry *structuredPropertyDefEntry) types.StructuredPropertyDefinition {
	def := types.StructuredPropertyDefinition{
		URN: urn,
	}

	if entry == nil {
		return def
	}

	def.QualifiedName = entry.QualifiedName
	def.DisplayName = entry.DisplayName
	def.Description = entry.Description
	def.ValueType = entry.ValueType.Info.Type
	def.Cardinality = entry.Cardinality

	for _, et := range entry.EntityTypes {
		if et.Info.Type != "" {
			def.EntityTypes = append(def.EntityTypes, et.Info.Type)
		}
	}

	for _, av := range entry.AllowedValues {
		val := av.Value.toAny()
		if val == nil {
			continue
		}
		def.AllowedValues = append(def.AllowedValues, types.AllowedValue{
			Value:       fmt.Sprintf("%v", val),
			Description: av.Description,
		})
	}

	return def
}
