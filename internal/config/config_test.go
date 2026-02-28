package config

import (
	"testing"
)

func TestValidate_NoAuth(t *testing.T) {
	c := &Config{}
	if err := c.validate(); err != nil {
		t.Errorf("zero auth should be allowed (connect at runtime): %v", err)
	}
}

func TestHasAuth_Empty(t *testing.T) {
	c := &Config{}
	if c.HasAuth() {
		t.Error("expected HasAuth() == false for empty config")
	}
}

func TestHasAuth_WithUsername(t *testing.T) {
	c := &Config{Username: "admin", Password: "pass"}
	if !c.HasAuth() {
		t.Error("expected HasAuth() == true when username is set")
	}
}

func TestValidateAuth_NoAuth(t *testing.T) {
	c := &Config{}
	if err := c.ValidateAuth(); err == nil {
		t.Error("expected error from ValidateAuth when no auth configured")
	}
}

func TestValidateAuth_ValidUpdb(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", Username: "admin", Password: "secret"}
	if err := c.ValidateAuth(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_IdentityFileOnly(t *testing.T) {
	c := &Config{IdentityFile: "/some/file.json"}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_IdentityFileWithControllerURL(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", IdentityFile: "/some/file.json"}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_UpdbBothSet(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", Username: "admin", Password: "secret"}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_UpdbMissingPassword(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", Username: "admin"}
	if err := c.validate(); err == nil {
		t.Error("expected error when username set but password missing")
	}
}

func TestValidate_UpdbMissingUsername(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", Password: "secret"}
	if err := c.validate(); err == nil {
		t.Error("expected error when password set but username missing")
	}
}

func TestValidate_CertBothSet(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", CertFile: "/cert.pem", KeyFile: "/key.pem"}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_CertMissingKey(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", CertFile: "/cert.pem"}
	if err := c.validate(); err == nil {
		t.Error("expected error when cert set but key missing")
	}
}

func TestValidate_CertMissingCert(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", KeyFile: "/key.pem"}
	if err := c.validate(); err == nil {
		t.Error("expected error when key set but cert missing")
	}
}

func TestValidate_MultipleAuthMethods(t *testing.T) {
	c := &Config{
		ControllerURL: "https://ctrl:1280",
		IdentityFile:  "/some/file.json",
		Username:      "admin",
		Password:      "secret",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when multiple auth methods configured")
	}
}

func TestValidate_UpdbNoControllerURL(t *testing.T) {
	c := &Config{Username: "admin", Password: "secret"}
	if err := c.validate(); err == nil {
		t.Error("expected error when controller URL missing for updb auth")
	}
}

func TestValidate_IdentityFileNoControllerURL(t *testing.T) {
	// Identity file auth is allowed without controller URL (it may be in the file)
	c := &Config{IdentityFile: "/some/file.json"}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error (identity file without controller URL should be allowed): %v", err)
	}
}

// ext-jwt (static token) tests

func TestValidate_ExtJWTToken_Valid(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", ExtJWTToken: "eyJhbGci.fake.token"}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ExtJWTFile_Valid(t *testing.T) {
	c := &Config{ControllerURL: "https://ctrl:1280", ExtJWTFile: "/path/to/token.jwt"}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_ExtJWTBothTokenAndFile_Conflict(t *testing.T) {
	c := &Config{
		ControllerURL: "https://ctrl:1280",
		ExtJWTToken:   "eyJhbGci.fake.token",
		ExtJWTFile:    "/path/to/token.jwt",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when both --ext-jwt-token and --ext-jwt-file are set")
	}
}

func TestValidate_ExtJWTNoControllerURL(t *testing.T) {
	c := &Config{ExtJWTToken: "eyJhbGci.fake.token"}
	if err := c.validate(); err == nil {
		t.Error("expected error when controller URL missing for ext-jwt auth")
	}
}

func TestValidate_ExtJWTAndUpdb_Conflict(t *testing.T) {
	c := &Config{
		ControllerURL: "https://ctrl:1280",
		ExtJWTToken:   "eyJhbGci.fake.token",
		Username:      "admin",
		Password:      "secret",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when ext-jwt and updb both configured")
	}
}

// OIDC client credentials tests

func TestValidate_OIDCAllSet_Valid(t *testing.T) {
	c := &Config{
		ControllerURL:    "https://ctrl:1280",
		OIDCIssuer:       "https://idp.example.com",
		OIDCClientID:     "my-client",
		OIDCClientSecret: "my-secret",
	}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_OIDCWithAudienceAndTokenURL_Valid(t *testing.T) {
	c := &Config{
		ControllerURL:    "https://ctrl:1280",
		OIDCIssuer:       "https://idp.example.com",
		OIDCClientID:     "my-client",
		OIDCClientSecret: "my-secret",
		OIDCAudience:     "https://ctrl.example.com",
		OIDCTokenURL:     "https://idp.example.com/oauth/token",
	}
	if err := c.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_OIDCMissingClientID(t *testing.T) {
	c := &Config{
		ControllerURL:    "https://ctrl:1280",
		OIDCIssuer:       "https://idp.example.com",
		OIDCClientSecret: "my-secret",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when OIDC client ID missing")
	}
}

func TestValidate_OIDCMissingClientSecret(t *testing.T) {
	c := &Config{
		ControllerURL: "https://ctrl:1280",
		OIDCIssuer:    "https://idp.example.com",
		OIDCClientID:  "my-client",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when OIDC client secret missing")
	}
}

func TestValidate_OIDCMissingIssuer(t *testing.T) {
	c := &Config{
		ControllerURL:    "https://ctrl:1280",
		OIDCClientID:     "my-client",
		OIDCClientSecret: "my-secret",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when OIDC issuer missing")
	}
}

func TestValidate_OIDCNoControllerURL(t *testing.T) {
	c := &Config{
		OIDCIssuer:       "https://idp.example.com",
		OIDCClientID:     "my-client",
		OIDCClientSecret: "my-secret",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when controller URL missing for OIDC auth")
	}
}

func TestValidate_OIDCAndUpdb_Conflict(t *testing.T) {
	c := &Config{
		ControllerURL:    "https://ctrl:1280",
		OIDCIssuer:       "https://idp.example.com",
		OIDCClientID:     "my-client",
		OIDCClientSecret: "my-secret",
		Username:         "admin",
		Password:         "secret",
	}
	if err := c.validate(); err == nil {
		t.Error("expected error when OIDC and updb both configured")
	}
}
