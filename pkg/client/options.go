package client

import (
	"strings"
	"unicode"
)

// SearchOption configures search behavior.
type SearchOption func(*searchOptions)

type searchOptions struct {
	entityType string
	limit      int
	offset     int
	filters    map[string][]string
}

// toEnumCase converts camelCase or PascalCase strings to SCREAMING_SNAKE_CASE.
// Examples: glossaryTerm -> GLOSSARY_TERM, dataProduct -> DATA_PRODUCT.
// Strings that are already uppercase (with or without underscores) pass through unchanged.
func toEnumCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	var result strings.Builder
	for i, r := range runes {
		// Insert underscore before uppercase if preceded by lowercase
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runes[i-1]) {
			_, _ = result.WriteRune('_')
		}
		_, _ = result.WriteRune(unicode.ToUpper(r))
	}
	return result.String()
}

// WithEntityType filters search by entity type.
// Valid types: DATASET, DASHBOARD, DATA_FLOW, DATA_JOB, CONTAINER, DOMAIN,
// TAG, GLOSSARY_TERM, CORP_USER, CORP_GROUP, DATA_PRODUCT, etc.
// The entity type is normalized to SCREAMING_SNAKE_CASE (e.g., glossaryTerm -> GLOSSARY_TERM).
func WithEntityType(entityType string) SearchOption {
	return func(o *searchOptions) {
		o.entityType = toEnumCase(entityType)
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
// The direction is normalized to uppercase.
func WithDirection(dir string) LineageOption {
	return func(o *lineageOptions) {
		o.direction = strings.ToUpper(dir)
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
