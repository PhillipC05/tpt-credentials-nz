package services

import (
	"testing"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/registry"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/testutil"
)

// alwaysVerifiedClient is a test-only registry client that always returns verified.
type alwaysVerifiedClient struct{}

func (a *alwaysVerifiedClient) Verify(_, _ string) (registry.Result, error) {
	return registry.Result{Verified: true, Notes: "test stub"}, nil
}

func newTestService(store *testutil.MockStore) *CredentialService {
	qrSvc := NewQRService(store)
	return NewCredentialService(store, qrSvc).WithRegistryClient(func(_ string) registry.Client {
		return &alwaysVerifiedClient{}
	})
}

func TestServiceCreateCredential(t *testing.T) {
	store := testutil.NewMockStore()
	svc := newTestService(store)

	cred, err := svc.CreateCredential("user-1", &models.CreateCredentialRequest{
		ProfessionalBodySlug: "nz-medical-council",
		LicenceNumber:        "MED123",
		FullName:             "Jane Smith",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.ID == "" {
		t.Error("credential ID must be set")
	}
	if cred.Status != models.CredentialStatusPending {
		t.Errorf("expected pending status, got %s", cred.Status)
	}
	if cred.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", cred.UserID)
	}
	if _, ok := store.Creds[cred.ID]; !ok {
		t.Error("credential not persisted in store")
	}
}

func TestServiceRefreshCredential(t *testing.T) {
	store := testutil.NewMockStore()
	store.Creds["cred-1"] = &models.Credential{
		ID:                 "cred-1",
		UserID:             "user-1",
		ProfessionalBodyID: "pb1",
		LicenceNumber:      "MED123",
		FullName:           "Jane Smith",
		Status:             models.CredentialStatusPending,
	}
	svc := newTestService(store)

	updated, err := svc.RefreshCredentialStatus("cred-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The stub callExternalRegistry always returns active for non-empty licence numbers.
	if updated.Status != models.CredentialStatusActive {
		t.Errorf("expected active after refresh, got %s", updated.Status)
	}
	if store.Creds["cred-1"].ExpiresAt == nil {
		t.Error("expires_at must be set after verification")
	}
}

func TestServiceRevokeCredential(t *testing.T) {
	store := testutil.NewMockStore()
	store.Creds["cred-1"] = &models.Credential{
		ID:     "cred-1",
		UserID: "user-1",
		Status: models.CredentialStatusActive,
	}
	svc := newTestService(store)

	if err := svc.RevokeCredential("cred-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.Creds["cred-1"].Status != models.CredentialStatusRevoked {
		t.Errorf("expected revoked, got %s", store.Creds["cred-1"].Status)
	}
}
