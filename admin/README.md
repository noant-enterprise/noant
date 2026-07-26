# NOANT Admin Panel (CEO Command Center)

Internal admin dashboard for monitoring and managing the NOANT platform.

## Tech Stack

- React 18 + TypeScript (strict mode)
- Vite 6 (dev server on port 3002)
- Tailwind CSS (dark theme)
- Recharts (charts)
- Lucide React (icons)

## Pages

| Route | Purpose |
|-------|---------|
| `/` | Dashboard — live overview (MRR, users, visitors, system status) |
| `/pipeline` | Sales CRM — field meeting leads, status pipeline, QR code sharing |
| `/customers` | Customer list with search, plan filtering, detail view |
| `/analytics` | Landing page analytics (visitors, funnel, sources, bounce rate) |
| `/revenue` | Revenue tracking (MRR, churn, failed payments, billing table) |
| `/ai` | AI health (accuracy, response time, unknown questions, sentiment) |
| `/knowledge` | Knowledge base manager (upload docs, train AI answers) |
| `/system` | System health (API, TiDB, Redis, WebSocket latency) |
| `/audit-logs` | Audit trail with search, action filters, user details |
| `/settings` | Profile, team, API keys |
| `/login` | Admin login (email + password, role-gated) |

## Realtime

- **WebSocket** — connects to `ws://backend/api/v1/admin/ws` for instant updates (lead_created, lead_updated, user_signed_up)
- **Auto-refresh polling** — fallback when WS disconnects (15–30s intervals per page)
- **useAdminWS hook** — typed WebSocket event listener with auto-reconnect
- **useAutoRefresh hook** — WS + interval combo for any data refetch

## Project Structure

```
admin/src/
├── app/              # Page components (one per route)
├── components/
│   ├── data/         # StatCard, LiveFeed, AlertBanner
│   ├── layout/       # Shell, Sidebar, Topbar, CommandPalette
│   └── ui/           # Skeleton, Feedback (EmptyState, ErrorBanner)
├── lib/
│   ├── api.ts        # Admin API client (all endpoints)
│   ├── utils.ts      # Formatters, timeAgo
│   └── hooks/        # Data hooks (useAnalytics, useRevenue, etc.)
├── types/            # TypeScript interfaces
└── App.tsx           # Router with protected routes
```

## Setup

```bash
cd admin
npm install
npm run dev        # → http://localhost:3002
```

The dev server proxies `/api` to `http://localhost:8080` (backend).

## Scripts

| Command | Purpose |
|---------|---------|
| `npm run dev` | Start dev server on port 3002 |
| `npm run build` | TypeScript check + production build |
| `npm run preview` | Preview production build |
| `npx tsc -b` | Type-check only |
| `npx vitest run` | Unit tests |

## Keyboard Shortcuts

- **Ctrl+K** — Open command palette (search all pages)
