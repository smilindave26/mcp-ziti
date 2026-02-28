package integration

import (
	"context"
	"testing"

	mgmtPC "github.com/openziti/edge-api/rest_management_api_client/posture_checks"
)

func TestListPostureChecks(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.PostureChecks.ListPostureChecks(
		mgmtPC.NewListPostureChecksParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list posture checks: %v", err)
	}
	// Quickstart may have no posture checks; just verify the call succeeds
	_ = resp.GetPayload().Data
}

func TestListPostureCheckTypes(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.PostureChecks.ListPostureCheckTypes(
		mgmtPC.NewListPostureCheckTypesParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list posture check types: %v", err)
	}
	// Verify the call succeeds; log count for diagnostics
	t.Logf("list posture check types returned %d results", len(resp.GetPayload().Data))
}
