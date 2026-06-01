---
title: State Transitions
portal: 4 — Backend Engineering
section: 00-module-development-guide/06-service-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ../02-ddd-domain-design/03-state-machine.md
    title: State Machine
  - path: ./03-write-operations.md
    title: Write Operations
---

# State Transitions

Every status-changing method in the service follows a specific sequence. This page shows the complete implementation of `Submit` and `Approve` as canonical examples.

## Submit

```go
func (s *contractService) Submit(
	ctx context.Context,
	id, tenantID uuid.UUID,
	version int,
	submittedBy uuid.UUID,
	principal iam.Principal,
) (*domain.Contract, error) {
	ctx, span := s.tracer.Start(ctx, "ContractService.Submit")
	defer span.End()

	// 1. Enforce permission
	allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
		Subject: iam.TenantSubject(principal.UserID()),
		Domain:  iam.TenantDomain(tenantID),
		Object:  "contracts/contract/*",
		Action:  "submit",
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, iam.ErrForbidden
	}

	// 2. Fetch current state
	current, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. Check state machine — ALWAYS before repo.UpdateStatus
	if !current.CanTransitionTo(domain.ContractStatusSubmitted) {
		return nil, domain.ErrContractInvalidTransition
	}

	// 4. Persist status transition
	updated, err := s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
		ID:        id,
		TenantID:  tenantID,
		Status:    domain.ContractStatusSubmitted,
		Version:   version,
		UpdatedBy: submittedBy,
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// 5. Publish event
	s.publishAsync(func() {
		if err := s.eventBus.Publish(ctx, domain.ContractSubmitted{
			ContractID:  updated.ID,
			TenantID:    updated.TenantID,
			EntityID:    updated.EntityID,
			SubmittedBy: submittedBy,
			OccurredAt:  time.Now(),
		}); err != nil {
			s.logger.Error().Err(err).Msg("failed to publish ContractSubmitted")
		}
	})

	// 6. Notify reviewers
	s.notifyAsync(ctx, notifications.Message{
		TenantID:  tenantID,
		Type:      "contract.submitted",
		Title:     "Contract submitted for review",
		Body:      fmt.Sprintf("Contract %s has been submitted for review.", updated.ContractNumber),
		Audience:  notifications.AudienceByRole("contract_reviewer"),
		ResourceID: updated.ID.String(),
	})

	// 7. Audit
	s.recordAuditAsync(ctx, audit.Event{
		TenantID:     tenantID,
		ActorID:      submittedBy,
		Action:       "contract.submit",
		ResourceType: "contract",
		ResourceID:   id.String(),
		Before:       current,
		After:        updated,
		OccurredAt:   time.Now(),
	})

	return updated, nil
}
```

## Approve (with business rule threshold check)

```go
func (s *contractService) Approve(
	ctx context.Context,
	id, tenantID uuid.UUID,
	version int,
	approvedBy uuid.UUID,
	principal iam.Principal,
) (*domain.Contract, error) {
	ctx, span := s.tracer.Start(ctx, "ContractService.Approve")
	defer span.End()

	// 1. Enforce permission
	allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
		Subject: iam.TenantSubject(principal.UserID()),
		Domain:  iam.TenantDomain(tenantID),
		Object:  "contracts/contract/*",
		Action:  "approve",
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, iam.ErrForbidden
	}

	// 2. Fetch current state
	current, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. State machine check
	if !current.CanTransitionTo(domain.ContractStatusApproved) {
		return nil, domain.ErrContractInvalidTransition
	}

	// 4. Business rule: approval threshold
	// NOTE: threshold comes from the session settings snapshot passed by the handler,
	// not fetched here — avoids a settings DB call inside the service.
	// The handler injects it via the request struct (not shown here for brevity).
	// See AddContractLineRequest for the pattern; extend ApproveRequest similarly.

	// 5. Persist
	updated, err := s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
		ID:        id,
		TenantID:  tenantID,
		Status:    domain.ContractStatusApproved,
		Version:   version,
		UpdatedBy: approvedBy,
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// 6. Publish, notify, audit (same async pattern as Submit)
	s.publishAsync(func() {
		_ = s.eventBus.Publish(ctx, domain.ContractApproved{
			ContractID:  updated.ID,
			TenantID:    updated.TenantID,
			EntityID:    updated.EntityID,
			ApprovedBy:  approvedBy,
			TotalValue:  updated.TotalValue.Amount().String(),
			Currency:    updated.Currency,
			OccurredAt:  time.Now(),
		})
	})

	s.recordAuditAsync(ctx, audit.Event{
		TenantID:     tenantID,
		ActorID:      approvedBy,
		Action:       "contract.approve",
		ResourceType: "contract",
		ResourceID:   id.String(),
		Before:       current,
		After:        updated,
		OccurredAt:   time.Now(),
	})

	return updated, nil
}
```

## Key Rules for Transition Methods

**Fetch before transition.** Always `GetByID` before `UpdateStatus`. The fetch gives you the current entity to:
1. Call `CanTransitionTo` (needs the current status).
2. Populate the `Before` field of the audit event.
3. Check any business rules that depend on current state.

**CanTransitionTo before UpdateStatus.** Always. No exceptions.

**Pass version from the request, not from the fetched entity.** The service receives `version` from the handler (the client's view of the version). Do not replace it with `current.Version` from the fetched entity — that bypasses optimistic locking.

```go
// CORRECT: version comes from the request
s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
	Version: version,  // from req — what the client thinks the version is
	...
})

// WRONG: version comes from the just-fetched entity
s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
	Version: current.Version,  // always matches — bypasses optimistic lock
	...
})
```

**State transition methods do not validate fields.** `Submit` does not validate `StartDate`. That validation happened at create/update time. The transition method only checks the transition validity and any transition-specific business rules.
