// Package widget defines the WidgetTree intermediate representation (IR) used
// by the SDUI framework.
//
// The WidgetTree sits between the SDUI generator and any renderer. The amis
// renderer converts a WidgetTree to amis JSON; future renderers may target
// React Native, PDF, or accessibility trees without touching the generator.
//
// ADR-006: The Node struct, NodeKind constants, DataSource, and ActionNode are
// public framework contracts frozen at v1.0. They must not be changed without
// an ADR revision.
//
// This package has no internal dependencies — only the standard library.
package widget

// Node is a single widget in the WidgetTree.
// Nodes form a tree via Children. Each Node has a semantic Kind that
// renderers translate to renderer-specific types.
//
// Normative requirements (ADR-006):
//   - NodeKind values MUST be one of the defined constants.
//   - Node.Hidden: true MUST NOT be used for permission-gating — absent nodes
//     are used instead (hidden elements can be revealed by client manipulation;
//     absent elements cannot).
//   - Node.Props MUST be used only for renderer-specific properties not
//     covered by the typed fields.
//   - The WidgetTree MUST be constructable without importing awo/sdui/amis.
type Node struct {
	// Kind is the semantic widget type (required).
	Kind NodeKind

	// ID is the optional stable identifier for this node.
	// Used by renderers for reference (e.g., dialog target IDs).
	ID string

	// Label is the human-readable display text.
	Label string

	// Name is the field binding name.
	// For form fields: the field name in EntityRecord.Data.
	// For lists: not used.
	Name string

	// Required marks a form field as required in the UI.
	// Does not duplicate server-side validation — both apply independently.
	Required bool

	// ReadOnly marks a form field as non-editable.
	ReadOnly bool

	// Hidden marks a node as absent from the rendered output.
	//
	// IMPORTANT: Do not use Hidden for permission-gating. Absent nodes (not
	// included in the tree) must be used instead. Hidden is reserved for
	// internal bookkeeping fields that are API-accessible but not shown in UI.
	Hidden bool

	// Props contains renderer-specific properties not expressible in the typed
	// fields above. Props are merged last and override computed defaults.
	// Use sparingly — prefer typed fields where possible.
	Props map[string]any

	// Children are the nested nodes (for page, form, section, tabs, table).
	Children []*Node

	// DataSource configures remote data fetching for list and select nodes.
	DataSource *DataSource

	// Actions are the action buttons associated with this node.
	Actions []*ActionNode
}

// NodeKind is the semantic widget type. Renderers map each NodeKind to a
// renderer-specific component type (e.g., amis component type strings).
type NodeKind string

const (
	// ── Structural ───────────────────────────────────────────────────────────

	// NodePage is the top-level page container.
	NodePage NodeKind = "page"

	// NodeForm is a form container. Children are the form fields.
	// DataSource provides the submit API endpoint.
	NodeForm NodeKind = "form"

	// NodeList is a paginated record list (amis crud2).
	// DataSource provides the data API endpoint. Children are column nodes.
	NodeList NodeKind = "list"

	// NodeSection is a visual grouping within a form (amis group).
	NodeSection NodeKind = "section"

	// NodeTabs is tabbed navigation. Each child NodeSection becomes one tab.
	NodeTabs NodeKind = "tabs"

	// NodeTable is an inline child record table.
	// DataSource provides the sub-list API endpoint.
	NodeTable NodeKind = "table"

	// NodeDialog is a modal dialog.
	NodeDialog NodeKind = "dialog"

	// ── Input fields ─────────────────────────────────────────────────────────

	// NodeField is a generic single-line text field (used for DynamicLink).
	NodeField NodeKind = "field"

	// NodeText is a single-line text input (FieldTypeData).
	NodeText NodeKind = "text"

	// NodeTextArea is a multi-line text input (FieldTypeLongText, SmallText).
	NodeTextArea NodeKind = "textarea"

	// NodeNumber is a numeric input (FieldTypeInt, Float, Currency).
	NodeNumber NodeKind = "number"

	// NodeSelect is a dropdown or autocomplete (FieldTypeSelect, Link).
	NodeSelect NodeKind = "select"

	// NodeDate is a date picker (FieldTypeDate). Format: YYYY-MM-DD.
	NodeDate NodeKind = "date"

	// NodeDateTime is a datetime picker (FieldTypeDateTime, FieldTypeTime).
	NodeDateTime NodeKind = "datetime"

	// NodeSwitch is a toggle switch (FieldTypeBool).
	NodeSwitch NodeKind = "switch"

	// NodeEditor is a JSON editor (FieldTypeJSON).
	NodeEditor NodeKind = "editor"

	// ── Interactive ──────────────────────────────────────────────────────────

	// NodeButton is an action trigger rendered as a button.
	NodeButton NodeKind = "button"
)

// DataSource configures remote data fetching for NodeList and NodeSelect nodes
// that use server-side search.
type DataSource struct {
	// URL is the API endpoint.
	// May contain amis-style template variables: ${keywords}, ${page}, ${tenant_id}
	URL string

	// Method is the HTTP method (default: "GET").
	Method string

	// SendOn is a condition expression for conditional data fetching.
	// Empty means always fetch.
	SendOn string

	// LabelField is the field name used as the display label in select options.
	LabelField string

	// ValueField is the field name used as the select option value (default: "id").
	ValueField string
}

// ActionNode represents an action button associated with a Node (typically on
// list rows, detail views, or form toolbars).
type ActionNode struct {
	// Label is the button display text.
	Label string

	// ActionType is the semantic type: "submit", "dialog", "link", "ajax".
	ActionType string

	// Level controls the visual prominence: "primary", "default", "warning", "danger".
	Level string

	// Href is the navigation target for "link" type actions.
	Href string

	// API is the endpoint for "ajax" type actions.
	API string

	// ConfirmText is the confirmation dialog message shown before execution.
	// Empty means no confirmation required.
	ConfirmText string
}
