package amis

import (
	"awo.so/framework/def"
)

// PageOpts controls which optional sections are rendered.
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

// CRUDPage generates a full AMIS CRUD page schema for an EntityDefinition.
// The schema includes a toolbar with Create button, a searchable table,
// and inline edit/delete actions.
func CRUDPage(def *def.EntityDefinition, opts PageOpts) map[string]any {
	base := opts.APIBase
	if base == "" {
		base = "/api"
	}
	apiURL := base + "/" + def.TableName()

	title := opts.Title
	if title == "" {
		title = def.Label
		if title == "" {
			title = def.Name
		}
	}

	columns := tableColumns(def, opts)
	if !opts.ReadOnly {
		columns = append(columns, actionColumn(apiURL))
	}

	toolbar := []any{}
	if !opts.ReadOnly {
		toolbar = append(toolbar, createButton(def, opts))
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
// When id is empty the form POSTs to apiURL; otherwise PUTs to apiURL/$id.
func FormPage(def *def.EntityDefinition, opts PageOpts) map[string]any {
	base := opts.APIBase
	if base == "" {
		base = "/api"
	}
	apiURL := base + "/" + def.TableName()

	title := opts.Title
	if title == "" {
		title = def.Label
	}

	return map[string]any{
		"type":  "page",
		"title": title,
		"body":  FormSchema(def, apiURL, opts),
	}
}

// FormSchema generates an AMIS form schema (without the page wrapper).
func FormSchema(def *def.EntityDefinition, apiURL string, opts PageOpts) map[string]any {
	controls := formControls(def, opts)

	return map[string]any{
		"type": "form",
		"api": map[string]any{
			"method": "post",
			"url":    apiURL,
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

// ──────────────────────────────────────────────────────────────────
// Internal builders
// ──────────────────────────────────────────────────────────────────

func tableColumns(entity *def.EntityDefinition, opts PageOpts) []any {
	cols := []any{
		map[string]any{"name": "id", "label": "ID", "type": "text", "toggled": false},
	}
	for _, f := range allFields(entity, opts) {
		if opts.ExcludeFields[f.Name] || f.Hidden || f.IsSensitive {
			continue
		}
		cols = append(cols, ColumnDef(f))
	}
	return cols
}

// allFields returns def.Fields merged with opts.ExtraFields.
func allFields(entity *def.EntityDefinition, opts PageOpts) []*def.FieldDef {
	if len(opts.ExtraFields) == 0 {
		return entity.Fields
	}
	seen := make(map[string]struct{}, len(entity.Fields))
	for _, f := range entity.Fields {
		seen[f.Name] = struct{}{}
	}
	merged := make([]*def.FieldDef, len(entity.Fields), len(entity.Fields)+len(opts.ExtraFields))
	copy(merged, entity.Fields)
	for _, f := range opts.ExtraFields {
		if _, clash := seen[f.Name]; !clash {
			merged = append(merged, f)
		}
	}
	return merged
}

func formControls(entity *def.EntityDefinition, opts PageOpts) []any {
	// Hidden id field for edit mode.
	controls := []any{
		map[string]any{"type": "hidden", "name": "id"},
	}

	// Group by Section if any field has one.
	sections := groupBySection(entity, opts)
	if len(sections) > 1 || (len(sections) == 1 && sections[0].name != "") {
		for _, s := range sections {
			body := make([]any, len(s.fields))
			for i, f := range s.fields {
				body[i] = FormControl(f)
			}
			if s.name == "" {
				controls = append(controls, body...)
			} else {
				controls = append(controls, map[string]any{
					"type":  "fieldset",
					"title": s.name,
					"body":  body,
				})
			}
		}
	} else {
		for _, f := range allFields(entity, opts) {
			if opts.ExcludeFields[f.Name] || f.Hidden {
				continue
			}
			controls = append(controls, FormControl(f))
		}
	}
	return controls
}

type sectionGroup struct {
	name   string
	fields []*def.FieldDef
}

func groupBySection(entity *def.EntityDefinition, opts PageOpts) []sectionGroup {
	order := []string{}
	groups := map[string]*sectionGroup{}

	for _, f := range allFields(entity, opts) {
		if opts.ExcludeFields[f.Name] || f.Hidden {
			continue
		}
		sec := f.Section
		if _, ok := groups[sec]; !ok {
			order = append(order, sec)
			groups[sec] = &sectionGroup{name: sec}
		}
		groups[sec].fields = append(groups[sec].fields, f)
	}

	out := make([]sectionGroup, 0, len(order))
	for _, name := range order {
		out = append(out, *groups[name])
	}
	return out
}

func createButton(entity *def.EntityDefinition, opts PageOpts) map[string]any {
	label := "Create " + entity.Label
	if entity.Label == "" {
		label = "Create " + entity.Name
	}
	return map[string]any{
		"type":       "button",
		"label":      label,
		"level":      "primary",
		"actionType": "dialog",
		"dialog": map[string]any{
			"title": label,
			"body":  FormSchema(entity, opts.APIBase+"/"+entity.TableName(), opts),
		},
	}
}

func searchBar() map[string]any {
	return map[string]any{
		"type":        "search-box",
		"name":        "q",
		"placeholder": "Search...",
		"align":       "right",
	}
}

func actionColumn(apiURL string) map[string]any {
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
					"body": map[string]any{
						"type":    "form",
						"initApi": apiURL + "/${id}",
						"api": map[string]any{
							"method": "put",
							"url":    apiURL + "/${id}",
						},
					},
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
					"url":    apiURL + "/${id}",
				},
			},
		},
	}
}
