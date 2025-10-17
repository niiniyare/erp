package engine

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/niiniyare/erp/web/components/atoms"
)

// SchemaTemplRenderer converts TemplComponent to actual Templ components
type SchemaTemplRenderer struct {
	componentMap map[string]TemplComponentRenderer
}

// TemplComponentRenderer defines how to convert a TemplComponent to a Templ component
type TemplComponentRenderer func(component TemplComponent) templ.Component

// NewSchemaTemplRenderer creates a new schema-to-templ renderer
func NewSchemaTemplRenderer() *SchemaTemplRenderer {
	renderer := &SchemaTemplRenderer{
		componentMap: make(map[string]TemplComponentRenderer),
	}
	
	// Register built-in component renderers
	renderer.registerBuiltinRenderers()
	
	return renderer
}

// registerBuiltinRenderers registers the core UI component renderers
func (str *SchemaTemplRenderer) registerBuiltinRenderers() {
	// Button components
	str.componentMap["Button"] = str.renderButton
	str.componentMap["button"] = str.renderButton
	
	// ButtonGroup components
	str.componentMap["ButtonGroup"] = str.renderButtonGroup
	str.componentMap["button-group"] = str.renderButtonGroup
	
	// Form components
	str.componentMap["Input"] = str.renderInput
	str.componentMap["input"] = str.renderInput
	str.componentMap["Textarea"] = str.renderTextarea
	str.componentMap["textarea"] = str.renderTextarea
	str.componentMap["Checkbox"] = str.renderCheckbox
	str.componentMap["checkbox"] = str.renderCheckbox
	str.componentMap["Select"] = str.renderSelect
	str.componentMap["select"] = str.renderSelect
	
	// Layout components
	str.componentMap["Container"] = str.renderContainer
	str.componentMap["container"] = str.renderContainer
	str.componentMap["Card"] = str.renderCard
	str.componentMap["card"] = str.renderCard
	str.componentMap["Panel"] = str.renderPanel
	str.componentMap["panel"] = str.renderPanel
	
	// Data components
	str.componentMap["Table"] = str.renderTable
	str.componentMap["table"] = str.renderTable
	str.componentMap["List"] = str.renderList
	str.componentMap["list"] = str.renderList
}

// RenderComponent converts a TemplComponent to a Templ component
func (str *SchemaTemplRenderer) RenderComponent(component TemplComponent) templ.Component {
	renderer, exists := str.componentMap[component.Type]
	if !exists {
		// Return a fallback component for unknown types
		return str.renderFallback(component)
	}
	
	return renderer(component)
}

// RegisterRenderer allows custom component renderers to be registered
func (str *SchemaTemplRenderer) RegisterRenderer(componentType string, renderer TemplComponentRenderer) {
	str.componentMap[componentType] = renderer
}

// ============================================================================
// COMPONENT RENDERERS
// ============================================================================

// renderButton converts TemplComponent to atoms.Button
func (str *SchemaTemplRenderer) renderButton(component TemplComponent) templ.Component {
	// Extract props from the component
	props := str.extractButtonProps(component)
	
	// Return the actual Templ button component
	return atoms.Button(props)
}

// renderButtonGroup converts TemplComponent to a group of buttons
func (str *SchemaTemplRenderer) renderButtonGroup(component TemplComponent) templ.Component {
	// For ButtonGroup, we need to render multiple buttons
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		// Extract buttons from props
		buttons := str.extractButtonGroupButtons(component)
		
		// Render button group wrapper
		if _, err := w.Write([]byte(`<div class="btn-group" role="group">`)); err != nil {
			return err
		}
		
		// Render each button
		for _, buttonProps := range buttons {
			buttonComponent := atoms.Button(buttonProps)
			if err := buttonComponent.Render(ctx, w); err != nil {
				return err
			}
		}
		
		// Close wrapper
		if _, err := w.Write([]byte(`</div>`)); err != nil {
			return err
		}
		
		return nil
	})
}

// renderInput converts TemplComponent to atoms.Input
func (str *SchemaTemplRenderer) renderInput(component TemplComponent) templ.Component {
	props := str.extractInputProps(component)
	return atoms.Input(props)
}

// renderTextarea converts TemplComponent to atoms.Textarea  
func (str *SchemaTemplRenderer) renderTextarea(component TemplComponent) templ.Component {
	props := str.extractTextareaProps(component)
	return atoms.Textarea(props)
}

// renderCheckbox converts TemplComponent to atoms.Checkbox
func (str *SchemaTemplRenderer) renderCheckbox(component TemplComponent) templ.Component {
	props := str.extractCheckboxProps(component)
	return atoms.Checkbox(props)
}

// renderSelect converts TemplComponent to atoms.Select
func (str *SchemaTemplRenderer) renderSelect(component TemplComponent) templ.Component {
	props := str.extractSelectProps(component)
	return atoms.Select(props)
}

// renderContainer renders a layout container
func (str *SchemaTemplRenderer) renderContainer(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		class := str.extractClass(component, "container")
		
		if _, err := w.Write([]byte(fmt.Sprintf(`<div class="%s">`, class))); err != nil {
			return err
		}
		
		// Render children if any
		for _, child := range component.Children {
			childComponent := str.RenderComponent(child)
			if err := childComponent.Render(ctx, w); err != nil {
				return err
			}
		}
		
		if _, err := w.Write([]byte(`</div>`)); err != nil {
			return err
		}
		
		return nil
	})
}

// renderCard renders a card component
func (str *SchemaTemplRenderer) renderCard(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		class := str.extractClass(component, "card")
		
		if _, err := w.Write([]byte(fmt.Sprintf(`<div class="%s">`, class))); err != nil {
			return err
		}
		
		// Render card body
		if _, err := w.Write([]byte(`<div class="card-body">`)); err != nil {
			return err
		}
		
		// Render children
		for _, child := range component.Children {
			childComponent := str.RenderComponent(child)
			if err := childComponent.Render(ctx, w); err != nil {
				return err
			}
		}
		
		if _, err := w.Write([]byte(`</div></div>`)); err != nil {
			return err
		}
		
		return nil
	})
}

// renderPanel renders a panel component
func (str *SchemaTemplRenderer) renderPanel(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		class := str.extractClass(component, "panel")
		
		if _, err := w.Write([]byte(fmt.Sprintf(`<div class="%s">`, class))); err != nil {
			return err
		}
		
		// Render children
		for _, child := range component.Children {
			childComponent := str.RenderComponent(child)
			if err := childComponent.Render(ctx, w); err != nil {
				return err
			}
		}
		
		if _, err := w.Write([]byte(`</div>`)); err != nil {
			return err
		}
		
		return nil
	})
}

// renderTable renders a data table
func (str *SchemaTemplRenderer) renderTable(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		class := str.extractClass(component, "table table-striped")
		
		if _, err := w.Write([]byte(fmt.Sprintf(`<table class="%s">`, class))); err != nil {
			return err
		}
		
		// For now, render a simple placeholder table
		if _, err := w.Write([]byte(`
			<thead>
				<tr><th>Column 1</th><th>Column 2</th><th>Actions</th></tr>
			</thead>
			<tbody>
				<tr><td>Data 1</td><td>Data 2</td><td><button class="btn btn-sm">Edit</button></td></tr>
			</tbody>
		`)); err != nil {
			return err
		}
		
		if _, err := w.Write([]byte(`</table>`)); err != nil {
			return err
		}
		
		return nil
	})
}

// renderList renders a data list
func (str *SchemaTemplRenderer) renderList(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		class := str.extractClass(component, "list-group")
		
		if _, err := w.Write([]byte(fmt.Sprintf(`<ul class="%s">`, class))); err != nil {
			return err
		}
		
		// Render list items
		if _, err := w.Write([]byte(`
			<li class="list-group-item">List Item 1</li>
			<li class="list-group-item">List Item 2</li>
			<li class="list-group-item">List Item 3</li>
		`)); err != nil {
			return err
		}
		
		if _, err := w.Write([]byte(`</ul>`)); err != nil {
			return err
		}
		
		return nil
	})
}

// renderFallback renders a fallback component for unknown types
func (str *SchemaTemplRenderer) renderFallback(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := w.Write([]byte(fmt.Sprintf(
			`<div class="alert alert-warning">Unknown component type: %s</div>`,
			component.Type,
		))); err != nil {
			return err
		}
		return nil
	})
}

// ============================================================================
// PROP EXTRACTION HELPERS
// ============================================================================

// extractButtonProps converts component props to atoms.ButtonProps
func (str *SchemaTemplRenderer) extractButtonProps(component TemplComponent) atoms.ButtonProps {
	props := atoms.ButtonProps{
		Type:     "button",
		Variant:  atoms.ButtonPrimary,
		Size:     atoms.ButtonSizeMD,
		Disabled: false,
		Loading:  false,
	}
	
	// Type assertion to extract props
	if propsMap, ok := component.Props.(map[string]interface{}); ok {
		if text, ok := propsMap["text"].(string); ok {
			props.Text = text
		}
		if btnType, ok := propsMap["type"].(string); ok {
			props.Type = btnType
		}
		if variant, ok := propsMap["variant"].(string); ok {
			props.Variant = atoms.ButtonVariant(variant)
		}
		if size, ok := propsMap["size"].(string); ok {
			props.Size = atoms.ButtonSize(size)
		}
		if disabled, ok := propsMap["disabled"].(bool); ok {
			props.Disabled = disabled
		}
		if loading, ok := propsMap["loading"].(bool); ok {
			props.Loading = loading
		}
		if id, ok := propsMap["id"].(string); ok {
			props.ID = id
		}
		if class, ok := propsMap["className"].(string); ok {
			props.Class = class
		}
		if onClick, ok := propsMap["onClick"].(string); ok {
			props.OnClick = onClick
		}
	}
	
	// Apply CSS from component if available
	if component.CSS != "" {
		if props.Class != "" {
			props.Class += " " + component.CSS
		} else {
			props.Class = component.CSS
		}
	}
	
	return props
}

// extractButtonGroupButtons extracts button configurations from ButtonGroup props
func (str *SchemaTemplRenderer) extractButtonGroupButtons(component TemplComponent) []atoms.ButtonProps {
	var buttons []atoms.ButtonProps
	
	if propsMap, ok := component.Props.(map[string]interface{}); ok {
		if buttonsData, ok := propsMap["buttons"].([]interface{}); ok {
			for _, buttonData := range buttonsData {
				if buttonMap, ok := buttonData.(map[string]interface{}); ok {
					// Create a temporary TemplComponent for each button
					tempComponent := TemplComponent{
						Type:  "Button",
						Props: buttonMap,
						CSS:   component.CSS,
					}
					buttonProps := str.extractButtonProps(tempComponent)
					buttons = append(buttons, buttonProps)
				}
			}
		}
	}
	
	// If no buttons specified, create a default button
	if len(buttons) == 0 {
		buttons = append(buttons, str.extractButtonProps(component))
	}
	
	return buttons
}

// extractInputProps converts component props to atoms.InputProps
func (str *SchemaTemplRenderer) extractInputProps(component TemplComponent) atoms.InputProps {
	props := atoms.InputProps{
		Type:     atoms.InputText,
		Size:     atoms.InputSizeMD,
		Required: false,
		Disabled: false,
		ReadOnly: false,
	}
	
	if propsMap, ok := component.Props.(map[string]interface{}); ok {
		if label, ok := propsMap["label"].(string); ok {
			props.Label = label
		}
		if name, ok := propsMap["name"].(string); ok {
			props.Name = name
		}
		if value, ok := propsMap["value"].(string); ok {
			props.Value = value
		}
		if placeholder, ok := propsMap["placeholder"].(string); ok {
			props.Placeholder = placeholder
		}
		if inputType, ok := propsMap["type"].(string); ok {
			props.Type = atoms.InputType(inputType)
		}
		if size, ok := propsMap["size"].(string); ok {
			props.Size = atoms.InputSize(size)
		}
		if required, ok := propsMap["required"].(bool); ok {
			props.Required = required
		}
		if disabled, ok := propsMap["disabled"].(bool); ok {
			props.Disabled = disabled
		}
		if readonly, ok := propsMap["readonly"].(bool); ok {
			props.ReadOnly = readonly
		}
		if id, ok := propsMap["id"].(string); ok {
			props.ID = id
		}
		if class, ok := propsMap["className"].(string); ok {
			props.Class = class
		}
	}
	
	// Apply CSS from component
	if component.CSS != "" {
		if props.Class != "" {
			props.Class += " " + component.CSS
		} else {
			props.Class = component.CSS
		}
	}
	
	return props
}

// extractTextareaProps converts component props to atoms.TextareaProps
func (str *SchemaTemplRenderer) extractTextareaProps(component TemplComponent) atoms.TextareaProps {
	props := atoms.TextareaProps{
		Rows:     3,
		Required: false,
		Disabled: false,
		ReadOnly: false,
	}
	
	if propsMap, ok := component.Props.(map[string]interface{}); ok {
		if label, ok := propsMap["label"].(string); ok {
			props.Label = label
		}
		if name, ok := propsMap["name"].(string); ok {
			props.Name = name
		}
		if value, ok := propsMap["value"].(string); ok {
			props.Value = value
		}
		if placeholder, ok := propsMap["placeholder"].(string); ok {
			props.Placeholder = placeholder
		}
		if rows, ok := propsMap["rows"].(float64); ok {
			props.Rows = int(rows)
		}
		if required, ok := propsMap["required"].(bool); ok {
			props.Required = required
		}
		if disabled, ok := propsMap["disabled"].(bool); ok {
			props.Disabled = disabled
		}
		if readonly, ok := propsMap["readonly"].(bool); ok {
			props.ReadOnly = readonly
		}
		if id, ok := propsMap["id"].(string); ok {
			props.ID = id
		}
		if class, ok := propsMap["className"].(string); ok {
			props.Class = class
		}
	}
	
	// Apply CSS from component
	if component.CSS != "" {
		if props.Class != "" {
			props.Class += " " + component.CSS
		} else {
			props.Class = component.CSS
		}
	}
	
	return props
}

// extractCheckboxProps converts component props to atoms.CheckboxProps
func (str *SchemaTemplRenderer) extractCheckboxProps(component TemplComponent) atoms.CheckboxProps {
	props := atoms.CheckboxProps{
		Size:     atoms.CheckboxSizeMD,
		Checked:  false,
		Required: false,
		Disabled: false,
	}
	
	if propsMap, ok := component.Props.(map[string]interface{}); ok {
		if label, ok := propsMap["label"].(string); ok {
			props.Label = label
		}
		if name, ok := propsMap["name"].(string); ok {
			props.Name = name
		}
		if value, ok := propsMap["value"].(string); ok {
			props.Value = value
		}
		if checked, ok := propsMap["checked"].(bool); ok {
			props.Checked = checked
		}
		if size, ok := propsMap["size"].(string); ok {
			props.Size = atoms.CheckboxSize(size)
		}
		if required, ok := propsMap["required"].(bool); ok {
			props.Required = required
		}
		if disabled, ok := propsMap["disabled"].(bool); ok {
			props.Disabled = disabled
		}
		if id, ok := propsMap["id"].(string); ok {
			props.ID = id
		}
		if class, ok := propsMap["className"].(string); ok {
			props.Class = class
		}
	}
	
	// Apply CSS from component
	if component.CSS != "" {
		if props.Class != "" {
			props.Class += " " + component.CSS
		} else {
			props.Class = component.CSS
		}
	}
	
	return props
}

// extractSelectProps converts component props to atoms.SelectProps
func (str *SchemaTemplRenderer) extractSelectProps(component TemplComponent) atoms.SelectProps {
	props := atoms.SelectProps{
		SelectSize: atoms.SelectSizeMD,
		Required:   false,
		Disabled:   false,
		Multiple:   false,
	}
	
	if propsMap, ok := component.Props.(map[string]interface{}); ok {
		if label, ok := propsMap["label"].(string); ok {
			props.Label = label
		}
		if name, ok := propsMap["name"].(string); ok {
			props.Name = name
		}
		if value, ok := propsMap["value"].(string); ok {
			props.Value = value
		}
		if size, ok := propsMap["size"].(string); ok {
			props.SelectSize = atoms.SelectSize(size)
		}
		if required, ok := propsMap["required"].(bool); ok {
			props.Required = required
		}
		if disabled, ok := propsMap["disabled"].(bool); ok {
			props.Disabled = disabled
		}
		if multiple, ok := propsMap["multiple"].(bool); ok {
			props.Multiple = multiple
		}
		if id, ok := propsMap["id"].(string); ok {
			props.ID = id
		}
		if class, ok := propsMap["className"].(string); ok {
			props.Class = class
		}
		
		// Extract options
		if optionsData, ok := propsMap["options"].([]interface{}); ok {
			for _, optionData := range optionsData {
				if optionMap, ok := optionData.(map[string]interface{}); ok {
					option := atoms.SelectOption{}
					if label, ok := optionMap["label"].(string); ok {
						option.Label = label
					}
					if value, ok := optionMap["value"].(string); ok {
						option.Value = value
					}
					if selected, ok := optionMap["selected"].(bool); ok {
						option.Selected = selected
					}
					if disabled, ok := optionMap["disabled"].(bool); ok {
						option.Disabled = disabled
					}
					props.Options = append(props.Options, option)
				}
			}
		}
	}
	
	// Apply CSS from component
	if component.CSS != "" {
		if props.Class != "" {
			props.Class += " " + component.CSS
		} else {
			props.Class = component.CSS
		}
	}
	
	return props
}

// extractClass extracts CSS class from component with fallback
func (str *SchemaTemplRenderer) extractClass(component TemplComponent, defaultClass string) string {
	class := defaultClass
	
	if propsMap, ok := component.Props.(map[string]interface{}); ok {
		if className, ok := propsMap["className"].(string); ok {
			if className != "" {
				class = className
			}
		}
	}
	
	// Append component CSS if available
	if component.CSS != "" {
		class += " " + component.CSS
	}
	
	return class
}