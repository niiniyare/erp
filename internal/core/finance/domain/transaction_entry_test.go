package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo.so/internal/core/finance/domain"
)

// ============================================================================
// Suite
// ============================================================================

type TransactionEntrySuite struct {
	suite.Suite
	req      *require.Assertions
	tenantID uuid.UUID
	txnID    uuid.UUID
}

func TestTransactionEntrySuite(t *testing.T) {
	suite.Run(t, new(TransactionEntrySuite))
}

func (s *TransactionEntrySuite) SetupTest() {
	s.req = require.New(s.T())
	s.tenantID = uuid.New()
	s.txnID = uuid.New()
}

// ============================================================================
// Helpers
// ============================================================================

func (s *TransactionEntrySuite) debitEntry(amount int64) domain.TransactionEntry {
	return domain.TransactionEntry{
		TenantID:      s.tenantID,
		TransactionID: s.txnID,
		EntryNumber:   1,
		AccountID:     uuid.New(),
		DebitAmount:   decimal.NewFromInt(amount),
		CreditAmount:  decimal.Zero,
		Description:   "Debit entry",
		ExchangeRate:  decimal.NewFromInt(1),
	}
}

func (s *TransactionEntrySuite) creditEntry(amount int64) domain.TransactionEntry {
	return domain.TransactionEntry{
		TenantID:      s.tenantID,
		TransactionID: s.txnID,
		EntryNumber:   2,
		AccountID:     uuid.New(),
		DebitAmount:   decimal.Zero,
		CreditAmount:  decimal.NewFromInt(amount),
		Description:   "Credit entry",
		ExchangeRate:  decimal.NewFromInt(1),
	}
}

// ============================================================================
// IsDebit / IsCredit
// ============================================================================

func (s *TransactionEntrySuite) TestIsDebit_NonZeroDebitAmount_True() {
	e := s.debitEntry(500)
	s.req.True(e.IsDebit())
	s.req.False(e.IsCredit())
}

func (s *TransactionEntrySuite) TestIsCredit_NonZeroCreditAmount_True() {
	e := s.creditEntry(500)
	s.req.True(e.IsCredit())
	s.req.False(e.IsDebit())
}

// ============================================================================
// GetEffectiveAmount
// ============================================================================

func (s *TransactionEntrySuite) TestGetEffectiveAmount_DebitEntry_ReturnsDebit() {
	e := s.debitEntry(750)
	s.req.True(e.GetEffectiveAmount().Equal(decimal.NewFromInt(750)))
}

func (s *TransactionEntrySuite) TestGetEffectiveAmount_CreditEntry_ReturnsCredit() {
	e := s.creditEntry(750)
	s.req.True(e.GetEffectiveAmount().Equal(decimal.NewFromInt(750)))
}

// ============================================================================
// GetSignedAmount
// ============================================================================

func (s *TransactionEntrySuite) TestGetSignedAmount_Debit_Positive() {
	e := s.debitEntry(300)
	s.req.True(e.GetSignedAmount().Equal(decimal.NewFromInt(300)))
}

func (s *TransactionEntrySuite) TestGetSignedAmount_Credit_Negative() {
	e := s.creditEntry(300)
	s.req.True(e.GetSignedAmount().Equal(decimal.NewFromInt(-300)))
}

// ============================================================================
// Validate — single-side constraint (core double-entry rule)
// ============================================================================

func (s *TransactionEntrySuite) TestValidate_ValidDebitEntry_NoErrors() {
	e := s.debitEntry(1000)
	s.req.Empty(e.Validate())
}

func (s *TransactionEntrySuite) TestValidate_ValidCreditEntry_NoErrors() {
	e := s.creditEntry(1000)
	s.req.Empty(e.Validate())
}

func (s *TransactionEntrySuite) TestValidate_BothAmountsZero_MISSING_AMOUNT_Error() {
	// Both zero violates double-entry: every entry must move value somewhere.
	e := s.debitEntry(0)
	e.CreditAmount = decimal.Zero
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Code == "MISSING_AMOUNT" {
			found = true
		}
	}
	s.req.True(found, "expected MISSING_AMOUNT when both amounts are zero")
}

func (s *TransactionEntrySuite) TestValidate_BothAmountsNonZero_DUAL_AMOUNTS_Error() {
	// Both non-zero violates double-entry: can't be debit AND credit simultaneously.
	e := s.debitEntry(500)
	e.CreditAmount = decimal.NewFromInt(500)
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Code == "DUAL_AMOUNTS" {
			found = true
		}
	}
	s.req.True(found, "expected DUAL_AMOUNTS when both amounts are non-zero")
}

func (s *TransactionEntrySuite) TestValidate_NegativeDebitAmount_Error() {
	e := s.debitEntry(0)
	e.DebitAmount = decimal.NewFromInt(-1)
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "debit_amount" {
			found = true
		}
	}
	s.req.True(found, "expected error on negative debit_amount")
}

func (s *TransactionEntrySuite) TestValidate_NegativeCreditAmount_Error() {
	e := s.creditEntry(0)
	e.CreditAmount = decimal.NewFromInt(-1)
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "credit_amount" {
			found = true
		}
	}
	s.req.True(found, "expected error on negative credit_amount")
}

func (s *TransactionEntrySuite) TestValidate_ZeroExchangeRate_Error() {
	e := s.debitEntry(100)
	e.ExchangeRate = decimal.Zero
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "exchange_rate" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionEntrySuite) TestValidate_MissingTenantID_Error() {
	e := s.debitEntry(100)
	e.TenantID = uuid.Nil
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "tenant_id" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionEntrySuite) TestValidate_MissingAccountID_Error() {
	e := s.debitEntry(100)
	e.AccountID = uuid.Nil
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "account_id" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionEntrySuite) TestValidate_EntryNumberZero_Error() {
	e := s.debitEntry(100)
	e.EntryNumber = 0
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "entry_number" {
			found = true
		}
	}
	s.req.True(found)
}

// ============================================================================
// Validate — multi-currency
// ============================================================================

func (s *TransactionEntrySuite) TestValidate_ValidOriginalCurrency_NoErrors() {
	e := s.debitEntry(100)
	curr := "EUR"
	e.OriginalCurrency = &curr
	e.OriginalAmount = decimal.NewFromInt(90)
	s.req.Empty(e.Validate())
}

func (s *TransactionEntrySuite) TestValidate_OriginalCurrencyWrongLength_Error() {
	e := s.debitEntry(100)
	curr := "EU" // must be 3 chars
	e.OriginalCurrency = &curr
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "original_currency" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionEntrySuite) TestValidate_OriginalAmountWithoutCurrency_Error() {
	e := s.debitEntry(100)
	e.OriginalCurrency = nil
	e.OriginalAmount = decimal.NewFromInt(90) // non-zero but no currency
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "original_amount" && err.Code == "INCONSISTENT_CURRENCY_DATA" {
			found = true
		}
	}
	s.req.True(found)
}

// ============================================================================
// Validate — tax
// ============================================================================

func (s *TransactionEntrySuite) TestValidate_TaxRateOutOfRange_Error() {
	e := s.debitEntry(100)
	e.TaxRate = decimal.NewFromInt(101) // > 100%
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "tax_rate" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *TransactionEntrySuite) TestValidate_NegativeTaxAmount_Error() {
	e := s.debitEntry(100)
	e.TaxAmount = decimal.NewFromInt(-5)
	errs := e.Validate()
	found := false
	for _, err := range errs {
		if err.Field == "tax_amount" {
			found = true
		}
	}
	s.req.True(found)
}

// ============================================================================
// CreateCounterEntry
// ============================================================================

func (s *TransactionEntrySuite) TestCreateCounterEntry_DebitEntry_CreatesMatchingCredit() {
	debit := s.debitEntry(1500)
	counterAccountID := uuid.New()
	counter := debit.CreateCounterEntry(counterAccountID, "Counter entry")

	s.req.Equal(counterAccountID, counter.AccountID)
	s.req.True(counter.CreditAmount.Equal(debit.DebitAmount), "counter credit should equal original debit")
	s.req.True(counter.DebitAmount.IsZero(), "counter debit should be zero")
	s.req.Equal(debit.EntryNumber+1, counter.EntryNumber)
}

func (s *TransactionEntrySuite) TestCreateCounterEntry_CreditEntry_CreatesMatchingDebit() {
	credit := s.creditEntry(1500)
	counterAccountID := uuid.New()
	counter := credit.CreateCounterEntry(counterAccountID, "Counter entry")

	s.req.True(counter.DebitAmount.Equal(credit.CreditAmount), "counter debit should equal original credit")
	s.req.True(counter.CreditAmount.IsZero(), "counter credit should be zero")
}

// ============================================================================
// ValidateAmountConsistency (multi-currency)
// ============================================================================

func (s *TransactionEntrySuite) TestValidateAmountConsistency_CorrectConversion_NoErrors() {
	e := s.debitEntry(108) // USD, after conversion from EUR at rate 1.08
	curr := "EUR"
	e.OriginalCurrency = &curr
	e.OriginalAmount = decimal.NewFromInt(100)
	e.ExchangeRate = decimal.NewFromFloat(1.08)
	s.req.Empty(e.ValidateAmountConsistency())
}

func (s *TransactionEntrySuite) TestValidateAmountConsistency_LargeDiscrepancy_Error() {
	e := s.debitEntry(200) // claims 200 but rate * original = 108
	curr := "EUR"
	e.OriginalCurrency = &curr
	e.OriginalAmount = decimal.NewFromInt(100)
	e.ExchangeRate = decimal.NewFromFloat(1.08)
	errs := e.ValidateAmountConsistency()
	found := false
	for _, err := range errs {
		if err.Code == "CURRENCY_CONVERSION_MISMATCH" {
			found = true
		}
	}
	s.req.True(found, "expected CURRENCY_CONVERSION_MISMATCH for large discrepancy")
}

// ============================================================================
// Clone
// ============================================================================

func (s *TransactionEntrySuite) TestClone_ProducesDeepCopy() {
	e := s.debitEntry(100)
	ref := "REF-001"
	e.Reference = &ref

	cloned := e.Clone()

	s.req.NotEqual(e.ID, cloned.ID, "clone should have a new UUID")
	*cloned.Reference = "CHANGED"
	s.req.Equal("REF-001", *e.Reference, "original reference must not change when clone is mutated")
}

// ============================================================================
// Reconciliation
// ============================================================================

func (s *TransactionEntrySuite) TestIsReconciled_BothFlagsSet_True() {
	e := s.debitEntry(100)
	e.MarkReconciled("RECON-001")
	s.req.True(e.IsReconciled())
}

func (s *TransactionEntrySuite) TestIsReconciled_AfterUnmark_False() {
	e := s.debitEntry(100)
	e.MarkReconciled("RECON-001")
	e.UnmarkReconciled()
	s.req.False(e.IsReconciled())
}
