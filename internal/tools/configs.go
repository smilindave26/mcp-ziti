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
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.listTypes)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-config-type",
		Description: "Get a single config type by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.getType)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-config-type",
		Description: "Create a new config type. schema is an optional JSON Schema object.",
	}, t.createType)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-config-type",
		Description: "Permanently delete a config type by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: func() *bool { b := true; return &b }()},
	}, t.deleteType)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-configs",
		Description: "List service configurations. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-config",
		Description: "Get a single service configuration by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-config",
		Description: "Create a new service configuration. data is a JSON object conforming to the config type's schema.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-config",
		Description: "Update an existing service configuration's name or data.",
	}, t.update)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-config",
		Description: "Permanently delete a service configuration by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type configTools struct{ zc *ziticlient.Client }

// --- Config Types ---

type listConfigTypesInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *configTools) listTypes(ctx context.Context, _ *mcp.CallToolRequest, in listConfigTypesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewListConfigTypesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Config.ListConfigTypes(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list config types: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getConfigTypeInput struct {
	ID string `json:"id" jsonschema:"required,config type ID"`
}

func (t *configTools) getType(ctx context.Context, _ *mcp.CallToolRequest, in getConfigTypeInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewDetailConfigTypeParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Config.DetailConfigType(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get config type %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type createConfigTypeInput struct {
	Name   string `json:"name"             jsonschema:"required,config type name"`
	Schema any    `json:"schema,omitempty" jsonschema:"optional JSON Schema object for validating config data"`
}

func (t *configTools) createType(ctx context.Context, _ *mcp.CallToolRequest, in createConfigTypeInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

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

type deleteConfigTypeInput struct {
	ID string `json:"id" jsonschema:"required,config type ID to delete"`
}

func (t *configTools) deleteType(ctx context.Context, _ *mcp.CallToolRequest, in deleteConfigTypeInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewDeleteConfigTypeParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Config.DeleteConfigType(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete config type %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}

// --- Configs ---

type listConfigsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *configTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listConfigsInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewListConfigsParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Config.ListConfigs(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list configs: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getConfigInput struct {
	ID string `json:"id" jsonschema:"required,config ID"`
}

func (t *configTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getConfigInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewDetailConfigParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Config.DetailConfig(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get config %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type createConfigInput struct {
	Name         string         `json:"name"         jsonschema:"required,config name"`
	ConfigTypeID string         `json:"configTypeId" jsonschema:"required,ID of the config type this config conforms to"`
	Data         map[string]any `json:"data"         jsonschema:"required,JSON object containing the config data"`
}

func (t *configTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createConfigInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.ConfigTypeID == "" {
		return nil, nil, fmt.Errorf("configTypeId is required")
	}
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
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
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

type deleteConfigInput struct {
	ID string `json:"id" jsonschema:"required,config ID to delete"`
}

func (t *configTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteConfigInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtConfig.NewDeleteConfigParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Config.DeleteConfig(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete config %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
