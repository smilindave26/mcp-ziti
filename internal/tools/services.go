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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-services",
		Description: "List services in the Ziti network. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "services", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtService.NewListServicesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Service.ListServices(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-service",
		Description: "Get a single service by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "service", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Service.DetailService(
			mgmtService.NewDetailServiceParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

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
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "service", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.Service.DeleteService(
			mgmtService.NewDeleteServiceParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type serviceTools struct{ zc *ziticlient.Client }

type createServiceInput struct {
	Name               string   `json:"name"                        jsonschema:"required,service name"`
	EncryptionRequired bool     `json:"encryptionRequired"          jsonschema:"required,enforce end-to-end encryption on all connections"`
	RoleAttributes     []string `json:"roleAttributes,omitempty"    jsonschema:"role attribute strings"`
	Configs            []string `json:"configs,omitempty"            jsonschema:"list of config IDs to attach to this service"`
}

func (t *serviceTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createServiceInput) (*mcp.CallToolResult, any, error) {
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
