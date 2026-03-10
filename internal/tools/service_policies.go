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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-service-policies",
		Description: "List service policies. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "service policies", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtServicePolicy.NewListServicePoliciesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.ServicePolicy.ListServicePolicies(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-service-policy",
		Description: "Get a single service policy by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "service policy", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.ServicePolicy.DetailServicePolicy(
			mgmtServicePolicy.NewDetailServicePolicyParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

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
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "service policy", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.ServicePolicy.DeleteServicePolicy(
			mgmtServicePolicy.NewDeleteServicePolicyParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type servicePolicyTools struct{ zc *ziticlient.Client }

type createServicePolicyInput struct {
	Name          string   `json:"name"                      jsonschema:"required,policy name"`
	Type          string   `json:"type"                      jsonschema:"required,Dial or Bind"`
	Semantic      string   `json:"semantic"                  jsonschema:"required,AllOf or AnyOf"`
	ServiceRoles  []string `json:"serviceRoles,omitempty"    jsonschema:"service role selectors e.g. ['#tag','@name']"`
	IdentityRoles []string `json:"identityRoles,omitempty"   jsonschema:"identity role selectors e.g. ['#tag','@name']"`
}

func (t *servicePolicyTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createServicePolicyInput) (*mcp.CallToolResult, any, error) {
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
	return jsonResult(resp.GetPayload().Data)
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
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}
