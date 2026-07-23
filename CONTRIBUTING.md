# Contributing to NOANT

Thank you for considering contributing to NOANT. This guide covers everything you need to get started.

## Prerequisites

- **Go** 1.25+
- **Node.js** 22+
- **TiDB** or **MySQL** 8.0+
- **Redis** 7+ (optional, falls back to in-memory cache)
- **Docker** & Docker Compose (recommended for local infra)

## Getting Started

```bash
# Clone the repo
git clone https://github.com/your-org/noant.git
cd noant

# Copy environment template
cp backend/.env.example backend/.env
# Edit .env with your credentials (Groq API key, Polar.sh, etc.)

# Start MySQL + Redis via Docker
docker compose up -d mysql redis

# Run the backend
cd backend
go run main.go

# In another terminal, run the frontend
cd frontend
npm install
npm run dev
```

Or use the single command to spin up everything:

```bash
docker compose up -d
make dev
```

## Project Structure

```
noant/
├── backend/              # Go backend (Gin framework)
│   ├── main.go           # Entrypoint, DI wiring, routes
│   ├── internal/
│   │   ├── handler/      # HTTP handlers (25 files, split by domain)
│   │   ├── service/      # Business logic layer (49 files)
│   │   ├── repository/   # Data access layer (27 files, MySQL + Redis)
│   │   ├── infrastructure/ # DB, Redis, logging, metrics, migrations
│   │   ├── middleware/    # Auth, CSRF, rate limiting, sanitization
│   │   ├── domain/       # Domain models
│   │   ├── errors/       # Custom error types
│   │   └── utils/        # Shared utilities
│   ├── migrations/       # SQL migrations
│   └── config/           # Configuration loaders
├── frontend/             # React + TypeScript (Vite)
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── app/          # Page-level views
│   │   ├── hooks/        # Custom React hooks
│   │   ├── contexts/     # React contexts
│   │   ├── lib/          # API clients, utilities
│   │   ├── types/        # TypeScript type definitions
│   │   └── test/         # Test utilities
│   └── e2e/              # Playwright end-to-end tests
├── monitoring/           # Prometheus/Grafana config
├── OpenWA/               # OpenWA WhatsApp integration
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Development Workflow

1. **Branch from `main`**: `git checkout -b feat/your-feature`
2. **Make your changes** in the appropriate layer (handler → service → repository)
3. **Write or update tests** for your changes
4. **Run the linter and tests** before committing
5. **Create a PR** against `main`

## Code Style

### Backend (Go)

- **Linter**: `golangci-lint` with the project config (`.golangci.yaml`)
- **Formatter**: `gofmt` (or `goimports`) — no manual formatting needed
- **Lint before pushing**:
  ```bash
  make lint-backend
  # or
  cd backend && golangci-lint run ./...
  ```
- Do not add comments to code unless explicitly asked
- Follow idiomatic Go conventions: short variable names, early returns, table-driven tests
- Use the existing error handling patterns in `internal/errors/`

### Frontend (TypeScript/React)

- **Linter**: ESLint — run with `npm run lint` from `frontend/`
- **Formatter**: Prettier — run with `npm run format` from `frontend/`
- **TypeScript**: strict mode is enabled; do not use `any` unless absolutely necessary
- **Lint before pushing**:
  ```bash
  make lint-frontend
  # or
  cd frontend && npm run lint
  ```
- Prefer functional components with hooks
- Use existing component patterns from `src/components/`

## Testing

### Backend

```bash
cd backend
go test -v -race ./...

# or via Makefile from project root
make test-backend
```

### Frontend

```bash
cd frontend
npm test           # single run
npm run test:watch # watch mode

# or via Makefile from project root
make test-frontend
```

### End-to-End

```bash
cd frontend
npx playwright install   # first time only
npx playwright test

# or via Makefile from project root
make e2e
```

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

| Prefix | Use for |
|--------|---------|
| `feat:` | New feature |
| `fix:` | Bug fix |
| `chore:` | Maintenance, deps, tooling |
| `docs:` | Documentation only |
| `refactor:` | Code restructuring without behavior change |
| `test:` | Adding or updating tests |

Examples:
```
feat: add Telegram webhook handler
fix: resolve race condition in credit deduction
chore: bump Go to 1.25
docs: update architecture diagram
```

## Pull Request Guidelines

- **Keep PRs small and focused** — one feature or fix per PR
- **Describe your changes** clearly in the PR description
- **Link related issues** (e.g., `Closes #42`)
- **Ensure CI passes** — lint, tests, and build must all be green
- **Request review** from at least one maintainer before merging
- **Squash merge** into `main` to keep history clean

## Environment Variables

All configuration is driven by environment variables. See `.env.example` for the full list with defaults.

Key variables:
- `GROQ_API_KEY` — AI inference API key
- `POLAR_ACCESS_TOKEN` — Payment processing
- `TIDB_HOST`, `TIDB_PORT`, `TIDB_USER`, `TIDB_PASSWORD`, `TIDB_DATABASE` — Database
- `REDIS_URL` — Cache (optional)
- `JWT_SECRET` — Authentication signing key
- `SENTRY_DSN` — Sentry error monitoring DSN (optional, set to empty to disable)
- `OPENWA_*` — WhatsApp integration settings

## Architecture Decisions

For detailed architectural rationale — data flow, security model, observability, and background jobs — see [ARCHITECTURE.md](./ARCHITECTURE.md).
