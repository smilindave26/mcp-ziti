package integration

import (
	"context"
	"testing"

	mgmtEJS "github.com/openziti/edge-api/rest_management_api_client/external_jwt_signer"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListExternalJWTSigner(t *testing.T) {
	ctx := context.Background()
	name := "test-ejs-" + uniqueSuffix()
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.ExternalJWTSigner.CreateExternalJWTSigner(
		mgmtEJS.NewCreateExternalJWTSignerParams().WithContext(ctx).WithExternalJWTSigner(&rest_model.ExternalJWTSignerCreate{
			Name:     &name,
			Issuer:   ptr("https://issuer.example.com"),
			Audience: ptr("https://ctrl.example.com"),
			Enabled:  ptr(true),
			CertPem:  &certPEM,
		}), nil)
	if err != nil {
		t.Fatalf("create external JWT signer: %v", err)
	}
	ejsID := createResp.GetPayload().Data.ID
	defer deleteEJS(t, ctx, ejsID)

	listResp, err := mgmt.ExternalJWTSigner.ListExternalJWTSigners(
		mgmtEJS.NewListExternalJWTSignersParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list external JWT signers: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected signer %q in list, got 0 results", name)
	}
}

func TestGetExternalJWTSigner(t *testing.T) {
	ctx := context.Background()
	name := "test-ejs-get-" + uniqueSuffix()
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.ExternalJWTSigner.CreateExternalJWTSigner(
		mgmtEJS.NewCreateExternalJWTSignerParams().WithContext(ctx).WithExternalJWTSigner(&rest_model.ExternalJWTSignerCreate{
			Name:     &name,
			Issuer:   ptr("https://issuer.example.com"),
			Audience: ptr("https://ctrl.example.com"),
			Enabled:  ptr(true),
			CertPem:  &certPEM,
		}), nil)
	if err != nil {
		t.Fatalf("create external JWT signer: %v", err)
	}
	ejsID := createResp.GetPayload().Data.ID
	defer deleteEJS(t, ctx, ejsID)

	getResp, err := mgmt.ExternalJWTSigner.DetailExternalJWTSigner(
		mgmtEJS.NewDetailExternalJWTSignerParams().WithContext(ctx).WithID(ejsID), nil)
	if err != nil {
		t.Fatalf("get external JWT signer: %v", err)
	}
	if *getResp.GetPayload().Data.ID != ejsID {
		t.Errorf("expected id %q, got %q", ejsID, *getResp.GetPayload().Data.ID)
	}
}

func TestUpdateExternalJWTSigner(t *testing.T) {
	ctx := context.Background()
	name := "test-ejs-update-" + uniqueSuffix()
	updatedName := name + "-updated"
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.ExternalJWTSigner.CreateExternalJWTSigner(
		mgmtEJS.NewCreateExternalJWTSignerParams().WithContext(ctx).WithExternalJWTSigner(&rest_model.ExternalJWTSignerCreate{
			Name:     &name,
			Issuer:   ptr("https://issuer.example.com"),
			Audience: ptr("https://ctrl.example.com"),
			Enabled:  ptr(true),
			CertPem:  &certPEM,
		}), nil)
	if err != nil {
		t.Fatalf("create external JWT signer: %v", err)
	}
	ejsID := createResp.GetPayload().Data.ID
	defer deleteEJS(t, ctx, ejsID)

	_, err = mgmt.ExternalJWTSigner.UpdateExternalJWTSigner(
		mgmtEJS.NewUpdateExternalJWTSignerParams().WithContext(ctx).WithID(ejsID).WithExternalJWTSigner(&rest_model.ExternalJWTSignerUpdate{
			Name:     &updatedName,
			Issuer:   ptr("https://issuer.example.com"),
			Audience: ptr("https://ctrl.example.com"),
			Enabled:  ptr(false),
			CertPem:  &certPEM,
		}), nil)
	if err != nil {
		t.Fatalf("update external JWT signer: %v", err)
	}

	getResp, err := mgmt.ExternalJWTSigner.DetailExternalJWTSigner(
		mgmtEJS.NewDetailExternalJWTSignerParams().WithContext(ctx).WithID(ejsID), nil)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if *getResp.GetPayload().Data.Name != updatedName {
		t.Errorf("expected name %q, got %q", updatedName, *getResp.GetPayload().Data.Name)
	}
}

func TestDeleteExternalJWTSigner(t *testing.T) {
	ctx := context.Background()
	name := "test-ejs-delete-" + uniqueSuffix()
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.ExternalJWTSigner.CreateExternalJWTSigner(
		mgmtEJS.NewCreateExternalJWTSignerParams().WithContext(ctx).WithExternalJWTSigner(&rest_model.ExternalJWTSignerCreate{
			Name:     &name,
			Issuer:   ptr("https://issuer.example.com"),
			Audience: ptr("https://ctrl.example.com"),
			Enabled:  ptr(true),
			CertPem:  &certPEM,
		}), nil)
	if err != nil {
		t.Fatalf("create external JWT signer: %v", err)
	}
	ejsID := createResp.GetPayload().Data.ID

	if _, err := mgmt.ExternalJWTSigner.DeleteExternalJWTSigner(
		mgmtEJS.NewDeleteExternalJWTSignerParams().WithContext(ctx).WithID(ejsID), nil); err != nil {
		t.Fatalf("delete external JWT signer: %v", err)
	}

	listResp, err := mgmt.ExternalJWTSigner.ListExternalJWTSigners(
		mgmtEJS.NewListExternalJWTSignersParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected signer to be deleted, but it still appears in list")
	}
}

func deleteEJS(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: cleanup EJS %q: get client: %v", id, err)
		return
	}
	if _, err := mgmt.ExternalJWTSigner.DeleteExternalJWTSigner(
		mgmtEJS.NewDeleteExternalJWTSignerParams().WithContext(ctx).WithID(id), nil); err != nil {
		t.Logf("WARN: cleanup EJS %q: %v", id, err)
	}
}
