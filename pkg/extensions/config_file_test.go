package extensions

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/txn2/mcp-datahub/pkg/tools"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestFromBytes_JSON(t *testing.T) {
	data := []byte(`{
		"datahub": {
			"url": "https://datahub.example.com",
			"token": "test-token",
			"timeout": "45s",
			"connection_name": "prod",
			"write_enabled": true
		},
		"toolkit": {
			"default_limit": 20,
			"max_limit": 200,
			"max_lineage_depth": 10
		},
		"extensions": {
			"logging": true,
			"metrics": false,
			"errors": true
		}
	}`)

	cfg, err := FromBytes(data, "json")
	if err != nil {
		t.Fatalf("FromBytes() error: %v", err)
	}

	if cfg.DataHub.URL != "https://datahub.example.com" {
		t.Errorf("URL = %q, want %q", cfg.DataHub.URL, "https://datahub.example.com")
	}
	if cfg.DataHub.Token != "test-token" {
		t.Errorf("Token = %q, want %q", cfg.DataHub.Token, "test-token")
	}
	if cfg.DataHub.Timeout.Duration != 45*time.Second {
		t.Errorf("Timeout = %v, want 45s", cfg.DataHub.Timeout.Duration)
	}
	if cfg.DataHub.ConnectionName != "prod" {
		t.Errorf("ConnectionName = %q, want %q", cfg.DataHub.ConnectionName, "prod")
	}
	if cfg.DataHub.WriteEnabled == nil || !*cfg.DataHub.WriteEnabled {
		t.Error("WriteEnabled should be true")
	}
	if cfg.Toolkit.DefaultLimit != 20 {
		t.Errorf("DefaultLimit = %d, want 20", cfg.Toolkit.DefaultLimit)
	}
	if cfg.Toolkit.MaxLimit != 200 {
		t.Errorf("MaxLimit = %d, want 200", cfg.Toolkit.MaxLimit)
	}
	if cfg.Toolkit.MaxLineageDepth != 10 {
		t.Errorf("MaxLineageDepth = %d, want 10", cfg.Toolkit.MaxLineageDepth)
	}
	if cfg.Extensions.Logging == nil || !*cfg.Extensions.Logging {
		t.Error("Logging should be true")
	}
	if cfg.Extensions.Metrics == nil || *cfg.Extensions.Metrics {
		t.Error("Metrics should be false")
	}
}

func TestFromBytes_YAML(t *testing.T) {
	data := []byte(`
datahub:
  url: https://datahub.example.com
  token: my-token
  timeout: "1m"
  write_enabled: false

toolkit:
  default_limit: 15
  descriptions:
    datahub_search: "Custom search description"
    datahub_get_entity: "Custom entity description"

extensions:
  logging: true
  metadata: true
`)

	cfg, err := FromBytes(data, "yaml")
	if err != nil {
		t.Fatalf("FromBytes() error: %v", err)
	}

	if cfg.DataHub.URL != "https://datahub.example.com" {
		t.Errorf("URL = %q", cfg.DataHub.URL)
	}
	if cfg.DataHub.Timeout.Duration != time.Minute {
		t.Errorf("Timeout = %v, want 1m", cfg.DataHub.Timeout.Duration)
	}
	if cfg.DataHub.WriteEnabled == nil || *cfg.DataHub.WriteEnabled {
		t.Error("WriteEnabled should be false")
	}
	if cfg.Toolkit.DefaultLimit != 15 {
		t.Errorf("DefaultLimit = %d, want 15", cfg.Toolkit.DefaultLimit)
	}
	if cfg.Toolkit.Descriptions["datahub_search"] != "Custom search description" {
		t.Errorf("Descriptions[search] = %q", cfg.Toolkit.Descriptions["datahub_search"])
	}
	if cfg.Extensions.Logging == nil || !*cfg.Extensions.Logging {
		t.Error("Logging should be true")
	}
	if cfg.Extensions.Metadata == nil || !*cfg.Extensions.Metadata {
		t.Error("Metadata should be true")
	}
}

func TestFromBytes_InvalidFormat(t *testing.T) {
	_, err := FromBytes([]byte("{}"), "toml")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestFromBytes_InvalidJSON(t *testing.T) {
	_, err := FromBytes([]byte("{invalid json"), "json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFromBytes_InvalidYAML(t *testing.T) {
	_, err := FromBytes([]byte("\t\tinvalid:\nyaml: ["), "yaml")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a YAML config file
	yamlPath := filepath.Join(tmpDir, "config.yaml")
	yamlData := []byte(`
datahub:
  url: https://test.datahub.io
  token: file-token
toolkit:
  default_limit: 25
`)
	if err := os.WriteFile(yamlPath, yamlData, 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := FromFile(yamlPath)
	if err != nil {
		t.Fatalf("FromFile() error: %v", err)
	}

	if cfg.DataHub.URL != "https://test.datahub.io" {
		t.Errorf("URL = %q", cfg.DataHub.URL)
	}
	if cfg.Toolkit.DefaultLimit != 25 {
		t.Errorf("DefaultLimit = %d, want 25", cfg.Toolkit.DefaultLimit)
	}
}

func TestFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()

	jsonPath := filepath.Join(tmpDir, "config.json")
	jsonData := []byte(`{"datahub": {"url": "https://json.datahub.io"}}`)
	if err := os.WriteFile(jsonPath, jsonData, 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := FromFile(jsonPath)
	if err != nil {
		t.Fatalf("FromFile() error: %v", err)
	}

	if cfg.DataHub.URL != "https://json.datahub.io" {
		t.Errorf("URL = %q", cfg.DataHub.URL)
	}
}

func TestFromFile_YML(t *testing.T) {
	tmpDir := t.TempDir()

	ymlPath := filepath.Join(tmpDir, "config.yml")
	ymlData := []byte(`datahub:
  url: https://yml.datahub.io
`)
	if err := os.WriteFile(ymlPath, ymlData, 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := FromFile(ymlPath)
	if err != nil {
		t.Fatalf("FromFile() error: %v", err)
	}

	if cfg.DataHub.URL != "https://yml.datahub.io" {
		t.Errorf("URL = %q", cfg.DataHub.URL)
	}
}

func TestFromFile_UnsupportedExtension(t *testing.T) {
	tmpDir := t.TempDir()

	tomlPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(tomlPath, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := FromFile(tomlPath)
	if err == nil {
		t.Error("expected error for unsupported extension")
	}
}

func TestFromFile_NonExistent(t *testing.T) {
	_, err := FromFile("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	yamlPath := filepath.Join(tmpDir, "config.yaml")
	yamlData := []byte(`
datahub:
  url: https://file.datahub.io
  token: file-token
  timeout: "10s"
  connection_name: file-conn
`)
	if err := os.WriteFile(yamlPath, yamlData, 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set env overrides
	envVars := map[string]string{
		"DATAHUB_URL":             "https://env.datahub.io",
		"DATAHUB_TOKEN":           "env-token",
		"DATAHUB_TIMEOUT":         "30s",
		"DATAHUB_CONNECTION_NAME": "env-conn",
		"DATAHUB_WRITE_ENABLED":   "true",
	}

	originals := make(map[string]string)
	for k, v := range envVars {
		originals[k] = os.Getenv(k)
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("failed to set %s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for k, v := range originals {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	})

	cfg, err := LoadConfig(yamlPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.DataHub.URL != "https://env.datahub.io" {
		t.Errorf("URL = %q, want env override", cfg.DataHub.URL)
	}
	if cfg.DataHub.Token != "env-token" {
		t.Errorf("Token = %q, want env override", cfg.DataHub.Token)
	}
	if cfg.DataHub.Timeout.Duration != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", cfg.DataHub.Timeout.Duration)
	}
	if cfg.DataHub.ConnectionName != "env-conn" {
		t.Errorf("ConnectionName = %q, want env override", cfg.DataHub.ConnectionName)
	}
	if cfg.DataHub.WriteEnabled == nil || !*cfg.DataHub.WriteEnabled {
		t.Error("WriteEnabled should be true from env")
	}
}

func TestLoadConfig_InvalidTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	yamlPath := filepath.Join(tmpDir, "config.yaml")
	yamlData := []byte(`datahub:
  url: https://test.datahub.io
`)
	if err := os.WriteFile(yamlPath, yamlData, 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	orig := os.Getenv("DATAHUB_TIMEOUT")
	if err := os.Setenv("DATAHUB_TIMEOUT", "invalid"); err != nil {
		t.Fatalf("failed to set DATAHUB_TIMEOUT: %v", err)
	}
	t.Cleanup(func() {
		if orig == "" {
			_ = os.Unsetenv("DATAHUB_TIMEOUT")
		} else {
			_ = os.Setenv("DATAHUB_TIMEOUT", orig)
		}
	})

	_, err := LoadConfig(yamlPath)
	if err == nil {
		t.Error("expected error for invalid timeout")
	}
}

func TestLoadConfig_EnvExpansion(t *testing.T) {
	tmpDir := t.TempDir()

	yamlPath := filepath.Join(tmpDir, "config.yaml")
	yamlData := []byte(`
datahub:
  url: https://test.datahub.io
  token: "${TEST_DH_SECRET_TOKEN}"
`)
	if err := os.WriteFile(yamlPath, yamlData, 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Clear env vars that would override
	for _, k := range []string{"DATAHUB_URL", "DATAHUB_TOKEN", "DATAHUB_TIMEOUT", "DATAHUB_CONNECTION_NAME", "DATAHUB_WRITE_ENABLED"} {
		orig := os.Getenv(k)
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("failed to unset %s: %v", k, err)
		}
		t.Cleanup(func() {
			if orig != "" {
				_ = os.Setenv(k, orig)
			}
		})
	}

	orig := os.Getenv("TEST_DH_SECRET_TOKEN")
	if err := os.Setenv("TEST_DH_SECRET_TOKEN", "expanded-secret"); err != nil {
		t.Fatalf("failed to set TEST_DH_SECRET_TOKEN: %v", err)
	}
	t.Cleanup(func() {
		if orig == "" {
			_ = os.Unsetenv("TEST_DH_SECRET_TOKEN")
		} else {
			_ = os.Setenv("TEST_DH_SECRET_TOKEN", orig)
		}
	})

	cfg, err := LoadConfig(yamlPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.DataHub.Token != "expanded-secret" {
		t.Errorf("Token = %q, want expanded env var", cfg.DataHub.Token)
	}
}

func TestClientConfig(t *testing.T) {
	sc := ServerConfig{
		DataHub: DataHubConfig{
			URL:     "https://test.datahub.io",
			Token:   "test-token",
			Timeout: Duration{Duration: 45 * time.Second},
		},
		Toolkit: ToolkitConfig{
			DefaultLimit:    20,
			MaxLimit:        200,
			MaxLineageDepth: 8,
		},
	}

	cfg := sc.ClientConfig()

	if cfg.URL != "https://test.datahub.io" {
		t.Errorf("URL = %q", cfg.URL)
	}
	if cfg.Token != "test-token" {
		t.Errorf("Token = %q", cfg.Token)
	}
	if cfg.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v", cfg.Timeout)
	}
	if cfg.DefaultLimit != 20 {
		t.Errorf("DefaultLimit = %d", cfg.DefaultLimit)
	}
	if cfg.MaxLimit != 200 {
		t.Errorf("MaxLimit = %d", cfg.MaxLimit)
	}
	if cfg.MaxLineageDepth != 8 {
		t.Errorf("MaxLineageDepth = %d", cfg.MaxLineageDepth)
	}
}

func TestToolsConfig(t *testing.T) {
	we := true
	sc := ServerConfig{
		DataHub: DataHubConfig{
			WriteEnabled: &we,
		},
		Toolkit: ToolkitConfig{
			DefaultLimit:    15,
			MaxLimit:        150,
			MaxLineageDepth: 3,
		},
	}

	cfg := sc.ToolsConfig()

	if cfg.DefaultLimit != 15 {
		t.Errorf("DefaultLimit = %d", cfg.DefaultLimit)
	}
	if cfg.MaxLimit != 150 {
		t.Errorf("MaxLimit = %d", cfg.MaxLimit)
	}
	if cfg.MaxLineageDepth != 3 {
		t.Errorf("MaxLineageDepth = %d", cfg.MaxLineageDepth)
	}
	if !cfg.WriteEnabled {
		t.Error("WriteEnabled should be true")
	}
}

func TestToolsConfig_NoWriteEnabled(t *testing.T) {
	sc := ServerConfig{}
	cfg := sc.ToolsConfig()
	if cfg.WriteEnabled {
		t.Error("WriteEnabled should be false when nil")
	}
}

func TestExtConfig(t *testing.T) {
	sc := ServerConfig{
		Extensions: ExtFileConfig{
			Logging:  boolPtr(true),
			Metrics:  boolPtr(true),
			Metadata: boolPtr(false),
			Errors:   boolPtr(false),
		},
	}

	cfg := sc.ExtConfig()

	if !cfg.EnableLogging {
		t.Error("EnableLogging should be true")
	}
	if !cfg.EnableMetrics {
		t.Error("EnableMetrics should be true")
	}
	if cfg.EnableMetadata {
		t.Error("EnableMetadata should be false")
	}
	if cfg.EnableErrorHelp {
		t.Error("EnableErrorHelp should be false")
	}
}

func TestExtConfig_Defaults(t *testing.T) {
	sc := ServerConfig{}
	cfg := sc.ExtConfig()

	// Should use defaults when pointers are nil
	if cfg.EnableLogging {
		t.Error("EnableLogging should default to false")
	}
	if !cfg.EnableErrorHelp {
		t.Error("EnableErrorHelp should default to true")
	}
}

func TestDescriptionsMap(t *testing.T) {
	sc := ServerConfig{
		Toolkit: ToolkitConfig{
			Descriptions: map[string]string{
				"datahub_search":     "Custom search",
				"datahub_get_entity": "Custom entity",
			},
		},
	}

	m := sc.DescriptionsMap()
	if len(m) != 2 {
		t.Fatalf("DescriptionsMap() len = %d, want 2", len(m))
	}
	if m[tools.ToolSearch] != "Custom search" {
		t.Errorf("DescriptionsMap()[search] = %q", m[tools.ToolSearch])
	}
	if m[tools.ToolGetEntity] != "Custom entity" {
		t.Errorf("DescriptionsMap()[entity] = %q", m[tools.ToolGetEntity])
	}
}

func TestDescriptionsMap_Empty(t *testing.T) {
	sc := ServerConfig{}
	m := sc.DescriptionsMap()
	if m != nil {
		t.Error("DescriptionsMap() should return nil for empty descriptions")
	}
}

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()

	if cfg.DataHub.Timeout.Duration != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", cfg.DataHub.Timeout.Duration)
	}
	if cfg.Toolkit.DefaultLimit != 10 {
		t.Errorf("DefaultLimit = %d, want 10", cfg.Toolkit.DefaultLimit)
	}
	if cfg.Toolkit.MaxLimit != 100 {
		t.Errorf("MaxLimit = %d, want 100", cfg.Toolkit.MaxLimit)
	}
	if cfg.Toolkit.MaxLineageDepth != 5 {
		t.Errorf("MaxLineageDepth = %d, want 5", cfg.Toolkit.MaxLineageDepth)
	}
	if cfg.Extensions.Errors == nil || !*cfg.Extensions.Errors {
		t.Error("Errors should be enabled by default")
	}
}

func TestDuration_JSON_Roundtrip(t *testing.T) {
	d := Duration{Duration: 45 * time.Second}

	data, err := d.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	var d2 Duration
	if err := d2.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if d2.Duration != d.Duration {
		t.Errorf("roundtrip: got %v, want %v", d2.Duration, d.Duration)
	}
}

func TestDuration_YAML_Roundtrip(t *testing.T) {
	d := Duration{Duration: 2 * time.Minute}

	val, err := d.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error: %v", err)
	}

	s, ok := val.(string)
	if !ok {
		t.Fatalf("MarshalYAML() returned %T, want string", val)
	}
	if s != "2m0s" {
		t.Errorf("MarshalYAML() = %q, want %q", s, "2m0s")
	}
}

func TestDuration_UnmarshalJSON_Invalid(t *testing.T) {
	var d Duration

	// Not a string
	if err := d.UnmarshalJSON([]byte("123")); err == nil {
		t.Error("expected error for non-string")
	}

	// Invalid duration
	if err := d.UnmarshalJSON([]byte(`"not-a-duration"`)); err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestPartialConfig_Defaults(t *testing.T) {
	// Only set one field, rest should use defaults
	data := []byte(`{"toolkit": {"default_limit": 50}}`)
	cfg, err := FromBytes(data, "json")
	if err != nil {
		t.Fatalf("FromBytes() error: %v", err)
	}

	if cfg.Toolkit.DefaultLimit != 50 {
		t.Errorf("DefaultLimit = %d, want 50", cfg.Toolkit.DefaultLimit)
	}
	// Defaults should still be applied
	if cfg.Toolkit.MaxLimit != 100 {
		t.Errorf("MaxLimit = %d, want default 100", cfg.Toolkit.MaxLimit)
	}
	if cfg.DataHub.Timeout.Duration != 30*time.Second {
		t.Errorf("Timeout = %v, want default 30s", cfg.DataHub.Timeout.Duration)
	}
}
