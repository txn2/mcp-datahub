# Installation

## Homebrew (macOS)

The easiest way to install on macOS:

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

## Claude Desktop Configuration

Add to your `claude_desktop_config.json`:

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

## Verify Installation

```bash
mcp-datahub --version
```

## Next Steps

- [Configuration Options](configuration.md)
- [Available Tools](tools.md)
