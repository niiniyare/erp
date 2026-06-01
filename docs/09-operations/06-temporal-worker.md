---
title: Temporal Worker Stuck Runbook
portal: 9 — Operations
section: 09-operations
audience: [devops, sre, backend-engineer]
related:
  - "[Operations Overview](01-operations-overview.md)"
  - "[Temporal Architecture](../03-platform-architecture/07-temporal-architecture/01-temporal-overview.md)"
  - "[Event Outbox Backlog](07-event-outbox.md)"
---

# Temporal Worker Stuck Runbook

**Trigger**: Workflows not progressing, Temporal UI shows workflows stuck in `Running` state, or alert `awoerp-temporal-workflow-failures`.

## 1. Access Temporal UI

```bash
# Port-forward Temporal UI (if not exposed via VPN)
kubectl port-forward svc/temporal-ui 8088:8088 -n production
# Open http://localhost:8088
```

In the UI:
- Check **Running** workflows — are they stuck on an activity?
- Check **Failed** workflows — what error?
- Check **Timed Out** workflows — activity or workflow timeout?

## 2. Check Worker Logs

```bash
# Workers run in the same pod as the API server
kubectl logs -l app=awoerp-server -n production --since=30m \
  | grep -E '"module":"temporal|worker|activity|workflow"'

# Look for worker connection errors
kubectl logs -l app=awoerp-server -n production --since=30m \
  | grep -E 'temporal.*error|unable to connect|dial tcp'
```

## 3. Diagnose Stuck Workflows

### Workers Not Polling

If no new workflow tasks are being picked up:

```bash
# Check Temporal server connectivity from app pod
kubectl exec -it deploy/awoerp-server -n production -- \
  nc -zv temporal-frontend 7233
```

If connection refused → Temporal server is down. Check Temporal deployment:

```bash
kubectl get pods -n temporal
kubectl logs -l app=temporal -n temporal --tail=50
```

### Activity Failing Repeatedly

In Temporal UI, open a stuck workflow → click the failing activity → view the last error.

Common causes:
- DB connection error inside activity → transient, will auto-retry
- Business rule violation with retryable error → should be non-retryable, fix code
- Activity timeout too short for workload → increase `StartToCloseTimeout`

### Workflow Stuck on Signal

A workflow awaiting a signal (`ContractApproval`) will sit in Running indefinitely until signaled. This is normal behavior — verify with business team whether the workflow is legitimately waiting for approval.

To check:

```bash
# Using tctl
tctl wf describe --workflow-id contract-approval-{uuid}
```

If the signal should have been sent but wasn't, the service that sends the signal failed silently. Check the HTTP handler logs around the time of the action (e.g., approve button click).

## 4. Manually Signal a Stuck Workflow

If a workflow is legitimately stuck waiting for a signal that was lost:

```bash
# Send approval signal manually
tctl wf signal \
  --workflow-id contract-approval-{uuid} \
  --name approval-decision \
  --input '{"approved": true, "comment": "Manual recovery", "approver_id": "..."}'
```

**Confirm with the business team before sending manual signals** — this triggers real state changes.

## 5. Terminate and Restart a Workflow

If a workflow is in a bad state and cannot recover:

```bash
# Terminate workflow
tctl wf terminate \
  --workflow-id contract-approval-{uuid} \
  --reason "Manual recovery: stuck workflow"
```

Then restart from the application:

```bash
# Via API (if endpoint exists)
curl -X POST /api/v1/contracts/{id}/restart-workflow \
  -H "Authorization: Bearer {token}"

# Or via awoctl
awoctl workflows restart --workflow-id contract-approval-{uuid}
```

## 6. Bulk Failed Workflows

If many workflows failed due to a deployment issue:

```bash
# List failed workflows in Temporal UI or via tctl
tctl wf list --status Failed --query 'WorkflowType="ContractApprovalWorkflow"'

# After fixing the root cause, re-run all failed workflows
# (requires custom awoctl command or manual script)
```

## 7. Worker Registration Check

If no workers are polling:

```bash
# Verify task queues have registered workers
tctl tq describe --task-queue awoerp.contracts
tctl tq describe --task-queue awoerp.finance
```

Look for `pollers` in the output. If empty → workers are not connecting. Restart pods:

```bash
kubectl rollout restart deployment/awoerp-server -n production
```

## Resolution Criteria

- Temporal UI shows workflows progressing (no new stuck ones appearing)
- `tctl tq describe` shows active pollers on all task queues
- No `workflow_task_failed` errors in worker logs
- Failed workflow count returns to 0 for critical workflow types
