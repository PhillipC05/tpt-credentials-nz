package services

import (
	"fmt"
	"log"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/repository"
)

// CredentialService handles business logic for professional credentials
// including verification against external professional body registries.
type CredentialService struct {
	repo     *repository.CredentialRepository
	qrSvc    *QRService
}

// NewCredentialService creates a new CredentialService.
func NewCredentialService(repo *repository.CredentialRepository, qrSvc *QRService) *CredentialService {
	return &CredentialService{
		repo:  repo,
		qrSvc: qrSvc,
	}
}

// CreateCredential links a user's RealMe identity to a professional licence.
// It performs an initial verification check against the professional body's
// registry.
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

	// Perform initial verification asynchronously
	go func() {
		if err := s.VerifyCredentialWithBody(cred); err != nil {
			log.Printf("initial verification failed for credential %s: %v", cred.ID, err)
		}
	}()

	return cred, nil
}

// VerifyCredentialWithBody checks the credential status against the external
// professional body registry. This simulates calling an external API; in
// production this would make HTTP requests to the body's verification endpoint.
func (s *CredentialService) VerifyCredentialWithBody(cred *models.Credential) (bool, error) {
	pb, err := s.repo.GetProfessionalBodyByID(cred.ProfessionalBodyID)
	if err != nil {
		return false, fmt.Errorf("failed to get professional body: %w", err)
	}

	// Simulate an external API call to the professional body's registry.
	// In production, this would make an HTTP GET to pb.BaseURL with the
	// licence number and parse the response.
	verified, status, err := s.callExternalRegistry(pb, cred.LicenceNumber, cred.FullName)
	if err != nil {
		return false, fmt.Errorf("external verification failed: %w", err)
	}

	now := time.Now()
	if verified {
		cred.Status = status
		cred.VerifiedAt = &now
		cred.LastCheckedAt = &now

		// Set a default expiry (e.g., 1 year from verification)
		expiry := now.AddDate(1, 0, 0)
		cred.ExpiresAt = &expiry

		if err := s.repo.UpdateCredentialVerification(cred.ID, status, now); err != nil {
			return false, fmt.Errorf("failed to update credential verification: %w", err)
		}
	} else {
		cred.Status = models.CredentialStatusPending
		cred.LastCheckedAt = &now
		if err := s.repo.UpdateCredentialVerification(cred.ID, models.CredentialStatusPending, now); err != nil {
			return false, fmt.Errorf("failed to update credential status: %w", err)
		}
	}

	return verified, nil
}

// RefreshCredentialStatus performs a live status check against the
// professional body's registry and updates the credential accordingly.
func (s *CredentialService) RefreshCredentialStatus(credID string) (*models.Credential, error) {
	cred, err := s.repo.GetCredentialByID(credID)
	if err != nil {
		return nil, err
	}

	if _, err := s.VerifyCredentialWithBody(cred); err != nil {
		return nil, fmt.Errorf("failed to refresh credential: %w", err)
	}

	return s.repo.GetCredentialByID(credID)
}

// RevokeCredential sets a credential status to revoked.
func (s *CredentialService) RevokeCredential(credID string) error {
	return s.repo.UpdateCredentialStatus(credID, models.CredentialStatusRevoked)
}

// GetCredential returns a credential by ID.
func (s *CredentialService) GetCredential(credID string) (*models.Credential, error) {
	return s.repo.GetCredentialByID(credID)
}

// ListCredentials returns all credentials for a user.
func (s *CredentialService) ListCredentials(userID string) ([]models.Credential, error) {
	creds, err := s.repo.ListCredentialsByUser(userID)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

// GetPublicCredential returns a sanitised credential view for public
// verification. No user ID is exposed.
func (s *CredentialService) GetPublicCredential(credID string) (*models.CredentialResponse, error) {
	cred, err := s.repo.GetCredentialByID(credID)
	if err != nil {
		return nil, err
	}

	pb, err := s.repo.GetProfessionalBodyByID(cred.ProfessionalBodyID)
	if err != nil {
		return nil, err
	}

	return &models.CredentialResponse{
		ID:            cred.ID,
		Professional:  pb.Name,
		FullName:      cred.FullName,
		LicenceNumber: cred.LicenceNumber,
		Status:        cred.Status,
		VerifiedAt:    cred.VerifiedAt,
		ExpiresAt:     cred.ExpiresAt,
	}, nil
}

// callExternalRegistry simulates an HTTP request to a professional body's
// verification API. In production, this would implement the specific API
// protocol for each body (e.g., NZ Medical Council, NZ Law Society).
//
// This is a stub that always returns verified=true for valid-looking licence
// numbers. Real implementations should handle:
//   - HTTP timeouts and retries
//   - API-specific authentication
//   - Response parsing for each body's format
//   - Rate limiting
//   - Circuit breaking for unavailable services
func (s *CredentialService) callExternalRegistry(pb *models.ProfessionalBody, licenceNumber, fullName string) (bool, models.CredentialStatus, error) {
	log.Printf("verifying licence %s with %s (%s)", licenceNumber, pb.Name, pb.BaseURL)

	// Simulate network latency
	time.Sleep(100 * time.Millisecond)

	// Stub: Always succeed for non-empty licence numbers
	if licenceNumber == "" {
		return false, models.CredentialStatusPending, fmt.Errorf("empty licence number")
	}

	return true, models.CredentialStatusActive, nil
}