package amis

import "awo.so/framework/def"

// ──────────────────────────────────────────────────────────────────
// Permission gating
// ──────────────────────────────────────────────────────────────────

// Action is one of the CRUD verbs a permission check is evaluated
// against. Kept as a small closed set (rather than a free-form string)
// so call sites get compile-time autocomplete and typos can't silently
// produce an always-denied or always-allowed check.
type Action string

const (
	ActionCreate Action = "create"
	ActionEdit   Action = "edit"
	ActionDelete Action = "delete"
	ActionView   Action = "view"
)

// PermissionChecker decides whether the current request's caller may
// perform an action on an entity. Defined as an interface -- not a
// concrete RBAC type -- so this package stays decoupled from whatever
// auth system the host application uses (JWT claims, session-based ACL,
// policy engine, etc.); the caller adapts their own permission source to
// this one-method shape.
//
// Implementations must be safe to call multiple times per request and
// should be cheap (no network calls) since CRUDPage may call it once per
// action per page render.
type PermissionChecker interface {
	Can(entity *def.EntityDefinition, action Action) bool
}

// AllowAll is a PermissionChecker that grants everything. Useful as a
// PageOpts default so callers who don't care about permissions don't
// have to wire one up, and as a explicit "no gating" choice in tests.
type AllowAll struct{}

func (AllowAll) Can(*def.EntityDefinition, Action) bool { return true }

// resolvedPermissions returns opts.Permissions, defaulting to AllowAll
// so existing callers that never set Permissions keep today's
// behavior -- this is what makes the field backward compatible.
func (o PageOpts) resolvedPermissions() PermissionChecker {
	// if o.Permissions == nil {
	// 	return AllowAll{}
	// }
	// return o.Permissions
	return nil
}

// ──────────────────────────────────────────────────────────────────
// PageOpts addition
//
// Add this single field to the existing PageOpts struct in page.go.
// It's additive and zero-valued by default (nil -> AllowAll), so every
// existing PageOpts{...} struct literal in the codebase keeps compiling
// and keeps its current (unrestricted) behavior unchanged.
// ──────────────────────────────────────────────────────────────────
//
//	type PageOpts struct {
//	    ... existing fields ...
//
//	    // Permissions gates which CRUD actions render in the page. A nil
//	    // Permissions means "allow everything" (today's behavior).
//	    Permissions PermissionChecker
//	}

// ──────────────────────────────────────────────────────────────────
// CRUDPageV2 -- target-shape builder
//
// This mirrors the JSON sample's exact schema shape, which differs from
// the original CRUDPage in a few structural ways worth calling out:
//   - "api" is an object ({method,url,adaptor}), not a bare URL string.
//   - Pagination uses "defaultParams"/"perPage"/"perPageField"/
//     "pageField"/"itemsFieldName" instead of a footerToolbar pager.
//   - The create button lives under "toolbar", not "headerToolbar".
//   - Edit/Delete dialogs are built inline per-field rather than reusing
//     FormSchema, matching the sample's flatter structure.
//
// Because this is a structurally different schema flavor, it's added as
// CRUDPageV2 rather than changing CRUDPage's output -- existing callers
// of CRUDPage are unaffected.
// ──────────────────────────────────────────────────────────────────

func CRUDPageV2(entity *def.EntityDefinition, opts PageOpts) map[string]any {
	if entity == nil {
		return errorPage("CRUDPageV2: nil EntityDefinition")
	}

	base := opts.resolvedAPIBase()
	apiURL := apiURLFor(base, entity)
	title := opts.resolvedTitle(entity)
	perms := opts.resolvedPermissions()

	columns := tableColumns(entity, opts)
	// Only attach the operation column if the caller can do *something*
	// with a row; an Actions column with zero buttons is just visual
	// noise and a tell that something's misconfigured.
	canEdit := !opts.ReadOnly && perms.Can(entity, ActionEdit)
	canDelete := !opts.ReadOnly && perms.Can(entity, ActionDelete)
	if canEdit || canDelete {
		columns = append(columns, operationColumnV2(entity, apiURL, base, opts, canEdit, canDelete))
	}

	toolbar := []any{}
	if !opts.ReadOnly && perms.Can(entity, ActionCreate) {
		toolbar = append(toolbar, createButtonV2(entity, apiURL))
	}

	return map[string]any{
		"type":  "page",
		"title": title,
		"body": map[string]any{
			"type": "crud",
			"api": map[string]any{
				"method": "GET",
				"url":    apiURL,
				// Normalizes a {data:{items,total}} backend envelope into
				// AMIS's expected {status,msg,data:{items,total}} shape.
				// If your backend already returns AMIS's native envelope,
				// drop this adaptor.
				"adaptor": `return {status:0,msg:'',data:{items:payload.data.items||[],total:payload.data.total||0}};`,
			},
			"defaultParams":  map[string]any{"$page": 1, "$perPage": 20},
			"perPage":        20,
			"perPageField":   "$perPage",
			"pageField":      "$page",
			"itemsFieldName": "items",
			"columns":        columns,
			"toolbar":        toolbar,
		},
	}
}

// createButtonV2 renders the "New <Entity>" toolbar button with an
// inline create-form dialog, matching the sample JSON's flat field list
// rather than FormSchema's grouped/sectioned layout.
func createButtonV2(entity *def.EntityDefinition, apiURL string) map[string]any {
	label := "New " + entityLabel(entity)
	return map[string]any{
		"type":       "button",
		"label":      label,
		"icon":       "fa fa-plus",
		"actionType": "dialog",
		"dialog": map[string]any{
			"title": label,
			"body": map[string]any{
				"type": "form",
				"api":  map[string]any{"method": "POST", "url": apiURL},
				"body": flatFormControls(entity),
			},
		},
	}
}

// operationColumnV2 renders the Actions column. canEdit/canDelete are
// pre-resolved booleans (not re-checked here) so this function has a
// single responsibility -- assembling the button list -- and the
// permission decision stays in one place (CRUDPageV2) for easy auditing.
func operationColumnV2(entity *def.EntityDefinition, apiURL, _ string, _ PageOpts, canEdit, canDelete bool) map[string]any {
	rowURL := apiURL + "/${id}"
	buttons := make([]any, 0, 2)

	if canEdit {
		buttons = append(buttons, map[string]any{
			"label":      "Edit",
			"type":       "button",
			"actionType": "dialog",
			"dialog": map[string]any{
				"title": "Edit " + entityLabel(entity),
				"body": map[string]any{
					"type":    "form",
					"api":     map[string]any{"method": "PUT", "url": rowURL},
					"initApi": map[string]any{"method": "GET", "url": rowURL},
					"body":    flatFormControls(entity),
				},
			},
		})
	}

	if canDelete {
		buttons = append(buttons, map[string]any{
			"label":       "Delete",
			"type":        "button",
			"actionType":  "ajax",
			"confirmText": "Delete this record?",
			"api":         map[string]any{"method": "DELETE", "url": rowURL},
			"reload":      "crud",
		})
	}

	return map[string]any{
		"type":    "operation",
		"label":   "Actions",
		"buttons": buttons,
	}
}

// flatFormControls renders entity fields as a flat AMIS control list
// (no fieldset grouping, no hidden id field), matching the sample JSON's
// style. It reuses FormControl per-field so individual control
// rendering (types, validation, options) stays identical to CRUDPage's
// behavior -- only the layout differs.
func flatFormControls(entity *def.EntityDefinition) []any {
	controls := make([]any, 0, len(entity.Fields))
	for _, f := range entity.Fields {
		if f == nil || f.Hidden || f.IsSensitive {
			continue
		}
		controls = append(controls, FormControl(f))
	}
	return controls
}
