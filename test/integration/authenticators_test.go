package integration

import (
	"context"
	"testing"

	mgmtAuth "github.com/openziti/edge-api/rest_management_api_client/authenticator"
)

func TestListAuthenticators(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	// Quickstart creates an admin identity with an updb authenticator
	resp, err := mgmt.Authenticator.ListAuthenticators(
		mgmtAuth.NewListAuthenticatorsParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list authenticators: %v", err)
	}
	// Quickstart should have at least the admin updb authenticator; log for diagnostics
	t.Logf("list authenticators returned %d results", len(resp.GetPayload().Data))
}

func TestGetAuthenticator(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	listResp, err := mgmt.Authenticator.ListAuthenticators(
		mgmtAuth.NewListAuthenticatorsParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list authenticators: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Skip("no authenticators present")
	}

	id := *listResp.GetPayload().Data[0].ID
	getResp, err := mgmt.Authenticator.DetailAuthenticator(
		mgmtAuth.NewDetailAuthenticatorParams().WithContext(ctx).WithID(id), nil)
	if err != nil {
		t.Fatalf("get authenticator %q: %v", id, err)
	}
	if *getResp.GetPayload().Data.ID != id {
		t.Errorf("expected id %q, got %q", id, *getResp.GetPayload().Data.ID)
	}
}
