# mcp-ziti

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that exposes the [OpenZiti](https://openziti.io) Management API to AI agents. It lets any MCP-compatible client — Claude Desktop, Cursor, VS Code Copilot, and others — create, inspect, and manage the resources in an OpenZiti network using natural language.

The server communicates over STDIO. It can authenticate with an OpenZiti controller at startup using one of five methods (identity JSON file, username/password, client certificate, external JWT token, or OIDC client credentials), or it can start **without any credentials** and let the AI agent connect to a controller at runtime via the `connect-controller` tool. For interactive OIDC login via browser (OAuth 2.0 Device Authorization Grant), the agent can use `start-oidc-login` and `complete-oidc-login`.

---

## Tools

| Category | Tool | Description |
|---|---|---|
| **Connection** | `connect-controller` | Connect (or reconnect) to a Ziti controller at runtime |
| | `disconnect-controller` | Disconnect from the current controller and clear credentials |
| | `get-controller-status` | Get the current connection status and controller URL |
| | `start-oidc-login` | Start an interactive OIDC login via browser (Device Authorization Grant) |
| | `complete-oidc-login` | Complete the OIDC device login and connect to the controller |
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
| **Service Edge Router Policies** | `list-service-edge-router-policies` | List service edge router policies |
| | `get-service-edge-router-policy` | Get a single service edge router policy by ID |
| | `create-service-edge-router-policy` | Create a service edge router policy |
| | `update-service-edge-router-policy` | Update a service edge router policy |
| | `delete-service-edge-router-policy` | Permanently delete a service edge router policy |
| **Edge Routers** | `list-edge-routers` | List edge routers |
| | `get-edge-router` | Get a single edge router by ID |
| **Routers** | `list-routers` | List fabric routers (edge and non-edge) |
| | `get-router` | Get a single fabric router by ID |
| **Authenticators** | `list-authenticators` | List authenticators (credentials attached to identities) |
| | `get-authenticator` | Get a single authenticator by ID |
| | `update-authenticator` | Update username/password for an updb authenticator |
| | `delete-authenticator` | Permanently delete an authenticator |
| **Enrollments** | `list-enrollments` | List pending enrollments |
| | `get-enrollment` | Get a single enrollment by ID |
| | `create-enrollment` | Create a new enrollment (ott, ottca, or updb) |
| | `delete-enrollment` | Delete a pending enrollment |
| **Certificate Authorities** | `list-certificate-authorities` | List CAs |
| | `get-certificate-authority` | Get a single CA by ID |
| | `create-certificate-authority` | Create a new CA |
| | `update-certificate-authority` | Update a CA |
| | `delete-certificate-authority` | Permanently delete a CA |
| **External JWT Signers** | `list-external-jwt-signers` | List external JWT signers |
| | `get-external-jwt-signer` | Get a single external JWT signer by ID |
| | `create-external-jwt-signer` | Create an external JWT signer (cert or JWKS) |
| | `update-external-jwt-signer` | Update an external JWT signer |
| | `delete-external-jwt-signer` | Permanently delete an external JWT signer |
| **Auth Policies** | `list-auth-policies` | List authentication policies |
| | `get-auth-policy` | Get a single auth policy by ID |
| | `create-auth-policy` | Create an authentication policy |
| | `update-auth-policy` | Update an authentication policy |
| | `delete-auth-policy` | Permanently delete an auth policy |
| **Configs** | `list-config-types` | List service config types |
| | `get-config-type` | Get a single config type by ID |
| | `create-config-type` | Create a new config type |
| | `delete-config-type` | Permanently delete a config type |
| | `list-configs` | List service configurations |
| | `get-config` | Get a single configuration by ID |
| | `create-config` | Create a new service configuration |
| | `update-config` | Update a service configuration |
| | `delete-config` | Permanently delete a service configuration |
| **Posture Checks** | `list-posture-checks` | List posture checks |
| | `get-posture-check` | Get a single posture check by ID |
| | `list-posture-check-types` | List available posture check types |
| | `delete-posture-check` | Permanently delete a posture check |
| **Terminators** | `list-terminators` | List terminators |
| | `get-terminator` | Get a single terminator by ID |
| | `create-terminator` | Create a terminator linking a service to a router address |
| | `delete-terminator` | Permanently delete a terminator |
| **Sessions** | `list-api-sessions` | List active management API sessions |
| | `get-api-session` | Get a single API session by ID |
| | `delete-api-session` | Force-delete an API session |
| | `list-sessions` | List active network (data-plane) sessions |
| | `get-session` | Get a single network session by ID |
| | `delete-session` | Terminate a network session |
| **Role Attributes** | `list-identity-role-attributes` | List role attributes in use on identities |
| | `list-edge-router-role-attributes` | List role attributes in use on edge routers |
| | `list-service-role-attributes` | List role attributes in use on services |
| | `list-posture-check-role-attributes` | List role attributes in use on posture checks |
| **Network** | `get-controller-version` | Get controller version and build info |
| | `list-summary` | Get a resource count summary for the whole network |
| | `list-controllers` | List controllers in an HA cluster |
| **Database** | `create-database-snapshot` | Trigger an immediate database backup snapshot |
| | `check-data-integrity` | Run an integrity check on the controller database |
| | `fix-data-integrity` | Attempt to fix data integrity issues automatically |

List tools accept `filter`, `limit` (default 100, max 500), and `offset` parameters for filtering and pagination.

---

## Prerequisites

- An [OpenZiti](https://openziti.io) controller reachable from the machine running the MCP server

---

## Installation

### Download a release

Pre-built binaries for macOS, Linux, and Windows are available on the [GitHub Releases page](https://github.com/smilindave26/mcp-ziti/releases).

1. Download the archive for your platform from the [latest release](https://github.com/smilindave26/mcp-ziti/releases/latest).
2. Extract the `ziti-mcp` binary and place it somewhere on your PATH (e.g. `/usr/local/bin`).

```bash
# Example: macOS Apple Silicon
ZMCP_VER="$(curl -sL https://api.github.com/repos/smilindave26/mcp-ziti/releases/latest | sed -n '/tag_name/s/.*v\([0-9]*\.[0-9]*\.[0-9]*\).*/\1/p')"
curl -sL "https://github.com/smilindave26/mcp-ziti/releases/download/v${ZMCP_VER}/mcp-ziti_${ZMCP_VER}_darwin_arm64.tar.gz" | tar xz ziti-mcp
sudo mv ziti-mcp /usr/local/bin/
```

A rolling **Development Build** pre-release with the latest binaries from `main` is also available for testing.

### From source

Requires Go 1.24 or later.

```bash
git clone https://github.com/smilindave26/mcp-ziti.git
cd mcp-ziti
go build -o ziti-mcp .
```

The resulting `ziti-mcp` binary is a single self-contained executable with no runtime dependencies.

---

## Authentication

Credentials are **optional at startup**. If no authentication is configured, the server starts in a disconnected state and the AI agent can connect later using the `connect-controller` tool. When credentials are provided, exactly one authentication method must be used. All options are available as CLI flags or environment variables (flags take precedence).

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
ziti-mcp --identity-file /path/to/identity.json

# Override the controller URL if needed
ziti-mcp --identity-file /path/to/identity.json \
          --controller https://ctrl.example.com:1280
```

Environment variables:
```bash
ZITI_IDENTITY_FILE=/path/to/identity.json ziti-mcp
```

### Username / password

Authenticates with the controller's built-in user database (updb). Requires `--controller`.

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --username admin \
          --password secret
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_USERNAME=admin \
ZITI_PASSWORD=secret \
ziti-mcp
```

### Client certificate

Authenticates using a TLS client certificate and private key. Requires `--controller`.

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --cert /path/to/client.crt \
          --key  /path/to/client.key
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_CERT_FILE=/path/to/client.crt \
ZITI_KEY_FILE=/path/to/client.key \
ziti-mcp
```

### External JWT (static token)

Authenticates using a pre-issued JWT — for example a service account token from your IdP or a long-lived API token issued by the controller. Requires `--controller`.

Provide the token directly as a string:

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --ext-jwt-token eyJhbGciOiJSUzI1NiJ9...
```

Or point to a file containing the token (useful for Kubernetes-mounted secrets):

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --ext-jwt-file /var/run/secrets/token.jwt
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_EXT_JWT_TOKEN=eyJhbGciOiJSUzI1NiJ9... \
ziti-mcp
```

### OIDC client credentials

Authenticates using the [OAuth 2.0 client credentials flow](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4). A fresh token is fetched from the IdP on each session, so no manual token rotation is needed. Requires `--controller` and `--oidc-issuer`.

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --oidc-issuer https://idp.example.com \
          --oidc-client-id my-client \
          --oidc-client-secret my-secret
```

Optional extras:

```bash
# Restrict the token audience
ziti-mcp ... --oidc-audience https://ctrl.example.com

# Skip OIDC discovery and use a known token endpoint directly
ziti-mcp ... --oidc-token-url https://idp.example.com/oauth/token
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_OIDC_ISSUER=https://idp.example.com \
ZITI_OIDC_CLIENT_ID=my-client \
ZITI_OIDC_CLIENT_SECRET=my-secret \
ziti-mcp
```

### Interactive OIDC login (browser)

For users configured via a 3rd-party IdP who don't have a client secret, the AI agent can initiate an interactive browser login using the [OAuth 2.0 Device Authorization Grant (RFC 8628)](https://datatracker.ietf.org/doc/html/rfc8628).

1. The agent calls `start-oidc-login` with the controller URL, OIDC issuer, and client ID
2. The tool returns a verification URL and a user code — the agent presents both to the user
3. The user opens the URL in their browser, enters the code, and authenticates with the IdP
4. The agent calls `complete-oidc-login` which polls the IdP until authentication completes, then connects

Pre-configure the connection details at startup so the agent doesn't need to provide them each time — omit `--oidc-client-secret` to start in disconnected mode with OIDC defaults ready:

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --oidc-issuer https://idp.example.com \
          --oidc-client-id my-public-client
```

Or in your MCP server config:
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/ziti-mcp",
      "args": [
        "--controller", "https://ctrl.example.com:1280",
        "--oidc-issuer", "https://idp.example.com",
        "--oidc-client-id", "my-public-client"
      ]
    }
  }
}
```

The agent can then simply call `start-oidc-login` with no parameters, and all the connection details are filled in from the startup config.

This requires the IdP to support the Device Authorization Grant flow (Auth0, Okta, and Keycloak all do).

### Optional CA override

By default the server fetches the controller's CA bundle from its well-known endpoint. To use a custom CA instead, add `--ca` (or `ZITI_CA_FILE`) to any of the methods above:

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --username admin --password secret \
          --ca /path/to/ca-bundle.pem
```

---

## Agent Configuration

The server communicates over STDIO. Configure it as an MCP server in your agent's settings by pointing to the binary and passing your preferred authentication flags.

> **Windows users:** JSON requires backslashes to be escaped as `\\`. Use double backslashes in all file paths:
>
> ```json
> {
>   "mcpServers": {
>     "ziti": {
>       "command": "C:\\Users\\you\\ziti-mcp.exe",
>       "args": ["--identity-file", "C:\\Users\\you\\.ziti\\identity.json"]
>     }
>   }
> }
> ```

You can also start the server **with no credentials at all** — the agent can then connect to any controller at runtime using the `connect-controller` tool:

```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/ziti-mcp"
    }
  }
}
```

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

**Identity file:**
```json
{
  "mcpServers": {
    "ziti": {
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
      "command": "/usr/local/bin/ziti-mcp",
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
| `--oidc-issuer` | `ZITI_OIDC_ISSUER` | OIDC issuer URL |
| `--oidc-client-id` | `ZITI_OIDC_CLIENT_ID` | OIDC client ID |
| `--oidc-client-secret` | `ZITI_OIDC_CLIENT_SECRET` | OIDC client secret (required for client credentials flow, omit for interactive login) |
| `--oidc-audience` | `ZITI_OIDC_AUDIENCE` | OIDC audience claim (optional) |
| `--oidc-token-url` | `ZITI_OIDC_TOKEN_URL` | OIDC token endpoint URL — skips discovery (optional) |

---

## Building and Testing

### Build

```bash
go build ./...

# Build to an explicit output path
go build -o ziti-mcp .
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
