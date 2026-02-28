package mcp_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/tools"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

// protocolSession is connected to an in-process MCP server backed by a stub
// ziticlient.  Tool calls fail gracefully (wrapped in CallToolResult.IsError),
// but the full MCP protocol layer—schema generation, input validation, error
// wrapping—is exercised without needing a real Ziti controller.
var protocolSession *mcp.ClientSession

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())

	srv := mcp.NewServer(&mcp.Implementation{Name: "mcp-ziti", Version: "test"}, nil)
	tools.RegisterAll(srv, ziticlient.NewForTest())

	ct, st := mcp.NewInMemoryTransports()

	srvSession, err := srv.Connect(ctx, st, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: server.Connect:", err)
		cancel()
		os.Exit(1)
	}

	cli := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	protocolSession, err = cli.Connect(ctx, ct, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: client.Connect:", err)
		cancel()
		os.Exit(1)
	}

	code := m.Run()

	protocolSession.Close() //nolint:errcheck
	cancel()
	srvSession.Wait() //nolint:errcheck

	os.Exit(code)
}
