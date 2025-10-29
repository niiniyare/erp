// Package schema provides a JSON-driven UI schema system for building enterprise forms and components.
// Backend developers can define complete UIs by writing JSON files - no frontend code required.
//
// Example usage:
//
//	schema := schema.NewSchema("user-form", schema.TypeForm, "Create User")
//	schema.AddField(schema.Field{Name: "email", Type: schema.FieldEmail, Label: "Email"})
//	schema.AddAction(schema.Action{ID: "submit", Type: schema.ActionSubmit, Text: "Save"})
package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// Schema is the main entry point - defines a complete form, page, or UI component.
// All UI elements are configured through this structure via JSON.
type Schema struct {
	// Identity
	ID          string `json:"id" validate:"required,min=1,max=100" example:"user-create-form"`
	Type        Type   `json:"type" validate:"required" example:"form"`
	Version     string `json:"version,omitempty" validate:"semver" example:"1.2.0"`
	Title       string `json:"title" validate:"required,min=1,max=200" example:"Create User"`
	Description string `json:"description,omitempty" validate:"max=1000" example:"Add a new user to the system"`

	// Core UI structure
	Config  *Config  `json:"config,omitempty"`                  // HTTP/API configuration
	Layout  *Layout  `json:"layout,omitempty"`                  // Visual layout (grid, tabs, sections)
	Fields  []Field  `json:"fields,omitempty" validate:"dive"`  // Form fields and inputs
	Actions []Action `json:"actions,omitempty" validate:"dive"` // Buttons and actions

	// Enterprise features
	Security   *Security   `json:"security,omitempty"`   // CSRF, rate limiting, encryption
	Tenant     *Tenant     `json:"tenant,omitempty"`     // Multi-tenancy isolation
	Workflow   *Workflow   `json:"workflow,omitempty"`   // Approval workflows
	Validation *Validation `json:"validation,omitempty"` // Cross-field validation rules
	Events     *Events     `json:"events,omitempty"`     // Lifecycle event handlers
	I18n       *I18n       `json:"i18n,omitempty"`       // Internationalization

	// Frontend framework integration
	HTMX   *HTMX   `json:"htmx,omitempty"`   // HTMX configuration
	Alpine *Alpine `json:"alpine,omitempty"` // Alpine.js configuration

	// Organization and discovery
	Meta     *Meta    `json:"meta,omitempty"`                          // Creation/update metadata
	Tags     []string `json:"tags,omitempty" validate:"dive,alphanum"` // Search tags
	Category string   `json:"category,omitempty" validate:"alphanum"`  // UI category
	Module   string   `json:"module,omitempty" validate:"alphanum"`    // ERP module name

	// Runtime state (managed by system, not in JSON)
	State   *State   `json:"state,omitempty"`   // Form values, errors, dirty state
	Context *Context `json:"context,omitempty"` // Request context, user info, permissions
}

// Type defines what kind of UI element this schema represents
type Type string

const (
	TypeForm      Type = "form"      // Complete form with fields and submission
	TypeComponent Type = "component" // Reusable UI component
	TypeLayout    Type = "layout"    // Layout container only
	TypeWorkflow  Type = "workflow"  // Multi-step workflow
	TypeTheme     Type = "theme"     // Theme/styling configuration
	TypePage      Type = "page"      // Full page layout
)

// Config holds HTTP and API configuration for form submission
type Config struct {
	Action   string            `json:"action,omitempty" validate:"url" example:"/api/v1/users"`
	Method   string            `json:"method,omitempty" validate:"oneof=GET POST PUT PATCH DELETE" example:"POST"`
	Target   string            `json:"target,omitempty" validate:"html_id" example:"#main-content"` // HTMX target
	Encoding string            `json:"encoding,omitempty" validate:"oneof=application/json multipart/form-data application/x-www-form-urlencoded"`
	Timeout  int               `json:"timeout,omitempty" validate:"min=0,max=300000" example:"30000"` // milliseconds
	Headers  map[string]string `json:"headers,omitempty"`                                             // Custom request headers
	Params   map[string]string `json:"params,omitempty"`                                              // Query parameters
	Cache    bool              `json:"cache,omitempty"`                                               // Enable response caching
	CacheTTL int               `json:"cacheTTL,omitempty" validate:"min=0" example:"300"`             // seconds
}

// State represents runtime form state - managed by the frontend
type State struct {
	Values      map[string]any    `json:"values,omitempty"`      // Current field values
	Errors      map[string]string `json:"errors,omitempty"`      // Validation errors by field
	Touched     map[string]bool   `json:"touched,omitempty"`     // Fields user has interacted with
	Dirty       map[string]bool   `json:"dirty,omitempty"`       // Fields that changed from default
	Valid       bool              `json:"valid,omitempty"`       // Overall form validity
	Submitting  bool              `json:"submitting,omitempty"`  // Submission in progress
	SubmitCount int               `json:"submitCount,omitempty"` // Number of submission attempts
	LastUpdated time.Time         `json:"lastUpdated,omitempty"` // Last state change timestamp
	CurrentStep string            `json:"currentStep,omitempty"` // For multi-step forms
	CurrentTab  string            `json:"currentTab,omitempty"`  // For tabbed forms
}

// Context provides runtime execution context - injected by the system
type Context struct {
	// User and session
	UserID    string `json:"userId,omitempty" validate:"uuid"`
	TenantID  string `json:"tenantId,omitempty" validate:"uuid"`
	SessionID string `json:"sessionId,omitempty" validate:"uuid"`
	RequestID string `json:"requestId,omitempty" validate:"uuid"`

	// Request metadata
	IP        string `json:"ip,omitempty" validate:"ip"`
	UserAgent string `json:"userAgent,omitempty"`
	Locale    string `json:"locale,omitempty" validate:"locale"`
	Timezone  string `json:"timezone,omitempty" validate:"timezone"`

	// Authorization
	Permissions []string `json:"permissions,omitempty"` // User permissions
	Roles       []string `json:"roles,omitempty"`       // User roles

	// Environment
	Environment string         `json:"environment,omitempty" validate:"oneof=development staging production"`
	Debug       bool           `json:"debug,omitempty"` // Enable debug mode
	Data        map[string]any `json:"data,omitempty"`  // Custom context data
}

// Core interfaces - implement these to extend the schema system

// SchemaValidator validates schema structure and submitted data
type SchemaValidator interface {
	ValidateSchema(ctx context.Context, schema *Schema) error
	ValidateData(ctx context.Context, schema *Schema, data map[string]any) error
}

// SchemaRenderer converts schema to HTML/templates
type SchemaRenderer interface {
	Render(ctx context.Context, schema *Schema, data map[string]any) (string, error)
	RenderField(ctx context.Context, field *Field, value any) (string, error)
}

// SchemaRegistry manages schema storage and retrieval
type SchemaRegistry interface {
	Register(ctx context.Context, schema *Schema) error
	Get(ctx context.Context, id string) (*Schema, error)
	List(ctx context.Context, filter map[string]any) ([]*Schema, error)
	Update(ctx context.Context, schema *Schema) error
	Delete(ctx context.Context, id string) error
	GetVersion(ctx context.Context, id, version string) (*Schema, error) // Version support
}

// ConditionEvaluator evaluates conditional display/requirement logic
type ConditionEvaluator interface {
	Evaluate(ctx context.Context, condition *condition.ConditionGroup, data map[string]any) (bool, error)
	Compile(condition *condition.ConditionGroup) error // Pre-compile for performance
}

// DataSourceResolver resolves dynamic options for select fields
type DataSourceResolver interface {
	Resolve(ctx context.Context, source *DataSource, params map[string]string) ([]Option, error)
	InvalidateCache(ctx context.Context, source *DataSource) error // Clear cached options
}

// TransformProcessor handles data transformation
type TransformProcessor interface {
	Transform(ctx context.Context, transform *Transform, value any) (any, error)
}

// Constructor - creates a new schema with sensible defaults
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
			Values:      make(map[string]any),
			Errors:      make(map[string]string),
			Touched:     make(map[string]bool),
			Dirty:       make(map[string]bool),
			Valid:       true,
			LastUpdated: now,
		},
		Fields:  []Field{},
		Actions: []Action{},
	}
}

// AddField appends a field to the schema
func (s *Schema) AddField(field Field) {
	if s.Fields == nil {
		s.Fields = []Field{}
	}
	s.Fields = append(s.Fields, field)
	s.updateTimestamp()
}

// AddAction appends an action button to the schema
func (s *Schema) AddAction(action Action) {
	if s.Actions == nil {
		s.Actions = []Action{}
	}
	s.Actions = append(s.Actions, action)
	s.updateTimestamp()
}

// GetField retrieves a field by name
func (s *Schema) GetField(name string) (*Field, bool) {
	for i := range s.Fields {
		if s.Fields[i].Name == name {
			return &s.Fields[i], true
		}
	}
	return nil, false
}

// HasField checks if a field exists
func (s *Schema) HasField(name string) bool {
	_, exists := s.GetField(name)
	return exists
}

// GetVisibleFields returns fields visible for given data context
func (s *Schema) GetVisibleFields(data map[string]any) []Field {
	visible := make([]Field, 0, len(s.Fields))
	for _, field := range s.Fields {
		if field.IsVisible(data) {
			visible = append(visible, field)
		}
	}
	return visible
}

// GetRequiredFields returns required fields for given data context
func (s *Schema) GetRequiredFields(data map[string]any) []Field {
	required := make([]Field, 0)
	for _, field := range s.Fields {
		if field.IsRequired(data) && field.IsVisible(data) {
			required = append(required, field)
		}
	}
	return required
}

// DetectCircularDependencies checks for circular field dependencies
func (s *Schema) DetectCircularDependencies() error {
	visited := make(map[string]bool)
	stack := make(map[string]bool)

	for _, field := range s.Fields {
		if err := s.checkCycle(field.Name, visited, stack); err != nil {
			return err
		}
	}
	return nil
}

// checkCycle performs DFS to detect circular dependencies
func (s *Schema) checkCycle(fieldName string, visited, stack map[string]bool) error {
	if stack[fieldName] {
		return NewValidationError(
			"circular_dependency",
			"circular dependency detected involving field: "+fieldName,
		).WithField(fieldName)
	}
	if visited[fieldName] {
		return nil
	}

	visited[fieldName] = true
	stack[fieldName] = true

	field, exists := s.GetField(fieldName)
	if exists {
		for _, dep := range field.Dependencies {
			if err := s.checkCycle(dep, visited, stack); err != nil {
				return err
			}
		}
	}

	stack[fieldName] = false
	return nil
}

// Validate performs comprehensive schema validation
func (s *Schema) Validate() error {
	collector := NewErrorCollector()

	// Basic field validation
	if s.ID == "" {
		collector.AddError(ErrInvalidSchemaID)
	}
	if s.Type == "" {
		collector.AddError(ErrInvalidSchemaType)
	}
	if s.Title == "" {
		collector.AddError(ErrInvalidSchemaTitle)
	}

	// Validate all fields
	for i, field := range s.Fields {
		if err := field.Validate(context.Background()); err != nil {
			if schemaErr, ok := err.(SchemaError); ok {
				collector.AddFieldError(fmt.Sprintf("Fields[%d]", i), schemaErr)
			} else {
				collector.AddValidationError(
					fmt.Sprintf("Fields[%d]", i),
					"field_validation",
					err.Error(),
				)
			}
		}
	}

	// Validate all actions
	for i, action := range s.Actions {
		if err := action.Validate(context.Background()); err != nil {
			if schemaErr, ok := err.(SchemaError); ok {
				collector.AddFieldError(fmt.Sprintf("Actions[%d]", i), schemaErr)
			} else {
				collector.AddValidationError(
					fmt.Sprintf("Actions[%d]", i),
					"action_validation",
					err.Error(),
				)
			}
		}
	}

	// Check for duplicate field names
	seen := make(map[string]bool)
	for _, field := range s.Fields {
		if seen[field.Name] {
			collector.AddFieldError(
				field.Name,
				ErrFieldDuplicate.WithDetail("name", field.Name),
			)
		}
		seen[field.Name] = true
	}

	// Validate field dependencies exist
	for _, field := range s.Fields {
		for _, dep := range field.Dependencies {
			if !s.HasField(dep) {
				collector.AddValidationError(
					field.Name,
					"invalid_dependency",
					fmt.Sprintf("depends on non-existent field: %s", dep),
				)
			}
		}
	}

	// Check for circular dependencies
	if err := s.DetectCircularDependencies(); err != nil {
		if schemaErr, ok := err.(SchemaError); ok {
			collector.AddError(schemaErr)
		} else {
			collector.AddValidationError("", "circular_dependency", err.Error())
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// Clone creates a deep copy of the schema
func (s *Schema) Clone() *Schema {
	data, _ := json.Marshal(s)
	var clone Schema
	json.Unmarshal(data, &clone)
	return &clone
}

// updateTimestamp updates the schema's last modified timestamp
func (s *Schema) updateTimestamp() {
	if s.Meta == nil {
		s.Meta = &Meta{}
	}
	s.Meta.UpdatedAt = time.Now()
}

// MarshalJSON implements custom JSON marshaling
func (s *Schema) MarshalJSON() ([]byte, error) {
	type Alias Schema
	s.updateTimestamp()
	return json.Marshal((*Alias)(s))
}

// UnmarshalJSON implements custom JSON unmarshaling with validation
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

// Interface implementation checks
func (s *Schema) GetID() string      { return s.ID }
func (s *Schema) GetType() string    { return string(s.Type) }
func (s *Schema) GetMetadata() *Meta { return s.Meta }
func (s *Schema) GetVersion() string { return s.Version }
