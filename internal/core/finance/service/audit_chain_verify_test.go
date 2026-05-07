// Package service_test — AuditChainVerifier unit tests.
//
// Proves that:
// - HASH_MISMATCH is detected when a payload is tampered after chain entry is written.
// - SEQUENCE_GAP is detected when an entry is missing from the chain.
// - A clean chain passes verification with zero violations.
// No database required — uses MockAuditChainRepository.
package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"

	"awo.so/internal/core/finance/service"
	"awo.so/internal/shared"
	"awo.so/internal/shared/metrics"
)

// computeTestHash mirrors the production formula so tests can build valid entries.
func computeTestHash(prevHash, eventType string, payload []byte, createdAt time.Time) string {
	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write([]byte(eventType))
	h.Write(payload)
	h.Write([]byte(createdAt.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(h.Sum(nil))
}

// buildEntry creates a properly-hashed AuditChainEntry for a given sequence position.
func buildEntry(seq int64, prevHash string, tenantID uuid.UUID) *service.AuditChainEntry {
	payload, _ := json.Marshal(map[string]string{"event": "TEST", "seq": string(rune('0' + seq))})
	createdAt := time.Date(2026, 1, 1, 0, 0, int(seq), 0, time.UTC)
	hash := computeTestHash(prevHash, "TEST_EVENT", payload, createdAt)
	return &service.AuditChainEntry{
		ID:        uuid.New(),
		TenantID:  tenantID,
		OutboxID:  uuid.New(),
		Sequence:  seq,
		EventType: "TEST_EVENT",
		Payload:   payload,
		PrevHash:  prevHash,
		ChainHash: hash,
		CreatedAt: createdAt,
	}
}

// ============================================================================
// FIN-CHAIN-001: Hash mismatch detected when payload is tampered
// ============================================================================

func TestAuditChainVerifier_HashMismatch_Detected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	// Build a valid entry, then tamper with its payload.
	entry := buildEntry(1, "", tenantID)
	entry.Payload = []byte(`{"tampered":"yes"}`) // hash no longer matches

	mockRepo := service.NewMockAuditChainRepository(ctrl)
	mockRepo.EXPECT().GetChainStats(gomock.Any(), tenantID).Return(int64(1), int64(1), nil)
	mockRepo.EXPECT().ListChainRange(gomock.Any(), tenantID, int64(1), int64(1)).Return([]*service.AuditChainEntry{entry}, nil)

	verifier := service.NewAuditChainVerifier(mockRepo, metrics.NewNoOpMetricsProvider())
	report, err := verifier.VerifyChain(ctx, 1, 1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == "HASH_MISMATCH" && v.Sequence == 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected HASH_MISMATCH violation at seq=1, violations: %+v", report.Violations)
	}
	if report.Healthy {
		t.Error("report should not be healthy when hash mismatch detected")
	}
}

// ============================================================================
// FIN-CHAIN-002: Sequence gap detected when an entry is missing
// ============================================================================

func TestAuditChainVerifier_SequenceGap_Detected(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	// Build entries seq=1 and seq=3 — seq=2 is missing.
	e1 := buildEntry(1, "", tenantID)
	e3 := buildEntry(3, e1.ChainHash, tenantID)

	mockRepo := service.NewMockAuditChainRepository(ctrl)
	mockRepo.EXPECT().GetChainStats(gomock.Any(), tenantID).Return(int64(3), int64(2), nil)
	mockRepo.EXPECT().ListChainRange(gomock.Any(), tenantID, int64(1), int64(3)).Return([]*service.AuditChainEntry{e1, e3}, nil)

	verifier := service.NewAuditChainVerifier(mockRepo, metrics.NewNoOpMetricsProvider())
	report, err := verifier.VerifyChain(ctx, 1, 3, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, v := range report.Violations {
		if v.Kind == "SEQUENCE_GAP" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected SEQUENCE_GAP violation, violations: %+v", report.Violations)
	}
	if report.Healthy {
		t.Error("report should not be healthy when sequence gap detected")
	}
}

// ============================================================================
// FIN-CHAIN-003: Valid chain passes verification with zero violations
// ============================================================================

func TestAuditChainVerifier_ValidChain_NoViolations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	// Build a 3-entry chain with correct hashes.
	e1 := buildEntry(1, "", tenantID)
	e2 := buildEntry(2, e1.ChainHash, tenantID)
	e3 := buildEntry(3, e2.ChainHash, tenantID)

	mockRepo := service.NewMockAuditChainRepository(ctrl)
	mockRepo.EXPECT().GetChainStats(gomock.Any(), tenantID).Return(int64(3), int64(3), nil)
	mockRepo.EXPECT().ListChainRange(gomock.Any(), tenantID, int64(1), int64(3)).Return([]*service.AuditChainEntry{e1, e2, e3}, nil)

	verifier := service.NewAuditChainVerifier(mockRepo, metrics.NewNoOpMetricsProvider())
	report, err := verifier.VerifyChain(ctx, 1, 3, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !report.Healthy {
		t.Errorf("expected healthy report for valid chain, violations: %+v", report.Violations)
	}
	if len(report.Violations) > 0 {
		t.Errorf("expected zero violations, got: %+v", report.Violations)
	}
}

// ============================================================================
// FIN-CHAIN-004: Nil verifier returns healthy empty report without panicking
// ============================================================================

func TestAuditChainVerifier_Nil_ReturnsHealthy(t *testing.T) {
	tenantID := uuid.New()
	ctx := shared.WithTenantID(context.Background(), tenantID)

	var verifier *service.AuditChainVerifier
	report, err := verifier.VerifyChain(ctx, 1, 10, 100)
	if err != nil {
		t.Fatalf("nil verifier should not return error: %v", err)
	}
	if !report.Healthy {
		t.Error("nil verifier should return healthy report")
	}
}
