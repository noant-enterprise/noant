# NOANT Monitoring

Observability stack: Prometheus metrics, Grafana dashboards, alerting rules.

## Components

### Prometheus

| File | Purpose |
|------|---------|
| `prometheus/prometheus.yml` | Scrape config — pulls metrics from backend `:9090/metrics` |
| `prometheus/recording-rules.yml` | Pre-computed PromQL expressions (19 rules) for dashboard performance |

**Scrape targets:**
- Backend: `http://backend:8080/metrics`
- Redis exporter (if deployed)

### Grafana

| File | Purpose |
|------|---------|
| `grafana/dashboard.json` | Main NOANT dashboard (15 panels) |
| `grafana/alerting.yml` | Alert rules (11 alert conditions) |
| `grafana/contact-points.yml` | Notification targets (email, Slack, PagerDuty) |
| `grafana/notification-policies.yml` | Alert routing and escalation |

## Dashboard Panels (15)

| Row | Panel | Metric |
|-----|-------|--------|
| Overview | Requests/sec | `rate(http_requests_total[5m])` |
| Overview | Error rate | `rate(http_requests_total{status=~"5.."}[5m])` |
| Overview | P95 latency | `histogram_quantile(0.95, ...)` |
| Overview | Active WebSockets | `ws_connections_active` |
| AI | Response time | `ai_response_duration_seconds` |
| AI | Accuracy rate | `ai_accuracy_ratio` |
| AI | Unknown questions | `ai_unknown_total` |
| Infrastructure | DB connections | `db_connections_active` |
| Infrastructure | Redis memory | `redis_memory_used_bytes` |
| Infrastructure | Cache hit rate | `cache_hits / (cache_hits + cache_misses)` |
| Business | API calls by user | `sum by (user_id) (api_calls_total)` |
| Business | Plan distribution | `users_by_plan` |
| Business | Revenue | `revenue_total` |
| Security | Auth failures | `rate(auth_failures_total[5m])` |
| Security | Rate limit hits | `rate(rate_limit_total[5m])` |

## Alert Rules (11)

| Alert | Condition | Severity |
|-------|-----------|----------|
| HighErrorRate | > 5% errors for 5m | Critical |
| HighLatency | P95 > 2s for 5m | Warning |
| DatabaseDown | Unhealthy for 1m | Critical |
| RedisDown | Unhealthy for 1m | Critical |
| HighMemory | > 85% for 5m | Warning |
| HighCPU | > 80% for 5m | Warning |
| AIAccuracyDrop | < 80% for 10m | Warning |
| UnknownQuestionsSpike | > 50/hr | Info |
| WebSocketDisconnects | > 100/hr | Warning |
| AuthFailureSpike | > 50/hr | Warning |
| DiskSpaceLow | < 10% free | Critical |

## Recording Rules (19)

Pre-computed metrics for dashboard performance. Key rules:

- `noant:http_requests:rate5m` — request rate per endpoint
- `noant:errors:rate5m` — error rate per endpoint
- `noant:latency:p95_5m` — P95 latency bucket
- `noant:ai:accuracy_ratio` — AI accuracy rolling average
- `noant:db:connections_active` — active DB connections
- `noant:ws:connections_active` — active WebSocket connections

## Setup

### Docker Compose (Production)

Add to `docker-compose.production.yml`:

```yaml
services:
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./monitoring/prometheus:/etc/prometheus
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    volumes:
      - ./monitoring/grafana:/var/lib/grafana/provisioning
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: your-password
```

### Standalone

```bash
# Prometheus
prometheus --config.file=monitoring/prometheus/prometheus.yml

# Grafana
grafana-server --homepath=/usr/share/grafana
```

## Accessing

| Service | Default URL | Credentials |
|---------|-------------|-------------|
| Grafana | `http://localhost:3000` | admin / (set password) |
| Prometheus | `http://localhost:9090` | None |

In production, put both behind Nginx with authentication.

## Sentry Integration

Error monitoring is separate from Prometheus/Grafana. Sentry captures:

- **Backend:** All panics, 5xx errors, slow queries (> 2s)
- **Frontend:** Unhandled exceptions, uncaught promise rejections
- **Environment:** `production` / `development`
- **Release:** `noant@2.0.0`

20% of transactions are sampled for tracing. Sensitive headers (Authorization, X-API-Key) are scrubbed before sending.

See `scripts/SENTRY_DASHBOARD.md` for the Sentry dashboard setup guide (27 widgets, 8 alert rules).
