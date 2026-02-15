package extensions

import (
	"io"
	"os"
	"strings"

	"github.com/txn2/mcp-datahub/pkg/tools"
)

// Config controls which extensions are enabled.
type Config struct {
	// EnableLogging enables structured logging of tool calls.
	EnableLogging bool

	// EnableMetrics enables metrics collection.
	EnableMetrics bool

	// EnableMetadata enables metadata enrichment on tool results.
	EnableMetadata bool

	// EnableErrorHelp enables error hint enrichment.
	EnableErrorHelp bool

	// LogOutput is the writer for logging output. Defaults to os.Stderr.
	LogOutput io.Writer
}

// DefaultConfig returns a Config with error hints enabled by default.
func DefaultConfig() Config {
	return Config{
		EnableErrorHelp: true,
	}
}

// FromEnv loads extension configuration from environment variables.
func FromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("MCP_DATAHUB_EXT_LOGGING"); v != "" {
		cfg.EnableLogging = parseBool(v)
	}
	if v := os.Getenv("MCP_DATAHUB_EXT_METRICS"); v != "" {
		cfg.EnableMetrics = parseBool(v)
	}
	if v := os.Getenv("MCP_DATAHUB_EXT_METADATA"); v != "" {
		cfg.EnableMetadata = parseBool(v)
	}
	if v := os.Getenv("MCP_DATAHUB_EXT_ERRORS"); v != "" {
		cfg.EnableErrorHelp = parseBool(v)
	}

	return cfg
}

// BuildToolkitOptions converts extension config into toolkit options.
func BuildToolkitOptions(cfg Config) []tools.ToolkitOption {
	var opts []tools.ToolkitOption

	if cfg.EnableLogging {
		output := cfg.LogOutput
		if output == nil {
			output = os.Stderr
		}
		opts = append(opts, tools.WithMiddleware(NewLoggingMiddleware(output)))
	}

	if cfg.EnableMetrics {
		collector := NewInMemoryCollector()
		opts = append(opts, tools.WithMiddleware(NewMetricsMiddleware(collector)))
	}

	if cfg.EnableMetadata {
		opts = append(opts, tools.WithMiddleware(NewMetadataMiddleware()))
	}

	if cfg.EnableErrorHelp {
		opts = append(opts, tools.WithMiddleware(NewErrorHintMiddleware()))
	}

	return opts
}

// parseBool returns true for "true", "1", "yes" (case-insensitive).
func parseBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
