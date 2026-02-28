package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtIdentity "github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_model"
)

func registerIdentityTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &identityTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-identities",
		Description: "List identities in the Ziti network. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-identity",
		Description: "Get a single identity by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-identity",
		Description: "Create a new identity. type must be one of: Device, User, Router, Service.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-identity",
		Description: "Update an existing identity's name, admin flag, or role attributes. All provided fields replace existing values.",
	}, t.update)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-identity",
		Description: "Permanently delete an identity by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type identityTools struct{ zc *ziticlient.Client }

type listIdentitiesInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression, e.g. name contains \"test\""`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *identityTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listIdentitiesInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtIdentity.NewListIdentitiesParams().WithContext(ctx)
	params.Limit = &limit
	params.Offset = &offset
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Identity.ListIdentities(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list identities: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type getIdentityInput struct {
	ID string `json:"id" jsonschema:"required,identity ID"`
}

func (t *identityTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getIdentityInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtIdentity.NewDetailIdentityParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Identity.DetailIdentity(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get identity %q: %w", in.ID, err)
	}
	return nil, resp.GetPayload().Data, nil
}

type createIdentityInput struct {
	Name           string   `json:"name"              jsonschema:"required,identity name"`
	Type           string   `json:"type"              jsonschema:"required,identity type: Device, User, Router, or Service"`
	IsAdmin        bool     `json:"isAdmin,omitempty" jsonschema:"whether the identity has admin privileges"`
	RoleAttributes []string `json:"roleAttributes,omitempty" jsonschema:"role attribute strings e.g. ['#servers','#db']"`
}

func (t *identityTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createIdentityInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.Type == "" {
		return nil, nil, fmt.Errorf("type is required (Device, User, Router, or Service)")
	}

	idType := rest_model.IdentityType(in.Type)
	body := &rest_model.IdentityCreate{
		Name:    &in.Name,
		Type:    &idType,
		IsAdmin: &in.IsAdmin,
	}
	if len(in.RoleAttributes) > 0 {
		attrs := rest_model.Attributes(in.RoleAttributes)
		body.RoleAttributes = &attrs
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtIdentity.NewCreateIdentityParams().WithContext(ctx).WithIdentity(body)
	resp, err := mgmt.Identity.CreateIdentity(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create identity: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type updateIdentityInput struct {
	ID             string   `json:"id"              jsonschema:"required,identity ID to update"`
	Name           string   `json:"name"            jsonschema:"required,new name for the identity"`
	Type           string   `json:"type"            jsonschema:"required,identity type: Device, User, Router, or Service"`
	IsAdmin        bool     `json:"isAdmin"         jsonschema:"whether the identity has admin privileges"`
	RoleAttributes []string `json:"roleAttributes,omitempty" jsonschema:"role attribute strings"`
}

func (t *identityTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateIdentityInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.Type == "" {
		return nil, nil, fmt.Errorf("type is required (Device, User, Router, or Service)")
	}

	idType := rest_model.IdentityType(in.Type)
	body := &rest_model.IdentityUpdate{
		Name:    &in.Name,
		Type:    &idType,
		IsAdmin: &in.IsAdmin,
	}
	if len(in.RoleAttributes) > 0 {
		attrs := rest_model.Attributes(in.RoleAttributes)
		body.RoleAttributes = &attrs
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtIdentity.NewUpdateIdentityParams().WithContext(ctx).WithID(in.ID).WithIdentity(body)
	_, err = mgmt.Identity.UpdateIdentity(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update identity %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "updated", "id": in.ID}, nil
}

type deleteIdentityInput struct {
	ID string `json:"id" jsonschema:"required,identity ID to delete"`
}

func (t *identityTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteIdentityInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtIdentity.NewDeleteIdentityParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Identity.DeleteIdentity(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete identity %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "deleted", "id": in.ID}, nil
}
