# NOANT Deployment Guide

Deployment options for the NOANT platform.

## Quick Start (Local Development)

```bash
# 1. Start backend (requires Go 1.25 + TiDB)
cd backend && go run .

# 2. Start admin panel (port 3002)
cd admin && npm install && npm run dev

# 3. Start frontend (port 5173)
cd frontend && npm install && npm run dev
```

## Production Deployment Options

### Option A: VPS Self-Hosted (Recommended)

Full control, lowest cost. Works with any VPS provider (DigitalOcean, Hetzner, AWS EC2).

**Prerequisites:**
- Ubuntu 22.04+ VPS with 2GB+ RAM
- Domain name pointed at your VPS IP
- TiDB Cloud free tier account (or self-hosted TiDB)

**Steps:**

```bash
# 1. SSH into your VPS
ssh root@your-vps-ip

# 2. Run setup script (installs Docker, configures firewall)
git clone https://github.com/divineshedrack33220/noant.git
cd noant/deploy
chmod +x setup-vps.sh && ./setup-vps.sh

# 3. Configure environment
cp .env.example .env
nano .env  # Fill in your values

# 4. Deploy
chmod +x deploy.sh && ./deploy.sh deploy
```

**What gets deployed:**
- Backend (Go API) on port 8080
- Frontend (React) on port 3001
- Admin panel on port 3002
- Nginx reverse proxy on ports 80/443
- Certbot for SSL certificates

**Blue-Green Deployment:**

```bash
./deploy.sh deploy    # Deploy new version (zero downtime)
./deploy.sh rollback  # Rollback to previous version
./deploy.sh status    # Check current deployment status
./deploy.sh build     # Rebuild Docker images
```

### Option B: Render (Managed Cloud)

One-click deployment with Render Blueprint.

1. Push to GitHub
2. Connect repo to Render
3. Render auto-detects `render.yaml` Blueprint
4. Set environment variables in Render dashboard
5. Deploy

### Option C: Docker Compose (Local/Any Cloud)

```bash
cd deploy

# Production stack (external TiDB + Redis)
docker compose -f docker-compose.production.yml up -d

# Or blue-green (local TiDB + Redis)
docker compose -f docker-compose.infra.yml up -d
docker compose -f docker-compose.app.yml up -d
```

## Architecture

```
Internet → Nginx (443) → Backend (8080) + Frontend (3001) + Admin (3002)
                                  ↓
                           TiDB Cloud (external)
                                  ↓
                            Redis (local Docker)
```

## SSL Certificates

Certbot auto-renews certificates. For manual setup:

```bash
certbot --nginx -d noant.yourdomain.com
certbot --nginx -d api.noant.yourdomain.com
```

## Environment Variables

See `deploy/.env.example` for the complete template with all variables:

| Category | Variables |
|----------|-----------|
| Database | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` |
| Redis | `REDIS_URL` |
| Auth | `JWT_SECRET`, `SESSION_SECRET` |
| AI | `GROQ_API_KEY` |
| WhatsApp | `OPENWA_*` |
| Payments | `POLAR_*` |
| Monitoring | `SENTRY_DNS`, `SENTRY_AUTH_TOKEN` |
| Domain | `CORS_ORIGINS`, `FRONTEND_URL`, `ADMIN_URL` |

## Monitoring

See `monitoring/` directory for Grafana dashboards and Prometheus configs.

- Grafana: `http://your-vps:3000` (default port, change in production)
- Prometheus: `http://your-vps:9090`
- Sentry: `https://sentry.io`

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Migration fails | Ensure `multiStatements=true` in TiDB DSN |
| WebSocket disconnects | Check CORS_ORIGINS includes your domain |
| 500 on auth endpoints | Check JWT_SECRET is set and consistent |
| Redis unavailable | App runs in offline mode (in-memory fallback) |
| WhatsApp not connecting | Check OpenWA session status at `/api/v1/whatsapp/status` |
