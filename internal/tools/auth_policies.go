package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtAP "github.com/openziti/edge-api/rest_management_api_client/auth_policy"
	"github.com/openziti/edge-api/rest_model"
)

func registerAuthPolicyTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &authPolicyTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-auth-policies",
		Description: "List authentication policies. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "auth policies", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtAP.NewListAuthPoliciesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.AuthPolicy.ListAuthPolicies(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-auth-policy",
		Description: "Get a single authentication policy by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "auth policy", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.AuthPolicy.DetailAuthPolicy(
			mgmtAP.NewDetailAuthPolicyParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-auth-policy",
		Description: "Create a new authentication policy controlling which authentication methods are allowed.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-auth-policy",
		Description: "Update an existing authentication policy.",
	}, t.update)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-auth-policy",
		Description: "Permanently delete an authentication policy by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "auth policy", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.AuthPolicy.DeleteAuthPolicy(
			mgmtAP.NewDeleteAuthPolicyParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type authPolicyTools struct{ zc *ziticlient.Client }

// authPolicyPrimaryInput captures primary auth method settings.
type authPolicyPrimaryInput struct {
	CertAllowed            bool  `json:"certAllowed"            jsonschema:"allow certificate authentication"`
	CertAllowExpiredCerts  bool  `json:"certAllowExpiredCerts"  jsonschema:"allow expired certificates"`
	UpdbAllowed            bool  `json:"updbAllowed"            jsonschema:"allow username/password authentication"`
	UpdbMinPasswordLength  int64 `json:"updbMinPasswordLength,omitempty" jsonschema:"minimum password length"`
	UpdbRequireMixedCase   bool  `json:"updbRequireMixedCase,omitempty"  jsonschema:"require mixed case passwords"`
	UpdbRequireNumberChar  bool  `json:"updbRequireNumberChar,omitempty" jsonschema:"require numeric character in passwords"`
	UpdbRequireSpecialChar bool  `json:"updbRequireSpecialChar,omitempty" jsonschema:"require special character in passwords"`
	UpdbMaxAttempts        int64 `json:"updbMaxAttempts,omitempty"       jsonschema:"max failed login attempts (0 = unlimited)"`
	UpdbLockoutMinutes     int64 `json:"updbLockoutMinutes,omitempty"    jsonschema:"lockout duration after max failed attempts"`
}

type createAuthPolicyInput struct {
	Name                 string                 `json:"name"                       jsonschema:"required,policy name"`
	Primary              authPolicyPrimaryInput `json:"primary"                    jsonschema:"primary authentication method settings"`
	SecondaryRequireTotp bool                   `json:"secondaryRequireTotp,omitempty" jsonschema:"require TOTP as secondary factor"`
	SecondaryExtJWTID    string                 `json:"secondaryExtJwtSignerId,omitempty" jsonschema:"require a specific ext-jwt signer as secondary factor"`
}

func buildAuthPolicyBody(name string, in authPolicyPrimaryInput, requireTotp bool, extJWTID string) *rest_model.AuthPolicyCreate {
	primary := &rest_model.AuthPolicyPrimary{
		Cert: &rest_model.AuthPolicyPrimaryCert{
			Allowed:           &in.CertAllowed,
			AllowExpiredCerts: &in.CertAllowExpiredCerts,
		},
		Updb: &rest_model.AuthPolicyPrimaryUpdb{
			Allowed:           &in.UpdbAllowed,
			RequireMixedCase:  &in.UpdbRequireMixedCase,
			RequireNumberChar: &in.UpdbRequireNumberChar,
			RequireSpecialChar: &in.UpdbRequireSpecialChar,
		},
	}
	if in.UpdbMinPasswordLength > 0 {
		primary.Updb.MinPasswordLength = &in.UpdbMinPasswordLength
	}
	if in.UpdbMaxAttempts > 0 {
		primary.Updb.MaxAttempts = &in.UpdbMaxAttempts
	}
	if in.UpdbLockoutMinutes > 0 {
		primary.Updb.LockoutDurationMinutes = &in.UpdbLockoutMinutes
	}

	secondary := &rest_model.AuthPolicySecondary{
		RequireTotp: &requireTotp,
	}
	if extJWTID != "" {
		secondary.RequireExtJWTSigner = &extJWTID
	}

	return &rest_model.AuthPolicyCreate{
		Name:      &name,
		Primary:   primary,
		Secondary: secondary,
	}
}

func (t *authPolicyTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createAuthPolicyInput) (*mcp.CallToolResult, any, error) {
	body := buildAuthPolicyBody(in.Name, in.Primary, in.SecondaryRequireTotp, in.SecondaryExtJWTID)

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAP.NewCreateAuthPolicyParams().WithContext(ctx).WithAuthPolicy(body)
	resp, err := mgmt.AuthPolicy.CreateAuthPolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create auth policy: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type updateAuthPolicyInput struct {
	ID                   string                 `json:"id"                         jsonschema:"required,auth policy ID to update"`
	Name                 string                 `json:"name"                       jsonschema:"required,policy name"`
	Primary              authPolicyPrimaryInput `json:"primary"                    jsonschema:"primary authentication method settings"`
	SecondaryRequireTotp bool                   `json:"secondaryRequireTotp,omitempty" jsonschema:"require TOTP as secondary factor"`
	SecondaryExtJWTID    string                 `json:"secondaryExtJwtSignerId,omitempty" jsonschema:"require a specific ext-jwt signer as secondary factor"`
}

func (t *authPolicyTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateAuthPolicyInput) (*mcp.CallToolResult, any, error) {
	createBody := buildAuthPolicyBody(in.Name, in.Primary, in.SecondaryRequireTotp, in.SecondaryExtJWTID)
	body := &rest_model.AuthPolicyUpdate{
		AuthPolicyCreate: *createBody,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtAP.NewUpdateAuthPolicyParams().WithContext(ctx).WithID(in.ID).WithAuthPolicy(body)
	_, err = mgmt.AuthPolicy.UpdateAuthPolicy(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update auth policy %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}
