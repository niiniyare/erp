package filter

// Walk calls visitor for every node in the filter tree in pre-order
// (parent before children). Returns on the first non-nil error.
//
// Walk is useful for validating that all field references in a filter are
// valid column names before the filter is compiled to SQL, e.g.:
//
//	allowed := map[string]bool{"status": true, "created_at": true}
//	err := f.Walk(func(n *Filter) error {
//	    if n.Field != "" && !allowed[n.Field] {
//	        return fmt.Errorf("unknown field %q", n.Field)
//	    }
//	    return nil
//	})
func (f *Filter) Walk(visitor func(*Filter) error) error {
	if f.IsNone() {
		return nil
	}
	if err := visitor(f); err != nil {
		return err
	}
	switch f.Kind {
	case KindAnd, KindOr:
		for _, child := range f.Children {
			if err := child.Walk(visitor); err != nil {
				return err
			}
		}
	case KindNot:
		if f.Inner != nil {
			return f.Inner.Walk(visitor)
		}
	case KindJSONPath:
		// Inner holds the scalar op; its Field is intentionally "" (the
		// Column field of the parent node names the JSONB column).
		// Do NOT recurse into Inner — it would visit a phantom empty Field.
	}
	return nil
}

// Fields returns all column/field names referenced by leaf predicates in
// the filter tree. JSONPath and JSONContains nodes contribute their Column;
// all other nodes contribute their Field. The returned slice is deduplicated
// and ordered by first appearance.
//
// This is useful for building column allowlists:
//
//	for _, col := range f.Fields() {
//	    if !allowedCols[col] {
//	        return fmt.Errorf("unknown filter column %q", col)
//	    }
//	}
func (f *Filter) Fields() []string {
	seen := make(map[string]bool)
	var out []string
	_ = f.Walk(func(n *Filter) error {
		var name string
		switch n.Kind {
		case KindJSONPath, KindJSONContains:
			name = n.Column
		default:
			name = n.Field
		}
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
		return nil
	})
	return out
}
