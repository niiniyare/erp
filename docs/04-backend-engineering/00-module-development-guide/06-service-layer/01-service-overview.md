---
title: Service Layer Overview
portal: 4 — Backend Engineering
section: 00-module-development-guide/06-service-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-service-interface.md
    title: Service Interface
  - path: ./03-write-operations.md
    title: Write Operations
  - path: ./04-authorization.md
    title: Authorization
---

# Service Layer Overview

The service is the application layer — it orchestrates domain logic, persistence, and platform services. It is the only layer that holds business rules.

## Responsibilities

| Responsibility | Where in service |
|---------------|-----------------|
| Permission enforcement | First check in every method |
| Business rule validation | After permission, before repo call |
| State machine transitions | `entity.CanTransitionTo()` before `repo.UpdateStatus` |
| Optimistic lock conflict handling | Propagate `ErrContractConflict` to caller |
| Domain event publication | After successful repo write |
| Audit trail recording | After successful repo write |
| Notification dispatch | After successful repo write (async) |
| Feature flag gating | Before applying new behaviour |
| Settings lookup | When default behaviour is tenant-configurable |

## What Does NOT Go in the Service

| Thing | Correct location |
|-------|-----------------|
| SQL queries | `db/queries/<module>.sql` |
| HTTP status codes | `handlers/errors.go` |
| JSON serialisation | `handlers/response.go` |
| Request DTO parsing | `handlers/request.go` |
| SQLC types | `repository/<noun>_sqlc.go` |
| Session extraction from HTTP context | `handlers/<module>/handler.go` |

## Constructor Dependencies

```go
type contractService struct {
	repo     repository.ContractRepository
	lineRepo repository.ContractLineRepository
	authzSvc iam.AuthzService
	auditSvc audit.Service
	notifSvc notifications.Service
	eventBus events.Bus
	settings settings.Snapshot  // or passed per-call from session
	logger   logger.Logger
	tracer   tracing.Service
	metrics  metrics.MetricsProvider
}

func NewContractService(
	repo     repository.ContractRepository,
	lineRepo repository.ContractLineRepository,
	authzSvc iam.AuthzService,
	auditSvc audit.Service,
	notifSvc notifications.Service,
	eventBus events.Bus,
	logger   logger.Logger,
	tracer   tracing.Service,
	metrics  metrics.MetricsProvider,
) ContractService {
	return &contractService{
		repo:     repo,
		lineRepo: lineRepo,
		authzSvc: authzSvc,
		auditSvc: auditSvc,
		notifSvc: notifSvc,
		eventBus: eventBus,
		logger:   logger,
		tracer:   tracer,
		metrics:  metrics,
	}
}
```

Dependencies are injected via Wire. The service does not construct any of them.

## Standard Write Method Sequence

Every write method follows this exact order:

```
1. Start span (instrumentation)
2. Enforce permission (authzSvc.Enforce)
3. Validate business rules
4. If status change: check CanTransitionTo
5. Call repository method
6. Publish domain event (async goroutine)
7. Record audit event (async goroutine)
8. Send notification (async goroutine)
9. Return domain entity
```

Never reorder steps 2–4. Permission must be checked before reading current state for business rules. Business rules must be checked before attempting persistence.

## Session vs Principal

The service does not receive the full `ResolvedSession`. The handler extracts the session from `c.Locals`, derives the principal fields needed, and passes them explicitly to service methods.

```go
// Handler extracts and passes principal context explicitly
sess := c.Locals(domain.LocalsKeySession).(*iam.ResolvedSession)

contract, err := h.service.Create(ctx, service.CreateContractRequest{
	TenantID:  sess.TenantID,
	EntityID:  sess.EntityScope.EntityID,
	CreatedBy: sess.UserID,
	Principal: sess.ToPrincipal(),
	// ...
})
```

The service receives `Principal` (a simple Subject + Domain struct), not the full session. This decouples the service from the HTTP context and makes it testable without constructing a full session.
