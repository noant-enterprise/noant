.PHONY: test test-backend test-frontend lint lint-backend lint-frontend build build-backend build-frontend docker-build docker-up docker-down e2e

# ── Tests ──────────────────────────────────────────────────────────────

test: test-backend test-frontend

test-backend:
	cd backend && go test -v -race ./...

test-frontend:
	cd frontend && npx vitest run

# ── Lint ───────────────────────────────────────────────────────────────

lint: lint-backend lint-frontend

lint-backend:
	cd backend && golangci-lint run ./...

lint-frontend:
	cd frontend && npx eslint src --ext .ts,.tsx

# ── Build ──────────────────────────────────────────────────────────────

build: build-backend build-frontend

build-backend:
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/main main.go

build-frontend:
	cd frontend && npm run build

# ── Docker ─────────────────────────────────────────────────────────────

docker-build:
	docker build -t noant:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ── E2E (Playwright) ──────────────────────────────────────────────────
# Requires: npm install -D @playwright/test (in frontend/) and npx playwright install

e2e:
	cd frontend && npx playwright test
