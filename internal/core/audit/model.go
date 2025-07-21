package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// AuditEvent represents the data for an audit log entry.
type AuditEvent struct {
	UserID        uuid.UUID
	EventType     string
	EventCategory string
	Severity      string
	EntityID      uuid.NullUUID
	Decision      string
	Reason        string
	Context       json.RawMessage
}

// Service defines the interface for the audit service.
type Service interface {
	Record(ctx context.Context, event AuditEvent) error
}

// Repository defines the interface for the audit repository.
type Repository interface {
	CreateAuditEvent(ctx context.Context, arg AuditEvent) error
}
