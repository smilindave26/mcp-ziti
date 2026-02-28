package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtCA "github.com/openziti/edge-api/rest_management_api_client/certificate_authority"
	"github.com/openziti/edge-api/rest_model"
)

func registerCertificateAuthorityTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &certificateAuthorityTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-certificate-authorities",
		Description: "List certificate authorities (CAs). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-certificate-authority",
		Description: "Get a single certificate authority by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-certificate-authority",
		Description: "Create a new certificate authority. certPem must be a PEM-encoded CA certificate.",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-certificate-authority",
		Description: "Update a certificate authority's name, enrollment settings, or role attributes.",
	}, t.update)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-certificate-authority",
		Description: "Permanently delete a certificate authority by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type certificateAuthorityTools struct{ zc *ziticlient.Client }

type listCAsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *certificateAuthorityTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listCAsInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtCA.NewListCasParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.CertificateAuthority.ListCas(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list certificate authorities: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type getCAInput struct {
	ID string `json:"id" jsonschema:"required,certificate authority ID"`
}

func (t *certificateAuthorityTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getCAInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtCA.NewDetailCaParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.CertificateAuthority.DetailCa(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get certificate authority %q: %w", in.ID, err)
	}
	return nil, resp.GetPayload().Data, nil
}

type createCAInput struct {
	Name                      string   `json:"name"                      jsonschema:"required,CA name"`
	CertPem                   string   `json:"certPem"                   jsonschema:"required,PEM-encoded CA certificate"`
	IsAuthEnabled             bool     `json:"isAuthEnabled"             jsonschema:"whether to allow authentication using this CA"`
	IsAutoCaEnrollmentEnabled bool     `json:"isAutoCaEnrollmentEnabled" jsonschema:"whether to auto-enroll identities via this CA"`
	IsOttCaEnrollmentEnabled  bool     `json:"isOttCaEnrollmentEnabled"  jsonschema:"whether to allow one-time-token enrollment via this CA"`
	IdentityRoles             []string `json:"identityRoles,omitempty"   jsonschema:"role attributes assigned to identities enrolled via this CA"`
	IdentityNameFormat        string   `json:"identityNameFormat,omitempty" jsonschema:"naming template for auto-enrolled identities"`
}

func (t *certificateAuthorityTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createCAInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.CertPem == "" {
		return nil, nil, fmt.Errorf("certPem is required")
	}

	body := &rest_model.CaCreate{
		Name:                      &in.Name,
		CertPem:                   &in.CertPem,
		IsAuthEnabled:             &in.IsAuthEnabled,
		IsAutoCaEnrollmentEnabled: &in.IsAutoCaEnrollmentEnabled,
		IsOttCaEnrollmentEnabled:  &in.IsOttCaEnrollmentEnabled,
		IdentityRoles:             rest_model.Roles(in.IdentityRoles),
		IdentityNameFormat:        in.IdentityNameFormat,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtCA.NewCreateCaParams().WithContext(ctx).WithCa(body)
	resp, err := mgmt.CertificateAuthority.CreateCa(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate authority: %w", err)
	}
	return nil, resp.GetPayload().Data, nil
}

type updateCAInput struct {
	ID                        string   `json:"id"                        jsonschema:"required,certificate authority ID to update"`
	Name                      string   `json:"name"                      jsonschema:"required,CA name"`
	IsAuthEnabled             bool     `json:"isAuthEnabled"             jsonschema:"whether to allow authentication using this CA"`
	IsAutoCaEnrollmentEnabled bool     `json:"isAutoCaEnrollmentEnabled" jsonschema:"whether to auto-enroll identities via this CA"`
	IsOttCaEnrollmentEnabled  bool     `json:"isOttCaEnrollmentEnabled"  jsonschema:"whether to allow one-time-token enrollment via this CA"`
	IdentityRoles             []string `json:"identityRoles,omitempty"   jsonschema:"role attributes assigned to auto-enrolled identities"`
	IdentityNameFormat        string   `json:"identityNameFormat,omitempty" jsonschema:"naming template for auto-enrolled identities"`
}

func (t *certificateAuthorityTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateCAInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	body := &rest_model.CaUpdate{
		Name:                      &in.Name,
		IsAuthEnabled:             &in.IsAuthEnabled,
		IsAutoCaEnrollmentEnabled: &in.IsAutoCaEnrollmentEnabled,
		IsOttCaEnrollmentEnabled:  &in.IsOttCaEnrollmentEnabled,
		IdentityRoles:             rest_model.Roles(in.IdentityRoles),
		IdentityNameFormat:        &in.IdentityNameFormat,
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtCA.NewUpdateCaParams().WithContext(ctx).WithID(in.ID).WithCa(body)
	_, err = mgmt.CertificateAuthority.UpdateCa(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update certificate authority %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "updated", "id": in.ID}, nil
}

type deleteCAInput struct {
	ID string `json:"id" jsonschema:"required,certificate authority ID to delete"`
}

func (t *certificateAuthorityTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteCAInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtCA.NewDeleteCaParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.CertificateAuthority.DeleteCa(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete certificate authority %q: %w", in.ID, err)
	}
	return nil, map[string]string{"status": "deleted", "id": in.ID}, nil
}
