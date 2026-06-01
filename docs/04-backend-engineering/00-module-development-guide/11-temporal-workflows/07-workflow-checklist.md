---
title: Workflow Checklist
portal: 4 — Backend Engineering
section: 00-module-development-guide/11-temporal-workflows
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-temporal-overview.md
    title: Temporal Overview
---

# Workflow Checklist

Use this checklist when adding a new Temporal workflow to a module.

## Workflow Code

- [ ] Workflow function is registered in `worker/worker.go`
- [ ] All side effects (DB, HTTP, time.Now) are in activities — none in workflow function
- [ ] Timer uses `workflow.NewTimer` (deterministic), not `time.Sleep`
- [ ] Signal channels use named constants from `workflows/` package
- [ ] Query handlers registered for observability (optional but recommended)
- [ ] Workflow ID is deterministic and unique per business entity (e.g., `contract-approval-{contractID}`)

## Activity Code

- [ ] All activities registered via `w.RegisterActivity(activityStruct)`
- [ ] All activities are idempotent (safe to retry)
- [ ] Business errors wrapped in `temporal.NewNonRetryableApplicationError` where appropriate
- [ ] Long activities (> 10s) call `activity.RecordHeartbeat`
- [ ] Activities use `activity.GetLogger(ctx)` not module logger

## Error Handling

- [ ] RetryPolicy configured per activity (not global default only)
- [ ] Saga compensations defer'd for multi-step workflows
- [ ] Cancellation signal handled if workflow may be cancelled

## Testing

- [ ] Workflow tests use `testsuite.WorkflowTestSuite`
- [ ] Both happy path and rejection/timeout paths tested
- [ ] Activity tests use real test database (not mocked)
- [ ] `AssertExpectations` called in `TearDownTest`

## Registration

- [ ] Worker provider added to `contracts.WorkerSet` in `contracts.go`
- [ ] `WorkerSet` included in `InitializeApp` via Wire
- [ ] Task queue constant matches across worker and workflow starters
- [ ] Cron schedules registered at startup with `IsAlreadyStartedError` guard

## Operations

- [ ] Workflow type name matches across client `.ExecuteWorkflow` call and registration
- [ ] Signal names are constants shared between sender and receiver
- [ ] Workflow visible in Temporal Web UI after first run
