package audit

import (
	"context"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

type repository struct {
	store db.Store
}

// NewRepository creates a new audit repository.
func NewRepository(store db.Store) Repository {
	return &repository{
		store: store,
	}
}

// CreateAuditEvent creates a new audit event in the database.
func (r *repository) CreateAuditEvent(ctx context.Context, arg AuditEvent) error {
	params := db.CreateAuditEventParams{
		EventType: arg.EventType,
		Reason:    arg.Reason,
		Context:   arg.Context,
	}

	if arg.UserID != uuid.Nil {
		params.UserID = &arg.UserID
	}
	if arg.EventCategory != "" {
		params.EventCategory = &arg.EventCategory
	}
	if arg.Severity != "" {
		params.Severity = &arg.Severity
	}
	if arg.EntityID.Valid {
		params.EntityID = &arg.EntityID.UUID
	}
	if arg.Decision != "" {
		params.Decision = &arg.Decision
	}

	_, err := r.store.CreateAuditEvent(ctx, params)
	return err
}
