# Umami MCP Server

Connect your Umami Analytics to any MCP client - Claude Desktop, VS Code, Zed, and more.

## Quick Start

### 1. Download

Get the latest release for your platform from [Releases](https://github.com/Macawls/umami-mcp-server/releases).

### 2. Configure Your MCP Client

<details>
<summary><strong>Claude Desktop</strong></summary>

Add to your Claude Desktop config:

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`  
**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "umami": {
      "command": "path/to/umami-mcp",
      "env": {
        "UMAMI_URL": "https://your-analytics.com",
        "UMAMI_USERNAME": "your-username",
        "UMAMI_PASSWORD": "your-password"
      }
    }
  }
}
```

Restart Claude Desktop to load the server.
</details>

<details>
<summary><strong>VS Code (Cline)</strong></summary>

Add to your VS Code settings (`Ctrl/Cmd + ,` → Extensions → Cline):

```json
{
  "cline.mcpServers": {
    "umami": {
      "command": "path/to/umami-mcp",
      "env": {
        "UMAMI_URL": "https://your-analytics.com",
        "UMAMI_USERNAME": "your-username", 
        "UMAMI_PASSWORD": "your-password"
      }
    }
  }
}
```

Or add to `.vscode/settings.json` in your workspace.
</details>

<details>
<summary><strong>Zed</strong></summary>

Add to your Zed settings:

```json
{
  "assistant": {
    "version": "2",
    "mcp_servers": {
      "umami": {
        "command": "path/to/umami-mcp",
        "env": {
          "UMAMI_URL": "https://your-analytics.com",
          "UMAMI_USERNAME": "your-username",
          "UMAMI_PASSWORD": "your-password"
        }
      }
    }
  }
}
```
</details>

<details>
<summary><strong>Other MCP Clients</strong></summary>

For any MCP-compatible client, you'll need:

- **Command**: Path to the umami-mcp binary
- **Environment Variables**:
  - `UMAMI_URL`: Your Umami instance URL
  - `UMAMI_USERNAME`: Your username
  - `UMAMI_PASSWORD`: Your password

Check your client's documentation for specific configuration format.
</details>

## Available Tools

- **get_websites** - List all your websites
- **get_stats** - Get visitor statistics
- **get_pageviews** - View page traffic over time
- **get_metrics** - See browsers, countries, devices, and more
- **get_active** - Current active visitors

## Example Prompts

- "Show me all my Umami websites"
- "Get stats for website [ID] from last week"
- "What browsers are visitors using?"
- "How many active visitors right now?"

## Alternative Configuration

Instead of environment variables, create a `config.yaml` file next to the binary:

```yaml
umami_url: https://your-analytics.com
username: your-username
password: your-password
```

Environment variables take priority over the config file.

## Troubleshooting

### Binary won't run
- **macOS**: Run `xattr -c umami-mcp` to remove quarantine
- **Linux**: Run `chmod +x umami-mcp` to make executable

### Connection errors
- Verify your Umami instance is accessible
- Check your credentials are correct
- Ensure the URL has no trailing slash

### Tools not showing up
- Check your MCP client logs for errors
- Verify the binary path is absolute
- Try running the binary directly to check for errors

## Build from Source

```bash
git clone https://github.com/Macawls/umami-mcp-server.git
cd umami-mcp-server
go build -o umami-mcp
```

## License

MIT