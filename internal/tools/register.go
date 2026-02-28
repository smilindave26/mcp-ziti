package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

// RegisterAll registers all Ziti management tools with the MCP server.
func RegisterAll(s *mcp.Server, zc *ziticlient.Client) {
	registerConnectionTools(s, zc)
	registerIdentityTools(s, zc)
	registerServiceTools(s, zc)
	registerServicePolicyTools(s, zc)
	registerEdgeRouterPolicyTools(s, zc)
	registerEdgeRouterTools(s, zc)
	registerNetworkTools(s, zc)
	registerAuthenticatorTools(s, zc)
	registerEnrollmentTools(s, zc)
	registerCertificateAuthorityTools(s, zc)
	registerExternalJWTSignerTools(s, zc)
	registerAuthPolicyTools(s, zc)
	registerConfigTools(s, zc)
	registerPostureCheckTools(s, zc)
	registerTerminatorTools(s, zc)
	registerServiceEdgeRouterPolicyTools(s, zc)
	registerAPISessionTools(s, zc)
	registerSessionTools(s, zc)
	registerRouterTools(s, zc)
	registerDatabaseTools(s, zc)
	registerControllerTools(s, zc)
	registerRoleAttributeTools(s, zc)
}
