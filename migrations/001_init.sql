SET search_path TO credentials;

-- Professional Credentials Wallet Database Schema
-- Supports linking RealMe-verified identities to professional licences
-- and QR-code-based third-party verification.

BEGIN;

-- Users table (shared auth model)
CREATE TABLE IF NOT EXISTS users (
    id          TEXT PRIMARY KEY,
    email       TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Professional bodies (e.g., NZ Medical Council, NZ Law Society, etc.)
CREATE TABLE IF NOT EXISTS professional_bodies (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT UNIQUE NOT NULL,
    base_url    TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Credentials linking a user's RealMe identity to a professional licence
CREATE TABLE IF NOT EXISTS credentials (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    professional_body_id TEXT NOT NULL REFERENCES professional_bodies(id),
    licence_number       TEXT NOT NULL,
    full_name            TEXT NOT NULL,
    status               TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('active', 'pending', 'revoked', 'expired')),
    verified_at          TIMESTAMPTZ,
    expires_at           TIMESTAMPTZ,
    last_checked_at      TIMESTAMPTZ,
    metadata             TEXT DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- QR tokens for one-time verification scans
CREATE TABLE IF NOT EXISTS qr_tokens (
    id            TEXT PRIMARY KEY,
    credential_id TEXT NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    token         TEXT UNIQUE NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    used_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_credentials_status ON credentials(status);
CREATE INDEX IF NOT EXISTS idx_qr_tokens_token ON qr_tokens(token);
CREATE INDEX IF NOT EXISTS idx_qr_tokens_credential_id ON qr_tokens(credential_id);
CREATE INDEX IF NOT EXISTS idx_professional_bodies_slug ON professional_bodies(slug);

-- Seed some New Zealand professional bodies
INSERT INTO professional_bodies (id, name, slug, base_url) VALUES
    ('pb_nzmc', 'Medical Council of New Zealand', 'nz-medical-council', 'https://example.com/api/mcnz'),
    ('pb_nzls', 'New Zealand Law Society', 'nz-law-society', 'https://example.com/api/nzls'),
    ('pb_nzpb', 'Psychologists Board of New Zealand', 'nz-psychologists-board', 'https://example.com/api/pbnz'),
    ('pb_nzne', 'Engineering New Zealand', 'engineering-nz', 'https://example.com/api/engineeringnz'),
    ('pb_nzna', 'Nursing Council of New Zealand', 'nz-nursing-council', 'https://example.com/api/ncnz'),
    ('pb_nzta', 'Teaching Council of New Zealand', 'nz-teaching-council', 'https://example.com/api/tcnz')
ON CONFLICT (id) DO NOTHING;

COMMIT;