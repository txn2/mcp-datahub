package tools

import "testing"

func strPtr(s string) *string {
	return &s
}

func TestDefaultDescription(t *testing.T) {
	tests := []struct {
		name     string
		toolName ToolName
		wantDesc string
	}{
		{
			name:     "known tool returns description",
			toolName: ToolSearch,
			wantDesc: defaultDescriptions[ToolSearch],
		},
		{
			name:     "unknown tool returns empty string",
			toolName: ToolName("nonexistent_tool"),
			wantDesc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultDescription(tt.toolName)
			if got != tt.wantDesc {
				t.Errorf("DefaultDescription(%q) = %q, want %q", tt.toolName, got, tt.wantDesc)
			}
		})
	}
}

func TestDefaultDescriptions_AllToolsCovered(t *testing.T) {
	allTools := append(AllTools(), WriteTools()...)
	for _, name := range allTools {
		desc := DefaultDescription(name)
		if desc == "" {
			t.Errorf("no default description for tool %q", name)
		}
	}
}

func TestGetDescription_Priority(t *testing.T) {
	toolkitDesc := "toolkit-level description"
	perRegDesc := "per-registration description"

	tk := &Toolkit{
		descriptions:    map[ToolName]string{ToolSearch: toolkitDesc},
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	tests := []struct {
		name string
		cfg  *toolConfig
		want string
	}{
		{
			name: "per-registration override takes highest priority",
			cfg:  &toolConfig{description: strPtr(perRegDesc)},
			want: perRegDesc,
		},
		{
			name: "toolkit-level override takes middle priority",
			cfg:  nil,
			want: toolkitDesc,
		},
		{
			name: "toolkit-level with nil description in cfg",
			cfg:  &toolConfig{},
			want: toolkitDesc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tk.getDescription(ToolSearch, tt.cfg)
			if got != tt.want {
				t.Errorf("getDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetDescription_NilConfig(t *testing.T) {
	// Toolkit with no overrides should return the default
	tk := &Toolkit{
		descriptions:    make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	got := tk.getDescription(ToolSearch, nil)
	want := defaultDescriptions[ToolSearch]
	if got != want {
		t.Errorf("getDescription() with no overrides = %q, want default", got)
	}
}

func TestGetDescription_DefaultFallback(t *testing.T) {
	// Toolkit with no overrides and empty config falls back to default
	tk := &Toolkit{
		descriptions:    make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	got := tk.getDescription(ToolGetEntity, &toolConfig{})
	want := defaultDescriptions[ToolGetEntity]
	if got != want {
		t.Errorf("getDescription() fallback = %q, want %q", got, want)
	}
}
