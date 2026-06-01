---
title: State Machine Design
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Domain Layer Overview](01-domain-overview.md)"
  - "[Service Layer](../06-service-layer/01-service-overview.md)"
  - "[Service Testing](../06-service-layer/03-service-testing.md)"
---

# State Machine Design

Every business entity with lifecycle (contracts, invoices, purchase orders) has a state machine. Define it in the domain layer.

## Pattern

```go
// domain/status.go

type ContractStatus string

const (
    StatusDraft       ContractStatus = "draft"
    StatusUnderReview ContractStatus = "under_review"
    StatusApproved    ContractStatus = "approved"
    StatusActive      ContractStatus = "active"
    StatusSuspended   ContractStatus = "suspended"
    StatusTerminated  ContractStatus = "terminated"
)

// AllStatuses used in validation and DB CHECK constraints
var AllStatuses = []ContractStatus{
    StatusDraft, StatusUnderReview, StatusApproved,
    StatusActive, StatusSuspended, StatusTerminated,
}

// Transitions defines the allowed state machine
var transitions = map[ContractStatus][]ContractStatus{
    StatusDraft:       {StatusUnderReview},
    StatusUnderReview: {StatusApproved, StatusDraft},  // approved or returned
    StatusApproved:    {StatusActive, StatusDraft},    // activated or returned
    StatusActive:      {StatusSuspended, StatusTerminated},
    StatusSuspended:   {StatusActive, StatusTerminated},
    StatusTerminated:  {},  // terminal — no further transitions
}

func (s ContractStatus) CanTransitionTo(next ContractStatus) bool {
    allowed, ok := transitions[s]
    if !ok {
        return false
    }
    for _, a := range allowed {
        if a == next {
            return true
        }
    }
    return false
}

// Helper methods for common checks
func (s ContractStatus) IsEditable() bool {
    return s == StatusDraft
}

func (s ContractStatus) IsTerminal() bool {
    return s == StatusTerminated
}

func (s ContractStatus) IsActive() bool {
    return s == StatusActive
}
```

## DB CHECK Constraint Alignment

The `CHECK` constraint in the migration must exactly match `AllStatuses`:

```sql
status text NOT NULL DEFAULT 'draft'
    CHECK (status IN ('draft','under_review','approved','active','suspended','terminated'))
```

When adding a new status:
1. Add to the Go constants
2. Add to `AllStatuses`
3. Add to the `transitions` map
4. Create a migration to update the `CHECK` constraint

## Service Layer Enforcement

```go
func (s *ContractService) Submit(ctx context.Context, sess iam.ResolvedSession, id uuid.UUID, version int) error {
    contract, err := s.repo.GetByID(ctx, id, sess.TenantID)
    if err != nil {
        return err
    }

    // State machine check — use domain method
    if !contract.Status.CanTransitionTo(domain.StatusUnderReview) {
        return domain.ErrContractNotEditable
    }

    _, err = s.repo.UpdateStatus(ctx, id, sess.TenantID, repository.UpdateStatusParams{
        Status:  domain.StatusUnderReview,
        Version: version,
    })
    return err
}
```

## SQL Guard

Add the state check to the UPDATE query as a DB-level guard:

```sql
-- name: SubmitContract :one
UPDATE contracts
SET status     = 'under_review',
    updated_at = now(),
    version    = version + 1
WHERE id        = @id
  AND version   = @version
  AND status    = 'draft'           -- DB-level state guard
  AND deleted_at IS NULL
RETURNING *;
```

If the service check passes but a concurrent request changed the status between GET and UPDATE, the DB query returns no rows → `ErrVersionConflict`. Double safety.

## Testing the State Machine

```go
func TestContractStatus_Transitions(t *testing.T) {
    matrix := map[domain.ContractStatus]struct {
        allowed []domain.ContractStatus
        denied  []domain.ContractStatus
    }{
        domain.StatusDraft: {
            allowed: []domain.ContractStatus{domain.StatusUnderReview},
            denied:  []domain.ContractStatus{domain.StatusActive, domain.StatusApproved, domain.StatusTerminated},
        },
        domain.StatusUnderReview: {
            allowed: []domain.ContractStatus{domain.StatusApproved, domain.StatusDraft},
            denied:  []domain.ContractStatus{domain.StatusActive, domain.StatusTerminated},
        },
        domain.StatusTerminated: {
            allowed: []domain.ContractStatus{},
            denied:  []domain.ContractStatus{domain.StatusDraft, domain.StatusActive},
        },
        // ... all statuses
    }

    for from, tc := range matrix {
        for _, to := range tc.allowed {
            assert.True(t, from.CanTransitionTo(to),
                "%s → %s should be allowed", from, to)
        }
        for _, to := range tc.denied {
            assert.False(t, from.CanTransitionTo(to),
                "%s → %s should be denied", from, to)
        }
    }
}
```

## State Diagram Reference

```
draft ──────────────────────────────────────► under_review
  ▲                                                 │
  │ (returned)                          (approved)  │
  │◄────────────── under_review ◄──────────────────-┘
  │                     │
  │ (returned)          ▼
  │◄──────────── approved ─────────────────────► active
                                                   │    │
                                       (suspended) ▼    │ (terminated)
                                              suspended  │
                                                   │     │
                                       (activated) ▼     ▼
                                               active  terminated
                                                              (terminal)
```
