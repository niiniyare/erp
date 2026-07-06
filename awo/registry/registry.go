package registry

import (
	"fmt"
	"regexp"
	"strings"

	"awo.so/awo/def"
)

// Registry is the sealed, validated collection of EntityDefinitions. Once
// built it is read-only and safe for concurrent access.
type Registry struct {
	defs   []def.EntityDefinition
	byName map[string]def.EntityDefinition
}

// All returns all registered definitions in registration order.
func (r *Registry) All() []def.EntityDefinition {
	out := make([]def.EntityDefinition, len(r.defs))
	copy(out, r.defs)
	return out
}

// Lookup returns the EntityDefinition with the given name, or nil if not found.
func (r *Registry) Lookup(name string) def.EntityDefinition {
	return r.byName[name]
}

// Count returns the total number of registered definitions.
func (r *Registry) Count() int { return len(r.defs) }

// SystemEntities returns only SystemDefinition entries.
func (r *Registry) SystemEntities() []def.EntityDefinition {
	var out []def.EntityDefinition
	for _, d := range r.defs {
		if d.IsSystem() {
			out = append(out, d)
		}
	}
	return out
}

// CustomEntities returns only CustomDefinition entries.
func (r *Registry) CustomEntities() []def.EntityDefinition {
	var out []def.EntityDefinition
	for _, d := range r.defs {
		if !d.IsSystem() {
			out = append(out, d)
		}
	}
	return out
}

// ByModule returns all definitions that belong to the named module.
func (r *Registry) ByModule(module string) []def.EntityDefinition {
	var out []def.EntityDefinition
	for _, d := range r.defs {
		if d.EntityModule() == module {
			out = append(out, d)
		}
	}
	return out
}

// mandatorySystemEntities lists entity names that MUST be declared as
// SystemDefinition. Declaring any of these as CustomDefinition is a fatal
// validation error.
var mandatorySystemEntities = map[string]string{
	"iam_user":              "IAM data; JSONB corruption risk",
	"platform_tenant":       "Platform identity; accessible before per-tenant schemas load",
	"finance_ledger_entry":  "Double-entry accounting; SQL numeric constraints",
	"finance_journal_entry": "Double-entry; debit/credit balance enforced at DB level",
	"finance_payment":       "Financial transaction; SQL constraints + audit triggers",
	"finance_tax_entry":     "KRA eTIMS record; regulatory compliance",
	"inventory_stock_move":  "Inventory accounting; SQL quantity constraints",
}

// entityNameRe validates the {module}_{noun} snake_case format.
var entityNameRe = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z][a-z0-9]*)+$`)

// fieldNameRe validates snake_case field names.
var fieldNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Build reads all definitions registered via [def.Register], validates them,
// seals the def registry, and returns a read-only [Registry].
//
// Build is called exactly once during the Initialization Phase. It panics if
// validation fails — a process that cannot build a valid registry cannot serve
// requests.
func Build() *Registry {
	// Seal the def registry immediately: no new registrations after Build.
	def.Seal()

	defs := def.All()
	if len(defs) == 0 {
		panic("registry.Build: no EntityDefinitions registered — did module init() functions run?")
	}

	byName := make(map[string]def.EntityDefinition, len(defs))
	for _, d := range defs {
		byName[d.EntityName()] = d
	}

	var errs []string

	for _, d := range defs {
		errs = append(errs, validateDefinition(d, byName)...)
	}

	if len(errs) > 0 {
		msg := strings.Join(errs, "\n  - ")
		panic(fmt.Sprintf("registry.Build: validation failed:\n  - %s", msg))
	}

	return &Registry{defs: defs, byName: byName}
}

// BuildFrom constructs a Registry from an explicit list of EntityDefinitions
// without touching the global def registry. Intended for use in tests and
// tooling that needs an isolated registry.
//
// Returns an error instead of panicking so callers (e.g. test harnesses) can
// report failures cleanly via t.Fatal.
func BuildFrom(defs []def.EntityDefinition) (*Registry, error) {
	if len(defs) == 0 {
		return nil, fmt.Errorf("registry.BuildFrom: no EntityDefinitions provided")
	}

	byName := make(map[string]def.EntityDefinition, len(defs))
	for _, d := range defs {
		byName[d.EntityName()] = d
	}

	var errs []string
	for _, d := range defs {
		errs = append(errs, validateDefinition(d, byName)...)
	}

	if len(errs) > 0 {
		msg := strings.Join(errs, "\n  - ")
		return nil, fmt.Errorf("registry.BuildFrom: validation failed:\n  - %s", msg)
	}

	return &Registry{defs: defs, byName: byName}, nil
}

func validateDefinition(d def.EntityDefinition, byName map[string]def.EntityDefinition) []string {
	var errs []string
	name := d.EntityName()
	module := d.EntityModule()

	// Name format
	if !entityNameRe.MatchString(name) {
		errs = append(errs, fmt.Sprintf(
			"entity %q: name must match {module}_{noun} snake_case format", name,
		))
	}

	// Module must be non-empty
	if module == "" {
		errs = append(errs, fmt.Sprintf("entity %q: Module is empty", name))
	}

	// Name must start with module prefix
	if module != "" && !strings.HasPrefix(name, module+"_") {
		errs = append(errs, fmt.Sprintf(
			"entity %q: name must begin with module prefix %q_", name, module,
		))
	}

	// Label required
	if d.EntityLabel() == "" {
		errs = append(errs, fmt.Sprintf("entity %q: Label is empty", name))
	}

	// Mandatory system entity check
	if reason, mandatory := mandatorySystemEntities[name]; mandatory && !d.IsSystem() {
		errs = append(errs, fmt.Sprintf(
			"entity %q: must be a SystemDefinition (%s)", name, reason,
		))
	}

	// Field validation
	fieldNames := make(map[string]bool)
	for _, f := range d.EntityFields() {
		if f.Name == "" {
			errs = append(errs, fmt.Sprintf("entity %q: field has empty Name", name))
			continue
		}
		if !fieldNameRe.MatchString(f.Name) {
			errs = append(errs, fmt.Sprintf(
				"entity %q: field %q name must be snake_case", name, f.Name,
			))
		}
		if fieldNames[f.Name] {
			errs = append(errs, fmt.Sprintf(
				"entity %q: duplicate field name %q", name, f.Name,
			))
		}
		fieldNames[f.Name] = true

		// Link target must be registered
		if f.Type == def.FieldTypeLink || f.Type == def.FieldTypeLinkList {
			if f.LinkTarget == "" {
				errs = append(errs, fmt.Sprintf(
					"entity %q: field %q (Link/LinkList) has empty LinkTarget", name, f.Name,
				))
			} else if byName[f.LinkTarget] == nil {
				errs = append(errs, fmt.Sprintf(
					"entity %q: field %q LinkTarget %q is not registered",
					name, f.Name, f.LinkTarget,
				))
			}
		}

		// NamingSeries requires a Series pattern
		if f.Type == def.FieldTypeNamingSeries && f.Series == "" {
			errs = append(errs, fmt.Sprintf(
				"entity %q: field %q (NamingSeries) has empty Series pattern", name, f.Name,
			))
		}

		// Select requires Options
		if f.Type == def.FieldTypeSelect && len(f.Options) == 0 {
			errs = append(errs, fmt.Sprintf(
				"entity %q: field %q (Select) has no Options", name, f.Name,
			))
		}

		// Currency must never be FieldTypeFloat — catch accidental misuse
		// (we cannot catch it at compile time for named fields, but we can
		// flag fields named "*amount*", "*total*", "*price*" that use Float)
		if f.Type == def.FieldTypeFloat {
			lower := strings.ToLower(f.Name)
			for _, hint := range []string{"amount", "total", "price", "cost", "fee", "tax"} {
				if strings.Contains(lower, hint) {
					errs = append(errs, fmt.Sprintf(
						"entity %q: field %q looks monetary but uses FieldTypeFloat — "+
							"use FieldTypeCurrency (numeric(20,4)) instead",
						name, f.Name,
					))
					break
				}
			}
		}
	}

	// Edge validation
	edgeNames := make(map[string]bool)
	for _, e := range d.EntityEdges() {
		if e.Name == "" {
			errs = append(errs, fmt.Sprintf("entity %q: edge has empty Name", name))
			continue
		}
		if edgeNames[e.Name] {
			errs = append(errs, fmt.Sprintf("entity %q: duplicate edge name %q", name, e.Name))
		}
		edgeNames[e.Name] = true

		if e.Target == "" {
			errs = append(errs, fmt.Sprintf(
				"entity %q: edge %q has empty Target", name, e.Name,
			))
		} else if byName[e.Target] == nil {
			errs = append(errs, fmt.Sprintf(
				"entity %q: edge %q Target %q is not registered", name, e.Name, e.Target,
			))
		}
	}

	// Action name uniqueness
	actionNames := make(map[string]bool)
	for _, a := range d.EntityActions() {
		if a.Name == "" {
			errs = append(errs, fmt.Sprintf("entity %q: action has empty Name", name))
			continue
		}
		if actionNames[a.Name] {
			errs = append(errs, fmt.Sprintf(
				"entity %q: duplicate action name %q", name, a.Name,
			))
		}
		actionNames[a.Name] = true

		if a.HandlerFunc == nil {
			errs = append(errs, fmt.Sprintf(
				"entity %q: action %q has nil HandlerFunc", name, a.Name,
			))
		}
	}

	return errs
}
