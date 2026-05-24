package ast

import (
	"fmt"
	"strings"

	"awo.so/internal/web/ui"
)

// validAPIMethods is the set of HTTP methods accepted by APISpec.
var validAPIMethods = map[string]bool{
	"get": true, "post": true, "put": true, "patch": true, "delete": true,
}

// ─── APISpec ──────────────────────────────────────────────────────────────────

// APISpec describes an AMIS API call.
// Compile() emits the canonical "method:url" string form.
//
// Required: Method (one of get/post/put/patch/delete) and URL.
type APISpec struct {
	Method  string            // "get" | "post" | "put" | "patch" | "delete"
	URL     string            // e.g. "/api/v1/finance/invoices"
	Headers map[string]string // optional extra request headers
	// Data is merged into the request body/params (POST/PUT only).
	Data map[string]any
	// ResponseData transforms the API response before AMIS consumes it.
	ResponseData map[string]any
	// SendOn is a boolean AMIS expression; the API call is skipped when false.
	SendOn string
}

// Validate checks that Method and URL are set and that Method is recognised.
func (a APISpec) Validate(ownerType string) error {
	if a.URL == "" {
		return ErrAPIURLEmpty(ownerType)
	}
	m := strings.ToLower(a.Method)
	if m == "" {
		return ErrAPIMethodInvalid(ownerType, "(empty)")
	}
	if !validAPIMethods[m] {
		return ErrAPIMethodInvalid(ownerType, a.Method)
	}
	return nil
}

// Compile emits the AMIS API object form so that headers and data can be included.
// When no headers/data/sendOn are set, a plain "method:url" string is sufficient for
// AMIS — callers may use CompileString() in that case.
func (a APISpec) Compile() any {
	if len(a.Headers) == 0 && len(a.Data) == 0 && len(a.ResponseData) == 0 && a.SendOn == "" {
		return a.CompileString()
	}
	obj := ui.M{
		"method": strings.ToLower(a.Method),
		"url":    a.URL,
	}
	if len(a.Headers) > 0 {
		obj["headers"] = a.Headers
	}
	if len(a.Data) > 0 {
		obj["data"] = a.Data
	}
	if len(a.ResponseData) > 0 {
		obj["responseData"] = a.ResponseData
	}
	if a.SendOn != "" {
		obj["sendOn"] = a.SendOn
	}
	return obj
}

// CompileString emits the "method:url" shorthand.
func (a APISpec) CompileString() string {
	return strings.ToLower(a.Method) + ":" + a.URL
}

// ─── TableColumn ──────────────────────────────────────────────────────────────

// TableColumn describes one column in a TableNode or CRUDNode.
type TableColumn struct {
	// Name is the data field key. Required.
	Name string
	// Label is the column header text. Defaults to Name when empty.
	Label string
	// Type is the AMIS column type: "text" (default) | "date" | "number" |
	// "currency" | "mapping" | "image" | "link" | "operation"
	Type string
	// Map holds value→HTML entries for Type "mapping" columns.
	// Each value must be a non-empty string; HTML templates are supported.
	// Ignored when Type is not "mapping".
	Map map[string]string
	// Width in pixels. 0 = auto.
	Width int
	// Sortable enables server-side sort for this column.
	Sortable bool
	// Fixed pins the column: "" | "left" | "right"
	Fixed string
	// VisibleOn is a boolean AMIS expression controlling column visibility.
	VisibleOn string
	// QuickEdit enables inline editing for this column.
	QuickEdit bool
}

func (c TableColumn) compile() ui.M {
	m := ui.M{"name": c.Name}
	label := c.Label
	if label == "" {
		label = c.Name
	}
	m["label"] = label
	if c.Type != "" {
		m["type"] = c.Type
	}
	if c.Type == "mapping" && len(c.Map) > 0 {
		// Convert map[string]string to map[string]any for JSON serialisation.
		mapped := make(map[string]any, len(c.Map))
		for k, v := range c.Map {
			mapped[k] = v
		}
		m["map"] = mapped
	}
	if c.Width > 0 {
		m["width"] = c.Width
	}
	if c.Sortable {
		m["sortable"] = true
	}
	if c.Fixed != "" {
		m["fixed"] = c.Fixed
	}
	if c.VisibleOn != "" {
		m["visibleOn"] = c.VisibleOn
	}
	if c.QuickEdit {
		m["quickEdit"] = true
	}
	return m
}

// ─── TableNode ────────────────────────────────────────────────────────────────

// TableNode renders a static or API-sourced data table.
// Maps to AMIS type "table".
//
// Required: at least one Column.
// Use CRUDNode for tables with pagination, filtering, and row actions.
type TableNode struct {
	Source  string        // data variable path (e.g. "${items}") — mutually exclusive with API
	API     *APISpec      // remote data source — mutually exclusive with Source
	Columns []TableColumn // Required: at least one
	// Title is an optional table caption.
	Title string
	// Striped enables alternating row colours.
	Striped bool
	// Bordered renders cell borders.
	Bordered bool
}

func (t TableNode) NodeType() string { return "table" }

func (t TableNode) Validate() error {
	if len(t.Columns) == 0 {
		return ErrRequiredField("table", "Columns")
	}
	for i, col := range t.Columns {
		if col.Name == "" {
			return ErrInvalidField("table", "Columns", fmt.Sprintf("index %d: Name must not be empty", i))
		}
	}
	if t.API != nil {
		if err := t.API.Validate("table"); err != nil {
			return err
		}
	}
	return nil
}

func (t TableNode) Compile() ui.M {
	cols := make(ui.A, 0, len(t.Columns))
	for _, c := range t.Columns {
		cols = append(cols, c.compile())
	}
	m := ui.M{
		"type":    "table",
		"columns": cols,
	}
	if t.Title != "" {
		m["title"] = t.Title
	}
	if t.Source != "" {
		m["source"] = t.Source
	}
	if t.API != nil {
		m["api"] = t.API.Compile()
	}
	if t.Striped {
		m["striped"] = true
	}
	if t.Bordered {
		m["bordered"] = true
	}
	return m
}

var _ Node = TableNode{}

// ─── CRUDNode ─────────────────────────────────────────────────────────────────

// CRUDNode is the primary listing/data-management node for ERP listing pages.
// Maps to AMIS type "crud".
//
// Structural invariants enforced at compile time (not by NormalizeStage):
//   - syncLocation is always true — emitted unconditionally.
//   - API.Method and API.URL are required.
//
// Required: API and at least one Column.
type CRUDNode struct {
	API        APISpec
	Columns    []TableColumn
	// Filter is the filter form shown above the table. nil = no filter.
	// Accepts any Node — typically FilterBarNode or FormNode.
	Filter     Node
	// Toolbar nodes appear in the top-right of the CRUD header.
	Toolbar    []Node
	// BulkActions appear when rows are selected.
	BulkActions []Node
	// RowActions appear in an "operation" column on each row.
	RowActions  []ActionNode
	// PrimaryKey is the field used for row identity (default: "id").
	PrimaryKey  string
	// PageSize is the default number of rows per page (default: 20).
	PageSize    int
	// Title is an optional heading above the CRUD component.
	Title       string
}

func (c CRUDNode) NodeType() string { return "crud" }

func (c CRUDNode) Validate() error {
	if err := c.API.Validate("crud"); err != nil {
		return err
	}
	if len(c.Columns) == 0 {
		return ErrRequiredField("crud", "Columns")
	}
	for i, col := range c.Columns {
		if col.Name == "" {
			return ErrInvalidField("crud", "Columns", fmt.Sprintf("index %d: Name must not be empty", i))
		}
	}
	return nil
}

func (c CRUDNode) Compile() ui.M {
	cols := make(ui.A, 0, len(c.Columns))
	for _, col := range c.Columns {
		cols = append(cols, col.compile())
	}

	// Append operation column when row actions are defined.
	if len(c.RowActions) > 0 {
		actions := make(ui.A, 0, len(c.RowActions))
		for _, a := range c.RowActions {
			actions = append(actions, a.Compile())
		}
		cols = append(cols, ui.M{
			"type":    "operation",
			"label":   "Actions",
			"buttons": actions,
		})
	}

	m := ui.M{
		"type":         "crud",
		"syncLocation": true, // structural invariant — always emitted
		"api":          c.API.Compile(),
		"columns":      cols,
	}

	pk := c.PrimaryKey
	if pk == "" {
		pk = "id"
	}
	m["primaryField"] = pk

	ps := c.PageSize
	if ps == 0 {
		ps = 20
	}
	m["perPage"] = ps

	if c.Title != "" {
		m["title"] = c.Title
	}
	if c.Filter != nil {
		m["filter"] = c.Filter.Compile()
	}
	if len(c.Toolbar) > 0 {
		m["toolbar"] = compileNodes(c.Toolbar)
	}
	if len(c.BulkActions) > 0 {
		m["bulkActions"] = compileNodes(c.BulkActions)
	}

	return m
}

func (c CRUDNode) Children() []Node {
	var all []Node
	all = append(all, c.Toolbar...)
	all = append(all, c.BulkActions...)
	if c.Filter != nil && c.Filter != Node(nil) {
		all = append(all, c.Filter)
	}
	// Include RowActions so CompileTree validates them (e.g. dialog/drawer children).
	for _, ra := range c.RowActions {
		all = append(all, ra)
	}
	return all
}

var _ Node = CRUDNode{}
var _ ContainerNode = CRUDNode{}

// ─── ChartNode ────────────────────────────────────────────────────────────────

// ChartNode renders an ECharts chart.
// Maps to AMIS type "chart".
//
// Structural invariant: style.background is always "transparent" — emitted
// unconditionally to ensure dark-mode compatibility.
//
// Required: Config must have at least one key.
type ChartNode struct {
	// Config is the ECharts option object.
	Config map[string]any
	// API is an optional remote data source. When set, AMIS fetches data and
	// passes it to Config via the replaceChartOption mechanism.
	API    *APISpec
	Height int    // pixel height. Default: 300.
	Width  string // CSS width, e.g. "100%". Default: "100%".
}

func (c ChartNode) NodeType() string { return "chart" }

func (c ChartNode) Validate() error {
	if len(c.Config) == 0 {
		return ErrRequiredField("chart", "Config")
	}
	if c.API != nil {
		if err := c.API.Validate("chart"); err != nil {
			return err
		}
	}
	return nil
}

func (c ChartNode) Compile() ui.M {
	h := c.Height
	if h == 0 {
		h = 300
	}
	w := c.Width
	if w == "" {
		w = "100%"
	}
	m := ui.M{
		"type":   "chart",
		"config": c.Config,
		"style": ui.M{
			"background": "transparent", // structural invariant — always emitted
			"height":     fmt.Sprintf("%dpx", h),
			"width":      w,
		},
	}
	if c.API != nil {
		m["api"] = c.API.Compile()
	}
	return m
}

var _ Node = ChartNode{}

// ─── CardNode ─────────────────────────────────────────────────────────────────

// CardNode renders a card (panel with header and body).
// Maps to AMIS type "panel".
//
// Required: at least one node in Body.
type CardNode struct {
	Title    string
	SubTitle string
	Body     []Node
	Footer   []Node
	CSSClass string
	// Collapsible allows the card body to be toggled.
	Collapsible bool
}

func (c CardNode) NodeType() string { return "panel" }

func (c CardNode) Validate() error {
	if len(c.Body) == 0 {
		return ErrRequiredField("panel", "Body")
	}
	return nil
}

func (c CardNode) Compile() ui.M {
	m := ui.M{
		"type": "panel",
		"body": compileNodes(c.Body),
	}
	if c.Title != "" {
		m["title"] = c.Title
	}
	if c.SubTitle != "" {
		m["subTitle"] = c.SubTitle
	}
	if len(c.Footer) > 0 {
		m["footer"] = compileNodes(c.Footer)
	}
	if c.CSSClass != "" {
		m["className"] = c.CSSClass
	}
	if c.Collapsible {
		m["collapsible"] = true
	}
	return m
}

func (c CardNode) Children() []Node {
	return mergeNodeSlices(c.Body, c.Footer)
}

var _ Node = CardNode{}
var _ ContainerNode = CardNode{}

// ─── StatNode ─────────────────────────────────────────────────────────────────

// StatNode renders a single KPI metric with label, value, and optional trend.
// Maps to AMIS type "statistic" (custom wrapper around AMIS tpl for ERP use).
//
// Required: Label and ValueKey.
type StatNode struct {
	// Label is the metric name (e.g. "Total Revenue").
	Label string
	// ValueKey is the data key holding the metric value (e.g. "revenue").
	ValueKey string
	// Format controls value display: "number" (default) | "currency" | "percent"
	Format string
	// Currency code for "currency" format (e.g. "KES"). Defaults to tenant currency.
	Currency string
	// TrendKey is the data key for the trend value (positive = up, negative = down).
	TrendKey string
	// TrendMode: "up_is_good" (green up) | "down_is_good" (green down, for costs/debt)
	TrendMode string
	// IconClass is an optional Font Awesome icon class (e.g. "fa fa-dollar").
	IconClass string
}

func (s StatNode) NodeType() string { return "tpl" }

func (s StatNode) Validate() error {
	if s.Label == "" {
		return ErrRequiredField("stat", "Label")
	}
	if s.ValueKey == "" {
		return ErrRequiredField("stat", "ValueKey")
	}
	return nil
}

func (s StatNode) Compile() ui.M {
	// Emit as AMIS "tpl" — the CSS class drives the ERP stat card styling.
	// Values are bound via AMIS data expressions.
	tpl := fmt.Sprintf(`<div class="erp-stat-card">`)
	if s.IconClass != "" {
		tpl += fmt.Sprintf(`<i class="%s erp-stat-icon"></i>`, s.IconClass)
	}
	tpl += fmt.Sprintf(`<div class="erp-stat-label">%s</div>`, s.Label)
	tpl += fmt.Sprintf(`<div class="erp-stat-value">${%s}</div>`, s.ValueKey)
	if s.TrendKey != "" {
		tpl += fmt.Sprintf(`<div class="erp-stat-trend erp-stat-trend--%s">${%s}</div>`,
			s.TrendMode, s.TrendKey)
	}
	tpl += `</div>`

	return ui.M{
		"type": "tpl",
		"tpl":  tpl,
	}
}

var _ Node = StatNode{}

// ─── TimelineNode ─────────────────────────────────────────────────────────────

// TimelineNode renders a vertical timeline of events.
// Maps to AMIS type "timeline".
//
// Required: either Items or API.
type TimelineNode struct {
	// Items is a static list of timeline entries.
	Items []TimelineItem
	// API fetches timeline entries dynamically.
	API *APISpec
	// Direction: "vertical" (default) | "horizontal"
	Direction string
}

// TimelineItem is a single entry in a TimelineNode.
type TimelineItem struct {
	Time    string // display time label
	Title   string
	Detail  string
	Color   string // dot colour: "green" | "red" | "blue" | "grey" | custom hex
	Icon    string // optional icon class
}

func (t TimelineNode) NodeType() string { return "timeline" }

func (t TimelineNode) Validate() error {
	if len(t.Items) == 0 && t.API == nil {
		return ErrRequiredField("timeline", "Items or API")
	}
	if t.API != nil {
		if err := t.API.Validate("timeline"); err != nil {
			return err
		}
	}
	return nil
}

func (t TimelineNode) Compile() ui.M {
	m := ui.M{"type": "timeline"}
	if len(t.Items) > 0 {
		items := make(ui.A, 0, len(t.Items))
		for _, item := range t.Items {
			entry := ui.M{"time": item.Time, "title": item.Title}
			if item.Detail != "" {
				entry["detail"] = item.Detail
			}
			if item.Color != "" {
				entry["color"] = item.Color
			}
			if item.Icon != "" {
				entry["icon"] = item.Icon
			}
			items = append(items, entry)
		}
		m["items"] = items
	}
	if t.API != nil {
		m["api"] = t.API.Compile()
	}
	if t.Direction != "" {
		m["direction"] = t.Direction
	}
	return m
}

var _ Node = TimelineNode{}

// ─── TreeNode ─────────────────────────────────────────────────────────────────

// TreeNode renders a collapsible tree view.
// Maps to AMIS type "tree".
//
// Required: API (tree data is almost always remote in ERP contexts).
type TreeNode struct {
	API           APISpec
	LabelField    string // field for node display label (default: "label")
	ValueField    string // field for node value (default: "value")
	ChildrenField string // field for child nodes (default: "children")
	Multiple      bool   // allow multi-select
	Cascade       bool   // parent selection cascades to children
}

func (t TreeNode) NodeType() string { return "tree" }

func (t TreeNode) Validate() error {
	return t.API.Validate("tree")
}

func (t TreeNode) Compile() ui.M {
	m := ui.M{
		"type": "tree",
		"source": t.API.Compile(),
	}
	lf := t.LabelField
	if lf == "" {
		lf = "label"
	}
	vf := t.ValueField
	if vf == "" {
		vf = "value"
	}
	cf := t.ChildrenField
	if cf == "" {
		cf = "children"
	}
	m["labelField"] = lf
	m["valueField"] = vf
	m["childrenField"] = cf
	if t.Multiple {
		m["multiple"] = true
	}
	if t.Cascade {
		m["cascade"] = true
	}
	return m
}

var _ Node = TreeNode{}

// ─── MappingNode ──────────────────────────────────────────────────────────────

// MappingNode renders a value mapped to a display string (typically styled HTML).
// Maps to AMIS type "mapping".
//
// Use for status badges, label chips, and any enum field that needs colour coding.
// Each entry in Map is: data-value → HTML string (e.g. badge markup).
// A "*" key acts as the catch-all fallback when no key matches the data value.
//
// Required: Name and at least one entry in Map.
type MappingNode struct {
	// Name is the data field key whose value is looked up in Map.
	Name      string
	// Label is the field label (used in form layout or table header).
	Label     string
	// Map is the value→display mapping. Values are HTML strings.
	// Example: {"draft": "<span class='badge badge-warning'>Draft</span>", "*": "${value}"}
	Map       map[string]string
	// VisibleOn is a boolean AMIS expression controlling visibility.
	VisibleOn string
}

func (m MappingNode) NodeType() string { return "mapping" }

func (m MappingNode) Validate() error {
	if m.Name == "" {
		return ErrRequiredField("mapping", "Name")
	}
	if len(m.Map) == 0 {
		return ErrRequiredField("mapping", "Map")
	}
	return nil
}

func (m MappingNode) Compile() ui.M {
	mapped := make(map[string]any, len(m.Map))
	for k, v := range m.Map {
		mapped[k] = v
	}
	out := ui.M{
		"type": "mapping",
		"name": m.Name,
		"map":  mapped,
	}
	if m.Label != "" {
		out["label"] = m.Label
	}
	if m.VisibleOn != "" {
		out["visibleOn"] = m.VisibleOn
	}
	return out
}

var _ Node = MappingNode{}

// ─── PropertyNode ─────────────────────────────────────────────────────────────

// PropertyNode renders a read-only key-value description list.
// Maps to AMIS type "property".
//
// Use for document detail views, record summary panels, and any read-only
// field group. Each item is a label + AMIS expression string.
//
// Required: at least one Item.
type PropertyNode struct {
	// Title is an optional panel heading rendered above the property grid.
	Title string
	// Column is the number of label-value pairs per row (default: 3).
	Column int
	// Items are the label-value pairs. Content is an AMIS expression string.
	// Example: PropertyItem{Label: "Amount", Content: "${amount|number}"}
	Items []PropertyItem
}

// PropertyItem is one label-value row in a PropertyNode.
type PropertyItem struct {
	// Label is the field name displayed on the left.
	Label string
	// Content is an AMIS expression rendered on the right.
	// Plain data keys: "${ref_number}"
	// With format filters: "${amount|number}", "${date|date:YYYY-MM-DD}", "${rate|percent}"
	Content string
}

func (p PropertyNode) NodeType() string { return "property" }

func (p PropertyNode) Validate() error {
	if len(p.Items) == 0 {
		return ErrRequiredField("property", "Items")
	}
	for i, item := range p.Items {
		if item.Label == "" {
			return ErrInvalidField("property", "Items", formatColumnErr(i, "Label must not be empty"))
		}
		if item.Content == "" {
			return ErrInvalidField("property", "Items", formatColumnErr(i, "Content must not be empty"))
		}
	}
	return nil
}

func (p PropertyNode) Compile() ui.M {
	items := make(ui.A, 0, len(p.Items))
	for _, item := range p.Items {
		items = append(items, ui.M{
			"label":   item.Label,
			"content": item.Content,
		})
	}
	col := p.Column
	if col <= 0 {
		col = 3
	}
	m := ui.M{
		"type":   "property",
		"column": col,
		"items":  items,
	}
	if p.Title != "" {
		m["title"] = p.Title
	}
	return m
}

var _ Node = PropertyNode{}

// ─── ActionNode (display helper) ─────────────────────────────────────────────
// ActionNode is defined here (before form.go) because CRUDNode uses it in
// RowActions. The full form-oriented ActionNode definition lives in form.go.
// This is the shared canonical definition.

// ActionNode renders a button or link that triggers an AMIS action.
// Maps to AMIS type "button".
//
// Required: Label and ActionType.
// For "dialog" ActionType: set Dialog (inline) or Target (named reference).
// For "drawer" ActionType: set Drawer (inline) or Target (named reference).
type ActionNode struct {
	Label       string
	// ActionType: "ajax" | "dialog" | "drawer" | "link" | "submit" | "reset" |
	//             "copy" | "download" | "reload" | "close"
	ActionType  string
	// API is required when ActionType is "ajax".
	API         *APISpec
	// Dialog is an inline dialog definition for ActionType "dialog".
	// Takes precedence over Target when both are set.
	Dialog      *DialogNode
	// Drawer is an inline drawer definition for ActionType "drawer".
	// Takes precedence over Target when both are set.
	Drawer      *DrawerNode
	// Target is a named dialog/drawer reference or URL for link types.
	// Use Dialog/Drawer fields instead for inline definitions.
	Target      string
	// Level controls button colour: "primary" | "success" | "warning" | "danger" |
	//                               "info" | "default" | "link"
	Level       string
	// Size: "xs" | "sm" | "md" | "lg"
	Size        string
	// VisibleOn is a boolean AMIS expression controlling visibility.
	VisibleOn   string
	// DisabledOn is a boolean AMIS expression controlling disabled state.
	DisabledOn  string
	// ConfirmText shows a confirmation dialog before executing the action.
	ConfirmText string
	// Icon is an optional Font Awesome class (e.g. "fa fa-check").
	Icon        string
}

func (a ActionNode) NodeType() string { return "button" }

func (a ActionNode) Validate() error {
	if a.Label == "" {
		return ErrRequiredField("button", "Label")
	}
	if a.ActionType == "" {
		return ErrRequiredField("button", "ActionType")
	}
	if a.ActionType == "ajax" && a.API == nil {
		return ErrInvalidField("button", "API", "required when ActionType is \"ajax\"")
	}
	if a.ActionType == "dialog" && a.Dialog == nil && a.Target == "" {
		return ErrInvalidField("button", "Dialog", "required when ActionType is \"dialog\" (or set Target for named reference)")
	}
	if a.ActionType == "drawer" && a.Drawer == nil && a.Target == "" {
		return ErrInvalidField("button", "Drawer", "required when ActionType is \"drawer\" (or set Target for named reference)")
	}
	if a.API != nil {
		if err := a.API.Validate("button"); err != nil {
			return err
		}
	}
	if a.Dialog != nil {
		if err := a.Dialog.Validate(); err != nil {
			return err
		}
	}
	if a.Drawer != nil {
		if err := a.Drawer.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a ActionNode) Compile() ui.M {
	m := ui.M{
		"type":       "button",
		"label":      a.Label,
		"actionType": a.ActionType,
	}
	if a.API != nil {
		m["api"] = a.API.Compile()
	}
	// Inline dialog/drawer take precedence over Target string reference.
	if a.Dialog != nil {
		m["dialog"] = a.Dialog.Compile()
	} else if a.Drawer != nil {
		m["drawer"] = a.Drawer.Compile()
	} else if a.Target != "" {
		m["target"] = a.Target
	}
	if a.Level != "" {
		m["level"] = a.Level
	}
	if a.Size != "" {
		m["size"] = a.Size
	}
	if a.VisibleOn != "" {
		m["visibleOn"] = a.VisibleOn
	}
	if a.DisabledOn != "" {
		m["disabledOn"] = a.DisabledOn
	}
	if a.ConfirmText != "" {
		m["confirmText"] = a.ConfirmText
	}
	if a.Icon != "" {
		m["icon"] = a.Icon
	}
	return m
}

var _ Node = ActionNode{}
