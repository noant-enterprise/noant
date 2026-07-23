# Changelog

All notable changes to NOANT Enterprise will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.0] - 2026-07-23

### Added
- Multi-tenancy organization scoping: `organizations` table, `org_id` columns on all data tables
- `getScopeID()` helper for org-aware handler scoping
- Migration 018: organizations table + org_id columns on core tables
- Migration 019: org_id columns on remaining tables (credits, subscriptions, templates, archives, API keys)
- Migration 020: delivery_status and external_id columns on messages
- Migration 021: backfill org_id for training tables
- Sentry error monitoring (Go + React) with DSN configuration
- Sentry middleware: `SentryContextMiddleware`, breadcrumbs, user context tags
- SentryGo SDK v0.48.0 integration with sentrygin middleware
- Sentry React SDK with browser tracing, session replay
- Sentry dashboard setup script and manual guide (27 widgets)
- `refreshAlerts()` in SidebarAlertsContext for immediate badge clearing
- OpenWA reliability: graceful shutdown, delivery tracking, session persistence, message dedup
- OpenWA efficiency: retry jitter, incoming webhook retry, priority aging, queue alerting
- OpenWA session health dashboard and session rotation after 5000 messages

### Fixed
- Multi-tenancy scoping bug: handlers used `getUserID(c)` but repos query by `org_id`
- All 25 handler files updated to use `getScopeID(c)` for org-scoped operations
- 10 service-layer struct creations missing `OrgID` field
- Team page redirect: `ListTeam` returns empty list instead of 401 when no org exists
- Alert badge not clearing: Header and notifications page now call `refreshAlerts()`
- AI response: removed hardcoded "next best option" append from trained answers
- WhatsApp campaign opt-out compliance
- AI reply routing via queue instead of synchronous
- Delivery status tracking for WhatsApp messages
- Rate limit backoff for OpenWA message sending
- Redis failure graceful degradation logging

### Changed
- Handler file count: 14 → 25 files
- Service file count: 13 → 49 files
- Repository file count: 19 → 27 files
- Migration count: 17 → 20 migrations

## [2.0.0] - 2026-07-21

### Added
- AI-powered customer support with Groq Llama 3.3 integration
- Multi-channel messaging: WhatsApp (OpenWA), Telegram, Web Widget
- Real-time streaming AI responses (SSE)
- Circuit breaker pattern for AI API resilience
- Parallel database and AI calls for faster response times
- Human-in-the-loop handoff system
- Campaign management with broadcast messaging
- Credit-based billing system with Polar.sh integration
- Team management with role-based access control
- API key management for third-party integrations
- Comprehensive analytics dashboard with 9+ report types
- Push notifications with VAPID/WebPush
- WhatsApp template management and submission
- Onboarding wizard with industry templates
- CSAT rating system
- Background job scheduler with Redis-backed queue
- 17 database migrations with auto-migration runner
- Comprehensive test suite: 742+ tests (unit, integration, E2E, chaos, benchmarks)
- CI/CD pipeline with 11 GitHub Actions jobs
- Security scanning: govulncheck, npm audit, CodeQL, SNYK
- Prometheus metrics with 15+ counters/histograms/gauges
- Grafana dashboard with 15 panels and 11 alert rules
- Prometheus recording rules for optimized queries
- Structured JSON logging with request ID tracking
- OWASP security headers, CSRF protection, rate limiting
- JWT authentication with httpOnly cookies and refresh tokens
- Database query tracing wrapper (TracedDB)
- API versioning middleware with deprecation headers
- Load testing framework (k6) with CI threshold gating
- Integration tests with testcontainers (TiDB + Redis)
- Chaos/fault injection tests for circuit breaker
- Performance benchmarks for critical utility functions
- OpenAPI 3.0.3 specification (130 endpoints)
- Incident response runbook with 8 scenarios
- Staging environment tuning guide
- Contributing guide and architecture documentation

### Changed
- Refactored monolithic files into 46 domain-specific files
- Split service.go (4488 lines) → 13 domain files
- Split handler.go (2717 lines) → 14 domain files
- Split repository.go (2395 lines) → 19 domain files
- Merged GenerateResponse + GenerateStreamingResponse into generateResponseCore
- Replaced magic numbers with named constants
- Pre-compiled regex patterns for performance
- Standardized error responses with ClassifyError middleware
- Auth refresh loop fix with hasFailedRefresh flag
- WebSocket reconnect with max 15 retries
- Redis nil pointer guards in PlanService and ChatService

### Security
- OWASP security headers on all responses
- CSRF protection via origin/Referer validation
- Rate limiting on all endpoints (IP-based and user-based)
- Input sanitization middleware
- Path traversal protection in static file serving
- Vulnerability disclosure policy (SECURITY.md)
