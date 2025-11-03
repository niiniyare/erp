package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

// TemplRenderer provides templ template integration for the schema system
type TemplRenderer struct {
	registry *RendererRegistry
	tokens   TokenResolver
}

// NewTemplRenderer creates a new templ-based renderer
func NewTemplRenderer(registry *RendererRegistry, tokens TokenResolver) *TemplRenderer {
	if registry == nil {
		registry = NewRendererRegistry(tokens)
	}
	return &TemplRenderer{
		registry: registry,
		tokens:   tokens,
	}
}

// RenderForm implements FormRenderer interface using templ templates
func (tr *TemplRenderer) RenderForm(ctx context.Context, schema *Schema, data map[string]any) (string, error) {
	return tr.RenderFormWithErrors(ctx, schema, data, nil)
}

// RenderFormWithErrors renders a complete form with error handling using templ
func (tr *TemplRenderer) RenderFormWithErrors(ctx context.Context, schema *Schema, data map[string]any, errors map[string][]string) (string, error) {
	if schema.Type != TypeForm {
		return "", NewRenderError("invalid_schema_type", "schema must be of type 'form'")
	}

	// Use templ to render the form
	component := FormTemplate(schema, data, errors, tr.registry, tr.tokens)

	// Render the templ component to string
	var builder strings.Builder
	err := component.Render(ctx, &builder)
	if err != nil {
		return "", NewRenderError("templ_render_failed", fmt.Sprintf("failed to render templ template: %v", err))
	}

	return builder.String(), nil
}

// RenderField implements FormRenderer interface using templ templates
func (tr *TemplRenderer) RenderField(ctx context.Context, field *Field, value any) (string, error) {
	// Use the registry to render individual fields
	return tr.registry.RenderField(ctx, field, value, nil)
}

// RenderAction implements FormRenderer interface using templ templates
func (tr *TemplRenderer) RenderAction(ctx context.Context, action *Action) (string, error) {
	// Render action button using templ
	component := ActionTemplate(action, tr.tokens)

	var builder strings.Builder
	err := component.Render(ctx, &builder)
	if err != nil {
		return "", NewRenderError("templ_render_failed", fmt.Sprintf("failed to render action template: %v", err))
	}

	return builder.String(), nil
}

// RenderFieldWithTempl renders a field using templ templates
func (tr *TemplRenderer) RenderFieldWithTempl(ctx context.Context, field *Field, value any, errors []string) (templ.Component, error) {
	return FieldTemplate(field, value, errors, tr.tokens), nil
}

// RenderFormAsComponent returns the form as a templ.Component for composition
func (tr *TemplRenderer) RenderFormAsComponent(schema *Schema, data map[string]any, errors map[string][]string) templ.Component {
	return FormTemplate(schema, data, errors, tr.registry, tr.tokens)
}

// RenderActionAsComponent returns the action as a templ.Component
func (tr *TemplRenderer) RenderActionAsComponent(action *Action) templ.Component {
	return ActionTemplate(action, tr.tokens)
}

// RenderFieldAsComponent returns the field as a templ.Component
func (tr *TemplRenderer) RenderFieldAsComponent(field *Field, value any, errors []string) templ.Component {
	return FieldTemplate(field, value, errors, tr.tokens)
}

// Helper method to resolve design tokens
func (tr *TemplRenderer) ResolveToken(token string) string {
	if tr.tokens != nil {
		if value, err := tr.tokens.ResolveToken(token); err == nil {
			return value
		}
	}
	return ""
}

// Helper method to generate CSS variables from design tokens
func (tr *TemplRenderer) GenerateTokenCSS(schema *Schema) string {
	if tr.tokens == nil {
		return ""
	}

	var cssVars []string

	// Generate common design token CSS variables
	tokens := tr.tokens.GetTokens()
	for token, value := range tokens {
		// Convert token name to CSS variable format
		cssVar := fmt.Sprintf("--token-%s", strings.ReplaceAll(token, ".", "-"))
		cssVars = append(cssVars, fmt.Sprintf("%s: %s", cssVar, value))
	}

	if len(cssVars) > 0 {
		return fmt.Sprintf(":root { %s }", strings.Join(cssVars, "; "))
	}

	return ""
}

// GetRequiredAssets returns all required CSS and JS assets
func (tr *TemplRenderer) GetRequiredAssets() []string {
	if tr.registry != nil {
		return tr.registry.GetAllRequiredAssets()
	}
	return []string{}
}
