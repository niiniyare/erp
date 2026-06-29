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

type BudgetSuite struct {
	suite.Suite
	req *require.Assertions
}

func TestBudgetSuite(t *testing.T) {
	suite.Run(t, new(BudgetSuite))
}

func (s *BudgetSuite) SetupTest() {
	s.req = require.New(s.T())
}

// ============================================================================
// Helpers
// ============================================================================

func newValidBudget() domain.Budget {
	return domain.Budget{
		TenantID:     uuid.New(),
		FiscalYearID: uuid.New(),
		Name:         "FY2026 Operating Budget",
		BudgetType:   domain.BudgetTypeAnnual,
		Status:       domain.BudgetStatusDraft,
		CurrencyCode: "USD",
		Version:      1,
	}
}

func newValidLineItem(budgetID, tenantID uuid.UUID) domain.BudgetLineItem {
	return domain.BudgetLineItem{
		BudgetID:       budgetID,
		TenantID:       tenantID,
		AccountID:      uuid.New(),
		BudgetedAmount: decimal.NewFromInt(50_000),
	}
}

// ============================================================================
// BudgetStatus
// ============================================================================

func (s *BudgetSuite) TestBudgetStatus_AllDefined_AreValid() {
	valid := []domain.BudgetStatus{
		domain.BudgetStatusDraft,
		domain.BudgetStatusSubmitted,
		domain.BudgetStatusApproved,
		domain.BudgetStatusRejected,
		domain.BudgetStatusRevised,
		domain.BudgetStatusClosed,
	}
	for _, st := range valid {
		s.req.True(st.IsValid(), "expected %s to be valid", st)
	}
	s.req.False(domain.BudgetStatus("BOGUS").IsValid())
}

func (s *BudgetSuite) TestBudgetStatus_IsEditable_OnlyDraftAndRejected() {
	cases := []struct {
		status   domain.BudgetStatus
		editable bool
	}{
		{domain.BudgetStatusDraft, true},
		{domain.BudgetStatusRejected, true},
		{domain.BudgetStatusSubmitted, false},
		{domain.BudgetStatusApproved, false},
		{domain.BudgetStatusRevised, false},
		{domain.BudgetStatusClosed, false},
	}
	for _, tc := range cases {
		s.req.Equal(tc.editable, tc.status.IsEditable(), "status=%s", tc.status)
	}
}

// ============================================================================
// BudgetType
// ============================================================================

func (s *BudgetSuite) TestBudgetType_AllDefined_AreValid() {
	valid := []domain.BudgetType{
		domain.BudgetTypeAnnual,
		domain.BudgetTypeQuarterly,
		domain.BudgetTypeProject,
		domain.BudgetTypeDepartment,
		domain.BudgetTypeCapex,
	}
	for _, bt := range valid {
		s.req.True(bt.IsValid(), "expected %s to be valid", bt)
	}
	s.req.False(domain.BudgetType("ROLLING").IsValid())
}

// ============================================================================
// Budget.Validate
// ============================================================================

func (s *BudgetSuite) TestBudgetValidate_ValidBudget_NoErrors() {
	b := newValidBudget()
	s.req.Empty(b.Validate())
}

func (s *BudgetSuite) TestBudgetValidate_MissingTenantID_Error() {
	b := newValidBudget()
	b.TenantID = uuid.Nil
	errs := b.Validate()
	s.req.NotEmpty(errs)
	found := false
	for _, e := range errs {
		if e.Field == "tenant_id" && e.Code == "REQUIRED_FIELD" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_MissingFiscalYearID_Error() {
	b := newValidBudget()
	b.FiscalYearID = uuid.Nil
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "fiscal_year_id" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_EmptyName_Error() {
	b := newValidBudget()
	b.Name = ""
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_WhitespaceOnlyName_Error() {
	b := newValidBudget()
	b.Name = "   "
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "name" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_InvalidBudgetType_Error() {
	b := newValidBudget()
	b.BudgetType = domain.BudgetType("UNKNOWN")
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "budget_type" && e.Code == "INVALID_VALUE" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_InvalidStatus_Error() {
	b := newValidBudget()
	b.Status = domain.BudgetStatus("LIMBO")
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "status" && e.Code == "INVALID_VALUE" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_CurrencyCodeNot3Chars_Error() {
	b := newValidBudget()
	b.CurrencyCode = "US" // must be exactly 3
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "currency_code" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_CurrencyCode4Chars_Error() {
	b := newValidBudget()
	b.CurrencyCode = "USDD"
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "currency_code" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_VersionZero_Error() {
	b := newValidBudget()
	b.Version = 0
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "version" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_VersionNegative_Error() {
	b := newValidBudget()
	b.Version = -1
	errs := b.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "version" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestBudgetValidate_MultipleErrors_ReturnsAll() {
	b := domain.Budget{} // zero value — everything is invalid
	errs := b.Validate()
	s.req.GreaterOrEqual(len(errs), 4, "expected at least 4 validation errors for zero-value budget")
}

// ============================================================================
// BudgetLineItem.Validate
// ============================================================================

func (s *BudgetSuite) TestLineItemValidate_Valid_NoErrors() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	s.req.Empty(li.Validate())
}

func (s *BudgetSuite) TestLineItemValidate_MissingBudgetID_Error() {
	b := newValidBudget()
	li := newValidLineItem(uuid.Nil, b.TenantID)
	errs := li.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "budget_id" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestLineItemValidate_MissingTenantID_Error() {
	li := newValidLineItem(uuid.New(), uuid.Nil)
	errs := li.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "tenant_id" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestLineItemValidate_MissingAccountID_Error() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.AccountID = uuid.Nil
	errs := li.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "account_id" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestLineItemValidate_NegativeBudgetedAmount_Error() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.NewFromInt(-1)
	errs := li.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "budgeted_amount" {
			found = true
		}
	}
	s.req.True(found)
}

func (s *BudgetSuite) TestLineItemValidate_ZeroBudgetedAmount_Valid() {
	// Zero budgets are allowed (line placeholder for future allocation)
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.Zero
	s.req.Empty(li.Validate())
}

// ============================================================================
// BudgetLineItem.IsOverBudget
// ============================================================================

func (s *BudgetSuite) TestIsOverBudget_ActualExceedsBudgeted_True() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.ActualAmount = li.BudgetedAmount.Add(decimal.NewFromInt(1))
	s.req.True(li.IsOverBudget())
}

func (s *BudgetSuite) TestIsOverBudget_ActualEqualsBudgeted_False() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.ActualAmount = li.BudgetedAmount
	s.req.False(li.IsOverBudget())
}

func (s *BudgetSuite) TestIsOverBudget_ActualBelowBudgeted_False() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.ActualAmount = li.BudgetedAmount.Sub(decimal.NewFromInt(1))
	s.req.False(li.IsOverBudget())
}

// ============================================================================
// BudgetLineItem.VariancePercent
// ============================================================================

func (s *BudgetSuite) TestVariancePercent_HalfSpent_Returns50Percent() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.NewFromInt(100)
	li.ActualAmount = decimal.NewFromInt(50)
	// variance = (100 - 50) / 100 * 100 = 50%
	pct := li.VariancePercent()
	s.req.True(pct.Equal(decimal.NewFromInt(50)), "expected 50%%, got %s", pct)
}

func (s *BudgetSuite) TestVariancePercent_OverBudget_ReturnsNegativePercent() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.NewFromInt(100)
	li.ActualAmount = decimal.NewFromInt(120)
	// variance = (100 - 120) / 100 * 100 = -20%
	pct := li.VariancePercent()
	s.req.True(pct.Equal(decimal.NewFromInt(-20)), "expected -20%%, got %s", pct)
}

func (s *BudgetSuite) TestVariancePercent_ZeroBudget_ReturnsZero() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.Zero
	li.ActualAmount = decimal.NewFromInt(500)
	// Avoid division-by-zero — must return zero
	pct := li.VariancePercent()
	s.req.True(pct.IsZero(), "expected 0 for zero budget, got %s", pct)
}

func (s *BudgetSuite) TestVariancePercent_UnderBudget_ReturnsPositivePercent() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.NewFromInt(200)
	li.ActualAmount = decimal.NewFromInt(50)
	// (200-50)/200 * 100 = 75%
	pct := li.VariancePercent()
	s.req.True(pct.Equal(decimal.NewFromInt(75)), "expected 75%%, got %s", pct)
}

// ============================================================================
// BudgetLineItem.ExceedsVarianceThreshold
// ============================================================================

func (s *BudgetSuite) TestExceedsVarianceThreshold_VarianceAboveThreshold_True() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.NewFromInt(100)
	li.ActualAmount = decimal.NewFromInt(125) // -25% variance
	threshold := decimal.NewFromInt(10)       // 10% threshold
	s.req.True(li.ExceedsVarianceThreshold(threshold))
}

func (s *BudgetSuite) TestExceedsVarianceThreshold_VarianceBelowThreshold_False() {
	b := newValidBudget()
	li := newValidLineItem(uuid.New(), b.TenantID)
	li.BudgetedAmount = decimal.NewFromInt(100)
	li.ActualAmount = decimal.NewFromInt(105) // -5% variance
	threshold := decimal.NewFromInt(10)       // 10% threshold
	s.req.False(li.ExceedsVarianceThreshold(threshold))
}
