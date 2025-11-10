package schema

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
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
	Tags       []string               `json:"tags,omitempty"`
	License    string                 `json:"license,omitempty"`
	Repository string                 `json:"repository,omitempty" validate:"url"`
	Homepage   string                 `json:"homepage,omitempty" validate:"url"`
	Preview    string                 `json:"preview,omitempty" validate:"url"` // Preview image URL
	CustomData map[string]any `json:"customData,omitempty"`
	CreatedAt  time.Time              `json:"createdAt,omitempty"`
	UpdatedAt  time.Time              `json:"updatedAt,omitempty"`
}

// ThemeOverrides represents customizations that can be applied to a base theme.
// This enables runtime theme customization without modifying the original theme.
type ThemeOverrides struct {
	// Token overrides - specific token values to override
	TokenOverrides map[string]string `json:"tokenOverrides,omitempty"`

	// Component customizations
	ComponentOverrides map[string]any `json:"componentOverrides,omitempty"`

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
	TenantID    string                 `json:"tenantId" validate:"required"`
	BaseThemeID string                 `json:"baseThemeId" validate:"required"`
	Overrides   *ThemeOverrides        `json:"overrides,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// ThemeCacheEntry represents a cached theme with timestamp
type ThemeCacheEntry struct {
	theme     *Theme
	timestamp time.Time
}

// ThemeCache implements a bounded LRU cache with TTL
type ThemeCache struct {
	entries map[string]*ThemeCacheEntry
	order   []string // LRU order (most recent first)
	maxSize int
	ttl     time.Duration
	mu      sync.RWMutex
}

// NewThemeCache creates a new LRU cache with TTL
func NewThemeCache(maxSize int, ttl time.Duration) *ThemeCache {
	return &ThemeCache{
		entries: make(map[string]*ThemeCacheEntry),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get retrieves a theme from cache
func (tc *ThemeCache) Get(key string) (*Theme, bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	entry, exists := tc.entries[key]
	if !exists {
		return nil, false
	}

	// Check TTL
	if time.Since(entry.timestamp) > tc.ttl {
		tc.removeUnsafe(key)
		return nil, false
	}

	// Move to front (most recently used)
	tc.moveToFrontUnsafe(key)
	return entry.theme, true
}

// Set stores a theme in cache
func (tc *ThemeCache) Set(key string, theme *Theme) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// If key exists, update and move to front
	if _, exists := tc.entries[key]; exists {
		tc.entries[key] = &ThemeCacheEntry{
			theme:     theme,
			timestamp: time.Now(),
		}
		tc.moveToFrontUnsafe(key)
		return
	}

	// If at capacity, remove LRU item
	if len(tc.entries) >= tc.maxSize {
		lruKey := tc.order[len(tc.order)-1]
		tc.removeUnsafe(lruKey)
	}

	// Add new entry
	tc.entries[key] = &ThemeCacheEntry{
		theme:     theme,
		timestamp: time.Now(),
	}
	tc.order = append([]string{key}, tc.order...)
}

// Clear removes all entries from cache
func (tc *ThemeCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.entries = make(map[string]*ThemeCacheEntry)
	tc.order = tc.order[:0]
}

// removeUnsafe removes an entry (must hold lock)
func (tc *ThemeCache) removeUnsafe(key string) {
	delete(tc.entries, key)

	// Remove from order slice
	for i, k := range tc.order {
		if k == key {
			tc.order = append(tc.order[:i], tc.order[i+1:]...)
			break
		}
	}
}

// moveToFrontUnsafe moves key to front of LRU order (must hold lock)
func (tc *ThemeCache) moveToFrontUnsafe(key string) {
	// Remove from current position
	for i, k := range tc.order {
		if k == key {
			tc.order = append(tc.order[:i], tc.order[i+1:]...)
			break
		}
	}

	// Add to front
	tc.order = append([]string{key}, tc.order...)
}

// ThemeManager provides high-level theme management with runtime customization.
// Handles theme registration, resolution, and customization with caching.
type ThemeManager struct {
	registry     *ThemeRegistry
	tokenManager *TokenRegistry
	cache        *ThemeCache
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
		cache:        NewThemeCache(100, 10*time.Minute), // 100 themes, 10min TTL
	}
}

// RegisterTheme registers a new theme in the manager.
//
// Example:
//
//	manager := schema.GetGlobalThemeManager()
//
//	// Create and register a theme
//	theme, err := schema.NewTheme("Company Theme").
//	    WithID("company-v1").
//	    WithTokens(schema.GetDefaultTokens()).
//	    Build()
//	if err != nil {
//	    return err
//	}
//
//	// Register the theme
//	err = manager.RegisterTheme(theme)
//	if err != nil {
//	    return fmt.Errorf("failed to register theme: %w", err)
//	}
//
//	fmt.Printf("Theme %s registered successfully\n", theme.ID)
func (tm *ThemeManager) RegisterTheme(theme *Theme) error {
	if err := tm.ValidateTheme(theme); err != nil {
		return WrapError(err, "theme_validation_failed", "theme validation failed")
	}

	return tm.registry.Register(theme)
}

// GetTheme retrieves a theme by ID with caching.
//
// Example:
//
//	manager := schema.GetGlobalThemeManager()
//	ctx := context.Background()
//
//	// Get a theme by ID
//	theme, err := manager.GetTheme(ctx, "company-v1")
//	if err != nil {
//	    if schema.IsNotFoundError(err) {
//	        fmt.Println("Theme not found")
//	        return nil
//	    }
//	    return fmt.Errorf("failed to get theme: %w", err)
//	}
//
//	fmt.Printf("Retrieved theme: %s (%s)\n", theme.Name, theme.Version)
//
//	// List all available themes
//	themes, err := manager.ListThemes(ctx)
//	if err != nil {
//	    return err
//	}
//
//	for _, t := range themes {
//	    fmt.Printf("- %s: %s\n", t.ID, t.Name)
//	}
func (tm *ThemeManager) GetTheme(ctx context.Context, themeID string) (*Theme, error) {
	// Check cache first
	if cached, exists := tm.cache.Get(themeID); exists {
		return cached, nil
	}

	// Get from registry
	theme, err := tm.registry.Get(themeID)
	if err != nil {
		return nil, err
	}

	// Cache the theme
	tm.cache.Set(themeID, theme)

	return theme, nil
}

// GetThemeWithOverrides applies overrides to a base theme and returns the result.
//
// Example:
//
//	manager := schema.GetGlobalThemeManager()
//	ctx := context.Background()
//
//	// Create tenant-specific overrides
//	overrides := &schema.ThemeOverrides{
//	    TokenOverrides: map[string]string{
//	        "colors.primary.500":   "#f59e0b", // Custom brand color
//	        "spacing.base":         "1rem",    // Custom spacing
//	        "typography.family":    "'Custom Font', sans-serif",
//	    },
//	    ComponentOverrides: map[string]any{
//	        "button.borderRadius": "8px",
//	        "card.shadow":        "0 4px 12px rgba(0,0,0,0.15)",
//	    },
//	    CustomCSS: `
//	        .header { background: linear-gradient(135deg, #f59e0b, #d97706); }
//	        .logo { font-family: 'Custom Font', sans-serif; }
//	    `,
//	}
//
//	// Apply overrides to base theme
//	customTheme, err := manager.GetThemeWithOverrides(ctx, "base-theme", overrides)
//	if err != nil {
//	    return fmt.Errorf("failed to apply overrides: %w", err)
//	}
//
//	fmt.Printf("Custom theme created with %d token overrides\n", len(overrides.TokenOverrides))
func (tm *ThemeManager) GetThemeWithOverrides(ctx context.Context, themeID string, overrides *ThemeOverrides) (*Theme, error) {
	// Get base theme
	baseTheme, err := tm.GetTheme(ctx, themeID)
	if err != nil {
		return nil, err
	}

	if overrides == nil {
		return baseTheme, nil
	}

	// Create a deep copy of the base theme to avoid mutations
	customizedTheme, err := deepCopyTheme(baseTheme)
	if err != nil {
		return nil, WrapError(err, "theme_customization_failed", "failed to create customized theme copy")
	}

	// Update identity for customized theme
	customizedTheme.ID = baseTheme.ID + "_customized"
	customizedTheme.Name = baseTheme.Name + " (Customized)"
	customizedTheme.createdAt = time.Now()
	customizedTheme.updatedAt = time.Now()

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

	// Apply token overrides with deep modification
	if overrides.TokenOverrides != nil && len(overrides.TokenOverrides) > 0 {
		err := tm.applyTokenOverrides(customizedTheme, overrides.TokenOverrides)
		if err != nil {
			return nil, WrapError(err, "token_override_failed", "failed to apply token overrides")
		}
	}

	// Apply dark mode overrides
	if overrides.DarkModeOverrides != nil && len(overrides.DarkModeOverrides) > 0 {
		err := tm.applyDarkModeOverrides(customizedTheme, overrides.DarkModeOverrides)
		if err != nil {
			return nil, WrapError(err, "dark_mode_override_failed", "failed to apply dark mode overrides")
		}
	}

	return customizedTheme, nil
}

// applyTokenOverrides applies token path overrides to a theme's tokens
func (tm *ThemeManager) applyTokenOverrides(theme *Theme, overrides map[string]string) error {
	if theme.Tokens == nil {
		return NewValidationError("tokens_nil", "theme tokens are nil")
	}

	for path, value := range overrides {
		err := tm.setTokenValue(theme.Tokens, path, value)
		if err != nil {
			return WrapError(err, "token_override_failed", "failed to override token: "+path)
		}
	}

	return nil
}

// applyDarkModeOverrides applies dark mode token overrides
func (tm *ThemeManager) applyDarkModeOverrides(theme *Theme, overrides map[string]string) error {
	// Ensure dark mode configuration exists
	if theme.DarkMode == nil {
		theme.DarkMode = &DarkModeConfig{
			Enabled:  true,
			Strategy: "class",
		}
	}

	// Ensure dark tokens exist
	if theme.DarkMode.DarkTokens == nil {
		// Create a copy of the base tokens for dark mode customization
		darkTokens, err := deepCopyTokens(theme.Tokens)
		if err != nil {
			return WrapError(err, "dark_tokens_copy_failed", "failed to create dark tokens copy")
		}
		theme.DarkMode.DarkTokens = darkTokens
	}

	// Apply overrides to dark tokens
	for path, value := range overrides {
		err := tm.setTokenValue(theme.DarkMode.DarkTokens, path, value)
		if err != nil {
			return WrapError(err, "dark_mode_override_failed", "failed to override dark mode token: "+path)
		}
	}

	return nil
}

// setTokenValue sets a token value at the specified path using reflection
func (tm *ThemeManager) setTokenValue(tokens *DesignTokens, path, value string) error {
	// This is a simplified implementation - in production you might want
	// a more robust path parser
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return NewValidationError("invalid_token_path", "token path must have at least two parts: "+path)
	}

	// Use reflection to navigate and set the token value
	current := reflect.ValueOf(tokens).Elem()

	// Navigate to the target field
	for _, part := range parts[:len(parts)-1] {
		field := findStructFieldByName(current, part)
		if !field.IsValid() {
			return NewValidationError("token_field_not_found", "token field not found: "+part+" in path "+path)
		}

		// Handle pointer fields
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				// Create new instance if nil
				field.Set(reflect.New(field.Type().Elem()))
			}
			current = field.Elem()
		} else {
			current = field
		}
	}

	// Set the final value
	finalPart := parts[len(parts)-1]
	field := findStructFieldByName(current, finalPart)
	if !field.IsValid() {
		return NewValidationError("token_field_not_found", "final token field not found: "+finalPart+" in path "+path)
	}

	if !field.CanSet() {
		return NewValidationError("token_field_not_settable", "token field is not settable: "+finalPart)
	}

	// Set the value based on field type
	if field.Kind() == reflect.String {
		field.SetString(value)
	} else if field.Type() == reflect.TypeOf(TokenReference("")) {
		field.Set(reflect.ValueOf(TokenReference(value)))
	} else {
		return NewValidationError("unsupported_token_type", "unsupported token field type for: "+finalPart)
	}

	return nil
}

// findStructFieldByName finds a struct field by name (case-insensitive with JSON tag support)
func findStructFieldByName(structValue reflect.Value, fieldName string) reflect.Value {
	structType := structValue.Type()

	// First try exact match
	if field := structValue.FieldByName(fieldName); field.IsValid() {
		return field
	}

	// Then try case-insensitive match and JSON tags
	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)

		// Check field name (case-insensitive)
		if strings.EqualFold(field.Name, fieldName) {
			return structValue.Field(i)
		}

		// Check JSON tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagParts := strings.Split(jsonTag, ",")
			if len(tagParts) > 0 && strings.EqualFold(tagParts[0], fieldName) {
				return structValue.Field(i)
			}
		}
	}

	return reflect.Value{} // Invalid value
}

// deepCopyTokens creates a deep copy of design tokens
func deepCopyTokens(source *DesignTokens) (*DesignTokens, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, WrapError(err, "tokens_marshal_failed", "failed to marshal tokens for deep copy")
	}

	var copied DesignTokens
	err = json.Unmarshal(data, &copied)
	if err != nil {
		return nil, WrapError(err, "tokens_unmarshal_failed", "failed to unmarshal tokens for deep copy")
	}

	return &copied, nil
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
	tm.cache.Clear()
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

// deepCopyTheme creates a deep copy of a theme for safe cloning
func deepCopyTheme(source *Theme) (*Theme, error) {
	// Use JSON marshaling/unmarshaling for deep copy
	// This ensures all nested structures are properly copied
	data, err := json.Marshal(source)
	if err != nil {
		return nil, WrapError(err, "deep_copy_marshal_failed", "failed to marshal theme for deep copy")
	}

	var copied Theme
	err = json.Unmarshal(data, &copied)
	if err != nil {
		return nil, WrapError(err, "deep_copy_unmarshal_failed", "failed to unmarshal theme for deep copy")
	}

	// Preserve internal fields that don't serialize
	copied.createdAt = source.createdAt
	copied.updatedAt = source.updatedAt

	return &copied, nil
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

	// Create a deep copy of the theme
	clonedTheme, err := deepCopyTheme(sourceTheme)
	if err != nil {
		return nil, WrapError(err, "theme_clone_failed", "failed to clone theme")
	}

	// Update the copied theme with new identity
	clonedTheme.ID = newID
	clonedTheme.Name = sourceTheme.Name + " (Copy)"
	clonedTheme.createdAt = time.Now()
	clonedTheme.updatedAt = time.Now()

	tr.themes[newID] = clonedTheme
	return clonedTheme, nil
}

// ThemeBuilder provides a fluent interface for constructing themes.
type ThemeBuilder struct {
	theme *Theme
}

// NewTheme creates a new theme builder with the given name.
//
// Example:
//
//	// Create a basic theme
//	theme, err := schema.NewTheme("Corporate Theme").
//	    WithID("corp-v2").
//	    WithDescription("Modern corporate design system").
//	    WithVersion("2.1.0").
//	    WithAuthor("Design Team").
//	    WithTokens(schema.GetDefaultTokens()).
//	    Build()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create theme with dark mode and accessibility
//	theme, err := schema.NewTheme("Accessible Theme").
//	    WithID("accessible-v1").
//	    WithTokens(schema.GetDefaultTokens()).
//	    WithDarkMode(&schema.DarkModeConfig{
//	        Enabled:  true,
//	        Strategy: "class",
//	    }).
//	    WithAccessibility(&schema.AccessibilityConfig{
//	        HighContrast:     true,
//	        MinContrastRatio: 4.5,
//	        KeyboardNav:      true,
//	    }).
//	    Build()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Register theme with global manager
//	manager := schema.GetGlobalThemeManager()
//	err = manager.RegisterTheme(theme)
//	if err != nil {
//	    log.Fatal(err)
//	}
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
	globalThemeRegistry *ThemeRegistry
	globalThemeManager  *ThemeManager
	globalTenantManager *TenantThemeManager
	themeRegistryOnce   sync.Once
	themeManagerOnce    sync.Once
	tenantManagerOnce   sync.Once
)

// GetGlobalThemeRegistry returns the global theme registry instance.
func GetGlobalThemeRegistry() *ThemeRegistry {
	themeRegistryOnce.Do(func() {
		globalThemeRegistry = NewThemeRegistry()
		// Register default themes
		if err := CreateDefaultThemes(); err != nil {
			// Log error but continue - themes are not critical for basic functionality
			// In production, you might want to handle this differently
			_ = err
		}
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

// ApplyTheme applies a theme to a schema.
// This is the integration point with the existing schema system.
//
// Example:
//
//	// Create a schema
//	schemaBuilder := schema.NewBuilder("user-profile").
//	    WithTitle("User Profile Form").
//	    WithDescription("Update user profile information")
//
//	schemaBuilder.AddField(&schema.Field{
//	    Name:     "email",
//	    Type:     schema.FieldTypeEmail,
//	    Label:    "Email Address",
//	    Required: true,
//	})
//
//	schemaBuilder.AddField(&schema.Field{
//	    Name:  "name",
//	    Type:  schema.FieldTypeText,
//	    Label: "Full Name",
//	})
//
//	userSchema, err := schemaBuilder.Build()
//	if err != nil {
//	    return err
//	}
//
//	// Apply a theme to the schema
//	ctx := context.Background()
//	err = userSchema.ApplyTheme(ctx, "corporate-v2")
//	if err != nil {
//	    return fmt.Errorf("failed to apply theme: %w", err)
//	}
//
//	fmt.Printf("Applied theme %s to schema %s\n", userSchema.Meta.Theme.ID, userSchema.Title)
//
//	// The schema now uses the theme's design tokens for consistent styling
//	// Components will automatically resolve tokens like:
//	// - {semantic.colors.background.default} for backgrounds
//	// - {components.input.border} for input styling
//	// - {semantic.spacing.component.default} for spacing
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
		ID:            override.ID,
		Name:          override.Name,
		Description:   override.Description,
		Version:       override.Version,
		Author:        override.Author,
		Tokens:        override.Tokens,
		DarkMode:      override.DarkMode,
		Accessibility: override.Accessibility,
		Meta:          override.Meta,
		CustomCSS:     override.CustomCSS,
		CustomJS:      override.CustomJS,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
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

// ExportTheme exports a theme to JSON format with pretty formatting.
func ExportTheme(theme *Theme) ([]byte, error) {
	if theme == nil {
		return nil, NewValidationError("theme", "theme is required")
	}
	
	// Export with pretty formatting and sorted keys for consistency
	return json.MarshalIndent(theme, "", "  ")
}

// ImportTheme imports a theme from JSON format with validation.
func ImportTheme(data []byte) (*Theme, error) {
	if len(data) == 0 {
		return nil, NewValidationError("data", "theme data is required")
	}
	
	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, WrapError(err, "theme_unmarshal_failed", "failed to unmarshal theme")
	}
	
	// Validate the imported theme
	if err := ValidateTheme(&theme); err != nil {
		return nil, WrapError(err, "theme_validation_failed", "imported theme validation failed")
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
