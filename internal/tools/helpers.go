package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtClient "github.com/openziti/edge-api/rest_management_api_client"
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

// mgmtAPI is a type alias so tool files can reference the management client
// type without importing the package themselves.
type mgmtAPI = mgmtClient.ZitiEdgeManagement

// Shared input types for standard CRUD operations.

type listInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression, e.g. name contains 'test'"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

type getInput struct {
	ID string `json:"id" jsonschema:"required,resource ID"`
}

type deleteInput struct {
	ID string `json:"id" jsonschema:"required,resource ID to delete"`
}

// Common tool annotations.
var (
	readOnlyAnnotation    = &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true}
	destructiveAnnotation = &mcp.ToolAnnotations{DestructiveHint: boolPtr(true), IdempotentHint: true}
)

func boolPtr(b bool) *bool { return &b }

// CRUD handler function types.
type (
	listFunc   func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error)
	getFunc    func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error)
	deleteFunc func(ctx context.Context, mgmt *mgmtAPI, id string) error
)

// makeListHandler creates a standard list tool handler.
func makeListHandler(zc *ziticlient.Client, name string, fn listFunc) func(context.Context, *mcp.CallToolRequest, listInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in listInput) (*mcp.CallToolResult, any, error) {
		mgmt, err := zc.Mgmt()
		if err != nil {
			return nil, nil, err
		}
		limit := clampLimit(in.Limit)
		offset := in.Offset
		var filter *string
		if in.Filter != "" {
			filter = &in.Filter
		}
		data, err := fn(ctx, mgmt, filter, &limit, &offset)
		if err != nil {
			return nil, nil, fmt.Errorf("list %s: %w", name, err)
		}
		return jsonResult(data)
	}
}

// makeGetHandler creates a standard get-by-ID tool handler.
func makeGetHandler(zc *ziticlient.Client, name string, fn getFunc) func(context.Context, *mcp.CallToolRequest, getInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in getInput) (*mcp.CallToolResult, any, error) {
		mgmt, err := zc.Mgmt()
		if err != nil {
			return nil, nil, err
		}
		data, err := fn(ctx, mgmt, in.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get %s %q: %w", name, in.ID, err)
		}
		return jsonResult(data)
	}
}

// makeDeleteHandler creates a standard delete-by-ID tool handler.
func makeDeleteHandler(zc *ziticlient.Client, name string, fn deleteFunc) func(context.Context, *mcp.CallToolRequest, deleteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in deleteInput) (*mcp.CallToolResult, any, error) {
		mgmt, err := zc.Mgmt()
		if err != nil {
			return nil, nil, err
		}
		if err := fn(ctx, mgmt, in.ID); err != nil {
			return nil, nil, fmt.Errorf("delete %s %q: %w", name, in.ID, err)
		}
		return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
	}
}

// stripFields lists top-level and nested keys removed from every API response
// to reduce token usage. These fields add no value for LLM-driven management.
var stripFields = map[string]bool{
	"_links":  true, // HAL hypermedia URLs
	"envInfo": true, // OS/arch on identities
	"sdkInfo": true, // SDK version/build info on identities
}

// stripDenylistKeys recursively removes denylist keys from maps and slices.
func stripDenylistKeys(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k := range val {
			if stripFields[k] {
				delete(val, k)
			} else {
				val[k] = stripDenylistKeys(val[k])
			}
		}
		return val
	case []any:
		for i, item := range val {
			val[i] = stripDenylistKeys(item)
		}
		return val
	default:
		return v
	}
}

// jsonResult marshals v to JSON, strips noisy API fields, and returns it as a
// text content tool result. The Out value is nil so the go-sdk does not set
// StructuredContent (which must be a JSON object per the MCP spec; returning
// arrays or scalars there breaks clients that validate the response).
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}

	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, fmt.Errorf("unmarshal for stripping: %w", err)
	}
	raw = stripDenylistKeys(raw)

	b, err = json.Marshal(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("re-marshal result: %w", err)
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
