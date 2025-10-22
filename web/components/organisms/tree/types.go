package tree

import (
	"github.com/niiniyare/erp/web/components/atoms"
)

// TreeNode represents a single node in the tree structure
type TreeNode struct {
	// Core node properties
	ID    string `json:"id"`
	Label string `json:"label"`
	Value any    `json:"value,omitempty"`

	// Hierarchical structure
	Children []TreeNode `json:"children,omitempty"`
	ParentID string     `json:"parentId,omitempty"`
	Path     []string   `json:"path,omitempty"`
	Level    int        `json:"level,omitempty"`

	// Node state
	Expanded    bool `json:"expanded,omitempty"`
	Selected    bool `json:"selected,omitempty"`
	Disabled    bool `json:"disabled,omitempty"`
	Hidden      bool `json:"hidden,omitempty"`
	HasChildren bool `json:"hasChildren,omitempty"`
	IsLeaf      bool `json:"isLeaf,omitempty"`

	// Visual properties
	Icon          string `json:"icon,omitempty"`
	ExpandedIcon  string `json:"expandedIcon,omitempty"`
	CollapsedIcon string `json:"collapsedIcon,omitempty"`
	Badge         string `json:"badge,omitempty"`
	Description   string `json:"description,omitempty"`

	// Navigation and interaction
	URL string `json:"url,omitempty"`

	// Column data - stores values for each column
	Data map[string]interface{} `json:"data,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TreeColumn defines a column in multi-column tree view
type TreeColumn struct {
	// Core properties
	Field string `json:"field"`
	Label string `json:"label"`

	// Layout
	Width    string `json:"width,omitempty"`    // e.g., "40%", "200px", "1fr"
	MinWidth string `json:"minWidth,omitempty"` // e.g., "100px"
	Align    string `json:"align,omitempty"`    // left, center, right

	// Data formatting
	Type   string `json:"type,omitempty"`   // text, number, currency, percentage, date, custom
	Format string `json:"format,omitempty"` // format string for numbers/dates

	// Aggregation for parent nodes
	Aggregation *TreeAggregation `json:"aggregation,omitempty"`

	// Features
	Sortable bool   `json:"sortable"`
	Render   string `json:"render,omitempty"` // custom template/function

	// Styling
	ClassName       string `json:"className,omitempty"`
	HeaderClassName string `json:"headerClassName,omitempty"`

	// Visibility
	Visible   bool   `json:"visible"`
	VisibleOn string `json:"visibleOn,omitempty"`
}

// TreeAggregation defines how to aggregate values in parent nodes
type TreeAggregation struct {
	Enabled            bool   `json:"enabled"`
	Method             string `json:"method"` // sum, average, count, min, max, custom
	ExcludeParentValue bool   `json:"excludeParentValue"`
	DisplayMode        string `json:"displayMode"` // replace, append, both
	AppendFormat       string `json:"appendFormat,omitempty"`
	Recursive          bool   `json:"recursive"`
	CustomFunction     string `json:"customFunction,omitempty"`
}

// TreeConfig defines tree configuration options
type TreeConfig struct {
	// Selection behavior
	Selectable  bool `json:"selectable"`
	MultiSelect bool `json:"multiSelect"`
	Checkboxes  bool `json:"checkboxes"`
	Cascade     bool `json:"cascade"`

	// Expansion behavior
	Expandable      bool     `json:"expandable"`
	Accordion       bool     `json:"accordion"`
	ShowRoot        bool     `json:"showRoot"`
	ExpandAll       bool     `json:"expandAll"`
	CollapseAll     bool     `json:"collapseAll"`
	DefaultExpanded []string `json:"defaultExpanded,omitempty"`
	DefaultSelected []string `json:"defaultSelected,omitempty"`

	// Visual configuration
	ShowLines    bool `json:"showLines"`
	ShowIcons    bool `json:"showIcons"`
	ShowTooltips bool `json:"showTooltips"`

	// Interaction features
	Draggable  bool `json:"draggable"`
	Sortable   bool `json:"sortable"`
	Searchable bool `json:"searchable"`
	Filterable bool `json:"filterable"`

	// Performance options
	Virtual  bool `json:"virtual"`
	LazyLoad bool `json:"lazyLoad"`

	// Layout options
	MaxHeight  string `json:"maxHeight,omitempty"`
	Indent     string `json:"indent,omitempty"`
	NodeHeight string `json:"nodeHeight,omitempty"`

	// Column-specific options
	ShowHeaders      bool `json:"showHeaders"`
	HeaderSticky     bool `json:"headerSticky"`
	ColumnResizable  bool `json:"columnResizable"`
	ShowColumnBorder bool `json:"showColumnBorder"`
}

// TreeAction represents an action available for tree nodes
type TreeAction struct {
	// Core action properties
	Text    string        `json:"text"`
	Icon    string        `json:"icon,omitempty"`
	Variant atoms.Variant `json:"variant"`
	Size    atoms.Size    `json:"size"`

	// Event handling
	OnClick string `json:"onclick,omitempty"`

	// HTMX attributes
	HxPost    string `json:"hxPost,omitempty"`
	HxGet     string `json:"hxGet,omitempty"`
	HxDelete  string `json:"hxDelete,omitempty"`
	HxTarget  string `json:"hxTarget,omitempty"`
	HxSwap    string `json:"hxSwap,omitempty"`
	HxConfirm string `json:"hxConfirm,omitempty"`

	// Action scope and visibility
	ShowWhen  *atoms.SchemaExpression `json:"showWhen,omitempty"`
	VisibleOn string                  `json:"visibleOn,omitempty"` // "hover", "always", "selected"
	Disabled  bool                    `json:"disabled,omitempty"`

	// Styling and accessibility
	Tooltip  string `json:"tooltip,omitempty"`
	ID       string `json:"id,omitempty"`
	Class    string `json:"className,omitempty"`
	Position string `json:"position,omitempty"` // left, right
}

// TreeSearchConfig defines search configuration options
type TreeSearchConfig struct {
	Placeholder      string   `json:"placeholder"`
	Debounce         int      `json:"debounce"`
	MinLength        int      `json:"minLength"`
	Clearable        bool     `json:"clearable"`
	HighlightMatches bool     `json:"highlightMatches"`
	ExpandMatches    bool     `json:"expandMatches"`
	CaseSensitive    bool     `json:"caseSensitive"`
	SearchFields     []string `json:"searchFields,omitempty"`
}

// TreeStyling defines visual styling options
type TreeStyling struct {
	Variant  atoms.Variant `json:"variant"` // default, dark, light, minimal
	Size     atoms.Size    `json:"size"`    // xs, sm, md, lg, xl
	Rounded  bool          `json:"rounded"`
	Shadow   bool          `json:"shadow"`
	Bordered bool          `json:"bordered"`
	Striped  bool          `json:"striped"`
	Hover    bool          `json:"hover"`
}

// TreeLoadingConfig defines loading state configuration
type TreeLoadingConfig struct {
	Show          bool   `json:"show"`
	Text          string `json:"text"`
	Skeleton      bool   `json:"skeleton"`
	SkeletonCount int    `json:"skeletonCount"`
}

// TreeEmptyConfig defines empty state configuration
type TreeEmptyConfig struct {
	Message          string `json:"message"`
	Icon             string `json:"icon,omitempty"`
	ShowCreateButton bool   `json:"showCreateButton"`
	CreateText       string `json:"createText,omitempty"`
	CreateAction     string `json:"createAction,omitempty"`
}

// TreeProps defines properties for the Tree component
type TreeProps struct {
	// Base properties (ID, class, etc.)
	atoms.BaseProps

	// Accessibility support
	atoms.AccessibilityProps

	// Tree structure and data
	Data []TreeNode `json:"data"`

	// Column definitions for multi-column view
	Columns []TreeColumn `json:"columns,omitempty"`

	// Configuration
	Config TreeConfig `json:"config"`

	// Actions
	Actions     []TreeAction `json:"actions,omitempty"`
	NodeActions []TreeAction `json:"nodeActions,omitempty"`

	// Search configuration
	SearchProps *TreeSearchConfig `json:"searchProps,omitempty"`

	// Server-side processing
	DataUrl     string `json:"dataUrl,omitempty"`
	LazyLoadUrl string `json:"lazyLoadUrl,omitempty"`
	SearchUrl   string `json:"searchUrl,omitempty"`

	// HTMX attributes
	HxGet     string `json:"hxGet,omitempty"`
	HxPost    string `json:"hxPost,omitempty"`
	HxTarget  string `json:"hxTarget,omitempty"`
	HxSwap    string `json:"hxSwap,omitempty"`
	HxTrigger string `json:"hxTrigger,omitempty"`

	// Styling and states
	Styling TreeStyling       `json:"styling,omitempty"`
	Loading TreeLoadingConfig `json:"loading,omitempty"`
	Empty   TreeEmptyConfig   `json:"empty,omitempty"`

	// Tree metadata
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// Event handlers
	OnNodeClick    string `json:"onNodeClick,omitempty"`
	OnNodeSelect   string `json:"onNodeSelect,omitempty"`
	OnNodeExpand   string `json:"onNodeExpand,omitempty"`
	OnNodeCollapse string `json:"onNodeCollapse,omitempty"`
	OnSearch       string `json:"onSearch,omitempty"`
	OnDrop         string `json:"onDrop,omitempty"`

	// Schema expression support
	HiddenOn   *atoms.SchemaExpression `json:"hiddenOn,omitempty"`
	VisibleOn  *atoms.SchemaExpression `json:"visibleOn,omitempty"`
	DisabledOn *atoms.SchemaExpression `json:"disabledOn,omitempty"`

	// Static display mode
	Static            bool                    `json:"static"`
	StaticOn          *atoms.SchemaExpression `json:"staticOn,omitempty"`
	StaticPlaceholder string                  `json:"staticPlaceholder,omitempty"`
	StaticClassName   string                  `json:"staticClassName,omitempty"`

	// Mobile UI optimization
	UseMobileUI bool `json:"useMobileUI,omitempty"`
}

// ============================================================================
// CONSTRUCTOR FUNCTIONS
// ============================================================================

// NewTreeNode creates a new tree node with sensible defaults
func NewTreeNode(id, label string) TreeNode {
	return TreeNode{
		ID:          id,
		Label:       label,
		Level:       0,
		Expanded:    false,
		Selected:    false,
		Disabled:    false,
		Hidden:      false,
		HasChildren: false,
		IsLeaf:      true,
		Data:        make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}
}

// NewTreeColumn creates a new column with sensible defaults
func NewTreeColumn(field, label string) TreeColumn {
	return TreeColumn{
		Field:    field,
		Label:    label,
		Type:     "text",
		Align:    "left",
		Visible:  true,
		Sortable: false,
	}
}

// NewTreeAction creates a new tree action with sensible defaults
func NewTreeAction(text, icon string) TreeAction {
	return TreeAction{
		Text:      text,
		Icon:      icon,
		Variant:   atoms.VariantDefault,
		Size:      atoms.SizeSM,
		VisibleOn: "hover",
	}
}

// NewTreeProps creates a new Tree with sensible defaults
func NewTreeProps(id string) TreeProps {
	return TreeProps{
		BaseProps: atoms.BaseProps{
			ID: atoms.EnsureID(id, "tree"),
		},
		Config: TreeConfig{
			Selectable:       false,
			MultiSelect:      false,
			Checkboxes:       false,
			Cascade:          true,
			Expandable:       true,
			Accordion:        false,
			ShowRoot:         true,
			ShowLines:        true,
			ShowIcons:        true,
			ShowTooltips:     false,
			Draggable:        false,
			Sortable:         false,
			Searchable:       false,
			Filterable:       false,
			Virtual:          false,
			LazyLoad:         false,
			Indent:           "1.5rem",
			NodeHeight:       "2.5rem",
			ShowHeaders:      true,
			HeaderSticky:     false,
			ColumnResizable:  false,
			ShowColumnBorder: false,
		},
		Styling: TreeStyling{
			Variant:  atoms.VariantDefault,
			Size:     atoms.SizeMD,
			Rounded:  true,
			Shadow:   false,
			Bordered: false,
			Striped:  false,
			Hover:    true,
		},
		Loading: TreeLoadingConfig{
			Show:          false,
			Text:          "Loading tree...",
			Skeleton:      true,
			SkeletonCount: 5,
		},
		Empty: TreeEmptyConfig{
			Message:          "No data available",
			ShowCreateButton: false,
			CreateText:       "Add Root Node",
		},
	}
}

// ============================================================================
// TREE NODE HELPER METHODS
// ============================================================================

// AddChild adds a child node to the current node
func (n *TreeNode) AddChild(child TreeNode) {
	child.ParentID = n.ID
	child.Level = n.Level + 1
	child.Path = append(n.Path, n.ID)
	n.Children = append(n.Children, child)
	n.HasChildren = true
	n.IsLeaf = false
}

// SetData sets a data value for a column field
func (n *TreeNode) SetData(field string, value interface{}) {
	if n.Data == nil {
		n.Data = make(map[string]interface{})
	}
	n.Data[field] = value
}

// GetData retrieves a data value for a column field
func (n *TreeNode) GetData(field string) interface{} {
	if n.Data == nil {
		return nil
	}
	return n.Data[field]
}

// IsRoot returns true if the node has no parent
func (n *TreeNode) IsRoot() bool {
	return n.ParentID == ""
}

// GetDepth returns the depth level of the node
func (n *TreeNode) GetDepth() int {
	return n.Level
}

// ToggleExpanded toggles the expanded state of the node
func (n *TreeNode) ToggleExpanded() {
	if n.HasChildren {
		n.Expanded = !n.Expanded
	}
}

// ToggleSelected toggles the selected state of the node
func (n *TreeNode) ToggleSelected() {
	n.Selected = !n.Selected
}

// ============================================================================
// TREE COLUMN HELPER METHODS
// ============================================================================

// WithAggregation adds aggregation configuration to a column
func (c TreeColumn) WithAggregation(method string, excludeParent bool) TreeColumn {
	c.Aggregation = &TreeAggregation{
		Enabled:            true,
		Method:             method,
		ExcludeParentValue: excludeParent,
		DisplayMode:        "replace",
		Recursive:          true,
	}
	return c
}

// WithWidth sets the column width
func (c TreeColumn) WithWidth(width string) TreeColumn {
	c.Width = width
	return c
}

// WithAlign sets the column alignment
func (c TreeColumn) WithAlign(align string) TreeColumn {
	c.Align = align
	return c
}

// WithType sets the column type
func (c TreeColumn) WithType(colType string) TreeColumn {
	c.Type = colType
	return c
}

// WithFormat sets the column format
func (c TreeColumn) WithFormat(format string) TreeColumn {
	c.Format = format
	return c
}

// AsSortable makes the column sortable
func (c TreeColumn) AsSortable() TreeColumn {
	c.Sortable = true
	return c
}

// ============================================================================
// TREE PROPS BUILDER METHODS
// ============================================================================

// WithData returns a new TreeProps with specified data
func (p TreeProps) WithData(data []TreeNode) TreeProps {
	p.Data = data
	return p
}

// WithColumns returns a new TreeProps with specified columns
func (p TreeProps) WithColumns(columns ...TreeColumn) TreeProps {
	p.Columns = columns
	return p
}

// WithSearchable returns a new TreeProps with search enabled
func (p TreeProps) WithSearchable(searchable bool) TreeProps {
	p.Config.Searchable = searchable
	if searchable && p.SearchProps == nil {
		p.SearchProps = &TreeSearchConfig{
			Placeholder:      "Search tree...",
			Debounce:         300,
			MinLength:        0,
			Clearable:        true,
			HighlightMatches: true,
			ExpandMatches:    true,
			CaseSensitive:    false,
			SearchFields:     []string{"label"},
		}
	}
	return p
}

// WithSelectable returns a new TreeProps with selection enabled
func (p TreeProps) WithSelectable(selectable, multiSelect bool) TreeProps {
	p.Config.Selectable = selectable
	p.Config.MultiSelect = multiSelect
	if selectable {
		p.Config.Checkboxes = true
	}
	return p
}

// WithLazyLoad returns a new TreeProps with lazy loading enabled
func (p TreeProps) WithLazyLoad(lazyLoadUrl string) TreeProps {
	p.Config.LazyLoad = true
	p.LazyLoadUrl = lazyLoadUrl
	return p
}

// WithStyling returns a new TreeProps with specified styling
func (p TreeProps) WithStyling(variant atoms.Variant, size atoms.Size) TreeProps {
	p.Styling.Variant = variant
	p.Styling.Size = size
	return p
}

// WithActions returns a new TreeProps with specified actions
func (p TreeProps) WithActions(actions ...TreeAction) TreeProps {
	p.Actions = append(p.Actions, actions...)
	return p
}

// WithNodeActions returns a new TreeProps with specified node actions
func (p TreeProps) WithNodeActions(actions ...TreeAction) TreeProps {
	p.NodeActions = append(p.NodeActions, actions...)
	return p
}

// AsCompact returns a new TreeProps with compact styling
func (p TreeProps) AsCompact() TreeProps {
	p.Styling.Size = atoms.SizeSM
	p.Config.NodeHeight = "2rem"
	p.Config.Indent = "1rem"
	return p
}

// AsExpandAll returns a new TreeProps with all nodes expanded by default
func (p TreeProps) AsExpandAll() TreeProps {
	p.Config.ExpandAll = true
	return p
}

// WithHeaders enables/disables column headers
func (p TreeProps) WithHeaders(show bool) TreeProps {
	p.Config.ShowHeaders = show
	return p
}

// ============================================================================
// SPECIALIZED TREE CONSTRUCTORS
// ============================================================================

// NewChartOfAccountsTree creates a tree configured for chart of accounts
func NewChartOfAccountsTree(id string) TreeProps {
	props := NewTreeProps(id)

	props.Columns = []TreeColumn{
		NewTreeColumn("accountName", "Account Name").
			WithWidth("40%").
			WithType("text"),
		NewTreeColumn("accountCode", "Code").
			WithWidth("15%").
			WithType("text").
			WithAlign("center"),
		NewTreeColumn("accountType", "Type").
			WithWidth("15%").
			WithType("text"),
		NewTreeColumn("balance", "Balance").
			WithWidth("20%").
			WithType("currency").
			WithFormat("$0,0.00").
			WithAlign("right").
			WithAggregation("sum", true),
		NewTreeColumn("currency", "Currency").
			WithWidth("10%").
			WithType("text").
			WithAlign("center"),
	}

	props.Config.Selectable = true
	props.Config.MultiSelect = false
	props.Config.Checkboxes = false
	props.Config.ShowLines = true
	props.Config.ShowHeaders = true
	props.Styling.Size = atoms.SizeMD

	return props
}

// NewOrganizationTree creates a tree configured for organizational hierarchy
func NewOrganizationTree(id string) TreeProps {
	props := NewTreeProps(id)

	props.Columns = []TreeColumn{
		NewTreeColumn("name", "Organization Unit").
			WithWidth("40%").
			WithType("text"),
		NewTreeColumn("type", "Type").
			WithWidth("15%").
			WithType("text"),
		NewTreeColumn("location", "Location").
			WithWidth("20%").
			WithType("text"),
		NewTreeColumn("employeeCount", "Employees").
			WithWidth("12.5%").
			WithType("number").
			WithAlign("right").
			WithAggregation("sum", true),
		NewTreeColumn("budget", "Budget").
			WithWidth("12.5%").
			WithType("currency").
			WithFormat("$0,0").
			WithAlign("right").
			WithAggregation("sum", true),
	}

	props.Config.ShowHeaders = true
	props.Config.ShowLines = true
	props.Config.ExpandAll = false

	return props
}

// NewLocationTree creates a tree configured for geographic hierarchy
func NewLocationTree(id string) TreeProps {
	props := NewTreeProps(id)

	props.Columns = []TreeColumn{
		NewTreeColumn("locationName", "Location").
			WithWidth("40%").
			WithType("text"),
		NewTreeColumn("locationType", "Type").
			WithWidth("15%").
			WithType("text"),
		NewTreeColumn("population", "Population").
			WithWidth("15%").
			WithType("number").
			WithFormat("0,0").
			WithAlign("right").
			WithAggregation("sum", true),
		NewTreeColumn("area", "Area (km²)").
			WithWidth("15%").
			WithType("number").
			WithFormat("0,0.00").
			WithAlign("right").
			WithAggregation("sum", true),
		NewTreeColumn("gdp", "GDP").
			WithWidth("15%").
			WithType("currency").
			WithFormat("$0.0a").
			WithAlign("right").
			WithAggregation("sum", false),
	}

	props.Config.LazyLoad = true
	props.Config.ShowHeaders = true
	props.WithSearchable(true)

	return props
}

// NewAccountNode creates a chart of accounts node
func NewAccountNode(code, name, accountType string, balance float64) TreeNode {
	node := NewTreeNode(code, name)
	node.SetData("accountName", name)
	node.SetData("accountCode", code)
	node.SetData("accountType", accountType)
	node.SetData("balance", balance)
	node.SetData("currency", "USD")
	return node
}

