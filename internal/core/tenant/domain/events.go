package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event is the base for all tenant domain events.
type Event struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	Timestamp time.Time `json:"timestamp"`
}

type TenantCreated struct {
	Event
	Name string `json:"name"`
}

type TenantActivated struct {
	Event
}

type TenantSuspended struct {
	Event
	Reason string `json:"reason"`
}

type TenantArchived struct {
	Event
}

type TenantUpdated struct {
	Event
	Fields []string `json:"fields"`
}

type TenantDeleted struct {
	Event
}
