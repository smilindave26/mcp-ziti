package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtPC "github.com/openziti/edge-api/rest_management_api_client/posture_checks"
)

func registerPostureCheckTools(s *mcp.Server, zc *ziticlient.Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-posture-checks",
		Description: "List posture checks (zero-trust device health requirements). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "posture checks", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtPC.NewListPostureChecksParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.PostureChecks.ListPostureChecks(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-posture-check",
		Description: "Get a single posture check by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "posture check", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.PostureChecks.DetailPostureCheck(
			mgmtPC.NewDetailPostureCheckParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-posture-check-types",
		Description: "List available posture check types (e.g. OS, domain, MFA, process).",
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "posture check types", func(ctx context.Context, mgmt *mgmtAPI, _ *string, limit, offset *int64) (any, error) {
		params := mgmtPC.NewListPostureCheckTypesParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		resp, err := mgmt.PostureChecks.ListPostureCheckTypes(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-posture-check",
		Description: "Permanently delete a posture check by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "posture check", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.PostureChecks.DeletePostureCheck(
			mgmtPC.NewDeletePostureCheckParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}
