package ui

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

// ============================================================================
// ENHANCED COMPONENT BUILDERS WITH STYLING
// ============================================================================

// CreateStyledButton creates a button with integrated CSS styling
func CreateStyledButton(id, text, variant string, schemaDir string) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	styles := factory.ButtonStyles(variant)
	
	return NewComponent(ComponentButton, id).
		WithLabel(text).
		WithStyles(styles).
		WithConfig(ButtonConfig{
			Text: text,
		})
}

// CreateStyledInput creates an input with integrated CSS styling
func CreateStyledInput(id, inputType, state string, schemaDir string) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	styles := factory.InputStyles(state)
	
	return NewComponent(ComponentInput, id).
		WithStyles(styles).
		WithConfig(InputConfig{
			InputType: InputType(inputType),
		})
}

// CreateStyledCard creates a card with integrated CSS styling
func CreateStyledCard(id, elevation string, schemaDir string) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	styles := factory.CardStyles(elevation)
	
	return NewComponent(ComponentCard, id).
		WithStyles(styles)
}

// CreateStyledContainer creates a container with integrated CSS styling
func CreateStyledContainer(id, maxWidth string, schemaDir string) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	styles := factory.ContainerStyles(maxWidth)
	
	return NewComponent(ComponentContainer, id).
		WithStyles(styles)
}

// CreateStyledTable creates a table with integrated CSS styling
func CreateStyledTable(id string, columns []TableColumn, variant string, schemaDir string) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	styles := factory.TableStyles(variant)
	
	return NewComponent(ComponentTable, id).
		WithStyles(styles).
		WithConfig(TableConfig{
			Columns: columns,
		})
}

// ============================================================================
// FLUENT STYLING API EXTENSIONS
// ============================================================================

// WithStyledVariant applies predefined styling variant to component
func (b *ComponentBuilder) WithStyledVariant(variant, schemaDir string) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	
	var styles *css.Styles
	
	switch b.component.Type {
	case ComponentButton:
		styles = factory.ButtonStyles(variant)
	case ComponentInput, ComponentTextarea:
		styles = factory.InputStyles(variant)
	case ComponentCard:
		styles = factory.CardStyles(variant)
	case ComponentContainer:
		styles = factory.ContainerStyles(variant)
	case ComponentTable:
		styles = factory.TableStyles(variant)
	default:
		// For unknown components, create basic styles
		styles = css.NewStyles(schemaDir)
	}
	
	return b.WithStyles(styles)
}

// WithCustomCSS adds custom CSS properties to component
func (b *ComponentBuilder) WithCustomCSS(property, value string) *ComponentBuilder {
	if b.component.Styles == nil {
		b.component.Styles = css.NewStyles("")
	}
	
	b.component.Styles.WithCustomProperty(property, value)
	return b
}

// WithResponsiveStyles applies responsive styling to component
func (b *ComponentBuilder) WithResponsiveStyles(schemaDir string, configurator func(*css.ResponsiveStyleBuilder)) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	builder := factory.ResponsiveStyles()
	
	configurator(builder)
	
	return b.WithStyles(builder.Build())
}

// WithThemeStyles applies theme-aware styling to component
func (b *ComponentBuilder) WithThemeStyles(theme, schemaDir string, configurator func(*css.ThemeStyleBuilder)) *ComponentBuilder {
	factory := css.NewFactory(schemaDir)
	builder := factory.ThemeStyles(theme)
	
	configurator(builder)
	
	return b.WithStyles(builder.Build())
}

// ============================================================================
// COMPONENT VALIDATION WITH STYLING
// ============================================================================

// ValidateComponentWithStyles validates both component structure and CSS styles
func ValidateComponentWithStyles(ctx context.Context, registry ComponentRegistry, component Component, schemaDir string) []ValidationError {
	var errors []ValidationError
	
	// Validate component structure
	if err := registry.Validate(ctx, component); err != nil {
		errors = append(errors, ValidationError{
			Component: component.ID,
			Field:     "structure",
			Message:   err.Error(),
		})
	}
	
	// Validate CSS styles if present
	if component.Styles != nil {
		validator := css.NewValidator(schemaDir)
		cssErrors := validator.ValidateStyles(component.Styles)
		
		for _, cssErr := range cssErrors {
			errors = append(errors, ValidationError{
				Component: component.ID,
				Field:     fmt.Sprintf("styles.%s", cssErr.Property),
				Message:   cssErr.Message,
			})
		}
	}
	
	// Recursively validate children
	for _, child := range component.Children {
		childErrors := ValidateComponentWithStyles(ctx, registry, child, schemaDir)
		errors = append(errors, childErrors...)
	}
	
	return errors
}

// ValidationError represents component and styling validation errors
type ValidationError struct {
	Component string `json:"component"`
	Field     string `json:"field"`
	Message   string `json:"message"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("component '%s' field '%s': %s", ve.Component, ve.Field, ve.Message)
}

// ============================================================================
// COMPONENT RENDERING WITH STYLES
// ============================================================================

// ComponentRenderer handles rendering components with integrated styling
type ComponentRenderer struct {
	schemaDir string
}

// NewComponentRenderer creates a new component renderer
func NewComponentRenderer(schemaDir string) *ComponentRenderer {
	return &ComponentRenderer{
		schemaDir: schemaDir,
	}
}

// RenderComponent renders a component with its styles
func (r *ComponentRenderer) RenderComponent(component Component) (RenderedComponent, error) {
	rendered := RenderedComponent{
		ID:       component.ID,
		Type:     string(component.Type),
		Label:    component.Label,
		Class:    component.Class,
		Children: make([]RenderedComponent, 0, len(component.Children)),
	}
	
	// Add CSS styles if present
	if component.Styles != nil {
		rendered.CSS = component.Styles.ToCSS()
		rendered.CSSClass = component.Styles.ToCSSClass(fmt.Sprintf("component-%s", component.ID))
	}
	
	// Add custom class if specified
	if component.Class != "" {
		if rendered.CSS != "" {
			rendered.CSS = fmt.Sprintf("%s; %s", rendered.CSS, component.Class)
		} else {
			rendered.CSS = component.Class
		}
	}
	
	// Parse component configuration
	if len(component.Config) > 0 {
		var config map[string]interface{}
		if err := json.Unmarshal(component.Config, &config); err != nil {
			return rendered, fmt.Errorf("failed to parse component config: %w", err)
		}
		rendered.Config = config
	}
	
	// Recursively render children
	for _, child := range component.Children {
		childRendered, err := r.RenderComponent(child)
		if err != nil {
			return rendered, fmt.Errorf("failed to render child component: %w", err)
		}
		rendered.Children = append(rendered.Children, childRendered)
	}
	
	return rendered, nil
}

// RenderedComponent represents a component ready for template rendering
type RenderedComponent struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Label    string                 `json:"label"`
	Class    string                 `json:"class"`
	CSS      string                 `json:"css"`
	CSSClass string                 `json:"css_class"`
	Config   map[string]interface{} `json:"config"`
	Children []RenderedComponent    `json:"children"`
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// MergeStyles merges multiple CSS styles objects
func MergeStyles(schemaDir string, styles ...*css.Styles) *css.Styles {
	merged := css.NewStyles(schemaDir)
	
	for _, style := range styles {
		if style == nil {
			continue
		}
		
		// Merge all properties
		properties := style.GetAllProperties()
		for prop, value := range properties {
			if value != "" {
				// Apply the property to merged styles using reflection or direct assignment
				// This would require extending the Styles struct with setter methods
				merged.WithCustomProperty(prop, value)
			}
		}
		
		// Merge custom properties
		for prop, value := range style.Custom {
			merged.WithCustomProperty(prop, value)
		}
	}
	
	return merged
}

// GenerateStylesheet generates a complete CSS stylesheet from components
func GenerateStylesheet(components []Component) string {
	var cssRules []string
	
	for _, component := range components {
		if component.Styles != nil {
			cssClass := component.Styles.ToCSSClass(fmt.Sprintf("component-%s", component.ID))
			if cssClass != "" {
				cssRules = append(cssRules, cssClass)
			}
		}
		
		// Recursively process children
		if len(component.Children) > 0 {
			childStylesheet := GenerateStylesheet(component.Children)
			if childStylesheet != "" {
				cssRules = append(cssRules, childStylesheet)
			}
		}
	}
	
	if len(cssRules) == 0 {
		return ""
	}
	
	return fmt.Sprintf("/* Generated Component Styles */\n%s", 
		fmt.Sprintf("%s\n", cssRules))
}

// ExtractInlineStyles extracts inline styles from components for CSP compliance
func ExtractInlineStyles(components []Component) (map[string]string, []Component) {
	inlineStyles := make(map[string]string)
	cleanedComponents := make([]Component, 0, len(components))
	
	for _, component := range components {
		cleaned := component
		
		if component.Styles != nil && component.Styles.ToCSS() != "" {
			className := fmt.Sprintf("component-%s", component.ID)
			inlineStyles[className] = component.Styles.ToCSS()
			
			// Remove inline styles and add class name
			cleaned.Styles = nil
			if cleaned.Class == "" {
				cleaned.Class = className
			} else {
				cleaned.Class = fmt.Sprintf("%s %s", cleaned.Class, className)
			}
		}
		
		// Recursively process children
		if len(component.Children) > 0 {
			childStyles, cleanedChildren := ExtractInlineStyles(component.Children)
			for className, styles := range childStyles {
				inlineStyles[className] = styles
			}
			cleaned.Children = cleanedChildren
		}
		
		cleanedComponents = append(cleanedComponents, cleaned)
	}
	
	return inlineStyles, cleanedComponents
}

// ============================================================================
// EXAMPLE USAGE PATTERNS
// ============================================================================

// ExampleCreateStyledForm demonstrates creating a complete styled form
func ExampleCreateStyledForm(schemaDir string) Component {
	// Create form with styling
	form := NewComponent(ComponentForm, "user-form").
		WithLabel("User Registration Form").
		WithStyledVariant("default", schemaDir).
		WithResponsiveStyles(schemaDir, func(builder *css.ResponsiveStyleBuilder) {
			builder.Base(css.NewStyles(schemaDir).WithPadding("1rem")).
				MD(func(s *css.Styles) {
					s.WithPadding("2rem")
				})
		}).
		WithChildren(
			// Styled input fields
			CreateStyledInput("first-name", "text", "default", schemaDir).
				WithLabel("First Name").
				WithName("first_name").
				Required().
				Build(),
			
			CreateStyledInput("email", "email", "default", schemaDir).
				WithLabel("Email Address").
				WithName("email").
				Required().
				Build(),
			
			// Styled submit button
			CreateStyledButton("submit", "Create Account", "primary", schemaDir).
				WithCustomCSS("margin-top", "1rem").
				Build(),
		).
		Build()
	
	return form
}