package types

// QueryList represents a list of queries associated with a dataset.
type QueryList struct {
	// Queries is the list of queries.
	Queries []Query `json:"queries"`

	// Total is the total number of queries.
	Total int `json:"total"`
}

// Query represents a SQL query associated with a dataset.
type Query struct {
	// URN is the query URN.
	URN string `json:"urn,omitempty"`

	// Name is the query name.
	Name string `json:"name,omitempty"`

	// Statement is the SQL query text.
	Statement string `json:"statement"`

	// Description is the query description.
	Description string `json:"description,omitempty"`

	// Source indicates how the query was created (e.g., "MANUAL", "SYSTEM").
	Source string `json:"source,omitempty"`

	// CreatedBy is who created the query.
	CreatedBy string `json:"created_by,omitempty"`

	// Created is when the query was created.
	Created int64 `json:"created,omitempty"`

	// LastRun is when the query was last executed.
	LastRun int64 `json:"last_run,omitempty"`

	// RunCount is how many times the query has been run.
	RunCount int `json:"run_count,omitempty"`
}
