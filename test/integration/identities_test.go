package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	mgmtIdentity "github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListIdentity(t *testing.T) {
	ctx := context.Background()
	name := "test-identity-list-" + uniqueSuffix()

	id := createIdentity(t, ctx, name)
	defer deleteIdentity(t, ctx, id)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtIdentity.NewListIdentitiesParams().WithContext(ctx)
	params.Filter = ptr(`name = "` + name + `"`)
	resp, err := mgmt.Identity.ListIdentities(params, nil)
	if err != nil {
		t.Fatalf("list identities: %v", err)
	}

	if len(resp.GetPayload().Data) == 0 {
		t.Fatalf("expected identity %q in list, got 0 results", name)
	}
	if *resp.GetPayload().Data[0].Name != name {
		t.Errorf("expected name %q, got %q", name, *resp.GetPayload().Data[0].Name)
	}
}

func TestGetIdentity(t *testing.T) {
	ctx := context.Background()
	name := "test-identity-get-" + uniqueSuffix()

	id := createIdentity(t, ctx, name)
	defer deleteIdentity(t, ctx, id)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtIdentity.NewDetailIdentityParams().WithContext(ctx).WithID(id)
	resp, err := mgmt.Identity.DetailIdentity(params, nil)
	if err != nil {
		t.Fatalf("get identity: %v", err)
	}

	got := resp.GetPayload().Data
	if *got.ID != id {
		t.Errorf("expected id %q, got %q", id, *got.ID)
	}
	if *got.Name != name {
		t.Errorf("expected name %q, got %q", name, *got.Name)
	}
}

func TestUpdateIdentity(t *testing.T) {
	ctx := context.Background()
	name := "test-identity-update-" + uniqueSuffix()
	updatedName := name + "-updated"

	id := createIdentity(t, ctx, name)
	defer deleteIdentity(t, ctx, id)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	idType := rest_model.IdentityTypeDevice
	updateParams := mgmtIdentity.NewUpdateIdentityParams().WithContext(ctx).WithID(id).WithIdentity(&rest_model.IdentityUpdate{
		Name:    &updatedName,
		Type:    &idType,
		IsAdmin: ptr(false),
	})
	_, err = mgmt.Identity.UpdateIdentity(updateParams, nil)
	if err != nil {
		t.Fatalf("update identity: %v", err)
	}

	getParams := mgmtIdentity.NewDetailIdentityParams().WithContext(ctx).WithID(id)
	resp, err := mgmt.Identity.DetailIdentity(getParams, nil)
	if err != nil {
		t.Fatalf("get identity after update: %v", err)
	}
	if *resp.GetPayload().Data.Name != updatedName {
		t.Errorf("expected updated name %q, got %q", updatedName, *resp.GetPayload().Data.Name)
	}
}

func TestDeleteIdentity(t *testing.T) {
	ctx := context.Background()
	name := "test-identity-delete-" + uniqueSuffix()

	id := createIdentity(t, ctx, name)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	deleteParams := mgmtIdentity.NewDeleteIdentityParams().WithContext(ctx).WithID(id)
	_, err = mgmt.Identity.DeleteIdentity(deleteParams, nil)
	if err != nil {
		t.Fatalf("delete identity: %v", err)
	}

	listParams := mgmtIdentity.NewListIdentitiesParams().WithContext(ctx)
	listParams.Filter = ptr(`name = "` + name + `"`)
	resp, err := mgmt.Identity.ListIdentities(listParams, nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(resp.GetPayload().Data) > 0 {
		t.Errorf("expected identity to be deleted, but it still appears in list")
	}
}

// createIdentity is a test helper that creates a Device identity and returns its ID.
func createIdentity(t *testing.T, ctx context.Context, name string) string {
	t.Helper()

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
	resp, err := mgmt.Identity.CreateIdentity(params, nil)
	if err != nil {
		t.Fatalf("create identity %q: %v", name, err)
	}
	return resp.GetPayload().Data.ID
}

// deleteIdentity is a test helper that deletes an identity by ID (best-effort cleanup).
func deleteIdentity(t *testing.T, ctx context.Context, id string) {
	t.Helper()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: could not get mgmt client for cleanup: %v", err)
		return
	}

	params := mgmtIdentity.NewDeleteIdentityParams().WithContext(ctx).WithID(id)
	if _, err := mgmt.Identity.DeleteIdentity(params, nil); err != nil {
		t.Logf("WARN: cleanup delete identity %q: %v", id, err)
	}
}

// uniqueSuffix returns a timestamp-based string for unique resource names.
func uniqueSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
