package integration

import (
	"context"
	"testing"

	mgmtRouter "github.com/openziti/edge-api/rest_management_api_client/router"
)

func TestListRouters(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.Router.ListRouters(
		mgmtRouter.NewListRoutersParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list routers: %v", err)
	}
	if len(resp.GetPayload().Data) == 0 {
		t.Error("expected at least one router from quickstart, got 0")
	}
}

func TestGetRouter(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	listResp, err := mgmt.Router.ListRouters(
		mgmtRouter.NewListRoutersParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list routers: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Skip("no routers present")
	}

	id := *listResp.GetPayload().Data[0].ID
	getResp, err := mgmt.Router.DetailRouter(
		mgmtRouter.NewDetailRouterParams().WithContext(ctx).WithID(id), nil)
	if err != nil {
		t.Fatalf("get router %q: %v", id, err)
	}
	if *getResp.GetPayload().Data.ID != id {
		t.Errorf("expected id %q, got %q", id, *getResp.GetPayload().Data.ID)
	}
}
