package integration

import (
	"context"
	"testing"

	mgmtER "github.com/openziti/edge-api/rest_management_api_client/edge_router"
)

func TestListEdgeRouters_QuickstartHasAtLeastOne(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	limit := int64(100)
	offset := int64(0)
	params := mgmtER.NewListEdgeRoutersParams().WithContext(ctx)
	params.Limit = &limit
	params.Offset = &offset

	resp, err := mgmt.EdgeRouter.ListEdgeRouters(params, nil)
	if err != nil {
		t.Fatalf("list edge routers: %v", err)
	}

	if len(resp.GetPayload().Data) == 0 {
		t.Error("expected at least one edge router from quickstart, got 0")
	}
}

func TestGetEdgeRouter_ByID(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	// First list to get a valid ID
	limit := int64(1)
	offset := int64(0)
	listParams := mgmtER.NewListEdgeRoutersParams().WithContext(ctx)
	listParams.Limit = &limit
	listParams.Offset = &offset

	listResp, err := mgmt.EdgeRouter.ListEdgeRouters(listParams, nil)
	if err != nil {
		t.Fatalf("list edge routers: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Skip("no edge routers available to test get-by-id")
	}

	routerID := *listResp.GetPayload().Data[0].ID

	getParams := mgmtER.NewDetailEdgeRouterParams().WithContext(ctx).WithID(routerID)
	getResp, err := mgmt.EdgeRouter.DetailEdgeRouter(getParams, nil)
	if err != nil {
		t.Fatalf("get edge router %q: %v", routerID, err)
	}

	got := getResp.GetPayload().Data
	if *got.ID != routerID {
		t.Errorf("expected id %q, got %q", routerID, *got.ID)
	}
}

func TestGetEdgeRouter_NonExistent_ReturnsError(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	params := mgmtER.NewDetailEdgeRouterParams().WithContext(ctx).WithID("does-not-exist-00000000")
	_, err = mgmt.EdgeRouter.DetailEdgeRouter(params, nil)
	if err == nil {
		t.Error("expected error when fetching non-existent edge router")
	}
}
