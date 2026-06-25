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

// RunBefore executes all before_save hooks matching op, in registration order.
// Any error aborts the chain and is returned immediately.
func (e *Executor) RunBefore(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, true)
}

// RunAfter executes all after_save hooks matching op, in registration order.
// Error rolls back the enclosing transaction — callers must propagate it.
func (e *Executor) RunAfter(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation) error {
	return run(ctx, def, m, false)
}

func run(ctx context.Context, def *definition.EntityDefinition, m *definition.Mutation, before bool) error {
	for i, h := range def.Hooks {
		if !h.Ops.Is(m.Op) {
			continue
		}
		if h.Before != before {
			continue
		}
		if err := h.Fn(ctx, m); err != nil {
			phase := "after_save"
			if before {
				phase = "before_save"
			}
			return fmt.Errorf("hook[%d] %s/%s %s: %w", i, def.Name, m.Op, phase, err)
		}
	}
	return nil
}
