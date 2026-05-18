package ast

import (
	"fmt"

	"awo.so/internal/web/ui"
)

// ─── PageNode ─────────────────────────────────────────────────────────────────

// PageNode is the root node for every compiled page schema.
// Maps to AMIS type "page".
//
// Required: Title.
// Body is the main content area. AsideBody renders in the left column when set.
type PageNode struct {
	Title     string
	Body      []Node
	AsideBody []Node // optional left-side panel
	Toolbar   []Node // optional toolbar above body
	CSSClass  string
	// SubTitle is a smaller secondary heading rendered below Title.
	SubTitle string
	// Remark is a tooltip text placed next to the title.
	Remark string
}

func (p PageNode) NodeType() string { return "page" }

func (p PageNode) Validate() error {
	if p.Title == "" {
		return ErrRequiredField("page", "Title")
	}
	return nil
}

func (p PageNode) Compile() ui.M {
	m := ui.M{
		"type":  "page",
		"title": p.Title,
	}
	if p.SubTitle != "" {
		m["subTitle"] = p.SubTitle
	}
	if p.Remark != "" {
		m["remark"] = p.Remark
	}
	if p.CSSClass != "" {
		m["className"] = p.CSSClass
	}
	if len(p.Body) > 0 {
		m["body"] = compileNodes(p.Body)
	}
	if len(p.AsideBody) > 0 {
		m["aside"] = compileNodes(p.AsideBody)
	}
	if len(p.Toolbar) > 0 {
		m["toolbar"] = compileNodes(p.Toolbar)
	}
	return m
}

func (p PageNode) Children() []Node {
	return mergeNodeSlices(p.Body, p.AsideBody, p.Toolbar)
}

var _ Node = PageNode{}
var _ ContainerNode = PageNode{}

// ─── GridNode ─────────────────────────────────────────────────────────────────

// GridNode renders children in a CSS-grid layout.
// Maps to AMIS type "grid".
//
// Required: at least one Column.
type GridNode struct {
	Columns []GridColumn
	// Gap controls spacing between columns. Valid values: "xs" | "sm" | "md" | "lg" | "none".
	Gap string
}

// GridColumn is one column in a GridNode.
type GridColumn struct {
	// Body is the content of the column. Must have at least one node.
	Body []Node
	// MD is the Bootstrap-style column width (1–12). 0 means auto.
	MD int
}

func (g GridNode) NodeType() string { return "grid" }

func (g GridNode) Validate() error {
	if len(g.Columns) == 0 {
		return ErrRequiredField("grid", "Columns")
	}
	for i, col := range g.Columns {
		if len(col.Body) == 0 {
			return ErrInvalidField("grid", "Columns", formatColumnErr(i, "Body must not be empty"))
		}
		if col.MD < 0 || col.MD > 12 {
			return ErrInvalidField("grid", "Columns", formatColumnErr(i, "MD must be 0–12"))
		}
	}
	return nil
}

func (g GridNode) Compile() ui.M {
	cols := make(ui.A, 0, len(g.Columns))
	for _, col := range g.Columns {
		c := ui.M{"body": compileNodes(col.Body)}
		if col.MD > 0 {
			c["md"] = col.MD
		}
		cols = append(cols, c)
	}
	m := ui.M{
		"type":    "grid",
		"columns": cols,
	}
	if g.Gap != "" {
		m["gap"] = g.Gap
	}
	return m
}

func (g GridNode) Children() []Node {
	var all []Node
	for _, col := range g.Columns {
		all = append(all, col.Body...)
	}
	return all
}

var _ Node = GridNode{}
var _ ContainerNode = GridNode{}

// ─── FlexNode ─────────────────────────────────────────────────────────────────

// FlexNode renders children using CSS flexbox.
// Maps to AMIS type "flex".
//
// Required: at least one item in Items.
type FlexNode struct {
	Items     []Node
	Direction string // "row" (default) | "column"
	Justify   string // flex justify-content: "flex-start" | "center" | "flex-end" | "space-between" | "space-around"
	Align     string // flex align-items: "flex-start" | "center" | "flex-end" | "stretch"
	Gap       string // spacing between items: "xs" | "sm" | "md" | "lg"
	Wrap      bool   // allow items to wrap
}

func (f FlexNode) NodeType() string { return "flex" }

func (f FlexNode) Validate() error {
	if len(f.Items) == 0 {
		return ErrRequiredField("flex", "Items")
	}
	if f.Direction != "" && f.Direction != "row" && f.Direction != "column" {
		return ErrInvalidField("flex", "Direction", "must be \"row\" or \"column\"")
	}
	return nil
}

func (f FlexNode) Compile() ui.M {
	m := ui.M{
		"type":  "flex",
		"items": compileNodes(f.Items),
	}
	if f.Direction != "" {
		m["direction"] = f.Direction
	}
	if f.Justify != "" {
		m["justify"] = f.Justify
	}
	if f.Align != "" {
		m["alignItems"] = f.Align
	}
	if f.Gap != "" {
		m["gap"] = f.Gap
	}
	if f.Wrap {
		m["wrap"] = true
	}
	return m
}

func (f FlexNode) Children() []Node { return f.Items }

var _ Node = FlexNode{}
var _ ContainerNode = FlexNode{}

// ─── TabsNode ─────────────────────────────────────────────────────────────────

// TabsNode renders content in a tabbed interface.
// Maps to AMIS type "tabs".
//
// Required: at least one Tab.
type TabsNode struct {
	Tabs []Tab
	// Mode controls tab visual style: "line" (default) | "card" | "radio" | "tiled"
	Mode string
	// Mountable controls whether inactive tab bodies are mounted in the DOM.
	// false = lazy-mount (better performance for large schemas).
	Mountable bool
}

// Tab is a single tab item within a TabsNode.
type Tab struct {
	// Title is the tab label. Required.
	Title string
	// Body is the content rendered when the tab is active.
	Body []Node
	// Hash is the URL fragment used for deep-linking (e.g. "#details").
	Hash string
	// Icon is an optional icon class (e.g. "fa fa-user").
	Icon string
	// VisibleOn is a boolean AMIS expression controlling tab visibility.
	// Must not contain IAM keywords — enforced by ValidateStage.
	VisibleOn string
}

func (t TabsNode) NodeType() string { return "tabs" }

func (t TabsNode) Validate() error {
	if len(t.Tabs) == 0 {
		return ErrRequiredField("tabs", "Tabs")
	}
	for i, tab := range t.Tabs {
		if tab.Title == "" {
			return ErrInvalidField("tabs", "Tabs", formatColumnErr(i, "Title must not be empty"))
		}
	}
	return nil
}

func (t TabsNode) Compile() ui.M {
	tabs := make(ui.A, 0, len(t.Tabs))
	for _, tab := range t.Tabs {
		item := ui.M{"title": tab.Title}
		if len(tab.Body) > 0 {
			item["body"] = compileNodes(tab.Body)
		}
		if tab.Hash != "" {
			item["hash"] = tab.Hash
		}
		if tab.Icon != "" {
			item["icon"] = tab.Icon
		}
		if tab.VisibleOn != "" {
			item["visibleOn"] = tab.VisibleOn
		}
		tabs = append(tabs, item)
	}
	m := ui.M{
		"type": "tabs",
		"tabs": tabs,
	}
	if t.Mode != "" {
		m["mode"] = t.Mode
	}
	if t.Mountable {
		m["mountable"] = true
	}
	return m
}

func (t TabsNode) Children() []Node {
	var all []Node
	for _, tab := range t.Tabs {
		all = append(all, tab.Body...)
	}
	return all
}

var _ Node = TabsNode{}
var _ ContainerNode = TabsNode{}

// ─── SplitPaneNode ────────────────────────────────────────────────────────────

// SplitPaneNode renders two side-by-side panels.
// Maps to AMIS type "grid" with two columns, but semantically signals a
// master-detail or left-right split — used in detail view pages.
//
// Required: both Left and Right must be non-nil.
type SplitPaneNode struct {
	Left      Node
	Right     Node
	LeftWidth int // Bootstrap MD col width for left pane (1–11, default 4)
}

func (s SplitPaneNode) NodeType() string { return "split_pane" }

func (s SplitPaneNode) Validate() error {
	if s.Left == nil {
		return ErrRequiredField("split_pane", "Left")
	}
	if s.Right == nil {
		return ErrRequiredField("split_pane", "Right")
	}
	w := s.LeftWidth
	if w != 0 && (w < 1 || w > 11) {
		return ErrInvalidField("split_pane", "LeftWidth", "must be 1–11")
	}
	return nil
}

func (s SplitPaneNode) Compile() ui.M {
	leftMD := s.LeftWidth
	if leftMD == 0 {
		leftMD = 4
	}
	rightMD := 12 - leftMD
	return ui.M{
		"type": "grid",
		"columns": ui.A{
			ui.M{"md": leftMD, "body": s.Left.Compile()},
			ui.M{"md": rightMD, "body": s.Right.Compile()},
		},
	}
}

func (s SplitPaneNode) Children() []Node {
	nodes := make([]Node, 0, 2)
	if s.Left != nil {
		nodes = append(nodes, s.Left)
	}
	if s.Right != nil {
		nodes = append(nodes, s.Right)
	}
	return nodes
}

var _ Node = SplitPaneNode{}
var _ ContainerNode = SplitPaneNode{}

// ─── SectionNode ─────────────────────────────────────────────────────────────

// SectionNode is a collapsible, labelled content group.
// Maps to AMIS type "collapse". Used extensively in document forms to group
// related fields (e.g. "Address Details", "Payment Terms").
//
// Required: Title and at least one node in Body.
type SectionNode struct {
	Title     string
	Body      []Node
	Collapsed bool   // initial collapsed state
	CSSClass  string
}

func (s SectionNode) NodeType() string { return "collapse" }

func (s SectionNode) Validate() error {
	if s.Title == "" {
		return ErrRequiredField("collapse", "Title")
	}
	if len(s.Body) == 0 {
		return ErrRequiredField("collapse", "Body")
	}
	return nil
}

func (s SectionNode) Compile() ui.M {
	m := ui.M{
		"type":      "collapse",
		"header":    s.Title,
		"collapsed": s.Collapsed,
		"body":      compileNodes(s.Body),
	}
	if s.CSSClass != "" {
		m["className"] = s.CSSClass
	}
	return m
}

func (s SectionNode) Children() []Node { return s.Body }

var _ Node = SectionNode{}
var _ ContainerNode = SectionNode{}

// ─── helpers ──────────────────────────────────────────────────────────────────

// compileNodes emits a []Node as an AMIS array.
// Used by all container node Compile() implementations.
func compileNodes(nodes []Node) ui.A {
	out := make(ui.A, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.Compile())
	}
	return out
}

// mergeNodeSlices concatenates multiple node slices without allocating when empty.
func mergeNodeSlices(slices ...[]Node) []Node {
	var total int
	for _, s := range slices {
		total += len(s)
	}
	if total == 0 {
		return nil
	}
	merged := make([]Node, 0, total)
	for _, s := range slices {
		merged = append(merged, s...)
	}
	return merged
}

// formatColumnErr formats an index-based error suffix for column/tab field errors.
func formatColumnErr(index int, reason string) string {
	return "index " + itoa(index) + ": " + reason
}

// itoa is a zero-allocation int-to-string for small indices.
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	// fallback for larger indices — not performance-critical
	return fmt.Sprintf("%d", n)
}
