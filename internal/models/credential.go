package models

import "time"

// ProfessionalBody represents a registered professional organisation such as
// the New Zealand Medical Council, New Zealand Law Society, etc.
type ProfessionalBody struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	BaseURL   string    `json:"base_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CredentialStatus represents the current state of a professional credential.
type CredentialStatus string

const (
	CredentialStatusActive    CredentialStatus = "active"
	CredentialStatusPending   CredentialStatus = "pending"
	CredentialStatusRevoked   CredentialStatus = "revoked"
	CredentialStatusExpired   CredentialStatus = "expired"
)

// Credential represents a verified professional licence or registration.
type Credential struct {
	ID                 string           `json:"id"`
	UserID             string           `json:"user_id"`
	ProfessionalBodyID string           `json:"professional_body_id"`
	LicenceNumber      string           `json:"licence_number"`
	FullName           string           `json:"full_name"`
	Status             CredentialStatus `json:"status"`
	VerifiedAt         *time.Time       `json:"verified_at,omitempty"`
	ExpiresAt          *time.Time       `json:"expires_at,omitempty"`
	LastCheckedAt      *time.Time       `json:"last_checked_at,omitempty"`
	Metadata           string           `json:"metadata,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// QRToken represents a one-time-use or time-limited token embedded in a QR code
// used for third-party verification of a credential.
type QRToken struct {
	ID           string     `json:"id"`
	CredentialID string     `json:"credential_id"`
	Token        string     `json:"token"`
	ExpiresAt    time.Time  `json:"expires_at"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// CreateCredentialRequest is the payload for linking a RealMe identity to
// a professional licence.
type CreateCredentialRequest struct {
	ProfessionalBodySlug string `json:"professional_body_slug"`
	LicenceNumber        string `json:"licence_number"`
	FullName             string `json:"full_name"`
}

// CredentialResponse is the public-facing representation of a credential.
type CredentialResponse struct {
	ID            string           `json:"id"`
	Professional  string           `json:"professional"`
	FullName      string           `json:"full_name"`
	LicenceNumber string           `json:"licence_number"`
	Status        CredentialStatus `json:"status"`
	VerifiedAt    *time.Time       `json:"verified_at,omitempty"`
	ExpiresAt     *time.Time       `json:"expires_at,omitempty"`
}

// VerifyResponse is the payload returned when a QR code is scanned.
type VerifyResponse struct {
	Valid        bool              `json:"valid"`
	FullName     string            `json:"full_name"`
	Professional string            `json:"professional"`
	LicenceNumber string           `json:"licence_number"`
	Status       CredentialStatus  `json:"status"`
	VerifiedAt   *time.Time        `json:"verified_at,omitempty"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
}