package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/auth"
	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

const (
	// Default fallbacks for device authorization responses that omit these fields.
	defaultPollingIntervalSecs  = 5
	defaultDeviceCodeExpirySecs = 300
)

func registerConnectionTools(s *mcp.Server, zc *ziticlient.Client, cfg *config.Config) {
	t := &connectionTools{zc: zc, cfg: cfg}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "connect-controller",
		Description: "Connect (or reconnect) to a Ziti controller. Provide exactly one authentication method: identity JSON, username/password, client certificate, external JWT, or OIDC client credentials. For interactive OIDC login via browser, use start-oidc-login instead.",
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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "start-oidc-login",
		Description: "Start an interactive OIDC login using the OAuth 2.0 Device Authorization Grant (RFC 8628). Returns a verification URL and user code for the user to enter in their browser. After the user completes browser authentication, call complete-oidc-login to finish connecting. Parameters are optional if pre-configured at startup via --controller, --oidc-issuer, --oidc-client-id, and --oidc-audience.",
	}, t.startOIDCLogin)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "complete-oidc-login",
		Description: "Complete an interactive OIDC login started by start-oidc-login. Polls the IdP token endpoint until the user completes browser authentication, then connects to the controller.",
	}, t.completeOIDCLogin)
}

type connectionTools struct {
	zc        *ziticlient.Client
	cfg       *config.Config // startup config, may be nil; provides OIDC defaults
	mu        sync.Mutex
	oidcState *deviceAuthState
}

// deviceAuthState holds in-progress device authorization flow state.
type deviceAuthState struct {
	tokenEndpoint string
	deviceCode    string
	clientID      string
	interval      int // polling interval in seconds
	expiresAt     time.Time
	controllerURL string
	caFile        string
}

type connectControllerInput struct {
	ControllerURL string `json:"controllerUrl,omitempty" jsonschema:"controller URL, e.g. https://ctrl.example.com:1280 — required unless using identityFile or identityJson"`

	// Identity file auth — prefer identityFile (path) over identityJson (inline content)
	IdentityFile string `json:"identityFile,omitempty" jsonschema:"path to a Ziti identity JSON file on disk (preferred over identityJson)"`
	IdentityJSON string `json:"identityJson,omitempty" jsonschema:"inline Ziti identity JSON content — only use if identityFile is not available"`

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
	// Prefer file path over inline JSON — much faster for the LLM to pass a path
	identity := in.IdentityFile
	if identity == "" {
		identity = in.IdentityJSON
	}

	cfg := &config.Config{
		ControllerURL:    in.ControllerURL,
		IdentityFile:     identity,
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

// --- OAuth 2.0 Device Authorization Grant (RFC 8628) login tools ---

type startOIDCLoginInput struct {
	ControllerURL string   `json:"controllerUrl,omitempty" jsonschema:"controller URL — uses startup --controller default if omitted"`
	OIDCIssuer    string   `json:"oidcIssuer,omitempty" jsonschema:"OIDC issuer URL — uses startup --oidc-issuer default if omitted"`
	OIDCClientID  string   `json:"oidcClientId,omitempty" jsonschema:"OIDC client ID — uses startup --oidc-client-id default if omitted"`
	OIDCAudience  string   `json:"oidcAudience,omitempty" jsonschema:"OIDC audience (optional) — uses startup --oidc-audience default if omitted"`
	CAFile        string   `json:"caFile,omitempty" jsonschema:"path to CA bundle PEM file (optional)"`
	Scopes        []string `json:"scopes,omitempty" jsonschema:"OAuth scopes (defaults to openid)"`
}

// deviceAuthResponse is the JSON response from the device authorization endpoint.
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func (t *connectionTools) startOIDCLogin(_ context.Context, _ *mcp.CallToolRequest, in startOIDCLoginInput) (*mcp.CallToolResult, any, error) {
	// Apply defaults from startup config for any fields not provided by the LLM
	if t.cfg != nil {
		if in.ControllerURL == "" {
			in.ControllerURL = t.cfg.ControllerURL
		}
		if in.OIDCIssuer == "" {
			in.OIDCIssuer = t.cfg.OIDCIssuer
		}
		if in.OIDCClientID == "" {
			in.OIDCClientID = t.cfg.OIDCClientID
		}
		if in.OIDCAudience == "" {
			in.OIDCAudience = t.cfg.OIDCAudience
		}
		if in.CAFile == "" {
			in.CAFile = t.cfg.CAFile
		}
	}

	// Validate required fields (after applying defaults)
	if in.ControllerURL == "" {
		return nil, nil, fmt.Errorf("controllerUrl is required (provide it or configure --controller at startup)")
	}
	if in.OIDCIssuer == "" {
		return nil, nil, fmt.Errorf("oidcIssuer is required (provide it or configure --oidc-issuer at startup)")
	}
	if in.OIDCClientID == "" {
		return nil, nil, fmt.Errorf("oidcClientId is required (provide it or configure --oidc-client-id at startup)")
	}

	// Discover OIDC endpoints
	endpoints, err := auth.DiscoverOIDCEndpoints(in.OIDCIssuer)
	if err != nil {
		return nil, nil, fmt.Errorf("OIDC discovery for %q: %w", in.OIDCIssuer, err)
	}
	if endpoints.DeviceAuthorizationEndpoint == "" {
		return nil, nil, fmt.Errorf("OIDC issuer %q does not support the device authorization flow (no device_authorization_endpoint in discovery document)", in.OIDCIssuer)
	}

	// Set up scopes
	scopes := in.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid"}
	}

	// Request device code from the IdP
	formData := url.Values{
		"client_id": {in.OIDCClientID},
		"scope":     {strings.Join(scopes, " ")},
	}
	if in.OIDCAudience != "" {
		formData.Set("audience", in.OIDCAudience)
	}

	//nolint:gosec // URL is from OIDC discovery, not user input
	resp, err := http.PostForm(endpoints.DeviceAuthorizationEndpoint, formData)
	if err != nil {
		return nil, nil, fmt.Errorf("requesting device authorization: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading device authorization response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("device authorization request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var deviceResp deviceAuthResponse
	if err := json.Unmarshal(body, &deviceResp); err != nil {
		return nil, nil, fmt.Errorf("decoding device authorization response: %w", err)
	}

	if deviceResp.DeviceCode == "" || deviceResp.UserCode == "" {
		return nil, nil, fmt.Errorf("device authorization response missing device_code or user_code")
	}

	interval := deviceResp.Interval
	if interval <= 0 {
		interval = defaultPollingIntervalSecs
	}

	expiresIn := deviceResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = defaultDeviceCodeExpirySecs
	}

	// Store state for complete-oidc-login
	t.mu.Lock()
	t.oidcState = &deviceAuthState{
		tokenEndpoint: endpoints.TokenEndpoint,
		deviceCode:    deviceResp.DeviceCode,
		clientID:      in.OIDCClientID,
		interval:      interval,
		expiresAt:     time.Now().Add(time.Duration(expiresIn) * time.Second),
		controllerURL: in.ControllerURL,
		caFile:        in.CAFile,
	}
	t.mu.Unlock()

	// Build the user-facing message
	var codeStep string
	var verificationURL string
	if deviceResp.VerificationURIComplete != "" {
		verificationURL = deviceResp.VerificationURIComplete
		codeStep = fmt.Sprintf("   (The code **%s** is pre-filled in the URL)\n", deviceResp.UserCode)
	} else {
		verificationURL = deviceResp.VerificationURI
		codeStep = fmt.Sprintf("2. Enter this code when prompted: **%s**\n", deviceResp.UserCode)
	}

	message := fmt.Sprintf(
		"OIDC device login initiated. Please ask the user to:\n\n"+
			"1. Open this URL in their browser: %s\n"+
			"%s"+
			"3. Complete authentication with the identity provider\n\n"+
			"After the user completes authentication, call complete-oidc-login to finish connecting.\n"+
			"The code expires in %d seconds.",
		verificationURL, codeStep, expiresIn)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
	}, nil, nil
}

type completeOIDCLoginInput struct{}

// tokenResponse is the JSON response from the token endpoint during device code polling.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func (t *connectionTools) completeOIDCLogin(ctx context.Context, _ *mcp.CallToolRequest, _ completeOIDCLoginInput) (*mcp.CallToolResult, any, error) {
	t.mu.Lock()
	state := t.oidcState
	t.mu.Unlock()

	if state == nil {
		return nil, nil, fmt.Errorf("no OIDC login in progress — call start-oidc-login first")
	}

	// Poll the token endpoint until the user completes authentication
	token, err := pollForToken(ctx, state)

	// Clear state regardless of outcome
	t.mu.Lock()
	t.oidcState = nil
	t.mu.Unlock()

	if err != nil {
		return nil, nil, err
	}

	// Build auth result from the access token
	authResult, err := auth.FromToken(state.controllerURL, token, state.caFile)
	if err != nil {
		return nil, nil, fmt.Errorf("building authenticator from token: %w", err)
	}

	// Connect to the controller
	if err := t.zc.ConnectWithAuth(authResult); err != nil {
		return nil, nil, err
	}

	return t.statusResult()
}

// pollForToken polls the token endpoint at the specified interval until the user
// completes authentication, the device code expires, or the context is cancelled.
func pollForToken(ctx context.Context, state *deviceAuthState) (string, error) {
	ticker := time.NewTicker(time.Duration(state.interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("cancelled while waiting for user to authenticate: %w", ctx.Err())
		case <-ticker.C:
			if time.Now().After(state.expiresAt) {
				return "", fmt.Errorf("device code expired — call start-oidc-login to try again")
			}

			token, done, err := requestToken(state)
			if err != nil {
				return "", err
			}
			if done {
				return token, nil
			}
			// Not done yet — continue polling
		}
	}
}

// requestToken makes a single token request. Returns (token, done, error).
// done=false means keep polling; done=true means we have a token or a terminal error.
func requestToken(state *deviceAuthState) (string, bool, error) {
	formData := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {state.deviceCode},
		"client_id":   {state.clientID},
	}

	//nolint:gosec // URL is from OIDC discovery
	resp, err := http.PostForm(state.tokenEndpoint, formData)
	if err != nil {
		return "", true, fmt.Errorf("polling token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", true, fmt.Errorf("reading token response: %w", err)
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", true, fmt.Errorf("decoding token response: %w", err)
	}

	// Check for pending/slow_down responses (keep polling)
	switch tokenResp.Error {
	case "authorization_pending":
		return "", false, nil
	case "slow_down":
		// IdP wants us to slow down — the next tick will respect the interval
		return "", false, nil
	case "access_denied":
		return "", true, fmt.Errorf("user denied the authorization request")
	case "expired_token":
		return "", true, fmt.Errorf("device code expired — call start-oidc-login to try again")
	case "":
		// No error — we should have a token
		if tokenResp.AccessToken == "" {
			return "", true, fmt.Errorf("token response missing access_token")
		}
		return tokenResp.AccessToken, true, nil
	default:
		desc := tokenResp.ErrorDesc
		if desc == "" {
			desc = tokenResp.Error
		}
		return "", true, fmt.Errorf("token endpoint error: %s", desc)
	}
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
		summary = "Not connected to a Ziti controller. Connect before calling any other tools.\n\n" +
			"IMPORTANT: Present ALL of the following options to the user as separate choices. Do NOT combine or merge them:\n" +
			"  1. Interactive browser login — user opens a link and authenticates in their browser (best for human users with a 3rd-party identity provider)\n" +
			"  2. Identity JSON file — provide a path to a Ziti identity .json file on disk\n" +
			"  3. Username and password — authenticate with the controller's built-in user database\n" +
			"  4. Client certificate — authenticate with a TLS client cert and private key\n" +
			"  5. External JWT token — authenticate with a pre-issued JWT string\n" +
			"  6. OIDC client credentials — authenticate with a client ID and secret from an identity provider (for service accounts)"
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
