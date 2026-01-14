package types

// SchemaMetadata represents the schema of a dataset.
type SchemaMetadata struct {
	// Name is the schema name.
	Name string `json:"name,omitempty"`

	// PlatformSchema is the platform-specific schema representation.
	PlatformSchema string `json:"platform_schema,omitempty"`

	// Version is the schema version.
	Version int64 `json:"version,omitempty"`

	// Fields is the list of schema fields.
	Fields []SchemaField `json:"fields"`

	// PrimaryKeys lists the primary key field paths.
	PrimaryKeys []string `json:"primary_keys,omitempty"`

	// ForeignKeys lists foreign key relationships.
	ForeignKeys []ForeignKey `json:"foreign_keys,omitempty"`

	// Hash is the schema hash for change detection.
	Hash string `json:"hash,omitempty"`
}

// SchemaField represents a field in a dataset schema.
type SchemaField struct {
	// FieldPath is the full path to this field (e.g., "user.address.city").
	FieldPath string `json:"field_path"`

	// Type is the field's data type.
	Type string `json:"type"`

	// NativeType is the platform-specific type.
	NativeType string `json:"native_type,omitempty"`

	// Description is the field description.
	Description string `json:"description,omitempty"`

	// Nullable indicates if the field can be null.
	Nullable bool `json:"nullable"`

	// IsPartitionKey indicates if this is a partition key.
	IsPartitionKey bool `json:"is_partition_key,omitempty"`

	// Tags lists the tags applied to this field.
	Tags []Tag `json:"tags,omitempty"`

	// GlossaryTerms lists glossary terms for this field.
	GlossaryTerms []GlossaryTerm `json:"glossary_terms,omitempty"`

	// JSONPath is the JSON path for nested fields.
	JSONPath string `json:"json_path,omitempty"`
}

// ForeignKey represents a foreign key relationship.
type ForeignKey struct {
	// Name is the constraint name.
	Name string `json:"name,omitempty"`

	// SourceFields are the source field paths.
	SourceFields []string `json:"source_fields"`

	// ForeignDataset is the referenced dataset URN.
	ForeignDataset string `json:"foreign_dataset"`

	// ForeignFields are the referenced field paths.
	ForeignFields []string `json:"foreign_fields"`
}
