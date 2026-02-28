package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/netfoundry/mcp-ziti-golang/internal/ziticlient"
	mgmtEnroll "github.com/openziti/edge-api/rest_management_api_client/enrollment"
	"github.com/openziti/edge-api/rest_model"
)

func registerEnrollmentTools(s *mcp.Server, zc *ziticlient.Client) {
	t := &enrollmentTools{zc: zc}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-enrollments",
		Description: "List pending enrollments. Returns up to `limit` results (default 100, max 500). Use `offset` to paginate.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.list)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-enrollment",
		Description: "Get a single enrollment by ID.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, t.get)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-enrollment",
		Description: "Create a new enrollment for an identity. method must be one of: ott, ottca, updb. expiresAt is an RFC3339 timestamp (e.g. 2026-01-01T00:00:00Z).",
	}, t.create)

	destructive := true
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-enrollment",
		Description: "Delete a pending enrollment by ID.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, t.delete)
}

type enrollmentTools struct{ zc *ziticlient.Client }

type listEnrollmentsInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional filter expression"`
	Limit  int64  `json:"limit,omitempty"  jsonschema:"max results to return (default 100, max 500)"`
	Offset int64  `json:"offset,omitempty" jsonschema:"number of results to skip for pagination"`
}

func (t *enrollmentTools) list(ctx context.Context, _ *mcp.CallToolRequest, in listEnrollmentsInput) (*mcp.CallToolResult, any, error) {
	limit, offset := clampLimit(in.Limit), in.Offset

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEnroll.NewListEnrollmentsParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	if in.Filter != "" {
		params.Filter = &in.Filter
	}

	resp, err := mgmt.Enrollment.ListEnrollments(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("list enrollments: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type getEnrollmentInput struct {
	ID string `json:"id" jsonschema:"required,enrollment ID"`
}

func (t *enrollmentTools) get(ctx context.Context, _ *mcp.CallToolRequest, in getEnrollmentInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEnroll.NewDetailEnrollmentParams().WithContext(ctx).WithID(in.ID)
	resp, err := mgmt.Enrollment.DetailEnrollment(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("get enrollment %q: %w", in.ID, err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type createEnrollmentInput struct {
	IdentityID string `json:"identityId" jsonschema:"required,identity ID to enroll"`
	Method     string `json:"method"     jsonschema:"required,enrollment method: ott, ottca, or updb"`
	ExpiresAt  string `json:"expiresAt"  jsonschema:"required,expiry timestamp in RFC3339 format e.g. 2026-01-01T00:00:00Z"`
	CaID       string `json:"caId,omitempty" jsonschema:"CA ID for ottca method"`
	Username   string `json:"username,omitempty" jsonschema:"username for updb method"`
}

func (t *enrollmentTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createEnrollmentInput) (*mcp.CallToolResult, any, error) {
	if in.IdentityID == "" {
		return nil, nil, fmt.Errorf("identityId is required")
	}
	if in.Method == "" {
		return nil, nil, fmt.Errorf("method is required (ott, ottca, or updb)")
	}
	if in.ExpiresAt == "" {
		return nil, nil, fmt.Errorf("expiresAt is required (RFC3339 timestamp)")
	}

	ts, err := time.Parse(time.RFC3339, in.ExpiresAt)
	if err != nil {
		return nil, nil, fmt.Errorf("expiresAt must be RFC3339 format: %w", err)
	}
	expiresAt := strfmt.DateTime(ts)
	method := in.Method

	body := &rest_model.EnrollmentCreate{
		IdentityID: &in.IdentityID,
		Method:     &method,
		ExpiresAt:  &expiresAt,
	}
	if in.CaID != "" {
		body.CaID = &in.CaID
	}
	if in.Username != "" {
		body.Username = &in.Username
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEnroll.NewCreateEnrollmentParams().WithContext(ctx).WithEnrollment(body)
	resp, err := mgmt.Enrollment.CreateEnrollment(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create enrollment: %w", err)
	}
	return jsonResult(resp.GetPayload().Data)
}

type deleteEnrollmentInput struct {
	ID string `json:"id" jsonschema:"required,enrollment ID to delete"`
}

func (t *enrollmentTools) delete(ctx context.Context, _ *mcp.CallToolRequest, in deleteEnrollmentInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	mgmt, err := t.zc.Mgmt()
	if err != nil {
		return nil, nil, err
	}

	params := mgmtEnroll.NewDeleteEnrollmentParams().WithContext(ctx).WithID(in.ID)
	_, err = mgmt.Enrollment.DeleteEnrollment(params, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("delete enrollment %q: %w", in.ID, err)
	}
	return jsonResult(map[string]string{"status": "deleted", "id": in.ID})
}
