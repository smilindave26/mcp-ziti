package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
)

// RegisterAll registers all Ziti management tools with the MCP server.
func RegisterAll(s *mcp.Server, zc *ziticlient.Client) {
	registerIdentityTools(s, zc)
	registerServiceTools(s, zc)
	registerServicePolicyTools(s, zc)
	registerEdgeRouterPolicyTools(s, zc)
	registerEdgeRouterTools(s, zc)
	registerNetworkTools(s, zc)
}
