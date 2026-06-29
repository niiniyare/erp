// Package hooks executes EntityDefinition lifecycle hooks within write transactions.
package hooks

import (
	"context"
	"fmt"

	"awo.so/framework/definition"
)

// Executor runs the hook chain for a given entity and operation.
type Executor struct{}

// New returns an Executor. Stateless; safe to share.
func New() *Executor { return &Executor{} }

// RunBeforeValidate executes before_validate hooks. Runs outside the DB transaction.
func (e *Executor) RunBeforeValidate(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, definition.HookBeforeValidate)
}

// RunBefore executes before_save hooks. Runs outside the DB transaction.
// Any error aborts the chain and is returned immediately.
func (e *Executor) RunBefore(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, definition.HookBeforeSave)
}

// RunAfter executes after_save hooks. Runs inside the DB transaction.
// Error rolls back the enclosing transaction — callers must propagate it.
func (e *Executor) RunAfter(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, definition.HookAfterSave)
}

// RunBeforeDelete executes before_delete hooks. Runs outside the DB transaction.
func (e *Executor) RunBeforeDelete(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, definition.HookBeforeDelete)
}

// RunOnSubmit executes on_submit hooks. Runs inside the DB transaction.
func (e *Executor) RunOnSubmit(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, definition.HookOnSubmit)
}

// RunOnCancel executes on_cancel hooks. Runs inside the DB transaction.
func (e *Executor) RunOnCancel(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, definition.HookOnCancel)
}

func run(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation, timing definition.HookTiming) error {
	for i, h := range def.Hooks {
		if !h.EffectiveOps().Is(m.Op) {
			continue
		}
		if h.EffectiveTiming() != timing {
			continue
		}
		name := h.Name
		if name == "" {
			name = fmt.Sprintf("hook[%d]", i)
		}
		if err := h.Fn(ctx, m); err != nil {
			return fmt.Errorf("%s %s/%s %s: %w", name, def.Name, m.Op, timing, err)
		}
	}
	return nil
}
