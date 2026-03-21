# Chirpy

A RESTful API server for a social messaging platform, built with Go. Users can post short messages ("chirps"), authenticate with JWT tokens, and upgrade to premium via webhook integration.

Built as a [Boot.dev](https://boot.dev) guided project.

## Table of Contents

- [Motivation](#motivation)
- [Installation](#installation)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [Running the Server](#running-the-server)
- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
  - [Health Check](#health-check)
  - [Users](#users)
  - [Authentication Endpoints](#authentication-endpoints)
  - [Chirps](#chirps)
  - [Polka Webhooks](#polka-webhooks)
  - [Admin](#admin)
- [Project Structure](#project-structure)

## Motivation

Chirpy is a Twitter-like API that demonstrates core backend concepts in Go:

- RESTful API design using Go's standard `net/http` library
- JWT-based authentication with access and refresh tokens
- PostgreSQL database with migrations (Goose) and type-safe queries (SQLC)
- Webhook integration for external payment processing
- Middleware patterns for metrics tracking

## Installation

**Prerequisites:**

- [Go](https://golang.org/dl/) 1.25.0+
- [PostgreSQL](https://www.postgresql.org/download/)
- [Goose](https://github.com/pressly/goose) (for database migrations)
- [SQLC](https://sqlc.dev/) (only if modifying SQL queries)

```bash
git clone https://github.com/pgrigorakis/chirpy.git
cd chirpy
go mod download
```

## Configuration

Create a `.env` file in the project root with the following variables:

| Variable     | Description                          | Example                                                      |
|--------------|--------------------------------------|--------------------------------------------------------------|
| `DB_URL`     | PostgreSQL connection string         | `postgres://user:pass@localhost:5432/chirpy?sslmode=disable`  |
| `JWT_SECRET` | Secret key for signing JWT tokens    | `my-super-secret-key`                                        |
| `PLATFORM`   | Set to `dev` to enable admin reset   | `dev`                                                        |
| `POLKA_KEY`  | API key for Polka webhook validation | `f271c81ff7084ee5b99a5091b42d486e`                           |

```bash
# .env
DB_URL="postgres://user:pass@localhost:5432/chirpy?sslmode=disable"
JWT_SECRET="your-secret-key-here"
PLATFORM="dev"
POLKA_KEY="your-polka-api-key"
```

## Database Setup

1. Create the PostgreSQL database:

```bash
createdb chirpy
```

2. Run migrations with Goose:

```bash
goose -dir sql/schema postgres "your-db-url-here" up
```

This creates three tables:
- **users** — id, email, hashed_password, is_chirpy_red (premium flag)
- **chirps** — id, body (max 140 chars), user_id (foreign key)
- **refresh_tokens** — token, user_id, expires_at, revoked_at

## Running the Server

```bash
go build -o chirpy && ./chirpy
```

Or directly:

```bash
go run .
```

The server starts on **port 8080**:

```
Serving files from . on port: 8080
```

## Authentication

Chirpy uses a **two-token authentication system**:

### Access Token (JWT)
- Issued on login, expires in **1 hour**
- Signed with HS256 using the `JWT_SECRET`
- Used in the `Authorization` header for protected endpoints:
  ```
  Authorization: Bearer <access_token>
  ```

### Refresh Token
- Issued on login, expires in **60 days**
- A random 256-bit hex string stored in the database
- Used to obtain a new access token via `POST /api/refresh`
- Can be revoked via `POST /api/revoke`

### Password Hashing
- Passwords are hashed using **Argon2id** before storage

## API Endpoints

### Health Check

#### `GET /api/healthz`

Returns server health status.

```bash
curl http://localhost:8080/api/healthz
```

**Response:** `200 OK`
```
OK
```

---

### Users

#### `POST /api/users` — Register a new user

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepassword"}'
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

#### `PUT /api/users` — Update email and password

Requires authentication.

```bash
curl -X PUT http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"email": "newemail@example.com", "password": "newpassword"}'
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "email": "newemail@example.com",
  "is_chirpy_red": false
}
```

---

### Authentication Endpoints

#### `POST /api/login` — Log in and receive tokens

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "securepassword"}'
```

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "<jwt_access_token>",
  "refresh_token": "<hex_refresh_token>"
}
```

#### `POST /api/refresh` — Get a new access token

```bash
curl -X POST http://localhost:8080/api/refresh \
  -H "Authorization: Bearer <refresh_token>"
```

**Response:** `200 OK`
```json
{
  "token": "<new_jwt_access_token>"
}
```

#### `POST /api/revoke` — Revoke a refresh token

```bash
curl -X POST http://localhost:8080/api/revoke \
  -H "Authorization: Bearer <refresh_token>"
```

**Response:** `204 No Content`

---

### Chirps

#### `POST /api/chirps` — Create a chirp

Requires authentication. Body must be **140 characters or fewer**. Profane words (`kerfuffle`, `sharbert`, `fornax`) are automatically censored with `****`.

```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"body": "Hello, world! This is my first chirp!"}'
```

**Response:** `201 Created`
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "body": "Hello, world! This is my first chirp!",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### `GET /api/chirps` — List all chirps

Supports optional query parameters for filtering and sorting.

| Parameter   | Type   | Description                              |
|-------------|--------|------------------------------------------|
| `author_id` | UUID   | Filter chirps by a specific author       |
| `sort`      | string | `asc` (default) or `desc` by created_at |

```bash
# All chirps (ascending by default)
curl http://localhost:8080/api/chirps

# Filter by author, sorted newest first
curl "http://localhost:8080/api/chirps?author_id=550e8400-...&sort=desc"
```

**Response:** `200 OK`
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z",
    "body": "Hello, world!",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }
]
```

#### `GET /api/chirps/{chirpID}` — Get a single chirp

```bash
curl http://localhost:8080/api/chirps/660e8400-e29b-41d4-a716-446655440000
```

**Response:** `200 OK`
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "body": "Hello, world!",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error:** `404 Not Found` if chirp doesn't exist.

#### `DELETE /api/chirps/{chirpID}` — Delete a chirp

Requires authentication. Only the chirp's author can delete it.

```bash
curl -X DELETE http://localhost:8080/api/chirps/660e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <access_token>"
```

**Response:** `204 No Content`

**Error:** `403 Forbidden` if you are not the chirp's author.

---

### Polka Webhooks

#### `POST /api/polka/webhooks` — Handle Polka payment events

Authenticated with the Polka API key. Upgrades a user to Chirpy Red (premium) when a `user.upgraded` event is received.

```bash
curl -X POST http://localhost:8080/api/polka/webhooks \
  -H "Content-Type: application/json" \
  -H "Authorization: ApiKey <polka_key>" \
  -d '{"event": "user.upgraded", "data": {"user_id": "550e8400-..."}}'
```

**Response:** `204 No Content`

Non-`user.upgraded` events return `204` with no action taken.

---

### Admin

#### `GET /admin/metrics` — View file server hit count

Returns an HTML page showing how many times the `/app/` file server has been accessed.

```bash
curl http://localhost:8080/admin/metrics
```

#### `POST /admin/reset` — Reset the database (dev only)

Only available when `PLATFORM=dev`. Deletes all users (and cascades to chirps and refresh tokens) and resets the hit counter.

```bash
curl -X POST http://localhost:8080/admin/reset
```

**Response:** `200 OK` (dev) or `403 Forbidden` (non-dev).

---

### Static Files

#### `GET /app/*` — Serve static files

Serves files from the project root. Each request increments the metrics hit counter.

## Project Structure

```
chirpy/
├── main.go                        # Entry point, route registration
├── helpers.go                     # JSON response helpers
├── metrics.go                     # File server hit counter middleware
├── reset.go                       # Admin reset handler
├── status.go                      # Health check handler
├── handler_chirps_create.go       # POST /api/chirps
├── handler_chirps_delete.go       # DELETE /api/chirps/{chirpID}
├── handler_chirps_get.go          # GET /api/chirps, GET /api/chirps/{chirpID}
├── handler_users_create.go        # POST /api/users
├── handler_users_login.go         # POST /api/login
├── handler_users_update.go        # PUT /api/users
├── handler_users_upgrade.go       # POST /api/polka/webhooks
├── handler_refresh.go             # POST /api/refresh, POST /api/revoke
├── internal/
│   ├── auth/                      # JWT, password hashing, token extraction
│   └── database/                  # SQLC-generated DB queries and models
└── sql/
    ├── schema/                    # Goose migration files (001-005)
    └── queries/                   # SQL queries used by SQLC
```
