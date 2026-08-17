package finance

import (
	"fmt"

	"awo.so/awo/def"
)

// stubAction returns a HandlerFunc that rejects the action with a "not implemented"
// error. Used as a placeholder until concrete business logic is wired.
func stubAction(name string) def.ActionHandlerFunc {
	return func(ctx *def.ActionContext) (*def.ActionResult, error) {
		return nil, fmt.Errorf("action %q: not yet implemented", name)
	}
}
