package auth

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	openapiclient "github.com/go-openapi/runtime/client"
	"github.com/netfoundry/mcp-ziti-golang/internal/config"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_management_api_client/authentication"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/edge-api/rest_util"
	"golang.org/x/oauth2/clientcredentials"
)

// Result holds the authenticator and normalized controller URL ready to use with
// the management API.
type Result struct {
	Authenticator rest_util.Authenticator
	ControllerURL string // normalized to https://host:port (no path)
}

// Build constructs an Authenticator and controller URL from the provided config.
func Build(cfg *config.Config) (*Result, error) {
	switch {
	case cfg.IdentityFile != "":
		return fromIdentityFile(cfg)
	case cfg.Username != "":
		return fromUpdb(cfg)
	case cfg.ExtJWTToken != "" || cfg.ExtJWTFile != "":
		return fromExtJWT(cfg)
	case cfg.OIDCIssuer != "":
		return fromOIDC(cfg)
	default:
		return fromCertFiles(cfg)
	}
}

// fromUpdb builds an Authenticator using username/password.
func fromUpdb(cfg *config.Config) (*Result, error) {
	ctrlURL, err := normalizeURL(cfg.ControllerURL)
	if err != nil {
		return nil, err
	}

	caPool, err := resolveCAPool(cfg.CAFile, ctrlURL)
	if err != nil {
		return nil, err
	}

	auth := rest_util.NewAuthenticatorUpdb(cfg.Username, cfg.Password)
	auth.RootCas = caPool

	return &Result{Authenticator: auth, ControllerURL: ctrlURL}, nil
}

// fromCertFiles builds an Authenticator using explicit cert/key PEM files.
func fromCertFiles(cfg *config.Config) (*Result, error) {
	ctrlURL, err := normalizeURL(cfg.ControllerURL)
	if err != nil {
		return nil, err
	}

	certPEM, err := os.ReadFile(cfg.CertFile)
	if err != nil {
		return nil, fmt.Errorf("reading cert file: %w", err)
	}

	keyPEM, err := os.ReadFile(cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("reading key file: %w", err)
	}

	x509Cert, privateKey, err := parseCertAndKey(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	caPool, err := resolveCAPool(cfg.CAFile, ctrlURL)
	if err != nil {
		return nil, err
	}

	auth := rest_util.NewAuthenticatorCert(x509Cert, privateKey)
	auth.RootCas = caPool

	return &Result{Authenticator: auth, ControllerURL: ctrlURL}, nil
}

// fromExtJWT builds an Authenticator using a static external JWT token.
// The token may be provided as a literal string (cfg.ExtJWTToken) or read
// from a file (cfg.ExtJWTFile).
func fromExtJWT(cfg *config.Config) (*Result, error) {
	ctrlURL, err := normalizeURL(cfg.ControllerURL)
	if err != nil {
		return nil, err
	}

	var token string
	if cfg.ExtJWTToken != "" {
		token = cfg.ExtJWTToken
	} else {
		data, err := os.ReadFile(cfg.ExtJWTFile)
		if err != nil {
			return nil, fmt.Errorf("reading JWT file: %w", err)
		}
		token = strings.TrimSpace(string(data))
	}

	if token == "" {
		return nil, fmt.Errorf("external JWT is empty")
	}

	caPool, err := resolveCAPool(cfg.CAFile, ctrlURL)
	if err != nil {
		return nil, err
	}

	tokenFunc := func() (string, error) { return token, nil }
	a := &extJWTAuthenticator{tokenFunc: tokenFunc}
	a.RootCas = caPool

	return &Result{Authenticator: a, ControllerURL: ctrlURL}, nil
}

// fromOIDC builds an Authenticator using the OIDC client credentials flow.
// A fresh JWT is fetched from the token endpoint on every Authenticate() call,
// so session refresh works without operator intervention.
func fromOIDC(cfg *config.Config) (*Result, error) {
	ctrlURL, err := normalizeURL(cfg.ControllerURL)
	if err != nil {
		return nil, err
	}

	tokenURL := cfg.OIDCTokenURL
	if tokenURL == "" {
		tokenURL, err = discoverTokenURL(cfg.OIDCIssuer)
		if err != nil {
			return nil, fmt.Errorf("OIDC discovery for %q: %w", cfg.OIDCIssuer, err)
		}
	}

	endpointParams := url.Values{}
	if cfg.OIDCAudience != "" {
		endpointParams.Set("audience", cfg.OIDCAudience)
	}

	oauthCfg := clientcredentials.Config{
		ClientID:       cfg.OIDCClientID,
		ClientSecret:   cfg.OIDCClientSecret,
		TokenURL:       tokenURL,
		EndpointParams: endpointParams,
	}

	caPool, err := resolveCAPool(cfg.CAFile, ctrlURL)
	if err != nil {
		return nil, err
	}

	tokenFunc := func() (string, error) {
		tok, err := oauthCfg.Token(context.Background())
		if err != nil {
			return "", fmt.Errorf("fetching OIDC token: %w", err)
		}
		return tok.AccessToken, nil
	}

	a := &extJWTAuthenticator{tokenFunc: tokenFunc}
	a.RootCas = caPool

	return &Result{Authenticator: a, ControllerURL: ctrlURL}, nil
}

// extJWTAuthenticator implements rest_util.Authenticator for the ext-jwt
// authentication method. tokenFunc is called on every Authenticate() invocation,
// allowing callers to supply fresh tokens (e.g. via OIDC) or a static value.
type extJWTAuthenticator struct {
	rest_util.AuthenticatorBase
	tokenFunc func() (string, error)
}

func (a *extJWTAuthenticator) Authenticate(ctrlURL *url.URL) (*rest_model.CurrentAPISessionDetail, error) {
	token, err := a.tokenFunc()
	if err != nil {
		return nil, fmt.Errorf("obtaining JWT: %w", err)
	}

	httpClient, err := a.BuildHttpClient()
	if err != nil {
		return nil, err
	}

	path := rest_management_api_client.DefaultBasePath
	if ctrlURL.Path != "" && ctrlURL.Path != "/" {
		path = ctrlURL.Path
	}

	clientRuntime := openapiclient.NewWithClient(ctrlURL.Host, path, rest_management_api_client.DefaultSchemes, httpClient)
	clientRuntime.DefaultAuthentication = &rest_util.HeaderAuth{
		HeaderName:  "Authorization",
		HeaderValue: "Bearer " + token,
	}

	client := rest_management_api_client.New(clientRuntime, nil)

	params := &authentication.AuthenticateParams{
		Auth: &rest_model.Authenticate{
			ConfigTypes: a.ConfigTypes,
			EnvInfo:     a.EnvInfo,
			SdkInfo:     a.SdkInfo,
		},
		Method:  "ext-jwt",
		Context: context.Background(),
	}

	resp, err := client.Authentication.Authenticate(params)
	if err != nil {
		return nil, fmt.Errorf("ext-jwt authentication: %w", err)
	}

	if resp.GetPayload() == nil {
		return nil, fmt.Errorf("ext-jwt authentication: nil payload")
	}

	return resp.GetPayload().Data, nil
}

func (a *extJWTAuthenticator) BuildHttpClient() (*http.Client, error) {
	return a.BuildHttpClientWithModifyTls(nil)
}

// discoverTokenURL fetches the OIDC discovery document from the issuer and
// returns the token_endpoint URL. If OIDCTokenURL is set in config, this
// function is never called.
func discoverTokenURL(issuer string) (string, error) {
	discoveryURL := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	//nolint:gosec // URL is caller-provided configuration, not user input
	resp, err := http.Get(discoveryURL)
	if err != nil {
		return "", fmt.Errorf("fetching OIDC discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OIDC discovery returned HTTP %d from %s", resp.StatusCode, discoveryURL)
	}

	var doc struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("decoding OIDC discovery document: %w", err)
	}
	if doc.TokenEndpoint == "" {
		return "", fmt.Errorf("OIDC discovery document at %s missing token_endpoint", discoveryURL)
	}

	return doc.TokenEndpoint, nil
}

// identityFileJSON is the top-level structure of a Ziti identity JSON file.
type identityFileJSON struct {
	ZtAPI string `json:"ztAPI"`
	ID    struct {
		Cert string `json:"cert"`
		Key  string `json:"key"`
		CA   string `json:"ca"`
	} `json:"id"`
}

// fromIdentityFile builds an Authenticator by parsing a Ziti identity JSON.
// The controller URL is extracted from the ztAPI field unless overridden in cfg.
// cfg.IdentityFile may be a file path or inline JSON content (detected by
// leading '{').
func fromIdentityFile(cfg *config.Config) (*Result, error) {
	var data []byte
	if strings.HasPrefix(strings.TrimSpace(cfg.IdentityFile), "{") {
		data = []byte(cfg.IdentityFile)
	} else {
		var err error
		data, err = os.ReadFile(cfg.IdentityFile)
		if err != nil {
			return nil, fmt.Errorf("reading identity file: %w", err)
		}
	}

	var idFile identityFileJSON
	if err := json.Unmarshal(data, &idFile); err != nil {
		return nil, fmt.Errorf("parsing identity file: %w", err)
	}

	// Determine controller URL: explicit flag/env takes precedence, then ztAPI field
	rawCtrlURL := cfg.ControllerURL
	if rawCtrlURL == "" {
		if idFile.ZtAPI == "" {
			return nil, fmt.Errorf("identity file has no ztAPI field; provide --controller explicitly")
		}
		rawCtrlURL = idFile.ZtAPI
	}

	ctrlURL, err := normalizeURL(rawCtrlURL)
	if err != nil {
		return nil, err
	}

	// Parse cert/key from the identity file
	certPEM := stripPEMPrefix(idFile.ID.Cert)
	keyPEM := stripPEMPrefix(idFile.ID.Key)
	caPEM := stripPEMPrefix(idFile.ID.CA)

	x509Cert, privateKey, err := parseCertAndKey([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("identity file credentials: %w", err)
	}

	// Build CA pool: prefer explicit --ca flag, then identity file CA, then well-known
	var caPool *x509.CertPool
	if cfg.CAFile != "" {
		caPool, err = loadCAPool(cfg.CAFile)
	} else if caPEM != "" {
		caPool, err = parseCAPool([]byte(caPEM))
	} else {
		caPool, err = rest_util.GetControllerWellKnownCaPool(ctrlURL)
	}
	if err != nil {
		return nil, fmt.Errorf("building CA pool: %w", err)
	}

	auth := rest_util.NewAuthenticatorCert(x509Cert, privateKey)
	auth.RootCas = caPool

	return &Result{Authenticator: auth, ControllerURL: ctrlURL}, nil
}

// normalizeURL ensures the URL is reduced to https://host:port so that the
// management API base path (/edge/management/v1) is applied correctly.
// Identity files store ztAPI as https://host:port/edge/client/v1 — the path
// must be stripped before passing to management API helpers.
func normalizeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid controller URL %q: %w", raw, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid controller URL %q: missing host", raw)
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

// resolveCAPool returns a CA pool: loads from caFile if set, otherwise fetches
// the controller's well-known CA bundle.
func resolveCAPool(caFile, ctrlURL string) (*x509.CertPool, error) {
	if caFile != "" {
		return loadCAPool(caFile)
	}
	pool, err := rest_util.GetControllerWellKnownCaPool(ctrlURL)
	if err != nil {
		return nil, fmt.Errorf("fetching well-known CAs from %s: %w", ctrlURL, err)
	}
	return pool, nil
}

// loadCAPool reads a PEM file and returns an x509.CertPool.
func loadCAPool(path string) (*x509.CertPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading CA file %q: %w", path, err)
	}
	return parseCAPool(data)
}

// parseCAPool parses PEM-encoded certificates into a CertPool.
func parseCAPool(pemData []byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("no valid certificates found in CA PEM data")
	}
	return pool, nil
}

// parseCertAndKey decodes PEM cert and key bytes into their Go types.
func parseCertAndKey(certPEM, keyPEM []byte) (*x509.Certificate, interface{}, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode certificate PEM block")
	}

	x509Cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing certificate: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode private key PEM block")
	}

	privateKey, err := parsePrivateKey(keyBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing private key: %w", err)
	}

	return x509Cert, privateKey, nil
}

// parsePrivateKey attempts to parse various private key types from a PEM block.
func parsePrivateKey(block *pem.Block) (interface{}, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		// Try PKCS8 as a fallback for unknown types
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unsupported key type %q", block.Type)
		}
		return key, nil
	}
}

// stripPEMPrefix removes the "pem:" prefix that Ziti identity files prepend to
// PEM-encoded values.
func stripPEMPrefix(s string) string {
	return strings.TrimPrefix(s, "pem:")
}
