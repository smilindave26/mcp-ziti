# Tools

All list tools accept `filter`, `limit` (default 100, max 500), and `offset` parameters for filtering and pagination.

## Connection

| Tool | Description |
|------|-------------|
| `connect-controller` | Connect (or reconnect) to a Ziti controller at runtime |
| `disconnect-controller` | Disconnect from the current controller and clear credentials |
| `get-controller-status` | Get the current connection status and controller URL |
| `start-oidc-login` | Start an interactive OIDC login via browser (Device Authorization Grant) |
| `complete-oidc-login` | Complete the OIDC device login and connect to the controller |

## Identities

| Tool | Description |
|------|-------------|
| `list-identities` | List identities with optional filter and pagination |
| `get-identity` | Get a single identity by ID |
| `create-identity` | Create a new identity (Device, User, Router, or Service) |
| `update-identity` | Rename, change type, toggle admin flag, or update role attributes |
| `delete-identity` | Permanently delete an identity |

## Services

| Tool | Description |
|------|-------------|
| `list-services` | List services with optional filter and pagination |
| `get-service` | Get a single service by ID |
| `create-service` | Create a new service |
| `update-service` | Update a service's name, encryption setting, or role attributes |
| `delete-service` | Permanently delete a service |

## Service Policies

| Tool | Description |
|------|-------------|
| `list-service-policies` | List service policies |
| `get-service-policy` | Get a single service policy by ID |
| `create-service-policy` | Create a Dial or Bind service policy |
| `update-service-policy` | Update a service policy |
| `delete-service-policy` | Permanently delete a service policy |

## Edge Router Policies

| Tool | Description |
|------|-------------|
| `list-edge-router-policies` | List edge router policies |
| `get-edge-router-policy` | Get a single edge router policy by ID |
| `create-edge-router-policy` | Create an edge router policy |
| `delete-edge-router-policy` | Permanently delete an edge router policy |

## Service Edge Router Policies

| Tool | Description |
|------|-------------|
| `list-service-edge-router-policies` | List service edge router policies |
| `get-service-edge-router-policy` | Get a single service edge router policy by ID |
| `create-service-edge-router-policy` | Create a service edge router policy |
| `update-service-edge-router-policy` | Update a service edge router policy |
| `delete-service-edge-router-policy` | Permanently delete a service edge router policy |

## Edge Routers

| Tool | Description |
|------|-------------|
| `list-edge-routers` | List edge routers |
| `get-edge-router` | Get a single edge router by ID |

## Routers

| Tool | Description |
|------|-------------|
| `list-routers` | List fabric routers (edge and non-edge) |
| `get-router` | Get a single fabric router by ID |

## Authenticators

| Tool | Description |
|------|-------------|
| `list-authenticators` | List authenticators (credentials attached to identities) |
| `get-authenticator` | Get a single authenticator by ID |
| `update-authenticator` | Update username/password for an updb authenticator |
| `delete-authenticator` | Permanently delete an authenticator |

## Enrollments

| Tool | Description |
|------|-------------|
| `list-enrollments` | List pending enrollments |
| `get-enrollment` | Get a single enrollment by ID |
| `create-enrollment` | Create a new enrollment (ott, ottca, or updb) |
| `delete-enrollment` | Delete a pending enrollment |

## Certificate Authorities

| Tool | Description |
|------|-------------|
| `list-certificate-authorities` | List CAs |
| `get-certificate-authority` | Get a single CA by ID |
| `create-certificate-authority` | Create a new CA |
| `update-certificate-authority` | Update a CA |
| `delete-certificate-authority` | Permanently delete a CA |

## External JWT Signers

| Tool | Description |
|------|-------------|
| `list-external-jwt-signers` | List external JWT signers |
| `get-external-jwt-signer` | Get a single external JWT signer by ID |
| `create-external-jwt-signer` | Create an external JWT signer (cert or JWKS) |
| `update-external-jwt-signer` | Update an external JWT signer |
| `delete-external-jwt-signer` | Permanently delete an external JWT signer |

## Auth Policies

| Tool | Description |
|------|-------------|
| `list-auth-policies` | List authentication policies |
| `get-auth-policy` | Get a single auth policy by ID |
| `create-auth-policy` | Create an authentication policy |
| `update-auth-policy` | Update an authentication policy |
| `delete-auth-policy` | Permanently delete an auth policy |

## Configs

| Tool | Description |
|------|-------------|
| `list-config-types` | List service config types |
| `get-config-type` | Get a single config type by ID |
| `create-config-type` | Create a new config type |
| `delete-config-type` | Permanently delete a config type |
| `list-configs` | List service configurations |
| `get-config` | Get a single configuration by ID |
| `create-config` | Create a new service configuration |
| `update-config` | Update a service configuration |
| `delete-config` | Permanently delete a service configuration |

## Posture Checks

| Tool | Description |
|------|-------------|
| `list-posture-checks` | List posture checks |
| `get-posture-check` | Get a single posture check by ID |
| `list-posture-check-types` | List available posture check types |
| `delete-posture-check` | Permanently delete a posture check |

## Terminators

| Tool | Description |
|------|-------------|
| `list-terminators` | List terminators |
| `get-terminator` | Get a single terminator by ID |
| `create-terminator` | Create a terminator linking a service to a router address |
| `delete-terminator` | Permanently delete a terminator |

## Sessions

| Tool | Description |
|------|-------------|
| `list-api-sessions` | List active management API sessions |
| `get-api-session` | Get a single API session by ID |
| `delete-api-session` | Force-delete an API session |
| `list-sessions` | List active network (data-plane) sessions |
| `get-session` | Get a single network session by ID |
| `delete-session` | Terminate a network session |

## Role Attributes

| Tool | Description |
|------|-------------|
| `list-identity-role-attributes` | List role attributes in use on identities |
| `list-edge-router-role-attributes` | List role attributes in use on edge routers |
| `list-service-role-attributes` | List role attributes in use on services |
| `list-posture-check-role-attributes` | List role attributes in use on posture checks |

## Network

| Tool | Description |
|------|-------------|
| `get-controller-version` | Get controller version and build info |
| `list-summary` | Get a resource count summary for the whole network |
| `list-controllers` | List controllers in an HA cluster |

## Database

| Tool | Description |
|------|-------------|
| `create-database-snapshot` | Trigger an immediate database backup snapshot |
| `check-data-integrity` | Run an integrity check on the controller database |
| `fix-data-integrity` | Attempt to fix data integrity issues automatically |
