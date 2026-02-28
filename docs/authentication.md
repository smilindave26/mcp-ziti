# Authentication

Credentials are **optional at startup**. If no authentication is configured, the server starts in a disconnected state and the AI agent can connect later using the `connect-controller` tool. When credentials are provided, exactly one authentication method must be used. All options are available as CLI flags or environment variables (flags take precedence).

## Identity JSON file

A Ziti identity file contains the controller URL, client certificate, and CA bundle in a single JSON file. This is the simplest option when you already have an enrolled identity.

!!! tip "How to get an identity file"
    In your [OpenZiti](https://openziti.io) network, create an identity and download its one-time enrollment JWT, then enroll it with the `ziti` CLI to produce the JSON file:

    ```bash
    # Create an identity via the CLI or web console, then enroll with the JWT
    ziti edge enroll --jwt /path/to/identity.jwt --out /path/to/identity.json
    ```

    The resulting `identity.json` is what you pass to `--identity-file`. See the [OpenZiti enrollment docs](https://openziti.io/docs/learn/core-concepts/identities/enrolling) for full details.

```bash
# The controller URL is read from the ztAPI field inside the file
ziti-mcp --identity-file /path/to/identity.json

# Override the controller URL if needed
ziti-mcp --identity-file /path/to/identity.json \
          --controller https://ctrl.example.com:1280
```

Environment variables:
```bash
ZITI_IDENTITY_FILE=/path/to/identity.json ziti-mcp
```

## Username / password

Authenticates with the controller's built-in user database (updb). Requires `--controller`.

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --username admin \
          --password secret
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_USERNAME=admin \
ZITI_PASSWORD=secret \
ziti-mcp
```

## Client certificate

Authenticates using a TLS client certificate and private key. Requires `--controller`.

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --cert /path/to/client.crt \
          --key  /path/to/client.key
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_CERT_FILE=/path/to/client.crt \
ZITI_KEY_FILE=/path/to/client.key \
ziti-mcp
```

## External JWT (static token)

Authenticates using a pre-issued JWT — for example a service account token from your IdP or a long-lived API token issued by the controller. Requires `--controller`.

Provide the token directly as a string:

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --ext-jwt-token eyJhbGciOiJSUzI1NiJ9...
```

Or point to a file containing the token (useful for Kubernetes-mounted secrets):

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --ext-jwt-file /var/run/secrets/token.jwt
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_EXT_JWT_TOKEN=eyJhbGciOiJSUzI1NiJ9... \
ziti-mcp
```

## OIDC client credentials

Authenticates using the [OAuth 2.0 client credentials flow](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4). A fresh token is fetched from the IdP on each session, so no manual token rotation is needed. Requires `--controller` and `--oidc-issuer`.

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --oidc-issuer https://idp.example.com \
          --oidc-client-id my-client \
          --oidc-client-secret my-secret
```

Optional extras:

```bash
# Restrict the token audience
ziti-mcp ... --oidc-audience https://ctrl.example.com

# Skip OIDC discovery and use a known token endpoint directly
ziti-mcp ... --oidc-token-url https://idp.example.com/oauth/token
```

Environment variables:
```bash
ZITI_CONTROLLER_URL=https://ctrl.example.com:1280 \
ZITI_OIDC_ISSUER=https://idp.example.com \
ZITI_OIDC_CLIENT_ID=my-client \
ZITI_OIDC_CLIENT_SECRET=my-secret \
ziti-mcp
```

## Interactive OIDC login (browser)

For users configured via a 3rd-party identity provider who don't have a client secret, the AI agent can initiate an interactive browser login using the [OAuth 2.0 Device Authorization Grant (RFC 8628)](https://datatracker.ietf.org/doc/html/rfc8628).

**When to use this:** Your Ziti network uses an external IdP (e.g. Auth0, Okta, Keycloak) for user authentication, and you want to log in interactively rather than providing a client secret.

**Two-step flow:**

1. The agent calls `start-oidc-login` (with `controllerUrl`, `oidcIssuer`, and `oidcClientId`, or these can come from startup defaults)
2. The tool returns a verification URL and a user code
3. The agent presents the URL and code to the user, who opens the URL in their browser and enters the code
4. The user authenticates with the IdP in the browser
5. The agent calls `complete-oidc-login` — the tool polls the IdP token endpoint until authentication completes, then connects

**Pre-configured defaults:** You can set `--controller`, `--oidc-issuer`, `--oidc-client-id`, and `--oidc-audience` at startup without `--oidc-client-secret`. The server starts disconnected, and those values are used as defaults for `start-oidc-login` so the agent doesn't need to provide them:

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --oidc-issuer https://idp.example.com \
          --oidc-client-id my-public-client
```

!!! note
    The IdP must support the Device Authorization Grant flow. Auth0, Okta, and Keycloak all support it. No redirect URIs need to be configured.

## Optional CA override

By default the server fetches the controller's CA bundle from its well-known endpoint. To use a custom CA instead, add `--ca` (or `ZITI_CA_FILE`) to any of the methods above:

```bash
ziti-mcp --controller https://ctrl.example.com:1280 \
          --username admin --password secret \
          --ca /path/to/ca-bundle.pem
```

## Flag reference

| Flag | Env var | Description |
|------|---------|-------------|
| `--controller` | `ZITI_CONTROLLER_URL` | Controller URL, e.g. `https://ctrl.example.com:1280` |
| `--identity-file` | `ZITI_IDENTITY_FILE` | Path to a Ziti identity JSON file |
| `--username` | `ZITI_USERNAME` | Username for updb authentication |
| `--password` | `ZITI_PASSWORD` | Password for updb authentication |
| `--cert` | `ZITI_CERT_FILE` | Path to a PEM client certificate file |
| `--key` | `ZITI_KEY_FILE` | Path to a PEM private key file |
| `--ca` | `ZITI_CA_FILE` | Path to a PEM CA bundle (optional override) |
| `--ext-jwt-token` | `ZITI_EXT_JWT_TOKEN` | External JWT token string |
| `--ext-jwt-file` | `ZITI_EXT_JWT_FILE` | Path to a file containing an external JWT |
| `--oidc-issuer` | `ZITI_OIDC_ISSUER` | OIDC issuer URL |
| `--oidc-client-id` | `ZITI_OIDC_CLIENT_ID` | OIDC client ID |
| `--oidc-client-secret` | `ZITI_OIDC_CLIENT_SECRET` | OIDC client secret (required for client credentials flow, omit for interactive login) |
| `--oidc-audience` | `ZITI_OIDC_AUDIENCE` | OIDC audience claim (optional) |
| `--oidc-token-url` | `ZITI_OIDC_TOKEN_URL` | OIDC token endpoint URL — skips discovery (optional) |
