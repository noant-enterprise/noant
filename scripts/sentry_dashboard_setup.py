#!/usr/bin/env python3
"""
NOANT — Sentry Dashboard Setup
Creates a world-class monitoring dashboard via Sentry Management API.

Usage:
  export SENTRY_AUTH_TOKEN=your_auth_token
  export SENTRY_ORG=divineshedrack33220
  export SENTRY_DSN_ORG_ID=4511779671375872  # from DSN
  pip install requests
  python3 sentry_dashboard_setup.py
"""

import os
import sys
import json
import time

try:
    import requests
except ImportError:
    print("pip install requests")
    sys.exit(1)

SENTRY_BASE = "https://sentry.io/api/0"
AUTH_TOKEN = os.environ.get("SENTRY_AUTH_TOKEN", "")
ORG = os.environ.get("SENTRY_ORG", "")

if not AUTH_TOKEN or not ORG:
    print("Set SENTRY_AUTH_TOKEN and SENTRY_ORG environment variables")
    sys.exit(1)

HEADERS = {
    "Authorization": f"Bearer {AUTH_TOKEN}",
    "Content-Type": "application/json",
}


def api(method, path, data=None):
    url = f"{SENTRY_BASE}{path}"
    r = requests.request(method, url, headers=HEADERS, json=data, timeout=30)
    if r.status_code >= 400:
        print(f"  [{r.status_code}] {path}: {r.text[:200]}")
        return None
    return r.json() if r.content else {}


def get_or_create_project():
    """Find the noant project or use first available."""
    projects = api("GET", f"/organizations/{ORG}/projects/")
    if not projects:
        print("No projects found. Create one in Sentry first.")
        sys.exit(1)
    for p in projects:
        if "noant" in p["slug"].lower():
            return p["slug"], p["id"]
    return projects[0]["slug"], projects[0]["id"]


def get_or_create_dashboard(title):
    """Find existing dashboard or create new one."""
    dashboards = api("GET", f"/organizations/{ORG}/dashboards/")
    if dashboards and "dashboards" in dashboards:
        for d in dashboards["dashboards"]:
            if d["title"] == title:
                print(f"  Found existing dashboard: {title} (id={d['id']})")
                return d["id"]

    result = api("POST", f"/organizations/{ORG}/dashboards/", {
        "title": title,
        "description": "NOANT world-class monitoring dashboard",
        "dashboardType": "owner",
    })
    if result and "id" in result:
        print(f"  Created dashboard: {title} (id={result['id']})")
        return result["id"]
    print(f"  Failed to create dashboard: {result}")
    return None


def add_widget(dashboard_id, widget):
    """Add a widget to dashboard."""
    result = api(
        "POST",
        f"/organizations/{ORG}/dashboards/{dashboard_id}/widgets/",
        widget,
    )
    if result and "id" in result:
        print(f"    Widget: {widget['title']} (id={result['id']})")
        return result["id"]
    return None


# ============================================================
# WIDGET DEFINITIONS
# ============================================================

WIDGETS = [
    # ---- ROW 1: TOP-LEVEL KPIs ----
    {
        "title": "Total Errors (24h)",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Errors",
            "fields": ["count()"],
            "conditions": "level:error",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Affected Users (24h)",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Users",
            "fields": ["count_unique(user)"],
            "conditions": "level:error",
            "aggregates": [{"fields": ["count_unique(user)"], "kind": "function"}],
        }],
    },
    {
        "title": "Error Rate %",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Rate",
            "fields": ["failure_rate()"],
            "conditions": "",
            "aggregates": [{"fields": ["failure_rate()"], "kind": "function"}],
        }],
    },
    {
        "title": "P95 Latency",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "P95",
            "fields": ["p95(transaction.duration)"],
            "conditions": "",
            "aggregates": [{"fields": ["p95(transaction.duration)"], "kind": "function"}],
        }],
    },

    # ---- ROW 2: ERROR TRENDS ----
    {
        "title": "Errors Over Time",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "Errors",
            "fields": ["count()"],
            "conditions": "level:error",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Errors by Endpoint",
        "displayType": "top_n",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "By Path",
            "fields": ["count()"],
            "conditions": "level:error",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
            "orderby": "-count()",
        }],
        "displayOptions": {"topEvents": 10, "yAxis": "count()"},
    },
    {
        "title": "New vs Recurring Errors",
        "displayType": "stacked_area",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [
            {
                "name": "New",
                "fields": ["count()"],
                "conditions": "level:error issue.type:regression",
                "aggregates": [{"fields": ["count()"], "kind": "function"}],
            },
            {
                "name": "Recurring",
                "fields": ["count()"],
                "conditions": "level:error -issue.type:regression",
                "aggregates": [{"fields": ["count()"], "kind": "function"}],
            },
        ],
    },

    # ---- ROW 3: BACKEND (Go/Gin) ----
    {
        "title": "Backend Errors by Status Code",
        "displayType": "top_n",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "By Status",
            "fields": ["count()"],
            "conditions": "level:error",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
            "orderby": "-count()",
        }],
        "displayOptions": {"topEvents": 10, "yAxis": "count()"},
    },
    {
        "title": "API Request Duration (P50/P95/P99)",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [
            {
                "name": "P50",
                "fields": ["p50(transaction.duration)"],
                "conditions": "",
                "aggregates": [{"fields": ["p50(transaction.duration)"], "kind": "function"}],
            },
            {
                "name": "P95",
                "fields": ["p95(transaction.duration)"],
                "conditions": "",
                "aggregates": [{"fields": ["p95(transaction.duration)"], "kind": "function"}],
            },
            {
                "name": "P99",
                "fields": ["p99(transaction.duration)"],
                "conditions": "",
                "aggregates": [{"fields": ["p99(transaction.duration)"], "kind": "function"}],
            },
        ],
    },
    {
        "title": "Database Errors",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "DB Errors",
            "fields": ["count()"],
            "conditions": "level:error (database OR sql OR tidb OR connection pool)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },

    # ---- ROW 4: AI/LLM (Groq) ----
    {
        "title": "AI Response Failures",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "AI Failures",
            "fields": ["count()"],
            "conditions": "level:error (groq OR ai OR llama OR model OR chat)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "AI Response Latency (P95)",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "AI P95",
            "fields": ["p95(transaction.duration)"],
            "conditions": "transaction:/api/v1/chats/direct-chat OR transaction:/api/v1/chats/stream",
            "aggregates": [{"fields": ["p95(transaction.duration)"], "kind": "function"}],
        }],
    },
    {
        "title": "Groq Rate Limit Events",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Rate Limited",
            "fields": ["count()"],
            "conditions": "(429 OR rate_limit OR too_many_requests)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },

    # ---- ROW 5: WhatsApp/OpenWA ----
    {
        "title": "WhatsApp Errors",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "WA Errors",
            "fields": ["count()"],
            "conditions": "level:error (openwa OR whatsapp OR webhook OR session)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Webhook Delivery Failures",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "Webhook Fails",
            "fields": ["count()"],
            "conditions": "level:error (openwa OR whatsapp) (webhook OR delivery OR send)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Session Reconnection Events",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "Reconnects",
            "fields": ["count()"],
            "conditions": "(reconnect OR session OR disconnected OR rotated)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },

    # ---- ROW 6: SECURITY ----
    {
        "title": "Auth Failures (24h)",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Auth Fails",
            "fields": ["count()"],
            "conditions": "level:error (unauthorized OR invalid_credentials OR forbidden OR login OR auth)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Rate Limiting Events",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "Rate Limited",
            "fields": ["count()"],
            "conditions": "(429 OR rate_limit OR too_many_requests OR throttled)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Security Errors by User",
        "displayType": "top_n",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "By User",
            "fields": ["count_unique(user)"],
            "conditions": "level:error (unauthorized OR forbidden OR csrf OR invalid_token)",
            "aggregates": [{"fields": ["count_unique(user)"], "kind": "function"}],
            "orderby": "-count_unique(user)",
        }],
        "displayOptions": {"topEvents": 10, "yAxis": "count_unique(user)"},
    },

    # ---- ROW 7: FRONTEND ----
    {
        "title": "Frontend JS Errors",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "JS Errors",
            "fields": ["count()"],
            "conditions": "level:error (TypeError OR ReferenceError OR SyntaxError OR undefined OR null)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Frontend Errors Over Time",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "JS Trend",
            "fields": ["count()"],
            "conditions": "level:error platform:javascript",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Most Common JS Errors",
        "displayType": "top_n",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Top JS",
            "fields": ["count()"],
            "conditions": "level:error platform:javascript",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
            "orderby": "-count()",
        }],
        "displayOptions": {"topEvents": 10, "yAxis": "count()"},
    },

    # ---- ROW 8: RELEASE & DEPLOY ----
    {
        "title": "Errors by Release",
        "displayType": "top_n",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "By Release",
            "fields": ["count()"],
            "conditions": "level:error",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
            "orderby": "-count()",
        }],
        "displayOptions": {"topEvents": 5, "yAxis": "count()"},
    },
    {
        "title": "Error-Free Session Rate",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Session Health",
            "fields": ["count_unique(session)"],
            "conditions": "level:error",
            "aggregates": [{"fields": ["count_unique(session)"], "kind": "function"}],
        }],
    },

    # ---- ROW 9: INFRASTRUCTURE ----
    {
        "title": "Redis Errors",
        "displayType": "big_number",
        "widgetType": "discover",
        "interval": "1d",
        "queries": [{
            "name": "Redis",
            "fields": ["count()"],
            "conditions": "level:error (redis OR connection refused OR ECONNREFUSED)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "HTTP 5xx Errors Over Time",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "5xx",
            "fields": ["count()"],
            "conditions": "level:error (500 OR 502 OR 503 OR 504 OR INTERNAL_ERROR OR SERVICE_UNAVAILABLE)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
    {
        "title": "Unhandled Exceptions",
        "displayType": "line",
        "widgetType": "discover",
        "interval": "1h",
        "queries": [{
            "name": "Panics",
            "fields": ["count()"],
            "conditions": "level:error (panic OR goroutine OR unexpected)",
            "aggregates": [{"fields": ["count()"], "kind": "function"}],
        }],
    },
]


def main():
    print(f"NOANT Sentry Dashboard Setup")
    print(f"Organization: {ORG}")
    print()

    slug, project_id = get_or_create_project()
    print(f"Project: {slug} (id={project_id})")
    print()

    dashboard_id = get_or_create_dashboard("NOANT — Production Dashboard")
    if not dashboard_id:
        print("Cannot create dashboard. Check permissions.")
        sys.exit(1)

    print(f"\nAdding {len(WIDGETS)} widgets...")
    created = 0
    for widget in WIDGETS:
        result = add_widget(dashboard_id, widget)
        if result:
            created += 1
        time.sleep(0.3)  # rate limit

    print(f"\nDone! Created {created}/{len(WIDGETS)} widgets.")
    print(f"\nDashboard URL: https://sentry.io/organizations/{ORG}/dashboards/")
    print(f"\nNext steps:")
    print(f"  1. Open the dashboard in Sentry")
    print(f"  2. Arrange widgets into rows (drag & drop)")
    print(f"  3. Set default time range to 'Last 24 hours'")
    print(f"  4. Add dashboard to favorites for your team")


if __name__ == "__main__":
    main()
