package schemas

import (
	"context"
	"time"

	"github.com/a-h/templ"
)

// PageSchema defines the complete structure for a JSON-driven page
type PageSchema struct {
	ID          string                 `json:"id"`
	Version     string                 `json:"version,omitempty"`
	Layout      string                 `json:"layout"` // "app", "auth", "minimal", "base"
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Components  []ComponentDefinition  `json:"components"`
	DataSources []DataSource           `json:"dataSources,omitempty"`
	Actions     []ActionDefinition     `json:"actions,omitempty"`
	Permissions *PermissionRules       `json:"permissions,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
	CreatedAt   time.Time              `json:"createdAt,omitempty"`
	UpdatedAt   time.Time              `json:"updatedAt,omitempty"`
}

// ComponentDefinition describes how to render a component within a schema
type ComponentDefinition struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "atoms.button", "organisms.table", etc.
	Props       map[string]any `json:"props"`
	Layout      *LayoutRules           `json:"layout,omitempty"`
	Children    []ComponentDefinition  `json:"children,omitempty"`
	DataSource  string                 `json:"dataSource,omitempty"`
	Permissions *PermissionRules       `json:"permissions,omitempty"`
	Conditions  []RenderCondition      `json:"conditions,omitempty"`
	Events      []EventHandler         `json:"events,omitempty"`
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
	Type       string            `json:"type"` // "api", "static", "computed"
	Endpoint   string            `json:"endpoint,omitempty"`
	Method     string            `json:"method,omitempty"` // "GET", "POST"
	Headers    map[string]string `json:"headers,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Transform  string            `json:"transform,omitempty"` // Data transformation expression
	Cache      *CacheConfig      `json:"cache,omitempty"`
	Refresh    *RefreshConfig    `json:"refresh,omitempty"`
	Fallback   any       `json:"fallback,omitempty"` // Fallback data
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
	ID          string            `json:"id"`
	Type        string            `json:"type"` // "navigate", "submit", "api", "event"
	Label       string            `json:"label,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	Target      string            `json:"target,omitempty"`  // URL, endpoint, or element selector
	Method      string            `json:"method,omitempty"`  // HTTP method for API calls
	Data        map[string]string `json:"data,omitempty"`    // Data to send
	Confirm     *ConfirmDialog    `json:"confirm,omitempty"` // Confirmation dialog
	Success     *ActionResponse   `json:"success,omitempty"` // Success handling
	Error       *ActionResponse   `json:"error,omitempty"`   // Error handling
	Permissions []string          `json:"permissions,omitempty"`
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
	Field    string      `json:"field"`            // Field to check
	Operator string      `json:"operator"`         // "eq", "ne", "in", "contains"
	Value    any `json:"value"`            // Value to compare
	Source   string      `json:"source,omitempty"` // "user", "context", "data"
}

// RenderCondition defines when a component should be rendered
type RenderCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "eq", "ne", "gt", "lt", "in", "exists"
	Value    any `json:"value"`
	Source   string      `json:"source,omitempty"` // "props", "data", "user", "context"
}

// EventHandler defines client-side event handling
type EventHandler struct {
	Event   string                 `json:"event"`                     // "click", "change", "submit"
	Action  string                 `json:"action"`                    // Action ID or inline handler
	Target  string                 `json:"target,omitempty"`          // Element selector for delegation
	Data    map[string]any `json:"data,omitempty"`            // Additional data for handler
	Prevent bool                   `json:"preventDefault,omitempty"`  // Prevent default behavior
	Stop    bool                   `json:"stopPropagation,omitempty"` // Stop event bubbling
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
	IDs         []string  `json:"ids,omitempty"`
	Layout      string    `json:"layout,omitempty"`
	CreatedFrom time.Time `json:"createdFrom,omitempty"`
	CreatedTo   time.Time `json:"createdTo,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
}

// RenderContext provides context information for schema rendering
type RenderContext struct {
	User        any            `json:"user,omitempty"`
	Permissions []string               `json:"permissions"`
	Data        map[string]any `json:"data,omitempty"`
	Theme       string                 `json:"theme,omitempty"`
	Language    string                 `json:"language,omitempty"`
	Timezone    string                 `json:"timezone,omitempty"`
	Debug       bool                   `json:"debug,omitempty"`
	RequestID   string                 `json:"requestId,omitempty"`
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
	ErrCodeInvalidProps      = "INVALID_PROPS"
	ErrCodePermissionDenied  = "PERMISSION_DENIED"
	ErrCodeDataSourceError   = "DATA_SOURCE_ERROR"
	ErrCodeLayoutError       = "LAYOUT_ERROR"
	ErrCodeValidationError   = "VALIDATION_ERROR"
	ErrCodeRenderError       = "RENDER_ERROR"
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
