package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtInfo "github.com/openziti/edge-api/rest_management_api_client/informational"
)

// requiredAPIPath is the management API path this tool was built against.
const requiredAPIPath = "/edge/management/v1"

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
		Description: "Get the current connection status and controller URL.",
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

func (t *connectionTools) connect(ctx context.Context, _ *mcp.CallToolRequest, in connectControllerInput) (*mcp.CallToolResult, any, error) {
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

	result := map[string]any{
		"connected":     true,
		"controllerUrl": t.zc.ControllerURL(),
	}

	// Fetch version info from the controller and assess compatibility.
	if versionInfo, err := t.fetchVersionInfo(ctx); err != nil {
		slog.Warn("could not fetch controller version info", "error", err)
		result["versionWarning"] = fmt.Sprintf("connected successfully but could not retrieve version info: %v", err)
	} else {
		for k, v := range versionInfo {
			result[k] = v
		}
	}

	return jsonResult(result)
}

// fetchVersionInfo calls the controller's /version endpoint and returns a map
// with the controller version, supported API versions, and a compatibility assessment.
func (t *connectionTools) fetchVersionInfo(ctx context.Context) (map[string]any, error) {
	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, err
	}

	resp, err := mgmt.Informational.ListVersion(mgmtInfo.NewListVersionParams().WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("list version: %w", err)
	}

	data := resp.GetPayload().Data

	info := map[string]any{
		"controllerVersion": data.Version,
		"buildDate":         data.BuildDate,
		"runtimeVersion":    data.RuntimeVersion,
		"thisToolBuiltFor":  requiredAPIPath,
	}

	// Flatten the apiVersions into a readable list and check compatibility.
	compatible := false
	var supported []map[string]any
	for group, versions := range data.APIVersions {
		for label, apiVer := range versions {
			entry := map[string]any{
				"group":   group,
				"label":   label,
				"version": apiVer.Version,
			}
			if apiVer.Path != nil {
				entry["path"] = *apiVer.Path
				if *apiVer.Path == requiredAPIPath {
					compatible = true
				}
			}
			supported = append(supported, entry)
		}
	}
	info["controllerAPIVersions"] = supported
	info["compatible"] = compatible

	if compatible {
		info["compatibilityNote"] = fmt.Sprintf(
			"Controller %s advertises %s which matches this tool's API path. "+
				"The controller may have added new required fields since this tool's client library (edge-api v0.26.56) was generated; "+
				"if you see validation errors on create/update operations, the client library may need updating.",
			data.Version, requiredAPIPath)
	} else {
		info["compatibilityNote"] = fmt.Sprintf(
			"WARNING: Controller %s does not advertise %s. "+
				"This tool was built for that API path and may not work correctly. "+
				"Operations may fail with unexpected errors.",
			data.Version, requiredAPIPath)
	}

	return info, nil
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
	return jsonResult(map[string]any{
		"connected":     t.zc.Connected(),
		"controllerUrl": t.zc.ControllerURL(),
	})
}
