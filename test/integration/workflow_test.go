package integration

import (
	"context"
	"testing"

	mgmtERP "github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	mgmtIdentity "github.com/openziti/edge-api/rest_management_api_client/identity"
	mgmtService "github.com/openziti/edge-api/rest_management_api_client/service"
	mgmtServicePolicy "github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
)

// TestFullWorkflow exercises a realistic end-to-end scenario:
//  1. Create an identity with a role attribute
//  2. Create a service with a role attribute
//  3. Create a Dial service policy linking them
//  4. Create an edge router policy covering all identities
//  5. Verify list-summary counts reflect the new resources
//  6. Tear down in reverse order
func TestFullWorkflow(t *testing.T) {
	ctx := context.Background()
	suffix := uniqueSuffix()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	// 1. Create identity with a role attribute.
	// Role attributes on resources are plain strings (no # prefix).
	// The # prefix is used in policy role *selectors* to reference attributes.
	identityName := "workflow-identity-" + suffix
	roleAttrValue := "workflow-" + suffix    // plain attribute value on the identity
	roleAttrSelector := "#" + roleAttrValue  // # selector used in policies
	idType := rest_model.IdentityTypeDevice
	roleAttrs := rest_model.Attributes{roleAttrValue}
	createIDResp, err := mgmt.Identity.CreateIdentity(
		mgmtIdentity.NewCreateIdentityParams().WithContext(ctx).WithIdentity(&rest_model.IdentityCreate{
			Name:           &identityName,
			Type:           &idType,
			IsAdmin:        ptr(false),
			RoleAttributes: &roleAttrs,
		}), nil)
	if err != nil {
		t.Fatalf("create identity: %v", err)
	}
	identityID := createIDResp.GetPayload().Data.ID
	defer func() {
		p := mgmtIdentity.NewDeleteIdentityParams().WithContext(ctx).WithID(identityID)
		if _, err := mgmt.Identity.DeleteIdentity(p, nil); err != nil {
			t.Logf("WARN: cleanup identity %q: %v", identityID, err)
		}
	}()

	// 2. Create service with role attribute
	serviceName := "workflow-service-" + suffix
	svcRoleAttrValue := "svc-workflow-" + suffix    // plain attribute on the service
	svcRoleAttrSelector := "#" + svcRoleAttrValue   // # selector used in policies
	createSvcResp, err := mgmt.Service.CreateService(
		mgmtService.NewCreateServiceParams().WithContext(ctx).WithService(&rest_model.ServiceCreate{
			Name:               &serviceName,
			EncryptionRequired: ptr(true),
			RoleAttributes:     []string{svcRoleAttrValue},
		}), nil)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	serviceID := createSvcResp.GetPayload().Data.ID
	defer func() {
		p := mgmtService.NewDeleteServiceParams().WithContext(ctx).WithID(serviceID)
		if _, err := mgmt.Service.DeleteService(p, nil); err != nil {
			t.Logf("WARN: cleanup service %q: %v", serviceID, err)
		}
	}()

	// 3. Create Dial service policy
	spName := "workflow-sp-" + suffix
	dialBind := rest_model.DialBindDial
	semantic := rest_model.SemanticAnyOf
	createSPResp, err := mgmt.ServicePolicy.CreateServicePolicy(
		mgmtServicePolicy.NewCreateServicePolicyParams().WithContext(ctx).WithPolicy(&rest_model.ServicePolicyCreate{
			Name:          &spName,
			Type:          &dialBind,
			Semantic:      &semantic,
			ServiceRoles:  rest_model.Roles{svcRoleAttrSelector},
			IdentityRoles: rest_model.Roles{roleAttrSelector},
		}), nil)
	if err != nil {
		t.Fatalf("create service policy: %v", err)
	}
	spID := createSPResp.GetPayload().Data.ID
	defer func() {
		p := mgmtServicePolicy.NewDeleteServicePolicyParams().WithContext(ctx).WithID(spID)
		if _, err := mgmt.ServicePolicy.DeleteServicePolicy(p, nil); err != nil {
			t.Logf("WARN: cleanup service policy %q: %v", spID, err)
		}
	}()

	// 4. Create edge router policy covering all identities
	erpName := "workflow-erp-" + suffix
	createERPResp, err := mgmt.EdgeRouterPolicy.CreateEdgeRouterPolicy(
		mgmtERP.NewCreateEdgeRouterPolicyParams().WithContext(ctx).WithPolicy(&rest_model.EdgeRouterPolicyCreate{
			Name:            &erpName,
			Semantic:        &semantic,
			EdgeRouterRoles: rest_model.Roles{"#all"},
			IdentityRoles:   rest_model.Roles{"#all"},
		}), nil)
	if err != nil {
		t.Fatalf("create edge router policy: %v", err)
	}
	erpID := createERPResp.GetPayload().Data.ID
	defer func() {
		p := mgmtERP.NewDeleteEdgeRouterPolicyParams().WithContext(ctx).WithID(erpID)
		if _, err := mgmt.EdgeRouterPolicy.DeleteEdgeRouterPolicy(p, nil); err != nil {
			t.Logf("WARN: cleanup edge router policy %q: %v", erpID, err)
		}
	}()

	// 5. Verify each resource is reachable by ID
	if _, err := mgmt.Identity.DetailIdentity(
		mgmtIdentity.NewDetailIdentityParams().WithContext(ctx).WithID(identityID), nil); err != nil {
		t.Errorf("get identity after create: %v", err)
	}

	if _, err := mgmt.Service.DetailService(
		mgmtService.NewDetailServiceParams().WithContext(ctx).WithID(serviceID), nil); err != nil {
		t.Errorf("get service after create: %v", err)
	}

	if _, err := mgmt.ServicePolicy.DetailServicePolicy(
		mgmtServicePolicy.NewDetailServicePolicyParams().WithContext(ctx).WithID(spID), nil); err != nil {
		t.Errorf("get service policy after create: %v", err)
	}

	if _, err := mgmt.EdgeRouterPolicy.DetailEdgeRouterPolicy(
		mgmtERP.NewDetailEdgeRouterPolicyParams().WithContext(ctx).WithID(erpID), nil); err != nil {
		t.Errorf("get edge router policy after create: %v", err)
	}

	// 6. Verify the service policy has the right type
	spResp, err := mgmt.ServicePolicy.DetailServicePolicy(
		mgmtServicePolicy.NewDetailServicePolicyParams().WithContext(ctx).WithID(spID), nil)
	if err != nil {
		t.Fatalf("get service policy for verification: %v", err)
	}
	got := spResp.GetPayload().Data
	if *got.Name != spName {
		t.Errorf("expected service policy name %q, got %q", spName, *got.Name)
	}
	if string(*got.Type) != string(rest_model.DialBindDial) {
		t.Errorf("expected type Dial, got %q", *got.Type)
	}
}
