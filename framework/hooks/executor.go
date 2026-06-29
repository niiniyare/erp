// Package hooks executes EntityDefinition lifecycle hooks in the correct order
// for each write operation.
//
// # Hook execution model
//
//   - Hooks are registered on [definition.EntityDefinition] as a slice of
//     [definition.Hook] values. Each hook declares which [definition.Op]s and
//     which [definition.HookTiming] it applies to via [definition.Hook.Ops] and
//     [definition.Hook.Timing]; zero values are expanded to "all ops" and the
//     default timing by the [definition.Hook.EffectiveOps] /
//     [definition.Hook.EffectiveTiming] methods.
//
//   - [Runner.Run] is the single entry point. Callers pass the desired
//     [definition.HookTiming] explicitly so the call site documents when hooks
//     fire without requiring six separately-named methods.
//
//   - Hooks that panic are caught and converted to errors so a single
//     misbehaving hook cannot crash the server goroutine.
//
//   - If the context is already cancelled before the first hook runs, [Runner.Run]
//     returns the context error immediately without iterating the hook slice.
//
// # Typical call sequence (inside a write handler)
//
//	r := hooks.NewRunner(logger) // or hooks.DefaultRunner
//
//	if err := r.Run(ctx, def, m, definition.HookBeforeValidate); err != nil { … }
//	if err := validate(m); err != nil { … }
//	if err := r.Run(ctx, def, m, definition.HookBeforeSave); err != nil { … }
//	// begin tx
//	if err := persist(ctx, tx, m); err != nil { … }
//	if err := r.Run(ctx, def, m, definition.HookAfterSave); err != nil { … }
//	// commit tx
package hooks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"awo.so/framework/definition"
)

// ── Error type ────────────────────────────────────────────────────────────────

// HookError is returned when a hook function returns an error or panics.
// It carries enough structured context for callers to log or inspect the
// failure without parsing error strings.
type HookError struct {
	// HookName is the hook's Name field, or "hook[i]" for anonymous hooks.
	HookName string
	// Entity is the EntityDefinition.Name of the entity being mutated.
	Entity string
	// Op is the write operation (create, update, delete, …).
	Op definition.Op
	// Timing is the hook phase that failed.
	Timing definition.HookTiming
	// Panicked is true when the hook panicked rather than returning an error.
	Panicked bool
	// Err is the underlying error (or the recovered panic value converted to error).
	Err error
}

func (e *HookError) Error() string {
	phase := "hook"
	if e.Panicked {
		phase = "hook panic"
	}
	return fmt.Sprintf("%s %q on %s/%s %s: %s",
		phase, e.HookName, e.Entity, e.Op, e.Timing, e.Err)
}

// Unwrap allows errors.Is / errors.As to inspect the underlying cause.
func (e *HookError) Unwrap() error { return e.Err }

// ── Runner ────────────────────────────────────────────────────────────────────

// Runner executes the hook chain for a given entity and operation.
// The zero value is valid and uses a discarding logger; prefer [NewRunner].
type Runner struct {
	// log receives a structured warning for each hook panic before the error
	// is returned. It is never nil after construction.
	log *slog.Logger
}

// DefaultRunner is a package-level Runner usable when no logger is needed.
// Panics are still converted to errors; log output goes to the discard handler.
var DefaultRunner = &Runner{log: slog.New(noopHandler{})}

// NewRunner returns a Runner that uses logger for structured warning output
// when hooks panic. Passing nil falls back to the discard handler.
func NewRunner(logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.New(noopHandler{})
	}
	return &Runner{log: logger}
}

// Run executes all hooks registered on def whose Op and Timing match m.Op and
// timing, in registration order.
//
// Execution stops at the first failure (error or panic). If the context is
// already done before iteration begins, Run returns ctx.Err() immediately.
//
// Panics inside hook functions are recovered and returned as *HookError with
// Panicked == true so the calling write transaction can roll back cleanly.
func (r *Runner) Run(
	ctx context.Context,
	def *definition.EntityDefinition,
	m *definition.Mutation,
	timing definition.HookTiming,
) error {
	// Fast-path: refuse to start if the context is already cancelled.
	if err := ctx.Err(); err != nil {
		return err
	}

	for i, h := range def.Hooks {
		if !h.EffectiveOps().Is(m.Op) {
			continue
		}
		if h.EffectiveTiming() != timing {
			continue
		}
		if err := r.runOne(ctx, def, m, timing, h, i); err != nil {
			return err
		}
	}
	return nil
}

// runOne calls a single hook function and converts panics to *HookError.
func (r *Runner) runOne(
	ctx context.Context,
	def *definition.EntityDefinition,
	m *definition.Mutation,
	timing definition.HookTiming,
	h definition.HookDef,
	idx int,
) (retErr error) {
	name := hookName(h, idx)

	// Recover panics so a buggy hook cannot crash the server goroutine.
	defer func() {
		if rec := recover(); rec != nil {
			// Convert the panic value to an error.
			var err error
			switch v := rec.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
			r.log.WarnContext(
				ctx, "hook panicked",
				slog.String("hook", name),
				slog.String("entity", def.Name),
				slog.String("op", string(m.Op)),
				slog.String("timing", string(timing)),
				slog.Any("panic", rec),
			)
			retErr = &HookError{
				HookName: name,
				Entity:   def.Name,
				Op:       m.Op,
				Timing:   timing,
				Panicked: true,
				Err:      err,
			}
		}
	}()

	if err := h.Fn(ctx, m); err != nil {
		return &HookError{
			HookName: name,
			Entity:   def.Name,
			Op:       m.Op,
			Timing:   timing,
			Err:      err,
		}
	}
	return nil
}

// ── Convenience unwrap helpers ────────────────────────────────────────────────

// AsHookError unwraps err into a *HookError if one is present in the chain.
// Returns (nil, false) when err contains no *HookError.
//
// Useful when callers want to log the hook name or check Panicked:
//
//	if he, ok := hooks.AsHookError(err); ok && he.Panicked {
//	    metrics.Inc("hook.panics")
//	}
func AsHookError(err error) (*HookError, bool) {
	var he *HookError
	return he, errors.As(err, &he)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// hookName returns the hook's declared name, or a positional fallback.
// Using a fallback rather than a blank string keeps error messages actionable
// for hooks that were registered without an explicit name.
func hookName(h definition.HookDef, idx int) string {
	if h.Name != "" {
		return h.Name
	}
	return fmt.Sprintf("hook[%d]", idx)
}

// noopHandler is an slog.Handler that discards all log records.
// Used when no logger is provided to [NewRunner].
type noopHandler struct{}

func (noopHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (noopHandler) Handle(context.Context, slog.Record) error { return nil }
func (noopHandler) WithAttrs([]slog.Attr) slog.Handler        { return noopHandler{} }
func (noopHandler) WithGroup(string) slog.Handler             { return noopHandler{} }
