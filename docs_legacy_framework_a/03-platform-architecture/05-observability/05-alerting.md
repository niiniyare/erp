> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Alerting
portal: 3 — Platform Architecture
section: 03-platform-architecture
audience: [devops, sre]
related:
  - "[Observability Overview](01-observability-overview.md)"
  - "[Metrics](04-metrics.md)"
  - "[Runbook Overview](../../09-operations/01-runbook-overview.md)"
---

# Alerting

## Alert Overview

AwoERP uses Prometheus Alertmanager for alert routing. Alerts are routed to PagerDuty for P1/P2, Slack for P3.

## Prometheus Alert Rules

```yaml
# k8s/monitoring/alerts.yaml
groups:
  - name: awoerp.availability
    rules:
      - alert: ServiceDown
        expr: up{job="awoerp"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "AwoERP service is down"
          runbook: "https://docs.awoerp.com/09-operations/02-service-down"

      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{job="awoerp",status=~"5.."}[5m]))
          /
          sum(rate(http_requests_total{job="awoerp"}[5m])) > 0.01
        for: 5m
        labels:
          severity: high
        annotations:
          summary: "5xx error rate above 1%"
          runbook: "https://docs.awoerp.com/09-operations/04-high-error-rate"

  - name: awoerp.latency
    rules:
      - alert: HighP99Latency
        expr: |
          histogram_quantile(0.99,
            sum(rate(http_request_duration_seconds_bucket{job="awoerp"}[5m])) by (le)
          ) > 2
        for: 5m
        labels:
          severity: high
        annotations:
          summary: "P99 latency above 2s"

  - name: awoerp.database
    rules:
      - alert: DBPoolExhausted
        expr: db_pool_idle_connections{job="awoerp"} < 2
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "DB connection pool near exhaustion"
          runbook: "https://docs.awoerp.com/09-operations/03-database-issues"

      - alert: ReadinessProbeFailure
        expr: kube_pod_status_ready{namespace="production",condition="false"} > 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Pods failing readiness check"

  - name: awoerp.events
    rules:
      - alert: EventOutboxBacklog
        expr: event_outbox_undelivered_total{job="awoerp"} > 1000
        for: 5m
        labels:
          severity: low
        annotations:
          summary: "Event outbox backlog exceeds 1000"
          runbook: "https://docs.awoerp.com/09-operations/07-event-outbox"

  - name: awoerp.temporal
    rules:
      - alert: TemporalWorkflowFailures
        expr: temporal_workflow_failed_total{job="awoerp",workflow_type=~"ContractApproval.*"} > 0
        for: 1m
        labels:
          severity: high
        annotations:
          summary: "Critical workflow failures detected"
          runbook: "https://docs.awoerp.com/09-operations/06-temporal-worker"
```

## Alert Thresholds Reference

| Alert | Threshold | Severity | Runbook |
|-------|-----------|---------|---------|
| Service down | Prometheus `up` == 0 for 1m | P1 | service-down |
| 5xx error rate | > 1% over 5m | P2 | high-error-rate |
| P99 latency | > 2s over 5m | P2 | high-error-rate |
| DB pool exhaustion | idle < 2 for 2m | P1 | database-issues |
| Readiness failures | pods not ready 2m | P1 | service-down |
| Event outbox backlog | > 1000 for 5m | P3 | event-outbox |
| Temporal workflow failures | > 0 for critical types | P2 | temporal-worker |
| Pod OOMKilled | count > 0 | P2 | service-down |
| Migration job failed | job status == failed | P1 | migration-failure |

## Alertmanager Routing

```yaml
# alertmanager.yaml
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
      continue: true   # also send to Slack
    - match:
        severity: high
      receiver: pagerduty-high
      continue: true
    - match:
        severity: low
      receiver: slack-default

receivers:
  - name: pagerduty-critical
    pagerduty_configs:
      - routing_key: <PAGERDUTY_KEY>
        severity: critical
  - name: pagerduty-high
    pagerduty_configs:
      - routing_key: <PAGERDUTY_KEY>
        severity: error
  - name: slack-default
    slack_configs:
      - channel: '#alerts'
        send_resolved: true
```

## Silencing Non-Business-Hours Alerts

P3 alerts (event outbox backlog) can be silenced outside business hours using Alertmanager inhibition rules or time-based routing. P1/P2 alerts always page.

## Custom Module Metrics for Alerting

When building a module, register metrics that can be alerted on:

```go
// Register gauge for backlog metrics
pendingApprovals := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
    Name: "contracts_pending_approvals",
    Help: "Number of contracts awaiting approval",
})

// Update periodically (e.g., from a cron activity)
pendingApprovals.Set(float64(count))
```

Alert rule:

```yaml
- alert: ContractApprovalBacklog
  expr: contracts_pending_approvals > 50
  for: 30m
  labels:
    severity: low
  annotations:
    summary: "Many contracts awaiting approval"
```
