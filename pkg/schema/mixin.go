package schema

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// Mixin represents a reusable collection of fields and validation rules
// that can be included in multiple schemas to avoid duplication
type Mixin struct {
	// Identity
	ID          string `json:"id" validate:"required,min=1,max=100" example:"audit_fields"`
	Name        string `json:"name" validate:"required,min=1,max=200" example:"Audit Fields"`
	Description string `json:"description,omitempty" validate:"max=500" example:"Standard audit fields for tracking changes"`
	Version     string `json:"version,omitempty" validate:"semver" example:"1.0.0"`

	// Content
	Fields   []Field   `json:"fields" validate:"dive"`            // Fields to include
	Actions  []Action  `json:"actions,omitempty" validate:"dive"` // Actions to include
	Events   *Events   `json:"events,omitempty"`                  // Event handlers
	Security *Security `json:"security,omitempty"`                // Security settings

	// Organization
	Category string         `json:"category,omitempty" validate:"alphanum"`  // Category for organization
	Tags     []string       `json:"tags,omitempty" validate:"dive,alphanum"` // Search tags
	Metadata map[string]any `json:"metadata,omitempty"`                      // Custom metadata

	// Tracking
	Meta *Meta `json:"meta,omitempty"` // Creation/update metadata
}

// MixinRegistry manages available mixins with thread-safe operations
type MixinRegistry struct {
	mixins map[string]*Mixin
	mu     sync.RWMutex
}

// NewMixinRegistry creates a new mixin registry with built-in mixins
func NewMixinRegistry() *MixinRegistry {
	registry := &MixinRegistry{
		mixins: make(map[string]*Mixin),
	}

	// Register built-in mixins
	registry.registerBuiltInMixins()

	return registry
}

// Register adds a mixin to the registry
func (r *MixinRegistry) Register(mixin *Mixin) error {
	return r.register(mixin, true)
}

// registerBuiltIn adds a mixin without validation (for built-in mixins)
func (r *MixinRegistry) registerBuiltIn(mixin *Mixin) error {
	return r.register(mixin, false)
}

// register is the internal registration method with thread safety
func (r *MixinRegistry) register(mixin *Mixin, validate bool) error {
	if mixin.ID == "" {
		return NewValidationError("mixin_id", "mixin ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate registration
	if _, exists := r.mixins[mixin.ID]; exists {
		return NewValidationError("mixin_duplicate", fmt.Sprintf("mixin %s already exists", mixin.ID))
	}

	// Only validate user-defined mixins
	if validate {
		if err := mixin.Validate(); err != nil {
			return fmt.Errorf("invalid mixin %s: %w", mixin.ID, err)
		}
	}

	r.mixins[mixin.ID] = mixin
	return nil
}

// Get retrieves a mixin by ID with thread safety
func (r *MixinRegistry) Get(id string) (*Mixin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mixin, exists := r.mixins[id]
	if !exists {
		return nil, NewValidationError("mixin_not_found", "mixin not found: "+id)
	}
	return mixin, nil
}

// List returns all registered mixins with thread safety
func (r *MixinRegistry) List() map[string]*Mixin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*Mixin)
	maps.Copy(result, r.mixins)
	return result
}

// ApplyMixin applies a mixin to a schema (backward compatible)
func (r *MixinRegistry) ApplyMixin(schema *Schema, mixinID string, prefix string) error {
	return r.ApplyMixinWithEvaluator(schema, mixinID, prefix, nil)
}

// ApplyMixinWithEvaluator applies a mixin to a schema with evaluator injection
func (r *MixinRegistry) ApplyMixinWithEvaluator(schema *Schema, mixinID string, prefix string, evaluator *condition.Evaluator) error {
	mixin, err := r.Get(mixinID)
	if err != nil {
		return err
	}

	// Apply fields with optional prefix
	for _, field := range mixin.Fields {
		newField := field // Copy field
		if prefix != "" {
			newField.Name = prefix + "_" + field.Name
		}

		// Set evaluator on field if provided
		if evaluator != nil {
			newField.SetEvaluator(evaluator)
		}

		// Check for conflicts
		if schema.HasField(newField.Name) {
			return NewValidationError(
				"field_conflict",
				fmt.Sprintf("field %s already exists in schema", newField.Name),
			)
		}

		schema.AddField(newField)
	}

	// Apply actions
	for _, action := range mixin.Actions {
		newAction := action // Copy action
		if prefix != "" {
			newAction.ID = prefix + "_" + action.ID
		}

		// Check for conflicts
		for _, existingAction := range schema.Actions {
			if existingAction.ID == newAction.ID {
				return NewValidationError(
					"action_conflict",
					fmt.Sprintf("action %s already exists in schema", newAction.ID),
				)
			}
		}

		schema.AddAction(newAction)
	}

	// Track applied mixin
	if schema.Meta == nil {
		schema.Meta = &Meta{}
	}
	if schema.Meta.CustomData == nil {
		schema.Meta.CustomData = make(map[string]any)
	}

	appliedMixins, exists := schema.Meta.CustomData["applied_mixins"]
	if !exists {
		appliedMixins = []string{}
	}

	mixinList, ok := appliedMixins.([]string)
	mixinRef := mixinID
	if !ok {
		mixinList = []string{}
	}
	schema.Meta.CustomData["applied_mixins"] = append(mixinList, mixinRef)

	return nil
}

// Unregister removes a mixin from the registry
func (r *MixinRegistry) Unregister(mixinID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.mixins[mixinID]; !exists {
		return NewValidationError("mixin_not_found", fmt.Sprintf("mixin %s not found", mixinID))
	}

	delete(r.mixins, mixinID)
	return nil
}

// Update updates an existing mixin
func (r *MixinRegistry) Update(mixin *Mixin) error {
	if mixin == nil || mixin.ID == "" {
		return NewValidationError("mixin_invalid", "mixin is required with valid ID")
	}

	if err := mixin.Validate(); err != nil {
		return fmt.Errorf("invalid mixin %s: %w", mixin.ID, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.mixins[mixin.ID]; !exists {
		return NewValidationError("mixin_not_found", fmt.Sprintf("mixin %s not found", mixin.ID))
	}

	r.mixins[mixin.ID] = mixin
	return nil
}

// Validate checks if mixin configuration is valid
func (m *Mixin) Validate() error {
	collector := NewErrorCollector()

	if m.ID == "" {
		collector.AddError(NewValidationError("mixin_id", "mixin ID is required"))
	}

	if m.Name == "" {
		collector.AddError(NewValidationError("mixin_name", "mixin name is required"))
	}

	// Validate fields
	fieldNames := make(map[string]bool)
	for i, field := range m.Fields {
		if err := field.Validate(context.Background()); err != nil {
			collector.AddFieldError(fmt.Sprintf("Fields[%d]", i), err.(SchemaError))
		}

		// Check for duplicate field names within mixin
		if fieldNames[field.Name] {
			collector.AddValidationError(
				field.Name,
				"duplicate_field",
				"duplicate field name in mixin: "+field.Name,
			)
		}
		fieldNames[field.Name] = true
	}

	// Validate actions
	for i, action := range m.Actions {
		if err := action.Validate(context.Background()); err != nil {
			collector.AddFieldError(fmt.Sprintf("Actions[%d]", i), err.(SchemaError))
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// Clone creates a deep copy of the mixin
func (m *Mixin) Clone() *Mixin {
	return &Mixin{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Version:     m.Version,
		Fields:      append([]Field{}, m.Fields...),
		Actions:     append([]Action{}, m.Actions...),
		Category:    m.Category,
		Tags:        append([]string{}, m.Tags...),
		Metadata:    copyMap(m.Metadata),
	}
}

// registerBuiltInMixins registers common mixins
func (r *MixinRegistry) registerBuiltInMixins() {
	// Audit Fields Mixin
	auditMixin := &Mixin{
		ID:          "audit_fields",
		Name:        "Audit Fields",
		Description: "Standard audit fields for tracking record changes",
		Version:     "1.0.0",
		Category:    "system",
		Tags:        []string{"audit", "tracking", "system"},
		Fields: []Field{
			{
				Name:        "created_at",
				Type:        FieldDateTime,
				Label:       "Created At",
				Description: "When the record was created",
				Readonly:    true,
				Default:     "NOW()",
			},
			{
				Name:        "updated_at",
				Type:        FieldDateTime,
				Label:       "Updated At",
				Description: "When the record was last updated",
				Readonly:    true,
				Default:     "NOW()",
			},
			{
				Name:        "created_by",
				Type:        FieldText,
				Label:       "Created By",
				Description: "User who created the record",
				Readonly:    true,
			},
			{
				Name:        "updated_by",
				Type:        FieldText,
				Label:       "Updated By",
				Description: "User who last updated the record",
				Readonly:    true,
			},
		},
		Meta: &Meta{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	r.registerBuiltIn(auditMixin)

	// Address Fields Mixin
	addressMixin := &Mixin{
		ID:          "address_fields",
		Name:        "Address Fields",
		Description: "Standard address fields for locations",
		Version:     "1.0.0",
		Category:    "location",
		Tags:        []string{"address", "location", "contact"},
		Fields: []Field{
			{
				Name:        "street_address",
				Type:        FieldText,
				Label:       "Street Address",
				Description: "Street address line 1",
				Required:    true,
				Validation: &FieldValidation{
					MaxLength: intPtr(200),
				},
			},
			{
				Name:        "street_address_2",
				Type:        FieldText,
				Label:       "Street Address 2",
				Description: "Street address line 2 (optional)",
				Validation: &FieldValidation{
					MaxLength: intPtr(200),
				},
			},
			{
				Name:        "city",
				Type:        FieldText,
				Label:       "City",
				Description: "City or locality",
				Required:    true,
				Validation: &FieldValidation{
					MaxLength: intPtr(100),
				},
			},
			{
				Name:        "state_province",
				Type:        FieldText,
				Label:       "State/Province",
				Description: "State, province, or region",
				Validation: &FieldValidation{
					MaxLength: intPtr(100),
				},
			},
			{
				Name:        "postal_code",
				Type:        FieldText,
				Label:       "Postal Code",
				Description: "ZIP or postal code",
				Validation: &FieldValidation{
					MaxLength: intPtr(20),
				},
			},
			{
				Name:        "country",
				Type:        FieldSelect,
				Label:       "Country",
				Description: "Country",
				Required:    true,
				Default:     "US",
			},
		},
		Meta: &Meta{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	r.registerBuiltIn(addressMixin)

	// Contact Fields Mixin
	contactMixin := &Mixin{
		ID:          "contact_fields",
		Name:        "Contact Fields",
		Description: "Standard contact information fields",
		Version:     "1.0.0",
		Category:    "contact",
		Tags:        []string{"contact", "communication"},
		Fields: []Field{
			{
				Name:        "phone",
				Type:        FieldPhone,
				Label:       "Phone",
				Description: "Primary phone number",
				Validation: &FieldValidation{
					Format: "phone",
				},
			},
			{
				Name:        "email",
				Type:        FieldEmail,
				Label:       "Email",
				Description: "Email address",
				Validation: &FieldValidation{
					Format: "email",
				},
			},
			{
				Name:        "fax",
				Type:        FieldPhone,
				Label:       "Fax",
				Description: "Fax number",
				Validation: &FieldValidation{
					Format: "phone",
				},
			},
			{
				Name:        "website",
				Type:        FieldURL,
				Label:       "Website",
				Description: "Website URL",
				Validation: &FieldValidation{
					Format: "url",
				},
			},
		},
		Meta: &Meta{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	r.registerBuiltIn(contactMixin)

	// Status Fields Mixin
	statusMixin := &Mixin{
		ID:          "status_fields",
		Name:        "Status Fields",
		Description: "Standard status and lifecycle fields",
		Version:     "1.0.0",
		Category:    "system",
		Tags:        []string{"status", "lifecycle", "system"},
		Fields: []Field{
			{
				Name:        "status",
				Type:        FieldSelect,
				Label:       "Status",
				Description: "Current status of the record",
				Required:    true,
				Default:     "active",
				Options: []Option{
					{Value: "active", Label: "Active"},
					{Value: "inactive", Label: "Inactive"},
					{Value: "pending", Label: "Pending"},
					{Value: "archived", Label: "Archived"},
				},
			},
			{
				Name:        "is_active",
				Type:        FieldCheckbox,
				Label:       "Is Active",
				Description: "Whether the record is active",
				Default:     true,
			},
			{
				Name:        "is_deleted",
				Type:        FieldCheckbox,
				Label:       "Is Deleted",
				Description: "Soft delete flag",
				Default:     false,
				Hidden:      true,
			},
		},
		Meta: &Meta{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	r.registerBuiltIn(statusMixin)
}

// Helper functions

func copyMap(original map[string]any) map[string]any {
	if original == nil {
		return nil
	}
	copy := make(map[string]any)
	maps.Copy(copy, original)
	return copy
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
