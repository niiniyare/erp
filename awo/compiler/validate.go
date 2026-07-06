package compiler

import (
	"fmt"
	"regexp"
	"strings"

	"awo.so/awo/def"
	"awo.so/awo/registry"
)

// entityNameRe matches valid entity names: lowercase letters/digits/underscores,
// must start with a letter, must end with a letter or digit.
var entityNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]*[a-z0-9]$`)

// Validate runs semantic validation against the registry and returns Diagnostics.
// Called internally by Compile; also exported for tooling.
func Validate(reg *registry.Registry) Diagnostics {
	var ds Diagnostics
	defs := reg.All()

	// Build name set for link-target resolution.
	names := make(map[string]bool, len(defs))
	for _, d := range defs {
		names[d.EntityName()] = true
	}

	for _, d := range defs {
		name := d.EntityName()

		// Entity name format.
		if !entityNameRe.MatchString(name) {
			ds = append(ds, Diagnostic{
				Severity:   SeverityError,
				EntityName: name,
				Message:    fmt.Sprintf("entity name %q does not match ^[a-z][a-z0-9_]*[a-z0-9]$", name),
			})
		}

		// Entity name must contain at least one underscore (module_noun convention).
		if !strings.Contains(name, "_") {
			ds = append(ds, Diagnostic{
				Severity:   SeverityWarning,
				EntityName: name,
				Message:    "entity name should follow {module}_{noun} convention (no underscore found)",
			})
		}

		// Field-level checks.
		fieldNames := make(map[string]bool)
		for _, f := range d.EntityFields() {
			// Duplicate field name.
			if fieldNames[f.Name] {
				ds = append(ds, Diagnostic{
					Severity:   SeverityError,
					EntityName: name,
					FieldName:  f.Name,
					Message:    "duplicate field name",
				})
			}
			fieldNames[f.Name] = true

			// Money must not be FieldTypeFloat or FieldTypeInt — flag when label hints money.
			// (We cannot detect by name alone perfectly, but flag explicit misuse patterns.)
			// The real rule: FieldTypeCurrency is the only correct money type.
			// Flag if a field named *_amount, *_price, *_cost, *_total uses float/int.
			if f.Type == def.FieldTypeFloat || f.Type == def.FieldTypeInt {
				lower := strings.ToLower(f.Name)
				for _, suffix := range []string{"_amount", "_price", "_cost", "_total", "_fee", "_tax"} {
					if strings.HasSuffix(lower, suffix) {
						ds = append(ds, Diagnostic{
							Severity:   SeverityWarning,
							EntityName: name,
							FieldName:  f.Name,
							Message:    fmt.Sprintf("field %q looks like money but uses type %q — use FieldTypeCurrency instead", f.Name, f.Type),
						})
						break
					}
				}
			}

			// NamingSeries must have a non-empty pattern.
			if f.Type == def.FieldTypeNamingSeries && f.Series == "" {
				ds = append(ds, Diagnostic{
					Severity:   SeverityError,
					EntityName: name,
					FieldName:  f.Name,
					Message:    "FieldTypeNamingSeries field must have a non-empty Series pattern",
				})
			}

			// Link targets must exist.
			if (f.Type == def.FieldTypeLink || f.Type == def.FieldTypeLinkList) && f.LinkTarget != "" {
				if !names[f.LinkTarget] {
					ds = append(ds, Diagnostic{
						Severity:   SeverityError,
						EntityName: name,
						FieldName:  f.Name,
						Message:    fmt.Sprintf("link target %q is not registered", f.LinkTarget),
					})
				}
			}

			// Required + Default is redundant (warning only — default overrides required on omission).
			if f.Required && f.Default != nil {
				ds = append(ds, Diagnostic{
					Severity:   SeverityWarning,
					EntityName: name,
					FieldName:  f.Name,
					Message:    "field is both Required and has a Default — the default satisfies the required constraint; consider removing Required",
				})
			}

			// Immutable without Required or Default means the field will always be zero on create.
			if f.Immutable && !f.Required && f.Default == nil {
				ds = append(ds, Diagnostic{
					Severity:   SeverityWarning,
					EntityName: name,
					FieldName:  f.Name,
					Message:    "Immutable field has neither Required nor Default — it will always be zero-valued and cannot be changed after creation",
				})
			}
		}

		// Edge name uniqueness.
		edgeNames := make(map[string]bool)
		for _, e := range d.EntityEdges() {
			if edgeNames[e.Name] {
				ds = append(ds, Diagnostic{
					Severity:   SeverityError,
					EntityName: name,
					Message:    fmt.Sprintf("duplicate edge name %q", e.Name),
				})
			}
			edgeNames[e.Name] = true
		}
	}

	return ds
}
