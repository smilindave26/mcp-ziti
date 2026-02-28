package ziticlient

import (
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

// Client wraps the OpenZiti Management API client with transparent session refresh.
// Before each use, call Mgmt() which re-authenticates if the session is near expiry.
type Client struct {
	authenticator rest_util.Authenticator
	ctrlURL       *url.URL
	mgmt          *rest_management_api_client.ZitiEdgeManagement
	expiresAt     time.Time
	mu            sync.Mutex
}

// New creates a Client by authenticating with the controller using the provided config.
func New(cfg *config.Config) (*Client, error) {
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

	return c, nil
}

// Mgmt returns the management API client, refreshing the session if it is
// within the refresh window of expiry.
func (c *Client) Mgmt() (*rest_management_api_client.ZitiEdgeManagement, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

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
