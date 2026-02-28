package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtServicePolicy "github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
)

func registerServicePolicyTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &servicePolicyTools{zc: zc}
	destructive := true

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-service-policies",
		Description: "List service policies. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-service-policy",
		Description: "Get a single service policy by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-service-policy",
		Description: "Create a service policy linking identities to services. type must be Dial or Bind. semantic must be AllOf or AnyOf. Roles use #tag or @name syntax.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-service-policy",
		Description: "Update a service policy's name, type, semantic, or role assignments.",
	}, t.update)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-service-policy",
		Description: "Permanently delete a service policy by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type servicePolicyTools struct{ zc *ziticlient.Client }

type listServicePoliciesInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *servicePolicyTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listServicePoliciesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtServicePolicy.NewListServicePoliciesParams().WithContext(ctx)
	params.Limit = &limit
	params.Offset = &offset
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.ServicePolicy.ListServicePolicies(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list service policies: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type getServicePolicyInput struct {
	ID string `json:"id" jsonschema:"required,service policy ID"`
}

func (t *servicePolicyTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getServicePolicyInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtServicePolicy.NewDetailServicePolicyParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.ServicePolicy.DetailServicePolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get service policy %q: %w", in.ID, err)
	}
	return nil, resp.GetPayload().Data, nil
}

type createServicePolicyInput struct {
	Name          string   `json:"name"                      jsonschema:"required,policy name"`
	Type          string   `json:"type"                      jsonschema:"required,Dial or Bind"`
	Semantic      string   `json:"semantic"                  jsonschema:"required,AllOf or AnyOf"`
	ServiceRoles  []string `json:"serviceRoles,omitempty"    jsonschema:"service role selectors e.g. ['#tag','@name']"`
	IdentityRoles []string `json:"identityRoles,omitempty"   jsonschema:"identity role selectors e.g. ['#tag','@name']"`
}

func (t *servicePolicyTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createServicePolicyInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.Type == "" {
		return nil, nil, fmt.Errorf("type is required (Dial or Bind)")
	}
	if in.Semantic == "" {
		return nil, nil, fmt.Errorf("semantic is required (AllOf or AnyOf)")
	}

	dialBind := rest_model.DialBind(in.Type)
	semantic := rest_model.Semantic(in.Semantic)
	body := &rest_model.ServicePolicyCreate{
		Name:          &in.Name,
		Type:          &dialBind,
		Semantic:      &semantic,
		ServiceRoles:  rest_model.Roles(in.ServiceRoles),
		IdentityRoles: rest_model.Roles(in.IdentityRoles),
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtServicePolicy.NewCreateServicePolicyParams().WithContext(ctx).WithPolicy(body)
	resp, err := mgmt.ServicePolicy.CreateServicePolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create service policy: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type updateServicePolicyInput struct {
	ID            string   `json:"id"                        jsonschema:"required,service policy ID to update"`
	Name          string   `json:"name"                      jsonschema:"required,new policy name"`
	Type          string   `json:"type"                      jsonschema:"required,Dial or Bind"`
	Semantic      string   `json:"semantic"                  jsonschema:"required,AllOf or AnyOf"`
	ServiceRoles  []string `json:"serviceRoles,omitempty"    jsonschema:"service role selectors"`
	IdentityRoles []string `json:"identityRoles,omitempty"   jsonschema:"identity role selectors"`
}

func (t *servicePolicyTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateServicePolicyInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.Type == "" {
		return nil, nil, fmt.Errorf("type is required (Dial or Bind)")
	}
	if in.Semantic == "" {
		return nil, nil, fmt.Errorf("semantic is required (AllOf or AnyOf)")
	}

	dialBind := rest_model.DialBind(in.Type)
	semantic := rest_model.Semantic(in.Semantic)
	body := &rest_model.ServicePolicyUpdate{
		Name:          &in.Name,
		Type:          &dialBind,
		Semantic:      &semantic,
		ServiceRoles:  rest_model.Roles(in.ServiceRoles),
		IdentityRoles: rest_model.Roles(in.IdentityRoles),
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtServicePolicy.NewUpdateServicePolicyParams().WithContext(ctx).WithID(in.ID).WithPolicy(body)
	_, err = mgmt.ServicePolicy.UpdateServicePolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update service policy %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "updated", "id": in.ID}, nil
}

type deleteServicePolicyInput struct {
	ID string `json:"id" jsonschema:"required,service policy ID to delete"`
}

func (t *servicePolicyTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteServicePolicyInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtServicePolicy.NewDeleteServicePolicyParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.ServicePolicy.DeleteServicePolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete service policy %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "deleted", "id": in.ID}, nil
}
