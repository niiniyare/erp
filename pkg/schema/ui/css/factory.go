package css

import "fmt"

// Factory provides convenient methods for creating styled components
type Factory struct {
	schemaDir string
}

// NewFactory creates a new CSS factory
func NewFactory(schemaDir string) *Factory {
	return &Factory{
		schemaDir: schemaDir,
	}
}

// GetSchemaDir returns the schema directory used by this factory
func (f *Factory) GetSchemaDir() string {
	return f.schemaDir
}

// ============================================================================
// COMPONENT STYLE FACTORIES
// ============================================================================

// ButtonStyles creates a styled button with various presets
func (f *Factory) ButtonStyles(variant string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	// Base button styles
	styles.WithDisplay("inline-flex").
		WithAlignItems("center").
		WithJustifyContent("center").
		WithPadding("0.5rem 1rem").
		WithBorderRadius("0.375rem").
		WithFontWeight("500").
		WithTransition("all 0.2s ease-in-out")
	
	// Variant-specific styles
	switch variant {
	case "primary":
		styles.WithBackgroundColor("#3b82f6").
			WithColor("#ffffff").
			WithCustomProperty("--hover-bg", "#2563eb")
	case "secondary":
		styles.WithBackgroundColor("#6b7280").
			WithColor("#ffffff").
			WithCustomProperty("--hover-bg", "#4b5563")
	case "success":
		styles.WithBackgroundColor("#10b981").
			WithColor("#ffffff").
			WithCustomProperty("--hover-bg", "#059669")
	case "danger":
		styles.WithBackgroundColor("#ef4444").
			WithColor("#ffffff").
			WithCustomProperty("--hover-bg", "#dc2626")
	case "outline":
		styles.WithBackgroundColor("transparent").
			WithColor("#3b82f6").
			WithCustomProperty("border", "1px solid #3b82f6").
			WithCustomProperty("--hover-bg", "#f3f4f6")
	default:
		// Default button
		styles.WithBackgroundColor("#f3f4f6").
			WithColor("#374151").
			WithCustomProperty("--hover-bg", "#e5e7eb")
	}
	
	return styles
}

// InputStyles creates styled form input
func (f *Factory) InputStyles(state string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	// Base input styles
	styles.WithDisplay("block").
		WithWidth("100%").
		WithPadding("0.5rem 0.75rem").
		WithBorderRadius("0.375rem").
		WithFontSize("1rem").
		WithTransition("border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out")
	
	// State-specific styles
	switch state {
	case "error":
		styles.WithCustomProperty("border", "1px solid #ef4444").
			WithCustomProperty("--focus-border", "#ef4444").
			WithCustomProperty("--focus-shadow", "0 0 0 3px rgba(239, 68, 68, 0.1)")
	case "success":
		styles.WithCustomProperty("border", "1px solid #10b981").
			WithCustomProperty("--focus-border", "#10b981").
			WithCustomProperty("--focus-shadow", "0 0 0 3px rgba(16, 185, 129, 0.1)")
	case "disabled":
		styles.WithBackgroundColor("#f9fafb").
			WithColor("#6b7280").
			WithCustomProperty("border", "1px solid #d1d5db").
			WithCustomProperty("cursor", "not-allowed")
	default:
		// Default input
		styles.WithBackgroundColor("#ffffff").
			WithCustomProperty("border", "1px solid #d1d5db").
			WithCustomProperty("--focus-border", "#3b82f6").
			WithCustomProperty("--focus-shadow", "0 0 0 3px rgba(59, 130, 246, 0.1)")
	}
	
	return styles
}

// CardStyles creates styled card component
func (f *Factory) CardStyles(elevation string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	// Base card styles
	styles.WithBackgroundColor("#ffffff").
		WithBorderRadius("0.5rem").
		WithCustomProperty("border", "1px solid #e5e7eb")
	
	// Elevation-specific shadows
	switch elevation {
	case "sm":
		styles.WithBoxShadow("0 1px 2px 0 rgba(0, 0, 0, 0.05)")
	case "md":
		styles.WithBoxShadow("0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)")
	case "lg":
		styles.WithBoxShadow("0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)")
	case "xl":
		styles.WithBoxShadow("0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)")
	case "none":
		// No shadow
	default:
		// Default elevation
		styles.WithBoxShadow("0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)")
	}
	
	return styles
}

// ContainerStyles creates styled container
func (f *Factory) ContainerStyles(maxWidth string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	// Base container styles
	styles.WithWidth("100%").
		WithCustomProperty("margin", "0 auto").
		WithPadding("0 1rem")
	
	// Max width variants
	switch maxWidth {
	case "sm":
		styles.WithMaxWidth("640px")
	case "md":
		styles.WithMaxWidth("768px")
	case "lg":
		styles.WithMaxWidth("1024px")
	case "xl":
		styles.WithMaxWidth("1280px")
	case "2xl":
		styles.WithMaxWidth("1536px")
	case "full":
		styles.WithMaxWidth("100%")
	default:
		styles.WithMaxWidth("1200px")
	}
	
	return styles
}

// FlexStyles creates flexbox layout styles
func (f *Factory) FlexStyles(direction, justify, align string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	styles.WithDisplay("flex")
	
	if direction != "" {
		styles.WithFlexDirection(direction)
	}
	
	if justify != "" {
		styles.WithJustifyContent(justify)
	}
	
	if align != "" {
		styles.WithAlignItems(align)
	}
	
	return styles
}

// GridStyles creates CSS grid layout styles
func (f *Factory) GridStyles(columns, gap string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	styles.WithDisplay("grid")
	
	if columns != "" {
		styles.WithGridTemplateColumns(columns)
	}
	
	if gap != "" {
		styles.WithGridGap(gap)
	}
	
	return styles
}

// ModalStyles creates styled modal overlay
func (f *Factory) ModalStyles(size string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	// Base modal styles
	styles.WithPosition("fixed").
		WithCustomProperty("top", "0").
		WithCustomProperty("left", "0").
		WithWidth("100%").
		WithHeight("100%").
		WithBackgroundColor("rgba(0, 0, 0, 0.5)").
		WithDisplay("flex").
		WithAlignItems("center").
		WithJustifyContent("center").
		WithZIndex("1000")
	
	return styles
}

// ModalContentStyles creates styled modal content
func (f *Factory) ModalContentStyles(size string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	// Base modal content styles
	styles.WithBackgroundColor("#ffffff").
		WithBorderRadius("0.5rem").
		WithBoxShadow("0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)").
		WithPadding("1.5rem").
		WithMaxHeight("90vh").
		WithOverflow("auto")
	
	// Size variants
	switch size {
	case "sm":
		styles.WithMaxWidth("400px")
	case "md":
		styles.WithMaxWidth("500px")
	case "lg":
		styles.WithMaxWidth("800px")
	case "xl":
		styles.WithMaxWidth("1200px")
	case "full":
		styles.WithMaxWidth("95vw")
	default:
		styles.WithMaxWidth("600px")
	}
	
	return styles
}

// TableStyles creates styled table
func (f *Factory) TableStyles(variant string) *Styles {
	styles := NewStyles(f.schemaDir)
	
	// Base table styles
	styles.WithWidth("100%").
		WithCustomProperty("border-collapse", "collapse").
		WithBackgroundColor("#ffffff")
	
	switch variant {
	case "bordered":
		styles.WithCustomProperty("border", "1px solid #e5e7eb")
	case "striped":
		styles.WithCustomProperty("--stripe-bg", "#f9fafb")
	case "hover":
		styles.WithCustomProperty("--hover-bg", "#f3f4f6")
	default:
		// Default table
	}
	
	return styles
}

// ============================================================================
// RESPONSIVE UTILITIES
// ============================================================================

// ResponsiveStyles creates responsive breakpoint styles
func (f *Factory) ResponsiveStyles() *ResponsiveStyleBuilder {
	return &ResponsiveStyleBuilder{
		factory: f,
		styles:  NewStyles(f.schemaDir),
	}
}

// ResponsiveStyleBuilder provides a fluent API for responsive styles
type ResponsiveStyleBuilder struct {
	factory *Factory
	styles  *Styles
}

// Base sets the base styles (mobile-first)
func (r *ResponsiveStyleBuilder) Base(styles *Styles) *ResponsiveStyleBuilder {
	// Merge base styles
	r.styles = styles
	return r
}

// SM sets styles for small screens and up (640px+)
func (r *ResponsiveStyleBuilder) SM(modifier func(*Styles)) *ResponsiveStyleBuilder {
	// In a real implementation, this would add media query styles
	// For now, we'll add them as custom properties
	smStyles := NewStyles(r.factory.schemaDir)
	modifier(smStyles)
	
	// Add as custom properties with sm prefix
	properties := smStyles.getAllProperties()
	for prop, value := range properties {
		if value != "" {
			r.styles.WithCustomProperty(fmt.Sprintf("--sm-%s", prop), value)
		}
	}
	
	return r
}

// MD sets styles for medium screens and up (768px+)
func (r *ResponsiveStyleBuilder) MD(modifier func(*Styles)) *ResponsiveStyleBuilder {
	mdStyles := NewStyles(r.factory.schemaDir)
	modifier(mdStyles)
	
	properties := mdStyles.getAllProperties()
	for prop, value := range properties {
		if value != "" {
			r.styles.WithCustomProperty(fmt.Sprintf("--md-%s", prop), value)
		}
	}
	
	return r
}

// LG sets styles for large screens and up (1024px+)
func (r *ResponsiveStyleBuilder) LG(modifier func(*Styles)) *ResponsiveStyleBuilder {
	lgStyles := NewStyles(r.factory.schemaDir)
	modifier(lgStyles)
	
	properties := lgStyles.getAllProperties()
	for prop, value := range properties {
		if value != "" {
			r.styles.WithCustomProperty(fmt.Sprintf("--lg-%s", prop), value)
		}
	}
	
	return r
}

// XL sets styles for extra large screens and up (1280px+)
func (r *ResponsiveStyleBuilder) XL(modifier func(*Styles)) *ResponsiveStyleBuilder {
	xlStyles := NewStyles(r.factory.schemaDir)
	modifier(xlStyles)
	
	properties := xlStyles.getAllProperties()
	for prop, value := range properties {
		if value != "" {
			r.styles.WithCustomProperty(fmt.Sprintf("--xl-%s", prop), value)
		}
	}
	
	return r
}

// Build returns the final responsive styles
func (r *ResponsiveStyleBuilder) Build() *Styles {
	return r.styles
}

// ============================================================================
// THEME UTILITIES
// ============================================================================

// ThemeStyles creates theme-aware styles
func (f *Factory) ThemeStyles(theme string) *ThemeStyleBuilder {
	return &ThemeStyleBuilder{
		factory: f,
		theme:   theme,
		styles:  NewStyles(f.schemaDir),
	}
}

// ThemeStyleBuilder provides theme-aware styling
type ThemeStyleBuilder struct {
	factory *Factory
	theme   string
	styles  *Styles
}

// Primary sets primary color styles
func (t *ThemeStyleBuilder) Primary(lightColor, darkColor string) *ThemeStyleBuilder {
	if t.theme == "dark" {
		t.styles.WithColor(darkColor)
	} else {
		t.styles.WithColor(lightColor)
	}
	return t
}

// Background sets background color styles
func (t *ThemeStyleBuilder) Background(lightBg, darkBg string) *ThemeStyleBuilder {
	if t.theme == "dark" {
		t.styles.WithBackgroundColor(darkBg)
	} else {
		t.styles.WithBackgroundColor(lightBg)
	}
	return t
}

// Border sets border color styles
func (t *ThemeStyleBuilder) Border(lightBorder, darkBorder string) *ThemeStyleBuilder {
	if t.theme == "dark" {
		t.styles.WithCustomProperty("border-color", darkBorder)
	} else {
		t.styles.WithCustomProperty("border-color", lightBorder)
	}
	return t
}

// Build returns the theme-aware styles
func (t *ThemeStyleBuilder) Build() *Styles {
	return t.styles
}