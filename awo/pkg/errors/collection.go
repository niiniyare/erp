package errors

import (
	"fmt"
	"strings"
	"time"
)

// severityRank gives each Severity a numeric rank so "what's the worst
// severity in this collection" can be computed with a simple max instead
// of a chain of if/else conditions that's easy to get subtly wrong when a
// new severity is added later.
var severityRank = map[Severity]int{
	SeverityInfo:     0,
	SeverityWarning:  1,
	SeverityError:    2,
	SeverityCritical: 3,
}

// Collection aggregates multiple, possibly-unrelated errors produced by a
// single logical operation — e.g. a bulk import that processes 500 rows
// and wants to report every row that failed, not just the first one.
//
// This is distinct from FieldErrors: FieldErrors is specifically "several
// problems with several fields of ONE request", while Collection is
// "several independent errors from ONE operation", which may or may not
// themselves be FieldErrors, AwoErrors, or anything else.
type Collection struct {
	errs      []error
	Operation string         // what operation produced these errors, e.g. "bulk_import_journal_entries"
	TenantID  string
	RequestID string
	Timestamp time.Time
	severity  Severity // highest severity seen so far, see Add
}

// NewCollection creates an empty Collection for the named operation.
func NewCollection(operation string) *Collection {
	return &Collection{
		Operation: operation,
		Timestamp: time.Now(),
		severity:  SeverityInfo,
	}
}

// Add records err in the collection (no-op if err is nil) and updates the
// collection's tracked max severity if err carries a higher one.
func (c *Collection) Add(err error) {
	if err == nil {
		return
	}
	c.errs = append(c.errs, err)

	var sev interface{ GetSeverity() Severity }
	if s, ok := err.(interface{ GetSeverity() Severity }); ok {
		sev = s
	}
	if sev != nil {
		if severityRank[sev.GetSeverity()] > severityRank[c.severity] {
			c.severity = sev.GetSeverity()
		}
	}
}

// HasErrors reports whether anything has been added to the collection.
func (c *Collection) HasErrors() bool { return len(c.errs) > 0 }

// Count returns the number of errors recorded.
func (c *Collection) Count() int { return len(c.errs) }

// Errors returns the raw slice of recorded errors, in the order they were added.
func (c *Collection) Errors() []error { return c.errs }

// Severity returns the highest severity among all recorded errors.
func (c *Collection) Severity() Severity { return c.severity }

// ByCategory returns only the errors whose Category matches (via the
// Categorized interface — works for AwoError and any module error type
// implementing it, not just AwoError specifically).
func (c *Collection) ByCategory(category Category) []error {
	var out []error
	for _, err := range c.errs {
		var cat Categorized
		if cc, ok := err.(Categorized); ok {
			cat = cc
		}
		if cat != nil && cat.ErrorCategory() == category {
			out = append(out, err)
		}
	}
	return out
}

// Error implements the standard error interface.
func (c *Collection) Error() string {
	switch len(c.errs) {
	case 0:
		return "no errors"
	case 1:
		return c.errs[0].Error()
	default:
		msgs := make([]string, len(c.errs))
		for i, e := range c.errs {
			msgs[i] = e.Error()
		}
		return fmt.Sprintf("%d errors occurred during %s: [%s]", len(c.errs), c.Operation, strings.Join(msgs, "; "))
	}
}

// ErrorCode implements Coded. A mixed collection is reported under
// CodeValidationFailed if it's entirely field errors, or CodeInternal
// otherwise — for anything more specific, inspect Errors() directly.
func (c *Collection) ErrorCode() Code {
	if len(c.errs) > 0 {
		allValidation := true
		for _, e := range c.errs {
			if _, ok := e.(FieldErrors); !ok {
				if _, ok := e.(FieldError); !ok {
					allValidation = false
					break
				}
			}
		}
		if allValidation {
			return CodeValidationFailed
		}
	}
	return CodeInternal
}

// HTTPStatus implements HTTPStatuser, mirroring the same "all validation
// errors → 400, otherwise 500" rule as ErrorCode.
func (c *Collection) HTTPStatus() int {
	if c.ErrorCode() == CodeValidationFailed {
		return 400
	}
	return 500
}
