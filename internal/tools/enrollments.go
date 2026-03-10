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
		Annotations: readOnlyAnnotation,
	}, makeListHandler(zc, "enrollments", func(ctx context.Context, mgmt *mgmtAPI, filter *string, limit, offset *int64) (any, error) {
		params := mgmtEnroll.NewListEnrollmentsParams().WithContext(ctx).WithLimit(limit).WithOffset(offset)
		params.Filter = filter
		resp, err := mgmt.Enrollment.ListEnrollments(params, nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-enrollment",
		Description: "Get a single enrollment by ID.",
		Annotations: readOnlyAnnotation,
	}, makeGetHandler(zc, "enrollment", func(ctx context.Context, mgmt *mgmtAPI, id string) (any, error) {
		resp, err := mgmt.Enrollment.DetailEnrollment(
			mgmtEnroll.NewDetailEnrollmentParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}
		return resp.GetPayload().Data, nil
	}))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create-enrollment",
		Description: "Create a new enrollment for an identity. method must be one of: ott, ottca, updb. expiresAt is an RFC3339 timestamp (e.g. 2026-01-01T00:00:00Z).",
	}, t.create)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-enrollment",
		Description: "Delete a pending enrollment by ID.",
		Annotations: destructiveAnnotation,
	}, makeDeleteHandler(zc, "enrollment", func(ctx context.Context, mgmt *mgmtAPI, id string) error {
		_, err := mgmt.Enrollment.DeleteEnrollment(
			mgmtEnroll.NewDeleteEnrollmentParams().WithContext(ctx).WithID(id), nil)
		return err
	}))
}

type enrollmentTools struct{ zc *ziticlient.Client }

type createEnrollmentInput struct {
	IdentityID string `json:"identityId" jsonschema:"required,identity ID to enroll"`
	Method     string `json:"method"     jsonschema:"required,enrollment method: ott, ottca, or updb"`
	ExpiresAt  string `json:"expiresAt"  jsonschema:"required,expiry timestamp in RFC3339 format e.g. 2026-01-01T00:00:00Z"`
	CaID       string `json:"caId,omitempty" jsonschema:"CA ID for ottca method"`
	Username   string `json:"username,omitempty" jsonschema:"username for updb method"`
}

func (t *enrollmentTools) create(ctx context.Context, _ *mcp.CallToolRequest, in createEnrollmentInput) (*mcp.CallToolResult, any, error) {
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
