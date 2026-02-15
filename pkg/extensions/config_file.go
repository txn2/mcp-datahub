package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/tools"
)

// ServerConfig is the top-level configuration loaded from a file.
type ServerConfig struct {
	DataHub    DataHubConfig `json:"datahub" yaml:"datahub"`
	Toolkit    ToolkitConfig `json:"toolkit" yaml:"toolkit"`
	Extensions ExtFileConfig `json:"extensions" yaml:"extensions"`
}

// DataHubConfig configures the DataHub connection.
type DataHubConfig struct {
	URL            string   `json:"url" yaml:"url"`
	Token          string   `json:"token" yaml:"token"`
	Timeout        Duration `json:"timeout" yaml:"timeout"`
	ConnectionName string   `json:"connection_name" yaml:"connection_name"`
	WriteEnabled   *bool    `json:"write_enabled" yaml:"write_enabled"`
}

// ToolkitConfig configures toolkit behavior.
type ToolkitConfig struct {
	DefaultLimit    int               `json:"default_limit" yaml:"default_limit"`
	MaxLimit        int               `json:"max_limit" yaml:"max_limit"`
	MaxLineageDepth int               `json:"max_lineage_depth" yaml:"max_lineage_depth"`
	Descriptions    map[string]string `json:"descriptions" yaml:"descriptions"`
}

// ExtFileConfig configures extensions via file.
// Pointer bools allow distinguishing "not set" from "set to false".
type ExtFileConfig struct {
	Logging  *bool `json:"logging" yaml:"logging"`
	Metrics  *bool `json:"metrics" yaml:"metrics"`
	Metadata *bool `json:"metadata" yaml:"metadata"`
	Errors   *bool `json:"errors" yaml:"errors"`
}

// Duration wraps time.Duration for JSON/YAML unmarshalling.
// Accepts Go duration strings like "30s", "5m", "1h".
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dur
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dur
	return nil
}

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// MarshalYAML implements yaml.Marshaler.
func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// DefaultServerConfig returns a ServerConfig with sensible defaults.
func DefaultServerConfig() ServerConfig {
	errorsEnabled := true
	return ServerConfig{
		DataHub: DataHubConfig{
			Timeout: Duration{Duration: 30 * time.Second},
		},
		Toolkit: ToolkitConfig{
			DefaultLimit:    10,
			MaxLimit:        100,
			MaxLineageDepth: 5,
		},
		Extensions: ExtFileConfig{
			Errors: &errorsEnabled,
		},
	}
}

// FromFile loads a ServerConfig from a YAML or JSON file.
// The format is determined by the file extension (.yaml, .yml, .json).
func FromFile(path string) (ServerConfig, error) {
	data, err := os.ReadFile(path) //#nosec G304 -- path is user-provided config file
	if err != nil {
		return ServerConfig{}, fmt.Errorf("reading config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var format string
	switch ext {
	case ".yaml", ".yml":
		format = "yaml"
	case ".json":
		format = "json"
	default:
		return ServerConfig{}, fmt.Errorf("unsupported config file extension: %s", ext)
	}

	return FromBytes(data, format)
}

// FromBytes parses a ServerConfig from raw bytes in the given format ("json" or "yaml").
func FromBytes(data []byte, format string) (ServerConfig, error) {
	cfg := DefaultServerConfig()

	switch format {
	case "json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return ServerConfig{}, fmt.Errorf("parsing JSON config: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return ServerConfig{}, fmt.Errorf("parsing YAML config: %w", err)
		}
	default:
		return ServerConfig{}, fmt.Errorf("unsupported format: %s", format)
	}

	return cfg, nil
}

// LoadConfig loads a ServerConfig from a file and applies environment variable overrides.
// Environment variables take precedence over file values for sensitive fields.
func LoadConfig(path string) (ServerConfig, error) {
	cfg, err := FromFile(path)
	if err != nil {
		return cfg, err
	}

	// Environment overrides for connection settings
	if v := os.Getenv("DATAHUB_URL"); v != "" {
		cfg.DataHub.URL = v
	}
	if v := os.Getenv("DATAHUB_TOKEN"); v != "" {
		cfg.DataHub.Token = v
	}
	if v := os.Getenv("DATAHUB_TIMEOUT"); v != "" {
		dur, parseErr := time.ParseDuration(v)
		if parseErr != nil {
			return cfg, fmt.Errorf("invalid DATAHUB_TIMEOUT: %w", parseErr)
		}
		cfg.DataHub.Timeout = Duration{Duration: dur}
	}
	if v := os.Getenv("DATAHUB_CONNECTION_NAME"); v != "" {
		cfg.DataHub.ConnectionName = v
	}
	if v := os.Getenv("DATAHUB_WRITE_ENABLED"); v != "" {
		b := parseBool(v)
		cfg.DataHub.WriteEnabled = &b
	}

	// Expand environment variables in token (for $VAR or ${VAR} patterns)
	cfg.DataHub.Token = os.ExpandEnv(cfg.DataHub.Token)

	return cfg, nil
}

// ClientConfig converts the file config to a client.Config.
func (sc *ServerConfig) ClientConfig() client.Config {
	cfg := client.DefaultConfig()
	if sc.DataHub.URL != "" {
		cfg.URL = sc.DataHub.URL
	}
	if sc.DataHub.Token != "" {
		cfg.Token = sc.DataHub.Token
	}
	if sc.DataHub.Timeout.Duration > 0 {
		cfg.Timeout = sc.DataHub.Timeout.Duration
	}
	if sc.Toolkit.DefaultLimit > 0 {
		cfg.DefaultLimit = sc.Toolkit.DefaultLimit
	}
	if sc.Toolkit.MaxLimit > 0 {
		cfg.MaxLimit = sc.Toolkit.MaxLimit
	}
	if sc.Toolkit.MaxLineageDepth > 0 {
		cfg.MaxLineageDepth = sc.Toolkit.MaxLineageDepth
	}
	return cfg
}

// ToolsConfig converts the file config to a tools.Config.
func (sc *ServerConfig) ToolsConfig() tools.Config {
	cfg := tools.DefaultConfig()
	if sc.Toolkit.DefaultLimit > 0 {
		cfg.DefaultLimit = sc.Toolkit.DefaultLimit
	}
	if sc.Toolkit.MaxLimit > 0 {
		cfg.MaxLimit = sc.Toolkit.MaxLimit
	}
	if sc.Toolkit.MaxLineageDepth > 0 {
		cfg.MaxLineageDepth = sc.Toolkit.MaxLineageDepth
	}
	if sc.DataHub.WriteEnabled != nil {
		cfg.WriteEnabled = *sc.DataHub.WriteEnabled
	}
	return cfg
}

// ExtConfig converts the file config to an extensions.Config.
func (sc *ServerConfig) ExtConfig() Config {
	cfg := DefaultConfig()
	if sc.Extensions.Logging != nil {
		cfg.EnableLogging = *sc.Extensions.Logging
	}
	if sc.Extensions.Metrics != nil {
		cfg.EnableMetrics = *sc.Extensions.Metrics
	}
	if sc.Extensions.Metadata != nil {
		cfg.EnableMetadata = *sc.Extensions.Metadata
	}
	if sc.Extensions.Errors != nil {
		cfg.EnableErrorHelp = *sc.Extensions.Errors
	}
	return cfg
}

// DescriptionsMap converts string-keyed descriptions to tools.ToolName-keyed map.
func (sc *ServerConfig) DescriptionsMap() map[tools.ToolName]string {
	if len(sc.Toolkit.Descriptions) == 0 {
		return nil
	}
	m := make(map[tools.ToolName]string, len(sc.Toolkit.Descriptions))
	for k, v := range sc.Toolkit.Descriptions {
		m[tools.ToolName(k)] = v
	}
	return m
}
