package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtAPISession "github.com/openziti/edge-api/rest_management_api_client/api_session"
)

func registerAPISessionTools(s *mcp.Server, zc *ziticlient.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-api-sessions",
		Description: "List active API sessions (authenticated management connections). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "api sessions", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtAPISession.NewListAPISessionsParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.APISession.ListAPISessions(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-api-session",
		Description: "Get a single API session by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "api session", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.APISession.DetailAPISessions(
			mgmtAPISession.NewDetailAPISessionsParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-api-session",
		Description: "Delete (force-logout) an API session by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, makeDeleteHandler(zc, "api session", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.APISession.DeleteAPISessions(
			mgmtAPISession.NewDeleteAPISessionsParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}
