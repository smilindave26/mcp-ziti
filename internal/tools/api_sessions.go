package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtAPISession "github.com/openziti/edge-api/rest_management_api_client/api_session"
)

func registerAPISessionTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &apiSessionTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-api-sessions",
		Description: "List active API sessions (authenticated management connections). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-api-session",
		Description: "Get a single API session by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-api-session",
		Description: "Delete (force-logout) an API session by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, t.delete)
}

type apiSessionTools struct{ zc *ziticlient.Client }

type listAPISessionsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *apiSessionTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listAPISessionsInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAPISession.NewListAPISessionsParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.APISession.ListAPISessions(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list api sessions: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type getAPISessionInput struct {
	ID string `json:"id" jsonschema:"required,API session ID"`
}

func (t *apiSessionTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getAPISessionInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAPISession.NewDetailAPISessionsParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.APISession.DetailAPISessions(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get api session %q: %w", in.ID, err)
	}
	return nil, resp.GetPayload().Data, nil
}

type deleteAPISessionInput struct {
	ID string `json:"id" jsonschema:"required,API session ID to delete"`
}

func (t *apiSessionTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteAPISessionInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAPISession.NewDeleteAPISessionsParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.APISession.DeleteAPISessions(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete api session %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "deleted", "id": in.ID}, nil
}
