package schema

import (
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// Supporting types for the main schema

// Section represents a logical grouping of fields
type Section struct {
	ID          string                    `json:"id" validate:"required" example:"customer-info"`
	Title       string                    `json:"title" validate:"required" example:"Customer Information"`
	Description string                    `json:"description,omitempty" validate:"max=500"`
	Fields      []string                  `json:"fields" validate:"required,dive,fieldname"`
	Icon        string                    `json:"icon,omitempty" validate:"icon_name"`
	Collapsible bool                      `json:"collapsible,omitempty"`
	Collapsed   bool                      `json:"collapsed,omitempty"`
	Order       int                       `json:"order,omitempty"`
	Conditional *Conditional              `json:"conditional,omitempty"`
	Permissions *SectionPermissions       `json:"permissions,omitempty"`
}

// Group represents a visual grouping of fields
type Group struct {
	ID      string   `json:"id" validate:"required" example:"basic-info"`
	Title   string   `json:"title,omitempty" example:"Basic Information"`
	Fields  []string `json:"fields" validate:"required,dive,fieldname"`
	Layout  string   `json:"layout" validate:"oneof=horizontal vertical grid" example:"horizontal"`
	Border  bool     `json:"border,omitempty"`
	Padding string   `json:"padding,omitempty" validate:"css_size"`
	Order   int      `json:"order,omitempty"`
}

// Tab represents a tab in tabbed layouts
type Tab struct {
	ID          string                    `json:"id" validate:"required" example:"basic-tab"`
	Title       string                    `json:"title" validate:"required" example:"Basic Info"`
	Icon        string                    `json:"icon,omitempty" validate:"icon_name"`
	Fields      []string                  `json:"fields,omitempty" validate:"dive,fieldname"`
	Sections    []string                  `json:"sections,omitempty" validate:"dive,fieldname"`
	Badge       string                    `json:"badge,omitempty"`
	Disabled    bool                      `json:"disabled,omitempty"`
	Order       int                       `json:"order,omitempty"`
	Conditional *Conditional              `json:"conditional,omitempty"`
}

// Step represents a step in multi-step forms
type Step struct {
	ID          string                    `json:"id" validate:"required" example:"step-1"`
	Title       string                    `json:"title" validate:"required" example:"Basic Information"`
	Description string                    `json:"description,omitempty" validate:"max=500"`
	Fields      []string                  `json:"fields,omitempty" validate:"dive,fieldname"`
	Sections    []string                  `json:"sections,omitempty" validate:"dive,fieldname"`
	Optional    bool                      `json:"optional,omitempty"`
	Order       int                       `json:"order,omitempty"`
	Validation  *StepValidation           `json:"validation,omitempty"`
	Conditional *Conditional              `json:"conditional,omitempty"`
}

// Breakpoints defines responsive breakpoints
type Breakpoints struct {
	Mobile  *BreakpointConfig `json:"mobile,omitempty"`  // < 768px
	Tablet  *BreakpointConfig `json:"tablet,omitempty"`  // >= 768px
	Desktop *BreakpointConfig `json:"desktop,omitempty"` // >= 1024px
	Wide    *BreakpointConfig `json:"wide,omitempty"`    // >= 1280px
}

// BreakpointConfig defines layout at specific breakpoint
type BreakpointConfig struct {
	Columns int    `json:"columns" validate:"min=1,max=24"`
	Gap     string `json:"gap,omitempty" validate:"css_size"`
	Hidden  bool   `json:"hidden,omitempty"`
}

// FieldLayout defines individual field positioning
type FieldLayout struct {
	ColSpan  int    `json:"colSpan,omitempty" validate:"min=1,max=24" example:"2"`
	Row      int    `json:"row,omitempty" validate:"min=1"`
	Order    int    `json:"order,omitempty"`
	Width    string `json:"width,omitempty" validate:"css_size"`
	MinWidth string `json:"minWidth,omitempty" validate:"css_size"`
	MaxWidth string `json:"maxWidth,omitempty" validate:"css_size"`
	Height   string `json:"height,omitempty" validate:"css_size"`
	Class    string `json:"class,omitempty" validate:"css_class"`
	Style    string `json:"style,omitempty" validate:"css_style"`
}

// Style defines visual styling
type Style struct {
	Background string `json:"background,omitempty" validate:"css_color"`
	Color      string `json:"color,omitempty" validate:"css_color"`
	Border     string `json:"border,omitempty" validate:"css_border"`
	Padding    string `json:"padding,omitempty" validate:"css_size"`
	Margin     string `json:"margin,omitempty" validate:"css_size"`
	FontSize   string `json:"fontSize,omitempty" validate:"css_size"`
	FontWeight string `json:"fontWeight,omitempty" validate:"css_font_weight"`
	TextAlign  string `json:"textAlign,omitempty" validate:"oneof=left center right justify"`
	Class      string `json:"class,omitempty" validate:"css_class"`
	Custom     string `json:"custom,omitempty" validate:"css_style"`
}

// FieldValidation defines field-level validation rules
type FieldValidation struct {
	Required    bool                      `json:"required,omitempty"`
	MinLength   *int                      `json:"minLength,omitempty" validate:"min=0"`
	MaxLength   *int                      `json:"maxLength,omitempty" validate:"min=0"`
	Min         *float64                  `json:"min,omitempty"`
	Max         *float64                  `json:"max,omitempty"`
	Pattern     string                    `json:"pattern,omitempty" validate:"regexp"`
	Email       bool                      `json:"email,omitempty"`
	URL         bool                      `json:"url,omitempty"`
	Unique      bool                      `json:"unique,omitempty"`
	Custom      []string                  `json:"custom,omitempty"`
	CrossField  []CrossFieldRule          `json:"crossField,omitempty" validate:"dive"`
	Async       *AsyncValidation          `json:"async,omitempty"`
	Conditions  *condition.ConditionGroup `json:"conditions,omitempty"`
	Formula     string                    `json:"formula,omitempty" validate:"js_expression"`
	Messages    map[string]string         `json:"messages,omitempty"`
}

// Transform defines data transformation
type Transform struct {
	Input  string `json:"input,omitempty" validate:"js_function"`
	Output string `json:"output,omitempty" validate:"js_function"`
	Format string `json:"format,omitempty"`
}

// Mask defines input masking
type Mask struct {
	Pattern     string                 `json:"pattern" validate:"required"`
	Placeholder string                 `json:"placeholder,omitempty"`
	Type        string                 `json:"type,omitempty" validate:"oneof=phone currency date custom"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

// FieldEvents defines field-level event handlers
type FieldEvents struct {
	OnChange   string `json:"onChange,omitempty" validate:"js_function"`
	OnFocus    string `json:"onFocus,omitempty" validate:"js_function"`
	OnBlur     string `json:"onBlur,omitempty" validate:"js_function"`
	OnInput    string `json:"onInput,omitempty" validate:"js_function"`
	OnKeyPress string `json:"onKeyPress,omitempty" validate:"js_function"`
	OnKeyUp    string `json:"onKeyUp,omitempty" validate:"js_function"`
	OnKeyDown  string `json:"onKeyDown,omitempty" validate:"js_function"`
	OnClick    string `json:"onClick,omitempty" validate:"js_function"`
	OnMount    string `json:"onMount,omitempty" validate:"js_function"`
	OnUnmount  string `json:"onUnmount,omitempty" validate:"js_function"`
	Debounce   int    `json:"debounce,omitempty" validate:"min=0,max=10000"`
	Throttle   int    `json:"throttle,omitempty" validate:"min=0,max=10000"`
}

// Conditional defines conditional logic
type Conditional struct {
	Show    *condition.ConditionGroup `json:"show,omitempty"`
	Hide    *condition.ConditionGroup `json:"hide,omitempty"`
	Enable  *condition.ConditionGroup `json:"enable,omitempty"`
	Disable *condition.ConditionGroup `json:"disable,omitempty"`
	Require *condition.ConditionGroup `json:"require,omitempty"`
	
	// Formula-based conditions (simplified expressions)
	ShowIf     string `json:"showIf,omitempty" validate:"js_expression"`
	HideIf     string `json:"hideIf,omitempty" validate:"js_expression"`
	EnableIf   string `json:"enableIf,omitempty" validate:"js_expression"`
	DisableIf  string `json:"disableIf,omitempty" validate:"js_expression"`
	RequireIf  string `json:"requireIf,omitempty" validate:"js_expression"`
	ValidateIf string `json:"validateIf,omitempty" validate:"js_expression"`
}

// DataSource defines dynamic data loading
type DataSource struct {
	Type        DataSourceType    `json:"type" validate:"required"`
	URL         string            `json:"url,omitempty" validate:"url"`
	Method      string            `json:"method,omitempty" validate:"oneof=GET POST PUT PATCH DELETE"`
	Headers     map[string]string `json:"headers,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
	Transform   string            `json:"transform,omitempty" validate:"js_function"`
	Cache       bool              `json:"cache,omitempty"`
	CacheTTL    int               `json:"cacheTTL,omitempty" validate:"min=0"`
	LazyLoad    bool              `json:"lazyLoad,omitempty"`
	SearchField string            `json:"searchField,omitempty" validate:"fieldname"`
	ValueField  string            `json:"valueField,omitempty" validate:"fieldname"`
	LabelField  string            `json:"labelField,omitempty" validate:"fieldname"`
	Tenant      bool              `json:"tenant,omitempty"`
	DependsOn   []string          `json:"dependsOn,omitempty" validate:"dive,fieldname"`
	Data        []Option          `json:"data,omitempty" validate:"dive"`
}

// DataSourceType defines data source types
type DataSourceType string

const (
	DataSourceStatic   DataSourceType = "static"
	DataSourceAPI      DataSourceType = "api"
	DataSourceDatabase DataSourceType = "database"
	DataSourceFunction DataSourceType = "function"
)

// FieldHTMX defines field-level HTMX configuration
type FieldHTMX struct {
	Get     string            `json:"get,omitempty" validate:"url"`
	Post    string            `json:"post,omitempty" validate:"url"`
	Trigger string            `json:"trigger,omitempty"`
	Target  string            `json:"target,omitempty" validate:"css_selector"`
	Swap    string            `json:"swap,omitempty" validate:"oneof=innerHTML outerHTML beforebegin afterbegin beforeend afterend"`
	Delay   string            `json:"delay,omitempty" validate:"duration"`
	Include string            `json:"include,omitempty" validate:"css_selector"`
	Headers map[string]string `json:"headers,omitempty"`
}

// FieldAlpine defines field-level Alpine.js configuration
type FieldAlpine struct {
	XModel      string            `json:"xModel,omitempty" validate:"js_variable"`
	XShow       string            `json:"xShow,omitempty" validate:"js_expression"`
	XIf         string            `json:"xIf,omitempty" validate:"js_expression"`
	XBind       map[string]string `json:"xBind,omitempty"`
	XOn         map[string]string `json:"xOn,omitempty"`
	XText       string            `json:"xText,omitempty" validate:"js_expression"`
	XHTML       string            `json:"xHtml,omitempty" validate:"js_expression"`
	XRef        string            `json:"xRef,omitempty" validate:"js_variable"`
	XTransition string            `json:"xTransition,omitempty"`
}

// Permission types for different schema elements
type FieldPermissions struct {
	View   *condition.ConditionGroup `json:"view,omitempty"`
	Edit   *condition.ConditionGroup `json:"edit,omitempty"`
	Roles  []string                  `json:"roles,omitempty" validate:"dive,alphanum"`
	Groups []string                  `json:"groups,omitempty" validate:"dive,alphanum"`
}

type SectionPermissions struct {
	View   *condition.ConditionGroup `json:"view,omitempty"`
	Edit   *condition.ConditionGroup `json:"edit,omitempty"`
	Roles  []string                  `json:"roles,omitempty" validate:"dive,alphanum"`
	Groups []string                  `json:"groups,omitempty" validate:"dive,alphanum"`
}

type ActionPermissions struct {
	Execute *condition.ConditionGroup `json:"execute,omitempty"`
	Roles   []string                  `json:"roles,omitempty" validate:"dive,alphanum"`
	Groups  []string                  `json:"groups,omitempty" validate:"dive,alphanum"`
}

// Action configuration types
type ActionConfig struct {
	URL         string            `json:"url,omitempty" validate:"url"`
	Method      string            `json:"method,omitempty" validate:"oneof=GET POST PUT PATCH DELETE"`
	Target      string            `json:"target,omitempty" validate:"css_selector"`
	Redirect    string            `json:"redirect,omitempty" validate:"url"`
	Download    bool              `json:"download,omitempty"`
	NewWindow   bool              `json:"newWindow,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Payload     interface{}       `json:"payload,omitempty"`
	Transform   string            `json:"transform,omitempty" validate:"js_function"`
	OnSuccess   string            `json:"onSuccess,omitempty" validate:"js_function"`
	OnError     string            `json:"onError,omitempty" validate:"js_function"`
	Timeout     int               `json:"timeout,omitempty" validate:"min=0,max=300000"`
}

type Confirm struct {
	Title       string `json:"title" validate:"required,max=100"`
	Message     string `json:"message" validate:"required,max=500"`
	ConfirmText string `json:"confirmText,omitempty" validate:"max=50"`
	CancelText  string `json:"cancelText,omitempty" validate:"max=50"`
	Type        string `json:"type,omitempty" validate:"oneof=info warning danger success"`
	Icon        string `json:"icon,omitempty" validate:"icon_name"`
}

type ActionHTMX struct {
	Post      string            `json:"post,omitempty" validate:"url"`
	Get       string            `json:"get,omitempty" validate:"url"`
	Target    string            `json:"target,omitempty" validate:"css_selector"`
	Swap      string            `json:"swap,omitempty" validate:"oneof=innerHTML outerHTML beforebegin afterbegin beforeend afterend"`
	Indicator string            `json:"indicator,omitempty" validate:"css_selector"`
	Confirm   string            `json:"confirm,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type ActionAlpine struct {
	XOn   map[string]string `json:"xOn,omitempty"`
	XBind map[string]string `json:"xBind,omitempty"`
}

// Security types
type CSRF struct {
	Enabled    bool   `json:"enabled"`
	TokenField string `json:"tokenField,omitempty" validate:"fieldname"`
	HeaderName string `json:"headerName,omitempty" validate:"http_header"`
}

type RateLimit struct {
	Enabled       bool   `json:"enabled"`
	MaxRequests   int    `json:"maxRequests" validate:"min=1"`
	WindowSeconds int    `json:"windowSeconds" validate:"min=1"`
	Strategy      string `json:"strategy" validate:"oneof=sliding fixed"`
}

type Encryption struct {
	Enabled         bool     `json:"enabled"`
	EncryptedFields []string `json:"encryptedFields,omitempty" validate:"dive,fieldname"`
	Algorithm       string   `json:"algorithm,omitempty" validate:"oneof=AES-256-GCM ChaCha20-Poly1305"`
}

// Workflow types
type WorkflowAction struct {
	ID          string                    `json:"id" validate:"required,uuid"`
	Label       string                    `json:"label" validate:"required,max=100"`
	Type        string                    `json:"type" validate:"required,oneof=approve reject submit cancel escalate"`
	Icon        string                    `json:"icon,omitempty" validate:"icon_name"`
	Button      Action                    `json:"button"`
	Conditions  *condition.ConditionGroup `json:"conditions,omitempty"`
	Confirm     *Confirm                  `json:"confirm,omitempty"`
	OnExecute   string                    `json:"onExecute,omitempty" validate:"js_function"`
	Permissions *ActionPermissions        `json:"permissions,omitempty"`
	Order       int                       `json:"order,omitempty"`
}

type WorkflowTransition struct {
	From       string                    `json:"from" validate:"required"`
	To         string                    `json:"to" validate:"required"`
	Action     string                    `json:"action" validate:"required"`
	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
	Formula    string                    `json:"formula,omitempty" validate:"js_expression"`
	OnSuccess  string                    `json:"onSuccess,omitempty" validate:"js_function"`
	OnError    string                    `json:"onError,omitempty" validate:"js_function"`
	Timeout    int                       `json:"timeout,omitempty" validate:"min=0"`
}

type Notification struct {
	ID         string                    `json:"id" validate:"required,uuid"`
	Event      string                    `json:"event" validate:"required"`
	Recipients []string                  `json:"recipients" validate:"required,dive,email"`
	Template   string                    `json:"template" validate:"required"`
	Channels   []string                  `json:"channels" validate:"required,dive,oneof=email sms push webhook"`
	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
	Delay      int                       `json:"delay,omitempty" validate:"min=0"`
}

type ApprovalConfig struct {
	Required   bool                      `json:"required"`
	Type       string                    `json:"type" validate:"oneof=sequential parallel any"`
	Approvers  []string                  `json:"approvers" validate:"dive,uuid"`
	MinVotes   int                       `json:"minVotes,omitempty" validate:"min=1"`
	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
	Timeout    int                       `json:"timeout,omitempty" validate:"min=0"`
	Escalation *EscalationConfig         `json:"escalation,omitempty"`
}

type EscalationConfig struct {
	After      int      `json:"after" validate:"min=1"`
	Recipients []string `json:"recipients" validate:"required,dive,uuid"`
	Action     string   `json:"action,omitempty" validate:"oneof=auto-approve notify-admin escalate-higher"`
}

// Validation types
type Validation struct {
	Mode            ValidationMode     `json:"mode" validate:"oneof=onChange onBlur onSubmit all"`
	RevalidateMode  ValidationMode     `json:"revalidateMode" validate:"oneof=onChange onBlur onSubmit all"`
	ShowErrors      string             `json:"showErrors" validate:"oneof=all touched submitted"`
	FocusOnError    bool               `json:"focusOnError,omitempty"`
	ScrollToError   bool               `json:"scrollToError,omitempty"`
	CrossFieldRules []CrossFieldRule   `json:"crossFieldRules,omitempty" validate:"dive"`
	CustomRules     []CustomRule       `json:"customRules,omitempty" validate:"dive"`
	AsyncRules      []AsyncValidation  `json:"asyncRules,omitempty" validate:"dive"`
	OnValidate      string             `json:"onValidate,omitempty" validate:"js_function"`
}

type ValidationMode string

const (
	ValidationOnChange ValidationMode = "onChange"
	ValidationOnBlur   ValidationMode = "onBlur"
	ValidationOnSubmit ValidationMode = "onSubmit"
	ValidationAll      ValidationMode = "all"
)

type CrossFieldRule struct {
	ID         string                    `json:"id" validate:"required"`
	Message    string                    `json:"message" validate:"required,max=200"`
	Fields     []string                  `json:"fields" validate:"required,dive,fieldname"`
	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
	Formula    string                    `json:"formula,omitempty" validate:"js_expression"`
}

type CustomRule struct {
	ID       string `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Function string `json:"function" validate:"required,js_function"`
	Message  string `json:"message" validate:"required,max=200"`
	Async    bool   `json:"async,omitempty"`
}

type AsyncValidation struct {
	URL         string            `json:"url" validate:"required,url"`
	Method      string            `json:"method" validate:"oneof=GET POST PUT PATCH"`
	Headers     map[string]string `json:"headers,omitempty"`
	Debounce    int               `json:"debounce,omitempty" validate:"min=0,max=10000"`
	ValidateOn  []string          `json:"validateOn,omitempty" validate:"dive,oneof=change blur submit"`
	OnSuccess   string            `json:"onSuccess,omitempty" validate:"js_function"`
	OnError     string            `json:"onError,omitempty" validate:"js_function"`
	CacheResult bool              `json:"cacheResult,omitempty"`
	Timeout     int               `json:"timeout,omitempty" validate:"min=0,max=30000"`
}

type StepValidation struct {
	Required   bool                      `json:"required"`
	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
	OnNext     string                    `json:"onNext,omitempty" validate:"js_function"`
	OnPrevious string                    `json:"onPrevious,omitempty" validate:"js_function"`
	Custom     []CustomRule              `json:"custom,omitempty" validate:"dive"`
}

// I18n types
type I18n struct {
	Enabled          bool              `json:"enabled"`
	DefaultLocale    string            `json:"defaultLocale" validate:"locale"`
	SupportedLocales []string          `json:"supportedLocales" validate:"dive,locale"`
	Translations     map[string]string `json:"translations,omitempty"`
	DateFormat       string            `json:"dateFormat,omitempty"`
	TimeFormat       string            `json:"timeFormat,omitempty"`
	NumberFormat     string            `json:"numberFormat,omitempty"`
	CurrencyFormat   string            `json:"currencyFormat,omitempty"`
	RTL              bool              `json:"rtl,omitempty"`
}

// Changelog for version tracking
type ChangelogEntry struct {
	Version     string    `json:"version" validate:"required,semver"`
	Date        time.Time `json:"date" validate:"required"`
	Author      string    `json:"author" validate:"required,uuid"`
	Changes     []string  `json:"changes" validate:"required,dive,max=200"`
	Type        string    `json:"type" validate:"oneof=major minor patch"`
	Breaking    bool      `json:"breaking,omitempty"`
	Description string    `json:"description,omitempty" validate:"max=500"`
}

// Helper functions for pointer types
func IntPtr(i int) *int {
	return &i
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func StringPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

// Validation helper methods
func (f *Field) HasValidation() bool {
	return f.Validation != nil
}

func (f *Field) HasDataSource() bool {
	return f.DataSource != nil
}

func (f *Field) HasConditional() bool {
	return f.Conditional != nil
}

func (s *Schema) HasWorkflow() bool {
	return s.Workflow != nil && s.Workflow.Enabled
}

func (s *Schema) HasMultiTenant() bool {
	return s.Tenant != nil && s.Tenant.Enabled
}

func (s *Schema) HasHTMX() bool {
	return s.HTMX != nil && s.HTMX.Enabled
}

func (s *Schema) HasAlpine() bool {
	return s.Alpine != nil && s.Alpine.Enabled
}