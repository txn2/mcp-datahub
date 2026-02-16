package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDefaultAnnotations(t *testing.T) {
	tests := []struct {
		name    ToolName
		wantNil bool
	}{
		{ToolSearch, false},
		{ToolGetEntity, false},
		{ToolGetSchema, false},
		{ToolGetLineage, false},
		{ToolGetColumnLineage, false},
		{ToolGetQueries, false},
		{ToolGetGlossaryTerm, false},
		{ToolListTags, false},
		{ToolListDomains, false},
		{ToolListDataProducts, false},
		{ToolGetDataProduct, false},
		{ToolListConnections, false},
		{ToolUpdateDescription, false},
		{ToolAddTag, false},
		{ToolRemoveTag, false},
		{ToolAddGlossaryTerm, false},
		{ToolRemoveGlossaryTerm, false},
		{ToolAddLink, false},
		{ToolRemoveLink, false},
		{ToolName("unknown_tool"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			ann := DefaultAnnotations(tt.name)
			if tt.wantNil && ann != nil {
				t.Errorf("expected nil annotations for %s", tt.name)
			}
			if !tt.wantNil && ann == nil {
				t.Errorf("expected non-nil annotations for %s", tt.name)
			}
		})
	}
}

func TestDefaultAnnotations_AllToolsCovered(t *testing.T) {
	for _, name := range AllTools() {
		if DefaultAnnotations(name) == nil {
			t.Errorf("tool %s has no default annotations", name)
		}
	}
}

func TestDefaultAnnotations_ReadOnlyTools(t *testing.T) {
	readOnlyTools := []ToolName{
		ToolSearch, ToolGetEntity, ToolGetSchema, ToolGetLineage,
		ToolGetColumnLineage, ToolGetQueries, ToolGetGlossaryTerm,
		ToolListTags, ToolListDomains, ToolListDataProducts,
		ToolGetDataProduct, ToolListConnections,
	}

	for _, name := range readOnlyTools {
		t.Run(string(name), func(t *testing.T) {
			ann := DefaultAnnotations(name)
			if !ann.ReadOnlyHint {
				t.Errorf("expected ReadOnlyHint=true for %s", name)
			}
			if !ann.IdempotentHint {
				t.Errorf("expected IdempotentHint=true for %s", name)
			}
			if ann.OpenWorldHint == nil || *ann.OpenWorldHint {
				t.Errorf("expected OpenWorldHint=false for %s", name)
			}
		})
	}
}

func TestDefaultAnnotations_WriteTools(t *testing.T) {
	writeTools := []ToolName{
		ToolUpdateDescription, ToolAddTag, ToolRemoveTag,
		ToolAddGlossaryTerm, ToolRemoveGlossaryTerm,
		ToolAddLink, ToolRemoveLink,
	}

	for _, name := range writeTools {
		t.Run(string(name), func(t *testing.T) {
			ann := DefaultAnnotations(name)
			if ann.ReadOnlyHint {
				t.Errorf("expected ReadOnlyHint=false for %s", name)
			}
			if ann.DestructiveHint == nil || *ann.DestructiveHint {
				t.Errorf("expected DestructiveHint=false for %s", name)
			}
			if !ann.IdempotentHint {
				t.Errorf("expected IdempotentHint=true for %s", name)
			}
			if ann.OpenWorldHint == nil || *ann.OpenWorldHint {
				t.Errorf("expected OpenWorldHint=false for %s", name)
			}
		})
	}
}

func TestGetAnnotations_Priority(t *testing.T) {
	tests := []struct {
		name         string
		toolkitAnns  map[ToolName]*mcp.ToolAnnotations
		cfgAnns      *mcp.ToolAnnotations
		toolName     ToolName
		wantReadOnly bool
	}{
		{
			name:         "default only",
			toolkitAnns:  nil,
			cfgAnns:      nil,
			toolName:     ToolSearch,
			wantReadOnly: true,
		},
		{
			name: "toolkit override",
			toolkitAnns: map[ToolName]*mcp.ToolAnnotations{
				ToolSearch: {ReadOnlyHint: false},
			},
			cfgAnns:      nil,
			toolName:     ToolSearch,
			wantReadOnly: false,
		},
		{
			name:         "per-registration override",
			toolkitAnns:  nil,
			cfgAnns:      &mcp.ToolAnnotations{ReadOnlyHint: false},
			toolName:     ToolSearch,
			wantReadOnly: false,
		},
		{
			name: "per-registration beats toolkit",
			toolkitAnns: map[ToolName]*mcp.ToolAnnotations{
				ToolSearch: {ReadOnlyHint: true},
			},
			cfgAnns:      &mcp.ToolAnnotations{ReadOnlyHint: false},
			toolName:     ToolSearch,
			wantReadOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &Toolkit{
				annotations: make(map[ToolName]*mcp.ToolAnnotations),
			}
			if tt.toolkitAnns != nil {
				tk.annotations = tt.toolkitAnns
			}

			var cfg *toolConfig
			if tt.cfgAnns != nil {
				cfg = &toolConfig{annotations: tt.cfgAnns}
			}

			got := tk.getAnnotations(tt.toolName, cfg)
			if got == nil {
				t.Fatal("getAnnotations() returned nil")
			}
			if got.ReadOnlyHint != tt.wantReadOnly {
				t.Errorf("ReadOnlyHint = %v, want %v",
					got.ReadOnlyHint, tt.wantReadOnly)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	truePtr := boolPtr(true)
	falsePtr := boolPtr(false)

	if truePtr == nil || !*truePtr {
		t.Error("boolPtr(true) should return *true")
	}
	if falsePtr == nil || *falsePtr {
		t.Error("boolPtr(false) should return *false")
	}
}
