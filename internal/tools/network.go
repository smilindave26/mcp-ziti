package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtInfo "github.com/openziti/edge-api/rest_management_api_client/informational"
)

func registerNetworkTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &networkTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-controller-version",
		Description: "Get the Ziti controller version and build information.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.version)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-summary",
		Description: "Get a count summary of all resource types in the Ziti network (identities, services, policies, edge routers).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.summary)
}

type networkTools struct{ zc *ziticlient.Client }

func (t *networkTools) version(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtInfo.NewListVersionParams().WithContext(ctx)
	resp, err := mgmt.Informational.ListVersion(params)
	if err != nil {
		return nil, nil, fmt.Errorf("get controller version: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type summaryResult struct {
	Identities        int64 `json:"identities"`
	Services          int64 `json:"services"`
	ServicePolicies   int64 `json:"servicePolicies"`
	EdgeRouterPolicies int64 `json:"edgeRouterPolicies"`
	EdgeRouters       int64 `json:"edgeRouters"`
}

func (t *networkTools) summary(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	zero := int64(0)
	one := int64(1)

	result := summaryResult{}

	// Count each resource type by fetching limit=1 and reading the total
	identResp, err := mgmt.Identity.ListIdentities(
		newCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting identities: %w", err)
	}
	if identResp.GetPayload().Meta.Pagination != nil {
		result.Identities = *identResp.GetPayload().Meta.Pagination.TotalCount
	}

	svcResp, err := mgmt.Service.ListServices(
		newServiceCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting services: %w", err)
	}
	if svcResp.GetPayload().Meta.Pagination != nil {
		result.Services = *svcResp.GetPayload().Meta.Pagination.TotalCount
	}

	spResp, err := mgmt.ServicePolicy.ListServicePolicies(
		newSPCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting service policies: %w", err)
	}
	if spResp.GetPayload().Meta.Pagination != nil {
		result.ServicePolicies = *spResp.GetPayload().Meta.Pagination.TotalCount
	}

	erpResp, err := mgmt.EdgeRouterPolicy.ListEdgeRouterPolicies(
		newERPCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting edge router policies: %w", err)
	}
	if erpResp.GetPayload().Meta.Pagination != nil {
		result.EdgeRouterPolicies = *erpResp.GetPayload().Meta.Pagination.TotalCount
	}

	erResp, err := mgmt.EdgeRouter.ListEdgeRouters(
		newERCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting edge routers: %w", err)
	}
	if erResp.GetPayload().Meta.Pagination != nil {
		result.EdgeRouters = *erResp.GetPayload().Meta.Pagination.TotalCount
	}

	return jsonResult(result)
}
