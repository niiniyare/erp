package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Common workflow types and enums (avoid duplicates with existing types)
type ProcessingStatus string
type ReversalStatus string
type AccountCreationStatus string
type AccountClosureStatus string
type ReconciliationStatus string
type AuditStatus string
type FraudDetectionStatus string
type ClosingStatus string
type StepStatus string

const (
	// Note: ApprovalStatus already defined in types.go

	// Processing statuses
	ProcessingStatusPending   ProcessingStatus = "pending"
	ProcessingStatusCompleted ProcessingStatus = "completed"
	ProcessingStatusFailed    ProcessingStatus = "failed"

	// Reversal statuses
	ReversalStatusPending   ReversalStatus = "pending"
	ReversalStatusCompleted ReversalStatus = "completed"
	ReversalStatusRejected  ReversalStatus = "rejected"
	ReversalStatusFailed    ReversalStatus = "failed"

	// Account creation statuses
	AccountCreationStatusPending   AccountCreationStatus = "pending"
	AccountCreationStatusCompleted AccountCreationStatus = "completed"
	AccountCreationStatusFailed    AccountCreationStatus = "failed"

	// Account closure statuses
	AccountClosureStatusPending   AccountClosureStatus = "pending"
	AccountClosureStatusCompleted AccountClosureStatus = "completed"
	AccountClosureStatusFailed    AccountClosureStatus = "failed"

	// Reconciliation statuses
	ReconciliationStatusPending     ReconciliationStatus = "pending"
	ReconciliationStatusReconciled  ReconciliationStatus = "reconciled"
	ReconciliationStatusDiscrepancy ReconciliationStatus = "discrepancy"
	ReconciliationStatusFailed      ReconciliationStatus = "failed"

	// Audit statuses
	AuditStatusPending   AuditStatus = "pending"
	AuditStatusCompleted AuditStatus = "completed"
	AuditStatusFailed    AuditStatus = "failed"

	// Fraud detection statuses
	FraudDetectionStatusPending   FraudDetectionStatus = "pending"
	FraudDetectionStatusCompleted FraudDetectionStatus = "completed"
	FraudDetectionStatusFailed    FraudDetectionStatus = "failed"

	// Closing statuses
	ClosingStatusPending   ClosingStatus = "pending"
	ClosingStatusCompleted ClosingStatus = "completed"
	ClosingStatusFailed    ClosingStatus = "failed"

	// Step statuses
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// Transaction Approval Workflow Types
type TransactionApprovalWorkflowInput struct {
	TransactionID uuid.UUID      `json:"transaction_id"`
	SubmittedBy   string         `json:"submitted_by"`
	SubmittedAt   time.Time      `json:"submitted_at"`
	Amount        decimal.Decimal `json:"amount"`
	AccountType   string         `json:"account_type"`
}

type TransactionApprovalWorkflowResult struct {
	TransactionID     uuid.UUID       `json:"transaction_id"`
	Status           ApprovalStatus  `json:"status"`
	ApprovedBy       []string        `json:"approved_by,omitempty"`
	RejectedBy       []string        `json:"rejected_by,omitempty"`
	Comments         []string        `json:"comments,omitempty"`
	ValidationErrors []string        `json:"validation_errors,omitempty"`
	StartTime        time.Time       `json:"start_time"`
	CompletedTime    time.Time       `json:"completed_time"`
	Duration         time.Duration   `json:"duration"`
	Error            string          `json:"error,omitempty"`
}

// Transaction Processing Workflow Types
type TransactionProcessingWorkflowInput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
}

type TransactionProcessingWorkflowResult struct {
	TransactionID    uuid.UUID        `json:"transaction_id"`
	Status          ProcessingStatus `json:"status"`
	PostingReference string           `json:"posting_reference,omitempty"`
	StartTime       time.Time        `json:"start_time"`
	CompletedTime   time.Time        `json:"completed_time"`
	Duration        time.Duration    `json:"duration"`
	Error           string           `json:"error,omitempty"`
}

// Transaction Reversal Workflow Types
type TransactionReversalWorkflowInput struct {
	OriginalTransactionID uuid.UUID `json:"original_transaction_id"`
	ReversalReason       string    `json:"reversal_reason"`
	InitiatedBy          string    `json:"initiated_by"`
}

type TransactionReversalWorkflowResult struct {
	OriginalTransactionID uuid.UUID      `json:"original_transaction_id"`
	ReversalTransactionID uuid.UUID      `json:"reversal_transaction_id"`
	Status               ReversalStatus `json:"status"`
	PostingReference     string         `json:"posting_reference,omitempty"`
	StartTime           time.Time      `json:"start_time"`
	CompletedTime       time.Time      `json:"completed_time"`
	Duration            time.Duration  `json:"duration"`
	Error               string         `json:"error,omitempty"`
}

// Bulk Transaction Workflow Types
type BulkTransactionWorkflowInput struct {
	BatchID        string      `json:"batch_id"`
	TransactionIDs []uuid.UUID `json:"transaction_ids"`
	Concurrency    int         `json:"concurrency,omitempty"`
}

type BulkTransactionWorkflowResult struct {
	BatchID                string                                        `json:"batch_id"`
	Results                map[string]TransactionProcessingWorkflowResult `json:"results"`
	SuccessfulTransactions []uuid.UUID                                   `json:"successful_transactions"`
	FailedTransactions     []uuid.UUID                                   `json:"failed_transactions"`
	TotalProcessed         int                                           `json:"total_processed"`
	SuccessRate           float64                                       `json:"success_rate"`
	StartTime             time.Time                                     `json:"start_time"`
	CompletedTime         time.Time                                     `json:"completed_time"`
	Duration              time.Duration                                 `json:"duration"`
}

// Account Creation Workflow Types
type AccountCreationWorkflowInput struct {
	AccountCode        string      `json:"account_code"`
	AccountName        string      `json:"account_name"`
	AccountType        string      `json:"account_type"`
	AccountClass       string      `json:"account_class"`
	ParentAccountCode  string      `json:"parent_account_code,omitempty"`
	CurrencyCode       string      `json:"currency_code"`
	Description        string      `json:"description,omitempty"`
	IsActive          bool        `json:"is_active"`
	AllowManualJournal bool        `json:"allow_manual_journal"`
	CreatedBy         string      `json:"created_by"`
}

type AccountCreationWorkflowResult struct {
	AccountCode      string                `json:"account_code"`
	AccountID        uuid.UUID             `json:"account_id"`
	Status          AccountCreationStatus `json:"status"`
	ValidationErrors []string              `json:"validation_errors,omitempty"`
	StartTime       time.Time             `json:"start_time"`
	CompletedTime   time.Time             `json:"completed_time"`
	Duration        time.Duration         `json:"duration"`
	Error           string                `json:"error,omitempty"`
}

// Account Closure Workflow Types
type AccountClosureWorkflowInput struct {
	AccountID     uuid.UUID `json:"account_id"`
	ClosureReason string    `json:"closure_reason"`
	ClosedBy      string    `json:"closed_by"`
	ClosureDate   time.Time `json:"closure_date"`
}

type AccountClosureWorkflowResult struct {
	AccountID    uuid.UUID            `json:"account_id"`
	Status      AccountClosureStatus `json:"status"`
	Dependencies []string             `json:"dependencies,omitempty"`
	StartTime   time.Time            `json:"start_time"`
	CompletedTime time.Time           `json:"completed_time"`
	Duration    time.Duration        `json:"duration"`
	Error       string               `json:"error,omitempty"`
}

// Account Reconciliation Workflow Types
type AccountReconciliationWorkflowInput struct {
	AccountID            uuid.UUID `json:"account_id"`
	ReconciliationPeriod string    `json:"reconciliation_period"`
	PeriodFrom          time.Time `json:"period_from"`
	PeriodTo            time.Time `json:"period_to"`
}

type AccountReconciliationWorkflowResult struct {
	AccountID       uuid.UUID            `json:"account_id"`
	Period         string               `json:"period"`
	Status         ReconciliationStatus `json:"status"`
	ExpectedBalance decimal.Decimal      `json:"expected_balance"`
	ActualBalance  decimal.Decimal      `json:"actual_balance"`
	Discrepancy    decimal.Decimal      `json:"discrepancy"`
	ReportID       uuid.UUID            `json:"report_id,omitempty"`
	ReportPath     string               `json:"report_path,omitempty"`
	StartTime      time.Time            `json:"start_time"`
	CompletedTime  time.Time            `json:"completed_time"`
	Duration       time.Duration        `json:"duration"`
	Error          string               `json:"error,omitempty"`
}

// Compliance Audit Workflow Types
type ComplianceAuditWorkflowInput struct {
	AuditType               string            `json:"audit_type"`
	AuditPeriod            string            `json:"audit_period"`
	AuditScope             []string          `json:"audit_scope"`
	ComplianceRules        []string          `json:"compliance_rules"`
	PeriodFrom             time.Time         `json:"period_from"`
	PeriodTo               time.Time         `json:"period_to"`
	InitiatedBy            string            `json:"initiated_by"`
	NotificationRecipients []string          `json:"notification_recipients"`
}

type ComplianceAuditWorkflowResult struct {
	AuditID        uuid.UUID              `json:"audit_id"`
	AuditType      string                 `json:"audit_type"`
	Period         string                 `json:"period"`
	Status         AuditStatus            `json:"status"`
	ComplianceScore float64               `json:"compliance_score"`
	ViolationCount int                    `json:"violation_count"`
	Violations     []ComplianceViolation  `json:"violations,omitempty"`
	ReportID       uuid.UUID              `json:"report_id,omitempty"`
	ReportPath     string                 `json:"report_path,omitempty"`
	StartTime      time.Time              `json:"start_time"`
	CompletedTime  time.Time              `json:"completed_time"`
	Duration       time.Duration          `json:"duration"`
	Error          string                 `json:"error,omitempty"`
}

// Fraud Detection Workflow Types
type FraudDetectionWorkflowInput struct {
	DetectionType       string                    `json:"detection_type"`
	DetectionRules      []string                  `json:"detection_rules"`
	MonitoringPeriod    string                    `json:"monitoring_period"`
	TransactionData     []TransactionDataPoint    `json:"transaction_data"`
	HistoricalData      []HistoricalDataPoint     `json:"historical_data"`
	AlertThreshold      float64                   `json:"alert_threshold"`
	CriticalThreshold   float64                   `json:"critical_threshold"`
	AlertRecipients     []string                  `json:"alert_recipients"`
	InitiatedBy         string                    `json:"initiated_by"`
}

type FraudDetectionWorkflowResult struct {
	DetectionID          uuid.UUID             `json:"detection_id"`
	DetectionType        string                `json:"detection_type"`
	Status              FraudDetectionStatus  `json:"status"`
	AnalyzedTransactions int                   `json:"analyzed_transactions"`
	HighRiskTransactions int                   `json:"high_risk_transactions"`
	Alerts              []FraudAlert          `json:"alerts"`
	RiskFactors         []string              `json:"risk_factors"`
	ReportID            uuid.UUID             `json:"report_id,omitempty"`
	ReportPath          string                `json:"report_path,omitempty"`
	StartTime           time.Time             `json:"start_time"`
	CompletedTime       time.Time             `json:"completed_time"`
	Duration            time.Duration         `json:"duration"`
	Error               string                `json:"error,omitempty"`
}

// Month-End and Year-End Closing Workflow Types
type MonthEndClosingWorkflowInput struct {
	ClosingPeriod           string            `json:"closing_period"`
	TenantID                uuid.UUID         `json:"tenant_id"`
	ValidationRules         []string          `json:"validation_rules"`
	AccountFilter           []string          `json:"account_filter,omitempty"`
	AdjustmentRules         []string          `json:"adjustment_rules"`
	AssetFilter             []string          `json:"asset_filter,omitempty"`
	StatementTypes          []string          `json:"statement_types"`
	NotificationRecipients  []string          `json:"notification_recipients"`
	InitiatedBy             string            `json:"initiated_by"`
}

type MonthEndClosingWorkflowResult struct {
	ClosingPeriod      string                      `json:"closing_period"`
	Status            ClosingStatus               `json:"status"`
	Steps             map[string]ClosingStepResult `json:"steps"`
	FinancialStatements []FinancialStatement       `json:"financial_statements,omitempty"`
	ValidationErrors   []string                    `json:"validation_errors,omitempty"`
	StartTime         time.Time                   `json:"start_time"`
	CompletedTime     time.Time                   `json:"completed_time"`
	Duration          time.Duration               `json:"duration"`
	Error             string                      `json:"error,omitempty"`
}

type YearEndClosingWorkflowInput struct {
	FiscalYear             string              `json:"fiscal_year"`
	TenantID               uuid.UUID           `json:"tenant_id"`
	ValidationRules        []string            `json:"validation_rules"`
	AssetFilter            []string            `json:"asset_filter,omitempty"`
	AccrualRules           []string            `json:"accrual_rules"`
	StatementTypes         []string            `json:"statement_types"`
	ArchiveSettings        ArchiveSettings     `json:"archive_settings"`
	NotificationRecipients []string            `json:"notification_recipients"`
	InitiatedBy            string              `json:"initiated_by"`
}

type YearEndClosingWorkflowResult struct {
	FiscalYear          string                      `json:"fiscal_year"`
	Status             ClosingStatus               `json:"status"`
	Steps              map[string]ClosingStepResult `json:"steps"`
	FinancialStatements []FinancialStatement        `json:"financial_statements,omitempty"`
	ArchiveReference    string                      `json:"archive_reference,omitempty"`
	UnclosedMonths     []string                    `json:"unclosed_months,omitempty"`
	ValidationErrors   []string                    `json:"validation_errors,omitempty"`
	StartTime          time.Time                   `json:"start_time"`
	CompletedTime      time.Time                   `json:"completed_time"`
	Duration           time.Duration               `json:"duration"`
	Error              string                      `json:"error,omitempty"`
}

// Supporting types for workflow activities
type ClosingStepResult struct {
	StepName         string     `json:"step_name"`
	Status           StepStatus `json:"status"`
	CompletedAt      time.Time  `json:"completed_at"`
	ProcessedCount   int        `json:"processed_count,omitempty"`
	DiscrepancyCount int        `json:"discrepancy_count,omitempty"`
	Error            string     `json:"error,omitempty"`
}

type ComplianceViolation struct {
	ViolationID   uuid.UUID `json:"violation_id"`
	RuleID        string    `json:"rule_id"`
	Description   string    `json:"description"`
	Severity      string    `json:"severity"`
	TransactionID uuid.UUID `json:"transaction_id,omitempty"`
	AccountID     uuid.UUID `json:"account_id,omitempty"`
	Amount        decimal.Decimal `json:"amount,omitempty"`
	DetectedAt    time.Time `json:"detected_at"`
}

type FraudAlert struct {
	AlertID       uuid.UUID         `json:"alert_id"`
	TransactionID uuid.UUID         `json:"transaction_id"`
	AlertType     FraudAlertType    `json:"alert_type"`
	RiskScore     float64           `json:"risk_score"`
	RiskFactors   []string          `json:"risk_factors"`
	Status        FraudAlertStatus  `json:"status"`
	DetectedAt    time.Time         `json:"detected_at"`
	ReviewedAt    time.Time         `json:"reviewed_at,omitempty"`
	ReviewedBy    string            `json:"reviewed_by,omitempty"`
}

type FraudAlertType string
type FraudAlertStatus string
type AlertType string
type AlertPriority string

const (
	FraudAlertTypeHighRisk         FraudAlertType = "high_risk"
	FraudAlertTypeAnomalous        FraudAlertType = "anomalous"
	FraudAlertTypePatternMatch     FraudAlertType = "pattern_match"

	FraudAlertStatusActive         FraudAlertStatus = "active"
	FraudAlertStatusInvestigating  FraudAlertStatus = "investigating"
	FraudAlertStatusResolved       FraudAlertStatus = "resolved"
	FraudAlertStatusFalsePositive  FraudAlertStatus = "false_positive"

	AlertTypeViolation             AlertType = "violation"
	AlertTypeFraud                 AlertType = "fraud"
	AlertTypeCompliance            AlertType = "compliance"

	AlertPriorityLow               AlertPriority = "low"
	AlertPriorityMedium            AlertPriority = "medium"
	AlertPriorityHigh              AlertPriority = "high"
	AlertPriorityCritical          AlertPriority = "critical"
)

type TransactionDataPoint struct {
	TransactionID uuid.UUID       `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	AccountID     uuid.UUID       `json:"account_id"`
	Timestamp     time.Time       `json:"timestamp"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
}

type HistoricalDataPoint struct {
	Period        string          `json:"period"`
	AccountID     uuid.UUID       `json:"account_id"`
	AverageAmount decimal.Decimal `json:"average_amount"`
	TransactionCount int          `json:"transaction_count"`
	Patterns      []string        `json:"patterns"`
}

type FinancialStatement struct {
	StatementID   uuid.UUID `json:"statement_id"`
	StatementType string    `json:"statement_type"`
	Period        string    `json:"period"`
	FilePath      string    `json:"file_path"`
	Format        string    `json:"format"`
	GeneratedAt   time.Time `json:"generated_at"`
	GeneratedBy   string    `json:"generated_by"`
}

type ArchiveSettings struct {
	ArchiveFormat     string    `json:"archive_format"`
	ArchiveLocation   string    `json:"archive_location"`
	CompressionLevel  int       `json:"compression_level"`
	RetentionPeriod   time.Duration `json:"retention_period"`
	IncludeAttachments bool     `json:"include_attachments"`
}

// Signal types for approval workflows
type ApprovalSignal struct {
	ApproverID string `json:"approver_id"`
	Comment    string `json:"comment,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type RejectionSignal struct {
	RejectorID string `json:"rejector_id"`
	Comment    string `json:"comment"`
	Timestamp  time.Time `json:"timestamp"`
}

// Additional supporting input/result types referenced in workflows
type TransactionValidationInput struct {
	TransactionID   uuid.UUID     `json:"transaction_id"`
	ValidationType  ValidationType `json:"validation_type"`
}

type TransactionValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type ValidationType string

const (
	ValidationTypeApproval  ValidationType = "approval"
	ValidationTypePosting   ValidationType = "posting"
	ValidationTypeReversal  ValidationType = "reversal"
)

type ApprovalRequirementsInput struct {
	TransactionID uuid.UUID       `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	AccountType   string          `json:"account_type"`
}

type ApprovalRequirementsResult struct {
	RequiredApprovers    []string      `json:"required_approvers"`
	RequiredApprovalCount int          `json:"required_approval_count"`
	ApprovalTimeout      time.Duration `json:"approval_timeout"`
}

type ApprovalProcessResult struct {
	Status     ApprovalStatus `json:"status"`
	ApprovedBy []string       `json:"approved_by,omitempty"`
	RejectedBy []string       `json:"rejected_by,omitempty"`
	Comments   []string       `json:"comments,omitempty"`
}

// Placeholder types for activities - these would be expanded as activities are implemented
type DoubleEntryValidationInput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
}

type DoubleEntryValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type LedgerPostingInput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
}

type LedgerPostingResult struct {
	PostingReference  string              `json:"posting_reference"`
	ProcessedEntries []TransactionEntry  `json:"processed_entries"`
}

type BalanceUpdateInput struct {
	TransactionID uuid.UUID           `json:"transaction_id"`
	Entries      []TransactionEntry  `json:"entries"`
}

type ApprovalNotificationInput struct {
	TransactionID uuid.UUID      `json:"transaction_id"`
	Status       ApprovalStatus `json:"status"`
	Recipients   []string       `json:"recipients"`
	Comments     []string       `json:"comments,omitempty"`
}

// Add more activity input/result types as needed...
// This file can be expanded as more workflow activities are implemented