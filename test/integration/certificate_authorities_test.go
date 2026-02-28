package integration

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	mgmtCA "github.com/openziti/edge-api/rest_management_api_client/certificate_authority"
	"github.com/openziti/edge-api/rest_model"
)

func TestCreateAndListCA(t *testing.T) {
	ctx := context.Background()
	name := "test-ca-" + uniqueSuffix()
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.CertificateAuthority.CreateCa(
		mgmtCA.NewCreateCaParams().WithContext(ctx).WithCa(&rest_model.CaCreate{
			Name:                      &name,
			CertPem:                   &certPEM,
			IsAuthEnabled:             ptr(false),
			IsAutoCaEnrollmentEnabled: ptr(false),
			IsOttCaEnrollmentEnabled:  ptr(false),
			IdentityRoles:             rest_model.Roles{},
		}), nil)
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	caID := createResp.GetPayload().Data.ID
	defer deleteCA(t, ctx, caID)

	listResp, err := mgmt.CertificateAuthority.ListCas(
		mgmtCA.NewListCasParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list CAs: %v", err)
	}
	if len(listResp.GetPayload().Data) == 0 {
		t.Fatalf("expected CA %q in list, got 0 results", name)
	}
	if *listResp.GetPayload().Data[0].Name != name {
		t.Errorf("expected name %q, got %q", name, *listResp.GetPayload().Data[0].Name)
	}
}

func TestGetCA(t *testing.T) {
	ctx := context.Background()
	name := "test-ca-get-" + uniqueSuffix()
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.CertificateAuthority.CreateCa(
		mgmtCA.NewCreateCaParams().WithContext(ctx).WithCa(&rest_model.CaCreate{
			Name:                      &name,
			CertPem:                   &certPEM,
			IsAuthEnabled:             ptr(false),
			IsAutoCaEnrollmentEnabled: ptr(false),
			IsOttCaEnrollmentEnabled:  ptr(false),
			IdentityRoles:             rest_model.Roles{},
		}), nil)
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	caID := createResp.GetPayload().Data.ID
	defer deleteCA(t, ctx, caID)

	getResp, err := mgmt.CertificateAuthority.DetailCa(
		mgmtCA.NewDetailCaParams().WithContext(ctx).WithID(caID), nil)
	if err != nil {
		t.Fatalf("get CA: %v", err)
	}
	if *getResp.GetPayload().Data.ID != caID {
		t.Errorf("expected id %q, got %q", caID, *getResp.GetPayload().Data.ID)
	}
}

func TestUpdateCA(t *testing.T) {
	ctx := context.Background()
	name := "test-ca-update-" + uniqueSuffix()
	updatedName := name + "-updated"
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.CertificateAuthority.CreateCa(
		mgmtCA.NewCreateCaParams().WithContext(ctx).WithCa(&rest_model.CaCreate{
			Name:                      &name,
			CertPem:                   &certPEM,
			IsAuthEnabled:             ptr(false),
			IsAutoCaEnrollmentEnabled: ptr(false),
			IsOttCaEnrollmentEnabled:  ptr(false),
			IdentityRoles:             rest_model.Roles{},
		}), nil)
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	caID := createResp.GetPayload().Data.ID
	defer deleteCA(t, ctx, caID)

	emptyFormat := ""
	_, err = mgmt.CertificateAuthority.UpdateCa(
		mgmtCA.NewUpdateCaParams().WithContext(ctx).WithID(caID).WithCa(&rest_model.CaUpdate{
			Name:                      &updatedName,
			IsAuthEnabled:             ptr(false),
			IsAutoCaEnrollmentEnabled: ptr(false),
			IsOttCaEnrollmentEnabled:  ptr(false),
			IdentityRoles:             rest_model.Roles{},
			IdentityNameFormat:        &emptyFormat,
		}), nil)
	if err != nil {
		t.Fatalf("update CA: %v", err)
	}

	getResp, err := mgmt.CertificateAuthority.DetailCa(
		mgmtCA.NewDetailCaParams().WithContext(ctx).WithID(caID), nil)
	if err != nil {
		t.Fatalf("get CA after update: %v", err)
	}
	if *getResp.GetPayload().Data.Name != updatedName {
		t.Errorf("expected name %q, got %q", updatedName, *getResp.GetPayload().Data.Name)
	}
}

func TestDeleteCA(t *testing.T) {
	ctx := context.Background()
	name := "test-ca-delete-" + uniqueSuffix()
	certPEM := generateCACert(t)

	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Fatal(err)
	}

	createResp, err := mgmt.CertificateAuthority.CreateCa(
		mgmtCA.NewCreateCaParams().WithContext(ctx).WithCa(&rest_model.CaCreate{
			Name:                      &name,
			CertPem:                   &certPEM,
			IsAuthEnabled:             ptr(false),
			IsAutoCaEnrollmentEnabled: ptr(false),
			IsOttCaEnrollmentEnabled:  ptr(false),
			IdentityRoles:             rest_model.Roles{},
		}), nil)
	if err != nil {
		t.Fatalf("create CA: %v", err)
	}
	caID := createResp.GetPayload().Data.ID

	if _, err := mgmt.CertificateAuthority.DeleteCa(
		mgmtCA.NewDeleteCaParams().WithContext(ctx).WithID(caID), nil); err != nil {
		t.Fatalf("delete CA: %v", err)
	}

	listResp, err := mgmt.CertificateAuthority.ListCas(
		mgmtCA.NewListCasParams().WithContext(ctx).WithFilter(ptr(`name = "`+name+`"`)), nil)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(listResp.GetPayload().Data) > 0 {
		t.Error("expected CA to be deleted, but it still appears in list")
	}
}

// deleteCA is a cleanup helper.
func deleteCA(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	mgmt, err := testClient.Mgmt()
	if err != nil {
		t.Logf("WARN: cleanup CA %q: get client: %v", id, err)
		return
	}
	if _, err := mgmt.CertificateAuthority.DeleteCa(
		mgmtCA.NewDeleteCaParams().WithContext(ctx).WithID(id), nil); err != nil {
		t.Logf("WARN: cleanup CA %q: %v", id, err)
	}
}

// generateCACert returns a PEM-encoded self-signed CA certificate.
func generateCACert(t *testing.T) string {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
}
