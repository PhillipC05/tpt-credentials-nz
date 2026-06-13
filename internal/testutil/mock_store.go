package testutil

import (
	"fmt"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/google/uuid"
)

// MockStore is an in-memory implementation of repository.Store for tests.
type MockStore struct {
	Bodies      map[string]*models.ProfessionalBody
	Creds       map[string]*models.Credential
	QRTokens    map[string]*models.QRToken
	ScanLogs    []*models.QRScanLog
	Events      map[string][]*models.CredentialEvent
	Verifiers   map[string]*models.VerifierAccount
	Webhooks    []*models.WebhookEndpoint
	Visibility  map[string]*models.CredentialVisibility
}

// NewMockStore returns a MockStore seeded with one professional body.
func NewMockStore() *MockStore {
	return &MockStore{
		Bodies: map[string]*models.ProfessionalBody{
			"pb1": {
				ID:      "pb1",
				Name:    "Medical Council of New Zealand",
				Slug:    "nz-medical-council",
				BaseURL: "https://example.com/api/mcnz",
			},
		},
		Creds:     make(map[string]*models.Credential),
		QRTokens:  make(map[string]*models.QRToken),
		Events:    make(map[string][]*models.CredentialEvent),
		Verifiers: make(map[string]*models.VerifierAccount),
		Visibility: make(map[string]*models.CredentialVisibility),
	}
}

func (m *MockStore) GetProfessionalBodyBySlug(slug string) (*models.ProfessionalBody, error) {
	for _, pb := range m.Bodies {
		if pb.Slug == slug {
			return pb, nil
		}
	}
	return nil, fmt.Errorf("professional body not found: %s", slug)
}

func (m *MockStore) GetProfessionalBodyByID(id string) (*models.ProfessionalBody, error) {
	if pb, ok := m.Bodies[id]; ok {
		return pb, nil
	}
	return nil, fmt.Errorf("professional body not found: %s", id)
}

func (m *MockStore) ListProfessionalBodies() ([]models.ProfessionalBody, error) {
	var result []models.ProfessionalBody
	for _, pb := range m.Bodies {
		result = append(result, *pb)
	}
	return result, nil
}

func (m *MockStore) CreateCredential(cred *models.Credential) error {
	if cred.ID == "" {
		cred.ID = uuid.New().String()
	}
	now := time.Now()
	cred.CreatedAt = now
	cred.UpdatedAt = now
	m.Creds[cred.ID] = cred
	return nil
}

func (m *MockStore) GetCredentialByID(id string) (*models.Credential, error) {
	if c, ok := m.Creds[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("credential not found: %s", id)
}

func (m *MockStore) ListCredentialsByUser(userID string) ([]models.Credential, error) {
	var result []models.Credential
	for _, c := range m.Creds {
		if c.UserID == userID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *MockStore) UpdateCredentialStatus(id string, status models.CredentialStatus) error {
	c, ok := m.Creds[id]
	if !ok {
		return fmt.Errorf("credential not found: %s", id)
	}
	c.Status = status
	return nil
}

func (m *MockStore) UpdateCredentialVerification(id string, status models.CredentialStatus, verifiedAt time.Time, expiresAt *time.Time) error {
	c, ok := m.Creds[id]
	if !ok {
		return fmt.Errorf("credential not found: %s", id)
	}
	c.Status = status
	c.VerifiedAt = &verifiedAt
	c.LastCheckedAt = &verifiedAt
	c.ExpiresAt = expiresAt
	return nil
}

func (m *MockStore) CreateQRToken(token *models.QRToken) error {
	m.QRTokens[token.Token] = token
	return nil
}

func (m *MockStore) GetQRTokenByToken(token string) (*models.QRToken, error) {
	if qt, ok := m.QRTokens[token]; ok {
		return qt, nil
	}
	return nil, fmt.Errorf("QR token not found: %s", token)
}

func (m *MockStore) MarkQRTokenUsed(id string) error {
	for _, qt := range m.QRTokens {
		if qt.ID == id {
			if qt.UsedAt != nil {
				return fmt.Errorf("QR token already used or not found: %s", id)
			}
			now := time.Now()
			qt.UsedAt = &now
			return nil
		}
	}
	return fmt.Errorf("QR token already used or not found: %s", id)
}

func (m *MockStore) CreateQRScanLog(entry *models.QRScanLog) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.ScannedAt.IsZero() {
		entry.ScannedAt = time.Now()
	}
	m.ScanLogs = append(m.ScanLogs, entry)
	return nil
}

func (m *MockStore) CreateCredentialEvent(event *models.CredentialEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	m.Events[event.CredentialID] = append(m.Events[event.CredentialID], event)
	return nil
}

func (m *MockStore) ListCredentialEvents(credentialID string) ([]models.CredentialEvent, error) {
	var result []models.CredentialEvent
	for _, e := range m.Events[credentialID] {
		result = append(result, *e)
	}
	return result, nil
}

func (m *MockStore) CreateVerifierAccount(account *models.VerifierAccount) error {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now
	m.Verifiers[account.ID] = account
	return nil
}

func (m *MockStore) GetVerifierAccountByEmail(email string) (*models.VerifierAccount, error) {
	for _, v := range m.Verifiers {
		if v.Email == email {
			return v, nil
		}
	}
	return nil, fmt.Errorf("verifier account not found: %s", email)
}

func (m *MockStore) GetVerifierAccountByID(id string) (*models.VerifierAccount, error) {
	if v, ok := m.Verifiers[id]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("verifier account not found: %s", id)
}

func (m *MockStore) CreateWebhookEndpoint(endpoint *models.WebhookEndpoint) error {
	if endpoint.ID == "" {
		endpoint.ID = uuid.New().String()
	}
	endpoint.CreatedAt = time.Now()
	m.Webhooks = append(m.Webhooks, endpoint)
	return nil
}

func (m *MockStore) ListWebhookEndpointsByVerifier(verifierID string) ([]models.WebhookEndpoint, error) {
	var result []models.WebhookEndpoint
	for _, wh := range m.Webhooks {
		if wh.VerifierID == verifierID {
			result = append(result, *wh)
		}
	}
	return result, nil
}

func (m *MockStore) ListActiveWebhookEndpoints() ([]models.WebhookEndpoint, error) {
	var result []models.WebhookEndpoint
	for _, wh := range m.Webhooks {
		if wh.Active {
			result = append(result, *wh)
		}
	}
	return result, nil
}

func (m *MockStore) GetCredentialVisibility(credentialID string) (*models.CredentialVisibility, error) {
	if v, ok := m.Visibility[credentialID]; ok {
		return v, nil
	}
	return &models.CredentialVisibility{
		CredentialID:      credentialID,
		ShowLicenceNumber: true,
		ShowExpiry:        true,
		ShowVerifiedAt:    true,
		UpdatedAt:         time.Now(),
	}, nil
}

func (m *MockStore) UpsertCredentialVisibility(v *models.CredentialVisibility) error {
	v.UpdatedAt = time.Now()
	m.Visibility[v.CredentialID] = v
	return nil
}
