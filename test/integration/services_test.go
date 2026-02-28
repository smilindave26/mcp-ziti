package integration

import (
	"context"
	"testing"

	mgmtService "github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListService(t *testing.T) {
	ctx := context.Background()
	name := "test-service-list-" + uniqueSuffix()

	id := createService(t, ctx, name)
	defer deleteService(t, ctx, id)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtService.NewListServicesParams().WithContext(ctx)
	params.Filter = ptr(`name = "` + name + `"`)
	resp, err := mgmt.Service.ListServices(params, nil)
	if err != nil {
		t.Fatalf("list services: %v", err)
	}

	if len(resp.GetPayload().Data) == 0 {
		t.Fatalf("expected service %q in list, got 0 results", name)
	}
	if *resp.GetPayload().Data[0].Name != name {
		t.Errorf("expected name %q, got %q", name, *resp.GetPayload().Data[0].Name)
	}
}

func TestGetService(t *testing.T) {
	ctx := context.Background()
	name := "test-service-get-" + uniqueSuffix()

	id := createService(t, ctx, name)
	defer deleteService(t, ctx, id)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtService.NewDetailServiceParams().WithContext(ctx).WithID(id)
	resp, err := mgmt.Service.DetailService(params, nil)
	if err != nil {
		t.Fatalf("get service: %v", err)
	}

	got := resp.GetPayload().Data
	if *got.ID != id {
		t.Errorf("expected id %q, got %q", id, *got.ID)
	}
	if *got.Name != name {
		t.Errorf("expected name %q, got %q", name, *got.Name)
	}
}

func TestDeleteService(t *testing.T) {
	ctx := context.Background()
	name := "test-service-delete-" + uniqueSuffix()

	id := createService(t, ctx, name)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	deleteParams := mgmtService.NewDeleteServiceParams().WithContext(ctx).WithID(id)
	_, err = mgmt.Service.DeleteService(deleteParams, nil)
	if err != nil {
		t.Fatalf("delete service: %v", err)
	}

	listParams := mgmtService.NewListServicesParams().WithContext(ctx)
	listParams.Filter = ptr(`name = "` + name + `"`)
	resp, err := mgmt.Service.ListServices(listParams, nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(resp.GetPayload().Data) > 0 {
		t.Errorf("expected service to be deleted, but it still appears in list")
	}
}

func TestUpdateService(t *testing.T) {
	ctx := context.Background()
	name := "test-service-update-" + uniqueSuffix()
	updatedName := name + "-updated"

	id := createService(t, ctx, name)
	defer deleteService(t, ctx, id)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	updateParams := mgmtService.NewUpdateServiceParams().WithContext(ctx).WithID(id).WithService(&rest_model.ServiceUpdate{
		Name:               &updatedName,
		EncryptionRequired: false,
	})
	_, err = mgmt.Service.UpdateService(updateParams, nil)
	if err != nil {
		t.Fatalf("update service: %v", err)
	}

	getParams := mgmtService.NewDetailServiceParams().WithContext(ctx).WithID(id)
	resp, err := mgmt.Service.DetailService(getParams, nil)
	if err != nil {
		t.Fatalf("get service after update: %v", err)
	}
	if *resp.GetPayload().Data.Name != updatedName {
		t.Errorf("expected updated name %q, got %q", updatedName, *resp.GetPayload().Data.Name)
	}
}

// createService is a test helper that creates a service and returns its ID.
func createService(t *testing.T, ctx context.Context, name string) string {
	t.Helper()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtService.NewCreateServiceParams().WithContext(ctx).WithService(&rest_model.ServiceCreate{
		Name:               &name,
		EncryptionRequired: ptr(true),
	})
	resp, err := mgmt.Service.CreateService(params, nil)
	if err != nil {
		t.Fatalf("create service %q: %v", name, err)
	}
	return resp.GetPayload().Data.ID
}

// deleteService is a test helper that deletes a service by ID (best-effort cleanup).
func deleteService(t *testing.T, ctx context.Context, id string) {
	t.Helper()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: could not get mgmt client for cleanup: %v", err)
		return
	}

	params := mgmtService.NewDeleteServiceParams().WithContext(ctx).WithID(id)
	if _, err := mgmt.Service.DeleteService(params, nil); err != nil {
		t.Logf("WARN: cleanup delete service %q: %v", id, err)
	}
}
