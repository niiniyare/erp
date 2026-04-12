package pipeline

import (
	"context"

	db "awo.so/db/sqlc"
)

// TxHook is a function that executes inside the same database transaction as
// the operation's primary write. It receives a tenant-scoped Store and the
// current OperationContext.
//
// If a TxHook returns an error the entire transaction — including the primary
// write — is rolled back. This implements the Transactional Outbox Pattern:
// side-effect rows (domain_events, workflow_trigger_queue, budget_actuals) are
// written atomically with the business record they depend on.
//
// TxHooks are registered by stages during their Execute() call, not at startup.
// They fire after all stages complete but before the transaction commits.
type TxHook struct {
	// Name uniquely identifies this hook for logging and debugging.
	Name string

	// Priority controls execution order within the same transaction.
	// Lower values execute first.
	Priority int

	// Fn is the hook body. store is a tenant-scoped Store bound to the open,
	// uncommitted transaction — tenant context is already set via
	// set_tenant_context() so all queries are RLS-filtered.
	Fn func(ctx context.Context, store db.Store, opCtx *OperationContext) error
}
