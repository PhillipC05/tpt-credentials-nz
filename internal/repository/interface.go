package repository

import (
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
)

// Store is the data-access interface used by the service layer. The concrete
// CredentialRepository implements this; the testutil.MockStore implements it
// for unit tests.
type Store interface {
	// Professional bodies
	GetProfessionalBodyBySlug(slug string) (*models.ProfessionalBody, error)
	GetProfessionalBodyByID(id string) (*models.ProfessionalBody, error)
	ListProfessionalBodies() ([]models.ProfessionalBody, error)

	// Credentials
	CreateCredential(cred *models.Credential) error
	GetCredentialByID(id string) (*models.Credential, error)
	ListCredentialsByUser(userID string) ([]models.Credential, error)
	UpdateCredentialStatus(id string, status models.CredentialStatus) error
	UpdateCredentialVerification(id string, status models.CredentialStatus, verifiedAt time.Time, expiresAt *time.Time) error

	// QR tokens
	CreateQRToken(token *models.QRToken) error
	GetQRTokenByToken(token string) (*models.QRToken, error)
	MarkQRTokenUsed(id string) error

	// Audit log
	CreateQRScanLog(entry *models.QRScanLog) error

	// Credential events (timeline)
	CreateCredentialEvent(event *models.CredentialEvent) error
	ListCredentialEvents(credentialID string) ([]models.CredentialEvent, error)

	// Verifier accounts
	CreateVerifierAccount(account *models.VerifierAccount) error
	GetVerifierAccountByEmail(email string) (*models.VerifierAccount, error)
	GetVerifierAccountByID(id string) (*models.VerifierAccount, error)

	// Webhooks
	CreateWebhookEndpoint(endpoint *models.WebhookEndpoint) error
	ListWebhookEndpointsByVerifier(verifierID string) ([]models.WebhookEndpoint, error)
	ListActiveWebhookEndpoints() ([]models.WebhookEndpoint, error)

	// Selective disclosure
	GetCredentialVisibility(credentialID string) (*models.CredentialVisibility, error)
	UpsertCredentialVisibility(v *models.CredentialVisibility) error
}
