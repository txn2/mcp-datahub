package types

// ColumnLineage represents fine-grained column-level lineage for a dataset.
type ColumnLineage struct {
	// DatasetURN is the URN of the dataset this lineage is for.
	DatasetURN string `json:"dataset_urn"`

	// Mappings contains the column-level lineage mappings.
	Mappings []ColumnLineageMapping `json:"mappings"`
}

// ColumnLineageMapping represents a single column lineage relationship.
// It describes how a downstream column is derived from one or more upstream columns.
type ColumnLineageMapping struct {
	// DownstreamColumn is the field path in the downstream dataset.
	DownstreamColumn string `json:"downstream_column"`

	// UpstreamDataset is the URN of the upstream dataset.
	UpstreamDataset string `json:"upstream_dataset"`

	// UpstreamColumn is the field path in the upstream dataset.
	UpstreamColumn string `json:"upstream_column"`

	// Transform describes the transformation operation (optional).
	// Examples: "IDENTITY", "TRANSFORM", "AGGREGATE"
	Transform string `json:"transform,omitempty"`

	// Query is the URN of the query that created this lineage (optional).
	Query string `json:"query,omitempty"`

	// ConfidenceScore indicates the confidence of the lineage mapping (optional).
	// Values typically range from 0.0 to 1.0.
	ConfidenceScore float64 `json:"confidence_score,omitempty"`
}
