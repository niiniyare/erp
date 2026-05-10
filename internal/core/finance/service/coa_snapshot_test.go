// Package service_test — COASnapshotService tests.
//
// Tests (FIN-SNAP-*):
//
//	001 — CreateSnapshot captures all accounts, stores correct entry count
//	002 — CreateSnapshot produces non-empty SHA-256 hash
//	003 — CreateSnapshot on empty COA produces empty snapshot (entry_count=0)
//	004 — ValidateSnapshotHash returns true for untampered snapshot
//	005 — ValidateSnapshotHash returns false after entries reordered (tamper simulation)
//	006 — GetByID returns the snapshot after creation
//	007 — GetAtTime returns the snapshot taken on or before the query time
//	008 — GetAtTime returns nil when no snapshot exists before the query time
//	009 — CreateSnapshot fails when repo returns error
//	010 — Two snapshots of same COA produce identical hashes (deterministic)
//	011 — Snapshot after COA change produces different hash
//	012 — CreateSnapshot without tenant ID in context returns error
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

// helpers ----------------------------------------------------------------

func snapCtx() context.Context {
	return shared.WithTenantID(context.Background(), uuid.New())
}

func newSnapSvc(accounts *p17AccountRepo, snapRepo *p17StubSnapshotRepo) *service.COASnapshotService {
	return service.NewCOASnapshotService(accounts, snapRepo, metrics.NewNoOpMetricsProvider())
}

func makeAccount(code string) *domain.Accounts {
	curr := "USD"
	return &domain.Accounts{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		AccountCode:   code,
		AccountName:   "Acct " + code,
		RootType:      domain.RootTypeAsset,
		NormalBalance: domain.NormalBalanceDebit,
		Status:        domain.AccountStatusActive,
		IsActive:      true,
		CurrencyCode:  &curr,
		CurrentBalance: decimal.Zero,
	}
}

func makeCreateReq() service.CreateSnapshotRequest {
	return service.CreateSnapshotRequest{CreatedBy: uuid.New()}
}

// FIN-SNAP-001 -----------------------------------------------------------

func TestSnapshot_Create_CapturesAccountCount(t *testing.T) {
	a1 := makeAccount("10010001")
	a2 := makeAccount("20010001")
	accounts := newP17AccountRepo(a1, a2)
	snapRepo := &p17StubSnapshotRepo{}

	svc := newSnapSvc(accounts, snapRepo)
	snap, err := svc.CreateSnapshot(snapCtx(), makeCreateReq())
	if err != nil {
		t.Fatalf("FIN-SNAP-001: unexpected error: %v", err)
	}
	if snap.EntryCount != 2 {
		t.Errorf("FIN-SNAP-001: expected EntryCount=2, got %d", snap.EntryCount)
	}
	if len(snap.Accounts) != 2 {
		t.Errorf("FIN-SNAP-001: expected 2 snapshot entries, got %d", len(snap.Accounts))
	}
}

// FIN-SNAP-002 -----------------------------------------------------------

func TestSnapshot_Create_ProducesNonEmptyHash(t *testing.T) {
	a1 := makeAccount("10010001")
	accounts := newP17AccountRepo(a1)
	snapRepo := &p17StubSnapshotRepo{}

	svc := newSnapSvc(accounts, snapRepo)
	snap, err := svc.CreateSnapshot(snapCtx(), makeCreateReq())
	if err != nil {
		t.Fatalf("FIN-SNAP-002: unexpected error: %v", err)
	}
	if snap.Hash == "" {
		t.Error("FIN-SNAP-002: expected non-empty hash")
	}
	if len(snap.Hash) != 64 {
		t.Errorf("FIN-SNAP-002: expected 64-char SHA-256 hex, got len=%d", len(snap.Hash))
	}
}

// FIN-SNAP-003 -----------------------------------------------------------

func TestSnapshot_Create_EmptyCOA(t *testing.T) {
	accounts := newP17AccountRepo() // no accounts
	snapRepo := &p17StubSnapshotRepo{}

	svc := newSnapSvc(accounts, snapRepo)
	snap, err := svc.CreateSnapshot(snapCtx(), makeCreateReq())
	if err != nil {
		t.Fatalf("FIN-SNAP-003: unexpected error: %v", err)
	}
	if snap.EntryCount != 0 {
		t.Errorf("FIN-SNAP-003: expected EntryCount=0, got %d", snap.EntryCount)
	}
}

// FIN-SNAP-004 -----------------------------------------------------------

func TestSnapshot_ValidateHash_Untampered(t *testing.T) {
	a1 := makeAccount("10010001")
	accounts := newP17AccountRepo(a1)
	snapRepo := &p17StubSnapshotRepo{}

	svc := newSnapSvc(accounts, snapRepo)
	snap, err := svc.CreateSnapshot(snapCtx(), makeCreateReq())
	if err != nil {
		t.Fatalf("FIN-SNAP-004: create error: %v", err)
	}

	ctx := snapCtx()
	valid, err := svc.ValidateSnapshotHash(ctx, snap.ID)
	if err != nil {
		t.Fatalf("FIN-SNAP-004: validate error: %v", err)
	}
	if !valid {
		t.Error("FIN-SNAP-004: expected hash to be valid for untampered snapshot")
	}
}

// FIN-SNAP-005 -----------------------------------------------------------

func TestSnapshot_ValidateHash_TamperedReturnsFalse(t *testing.T) {
	a1 := makeAccount("10010001")
	accounts := newP17AccountRepo(a1)
	snapRepo := &p17StubSnapshotRepo{}

	svc := newSnapSvc(accounts, snapRepo)
	snap, err := svc.CreateSnapshot(snapCtx(), makeCreateReq())
	if err != nil {
		t.Fatalf("FIN-SNAP-005: create error: %v", err)
	}

	// Tamper: modify stored hash to simulate corruption.
	snap.Hash = "0000000000000000000000000000000000000000000000000000000000000000"

	valid, err := svc.ValidateSnapshotHash(snapCtx(), snap.ID)
	if err != nil {
		t.Fatalf("FIN-SNAP-005: validate error: %v", err)
	}
	if valid {
		t.Error("FIN-SNAP-005: expected hash validation to fail for tampered snapshot")
	}
}

// FIN-SNAP-006 -----------------------------------------------------------

func TestSnapshot_GetByID_ReturnsAfterCreate(t *testing.T) {
	a1 := makeAccount("10010001")
	accounts := newP17AccountRepo(a1)
	snapRepo := &p17StubSnapshotRepo{}

	svc := newSnapSvc(accounts, snapRepo)
	snap, err := svc.CreateSnapshot(snapCtx(), makeCreateReq())
	if err != nil {
		t.Fatalf("FIN-SNAP-006: create error: %v", err)
	}

	got, err := svc.GetByID(snapCtx(), snap.ID)
	if err != nil {
		t.Fatalf("FIN-SNAP-006: get error: %v", err)
	}
	if got == nil || got.ID != snap.ID {
		t.Error("FIN-SNAP-006: expected to retrieve created snapshot by ID")
	}
}

// FIN-SNAP-007 -----------------------------------------------------------

func TestSnapshot_GetAtTime_ReturnsSnapshotBeforeQuery(t *testing.T) {
	a1 := makeAccount("10010001")
	tenantID := uuid.New()
	accounts := newP17AccountRepo(a1)
	snapRepo := &p17StubSnapshotRepo{}

	ctx := shared.WithTenantID(context.Background(), tenantID)
	svc := newSnapSvc(accounts, snapRepo)

	// Pre-populate a snapshot in the repo directly.
	past := time.Now().Add(-24 * time.Hour)
	existing := &domain.COASnapshot{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SnapshotAt: past,
		Hash:       "abc",
		CreatedAt:  past,
		CreatedBy:  uuid.New(),
	}
	snapRepo.store = append(snapRepo.store, existing)

	got, err := svc.GetAtTime(ctx, time.Now())
	if err != nil {
		t.Fatalf("FIN-SNAP-007: get error: %v", err)
	}
	if got == nil {
		t.Fatal("FIN-SNAP-007: expected a snapshot, got nil")
	}
	if got.ID != existing.ID {
		t.Error("FIN-SNAP-007: expected the pre-existing snapshot")
	}
}

// FIN-SNAP-008 -----------------------------------------------------------

func TestSnapshot_GetAtTime_NilWhenNoneExists(t *testing.T) {
	accounts := newP17AccountRepo()
	snapRepo := &p17StubSnapshotRepo{}
	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	svc := newSnapSvc(accounts, snapRepo)
	got, err := svc.GetAtTime(ctx, time.Now())
	if err != nil {
		t.Fatalf("FIN-SNAP-008: unexpected error: %v", err)
	}
	if got != nil {
		t.Error("FIN-SNAP-008: expected nil when no snapshot exists")
	}
}

// FIN-SNAP-009 -----------------------------------------------------------

func TestSnapshot_Create_PropagatesRepoError(t *testing.T) {
	accounts := newP17AccountRepo(makeAccount("10010001"))
	snapRepo := &p17StubSnapshotRepo{
		fnCreate: func(_ context.Context, _ *domain.COASnapshot) error {
			return errors.New("DB write failed")
		},
	}

	svc := newSnapSvc(accounts, snapRepo)
	_, err := svc.CreateSnapshot(snapCtx(), makeCreateReq())
	if err == nil {
		t.Error("FIN-SNAP-009: expected error from repo, got nil")
	}
}

// FIN-SNAP-010 -----------------------------------------------------------

func TestSnapshot_Create_Deterministic_SameCOA(t *testing.T) {
	// Two separate snapshots of the same COA set must have identical hashes.
	a1 := makeAccount("10010001")
	a2 := makeAccount("20010001")

	// Create two separate repo instances with identical account sets.
	accounts1 := newP17AccountRepo(a1, a2)
	accounts2 := newP17AccountRepo(a1, a2) // same objects
	snapRepo1 := &p17StubSnapshotRepo{}
	snapRepo2 := &p17StubSnapshotRepo{}

	svc1 := newSnapSvc(accounts1, snapRepo1)
	svc2 := newSnapSvc(accounts2, snapRepo2)

	snap1, err1 := svc1.CreateSnapshot(snapCtx(), makeCreateReq())
	snap2, err2 := svc2.CreateSnapshot(snapCtx(), makeCreateReq())
	if err1 != nil || err2 != nil {
		t.Fatalf("FIN-SNAP-010: create errors: %v / %v", err1, err2)
	}

	if snap1.Hash != snap2.Hash {
		t.Errorf("FIN-SNAP-010: hashes differ for identical COA: %s vs %s", snap1.Hash, snap2.Hash)
	}
}

// FIN-SNAP-011 -----------------------------------------------------------

func TestSnapshot_Create_DifferentCOA_DifferentHash(t *testing.T) {
	a1 := makeAccount("10010001")
	a2 := makeAccount("20010001")

	accounts1 := newP17AccountRepo(a1)
	accounts2 := newP17AccountRepo(a1, a2) // a2 added
	snapRepo1 := &p17StubSnapshotRepo{}
	snapRepo2 := &p17StubSnapshotRepo{}

	svc1 := newSnapSvc(accounts1, snapRepo1)
	svc2 := newSnapSvc(accounts2, snapRepo2)

	snap1, _ := svc1.CreateSnapshot(snapCtx(), makeCreateReq())
	snap2, _ := svc2.CreateSnapshot(snapCtx(), makeCreateReq())

	if snap1.Hash == snap2.Hash {
		t.Error("FIN-SNAP-011: expected different hashes for different COA sets")
	}
}

// FIN-SNAP-012 -----------------------------------------------------------

func TestSnapshot_Create_NoTenantInContext_ReturnsError(t *testing.T) {
	accounts := newP17AccountRepo(makeAccount("10010001"))
	snapRepo := &p17StubSnapshotRepo{}
	svc := newSnapSvc(accounts, snapRepo)

	_, err := svc.CreateSnapshot(context.Background(), makeCreateReq()) // no tenant
	if err == nil {
		t.Error("FIN-SNAP-012: expected error when tenant ID missing, got nil")
	}
}
