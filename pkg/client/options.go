package client

// SearchOption configures search behavior.
type SearchOption func(*searchOptions)

type searchOptions struct {
	entityType string
	limit      int
	offset     int
	filters    map[string][]string
}

// WithEntityType filters search by entity type.
// Valid types: DATASET, DASHBOARD, DATA_FLOW, DATA_JOB, CONTAINER, DOMAIN,
// TAG, GLOSSARY_TERM, CORP_USER, CORP_GROUP, DATA_PRODUCT, etc.
func WithEntityType(entityType string) SearchOption {
	return func(o *searchOptions) {
		o.entityType = entityType
	}
}

// WithEntityTypes is an alias for WithEntityType (uses first type).
// Deprecated: Use WithEntityType instead.
func WithEntityTypes(types ...string) SearchOption {
	return func(o *searchOptions) {
		if len(types) > 0 {
			o.entityType = types[0]
		}
	}
}

// WithLimit sets the maximum number of results.
func WithLimit(limit int) SearchOption {
	return func(o *searchOptions) {
		o.limit = limit
	}
}

// WithOffset sets the result offset for pagination.
func WithOffset(offset int) SearchOption {
	return func(o *searchOptions) {
		o.offset = offset
	}
}

// WithFilters adds search filters.
func WithFilters(filters map[string][]string) SearchOption {
	return func(o *searchOptions) {
		o.filters = filters
	}
}

// LineageOption configures lineage queries.
type LineageOption func(*lineageOptions)

type lineageOptions struct {
	direction string
	depth     int
}

// WithDirection sets the lineage direction (UPSTREAM or DOWNSTREAM).
func WithDirection(dir string) LineageOption {
	return func(o *lineageOptions) {
		o.direction = dir
	}
}

// WithDepth sets the maximum lineage traversal depth.
func WithDepth(depth int) LineageOption {
	return func(o *lineageOptions) {
		o.depth = depth
	}
}

// Constants for lineage directions.
const (
	LineageDirectionUpstream   = "UPSTREAM"
	LineageDirectionDownstream = "DOWNSTREAM"
)
