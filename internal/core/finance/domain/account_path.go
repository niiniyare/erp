package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// =============================================================================
// MaterialisedPath
// =============================================================================

// MaterialisedPath is a value type that encodes an ancestor chain as a
// slash-delimited string of UUIDs, e.g.:
//
//	/550e8400-e29b-41d4-a716-446655440000/6ba7b810-9dad-11d1-80b4-00c04fd430c8/
//
// Rules enforced by this type:
//   - Every segment is a valid, non-nil UUID.
//   - The string always starts and ends with a separator so that a plain
//     substring search (LIKE '%/id/%') cannot produce false positives at
//     the boundaries.
//   - The empty path ("") is the valid root sentinel; it means "no ancestors".
//
// All hierarchy path fields across the domain (Account, ReportingGroup,
// CostCenter, …) use this type so path construction and query predicates
// are consistent.
type MaterialisedPath string

const pathSep = "/"

// NewMaterialisedPath constructs a MaterialisedPath from an ordered slice of
// ancestor UUIDs (root first, immediate parent last).
// An empty slice returns the root sentinel ("").
func NewMaterialisedPath(segments []uuid.UUID) MaterialisedPath {
	if len(segments) == 0 {
		return MaterialisedPath("")
	}
	var b strings.Builder
	b.WriteString(pathSep)
	for _, id := range segments {
		b.WriteString(id.String())
		b.WriteString(pathSep)
	}
	return MaterialisedPath(b.String())
}

// Append returns a new path with nodeID appended as the next segment.
// This is used by the repository when persisting a newly created child node.
func (p MaterialisedPath) Append(nodeID uuid.UUID) MaterialisedPath {
	if p == "" {
		return MaterialisedPath(pathSep + nodeID.String() + pathSep)
	}
	return MaterialisedPath(string(p) + nodeID.String() + pathSep)
}

// ContainsAncestor reports whether ancestorID appears anywhere in the path.
// This is an O(1) string operation and maps directly to a SQL LIKE predicate:
//
//	WHERE path LIKE '%/' || $ancestorID || '/%'
func (p MaterialisedPath) ContainsAncestor(ancestorID uuid.UUID) bool {
	if p == "" {
		return false
	}
	return strings.Contains(string(p), pathSep+ancestorID.String()+pathSep)
}

// IsDescendantOf returns true when p represents a node that is a descendant
// of the node identified by ancestorID.
// Equivalent to ContainsAncestor; provided as an alias for readability at
// call sites that phrase the question from the child's perspective.
func (p MaterialisedPath) IsDescendantOf(ancestorID uuid.UUID) bool {
	return p.ContainsAncestor(ancestorID)
}

// Depth returns the number of ancestors encoded in the path.
// A root node (empty path) has depth 0.
// A direct child of the root has depth 1.
func (p MaterialisedPath) Depth() int {
	if p == "" {
		return 0
	}
	// Count non-empty segments between separators.
	parts := strings.Split(strings.Trim(string(p), pathSep), pathSep)
	return len(parts)
}

// Segments parses and returns the ancestor UUIDs in order (root first).
// Returns an error if any segment is not a valid UUID; this would indicate
// a corrupt path that should be flagged and corrected.
func (p MaterialisedPath) Segments() ([]uuid.UUID, error) {
	if p == "" {
		return nil, nil
	}
	raw := strings.Split(strings.Trim(string(p), pathSep), pathSep)
	ids := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("materialised path contains invalid UUID segment %q: %w", s, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ImmediateParentID returns the last segment of the path, which is the direct
// parent's UUID. Returns uuid.Nil and false for a root node (empty path).
func (p MaterialisedPath) ImmediateParentID() (uuid.UUID, bool) {
	segs, err := p.Segments()
	if err != nil || len(segs) == 0 {
		return uuid.Nil, false
	}
	return segs[len(segs)-1], true
}

// String returns the raw path string. Implements fmt.Stringer.
func (p MaterialisedPath) String() string { return string(p) }

// IsRoot returns true when the path is the root sentinel (empty string).
func (p MaterialisedPath) IsRoot() bool { return p == "" }

// Validate checks that the path is either the root sentinel or a well-formed
// slash-delimited sequence of valid UUIDs.
func (p MaterialisedPath) Validate() error {
	if p.IsRoot() {
		return nil
	}
	s := string(p)
	if !strings.HasPrefix(s, pathSep) || !strings.HasSuffix(s, pathSep) {
		return fmt.Errorf(
			"materialised path %q must start and end with %q", s, pathSep,
		)
	}
	if _, err := p.Segments(); err != nil {
		return err
	}
	return nil
}
