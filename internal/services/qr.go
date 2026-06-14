package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/repository"
	qrcode "github.com/skip2/go-qrcode"
)

// QRService handles QR token generation and resolution for credential
// verification.
type QRService struct {
	repo          repository.Store
	tokenDuration time.Duration
}

// NewQRService creates a new QRService.
func NewQRService(repo repository.Store) *QRService {
	return &QRService{
		repo:          repo,
		tokenDuration: 30 * time.Minute,
	}
}

// GenerateQRToken creates a new cryptographically random token for a credential
// and returns it along with a base64-encoded QR code PNG.
func (s *QRService) GenerateQRToken(credID string) (string, string, error) {
	token, err := s.generateRandomToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	qrToken := &models.QRToken{
		CredentialID: credID,
		Token:        token,
		ExpiresAt:    time.Now().Add(s.tokenDuration),
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateQRToken(qrToken); err != nil {
		return "", "", fmt.Errorf("failed to save QR token: %w", err)
	}

	verifyURL := s.GetVerificationURL(token)
	qrPNG, err := qrcode.Encode(verifyURL, qrcode.Medium, 256)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	return token, base64.StdEncoding.EncodeToString(qrPNG), nil
}

// ResolveToken validates a QR token and returns the associated credential.
// Tokens can only be used once and must not be expired.
// Also returns the token DB ID and credential ID for audit logging.
func (s *QRService) ResolveToken(token string) (*models.CredentialResponse, string, string, error) {
	qt, err := s.repo.GetQRTokenByToken(token)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid token: %w", err)
	}

	if qt.UsedAt != nil {
		return nil, qt.ID, qt.CredentialID, fmt.Errorf("token has already been used")
	}

	if time.Now().After(qt.ExpiresAt) {
		return nil, qt.ID, qt.CredentialID, fmt.Errorf("token has expired")
	}

	if err := s.repo.MarkQRTokenUsed(qt.ID); err != nil {
		return nil, qt.ID, qt.CredentialID, fmt.Errorf("failed to mark token as used: %w", err)
	}

	cred, err := s.repo.GetCredentialByID(qt.CredentialID)
	if err != nil {
		return nil, qt.ID, qt.CredentialID, fmt.Errorf("failed to resolve credential: %w", err)
	}

	pb, err := s.repo.GetProfessionalBodyByID(cred.ProfessionalBodyID)
	if err != nil {
		return nil, qt.ID, qt.CredentialID, fmt.Errorf("failed to get professional body: %w", err)
	}

	return &models.CredentialResponse{
		ID:            cred.ID,
		Professional:  pb.Name,
		FullName:      cred.FullName,
		LicenceNumber: cred.LicenceNumber,
		Status:        cred.Status,
		VerifiedAt:    cred.VerifiedAt,
		ExpiresAt:     cred.ExpiresAt,
	}, qt.ID, qt.CredentialID, nil
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
// Uses the TPT_BASE_URL environment variable when set.
func (s *QRService) GetVerificationURL(token string) string {
	base := os.Getenv("TPT_BASE_URL")
	if base == "" {
		base = "https://tpt.nz"
	}
	return fmt.Sprintf("%s/verify/%s", base, token)
}

// LogQRScan persists a QR scan audit entry and writes a log line.
func (s *QRService) LogQRScan(tokenID, credentialID string, success bool, remoteAddr string) {
	status := "success"
	if !success {
		status = "failed"
	}
	tok := tokenID
	if len(tok) > 8 {
		tok = tok[:8]
	}
	log.Printf("QR scan [%s] token=%s remote=%s", status, tok, remoteAddr)

	entry := &models.QRScanLog{
		TokenID:      tokenID,
		CredentialID: credentialID,
		VerifierIP:   remoteAddr,
		Success:      success,
	}
	if err := s.repo.CreateQRScanLog(entry); err != nil {
		log.Printf("QR scan audit log write failed: %v", err)
	}
}