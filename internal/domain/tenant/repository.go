package tenant

import "context"

type Repo interface {
	// WithTx(context.Context, func(*sql.Tx) error) error
	Create(context.Context, *Tenant) error
	UpdateConfig(context.Context, int, map[string]interface{}) error
}
