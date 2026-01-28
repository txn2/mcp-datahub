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

// rawEditableSchemaMetadata is the internal struct for unmarshaling editableSchemaMetadata responses.
// This contains user-edited column metadata added via the DataHub UI.
type rawEditableSchemaMetadata struct {
	EditableSchemaFieldInfo []rawEditableSchemaFieldInfo `json:"editableSchemaFieldInfo"`
}

// rawEditableSchemaFieldInfo is the internal struct for unmarshaling editable field info.
type rawEditableSchemaFieldInfo struct {
	FieldPath     string `json:"fieldPath"`
	Description   string `json:"description"`
	GlossaryTerms struct {
		Terms []struct {
			Term struct {
				URN  string `json:"urn"`
				Name string `json:"name"`
			} `json:"term"`
		} `json:"terms"`
	} `json:"glossaryTerms"`
	Tags struct {
		Tags []struct {
			Tag struct {
				URN  string `json:"urn"`
				Name string `json:"name"`
			} `json:"tag"`
		} `json:"tags"`
	} `json:"tags"`
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

// mergeEditableSchemaMetadata merges UI-edited metadata into schema fields.
// When editable metadata exists for a field, it REPLACES the ingested metadata
// because the user edited it for a reason - their edits are the source of truth.
func mergeEditableSchemaMetadata(schema *types.SchemaMetadata, edited rawEditableSchemaMetadata) {
	if schema == nil || len(edited.EditableSchemaFieldInfo) == 0 {
		return
	}

	// Build lookup map by fieldPath
	editedByPath := make(map[string]*rawEditableSchemaFieldInfo)
	for i := range edited.EditableSchemaFieldInfo {
		info := &edited.EditableSchemaFieldInfo[i]
		editedByPath[info.FieldPath] = info
	}

	// Merge into schema fields
	for i := range schema.Fields {
		field := &schema.Fields[i]
		editedInfo, ok := editedByPath[field.FieldPath]
		if !ok {
			continue
		}

		// Override description if edited (UI description takes precedence)
		if editedInfo.Description != "" {
			field.Description = editedInfo.Description
		}

		// Replace glossary terms if any were edited (user's edits are source of truth)
		if len(editedInfo.GlossaryTerms.Terms) > 0 {
			field.GlossaryTerms = nil
			for _, gt := range editedInfo.GlossaryTerms.Terms {
				field.GlossaryTerms = append(field.GlossaryTerms, types.GlossaryTerm{
					URN:  gt.Term.URN,
					Name: gt.Term.Name,
				})
			}
		}

		// Replace tags if any were edited (user's edits are source of truth)
		if len(editedInfo.Tags.Tags) > 0 {
			field.Tags = nil
			for _, t := range editedInfo.Tags.Tags {
				field.Tags = append(field.Tags, types.Tag{
					URN:  t.Tag.URN,
					Name: t.Tag.Name,
				})
			}
		}
	}
}
