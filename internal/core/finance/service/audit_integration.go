package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"awo.so/internal/core/audit"
	"awo.so/internal/shared"
	"awo.so/internal/shared/logger"
)

// Finance audit event types — used as EventType in audit.CreateAuditEventRequest.
const (
	auditTypeTxnCreated     = "FINANCE_TXN_CREATED"
	auditTypeTxnPosted      = "FINANCE_TXN_POSTED"
	auditTypeTxnApproved    = "FINANCE_TXN_APPROVED"
	auditTypeTxnRejected    = "FINANCE_TXN_REJECTED"
	auditTypeTxnReversed    = "FINANCE_TXN_REVERSED"
	auditTypeRecurringCreated = "FINANCE_RECURRING_TXN_CREATED"
	auditTypePeriodStatus   = "FINANCE_PERIOD_STATUS_CHANGED"
	auditTypeStmtImported   = "FINANCE_BANK_STMT_IMPORTED"
	auditTypeReconciled     = "FINANCE_RECONCILIATION_COMPLETED"
	auditTypeLineMatched    = "FINANCE_STMT_LINE_MATCHED"
	auditTypeLineUnmatched  = "FINANCE_STMT_LINE_UNMATCHED"
)

const (
	auditCategoryFinance = "FINANCE"
	auditSeverityInfo    = "INFO"
	auditSeverityMedium  = "MEDIUM"
	auditSeverityHigh    = "HIGH"
)

// fireAudit writes a finance audit event synchronously.
// Errors are logged at ERROR level but never propagated — audit failures
// must not fail financial mutations.
func fireAudit(ctx context.Context, svc audit.Service, req audit.CreateAuditEventRequest) {
	if svc == nil {
		return
	}
	if _, err := svc.CreateAuditEvent(ctx, req); err != nil {
		logger.ErrorContext(ctx, "finance audit write failed", logger.Fields{
			"event_type": req.EventType,
			"error":      err.Error(),
		})
	}
}

// auditUserID extracts the caller's user ID for audit events.
// Returns nil when no authenticated user is in context (system calls).
func auditUserID(ctx context.Context) *uuid.UUID {
	id, ok := shared.GetUserID(ctx)
	if !ok || id == uuid.Nil {
		return nil
	}
	return &id
}

// auditCtx marshals arbitrary fields into JSON for the audit Context field.
// Returns nil on marshal error — audit still fires, just without context payload.
func auditCtx(fields map[string]any) json.RawMessage {
	b, err := json.Marshal(fields)
	if err != nil {
		return nil
	}
	return b
}

// uuidPtr returns a pointer to v. Returns nil for uuid.Nil.
func uuidPtr(v uuid.UUID) *uuid.UUID {
	if v == uuid.Nil {
		return nil
	}
	return &v
}

// derefUUIDPtr dereferences a *uuid.UUID for uuidPtr — safe for nil (returns nil).
func derefUUIDPtr(p *uuid.UUID) *uuid.UUID {
	if p == nil || *p == uuid.Nil {
		return nil
	}
	return p
}
