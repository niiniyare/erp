package tenant

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        int32
	Uuid      uuid.UUID
	Name      string
	Subdomain string
	Status    string
	Industry  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
