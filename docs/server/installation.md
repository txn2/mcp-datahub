# Installation

## Claude Desktop (MCPB Bundle)

The easiest way to install for Claude Desktop - download and double-click:

1. Go to the [releases page](https://github.com/txn2/mcp-datahub/releases)
2. Download the `.mcpb` bundle for your platform:
   - **macOS Apple Silicon (M1/M2/M3/M4)**: `mcp-datahub_X.X.X_darwin_arm64.mcpb`
   - **macOS Intel**: `mcp-datahub_X.X.X_darwin_amd64.mcpb`
   - **Windows**: `mcp-datahub_X.X.X_windows_amd64.mcpb`
3. Double-click to install
4. Configure your DataHub URL and token when prompted

## Homebrew (macOS)

```bash
brew install txn2/tap/mcp-datahub
```

## Go Install

If you have Go installed:

```bash
go install github.com/txn2/mcp-datahub/cmd/mcp-datahub@latest
```

## Download Binary

Download pre-built binaries from the [releases page](https://github.com/txn2/mcp-datahub/releases).

Available platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Docker

```bash
docker pull ghcr.io/txn2/mcp-datahub:latest

docker run -e DATAHUB_URL=https://datahub.company.com \
           -e DATAHUB_TOKEN=your_token \
           ghcr.io/txn2/mcp-datahub:latest
```

## Claude Desktop Manual Configuration

If you installed via Homebrew or binary download, add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "datahub": {
      "command": "/opt/homebrew/bin/mcp-datahub",
      "env": {
        "DATAHUB_URL": "https://datahub.example.com",
        "DATAHUB_TOKEN": "your_token"
      }
    }
  }
}
```

### Multi-Server Configuration

To connect to multiple DataHub instances:

```json
{
  "mcpServers": {
    "datahub": {
      "command": "/opt/homebrew/bin/mcp-datahub",
      "env": {
        "DATAHUB_URL": "https://prod.datahub.example.com",
        "DATAHUB_TOKEN": "prod-token",
        "DATAHUB_CONNECTION_NAME": "prod",
        "DATAHUB_ADDITIONAL_SERVERS": "{\"staging\":{\"url\":\"https://staging.datahub.example.com\",\"token\":\"staging-token\"}}"
      }
    }
  }
}
```

## Claude Code CLI

```bash
claude mcp add datahub \
  -e DATAHUB_URL=https://datahub.example.com \
  -e DATAHUB_TOKEN=your-token \
  -- mcp-datahub
```

## Verify Installation

```bash
mcp-datahub --version
```

## Next Steps

- [Configuration Options](configuration.md)
- [Available Tools](tools.md)
