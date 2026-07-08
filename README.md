# NOANT Enterprise — v2.0

Autonomous & Human-in-the-Loop AI Customer Support Platform. Built in Nigeria. For the World.

---

## Table of Contents

1. [Overview](#overview)
2. [Technology Stack](#technology-stack)
3. [Project Structure](#project-structure)
4. [Architecture & System Topology](#architecture--system-topology)
5. [AI Orchestration Flow](#ai-orchestration-flow)
6. [Security Features](#security-features)
7. [Core Subsystems](#core-subsystems)
8. [Configuration Reference](#configuration-reference)
9. [Installation & Deployment](#installation--deployment)
10. [Testing](#testing)
11. [API Directory](#api-directory)
12. [Operations](#operations)
13. [Database Schema](#database-schema)

---

## Overview

NOANT Enterprise is a state-of-the-art, high-throughput, multi-tenant customer relationship and messaging platform. It features a **hybrid automation model** — balancing high-confidence AI answering capabilities with seamless, real-time human escalation across WhatsApp, Telegram, Facebook Messenger, and embedded Web Widgets.

### Brand Identity

The logo is modeled after an advanced conversation-loop icon:

- **Outer Dashed Ring** — Multi-channel messaging endpoints spinning in synchrony.
- **Solid Black Core** — Central AI Brain container.
- **Three White Dots** — Message propagation, growing as intelligence accumulates.

---

## Technology Stack

| Layer | Technology | Purpose |
|---|---|---|
| **Backend** | Go 1.22, Gin | HTTP API server, WebSocket hub, business logic |
| **Frontend** | React 18, TypeScript, Vite | Single-page application dashboard |
| **AI** | Groq Llama 3.3 (via REST API) | Natural language understanding & response generation |
| **Database** | TiDB Cloud (MySQL-compatible) | Distributed SQL for conversations, users, training data |
| **Cache & Queue** | Redis (Upstash) | Session tokens, rate limits, AI history, job queue |
| **Payments** | Polar API | Subscription & credit pack processing |
| **Embeddings** | Custom Go vector search | Semantic QA pair matching |
| **Deployment** | Render, Docker | Unified static + API server |

---

## Project Structure

```
noant/
├── backend/
│   ├── main.go                    # Entry point, router, middleware wiring
│   ├── config/
│   │   └── config.go              # Environment-based configuration
│   ├── migrations/                # SQL migration files (sequenced)
│   ├── internal/
│   │   ├── domain/
│   │   │   └── models.go          # Core domain types (User, Conversation, etc.)
│   │   ├── infrastructure/
│   │   │   ├── db.go              # TiDB connection pool
│   │   │   ├── redis.go           # Redis client wrapper
│   │   │   ├── cache.go           # Generic cache layer
│   │   │   ├── bottleneck.go      # Concurrency limiter
│   │   │   ├── jobqueue.go        # Background job scheduler
│   │   │   ├── blacklist.go       # In-memory token blacklist (Redis fallback)
│   │   │   ├── memory_ratelimit.go# In-memory rate limiter (Redis fallback)
│   │   │   └── logger.go          # Structured logger
│   │   ├── middleware/
│   │   │   ├── auth.go            # JWT auth, CSP, rate limiting, token blacklist
│   │   │   ├── csrf.go            # Origin/Referer CSRF validation
│   │   │   ├── bodylimit.go       # 1 MB request body limit
│   │   │   ├── sanitize.go        # XSS sanitization middleware
│   │   │   ├── audit.go           # Request audit logging
│   │   │   └── websocket_auth.go  # WebSocket origin validation
│   │   ├── handler/
│   │   │   ├── handler.go         # Auth, Chat, Integration, Inventory handlers
│   │   │   ├── websocket.go       # WebSocket hub & connection management
│   │   │   ├── health.go          # /health endpoint
│   │   │   ├── notifications.go   # Notification polling
│   │   │   └── background.go      # Background task endpoints
│   │   ├── service/
│   │   │   ├── service.go         # Auth, Chat, AI Brain, Integration services
│   │   │   ├── ai_sales.go        # Sales mode AI logic
│   │   │   ├── embedding.go       # Vector search & QA matching
│   │   │   ├── email.go           # Email sending (SMTP, Resend)
│   │   │   ├── plan.go            # Plan gating & enforcement
│   │   │   ├── credit.go          # Credit balance management
│   │   │   ├── campaign.go        # Broadcast campaign scheduling
│   │   │   ├── openwa.go          # Open WhatsApp API integration
│   │   │   ├── telegram.go        # Telegram bot integration
│   │   │   ├── dbmanager.go       # Data retention cleanup jobs
│   │   │   └── notifications.go   # Notification service
│   │   ├── repository/
│   │   │   ├── repository.go      # All DB repository implementations
│   │   │   ├── uow.go             # Unit of Work (transaction wrapper)
│   │   │   ├── audit.go           # Audit log repository
│   │   │   ├── notifications.go   # Notification repository
│   │   │   └── fcm_repository.go  # FCM push notification repository
│   │   └── utils/
│   │       ├── errors.go          # Standardized error responses
│   │       └── sanitize.go        # Reflection-based struct sanitizer
│   └── .env.example               # Environment variable template
├── frontend/
│   ├── src/
│   │   ├── app/                   # Next.js-style App Router pages
│   │   ├── components/            # Reusable UI components
│   │   ├── hooks/                 # Custom React hooks
│   │   ├── lib/                   # API client, WebSocket manager
│   │   ├── contexts/              # React context providers
│   │   └── types/                 # TypeScript type definitions
│   └── public/                    # Static assets, favicons
├── brain/                         # Design assets (logos, branding)
└── README.md
```

---

## Architecture & System Topology

The platform is designed around a decoupled microservice-compatible structure enforcing strict separation of concerns, multi-tenant workspace isolation, and real-time state synchronization.

```mermaid
graph TD
    subgraph Clients ["Client Gateways"]
        WA[WhatsApp App / Twilio]
        TG[Telegram Bot API]
        FB[Facebook Messenger API]
        WB[Web Widget Custom JS]
    end

    subgraph Edge ["Network & Security Layer"]
        R_PROXY[Reverse Proxy / CORS Guard]
        JWT_AUTH[JWT Auth Validator]
        RATE_LIM[Redis Rate Limiter / In-Memory Fallback]
        CSRF[CSRF Origin Validation]
        BODY_LIMIT[1 MB Body Limit]
        SANITIZE[XSS Sanitizer]
    end

    subgraph Backend ["Go 1.22 App Service (Gin)"]
        H_CHATS[ChatHandler]
        H_INTEG[IntegrationHandler]
        H_TRAIN[TrainingHandler]
        H_ANALY[AnalyticsHandler]

        S_INTEG[IntegrationService]
        S_WIDGET[WidgetService]
        S_CHAT[ChatService]

        AI_BRAIN[Groq Llama 3.3 AI Brain]
        AI_VALIDATE[Response Validator / Anti-Hallucination]
        WS_HUB[WebSocket Hub / Broadcast Server]
        JOB_Q[Background Job Queue]
    end

    subgraph Storage ["Distributed Persistence"]
        DB_TIDB[TiDB Cloud SQL Distributed Cluster]
        CACHE_RED[Upstash Cache & Memory Store]
        MEM_BLACKLIST[In-Memory Token Blacklist]
        MEM_RATELIMIT[In-Memory Rate Limiter]
    end

    subgraph UI ["React SPA Frontend (Vite)"]
        RT_CONN[Network Context & Banners]
        WC_CTX[Widget Config Context]
        RT_CHATS[Chats Inbox View]
        ANALY_DASH[Overview Analytics]
    end

    WA --> R_PROXY
    TG --> R_PROXY
    FB --> R_PROXY
    WB --> R_PROXY

    R_PROXY --> SANITIZE
    SANITIZE --> BODY_LIMIT
    BODY_LIMIT --> CSRF
    CSRF --> JWT_AUTH
    JWT_AUTH --> RATE_LIM

    RATE_LIM --> H_CHATS
    RATE_LIM --> H_INTEG
    RATE_LIM --> H_TRAIN
    RATE_LIM --> H_ANALY

    H_CHATS --> S_CHAT
    H_INTEG --> S_INTEG
    H_TRAIN --> S_CHAT
    H_ANALY --> S_CHAT

    S_CHAT --> AI_BRAIN
    AI_BRAIN --> AI_VALIDATE
    S_CHAT --> WS_HUB
    S_WIDGET --> DB_TIDB

    S_CHAT --> DB_TIDB
    S_CHAT --> CACHE_RED
    S_CHAT --> MEM_BLACKLIST
    S_CHAT --> MEM_RATELIMIT

    JOB_Q --> DB_TIDB

    WS_HUB -.->|WebSocket Real-time Feed| UI
    UI -->|API Requests| R_PROXY
```

---

## AI Orchestration Flow

The "In-the-Round" orchestration describes how a message from an external user is consumed by the AI, verified against security and anti-hallucination thresholds, and either auto-replied or escalated to a live agent.

```mermaid
sequenceDiagram
    autonumber
    actor Customer as WhatsApp / Widget User
    participant WH as Webhooks / Backend Handler
    participant AI as AI Brain & Llama 3.3
    participant VAL as Response Validator
    participant DB as TiDB Database
    participant WS as WebSocket Broadcast Hub
    participant Client as React Dashboard (Agent View)
    actor Agent as Support Agent

    Customer->>WH: Sends message (e.g. "Do you ship to Lagos?")
    WH->>DB: Logs incoming message as unread
    WH->>WS: Broadcasts `new_message` event
    WS->>Client: Triggers real-time conversation list refresh

    WH->>AI: Prompts AI Brain with Query & Context
    activate AI
    AI->>DB: Semantic database search matching QA pairs
    DB-->>AI: Returns closest matching candidate (Confidence score)

    alt Confidence Score >= 0.70 (High Confidence Answering)
        AI-->>WH: Generates high-confidence response
        WH->>VAL: Validate response for hallucinated prices
        VAL-->>WH: Passes validation (confidence preserved)
        WH->>DB: Inserts AI message into `messages`
        WH->>WS: Broadcasts `new_message` to websocket
        WS->>Client: Automatically displays AI's answer
        WH->>Customer: Delivers reply back to customer chat
    else Confidence Score < 0.70 (Anti-Hallucination Guardrail)
        AI-->>WH: Flags Query as Ambiguous / Unknown
        deactivate AI
        WH->>DB: Updates Conversation Status to 'escalated'
        WH->>DB: Inserts task in `unknown_questions` queue
        WH->>DB: Inserts fallback warning text
        WH->>WS: Broadcasts `unknown_question` & `integration_update`
        WS->>Client: Visual pulse alert, escalated badge
        WH->>Customer: "I'm passing you to a human agent..."

        Agent->>Client: Selects conversation thread
        Client->>WH: GET /chats/conversations/:id
        WH->>DB: Calls MarkRead, updates unread to 0
        DB-->>Client: Returns conversation thread & messages
        Agent->>WH: Clicks 'Takeover' conversation
        WH->>DB: Sets `taken_over_by` as Agent ID

        Agent->>Client: Resolves ticket & submits corrected answer
        Client->>WH: POST /training/unknown-questions/:id/train
        WH->>DB: Inserts new QA pair, updates status to 'trained'
        WH-->>Agent: Training complete! Knowledge Base updated.
    end
```

---

## Security Features

### Authentication & Session Management

| Feature | Implementation |
|---|---|
| **JWT-based auth** | Access tokens (24h default, 7d with remember-me) + refresh tokens (7d / 30d) |
| **Cookie security** | `SameSite=LaxMode`, `Secure` in production, `HttpOnly` |
| **Token revocation** | Redis blacklist with in-memory fallback when Redis is unavailable |
| **Password hashing** | bcrypt with cost factor 12 |
| **Password reset** | Single-use tokens stored in Redis with 1-hour TTL; rate-limited to 3/hour per email |
| **Force password change** | `must_change_password` flag enforced at login and route level |

### API Protection

| Feature | Implementation |
|---|---|
| **CSRF** | Origin/Referer header validation on all POST/PUT/PATCH/DELETE requests |
| **Rate limiting** | Redis-based sliding window with in-memory fallback (3 req/hour for password reset, 500 req/min per user for chats) |
| **Request body limit** | 1 MB cap enforced before any JSON parsing |
| **XSS sanitization** | Reflection-based struct sanitizer strips HTML, event handlers, script tags from all input |
| **File upload** | 2 MB limit, `.csv` only, content sniffing (512 bytes), CSV data sanitized before persistence |
| **Error hardening** | All internal errors return generic `"An unexpected error occurred"` — no stack or detail leakage |
| **CORS** | Configurable allowlist, credentials enabled |

### Content Security

- **CSP**: `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; img-src 'self' data: blob:; connect-src 'self' api.groq.com; frame-ancestors 'none'`
- **Headers**: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`
- **Permissions**: `camera=(), microphone=(), geolocation=()`

### Redis Failure Resilience

When Redis is unavailable, the system degrades gracefully:

| Feature | Fallback |
|---|---|
| **Rate limiting** | In-memory sliding window with periodic cleanup |
| **Token blacklist** | In-memory map with 24h TTL and 10-minute cleanup goroutine |
| **Auth middleware** | Falls back to in-memory blacklist check when Redis is nil |
| **Conversation history** | Loaded from MySQL instead of Redis cache |

---

## Core Subsystems

### 1. AI Brain & Anti-Hallucination Guardrail

The Go backend implements a LangChain-style orchestration layer:

- **Semantic Search**: Queries `qa_pairs` using full-text search, returns closest match with confidence score.
- **Confidence Threshold**: Strict boundary of **0.70**. Below it, the system rejects AI generation, logs the gap, and escalates.
- **Response Validation**: Extracts price claims (e.g. `₦1,500`) from AI responses using regex and cross-references against the user's inventory. If a claimed price doesn't match any inventory item, the response is replaced with a safe fallback.
- **Hallucination Signals**: Phrases like "according to my training", "based on my knowledge", "generally speaking" reduce confidence by 30%.
- **Context Window**: Last 10 conversation turns stored in Redis (3-day TTL), falls back to MySQL.

### 2. Multi-Channel Integration Hub

Supports bidirectional messaging across:

| Channel | Protocol | Webhook | Polling |
|---|---|---|---|
| WhatsApp | OpenWA API (self-hosted) | Yes | No |
| Telegram | Bot API | Yes | No |
| Facebook Messenger | Graph API | Yes | No |
| Instagram | Graph API | Yes | No |
| Web Widget | Custom JS SDK + WebSocket | No | Yes |

### 3. Background Job Scheduler

Built-in job queue with Redis persistence handles recurring maintenance:

| Job | Interval | Purpose |
|---|---|---|
| `health_check` | 5 min | Tests all integration channel connectivity |
| `cache_cleanup` | 15 min | Evicts stale cache entries |
| `handoff_reminder` | 15 min | Sends reminders for pending handoffs |
| `check_credit_expiry` | 24 h | Expires stale credit balances |
| `process_campaigns_start` | 24 h | Activates scheduled broadcast campaigns |
| `db_cleanup_all` | 6 h | Runs all data retention cleanups |
| `openwa_webhook_repair` | 30 min | Repairs OpenWA webhook URLs |
| `free_weekly_reset` | 7 d | Resets free plan weekly usage counters |

### 4. Real-Time WebSocket Hub

- **Connection Management**: Single hub instance, client map keyed by remote address.
- **Keepalive**: Ping messages every 30 seconds.
- **Broadcast**: Channel with 256-buffer, drops messages when full (non-blocking).
- **Reconnection**: Frontend uses exponential backoff starting at 1s, doubling per attempt, capped at 30s, reset on successful connect.

### 5. Plan Gating & Credit System

| Plan | Price | Inventory Limit | Team Seats | Channels |
|---|---|---|---|---|
| Free | ₦0 | 10 items | 1 (owner only) | Web Widget |
| Pulse | Via Polar | 50 items | Up to 3 | Web + Telegram |
| Pro | Via Polar | 500 items | Up to 10 | All channels |
| Enterprise | Via Polar | Unlimited | Unlimited | All channels + custom |

- Limits enforced at API layer; frontend shows upgrade modal on quota exceeded.
- Credit packs purchasable for per-message AI usage.
- Polar webhooks verified via HMAC-SHA256 signatures.

### 6. Data Retention & Cleanup

Automatic cleanup schedules (configurable via environment variables):

| Entity | Retention | Action |
|---|---|---|
| Resolved conversations | 90 days | Soft-delete |
| Abandoned conversations | 30 days | Soft-delete |
| Orphaned messages | Immediate | Delete when parent conversation is removed |
| Unknown questions | 30 days | Delete |
| Expired handoffs | 7 days | Close |
| Audit logs | 90 days | Archive/delete |
| Notifications | 30 days | Delete |
| Inactive integrations | 30 days | Deactivate |
| Expired trials | 14 days after expiry | Downgrade to free |
| Expired credits | 365 days | Reset to zero |
| Completed campaigns | 30 days after end | Archive |

### 7. OpenWA Messaging Pipeline

The OpenWA (self-hosted WhatsApp API) subsystem is a production-grade messaging pipeline with six integrated layers:

```mermaid
graph TD
    subgraph Inbound ["Webhook Intake"]
        WH[WhatsApp Webhook<br/>POST /api/v1/openwa/webhook]
        HMAC[HMAC-SHA256<br/>Signature Verification]
    end

    subgraph Core ["Messaging Core"]
        CB[Circuit Breaker<br/>5 failures → 60s block]
        RL[Rate Limiter<br/>Sliding Window / session]
        SQ[Send Queue<br/>Redis FIFO + Priority]
        DLQ[Dead-Letter Queue<br/>Max 5 retries]
        WK[Session Worker<br/>Poll: 200ms]
    end

    subgraph Session ["Session Layer"]
        SM[Session Manager<br/>Health Check: 30s interval]
        AR[Auto-Reconnect<br/>Backoff: 30s→30m max]
        QR[QR Code Store<br/>Redis TTL: 5 min]
    end

    subgraph Media ["Media Pipeline"]
        MH[Media Handler<br/>Download + MIME Detect]
        FS[File System Store<br/>Configurable Retention]
        TH[Thumbnail Gen<br/>Nearest-neighbor, ≤400px]
    end

    subgraph Templates ["Template Engine"]
        TS[Template Service<br/>CRUD + Variable Substitution]
        IM[Interactive Messages<br/>List / Buttons / Catalog]
        CL[Common Template Library<br/>6 built-in templates]
    end

    subgraph Campaign ["Campaign Broadcast"]
        CM[Campaign Bridge<br/>Batch: 50 msg / 2s spread]
        OT[Opt-Out Tracker<br/>STOP keyword detection]
        AN[Delivery Analytics<br/>Read rate, failure rate]
    end

    WH --> HMAC
    HMAC --> CB
    CB -->|enqueue| SQ
    SQ -->|dequeue| WK
    WK -->|send| OWA[OpenWA Server]
    WK -->|download| MH

    SM -->|health poll| OWA
    SM -->|failure| AR
    AR -->|reconnect| OWA

    RL -.->|per-session limits| WK
    TS -->|render template| WK
    CM -->|batch enqueue| SQ
    CM -->|check opt-out| OT
```

#### Layer Details

| Layer | Files | Persistence | Key Features |
|---|---|---|---|
| **Rate Limiter** | `openwa_queue.go` | In-memory sliding window | 20 text/min, 10 media/min, 30 template/min, burst 5 |
| **Send Queue** | `openwa_queue.go` | Redis (FIFO) | Priority levels, exponential backoff (max 5), dead-letter |
| **Session Manager** | `openwa_session.go` | Redis (status) | 30s health poll, auto-reconnect after 3 failures |
| **Media Handler** | `openwa_media.go` | File system | MIME detection, thumbnail generation, 90d cleanup |
| **Template Service** | `openwa_templates.go` | DB (repository) | CRUD, variable substitution, interactive messages |
| **Campaign Bridge** | `openwa_campaign.go` | DB (repository) | 50 msg/batch, 2s spread, 20% failure threshold |
| **Circuit Breaker** | `openwa.go` | In-memory | 5 consecutive failures → 60s open state |

#### Message Flow (Outbound)

```
Client Request → EnqueueMessage()
  → RateLimiter.Allow()          # sliding window check
  → Queue.Push()                 # Redis FIFO with priority
  → SessionWorker.process()      # polling every 200ms
    → CircuitBreaker.Call()      # wraps HTTP call
      → HTTP POST to OpenWA      # connection pooled, 30s timeout
        → RateLimit header parse # tracks X-RateLimit-Remaining
```

#### Redis Key Layout

| Key Pattern | TTL | Purpose |
|---|---|---|
| `queue:<sessionID>` | Persistent | FIFO message queue per session |
| `ratelimit:<sessionID>:<type>` | 1 min | Sliding window counters |
| `session:state:<sessionID>` | Persistent | Connection state & metrics |
| `session:qr:<sessionID>` | 5 min | QR code for reconnection |
| `retry:<msgID>` | 24 h | Retry attempt counter |
| `deadletter:<sessionID>` | 7 d | Failed messages |---

## Configuration Reference

All configuration is via environment variables. Copy `backend/.env.example` to `backend/.env`.

### Server

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `NODE_ENV` | `development` | `production` enables ReleaseMode |
| `APP_URL` | `http://localhost:8080` | Public-facing URL |
| `CORS_ORIGINS` | `http://localhost:5173` | Comma-separated allowed origins |

### Database (TiDB)

| Variable | Default | Description |
|---|---|---|
| `DB_DSN` | — | TiDB connection string |
| `DB_POOL_SIZE` | `20` | Max open connections |
| `DB_MAX_IDLE` | `5` | Max idle connections |
| `DB_CONN_MAX_LIFETIME` | `5m` | Connection max lifetime |

### Redis

| Variable | Default | Description |
|---|---|---|
| `REDIS_URL` | — | Redis connection URL |
| `REDIS_SHORT_TTL` | `259200` | AI conversation history TTL (seconds, default 3 days) |
| `CACHE_TTL` | `300` | Generic cache TTL (seconds) |

### Authentication

| Variable | Default | Description |
|---|---|---|
| `JWT_SECRET` | — | HMAC signing key (min 32 chars) |
| `JWT_ACCESS_TTL` | `24h` | Access token TTL |
| `JWT_REFRESH_TTL` | `168h` | Refresh token TTL (7 days) |

### AI / Groq

| Variable | Default | Description |
|---|---|---|
| `GROQ_API_KEY` | — | Groq API key (comma-separated for round-robin) |
| `GROQ_MODEL` | `llama3-70b-8192` | Model identifier |

### Integrations

| Variable | Default | Description |
|---|---|---|
| `OPENWA_BASE_URL` | — | OpenWA self-hosted server URL |
| `OPENWA_API_KEY` | — | OpenWA API authentication key |
| `OPENWA_WEBOOK_SECRET` | — | OpenWA webhook verification secret |
| `OPENWA_RATE_LIMIT_TEXT` | `20` | Text messages per minute per session |
| `OPENWA_RATE_LIMIT_MEDIA` | `10` | Media messages per minute per session |
| `OPENWA_RATE_LIMIT_TEMPLATE` | `30` | Template messages per minute per session |
| `OPENWA_RATE_LIMIT_BURST` | `5` | Burst allowance over sliding window |
| `OPENWA_QUEUE_MAX_DEPTH` | `10000` | Max queued messages per session |
| `OPENWA_MEDIA_DIR` | `./media` | File system path for media storage |
| `OPENWA_MEDIA_RETENTION_DAYS` | `90` | Media file retention period |
| `OPENWA_SESSION_HEALTH_INTERVAL` | `30` | Session health poll interval (seconds) |
| `OPENWA_MAX_RECONNECT_ATTEMPTS` | `10` | Max reconnection attempts before giving up |
| `OPENWA_CONNECTION_POOL_SIZE` | `10` | Max idle HTTP connections to OpenWA |
| `OPENWA_CONNECTION_TIMEOUT` | `30` | TCP connection timeout (seconds) |
| `OPENWA_REQUEST_TIMEOUT` | `60` | HTTP request timeout (seconds) |
| `TELEGRAM_BOT_TOKEN` | — | Telegram bot token |
| `FACEBOOK_ACCESS_TOKEN` | — | Facebook Graph API token |
| `INSTAGRAM_ACCESS_TOKEN` | — | Instagram Graph API token |

### Email

| Variable | Default | Description |
|---|---|---|
| `EMAIL_PROVIDER` | `resend` | `resend` or `smtp` |
| `RESEND_API_KEY` | — | Resend.com API key |
| `SMTP_HOST` | — | SMTP server host |
| `SMTP_PORT` | `587` | SMTP server port |
| `SMTP_USER` | — | SMTP username |
| `SMTP_PASS` | — | SMTP password |
| `FROM_EMAIL` | — | Sender email address |

### Payments

| Variable | Default | Description |
|---|---|---|
| `POLAR_ACCESS_TOKEN` | — | Polar API authentication token |
| `POLAL_WEBHOOK_SECRET` | — | Polar webhook signing secret |
| `POLAR_ORGANIZATION_ID` | — | Polar organization identifier |

---

## Installation & Deployment

### Prerequisites

- Go 1.22+
- Node.js 18+ (with npm)
- MySQL client or TiDB Cloud access
- Redis (Upstash or self-hosted)

### Local Development

**Backend:**

```bash
cd backend
cp .env.example .env
# Edit .env with your credentials
go mod download
go run main.go
```

**Frontend:**

```bash
cd frontend
npm install
npm run dev
```

**Run all migrations:**

```bash
for f in backend/migrations/*.sql; do
  mysql -h your-tidb-host -P 4000 -u your-user -p noant < "$f"
done
```

### Production Build (Unified Server)

```bash
cd frontend && npm install && npm run build
cd ../backend && mkdir -p static && cp -r ../frontend/dist/* ./static/
go build -o bin/main main.go
./bin/main
```

The backend serves the compiled React SPA from `./static/` on the same port — no separate frontend hosting needed.

### Docker

```bash
# Build
docker build -f backend/Dockerfile -t noant-backend .
docker build -f frontend/Dockerfile -t noant-frontend .

# Or use docker-compose (if available)
docker compose up -d
```

---

## Testing

```bash
# All backend tests
cd backend && go test -count=1 ./...

# Specific packages
go test -count=1 ./internal/infrastructure/...
go test -count=1 ./internal/middleware/...
go test -count=1 ./internal/service/...
go test -count=1 ./internal/utils/...

# Frontend type check
cd frontend && npx tsc --noEmit

# Frontend tests (if configured)
npm test
```

### Test Coverage Areas

| Package | Tests | What's Covered |
|---|---|---|
| `infrastructure` | Rate limiter, blacklist | Allow/deny logic, window reset, key isolation, cleanup, TTL expiry |
| `middleware` | Body limit | Safe method pass-through, oversized rejection, streaming truncation |
| `service` | AI sales, duplicate reply, WhatsApp identity, Polar | Business logic correctness |
| `utils` | Sanitization | XSS stripping, control char removal, struct traversal, validation regexes |

---

## API Directory

All endpoints are prefixed with `/api/v1` and return JSON. Authentication via `Authorization: Bearer <token>` header or `noant_access` cookie.

### Authentication

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/auth/register` | Create tenant account | No |
| `POST` | `/auth/login` | Log in (supports `remember_me`) | No |
| `POST` | `/auth/refresh` | Refresh session token | Yes |
| `POST` | `/auth/logout` | Revoke tokens (Redis + in-memory blacklist) | Yes |
| `POST` | `/auth/change-password` | Change password (clears must_change_password flag) | Yes |
| `POST` | `/auth/forgot-password` | Request password reset (rate-limited 3/hour) | No |
| `POST` | `/auth/reset-password` | Complete password reset with token | No |
| `POST` | `/auth/verify-email` | Verify email with code | No |
| `GET` | `/auth/me` | Get current user profile | Yes |
| `PUT` | `/auth/profile` | Update profile | Yes |

### Chats & Conversations

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/chats/conversations` | List conversations (paginated, `page`/`limit`) | Yes |
| `GET` | `/chats/conversations/:id` | Get conversation with messages (paginated) | Yes |
| `POST` | `/chats/conversations/:id/messages` | Send message | Yes |
| `PUT` | `/chats/conversations/:id/takeover` | Human takeover from AI | Yes |
| `POST` | `/chats/conversations/:id/escalate` | Escalate conversation | Yes |
| `POST` | `/chats/direct-chat` | Start test chat with AI | Yes |
| `DELETE` | `/chats/clear` | Clear all conversations | Yes |
| `POST` | `/chats/typing` | Broadcast typing indicator via WebSocket | Yes |

### Training

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/training/categories` | List QA categories | Yes |
| `POST` | `/training/categories` | Create category | Yes |
| `PUT` | `/training/categories/:id` | Update category | Yes |
| `DELETE` | `/training/categories/:id` | Delete category | Yes |
| `GET` | `/training/qa` | List QA pairs (filterable by category) | Yes |
| `POST` | `/training/qa` | Create QA pair | Yes |
| `PUT` | `/training/qa/:id` | Update QA pair | Yes |
| `DELETE` | `/training/qa/:id` | Delete QA pair | Yes |
| `POST` | `/training/bulk-qa` | Bulk import QA pairs | Yes |
| `POST` | `/training/csv-upload` | Upload CSV with questions/answers | Yes |
| `GET` | `/training/search` | Search QA pairs | Yes |
| `GET` | `/training/unknown-questions` | List knowledge gaps | Yes |
| `POST` | `/training/unknown-questions/:id/train` | Train & resolve unknown question | Yes |

### Integrations

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/integrations/list` | List all channel integrations | Yes |
| `POST` | `/integrations/connect` | Connect a channel | Yes |
| `POST` | `/integrations/disconnect/:channel` | Disconnect a channel | Yes |
| `PUT` | `/integrations/:id` | Update integration config | Yes |
| `DELETE` | `/integrations/:id` | Remove integration | Yes |

### Widget Config

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/widget/config` | Get widget configuration | Yes |
| `POST` | `/widget/config` | Update widget configuration | Yes |
| `GET` | `/widget/public/:id` | Public widget config (no auth) | No |

### Payments & Billing

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/payments/plans` | List subscription plans | No |
| `POST` | `/payments/subscribe` | Initiate subscription checkout | Yes |
| `POST` | `/payments/webhook` | Polar webhook receiver | No (HMAC verified) |
| `GET` | `/payments/status` | Current subscription status | Yes |

### Inventory

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/inventory` | List inventory items | Yes |
| `POST` | `/inventory` | Create inventory item | Yes |
| `GET` | `/inventory/search` | Search inventory | Yes |
| `GET` | `/inventory/:id` | Get item details | Yes |
| `PUT` | `/inventory/:id` | Update item | Yes |
| `DELETE` | `/inventory/:id` | Delete item | Yes |

### Handoffs & Leads

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/handoffs` | List handoffs/leads | Yes |
| `GET` | `/handoffs/:id` | Get handoff details | Yes |
| `PUT` | `/handoffs/status` | Update handoff status | Yes |

### Credits

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/credits/balance` | Get credit balance | Yes |
| `GET` | `/credits/limits` | Get plan limits | Yes |
| `POST` | `/credits/purchase` | Purchase credit pack | Yes |
| `GET` | `/credits/history` | Credit transaction history | Yes |

### Analytics

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/analytics/overview` | Dashboard overview stats | Yes |
| `GET` | `/analytics/insights` | AI-generated business insights | Yes |

### Campaigns

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/campaigns` | List campaigns | Yes |
| `POST` | `/campaigns` | Schedule campaign | Yes |
| `DELETE` | `/campaigns/:id` | Cancel campaign | Yes |

### Teams

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/teams/members` | List team members | Yes |
| `POST` | `/teams/invite` | Invite team member | Yes |
| `DELETE` | `/teams/members/:id` | Remove team member | Yes |

### API Keys

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/api-keys` | List API keys | Yes |
| `POST` | `/api-keys` | Generate API key | Yes |
| `DELETE` | `/api-keys/:id` | Revoke API key | Yes |

### Notifications

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/notifications` | List notifications | Yes |
| `PUT` | `/notifications/:id/read` | Mark as read | Yes |

### WhatsApp (OpenWA)

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/channels/whatsapp/health` | WhatsApp connection health | Yes |
| `POST` | `/channels/whatsapp/verify` | Verify WhatsApp number | Yes |
| `POST` | `/channels/whatsapp/send-test` | Send test WhatsApp message | Yes |
| `POST` | `/openwa/webhook` | Inbound message webhook | No (HMAC) |
| `GET` | `/openwa/status` | Session connection status | Yes |
| `POST` | `/openwa/restart` | Force session restart | Yes |

### WhatsApp Templates

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/templates` | List message templates | Yes |
| `POST` | `/templates` | Create template | Yes |
| `GET` | `/templates/:id` | Get template by ID | Yes |
| `PUT` | `/templates/:id` | Update template | Yes |
| `DELETE` | `/templates/:id` | Delete template | Yes |
| `POST` | `/templates/:id/submit` | Submit for approval | Yes |
| `POST` | `/templates/send` | Send templated message | Yes |
| `GET` | `/templates/common` | List common templates | Yes |

### WhatsApp Campaigns

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/whatsapp/campaigns/broadcast` | Broadcast campaign | Yes |
| `GET` | `/whatsapp/campaigns/:id/analytics` | Campaign delivery analytics | Yes |

### WhatsApp Media

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/chats/conversations/:id/media` | Upload media to conversation | Yes |
| `GET` | `/chats/conversations/:id/media` | List conversation media | Yes |
| `GET` | `/chats/conversations/media/:mediaID` | Download media file | Yes |
| `GET` | `/chats/conversations/media/:mediaID/thumbnail` | Get media thumbnail | Yes |

### WhatsApp Admin

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/whatsapp/admin/queue/stats` | Queue depth & processed stats | Yes |
| `GET` | `/whatsapp/admin/sessions` | List managed sessions | Yes |
| `GET` | `/whatsapp/admin/sessions/:id/metrics` | Per-session metrics | Yes |
| `POST` | `/whatsapp/admin/sessions/:id/reconnect` | Force reconnection | Yes |

### WhatsApp Interactive

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `POST` | `/whatsapp/interactive/list` | Send interactive list message | Yes |
| `POST` | `/whatsapp/interactive/buttons` | Send interactive buttons message | Yes |

### System

| Method | Endpoint | Description | Auth |
|---|---|---|---|
| `GET` | `/health` | Health check (DB, Redis, Groq keys) | No |
| `GET` | `/ping` | Liveness probe | No |
| `GET` | `/metrics` | Prometheus metrics | No |

---

## Operations

### Health Checks

| Component | Check | Frequency |
|---|---|---|
| Database | `PingContext()` | On-demand via `/health` |
| Redis | `Ping()` | On-demand via `/health` |
| Groq API | Key format validation | On-demand via `/health` |
| Integrations (Telegram, WhatsApp, FB, IG) | API connectivity test | Every 5 min (background goroutine + job queue) |

### Redis Key Lifecycle

| Key Pattern | TTL | Created By |
|---|---|---|
| `refresh:<token>` | 7 days | Login, token refresh |
| `blacklist:<token>` | 24 hours | Logout, token revocation |
| `reset:<token>` | 1 hour | Forgot password |
| `conv:<id>:history` | 3 days | AI conversation turn storage |
| `cache:<key>` | 5 minutes | Generic cache |
| `ratelimit:*` | Window duration | Rate limit middleware |
| `user:<id>` | 5 minutes | User cache |
| `job:<id>` | 24 hours | Background job queue |
| `free_weekly:<uid>` | Until Monday midnight | Free plan counter |
| `queue:<sessionID>` | Persistent | OpenWA message queue (FIFO) |
| `ratelimit:<sessionID>:<type>` | 1 min | OpenWA sliding window counters |
| `session:state:<sessionID>` | Persistent | OpenWA connection state & metrics |
| `session:qr:<sessionID>` | 5 min | OpenWA QR code for reconnection |
| `retry:<msgID>` | 24 h | OpenWA retry attempt counter |
| `deadletter:<sessionID>` | 7 d | OpenWA failed messages |

### Circuit Breaker (AI API Calls)

- **Threshold**: 3 consecutive failures
- **State**: `closed` → `open` → `half-open` → `closed`
- **Recovery**: 60-second wait before transitioning to half-open on next request
- **Scope**: Per-instance, in-memory (not shared across replicas)

### WebSocket Reconnection

The frontend WebSocket client uses exponential backoff:

```
Attempt 1: 1s
Attempt 2: 2s
Attempt 3: 4s
Attempt 4: 8s
...
Cap: 30s
Reset: On successful connection
```

---

## Database Schema

### Entity Relationship

```
┌─────────────────────────────────┐       ┌─────────────────────────────────┐
│          users                  │       │          conversations          │
├─────────────────────────────────┤       ├─────────────────────────────────┤
│ id (PK)         VARCHAR(36)     │◄──┐   │ id (PK)         VARCHAR(36)     │◄──┐
│ email           VARCHAR(255)    │   │   │ user_id (FK)    VARCHAR(36)     │   │
│ password_hash   VARCHAR(255)    │   │   │ customer_name   VARCHAR(100)    │   │
│ first_name      VARCHAR(100)    │   │   │ customer_phone  VARCHAR(20)     │   │
│ last_name       VARCHAR(100)    │   │   │ customer_email  VARCHAR(255)    │   │
│ role            ENUM            │   │   │ channel         VARCHAR(50)     │   │
│ plan_id         VARCHAR(50)     │   │   │ status          ENUM            │   │
│ is_active       BOOLEAN         │   │   │ intent          VARCHAR(50)     │   │
│ is_verified     BOOLEAN         │   │   │ priority        ENUM            │   │
│ trial_expires_at TIMESTAMP      │   │   │ is_ai_transferred BOOLEAN       │   │
│ created_at      TIMESTAMP       │   │   │ taken_over_by   VARCHAR(36)    │   │
└─────────────────────────────────┘   │   │ folder_id       VARCHAR(36)     │   │
                                      │   │ created_at      TIMESTAMP       │   │
┌─────────────────────────────────┐   │   │ updated_at      TIMESTAMP       │   │
│          integrations           │   │   └─────────────────────────────────┘   │
├─────────────────────────────────┤   │                                         │
│ id (PK)         VARCHAR(36)     │   │   ┌─────────────────────────────────┐   │
│ user_id (FK)    VARCHAR(36)     ├───┘   │          messages               │   │
│ channel         VARCHAR(50)     │       ├─────────────────────────────────┤   │
│ status          ENUM            │       │ id (PK)         VARCHAR(36)     │   │
│ config          JSON            │       │ conversation_id VARCHAR(36)     ├───┘
│ created_at      TIMESTAMP       │       │ sender_type     ENUM            │
│ updated_at      TIMESTAMP       │       │ content         TEXT            │
└─────────────────────────────────┘       │ is_read         BOOLEAN         │
                                          │ confidence      DECIMAL         │
┌─────────────────────────────────┐       │ source          VARCHAR(50)     │
│          qa_pairs              │       │ created_at      TIMESTAMP       │
├─────────────────────────────────┤       └─────────────────────────────────┘
│ id (PK)         VARCHAR(36)     │
│ category_id (FK) VARCHAR(36)   │       ┌─────────────────────────────────┐
│ question        TEXT            │       │       categories               │
│ answer          TEXT            │       ├─────────────────────────────────┤
│ variations      JSON            │       │ id (PK)         VARCHAR(36)     │
│ is_active       BOOLEAN         │       │ user_id (FK)    VARCHAR(36)     │
│ usage_count     INT             │       │ name            VARCHAR(100)    │
│ created_at      TIMESTAMP       │       │ description     TEXT            │
└─────────────────────────────────┘       │ color           VARCHAR(7)      │
                                          │ created_at      TIMESTAMP       │
┌─────────────────────────────────┐       └─────────────────────────────────┘
│     unknown_questions          │
├─────────────────────────────────┤       ┌─────────────────────────────────┐
│ id (PK)         VARCHAR(36)     │       │          handoffs              │
│ user_id (FK)    VARCHAR(36)     │       ├─────────────────────────────────┤
│ question        TEXT            │       │ id (PK)         VARCHAR(36)     │
│ conversation_id VARCHAR(36)     │       │ conversation_id VARCHAR(36)     │
│ channel         VARCHAR(50)     │       │ customer_phone  VARCHAR(20)     │
│ status          ENUM            │       │ product_name    VARCHAR(255)    │
│ suggested_answer TEXT           │       │ original_price  DECIMAL         │
│ created_at      TIMESTAMP       │       │ agreed_price    DECIMAL         │
└─────────────────────────────────┘       │ status          ENUM            │
                                          │ created_at      TIMESTAMP       │
┌─────────────────────────────────┐       └─────────────────────────────────┘
│      inventory_items           │
├─────────────────────────────────┤       ┌─────────────────────────────────┐
│ id (PK)         VARCHAR(36)     │       │       widget_configs           │
│ user_id (FK)    VARCHAR(36)     │       ├─────────────────────────────────┤
│ type            ENUM            │       │ id (PK)         VARCHAR(36)     │
│ name            VARCHAR(255)    │       │ user_id (FK)    VARCHAR(36)     │
│ description     TEXT            │       │ config          JSON            │
│ price           DECIMAL         │       │ is_active       BOOLEAN         │
│ min_price       DECIMAL         │       │ created_at      TIMESTAMP       │
│ stock_quantity  INT             │       └─────────────────────────────────┘
│ is_active       BOOLEAN         │
│ created_at      TIMESTAMP       │       ┌─────────────────────────────────┐
└─────────────────────────────────┘       │       audit_logs               │
                                          ├─────────────────────────────────┤
┌─────────────────────────────────┐       │ id (PK)         VARCHAR(36)     │
│       credit_balances          │       │ user_id (FK)    VARCHAR(36)     │
├─────────────────────────────────┤       │ action          VARCHAR(100)    │
│ id (PK)         VARCHAR(36)     │       │ resource_type   VARCHAR(50)     │
│ user_id (FK)    VARCHAR(36)     │       │ resource_id     VARCHAR(36)     │
│ balance         DECIMAL         │       │ details         JSON            │
│ expires_at      TIMESTAMP       │       │ created_at      TIMESTAMP       │
│ last_updated_at TIMESTAMP       │       └─────────────────────────────────┘
└─────────────────────────────────┘
```

### Additional Tables

- `team_members` — Team invitations and role assignments
- `api_keys` — Programmatic API access keys
- `archived_conversations` — Soft-deleted conversation archive
- `subscriptions` — Polar subscription lifecycle records
- `credit_purchases` — Credit pack purchase history
- `campaign_schedules` — Broadcast campaign scheduling
- `notifications` — In-app notification center
- `fcm_tokens` — Firebase Cloud Messaging push tokens
- `fcm_preferences` — Per-user notification preferences

---

## License

MIT License. Built for African businesses.
