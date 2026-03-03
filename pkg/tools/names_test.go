package tools

import (
	"testing"
)

func TestToolNameConstants(t *testing.T) {
	// Verify tool name format is consistent
	names := []struct {
		name     ToolName
		expected string
	}{
		{ToolSearch, "datahub_search"},
		{ToolGetEntity, "datahub_get_entity"},
		{ToolGetSchema, "datahub_get_schema"},
		{ToolGetLineage, "datahub_get_lineage"},
		{ToolGetQueries, "datahub_get_queries"},
		{ToolBrowse, "datahub_browse"},
		{ToolGetGlossaryTerm, "datahub_get_glossary_term"},
		{ToolGetDataProduct, "datahub_get_data_product"},
		{ToolListConnections, "datahub_list_connections"},
	}

	for _, tc := range names {
		t.Run(string(tc.name), func(t *testing.T) {
			if string(tc.name) != tc.expected {
				t.Errorf("ToolName = %s, want %s", tc.name, tc.expected)
			}
		})
	}
}

func TestDeprecatedToolNameConstants(t *testing.T) {
	// Verify deprecated tool names are preserved for backward compatibility
	names := []struct {
		name     ToolName
		expected string
	}{
		{ToolListTags, "datahub_list_tags"},
		{ToolListDomains, "datahub_list_domains"},
		{ToolListDataProducts, "datahub_list_data_products"},
		{ToolGetColumnLineage, "datahub_get_column_lineage"},
	}

	for _, tc := range names {
		t.Run(string(tc.name), func(t *testing.T) {
			if string(tc.name) != tc.expected {
				t.Errorf("ToolName = %s, want %s", tc.name, tc.expected)
			}
		})
	}
}

func TestAllTools(t *testing.T) {
	tools := AllTools()

	// Should return 9 primary read tools (excludes deprecated)
	expectedCount := 9
	if len(tools) != expectedCount {
		t.Errorf("AllTools() count = %d, want %d", len(tools), expectedCount)
	}

	// Should contain all expected tools
	expectedTools := map[ToolName]bool{
		ToolSearch:          true,
		ToolGetEntity:       true,
		ToolGetSchema:       true,
		ToolGetLineage:      true,
		ToolGetQueries:      true,
		ToolBrowse:          true,
		ToolGetGlossaryTerm: true,
		ToolGetDataProduct:  true,
		ToolListConnections: true,
	}

	for _, tool := range tools {
		if !expectedTools[tool] {
			t.Errorf("AllTools() contains unexpected tool: %s", tool)
		}
		delete(expectedTools, tool)
	}

	for tool := range expectedTools {
		t.Errorf("AllTools() missing tool: %s", tool)
	}
}

func TestDeprecatedTools(t *testing.T) {
	tools := DeprecatedTools()

	expectedCount := 4
	if len(tools) != expectedCount {
		t.Errorf("DeprecatedTools() count = %d, want %d", len(tools), expectedCount)
	}

	expectedTools := map[ToolName]bool{
		ToolListTags:         true,
		ToolListDomains:      true,
		ToolListDataProducts: true,
		ToolGetColumnLineage: true,
	}

	for _, tool := range tools {
		if !expectedTools[tool] {
			t.Errorf("DeprecatedTools() contains unexpected tool: %s", tool)
		}
		delete(expectedTools, tool)
	}

	for tool := range expectedTools {
		t.Errorf("DeprecatedTools() missing tool: %s", tool)
	}
}

func TestAllToolsNoDuplicates(t *testing.T) {
	tools := AllTools()
	seen := make(map[ToolName]bool)

	for _, tool := range tools {
		if seen[tool] {
			t.Errorf("AllTools() contains duplicate: %s", tool)
		}
		seen[tool] = true
	}
}

func TestAllToolsAndDeprecatedNoOverlap(t *testing.T) {
	allSet := make(map[ToolName]bool)
	for _, tool := range AllTools() {
		allSet[tool] = true
	}

	for _, tool := range DeprecatedTools() {
		if allSet[tool] {
			t.Errorf("DeprecatedTools() overlaps with AllTools(): %s", tool)
		}
	}
}
