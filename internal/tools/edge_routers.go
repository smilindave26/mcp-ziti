package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtER "github.com/openziti/edge-api/rest_management_api_client/edge_router"
)

func registerEdgeRouterTools(s *mcp.Server, zc *ziticlient.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-edge-routers",
		Description: "List edge routers in the Ziti network. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "edge routers", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtER.NewListEdgeRoutersParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.EdgeRouter.ListEdgeRouters(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-edge-router",
		Description: "Get a single edge router by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "edge router", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.EdgeRouter.DetailEdgeRouter(
			mgmtER.NewDetailEdgeRouterParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))
}
