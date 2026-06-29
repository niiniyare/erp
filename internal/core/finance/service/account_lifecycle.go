// Package service — account lifecycle service.
//
// AccountLifecycleService wraps AccountService with formal lifecycle workflow
// enforcement. All state transitions go through this service — never directly
// through UpdateAccount — so that:
//
//   - Every transition is validated against the state machine.
//   - SOD constraints are enforced (approver ≠ creator).
//   - Every transition is recorded as an AccountLifecycleEvent.
//   - Business preconditions (zero balance for close, validation for activate)
//     are checked before writing.
//
// This service does NOT manage opening balances (see OpeningBalanceEngine)
// or integrity scans (see COAIntegrityService).
//
// All methods are nil-safe on the service receiver.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// AccountLifecycleService orchestrates formal account lifecycle transitions.
type AccountLifecycleService struct {
	accounts domain.AccountsRepository
	events   domain.AccountLifecycleEventRepository // nil = events not persisted
	governor *AccountMutationGovernor
	metrics  metrics.MetricsProvider
}

// NewAccountLifecycleService constructs the service.
// events may be nil in tests or when event persistence is not yet wired.
func NewAccountLifecycleService(
	accounts domain.AccountsRepository,
	events domain.AccountLifecycleEventRepository,
	governor *AccountMutationGovernor,
	m metrics.MetricsProvider,
) *AccountLifecycleService {
	return &AccountLifecycleService{
		accounts: accounts,
		events:   events,
		governor: governor,
		metrics:  m,
	}
}

// =============================================================================
// Approval workflow
// =============================================================================

// SubmitForApproval transitions a DRAFT account to PENDING_APPROVAL.
// The submitter (submitterID) cannot later approve the same account (SOD).
func (s *AccountLifecycleService) SubmitForApproval(
	ctx context.Context,
	accountID uuid.UUID,
	submitterID uuid.UUID,
	reason string,
) error {
	if s == nil {
		return nil
	}

	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	if account.Status != domain.AccountStatusDraft {
		return fmt.Errorf("%w: current status is %s", domain.ErrAccountNotInDraft, account.Status)
	}

	// Validate the account is structurally sound before submission.
	if errs := account.Validate(); len(errs) > 0 {
		return fmt.Errorf("account validation failed before submission: %v", errs)
	}

	if err := s.governor.ValidateStatusTransition(ctx, account, domain.AccountStatusPendingApproval, submitterID); err != nil {
		return err
	}

	// Persist transition.
	account.Status = domain.AccountStatusPendingApproval
	if err := s.accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account status: %w", err)
	}

	s.appendEvent(ctx, &domain.AccountLifecycleEvent{
		ID:         uuid.New(),
		TenantID:   account.TenantID,
		AccountID:  accountID,
		Kind:       domain.EventKindSubmitted,
		FromStatus: domain.AccountStatusDraft,
		ToStatus:   domain.AccountStatusPendingApproval,
		ActorID:    submitterID,
		Reason:     reason,
		OccurredAt: time.Now(),
	})

	s.recordTransition(ctx, "submitted", account.ID)
	return nil
}

// Approve transitions a PENDING_APPROVAL account to ACTIVE.
// Enforces SOD: approverID cannot equal account.CreatedBy.
func (s *AccountLifecycleService) Approve(
	ctx context.Context,
	accountID uuid.UUID,
	approverID uuid.UUID,
	notes string,
) error {
	if s == nil {
		return nil
	}

	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	if account.Status != domain.AccountStatusPendingApproval {
		return fmt.Errorf("%w: current status is %s", domain.ErrAccountNotPendingApproval, account.Status)
	}

	// Governor enforces SOD (approver ≠ creator) and state machine.
	if err := s.governor.ValidateStatusTransition(ctx, account, domain.AccountStatusActive, approverID); err != nil {
		return err
	}

	// Re-validate before activation to catch any drift since submission.
	if errs := account.Validate(); len(errs) > 0 {
		return fmt.Errorf("account validation failed before approval: %v", errs)
	}

	account.Status = domain.AccountStatusActive
	account.IsActive = true
	updatedBy := approverID
	account.UpdatedBy = &updatedBy
	if err := s.accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to activate account: %w", err)
	}

	s.appendEvent(ctx, &domain.AccountLifecycleEvent{
		ID:         uuid.New(),
		TenantID:   account.TenantID,
		AccountID:  accountID,
		Kind:       domain.EventKindApproved,
		FromStatus: domain.AccountStatusPendingApproval,
		ToStatus:   domain.AccountStatusActive,
		ActorID:    approverID,
		Reason:     notes,
		OccurredAt: time.Now(),
	})

	s.recordTransition(ctx, "approved", account.ID)
	return nil
}

// Reject returns a PENDING_APPROVAL account to DRAFT for rework.
func (s *AccountLifecycleService) Reject(
	ctx context.Context,
	accountID uuid.UUID,
	approverID uuid.UUID,
	reason string,
) error {
	if s == nil {
		return nil
	}

	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}

	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	if account.Status != domain.AccountStatusPendingApproval {
		return fmt.Errorf("%w: current status is %s", domain.ErrAccountNotPendingApproval, account.Status)
	}

	if err := s.governor.ValidateStatusTransition(ctx, account, domain.AccountStatusDraft, approverID); err != nil {
		return err
	}

	account.Status = domain.AccountStatusDraft
	if err := s.accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to reject account: %w", err)
	}

	s.appendEvent(ctx, &domain.AccountLifecycleEvent{
		ID:         uuid.New(),
		TenantID:   account.TenantID,
		AccountID:  accountID,
		Kind:       domain.EventKindRejected,
		FromStatus: domain.AccountStatusPendingApproval,
		ToStatus:   domain.AccountStatusDraft,
		ActorID:    approverID,
		Reason:     reason,
		OccurredAt: time.Now(),
	})

	s.recordTransition(ctx, "rejected", account.ID)
	return nil
}

// =============================================================================
// Operational transitions
// =============================================================================

// Freeze transitions an account to FROZEN. No preconditions on balance;
// freeze is an emergency operation. Records lifecycle event.
func (s *AccountLifecycleService) Freeze(
	ctx context.Context,
	accountID uuid.UUID,
	operatorID uuid.UUID,
	reason string,
) error {
	if s == nil {
		return nil
	}
	if reason == "" {
		return fmt.Errorf("freeze reason is required")
	}
	return s.transitionWithEvent(ctx, accountID, operatorID,
		domain.AccountStatusFrozen, domain.EventKindFrozen, reason)
}

// Unfreeze transitions a FROZEN account back to ACTIVE.
// Re-runs Validate() before allowing the account to accept postings again.
func (s *AccountLifecycleService) Unfreeze(
	ctx context.Context,
	accountID uuid.UUID,
	operatorID uuid.UUID,
) error {
	if s == nil {
		return nil
	}

	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	if account.Status != domain.AccountStatusFrozen {
		return fmt.Errorf("account is not frozen (current status: %s)", account.Status)
	}

	if err := s.governor.ValidateStatusTransition(ctx, account, domain.AccountStatusActive, operatorID); err != nil {
		return err
	}

	// Re-validate before restoring posting capability.
	if errs := account.Validate(); len(errs) > 0 {
		return fmt.Errorf("account fails validation after freeze period: %v", errs)
	}

	account.Status = domain.AccountStatusActive
	account.IsActive = true
	if err := s.accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to unfreeze account: %w", err)
	}

	s.appendEvent(ctx, &domain.AccountLifecycleEvent{
		ID:         uuid.New(),
		TenantID:   account.TenantID,
		AccountID:  accountID,
		Kind:       domain.EventKindUnfrozen,
		FromStatus: domain.AccountStatusFrozen,
		ToStatus:   domain.AccountStatusActive,
		ActorID:    operatorID,
		OccurredAt: time.Now(),
	})

	s.recordTransition(ctx, "unfrozen", account.ID)
	return nil
}

// Close permanently closes an account. The account must:
//   - Currently be INACTIVE.
//   - Have a zero balance.
//   - Have no child accounts.
//
// Closure is irreversible. The account is retained for historical reporting.
func (s *AccountLifecycleService) Close(
	ctx context.Context,
	accountID uuid.UUID,
	operatorID uuid.UUID,
	reason string,
) error {
	if s == nil {
		return nil
	}

	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	if account.Status != domain.AccountStatusInactive {
		return fmt.Errorf("account must be INACTIVE before closing (current: %s)", account.Status)
	}

	// Governor validates zero balance and no children.
	if err := s.governor.ValidateArchival(ctx, account); err != nil {
		return err
	}

	if err := s.governor.ValidateStatusTransition(ctx, account, domain.AccountStatusClosed, operatorID); err != nil {
		return err
	}

	account.Status = domain.AccountStatusClosed
	account.IsActive = false
	if err := s.accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to close account: %w", err)
	}

	s.appendEvent(ctx, &domain.AccountLifecycleEvent{
		ID:         uuid.New(),
		TenantID:   account.TenantID,
		AccountID:  accountID,
		Kind:       domain.EventKindClosed,
		FromStatus: domain.AccountStatusInactive,
		ToStatus:   domain.AccountStatusClosed,
		ActorID:    operatorID,
		Reason:     reason,
		OccurredAt: time.Now(),
	})

	s.recordTransition(ctx, "closed", account.ID)
	return nil
}

// Archive moves a CLOSED account to long-term ARCHIVED storage.
// Archived accounts are read-only forever.
func (s *AccountLifecycleService) Archive(
	ctx context.Context,
	accountID uuid.UUID,
	operatorID uuid.UUID,
) error {
	if s == nil {
		return nil
	}
	return s.transitionWithEvent(ctx, accountID, operatorID,
		domain.AccountStatusArchived, domain.EventKindArchived, "archive")
}

// Suspend places an ACTIVE account on administrative hold.
func (s *AccountLifecycleService) Suspend(
	ctx context.Context,
	accountID uuid.UUID,
	operatorID uuid.UUID,
	reason string,
) error {
	if s == nil {
		return nil
	}
	if reason == "" {
		return fmt.Errorf("suspension reason is required")
	}
	return s.transitionWithEvent(ctx, accountID, operatorID,
		domain.AccountStatusSuspended, domain.EventKindSuspended, reason)
}

// PlaceComplianceHold applies a regulatory compliance hold to an account.
func (s *AccountLifecycleService) PlaceComplianceHold(
	ctx context.Context,
	accountID uuid.UUID,
	operatorID uuid.UUID,
	reason string,
) error {
	if s == nil {
		return nil
	}
	if reason == "" {
		return fmt.Errorf("compliance hold reason is required")
	}
	return s.transitionWithEvent(ctx, accountID, operatorID,
		domain.AccountStatusComplianceHold, domain.EventKindComplianceHold, reason)
}

// =============================================================================
// Lifecycle event history
// =============================================================================

// GetLifecycleHistory returns all lifecycle events for an account in
// chronological order. Returns an empty slice when the events repository
// is not wired (nil events).
func (s *AccountLifecycleService) GetLifecycleHistory(
	ctx context.Context,
	accountID uuid.UUID,
) ([]*domain.AccountLifecycleEvent, error) {
	if s == nil || s.events == nil {
		return nil, nil
	}
	return s.events.ListForAccount(ctx, accountID)
}

// =============================================================================
// Internal helpers
// =============================================================================

// transitionWithEvent is the shared implementation for transitions that do not
// require special preconditions beyond the state machine and governor.
func (s *AccountLifecycleService) transitionWithEvent(
	ctx context.Context,
	accountID uuid.UUID,
	actorID uuid.UUID,
	to domain.AccountStatus,
	kind domain.AccountLifecycleEventKind,
	reason string,
) error {
	account, err := s.accounts.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	from := account.Status

	if err := s.governor.ValidateStatusTransition(ctx, account, to, actorID); err != nil {
		return err
	}

	account.Status = to
	if to == domain.AccountStatusActive {
		account.IsActive = true
	} else if to == domain.AccountStatusInactive || to == domain.AccountStatusClosed || to == domain.AccountStatusArchived {
		account.IsActive = false
	}

	if err := s.accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to transition account to %s: %w", to, err)
	}

	s.appendEvent(ctx, &domain.AccountLifecycleEvent{
		ID:         uuid.New(),
		TenantID:   account.TenantID,
		AccountID:  accountID,
		Kind:       kind,
		FromStatus: from,
		ToStatus:   to,
		ActorID:    actorID,
		Reason:     reason,
		OccurredAt: time.Now(),
	})

	s.recordTransition(ctx, string(kind), account.ID)
	return nil
}

func (s *AccountLifecycleService) appendEvent(ctx context.Context, event *domain.AccountLifecycleEvent) {
	if s.events == nil {
		return
	}
	if err := s.events.Append(ctx, event); err != nil {
		// Event persistence failure must NOT block the transition — the
		// structural change has already been committed. Log and alert.
		logger.ErrorContext(ctx, "failed to persist account lifecycle event", logger.Fields{
			"account_id": event.AccountID.String(),
			"kind":       string(event.Kind),
			"error":      err.Error(),
		})
	}
}

func (s *AccountLifecycleService) recordTransition(ctx context.Context, kind string, accountID uuid.UUID) {
	tenantID, _ := shared.GetTenantID(ctx)
	if s.metrics != nil {
		s.metrics.IncrementCounter("finance_account_lifecycle_transitions_total", metrics.Fields{
			"kind":      kind,
			"tenant_id": tenantID.String(),
		})
	}
	logger.InfoContext(ctx, "account lifecycle transition", logger.Fields{
		"kind":       kind,
		"account_id": accountID.String(),
	})
}
