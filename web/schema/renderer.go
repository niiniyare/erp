package schema

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
)

// FieldRenderer interface for rendering different field types
type FieldRenderer interface {
	Render(ctx context.Context, field *Field, value any, errors []string) (string, error)
	SupportsFieldType(fieldType FieldType) bool
	GetRequiredAssets() []string
}

// RendererRegistry manages field renderers
type RendererRegistry struct {
	renderers map[FieldType]FieldRenderer
	tokens    TokenResolver
}

// NewRendererRegistry creates a new renderer registry
func NewRendererRegistry(tokens TokenResolver) *RendererRegistry {
	registry := &RendererRegistry{
		renderers: make(map[FieldType]FieldRenderer),
		tokens:    tokens,
	}

	// Register default renderers
	registry.registerDefaultRenderers()
	return registry
}

// RegisterRenderer registers a renderer for a field type
func (r *RendererRegistry) RegisterRenderer(fieldType FieldType, renderer FieldRenderer) {
	r.renderers[fieldType] = renderer
}

// RenderField renders a field using the appropriate renderer
func (r *RendererRegistry) RenderField(ctx context.Context, field *Field, value any, errors []string) (string, error) {
	renderer, exists := r.renderers[field.Type]
	if !exists {
		return "", NewRenderError("renderer_not_found", fmt.Sprintf("no renderer found for field type: %s", field.Type))
	}

	return renderer.Render(ctx, field, value, errors)
}

// GetFieldRenderer returns the renderer for a field type
func (r *RendererRegistry) GetFieldRenderer(fieldType FieldType) (FieldRenderer, bool) {
	renderer, exists := r.renderers[fieldType]
	return renderer, exists
}

// GetAllRequiredAssets returns all required assets from all renderers
func (r *RendererRegistry) GetAllRequiredAssets() []string {
	var assets []string
	seen := make(map[string]bool)

	for _, renderer := range r.renderers {
		for _, asset := range renderer.GetRequiredAssets() {
			if !seen[asset] {
				assets = append(assets, asset)
				seen[asset] = true
			}
		}
	}

	return assets
}

// BaseRenderer provides common functionality for all renderers
type BaseRenderer struct {
	tokens TokenResolver
}

// NewBaseRenderer creates a new base renderer
func NewBaseRenderer(tokens TokenResolver) *BaseRenderer {
	return &BaseRenderer{tokens: tokens}
}

// RenderContainer renders the field container with label, validation, etc.
func (br *BaseRenderer) RenderContainer(field *Field, inputHTML string, errors []string) string {
	var parts []string

	// Field wrapper with error state
	classes := []string{"field-group"}
	if len(errors) > 0 {
		classes = append(classes, "field-group--error")
	}
	if field.Required {
		classes = append(classes, "field-group--required")
	}

	// Apply design tokens
	style := br.resolveFieldStyle(field)

	parts = append(parts, fmt.Sprintf(`<div class="%s" style="%s">`, strings.Join(classes, " "), style))

	// Label
	if field.Label != "" {
		labelClass := "field-label"
		if field.Required {
			labelClass += " field-label--required"
		}

		ariaLabel := field.Label
		if field.Required {
			ariaLabel += " (required)"
		}

		parts = append(parts, fmt.Sprintf(
			`<label for="%s" class="%s" aria-label="%s">%s</label>`,
			html.EscapeString(field.Name),
			labelClass,
			html.EscapeString(ariaLabel),
			html.EscapeString(field.Label),
		))
	}

	// Description
	if field.Description != "" {
		parts = append(parts, fmt.Sprintf(
			`<div class="field-description" id="%s-description">%s</div>`,
			html.EscapeString(field.Name),
			html.EscapeString(field.Description),
		))
	}

	// Input
	parts = append(parts, inputHTML)

	// Validation errors
	if len(errors) > 0 {
		parts = append(parts, `<div class="field-errors" role="alert">`)
		for _, err := range errors {
			parts = append(parts, fmt.Sprintf(`<span class="field-error">%s</span>`, html.EscapeString(err)))
		}
		parts = append(parts, `</div>`)
	}

	// Help text (more detailed help)
	if field.Help != "" {
		parts = append(parts, fmt.Sprintf(
			`<div class="field-help" id="%s-help">%s</div>`,
			html.EscapeString(field.Name),
			html.EscapeString(field.Help),
		))
	}

	parts = append(parts, `</div>`)

	return strings.Join(parts, "\n")
}

// resolveFieldStyle resolves design tokens to CSS
func (br *BaseRenderer) resolveFieldStyle(field *Field) string {
	if br.tokens == nil || field.Style == nil {
		return ""
	}

	var styles []string

	// Resolve design tokens
	if field.Style.BackgroundToken != "" {
		if value, err := br.tokens.ResolveToken(field.Style.BackgroundToken); err == nil {
			styles = append(styles, fmt.Sprintf("background: %s", value))
		}
	} else if field.Style.Background != "" {
		styles = append(styles, fmt.Sprintf("background: %s", field.Style.Background))
	}

	if field.Style.ColorToken != "" {
		if value, err := br.tokens.ResolveToken(field.Style.ColorToken); err == nil {
			styles = append(styles, fmt.Sprintf("color: %s", value))
		}
	} else if field.Style.Color != "" {
		styles = append(styles, fmt.Sprintf("color: %s", field.Style.Color))
	}

	if field.Style.SpacingToken != "" {
		if value, err := br.tokens.ResolveToken(field.Style.SpacingToken); err == nil {
			styles = append(styles, fmt.Sprintf("padding: %s", value))
		}
	}

	// Add other style properties
	if field.Style.Border != "" {
		styles = append(styles, fmt.Sprintf("border: %s", field.Style.Border))
	}
	if field.Style.Padding != "" {
		styles = append(styles, fmt.Sprintf("padding: %s", field.Style.Padding))
	}
	if field.Style.Margin != "" {
		styles = append(styles, fmt.Sprintf("margin: %s", field.Style.Margin))
	}

	return strings.Join(styles, "; ")
}

// buildInputAttributes builds common input attributes
func (br *BaseRenderer) buildInputAttributes(field *Field, value any) map[string]string {
	attrs := map[string]string{
		"id":   field.Name,
		"name": field.Name,
	}

	// Add class
	classes := []string{"field-input"}
	if field.Type != "" {
		classes = append(classes, fmt.Sprintf("field-input--%s", string(field.Type)))
	}
	attrs["class"] = strings.Join(classes, " ")

	// Required
	if field.Required {
		attrs["required"] = "required"
		attrs["aria-required"] = "true"
	}

	// Disabled
	if field.Disabled {
		attrs["disabled"] = "disabled"
		attrs["aria-disabled"] = "true"
	}

	// Readonly
	if field.Readonly {
		attrs["readonly"] = "readonly"
		attrs["aria-readonly"] = "true"
	}

	// Placeholder
	if field.Placeholder != "" {
		attrs["placeholder"] = field.Placeholder
	}

	// Value
	if value != nil {
		attrs["value"] = fmt.Sprintf("%v", value)
	} else if field.Default != nil {
		attrs["value"] = fmt.Sprintf("%v", field.Default)
	}

	// ARIA attributes
	var describedByParts []string
	if field.Description != "" {
		describedByParts = append(describedByParts, field.Name+"-description")
	}
	if field.Help != "" {
		describedByParts = append(describedByParts, field.Name+"-help")
	}
	if len(describedByParts) > 0 {
		attrs["aria-describedby"] = strings.Join(describedByParts, " ")
	}

	// Validation attributes
	if field.Validation != nil {
		if field.Validation.MinLength != nil {
			attrs["minlength"] = strconv.Itoa(*field.Validation.MinLength)
		}
		if field.Validation.MaxLength != nil {
			attrs["maxlength"] = strconv.Itoa(*field.Validation.MaxLength)
		}
		if field.Validation.Min != nil {
			attrs["min"] = fmt.Sprintf("%g", *field.Validation.Min)
		}
		if field.Validation.Max != nil {
			attrs["max"] = fmt.Sprintf("%g", *field.Validation.Max)
		}
		if field.Validation.Pattern != "" {
			attrs["pattern"] = field.Validation.Pattern
		}
	}

	// HTMX attributes
	if field.HTMX != nil {
		if field.HTMX.Get != "" {
			attrs["hx-get"] = field.HTMX.Get
		}
		if field.HTMX.Post != "" {
			attrs["hx-post"] = field.HTMX.Post
		}
		if field.HTMX.Target != "" {
			attrs["hx-target"] = field.HTMX.Target
		}
		if field.HTMX.Swap != "" {
			attrs["hx-swap"] = field.HTMX.Swap
		}
		if field.HTMX.Trigger != "" {
			attrs["hx-trigger"] = field.HTMX.Trigger
		}
		if field.HTMX.Delay != "" {
			attrs["hx-delay"] = field.HTMX.Delay
		}
		if field.HTMX.Include != "" {
			attrs["hx-include"] = field.HTMX.Include
		}
	}

	// Alpine.js attributes
	if field.Alpine != nil {
		if field.Alpine.XModel != "" {
			attrs["x-model"] = field.Alpine.XModel
		}
		if field.Alpine.XShow != "" {
			attrs["x-show"] = field.Alpine.XShow
		}
		if field.Alpine.XIf != "" {
			attrs["x-if"] = field.Alpine.XIf
		}
		if field.Alpine.XText != "" {
			attrs["x-text"] = field.Alpine.XText
		}
		if field.Alpine.XRef != "" {
			attrs["x-ref"] = field.Alpine.XRef
		}

		// Add x-bind attributes
		for key, value := range field.Alpine.XBind {
			attrs[fmt.Sprintf("x-bind:%s", key)] = value
		}

		// Add x-on attributes
		for event, handler := range field.Alpine.XOn {
			attrs[fmt.Sprintf("x-on:%s", event)] = handler
		}
	}

	// Event handlers
	if field.Events != nil {
		if field.Events.OnChange != "" {
			attrs["onchange"] = field.Events.OnChange
		}
		if field.Events.OnFocus != "" {
			attrs["onfocus"] = field.Events.OnFocus
		}
		if field.Events.OnBlur != "" {
			attrs["onblur"] = field.Events.OnBlur
		}
		if field.Events.OnInput != "" {
			attrs["oninput"] = field.Events.OnInput
		}
		if field.Events.OnKeyPress != "" {
			attrs["onkeypress"] = field.Events.OnKeyPress
		}
		if field.Events.OnKeyUp != "" {
			attrs["onkeyup"] = field.Events.OnKeyUp
		}
		if field.Events.OnKeyDown != "" {
			attrs["onkeydown"] = field.Events.OnKeyDown
		}
		if field.Events.OnClick != "" {
			attrs["onclick"] = field.Events.OnClick
		}
	}

	return attrs
}

// attributesToString converts attributes map to HTML string
func (br *BaseRenderer) attributesToString(attrs map[string]string) string {
	var parts []string
	for key, value := range attrs {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, html.EscapeString(value)))
	}
	return strings.Join(parts, " ")
}

// TokenResolver interface for resolving design tokens
type TokenResolver interface {
	ResolveToken(token string) (string, error)
	HasToken(token string) bool
	GetTokens() map[string]string
}

// DefaultTokenResolver provides basic token resolution
type DefaultTokenResolver struct {
	tokens map[string]string
}

// NewDefaultTokenResolver creates a default token resolver
func NewDefaultTokenResolver() *DefaultTokenResolver {
	return &DefaultTokenResolver{
		tokens: getDefaultTokens(),
	}
}

// ResolveToken resolves a design token to its CSS value
func (tr *DefaultTokenResolver) ResolveToken(token string) (string, error) {
	if value, exists := tr.tokens[token]; exists {
		return value, nil
	}
	return "", NewRenderError("token_not_found", fmt.Sprintf("design token not found: %s", token))
}

// HasToken checks if a token exists
func (tr *DefaultTokenResolver) HasToken(token string) bool {
	_, exists := tr.tokens[token]
	return exists
}

// GetTokens returns all available tokens
func (tr *DefaultTokenResolver) GetTokens() map[string]string {
	return tr.tokens
}

// AddToken adds or updates a token
func (tr *DefaultTokenResolver) AddToken(token, value string) {
	tr.tokens[token] = value
}

// getDefaultTokens returns the default design tokens
func getDefaultTokens() map[string]string {
	return map[string]string{
		// Input tokens
		"input.background":          "var(--color-white)",
		"input.text":                "var(--color-gray-900)",
		"input.border":              "var(--color-gray-300)",
		"input.border.focus":        "var(--color-blue-500)",
		"input.placeholder":         "var(--color-gray-500)",
		"input.disabled.background": "var(--color-gray-100)",
		"input.disabled.text":       "var(--color-gray-400)",
		"input.error.border":        "var(--color-red-500)",
		"input.error.background":    "var(--color-red-50)",

		// Button tokens
		"button.primary.background":       "var(--color-blue-600)",
		"button.primary.text":             "var(--color-white)",
		"button.primary.border":           "var(--color-blue-600)",
		"button.primary.hover.background": "var(--color-blue-700)",
		"button.secondary.background":     "var(--color-white)",
		"button.secondary.text":           "var(--color-gray-700)",
		"button.secondary.border":         "var(--color-gray-300)",
		"button.outline.background":       "transparent",
		"button.outline.text":             "var(--color-blue-600)",
		"button.outline.border":           "var(--color-blue-600)",

		// Select tokens
		"select.background":              "var(--color-white)",
		"select.text":                    "var(--color-gray-900)",
		"select.border":                  "var(--color-gray-300)",
		"select.arrow":                   "var(--color-gray-500)",
		"select.option.background":       "var(--color-white)",
		"select.option.text":             "var(--color-gray-900)",
		"select.option.hover.background": "var(--color-blue-50)",

		// Spacing tokens
		"spacing.xs":  "0.25rem",
		"spacing.sm":  "0.5rem",
		"spacing.md":  "0.75rem",
		"spacing.lg":  "1rem",
		"spacing.xl":  "1.25rem",
		"spacing.2xl": "1.5rem",
		"spacing.3xl": "2rem",

		// Typography tokens
		"font.size.xs":         "0.75rem",
		"font.size.sm":         "0.875rem",
		"font.size.md":         "1rem",
		"font.size.lg":         "1.125rem",
		"font.size.xl":         "1.25rem",
		"font.weight.normal":   "400",
		"font.weight.medium":   "500",
		"font.weight.semibold": "600",
		"font.weight.bold":     "700",

		// Border radius tokens
		"border.radius.sm": "0.125rem",
		"border.radius.md": "0.25rem",
		"border.radius.lg": "0.5rem",
		"border.radius.xl": "0.75rem",

		// Shadow tokens
		"shadow.sm": "0 1px 2px 0 rgba(0, 0, 0, 0.05)",
		"shadow.md": "0 4px 6px -1px rgba(0, 0, 0, 0.1)",
		"shadow.lg": "0 10px 15px -3px rgba(0, 0, 0, 0.1)",
	}
}

// registerDefaultRenderers registers all default field renderers
func (r *RendererRegistry) registerDefaultRenderers() {
	baseRenderer := NewBaseRenderer(r.tokens)

	// Text-based renderers
	r.RegisterRenderer(FieldText, NewTextRenderer(baseRenderer))
	r.RegisterRenderer(FieldEmail, NewEmailRenderer(baseRenderer))
	r.RegisterRenderer(FieldPassword, NewPasswordRenderer(baseRenderer))
	r.RegisterRenderer(FieldURL, NewURLRenderer(baseRenderer))
	r.RegisterRenderer(FieldPhone, NewPhoneRenderer(baseRenderer))
	r.RegisterRenderer(FieldHidden, NewHiddenRenderer(baseRenderer))

	// Number renderers
	r.RegisterRenderer(FieldNumber, NewNumberRenderer(baseRenderer))
	r.RegisterRenderer(FieldCurrency, NewCurrencyRenderer(baseRenderer))
	r.RegisterRenderer(FieldSlider, NewSliderRenderer(baseRenderer))
	r.RegisterRenderer(FieldRating, NewRatingRenderer(baseRenderer))

	// Date/Time renderers
	r.RegisterRenderer(FieldDate, NewDateRenderer(baseRenderer))
	r.RegisterRenderer(FieldTime, NewTimeRenderer(baseRenderer))
	r.RegisterRenderer(FieldDateTime, NewDateTimeRenderer(baseRenderer))
	r.RegisterRenderer(FieldDateRange, NewDateRangeRenderer(baseRenderer))

	// Text area renderers
	r.RegisterRenderer(FieldTextarea, NewTextareaRenderer(baseRenderer))
	r.RegisterRenderer(FieldRichText, NewRichTextRenderer(baseRenderer))
	r.RegisterRenderer(FieldCode, NewCodeRenderer(baseRenderer))
	r.RegisterRenderer(FieldJSON, NewJSONRenderer(baseRenderer))

	// Selection renderers
	r.RegisterRenderer(FieldSelect, NewSelectRenderer(baseRenderer))
	r.RegisterRenderer(FieldMultiSelect, NewMultiSelectRenderer(baseRenderer))
	r.RegisterRenderer(FieldRadio, NewRadioRenderer(baseRenderer))
	r.RegisterRenderer(FieldCheckbox, NewCheckboxRenderer(baseRenderer))
	r.RegisterRenderer(FieldSwitch, NewSwitchRenderer(baseRenderer))
	r.RegisterRenderer(FieldTreeSelect, NewTreeSelectRenderer(baseRenderer))
	r.RegisterRenderer(FieldCascader, NewCascaderRenderer(baseRenderer))
	r.RegisterRenderer(FieldTransfer, NewTransferRenderer(baseRenderer))

	// File renderers
	r.RegisterRenderer(FieldFile, NewFileRenderer(baseRenderer))
	r.RegisterRenderer(FieldImage, NewImageRenderer(baseRenderer))
	r.RegisterRenderer(FieldSignature, NewSignatureRenderer(baseRenderer))

	// Special renderers
	r.RegisterRenderer(FieldColor, NewColorRenderer(baseRenderer))
	r.RegisterRenderer(FieldTags, NewTagsRenderer(baseRenderer))
	r.RegisterRenderer(FieldLocation, NewLocationRenderer(baseRenderer))
	r.RegisterRenderer(FieldRelation, NewRelationRenderer(baseRenderer))
	r.RegisterRenderer(FieldAutoComplete, NewAutoCompleteRenderer(baseRenderer))

	// Display renderers
	r.RegisterRenderer(FieldDisplay, NewDisplayRenderer(baseRenderer))
	r.RegisterRenderer(FieldDivider, NewDividerRenderer(baseRenderer))
	r.RegisterRenderer(FieldHTML, NewHTMLRenderer(baseRenderer))
}

