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
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "identities", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtIdentity.NewListIdentitiesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Identity.ListIdentities(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-identity",
		Description: "Get a single identity by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "identity", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Identity.DetailIdentity(
			mgmtIdentity.NewDetailIdentityParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-identity",
		Description: "Create a new identity. type must be one of: Device, User, Router, Service.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-identity",
		Description: "Update an existing identity's name, admin flag, or role attributes. All provided fields replace existing values.",
	}, t.update)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-identity",
		Description: "Permanently delete an identity by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "identity", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.Identity.DeleteIdentity(
			mgmtIdentity.NewDeleteIdentityParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type identityTools struct{ zc *ziticlient.Client }

type createIdentityInput struct {
	Name           string   `json:"name"              jsonschema:"required,identity name"`
	Type           string   `json:"type"              jsonschema:"required,identity type: Device, User, Router, or Service"`
	IsAdmin        bool     `json:"isAdmin,omitempty" jsonschema:"whether the identity has admin privileges"`
	RoleAttributes []string `json:"roleAttributes,omitempty" jsonschema:"role attribute strings e.g. ['#servers','#db']"`
}

func (t *identityTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createIdentityInput) (*mcp.CallToolResult, any, error) {
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
	return jsonResult(resp.GetPayload().Data)
}

type updateIdentityInput struct {
	ID             string   `json:"id"              jsonschema:"required,identity ID to update"`
	Name           string   `json:"name"            jsonschema:"required,new name for the identity"`
	Type           string   `json:"type"            jsonschema:"required,identity type: Device, User, Router, or Service"`
	IsAdmin        bool     `json:"isAdmin"         jsonschema:"whether the identity has admin privileges"`
	RoleAttributes []string `json:"roleAttributes,omitempty" jsonschema:"role attribute strings"`
}

func (t *identityTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateIdentityInput) (*mcp.CallToolResult, any, error) {
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
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}
