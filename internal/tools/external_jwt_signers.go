package tools

import (
	"context"
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtEJS "github.com/openziti/edge-api/rest_management_api_client/external_jwt_signer"
	"github.com/openziti/edge-api/rest_model"
)

func registerExternalJWTSignerTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &externalJWTSignerTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-external-jwt-signers",
		Description: "List external JWT signers. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-external-jwt-signer",
		Description: "Get a single external JWT signer by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-external-jwt-signer",
		Description: "Create a new external JWT signer. Provide either certPem (for static cert validation) or jwksEndpoint (for JWKS-based validation).",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-external-jwt-signer",
		Description: "Update an external JWT signer.",
	}, t.update)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-external-jwt-signer",
		Description: "Permanently delete an external JWT signer by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type externalJWTSignerTools struct{ zc *ziticlient.Client }

type listEJSInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *externalJWTSignerTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listEJSInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEJS.NewListExternalJWTSignersParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.ExternalJWTSigner.ListExternalJWTSigners(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list external JWT signers: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getEJSInput struct {
	ID string `json:"id" jsonschema:"required,external JWT signer ID"`
}

func (t *externalJWTSignerTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getEJSInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEJS.NewDetailExternalJWTSignerParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.ExternalJWTSigner.DetailExternalJWTSigner(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get external JWT signer %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type createEJSInput struct {
	Name          string `json:"name"                    jsonschema:"required,signer name"`
	Issuer        string `json:"issuer"                  jsonschema:"required,expected JWT issuer claim"`
	Audience      string `json:"audience"                jsonschema:"required,expected JWT audience claim"`
	Enabled       bool   `json:"enabled"                 jsonschema:"whether the signer is active"`
	CertPem       string `json:"certPem,omitempty"       jsonschema:"PEM-encoded signing certificate (mutually exclusive with jwksEndpoint)"`
	JwksEndpoint  string `json:"jwksEndpoint,omitempty"  jsonschema:"JWKS endpoint URL (mutually exclusive with certPem)"`
	Kid           string `json:"kid,omitempty"           jsonschema:"key ID hint"`
	ClaimsProperty string `json:"claimsProperty,omitempty" jsonschema:"JWT claim to use as the identity lookup property"`
	UseExternalID bool   `json:"useExternalId,omitempty" jsonschema:"use externalId field for identity lookup"`
	ClientID      string `json:"clientId,omitempty"      jsonschema:"OAuth client ID for external auth URL flow"`
	ExternalAuthURL string `json:"externalAuthUrl,omitempty" jsonschema:"URL for external authentication"`
}

func (t *externalJWTSignerTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createEJSInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.Issuer == "" {
		return nil, nil, fmt.Errorf("issuer is required")
	}
	if in.Audience == "" {
		return nil, nil, fmt.Errorf("audience is required")
	}
	if in.CertPem == "" && in.JwksEndpoint == "" {
		return nil, nil, fmt.Errorf("one of certPem or jwksEndpoint is required")
	}

	body := &rest_model.ExternalJWTSignerCreate{
		Name:     &in.Name,
		Issuer:   &in.Issuer,
		Audience: &in.Audience,
		Enabled:  &in.Enabled,
	}
	if in.CertPem != "" {
		body.CertPem = &in.CertPem
	}
	if in.JwksEndpoint != "" {
		jwks := strfmt.URI(in.JwksEndpoint)
		body.JwksEndpoint = &jwks
	}
	if in.Kid != "" {
		body.Kid = &in.Kid
	}
	if in.ClaimsProperty != "" {
		body.ClaimsProperty = &in.ClaimsProperty
	}
	if in.UseExternalID {
		body.UseExternalID = &in.UseExternalID
	}
	if in.ClientID != "" {
		body.ClientID = &in.ClientID
	}
	if in.ExternalAuthURL != "" {
		body.ExternalAuthURL = &in.ExternalAuthURL
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEJS.NewCreateExternalJWTSignerParams().WithContext(ctx).WithExternalJWTSigner(body)
	resp, err := mgmt.ExternalJWTSigner.CreateExternalJWTSigner(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create external JWT signer: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type updateEJSInput struct {
	ID            string `json:"id"                      jsonschema:"required,external JWT signer ID to update"`
	Name          string `json:"name"                    jsonschema:"required,signer name"`
	Issuer        string `json:"issuer"                  jsonschema:"required,expected JWT issuer claim"`
	Audience      string `json:"audience"                jsonschema:"required,expected JWT audience claim"`
	Enabled       bool   `json:"enabled"                 jsonschema:"whether the signer is active"`
	CertPem       string `json:"certPem,omitempty"       jsonschema:"PEM-encoded signing certificate"`
	JwksEndpoint  string `json:"jwksEndpoint,omitempty"  jsonschema:"JWKS endpoint URL"`
	Kid           string `json:"kid,omitempty"           jsonschema:"key ID hint"`
	ClaimsProperty string `json:"claimsProperty,omitempty" jsonschema:"JWT claim used for identity lookup"`
	UseExternalID bool   `json:"useExternalId,omitempty" jsonschema:"use externalId field for identity lookup"`
	ClientID      string `json:"clientId,omitempty"      jsonschema:"OAuth client ID"`
	ExternalAuthURL string `json:"externalAuthUrl,omitempty" jsonschema:"URL for external authentication"`
}

func (t *externalJWTSignerTools) update(ctx context.Context, _ *mcp.CallToolRequest, in updateEJSInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if in.Issuer == "" {
		return nil, nil, fmt.Errorf("issuer is required")
	}
	if in.Audience == "" {
		return nil, nil, fmt.Errorf("audience is required")
	}

	body := &rest_model.ExternalJWTSignerUpdate{
		Name:     &in.Name,
		Issuer:   &in.Issuer,
		Audience: &in.Audience,
		Enabled:  &in.Enabled,
	}
	if in.CertPem != "" {
		body.CertPem = &in.CertPem
	}
	if in.JwksEndpoint != "" {
		jwks := strfmt.URI(in.JwksEndpoint)
		body.JwksEndpoint = &jwks
	}
	if in.Kid != "" {
		body.Kid = &in.Kid
	}
	if in.ClaimsProperty != "" {
		body.ClaimsProperty = &in.ClaimsProperty
	}
	if in.UseExternalID {
		body.UseExternalID = &in.UseExternalID
	}
	if in.ClientID != "" {
		body.ClientID = &in.ClientID
	}
	if in.ExternalAuthURL != "" {
		body.ExternalAuthURL = &in.ExternalAuthURL
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEJS.NewUpdateExternalJWTSignerParams().WithContext(ctx).WithID(in.ID).WithExternalJWTSigner(body)
	_, err = mgmt.ExternalJWTSigner.UpdateExternalJWTSigner(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("update external JWT signer %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "updated", "id": in.ID})
}

type deleteEJSInput struct {
	ID string `json:"id" jsonschema:"required,external JWT signer ID to delete"`
}

func (t *externalJWTSignerTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteEJSInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEJS.NewDeleteExternalJWTSignerParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.ExternalJWTSigner.DeleteExternalJWTSigner(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete external JWT signer %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
