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
		Version: "0.1.0",
	}, nil)

	tools.RegisterAll(s, zc)

	slog.Info("starting MCP server over STDIO")
	return s.Run(context.Background(), &mcp.StdioTransport{})
}
