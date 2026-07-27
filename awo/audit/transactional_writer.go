package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"awo.so/awo/tx"
)

// TransactionalWriter is the production AuditWriter implementation.
// It writes AuditRecord values to platform_audit_log using the connection
// (transaction or pool) carried in context.
//
// When called from within an EntityService repo.WithTx callback, the record
// is written inside the active transaction — guaranteeing that no mutation
// escapes the audit trail. If the Write call itself fails and the failure
// policy is Propagate, the transaction rolls back together with the mutation.
//
// When called outside a transaction (e.g. standalone auth events), the writer
// falls back to the pool for a direct INSERT.
//
// TransactionalWriter never imports contrib/pgx — it accesses the driver
// connection exclusively via tx.QuerierFromContext to avoid circular imports.
type TransactionalWriter struct {
	pool *pgxpool.Pool
}

// NewTransactionalWriter returns a TransactionalWriter backed by pool.
// pool is used for auth/security events that are written outside a transaction.
func NewTransactionalWriter(pool *pgxpool.Pool) *TransactionalWriter {
	return &TransactionalWriter{pool: pool}
}

// Write inserts record into platform_audit_log. It uses the active transaction
// from ctx when available, falling back to the pool otherwise.
//
// Write validates the record before attempting the insert. Validation errors
// are always returned to the caller regardless of the failure policy (the
// policy governs DB errors, not programming errors).
func (w *TransactionalWriter) Write(ctx context.Context, record AuditRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	// Assign ID and timestamp if not already set by caller.
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}

	beforeJSON, err := marshalNullable(record.BeforeData)
	if err != nil {
		return errorf("audit: marshal BeforeData: %w", err)
	}
	afterJSON, err := marshalNullable(record.AfterData)
	if err != nil {
		return errorf("audit: marshal AfterData: %w", err)
	}
	contextJSON, err := marshalNullable(record.Context)
	if err != nil {
		return errorf("audit: marshal Context: %w", err)
	}
	complianceJSON, err := marshalNullable(record.ComplianceFlags)
	if err != nil {
		return errorf("audit: marshal ComplianceFlags: %w", err)
	}

	changedFieldsJSON, err := marshalNullable(record.ChangedFields)
	if err != nil {
		return errorf("audit: marshal ChangedFields: %w", err)
	}

	actorID := record.ActorID()
	saID := record.ServiceAccountID()

	var actorIDArg, saIDArg interface{}
	if actorID != uuid.Nil {
		actorIDArg = actorID
	}
	if saID != uuid.Nil {
		saIDArg = saID
	}

	var systemActorArg interface{}
	if record.SystemActor != "" {
		systemActorArg = string(record.SystemActor)
	}

	const insertSQL = `
INSERT INTO platform_audit_log (
	id, tenant_id, request_id,
	entity_name, record_id, operation,
	actor_id, service_account_id, system_actor,
	ip_address, session_id,
	before_data, after_data, changed_fields,
	event_category, severity, risk_score, compliance_flags,
	context, created_at
) VALUES (
	$1, $2, $3,
	$4, $5, $6,
	$7, $8, $9,
	$10, $11,
	$12, $13, $14,
	$15, $16, $17, $18,
	$19, $20
)`

	args := []any{
		record.ID, record.TenantID, nullableString(record.RequestID),
		record.EntityName, nullableUUID(record.RecordID), string(record.Operation),
		actorIDArg, saIDArg, systemActorArg,
		nullableString(record.IPAddress), nullableString(record.SessionID),
		beforeJSON, afterJSON, changedFieldsJSON,
		string(record.EventCategory), string(record.Severity), record.RiskScore, complianceJSON,
		contextJSON, record.CreatedAt,
	}

	// Prefer the active transaction from context; fall back to pool.
	if q, ok := tx.QuerierFromContext(ctx); ok {
		_, err = q.ExecSQL(ctx, insertSQL, args...)
	} else {
		_, err = w.pool.Exec(ctx, insertSQL, args...)
	}
	return err
}

// marshalNullable marshals v to JSON bytes. Returns nil when v is nil or
// empty (nil map, nil slice) so that the DB column stores NULL rather than
// an empty JSON object/array.
func marshalNullable(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// Treat empty JSON objects and arrays as NULL.
	s := string(b)
	if s == "{}" || s == "[]" || s == "null" {
		return nil, nil
	}
	return b, nil
}

// nullableString returns nil when s is empty, preserving SQL NULL semantics.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullableUUID returns nil when u is the zero UUID, preserving SQL NULL.
func nullableUUID(u uuid.UUID) interface{} {
	if u == uuid.Nil {
		return nil
	}
	return u
}
