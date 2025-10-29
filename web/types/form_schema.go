package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// FormSchema defines the structure for dynamically generating forms
type FormSchema struct {
	ID          string        `json:"id"`
	Version     string        `json:"version,omitempty"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Action      string        `json:"action"`
	Method      string        `json:"method"`
	Fields      []FieldSchema `json:"fields"`
	Submit      ButtonSchema  `json:"submit"`
	Cancel      *ButtonSchema `json:"cancel,omitempty"`
	Reset       *ButtonSchema `json:"reset,omitempty"`
	Layout      FormLayout    `json:"layout"`

	// Multi-tenancy support
	TenantID    string      `json:"tenantId,omitempty"`
	TenantScope TenantScope `json:"tenantScope,omitempty"`

	// Conditional rendering using condition package
	VisibilityRules *condition.ConditionGroup `json:"visibilityRules,omitempty"`
	EnabledRules    *condition.ConditionGroup `json:"enabledRules,omitempty"`

	// HTMX integration
	HTMXConfig *HTMXConfig `json:"htmx,omitempty"`

	// Alpine.js integration
	AlpineConfig *AlpineConfig `json:"alpine,omitempty"`

	// Enterprise features
	Security    SecurityConfig  `json:"security"`
	Workflow    *WorkflowConfig `json:"workflow,omitempty"`
	Audit       AuditConfig     `json:"audit"`
	I18n        *I18nConfig     `json:"i18n,omitempty"`
	Permissions []string        `json:"permissions,omitempty"`

	// UI/UX ents
	Notifications NotificationConfig `json:"notifications"`
	Loading       LoadingConfig      `json:"loading"`
	Validation    FormValidation     `json:"validation"`

	// Data handling
	InitialData   map[string]any       `json:"initialData,omitempty"`
	DataTransform *DataTransformConfig `json:"dataTransform,omitempty"`

	// Metadata
	Tags      []string  `json:"tags,omitempty"`
	Category  string    `json:"category,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	CreatedBy string    `json:"createdBy,omitempty"`
	UpdatedBy string    `json:"updatedBy,omitempty"`
}

// TenantScope defines multi-tenant data scope
type TenantScope string

const (
	TenantScopeGlobal     TenantScope = "global"     // Available to all tenants
	TenantScopeOrg        TenantScope = "org"        // Organization level
	TenantScopeDepartment TenantScope = "department" // Department level
	TenantScopeUser       TenantScope = "user"       // User level
)

// FieldSchema defines individual form field specifications
type FieldSchema struct {
	Name         string            `json:"name"`
	Type         FieldType         `json:"type"`
	Label        string            `json:"label"`
	Placeholder  string            `json:"placeholder,omitempty"`
	Required     bool              `json:"required"`
	Disabled     bool              `json:"disabled"`
	Readonly     bool              `json:"readonly"`
	Value        any               `json:"value,omitempty"`
	DefaultValue any               `json:"defaultValue,omitempty"`
	Validation   ValidationRules   `json:"validation,omitempty"`
	Options      []OptionSchema    `json:"options,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
	HelpText     string            `json:"helpText,omitempty"`
	ErrorText    string            `json:"errorText,omitempty"`
	Layout       FieldLayout       `json:"layout"`

	// Conditional logic using condition package
	VisibilityRules *condition.ConditionGroup `json:"visibilityRules,omitempty"`
	EnabledRules    *condition.ConditionGroup `json:"enabledRules,omitempty"`
	RequiredRules   *condition.ConditionGroup `json:"requiredRules,omitempty"`

	// Dynamic behavior
	DependsOn []string        `json:"dependsOn,omitempty"`
	Triggers  []TriggerConfig `json:"triggers,omitempty"`

	// HTMX field-level config
	HTMXConfig *HTMXFieldConfig `json:"htmx,omitempty"`

	// Alpine.js field-level config
	AlpineConfig *AlpineFieldConfig `json:"alpine,omitempty"`

	// Data source for dynamic fields
	DataSource *DataSourceConfig `json:"dataSource,omitempty"`

	// Field-specific features
	Prefix  string `json:"prefix,omitempty"`
	Suffix  string `json:"suffix,omitempty"`
	Icon    string `json:"icon,omitempty"`
	Tooltip string `json:"tooltip,omitempty"`

	// Multi-tenant field config
	TenantRestricted bool `json:"tenantRestricted"`

	// Formatting
	Format *FormatConfig `json:"format,omitempty"`
	Mask   string        `json:"mask,omitempty"`

	// Permissions
	ViewPermissions []string `json:"viewPermissions,omitempty"`
	EditPermissions []string `json:"editPermissions,omitempty"`
}

// HTMXConfig defines HTMX-specific configuration
type HTMXConfig struct {
	Enabled bool `json:"enabled"`

	// Form submission
	PostURL      string   `json:"postUrl,omitempty"`
	GetURL       string   `json:"getUrl,omitempty"`
	SwapStrategy HTMXSwap `json:"swap"`
	Target       string   `json:"target,omitempty"`

	// Behavior
	PushURL    bool `json:"pushUrl"`
	ReplaceURL bool `json:"replaceUrl"`
	Boost      bool `json:"boost"`

	// Loading states
	Indicator string `json:"indicator,omitempty"`

	// Triggers
	Trigger string `json:"trigger,omitempty"`

	// Headers
	Headers map[string]string `json:"headers,omitempty"`

	// Validation
	Validate bool `json:"validate"`

	// Response handling
	SuccessTarget string `json:"successTarget,omitempty"`
	ErrorTarget   string `json:"errorTarget,omitempty"`

	// Extensions
	Extensions []string `json:"extensions,omitempty"`
}

// HTMXSwap defines HTMX swap strategies
type HTMXSwap string

const (
	HTMXSwapInnerHTML   HTMXSwap = "innerHTML"
	HTMXSwapOuterHTML   HTMXSwap = "outerHTML"
	HTMXSwapBeforeBegin HTMXSwap = "beforebegin"
	HTMXSwapAfterBegin  HTMXSwap = "afterbegin"
	HTMXSwapBeforeEnd   HTMXSwap = "beforeend"
	HTMXSwapAfterEnd    HTMXSwap = "afterend"
	HTMXSwapDelete      HTMXSwap = "delete"
	HTMXSwapNone        HTMXSwap = "none"
)

// HTMXFieldConfig defines field-level HTMX configuration
type HTMXFieldConfig struct {
	// Auto-complete/search
	GetURL  string   `json:"getUrl,omitempty"`
	Trigger string   `json:"trigger,omitempty"`
	Target  string   `json:"target,omitempty"`
	Swap    HTMXSwap `json:"swap,omitempty"`

	// Debouncing
	Delay string `json:"delay,omitempty"` // e.g., "500ms"

	// Dependencies
	Include string `json:"include,omitempty"`

	// Sync
	Sync string `json:"sync,omitempty"`
}

// AlpineConfig defines Alpine.js-specific configuration
type AlpineConfig struct {
	Enabled bool `json:"enabled"`

	// Data model
	XData string `json:"xData,omitempty"`

	// Reactive state
	XModelModifiers []string `json:"xModelModifiers,omitempty"` // debounce, throttle, lazy

	// Initialization
	XInit string `json:"xInit,omitempty"`

	// Effects
	XEffect string `json:"xEffect,omitempty"`

	// Custom directives
	Directives map[string]string `json:"directives,omitempty"`

	// Stores
	Store string `json:"store,omitempty"`

	// Cloak
	Cloak bool `json:"cloak"`
}

// AlpineFieldConfig defines field-level Alpine.js configuration
type AlpineFieldConfig struct {
	// Binding
	XModel          string   `json:"xModel,omitempty"`
	XModelModifiers []string `json:"xModelModifiers,omitempty"`

	// Display
	XShow string `json:"xShow,omitempty"`
	XIf   string `json:"xIf,omitempty"`

	// Events
	XOn map[string]string `json:"xOn,omitempty"` // event: handler

	// Binding
	XBind map[string]string `json:"xBind,omitempty"` // attr: expression

	// Text/HTML
	XText string `json:"xText,omitempty"`
	XHTML string `json:"xHtml,omitempty"`

	// Ref
	XRef string `json:"xRef,omitempty"`

	// Transition
	XTransition *TransitionConfig `json:"xTransition,omitempty"`
}

// TransitionConfig defines Alpine.js transition configuration
type TransitionConfig struct {
	Enter      string `json:"enter,omitempty"`
	EnterStart string `json:"enterStart,omitempty"`
	EnterEnd   string `json:"enterEnd,omitempty"`
	Leave      string `json:"leave,omitempty"`
	LeaveStart string `json:"leaveStart,omitempty"`
	LeaveEnd   string `json:"leaveEnd,omitempty"`
}

// TriggerConfig defines field triggers
type TriggerConfig struct {
	Event     TriggerEvent              `json:"event"`
	Action    TriggerAction             `json:"action"`
	Target    string                    `json:"target,omitempty"`
	Value     any                       `json:"value,omitempty"`
	Condition *condition.ConditionGroup `json:"condition,omitempty"`
	Debounce  int                       `json:"debounce,omitempty"` // milliseconds
}

// TriggerEvent defines trigger events
type TriggerEvent string

const (
	TriggerEventChange TriggerEvent = "change"
	TriggerEventBlur   TriggerEvent = "blur"
	TriggerEventFocus  TriggerEvent = "focus"
	TriggerEventInput  TriggerEvent = "input"
	TriggerEventClick  TriggerEvent = "click"
)

// TriggerAction defines trigger actions
type TriggerAction string

const (
	TriggerActionShow       TriggerAction = "show"
	TriggerActionHide       TriggerAction = "hide"
	TriggerActionEnable     TriggerAction = "enable"
	TriggerActionDisable    TriggerAction = "disable"
	TriggerActionSetValue   TriggerAction = "setValue"
	TriggerActionClearValue TriggerAction = "clearValue"
	TriggerActionValidate   TriggerAction = "validate"
	TriggerActionFetch      TriggerAction = "fetch"
	TriggerActionSubmit     TriggerAction = "submit"
)

// DataSourceConfig defines dynamic data source configuration
type DataSourceConfig struct {
	Type    DataSourceType    `json:"type"`
	URL     string            `json:"url,omitempty"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Params  map[string]string `json:"params,omitempty"`

	// Static data
	Data []OptionSchema `json:"data,omitempty"`

	// Transform response
	Transform string `json:"transform,omitempty"` // JSONPath or JS expression

	// Caching
	Cache         bool `json:"cache"`
	CacheDuration int  `json:"cacheDuration,omitempty"` // seconds

	// Filtering
	SearchParam string `json:"searchParam,omitempty"`
	FilterParam string `json:"filterParam,omitempty"`

	// Pagination
	Paginated bool `json:"paginated"`
	PageSize  int  `json:"pageSize,omitempty"`

	// Dependencies
	DependsOn []string `json:"dependsOn,omitempty"`

	// Multi-tenant filtering
	TenantFiltered bool `json:"tenantFiltered"`
}

// DataSourceType defines data source types
type DataSourceType string

const (
	DataSourceTypeStatic   DataSourceType = "static"
	DataSourceTypeAPI      DataSourceType = "api"
	DataSourceTypeFunction DataSourceType = "function"
)

// SecurityConfig defines form security settings
type SecurityConfig struct {
	CSRF           CSRFConfig       `json:"csrf"`
	RateLimiting   *RateLimitConfig `json:"rateLimiting,omitempty"`
	Encryption     EncryptionConfig `json:"encryption"`
	SanitizeInput  bool             `json:"sanitizeInput"`
	AllowedOrigins []string         `json:"allowedOrigins,omitempty"`
}

// CSRFConfig defines CSRF protection
type CSRFConfig struct {
	Enabled    bool   `json:"enabled"`
	TokenField string `json:"tokenField"`
	HeaderName string `json:"headerName"`
}

// RateLimitConfig defines rate limiting
type RateLimitConfig struct {
	Enabled       bool   `json:"enabled"`
	MaxRequests   int    `json:"maxRequests"`
	WindowSeconds int    `json:"windowSeconds"`
	Strategy      string `json:"strategy"` // "sliding", "fixed"
}

// EncryptionConfig defines field encryption
type EncryptionConfig struct {
	Enabled         bool     `json:"enabled"`
	EncryptedFields []string `json:"encryptedFields,omitempty"`
	Algorithm       string   `json:"algorithm,omitempty"`
}

// WorkflowConfig defines workflow integration
type WorkflowConfig struct {
	Enabled          bool            `json:"enabled"`
	WorkflowID       string          `json:"workflowId,omitempty"`
	Stages           []WorkflowStage `json:"stages,omitempty"`
	ApprovalRequired bool            `json:"approvalRequired"`
	Approvers        []string        `json:"approvers,omitempty"`
}

// WorkflowStage defines workflow stages
type WorkflowStage struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	Order     int                       `json:"order"`
	Condition *condition.ConditionGroup `json:"condition,omitempty"`
	Actions   []WorkflowAction          `json:"actions,omitempty"`
}

// WorkflowAction defines workflow actions
type WorkflowAction struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config,omitempty"`
}

// AuditConfig defines audit logging
type AuditConfig struct {
	Enabled        bool     `json:"enabled"`
	LogChanges     bool     `json:"logChanges"`
	LogViews       bool     `json:"logViews"`
	LogSubmissions bool     `json:"logSubmissions"`
	ExcludeFields  []string `json:"excludeFields,omitempty"`
	RetentionDays  int      `json:"retentionDays,omitempty"`
}

// I18nConfig defines internationalization
type I18nConfig struct {
	Enabled        bool                         `json:"enabled"`
	DefaultLocale  string                       `json:"defaultLocale"`
	Locales        []string                     `json:"locales"`
	Translations   map[string]map[string]string `json:"translations,omitempty"`
	FallbackLocale string                       `json:"fallbackLocale,omitempty"`
}

// NotificationConfig defines notification settings
type NotificationConfig struct {
	Success     NotificationStyle    `json:"success"`
	Error       NotificationStyle    `json:"error"`
	Warning     NotificationStyle    `json:"warning"`
	Info        NotificationStyle    `json:"info"`
	Position    NotificationPosition `json:"position"`
	Duration    int                  `json:"duration"` // milliseconds
	Dismissible bool                 `json:"dismissible"`
}

// NotificationStyle defines notification styling
type NotificationStyle struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message,omitempty"`
	Icon    string `json:"icon,omitempty"`
	Sound   bool   `json:"sound"`
}

// NotificationPosition defines notification position
type NotificationPosition string

const (
	NotificationPositionTopRight     NotificationPosition = "top-right"
	NotificationPositionTopLeft      NotificationPosition = "top-left"
	NotificationPositionTopCenter    NotificationPosition = "top-center"
	NotificationPositionBottomRight  NotificationPosition = "bottom-right"
	NotificationPositionBottomLeft   NotificationPosition = "bottom-left"
	NotificationPositionBottomCenter NotificationPosition = "bottom-center"
)

// LoadingConfig defines loading states
type LoadingConfig struct {
	Enabled     bool   `json:"enabled"`
	Type        string `json:"type"` // "spinner", "progress", "skeleton"
	Message     string `json:"message,omitempty"`
	Overlay     bool   `json:"overlay"`
	DisableForm bool   `json:"disableForm"`
}

// FormValidation defines form-level validation
type FormValidation struct {
	Mode            ValidationMode    `json:"mode"`
	RevalidateMode  ValidationMode    `json:"revalidateMode"`
	ShowErrorsOn    ValidationTrigger `json:"showErrorsOn"`
	ScrollToError   bool              `json:"scrollToError"`
	FocusFirstError bool              `json:"focusFirstError"`
}

// ValidationMode defines when validation occurs
type ValidationMode string

const (
	ValidationModeOnChange ValidationMode = "onChange"
	ValidationModeOnBlur   ValidationMode = "onBlur"
	ValidationModeOnSubmit ValidationMode = "onSubmit"
	ValidationModeAll      ValidationMode = "all"
)

// ValidationTrigger defines when to show errors
type ValidationTrigger string

const (
	ValidationTriggerImmediate ValidationTrigger = "immediate"
	ValidationTriggerTouch     ValidationTrigger = "touch"
	ValidationTriggerSubmit    ValidationTrigger = "submit"
)

// DataTransformConfig defines data transformation
type DataTransformConfig struct {
	BeforeSubmit string            `json:"beforeSubmit,omitempty"` // JS function name
	AfterLoad    string            `json:"afterLoad,omitempty"`    // JS function name
	FieldMapping map[string]string `json:"fieldMapping,omitempty"` // source: target
}

// FormatConfig defines field formatting
type FormatConfig struct {
	Type               FormatType `json:"type"`
	Pattern            string     `json:"pattern,omitempty"`
	DecimalPlaces      int        `json:"decimalPlaces,omitempty"`
	ThousandsSeparator string     `json:"thousandsSeparator,omitempty"`
	DecimalSeparator   string     `json:"decimalSeparator,omitempty"`
	Prefix             string     `json:"prefix,omitempty"`
	Suffix             string     `json:"suffix,omitempty"`
	DateFormat         string     `json:"dateFormat,omitempty"`
	TimeFormat         string     `json:"timeFormat,omitempty"`
}

// FormatType defines formatting types
type FormatType string

const (
	FormatTypeNumber   FormatType = "number"
	FormatTypeCurrency FormatType = "currency"
	FormatTypePercent  FormatType = "percent"
	FormatTypeDate     FormatType = "date"
	FormatTypeTime     FormatType = "time"
	FormatTypePhone    FormatType = "phone"
	FormatTypeCustom   FormatType = "custom"
)

// ValidationRules with condition package integration
type ValidationRules struct {
	MinLength    *int     `json:"minLength,omitempty"`
	MaxLength    *int     `json:"maxLength,omitempty"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
	Pattern      string   `json:"pattern,omitempty"`
	CustomRules  []string `json:"customRules,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`

	// Email validation
	Email bool `json:"email,omitempty"`

	// URL validation
	URL bool `json:"url,omitempty"`

	// File validation
	MinFiles     *int     `json:"minFiles,omitempty"`
	MaxFiles     *int     `json:"maxFiles,omitempty"`
	MaxFileSize  *int64   `json:"maxFileSize,omitempty"` // bytes
	AllowedTypes []string `json:"allowedTypes,omitempty"`

	// Conditional validation using condition package
	ConditionalRules []ConditionalValidation `json:"conditionalRules,omitempty"`

	// Custom error messages
	Messages map[string]string `json:"messages,omitempty"`

	// Cross-field validation
	MatchField string `json:"matchField,omitempty"`

	// Async validation
	AsyncValidator *AsyncValidatorConfig `json:"asyncValidator,omitempty"`
}

// ConditionalValidation defines conditional validation rules
type ConditionalValidation struct {
	Condition *condition.ConditionGroup `json:"condition"`
	Rules     ValidationRules           `json:"rules"`
}

// AsyncValidatorConfig defines async validation
type AsyncValidatorConfig struct {
	URL          string `json:"url"`
	Method       string `json:"method"`
	Debounce     int    `json:"debounce,omitempty"` // milliseconds
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type FieldType string

// Additional field types for enterprise ERP
const (
	FieldTypeColorPicker    FieldType = "colorpicker"
	FieldTypeRichText       FieldType = "richtext"
	FieldTypeDateRange      FieldType = "daterange"
	FieldTypePhoneNumber    FieldType = "phone"
	FieldTypeURL            FieldType = "url"
	FieldTypeCurrency       FieldType = "currency"
	FieldTypeEmail          FieldType = "email"
	FieldTypeJSON           FieldType = "json"
	FieldTypeCode           FieldType = "code"
	FieldTypeSignature      FieldType = "signature"
	FieldTypeRating         FieldType = "rating"
	FieldTypeImageUpload    FieldType = "imageupload"
	FieldTypeFileMultiple   FieldType = "filemultiple"
	FieldTypeAutoComplete   FieldType = "autocomplete"
	FieldTypeChipInput      FieldType = "chipinput"
	FieldTypeLocationPicker FieldType = "location"
	FieldTypeTree           FieldType = "tree"
	FieldTypeTransfer       FieldType = "transfer"
	FieldTypeCascader       FieldType = "cascader"
)

// FormLayout with responsive breakpoints
type FormLayout struct {
	Columns     int          `json:"columns"`
	Gap         string       `json:"gap"`
	Direction   string       `json:"direction"`
	Responsive  bool         `json:"responsive"`
	Sections    []Section    `json:"sections,omitempty"`
	FieldGroups []FieldGroup `json:"fieldGroups,omitempty"`

	// Responsive breakpoints
	Breakpoints *Breakpoints `json:"breakpoints,omitempty"`

	// Grid system
	GridSystem GridSystem `json:"gridSystem,omitempty"`
}

// Breakpoints defines responsive breakpoints
type Breakpoints struct {
	Mobile  LayoutConfig `json:"mobile,omitempty"`  // < 768px
	Tablet  LayoutConfig `json:"tablet,omitempty"`  // >= 768px
	Desktop LayoutConfig `json:"desktop,omitempty"` // >= 1024px
	Wide    LayoutConfig `json:"wide,omitempty"`    // >= 1280px
}

// LayoutConfig defines layout at breakpoint
type LayoutConfig struct {
	Columns int    `json:"columns"`
	Gap     string `json:"gap"`
}

// GridSystem defines grid configuration
type GridSystem string

const (
	GridSystem12   GridSystem = "12"
	GridSystem24   GridSystem = "24"
	GridSystemFlex GridSystem = "flex"
)

// Helper functions for condition integration
func (f *FieldSchema) IsVisible(ctx *condition.EvalContext) (bool, error) {
	if f.VisibilityRules == nil {
		return true, nil
	}

	evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
	return evaluator.Evaluate(context.Background(), f.VisibilityRules, ctx)
}

func (f *FieldSchema) IsEnabled(ctx *condition.EvalContext) (bool, error) {
	if f.EnabledRules == nil {
		return !f.Disabled, nil
	}

	evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
	return evaluator.Evaluate(context.Background(), f.EnabledRules, ctx)
}

func (f *FieldSchema) IsRequired(ctx *condition.EvalContext) (bool, error) {
	if f.RequiredRules == nil {
		return f.Required, nil
	}

	evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
	return evaluator.Evaluate(context.Background(), f.RequiredRules, ctx)
}

// Example preset for enterprise multi-tenant form
var PresetForms = map[string]FormSchema{
	"tenant_user_create": {
		ID:          "create-tenant-user-form",
		Title:       "Create User",
		Description: "Add a new user to the organization",
		Action:      "/api/v1/tenants/{{tenantId}}/users",
		Method:      "POST",
		TenantScope: TenantScopeOrg,

		HTMXConfig: &HTMXConfig{
			Enabled:      true,
			PostURL:      "/api/v1/tenants/{{tenantId}}/users",
			SwapStrategy: HTMXSwapInnerHTML,
			Target:       "#user-list",
			PushURL:      true,
			Validate:     true,
		},

		AlpineConfig: &AlpineConfig{
			Enabled: true,
			XData:   "userForm",
			Cloak:   true,
		},

		Security: SecurityConfig{
			CSRF: CSRFConfig{
				Enabled:    true,
				TokenField: "_csrf",
				HeaderName: "X-CSRF-Token",
			},
			SanitizeInput: true,
		},

		Audit: AuditConfig{
			Enabled:        true,
			LogChanges:     true,
			LogSubmissions: true,
			ExcludeFields:  []string{"password"},
		},

		Fields: []FieldSchema{
			{
				Name:        "email",
				Type:        FieldTypeEmail,
				Label:       "Email Address",
				Placeholder: "user@company.com",
				Required:    true,
				Validation: ValidationRules{
					Email:     true,
					MaxLength: intPtr(255),
					AsyncValidator: &AsyncValidatorConfig{
						URL:          "/api/v1/validate/email",
						Method:       "POST",
						Debounce:     500,
						ErrorMessage: "Email already exists",
					},
				},
				HTMXConfig: &HTMXFieldConfig{
					GetURL:  "/api/v1/validate/email",
					Trigger: "keyup changed delay:500ms",
					Target:  "#email-validation",
					Swap:    HTMXSwapInnerHTML,
				},
				Layout: FieldLayout{ColSpan: 2},
			},
			{
				Name:     "role",
				Type:     FieldTypeSelect,
				Label:    "Role",
				Required: true,
				DataSource: &DataSourceConfig{
					Type:           DataSourceTypeAPI,
					URL:            "/api/v1/tenants/{{tenantId}}/roles",
					Method:         "GET",
					Cache:          true,
					CacheDuration:  300,
					TenantFiltered: true,
				},
				Layout: FieldLayout{ColSpan: 1},
			},
			{
				Name:   "sendInvite",
				Type:   FieldTypeSwitch,
				Label:  "Send invitation email",
				Value:  true,
				Layout: FieldLayout{ColSpan: 1},
			},
		},

		Submit: ButtonSchema{
			Text:    "Create User",
			Type:    ButtonTypeSubmit,
			Variant: ButtonVariantPrimary,
			Size:    ButtonSizeMedium,
		},

		Layout: FormLayout{
			Columns:    2,
			Gap:        "4",
			Direction:  "vertical",
			Responsive: true,
			GridSystem: GridSystem12,
		},

		Notifications: NotificationConfig{
			Success: NotificationStyle{
				Enabled: true,
				Message: "User created successfully",
				Icon:    "check-circle",
			},
			Error: NotificationStyle{
				Enabled: true,
				Message: "Failed to create user",
				Icon:    "alert-circle",
			},
			Position:    NotificationPositionTopRight,
			Duration:    3000,
			Dismissible: true,
		},

		Permissions: []string{"users.create"},
	},
}

// FormRenderer handles server-side rendering
type FormRenderer struct {
	schema      FormSchema
	evalContext *condition.EvalContext
	tenantID    string
	userID      string
	permissions []string
}

// NewFormRenderer creates a new form renderer
func NewFormRenderer(schema FormSchema, tenantID, userID string, permissions []string) *FormRenderer {
	return &FormRenderer{
		schema:      schema,
		tenantID:    tenantID,
		userID:      userID,
		permissions: permissions,
	}
}

// Render generates HTMX-compatible HTML with Alpine.js directives
func (r *FormRenderer) Render(data map[string]any) (string, error) {
	// Check form-level permissions
	if !r.hasPermission(r.schema.Permissions) {
		return "", ErrPermissionDenied
	}

	// Create evaluation context for conditional rendering
	evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
	r.evalContext = evalCtx

	// Filter fields based on visibility and permissions
	visibleFields, err := r.filterFields()
	if err != nil {
		return "", err
	}

	// Generate HTML
	return r.generateHTML(visibleFields, data)
}

// filterFields returns only visible and permitted fields
func (r *FormRenderer) filterFields() ([]FieldSchema, error) {
	var filtered []FieldSchema

	for _, field := range r.schema.Fields {
		// Check field-level permissions
		if !r.hasPermission(field.ViewPermissions) {
			continue
		}

		// Check visibility rules
		if field.VisibilityRules != nil {
			visible, err := field.IsVisible(r.evalContext)
			if err != nil {
				return nil, err
			}
			if !visible {
				continue
			}
		}

		filtered = append(filtered, field)
	}

	return filtered, nil
}

// hasPermission checks if user has required permissions
func (r *FormRenderer) hasPermission(required []string) bool {
	if len(required) == 0 {
		return true
	}

	permMap := make(map[string]bool)
	for _, p := range r.permissions {
		permMap[p] = true
	}

	for _, req := range required {
		if !permMap[req] {
			return false
		}
	}

	return true
}

// generateHTML creates the actual HTML markup
func (r *FormRenderer) generateHTML(fields []FieldSchema, data map[string]any) (string, error) {
	// Implementation would generate HTMX + Alpine.js markup
	// This is a placeholder - actual implementation would be more complex
	return "", nil
}

// FormValidator handles server-side validation
type FormValidator struct {
	schema      FormSchema
	evalContext *condition.EvalContext
}

// NewFormValidator creates a new validator
func NewFormValidator(schema FormSchema) *FormValidator {
	return &FormValidator{
		schema: schema,
	}
}

// Validate validates form data against schema
func (v *FormValidator) Validate(data map[string]any) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: make(map[string][]string),
	}

	// Create evaluation context for conditional validation
	v.evalContext = condition.NewEvalContext(data, condition.DefaultEvalOptions())

	for _, field := range v.schema.Fields {
		fieldErrors := v.validateField(field, data)
		if len(fieldErrors) > 0 {
			result.Valid = false
			result.Errors[field.Name] = fieldErrors
		}
	}

	return result
}

// validateField validates a single field
func (v *FormValidator) validateField(field FieldSchema, data map[string]any) []string {
	var errors []string

	value, exists := data[field.Name]

	// Check if field is required
	required := field.Required
	if field.RequiredRules != nil {
		isRequired, err := field.IsRequired(v.evalContext)
		if err == nil {
			required = isRequired
		}
	}

	// Required validation
	if required && (!exists || value == nil || value == "") {
		msg := field.Validation.Messages["required"]
		if msg == "" {
			msg = fmt.Sprintf("%s is required", field.Label)
		}
		errors = append(errors, msg)
		return errors
	}

	// Skip other validations if field is empty and not required
	if !exists || value == nil || value == "" {
		return errors
	}

	// Type-specific validation
	errors = append(errors, v.validateByType(field, value)...)

	// Pattern validation
	if field.Validation.Pattern != "" {
		errors = append(errors, v.validatePattern(field, value)...)
	}

	// Range validation
	errors = append(errors, v.validateRange(field, value)...)

	// Length validation
	errors = append(errors, v.validateLength(field, value)...)

	// Email validation
	if field.Validation.Email {
		errors = append(errors, v.validateEmail(field, value)...)
	}

	// URL validation
	if field.Validation.URL {
		errors = append(errors, v.validateURL(field, value)...)
	}

	// Conditional validation
	errors = append(errors, v.validateConditional(field, value)...)

	// Cross-field validation
	if field.Validation.MatchField != "" {
		errors = append(errors, v.validateMatch(field, value, data)...)
	}

	return errors
}

// ValidationResult represents validation outcome
type ValidationResult struct {
	Valid  bool                `json:"valid"`
	Errors map[string][]string `json:"errors"`
}

// FormStateManager manages form state in-memory or with cache
type FormStateManager struct {
	states map[string]*FormState // sessionID -> state
	mu     sync.RWMutex
}

// NewFormStateManager creates a new state manager
func NewFormStateManager() *FormStateManager {
	return &FormStateManager{
		states: make(map[string]*FormState),
	}
}

// GetState retrieves form state
func (m *FormStateManager) GetState(sessionID string) *FormState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.states[sessionID]
}

// UpdateState updates form state
func (m *FormStateManager) UpdateState(sessionID string, state *FormState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state.LastUpdated = time.Now()
	m.states[sessionID] = state
}

// ClearState removes form state
func (m *FormStateManager) ClearState(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, sessionID)
}

// FormSchemaRegistry manages form schema definitions
type FormSchemaRegistry struct {
	schemas map[string]FormSchema
	mu      sync.RWMutex
}

// NewFormSchemaRegistry creates a registry
func NewFormSchemaRegistry() *FormSchemaRegistry {
	return &FormSchemaRegistry{
		schemas: make(map[string]FormSchema),
	}
}

// Register registers a form schema
func (r *FormSchemaRegistry) Register(schema FormSchema) error {
	if err := schema.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.schemas[schema.ID] = schema
	return nil
}

// Get retrieves a schema by ID
func (r *FormSchemaRegistry) Get(id string) (FormSchema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schema, exists := r.schemas[id]
	if !exists {
		return FormSchema{}, ErrSchemaNotFound
	}

	return schema, nil
}

// LoadFromJSON loads schema from JSON file
func (r *FormSchemaRegistry) LoadFromJSON(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var schema FormSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return err
	}

	return r.Register(schema)
}

// LoadFromDirectory loads all schemas from directory
func (r *FormSchemaRegistry) LoadFromDirectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filepath := filepath.Join(dir, file.Name())
			if err := r.LoadFromJSON(filepath); err != nil {
				return fmt.Errorf("failed to load %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

// List returns all registered schemas
func (r *FormSchemaRegistry) List() []FormSchema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schemas := make([]FormSchema, 0, len(r.schemas))
	for _, schema := range r.schemas {
		schemas = append(schemas, schema)
	}

	return schemas
}

// FormSchema validation
func (f *FormSchema) Validate() error {
	if f.ID == "" {
		return errors.New("form ID is required")
	}

	if f.Title == "" {
		return errors.New("form title is required")
	}

	if f.Action == "" {
		return errors.New("form action is required")
	}

	if f.Method == "" {
		return errors.New("form method is required")
	}

	if len(f.Fields) == 0 {
		return errors.New("form must have at least one field")
	}

	// Validate each field
	fieldNames := make(map[string]bool)
	for _, field := range f.Fields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}

		// Check for duplicate field names
		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name: %s", field.Name)
		}
		fieldNames[field.Name] = true
	}

	return nil
}

// FieldSchema validation
func (f *FieldSchema) Validate() error {
	if f.Name == "" {
		return errors.New("field name is required")
	}

	if f.Type == "" {
		return errors.New("field type is required")
	}

	if f.Label == "" {
		return errors.New("field label is required")
	}

	// Validate field type is known
	validTypes := map[FieldType]bool{
		FieldTypeText: true, FieldTypeEmail: true, FieldTypePassword: true,
		FieldTypeNumber: true, FieldTypeDate: true, FieldTypeTime: true,
		FieldTypeDateTime: true, FieldTypeTextarea: true, FieldTypeSelect: true,
		FieldTypeMultiSelect: true, FieldTypeCheckbox: true, FieldTypeRadio: true,
		FieldTypeFile: true, FieldTypeHidden: true, FieldTypeSwitch: true,
		FieldTypeSlider: true, FieldTypeTagsInput: true, FieldTypeCurrency: true,
		FieldTypeColorPicker: true, FieldTypeRichText: true, FieldTypeDateRange: true,
		FieldTypePhoneNumber: true, FieldTypeURL: true, FieldTypeJSON: true,
		FieldTypeCode: true, FieldTypeSignature: true, FieldTypeRating: true,
		FieldTypeImageUpload: true, FieldTypeFileMultiple: true, FieldTypeAutoComplete: true,
		FieldTypeChipInput: true, FieldTypeLocationPicker: true, FieldTypeTree: true,
		FieldTypeTransfer: true, FieldTypeCascader: true,
	}

	if !validTypes[f.Type] {
		return fmt.Errorf("invalid field type: %s", f.Type)
	}

	// Validate options for select/radio/checkbox fields
	if f.Type == FieldTypeSelect || f.Type == FieldTypeRadio || f.Type == FieldTypeMultiSelect {
		if len(f.Options) == 0 && f.DataSource == nil {
			return errors.New("select/radio/multiselect fields require options or data source")
		}
	}

	return nil
}

// Errors
var (
	ErrSchemaNotFound   = errors.New("schema not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidData      = errors.New("invalid data")
	ErrValidationFailed = errors.New("validation failed")
)

// Additional helper methods
func (v *FormValidator) validateByType(field FieldSchema, value any) []string {
	var errors []string

	switch field.Type {
	case FieldTypeEmail:
		if !v.isValidEmail(value) {
			errors = append(errors, "Invalid email format")
		}
	case FieldTypeURL:
		if !v.isValidURL(value) {
			errors = append(errors, "Invalid URL format")
		}
	case FieldTypeNumber, FieldTypeCurrency:
		if !v.isNumeric(value) {
			errors = append(errors, "Must be a valid number")
		}
	case FieldTypeDate, FieldTypeDateTime, FieldTypeDateRange:
		if !v.isValidDate(value, field.Type) {
			errors = append(errors, "Invalid date format")
		}
	case FieldTypePhoneNumber:
		if !v.isValidPhone(value) {
			errors = append(errors, "Invalid phone number")
		}
	case FieldTypeJSON:
		if !v.isValidJSON(value) {
			errors = append(errors, "Invalid JSON format")
		}
	}

	return errors
}

func (v *FormValidator) validatePattern(field FieldSchema, value any) []string {
	var errors []string

	strVal, ok := value.(string)
	if !ok {
		return errors
	}

	matched, err := regexp.MatchString(field.Validation.Pattern, strVal)
	if err != nil || !matched {
		msg := field.Validation.Messages["pattern"]
		if msg == "" {
			msg = fmt.Sprintf("%s format is invalid", field.Label)
		}
		errors = append(errors, msg)
	}

	return errors
}

func (v *FormValidator) validateRange(field FieldSchema, value any) []string {
	var errors []string

	numVal, err := v.toFloat64(value)
	if err != nil {
		return errors
	}

	if field.Validation.Min != nil && numVal < *field.Validation.Min {
		msg := field.Validation.Messages["min"]
		if msg == "" {
			msg = fmt.Sprintf("%s must be at least %.2f", field.Label, *field.Validation.Min)
		}
		errors = append(errors, msg)
	}

	if field.Validation.Max != nil && numVal > *field.Validation.Max {
		msg := field.Validation.Messages["max"]
		if msg == "" {
			msg = fmt.Sprintf("%s must be at most %.2f", field.Label, *field.Validation.Max)
		}
		errors = append(errors, msg)
	}

	return errors
}

func (v *FormValidator) validateLength(field FieldSchema, value any) []string {
	var errors []string

	strVal := fmt.Sprintf("%v", value)
	length := len(strVal)

	if field.Validation.MinLength != nil && length < *field.Validation.MinLength {
		msg := field.Validation.Messages["minLength"]
		if msg == "" {
			msg = fmt.Sprintf("%s must be at least %d characters", field.Label, *field.Validation.MinLength)
		}
		errors = append(errors, msg)
	}

	if field.Validation.MaxLength != nil && length > *field.Validation.MaxLength {
		msg := field.Validation.Messages["maxLength"]
		if msg == "" {
			msg = fmt.Sprintf("%s must be at most %d characters", field.Label, *field.Validation.MaxLength)
		}
		errors = append(errors, msg)
	}

	return errors
}

func (v *FormValidator) validateEmail(field FieldSchema, value any) []string {
	var errors []string

	if !v.isValidEmail(value) {
		msg := field.Validation.Messages["email"]
		if msg == "" {
			msg = fmt.Sprintf("%s must be a valid email address", field.Label)
		}
		errors = append(errors, msg)
	}

	return errors
}

func (v *FormValidator) validateURL(field FieldSchema, value any) []string {
	var errors []string

	if !v.isValidURL(value) {
		msg := field.Validation.Messages["url"]
		if msg == "" {
			msg = fmt.Sprintf("%s must be a valid URL", field.Label)
		}
		errors = append(errors, msg)
	}

	return errors
}

func (v *FormValidator) validateConditional(field FieldSchema, value any) []string {
	var errors []string

	for _, condRule := range field.Validation.ConditionalRules {
		// Evaluate condition
		evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
		matches, err := evaluator.Evaluate(context.Background(), condRule.Condition, v.evalContext)

		if err != nil || !matches {
			continue
		}

		// Apply conditional validation rules
		tempField := field
		tempField.Validation = condRule.Rules
		errors = append(errors, v.validateField(tempField, map[string]any{
			field.Name: value,
		})...)
	}

	return errors
}

func (v *FormValidator) validateMatch(field FieldSchema, value any, data map[string]any) []string {
	var errors []string

	matchValue, exists := data[field.Validation.MatchField]
	if !exists {
		return errors
	}

	if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", matchValue) {
		msg := field.Validation.Messages["match"]
		if msg == "" {
			msg = fmt.Sprintf("%s must match %s", field.Label, field.Validation.MatchField)
		}
		errors = append(errors, msg)
	}

	return errors
}

// Helper validation methods
func (v *FormValidator) isValidEmail(value any) bool {
	strVal, ok := value.(string)
	if !ok {
		return false
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strVal)
}

func (v *FormValidator) isValidURL(value any) bool {
	strVal, ok := value.(string)
	if !ok {
		return false
	}

	_, err := url.Parse(strVal)
	return err == nil && (strings.HasPrefix(strVal, "http://") || strings.HasPrefix(strVal, "https://"))
}

func (v *FormValidator) isNumeric(value any) bool {
	_, err := v.toFloat64(value)
	return err == nil
}

func (v *FormValidator) isValidDate(value any, fieldType FieldType) bool {
	strVal, ok := value.(string)
	if !ok {
		return false
	}

	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, strVal); err == nil {
			return true
		}
	}

	return false
}

func (v *FormValidator) isValidPhone(value any) bool {
	strVal, ok := value.(string)
	if !ok {
		return false
	}

	// Basic phone validation - adjust regex for your needs
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(strings.ReplaceAll(strVal, " ", ""))
}

func (v *FormValidator) isValidJSON(value any) bool {
	strVal, ok := value.(string)
	if !ok {
		return false
	}

	var js json.RawMessage
	return json.Unmarshal([]byte(strVal), &js) == nil
}

func (v *FormValidator) toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, errors.New("not a number")
	}
}

// DataSourceResolver resolves dynamic data sources
type DataSourceResolver struct {
	httpClient *http.Client
	cache      *DataSourceCache
	tenantID   string
}

// NewDataSourceResolver creates a resolver
func NewDataSourceResolver(tenantID string) *DataSourceResolver {
	return &DataSourceResolver{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cache:      NewDataSourceCache(),
		tenantID:   tenantID,
	}
}

// Resolve resolves a data source to options
func (r *DataSourceResolver) Resolve(ds *DataSourceConfig, params map[string]string) ([]OptionSchema, error) {
	if ds == nil {
		return nil, nil
	}

	switch ds.Type {
	case DataSourceTypeStatic:
		return ds.Data, nil

	case DataSourceTypeAPI:
		return r.resolveAPI(ds, params)

	case DataSourceTypeFunction:
		return r.resolveFunction(ds, params)

	default:
		return nil, fmt.Errorf("unknown data source type: %s", ds.Type)
	}
}

// resolveAPI fetches data from API
func (r *DataSourceResolver) resolveAPI(ds *DataSourceConfig, params map[string]string) ([]OptionSchema, error) {
	// Check cache
	if ds.Cache {
		cacheKey := r.buildCacheKey(ds, params)
		if cached := r.cache.Get(cacheKey); cached != nil {
			return cached, nil
		}
	}

	// Build URL with parameters
	url := r.buildURL(ds.URL, params)

	// Add tenant filter if enabled
	if ds.TenantFiltered {
		url = strings.ReplaceAll(url, "{{tenantId}}", r.tenantID)
	}

	// Create request
	req, err := http.NewRequest(ds.Method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range ds.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result []OptionSchema
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Cache result
	if ds.Cache {
		cacheKey := r.buildCacheKey(ds, params)
		r.cache.Set(cacheKey, result, time.Duration(ds.CacheDuration)*time.Second)
	}

	return result, nil
}

func (r *DataSourceResolver) resolveFunction(ds *DataSourceConfig, params map[string]string) ([]OptionSchema, error) {
	// Placeholder for custom function resolution
	return nil, errors.New("function data sources not yet implemented")
}

func (r *DataSourceResolver) buildURL(baseURL string, params map[string]string) string {
	url := baseURL
	for key, value := range params {
		placeholder := "{{" + key + "}}"
		url = strings.ReplaceAll(url, placeholder, value)
	}
	return url
}

func (r *DataSourceResolver) buildCacheKey(ds *DataSourceConfig, params map[string]string) string {
	parts := []string{ds.URL}
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "|")
}

// DataSourceCache provides simple in-memory caching
type DataSourceCache struct {
	data map[string]cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	options   []OptionSchema
	expiresAt time.Time
}

func NewDataSourceCache() *DataSourceCache {
	cache := &DataSourceCache{
		data: make(map[string]cacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

func (c *DataSourceCache) Get(key string) []OptionSchema {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.options
}

func (c *DataSourceCache) Set(key string, options []OptionSchema, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		options:   options,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *DataSourceCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *DataSourceCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.expiresAt) {
			delete(c.data, key)
		}
	}
}
