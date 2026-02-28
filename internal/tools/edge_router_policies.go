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
	destructive := true

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-edge-router-policies",
		Description: "List edge router policies. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-edge-router-policy",
		Description: "Get a single edge router policy by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-edge-router-policy",
		Description: "Create an edge router policy linking identities to edge routers. semantic must be AllOf or AnyOf. Roles use #tag or @name syntax.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-edge-router-policy",
		Description: "Permanently delete an edge router policy by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type edgeRouterPolicyTools struct{ zc *ziticlient.Client }

type listEdgeRouterPoliciesInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *edgeRouterPolicyTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listEdgeRouterPoliciesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtERP.NewListEdgeRouterPoliciesParams().WithContext(ctx)
	params.Limit = &limit
	params.Offset = &offset
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.EdgeRouterPolicy.ListEdgeRouterPolicies(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list edge router policies: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getEdgeRouterPolicyInput struct {
	ID string `json:"id" jsonschema:"required,edge router policy ID"`
}

func (t *edgeRouterPolicyTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getEdgeRouterPolicyInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtERP.NewDetailEdgeRouterPolicyParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.EdgeRouterPolicy.DetailEdgeRouterPolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get edge router policy %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type createEdgeRouterPolicyInput struct {
	Name            string   `json:"name"                       jsonschema:"required,policy name"`
	Semantic        string   `json:"semantic"                   jsonschema:"required,AllOf or AnyOf"`
	EdgeRouterRoles []string `json:"edgeRouterRoles,omitempty"  jsonschema:"edge router role selectors e.g. ['#tag','@name']"`
	IdentityRoles   []string `json:"identityRoles,omitempty"    jsonschema:"identity role selectors e.g. ['#tag','@name']"`
}

func (t *edgeRouterPolicyTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createEdgeRouterPolicyInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.Semantic == "" {
		return nil, nil, fmt.Errorf("semantic is required (AllOf or AnyOf)")
	}

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

type deleteEdgeRouterPolicyInput struct {
	ID string `json:"id" jsonschema:"required,edge router policy ID to delete"`
}

func (t *edgeRouterPolicyTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteEdgeRouterPolicyInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtERP.NewDeleteEdgeRouterPolicyParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.EdgeRouterPolicy.DeleteEdgeRouterPolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete edge router policy %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
