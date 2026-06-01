---
title: Monitoring and Alerting
portal: 6 — DevOps
section: 06-devops
audience: [devops, sre]
related:
  - "[DevOps Overview](01-devops-overview.md)"
  - "[Operations Overview](../09-operations/01-operations-overview.md)"
  - "[Metrics](../03-platform-architecture/05-observability/04-metrics.md)"
---

# Monitoring and Alerting

## Metrics Stack

| Component | Role |
|-----------|------|
| Prometheus | Scrapes `/metrics` endpoint (port 9090), stores time-series |
| Grafana | Dashboard visualization |
| Alertmanager | Routes alerts to PagerDuty / Slack |
| OTel Collector | Receives traces, forwards to Jaeger/Tempo |

## Key Metrics

### HTTP

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Label: `method`, `path`, `status` |
| `http_request_duration_seconds` | Histogram | Buckets: 10ms–10s |
| `http_requests_in_flight` | Gauge | Active requests |

### Database

| Metric | Type | Description |
|--------|------|-------------|
| `db_pool_open_connections` | Gauge | Active pool connections |
| `db_pool_idle_connections` | Gauge | Idle pool connections |
| `db_query_duration_seconds` | Histogram | Label: `query` (hashed) |

### Business

| Metric | Type | Description |
|--------|------|-------------|
| `contracts_created_total` | Counter | Label: `tenant_id`, `contract_type` |
| `contracts_status_transitions_total` | Counter | Label: `from`, `to` |
| `event_outbox_pending` | Gauge | Undelivered outbox events |
| `temporal_workflow_backlog` | Gauge | Label: `task_queue` |

## Prometheus Scrape Config

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'awoerp'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names: ['awoerp']
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        target_label: __metrics_path__
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_port]
        target_label: __address__
        replacement: $1:9090
```

Pod annotations to enable scraping:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "9090"
  prometheus.io/path: "/metrics"
```

## Alert Rules

```yaml
# alerts.yml
groups:
  - name: awoerp.api
    rules:

    - alert: HighErrorRate
      expr: |
        rate(http_requests_total{status=~"5.."}[5m]) /
        rate(http_requests_total[5m]) > 0.01
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "HTTP 5xx error rate > 1% for 2 minutes"
        runbook: "https://docs.internal/ops/04-high-error-rate"

    - alert: HighP99Latency
      expr: |
        histogram_quantile(0.99,
          rate(http_request_duration_seconds_bucket[5m])) > 2
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "P99 latency > 2s"

    - alert: DBPoolExhausted
      expr: db_pool_idle_connections == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Database connection pool exhausted"
        runbook: "https://docs.internal/ops/03-database-issues"

    - alert: EventOutboxBacklog
      expr: event_outbox_pending > 500
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Event outbox has {{ $value }} pending events"
        runbook: "https://docs.internal/ops/07-event-outbox"

    - alert: TemporalWorkerDown
      expr: temporal_workflow_backlog > 100
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Temporal task queue backlog growing — worker may be down"
        runbook: "https://docs.internal/ops/06-temporal-worker"

    - alert: ServiceDown
      expr: up{job="awoerp"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "AwoERP pod is down"
        runbook: "https://docs.internal/ops/02-service-down"
```

## Grafana Dashboards

### Required Dashboards

| Dashboard | Key Panels |
|-----------|-----------|
| API Overview | Request rate, error rate, p50/p99 latency |
| Database | Pool utilization, query latency by type |
| Business Metrics | Contracts created/day, active contracts, pipeline throughput |
| Infrastructure | Pod CPU/memory, restarts |
| Temporal | Workflow success/failure, queue depth |
| Events | Outbox pending, delivery rate |

### Key Queries

**Error rate (5m window):**
```promql
rate(http_requests_total{status=~"5.."}[5m])
/ rate(http_requests_total[5m]) * 100
```

**P99 latency:**
```promql
histogram_quantile(0.99,
  rate(http_request_duration_seconds_bucket[5m]))
```

**DB pool utilization:**
```promql
db_pool_open_connections / db_pool_open_connections offset 1d
```

**Contracts created last 24h:**
```promql
increase(contracts_created_total[24h])
```

## Alertmanager Routing

```yaml
# alertmanager.yml
route:
  group_by: ['alertname', 'cluster']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: slack-default
  routes:
    - match:
        severity: critical
      receiver: pagerduty-critical
      continue: true
    - match:
        severity: critical
      receiver: slack-critical

receivers:
  - name: pagerduty-critical
    pagerduty_configs:
      - routing_key: "$PAGERDUTY_KEY"
        description: "{{ .CommonAnnotations.summary }}"
        details:
          runbook: "{{ .CommonAnnotations.runbook }}"

  - name: slack-critical
    slack_configs:
      - api_url: "$SLACK_WEBHOOK_CRITICAL"
        channel: '#alerts-critical'
        title: "CRITICAL: {{ .CommonAnnotations.summary }}"
        text: "Runbook: {{ .CommonAnnotations.runbook }}"

  - name: slack-default
    slack_configs:
      - api_url: "$SLACK_WEBHOOK"
        channel: '#alerts'
        title: "{{ .GroupLabels.alertname }}"
```

## Health Endpoints

| Endpoint | Port | Purpose | Used by |
|----------|------|---------|---------|
| `GET /health/live` | 8081 | Is process alive? | K8s liveness probe |
| `GET /health/ready` | 8081 | Is DB + Redis up? | K8s readiness probe |
| `GET /health/startup` | 8081 | Did migrations run? | K8s startup probe |
| `GET /metrics` | 9090 | Prometheus metrics | Prometheus scraper |

Ready check response:

```json
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "temporal": "ok"
  }
}
```

If any check fails, response is `503 Service Unavailable` and the pod is removed from load balancer traffic.

## SLO Targets

| SLO | Target | Alert threshold |
|-----|--------|----------------|
| API availability | 99.9% uptime | Page at 99.5% |
| P99 latency | < 500ms | Page at > 2s |
| Error rate | < 0.1% | Page at > 1% |
| DB query p99 | < 100ms | Alert at > 500ms |
