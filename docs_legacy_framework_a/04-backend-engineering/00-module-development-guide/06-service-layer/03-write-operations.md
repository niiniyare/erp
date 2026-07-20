> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Write Operations
portal: 4 — Backend Engineering
section: 00-module-development-guide/06-service-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./04-authorization.md
    title: Authorization
  - path: ./05-state-transitions.md
    title: State Transitions
---

# Write Operations

This page shows the complete implementation of the core write operations: `Create`, `Update`, and `Submit` (status transition). The pattern is the same for all write methods.

## Create

```go
// internal/core/contracts/service/contract.go
func (s *contractService) Create(ctx context.Context, req CreateContractRequest) (*domain.Contract, error) {
	// 1. Start span
	ctx, span := s.tracer.Start(ctx, "ContractService.Create")
	defer span.End()

	// 2. Enforce permission
	allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
		Subject: iam.TenantSubject(req.Principal.UserID()),
		Domain:  iam.TenantDomain(req.TenantID),
		Object:  "contracts/contract/*",
		Action:  "create",
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, iam.ErrForbidden
	}

	// 3. Parse and validate input
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}
	if !endDate.After(startDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	// 4. Business rule: check for duplicate contract number
	exists, err := s.repo.ExistsWithNumber(ctx, req.ContractNumber, req.TenantID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrContractAlreadyExists
	}

	// 5. Persist
	contract, err := s.repo.Create(ctx, repository.CreateContractParams{
		TenantID:       req.TenantID,
		EntityID:       req.EntityID,
		ContractNumber: req.ContractNumber,
		Title:          req.Title,
		Description:    req.Description,
		VendorID:       req.VendorID,
		ContractType:   req.ContractType,
		StartDate:      startDate,
		EndDate:        endDate,
		TotalValue:     decimal.Zero,  // starts at zero; lines added separately
		Currency:       req.Currency,
		CreatedBy:      req.CreatedBy,
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// 6. Set span attributes after successful write
	span.SetAttributes(
		attribute.String("contract.id", contract.ID.String()),
		attribute.String("tenant.id", contract.TenantID.String()),
	)

	// 7. Publish domain event (async — never block response)
	s.publishAsync(func() {
		evt := domain.ContractCreated{
			ContractID:     contract.ID,
			TenantID:       contract.TenantID,
			EntityID:       contract.EntityID,
			ContractNumber: contract.ContractNumber,
			CreatedBy:      contract.CreatedBy,
			OccurredAt:     time.Now(),
		}
		if err := s.eventBus.Publish(ctx, evt); err != nil {
			s.logger.Error().Err(err).Msg("failed to publish ContractCreated")
		}
	})

	// 8. Record audit (async)
	s.recordAuditAsync(ctx, audit.Event{
		TenantID:     contract.TenantID,
		EntityID:     contract.EntityID,
		ActorID:      contract.CreatedBy,
		Action:       "contract.create",
		ResourceType: "contract",
		ResourceID:   contract.ID.String(),
		After:        contract,
		OccurredAt:   time.Now(),
	})

	// 9. Increment metric
	s.metrics.IncrementCounter("contracts.created", 1)

	return contract, nil
}
```

## Update

```go
func (s *contractService) Update(ctx context.Context, req UpdateContractRequest) (*domain.Contract, error) {
	ctx, span := s.tracer.Start(ctx, "ContractService.Update")
	defer span.End()

	// 1. Enforce permission
	allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
		Subject: iam.TenantSubject(req.Principal.UserID()),
		Domain:  iam.TenantDomain(req.TenantID),
		Object:  "contracts/contract/*",
		Action:  "update",
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, iam.ErrForbidden
	}

	// 2. Fetch current state for business rule checks
	current, err := s.repo.GetByID(ctx, req.ID, req.TenantID)
	if err != nil {
		return nil, err
	}

	// 3. Business rule: only draft contracts are editable
	if !current.IsEditable() {
		return nil, domain.ErrContractNotEditable
	}

	// 4. Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	// 5. Persist (optimistic lock via Version)
	updated, err := s.repo.Update(ctx, repository.UpdateContractParams{
		ID:           req.ID,
		TenantID:     req.TenantID,
		Title:        req.Title,
		Description:  req.Description,
		VendorID:     req.VendorID,
		ContractType: req.ContractType,
		StartDate:    startDate,
		EndDate:      endDate,
		Currency:     req.Currency,
		TotalValue:   current.TotalValue.Amount(),  // preserve existing total
		Version:      req.Version,
		UpdatedBy:    req.UpdatedBy,
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// 6. Async: audit + event
	s.recordAuditAsync(ctx, audit.Event{
		TenantID:     updated.TenantID,
		ActorID:      updated.UpdatedBy,
		Action:       "contract.update",
		ResourceType: "contract",
		ResourceID:   updated.ID.String(),
		Before:       current,
		After:        updated,
		OccurredAt:   time.Now(),
	})

	return updated, nil
}
```

## publishAsync Helper

```go
// publishAsync runs fn in a goroutine with a background context and timeout.
// Errors are logged; never block the caller.
func (s *contractService) publishAsync(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error().Interface("panic", r).Msg("recovered from panic in async publish")
			}
		}()
		fn()
	}()
}

// recordAuditAsync fires an audit event asynchronously.
func (s *contractService) recordAuditAsync(ctx context.Context, evt audit.Event) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error().Interface("panic", r).Msg("recovered from panic in audit recording")
			}
		}()
		auditCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.auditSvc.Record(auditCtx, evt); err != nil {
			s.logger.Error().Err(err).Str("action", evt.Action).Msg("failed to record audit event")
		}
	}()
}
```

The deferred `recover()` in async goroutines ensures a panic in audit/event publishing does not crash the server. The response has already been sent before the goroutine runs.
