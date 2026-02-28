package integration

import (
	"context"
	"testing"

	mgmtERP "github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	mgmtServicePolicy "github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
)

// Note: uniqueSuffix and ptr are defined in identities_test.go and main_test.go respectively.

func TestCreateServicePolicy(t *testing.T) {
	ctx := context.Background()
	policyName := "test-sp-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	// Use #all role selectors so no specific resource IDs are needed
	dialBind := rest_model.DialBindDial
	semantic := rest_model.SemanticAnyOf
	createParams := mgmtServicePolicy.NewCreateServicePolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServicePolicyCreate{
		Name:          &policyName,
		Type:          &dialBind,
		Semantic:      &semantic,
		ServiceRoles:  rest_model.Roles{"#all"},
		IdentityRoles: rest_model.Roles{"#all"},
	})
	resp, err := mgmt.ServicePolicy.CreateServicePolicy(createParams, nil)
	if err != nil {
		t.Fatalf("create service policy: %v", err)
	}
	policyID := resp.GetPayload().Data.ID
	defer func() {
		params := mgmtServicePolicy.NewDeleteServicePolicyParams().WithContext(ctx).WithID(policyID)
		if _, err := mgmt.ServicePolicy.DeleteServicePolicy(params, nil); err != nil {
			t.Logf("WARN: cleanup delete service policy %q: %v", policyID, err)
		}
	}()

	// Confirm it appears in the list
	listParams := mgmtServicePolicy.NewListServicePoliciesParams().WithContext(ctx)
	listParams.Filter = ptr(`name = "` + policyName + `"`)
	listResp, err := mgmt.ServicePolicy.ListServicePolicies(listParams, nil)
	if err != nil {
		t.Fatalf("list service policies: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected service policy %q in list, got 0 results", policyName)
	}
	got := listResp.GetPayload().Data[0]
	if *got.Name != policyName {
		t.Errorf("expected name %q, got %q", policyName, *got.Name)
	}
	if string(*got.Type) != string(rest_model.DialBindDial) {
		t.Errorf("expected type Dial, got %q", *got.Type)
	}
}

func TestGetServicePolicy(t *testing.T) {
	ctx := context.Background()
	policyName := "test-sp-get-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	dialBind := rest_model.DialBindDial
	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.ServicePolicy.CreateServicePolicy(
		mgmtServicePolicy.NewCreateServicePolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServicePolicyCreate{
			Name:          &policyName,
			Type:          &dialBind,
			Semantic:      &semantic,
			ServiceRoles:  rest_model.Roles{"#all"},
			IdentityRoles: rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create service policy: %v", err)
	}
	policyID := createResp.GetPayload().Data.ID
	defer func() {
		params := mgmtServicePolicy.NewDeleteServicePolicyParams().WithContext(ctx).WithID(policyID)
		if _, err := mgmt.ServicePolicy.DeleteServicePolicy(params, nil); err != nil {
			t.Logf("WARN: cleanup delete service policy %q: %v", policyID, err)
		}
	}()

	getResp, err := mgmt.ServicePolicy.DetailServicePolicy(
		mgmtServicePolicy.NewDetailServicePolicyParams().WithContext(ctx).WithID(policyID), nil)
	if err != nil {
		t.Fatalf("get service policy: %v", err)
	}
	got := getResp.GetPayload().Data
	if *got.ID != policyID {
		t.Errorf("expected id %q, got %q", policyID, *got.ID)
	}
	if *got.Name != policyName {
		t.Errorf("expected name %q, got %q", policyName, *got.Name)
	}
}

func TestUpdateServicePolicy(t *testing.T) {
	ctx := context.Background()
	policyName := "test-sp-update-" + uniqueSuffix()
	updatedName := policyName + "-updated"

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	dialBind := rest_model.DialBindDial
	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.ServicePolicy.CreateServicePolicy(
		mgmtServicePolicy.NewCreateServicePolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServicePolicyCreate{
			Name:          &policyName,
			Type:          &dialBind,
			Semantic:      &semantic,
			ServiceRoles:  rest_model.Roles{"#all"},
			IdentityRoles: rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create service policy: %v", err)
	}
	policyID := createResp.GetPayload().Data.ID
	defer func() {
		params := mgmtServicePolicy.NewDeleteServicePolicyParams().WithContext(ctx).WithID(policyID)
		if _, err := mgmt.ServicePolicy.DeleteServicePolicy(params, nil); err != nil {
			t.Logf("WARN: cleanup delete service policy %q: %v", policyID, err)
		}
	}()

	bindType := rest_model.DialBindBind
	updateParams := mgmtServicePolicy.NewUpdateServicePolicyParams().WithContext(ctx).WithID(policyID).WithPolicy(&rest_model.ServicePolicyUpdate{
		Name:          &updatedName,
		Type:          &bindType,
		Semantic:      &semantic,
		ServiceRoles:  rest_model.Roles{"#all"},
		IdentityRoles: rest_model.Roles{"#all"},
	})
	if _, err := mgmt.ServicePolicy.UpdateServicePolicy(updateParams, nil); err != nil {
		t.Fatalf("update service policy: %v", err)
	}

	getResp, err := mgmt.ServicePolicy.DetailServicePolicy(
		mgmtServicePolicy.NewDetailServicePolicyParams().WithContext(ctx).WithID(policyID), nil)
	if err != nil {
		t.Fatalf("get service policy after update: %v", err)
	}
	got := getResp.GetPayload().Data
	if *got.Name != updatedName {
		t.Errorf("expected updated name %q, got %q", updatedName, *got.Name)
	}
	if string(*got.Type) != string(rest_model.DialBindBind) {
		t.Errorf("expected type Bind, got %q", *got.Type)
	}
}

func TestDeleteServicePolicy(t *testing.T) {
	ctx := context.Background()
	policyName := "test-sp-delete-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	dialBind := rest_model.DialBindDial
	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.ServicePolicy.CreateServicePolicy(
		mgmtServicePolicy.NewCreateServicePolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServicePolicyCreate{
			Name:          &policyName,
			Type:          &dialBind,
			Semantic:      &semantic,
			ServiceRoles:  rest_model.Roles{"#all"},
			IdentityRoles: rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create service policy: %v", err)
	}
	policyID := createResp.GetPayload().Data.ID

	deleteParams := mgmtServicePolicy.NewDeleteServicePolicyParams().WithContext(ctx).WithID(policyID)
	if _, err := mgmt.ServicePolicy.DeleteServicePolicy(deleteParams, nil); err != nil {
		t.Fatalf("delete service policy: %v", err)
	}

	listParams := mgmtServicePolicy.NewListServicePoliciesParams().WithContext(ctx)
	listParams.Filter = ptr(`name = "` + policyName + `"`)
	listResp, err := mgmt.ServicePolicy.ListServicePolicies(listParams, nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected service policy to be deleted, but it still appears in list")
	}
}

func TestGetEdgeRouterPolicy(t *testing.T) {
	ctx := context.Background()
	policyName := "test-erp-get-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.EdgeRouterPolicy.CreateEdgeRouterPolicy(
		mgmtERP.NewCreateEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.EdgeRouterPolicyCreate{
			Name:            &policyName,
			Semantic:        &semantic,
			EdgeRouterRoles: rest_model.Roles{"#all"},
			IdentityRoles:   rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create edge router policy: %v", err)
	}
	erpID := createResp.GetPayload().Data.ID
	defer func() {
		params := mgmtERP.NewDeleteEdgeRouterPolicyParams().WithContext(ctx).WithID(erpID)
		if _, err := mgmt.EdgeRouterPolicy.DeleteEdgeRouterPolicy(params, nil); err != nil {
			t.Logf("WARN: cleanup delete edge router policy %q: %v", erpID, err)
		}
	}()

	getResp, err := mgmt.EdgeRouterPolicy.DetailEdgeRouterPolicy(
		mgmtERP.NewDetailEdgeRouterPolicyParams().WithContext(ctx).WithID(erpID), nil)
	if err != nil {
		t.Fatalf("get edge router policy: %v", err)
	}
	got := getResp.GetPayload().Data
	if *got.ID != erpID {
		t.Errorf("expected id %q, got %q", erpID, *got.ID)
	}
	if *got.Name != policyName {
		t.Errorf("expected name %q, got %q", policyName, *got.Name)
	}
}

func TestDeleteEdgeRouterPolicy(t *testing.T) {
	ctx := context.Background()
	policyName := "test-erp-delete-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	semantic := rest_model.SemanticAnyOf
	createResp, err := mgmt.EdgeRouterPolicy.CreateEdgeRouterPolicy(
		mgmtERP.NewCreateEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.EdgeRouterPolicyCreate{
			Name:            &policyName,
			Semantic:        &semantic,
			EdgeRouterRoles: rest_model.Roles{"#all"},
			IdentityRoles:   rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create edge router policy: %v", err)
	}
	erpID := createResp.GetPayload().Data.ID

	deleteParams := mgmtERP.NewDeleteEdgeRouterPolicyParams().WithContext(ctx).WithID(erpID)
	if _, err := mgmt.EdgeRouterPolicy.DeleteEdgeRouterPolicy(deleteParams, nil); err != nil {
		t.Fatalf("delete edge router policy: %v", err)
	}

	listParams := mgmtERP.NewListEdgeRouterPoliciesParams().WithContext(ctx)
	listParams.Filter = ptr(`name = "` + policyName + `"`)
	listResp, err := mgmt.EdgeRouterPolicy.ListEdgeRouterPolicies(listParams, nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected edge router policy to be deleted, but it still appears in list")
	}
}

func TestCreateEdgeRouterPolicy(t *testing.T) {
	ctx := context.Background()
	policyName := "test-erp-" + uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	// Use #all role selectors so no specific resource IDs are needed
	semantic := rest_model.SemanticAnyOf
	createParams := mgmtERP.NewCreateEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.EdgeRouterPolicyCreate{
		Name:            &policyName,
		Semantic:        &semantic,
		EdgeRouterRoles: rest_model.Roles{"#all"},
		IdentityRoles:   rest_model.Roles{"#all"},
	})
	resp, err := mgmt.EdgeRouterPolicy.CreateEdgeRouterPolicy(createParams, nil)
	if err != nil {
		t.Fatalf("create edge router policy: %v", err)
	}
	erpID := resp.GetPayload().Data.ID
	defer func() {
		params := mgmtERP.NewDeleteEdgeRouterPolicyParams().WithContext(ctx).WithID(erpID)
		if _, err := mgmt.EdgeRouterPolicy.DeleteEdgeRouterPolicy(params, nil); err != nil {
			t.Logf("WARN: cleanup delete edge router policy %q: %v", erpID, err)
		}
	}()

	// Confirm it appears in the list
	listParams := mgmtERP.NewListEdgeRouterPoliciesParams().WithContext(ctx)
	listParams.Filter = ptr(`name = "` + policyName + `"`)
	listResp, err := mgmt.EdgeRouterPolicy.ListEdgeRouterPolicies(listParams, nil)
	if err != nil {
		t.Fatalf("list edge router policies: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected edge router policy %q in list, got 0 results", policyName)
	}
	if *listResp.GetPayload().Data[0].Name != policyName {
		t.Errorf("expected name %q, got %q", policyName, *listResp.GetPayload().Data[0].Name)
	}
}
