---
title: Service Patterns
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Service Layer Overview](01-service-overview.md)"
  - "[Service Testing](03-service-testing.md)"
  - "[RLS and Tenant Isolation](../05-repository-layer/02-rls-and-tenant-isolation.md)"
---

# Service Patterns

Common patterns for service method implementations.

## List with Entity Scope

When a user's session has an EntityScope restriction, filter results to that scope:

```go
func (s *contractService) List(ctx context.Context, req ListContractsRequest) ([]*domain.Contract, int, error) {
    ctx, span := s.tracer.Start(ctx, "contract.service.list")
    defer span.End()

    allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
        Subject: req.Principal.Subject,
        Domain:  req.Principal.Domain,
        Object:  fmt.Sprintf("tenants/%s/contracts", req.TenantID),
        Action:  "contracts.contract.read",
    })
    if err != nil || !allowed {
        return nil, 0, iam.ErrForbidden
    }

    // EntityScope: if session is scoped to a subtree, filter by entity
    params := repository.ListContractsParams{
        TenantID: req.TenantID,
        Status:   req.Status,
        Limit:    req.Limit,
        Offset:   req.Offset,
    }
    if req.Principal.EntityScope != nil && !req.Principal.EntityScope.IsAll() {
        params.EntityID = &req.Principal.EntityScope.EntityID
    }

    contracts, total, err := s.repo.List(ctx, params)
    if err != nil {
        return nil, 0, err
    }
    return contracts, int(total), nil
}
```

## Update with Before/After Audit

Capture state before mutation for complete audit trail:

```go
func (s *contractService) Update(ctx context.Context, req UpdateContractRequest) (*domain.Contract, error) {
    ctx, span := s.tracer.Start(ctx, "contract.service.update")
    defer span.End()

    // 1. Authorize
    allowed, _ := s.authzSvc.Enforce(ctx, iam.Request{
        Subject: req.Principal.Subject,
        Domain:  req.Principal.Domain,
        Object:  fmt.Sprintf("tenants/%s/contracts/%s", req.TenantID, req.ID),
        Action:  "contracts.contract.update",
    })
    if !allowed {
        return nil, iam.ErrForbidden
    }

    // 2. Fetch before state (for audit)
    before, err := s.repo.GetByID(ctx, req.ID, req.TenantID)
    if err != nil {
        return nil, err
    }
    if !before.IsEditable() {
        return nil, domain.ErrContractNotEditable
    }

    // 3. Persist
    updated, err := s.repo.Update(ctx, repository.UpdateContractParams{
        ID:          req.ID,
        TenantID:    req.TenantID,
        Title:       req.Title,
        Description: req.Description,
        TotalValue:  req.TotalValue,
        Currency:    req.Currency,
        StartDate:   req.StartDate,
        EndDate:     req.EndDate,
        VendorID:    req.VendorID,
        AssignedTo:  req.AssignedTo,
        Version:     req.Version,
        UpdatedBy:   req.UserID,
    })
    if err != nil {
        return nil, err
    }

    // 4. Audit with before + after (async)
    s.recordAuditAsync(ctx, audit.Event{
        TenantID:     req.TenantID,
        Action:       "contract.updated",
        ResourceType: "contract",
        ResourceID:   req.ID.String(),
        ActorID:      req.UserID,
        Before:       marshalJSON(before),
        After:        marshalJSON(updated),
    })

    return updated, nil
}
```

## Conditional Notification

Only notify when the action is meaningful to recipients:

```go
func (s *contractService) Approve(ctx context.Context, req ApproveContractRequest) (*domain.Contract, error) {
    ctx, span := s.tracer.Start(ctx, "contract.service.approve")
    defer span.End()

    allowed, _ := s.authzSvc.Enforce(ctx, iam.Request{
        Subject: req.Principal.Subject,
        Domain:  req.Principal.Domain,
        Object:  fmt.Sprintf("tenants/%s/contracts/%s", req.TenantID, req.ContractID),
        Action:  "contracts.contract.approve",
    })
    if !allowed {
        return nil, iam.ErrForbidden
    }

    contract, err := s.repo.GetByID(ctx, req.ContractID, req.TenantID)
    if err != nil {
        return nil, err
    }
    if !contract.CanTransitionTo(domain.ContractStatusApproved) {
        return nil, domain.ErrContractInvalidTransition
    }

    updated, err := s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
        ID: req.ContractID, TenantID: req.TenantID,
        Status: domain.ContractStatusApproved, UpdatedBy: req.UserID,
    })
    if err != nil {
        return nil, err
    }

    // Notify submitter only — they care about approval
    s.publishNotificationAsync(port.Notification{
        TenantID:     req.TenantID,
        Audience:     port.AudienceUser,
        RecipientID:  contract.CreatedBy,   // submitter
        Category:     domain.NotifContractApproved,
        Title:        "Contract Approved",
        Body:         fmt.Sprintf("Contract %s has been approved.", updated.ContractNumber),
        ResourceID:   updated.ID,
        ResourceType: "contract",
        Priority:     port.PriorityNormal,
    })

    s.recordAuditAsync(ctx, audit.Event{
        TenantID: req.TenantID, Action: "contract.approved",
        ResourceType: "contract", ResourceID: req.ContractID.String(),
        ActorID: req.UserID, After: marshalJSON(updated),
    })

    s.publishAsync(ctx, domain.ContractApprovedEvent{
        ContractID: updated.ID, TenantID: req.TenantID,
        ApprovedBy: req.UserID, OccurredAt: time.Now().UTC(),
    })

    return updated, nil
}
```

## Batch Operation

Process a slice of IDs, collecting errors without aborting:

```go
func (s *contractService) BulkTerminate(
    ctx context.Context,
    req BulkTerminateRequest,
) (BulkResult, error) {
    ctx, span := s.tracer.Start(ctx, "contract.service.bulk_terminate")
    defer span.End()

    allowed, _ := s.authzSvc.Enforce(ctx, iam.Request{
        Subject: req.Principal.Subject,
        Domain:  req.Principal.Domain,
        Object:  fmt.Sprintf("tenants/%s/contracts", req.TenantID),
        Action:  "contracts.contract.terminate",
    })
    if !allowed {
        return BulkResult{}, iam.ErrForbidden
    }

    result := BulkResult{Total: len(req.ContractIDs)}
    for _, id := range req.ContractIDs {
        contract, err := s.repo.GetByID(ctx, id, req.TenantID)
        if err != nil {
            result.Errors = append(result.Errors, BulkError{ID: id, Err: err})
            continue
        }
        if !contract.CanTransitionTo(domain.ContractStatusTerminated) {
            result.Errors = append(result.Errors, BulkError{
                ID:  id,
                Err: domain.ErrContractInvalidTransition,
            })
            continue
        }
        _, err = s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
            ID: id, TenantID: req.TenantID,
            Status: domain.ContractStatusTerminated, UpdatedBy: req.UserID,
        })
        if err != nil {
            result.Errors = append(result.Errors, BulkError{ID: id, Err: err})
            continue
        }
        result.Succeeded++
        s.recordAuditAsync(ctx, audit.Event{
            TenantID: req.TenantID, Action: "contract.terminated",
            ResourceType: "contract", ResourceID: id.String(),
            ActorID: req.UserID,
        })
    }
    return result, nil
}
```

## Cross-Module Call via Interface

Never import another module's concrete package. Use an interface:

```go
// In contracts/service package:
// VendorService is the port for vendor lookups — injected via Wire
type VendorService interface {
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*VendorInfo, error)
}

func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
    // Validate vendor exists in this tenant
    vendor, err := s.vendorSvc.GetByID(ctx, req.VendorID, req.TenantID)
    if err != nil {
        if errors.Is(err, ErrVendorNotFound) {
            return nil, domain.NewBusinessError(400, "vendor not found", nil)
        }
        return nil, err
    }
    _ = vendor // use vendor.Name in contract creation if needed
    // ...
}
```

In `contracts/wire.go`:
```go
// Wire the interface to the concrete vendor service from the vendors module
wire.Bind(new(service.VendorService), new(*vendorservice.VendorServiceImpl)),
```

## Idempotent External Call

Use a stored idempotency key to prevent double-execution:

```go
func (s *contractService) RegisterWithExternal(ctx context.Context, contractID, tenantID uuid.UUID) error {
    contract, err := s.repo.GetByID(ctx, contractID, tenantID)
    if err != nil {
        return err
    }

    // Idempotency key: deterministic from contract ID
    idempotencyKey := fmt.Sprintf("contracts:external_register:%s", contractID)

    // Check if already registered (stored in contract metadata or a flags table)
    if contract.ExternallyRegistered {
        return nil // already done, safe to return
    }

    if err := s.externalClient.Register(ctx, external.RegisterRequest{
        IdempotencyKey: idempotencyKey,
        ContractNumber: contract.ContractNumber,
        TenantID:       tenantID.String(),
    }); err != nil {
        return fmt.Errorf("external registration failed: %w", err)
    }

    // Mark as registered to prevent future duplicates
    return s.repo.SetExternallyRegistered(ctx, contractID, tenantID)
}
```

## Helper: marshalJSON

```go
func marshalJSON(v any) json.RawMessage {
    b, _ := json.Marshal(v)
    return b
}
```
