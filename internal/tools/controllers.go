package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtCtrl "github.com/openziti/edge-api/rest_management_api_client/controllers"
)

func registerControllerTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &controllerTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-controllers",
		Description: "List controllers in an HA (high-availability) cluster.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)
}

type controllerTools struct{ zc *ziticlient.Client }

type listControllersInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *controllerTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listControllersInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtCtrl.NewListControllersParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Controllers.ListControllers(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list controllers: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}
