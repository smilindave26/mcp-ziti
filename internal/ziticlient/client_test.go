package ziticlient

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/openziti/edge-api/rest_model"
)

// mockAuthenticator implements rest_util.Authenticator for session refresh testing.
type mockAuthenticator struct {
	callCount int
	token     string
	expiresIn time.Duration
}

func (m *mockAuthenticator) Authenticate(_ *url.URL) (*rest_model.CurrentAPISessionDetail, error) {
	m.callCount++
	token := m.token
	exp := strfmt.DateTime(time.Now().Add(m.expiresIn))
	expSecs := int64(m.expiresIn.Seconds())
	authID := "mock-auth"
	return &rest_model.CurrentAPISessionDetail{
		ExpirationSeconds: &expSecs,
		ExpiresAt:         &exp,
		APISessionDetail: rest_model.APISessionDetail{
			Token:           &token,
			AuthenticatorID: &authID,
			AuthQueries:     rest_model.AuthQueryList{},
			ConfigTypes:     []string{},
		},
	}, nil
}

func (m *mockAuthenticator) BuildHttpClient() (*http.Client, error) {
	return &http.Client{}, nil
}

func (m *mockAuthenticator) SetInfo(_ *rest_model.EnvInfo, _ *rest_model.SdkInfo) {}

// newTestClient creates a Client directly with the mock authenticator,
// bypassing New() to avoid needing a live controller.
func newTestClient(t *testing.T, mock *mockAuthenticator) *Client {
	t.Helper()
	ctrlURL, err := url.Parse("https://localhost:1280")
	if err != nil {
		t.Fatalf("parsing URL: %v", err)
	}
	c := &Client{
		authenticator: mock,
		ctrlURL:       ctrlURL,
	}
	if err := c.authenticate(); err != nil {
		t.Fatalf("initial authenticate: %v", err)
	}
	c.connected = true
	return c
}

func TestMgmt_ErrNotConnected(t *testing.T) {
	c := &Client{} // disconnected
	_, err := c.Mgmt()
	if err != ErrNotConnected {
		t.Errorf("expected ErrNotConnected, got %v", err)
	}
}

func TestConnected_Disconnected(t *testing.T) {
	c := &Client{}
	if c.Connected() {
		t.Error("expected Connected() == false for new empty client")
	}
}

func TestControllerURL_Disconnected(t *testing.T) {
	c := &Client{}
	if got := c.ControllerURL(); got != "" {
		t.Errorf("expected empty ControllerURL, got %q", got)
	}
}

func TestMgmt_RefreshesSessionNearExpiry(t *testing.T) {
	mock := &mockAuthenticator{token: "initial-token", expiresIn: 30 * time.Minute}
	c := newTestClient(t, mock)
	afterInit := mock.callCount // should be 1

	// Force expiry into refresh window (< 5 minutes remaining)
	c.expiresAt = time.Now().Add(2 * time.Minute)

	if _, err := c.Mgmt(); err != nil {
		t.Fatalf("Mgmt() returned error: %v", err)
	}

	if mock.callCount != afterInit+1 {
		t.Errorf("expected %d authenticate calls, got %d", afterInit+1, mock.callCount)
	}
}

func TestMgmt_NoRefreshWhenSessionFresh(t *testing.T) {
	mock := &mockAuthenticator{token: "test-token", expiresIn: 30 * time.Minute}
	c := newTestClient(t, mock)
	afterInit := mock.callCount

	// Session is well within the refresh window — no refresh expected
	c.expiresAt = time.Now().Add(60 * time.Minute)

	if _, err := c.Mgmt(); err != nil {
		t.Fatalf("Mgmt() returned error: %v", err)
	}

	if mock.callCount != afterInit {
		t.Errorf("expected no re-auth, but authenticate was called again (count=%d)", mock.callCount)
	}
}

func TestMgmt_RefreshesExpiredSession(t *testing.T) {
	mock := &mockAuthenticator{token: "expired-token", expiresIn: 30 * time.Minute}
	c := newTestClient(t, mock)
	afterInit := mock.callCount

	// Session already expired
	c.expiresAt = time.Now().Add(-1 * time.Minute)

	if _, err := c.Mgmt(); err != nil {
		t.Fatalf("Mgmt() returned error: %v", err)
	}

	if mock.callCount != afterInit+1 {
		t.Errorf("expected re-auth for expired session, got %d calls (want %d)", mock.callCount, afterInit+1)
	}
}

func TestMgmt_ReturnsNonNilClient(t *testing.T) {
	mock := &mockAuthenticator{token: "valid-token", expiresIn: 30 * time.Minute}
	c := newTestClient(t, mock)
	c.expiresAt = time.Now().Add(60 * time.Minute) // fresh session

	mgmt, err := c.Mgmt()
	if err != nil {
		t.Fatalf("Mgmt() returned error: %v", err)
	}
	if mgmt == nil {
		t.Error("expected non-nil management client")
	}
}
