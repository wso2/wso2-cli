# WSO2 Integration Platform CLI

A CLI and MCP Server for [WSO2 Integration Platform](https://wso2.com/integration-platform/) — create and manage Automations, APIs, AI Agents, MCP Servers, Event Integrations, and File Integrations.

---

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [MCP Server](#mcp-server)
  - [Option 1 — Claude Connectors (.mcpb)](#option-1--claude-connectors-mcpb)
  - [Option 2 — Claude Code & MCP agents (npm)](#option-2--claude-code--mcp-agents-npm)
  - [Option 3 — Direct binary](#option-3--direct-binary)
  - [Authentication](#authentication)
- [Development](#development)
- [Release](#release)

---

## Installation

### macOS / Linux / WSL

```bash
curl -sSfL https://raw.githubusercontent.com/wso2/integration-platform-tools/main/scripts/install.sh | sh
```

### Windows (PowerShell)

```powershell
iwr https://raw.githubusercontent.com/wso2/integration-platform-tools/main/scripts/install.ps1 -useb | iex
```

### npm

```bash
npm install -g @pcnfernando-wso2/integration-platform-mcp
```

### Manual download

Download the binary for your platform from the [GitHub releases page](https://github.com/wso2/integration-platform-tools/releases).

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `wso2-integration-platform-VERSION-darwin-arm64.zip` |
| macOS (Intel) | `wso2-integration-platform-VERSION-darwin-amd64.zip` |
| Linux (x64) | `wso2-integration-platform-VERSION-linux-amd64.tar.gz` |
| Linux (ARM64) | `wso2-integration-platform-VERSION-linux-arm64.tar.gz` |
| Windows (x64) | `wso2-integration-platform-VERSION-windows-amd64.zip` |

---

## Getting Started

```bash
# 1. Sign in
wso2-integration-platform login

# 2. List your projects
wso2-integration-platform list projects

# 3. List your integrations
wso2-integration-platform list integrations

# 4. Describe a resource
wso2-integration-platform describe <project|component>
```

---

## MCP Server

The CLI includes a built-in MCP (Model Context Protocol) server that exposes all platform operations as tools for AI assistants. Three distribution options are supported:

### Option 1 — Claude Connectors (.mcpb)

Install directly into Claude via the native connector bundle format. No separate CLI install needed.

1. Download `wso2-integration-platform-VERSION-OS-ARCH.mcpb` from the [releases page](https://github.com/wso2/integration-platform-tools/releases).
2. Open Claude → **Settings → Connectors → Install from file**.
3. Select the `.mcpb` file.
4. Claude will prompt you to log in the first time a tool is used.

### Option 2 — Claude Code & MCP agents (npm)

Best for Claude Code and any agent that supports MCP stdio transport. The npm package downloads the correct binary for your platform automatically.

**Install and configure in one step:**

```bash
# Install the npm package
npm install -g @pcnfernando-wso2/integration-platform-mcp

# Log in (stored credentials are reused across sessions)
wso2-integration-platform login

# Add to Claude Code
claude mcp add wso2-integration-platform -- wso2-integration-platform start-mcp-server
```

**Or configure your agent manually.** The server speaks MCP over stdio, so any
MCP-capable client can run it. Note that clients disagree on the top-level key —
VS Code uses `servers`, most others use `mcpServers`. Using the wrong one is the
most common reason a client shows no tools.

<details open>
<summary><b>VS Code (GitHub Copilot / Copilot Chat)</b></summary>

Create `.vscode/mcp.json` in your workspace (or add to your user settings).
VS Code uses `servers` and requires `type`:

```json
{
  "servers": {
    "wso2-integration-platform": {
      "type": "stdio",
      "command": "wso2-integration-platform",
      "args": ["start-mcp-server"]
    }
  }
}
```

Then open the Copilot Chat view and switch to **Agent** mode — MCP tools are only
available there, not in plain chat.
</details>

<details>
<summary><b>Claude Desktop</b></summary>

Add to `claude_desktop_config.json`
(macOS: `~/Library/Application Support/Claude/`, Windows: `%APPDATA%\Claude\`):

```json
{
  "mcpServers": {
    "wso2-integration-platform": {
      "command": "wso2-integration-platform",
      "args": ["start-mcp-server"]
    }
  }
}
```

Restart Claude Desktop afterwards — it only picks up new servers on start.
For a no-config alternative, use the `.mcpb` connector in Option 1 instead.
</details>

<details>
<summary><b>Cursor / Windsurf / other <code>mcpServers</code> clients</b></summary>

Cursor reads `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` (per project):

```json
{
  "mcpServers": {
    "wso2-integration-platform": {
      "command": "wso2-integration-platform",
      "args": ["start-mcp-server"]
    }
  }
}
```

The same block works for any client using the `mcpServers` convention; only the
file location differs.
</details>

<details>
<summary><b>Without a global install (<code>npx</code>)</b></summary>

```json
{
  "mcpServers": {
    "wso2-integration-platform": {
      "command": "npx",
      "args": ["-y", "@pcnfernando-wso2/integration-platform-mcp", "start-mcp-server"]
    }
  }
}
```

For VS Code, use the same values under `servers` with `"type": "stdio"`.
</details>

<details>
<summary><b>Remote / HTTP transport</b></summary>

For containers, shared instances, or clients that connect over HTTP rather than
spawning a process:

```bash
wso2-integration-platform start-mcp-server --http --port 8080
```

It serves MCP at `http://<host>:<port>/mcp` (default port `8080`). Authentication
is **per request** in this mode — pass a token as `Authorization: Bearer <token>`;
the server does not use the local credential store. VS Code example:

```json
{
  "servers": {
    "wso2-integration-platform": {
      "type": "http",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

Three tools are intentionally absent over HTTP (`login`, `check_login_status`,
`change_org`) — they are meaningless when auth arrives per request and the
organization derives from the token.
</details>

**Verifying it works.** Ask the agent to call `get_started`. It should return the
platform knowledge base. If the client lists no tools at all, check the top-level
key (`servers` vs `mcpServers`) and that `wso2-integration-platform` is on the
`PATH` of the process that launched the client — GUI apps often do not inherit a
shell `PATH`, in which case use an absolute path for `command`.

### Option 3 — Direct binary

For environments where npm is not available (CI/CD, WSL, Docker).

```bash
# Install the CLI
curl -sSfL https://raw.githubusercontent.com/wso2/integration-platform-tools/main/scripts/install.sh | sh

# Log in
wso2-integration-platform login

# Add to Claude Code
claude mcp add wso2-integration-platform -- wso2-integration-platform start-mcp-server
```

### Authentication

The MCP server supports three authentication methods:

| Method | How |
|--------|-----|
| **Pre-auth via CLI** (recommended) | Run `wso2-integration-platform login` before starting the MCP server. Credentials are stored encrypted and reused automatically. |
| **Personal Access Token (PAT)** | Set `WSO2IP_PAT=<your-token>` in the MCP server's `env` config. No interactive login needed. |
| **In-session via MCP tool** | Ask Claude to call the `login` tool. Claude will provide a browser URL. After signing in, call `check_login_status` to confirm. |

**Using a PAT in Claude Code:**

```json
{
  "mcpServers": {
    "wso2-integration-platform": {
      "command": "wso2-integration-platform",
      "args": ["start-mcp-server"],
      "env": {
        "WSO2IP_PAT": "your-personal-access-token"
      }
    }
  }
}
```

---

## Development

`main` is the stable branch. `dev` is the development branch. All changes should target `dev`; `main` is updated at release time.

### Feature development

1. Branch from `dev`
2. Make changes
3. Open a PR targeting `dev`

### Bug fixes

1. Branch from `main`
2. Make changes
3. Open a PR targeting `main`

### Build commands

```bash
make build                          # Build CLI binary → bin/wso2-integration-platform
make run                            # Run locally via go run
make test                           # Run unit tests

# Cross-platform compilation
OS=linux  ARCH=amd64 make compile_ip   # bin/linux-amd64/wso2-integration-platform
OS=darwin ARCH=arm64 make compile_ip   # bin/darwin-arm64/wso2-integration-platform
OS=linux  ARCH=amd64 make bundle_mcpb  # bin/linux-amd64/wso2-integration-platform-linux-amd64.mcpb
```

Building requires `go-bindata` on your `PATH` for i18n asset packaging:

```bash
go install github.com/go-bindata/go-bindata/...@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

---

## Release

### Release cycle

Bi-weekly, every second Thursday. All changes merged to `dev` are included. Critical bug fixes are excepted.

### Release process

1. Merge `dev` → `main`
2. Draft a new release in GitHub with the version tag
3. Publish the release

The release workflow automatically:
- Builds `wso2-integration-platform` binaries for all platforms
- Packages `.mcpb` Claude Connector bundles for Linux and macOS
- Attaches all artifacts to the GitHub release
- Publishes `@pcnfernando-wso2/integration-platform-mcp` to npm (stable releases only)

> **Note:** Publishing to npm requires the `NPM_TOKEN` secret to be set in the repository.
