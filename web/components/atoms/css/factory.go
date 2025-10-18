// Package css provides component styling factories for atomic components
package css

// Factory provides convenient methods for creating styled atomic components
type Factory struct {
	theme *Theme
}

// NewFactory creates a new CSS factory with default theme
func NewFactory() *Factory {
	return &Factory{
		theme: DefaultTheme(),
	}
}

// NewFactoryWithTheme creates a CSS factory with custom theme
func NewFactoryWithTheme(theme *Theme) *Factory {
	return &Factory{
		theme: theme,
	}
}

// ============================================================================
// COMPONENT STYLE FACTORIES
// ============================================================================

// ButtonStyles creates styled button with various presets
func (f *Factory) ButtonStyles(variant string) *Styles {
	styles := NewStyles()

	// Base button styles
	styles.WithDisplay("inline-flex").
		WithAlignItems("center").
		WithJustifyContent("center").
		WithPadding("0.5rem 1rem").
		WithBorderRadius("0.375rem").
		WithFontWeight("500").
		WithTransition("all 0.2s ease-in-out").
		WithCustomProperty("cursor", "pointer").
		WithCustomProperty("user-select", "none").
		WithCustomProperty("text-decoration", "none").
		WithCustomProperty("outline", "none")

	// Variant-specific styles
	switch variant {
	case "primary":
		styles.WithBackgroundColor(f.theme.Colors["primary"]).
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid "+f.theme.Colors["primary"]).
			WithBoxShadow(f.theme.Shadows["sm"])
	case "secondary":
		styles.WithBackgroundColor(f.theme.Colors["secondary-light"]).
			WithColor(f.theme.Colors["secondary-dark"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["secondary"])
	case "success":
		styles.WithBackgroundColor(f.theme.Colors["success"]).
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid "+f.theme.Colors["success"])
	case "danger", "error":
		styles.WithBackgroundColor(f.theme.Colors["error"]).
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid "+f.theme.Colors["error"])
	case "warning":
		styles.WithBackgroundColor(f.theme.Colors["warning"]).
			WithColor("#ffffff").
			WithCustomProperty("border", "1px solid "+f.theme.Colors["warning"])
	case "outline":
		styles.WithBackgroundColor("transparent").
			WithColor(f.theme.Colors["primary"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["primary"])
	case "ghost":
		styles.WithBackgroundColor("transparent").
			WithColor(f.theme.Colors["primary"]).
			WithCustomProperty("border", "1px solid transparent")
	default:
		// Default button
		styles.WithBackgroundColor(f.theme.Colors["neutral-light"]).
			WithColor(f.theme.Colors["neutral-dark"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral"])
	}

	return styles
}

// InputStyles creates styled form input
func (f *Factory) InputStyles(state string) *Styles {
	styles := NewStyles()

	// Base input styles
	styles.WithDisplay("block").
		WithWidth("100%").
		WithPadding("0.5rem 0.75rem").
		WithBorderRadius("0.375rem").
		WithFontSize("1rem").
		WithTransition("border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out").
		WithBackgroundColor("#ffffff")

	// State-specific styles
	switch state {
	case "error":
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["error"]).
			WithColor(f.theme.Colors["error"])
	case "success":
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["success"]).
			WithColor(f.theme.Colors["success"])
	case "disabled":
		styles.WithBackgroundColor(f.theme.Colors["neutral-lighter"]).
			WithColor(f.theme.Colors["neutral"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"]).
			WithCustomProperty("cursor", "not-allowed")
	default:
		// Default input
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"]).
			WithColor(f.theme.Colors["neutral-dark"])
	}

	return styles
}

// CheckboxStyles creates styled checkbox
func (f *Factory) CheckboxStyles(variant string) *Styles {
	styles := NewStyles()

	// Base checkbox styles
	styles.WithDisplay("inline-block").
		WithWidth("1rem").
		WithHeight("1rem").
		WithBorderRadius("0.25rem").
		WithTransition("all 0.2s ease-in-out").
		WithCustomProperty("cursor", "pointer")

	// Variant-specific styles
	switch variant {
	case "success":
		styles.WithBackgroundColor(f.theme.Colors["success"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["success"])
	case "error":
		styles.WithBackgroundColor(f.theme.Colors["error"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["error"])
	default:
		styles.WithBackgroundColor("#ffffff").
			WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"])
	}

	return styles
}

// RadioStyles creates styled radio button
func (f *Factory) RadioStyles(variant string) *Styles {
	styles := NewStyles()

	// Base radio styles
	styles.WithDisplay("inline-block").
		WithWidth("1rem").
		WithHeight("1rem").
		WithBorderRadius("50%").
		WithTransition("all 0.2s ease-in-out").
		WithCustomProperty("cursor", "pointer")

	// Variant-specific styles
	switch variant {
	case "success":
		styles.WithBackgroundColor(f.theme.Colors["success"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["success"])
	case "error":
		styles.WithBackgroundColor(f.theme.Colors["error"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["error"])
	default:
		styles.WithBackgroundColor("#ffffff").
			WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"])
	}

	return styles
}

// SelectStyles creates styled select dropdown
func (f *Factory) SelectStyles(state string) *Styles {
	styles := NewStyles()

	// Base select styles
	styles.WithDisplay("block").
		WithWidth("100%").
		WithPadding("0.5rem 2rem 0.5rem 0.75rem").
		WithBorderRadius("0.375rem").
		WithFontSize("1rem").
		WithTransition("border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out").
		WithBackgroundColor("#ffffff").
		WithCustomProperty("cursor", "pointer")

	// State-specific styles
	switch state {
	case "error":
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["error"])
	case "success":
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["success"])
	case "disabled":
		styles.WithBackgroundColor(f.theme.Colors["neutral-lighter"]).
			WithColor(f.theme.Colors["neutral"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"]).
			WithCustomProperty("cursor", "not-allowed")
	default:
		// Default select
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"]).
			WithColor(f.theme.Colors["neutral-dark"])
	}

	return styles
}

// TextareaStyles creates styled textarea
func (f *Factory) TextareaStyles(state string) *Styles {
	styles := NewStyles()

	// Base textarea styles
	styles.WithDisplay("block").
		WithWidth("100%").
		WithPadding("0.5rem 0.75rem").
		WithBorderRadius("0.375rem").
		WithFontSize("1rem").
		WithTransition("border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out").
		WithBackgroundColor("#ffffff").
		WithCustomProperty("resize", "vertical").
		WithCustomProperty("min-height", "4rem")

	// State-specific styles
	switch state {
	case "error":
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["error"])
	case "success":
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["success"])
	case "disabled":
		styles.WithBackgroundColor(f.theme.Colors["neutral-lighter"]).
			WithColor(f.theme.Colors["neutral"]).
			WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"]).
			WithCustomProperty("cursor", "not-allowed")
	default:
		// Default textarea
		styles.WithCustomProperty("border", "1px solid "+f.theme.Colors["neutral-light"]).
			WithColor(f.theme.Colors["neutral-dark"])
	}

	return styles
}

// IconStyles creates styled icon
func (f *Factory) IconStyles(size string) *Styles {
	styles := NewStyles()

	// Base icon styles
	styles.WithDisplay("inline-block").
		WithCustomProperty("vertical-align", "middle")

	// Size variants
	switch size {
	case "xs":
		styles.WithWidth("0.75rem").WithHeight("0.75rem")
	case "sm":
		styles.WithWidth("1rem").WithHeight("1rem")
	case "md":
		styles.WithWidth("1.25rem").WithHeight("1.25rem")
	case "lg":
		styles.WithWidth("1.5rem").WithHeight("1.5rem")
	case "xl":
		styles.WithWidth("2rem").WithHeight("2rem")
	default:
		// Default size (md)
		styles.WithWidth("1.25rem").WithHeight("1.25rem")
	}

	return styles
}

// SpinnerStyles creates styled spinner
func (f *Factory) SpinnerStyles(variant string) *Styles {
	styles := NewStyles()

	// Base spinner styles
	styles.WithDisplay("inline-block").
		WithWidth("1rem").
		WithHeight("1rem").
		WithBorderRadius("50%").
		WithCustomProperty("animation", "spin 1s linear infinite")

	// Variant-specific styles
	switch variant {
	case "border":
		styles.WithCustomProperty("border", "2px solid "+f.theme.Colors["neutral-lighter"]).
			WithCustomProperty("border-top", "2px solid "+f.theme.Colors["primary"])
	case "dots":
		styles.WithBackgroundColor(f.theme.Colors["primary"])
	case "pulse":
		styles.WithBackgroundColor(f.theme.Colors["primary"]).
			WithCustomProperty("animation", "pulse 1.5s ease-in-out infinite")
	default:
		// Default border spinner
		styles.WithCustomProperty("border", "2px solid "+f.theme.Colors["neutral-lighter"]).
			WithCustomProperty("border-top", "2px solid "+f.theme.Colors["primary"])
	}

	return styles
}

// ToggleStyles creates styled toggle/switch
func (f *Factory) ToggleStyles(variant string) *Styles {
	styles := NewStyles()

	// Base toggle styles
	styles.WithPosition("relative").
		WithDisplay("inline-block").
		WithWidth("2.5rem").
		WithHeight("1.25rem").
		WithBorderRadius("0.625rem").
		WithTransition("all 0.2s ease-in-out").
		WithCustomProperty("cursor", "pointer")

	// Variant-specific styles
	switch variant {
	case "success":
		styles.WithBackgroundColor(f.theme.Colors["success"])
	case "danger":
		styles.WithBackgroundColor(f.theme.Colors["error"])
	default:
		// Default toggle
		styles.WithBackgroundColor(f.theme.Colors["primary"])
	}

	return styles
}
