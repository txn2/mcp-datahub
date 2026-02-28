package tools

import "testing"

func TestDefaultTitle(t *testing.T) {
	tests := []struct {
		name      string
		toolName  ToolName
		wantEmpty bool
	}{
		{name: "known tool returns title", toolName: ToolSearch, wantEmpty: false},
		{name: "unknown tool returns empty", toolName: ToolName("nonexistent_tool"), wantEmpty: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultTitle(tt.toolName)
			if tt.wantEmpty && got != "" {
				t.Errorf("DefaultTitle(%q) = %q, want empty string", tt.toolName, got)
			}
			if !tt.wantEmpty && got == "" {
				t.Errorf("DefaultTitle(%q) returned empty string, want non-empty", tt.toolName)
			}
		})
	}
}

func TestDefaultTitle_AllToolsCovered(t *testing.T) {
	allTools := append(AllTools(), WriteTools()...)
	for _, name := range allTools {
		title := DefaultTitle(name)
		if title == "" {
			t.Errorf("no default title for tool %q", name)
		}
	}
}

func TestGetTitle_Priority(t *testing.T) {
	toolkitTitle := "toolkit-level title"
	perRegTitle := "per-registration title"

	tk := &Toolkit{
		titles:          map[ToolName]string{ToolSearch: toolkitTitle},
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
			cfg:  &toolConfig{title: strPtr(perRegTitle)},
			want: perRegTitle,
		},
		{
			name: "toolkit-level override takes middle priority",
			cfg:  nil,
			want: toolkitTitle,
		},
		{
			name: "toolkit-level with nil title in cfg",
			cfg:  &toolConfig{},
			want: toolkitTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tk.getTitle(ToolSearch, tt.cfg)
			if got != tt.want {
				t.Errorf("getTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTitle_DefaultFallback(t *testing.T) {
	tk := &Toolkit{
		titles:          make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	got := tk.getTitle(ToolSearch, nil)
	want := defaultTitles[ToolSearch]
	if got != want {
		t.Errorf("getTitle() with no overrides = %q, want %q", got, want)
	}
}

func TestGetTitle_NilConfig(t *testing.T) {
	tk := &Toolkit{
		titles:          make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	got := tk.getTitle(ToolGetEntity, &toolConfig{})
	want := defaultTitles[ToolGetEntity]
	if got != want {
		t.Errorf("getTitle() with empty cfg = %q, want %q", got, want)
	}
}
