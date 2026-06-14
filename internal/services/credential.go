package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/registry"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/repository"
)

// RegistryClientFunc resolves a registry client for a professional body slug.
// Override in tests to avoid real HTTP calls.
type RegistryClientFunc func(slug string) registry.Client

// CredentialService handles business logic for professional credentials
// including verification against external professional body registries.
type CredentialService struct {
	repo          repository.Store
	qrSvc         *QRService
	registryClient RegistryClientFunc
}

// NewCredentialService creates a new CredentialService.
func NewCredentialService(repo repository.Store, qrSvc *QRService) *CredentialService {
	return &CredentialService{
		repo:          repo,
		qrSvc:         qrSvc,
		registryClient: registry.BySlug,
	}
}

// WithRegistryClient overrides the registry client factory (for tests).
func (s *CredentialService) WithRegistryClient(fn RegistryClientFunc) *CredentialService {
	s.registryClient = fn
	return s
}

// CreateCredential links a user's RealMe identity to a professional licence.
func (s *CredentialService) CreateCredential(userID string, req *models.CreateCredentialRequest) (*models.Credential, error) {
	pb, err := s.repo.GetProfessionalBodyBySlug(req.ProfessionalBodySlug)
	if err != nil {
		return nil, fmt.Errorf("invalid professional body: %w", err)
	}

	cred := &models.Credential{
		UserID:             userID,
		ProfessionalBodyID: pb.ID,
		LicenceNumber:      req.LicenceNumber,
		FullName:           req.FullName,
		Status:             models.CredentialStatusPending,
	}

	if err := s.repo.CreateCredential(cred); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	_ = s.repo.CreateCredentialEvent(&models.CredentialEvent{
		CredentialID: cred.ID,
		EventType:    models.EventCreated,
		ToStatus:     string(models.CredentialStatusPending),
	})

	go func() {
		if _, err := s.VerifyCredentialWithBody(cred); err != nil {
			log.Printf("initial verification failed for credential %s: %v", cred.ID, err)
		}
	}()

	return cred, nil
}

// VerifyCredentialWithBody checks the credential against the external registry
// and updates the stored status accordingly.
func (s *CredentialService) VerifyCredentialWithBody(cred *models.Credential) (bool, error) {
	pb, err := s.repo.GetProfessionalBodyByID(cred.ProfessionalBodyID)
	if err != nil {
		return false, fmt.Errorf("failed to get professional body: %w", err)
	}

	client := s.registryClient(pb.Slug)
	result, err := client.Verify(cred.LicenceNumber, cred.FullName)

	now := time.Now()
	prevStatus := string(cred.Status)

	if err != nil {
		log.Printf("registry verification error for %s/%s: %v", pb.Slug, cred.LicenceNumber, err)
		cred.LastCheckedAt = &now
		_ = s.repo.UpdateCredentialVerification(cred.ID, models.CredentialStatusPending, now, cred.ExpiresAt)
		return false, fmt.Errorf("external verification failed: %w", err)
	}

	if result.Verified {
		cred.Status = models.CredentialStatusActive
		cred.VerifiedAt = &now
		cred.LastCheckedAt = &now
		expiry := now.AddDate(1, 0, 0)
		cred.ExpiresAt = &expiry
		if err := s.repo.UpdateCredentialVerification(cred.ID, models.CredentialStatusActive, now, &expiry); err != nil {
			return false, fmt.Errorf("failed to update credential: %w", err)
		}
		_ = s.repo.CreateCredentialEvent(&models.CredentialEvent{
			CredentialID: cred.ID,
			EventType:    models.EventVerified,
			FromStatus:   prevStatus,
			ToStatus:     string(models.CredentialStatusActive),
			Notes:        result.Notes,
		})
		go s.deliverWebhooks("credential.verified", cred)
	} else {
		cred.LastCheckedAt = &now
		_ = s.repo.UpdateCredentialVerification(cred.ID, models.CredentialStatusPending, now, nil)
	}

	return result.Verified, nil
}

// RefreshCredentialStatus performs a live re-check and returns the updated credential.
func (s *CredentialService) RefreshCredentialStatus(credID string) (*models.Credential, error) {
	cred, err := s.repo.GetCredentialByID(credID)
	if err != nil {
		return nil, err
	}

	prevStatus := string(cred.Status)
	if _, err := s.VerifyCredentialWithBody(cred); err != nil {
		return nil, fmt.Errorf("failed to refresh credential: %w", err)
	}

	updated, err := s.repo.GetCredentialByID(credID)
	if err != nil {
		return nil, err
	}

	if string(updated.Status) != prevStatus {
		_ = s.repo.CreateCredentialEvent(&models.CredentialEvent{
			CredentialID: credID,
			EventType:    models.EventRefreshed,
			FromStatus:   prevStatus,
			ToStatus:     string(updated.Status),
		})
	}
	return updated, nil
}

// RevokeCredential sets a credential status to revoked and fires webhooks.
func (s *CredentialService) RevokeCredential(credID string) error {
	cred, err := s.repo.GetCredentialByID(credID)
	if err != nil {
		return err
	}
	prevStatus := string(cred.Status)
	if err := s.repo.UpdateCredentialStatus(credID, models.CredentialStatusRevoked); err != nil {
		return err
	}
	_ = s.repo.CreateCredentialEvent(&models.CredentialEvent{
		CredentialID: credID,
		EventType:    models.EventRevoked,
		FromStatus:   prevStatus,
		ToStatus:     string(models.CredentialStatusRevoked),
	})
	cred.Status = models.CredentialStatusRevoked
	go s.deliverWebhooks("credential.revoked", cred)
	return nil
}

// GetCredential returns a credential by ID.
func (s *CredentialService) GetCredential(credID string) (*models.Credential, error) {
	return s.repo.GetCredentialByID(credID)
}

// ListCredentials returns all credentials for a user.
func (s *CredentialService) ListCredentials(userID string) ([]models.Credential, error) {
	return s.repo.ListCredentialsByUser(userID)
}

// GetPublicCredential returns a sanitised credential view honouring visibility settings.
func (s *CredentialService) GetPublicCredential(credID string) (*models.CredentialResponse, error) {
	cred, err := s.repo.GetCredentialByID(credID)
	if err != nil {
		return nil, err
	}

	if cred.Status == models.CredentialStatusRevoked {
		return nil, fmt.Errorf("credential is revoked")
	}

	pb, err := s.repo.GetProfessionalBodyByID(cred.ProfessionalBodyID)
	if err != nil {
		return nil, err
	}

	vis, err := s.repo.GetCredentialVisibility(credID)
	if err != nil {
		vis = &models.CredentialVisibility{ShowLicenceNumber: true, ShowExpiry: true, ShowVerifiedAt: true}
	}

	resp := &models.CredentialResponse{
		ID:           cred.ID,
		Professional: pb.Name,
		FullName:     cred.FullName,
		Status:       cred.Status,
	}
	if vis.ShowLicenceNumber {
		resp.LicenceNumber = cred.LicenceNumber
	}
	if vis.ShowVerifiedAt {
		resp.VerifiedAt = cred.VerifiedAt
	}
	if vis.ShowExpiry {
		resp.ExpiresAt = cred.ExpiresAt
	}
	return resp, nil
}

// GetCredentialEvents returns the lifecycle timeline for a credential.
func (s *CredentialService) GetCredentialEvents(credID string) ([]models.CredentialEvent, error) {
	return s.repo.ListCredentialEvents(credID)
}

// GetVerifiableCredential returns a W3C VC Data Model v1.1 representation.
func (s *CredentialService) GetVerifiableCredential(credID string) (*models.VerifiableCredential, error) {
	cred, err := s.repo.GetCredentialByID(credID)
	if err != nil {
		return nil, err
	}
	pb, err := s.repo.GetProfessionalBodyByID(cred.ProfessionalBodyID)
	if err != nil {
		return nil, err
	}

	vc := &models.VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://tpt.nz/credentials/v1",
		},
		ID:   fmt.Sprintf("https://tpt.nz/credentials/%s", cred.ID),
		Type: []string{"VerifiableCredential", "ProfessionalLicenceCredential"},
		Issuer: "https://tpt.nz",
		CredentialSubject: map[string]interface{}{
			"fullName":         cred.FullName,
			"licenceNumber":    cred.LicenceNumber,
			"professionalBody": pb.Name,
			"status":           string(cred.Status),
		},
	}
	if cred.VerifiedAt != nil {
		vc.IssuanceDate = cred.VerifiedAt.UTC().Format(time.RFC3339)
	} else {
		vc.IssuanceDate = cred.CreatedAt.UTC().Format(time.RFC3339)
	}
	if cred.ExpiresAt != nil {
		vc.ExpirationDate = cred.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return vc, nil
}

// UpdateVisibility sets selective disclosure preferences for a credential.
func (s *CredentialService) UpdateVisibility(credID string, v *models.CredentialVisibility) error {
	v.CredentialID = credID
	if err := s.repo.UpsertCredentialVisibility(v); err != nil {
		return err
	}
	_ = s.repo.CreateCredentialEvent(&models.CredentialEvent{
		CredentialID: credID,
		EventType:    models.EventVisibilityChanged,
	})
	return nil
}

// GetVisibility returns the current visibility settings for a credential.
func (s *CredentialService) GetVisibility(credID string) (*models.CredentialVisibility, error) {
	return s.repo.GetCredentialVisibility(credID)
}

// deliverWebhooks fires all active webhook endpoints for the given event type.
func (s *CredentialService) deliverWebhooks(event string, cred *models.Credential) {
	endpoints, err := s.repo.ListActiveWebhookEndpoints()
	if err != nil {
		log.Printf("webhook: failed to list endpoints: %v", err)
		return
	}

	payload, err := json.Marshal(map[string]interface{}{
		"event":         event,
		"credential_id": cred.ID,
		"status":        string(cred.Status),
		"occurred_at":   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	for _, ep := range endpoints {
		if !strings.Contains(ep.Events, event) {
			continue
		}

		sig := webhookSignature(ep.Secret, payload)
		req, err := http.NewRequest("POST", ep.URL, bytes.NewReader(payload))
		if err != nil {
			log.Printf("webhook: bad endpoint URL %s: %v", ep.URL, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-TPT-Signature", "sha256="+sig)
		req.Header.Set("X-TPT-Event", event)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("webhook: delivery to %s failed: %v", ep.URL, err)
			continue
		}
		resp.Body.Close()
		log.Printf("webhook: delivered %s to %s — status %d", event, ep.URL, resp.StatusCode)
	}
}

func webhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
