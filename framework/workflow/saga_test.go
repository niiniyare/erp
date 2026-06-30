package workflow_test

import (
	"context"
	"errors"
	"testing"

	"awo.so/framework/workflow"
)

func TestActivitySaga_Rollback_CallsLIFO(t *testing.T) {
	s := workflow.NewActivitySaga()
	order := []int{}

	s.AddCompensation(func(_ context.Context) error { order = append(order, 1); return nil })
	s.AddCompensation(func(_ context.Context) error { order = append(order, 2); return nil })
	s.AddCompensation(func(_ context.Context) error { order = append(order, 3); return nil })

	err := s.Rollback(context.Background(), errors.New("step 3 failed"))
	if err == nil {
		t.Fatal("want non-nil error from Rollback")
	}

	if len(order) != 3 || order[0] != 3 || order[1] != 2 || order[2] != 1 {
		t.Errorf("want LIFO order [3 2 1], got %v", order)
	}
}

func TestActivitySaga_Rollback_AllCompensationsRunDespiteErrors(t *testing.T) {
	s := workflow.NewActivitySaga()
	ran := 0

	s.AddCompensation(func(_ context.Context) error { ran++; return errors.New("comp 1 failed") })
	s.AddCompensation(func(_ context.Context) error { ran++; return nil })
	s.AddCompensation(func(_ context.Context) error { ran++; return errors.New("comp 3 failed") })

	err := s.Rollback(context.Background(), errors.New("original"))
	if err == nil {
		t.Fatal("want error")
	}
	if ran != 3 {
		t.Errorf("want all 3 compensations attempted, got %d", ran)
	}
}

func TestActivitySaga_Rollback_WrapsOriginalError(t *testing.T) {
	s := workflow.NewActivitySaga()
	original := errors.New("db write failed")

	err := s.Rollback(context.Background(), original)
	if !errors.Is(err, original) {
		t.Errorf("want original error in chain, got %v", err)
	}
}

func TestActivitySaga_NoCompensations_ReturnsWrapped(t *testing.T) {
	s := workflow.NewActivitySaga()
	original := errors.New("nothing to compensate")
	err := s.Rollback(context.Background(), original)
	if err == nil {
		t.Fatal("want non-nil")
	}
	if !errors.Is(err, original) {
		t.Errorf("want original wrapped, got %v", err)
	}
}
