package client

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestGraphQLQueriesMatchSchema validates that every GraphQL query/mutation
// constant in the client package uses types and fields that exist in the
// upstream DataHub GraphQL schema files (testdata/datahub-schema/*.graphql).
//
// This catches mismatched field names, incorrect type references, and
// invalid fragment targets before they reach production.
//
// Run `make schema-sync` to update the schema files when targeting a new
// DataHub version, then `make schema-check` to validate.
func TestGraphQLQueriesMatchSchema(t *testing.T) {
	schema := loadSchema(t)
	if len(schema.types) == 0 {
		t.Skip("no schema files found — run 'make schema-sync' first")
	}

	// Every GraphQL query/mutation constant in the client package.
	// Map key is the const name for readable error messages.
	queries := map[string]string{
		// Core queries (queries.go)
		"SearchQuery":                 SearchQuery,
		"GetEntityQuery":              GetEntityQuery,
		"GetSchemaQuery":              GetSchemaQuery,
		"GetLineageQuery":             GetLineageQuery,
		"GetQueriesQuery":             GetQueriesQuery,
		"GetUsageStatsQueriesQuery":   GetUsageStatsQueriesQuery,
		"GetGlossaryTermQuery":        GetGlossaryTermQuery,
		"ListTagsQuery":               ListTagsQuery,
		"ListDomainsQuery":            ListDomainsQuery,
		"PingQuery":                   PingQuery,
		"ListDataProductsQuery":       ListDataProductsQuery,
		"GetDataProductQuery":         GetDataProductQuery,
		"GetDataProductEntitiesQuery": GetDataProductEntitiesQuery,
		"GetColumnLineageQuery":       GetColumnLineageQuery,
		"BatchGetSchemasQuery":        BatchGetSchemasQuery,

		// Query CRUD mutations (queries.go)
		"CreateQueryMutation": CreateQueryMutation,
		"UpdateQueryMutation": UpdateQueryMutation,
		"DeleteQueryMutation": DeleteQueryMutation,

		// Documents (documents.go, context_documents.go)
		"GetDocumentQuery":         GetDocumentQuery,
		"GetRelatedDocumentsQuery": GetRelatedDocumentsQuery,
		"GetContextDocumentsQuery": GetContextDocumentsQuery,

		// Document mutations (write_documents.go)
		"updateDocumentContentsMutation":        updateDocumentContentsMutation,
		"updateDocumentStatusMutation":          updateDocumentStatusMutation,
		"updateDocumentRelatedEntitiesMutation": updateDocumentRelatedEntitiesMutation,
		"updateDocumentSubTypeMutation":         updateDocumentSubTypeMutation,
		"deleteDocumentMutation":                deleteDocumentMutation,

		// Structured properties (structured_properties.go)
		"GetStructuredPropertiesQuery":           GetStructuredPropertiesQuery,
		"ListStructuredPropertyDefinitionsQuery": ListStructuredPropertyDefinitionsQuery,
		"UpsertStructuredPropertiesMutation":     UpsertStructuredPropertiesMutation,
		"RemoveStructuredPropertiesMutation":     RemoveStructuredPropertiesMutation,

		// Data contracts (data_contracts.go)
		"GetDataContractQuery": GetDataContractQuery,

		// Incidents (incidents.go)
		"GetIncidentsQuery":            GetIncidentsQuery,
		"RaiseIncidentMutation":        RaiseIncidentMutation,
		"UpdateIncidentStatusMutation": UpdateIncidentStatusMutation,

		// Incident updates (write_incidents.go)
		"updateIncidentMutation": updateIncidentMutation,

		// Semantic search (semantic_search.go)
		"SemanticSearchQuery": SemanticSearchQuery,

		// Document search (search_documents.go)
		"SearchDocumentsQuery": SearchDocumentsQuery,

		// Tag/term/description mutations (write_graphql.go)
		"AddTagMutation":            AddTagMutation,
		"RemoveTagMutation":         RemoveTagMutation,
		"AddTermMutation":           AddTermMutation,
		"RemoveTermMutation":        RemoveTermMutation,
		"UpdateDescriptionMutation": UpdateDescriptionMutation,

		// Owner mutations (write_owners.go)
		"addOwnerMutation":    addOwnerMutation,
		"removeOwnerMutation": removeOwnerMutation,

		// Domain mutations (write_domains.go)
		"setDomainMutation":   setDomainMutation,
		"unsetDomainMutation": unsetDomainMutation,

		// Entity creation mutations (write_entities.go)
		"createTagMutation":                createTagMutation,
		"createDomainMutation":             createDomainMutation,
		"createGlossaryTermMutation":       createGlossaryTermMutation,
		"createDataProductMutation":        createDataProductMutation,
		"createDocumentMutation":           createDocumentMutation,
		"createApplicationMutation":        createApplicationMutation,
		"createStructuredPropertyMutation": createStructuredPropertyMutation,
		"upsertDataContractMutation":       upsertDataContractMutation,

		// Delete mutations (write_delete.go)
		"deleteTagMutation":                deleteTagMutation,
		"deleteDomainMutation":             deleteDomainMutation,
		"deleteGlossaryEntityMutation":     deleteGlossaryEntityMutation,
		"deleteDataProductMutation":        deleteDataProductMutation,
		"deleteApplicationMutation":        deleteApplicationMutation,
		"deleteStructuredPropertyMutation": deleteStructuredPropertyMutation,

		// Structured property CRUD (write_structured_properties_crud.go)
		"updateStructuredPropertyMutation": updateStructuredPropertyMutation,
	}

	for name, queryStr := range queries {
		t.Run(name, func(t *testing.T) {
			validateQuery(t, name, queryStr, schema)
		})
	}
}

// schemaData holds parsed type definitions from the GraphQL schema files.
type schemaData struct {
	// types maps type name → set of field names.
	// Covers both "type Foo" and "input Foo" definitions.
	types map[string]map[string]bool

	// fragmentTargets is the set of valid type names that fragments can target.
	fragmentTargets map[string]bool
}

// loadSchema reads all .graphql files from testdata/datahub-schema/ and
// parses type/input definitions with their fields.
func loadSchema(t *testing.T) schemaData {
	t.Helper()

	_, thisFile, _, _ := runtime.Caller(0)
	schemaDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "datahub-schema")

	files, err := filepath.Glob(filepath.Join(schemaDir, "*.graphql"))
	if err != nil || len(files) == 0 {
		return schemaData{}
	}

	sd := schemaData{
		types:           make(map[string]map[string]bool),
		fragmentTargets: make(map[string]bool),
	}

	// Matches: type Foo {, type Foo implements Bar {, input Foo {, extend type Foo {, enum Foo {
	typeRe := regexp.MustCompile(`(?:extend\s+)?(?:type|input|interface|enum)\s+(\w+)(?:\s+implements\s+\w+)?\s*\{`)
	// Matches field definitions: fieldName: Type or fieldName(args): Type
	fieldRe := regexp.MustCompile(`^\s+(\w+)\s*[\(:]`)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("reading schema file %s: %v", f, err)
		}

		lines := strings.Split(string(data), "\n")
		var currentType string
		braceDepth := 0

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Skip comments and directives
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "\"\"\"") || trimmed == "" {
				continue
			}

			// Check for type/input definition
			if m := typeRe.FindStringSubmatch(trimmed); m != nil {
				currentType = m[1]
				if sd.types[currentType] == nil {
					sd.types[currentType] = make(map[string]bool)
				}
				sd.fragmentTargets[currentType] = true
				braceDepth = 1
				continue
			}

			if currentType != "" {
				braceDepth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")

				if braceDepth <= 0 {
					currentType = ""
					braceDepth = 0
					continue
				}

				// Only parse top-level fields (brace depth 1 means we're inside the type body)
				if braceDepth == 1 {
					if fm := fieldRe.FindStringSubmatch(line); fm != nil {
						sd.types[currentType][fm[1]] = true
					}
				}
			}
		}
	}

	return sd
}

// validateQuery checks a GraphQL query string against the schema.
func validateQuery(t *testing.T, name, queryStr string, schema schemaData) {
	t.Helper()

	// Check fragment targets
	fragRe := regexp.MustCompile(`fragment\s+\w+\s+on\s+(\w+)`)
	for _, m := range fragRe.FindAllStringSubmatch(queryStr, -1) {
		typeName := m[1]
		if !schema.fragmentTargets[typeName] {
			t.Errorf("%s: fragment targets unknown type %q", name, typeName)
		}
	}

	// Check top-level query/mutation field names exist in Query/Mutation type
	topFieldRe := regexp.MustCompile(`(?:query|mutation)\s+\w+\([^)]*\)\s*\{\s*(\w+)`)
	for _, m := range topFieldRe.FindAllStringSubmatch(queryStr, -1) {
		field := m[1]
		inQuery := schema.types["Query"][field]
		inMutation := schema.types["Mutation"][field]
		if !inQuery && !inMutation {
			t.Errorf("%s: top-level field %q not found in Query or Mutation type", name, field)
		}
	}

	// Check inline fragment type names (... on TypeName)
	inlineFragRe := regexp.MustCompile(`\.\.\.\s+on\s+(\w+)`)
	for _, m := range inlineFragRe.FindAllStringSubmatch(queryStr, -1) {
		typeName := m[1]
		if !schema.fragmentTargets[typeName] {
			t.Errorf("%s: inline fragment targets unknown type %q", name, typeName)
		}
	}

	// Check input type references ($var: TypeName!)
	inputTypeRe := regexp.MustCompile(`\$\w+:\s*\[?(\w+)!?\]?!?`)
	for _, m := range inputTypeRe.FindAllStringSubmatch(queryStr, -1) {
		typeName := m[1]
		// Skip scalar types
		if isScalar(typeName) {
			continue
		}
		if !schema.fragmentTargets[typeName] {
			t.Errorf("%s: references unknown input type %q", name, typeName)
		}
	}
}

func isScalar(name string) bool {
	switch name {
	case "String", "Int", "Float", "Boolean", "ID", "Long":
		return true
	}
	return false
}
