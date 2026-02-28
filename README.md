# mcp-ziti

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that exposes the [OpenZiti](https://openziti.io) Management API to AI agents. It lets any MCP-compatible client — Claude Desktop, Cursor, VS Code Copilot, and others — create, inspect, and manage the resources in an OpenZiti network using natural language.

The server communicates over STDIO and authenticates with an OpenZiti controller using one of five methods: an identity JSON file, username/password, a client certificate, an external JWT token, or OIDC client credentials.

---

## Tools

| Category | Tool | Description |
|---|---|---|
| **Identities** | `list-identities` | List identities with optional filter and pagination |
| | `get-identity` | Get a single identity by ID |
| | `create-identity` | Create a new identity (Device, User, Router, or Service) |
| | `update-identity` | Rename, change type, toggle admin flag, or update role attributes |
| | `delete-identity` | Permanently delete an identity |
| **Services** | `list-services` | List services with optional filter and pagination |
| | `get-service` | Get a single service by ID |
| | `create-service` | Create a new service |
| | `update-service` | Update a service's name, encryption setting, or role attributes |
| | `delete-service` | Permanently delete a service |
| **Service Policies** | `list-service-policies` | List service policies |
| | `get-service-policy` | Get a single service policy by ID |
| | `create-service-policy` | Create a Dial or Bind service policy |
| | `update-service-policy` | Update a service policy |
| | `delete-service-policy` | Permanently delete a service policy |
| **Edge Router Policies** | `list-edge-router-policies` | List edge router policies |
| | `get-edge-router-policy` | Get a single edge router policy by ID |
| | `create-edge-router-policy` | Create an edge router policy |
| | `delete-edge-router-policy` | Permanently delete an edge router policy |
| **Edge Routers** | `list-edge-routers` | List edge routers |
| | `get-edge-router` | Get a single edge router by ID |
| **Network** | `get-controller-version` | Get controller version and build info |
| | `list-summary` | Get a resource count summary for the whole network |

List tools accept `filter`, `limit` (default 100, max 500), and `offset` parameters for filtering and pagination.

---

## Prerequisites

- Go 1.24 or later
- An [OpenZiti](https://openziti.io) controller reachable from the machine running the MCP server

---

## Installation

**From source:**

```bash
git clone https://github.com/netfoundry/mcp-ziti-golang.git
cd mcp-ziti-golang
go build -o mcp-ziti .
```

The resulting `mcp-ziti` binary is a single self-contained executable with no runtime dependencies.

---

## Authentication

Exactly one authentication method must be configured. All options are available as CLI flags or environment variables (flags take precedence).

### Identity JSON file

A Ziti identity file contains the controller URL, client certificate, and CA bundle in a single JSON file. This is the simplest option when you already have an enrolled identity.

> **How to get an identity file:** In your [OpenZiti](https://openziti.io) network, create an identity and download its one-time enrollment JWT, then enroll it with the `ziti` CLI to produce the JSON file:
> ```bash
> # Create an identity via the CLI or web console, then enroll with the JWT
> ziti edge enroll --jwt /path/to/identity.jwt --out /path/to/identity.json
> ```
> The resulting `identity.json` is what you pass to `--identity-file`. See the [OpenZiti enrollment docs](https://openziti.io/docs/learn/core-concepts/identities/enrolling) for full details, including enrollment via the [CloudZiti console](https://cloudziti.io) or the management API.

```bash
# The controller URL is read from the ztAPI field inside the file
mcp-ziti --identity-file /path/to/identity.json

# Override the controller URL if needed
mcp-ziti --identity-file /path/to/identity.json \
          --controller https://ctrl.example.com:1280
```

Environment variables:
```bash
ZITI_IDENTITY_FILE=/path/to/identity.json mcp-ziti
```

### Username / password

Authenticates with the controller's built-in user database (updb). Requires `--controller`.

```bash
mcp-ziti --controller https://ctrl.example.com:1280 \
          --username admin \
          --password secret
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_USERNAME=admin \
ZITI_PASSWORD=secret \
mcp-ziti
```

### Client certificate

Authenticates using a TLS client certificate and private key. Requires `--controller`.

```bash
mcp-ziti --controller https://ctrl.example.com:1280 \
          --cert /path/to/client.crt \
          --key  /path/to/client.key
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_CERT_FILE=/path/to/client.crt \
ZITI_KEY_FILE=/path/to/client.key \
mcp-ziti
```

### External JWT (static token)

Authenticates using a pre-issued JWT — for example a service account token from your IdP or a long-lived API token issued by the controller. Requires `--controller`.

Provide the token directly as a string:

```bash
mcp-ziti --controller https://ctrl.example.com:1280 \
          --ext-jwt-token eyJhbGciOiJSUzI1NiJ9...
```

Or point to a file containing the token (useful for Kubernetes-mounted secrets):

```bash
mcp-ziti --controller https://ctrl.example.com:1280 \
          --ext-jwt-file /var/run/secrets/token.jwt
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_EXT_JWT_TOKEN=eyJhbGciOiJSUzI1NiJ9... \
mcp-ziti
```

### OIDC client credentials

Authenticates using the [OAuth 2.0 client credentials flow](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4). A fresh token is fetched from the IdP on each session, so no manual token rotation is needed. Requires `--controller` and `--oidc-issuer`.

```bash
mcp-ziti --controller https://ctrl.example.com:1280 \
          --oidc-issuer https://idp.example.com \
          --oidc-client-id my-client \
          --oidc-client-secret my-secret
```

Optional extras:

```bash
# Restrict the token audience
mcp-ziti ... --oidc-audience https://ctrl.example.com

# Skip OIDC discovery and use a known token endpoint directly
mcp-ziti ... --oidc-token-url https://idp.example.com/oauth/token
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_OIDC_ISSUER=https://idp.example.com \
ZITI_OIDC_CLIENT_ID=my-client \
ZITI_OIDC_CLIENT_SECRET=my-secret \
mcp-ziti
```

### Optional CA override

By default the server fetches the controller's CA bundle from its well-known endpoint. To use a custom CA instead, add `--ca` (or `ZITI_CA_FILE`) to any of the methods above:

```bash
mcp-ziti --controller https://ctrl.example.com:1280 \
          --username admin --password secret \
          --ca /path/to/ca-bundle.pem
```

---

## Agent Configuration

The server communicates over STDIO. Configure it as an MCP server in your agent's settings by pointing to the binary and passing your preferred authentication flags.

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

**Identity file:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": ["--identity-file", "/path/to/identity.json"]
    }
  }
}
```

**Username / password:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--username", "admin",
        "--password", "secret"
      ]
    }
  }
}
```

**Client certificate:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--cert", "/path/to/client.crt",
        "--key",  "/path/to/client.key"
      ]
    }
  }
}
```

**Using environment variables** (keeps credentials out of the config file):
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "env": {
        "ZITI_CONTROLLER_URL": "https://ctrl.example.com:1280",
        "ZITI_USERNAME": "admin",
        "ZITI_PASSWORD": "secret"
      }
    }
  }
}
```

### Cursor

Edit `.cursor/mcp.json` in your project root, or `~/.cursor/mcp.json` for a global configuration:

**Identity file:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": ["--identity-file", "/path/to/identity.json"]
    }
  }
}
```

**Username / password:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--username", "admin",
        "--password", "secret"
      ]
    }
  }
}
```

**Client certificate:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--cert", "/path/to/client.crt",
        "--key",  "/path/to/client.key"
      ]
    }
  }
}
```

### VS Code (GitHub Copilot)

Add to your `.vscode/mcp.json` or user `settings.json` under `"mcp"`:

**Identity file:**
```json
{
  "servers": {
    "ziti": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-ziti",
      "args": ["--identity-file", "/path/to/identity.json"]
    }
  }
}
```

**Username / password:**
```json
{
  "servers": {
    "ziti": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-ziti",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--username", "admin",
        "--password", "secret"
      ]
    }
  }
}
```

**Client certificate:**
```json
{
  "servers": {
    "ziti": {
      "type": "stdio",
      "command": "/usr/local/bin/mcp-ziti",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--cert", "/path/to/client.crt",
        "--key",  "/path/to/client.key"
      ]
    }
  }
}
```

### Claude Code (CLI)

Add to your project's `.claude/settings.json` or `~/.claude/settings.json`:

**Identity file:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": ["--identity-file", "/path/to/identity.json"]
    }
  }
}
```

**Username / password:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/mcp-ziti",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--username", "admin",
        "--password", "secret"
      ]
    }
  }
}
```

---

## Flag Reference

| Flag | Env var | Description |
|---|---|---|
| `--controller` | `ZITI_CONTROLLER_URL` | Controller URL, e.g. `https://ctrl.example.com:1280` |
| `--identity-file` | `ZITI_IDENTITY_FILE` | Path to a Ziti identity JSON file |
| `--username` | `ZITI_USERNAME` | Username for updb authentication |
| `--password` | `ZITI_PASSWORD` | Password for updb authentication |
| `--cert` | `ZITI_CERT_FILE` | Path to a PEM client certificate file |
| `--key` | `ZITI_KEY_FILE` | Path to a PEM private key file |
| `--ca` | `ZITI_CA_FILE` | Path to a PEM CA bundle (optional override) |
| `--ext-jwt-token` | `ZITI_EXT_JWT_TOKEN` | External JWT token string |
| `--ext-jwt-file` | `ZITI_EXT_JWT_FILE` | Path to a file containing an external JWT |
| `--oidc-issuer` | `ZITI_OIDC_ISSUER` | OIDC issuer URL for client credentials flow |
| `--oidc-client-id` | `ZITI_OIDC_CLIENT_ID` | OIDC client ID |
| `--oidc-client-secret` | `ZITI_OIDC_CLIENT_SECRET` | OIDC client secret |
| `--oidc-audience` | `ZITI_OIDC_AUDIENCE` | OIDC audience claim (optional) |
| `--oidc-token-url` | `ZITI_OIDC_TOKEN_URL` | OIDC token endpoint URL — skips discovery (optional) |

---

## Building and Testing

### Build

```bash
go build ./...

# Build to an explicit output path
go build -o mcp-ziti .
```

### Unit tests

No external dependencies required.

```bash
go test ./internal/...
```

### Integration tests

Integration tests spin up a live [OpenZiti](https://openziti.io) network using `ziti edge quickstart` and run all 23 tools against it. They are skipped automatically if the `ziti` binary is not in PATH.

Install the `ziti` binary from the [OpenZiti releases page](https://github.com/openziti/ziti/releases), then:

```bash
go test -v -timeout 5m ./test/integration/...
```

Run a single test by name:

```bash
go test -v -timeout 5m -run TestFullWorkflow ./test/integration/...
```

### Lint

```bash
golangci-lint run ./...
```
