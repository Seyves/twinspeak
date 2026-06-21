# TwinSpeak Agent Guide

## Architecture

**Monorepo**: `backend/` (Go) + `frontend/` (TanStack Start/React)

- **Backend**: Go 1.25, Fiber, PostgreSQL, sqlc for type-safe queries
  - Entry: `backend/cmd/twinspeek/main.go`
  - Config: `backend/config.yaml` (gitignored, see `config-example.yaml`)
  - DB: schema in `schema.sql`, queries in `query.sql`
  - Hot reload: Air (`.air.toml`)
- **Frontend**: TanStack Start, React 19, Tailwind CSS v4, Vite
  - Entry: `frontend/src/router.tsx`
  - Routes: file-based in `frontend/src/routes/`
  - API proxy: `/api/v1/**` → `backend:8080` (vite.config.ts:27-32)
  - WebSocket proxy: `/api/v1/ws` → `ws://backend:8080` (vite.config.ts:44-48)

## Developer Commands

All development happens through Docker Compose:

```bash
# Start everything (backend, frontend, db)
docker compose up

# Backend runs: air -c .air.toml (hot reload)
# Frontend runs: pnpm dev --host (port 4321)
# DB: PostgreSQL 18 on port 5432
```

**Testing backend**:

```bash
./test.sh
# Resets DB volume, runs `go test ./...` in docker-compose.test.yml
```

Frontend has Vitest configured (`pnpm test`) but no tests exist yet.

## Database Workflow

- **Schema**: edit `backend/schema.sql`
- **Queries**: edit `backend/query.sql` with sqlc annotations
- **Codegen**: `sqlc generate` → outputs to `backend/internal/db/`
- Tests expect `DB_URL` env var (set in docker-compose.test.yml:31)

## Backend Testing

- Tests use real PostgreSQL (docker-compose.test.yml)
- Each test wraps in a transaction (`prepare(t)` helper in auth_test.go:21-30)
- DB is reset before each test run (test.sh:5)
- Example: `backend/internal/auth/auth_test.go`, `backend/internal/billing/billing_test.go`

## Config Quirks

- Backend config is at `backend/config.yaml` (gitignored, contains secrets)
- Template: `backend/config-example.yaml`
- Air build: `go build -o ./tmp/main ./cmd/twinspeek/main.go` (air.toml:8)
- Air passes `-c ./config.yaml` to binary (air.toml:6)
- Two pipeline options: `gladia` or `whisper` (main.go:47-60)

## Frontend Style

- Prettier config: **no semicolons**, single quotes, 4-space indent (.prettierrc.json)
- Tailwind CSS v4 via Vite plugin
- Path alias: `@/*` → `./src/*` (tsconfig.json:10-13)
- Also supports: `#/*` → `./src/*` (package.json:5-7)

## Frontend State Management

**Jotai with TanStack Query Integration:**

- Read-only queries: use `atomWithQuery` for GET requests (e.g., `messagesAtom`, `preferencesAtom`)
- Write operations: **always use `atomWithMutation`** for POST/PUT/DELETE requests that need lifecycle management
  - Provides loading states (`isPending`)
  - Handles success/error callbacks
  - Supports query invalidation via `queryClientAtom`
  - Integrates with toast notifications using `toast.promise`

**Key points:**

- Mutations handle query invalidation internally via `onSuccess`
- Use `toast.promise` for consistent user feedback
- Never manually manage loading/error states when `atomWithMutation` is available

## API Integration

Frontend assumes backend is at `backend:8080` in Docker network. Proxied through:

- Nitro route rules for REST (vite.config.ts:26-33)
- Vite dev server proxy for WebSocket (vite.config.ts:43-49)

## Directory Ownership

- `backend/internal/`: auth, billing, clients, config, db, googleauth, metrics, preferences, server, speechpipeline, users
- `backend/cmd/`: twinspeek (main server), cli
- `frontend/src/`: api, assets, atoms, components, definitions, hooks, lib, routes

## Backend Code Style

**Error Handling:**

- Use named error variables for specific cases: `ErrInvalidCredentials`, `ErrMaliciousSuspicion`
- Every error should be wrapped in the simular manner: `log.Errorf("Error during operation: %s", err.Error())`
- Return fiber errors with status codes in trasport layer: `fiber.NewError(fiber.StatusBadRequest, "user message")`

**Transactions:**

- Always use `defer tx.Rollback(ctx)` after `Begin`
- Create transaction-scoped queries: `qtx := queries.WithTx(tx)`
- Commit explicitly at the end: `tx.Commit(ctx)`

**Module Pattern:**

- Each feature gets its own package under `internal/`
- Modules have a struct with dependencies and a `New()` constructor
- Modules should have as little dependencies as possible.
- Modules should be testable by themselves

**Constants:**

- Define at package level or const block
- Use descriptive names: `accessTokenLifetime`, `emailUnverifiedCookie`

**Cookies:**

- Use secure cookies: `HTTPOnly=true`, `Secure=true`, `SameSite=Strict`
- Exception: `email_unverified` cookie needs `HTTPOnly=false` for frontend access
- Helper pattern: `getSecureCookie()`, `getEmailUnverifiedCookie()`
- Use negative flags for transient state (e.g., `email_unverified=true` for new users, removed after verification)

## Common Gotchas

- **DB name inconsistency**: schema uses `twinspeak`, config.yaml references `twinspeek` (note spelling)
- **SQLC overrides**: UUIDs → `github.com/google/uuid`, timestamptz → `time.Time` (sqlc.yaml:16-38)
- **Test isolation**: `./test.sh` always nukes DB volume (test.sh:5)
- **Go version**: uses Go 1.25 (go.mod:3, Dockerfile)
- **Frontend port**: always 4321 (package.json:9,12; vite.config.ts:42)
- **Backend port**: always 8080 (config.yaml:1; docker-compose.yml:52)
- **Backend doesn't know frontend routes**: Never return redirects to frontend paths, return errors instead
- **SSR limitations**: Avoid heavy data fetching in SSR server functions, use them only for lightweight checks (cookies, etc.)
- **Cookie-based state**: Prefer cookie checks over DB queries for performance (with DB fallback for security)
