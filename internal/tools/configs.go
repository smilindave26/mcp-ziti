package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtConfig "github.com/openziti/edge-api/rest_management_api_client/config"
	"github.com/openziti/edge-api/rest_model"
)

func registerConfigTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &configTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-config-types",
		Description: "List service config types (schemas for service configurations). Returns up to `limit` results (default 100, max 500).",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "config types", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtConfig.NewListConfigTypesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Config.ListConfigTypes(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-config-type",
		Description: "Get a single config type by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "config type", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Config.DetailConfigType(
			mgmtConfig.NewDetailConfigTypeParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-config-type",
		Description: "Create a new config type. schema is an optional JSON Schema object.",
	}, t.createType)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-config-type",
		Description: "Permanently delete a config type by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "config type", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.Config.DeleteConfigType(
			mgmtConfig.NewDeleteConfigTypeParams().WithContext(ctx).WithID(id), nil)
		return err
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-configs",
		Description: "List service configurations. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "configs", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtConfig.NewListConfigsParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Config.ListConfigs(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-config",
		Description: "Get a single service configuration by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "config", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Config.DetailConfig(
			mgmtConfig.NewDetailConfigParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-config",
		Description: "Create a new service configuration. data is a JSON object conforming to the config type's schema.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-config",
		Description: "Update an existing service configuration's name or data.",
	}, t.update)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-config",
		Description: "Permanently delete a service configuration by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "config", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.Config.DeleteConfig(
			mgmtConfig.NewDeleteConfigParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type configTools struct{ zc *ziticlient.Client }

// --- Config Types ---

type createConfigTypeInput struct {
	Name   string `json:"name"             jsonschema:"required,config type name"`
	Schema any    `json:"schema,omitempty" jsonschema:"optional JSON Schema object for validating config data"`
}

func (t *configTools) createType(ctx context.Context, _ *mcp.CallToolRequest, in createConfigTypeInput) (*mcp.CallToolResult, any, error) {
	body := &rest_model.ConfigTypeCreate{
		Name:   &in.Name,
		Schema: in.Schema,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewCreateConfigTypeParams().WithContext(ctx).WithConfigType(body)
	resp, err := mgmt.Config.CreateConfigType(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create config type: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

// --- Configs ---

type createConfigInput struct {
	Name         string         `json:"name"         jsonschema:"required,config name"`
	ConfigTypeID string         `json:"configTypeId" jsonschema:"required,ID of the config type this config conforms to"`
	Data         map[string]any `json:"data"         jsonschema:"required,JSON object containing the config data"`
}

func (t *configTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createConfigInput) (*mcp.CallToolResult, any, error) {
	if in.Data == nil {
		return nil, nil, fmt.Errorf("data is required")
	}

	var data any = in.Data
	body := &rest_model.ConfigCreate{
		Name:         &in.Name,
		ConfigTypeID: &in.ConfigTypeID,
		Data:         &data,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewCreateConfigParams().WithContext(ctx).WithConfig(body)
	resp, err := mgmt.Config.CreateConfig(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create config: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type updateConfigInput struct {
	ID   string         `json:"id"   jsonschema:"required,config ID to update"`
	Name string         `json:"name" jsonschema:"required,config name"`
	Data map[string]any `json:"data" jsonschema:"required,JSON object containing the updated config data"`
}

func (t *configTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateConfigInput) (*mcp.CallToolResult, any, error) {
	if in.Data == nil {
		return nil, nil, fmt.Errorf("data is required")
	}

	var updateData any = in.Data
	body := &rest_model.ConfigUpdate{
		Name: &in.Name,
		Data: &updateData,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewUpdateConfigParams().WithContext(ctx).WithID(in.ID).WithConfig(body)
	_, err = mgmt.Config.UpdateConfig(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update config %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}
