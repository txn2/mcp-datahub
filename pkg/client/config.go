package client

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// APIVersionV1 selects the legacy Rest.li API endpoints (DataHub <= 1.3.x).
const APIVersionV1 = "v1"

// APIVersionV3 selects the OpenAPI v3 endpoints (DataHub >= 1.4.0).
const APIVersionV3 = "v3"

// Config holds the DataHub client configuration.
type Config struct {
	// URL is the DataHub GMS URL (required).
	URL string

	// Token is the personal access token (required).
	Token string

	// Timeout is the request timeout. Default: 30s.
	Timeout time.Duration

	// RetryMax is the maximum retry attempts. Default: 3.
	RetryMax int

	// DefaultLimit is the default search result limit. Default: 10.
	DefaultLimit int

	// MaxLimit is the maximum allowed limit. Default: 100.
	MaxLimit int

	// MaxLineageDepth is the maximum lineage traversal depth. Default: 5.
	MaxLineageDepth int

	// APIVersion selects the REST API version: "v1" (legacy Rest.li, default)
	// or "v3" (OpenAPI v3, DataHub >= 1.4.0).
	APIVersion string

	// Debug enables debug logging. Default: false.
	Debug bool

	// Logger is the logger for debug output. If nil, a NopLogger is used.
	// When Debug is true and Logger is nil, a StdLogger is created automatically.
	Logger Logger
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		Timeout:         30 * time.Second,
		RetryMax:        3,
		DefaultLimit:    10,
		MaxLimit:        100,
		MaxLineageDepth: 5,
	}
}

// FromEnv creates a Config from environment variables.
func FromEnv() (Config, error) {
	cfg := DefaultConfig()

	cfg.URL = os.Getenv("DATAHUB_URL")
	cfg.Token = os.Getenv("DATAHUB_TOKEN")

	if timeout := os.Getenv("DATAHUB_TIMEOUT"); timeout != "" {
		secs, err := strconv.Atoi(timeout)
		if err != nil {
			return cfg, fmt.Errorf("invalid DATAHUB_TIMEOUT: %w", err)
		}
		cfg.Timeout = time.Duration(secs) * time.Second
	}

	if retryMax := os.Getenv("DATAHUB_RETRY_MAX"); retryMax != "" {
		val, err := strconv.Atoi(retryMax)
		if err != nil {
			return cfg, fmt.Errorf("invalid DATAHUB_RETRY_MAX: %w", err)
		}
		cfg.RetryMax = val
	}

	if defaultLimit := os.Getenv("DATAHUB_DEFAULT_LIMIT"); defaultLimit != "" {
		val, err := strconv.Atoi(defaultLimit)
		if err != nil {
			return cfg, fmt.Errorf("invalid DATAHUB_DEFAULT_LIMIT: %w", err)
		}
		cfg.DefaultLimit = val
	}

	if maxLimit := os.Getenv("DATAHUB_MAX_LIMIT"); maxLimit != "" {
		val, err := strconv.Atoi(maxLimit)
		if err != nil {
			return cfg, fmt.Errorf("invalid DATAHUB_MAX_LIMIT: %w", err)
		}
		cfg.MaxLimit = val
	}

	if maxDepth := os.Getenv("DATAHUB_MAX_LINEAGE_DEPTH"); maxDepth != "" {
		val, err := strconv.Atoi(maxDepth)
		if err != nil {
			return cfg, fmt.Errorf("invalid DATAHUB_MAX_LINEAGE_DEPTH: %w", err)
		}
		cfg.MaxLineageDepth = val
	}

	if apiVersion := os.Getenv("DATAHUB_API_VERSION"); apiVersion != "" {
		cfg.APIVersion = apiVersion
	}

	if debug := os.Getenv("DATAHUB_DEBUG"); debug != "" {
		cfg.Debug = debug == "1" || debug == "true"
	}

	return cfg, nil
}

// isV3 returns true when the config selects the OpenAPI v3 endpoints.
func (c Config) isV3() bool {
	return c.APIVersion == APIVersionV3
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("DATAHUB_URL is required")
	}
	if c.Token == "" {
		return fmt.Errorf("DATAHUB_TOKEN is required")
	}
	return nil
}
