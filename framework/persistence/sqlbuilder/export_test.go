// export_test.go exposes internal functions for white-box testing.
// Compiled only during `go test` (package sqlbuilder, not sqlbuilder_test).
package sqlbuilder

import "awo.so/framework/filter"

// ExportedPgIdent exposes pgIdent for unit tests.
func ExportedPgIdent(name string) (string, error) {
	return pgIdent(name)
}

// ExportedFilterToSQL exposes filterToSQL for unit tests.
func ExportedFilterToSQL(f *filter.Filter, startIdx int) (string, []any, error) {
	return filterToSQL(f, startIdx)
}
