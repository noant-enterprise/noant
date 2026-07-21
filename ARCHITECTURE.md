# NOANT Architecture

## System Overview

NOANT is an AI-powered customer support platform with multi-channel integration. It enables businesses to handle customer conversations across WhatsApp (via OpenWA), Telegram, and a web widget — all driven by an AI response engine with human-in-the-loop capabilities.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25 + Gin |
| Frontend | React 18 + TypeScript + Vite |
| Database | TiDB (MySQL-compatible) |
| Cache | Redis 7 (optional, falls back to in-memory) |
| AI | Groq API (Llama 3.3) |
| Payments | Polar.sh |
| WhatsApp | OpenWA (self-hosted) |
| Monitoring | Prometheus + Grafana |

## Backend Architecture

The backend follows a layered architecture with clear separation of concerns:

```
HTTP Request
    │
    ▼
┌─────────────┐
│  Middleware  │  Auth, CSRF, rate limiting, sanitization
└─────┬───────┘
      │
      ▼
┌─────────────┐
│   Handler   │  HTTP request/response, validation, error mapping
└─────┬───────┘
      │
      ▼
┌─────────────┐
│   Service   │  Business logic, orchestration
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Repository  │  Data access (MySQL queries + Redis caching)
└─────┬───────┘
      │
      ▼
┌──────────────┐
│Infrastructure│  DB connections, Redis, logging, metrics, migrations
└──────────────┘
```

### Handlers (`internal/handler/`)

HTTP layer responsible for:
- Parsing and validating request bodies
- Calling service methods
- Mapping errors to HTTP status codes
- Returning structured JSON responses

### Services (`internal/service/`)

Business logic layer responsible for:
- Orchestrating cross-cutting concerns (e.g., AI + chat + credits)
- Enforcing business rules and invariants
- Coordinating between repositories
- Managing external API integrations (Groq, Polar.sh, OpenWA, Telegram)

### Repositories (`internal/repository/`)

Data access layer responsible for:
- MySQL queries via `database/sql`
- Redis caching with fallback to in-memory
- Domain-specific query methods
- Interface-based design for testability (mocks in `internal/repository/mock.go`)

### Infrastructure (`internal/infrastructure/`)

Cross-cutting technical concerns:
- `db.go` — Database connection pool and health checks
- `redis.go` — Redis client initialization
- `cache.go` — Multi-tier caching (Redis → in-memory)
- `logger.go` — Structured JSON logging via `slog`
- `metrics.go` — Prometheus metrics registration and collection
- `prometheus.go` — Prometheus HTTP endpoint
- `migrations.go` — Auto-run SQL migrations on startup
- `jobqueue.go` — Background job queue with Redis-backed workers
- `bottleneck.go` — Rate limiter implementation
- `memory_ratelimit.go` — In-memory rate limiter (no Redis required)
- `blacklist.go` — IP/token blacklist

### Middleware (`internal/middleware/`)

- `auth.go` — JWT access + refresh token validation
- `csrf.go` — Origin header validation, double-submit cookie pattern
- `bodylimit.go` — Request body size limits
- `sanitize_middleware.go` — Input sanitization
- `audit.go` — Request audit logging
- `websocket_auth.go` — WebSocket upgrade authentication

## Key Files

### Backend

| File | Purpose |
|------|---------|
| `backend/main.go` | Application entrypoint, DI wiring, route registration |
| `backend/internal/service/aibrain.go` | AI response orchestration (main entry) |
| `backend/internal/service/aibrain_groq.go` | Groq API integration |
| `backend/internal/service/aibrain_intent.go` | Intent classification |
| `backend/internal/service/aibrain_prompt.go` | Prompt construction |
| `backend/internal/service/aibrain_response.go` | Response humanization |
| `backend/internal/service/aibrain_search.go` | QA knowledge base search |
| `backend/internal/service/openwa.go` | WhatsApp session management |
| `backend/internal/service/openwa_webhook.go` | WhatsApp webhook processing |
| `backend/internal/handler/` | 14+ domain handler files |
| `backend/internal/service/` | 13+ domain service files |
| `backend/internal/repository/` | 19+ domain repo files + interfaces + mocks |

### Frontend

| Path | Purpose |
|------|---------|
| `frontend/src/App.tsx` | Root component, routing |
| `frontend/src/app/` | Page-level views |
| `frontend/src/components/` | Reusable UI components |
| `frontend/src/hooks/` | Custom React hooks |
| `frontend/src/contexts/` | React context providers |
| `frontend/src/lib/` | API clients, utilities |
| `frontend/src/types/` | TypeScript type definitions |

## Data Flow

A typical message lifecycle:

```
1. User sends message via WhatsApp / Telegram / Web Widget
        │
2. Webhook receives message → routes to channel handler
   (openwa_handler / telegram / websocket)
        │
3. Chat service creates or finds existing conversation
        │
4. AI Brain processes the message:
   a. Intent classification (support, sales, general, etc.)
   b. QA knowledge base search (vector similarity)
   c. Groq API call (Llama 3.3) with context + history
   d. Response humanization (tone, length, personality)
        │
5. Response sent back through the same channel
        │
6. Metrics recorded + audit log written
```

## Security

### Authentication
- JWT-based with access + refresh token rotation
- Tokens stored in `httpOnly` cookies (not accessible via JavaScript)
- Short-lived access tokens (15 min), longer refresh tokens (7 days)

### CSRF Protection
- Origin header validation on all state-changing requests
- Double-submit cookie pattern for browser clients

### Rate Limiting
- IP-based rate limiting for unauthenticated endpoints
- User-based rate limiting for authenticated endpoints
- Configurable per-route limits
- Falls back to in-memory limiter when Redis is unavailable

### Other
- OWASP security headers (X-Content-Type-Options, X-Frame-Options, CSP, etc.)
- Input sanitization on all user-provided data
- Path traversal protection on file-serving routes
- Request body size limits to prevent abuse

## Observability

### Metrics (Prometheus)
- Request counts and latencies per route
- AI inference metrics (tokens, latency, cost)
- Credit consumption metrics
- OpenWA session health
- Custom business metrics

### Logging
- Structured JSON logging via Go `slog`
- Request ID propagation for distributed tracing
- Log levels: debug, info, warn, error

### Audit Trail
- All state-changing operations are audit-logged
- Tracks: actor, action, resource, timestamp, before/after values
- Queryable via the audit handler

## Background Jobs

The job queue runs on a Redis-backed worker system (falls back to in-memory processing when Redis is unavailable).

### Scheduled Tasks
- **Health checks** — periodic service health monitoring
- **Cache cleanup** — evict stale entries from in-memory cache
- **Credit expiry** — expire unused credits after configured TTL
- **Campaign processing** — send scheduled campaign messages in batches
- **DB cleanup** — archive or purge old records per retention policy
- **OpenWA session keepalive** — prevent WhatsApp session disconnection
- **Metrics flush** — aggregate and export metrics periodically

## Database

TiDB is the primary datastore (MySQL-compatible). Schema is managed via SQL migration files in `backend/migrations/`. Migrations run automatically on application startup via `infrastructure/migrations.go`.

Key domain tables:
- `users` — user accounts and authentication
- `conversations` — chat sessions (channel-agnostic)
- `messages` — individual messages within conversations
- `qa_entries` — knowledge base for AI responses
- `credits` — user credit balances
- `campaigns` / `campaign_recipients` — outbound messaging campaigns
- `integrations` — connected channel configurations
- `subscriptions` — payment/plan subscriptions
- `audit_logs` — operation audit trail
