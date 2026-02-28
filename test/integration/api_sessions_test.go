package integration

import (
	"context"
	"testing"

	mgmtAPISession "github.com/openziti/edge-api/rest_management_api_client/api_session"
	mgmtSession "github.com/openziti/edge-api/rest_management_api_client/session"
)

func TestListAPISessions(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.APISession.ListAPISessions(
		mgmtAPISession.NewListAPISessionsParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list api sessions: %v", err)
	}
	// We must have at least our own session
	if len(resp.GetPayload().Data) == 0 {
		t.Error("expected at least one API session (our own), got 0")
	}
}

func TestGetAPISession(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	listResp, err := mgmt.APISession.ListAPISessions(
		mgmtAPISession.NewListAPISessionsParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list api sessions: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Skip("no API sessions present")
	}

	id := *listResp.GetPayload().Data[0].ID
	getResp, err := mgmt.APISession.DetailAPISessions(
		mgmtAPISession.NewDetailAPISessionsParams().WithContext(ctx).WithID(id), nil)
	if err != nil {
		t.Fatalf("get api session %q: %v", id, err)
	}
	if *getResp.GetPayload().Data.ID != id {
		t.Errorf("expected id %q, got %q", id, *getResp.GetPayload().Data.ID)
	}
}

func TestListSessions(t *testing.T) {
	ctx := context.Background()

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := mgmt.Session.ListSessions(
		mgmtSession.NewListSessionsParams().WithContext(ctx), nil)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	// May be empty in quickstart; just verify no error
	_ = resp.GetPayload().Data
}
