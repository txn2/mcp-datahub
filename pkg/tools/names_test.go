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
		{ToolGetColumnLineage, "datahub_get_column_lineage"},
		{ToolGetQueries, "datahub_get_queries"},
		{ToolGetGlossaryTerm, "datahub_get_glossary_term"},
		{ToolListTags, "datahub_list_tags"},
		{ToolListDomains, "datahub_list_domains"},
		{ToolListDataProducts, "datahub_list_data_products"},
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

func TestAllTools(t *testing.T) {
	tools := AllTools()

	// Should return all 12 tools
	expectedCount := 12
	if len(tools) != expectedCount {
		t.Errorf("AllTools() count = %d, want %d", len(tools), expectedCount)
	}

	// Should contain all expected tools
	expectedTools := map[ToolName]bool{
		ToolSearch:           true,
		ToolGetEntity:        true,
		ToolGetSchema:        true,
		ToolGetLineage:       true,
		ToolGetColumnLineage: true,
		ToolGetQueries:       true,
		ToolGetGlossaryTerm:  true,
		ToolListTags:         true,
		ToolListDomains:      true,
		ToolListDataProducts: true,
		ToolGetDataProduct:   true,
		ToolListConnections:  true,
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
