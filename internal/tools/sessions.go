package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtSession "github.com/openziti/edge-api/rest_management_api_client/session"
)

func registerSessionTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &sessionTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-sessions",
		Description: "List active network sessions (data-plane connections). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-session",
		Description: "Get a single network session by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-session",
		Description: "Terminate a network session by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, t.delete)
}

type sessionTools struct{ zc *ziticlient.Client }

type listSessionsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *sessionTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listSessionsInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtSession.NewListSessionsParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Session.ListSessions(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list sessions: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type getSessionInput struct {
	ID string `json:"id" jsonschema:"required,session ID"`
}

func (t *sessionTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getSessionInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtSession.NewDetailSessionParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Session.DetailSession(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get session %q: %w", in.ID, err)
	}
	return nil, resp.GetPayload().Data, nil
}

type deleteSessionInput struct {
	ID string `json:"id" jsonschema:"required,session ID to terminate"`
}

func (t *sessionTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteSessionInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtSession.NewDeleteSessionParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Session.DeleteSession(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete session %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "deleted", "id": in.ID}, nil
}
