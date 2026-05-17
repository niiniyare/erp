package contract

import "awo.so/internal/core/iam"

// ExportNewSessionContext exposes the unexported newSessionContext constructor
// for use in external test packages (package contract_test).
//
// This file is compiled only during `go test` — it is never part of the
// production binary.
var ExportNewSessionContext = func(s *iam.ResolvedSession) SessionContext {
	return newSessionContext(s)
}
