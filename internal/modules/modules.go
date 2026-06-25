// Package modules is the single import point that registers all built-in
// EntityDefinitions. Import it with a blank identifier in main or wire:
//
//	import _ "awo.so/internal/modules"
package modules

import (
	_ "awo.so/internal/modules/crm"
	_ "awo.so/internal/modules/finance"
	_ "awo.so/internal/modules/hr"
	_ "awo.so/internal/modules/platform"
)
