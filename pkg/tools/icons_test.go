package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDefaultIcons(t *testing.T) {
	icons := DefaultIcons()
	if len(icons) != 1 {
		t.Fatalf("expected 1 default icon, got %d", len(icons))
	}
	if icons[0].Source == "" {
		t.Error("default icon source should not be empty")
	}
	if icons[0].MIMEType != "image/svg+xml" {
		t.Errorf("expected MIME type image/svg+xml, got %s", icons[0].MIMEType)
	}
}

func TestGetIcons_DefaultPriority(t *testing.T) {
	tk := newBaseToolkit(Config{})
	icons := tk.getIcons(ToolSearch, nil)
	if len(icons) != 1 {
		t.Fatalf("expected 1 icon, got %d", len(icons))
	}
	if icons[0].Source != defaultIcons[0].Source {
		t.Errorf("expected default icon source, got %s", icons[0].Source)
	}
}

func TestGetIcons_ToolkitOverride(t *testing.T) {
	tk := newBaseToolkit(Config{})
	customIcons := []mcp.Icon{{Source: "https://example.com/custom.svg", MIMEType: "image/svg+xml"}}
	tk.icons[ToolSearch] = customIcons

	// ToolSearch should get the override
	icons := tk.getIcons(ToolSearch, nil)
	if len(icons) != 1 {
		t.Fatalf("expected 1 icon, got %d", len(icons))
	}
	if icons[0].Source != "https://example.com/custom.svg" {
		t.Errorf("expected custom icon source, got %s", icons[0].Source)
	}

	// ToolGetEntity should get the default
	icons = tk.getIcons(ToolGetEntity, nil)
	if icons[0].Source != defaultIcons[0].Source {
		t.Errorf("expected default icon source for ToolGetEntity, got %s", icons[0].Source)
	}
}

func TestGetIcons_PerRegistrationOverride(t *testing.T) {
	tk := newBaseToolkit(Config{})
	tk.icons[ToolSearch] = []mcp.Icon{{Source: "https://example.com/toolkit.svg"}}

	cfg := &toolConfig{
		icons: []mcp.Icon{{Source: "https://example.com/registration.svg"}},
	}
	icons := tk.getIcons(ToolSearch, cfg)
	if len(icons) != 1 {
		t.Fatalf("expected 1 icon, got %d", len(icons))
	}
	if icons[0].Source != "https://example.com/registration.svg" {
		t.Errorf("expected per-registration icon source, got %s", icons[0].Source)
	}
}

func TestGetIcons_NilConfig(t *testing.T) {
	tk := newBaseToolkit(Config{})
	icons := tk.getIcons(ToolSearch, nil)
	if len(icons) == 0 {
		t.Fatal("expected default icons, got empty")
	}
}

func TestGetIcons_EmptyToolConfig(t *testing.T) {
	tk := newBaseToolkit(Config{})
	cfg := &toolConfig{}
	icons := tk.getIcons(ToolSearch, cfg)
	if len(icons) == 0 {
		t.Fatal("expected default icons, got empty")
	}
	if icons[0].Source != defaultIcons[0].Source {
		t.Errorf("expected default icon, got %s", icons[0].Source)
	}
}
