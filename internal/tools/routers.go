package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtRouter "github.com/openziti/edge-api/rest_management_api_client/router"
)

func registerRouterTools(s *mcp.Server, zc *ziticlient.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-routers",
		Description: "List fabric routers (both edge and non-edge). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "routers", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtRouter.NewListRoutersParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Router.ListRouters(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-router",
		Description: "Get a single fabric router by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "router", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Router.DetailRouter(
			mgmtRouter.NewDetailRouterParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))
}
