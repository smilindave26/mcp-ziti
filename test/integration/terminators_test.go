package integration

import (
	"context"
	"testing"

	mgmtTerm "github.com/openziti/edge-api/rest_management_api_client/terminator"
)

func TestListTerminators(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.Terminator.ListTerminators(
		mgmtTerm.NewListTerminatorsParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list terminators: %v", err)
	}
	// Quickstart may have no terminators; just verify the call succeeds
	_ = resp.GetPayload().Data
}
