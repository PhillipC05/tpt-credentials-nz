SET search_path TO credentials;

BEGIN;

-- Persistent QR scan audit trail
CREATE TABLE IF NOT EXISTS qr_scan_logs (
    id            TEXT PRIMARY KEY,
    token_id      TEXT NOT NULL REFERENCES qr_tokens(id) ON DELETE CASCADE,
    credential_id TEXT NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    scanned_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verifier_ip   TEXT NOT NULL DEFAULT '',
    success       BOOLEAN NOT NULL DEFAULT TRUE
);

-- Credential lifecycle timeline
CREATE TABLE IF NOT EXISTS credential_events (
    id            TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    event_type    TEXT NOT NULL CHECK (event_type IN ('created', 'verified', 'refreshed', 'revoked', 'expired', 'visibility_changed')),
    from_status   TEXT DEFAULT '',
    to_status     TEXT DEFAULT '',
    notes         TEXT DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Verifier/employer accounts (B2B tier)
CREATE TABLE IF NOT EXISTS verifier_accounts (
    id            TEXT PRIMARY KEY,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    organisation  TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Webhook endpoints registered by verifiers
CREATE TABLE IF NOT EXISTS webhook_endpoints (
    id            TEXT PRIMARY KEY,
    verifier_id   TEXT NOT NULL REFERENCES verifier_accounts(id) ON DELETE CASCADE,
    url           TEXT NOT NULL,
    secret        TEXT NOT NULL,
    events        TEXT NOT NULL DEFAULT 'credential.revoked,credential.expired,credential.verified',
    active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Credentials shared with a verifier (roster tracking)
CREATE TABLE IF NOT EXISTS credential_shares (
    id            TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    verifier_id   TEXT NOT NULL REFERENCES verifier_accounts(id) ON DELETE CASCADE,
    shared_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(credential_id, verifier_id)
);

-- Visibility preferences per credential (selective disclosure)
CREATE TABLE IF NOT EXISTS credential_visibility (
    credential_id      TEXT PRIMARY KEY REFERENCES credentials(id) ON DELETE CASCADE,
    show_licence_number BOOLEAN NOT NULL DEFAULT TRUE,
    show_expiry         BOOLEAN NOT NULL DEFAULT TRUE,
    show_verified_at    BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_qr_scan_logs_credential_id ON qr_scan_logs(credential_id);
CREATE INDEX IF NOT EXISTS idx_qr_scan_logs_token_id ON qr_scan_logs(token_id);
CREATE INDEX IF NOT EXISTS idx_credential_events_credential_id ON credential_events(credential_id);
CREATE INDEX IF NOT EXISTS idx_credential_events_created_at ON credential_events(created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_endpoints_verifier_id ON webhook_endpoints(verifier_id);
CREATE INDEX IF NOT EXISTS idx_credential_shares_credential_id ON credential_shares(credential_id);
CREATE INDEX IF NOT EXISTS idx_credential_shares_verifier_id ON credential_shares(verifier_id);

COMMIT;
