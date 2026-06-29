// Package naming is a compatibility shim. New code should import awo.so/framework/platform/naming.
package naming

import platformnaming "awo.so/framework/platform/naming"

type ExecOneRow = platformnaming.ExecOneRow

var (
	Next   = platformnaming.Next
	Format = platformnaming.Format
	Stamp  = platformnaming.Stamp
)
