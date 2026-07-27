// Package amis converts a WidgetTree ([]*widget.Node) into amis JSON
// (map[string]any) ready for serialisation and delivery to the browser.
//
// ADR-006: The amis renderer is one possible renderer for the WidgetTree IR.
// It MUST NOT accept raw field definitions or entity definitions — its sole
// input is always widget.Node values.
//
// Normative requirements:
//   - Render returns an error for any unknown NodeKind.
//   - Node.Hidden: true causes the node to be omitted from output entirely.
//   - Node.Props is merged last and overrides all computed defaults.
//   - Renderer does not import awo/def, awo/registry, or awo/compiler.
//   - Output is deterministic for the same input tree.
package amis

import (
	"fmt"

	"awo.so/awo/sdui/widget"
)

// Renderer converts a WidgetTree root node into an amis JSON schema.
// The returned map is safe for JSON serialisation via encoding/json.
type Renderer interface {
	// Render translates root and all its descendants into an amis page schema.
	// Returns an error if root is nil or any node contains an unknown NodeKind.
	Render(root *widget.Node) (map[string]any, error)
}

// DefaultRenderer is the production implementation of Renderer.
// It handles all NodeKind constants defined in awo/sdui/widget.
// Construct via New().
type DefaultRenderer struct{}

// New returns a DefaultRenderer ready for use.
func New() *DefaultRenderer {
	return &DefaultRenderer{}
}

// Render translates root and all its descendants into an amis page schema.
// Unknown NodeKind values cause Render to return an error — they are never
// silently dropped or rendered as empty divs.
func (r *DefaultRenderer) Render(root *widget.Node) (map[string]any, error) {
	if root == nil {
		return nil, fmt.Errorf("amis.Render: root node must not be nil")
	}
	return r.renderNode(root)
}

// renderNode converts a single Node and its subtree to an amis schema object.
func (r *DefaultRenderer) renderNode(n *widget.Node) (map[string]any, error) {
	if n == nil {
		return nil, fmt.Errorf("amis.renderNode: nil node in tree")
	}

	// Hidden nodes are entirely omitted — not rendered with hidden:true.
	// (See WIDGET_TREE_SPEC.md §7 — permission-gated elements must be absent.)
	if n.Hidden {
		return nil, nil
	}

	var out map[string]any
	var err error

	switch n.Kind {
	case widget.NodePage:
		out, err = r.renderPage(n)
	case widget.NodeForm:
		out, err = r.renderForm(n)
	case widget.NodeList:
		out, err = r.renderList(n)
	case widget.NodeSection:
		out, err = r.renderSection(n)
	case widget.NodeTabs:
		out, err = r.renderTabs(n)
	case widget.NodeTable:
		out, err = r.renderTable(n)
	case widget.NodeDialog:
		out, err = r.renderDialog(n)
	case widget.NodeButton:
		out, err = r.renderButton(n)
	case widget.NodeText, widget.NodeField:
		out, err = r.renderText(n)
	case widget.NodeTextArea:
		out, err = r.renderTextArea(n)
	case widget.NodeNumber:
		out, err = r.renderNumber(n)
	case widget.NodeSelect:
		out, err = r.renderSelect(n)
	case widget.NodeDate:
		out, err = r.renderDate(n)
	case widget.NodeDateTime:
		out, err = r.renderDateTime(n)
	case widget.NodeSwitch:
		out, err = r.renderSwitch(n)
	case widget.NodeEditor:
		out, err = r.renderEditor(n)
	default:
		return nil, fmt.Errorf("amis.renderNode: unknown NodeKind %q — must be one of the defined widget.NodeKind constants", n.Kind)
	}

	if err != nil {
		return nil, err
	}

	// Merge Node.Props last — they override all computed defaults.
	for k, v := range n.Props {
		out[k] = v
	}

	// Set stable ID if present.
	if n.ID != "" {
		out["id"] = n.ID
	}

	return out, nil
}

// renderChildren converts a slice of Nodes into a slice of amis schema objects.
// Nil results (from Hidden nodes) are omitted.
func (r *DefaultRenderer) renderChildren(nodes []*widget.Node) ([]any, error) {
	var out []any
	for _, child := range nodes {
		m, err := r.renderNode(child)
		if err != nil {
			return nil, err
		}
		if m != nil { // nil == hidden
			out = append(out, m)
		}
	}
	return out, nil
}

// renderActions converts ActionNode values to amis button schema objects.
func (r *DefaultRenderer) renderActions(actions []*widget.ActionNode) []any {
	out := make([]any, 0, len(actions))
	for _, a := range actions {
		if a == nil {
			continue
		}
		btn := map[string]any{
			"type":       "button",
			"label":      a.Label,
			"actionType": a.ActionType,
			"level":      a.Level,
		}
		if a.Href != "" {
			btn["link"] = a.Href
		}
		if a.API != "" {
			btn["api"] = a.API
		}
		if a.ConfirmText != "" {
			btn["confirmText"] = a.ConfirmText
		}
		out = append(out, btn)
	}
	return out
}

// applyCommon sets name, label, required, disabled on a field schema.
func applyCommon(n *widget.Node, out map[string]any) {
	if n.Name != "" {
		out["name"] = n.Name
	}
	if n.Label != "" {
		out["label"] = n.Label
	}
	if n.Required {
		out["required"] = true
	}
	if n.ReadOnly {
		out["disabled"] = true
	}
}

// ── Structural renderers ──────────────────────────────────────────────────────

func (r *DefaultRenderer) renderPage(n *widget.Node) (map[string]any, error) {
	body, err := r.renderChildren(n.Children)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"type": "page",
	}
	if n.Label != "" {
		out["title"] = n.Label
	}
	if len(body) > 0 {
		out["body"] = body
	}
	if acts := r.renderActions(n.Actions); len(acts) > 0 {
		out["toolbar"] = acts
	}
	return out, nil
}

func (r *DefaultRenderer) renderForm(n *widget.Node) (map[string]any, error) {
	body, err := r.renderChildren(n.Children)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"type": "form",
		"body": body,
	}
	if n.Label != "" {
		out["title"] = n.Label
	}
	if n.DataSource != nil && n.DataSource.URL != "" {
		out["api"] = buildAPI(n.DataSource)
		out["initApi"] = n.DataSource.Method + ":" + n.DataSource.URL + "/${id}"
	}
	if acts := r.renderActions(n.Actions); len(acts) > 0 {
		out["actions"] = acts
	}
	return out, nil
}

func (r *DefaultRenderer) renderList(n *widget.Node) (map[string]any, error) {
	columns, err := r.renderListColumns(n.Children)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"type":    "crud2",
		"columns": columns,
	}
	if n.DataSource != nil && n.DataSource.URL != "" {
		out["api"] = buildAPI(n.DataSource)
	}
	if n.Label != "" {
		out["title"] = n.Label
	}
	if acts := r.renderActions(n.Actions); len(acts) > 0 {
		out["toolbar"] = acts
	}
	return out, nil
}

// renderListColumns converts column child nodes to amis column descriptor objects.
// Column amis types are determined by NodeKind (AMIS_RENDERER_SPEC.md §6).
func (r *DefaultRenderer) renderListColumns(nodes []*widget.Node) ([]any, error) {
	var cols []any
	for _, child := range nodes {
		if child == nil || child.Hidden {
			continue
		}
		colType := nodeKindToColumnType(child.Kind)
		col := map[string]any{
			"name":  child.Name,
			"label": child.Label,
			"type":  colType,
		}
		if child.Kind == widget.NodeDate {
			col["format"] = "YYYY-MM-DD"
		}
		if child.Kind == widget.NodeDateTime {
			col["format"] = "YYYY-MM-DDTHH:mm:ssZ"
		}
		// Merge any column-level Props.
		for k, v := range child.Props {
			col[k] = v
		}
		cols = append(cols, col)
	}
	return cols, nil
}

// nodeKindToColumnType maps NodeKind to the amis column type for list views.
func nodeKindToColumnType(kind widget.NodeKind) string {
	switch kind {
	case widget.NodeNumber:
		return "tpl"
	case widget.NodeDate:
		return "date"
	case widget.NodeDateTime:
		return "datetime"
	case widget.NodeSwitch:
		return "status"
	case widget.NodeSelect:
		return "mapping"
	default:
		return "text"
	}
}

func (r *DefaultRenderer) renderSection(n *widget.Node) (map[string]any, error) {
	body, err := r.renderChildren(n.Children)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"type": "group",
		"body": body,
	}
	if n.Label != "" {
		out["label"] = n.Label
	}
	return out, nil
}

func (r *DefaultRenderer) renderTabs(n *widget.Node) (map[string]any, error) {
	var tabs []any
	for _, child := range n.Children {
		if child == nil || child.Hidden {
			continue
		}
		body, err := r.renderChildren(child.Children)
		if err != nil {
			return nil, err
		}
		tab := map[string]any{
			"title": child.Label,
			"body":  body,
		}
		tabs = append(tabs, tab)
	}
	return map[string]any{
		"type": "tabs",
		"tabs": tabs,
	}, nil
}

func (r *DefaultRenderer) renderTable(n *widget.Node) (map[string]any, error) {
	cols, err := r.renderListColumns(n.Children)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"type":    "table",
		"columns": cols,
	}
	if n.DataSource != nil && n.DataSource.URL != "" {
		out["source"] = n.DataSource.URL
	}
	if n.Label != "" {
		out["label"] = n.Label
	}
	return out, nil
}

func (r *DefaultRenderer) renderDialog(n *widget.Node) (map[string]any, error) {
	body, err := r.renderChildren(n.Children)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"type": "dialog",
		"body": body,
	}
	if n.Label != "" {
		out["title"] = n.Label
	}
	return out, nil
}

// ── Field renderers ───────────────────────────────────────────────────────────

func (r *DefaultRenderer) renderButton(n *widget.Node) (map[string]any, error) {
	out := map[string]any{
		"type":  "button",
		"label": n.Label,
	}
	if n.ReadOnly {
		out["disabled"] = true
	}
	return out, nil
}

func (r *DefaultRenderer) renderText(n *widget.Node) (map[string]any, error) {
	out := map[string]any{"type": "input-text"}
	applyCommon(n, out)
	return out, nil
}

func (r *DefaultRenderer) renderTextArea(n *widget.Node) (map[string]any, error) {
	out := map[string]any{"type": "textarea"}
	applyCommon(n, out)
	return out, nil
}

func (r *DefaultRenderer) renderNumber(n *widget.Node) (map[string]any, error) {
	out := map[string]any{"type": "input-number"}
	applyCommon(n, out)
	return out, nil
}

func (r *DefaultRenderer) renderSelect(n *widget.Node) (map[string]any, error) {
	out := map[string]any{"type": "select"}
	applyCommon(n, out)
	if n.DataSource != nil && n.DataSource.URL != "" {
		src := map[string]any{
			"url":    n.DataSource.URL,
			"method": methodOrDefault(n.DataSource.Method),
		}
		if n.DataSource.SendOn != "" {
			src["sendOn"] = n.DataSource.SendOn
		}
		out["source"] = src
		if n.DataSource.LabelField != "" {
			out["labelField"] = n.DataSource.LabelField
		}
		vf := n.DataSource.ValueField
		if vf == "" {
			vf = "id"
		}
		out["valueField"] = vf
	}
	return out, nil
}

func (r *DefaultRenderer) renderDate(n *widget.Node) (map[string]any, error) {
	out := map[string]any{
		"type":   "input-date",
		"format": "YYYY-MM-DD",
	}
	applyCommon(n, out)
	return out, nil
}

func (r *DefaultRenderer) renderDateTime(n *widget.Node) (map[string]any, error) {
	out := map[string]any{
		"type":   "input-datetime",
		"format": "YYYY-MM-DDTHH:mm:ssZ",
	}
	applyCommon(n, out)
	return out, nil
}

func (r *DefaultRenderer) renderSwitch(n *widget.Node) (map[string]any, error) {
	out := map[string]any{"type": "switch"}
	applyCommon(n, out)
	return out, nil
}

func (r *DefaultRenderer) renderEditor(n *widget.Node) (map[string]any, error) {
	out := map[string]any{"type": "json-editor"}
	applyCommon(n, out)
	return out, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// buildAPI converts a DataSource into an amis api descriptor.
func buildAPI(ds *widget.DataSource) map[string]any {
	api := map[string]any{
		"url":    ds.URL,
		"method": methodOrDefault(ds.Method),
	}
	if ds.SendOn != "" {
		api["sendOn"] = ds.SendOn
	}
	return api
}

func methodOrDefault(method string) string {
	if method == "" {
		return "GET"
	}
	return method
}
