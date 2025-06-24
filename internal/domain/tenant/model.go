package tenant

import "time"

type Tenant struct {
	ID        int32
	Uuid      string
	Name      string
	Subdomain string
	Status    string
	Industry  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
