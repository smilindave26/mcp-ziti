package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtTerm "github.com/openziti/edge-api/rest_management_api_client/terminator"
	"github.com/openziti/edge-api/rest_model"
)

func registerTerminatorTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &terminatorTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-terminators",
		Description: "List terminators (service endpoints hosted by SDK applications or routers). Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-terminator",
		Description: "Get a single terminator by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-terminator",
		Description: "Create a terminator linking a service to a router address. binding is typically 'transport'. precedence is one of: default, required, failed.",
	}, t.create)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-terminator",
		Description: "Permanently delete a terminator by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type terminatorTools struct{ zc *ziticlient.Client }

type listTerminatorsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *terminatorTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listTerminatorsInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtTerm.NewListTerminatorsParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Terminator.ListTerminators(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list terminators: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getTerminatorInput struct {
	ID string `json:"id" jsonschema:"required,terminator ID"`
}

func (t *terminatorTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getTerminatorInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtTerm.NewDetailTerminatorParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Terminator.DetailTerminator(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get terminator %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type createTerminatorInput struct {
	ServiceID  string `json:"serviceId"            jsonschema:"required,service ID this terminator hosts"`
	RouterID   string `json:"routerId"             jsonschema:"required,edge router ID hosting the terminator"`
	Binding    string `json:"binding"              jsonschema:"required,binding type, typically 'transport'"`
	Address    string `json:"address"              jsonschema:"required,address the router dials to reach the service"`
	Precedence string `json:"precedence,omitempty" jsonschema:"terminator precedence: default, required, or failed"`
	Cost       int64  `json:"cost,omitempty"       jsonschema:"terminator cost (0-65535, lower is preferred)"`
}

func (t *terminatorTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createTerminatorInput) (*mcp.CallToolResult, any, error) {
	if in.ServiceID == "" {
		return nil, nil, fmt.Errorf("serviceId is required")
	}
	if in.RouterID == "" {
		return nil, nil, fmt.Errorf("routerId is required")
	}
	if in.Binding == "" {
		return nil, nil, fmt.Errorf("binding is required")
	}
	if in.Address == "" {
		return nil, nil, fmt.Errorf("address is required")
	}

	body := &rest_model.TerminatorCreate{
		Service: &in.ServiceID,
		Router:  &in.RouterID,
		Binding: &in.Binding,
		Address: &in.Address,
	}
	if in.Precedence != "" {
		prec := rest_model.TerminatorPrecedence(in.Precedence)
		body.Precedence = prec
	}
	if in.Cost > 0 {
		cost := rest_model.TerminatorCost(in.Cost)
		body.Cost = &cost
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtTerm.NewCreateTerminatorParams().WithContext(ctx).WithTerminator(body)
	resp, err := mgmt.Terminator.CreateTerminator(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create terminator: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type deleteTerminatorInput struct {
	ID string `json:"id" jsonschema:"required,terminator ID to delete"`
}

func (t *terminatorTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteTerminatorInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtTerm.NewDeleteTerminatorParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Terminator.DeleteTerminator(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete terminator %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
