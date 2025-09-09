package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Account Activity Types
type AccountCodeValidationInput struct {
	AccountCode string `json:"account_code"`
	AccountType string `json:"account_type"`
}

type AccountCodeValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type AccountHierarchyValidationInput struct {
	AccountCode  string `json:"account_code"`
	ParentCode   string `json:"parent_code,omitempty"`
	AccountType  string `json:"account_type"`
	AccountClass string `json:"account_class"`
}

type AccountHierarchyValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type AccountCreationInput struct {
	AccountCode        string `json:"account_code"`
	AccountName        string `json:"account_name"`
	AccountType        string `json:"account_type"`
	AccountClass       string `json:"account_class"`
	ParentAccountCode  string `json:"parent_account_code,omitempty"`
	CurrencyCode       string `json:"currency_code"`
	Description        string `json:"description,omitempty"`
	IsActive          bool   `json:"is_active"`
	AllowManualJournal bool   `json:"allow_manual_journal"`
	CreatedBy         string `json:"created_by"`
}

type AccountCreationResult struct {
	AccountID uuid.UUID `json:"account_id"`
}

type AccountNotificationInput struct {
	AccountID   uuid.UUID `json:"account_id"`
	AccountCode string    `json:"account_code"`
	Action      string    `json:"action"`
	Recipients  []string  `json:"recipients"`
}

type AccountBalanceCheckInput struct {
	AccountID uuid.UUID `json:"account_id"`
}

type AccountBalanceCheckResult struct {
	IsZero  bool            `json:"is_zero"`
	Balance decimal.Decimal `json:"balance"`
}

type AccountDependencyCheckInput struct {
	AccountID uuid.UUID `json:"account_id"`
}

type AccountDependencyCheckResult struct {
	HasDependencies bool     `json:"has_dependencies"`
	Dependencies   []string `json:"dependencies,omitempty"`
}

type AccountClosureInput struct {
	AccountID     uuid.UUID `json:"account_id"`
	ClosureReason string    `json:"closure_reason"`
	ClosedBy      string    `json:"closed_by"`
	ClosureDate   time.Time `json:"closure_date"`
}

type AccountTransactionDataInput struct {
	AccountID  uuid.UUID `json:"account_id"`
	PeriodFrom time.Time `json:"period_from"`
	PeriodTo   time.Time `json:"period_to"`
}

type AccountTransactionDataResult struct {
	Transactions   []TransactionDetail  `json:"transactions"`
	OpeningBalance decimal.Decimal      `json:"opening_balance"`
}

type TransactionDetail struct {
	TransactionID uuid.UUID       `json:"transaction_id"`
	Date          time.Time       `json:"date"`
	Amount        decimal.Decimal `json:"amount"`
	Description   string          `json:"description"`
	Reference     string          `json:"reference"`
}

type BalanceCalculationInput struct {
	AccountID      uuid.UUID            `json:"account_id"`
	Transactions   []TransactionDetail  `json:"transactions"`
	OpeningBalance decimal.Decimal      `json:"opening_balance"`
}

type BalanceCalculationResult struct {
	ExpectedBalance decimal.Decimal `json:"expected_balance"`
}

type ActualBalanceInput struct {
	AccountID uuid.UUID `json:"account_id"`
	AsOfDate  time.Time `json:"as_of_date"`
}

type ActualBalanceResult struct {
	Balance decimal.Decimal `json:"balance"`
}

type ReconciliationReportInput struct {
	AccountID       uuid.UUID            `json:"account_id"`
	Period          string               `json:"period"`
	ExpectedBalance decimal.Decimal      `json:"expected_balance"`
	ActualBalance   decimal.Decimal      `json:"actual_balance"`
	Discrepancy     decimal.Decimal      `json:"discrepancy"`
	Transactions    []TransactionDetail  `json:"transactions"`
}

type ReconciliationReportResult struct {
	ReportID   uuid.UUID `json:"report_id"`
	ReportPath string    `json:"report_path"`
}

// Compliance and Audit Activity Types
type AuditInitializationInput struct {
	AuditType       string   `json:"audit_type"`
	AuditPeriod     string   `json:"audit_period"`
	AuditScope      []string `json:"audit_scope"`
	InitiatedBy     string   `json:"initiated_by"`
	ComplianceRules []string `json:"compliance_rules"`
}

type AuditInitializationResult struct {
	AuditID uuid.UUID `json:"audit_id"`
}

type AuditDataCollectionInput struct {
	AuditID    uuid.UUID `json:"audit_id"`
	AuditType  string    `json:"audit_type"`
	AuditScope []string  `json:"audit_scope"`
	PeriodFrom time.Time `json:"period_from"`
	PeriodTo   time.Time `json:"period_to"`
}

type AuditDataCollectionResult struct {
	CollectedData map[string]any `json:"collected_data"`
}

type ComplianceCheckInput struct {
	AuditID         uuid.UUID      `json:"audit_id"`
	ComplianceRules []string       `json:"compliance_rules"`
	AuditData       map[string]any `json:"audit_data"`
}

type ComplianceCheckResults struct {
	ComplianceScore float64               `json:"compliance_score"`
	Violations      []ComplianceViolation `json:"violations"`
	Recommendations []string              `json:"recommendations"`
	MaxSeverity     string                `json:"max_severity"`
}

type AuditReportInput struct {
	AuditID         uuid.UUID             `json:"audit_id"`
	AuditType       string                `json:"audit_type"`
	ComplianceScore float64               `json:"compliance_score"`
	Violations      []ComplianceViolation `json:"violations"`
	Recommendations []string              `json:"recommendations"`
}

type AuditReportResult struct {
	ReportID   uuid.UUID `json:"report_id"`
	ReportPath string    `json:"report_path"`
}

type ComplianceAlertInput struct {
	AuditID     uuid.UUID             `json:"audit_id"`
	AlertType   AlertType             `json:"alert_type"`
	Violations  []ComplianceViolation `json:"violations"`
	Severity    string                `json:"severity"`
	Recipients  []string              `json:"recipients"`
}

// Fraud Detection Activity Types
type FraudDetectionInitInput struct {
	DetectionType    string            `json:"detection_type"`
	DetectionRules   []string          `json:"detection_rules"`
	MonitoringPeriod string            `json:"monitoring_period"`
	InitiatedBy      string            `json:"initiated_by"`
}

type FraudDetectionInitResult struct {
	DetectionID uuid.UUID `json:"detection_id"`
}

type FraudPatternAnalysisInput struct {
	DetectionID     uuid.UUID               `json:"detection_id"`
	DetectionRules  []string                `json:"detection_rules"`
	TransactionData []TransactionDataPoint  `json:"transaction_data"`
	HistoricalData  []HistoricalDataPoint   `json:"historical_data"`
}

type FraudPatternAnalysisResult struct {
	DetectedPatterns []FraudPattern `json:"detected_patterns"`
	RiskFactors      []string       `json:"risk_factors"`
}

type FraudPattern struct {
	PatternID   uuid.UUID `json:"pattern_id"`
	PatternType string    `json:"pattern_type"`
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	Severity    string    `json:"severity"`
}

type FraudRiskScoringInput struct {
	DetectionID uuid.UUID      `json:"detection_id"`
	Patterns    []FraudPattern `json:"patterns"`
	RiskFactors []string       `json:"risk_factors"`
}

type FraudRiskScoringResult struct {
	HighRiskTransactions []HighRiskTransaction `json:"high_risk_transactions"`
	TopRiskFactors       []string              `json:"top_risk_factors"`
}

type HighRiskTransaction struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	RiskScore     float64   `json:"risk_score"`
	RiskFactors   []string  `json:"risk_factors"`
}

type FraudAlertNotificationInput struct {
	DetectionID uuid.UUID     `json:"detection_id"`
	Alert       FraudAlert    `json:"alert"`
	Recipients  []string      `json:"recipients"`
	Priority    AlertPriority `json:"priority"`
}

type FraudDetectionReportInput struct {
	DetectionID          uuid.UUID    `json:"detection_id"`
	DetectionType        string       `json:"detection_type"`
	AnalyzedTransactions int          `json:"analyzed_transactions"`
	HighRiskTransactions int          `json:"high_risk_transactions"`
	Alerts               []FraudAlert `json:"alerts"`
	RiskFactors          []string     `json:"risk_factors"`
}

type FraudDetectionReportResult struct {
	ReportID   uuid.UUID `json:"report_id"`
	ReportPath string    `json:"report_path"`
}

// Transaction Processing Activity Types
type TransactionReversalInput struct {
	OriginalTransactionID uuid.UUID `json:"original_transaction_id"`
	ReversalReason        string    `json:"reversal_reason"`
}

type ReversalValidationInput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Reason        string    `json:"reason"`
}

type ReversalValidationResult struct {
	IsEligible       bool     `json:"is_eligible"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type ReversalCreationInput struct {
	OriginalTransactionID uuid.UUID `json:"original_transaction_id"`
	ReversalReason        string    `json:"reversal_reason"`
	InitiatedBy          string    `json:"initiated_by"`
}

type ReversalCreationResult struct {
	ReversalTransactionID uuid.UUID `json:"reversal_transaction_id"`
}

// Period Closing Activity Types
type PreClosingValidationInput struct {
	ClosingPeriod   string    `json:"closing_period"`
	TenantID        uuid.UUID `json:"tenant_id"`
	ValidationRules []string  `json:"validation_rules"`
}

type PreClosingValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type AccountReconciliationBatchInput struct {
	ClosingPeriod string    `json:"closing_period"`
	TenantID      uuid.UUID `json:"tenant_id"`
	AccountFilter []string  `json:"account_filter,omitempty"`
}

type AccountReconciliationBatchResult struct {
	ProcessedAccounts int                                  `json:"processed_accounts"`
	DiscrepancyCount  int                                  `json:"discrepancy_count"`
	Results           []AccountReconciliationWorkflowResult `json:"results"`
}

type AdjustingEntriesInput struct {
	ClosingPeriod         string                            `json:"closing_period"`
	TenantID             uuid.UUID                         `json:"tenant_id"`
	ReconciliationResults []AccountReconciliationWorkflowResult `json:"reconciliation_results"`
	AdjustmentRules      []string                          `json:"adjustment_rules"`
}

type AdjustingEntriesResult struct {
	GeneratedEntries []AdjustingEntry `json:"generated_entries"`
}

type AdjustingEntry struct {
	EntryID     uuid.UUID       `json:"entry_id"`
	AccountID   uuid.UUID       `json:"account_id"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
}

type DepreciationCalculationInput struct {
	ClosingPeriod string    `json:"closing_period"`
	TenantID      uuid.UUID `json:"tenant_id"`
	AssetFilter   []string  `json:"asset_filter,omitempty"`
}

type DepreciationCalculationResult struct {
	ProcessedAssets     int                 `json:"processed_assets"`
	DepreciationEntries []DepreciationEntry `json:"depreciation_entries"`
}

type DepreciationEntry struct {
	EntryID           uuid.UUID       `json:"entry_id"`
	AssetID           uuid.UUID       `json:"asset_id"`
	DepreciationAmount decimal.Decimal `json:"depreciation_amount"`
	AccumulatedAmount  decimal.Decimal `json:"accumulated_amount"`
	Method            string          `json:"method"`
}

type ClosingEntriesInput struct {
	ClosingPeriod       string              `json:"closing_period"`
	TenantID           uuid.UUID           `json:"tenant_id"`
	AdjustingEntries   []AdjustingEntry    `json:"adjusting_entries"`
	DepreciationEntries []DepreciationEntry `json:"depreciation_entries"`
}

type ClosingEntriesResult struct {
	GeneratedEntries []ClosingEntry `json:"generated_entries"`
}

type ClosingEntry struct {
	EntryID     uuid.UUID       `json:"entry_id"`
	AccountID   uuid.UUID       `json:"account_id"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	EntryType   string          `json:"entry_type"`
}

type FinancialStatementsInput struct {
	ClosingPeriod  string    `json:"closing_period"`
	TenantID       uuid.UUID `json:"tenant_id"`
	StatementTypes []string  `json:"statement_types"`
}

type FinancialStatementsResult struct {
	GeneratedStatements []FinancialStatement `json:"generated_statements"`
}

type FinalizeClosingInput struct {
	ClosingPeriod        string               `json:"closing_period"`
	TenantID            uuid.UUID            `json:"tenant_id"`
	ClosingEntries      []ClosingEntry       `json:"closing_entries"`
	FinancialStatements []FinancialStatement `json:"financial_statements"`
	FinalizedBy         string               `json:"finalized_by"`
}

type ClosingNotificationInput struct {
	ClosingPeriod       string               `json:"closing_period"`
	ClosingStatus       ClosingStatus        `json:"closing_status"`
	FinancialStatements []FinancialStatement `json:"financial_statements"`
	Recipients          []string             `json:"recipients"`
}

// Year-End Closing Activity Types
type PreYearEndValidationInput struct {
	FiscalYear      string    `json:"fiscal_year"`
	TenantID        uuid.UUID `json:"tenant_id"`
	ValidationRules []string  `json:"validation_rules"`
}

type PreYearEndValidationResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type MonthlyClosingCheckInput struct {
	FiscalYear string    `json:"fiscal_year"`
	TenantID   uuid.UUID `json:"tenant_id"`
}

type MonthlyClosingCheckResult struct {
	AllMonthsClosed bool     `json:"all_months_closed"`
	UnclosedMonths  []string `json:"unclosed_months,omitempty"`
}

type AnnualDepreciationInput struct {
	FiscalYear  string    `json:"fiscal_year"`
	TenantID    uuid.UUID `json:"tenant_id"`
	AssetFilter []string  `json:"asset_filter,omitempty"`
}

type AnnualDepreciationResult struct {
	ProcessedAssets     int                 `json:"processed_assets"`
	DepreciationEntries []DepreciationEntry `json:"depreciation_entries"`
}

type YearEndAccrualsInput struct {
	FiscalYear   string    `json:"fiscal_year"`
	TenantID     uuid.UUID `json:"tenant_id"`
	AccrualRules []string  `json:"accrual_rules"`
}

type YearEndAccrualsResult struct {
	ProcessedAccruals []AccrualEntry `json:"processed_accruals"`
	AccrualEntries    []AccrualEntry `json:"accrual_entries"`
}

type AccrualEntry struct {
	EntryID      uuid.UUID       `json:"entry_id"`
	AccountID    uuid.UUID       `json:"account_id"`
	Amount       decimal.Decimal `json:"amount"`
	Description  string          `json:"description"`
	AccrualType  string          `json:"accrual_type"`
	PeriodStart  time.Time       `json:"period_start"`
	PeriodEnd    time.Time       `json:"period_end"`
}

type YearEndClosingEntriesInput struct {
	FiscalYear          string              `json:"fiscal_year"`
	TenantID           uuid.UUID           `json:"tenant_id"`
	DepreciationEntries []DepreciationEntry `json:"depreciation_entries"`
	AccrualEntries     []AccrualEntry      `json:"accrual_entries"`
}

type YearEndClosingEntriesResult struct {
	GeneratedEntries []ClosingEntry `json:"generated_entries"`
}

type AnnualFinancialStatementsInput struct {
	FiscalYear       string    `json:"fiscal_year"`
	TenantID         uuid.UUID `json:"tenant_id"`
	StatementTypes   []string  `json:"statement_types"`
	IncludePriorYear bool      `json:"include_prior_year"`
}

type AnnualFinancialStatementsResult struct {
	GeneratedStatements []FinancialStatement `json:"generated_statements"`
}

type FiscalYearArchiveInput struct {
	FiscalYear      string          `json:"fiscal_year"`
	TenantID        uuid.UUID       `json:"tenant_id"`
	ArchiveSettings ArchiveSettings `json:"archive_settings"`
}

type FiscalYearArchiveResult struct {
	ArchiveReference string `json:"archive_reference"`
	ArchivedRecords  int    `json:"archived_records"`
}

type FinalizeYearEndClosingInput struct {
	FiscalYear          string               `json:"fiscal_year"`
	TenantID           uuid.UUID            `json:"tenant_id"`
	ClosingEntries     []ClosingEntry       `json:"closing_entries"`
	FinancialStatements []FinancialStatement `json:"financial_statements"`
	ArchiveReference   string               `json:"archive_reference"`
	FinalizedBy        string               `json:"finalized_by"`
}