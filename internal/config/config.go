package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all runtime configuration for the MCP server.
type Config struct {
	ControllerURL string // --controller / ZITI_CONTROLLER_URL
	IdentityFile  string // --identity-file / ZITI_IDENTITY_FILE
	Username      string // --username / ZITI_USERNAME
	Password      string // --password / ZITI_PASSWORD
	CertFile      string // --cert / ZITI_CERT_FILE
	KeyFile       string // --key / ZITI_KEY_FILE
	CAFile        string // --ca / ZITI_CA_FILE (optional CA bundle override)

	// External JWT (static token or file)
	ExtJWTToken string // --ext-jwt-token / ZITI_EXT_JWT_TOKEN
	ExtJWTFile  string // --ext-jwt-file  / ZITI_EXT_JWT_FILE

	// OIDC client credentials
	OIDCIssuer       string // --oidc-issuer        / ZITI_OIDC_ISSUER
	OIDCClientID     string // --oidc-client-id     / ZITI_OIDC_CLIENT_ID
	OIDCClientSecret string // --oidc-client-secret / ZITI_OIDC_CLIENT_SECRET
	OIDCAudience     string // --oidc-audience      / ZITI_OIDC_AUDIENCE     (optional)
	OIDCTokenURL     string // --oidc-token-url     / ZITI_OIDC_TOKEN_URL    (optional override, skips discovery)
}

// Load parses CLI flags then falls back to environment variables for any unset
// fields. CLI flags always take precedence.
func Load() (*Config, error) {
	c := &Config{}

	flag.StringVar(&c.ControllerURL, "controller", "", "Controller URL, e.g. https://ctrl.example.com:1280 (env: ZITI_CONTROLLER_URL)")
	flag.StringVar(&c.IdentityFile, "identity-file", "", "Path to Ziti identity JSON file (env: ZITI_IDENTITY_FILE)")
	flag.StringVar(&c.Username, "username", "", "Username for updb authentication (env: ZITI_USERNAME)")
	flag.StringVar(&c.Password, "password", "", "Password for updb authentication (env: ZITI_PASSWORD)")
	flag.StringVar(&c.CertFile, "cert", "", "Path to client certificate PEM file (env: ZITI_CERT_FILE)")
	flag.StringVar(&c.KeyFile, "key", "", "Path to client private key PEM file (env: ZITI_KEY_FILE)")
	flag.StringVar(&c.CAFile, "ca", "", "Path to CA bundle PEM file — optional override (env: ZITI_CA_FILE)")
	flag.StringVar(&c.ExtJWTToken, "ext-jwt-token", "", "External JWT token string (env: ZITI_EXT_JWT_TOKEN)")
	flag.StringVar(&c.ExtJWTFile, "ext-jwt-file", "", "Path to file containing an external JWT (env: ZITI_EXT_JWT_FILE)")
	flag.StringVar(&c.OIDCIssuer, "oidc-issuer", "", "OIDC issuer URL for client credentials flow (env: ZITI_OIDC_ISSUER)")
	flag.StringVar(&c.OIDCClientID, "oidc-client-id", "", "OIDC client ID (env: ZITI_OIDC_CLIENT_ID)")
	flag.StringVar(&c.OIDCClientSecret, "oidc-client-secret", "", "OIDC client secret (env: ZITI_OIDC_CLIENT_SECRET)")
	flag.StringVar(&c.OIDCAudience, "oidc-audience", "", "OIDC audience claim — optional (env: ZITI_OIDC_AUDIENCE)")
	flag.StringVar(&c.OIDCTokenURL, "oidc-token-url", "", "OIDC token endpoint URL — optional, skips discovery (env: ZITI_OIDC_TOKEN_URL)")
	flag.Parse()

	// Fall back to env vars for any flag not explicitly set
	if c.ControllerURL == "" {
		c.ControllerURL = os.Getenv("ZITI_CONTROLLER_URL")
	}
	if c.IdentityFile == "" {
		c.IdentityFile = os.Getenv("ZITI_IDENTITY_FILE")
	}
	if c.Username == "" {
		c.Username = os.Getenv("ZITI_USERNAME")
	}
	if c.Password == "" {
		c.Password = os.Getenv("ZITI_PASSWORD")
	}
	if c.CertFile == "" {
		c.CertFile = os.Getenv("ZITI_CERT_FILE")
	}
	if c.KeyFile == "" {
		c.KeyFile = os.Getenv("ZITI_KEY_FILE")
	}
	if c.CAFile == "" {
		c.CAFile = os.Getenv("ZITI_CA_FILE")
	}
	if c.ExtJWTToken == "" {
		c.ExtJWTToken = os.Getenv("ZITI_EXT_JWT_TOKEN")
	}
	if c.ExtJWTFile == "" {
		c.ExtJWTFile = os.Getenv("ZITI_EXT_JWT_FILE")
	}
	if c.OIDCIssuer == "" {
		c.OIDCIssuer = os.Getenv("ZITI_OIDC_ISSUER")
	}
	if c.OIDCClientID == "" {
		c.OIDCClientID = os.Getenv("ZITI_OIDC_CLIENT_ID")
	}
	if c.OIDCClientSecret == "" {
		c.OIDCClientSecret = os.Getenv("ZITI_OIDC_CLIENT_SECRET")
	}
	if c.OIDCAudience == "" {
		c.OIDCAudience = os.Getenv("ZITI_OIDC_AUDIENCE")
	}
	if c.OIDCTokenURL == "" {
		c.OIDCTokenURL = os.Getenv("ZITI_OIDC_TOKEN_URL")
	}

	return c, c.validate()
}

func (c *Config) validate() error {
	hasIdentity := c.IdentityFile != ""
	hasCreds := c.Username != "" || c.Password != ""
	hasCert := c.CertFile != "" || c.KeyFile != ""
	hasExtJWT := c.ExtJWTToken != "" || c.ExtJWTFile != ""
	hasOIDC := c.OIDCIssuer != "" || c.OIDCClientID != "" || c.OIDCClientSecret != ""

	count := 0
	if hasIdentity {
		count++
	}
	if hasCreds {
		count++
	}
	if hasCert {
		count++
	}
	if hasExtJWT {
		count++
	}
	if hasOIDC {
		count++
	}

	if count == 0 {
		// Zero auth is allowed — the LLM can connect at runtime via connect-controller.
		return nil
	}
	if count > 1 {
		return fmt.Errorf("multiple authentication methods configured: use exactly one of --identity-file, --username/--password, --cert/--key, --ext-jwt-token/--ext-jwt-file, or --oidc-issuer/--oidc-client-id/--oidc-client-secret")
	}

	if hasCreds {
		if c.Username == "" {
			return fmt.Errorf("--password requires --username")
		}
		if c.Password == "" {
			return fmt.Errorf("--username requires --password")
		}
	}

	if hasCert {
		if c.CertFile == "" {
			return fmt.Errorf("--key requires --cert")
		}
		if c.KeyFile == "" {
			return fmt.Errorf("--cert requires --key")
		}
	}

	if hasExtJWT {
		if c.ExtJWTToken != "" && c.ExtJWTFile != "" {
			return fmt.Errorf("--ext-jwt-token and --ext-jwt-file are mutually exclusive")
		}
	}

	if hasOIDC {
		if c.OIDCIssuer == "" {
			return fmt.Errorf("--oidc-client-id/--oidc-client-secret require --oidc-issuer")
		}
		if c.OIDCClientID == "" {
			return fmt.Errorf("--oidc-issuer requires --oidc-client-id")
		}
		if c.OIDCClientSecret == "" {
			return fmt.Errorf("--oidc-issuer requires --oidc-client-secret")
		}
	}

	// ControllerURL is required for all methods except identity file (which embeds it)
	if c.ControllerURL == "" && !hasIdentity {
		return fmt.Errorf("--controller is required when not using --identity-file")
	}

	return nil
}

// HasAuth returns true if at least one authentication method is configured.
func (c *Config) HasAuth() bool {
	return c.IdentityFile != "" ||
		c.Username != "" || c.Password != "" ||
		c.CertFile != "" || c.KeyFile != "" ||
		c.ExtJWTToken != "" || c.ExtJWTFile != "" ||
		c.OIDCIssuer != "" || c.OIDCClientID != "" || c.OIDCClientSecret != ""
}

// ValidateAuth validates auth fields, requiring that at least one auth method is present.
// Used by connect-controller to validate runtime-supplied credentials.
func (c *Config) ValidateAuth() error {
	if !c.HasAuth() {
		return fmt.Errorf("no authentication configured: provide identity file, username/password, cert/key, ext-jwt-token/ext-jwt-file, or oidc credentials")
	}
	return c.validate()
}
