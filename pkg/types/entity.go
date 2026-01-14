package types

// Entity represents a DataHub entity with common metadata.
type Entity struct {
	// URN is the unique identifier for this entity.
	URN string `json:"urn"`

	// Type is the entity type (dataset, dashboard, dataFlow, etc.).
	Type string `json:"type"`

	// Name is the display name of the entity.
	Name string `json:"name"`

	// Description is the business description.
	Description string `json:"description,omitempty"`

	// Owners lists the owners of this entity.
	Owners []Owner `json:"owners,omitempty"`

	// Tags lists the tags applied to this entity.
	Tags []Tag `json:"tags,omitempty"`

	// GlossaryTerms lists the glossary terms associated with this entity.
	GlossaryTerms []GlossaryTerm `json:"glossary_terms,omitempty"`

	// Domain is the data domain this entity belongs to.
	Domain *Domain `json:"domain,omitempty"`

	// Platform is the data platform (for datasets).
	Platform string `json:"platform,omitempty"`

	// Deprecation contains deprecation info if the entity is deprecated.
	Deprecation *Deprecation `json:"deprecation,omitempty"`

	// Properties contains additional entity-specific properties.
	Properties map[string]any `json:"properties,omitempty"`

	// Created is the creation timestamp.
	Created int64 `json:"created,omitempty"`

	// LastModified is the last modification timestamp.
	LastModified int64 `json:"last_modified,omitempty"`
}

// Deprecation contains deprecation information.
type Deprecation struct {
	// Deprecated: This field indicates if the entity is deprecated.
	Deprecated bool `json:"deprecated"`

	// Note is the deprecation note.
	Note string `json:"note,omitempty"`

	// Actor is who deprecated the entity.
	Actor string `json:"actor,omitempty"`

	// DecommissionTime is when the entity will be decommissioned.
	DecommissionTime int64 `json:"decommission_time,omitempty"`
}

// Dataset represents a DataHub dataset entity.
type Dataset struct {
	Entity

	// Schema contains the dataset schema.
	Schema *SchemaMetadata `json:"schema,omitempty"`

	// SubTypes are the dataset sub-types (table, view, etc.).
	SubTypes []string `json:"sub_types,omitempty"`
}

// Dashboard represents a DataHub dashboard entity.
type Dashboard struct {
	Entity

	// DashboardURL is the URL to the dashboard.
	DashboardURL string `json:"dashboard_url,omitempty"`

	// Charts lists the charts in this dashboard.
	Charts []string `json:"charts,omitempty"`
}

// Pipeline represents a DataHub data pipeline entity.
type Pipeline struct {
	Entity

	// DataFlow is the parent data flow URN.
	DataFlow string `json:"data_flow,omitempty"`

	// Inputs are the input dataset URNs.
	Inputs []string `json:"inputs,omitempty"`

	// Outputs are the output dataset URNs.
	Outputs []string `json:"outputs,omitempty"`
}
