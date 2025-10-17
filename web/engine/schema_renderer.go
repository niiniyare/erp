package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/niiniyare/erp/web/components/atoms"
)

// JSONSchemaRenderer provides a unified interface for rendering components from JSON schemas
// It bridges the gap between JSON schema definitions and Templ components
type JSONSchemaRenderer struct {
	factory   *SchemaFactory
	templates map[string]TemplateFunc
}

// TemplateFunc represents a function that can render a Templ component
type TemplateFunc func(props interface{}) string

// NewJSONSchemaRenderer creates a new unified schema renderer
func NewJSONSchemaRenderer(schemaDir string) (*JSONSchemaRenderer, error) {
	factory, err := NewSchemaFactory(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema factory: %w", err)
	}

	renderer := &JSONSchemaRenderer{
		factory:   factory,
		templates: make(map[string]TemplateFunc),
	}

	// Register Templ template functions
	renderer.registerTemplateRenderers()

	return renderer, nil
}

// registerTemplateRenderers registers Templ component renderers
func (sr *JSONSchemaRenderer) registerTemplateRenderers() {
	// Atom components
	sr.templates["Button"] = sr.renderButton
	sr.templates["Input"] = sr.renderInput
	sr.templates["Textarea"] = sr.renderTextarea
	sr.templates["Select"] = sr.renderSelect
	sr.templates["Checkbox"] = sr.renderCheckbox
	sr.templates["Radio"] = sr.renderRadio

	// Molecule components  
	sr.templates["Card"] = sr.renderCard
	sr.templates["Form"] = sr.renderForm
	sr.templates["Table"] = sr.renderTable

	// Layout components
	sr.templates["Container"] = sr.renderContainer
}

// RenderComponent renders a component from JSON schema using Templ
func (sr *JSONSchemaRenderer) RenderComponent(ctx context.Context, schemaType string, props map[string]interface{}) (string, error) {
	// Get component from schema factory
	component, err := sr.factory.RenderFromSchema(ctx, schemaType, props)
	if err != nil {
		return "", fmt.Errorf("failed to render from schema: %w", err)
	}

	// Get template renderer
	templateFunc, exists := sr.templates[component.Type]
	if !exists {
		return "", fmt.Errorf("template renderer not found for type: %s", component.Type)
	}

	// Render using Templ
	html := templateFunc(component.Props)
	
	// Apply additional attributes and events
	html = sr.applyComponentEnhancements(html, component)

	return html, nil
}

// RenderComponentTree renders a tree of components
func (sr *JSONSchemaRenderer) RenderComponentTree(ctx context.Context, components []ComponentDefinition) (string, error) {
	var html strings.Builder

	for _, comp := range components {
		// Convert component definition to props
		props := sr.componentToProps(comp)
		
		// Render component
		componentHTML, err := sr.RenderComponent(ctx, comp.SchemaType, props)
		if err != nil {
			return "", fmt.Errorf("failed to render component %s: %w", comp.ID, err)
		}

		html.WriteString(componentHTML)
	}

	return html.String(), nil
}

// ComponentDefinition defines a component instance with its configuration
type ComponentDefinition struct {
	ID         string                 `json:"id"`
	SchemaType string                 `json:"schemaType"`
	Props      map[string]interface{} `json:"props"`
	Children   []ComponentDefinition  `json:"children,omitempty"`
	Conditions []RenderCondition      `json:"conditions,omitempty"`
}

// RenderCondition defines when a component should be rendered
type RenderCondition struct {
	Type     string `json:"type"`     // "permission", "feature", "data"
	Operator string `json:"operator"` // "equals", "contains", "exists"
	Value    string `json:"value"`
	Negate   bool   `json:"negate,omitempty"`
}

// componentToProps converts ComponentDefinition to props map
func (sr *JSONSchemaRenderer) componentToProps(comp ComponentDefinition) map[string]interface{} {
	props := make(map[string]interface{})
	
	// Copy all props
	for key, value := range comp.Props {
		props[key] = value
	}
	
	// Add metadata
	props["id"] = comp.ID
	props["_schemaType"] = comp.SchemaType
	
	return props
}

// applyComponentEnhancements applies additional attributes and events to HTML
func (sr *JSONSchemaRenderer) applyComponentEnhancements(html string, component TemplComponent) string {
	// Apply CSS classes
	if component.CSS != "" {
		html = sr.addCSSClasses(html, component.CSS)
	}

	// Apply attributes
	for attr, value := range component.Attributes {
		html = sr.addAttribute(html, attr, value)
	}

	// Apply events (HTMX attributes)
	for event, handler := range component.Events {
		switch event {
		case "click":
			if handler != "" {
				html = sr.addAttribute(html, "hx-post", handler)
				html = sr.addAttribute(html, "hx-trigger", "click")
			}
		case "change":
			if handler != "" {
				html = sr.addAttribute(html, "hx-post", handler)
				html = sr.addAttribute(html, "hx-trigger", "change")
			}
		}
	}

	return html
}

// addCSSClasses adds CSS classes to HTML element
func (sr *JSONSchemaRenderer) addCSSClasses(html, classes string) string {
	// Simple implementation - in production, use proper HTML parsing
	if strings.Contains(html, "class=\"") {
		return strings.Replace(html, "class=\"", fmt.Sprintf("class=\"%s ", classes), 1)
	} else {
		// Add class attribute to first tag
		tagEnd := strings.Index(html, ">")
		if tagEnd > 0 {
			return html[:tagEnd] + fmt.Sprintf(` class="%s"`, classes) + html[tagEnd:]
		}
	}
	return html
}

// addAttribute adds an attribute to HTML element
func (sr *JSONSchemaRenderer) addAttribute(html, attr, value string) string {
	// Simple implementation - in production, use proper HTML parsing
	tagEnd := strings.Index(html, ">")
	if tagEnd > 0 {
		return html[:tagEnd] + fmt.Sprintf(` %s="%s"`, attr, value) + html[tagEnd:]
	}
	return html
}

// ============================================================================
// TEMPL COMPONENT RENDERERS
// ============================================================================

// renderButton renders a Button component using atoms.Button
func (sr *JSONSchemaRenderer) renderButton(props interface{}) string {
	buttonProps, ok := props.(atoms.ButtonProps)
	if !ok {
		return "<!-- Invalid button props -->"
	}

	// In a real implementation, this would call the actual Templ component
	// For now, return a simple HTML representation
	return fmt.Sprintf(`<button type="%s" class="btn btn-%s btn-%s" %s>%s</button>`,
		buttonProps.Type,
		buttonProps.Variant,
		buttonProps.Size,
		sr.renderButtonAttributes(buttonProps),
		buttonProps.Text,
	)
}

func (sr *JSONSchemaRenderer) renderButtonAttributes(props atoms.ButtonProps) string {
	var attrs []string
	
	if props.ID != "" {
		attrs = append(attrs, fmt.Sprintf(`id="%s"`, props.ID))
	}
	if props.Disabled {
		attrs = append(attrs, "disabled")
	}
	if props.OnClick != "" {
		attrs = append(attrs, fmt.Sprintf(`onclick="%s"`, props.OnClick))
	}
	if props.AriaLabel != "" {
		attrs = append(attrs, fmt.Sprintf(`aria-label="%s"`, props.AriaLabel))
	}
	
	return strings.Join(attrs, " ")
}

// renderInput renders an Input component using atoms.Input
func (sr *JSONSchemaRenderer) renderInput(props interface{}) string {
	inputProps, ok := props.(atoms.InputProps)
	if !ok {
		return "<!-- Invalid input props -->"
	}

	return fmt.Sprintf(`<input type="%s" value="%s" placeholder="%s" %s>`,
		inputProps.Type,
		inputProps.Value,
		inputProps.Placeholder,
		sr.renderInputAttributes(inputProps),
	)
}

func (sr *JSONSchemaRenderer) renderInputAttributes(props atoms.InputProps) string {
	var attrs []string
	
	if props.ID != "" {
		attrs = append(attrs, fmt.Sprintf(`id="%s"`, props.ID))
	}
	if props.Name != "" {
		attrs = append(attrs, fmt.Sprintf(`name="%s"`, props.Name))
	}
	if props.Required {
		attrs = append(attrs, "required")
	}
	if props.Disabled {
		attrs = append(attrs, "disabled")
	}
	if props.ReadOnly {
		attrs = append(attrs, "readonly")
	}
	if props.MaxLength > 0 {
		attrs = append(attrs, fmt.Sprintf(`maxlength="%d"`, props.MaxLength))
	}
	if props.MinLength > 0 {
		attrs = append(attrs, fmt.Sprintf(`minlength="%d"`, props.MinLength))
	}
	if props.Pattern != "" {
		attrs = append(attrs, fmt.Sprintf(`pattern="%s"`, props.Pattern))
	}
	if props.AriaLabel != "" {
		attrs = append(attrs, fmt.Sprintf(`aria-label="%s"`, props.AriaLabel))
	}
	
	return strings.Join(attrs, " ")
}

// renderTextarea renders a Textarea component
func (sr *JSONSchemaRenderer) renderTextarea(props interface{}) string {
	textareaProps, ok := props.(atoms.TextareaProps)
	if !ok {
		return "<!-- Invalid textarea props -->"
	}

	return fmt.Sprintf(`<textarea %s>%s</textarea>`,
		sr.renderTextareaAttributes(textareaProps),
		textareaProps.Value,
	)
}

func (sr *JSONSchemaRenderer) renderTextareaAttributes(props atoms.TextareaProps) string {
	var attrs []string
	
	if props.ID != "" {
		attrs = append(attrs, fmt.Sprintf(`id="%s"`, props.ID))
	}
	if props.Name != "" {
		attrs = append(attrs, fmt.Sprintf(`name="%s"`, props.Name))
	}
	if props.Placeholder != "" {
		attrs = append(attrs, fmt.Sprintf(`placeholder="%s"`, props.Placeholder))
	}
	if props.Rows > 0 {
		attrs = append(attrs, fmt.Sprintf(`rows="%d"`, props.Rows))
	}
	if props.Cols > 0 {
		attrs = append(attrs, fmt.Sprintf(`cols="%d"`, props.Cols))
	}
	if props.MaxLength > 0 {
		attrs = append(attrs, fmt.Sprintf(`maxlength="%d"`, props.MaxLength))
	}
	if props.Required {
		attrs = append(attrs, "required")
	}
	if props.Disabled {
		attrs = append(attrs, "disabled")
	}
	if props.ReadOnly {
		attrs = append(attrs, "readonly")
	}
	if !props.Resizable {
		attrs = append(attrs, `style="resize: none"`)
	}
	
	return strings.Join(attrs, " ")
}

// renderSelect renders a Select component
func (sr *JSONSchemaRenderer) renderSelect(props interface{}) string {
	selectProps, ok := props.(atoms.SelectProps)
	if !ok {
		return "<!-- Invalid select props -->"
	}

	var options strings.Builder
	for _, option := range selectProps.Options {
		selected := ""
		if option.Value == selectProps.Value {
			selected = " selected"
		}
		disabled := ""
		if option.Disabled {
			disabled = " disabled"
		}
		
		options.WriteString(fmt.Sprintf(`<option value="%s"%s%s>%s</option>`,
			option.Value, selected, disabled, option.Label))
	}

	return fmt.Sprintf(`<select %s>%s</select>`,
		sr.renderSelectAttributes(selectProps),
		options.String(),
	)
}

func (sr *JSONSchemaRenderer) renderSelectAttributes(props atoms.SelectProps) string {
	var attrs []string
	
	if props.ID != "" {
		attrs = append(attrs, fmt.Sprintf(`id="%s"`, props.ID))
	}
	if props.Name != "" {
		attrs = append(attrs, fmt.Sprintf(`name="%s"`, props.Name))
	}
	if props.Multiple {
		attrs = append(attrs, "multiple")
	}
	if props.Required {
		attrs = append(attrs, "required")
	}
	if props.Disabled {
		attrs = append(attrs, "disabled")
	}
	if props.AriaLabel != "" {
		attrs = append(attrs, fmt.Sprintf(`aria-label="%s"`, props.AriaLabel))
	}
	
	return strings.Join(attrs, " ")
}

// renderCheckbox renders a Checkbox component
func (sr *JSONSchemaRenderer) renderCheckbox(props interface{}) string {
	checkboxProps, ok := props.(atoms.CheckboxProps)
	if !ok {
		return "<!-- Invalid checkbox props -->"
	}

	checked := ""
	if checkboxProps.Checked {
		checked = " checked"
	}

	return fmt.Sprintf(`<label class="checkbox-wrapper">
		<input type="checkbox" value="%s"%s %s>
		<span class="checkbox-label">%s</span>
	</label>`,
		checkboxProps.Value,
		checked,
		sr.renderCheckboxAttributes(checkboxProps),
		checkboxProps.Label,
	)
}

func (sr *JSONSchemaRenderer) renderCheckboxAttributes(props atoms.CheckboxProps) string {
	var attrs []string
	
	if props.ID != "" {
		attrs = append(attrs, fmt.Sprintf(`id="%s"`, props.ID))
	}
	if props.Name != "" {
		attrs = append(attrs, fmt.Sprintf(`name="%s"`, props.Name))
	}
	if props.Required {
		attrs = append(attrs, "required")
	}
	if props.Disabled {
		attrs = append(attrs, "disabled")
	}
	if props.AriaLabel != "" {
		attrs = append(attrs, fmt.Sprintf(`aria-label="%s"`, props.AriaLabel))
	}
	
	return strings.Join(attrs, " ")
}

// renderRadio renders a Radio component
func (sr *JSONSchemaRenderer) renderRadio(props interface{}) string {
	// Placeholder implementation
	return `<div class="radio-group"><!-- Radio component --></div>`
}

// renderCard renders a Card component
func (sr *JSONSchemaRenderer) renderCard(props interface{}) string {
	// Placeholder implementation
	return `<div class="card"><!-- Card component --></div>`
}

// renderForm renders a Form component
func (sr *JSONSchemaRenderer) renderForm(props interface{}) string {
	// Placeholder implementation
	return `<form><!-- Form component --></form>`
}

// renderTable renders a Table component
func (sr *JSONSchemaRenderer) renderTable(props interface{}) string {
	// Placeholder implementation
	return `<table class="table"><!-- Table component --></table>`
}

// renderContainer renders a Container component
func (sr *JSONSchemaRenderer) renderContainer(props interface{}) string {
	// Placeholder implementation
	return `<div class="container"><!-- Container component --></div>`
}

// ============================================================================
// SCHEMA INTEGRATION UTILITIES
// ============================================================================

// ConvertSchemaToComponent converts a JSON schema definition to a component tree
func (sr *JSONSchemaRenderer) ConvertSchemaToComponent(schemaData map[string]interface{}) (ComponentDefinition, error) {
	componentType, ok := schemaData["type"].(string)
	if !ok {
		return ComponentDefinition{}, fmt.Errorf("component type is required")
	}

	// Generate unique ID if not provided
	id, ok := schemaData["id"].(string)
	if !ok {
		id = fmt.Sprintf("%s_%d", componentType, len(schemaData))
	}

	// Extract props (all properties except special ones)
	props := make(map[string]interface{})
	for key, value := range schemaData {
		switch key {
		case "id", "type", "children", "conditions":
			// Skip these special properties
		default:
			props[key] = value
		}
	}

	component := ComponentDefinition{
		ID:         id,
		SchemaType: componentType + "Schema", // Append "Schema" to match schema names
		Props:      props,
	}

	// Handle children
	if childrenData, exists := schemaData["children"]; exists {
		if children, ok := childrenData.([]interface{}); ok {
			for _, child := range children {
				if childMap, ok := child.(map[string]interface{}); ok {
					childComponent, err := sr.ConvertSchemaToComponent(childMap)
					if err != nil {
						return component, fmt.Errorf("failed to convert child component: %w", err)
					}
					component.Children = append(component.Children, childComponent)
				}
			}
		}
	}

	return component, nil
}

// RenderFromJSON renders components directly from JSON string
func (sr *JSONSchemaRenderer) RenderFromJSON(ctx context.Context, jsonData string) (string, error) {
	// Parse JSON into component definition
	var componentData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &componentData); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to component definition
	component, err := sr.ConvertSchemaToComponent(componentData)
	if err != nil {
		return "", fmt.Errorf("failed to convert schema: %w", err)
	}

	// Render component
	return sr.RenderComponent(ctx, component.SchemaType, component.Props)
}

// GetAvailableComponents returns all available component types
func (sr *JSONSchemaRenderer) GetAvailableComponents() []string {
	return sr.factory.GetAvailableSchemas()
}

// ValidateComponent validates a component definition against its schema
func (sr *JSONSchemaRenderer) ValidateComponent(component ComponentDefinition) error {
	return sr.factory.ValidateAgainstSchema(component.SchemaType, component.Props)
}