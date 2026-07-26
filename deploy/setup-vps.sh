#!/bin/bash
set -euo pipefail

# ─── NOANT VPS Setup Script ────────────────────────────────────
# Run this on a fresh Ubuntu/Debian VPS to deploy NOANT.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/.../setup-vps.sh | bash
#   OR: scp setup-vps.sh user@vps:~ && ssh user@vps 'bash setup-vps.sh'
# ────────────────────────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[noant]${NC} $1"; }
warn() { echo -e "${YELLOW}[warn]${NC} $1"; }
err() { echo -e "${RED}[error]${NC} $1"; exit 1; }

# ─── Step 1: System packages ────────────────────────────────────
log "Step 1/6: Installing system packages..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq curl git ufw fail2ban

# ─── Step 2: Docker ─────────────────────────────────────────────
log "Step 2/6: Installing Docker..."
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker
    usermod -aG docker $USER 2>/dev/null || true
fi

if ! command -v docker compose &>/dev/null; then
    # docker compose plugin should come with Docker
    log "Docker Compose plugin detected: $(docker compose version 2>&1)"
fi

# ─── Step 3: Clone repo ────────────────────────────────────────
log "Step 3/6: Cloning NOANT..."
DEPLOY_DIR="/opt/noant"
if [ -d "$DEPLOY_DIR" ]; then
    cd "$DEPLOY_DIR" && git pull
else
    git clone https://github.com/noant-enterprise/noant.git "$DEPLOY_DIR"
fi
cd "$DEPLOY_DIR/deploy"

# ─── Step 4: Configure environment ─────────────────────────────
log "Step 4/6: Configuring environment..."
if [ ! -f .env ]; then
    cp .env.production .env

    # Generate random secrets
    JWT_SECRET=$(openssl rand -hex 32)
    CSRF_SECRET=$(openssl rand -hex 32)
    REDIS_PASS=$(openssl rand -hex 16)

    sed -i "s/replace_with_random_64_char_string/$JWT_SECRET/" .env
    sed -i "s/replace_with_another_random_string/$CSRF_SECRET/" .env
    sed -i "s/changeme_redis_password_here/$REDIS_PASS/" .env

    echo ""
    warn "═══════════════════════════════════════════════════"
    warn "  .env file created. You MUST edit it with real values:"
    warn ""
    warn "  Required:"
    warn "    - TIDB_HOST/USER/PASSWORD (from TiDB Cloud)"
    warn "    - GROQ_API_KEY (from console.groq.com)"
    warn "    - DOMAIN (your domain name)"
    warn "    - LETSENCRYPT_EMAIL"
    warn ""
    warn "  Edit: $DEPLOY_DIR/deploy/.env"
    warn "═══════════════════════════════════════════════════"
    echo ""
else
    log ".env already exists, skipping"
fi

# ─── Step 5: Firewall ──────────────────────────────────────────
log "Step 5/6: Configuring firewall..."
ufw --force reset >/dev/null 2>&1
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw --force enable

# Fail2ban
systemctl enable --now fail2ban 2>/dev/null || true

# ─── Step 6: Build and start ───────────────────────────────────
log "Step 6/6: Building and starting NOANT..."
docker compose -f docker-compose.production.yml --env-file .env up -d --build

# Wait for health
log "Waiting for backend to be healthy..."
for i in $(seq 1 30); do
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        log "Backend is healthy!"
        break
    fi
    sleep 2
done

echo ""
echo "═══════════════════════════════════════════════════"
log "  NOANT is deployed!"
echo ""
log "  Frontend:  http://localhost (after DNS setup)"
log "  Admin:     http://localhost/admin"
log "  API:       http://localhost/api/v1/health"
log "  Backend:   http://localhost:8080/health"
echo ""
log "  Next steps:"
log "    1. Point your domain DNS to this VPS IP"
log "    2. Edit .env with TiDB Cloud + Groq credentials"
log "    3. Restart: cd /opt/noant/deploy && docker compose -f docker-compose.production.yml --env-file .env up -d"
log "    4. Register first user: curl -X POST http://localhost/api/v1/auth/register ..."
echo "═══════════════════════════════════════════════════"
