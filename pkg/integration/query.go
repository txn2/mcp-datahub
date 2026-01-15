package integration

import "context"

// QueryProvider provides query execution context for DataHub entities.
// Implementations typically connect to Trino, Spark, or other SQL engines
// to provide table resolution, query examples, and availability information.
//
// This interface enables mcp-trino (or other query toolkits) to inject
// execution context into mcp-datahub tools WITHOUT creating import cycles.
//
// All methods should return nil (not an error) if information is not found.
// Errors should only be returned for connection/authentication failures
// or other exceptional conditions.
//
// Implementations must be safe for concurrent use.
type QueryProvider interface {
	// Name returns the provider name (e.g., "trino", "spark", "presto").
	// Used for logging, metrics, and source attribution.
	Name() string

	// ResolveTable maps a DataHub URN to a query engine table identifier.
	// Returns nil if the URN cannot be resolved to a queryable table.
	//
	// Example URN: urn:li:dataset:(urn:li:dataPlatform:trino,catalog.schema.table,PROD)
	// Returns: &TableIdentifier{Catalog: "catalog", Schema: "schema", Table: "table"}
	ResolveTable(ctx context.Context, urn string) (*TableIdentifier, error)

	// GetTableAvailability checks if a DataHub entity is available as a queryable table.
	// This is used to enrich search results with query availability status.
	// Returns nil if availability cannot be determined (not an error).
	GetTableAvailability(ctx context.Context, urn string) (*TableAvailability, error)

	// GetQueryExamples returns sample SQL queries for a DataHub entity.
	// Returns an empty slice if no examples are available (not an error).
	// Examples might include: SELECT samples, common aggregations, join patterns.
	GetQueryExamples(ctx context.Context, urn string) ([]QueryExample, error)

	// GetExecutionContext returns execution context for lineage bridging.
	// Given a set of DataHub URNs (e.g., from lineage), returns information
	// about how they relate to query execution (sources, targets, transformations).
	// Returns nil if no execution context is available (not an error).
	GetExecutionContext(ctx context.Context, urns []string) (*ExecutionContext, error)

	// Close releases any resources held by the provider.
	// Implementations should be idempotent.
	Close() error
}

// NoOpQueryProvider is a QueryProvider that does nothing.
// Use this as a placeholder or for testing.
type NoOpQueryProvider struct{}

// Name implements QueryProvider.
func (n *NoOpQueryProvider) Name() string { return "noop" }

// ResolveTable implements QueryProvider.
func (n *NoOpQueryProvider) ResolveTable(_ context.Context, _ string) (*TableIdentifier, error) {
	return nil, nil
}

// GetTableAvailability implements QueryProvider.
func (n *NoOpQueryProvider) GetTableAvailability(_ context.Context, _ string) (*TableAvailability, error) {
	return nil, nil
}

// GetQueryExamples implements QueryProvider.
func (n *NoOpQueryProvider) GetQueryExamples(_ context.Context, _ string) ([]QueryExample, error) {
	return nil, nil
}

// GetExecutionContext implements QueryProvider.
func (n *NoOpQueryProvider) GetExecutionContext(_ context.Context, _ []string) (*ExecutionContext, error) {
	return nil, nil
}

// Close implements QueryProvider.
func (n *NoOpQueryProvider) Close() error { return nil }

// Verify NoOpQueryProvider implements QueryProvider.
var _ QueryProvider = (*NoOpQueryProvider)(nil)
