package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtER "github.com/openziti/edge-api/rest_management_api_client/edge_router"
)

func registerEdgeRouterTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &edgeRouterTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-edge-routers",
		Description: "List edge routers in the Ziti network. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-edge-router",
		Description: "Get a single edge router by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)
}

type edgeRouterTools struct{ zc *ziticlient.Client }

type listEdgeRoutersInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *edgeRouterTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listEdgeRoutersInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtER.NewListEdgeRoutersParams().WithContext(ctx)
	params.Limit = &limit
	params.Offset = &offset
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.EdgeRouter.ListEdgeRouters(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list edge routers: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getEdgeRouterInput struct {
	ID string `json:"id" jsonschema:"required,edge router ID"`
}

func (t *edgeRouterTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getEdgeRouterInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtER.NewDetailEdgeRouterParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.EdgeRouter.DetailEdgeRouter(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get edge router %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}
