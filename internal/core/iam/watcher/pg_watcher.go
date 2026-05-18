package watcher

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
)

// pgChannel is the PostgreSQL NOTIFY channel used for policy reload signals.
// All instances LISTEN on this channel; any mutation triggers a NOTIFY.
const pgChannel = "iam_policy_changed"

// PgWatcher implements PolicyWatcher using PostgreSQL LISTEN/NOTIFY.
//
// Convergence window:
//
//	NOTIFY is delivered to all listening connections on the same PostgreSQL
//	cluster within the transaction commit round-trip (typically < 10 ms on a
//	local network). This is orders of magnitude faster than the 30-second
//	auto-reload safety net, which remains active.
//
// Dedicated connection:
//
//	LISTEN requires a persistent dedicated connection — it cannot use a pool
//	connection that is checked out transiently. PgWatcher.Watch acquires one
//	connection from the pool and holds it for the lifetime of the listener.
//	This uses one extra connection per application instance; size your pool
//	accordingly (pool_max += number_of_instances).
//
// Missed notifications:
//
//	If the listener connection drops, WaitForNotification returns an error.
//	PgWatcher backs off for 1 second and retries the acquire + LISTEN loop.
//	Any mutations that occurred during the outage are picked up by the next
//	periodic StartAutoLoadPolicy tick (≤ 30 seconds).
type PgWatcher struct {
	pool    *pgxpool.Pool
	log     logger.Logger
	metrics metrics.MetricsProvider
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewPgWatcher creates a PgWatcher. pool must not be nil.
// log and metrics are optional (nil-safe).
func NewPgWatcher(pool *pgxpool.Pool, log logger.Logger, mp metrics.MetricsProvider) *PgWatcher {
	return &PgWatcher{
		pool:    pool,
		log:     log,
		metrics: mp,
		done:    make(chan struct{}),
	}
}

// Notify executes NOTIFY iam_policy_changed on any pool connection.
// The signal reaches all instances currently LISTENING on the same PG cluster.
func (w *PgWatcher) Notify(ctx context.Context) error {
	_, err := w.pool.Exec(ctx, "SELECT pg_notify($1, '')", pgChannel)
	if err != nil {
		w.incrCounter("iam.watcher.notify_error")
		return &WatcherError{Op: "Notify", Err: err}
	}
	w.incrCounter("iam.watcher.notify_sent")
	return nil
}

// Watch acquires a dedicated connection, issues LISTEN iam_policy_changed, and
// starts a background goroutine that calls fn on each notification.
//
// The goroutine exits when ctx is cancelled or Close is called.
// If the connection drops, the goroutine re-acquires a connection and
// re-issues LISTEN automatically (with a 1-second back-off).
func (w *PgWatcher) Watch(ctx context.Context, fn func()) error {
	watchCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	go func() {
		defer close(w.done)
		w.listenLoop(watchCtx, fn)
	}()

	return nil
}

// listenLoop runs the LISTEN/WaitForNotification cycle.
// On connection loss it backs off 1 second and retries.
func (w *PgWatcher) listenLoop(ctx context.Context, fn func()) {
	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := w.pool.Acquire(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logWarn("watcher: acquire connection failed; retrying", err)
			w.incrCounter("iam.watcher.acquire_error")
			w.backoff(ctx)
			continue
		}

		if _, err := conn.Exec(ctx, "LISTEN "+pgChannel); err != nil {
			conn.Release()
			if ctx.Err() != nil {
				return
			}
			w.logWarn("watcher: LISTEN failed; retrying", err)
			w.incrCounter("iam.watcher.listen_error")
			w.backoff(ctx)
			continue
		}

		w.logInfo("watcher: listening on channel " + pgChannel)

		// Receive notifications until the context is cancelled or the
		// connection dies.
		for {
			_, err := conn.Conn().WaitForNotification(ctx)
			if ctx.Err() != nil {
				conn.Release()
				return
			}
			if err != nil {
				conn.Release()
				w.logWarn("watcher: notification error; reconnecting", err)
				w.incrCounter("iam.watcher.recv_error")
				w.backoff(ctx)
				break // re-acquire outer loop
			}

			w.incrCounter("iam.watcher.reload_triggered")
			fn()
		}
	}
}

// Close cancels the watch goroutine and waits for it to exit (up to 5 s).
func (w *PgWatcher) Close() error {
	if w.cancel == nil {
		return nil // Watch was never called
	}
	w.cancel()
	select {
	case <-w.done:
	case <-time.After(5 * time.Second):
		// Goroutine didn't exit cleanly; log if possible and continue.
		w.logWarn("watcher: Close timed out waiting for goroutine", nil)
	}
	return nil
}

// ── internal helpers ─────────────────────────────────────────────────────────

func (w *PgWatcher) backoff(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
	}
}

func (w *PgWatcher) incrCounter(name string) {
	if w.metrics != nil {
		w.metrics.IncrementCounter(name, nil)
	}
}

func (w *PgWatcher) logInfo(msg string) {
	if w.log != nil {
		w.log.Info(msg)
	}
}

func (w *PgWatcher) logWarn(msg string, err error) {
	if w.log == nil {
		return
	}
	fields := logger.Fields{}
	if err != nil {
		fields["error"] = err.Error()
	}
	w.log.Warn(msg, fields)
}
