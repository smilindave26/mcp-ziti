package integration

import (
	"context"
	"testing"

	mgmtRA "github.com/openziti/edge-api/rest_management_api_client/role_attributes"
)

func TestListIdentityRoleAttributes(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgmt.RoleAttributes.ListIdentityRoleAttributes(
		mgmtRA.NewListIdentityRoleAttributesParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list identity role attributes: %v", err)
	}
}

func TestListEdgeRouterRoleAttributes(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgmt.RoleAttributes.ListEdgeRouterRoleAttributes(
		mgmtRA.NewListEdgeRouterRoleAttributesParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list edge router role attributes: %v", err)
	}
}

func TestListServiceRoleAttributes(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgmt.RoleAttributes.ListServiceRoleAttributes(
		mgmtRA.NewListServiceRoleAttributesParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list service role attributes: %v", err)
	}
}

func TestListPostureCheckRoleAttributes(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	_, err = mgmt.RoleAttributes.ListPostureCheckRoleAttributes(
		mgmtRA.NewListPostureCheckRoleAttributesParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list posture check role attributes: %v", err)
	}
}
