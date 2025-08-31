package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// TransactionWorkflowEngine handles approval workflows and business processes for transactions
type TransactionWorkflowEngine interface {
	// SubmitForApproval submits a transaction for approval workflow
	SubmitForApproval(ctx context.Context, req SubmitApprovalRequest) (*ApprovalWorkflowResult, error)

	// ApproveTransaction approves a transaction in the workflow
	ApproveTransaction(ctx context.Context, req ApprovalRequest) (*ApprovalResult, error)

	// RejectTransaction rejects a transaction in the workflow
	RejectTransaction(ctx context.Context, req RejectionRequest) (*RejectionResult, error)

	// RequestMoreInformation requests more information for a transaction
	RequestMoreInformation(ctx context.Context, req InfoRequest) (*InfoRequestResult, error)

	// UpdateTransactionInfo updates transaction information in response to requests
	UpdateTransactionInfo(ctx context.Context, req UpdateInfoRequest) (*UpdateInfoResult, error)

	// GetApprovalHistory gets the full approval history for a transaction
	GetApprovalHistory(ctx context.Context, transactionID uuid.UUID) (*ApprovalHistoryResult, error)

	// GetPendingApprovals gets pending approvals for a user or role
	GetPendingApprovals(ctx context.Context, req PendingApprovalsRequest) (*PendingApprovalsResult, error)

	// EscalateApproval escalates approval to next level or administrator
	EscalateApproval(ctx context.Context, req EscalationRequest) (*EscalationResult, error)

	// ValidateApprovalAuthority validates if a user can approve a transaction
	ValidateApprovalAuthority(ctx context.Context, req AuthorityValidationRequest) (*AuthorityValidationResult, error)
}

// SubmitApprovalRequest represents a request to submit transaction for approval
type SubmitApprovalRequest struct {
	TransactionID     uuid.UUID    `json:"transaction_id"`
	SubmittedBy       uuid.UUID    `json:"submitted_by"`
	ApprovalType      ApprovalType `json:"approval_type"`
	RequestedBy       uuid.UUID    `json:"requested_by,omitempty"`
	Priority          Priority     `json:"priority"`
	Comments          string       `json:"comments,omitempty"`
	RequiredApprovers []uuid.UUID  `json:"required_approvers,omitempty"`
	DueDate           *time.Time   `json:"due_date,omitempty"`
}

// ApprovalRequest represents a request to approve a transaction
type ApprovalRequest struct {
	TransactionID    uuid.UUID      `json:"transaction_id"`
	ApprovedBy       uuid.UUID      `json:"approved_by"`
	ApprovalLevel    int32          `json:"approval_level"`
	ApprovalComments string         `json:"approval_comments,omitempty"`
	ApprovalAction   ApprovalAction `json:"approval_action"`
	ConditionalTerms []string       `json:"conditional_terms,omitempty"`
	EffectiveDate    *time.Time     `json:"effective_date,omitempty"`
}

// RejectionRequest represents a request to reject a transaction
type RejectionRequest struct {
	TransactionID     uuid.UUID       `json:"transaction_id"`
	RejectedBy        uuid.UUID       `json:"rejected_by"`
	RejectionReason   RejectionReason `json:"rejection_reason"`
	RejectionComments string          `json:"rejection_comments"`
	SuggestedChanges  []string        `json:"suggested_changes,omitempty"`
	AllowResubmission bool            `json:"allow_resubmission"`
}

// InfoRequest represents a request for more information
type InfoRequest struct {
	TransactionID       uuid.UUID  `json:"transaction_id"`
	RequestedBy         uuid.UUID  `json:"requested_by"`
	InformationRequired []string   `json:"information_required"`
	RequestComments     string     `json:"request_comments"`
	DueDate             *time.Time `json:"due_date,omitempty"`
}

// UpdateInfoRequest represents an update to transaction information
type UpdateInfoRequest struct {
	TransactionID       uuid.UUID            `json:"transaction_id"`
	UpdatedBy           uuid.UUID            `json:"updated_by"`
	InformationUpdates  []InformationUpdate  `json:"information_updates"`
	UpdateComments      string               `json:"update_comments,omitempty"`
	SupportingDocuments []SupportingDocument `json:"supporting_documents,omitempty"`
}

// PendingApprovalsRequest represents a request for pending approvals
type PendingApprovalsRequest struct {
	UserID        *uuid.UUID     `json:"user_id,omitempty"`
	RoleID        *uuid.UUID     `json:"role_id,omitempty"`
	ApprovalTypes []ApprovalType `json:"approval_types,omitempty"`
	Priority      *Priority      `json:"priority,omitempty"`
	OlderThan     *time.Time     `json:"older_than,omitempty"`
	Limit         int32          `json:"limit"`
	Offset        int32          `json:"offset"`
}

// EscalationRequest represents a request to escalate approval
type EscalationRequest struct {
	TransactionID     uuid.UUID        `json:"transaction_id"`
	EscalatedBy       uuid.UUID        `json:"escalated_by"`
	EscalationReason  EscalationReason `json:"escalation_reason"`
	EscalationLevel   int32            `json:"escalation_level"`
	EscalationComment string           `json:"escalation_comment,omitempty"`
	UrgentFlag        bool             `json:"urgent_flag"`
}

// AuthorityValidationRequest represents a request to validate approval authority
type AuthorityValidationRequest struct {
	UserID        uuid.UUID    `json:"user_id"`
	TransactionID uuid.UUID    `json:"transaction_id"`
	ApprovalLevel int32        `json:"approval_level"`
	ApprovalType  ApprovalType `json:"approval_type"`
}

// ApprovalType represents the type of approval required
type ApprovalType string

const (
	ApprovalTypeFinancial  ApprovalType = "FINANCIAL"
	ApprovalTypeBudgetary  ApprovalType = "BUDGETARY"
	ApprovalTypeCompliance ApprovalType = "COMPLIANCE"
	ApprovalTypeManagerial ApprovalType = "MANAGERIAL"
	ApprovalTypeExecutive  ApprovalType = "EXECUTIVE"
	ApprovalTypeTechnical  ApprovalType = "TECHNICAL"
	ApprovalTypeAudit      ApprovalType = "AUDIT"
)

// Priority represents the priority level of approval
type Priority string

const (
	PriorityLow      Priority = "LOW"
	PriorityNormal   Priority = "NORMAL"
	PriorityHigh     Priority = "HIGH"
	PriorityUrgent   Priority = "URGENT"
	PriorityCritical Priority = "CRITICAL"
)

// ApprovalAction represents the action taken during approval
type ApprovalAction string

const (
	ApprovalActionApprove            ApprovalAction = "APPROVE"
	ApprovalActionApproveConditional ApprovalAction = "APPROVE_CONDITIONAL"
	ApprovalActionDelegate           ApprovalAction = "DELEGATE"
	ApprovalActionReassign           ApprovalAction = "REASSIGN"
)

// RejectionReason represents the reason for rejection
type RejectionReason string

const (
	RejectionReasonInsufficientDocumentation RejectionReason = "INSUFFICIENT_DOCUMENTATION"
	RejectionReasonBudgetConstraints         RejectionReason = "BUDGET_CONSTRAINTS"
	RejectionReasonPolicyViolation           RejectionReason = "POLICY_VIOLATION"
	RejectionReasonComplianceIssue           RejectionReason = "COMPLIANCE_ISSUE"
	RejectionReasonIncorrectInformation      RejectionReason = "INCORRECT_INFORMATION"
	RejectionReasonDuplicateRequest          RejectionReason = "DUPLICATE_REQUEST"
	RejectionReasonOther                     RejectionReason = "OTHER"
)

// EscalationReason represents the reason for escalation
type EscalationReason string

const (
	EscalationReasonTimeout          EscalationReason = "TIMEOUT"
	EscalationReasonComplexity       EscalationReason = "COMPLEXITY"
	EscalationReasonHighValue        EscalationReason = "HIGH_VALUE"
	EscalationReasonPolicyException  EscalationReason = "POLICY_EXCEPTION"
	EscalationReasonUrgency          EscalationReason = "URGENCY"
	EscalationReasonConflictInterest EscalationReason = "CONFLICT_OF_INTEREST"
)

// InformationUpdate represents an update to requested information
type InformationUpdate struct {
	Field        string `json:"field"`
	OldValue     string `json:"old_value,omitempty"`
	NewValue     string `json:"new_value"`
	UpdateReason string `json:"update_reason,omitempty"`
}

// SupportingDocument represents a supporting document
type SupportingDocument struct {
	DocumentID   uuid.UUID `json:"document_id"`
	DocumentType string    `json:"document_type"`
	DocumentName string    `json:"document_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	ContentType  string    `json:"content_type"`
}

// ApprovalWorkflowResult represents the result of submitting for approval
type ApprovalWorkflowResult struct {
	WorkflowID          uuid.UUID      `json:"workflow_id"`
	TransactionID       uuid.UUID      `json:"transaction_id"`
	WorkflowStatus      WorkflowStatus `json:"workflow_status"`
	ApprovalLevel       int32          `json:"approval_level"`
	RequiredApprovers   []ApproverInfo `json:"required_approvers"`
	EstimatedCompletion *time.Time     `json:"estimated_completion,omitempty"`
	WorkflowSteps       []WorkflowStep `json:"workflow_steps"`
	Success             bool           `json:"success"`
	Errors              []string       `json:"errors,omitempty"`
}

// ApprovalResult represents the result of an approval action
type ApprovalResult struct {
	TransactionID     uuid.UUID      `json:"transaction_id"`
	ApprovalID        uuid.UUID      `json:"approval_id"`
	ApprovalStatus    ApprovalStatus `json:"approval_status"`
	NextApprovalLevel *int32         `json:"next_approval_level,omitempty"`
	NextApprovers     []ApproverInfo `json:"next_approvers,omitempty"`
	WorkflowComplete  bool           `json:"workflow_complete"`
	FinalDecision     *FinalDecision `json:"final_decision,omitempty"`
	Success           bool           `json:"success"`
	Errors            []string       `json:"errors,omitempty"`
}

// RejectionResult represents the result of a rejection action
type RejectionResult struct {
	TransactionID       uuid.UUID       `json:"transaction_id"`
	RejectionID         uuid.UUID       `json:"rejection_id"`
	RejectionStatus     RejectionStatus `json:"rejection_status"`
	WorkflowStatus      WorkflowStatus  `json:"workflow_status"`
	ResubmissionAllowed bool            `json:"resubmission_allowed"`
	Success             bool            `json:"success"`
	Errors              []string        `json:"errors,omitempty"`
}

// InfoRequestResult represents the result of requesting information
type InfoRequestResult struct {
	TransactionID   uuid.UUID     `json:"transaction_id"`
	InfoRequestID   uuid.UUID     `json:"info_request_id"`
	RequestStatus   RequestStatus `json:"request_status"`
	ResponseDueDate *time.Time    `json:"response_due_date,omitempty"`
	Success         bool          `json:"success"`
	Errors          []string      `json:"errors,omitempty"`
}

// UpdateInfoResult represents the result of updating information
type UpdateInfoResult struct {
	TransactionID       uuid.UUID    `json:"transaction_id"`
	UpdateID            uuid.UUID    `json:"update_id"`
	UpdateStatus        UpdateStatus `json:"update_status"`
	WorkflowReactivated bool         `json:"workflow_reactivated"`
	NextAction          *NextAction  `json:"next_action,omitempty"`
	Success             bool         `json:"success"`
	Errors              []string     `json:"errors,omitempty"`
}

// ApprovalHistoryResult represents the approval history for a transaction
type ApprovalHistoryResult struct {
	TransactionID       uuid.UUID        `json:"transaction_id"`
	WorkflowID          uuid.UUID        `json:"workflow_id"`
	ApprovalHistory     []ApprovalRecord `json:"approval_history"`
	CurrentStatus       WorkflowStatus   `json:"current_status"`
	CurrentLevel        int32            `json:"current_level"`
	TotalLevelsRequired int32            `json:"total_levels_required"`
	WorkflowDuration    time.Duration    `json:"workflow_duration"`
}

// PendingApprovalsResult represents pending approvals for a user
type PendingApprovalsResult struct {
	PendingApprovals []PendingApproval `json:"pending_approvals"`
	TotalCount       int32             `json:"total_count"`
	OverdueCount     int32             `json:"overdue_count"`
	UrgentCount      int32             `json:"urgent_count"`
}

// EscalationResult represents the result of escalating an approval
type EscalationResult struct {
	TransactionID       uuid.UUID      `json:"transaction_id"`
	EscalationID        uuid.UUID      `json:"escalation_id"`
	NewApprovalLevel    int32          `json:"new_approval_level"`
	NewApprovers        []ApproverInfo `json:"new_approvers"`
	EscalationTimestamp time.Time      `json:"escalation_timestamp"`
	Success             bool           `json:"success"`
	Errors              []string       `json:"errors,omitempty"`
}

// AuthorityValidationResult represents the result of validating approval authority
type AuthorityValidationResult struct {
	HasAuthority      bool           `json:"has_authority"`
	AuthorityLevel    int32          `json:"authority_level"`
	AuthorityTypes    []ApprovalType `json:"authority_types"`
	ValidationReasons []string       `json:"validation_reasons,omitempty"`
	Restrictions      []string       `json:"restrictions,omitempty"`
}

// Supporting types

type WorkflowStatus string

const (
	WorkflowStatusPending    WorkflowStatus = "PENDING"
	WorkflowStatusInProgress WorkflowStatus = "IN_PROGRESS"
	WorkflowStatusApproved   WorkflowStatus = "APPROVED"
	WorkflowStatusRejected   WorkflowStatus = "REJECTED"
	WorkflowStatusCancelled  WorkflowStatus = "CANCELLED"
	WorkflowStatusEscalated  WorkflowStatus = "ESCALATED"
	WorkflowStatusOnHold     WorkflowStatus = "ON_HOLD"
)

type ApprovalStatus string

const (
	ApprovalStatusPending       ApprovalStatus = "PENDING"
	ApprovalStatusApproved      ApprovalStatus = "APPROVED"
	ApprovalStatusRejected      ApprovalStatus = "REJECTED"
	ApprovalStatusConditional   ApprovalStatus = "CONDITIONAL"
	ApprovalStatusDelegated     ApprovalStatus = "DELEGATED"
	ApprovalStatusEscalated     ApprovalStatus = "ESCALATED"
	ApprovalStatusInfoRequested ApprovalStatus = "INFO_REQUESTED"
)

type RejectionStatus string

const (
	RejectionStatusRejected            RejectionStatus = "REJECTED"
	RejectionStatusResubmissionAllowed RejectionStatus = "RESUBMISSION_ALLOWED"
	RejectionStatusFinalRejection      RejectionStatus = "FINAL_REJECTION"
)

type RequestStatus string

const (
	RequestStatusPending  RequestStatus = "PENDING"
	RequestStatusProvided RequestStatus = "PROVIDED"
	RequestStatusOverdue  RequestStatus = "OVERDUE"
)

type UpdateStatus string

const (
	UpdateStatusApplied  UpdateStatus = "APPLIED"
	UpdateStatusPending  UpdateStatus = "PENDING"
	UpdateStatusRejected UpdateStatus = "REJECTED"
)

type ApproverInfo struct {
	UserID       uuid.UUID    `json:"user_id"`
	UserName     string       `json:"user_name"`
	UserEmail    string       `json:"user_email"`
	ApprovalType ApprovalType `json:"approval_type"`
	Level        int32        `json:"level"`
	DueDate      *time.Time   `json:"due_date,omitempty"`
}

type WorkflowStep struct {
	StepID        uuid.UUID      `json:"step_id"`
	StepNumber    int32          `json:"step_number"`
	StepName      string         `json:"step_name"`
	StepType      string         `json:"step_type"`
	Status        WorkflowStatus `json:"status"`
	Approvers     []ApproverInfo `json:"approvers"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	EstimatedTime time.Duration  `json:"estimated_time"`
}

type FinalDecision struct {
	Decision      string    `json:"decision"`
	DecisionDate  time.Time `json:"decision_date"`
	DecisionBy    uuid.UUID `json:"decision_by"`
	DecisionNotes string    `json:"decision_notes,omitempty"`
}

type ApprovalRecord struct {
	ApprovalID     uuid.UUID      `json:"approval_id"`
	ApproverID     uuid.UUID      `json:"approver_id"`
	ApproverName   string         `json:"approver_name"`
	ApprovalLevel  int32          `json:"approval_level"`
	ApprovalAction ApprovalAction `json:"approval_action"`
	ApprovalStatus ApprovalStatus `json:"approval_status"`
	ApprovalDate   time.Time      `json:"approval_date"`
	Comments       string         `json:"comments,omitempty"`
	TimeToDecision time.Duration  `json:"time_to_decision"`
}

type PendingApproval struct {
	TransactionID          uuid.UUID    `json:"transaction_id"`
	TransactionNumber      string       `json:"transaction_number"`
	TransactionDescription string       `json:"transaction_description"`
	Amount                 string       `json:"amount"`
	Currency               string       `json:"currency"`
	SubmittedBy            uuid.UUID    `json:"submitted_by"`
	SubmittedAt            time.Time    `json:"submitted_at"`
	ApprovalType           ApprovalType `json:"approval_type"`
	Priority               Priority     `json:"priority"`
	DueDate                *time.Time   `json:"due_date,omitempty"`
	IsOverdue              bool         `json:"is_overdue"`
	DaysWaiting            int32        `json:"days_waiting"`
}

type NextAction struct {
	ActionType    string     `json:"action_type"`
	ActionBy      uuid.UUID  `json:"action_by"`
	ActionDue     *time.Time `json:"action_due,omitempty"`
	ActionDetails string     `json:"action_details,omitempty"`
}

// transactionWorkflowEngine implements TransactionWorkflowEngine
type transactionWorkflowEngine struct {
	transactionRepository domain.TransactionRepository
	tracing               tracing.TracingService
}

// TransactionWorkflowEngineDeps represents dependencies for the workflow engine
type TransactionWorkflowEngineDeps struct {
	TransactionRepository domain.TransactionRepository
	Tracing               tracing.TracingService
}

// NewTransactionWorkflowEngine creates a new transaction workflow engine
func NewTransactionWorkflowEngine(deps TransactionWorkflowEngineDeps) TransactionWorkflowEngine {
	return &transactionWorkflowEngine{
		transactionRepository: deps.TransactionRepository,
		tracing:               deps.Tracing,
	}
}

// SubmitForApproval submits a transaction for approval workflow
func (e *transactionWorkflowEngine) SubmitForApproval(ctx context.Context, req SubmitApprovalRequest) (*ApprovalWorkflowResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.SubmitForApproval")
	defer span.End()

	result := &ApprovalWorkflowResult{
		WorkflowID:        uuid.New(),
		TransactionID:     req.TransactionID,
		WorkflowStatus:    WorkflowStatusPending,
		ApprovalLevel:     1,
		RequiredApprovers: []ApproverInfo{},
		WorkflowSteps:     []WorkflowStep{},
		Success:           false,
		Errors:            []string{},
	}

	span.SetAttributes(
		attribute.String("transaction_id", req.TransactionID.String()),
		attribute.String("submitted_by", req.SubmittedBy.String()),
		attribute.String("approval_type", string(req.ApprovalType)),
		attribute.String("priority", string(req.Priority)),
	)

	// 1. Load and validate transaction
	transaction, err := e.transactionRepository.GetByID(ctx, req.TransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load transaction: %v", err))
		return result, err
	}

	// 2. Validate transaction is in correct state for approval
	if transaction.TransactionStatus != domain.TransactionStatusDraft &&
		transaction.TransactionStatus != domain.TransactionStatusPendingApproval {
		result.Errors = append(result.Errors, fmt.Sprintf("Transaction is not in a state that can be submitted for approval: %s", transaction.TransactionStatus))
		return result, fmt.Errorf("transaction cannot be submitted for approval")
	}

	// 3. Determine approval workflow based on transaction and approval type
	workflowSteps, err := e.determineApprovalWorkflow(ctx, transaction, req.ApprovalType, req.Priority)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to determine approval workflow: %v", err))
		return result, err
	}

	result.WorkflowSteps = workflowSteps
	result.RequiredApprovers = e.extractApproversFromSteps(workflowSteps)

	// 4. Set estimated completion time
	if len(workflowSteps) > 0 {
		totalEstimatedTime := time.Duration(0)
		for _, step := range workflowSteps {
			totalEstimatedTime += step.EstimatedTime
		}
		estimatedCompletion := time.Now().Add(totalEstimatedTime)
		result.EstimatedCompletion = &estimatedCompletion
	}

	// 5. Update transaction status
	transaction.TransactionStatus = domain.TransactionStatusPendingApproval
	transaction.ApprovalRequired = true
	transaction.ApprovalStatus = domain.ApprovalStatusPending

	err = e.transactionRepository.Update(ctx, transaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update transaction status: %v", err))
		return result, err
	}

	// 6. TODO: Create workflow record in database
	// 7. TODO: Send notifications to approvers

	result.Success = true
	result.WorkflowStatus = WorkflowStatusInProgress

	span.SetAttributes(
		attribute.String("workflow_id", result.WorkflowID.String()),
		attribute.Bool("success", result.Success),
		attribute.Int64("approval_levels", int64(len(workflowSteps))),
	)

	return result, nil
}

// ApproveTransaction approves a transaction in the workflow
func (e *transactionWorkflowEngine) ApproveTransaction(ctx context.Context, req ApprovalRequest) (*ApprovalResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.ApproveTransaction")
	defer span.End()

	result := &ApprovalResult{
		TransactionID:    req.TransactionID,
		ApprovalID:       uuid.New(),
		ApprovalStatus:   ApprovalStatusApproved,
		WorkflowComplete: false,
		Success:          false,
		Errors:           []string{},
	}

	span.SetAttributes(
		attribute.String("transaction_id", req.TransactionID.String()),
		attribute.String("approved_by", req.ApprovedBy.String()),
		attribute.Int64("approval_level", int64(req.ApprovalLevel)),
		attribute.String("approval_action", string(req.ApprovalAction)),
	)

	// 1. Load and validate transaction
	transaction, err := e.transactionRepository.GetByID(ctx, req.TransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load transaction: %v", err))
		return result, err
	}

	// 2. Validate transaction is pending approval
	if transaction.TransactionStatus != domain.TransactionStatusPendingApproval {
		result.Errors = append(result.Errors, "Transaction is not pending approval")
		return result, fmt.Errorf("transaction is not pending approval")
	}

	// 3. Validate approver authority
	authorityReq := AuthorityValidationRequest{
		UserID:        req.ApprovedBy,
		TransactionID: req.TransactionID,
		ApprovalLevel: req.ApprovalLevel,
		ApprovalType:  ApprovalTypeFinancial, // Default for now
	}

	authorityResult, err := e.ValidateApprovalAuthority(ctx, authorityReq)
	if err != nil || !authorityResult.HasAuthority {
		result.Errors = append(result.Errors, "Approver does not have authority to approve this transaction")
		return result, fmt.Errorf("insufficient approval authority")
	}

	// 4. Process approval action
	switch req.ApprovalAction {
	case ApprovalActionApprove:
		result.ApprovalStatus = ApprovalStatusApproved
	case ApprovalActionApproveConditional:
		result.ApprovalStatus = ApprovalStatusConditional
	case ApprovalActionDelegate:
		result.ApprovalStatus = ApprovalStatusDelegated
	case ApprovalActionReassign:
		result.ApprovalStatus = ApprovalStatusEscalated
	}

	// 5. Determine if workflow is complete or needs next level
	isComplete, nextLevel, nextApprovers := e.determineWorkflowProgress(ctx, transaction, req.ApprovalLevel)

	result.WorkflowComplete = isComplete
	if !isComplete {
		result.NextApprovalLevel = &nextLevel
		result.NextApprovers = nextApprovers
	}

	// 6. Update transaction based on workflow status
	if isComplete {
		// Final approval - mark as approved
		transaction.TransactionStatus = domain.TransactionStatusApproved
		transaction.ApprovalStatus = domain.ApprovalStatusApproved
		transaction.ApprovedBy = &req.ApprovedBy
		now := time.Now()
		transaction.ApprovedAt = &now
		transaction.ApprovalNotes = &req.ApprovalComments

		result.FinalDecision = &FinalDecision{
			Decision:      "APPROVED",
			DecisionDate:  now,
			DecisionBy:    req.ApprovedBy,
			DecisionNotes: req.ApprovalComments,
		}
	} else {
		// Partial approval - keep pending but update level
		transaction.ApprovalStatus = domain.ApprovalStatusPartiallyApproved
	}

	err = e.transactionRepository.Update(ctx, transaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update transaction: %v", err))
		return result, err
	}

	// 7. TODO: Record approval in workflow history
	// 8. TODO: Send notifications for next approvers or completion

	result.Success = true

	span.SetAttributes(
		attribute.String("approval_id", result.ApprovalID.String()),
		attribute.Bool("workflow_complete", result.WorkflowComplete),
		attribute.Bool("success", result.Success),
	)

	return result, nil
}

// RejectTransaction rejects a transaction in the workflow
func (e *transactionWorkflowEngine) RejectTransaction(ctx context.Context, req RejectionRequest) (*RejectionResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.RejectTransaction")
	defer span.End()

	result := &RejectionResult{
		TransactionID:       req.TransactionID,
		RejectionID:         uuid.New(),
		RejectionStatus:     RejectionStatusRejected,
		WorkflowStatus:      WorkflowStatusRejected,
		ResubmissionAllowed: req.AllowResubmission,
		Success:             false,
		Errors:              []string{},
	}

	// Load transaction and update status
	transaction, err := e.transactionRepository.GetByID(ctx, req.TransactionID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load transaction: %v", err))
		return result, err
	}

	// Update transaction status
	transaction.TransactionStatus = domain.TransactionStatusRejected
	transaction.ApprovalStatus = domain.ApprovalStatusRejected

	err = e.transactionRepository.Update(ctx, transaction)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to update transaction: %v", err))
		return result, err
	}

	result.Success = true
	return result, nil
}

// RequestMoreInformation requests more information for a transaction
func (e *transactionWorkflowEngine) RequestMoreInformation(ctx context.Context, req InfoRequest) (*InfoRequestResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.RequestMoreInformation")
	defer span.End()

	result := &InfoRequestResult{
		TransactionID: req.TransactionID,
		InfoRequestID: uuid.New(),
		RequestStatus: RequestStatusPending,
		Success:       false,
		Errors:        []string{},
	}

	if req.DueDate != nil {
		result.ResponseDueDate = req.DueDate
	}

	// TODO: Implement information request logic
	result.Success = true
	return result, nil
}

// UpdateTransactionInfo updates transaction information in response to requests
func (e *transactionWorkflowEngine) UpdateTransactionInfo(ctx context.Context, req UpdateInfoRequest) (*UpdateInfoResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.UpdateTransactionInfo")
	defer span.End()

	result := &UpdateInfoResult{
		TransactionID:       req.TransactionID,
		UpdateID:            uuid.New(),
		UpdateStatus:        UpdateStatusApplied,
		WorkflowReactivated: true,
		Success:             false,
		Errors:              []string{},
	}

	// TODO: Implement information update logic
	result.Success = true
	return result, nil
}

// GetApprovalHistory gets the full approval history for a transaction
func (e *transactionWorkflowEngine) GetApprovalHistory(ctx context.Context, transactionID uuid.UUID) (*ApprovalHistoryResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.GetApprovalHistory")
	defer span.End()

	result := &ApprovalHistoryResult{
		TransactionID:       transactionID,
		WorkflowID:          uuid.New(), // Would be loaded from database
		ApprovalHistory:     []ApprovalRecord{},
		CurrentStatus:       WorkflowStatusPending,
		CurrentLevel:        1,
		TotalLevelsRequired: 1,
		WorkflowDuration:    time.Duration(0),
	}

	// TODO: Implement approval history loading from database
	return result, nil
}

// GetPendingApprovals gets pending approvals for a user or role
func (e *transactionWorkflowEngine) GetPendingApprovals(ctx context.Context, req PendingApprovalsRequest) (*PendingApprovalsResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.GetPendingApprovals")
	defer span.End()

	result := &PendingApprovalsResult{
		PendingApprovals: []PendingApproval{},
		TotalCount:       0,
		OverdueCount:     0,
		UrgentCount:      0,
	}

	// TODO: Implement pending approvals loading from database
	// This would typically involve querying pending transactions and matching with user roles/permissions

	return result, nil
}

// EscalateApproval escalates approval to next level or administrator
func (e *transactionWorkflowEngine) EscalateApproval(ctx context.Context, req EscalationRequest) (*EscalationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.EscalateApproval")
	defer span.End()

	result := &EscalationResult{
		TransactionID:       req.TransactionID,
		EscalationID:        uuid.New(),
		NewApprovalLevel:    req.EscalationLevel,
		NewApprovers:        []ApproverInfo{},
		EscalationTimestamp: time.Now(),
		Success:             false,
		Errors:              []string{},
	}

	// TODO: Implement escalation logic
	result.Success = true
	return result, nil
}

// ValidateApprovalAuthority validates if a user can approve a transaction
func (e *transactionWorkflowEngine) ValidateApprovalAuthority(ctx context.Context, req AuthorityValidationRequest) (*AuthorityValidationResult, error) {
	ctx, span := e.tracing.StartSpan(ctx, "TransactionWorkflowEngine.ValidateApprovalAuthority")
	defer span.End()

	result := &AuthorityValidationResult{
		HasAuthority:      true, // Simplified - assume authority for now
		AuthorityLevel:    req.ApprovalLevel,
		AuthorityTypes:    []ApprovalType{req.ApprovalType},
		ValidationReasons: []string{},
		Restrictions:      []string{},
	}

	// TODO: Implement proper authority validation
	// This would check user roles, permissions, delegation rules, etc.

	return result, nil
}

// Helper methods

func (e *transactionWorkflowEngine) determineApprovalWorkflow(ctx context.Context, transaction *domain.Transaction, approvalType ApprovalType, priority Priority) ([]WorkflowStep, error) {
	// Simplified workflow determination
	// In production, this would be more complex based on transaction amount, type, etc.

	steps := []WorkflowStep{
		{
			StepID:        uuid.New(),
			StepNumber:    1,
			StepName:      "Initial Approval",
			StepType:      "APPROVAL",
			Status:        WorkflowStatusPending,
			Approvers:     []ApproverInfo{},
			EstimatedTime: 24 * time.Hour,
		},
	}

	// Add additional steps for high-value transactions
	if transaction.TotalDebitAmount.GreaterThan(decimal.NewFromInt(10000)) {
		steps = append(steps, WorkflowStep{
			StepID:        uuid.New(),
			StepNumber:    2,
			StepName:      "Executive Approval",
			StepType:      "EXECUTIVE_APPROVAL",
			Status:        WorkflowStatusPending,
			Approvers:     []ApproverInfo{},
			EstimatedTime: 48 * time.Hour,
		})
	}

	return steps, nil
}

func (e *transactionWorkflowEngine) extractApproversFromSteps(steps []WorkflowStep) []ApproverInfo {
	var allApprovers []ApproverInfo
	for _, step := range steps {
		allApprovers = append(allApprovers, step.Approvers...)
	}
	return allApprovers
}

func (e *transactionWorkflowEngine) determineWorkflowProgress(ctx context.Context, transaction *domain.Transaction, currentLevel int32) (bool, int32, []ApproverInfo) {
	// Simplified logic - in production this would check actual workflow configuration

	// For now, assume single level approval for most transactions
	maxLevels := int32(1)

	// High value transactions need 2 levels
	if transaction.TotalDebitAmount.GreaterThan(decimal.NewFromInt(10000)) {
		maxLevels = 2
	}

	isComplete := currentLevel >= maxLevels
	nextLevel := currentLevel + 1

	// TODO: Load next approvers from configuration
	var nextApprovers []ApproverInfo

	return isComplete, nextLevel, nextApprovers
}
