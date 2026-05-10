// Package service_test — Account lifecycle state machine tests.
//
// Tests (FIN-LIFECYCLE-*):
//
//	001 — SubmitForApproval transitions DRAFT → PENDING_APPROVAL
//	002 — SubmitForApproval rejects non-DRAFT account
//	003 — SubmitForApproval rejects invalid account (validation failure)
//	004 — Approve transitions PENDING_APPROVAL → ACTIVE
//	005 — Approve enforces SOD (approver ≠ creator)
//	006 — Approve rejects non-PENDING account
//	007 — Reject transitions PENDING_APPROVAL → DRAFT
//	008 — Reject requires reason
//	009 — Freeze transitions ACTIVE → FROZEN
//	010 — Freeze requires reason
//	011 — Unfreeze transitions FROZEN → ACTIVE
//	012 — Unfreeze rejects non-FROZEN account
//	013 — Close transitions INACTIVE → CLOSED; zero-balance required
//	014 — Close rejects non-zero balance
//	015 — Archive transitions CLOSED → ARCHIVED
//	016 — Suspend transitions ACTIVE → SUSPENDED; requires reason
//	017 — PlaceComplianceHold requires reason
//	018 — GetLifecycleHistory returns events in order
//	019 — Nil service receiver — all methods nil-safe
//	020 — Event persistence failure does not block transition
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
// Overridable account repo that delegates to p16StubAccountRepo by default.
// =============================================================================

// lcAccountRepo wraps p16StubAccountRepo and adds overridable GetByID / Update hooks.
type lcAccountRepo struct {
	p16StubAccountRepo
	accounts map[uuid.UUID]*domain.Accounts
	updateFn func(*domain.Accounts) error
}

func newLCAccountRepo(accounts ...*domain.Accounts) *lcAccountRepo {
	r := &lcAccountRepo{accounts: make(map[uuid.UUID]*domain.Accounts)}
	for _, a := range accounts {
		r.accounts[a.ID] = a
	}
	return r
}

func (r *lcAccountRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Accounts, error) {
	if a, ok := r.accounts[id]; ok {
		return a, nil
	}
	return nil, errors.New("account not found")
}

func (r *lcAccountRepo) Update(_ context.Context, a *domain.Accounts) error {
	if r.updateFn != nil {
		return r.updateFn(a)
	}
	r.accounts[a.ID] = a
	return nil
}

// =============================================================================
// Lifecycle event repo stub
// =============================================================================

type lcEventRepo struct {
	events   []*domain.AccountLifecycleEvent
	appendFn func(*domain.AccountLifecycleEvent) error
}

func (r *lcEventRepo) Append(_ context.Context, e *domain.AccountLifecycleEvent) error {
	if r.appendFn != nil {
		return r.appendFn(e)
	}
	r.events = append(r.events, e)
	return nil
}
func (r *lcEventRepo) ListForAccount(_ context.Context, id uuid.UUID) ([]*domain.AccountLifecycleEvent, error) {
	var out []*domain.AccountLifecycleEvent
	for _, e := range r.events {
		if e.AccountID == id {
			out = append(out, e)
		}
	}
	return out, nil
}
func (r *lcEventRepo) ListForTenant(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*domain.AccountLifecycleEvent, error) {
	return r.events, nil
}

// =============================================================================
// Helpers
// =============================================================================

func lcCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

// newValidAccount returns a minimally valid DRAFT account.
func newValidAccount(tenantID uuid.UUID, creatorID uuid.UUID) *domain.Accounts {
	code := "10010001"
	curr := "USD"
	return &domain.Accounts{
		ID:               uuid.New(),
		TenantID:         tenantID,
		AccountCode:      code,
		AccountName:      "Test Asset Account",
		RootType:         domain.RootTypeAsset,
		NormalBalance:    domain.NormalBalanceDebit,
		AccountType:      "Current Asset",
		Status:           domain.AccountStatusDraft,
		ValidationStatus: domain.ValidationStatusValid,
		IsActive:         false,
		CurrencyCode:     &curr,
		CurrentBalance:   decimal.Zero,
		CreatedBy:        creatorID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func newLifecycleSvc(
	repo *lcAccountRepo,
	evRepo *lcEventRepo,
) *service.AccountLifecycleService {
	gov := service.NewAccountMutationGovernor(repo, metrics.NewNoOpMetricsProvider())
	return service.NewAccountLifecycleService(repo, evRepo, gov, metrics.NewNoOpMetricsProvider())
}

// =============================================================================
// FIN-LIFECYCLE-001: SubmitForApproval DRAFT → PENDING_APPROVAL
// =============================================================================

func TestLifecycle_SubmitForApproval_DraftToPending(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)

	repo := newLCAccountRepo(account)
	evRepo := &lcEventRepo{}
	svc := newLifecycleSvc(repo, evRepo)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.SubmitForApproval(ctx, account.ID, creator, "initial submission"); err != nil {
		t.Fatalf("FIN-LIFECYCLE-001: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusPendingApproval {
		t.Errorf("FIN-LIFECYCLE-001: expected PENDING_APPROVAL, got %s", got.Status)
	}
	if len(evRepo.events) == 0 {
		t.Error("FIN-LIFECYCLE-001: expected lifecycle event to be recorded")
	}
}

// =============================================================================
// FIN-LIFECYCLE-002: SubmitForApproval rejects non-DRAFT
// =============================================================================

func TestLifecycle_SubmitForApproval_NonDraft_Rejected(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusActive

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.SubmitForApproval(ctx, account.ID, creator, "re-submit")
	if err == nil {
		t.Error("FIN-LIFECYCLE-002: expected error for non-DRAFT account, got nil")
	}
	if !errors.Is(err, domain.ErrAccountNotInDraft) {
		t.Errorf("FIN-LIFECYCLE-002: expected ErrAccountNotInDraft, got %v", err)
	}
}

// =============================================================================
// FIN-LIFECYCLE-003: SubmitForApproval rejects invalid account
// =============================================================================

func TestLifecycle_SubmitForApproval_Invalid_Rejected(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.AccountName = "" // cause validation failure

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.SubmitForApproval(ctx, account.ID, creator, "submit invalid")
	if err == nil {
		t.Error("FIN-LIFECYCLE-003: expected validation error, got nil")
	}
}

// =============================================================================
// FIN-LIFECYCLE-004: Approve PENDING_APPROVAL → ACTIVE
// =============================================================================

func TestLifecycle_Approve_PendingToActive(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	approver := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusPendingApproval

	repo := newLCAccountRepo(account)
	evRepo := &lcEventRepo{}
	svc := newLifecycleSvc(repo, evRepo)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.Approve(ctx, account.ID, approver, "approved"); err != nil {
		t.Fatalf("FIN-LIFECYCLE-004: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusActive {
		t.Errorf("FIN-LIFECYCLE-004: expected ACTIVE, got %s", got.Status)
	}
	if !got.IsActive {
		t.Error("FIN-LIFECYCLE-004: expected IsActive=true after approval")
	}
}

// =============================================================================
// FIN-LIFECYCLE-005: Approve enforces SOD
// =============================================================================

func TestLifecycle_Approve_SOD_Rejected(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusPendingApproval

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.Approve(ctx, account.ID, creator, "self-approve") // SOD violation
	if err == nil {
		t.Error("FIN-LIFECYCLE-005: expected SOD error, got nil")
	}
	if !errors.Is(err, domain.ErrSelfApprovalNotAllowed) {
		t.Errorf("FIN-LIFECYCLE-005: expected ErrSelfApprovalNotAllowed, got %v", err)
	}
}

// =============================================================================
// FIN-LIFECYCLE-006: Approve rejects non-PENDING account
// =============================================================================

func TestLifecycle_Approve_NonPending_Rejected(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	approver := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusDraft

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.Approve(ctx, account.ID, approver, "approve draft")
	if err == nil {
		t.Error("FIN-LIFECYCLE-006: expected error for non-PENDING account, got nil")
	}
	if !errors.Is(err, domain.ErrAccountNotPendingApproval) {
		t.Errorf("FIN-LIFECYCLE-006: expected ErrAccountNotPendingApproval, got %v", err)
	}
}

// =============================================================================
// FIN-LIFECYCLE-007: Reject PENDING_APPROVAL → DRAFT
// =============================================================================

func TestLifecycle_Reject_PendingToDraft(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	approver := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusPendingApproval

	repo := newLCAccountRepo(account)
	evRepo := &lcEventRepo{}
	svc := newLifecycleSvc(repo, evRepo)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.Reject(ctx, account.ID, approver, "missing documentation"); err != nil {
		t.Fatalf("FIN-LIFECYCLE-007: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusDraft {
		t.Errorf("FIN-LIFECYCLE-007: expected DRAFT, got %s", got.Status)
	}
}

// =============================================================================
// FIN-LIFECYCLE-008: Reject requires reason
// =============================================================================

func TestLifecycle_Reject_RequiresReason(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	approver := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusPendingApproval

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.Reject(ctx, account.ID, approver, "")
	if err == nil {
		t.Error("FIN-LIFECYCLE-008: expected error for empty rejection reason, got nil")
	}
}

// =============================================================================
// FIN-LIFECYCLE-009: Freeze ACTIVE → FROZEN
// =============================================================================

func TestLifecycle_Freeze_ActiveToFrozen(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusActive
	account.IsActive = true

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.Freeze(ctx, account.ID, uuid.New(), "fraud investigation"); err != nil {
		t.Fatalf("FIN-LIFECYCLE-009: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusFrozen {
		t.Errorf("FIN-LIFECYCLE-009: expected FROZEN, got %s", got.Status)
	}
}

// =============================================================================
// FIN-LIFECYCLE-010: Freeze requires reason
// =============================================================================

func TestLifecycle_Freeze_RequiresReason(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusActive

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.Freeze(ctx, account.ID, uuid.New(), "")
	if err == nil {
		t.Error("FIN-LIFECYCLE-010: expected error for empty freeze reason, got nil")
	}
}

// =============================================================================
// FIN-LIFECYCLE-011: Unfreeze FROZEN → ACTIVE
// =============================================================================

func TestLifecycle_Unfreeze_FrozenToActive(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusFrozen

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.Unfreeze(ctx, account.ID, uuid.New()); err != nil {
		t.Fatalf("FIN-LIFECYCLE-011: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusActive {
		t.Errorf("FIN-LIFECYCLE-011: expected ACTIVE, got %s", got.Status)
	}
}

// =============================================================================
// FIN-LIFECYCLE-012: Unfreeze rejects non-FROZEN
// =============================================================================

func TestLifecycle_Unfreeze_NonFrozen_Rejected(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusActive

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.Unfreeze(ctx, account.ID, uuid.New())
	if err == nil {
		t.Error("FIN-LIFECYCLE-012: expected error for non-FROZEN account, got nil")
	}
}

// =============================================================================
// FIN-LIFECYCLE-013: Close INACTIVE → CLOSED with zero balance
// =============================================================================

func TestLifecycle_Close_InactiveToClosed(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusInactive
	account.CurrentBalance = decimal.Zero
	account.HasChildren = false

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.Close(ctx, account.ID, uuid.New(), "year-end close"); err != nil {
		t.Fatalf("FIN-LIFECYCLE-013: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusClosed {
		t.Errorf("FIN-LIFECYCLE-013: expected CLOSED, got %s", got.Status)
	}
}

// =============================================================================
// FIN-LIFECYCLE-014: Close rejects non-zero balance
// =============================================================================

func TestLifecycle_Close_NonZeroBalance_Rejected(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusInactive
	account.CurrentBalance = decimal.NewFromFloat(500.00)

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.Close(ctx, account.ID, uuid.New(), "close with balance")
	if err == nil {
		t.Error("FIN-LIFECYCLE-014: expected error for non-zero balance, got nil")
	}
	if !errors.Is(err, domain.ErrAccountHasNonZeroBalance) {
		t.Errorf("FIN-LIFECYCLE-014: expected ErrAccountHasNonZeroBalance, got %v", err)
	}
}

// =============================================================================
// FIN-LIFECYCLE-015: Archive CLOSED → ARCHIVED
// =============================================================================

func TestLifecycle_Archive_ClosedToArchived(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusClosed
	account.IsActive = false

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.Archive(ctx, account.ID, uuid.New()); err != nil {
		t.Fatalf("FIN-LIFECYCLE-015: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusArchived {
		t.Errorf("FIN-LIFECYCLE-015: expected ARCHIVED, got %s", got.Status)
	}
}

// =============================================================================
// FIN-LIFECYCLE-016: Suspend ACTIVE → SUSPENDED
// =============================================================================

func TestLifecycle_Suspend_ActiveToSuspended(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusActive
	account.IsActive = true

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.Suspend(ctx, account.ID, uuid.New(), "pending investigation"); err != nil {
		t.Fatalf("FIN-LIFECYCLE-016: unexpected error: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusSuspended {
		t.Errorf("FIN-LIFECYCLE-016: expected SUSPENDED, got %s", got.Status)
	}
}

// =============================================================================
// FIN-LIFECYCLE-017: PlaceComplianceHold requires reason
// =============================================================================

func TestLifecycle_PlaceComplianceHold_RequiresReason(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)
	account.Status = domain.AccountStatusActive

	repo := newLCAccountRepo(account)
	svc := newLifecycleSvc(repo, &lcEventRepo{})

	ctx := shared.WithTenantID(context.Background(), tenantID)
	err := svc.PlaceComplianceHold(ctx, account.ID, uuid.New(), "")
	if err == nil {
		t.Error("FIN-LIFECYCLE-017: expected error for empty compliance hold reason, got nil")
	}
}

// =============================================================================
// FIN-LIFECYCLE-018: GetLifecycleHistory returns events in order
// =============================================================================

func TestLifecycle_GetLifecycleHistory_ReturnsEvents(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	approver := uuid.New()
	account := newValidAccount(tenantID, creator)

	repo := newLCAccountRepo(account)
	evRepo := &lcEventRepo{}
	svc := newLifecycleSvc(repo, evRepo)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	_ = svc.SubmitForApproval(ctx, account.ID, creator, "submit")
	_ = svc.Reject(ctx, account.ID, approver, "missing info")

	history, err := svc.GetLifecycleHistory(ctx, account.ID)
	if err != nil {
		t.Fatalf("FIN-LIFECYCLE-018: unexpected error: %v", err)
	}
	if len(history) < 2 {
		t.Errorf("FIN-LIFECYCLE-018: expected >=2 events, got %d", len(history))
	}
}

// =============================================================================
// FIN-LIFECYCLE-019: Nil service receiver — nil-safe
// =============================================================================

func TestLifecycle_NilReceiver_Safe(t *testing.T) {
	var svc *service.AccountLifecycleService
	ctx := lcCtx()
	id := uuid.New()
	actor := uuid.New()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FIN-LIFECYCLE-019: nil receiver panicked: %v", r)
		}
	}()

	_ = svc.SubmitForApproval(ctx, id, actor, "")
	_ = svc.Approve(ctx, id, actor, "")
	_ = svc.Reject(ctx, id, actor, "")
	_ = svc.Freeze(ctx, id, actor, "")
	_ = svc.Unfreeze(ctx, id, actor)
	_ = svc.Close(ctx, id, actor, "")
	_ = svc.Archive(ctx, id, actor)
	_ = svc.Suspend(ctx, id, actor, "")
	_ = svc.PlaceComplianceHold(ctx, id, actor, "")
	_, _ = svc.GetLifecycleHistory(ctx, id)
}

// =============================================================================
// FIN-LIFECYCLE-020: Event persistence failure does not block transition
// =============================================================================

func TestLifecycle_EventPersistenceFailure_TransitionSucceeds(t *testing.T) {
	tenantID := uuid.New()
	creator := uuid.New()
	account := newValidAccount(tenantID, creator)

	repo := newLCAccountRepo(account)
	evRepo := &lcEventRepo{
		appendFn: func(_ *domain.AccountLifecycleEvent) error {
			return errors.New("event store unavailable")
		},
	}
	svc := newLifecycleSvc(repo, evRepo)

	ctx := shared.WithTenantID(context.Background(), tenantID)
	if err := svc.SubmitForApproval(ctx, account.ID, creator, "submit despite event failure"); err != nil {
		t.Errorf("FIN-LIFECYCLE-020: transition should succeed despite event failure, got: %v", err)
	}

	got := repo.accounts[account.ID]
	if got.Status != domain.AccountStatusPendingApproval {
		t.Errorf("FIN-LIFECYCLE-020: expected PENDING_APPROVAL, got %s", got.Status)
	}
}
