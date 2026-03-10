package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtCtrl "github.com/openziti/edge-api/rest_management_api_client/controllers"
)

func registerControllerTools(s *mcp.Server, zc *ziticlient.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-controllers",
		Description: "List controllers in an HA (high-availability) cluster.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "controllers", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtCtrl.NewListControllersParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Controllers.ListControllers(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))
}
