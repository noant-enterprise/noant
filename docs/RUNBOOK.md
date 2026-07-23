# NOANT Incident Response Runbook

## Severity Levels

- **SEV1** (Critical): Platform down, data loss risk, security breach. All hands on deck. Page immediately.
- **SEV2** (High): Major feature broken, significant performance degradation. On-call engineer responds within 15 minutes.
- **SEV3** (Medium): Minor feature broken, cosmetic issues. Next business day response.
- **SEV4** (Low): Monitoring alert, no user impact. Address in normal workflow.

---

## Common Incidents & Resolution

### 1. API Returns 500 Errors

**Symptoms**: Elevated 5xx rate on Grafana Request Rate panel. Users reporting errors.

**Diagnosis**:
- Grafana dashboard → Request Rate panel (filter by status=500)
- Check backend logs for panic traces: `journalctl -u noant --since "10 minutes ago"`
- Identify affected endpoints and affected time window

**Resolution**:
1. Check database connectivity: `mysql -u noant -p -e "SELECT 1" noant_prod`
2. Check Redis availability: `redis-cli ping`
3. Check Groq API status: https://status.groq.com
4. If caused by a bad deploy, see **Rollback Procedures** below

---

### 2. AI Responses Failing / Slow

**Symptoms**: Users report AI chat not responding or timing out. Grafana AI Response Rate drops.

**Diagnosis**:
- Grafana → AI Response Rate, AI Duration panels
- Check Groq API status page: https://status.groq.com
- Review backend logs for `CircuitBreakerOpen` errors

**Resolution**:
1. Circuit breaker may have tripped (auto-recovers in 60s) — wait and monitor
2. If persistent, check Groq API key validity and rate limits
3. Verify outbound network connectivity to `api.groq.com`
4. If Groq is fully down, communicate to users; fallback options are not available for AI features

---

### 3. Database Connection Exhaustion

**Symptoms**: Timeouts on all DB-dependent endpoints. Grafana DB Connections gauge spikes.

**Diagnosis**:
- Grafana → DB Connections gauge, query duration panels
- Check running queries: `SHOW PROCESSLIST` on MySQL
- Look for long-running queries or deadlocks

**Resolution**:
1. Kill stalled queries: `KILL <process_id>`
2. Check connection pool settings in backend config
3. Restart the backend service to release leaked connections
4. If recurring, review code for missing connection releases (defer rows.Close(), etc.)

---

### 4. Redis Down / Unavailable

**Symptoms**: Rate limiting, session caching, and token blacklisting fail. Backend logs show Redis connection errors.

**Diagnosis**:
- Grafana → Redis Connections gauge
- `redis-cli ping` from the application server
- Check Redis process: `systemctl status redis`

**Resolution**:
1. Backend falls back to in-memory mode automatically (no crash)
2. Restart Redis: `systemctl restart redis`
3. After restart, verify data integrity — session caches will be cold
4. If Redis data is corrupted, flush and let caches repopulate: `redis-cli FLUSHALL`

---

### 5. WebSocket Connection Issues

**Symptoms**: Real-time features (chat, live updates) not working. Clients fail to connect or disconnect immediately.

**Diagnosis**:
- Check backend logs for WebSocket upgrade errors
- Verify CORS origins match the client domain
- Check reverse proxy (nginx/caddy) WebSocket configuration

**Resolution**:
1. Ensure proxy passes `Upgrade` and `Connection` headers:
   ```
   proxy_set_header Upgrade $http_upgrade;
   proxy_set_header Connection "upgrade";
   ```
2. Verify `WS_ALLOWED_ORIGINS` env var includes the client domain
3. Check that the WebSocket auth middleware isn't rejecting valid tokens
4. Restart backend if goroutine leak is suspected (see **High Memory Usage**)

---

### 6. WhatsApp Channel Down

**Symptoms**: WhatsApp messages not being sent or received. OpenWA integration reports errors.

**Diagnosis**:
- Check OpenWA session status: `curl -s https://your-domain/api/v1/openwa/status`
- Check backend logs for OpenWA connection errors
- Verify OpenWA server is running and reachable

**Resolution**:
1. Restart the OpenWA session via the API or dashboard
2. Check OpenWA server health and logs
3. Verify the WhatsApp account hasn't been logged out (QR re-scan may be needed)
4. If the OpenWA server itself is down, restart it: `systemctl restart openwa`

---

### 7. High Memory Usage

**Symptoms**: OOM kills, slow response times, Grafana memory metrics climbing steadily.

**Diagnosis**:
- Check system metrics: `free -h`, `top`
- Go pprof heap dump: `curl -o /tmp/heap.pbprof http://localhost:6060/debug/pprof/heap`
- Analyze with `go tool pprof /tmp/heap.pbprof`

**Resolution**:
1. Check for goroutine leaks: `curl http://localhost:6060/debug/pprof/goroutine?debug=1`
2. Look for growing heap in pprof — identifies memory-intensive operations
3. If caused by a specific request pattern, rate-limit that endpoint
4. As a temporary fix, restart the service: `systemctl restart noant`
5. If recurring, profile with CPU and memory profiles to find the root cause

---

### 8. Rate Limiting Triggered

**Symptoms**: Clients receiving 429 Too Many Requests responses.

**Diagnosis**:
- Check 429 response rate in Grafana or backend logs
- Identify if the traffic is legitimate or abusive (check IP patterns, user agents)
- Review X-RateLimit-* headers in responses

**Resolution**:
- **Legitimate traffic spike**: Temporarily increase rate limits in backend config, then investigate why usage increased
- **Abuse / bot traffic**: Add IP to blocklist, tighten rate limits for anonymous endpoints
- **Client bug causing retries**: Notify the client team; implement exponential backoff on the client side

### 9. Multi-Tenancy Data Leakage

**Symptoms**: User sees data belonging to another organization. Grafana anomaly alerts on cross-org queries.

**Diagnosis**:
- Check Sentry for any 401/403 errors on org-scoped endpoints
- Verify `org_id` is being set in JWT claims: check `users.org_id` is not NULL
- Check backend logs for queries missing `org_id` filter

**Resolution**:
1. Verify all handler methods use `getScopeID(c)` instead of `getUserID(c)` for org-scoped repos
2. Check service layer: all domain struct creations must set `OrgID`
3. Run migration 021 to backfill any missing `org_id` values
4. If confirmed breach, rotate all JWT tokens and audit access logs

---

## Rollback Procedures

### Backend Rollback

```bash
# Revert to previous binary
cp /opt/noant/bin/main.prev /opt/noant/bin/main
systemctl restart noant

# Verify health after restart
curl -s https://your-domain/health | jq .
```

### Database Rollback

```bash
# List applied migrations
mysql -u noant -p noant_prod -e "SELECT * FROM schema_migrations ORDER BY version;"

# Manually revert specific migration (use with caution)
# Identify the down migration SQL and run it against the database
# Always take a backup first:
mysqldump -u noant -p noant_prod > /tmp/noant_backup_$(date +%Y%m%d_%H%M%S).sql
```

### Frontend Rollback

```bash
# If using CDN, revert the deploy
# If self-hosted:
cp -r /opt/noant/frontend.prev /opt/noant/frontend
# Clear CDN cache if applicable
```

---

## Escalation Contacts

| Role | Contact |
|------|---------|
| Platform Team | platform@noant.example.com |
| DevOps On-Call | @noant-oncall (Slack) |
| Security | security@noant.example.com |
| Database Admin | dba@noant.example.com |

**Escalation Path**: On-Call Engineer → Platform Team Lead → CTO (for SEV1 only)

---

## Monitoring & Alerts

| Tool | URL |
|------|-----|
| Grafana | https://grafana.noant.example.com/d/noant-main |
| Prometheus | https://prometheus.noant.example.com |
| AlertManager | https://alertmanager.noant.example.com |
| Uptime Monitor | https://uptime.noant.example.com |
| Sentry | https://sentry.io/organizations/noant/projects/ |

---

## Health Checks

| Endpoint | Purpose | Auth Required |
|----------|---------|---------------|
| `GET /health` | Full health check (DB + Redis + Groq) | No |
| `GET /ping` | Simple liveness probe (returns 200) | No |
| `GET /metrics` | Prometheus metrics | Yes (admin) |

**Quick health check**:
```bash
curl -s https://your-domain/health | jq .
```

**Expected healthy response**:
```json
{
  "status": "healthy",
  "database": "ok",
  "redis": "ok",
  "groq": "ok"
}
```

---

## Staging Environment Tuning Guide

After deploying to staging, run these steps to calibrate thresholds:

### 1. Load Test Baseline

Run k6 smoke test against staging to establish baseline metrics:

```bash
k6 run --summary-export=baseline.json tests/load/smoke.js
```

Record the P95 latency and error rate. These become your production baselines.

### 2. Grafana Alert Threshold Tuning

After 1 week of production data, adjust alert thresholds in `monitoring/grafana/alerting.yml`:

| Alert | Initial Threshold | Tune Based On |
|-------|------------------|---------------|
| HighErrorRate | >5% | If P99 is always 3-4%, raise to 10% |
| HighLatencyP95 | >3s | If normal P95 is 1.5s, keep; if 2.8s, raise to 4s |
| AIFailureRate | >20% | Usually accurate; Groq outages will trigger this |
| DatabaseConnectionExhaustion | >80 | Adjust based on your pool size |
| OpenWAQueueBacklog | >100 | Adjust based on message volume |

### 3. Recording Rules Validation

After deploying recording rules, verify in Prometheus:

```
noant:request_rate:rate5m
noant:error_rate:ratio5m
noant:latency:p95_5m
```

These should be populated within 15 seconds. If not, check rule syntax:

```bash
promtool check rules monitoring/prometheus/recording-rules.yml
```

### 4. DB Query Tracing Validation

After deploying TracedDB, verify metrics are being recorded:

```bash
curl -s localhost:8080/metrics | grep noant_db_queries_total
```

You should see counters incrementing for `exec`, `query`, `query_row`, and `ping` operations.

### 5. Capacity Planning

Use these metrics to plan capacity:

- **CPU**: `noant:request_rate:rate5m` × average CPU per request
- **Memory**: `noant:connections:current` × average memory per connection
- **DB connections**: `noant_db_connections` peak × 1.5 safety margin
- **Redis connections**: `noant_redis_connections` peak × 2 safety margin

### 6. Performance Regression Detection

Set up a weekly k6 run in staging and compare with baseline:

```bash
k6 run --summary-export=current.json tests/load/stress.js
# Compare current.json P95 with baseline.json P95
# If P95 increased by >20%, investigate before promoting to production
```
