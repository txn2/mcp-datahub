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
    DefaultLimit    int // Default search limit
    MaxLimit        int // Maximum search limit
    MaxLineageDepth int // Max lineage depth
}
```

### Example

```go
toolkit := tools.NewToolkit(datahubClient, tools.Config{
    DefaultLimit:    20,
    MaxLimit:        50,
    MaxLineageDepth: 3,
})
```

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
