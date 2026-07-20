> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Temporal Workflows Overview
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Workflow Patterns Reference](08-workflow-patterns-reference.md)"
  - "[Temporal Architecture](../../../03-platform-architecture/07-temporal-architecture/01-temporal-overview.md)"
  - "[Wire Registration](../09-wire-registration/01-wire-registration-overview.md)"
---

# Temporal Workflows Overview

Temporal handles long-running, multi-step business processes that span minutes, hours, or days. For contracts, it manages approval workflows, expiry processing, and bulk operations.

## When to Use Temporal

Use Temporal (not direct service calls) when:

| Condition | Use Temporal? |
|-----------|--------------|
| Operation completes in < 1 second | No — handle in service layer |
| Operation requires human approval step | Yes |
| Operation may run for hours/days | Yes |
| Operation retries on external service failure | Yes |
| Operation coordinates multiple services | Yes |
| Operation requires compensation (saga) | Yes |
| Scheduled/recurring job | Yes |

## Contracts Workflows

| Workflow | Trigger | Duration |
|----------|---------|---------|
| `ContractApprovalWorkflow` | Contract submitted | Minutes to days |
| `ContractExpiryWorkflow` | Nightly cron | Minutes |
| `ContractBulkImportWorkflow` | User uploads CSV | Minutes to hours |
| `ContractRenewalReminderWorkflow` | Scheduled, 30 days before expiry | Seconds |

## Architecture

```
HTTP Handler
    │
    ├── Simple operations → Service Layer (sync)
    │
    └── Long-running operations → Temporal Client
                                        │
                                        ▼
                                  Temporal Server
                                        │
                                        ▼
                                  Worker (separate process)
                                        │
                                        ├── Workflow orchestrator
                                        └── Activities (DB, email, notifications)
```

## Temporal Client vs. Worker

- **Client**: used by the web server to start/signal/query workflows
- **Worker**: separate process that executes workflow and activity code

Both are wired via the same `temporal.go` provider in Wire.

## Package Layout

```
internal/core/contracts/
├── workflows/
│   ├── approval_workflow.go        # ContractApprovalWorkflow
│   ├── expiry_workflow.go          # ContractExpiryWorkflow
│   ├── bulk_import_workflow.go     # ContractBulkImportWorkflow
│   └── workflow_test.go            # workflow unit tests (replayer)
├── activities/
│   ├── contract_activities.go      # NotifyReviewers, SendApprovalEmail, ...
│   └── activity_test.go
└── worker/
    └── worker.go                   # RegisterWorkflow + RegisterActivity
```

## Task Queue

All contracts workflows run on a dedicated task queue:

```go
const ContractsTaskQueue = "contracts"
```

Using a dedicated task queue per module prevents one module's worker backlog from affecting other modules.

## Key Constraints

- **Workflow code must be deterministic**: no random numbers, no time.Now(), no direct I/O
- **Activities may be non-deterministic**: DB queries, HTTP calls, time.Now() go in activities
- **Idempotency**: activities must be idempotent — Temporal may re-execute them on retry
- **Heartbeating**: long activities (> 10s) must call `activity.RecordHeartbeat`
- **Context**: use `activity.GetLogger(ctx)` not the module's logger in activities
