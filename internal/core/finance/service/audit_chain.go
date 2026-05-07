package service

// audit_chain.go — Tamper-evident audit chain for CRITICAL finance events.
//
// Every delivered CRITICAL outbox entry is appended to a hash chain stored in
// finance_audit_chain. Each entry includes:
//   - A monotonically increasing sequence number (per tenant)
//   - SHA256(prevHash || eventType || payload || createdAt as RFC3339)
//
// AuditChainVerifier walks the chain and detects:
//   - Hash mismatches (payload tampered after delivery)
//   - Sequence gaps (entries deleted from the chain table)
//   - Broken prev-hash linkage (rows reordered or inserted)
//
// This is NOT a blockchain. It is a forensic tamper-evidence mechanism using
// a standard hash chain backed by a single database table.
//
// All methods are nil-safe.

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// AuditChainEntry is a single link in the CRITICAL-event hash chain for a tenant.
type AuditChainEntry struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	OutboxID   uuid.UUID `json:"outbox_id"`   // FK to finance_audit_outbox
	Sequence   int64     `json:"sequence"`    // Monotonically increasing per tenant
	EventType  string    `json:"event_type"`
	Payload    []byte    `json:"payload"`     // JSON-encoded audit.CreateAuditEventRequest
	PrevHash   string    `json:"prev_hash"`   // ChainHash of the previous entry; "" for first
	ChainHash  string    `json:"chain_hash"`  // SHA256(prevHash+eventType+payload+createdAt)
	CreatedAt  time.Time `json:"created_at"`
}

// computeChainHash calculates the deterministic chain hash for an entry.
func computeChainHash(prevHash, eventType string, payload []byte, createdAt time.Time) string {
	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write([]byte(eventType))
	h.Write(payload)
	h.Write([]byte(createdAt.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(h.Sum(nil))
}

// AuditChainRepository persists and queries the CRITICAL-event hash chain.
type AuditChainRepository interface {
	// AppendChainEntry appends a new link. Sequence must be tenant-monotonic.
	// Idempotent by OutboxID: silently succeeds if the outbox_id already exists.
	AppendChainEntry(ctx context.Context, entry *AuditChainEntry) error

	// GetLastChainEntry returns the most recent entry for the tenant, or nil if
	// the chain is empty (first event).
	GetLastChainEntry(ctx context.Context, tenantID uuid.UUID) (*AuditChainEntry, error)

	// ListChainRange returns entries in ascending sequence order for the given range.
	ListChainRange(ctx context.Context, tenantID uuid.UUID, fromSeq, toSeq int64) ([]*AuditChainEntry, error)

	// GetChainStats returns the max sequence and entry count for a tenant.
	GetChainStats(ctx context.Context, tenantID uuid.UUID) (maxSeq int64, count int64, err error)
}

// ── AuditChainWriter ──────────────────────────────────────────────────────────

// AuditChainWriter appends a delivered CRITICAL outbox entry to the hash chain.
// Called by AuditOutboxProcessor after each successful delivery.
// A nil *AuditChainWriter silently no-ops.
type AuditChainWriter struct {
	chainRepo AuditChainRepository
	metrics   metrics.MetricsProvider
}

// NewAuditChainWriter creates the writer.
func NewAuditChainWriter(chainRepo AuditChainRepository, m metrics.MetricsProvider) *AuditChainWriter {
	return &AuditChainWriter{chainRepo: chainRepo, metrics: m}
}

// AppendDelivered appends a delivered outbox entry to the tenant's hash chain.
// On failure the chain write is logged but does NOT fail the delivery — the
// outbox delivery is already marked DELIVERED. The gap detector will surface
// the missing chain entry.
func (w *AuditChainWriter) AppendDelivered(ctx context.Context, outboxID uuid.UUID, eventType string, payload json.RawMessage, deliveredAt time.Time) {
	if w == nil || w.chainRepo == nil {
		return
	}
	tenantID, _ := shared.GetTenantID(ctx)

	last, err := w.chainRepo.GetLastChainEntry(ctx, tenantID)
	if err != nil {
		logger.ErrorContext(ctx, "audit chain: failed to get last entry — chain append skipped", logger.Fields{
			"outbox_id":  outboxID.String(),
			"event_type": eventType,
			"error":      err.Error(),
		})
		w.metrics.IncrementCounter("finance_audit_chain_write_failures_total", metrics.Fields{
			"event_type": eventType,
		})
		return
	}

	prevHash := ""
	nextSeq := int64(1)
	if last != nil {
		prevHash = last.ChainHash
		nextSeq = last.Sequence + 1
	}

	payloadBytes := []byte(payload)
	chainHash := computeChainHash(prevHash, eventType, payloadBytes, deliveredAt)

	entry := &AuditChainEntry{
		ID:        uuid.New(),
		TenantID:  tenantID,
		OutboxID:  outboxID,
		Sequence:  nextSeq,
		EventType: eventType,
		Payload:   payloadBytes,
		PrevHash:  prevHash,
		ChainHash: chainHash,
		CreatedAt: deliveredAt,
	}

	if appendErr := w.chainRepo.AppendChainEntry(ctx, entry); appendErr != nil {
		logger.ErrorContext(ctx, "audit chain: append failed — forensic chain has a gap", logger.Fields{
			"outbox_id":  outboxID.String(),
			"event_type": eventType,
			"sequence":   nextSeq,
			"error":      appendErr.Error(),
		})
		w.metrics.IncrementCounter("finance_audit_chain_write_failures_total", metrics.Fields{
			"event_type": eventType,
		})
		return
	}
	w.metrics.IncrementCounter("finance_audit_chain_appended_total", metrics.Fields{
		"event_type": eventType,
	})
}

// ── AuditChainVerifier ────────────────────────────────────────────────────────

// ChainVerificationReport summarises the result of a hash-chain verification scan.
type ChainVerificationReport struct {
	TenantID      uuid.UUID         `json:"tenant_id"`
	TotalEntries  int64             `json:"total_entries"`
	Violations    []ChainViolation  `json:"violations,omitempty"`
	Healthy       bool              `json:"healthy"`
	VerifiedAt    time.Time         `json:"verified_at"`
}

// ChainViolation describes a single detected tamper or gap in the audit chain.
type ChainViolation struct {
	Kind     string `json:"kind"`
	Sequence int64  `json:"sequence"`
	Detail   string `json:"detail"`
}

// AuditChainVerifier reads and verifies the CRITICAL-event hash chain for a tenant.
// A nil *AuditChainVerifier returns a healthy empty report.
type AuditChainVerifier struct {
	chainRepo AuditChainRepository
	metrics   metrics.MetricsProvider
}

// NewAuditChainVerifier creates the verifier.
func NewAuditChainVerifier(chainRepo AuditChainRepository, m metrics.MetricsProvider) *AuditChainVerifier {
	return &AuditChainVerifier{chainRepo: chainRepo, metrics: m}
}

// VerifyChain walks the tenant's hash chain from fromSeq to toSeq and verifies:
//  1. No sequence gaps.
//  2. Each entry's ChainHash matches the recomputed value.
//  3. Each entry's PrevHash matches the previous entry's ChainHash.
//
// Paged in blocks of pageSize to avoid OOM on large chains.
func (v *AuditChainVerifier) VerifyChain(ctx context.Context, fromSeq, toSeq int64, pageSize int) (*ChainVerificationReport, error) {
	if v == nil || v.chainRepo == nil {
		tenantID, _ := shared.GetTenantID(ctx)
		return &ChainVerificationReport{TenantID: tenantID, Healthy: true, VerifiedAt: time.Now()}, nil
	}
	tenantID, _ := shared.GetTenantID(ctx)
	report := &ChainVerificationReport{
		TenantID:   tenantID,
		VerifiedAt: time.Now(),
	}

	_, totalCount, statsErr := v.chainRepo.GetChainStats(ctx, tenantID)
	if statsErr != nil {
		return nil, fmt.Errorf("audit chain: get stats: %w", statsErr)
	}
	report.TotalEntries = totalCount

	var prevHash string
	expectedSeq := fromSeq

	for start := fromSeq; start <= toSeq; start += int64(pageSize) {
		end := start + int64(pageSize) - 1
		if end > toSeq {
			end = toSeq
		}
		entries, err := v.chainRepo.ListChainRange(ctx, tenantID, start, end)
		if err != nil {
			return nil, fmt.Errorf("audit chain: list range [%d,%d]: %w", start, end, err)
		}

		for _, entry := range entries {
			// Check sequence gap.
			if entry.Sequence != expectedSeq {
				report.Violations = append(report.Violations, ChainViolation{
					Kind:     "SEQUENCE_GAP",
					Sequence: expectedSeq,
					Detail:   fmt.Sprintf("expected sequence %d, got %d — entries may have been deleted", expectedSeq, entry.Sequence),
				})
				v.metrics.IncrementCounter("finance_audit_chain_violations_total", metrics.Fields{
					"kind": "SEQUENCE_GAP",
				})
				// Advance to actual sequence so we continue checking from here.
				expectedSeq = entry.Sequence
			}

			// Check prev-hash linkage.
			if entry.Sequence > fromSeq && entry.PrevHash != prevHash {
				report.Violations = append(report.Violations, ChainViolation{
					Kind:     "PREV_HASH_MISMATCH",
					Sequence: entry.Sequence,
					Detail:   fmt.Sprintf("seq %d: prev_hash %q does not match previous entry's chain_hash %q", entry.Sequence, entry.PrevHash, prevHash),
				})
				v.metrics.IncrementCounter("finance_audit_chain_violations_total", metrics.Fields{
					"kind": "PREV_HASH_MISMATCH",
				})
			}

			// Recompute hash and verify.
			recomputed := computeChainHash(entry.PrevHash, entry.EventType, entry.Payload, entry.CreatedAt)
			if recomputed != entry.ChainHash {
				report.Violations = append(report.Violations, ChainViolation{
					Kind:     "HASH_MISMATCH",
					Sequence: entry.Sequence,
					Detail:   fmt.Sprintf("seq %d: stored hash %q != recomputed %q — payload may have been tampered", entry.Sequence, entry.ChainHash, recomputed),
				})
				v.metrics.IncrementCounter("finance_audit_chain_violations_total", metrics.Fields{
					"kind": "HASH_MISMATCH",
				})
			}

			prevHash = entry.ChainHash
			expectedSeq = entry.Sequence + 1
		}
	}

	report.Healthy = len(report.Violations) == 0
	if !report.Healthy {
		logger.ErrorContext(ctx, "audit chain: verification FAILED — tamper or deletion detected", logger.Fields{
			"tenant_id":       tenantID.String(),
			"violations":      len(report.Violations),
			"from_seq":        fromSeq,
			"to_seq":          toSeq,
		})
	}
	return report, nil
}
