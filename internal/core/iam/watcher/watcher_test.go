package watcher_test

// watcher_test.go — AUTHZ-DIST-1 through AUTHZ-DIST-5.
//
// Tests the PolicyWatcher contract using a chanWatcher test double.
// No database or network required.
//
// AUTHZ-DIST-1: Notify triggers Watch callback on the same watcher.
// AUTHZ-DIST-2: Multiple Watch subscribers each receive the notification.
// AUTHZ-DIST-3: Watch stops delivering after context cancel.
// AUTHZ-DIST-4: Notify with no active Watch subscriber does not block or panic.
// AUTHZ-DIST-5: Watcher error on Notify is surfaced as a non-nil error.

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"awo.so/internal/core/iam/watcher"
)

// ---------------------------------------------------------------------------
// chanWatcher — test double for PolicyWatcher
// ---------------------------------------------------------------------------
//
// chanWatcher simulates the watcher mechanism without any network or DB:
//   - Notify sends a signal to all registered subscriber channels.
//   - Watch reads from a subscriber channel and calls fn in a goroutine.
//   - Close is a no-op; Watch goroutines exit when the ctx is cancelled.

type chanWatcher struct {
	mu   sync.Mutex
	subs []chan struct{}
}

func (w *chanWatcher) Notify(_ context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ch := range w.subs {
		select {
		case ch <- struct{}{}:
		default: // subscriber is busy; drop the signal (mirrors real pub/sub)
		}
	}
	return nil
}

func (w *chanWatcher) Watch(ctx context.Context, fn func()) error {
	ch := make(chan struct{}, 1)
	w.mu.Lock()
	w.subs = append(w.subs, ch)
	w.mu.Unlock()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				fn()
			}
		}
	}()
	return nil
}

func (w *chanWatcher) Close() error { return nil }

// subscribe returns the number of current subscribers (for assertions).
func (w *chanWatcher) subscriberCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.subs)
}

// ---------------------------------------------------------------------------
// errWatcher — always returns an error from Notify (AUTHZ-DIST-5)
// ---------------------------------------------------------------------------

var errNotifyFailed = errors.New("simulated notify failure")

type errWatcher struct{ watcher.NoopWatcher }

func (errWatcher) Notify(_ context.Context) error { return errNotifyFailed }

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// awaitCall returns true if done is closed within timeout.
func awaitCall(done <-chan struct{}, timeout time.Duration) bool {
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// ---------------------------------------------------------------------------
// AUTHZ-DIST-1: Notify triggers Watch callback
// ---------------------------------------------------------------------------

func TestAUTHZ_DIST_1_NotifyTriggers_WatchCallback(t *testing.T) {
	w := &chanWatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := make(chan struct{}, 1)
	require.NoError(t, w.Watch(ctx, func() { called <- struct{}{} }))

	require.NoError(t, w.Notify(context.Background()))

	assert.True(t, awaitCall(called, 200*time.Millisecond),
		"Watch callback must fire after Notify")
}

// ---------------------------------------------------------------------------
// AUTHZ-DIST-2: Multiple subscribers each receive the notification
// ---------------------------------------------------------------------------

func TestAUTHZ_DIST_2_MultipleSubscribers_AllReceive(t *testing.T) {
	w := &chanWatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const n = 3
	var count atomic.Int32
	done := make(chan struct{}, n)

	for range n {
		require.NoError(t, w.Watch(ctx, func() {
			count.Add(1)
			done <- struct{}{}
		}))
	}

	assert.Equal(t, n, w.subscriberCount())

	require.NoError(t, w.Notify(context.Background()))

	received := 0
	deadline := time.After(300 * time.Millisecond)
	for received < n {
		select {
		case <-done:
			received++
		case <-deadline:
			t.Fatalf("only %d/%d subscribers received notification", received, n)
		}
	}
	assert.Equal(t, int32(n), count.Load())
}

// ---------------------------------------------------------------------------
// AUTHZ-DIST-3: Watch stops delivering after context cancel
// ---------------------------------------------------------------------------

func TestAUTHZ_DIST_3_WatchStops_AfterContextCancel(t *testing.T) {
	w := &chanWatcher{}
	ctx, cancel := context.WithCancel(context.Background())

	var count atomic.Int32
	require.NoError(t, w.Watch(ctx, func() { count.Add(1) }))

	// First notify — must fire.
	require.NoError(t, w.Notify(context.Background()))
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), count.Load(), "callback must fire before cancel")

	// Cancel the Watch context.
	cancel()
	time.Sleep(50 * time.Millisecond) // let goroutine exit

	// Second notify — must NOT fire.
	require.NoError(t, w.Notify(context.Background()))
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(1), count.Load(), "callback must not fire after cancel")
}

// ---------------------------------------------------------------------------
// AUTHZ-DIST-4: Notify with no subscribers does not block or panic
// ---------------------------------------------------------------------------

func TestAUTHZ_DIST_4_NotifyWithNoSubscribers_IsNoop(t *testing.T) {
	w := &chanWatcher{}
	assert.Equal(t, 0, w.subscriberCount())

	// Must complete without blocking, panicking, or returning an error.
	done := make(chan error, 1)
	go func() { done <- w.Notify(context.Background()) }()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Notify with no subscribers blocked for > 100ms")
	}
}

// ---------------------------------------------------------------------------
// AUTHZ-DIST-5: Watcher error on Notify is surfaced as a non-nil error
// ---------------------------------------------------------------------------

func TestAUTHZ_DIST_5_NotifyError_IsSurfaced(t *testing.T) {
	w := errWatcher{}
	err := w.Notify(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, errNotifyFailed,
		"Notify error must be returned to the caller so it can be logged / metered")
}

// ---------------------------------------------------------------------------
// NoopWatcher contract tests
// ---------------------------------------------------------------------------

func TestNoopWatcher_AllMethodsAreNoops(t *testing.T) {
	var w watcher.NoopWatcher
	ctx := context.Background()

	assert.NoError(t, w.Notify(ctx))
	assert.NoError(t, w.Watch(ctx, func() { t.Fatal("NoopWatcher callback must never fire") }))
	assert.NoError(t, w.Notify(ctx)) // even after Watch — must not call fn
	assert.NoError(t, w.Close())
	assert.NoError(t, w.Close()) // idempotent
}

// ---------------------------------------------------------------------------
// WatcherError formatting
// ---------------------------------------------------------------------------

func TestWatcherError_FormatsCorrectly(t *testing.T) {
	inner := errors.New("connection refused")
	err := &watcher.WatcherError{Op: "Notify", Err: inner}

	assert.Contains(t, err.Error(), "Notify")
	assert.Contains(t, err.Error(), "connection refused")
	assert.ErrorIs(t, err, inner, "WatcherError must unwrap to inner error")
}
