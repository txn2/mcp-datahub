// Package server provides the default MCP server setup for mcp-datahub.
package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/multiserver"
	"github.com/txn2/mcp-datahub/pkg/tools"
)

// Version is the MCP server version.
const Version = "0.1.0"

// Options configures the server.
type Options struct {
	// MultiServerConfig is the multi-server configuration.
	// If nil, will be loaded from environment via multiserver.FromEnv().
	MultiServerConfig *multiserver.Config

	// ToolkitConfig is the toolkit configuration.
	ToolkitConfig tools.Config
}

// DefaultOptions returns default server options.
// Note: MultiServerConfig is loaded from environment when nil.
func DefaultOptions() Options {
	return Options{
		MultiServerConfig: nil, // Loaded from env in New()
		ToolkitConfig:     tools.DefaultConfig(),
	}
}

// New creates a new MCP server with DataHub tools.
// Returns the MCP server and the connection manager for cleanup.
// The server starts even if unconfigured - tools will return helpful errors.
func New(opts Options) (*mcp.Server, *multiserver.Manager, error) {
	// Load multi-server config from environment if not provided
	var msCfg multiserver.Config
	if opts.MultiServerConfig != nil {
		msCfg = *opts.MultiServerConfig
	} else {
		var err error
		msCfg, err = multiserver.FromEnv()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load server configuration: %w", err)
		}
	}

	// Check configuration but don't fail - store error for tools to report
	var configErr error
	if err := msCfg.Primary.Validate(); err != nil {
		configErr = fmt.Errorf("datahub connection not configured: %w - please set DATAHUB_URL and DATAHUB_TOKEN", err)
	}

	// Create connection manager
	mgr := multiserver.NewManager(msCfg)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-datahub",
		Version: Version,
	}, nil)

	// Apply write-enabled from environment if not already set in config
	if !opts.ToolkitConfig.WriteEnabled {
		if v := os.Getenv("DATAHUB_WRITE_ENABLED"); strings.EqualFold(v, "true") || v == "1" {
			opts.ToolkitConfig.WriteEnabled = true
		}
	}

	// Build toolkit options
	var toolkitOpts []tools.ToolkitOption

	// If unconfigured, add middleware that returns helpful error for all tools
	if configErr != nil {
		toolkitOpts = append(toolkitOpts, tools.WithMiddleware(
			tools.BeforeFunc(func(_ context.Context, _ *tools.ToolContext) (context.Context, error) {
				return nil, configErr
			}),
		))
	}

	// Create toolkit with multi-server manager and register tools
	toolkit := tools.NewToolkitWithManager(mgr, opts.ToolkitConfig, toolkitOpts...)
	toolkit.RegisterAll(server)

	return server, mgr, nil
}
