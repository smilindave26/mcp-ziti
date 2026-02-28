package integration

import (
	"context"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	mgmtEnroll "github.com/openziti/edge-api/rest_management_api_client/enrollment"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListEnrollment(t *testing.T) {
	ctx := context.Background()

	identityID := createIdentity(t, ctx, "test-enroll-identity-"+uniqueSuffix())
	defer deleteIdentity(t, ctx, identityID)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	method := "ott"
	expiresAt := strfmt.DateTime(time.Now().Add(24 * time.Hour))
	createResp, err := mgmt.Enrollment.CreateEnrollment(
		mgmtEnroll.NewCreateEnrollmentParams().WithContext(ctx).WithEnrollment(&rest_model.EnrollmentCreate{
			IdentityID: &identityID,
			Method:     &method,
			ExpiresAt:  &expiresAt,
		}), nil)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}
	enrollmentID := createResp.GetPayload().Data.ID
	defer func() {
		params := mgmtEnroll.NewDeleteEnrollmentParams().WithContext(ctx).WithID(enrollmentID)
		if _, err := mgmt.Enrollment.DeleteEnrollment(params, nil); err != nil {
			t.Logf("WARN: cleanup enrollment %q: %v", enrollmentID, err)
		}
	}()

	listResp, err := mgmt.Enrollment.ListEnrollments(
		mgmtEnroll.NewListEnrollmentsParams().WithContext(ctx).WithFilter(ptr(`identityId = "`+identityID+`"`)), nil)
	if err != nil {
		t.Fatalf("list enrollments: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Error("expected enrollment in list, got 0 results")
	}
}

func TestGetEnrollment(t *testing.T) {
	ctx := context.Background()

	identityID := createIdentity(t, ctx, "test-enroll-get-"+uniqueSuffix())
	defer deleteIdentity(t, ctx, identityID)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	method := "ott"
	expiresAt := strfmt.DateTime(time.Now().Add(24 * time.Hour))
	createResp, err := mgmt.Enrollment.CreateEnrollment(
		mgmtEnroll.NewCreateEnrollmentParams().WithContext(ctx).WithEnrollment(&rest_model.EnrollmentCreate{
			IdentityID: &identityID,
			Method:     &method,
			ExpiresAt:  &expiresAt,
		}), nil)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}
	enrollmentID := createResp.GetPayload().Data.ID
	defer func() {
		params := mgmtEnroll.NewDeleteEnrollmentParams().WithContext(ctx).WithID(enrollmentID)
		if _, err := mgmt.Enrollment.DeleteEnrollment(params, nil); err != nil {
			t.Logf("WARN: cleanup enrollment %q: %v", enrollmentID, err)
		}
	}()

	getResp, err := mgmt.Enrollment.DetailEnrollment(
		mgmtEnroll.NewDetailEnrollmentParams().WithContext(ctx).WithID(enrollmentID), nil)
	if err != nil {
		t.Fatalf("get enrollment: %v", err)
	}
	if *getResp.GetPayload().Data.ID != enrollmentID {
		t.Errorf("expected id %q, got %q", enrollmentID, *getResp.GetPayload().Data.ID)
	}
}

func TestDeleteEnrollment(t *testing.T) {
	ctx := context.Background()

	identityID := createIdentity(t, ctx, "test-enroll-delete-"+uniqueSuffix())
	defer deleteIdentity(t, ctx, identityID)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	method := "ott"
	expiresAt := strfmt.DateTime(time.Now().Add(24 * time.Hour))
	createResp, err := mgmt.Enrollment.CreateEnrollment(
		mgmtEnroll.NewCreateEnrollmentParams().WithContext(ctx).WithEnrollment(&rest_model.EnrollmentCreate{
			IdentityID: &identityID,
			Method:     &method,
			ExpiresAt:  &expiresAt,
		}), nil)
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}
	enrollmentID := createResp.GetPayload().Data.ID

	if _, err := mgmt.Enrollment.DeleteEnrollment(
		mgmtEnroll.NewDeleteEnrollmentParams().WithContext(ctx).WithID(enrollmentID), nil); err != nil {
		t.Fatalf("delete enrollment: %v", err)
	}

	listResp, err := mgmt.Enrollment.ListEnrollments(
		mgmtEnroll.NewListEnrollmentsParams().WithContext(ctx).WithFilter(ptr(`id = "`+enrollmentID+`"`)), nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected enrollment to be deleted, but it still appears in list")
	}
}
