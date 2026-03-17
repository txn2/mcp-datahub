package types

// CreateDocumentInput contains parameters for creating a new context document.
type CreateDocumentInput struct {
	// Title is the document title (required).
	Title string

	// Content is the document body text.
	Content string

	// Status is the publication state: PUBLISHED or UNPUBLISHED.
	Status string

	// SubType is the document sub-type classification.
	SubType string

	// RelatedAssetURNs are entity URNs to link to this document.
	RelatedAssetURNs []string

	// GlobalContext controls whether the document appears in global search.
	GlobalContext bool
}

// CreateStructuredPropertyInput contains parameters for creating a structured property definition.
type CreateStructuredPropertyInput struct {
	// QualifiedName is the fully qualified property name (required, e.g., "io.acryl.privacy.retentionTime").
	QualifiedName string

	// DisplayName is the human-readable name shown in the UI.
	DisplayName string

	// Description explains the purpose and usage of this property.
	Description string

	// ValueType is the data type for values (required, e.g., "string", "number", "date", "urn").
	ValueType string

	// Cardinality indicates whether the property accepts a single value or multiple.
	// Values: "SINGLE", "MULTIPLE". Defaults to "SINGLE".
	Cardinality string

	// EntityTypes lists which entity types this property can be applied to.
	EntityTypes []string

	// AllowedValues constrains the set of valid values when defined.
	AllowedValues []AllowedValue
}

// UpsertDataContractInput contains parameters for creating or updating a data contract.
type UpsertDataContractInput struct {
	// DatasetURN is the dataset to apply the contract to (required).
	DatasetURN string

	// SchemaAssertionURNs are assertion URNs for schema validation.
	SchemaAssertionURNs []string

	// FreshnessAssertionURNs are assertion URNs for freshness validation.
	FreshnessAssertionURNs []string

	// DataQualityAssertionURNs are assertion URNs for data quality validation.
	DataQualityAssertionURNs []string
}

// UpdateIncidentInput contains parameters for updating an existing incident.
// Note: type and customType are set at creation time via RaiseIncidentInput
// and cannot be changed via updateIncident.
type UpdateIncidentInput struct {
	// Title is the updated title (empty means no change).
	Title string

	// Description is the updated description (empty means no change).
	Description string

	// Priority is the updated priority (empty means no change).
	// Values: LOW, MEDIUM, HIGH, CRITICAL.
	Priority string
}

// UpdateStructuredPropertyInput contains parameters for updating a structured property definition.
type UpdateStructuredPropertyInput struct {
	// DisplayName is the updated human-readable name.
	DisplayName string

	// Description is the updated description.
	Description string

	// NewAllowedValues are additional allowed values to add.
	NewAllowedValues []AllowedValue
}
