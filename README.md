# Caregiver Shift Tracker API

Backend service for the Blue Horn Tech caregiver scheduling assignment. The API is written in Go (1.22), follows a clean architecture layering, and persists data in PostgreSQL. It exposes REST endpoints for schedules, tasks, and a lightweight OAuth2/OIDC-style token flow so the frontend can authenticate via client credentials.

## Tech Stack

- Go 1.22
- Gin as the HTTP router
- sqlx with pgx driver for PostgreSQL
- Zap structured logging
- JWT (HS256) tokens for the faux OIDC flow
- Plain SQL migrations (stored in `/migrations`)

## Project Layout

```
backend/
  cmd/api                 # application entrypoint
  internal/               # business layers
    app/                  # dependency wiring
    config/               # environment-driven configuration + DB bootstrap
    domain/               # entities and domain errors
    middleware/           # logging factory + auth/request middleware
    handler/              # HTTP & docs handlers
    docs/                 # embedded Swagger/OpenAPI assets
    repository/           # interfaces + Postgres implementations
    router/               # Gin router wiring
    usecase/              # business logic coordinators
  migrations/             # SQL migrations + seed data
  .env.example            # sample environment variables
```

## Prerequisites

- Go 1.22+
- Docker (recommended for local PostgreSQL)
- Alternatively: PostgreSQL 14+ installed locally

## Configuration

Copy the sample environment file and adjust as needed:

```bash
cp backend/.env.example backend/.env
```

Key variables:

- `APP_HOST`, `APP_PORT` – server bind address.
- `DB_*` – Postgres connection settings.
- `AUTH_*` – secrets + token metadata for the pseudo-OIDC flow. Update the secrets for production use.

## Quick start (recommended)

Use Docker Compose (provided in the repository root) to start PostgreSQL and run the API. From the repository root run:

```bash
# 1. Start PostgreSQL
docker compose up -d postgres

# 2. Copy env configuration (edit if needed)
cp backend/.env.example backend/.env

# 3. Apply the SQL migrations (seeds demo data)
docker compose exec -T postgres psql -U postgres -d care_tracker < backend/migrations/0001_init.sql
docker compose exec -T postgres psql -U postgres -d care_tracker < backend/migrations/0002_logs_caregivers.sql

# 4. Start the API
cd backend
go run ./cmd
```

When the server prints `🚀 care-shift-tracker` it is ready on `http://localhost:8080`. Press `ctrl+c` to stop it, then tear down the database with `docker compose down` (or keep it running for the frontend).

## Database Setup

If you prefer a locally installed Postgres, create the database and run the migrations manually:

1. Create the database:
   ```bash
   createdb care_tracker
   ```
2. Run the migration SQL (includes seed data and attendance logs table):
   ```bash
   psql "$DB_USER" -h "$DB_HOST" -d "$DB_NAME" -f migrations/0001_init.sql
   psql "$DB_USER" -h "$DB_HOST" -d "$DB_NAME" -f migrations/0002_logs_caregivers.sql
   ```
   (Ensure the environment variables in your shell line up with `.env`.)

The migration seeds a default caregiver, schedules, tasks, and an OAuth client (`caregiver-app`) with the plaintext secret `caregiver-secret` (hashed in the DB).

## Running the API

```bash
cd backend
go run ./cmd
```

or build and run:

```bash
cd backend
go build -o bin/api ./cmd
./bin/api
```

The server listens on `APP_HOST:APP_PORT` (defaults to `0.0.0.0:8080`). A healthcheck is available at `GET /healthz`.

## Auth Flow (OIDC “Wannabe”)

The `/api/auth/token` endpoint implements a simplified client-credentials exchange:

- Request body: form-encoded or JSON with `grant_type=client_credentials`, `client_id`, `client_secret`, and optional `scope`.
- Response: access token (HS256 JWT), ID token (HS256), token type, expires in seconds, granted scope, and caregiver profile payload.
- The default seeded client is `caregiver-app` / `caregiver-secret`.

Example request:

```bash
curl -X POST http://localhost:8080/api/auth/token \
  -H 'Content-Type: application/json' \
  -d '{
        "grant_type": "client_credentials",
        "client_id": "caregiver-app",
        "client_secret": "caregiver-secret"
      }'
```

Use the returned access token as a Bearer token on protected routes.

## API Documentation

- Swagger UI: `GET /docs`
- Raw OpenAPI spec: `GET /docs/openapi.yaml`

Both assets are embedded in the binary so no extra tooling is required.

## API Surface

| Method  | Path                       | Description                                               |
| ------- | -------------------------- | --------------------------------------------------------- |
| `POST`  | `/api/auth/token`          | Obtain access + ID token                                  |
| `GET`   | `/api/schedules`           | List schedules (filter by `status`, `date` query params)  |
| `GET`   | `/api/schedules/today`     | Today’s schedules + metrics                               |
| `GET`   | `/api/schedules/metrics`   | Aggregate counts for a given date (`?date=YYYY-MM-DD`)    |
| `GET`   | `/api/schedules/:id`       | Schedule detail with tasks and client info                |
| `POST`  | `/api/schedules/:id/start` | Clock-in; requires `latitude` & `longitude`               |
| `POST`  | `/api/schedules/:id/end`   | Clock-out; requires `latitude` & `longitude`              |
| `POST`  | `/api/schedules/:id/tasks` | Add a new care task to the schedule                       |
| `PATCH` | `/api/tasks/:taskId`       | Update task status (complete or not-complete with reason) |

All `/api/*` endpoints except `/api/auth/token` require the Bearer access token header.

## Logging

- Structured JSON logs are emitted to stdout via Zap.
- Every request also lands in the `request_logs` Postgres table (method, path, status, latency, IP, user-agent) for later auditing/analytics.

## Testing the Flow Quickly

```bash
# 1) Token
token=$(curl -s -X POST http://localhost:8080/api/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"caregiver-app","client_secret":"caregiver-secret"}' | jq -r '.access_token')

# 2) List schedules
curl -H "Authorization: Bearer $token" http://localhost:8080/api/schedules
```

## Notes & Next Steps

- The logging middleware currently records request metadata; hook in trace IDs or structured request bodies as needed.
- Schedule status transitions are guarded in the use case layer; extend with additional validation or auditing tables for production environments.
- Swap secrets and token signing to asymmetric keys before deploying to a shared environment.
- Integrate a migration runner (e.g., `golang-migrate` or `goose`) if automated migration flow is preferred.
- Geolocation fallback: if the browser denies access, the frontend should prompt for manual latitude/longitude entry and send those values to the backend. Document any alternative flows in the frontend README when implemented.
- Run unit tests with `go test ./...` to validate use cases and auth flows.
