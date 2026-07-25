#!/bin/bash
set -euo pipefail

# ─── NOANT Blue-Green Deploy Script ────────────────────────────
# Deploys new code with zero downtime by switching between
# blue and green app stacks.
#
# Usage:
#   ./deploy.sh              # Deploy to inactive color
#   ./deploy.sh rollback     # Rollback to previous color
#   ./deploy.sh status       # Show which color is live
#   ./deploy.sh build        # Build both colors without switching
# ────────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
STATE_FILE="$SCRIPT_DIR/.active-color"
LOG_FILE="$SCRIPT_DIR/deploy.log"

# Port mapping: blue=8081/3001, green=8082/3002
declare -A APP_PORTS=( [blue]=8081 [green]=8082 )
declare -A FE_PORTS=( [blue]=3001 [green]=3002 )

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[deploy]${NC} $1"; echo "[$(date -Iseconds)] $1" >> "$LOG_FILE"; }
warn() { echo -e "${YELLOW}[warn]${NC} $1"; }
err() { echo -e "${RED}[error]${NC} $1"; exit 1; }

get_active_color() {
    if [ -f "$STATE_FILE" ]; then
        cat "$STATE_FILE"
    else
        echo "blue"
    fi
}

get_inactive_color() {
    local active
    active=$(get_active_color)
    if [ "$active" = "blue" ]; then echo "green"; else echo "blue"; fi
}

wait_for_health() {
    local color=$1
    local port=${APP_PORTS[$color]}
    local max_wait=120
    local elapsed=0

    log "Waiting for ${color} backend to be healthy on port ${port}..."
    while [ $elapsed -lt $max_wait ]; do
        if curl -sf "http://127.0.0.1:${port}/health" > /dev/null 2>&1; then
            log "${color} backend is healthy!"
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
    done

    err "${color} backend failed health check after ${max_wait}s"
}

switch_nginx() {
    local color=$1
    local conf="$SCRIPT_DIR/nginx/${color}.conf"
    local target="/etc/nginx/conf.d/noant.conf"

    if [ ! -f "$conf" ]; then
        err "Nginx config not found: $conf"
    fi

    sudo cp "$conf" "$target"
    sudo nginx -t 2>/dev/null || err "Nginx config test failed"
    sudo nginx -s reload 2>/dev/null || sudo systemctl reload nginx
    log "Nginx switched to ${color}"
}

# ─── Commands ─────────────────────────────────────────────────

cmd_deploy() {
    local active inactive
    active=$(get_active_color)
    inactive=$(get_inactive_color)

    log "═══════════════════════════════════════════════"
    log "  Deploying: ${active} → ${inactive}"
    log "═══════════════════════════════════════════════"

    # Step 1: Build the inactive color
    log "Step 1/5: Building ${inactive} app..."
    cd "$SCRIPT_DIR"
    COLOR=$inactive \
      APP_PORT=${APP_PORTS[$inactive]} \
      FRONTEND_PORT=${FE_PORTS[$inactive]} \
      docker compose -f docker-compose.app.yml -p "noant-${inactive}" build --no-cache 2>&1 | tee -a "$LOG_FILE"

    # Step 2: Start the inactive color
    log "Step 2/5: Starting ${inactive} app..."
    COLOR=$inactive \
      APP_PORT=${APP_PORTS[$inactive]} \
      FRONTEND_PORT=${FE_PORTS[$inactive]} \
      docker compose -f docker-compose.app.yml -p "noant-${inactive}" up -d 2>&1 | tee -a "$LOG_FILE"

    # Step 3: Wait for health check
    log "Step 3/5: Health check..."
    wait_for_health "$inactive"

    # Step 4: Switch nginx
    log "Step 4/5: Switching traffic to ${inactive}..."
    switch_nginx "$inactive"

    # Step 5: Stop old color
    log "Step 5/5: Stopping ${active} app..."
    COLOR=$inactive \
      APP_PORT=${APP_PORTS[$inactive]} \
      FRONTEND_PORT=${FE_PORTS[$inactive]} \
      docker compose -f docker-compose.app.yml -p "noant-${active}" down 2>&1 | tee -a "$LOG_FILE" || true

    # Update state
    echo "$inactive" > "$STATE_FILE"

    log "═══════════════════════════════════════════════"
    log "  Deploy complete! Now live: ${inactive}"
    log "═══════════════════════════════════════════════"
}

cmd_rollback() {
    local active inactive
    active=$(get_active_color)
    inactive=$(get_inactive_color)

    log "Rolling back: ${active} → ${inactive}"

    # Check if the inactive (previous) color's containers exist and are healthy
    local port=${APP_PORTS[$inactive]}
    if ! curl -sf "http://127.0.0.1:${port}/health" > /dev/null 2>&1; then
        err "Cannot rollback: ${inactive} is not running. Redeploy instead."
    fi

    switch_nginx "$inactive"
    echo "$inactive" > "$STATE_FILE"

    log "Rollback complete! Now live: ${inactive}"
}

cmd_status() {
    local active
    active=$(get_active_color)
    echo ""
    echo "  Active color: ${active}"
    echo ""
    echo "  Blue:   app=:${APP_PORTS[blue]}  frontend=:${FE_PORTS[blue]}"
    echo "  Green:  app=:${APP_PORTS[green]}  frontend=:${FE_PORTS[green]}"
    echo ""

    # Check health
    for color in blue green; do
        local port=${APP_PORTS[$color]}
        if curl -sf "http://127.0.0.1:${port}/health" > /dev/null 2>&1; then
            echo "  ${color}: HEALTHY"
        else
            echo "  ${color}: NOT RUNNING"
        fi
    done
    echo ""
}

cmd_build() {
    log "Building both colors..."
    for color in blue green; do
        log "Building ${color}..."
        cd "$SCRIPT_DIR"
        COLOR=$color \
          APP_PORT=${APP_PORTS[$color]} \
          FRONTEND_PORT=${FE_PORTS[$color]} \
          docker compose -f docker-compose.app.yml -p "noant-${color}" build --no-cache 2>&1 | tee -a "$LOG_FILE"
    done
    log "Build complete!"
}

cmd_init() {
    log "Initializing infrastructure..."
    cd "$SCRIPT_DIR"
    docker compose -f docker-compose.infra.yml -p noant-infra up -d 2>&1 | tee -a "$LOG_FILE"

    log "Deploying initial blue stack..."
    COLOR=blue \
      APP_PORT=${APP_PORTS[blue]} \
      FRONTEND_PORT=${FE_PORTS[blue]} \
      docker compose -f docker-compose.app.yml -p noant-blue up -d --build 2>&1 | tee -a "$LOG_FILE"

    wait_for_health "blue"
    switch_nginx "blue"
    echo "blue" > "$STATE_FILE"

    log "Init complete! NOANT is live on blue."
}

# ─── Main ─────────────────────────────────────────────────────

case "${1:-deploy}" in
    deploy)   cmd_deploy ;;
    rollback) cmd_rollback ;;
    status)   cmd_status ;;
    build)    cmd_build ;;
    init)     cmd_init ;;
    *)
        echo "Usage: $0 {deploy|rollback|status|build|init}"
        exit 1
        ;;
esac
