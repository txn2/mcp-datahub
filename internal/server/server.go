// Package server provides the default MCP server setup for mcp-datahub.
package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/tools"
)

// Version is the MCP server version.
const Version = "0.1.0"

// Options configures the server.
type Options struct {
	// ClientConfig is the DataHub client configuration.
	// If nil, will be loaded from environment.
	ClientConfig *client.Config

	// ToolkitConfig is the toolkit configuration.
	ToolkitConfig tools.Config
}

// DefaultOptions returns default server options.
func DefaultOptions() Options {
	return Options{
		ClientConfig:  nil, // Loaded from env in New()
		ToolkitConfig: tools.DefaultConfig(),
	}
}

// New creates a new MCP server with DataHub tools.
// Returns the MCP server and the client for cleanup.
// The server starts even if unconfigured - tools will return helpful errors.
func New(opts Options) (*mcp.Server, *client.Client, error) {
	// Load client config from environment if not provided
	var clientCfg client.Config
	if opts.ClientConfig != nil {
		clientCfg = *opts.ClientConfig
	} else {
		var err error
		clientCfg, err = client.FromEnv()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Check configuration but don't fail - store error for tools to report
	var configErr error
	if err := clientCfg.Validate(); err != nil {
		configErr = fmt.Errorf("datahub connection not configured: %w - please set DATAHUB_URL and DATAHUB_TOKEN", err)
	}

	// Create DataHub client
	var datahubClient *client.Client
	if configErr == nil {
		var err error
		datahubClient, err = client.New(clientCfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create client: %w", err)
		}
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-datahub",
		Version: Version,
	}, nil)

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

	// Create toolkit and register tools
	// Use a wrapper client that handles nil case
	var toolClient tools.DataHubClient
	if datahubClient != nil {
		toolClient = datahubClient
	}

	toolkit := tools.NewToolkit(toolClient, opts.ToolkitConfig, toolkitOpts...)
	toolkit.RegisterAll(server)

	return server, datahubClient, nil
}
