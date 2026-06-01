---
title: Domain Tests
portal: 4 — Backend Engineering
section: 00-module-development-guide/22-testing-guide
audience: [backend-engineer, tech-lead]
related:
  - path: ./01-testing-overview.md
    title: Testing Overview
  - path: ../02-ddd-domain-design/03-state-machine.md
    title: State Machine
---

# Domain Tests

Domain tests are pure unit tests — zero external dependencies, no mocks, no database. They test state machine transitions, value object validation, and entity helper methods.

## State Machine Test (full matrix)

```go
// internal/core/contracts/domain/contract_test.go
package domain_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"awo.so/internal/core/contracts/domain"
)

func TestContractCanTransitionTo(t *testing.T) {
	// Full transition matrix — every valid and invalid combination
	tests := []struct {
		name    string
		from    domain.ContractStatus
		to      domain.ContractStatus
		allowed bool
	}{
		// Draft transitions
		{"draft → submitted (allowed)", domain.ContractStatusDraft, domain.ContractStatusSubmitted, true},
		{"draft → active (blocked)",    domain.ContractStatusDraft, domain.ContractStatusActive, false},
		{"draft → terminated (blocked)",domain.ContractStatusDraft, domain.ContractStatusTerminated, false},
		{"draft → draft (blocked)",     domain.ContractStatusDraft, domain.ContractStatusDraft, false},

		// Submitted transitions
		{"submitted → under_review (allowed)", domain.ContractStatusSubmitted, domain.ContractStatusUnderReview, true},
		{"submitted → draft (allowed)",        domain.ContractStatusSubmitted, domain.ContractStatusDraft, true},
		{"submitted → approved (blocked)",     domain.ContractStatusSubmitted, domain.ContractStatusApproved, false},
		{"submitted → active (blocked)",       domain.ContractStatusSubmitted, domain.ContractStatusActive, false},

		// Under review transitions
		{"under_review → approved (allowed)", domain.ContractStatusUnderReview, domain.ContractStatusApproved, true},
		{"under_review → draft (allowed)",    domain.ContractStatusUnderReview, domain.ContractStatusDraft, true},
		{"under_review → active (blocked)",   domain.ContractStatusUnderReview, domain.ContractStatusActive, false},

		// Approved transitions
		{"approved → active (allowed)",  domain.ContractStatusApproved, domain.ContractStatusActive, true},
		{"approved → draft (allowed)",   domain.ContractStatusApproved, domain.ContractStatusDraft, true},
		{"approved → submitted (blocked)", domain.ContractStatusApproved, domain.ContractStatusSubmitted, false},

		// Active transitions
		{"active → suspended (allowed)",   domain.ContractStatusActive, domain.ContractStatusSuspended, true},
		{"active → terminated (allowed)",  domain.ContractStatusActive, domain.ContractStatusTerminated, true},
		{"active → draft (blocked)",       domain.ContractStatusActive, domain.ContractStatusDraft, false},
		{"active → approved (blocked)",    domain.ContractStatusActive, domain.ContractStatusApproved, false},

		// Suspended transitions
		{"suspended → active (allowed)",     domain.ContractStatusSuspended, domain.ContractStatusActive, true},
		{"suspended → terminated (allowed)", domain.ContractStatusSuspended, domain.ContractStatusTerminated, true},
		{"suspended → draft (blocked)",      domain.ContractStatusSuspended, domain.ContractStatusDraft, false},

		// Terminated (terminal state)
		{"terminated → draft (blocked)",    domain.ContractStatusTerminated, domain.ContractStatusDraft, false},
		{"terminated → active (blocked)",   domain.ContractStatusTerminated, domain.ContractStatusActive, false},
		{"terminated → submitted (blocked)",domain.ContractStatusTerminated, domain.ContractStatusSubmitted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &domain.Contract{Status: tt.from}
			got := c.CanTransitionTo(tt.to)
			assert.Equal(t, tt.allowed, got,
				"CanTransitionTo(%q → %q): expected %v, got %v",
				tt.from, tt.to, tt.allowed, got)
		})
	}
}
```

## Value Object Tests

```go
// internal/core/contracts/domain/value_objects_test.go
package domain_test

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"awo.so/internal/core/contracts/domain"
)

func TestNewContractValue(t *testing.T) {
	t.Run("positive value is valid", func(t *testing.T) {
		v, err := domain.NewContractValue(decimal.NewFromFloat(100.50))
		require.NoError(t, err)
		assert.Equal(t, "100.5", v.Amount().String())
	})

	t.Run("zero is valid", func(t *testing.T) {
		v, err := domain.NewContractValue(decimal.Zero)
		require.NoError(t, err)
		assert.True(t, v.IsZero())
	})

	t.Run("negative is invalid", func(t *testing.T) {
		_, err := domain.NewContractValue(decimal.NewFromFloat(-0.01))
		assert.ErrorIs(t, err, domain.ErrContractValueNegative)
	})
}

func TestNewContractNumber(t *testing.T) {
	t.Run("valid format", func(t *testing.T) {
		n, err := domain.NewContractNumber("CONT-2025-0042")
		require.NoError(t, err)
		assert.Equal(t, "CONT-2025-0042", n.Value())
	})

	t.Run("empty is invalid", func(t *testing.T) {
		_, err := domain.NewContractNumber("")
		assert.ErrorIs(t, err, domain.ErrContractNumberEmpty)
	})

	t.Run("wrong format is invalid", func(t *testing.T) {
		_, err := domain.NewContractNumber("CONTRACT-001")
		assert.ErrorIs(t, err, domain.ErrContractNumberInvalid)
	})
}
```

## Entity Helper Tests

```go
func TestContract_IsEditable(t *testing.T) {
	assert.True(t, (&domain.Contract{Status: domain.ContractStatusDraft}).IsEditable())
	assert.False(t, (&domain.Contract{Status: domain.ContractStatusSubmitted}).IsEditable())
	assert.False(t, (&domain.Contract{Status: domain.ContractStatusActive}).IsEditable())
}

func TestContract_IsDeleted(t *testing.T) {
	now := time.Now()
	assert.True(t, (&domain.Contract{DeletedAt: &now}).IsDeleted())
	assert.False(t, (&domain.Contract{DeletedAt: nil}).IsDeleted())
}
```

These tests have no external dependencies, run in milliseconds, and never need a database.
