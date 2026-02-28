package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

func registerConnectionTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &connectionTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "connect-controller",
		Description: "Connect (or reconnect) to a Ziti controller. Provide exactly one authentication method: identity JSON, username/password, client certificate, external JWT, or OIDC client credentials.",
	}, t.connect)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "disconnect-controller",
		Description: "Disconnect from the current Ziti controller, clearing all credentials and session state.",
	}, t.disconnect)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-controller-status",
		Description: "Get the current connection status, controller URL, and API version compatibility info.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.status)
}

type connectionTools struct{ zc *ziticlient.Client }

type connectControllerInput struct {
	ControllerURL string `json:"controllerUrl,omitempty" jsonschema:"controller URL, e.g. https://ctrl.example.com:1280 — required unless using identityJson"`

	// Identity file auth (inline JSON content)
	IdentityJSON string `json:"identityJson,omitempty" jsonschema:"inline Ziti identity JSON content"`

	// Username/password auth
	Username string `json:"username,omitempty" jsonschema:"username for updb authentication"`
	Password string `json:"password,omitempty" jsonschema:"password for updb authentication"`

	// Client certificate auth
	CertPEM string `json:"certPem,omitempty" jsonschema:"client certificate PEM content (path to file)"`
	KeyPEM  string `json:"keyPem,omitempty"  jsonschema:"client private key PEM content (path to file)"`

	// External JWT auth
	ExtJWTToken string `json:"extJwtToken,omitempty" jsonschema:"external JWT token string"`

	// OIDC client credentials auth
	OIDCIssuer       string `json:"oidcIssuer,omitempty"       jsonschema:"OIDC issuer URL"`
	OIDCClientID     string `json:"oidcClientId,omitempty"     jsonschema:"OIDC client ID"`
	OIDCClientSecret string `json:"oidcClientSecret,omitempty" jsonschema:"OIDC client secret"`
	OIDCAudience     string `json:"oidcAudience,omitempty"     jsonschema:"OIDC audience (optional)"`
	OIDCTokenURL     string `json:"oidcTokenUrl,omitempty"     jsonschema:"OIDC token endpoint URL — optional, skips discovery"`

	// Optional CA override
	CAFile string `json:"caFile,omitempty" jsonschema:"path to CA bundle PEM file (optional)"`
}

func (t *connectionTools) connect(_ context.Context, _ *mcp.CallToolRequest, in connectControllerInput) (*mcp.CallToolResult, any, error) {
	cfg := &config.Config{
		ControllerURL:    in.ControllerURL,
		IdentityFile:     in.IdentityJSON,
		Username:         in.Username,
		Password:         in.Password,
		CertFile:         in.CertPEM,
		KeyFile:          in.KeyPEM,
		ExtJWTToken:      in.ExtJWTToken,
		OIDCIssuer:       in.OIDCIssuer,
		OIDCClientID:     in.OIDCClientID,
		OIDCClientSecret: in.OIDCClientSecret,
		OIDCAudience:     in.OIDCAudience,
		OIDCTokenURL:     in.OIDCTokenURL,
		CAFile:           in.CAFile,
	}

	if err := t.zc.Connect(cfg); err != nil {
		return nil, nil, err
	}

	return t.statusResult()
}

type disconnectControllerInput struct{}

func (t *connectionTools) disconnect(_ context.Context, _ *mcp.CallToolRequest, _ disconnectControllerInput) (*mcp.CallToolResult, any, error) {
	if err := t.zc.Disconnect(); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"connected": false,
	})
}

type getControllerStatusInput struct{}

func (t *connectionTools) status(_ context.Context, _ *mcp.CallToolRequest, _ getControllerStatusInput) (*mcp.CallToolResult, any, error) {
	return t.statusResult()
}

// statusResult builds a tool result with a human-readable compatibility summary
// as a separate text block (so the LLM relays it to the user) followed by the
// full JSON details.
func (t *connectionTools) statusResult() (*mcp.CallToolResult, any, error) {
	data := map[string]any{
		"connected":     t.zc.Connected(),
		"controllerUrl": t.zc.ControllerURL(),
	}

	var summary string
	if info := t.zc.GetVersionInfo(); info != nil {
		data["controllerVersion"] = info.ControllerVersion
		data["buildDate"] = info.BuildDate
		data["runtimeVersion"] = info.RuntimeVersion
		data["controllerAPIVersions"] = info.APIVersions
		data["thisToolBuiltFor"] = info.ThisToolBuiltFor
		data["edgeApiModule"] = info.EdgeAPIModule
		data["compatible"] = info.Compatible
		data["compatibilityNote"] = info.CompatibilityNote

		if info.Compatible {
			summary = fmt.Sprintf(
				"Connected to controller %s at %s.\n\n"+
					"API Compatibility: COMPATIBLE\n"+
					"  This tool uses: %s (edge-api %s)\n"+
					"  Controller supports: %s\n\n"+
					"Note: %s",
				info.ControllerVersion, t.zc.ControllerURL(),
				info.ThisToolBuiltFor, info.EdgeAPIModule,
				info.ThisToolBuiltFor,
				info.CompatibilityNote)
		} else {
			summary = fmt.Sprintf(
				"Connected to controller %s at %s.\n\n"+
					"API Compatibility: NOT COMPATIBLE — OPERATIONS MAY FAIL\n"+
					"  This tool uses: %s (edge-api %s)\n"+
					"  Controller does NOT advertise %s\n\n"+
					"Warning: %s",
				info.ControllerVersion, t.zc.ControllerURL(),
				info.ThisToolBuiltFor, info.EdgeAPIModule,
				info.ThisToolBuiltFor,
				info.CompatibilityNote)
		}
	} else if !t.zc.Connected() {
		summary = "Not connected to a Ziti controller. Use connect-controller to connect."
	} else {
		summary = fmt.Sprintf("Connected to %s. Version info unavailable.", t.zc.ControllerURL())
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil, nil
}
