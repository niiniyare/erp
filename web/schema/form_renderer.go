package schema

import (
	"context"
	"fmt"
	"html"
	"strings"
)

// DefaultFormRenderer handles rendering complete forms
type DefaultFormRenderer struct {
	registry *RendererRegistry
	tokens   TokenResolver
}

// NewDefaultFormRenderer creates a new form renderer
func NewDefaultFormRenderer(registry *RendererRegistry, tokens TokenResolver) *DefaultFormRenderer {
	return &DefaultFormRenderer{
		registry: registry,
		tokens:   tokens,
	}
}

// buildFormAttributes builds form tag attributes
func (fr *DefaultFormRenderer) buildFormAttributes(schema *Schema) map[string]string {
	attrs := map[string]string{
		"id":     schema.ID,
		"class":  "schema-form",
		"method": "POST",
	}

	// Add form configuration
	if schema.Config != nil {
		if schema.Config.Action != "" {
			attrs["action"] = schema.Config.Action
		}
		if schema.Config.Method != "" {
			attrs["method"] = schema.Config.Method
		}
		if schema.Config.Target != "" {
			attrs["target"] = schema.Config.Target
		}
		if schema.Config.Encoding != "" {
			attrs["enctype"] = schema.Config.Encoding
		}
	}

	// Add CSS classes based on features
	var classes []string
	classes = append(classes, "schema-form")

	if schema.Layout != nil {
		classes = append(classes, fmt.Sprintf("schema-form--columns-%d", schema.Layout.Columns))
		if schema.Layout.Gap != "" {
			classes = append(classes, "schema-form--custom-gap")
		}
	}

	if schema.HasHTMX() {
		classes = append(classes, "schema-form--htmx")
	}

	if schema.HasAlpine() {
		classes = append(classes, "schema-form--alpine")
	}

	attrs["class"] = strings.Join(classes, " ")

	// Add HTMX attributes
	if schema.HTMX != nil && schema.HTMX.Enabled {
		if schema.HTMX.Post != "" {
			attrs["hx-post"] = schema.HTMX.Post
		}
		if schema.HTMX.Get != "" {
			attrs["hx-get"] = schema.HTMX.Get
		}
		if schema.HTMX.Target != "" {
			attrs["hx-target"] = schema.HTMX.Target
		}
		if schema.HTMX.Swap != "" {
			attrs["hx-swap"] = schema.HTMX.Swap
		}
		if schema.HTMX.Trigger != "" {
			attrs["hx-trigger"] = schema.HTMX.Trigger
		}
		if schema.HTMX.Validate {
			attrs["hx-validate"] = "true"
		}
	}

	// Add Alpine.js attributes
	if schema.Alpine != nil && schema.Alpine.Enabled {
		if schema.Alpine.XData != "" {
			attrs["x-data"] = schema.Alpine.XData
		}
		if schema.Alpine.XInit != "" {
			attrs["x-init"] = schema.Alpine.XInit
		}
	}

	// Add event handlers
	if schema.Events != nil {
		if schema.Events.OnSubmit != "" {
			attrs["onsubmit"] = schema.Events.OnSubmit
		}
		if schema.Events.OnMount != "" {
			attrs["data-on-mount"] = schema.Events.OnMount
		}
	}

	// Add design token CSS variables
	if style := fr.resolveFormStyle(schema); style != "" {
		attrs["style"] = style
	}

	return attrs
}

// renderFormHeader renders form title and description
func (fr *DefaultFormRenderer) renderFormHeader(schema *Schema) string {
	if schema.Title == "" && schema.Description == "" {
		return ""
	}

	var parts []string
	parts = append(parts, `<div class="form-header">`)

	if schema.Title != "" {
		parts = append(parts, fmt.Sprintf(`<h1 class="form-title">%s</h1>`, html.EscapeString(schema.Title)))
	}

	if schema.Description != "" {
		parts = append(parts, fmt.Sprintf(`<p class="form-description">%s</p>`, html.EscapeString(schema.Description)))
	}

	parts = append(parts, `</div>`)
	return strings.Join(parts, "\n")
}

// renderCSRFToken renders CSRF protection token
func (fr *DefaultFormRenderer) renderCSRFToken(csrf *CSRF) string {
	tokenField := csrf.TokenField
	if tokenField == "" {
		tokenField = "csrf_token"
	}

	// In production, this would get the actual CSRF token from the security system
	return fmt.Sprintf(`<input type="hidden" name="%s" value="csrf_token_placeholder">`, tokenField)
}

// renderFormContent renders the main form content based on layout
func (fr *DefaultFormRenderer) renderFormContent(ctx context.Context, schema *Schema, data map[string]any, errors map[string][]string) (string, error) {
	if schema.Layout == nil {
		// Simple field-by-field rendering
		return fr.renderFieldList(ctx, schema.Fields, data, errors)
	}

	// Render based on layout type
	switch {
	case len(schema.Layout.Sections) > 0:
		return fr.renderSections(ctx, schema, data, errors)
	case len(schema.Layout.Tabs) > 0:
		return fr.renderTabs(ctx, schema, data, errors)
	case len(schema.Layout.Steps) > 0:
		return fr.renderSteps(ctx, schema, data, errors)
	default:
		return fr.renderFieldList(ctx, schema.Fields, data, errors)
	}
}

// renderFieldList renders a simple list of fields
func (fr *DefaultFormRenderer) renderFieldList(ctx context.Context, fields []Field, data map[string]any, errors map[string][]string) (string, error) {
	var fieldHTMLs []string

	for _, field := range fields {
		value := data[field.Name]
		fieldErrors := errors[field.Name]

		fieldHTML, err := fr.registry.RenderField(ctx, &field, value, fieldErrors)
		if err != nil {
			return "", err
		}

		fieldHTMLs = append(fieldHTMLs, fieldHTML)
	}

	return fmt.Sprintf(`<div class="form-fields">%s</div>`, strings.Join(fieldHTMLs, "\n")), nil
}

// renderSections renders form sections
func (fr *DefaultFormRenderer) renderSections(ctx context.Context, schema *Schema, data map[string]any, errors map[string][]string) (string, error) {
	var sectionHTMLs []string

	// Create field lookup map
	fieldMap := make(map[string]*Field)
	for i := range schema.Fields {
		fieldMap[schema.Fields[i].Name] = &schema.Fields[i]
	}

	for _, section := range schema.Layout.Sections {
		var sectionFields []Field
		for _, fieldName := range section.Fields {
			if field, exists := fieldMap[fieldName]; exists {
				sectionFields = append(sectionFields, *field)
			}
		}

		sectionHTML, err := fr.renderSection(ctx, section, sectionFields, data, errors)
		if err != nil {
			return "", err
		}

		sectionHTMLs = append(sectionHTMLs, sectionHTML)
	}

	return fmt.Sprintf(`<div class="form-sections">%s</div>`, strings.Join(sectionHTMLs, "\n")), nil
}

// renderSection renders a single section
func (fr *DefaultFormRenderer) renderSection(ctx context.Context, section Section, fields []Field, data map[string]any, errors map[string][]string) (string, error) {
	var parts []string

	// Section classes
	classes := []string{"form-section"}
	if section.Collapsible {
		classes = append(classes, "form-section--collapsible")
		if section.Collapsed {
			classes = append(classes, "form-section--collapsed")
		}
	}

	sectionAttrs := map[string]string{
		"id":    section.ID,
		"class": strings.Join(classes, " "),
	}

	parts = append(parts, fmt.Sprintf(`<section %s>`, fr.attributesToString(sectionAttrs)))

	// Section header
	if section.Title != "" {
		headerClass := "section-header"
		if section.Collapsible {
			headerClass += " section-header--collapsible"
		}

		headerHTML := fmt.Sprintf(`<header class="%s">`, headerClass)
		if section.Icon != "" {
			headerHTML += fmt.Sprintf(`<span class="section-icon">%s</span>`, html.EscapeString(section.Icon))
		}
		headerHTML += fmt.Sprintf(`<h2 class="section-title">%s</h2>`, html.EscapeString(section.Title))

		if section.Collapsible {
			headerHTML += `<button type="button" class="section-toggle" onclick="toggleSection(this)" aria-expanded="true">▼</button>`
		}
		headerHTML += `</header>`

		parts = append(parts, headerHTML)
	}

	if section.Description != "" {
		parts = append(parts, fmt.Sprintf(`<p class="section-description">%s</p>`, html.EscapeString(section.Description)))
	}

	// Section content
	parts = append(parts, `<div class="section-content">`)

	fieldsHTML, err := fr.renderFieldList(ctx, fields, data, errors)
	if err != nil {
		return "", err
	}

	parts = append(parts, fieldsHTML)
	parts = append(parts, `</div>`)
	parts = append(parts, `</section>`)

	return strings.Join(parts, "\n"), nil
}

// renderTabs renders tabbed layout
func (fr *DefaultFormRenderer) renderTabs(ctx context.Context, schema *Schema, data map[string]any, errors map[string][]string) (string, error) {
	var parts []string

	// Tab navigation
	parts = append(parts, `<div class="form-tabs">`)
	parts = append(parts, `<nav class="tab-nav" role="tablist">`)

	for i, tab := range schema.Layout.Tabs {
		activeClass := ""
		if i == 0 {
			activeClass = " tab-nav-item--active"
		}

		tabHTML := fmt.Sprintf(`<button type="button" class="tab-nav-item%s" role="tab" onclick="switchTab('%s', %d)" aria-selected="%t">`,
			activeClass, schema.ID, i, i == 0)

		if tab.Icon != "" {
			tabHTML += fmt.Sprintf(`<span class="tab-icon">%s</span>`, html.EscapeString(tab.Icon))
		}

		tabHTML += html.EscapeString(tab.Title)

		if tab.Badge != "" {
			tabHTML += fmt.Sprintf(`<span class="tab-badge">%s</span>`, html.EscapeString(tab.Badge))
		}

		tabHTML += `</button>`
		parts = append(parts, tabHTML)
	}

	parts = append(parts, `</nav>`)

	// Tab content
	for i, tab := range schema.Layout.Tabs {
		activeClass := ""
		if i == 0 {
			activeClass = " tab-content--active"
		}

		parts = append(parts, fmt.Sprintf(`<div class="tab-content%s" role="tabpanel" id="%s-tab-%d">`,
			activeClass, schema.ID, i))

		// Get fields for this tab
		var tabFields []Field
		fieldMap := make(map[string]*Field)
		for j := range schema.Fields {
			fieldMap[schema.Fields[j].Name] = &schema.Fields[j]
		}

		for _, fieldName := range tab.Fields {
			if field, exists := fieldMap[fieldName]; exists {
				tabFields = append(tabFields, *field)
			}
		}

		fieldsHTML, err := fr.renderFieldList(ctx, tabFields, data, errors)
		if err != nil {
			return "", err
		}

		parts = append(parts, fieldsHTML)
		parts = append(parts, `</div>`)
	}

	parts = append(parts, `</div>`)
	return strings.Join(parts, "\n"), nil
}

// renderSteps renders multi-step layout
func (fr *DefaultFormRenderer) renderSteps(ctx context.Context, schema *Schema, data map[string]any, errors map[string][]string) (string, error) {
	var parts []string

	// Step progress indicator
	parts = append(parts, `<div class="form-steps">`)
	parts = append(parts, `<div class="step-progress">`)

	for i, step := range schema.Layout.Steps {
		stepClass := "step-indicator"
		if i == 0 {
			stepClass += " step-indicator--active"
		}

		parts = append(parts, fmt.Sprintf(`<div class="%s">`, stepClass))
		parts = append(parts, fmt.Sprintf(`<span class="step-number">%d</span>`, i+1))
		parts = append(parts, fmt.Sprintf(`<span class="step-title">%s</span>`, html.EscapeString(step.Title)))
		parts = append(parts, `</div>`)
	}

	parts = append(parts, `</div>`)

	// Step content
	for i, step := range schema.Layout.Steps {
		activeClass := ""
		if i == 0 {
			activeClass = " step-content--active"
		}

		parts = append(parts, fmt.Sprintf(`<div class="step-content%s" id="%s-step-%d">`,
			activeClass, schema.ID, i))

		if step.Title != "" {
			parts = append(parts, fmt.Sprintf(`<h3 class="step-title">%s</h3>`, html.EscapeString(step.Title)))
		}

		if step.Description != "" {
			parts = append(parts, fmt.Sprintf(`<p class="step-description">%s</p>`, html.EscapeString(step.Description)))
		}

		// Get fields for this step
		var stepFields []Field
		fieldMap := make(map[string]*Field)
		for j := range schema.Fields {
			fieldMap[schema.Fields[j].Name] = &schema.Fields[j]
		}

		for _, fieldName := range step.Fields {
			if field, exists := fieldMap[fieldName]; exists {
				stepFields = append(stepFields, *field)
			}
		}

		fieldsHTML, err := fr.renderFieldList(ctx, stepFields, data, errors)
		if err != nil {
			return "", err
		}

		parts = append(parts, fieldsHTML)

		// Step navigation
		parts = append(parts, `<div class="step-navigation">`)
		if i > 0 {
			parts = append(parts, fmt.Sprintf(`<button type="button" class="step-btn step-btn--previous" onclick="previousStep('%s')">Previous</button>`, schema.ID))
		}
		if i < len(schema.Layout.Steps)-1 {
			parts = append(parts, fmt.Sprintf(`<button type="button" class="step-btn step-btn--next" onclick="nextStep('%s')">Next</button>`, schema.ID))
		}
		parts = append(parts, `</div>`)

		parts = append(parts, `</div>`)
	}

	parts = append(parts, `</div>`)
	return strings.Join(parts, "\n"), nil
}

// renderFormActions renders form action buttons
func (fr *DefaultFormRenderer) renderFormActions(schema *Schema) string {
	if len(schema.Actions) == 0 {
		return ""
	}

	var actionHTMLs []string

	for _, action := range schema.Actions {
		actionHTML := fr.renderAction(action)
		actionHTMLs = append(actionHTMLs, actionHTML)
	}

	return fmt.Sprintf(`<div class="form-actions">%s</div>`, strings.Join(actionHTMLs, "\n"))
}

// renderAction renders a single action button
func (fr *DefaultFormRenderer) renderAction(action Action) string {
	classes := []string{"action-button"}
	classes = append(classes, fmt.Sprintf("action-button--%s", action.Type))
	if action.Variant != "" {
		classes = append(classes, fmt.Sprintf("action-button--%s", action.Variant))
	}

	attrs := map[string]string{
		"id":    action.ID,
		"type":  string(action.Type),
		"class": strings.Join(classes, " "),
	}

	if action.Disabled {
		attrs["disabled"] = "disabled"
	}

	// Add action configuration
	if action.Config != nil {
		if action.Config.URL != "" && action.Type == ActionCustom {
			attrs["onclick"] = fmt.Sprintf("handleCustomAction('%s', '%s')", action.ID, action.Config.URL)
		}
	}

	// Add HTMX attributes
	if action.HTMX != nil {
		if action.HTMX.Post != "" {
			attrs["hx-post"] = action.HTMX.Post
		}
		if action.HTMX.Get != "" {
			attrs["hx-get"] = action.HTMX.Get
		}
		if action.HTMX.Target != "" {
			attrs["hx-target"] = action.HTMX.Target
		}
		if action.HTMX.Confirm != "" {
			attrs["hx-confirm"] = action.HTMX.Confirm
		}
	}

	// Add confirmation dialog
	if action.Confirm != nil {
		attrs["onclick"] = fmt.Sprintf("confirmAction('%s', '%s', '%s')",
			action.ID, action.Confirm.Title, action.Confirm.Message)
	}

	// Note: Action styling would be handled via CSS classes and design tokens

	return fmt.Sprintf(`<button %s>%s</button>`, fr.attributesToString(attrs), html.EscapeString(action.Text))
}

// resolveFormStyle resolves form-level design tokens
func (fr *DefaultFormRenderer) resolveFormStyle(schema *Schema) string {
	if fr.tokens == nil {
		return ""
	}

	var styles []string

	// Add layout-specific styles
	if schema.Layout != nil {
		if schema.Layout.Gap != "" {
			if value, err := fr.tokens.ResolveToken("spacing." + schema.Layout.Gap); err == nil {
				styles = append(styles, fmt.Sprintf("gap: %s", value))
			} else {
				styles = append(styles, fmt.Sprintf("gap: %s", schema.Layout.Gap))
			}
		}
	}

	return strings.Join(styles, "; ")
}

// renderRequiredScripts renders JavaScript dependencies
func (fr *DefaultFormRenderer) renderRequiredScripts(schema *Schema) string {
	assets := fr.registry.GetAllRequiredAssets()

	var scripts []string
	for _, asset := range assets {
		if strings.HasSuffix(asset, ".js") {
			scripts = append(scripts, fmt.Sprintf(`<script src="/assets/%s"></script>`, asset))
		}
	}

	if len(scripts) > 0 {
		return strings.Join(scripts, "\n")
	}

	return ""
}

// attributesToString converts attributes map to HTML string
func (fr *DefaultFormRenderer) attributesToString(attrs map[string]string) string {
	var parts []string
	for key, value := range attrs {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, key, html.EscapeString(value)))
	}
	return strings.Join(parts, " ")
}

// RenderFormWithErrors renders a complete form with error handling (extended interface)
func (fr *DefaultFormRenderer) RenderFormWithErrors(ctx context.Context, schema *Schema, data map[string]any, errors map[string][]string) (string, error) {
	if schema.Type != TypeForm {
		return "", NewRenderError("invalid_schema_type", "schema must be of type 'form'")
	}

	var parts []string

	// Form opening tag with attributes
	formAttrs := fr.buildFormAttributes(schema)
	parts = append(parts, fmt.Sprintf(`<form %s>`, fr.attributesToString(formAttrs)))

	// Form header (title, description)
	if header := fr.renderFormHeader(schema); header != "" {
		parts = append(parts, header)
	}

	// CSRF token if security is enabled
	if schema.Security != nil && schema.Security.CSRF != nil && schema.Security.CSRF.Enabled {
		parts = append(parts, fr.renderCSRFToken(schema.Security.CSRF))
	}

	// Render form content based on layout
	content, err := fr.renderFormContent(ctx, schema, data, errors)
	if err != nil {
		return "", err
	}
	parts = append(parts, content)

	// Form actions (submit, reset, etc.)
	if actions := fr.renderFormActions(schema); actions != "" {
		parts = append(parts, actions)
	}

	// Form closing tag
	parts = append(parts, `</form>`)

	// Required assets and scripts
	if scripts := fr.renderRequiredScripts(schema); scripts != "" {
		parts = append(parts, scripts)
	}

	return strings.Join(parts, "\n"), nil
}

// Interface implementation methods for FormRenderer

// RenderForm implements FormRenderer interface (without errors parameter)
func (fr *DefaultFormRenderer) RenderForm(ctx context.Context, schema *Schema, data map[string]any) (string, error) {
	return fr.RenderFormWithErrors(ctx, schema, data, nil)
}

// RenderField implements FormRenderer interface
func (fr *DefaultFormRenderer) RenderField(ctx context.Context, field *Field, value any) (string, error) {
	return fr.registry.RenderField(ctx, field, value, nil)
}

// RenderAction implements FormRenderer interface
func (fr *DefaultFormRenderer) RenderAction(ctx context.Context, action *Action) (string, error) {
	return fr.renderAction(*action), nil
}
