package client

import "github.com/txn2/mcp-datahub/pkg/types"

// rawSchemaMetadata is the internal struct for unmarshaling schema responses.
type rawSchemaMetadata struct {
	Name           string   `json:"name"`
	Version        int64    `json:"version"`
	Hash           string   `json:"hash"`
	PrimaryKeys    []string `json:"primaryKeys"`
	PlatformSchema struct {
		Schema string `json:"schema"`
	} `json:"platformSchema"`
	Fields      []rawSchemaField `json:"fields"`
	ForeignKeys []rawForeignKey  `json:"foreignKeys"`
}

// rawSchemaField is the internal struct for unmarshaling field responses.
type rawSchemaField struct {
	FieldPath      string `json:"fieldPath"`
	Type           string `json:"type"`
	NativeDataType string `json:"nativeDataType"`
	Description    string `json:"description"`
	Nullable       bool   `json:"nullable"`
	IsPartOfKey    bool   `json:"isPartOfKey"`
	Tags           struct {
		Tags []struct {
			Tag struct {
				URN  string `json:"urn"`
				Name string `json:"name"`
			} `json:"tag"`
		} `json:"tags"`
	} `json:"tags"`
	GlossaryTerms struct {
		Terms []struct {
			Term struct {
				URN  string `json:"urn"`
				Name string `json:"name"`
			} `json:"term"`
		} `json:"terms"`
	} `json:"glossaryTerms"`
}

// rawForeignKey is the internal struct for unmarshaling foreign key responses.
type rawForeignKey struct {
	Name         string `json:"name"`
	SourceFields []struct {
		FieldPath string `json:"fieldPath"`
	} `json:"sourceFields"`
	ForeignDataset struct {
		URN string `json:"urn"`
	} `json:"foreignDataset"`
	ForeignFields []struct {
		FieldPath string `json:"fieldPath"`
	} `json:"foreignFields"`
}

// parseSchemaMetadata converts raw schema data to types.SchemaMetadata.
func parseSchemaMetadata(raw rawSchemaMetadata) *types.SchemaMetadata {
	schema := &types.SchemaMetadata{
		Name:           raw.Name,
		Version:        raw.Version,
		Hash:           raw.Hash,
		PrimaryKeys:    raw.PrimaryKeys,
		PlatformSchema: raw.PlatformSchema.Schema,
	}

	for _, f := range raw.Fields {
		field := parseSchemaField(f)
		schema.Fields = append(schema.Fields, field)
	}

	for _, fk := range raw.ForeignKeys {
		foreignKey := parseForeignKey(fk)
		schema.ForeignKeys = append(schema.ForeignKeys, foreignKey)
	}

	return schema
}

// parseSchemaField converts a raw field to types.SchemaField.
func parseSchemaField(f rawSchemaField) types.SchemaField {
	field := types.SchemaField{
		FieldPath:      f.FieldPath,
		Type:           f.Type,
		NativeType:     f.NativeDataType,
		Description:    f.Description,
		Nullable:       f.Nullable,
		IsPartitionKey: f.IsPartOfKey,
	}

	for _, t := range f.Tags.Tags {
		field.Tags = append(field.Tags, types.Tag{
			URN:  t.Tag.URN,
			Name: t.Tag.Name,
		})
	}

	for _, gt := range f.GlossaryTerms.Terms {
		field.GlossaryTerms = append(field.GlossaryTerms, types.GlossaryTerm{
			URN:  gt.Term.URN,
			Name: gt.Term.Name,
		})
	}

	return field
}

// parseForeignKey converts a raw foreign key to types.ForeignKey.
func parseForeignKey(fk rawForeignKey) types.ForeignKey {
	foreignKey := types.ForeignKey{
		Name:           fk.Name,
		ForeignDataset: fk.ForeignDataset.URN,
	}

	for _, sf := range fk.SourceFields {
		foreignKey.SourceFields = append(foreignKey.SourceFields, sf.FieldPath)
	}
	for _, ff := range fk.ForeignFields {
		foreignKey.ForeignFields = append(foreignKey.ForeignFields, ff.FieldPath)
	}

	return foreignKey
}
