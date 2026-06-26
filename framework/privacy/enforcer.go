// Package privacy evaluates EntityDefinition policy chains for read/write access.
package privacy

import (
	"context"
	"errors"
	"fmt"

	"awo.so/framework/definition"
)

// Enforcer evaluates policy chains. Stateless; safe to share.
type Enforcer struct{}

// New returns an Enforcer.
func New() *Enforcer { return &Enforcer{} }

// Allow returns nil if the viewer may perform op on record, or an error otherwise.
//
// Chain semantics:
//   - ErrAllow  → grant, stop
//   - ErrDeny   → deny, stop
//   - ErrSkip   → continue to next policy
//   - any other → deny + log (treated as ErrDeny)
//
// Fail-closed: if chain exhausts without ErrAllow → deny.
func (e *Enforcer) Allow(
	ctx context.Context,
	def *definition.EntityDefinition,
	viewer definition.ViewerContext,
	op definition.Op,
	record definition.Record,
) error {
	for i, p := range def.Policies {
		if !p.EffectiveOps().Is(op) {
			continue
		}
		err := p.Fn(ctx, viewer, op, record)
		switch {
		case errors.Is(err, definition.ErrAllow):
			return nil
		case errors.Is(err, definition.ErrDeny):
			return fmt.Errorf("policy[%d] denied %s/%s for actor %s",
				i, def.Name, op, viewer.ActorID())
		case errors.Is(err, definition.ErrSkip):
			continue
		case err != nil:
			// Unexpected error treated as deny.
			return fmt.Errorf("policy[%d] error %s/%s: %w", i, def.Name, op, err)
		default:
			// nil return from policy treated as ErrSkip.
			continue
		}
	}
	// No policy granted — fail closed.
	return fmt.Errorf("no policy granted %s/%s for actor %s (fail-closed)",
		def.Name, op, viewer.ActorID())
}

