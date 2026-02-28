package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtRouter "github.com/openziti/edge-api/rest_management_api_client/router"
)

func registerRouterTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &routerTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-routers",
		Description: "List fabric routers (both edge and non-edge). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-router",
		Description: "Get a single fabric router by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)
}

type routerTools struct{ zc *ziticlient.Client }

type listRoutersInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *routerTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listRoutersInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtRouter.NewListRoutersParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Router.ListRouters(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list routers: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getRouterInput struct {
	ID string `json:"id" jsonschema:"required,router ID"`
}

func (t *routerTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getRouterInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtRouter.NewDetailRouterParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Router.DetailRouter(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get router %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}
