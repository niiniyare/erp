// Package customfield is a compatibility shim. New code should import awo.so/framework/platform/customfield.
package customfield

import platformcf "awo.so/framework/platform/customfield"

type LoadFunc = platformcf.LoadFunc
type SaveFunc = platformcf.SaveFunc
type Registry = platformcf.Registry

var (
	New          = platformcf.New
	MergedFields = platformcf.MergedFields
)
