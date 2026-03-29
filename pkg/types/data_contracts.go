package types

// DataContract represents a data contract on a dataset (DataHub 1.3.x+).
// Contracts bundle freshness, schema, and data quality assertions into
// a single pass/fail quality signal.
type DataContract struct {
	// Status is the overall contract state: PASSING, FAILING, or NOT_APPLICABLE.
	Status string `json:"status"`

	// AssertionResults contains the assertion references grouped by category.
	AssertionResults []AssertionResult `json:"assertion_results,omitempty"`
}

// AssertionResult represents a single assertion reference within a data contract.
type AssertionResult struct {
	// AssertionURN identifies the assertion.
	AssertionURN string `json:"assertion_urn"`

	// Type is the assertion category: FRESHNESS, SCHEMA, or DATA_QUALITY.
	Type string `json:"type"`
}
