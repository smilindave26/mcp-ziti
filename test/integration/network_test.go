package integration

import (
	"context"
	"testing"

	mgmtER "github.com/openziti/edge-api/rest_management_api_client/edge_router"
	mgmtIdentity "github.com/openziti/edge-api/rest_management_api_client/identity"
	mgmtInfo "github.com/openziti/edge-api/rest_management_api_client/informational"
	mgmtService "github.com/openziti/edge-api/rest_management_api_client/service"
)

func TestGetControllerVersion(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.Informational.ListVersion(mgmtInfo.NewListVersionParams().WithContext(ctx))
	if err != nil {
		t.Fatalf("get controller version: %v", err)
	}

	data := resp.GetPayload().Data
	if data == nil {
		t.Fatal("expected non-nil version data")
	}
	if data.Version == "" {
		t.Error("expected non-empty version string")
	}
}

func TestListSummary_IdentityCountNonNegative(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	limit := int64(1)
	offset := int64(0)
	params := mgmtIdentity.NewListIdentitiesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	resp, err := mgmt.Identity.ListIdentities(params, nil)
	if err != nil {
		t.Fatalf("list identities for count: %v", err)
	}
	if resp.GetPayload().Meta.Pagination == nil {
		t.Error("expected pagination metadata in identity list response")
	} else if *resp.GetPayload().Meta.Pagination.TotalCount < 0 {
		t.Error("identity total count should not be negative")
	}
}

func TestListSummary_ServiceCountNonNegative(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	limit := int64(1)
	offset := int64(0)
	params := mgmtService.NewListServicesParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	resp, err := mgmt.Service.ListServices(params, nil)
	if err != nil {
		t.Fatalf("list services for count: %v", err)
	}
	if resp.GetPayload().Meta.Pagination == nil {
		t.Error("expected pagination metadata in service list response")
	} else if *resp.GetPayload().Meta.Pagination.TotalCount < 0 {
		t.Error("service total count should not be negative")
	}
}

func TestListSummary_QuickstartHasEdgeRouter(t *testing.T) {
	ctx := context.Background()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	limit := int64(1)
	offset := int64(0)
	params := mgmtER.NewListEdgeRoutersParams().WithContext(ctx).WithLimit(&limit).WithOffset(&offset)
	resp, err := mgmt.EdgeRouter.ListEdgeRouters(params, nil)
	if err != nil {
		t.Fatalf("list edge routers for count: %v", err)
	}
	if resp.GetPayload().Meta.Pagination == nil {
		t.Fatal("expected pagination metadata in edge router response")
	}
	if *resp.GetPayload().Meta.Pagination.TotalCount < 1 {
		t.Errorf("expected at least 1 edge router from quickstart, got %d",
			*resp.GetPayload().Meta.Pagination.TotalCount)
	}
}
