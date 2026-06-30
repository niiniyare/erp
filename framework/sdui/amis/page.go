package amis

import (
	"fmt"

	"awo.so/framework/def"
)

// PageOpts controls which optional sections are rendered.
//
// PageOpts is intentionally a plain struct (not an interface or functional
// options) to preserve backward compatibility with existing call sites that
// construct it as a struct literal. New optional behavior should be added
// as additional fields with safe zero values, never by changing the
// meaning of an existing field.
type PageOpts struct {
	// APIBase is the REST prefix, e.g. "/api". Defaults to "/api".
	APIBase string

	// Title overrides the entity label as the page heading.
	Title string

	// ExcludeFields hides specific fields from both form and table.
	ExcludeFields map[string]bool

	// ReadOnly renders forms in view-only mode (no create/edit/delete).
	ReadOnly bool

	// ExtraFields are appended after the entity's base fields.
	// Use this to inject tenant-specific custom fields loaded from the
	// customfield.Registry at request time.
	ExtraFields []*def.FieldDef
}

// resolvedAPIBase returns opts.APIBase with the historical "/api" default
// applied. Centralizing this avoids the bug where some builders honored the
// default and others (e.g. the old createButton) silently didn't.
func (o PageOpts) resolvedAPIBase() string {
	if o.APIBase == "" {
		return "/api"
	}
	return o.APIBase
}

// resolvedTitle returns the page heading, falling back from Title -> Label -> Name.
func (o PageOpts) resolvedTitle(entity *def.EntityDefinition) string {
	if o.Title != "" {
		return o.Title
	}
	return entityLabel(entity)
}

// entityLabel returns entity.Label, falling back to entity.Name so callers
// never render a blank heading or blank validation message.
func entityLabel(entity *def.EntityDefinition) string {
	if entity == nil {
		return ""
	}
	if entity.Label != "" {
		return entity.Label
	}
	return entity.Name
}

// apiURLFor builds the collection endpoint for an entity under the given
// (already-defaulted) API base, e.g. "/api" + "users" -> "/api/users".
func apiURLFor(base string, entity *def.EntityDefinition) string {
	return base + "/" + entity.TableName()
}

// CRUDPage generates a full AMIS CRUD page schema for an EntityDefinition.
// The schema includes a toolbar with Create button, a searchable table,
// and inline edit/delete actions.
//
// CRUDPage is safe to call with entity == nil; it returns a minimal error
// page schema rather than panicking, since schema builders are typically
// invoked deep inside an HTTP handler where a panic would surface as an
// opaque 500 instead of a clear diagnostic.
func CRUDPage(entity *def.EntityDefinition, opts PageOpts) map[string]any {
	if entity == nil {
		return errorPage("CRUDPage: nil EntityDefinition")
	}

	base := opts.resolvedAPIBase()
	apiURL := apiURLFor(base, entity)
	title := opts.resolvedTitle(entity)

	columns := tableColumns(entity, opts)
	if !opts.ReadOnly {
		// actionColumn now receives the entity/opts/base it needs to build
		// a real, populated Edit form instead of an empty dialog body.
		columns = append(columns, actionColumn(entity, apiURL, base, opts))
	}

	toolbar := make([]any, 0, 2)
	if !opts.ReadOnly {
		toolbar = append(toolbar, createButton(entity, base, opts))
	}
	toolbar = append(toolbar, searchBar())

	return map[string]any{
		"type":  "page",
		"title": title,
		"body": map[string]any{
			"type":          "crud",
			"api":           apiURL,
			"syncLocation":  false,
			"headerToolbar": toolbar,
			"columns":       columns,
			"footerToolbar": []any{
				"statistics",
				map[string]any{"type": "pagination", "layout": "perPage,pager,go"},
			},
		},
	}
}

// FormPage generates an AMIS form page for create or edit.
// When the record id is empty the rendered form POSTs to apiURL;
// otherwise it PUTs to apiURL/$id (see FormSchema's adaptor).
func FormPage(entity *def.EntityDefinition, opts PageOpts) map[string]any {
	if entity == nil {
		return errorPage("FormPage: nil EntityDefinition")
	}

	base := opts.resolvedAPIBase()
	apiURL := apiURLFor(base, entity)
	title := opts.resolvedTitle(entity)

	return map[string]any{
		"type":  "page",
		"title": title,
		"body":  FormSchema(entity, apiURL, opts),
	}
}

// FormSchema generates an AMIS form schema (without the page wrapper).
//
// apiURL must be the fully-resolved collection endpoint (including API
// base), e.g. "/api/users". Callers that already have a resolved base
// should use apiURLFor rather than recomputing it, to avoid the
// double-resolution bug where some builders re-read opts.APIBase and
// produced inconsistent URLs from the same PageOpts value.
func FormSchema(entity *def.EntityDefinition, apiURL string, opts PageOpts) map[string]any {
	if entity == nil {
		return map[string]any{"type": "form", "body": []any{}}
	}

	controls := formControls(entity, opts)

	return map[string]any{
		"type": "form",
		"api": map[string]any{
			"method": "post",
			"url":    apiURL,
			// Switch to PUT against the record's own URL when editing an
			// existing record (payload.id is set). Kept as a client-side
			// adaptor, matching the original schema's contract, so AMIS
			// forms can be reused unmodified for both create and edit.
			"adaptor": `
if (payload.id) {
  api.method = 'put';
  api.url = api.url + '/' + payload.id;
}
return api;`,
		},
		"body":             controls,
		"resetAfterSubmit": false,
		"wrapWithPanel":    true,
	}
}

// errorPage returns a minimal AMIS-renderable page communicating a
// programmer error (e.g. nil EntityDefinition) instead of panicking.
// Schema builders run inside request handlers; a panic here would take
// down the request instead of rendering a diagnosable message.
func errorPage(msg string) map[string]any {
	return map[string]any{
		"type":  "page",
		"title": "Error",
		"body": map[string]any{
			"type":  "alert",
			"level": "danger",
			"body":  msg,
		},
	}
}

// ──────────────────────────────────────────────────────────────────
// Internal builders
// ──────────────────────────────────────────────────────────────────

// tableColumns builds the AMIS column list for the CRUD table, excluding
// hidden, explicitly-excluded, and sensitive fields. The leading "id"
// column is always present and not toggleable off by default.
func tableColumns(entity *def.EntityDefinition, opts PageOpts) []any {
	cols := []any{
		map[string]any{"name": "id", "label": "ID", "type": "text", "toggled": false},
	}
	for _, f := range allFields(entity, opts) {
		if f == nil || skipField(f, opts) {
			continue
		}
		cols = append(cols, ColumnDef(f))
	}
	return cols
}

// skipField centralizes the visibility rule so the table and every form
// path apply the *same* exclusion logic. Previously the table additionally
// filtered IsSensitive while the form builders did not, which meant
// "sensitive" fields (e.g. secrets, internal identifiers) were hidden from
// the list view but still rendered—and editable—in create/edit forms.
func skipField(f *def.FieldDef, opts PageOpts) bool {
	return opts.ExcludeFields[f.Name] || f.Hidden || f.IsSensitive
}

// allFields returns entity.Fields merged with opts.ExtraFields, skipping
// any extra field whose name collides with a base field (base fields win).
// Safe to call with a nil entity (returns opts.ExtraFields verbatim).
func allFields(entity *def.EntityDefinition, opts PageOpts) []*def.FieldDef {
	if entity == nil {
		return opts.ExtraFields
	}
	if len(opts.ExtraFields) == 0 {
		return entity.Fields
	}

	seen := make(map[string]struct{}, len(entity.Fields))
	for _, f := range entity.Fields {
		if f != nil {
			seen[f.Name] = struct{}{}
		}
	}

	merged := make([]*def.FieldDef, len(entity.Fields), len(entity.Fields)+len(opts.ExtraFields))
	copy(merged, entity.Fields)
	for _, f := range opts.ExtraFields {
		if f == nil {
			continue
		}
		if _, clash := seen[f.Name]; !clash {
			merged = append(merged, f)
		}
	}
	return merged
}

// formControls builds the AMIS form body: a hidden id field followed by
// either grouped fieldsets (if any field declares a Section) or a flat
// list of controls.
func formControls(entity *def.EntityDefinition, opts PageOpts) []any {
	controls := []any{
		map[string]any{"type": "hidden", "name": "id"},
	}

	sections := groupBySection(entity, opts)

	// A single unnamed section means no field declared a Section, so we
	// render flat; anything else (multiple sections, or one named section)
	// renders as fieldsets. This mirrors the original behavior but now
	// shares one field-selection path with tableColumns via skipField,
	// instead of formControls/groupBySection re-implementing the filter.
	if len(sections) == 1 && sections[0].name == "" {
		for _, f := range sections[0].fields {
			controls = append(controls, FormControl(f))
		}
		return controls
	}

	for _, s := range sections {
		body := make([]any, len(s.fields))
		for i, f := range s.fields {
			body[i] = FormControl(f)
		}
		if s.name == "" {
			controls = append(controls, body...)
			continue
		}
		controls = append(controls, map[string]any{
			"type":  "fieldset",
			"title": s.name,
			"body":  body,
		})
	}
	return controls
}

// sectionGroup is an ordered bucket of fields sharing the same FieldDef.Section.
type sectionGroup struct {
	name   string
	fields []*def.FieldDef
}

// groupBySection partitions visible fields into sectionGroups, preserving
// first-seen order of section names (including the empty/default section).
// It applies the same skipField rule as tableColumns, so sensitive fields
// are now excluded from forms exactly as they already were from tables.
func groupBySection(entity *def.EntityDefinition, opts PageOpts) []sectionGroup {
	order := make([]string, 0, 4)
	groups := make(map[string]*sectionGroup, 4)

	for _, f := range allFields(entity, opts) {
		if f == nil || skipField(f, opts) {
			continue
		}
		sec := f.Section
		g, ok := groups[sec]
		if !ok {
			g = &sectionGroup{name: sec}
			groups[sec] = g
			order = append(order, sec)
		}
		g.fields = append(g.fields, f)
	}

	out := make([]sectionGroup, 0, len(order))
	for _, name := range order {
		out = append(out, *groups[name])
	}
	if len(out) == 0 {
		// Keep formControls' "flat" branch well-defined even when every
		// field was filtered out (e.g. ExcludeFields covers everything).
		out = append(out, sectionGroup{})
	}
	return out
}

// createButton renders the toolbar "Create <Entity>" button, which opens
// a dialog containing a full create form built via FormSchema.
//
// base must be the already-resolved API base (see PageOpts.resolvedAPIBase)
// so the create dialog and the surrounding CRUD table always agree on the
// same prefix, even when opts.APIBase is empty.
func createButton(entity *def.EntityDefinition, base string, opts PageOpts) map[string]any {
	label := fmt.Sprintf("Create %s", entityLabel(entity))
	apiURL := apiURLFor(base, entity)

	return map[string]any{
		"type":       "button",
		"label":      label,
		"level":      "primary",
		"actionType": "dialog",
		"dialog": map[string]any{
			"title": label,
			"body":  FormSchema(entity, apiURL, opts),
		},
	}
}

// searchBar renders the standard right-aligned table search box.
func searchBar() map[string]any {
	return map[string]any{
		"type":        "search-box",
		"name":        "q",
		"placeholder": "Search...",
		"align":       "right",
	}
}

// actionColumn renders the per-row Edit/Delete operation column.
//
// This is the fix for the headline production bug: the previous version
// only received apiURL and built an Edit dialog with no form controls at
// all (no "body"), so clicking Edit opened an empty dialog. It now
// receives entity/opts and reuses FormSchema -- the same builder used for
// Create -- so Edit and Create always render the same set of controls,
// validations, and field types, and can never drift apart.
func actionColumn(entity *def.EntityDefinition, apiURL, _ string, opts PageOpts) map[string]any {
	editAPIURL := apiURL + "/${id}"

	// Build the edit form schema, then override its api/initApi so it
	// loads and saves the specific row instead of the collection.
	editForm := FormSchema(entity, apiURL, opts)
	editForm["initApi"] = editAPIURL
	editForm["api"] = map[string]any{
		"method": "put",
		"url":    editAPIURL,
	}

	return map[string]any{
		"type":  "operation",
		"label": "Actions",
		"buttons": []any{
			map[string]any{
				"label":      "Edit",
				"type":       "button",
				"level":      "link",
				"actionType": "dialog",
				"dialog": map[string]any{
					"title": "Edit",
					"body":  editForm,
				},
			},
			map[string]any{
				"label":       "Delete",
				"type":        "button",
				"level":       "link",
				"className":   "text-danger",
				"actionType":  "ajax",
				"confirmText": "Delete this record?",
				"api": map[string]any{
					"method": "delete",
					"url":    editAPIURL,
				},
			},
		},
	}
}
