// Package pipeline provides composable pipeline stages for finance operations.
// Stages are registered with a StageRegistry and executed by PipelineBuilder.
// See internal/pipeline for the core pipeline engine.
package pipeline

import (
	"time"

	"github.com/google/uuid"
)

// PostTransactionInput is the typed input placed in OperationContext.Input for
// the "finance.transaction.post" operation.
type PostTransactionInput struct {
	TransactionID uuid.UUID
	PostingDate   time.Time // zero value → use time.Now() at post time
	TenantID      uuid.UUID
	PostedBy      uuid.UUID
}

// ── Data-map key constants ────────────────────────────────────────────────────
// Stages communicate through opCtx.Data using these well-known keys.

const (
	// KeyTransaction holds the loaded *domain.Transaction.
	KeyTransaction = "gl.transaction"

	// KeyEntries holds the loaded []domain.TransactionEntry.
	KeyEntries = "gl.entries"

	// KeyPeriod holds the *domain.AccountingPeriod resolved for the posting date.
	// Set by PeriodCheckStage when a PeriodRepository is available.
	KeyPeriod = "gl.period"
)
