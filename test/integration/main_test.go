package integration

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

const (
	quickstartAddr = "https://localhost:1280"
	adminUser      = "admin"
	adminPass      = "admin"
)

// testClient is the shared management client used by all integration tests.
var testClient *ziticlient.Client

func TestMain(m *testing.M) {
	// Skip the entire suite if the ziti binary is not available
	if _, err := exec.LookPath("ziti"); err != nil {
		fmt.Fprintln(os.Stderr, "SKIP: 'ziti' binary not found in PATH — skipping integration tests")
		os.Exit(0)
	}

	tmpDir, err := os.MkdirTemp("", "ziti-quickstart-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: creating temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Start ziti edge quickstart
	cmd := exec.Command("ziti", "edge", "quickstart",
		"--home", tmpDir,
		"--ctrl-address", "localhost",
		"--ctrl-port", "1280",
		"--router-address", "localhost",
		"--router-port", "3022",
	)
	cmd.Stdout = os.Stderr // redirect process output to stderr, keep stdout clean
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: starting ziti edge quickstart:", err)
		os.Exit(1)
	}
	defer cmd.Process.Kill() //nolint:errcheck

	// Wait for controller to be ready (up to 90 seconds)
	if err := waitForController(quickstartAddr, 90*time.Second); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: controller not ready:", err)
		os.Exit(1)
	}

	// Create the shared test client
	testClient, err = ziticlient.New(&config.Config{
		ControllerURL: quickstartAddr,
		Username:      adminUser,
		Password:      adminPass,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: creating test client:", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// waitForController polls the controller's /version endpoint until it responds
// or the timeout elapses.
func waitForController(addr string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(addr + "/edge/management/v1/version")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				fmt.Fprintln(os.Stderr, "controller is ready")
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("controller at %s did not become ready within %s", addr, timeout)
}

// ptr returns a pointer to v. Useful for required pointer fields in API models.
func ptr[T any](v T) *T { return &v }
