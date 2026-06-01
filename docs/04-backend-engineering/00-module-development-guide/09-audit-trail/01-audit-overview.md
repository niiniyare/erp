---
title: Audit Trail Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/09-audit-trail
audience: [backend-engineer, tech-lead]
related:
  - path: ../07-core-integrations/05-audit-integration.md
    title: Audit Integration
  - path: ./02-audit-events.md
    title: Audit Events
---

# Audit Trail Overview

The audit trail records an immutable history of every state-changing operation. It answers: who did what to which resource, when, and what changed.

## Platform Audit Service

The audit service is a platform service — provided by `awo.so/internal/platform/audit`. Business modules call it; they do not implement their own audit storage.

```go
// Injected into every service constructor
type ContractService struct {
	auditSvc audit.Service
}
```

## What Gets Audited

| Operation | Audit required |
|-----------|---------------|
| Create contract | Yes |
| Update contract fields | Yes |
| Any status transition | Yes |
| Add/remove contract line | Yes |
| Soft delete | Yes |
| List/read | No |

## Audit Event Fields

```go
audit.Event{
	TenantID:     contract.TenantID,           // required: tenant scope
	EntityID:     contract.EntityID,           // required: org unit scope
	ActorID:      actorID,                     // required: who performed the action
	Action:       "contract.create",           // required: what happened
	ResourceType: "contract",                  // required: type of resource
	ResourceID:   contract.ID.String(),        // required: which resource
	Before:       beforeSnapshot,              // optional: state before change
	After:        afterSnapshot,               // optional: state after change
	Metadata:     map[string]string{...},      // optional: extra context
	OccurredAt:   time.Now(),                  // required: when
}
```

## Async Recording

```go
go func() {
	auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error().Interface("panic", r).Msg("audit goroutine panic")
		}
	}()
	if err := s.auditSvc.Record(auditCtx, evt); err != nil {
		s.logger.Error().Err(err).Str("action", evt.Action).Msg("audit record failed")
	}
}()
```

Use `context.Background()` — the request context is cancelled by the time the goroutine runs. The 5-second timeout prevents the goroutine from hanging indefinitely.

## Action Naming

`<resource_type>.<verb>` — matches the permission string verb where possible.

```
contract.create
contract.update
contract.submit
contract.approve
contract.activate
contract.suspend
contract.terminate
contract.delete
contract_line.create
contract_line.update
contract_line.delete
```
