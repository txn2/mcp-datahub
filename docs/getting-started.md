# Getting Started

Get mcp-datahub running in 5 minutes.

## Prerequisites

- A running DataHub instance
- A DataHub personal access token
- One of: Claude Desktop, Claude Code, Cursor, or another MCP client

## Step 1: Install

Choose your preferred installation method:

**Claude Desktop (Easiest)**

Download the `.mcpb` bundle from the [releases page](https://github.com/txn2/mcp-datahub/releases) and double-click to install.

**Homebrew (macOS)**

```bash
brew install txn2/tap/mcp-datahub
```

**Go Install**

```bash
go install github.com/txn2/mcp-datahub/cmd/mcp-datahub@latest
```

## Step 2: Get Your DataHub Token

1. Log into your DataHub instance
2. Navigate to Settings (gear icon in top right)
3. Select "Access Tokens"
4. Click "Generate New Token"
5. Give it a descriptive name like "mcp-datahub"
6. Copy the token value

## Step 3: Configure

If you used the MCPB bundle, you will be prompted for configuration during installation.

For other installation methods, set environment variables:

```bash
export DATAHUB_URL=https://your-datahub.company.com
export DATAHUB_TOKEN=your_token_here
```

Or configure Claude Desktop manually by editing `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "datahub": {
      "command": "mcp-datahub",
      "env": {
        "DATAHUB_URL": "https://your-datahub.company.com",
        "DATAHUB_TOKEN": "your_token_here"
      }
    }
  }
}
```

## Step 4: Verify

Restart your MCP client (Claude Desktop, etc.) and try a simple query:

> "Search DataHub for customer datasets"

You should see search results from your DataHub instance.

## What You Can Do Now

With mcp-datahub connected, you can ask your AI assistant to:

- **Search**: "Find all datasets related to orders"
- **Explore schemas**: "What fields are in the customers table?"
- **Trace lineage**: "What downstream dashboards use this dataset?"
- **Understand context**: "What does the PII glossary term mean?"
- **Browse domains**: "List all data domains in our catalog"

## Next Steps

- [Available Tools Reference](server/tools.md): Full documentation of all tools
- [Configuration Options](server/configuration.md): Customize timeouts, limits, and more
- [Multi-Server Setup](server/configuration.md#multi-server-configuration): Connect to multiple DataHub instances
- [Library Usage](library/index.md): Build custom MCP servers with authentication

## Common Issues

**"unauthorized" error**: Your token may be invalid or expired. Generate a new one from DataHub Settings.

**"connection refused" error**: Check that your DATAHUB_URL is correct and the server is accessible.

**No results found**: Verify that data exists in your DataHub instance by checking the DataHub UI.

See the [Troubleshooting Guide](support/troubleshooting.md) for more help.
