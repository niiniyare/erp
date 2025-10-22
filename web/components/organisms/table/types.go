package table

import (
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/web/components/molecules"
)

// DataTableColumn defines a column configuration
// Aligned with TableColumn schema definition
type DataTableColumn struct {
	// Basic column properties
	Key        string `json:"key"`
	Label      string `json:"label"`
	Sortable   bool   `json:"sortable"`
	Searchable bool   `json:"searchable"`
	Hidden     bool   `json:"hidden"`
	Width      string `json:"width,omitempty"`
	Type       string `json:"type,omitempty"` // "text", "number", "date", "html"
	Formatter  string `json:"formatter,omitempty"`
	Format     string `json:"format,omitempty"` // Date/number format
	Class      string `json:"className,omitempty"`
	
	// Alignment and responsive behavior
	Align      string `json:"align,omitempty"`      // "left", "center", "right"
	Fixed      bool   `json:"fixed,omitempty"`      // Fixed column position
	Resizable  bool   `json:"resizable,omitempty"`  // Column can be resized
	MinWidth   string `json:"minWidth,omitempty"`   // Minimum column width
	MaxWidth   string `json:"maxWidth,omitempty"`   // Maximum column width
	
	// Schema-aligned properties
	DisabledOn *atoms.SchemaExpression `json:"disabledOn,omitempty"` // Dynamic disabled condition
	HiddenOn   *atoms.SchemaExpression `json:"hiddenOn,omitempty"`   // Dynamic hidden condition
}

// DataTableAction defines available actions
// Enhanced with atoms architecture alignment
type DataTableAction struct {
	// Core action properties
	Text    string        `json:"text"`
	Icon    string        `json:"icon"`
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
	Bulk      bool                    `json:"bulk,omitempty"`      // For bulk actions
	Single    bool                    `json:"single,omitempty"`    // For single row actions
	ShowWhen  *atoms.SchemaExpression `json:"showWhen,omitempty"`  // Conditional display
	Disabled  bool                    `json:"disabled,omitempty"`  // Action disabled state
	
	// Styling and accessibility
	Tooltip   string `json:"tooltip,omitempty"`   // Tooltip text
	ID        string `json:"id,omitempty"`        // Action ID
	Class     string `json:"className,omitempty"` // Custom CSS classes
}

// DataTableConfig defines table configuration
// Enhanced with comprehensive table options
type DataTableConfig struct {
	// Core table features
	Searchable   bool `json:"searchable"`
	Sortable     bool `json:"sortable"`
	Paging       bool `json:"paging"`
	PageLength   int  `json:"pageLength"`
	Info         bool `json:"info"`
	LengthChange bool `json:"lengthChange"`
	
	// Layout and display
	ScrollY        string `json:"scrollY,omitempty"`        // Vertical scroll height
	ScrollX        bool   `json:"scrollX,omitempty"`        // Horizontal scroll
	FixedHeader    bool   `json:"fixedHeader"`              // Fixed header on scroll
	FixedColumns   int    `json:"fixedColumns,omitempty"`   // Number of fixed columns
	Responsive     bool   `json:"responsive"`               // Responsive table behavior
	AutoWidth      bool   `json:"autoWidth"`                // Auto-calculate column widths
	
	// Data handling
	ServerSide     bool   `json:"serverSide"`               // Server-side processing
	AjaxURL        string `json:"ajaxUrl,omitempty"`        // Ajax data URL
	DeferRender    bool   `json:"deferRender,omitempty"`    // Defer rendering for performance
	
	// Selection and interaction
	RowSelection   bool   `json:"rowSelection"`             // Enable row selection
	MultiSelect    bool   `json:"multiSelect,omitempty"`    // Multiple row selection
	SelectOnClick  bool   `json:"selectOnClick,omitempty"`  // Select row on click
	
	// Styling options
	Striped        bool   `json:"striped,omitempty"`        // Alternating row colors
	Hover          bool   `json:"hover,omitempty"`          // Hover effects
	Bordered       bool   `json:"bordered,omitempty"`       // Table borders
	Compact        bool   `json:"compact,omitempty"`        // Compact spacing
	
	// Performance options
	StateSave      bool   `json:"stateSave,omitempty"`      // Save table state
	StateDuration  int    `json:"stateDuration,omitempty"`  // State duration in seconds
}

// DataTableProps defines properties for the DataTable organism
// Aligned with atoms architecture using shared property structs
type DataTableProps struct {
	// Base properties (ID, class, etc.)
	atoms.BaseProps
	
	// Accessibility support
	atoms.AccessibilityProps
	
	// Table structure
	Columns []DataTableColumn `json:"columns"`
	Data    []map[string]any  `json:"data,omitempty"`

	// Configuration
	Config DataTableConfig `json:"config"`

	// Actions
	Actions     []DataTableAction `json:"actions,omitempty"`
	BulkActions []DataTableAction `json:"bulkActions,omitempty"`

	// Search and filters
	SearchProps *molecules.SearchProps `json:"searchProps,omitempty"`
	ShowFilters bool                   `json:"showFilters,omitempty"`
	Filterable  bool                   `json:"filterable,omitempty"`

	// Server-side processing
	DataUrl   string `json:"dataUrl,omitempty"`   // HTMX endpoint for data
	SearchUrl string `json:"searchUrl,omitempty"` // HTMX endpoint for search
	FilterUrl string `json:"filterUrl,omitempty"` // HTMX endpoint for filtering

	// HTMX attributes
	HxGet     string `json:"hxGet,omitempty"`     // HTMX GET endpoint
	HxPost    string `json:"hxPost,omitempty"`    // HTMX POST endpoint
	HxTarget  string `json:"hxTarget,omitempty"`  // HTMX target selector
	HxSwap    string `json:"hxSwap,omitempty"`    // HTMX swap strategy
	HxTrigger string `json:"hxTrigger,omitempty"` // HTMX trigger events

	// Table metadata
	Caption     string `json:"caption,omitempty"`     // Table caption
	Summary     string `json:"summary,omitempty"`     // Table summary for accessibility
	Title       string `json:"title,omitempty"`       // Table title
	Description string `json:"description,omitempty"` // Table description
	
	// Loading and empty states
	Loading     bool   `json:"loading,omitempty"`     // Show loading state
	LoadingText string `json:"loadingText,omitempty"` // Custom loading message
	EmptyText   string `json:"emptyText,omitempty"`   // Custom empty state message
	ErrorText   string `json:"errorText,omitempty"`   // Custom error message
	
	// Theme and styling variants
	Variant atoms.Variant `json:"variant,omitempty"` // Table style variant
	Size    atoms.Size    `json:"size,omitempty"`    // Table size
	
	// Event handlers
	OnRowClick    string `json:"onRowClick,omitempty"`    // Row click handler
	OnRowSelect   string `json:"onRowSelect,omitempty"`   // Row selection handler
	OnSort        string `json:"onSort,omitempty"`        // Sort handler
	OnFilter      string `json:"onFilter,omitempty"`      // Filter handler
	OnPageChange  string `json:"onPageChange,omitempty"`  // Page change handler
}

// PageInfo represents pagination information
// Enhanced with comprehensive pagination data
type PageInfo struct {
	Current     int `json:"current"`     // Current page number (1-based)
	Total       int `json:"total"`       // Total number of pages
	Start       int `json:"start"`       // Start record number (1-based)
	End         int `json:"end"`         // End record number
	TotalRecords int `json:"totalRecords"` // Total number of records
	PageSize    int `json:"pageSize"`    // Records per page
	HasNext     bool `json:"hasNext"`     // Has next page
	HasPrev     bool `json:"hasPrev"`     // Has previous page
}

// ============================================================================
// TABLE HELPER FUNCTIONS
// ============================================================================

// NewDataTableColumn creates a new column with sensible defaults
func NewDataTableColumn(key, label string) DataTableColumn {
	return DataTableColumn{
		Key:        key,
		Label:      label,
		Sortable:   true,
		Searchable: true,
		Type:       "text",
		Align:      "left",
		Resizable:  true,
	}
}

// NewDataTableAction creates a new action with sensible defaults
func NewDataTableAction(text, icon string) DataTableAction {
	return DataTableAction{
		Text:    text,
		Icon:    icon,
		Variant: atoms.VariantDefault,
		Size:    atoms.SizeSM,
		Single:  true,
	}
}

// NewDataTableProps creates a new DataTable with sensible defaults
func NewDataTableProps(id string) DataTableProps {
	return DataTableProps{
		BaseProps: atoms.BaseProps{
			ID: atoms.EnsureID(id, "datatable"),
		},
		Config: DataTableConfig{
			Searchable:   true,
			Sortable:     true,
			Paging:       true,
			PageLength:   25,
			Info:         true,
			LengthChange: true,
			FixedHeader:  false,
			Responsive:   true,
			AutoWidth:    true,
			Hover:        true,
		},
		Variant:     atoms.VariantDefault,
		Size:        atoms.SizeMD,
		LoadingText: "Loading...",
		EmptyText:   "No data available",
		ErrorText:   "Error loading data",
	}
}
