# Build and Test

## Prerequisites

- Go 1.24+
- (For integration tests) `ziti` binary in PATH — install from https://github.com/openziti/ziti/releases

## Build

```bash
# Build binary to default output
go build ./...

# Build with explicit output path
go build -o bin/mcp-ziti .
```

## Run

The server communicates over STDIO (MCP protocol). Start it with one of the three
auth methods:

**Identity JSON file:**
```bash
./bin/mcp-ziti --controller https://ctrl.example.com:1280 --identity-file /path/to/identity.json

# If the identity file contains a ztAPI field, --controller can be omitted:
./bin/mcp-ziti --identity-file /path/to/identity.json
```

**Username / password:**
```bash
./bin/mcp-ziti --controller https://ctrl.example.com:1280 --username admin --password secret
```

**Client certificate:**
```bash
./bin/mcp-ziti --controller https://ctrl.example.com:1280 --cert /path/to/client.crt --key /path/to/client.key
```

**Environment variables** (all flags have env var equivalents):
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_USERNAME=admin \
ZITI_PASSWORD=secret \
./bin/mcp-ziti
```
CLI flags take precedence over env vars.

**Optional CA override:**
```bash
./bin/mcp-ziti --controller https://ctrl.example.com:1280 --username admin --password secret --ca /path/to/ca.pem
```
If `--ca` is not set, the controller's well-known CA bundle is fetched automatically.

## Unit Tests

```bash
go test ./internal/...
```

## Integration Tests

Integration tests spin up a local Ziti network using `ziti edge quickstart`.
They are skipped automatically if `ziti` is not found in PATH.

```bash
go test -v -timeout 3m ./test/integration/...
```

To run a specific test:
```bash
go test -v -timeout 3m -run TestCreateAndListIdentity ./test/integration/...
```

## Lint

```bash
golangci-lint run ./...
```

## MCP Client Configuration (Claude Desktop example)

```json
{
  "mcpServers": {
    "ziti": {
      "command": "/path/to/mcp-ziti",
      "args": ["--controller", "https://ctrl.example.com:1280", "--username", "admin", "--password", "secret"]
    }
  }
}
```
