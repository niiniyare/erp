// Package audit is a compatibility shim. New code should import awo.so/framework/platform/audit.
package audit

import platformaudit "awo.so/framework/platform/audit"

type Entry = platformaudit.Entry
type WriteFunc = platformaudit.WriteFunc
type ExecFunc = platformaudit.ExecFunc

var Write = platformaudit.Write
