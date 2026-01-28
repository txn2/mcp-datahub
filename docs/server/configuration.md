# Configuration

All configuration is done through environment variables.

## Required Variables

| Variable | Description |
|----------|-------------|
| `DATAHUB_URL` | DataHub GMS URL (e.g., `https://datahub.company.com`) |
| `DATAHUB_TOKEN` | Personal access token from DataHub |

## Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATAHUB_TIMEOUT` | Request timeout in seconds | `30` |
| `DATAHUB_RETRY_MAX` | Maximum retry attempts | `3` |
| `DATAHUB_DEFAULT_LIMIT` | Default search result limit | `10` |
| `DATAHUB_MAX_LIMIT` | Maximum allowed limit | `100` |
| `DATAHUB_MAX_LINEAGE_DEPTH` | Maximum lineage traversal depth | `5` |
| `DATAHUB_CONNECTION_NAME` | Display name for primary connection | `datahub` |
| `DATAHUB_ADDITIONAL_SERVERS` | JSON map of additional servers | (empty) |
| `DATAHUB_DEBUG` | Enable debug logging (`1` or `true`) | `false` |

## Example Configuration

```bash
# Required
export DATAHUB_URL=https://datahub.company.com
export DATAHUB_TOKEN=your_personal_access_token

# Optional tuning
export DATAHUB_TIMEOUT=60
export DATAHUB_DEFAULT_LIMIT=20
export DATAHUB_MAX_LIMIT=50

# Debug logging (for troubleshooting)
export DATAHUB_DEBUG=1
```

## Multi-Server Configuration

Connect to multiple DataHub instances simultaneously. Useful for:

- Production and staging environments
- Multi-tenant deployments
- Cross-environment metadata comparison

### Setting Up Multiple Servers

```bash
# Primary server configuration
export DATAHUB_URL=https://prod.datahub.example.com/api/graphql
export DATAHUB_TOKEN=prod-token
export DATAHUB_CONNECTION_NAME=prod  # Optional: customize display name

# Additional servers as JSON
export DATAHUB_ADDITIONAL_SERVERS='{
  "staging": {
    "url": "https://staging.datahub.example.com/api/graphql",
    "token": "staging-token"
  },
  "dev": {
    "url": "https://dev.datahub.example.com/api/graphql"
  }
}'
```

### Additional Server Options

Each additional server can override these settings (inherits from primary if not specified):

| Field | Description |
|-------|-------------|
| `url` | DataHub GMS URL (required) |
| `token` | Access token (inherits from primary) |
| `timeout` | Request timeout in seconds |
| `retry_max` | Maximum retry attempts |
| `default_limit` | Default search limit |
| `max_limit` | Maximum allowed limit |
| `max_lineage_depth` | Maximum lineage depth |

### Using Multiple Servers

1. Use `datahub_list_connections` to see available connections
2. Pass the `connection` parameter to any tool to target a specific server
3. If `connection` is omitted, the default (primary) server is used

```
# Example: Search staging server
datahub_search query="customers" connection="staging"
```

## Getting a DataHub Token

1. Log into DataHub
2. Go to Settings > Access Tokens
3. Generate a new token with appropriate permissions
4. Copy the token value

## Security Considerations

- Never commit tokens to version control
- Use environment variables or secret management
- Tokens should have minimal required permissions
- Rotate tokens periodically
