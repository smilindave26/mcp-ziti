package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	fabricRaft "github.com/openziti/fabric/controller/rest_client/raft"
)

func registerClusterTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &clusterTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-cluster-members",
		Description: "List raft cluster members (controller nodes). Shows address, leader status, connected state, version, and voter status. Only available if the controller exposes the fabric API.",
		Annotations: readOnlyAnnotation,
	}, t.listMembers)
}

type clusterTools struct{ zc *ziticlient.Client }

func (t *clusterTools) listMembers(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	fabric, err := t.zc.Fabric()
	if err != nil {
		return nil, nil, err
	}

	resp, err := fabric.Raft.RaftListMembers(
		fabricRaft.NewRaftListMembersParams().WithContext(ctx))
	if err != nil {
		return nil, nil, fmt.Errorf("list cluster members: %w", err)
	}
	return jsonResult(resp.GetPayload().Values)
}
