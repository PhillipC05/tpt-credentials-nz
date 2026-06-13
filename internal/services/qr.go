package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/repository"
	qrcode "github.com/skip2/go-qrcode"
)

// QRService handles QR token generation and resolution for credential
// verification.
type QRService struct {
	repo          *repository.CredentialRepository
	tokenDuration time.Duration
}

// NewQRService creates a new QRService.
func NewQRService(repo *repository.CredentialRepository) *QRService {
	return &QRService{
		repo:          repo,
		tokenDuration: 30 * time.Minute, // QR tokens valid for 30 minutes
	}
}

// GenerateQRToken creates a new cryptographically random token for a credential
// and returns it along with a QR code PNG.
func (s *QRService) GenerateQRToken(credID string) (string, []byte, error) {
	token, err := s.generateRandomToken()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	qrToken := &models.QRToken{
		CredentialID: credID,
		Token:        token,
		ExpiresAt:    time.Now().Add(s.tokenDuration),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateQRToken(qrToken); err != nil {
		return "", nil, fmt.Errorf("failed to save QR token: %w", err)
	}

	// Generate QR code as PNG bytes
	// The URL format should match the public verification endpoint
	verifyURL := fmt.Sprintf("https://tpt.nz/verify/%s", token)
	qrPNG, err := qrcode.Encode(verifyURL, qrcode.Medium, 256)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return token, qrPNG, nil
}

// ResolveToken validates a QR token and returns the associated credential.
// Tokens can only be used once and must not be expired.
func (s *QRService) ResolveToken(token string) (*models.CredentialResponse, error) {
	qt, err := s.repo.GetQRTokenByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token has been used
	if qt.UsedAt != nil {
		return nil, fmt.Errorf("token has already been used")
	}

	// Check if token has expired
	if time.Now().After(qt.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// Mark token as used (one-time use)
	if err := s.repo.MarkQRTokenUsed(qt.ID); err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Retrieve credential and build public response
	cred, err := s.repo.GetCredentialByID(qt.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve credential: %w", err)
	}

	pb, err := s.repo.GetProfessionalBodyByID(cred.ProfessionalBodyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get professional body: %w", err)
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

// generateRandomToken creates a cryptographically secure random hex string.
func (s *QRService) generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetVerificationURL returns the full verification URL for a given token.
func (s *QRService) GetVerificationURL(token string) string {
	return fmt.Sprintf("https://tpt.nz/verify/%s", token)
}

// LogQRScan logs a QR scan attempt for audit purposes.
func (s *QRService) LogQRScan(token string, success bool, remoteAddr string) {
	status := "success"
	if !success {
		status = "failed"
	}
	log.Printf("QR scan [%s] token=%s remote=%s", status, token[:8], remoteAddr)
}