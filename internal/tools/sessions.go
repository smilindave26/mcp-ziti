package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtSession "github.com/openziti/edge-api/rest_management_api_client/session"
)

func registerSessionTools(s *mcp.Server, zc *ziticlient.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-sessions",
		Description: "List active network sessions (data-plane connections). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "sessions", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtSession.NewListSessionsParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Session.ListSessions(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-session",
		Description: "Get a single network session by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "session", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Session.DetailSession(
			mgmtSession.NewDetailSessionParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-session",
		Description: "Terminate a network session by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: boolPtr(true)},
	}, makeDeleteHandler(zc, "session", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.Session.DeleteSession(
			mgmtSession.NewDeleteSessionParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}
