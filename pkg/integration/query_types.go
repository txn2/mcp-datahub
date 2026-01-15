package integration

import "time"

// TableIdentifier uniquely identifies a table in a query engine.
// This structure is intentionally compatible with (but not imported from)
// mcp-trino's semantic.TableIdentifier to maintain island architecture.
type TableIdentifier struct {
	// Connection is the named connection (empty for default).
	Connection string `json:"connection,omitempty"`

	// Catalog is the catalog/database name.
	Catalog string `json:"catalog"`

	// Schema is the schema name.
	Schema string `json:"schema"`

	// Table is the table name.
	Table string `json:"table"`
}

// String returns the fully-qualified table name.
func (t TableIdentifier) String() string {
	if t.Connection != "" {
		return t.Connection + ":" + t.Catalog + "." + t.Schema + "." + t.Table
	}
	return t.Catalog + "." + t.Schema + "." + t.Table
}

// TableAvailability represents whether a DataHub entity is available
// as a queryable table in the connected query engine.
type TableAvailability struct {
	// Available indicates if the table exists and is queryable.
	Available bool `json:"available"`

	// Table is the resolved table identifier (nil if not available).
	Table *TableIdentifier `json:"table,omitempty"`

	// Connection is the query engine connection where the table is available.
	Connection string `json:"connection,omitempty"`

	// Error explains why the table is not available (if Available is false).
	Error string `json:"error,omitempty"`

	// LastChecked is when availability was last verified.
	LastChecked time.Time `json:"last_checked,omitempty"`

	// RowCount is an optional estimate of table rows (if known).
	RowCount *int64 `json:"row_count,omitempty"`

	// LastUpdated is when the table data was last modified (if known).
	LastUpdated *time.Time `json:"last_updated,omitempty"`
}

// QueryExample represents a sample SQL query for a DataHub entity.
type QueryExample struct {
	// Name is a short identifier for the example.
	Name string `json:"name"`

	// Description explains what the query does.
	Description string `json:"description,omitempty"`

	// SQL is the executable SQL statement.
	SQL string `json:"sql"`

	// Category classifies the example type.
	// Common values: "sample", "aggregation", "join", "filter", "common"
	Category string `json:"category,omitempty"`

	// Source indicates where this example came from.
	// Values: "generated", "history", "template", "documentation"
	Source string `json:"source,omitempty"`
}

// ExecutionContext provides query execution context for lineage bridging.
// This connects DataHub lineage information to query engine execution details.
type ExecutionContext struct {
	// Tables maps DataHub URNs to their resolved table identifiers.
	Tables map[string]*TableIdentifier `json:"tables,omitempty"`

	// Connections lists the query engine connections involved.
	Connections []string `json:"connections,omitempty"`

	// Queries are relevant queries that involve these entities.
	Queries []ExecutionQuery `json:"queries,omitempty"`

	// Source indicates which provider supplied this context.
	Source string `json:"source,omitempty"`
}

// ExecutionQuery represents a query that involves lineage entities.
type ExecutionQuery struct {
	// SQL is the query text.
	SQL string `json:"sql,omitempty"`

	// Sources are the source table URNs.
	Sources []string `json:"sources,omitempty"`

	// Targets are the target table URNs (for INSERT/CREATE).
	Targets []string `json:"targets,omitempty"`

	// ExecutedAt is when the query was last executed (if known).
	ExecutedAt *time.Time `json:"executed_at,omitempty"`

	// QueryID is the query engine's query identifier.
	QueryID string `json:"query_id,omitempty"`
}
