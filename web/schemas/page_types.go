package schemas

import (
	"context"
	"encoding/json"
	"time"

	"github.com/a-h/templ"
)

// PageSchema defines the complete structure for a JSON-driven page
type PageSchema struct {
	ID          string                `json:"id"`
	Version     string                `json:"version,omitempty"`
	Layout      string                `json:"layout"` // "app", "auth", "minimal", "base"
	Title       string                `json:"title"`
	Description string                `json:"description,omitempty"`
	Components  []ComponentDefinition `json:"components"`
	DataSources []DataSource          `json:"dataSources,omitempty"`
	Actions     []ActionDefinition    `json:"actions,omitempty"`
	Permissions *PermissionRules      `json:"permissions,omitempty"`
	Meta        map[string]any        `json:"meta,omitempty"`
	CreatedAt   *time.Time            `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time            `json:"updatedAt,omitempty"`

	// Reusable component definitions (like "$ref" pattern)
	Definitions map[string]*ComponentDefinition `json:"definitions,omitempty"`

	// Page-level HTMX configuration
	HTMX *HTMXConfig `json:"htmx,omitempty"`

	// Page-level Alpine.js configuration
	Alpine *AlpineConfig `json:"alpine,omitempty"`
}

// HTMXConfig defines page-level HTMX settings
type HTMXConfig struct {
	Timeout          int    `json:"timeout,omitempty"`          // Request timeout in ms
	HistoryEnabled   bool   `json:"historyEnabled,omitempty"`   // Enable history support
	DefaultSwapStyle string `json:"defaultSwapStyle,omitempty"` // Default swap strategy
	IndicatorClass   string `json:"indicatorClass,omitempty"`   // Loading indicator class
	ScrollBehavior   string `json:"scrollBehavior,omitempty"`   // "smooth", "auto"
	DefaultSwapDelay int    `json:"defaultSwapDelay,omitempty"` // Delay in ms
}

// AlpineConfig defines page-level Alpine.js settings
type AlpineConfig struct {
	Plugins []string          `json:"plugins,omitempty"` // "morph", "focus", "intersect"
	Stores  map[string]string `json:"stores,omitempty"`  // Global Alpine stores
	CSP     bool              `json:"csp,omitempty"`     // Content Security Policy mode
}

// ComponentDefinition describes how to render a component within a schema
type ComponentDefinition struct {
	ID          string                `json:"id"`
	Type        string                `json:"type"` // "atoms.button", "organisms.table", "crud", "form"
	Props       map[string]any        `json:"props"`
	Layout      *LayoutRules          `json:"layout,omitempty"`
	Children    []ComponentDefinition `json:"children,omitempty"`
	DataSource  string                `json:"dataSource,omitempty"`
	Permissions *PermissionRules      `json:"permissions,omitempty"`
	Conditions  []RenderCondition     `json:"conditions,omitempty"`
	Events      []EventHandler        `json:"events,omitempty"`

	// Form field configuration
	Name        string           `json:"name,omitempty"`        // Field name for forms
	Label       string           `json:"label,omitempty"`       // Field label
	Value       any              `json:"value,omitempty"`       // Default value
	Placeholder string           `json:"placeholder,omitempty"` // Input placeholder
	Required    bool             `json:"required,omitempty"`    // Required field
	Validations []ValidationRule `json:"validations,omitempty"` // Validation rules
	ValidateOn  string           `json:"validateOn,omitempty"`  // "change", "blur", "submit"

	// Expression-based visibility and state
	VisibleOn  string `json:"visibleOn,omitempty"`  // Expression: "${status === 'active'}"
	HiddenOn   string `json:"hiddenOn,omitempty"`   // Expression: "${role !== 'admin'}"
	DisabledOn string `json:"disabledOn,omitempty"` // Expression: "${!canEdit}"
	StaticOn   string `json:"staticOn,omitempty"`   // Render as static when true

	// Data binding and templates
	Source string `json:"source,omitempty"` // Data path: "${items}", "${user.name}"
	Tpl    string `json:"tpl,omitempty"`    // Template: "${name} - ${age}"

	// API integration
	API         *APIDefinition `json:"api,omitempty"`         // Primary API endpoint
	InitAPI     *APIDefinition `json:"initApi,omitempty"`     // Initialization API
	SchemaAPI   *APIDefinition `json:"schemaApi,omitempty"`   // Dynamic schema loading
	InitFetch   bool           `json:"initFetch,omitempty"`   // Fetch data on mount
	InitFetchOn string         `json:"initFetchOn,omitempty"` // Conditional init fetch

	// Alpine.js directives
	XData  string            `json:"xData,omitempty"`  // x-data attribute
	XShow  string            `json:"xShow,omitempty"`  // x-show expression
	XIf    string            `json:"xIf,omitempty"`    // x-if expression
	XBind  map[string]string `json:"xBind,omitempty"`  // x-bind:* attributes
	XModel string            `json:"xModel,omitempty"` // x-model binding
	XOn    map[string]string `json:"xOn,omitempty"`    // x-on:* event handlers
	XText  string            `json:"xText,omitempty"`  // x-text content
	XHTML  string            `json:"xHtml,omitempty"`  // x-html content

	// HTMX directives (for partial rendering and OOB swaps)
	IsPartial bool   `json:"isPartial,omitempty"` // Render without layout wrapper
	OOBSwap   string `json:"oobSwap,omitempty"`   // Out-of-band swap: "true" or selector

	// Component-specific configurations (polymorphic based on Type)
	CRUD   *CRUDConfig   `json:"crud,omitempty"`   // For type="crud"
	Form   *FormConfig   `json:"form,omitempty"`   // For type="form"
	Dialog *DialogConfig `json:"dialog,omitempty"` // For type="dialog"
	Drawer *DrawerConfig `json:"drawer,omitempty"` // For type="drawer"
	Wizard *WizardConfig `json:"wizard,omitempty"` // For type="wizard"

	// Styling
	Class    string `json:"class,omitempty"`    // Additional CSS classes
	Disabled bool   `json:"disabled,omitempty"` // Disabled state
}

// APIDefinition describes an API endpoint configuration
type APIDefinition struct {
	Method  string            `json:"method"`            // GET, POST, PUT, DELETE
	URL     string            `json:"url"`               // Endpoint URL (supports ${} expressions)
	Data    map[string]any    `json:"data,omitempty"`    // Request data
	Headers map[string]string `json:"headers,omitempty"` // Custom headers
	SendOn  string            `json:"sendOn,omitempty"`  // Condition expression for sending

	// Data transformation
	Adaptor        string `json:"adaptor,omitempty"`        // Response transformation script
	RequestAdaptor string `json:"requestAdaptor,omitempty"` // Request transformation script

	// Response handling
	DataType     string        `json:"dataType,omitempty"`     // "json", "form-data", "text"
	ResponseType string        `json:"responseType,omitempty"` // Expected response format
	Cache        *CacheConfig  `json:"cache,omitempty"`        // Caching configuration
	Timeout      time.Duration `json:"timeout,omitempty"`      // Request timeout

	// HTMX-specific
	HXTarget  string `json:"hxTarget,omitempty"`  // Target element selector
	HXSwap    string `json:"hxSwap,omitempty"`    // Swap strategy
	HXInclude string `json:"hxInclude,omitempty"` // Additional data to include
	HXSelect  string `json:"hxSelect,omitempty"`  // Select portion of response
}

// CRUDConfig defines CRUD table/list configuration
type CRUDConfig struct {
	API           *APIDefinition       `json:"api"`                     // Data fetching endpoint
	Filter        *ComponentDefinition `json:"filter,omitempty"`        // Filter form
	HeaderToolbar []ActionDefinition   `json:"headerToolbar,omitempty"` // Toolbar actions
	ItemActions   []ActionDefinition   `json:"itemActions,omitempty"`   // Per-item actions
	BulkActions   []ActionDefinition   `json:"bulkActions,omitempty"`   // Bulk operations
	Columns       []ColumnDefinition   `json:"columns"`                 // Table columns
	QuickEdit     *ComponentDefinition `json:"quickEdit,omitempty"`     // Inline editing form
	QuickSave     *APIDefinition       `json:"quickSave,omitempty"`     // Quick save endpoint
	PerPage       int                  `json:"perPage,omitempty"`       // Items per page
	Draggable     bool                 `json:"draggable,omitempty"`     // Enable drag-drop reorder
	Sortable      bool                 `json:"sortable,omitempty"`      // Enable column sorting
	Selectable    bool                 `json:"selectable,omitempty"`    // Enable row selection
	Messages      *CRUDMessages        `json:"messages,omitempty"`      // User-facing messages
}

// ColumnDefinition defines a table column
type ColumnDefinition struct {
	Name       string             `json:"name"`                 // Data field name
	Label      string             `json:"label"`                // Column header
	Type       string             `json:"type,omitempty"`       // "text", "date", "image", "operation"
	Sortable   bool               `json:"sortable,omitempty"`   // Can sort by this column
	Searchable bool               `json:"searchable,omitempty"` // Can search this column
	Tpl        string             `json:"tpl,omitempty"`        // Custom template
	Width      string             `json:"width,omitempty"`      // Column width
	Fixed      string             `json:"fixed,omitempty"`      // "left", "right"
	Toggled    bool               `json:"toggled,omitempty"`    // Initially visible
	Breakpoint string             `json:"breakpoint,omitempty"` // Hide on smaller screens
	Buttons    []ActionDefinition `json:"buttons,omitempty"`    // Action buttons (for type="operation")
}

// CRUDMessages defines user-facing messages for CRUD operations
type CRUDMessages struct {
	FetchFailed   string `json:"fetchFailed,omitempty"`
	SaveSuccess   string `json:"saveSuccess,omitempty"`
	SaveFailed    string `json:"saveFailed,omitempty"`
	DeleteConfirm string `json:"deleteConfirm,omitempty"`
	DeleteSuccess string `json:"deleteSuccess,omitempty"`
	DeleteFailed  string `json:"deleteFailed,omitempty"`
}

// FormConfig defines form configuration
type FormConfig struct {
	API              *APIDefinition        `json:"api"`                        // Form submission endpoint
	InitAPI          *APIDefinition        `json:"initApi,omitempty"`          // Load initial data
	Mode             string                `json:"mode,omitempty"`             // "normal", "horizontal", "inline"
	Body             []ComponentDefinition `json:"body"`                       // Form fields
	Actions          []ActionDefinition    `json:"actions,omitempty"`          // Form actions (submit, reset)
	Messages         *FormMessages         `json:"messages,omitempty"`         // User messages
	LabelWidth       string                `json:"labelWidth,omitempty"`       // Label column width
	WrapWithPanel    bool                  `json:"wrapWithPanel,omitempty"`    // Wrap in panel
	Redirect         string                `json:"redirect,omitempty"`         // Redirect after submit
	Reload           string                `json:"reload,omitempty"`           // Component to reload after submit
	PristineCheck    bool                  `json:"pristineCheck,omitempty"`    // Check if form is modified
	SubmitOnChange   bool                  `json:"submitOnChange,omitempty"`   // Auto-submit on field change
	SubmitOnInit     bool                  `json:"submitOnInit,omitempty"`     // Submit on initialization
	ResetAfterSubmit bool                  `json:"resetAfterSubmit,omitempty"` // Reset form after successful submit
	Debug            bool                  `json:"debug,omitempty"`            // Show debug info
}

// FormMessages defines form user messages
type FormMessages struct {
	FetchSuccess   string `json:"fetchSuccess,omitempty"`
	FetchFailed    string `json:"fetchFailed,omitempty"`
	SaveSuccess    string `json:"saveSuccess,omitempty"`
	SaveFailed     string `json:"saveFailed,omitempty"`
	ValidateFailed string `json:"validateFailed,omitempty"`
}

// DialogConfig defines modal dialog configuration
type DialogConfig struct {
	Title           string               `json:"title"`
	Body            *ComponentDefinition `json:"body"`                      // Dialog content
	Size            string               `json:"size,omitempty"`            // "sm", "md", "lg", "xl", "full"
	Actions         []ActionDefinition   `json:"actions,omitempty"`         // Dialog action buttons
	CloseOnEsc      bool                 `json:"closeOnEsc,omitempty"`      // Allow ESC to close
	CloseOnOutside  bool                 `json:"closeOnOutside,omitempty"`  // Close on backdrop click
	ShowCloseButton bool                 `json:"showCloseButton,omitempty"` // Show X button
	ShowErrorMsg    bool                 `json:"showErrorMsg,omitempty"`    // Show error messages
	ShowLoading     bool                 `json:"showLoading,omitempty"`     // Show loading state
	Draggable       bool                 `json:"draggable,omitempty"`       // Allow dragging
}

// DrawerConfig defines side drawer configuration
type DrawerConfig struct {
	Title           string               `json:"title"`
	Body            *ComponentDefinition `json:"body"`
	Position        string               `json:"position,omitempty"` // "left", "right", "top", "bottom"
	Size            string               `json:"size,omitempty"`     // "sm", "md", "lg", "xl"
	Overlay         bool                 `json:"overlay,omitempty"`  // Show backdrop
	CloseOnEsc      bool                 `json:"closeOnEsc,omitempty"`
	CloseOnOutside  bool                 `json:"closeOnOutside,omitempty"`
	ShowCloseButton bool                 `json:"showCloseButton,omitempty"`
	Resizable       bool                 `json:"resizable,omitempty"` // Allow resize
}

// WizardConfig defines multi-step wizard configuration
type WizardConfig struct {
	Steps             []WizardStep   `json:"steps"`
	API               *APIDefinition `json:"api,omitempty"`               // Final submission
	InitAPI           *APIDefinition `json:"initApi,omitempty"`           // Load initial data
	Mode              string         `json:"mode,omitempty"`              // "horizontal", "vertical"
	ActionPrevLabel   string         `json:"actionPrevLabel,omitempty"`   // Previous button text
	ActionNextLabel   string         `json:"actionNextLabel,omitempty"`   // Next button text
	ActionFinishLabel string         `json:"actionFinishLabel,omitempty"` // Finish button text
	Redirect          string         `json:"redirect,omitempty"`          // Redirect on completion
	Reload            string         `json:"reload,omitempty"`            // Reload component on completion
}

// WizardStep defines a single wizard step
type WizardStep struct {
	Title      string               `json:"title"`
	Subtitle   string               `json:"subtitle,omitempty"`
	Body       *ComponentDefinition `json:"body"`                 // Step content (usually a form)
	API        *APIDefinition       `json:"api,omitempty"`        // Step-specific API
	JumpableOn string               `json:"jumpableOn,omitempty"` // Condition to allow jumping to this step
	VisibleOn  string               `json:"visibleOn,omitempty"`  // Condition to show step
	Icon       string               `json:"icon,omitempty"`
}

// ValidationRule defines field validation
type ValidationRule struct {
	// Standard validations
	Required  bool     `json:"required,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`   // Regex pattern
	Min       *float64 `json:"min,omitempty"`       // Minimum value
	Max       *float64 `json:"max,omitempty"`       // Maximum value
	MinLength *int     `json:"minLength,omitempty"` // Minimum string length
	MaxLength *int     `json:"maxLength,omitempty"` // Maximum string length

	// Type validations
	IsEmail    bool `json:"isEmail,omitempty"`
	IsURL      bool `json:"isUrl,omitempty"`
	IsNumeric  bool `json:"isNumeric,omitempty"`
	IsInt      bool `json:"isInt,omitempty"`
	IsAlpha    bool `json:"isAlpha,omitempty"`
	IsAlphaNum bool `json:"isAlphaNum,omitempty"`

	// Custom validation
	Validator  string `json:"validator,omitempty"`  // Expression or function name
	MatchField string `json:"matchField,omitempty"` // Must match another field

	// Error messages
	Message         string            `json:"message,omitempty"`
	ValidatorErrors map[string]string `json:"validatorErrors,omitempty"` // Custom messages per validation
}

// LayoutRules define positioning and styling for components
type LayoutRules struct {
	Container  string            `json:"container,omitempty"` // "grid", "flex", "block"
	Position   string            `json:"position,omitempty"`  // "relative", "absolute", "fixed"
	Grid       *GridLayout       `json:"grid,omitempty"`
	Flex       *FlexLayout       `json:"flex,omitempty"`
	Spacing    *SpacingRules     `json:"spacing,omitempty"`
	Responsive map[string]string `json:"responsive,omitempty"` // Breakpoint → layout changes
	Classes    []string          `json:"classes,omitempty"`    // Additional CSS classes
	Styles     map[string]string `json:"styles,omitempty"`     // Inline styles
}

// GridLayout defines CSS Grid properties
type GridLayout struct {
	Columns    string `json:"columns"` // "repeat(3, 1fr)", "300px 1fr"
	Rows       string `json:"rows,omitempty"`
	Gap        string `json:"gap,omitempty"` // "1rem", "16px"
	ColumnSpan int    `json:"columnSpan,omitempty"`
	RowSpan    int    `json:"rowSpan,omitempty"`
	Area       string `json:"area,omitempty"` // Grid area name
}

// FlexLayout defines CSS Flexbox properties
type FlexLayout struct {
	Direction string `json:"direction,omitempty"` // "row", "column"
	Wrap      string `json:"wrap,omitempty"`      // "wrap", "nowrap"
	Justify   string `json:"justify,omitempty"`   // "center", "space-between"
	Align     string `json:"align,omitempty"`     // "center", "stretch"
	Gap       string `json:"gap,omitempty"`
	Grow      int    `json:"grow,omitempty"`
	Shrink    int    `json:"shrink,omitempty"`
	Basis     string `json:"basis,omitempty"`
}

// SpacingRules define margin and padding
type SpacingRules struct {
	Margin   string `json:"margin,omitempty"` // "1rem", "16px 24px"
	Padding  string `json:"padding,omitempty"`
	MarginX  string `json:"marginX,omitempty"`
	MarginY  string `json:"marginY,omitempty"`
	PaddingX string `json:"paddingX,omitempty"`
	PaddingY string `json:"paddingY,omitempty"`
}

// DataSource defines data fetching configuration
type DataSource struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // "api", "static", "computed", "sse"
	Endpoint   string            `json:"endpoint,omitempty"`
	Method     string            `json:"method,omitempty"` // "GET", "POST"
	Headers    map[string]string `json:"headers,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Transform  string            `json:"transform,omitempty"` // Data transformation expression
	Cache      *CacheConfig      `json:"cache,omitempty"`
	Refresh    *RefreshConfig    `json:"refresh,omitempty"`
	Fallback   any               `json:"fallback,omitempty"` // Fallback data

	// Server-Sent Events configuration
	SSE      bool `json:"sse,omitempty"`      // Use SSE for real-time updates
	SSERetry int  `json:"sseRetry,omitempty"` // Retry interval in ms
}

// CacheConfig defines caching behavior for data sources
type CacheConfig struct {
	TTL        time.Duration `json:"ttl"`
	Key        string        `json:"key,omitempty"`
	Invalidate []string      `json:"invalidate,omitempty"` // Events that invalidate cache
}

// RefreshConfig defines automatic refresh behavior
type RefreshConfig struct {
	Interval  time.Duration `json:"interval,omitempty"`
	Trigger   string        `json:"trigger,omitempty"`   // "visibility", "focus", "manual"
	Condition string        `json:"condition,omitempty"` // Condition for refresh
}

// ActionDefinition describes user actions and their handlers
type ActionDefinition struct {
	ID         string `json:"id"`
	Type       string `json:"type,omitempty"` // Backward compatibility
	ActionType string `json:"actionType"`     // "ajax", "dialog", "drawer", "url", "link", "submit", "reset", "reload"
	Label      string `json:"label,omitempty"`
	Icon       string `json:"icon,omitempty"`
	Level      string `json:"level,omitempty"` // "primary", "secondary", "danger", "warning"
	Size       string `json:"size,omitempty"`  // "sm", "md", "lg"

	// Legacy fields (backward compatible)
	Target      string            `json:"target,omitempty"`  // URL, endpoint, or element selector
	Method      string            `json:"method,omitempty"`  // HTTP method for API calls
	Data        map[string]string `json:"data,omitempty"`    // Data to send
	Confirm     *ConfirmDialog    `json:"confirm,omitempty"` // Confirmation dialog
	Success     *ActionResponse   `json:"success,omitempty"` // Success handling
	Error       *ActionResponse   `json:"error,omitempty"`   // Error handling
	Permissions []string          `json:"permissions,omitempty"`

	// API action
	API *APIDefinition `json:"api,omitempty"` // For actionType="ajax"

	// Dialog/Drawer actions
	Dialog *DialogConfig `json:"dialog,omitempty"` // For actionType="dialog"
	Drawer *DrawerConfig `json:"drawer,omitempty"` // For actionType="drawer"

	// Navigation
	Link     string `json:"link,omitempty"`     // For actionType="url" or "link"
	Blank    bool   `json:"blank,omitempty"`    // Open in new tab
	Redirect string `json:"redirect,omitempty"` // Redirect URL

	// Post-action behavior
	Reload      string       `json:"reload,omitempty"`      // Component/page to reload
	Toast       *ToastConfig `json:"toast,omitempty"`       // Show toast message
	Close       bool         `json:"close,omitempty"`       // Close current dialog/drawer
	Required    []string     `json:"required,omitempty"`    // Required fields
	ConfirmText string       `json:"confirmText,omitempty"` // Quick confirmation text

	// HTMX-specific fields
	HXGet     string `json:"hxGet,omitempty"`     // hx-get
	HXPost    string `json:"hxPost,omitempty"`    // hx-post
	HXPut     string `json:"hxPut,omitempty"`     // hx-put
	HXPatch   string `json:"hxPatch,omitempty"`   // hx-patch
	HXDelete  string `json:"hxDelete,omitempty"`  // hx-delete
	HXTarget  string `json:"hxTarget,omitempty"`  // hx-target
	HXSwap    string `json:"hxSwap,omitempty"`    // hx-swap
	HXTrigger string `json:"hxTrigger,omitempty"` // hx-trigger
	HXPushURL bool   `json:"hxPushUrl,omitempty"` // hx-push-url
	HXSelect  string `json:"hxSelect,omitempty"`  // hx-select
	HXInclude string `json:"hxInclude,omitempty"` // hx-include
}

// ToastConfig defines toast notification configuration
type ToastConfig struct {
	Title       string        `json:"title,omitempty"`
	Body        string        `json:"body,omitempty"`
	Level       string        `json:"level,omitempty"`    // "info", "success", "warning", "error"
	Position    string        `json:"position,omitempty"` // "top-right", "bottom-center", etc.
	Duration    time.Duration `json:"duration,omitempty"` // Display duration
	CloseButton bool          `json:"closeButton,omitempty"`
}

// ConfirmDialog defines confirmation dialog for actions
type ConfirmDialog struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Confirm string `json:"confirm,omitempty"` // Confirm button text
	Cancel  string `json:"cancel,omitempty"`  // Cancel button text
}

// ActionResponse defines response handling for actions
type ActionResponse struct {
	Type     string            `json:"type"` // "redirect", "refresh", "toast", "update"
	Target   string            `json:"target,omitempty"`
	Message  string            `json:"message,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	Duration time.Duration     `json:"duration,omitempty"` // For toast messages

	// HTMX response headers
	HXTrigger  string `json:"hxTrigger,omitempty"`  // HX-Trigger header
	HXRedirect string `json:"hxRedirect,omitempty"` // HX-Redirect
	HXRefresh  bool   `json:"hxRefresh,omitempty"`  // HX-Refresh
	HXReswap   string `json:"hxReswap,omitempty"`   // HX-Reswap
	HXRetarget string `json:"hxRetarget,omitempty"` // HX-Retarget
}

// PermissionRules define access control for components and actions
type PermissionRules struct {
	RequiredRoles       []string             `json:"requiredRoles,omitempty"`
	RequiredPermissions []string             `json:"requiredPermissions,omitempty"`
	ExcludedRoles       []string             `json:"excludedRoles,omitempty"`
	Conditions          []PermissionCheck    `json:"conditions,omitempty"`
	Fallback            *ComponentDefinition `json:"fallback,omitempty"` // Component to show if access denied
}

// PermissionCheck defines conditional permission logic
type PermissionCheck struct {
	Field    string `json:"field"`            // Field to check
	Operator string `json:"operator"`         // "eq", "ne", "in", "contains"
	Value    any    `json:"value"`            // Value to compare
	Source   string `json:"source,omitempty"` // "user", "context", "data"
}

// RenderCondition defines when a component should be rendered
type RenderCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // "eq", "ne", "gt", "lt", "in", "exists"
	Value    any    `json:"value"`
	Source   string `json:"source,omitempty"` // "props", "data", "user", "context"
}

// EventHandler defines client-side event handling
type EventHandler struct {
	Event   string         `json:"event"`                     // "click", "change", "submit"
	Action  string         `json:"action"`                    // Action ID or inline handler
	Target  string         `json:"target,omitempty"`          // Element selector for delegation
	Data    map[string]any `json:"data,omitempty"`            // Additional data for handler
	Prevent bool           `json:"preventDefault,omitempty"`  // Prevent default behavior
	Stop    bool           `json:"stopPropagation,omitempty"` // Stop event bubbling

	// HTMX event attributes
	HXGet     string `json:"hxGet,omitempty"`     // hx-get
	HXPost    string `json:"hxPost,omitempty"`    // hx-post
	HXPut     string `json:"hxPut,omitempty"`     // hx-put
	HXPatch   string `json:"hxPatch,omitempty"`   // hx-patch
	HXDelete  string `json:"hxDelete,omitempty"`  // hx-delete
	HXTarget  string `json:"hxTarget,omitempty"`  // hx-target
	HXSwap    string `json:"hxSwap,omitempty"`    // hx-swap
	HXTrigger string `json:"hxTrigger,omitempty"` // hx-trigger override
	HXInclude string `json:"hxInclude,omitempty"` // hx-include
}

// SchemaValidator defines validation interface for schemas
type SchemaValidator interface {
	ValidatePageSchema(schema *PageSchema) []ValidationError
	ValidateComponentDefinition(def *ComponentDefinition) []ValidationError
	ValidateDataSource(ds *DataSource) []ValidationError
	ValidatePermissions(perms *PermissionRules) []ValidationError
}

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"` // JSON path to the error
}

// ComponentResolver defines interface for resolving components from schema paths
type ComponentResolver interface {
	Resolve(componentType string) (templ.Component, error)
	Register(componentType string, factory ComponentFactory) error
	List() []string
	Exists(componentType string) bool
}

// ComponentFactory creates components from schema definitions
type ComponentFactory func(def *ComponentDefinition, ctx context.Context) (templ.Component, error)

// SchemaRenderer converts schemas to renderable templates
type SchemaRenderer interface {
	RenderPage(schema *PageSchema, ctx context.Context) (templ.Component, error)
	RenderComponent(def *ComponentDefinition, ctx context.Context) (templ.Component, error)
	RenderLayout(layoutType string, content templ.Component, ctx context.Context) (templ.Component, error)
}

// PermissionChecker validates user permissions against schema requirements
type PermissionChecker interface {
	CheckPermissions(rules *PermissionRules, ctx context.Context) (bool, error)
	CheckRoles(roles []string, ctx context.Context) (bool, error)
	CheckConditions(conditions []PermissionCheck, ctx context.Context) (bool, error)
}

// DataProvider fetches and manages data for schema components
type DataProvider interface {
	GetData(source *DataSource, ctx context.Context) (any, error)
	InvalidateCache(sourceID string) error
	RefreshData(sourceID string, ctx context.Context) (any, error)
}

// SchemaManager handles schema storage, versioning, and retrieval
type SchemaManager interface {
	GetSchema(id string, version ...string) (*PageSchema, error)
	SaveSchema(schema *PageSchema) error
	ListSchemas(filter SchemaFilter) ([]*PageSchema, error)
	DeleteSchema(id string, version ...string) error
	ValidateSchema(schema *PageSchema) []ValidationError
}

// SchemaFilter defines filtering criteria for schema queries
type SchemaFilter struct {
	IDs         []string   `json:"ids,omitempty"`
	Layout      string     `json:"layout,omitempty"`
	CreatedFrom *time.Time `json:"createdFrom,omitempty"`
	CreatedTo   *time.Time `json:"createdTo,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
}

// RenderContext provides context information for schema rendering
type RenderContext struct {
	User        any            `json:"user,omitempty"`
	Permissions []string       `json:"permissions"`
	Data        map[string]any `json:"data,omitempty"`
	Theme       string         `json:"theme,omitempty"`
	Language    string         `json:"language,omitempty"`
	Timezone    string         `json:"timezone,omitempty"`
	Debug       bool           `json:"debug,omitempty"`
	RequestID   string         `json:"requestId,omitempty"`
}

// ExpressionEngine evaluates template expressions in component definitions
type ExpressionEngine interface {
	// Eval evaluates an expression with the given data context
	// Supports ${fieldName} and ${object.field} syntax
	Eval(expr string, data map[string]any) (any, error)

	// EvalBool evaluates an expression and returns a boolean result
	EvalBool(expr string, data map[string]any) (bool, error)

	// EvalString evaluates an expression and returns a string result
	EvalString(expr string, data map[string]any) (string, error)
}

// SchemaError represents errors in schema processing
type SchemaError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Component string `json:"component,omitempty"`
	Path      string `json:"path,omitempty"`
	Details   string `json:"details,omitempty"`
}

func (e *SchemaError) Error() string {
	if e.Component != "" {
		return "schema error in " + e.Component + ": " + e.Message
	}
	return "schema error: " + e.Message
}

// Common error codes for schema processing
const (
	ErrCodeComponentNotFound = "COMPONENT_NOT_FOUND"
	ErrCodeMissingComponent  = "MISSING_COMPONENT"
	ErrCodeInvalidProps      = "INVALID_PROPS"
	ErrCodePermissionDenied  = "PERMISSION_DENIED"
	ErrCodeDataSourceError   = "DATA_SOURCE_ERROR"
	ErrCodeLayoutError       = "LAYOUT_ERROR"
	ErrCodeValidationError   = "VALIDATION_ERROR"
	ErrCodeRenderError       = "RENDER_ERROR"
	ErrCodeExpressionError   = "EXPRESSION_ERROR"
	ErrCodeAPIError          = "API_ERROR"
)

// NewSchemaError creates a new schema error
func NewSchemaError(code, message string) *SchemaError {
	return &SchemaError{
		Code:    code,
		Message: message,
	}
}

// WithComponent adds component context to a schema error
func (e *SchemaError) WithComponent(component string) *SchemaError {
	e.Component = component
	return e
}

// WithPath adds path context to a schema error
func (e *SchemaError) WithPath(path string) *SchemaError {
	e.Path = path
	return e
}

// WithDetails adds additional details to a schema error
func (e *SchemaError) WithDetails(details string) *SchemaError {
	e.Details = details
	return e
}

// Helper methods for ComponentDefinition

// IsForm returns true if the component is a form
func (c *ComponentDefinition) IsForm() bool {
	return c.Type == "form" || c.Form != nil
}

// IsCRUD returns true if the component is a CRUD table
func (c *ComponentDefinition) IsCRUD() bool {
	return c.Type == "crud" || c.CRUD != nil
}

// IsDialog returns true if the component is a dialog
func (c *ComponentDefinition) IsDialog() bool {
	return c.Type == "dialog" || c.Dialog != nil
}

// IsDrawer returns true if the component is a drawer
func (c *ComponentDefinition) IsDrawer() bool {
	return c.Type == "drawer" || c.Drawer != nil
}

// IsWizard returns true if the component is a wizard
func (c *ComponentDefinition) IsWizard() bool {
	return c.Type == "wizard" || c.Wizard != nil
}

// GetActionType returns the action type, preferring ActionType over Type for backward compatibility
func (a *ActionDefinition) GetActionType() string {
	if a.ActionType != "" {
		return a.ActionType
	}
	return a.Type
}

// ToJSON converts ComponentDefinition to JSON string
func (c *ComponentDefinition) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
