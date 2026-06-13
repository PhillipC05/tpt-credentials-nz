package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
	"github.com/google/uuid"
)

// CredentialRepository handles database operations for credentials, professional
// bodies, and QR tokens.
type CredentialRepository struct {
	db *sql.DB
}

// NewCredentialRepository creates a new CredentialRepository.
func NewCredentialRepository(db *sql.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

// --- Professional Bodies ---

// GetProfessionalBodyBySlug retrieves a professional body by its slug.
func (r *CredentialRepository) GetProfessionalBodyBySlug(slug string) (*models.ProfessionalBody, error) {
	row := r.db.QueryRow(
		`SELECT id, name, slug, base_url, created_at, updated_at
		 FROM professional_bodies
		 WHERE slug = $1`,
		slug,
	)

	var pb models.ProfessionalBody
	err := row.Scan(&pb.ID, &pb.Name, &pb.Slug, &pb.BaseURL, &pb.CreatedAt, &pb.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("professional body not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to get professional body: %w", err)
	}
	return &pb, nil
}

// GetProfessionalBodyByID retrieves a professional body by its ID.
func (r *CredentialRepository) GetProfessionalBodyByID(id string) (*models.ProfessionalBody, error) {
	row := r.db.QueryRow(
		`SELECT id, name, slug, base_url, created_at, updated_at
		 FROM professional_bodies
		 WHERE id = $1`,
		id,
	)

	var pb models.ProfessionalBody
	err := row.Scan(&pb.ID, &pb.Name, &pb.Slug, &pb.BaseURL, &pb.CreatedAt, &pb.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("professional body not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get professional body: %w", err)
	}
	return &pb, nil
}

// ListProfessionalBodies returns all registered professional bodies.
func (r *CredentialRepository) ListProfessionalBodies() ([]models.ProfessionalBody, error) {
	rows, err := r.db.Query(
		`SELECT id, name, slug, base_url, created_at, updated_at
		 FROM professional_bodies
		 ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list professional bodies: %w", err)
	}
	defer rows.Close()

	var bodies []models.ProfessionalBody
	for rows.Next() {
		var pb models.ProfessionalBody
		if err := rows.Scan(&pb.ID, &pb.Name, &pb.Slug, &pb.BaseURL, &pb.CreatedAt, &pb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan professional body: %w", err)
		}
		bodies = append(bodies, pb)
	}
	return bodies, rows.Err()
}

// --- Credentials ---

// CreateCredential inserts a new credential for a user.
func (r *CredentialRepository) CreateCredential(cred *models.Credential) error {
	now := time.Now()
	cred.ID = uuid.New().String()
	cred.CreatedAt = now
	cred.UpdatedAt = now

	return r.db.QueryRow(
		`INSERT INTO credentials (id, user_id, professional_body_id, licence_number, full_name, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		cred.ID,
		cred.UserID,
		cred.ProfessionalBodyID,
		cred.LicenceNumber,
		cred.FullName,
		cred.Status,
		cred.CreatedAt,
		cred.UpdatedAt,
	).Scan(&cred.ID)
}

// GetCredentialByID retrieves a single credential by ID.
func (r *CredentialRepository) GetCredentialByID(id string) (*models.Credential, error) {
	row := r.db.QueryRow(
		`SELECT c.id, c.user_id, c.professional_body_id, c.licence_number, c.full_name,
		        c.status, c.verified_at, c.expires_at, c.last_checked_at, c.metadata,
		        c.created_at, c.updated_at
		 FROM credentials c
		 WHERE c.id = $1`,
		id,
	)

	var cred models.Credential
	err := row.Scan(
		&cred.ID, &cred.UserID, &cred.ProfessionalBodyID, &cred.LicenceNumber,
		&cred.FullName, &cred.Status, &cred.VerifiedAt, &cred.ExpiresAt,
		&cred.LastCheckedAt, &cred.Metadata, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("credential not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	return &cred, nil
}

// ListCredentialsByUser returns all credentials belonging to a user.
func (r *CredentialRepository) ListCredentialsByUser(userID string) ([]models.Credential, error) {
	rows, err := r.db.Query(
		`SELECT c.id, c.user_id, c.professional_body_id, c.licence_number, c.full_name,
		        c.status, c.verified_at, c.expires_at, c.last_checked_at, c.metadata,
		        c.created_at, c.updated_at
		 FROM credentials c
		 WHERE c.user_id = $1
		 ORDER BY c.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	defer rows.Close()

	var creds []models.Credential
	for rows.Next() {
		var cred models.Credential
		if err := rows.Scan(
			&cred.ID, &cred.UserID, &cred.ProfessionalBodyID, &cred.LicenceNumber,
			&cred.FullName, &cred.Status, &cred.VerifiedAt, &cred.ExpiresAt,
			&cred.LastCheckedAt, &cred.Metadata, &cred.CreatedAt, &cred.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		creds = append(creds, cred)
	}
	return creds, rows.Err()
}

// UpdateCredentialStatus updates the status of a credential.
func (r *CredentialRepository) UpdateCredentialStatus(id string, status models.CredentialStatus) error {
	result, err := r.db.Exec(
		`UPDATE credentials SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to update credential status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("credential not found: %s", id)
	}
	return nil
}

// UpdateCredentialVerification updates the verification timestamp, expiry, and status.
func (r *CredentialRepository) UpdateCredentialVerification(id string, status models.CredentialStatus, verifiedAt time.Time, expiresAt *time.Time) error {
	result, err := r.db.Exec(
		`UPDATE credentials SET status = $1, verified_at = $2, last_checked_at = $2, expires_at = $3, updated_at = $2 WHERE id = $4`,
		status, verifiedAt, expiresAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update credential verification: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("credential not found: %s", id)
	}
	return nil
}

// --- QR Tokens ---

// CreateQRToken inserts a new QR token for a credential.
func (r *CredentialRepository) CreateQRToken(token *models.QRToken) error {
	token.ID = uuid.New().String()
	return r.db.QueryRow(
		`INSERT INTO qr_tokens (id, credential_id, token, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		token.ID, token.CredentialID, token.Token, token.ExpiresAt, token.CreatedAt,
	).Scan(&token.ID)
}

// GetQRTokenByToken retrieves a QR token by its token value.
func (r *CredentialRepository) GetQRTokenByToken(token string) (*models.QRToken, error) {
	row := r.db.QueryRow(
		`SELECT id, credential_id, token, expires_at, used_at, created_at
		 FROM qr_tokens
		 WHERE token = $1`,
		token,
	)

	var qt models.QRToken
	err := row.Scan(&qt.ID, &qt.CredentialID, &qt.Token, &qt.ExpiresAt, &qt.UsedAt, &qt.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("QR token not found: %s", token)
		}
		return nil, fmt.Errorf("failed to get QR token: %w", err)
	}
	return &qt, nil
}

// MarkQRTokenUsed marks a QR token as used at the current time.
func (r *CredentialRepository) MarkQRTokenUsed(id string) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE qr_tokens SET used_at = $1 WHERE id = $2 AND used_at IS NULL`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("failed to mark QR token as used: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("QR token already used or not found: %s", id)
	}
	return nil
}

// --- Audit Log ---

// CreateQRScanLog persists a QR scan audit entry.
func (r *CredentialRepository) CreateQRScanLog(entry *models.QRScanLog) error {
	entry.ID = uuid.New().String()
	if entry.ScannedAt.IsZero() {
		entry.ScannedAt = time.Now()
	}
	_, err := r.db.Exec(
		`INSERT INTO qr_scan_logs (id, token_id, credential_id, scanned_at, verifier_ip, success)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.ID, entry.TokenID, entry.CredentialID, entry.ScannedAt, entry.VerifierIP, entry.Success,
	)
	if err != nil {
		return fmt.Errorf("failed to create QR scan log: %w", err)
	}
	return nil
}

// --- Credential Events ---

// CreateCredentialEvent inserts a new lifecycle event for a credential.
func (r *CredentialRepository) CreateCredentialEvent(event *models.CredentialEvent) error {
	event.ID = uuid.New().String()
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	_, err := r.db.Exec(
		`INSERT INTO credential_events (id, credential_id, event_type, from_status, to_status, notes, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.ID, event.CredentialID, event.EventType, event.FromStatus, event.ToStatus, event.Notes, event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create credential event: %w", err)
	}
	return nil
}

// ListCredentialEvents returns all events for a credential in chronological order.
func (r *CredentialRepository) ListCredentialEvents(credentialID string) ([]models.CredentialEvent, error) {
	rows, err := r.db.Query(
		`SELECT id, credential_id, event_type, from_status, to_status, notes, created_at
		 FROM credential_events
		 WHERE credential_id = $1
		 ORDER BY created_at ASC`,
		credentialID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list credential events: %w", err)
	}
	defer rows.Close()

	var events []models.CredentialEvent
	for rows.Next() {
		var e models.CredentialEvent
		if err := rows.Scan(&e.ID, &e.CredentialID, &e.EventType, &e.FromStatus, &e.ToStatus, &e.Notes, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan credential event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// --- Verifier Accounts ---

// CreateVerifierAccount inserts a new verifier/employer account.
func (r *CredentialRepository) CreateVerifierAccount(account *models.VerifierAccount) error {
	account.ID = uuid.New().String()
	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now
	_, err := r.db.Exec(
		`INSERT INTO verifier_accounts (id, email, password_hash, organisation, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		account.ID, account.Email, account.PasswordHash, account.Organisation, account.CreatedAt, account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create verifier account: %w", err)
	}
	return nil
}

// GetVerifierAccountByEmail retrieves a verifier account by email.
func (r *CredentialRepository) GetVerifierAccountByEmail(email string) (*models.VerifierAccount, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, organisation, created_at, updated_at
		 FROM verifier_accounts WHERE email = $1`,
		email,
	)
	var a models.VerifierAccount
	if err := row.Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Organisation, &a.CreatedAt, &a.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("verifier account not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get verifier account: %w", err)
	}
	return &a, nil
}

// GetVerifierAccountByID retrieves a verifier account by ID.
func (r *CredentialRepository) GetVerifierAccountByID(id string) (*models.VerifierAccount, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, organisation, created_at, updated_at
		 FROM verifier_accounts WHERE id = $1`,
		id,
	)
	var a models.VerifierAccount
	if err := row.Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Organisation, &a.CreatedAt, &a.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("verifier account not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get verifier account: %w", err)
	}
	return &a, nil
}

// --- Webhooks ---

// CreateWebhookEndpoint inserts a new webhook endpoint for a verifier.
func (r *CredentialRepository) CreateWebhookEndpoint(endpoint *models.WebhookEndpoint) error {
	endpoint.ID = uuid.New().String()
	endpoint.CreatedAt = time.Now()
	_, err := r.db.Exec(
		`INSERT INTO webhook_endpoints (id, verifier_id, url, secret, events, active, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		endpoint.ID, endpoint.VerifierID, endpoint.URL, endpoint.Secret, endpoint.Events, endpoint.Active, endpoint.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create webhook endpoint: %w", err)
	}
	return nil
}

// ListWebhookEndpointsByVerifier returns all webhook endpoints for a verifier.
func (r *CredentialRepository) ListWebhookEndpointsByVerifier(verifierID string) ([]models.WebhookEndpoint, error) {
	rows, err := r.db.Query(
		`SELECT id, verifier_id, url, secret, events, active, created_at
		 FROM webhook_endpoints WHERE verifier_id = $1`,
		verifierID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook endpoints: %w", err)
	}
	defer rows.Close()
	return r.scanWebhookEndpoints(rows)
}

// ListActiveWebhookEndpoints returns all active webhook endpoints across all verifiers.
func (r *CredentialRepository) ListActiveWebhookEndpoints() ([]models.WebhookEndpoint, error) {
	rows, err := r.db.Query(
		`SELECT id, verifier_id, url, secret, events, active, created_at
		 FROM webhook_endpoints WHERE active = TRUE`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list active webhook endpoints: %w", err)
	}
	defer rows.Close()
	return r.scanWebhookEndpoints(rows)
}

func (r *CredentialRepository) scanWebhookEndpoints(rows *sql.Rows) ([]models.WebhookEndpoint, error) {
	var endpoints []models.WebhookEndpoint
	for rows.Next() {
		var e models.WebhookEndpoint
		if err := rows.Scan(&e.ID, &e.VerifierID, &e.URL, &e.Secret, &e.Events, &e.Active, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan webhook endpoint: %w", err)
		}
		endpoints = append(endpoints, e)
	}
	return endpoints, rows.Err()
}

// --- Selective Disclosure ---

// GetCredentialVisibility returns the visibility settings for a credential.
// Returns default (show all) if no settings exist.
func (r *CredentialRepository) GetCredentialVisibility(credentialID string) (*models.CredentialVisibility, error) {
	row := r.db.QueryRow(
		`SELECT credential_id, show_licence_number, show_expiry, show_verified_at, updated_at
		 FROM credential_visibility WHERE credential_id = $1`,
		credentialID,
	)
	var v models.CredentialVisibility
	if err := row.Scan(&v.CredentialID, &v.ShowLicenceNumber, &v.ShowExpiry, &v.ShowVerifiedAt, &v.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return &models.CredentialVisibility{
				CredentialID:      credentialID,
				ShowLicenceNumber: true,
				ShowExpiry:        true,
				ShowVerifiedAt:    true,
				UpdatedAt:         time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get credential visibility: %w", err)
	}
	return &v, nil
}

// UpsertCredentialVisibility creates or updates visibility settings for a credential.
func (r *CredentialRepository) UpsertCredentialVisibility(v *models.CredentialVisibility) error {
	v.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`INSERT INTO credential_visibility (credential_id, show_licence_number, show_expiry, show_verified_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (credential_id) DO UPDATE
		 SET show_licence_number = EXCLUDED.show_licence_number,
		     show_expiry = EXCLUDED.show_expiry,
		     show_verified_at = EXCLUDED.show_verified_at,
		     updated_at = EXCLUDED.updated_at`,
		v.CredentialID, v.ShowLicenceNumber, v.ShowExpiry, v.ShowVerifiedAt, v.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert credential visibility: %w", err)
	}
	return nil
}