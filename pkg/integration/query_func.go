package integration

import "context"

// QueryProviderFunc allows implementing QueryProvider with individual functions.
// Any nil function returns nil/empty results (not errors).
//
// Example:
//
//	provider := &integration.QueryProviderFunc{
//	    NameFn: func() string { return "custom" },
//	    ResolveTableFn: func(ctx context.Context, urn string) (*TableIdentifier, error) {
//	        // Custom resolution logic
//	        return parseURN(urn), nil
//	    },
//	}
type QueryProviderFunc struct {
	NameFn                 func() string
	ResolveTableFn         func(ctx context.Context, urn string) (*TableIdentifier, error)
	GetTableAvailabilityFn func(ctx context.Context, urn string) (*TableAvailability, error)
	GetQueryExamplesFn     func(ctx context.Context, urn string) ([]QueryExample, error)
	GetExecutionContextFn  func(ctx context.Context, urns []string) (*ExecutionContext, error)
	CloseFn                func() error
}

// Name implements QueryProvider.
func (f *QueryProviderFunc) Name() string {
	if f.NameFn == nil {
		return "func"
	}
	return f.NameFn()
}

// ResolveTable implements QueryProvider.
func (f *QueryProviderFunc) ResolveTable(ctx context.Context, urn string) (*TableIdentifier, error) {
	if f.ResolveTableFn == nil {
		return nil, nil
	}
	return f.ResolveTableFn(ctx, urn)
}

// GetTableAvailability implements QueryProvider.
func (f *QueryProviderFunc) GetTableAvailability(ctx context.Context, urn string) (*TableAvailability, error) {
	if f.GetTableAvailabilityFn == nil {
		return nil, nil
	}
	return f.GetTableAvailabilityFn(ctx, urn)
}

// GetQueryExamples implements QueryProvider.
func (f *QueryProviderFunc) GetQueryExamples(ctx context.Context, urn string) ([]QueryExample, error) {
	if f.GetQueryExamplesFn == nil {
		return nil, nil
	}
	return f.GetQueryExamplesFn(ctx, urn)
}

// GetExecutionContext implements QueryProvider.
func (f *QueryProviderFunc) GetExecutionContext(ctx context.Context, urns []string) (*ExecutionContext, error) {
	if f.GetExecutionContextFn == nil {
		return nil, nil
	}
	return f.GetExecutionContextFn(ctx, urns)
}

// Close implements QueryProvider.
func (f *QueryProviderFunc) Close() error {
	if f.CloseFn == nil {
		return nil
	}
	return f.CloseFn()
}

// Verify QueryProviderFunc implements QueryProvider.
var _ QueryProvider = (*QueryProviderFunc)(nil)
