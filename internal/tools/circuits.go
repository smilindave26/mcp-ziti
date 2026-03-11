package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	fabricCircuit "github.com/openziti/fabric/controller/rest_client/circuit"
	"github.com/openziti/fabric/controller/rest_model"
)

func registerCircuitTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &circuitTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-circuits",
		Description: "List active circuits in the Ziti fabric. Only available if the controller exposes the fabric API.",
		Annotations: readOnlyAnnotation,
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-circuit",
		Description: "Get a single circuit by ID from the Ziti fabric.",
		Annotations: readOnlyAnnotation,
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-circuit",
		Description: "Delete (tear down) a circuit by ID. Set immediate=true to skip graceful shutdown.",
		Annotations: destructiveAnnotation,
	}, t.delete)
}

type circuitTools struct{ zc *ziticlient.Client }

func (t *circuitTools) list(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	fabric, err := t.zc.Fabric()
	if err != nil {
		return nil, nil, err
	}

	resp, err := fabric.Circuit.ListCircuits(
		fabricCircuit.NewListCircuitsParams().WithContext(ctx))
	if err != nil {
		return nil, nil, fmt.Errorf("list circuits: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

func (t *circuitTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getInput) (*mcp.CallToolResult, any, error) {
	fabric, err := t.zc.Fabric()
	if err != nil {
		return nil, nil, err
	}

	resp, err := fabric.Circuit.DetailCircuit(
		fabricCircuit.NewDetailCircuitParams().WithContext(ctx).WithID(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("get circuit %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type deleteCircuitInput struct {
	ID        string `json:"id"                  jsonschema:"required,circuit ID to delete"`
	Immediate bool   `json:"immediate,omitempty" jsonschema:"skip graceful shutdown and tear down immediately"`
}

func (t *circuitTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteCircuitInput) (*mcp.CallToolResult, any, error) {
	fabric, err := t.zc.Fabric()
	if err != nil {
		return nil, nil, err
	}

	params := fabricCircuit.NewDeleteCircuitParams().WithContext(ctx).WithID(in.ID)
	if in.Immediate {
		params.WithOptions(&rest_model.CircuitDelete{Immediate: true})
	}

	_, err = fabric.Circuit.DeleteCircuit(params)
	if err != nil {
		return nil, nil, fmt.Errorf("delete circuit %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
