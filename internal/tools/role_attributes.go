package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtRA "github.com/openziti/edge-api/rest_management_api_client/role_attributes"
)

func registerRoleAttributeTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &roleAttributeTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-identity-role-attributes",
		Description: "List all role attribute values currently in use on identities.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.listIdentity)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-edge-router-role-attributes",
		Description: "List all role attribute values currently in use on edge routers.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.listEdgeRouter)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-service-role-attributes",
		Description: "List all role attribute values currently in use on services.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.listService)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-posture-check-role-attributes",
		Description: "List all role attribute values currently in use on posture checks.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.listPostureCheck)
}

type roleAttributeTools struct{ zc *ziticlient.Client }

type listRoleAttributesInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *roleAttributeTools) listIdentity(ctx context.Context, _ *mcp.CallToolRequest, in listRoleAttributesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtRA.NewListIdentityRoleAttributesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.RoleAttributes.ListIdentityRoleAttributes(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list identity role attributes: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

func (t *roleAttributeTools) listEdgeRouter(ctx context.Context, _ *mcp.CallToolRequest, in listRoleAttributesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtRA.NewListEdgeRouterRoleAttributesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.RoleAttributes.ListEdgeRouterRoleAttributes(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list edge router role attributes: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

func (t *roleAttributeTools) listService(ctx context.Context, _ *mcp.CallToolRequest, in listRoleAttributesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtRA.NewListServiceRoleAttributesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.RoleAttributes.ListServiceRoleAttributes(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list service role attributes: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

func (t *roleAttributeTools) listPostureCheck(ctx context.Context, _ *mcp.CallToolRequest, in listRoleAttributesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtRA.NewListPostureCheckRoleAttributesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.RoleAttributes.ListPostureCheckRoleAttributes(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list posture check role attributes: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}
