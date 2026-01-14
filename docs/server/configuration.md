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

## Example Configuration

```bash
# Required
export DATAHUB_URL=https://datahub.company.com
export DATAHUB_TOKEN=your_personal_access_token

# Optional tuning
export DATAHUB_TIMEOUT=60
export DATAHUB_DEFAULT_LIMIT=20
export DATAHUB_MAX_LIMIT=50
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
