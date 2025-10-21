package engine

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
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
	str.componentMap["ButtonGroup"] = str.renderButtonGroup
	str.componentMap["button-group"] = str.renderButtonGroup

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

// renderButton converts TemplComponent to a button
func (str *SchemaTemplRenderer) renderButton(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		text := str.extractString(component.Props, "text", "Button")
		variant := str.extractString(component.Props, "variant", "primary")

		class := fmt.Sprintf("px-4 py-2 rounded font-medium %s", str.getButtonClass(variant))

		html := fmt.Sprintf(`<button class="%s">%s</button>`, class, text)
		_, err := w.Write([]byte(html))
		return err
	})
}

// renderButtonGroup converts TemplComponent to a group of buttons
func (str *SchemaTemplRenderer) renderButtonGroup(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		text := str.extractString(component.Props, "text", "Button")
		variant := str.extractString(component.Props, "variant", "primary")

		class := fmt.Sprintf("px-4 py-2 rounded font-medium %s", str.getButtonClass(variant))

		html := fmt.Sprintf(`<div class="inline-flex"><button class="%s">%s</button></div>`, class, text)
		_, err := w.Write([]byte(html))
		return err
	})
}

// renderContainer renders a layout container
func (str *SchemaTemplRenderer) renderContainer(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		class := str.extractString(component.Props, "className", "container")

		html := fmt.Sprintf(`<div class="%s">Container Content</div>`, class)
		_, err := w.Write([]byte(html))
		return err
	})
}

// renderCard renders a card component
func (str *SchemaTemplRenderer) renderCard(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		title := str.extractString(component.Props, "title", "Card")
		content := str.extractString(component.Props, "content", "Card content")

		html := fmt.Sprintf(`
			<div class="bg-white border border-gray-200 rounded-lg shadow-sm overflow-hidden">
				<div class="px-4 py-5 sm:p-6">
					<h3 class="text-lg leading-6 font-medium text-gray-900 mb-2">%s</h3>
					<p class="text-sm text-gray-600">%s</p>
				</div>
			</div>
		`, title, content)
		_, err := w.Write([]byte(html))
		return err
	})
}

// renderPanel renders a panel component
func (str *SchemaTemplRenderer) renderPanel(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		title := str.extractString(component.Props, "title", "Panel")
		content := str.extractString(component.Props, "content", "Panel content")

		html := fmt.Sprintf(`
			<div class="bg-white shadow rounded-lg">
				<div class="px-4 py-5 sm:p-6">
					<h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">%s</h3>
					<p class="text-sm text-gray-600">%s</p>
				</div>
			</div>
		`, title, content)
		_, err := w.Write([]byte(html))
		return err
	})
}

// renderTable renders a data table
func (str *SchemaTemplRenderer) renderTable(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		html := `
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Schema</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					<tr>
						<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">Component</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">Schema-generated</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">Active</span>
						</td>
					</tr>
				</tbody>
			</table>
		`
		_, err := w.Write([]byte(html))
		return err
	})
}

// renderList renders a data list
func (str *SchemaTemplRenderer) renderList(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		html := `
			<ul class="divide-y divide-gray-200">
				<li class="py-3 flex">
					<div class="flex-1">
						<h4 class="text-sm font-medium text-gray-900">Schema-driven component</h4>
						<p class="text-sm text-gray-500">Generated from JSON schema</p>
					</div>
				</li>
			</ul>
		`
		_, err := w.Write([]byte(html))
		return err
	})
}

// renderFallback renders a fallback component for unknown types
func (str *SchemaTemplRenderer) renderFallback(component TemplComponent) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		html := fmt.Sprintf(
			`<div class="alert alert-warning">Unknown component type: %s</div>`,
			component.Type,
		)
		_, err := w.Write([]byte(html))
		return err
	})
}

// Helper functions
func (str *SchemaTemplRenderer) extractString(props interface{}, key, defaultValue string) string {
	if propsMap, ok := props.(map[string]interface{}); ok {
		if value, exists := propsMap[key]; exists {
			if str, ok := value.(string); ok {
				return str
			}
		}
	}
	return defaultValue
}

func (str *SchemaTemplRenderer) getButtonClass(variant string) string {
	switch variant {
	case "primary":
		return "text-white bg-blue-600 hover:bg-blue-700"
	case "secondary":
		return "text-gray-700 bg-gray-200 hover:bg-gray-300"
	case "success":
		return "text-white bg-green-600 hover:bg-green-700"
	case "danger":
		return "text-white bg-red-600 hover:bg-red-700"
	default:
		return "text-gray-700 bg-gray-200 hover:bg-gray-300"
	}
}
