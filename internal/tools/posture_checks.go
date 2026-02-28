package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtPC "github.com/openziti/edge-api/rest_management_api_client/posture_checks"
)

func registerPostureCheckTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &postureCheckTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-posture-checks",
		Description: "List posture checks (zero-trust device health requirements). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-posture-check",
		Description: "Get a single posture check by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-posture-check-types",
		Description: "List available posture check types (e.g. OS, domain, MFA, process).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.listTypes)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-posture-check",
		Description: "Permanently delete a posture check by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type postureCheckTools struct{ zc *ziticlient.Client }

type listPostureChecksInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *postureCheckTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listPostureChecksInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtPC.NewListPostureChecksParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.PostureChecks.ListPostureChecks(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list posture checks: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getPostureCheckInput struct {
	ID string `json:"id" jsonschema:"required,posture check ID"`
}

func (t *postureCheckTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getPostureCheckInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtPC.NewDetailPostureCheckParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.PostureChecks.DetailPostureCheck(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get posture check %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type listPostureCheckTypesInput struct {
	Limit  int64 `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64 `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *postureCheckTools) listTypes(ctx context.Context, _ *mcp.CallToolRequest, in listPostureCheckTypesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtPC.NewListPostureCheckTypesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	resp, err := mgmt.PostureChecks.ListPostureCheckTypes(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list posture check types: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type deletePostureCheckInput struct {
	ID string `json:"id" jsonschema:"required,posture check ID to delete"`
}

func (t *postureCheckTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deletePostureCheckInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtPC.NewDeletePostureCheckParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.PostureChecks.DeletePostureCheck(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete posture check %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
