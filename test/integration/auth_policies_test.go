package integration

import (
	"context"
	"testing"

	mgmtAP "github.com/openziti/edge-api/rest_management_api_client/auth_policy"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListAuthPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-ap-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.AuthPolicy.CreateAuthPolicy(
		mgmtAP.NewCreateAuthPolicyParams().WithContext(ctx).WithAuthPolicy(makeAuthPolicy(name)), nil)
	if err != nil {
		t.Fatalf("create auth policy: %v", err)
	}
	apID := createResp.GetPayload().Data.ID
	defer deleteAuthPolicy(t, ctx, apID)

	listResp, err := mgmt.AuthPolicy.ListAuthPolicies(
		mgmtAP.NewListAuthPoliciesParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list auth policies: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected auth policy %q in list, got 0 results", name)
	}
}

func TestGetAuthPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-ap-get-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.AuthPolicy.CreateAuthPolicy(
		mgmtAP.NewCreateAuthPolicyParams().WithContext(ctx).WithAuthPolicy(makeAuthPolicy(name)), nil)
	if err != nil {
		t.Fatalf("create auth policy: %v", err)
	}
	apID := createResp.GetPayload().Data.ID
	defer deleteAuthPolicy(t, ctx, apID)

	getResp, err := mgmt.AuthPolicy.DetailAuthPolicy(
		mgmtAP.NewDetailAuthPolicyParams().WithContext(ctx).WithID(apID), nil)
	if err != nil {
		t.Fatalf("get auth policy: %v", err)
	}
	if *getResp.GetPayload().Data.ID != apID {
		t.Errorf("expected id %q, got %q", apID, *getResp.GetPayload().Data.ID)
	}
}

func TestUpdateAuthPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-ap-update-" + uniqueSuffix()
	updatedName := name + "-updated"

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.AuthPolicy.CreateAuthPolicy(
		mgmtAP.NewCreateAuthPolicyParams().WithContext(ctx).WithAuthPolicy(makeAuthPolicy(name)), nil)
	if err != nil {
		t.Fatalf("create auth policy: %v", err)
	}
	apID := createResp.GetPayload().Data.ID
	defer deleteAuthPolicy(t, ctx, apID)

	_, err = mgmt.AuthPolicy.UpdateAuthPolicy(
		mgmtAP.NewUpdateAuthPolicyParams().WithContext(ctx).WithID(apID).WithAuthPolicy(
			&rest_model.AuthPolicyUpdate{AuthPolicyCreate: *makeAuthPolicy(updatedName)},
		), nil)
	if err != nil {
		t.Fatalf("update auth policy: %v", err)
	}

	getResp, err := mgmt.AuthPolicy.DetailAuthPolicy(
		mgmtAP.NewDetailAuthPolicyParams().WithContext(ctx).WithID(apID), nil)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if *getResp.GetPayload().Data.Name != updatedName {
		t.Errorf("expected name %q, got %q", updatedName, *getResp.GetPayload().Data.Name)
	}
}

func TestDeleteAuthPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-ap-delete-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.AuthPolicy.CreateAuthPolicy(
		mgmtAP.NewCreateAuthPolicyParams().WithContext(ctx).WithAuthPolicy(makeAuthPolicy(name)), nil)
	if err != nil {
		t.Fatalf("create auth policy: %v", err)
	}
	apID := createResp.GetPayload().Data.ID

	if _, err := mgmt.AuthPolicy.DeleteAuthPolicy(
		mgmtAP.NewDeleteAuthPolicyParams().WithContext(ctx).WithID(apID), nil); err != nil {
		t.Fatalf("delete auth policy: %v", err)
	}

	listResp, err := mgmt.AuthPolicy.ListAuthPolicies(
		mgmtAP.NewListAuthPoliciesParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected auth policy to be deleted, but it still appears in list")
	}
}

// makeAuthPolicy builds a minimal AuthPolicyCreate allowing no auth methods.
// Tests that need a valid but harmless policy use this.
func makeAuthPolicy(name string) *rest_model.AuthPolicyCreate {
	return &rest_model.AuthPolicyCreate{
		Name: &name,
		Primary: &rest_model.AuthPolicyPrimary{
			Cert: &rest_model.AuthPolicyPrimaryCert{
				Allowed:           ptr(false),
				AllowExpiredCerts: ptr(false),
			},
			ExtJWT: &rest_model.AuthPolicyPrimaryExtJWT{
				Allowed:        ptr(false),
				AllowedSigners: []string{},
			},
			Updb: &rest_model.AuthPolicyPrimaryUpdb{
				Allowed:                ptr(false),
				LockoutDurationMinutes: ptr(int64(0)),
				MaxAttempts:            ptr(int64(0)),
				MinPasswordLength:      ptr(int64(5)),
				RequireMixedCase:       ptr(false),
				RequireNumberChar:      ptr(false),
				RequireSpecialChar:     ptr(false),
			},
		},
		Secondary: &rest_model.AuthPolicySecondary{
			RequireTotp: ptr(false),
		},
	}
}

func deleteAuthPolicy(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: cleanup auth policy %q: get client: %v", id, err)
		return
	}
	if _, err := mgmt.AuthPolicy.DeleteAuthPolicy(
		mgmtAP.NewDeleteAuthPolicyParams().WithContext(ctx).WithID(id), nil); err != nil {
		t.Logf("WARN: cleanup auth policy %q: %v", id, err)
	}
}
