package integration

import (
	"context"
	"testing"

	mgmtIdentity "github.com/openziti/edge-api/rest_management_api_client/identity"
	mgmtService "github.com/openziti/edge-api/rest_management_api_client/service"
	mgmtServicePolicy "github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
)

func TestGetNonExistentIdentity_ReturnsError(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtIdentity.NewDetailIdentityParams().WithContext(ctx).WithID("does-not-exist-00000000")
	_, err = mgmt.Identity.DetailIdentity(params, nil)
	if err == nil {
		t.Error("expected error when fetching non-existent identity")
	}
}

func TestGetNonExistentService_ReturnsError(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtService.NewDetailServiceParams().WithContext(ctx).WithID("does-not-exist-00000000")
	_, err = mgmt.Service.DetailService(params, nil)
	if err == nil {
		t.Error("expected error when fetching non-existent service")
	}
}

func TestDeleteNonExistentIdentity_ReturnsError(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtIdentity.NewDeleteIdentityParams().WithContext(ctx).WithID("does-not-exist-00000000")
	_, err = mgmt.Identity.DeleteIdentity(params, nil)
	if err == nil {
		t.Error("expected error when deleting non-existent identity")
	}
}

func TestDeleteNonExistentService_ReturnsError(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtService.NewDeleteServiceParams().WithContext(ctx).WithID("does-not-exist-00000000")
	_, err = mgmt.Service.DeleteService(params, nil)
	if err == nil {
		t.Error("expected error when deleting non-existent service")
	}
}

func TestCreateIdentity_DuplicateName_ReturnsError(t *testing.T) {
	ctx := context.Background()
	name := "test-dup-" + uniqueSuffix()

	id := createIdentity(t, ctx, name)
	defer deleteIdentity(t, ctx, id)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	idType := rest_model.IdentityTypeDevice
	params := mgmtIdentity.NewCreateIdentityParams().WithContext(ctx).WithIdentity(&rest_model.IdentityCreate{
		Name:    &name,
		Type:    &idType,
		IsAdmin: ptr(false),
	})
	_, err = mgmt.Identity.CreateIdentity(params, nil)
	if err == nil {
		t.Error("expected error when creating identity with duplicate name")
	}
}

func TestCreateServicePolicy_InvalidType_ReturnsError(t *testing.T) {
	ctx := context.Background()
	name := "test-bad-type-" + uniqueSuffix()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	badType := rest_model.DialBind("InvalidType")
	semantic := rest_model.SemanticAnyOf
	params := mgmtServicePolicy.NewCreateServicePolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServicePolicyCreate{
		Name:          &name,
		Type:          &badType,
		Semantic:      &semantic,
		ServiceRoles:  rest_model.Roles{"#all"},
		IdentityRoles: rest_model.Roles{"#all"},
	})
	_, err = mgmt.ServicePolicy.CreateServicePolicy(params, nil)
	if err == nil {
		t.Error("expected error when creating service policy with invalid type")
	}
}

func TestDeleteIdentity_DoubleDeletion_ReturnsError(t *testing.T) {
	ctx := context.Background()
	name := "test-double-del-" + uniqueSuffix()

	id := createIdentity(t, ctx, name)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	// First delete — should succeed
	params := mgmtIdentity.NewDeleteIdentityParams().WithContext(ctx).WithID(id)
	if _, err := mgmt.Identity.DeleteIdentity(params, nil); err != nil {
		t.Fatalf("first delete failed: %v", err)
	}

	// Second delete — should fail
	if _, err := mgmt.Identity.DeleteIdentity(params, nil); err == nil {
		t.Error("expected error on second delete of same identity")
	}
}
