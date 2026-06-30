package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RetryQueue persists workflow trigger failures so they can be re-attempted
// after Temporal recovers. It is consumed by a background goroutine or a
// separate process; exact delivery semantics are at-least-once.
//
// The queue is intentionally simple: it does not implement exponential backoff
// or dead-letter routing. Those belong in the Temporal server (retry policy on
// the workflow/activity, not on the trigger side).
type RetryQueue interface {
	// Enqueue records a failed workflow trigger for later re-attempt.
	Enqueue(ctx context.Context, item RetryItem) error

	// Drain pops all pending items. The handler is called for each; items where
	// handler returns nil are not re-queued. Items where handler errors are
	// re-enqueued with an incremented attempt count.
	Drain(ctx context.Context, handler func(ctx context.Context, item RetryItem) error) error

	// Len returns the number of items currently queued.
	Len(ctx context.Context) (int64, error)
}

// RetryItem is one queued workflow trigger that failed on its first attempt.
type RetryItem struct {
	// ID uniquely identifies this retry item; used for idempotent re-enqueue.
	ID uuid.UUID `json:"id"`

	// TenantID is the tenant that owns the triggering mutation.
	TenantID string `json:"tenant_id"`

	// EntityName is the registered entity type name.
	EntityName string `json:"entity_name"`

	// RecordID is the UUID of the mutated record.
	RecordID uuid.UUID `json:"record_id"`

	// Op is the mutation operation that fired the trigger.
	Op string `json:"op"`

	// WorkflowType is the Temporal workflow type to start.
	WorkflowType string `json:"workflow_type"`

	// TaskQueue is the Temporal task queue for the workflow.
	TaskQueue string `json:"task_queue"`

	// WorkflowID is the pre-computed Temporal workflow ID.
	WorkflowID string `json:"workflow_id"`

	// Attempts is the number of times this item has been tried (including the
	// initial attempt that failed before enqueue).
	Attempts int `json:"attempts"`

	// EnqueuedAt is when the item was first added to the queue.
	EnqueuedAt time.Time `json:"enqueued_at"`
}

// ── Redis implementation ───────────────────────────────────────────────────────

const (
	retryQueueKey    = "awo:wf:retry"
	retryMaxAttempts = 10
)

// RedisRetryQueue is a RetryQueue backed by a Redis list.
// Items are stored as JSON in a FIFO list. Failed re-deliveries are
// re-appended (RPUSH) after incrementing Attempts.
type RedisRetryQueue struct {
	rdb redis.UniversalClient
}

// NewRedisRetryQueue creates a RedisRetryQueue backed by rdb.
func NewRedisRetryQueue(rdb redis.UniversalClient) *RedisRetryQueue {
	return &RedisRetryQueue{rdb: rdb}
}

// Enqueue marshals item and pushes it to the right end of the Redis list.
func (q *RedisRetryQueue) Enqueue(ctx context.Context, item RetryItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	if item.EnqueuedAt.IsZero() {
		item.EnqueuedAt = time.Now().UTC()
	}
	b, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("retry queue: marshal: %w", err)
	}
	if err := q.rdb.RPush(ctx, retryQueueKey, b).Err(); err != nil {
		return fmt.Errorf("retry queue: enqueue: %w", err)
	}
	return nil
}

// Drain pops all items currently in the queue and calls handler for each.
// Items that handler cannot process are re-queued (up to retryMaxAttempts).
// Items that exceed retryMaxAttempts are discarded with a log warning embedded
// in the returned error slice (Drain never returns early on handler errors).
func (q *RedisRetryQueue) Drain(ctx context.Context, handler func(ctx context.Context, item RetryItem) error) error {
	// Snapshot current queue length so we don't loop indefinitely on re-queued items.
	n, err := q.rdb.LLen(ctx, retryQueueKey).Result()
	if err != nil {
		return fmt.Errorf("retry queue: llen: %w", err)
	}

	var errs []error
	for i := int64(0); i < n; i++ {
		raw, err := q.rdb.LPop(ctx, retryQueueKey).Bytes()
		if err == redis.Nil {
			break // queue emptied before we reached n
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("retry queue: lpop: %w", err))
			continue
		}

		var item RetryItem
		if err := json.Unmarshal(raw, &item); err != nil {
			// Corrupted item — discard.
			errs = append(errs, fmt.Errorf("retry queue: unmarshal: %w", err))
			continue
		}

		if err := handler(ctx, item); err != nil {
			item.Attempts++
			if item.Attempts >= retryMaxAttempts {
				errs = append(errs, fmt.Errorf("retry queue: item %s exceeded max attempts (%d), discarding: %w",
					item.ID, retryMaxAttempts, err))
				continue
			}
			// Re-queue at the back.
			if rqErr := q.Enqueue(ctx, item); rqErr != nil {
				errs = append(errs, fmt.Errorf("retry queue: re-enqueue %s: %w", item.ID, rqErr))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("retry queue: drain completed with %d error(s): %v", len(errs), errs)
	}
	return nil
}

// Len returns the number of items currently in the queue.
func (q *RedisRetryQueue) Len(ctx context.Context) (int64, error) {
	n, err := q.rdb.LLen(ctx, retryQueueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("retry queue: len: %w", err)
	}
	return n, nil
}
