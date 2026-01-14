package types

// ParsedURN represents a parsed DataHub URN.
type ParsedURN struct {
	// Raw is the original URN string.
	Raw string `json:"raw"`

	// EntityType is the type of entity (dataset, dashboard, dataFlow, etc.).
	EntityType string `json:"entity_type"`

	// Platform is the data platform (snowflake, bigquery, postgres, etc.).
	Platform string `json:"platform,omitempty"`

	// Name is the qualified name of the entity.
	Name string `json:"name"`

	// Env is the environment (PROD, DEV, etc.).
	Env string `json:"env,omitempty"`
}
