---
title: Service Patterns
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Service Layer Overview](01-service-overview.md)"
  - "[Repository Layer](../07-repository-layer/01-repository-overview.md)"
  - "[Error Handling](../17-error-handling/01-error-handling-overview.md)"
---

# Service Patterns

## List with Filtering

```go
func (s *ContractService) List(ctx context.Context, sess iam.ResolvedSession, params ListParams) ([]domain.Contract, int64, error) {
    if err := s.authz.Can(ctx, sess.ToPrincipal(), "contracts.contract.read"); err != nil {
        return nil, 0, domain.ErrForbidden
    }

    repoParams := repository.ListParams{
        Status: params.Status,
        Limit:  params.Limit,
        Offset: params.Offset,
    }

    // Apply entity scope if user doesn't have full visibility
    if sess.EntityScope.Type != iam.ScopeAll {
        repoParams.EntityID = sess.EntityID()
        repoParams.ScopeType = sess.EntityScope.Type
    }

    return s.repo.List(ctx, sess.TenantID, repoParams)
}
```

## Update with Before/After Audit

```go
func (s *ContractService) Update(ctx context.Context, sess iam.ResolvedSession, id uuid.UUID, req UpdateParams) (*domain.Contract, error) {
    if err := s.authz.Can(ctx, sess.ToPrincipal(), "contracts.contract.update"); err != nil {
        return nil, domain.ErrForbidden
    }

    // Capture before state
    before, err := s.repo.GetByID(ctx, id, sess.TenantID)
    if err != nil {
        return nil, err
    }
    if !before.Status.CanEdit() {
        return nil, domain.ErrContractNotEditable
    }

    after, err := s.repo.Update(ctx, id, sess.TenantID, mapToUpdateParams(req))
    if err != nil {
        return nil, err
    }

    go s.recordAuditAsync(sess, "contracts.contract.update", id, before, after)
    return after, nil
}
```

## Conditional Notification

```go
func (s *ContractService) Approve(ctx context.Context, sess iam.ResolvedSession, id uuid.UUID, version int, comment string) error {
    // ...approval logic...

    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        defer func() { recover() }()

        // Notify contract owner
        if err := s.notif.Send(ctx, notifications.Notification{
            TenantID:   sess.TenantID,
            UserID:     contract.OwnerID,           // recipient, not approver
            Category:   notifications.CategoryContractApproved,
            Title:      "Contract Approved",
            Body:       fmt.Sprintf("Your contract %s has been approved.", contract.ContractNumber),
            ResourceID: contract.ID,
        }); err != nil {
            s.logger.Error("notification failed", "error", err)
        }
    }()
    return nil
}
```

## Batch Operation

```go
func (s *ContractService) BulkTerminate(ctx context.Context, sess iam.ResolvedSession, ids []uuid.UUID, reason string) (*BulkResult, error) {
    if err := s.authz.Can(ctx, sess.ToPrincipal(), "contracts.contract.terminate"); err != nil {
        return nil, domain.ErrForbidden
    }
    if !sess.FeatureEnabled("contracts.bulk_operations") {
        return nil, &domain.BusinessError{
            Code: "FEATURE_NOT_ENABLED", Message: "bulk operations not enabled", Status: 403,
        }
    }

    result := &BulkResult{}
    for _, id := range ids {
        if err := s.terminate(ctx, sess, id, reason); err != nil {
            result.Failed = append(result.Failed, BulkError{ID: id, Error: err.Error()})
        } else {
            result.Succeeded = append(result.Succeeded, id)
        }
    }
    return result, nil
}
```

## Cross-Module Call via Interface

```go
type EntityLookup interface {
    GetEntityName(ctx context.Context, tenantID, entityID uuid.UUID) (string, error)
}

type ContractService struct {
    repo     repository.ContractRepository
    entities EntityLookup   // injected via Wire — provided by entities module
    // ...
}

func (s *ContractService) GetByID(ctx context.Context, sess iam.ResolvedSession, id uuid.UUID) (*ContractDetail, error) {
    contract, err := s.repo.GetByID(ctx, id, sess.TenantID)
    if err != nil {
        return nil, err
    }

    // Enrich with entity name — cross-module lookup
    entityName, err := s.entities.GetEntityName(ctx, sess.TenantID, contract.EntityID)
    if err != nil {
        s.logger.WarnContext(ctx, "entity lookup failed", "entity_id", contract.EntityID, "error", err)
        entityName = ""   // degrade gracefully
    }

    return &ContractDetail{Contract: contract, EntityName: entityName}, nil
}
```

## Idempotent External Call

```go
func (s *ContractService) SendToExternalSystem(ctx context.Context, contractID uuid.UUID) error {
    // Check if already sent
    contract, _ := s.repo.GetByID(ctx, contractID, tenantID)
    if contract.ExternalID != "" {
        return nil   // already sent — idempotent
    }

    externalID, err := s.externalClient.Submit(ctx, mapToExternalRequest(contract))
    if err != nil {
        return fmt.Errorf("external submit: %w", err)
    }

    // Persist the external ID
    return s.repo.SetExternalID(ctx, contractID, tenantID, externalID)
}
```
