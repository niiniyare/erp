---
title: Sending Notifications
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[Notification Types](02-notification-types.md)"
  - "[Service Layer Overview](../06-service-layer/01-service-overview.md)"
---

# Sending Notifications

## Integration in Service Methods

Call `publishNotificationAsync` after every successful state transition. Never block the response waiting for notification delivery.

```go
// internal/core/contracts/service/contract_service.go

func (s *contractService) Submit(
    ctx context.Context,
    contractID, tenantID uuid.UUID,
    version int,
    userID uuid.UUID,
    principal iam.Principal,
) (*domain.Contract, error) {
    span := s.tracer.StartSpan(ctx, "contract.service.submit")
    defer span.End()

    // 1. Authorize
    allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
        Subject: principal.Subject,
        Domain:  principal.Domain,
        Object:  fmt.Sprintf("tenants/%s/contracts/%s", tenantID, contractID),
        Action:  "contracts.contract.submit",
    })
    if err != nil || !allowed {
        return nil, iam.ErrForbidden
    }

    // 2. Fetch
    contract, err := s.repo.GetByID(ctx, contractID, tenantID)
    if err != nil {
        return nil, err
    }

    // 3. State machine check
    if !contract.CanTransitionTo(domain.ContractStatusSubmitted) {
        return nil, domain.ErrContractInvalidTransition
    }

    // 4. Persist
    updated, err := s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
        ID:        contractID,
        TenantID:  tenantID,
        Status:    domain.ContractStatusSubmitted,
        Version:   version,
        UpdatedBy: userID,
    })
    if err != nil {
        return nil, err
    }

    // 5. Notify reviewers — async, non-blocking
    s.publishNotificationAsync(
        buildSubmittedNotification(tenantID, contractID, updated.ContractNumber),
    )

    // 6. Audit — async
    s.recordAuditAsync(ctx, audit.Event{
        TenantID:     tenantID,
        Action:       "contract.submitted",
        ResourceType: "contract",
        ResourceID:   contractID.String(),
        ActorID:      userID,
        Before:       toJSON(contract),
        After:        toJSON(updated),
    })

    return updated, nil
}
```

## publishNotificationAsync Helper

```go
func (s *contractService) publishNotificationAsync(n port.Notification) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() {
            if r := recover(); r != nil {
                s.log.Error().
                    Interface("panic", r).
                    Str("category", n.Category).
                    Msg("notification goroutine panicked")
            }
        }()

        if err := s.notifSvc.Notify(ctx, n); err != nil {
            s.log.Warn().
                Err(err).
                Str("category", n.Category).
                Str("tenant_id", n.TenantID.String()).
                Msg("notification delivery failed — non-fatal")
            // Do NOT return error. Notification failure is non-fatal.
        }
    }()
}
```

## Conditional Notifications (Feature Flag)

Some notification categories can be toggled per-tenant. Check the feature flag before publishing:

```go
// In handler — extract flag from session before calling service
session := middleware.SessionFrom(c)
notifyOnSubmit := session.FeatureEnabled("contracts.notify_on_submit")

result, err := s.contractSvc.Submit(ctx, service.SubmitContractRequest{
    ContractID:           contractID,
    TenantID:             tenantID,
    Version:              req.Version,
    UserID:               session.UserID,
    Principal:            session.ToPrincipal(),
    NotifyOnSubmit:       notifyOnSubmit,
})
```

```go
// In service — respect the flag
func (s *contractService) Submit(ctx context.Context, req SubmitContractRequest) (*domain.Contract, error) {
    // ... state machine, persist ...

    if req.NotifyOnSubmit {
        s.publishNotificationAsync(buildSubmittedNotification(...))
    }

    return updated, nil
}
```

## Testing Notification Publishing

In service unit tests, assert that `Notify` was called (or not called) on the mock:

```go
func TestContractService_Submit_NotifieReviewers(t *testing.T) {
    repo := &mockContractRepo{}
    notifSvc := &mockNotifSvc{}

    authz := &mockAuthzService{}
    authz.On("Enforce", mock.Anything, mock.Anything).Return(true, nil)

    repo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).
        Return(&domain.Contract{
            Status:  domain.ContractStatusDraft,
            Version: 1,
        }, nil)
    repo.On("UpdateStatus", mock.Anything, mock.Anything).
        Return(&domain.Contract{
            Status:         domain.ContractStatusSubmitted,
            ContractNumber: "CONT-2025-0001",
        }, nil)

    notifSvc.On("Notify", mock.Anything, mock.MatchedBy(func(n port.Notification) bool {
        return n.Category == domain.NotifContractSubmitted
    })).Return(nil)

    svc := buildService(repo, notifSvc, authz)
    _, err := svc.Submit(context.Background(), SubmitContractRequest{
        NotifyOnSubmit: true,
        // ...
    })

    require.NoError(t, err)

    // Give goroutine time to run — or make notifSvc.Notify synchronous in tests
    time.Sleep(10 * time.Millisecond)
    notifSvc.AssertCalled(t, "Notify", mock.Anything, mock.Anything)
}
```

Prefer injecting a synchronous test double over `time.Sleep` — use a channel or `sync.WaitGroup` in the mock to signal completion:

```go
type syncNotifSvc struct {
    calls []port.Notification
    mu    sync.Mutex
}

func (m *syncNotifSvc) Notify(_ context.Context, n port.Notification) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.calls = append(m.calls, n)
    return nil
}
```
