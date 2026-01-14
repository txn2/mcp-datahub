package client

import (
	"testing"
)

func TestBuildDatasetURN(t *testing.T) {
	tests := []struct {
		name          string
		platform      string
		qualifiedName string
		env           string
		want          string
	}{
		{
			name:          "basic dataset",
			platform:      "snowflake",
			qualifiedName: "my_db.my_schema.my_table",
			env:           "PROD",
			want:          "urn:li:dataset:(urn:li:dataPlatform:snowflake,my_db.my_schema.my_table,PROD)",
		},
		{
			name:          "empty env defaults to PROD",
			platform:      "postgres",
			qualifiedName: "public.users",
			env:           "",
			want:          "urn:li:dataset:(urn:li:dataPlatform:postgres,public.users,PROD)",
		},
		{
			name:          "dev environment",
			platform:      "bigquery",
			qualifiedName: "project.dataset.table",
			env:           "DEV",
			want:          "urn:li:dataset:(urn:li:dataPlatform:bigquery,project.dataset.table,DEV)",
		},
		{
			name:          "name with special characters",
			platform:      "snowflake",
			qualifiedName: "db/schema/table with spaces",
			env:           "PROD",
			want:          "urn:li:dataset:(urn:li:dataPlatform:snowflake,db%2Fschema%2Ftable%20with%20spaces,PROD)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDatasetURN(tt.platform, tt.qualifiedName, tt.env)
			if got != tt.want {
				t.Errorf("BuildDatasetURN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildDashboardURN(t *testing.T) {
	tests := []struct {
		name        string
		platform    string
		dashboardID string
		want        string
	}{
		{
			name:        "looker dashboard",
			platform:    "looker",
			dashboardID: "123",
			want:        "urn:li:dashboard:(looker,123)",
		},
		{
			name:        "tableau dashboard",
			platform:    "tableau",
			dashboardID: "workbook/view",
			want:        "urn:li:dashboard:(tableau,workbook/view)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDashboardURN(tt.platform, tt.dashboardID)
			if got != tt.want {
				t.Errorf("BuildDashboardURN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildChartURN(t *testing.T) {
	got := BuildChartURN("looker", "456")
	want := "urn:li:chart:(looker,456)"
	if got != want {
		t.Errorf("BuildChartURN() = %v, want %v", got, want)
	}
}

func TestBuildDataFlowURN(t *testing.T) {
	got := BuildDataFlowURN("airflow", "my_dag", "production")
	want := "urn:li:dataFlow:(airflow,my_dag,production)"
	if got != want {
		t.Errorf("BuildDataFlowURN() = %v, want %v", got, want)
	}
}

func TestBuildDataJobURN(t *testing.T) {
	dataFlowURN := "urn:li:dataFlow:(airflow,my_dag,production)"
	got := BuildDataJobURN(dataFlowURN, "task_1")
	want := "urn:li:dataJob:(urn:li:dataFlow:(airflow,my_dag,production),task_1)"
	if got != want {
		t.Errorf("BuildDataJobURN() = %v, want %v", got, want)
	}
}

func TestBuildGlossaryTermURN(t *testing.T) {
	got := BuildGlossaryTermURN("business.revenue")
	want := "urn:li:glossaryTerm:business.revenue"
	if got != want {
		t.Errorf("BuildGlossaryTermURN() = %v, want %v", got, want)
	}
}

func TestBuildTagURN(t *testing.T) {
	got := BuildTagURN("PII")
	want := "urn:li:tag:PII"
	if got != want {
		t.Errorf("BuildTagURN() = %v, want %v", got, want)
	}
}

func TestBuildDomainURN(t *testing.T) {
	got := BuildDomainURN("marketing")
	want := "urn:li:domain:marketing"
	if got != want {
		t.Errorf("BuildDomainURN() = %v, want %v", got, want)
	}
}

func TestParseURN(t *testing.T) {
	tests := []struct {
		name        string
		urn         string
		wantType    string
		wantPlatf   string
		wantName    string
		wantEnv     string
		wantErr     bool
		errContains string
	}{
		{
			name:      "dataset URN",
			urn:       "urn:li:dataset:(urn:li:dataPlatform:snowflake,my_db.my_schema.my_table,PROD)",
			wantType:  "dataset",
			wantPlatf: "snowflake",
			wantName:  "my_db.my_schema.my_table",
			wantEnv:   "PROD",
		},
		{
			name:      "dataset URN with encoded name",
			urn:       "urn:li:dataset:(urn:li:dataPlatform:snowflake,db%2Fschema%2Ftable,PROD)",
			wantType:  "dataset",
			wantPlatf: "snowflake",
			wantName:  "db/schema/table",
			wantEnv:   "PROD",
		},
		{
			name:      "dashboard URN",
			urn:       "urn:li:dashboard:(looker,123)",
			wantType:  "dashboard",
			wantPlatf: "looker",
			wantName:  "123",
		},
		{
			name:      "chart URN",
			urn:       "urn:li:chart:(tableau,viz_1)",
			wantType:  "chart",
			wantPlatf: "tableau",
			wantName:  "viz_1",
		},
		{
			name:     "tag URN",
			urn:      "urn:li:tag:PII",
			wantType: "tag",
			wantName: "PII",
		},
		{
			name:     "glossary term URN",
			urn:      "urn:li:glossaryTerm:business.revenue",
			wantType: "glossaryTerm",
			wantName: "business.revenue",
		},
		{
			name:     "domain URN",
			urn:      "urn:li:domain:marketing",
			wantType: "domain",
			wantName: "marketing",
		},
		{
			name:        "invalid prefix",
			urn:         "invalid:urn",
			wantErr:     true,
			errContains: "must start with 'urn:li:'",
		},
		{
			name:        "dataset missing parentheses",
			urn:         "urn:li:dataset:invalid",
			wantErr:     true,
			errContains: "must have parentheses",
		},
		{
			name:        "dataset wrong part count",
			urn:         "urn:li:dataset:(platform,name)",
			wantErr:     true,
			errContains: "must have 3 parts",
		},
		{
			name:        "dataset invalid platform",
			urn:         "urn:li:dataset:(invalid,name,PROD)",
			wantErr:     true,
			errContains: "invalid platform URN",
		},
		{
			name:        "dashboard wrong part count",
			urn:         "urn:li:dashboard:(platform)",
			wantErr:     true,
			errContains: "must have 2 parts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseURN(tt.urn)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseURN() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseURN() error = %v, want containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseURN() unexpected error: %v", err)
				return
			}

			if parsed.EntityType != tt.wantType {
				t.Errorf("ParseURN() EntityType = %v, want %v", parsed.EntityType, tt.wantType)
			}
			if parsed.Platform != tt.wantPlatf {
				t.Errorf("ParseURN() Platform = %v, want %v", parsed.Platform, tt.wantPlatf)
			}
			if parsed.Name != tt.wantName {
				t.Errorf("ParseURN() Name = %v, want %v", parsed.Name, tt.wantName)
			}
			if tt.wantEnv != "" && parsed.Env != tt.wantEnv {
				t.Errorf("ParseURN() Env = %v, want %v", parsed.Env, tt.wantEnv)
			}
			if parsed.Raw != tt.urn {
				t.Errorf("ParseURN() Raw = %v, want %v", parsed.Raw, tt.urn)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
