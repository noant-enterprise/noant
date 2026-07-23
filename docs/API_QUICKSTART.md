# NOANT API Quickstart Guide

> **Base URL:** `http://localhost:8080/api/v1`

All requests and responses use `application/json` unless otherwise noted. Authentication uses JWT tokens delivered via `httpOnly` cookies or the `Authorization: Bearer` header.

---

## Table of Contents

- [Authentication](#authentication)
- [Making Authenticated Requests](#making-authenticated-requests)
- [Conversations & Messages](#conversations--messages)
- [AI Chat (Streaming)](#ai-chat-streaming)
- [Training (QA Pairs)](#training-qa-pairs)
- [Channels (WhatsApp)](#channels-whatsapp)
- [Templates](#templates)
- [Inventory](#inventory)
- [Handoffs](#handoffs)
- [Analytics](#analytics)
- [Credits](#credits)
- [Notifications](#notifications)
- [Campaigns](#campaigns)
- [Team](#team)
- [Widget](#widget)
- [Error Codes](#error-codes)
- [Rate Limits](#rate-limits)

---

## Authentication

### Register

Create a new account. A verification email is sent automatically.

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "you@example.com",
    "password": "securepass123",
    "first_name": "John",
    "last_name": "Doe",
    "company_name": "Acme Inc"
  }'
```

**Response `201 Created`:**
```json
{
  "message": "User registered successfully",
  "user": { "id": "...", "email": "you@example.com", ... }
}
```

### Verify Email

Use the 6-digit code from the verification email.

```bash
curl -X POST http://localhost:8080/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "you@example.com",
    "code": "482910"
  }'
```

**Response `200 OK`:**
```json
{
  "message": "Email verified successfully",
  "user": { ... },
  "trial_info": { "trial_expires_at": "...", "trial_days_left": 14 }
}
```

Sets `noant_access` and `noant_refresh` httpOnly cookies on success.

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "you@example.com",
    "password": "securepass123"
  }'
```

**Response `200 OK`:**
```json
{
  "user": { "id": "...", "email": "you@example.com", "plan_id": "free", ... },
  "trial_info": { "trial_ended": false, "trial_days_left": 14 }
}
```

Sets `noant_access` (24h) and `noant_refresh` (7d) httpOnly cookies.

**Errors:**
- `403` — `email_not_verified`
- `429` — Account temporarily locked (5 failed attempts, 15 min cooldown)
- `401` — Invalid credentials

### Refresh Token

Exchange the refresh cookie for a new access token.

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -b "noant_refresh=<refresh_token>"
```

**Response `200 OK`:**
```json
{ "message": "Session refreshed" }
```

### Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -b "noant_access=<token>; noant_refresh=<refresh_token>"
```

---

## Making Authenticated Requests

Use either method:

**Option 1 — Cookies (browser / same-origin):**
```bash
curl http://localhost:8080/api/v1/auth/me \
  -b "noant_access=<token>"
```

**Option 2 — Bearer token header:**
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

### Get Current User

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**
```json
{
  "user": {
    "id": "...",
    "email": "you@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "owner",
    "plan_id": "free",
    "credit_balance": 50,
    "is_verified": true,
    "onboarding_status": "complete"
  },
  "trial_info": { "trial_ended": false, "trial_days_left": 14 }
}
```

---

## Conversations & Messages

### List Conversations

```bash
curl "http://localhost:8080/api/v1/chats/conversations?page=1&limit=20&status=active" \
  -H "Authorization: Bearer <token>"
```

**Query Parameters:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | Page number |
| limit | int | 20 | Items per page (max 100) |
| status | string | — | Filter: `active`, `resolved`, `escalated`, `archived` |

**Response `200 OK`:**
```json
{
  "conversations": [
    {
      "id": "...",
      "customer_name": "Jane",
      "channel": "whatsapp",
      "status": "active",
      "intent": "buying",
      "priority": "medium",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "limit": 20,
  "has_more": true
}
```

### Get Conversation with Messages

```bash
curl "http://localhost:8080/api/v1/chats/conversations/<id>?page=1&limit=30" \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**
```json
{
  "conversation": { "id": "...", "status": "active", ... },
  "messages": [
    {
      "id": "...",
      "sender_type": "customer",
      "content": "Do you have this in blue?",
      "sequence": 1,
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 12,
  "has_more": true,
  "page": 1
}
```

### Send Message (Non-streaming)

Sends a message and triggers an asynchronous AI response (delivered via WebSocket).

```bash
curl -X POST http://localhost:8080/api/v1/chats/conversations/<id>/messages \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, how can I help?"}'
```

**Response `200 OK`:**
```json
{ "message": "Message sent" }
```

### Direct Chat (One-off)

Send a single message and get an AI response without creating a persistent conversation.

```bash
curl -X POST http://localhost:8080/api/v1/chats/direct-chat \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Test User",
    "message": "What are your business hours?",
    "channel": "web"
  }'
```

**Response `200 OK`:**
```json
{
  "conversation": { "id": "...", ... },
  "message": { "id": "...", "content": "We are open Mon-Fri 9am-5pm...", ... }
}
```

### Escalate Conversation

```bash
curl -X POST http://localhost:8080/api/v1/chats/conversations/<id>/escalate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Customer requests human agent"}'
```

### Human Takeover

Take over an AI-managed conversation as a human agent.

```bash
curl -X PUT http://localhost:8080/api/v1/chats/conversations/<id>/takeover \
  -H "Authorization: Bearer <token>"
```

### Rate Conversation (CSAT)

```bash
curl -X POST http://localhost:8080/api/v1/chats/conversations/<id>/rate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"score": 5, "feedback": "Great help!"}'
```

`score` must be between 1 and 5.

---

## AI Chat (Streaming)

### Stream Message via SSE

Send a message and receive the AI response token-by-token using Server-Sent Events.

```bash
curl -X POST http://localhost:8080/api/v1/chats/conversations/<id>/stream \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"content": "Tell me about your pricing plans"}' \
  --no-buffer
```

**Response (SSE stream):**

```
data: We offer three pricing plans:

data:  The Starter plan

data:  is free and includes

data:  basic AI responses...

data: [DONE]{"id":"msg_abc123","created_at":"2025-01-15T10:31:00Z","confidence":0.95,"source":"qa_match"}
```

**SSE Event Format:**

| Event | Format | Description |
|---|---|---|
| Token | `data: <text>` | Incremental response chunk |
| Error | `data: [ERROR]` | Generation failed |
| Done | `data: [DONE]<json>` | Stream complete; JSON metadata follows |

The `[DONE]` event includes a JSON object with `id`, `created_at`, `confidence`, and `source` fields.

### Assistant Chat (Onboarding)

A separate conversational assistant for platform setup guidance.

```bash
curl -X POST http://localhost:8080/api/v1/assistant/chat \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"message": "How do I set up my first QA pair?"}'
```

**Response `200 OK`:**
```json
{
  "content": "To set up your first QA pair...",
  "action": "guide",
  "steps": ["Go to Training", "Click Create QA Pair", "..."],
  "suggestions": ["Upload CSV", "Create Category"]
}
```

---

## Training (QA Pairs)

### Create a Category

```bash
curl -X POST http://localhost:8080/api/v1/training/categories \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Pricing", "description": "Price-related questions", "color": "#10b981"}'
```

### List Categories

```bash
curl http://localhost:8080/api/v1/training/categories \
  -H "Authorization: Bearer <token>"
```

### Create a QA Pair

```bash
curl -X POST http://localhost:8080/api/v1/training/qa \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "<category_uuid>",
    "question": "What is your return policy?",
    "answer": "We offer 30-day returns on all items.",
    "variations": ["Can I return this?", "How do I return a product?"]
  }'
```

### List Unknown Questions

View questions from customers that were not matched to any QA pair.

```bash
curl http://localhost:8080/api/v1/training/unknown-questions \
  -H "Authorization: Bearer <token>"
```

### Train an Unknown Question

Convert an unknown question into a new QA pair.

```bash
curl -X POST http://localhost:8080/api/v1/training/unknown-questions/<id>/train \
  -H "Authorization: Bearer <token>"
```

### Bulk Import QA Pairs

```bash
curl -X POST http://localhost:8080/api/v1/training/bulk-qa \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "<category_uuid>",
    "pairs": [
      {"question": "What are your hours?", "answer": "Mon-Fri 9am-5pm."},
      {"question": "Do you ship internationally?", "answer": "Yes, to 50+ countries."}
    ]
  }'
```

---

## Channels (WhatsApp)

### Get Integration Status

```bash
curl http://localhost:8080/api/v1/channels/whatsapp/status \
  -H "Authorization: Bearer <token>"
```

### Get Session Health

```bash
curl http://localhost:8080/api/v1/channels/whatsapp/sessions/health \
  -H "Authorization: Bearer <token>"
```

---

## Templates

### List Templates

```bash
curl http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer <token>"
```

### Create Template

```bash
curl -X POST http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "order_confirmation",
    "language": "en",
    "category": "utility",
    "body_text": "Your order {{1}} has been confirmed."
  }'
```

---

## Inventory

### List Products

```bash
curl http://localhost:8080/api/v1/inventory \
  -H "Authorization: Bearer <token>"
```

### Create Product

```bash
curl -X POST http://localhost:8080/api/v1/inventory \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Widget",
    "type": "product",
    "price": 25000,
    "description": "High-quality widget",
    "stock_quantity": 100
  }'
```

---

## Handoffs

### List Handoffs

```bash
curl http://localhost:8080/api/v1/handoffs \
  -H "Authorization: Bearer <token>"
```

### Update Handoff Status

```bash
curl -X PUT http://localhost:8080/api/v1/handoffs/<id> \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status": "sold", "final_price": 22000}'
```

---

## Analytics

### Dashboard Overview

```bash
curl http://localhost:8080/api/v1/analytics/overview \
  -H "Authorization: Bearer <token>"
```

### Unknown Questions Stats

```bash
curl http://localhost:8080/api/v1/analytics/unknown-questions \
  -H "Authorization: Bearer <token>"
```

---

## Credits

### Get Balance

```bash
curl http://localhost:8080/api/v1/credits/balance \
  -H "Authorization: Bearer <token>"
```

---

## Notifications

### List Notifications

```bash
curl "http://localhost:8080/api/v1/notifications?limit=20" \
  -H "Authorization: Bearer <token>"
```

### Mark All Read

```bash
curl -X POST http://localhost:8080/api/v1/notifications/read-all \
  -H "Authorization: Bearer <token>"
```

---

## Campaigns

### List Campaigns

```bash
curl http://localhost:8080/api/v1/campaigns \
  -H "Authorization: Bearer <token>"
```

### Create Campaign

```bash
curl -X POST http://localhost:8080/api/v1/campaigns \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Summer Sale",
    "start_date": "2026-08-01",
    "end_date": "2026-08-31"
  }'
```

---

## Team

### List Team Members

```bash
curl http://localhost:8080/api/v1/settings/team \
  -H "Authorization: Bearer <token>"
```

### Invite Member

```bash
curl -X POST http://localhost:8080/api/v1/settings/team/invite \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"email": "colleague@company.com", "role": "member"}'
```

---

## Widget

### Get Widget Config

```bash
curl http://localhost:8080/api/v1/widget \
  -H "Authorization: Bearer <token>"
```

### Public Chat (Embed)

```bash
curl -X POST http://localhost:8080/api/v1/widget/chat \
  -H "Content-Type: application/json" \
  -d '{"api_key": "widget_xxx", "message": "Hello!"}'
```

---

## Error Codes

| HTTP Status | Error | Meaning |
|---|---|---|
| 400 | `validation_error` | Request body failed validation |
| 400 | `invalid_code` | Email verification code is wrong |
| 401 | `authorization required` | No authentication token provided |
| 401 | `invalid or expired token` | JWT token is invalid or expired |
| 401 | `token revoked` | Token was logged out and blacklisted |
| 401 | `Invalid email or password` | Login credentials incorrect |
| 403 | `email_not_verified` | Account email not yet verified |
| 403 | `Admin access required` | Endpoint requires owner/admin role |
| 404 | `Conversation not found` | Conversation ID does not exist or not owned by user |
| 409 | `Registration failed` | Email already registered |
| 413 | — | Request body exceeds size limit (8MB) |
| 429 | `rate limit exceeded` | Too many requests for this endpoint group |
| 429 | `Account temporarily locked` | Too many failed logins (15 min lockout) |
| 429 | `too_many_attempts` | Email verification rate limited |
| 500 | Internal error | Server-side failure; check request_id header |

Every response includes an `X-Request-ID` header for tracing.

---

## Rate Limits

Rate limits are applied per endpoint group. Responses include `X-RateLimit-Limit` and `X-RateLimit-Remaining` headers. When exceeded, a `429` is returned with a `retry_after` field (seconds).

| Endpoint Group | Path Prefix | Limit | Window |
|---|---|---|---|
| Auth mutations | `/api/v1/auth/register`, `/login`, `/logout`, `/change-password`, `/forgot-password`, `/reset-password`, `/verify`, `/resend-verification` | 10 req/min | Per IP |
| Auth sessions | `/api/v1/auth/refresh`, `/me` | 120 req/min | Per IP |
| Chats | `/api/v1/chats/**` | 500 req/min | Per user |
| Training | `/api/v1/training/**` | 300 req/min | Per user |
| Analytics | `/api/v1/analytics/**` | 60 req/min | Per user |
| Integrations | `/api/v1/integrations/**` | 300 req/min | Per user |
| Settings | `/api/v1/settings/**` | 60 req/min | Per user |
| Credits | `/api/v1/credits/**` | 30 req/min | Per user |
| Campaigns | `/api/v1/campaigns/**` | 30 req/min | Per user |
| Templates | `/api/v1/templates/**` | 30 req/min | Per user |
| WhatsApp Admin | `/api/v1/whatsapp/admin/**` | default | — |
| Inventory | `/api/v1/inventory/**` | 60 req/min | Per user |
| Handoffs | `/api/v1/handoffs/**` | 60 req/min | Per user |
| Onboarding | `/api/v1/onboarding/**` | 30 req/min | Per user |
| Assistant | `/api/v1/assistant/**` | 30 req/min | Per user |
| Push | `/api/v1/push/**` | 30 req/min | Per user |
| Webhooks | `/api/v1/openwa/webhook`, `/api/v1/telegram/webhook` | 1000 req/min | Per IP |

---

## WebSocket

Connect to `ws://localhost:8080/ws` with a valid JWT token (via cookie or query param). The server pushes real-time events for conversation updates.

**Message types:**

```json
{
  "type": "new_message",
  "conversation_id": "...",
  "data": {
    "id": "...",
    "content": "...",
    "role": "ai",
    "confidence": 0.95,
    "source": "qa_match"
  }
}
```

```json
{
  "type": "typing_indicator",
  "conversation_id": "...",
  "data": { "is_typing": true }
}
```

---

## Health Check

```bash
curl http://localhost:8080/health
```

```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T10:30:00Z",
  "version": "2.0.0"
}
```
