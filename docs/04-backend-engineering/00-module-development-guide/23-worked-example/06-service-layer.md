---
title: Worked Example — Service Layer
portal: 4 — Backend Engineering
section: 00-module-development-guide/23-worked-example
audience: [backend-engineer, tech-lead]
related:
  - path: ../06-service-layer/03-write-operations.md
    title: Write Operations
  - path: ../06-service-layer/05-state-transitions.md
    title: State Transitions
---

# Worked Example — Service Layer

## Constructor

```go
// internal/core/contracts/service/contract_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/rs/zerolog"
    "go.opentelemetry.io/otel/trace"

    "awo.so/internal/core/contracts/domain"
    "awo.so/internal/core/contracts/repository"
    iam "awo.so/internal/core/iam"
    "awo.so/internal/core/audit"
    "awo.so/internal/core/notifications/port"
    "awo.so/internal/platform/eventbus"
)

type contractService struct {
    repo      repository.ContractRepository
    lineRepo  repository.ContractLineRepository
    authzSvc  iam.AuthzService
    auditSvc  audit.AuditService
    notifSvc  port.NotificationService
    eventBus  eventbus.Publisher
    log       zerolog.Logger
    tracer    trace.Tracer
}

func NewContractService(
    repo      repository.ContractRepository,
    lineRepo  repository.ContractLineRepository,
    authzSvc  iam.AuthzService,
    auditSvc  audit.AuditService,
    notifSvc  port.NotificationService,
    eventBus  eventbus.Publisher,
    log       zerolog.Logger,
    tracer    trace.Tracer,
) ContractService {
    return &contractService{
        repo:     repo,
        lineRepo: lineRepo,
        authzSvc: authzSvc,
        auditSvc: auditSvc,
        notifSvc: notifSvc,
        eventBus: eventBus,
        log:      log.With().Str("service", "contracts").Logger(),
        tracer:   tracer,
    }
}
```

## Create

```go
func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
    ctx, span := s.tracer.Start(ctx, "contract.service.create")
    defer span.End()

    // 1. Authorize
    allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
        Subject: req.Principal.Subject,
        Domain:  req.Principal.Domain,
        Object:  fmt.Sprintf("tenants/%s/contracts", req.TenantID),
        Action:  "contracts.contract.create",
    })
    if err != nil || !allowed {
        return nil, iam.ErrForbidden
    }

    // 2. Persist
    contract, err := s.repo.Create(ctx, repository.CreateContractParams{
        TenantID:       req.TenantID,
        EntityID:       req.EntityID,
        ContractNumber: req.ContractNumber,
        Title:          req.Title,
        Description:    req.Description,
        ContractType:   req.ContractType,
        TotalValue:     req.TotalValue,
        Currency:       req.Currency,
        StartDate:      req.StartDate,
        EndDate:        req.EndDate,
        VendorID:       req.VendorID,
        AssignedTo:     req.AssignedTo,
        CreatedBy:      req.UserID,
    })
    if err != nil {
        return nil, err
    }

    // 3. Audit (async)
    s.recordAuditAsync(ctx, audit.Event{
        TenantID: req.TenantID, Action: "contract.created",
        ResourceType: "contract", ResourceID: contract.ID.String(),
        ActorID: req.UserID, After: toJSON(contract),
    })

    // 4. Publish event (async)
    s.publishAsync(ctx, domain.ContractCreatedEvent{
        ContractID: contract.ID, TenantID: req.TenantID,
        ContractNumber: contract.ContractNumber, EntityID: req.EntityID,
        CreatedBy: req.UserID, OccurredAt: time.Now().UTC(),
    })

    return contract, nil
}
```

## Submit (State Transition)

```go
func (s *contractService) Submit(ctx context.Context, req SubmitContractRequest) (*domain.Contract, error) {
    ctx, span := s.tracer.Start(ctx, "contract.service.submit")
    defer span.End()

    // 1. Authorize
    allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
        Subject: req.Principal.Subject,
        Domain:  req.Principal.Domain,
        Object:  fmt.Sprintf("tenants/%s/contracts/%s", req.TenantID, req.ContractID),
        Action:  "contracts.contract.submit",
    })
    if err != nil || !allowed {
        return nil, iam.ErrForbidden
    }

    // 2. Fetch current state
    contract, err := s.repo.GetByID(ctx, req.ContractID, req.TenantID)
    if err != nil {
        return nil, err
    }

    // 3. State machine check
    if !contract.CanTransitionTo(domain.ContractStatusSubmitted) {
        return nil, domain.ErrContractInvalidTransition
    }

    // 4. Persist transition
    updated, err := s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
        ID: req.ContractID, TenantID: req.TenantID,
        Status: domain.ContractStatusSubmitted, UpdatedBy: req.UserID,
    })
    if err != nil {
        return nil, err
    }

    // 5. Notify reviewers (async)
    s.publishNotificationAsync(buildSubmittedNotification(req.TenantID, req.ContractID, updated.ContractNumber))

    // 6. Audit (async)
    s.recordAuditAsync(ctx, audit.Event{
        TenantID: req.TenantID, Action: "contract.submitted",
        ResourceType: "contract", ResourceID: req.ContractID.String(),
        ActorID: req.UserID, Before: toJSON(contract), After: toJSON(updated),
    })

    // 7. Publish event (async)
    s.publishAsync(ctx, domain.ContractSubmittedEvent{
        ContractID: updated.ID, TenantID: req.TenantID,
        ContractNumber: updated.ContractNumber, SubmittedBy: req.UserID,
        OccurredAt: time.Now().UTC(),
    })

    return updated, nil
}
```

## Async Helpers

```go
func (s *contractService) recordAuditAsync(ctx context.Context, e audit.Event) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.log.Error().Interface("panic", r).Msg("audit goroutine panicked")
            }
        }()
        if err := s.auditSvc.Record(ctx, e); err != nil {
            s.log.Warn().Err(err).Str("action", e.Action).Msg("audit record failed")
        }
    }()
}

func (s *contractService) publishAsync(ctx context.Context, e domain.Event) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.log.Error().Interface("panic", r).Msg("event publish goroutine panicked")
            }
        }()
        payload, err := json.Marshal(e)
        if err != nil {
            s.log.Error().Err(err).Msg("failed to marshal event")
            return
        }
        if err := s.eventBus.Publish(ctx, eventbus.Event{
            ID: uuid.New().String(), Topic: e.Topic(),
            TenantID: e.GetTenantID().String(), PublishedAt: time.Now().UTC(),
            Payload: payload,
        }); err != nil {
            s.log.Warn().Err(err).Str("topic", e.Topic()).Msg("event publish failed")
        }
    }()
}

func (s *contractService) publishNotificationAsync(n port.Notification) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.log.Error().Interface("panic", r).Msg("notification goroutine panicked")
            }
        }()
        if err := s.notifSvc.Notify(ctx, n); err != nil {
            s.log.Warn().Err(err).Str("category", n.Category).Msg("notification failed")
        }
    }()
}

func toJSON(v interface{}) json.RawMessage {
    b, _ := json.Marshal(v)
    return b
}
```
