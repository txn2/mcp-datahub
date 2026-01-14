package tools

// Config configures the DataHub toolkit behavior.
type Config struct {
	// DefaultLimit is the default search result limit. Default: 10.
	DefaultLimit int

	// MaxLimit is the maximum allowed limit. Default: 100.
	MaxLimit int

	// MaxLineageDepth is the maximum lineage traversal depth. Default: 5.
	MaxLineageDepth int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DefaultLimit:    10,
		MaxLimit:        100,
		MaxLineageDepth: 5,
	}
}

// normalizeConfig applies default values to a Config.
func normalizeConfig(cfg Config) Config {
	if cfg.DefaultLimit <= 0 {
		cfg.DefaultLimit = 10
	}
	if cfg.MaxLimit <= 0 {
		cfg.MaxLimit = 100
	}
	if cfg.MaxLineageDepth <= 0 {
		cfg.MaxLineageDepth = 5
	}
	return cfg
}
