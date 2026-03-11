package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	fabricCircuit "github.com/openziti/fabric/controller/rest_client/circuit"
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
		Description: "Get a count summary of all resource types in the Ziti network. Includes circuits if the fabric API is available.",
		Annotations: readOnlyAnnotation,
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
	Identities                 int64 `json:"identities"`
	Services                   int64 `json:"services"`
	ServicePolicies            int64 `json:"servicePolicies"`
	EdgeRouterPolicies         int64 `json:"edgeRouterPolicies"`
	ServiceEdgeRouterPolicies  int64 `json:"serviceEdgeRouterPolicies"`
	EdgeRouters                int64 `json:"edgeRouters"`
	Terminators                int64 `json:"terminators"`
	Configs                    int64 `json:"configs"`
	Circuits                   *int  `json:"circuits,omitempty"`
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

	serpResp, err := mgmt.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(
		newSERPCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting service edge router policies: %w", err)
	}
	if serpResp.GetPayload().Meta.Pagination != nil {
		result.ServiceEdgeRouterPolicies = *serpResp.GetPayload().Meta.Pagination.TotalCount
	}

	erResp, err := mgmt.EdgeRouter.ListEdgeRouters(
		newERCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting edge routers: %w", err)
	}
	if erResp.GetPayload().Meta.Pagination != nil {
		result.EdgeRouters = *erResp.GetPayload().Meta.Pagination.TotalCount
	}

	termResp, err := mgmt.Terminator.ListTerminators(
		newTermCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting terminators: %w", err)
	}
	if termResp.GetPayload().Meta.Pagination != nil {
		result.Terminators = *termResp.GetPayload().Meta.Pagination.TotalCount
	}

	cfgResp, err := mgmt.Config.ListConfigs(
		newConfigCountParams(ctx, &zero, &one), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("counting configs: %w", err)
	}
	if cfgResp.GetPayload().Meta.Pagination != nil {
		result.Configs = *cfgResp.GetPayload().Meta.Pagination.TotalCount
	}

	// Include circuit count if the fabric API is available.
	fabric, err := t.zc.Fabric()
	if err == nil {
		circResp, err := fabric.Circuit.ListCircuits(
			fabricCircuit.NewListCircuitsParams().WithContext(ctx))
		if err == nil {
			count := len(circResp.GetPayload().Data)
			result.Circuits = &count
		}
	} else if !errors.Is(err, ziticlient.ErrFabricNotAvailable) {
		return nil, nil, err
	}

	return jsonResult(result)
}
