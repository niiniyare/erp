---
title: Audit Integration
portal: 4 — Backend Engineering
section: 00-module-development-guide/07-core-integrations
audience: [backend-engineer, tech-lead]
related:
  - path: ./06-notification-integration.md
    title: Notification Integration
---

# Audit Integration

Every state-changing operation records an audit event. Audit events are immutable records that capture who did what to which resource and when, with before/after state snapshots.

## Audit Event Structure

```go
// Platform audit.Event — passed to auditSvc.Record()
type Event struct {
	TenantID     uuid.UUID
	EntityID     uuid.UUID
	ActorID      uuid.UUID
	Action       string       // "contract.create", "contract.approve", etc.
	ResourceType string       // "contract"
	ResourceID   string       // contract UUID as string
	Before       interface{}  // serialised to JSON — current state before change
	After        interface{}  // serialised to JSON — new state after change
	Metadata     map[string]string  // optional extra context
	OccurredAt   time.Time
}
```

## Calling the Audit Service

Always call audit asynchronously — never block the response:

```go
go func() {
	auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.auditSvc.Record(auditCtx, audit.Event{
		TenantID:     contract.TenantID,
		EntityID:     contract.EntityID,
		ActorID:      actorID,
		Action:       "contract.create",
		ResourceType: "contract",
		ResourceID:   contract.ID.String(),
		After:        contract,
		OccurredAt:   time.Now(),
	}); err != nil {
		s.logger.Error().Err(err).Msg("audit record failed")
	}
}()
```

Use `context.Background()` for the audit goroutine — not the request context. The request context is cancelled when the response is sent. The audit write must complete after the response.

## Action Name Convention

Format: `<module>.<verb>` — lowercase, dot-separated.

| Operation | Action name |
|-----------|------------|
| Create contract | `contract.create` |
| Update contract | `contract.update` |
| Submit contract | `contract.submit` |
| Approve contract | `contract.approve` |
| Activate contract | `contract.activate` |
| Suspend contract | `contract.suspend` |
| Terminate contract | `contract.terminate` |
| Delete contract | `contract.delete` |
| Add line | `contract_line.create` |
| Remove line | `contract_line.delete` |

## Before and After State

For write operations, pass both `Before` (the entity before the change) and `After` (the entity after the change). The audit service serialises them to JSON.

For creates, `Before` is nil (nothing existed before). For deletes, `After` is the deleted entity (preserving the final state).

```go
// Create: no Before
audit.Event{
	Before: nil,
	After:  contract,  // newly created
}

// Update: both Before and After
audit.Event{
	Before: currentContract,   // state before update
	After:  updatedContract,   // state after update
}

// Delete: After is the deleted entity
audit.Event{
	Before: contract,  // state at time of deletion
	After:  nil,
}
```

## What Requires Auditing

| Operation type | Audit required |
|---------------|----------------|
| Create | Yes |
| Update (any field) | Yes |
| Status transition | Yes |
| Delete | Yes |
| Read (list/detail) | No (generates too much noise) |
| Failed attempts | No (captured in structured logs, not audit trail) |

Some regulated modules (finance, HR) may require read auditing. Check with the compliance team before adding read audit events to a module.
