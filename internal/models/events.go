package models

import "time"

type QRScanLog struct {
	ID           string    `json:"id"`
	TokenID      string    `json:"token_id"`
	CredentialID string    `json:"credential_id"`
	ScannedAt    time.Time `json:"scanned_at"`
	VerifierIP   string    `json:"verifier_ip"`
	Success      bool      `json:"success"`
}

type CredentialEvent struct {
	ID           string    `json:"id"`
	CredentialID string    `json:"credential_id"`
	EventType    string    `json:"event_type"`
	FromStatus   string    `json:"from_status,omitempty"`
	ToStatus     string    `json:"to_status,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

const (
	EventCreated           = "created"
	EventVerified          = "verified"
	EventRefreshed         = "refreshed"
	EventRevoked           = "revoked"
	EventExpired           = "expired"
	EventVisibilityChanged = "visibility_changed"
)
