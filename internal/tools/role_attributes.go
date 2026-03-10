package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtRA "github.com/openziti/edge-api/rest_management_api_client/role_attributes"
)

func registerRoleAttributeTools(s *mcp.Server, zc *ziticlient.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-identity-role-attributes",
		Description: "List all role attribute values currently in use on identities.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "identity role attributes", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtRA.NewListIdentityRoleAttributesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.RoleAttributes.ListIdentityRoleAttributes(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-edge-router-role-attributes",
		Description: "List all role attribute values currently in use on edge routers.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "edge router role attributes", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtRA.NewListEdgeRouterRoleAttributesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.RoleAttributes.ListEdgeRouterRoleAttributes(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-service-role-attributes",
		Description: "List all role attribute values currently in use on services.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "service role attributes", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtRA.NewListServiceRoleAttributesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.RoleAttributes.ListServiceRoleAttributes(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-posture-check-role-attributes",
		Description: "List all role attribute values currently in use on posture checks.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "posture check role attributes", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtRA.NewListPostureCheckRoleAttributesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.RoleAttributes.ListPostureCheckRoleAttributes(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))
}
