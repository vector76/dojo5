# Dojo CRM

A lightweight CRM for managing a martial arts dojo or yoga studio. Compiles to a single binary with embedded frontend assets and SQLite database.

## Prerequisites

- Go 1.25+
- Node.js 18+ and npm

## Quick Start

```bash
# Build the single binary
make build

# Set a JWT secret and run
JWT_SECRET=your-secret-here ./dojo-crm
```

The server starts on `http://localhost:8080`. On first run, you'll be prompted to create an admin account.

## Configuration

| Environment Variable | Description | Default |
|---|---|---|
| `JWT_SECRET` | **Required.** Secret key for signing JWT tokens | (none) |
| `PORT` | HTTP server port | `8080` |
| `DB_PATH` | Path to SQLite database file | `dojo.db` |

The database is created automatically on first run.

## Build

```bash
make build      # Build frontend + backend into ./dojo-crm
make clean      # Remove build artifacts
```

The build process:
1. Installs frontend dependencies (`npm ci`)
2. Builds the React app (`vite build`)
3. Copies the dist into the Go embed directory
4. Compiles the Go binary with embedded assets
5. Copies migration files alongside the binary

## Development

```bash
make dev        # Run Vite dev server + Go server concurrently
```

This starts the Vite dev server (with HMR) and the Go backend side by side. The Go server uses `JWT_SECRET=dev-secret` in dev mode.

## Testing

```bash
make test       # Run all tests (frontend + backend)
```

Individual suites:

```bash
make frontend-test   # Vitest + React Testing Library (131 tests)
make backend-test    # Go test (6 packages)
```

## Deployment

The `dojo-crm` binary is self-contained. To deploy:

1. `make build`
2. Copy `dojo-crm` and the `migrations/` directory to the target machine
3. Set `JWT_SECRET` and run

The binary serves both the API (`/api/`) and the React frontend (with SPA fallback for client-side routing).

## User Roles

| Role | Capabilities |
|---|---|
| **admin** | Full access: manage users, classes, class types, payments, attendance, balances |
| **instructor** | View members, manage attendance and check-ins |
| **user** | View own profile, classes, payments, and attendance history |

## API Overview

All API endpoints are under `/api/`. Authenticated endpoints require `Authorization: Bearer <token>`.

- `POST /api/auth/setup` - Create initial admin (first run only)
- `POST /api/auth/login` - Login, returns JWT token
- `GET /api/auth/me` - Current user profile
- `GET/POST /api/users` - List/create users
- `GET/PUT/DELETE /api/users/:id` - User CRUD
- `PUT /api/users/:id/password` - Reset password (admin)
- `GET/PUT /api/users/:id/balance` - View/set expected balance
- `GET /api/users/:id/payments` - User payment history
- `GET /api/users/:id/attendance` - User attendance history
- `GET/POST/PUT/DELETE /api/class-types` - Class type management
- `GET/POST/PUT/DELETE /api/classes` - Class scheduling
- `GET/POST /api/classes/:id/attendance` - Class attendance
- `POST /api/payments` - Record a payment
- `GET /api/health` - Health check

Error responses use JSON format: `{"error": "message"}`.

## Project Structure

```
backend/
  cmd/server/        # Main entry point
  internal/
    auth/            # JWT middleware and role checks
    database/        # SQLite connection and migrations
    frontend/        # Embedded frontend assets + SPA handler
    handlers/        # HTTP handlers for all API routes
    models/          # Data models and repository layer
  migrations/        # SQL migration files
frontend/
  src/
    components/      # Shared components (Layout, ErrorBoundary, Toast, etc.)
    pages/           # Page components (Dashboard, Members, Classes, etc.)
    api.ts           # API client with JWT auth
    auth.tsx         # Auth context and provider
docs/                # Architecture and design documents
```
