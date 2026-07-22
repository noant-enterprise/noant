# NOANT — Sentry Dashboard Setup Guide

## Quick Setup (Automated)

```bash
export SENTRY_AUTH_TOKEN=your_auth_token_here
export SENTRY_ORG=divineshedrack33220
pip install requests
python3 scripts/sentry_dashboard_setup.py
```

## Manual Setup (Step by Step)

Go to: https://sentry.io/organizations/divineshedrack33220/dashboards/

Click **"Create Dashboard"** → Name: **"NOANT — Production Dashboard"**

---

## Dashboard Layout (9 Rows, 27 Widgets)

### Row 1: Top-Level KPIs (Big Number)

| Widget | Query | Conditions |
|--------|-------|------------|
| Total Errors (24h) | `count()` | `level:error` |
| Affected Users (24h) | `count_unique(user)` | `level:error` |
| Error Rate % | `failure_rate()` | *(none)* |
| P95 Latency | `p95(transaction.duration)` | *(none)* |

### Row 2: Error Trends (Line/Stacked/TopN)

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| Errors Over Time | Line | `count()` | `level:error` |
| Errors by Endpoint | Top 10 | `count()` | `level:error` |
| New vs Recurring | Stacked Area | `count()` | `issue.type:regression` / negative |

### Row 3: Backend Health (Go/Gin)

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| Errors by Status Code | Top 10 | `count()` | `level:error` |
| API Request Duration | Line (P50/P95/P99) | `p50/p95/p99(transaction.duration)` | *(none)* |
| Database Errors | Big Number | `count()` | `level:error (database OR sql OR tidb)` |

### Row 4: AI/LLM Monitoring (Groq)

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| AI Response Failures | Big Number | `count()` | `level:error (groq OR ai OR llama)` |
| AI Response Latency | Line | `p95(transaction.duration)` | `transaction:/api/v1/chats/direct-chat` |
| Groq Rate Limit Events | Big Number | `count()` | `429 OR rate_limit` |

### Row 5: WhatsApp/OpenWA

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| WhatsApp Errors | Big Number | `count()` | `level:error (openwa OR whatsapp OR webhook)` |
| Webhook Failures | Line | `count()` | `level:error (openwa OR whatsapp) (webhook OR delivery)` |
| Session Reconnections | Line | `count()` | `reconnect OR session OR disconnected` |

### Row 6: Security

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| Auth Failures | Big Number | `count()` | `level:error (unauthorized OR invalid_credentials)` |
| Rate Limiting Events | Line | `count()` | `429 OR rate_limit OR throttled` |
| Security Errors by User | Top 10 | `count_unique(user)` | `level:error (unauthorized OR forbidden)` |

### Row 7: Frontend (React/TypeScript)

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| Frontend JS Errors | Big Number | `count()` | `level:error (TypeError OR ReferenceError)` |
| Frontend Errors Trend | Line | `count()` | `level:error platform:javascript` |
| Most Common JS Errors | Top 10 | `count()` | `level:error platform:javascript` |

### Row 8: Releases & Deploy

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| Errors by Release | Top 5 | `count()` | `level:error` |
| Error-Free Sessions | Big Number | `count_unique(session)` | `level:error` |

### Row 9: Infrastructure

| Widget | Type | Query | Conditions |
|--------|------|-------|------------|
| Redis Errors | Big Number | `count()` | `level:error (redis OR connection refused)` |
| HTTP 5xx Errors | Line | `count()` | `level:error (500 OR 502 OR 503 OR 504)` |
| Unhandled Exceptions | Line | `count()` | `level:error (panic OR goroutine)` |

---

## Alert Rules (set up in Sentry → Alerts)

| Alert | Condition | Severity | Notify |
|-------|-----------|----------|--------|
| High Error Rate | >5% errors for 5min | Critical | Email + Slack |
| AI Failure Spike | >10 AI failures/min | Critical | Email |
| WhatsApp Session Down | Session disconnect >10min | Warning | Email |
| P99 Latency Spike | P99 >10s for 5min | Critical | Email |
| Database Exhaustion | DB errors >5/min | Critical | Email |
| Frontend JS Crash | >20 JS errors/hour | Warning | Email |
| Auth Brute Force | >50 auth failures/hour | Warning | Email |
| Redis Unavailable | Redis errors >1/min | Critical | Email |

---

## Issue Auto-Assignment Rules

Set up in Sentry → Settings → Ownership Rules:

| Pattern | Assign To |
|---------|-----------|
| `path:internal/service/*` | @backend-team |
| `path:internal/handler/*` | @backend-team |
| `path:internal/middleware/*` | @backend-team |
| `message:*groq*` | @ai-team |
| `message:*openwa*` | @whatsapp-team |
| `platform:javascript` | @frontend-team |
| `message:*redis*` | @infrastructure |
| `message:*database*` | @infrastructure |
