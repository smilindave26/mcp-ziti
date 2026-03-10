package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtAuth "github.com/openziti/edge-api/rest_management_api_client/authenticator"
	"github.com/openziti/edge-api/rest_model"
)

func registerAuthenticatorTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &authenticatorTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-authenticators",
		Description: "List authenticators (credentials attached to identities). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "authenticators", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtAuth.NewListAuthenticatorsParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Authenticator.ListAuthenticators(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-authenticator",
		Description: "Get a single authenticator by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "authenticator", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Authenticator.DetailAuthenticator(
			mgmtAuth.NewDetailAuthenticatorParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-authenticator",
		Description: "Update the username/password for an updb authenticator.",
	}, t.update)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-authenticator",
		Description: "Permanently delete an authenticator by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "authenticator", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.Authenticator.DeleteAuthenticator(
			mgmtAuth.NewDeleteAuthenticatorParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type authenticatorTools struct{ zc *ziticlient.Client }

type updateAuthenticatorInput struct {
	ID       string `json:"id"       jsonschema:"required,authenticator ID to update"`
	Username string `json:"username" jsonschema:"required,new username"`
	Password string `json:"password" jsonschema:"required,new password"`
}

func (t *authenticatorTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateAuthenticatorInput) (*mcp.CallToolResult, any, error) {
	username := rest_model.Username(in.Username)
	password := rest_model.Password(in.Password)
	body := &rest_model.AuthenticatorUpdate{
		Username: &username,
		Password: &password,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAuth.NewUpdateAuthenticatorParams().WithContext(ctx).WithID(in.ID).WithAuthenticator(body)
	_, err = mgmt.Authenticator.UpdateAuthenticator(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update authenticator %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}
