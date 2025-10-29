package types

//
// import (
// 	"time"
//
// 	"github.com/niiniyare/erp/condition"
// )
//
// // FormSchema defines the structure for dynamically generating forms
// // Supports multi-tenant ERP with HTMX/AlpineJS rendering
// type FormSchema struct {
// 	ID          string           `json:"id"`
// 	Title       string           `json:"title"`
// 	Description string           `json:"description,omitempty"`
// 	Action      string           `json:"action"`
// 	Method      string           `json:"method"`
// 	Fields      []FieldSchema    `json:"fields"`
// 	Submit      ButtonSchema     `json:"submit"`
// 	Cancel      *ButtonSchema    `json:"cancel,omitempty"`
// 	Layout      FormLayout       `json:"layout"`
// 	Permissions *PermissionRules `json:"permissions,omitempty"`
// 	Validation  *FormValidation  `json:"validation,omitempty"`
// 	Workflow    *WorkflowConfig  `json:"workflow,omitempty"`
// 	Tenant      *TenantConfig    `json:"tenant,omitempty"`
// 	Events      *FormEvents      `json:"events,omitempty"`
// 	I18n        *I18nConfig      `json:"i18n,omitempty"`
// 	Meta        *FormMeta        `json:"meta,omitempty"`
// }
//
// // FieldSchema defines individual form field specifications
// type FieldSchema struct {
// 	Name         string              `json:"name"`
// 	Type         FieldType           `json:"type"`
// 	Label        string              `json:"label"`
// 	Placeholder  string              `json:"placeholder,omitempty"`
// 	Required     bool                `json:"required"`
// 	Disabled     bool                `json:"disabled"`
// 	Readonly     bool                `json:"readonly"`
// 	Value        interface{}         `json:"value,omitempty"`
// 	DefaultValue interface{}         `json:"defaultValue,omitempty"`
// 	Validation   ValidationRules     `json:"validation,omitempty"`
// 	Options      []OptionSchema      `json:"options,omitempty"`
// 	Attributes   map[string]string   `json:"attributes,omitempty"`
// 	HelpText     string              `json:"helpText,omitempty"`
// 	ErrorText    string              `json:"errorText,omitempty"`
// 	Layout       FieldLayout         `json:"layout"`
// 	Conditional  *ConditionalRules   `json:"conditional,omitempty"`
// 	DataSource   *DataSource         `json:"dataSource,omitempty"`
// 	Transform    *FieldTransform     `json:"transform,omitempty"`
// 	Dependencies []string            `json:"dependencies,omitempty"`
// 	Permissions  *FieldPermissions   `json:"permissions,omitempty"`
// 	Events       *FieldEvents        `json:"events,omitempty"`
// 	HTMX         *HTMXConfig         `json:"htmx,omitempty"`
// 	Alpine       *AlpineConfig       `json:"alpine,omitempty"`
// 	Mask         *InputMask          `json:"mask,omitempty"`
// 	AutoComplete *AutoCompleteConfig `json:"autoComplete,omitempty"`
// }
//
// // FieldType represents different input field types
// type FieldType string
//
// const (
// 	FieldTypeText         FieldType = "text"
// 	FieldTypeEmail        FieldType = "email"
// 	FieldTypePassword     FieldType = "password"
// 	FieldTypeNumber       FieldType = "number"
// 	FieldTypeDate         FieldType = "date"
// 	FieldTypeTime         FieldType = "time"
// 	FieldTypeDateTime     FieldType = "datetime"
// 	FieldTypeTextarea     FieldType = "textarea"
// 	FieldTypeSelect       FieldType = "select"
// 	FieldTypeMultiSelect  FieldType = "multiselect"
// 	FieldTypeCheckbox     FieldType = "checkbox"
// 	FieldTypeRadio        FieldType = "radio"
// 	FieldTypeFile         FieldType = "file"
// 	FieldTypeHidden       FieldType = "hidden"
// 	FieldTypeSwitch       FieldType = "switch"
// 	FieldTypeSlider       FieldType = "slider"
// 	FieldTypeTagsInput    FieldType = "tagsinput"
// 	FieldTypeCurrency     FieldType = "currency"
// 	FieldTypeColorPicker  FieldType = "colorpicker"
// 	FieldTypeRichText     FieldType = "richtext"
// 	FieldTypeDateRange    FieldType = "daterange"
// 	FieldTypePhoneNumber  FieldType = "phonenumber"
// 	FieldTypeRating       FieldType = "rating"
// 	FieldTypeTreeSelect   FieldType = "treeselect"
// 	FieldTypeCascader     FieldType = "cascader"
// 	FieldTypeTransfer     FieldType = "transfer"
// 	FieldTypeCodeEditor   FieldType = "codeeditor"
// 	FieldTypeJsonEditor   FieldType = "jsoneditor"
// 	FieldTypeImageUpload  FieldType = "imageupload"
// 	FieldTypeSignature    FieldType = "signature"
// 	FieldTypeLocation     FieldType = "location"
// 	FieldTypeRelationship FieldType = "relationship"
// )
//
// // ConditionalRules defines when fields are visible/enabled/required
// // Integrates with condition package for runtime evaluation
// type ConditionalRules struct {
// 	Show    *condition.ConditionGroup `json:"show,omitempty"`
// 	Hide    *condition.ConditionGroup `json:"hide,omitempty"`
// 	Enable  *condition.ConditionGroup `json:"enable,omitempty"`
// 	Disable *condition.ConditionGroup `json:"disable,omitempty"`
// 	Require *condition.ConditionGroup `json:"require,omitempty"`
// 	// Formula-based conditions (uses expr-lang)
// 	ShowIf     string `json:"showIf,omitempty"`
// 	HideIf     string `json:"hideIf,omitempty"`
// 	EnableIf   string `json:"enableIf,omitempty"`
// 	DisableIf  string `json:"disableIf,omitempty"`
// 	RequireIf  string `json:"requireIf,omitempty"`
// 	ValidateIf string `json:"validateIf,omitempty"`
// }
//
// // ValidationRules defines validation constraints for fields
// type ValidationRules struct {
// 	MinLength    *int                      `json:"minLength,omitempty"`
// 	MaxLength    *int                      `json:"maxLength,omitempty"`
// 	Min          *float64                  `json:"min,omitempty"`
// 	Max          *float64                  `json:"max,omitempty"`
// 	Pattern      string                    `json:"pattern,omitempty"`
// 	CustomRules  []string                  `json:"customRules,omitempty"`
// 	Dependencies []string                  `json:"dependencies,omitempty"`
// 	Email        bool                      `json:"email,omitempty"`
// 	URL          bool                      `json:"url,omitempty"`
// 	MinFiles     *int                      `json:"minFiles,omitempty"`
// 	MaxFiles     *int                      `json:"maxFiles,omitempty"`
// 	AllowedTypes []string                  `json:"allowedTypes,omitempty"`
// 	MaxFileSize  *int64                    `json:"maxFileSize,omitempty"`
// 	UniqueField  bool                      `json:"uniqueField,omitempty"`
// 	Async        *AsyncValidation          `json:"async,omitempty"`
// 	Conditions   *condition.ConditionGroup `json:"conditions,omitempty"`
// 	Formula      string                    `json:"formula,omitempty"`
// }
//
// // AsyncValidation for server-side validation
// type AsyncValidation struct {
// 	URL         string            `json:"url"`
// 	Method      string            `json:"method"`
// 	Debounce    int               `json:"debounce"` // milliseconds
// 	Headers     map[string]string `json:"headers,omitempty"`
// 	OnSuccess   string            `json:"onSuccess,omitempty"`
// 	OnError     string            `json:"onError,omitempty"`
// 	ValidateOn  []string          `json:"validateOn,omitempty"` // blur, change, submit
// 	CacheResult bool              `json:"cacheResult"`
// }
//
// // DataSource defines dynamic data loading
// type DataSource struct {
// 	Type        DataSourceType    `json:"type"`
// 	URL         string            `json:"url,omitempty"`
// 	Method      string            `json:"method,omitempty"`
// 	Headers     map[string]string `json:"headers,omitempty"`
// 	Params      map[string]string `json:"params,omitempty"`
// 	Transform   string            `json:"transform,omitempty"` // JavaScript function
// 	Cache       bool              `json:"cache"`
// 	CacheTTL    int               `json:"cacheTTL,omitempty"` // seconds
// 	DependsOn   []string          `json:"dependsOn,omitempty"`
// 	LazyLoad    bool              `json:"lazyLoad"`
// 	SearchField string            `json:"searchField,omitempty"`
// 	ValueField  string            `json:"valueField,omitempty"`
// 	LabelField  string            `json:"labelField,omitempty"`
// 	Tenant      bool              `json:"tenant"` // Filter by tenant
// }
//
// // DataSourceType defines where data comes from
// type DataSourceType string
//
// const (
// 	DataSourceStatic   DataSourceType = "static"
// 	DataSourceAPI      DataSourceType = "api"
// 	DataSourceDatabase DataSourceType = "database"
// 	DataSourceFunction DataSourceType = "function"
// )
//
// // FieldTransform defines data transformation
// type FieldTransform struct {
// 	Input  string `json:"input,omitempty"`  // Transform before display
// 	Output string `json:"output,omitempty"` // Transform before submit
// 	Format string `json:"format,omitempty"` // Display format
// }
//
// // FieldPermissions defines field-level access control
// type FieldPermissions struct {
// 	View   *condition.ConditionGroup `json:"view,omitempty"`
// 	Edit   *condition.ConditionGroup `json:"edit,omitempty"`
// 	Roles  []string                  `json:"roles,omitempty"`
// 	Groups []string                  `json:"groups,omitempty"`
// }
//
// // FieldEvents defines field-level event handlers
// type FieldEvents struct {
// 	OnChange   string `json:"onChange,omitempty"`
// 	OnFocus    string `json:"onFocus,omitempty"`
// 	OnBlur     string `json:"onBlur,omitempty"`
// 	OnInput    string `json:"onInput,omitempty"`
// 	OnKeyPress string `json:"onKeyPress,omitempty"`
// 	OnKeyUp    string `json:"onKeyUp,omitempty"`
// 	OnKeyDown  string `json:"onKeyDown,omitempty"`
// 	OnClick    string `json:"onClick,omitempty"`
// 	OnMount    string `json:"onMount,omitempty"`
// 	OnUnmount  string `json:"onUnmount,omitempty"`
// 	Debounce   int    `json:"debounce,omitempty"` // milliseconds
// 	Throttle   int    `json:"throttle,omitempty"` // milliseconds
// }
//
// // HTMXConfig for HTMX-specific attributes
// type HTMXConfig struct {
// 	Get       string            `json:"get,omitempty"`
// 	Post      string            `json:"post,omitempty"`
// 	Put       string            `json:"put,omitempty"`
// 	Delete    string            `json:"delete,omitempty"`
// 	Patch     string            `json:"patch,omitempty"`
// 	Trigger   string            `json:"trigger,omitempty"`
// 	Target    string            `json:"target,omitempty"`
// 	Swap      string            `json:"swap,omitempty"`
// 	SwapOOB   string            `json:"swapOob,omitempty"`
// 	Select    string            `json:"select,omitempty"`
// 	Indicator string            `json:"indicator,omitempty"`
// 	PushURL   string            `json:"pushUrl,omitempty"`
// 	Vals      string            `json:"vals,omitempty"`
// 	Confirm   string            `json:"confirm,omitempty"`
// 	Headers   map[string]string `json:"headers,omitempty"`
// 	Include   string            `json:"include,omitempty"`
// 	Boost     bool              `json:"boost,omitempty"`
// 	Disabled  string            `json:"disabled,omitempty"`
// 	Sync      string            `json:"sync,omitempty"`
// 	Validate  bool              `json:"validate,omitempty"`
// 	Encoding  string            `json:"encoding,omitempty"`
// }
//
// // AlpineConfig for Alpine.js directives
// type AlpineConfig struct {
// 	XData       string `json:"xData,omitempty"`
// 	XInit       string `json:"xInit,omitempty"`
// 	XShow       string `json:"xShow,omitempty"`
// 	XIf         string `json:"xIf,omitempty"`
// 	XFor        string `json:"xFor,omitempty"`
// 	XModel      string `json:"xModel,omitempty"`
// 	XModelable  string `json:"xModelable,omitempty"`
// 	XBind       string `json:"xBind,omitempty"`
// 	XOn         string `json:"xOn,omitempty"`
// 	XText       string `json:"xText,omitempty"`
// 	XHTML       string `json:"xHtml,omitempty"`
// 	XRef        string `json:"xRef,omitempty"`
// 	XCloak      bool   `json:"xCloak,omitempty"`
// 	XTransition string `json:"xTransition,omitempty"`
// 	XTeleport   string `json:"xTeleport,omitempty"`
// }
//
// // InputMask for formatted input
// type InputMask struct {
// 	Pattern     string                 `json:"pattern"`
// 	Placeholder string                 `json:"placeholder,omitempty"`
// 	Type        string                 `json:"type,omitempty"` // phone, currency, date, custom
// 	Options     map[string]interface{} `json:"options,omitempty"`
// }
//
// // AutoCompleteConfig for autocomplete functionality
// type AutoCompleteConfig struct {
// 	Enabled      bool              `json:"enabled"`
// 	Source       string            `json:"source"`
// 	MinChars     int               `json:"minChars"`
// 	MaxResults   int               `json:"maxResults"`
// 	Debounce     int               `json:"debounce"`
// 	Template     string            `json:"template,omitempty"`
// 	OnSelect     string            `json:"onSelect,omitempty"`
// 	Headers      map[string]string `json:"headers,omitempty"`
// 	CacheResults bool              `json:"cacheResults"`
// }
//
// // OptionSchema defines options for select, radio, and checkbox fields
// type OptionSchema struct {
// 	Value       string                 `json:"value"`
// 	Label       string                 `json:"label"`
// 	Disabled    bool                   `json:"disabled"`
// 	Selected    bool                   `json:"selected"`
// 	Group       string                 `json:"group,omitempty"`
// 	Icon        string                 `json:"icon,omitempty"`
// 	Color       string                 `json:"color,omitempty"`
// 	Description string                 `json:"description,omitempty"`
// 	Meta        map[string]interface{} `json:"meta,omitempty"`
// 	Children    []OptionSchema         `json:"children,omitempty"` // For tree select
// }
//
// // ButtonSchema defines button specifications
// type ButtonSchema struct {
// 	Text       string            `json:"text"`
// 	Type       ButtonType        `json:"type"`
// 	Variant    ButtonVariant     `json:"variant"`
// 	Size       ButtonSize        `json:"size"`
// 	Icon       string            `json:"icon,omitempty"`
// 	IconPos    string            `json:"iconPos,omitempty"` // left, right
// 	Loading    bool              `json:"loading"`
// 	Disabled   bool              `json:"disabled"`
// 	Attributes map[string]string `json:"attributes,omitempty"`
// 	HTMX       *HTMXConfig       `json:"htmx,omitempty"`
// 	Alpine     *AlpineConfig     `json:"alpine,omitempty"`
// 	Confirm    *ConfirmDialog    `json:"confirm,omitempty"`
// }
//
// // ButtonType represents different button types
// type ButtonType string
//
// const (
// 	ButtonTypeSubmit ButtonType = "submit"
// 	ButtonTypeButton ButtonType = "button"
// 	ButtonTypeReset  ButtonType = "reset"
// )
//
// // ButtonVariant represents different button styles
// type ButtonVariant string
//
// const (
// 	ButtonVariantPrimary     ButtonVariant = "primary"
// 	ButtonVariantSecondary   ButtonVariant = "secondary"
// 	ButtonVariantOutline     ButtonVariant = "outline"
// 	ButtonVariantGhost       ButtonVariant = "ghost"
// 	ButtonVariantLink        ButtonVariant = "link"
// 	ButtonVariantDestructive ButtonVariant = "destructive"
// 	ButtonVariantSuccess     ButtonVariant = "success"
// 	ButtonVariantWarning     ButtonVariant = "warning"
// 	ButtonVariantInfo        ButtonVariant = "info"
// )
//
// // ButtonSize represents different button sizes
// type ButtonSize string
//
// const (
// 	ButtonSizeSmall  ButtonSize = "sm"
// 	ButtonSizeMedium ButtonSize = "md"
// 	ButtonSizeLarge  ButtonSize = "lg"
// 	ButtonSizeIcon   ButtonSize = "icon"
// )
//
// // ConfirmDialog for confirmation prompts
// type ConfirmDialog struct {
// 	Title       string `json:"title"`
// 	Message     string `json:"message"`
// 	ConfirmText string `json:"confirmText,omitempty"`
// 	CancelText  string `json:"cancelText,omitempty"`
// 	Type        string `json:"type,omitempty"` // info, warning, danger
// }
//
// // FormLayout defines the overall form layout
// type FormLayout struct {
// 	Columns     int          `json:"columns"`
// 	Gap         string       `json:"gap"`
// 	Direction   string       `json:"direction"`
// 	Responsive  bool         `json:"responsive"`
// 	Sections    []Section    `json:"sections,omitempty"`
// 	FieldGroups []FieldGroup `json:"fieldGroups,omitempty"`
// 	Tabs        []Tab        `json:"tabs,omitempty"`
// 	Steps       []Step       `json:"steps,omitempty"`
// }
//
// // FieldLayout defines individual field layout
// type FieldLayout struct {
// 	ColSpan   int    `json:"colSpan"`
// 	Row       int    `json:"row,omitempty"`
// 	Order     int    `json:"order,omitempty"`
// 	Width     string `json:"width,omitempty"`
// 	MinWidth  string `json:"minWidth,omitempty"`
// 	MaxWidth  string `json:"maxWidth,omitempty"`
// 	ClassName string `json:"className,omitempty"`
// }
//
// // Section represents a logical grouping of fields
// type Section struct {
// 	ID          string                    `json:"id"`
// 	Title       string                    `json:"title"`
// 	Description string                    `json:"description,omitempty"`
// 	Fields      []string                  `json:"fields"`
// 	Collapsible bool                      `json:"collapsible"`
// 	Collapsed   bool                      `json:"collapsed"`
// 	Icon        string                    `json:"icon,omitempty"`
// 	Conditional *condition.ConditionGroup `json:"conditional,omitempty"`
// 	Permissions *FieldPermissions         `json:"permissions,omitempty"`
// }
//
// // FieldGroup represents a visual grouping of fields
// type FieldGroup struct {
// 	ID      string   `json:"id"`
// 	Title   string   `json:"title,omitempty"`
// 	Fields  []string `json:"fields"`
// 	Layout  string   `json:"layout"` // horizontal, vertical, grid
// 	Border  bool     `json:"border"`
// 	Padding string   `json:"padding"`
// }
//
// // Tab for tabbed form layouts
// type Tab struct {
// 	ID          string                    `json:"id"`
// 	Title       string                    `json:"title"`
// 	Icon        string                    `json:"icon,omitempty"`
// 	Fields      []string                  `json:"fields"`
// 	Sections    []string                  `json:"sections,omitempty"`
// 	Badge       string                    `json:"badge,omitempty"`
// 	Disabled    bool                      `json:"disabled"`
// 	Conditional *condition.ConditionGroup `json:"conditional,omitempty"`
// }
//
// // Step for multi-step forms
// type Step struct {
// 	ID          string                    `json:"id"`
// 	Title       string                    `json:"title"`
// 	Description string                    `json:"description,omitempty"`
// 	Fields      []string                  `json:"fields"`
// 	Sections    []string                  `json:"sections,omitempty"`
// 	Validation  *StepValidation           `json:"validation,omitempty"`
// 	Conditional *condition.ConditionGroup `json:"conditional,omitempty"`
// 	Optional    bool                      `json:"optional"`
// }
//
// // StepValidation defines validation for wizard steps
// type StepValidation struct {
// 	Required   bool                      `json:"required"`
// 	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
// 	OnNext     string                    `json:"onNext,omitempty"`
// 	OnPrevious string                    `json:"onPrevious,omitempty"`
// }
//
// // PermissionRules defines form-level access control
// type PermissionRules struct {
// 	View   *condition.ConditionGroup `json:"view,omitempty"`
// 	Create *condition.ConditionGroup `json:"create,omitempty"`
// 	Edit   *condition.ConditionGroup `json:"edit,omitempty"`
// 	Delete *condition.ConditionGroup `json:"delete,omitempty"`
// 	Submit *condition.ConditionGroup `json:"submit,omitempty"`
// 	Roles  []string                  `json:"roles,omitempty"`
// 	Groups []string                  `json:"groups,omitempty"`
// }
//
// // FormValidation defines form-level validation
// type FormValidation struct {
// 	ValidateOn      []string                  `json:"validateOn,omitempty"` // change, blur, submit
// 	ShowErrors      string                    `json:"showErrors,omitempty"` // all, touched, submitted
// 	FocusOnError    bool                      `json:"focusOnError"`
// 	ScrollToError   bool                      `json:"scrollToError"`
// 	CrossFieldRules []CrossFieldRule          `json:"crossFieldRules,omitempty"`
// 	Conditions      *condition.ConditionGroup `json:"conditions,omitempty"`
// 	Formula         string                    `json:"formula,omitempty"`
// 	OnValidate      string                    `json:"onValidate,omitempty"`
// }
//
// // CrossFieldRule for validating relationships between fields
// type CrossFieldRule struct {
// 	ID         string                    `json:"id"`
// 	Message    string                    `json:"message"`
// 	Fields     []string                  `json:"fields"`
// 	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
// 	Formula    string                    `json:"formula,omitempty"`
// }
//
// // WorkflowConfig for workflow integration
// type WorkflowConfig struct {
// 	Enabled       bool                 `json:"enabled"`
// 	WorkflowID    string               `json:"workflowId"`
// 	Stage         string               `json:"stage,omitempty"`
// 	Status        string               `json:"status,omitempty"`
// 	Actions       []WorkflowAction     `json:"actions,omitempty"`
// 	Transitions   []WorkflowTransition `json:"transitions,omitempty"`
// 	Notifications []NotificationRule   `json:"notifications,omitempty"`
// 	Approvals     *ApprovalConfig      `json:"approvals,omitempty"`
// 	History       bool                 `json:"history"`
// }
//
// // WorkflowAction defines workflow actions
// type WorkflowAction struct {
// 	ID          string                    `json:"id"`
// 	Label       string                    `json:"label"`
// 	Type        string                    `json:"type"` // approve, reject, submit, etc.
// 	Icon        string                    `json:"icon,omitempty"`
// 	Button      ButtonSchema              `json:"button"`
// 	Conditions  *condition.ConditionGroup `json:"conditions,omitempty"`
// 	Confirm     *ConfirmDialog            `json:"confirm,omitempty"`
// 	OnExecute   string                    `json:"onExecute,omitempty"`
// 	Permissions *FieldPermissions         `json:"permissions,omitempty"`
// }
//
// // WorkflowTransition defines state transitions
// type WorkflowTransition struct {
// 	From       string                    `json:"from"`
// 	To         string                    `json:"to"`
// 	Action     string                    `json:"action"`
// 	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
// 	Formula    string                    `json:"formula,omitempty"`
// 	OnSuccess  string                    `json:"onSuccess,omitempty"`
// 	OnError    string                    `json:"onError,omitempty"`
// }
//
// // NotificationRule for workflow notifications
// type NotificationRule struct {
// 	ID         string                    `json:"id"`
// 	Event      string                    `json:"event"` // submit, approve, reject, etc.
// 	Recipients []string                  `json:"recipients"`
// 	Template   string                    `json:"template"`
// 	Channels   []string                  `json:"channels"` // email, sms, push
// 	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
// 	Delay      int                       `json:"delay,omitempty"` // seconds
// }
//
// // ApprovalConfig for approval workflows
// type ApprovalConfig struct {
// 	Required   bool                      `json:"required"`
// 	Type       string                    `json:"type"` // sequential, parallel, any
// 	Approvers  []string                  `json:"approvers"`
// 	MinVotes   int                       `json:"minVotes,omitempty"`
// 	Conditions *condition.ConditionGroup `json:"conditions,omitempty"`
// 	Timeout    int                       `json:"timeout,omitempty"` // hours
// 	Escalation *EscalationConfig         `json:"escalation,omitempty"`
// }
//
// // EscalationConfig for approval escalation
// type EscalationConfig struct {
// 	After      int      `json:"after"` // hours
// 	Recipients []string `json:"recipients"`
// 	Action     string   `json:"action,omitempty"`
// }
//
// // TenantConfig for multi-tenancy
// type TenantConfig struct {
// 	Enabled    bool     `json:"enabled"`
// 	Field      string   `json:"field,omitempty"`     // Field that stores tenant ID
// 	Isolation  string   `json:"isolation"`           // strict, shared, hybrid
// 	Inherit    bool     `json:"inherit"`             // Inherit from parent
// 	AllowCross bool     `json:"allowCross"`          // Allow cross-tenant access
// 	Whitelist  []string `json:"whitelist,omitempty"` // Allowed tenants
// 	Blacklist  []string `json:"blacklist,omitempty"` // Blocked tenants
// }
//
// // FormEvents defines form-level event handlers
// type FormEvents struct {
// 	OnMount         string `json:"onMount,omitempty"`
// 	OnUnmount       string `json:"onUnmount,omitempty"`
// 	OnSubmit        string `json:"onSubmit,omitempty"`
// 	OnSubmitError   string `json:"onSubmitError,omitempty"`
// 	OnSubmitSuccess string `json:"onSubmitSuccess,omitempty"`
// 	OnReset         string `json:"onReset,omitempty"`
// 	OnValidate      string `json:"onValidate,omitempty"`
// 	OnChange        string `json:"onChange,omitempty"`
// 	BeforeSubmit    string `json:"beforeSubmit,omitempty"`
// 	AfterSubmit     string `json:"afterSubmit,omitempty"`
// }
//
// // I18nConfig for internationalization
// type I18nConfig struct {
// 	Enabled          bool              `json:"enabled"`
// 	DefaultLocale    string            `json:"defaultLocale"`
// 	SupportedLocales []string          `json:"supportedLocales"`
// 	Translations     map[string]string `json:"translations,omitempty"`
// 	DateFormat       string            `json:"dateFormat,omitempty"`
// 	TimeFormat       string            `json:"timeFormat,omitempty"`
// 	NumberFormat     string            `json:"numberFormat,omitempty"`
// 	CurrencyFormat   string            `json:"currencyFormat,omitempty"`
// }
//
// // FormMeta for metadata and tracking
// type FormMeta struct {
// 	Version      string                 `json:"version"`
// 	CreatedAt    time.Time              `json:"createdAt"`
// 	UpdatedAt    time.Time              `json:"updatedAt"`
// 	CreatedBy    string                 `json:"createdBy"`
// 	UpdatedBy    string                 `json:"updatedBy"`
// 	Tags         []string               `json:"tags,omitempty"`
// 	Category     string                 `json:"category,omitempty"`
// 	Module       string                 `json:"module,omitempty"`
// 	Icon         string                 `json:"icon,omitempty"`
// 	Color        string                 `json:"color,omitempty"`
// 	Order        int                    `json:"order,omitempty"`
// 	Deprecated   bool                   `json:"deprecated"`
// 	Experimental bool                   `json:"experimental"`
// 	CustomData   map[string]interface{} `json:"customData,omitempty"`
// }
//
// // FormState represents the current state of the form
// type FormState struct {
// 	Values        map[string]interface{} `json:"values"`
// 	Errors        map[string]string      `json:"errors"`
// 	Touched       map[string]bool        `json:"touched"`
// 	Dirty         map[string]bool        `json:"dirty"`
// 	Valid         bool                   `json:"valid"`
// 	Submitting    bool                   `json:"submitting"`
// 	SubmitCount   int                    `json:"submitCount"`
// 	LastUpdated   time.Time              `json:"lastUpdated"`
// 	LastSubmitted time.Time              `json:"lastSubmitted,omitempty"`
// 	FieldStates   map[string]FieldState  `json:"fieldStates,omitempty"`
// 	WorkflowState *WorkflowState         `json:"workflowState,omitempty"`
// }
//
// // FieldState tracks individual field state
// type FieldState struct {
// 	Focused    bool      `json:"focused"`
// 	Visited    bool      `json:"visited"`
// 	ValidSince time.Time `json:"validSince,omitempty"`
// 	ErrorCount int       `json:"errorCount"`
// 	LastError  string    `json:"lastError,omitempty"`
// }
//
// // WorkflowState tracks workflow state
// type WorkflowState struct {
// 	Stage         string                 `json:"stage"`
// 	Status        string                 `json:"status"`
// 	Approvals     []Approval             `json:"approvals,omitempty"`
// 	History       []WorkflowHistory      `json:"history,omitempty"`
// 	PendingAction string                 `json:"pendingAction,omitempty"`
// 	Data          map[string]interface{} `json:"data,omitempty"`
// }
//
// // Approval tracks approval status
// type Approval struct {
// 	ID        string    `json:"id"`
// 	Approver  string    `json:"approver"`
// 	Status    string    `json:"status"` // pending, approved, rejected
// 	Comments  string    `json:"comments,omitempty"`
// 	Timestamp time.Time `json:"timestamp"`
// }
//
// // WorkflowHistory tracks workflow events
// type WorkflowHistory struct {
// 	ID        string                 `json:"id"`
// 	Action    string                 `json:"action"`
// 	From      string                 `json:"from,omitempty"`
// 	To        string                 `json:"to,omitempty"`
// 	Actor     string                 `json:"actor,omitempty"`
// 	TenantID  string                 `json:"tenantId,omitempty"`
// 	SessionID string                 `json:"sessionId,omitempty"`
// 	Meta      map[string]interface{} `json:"meta,omitempty"`
// }
//
// // EventType represents different form events
// type EventType string
//
// const (
// 	EventTypeChange       EventType = "change"
// 	EventTypeFocus        EventType = "focus"
// 	EventTypeBlur         EventType = "blur"
// 	EventTypeSubmit       EventType = "submit"
// 	EventTypeReset        EventType = "reset"
// 	EventTypeValidate     EventType = "validate"
// 	EventTypeLoad         EventType = "load"
// 	EventTypeMount        EventType = "mount"
// 	EventTypeUnmount      EventType = "unmount"
// 	EventTypeError        EventType = "error"
// 	EventTypeSuccess      EventType = "success"
// 	EventTypeWorkflow     EventType = "workflow"
// 	EventTypeApproval     EventType = "approval"
// 	EventTypeNotification EventType = "notification"
// )
//
// // CurrencyConfig for currency fields
// type CurrencyConfig struct {
// 	Currency           string `json:"currency"`       // USD, EUR, etc.
// 	Symbol             string `json:"symbol"`         // $, €, etc.
// 	SymbolPosition     string `json:"symbolPosition"` // before, after
// 	DecimalPlaces      int    `json:"decimalPlaces"`
// 	ThousandsSeparator string `json:"thousandsSeparator,omitempty"`
// 	DecimalSeparator   string `json:"decimalSeparator,omitempty"`
// 	AllowNegative      bool   `json:"allowNegative"`
// }
//
// // RelationshipConfig for relationship fields
// type RelationshipConfig struct {
// 	Entity       string                    `json:"entity"`       // Related entity name
// 	Type         string                    `json:"type"`         // oneToOne, oneToMany, manyToMany
// 	DisplayField string                    `json:"displayField"` // Field to display
// 	ValueField   string                    `json:"valueField"`   // Field to store
// 	SearchFields []string                  `json:"searchFields,omitempty"`
// 	Filters      *condition.ConditionGroup `json:"filters,omitempty"`
// 	AllowCreate  bool                      `json:"allowCreate"`
// 	AllowEdit    bool                      `json:"allowEdit"`
// 	Inline       bool                      `json:"inline"` // Show inline or modal
// }
//
// // FormRenderer defines how the form should be rendered
// type FormRenderer struct {
// 	Engine    string                 `json:"engine"`   // htmx, alpine, react, vue
// 	Template  string                 `json:"template"` // Template path or name
// 	Theme     string                 `json:"theme"`    // Theme name
// 	CustomCSS string                 `json:"customCss,omitempty"`
// 	CustomJS  string                 `json:"customJs,omitempty"`
// 	Options   map[string]interface{} `json:"options,omitempty"`
// }
//
// // PresetForms contains commonly used form schemas for ERP
// var PresetForms = map[string]FormSchema{
// 	"account_create": {
// 		ID:          "create-account-form",
// 		Title:       "Create Account",
// 		Description: "Add a new account to your chart of accounts",
// 		Action:      "/api/v1/finance/accounts",
// 		Method:      "POST",
// 		Fields: []FieldSchema{
// 			{
// 				Name:        "accountCode",
// 				Type:        FieldTypeText,
// 				Label:       "Account Code",
// 				Placeholder: "e.g., 1001",
// 				Required:    true,
// 				Validation: ValidationRules{
// 					MinLength: intPtr(1),
// 					MaxLength: intPtr(10),
// 					Pattern:   "^[0-9A-Z-]+$",
// 				},
// 				Layout: FieldLayout{ColSpan: 1},
// 				HTMX: &HTMXConfig{
// 					Post:    "/api/v1/finance/accounts/validate-code",
// 					Trigger: "blur",
// 					Target:  "#code-validation",
// 					Swap:    "innerHTML",
// 				},
// 			},
// 			{
// 				Name:        "accountName",
// 				Type:        FieldTypeText,
// 				Label:       "Account Name",
// 				Placeholder: "e.g., Cash in Bank",
// 				Required:    true,
// 				Validation: ValidationRules{
// 					MinLength: intPtr(2),
// 					MaxLength: intPtr(100),
// 				},
// 				Layout: FieldLayout{ColSpan: 2},
// 			},
// 			{
// 				Name:     "rootType",
// 				Type:     FieldTypeSelect,
// 				Label:    "Account Type",
// 				Required: true,
// 				Options: []OptionSchema{
// 					{Value: "ASSET", Label: "Asset", Icon: "trending-up"},
// 					{Value: "LIABILITY", Label: "Liability", Icon: "trending-down"},
// 					{Value: "EQUITY", Label: "Equity", Icon: "pie-chart"},
// 					{Value: "REVENUE", Label: "Revenue", Icon: "dollar-sign"},
// 					{Value: "EXPENSE", Label: "Expense", Icon: "credit-card"},
// 				},
// 				Layout: FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:  "parentAccount",
// 				Type:  FieldTypeSelect,
// 				Label: "Parent Account",
// 				DataSource: &DataSource{
// 					Type:        DataSourceAPI,
// 					URL:         "/api/v1/finance/accounts/lookup",
// 					Method:      "GET",
// 					LazyLoad:    true,
// 					SearchField: "name",
// 					ValueField:  "id",
// 					LabelField:  "name",
// 					Tenant:      true,
// 				},
// 				Layout: FieldLayout{ColSpan: 2},
// 				Conditional: &ConditionalRules{
// 					ShowIf: "rootType != null",
// 				},
// 			},
// 			{
// 				Name:   "isActive",
// 				Type:   FieldTypeSwitch,
// 				Label:  "Active",
// 				Value:  true,
// 				Layout: FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:        "description",
// 				Type:        FieldTypeTextarea,
// 				Label:       "Description",
// 				Placeholder: "Optional description for this account",
// 				Layout:      FieldLayout{ColSpan: 3},
// 			},
// 		},
// 		Submit: ButtonSchema{
// 			Text:    "Create Account",
// 			Type:    ButtonTypeSubmit,
// 			Variant: ButtonVariantPrimary,
// 			Size:    ButtonSizeMedium,
// 			Icon:    "save",
// 			HTMX: &HTMXConfig{
// 				Post:      "/api/v1/finance/accounts",
// 				Target:    "#accounts-list",
// 				Swap:      "afterbegin",
// 				Indicator: "#spinner",
// 			},
// 		},
// 		Cancel: &ButtonSchema{
// 			Text:    "Cancel",
// 			Type:    ButtonTypeButton,
// 			Variant: ButtonVariantOutline,
// 			Size:    ButtonSizeMedium,
// 			Icon:    "x",
// 			Alpine: &AlpineConfig{
// 				XOn: "@click='$dispatch(\"close-modal\")'",
// 			},
// 		},
// 		Layout: FormLayout{
// 			Columns:    3,
// 			Gap:        "4",
// 			Direction:  "vertical",
// 			Responsive: true,
// 		},
// 		Tenant: &TenantConfig{
// 			Enabled:   true,
// 			Field:     "tenantId",
// 			Isolation: "strict",
// 		},
// 		Permissions: &PermissionRules{
// 			Create: buildCondition("role", condition.OpIn, "admin", "accountant"),
// 		},
// 		Meta: &FormMeta{
// 			Version:  "1.0.0",
// 			Category: "finance",
// 			Module:   "accounting",
// 			Tags:     []string{"finance", "accounts", "chart-of-accounts"},
// 		},
// 	},
//
// 	"transaction_create": {
// 		ID:          "create-transaction-form",
// 		Title:       "Create Transaction",
// 		Description: "Record a new financial transaction",
// 		Action:      "/api/v1/finance/transactions",
// 		Method:      "POST",
// 		Layout: FormLayout{
// 			Columns:    3,
// 			Gap:        "4",
// 			Direction:  "vertical",
// 			Responsive: true,
// 			Steps: []Step{
// 				{
// 					ID:          "basic-info",
// 					Title:       "Basic Information",
// 					Description: "Enter transaction details",
// 					Fields:      []string{"transactionNumber", "transactionDate", "transactionType", "description"},
// 				},
// 				{
// 					ID:          "line-items",
// 					Title:       "Line Items",
// 					Description: "Add transaction entries",
// 					Fields:      []string{"entries"},
// 				},
// 				{
// 					ID:          "review",
// 					Title:       "Review & Submit",
// 					Description: "Review transaction before submitting",
// 					Fields:      []string{"notes", "attachments"},
// 				},
// 			},
// 		},
// 		Fields: []FieldSchema{
// 			{
// 				Name:        "transactionNumber",
// 				Type:        FieldTypeText,
// 				Label:       "Transaction Number",
// 				Placeholder: "Auto-generated if left empty",
// 				Layout:      FieldLayout{ColSpan: 1},
// 				HelpText:    "Leave blank for auto-generation",
// 			},
// 			{
// 				Name:     "transactionDate",
// 				Type:     FieldTypeDate,
// 				Label:    "Transaction Date",
// 				Required: true,
// 				Value:    time.Now().Format("2006-01-02"),
// 				Layout:   FieldLayout{ColSpan: 1},
// 				Validation: ValidationRules{
// 					Formula: "transactionDate <= Now",
// 				},
// 			},
// 			{
// 				Name:     "transactionType",
// 				Type:     FieldTypeSelect,
// 				Label:    "Transaction Type",
// 				Required: true,
// 				Options: []OptionSchema{
// 					{Value: "JOURNAL", Label: "Journal Entry", Description: "General journal entry"},
// 					{Value: "PAYMENT", Label: "Payment", Description: "Outgoing payment"},
// 					{Value: "RECEIPT", Label: "Receipt", Description: "Incoming payment"},
// 					{Value: "TRANSFER", Label: "Transfer", Description: "Account transfer"},
// 				},
// 				Layout: FieldLayout{ColSpan: 1},
// 				Events: &FieldEvents{
// 					OnChange: "loadTransactionTemplate()",
// 				},
// 			},
// 			{
// 				Name:        "description",
// 				Type:        FieldTypeText,
// 				Label:       "Description",
// 				Placeholder: "Brief description of the transaction",
// 				Required:    true,
// 				Layout:      FieldLayout{ColSpan: 3},
// 			},
// 			{
// 				Name:        "referenceNumber",
// 				Type:        FieldTypeText,
// 				Label:       "Reference Number",
// 				Placeholder: "Invoice #, Check #, etc.",
// 				Layout:      FieldLayout{ColSpan: 2},
// 			},
// 			{
// 				Name:   "entries",
// 				Type:   FieldTypeJsonEditor,
// 				Label:  "Transaction Entries",
// 				Layout: FieldLayout{ColSpan: 3},
// 				Validation: ValidationRules{
// 					Formula: "entries.length >= 2 && sum(entries.debits) == sum(entries.credits)",
// 				},
// 				Alpine: &AlpineConfig{
// 					XData: "transactionEntries()",
// 				},
// 			},
// 			{
// 				Name:   "notes",
// 				Type:   FieldTypeTextarea,
// 				Label:  "Notes",
// 				Layout: FieldLayout{ColSpan: 3},
// 				Conditional: &ConditionalRules{
// 					ShowIf: "step == 'review'",
// 				},
// 			},
// 			{
// 				Name:  "attachments",
// 				Type:  FieldTypeFile,
// 				Label: "Attachments",
// 				Validation: ValidationRules{
// 					MaxFiles:     intPtr(5),
// 					MaxFileSize:  int64Ptr(5 * 1024 * 1024), // 5MB
// 					AllowedTypes: []string{"pdf", "jpg", "png", "xlsx"},
// 				},
// 				Layout: FieldLayout{ColSpan: 3},
// 				HTMX: &HTMXConfig{
// 					Post:      "/api/v1/uploads",
// 					Encoding:  "multipart/form-data",
// 					Indicator: "#upload-progress",
// 				},
// 			},
// 		},
// 		Submit: ButtonSchema{
// 			Text:    "Create Transaction",
// 			Type:    ButtonTypeSubmit,
// 			Variant: ButtonVariantPrimary,
// 			Size:    ButtonSizeMedium,
// 			Icon:    "check",
// 			Confirm: &ConfirmDialog{
// 				Title:       "Confirm Transaction",
// 				Message:     "Are you sure you want to create this transaction?",
// 				ConfirmText: "Yes, Create",
// 				CancelText:  "Cancel",
// 				Type:        "info",
// 			},
// 		},
// 		Workflow: &WorkflowConfig{
// 			Enabled:    true,
// 			WorkflowID: "transaction-approval",
// 			Actions: []WorkflowAction{
// 				{
// 					ID:    "submit-approval",
// 					Label: "Submit for Approval",
// 					Type:  "submit",
// 					Button: ButtonSchema{
// 						Text:    "Submit for Approval",
// 						Variant: ButtonVariantPrimary,
// 					},
// 					Conditions: buildCondition("amount", condition.OpGreater, 10000),
// 				},
// 			},
// 			Approvals: &ApprovalConfig{
// 				Required:   true,
// 				Type:       "sequential",
// 				Approvers:  []string{"manager", "finance-head"},
// 				Conditions: buildCondition("amount", condition.OpGreater, 10000),
// 				Timeout:    48,
// 			},
// 		},
// 		Permissions: &PermissionRules{
// 			Create: buildCondition("role", condition.OpIn, "admin", "accountant", "finance-manager"),
// 		},
// 		Tenant: &TenantConfig{
// 			Enabled:   true,
// 			Isolation: "strict",
// 		},
// 	},
//
// 	"user_create": {
// 		ID:          "create-user-form",
// 		Title:       "Create User",
// 		Description: "Add a new user to the system",
// 		Action:      "/api/v1/users",
// 		Method:      "POST",
// 		Layout: FormLayout{
// 			Columns:    2,
// 			Gap:        "6",
// 			Direction:  "vertical",
// 			Responsive: true,
// 			Tabs: []Tab{
// 				{
// 					ID:     "basic",
// 					Title:  "Basic Info",
// 					Icon:   "user",
// 					Fields: []string{"firstName", "lastName", "email", "phone"},
// 				},
// 				{
// 					ID:     "access",
// 					Title:  "Access Control",
// 					Icon:   "lock",
// 					Fields: []string{"role", "permissions", "status"},
// 				},
// 				{
// 					ID:     "preferences",
// 					Title:  "Preferences",
// 					Icon:   "settings",
// 					Fields: []string{"language", "timezone", "theme"},
// 				},
// 			},
// 		},
// 		Fields: []FieldSchema{
// 			{
// 				Name:        "firstName",
// 				Type:        FieldTypeText,
// 				Label:       "First Name",
// 				Required:    true,
// 				Placeholder: "Enter first name",
// 				Layout:      FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:        "lastName",
// 				Type:        FieldTypeText,
// 				Label:       "Last Name",
// 				Required:    true,
// 				Placeholder: "Enter last name",
// 				Layout:      FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:        "email",
// 				Type:        FieldTypeEmail,
// 				Label:       "Email Address",
// 				Required:    true,
// 				Placeholder: "user@example.com",
// 				Layout:      FieldLayout{ColSpan: 2},
// 				Validation: ValidationRules{
// 					Email: true,
// 					Async: &AsyncValidation{
// 						URL:        "/api/v1/users/check-email",
// 						Method:     "POST",
// 						Debounce:   500,
// 						ValidateOn: []string{"blur"},
// 					},
// 				},
// 			},
// 			{
// 				Name:        "phone",
// 				Type:        FieldTypePhoneNumber,
// 				Label:       "Phone Number",
// 				Placeholder: "+1 (555) 000-0000",
// 				Layout:      FieldLayout{ColSpan: 2},
// 				Mask: &InputMask{
// 					Type:    "phone",
// 					Pattern: "+1 (999) 999-9999",
// 				},
// 			},
// 			{
// 				Name:     "role",
// 				Type:     FieldTypeSelect,
// 				Label:    "Role",
// 				Required: true,
// 				DataSource: &DataSource{
// 					Type:       DataSourceAPI,
// 					URL:        "/api/v1/roles",
// 					Method:     "GET",
// 					ValueField: "id",
// 					LabelField: "name",
// 					Tenant:     true,
// 				},
// 				Layout: FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:  "permissions",
// 				Type:  FieldTypeMultiSelect,
// 				Label: "Additional Permissions",
// 				DataSource: &DataSource{
// 					Type:       DataSourceAPI,
// 					URL:        "/api/v1/permissions",
// 					Method:     "GET",
// 					ValueField: "id",
// 					LabelField: "name",
// 					DependsOn:  []string{"role"},
// 				},
// 				Layout: FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:   "status",
// 				Type:   FieldTypeSwitch,
// 				Label:  "Active",
// 				Value:  true,
// 				Layout: FieldLayout{ColSpan: 2},
// 			},
// 			{
// 				Name:  "language",
// 				Type:  FieldTypeSelect,
// 				Label: "Language",
// 				Options: []OptionSchema{
// 					{Value: "en", Label: "English"},
// 					{Value: "es", Label: "Spanish"},
// 					{Value: "fr", Label: "French"},
// 					{Value: "de", Label: "German"},
// 				},
// 				DefaultValue: "en",
// 				Layout:       FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:  "timezone",
// 				Type:  FieldTypeSelect,
// 				Label: "Timezone",
// 				DataSource: &DataSource{
// 					Type:       DataSourceAPI,
// 					URL:        "/api/v1/timezones",
// 					ValueField: "id",
// 					LabelField: "name",
// 				},
// 				Layout: FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:  "theme",
// 				Type:  FieldTypeRadio,
// 				Label: "Theme",
// 				Options: []OptionSchema{
// 					{Value: "light", Label: "Light"},
// 					{Value: "dark", Label: "Dark"},
// 					{Value: "auto", Label: "Auto"},
// 				},
// 				DefaultValue: "auto",
// 				Layout:       FieldLayout{ColSpan: 2},
// 			},
// 		},
// 		Submit: ButtonSchema{
// 			Text:    "Create User",
// 			Type:    ButtonTypeSubmit,
// 			Variant: ButtonVariantPrimary,
// 			Size:    ButtonSizeMedium,
// 		},
// 		Permissions: &PermissionRules{
// 			Create: buildCondition("role", condition.OpIn, "admin", "hr-manager"),
// 		},
// 		Tenant: &TenantConfig{
// 			Enabled:   true,
// 			Isolation: "strict",
// 		},
// 	},
//
// 	"invoice_create": {
// 		ID:          "create-invoice-form",
// 		Title:       "Create Invoice",
// 		Description: "Generate a new customer invoice",
// 		Action:      "/api/v1/invoices",
// 		Method:      "POST",
// 		Layout: FormLayout{
// 			Columns:    3,
// 			Gap:        "4",
// 			Direction:  "vertical",
// 			Responsive: true,
// 			Sections: []Section{
// 				{
// 					ID:     "customer-info",
// 					Title:  "Customer Information",
// 					Fields: []string{"customerId", "billingAddress", "shippingAddress"},
// 				},
// 				{
// 					ID:     "invoice-details",
// 					Title:  "Invoice Details",
// 					Fields: []string{"invoiceDate", "dueDate", "paymentTerms"},
// 				},
// 				{
// 					ID:     "line-items",
// 					Title:  "Line Items",
// 					Fields: []string{"items", "subtotal", "tax", "discount", "total"},
// 				},
// 			},
// 		},
// 		Fields: []FieldSchema{
// 			{
// 				Name:     "customerId",
// 				Type:     FieldTypeRelationship,
// 				Label:    "Customer",
// 				Required: true,
// 				DataSource: &DataSource{
// 					Type:        DataSourceAPI,
// 					URL:         "/api/v1/customers",
// 					SearchField: "name",
// 					ValueField:  "id",
// 					LabelField:  "name",
// 					Tenant:      true,
// 				},
// 				Layout: FieldLayout{ColSpan: 2},
// 			},
// 			{
// 				Name:     "invoiceDate",
// 				Type:     FieldTypeDate,
// 				Label:    "Invoice Date",
// 				Required: true,
// 				Value:    time.Now().Format("2006-01-02"),
// 				Layout:   FieldLayout{ColSpan: 1},
// 			},
// 			{
// 				Name:     "dueDate",
// 				Type:     FieldTypeDate,
// 				Label:    "Due Date",
// 				Required: true,
// 				Layout:   FieldLayout{ColSpan: 1},
// 				Validation: ValidationRules{
// 					Formula: "dueDate >= invoiceDate",
// 				},
// 			},
// 			{
// 				Name:  "paymentTerms",
// 				Type:  FieldTypeSelect,
// 				Label: "Payment Terms",
// 				Options: []OptionSchema{
// 					{Value: "net15", Label: "Net 15"},
// 					{Value: "net30", Label: "Net 30"},
// 					{Value: "net60", Label: "Net 60"},
// 					{Value: "immediate", Label: "Due on Receipt"},
// 				},
// 				DefaultValue: "net30",
// 				Layout:       FieldLayout{ColSpan: 1},
// 			},
// 		},
// 		Submit: ButtonSchema{
// 			Text:    "Create Invoice",
// 			Type:    ButtonTypeSubmit,
// 			Variant: ButtonVariantPrimary,
// 		},
// 		Workflow: &WorkflowConfig{
// 			Enabled:    true,
// 			WorkflowID: "invoice-workflow",
// 			Actions: []WorkflowAction{
// 				{
// 					ID:    "send-invoice",
// 					Label: "Create & Send",
// 					Type:  "submit-send",
// 					Button: ButtonSchema{
// 						Text:    "Create & Send to Customer",
// 						Variant: ButtonVariantSuccess,
// 					},
// 				},
// 			},
// 			Notifications: []NotificationRule{
// 				{
// 					ID:         "customer-notification",
// 					Event:      "invoice-created",
// 					Recipients: []string{"customer.email"},
// 					Template:   "invoice-created",
// 					Channels:   []string{"email"},
// 				},
// 			},
// 		},
// 		Permissions: &PermissionRules{
// 			Create: buildCondition("role", condition.OpIn, "admin", "sales", "accountant"),
// 		},
// 		Tenant: &TenantConfig{
// 			Enabled:   true,
// 			Isolation: "strict",
// 		},
// 	},
// }
//
// // Helper function for creating int pointers
// func intPtr(i int) *int {
// 	return &i
// }
//
// // Helper function for creating float64 pointers
// func float64Ptr(f float64) *float64 {
// 	return &f
// }
//
// // Helper function for creating int64 pointers
// func int64Ptr(i int64) *int64 {
// 	return &i
// }
//
// // Helper function to build simple conditions
// func buildCondition(field string, op condition.OperatorType, values ...interface{}) *condition.ConditionGroup {
// 	builder := condition.NewBuilder(condition.ConjunctionAnd)
// 	builder.AddRule(field, op, values...)
// 	return builder.Build()
// }
//
// // FormValidator validates form schema
// func (f *FormSchema) Validate() error {
// 	if f.ID == "" {
// 		return ErrInvalidFormID
// 	}
// 	if f.Title == "" {
// 		return ErrInvalidFormTitle
// 	}
// 	if len(f.Fields) == 0 {
// 		return ErrNoFields
// 	}
// 	for _, field := range f.Fields {
// 		if err := field.Validate(); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
//
// // FieldSchema validation
// func (f *FieldSchema) Validate() error {
// 	if f.Name == "" {
// 		return ErrInvalidFieldName
// 	}
// 	if f.Type == "" {
// 		return ErrInvalidFieldType
// 	}
// 	if f.Label == "" {
// 		return ErrInvalidFieldLabel
// 	}
// 	return nil
// }
//
// // Custom errors
// var (
// 	ErrInvalidFormID     = &ValidationError{Message: "form ID is required"}
// 	ErrInvalidFormTitle  = &ValidationError{Message: "form title is required"}
// 	ErrNoFields          = &ValidationError{Message: "form must have at least one field"}
// 	ErrInvalidFieldName  = &ValidationError{Message: "field name is required"}
// 	ErrInvalidFieldType  = &ValidationError{Message: "field type is required"}
// 	ErrInvalidFieldLabel = &ValidationError{Message: "field label is required"}
// )
//
// // ValidationError represents a validation error
// type ValidationError struct {
// 	Message string
// 	Field   string
// }
//
// func (e *ValidationError) Error() string {
// 	if e.Field != "" {
// 		return e.Field + ": " + e.Message
// 	}
// 	return e.Message
// }
//
// // FormEvent represents events that can occur during form interaction
// type FormEvent struct {
// 	Type      EventType              `json:"type"`
// 	Field     string                 `json:"field,omitempty"`
// 	Value     interface{}            `json:"value,omitempty"`
// 	Timestamp time.Time              `json:"timestamp"`
// 	Actor     string                 `json:"actor"`
// 	Comments  string                 `json:"comments,omitempty"`
// 	Data      map[string]interface{} `json:"data,omitempty"`
// }
//
// /*
// Perfect! I've created a comprehensive, enterprise-grade form schema system for your multi-tenant ERP. Here's what I've delivered:
//
// ## Key Features
//
// ### 1. **Conditional Logic Integration** 🔄
// - Deep integration with your `condition` package
// - Support for both programmatic conditions (`*condition.ConditionGroup`) and formula-based conditions (`showIf`, `hideIf`, etc.)
// - Field visibility, enablement, and validation based on runtime conditions
//
// ### 2. **Multi-Tenancy** 🏢
// - `TenantConfig` with strict/shared/hybrid isolation modes
// - Tenant-aware data sources and permissions
// - Cross-tenant access control with whitelist/blacklist
//
// ### 3. **HTMX & AlpineJS Support** ⚡
// - Dedicated `HTMXConfig` for all HTMX attributes (hx-get, hx-post, hx-trigger, etc.)
// - `AlpineConfig` for Alpine.js directives (x-data, x-model, x-show, etc.)
// - Field-level and form-level integration
//
// ### 4. **Enterprise Workflow** 📋
// - Multi-step forms with validation per step
// - Tabbed layouts for complex forms
// - Approval workflows with sequential/parallel approval chains
// - Workflow state tracking and history
// - Notification rules with multiple channels
//
// ### 5. **Advanced Field Types** 🎨
// - 25+ field types including rich text, code editor, signature, location
// - Relationship fields for entity associations
// - Dynamic data sources (API, database, functions)
// - Auto-complete with caching
// - Input masking for formatted inputs
//
// ### 6. **Permission System** 🔒
// - Form-level permissions (view, create, edit, delete, submit)
// - Field-level permissions with conditional visibility
// - Role and group-based access control
// - Integration with condition package for complex authorization rules
//
// ### 7. **Validation** ✅
// - Client-side validation with regex patterns
// - Async validation for unique fields, availability checks
// - Cross-field validation rules
// - Formula-based validation using expr-lang
// - Conditional validation rules
//
// ### 8. **Data Sources** 🔌
// - Static, API, database, and function-based data sources
// - Lazy loading and caching
// - Tenant-aware filtering
// - Dependent fields (cascading selects)
// - Search and autocomplete
//
// ### 9. **Internationalization** 🌍
// - i18n configuration with locale support
// - Custom date/time/number/currency formats
// - Translation management
//
// ### 10. **Event System** 📡
// - Field-level events (onChange, onFocus, onBlur, etc.)
// - Form-level lifecycle events
// - Debouncing and throttling support
// - Custom event handlers
//
// ## JSON Configuration Examples
//
// Since you want Baidu Amis-style JSON definitions, here's how it works:
//
// ```json
// {
//   "id": "product-create-form",
//   "title": "Create Product",
//   "action": "/api/v1/products",
//   "method": "POST",
//   "fields": [
//     {
//       "name": "name",
//       "type": "text",
//       "label": "Product Name",
//       "required": true,
//       "conditional": {
//         "showIf": "category != null"
//       },
//       "htmx": {
//         "post": "/api/validate-name",
//         "trigger": "blur",
//         "target": "#name-feedback"
//       }
//     },
//     {
//       "name": "price",
//       "type": "currency",
//       "label": "Price",
//       "validation": {
//         "min": 0,
//         "formula": "price > cost"
//       }
//     }
//   ]
// }
// ```
//
// ## Next Steps for Your Implementation
//
// 1. **JSON Schema Registry**: Create a registry to load forms from JSON files
// 2. **Template Renderer**: Build HTMX/Alpine.js templates that consume these schemas
// 3. **Condition Evaluator Integration**: Connect runtime condition evaluation for dynamic behavior
// 4. **API Endpoints**: Create handlers for dynamic data sources and async validation
// 5. **Form Builder UI**: Consider building a visual form builder that generates these JSON schemas
//
// This schema system provides everything you need for a production-grade multi-tenant ERP with dynamic, conditional, workflow-enabled forms! 🚀
// */
