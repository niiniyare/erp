// Package service — COA Snapshot Service.
//
// COASnapshotService creates and retrieves immutable point-in-time snapshots
// of the Chart of Accounts structure. Snapshots are the binding contract
// between reports and the COA: every financial report must reference a
// snapshot, not the live COA, so historical reports remain reproducible even
// after structural mutations.
//
// # Snapshot lifecycle
//
//  1. CreateSnapshot — loads all active accounts, computes deterministic SHA-256
//     hash over the canonical ordered set, persists the COASnapshot.
//  2. GetSnapshotForReport — retrieves the snapshot identified in a report
//     request; falls back to GetAtTime when no explicit snapshot is provided.
//  3. ValidateSnapshotHash — recomputes the hash from stored entries and
//     compares against the stored hash to detect tampering.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/finance/domain"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// COASnapshotService creates and retrieves immutable COA snapshots.
type COASnapshotService struct {
	accounts  domain.AccountsRepository
	snapshots domain.COASnapshotRepository
	metrics   metrics.MetricsProvider
}

// NewCOASnapshotService constructs the snapshot service.
func NewCOASnapshotService(
	accounts domain.AccountsRepository,
	snapshots domain.COASnapshotRepository,
	m metrics.MetricsProvider,
) *COASnapshotService {
	return &COASnapshotService{
		accounts:  accounts,
		snapshots: snapshots,
		metrics:   m,
	}
}

// =============================================================================
// Snapshot creation
// =============================================================================

// CreateSnapshotRequest specifies inputs for a new COA snapshot.
type CreateSnapshotRequest struct {
	// PeriodID optionally links the snapshot to a fiscal period.
	// When set, the snapshot is interpreted as the COA state at period open.
	PeriodID  *uuid.UUID
	CreatedBy uuid.UUID
}

// CreateSnapshot captures the current COA structure for the tenant in ctx
// and persists it as an immutable snapshot.
//
// The snapshot hash is a deterministic SHA-256 of all account entries in
// canonical (AccountCode ascending) order. Recomputing from the same entries
// must produce the same hash.
func (s *COASnapshotService) CreateSnapshot(
	ctx context.Context,
	req CreateSnapshotRequest,
) (*domain.COASnapshot, error) {
	if s == nil {
		return nil, fmt.Errorf("COASnapshotService is nil")
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("create snapshot: tenant ID missing from context")
	}

	// Load all accounts for the tenant.
	accounts, err := s.accounts.List(ctx, &domain.AccountFilter{})
	if err != nil {
		return nil, fmt.Errorf("create snapshot: load accounts: %w", err)
	}

	// Build snapshot entries from live accounts.
	entries := make([]domain.COASnapshotEntry, 0, len(accounts))
	for _, a := range accounts {
		entry := domain.COASnapshotEntry{
			AccountID:     a.ID,
			AccountCode:   a.AccountCode,
			AccountName:   a.AccountName,
			RootType:      a.RootType,
			NormalBalance: a.NormalBalance,
			Status:        a.Status,
			ParentID:      a.ParentAccountID,
			AccountLevel:  a.AccountLevel,
			CurrencyCode:  a.CurrencyCode,
			IsActive:      a.IsActive,
		}
		entries = append(entries, entry)
	}

	// Sort canonically by AccountCode for deterministic hash.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].AccountCode < entries[j].AccountCode
	})

	hash, err := computeSnapshotHash(entries)
	if err != nil {
		return nil, fmt.Errorf("create snapshot: compute hash: %w", err)
	}

	now := time.Now().UTC()
	snap := &domain.COASnapshot{
		ID:         uuid.New(),
		TenantID:   tenantID,
		SnapshotAt: now,
		PeriodID:   req.PeriodID,
		Accounts:   entries,
		Hash:       hash,
		EntryCount: len(entries),
		CreatedBy:  req.CreatedBy,
		CreatedAt:  now,
	}

	if err := s.snapshots.Create(ctx, snap); err != nil {
		return nil, fmt.Errorf("create snapshot: persist: %w", err)
	}

	s.emitSnapshotCreated(ctx, snap)

	logger.InfoContext(ctx, "COA snapshot created", logger.Fields{
		"snapshot_id": snap.ID.String(),
		"tenant_id":   tenantID.String(),
		"entry_count": snap.EntryCount,
		"hash":        snap.Hash[:16] + "...",
	})

	return snap, nil
}

// =============================================================================
// Snapshot retrieval
// =============================================================================

// GetByID retrieves a snapshot by its ID.
func (s *COASnapshotService) GetByID(ctx context.Context, id uuid.UUID) (*domain.COASnapshot, error) {
	if s == nil {
		return nil, fmt.Errorf("COASnapshotService is nil")
	}
	return s.snapshots.GetByID(ctx, id)
}

// GetForPeriod retrieves the snapshot bound to a fiscal period.
func (s *COASnapshotService) GetForPeriod(
	ctx context.Context,
	periodID uuid.UUID,
) (*domain.COASnapshot, error) {
	if s == nil {
		return nil, fmt.Errorf("COASnapshotService is nil")
	}
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("get snapshot for period: tenant ID missing from context")
	}
	return s.snapshots.GetForPeriod(ctx, tenantID, periodID)
}

// GetAtTime retrieves the most recent snapshot taken on or before asOfTime.
// Returns nil, nil when no snapshot exists before that time.
func (s *COASnapshotService) GetAtTime(
	ctx context.Context,
	asOfTime time.Time,
) (*domain.COASnapshot, error) {
	if s == nil {
		return nil, fmt.Errorf("COASnapshotService is nil")
	}
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("get snapshot at time: tenant ID missing from context")
	}
	return s.snapshots.GetAtTime(ctx, tenantID, asOfTime)
}

// =============================================================================
// Hash validation (tamper detection)
// =============================================================================

// ValidateSnapshotHash recomputes the hash from the snapshot's stored entries
// and compares it against the stored hash.
//
// Returns (true, nil) when valid, (false, nil) when tampered, (false, err) on error.
func (s *COASnapshotService) ValidateSnapshotHash(
	ctx context.Context,
	snapshotID uuid.UUID,
) (bool, error) {
	if s == nil {
		return false, fmt.Errorf("COASnapshotService is nil")
	}

	snap, err := s.snapshots.GetByID(ctx, snapshotID)
	if err != nil {
		return false, fmt.Errorf("validate snapshot hash: load snapshot: %w", err)
	}

	// Entries must be in canonical order before hashing.
	sorted := make([]domain.COASnapshotEntry, len(snap.Accounts))
	copy(sorted, snap.Accounts)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].AccountCode < sorted[j].AccountCode
	})

	recomputed, err := computeSnapshotHash(sorted)
	if err != nil {
		return false, fmt.Errorf("validate snapshot hash: recompute: %w", err)
	}

	valid := recomputed == snap.Hash

	if !valid {
		logger.WarnContext(ctx, "COA snapshot hash mismatch — possible tampering", logger.Fields{
			"snapshot_id": snapshotID.String(),
			"stored":      snap.Hash,
			"recomputed":  recomputed,
		})
		s.emitHashMismatch(ctx, snapshotID)
	}

	return valid, nil
}

// =============================================================================
// Deterministic hash computation
// =============================================================================

// computeSnapshotHash computes a deterministic SHA-256 of the ordered entries.
// The input slice MUST be sorted in canonical order before calling this.
func computeSnapshotHash(entries []domain.COASnapshotEntry) (string, error) {
	h := sha256.New()
	for _, e := range entries {
		// Marshal each entry to canonical JSON and write to the hash.
		b, err := json.Marshal(e)
		if err != nil {
			return "", fmt.Errorf("marshal entry %s: %w", e.AccountCode, err)
		}
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// =============================================================================
// Metrics
// =============================================================================

func (s *COASnapshotService) emitSnapshotCreated(ctx context.Context, snap *domain.COASnapshot) {
	if s.metrics == nil {
		return
	}
	s.metrics.IncrementCounter("finance_coa_snapshots_created_total", metrics.Fields{
		"tenant_id": snap.TenantID.String(),
	})
}

func (s *COASnapshotService) emitHashMismatch(ctx context.Context, snapshotID uuid.UUID) {
	if s.metrics == nil {
		return
	}
	s.metrics.IncrementCounter("finance_coa_snapshot_hash_mismatches_total", metrics.Fields{
		"snapshot_id": snapshotID.String(),
	})
}
