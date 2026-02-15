// Package extensions provides optional middleware and configuration for mcp-datahub.
//
// Extensions add cross-cutting concerns like logging, metrics, error hints,
// and metadata enrichment to DataHub MCP tools. All extensions are opt-in
// and configured via environment variables or a config file.
//
// # Environment Variables
//
// Extensions can be enabled via environment variables:
//
//   - MCP_DATAHUB_EXT_LOGGING: Enable structured logging of tool calls ("true"/"1")
//   - MCP_DATAHUB_EXT_METRICS: Enable metrics collection ("true"/"1")
//   - MCP_DATAHUB_EXT_METADATA: Enable metadata enrichment on results ("true"/"1")
//   - MCP_DATAHUB_EXT_ERRORS: Enable error hint enrichment ("true"/"1", default: "true")
//
// # Config File
//
// For file-based configuration, use [FromFile] or [LoadConfig] to load
// YAML or JSON config files. See [ServerConfig] for the full schema.
//
// # Usage
//
//	cfg := extensions.FromEnv()
//	opts := extensions.BuildToolkitOptions(cfg)
//	toolkit := tools.NewToolkit(client, toolsCfg, opts...)
package extensions
