package tools

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultLimit != 10 {
		t.Errorf("DefaultConfig() DefaultLimit = %d, want 10", cfg.DefaultLimit)
	}
	if cfg.MaxLimit != 100 {
		t.Errorf("DefaultConfig() MaxLimit = %d, want 100", cfg.MaxLimit)
	}
	if cfg.MaxLineageDepth != 5 {
		t.Errorf("DefaultConfig() MaxLineageDepth = %d, want 5", cfg.MaxLineageDepth)
	}
}

func TestNormalizeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    Config
		expected Config
	}{
		{
			name:  "all zeros get defaults",
			input: Config{},
			expected: Config{
				DefaultLimit:    10,
				MaxLimit:        100,
				MaxLineageDepth: 5,
			},
		},
		{
			name: "negative values get defaults",
			input: Config{
				DefaultLimit:    -1,
				MaxLimit:        -5,
				MaxLineageDepth: -10,
			},
			expected: Config{
				DefaultLimit:    10,
				MaxLimit:        100,
				MaxLineageDepth: 5,
			},
		},
		{
			name: "positive values preserved",
			input: Config{
				DefaultLimit:    20,
				MaxLimit:        200,
				MaxLineageDepth: 10,
			},
			expected: Config{
				DefaultLimit:    20,
				MaxLimit:        200,
				MaxLineageDepth: 10,
			},
		},
		{
			name: "partial values",
			input: Config{
				DefaultLimit:    50,
				MaxLimit:        0,
				MaxLineageDepth: 3,
			},
			expected: Config{
				DefaultLimit:    50,
				MaxLimit:        100,
				MaxLineageDepth: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeConfig(tt.input)

			if result.DefaultLimit != tt.expected.DefaultLimit {
				t.Errorf("normalizeConfig() DefaultLimit = %d, want %d", result.DefaultLimit, tt.expected.DefaultLimit)
			}
			if result.MaxLimit != tt.expected.MaxLimit {
				t.Errorf("normalizeConfig() MaxLimit = %d, want %d", result.MaxLimit, tt.expected.MaxLimit)
			}
			if result.MaxLineageDepth != tt.expected.MaxLineageDepth {
				t.Errorf("normalizeConfig() MaxLineageDepth = %d, want %d", result.MaxLineageDepth, tt.expected.MaxLineageDepth)
			}
		})
	}
}
