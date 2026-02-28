package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/netfoundry/mcp-ziti-golang/internal/tools"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	// All log output goes to stderr so it never corrupts the STDIO MCP stream.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	zc, err := ziticlient.New(cfg)
	if err != nil {
		return fmt.Errorf("connecting to Ziti controller: %w", err)
	}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "ziti-mcp",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: buildInstructions(zc, cfg),
	})

	tools.RegisterAll(s, zc, cfg)

	slog.Info("starting MCP server over STDIO")
	return s.Run(context.Background(), &mcp.StdioTransport{})
}

// buildInstructions returns the MCP server instructions shown to the LLM
// during initialization, including connection status and version compatibility.
func buildInstructions(zc *ziticlient.Client, cfg *config.Config) string {
	base := "This server exposes the OpenZiti Management API. " +
		"Use the provided tools to create, inspect, and manage resources in an OpenZiti network."

	if !zc.Connected() {
		hint := "STATUS: Not connected to a controller. " +
			"Connect before calling any other tools.\n\n" +
			"IMPORTANT: Present ALL of the following options to the user as separate choices. Do NOT combine or merge them:\n" +
			"  1. Interactive browser login — user opens a link and authenticates in their browser (best for human users with a 3rd-party identity provider)\n" +
			"  2. Identity JSON file — provide a path to a Ziti identity .json file on disk\n" +
			"  3. Username and password — authenticate with the controller's built-in user database\n" +
			"  4. Client certificate — authenticate with a TLS client cert and private key\n" +
			"  5. External JWT token — authenticate with a pre-issued JWT string\n" +
			"  6. OIDC client credentials — authenticate with a client ID and secret from an identity provider (for service accounts)"

		if cfg != nil && cfg.OIDCIssuer != "" && cfg.OIDCClientID != "" {
			hint += fmt.Sprintf(
				"\n\nOIDC defaults are pre-configured (issuer: %s, client: %s). "+
					"If the user picks option 1, call start-oidc-login with no parameters.",
				cfg.OIDCIssuer, cfg.OIDCClientID)
		}

		return base + "\n\n" + hint
	}

	info := zc.GetVersionInfo()
	if info == nil {
		return base + "\n\n" +
			fmt.Sprintf("STATUS: Connected to %s.", zc.ControllerURL())
	}

	return base + "\n\n" +
		fmt.Sprintf("STATUS: Connected to %s (controller %s).\n", zc.ControllerURL(), info.ControllerVersion) +
		fmt.Sprintf("API compatibility: %s\n", info.CompatibilityNote)
}
