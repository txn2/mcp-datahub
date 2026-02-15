# Configuration Reference

Complete configuration reference for mcp-datahub.

## Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `DATAHUB_URL` | DataHub GMS URL | `https://datahub.company.com` |
| `DATAHUB_TOKEN` | Personal access token | `eyJhbGciOiJIUzI1NiIs...` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `DATAHUB_TIMEOUT` | HTTP request timeout (seconds) | `30` |
| `DATAHUB_RETRY_MAX` | Maximum retry attempts for failed requests | `3` |
| `DATAHUB_DEFAULT_LIMIT` | Default search result limit | `10` |
| `DATAHUB_MAX_LIMIT` | Maximum allowed search limit | `100` |
| `DATAHUB_MAX_LINEAGE_DEPTH` | Maximum lineage traversal depth | `5` |
| `DATAHUB_CONNECTION_NAME` | Display name for primary connection | `datahub` |
| `DATAHUB_ADDITIONAL_SERVERS` | JSON map of additional servers | (empty) |
| `DATAHUB_WRITE_ENABLED` | Enable write operations (`true` or `1`) | `false` |
| `DATAHUB_DEBUG` | Enable debug logging (`1` or `true`) | `false` |

### Extensions

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_DATAHUB_EXT_LOGGING` | Enable structured logging of tool calls | `false` |
| `MCP_DATAHUB_EXT_METRICS` | Enable metrics collection | `false` |
| `MCP_DATAHUB_EXT_METADATA` | Enable metadata enrichment on results | `false` |
| `MCP_DATAHUB_EXT_ERRORS` | Enable error hint enrichment | `true` |

## Client Configuration

When using as a library, configure via the `Config` struct:

```go
type Config struct {
    URL             string        // DataHub GMS URL (required)
    Token           string        // API token (required)
    Timeout         time.Duration // Request timeout
    RetryMax        int           // Max retries
    DefaultLimit    int           // Default search limit
    MaxLimit        int           // Maximum search limit
    MaxLineageDepth int           // Max lineage depth
    Debug           bool          // Enable debug logging
    Logger          Logger        // Custom logger (nil = auto-select)
}
```

### DefaultConfig

Returns default configuration values:

```go
func DefaultConfig() Config {
    return Config{
        Timeout:         30 * time.Second,
        RetryMax:        3,
        DefaultLimit:    10,
        MaxLimit:        100,
        MaxLineageDepth: 5,
        Debug:           false,
        Logger:          nil, // Uses NopLogger; StdLogger when Debug=true
    }
}
```

### FromEnv

Loads configuration from environment variables:

```go
cfg, err := client.FromEnv()
if err != nil {
    log.Fatal(err)
}
```

## Toolkit Configuration

```go
type Config struct {
    DefaultLimit    int           // Default search limit
    MaxLimit        int           // Maximum search limit
    MaxLineageDepth int           // Max lineage depth
    WriteEnabled    bool          // Enable write operations
    Debug           bool          // Enable debug logging
    Logger          client.Logger // Custom logger (nil = auto-select)
}
```

### Example

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{
    DefaultLimit:    20,
    MaxLimit:        50,
    MaxLineageDepth: 3,
    WriteEnabled:    true,
})
```

## Debug Logging

Enable debug logging to troubleshoot issues with DataHub connectivity, GraphQL queries, and tool execution.

### Via Environment Variable

```bash
export DATAHUB_DEBUG=1
./mcp-datahub
```

### Programmatic Configuration

```go
// Auto-create StdLogger when Debug=true
cfg := client.Config{
    URL:   "https://datahub.example.com",
    Token: "token",
    Debug: true,
}

// Or provide a custom logger
cfg := client.Config{
    URL:    "https://datahub.example.com",
    Token:  "token",
    Logger: myCustomLogger,
}
```

### Logger Interface

The `Logger` interface is compatible with `slog.Logger` patterns:

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

Built-in implementations:
- `NopLogger` - Discards all output (default when debug disabled)
- `StdLogger` - Writes to stderr with structured key-value format

### Log Output

When debug logging is enabled, you'll see:

```
[datahub] DEBUG: executing GraphQL query [operation=GetEntity endpoint=https://... request_size=256]
[datahub] DEBUG: received response [status=200 response_size=1024]
[datahub] DEBUG: request completed [operation=GetEntity duration_ms=150 attempts=1]
```

## Description Overrides

Customize tool descriptions to match your deployment:

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{},
    tools.WithDescriptions(map[tools.ToolName]string{
        tools.ToolSearch: "Search our internal data catalog",
        tools.ToolGetEntity: "Get metadata for a dataset in our catalog",
    }),
)
```

Description priority (highest to lowest):

1. Per-registration override via `WithDescription()`
2. Toolkit-level override via `WithDescriptions()`
3. Built-in default description

## Extensions Configuration

The `extensions` package provides built-in middleware and config file support.

### Loading from Environment

```go
import "github.com/txn2/mcp-datahub/pkg/extensions"

cfg := extensions.FromEnv()
opts := extensions.BuildToolkitOptions(cfg)
toolkit := tools.NewToolkit(datahubClient, toolsCfg, opts...)
```

### Extensions Config Struct

```go
type Config struct {
    EnableLogging   bool      // Structured logging of tool calls
    EnableMetrics   bool      // Metrics collection
    EnableMetadata  bool      // Metadata enrichment on results
    EnableErrorHelp bool      // Error hint enrichment (default: true)
    LogOutput       io.Writer // Custom log output (default: os.Stderr)
}
```

### Config File Support

Load configuration from YAML or JSON files:

```go
serverCfg, err := extensions.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

clientCfg := serverCfg.ClientConfig()     // -> client.Config
toolsCfg := serverCfg.ToolsConfig()       // -> tools.Config
extCfg := serverCfg.ExtConfig()           // -> extensions.Config
descs := serverCfg.DescriptionsMap()      // -> map[tools.ToolName]string
```

#### YAML Config File Format

```yaml
datahub:
  url: https://datahub.example.com
  token: "${DATAHUB_TOKEN}"
  timeout: "30s"
  connection_name: prod
  write_enabled: true

toolkit:
  default_limit: 20
  max_limit: 50
  max_lineage_depth: 3
  descriptions:
    datahub_search: "Search our internal data catalog"

extensions:
  logging: true
  metrics: false
  metadata: false
  errors: true
```

Environment variables override file values for sensitive fields (`DATAHUB_URL`, `DATAHUB_TOKEN`, `DATAHUB_TIMEOUT`, `DATAHUB_CONNECTION_NAME`, `DATAHUB_WRITE_ENABLED`). Token values support `$VAR` / `${VAR}` expansion.

## Validation

Configuration is validated on client creation:

- `URL` must be non-empty
- `Token` must be non-empty
- Numeric limits must be positive

Invalid configuration returns an error:

```go
client, err := client.New(cfg)
if err != nil {
    // Handle validation error
}
```
