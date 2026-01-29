package client

import (
	"testing"
)

func TestToEnumCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"single lowercase char", "a", "A"},
		{"single uppercase char", "A", "A"},
		{"all uppercase", "DATASET", "DATASET"},
		{"all lowercase", "dataset", "DATASET"},
		{"camelCase", "glossaryTerm", "GLOSSARY_TERM"},
		{"PascalCase", "GlossaryTerm", "GLOSSARY_TERM"},
		{"multi-word camelCase", "dataProduct", "DATA_PRODUCT"},
		{"multi-word PascalCase", "DataProduct", "DATA_PRODUCT"},
		{"three words camelCase", "corpUserGroup", "CORP_USER_GROUP"},
		{"already screaming snake", "GLOSSARY_TERM", "GLOSSARY_TERM"},
		{"mixed with numbers", "mlModel", "ML_MODEL"},
		{"consecutive caps", "MLModel", "MLMODEL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toEnumCase(tt.input)
			if got != tt.want {
				t.Errorf("toEnumCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWithEntityType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already uppercase", "DASHBOARD", "DASHBOARD"},
		{"camelCase glossaryTerm", "glossaryTerm", "GLOSSARY_TERM"},
		{"camelCase dataProduct", "dataProduct", "DATA_PRODUCT"},
		{"camelCase corpUser", "corpUser", "CORP_USER"},
		{"lowercase", "dataset", "DATASET"},
		{"PascalCase", "GlossaryTerm", "GLOSSARY_TERM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &searchOptions{}
			WithEntityType(tt.input)(opts)
			if opts.entityType != tt.want {
				t.Errorf("WithEntityType(%q) = %q, want %q", tt.input, opts.entityType, tt.want)
			}
		})
	}
}

func TestWithLimit(t *testing.T) {
	opts := &searchOptions{}
	WithLimit(50)(opts)

	if opts.limit != 50 {
		t.Errorf("WithLimit() = %d, want 50", opts.limit)
	}
}

func TestWithOffset(t *testing.T) {
	opts := &searchOptions{}
	WithOffset(10)(opts)

	if opts.offset != 10 {
		t.Errorf("WithOffset() = %d, want 10", opts.offset)
	}
}

func TestWithFilters(t *testing.T) {
	opts := &searchOptions{}
	filters := map[string][]string{
		"platform": {"snowflake", "bigquery"},
	}
	WithFilters(filters)(opts)

	if len(opts.filters) != 1 {
		t.Errorf("WithFilters() count = %d, want 1", len(opts.filters))
	}
}

func TestLineageOptions(t *testing.T) {
	tests := []struct {
		name      string
		applyOpts func(*lineageOptions)
		checkFunc func(*lineageOptions) bool
	}{
		{
			name: "direction upstream",
			applyOpts: func(o *lineageOptions) {
				WithDirection(LineageDirectionUpstream)(o)
			},
			checkFunc: func(o *lineageOptions) bool {
				return o.direction == LineageDirectionUpstream
			},
		},
		{
			name: "direction downstream",
			applyOpts: func(o *lineageOptions) {
				WithDirection(LineageDirectionDownstream)(o)
			},
			checkFunc: func(o *lineageOptions) bool {
				return o.direction == LineageDirectionDownstream
			},
		},
		{
			name: "direction lowercase upstream normalized",
			applyOpts: func(o *lineageOptions) {
				WithDirection("upstream")(o)
			},
			checkFunc: func(o *lineageOptions) bool {
				return o.direction == LineageDirectionUpstream
			},
		},
		{
			name: "direction lowercase downstream normalized",
			applyOpts: func(o *lineageOptions) {
				WithDirection("downstream")(o)
			},
			checkFunc: func(o *lineageOptions) bool {
				return o.direction == LineageDirectionDownstream
			},
		},
		{
			name: "direction mixed case normalized",
			applyOpts: func(o *lineageOptions) {
				WithDirection("Upstream")(o)
			},
			checkFunc: func(o *lineageOptions) bool {
				return o.direction == LineageDirectionUpstream
			},
		},
		{
			name: "depth",
			applyOpts: func(o *lineageOptions) {
				WithDepth(3)(o)
			},
			checkFunc: func(o *lineageOptions) bool {
				return o.depth == 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &lineageOptions{}
			tt.applyOpts(opts)
			if !tt.checkFunc(opts) {
				t.Errorf("Option not applied correctly")
			}
		})
	}
}
