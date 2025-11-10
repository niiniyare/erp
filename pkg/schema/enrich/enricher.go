package enrich

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	// I18n support
	GetPreferredLocale() string
}

// TenantOverride represents tenant-specific customizations
type TenantOverride struct {
	FieldName    string `json:"field_name"`
	Label        string `json:"label,omitempty"`
	Required     *bool  `json:"required,omitempty"`
	Hidden       *bool  `json:"hidden,omitempty"`
	DefaultValue any    `json:"default_value,omitempty"`
	Placeholder  string `json:"placeholder,omitempty"`
	Help         string `json:"help,omitempty"`
}

// TenantCustomization holds all tenant-specific overrides for a schema
type TenantCustomization struct {
	SchemaID  string                    `json:"schema_id"`
	TenantID  string                    `json:"tenant_id"`
	Overrides map[string]TenantOverride `json:"overrides"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
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

	// EnrichWithLocale enriches a schema with localization support
	EnrichWithLocale(ctx context.Context, schema *schema.Schema, user User, locale string) (*schema.Schema, error)

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

// Enrich enriches a schema with functional options (implements schema.Enricher interface)
func (e *DefaultEnricher) Enrich(ctx context.Context, s *schema.Schema, opts ...schema.EnrichOption) error {
	if s == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Build configuration from options
	config := &schema.EnrichConfig{
		User:           uuid.Nil,
		Environment:    "production",
		Features:       make(map[string]bool),
		FieldFilter:    nil,
		DepthLimit:     10,
		SkipValidation: false,
		Extensions:     make(map[string]any),
	}

	// Apply all options to build the configuration
	for _, opt := range opts {
		opt(config)
	}

	// Extract user information from config extensions or use default
	var user User
	if userInterface, exists := config.Extensions["user"]; exists {
		if u, ok := userInterface.(User); ok {
			user = u
		} else {
			return fmt.Errorf("invalid user type in extensions")
		}
	} else if config.User != uuid.Nil {
		// Create a basic user from UUID if no full user provided
		user = &BasicUser{
			ID:          config.User.String(),
			TenantID:    extractTenantFromConfig(config),
			Permissions: extractPermissionsFromConfig(config),
			Roles:       extractRolesFromConfig(config),
		}
	} else {
		return fmt.Errorf("no user information provided in enrichment options")
	}

	// Get tenant customizations if provider is available
	var customization *TenantCustomization
	if e.tenantProvider != nil {
		var err error
		customization, err = e.tenantProvider.GetCustomization(ctx, s.ID, user.GetTenantID())
		if err != nil {
			// Log error but continue with enrichment
			fmt.Printf("Warning: failed to get tenant customization: %v\n", err)
		}
	}

	// Create data context for dynamic defaults
	dataContext := map[string]any{
		"user_id":     user.GetID(),
		"tenant_id":   user.GetTenantID(),
		"timestamp":   time.Now(),
		"environment": config.Environment,
		"features":    config.Features,
	}

	// Add extensions to data context
	for key, value := range config.Extensions {
		dataContext[key] = value
	}

	// Enrich each field
	fieldsToProcess := s.Fields
	if config.FieldFilter != nil {
		fieldsToProcess = filterFields(s.Fields, config.FieldFilter)
	}

	for i := range fieldsToProcess {
		var field *schema.Field
		if config.FieldFilter != nil {
			// When filtered, we need to find the field in the original schema
			for j := range s.Fields {
				if s.Fields[j].Name == fieldsToProcess[i].Name {
					field = &s.Fields[j]
					break
				}
			}
		} else {
			// When not filtered, work directly with the original fields
			field = &s.Fields[i]
		}
		
		if field == nil {
			continue // Skip if field not found
		}

		// Apply tenant customizations
		if customization != nil {
			e.applyTenantCustomization(field, customization)
		}

		// Enrich the field
		if err := e.EnrichField(ctx, field, user, dataContext); err != nil {
			if !config.SkipValidation {
				return fmt.Errorf("failed to enrich field %s: %w", field.Name, err)
			}
			// Log error but continue if validation is skipped
			fmt.Printf("Warning: failed to enrich field %s: %v\n", field.Name, err)
		}
	}

	// Enrich actions
	for i := range s.Actions {
		action := &s.Actions[i]
		e.enrichAction(ctx, action, user)
	}

	// Validate enriched schema if not skipped
	if !config.SkipValidation {
		if err := s.Validate(ctx); err != nil {
			return fmt.Errorf("enriched schema validation failed: %w", err)
		}
	}

	return nil
}

// EnrichWithUser enriches a schema with runtime data based on user context (legacy method)
func (e *DefaultEnricher) EnrichWithUser(ctx context.Context, schemaObj *schema.Schema, user User) (*schema.Schema, error) {
	if schemaObj == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}

	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	// Clone the schema to avoid modifying the original
	enriched, err := schemaObj.Clone()
	if err != nil {
		return nil, fmt.Errorf("clone failed Error:%w", err)
	}
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

// EnrichWithLocale enriches a schema with localization support using embedded translations
func (e *DefaultEnricher) EnrichWithLocale(ctx context.Context, schemaObj *schema.Schema, user User, locale string) (*schema.Schema, error) {
	if schemaObj == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}

	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	// Use user's preferred locale if none specified
	if locale == "" {
		locale = user.GetPreferredLocale()
	}
	if locale == "" {
		locale = "en" // Default to English
	}

	// Validate the locale is available
	if !schema.T_HasLocale(locale) {
		// Fallback to detected locale or English
		locale = schema.T_DetectLocale([]string{locale, "en"})
	}

	// First clone the schema to avoid modifying the original
	localizedSchema, err := schemaObj.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone schema: %w", err)
	}

	// Apply embedded translations to all fields
	for i := range localizedSchema.Fields {
		field := &localizedSchema.Fields[i]
		
		// Localize field labels using embedded translations
		field.Label = field.GetLocalizedLabel(locale)
		field.Placeholder = field.GetLocalizedPlaceholder(locale)
		field.Help = field.GetLocalizedHelp(locale)
		field.Description = field.GetLocalizedDescription(locale)
	}

	// Apply embedded translations to actions
	for i := range localizedSchema.Actions {
		action := &localizedSchema.Actions[i]
		// Use action ID as translation key
		action.Text = schema.T_Action(locale, action.ID)
	}

	// Localize schema-level content if I18n is available
	if localizedSchema.I18n != nil {
		if title, exists := localizedSchema.I18n.Title[locale]; exists && title != "" {
			localizedSchema.Title = title
		}
		if desc, exists := localizedSchema.I18n.Description[locale]; exists && desc != "" {
			localizedSchema.Description = desc
		}
	}

	// Then enrich with user context
	enrichedSchema, err := e.EnrichWithUser(ctx, localizedSchema, user)
	if err != nil {
		return nil, fmt.Errorf("enrichment failed: %w", err)
	}

	return enrichedSchema, nil
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

// BasicUser is a simple implementation of the User interface for enrichment
type BasicUser struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
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

// GetPreferredLocale returns user's preferred locale
func (u *DefaultUser) GetPreferredLocale() string {
	return "en" // Default implementation
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

// BasicUser interface implementation

// GetID returns the user ID
func (u *BasicUser) GetID() string {
	return u.ID
}

// GetTenantID returns the tenant ID
func (u *BasicUser) GetTenantID() string {
	return u.TenantID
}

// GetPermissions returns user permissions
func (u *BasicUser) GetPermissions() []string {
	return u.Permissions
}

// GetRoles returns user roles
func (u *BasicUser) GetRoles() []string {
	return u.Roles
}

// GetPreferredLocale returns user's preferred locale
func (u *BasicUser) GetPreferredLocale() string {
	return "en" // Default implementation
}

// HasPermission checks if user has specific permission
func (u *BasicUser) HasPermission(permission string) bool {
	for _, perm := range u.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// HasRole checks if user has specific role
func (u *BasicUser) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// Helper functions for extracting data from EnrichOption configurations

// extractTenantFromConfig extracts tenant ID from enrichment config
func extractTenantFromConfig(config *schema.EnrichConfig) string {
	if tenantInterface, exists := config.Extensions["tenant_id"]; exists {
		if tenantID, ok := tenantInterface.(string); ok {
			return tenantID
		}
	}
	if tenantInterface, exists := config.Extensions["tenant"]; exists {
		if tenantID, ok := tenantInterface.(string); ok {
			return tenantID
		}
	}
	return "default"
}

// extractPermissionsFromConfig extracts permissions from enrichment config
func extractPermissionsFromConfig(config *schema.EnrichConfig) []string {
	if permInterface, exists := config.Extensions["permissions"]; exists {
		if permissions, ok := permInterface.([]string); ok {
			return permissions
		}
		if permissions, ok := permInterface.([]any); ok {
			var result []string
			for _, perm := range permissions {
				if permStr, ok := perm.(string); ok {
					result = append(result, permStr)
				}
			}
			return result
		}
	}
	return []string{}
}

// extractRolesFromConfig extracts roles from enrichment config
func extractRolesFromConfig(config *schema.EnrichConfig) []string {
	if roleInterface, exists := config.Extensions["roles"]; exists {
		if roles, ok := roleInterface.([]string); ok {
			return roles
		}
		if roles, ok := roleInterface.([]any); ok {
			var result []string
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					result = append(result, roleStr)
				}
			}
			return result
		}
	}
	return []string{}
}

// filterFields filters fields based on the provided filter function
func filterFields(fields []schema.Field, filter func(*schema.Field) bool) []schema.Field {
	var filtered []schema.Field
	for _, field := range fields {
		if filter(&field) {
			filtered = append(filtered, field)
		}
	}
	return filtered
}

// Helper functions to create EnrichOptions

// WithUser creates an EnrichOption that sets the user in extensions
func WithUser(user User) schema.EnrichOption {
	return func(config *schema.EnrichConfig) {
		config.Extensions["user"] = user
	}
}

// WithTenant creates an EnrichOption that sets the tenant ID
func WithTenant(tenantID string) schema.EnrichOption {
	return func(config *schema.EnrichConfig) {
		config.Extensions["tenant_id"] = tenantID
	}
}

// WithPermissions creates an EnrichOption that sets permissions
func WithPermissions(permissions []string) schema.EnrichOption {
	return func(config *schema.EnrichConfig) {
		config.Extensions["permissions"] = permissions
	}
}

// WithRoles creates an EnrichOption that sets roles
func WithRoles(roles []string) schema.EnrichOption {
	return func(config *schema.EnrichConfig) {
		config.Extensions["roles"] = roles
	}
}

// WithFieldFilter creates an EnrichOption that sets a field filter
func WithFieldFilter(filter func(*schema.Field) bool) schema.EnrichOption {
	return func(config *schema.EnrichConfig) {
		config.FieldFilter = filter
	}
}

// WithEnvironment creates an EnrichOption that sets the environment
func WithEnvironment(env string) schema.EnrichOption {
	return func(config *schema.EnrichConfig) {
		config.Environment = env
	}
}
