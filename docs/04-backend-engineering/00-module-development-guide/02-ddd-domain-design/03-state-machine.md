---
title: State Machine
portal: 4 — Backend Engineering
section: 00-module-development-guide/02-ddd-domain-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-entity-design.md
    title: Entity Design
  - path: ./05-domain-errors.md
    title: Domain Errors
---

# State Machine

Every AwoERP entity with a `status` field has a state machine enforced by `CanTransitionTo()`. This page explains the design rules and the contracts state machine in full.

## Why CanTransitionTo on the Domain Entity

The state machine lives on the entity — not in the service, not in the database — because:

1. **It is domain logic, not infrastructure logic.** It belongs with the data it guards.
2. **It is testable without a database.** A unit test can call `entity.CanTransitionTo()` directly.
3. **It is the single source of truth.** Service code never hard-codes `if status == "draft"` checks — it always calls `CanTransitionTo`.
4. **It is visible alongside the status constants.** A reader sees all valid states and all transitions in one file.

## Contracts State Machine

```
draft ──────────────────────────────────────────────────►  submitted
  ▲                                                              │
  │ (reject / revise)                                            │
  └──────────────────────────────────────────────────────  under_review
                                                                 │
                                                                 ▼
                                                            approved
                                                                 │
                                                                 ▼
                                                            active ◄──────── suspended
                                                            │    │               ▲
                                                            │    └───────────────┘
                                                            │
                                                            ▼
                                                         terminated
```

### Transition Table

| From | To | Business trigger |
|------|----|-----------------|
| `draft` | `submitted` | User submits for review |
| `submitted` | `under_review` | Reviewer picks up the contract |
| `submitted` | `draft` | Returned for revision |
| `under_review` | `approved` | Reviewer approves |
| `under_review` | `draft` | Returned for revision |
| `approved` | `active` | Contract start date reached or manual activation |
| `approved` | `draft` | Reset before activation (edge case) |
| `active` | `suspended` | Temporary hold |
| `active` | `terminated` | Contract ended |
| `suspended` | `active` | Hold lifted |
| `suspended` | `terminated` | Terminated while on hold |
| `terminated` | _(none)_ | Terminal state — no exits |

### Implementation

```go
// CanTransitionTo reports whether transitioning from the current status to next
// is a valid state machine move.
func (c *Contract) CanTransitionTo(next ContractStatus) bool {
	allowed := map[ContractStatus][]ContractStatus{
		ContractStatusDraft:       {ContractStatusSubmitted},
		ContractStatusSubmitted:   {ContractStatusUnderReview, ContractStatusDraft},
		ContractStatusUnderReview: {ContractStatusApproved, ContractStatusDraft},
		ContractStatusApproved:    {ContractStatusActive, ContractStatusDraft},
		ContractStatusActive:      {ContractStatusSuspended, ContractStatusTerminated},
		ContractStatusSuspended:   {ContractStatusActive, ContractStatusTerminated},
		ContractStatusTerminated:  {}, // terminal state
	}
	for _, s := range allowed[c.Status] {
		if s == next {
			return true
		}
	}
	return false
}
```

### How the Service Uses It

```go
// internal/core/contracts/service/contract.go
func (s *contractService) Submit(ctx context.Context, id, tenantID uuid.UUID, updatedBy uuid.UUID) (*domain.Contract, error) {
	// 1. Fetch current state
	contract, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	// 2. Check state machine — always before calling repo.UpdateStatus
	if !contract.CanTransitionTo(domain.ContractStatusSubmitted) {
		return nil, domain.ErrContractInvalidTransition
	}

	// 3. Persist the transition
	return s.repo.UpdateStatus(ctx, repository.UpdateContractStatusParams{
		ID:        id,
		TenantID:  tenantID,
		Status:    domain.ContractStatusSubmitted,
		UpdatedBy: updatedBy,
		Version:   contract.Version,
	})
}
```

The service **never** reads the current status and decides the transition itself. It always fetches the entity, calls `CanTransitionTo`, then calls the repo. This ensures the state machine in the domain file is the only place transition logic lives.

## Testing the State Machine

State machine tests are pure unit tests — no database, no mocks:

```go
// internal/core/contracts/domain/contract_test.go
package domain_test

import (
	"testing"

	"awo.so/internal/core/contracts/domain"
)

func TestContractCanTransitionTo(t *testing.T) {
	tests := []struct {
		from    domain.ContractStatus
		to      domain.ContractStatus
		allowed bool
	}{
		{domain.ContractStatusDraft, domain.ContractStatusSubmitted, true},
		{domain.ContractStatusDraft, domain.ContractStatusActive, false},
		{domain.ContractStatusDraft, domain.ContractStatusTerminated, false},
		{domain.ContractStatusSubmitted, domain.ContractStatusUnderReview, true},
		{domain.ContractStatusSubmitted, domain.ContractStatusDraft, true},
		{domain.ContractStatusSubmitted, domain.ContractStatusActive, false},
		{domain.ContractStatusApproved, domain.ContractStatusActive, true},
		{domain.ContractStatusApproved, domain.ContractStatusDraft, true},
		{domain.ContractStatusApproved, domain.ContractStatusTerminated, false},
		{domain.ContractStatusActive, domain.ContractStatusSuspended, true},
		{domain.ContractStatusActive, domain.ContractStatusTerminated, true},
		{domain.ContractStatusActive, domain.ContractStatusDraft, false},
		{domain.ContractStatusTerminated, domain.ContractStatusDraft, false},
		{domain.ContractStatusTerminated, domain.ContractStatusActive, false},
	}

	for _, tt := range tests {
		c := &domain.Contract{Status: tt.from}
		got := c.CanTransitionTo(tt.to)
		if got != tt.allowed {
			t.Errorf("CanTransitionTo(%q → %q) = %v, want %v", tt.from, tt.to, got, tt.allowed)
		}
	}
}
```

This test must cover **every row** in the transition table plus representative invalid transitions from each state. Full matrix coverage ensures no regression when transitions are added or removed.

## Common Mistakes

**Using switch instead of map:**

```go
// FRAGILE — easy to forget a case; no exhaustiveness check
func (c *Contract) CanTransitionTo(next ContractStatus) bool {
    switch c.Status {
    case ContractStatusDraft:
        return next == ContractStatusSubmitted
    // ...
    }
    return false
}
```

The map pattern is preferred because it is data — readable as a table, easy to diff in code review.

**Checking status in the service instead of delegating to entity:**

```go
// WRONG — business rule in wrong layer
func (s *contractService) Submit(ctx context.Context, ...) error {
    if contract.Status != domain.ContractStatusDraft {
        return domain.ErrContractInvalidTransition
    }
    // ...
}
```

This bypasses `CanTransitionTo` and hard-codes a transition rule in the service. When a new transition is added to the map, this service check does not update automatically.

**Not calling CanTransitionTo before UpdateStatus:**

```go
// WRONG — missing guard
func (s *contractService) Submit(ctx context.Context, ...) error {
    // goes straight to repo — no state check
    return s.repo.UpdateStatus(ctx, ...)
}
```

Without the guard, any status can be written to any status — the state machine is effectively disabled.
