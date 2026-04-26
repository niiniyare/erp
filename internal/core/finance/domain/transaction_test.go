package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/finance/domain"
)

// ============================================================================
// Suite
// ============================================================================

type TransactionSuite struct {
	suite.Suite
	req *require.Assertions
}

func TestTransactionSuite(t *testing.T) {
	suite.Run(t, new(TransactionSuite))
}

func (s *TransactionSuite) SetupTest() {
	s.req = require.New(s.T())
}

// ============================================================================
// Helpers
// ============================================================================

// newBalancedEntries returns two entries (one debit, one credit) whose totals equal amount.
func newBalancedEntries(amount int64) []domain.TransactionEntry {
	tenantID := uuid.New()
	txnID := uuid.New()
	return []domain.TransactionEntry{
		{
			TenantID:      tenantID,
			TransactionID: txnID,
			EntryNumber:   1,
			AccountID:     uuid.New(),
			DebitAmount:   decimal.NewFromInt(amount),
			CreditAmount:  decimal.Zero,
			Description:   "Debit entry",
			ExchangeRate:  decimal.NewFromInt(1),
		},
		{
			TenantID:      tenantID,
			TransactionID: txnID,
			EntryNumber:   2,
			AccountID:     uuid.New(),
			DebitAmount:   decimal.Zero,
			CreditAmount:  decimal.NewFromInt(amount),
			Description:   "Credit entry",
			ExchangeRate:  decimal.NewFromInt(1),
		},
	}
}

// newValidTxn builds a minimal, valid Transaction ready for positive-path tests.
func newValidTxn() domain.Transaction {
	entries := newBalancedEntries(100)
	return domain.Transaction{
		TenantID:          uuid.New(),
		TransactionNumber: "TXN-2026-001",
		TransactionType:   domain.TransactionTypeManual,
		TransactionStatus: domain.TransactionStatusDraft,
		ApprovalStatus:    domain.ApprovalStatusNotRequired,
		TransactionDate:   time.Now(),
		Description:       "Test journal entry",
		CurrencyCode:      "USD",
		ExchangeRate:      decimal.NewFromInt(1),
		TotalDebitAmount:  decimal.NewFromInt(100),
		TotalCreditAmount: decimal.NewFromInt(100),
		Entries:           entries,
	}
}

// ============================================================================
// TransactionStatus state machine
// ============================================================================

func (s *TransactionSuite) TestIsEditable_OnlyDraftAndRejectedAreEditable() {
	cases := []struct {
		status   domain.TransactionStatus
		editable bool
	}{
		{domain.TransactionStatusDraft, true},
		{domain.TransactionStatusRejected, true},
		{domain.TransactionStatusPendingApproval, false},
		{domain.TransactionStatusApproved, false},
		{domain.TransactionStatusPosted, false},
		{domain.TransactionStatusCancelled, false},
		{domain.TransactionStatusReversed, false},
	}
	for _, tc := range cases {
		s.req.Equal(tc.editable, tc.status.IsEditable(), "status=%s", tc.status)
	}
}

func (s *TransactionSuite) TestTransactionStatus_AllDefinedStatuses_AreValid() {
	valid := []domain.TransactionStatus{
		domain.TransactionStatusDraft, domain.TransactionStatusPendingApproval,
		domain.TransactionStatusApproved, domain.TransactionStatusPosted,
		domain.TransactionStatusCancelled, domain.TransactionStatusReversed,
		domain.TransactionStatusRejected,
	}
	for _, st := range valid {
		s.req.True(st.IsValid(), "expected %s to be valid", st)
	}
	s.req.False(domain.TransactionStatus("UNKNOWN").IsValid())
}

func (s *TransactionSuite) TestTransactionStatus_IsActive_OnlyPosted() {
	s.req.True(domain.TransactionStatusPosted.IsActive())
	s.req.False(domain.TransactionStatusDraft.IsActive())
	s.req.False(domain.TransactionStatusApproved.IsActive())
}

// ============================================================================
// IsBalanced — double-entry invariant
// ============================================================================

func (s *TransactionSuite) TestIsBalanced_EqualDebitsAndCredits_True() {
	t := newValidTxn()
	s.req.True(t.IsBalanced())
}

func (s *TransactionSuite) TestIsBalanced_DebitsExceedCredits_False() {
	t := newValidTxn()
	t.TotalDebitAmount = decimal.NewFromInt(101)
	s.req.False(t.IsBalanced())
}

func (s *TransactionSuite) TestIsBalanced_CreditsExceedDebits_False() {
	t := newValidTxn()
	t.TotalCreditAmount = decimal.NewFromInt(101)
	s.req.False(t.IsBalanced())
}

func (s *TransactionSuite) TestIsBalanced_BothZero_True() {
	t := newValidTxn()
	t.TotalDebitAmount = decimal.Zero
	t.TotalCreditAmount = decimal.Zero
	s.req.True(t.IsBalanced())
}

func (s *TransactionSuite) TestIsBalanced_DecimalPrecision_MustBeExact() {
	// 100.001 ≠ 100.000 — even a fraction of a cent breaks the invariant
	t := newValidTxn()
	t.TotalDebitAmount = decimal.NewFromFloat(100.001)
	t.TotalCreditAmount = decimal.NewFromInt(100)
	s.req.False(t.IsBalanced())
}

// ============================================================================
// CalculateTotals
// ============================================================================

func (s *TransactionSuite) TestCalculateTotals_SetsDebitAndCreditFromEntries() {
	entries := newBalancedEntries(350)
	t := newValidTxn()
	t.Entries = entries
	t.TotalDebitAmount = decimal.Zero
	t.TotalCreditAmount = decimal.Zero

	t.CalculateTotals()

	s.req.True(t.TotalDebitAmount.Equal(decimal.NewFromInt(350)))
	s.req.True(t.TotalCreditAmount.Equal(decimal.NewFromInt(350)))
	s.req.True(t.IsBalanced())
}

func (s *TransactionSuite) TestCalculateTotals_MultipleDebits_SumsAll() {
	tenantID := uuid.New()
	txnID := uuid.New()
	one := func(n int, dr, cr int64, desc string) domain.TransactionEntry {
		return domain.TransactionEntry{
			TenantID: tenantID, TransactionID: txnID,
			EntryNumber:  int32(n),
			AccountID:    uuid.New(),
			DebitAmount:  decimal.NewFromInt(dr),
			CreditAmount: decimal.NewFromInt(cr),
			Description:  desc,
			ExchangeRate: decimal.NewFromInt(1),
		}
	}
	t := newValidTxn()
	t.Entries = []domain.TransactionEntry{
		one(1, 600, 0, "Dr A"),
		one(2, 400, 0, "Dr B"),
		one(3, 0, 1000, "Cr C"),
	}
	t.TotalDebitAmount = decimal.Zero
	t.TotalCreditAmount = decimal.Zero

	t.CalculateTotals()

	s.req.True(t.TotalDebitAmount.Equal(decimal.NewFromInt(1000)))
	s.req.True(t.TotalCreditAmount.Equal(decimal.NewFromInt(1000)))
}

// ============================================================================
// CanBePosted
// ============================================================================

func (s *TransactionSuite) TestCanBePosted_BalancedDraftNoApproval_True() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusDraft
	t.ApprovalRequired = false
	s.req.True(t.CanBePosted())
}

func (s *TransactionSuite) TestCanBePosted_BalancedApprovedStatus_True() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusApproved
	t.ApprovalRequired = false
	s.req.True(t.CanBePosted())
}

func (s *TransactionSuite) TestCanBePosted_Unbalanced_False() {
	t := newValidTxn()
	t.TotalDebitAmount = decimal.NewFromInt(150)
	s.req.False(t.CanBePosted())
}

func (s *TransactionSuite) TestCanBePosted_PostedStatus_False() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusPosted
	s.req.False(t.CanBePosted())
}

func (s *TransactionSuite) TestCanBePosted_RequiresApproval_PendingApproval_False() {
	t := newValidTxn()
	t.ApprovalRequired = true
	t.ApprovalStatus = domain.ApprovalStatusPending
	s.req.False(t.CanBePosted())
}

func (s *TransactionSuite) TestCanBePosted_RequiresApproval_FullyApproved_True() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusApproved
	t.ApprovalRequired = true
	t.ApprovalStatus = domain.ApprovalStatusApproved
	s.req.True(t.CanBePosted())
}

func (s *TransactionSuite) TestCanBePosted_CancelledStatus_False() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusCancelled
	s.req.False(t.CanBePosted())
}

// ============================================================================
// CanBeReversed
// ============================================================================

func (s *TransactionSuite) TestCanBeReversed_PostedAndNotReversed_True() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusPosted
	t.IsReversed = false
	s.req.True(t.CanBeReversed())
}

func (s *TransactionSuite) TestCanBeReversed_AlreadyReversed_False() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusPosted
	t.IsReversed = true
	s.req.False(t.CanBeReversed())
}

func (s *TransactionSuite) TestCanBeReversed_DraftStatus_False() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusDraft
	s.req.False(t.CanBeReversed())
}

func (s *TransactionSuite) TestCanBeReversed_ApprovedStatus_False() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusApproved
	s.req.False(t.CanBeReversed())
}

// ============================================================================
// CanBeEdited
// ============================================================================

func (s *TransactionSuite) TestCanBeEdited_Draft_NotReversed_True() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusDraft
	t.IsReversed = false
	s.req.True(t.CanBeEdited())
}

func (s *TransactionSuite) TestCanBeEdited_Draft_IsReversed_False() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusDraft
	t.IsReversed = true
	s.req.False(t.CanBeEdited())
}

func (s *TransactionSuite) TestCanBeEdited_Posted_False() {
	t := newValidTxn()
	t.TransactionStatus = domain.TransactionStatusPosted
	s.req.False(t.CanBeEdited())
}

// ============================================================================
// Validate — required fields and business rules
// ============================================================================

func (s *TransactionSuite) TestValidate_ValidTransaction_NoErrors() {
	t := newValidTxn()
	s.req.Empty(t.Validate())
}

func (s *TransactionSuite) TestValidate_MissingTenantID_Error() {
	t := newValidTxn()
	t.TenantID = uuid.Nil
	errs := t.Validate()
	s.req.NotEmpty(errs)
	s.req.Equal("tenant_id", errs[0].Field)
	s.req.Equal("REQUIRED_FIELD", errs[0].Code)
}

func (s *TransactionSuite) TestValidate_EmptyTransactionNumber_Error() {
	t := newValidTxn()
	t.TransactionNumber = ""
	s.req.NotEmpty(t.Validate())
}

func (s *TransactionSuite) TestValidate_TransactionNumberTooLong_Error() {
	t := newValidTxn()
	t.TransactionNumber = string(make([]byte, 51)) // 51 chars > max 50
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "transaction_number" && e.Code == "MAX_LENGTH_EXCEEDED" {
			found = true
		}
	}
	s.req.True(found, "expected MAX_LENGTH_EXCEEDED on transaction_number")
}

func (s *TransactionSuite) TestValidate_InvalidCurrencyCode_Error() {
	t := newValidTxn()
	t.CurrencyCode = "US" // must be exactly 3 chars
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "currency_code" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionSuite) TestValidate_ZeroExchangeRate_Error() {
	t := newValidTxn()
	t.ExchangeRate = decimal.Zero
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "exchange_rate" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionSuite) TestValidate_PostingDateBeforeTransactionDate_Error() {
	t := newValidTxn()
	yesterday := t.TransactionDate.Add(-24 * time.Hour)
	t.PostingDate = &yesterday
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Code == "INVALID_DATE_SEQUENCE" {
			found = true
		}
	}
	s.req.True(found, "expected INVALID_DATE_SEQUENCE")
}

func (s *TransactionSuite) TestValidate_UnbalancedEntries_Error() {
	t := newValidTxn()
	t.TotalDebitAmount = decimal.NewFromInt(200)
	// TotalCreditAmount stays at 100 → unbalanced
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Code == "TRANSACTION_UNBALANCED" {
			found = true
		}
	}
	s.req.True(found, "expected TRANSACTION_UNBALANCED error")
}

func (s *TransactionSuite) TestValidate_LessThanTwoEntries_Error() {
	t := newValidTxn()
	t.Entries = t.Entries[:1]
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Code == "INSUFFICIENT_ENTRIES" {
			found = true
		}
	}
	s.req.True(found, "expected INSUFFICIENT_ENTRIES error")
}

func (s *TransactionSuite) TestValidate_ApprovalRequired_StatusNotRequired_Error() {
	t := newValidTxn()
	t.ApprovalRequired = true
	t.ApprovalStatus = domain.ApprovalStatusNotRequired
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Code == "INCONSISTENT_APPROVAL_SETTINGS" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionSuite) TestValidate_ApprovalNotRequired_StatusPending_Error() {
	t := newValidTxn()
	t.ApprovalRequired = false
	t.ApprovalStatus = domain.ApprovalStatusPending
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Code == "INCONSISTENT_APPROVAL_SETTINGS" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionSuite) TestValidate_Recurring_MissingFrequency_Error() {
	t := newValidTxn()
	t.IsRecurring = true
	t.RecurringFrequency = nil
	t.NextRecurringDate = nil
	errs := t.Validate()
	hasFq := false
	hasDate := false
	for _, e := range errs {
		if e.Field == "recurring_frequency" {
			hasFq = true
		}
		if e.Field == "next_recurring_date" {
			hasDate = true
		}
	}
	s.req.True(hasFq, "expected recurring_frequency error")
	s.req.True(hasDate, "expected next_recurring_date error")
}

func (s *TransactionSuite) TestValidate_IsReversed_MissingReversedByID_Error() {
	t := newValidTxn()
	t.IsReversed = true
	t.ReversedByTransactionID = nil
	errs := t.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "reversed_by_transaction_id" {
			found = true
		}
	}
	s.req.True(found)
}

// ============================================================================
// IsSystemGenerated
// ============================================================================

func (s *TransactionSuite) TestIsSystemGenerated_SystemType_True() {
	t := newValidTxn()
	t.TransactionType = domain.TransactionTypeSystem
	s.req.True(t.IsSystemGenerated())
}

func (s *TransactionSuite) TestIsSystemGenerated_ManualType_False() {
	t := newValidTxn()
	t.TransactionType = domain.TransactionTypeManual
	s.req.False(t.IsSystemGenerated())
}

// ============================================================================
// GetNetAmount
// ============================================================================

func (s *TransactionSuite) TestGetNetAmount_BalancedTransaction_ReturnsZero() {
	t := newValidTxn()
	// net = |debit - credit| = |100 - 100| = 0
	s.req.True(t.GetNetAmount().IsZero())
}

// ============================================================================
// RejectionReason
// ============================================================================

func (s *TransactionSuite) TestRejectionReason_AllDefinedReasons_AreValid() {
	reasons := []domain.RejectionReason{
		domain.RejectionReasonInvalidAccount,
		domain.RejectionReasonPeriodClosed,
		domain.RejectionReasonBudgetExceeded,
		domain.RejectionReasonUnbalancedEntry,
		domain.RejectionReasonMissingDocumentation,
		domain.RejectionReasonDuplicateTransaction,
		domain.RejectionReasonAmountMismatch,
		domain.RejectionReasonUnauthorisedAccount,
		domain.RejectionReasonCurrencyMismatch,
		domain.RejectionReasonPolicyViolation,
		domain.RejectionReasonOther,
	}
	for _, r := range reasons {
		s.req.True(r.IsValid(), "reason %s should be valid", r)
	}
	s.req.False(domain.RejectionReason("MADE_UP").IsValid())
}
