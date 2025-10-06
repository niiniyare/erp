package workflow

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WorkflowOrchestrator defines the contract for transaction workflow orchestration
// This interface is owned by the finance core module, implemented by external packages
type WorkflowOrchestrator interface {
	// Transaction Processing Workflows
	ProcessTransaction(ctx context.Context, req ProcessTransactionRequest) (*ProcessResult, error)
	GetTransactionStatus(ctx context.Context, transactionID uuid.UUID) (*TransactionStatus, error)
	SubmitApproval(ctx context.Context, transactionID uuid.UUID, decision ApprovalDecision) error
	MonitorProgress(ctx context.Context, transactionID uuid.UUID) (*ProcessingProgress, error)

	// Batch Operations
	ProcessBatch(ctx context.Context, req BatchProcessRequest) (*BatchResult, error)

	// Reconciliation Workflows
	StartReconciliation(ctx context.Context, req ReconciliationRequest) (*ReconciliationResult, error)

	// Workflow Management
	CancelWorkflow(ctx context.Context, transactionID uuid.UUID, reason string) error
	RetryFailedWorkflow(ctx context.Context, transactionID uuid.UUID) error
}

// ProcessTransactionRequest contains all data needed to start transaction processing
type ProcessTransactionRequest struct {
	TransactionID    uuid.UUID                 `json:"transaction_id"`
	TransactionType  string                    `json:"transaction_type"`
	Amount           decimal.Decimal           `json:"amount"`
	Currency         string                    `json:"currency"`
	Description      string                    `json:"description"`
	PostingDate      time.Time                 `json:"posting_date"`
	Entries          []TransactionEntryRequest `json:"entries"`
	RequiresApproval bool                      `json:"requires_approval"`
	ApprovalLevel    ApprovalLevel             `json:"approval_level,omitempty"`
	Priority         ProcessingPriority        `json:"priority"`
	Metadata         map[string]any            `json:"metadata,omitempty"`
}

// TransactionEntryRequest represents individual transaction entries
type TransactionEntryRequest struct {
	AccountID    uuid.UUID       `json:"account_id"`
	Description  string          `json:"description"`
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`
	CostCenterID *uuid.UUID      `json:"cost_center_id,omitempty"`
	DepartmentID *uuid.UUID      `json:"department_id,omitempty"`
	ProjectID    *uuid.UUID      `json:"project_id,omitempty"`
}

// ProcessResult contains immediate response from workflow initiation
type ProcessResult struct {
	TransactionID       uuid.UUID     `json:"transaction_id"`
	WorkflowReference   string        `json:"workflow_reference"` // Internal reference, not exposed to API
	Status              string        `json:"status"`
	EstimatedCompletion time.Duration `json:"estimated_completion"`
	NextAction          string        `json:"next_action,omitempty"`
}

// TransactionStatus provides status information
type TransactionStatus struct {
	TransactionID    uuid.UUID        `json:"transaction_id"`
	Status           ProcessingStatus `json:"status"`
	CurrentStage     ProcessingStage  `json:"current_stage"`
	Progress         int              `json:"progress"` // 0-100 percentage
	RequiresApproval bool             `json:"requires_approval"`
	ApprovalHistory  []ApprovalRecord `json:"approval_history"`
	ProcessingSteps  []ProcessingStep `json:"processing_steps"`
	ErrorDetails     *ErrorDetails    `json:"error_details,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
}

// ApprovalDecision contains approval/rejection decision
type ApprovalDecision struct {
	Decision   ApprovalType `json:"decision"`
	Comments   string       `json:"comments"`
	ApproverID uuid.UUID    `json:"approver_id"`
	ApprovedAt time.Time    `json:"approved_at"`
	IPAddress  string       `json:"ip_address,omitempty"`
	UserAgent  string       `json:"user_agent,omitempty"`
}

// ProcessingProgress provides detailed progress information
type ProcessingProgress struct {
	TransactionID  uuid.UUID        `json:"transaction_id"`
	TotalSteps     int              `json:"total_steps"`
	CompletedSteps int              `json:"completed_steps"`
	CurrentStep    ProcessingStep   `json:"current_step"`
	RemainingSteps []ProcessingStep `json:"remaining_steps"`
	EstimatedTime  time.Duration    `json:"estimated_time"`
}

// BatchProcessRequest for bulk transaction processing
type BatchProcessRequest struct {
	BatchID      uuid.UUID                   `json:"batch_id"`
	Transactions []ProcessTransactionRequest `json:"transactions"`
	BatchOptions BatchOptions                `json:"batch_options"`
}

// BatchResult contains batch processing results
type BatchResult struct {
	BatchID        uuid.UUID           `json:"batch_id"`
	TotalCount     int                 `json:"total_count"`
	SuccessCount   int                 `json:"success_count"`
	FailureCount   int                 `json:"failure_count"`
	Results        []TransactionResult `json:"results"`
	ProcessingTime time.Duration       `json:"processing_time"`
}

// ReconciliationRequest for bank reconciliation workflows
type ReconciliationRequest struct {
	AccountID     uuid.UUID       `json:"account_id"`
	StatementDate time.Time       `json:"statement_date"`
	StatementData []byte          `json:"statement_data"`
	ReconcileType string          `json:"reconcile_type"`
	AutoMatch     bool            `json:"auto_match"`
	Tolerance     decimal.Decimal `json:"tolerance"`
}

// ReconciliationResult contains reconciliation workflow results
type ReconciliationResult struct {
	ReconciliationID uuid.UUID `json:"reconciliation_id"`
	Status           string    `json:"status"`
	MatchedCount     int       `json:"matched_count"`
	UnmatchedCount   int       `json:"unmatched_count"`
}

// Supporting Types and Enums

type ProcessingStatus string

const (
	ProcessingStatusDraft      ProcessingStatus = "draft"
	ProcessingStatusValidating ProcessingStatus = "validating"
	ProcessingStatusPending    ProcessingStatus = "pending_approval"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusPosting    ProcessingStatus = "posting"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
	ProcessingStatusCancelled  ProcessingStatus = "cancelled"
)

type ProcessingStage string

const (
	StageValidation     ProcessingStage = "validation"
	StageApproval       ProcessingStage = "approval"
	StagePosting        ProcessingStage = "posting"
	StageBalanceUpdate  ProcessingStage = "balance_update"
	StageNotification   ProcessingStage = "notification"
	StageReconciliation ProcessingStage = "reconciliation"
	StageCompletion     ProcessingStage = "completion"
)

type ApprovalType string

const (
	ApprovalTypeApprove  ApprovalType = "approve"
	ApprovalTypeReject   ApprovalType = "reject"
	ApprovalTypeDelegate ApprovalType = "delegate"
)

type ApprovalLevel string

const (
	ApprovalLevelNone      ApprovalLevel = "none"
	ApprovalLevelManager   ApprovalLevel = "manager"
	ApprovalLevelSenior    ApprovalLevel = "senior"
	ApprovalLevelExecutive ApprovalLevel = "executive"
)

type ProcessingPriority string

const (
	PriorityLow    ProcessingPriority = "low"
	PriorityNormal ProcessingPriority = "normal"
	PriorityHigh   ProcessingPriority = "high"
	PriorityUrgent ProcessingPriority = "urgent"
)

// ApprovalRecord tracks approval history
type ApprovalRecord struct {
	ID           uuid.UUID     `json:"id"`
	ApproverID   uuid.UUID     `json:"approver_id"`
	ApproverName string        `json:"approver_name"`
	Decision     ApprovalType  `json:"decision"`
	Comments     string        `json:"comments"`
	ApprovedAt   time.Time     `json:"approved_at"`
	Level        ApprovalLevel `json:"level"`
}

// ProcessingStep represents individual workflow step
type ProcessingStep struct {
	StepName    string          `json:"step_name"`
	Stage       ProcessingStage `json:"stage"`
	Status      string          `json:"status"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Duration    *time.Duration  `json:"duration,omitempty"`
	ErrorMsg    string          `json:"error_message,omitempty"`
}

// ErrorDetails provides detailed error information
type ErrorDetails struct {
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details"`
	Recoverable bool      `json:"recoverable"`
	RetryCount  int       `json:"retry_count"`
	LastRetry   time.Time `json:"last_retry"`
}

// TransactionResult for batch processing results
type TransactionResult struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Success       bool      `json:"success"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// BatchOptions for batch processing configuration
type BatchOptions struct {
	ConcurrentLimit int           `json:"concurrent_limit"`
	RetryAttempts   int           `json:"retry_attempts"`
	Timeout         time.Duration `json:"timeout"`
	FailFast        bool          `json:"fail_fast"`
}
