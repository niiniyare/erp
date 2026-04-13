package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5"

	db "awo.so/db/sqlc"
)

// txFrom extracts the underlying pgx.Tx from a db.Store that has been wrapped
// inside a WithTenant / WithTenantFromCtx callback. The store passed to the
// callback is always a db.TxStore; if it is not, the call site has a bug.
func txFrom(s db.Store) (pgx.Tx, error) {
	ts, ok := s.(db.TxStore)
	if !ok {
		return nil, fmt.Errorf("finance repository: store is not a TxStore — must be called inside a WithTenant callback")
	}
	return ts.GetTx(), nil
}
