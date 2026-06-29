// Package service_test — Phase 17 shared test stubs.
//
// Provides stubs for:
//   - COASnapshotRepository (p17StubSnapshotRepo)
//   - AccountsRepository   (p17StubAccountRepo — wraps p16StubAccountRepo)
//
// All stubs embed or hook-enable every method so tests only override what they need.
package service_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
)

// =============================================================================
// p17StubSnapshotRepo — COASnapshotRepository stub
// =============================================================================

type p17StubSnapshotRepo struct {
	fnCreate       func(ctx context.Context, snap *domain.COASnapshot) error
	fnGetByID      func(ctx context.Context, id uuid.UUID) (*domain.COASnapshot, error)
	fnGetForPeriod func(ctx context.Context, tenantID uuid.UUID, periodID uuid.UUID) (*domain.COASnapshot, error)
	fnGetAtTime    func(ctx context.Context, tenantID uuid.UUID, asOfTime time.Time) (*domain.COASnapshot, error)
	fnList         func(ctx context.Context, tenantID uuid.UUID) ([]*domain.COASnapshot, error)

	// Simple in-memory store — populated by Create, queried by Get*.
	store []*domain.COASnapshot
}

func (r *p17StubSnapshotRepo) Create(ctx context.Context, snap *domain.COASnapshot) error {
	if r.fnCreate != nil {
		return r.fnCreate(ctx, snap)
	}
	r.store = append(r.store, snap)
	return nil
}

func (r *p17StubSnapshotRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.COASnapshot, error) {
	if r.fnGetByID != nil {
		return r.fnGetByID(ctx, id)
	}
	for _, s := range r.store {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, nil
}

func (r *p17StubSnapshotRepo) GetForPeriod(ctx context.Context, tenantID uuid.UUID, periodID uuid.UUID) (*domain.COASnapshot, error) {
	if r.fnGetForPeriod != nil {
		return r.fnGetForPeriod(ctx, tenantID, periodID)
	}
	for _, s := range r.store {
		if s.TenantID == tenantID && s.PeriodID != nil && *s.PeriodID == periodID {
			return s, nil
		}
	}
	return nil, nil
}

func (r *p17StubSnapshotRepo) GetAtTime(ctx context.Context, tenantID uuid.UUID, asOfTime time.Time) (*domain.COASnapshot, error) {
	if r.fnGetAtTime != nil {
		return r.fnGetAtTime(ctx, tenantID, asOfTime)
	}
	// Return the most recent snapshot on or before asOfTime.
	var best *domain.COASnapshot
	for _, s := range r.store {
		if s.TenantID != tenantID {
			continue
		}
		if !s.SnapshotAt.After(asOfTime) {
			if best == nil || s.SnapshotAt.After(best.SnapshotAt) {
				best = s
			}
		}
	}
	return best, nil
}

func (r *p17StubSnapshotRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.COASnapshot, error) {
	if r.fnList != nil {
		return r.fnList(ctx, tenantID)
	}
	var out []*domain.COASnapshot
	for _, s := range r.store {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

// =============================================================================
// p17AccountRepo — minimal AccountsRepository for Phase 17 (wraps p16)
// =============================================================================

type p17AccountRepo struct {
	p16StubAccountRepo
	accounts []*domain.Accounts
	listErr  error
}

func newP17AccountRepo(accts ...*domain.Accounts) *p17AccountRepo {
	return &p17AccountRepo{accounts: accts}
}

func (r *p17AccountRepo) List(_ context.Context, _ *domain.AccountFilter) ([]*domain.Accounts, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.accounts, nil
}

func (r *p17AccountRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Accounts, error) {
	for _, a := range r.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}
