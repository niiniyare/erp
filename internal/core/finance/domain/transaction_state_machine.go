package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TransactionStateMachine manages state transitions for transactions
type TransactionStateMachine struct {
	transaction *Transaction
}

// NewTransactionStateMachine creates a new state machine for a transaction
func NewTransactionStateMachine(transaction *Transaction) *TransactionStateMachine {
	return &TransactionStateMachine{
		transaction: transaction,
	}
}

// StateTransition represents a state change operation
type StateTransition struct {
	FromStatus    TransactionStatus `json:"from_status"`
	ToStatus      TransactionStatus `json:"to_status"`
	TransitionBy  uuid.UUID         `json:"transition_by"`
	TransitionAt  time.Time         `json:"transition_at"`
	Notes         *string           `json:"notes,omitempty"`
	ReasonCode    *string           `json:"reason_code,omitempty"`
	ValidateRules bool              `json:"validate_rules"`
}

// TransactionStatusTransitions defines valid status transitions
var TransactionStatusTransitions = map[TransactionStatus][]TransactionStatus{
	TransactionStatusDraft: {
		TransactionStatusPendingApproval,
		TransactionStatusApproved,
		TransactionStatusCancelled,
	},
	TransactionStatusPendingApproval: {
		TransactionStatusApproved,
		TransactionStatusRejected,
		TransactionStatusCancelled,
		TransactionStatusDraft, // Allow return to draft for edits
	},
	TransactionStatusApproved: {
		TransactionStatusPosted,
		TransactionStatusCancelled,
		TransactionStatusDraft, // Allow return to draft for edits (if not posted)
	},
	TransactionStatusPosted: {
		TransactionStatusReversed,
		// Posted transactions generally cannot change status except reversal
	},
	TransactionStatusRejected: {
		TransactionStatusDraft,     // Allow fixing and resubmitting
		TransactionStatusCancelled, // Or cancel entirely
	},
	TransactionStatusCancelled: {
		// Cancelled transactions generally cannot change status
		// Only allow draft if explicitly permitted by business rules
	},
	TransactionStatusReversed: {
		// Reversed transactions cannot change status
	},
}

// CanTransitionTo checks if a transition from current status to target status is valid
func (tsm *TransactionStateMachine) CanTransitionTo(toStatus TransactionStatus) (bool, string) {
	currentStatus := tsm.transaction.TransactionStatus

	// Check if transition is defined in the rules
	validTransitions, exists := TransactionStatusTransitions[currentStatus]
	if !exists {
		return false, fmt.Sprintf("no transitions defined for status %s", currentStatus)
	}

	// Check if target status is in valid transitions
	for _, validStatus := range validTransitions {
		if validStatus == toStatus {
			return tsm.validateTransitionRules(currentStatus, toStatus)
		}
	}

	return false, fmt.Sprintf("invalid transition from %s to %s", currentStatus, toStatus)
}

// validateTransitionRules validates business rules for specific transitions
func (tsm *TransactionStateMachine) validateTransitionRules(from, to TransactionStatus) (bool, string) {
	t := tsm.transaction

	switch to {
	case TransactionStatusApproved:
		// Cannot approve without proper validation
		if !t.IsBalanced() {
			return false, "cannot approve unbalanced transaction"
		}
		if len(t.Entries) < 2 {
			return false, "cannot approve transaction with insufficient entries"
		}
		// Check if approval is actually required
		if !t.ApprovalRequired && from == TransactionStatusDraft {
			return false, "transaction does not require approval"
		}

	case TransactionStatusPosted:
		// Can only post approved transactions (or draft if no approval required)
		if t.ApprovalRequired && t.ApprovalStatus != ApprovalStatusApproved {
			return false, "cannot post transaction without approval when approval is required"
		}
		if !t.IsBalanced() {
			return false, "cannot post unbalanced transaction"
		}
		if t.ValidationStatus != ValidationStatusValid && t.ValidationStatus != ValidationStatusWarning {
			return false, "cannot post transaction with validation errors"
		}

	case TransactionStatusReversed:
		// Can only reverse posted transactions that haven't been reversed
		if from != TransactionStatusPosted {
			return false, "can only reverse posted transactions"
		}
		if t.IsReversed {
			return false, "transaction is already reversed"
		}

	case TransactionStatusDraft:
		// Can return to draft from certain states for editing
		if from == TransactionStatusPosted {
			return false, "cannot return posted transaction to draft status"
		}
		if from == TransactionStatusReversed {
			return false, "cannot return reversed transaction to draft status"
		}

	case TransactionStatusCancelled:
		// Cannot cancel posted or reversed transactions
		if from == TransactionStatusPosted {
			return false, "cannot cancel posted transaction - use reversal instead"
		}
		if from == TransactionStatusReversed {
			return false, "cannot cancel already reversed transaction"
		}

	case TransactionStatusPendingApproval:
		// Only draft transactions can be submitted for approval
		if from != TransactionStatusDraft {
			return false, "only draft transactions can be submitted for approval"
		}
		if !t.ApprovalRequired {
			return false, "transaction does not require approval"
		}
		if !t.IsBalanced() {
			return false, "cannot submit unbalanced transaction for approval"
		}

	case TransactionStatusRejected:
		// Can only reject transactions that are pending approval
		if from != TransactionStatusPendingApproval {
			return false, "can only reject transactions pending approval"
		}
	}

	return true, ""
}

// TransitionTo performs a state transition with validation
func (tsm *TransactionStateMachine) TransitionTo(transition StateTransition) error {
	// Validate the transition
	canTransition, reason := tsm.CanTransitionTo(transition.ToStatus)
	if !canTransition {
		return &StateTransitionError{
			FromStatus: transition.FromStatus,
			ToStatus:   transition.ToStatus,
			Reason:     reason,
		}
	}

	// Perform the transition
	_ = tsm.transaction.TransactionStatus
	tsm.transaction.TransactionStatus = transition.ToStatus
	tsm.transaction.UpdatedAt = transition.TransitionAt
	tsm.transaction.UpdatedBy = &transition.TransitionBy

	// Update related fields based on transition
	switch transition.ToStatus {
	case TransactionStatusApproved:
		tsm.transaction.ApprovalStatus = ApprovalStatusApproved
		tsm.transaction.ApprovedBy = &transition.TransitionBy
		tsm.transaction.ApprovedAt = &transition.TransitionAt
		if transition.Notes != nil {
			tsm.transaction.ApprovalNotes = transition.Notes
		}

	case TransactionStatusRejected:
		tsm.transaction.ApprovalStatus = ApprovalStatusRejected
		if transition.Notes != nil {
			tsm.transaction.ApprovalNotes = transition.Notes
		}
		if transition.ReasonCode != nil {
			rejectionReason := RejectionReason(*transition.ReasonCode)
			tsm.transaction.RejectionReason = &rejectionReason
		}

	case TransactionStatusPosted:
		tsm.transaction.PostedBy = &transition.TransitionBy
		tsm.transaction.PostedAt = &transition.TransitionAt
		now := time.Now()
		if tsm.transaction.PostingDate == nil {
			tsm.transaction.PostingDate = &now
		}

	case TransactionStatusReversed:
		tsm.transaction.IsReversed = true
		if transition.Notes != nil {
			tsm.transaction.ReversalReason = transition.Notes
		}

	case TransactionStatusDraft:
		// Reset approval fields when returning to draft
		tsm.transaction.ApprovalStatus = ApprovalStatusNotRequired
		tsm.transaction.ApprovedBy = nil
		tsm.transaction.ApprovedAt = nil
		tsm.transaction.ApprovalNotes = nil
		tsm.transaction.RejectionReason = nil

	case TransactionStatusPendingApproval:
		tsm.transaction.ApprovalStatus = ApprovalStatusPending

	case TransactionStatusCancelled:
		// No special field updates needed for cancellation
		break
	}

	// Log the transition (in a real system, this would go to an audit log)
	return nil
}

// GetValidTransitions returns all valid transitions from current state
func (tsm *TransactionStateMachine) GetValidTransitions() []TransactionStatus {
	currentStatus := tsm.transaction.TransactionStatus
	validTransitions, exists := TransactionStatusTransitions[currentStatus]
	if !exists {
		return []TransactionStatus{}
	}

	// Filter transitions based on business rules
	var allowedTransitions []TransactionStatus
	for _, status := range validTransitions {
		if canTransition, _ := tsm.validateTransitionRules(currentStatus, status); canTransition {
			allowedTransitions = append(allowedTransitions, status)
		}
	}

	return allowedTransitions
}

// GetTransitionHistory would return the history of transitions for this transaction
// In a full implementation, this would query an audit log table
func (tsm *TransactionStateMachine) GetTransitionHistory() []StateTransition {
	// TODO: Implement by querying audit log
	return []StateTransition{}
}

// ValidateCurrentState validates that the transaction's current state is consistent
func (tsm *TransactionStateMachine) ValidateCurrentState() []ValidationError {
	var errors []ValidationError
	t := tsm.transaction

	switch t.TransactionStatus {
	case TransactionStatusApproved:
		if t.ApprovalRequired && t.ApprovalStatus != ApprovalStatusApproved {
			errors = append(errors, ValidationError{
				Field:   "approval_status",
				Message: "Approved transaction must have approval status set to approved",
				Code:    "INCONSISTENT_STATE",
			})
		}
		if t.ApprovedBy == nil {
			errors = append(errors, ValidationError{
				Field:   "approved_by",
				Message: "Approved transaction must have approver recorded",
				Code:    "MISSING_AUDIT_DATA",
			})
		}

	case TransactionStatusPosted:
		if t.PostedBy == nil {
			errors = append(errors, ValidationError{
				Field:   "posted_by",
				Message: "Posted transaction must have poster recorded",
				Code:    "MISSING_AUDIT_DATA",
			})
		}
		if t.PostedAt == nil {
			errors = append(errors, ValidationError{
				Field:   "posted_at",
				Message: "Posted transaction must have posting timestamp",
				Code:    "MISSING_AUDIT_DATA",
			})
		}

	case TransactionStatusReversed:
		if !t.IsReversed {
			errors = append(errors, ValidationError{
				Field:   "is_reversed",
				Message: "Reversed transaction must have is_reversed flag set",
				Code:    "INCONSISTENT_STATE",
			})
		}

	case TransactionStatusRejected:
		if t.RejectionReason == nil {
			errors = append(errors, ValidationError{
				Field:   "rejection_reason",
				Message: "Rejected transaction should have rejection reason",
				Code:    "MISSING_AUDIT_DATA",
			})
		}
	}

	return errors
}

// StateTransitionError represents an error during state transition
type StateTransitionError struct {
	FromStatus TransactionStatus `json:"from_status"`
	ToStatus   TransactionStatus `json:"to_status"`
	Reason     string            `json:"reason"`
}

func (e *StateTransitionError) Error() string {
	return fmt.Sprintf("invalid state transition from %s to %s: %s", e.FromStatus, e.ToStatus, e.Reason)
}

// IsStateTransitionError checks if an error is a state transition error
func IsStateTransitionError(err error) bool {
	_, ok := err.(*StateTransitionError)
	return ok
}

// TransactionWorkflowEngine coordinates complex workflow operations
type TransactionWorkflowEngine struct {
	stateMachine *TransactionStateMachine
}

// NewTransactionWorkflowEngine creates a new workflow engine
func NewTransactionWorkflowEngine(transaction *Transaction) *TransactionWorkflowEngine {
	return &TransactionWorkflowEngine{
		stateMachine: NewTransactionStateMachine(transaction),
	}
}

// SubmitForApproval submits a transaction for approval workflow
func (twe *TransactionWorkflowEngine) SubmitForApproval(submittedBy uuid.UUID, notes *string) error {
	transition := StateTransition{
		FromStatus:   twe.stateMachine.transaction.TransactionStatus,
		ToStatus:     TransactionStatusPendingApproval,
		TransitionBy: submittedBy,
		TransitionAt: time.Now(),
		Notes:        notes,
	}

	return twe.stateMachine.TransitionTo(transition)
}

// Approve approves a pending transaction
func (twe *TransactionWorkflowEngine) Approve(approvedBy uuid.UUID, notes *string) error {
	transition := StateTransition{
		FromStatus:   twe.stateMachine.transaction.TransactionStatus,
		ToStatus:     TransactionStatusApproved,
		TransitionBy: approvedBy,
		TransitionAt: time.Now(),
		Notes:        notes,
	}

	return twe.stateMachine.TransitionTo(transition)
}

// Reject rejects a pending transaction
func (twe *TransactionWorkflowEngine) Reject(rejectedBy uuid.UUID, reason string, reasonCode *string) error {
	transition := StateTransition{
		FromStatus:   twe.stateMachine.transaction.TransactionStatus,
		ToStatus:     TransactionStatusRejected,
		TransitionBy: rejectedBy,
		TransitionAt: time.Now(),
		Notes:        &reason,
		ReasonCode:   reasonCode,
	}

	return twe.stateMachine.TransitionTo(transition)
}

// Post posts an approved transaction to the ledger
func (twe *TransactionWorkflowEngine) Post(postedBy uuid.UUID) error {
	transition := StateTransition{
		FromStatus:   twe.stateMachine.transaction.TransactionStatus,
		ToStatus:     TransactionStatusPosted,
		TransitionBy: postedBy,
		TransitionAt: time.Now(),
	}

	return twe.stateMachine.TransitionTo(transition)
}

// Cancel cancels a transaction (if allowed)
func (twe *TransactionWorkflowEngine) Cancel(cancelledBy uuid.UUID, reason *string) error {
	transition := StateTransition{
		FromStatus:   twe.stateMachine.transaction.TransactionStatus,
		ToStatus:     TransactionStatusCancelled,
		TransitionBy: cancelledBy,
		TransitionAt: time.Now(),
		Notes:        reason,
	}

	return twe.stateMachine.TransitionTo(transition)
}

// Reverse reverses a posted transaction
func (twe *TransactionWorkflowEngine) Reverse(reversedBy uuid.UUID, reason string) error {
	transition := StateTransition{
		FromStatus:   twe.stateMachine.transaction.TransactionStatus,
		ToStatus:     TransactionStatusReversed,
		TransitionBy: reversedBy,
		TransitionAt: time.Now(),
		Notes:        &reason,
	}

	return twe.stateMachine.TransitionTo(transition)
}

// GetWorkflowStatus returns the current workflow status
func (twe *TransactionWorkflowEngine) GetWorkflowStatus() TransactionWorkflowStatus {
	t := twe.stateMachine.transaction

	return TransactionWorkflowStatus{
		TransactionID:    t.ID,
		CurrentStatus:    t.TransactionStatus,
		ApprovalStatus:   t.ApprovalStatus,
		ApprovalRequired: t.ApprovalRequired,
		IsReversed:       t.IsReversed,
		ValidTransitions: twe.stateMachine.GetValidTransitions(),
		ValidationStatus: t.ValidationStatus,
		ValidationErrors: t.ValidationErrors,
	}
}

// TransactionWorkflowStatus represents the current workflow state
type TransactionWorkflowStatus struct {
	TransactionID    uuid.UUID           `json:"transaction_id"`
	CurrentStatus    TransactionStatus   `json:"current_status"`
	ApprovalStatus   ApprovalStatus      `json:"approval_status"`
	ApprovalRequired bool                `json:"approval_required"`
	IsReversed       bool                `json:"is_reversed"`
	ValidTransitions []TransactionStatus `json:"valid_transitions"`
	ValidationStatus ValidationStatus    `json:"validation_status"`
	ValidationErrors []ValidationError   `json:"validation_errors,omitempty"`
}

