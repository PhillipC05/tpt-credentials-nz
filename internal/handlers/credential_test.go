package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/services"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/testutil"
	"github.com/gorilla/mux"
)

func newTestServices(store *testutil.MockStore) (*services.CredentialService, *services.QRService) {
	qrSvc := services.NewQRService(store)
	credSvc := services.NewCredentialService(store, qrSvc)
	return credSvc, qrSvc
}

func TestHandlersCredentialCreate(t *testing.T) {
	store := testutil.NewMockStore()
	credSvc, qrSvc := newTestServices(store)
	h := NewCredentialHandler(credSvc, qrSvc)

	body := `{"professional_body_slug":"nz-medical-council","licence_number":"MED123","full_name":"Jane Smith"}`
	req := httptest.NewRequest(http.MethodPost, "/api/credentials", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	h.CreateCredential(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.Credential
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID == "" {
		t.Error("response must include a credential ID")
	}
	if resp.Status != models.CredentialStatusPending {
		t.Errorf("expected pending, got %s", resp.Status)
	}
}

func TestHandlersCredentialList(t *testing.T) {
	store := testutil.NewMockStore()
	credSvc, qrSvc := newTestServices(store)
	h := NewCredentialHandler(credSvc, qrSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/credentials", nil)
	req.Header.Set("X-User-ID", "user-1")
	w := httptest.NewRecorder()

	h.ListCredentials(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp []models.Credential
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp == nil {
		t.Error("response must be an array, not null")
	}
}

func TestHandlersPublicVerify(t *testing.T) {
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
	store.QRTokens["test-token"] = &models.QRToken{
		ID:           "qt-1",
		CredentialID: "cred-1",
		Token:        "test-token",
		ExpiresAt:    now.Add(30 * time.Minute),
		CreatedAt:    now,
	}

	credSvc, qrSvc := newTestServices(store)
	h := NewPublicHandler(credSvc, qrSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/verify/test-token", nil)
	req = mux.SetURLVars(req, map[string]string{"qr_id": "test-token"})
	w := httptest.NewRecorder()

	h.VerifyCredential(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["valid"] != true {
		t.Errorf("expected valid=true, got %v", resp["valid"])
	}
	if resp["full_name"] != "Jane Smith" {
		t.Errorf("expected Jane Smith, got %v", resp["full_name"])
	}
}
