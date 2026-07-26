# NOANT Backend (Go API Server)

REST API server powering the NOANT platform. Go 1.25, Gin web framework, TiDB (MySQL-compatible) + Redis.

## Tech Stack

- **Language:** Go 1.25
- **Web Framework:** Gin
- **Database:** TiDB Cloud (MySQL-compatible) via `database/sql`
- **Cache:** Redis (optional, in-memory fallback)
- **AI:** Groq API (Llama 3.3)
- **WebSocket:** gorilla/websocket
- **Monitoring:** Sentry, Prometheus metrics
- **Validation:** Playbook validator
- **Auth:** JWT (httpOnly cookies)

## Project Structure

```
backend/
├── main.go                    # Entry point, route wiring, DI
├── config/                    # Environment config loader
├── internal/
│   ├── domain/                # Domain types and interfaces
│   ├── errors/                # Custom error types
│   ├── handler/               # HTTP handlers (one per domain)
│   │   ├── handler.go         # Handlers struct, NewHandlers()
│   │   ├── auth_handler.go    # Login, register, refresh, verify
│   │   ├── chat_handler.go    # Send message, get messages
│   │   ├── admin_handler.go   # CEO Command Center endpoints
│   │   ├── websocket.go       # WS hub (chat + admin events)
│   │   └── ...                # ~25 handler files
│   ├── service/               # Business logic layer
│   │   ├── services.go        # Services aggregate
│   │   ├── auth.go            # Auth service (JWT, bcrypt, verification)
│   │   ├── chat.go            # AI orchestration, message handling
│   │   ├── openwa.go          # WhatsApp OpenWA integration
│   │   └── ...                # ~49 service files
│   ├── repository/            # Data access layer
│   │   ├── repositories.go    # Repos aggregate + RunInTx helper
│   │   ├── user_repo.go       # User queries
│   │   └── ...                # ~27 repo files
│   ├── infrastructure/        # Cross-cutting concerns
│   │   ├── db.go              # TiDB connection + migrations
│   │   ├── redis.go           # Redis client
│   │   ├── cache.go           # Cache layer
│   │   ├── job_queue.go       # Background job system
│   │   ├── metrics.go         # Prometheus metrics
│   │   ├── logger.go          # Structured logging (slog)
│   │   └── ws.go              # WebSocket helpers
│   ├── middleware/            # HTTP middleware
│   │   ├── auth.go            # JWT auth, cookie extraction
│   │   ├── admin.go           # RequireAdmin role check
│   │   ├── audit.go           # Audit logging
│   │   ├── ratelimit.go       # Rate limiting
│   │   └── cors.go            # CORS configuration
│   └── utils/                 # Response helpers
├── migrations/                # SQL migrations (TiDB-specific)
├── tests/                     # Integration tests (testcontainers)
├── docs/
│   └── openapi.yaml           # OpenAPI 3.0.3 spec (4995 lines)
└── go.mod                     # Go module definition
```

## Layer Architecture

```
HTTP Request → Middleware → Handler → Service → Repository → TiDB
                                    ↕
                              Infrastructure (Redis, Cache, Jobs, WS)
```

**Handlers** validate input, call services, return JSON.
**Services** contain business logic, orchestrate repos.
**Repositories** execute SQL queries, manage transactions.

## Setup

```bash
cd backend

# Requires Go 1.25+
export PATH=~/go125/go/bin:$PATH

# Install dependencies
go mod download

# Run
go run .

# Build
go build -o noant-server .
```

## Configuration

All config via environment variables (see root `.env.example`):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DB_HOST` | Yes | — | TiDB host |
| `DB_PORT` | Yes | 4000 | TiDB port |
| `DB_USER` | Yes | — | Database user |
| `DB_PASSWORD` | Yes | — | Database password |
| `DB_NAME` | Yes | noant | Database name |
| `REDIS_URL` | No | — | Redis URL (optional) |
| `JWT_SECRET` | Yes | — | JWT signing secret |
| `GROQ_API_KEY` | Yes | — | Groq API key |
| `SENTRY_DSN` | No | — | Sentry error tracking |
| `NODE_ENV` | No | development | Environment |

## Testing

```bash
# Unit + handler tests (fast, no Docker)
go test -short -count=1 ./...

# Integration tests (requires Docker for testcontainers)
go test -count=1 ./tests/...

# Specific package
go test -short -count=1 ./internal/handler/...

# With verbose output
go test -short -v -count=1 ./internal/service/...
```

## Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run
golangci-lint run

# From project root
~/go/bin/golangci-lint run
```

## Key Design Patterns

- **RunInTx** — `repos.RunInTx(ctx, func(tx) error)` wraps operations in a database transaction
- **Dual-scope queries** — conversation queries use `WHERE (org_id = ? AND org_id IS NOT NULL) OR (org_id IS NULL AND user_id = ?)` for multi-tenancy
- **Null-safe JSON** — slices initialized with `make()` (never nil) to avoid JSON `null` responses
- **BroadcastAdminEvent** — WebSocket hub pushes events to connected admin clients for realtime updates
- **describeAction** — audit middleware maps routes to human-readable action strings

## Adding a New Endpoint

1. Define request/response types in `handler/your_handler.go`
2. Add handler method
3. Add service method (if business logic needed)
4. Register route in `main.go` under appropriate group
5. Add OpenAPI spec in `docs/openapi.yaml`
6. Write tests in `handler/your_handler_test.go`
