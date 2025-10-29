// Package schema provides a comprehensive, production-ready JSON-driven UI schema system
// for enterprise applications with full validation, multi-tenancy, and framework integration.
package schema

import (
	"context"
	"encoding/json"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// Schema represents the unified schema for all UI components and forms
type Schema struct {
	// Core identification and metadata
	ID          string `json:"id" validate:"required,min=1,max=100" example:"user-create-form"`
	Type        Type   `json:"type" validate:"required" example:"form"`
	Version     string `json:"version,omitempty" validate:"semver" example:"1.2.0"`
	Title       string `json:"title" validate:"required,min=1,max=200" example:"Create User"`
	Description string `json:"description,omitempty" validate:"max=1000" example:"Add a new user to the system"`

	// Core configuration
	Config  *Config  `json:"config,omitempty"`
	Layout  *Layout  `json:"layout,omitempty"`
	Fields  []Field  `json:"fields,omitempty" validate:"dive"`
	Actions []Action `json:"actions,omitempty" validate:"dive"`

	// Enterprise features
	Security   *Security   `json:"security,omitempty"`
	Tenant     *Tenant     `json:"tenant,omitempty"`
	Workflow   *Workflow   `json:"workflow,omitempty"`
	Validation *Validation `json:"validation,omitempty"`
	Events     *Events     `json:"events,omitempty"`
	I18n       *I18n       `json:"i18n,omitempty"`

	// Framework integration
	HTMX   *HTMX   `json:"htmx,omitempty"`
	Alpine *Alpine `json:"alpine,omitempty"`

	// Metadata and tracking
	Meta     *Meta    `json:"meta,omitempty"`
	Tags     []string `json:"tags,omitempty" validate:"dive,alphanum"`
	Category string   `json:"category,omitempty" validate:"alphanum"`
	Module   string   `json:"module,omitempty" validate:"alphanum"`

	// Runtime state
	State   *State   `json:"state,omitempty"`
	Context *Context `json:"context,omitempty"`
}

// Type defines the schema type
type Type string

const (
	TypeForm      Type = "form"
	TypeComponent Type = "component"
	TypeLayout    Type = "layout"
	TypeWorkflow  Type = "workflow"
	TypeTheme     Type = "theme"
	TypePage      Type = "page"
)

// Config holds core configuration
type Config struct {
	Action   string            `json:"action,omitempty" validate:"url" example:"/api/v1/users"`
	Method   string            `json:"method,omitempty" validate:"oneof=GET POST PUT PATCH DELETE" example:"POST"`
	Target   string            `json:"target,omitempty" validate:"html_id" example:"#main-content"`
	Encoding string            `json:"encoding,omitempty" validate:"oneof=application/json multipart/form-data application/x-www-form-urlencoded"`
	Timeout  int               `json:"timeout,omitempty" validate:"min=0,max=300000" example:"30000"`
	Headers  map[string]string `json:"headers,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
	Cache    bool              `json:"cache,omitempty"`
	CacheTTL int               `json:"cacheTTL,omitempty" validate:"min=0" example:"300"`
}

// Layout defines the visual layout structure
type Layout struct {
	Type       LayoutType `json:"type" validate:"required" example:"grid"`
	Columns    int        `json:"columns,omitempty" validate:"min=1,max=24" example:"3"`
	Gap        string     `json:"gap,omitempty" validate:"css_size" example:"1rem"`
	Direction  string     `json:"direction,omitempty" validate:"oneof=row column row-reverse column-reverse" example:"column"`
	Wrap       bool       `json:"wrap,omitempty"`
	Responsive bool       `json:"responsive,omitempty"`

	// Advanced layout options
	Sections []Section `json:"sections,omitempty" validate:"dive"`
	Groups   []Group   `json:"groups,omitempty" validate:"dive"`
	Tabs     []Tab     `json:"tabs,omitempty" validate:"dive"`
	Steps    []Step    `json:"steps,omitempty" validate:"dive"`

	// Responsive breakpoints
	Breakpoints *Breakpoints `json:"breakpoints,omitempty"`
}

// LayoutType defines layout types
type LayoutType string

const (
	LayoutGrid     LayoutType = "grid"
	LayoutFlex     LayoutType = "flex"
	LayoutTabs     LayoutType = "tabs"
	LayoutSteps    LayoutType = "steps"
	LayoutSections LayoutType = "sections"
	LayoutGroups   LayoutType = "groups"
)

// Field represents a form field or UI component
type Field struct {
	// Core properties
	Name        string    `json:"name" validate:"required,min=1,max=100" example:"email"`
	Type        FieldType `json:"type" validate:"required" example:"email"`
	Label       string    `json:"label" validate:"required,min=1,max=200" example:"Email Address"`
	Placeholder string    `json:"placeholder,omitempty" validate:"max=200" example:"user@example.com"`

	// Field state
	Required bool `json:"required,omitempty"`
	Disabled bool `json:"disabled,omitempty"`
	Readonly bool `json:"readonly,omitempty"`
	Hidden   bool `json:"hidden,omitempty"`

	// Values and options
	Value   interface{} `json:"value,omitempty"`
	Default interface{} `json:"default,omitempty"`
	Options []Option    `json:"options,omitempty" validate:"dive"`

	// Validation and constraints
	Validation *FieldValidation `json:"validation,omitempty"`
	Transform  *Transform       `json:"transform,omitempty"`
	Mask       *Mask            `json:"mask,omitempty"`

	// Layout and positioning
	Layout *FieldLayout `json:"layout,omitempty"`
	Style  *Style       `json:"style,omitempty"`

	// Behavior and interaction
	Events      *FieldEvents `json:"events,omitempty"`
	Conditional *Conditional `json:"conditional,omitempty"`
	DataSource  *DataSource  `json:"dataSource,omitempty"`

	// Help and accessibility
	Help    string `json:"help,omitempty" validate:"max=500"`
	Error   string `json:"error,omitempty" validate:"max=200"`
	Icon    string `json:"icon,omitempty" validate:"icon_name"`
	Tooltip string `json:"tooltip,omitempty" validate:"max=200"`

	// Framework integration
	HTMX   *FieldHTMX   `json:"htmx,omitempty"`
	Alpine *FieldAlpine `json:"alpine,omitempty"`

	// Permissions and security
	Permissions *FieldPermissions `json:"permissions,omitempty"`

	// Dependencies
	Dependencies []string `json:"dependencies,omitempty" validate:"dive,fieldname"`
}

// FieldType defines all supported field types
type FieldType string

const (
	// Basic input types
	FieldText     FieldType = "text"
	FieldEmail    FieldType = "email"
	FieldPassword FieldType = "password"
	FieldNumber   FieldType = "number"
	FieldHidden   FieldType = "hidden"

	// Date and time
	FieldDate      FieldType = "date"
	FieldTime      FieldType = "time"
	FieldDateTime  FieldType = "datetime"
	FieldDateRange FieldType = "daterange"

	// Text areas and rich content
	FieldTextarea FieldType = "textarea"
	FieldRichText FieldType = "richtext"
	FieldCode     FieldType = "code"
	FieldJSON     FieldType = "json"

	// Selection types
	FieldSelect      FieldType = "select"
	FieldMultiSelect FieldType = "multiselect"
	FieldRadio       FieldType = "radio"
	FieldCheckbox    FieldType = "checkbox"
	FieldTreeSelect  FieldType = "treeselect"
	FieldCascader    FieldType = "cascader"
	FieldTransfer    FieldType = "transfer"

	// Interactive controls
	FieldSwitch FieldType = "switch"
	FieldSlider FieldType = "slider"
	FieldRating FieldType = "rating"
	FieldColor  FieldType = "color"

	// File and media
	FieldFile      FieldType = "file"
	FieldImage     FieldType = "image"
	FieldSignature FieldType = "signature"

	// Specialized inputs
	FieldPhone    FieldType = "phone"
	FieldURL      FieldType = "url"
	FieldCurrency FieldType = "currency"
	FieldTags     FieldType = "tags"
	FieldLocation FieldType = "location"

	// Relationship and data
	FieldRelation     FieldType = "relation"
	FieldAutoComplete FieldType = "autocomplete"

	// Display components
	FieldDisplay FieldType = "display"
	FieldDivider FieldType = "divider"
	FieldHTML    FieldType = "html"
)

// Action represents form actions and buttons
type Action struct {
	ID       string     `json:"id" validate:"required" example:"submit-btn"`
	Type     ActionType `json:"type" validate:"required" example:"submit"`
	Text     string     `json:"text" validate:"required,min=1,max=100" example:"Create User"`
	Variant  string     `json:"variant,omitempty" validate:"oneof=primary secondary outline ghost destructive" example:"primary"`
	Size     string     `json:"size,omitempty" validate:"oneof=sm md lg xl" example:"md"`
	Icon     string     `json:"icon,omitempty" validate:"icon_name" example:"user-plus"`
	Position string     `json:"position,omitempty" validate:"oneof=left right" example:"left"`

	// State and behavior
	Loading  bool `json:"loading,omitempty"`
	Disabled bool `json:"disabled,omitempty"`
	Hidden   bool `json:"hidden,omitempty"`

	// Configuration
	Config  *ActionConfig `json:"config,omitempty"`
	Confirm *Confirm      `json:"confirm,omitempty"`

	// Framework integration
	HTMX   *ActionHTMX   `json:"htmx,omitempty"`
	Alpine *ActionAlpine `json:"alpine,omitempty"`

	// Conditional display
	Conditional *Conditional       `json:"conditional,omitempty"`
	Permissions *ActionPermissions `json:"permissions,omitempty"`
}

// ActionType defines action types
type ActionType string

const (
	ActionSubmit ActionType = "submit"
	ActionReset  ActionType = "reset"
	ActionButton ActionType = "button"
	ActionLink   ActionType = "link"
	ActionCustom ActionType = "custom"
)

// Option represents select/radio/checkbox options
type Option struct {
	Value       string      `json:"value" validate:"required"`
	Label       string      `json:"label" validate:"required"`
	Description string      `json:"description,omitempty" validate:"max=200"`
	Icon        string      `json:"icon,omitempty" validate:"icon_name"`
	Color       string      `json:"color,omitempty" validate:"css_color"`
	Group       string      `json:"group,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Selected    bool        `json:"selected,omitempty"`
	Children    []Option    `json:"children,omitempty" validate:"dive"`
	Meta        interface{} `json:"meta,omitempty"`
}

// Security defines security and access control
type Security struct {
	CSRF         *CSRF       `json:"csrf,omitempty"`
	RateLimit    *RateLimit  `json:"rateLimit,omitempty"`
	Encryption   *Encryption `json:"encryption,omitempty"`
	Sanitization bool        `json:"sanitization,omitempty"`
	Origins      []string    `json:"origins,omitempty" validate:"dive,url"`
	ContentType  []string    `json:"contentType,omitempty"`
}

// Tenant defines multi-tenancy configuration
type Tenant struct {
	Enabled    bool     `json:"enabled"`
	Field      string   `json:"field,omitempty" validate:"fieldname" example:"tenant_id"`
	Isolation  string   `json:"isolation" validate:"oneof=strict shared hybrid" example:"strict"`
	Inherit    bool     `json:"inherit,omitempty"`
	AllowCross bool     `json:"allowCross,omitempty"`
	Whitelist  []string `json:"whitelist,omitempty" validate:"dive,uuid"`
	Blacklist  []string `json:"blacklist,omitempty" validate:"dive,uuid"`
}

// Workflow defines workflow and approval configuration
type Workflow struct {
	Enabled       bool                 `json:"enabled"`
	ID            string               `json:"id,omitempty" validate:"uuid"`
	Stage         string               `json:"stage,omitempty"`
	Status        string               `json:"status,omitempty"`
	Actions       []WorkflowAction     `json:"actions,omitempty" validate:"dive"`
	Transitions   []WorkflowTransition `json:"transitions,omitempty" validate:"dive"`
	Notifications []Notification       `json:"notifications,omitempty" validate:"dive"`
	Approvals     *ApprovalConfig      `json:"approvals,omitempty"`
	History       bool                 `json:"history,omitempty"`
}

// Events defines event handling configuration
type Events struct {
	OnMount         string `json:"onMount,omitempty" validate:"js_function"`
	OnUnmount       string `json:"onUnmount,omitempty" validate:"js_function"`
	OnSubmit        string `json:"onSubmit,omitempty" validate:"js_function"`
	OnSubmitSuccess string `json:"onSubmitSuccess,omitempty" validate:"js_function"`
	OnSubmitError   string `json:"onSubmitError,omitempty" validate:"js_function"`
	OnReset         string `json:"onReset,omitempty" validate:"js_function"`
	OnValidate      string `json:"onValidate,omitempty" validate:"js_function"`
	OnChange        string `json:"onChange,omitempty" validate:"js_function"`
	BeforeSubmit    string `json:"beforeSubmit,omitempty" validate:"js_function"`
	AfterSubmit     string `json:"afterSubmit,omitempty" validate:"js_function"`
}

// HTMX defines HTMX integration configuration
type HTMX struct {
	Enabled   bool              `json:"enabled"`
	Get       string            `json:"get,omitempty" validate:"url"`
	Post      string            `json:"post,omitempty" validate:"url"`
	Put       string            `json:"put,omitempty" validate:"url"`
	Delete    string            `json:"delete,omitempty" validate:"url"`
	Patch     string            `json:"patch,omitempty" validate:"url"`
	Trigger   string            `json:"trigger,omitempty"`
	Target    string            `json:"target,omitempty" validate:"css_selector"`
	Swap      string            `json:"swap,omitempty" validate:"oneof=innerHTML outerHTML beforebegin afterbegin beforeend afterend delete none"`
	Select    string            `json:"select,omitempty" validate:"css_selector"`
	Indicator string            `json:"indicator,omitempty" validate:"css_selector"`
	PushURL   string            `json:"pushUrl,omitempty" validate:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	Vals      string            `json:"vals,omitempty"`
	Confirm   string            `json:"confirm,omitempty"`
	Boost     bool              `json:"boost,omitempty"`
	Sync      string            `json:"sync,omitempty"`
	Validate  bool              `json:"validate,omitempty"`
}

// Alpine defines Alpine.js integration configuration
type Alpine struct {
	Enabled     bool   `json:"enabled"`
	XData       string `json:"xData,omitempty" validate:"js_object"`
	XInit       string `json:"xInit,omitempty" validate:"js_function"`
	XShow       string `json:"xShow,omitempty" validate:"js_expression"`
	XIf         string `json:"xIf,omitempty" validate:"js_expression"`
	XModel      string `json:"xModel,omitempty" validate:"js_variable"`
	XBind       string `json:"xBind,omitempty" validate:"js_object"`
	XOn         string `json:"xOn,omitempty" validate:"js_object"`
	XText       string `json:"xText,omitempty" validate:"js_expression"`
	XHTML       string `json:"xHtml,omitempty" validate:"js_expression"`
	XRef        string `json:"xRef,omitempty" validate:"js_variable"`
	XCloak      bool   `json:"xCloak,omitempty"`
	XTransition string `json:"xTransition,omitempty"`
	XTeleport   string `json:"xTeleport,omitempty" validate:"css_selector"`
}

// Meta contains metadata and tracking information
type Meta struct {
	CreatedAt    time.Time              `json:"createdAt,omitempty"`
	UpdatedAt    time.Time              `json:"updatedAt,omitempty"`
	CreatedBy    string                 `json:"createdBy,omitempty" validate:"uuid"`
	UpdatedBy    string                 `json:"updatedBy,omitempty" validate:"uuid"`
	Deprecated   bool                   `json:"deprecated,omitempty"`
	Experimental bool                   `json:"experimental,omitempty"`
	Changelog    []ChangelogEntry       `json:"changelog,omitempty" validate:"dive"`
	CustomData   map[string]interface{} `json:"customData,omitempty"`
}

// State represents runtime state
type State struct {
	Values      map[string]interface{} `json:"values,omitempty"`
	Errors      map[string]string      `json:"errors,omitempty"`
	Touched     map[string]bool        `json:"touched,omitempty"`
	Dirty       map[string]bool        `json:"dirty,omitempty"`
	Valid       bool                   `json:"valid,omitempty"`
	Submitting  bool                   `json:"submitting,omitempty"`
	SubmitCount int                    `json:"submitCount,omitempty"`
	LastUpdated time.Time              `json:"lastUpdated,omitempty"`
	CurrentStep string                 `json:"currentStep,omitempty"`
	CurrentTab  string                 `json:"currentTab,omitempty"`
}

// Context provides runtime context information
type Context struct {
	UserID      string                 `json:"userId,omitempty" validate:"uuid"`
	TenantID    string                 `json:"tenantId,omitempty" validate:"uuid"`
	SessionID   string                 `json:"sessionId,omitempty" validate:"uuid"`
	RequestID   string                 `json:"requestId,omitempty" validate:"uuid"`
	IP          string                 `json:"ip,omitempty" validate:"ip"`
	UserAgent   string                 `json:"userAgent,omitempty"`
	Locale      string                 `json:"locale,omitempty" validate:"locale"`
	Timezone    string                 `json:"timezone,omitempty" validate:"timezone"`
	Permissions []string               `json:"permissions,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Environment string                 `json:"environment,omitempty" validate:"oneof=development staging production"`
	Debug       bool                   `json:"debug,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// Core interfaces for the schema system
type (
	// SchemaValidator validates schema structure and data
	SchemaValidator interface {
		ValidateSchema(ctx context.Context, schema *Schema) error
		ValidateData(ctx context.Context, schema *Schema, data map[string]interface{}) error
	}

	// SchemaRenderer renders schema to HTML/templates
	SchemaRenderer interface {
		Render(ctx context.Context, schema *Schema, data map[string]interface{}) (string, error)
		RenderField(ctx context.Context, field *Field, value interface{}) (string, error)
	}

	// SchemaRegistry manages schema storage and retrieval
	SchemaRegistry interface {
		Register(ctx context.Context, schema *Schema) error
		Get(ctx context.Context, id string) (*Schema, error)
		List(ctx context.Context, filter map[string]interface{}) ([]*Schema, error)
		Update(ctx context.Context, schema *Schema) error
		Delete(ctx context.Context, id string) error
	}

	// ConditionEvaluator evaluates conditional logic
	ConditionEvaluator interface {
		Evaluate(ctx context.Context, condition *condition.ConditionGroup, data map[string]interface{}) (bool, error)
	}

	// DataSourceResolver resolves dynamic data sources
	DataSourceResolver interface {
		Resolve(ctx context.Context, source *DataSource, params map[string]string) ([]Option, error)
	}
)

// Schema methods implement BaseSchema interface
func (s *Schema) GetID() string {
	return s.ID
}

func (s *Schema) GetType() string {
	return string(s.Type)
}

func (s *Schema) GetMetadata() *Meta {
	return s.Meta
}

func (s *Schema) Validate() error {
	// Basic validation - full implementation would use validator package
	if s.ID == "" {
		return ErrInvalidSchemaID
	}
	if s.Type == "" {
		return ErrInvalidSchemaType
	}
	if s.Title == "" {
		return ErrInvalidSchemaTitle
	}
	return nil
}

// Helper functions for common operations
func NewSchema(id string, schemaType Type, title string) *Schema {
	now := time.Now()
	return &Schema{
		ID:      id,
		Type:    schemaType,
		Version: "1.0.0",
		Title:   title,
		Meta: &Meta{
			CreatedAt: now,
			UpdatedAt: now,
		},
		State: &State{
			Values:      make(map[string]interface{}),
			Errors:      make(map[string]string),
			Touched:     make(map[string]bool),
			Dirty:       make(map[string]bool),
			Valid:       true,
			LastUpdated: now,
		},
	}
}

func (s *Schema) AddField(field Field) {
	if s.Fields == nil {
		s.Fields = []Field{}
	}
	s.Fields = append(s.Fields, field)
	s.updateTimestamp()
}

func (s *Schema) AddAction(action Action) {
	if s.Actions == nil {
		s.Actions = []Action{}
	}
	s.Actions = append(s.Actions, action)
	s.updateTimestamp()
}

func (s *Schema) updateTimestamp() {
	if s.Meta == nil {
		s.Meta = &Meta{}
	}
	s.Meta.UpdatedAt = time.Now()
}

// Validation helpers
func (f *Field) IsVisible(data map[string]interface{}) bool {
	if f.Hidden {
		return false
	}
	if f.Conditional != nil {
		// Placeholder for condition evaluation
		return true
	}
	return true
}

func (f *Field) IsRequired(data map[string]interface{}) bool {
	if f.Required {
		return true
	}
	if f.Conditional != nil {
		// Placeholder for conditional requirement evaluation
		return false
	}
	return false
}

// JSON marshaling with proper validation
func (s *Schema) MarshalJSON() ([]byte, error) {
	type Alias Schema
	s.updateTimestamp()
	return json.Marshal((*Alias)(s))
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	type Alias Schema
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return s.Validate()
}
