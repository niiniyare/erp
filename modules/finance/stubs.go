package finance

import (
	"fmt"

	"awo.so/awo/def"
)

// stubAction returns an ActionHandlerFunc that returns a not-implemented error.
// Used for actions that are declared in the entity definition but whose business
// logic has not yet been implemented. The handler is intentionally non-nil so
// that registry.BuildFrom does not reject the ActionDef.
func stubAction(name string) def.ActionHandlerFunc {
	return func(_ *def.ActionContext) (*def.ActionResult, error) {
		return nil, fmt.Errorf("finance: action %q not yet implemented", name)
	}
}
