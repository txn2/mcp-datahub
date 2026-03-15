package types

// DataContract represents a data contract on a dataset (DataHub 1.4.x+).
// Contracts bundle freshness, schema, and data quality assertions into
// a single pass/fail quality signal.
type DataContract struct {
	// Status is the overall contract status: PASSING or FAILING.
	Status string `json:"status"`

	// AssertionResults contains the individual assertion outcomes.
	AssertionResults []AssertionResult `json:"assertion_results,omitempty"`
}

// AssertionResult represents a single assertion outcome within a data contract.
type AssertionResult struct {
	// AssertionURN identifies the assertion.
	AssertionURN string `json:"assertion_urn"`

	// Type is the assertion category: FRESHNESS, SCHEMA, or DATA_QUALITY.
	Type string `json:"type"`

	// ResultType is the assertion outcome (e.g., "SUCCESS", "FAILURE").
	ResultType string `json:"result_type"`

	// NativeResults contains platform-specific result details.
	NativeResults map[string]string `json:"native_results,omitempty"`
}
