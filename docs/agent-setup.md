# Agent Setup

The server communicates over STDIO. Configure it as an MCP server in your agent's settings by pointing to the binary and passing your preferred authentication flags.

!!! tip "Windows users"

    JSON requires backslashes to be escaped as `\\`. Use double backslashes in all file paths:

    ```json
    {
      "mcpServers": {
        "ziti": {
          "command": "C:\\Users\\you\\ziti-mcp.exe",
          "args": ["--identity-file", "C:\\Users\\you\\.ziti\\identity.json"]
        }
      }
    }
    ```

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

## Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

=== "Identity file"

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

=== "Username / password"

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

=== "Client certificate"

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

=== "Environment variables"

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

## Cursor

Edit `.cursor/mcp.json` in your project root, or `~/.cursor/mcp.json` for a global configuration:

=== "Identity file"

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

=== "Username / password"

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

=== "Client certificate"

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

## VS Code (GitHub Copilot)

Add to your `.vscode/mcp.json` or user `settings.json` under `"mcp"`:

=== "Identity file"

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

=== "Username / password"

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

=== "Client certificate"

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

## Claude Code (CLI)

Add to your project's `.claude/settings.json` or `~/.claude/settings.json`:

=== "Identity file"

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

=== "Username / password"

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
