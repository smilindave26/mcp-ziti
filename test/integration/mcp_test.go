package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/tools"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

// mcpSession creates an in-process MCP client connected to a server that uses
// zc as its Ziti backend.  The session is closed automatically when the test ends.
func mcpSession(t *testing.T, ctx context.Context, zc *ziticlient.Client) *mcp.ClientSession {
	t.Helper()

	srv := mcp.NewServer(&mcp.Implementation{Name: "mcp-ziti", Version: "test"}, nil)
	tools.RegisterAll(srv, zc, nil)

	ct, st := mcp.NewInMemoryTransports()

	srvSession, err := srv.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}

	cli := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	session, err := cli.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}

	t.Cleanup(func() {
		session.Close()   //nolint:errcheck
		srvSession.Wait() //nolint:errcheck
	})

	return session
}

// textContent returns the concatenated text from all TextContent blocks in a result.
func textContent(r *mcp.CallToolResult) string {
	var s string
	for _, c := range r.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			s += tc.Text
		}
	}
	return s
}

// TestListEdgeRouters_ViaProtocol_ResponseFormat calls list-edge-routers through
// the full MCP protocol stack and verifies the response shape:
//   - no protocol-level error
//   - Content contains a JSON array
//   - StructuredContent is nil (not set to an array, which would violate the spec)
func TestListEdgeRouters_ViaProtocol_ResponseFormat(t *testing.T) {
	ctx := context.Background()
	session := mcpSession(t, ctx, testClient)

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list-edge-routers",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %s", textContent(result))
	}

	text := textContent(result)
	if text == "" {
		t.Fatal("expected non-empty text content")
	}

	var items []any
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Errorf("Content is not valid JSON array: %v\nContent: %s", err, text)
	}

	if result.StructuredContent != nil {
		t.Errorf("expected StructuredContent=nil (spec requires object, not array), got %T",
			result.StructuredContent)
	}
}

// TestListIdentities_ViaProtocol_ResponseFormat applies the same response-format
// checks to the list-identities tool.
func TestListIdentities_ViaProtocol_ResponseFormat(t *testing.T) {
	ctx := context.Background()
	session := mcpSession(t, ctx, testClient)

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list-identities",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %s", textContent(result))
	}

	var items []any
	if err := json.Unmarshal([]byte(textContent(result)), &items); err != nil {
		t.Errorf("Content is not valid JSON array: %v", err)
	}

	if result.StructuredContent != nil {
		t.Errorf("expected StructuredContent=nil, got %T", result.StructuredContent)
	}
}

// TestGetEdgeRouter_ViaProtocol_ResponseFormat fetches the first edge router by
// ID and verifies the single-object response format.
func TestGetEdgeRouter_ViaProtocol_ResponseFormat(t *testing.T) {
	ctx := context.Background()
	session := mcpSession(t, ctx, testClient)

	// List first to get a valid ID.
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list-edge-routers",
		Arguments: map[string]any{},
	})
	if err != nil || listResult.IsError {
		t.Skipf("list-edge-routers failed, skipping get test: err=%v isError=%v", err, listResult.IsError)
	}

	var routers []map[string]any
	if err := json.Unmarshal([]byte(textContent(listResult)), &routers); err != nil || len(routers) == 0 {
		t.Skip("no edge routers available")
	}
	routerID, _ := routers[0]["id"].(string)
	if routerID == "" {
		t.Skip("edge router has no id field")
	}

	getResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get-edge-router",
		Arguments: map[string]any{"id": routerID},
	})
	if err != nil {
		t.Fatalf("CallTool get-edge-router: %v", err)
	}
	if getResult.IsError {
		t.Fatalf("tool returned error: %s", textContent(getResult))
	}

	var router map[string]any
	if err := json.Unmarshal([]byte(textContent(getResult)), &router); err != nil {
		t.Errorf("Content is not valid JSON object: %v", err)
	}

	if getResult.StructuredContent != nil {
		t.Errorf("expected StructuredContent=nil, got %T", getResult.StructuredContent)
	}
}
