package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtERP "github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_model"
)

func registerEdgeRouterPolicyTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &edgeRouterPolicyTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-edge-router-policies",
		Description: "List edge router policies. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "edge router policies", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtERP.NewListEdgeRouterPoliciesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.EdgeRouterPolicy.ListEdgeRouterPolicies(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-edge-router-policy",
		Description: "Get a single edge router policy by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "edge router policy", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.EdgeRouterPolicy.DetailEdgeRouterPolicy(
			mgmtERP.NewDetailEdgeRouterPolicyParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-edge-router-policy",
		Description: "Create an edge router policy linking identities to edge routers. semantic must be AllOf or AnyOf. Roles use #tag or @name syntax.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-edge-router-policy",
		Description: "Permanently delete an edge router policy by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "edge router policy", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.EdgeRouterPolicy.DeleteEdgeRouterPolicy(
			mgmtERP.NewDeleteEdgeRouterPolicyParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type edgeRouterPolicyTools struct{ zc *ziticlient.Client }

type createEdgeRouterPolicyInput struct {
	Name            string   `json:"name"                       jsonschema:"required,policy name"`
	Semantic        string   `json:"semantic"                   jsonschema:"required,AllOf or AnyOf"`
	EdgeRouterRoles []string `json:"edgeRouterRoles,omitempty"  jsonschema:"edge router role selectors e.g. ['#tag','@name']"`
	IdentityRoles   []string `json:"identityRoles,omitempty"    jsonschema:"identity role selectors e.g. ['#tag','@name']"`
}

func (t *edgeRouterPolicyTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createEdgeRouterPolicyInput) (*mcp.CallToolResult, any, error) {
	semantic := rest_model.Semantic(in.Semantic)
	body := &rest_model.EdgeRouterPolicyCreate{
		Name:            &in.Name,
		Semantic:        &semantic,
		EdgeRouterRoles: rest_model.Roles(in.EdgeRouterRoles),
		IdentityRoles:   rest_model.Roles(in.IdentityRoles),
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtERP.NewCreateEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(body)
	resp, err := mgmt.EdgeRouterPolicy.CreateEdgeRouterPolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create edge router policy: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}
