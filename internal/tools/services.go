package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtService "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_model"
)

func registerServiceTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &serviceTools{zc: zc}
	destructive := true

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-services",
		Description: "List services in the Ziti network. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-service",
		Description: "Get a single service by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-service",
		Description: "Create a new Ziti service. encryptionRequired enforces end-to-end encryption on all connections.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-service",
		Description: "Update an existing service's name, encryption setting, or role attributes.",
	}, t.update)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-service",
		Description: "Permanently delete a service by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type serviceTools struct{ zc *ziticlient.Client }

type listServicesInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *serviceTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listServicesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtService.NewListServicesParams().WithContext(ctx)
	params.Limit = &limit
	params.Offset = &offset
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Service.ListServices(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list services: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getServiceInput struct {
	ID string `json:"id" jsonschema:"required,service ID"`
}

func (t *serviceTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getServiceInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtService.NewDetailServiceParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Service.DetailService(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get service %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type createServiceInput struct {
	Name               string   `json:"name"                        jsonschema:"required,service name"`
	EncryptionRequired bool     `json:"encryptionRequired"          jsonschema:"required,enforce end-to-end encryption on all connections"`
	RoleAttributes     []string `json:"roleAttributes,omitempty"    jsonschema:"role attribute strings"`
	Configs            []string `json:"configs,omitempty"            jsonschema:"list of config IDs to attach to this service"`
}

func (t *serviceTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createServiceInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	body := &rest_model.ServiceCreate{
		Name:               &in.Name,
		EncryptionRequired: &in.EncryptionRequired,
		RoleAttributes:     in.RoleAttributes,
		Configs:            in.Configs,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtService.NewCreateServiceParams().WithContext(ctx).WithService(body)
	resp, err := mgmt.Service.CreateService(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create service: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type updateServiceInput struct {
	ID                 string   `json:"id"                          jsonschema:"required,service ID to update"`
	Name               string   `json:"name"                        jsonschema:"required,new service name"`
	EncryptionRequired bool     `json:"encryptionRequired"          jsonschema:"whether to enforce end-to-end encryption"`
	RoleAttributes     []string `json:"roleAttributes,omitempty"    jsonschema:"role attribute strings"`
	Configs            []string `json:"configs,omitempty"            jsonschema:"list of config IDs to attach to this service"`
}

func (t *serviceTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateServiceInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	body := &rest_model.ServiceUpdate{
		Name:               &in.Name,
		EncryptionRequired: in.EncryptionRequired,
		RoleAttributes:     in.RoleAttributes,
		Configs:            in.Configs,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtService.NewUpdateServiceParams().WithContext(ctx).WithID(in.ID).WithService(body)
	_, err = mgmt.Service.UpdateService(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update service %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}

type deleteServiceInput struct {
	ID string `json:"id" jsonschema:"required,service ID to delete"`
}

func (t *serviceTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteServiceInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtService.NewDeleteServiceParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Service.DeleteService(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete service %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
