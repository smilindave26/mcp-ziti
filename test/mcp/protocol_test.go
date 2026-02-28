package mcp_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestToolsList_AtLeast70Tools verifies that all tool registrations fired.
func TestToolsList_AtLeast70Tools(t *testing.T) {
	ctx := context.Background()
	result, err := protocolSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(result.Tools) < 70 {
		t.Errorf("expected at least 70 tools, got %d", len(result.Tools))
	}
}

// TestToolsList_AllInputSchemasAreObjects verifies that every tool's InputSchema
// has type "object".  A non-object schema would cause clients to reject tool calls.
func TestToolsList_AllInputSchemasAreObjects(t *testing.T) {
	ctx := context.Background()
	result, err := protocolSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		schema, ok := tool.InputSchema.(map[string]any)
		if !ok {
			t.Errorf("tool %q: InputSchema is not a map, got %T", tool.Name, tool.InputSchema)
			continue
		}
		typ, _ := schema["type"].(string)
		if typ != "object" {
			t.Errorf("tool %q: InputSchema.type = %q, want %q", tool.Name, typ, "object")
		}
	}
}

// TestToolCall_HandlerError_WrappedInContent verifies that errors returned by
// a tool handler appear in CallToolResult.Content with IsError=true, not as a
// protocol-level error.  This is the correct MCP behaviour: it lets the LLM
// see and react to the error rather than losing it in the transport layer.
func TestToolCall_HandlerError_WrappedInContent(t *testing.T) {
	ctx := context.Background()
	// list-edge-routers will fail because there is no real Ziti controller.
	result, err := protocolSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list-edge-routers",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected protocol error (expected handler error in Content): %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true, got false")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected error message in Content, got empty Content")
	}
}

// TestToolCall_StructuredContent_NilOnError verifies that StructuredContent is
// not populated on error responses.
func TestToolCall_StructuredContent_NilOnError(t *testing.T) {
	ctx := context.Background()
	result, err := protocolSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list-edge-routers",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if result.StructuredContent != nil {
		t.Errorf("expected StructuredContent=nil on error, got %T: %v",
			result.StructuredContent, result.StructuredContent)
	}
}

// TestToolCall_UnknownTool_ReturnsProtocolError verifies that calling a
// non-existent tool returns a protocol-level error (not a wrapped tool error).
func TestToolCall_UnknownTool_ReturnsProtocolError(t *testing.T) {
	ctx := context.Background()
	_, err := protocolSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "this-tool-does-not-exist",
	})
	if err == nil {
		t.Fatal("expected protocol error for unknown tool, got nil")
	}
}

// TestToolCall_AdditionalProperties_Rejected verifies that passing a field not
// declared in the InputSchema is rejected at the protocol level (before the
// handler runs) because the schema has additionalProperties: false.
func TestToolCall_AdditionalProperties_Rejected(t *testing.T) {
	ctx := context.Background()
	_, err := protocolSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list-edge-routers",
		Arguments: map[string]any{"unknown_field": "bad"},
	})
	if err == nil {
		t.Fatal("expected protocol error for unknown input property (additionalProperties: false), got nil")
	}
}
