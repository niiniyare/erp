package enrich

import (
	"context"
	"fmt"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// User represents a user in the system for enrichment purposes
type User interface {
	GetID() string
	GetTenantID() string
	GetPermissions() []string
	GetRoles() []string
	HasPermission(permission string) bool
	HasRole(role string) bool
}

// TenantOverride represents tenant-specific customizations
type TenantOverride struct {
	FieldName   string `json:"field_name"`
	Label       string `json:"label,omitempty"`
	Required    *bool  `json:"required,omitempty"`
	Hidden      *bool  `json:"hidden,omitempty"`
	DefaultValue any   `json:"default_value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Help        string `json:"help,omitempty"`
}

// TenantCustomization holds all tenant-specific overrides for a schema
type TenantCustomization struct {
	SchemaID   string                    `json:"schema_id"`
	TenantID   string                    `json:"tenant_id"`
	Overrides  map[string]TenantOverride `json:"overrides"`
	CreatedAt  time.Time                 `json:"created_at"`
	UpdatedAt  time.Time                 `json:"updated_at"`
}

// TenantProvider interface for retrieving tenant customizations
type TenantProvider interface {
	GetCustomization(ctx context.Context, schemaID, tenantID string) (*TenantCustomization, error)
}

// Enricher interface defines the contract for schema enrichment
type Enricher interface {
	// Enrich enriches a schema with runtime data based on user context
	Enrich(ctx context.Context, schema *schema.Schema, user User) (*schema.Schema, error)
	
	// EnrichField enriches a single field with runtime data
	EnrichField(ctx context.Context, field *schema.Field, user User, data map[string]any) error
	
	// SetTenantProvider sets the tenant customization provider
	SetTenantProvider(provider TenantProvider)
}

// DefaultEnricher is the default implementation of the Enricher interface
type DefaultEnricher struct {
	tenantProvider TenantProvider
}

// NewEnricher creates a new DefaultEnricher instance
func NewEnricher() *DefaultEnricher {
	return &DefaultEnricher{}
}

// SetTenantProvider sets the tenant customization provider
func (e *DefaultEnricher) SetTenantProvider(provider TenantProvider) {
	e.tenantProvider = provider
}

// Enrich enriches a schema with runtime data based on user context
func (e *DefaultEnricher) Enrich(ctx context.Context, schemaObj *schema.Schema, user User) (*schema.Schema, error) {
	if schemaObj == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}
	
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	// Clone the schema to avoid modifying the original
	enriched := schemaObj.Clone()
	
	// Get tenant customizations if provider is available
	var customization *TenantCustomization
	if e.tenantProvider != nil {
		var err error
		customization, err = e.tenantProvider.GetCustomization(ctx, enriched.ID, user.GetTenantID())
		if err != nil {
			// Log error but continue with enrichment
			// In production, you might want to use a proper logger
			fmt.Printf("Warning: failed to get tenant customization: %v\n", err)
		}
	}
	
	// Create data context for dynamic defaults
	dataContext := map[string]any{
		"user_id":   user.GetID(),
		"tenant_id": user.GetTenantID(),
		"timestamp": time.Now(),
	}
	
	// Enrich each field
	for i := range enriched.Fields {
		field := &enriched.Fields[i]
		
		// Apply tenant customizations
		if customization != nil {
			e.applyTenantCustomization(field, customization)
		}
		
		// Enrich the field
		if err := e.EnrichField(ctx, field, user, dataContext); err != nil {
			return nil, fmt.Errorf("failed to enrich field %s: %w", field.Name, err)
		}
	}
	
	// Enrich actions
	for i := range enriched.Actions {
		action := &enriched.Actions[i]
		e.enrichAction(ctx, action, user)
	}
	
	return enriched, nil
}

// EnrichField enriches a single field with runtime data
func (e *DefaultEnricher) EnrichField(ctx context.Context, field *schema.Field, user User, data map[string]any) error {
	if field == nil {
		return fmt.Errorf("field cannot be nil")
	}
	
	// Initialize runtime if not exists
	if field.Runtime == nil {
		field.Runtime = &schema.FieldRuntime{}
	}
	
	// Apply permission-based visibility and editability
	e.applyPermissions(field, user)
	
	// Apply dynamic default values
	e.applyDynamicDefaults(field, data)
	
	// Set tenant isolation field
	e.applyTenantIsolation(field, user)
	
	return nil
}

// applyPermissions applies permission-based field restrictions
func (e *DefaultEnricher) applyPermissions(field *schema.Field, user User) {
	// Check if field requires specific permission
	if field.RequirePermission != "" {
		hasPermission := user.HasPermission(field.RequirePermission)
		
		field.Runtime.Visible = hasPermission
		field.Runtime.Editable = hasPermission
		
		if !hasPermission {
			field.Runtime.Reason = "permission_required"
		}
	} else {
		// Default to visible and editable if no permission required
		field.Runtime.Visible = true
		field.Runtime.Editable = !field.Readonly
	}
	
	// Check role-based restrictions
	if len(field.RequireRoles) > 0 {
		hasRequiredRole := false
		for _, requiredRole := range field.RequireRoles {
			if user.HasRole(requiredRole) {
				hasRequiredRole = true
				break
			}
		}
		
		if !hasRequiredRole {
			field.Runtime.Visible = false
			field.Runtime.Editable = false
			field.Runtime.Reason = "role_required"
		}
	}
}

// applyDynamicDefaults applies dynamic default values based on context
func (e *DefaultEnricher) applyDynamicDefaults(field *schema.Field, data map[string]any) {
	// Only apply defaults if field has no current value
	if field.Value != nil {
		return
	}
	
	// Apply static default value if exists
	if field.Default != nil {
		field.Value = field.Default
		return
	}
	
	// Apply dynamic defaults based on field name
	switch field.Name {
	case "created_by", "user_id", "author_id":
		if userID, ok := data["user_id"].(string); ok {
			field.Value = userID
		}
	case "tenant_id", "organization_id":
		if tenantID, ok := data["tenant_id"].(string); ok {
			field.Value = tenantID
		}
	case "created_at", "timestamp":
		if timestamp, ok := data["timestamp"].(time.Time); ok {
			field.Value = timestamp.Format(time.RFC3339)
		}
	case "updated_at":
		field.Value = time.Now().Format(time.RFC3339)
	}
}

// applyTenantIsolation applies tenant isolation for multi-tenant fields
func (e *DefaultEnricher) applyTenantIsolation(field *schema.Field, user User) {
	// Automatically set tenant_id field for tenant isolation
	if field.Name == "tenant_id" {
		field.Value = user.GetTenantID()
		field.Hidden = true // Hide tenant_id from user
		field.Readonly = true
	}
}

// applyTenantCustomization applies tenant-specific customizations
func (e *DefaultEnricher) applyTenantCustomization(field *schema.Field, customization *TenantCustomization) {
	override, exists := customization.Overrides[field.Name]
	if !exists {
		return
	}
	
	// Apply label override
	if override.Label != "" {
		field.Label = override.Label
	}
	
	// Apply required override
	if override.Required != nil {
		field.Required = *override.Required
	}
	
	// Apply hidden override
	if override.Hidden != nil {
		field.Hidden = *override.Hidden
	}
	
	// Apply default value override
	if override.DefaultValue != nil {
		field.Default = override.DefaultValue
	}
	
	// Apply placeholder override
	if override.Placeholder != "" {
		field.Placeholder = override.Placeholder
	}
	
	// Apply help text override
	if override.Help != "" {
		field.Help = override.Help
	}
}

// enrichAction enriches actions with user permissions
func (e *DefaultEnricher) enrichAction(ctx context.Context, action *schema.Action, user User) {
	// Check action permissions
	if action.Permissions != nil {
		// Check view permissions
		if len(action.Permissions.View) > 0 {
			canView := false
			userPermissions := user.GetPermissions()
			for _, required := range action.Permissions.View {
				for _, userPerm := range userPermissions {
					if userPerm == required {
						canView = true
						break
					}
				}
				if canView {
					break
				}
			}
			action.Hidden = !canView
		}
		
		// Check execute permissions
		if len(action.Permissions.Execute) > 0 {
			canExecute := false
			userPermissions := user.GetPermissions()
			for _, required := range action.Permissions.Execute {
				for _, userPerm := range userPermissions {
					if userPerm == required {
						canExecute = true
						break
					}
				}
				if canExecute {
					break
				}
			}
			action.Disabled = !canExecute
		}
	}
}

// DefaultUser is a simple implementation of the User interface for testing
type DefaultUser struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

// GetID returns the user ID
func (u *DefaultUser) GetID() string {
	return u.ID
}

// GetTenantID returns the tenant ID
func (u *DefaultUser) GetTenantID() string {
	return u.TenantID
}

// GetPermissions returns user permissions
func (u *DefaultUser) GetPermissions() []string {
	return u.Permissions
}

// GetRoles returns user roles
func (u *DefaultUser) GetRoles() []string {
	return u.Roles
}

// HasPermission checks if user has specific permission
func (u *DefaultUser) HasPermission(permission string) bool {
	for _, perm := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// HasRole checks if user has specific role
func (u *DefaultUser) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}