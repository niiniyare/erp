// Package repository_test provides unit tests for cache consistency in the
// finance period repository. These tests do NOT require a database — they run
// in the standard `go test` pass.
package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"

	db "awo.so/db/sqlc"
	"awo.so/internal/core/finance/domain"
	"awo.so/internal/core/finance/repository"
	"awo.so/internal/platform/cache"
	"awo.so/internal/shared"
	"awo.so/internal/shared/tracing"
)

// cacheHitFY returns a DoAndReturn function that unmarshals a FiscalYear into dest.
func cacheHitFY(fy domain.FiscalYear) func(context.Context, string, any) error {
	return func(_ context.Context, _ string, dest any) error {
		if p, ok := dest.(*domain.FiscalYear); ok {
			*p = fy
			return nil
		}
		return errors.New("unexpected dest type")
	}
}

// cacheHitPeriod returns a DoAndReturn function that unmarshals an AccountingPeriod into dest.
func cacheHitPeriod(p domain.AccountingPeriod) func(context.Context, string, any) error {
	return func(_ context.Context, _ string, dest any) error {
		if pp, ok := dest.(*domain.AccountingPeriod); ok {
			*pp = p
			return nil
		}
		return errors.New("unexpected dest type")
	}
}

// ============================================================================
// FIN-CACHE-001: cache hit for GetFiscalYearByYear skips DB entirely
// ============================================================================

func TestGetFiscalYearByYear_CacheHit_SkipsDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	// Store must never be called when cache hits.
	mockStore.EXPECT().WithTenantFromCtx(gomock.Any(), gomock.Any()).Times(0)

	want := domain.FiscalYear{
		ID:        uuid.New(),
		Name:      "FY2026",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Year:      2026,
	}
	mockCache.EXPECT().
		GetMemory(gomock.Any(), "fiscal_year:year:2026", gomock.Any()).
		DoAndReturn(cacheHitFY(want))

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	got, err := repo.GetFiscalYearByYear(ctx, uuid.New(), 2026)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got ID %v, want %v", got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Errorf("got Name %q, want %q", got.Name, want.Name)
	}
}

// ============================================================================
// FIN-CACHE-002: cache hit for GetPeriodForDate skips DB entirely
// ============================================================================

func TestGetPeriodForDate_CacheHit_SkipsDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	mockStore.EXPECT().WithTenantFromCtx(gomock.Any(), gomock.Any()).Times(0)

	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	want := domain.AccountingPeriod{
		ID:     uuid.New(),
		Name:   "Mar 2026",
		Status: domain.PeriodStatusOpen,
	}
	mockCache.EXPECT().
		GetMemory(gomock.Any(), "period:date:2026-03-15", gomock.Any()).
		DoAndReturn(cacheHitPeriod(want))

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	got, err := repo.GetPeriodForDate(ctx, uuid.New(), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("got ID %v, want %v", got.ID, want.ID)
	}
}

// ============================================================================
// FIN-CACHE-003: nil cache is safe — no panic
// ============================================================================

func TestGetFiscalYearByYear_NilCache_NoPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	// With nil cache the DB path is taken; WithTenantFromCtx will be called once.
	mockStore.EXPECT().
		WithTenantFromCtx(gomock.Any(), gomock.Any()).
		Return(errors.New("no db available")).
		Times(1)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, nil)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	// Must not panic; must return an error from the store.
	_, err := repo.GetFiscalYearByYear(ctx, uuid.New(), 2026)
	if err == nil {
		t.Error("expected error when store returns error, got nil")
	}
}

func TestGetPeriodForDate_NilCache_NoPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockStore.EXPECT().
		WithTenantFromCtx(gomock.Any(), gomock.Any()).
		Return(errors.New("no db available")).
		Times(1)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, nil)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	_, err := repo.GetPeriodForDate(ctx, uuid.New(), time.Now())
	if err == nil {
		t.Error("expected error when store returns error, got nil")
	}
}

// ============================================================================
// FIN-CACHE-004: cache miss falls through to DB (verified by call count)
// ============================================================================

func TestGetFiscalYearByYear_CacheMiss_HitsDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	// Cache miss.
	mockCache.EXPECT().
		GetMemory(gomock.Any(), "fiscal_year:year:2026", gomock.Any()).
		Return(cache.ErrCacheMiss)

	// DB must be called once.
	mockStore.EXPECT().
		WithTenantFromCtx(gomock.Any(), gomock.Any()).
		Return(errors.New("simulated db error")).
		Times(1)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	_, err := repo.GetFiscalYearByYear(ctx, uuid.New(), 2026)
	if err == nil {
		t.Error("expected error from DB, got nil")
	}
}

func TestGetPeriodForDate_CacheMiss_HitsDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	mockCache.EXPECT().
		GetMemory(gomock.Any(), "period:date:2026-03-15", gomock.Any()).
		Return(cache.ErrCacheMiss)

	mockStore.EXPECT().
		WithTenantFromCtx(gomock.Any(), gomock.Any()).
		Return(errors.New("simulated db error")).
		Times(1)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	_, err := repo.GetPeriodForDate(ctx, uuid.New(), time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Error("expected error from DB, got nil")
	}
}

// ============================================================================
// FIN-CACHE-005: cache key format is deterministic and date-scoped
// ============================================================================

// TestCacheKeyFormat_Deterministic verifies repeated calls with the same args
// produce the same cache key (no randomness in key construction).
func TestCacheKeyFormat_Deterministic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	want := domain.FiscalYear{ID: uuid.New(), Year: 2026}

	// Both calls must hit the SAME key: "fiscal_year:year:2026"
	mockCache.EXPECT().
		GetMemory(gomock.Any(), "fiscal_year:year:2026", gomock.Any()).
		DoAndReturn(cacheHitFY(want)).
		Times(2)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	for i := 0; i < 2; i++ {
		got, err := repo.GetFiscalYearByYear(ctx, uuid.New(), 2026)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if got.ID != want.ID {
			t.Errorf("call %d: ID mismatch", i)
		}
	}
}

// ============================================================================
// FIN-CACHE-006: concurrent cache reads — no data race
// ============================================================================

// TestConcurrentCacheReads_NoDataRace verifies goroutines can safely read from
// cache simultaneously. Detected by go test -race.
func TestConcurrentCacheReads_NoDataRace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	want := domain.FiscalYear{ID: uuid.New(), Year: 2027}
	const goroutines = 20

	mockCache.EXPECT().
		GetMemory(gomock.Any(), "fiscal_year:year:2027", gomock.Any()).
		DoAndReturn(cacheHitFY(want)).
		Times(goroutines)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := shared.WithTenantID(context.Background(), uuid.New())
			got, err := repo.GetFiscalYearByYear(ctx, uuid.New(), 2027)
			if err != nil {
				errs <- err
				return
			}
			if got.ID != want.ID {
				errs <- errors.New("ID mismatch in concurrent read")
			}
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent read error: %v", err)
	}
}

// ============================================================================
// FIN-CACHE-007: UpdateFiscalYear cache invalidation — key is deleted
// ============================================================================

// TestUpdateFiscalYear_InvalidatesCache verifies that on successful update,
// the cache key for the fiscal year is deleted.
// The DB path is simulated by having WithTenantFromCtx return nil (success).
// Since txFrom(store) fails (MockStore is not a TxStore), the update itself
// will error, so we can only verify the cache DELETE is NOT called on error
// — which proves no stale invalidation on failed updates.
func TestUpdateFiscalYear_NoInvalidation_OnDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	// DB call fails inside the callback (txFrom fails — MockStore ≠ TxStore).
	// We verify the cache DeleteMemory is NOT called on error.
	mockStore.EXPECT().
		WithTenantFromCtx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, db.Store) error) error {
			// Pass the mock store itself into the callback.
			// txFrom will fail since MockStore doesn't implement TxStore.
			return fn(ctx, mockStore)
		}).
		Times(1)

	// DeleteMemory must NOT be called since the update errored.
	mockCache.EXPECT().DeleteMemory(gomock.Any(), gomock.Any()).Times(0)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	fy := &domain.FiscalYear{
		ID:        uuid.New(),
		Name:      "FY2026",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		IsClosed:  true,
	}
	err := repo.UpdateFiscalYear(ctx, fy)
	// Expect an error (txFrom fails).
	if err == nil {
		t.Error("expected error from DB path, got nil")
	}
}

// ============================================================================
// FIN-CACHE-008: UpdatePeriod cache invalidation — keys are deleted on success
// Same pattern as above: verify no stale DELETE on DB error.
// ============================================================================

func TestUpdatePeriod_NoInvalidation_OnDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := db.NewMockStore(ctrl)
	mockCache := cache.NewMockService(ctrl)

	mockStore.EXPECT().
		WithTenantFromCtx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, db.Store) error) error {
			return fn(ctx, mockStore) // txFrom will fail
		}).
		Times(1)

	mockCache.EXPECT().DeleteMemory(gomock.Any(), gomock.Any()).Times(0)

	repo := repository.NewPeriodRepository(mockStore, tracing.NewNoOpService(), nil, nil, mockCache)
	ctx := shared.WithTenantID(context.Background(), uuid.New())

	p := &domain.AccountingPeriod{
		ID:        uuid.New(),
		Status:    domain.PeriodStatusSoftClosed,
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	}
	err := repo.UpdatePeriod(ctx, p)
	if err == nil {
		t.Error("expected error from DB path, got nil")
	}
}
