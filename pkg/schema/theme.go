package schema

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// Theme represents a complete theme configuration for the UI system.
// A theme is essentially a configured set of design tokens with metadata.
//
// Design Philosophy:
//   - Themes are immutable once created (use ThemeManager for customization)
//   - All styling flows through design tokens (single source of truth)
//   - Supports runtime customization via ThemeManager
//   - Multi-tenant capable with TenantThemeManager
//
// Breaking Changes: The Theme structure is designed to be stable, but new
// optional fields may be added for enhanced functionality.
type Theme struct {
	// Identity
	ID          string `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty" validate:"semver"`
	Author      string `json:"author,omitempty"`

	// Design tokens - the single source of truth for all styling
	Tokens *DesignTokens `json:"tokens" validate:"required"`

	// Dark mode configuration
	DarkMode *DarkModeConfig `json:"darkMode,omitempty"`

	// Accessibility configuration
	Accessibility *AccessibilityConfig `json:"accessibility,omitempty"`

	// Metadata
	Meta *ThemeMeta `json:"meta,omitempty"`

	// Custom extensions (advanced users only)
	// Breaking Changes: CustomCSS/CustomJS may be deprecated in favor of
	// a more structured extension system.
	CustomCSS string `json:"customCSS,omitempty"`
	CustomJS  string `json:"customJS,omitempty"`

	// Internal fields (not serialized)
	createdAt time.Time
	updatedAt time.Time
}

// DarkModeConfig defines dark mode behavior and color overrides.
type DarkModeConfig struct {
	// Enabled indicates if dark mode is available
	Enabled bool `json:"enabled"`

	// Default indicates if dark mode should be the default
	Default bool `json:"default,omitempty"`

	// Strategy defines how dark mode is detected/applied
	// Valid values: "class", "media", "auto"
	Strategy string `json:"strategy,omitempty" validate:"oneof=class media auto"`

	// DarkTokens are token overrides specifically for dark mode
	// These override the base tokens when dark mode is active
	DarkTokens *DesignTokens `json:"darkTokens,omitempty"`
}

// AccessibilityConfig defines accessibility settings and preferences.
// These settings help meet WCAG 2.1 Level AA compliance.
type AccessibilityConfig struct {
	// ARIA
	AutoARIA        bool   `json:"autoAria,omitempty"` // Auto-generate ARIA attributes
	AriaLive        string `json:"ariaLive,omitempty" validate:"oneof=off polite assertive"`
	AriaDescribedBy bool   `json:"ariaDescribedBy,omitempty"` // Auto-link descriptions

	// Keyboard navigation
	KeyboardNav        bool `json:"keyboardNav,omitempty"`        // Enable enhanced keyboard nav
	FocusIndicator     bool `json:"focusIndicator,omitempty"`     // Enhanced focus indicators
	SkipLinks          bool `json:"skipLinks,omitempty"`          // Add skip navigation links
	TabIndexManagement bool `json:"tabIndexManagement,omitempty"` // Automatic tabindex management

	// Screen reader optimizations
	ScreenReaderOnly  bool `json:"screenReaderOnly,omitempty"`  // Screen reader enhancements
	LiveAnnouncements bool `json:"liveAnnouncements,omitempty"` // Announce dynamic changes

	// Contrast and visibility
	HighContrast     bool    `json:"highContrast,omitempty"`                             // High contrast mode
	MinContrastRatio float64 `json:"minContrastRatio,omitempty" validate:"gte=0,lte=21"` // WCAG ratio

	// Focus styling
	FocusOutlineColor string `json:"focusOutlineColor,omitempty"`
	FocusOutlineWidth string `json:"focusOutlineWidth,omitempty"`

	// Motion preferences
	ReducedMotion bool `json:"reducedMotion,omitempty"` // Respect prefers-reduced-motion
}

// ThemeMeta contains theme metadata and organizational information.
type ThemeMeta struct {
	Tags       []string          `json:"tags,omitempty"`
	License    string            `json:"license,omitempty"`
	Repository string            `json:"repository,omitempty" validate:"url"`
	Homepage   string            `json:"homepage,omitempty" validate:"url"`
	Preview    string            `json:"preview,omitempty" validate:"url"` // Preview image URL
	CustomData map[string]interface{} `json:"customData,omitempty"`
	CreatedAt  time.Time         `json:"createdAt,omitempty"`
	UpdatedAt  time.Time         `json:"updatedAt,omitempty"`
}

// ThemeOverrides represents customizations that can be applied to a base theme.
// This enables runtime theme customization without modifying the original theme.
type ThemeOverrides struct {
	// Token overrides - specific token values to override
	TokenOverrides map[string]string `json:"tokenOverrides,omitempty"`

	// Component customizations
	ComponentOverrides map[string]interface{} `json:"componentOverrides,omitempty"`

	// Custom CSS to inject
	CustomCSS string `json:"customCSS,omitempty"`

	// Accessibility overrides
	AccessibilityOverrides *AccessibilityConfig `json:"accessibilityOverrides,omitempty"`

	// Dark mode specific overrides
	DarkModeOverrides map[string]string `json:"darkModeOverrides,omitempty"`
}

// TenantConfig represents tenant-specific theming configuration.
// Enables multi-tenant theme customization with tenant isolation.
type TenantConfig struct {
	TenantID    string           `json:"tenantId" validate:"required"`
	BaseThemeID string           `json:"baseThemeId" validate:"required"`
	Overrides   *ThemeOverrides  `json:"overrides,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Active      bool             `json:"active"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

// ThemeManager provides high-level theme management with runtime customization.
// Handles theme registration, resolution, and customization with caching.
type ThemeManager struct {
	registry     *ThemeRegistry
	tokenManager *TokenRegistry
	cache        map[string]*Theme // TODO: Implement LRU cache
	mu           sync.RWMutex
}

// NewThemeManager creates a new theme manager with the given registries.
func NewThemeManager(themeRegistry *ThemeRegistry, tokenManager *TokenRegistry) *ThemeManager {
	if themeRegistry == nil {
		themeRegistry = NewThemeRegistry()
	}
	if tokenManager == nil {
		tokenManager = NewTokenRegistry()
	}
	
	return &ThemeManager{
		registry:     themeRegistry,
		tokenManager: tokenManager,
		cache:        make(map[string]*Theme),
	}
}

// RegisterTheme registers a new theme in the manager.
func (tm *ThemeManager) RegisterTheme(theme *Theme) error {
	if err := tm.ValidateTheme(theme); err != nil {
		return WrapError(err, "theme_validation_failed", "theme validation failed")
	}
	
	return tm.registry.Register(theme)
}

// GetTheme retrieves a theme by ID with caching.
func (tm *ThemeManager) GetTheme(ctx context.Context, themeID string) (*Theme, error) {
	// Check cache first
	tm.mu.RLock()
	if cached, exists := tm.cache[themeID]; exists {
		tm.mu.RUnlock()
		return cached, nil
	}
	tm.mu.RUnlock()
	
	// Get from registry
	theme, err := tm.registry.Get(themeID)
	if err != nil {
		return nil, err
	}
	
	// Cache the theme
	tm.mu.Lock()
	tm.cache[themeID] = theme
	tm.mu.Unlock()
	
	return theme, nil
}

// GetThemeWithOverrides applies overrides to a base theme and returns the result.
func (tm *ThemeManager) GetThemeWithOverrides(ctx context.Context, themeID string, overrides *ThemeOverrides) (*Theme, error) {
	// Get base theme
	baseTheme, err := tm.GetTheme(ctx, themeID)
	if err != nil {
		return nil, err
	}
	
	if overrides == nil {
		return baseTheme, nil
	}
	
	// Apply overrides to create a new theme
	customizedTheme := &Theme{
		ID:            baseTheme.ID + "_customized",
		Name:          baseTheme.Name + " (Customized)",
		Description:   baseTheme.Description,
		Version:       baseTheme.Version,
		Author:        baseTheme.Author,
		Tokens:        baseTheme.Tokens,
		DarkMode:      baseTheme.DarkMode,
		Accessibility: baseTheme.Accessibility,
		Meta:          baseTheme.Meta,
		CustomCSS:     baseTheme.CustomCSS,
		CustomJS:      baseTheme.CustomJS,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}
	
	// Apply accessibility overrides
	if overrides.AccessibilityOverrides != nil {
		customizedTheme.Accessibility = overrides.AccessibilityOverrides
	}
	
	// Apply custom CSS
	if overrides.CustomCSS != "" {
		if customizedTheme.CustomCSS != "" {
			customizedTheme.CustomCSS += "\n" + overrides.CustomCSS
		} else {
			customizedTheme.CustomCSS = overrides.CustomCSS
		}
	}
	
	// TODO: Apply token overrides (would require deep cloning and modification of tokens)
	// For now, we return the base theme with CSS customizations
	
	return customizedTheme, nil
}

// ResolveTheme fully resolves all token references in a theme.
func (tm *ThemeManager) ResolveTheme(ctx context.Context, theme *Theme) (*Theme, error) {
	if theme == nil {
		return nil, NewValidationError("theme_nil", "theme is nil")
	}
	
	// For now, delegate to token manager (full resolution would be complex)
	resolvedTokens, err := tm.tokenManager.ResolveAllTokens(ctx)
	if err != nil {
		return nil, WrapError(err, "token_resolution_failed", "failed to resolve tokens")
	}
	
	// Create a resolved copy of the theme
	resolvedTheme := *theme
	resolvedTheme.Tokens = resolvedTokens
	
	return &resolvedTheme, nil
}

// InvalidateCache clears the theme cache.
func (tm *ThemeManager) InvalidateCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.cache = make(map[string]*Theme)
}

// ListThemes returns all registered themes.
func (tm *ThemeManager) ListThemes(ctx context.Context) ([]*Theme, error) {
	return tm.registry.List()
}

// ValidateTheme validates a theme configuration for correctness.
func (tm *ThemeManager) ValidateTheme(theme *Theme) error {
	if theme == nil {
		return NewValidationError("theme_nil", "theme is nil")
	}
	
	if theme.ID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	if theme.Name == "" {
		return NewValidationError("theme_name_required", "theme name is required")
	}
	
	if theme.Tokens == nil {
		return NewValidationError("theme_tokens_required", "theme tokens are required")
	}
	
	// Validate tokens using the token registry
	return tm.tokenManager.resolver.ValidateReferences(theme.Tokens)
}

// TenantThemeManager provides multi-tenant theme management.
// Enables tenant-specific theme customization with proper isolation.
type TenantThemeManager struct {
	themeManager  *ThemeManager
	tenantConfigs map[string]*TenantConfig
	mu            sync.RWMutex
}

// NewTenantThemeManager creates a new tenant theme manager.
func NewTenantThemeManager(themeManager *ThemeManager) *TenantThemeManager {
	if themeManager == nil {
		themeManager = GetGlobalThemeManager()
	}
	
	return &TenantThemeManager{
		themeManager:  themeManager,
		tenantConfigs: make(map[string]*TenantConfig),
	}
}

// SetTenantTheme configures a theme for a specific tenant.
func (ttm *TenantThemeManager) SetTenantTheme(ctx context.Context, tenantID, themeID string, overrides *ThemeOverrides) error {
	if tenantID == "" {
		return NewValidationError("tenant_id_required", "tenant ID is required")
	}
	
	if themeID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	// Verify theme exists
	_, err := ttm.themeManager.GetTheme(ctx, themeID)
	if err != nil {
		return WrapError(err, "theme_not_found", "base theme not found")
	}
	
	ttm.mu.Lock()
	defer ttm.mu.Unlock()
	
	// Create or update tenant config
	config := &TenantConfig{
		TenantID:    tenantID,
		BaseThemeID: themeID,
		Overrides:   overrides,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// If config exists, preserve creation time
	if existing, exists := ttm.tenantConfigs[tenantID]; exists {
		config.CreatedAt = existing.CreatedAt
	}
	
	ttm.tenantConfigs[tenantID] = config
	return nil
}

// GetTenantTheme retrieves the resolved theme for a specific tenant.
func (ttm *TenantThemeManager) GetTenantTheme(ctx context.Context, tenantID string) (*Theme, error) {
	if tenantID == "" {
		return nil, NewValidationError("tenant_id_required", "tenant ID is required")
	}
	
	ttm.mu.RLock()
	config, exists := ttm.tenantConfigs[tenantID]
	ttm.mu.RUnlock()
	
	if !exists || !config.Active {
		// Return default theme for tenant without configuration
		return ttm.themeManager.GetTheme(ctx, "default")
	}
	
	// Get base theme and apply overrides
	return ttm.themeManager.GetThemeWithOverrides(ctx, config.BaseThemeID, config.Overrides)
}

// ResolveTenantContext determines the tenant from the request context.
func (ttm *TenantThemeManager) ResolveTenantContext(ctx context.Context) (string, error) {
	// Try to extract tenant ID from context
	if tenantID, ok := ctx.Value("tenantID").(string); ok && tenantID != "" {
		return tenantID, nil
	}
	
	// Try alternative context keys
	if tenantID, ok := ctx.Value("tenant_id").(string); ok && tenantID != "" {
		return tenantID, nil
	}
	
	if tenantID, ok := ctx.Value("tenant").(string); ok && tenantID != "" {
		return tenantID, nil
	}
	
	return "", NewValidationError("tenant_context_missing", "tenant context not found in request")
}

// InvalidateTenantCache clears the cache for a specific tenant.
func (ttm *TenantThemeManager) InvalidateTenantCache(tenantID string) {
	// For now, invalidate the entire theme cache
	// In a more sophisticated implementation, we would track
	// tenant-specific cached themes
	ttm.themeManager.InvalidateCache()
}

// ListTenantConfigs returns all tenant configurations.
func (ttm *TenantThemeManager) ListTenantConfigs(ctx context.Context) ([]*TenantConfig, error) {
	ttm.mu.RLock()
	defer ttm.mu.RUnlock()
	
	configs := make([]*TenantConfig, 0, len(ttm.tenantConfigs))
	for _, config := range ttm.tenantConfigs {
		configs = append(configs, config)
	}
	
	return configs, nil
}

// ThemeRegistry provides storage and retrieval of themes with thread safety.
type ThemeRegistry struct {
	themes map[string]*Theme
	mu     sync.RWMutex
}

// NewThemeRegistry creates a new theme registry.
func NewThemeRegistry() *ThemeRegistry {
	return &ThemeRegistry{
		themes: make(map[string]*Theme),
	}
}

// Register registers a new theme in the registry.
func (tr *ThemeRegistry) Register(theme *Theme) error {
	if theme == nil {
		return NewValidationError("theme_nil", "theme cannot be nil")
	}
	
	if theme.ID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	tr.mu.Lock()
	defer tr.mu.Unlock()
	
	// Check if theme already exists
	if _, exists := tr.themes[theme.ID]; exists {
		return NewConflictError("theme", "theme with ID already exists: "+theme.ID)
	}
	
	// Set creation timestamp
	theme.createdAt = time.Now()
	theme.updatedAt = time.Now()
	
	tr.themes[theme.ID] = theme
	return nil
}

// Get retrieves a theme by ID.
func (tr *ThemeRegistry) Get(themeID string) (*Theme, error) {
	if themeID == "" {
		return nil, NewValidationError("theme_id_required", "theme ID is required")
	}
	
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	
	theme, exists := tr.themes[themeID]
	if !exists {
		return nil, NewNotFoundError("theme", "theme not found: "+themeID)
	}
	
	return theme, nil
}

// Update updates an existing theme.
func (tr *ThemeRegistry) Update(theme *Theme) error {
	if theme == nil {
		return NewValidationError("theme_nil", "theme cannot be nil")
	}
	
	if theme.ID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	tr.mu.Lock()
	defer tr.mu.Unlock()
	
	// Check if theme exists
	if _, exists := tr.themes[theme.ID]; !exists {
		return NewNotFoundError("theme", "theme not found: "+theme.ID)
	}
	
	// Update timestamp
	theme.updatedAt = time.Now()
	
	tr.themes[theme.ID] = theme
	return nil
}

// Delete removes a theme from the registry.
func (tr *ThemeRegistry) Delete(themeID string) error {
	if themeID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	tr.mu.Lock()
	defer tr.mu.Unlock()
	
	// Check if theme exists
	if _, exists := tr.themes[themeID]; !exists {
		return NewNotFoundError("theme", "theme not found: "+themeID)
	}
	
	delete(tr.themes, themeID)
	return nil
}

// List returns all registered themes.
func (tr *ThemeRegistry) List() ([]*Theme, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	
	themes := make([]*Theme, 0, len(tr.themes))
	for _, theme := range tr.themes {
		themes = append(themes, theme)
	}
	
	return themes, nil
}

// Exists checks if a theme exists in the registry.
func (tr *ThemeRegistry) Exists(themeID string) bool {
	if themeID == "" {
		return false
	}
	
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	
	_, exists := tr.themes[themeID]
	return exists
}

// Clone creates a copy of an existing theme with a new ID.
func (tr *ThemeRegistry) Clone(sourceID, newID string) (*Theme, error) {
	if sourceID == "" {
		return nil, NewValidationError("source_id_required", "source theme ID is required")
	}
	
	if newID == "" {
		return nil, NewValidationError("new_id_required", "new theme ID is required")
	}
	
	tr.mu.Lock()
	defer tr.mu.Unlock()
	
	// Get source theme
	sourceTheme, exists := tr.themes[sourceID]
	if !exists {
		return nil, NewNotFoundError("theme", "source theme not found: "+sourceID)
	}
	
	// Check if new ID already exists
	if _, exists := tr.themes[newID]; exists {
		return nil, NewConflictError("theme", "theme with new ID already exists: "+newID)
	}
	
	// Create a copy of the theme
	clonedTheme := &Theme{
		ID:            newID,
		Name:          sourceTheme.Name + " (Copy)",
		Description:   sourceTheme.Description,
		Version:       sourceTheme.Version,
		Author:        sourceTheme.Author,
		Tokens:        sourceTheme.Tokens, // Shallow copy for now
		DarkMode:      sourceTheme.DarkMode,
		Accessibility: sourceTheme.Accessibility,
		Meta:          sourceTheme.Meta,
		CustomCSS:     sourceTheme.CustomCSS,
		CustomJS:      sourceTheme.CustomJS,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}
	
	tr.themes[newID] = clonedTheme
	return clonedTheme, nil
}

// ThemeBuilder provides a fluent interface for constructing themes.
type ThemeBuilder struct {
	theme *Theme
}

// NewTheme creates a new theme builder with the given name.
func NewTheme(name string) *ThemeBuilder {
	return &ThemeBuilder{
		theme: &Theme{
			Name:      name,
			createdAt: time.Now(),
			updatedAt: time.Now(),
		},
	}
}

// WithID sets the theme ID.
func (tb *ThemeBuilder) WithID(id string) *ThemeBuilder {
	tb.theme.ID = id
	return tb
}

// WithDescription sets the theme description.
func (tb *ThemeBuilder) WithDescription(description string) *ThemeBuilder {
	tb.theme.Description = description
	return tb
}

// WithVersion sets the theme version.
func (tb *ThemeBuilder) WithVersion(version string) *ThemeBuilder {
	tb.theme.Version = version
	return tb
}

// WithAuthor sets the theme author.
func (tb *ThemeBuilder) WithAuthor(author string) *ThemeBuilder {
	tb.theme.Author = author
	return tb
}

// WithTokens sets the design tokens.
func (tb *ThemeBuilder) WithTokens(tokens *DesignTokens) *ThemeBuilder {
	tb.theme.Tokens = tokens
	return tb
}

// WithDarkMode configures dark mode settings.
func (tb *ThemeBuilder) WithDarkMode(config *DarkModeConfig) *ThemeBuilder {
	tb.theme.DarkMode = config
	return tb
}

// WithAccessibility configures accessibility settings.
func (tb *ThemeBuilder) WithAccessibility(config *AccessibilityConfig) *ThemeBuilder {
	tb.theme.Accessibility = config
	return tb
}

// WithMeta sets theme metadata.
func (tb *ThemeBuilder) WithMeta(meta *ThemeMeta) *ThemeBuilder {
	tb.theme.Meta = meta
	return tb
}

// WithCustomCSS adds custom CSS to the theme.
func (tb *ThemeBuilder) WithCustomCSS(css string) *ThemeBuilder {
	tb.theme.CustomCSS = css
	return tb
}

// WithCustomJS adds custom JavaScript to the theme.
func (tb *ThemeBuilder) WithCustomJS(js string) *ThemeBuilder {
	tb.theme.CustomJS = js
	return tb
}

// Build constructs and returns the final theme.
func (tb *ThemeBuilder) Build() (*Theme, error) {
	if tb.theme == nil {
		return nil, NewValidationError("theme_nil", "theme is nil")
	}
	
	// Validate required fields
	if tb.theme.ID == "" {
		return nil, NewValidationError("theme_id_required", "theme ID is required")
	}
	
	if tb.theme.Name == "" {
		return nil, NewValidationError("theme_name_required", "theme name is required")
	}
	
	// Set default tokens if not provided
	if tb.theme.Tokens == nil {
		tb.theme.Tokens = GetDefaultTokens()
	}
	
	// Update timestamp
	tb.theme.updatedAt = time.Now()
	
	return tb.theme, nil
}

// BuildAndRegister builds the theme and registers it with the global registry.
func (tb *ThemeBuilder) BuildAndRegister() (*Theme, error) {
	theme, err := tb.Build()
	if err != nil {
		return nil, err
	}
	
	registry := GetGlobalThemeRegistry()
	err = registry.Register(theme)
	if err != nil {
		return nil, WrapError(err, "theme_registration_failed", "failed to register theme")
	}
	
	return theme, nil
}

// Global theme management instances
var (
	globalThemeRegistry     *ThemeRegistry
	globalThemeManager      *ThemeManager
	globalTenantManager     *TenantThemeManager
	themeRegistryOnce       sync.Once
	themeManagerOnce        sync.Once
	tenantManagerOnce       sync.Once
)

// GetGlobalThemeRegistry returns the global theme registry instance.
func GetGlobalThemeRegistry() *ThemeRegistry {
	themeRegistryOnce.Do(func() {
		globalThemeRegistry = NewThemeRegistry()
		// TODO: Register default themes
	})
	return globalThemeRegistry
}

// GetGlobalThemeManager returns the global theme manager instance.
func GetGlobalThemeManager() *ThemeManager {
	themeManagerOnce.Do(func() {
		registry := GetGlobalThemeRegistry()
		tokenRegistry := GetDefaultRegistry()
		globalThemeManager = NewThemeManager(registry, tokenRegistry)
	})
	return globalThemeManager
}

// GetGlobalTenantManager returns the global tenant theme manager instance.
func GetGlobalTenantManager() *TenantThemeManager {
	tenantManagerOnce.Do(func() {
		themeManager := GetGlobalThemeManager()
		globalTenantManager = NewTenantThemeManager(themeManager)
	})
	return globalTenantManager
}

// Utility functions for theme operations

// ApplyThemeToSchema applies a theme to a schema.
// This is the integration point with the existing schema system.
func (s *Schema) ApplyTheme(ctx context.Context, themeID string) error {
	if themeID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	themeManager := GetGlobalThemeManager()
	theme, err := themeManager.GetTheme(ctx, themeID)
	if err != nil {
		return WrapError(err, "theme_application_failed", "failed to apply theme to schema")
	}
	
	// Store theme reference in schema
	if s.Meta == nil {
		s.Meta = &Meta{}
	}
	if s.Meta.Theme == nil {
		s.Meta.Theme = &ThemeConfig{}
	}
	s.Meta.Theme.ID = themeID
	
	// Apply theme to layout if present
	if s.Layout != nil {
		s.Layout.ApplyTheme(theme)
	}
	
	return nil
}

// ApplyThemeWithOverrides applies a theme with custom overrides to a schema.
func (s *Schema) ApplyThemeWithOverrides(ctx context.Context, themeID string, overrides *ThemeOverrides) error {
	if themeID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	themeManager := GetGlobalThemeManager()
	theme, err := themeManager.GetThemeWithOverrides(ctx, themeID, overrides)
	if err != nil {
		return WrapError(err, "theme_application_failed", "failed to apply theme with overrides to schema")
	}
	
	// Store theme reference in schema
	if s.Meta == nil {
		s.Meta = &Meta{}
	}
	if s.Meta.Theme == nil {
		s.Meta.Theme = &ThemeConfig{}
	}
	s.Meta.Theme.ID = themeID
	s.Meta.Theme.Overrides = overrides
	
	// Apply theme to layout if present
	if s.Layout != nil {
		s.Layout.ApplyTheme(theme)
	}
	
	return nil
}

// GetThemeFromContext extracts theme information from the request context.
func GetThemeFromContext(ctx context.Context) (*Theme, error) {
	// Try to extract theme ID from context
	if themeID, ok := ctx.Value("themeID").(string); ok && themeID != "" {
		themeManager := GetGlobalThemeManager()
		return themeManager.GetTheme(ctx, themeID)
	}
	
	// Try alternative context keys
	if themeID, ok := ctx.Value("theme_id").(string); ok && themeID != "" {
		themeManager := GetGlobalThemeManager()
		return themeManager.GetTheme(ctx, themeID)
	}
	
	// Try tenant-based theme resolution
	tenantManager := GetGlobalTenantManager()
	tenantID, err := tenantManager.ResolveTenantContext(ctx)
	if err == nil && tenantID != "" {
		return tenantManager.GetTenantTheme(ctx, tenantID)
	}
	
	// Fall back to default theme
	return GetDefaultTheme()
}

// WithThemeContext adds theme information to the context.
func WithThemeContext(ctx context.Context, themeID string) context.Context {
	return context.WithValue(ctx, "themeID", themeID)
}

// WithTenantContext adds tenant information to the context.
func WithTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenantID", tenantID)
}

// Theme helper functions

// MergeThemes merges two themes, with the override theme taking precedence.
func MergeThemes(base, override *Theme) (*Theme, error) {
	if base == nil {
		return nil, NewValidationError("base_theme_nil", "base theme cannot be nil")
	}
	
	if override == nil {
		return base, nil // No override, return base
	}
	
	// Create a new theme by merging
	merged := &Theme{
		ID:          override.ID,
		Name:        override.Name,
		Description: override.Description,
		Version:     override.Version,
		Author:      override.Author,
		Tokens:      override.Tokens,
		DarkMode:    override.DarkMode,
		Accessibility: override.Accessibility,
		Meta:        override.Meta,
		CustomCSS:   override.CustomCSS,
		CustomJS:    override.CustomJS,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
	
	// Fall back to base values where override is empty
	if merged.Name == "" {
		merged.Name = base.Name
	}
	if merged.Description == "" {
		merged.Description = base.Description
	}
	if merged.Version == "" {
		merged.Version = base.Version
	}
	if merged.Author == "" {
		merged.Author = base.Author
	}
	if merged.Tokens == nil {
		merged.Tokens = base.Tokens
	}
	if merged.DarkMode == nil {
		merged.DarkMode = base.DarkMode
	}
	if merged.Accessibility == nil {
		merged.Accessibility = base.Accessibility
	}
	if merged.Meta == nil {
		merged.Meta = base.Meta
	}
	if merged.CustomCSS == "" {
		merged.CustomCSS = base.CustomCSS
	}
	if merged.CustomJS == "" {
		merged.CustomJS = base.CustomJS
	}
	
	return merged, nil
}

// ValidateTheme validates a theme configuration for correctness and completeness.
func ValidateTheme(theme *Theme) error {
	if theme == nil {
		return NewValidationError("theme_nil", "theme cannot be nil")
	}
	
	if theme.ID == "" {
		return NewValidationError("theme_id_required", "theme ID is required")
	}
	
	if theme.Name == "" {
		return NewValidationError("theme_name_required", "theme name is required")
	}
	
	if theme.Tokens == nil {
		return NewValidationError("theme_tokens_required", "theme tokens are required")
	}
	
	// Validate tokens using token registry
	tokenRegistry := GetDefaultRegistry()
	if err := tokenRegistry.resolver.ValidateReferences(theme.Tokens); err != nil {
		return WrapError(err, "theme_token_validation_failed", "theme token validation failed")
	}
	
	return nil
}

// ExportTheme exports a theme to JSON format.
func ExportTheme(theme *Theme) ([]byte, error) {
	// TODO: Implement theme export with pretty formatting
	return json.MarshalIndent(theme, "", "  ")
}

// ImportTheme imports a theme from JSON format.
func ImportTheme(data []byte) (*Theme, error) {
	// TODO: Implement theme import with validation
	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, WrapError(err, "theme_unmarshal_failed", "failed to unmarshal theme")
	}
	return &theme, nil
}

// CreateDefaultThemes creates and registers the default system themes.
func CreateDefaultThemes() error {
	registry := GetGlobalThemeRegistry()
	
	// Create default light theme
	lightTheme, err := NewTheme("Default Light").
		WithID("default").
		WithDescription("Default light theme for the ERP system").
		WithVersion("1.0.0").
		WithAuthor("Awo ERP Team").
		WithTokens(GetDefaultTokens()).
		Build()
	if err != nil {
		return WrapError(err, "default_theme_creation_failed", "failed to create default light theme")
	}
	
	err = registry.Register(lightTheme)
	if err != nil {
		return WrapError(err, "default_theme_registration_failed", "failed to register default light theme")
	}
	
	// Create default dark theme
	darkTheme, err := NewTheme("Default Dark").
		WithID("default-dark").
		WithDescription("Default dark theme for the ERP system").
		WithVersion("1.0.0").
		WithAuthor("Awo ERP Team").
		WithTokens(GetDefaultTokens()).
		WithDarkMode(&DarkModeConfig{
			Enabled:  true,
			Default:  true,
			Strategy: "class",
		}).
		Build()
	if err != nil {
		return WrapError(err, "dark_theme_creation_failed", "failed to create default dark theme")
	}
	
	err = registry.Register(darkTheme)
	if err != nil {
		return WrapError(err, "dark_theme_registration_failed", "failed to register default dark theme")
	}
	
	return nil
}

// GetDefaultTheme returns the default light theme.
func GetDefaultTheme() (*Theme, error) {
	registry := GetGlobalThemeRegistry()
	return registry.Get("default")
}

// GetDefaultDarkTheme returns the default dark theme.
func GetDefaultDarkTheme() (*Theme, error) {
	registry := GetGlobalThemeRegistry()
	return registry.Get("default-dark")
}