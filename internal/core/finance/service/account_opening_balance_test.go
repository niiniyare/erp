// Package service_test — OpeningBalanceEngine governance tests.
//
// Tests (FIN-OB-*):
//
//	001 — SetOpeningBalance creates PENDING request for valid account
//	002 — SetOpeningBalance rejects non-accepting account
//	003 — SetOpeningBalance rejects currency mismatch
//	004 — SetOpeningBalance rejects direction mismatch (credit vs DEBIT-normal)
//	005 — SetOpeningBalance rejects duplicate in-flight for same period
//	006 — ApproveOpeningBalance transitions PENDING → APPROVED
//	007 — ApproveOpeningBalance enforces SOD (approver ≠ submitter)
//	008 — ApproveOpeningBalance rejects non-PENDING balance
//	009 — RejectOpeningBalance transitions PENDING → REJECTED
//	010 — RejectOpeningBalance requires reason
//	011 — VoidOpeningBalance voids PENDING
//	012 — VoidOpeningBalance voids APPROVED
//	013 — VoidOpeningBalance rejects already-POSTED
//	014 — PostOpeningBalance transitions APPROVED → POSTED
//	015 — PostOpeningBalance rejects non-APPROVED balance
//	016 — PostOpeningBalance rejects re-posting of POSTED balance
//	017 — GetOpeningBalances returns all balances for account
//	018 — Nil engine — all methods nil-safe
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// =============================================================================
// Stubs
// =============================================================================

// obAccountRepo serves one account for GetByID.
type obAccountRepo struct {
	p16StubAccountRepo
	account *domain.Accounts
}

func (r *obAccountRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Accounts, error) {
	if r.account != nil && r.account.ID == id {
		return r.account, nil
	}
	return nil, errors.New("account not found")
}

// obBalanceRepo is an in-memory OpeningBalanceRepository.
type obBalanceRepo struct {
	balances map[uuid.UUID]*domain.OpeningBalance
}

func newOBBalanceRepo() *obBalanceRepo {
	return &obBalanceRepo{balances: make(map[uuid.UUID]*domain.OpeningBalance)}
}

func (r *obBalanceRepo) Create(_ context.Context, ob *domain.OpeningBalance) error {
	r.balances[ob.ID] = ob
	return nil
}

func (r *obBalanceRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.OpeningBalance, error) {
	if ob, ok := r.balances[id]; ok {
		return ob, nil
	}
	return nil, errors.New("opening balance not found")
}

func (r *obBalanceRepo) GetByAccount(_ context.Context, accountID uuid.UUID) ([]*domain.OpeningBalance, error) {
	var out []*domain.OpeningBalance
	for _, ob := range r.balances {
		if ob.AccountID == accountID {
			out = append(out, ob)
		}
	}
	return out, nil
}

func (r *obBalanceRepo) Update(_ context.Context, ob *domain.OpeningBalance) error {
	r.balances[ob.ID] = ob
	return nil
}

// =============================================================================
// Helpers
// =============================================================================

func obCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

// newActiveLeafAccount returns a VALID, ACTIVE, leaf account ready to accept postings.
func newActiveLeafAccount(tenantID uuid.UUID) *domain.Accounts {
	code := "10010001"
	curr := "USD"
	return &domain.Accounts{
		ID:               uuid.New(),
		TenantID:         tenantID,
		AccountCode:      code,
		AccountName:      "Test Asset",
		RootType:         domain.RootTypeAsset,
		NormalBalance:    domain.NormalBalanceDebit,
		AccountType:      "Current Asset",
		Status:           domain.AccountStatusActive,
		ValidationStatus: domain.ValidationStatusValid,
		IsActive:         true,
		HasChildren:      false,
		CurrencyCode:     &curr,
		CurrentBalance:   decimal.Zero,
		CreatedBy:        uuid.New(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func newOBEngine(account *domain.Accounts) (*service.OpeningBalanceEngine, *obBalanceRepo) {
	accountRepo := &obAccountRepo{account: account}
	balanceRepo := newOBBalanceRepo()
	eng := service.NewOpeningBalanceEngine(accountRepo, balanceRepo, nil, metrics.NewNoOpMetricsProvider())
	return eng, balanceRepo
}

func validOBReq(account *domain.Accounts) domain.OpeningBalanceRequest {
	return domain.OpeningBalanceRequest{
		AccountID:     account.ID,
		PeriodID:      uuid.New(),
		EffectiveDate: time.Now(),
		DebitAmount:   decimal.NewFromFloat(1000.00), // DEBIT for ASSET account
		CreditAmount:  decimal.Zero,
		CurrencyCode:  "USD",
		Reason:        "year-start opening balance",
		SubmittedBy:   uuid.New(),
	}
}

// =============================================================================
// FIN-OB-001: SetOpeningBalance creates PENDING request
// =============================================================================

func TestOB_SetOpeningBalance_CreatesPending(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, err := eng.SetOpeningBalance(obCtx(), req)
	if err != nil {
		t.Fatalf("FIN-OB-001: unexpected error: %v", err)
	}
	if ob.Status != domain.OpeningBalanceStatusPending {
		t.Errorf("FIN-OB-001: expected PENDING, got %s", ob.Status)
	}
	if _, ok := balRepo.balances[ob.ID]; !ok {
		t.Error("FIN-OB-001: opening balance not persisted")
	}
}

// =============================================================================
// FIN-OB-002: SetOpeningBalance rejects non-accepting account
// =============================================================================

func TestOB_SetOpeningBalance_NonAcceptingAccount_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	account.Status = domain.AccountStatusFrozen // cannot accept transactions
	account.IsActive = false

	eng, _ := newOBEngine(account)
	req := validOBReq(account)

	_, err := eng.SetOpeningBalance(obCtx(), req)
	if err == nil {
		t.Error("FIN-OB-002: expected error for frozen account, got nil")
	}
}

// =============================================================================
// FIN-OB-003: SetOpeningBalance rejects currency mismatch
// =============================================================================

func TestOB_SetOpeningBalance_CurrencyMismatch_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, _ := newOBEngine(account)

	req := validOBReq(account)
	req.CurrencyCode = "EUR" // account uses USD

	_, err := eng.SetOpeningBalance(obCtx(), req)
	if err == nil {
		t.Error("FIN-OB-003: expected currency mismatch error, got nil")
	}
	if !errors.Is(err, domain.ErrOpeningBalanceCurrencyMismatch) {
		t.Errorf("FIN-OB-003: expected ErrOpeningBalanceCurrencyMismatch, got %v", err)
	}
}

// =============================================================================
// FIN-OB-004: SetOpeningBalance rejects direction mismatch
// =============================================================================

func TestOB_SetOpeningBalance_DirectionMismatch_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID) // ASSET = DEBIT normal
	eng, _ := newOBEngine(account)

	req := validOBReq(account)
	req.DebitAmount = decimal.Zero
	req.CreditAmount = decimal.NewFromFloat(1000.00) // credit on DEBIT-normal = mismatch

	_, err := eng.SetOpeningBalance(obCtx(), req)
	if err == nil {
		t.Error("FIN-OB-004: expected direction mismatch error, got nil")
	}
	if !errors.Is(err, domain.ErrOpeningBalanceDirectionMismatch) {
		t.Errorf("FIN-OB-004: expected ErrOpeningBalanceDirectionMismatch, got %v", err)
	}
}

// =============================================================================
// FIN-OB-005: SetOpeningBalance rejects duplicate in-flight for same period
// =============================================================================

func TestOB_SetOpeningBalance_DuplicateInFlight_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, _ := newOBEngine(account)

	req := validOBReq(account)
	periodID := req.PeriodID

	// First submission succeeds.
	_, err := eng.SetOpeningBalance(obCtx(), req)
	if err != nil {
		t.Fatalf("FIN-OB-005: first submission failed: %v", err)
	}

	// Second submission for same period must fail.
	req2 := validOBReq(account)
	req2.PeriodID = periodID
	_, err = eng.SetOpeningBalance(obCtx(), req2)
	if err == nil {
		t.Error("FIN-OB-005: expected error for duplicate in-flight balance, got nil")
	}
}

// =============================================================================
// FIN-OB-006: ApproveOpeningBalance PENDING → APPROVED
// =============================================================================

func TestOB_ApproveOpeningBalance_PendingToApproved(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)

	approver := uuid.New() // different from req.SubmittedBy
	err := eng.ApproveOpeningBalance(obCtx(), ob.ID, approver, "approved")
	if err != nil {
		t.Fatalf("FIN-OB-006: unexpected error: %v", err)
	}

	got := balRepo.balances[ob.ID]
	if got.Status != domain.OpeningBalanceStatusApproved {
		t.Errorf("FIN-OB-006: expected APPROVED, got %s", got.Status)
	}
}

// =============================================================================
// FIN-OB-007: ApproveOpeningBalance enforces SOD
// =============================================================================

func TestOB_ApproveOpeningBalance_SOD_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, _ := newOBEngine(account)

	submitter := uuid.New()
	req := validOBReq(account)
	req.SubmittedBy = submitter
	ob, _ := eng.SetOpeningBalance(obCtx(), req)

	// Same person tries to approve.
	err := eng.ApproveOpeningBalance(obCtx(), ob.ID, submitter, "self-approve")
	if err == nil {
		t.Error("FIN-OB-007: expected SOD error, got nil")
	}
	if !errors.Is(err, domain.ErrOpeningBalanceSelfApproval) {
		t.Errorf("FIN-OB-007: expected ErrOpeningBalanceSelfApproval, got %v", err)
	}
}

// =============================================================================
// FIN-OB-008: ApproveOpeningBalance rejects non-PENDING
// =============================================================================

func TestOB_ApproveOpeningBalance_NonPending_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)

	// Manually mark as rejected.
	ob.Status = domain.OpeningBalanceStatusRejected
	balRepo.balances[ob.ID] = ob

	err := eng.ApproveOpeningBalance(obCtx(), ob.ID, uuid.New(), "approve rejected")
	if err == nil {
		t.Error("FIN-OB-008: expected error for non-PENDING balance, got nil")
	}
}

// =============================================================================
// FIN-OB-009: RejectOpeningBalance PENDING → REJECTED
// =============================================================================

func TestOB_RejectOpeningBalance_PendingToRejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)

	err := eng.RejectOpeningBalance(obCtx(), ob.ID, uuid.New(), "incorrect amount")
	if err != nil {
		t.Fatalf("FIN-OB-009: unexpected error: %v", err)
	}

	got := balRepo.balances[ob.ID]
	if got.Status != domain.OpeningBalanceStatusRejected {
		t.Errorf("FIN-OB-009: expected REJECTED, got %s", got.Status)
	}
}

// =============================================================================
// FIN-OB-010: RejectOpeningBalance requires reason
// =============================================================================

func TestOB_RejectOpeningBalance_RequiresReason(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, _ := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)

	err := eng.RejectOpeningBalance(obCtx(), ob.ID, uuid.New(), "")
	if err == nil {
		t.Error("FIN-OB-010: expected error for empty rejection reason, got nil")
	}
}

// =============================================================================
// FIN-OB-011: VoidOpeningBalance voids PENDING
// =============================================================================

func TestOB_VoidOpeningBalance_VoidsPending(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)

	err := eng.VoidOpeningBalance(obCtx(), ob.ID, uuid.New(), "period cancelled")
	if err != nil {
		t.Fatalf("FIN-OB-011: unexpected error: %v", err)
	}

	got := balRepo.balances[ob.ID]
	if got.Status != domain.OpeningBalanceStatusVoided {
		t.Errorf("FIN-OB-011: expected VOIDED, got %s", got.Status)
	}
}

// =============================================================================
// FIN-OB-012: VoidOpeningBalance voids APPROVED
// =============================================================================

func TestOB_VoidOpeningBalance_VoidsApproved(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)
	_ = eng.ApproveOpeningBalance(obCtx(), ob.ID, uuid.New(), "approved")

	err := eng.VoidOpeningBalance(obCtx(), ob.ID, uuid.New(), "reverted before posting")
	if err != nil {
		t.Fatalf("FIN-OB-012: unexpected error: %v", err)
	}

	got := balRepo.balances[ob.ID]
	if got.Status != domain.OpeningBalanceStatusVoided {
		t.Errorf("FIN-OB-012: expected VOIDED, got %s", got.Status)
	}
}

// =============================================================================
// FIN-OB-013: VoidOpeningBalance rejects already-POSTED
// =============================================================================

func TestOB_VoidOpeningBalance_PostedCannotBeVoided(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)
	_ = eng.ApproveOpeningBalance(obCtx(), ob.ID, uuid.New(), "approved")
	_, _ = eng.PostOpeningBalance(obCtx(), ob.ID, uuid.New())

	// Verify it's posted.
	got := balRepo.balances[ob.ID]
	if got.Status != domain.OpeningBalanceStatusPosted {
		t.Skipf("FIN-OB-013: skipping — posting not completed (status=%s)", got.Status)
	}

	err := eng.VoidOpeningBalance(obCtx(), ob.ID, uuid.New(), "too late")
	if err == nil {
		t.Error("FIN-OB-013: expected error for voiding POSTED balance, got nil")
	}
}

// =============================================================================
// FIN-OB-014: PostOpeningBalance APPROVED → POSTED
// =============================================================================

func TestOB_PostOpeningBalance_ApprovedToPosted(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)
	_ = eng.ApproveOpeningBalance(obCtx(), ob.ID, uuid.New(), "approved")

	posted, err := eng.PostOpeningBalance(obCtx(), ob.ID, uuid.New())
	if err != nil {
		t.Fatalf("FIN-OB-014: unexpected error: %v", err)
	}
	if posted.Status != domain.OpeningBalanceStatusPosted {
		t.Errorf("FIN-OB-014: expected POSTED, got %s", posted.Status)
	}
	if posted.TransactionID == nil {
		t.Error("FIN-OB-014: expected TransactionID to be set after posting")
	}
	_ = balRepo
}

// =============================================================================
// FIN-OB-015: PostOpeningBalance rejects non-APPROVED
// =============================================================================

func TestOB_PostOpeningBalance_NonApproved_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, _ := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)
	// ob is still PENDING — not approved

	_, err := eng.PostOpeningBalance(obCtx(), ob.ID, uuid.New())
	if err == nil {
		t.Error("FIN-OB-015: expected error for posting non-APPROVED balance, got nil")
	}
	if !errors.Is(err, domain.ErrOpeningBalanceNotApproved) {
		t.Errorf("FIN-OB-015: expected ErrOpeningBalanceNotApproved, got %v", err)
	}
}

// =============================================================================
// FIN-OB-016: PostOpeningBalance rejects re-posting
// =============================================================================

func TestOB_PostOpeningBalance_AlreadyPosted_Rejected(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, balRepo := newOBEngine(account)

	req := validOBReq(account)
	ob, _ := eng.SetOpeningBalance(obCtx(), req)
	_ = eng.ApproveOpeningBalance(obCtx(), ob.ID, uuid.New(), "approved")
	_, _ = eng.PostOpeningBalance(obCtx(), ob.ID, uuid.New())

	// Verify posted before testing re-post.
	if balRepo.balances[ob.ID].Status != domain.OpeningBalanceStatusPosted {
		t.Skip("FIN-OB-016: skipping — posting did not complete")
	}

	_, err := eng.PostOpeningBalance(obCtx(), ob.ID, uuid.New())
	if err == nil {
		t.Error("FIN-OB-016: expected error for re-posting, got nil")
	}
	if !errors.Is(err, domain.ErrOpeningBalanceAlreadyPosted) {
		t.Errorf("FIN-OB-016: expected ErrOpeningBalanceAlreadyPosted, got %v", err)
	}
}

// =============================================================================
// FIN-OB-017: GetOpeningBalances returns all for account
// =============================================================================

func TestOB_GetOpeningBalances_ReturnsAll(t *testing.T) {
	tenantID := uuid.New()
	account := newActiveLeafAccount(tenantID)
	eng, _ := newOBEngine(account)

	req1 := validOBReq(account)
	req1.Reason = "first"
	_, _ = eng.SetOpeningBalance(obCtx(), req1)

	req2 := validOBReq(account) // different period
	req2.Reason = "second"
	_, _ = eng.SetOpeningBalance(obCtx(), req2)

	balances, err := eng.GetOpeningBalances(obCtx(), account.ID)
	if err != nil {
		t.Fatalf("FIN-OB-017: unexpected error: %v", err)
	}
	if len(balances) < 2 {
		t.Errorf("FIN-OB-017: expected >=2 balances, got %d", len(balances))
	}
}

// =============================================================================
// FIN-OB-018: Nil engine — nil-safe
// =============================================================================

func TestOB_NilEngine_Safe(t *testing.T) {
	var eng *service.OpeningBalanceEngine
	ctx := obCtx()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-OB-018: nil engine panicked: %v", r)
		}
	}()

	_, _ = eng.SetOpeningBalance(ctx, domain.OpeningBalanceRequest{})
	_ = eng.ApproveOpeningBalance(ctx, uuid.New(), uuid.New(), "")
	_ = eng.RejectOpeningBalance(ctx, uuid.New(), uuid.New(), "")
	_ = eng.VoidOpeningBalance(ctx, uuid.New(), uuid.New(), "")
	_, _ = eng.PostOpeningBalance(ctx, uuid.New(), uuid.New())
	bs, _ := eng.GetOpeningBalances(ctx, uuid.New())
	if bs != nil {
		t.Error("FIN-OB-018: nil engine GetOpeningBalances should return nil")
	}
}
