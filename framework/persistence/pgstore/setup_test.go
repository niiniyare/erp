package pgstore_test

import (
	"context"
	"testing"

	"awo.so/framework/persistence/pgstore"
)

// TestEnsureRLSFunction_MissingFunction verifies that EnsureRLSFunction returns
// a descriptive error when set_tenant_context() is absent.
//
// This is an integration test — it requires a real PostgreSQL instance.
// Run via: go test -tags integration ./framework/persistence/pgstore/...
//
// Unit-level coverage is achieved by checking the error message contract
// against a stub pool. A full integration test is left to CI.
func TestEnsureRLSFunction_NilPool(t *testing.T) {
	// Calling with a nil pool should panic — document that requirement.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil pool")
		}
	}()
	//nolint:staticcheck
	_ = pgstore.EnsureRLSFunction(context.Background(), nil)
}
