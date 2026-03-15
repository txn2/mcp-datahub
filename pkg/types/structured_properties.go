package types

// StructuredPropertyDefinition describes a structured property type registered in DataHub.
// Structured properties (DataHub 1.4.x+) provide typed custom metadata with optional
// value constraints, replacing freeform custom properties for governed use cases.
type StructuredPropertyDefinition struct {
	// URN is the unique identifier for this structured property definition.
	URN string `json:"urn"`

	// QualifiedName is the fully qualified property name (e.g., "io.acryl.privacy.retentionTime").
	QualifiedName string `json:"qualified_name"`

	// DisplayName is the human-readable name shown in the UI.
	DisplayName string `json:"display_name,omitempty"`

	// Description explains the purpose and usage of this property.
	Description string `json:"description,omitempty"`

	// ValueType is the data type for values (e.g., "string", "number", "date", "urn").
	ValueType string `json:"value_type"`

	// Cardinality indicates whether the property accepts a single value or multiple.
	// Values: "SINGLE", "MULTIPLE".
	Cardinality string `json:"cardinality,omitempty"`

	// EntityTypes lists which entity types this property can be applied to.
	EntityTypes []string `json:"entity_types,omitempty"`

	// AllowedValues constrains the set of valid values when defined.
	AllowedValues []AllowedValue `json:"allowed_values,omitempty"`
}

// AllowedValue represents a permitted value for a structured property with optional description.
type AllowedValue struct {
	// Value is the allowed value.
	Value string `json:"value"`

	// Description explains when to use this value.
	Description string `json:"description,omitempty"`
}

// StructuredPropertyValue represents a structured property assignment on an entity.
type StructuredPropertyValue struct {
	// PropertyURN identifies which structured property definition this value belongs to.
	PropertyURN string `json:"property_urn"`

	// Definition contains the full property definition, when available.
	Definition *StructuredPropertyDefinition `json:"definition,omitempty"`

	// Values holds the assigned value(s). Elements are typically strings or numbers.
	Values []any `json:"values"`
}

// StructuredPropertyInput represents a structured property value to set on an entity.
type StructuredPropertyInput struct {
	// PropertyURN is the URN of the structured property definition.
	PropertyURN string

	// Values holds the value(s) to assign. Elements should be strings or numbers.
	Values []any
}
