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

	// Internal fields (not serialized)
	evaluator *condition.Evaluator `json:"-"` // Condition evaluator for dynamic behavior
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

// SetEvaluator sets the condition evaluator for the schema and all its fields
func (s *Schema) SetEvaluator(evaluator *condition.Evaluator) {
	s.evaluator = evaluator

	// Set evaluator on all fields
	for i := range s.Fields {
		s.Fields[i].SetEvaluator(evaluator)
	}
}

// GetEvaluator returns the schema's condition evaluator
func (s *Schema) GetEvaluator() *condition.Evaluator {
	return s.evaluator
}

// AddField appends a field to the schema
func (s *Schema) AddField(field Field) {
	if s.Fields == nil {
		s.Fields = []Field{}
	}

	// Set evaluator on new field if schema has one
	if s.evaluator != nil {
		field.SetEvaluator(s.evaluator)
	}

	s.Fields = append(s.Fields, field)
	s.updateTimestamp()
}

// AddFields appends multiple fields to the schema
func (s *Schema) AddFields(fields ...Field) {
	for _, field := range fields {
		s.AddField(field)
	}
}

// RemoveField removes a field by name
func (s *Schema) RemoveField(name string) bool {
	for i, field := range s.Fields {
		if field.Name == name {
			s.Fields = append(s.Fields[:i], s.Fields[i+1:]...)
			s.updateTimestamp()
			return true
		}
	}
	return false
}

// UpdateField updates an existing field by name
func (s *Schema) UpdateField(name string, updatedField Field) bool {
	for i, field := range s.Fields {
		if field.Name == name {
			// Preserve evaluator
			if s.evaluator != nil {
				updatedField.SetEvaluator(s.evaluator)
			}
			s.Fields[i] = updatedField
			s.updateTimestamp()
			return true
		}
	}
	return false
}

// AddAction appends an action button to the schema
func (s *Schema) AddAction(action Action) {
	if s.Actions == nil {
		s.Actions = []Action{}
	}
	s.Actions = append(s.Actions, action)
	s.updateTimestamp()
}

// AddActions appends multiple actions to the schema
func (s *Schema) AddActions(actions ...Action) {
	for _, action := range actions {
		s.AddAction(action)
	}
}

// RemoveAction removes an action by ID
func (s *Schema) RemoveAction(id string) bool {
	for i, action := range s.Actions {
		if action.ID == id {
			s.Actions = append(s.Actions[:i], s.Actions[i+1:]...)
			s.updateTimestamp()
			return true
		}
	}
	return false
}

// UpdateAction updates an existing action by ID
func (s *Schema) UpdateAction(id string, updatedAction Action) bool {
	for i, action := range s.Actions {
		if action.ID == id {
			s.Actions[i] = updatedAction
			s.updateTimestamp()
			return true
		}
	}
	return false
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

// GetFieldByIndex retrieves a field by index
func (s *Schema) GetFieldByIndex(index int) (*Field, bool) {
	if index >= 0 && index < len(s.Fields) {
		return &s.Fields[index], true
	}
	return nil, false
}

// GetAction retrieves an action by ID
func (s *Schema) GetAction(id string) (*Action, bool) {
	for i := range s.Actions {
		if s.Actions[i].ID == id {
			return &s.Actions[i], true
		}
	}
	return nil, false
}

// HasField checks if a field exists
func (s *Schema) HasField(name string) bool {
	_, exists := s.GetField(name)
	return exists
}

// HasAction checks if an action exists
func (s *Schema) HasAction(id string) bool {
	_, exists := s.GetAction(id)
	return exists
}

// GetFieldNames returns all field names
func (s *Schema) GetFieldNames() []string {
	names := make([]string, len(s.Fields))
	for i, field := range s.Fields {
		names[i] = field.Name
	}
	return names
}

// GetActionIDs returns all action IDs
func (s *Schema) GetActionIDs() []string {
	ids := make([]string, len(s.Actions))
	for i, action := range s.Actions {
		ids[i] = action.ID
	}
	return ids
}

// GetVisibleFields returns fields visible for given data context
func (s *Schema) GetVisibleFields(ctx context.Context, data map[string]any) []*Field {
	visible := make([]*Field, 0, len(s.Fields))
	for i := range s.Fields {
		if ok, _ := s.Fields[i].IsVisible(ctx, data); ok {
			visible = append(visible, &s.Fields[i])
		}
	}
	return visible
}

// GetHiddenFields returns fields hidden for given data context
func (s *Schema) GetHiddenFields(ctx context.Context, data map[string]any) []*Field {
	hidden := make([]*Field, 0, len(s.Fields))
	for i := range s.Fields {
		if ok, _ := s.Fields[i].IsVisible(ctx, data); !ok {
			hidden = append(hidden, &s.Fields[i])
		}
	}
	return hidden
}

// GetRequiredFields returns required fields for given data context
func (s *Schema) GetRequiredFields(ctx context.Context, data map[string]any) []*Field {
	required := make([]*Field, 0, len(s.Fields))
	for i := range s.Fields {
		field := &s.Fields[i]

		// Check if field is required (ignore errors, treat as not required)
		isRequired, err := field.IsRequired(ctx, data)
		if err != nil {
			continue
		}

		if isRequired {
			if ok, _ := field.IsVisible(ctx, data); ok {
				required = append(required, field)
			}
		}
	}
	return required
}

// GetOptionalFields returns optional fields for given data context
func (s *Schema) GetOptionalFields(ctx context.Context, data map[string]any) []*Field {
	optional := make([]*Field, 0, len(s.Fields))
	for i := range s.Fields {
		field := &s.Fields[i]

		isRequired, err := field.IsRequired(ctx, data)
		if err != nil || isRequired {
			continue
		}

		if ok, _ := field.IsVisible(ctx, data); ok {
			optional = append(optional, field)
		}
	}
	return optional
}

// GetFieldsByType returns all fields of a specific type
func (s *Schema) GetFieldsByType(fieldType FieldType) []*Field {
	fields := make([]*Field, 0)
	for i := range s.Fields {
		if s.Fields[i].Type == fieldType {
			fields = append(fields, &s.Fields[i])
		}
	}
	return fields
}

// GetEnabledActions returns all enabled (non-disabled) actions
func (s *Schema) GetEnabledActions(ctx context.Context, data map[string]any) []Action {
	enabled := make([]Action, 0, len(s.Actions))
	for _, action := range s.Actions {
		if !action.Disabled && !action.Hidden {
			// Check conditions if present
			if action.Condition != nil && s.evaluator != nil {
				evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
				result, err := s.evaluator.Evaluate(ctx, action.Condition, evalCtx)
				if err != nil || !result {
					continue
				}
			}
			enabled = append(enabled, action)
		}
	}
	return enabled
}

// ValidateData validates submitted data against the schema
func (s *Schema) ValidateData(ctx context.Context, data map[string]any) error {
	collector := NewErrorCollector()

	// Get visible required fields
	requiredFields := s.GetRequiredFields(ctx, data)

	// Check required fields
	for _, field := range requiredFields {
		value, exists := data[field.Name]
		if !exists || isEmpty(value) {
			collector.AddValidationError(
				field.Name,
				"required",
				fmt.Sprintf("%s is required", field.Label),
			)
		}
	}

	// Validate all provided field values
	for _, field := range s.Fields {
		value, exists := data[field.Name]
		if exists {
			if err := field.ValidateValue(ctx, value); err != nil {
				if schemaErr, ok := err.(SchemaError); ok {
					collector.AddFieldError(field.Name, schemaErr)
				} else {
					collector.AddValidationError(field.Name, "validation_failed", err.Error())
				}
			}
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
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

	// Check for duplicate action IDs
	seenActions := make(map[string]bool)
	for _, action := range s.Actions {
		if seenActions[action.ID] {
			collector.AddValidationError(
				action.ID,
				"duplicate_action",
				fmt.Sprintf("duplicate action ID: %s", action.ID),
			)
		}
		seenActions[action.ID] = true
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

// Clone creates a deep copy of the schema with evaluator preservation
func (s *Schema) Clone() *Schema {
	// Marshal to JSON and back for deep copy
	data, _ := json.Marshal(s)
	var clone Schema
	json.Unmarshal(data, &clone)

	// Preserve evaluator (doesn't serialize)
	clone.evaluator = s.evaluator

	// Set evaluator on all fields
	if clone.evaluator != nil {
		for i := range clone.Fields {
			clone.Fields[i].SetEvaluator(clone.evaluator)
		}
	}

	return &clone
}

// Reset resets the schema state
func (s *Schema) Reset() {
	if s.State == nil {
		s.State = &State{}
	}

	s.State.Values = make(map[string]any)
	s.State.Errors = make(map[string]string)
	s.State.Touched = make(map[string]bool)
	s.State.Dirty = make(map[string]bool)
	s.State.Valid = true
	s.State.Submitting = false
	s.State.SubmitCount = 0
	s.State.LastUpdated = time.Now()
}

// SetFieldValue sets a value in the schema state
func (s *Schema) SetFieldValue(fieldName string, value any) {
	if s.State == nil {
		s.State = &State{
			Values: make(map[string]any),
		}
	}
	if s.State.Values == nil {
		s.State.Values = make(map[string]any)
	}
	s.State.Values[fieldName] = value
	s.State.LastUpdated = time.Now()
}

// GetFieldValue gets a value from the schema state
func (s *Schema) GetFieldValue(fieldName string) (any, bool) {
	if s.State == nil || s.State.Values == nil {
		return nil, false
	}
	value, exists := s.State.Values[fieldName]
	return value, exists
}

// SetFieldError sets an error for a field
func (s *Schema) SetFieldError(fieldName string, errorMsg string) {
	if s.State == nil {
		s.State = &State{
			Errors: make(map[string]string),
		}
	}
	if s.State.Errors == nil {
		s.State.Errors = make(map[string]string)
	}
	s.State.Errors[fieldName] = errorMsg
	s.State.Valid = false
}

// ClearFieldError clears an error for a field
func (s *Schema) ClearFieldError(fieldName string) {
	if s.State != nil && s.State.Errors != nil {
		delete(s.State.Errors, fieldName)
		// Check if there are any remaining errors
		s.State.Valid = len(s.State.Errors) == 0
	}
}

// ClearAllErrors clears all field errors
func (s *Schema) ClearAllErrors() {
	if s.State != nil {
		s.State.Errors = make(map[string]string)
		s.State.Valid = true
	}
}

// HasErrors checks if the schema has any validation errors
func (s *Schema) HasErrors() bool {
	return s.State != nil && s.State.Errors != nil && len(s.State.Errors) > 0
}

// GetErrors returns all validation errors
func (s *Schema) GetErrors() map[string]string {
	if s.State == nil || s.State.Errors == nil {
		return make(map[string]string)
	}
	return s.State.Errors
}

// MarkFieldTouched marks a field as touched
func (s *Schema) MarkFieldTouched(fieldName string) {
	if s.State == nil {
		s.State = &State{
			Touched: make(map[string]bool),
		}
	}
	if s.State.Touched == nil {
		s.State.Touched = make(map[string]bool)
	}
	s.State.Touched[fieldName] = true
}

// IsFieldTouched checks if a field has been touched
func (s *Schema) IsFieldTouched(fieldName string) bool {
	return s.State != nil && s.State.Touched != nil && s.State.Touched[fieldName]
}

// MarkFieldDirty marks a field as dirty (changed from default)
func (s *Schema) MarkFieldDirty(fieldName string) {
	if s.State == nil {
		s.State = &State{
			Dirty: make(map[string]bool),
		}
	}
	if s.State.Dirty == nil {
		s.State.Dirty = make(map[string]bool)
	}
	s.State.Dirty[fieldName] = true
}

// IsFieldDirty checks if a field is dirty
func (s *Schema) IsFieldDirty(fieldName string) bool {
	return s.State != nil && s.State.Dirty != nil && s.State.Dirty[fieldName]
}

// IsDirty checks if any field is dirty
func (s *Schema) IsDirty() bool {
	if s.State == nil || s.State.Dirty == nil {
		return false
	}
	for _, dirty := range s.State.Dirty {
		if dirty {
			return true
		}
	}
	return false
}

// GetFieldCount returns the number of fields in the schema
func (s *Schema) GetFieldCount() int {
	return len(s.Fields)
}

// GetActionCount returns the number of actions in the schema
func (s *Schema) GetActionCount() int {
	return len(s.Actions)
}

// IsEmpty checks if the schema has no fields or actions
func (s *Schema) IsEmpty() bool {
	return len(s.Fields) == 0 && len(s.Actions) == 0
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

// ToJSON converts the schema to JSON string
func (s *Schema) ToJSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", WrapError(err, "json_marshal_failed", "failed to marshal schema to JSON")
	}
	return string(data), nil
}

// ToJSONPretty converts the schema to pretty-printed JSON string
func (s *Schema) ToJSONPretty() (string, error) {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", WrapError(err, "json_marshal_failed", "failed to marshal schema to JSON")
	}
	return string(data), nil
}

// FromJSON creates a schema from JSON string
func FromJSON(jsonStr string) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal([]byte(jsonStr), &schema); err != nil {
		return nil, WrapError(err, "json_unmarshal_failed", "failed to unmarshal schema from JSON")
	}
	return &schema, nil
}

// Interface implementation checks

func (s *Schema) GetID() string      { return s.ID }
func (s *Schema) GetType() string    { return string(s.Type) }
func (s *Schema) GetMetadata() *Meta { return s.Meta }
func (s *Schema) GetVersion() string { return s.Version }

// Utility functions

// isEmpty checks if a value is considered empty
//
//	func isEmpty(value any) bool {
//		if value == nil {
//			return true
//		}
//
//		switch v := value.(type) {
//		case string:
//			return v == ""
//		case []any:
//			return len(v) == 0
//		case map[string]any:
//			return len(v) == 0
//		case bool:
//			return false // boolean false is not considered empty
//		case int, int8, int16, int32, int64:
//			return false // zero numbers are not considered empty
//		case float32, float64:
//			return false // zero numbers are not considered empty
//		default:
//			return false
//		}
//	}
//
// MergeSchemas merges multiple schemas into one (experimental)
func MergeSchemas(base *Schema, others ...*Schema) *Schema {
	merged := base.Clone()

	for _, other := range others {
		// Merge fields (skip duplicates)
		for _, field := range other.Fields {
			if !merged.HasField(field.Name) {
				merged.AddField(field)
			}
		}

		// Merge actions (skip duplicates)
		for _, action := range other.Actions {
			if !merged.HasAction(action.ID) {
				merged.AddAction(action)
			}
		}

		// Merge tags
		if len(other.Tags) > 0 {
			tagSet := make(map[string]bool)
			for _, tag := range merged.Tags {
				tagSet[tag] = true
			}
			for _, tag := range other.Tags {
				if !tagSet[tag] {
					merged.Tags = append(merged.Tags, tag)
				}
			}
		}
	}

	merged.updateTimestamp()
	return merged
}

// FilterFields returns a new schema with only fields matching the predicate
func (s *Schema) FilterFields(predicate func(field Field) bool) *Schema {
	filtered := s.Clone()
	filtered.Fields = []Field{}

	for _, field := range s.Fields {
		if predicate(field) {
			filtered.AddField(field)
		}
	}

	return filtered
}

// MapFields applies a transformation to all fields
func (s *Schema) MapFields(transform func(field Field) Field) *Schema {
	mapped := s.Clone()

	for i, field := range mapped.Fields {
		mapped.Fields[i] = transform(field)
	}

	mapped.updateTimestamp()
	return mapped
}
