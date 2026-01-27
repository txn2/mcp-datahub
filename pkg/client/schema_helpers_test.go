package client

import (
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestParseSchemaField(t *testing.T) {
	tests := []struct {
		name     string
		input    rawSchemaField
		expected types.SchemaField
	}{
		{
			name: "basic field",
			input: rawSchemaField{
				FieldPath:      "customer_id",
				Type:           "NUMBER",
				NativeDataType: "INT64",
				Description:    "Customer identifier",
				Nullable:       false,
				IsPartOfKey:    true,
			},
			expected: types.SchemaField{
				FieldPath:      "customer_id",
				Type:           "NUMBER",
				NativeType:     "INT64",
				Description:    "Customer identifier",
				Nullable:       false,
				IsPartitionKey: true,
			},
		},
		{
			name: "field with tags",
			input: rawSchemaField{
				FieldPath: "email",
				Type:      "STRING",
				Tags: struct {
					Tags []struct {
						Tag struct {
							URN  string `json:"urn"`
							Name string `json:"name"`
						} `json:"tag"`
					} `json:"tags"`
				}{
					Tags: []struct {
						Tag struct {
							URN  string `json:"urn"`
							Name string `json:"name"`
						} `json:"tag"`
					}{
						{Tag: struct {
							URN  string `json:"urn"`
							Name string `json:"name"`
						}{URN: "urn:li:tag:pii", Name: "pii"}},
						{Tag: struct {
							URN  string `json:"urn"`
							Name string `json:"name"`
						}{URN: "urn:li:tag:sensitive", Name: "sensitive"}},
					},
				},
			},
			expected: types.SchemaField{
				FieldPath: "email",
				Type:      "STRING",
				Tags: []types.Tag{
					{URN: "urn:li:tag:pii", Name: "pii"},
					{URN: "urn:li:tag:sensitive", Name: "sensitive"},
				},
			},
		},
		{
			name: "field with glossary terms",
			input: rawSchemaField{
				FieldPath: "revenue",
				Type:      "NUMBER",
				GlossaryTerms: struct {
					Terms []struct {
						Term struct {
							URN  string `json:"urn"`
							Name string `json:"name"`
						} `json:"term"`
					} `json:"terms"`
				}{
					Terms: []struct {
						Term struct {
							URN  string `json:"urn"`
							Name string `json:"name"`
						} `json:"term"`
					}{
						{Term: struct {
							URN  string `json:"urn"`
							Name string `json:"name"`
						}{URN: "urn:li:glossaryTerm:Finance.Revenue", Name: "Revenue"}},
					},
				},
			},
			expected: types.SchemaField{
				FieldPath: "revenue",
				Type:      "NUMBER",
				GlossaryTerms: []types.GlossaryTerm{
					{URN: "urn:li:glossaryTerm:Finance.Revenue", Name: "Revenue"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseSchemaField(tc.input)

			if result.FieldPath != tc.expected.FieldPath {
				t.Errorf("FieldPath = %s, want %s", result.FieldPath, tc.expected.FieldPath)
			}
			if result.Type != tc.expected.Type {
				t.Errorf("Type = %s, want %s", result.Type, tc.expected.Type)
			}
			if result.NativeType != tc.expected.NativeType {
				t.Errorf("NativeType = %s, want %s", result.NativeType, tc.expected.NativeType)
			}
			if result.Nullable != tc.expected.Nullable {
				t.Errorf("Nullable = %v, want %v", result.Nullable, tc.expected.Nullable)
			}
			if result.IsPartitionKey != tc.expected.IsPartitionKey {
				t.Errorf("IsPartitionKey = %v, want %v", result.IsPartitionKey, tc.expected.IsPartitionKey)
			}
			if len(result.Tags) != len(tc.expected.Tags) {
				t.Errorf("Tags count = %d, want %d", len(result.Tags), len(tc.expected.Tags))
			}
			for i, tag := range result.Tags {
				if tag.URN != tc.expected.Tags[i].URN || tag.Name != tc.expected.Tags[i].Name {
					t.Errorf("Tag[%d] = %+v, want %+v", i, tag, tc.expected.Tags[i])
				}
			}
			if len(result.GlossaryTerms) != len(tc.expected.GlossaryTerms) {
				t.Errorf("GlossaryTerms count = %d, want %d", len(result.GlossaryTerms), len(tc.expected.GlossaryTerms))
			}
			for i, term := range result.GlossaryTerms {
				if term.URN != tc.expected.GlossaryTerms[i].URN || term.Name != tc.expected.GlossaryTerms[i].Name {
					t.Errorf("GlossaryTerm[%d] = %+v, want %+v", i, term, tc.expected.GlossaryTerms[i])
				}
			}
		})
	}
}

func TestParseForeignKey(t *testing.T) {
	tests := []struct {
		name     string
		input    rawForeignKey
		expected types.ForeignKey
	}{
		{
			name: "basic foreign key",
			input: rawForeignKey{
				Name: "fk_customer",
				SourceFields: []struct {
					FieldPath string `json:"fieldPath"`
				}{
					{FieldPath: "customer_id"},
				},
				ForeignDataset: struct {
					URN string `json:"urn"`
				}{URN: "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)"},
				ForeignFields: []struct {
					FieldPath string `json:"fieldPath"`
				}{
					{FieldPath: "id"},
				},
			},
			expected: types.ForeignKey{
				Name:           "fk_customer",
				SourceFields:   []string{"customer_id"},
				ForeignDataset: "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.customers,PROD)",
				ForeignFields:  []string{"id"},
			},
		},
		{
			name: "composite foreign key",
			input: rawForeignKey{
				Name: "fk_order_item",
				SourceFields: []struct {
					FieldPath string `json:"fieldPath"`
				}{
					{FieldPath: "order_id"},
					{FieldPath: "item_id"},
				},
				ForeignDataset: struct {
					URN string `json:"urn"`
				}{URN: "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.order_items,PROD)"},
				ForeignFields: []struct {
					FieldPath string `json:"fieldPath"`
				}{
					{FieldPath: "order_id"},
					{FieldPath: "id"},
				},
			},
			expected: types.ForeignKey{
				Name:           "fk_order_item",
				SourceFields:   []string{"order_id", "item_id"},
				ForeignDataset: "urn:li:dataset:(urn:li:dataPlatform:snowflake,prod.sales.order_items,PROD)",
				ForeignFields:  []string{"order_id", "id"},
			},
		},
		{
			name:  "empty foreign key",
			input: rawForeignKey{},
			expected: types.ForeignKey{
				Name:           "",
				ForeignDataset: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseForeignKey(tc.input)

			if result.Name != tc.expected.Name {
				t.Errorf("Name = %s, want %s", result.Name, tc.expected.Name)
			}
			if result.ForeignDataset != tc.expected.ForeignDataset {
				t.Errorf("ForeignDataset = %s, want %s", result.ForeignDataset, tc.expected.ForeignDataset)
			}
			if len(result.SourceFields) != len(tc.expected.SourceFields) {
				t.Errorf("SourceFields count = %d, want %d", len(result.SourceFields), len(tc.expected.SourceFields))
			}
			for i, sf := range result.SourceFields {
				if sf != tc.expected.SourceFields[i] {
					t.Errorf("SourceFields[%d] = %s, want %s", i, sf, tc.expected.SourceFields[i])
				}
			}
			if len(result.ForeignFields) != len(tc.expected.ForeignFields) {
				t.Errorf("ForeignFields count = %d, want %d", len(result.ForeignFields), len(tc.expected.ForeignFields))
			}
			for i, ff := range result.ForeignFields {
				if ff != tc.expected.ForeignFields[i] {
					t.Errorf("ForeignFields[%d] = %s, want %s", i, ff, tc.expected.ForeignFields[i])
				}
			}
		})
	}
}

func TestParseSchemaMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    rawSchemaMetadata
		expected *types.SchemaMetadata
	}{
		{
			name: "schema with fields and foreign keys",
			input: rawSchemaMetadata{
				Name:        "orders",
				Version:     1,
				Hash:        "abc123",
				PrimaryKeys: []string{"order_id"},
				PlatformSchema: struct {
					Schema string `json:"schema"`
				}{Schema: "CREATE TABLE orders (...)"},
				Fields: []rawSchemaField{
					{
						FieldPath: "order_id",
						Type:      "NUMBER",
					},
					{
						FieldPath: "customer_id",
						Type:      "NUMBER",
					},
				},
				ForeignKeys: []rawForeignKey{
					{
						Name: "fk_customer",
						SourceFields: []struct {
							FieldPath string `json:"fieldPath"`
						}{
							{FieldPath: "customer_id"},
						},
						ForeignDataset: struct {
							URN string `json:"urn"`
						}{URN: "urn:li:dataset:customers"},
						ForeignFields: []struct {
							FieldPath string `json:"fieldPath"`
						}{
							{FieldPath: "id"},
						},
					},
				},
			},
			expected: &types.SchemaMetadata{
				Name:           "orders",
				Version:        1,
				Hash:           "abc123",
				PrimaryKeys:    []string{"order_id"},
				PlatformSchema: "CREATE TABLE orders (...)",
				Fields: []types.SchemaField{
					{FieldPath: "order_id", Type: "NUMBER"},
					{FieldPath: "customer_id", Type: "NUMBER"},
				},
				ForeignKeys: []types.ForeignKey{
					{
						Name:           "fk_customer",
						SourceFields:   []string{"customer_id"},
						ForeignDataset: "urn:li:dataset:customers",
						ForeignFields:  []string{"id"},
					},
				},
			},
		},
		{
			name:  "empty schema",
			input: rawSchemaMetadata{},
			expected: &types.SchemaMetadata{
				Name:    "",
				Version: 0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseSchemaMetadata(tc.input)

			if result.Name != tc.expected.Name {
				t.Errorf("Name = %s, want %s", result.Name, tc.expected.Name)
			}
			if result.Version != tc.expected.Version {
				t.Errorf("Version = %d, want %d", result.Version, tc.expected.Version)
			}
			if result.Hash != tc.expected.Hash {
				t.Errorf("Hash = %s, want %s", result.Hash, tc.expected.Hash)
			}
			if result.PlatformSchema != tc.expected.PlatformSchema {
				t.Errorf("PlatformSchema = %s, want %s", result.PlatformSchema, tc.expected.PlatformSchema)
			}
			if len(result.Fields) != len(tc.expected.Fields) {
				t.Errorf("Fields count = %d, want %d", len(result.Fields), len(tc.expected.Fields))
			}
			if len(result.ForeignKeys) != len(tc.expected.ForeignKeys) {
				t.Errorf("ForeignKeys count = %d, want %d", len(result.ForeignKeys), len(tc.expected.ForeignKeys))
			}
		})
	}
}
