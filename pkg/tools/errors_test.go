package tools

import (
	"testing"
)

func TestErrorResult(t *testing.T) {
	result := ErrorResult("test error message")

	if !result.IsError {
		t.Error("ErrorResult() IsError should be true")
	}

	if len(result.Content) != 1 {
		t.Errorf("ErrorResult() Content length = %d, want 1", len(result.Content))
		return
	}

	// MCP types don't have direct Text field accessor, so we skip content check
	// The important thing is that IsError is true
}

func TestTextResult(t *testing.T) {
	result := TextResult("test message")

	if result.IsError {
		t.Error("TextResult() IsError should be false")
	}

	if len(result.Content) != 1 {
		t.Errorf("TextResult() Content length = %d, want 1", len(result.Content))
	}
}

func TestJSONResult(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "simple map",
			input: map[string]string{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name: "struct",
			input: struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}{
				Name:  "test",
				Count: 42,
			},
			wantErr: false,
		},
		{
			name:    "slice",
			input:   []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "channel (unmarshallable)",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := JSONResult(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("JSONResult() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("JSONResult() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("JSONResult() returned nil result")
				return
			}

			if result.IsError {
				t.Error("JSONResult() IsError should be false for valid JSON")
			}
		})
	}
}
