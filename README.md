<div style="display: flex; flex-wrap: wrap; gap: 2px">

  <a href="https://badge.fury.io/go/github.com%2FMacawls%2Fumami-mcp-server">
    <img src="https://badge.fury.io/go/github.com%2Fmacawls%2Fumami-mcp-server.svg" alt="Go project version" />
  </a>

  <a href="https://pkg.go.dev/github.com/Macawls/umami-mcp-server">
    <img src="https://pkg.go.dev/badge/github.com/Macawls/umami-mcp-server.svg" alt="Go Reference" />
  </a>

  <a href="https://github.com/Macawls/umami-mcp-server/actions/workflows/test.yml">
    <img src="https://github.com/Macawls/umami-mcp-server/actions/workflows/test.yml/badge.svg" alt="Test" />
  </a>

  <a href="https://github.com/Macawls/umami-mcp-server/actions/workflows/release.yml">
    <img src="https://github.com/Macawls/umami-mcp-server/actions/workflows/release.yml/badge.svg" alt="Release" />
  </a>

  <a href="https://github.com/Macawls/umami-mcp-server/actions/workflows/pages.yml">
    <img src="https://github.com/Macawls/umami-mcp-server/actions/workflows/pages.yml/badge.svg" alt="Deploy to GitHub Pages" />
  </a>

</div>

# Umami MCP Server

Connect your Umami Analytics to any MCP client - Claude Desktop, VS Code, Cursor, Windsurf, Zed, and more.

<img src=".github/workflows/insights.PNG" height="500">

## Prompts

### Analytics & Traffic

- "Give me a comprehensive analytics report for my website over the last 30 days"
- "Which pages are getting the most traffic this month? Show me the top 10"
- "Analyze my website's traffic patterns - when do I get the most visitors?"

### User Insights

- "Where are my visitors coming from? Break it down by country and city"
- "What devices and browsers are my users using?"
- "Show me the user journey - what pages do visitors typically view in sequence?"

### Real-time Monitoring

- "How many people are on my website right now? What pages are they viewing?"
- "Is my website experiencing any issues? Check if traffic has dropped significantly"

### Content & Campaign Analysis

- "Which blog posts should I update? Show me articles with declining traffic"
- "How did my recent email campaign perform? Track visitors from the campaign UTM"
- "Compare traffic from different social media platforms"

## Quick Start

### Option 1: Download Binary

Get the latest release for your platform from [Releases](https://github.com/Macawls/umami-mcp-server/releases)

### Option 2: Docker

```bash
docker run -i --rm \
  -e UMAMI_URL="https://your-instance.com" \
  -e UMAMI_USERNAME="username" \
  -e UMAMI_PASSWORD="password" \
  ghcr.io/macawls/umami-mcp-server
```

### Option 3: Go Install

```bash
go install github.com/Macawls/umami-mcp-server@latest
# Or specific version
go install github.com/Macawls/umami-mcp-server@v1.0.3
```

Installs to `~/go/bin/umami-mcp-server` (or `$GOPATH/bin`)

## Configure Your MCP Client

### Claude Desktop

Add to your Claude Desktop config:

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`  
**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "umami": {
      "command": "~/go/bin/umami-mcp-server",
      "env": {
        "UMAMI_URL": "https://your-umami-instance.com",
        "UMAMI_USERNAME": "your-username",
        "UMAMI_PASSWORD": "your-password"
      }
    }
  }
}
```

<details>
<summary>Docker version</summary>

```json
{
  "mcpServers": {
    "umami": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "UMAMI_URL",
        "-e",
        "UMAMI_USERNAME",
        "-e",
        "UMAMI_PASSWORD",
        "ghcr.io/macawls/umami-mcp-server"
      ],
      "env": {
        "UMAMI_URL": "https://your-umami-instance.com",
        "UMAMI_USERNAME": "your-username",
        "UMAMI_PASSWORD": "your-password"
      }
    }
  }
}
```

</details>

<details>
<summary>Secure prompts</summary>

```json
{
  "mcpServers": {
    "umami": {
      "command": "~/go/bin/umami-mcp-server",
      "env": {
        "UMAMI_URL": "${input:umami_url}",
        "UMAMI_USERNAME": "${input:umami_username}",
        "UMAMI_PASSWORD": "${input:umami_password}"
      }
    }
  },
  "inputs": [
    {
      "type": "promptString",
      "id": "umami_url",
      "description": "Umami instance URL"
    },
    {
      "type": "promptString",
      "id": "umami_username",
      "description": "Umami username"
    },
    {
      "type": "promptString",
      "id": "umami_password",
      "description": "Umami password",
      "password": true
    }
  ]
}
```

</details>

Restart Claude Desktop to load the server.

### VS Code (GitHub Copilot)

Enable agent mode and add MCP servers to access Umami from Copilot.

**For workspace:** Create `.vscode/mcp.json`

```json
{
  "servers": {
    "umami": {
      "command": "~/go/bin/umami-mcp-server",
      "env": {
        "UMAMI_URL": "https://your-umami-instance.com",
        "UMAMI_USERNAME": "your-username",
        "UMAMI_PASSWORD": "your-password"
      }
    }
  }
}
```

<details>
<summary>With secure prompts</summary>

```json
{
  "inputs": [
    {
      "type": "promptString",
      "id": "umami_url",
      "description": "Umami instance URL"
    },
    {
      "type": "promptString",
      "id": "umami_username",
      "description": "Umami username"
    },
    {
      "type": "promptString",
      "id": "umami_password",
      "description": "Umami password",
      "password": true
    }
  ],
  "servers": {
    "umami": {
      "command": "~/go/bin/umami-mcp-server",
      "env": {
        "UMAMI_URL": "${input:umami_url}",
        "UMAMI_USERNAME": "${input:umami_username}",
        "UMAMI_PASSWORD": "${input:umami_password}"
      }
    }
  }
}
```

</details>

Access via: Chat view → Agent mode → Tools button

### Other MCP Clients

<details>
<summary>Cursor, Windsurf, Zed, Cline</summary>

**Cursor:** `Ctrl/Cmd + Shift + P` → "Cursor Settings" → MCP section

**Windsurf:** Settings → MCP Settings → Add MCP Server  
Config location: `%APPDATA%\windsurf\mcp_settings.json` (Windows)

**Zed:** Settings → `assistant.mcp_servers`

**Cline:** VS Code Settings → Extensions → Cline → MCP Servers

All use similar JSON format as above. Docker and secure prompts work the same way.

</details>

## Available Tools

- **get_websites** - List all your websites
- **get_stats** - Get visitor statistics
- **get_pageviews** - View page traffic over time
- **get_metrics** - See browsers, countries, devices, and more
- **get_active** - Current active visitors

## Alternative Configuration

Instead of environment variables, create a `config.yaml` file next to the binary:

```yaml
umami_url: https://your-umami-instance.com
username: your-username
password: your-password
```

Environment variables take priority over the config file.

## Build from Source

```bash
git clone https://github.com/Macawls/umami-mcp-server.git
cd umami-mcp-server
go build -o umami-mcp
```

## Troubleshooting

### Binary won't run

- **macOS**: Run `xattr -c umami-mcp-server` to remove quarantine
- **Linux**: Run `chmod +x umami-mcp-server` to make executable

### Connection errors

- Verify your Umami instance is accessible
- Check your credentials are correct

### Tools not showing up

- Check your MCP client logs for errors
- Verify the binary path is absolute
- Try running the binary directly to check for errors

## License

MIT
