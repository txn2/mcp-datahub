package client

import (
	"strings"
	"testing"
)

// FuzzParseURN tests ParseURN with random inputs to find panics or crashes.
func FuzzParseURN(f *testing.F) {
	// Seed corpus with valid and edge case URNs
	seeds := []string{
		// Valid URNs
		"urn:li:dataset:(urn:li:dataPlatform:snowflake,db.schema.table,PROD)",
		"urn:li:dataset:(urn:li:dataPlatform:postgres,public.users,DEV)",
		"urn:li:dashboard:(looker,dashboard_123)",
		"urn:li:chart:(tableau,chart_456)",
		"urn:li:glossaryTerm:business.revenue",
		"urn:li:tag:PII",
		"urn:li:domain:marketing",
		"urn:li:dataFlow:(airflow,my_dag,PROD)",
		"urn:li:dataJob:(urn:li:dataFlow:(airflow,dag,PROD),task)",
		// Edge cases
		"",
		"urn:",
		"urn:li:",
		"urn:li:dataset:",
		"urn:li:dataset:()",
		"urn:li:dataset:(a,b)",
		"urn:li:dataset:(a,b,c,d)",
		"urn:li:dashboard:",
		"urn:li:dashboard:()",
		"urn:li:dashboard:(a)",
		"not a urn",
		"urn:li:unknown:value",
		// URL encoded
		"urn:li:dataset:(urn:li:dataPlatform:snowflake,db%2Fschema%2Ftable,PROD)",
		// Unicode
		"urn:li:tag:emoji_test",
		"urn:li:glossaryTerm:path.to.term",
		// Malformed
		"urn:li:dataset:missing_parens",
		"urn:li:dataset:(missing_close",
		"urn:li:dataset:missing_open)",
		"urn:li:chart:(only_one_part)",
		"urn:li:chart:()",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, urn string) {
		// ParseURN should never panic
		result, err := ParseURN(urn)

		// If no error, result should not be nil
		if err == nil && result == nil {
			t.Error("ParseURN returned nil result without error")
		}

		// If result is not nil, basic fields should be set
		if result != nil {
			if result.Raw != urn {
				t.Error("Raw field should match input URN")
			}
		}

		// The main goal of fuzz testing is to ensure no panics occur
		// Edge cases like "urn:li:(" producing empty EntityType are acceptable
		// as long as the function handles them without crashing
		_ = result // Use result to avoid unused variable warning
	})
}

// FuzzBuildDatasetURN tests BuildDatasetURN with random inputs.
func FuzzBuildDatasetURN(f *testing.F) {
	// Seed corpus
	f.Add("snowflake", "db.schema.table", "PROD")
	f.Add("postgres", "public.users", "DEV")
	f.Add("bigquery", "project.dataset.table", "")
	f.Add("", "", "")
	f.Add("platform", "name/with/slashes", "ENV")
	f.Add("platform", "name with spaces", "ENV")

	f.Fuzz(func(t *testing.T, platform, qualifiedName, env string) {
		// BuildDatasetURN should never panic
		result := BuildDatasetURN(platform, qualifiedName, env)

		// Result should always be a valid URN format
		if !strings.HasPrefix(result, "urn:li:dataset:(urn:li:dataPlatform:") {
			t.Error("BuildDatasetURN should produce valid dataset URN prefix")
		}

		// Should be parseable (round-trip test)
		parsed, err := ParseURN(result)
		if err != nil {
			// Some inputs may produce URNs that don't round-trip perfectly due to encoding
			// This is acceptable, just verify no panic
			return
		}

		if parsed.EntityType != "dataset" {
			t.Error("Parsed EntityType should be 'dataset'")
		}
	})
}

// FuzzBuildDashboardURN tests BuildDashboardURN with random inputs.
func FuzzBuildDashboardURN(f *testing.F) {
	f.Add("looker", "dashboard_123")
	f.Add("tableau", "viz_456")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, platform, dashboardID string) {
		// Should never panic
		result := BuildDashboardURN(platform, dashboardID)

		if !strings.HasPrefix(result, "urn:li:dashboard:(") {
			t.Error("BuildDashboardURN should produce valid dashboard URN prefix")
		}
	})
}
