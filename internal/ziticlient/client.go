package ziticlient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/netfoundry/mcp-ziti-golang/internal/auth"
	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/openziti/edge-api/rest_management_api_client"
	mgmtInfo "github.com/openziti/edge-api/rest_management_api_client/informational"
	"github.com/openziti/edge-api/rest_util"
)

const (
	refreshWindow = 5 * time.Minute

	// RequiredAPIPath is the management API path this client was built against.
	RequiredAPIPath = "/edge/management/v1"

	// EdgeAPIVersion is the version of the edge-api Go module used by this client.
	EdgeAPIVersion = "v0.26.56"
)

// APIVersionEntry describes a single API version advertised by the controller.
type APIVersionEntry struct {
	Group   string `json:"group"`
	Label   string `json:"label"`
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
}

// VersionInfo holds the controller version and API compatibility assessment.
type VersionInfo struct {
	ControllerVersion  string            `json:"controllerVersion"`
	BuildDate          string            `json:"buildDate,omitempty"`
	RuntimeVersion     string            `json:"runtimeVersion,omitempty"`
	APIVersions        []APIVersionEntry `json:"controllerAPIVersions"`
	ThisToolBuiltFor   string            `json:"thisToolBuiltFor"`
	EdgeAPIModule      string            `json:"edgeApiModule"`
	Compatible         bool              `json:"compatible"`
	CompatibilityNote  string            `json:"compatibilityNote"`
}

// ErrNotConnected is returned by Mgmt() when the client has no active controller connection.
var ErrNotConnected = errors.New("not connected to a Ziti controller — use the connect-controller tool to connect")

// Client wraps the OpenZiti Management API client with transparent session refresh.
// Before each use, call Mgmt() which re-authenticates if the session is near expiry.
// A Client may start in a disconnected state and be connected later via Connect().
type Client struct {
	authenticator rest_util.Authenticator
	ctrlURL       *url.URL
	mgmt          *rest_management_api_client.ZitiEdgeManagement
	expiresAt     time.Time
	connected     bool
	versionInfo   *VersionInfo
	mu            sync.Mutex
}

// NewForTest returns a disconnected Client for use in tests that exercise the
// MCP protocol layer without a real controller.
func NewForTest() *Client {
	return &Client{}
}

// New creates a Client from the provided config. If the config has auth
// credentials, it authenticates immediately. Otherwise it returns a
// disconnected Client that can be connected later via Connect().
func New(cfg *config.Config) (*Client, error) {
	if !cfg.HasAuth() {
		slog.Info("no credentials configured, starting disconnected")
		return &Client{}, nil
	}

	authResult, err := auth.Build(cfg)
	if err != nil {
		return nil, fmt.Errorf("building authenticator: %w", err)
	}

	ctrlURL, err := url.Parse(authResult.ControllerURL)
	if err != nil {
		return nil, fmt.Errorf("parsing controller URL: %w", err)
	}

	c := &Client{
		authenticator: authResult.Authenticator,
		ctrlURL:       ctrlURL,
	}

	if err := c.authenticate(); err != nil {
		return nil, fmt.Errorf("initial authentication failed: %w", err)
	}

	c.connected = true
	c.fetchAndLogVersionInfo()
	return c, nil
}

// Connect authenticates against a controller using the provided config,
// replacing any existing connection. Thread-safe.
func (c *Client) Connect(cfg *config.Config) error {
	if err := cfg.ValidateAuth(); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	authResult, err := auth.Build(cfg)
	if err != nil {
		return fmt.Errorf("building authenticator: %w", err)
	}

	ctrlURL, err := url.Parse(authResult.ControllerURL)
	if err != nil {
		return fmt.Errorf("parsing controller URL: %w", err)
	}

	c.mu.Lock()

	// Swap authenticator and URL for the new connection attempt.
	oldAuth, oldURL, oldConnected := c.authenticator, c.ctrlURL, c.connected
	c.authenticator = authResult.Authenticator
	c.ctrlURL = ctrlURL

	if err := c.authenticate(); err != nil {
		// Restore previous state on failure.
		c.authenticator, c.ctrlURL, c.connected = oldAuth, oldURL, oldConnected
		c.mu.Unlock()
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.connected = true
	c.mu.Unlock()

	// Fetch version info outside the lock (it calls Mgmt which acquires it).
	c.fetchAndLogVersionInfo()
	return nil
}

// Disconnect clears the active controller connection. Thread-safe.
// Returns ErrNotConnected if already disconnected.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrNotConnected
	}

	c.authenticator = nil
	c.ctrlURL = nil
	c.mgmt = nil
	c.expiresAt = time.Time{}
	c.connected = false
	c.versionInfo = nil

	slog.Info("disconnected from controller")
	return nil
}

// Connected returns true if the client has an active controller connection.
func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// ControllerURL returns the URL of the connected controller, or empty string if disconnected.
func (c *Client) ControllerURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ctrlURL == nil {
		return ""
	}
	return c.ctrlURL.String()
}

// GetVersionInfo returns the cached version info from the last successful connection,
// or nil if disconnected or version info was not fetched.
func (c *Client) GetVersionInfo() *VersionInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.versionInfo
}

// fetchAndLogVersionInfo fetches version info from the controller and logs it.
// Must NOT be called with c.mu held (it calls Mgmt which acquires the lock).
func (c *Client) fetchAndLogVersionInfo() {
	info, err := c.fetchVersionInfo()
	if err != nil {
		slog.Warn("could not fetch controller version info", "error", err)
		return
	}

	c.mu.Lock()
	c.versionInfo = info
	c.mu.Unlock()

	slog.Info("controller version info",
		"controllerVersion", info.ControllerVersion,
		"compatible", info.Compatible,
		"thisToolBuiltFor", info.ThisToolBuiltFor,
		"edgeApiModule", info.EdgeAPIModule,
	)
	if !info.Compatible {
		slog.Warn("API version mismatch", "note", info.CompatibilityNote)
	}
}

// fetchVersionInfo calls the controller's /version endpoint and builds a VersionInfo.
func (c *Client) fetchVersionInfo() (*VersionInfo, error) {
	mgmt, err := c.Mgmt()
	if err != nil {
		return nil, err
	}

	resp, err := mgmt.Informational.ListVersion(
		mgmtInfo.NewListVersionParams().WithContext(context.Background()))
	if err != nil {
		return nil, fmt.Errorf("list version: %w", err)
	}

	data := resp.GetPayload().Data

	info := &VersionInfo{
		ControllerVersion: data.Version,
		BuildDate:         data.BuildDate,
		RuntimeVersion:    data.RuntimeVersion,
		ThisToolBuiltFor:  RequiredAPIPath,
		EdgeAPIModule:     EdgeAPIVersion,
	}

	// Flatten apiVersions and check compatibility.
	for group, versions := range data.APIVersions {
		for label, apiVer := range versions {
			entry := APIVersionEntry{
				Group:   group,
				Label:   label,
				Version: apiVer.Version,
			}
			if apiVer.Path != nil {
				entry.Path = *apiVer.Path
				if *apiVer.Path == RequiredAPIPath {
					info.Compatible = true
				}
			}
			info.APIVersions = append(info.APIVersions, entry)
		}
	}

	if info.Compatible {
		info.CompatibilityNote = fmt.Sprintf(
			"Controller %s advertises %s which matches this tool's API path. "+
				"The controller may have added new required fields since this tool's "+
				"client library (edge-api %s) was generated; if you see validation "+
				"errors on create/update operations, the client library may need updating.",
			data.Version, RequiredAPIPath, EdgeAPIVersion)
	} else {
		info.CompatibilityNote = fmt.Sprintf(
			"WARNING: Controller %s does not advertise %s. "+
				"This tool was built for that API path and may not work correctly. "+
				"Operations may fail with unexpected errors.",
			data.Version, RequiredAPIPath)
	}

	return info, nil
}

// Mgmt returns the management API client, refreshing the session if it is
// within the refresh window of expiry. Returns ErrNotConnected if the client
// has no active connection.
func (c *Client) Mgmt() (*rest_management_api_client.ZitiEdgeManagement, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, ErrNotConnected
	}

	if time.Until(c.expiresAt) < refreshWindow {
		slog.Info("session token near expiry, refreshing", "expiresAt", c.expiresAt)
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("session refresh failed: %w", err)
		}
	}

	return c.mgmt, nil
}

// authenticate performs a fresh authentication against the controller and
// stores the new management client and session expiry.
// Must be called with c.mu held (or during construction before the client is shared).
func (c *Client) authenticate() error {
	session, err := c.authenticator.Authenticate(c.ctrlURL)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	if session.Token == nil || *session.Token == "" {
		return fmt.Errorf("controller returned empty session token")
	}

	httpClient, err := c.authenticator.BuildHttpClient()
	if err != nil {
		return fmt.Errorf("building HTTP client: %w", err)
	}

	mgmt, err := rest_util.NewEdgeManagementClientWithToken(httpClient, c.ctrlURL.String(), *session.Token)
	if err != nil {
		return fmt.Errorf("creating management client: %w", err)
	}

	c.mgmt = mgmt

	// Record session expiry for refresh logic
	if session.ExpiresAt != nil {
		c.expiresAt = time.Time(*session.ExpiresAt)
	} else {
		// Default to 30 minutes if server doesn't provide expiry
		c.expiresAt = time.Now().Add(30 * time.Minute)
	}

	slog.Info("authenticated successfully", "expiresAt", c.expiresAt)
	return nil
}
