package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"awo.so/framework/workflow"
)

func newTestRetryQueue(t *testing.T) (*workflow.RedisRetryQueue, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return workflow.NewRedisRetryQueue(rdb), mr
}

func testItem() workflow.RetryItem {
	return workflow.RetryItem{
		ID:           uuid.New(),
		TenantID:     "tenant-1",
		EntityName:   "invoice",
		RecordID:     uuid.New(),
		Op:           "create",
		WorkflowType: "InvoiceApprovalWorkflow",
		TaskQueue:    "finance",
		WorkflowID:   "tenant-1.invoice.abc.create",
	}
}

func TestRetryQueue_EnqueueAndLen(t *testing.T) {
	q, _ := newTestRetryQueue(t)
	ctx := context.Background()

	if err := q.Enqueue(ctx, testItem()); err != nil {
		t.Fatal(err)
	}
	n, err := q.Len(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("want 1, got %d", n)
	}
}

func TestRetryQueue_Drain_Success(t *testing.T) {
	q, _ := newTestRetryQueue(t)
	ctx := context.Background()

	_ = q.Enqueue(ctx, testItem())
	_ = q.Enqueue(ctx, testItem())

	called := 0
	err := q.Drain(ctx, func(_ context.Context, _ workflow.RetryItem) error {
		called++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if called != 2 {
		t.Errorf("want 2 calls, got %d", called)
	}
	n, _ := q.Len(ctx)
	if n != 0 {
		t.Errorf("want 0 after drain, got %d", n)
	}
}

func TestRetryQueue_Drain_FailedItemRequeued(t *testing.T) {
	q, _ := newTestRetryQueue(t)
	ctx := context.Background()

	_ = q.Enqueue(ctx, testItem())

	_ = q.Drain(ctx, func(_ context.Context, _ workflow.RetryItem) error {
		return errors.New("temporal unavailable")
	})

	n, _ := q.Len(ctx)
	if n != 1 {
		t.Errorf("want item re-queued (len=1), got %d", n)
	}
}

func TestRetryQueue_Drain_ExceedMaxAttempts_Discarded(t *testing.T) {
	q, _ := newTestRetryQueue(t)
	ctx := context.Background()

	item := testItem()
	item.Attempts = 9 // one more failure → discard
	_ = q.Enqueue(ctx, item)

	_ = q.Drain(ctx, func(_ context.Context, _ workflow.RetryItem) error {
		return errors.New("still failing")
	})

	n, _ := q.Len(ctx)
	if n != 0 {
		t.Errorf("want item discarded after max attempts, got len=%d", n)
	}
}
