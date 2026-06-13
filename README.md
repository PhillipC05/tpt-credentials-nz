# Professional Credentials Wallet — App 6

A RealMe-verified digital credentials wallet for New Zealand professionals.
Link your professional licences and qualifications to your verified identity and share them via QR code.

## Features

- **RealMe Verified Identity**: All credential holders must authenticate with RealMe Verified (Assertion Service)
- **Credential Issuance**: Verified professionals can self-attest credentials; future integration with licensing bodies
- **QR Verification**: Generate a time-limited QR code for third parties to verify credentials without an account
- **Credential Wallet**: Browse and manage all credentials linked to your FLT
- **Public Verification**: Anyone can verify a credential by scanning the QR code or entering the credential ID

## Architecture

```
packages/app-credentials/
├── cmd/server/main.go          # HTTP server entrypoint
├── internal/
│   ├── models/credential.go    # Domain types (Credential, CredentialType)
│   ├── repository/             # PostgreSQL data access
│   ├── services/credential.go  # Business logic (issue, revoke, verify)
│   └── handlers/
│       ├── auth.go             # RealMe authentication handlers
│       ├── credential.go       # Credential CRUD
│       └── public.go           # Unauthenticated verification endpoint
├── migrations/001_init.sql     # Database schema (credentials schema)
├── web/                        # Next.js 14 frontend
├── Dockerfile                  # Go backend container
└── docker-compose.yml          # Service composition (port 8094)
```

## Tech Stack

- **Backend**: Go 1.23, Chi router, pgx (PostgreSQL)
- **Auth**: RealMe Verified Identity (SAML 2.0, Assertion Service)
- **Frontend**: Next.js 14, TypeScript, Tailwind CSS
- **QR Code**: `skip2/go-qrcode` (server-side generation)
- **Database**: PostgreSQL 16

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 20+ (with pnpm)
- Docker and Docker Compose
- RealMe certificates (dev mock available)

### 1. Start Infrastructure

From the project root:

```bash
make dev
```

### 2. Database Migration

```bash
DATABASE_URL="postgres://tptnz:tptnz_dev@localhost:5432/tptnz?sslmode=disable" \
  atlas schema apply --dir "file://packages/app-credentials/migrations" \
  --url "$DATABASE_URL" --auto-approve
```

Or apply manually:

```bash
psql "postgres://tptnz:tptnz_dev@localhost:5432/tptnz?sslmode=disable" \
  -f packages/app-credentials/migrations/001_init.sql
```

### 3. Run the Backend

```bash
cd packages/app-credentials
DATABASE_URL="postgres://tptnz:tptnz_dev@localhost:5432/tptnz?sslmode=disable" \
  go run ./cmd/server
```

The server starts on `http://localhost:8094`.

### 4. Run the Frontend

```bash
cd packages/app-credentials/web
pnpm install
pnpm dev
```

### 5. Docker Compose (all-in-one)

```bash
docker compose -f docker-compose.yml \
  -f packages/app-credentials/docker-compose.yml up
```

## API Endpoints

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/login` | Initiate RealMe login |
| GET | `/auth/callback` | RealMe SAML callback |
| GET | `/auth/logout` | Clear session |
| GET | `/auth/metadata` | SAML SP metadata XML |
| GET | `/auth/status` | Current auth status |

### Credentials (requires RealMe Verified)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/credentials` | List own credentials |
| POST | `/credentials` | Issue a new credential |
| GET | `/credentials/{id}` | Get credential detail |
| DELETE | `/credentials/{id}` | Revoke a credential |
| POST | `/credentials/{id}/qr` | Generate QR verification token |

### Public Verification (no auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/verify/{token}` | Verify credential by QR token |
| GET | `/verify/id/{credentialId}` | Verify credential by ID |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDR` | `:8094` | Server listen address |
| `DATABASE_URL` | `postgres://tptnz:tptnz_dev@localhost:5432/tptnz?sslmode=disable` | PostgreSQL connection string |
| `REALME_ENVIRONMENT` | `mts` | mts, ite, or production |
| `REALME_CERT_FILE` | `certs/sp.crt` | SP certificate path |
| `REALME_KEY_FILE` | `certs/sp.key` | SP private key path |
| `REALME_ENTITY_ID` | `http://localhost:8094/auth/metadata` | SAML entity ID |
| `REALME_ACS_URL` | `http://localhost:8094/auth/callback` | SAML ACS URL |
| `REALME_IDP_METADATA_URL` | `http://localhost:8081/metadata` | IdP metadata URL |

## RealMe Registration

To use this app with real RealMe identities (ITE or Production environments),
you must register a Service Provider with the Department of Internal Affairs.

### MTS (Messaging Test Site) — Development

1. Generate a self-signed certificate and key:
   ```bash
   mkdir -p certs
   openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
     -keyout certs/sp.key -out certs/sp.crt \
     -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost"
   ```

2. In MTS, no formal registration is needed — use the mock IdP in
   `packages/realme-go/testenv/` for local development.

3. Start the mock IdP:
   ```bash
   cd packages/realme-go
   go run ./testenv/ -addr :8081
   ```

4. Configure the app to use the mock IdP:
   ```bash
   REALME_IDP_METADATA_URL=http://localhost:8081/metadata
   ```

### ITE (Integration Test Environment) — Pre-Production

1. Log in to the [RealMe Developer Portal](https://developers.realme.govt.nz/)
   and register a new service.
2. Submit your SP metadata XML (available at `GET /auth/metadata`) to DIA.
3. DIA will provide the ITE IdP metadata URL.
4. Generate a proper certificate (not self-signed) using the naming convention:
   `ite.{service-name}.{org-domain}.nz`
5. Configure environment variables:
   ```bash
   REALME_ENVIRONMENT=ite
   REALME_CERT_FILE=certs/ite.sp.crt
   REALME_KEY_FILE=certs/ite.sp.key
   REALME_IDP_METADATA_URL=<DIA-provided-ITE-url>
   ```

### Production

Follow the ITE steps above, substituting:
```bash
REALME_ENVIRONMENT=production
REALME_CERT_FILE=certs/prod.sp.crt
REALME_KEY_FILE=certs/prod.sp.key
REALME_IDP_METADATA_URL=<DIA-provided-prod-url>
```

## Testing

```bash
cd packages/app-credentials
go test ./...

# With race detection
go test -race ./...
```

## License

See the [LICENSE](../../LICENSE) file for details.
