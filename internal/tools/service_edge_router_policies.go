package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtSERP "github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_model"
)

func registerServiceEdgeRouterPolicyTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &serviceEdgeRouterPolicyTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-service-edge-router-policies",
		Description: "List service edge router policies. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "service edge router policies", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtSERP.NewListServiceEdgeRouterPoliciesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-service-edge-router-policy",
		Description: "Get a single service edge router policy by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "service edge router policy", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.ServiceEdgeRouterPolicy.DetailServiceEdgeRouterPolicy(
			mgmtSERP.NewDetailServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-service-edge-router-policy",
		Description: "Create a service edge router policy that controls which edge routers can host a service. semantic must be AllOf or AnyOf.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-service-edge-router-policy",
		Description: "Update a service edge router policy's name, roles, or semantic.",
	}, t.update)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-service-edge-router-policy",
		Description: "Permanently delete a service edge router policy by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "service edge router policy", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(
			mgmtSERP.NewDeleteServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type serviceEdgeRouterPolicyTools struct{ zc *ziticlient.Client }

type createSERPInput struct {
	Name            string   `json:"name"                   jsonschema:"required,policy name"`
	Semantic        string   `json:"semantic"               jsonschema:"required,role matching semantic: AllOf or AnyOf"`
	ServiceRoles    []string `json:"serviceRoles,omitempty" jsonschema:"service role selectors (e.g. ['#tag', '@id'])"`
	EdgeRouterRoles []string `json:"edgeRouterRoles,omitempty" jsonschema:"edge router role selectors (e.g. ['#all'])"`
}

func (t *serviceEdgeRouterPolicyTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createSERPInput) (*mcp.CallToolResult, any, error) {
	semantic := rest_model.Semantic(in.Semantic)
	body := &rest_model.ServiceEdgeRouterPolicyCreate{
		Name:            &in.Name,
		Semantic:        &semantic,
		ServiceRoles:    rest_model.Roles(in.ServiceRoles),
		EdgeRouterRoles: rest_model.Roles(in.EdgeRouterRoles),
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtSERP.NewCreateServiceEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(body)
	resp, err := mgmt.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create service edge router policy: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type updateSERPInput struct {
	ID              string   `json:"id"                     jsonschema:"required,policy ID to update"`
	Name            string   `json:"name"                   jsonschema:"required,policy name"`
	Semantic        string   `json:"semantic"               jsonschema:"required,role matching semantic: AllOf or AnyOf"`
	ServiceRoles    []string `json:"serviceRoles,omitempty" jsonschema:"service role selectors"`
	EdgeRouterRoles []string `json:"edgeRouterRoles,omitempty" jsonschema:"edge router role selectors"`
}

func (t *serviceEdgeRouterPolicyTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateSERPInput) (*mcp.CallToolResult, any, error) {
	semantic := rest_model.Semantic(in.Semantic)
	body := &rest_model.ServiceEdgeRouterPolicyUpdate{
		Name:            &in.Name,
		Semantic:        &semantic,
		ServiceRoles:    rest_model.Roles(in.ServiceRoles),
		EdgeRouterRoles: rest_model.Roles(in.EdgeRouterRoles),
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtSERP.NewUpdateServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(in.ID).WithPolicy(body)
	_, err = mgmt.ServiceEdgeRouterPolicy.UpdateServiceEdgeRouterPolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update service edge router policy %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}
