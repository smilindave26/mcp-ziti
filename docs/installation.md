# Installation

## Download a release

Pre-built binaries for macOS, Linux, and Windows are available on the [GitHub Releases page](https://github.com/smilindave26/mcp-ziti/releases).

1. Download the archive for your platform from the [latest release](https://github.com/smilindave26/mcp-ziti/releases/latest).
2. Extract the `ziti-mcp` binary and place it somewhere on your PATH (e.g. `/usr/local/bin`).

```bash
# Example: macOS Apple Silicon
curl -sL https://github.com/smilindave26/mcp-ziti/releases/latest/download/ziti-mcp_$(curl -sL https://api.github.com/repos/smilindave26/mcp-ziti/releases/latest | grep tag_name | cut -d'"' -f4 | sed 's/^v//')_darwin_arm64.tar.gz | tar xz
sudo mv ziti-mcp /usr/local/bin/
```

A rolling **Development Build** pre-release with the latest binaries from `main` is also available for testing.

## From source

Requires Go 1.24 or later.

```bash
git clone https://github.com/smilindave26/mcp-ziti.git
cd mcp-ziti
go build -o ziti-mcp .
```

The resulting `ziti-mcp` binary is a single self-contained executable with no runtime dependencies.
