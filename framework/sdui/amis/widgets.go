package amis

import (
	"fmt"

	"awo.so/framework/def"
)

// ──────────────────────────────────────────────────────────────────
// widget.go
//
// This file adds small, composable AMIS "widget" builders that sit one
// layer above the page/form/table builders in page.go and field.go.
// Where page.go answers "render this whole CRUD page" and field.go
// answers "render this one field", widget.go answers "render this one
// reusable dashboard component" -- a stat card, a quick filter chip
// strip, or a chart -- that callers can drop into a custom page layout
// alongside (or instead of) a full CRUDPage.
//
// Design goals, in priority order:
//
//  1. Never panic on bad input. Every exported builder here is reachable
//     from request-time metadata (entity definitions, registry-driven
//     custom fields) the same way page.go's builders are, so the same
//     nil-safety discipline applies.
//  2. Reuse existing helpers (entityLabel, fieldLabel, apiURLFor,
//     resolvedAPIBase) instead of re-deriving titles/URLs, so a future
//     change to e.g. API-base resolution only has to happen in one place.
//  3. Keep each widget's AMIS schema self-contained: a widget builder
//     returns a single map[string]any that can be embedded directly into
//     a "body" array, with no required surrounding wrapper.
// ──────────────────────────────────────────────────────────────────

// WidgetOpts configures the optional, widget-specific rendering knobs
// that don't belong on the more general-purpose PageOpts. It is a
// separate type (rather than extending PageOpts) so adding widget
// features never risks changing the meaning of existing PageOpts-based
// call sites -- consistent with how PageOpts itself avoids breaking
// changes by only ever growing new zero-valued fields.
type WidgetOpts struct {
	// APIBase is the REST prefix, e.g. "/api". Defaults to "/api",
	// resolved the same way as PageOpts.APIBase for consistency.
	APIBase string

	// RefreshSeconds, if > 0, makes the widget poll its data source on
	// an interval (AMIS "interval" property). Zero means no auto-refresh.
	RefreshSeconds int

	// Color overrides the widget's accent color. Empty uses the AMIS
	// theme default.
	Color string
}

// resolvedAPIBase mirrors PageOpts.resolvedAPIBase; duplicated narrowly
// here (rather than sharing a type) because WidgetOpts is intentionally
// decoupled from PageOpts -- see the WidgetOpts doc comment.
func (o WidgetOpts) resolvedAPIBase() string {
	if o.APIBase == "" {
		return "/api"
	}
	return o.APIBase
}

// StatCardWidget renders a single AMIS "stat" card summarizing a count
// or aggregate for an entity, e.g. "Total Users: 1,204". The value is
// fetched live from the entity's collection endpoint using AMIS's
// built-in data mapping rather than being computed server-side here,
// keeping this package's responsibility limited to schema generation
// (no business logic, no live API calls at schema-build time).
//
// field is optional: an empty string renders a row-count stat ("count");
// a non-empty field name renders a sum aggregate over that numeric field.
// fieldLabel mirrors the field on the card title for readability.
func StatCardWidget(entity *def.EntityDefinition, field string, opts WidgetOpts) map[string]any {
	if entity == nil {
		return errorWidget("StatCardWidget: nil EntityDefinition")
	}

	base := opts.resolvedAPIBase()
	apiURL := apiURLFor(base, entity)

	title := entityLabel(entity)
	valueExpr := "${count}"
	if field != "" {
		title = fmt.Sprintf("%s — %s", title, field)
		valueExpr = fmt.Sprintf("${sum_%s}", field)
	}

	schema := map[string]any{
		"type":  "panel",
		"title": title,
		"body": map[string]any{
			"type":      "tpl",
			"tpl":       valueExpr,
			"className": "text-2xl font-bold",
		},
		// initApi drives the card's data: AMIS fetches once on mount
		// (and again per RefreshSeconds if set) and exposes the response
		// body fields (e.g. "count", "sum_<field>") to the tpl above.
		"initApi": statAPIURL(apiURL, field),
	}
	if opts.Color != "" {
		schema["style"] = map[string]any{"borderTopColor": opts.Color}
	}
	if opts.RefreshSeconds > 0 {
		schema["interval"] = opts.RefreshSeconds * 1000 // AMIS expects ms.
	}
	return schema
}

// statAPIURL builds the aggregate endpoint queried by StatCardWidget.
// Kept as its own function so the query-string convention (a single
// "?aggregate=" parameter) lives in exactly one place and can evolve
// without touching StatCardWidget's schema-assembly logic.
func statAPIURL(collectionURL, field string) string {
	if field == "" {
		return collectionURL + "?aggregate=count"
	}
	return fmt.Sprintf("%s?aggregate=sum&field=%s", collectionURL, field)
}

// QuickFilterWidget renders a row of clickable filter chips for a
// FieldTypeSelect/FieldTypeMultiSelect field, letting a dashboard page
// offer one-click filtering (e.g. status = "open") without embedding a
// full CRUD table's built-in filter form.
//
// targetCRUDID must match the "id" of the crud schema node (set via
// AMIS's standard "id" property on the CRUDPage's body) that this widget
// should filter; QuickFilterWidget does not set that id itself, since
// page composition -- and therefore id uniqueness across a page -- is the
// caller's responsibility.
//
// Returns an errorWidget (never nil, never panics) if f is nil or is not
// a select-like field with Options, since chips have nothing to render
// in that case and a silently-empty widget is harder to debug than an
// explicit inline error.
func QuickFilterWidget(f *def.FieldDef, targetCRUDID string) map[string]any {
	if f == nil {
		return errorWidget("QuickFilterWidget: nil FieldDef")
	}
	if f.Type != def.FieldTypeSelect && f.Type != def.FieldTypeMultiSelect {
		return errorWidget(fmt.Sprintf("QuickFilterWidget: field %q is not a select field", f.Name))
	}
	if len(f.Options) == 0 {
		return errorWidget(fmt.Sprintf("QuickFilterWidget: field %q has no options", f.Name))
	}

	buttons := make([]any, 0, len(f.Options)+1)
	buttons = append(buttons, filterChip("All", targetCRUDID, f.Name, ""))
	for _, opt := range f.Options {
		buttons = append(buttons, filterChip(opt, targetCRUDID, f.Name, opt))
	}

	return map[string]any{
		"type": "button-group",
		"name": f.Name + "_quick_filter",
		"body": buttons,
	}
}

// filterChip builds a single AMIS button that re-queries the named CRUD
// component with fieldName=value (or clears the filter when value is "").
func filterChip(label, targetCRUDID, fieldName, value string) map[string]any {
	query := map[string]any{fieldName: value}
	return map[string]any{
		"type":       "button",
		"label":      label,
		"actionType": "reload",
		"target":     targetCRUDID,
		"args":       map[string]any{"query": query},
	}
}

// ChartWidget renders a simple AMIS line/bar chart sourced from an
// entity's collection endpoint, grouping by groupByField and aggregating
// valueField. It is intentionally minimal (one series, one aggregate) --
// callers needing multi-series or custom ECharts configs should build
// the "chart" schema node directly rather than extending this builder
// with an ever-growing options struct.
//
// kind selects the chart shape; unrecognized values fall back to "line"
// rather than producing an invalid AMIS schema.
func ChartWidget(entity *def.EntityDefinition, groupByField, valueField, kind string, opts WidgetOpts) map[string]any {
	if entity == nil {
		return errorWidget("ChartWidget: nil EntityDefinition")
	}
	if groupByField == "" || valueField == "" {
		return errorWidget("ChartWidget: groupByField and valueField are required")
	}

	chartType := "line"
	if kind == "bar" {
		chartType = "bar"
	}

	base := opts.resolvedAPIBase()
	apiURL := apiURLFor(base, entity)

	schema := map[string]any{
		"type": "chart",
		"api":  fmt.Sprintf("%s?group_by=%s&value=%s", apiURL, groupByField, valueField),
		"config": map[string]any{
			"xField": groupByField,
			"yField": valueField,
			"type":   chartType,
		},
	}
	if opts.RefreshSeconds > 0 {
		schema["interval"] = opts.RefreshSeconds * 1000
	}
	return schema
}

// errorWidget mirrors page.go's errorPage but as an embeddable schema
// node (no "page" wrapper), so a malformed widget call surfaces as a
// visible inline alert inside whatever layout it was embedded in,
// instead of either panicking or silently rendering nothing.
func errorWidget(msg string) map[string]any {
	return map[string]any{
		"type":  "alert",
		"level": "danger",
		"body":  msg,
	}
}
