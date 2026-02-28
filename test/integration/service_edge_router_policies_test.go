package integration

import (
	"context"
	"testing"

	mgmtSERP "github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListServiceEdgeRouterPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-serp-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(
		mgmtSERP.NewCreateServiceEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServiceEdgeRouterPolicyCreate{
			Name:            &name,
			Semantic:        &semantic,
			ServiceRoles:    rest_model.Roles{"#all"},
			EdgeRouterRoles: rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create service edge router policy: %v", err)
	}
	serpID := createResp.GetPayload().Data.ID
	defer deleteSERP(t, ctx, serpID)

	listResp, err := mgmt.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(
		mgmtSERP.NewListServiceEdgeRouterPoliciesParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list service edge router policies: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected policy %q in list, got 0 results", name)
	}
}

func TestGetServiceEdgeRouterPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-serp-get-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(
		mgmtSERP.NewCreateServiceEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServiceEdgeRouterPolicyCreate{
			Name:            &name,
			Semantic:        &semantic,
			ServiceRoles:    rest_model.Roles{"#all"},
			EdgeRouterRoles: rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create service edge router policy: %v", err)
	}
	serpID := createResp.GetPayload().Data.ID
	defer deleteSERP(t, ctx, serpID)

	getResp, err := mgmt.ServiceEdgeRouterPolicy.DetailServiceEdgeRouterPolicy(
		mgmtSERP.NewDetailServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(serpID), nil)
	if err != nil {
		t.Fatalf("get service edge router policy: %v", err)
	}
	if *getResp.GetPayload().Data.ID != serpID {
		t.Errorf("expected id %q, got %q", serpID, *getResp.GetPayload().Data.ID)
	}
}

func TestUpdateServiceEdgeRouterPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-serp-update-" + uniqueSuffix()
	updatedName := name + "-updated"

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(
		mgmtSERP.NewCreateServiceEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServiceEdgeRouterPolicyCreate{
			Name:            &name,
			Semantic:        &semantic,
			ServiceRoles:    rest_model.Roles{"#all"},
			EdgeRouterRoles: rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create service edge router policy: %v", err)
	}
	serpID := createResp.GetPayload().Data.ID
	defer deleteSERP(t, ctx, serpID)

	allOf := rest_model.SemanticAllOf
	_, err = mgmt.ServiceEdgeRouterPolicy.UpdateServiceEdgeRouterPolicy(
		mgmtSERP.NewUpdateServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(serpID).WithPolicy(&rest_model.ServiceEdgeRouterPolicyUpdate{
			Name:            &updatedName,
			Semantic:        &allOf,
			ServiceRoles:    rest_model.Roles{"#all"},
			EdgeRouterRoles: rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("update service edge router policy: %v", err)
	}

	getResp, err := mgmt.ServiceEdgeRouterPolicy.DetailServiceEdgeRouterPolicy(
		mgmtSERP.NewDetailServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(serpID), nil)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if *getResp.GetPayload().Data.Name != updatedName {
		t.Errorf("expected name %q, got %q", updatedName, *getResp.GetPayload().Data.Name)
	}
}

func TestDeleteServiceEdgeRouterPolicy(t *testing.T) {
	ctx := context.Background()
	name := "test-serp-delete-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(
		mgmtSERP.NewCreateServiceEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServiceEdgeRouterPolicyCreate{
			Name:            &name,
			Semantic:        &semantic,
			ServiceRoles:    rest_model.Roles{},
			EdgeRouterRoles: rest_model.Roles{},
		}), nil)
	if err != nil {
		t.Fatalf("create service edge router policy: %v", err)
	}
	serpID := createResp.GetPayload().Data.ID

	if _, err := mgmt.ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(
		mgmtSERP.NewDeleteServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(serpID), nil); err != nil {
		t.Fatalf("delete service edge router policy: %v", err)
	}

	listResp, err := mgmt.ServiceEdgeRouterPolicy.ListServiceEdgeRouterPolicies(
		mgmtSERP.NewListServiceEdgeRouterPoliciesParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected policy to be deleted, but it still appears in list")
	}
}

func deleteSERP(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: cleanup SERP %q: get client: %v", id, err)
		return
	}
	if _, err := mgmt.ServiceEdgeRouterPolicy.DeleteServiceEdgeRouterPolicy(
		mgmtSERP.NewDeleteServiceEdgeRouterPolicyParams().WithContext(ctx).WithID(id), nil); err != nil {
		t.Logf("WARN: cleanup SERP %q: %v", id, err)
	}
}
