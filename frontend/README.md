# NOANT Frontend (Customer App)

Main customer-facing web application. Users manage their AI customer support — connect WhatsApp, view conversations, upload knowledge base, track analytics, manage billing.

## Tech Stack

- React 18 + TypeScript (strict mode, `noUncheckedIndexedAccess`)
- Vite 6 (dev server on port 5173)
- React Router 6 (SPA routing)
- Tailwind CSS (dark theme)
- Recharts (charts)
- Lucide React (icons)
- Vitest + React Testing Library (463 tests)

## Routes

### Public
| Route | Page |
|-------|------|
| `/` | Landing page (marketing, pricing in Naira, features, FAQ) |
| `/login` | Login (email + password) |
| `/signup` | Registration |
| `/forgot-password` | Password reset request |
| `/reset-password` | Password reset form |
| `/verify-email` | Email verification |
| `/invite/:code` | Referral invite page (public, no auth) |
| `/terms` | Terms of Service |
| `/privacy` | Privacy Policy |

### Protected (requires login)
| Route | Page |
|-------|------|
| `/dashboard` | Main dashboard (stats, quick actions) |
| `/chats` | Live conversation list + chat view |
| `/channels` | Channel management (WhatsApp, Telegram, web widget) |
| `/knowledge` | Knowledge base (upload docs, manage training data) |
| `/insights` | Analytics (response time, resolution rate, top questions) |
| `/billing` | Subscription management (Free / Pro / Enterprise) |
| `/onboarding` | Setup wizard (3 steps) |
| `/settings` | Account settings, notifications, API keys |

## Project Structure

```
frontend/src/
├── app/                # Page components organized by route group
│   ├── landing/        # Marketing landing page
│   ├── (auth)/         # Login, signup, forgot/reset password, verify
│   ├── (dashboard)/    # Protected app pages (chats, channels, etc.)
│   ├── invite/[code]/  # Public referral invite
│   └── legal/          # Terms, Privacy pages
├── components/
│   ├── chat/           # ChatWindow, ChatList, MessageBubble
│   ├── layout/         # AppShell (sidebar + topbar + content)
│   └── ui/             # Toast, HelpModal, LoadingSpinner
├── lib/
│   ├── api.ts          # API client (fetch wrapper with cookie auth)
│   ├── utils.ts        # Formatters, helpers
│   └── hooks/          # Data hooks (useChats, useChannels, etc.)
├── contexts/           # React contexts (Auth, Toast, Theme)
└── types/              # TypeScript interfaces
```

## Setup

```bash
cd frontend
npm install
npm run dev        # → http://localhost:5173
```

The dev server proxies `/api` to `http://localhost:8080` (backend).

## Scripts

| Command | Purpose |
|---------|---------|
| `npm run dev` | Start dev server on port 5173 |
| `npm run build` | TypeScript check + production build |
| `npm run preview` | Preview production build |
| `npx tsc -b` | Type-check only |
| `npx vitest run` | Unit tests (463 tests, 52 files) |
| `npx playwright test` | E2E tests (Playwright) |

## E2E Tests

```bash
# Install browsers (first time only)
npx playwright install

# Run all E2E tests
npx playwright test

# Run specific test file
npx playwright test e2e/auth.spec.ts

# Run in headed mode (see browser)
npx playwright test --headed
```

E2E tests require the backend to be running (`go run .` in `backend/`).

## Design System

- **Primary:** `#0ea5e9` (sky blue)
- **Background:** `#0a0a0a` (noant-black)
- **Surface:** `#fafafa` (noant-paper)
- **CSS Vars:** `--bg-base`, `--bg-surface`, `--text-primary`, `--brand-sky`
- **Components:** rounded-xl borders, bg-bg-surface cards, text-text-primary text

## Environment Variables

| File | Purpose |
|------|---------|
| `.env.development` | Dev API URL (`http://localhost:8080`) |
| `.env.production` | Prod API URL, VAPID key, Sentry DSN |

Set `VITE_API_URL` to override the default API base URL.

## Key Patterns

- **Cookie auth** — all API calls use `credentials: 'include'` (httpOnly cookies, no bearer tokens)
- **Polling with retry** — chat list polls for new messages with max 30 retries
- **Optimistic UI** — status updates reflect immediately, errors revert
- **Null-safe access** — all optional fields use `??` fallbacks (TypeScript `noUncheckedIndexedAccess`)
