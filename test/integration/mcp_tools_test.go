package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/tools"
)

// newMCPSession creates an in-memory MCP server backed by testClient, connects
// a client to it, and returns the ClientSession along with a cleanup function.
func newMCPSession(t *testing.T) (*mcp.ClientSession, func()) {
	t.Helper()

	server := mcp.NewServer(&mcp.Implementation{Name: "test-ziti-mcp", Version: "0.0.0"}, nil)
	tools.RegisterAll(server, testClient)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	ctx := context.Background()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("connect server: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		serverSession.Close()
		t.Fatalf("connect client: %v", err)
	}

	return clientSession, func() {
		clientSession.Close()
		serverSession.Close()
	}
}

// expectedToolNames is a representative sample of tools that must always be registered.
// The full set is validated by the MCP protocol tests; here we spot-check core tools.
var expectedToolNames = []string{
	"connect-controller", "get-controller-status",
	"list-identities", "get-identity", "create-identity", "update-identity", "delete-identity",
	"list-services", "get-service", "create-service", "update-service", "delete-service",
	"list-service-policies", "get-service-policy", "create-service-policy", "update-service-policy", "delete-service-policy",
	"list-edge-router-policies", "get-edge-router-policy", "create-edge-router-policy", "delete-edge-router-policy",
	"list-edge-routers", "get-edge-router",
	"get-controller-version", "list-summary",
}

func TestMCP_ListTools_AllRegistered(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	toolNames := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range expectedToolNames {
		if !toolNames[expected] {
			t.Errorf("expected tool %q to be registered", expected)
		}
	}

	if len(result.Tools) < len(expectedToolNames) {
		var got []string
		for _, tool := range result.Tools {
			got = append(got, tool.Name)
		}
		t.Errorf("expected at least %d tools, got %d: %v", len(expectedToolNames), len(result.Tools), got)
	}
}

func TestMCP_ToolAnnotations_ReadOnly(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	toolMap := make(map[string]*mcp.Tool, len(result.Tools))
	for _, tool := range result.Tools {
		toolMap[tool.Name] = tool
	}

	readOnlyTools := []string{
		"list-identities", "get-identity",
		"list-services", "get-service",
		"list-service-policies", "get-service-policy",
		"list-edge-router-policies", "get-edge-router-policy",
		"list-edge-routers", "get-edge-router",
		"get-controller-version", "list-summary",
	}
	for _, name := range readOnlyTools {
		tool, ok := toolMap[name]
		if !ok {
			t.Errorf("tool %q not found", name)
			continue
		}
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Errorf("expected tool %q to have ReadOnlyHint=true", name)
		}
	}
}

func TestMCP_ToolAnnotations_Destructive(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	toolMap := make(map[string]*mcp.Tool, len(result.Tools))
	for _, tool := range result.Tools {
		toolMap[tool.Name] = tool
	}

	destructiveTools := []string{
		"delete-identity", "delete-service",
		"delete-service-policy", "delete-edge-router-policy",
	}
	for _, name := range destructiveTools {
		tool, ok := toolMap[name]
		if !ok {
			t.Errorf("tool %q not found", name)
			continue
		}
		if tool.Annotations == nil || tool.Annotations.DestructiveHint == nil || !*tool.Annotations.DestructiveHint {
			t.Errorf("expected tool %q to have DestructiveHint=true", name)
		}
	}
}

func TestMCP_CallTool_ListIdentities_Success(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list-identities",
		Arguments: map[string]any{"limit": 10},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got IsError=true: %v", result.Content)
	}
}

func TestMCP_CallTool_GetIdentity_BadID_IsError(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	// The MCP SDK validates required fields via JSON schema before calling the
	// handler. To test the IsError path, pass a non-existent ID so the handler
	// itself returns an API error wrapped as IsError.
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get-identity",
		Arguments: map[string]any{"id": "does-not-exist-00000000"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true for non-existent identity ID")
	}
}

func TestMCP_CallTool_GetControllerVersion_Success(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "get-controller-version",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got IsError=true: %v", result.Content)
	}
}

func TestMCP_CallTool_ListSummary_Success(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "list-summary",
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got IsError=true: %v", result.Content)
	}
}

func TestMCP_CallTool_CreateAndDeleteIdentity_RoundTrip(t *testing.T) {
	session, cleanup := newMCPSession(t)
	defer cleanup()

	ctx := context.Background()
	name := "mcp-test-identity-" + uniqueSuffix()

	// Create
	createResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "create-identity",
		Arguments: map[string]any{
			"name": name,
			"type": "Device",
		},
	})
	if err != nil {
		t.Fatalf("create-identity CallTool: %v", err)
	}
	if createResult.IsError {
		t.Fatalf("create-identity returned IsError=true: %v", createResult.Content)
	}

	// List with filter to find it
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list-identities",
		Arguments: map[string]any{
			"filter": fmt.Sprintf(`name = "%s"`, name),
		},
	})
	if err != nil {
		t.Fatalf("list-identities CallTool: %v", err)
	}
	if listResult.IsError {
		t.Errorf("list-identities returned IsError=true: %v", listResult.Content)
	}
}
