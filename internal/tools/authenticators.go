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
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-authenticator",
		Description: "Get a single authenticator by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-authenticator",
		Description: "Update the username/password for an updb authenticator.",
	}, t.update)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-authenticator",
		Description: "Permanently delete an authenticator by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type authenticatorTools struct{ zc *ziticlient.Client }

type listAuthenticatorsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *authenticatorTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listAuthenticatorsInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAuth.NewListAuthenticatorsParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Authenticator.ListAuthenticators(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list authenticators: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getAuthenticatorInput struct {
	ID string `json:"id" jsonschema:"required,authenticator ID"`
}

func (t *authenticatorTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getAuthenticatorInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAuth.NewDetailAuthenticatorParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Authenticator.DetailAuthenticator(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get authenticator %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type updateAuthenticatorInput struct {
	ID       string `json:"id"       jsonschema:"required,authenticator ID to update"`
	Username string `json:"username" jsonschema:"required,new username"`
	Password string `json:"password" jsonschema:"required,new password"`
}

func (t *authenticatorTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateAuthenticatorInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Username == "" {
		return nil, nil, fmt.Errorf("username is required")
	}
	if in.Password == "" {
		return nil, nil, fmt.Errorf("password is required")
	}

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

type deleteAuthenticatorInput struct {
	ID string `json:"id" jsonschema:"required,authenticator ID to delete"`
}

func (t *authenticatorTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteAuthenticatorInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAuth.NewDeleteAuthenticatorParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Authenticator.DeleteAuthenticator(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete authenticator %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
