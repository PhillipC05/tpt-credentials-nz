package services

import (
	"testing"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/testutil"
)

func TestQRServiceGenerateToken(t *testing.T) {
	store := testutil.NewMockStore()
	store.Creds["cred-1"] = &models.Credential{
		ID:     "cred-1",
		UserID: "user-1",
		Status: models.CredentialStatusActive,
	}
	svc := NewQRService(store)

	token, qrB64, err := svc.GenerateQRToken("cred-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token string")
	}
	if qrB64 == "" {
		t.Error("expected non-empty base64 QR code")
	}
	if _, ok := store.QRTokens[token]; !ok {
		t.Error("QR token not persisted in store")
	}
}

func TestQRServiceResolveToken(t *testing.T) {
	store := testutil.NewMockStore()
	store.Creds["cred-1"] = &models.Credential{
		ID:                 "cred-1",
		UserID:             "user-1",
		ProfessionalBodyID: "pb1",
		LicenceNumber:      "MED123",
		FullName:           "Jane Smith",
		Status:             models.CredentialStatusActive,
	}
	now := time.Now()
	store.QRTokens["valid-token"] = &models.QRToken{
		ID:           "qt-1",
		CredentialID: "cred-1",
		Token:        "valid-token",
		ExpiresAt:    now.Add(30 * time.Minute),
		CreatedAt:    now,
	}
	svc := NewQRService(store)

	resp, _, _, err := svc.ResolveToken("valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FullName != "Jane Smith" {
		t.Errorf("expected Jane Smith, got %s", resp.FullName)
	}
	if store.QRTokens["valid-token"].UsedAt == nil {
		t.Error("token must be marked used after resolution")
	}
}

func TestQRServiceExpiredToken(t *testing.T) {
	store := testutil.NewMockStore()
	store.QRTokens["expired-token"] = &models.QRToken{
		ID:           "qt-2",
		CredentialID: "cred-1",
		Token:        "expired-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	}
	svc := NewQRService(store)

	if _, _, _, err := svc.ResolveToken("expired-token"); err == nil {
		t.Error("expected error for expired token, got nil")
	}
}
