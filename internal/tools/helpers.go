package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mgmtER "github.com/openziti/edge-api/rest_management_api_client/edge_router"
	mgmtERP "github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	mgmtIdentity "github.com/openziti/edge-api/rest_management_api_client/identity"
	mgmtService "github.com/openziti/edge-api/rest_management_api_client/service"
	mgmtServicePolicy "github.com/openziti/edge-api/rest_management_api_client/service_policy"
)

const (
	defaultLimit int64 = 100
	maxLimit     int64 = 500
)

// jsonResult marshals v to JSON and returns it as a text content tool result.
// The Out value is nil so the go-sdk does not set StructuredContent (which must
// be a JSON object per the MCP spec; returning arrays or scalars there breaks
// clients that validate the response).
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}, nil, nil
}

// clampLimit returns the given limit clamped to [1, maxLimit].
// If limit is 0 (unset), it returns defaultLimit.
func clampLimit(limit int64) int64 {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

// Count-only param helpers used by the summary tool. They request limit=1 and
// offset=0 so we only need the pagination metadata (totalCount), not the items.

func newCountParams(ctx context.Context, offset, limit *int64) *mgmtIdentity.ListIdentitiesParams {
	return mgmtIdentity.NewListIdentitiesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
}

func newServiceCountParams(ctx context.Context, offset, limit *int64) *mgmtService.ListServicesParams {
	return mgmtService.NewListServicesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
}

func newSPCountParams(ctx context.Context, offset, limit *int64) *mgmtServicePolicy.ListServicePoliciesParams {
	return mgmtServicePolicy.NewListServicePoliciesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
}

func newERPCountParams(ctx context.Context, offset, limit *int64) *mgmtERP.ListEdgeRouterPoliciesParams {
	return mgmtERP.NewListEdgeRouterPoliciesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
}

func newERCountParams(ctx context.Context, offset, limit *int64) *mgmtER.ListEdgeRoutersParams {
	return mgmtER.NewListEdgeRoutersParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
}
