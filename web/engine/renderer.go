package engine

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/niiniyare/erp/web/layouts"
	"github.com/niiniyare/erp/web/schemas"
)

// SchemaRenderer implements the schema-to-template rendering engine
type SchemaRenderer struct {
	registry     *ComponentRegistry
	permissions  schemas.PermissionChecker
	dataProvider schemas.DataProvider
}

// NewSchemaRenderer creates a new schema renderer with the given dependencies
func NewSchemaRenderer(registry *ComponentRegistry, permissions schemas.PermissionChecker, dataProvider schemas.DataProvider) *SchemaRenderer {
	return &SchemaRenderer{
		registry:     registry,
		permissions:  permissions,
		dataProvider: dataProvider,
	}
}

// RenderPage converts a complete page schema into a renderable template
func (r *SchemaRenderer) RenderPage(schema *schemas.PageSchema, ctx context.Context) (templ.Component, error) {
	if schema == nil {
		return nil, schemas.NewSchemaError(schemas.ErrCodeValidationError, "page schema cannot be nil")
	}

	// Check page-level permissions
	if schema.Permissions != nil && r.permissions != nil {
		allowed, err := r.permissions.CheckPermissions(schema.Permissions, ctx)
		if err != nil {
			return nil, schemas.NewSchemaError(schemas.ErrCodePermissionDenied,
				fmt.Sprintf("permission check failed: %v", err)).WithPath("permissions")
		}
		if !allowed {
			return nil, schemas.NewSchemaError(schemas.ErrCodePermissionDenied,
				"insufficient permissions to access this page").WithPath("permissions")
		}
	}

	// Fetch data sources
	data := make(map[string]interface{})
	for _, ds := range schema.DataSources {
		if r.dataProvider != nil {
			value, err := r.dataProvider.GetData(&ds, ctx)
			if err != nil {
				return nil, schemas.NewSchemaError(schemas.ErrCodeDataSourceError,
					fmt.Sprintf("failed to fetch data source '%s': %v", ds.ID, err)).
					WithPath(fmt.Sprintf("dataSources.%s", ds.ID))
			}
			data[ds.ID] = value
		}
	}

	// Add data to context for component rendering
	ctx = context.WithValue(ctx, "schemaData", data)

	// Render page content
	pageContent, err := r.renderPageContent(schema.Components, ctx)
	if err != nil {
		return nil, err
	}

	// Wrap content in layout
	return r.RenderLayout(schema.Layout, pageContent, ctx)
}

// RenderComponent converts a component definition into a renderable template
func (r *SchemaRenderer) RenderComponent(def *schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	if def == nil {
		return nil, schemas.NewSchemaError(schemas.ErrCodeValidationError, "component definition cannot be nil")
	}

	// Check component-level permissions
	if def.Permissions != nil && r.permissions != nil {
		allowed, err := r.permissions.CheckPermissions(def.Permissions, ctx)
		if err != nil {
			return nil, schemas.NewSchemaError(schemas.ErrCodePermissionDenied,
				fmt.Sprintf("permission check failed: %v", err)).
				WithComponent(def.Type).WithPath("permissions")
		}
		if !allowed {
			// Check if there's a fallback component for permission denied
			if def.Permissions.Fallback != nil {
				return r.RenderComponent(def.Permissions.Fallback, ctx)
			}
			return nil, schemas.NewSchemaError(schemas.ErrCodePermissionDenied,
				"insufficient permissions to access this component").
				WithComponent(def.Type)
		}
	}

	// Check render conditions
	if len(def.Conditions) > 0 {
		shouldRender, err := r.evaluateConditions(def.Conditions, ctx)
		if err != nil {
			return nil, schemas.NewSchemaError(schemas.ErrCodeRenderError,
				fmt.Sprintf("failed to evaluate conditions: %v", err)).
				WithComponent(def.Type).WithPath("conditions")
		}
		if !shouldRender {
			return templ.NopComponent, nil
		}
	}

	// Inject data source data into props
	if def.DataSource != "" {
		data := r.getContextData(ctx, def.DataSource)
		if data != nil {
			if def.Props == nil {
				def.Props = make(map[string]interface{})
			}
			def.Props["data"] = data
		}
	}

	// Resolve and create the component
	component, err := r.registry.ResolveWithDefinition(def, ctx)
	if err != nil {
		return nil, err
	}

	// If component has children, render them and compose
	if len(def.Children) > 0 {
		children, err := r.renderComponents(def.Children, ctx)
		if err != nil {
			return nil, err
		}
		// For components with children, we need to create a wrapper
		return r.wrapWithChildren(component, children, def.Layout), nil
	}

	// Apply layout rules if specified
	if def.Layout != nil {
		return r.wrapWithLayout(component, def.Layout), nil
	}

	return component, nil
}

// RenderLayout wraps content in the specified layout
func (r *SchemaRenderer) RenderLayout(layoutType string, content templ.Component, ctx context.Context) (templ.Component, error) {
	if content == nil {
		content = templ.NopComponent
	}

	switch layoutType {
	case "app":
		props := r.getAppLayoutProps(ctx)
		return layouts.AppLayout(props, content), nil
	case "auth":
		props := r.getAuthLayoutProps(ctx)
		return layouts.AuthLayout(props, content), nil
	case "minimal":
		props := r.getMinimalLayoutProps(ctx)
		return layouts.MinimalLayout(props, content), nil
	case "base":
		props := r.getBaseLayoutProps(ctx)
		return layouts.BaseLayout(props, content), nil
	case "", "none":
		return content, nil
	default:
		// Try to resolve layout as a component
		layoutDef := &schemas.ComponentDefinition{
			Type: "layouts." + layoutType,
			Props: map[string]interface{}{
				"content": content,
			},
		}
		return r.registry.ResolveWithDefinition(layoutDef, ctx)
	}
}

// Helper methods

func (r *SchemaRenderer) renderPageContent(components []schemas.ComponentDefinition, ctx context.Context) (templ.Component, error) {
	if len(components) == 0 {
		return templ.NopComponent, nil
	}

	if len(components) == 1 {
		return r.RenderComponent(&components[0], ctx)
	}

	// Multiple components - wrap in a container
	renderedComponents, err := r.renderComponents(components, ctx)
	if err != nil {
		return nil, err
	}

	return r.createContainer(renderedComponents), nil
}

func (r *SchemaRenderer) renderComponents(components []schemas.ComponentDefinition, ctx context.Context) ([]templ.Component, error) {
	rendered := make([]templ.Component, 0, len(components))

	for i, comp := range components {
		component, err := r.RenderComponent(&comp, ctx)
		if err != nil {
			return nil, schemas.NewSchemaError(schemas.ErrCodeRenderError,
				fmt.Sprintf("failed to render component %d: %v", i, err)).
				WithPath(fmt.Sprintf("components[%d]", i))
		}
		rendered = append(rendered, component)
	}

	return rendered, nil
}

func (r *SchemaRenderer) evaluateConditions(conditions []schemas.RenderCondition, ctx context.Context) (bool, error) {
	for _, condition := range conditions {
		result, err := r.evaluateCondition(condition, ctx)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil // All conditions must be true
		}
	}
	return true, nil
}

func (r *SchemaRenderer) evaluateCondition(condition schemas.RenderCondition, ctx context.Context) (bool, error) {
	var value interface{}

	switch condition.Source {
	case "props":
		// Get from component props (this would need to be passed through context)
		value = r.getContextValue(ctx, "props."+condition.Field)
	case "data":
		// Get from data sources
		value = r.getContextData(ctx, condition.Field)
	case "user":
		// Get from user context
		value = r.getContextValue(ctx, "user."+condition.Field)
	case "context":
		// Get from general context
		value = r.getContextValue(ctx, condition.Field)
	default:
		// Default to context
		value = r.getContextValue(ctx, condition.Field)
	}

	return r.compareValues(value, condition.Operator, condition.Value), nil
}

func (r *SchemaRenderer) compareValues(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "eq", "==":
		return actual == expected
	case "ne", "!=":
		return actual != expected
	case "gt", ">":
		return r.compareNumeric(actual, expected, ">")
	case "lt", "<":
		return r.compareNumeric(actual, expected, "<")
	case "gte", ">=":
		return r.compareNumeric(actual, expected, ">=")
	case "lte", "<=":
		return r.compareNumeric(actual, expected, "<=")
	case "in":
		return r.valueInSlice(actual, expected)
	case "contains":
		return r.stringContains(actual, expected)
	case "exists":
		return actual != nil
	default:
		return false
	}
}

func (r *SchemaRenderer) compareNumeric(a, b interface{}, op string) bool {
	// Simple numeric comparison - would need more robust implementation
	af, aok := a.(float64)
	bf, bok := b.(float64)
	if !aok || !bok {
		return false
	}

	switch op {
	case ">":
		return af > bf
	case "<":
		return af < bf
	case ">=":
		return af >= bf
	case "<=":
		return af <= bf
	default:
		return false
	}
}

func (r *SchemaRenderer) valueInSlice(value, slice interface{}) bool {
	// Check if value is in slice - simplified implementation
	if s, ok := slice.([]interface{}); ok {
		for _, item := range s {
			if item == value {
				return true
			}
		}
	}
	return false
}

func (r *SchemaRenderer) stringContains(haystack, needle interface{}) bool {
	h, hok := haystack.(string)
	n, nok := needle.(string)
	if !hok || !nok {
		return false
	}
	return strings.Contains(h, n)
}

func (r *SchemaRenderer) getContextData(ctx context.Context, key string) interface{} {
	if data := ctx.Value("schemaData"); data != nil {
		if dataMap, ok := data.(map[string]interface{}); ok {
			return dataMap[key]
		}
	}
	return nil
}

func (r *SchemaRenderer) getContextValue(ctx context.Context, key string) interface{} {
	// This would need to be implemented based on how context values are structured
	// For now, return nil
	return nil
}

func (r *SchemaRenderer) wrapWithChildren(component templ.Component, children []templ.Component, layout *schemas.LayoutRules) templ.Component {
	// Create a wrapper that includes the component and its children
	// This is a simplified implementation
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := component.Render(ctx, w); err != nil {
			return err
		}
		for _, child := range children {
			if err := child.Render(ctx, w); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SchemaRenderer) wrapWithLayout(component templ.Component, layout *schemas.LayoutRules) templ.Component {
	if layout == nil {
		return component
	}

	// Apply layout rules by wrapping component in appropriate HTML structure
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		// Start container
		if err := r.writeLayoutStart(w, layout); err != nil {
			return err
		}

		// Render component
		if err := component.Render(ctx, w); err != nil {
			return err
		}

		// End container
		return r.writeLayoutEnd(w, layout)
	})
}

func (r *SchemaRenderer) writeLayoutStart(w io.Writer, layout *schemas.LayoutRules) error {
	classes := []string{}
	styles := make(map[string]string)

	// Add container classes
	switch layout.Container {
	case "grid":
		classes = append(classes, "grid")
		if layout.Grid != nil {
			if layout.Grid.Columns != "" {
				styles["grid-template-columns"] = layout.Grid.Columns
			}
			if layout.Grid.Rows != "" {
				styles["grid-template-rows"] = layout.Grid.Rows
			}
			if layout.Grid.Gap != "" {
				styles["gap"] = layout.Grid.Gap
			}
		}
	case "flex":
		classes = append(classes, "flex")
		if layout.Flex != nil {
			if layout.Flex.Direction != "" {
				classes = append(classes, "flex-"+layout.Flex.Direction)
			}
			if layout.Flex.Justify != "" {
				classes = append(classes, "justify-"+layout.Flex.Justify)
			}
			if layout.Flex.Align != "" {
				classes = append(classes, "items-"+layout.Flex.Align)
			}
			if layout.Flex.Gap != "" {
				styles["gap"] = layout.Flex.Gap
			}
		}
	case "block":
		classes = append(classes, "block")
	}

	// Add custom classes
	if len(layout.Classes) > 0 {
		classes = append(classes, layout.Classes...)
	}

	// Add spacing
	if layout.Spacing != nil {
		if layout.Spacing.Margin != "" {
			styles["margin"] = layout.Spacing.Margin
		}
		if layout.Spacing.Padding != "" {
			styles["padding"] = layout.Spacing.Padding
		}
	}

	// Add custom styles
	for k, v := range layout.Styles {
		styles[k] = v
	}

	// Write opening tag
	if _, err := w.Write([]byte("<div")); err != nil {
		return err
	}

	if len(classes) > 0 {
		classAttr := ` class="` + strings.Join(classes, " ") + `"`
		if _, err := w.Write([]byte(classAttr)); err != nil {
			return err
		}
	}

	if len(styles) > 0 {
		styleStr := ""
		for k, v := range styles {
			if styleStr != "" {
				styleStr += "; "
			}
			styleStr += k + ": " + v
		}
		styleAttr := ` style="` + styleStr + `"`
		if _, err := w.Write([]byte(styleAttr)); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte(">"))
	return err
}

func (r *SchemaRenderer) writeLayoutEnd(w io.Writer, layout *schemas.LayoutRules) error {
	_, err := w.Write([]byte("</div>"))
	return err
}

func (r *SchemaRenderer) createContainer(components []templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := w.Write([]byte(`<div class="space-y-6">`)); err != nil {
			return err
		}

		for _, comp := range components {
			if err := comp.Render(ctx, w); err != nil {
				return err
			}
		}

		_, err := w.Write([]byte("</div>"))
		return err
	})
}

// Layout props helpers

func (r *SchemaRenderer) getAppLayoutProps(ctx context.Context) layouts.AppLayoutProps {
	// Extract props from context - simplified implementation
	props := layouts.AppLayoutProps{
		Base: layouts.BaseLayoutProps{
			Title:       "ERP System",
			Description: "Enterprise Resource Planning System",
			ThemeMode:   "system",
		},
		SidebarOpen:      r.getBoolValue(ctx, "sidebarOpen", true),
		SidebarCollapsed: r.getBoolValue(ctx, "sidebarCollapsed", false),
		ShowBreadcrumbs:  r.getBoolValue(ctx, "showBreadcrumbs", true),
	}

	// Handle optional user field safely
	if userVal := r.getContextValue(ctx, "user"); userVal != nil {
		if user, ok := userVal.(*layouts.AppUser); ok {
			props.User = user
		}
	}

	return props
}

func (r *SchemaRenderer) getAuthLayoutProps(ctx context.Context) layouts.AuthLayoutProps {
	return layouts.AuthLayoutProps{
		Base: layouts.BaseLayoutProps{
			Title:       "Authentication - ERP System",
			Description: "Sign in to your ERP account",
			ThemeMode:   "system",
		},
		FormTitle: r.getStringValue(ctx, "formTitle", "Sign In"),
		AppName:   r.getStringValue(ctx, "appName", "ERP System"),
		AppLogo:   r.getStringValue(ctx, "appLogo", ""),
	}
}

func (r *SchemaRenderer) getMinimalLayoutProps(ctx context.Context) layouts.MinimalLayoutProps {
	return layouts.MinimalLayoutProps{
		Base: layouts.BaseLayoutProps{
			Title:       r.getStringValue(ctx, "title", "ERP System"),
			Description: r.getStringValue(ctx, "description", ""),
			ThemeMode:   "system",
		},
	}
}

func (r *SchemaRenderer) getBaseLayoutProps(ctx context.Context) layouts.BaseLayoutProps {
	return layouts.BaseLayoutProps{
		Title:       r.getStringValue(ctx, "title", "ERP System"),
		Description: r.getStringValue(ctx, "description", ""),
		ThemeMode:   "system",
	}
}

func (r *SchemaRenderer) getStringValue(ctx context.Context, key, defaultValue string) string {
	if val := r.getContextValue(ctx, key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (r *SchemaRenderer) getBoolValue(ctx context.Context, key string, defaultValue bool) bool {
	if val := r.getContextValue(ctx, key); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}
