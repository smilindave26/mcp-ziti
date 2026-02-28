package integration

import (
	"context"
	"testing"

	mgmtCtrl "github.com/openziti/edge-api/rest_management_api_client/controllers"
	mgmtDB "github.com/openziti/edge-api/rest_management_api_client/database"
)

func TestCreateDatabaseSnapshot(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgmt.Database.CreateDatabaseSnapshot(
		mgmtDB.NewCreateDatabaseSnapshotParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("create database snapshot: %v", err)
	}
}

func TestCheckDataIntegrity(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.Database.CheckDataIntegrity(
		mgmtDB.NewCheckDataIntegrityParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("check data integrity: %v", err)
	}
	// A healthy quickstart controller should report no errors
	payload := resp.GetPayload()
	if payload == nil {
		t.Fatal("expected non-nil payload from check data integrity")
	}
}

func TestListControllers(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.Controllers.ListControllers(
		mgmtCtrl.NewListControllersParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list controllers: %v", err)
	}
	// Non-HA quickstart may return an empty list; just verify the call succeeds
	t.Logf("list controllers returned %d results", len(resp.GetPayload().Data))
}
