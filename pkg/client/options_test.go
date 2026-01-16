package client

import (
	"testing"
)

func TestWithEntityType(t *testing.T) {
	opts := &searchOptions{}
	WithEntityType("DASHBOARD")(opts)

	if opts.entityType != "DASHBOARD" {
		t.Errorf("WithEntityType() = %s, want DASHBOARD", opts.entityType)
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
