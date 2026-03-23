package iam

// assignOpts is a package-level alias kept for whitebox test access.
// New code should use AssignOpts directly.
type assignOpts = AssignOpts

// nullableString returns nil for an empty string, otherwise a pointer to s.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
