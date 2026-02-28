package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtDB "github.com/openziti/edge-api/rest_management_api_client/database"
)

func registerDatabaseTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &databaseTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-database-snapshot",
		Description: "Trigger an immediate backup snapshot of the controller's database.",
	}, t.snapshot)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "check-data-integrity",
		Description: "Run an integrity check on the controller database and return any issues found.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, t.checkIntegrity)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "fix-data-integrity",
		Description: "Attempt to automatically fix data integrity issues found in the controller database.",
	}, t.fixIntegrity)
}

type databaseTools struct{ zc *ziticlient.Client }

type emptyInput struct{}

func (t *databaseTools) snapshot(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtDB.NewCreateDatabaseSnapshotParams().WithContext(ctx)
	_, err = mgmt.Database.CreateDatabaseSnapshot(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create database snapshot: %w", err)
	}
	return jsonResult(map[string]string{"status": "snapshot created"})
}

func (t *databaseTools) checkIntegrity(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtDB.NewCheckDataIntegrityParams().WithContext(ctx)
	resp, err := mgmt.Database.CheckDataIntegrity(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("check data integrity: %w", err)
	}
	return jsonResult(resp.GetPayload())
}

func (t *databaseTools) fixIntegrity(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtDB.NewFixDataIntegrityParams().WithContext(ctx)
	resp, err := mgmt.Database.FixDataIntegrity(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("fix data integrity: %w", err)
	}
	return jsonResult(resp.GetPayload())
}
