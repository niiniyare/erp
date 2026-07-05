package workflow

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"
)

// Saga manages a sequence of compensatable steps inside a Temporal workflow.
//
// Usage pattern:
//
//	s := workflow.NewSaga(ctx)
//	if err := workflow.ExecuteActivity(ctx, StepOneActivity, ...).Get(ctx, nil); err != nil {
//	    return s.Rollback(err)
//	}
//	s.AddCompensation(func(ctx workflow.Context) error {
//	    return workflow.ExecuteActivity(ctx, UndoStepOneActivity, ...).Get(ctx, nil)
//	})
//	// ... more steps
//	return result, nil
//
// Compensations run in LIFO order. Compensation errors are logged but do not
// prevent other compensations from running.
type Saga struct {
	ctx           workflow.Context
	compensations []sagaFn
}

// sagaFn is a single compensation step.
type sagaFn func(ctx workflow.Context) error

// NewSaga creates a Saga bound to the given workflow context.
func NewSaga(ctx workflow.Context) *Saga {
	return &Saga{ctx: ctx}
}

// AddCompensation registers a compensation to run on rollback.
// Compensations are called in LIFO order (last added, first run).
func (s *Saga) AddCompensation(fn func(ctx workflow.Context) error) {
	s.compensations = append(s.compensations, fn)
}

// Rollback runs all registered compensations in reverse order and wraps
// originalErr in a descriptive error. Compensation errors are collected and
// surfaced alongside the original error but do not abort remaining compensations.
func (s *Saga) Rollback(originalErr error) error {
	logger := workflow.GetLogger(s.ctx)
	var compErrs []error

	for i := len(s.compensations) - 1; i >= 0; i-- {
		if cErr := s.compensations[i](s.ctx); cErr != nil {
			logger.Error("saga compensation failed",
				"step", i,
				"compensation_error", cErr.Error(),
				"original_error", originalErr.Error(),
			)
			compErrs = append(compErrs, cErr)
		}
	}

	if len(compErrs) > 0 {
		return fmt.Errorf("saga rollback: original=%w; %d compensation error(s): %v",
			originalErr, len(compErrs), compErrs)
	}
	return fmt.Errorf("saga rollback: %w", originalErr)
}

// ── ActivitySaga — for use inside Temporal activities ─────────────────────────

// ActivitySaga manages compensatable steps that run inside Go code (e.g. inside
// a Temporal activity, or in tests). Unlike Saga it does not require a
// workflow.Context and can be used anywhere.
type ActivitySaga struct {
	compensations []activityFn
}

// activityFn is a compensation that runs in plain Go context.
type activityFn func(ctx context.Context) error

// NewActivitySaga creates an ActivitySaga.
func NewActivitySaga() *ActivitySaga {
	return &ActivitySaga{}
}

// AddCompensation registers a compensation step.
func (s *ActivitySaga) AddCompensation(fn func(ctx context.Context) error) {
	s.compensations = append(s.compensations, fn)
}

// Rollback runs all compensations in LIFO order and returns a wrapped error.
// All compensations are attempted regardless of intermediate failures.
func (s *ActivitySaga) Rollback(ctx context.Context, originalErr error) error {
	var compErrs []error

	for i := len(s.compensations) - 1; i >= 0; i-- {
		if cErr := s.compensations[i](ctx); cErr != nil {
			compErrs = append(compErrs, fmt.Errorf("compensation[%d]: %w", i, cErr))
		}
	}

	if len(compErrs) > 0 {
		return fmt.Errorf("saga rollback: original=%w; %d compensation error(s): %v",
			originalErr, len(compErrs), compErrs)
	}
	return fmt.Errorf("saga rollback: %w", originalErr)
}
