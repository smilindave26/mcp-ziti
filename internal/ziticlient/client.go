package ziticlient

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/netfoundry/mcp-ziti-golang/internal/auth"
	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_util"
)

const refreshWindow = 5 * time.Minute

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
	defer c.mu.Unlock()

	// Swap authenticator and URL for the new connection attempt.
	oldAuth, oldURL, oldConnected := c.authenticator, c.ctrlURL, c.connected
	c.authenticator = authResult.Authenticator
	c.ctrlURL = ctrlURL

	if err := c.authenticate(); err != nil {
		// Restore previous state on failure.
		c.authenticator, c.ctrlURL, c.connected = oldAuth, oldURL, oldConnected
		return fmt.Errorf("authentication failed: %w", err)
	}

	c.connected = true
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
