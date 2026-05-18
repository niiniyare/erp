// Package watcher provides the PolicyWatcher interface and implementations for
// reactive Casbin policy synchronisation across multiple application instances.
//
// Problem:
//
//	Each application instance holds an independent in-memory Casbin enforcer.
//	Without coordination, a policy or role mutation on instance A is invisible
//	to instance B until the next periodic reload (currently 30 seconds).
//
// Solution:
//
//	PolicyWatcher wraps a signalling channel (PostgreSQL LISTEN/NOTIFY or an
//	equivalent bus). After every successful mutation, the authz service calls
//	Notify. Every listening instance receives the signal and calls
//	enforcer.LoadPolicy() immediately, converging to the new state within the
//	network round-trip time rather than within the next 30-second window.
//
//	The 30-second StartAutoLoadPolicy remains active as a safety-net catch-all
//	for cases where a notification is missed (e.g. a transient outage).
//
// Wiring:
//
//	cfg := iam.Config{
//	    // ...
//	    Watcher: watcher.NewPgWatcher(pool, log, metrics),
//	}
//	svc, _ := iam.New(cfg)
//	// On graceful shutdown:
//	cfg.Watcher.Close()
package watcher

import (
	"context"
	"fmt"
)

// PolicyWatcher signals peer IAM nodes to reload their Casbin policy state.
//
// Contract:
//   - Notify is called by the authz service after every successful mutation
//     (AddPolicy, RemovePolicy, AssignRole, RevokeRole, BootstrapTenantAdmin).
//   - Watch registers a reload callback. It must be called before mutations are
//     expected. The callback is invoked from a background goroutine.
//   - All methods are safe for concurrent use.
//   - A nil PolicyWatcher is NOT safe; use NoopWatcher for no-op behaviour.
type PolicyWatcher interface {
	// Notify signals all peer nodes to reload their Casbin policy state.
	// Best-effort: the authz service logs errors but does NOT fail the mutation.
	Notify(ctx context.Context) error

	// Watch starts a background listener that calls fn each time a reload
	// signal arrives. The ctx controls the listener lifetime — cancel it or
	// call Close to stop. Watch returns immediately after registering the
	// listener; fn is called asynchronously.
	Watch(ctx context.Context, fn func()) error

	// Close stops the background listener and releases resources.
	// Safe to call multiple times. Close blocks until the listener exits.
	Close() error
}

// NoopWatcher is a PolicyWatcher that does nothing.
// Use in single-instance deployments or tests that do not need cross-node
// propagation — it is cheaper than nil-checking at every call site.
type NoopWatcher struct{}

func (NoopWatcher) Notify(_ context.Context) error         { return nil }
func (NoopWatcher) Watch(_ context.Context, _ func()) error { return nil }
func (NoopWatcher) Close() error                           { return nil }

// WatcherError wraps an error from a watcher operation with the operation name.
type WatcherError struct {
	Op  string
	Err error
}

func (e *WatcherError) Error() string { return fmt.Sprintf("iam.watcher.%s: %v", e.Op, e.Err) }
func (e *WatcherError) Unwrap() error { return e.Err }
