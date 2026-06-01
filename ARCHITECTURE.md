# NOANT Enterprise v2.0 — Complete Architecture Guide

> *Autonomous & Human-in-the-Loop AI Customer Support Platform*
> **Stack**: Go 1.22 (Gin) + React 18 (Vite) + TiDB + Redis + Groq Llama 3.3

---

## Table of Contents

1. [System Architecture Overview](#1-system-architecture-overview)
2. [Backend Architecture (Go)](#2-backend-architecture-go)
   - 2.1 Entry Point: main.go
   - 2.2 Layer 1: Handlers (HTTP Layer)
   - 2.3 Layer 2: Services (Business Logic)
   - 2.4 Layer 3: Repository (Data Access)
   - 2.5 Domain Models
   - 2.6 Middleware Stack
   - 2.7 Infrastructure Layer
   - 2.8 AI Brain (Groq Llama 3.3 Integration)
   - 2.9 WebSocket Hub
   - 2.10 Database Schema & Migrations
   - 2.11 Route Map (all 50+ endpoints)
3. [Frontend Architecture (React)](#3-frontend-architecture-react)
   - 3.1 Entry Point & App Bootstrap
   - 3.2 Routing Tree
   - 3.3 Context Layer (4 contexts)
   - 3.4 Custom Hooks (8 hooks)
   - 3.5 API Client & WebSocket Client
   - 3.6 UI Component Library
   - 3.7 Page-by-Page Breakdown
4. [Data Flow Sequences](#4-data-flow-sequences)
5. [Key Design Patterns](#5-key-design-patterns)
6. [Code Map — Every File & Its Purpose](#6-code-map--every-file--its-purpose)

---

## 1. System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    CLIENT GATEWAYS                               │
│   WhatsApp/Twilio   Telegram   Facebook   Web Widget (JS embed) │
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────────┐
│              NETWORK & SECURITY LAYER                            │
│   Reverse Proxy → JWT Auth Validator → Upstash Redis Rate Limiter│
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────────┐
│                  GO APPLICATION (Gin Framework)                  │
│                                                                  │
│   ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│   │  Handlers   │  │  Services    │  │     AI Brain          │  │
│   │  (HTTP)     │──│  (Business)  │──│  (Groq Llama 3.3)    │  │
│   └──────┬──────┘  └──────┬───────┘  └──────────────────────┘  │
│          │                │                                      │
│   ┌──────▼────────────────▼───────┐                              │
│   │       Repositories (DB)       │                              │
│   └──────┬────────────────┬───────┘                              │
│          │                │                                      │
│   ┌──────▼────┐    ┌──────▼──────┐    ┌──────────────────────┐  │
│   │  TiDB     │    │   Redis     │    │   WebSocket Hub      │  │
│   │  Cluster  │    │   Cache     │    │   (Broadcast Server) │  │
│   └───────────┘    └─────────────┘    └──────────┬───────────┘  │
└──────────────────────────────────────────────────┼──────────────┘
                                                   │
┌──────────────────────────────────────────────────▼──────────────┐
│                 REACT SPA (Vite)                                │
│   Dashboard: Overview · Chats · Teach · Insights · Channels    │
│   Settings · Notifications · Billing · Team · Widget           │
│   (Real-time via WebSocket + REST via httpOnly cookies)        │
└─────────────────────────────────────────────────────────────────┘
```

**Key principles:**
- **Decoupled layers**: Handlers → Services → Repositories → Database
- **Multi-tenant isolation**: Every query scoped by `user_id`
- **Hybrid AI model**: Semantic search + LLM intent classification + inventory-aware sales
- **Real-time push**: WebSocket hub broadcasts all events to connected dashboards
- **Optional Redis**: App degrades gracefully to in-memory caches if Redis is down
- **Zero hallucination**: AI only answers from training data and inventory — never from general knowledge

---

## 2. Backend Architecture (Go)

### 2.1 Entry Point: `backend/main.go`

**File**: `backend/main.go` (315 lines)

The `main()` function performs these steps in order:

```
1. config.Load()                    → Parse .env + env vars
2. infrastructure.NewLogger()       → Structured JSON logger (slog)
3. infrastructure.NewTiDBConnection → MySQL/TiDB connection pool (3 retries)
4. infrastructure.RunMigrations()   → Apply all .sql files in ./migrations/
5. CREATE TABLE IF NOT EXISTS audit_logs (direct fallback)
6. infrastructure.NewRedisClient()  → Optional; nil if unavailable
7. repository.NewRepositories()     → Wire all 14 repos with db + redis
8. repository.NewAuditRepository()
9. handler.NewWebSocketHub()        → Central broadcast hub
10. go wsHub.Run()                  → Start event loop goroutine
11. service.NewPolarService()       → Payment gateway (Polar.sh)
12. service.NewResendService()      → Email service (Resend)
13. service.NewServices()           → Wire all 11 services
14. infrastructure.NewCache()       → L1 memory + L2 Redis cache
15. infrastructure.NewBottleneck()  → Concurrency limiter (200 max, 1000 queue)
16. infrastructure.NewJobQueue()    → Background jobs (10 workers)
17. Register handlers + health check
18. Setup Gin engine with middleware stack
19. http.ListenAndServe()           → Start on :8080/:5000
20. Graceful shutdown (SIGINT/SIGTERM, 30s timeout)
```

**Background routines** (started via JobQueue):
- `health_check` every 5 minutes — pings Telegram, WhatsApp, Facebook, Instagram APIs
- `cache_cleanup` every 15 minutes — evicts stale cache entries
- `handoff_reminder` every 15 minutes — processes pending handoffs, sends reminders (max 3), auto-expires

### 2.2 Layer 1: Handlers (`internal/handler/`)

**File**: `internal/handler/handler.go` (1212 lines)

The `Handlers` struct aggregates all 13 sub-handlers:

```go
type Handlers struct {
    Auth         *AuthHandler
    Chat         *ChatHandler
    Training     *TrainingHandler
    Analytics    *AnalyticsHandler
    Integration  *IntegrationHandler
    Settings     *SettingsHandler
    Archive      *ArchiveHandler
    Payment      *PaymentHandler
    Audit        *AuditHandler
    Notification *NotificationHandler
    Widget       *WidgetHandler
    Inventory    *InventoryHandler
    Handoff      *HandoffHandler
}
```

**Handler pattern** (every handler follows this exactly):

```go
type XxxHandler struct {
    service *service.XxxService
    logger  *infrastructure.Logger
}

func NewXxxHandler(svc *service.XxxService, log *infrastructure.Logger) *XxxHandler {
    return &XxxHandler{service: svc, logger: log}
}

func (h *XxxHandler) SomeEndpoint(c *gin.Context) {
    // 1. Parse & validate request body (c.ShouldBindJSON)
    // 2. Call h.service.SomeMethod(ctx, args...)
    // 3. On error → utils.RespondXxx(c, message)
    // 4. On success → c.JSON(http.StatusOK, gin.H{...})
}
```

**AuthHandler** (7 endpoints):
- `Register` — validates email (required, email format), password (min 8), first/last name → calls `AuthService.Register` → returns 201
- `Login` — validates email + password → calls `AuthService.Login` → sets httpOnly cookies (`noant_access` + `noant_refresh`) → returns user + trial_info
- `RefreshToken` — reads refresh token from cookie or body → rotates tokens
- `Logout` — blacklists access token in Redis, deletes refresh token, clears cookies
- `ChangePassword` — requires current + new password (JWT auth)
- `ForgotPassword` — rate-limited (3/hr per email) → sends reset email via Resend
- `ResetPassword` — validates token (from Redis) + new password
- `Me` — returns current user profile + trial info

**ChatHandler** (6 endpoints):
- `DirectChat` — creates/finds active conversation → sends message → gets AI response → returns conversation + message
- `ListConversations` — paginated, enriched with `last_message` and `unread` count
- `GetConversation` — paginated messages, marks unread as read
- `SendMessage` — stores customer message, broadcasts `typing_indicator=true`, fires AI response **asynchronously** in goroutine, broadcasts result
- `HumanTakeover` — sets `taken_over_by` to agent ID
- `Escalate` — updates status to `escalated`, creates system message

**TrainingHandler** (12 endpoints):
- `SearchQAPairs`, `ListCategories`, `CreateCategory`, `DeleteCategory`
- `ListQAPairs`, `CreateQAPair`, `UpdateQAPair`, `DeleteQAPair`
- `BulkImport` — batch create Q&A pairs in transaction
- `ListUnknownQuestions`, `TrainUnknown` (converts question → Q&A pair), `IgnoreUnknown`
- `UploadCSV` — parses CSV (category, question, answer), auto-creates categories, upserts

**AnalyticsHandler** (4 endpoints):
- `Overview` — aggregate stats (totals, rates, response time, satisfaction)
- `ChannelDistribution` — COUNT + GROUP BY channel
- `Insights` — top 5 intents + peak hours (last 7 days)
- `Trends` — daily conversation counts for N days

**IntegrationHandler** (4 endpoints):
- `List` — all channels with status
- `Connect` — save channel config, broadcast `integration_update` via WebSocket
- `Disconnect` — remove channel config, broadcast via WebSocket
- `Test` — tests connectivity: Telegram (`getMe`), WhatsApp/Instagram/Facebook (Meta Graph API)

**SettingsHandler** (13 endpoints):
- Profile CRUD, API key management (`noant_`-prefixed keys), team management (invite/remove), notification preferences, account deletion, GDPR data export
- `InviteTeamMember` — creates team member record + sends invite email via Resend/SMTP

**ArchiveHandler** (6 endpoints):
- Folder CRUD (archive folders), move/remove chats, get status

**PaymentHandler** (4 endpoints):
- `ListPlans` — hardcoded Free/Starter/Pro/Enterprise with NGN pricing
- `Subscribe` — Polar.sh checkout → creates subscription
- `Webhook` — Polar.sh webhook events (subscription.created/active/cancelled/updated)
- `Status` — returns active subscription

**NotificationHandler** (4 endpoints):
- `List`, `UnreadCount`, `MarkRead`, `MarkAllRead`

**WidgetHandler** (4 endpoints):
- `Get` / `Upsert` (JWT auth) — read/write widget config
- `GetPublic` (API key) — public config for embedded widget
- `PublicChat` (API key) — unauthenticated chat via embedded widget

**InventoryHandler** (6 endpoints):
- `Create` — create product/service/package with name, price, min_price, stock
- `List` — list all inventory items with optional type filter
- `GetByID` — fetch single item
- `Update` — update item fields
- `Delete` — remove item
- `Search` — search items by name/description

**HandoffHandler** (3 endpoints):
- `List` — list handoffs with optional status filter (pending/sold/lost/expired)
- `GetByID` — fetch single handoff with customer and product details
- `UpdateStatus` — mark handoff as sold/lost/expired with notes and final price

**WebSocket handler** (`internal/handler/websocket.go`, 145 lines):
- `WebSocketHub` — central hub pattern: register/unregister/broadcast channels, client map, origin validation, 30s ping ticker
- `HandleWebSocket` — origin validation → upgrade → register → read loop (discards client messages — server→client only)
- `BroadcastMessage` — sends to all connected clients (channel capacity 256)

**Health handler** (`internal/handler/health.go`, 97 lines):
- Checks DB ping, Redis ping, Groq key validity
- Returns `"healthy"`, `"degraded"`, or 503

### 2.3 Layer 2: Services (`internal/service/`)

**File**: `internal/service/service.go` (1738 lines)

The `Services` struct aggregates all 13 services:

```go
type Services struct {
    Auth, Chat, Training, Analytics, Integration,
    Settings, Archive, Payment, Audit, Notification, Widget,
    Inventory, Handoff, OpenWA, Telegram
}
```

#### AIBrain — The Core AI Engine

```go
type AIBrain struct {
    cfg         *config.Config
    repos       *repository.Repositories
    redis       *infrastructure.RedisClient
    logger      *infrastructure.Logger
    keyIndex    int          // round-robin across Groq API keys
    keyMutex    sync.RWMutex
    cb          *CircuitBreaker  // 3 failures → 60s block
    broadcastFn func(convID, msgType string, data interface{})
    embeddings  *EmbeddingService  // semantic search
}
```

**AIBrain flow** (`GenerateResponse`) — Zero hallucination architecture:

```
Customer sends message
  │
  ▼
1. Greeting check → "Hi!" → local reply, no AI call
  │
  ▼
2. LLM Intent Classification (classifyIntent)
   ├── "handoff" → handleHandoff (create handoff record, notify owner)
   ├── "sales"   → handleSalesMode (search inventory, negotiate)
   └── "support" → default path
  │
  ▼
3. Support Mode:
   ├── Semantic search training data (embeddings + cosine similarity)
   │   └── Falls back to SQL LIKE if embeddings unavailable
   ├── Semantic search inventory
   ├── If training data match → return answer DIRECTLY (no Groq call)
   ├── If inventory match → return product info DIRECTLY (no Groq call)
   └── If nothing found → escalate + suggest similar Q&As
  │
  ▼
4. Sales Mode:
   ├── Search inventory (semantic search)
   ├── Build prompt with inventory context + min_price guardrails
   ├── Call Groq with strict "ONLY show inventory items" prompt
   ├── Validate response (no hallucinated prices/products)
   └── Local fallback if Groq fails (show inventory directly)
  │
  ▼
5. Handoff Mode:
   ├── Search inventory for mentioned product
   ├── Create handoff record in DB
   ├── Notify owner via WebSocket + notification
   └── Return: "Message [owner] on WhatsApp: [number]"
```

**Key AI rules enforced in prompts:**
- NEVER use general knowledge — only training data and inventory
- NEVER invent prices, products, or services
- NEVER say "I'm an AI" or mention platform names
- If training data contradicts general knowledge → training data wins
- Negotiation: offer small discounts, never below min_price
- Date/time aware for greetings and time-based context

**CircuitBreaker** states:
- `closed` → normal operation, allows requests
- `open` → after 3 consecutive failures, blocks all requests for 60s
- `half-open` → after 60s, allows one test request; success → closed, failure → open

**AuthService**:
- `Register` — duplicate email check → bcrypt hash (cost 12) → create user with 14-day trial
- `Login` — bcrypt compare → JWT (HS256, 24h access) + refresh token (32 random bytes hex, stored in Redis 7d)
- `RefreshToken` — validate refresh in Redis → rotate (new access + new refresh, old deleted)
- `Logout` — blacklist access token in Redis, delete refresh token
- `ForgotPassword` — rate-limited (3/hour/email) → generate reset token (Redis 1h TTL) → send via Resend
- `ResetPassword` — validate token → bcrypt + update → delete token

**ChatService**:
- `DirectChat` — rate limited (500/min per user, higher for business/enterprise) → find/create conversation → call AIBrain → store messages → escalate if low confidence
- `ListConversations` — paginated + enriched with last message + unread count
- `SendMessage` / `GenerateAIResponse` — store message → async AI response → broadcast via WebSocket

**TrainingService**:
- Full CRUD for categories and Q&A pairs
- `BulkImport` — transactional batch insert
- `UploadCSV` — parses CSV → auto-create categories → upsert Q&As (transactional)
- `TrainUnknown` — convert unknown question into Q&A pair

**AnalyticsService**:
- `Overview` — SQL aggregates for total conversations, today's count, active, resolved today, AI resolution rate (ratio of non-transferred), approximated response time (12.5–16.0s based on volume), satisfaction (94–98%)
- `ChannelDistribution` — `COUNT + GROUP BY channel`
- `Insights` — top 5 intents, peak hours (last 7 days)
- `Trends` — daily conversation counts for N days

**IntegrationService**:
- `Test` — channel-specific connectivity tests:
  - Telegram: `getMe` API with bot token
  - WhatsApp: Meta Graph API `/v21.0/{phone_number_id}`
  - Facebook: Meta Graph API page info
  - Instagram: Meta Graph API account info
  - Web: always succeeds

**WidgetService**:
- `Get` — returns config, creates default if none exists
- `GetByAPIKey` — public lookup for embedded widget
- `Upsert` — INSERT...ON DUPLICATE KEY UPDATE
- `PublicChat` — unauthenticated chat flow: validate API key → find/create conversation → rate limit (100/min) → store message → AI response → if escalated, create notification + send email

**PolarService** (`internal/service/polar.go`, 130 lines):
- `CreateCheckout` — creates Polar.sh checkout session → returns URL
- `VerifyWebhook` — HMAC-SHA256 signature verification
- `ProcessWebhook` — handles subscription.created/active/cancelled/updated events

**EmailService** (`internal/service/email.go`, 184 lines):
- Wraps Resend (primary) and SMTP (fallback)
- If `RESEND_API_KEY` is set → tries Resend first, falls back to SMTP on failure
- If no Resend key → uses SMTP directly
- `SendPasswordReset` — styled HTML email with reset link
- `SendNotificationEmail` — plain notification email

**ResendService** (`internal/service/resend.go`, 178 lines):
- Configurable `from` address via `RESEND_FROM` env var (default: `onboarding@resend.dev` — no domain needed)
- Configurable `apiURL` for reset links (uses `APIURL` from config)
- `SendPasswordReset` — styled HTML email with reset link
- `SendNotificationEmail` — plain notification email
- Both POST to `https://api.resend.com/emails` with Bearer auth

**VectorSearch** (`internal/service/vector.go`, 55 lines):
- Delegates to QAPair.Search (SQL LIKE)
- Falls back to word-by-word search for words > 4 chars
- TODO: Pinecone/Weaviate integration

**EmbeddingService** (`internal/service/embedding.go`, 392 lines):
- `GenerateEmbedding` — single text → vector via Groq `text-embedding-3-small`
- `GenerateEmbeddings` — batch text → vectors (one API call)
- `CosineSimilarity` — compares two vectors, returns 0.0–1.0
- `SemanticSearchQAPairs` — query embedding vs all Q&A embeddings → top matches above threshold (0.65)
- `SemanticSearchInventory` — same for inventory (threshold 0.6)
- `FindSimilarQA` — finds similar Q&As for unknown question suggestions
- **In-memory cache** per user, auto-refreshes every hour, invalidated on data changes
- Falls back to SQL LIKE if embedding API unavailable

**InventoryService**:
- Full CRUD for products/services/packages
- Cache invalidation on create/update/delete (triggers embedding rebuild)

**HandoffService**:
- `Create` — creates handoff record, notifies owner via WebSocket + notification
- `List` / `GetByID` — fetch handoffs with status filter
- `UpdateStatus` — mark sold/lost/expired, decrease inventory stock on sold
- `ProcessReminders` — background job: pending handoffs → reminders (max 3) → expire

### 2.4 Layer 3: Repository (`internal/repository/`)

**File**: `internal/repository/repository.go` (1170 lines)

```go
type Repositories struct {
    User, Conversation, Message, QAPair, Category, UnknownQ, Integration,
    Team, APIKey, Archive, Subscription, Audit, Notification, WidgetConfig,
    Inventory, Handoff
}
```

Every repository follows this pattern:
```go
type XxxRepository struct {
    db    *sql.DB
    redis *infrastructure.RedisClient
}

func NewXxxRepository(db *sql.DB, redis *infrastructure.RedisClient) *XxxRepository {
    return &XxxRepository{db: db, redis: redis}
}
```

**Key repositories:**

| Repository | Key Methods | Caching |
|---|---|---|
| `UserRepository` | Create, GetByEmail, GetByID, UpdateLastLogin, UpdatePassword, UpdateProfile, Delete, ExportUserData | Redis (GetByID) |
| `ConversationRepository` | Create, List (paginated + filtered), GetByID, UpdateStatus, FindActiveByCustomer, Takeover, GetOverview, CountByChannel, CountByIntent, CountByHour, CountByDate | No |
| `MessageRepository` | Create, ListByConversationPaginated, GetLastMessage, CountUnread, MarkRead | No |
| `QAPairRepository` | Create, BulkCreate, ListByCategory, Search (LIKE with punctuation cleanup), GetByQuestion, Update, IncrementUsage, Delete | No |
| `CategoryRepository` | GetByName, Create, List (with QACount via LEFT JOIN), Delete (cascade Q&As) | No |
| `UnknownQuestionRepository` | Create, List (filtered), UpdateStatus | No |
| `IntegrationRepository` | Create, ListByUser, ListActive, UpdateStatus, GetByUserAndChannel, Disconnect | No |
| `WidgetConfigRepository` | Get, GetByAPIKey, Upsert (ON DUPLICATE KEY UPDATE) | No |
| `InventoryRepository` | Create, GetByID, List (with type filter), Search (LIKE on name+description), Update, Delete, DecreaseStock | No |
| `HandoffRepository` | Create, GetByID, List (with status filter), UpdateStatus, GetPending, GetReadyForReminder, IncrementReminder, Expire | No |

**Unit of Work** (`internal/repository/uow.go`, 59 lines):
- Transaction wrapper: `Begin()`, `Commit()`, `Rollback()`
- Routes queries through active transaction if one exists

### 2.5 Domain Models (`internal/domain/models.go`)

All core domain structs with `json` and `db` tags:

| Struct | Purpose | Key Fields |
|---|---|---|
| `User` | Platform user (multi-tenant) | ID, Email, Password (json:"-"), FirstName, LastName, Role (owner/admin/agent), CompanyName, PlanID, TrialExpiresAt |
| `Conversation` | Chat session | ID, UserID, CustomerName/Phone/Email, Channel (telegram/whatsapp/web/instagram), Status (active/resolved/escalated/archived), Intent, Priority, TakenOverBy, Location, LastMessage, Unread |
| `Message` | Single message | ID, ConversationID, Role (ai/human/customer/system), Content, IsRead, Confidence, Source, Metadata |
| `QAPair` | Training knowledge base | ID, UserID, CategoryID, Question, Answer, Variations, Embedding(json:"-"), IsActive, UsageCount |
| `Category` | Q&A category | ID, UserID, Name, Description, Color, QACount |
| `UnknownQuestion` | AI knowledge gap | ID, UserID, Question, ConversationID, Channel, Status (pending/trained/ignored) |
| `Integration` | Connected channel | ID, UserID, Channel, Status, Config (map), WebhookURL, LastError |
| `WidgetConfig` | Chat widget settings | ID, UserID, BrandColor, Greeting, BotName, Position, WidgetAPIKey, IsActive |
| `InventoryItem` | Product/service/package | ID, UserID, Type (product/service/package), Name, Description, Price, MinPrice, StockQuantity, ImageURL, IsActive |
| `Handoff` | Sales handoff from AI to owner | ID, UserID, ConversationID, CustomerName/Phone/Whatsapp, ProductName, OriginalPrice, AgreedPrice, Quantity, Status (pending/sold/lost/expired), FinalPrice, OwnerNotes, ReminderCount, NextReminderAt |
| `AnalyticsOverview` | Dashboard stats | TotalConversations, ConversationsToday, ActiveConversations, ResolvedToday, AIResolutionRate, AvgResponseTime, Satisfaction |

### 2.6 Middleware Stack (`internal/middleware/`)

**auth.go** (316 lines):
- `AuthMiddleware()` — extracts JWT from Bearer header or `noant_access` cookie → validates signature → checks blacklist → verifies "access" type → sets `userID`, `email`, `role` in Gin context
- `GetAccessTokenFromRequest()` — checks Authorization header → falls back to cookie
- `GetRefreshTokenFromRequest()` — checks `noant_refresh` cookie → falls back to JSON body
- `SetAuthCookies()` / `ClearAuthCookies()` — httpOnly, SameSite=Lax, Secure if TLS
- `LoggerMiddleware()` — structured request logging (latency, status, IP, method, path)
- `SecurityHeaders()` — X-Content-Type-Options, X-Frame-Options, CSP, Referrer-Policy, Permissions-Policy
- `RateLimitMiddleware()` — IP-based rate limiting via Redis
- `RateLimitByUserMiddleware()` — user-based rate limiting via Redis
- `PlanEnforcementMiddleware()` — placeholder for plan-based access control
- `TrialExpirationMiddleware()` — returns 402 if trial expired

**audit.go** (69 lines):
- `AuditMiddleware()` — async logs POST/PUT/DELETE/PATCH actions to `audit_logs` table (goroutine with 5s timeout)

**websocket_auth.go** (66 lines):
- `WebSocketAuth()` — JWT validation for WebSocket upgrade requests

### 2.7 Infrastructure Layer (`internal/infrastructure/`)

| File | Purpose |
|---|---|
| `db.go` | TiDB connection with pool config (maxOpen=configurable, maxIdle=half), 3 retries |
| `redis.go` | Redis client wrapper: Get, Set, Delete, Exists, HSet/HGet, RateLimit, pipeline support |
| `cache.go` | L1 (in-memory) + L2 (Redis) cache with LRU-like eviction, hit/miss stats |
| `logger.go` | JSON structured logger wrapping slog, levels: debug/info/warn/error/fatal |
| `bottleneck.go` | Token-bucket concurrency limiter: max 200 concurrent, 1000 queue, per-group support |
| `jobqueue.go` | Background job processor: priority queue, 10 workers, exponential backoff (1s→2s→4s→30s), max 3 retries, recurring jobs, Redis persistence |
| `migrations.go` | Migration runner: tracks applied migrations in `schema_migrations` table, splits SQL by `;` |
| `metrics.go` | Prometheus counters: RequestsTotal, RequestDuration, AICallsTotal, AIDuration, DBConnections, RedisConnections |

### 2.8 WebSocket Hub (`internal/handler/websocket.go`)

```
WebSocketHub
  ├── clients      map[string]*websocket.Conn  (keyed by remote addr)
  ├── broadcast    chan WebSocketMessage        (capacity 256)
  ├── register     chan *Conn
  ├── unregister   chan *Conn
  └── allowedOrigins []string
```

**Message types broadcasted:**
- `new_message` — AI response or customer message delivered
- `typing_indicator` — AI thinking state (true/false)
- `unknown_question` — AI couldn't answer → training gap
- `integration_update` — channel connected/disconnected
- `notification` — system notification

### 2.9 Database Schema

**8 migration files** in `backend/migrations/`:

| Migration | Tables Created |
|---|---|
| `001_init.sql` | users, conversations, messages, categories, qa_pairs, unknown_questions, integrations, team_members, api_keys, archive_folders, subscriptions, payments |
| `003_audit_logs.sql` | audit_logs |
| `005_user_isolation.sql` | Adds user_id columns to categories, qa_pairs, unknown_questions (multi-tenant isolation) |
| `006_notifications_widget.sql` | notifications, widget_configs, user notification preference columns |
| `007_message_source.sql` | Adds `source` column to messages |
| `008_inventory_leads.sql` | inventory_items (products/services/packages), handoffs (sales leads), adds owner_whatsapp to users |

**Schema diagram (simplified):**

```
users ──── conversations ──── messages
  │            │
  ├──── qa_pairs ──── categories
  │
  ├──── unknown_questions
  │
  ├──── integrations
  │
  ├──── widget_configs
  │
  ├──── notifications
  │
  ├──── subscriptions ──── payments
  │
  ├──── inventory_items
  │
  ├──── handoffs
  │
  └──── team_members
```

### 2.10 Route Map

| Method | Path | Auth | Rate | Audited |
|---|---|---|---|---|
| **Auth** (10 req/min per IP) | | | | |
| POST | /api/v1/auth/register | No | IP | No |
| POST | /api/v1/auth/login | No | IP | No |
| POST | /api/v1/auth/refresh | No | IP | No |
| POST | /api/v1/auth/logout | No | IP | No |
| POST | /api/v1/auth/change-password | Yes | - | No |
| POST | /api/v1/auth/forgot-password | No | IP | No |
| POST | /api/v1/auth/reset-password | No | IP | No |
| GET | /api/v1/auth/me | Yes | - | No |
| **Chats** (500 req/min per user) | | | | |
| POST | /api/v1/chats/direct-chat | Yes | User | Yes |
| GET | /api/v1/chats/conversations | Yes | User | Yes |
| GET | /api/v1/chats/conversations/:id | Yes | User | Yes |
| POST | /api/v1/chats/conversations/:id/messages | Yes | User | Yes |
| PUT | /api/v1/chats/conversations/:id/takeover | Yes | User | Yes |
| POST | /api/v1/chats/conversations/:id/escalate | Yes | User | Yes |
| **Training** (30 req/min per user) | | | | |
| GET | /api/v1/training/search | Yes | User | Yes |
| GET | /api/v1/training/categories | Yes | User | Yes |
| POST | /api/v1/training/categories | Yes | User | Yes |
| DELETE | /api/v1/training/categories/:id | Yes | User | Yes |
| GET | /api/v1/training/categories/:id/qa | Yes | User | Yes |
| POST | /api/v1/training/qa | Yes | User | Yes |
| PUT | /api/v1/training/qa/:id | Yes | User | Yes |
| DELETE | /api/v1/training/qa/:id | Yes | User | Yes |
| POST | /api/v1/training/bulk-qa | Yes | User | Yes |
| POST | /api/v1/training/csv-upload | Yes | User | Yes |
| GET | /api/v1/training/unknown-questions | Yes | User | Yes |
| POST | /api/v1/training/unknown-questions/:id/train | Yes | User | Yes |
| POST | /api/v1/training/unknown-questions/:id/ignore | Yes | User | Yes |
| **Analytics** (60 req/min per user) | | | | |
| GET | /api/v1/analytics/overview | Yes | User | Yes |
| GET | /api/v1/analytics/channels | Yes | User | Yes |
| GET | /api/v1/analytics/insights | Yes | User | Yes |
| GET | /api/v1/analytics/trends | Yes | User | Yes |
| **Integrations** (30 req/min per user) | | | | |
| GET | /api/v1/integrations/list | Yes | User | Yes |
| POST | /api/v1/integrations/connect | Yes | User | Yes |
| POST | /api/v1/integrations/disconnect/:channel | Yes | User | Yes |
| POST | /api/v1/integrations/test/:channel | Yes | User | Yes |
| **Settings** (60 req/min per user) | | | | |
| GET/PUT | /api/v1/settings/profile | Yes | User | Yes |
| GET/POST/DELETE | /api/v1/settings/api-keys | Yes | User | Yes |
| GET/POST/DELETE | /api/v1/settings/team | Yes | User | Yes |
| GET/PUT | /api/v1/settings/notifications | Yes | User | Yes |
| DELETE | /api/v1/settings/account | Yes | User | Yes |
| GET | /api/v1/settings/account/export | Yes | User | Yes |
| GET | /api/v1/settings/audit-logs | Yes | User | Yes |
| **Notifications** | | | | |
| GET | /api/v1/notifications | Yes | - | No |
| GET | /api/v1/notifications/unread-count | Yes | - | No |
| POST | /api/v1/notifications/:id/read | Yes | - | No |
| POST | /api/v1/notifications/read-all | Yes | - | No |
| **Widget** | | | | |
| GET/POST | /api/v1/widget/config | Yes | - | No |
| GET | /api/v1/widget/public/config | API key | - | No |
| POST | /api/v1/widget/public/chat | API key | - | No |
| **Archive** | | | | |
| GET | /api/v1/archive/folders | Yes | - | Yes |
| POST/DELETE | /api/v1/archive/folders | Yes | - | Yes |
| POST | /api/v1/archive/move | Yes | - | Yes |
| POST | /api/v1/archive/remove | Yes | - | Yes |
| GET | /api/v1/archive/status | Yes | - | Yes |
| **Payments** | | | | |
| GET | /api/v1/payments/plans | No | - | No |
| POST | /api/v1/payments/subscribe | Yes | - | Yes |
| POST | /api/v1/payments/webhook | No | - | No |
| GET | /api/v1/payments/status | Yes | - | No |
| **Inventory** (60 req/min per user) | | | | |
| GET | /api/v1/inventory | Yes | User | Yes |
| POST | /api/v1/inventory | Yes | User | Yes |
| GET | /api/v1/inventory/search | Yes | User | Yes |
| GET | /api/v1/inventory/:id | Yes | User | Yes |
| PUT | /api/v1/inventory/:id | Yes | User | Yes |
| DELETE | /api/v1/inventory/:id | Yes | User | Yes |
| **Handoffs** (60 req/min per user) | | | | |
| GET | /api/v1/handoffs | Yes | User | Yes |
| GET | /api/v1/handoffs/:id | Yes | User | Yes |
| PUT | /api/v1/handoffs/status | Yes | User | Yes |
| **System** | | | | |
| GET | /health | No | - | No |
| GET | /metrics | No | - | No |
| GET | /ws | JWT query | - | No |

---

## 3. Frontend Architecture (React)

### 3.1 Entry Point & App Bootstrap

**File**: `frontend/index.html` → `frontend/src/main.tsx` → `frontend/src/App.tsx`

```
index.html
  └─ <div id="root">
      └─ main.tsx
           ├─ initTheme()  (reads localStorage 'noant_theme' or prefers-color-scheme)
           └─ <ErrorBoundary>
                └─ <NetworkProvider>
                     └─ <ToastProvider>
                          └─ <App />
                        </ToastProvider>
                   </NetworkProvider>
              </ErrorBoundary>
```

**App.tsx** — defines the router:

```tsx
const router = createBrowserRouter([
  {
    element: <AppShell />,    // OfflineBanner + CommandPalette + <Outlet />
    children: [
      // Auth routes (no protection, use AuthLayout)
      { path: '/login',  element: <AuthLayout><LoginPage /></AuthLayout> },
      { path: '/signup', element: <AuthLayout><SignupPage /></AuthLayout> },
      { path: '/forgot-password',  ... },
      { path: '/reset-password',   ... },
      // Dashboard routes (protected, use DashboardLayout)
      {
        path: '/',
        element: <ProtectedRoute>
                   <WidgetConfigProvider>
                     <SidebarAlertsProvider>
                       <DashboardLayout />
                     </SidebarAlertsProvider>
                   </WidgetConfigProvider>
                 </ProtectedRoute>,
        children: [
          { index: true, element: <OverviewPage /> },
          { path: 'chats',         element: <ChatsPage /> },
          { path: 'teach',        element: <TeachPage /> },
          { path: 'insights',     element: <InsightsPage /> },
          { path: 'channels',     element: <ChannelsPage /> },
          { path: 'setup',        element: <SetupPage /> },
          { path: 'settings',     element: <SettingsPage /> },
          { path: 'notifications',element: <NotificationsPage /> },
          { path: 'billing',      element: <BillingPage /> },
          { path: 'team',         element: <TeamPage /> },
          { path: 'widget',       element: <WidgetPage /> },
          { path: 'leads',        element: <LeadsPage /> },
          { path: 'inventory',    element: <InventoryPage /> },
        ]
      },
      { path: '*', element: <Navigate to='/' /> },
    ]
  }
])
```

**AppShell** runs a background session refresh every 20 minutes via `refreshToken()`.

### 3.2 Routing Tree

```
/login          → AuthLayout > LoginPage
/signup         → AuthLayout > SignupPage
/forgot-password → AuthLayout > ForgotPasswordPage
/reset-password  → AuthLayout > ResetPasswordPage
/               → DashboardLayout > OverviewPage
/chats          → DashboardLayout > ChatsPage
/teach          → DashboardLayout > TeachPage
/insights       → DashboardLayout > InsightsPage
/channels       → DashboardLayout > ChannelsPage
/setup          → DashboardLayout > SetupPage
/settings       → DashboardLayout > SettingsPage
/notifications  → DashboardLayout > NotificationsPage
/billing        → DashboardLayout > BillingPage
/team           → DashboardLayout > TeamPage
/widget         → DashboardLayout > WidgetPage
/leads          → DashboardLayout > LeadsPage
/inventory      → DashboardLayout > InventoryPage
```

### 3.3 Context Layer (4 contexts)

| Context | File | Purpose |
|---|---|---|
| `NetworkContext` | `contexts/NetworkContext.tsx` | Tracks `navigator.onLine`, exposes `isOnline` + `wasOffline` |
| `WidgetConfigContext` | `contexts/WidgetConfigContext.tsx` | Fetches/saves widget config from `/widget/config`, syncs with integrations |
| `SidebarAlertsContext` | `contexts/SidebarAlertsContext.tsx` | Polls + WebSocket for unread counts, unknown Qs, notifications, channel errors, billing. `channelIssues` counts only `status === 'error'` (not inactive/disconnected) |
| `ModalContext` | `contexts/ModalContext.tsx` | Imperative `showModal()` with confirm/cancel, loading, ESC/overlay-close |

### 3.4 Custom Hooks (8 hooks)

| Hook | File | Purpose |
|---|---|---|
| `useAPI` | `hooks/useAPI.ts` | Central data-fetching: auto-pagination, retry (max 3, exponential backoff 1s→2s→4s), toast on failure, GET dedup via inflight map |
| `useAuth` | `hooks/useAuth.ts` | Fetches current user on mount, returns `{user, loading, signOut, refreshUser}` |
| `useWebSocket` | `hooks/useWebSocket.ts` | Global WebSocket singleton with per-component subscribe/unsubscribe |
| `useInfiniteScroll` | `hooks/useInfiniteScroll.ts` | IntersectionObserver sentinel for load-more |
| `useKeyboardShortcuts` | `hooks/useKeyboardShortcuts.ts` | Register keyboard shortcuts with modifier matching |
| `useModal` | `hooks/useModal.ts` | Simple boolean toggle for local modal open/close |
| `useConfirm` | `hooks/useConfirm.ts` | Wrapper around ModalContext for imperative confirmations |
| `useNetwork` | `hooks/useNetwork.ts` | Re-exports from NetworkContext |

### 3.5 API Client & WebSocket Client

#### API Client (`lib/api.ts`)

```typescript
const API_BASE = import.meta.env.VITE_API_URL || '/api/v1';

// Singleton API object
api.get<T>(endpoint)      → GET request
api.post<T>(endpoint, body, isFormData?) → POST request
api.put<T>(endpoint, body) → PUT request
api.delete<T>(endpoint)    → DELETE request
```

**Features:**
- All requests use `credentials: 'include'` (httpOnly cookies for auth)
- GET requests deduplicated via `inflightRequests` Map (same endpoint → same Promise)
- **401 auto-refresh**: on 401, calls `refreshToken()` once, retries original request; if refresh fails, redirects to `/login`
- 3 retries with exponential backoff for 5xx/network errors
- Consistent error shape via `APIError` class (message + status + data)

#### WebSocket Client (`lib/websocket.ts`)

```typescript
class WebSocketManager {
  connect()       → Creates WebSocket to `ws[s]://host/ws`, sets up reconnect (3s delay)
  disconnect()    → Stops reconnection, closes socket
  onMessage(fn)   → Subscribes handler, returns unsubscribe function
}
```

**Key behaviors:**
- Singleton — one connection shared across all components
- Auto-reconnect on close (3s delay), stops when `disconnect()` called
- Message parsing handles snake_case, camelCase, and Go-style capitalized field names
- `useWebSocket` hook subscribes/unsubscribes per component (cleans up on unmount)

### 3.6 UI Component Library

All components live in `frontend/src/components/`:

**Layout** (`components/layout/`):
| Component | Purpose |
|---|---|
| `AuthLayout` | Split screen: dark brand panel (left) + form Outlet (right) |
| `DashboardLayout` | Shell: Sidebar + Header + BottomNav + `<Outlet />` |
| `Sidebar` | Nav links, sections (Your Workspace, Build Your Noant, Manage), alerts indicator, logout, collapsible on mobile |
| `Header` | Page title, notifications bell (with badge), theme toggle, user avatar dropdown |
| `BottomNav` | Mobile-only bottom tab bar (hides in chat thread view) |
| `MobileOverlay` | Backdrop overlay for mobile sidebar |

**Chat** (`components/chat/`):
| Component | Purpose |
|---|---|
| `ChatList` | Searchable conversation list with infinite scroll, status badges, unread counts |
| `ChatMessages` | Message thread with date separators, sender badges, confidence indicators, delivery status. Loading state uses `ConversationLoading` |
| `ChatInput` | Message input field + Send button + Takeover button + typing indicator |
| `CustomerInfo` | Customer details sidebar + conversation history timeline |
| `TypingIndicator` | Animated pulsing dots (WebSocket-driven), uses `dotPulse` keyframe animation |
| `ConversationLoading` | Theme-aware loading state: circle with pulsing dots. Sizes: sm/md/lg |

**Channels** (`components/channels/`):
| Component | Purpose |
|---|---|
| `ChannelCard` | Status card with connect/disconnect button, channel icon, last error |
| `ChannelIcon` | SVG icons for WhatsApp, Instagram, Telegram, Facebook, Discord, Web |
| `TokenDisplay` | Show/hide + copy-to-clipboard for tokens |
| `WhatsAppModal` | Connect form: phone_number_id, business_account_id, access_token |
| `TelegramModal` | Connect form: bot_token |
| `InstagramModal` | Connect form: instagram_id, page_access_token |
| `FacebookModal` | Connect form: page_id, page_access_token |
| `WebWidgetModal` | Configure web widget: position, bot name, greeting + embed code generator |

**Stats/Charts** (`components/stats/`):
| Component | Purpose |
|---|---|
| `StatCard` | Numeric stat card with label, value, change percentage arrow |
| `StatGrid` | 2-col (mobile) / 4-col (desktop) grid wrapper |
| `TrendChart` | Recharts AreaChart for conversation trends over time |
| `PeakHoursChart` | Recharts BarChart for hourly conversation volume |
| `ChannelDistributionChart` | Recharts PieChart (donut) for channel mix |
| `MetricRow` | Icon + label + value row for insight metrics |

**Training** (`components/training/`):
| Component | Purpose |
|---|---|
| `UploadZone` | Drag-and-drop CSV upload with progress bar |
| `UnknownQuestion` | Expandable card: train (answer + category) or ignore |
| `CategoryCard` | Color dot + name + Q&A count |
| `QAPair` | Question/answer display with left border color accent |

**UI Primitives** (`components/ui/`):
| Component | Purpose |
|---|---|
| `Toast` | Toast notification system (Context + Provider), auto-dismiss 4s, success/error/warning/info |
| `Button` | Variants (primary/accent/ghost/danger), sizes (sm/md/lg), loading spinner |
| `Input` | Styled text input with focus ring, error state |
| `Badge` | Status badges: sky/success/warning/error/neutral |
| `Card` | Card container + CardHeader/CardTitle/CardBody |
| `Skeleton` | Shimmer loading placeholder + StatSkeleton variant |
| `Avatar` | Gradient avatar with initials fallback, channel badge overlay |
| `Spinner` | SVG loading spinner with size variants |
| `Modal` | Portal modal: focus trap, ESC close, scroll lock, backdrop blur |
| `ConfirmModal` | Destructive confirmation dialog with type-to-confirm |
| `CommandPalette` | Cmd+K palette: search/filter across navigation, keyboard navigation |
| `ErrorBoundary` | Class-based error boundary with reload button |
| `OfflineBanner` | Sticky banner: red slide-down "offline" / green "back online" with fade |

### 3.7 Page-by-Page Breakdown

Each page follows this pattern:
```tsx
function SomePage() {
  const api = useAPI()
  const { data, get, loading } = api
  useEffect(() => { get('/api/v1/endpoint') }, [])
  if (loading) return <Skeleton />
  return <div>{/* render data */}</div>
}
```

#### Overview Page (`/`)

**File**: `frontend/src/app/(dashboard)/page.tsx` (209 lines)
**Data**: `GET /analytics/overview` + `GET /chats/conversations?limit=6`

Renders:
- **StatGrid** (4 cards): Conversations today, Resolved auto, Avg response time, Satisfaction
- **Quick actions**: Inbox, Channels, Teach, Insights (icon + label grid)
- **Recent activity table**: Desktop table / mobile list of recent conversations with channel icon, customer name, intent badge, status badge, timestamp
- **Empty state** with onboarding CTA when no conversations exist

#### Chats Page (`/chats`)

**File**: `frontend/src/app/(dashboard)/chats/page.tsx` (519 lines)
**Data**: `GET /chats/conversations` + `GET /chats/conversations/:id`

Three-panel layout:
1. **Left**: ChatList — searchable, filterable conversation list with infinite scroll, unread badges, channel icons
2. **Center**: ChatMessages + ChatInput — message thread with date separators, sender badges (Customer/AI/Agent/System), typing indicator, optimistic message sending
3. **Right (toggleable)**: CustomerInfo — customer details + history timeline

**Real-time features:**
- WebSocket subscription for `new_message`, `typing_indicator` events
- **AI response polling**: When message sent to AI, polls `GET /conversations/:id` every 800ms until AI response arrives (sets `pendingAI` set)
- Optimistic message rendering while waiting
- Mobile: single-column view with back button (URL search param `?id=` drives view state)

#### Teach Page (`/teach`)

**File**: `frontend/src/app/(dashboard)/teach/page.tsx` (~6500 lines, largest file)
**Data**: `GET /training/categories`, `GET /training/categories/:id/qa`, `GET /training/unknown-questions`, `POST /training/csv-upload`, `POST /knowledge-base/articles`

Three-tab layout:
1. **Categories**: Color-coded category cards with Q&A count, drag-to-reorder priority, create/delete
2. **Q&A Pairs**: Per-category list of question/answer pairs, create/edit/delete, bulk import
3. **Unknown Questions**: List of questions AI couldn't answer → train (provide answer + category) or ignore
4. **CSV Upload**: Drag-and-drop zone to bulk import Q&A pairs

#### Insights Page (`/insights`)

**File**: `frontend/src/app/(dashboard)/insights/page.tsx`
**Data**: `GET /analytics/overview`, `GET /analytics/trends`, `GET /analytics/insights`, `GET /analytics/channels`

Renders:
- **StatGrid**: Conversations today, Resolved auto, Avg response time, Satisfaction
- **TrendChart**: Recharts AreaChart (conversations over time)
- **PeakHoursChart**: Recharts BarChart (volume by hour)
- **ChannelDistributionChart**: Recharts PieChart (channel mix)
- **MetricRows**: Top intents, response times

#### Channels Page (`/channels`)

**File**: `frontend/src/app/(dashboard)/channels/page.tsx`
**Data**: `GET /integrations/list`

Renders:
- Grid of ChannelCard components for: WhatsApp, Instagram, Telegram, Facebook, Web Widget
- Each card shows: icon, channel name, status (connected/disconnected/error), connect/disconnect button
- Connect opens channel-specific modal with configuration fields
- Disconnect shows confirmation dialog
- Real-time WebSocket updates `integration_update` to refresh status

#### Setup Page (`/setup`)

**File**: `frontend/src/app/(dashboard)/setup/page.tsx`
**Data**: `GET /settings/profile`, `GET /settings/api-keys`, `GET /settings/team`

Tabbed layout:
1. **Profile**: Edit name, email, company, phone
2. **Team**: List team members with roles, invite new member, remove member
3. **Billing**: Current plan, subscription status
4. **API Keys**: Create/view/revoke API keys

#### Settings Page (`/settings`)

**File**: `frontend/src/app/(dashboard)/settings/page.tsx`
**Data**: `GET /settings/profile`, `GET /settings/notifications`

Tabbed layout:
1. **Profile**: Edit personal info
2. **Security**: Change password
3. **Notifications**: Toggle notification preferences (escalations, unknown questions, payments, security, team invites)
4. **Appearance**: Dark/light theme toggle
5. **Privacy**: Account deletion, data export

#### Notifications Page (`/notifications`)

**File**: `frontend/src/app/(dashboard)/notifications/page.tsx`
**Data**: `GET /notifications`

Renders:
- Notifications grouped by date
- Type-based icons (escalation, unknown_question, payment, team_invite, system)
- Mark-as-read / Mark-all-read buttons
- Click navigates to relevant page

#### Billing Page (`/billing`)

**File**: `frontend/src/app/(dashboard)/billing/page.tsx`
**Data**: `GET /billing/plan`

Renders:
- 4 pricing plans: Free, Starter, Pro, Business
- Feature comparison: AI responses, channels, team members
- NGN/USD pricing toggle
- "Popular" badge on recommended plan
- Subscribe button → Polar checkout

#### Team Page (`/team`)

**File**: `frontend/src/app/(dashboard)/team/page.tsx`
**Data**: `GET /team/members`

Renders:
- Team member list with avatar, name, email, role badge (Owner/Admin/Agent), status (Active/Pending)
- Invite member form (email + role selector)
- Remove member with confirmation dialog

#### Widget Page (`/widget`)

**File**: `frontend/src/app/(dashboard)/widget/page.tsx`
**Data**: `GET /widget/config` (via WidgetConfigContext)

Renders:
- Brand color picker (color input)
- Greeting message text input
- Bot name text input
- Position selector (bottom-left / bottom-right)
- Active/inactive toggle
- **Live preview**: Chat bubble preview in corner of page
- **Embed code**: Generated `<script>` tag for website integration

---

## 4. Data Flow Sequences

### 4.1 Customer Message → AI Response Flow

```
Customer sends "Do you ship to Lagos?"
  │
  ▼
Webhook/API → POST /api/v1/chats/conversations/:id/messages
  │
  ▼
ChatHandler.SendMessage
  ├── Parse & validate request
  ├── Store customer message in `messages` table
  ├── Broadcast "typing_indicator=true" via WebSocket
  ├── Launch goroutine: GenerateAIResponse
  │     │
  │     ▼
  │   ChatService.GenerateAIResponse
  │     │
  │     ▼
  │   AIBrain.GenerateResponse
  │     ├── 1. Fetch conversation + user
  │     ├── 2. Search Q&A (SQL LIKE match)
  │     │
  │     ├── [NO MATCH, confidence < 0.70]
  │     │   ├── Create UnknownQuestion record
  │     │   ├── Create Notification
  │     │   ├── Broadcast "unknown_question" via WebSocket
  │     │   └── Return fallback response
  │     │
  │     └── [MATCH, confidence ≥ 0.70]
  │         ├── Build LangChain prompt (system + 3 Q&A matches + 10-turn history)
  │         ├── Call Groq API (llama-3.3-70b-versatile)
  │         ├── Store conversation turn in Redis
  │         └── Return AI response
  │
  ├── Store AI message in `messages` table
  ├── Broadcast "new_message" + "typing_indicator=false" via WebSocket
  └── Return response to sender (via Twilio/Telegram/WebSocket)
```

### 4.2 WebSocket Real-time Update Flow

```
Backend Event                    WebSocket Hub                    React Dashboard
──────────────────────────────────────────────────────────────────────────────────
New message arrives
  │
  ▼
wsHub.BroadcastMessage({         ┌─────────────────┐
  Type: "new_message",           │  Hub event loop  │
  ConversationID: "...",         │  receives msg    │
  Data: { content: "..." }       │  iterates all    │
})                               │  connected WS    │
  │                              │  clients, sends  │
  ▼                              │  JSON message    │
hub.broadcast chan               └────────┬────────┘
                                          │
                                          ▼
                                   WebSocket.onmessage
                                          │
                                    WSMessage parsed
                                          │
                                    useWebSocket handlers
                                          │
                              ┌───────────┼───────────┐
                              ▼           ▼           ▼
                         ChatsPage    SidebarAlerts  Notification
                         (new msg)    (unread count)   (bell badge)
```

### 4.3 Authentication Flow

```
Login Page
  │
  ▼
POST /api/v1/auth/login  { email, password }
  │
  ▼
AuthService.Login
  ├── Find user by email
  ├── bcrypt.Compare(password, hash)
  ├── Generate JWT (24h, HS256, claims: user_id, email, role)
  ├── Generate refresh token (32 random bytes hex)
  ├── Store refresh in Redis (7d TTL)
  └── Set httpOnly cookies:
      ├── noant_access (JWT, 24h)
      └── noant_refresh (refresh token, 7d)
  │
  ▼
Response: { user, trial_info }
  │
  ▼
DashboardLayout renders
  ├── ProtectedRoute checks: useAuth().user exists
  ├── SidebarAlertsProvider starts polling
  └── WidgetConfigProvider fetches widget config
```

### 4.4 API Request Lifecycle

```
React Component
  │
  ▼
useAPI hook → api.get/post/put/delete
  │
  ├── [GET] Check inflightRequests map → deduplicate if pending
  │
  ▼
fetch(API_BASE + endpoint, { credentials: 'include' })
  │
  ▼
Gin Router → Middleware Stack:
  ├── RequestID → generate X-Request-ID
  ├── Logger → log method, path, latency
  ├── SecurityHeaders → CSP, XFO, etc.
  ├── CORS → origin validation
  ├── [auth routes] AuthMiddleware → JWT validation
  │     ├── Extract token from Cookie or Authorization header
  │     ├── Verify HS256 signature
  │     ├── Check Redis blacklist
  │     ├── Extract user_id, email, role → set in context
  │     └── [optional] TrialExpiration check
  ├── [audited routes] AuditMiddleware → async log to audit_logs
  │
  ▼
Handler → Service → Repository → TiDB/Redis
  │
  ▼
Gin Response (JSON)
  │
  ▼
fetch Response
  ├── [401] → auto-refresh → retry → [fail] redirect /login
  ├── [5xx] → retry with backoff (up to 3 times) → [fail] toast error
  └── [200] → JSON parsed → component re-renders
```

---

## 5. Key Design Patterns

### 5.1 Layered Architecture (Backend)

```
┌─────────────────────────────────────────┐
│           Handlers (HTTP)               │  ← Request parsing, response formatting
├─────────────────────────────────────────┤
│           Services (Business)           │  ← Business rules, AI orchestration
├─────────────────────────────────────────┤
│         Repositories (Data)             │  ← SQL queries, Redis caching
├─────────────────────────────────────────┤
│         Database / Cache                │  ← TiDB + Redis
└─────────────────────────────────────────┘
```

Each layer only talks to the layer directly below it. Handlers never call repositories directly.

### 5.2 WebSocket Hub Pattern

```
Central goroutine with 3 channels:
  register   chan *Client     → Add client to map
  unregister chan *Client     → Remove client from map
  broadcast  chan Message     → Send to all clients

Each client has its own send channel (buffered).
Hub never blocks — full broadcast channel drops messages (capacity 256).
```

### 5.3 Circuit Breaker Pattern

```
States: closed → open (3 failures) → half-open (after 60s)
  closed:    allow requests, count failures
  open:      reject all requests for 60s
  half-open: allow 1 request → success=closed, failure=open
```

### 5.4 Hook-based Data Fetching (Frontend)

```tsx
// Central pattern used on every page:
const api = useAPI()
const { data, get, loading, error, retry } = api

useEffect(() => { get('/endpoint') }, [])

if (loading) return <Skeleton />
if (error) return <Error onRetry={retry} />
return <Render data={data} />
```

### 5.5 Singleton WebSocket Manager

```
WebSocketManager (singleton)
  ├── One WebSocket connection for entire app
  ├── Set<MessageHandler> for subscribers
  ├── Auto-reconnect with 3s delay
  └── useWebSocket hook: subscribe/unsubscribe per component
```

### 5.6 Cookie-Based Auth

```
- httpOnly cookies prevent XSS token theft
- SameSite=Lax prevents CSRF
- Refresh token rotation: each use generates new pair
- Access token blacklist in Redis enables immediate logout
```

### 5.7 Round-Robin API Key Rotation

```
AIBrain.getNextAPIKey():
  key = keys[keyIndex]
  keyIndex = (keyIndex + 1) % len(keys)
  return key
```

---

## 6. Code Map — Every File & Its Purpose

### Backend (`backend/`)

```
backend/
├── main.go                    # Entry point: bootstrap, middleware, routes, graceful shutdown
├── go.mod / go.sum            # Module: noant (Go 1.25), deps: gin, jwt, gorilla/websocket, etc.
├── config/
│   └── config.go              # Config struct + Load() from env vars
├── internal/
│   ├── domain/
│   │   ├── models.go           # All domain models (User, Conversation, Message, QAPair, etc.)
│   │   └── audit.go            # AuditLog model
│   ├── handler/
│   │   ├── handler.go          # Handlers struct + Auth/Chat/Training/Analytics/Integration/Settings/Archive/Payment/Audit/Inventory/Handoff handlers
│   │   ├── websocket.go        # WebSocketHub: register/unregister/broadcast + HandleWebSocket
│   │   ├── health.go           # Health check handler
│   │   └── notifications.go    # Notification + Widget + SettingsNotif handlers
│   ├── service/
│   │   ├── service.go          # Services struct + AIBrain (Groq) + Auth/Chat/Training/Analytics/Integration/Settings/Archive/Inventory/Handoff
│   │   ├── embedding.go        # EmbeddingService: semantic search via Groq embeddings + cosine similarity
│   │   ├── notifications.go    # NotificationService + WidgetService + ResendService
│   │   ├── polar.go            # PolarService: Polar.sh payment integration
│   │   ├── resend.go           # ResendService: email sending
│   │   ├── vector.go           # VectorSearch: keyword search with word-by-word fallback
│   │   ├── retention.go        # RetentionService: data cleanup
│   │   └── 2fa.go              # TFAService: TOTP two-factor auth
│   ├── repository/
│   │   ├── repository.go       # All 14 repositories (User, Conversation, Message, QAPair, etc.)
│   │   ├── uow.go              # UnitOfWork: transaction wrapper
│   │   ├── audit.go            # AuditRepository
│   │   ├── notifications.go    # Notification + WidgetConfig repositories
│   │   ├── inventory.go        # InventoryRepository + HandoffRepository
│   ├── middleware/
│   │   ├── auth.go             # AuthMiddleware, RateLimit, SecurityHeaders, Logger, Cookies, etc.
│   │   ├── audit.go            # AuditMiddleware: async action logging
│   │   └── websocket_auth.go   # WebSocketAuth: JWT validation for WS upgrade
│   ├── infrastructure/
│   │   ├── db.go               # TiDB connection pool
│   │   ├── redis.go            # Redis client wrapper
│   │   ├── cache.go            # L1+L2 cache (memory + Redis)
│   │   ├── logger.go           # Structured JSON logger
│   │   ├── bottleneck.go       # Token-bucket concurrency limiter
│   │   ├── jobqueue.go         # Background job queue with priority, retries, scheduling
│   │   ├── migrations.go       # SQL migration runner
│   │   └── metrics.go          # Prometheus metrics
│   └── utils/
│       ├── errors.go            # Standardized error responses (400/401/403/404/409/429/500)
│       ├── crypto.go            # AES-256-GCM encryption/decryption
│       └── sanitize.go          # Input sanitization (XSS, SQL injection, email/phone validation)
├── migrations/
│   ├── 001_init.sql             # Core tables: users, conversations, messages, categories, qa_pairs, etc.
│   ├── 003_audit_logs.sql       # audit_logs table
│   ├── 004_audit_logs.sql       # Duplicate of 003
│   ├── 005_user_isolation.sql   # Multi-tenant: add user_id to categories, qa_pairs, unknown_questions
│   ├── 006_notifications_widget.sql  # notifications, widget_configs, user prefs
│   ├── 007_message_source.sql   # Add source column to messages
│   └── 008_inventory_leads.sql  # inventory_items, handoffs, owner_whatsapp
└── scripts/
    └── update_pagination.py      # Pagination helper script
```

### Frontend (`frontend/`)

```
frontend/
├── index.html                   # HTML entry: Inter font, favicons, <div id="root">
├── vite.config.ts               # Vite: React plugin, @/ alias, proxy /api+/ws→localhost:8080
├── tailwind.config.ts           # Custom colors (noant-*), Inter font, animations
├── tsconfig.json                # TypeScript config with @/ alias
├── package.json                 # Deps: react, react-router-dom, recharts, lucide-react, etc.
├── src/
│   ├── main.tsx                 # Bootstrap: initTheme, ErrorBoundary > NetworkProvider > Toast > App
│   ├── App.tsx                  # Router: createBrowserRouter with all auth + dashboard routes
│   ├── index.css                # Tailwind directives + CSS vars for light/dark themes + animations
│   ├── types/
│   │   └── index.ts             # All TypeScript interfaces: User, Conversation, Message, WSMessage, etc.
│   ├── lib/
│   │   ├── api.ts               # HTTP client: fetch wrapper, 401→refresh→redirect, retry, GET dedup
│   │   ├── auth.ts              # Auth service: login/signup/logout/getCurrentUser/refreshToken
│   │   ├── websocket.ts         # WebSocketManager singleton: connect, reconnect, subscribe
│   │   └── utils.ts             # cn(), timeAgo(), formatCurrency(), escapeHtml()
│   ├── hooks/
│   │   ├── useAPI.ts            # Central data-fetching: retry, dedup, loadMore, mergeData
│   │   ├── useAuth.ts           # Current user state: fetch on mount, signOut, refreshUser
│   │   ├── useWebSocket.ts      # WS singleton + per-component subscribe
│   │   ├── useInfiniteScroll.ts # IntersectionObserver load-more sentinel
│   │   ├── useKeyboardShortcuts.ts # Keyboard shortcut registration
│   │   ├── useModal.ts          # Boolean modal toggle
│   │   ├── useConfirm.ts        # Imperative confirmation dialog
│   │   └── useNetwork.ts        # Online/offline status
│   ├── contexts/
│   │   ├── NetworkContext.tsx    # navigator.onLine tracking
│   │   ├── WidgetConfigContext.tsx # Widget config fetch/save/sync-integrations
│   │   ├── SidebarAlertsContext.tsx # Polling + WS for unread counts, alerts
│   │   └── ModalContext.tsx     # Imperative showModal API
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── layout.tsx       # AuthLayout: brand panel + form
│   │   │   ├── login/page.tsx   # Login form
│   │   │   ├── signup/page.tsx  # Signup form
│   │   │   ├── forgot-password/page.tsx # Email input
│   │   │   └── reset-password/page.tsx  # Token + new password
│   │   └── (dashboard)/
│   │       ├── page.tsx         # Overview: stats + recent activity
│   │       ├── chats/page.tsx   # Chats: split-pane inbox, WebSocket, AI polling
│   │       ├── teach/page.tsx   # Teach: categories, Q&A, unknown Qs, CSV upload
│   │       ├── insights/page.tsx # Insights: charts + metrics
│   │       ├── channels/page.tsx # Channels: integration cards + connect modals
│   │       ├── setup/page.tsx   # Setup: profile/team/billing/api tabs
│   │       ├── settings/page.tsx # Settings: profile/security/notifications/appearance
│   │       ├── notifications/page.tsx # Notifications list
│   │       ├── billing/page.tsx # Billing: pricing plans
│   │       ├── team/page.tsx    # Team: member list + invite
│   │       ├── widget/page.tsx  # Widget: config editor + preview + embed code
│   │       ├── leads/page.tsx   # Leads: handoff pipeline (HOT/SOLD/LOST), status filters
│   │       └── inventory/page.tsx # Inventory: product/service/package CRUD, search, cards
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AuthLayout.tsx, DashboardLayout.tsx, Sidebar.tsx
│   │   │   ├── Header.tsx, BottomNav.tsx, MobileOverlay.tsx
│   │   ├── auth/
│   │   │   └── ProtectedRoute.tsx
│   │   ├── chat/
│   │   │   ├── ChatList.tsx, ChatMessages.tsx, ChatInput.tsx
│   │   │   ├── CustomerInfo.tsx, TypingIndicator.tsx, ConversationLoading.tsx
│   │   ├── channels/
│   │   │   ├── ChannelCard.tsx, ChannelIcon.tsx, TokenDisplay.tsx
│   │   │   ├── WhatsAppModal.tsx, TelegramModal.tsx
│   │   │   ├── InstagramModal.tsx, FacebookModal.tsx, WebWidgetModal.tsx
│   │   ├── stats/
│   │   │   ├── StatCard.tsx, StatGrid.tsx
│   │   │   ├── TrendChart.tsx, PeakHoursChart.tsx
│   │   │   ├── ChannelDistributionChart.tsx, MetricRow.tsx
│   │   ├── training/
│   │   │   ├── UploadZone.tsx, UnknownQuestion.tsx
│   │   │   ├── CategoryCard.tsx, QAPair.tsx
│   │   └── ui/
│   │       ├── Toast.tsx, Button.tsx, Input.tsx, Badge.tsx
│   │       ├── Card.tsx, Skeleton.tsx, Avatar.tsx, Spinner.tsx
│   │       ├── Modal.tsx, ConfirmModal.tsx
│   │       ├── CommandPalette.tsx, HelpModal.tsx
│   │       ├── ErrorBoundary.tsx, OfflineBanner.tsx
│   └── public/
│       ├── Logo A.png (favicon), Logo B.png, Logo&Name A.png
│       ├── favicon.svg, widget.js
│       └── (favicon references in index.html use Logo A.png)
```
