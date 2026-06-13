package models

import "time"

type VerifierAccount struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Organisation string    `json:"organisation"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WebhookEndpoint struct {
	ID         string    `json:"id"`
	VerifierID string    `json:"verifier_id"`
	URL        string    `json:"url"`
	Secret     string    `json:"-"`
	Events     string    `json:"events"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

type CredentialVisibility struct {
	CredentialID      string    `json:"credential_id"`
	ShowLicenceNumber bool      `json:"show_licence_number"`
	ShowExpiry        bool      `json:"show_expiry"`
	ShowVerifiedAt    bool      `json:"show_verified_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// VerifiableCredential is a W3C VC Data Model v1.1 representation.
type VerifiableCredential struct {
	Context           []string               `json:"@context"`
	ID                string                 `json:"id"`
	Type              []string               `json:"type"`
	Issuer            string                 `json:"issuer"`
	IssuanceDate      string                 `json:"issuanceDate"`
	ExpirationDate    string                 `json:"expirationDate,omitempty"`
	CredentialSubject map[string]interface{} `json:"credentialSubject"`
}
