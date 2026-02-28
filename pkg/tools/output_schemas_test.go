package tools

import (
	"encoding/json"
	"testing"
)

func TestDefaultOutputSchema(t *testing.T) {
	tests := []struct {
		name     string
		toolName ToolName
		wantNil  bool
	}{
		{name: "known tool returns schema", toolName: ToolSearch, wantNil: false},
		{name: "unknown tool returns nil", toolName: ToolName("nonexistent_tool"), wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultOutputSchema(tt.toolName)
			if tt.wantNil && got != nil {
				t.Errorf("DefaultOutputSchema(%q) = %v, want nil", tt.toolName, got)
			}
			if !tt.wantNil && got == nil {
				t.Errorf("DefaultOutputSchema(%q) returned nil, want non-nil", tt.toolName)
			}
		})
	}
}

func TestDefaultOutputSchema_AllToolsCovered(t *testing.T) {
	allTools := append(AllTools(), WriteTools()...)
	for _, name := range allTools {
		schema := DefaultOutputSchema(name)
		if schema == nil {
			t.Errorf("no default output schema for tool %q", name)
		}
	}
}

func TestDefaultOutputSchema_ValidJSON(t *testing.T) {
	allTools := append(AllTools(), WriteTools()...)
	for _, name := range allTools {
		schema := DefaultOutputSchema(name)
		if schema == nil {
			continue
		}
		var parsed map[string]any
		if err := json.Unmarshal(schema, &parsed); err != nil {
			t.Errorf("tool %q has invalid JSON schema: %v", name, err)
		}
		if schemaType, ok := parsed["type"]; !ok || schemaType != "object" {
			t.Errorf("tool %q schema missing top-level type=object", name)
		}
	}
}

func TestGetOutputSchema_Priority(t *testing.T) {
	customSchema := json.RawMessage(`{"type":"object","properties":{"custom":{"type":"string"}}}`)

	tk := &Toolkit{
		outputSchemas:   map[ToolName]any{ToolSearch: customSchema},
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	tests := []struct {
		name     string
		cfg      *toolConfig
		wantType string // "per-reg", "toolkit", "default"
	}{
		{
			name:     "per-registration override takes highest priority",
			cfg:      &toolConfig{outputSchema: json.RawMessage(`{"type":"object"}`)},
			wantType: "per-reg",
		},
		{
			name:     "toolkit-level override takes middle priority",
			cfg:      nil,
			wantType: "toolkit",
		},
		{
			name:     "empty cfg falls through to toolkit override",
			cfg:      &toolConfig{},
			wantType: "toolkit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tk.getOutputSchema(ToolSearch, tt.cfg)
			if got == nil {
				t.Fatal("getOutputSchema() returned nil")
			}

			// Verify it's the right schema by checking it's not nil and valid
			switch tt.wantType {
			case "per-reg":
				perReg, ok := tt.cfg.outputSchema.(json.RawMessage)
				if !ok {
					t.Fatal("per-reg schema is not json.RawMessage")
				}
				gotBytes, ok2 := got.(json.RawMessage)
				if !ok2 {
					t.Fatal("result is not json.RawMessage")
				}
				if string(gotBytes) != string(perReg) {
					t.Errorf("getOutputSchema() = %s, want per-reg %s", gotBytes, perReg)
				}
			case "toolkit":
				gotBytes, ok := got.(json.RawMessage)
				if !ok {
					t.Fatal("result is not json.RawMessage")
				}
				if string(gotBytes) != string(customSchema) {
					t.Errorf("getOutputSchema() = %s, want toolkit schema", gotBytes)
				}
			}
		})
	}
}

func TestGetOutputSchema_DefaultFallback(t *testing.T) {
	tk := &Toolkit{
		outputSchemas:   make(map[ToolName]any),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	got := tk.getOutputSchema(ToolSearch, nil)
	if got == nil {
		t.Fatal("getOutputSchema() with no overrides returned nil")
	}

	want := defaultOutputSchemas[ToolSearch]
	gotBytes, ok := got.(json.RawMessage)
	if !ok {
		t.Fatal("result is not json.RawMessage")
	}
	if string(gotBytes) != string(want) {
		t.Errorf("getOutputSchema() default = %s, want %s", gotBytes, want)
	}
}
