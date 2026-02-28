package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/netfoundry/mcp-ziti-golang/internal/config"
)

// normalizeURL tests

func TestNormalizeURL_StripPath(t *testing.T) {
	got, err := normalizeURL("https://ctrl.example.com:1280/edge/client/v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://ctrl.example.com:1280" {
		t.Errorf("expected %q, got %q", "https://ctrl.example.com:1280", got)
	}
}

func TestNormalizeURL_AlreadyNormalized(t *testing.T) {
	got, err := normalizeURL("https://ctrl.example.com:1280")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://ctrl.example.com:1280" {
		t.Errorf("expected unchanged URL, got %q", got)
	}
}

func TestNormalizeURL_MissingHost(t *testing.T) {
	_, err := normalizeURL("notaurl")
	if err == nil {
		t.Error("expected error for URL with no host")
	}
}

func TestNormalizeURL_InvalidURL(t *testing.T) {
	// %zz is an invalid percent-encoding, causing url.Parse to fail
	_, err := normalizeURL("https://%zz.example.com:1280")
	if err == nil {
		t.Error("expected error for URL with invalid percent-encoding")
	}
}

func TestNormalizeURL_StripQueryAndFragment(t *testing.T) {
	got, err := normalizeURL("https://ctrl:1280/some/path?foo=bar#frag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://ctrl:1280" {
		t.Errorf("expected path/query/fragment stripped, got %q", got)
	}
}

// stripPEMPrefix tests

func TestStripPEMPrefix_WithPrefix(t *testing.T) {
	got := stripPEMPrefix("pem:-----BEGIN CERTIFICATE-----")
	want := "-----BEGIN CERTIFICATE-----"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStripPEMPrefix_WithoutPrefix(t *testing.T) {
	got := stripPEMPrefix("-----BEGIN CERTIFICATE-----")
	if got != "-----BEGIN CERTIFICATE-----" {
		t.Errorf("unexpected result: %q", got)
	}
}

func TestStripPEMPrefix_Empty(t *testing.T) {
	got := stripPEMPrefix("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestStripPEMPrefix_OnlyPrefix(t *testing.T) {
	got := stripPEMPrefix("pem:")
	if got != "" {
		t.Errorf("expected empty string after stripping prefix, got %q", got)
	}
}

// parseCAPool tests

func TestParseCAPool_BadPEM(t *testing.T) {
	_, err := parseCAPool([]byte("this is not valid PEM data"))
	if err == nil {
		t.Error("expected error for bad PEM data")
	}
}

func TestParseCAPool_EmptyInput(t *testing.T) {
	_, err := parseCAPool([]byte(""))
	if err == nil {
		t.Error("expected error for empty PEM input")
	}
}

func TestParseCAPool_ValidCert(t *testing.T) {
	certPEM, _ := generateCertAndKey(t)
	pool, err := parseCAPool(certPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool == nil {
		t.Error("expected non-nil pool")
	}
}

// parseCertAndKey tests

func TestParseCertAndKey_BadCertPEM(t *testing.T) {
	_, _, err := parseCertAndKey([]byte("not a cert"), []byte("not a key"))
	if err == nil {
		t.Error("expected error for bad cert PEM")
	}
}

func TestParseCertAndKey_BadKeyPEM(t *testing.T) {
	certPEM, _ := generateCertAndKey(t)
	_, _, err := parseCertAndKey(certPEM, []byte("not a key"))
	if err == nil {
		t.Error("expected error for bad key PEM")
	}
}

func TestParseCertAndKey_ValidECKey(t *testing.T) {
	certPEM, keyPEM := generateCertAndKey(t)
	cert, key, err := parseCertAndKey(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cert == nil || key == nil {
		t.Error("expected non-nil cert and key")
	}
}

func TestParseCertAndKey_PKCS8Key(t *testing.T) {
	certPEM, keyPEM := generateCertAndKeyPKCS8(t)
	cert, key, err := parseCertAndKey(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("unexpected error with PKCS8 key: %v", err)
	}
	if cert == nil || key == nil {
		t.Error("expected non-nil cert and key")
	}
}

// fromIdentityFile tests

func TestFromIdentityFile_FileNotFound(t *testing.T) {
	cfg := &config.Config{IdentityFile: "/nonexistent/path/identity.json"}
	_, err := fromIdentityFile(cfg)
	if err == nil {
		t.Error("expected error when identity file does not exist")
	}
}

func TestFromIdentityFile_MalformedJSON(t *testing.T) {
	f, err := os.CreateTemp("", "identity-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("this is not json")
	f.Close()

	cfg := &config.Config{IdentityFile: f.Name()}
	_, err = fromIdentityFile(cfg)
	if err == nil {
		t.Error("expected error for malformed JSON identity file")
	}
}

func TestFromIdentityFile_MissingZtAPI(t *testing.T) {
	f, err := os.CreateTemp("", "identity-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Valid JSON but no ztAPI field
	f.WriteString(`{"id":{"cert":"","key":"","ca":""}}`)
	f.Close()

	// No controller URL in cfg, no ztAPI in file — should fail
	cfg := &config.Config{IdentityFile: f.Name()}
	_, err = fromIdentityFile(cfg)
	if err == nil {
		t.Error("expected error when ztAPI missing and no controller URL provided")
	}
}

func TestFromIdentityFile_BadPEMCert(t *testing.T) {
	f, err := os.CreateTemp("", "identity-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Has ztAPI but invalid cert/key PEM
	f.WriteString(`{"ztAPI":"https://ctrl:1280","id":{"cert":"notapem","key":"notapem","ca":""}}`)
	f.Close()

	cfg := &config.Config{IdentityFile: f.Name()}
	_, err = fromIdentityFile(cfg)
	if err == nil {
		t.Error("expected error for identity file with bad PEM cert")
	}
}

// Build tests

func TestBuild_NoCertFiles_Fails(t *testing.T) {
	// With a controller URL but no cert files, fromCertFiles should fail trying to read them
	cfg := &config.Config{ControllerURL: "https://ctrl:1280"}
	_, err := Build(cfg)
	if err == nil {
		t.Error("expected error when no cert files configured")
	}
}

// fromExtJWT tests

func TestFromExtJWT_StaticToken(t *testing.T) {
	cfg := &config.Config{
		ControllerURL: "https://ctrl:1280",
		ExtJWTToken:   "eyJhbGci.fake.token",
		CAFile:        writeTempCAFile(t),
	}
	result, err := fromExtJWT(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.Authenticator == nil {
		t.Error("expected non-nil result and authenticator")
	}
	if result.ControllerURL != "https://ctrl:1280" {
		t.Errorf("expected controller URL %q, got %q", "https://ctrl:1280", result.ControllerURL)
	}
}

func TestFromExtJWT_TokenFromFile(t *testing.T) {
	f, err := os.CreateTemp("", "jwt-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	fmt.Fprint(f, "  eyJhbGci.file.token  \n")
	f.Close()

	cfg := &config.Config{
		ControllerURL: "https://ctrl:1280",
		ExtJWTFile:    f.Name(),
		CAFile:        writeTempCAFile(t),
	}
	result, err := fromExtJWT(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.Authenticator == nil {
		t.Error("expected non-nil result and authenticator")
	}
}

func TestFromExtJWT_FileNotFound(t *testing.T) {
	cfg := &config.Config{
		ControllerURL: "https://ctrl:1280",
		ExtJWTFile:    "/nonexistent/path/token.jwt",
		CAFile:        writeTempCAFile(t),
	}
	_, err := fromExtJWT(cfg)
	if err == nil {
		t.Error("expected error when JWT file does not exist")
	}
}

func TestFromExtJWT_EmptyFile(t *testing.T) {
	f, err := os.CreateTemp("", "jwt-empty-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close() // write nothing

	cfg := &config.Config{
		ControllerURL: "https://ctrl:1280",
		ExtJWTFile:    f.Name(),
		CAFile:        writeTempCAFile(t),
	}
	_, err = fromExtJWT(cfg)
	if err == nil {
		t.Error("expected error when JWT file is empty")
	}
}

// discoverTokenURL tests

func TestDiscoverTokenURL_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token_endpoint":"https://idp.example.com/oauth/token"}`)
	}))
	defer srv.Close()

	tokenURL, err := discoverTokenURL(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokenURL != "https://idp.example.com/oauth/token" {
		t.Errorf("expected token URL %q, got %q", "https://idp.example.com/oauth/token", tokenURL)
	}
}

func TestDiscoverTokenURL_TrailingSlashIssuer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, `{"token_endpoint":"https://idp.example.com/token"}`)
	}))
	defer srv.Close()

	// Issuer with trailing slash — discoverTokenURL should strip it
	tokenURL, err := discoverTokenURL(srv.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error with trailing-slash issuer: %v", err)
	}
	if tokenURL == "" {
		t.Error("expected non-empty token URL")
	}
}

func TestDiscoverTokenURL_Non200Response(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := discoverTokenURL(srv.URL)
	if err == nil {
		t.Error("expected error for non-200 OIDC discovery response")
	}
}

func TestDiscoverTokenURL_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "this is not json")
	}))
	defer srv.Close()

	_, err := discoverTokenURL(srv.URL)
	if err == nil {
		t.Error("expected error for invalid JSON in discovery document")
	}
}

func TestDiscoverTokenURL_MissingTokenEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"issuer":"https://idp.example.com"}`)
	}))
	defer srv.Close()

	_, err := discoverTokenURL(srv.URL)
	if err == nil {
		t.Error("expected error when token_endpoint is absent from discovery document")
	}
}

// writeTempCAFile writes a self-signed cert to a temp file and returns the path.
// Used by tests that need a CAFile to avoid live network calls.
func writeTempCAFile(t *testing.T) string {
	t.Helper()
	certPEM, _ := generateCertAndKey(t)
	f, err := os.CreateTemp("", "ca-*.pem")
	if err != nil {
		t.Fatalf("creating temp CA file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.Write(certPEM); err != nil {
		t.Fatalf("writing temp CA file: %v", err)
	}
	f.Close()
	return f.Name()
}

// helpers

// generateCertAndKey creates a self-signed EC certificate and EC private key for testing.
func generateCertAndKey(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating EC key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("creating certificate: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("marshaling EC key: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}

// generateCertAndKeyPKCS8 creates a self-signed certificate with a PKCS8-encoded key.
func generateCertAndKeyPKCS8(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating EC key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "test-pkcs8"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("creating certificate: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshaling PKCS8 key: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}
