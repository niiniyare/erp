package table

import (
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/web/components/molecules"
)

// DataTableColumn defines a column configuration
type DataTableColumn struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	Sortable   bool   `json:"sortable"`
	Searchable bool   `json:"searchable"`
	Hidden     bool   `json:"hidden"`
	Width      string `json:"width,omitempty"`
	Type       string `json:"type,omitempty"` // "text", "number", "date", "html"
	Formatter  string `json:"formatter,omitempty"`
	Class      string `json:"class,omitempty"`
}

// DataTableAction defines available actions
type DataTableAction struct {
	Text      string              `json:"text"`
	Icon      string              `json:"icon"`
	Variant   atoms.ButtonVariant `json:"variant"`
	OnClick   string              `json:"onclick,omitempty"`
	HxPost    string              `json:"hxPost,omitempty"`
	HxGet     string              `json:"hxGet,omitempty"`
	HxDelete  string              `json:"hxDelete,omitempty"`
	HxTarget  string              `json:"hxTarget,omitempty"`
	HxSwap    string              `json:"hxSwap,omitempty"`
	HxConfirm string              `json:"hxConfirm,omitempty"`
	ShowWhen  bool                `json:"showWhen,omitempty"` // Conditional display
	Bulk      bool                `json:"bulk,omitempty"`     // For bulk actions
	Single    bool                `json:"single,omitempty"`   // For single row actions
}

// DataTableConfig defines table configuration
type DataTableConfig struct {
	Searchable     bool   `json:"searchable"`
	Sortable       bool   `json:"sortable"`
	Paging         bool   `json:"paging"`
	PageLength     int    `json:"pageLength"`
	Info           bool   `json:"info"`
	LengthChange   bool   `json:"lengthChange"`
	ScrollY        string `json:"scrollY,omitempty"`
	FixedHeader    bool   `json:"fixedHeader"`
	ResponsiveSync bool   `json:"responsiveSync"`
	ServerSide     bool   `json:"serverSide"`
}

// DataTableProps defines properties for the DataTable organism
type DataTableProps struct {
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

	// Server-side processing
	DataUrl   string `json:"dataUrl,omitempty"`   // HTMX endpoint for data
	SearchUrl string `json:"searchUrl,omitempty"` // HTMX endpoint for search

	// HTMX attributes
	HxGet    string `json:"hxGet,omitempty"`    // HTMX GET endpoint
	HxPost   string `json:"hxPost,omitempty"`   // HTMX POST endpoint
	HxTarget string `json:"hxTarget,omitempty"` // HTMX target selector
	HxSwap   string `json:"hxSwap,omitempty"`   // HTMX swap strategy

	// Styling
	Striped  bool `json:"striped"`
	Bordered bool `json:"bordered"`
	Hover    bool `json:"hover"`
	Compact  bool `json:"compact"`

	// HTML attributes
	ID         string `json:"id,omitempty"`
	Class      string `json:"class,omitempty"`
	DataTestID string `json:"dataTestId,omitempty"`

	// Accessibility
	Caption   string `json:"caption,omitempty"`
	AriaLabel string `json:"ariaLabel,omitempty"`
}

// PageInfo represents pagination information
type PageInfo struct {
	Current int `json:"current"`
	Total   int `json:"total"`
	Start   int `json:"start"`
	End     int `json:"end"`
}
