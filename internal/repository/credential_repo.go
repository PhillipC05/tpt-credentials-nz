package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PassThePlat/TPT-NZ-Public/packages/app-credentials/internal/models"
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
	cred.CreatedAt = now
	cred.UpdatedAt = now

	return r.db.QueryRow(
		`INSERT INTO credentials (user_id, professional_body_id, licence_number, full_name, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
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
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("credential not found: %s", id)
	}
	return nil
}

// UpdateCredentialVerification updates the verification timestamp and status.
func (r *CredentialRepository) UpdateCredentialVerification(id string, status models.CredentialStatus, verifiedAt time.Time) error {
	result, err := r.db.Exec(
		`UPDATE credentials SET status = $1, verified_at = $2, last_checked_at = $2, updated_at = $2 WHERE id = $3`,
		status, verifiedAt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update credential verification: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("credential not found: %s", id)
	}
	return nil
}

// --- QR Tokens ---

// CreateQRToken inserts a new QR token for a credential.
func (r *CredentialRepository) CreateQRToken(token *models.QRToken) error {
	return r.db.QueryRow(
		`INSERT INTO qr_tokens (credential_id, token, expires_at, created_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		token.CredentialID, token.Token, token.ExpiresAt, token.CreatedAt,
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
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("QR token already used or not found: %s", id)
	}
	return nil
}