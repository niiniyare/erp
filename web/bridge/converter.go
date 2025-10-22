package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/pkg/schema/ui/css"
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/web/components/organisms/tree"
)

// Bridge connects the schema system with Templ components.
// It provides conversion between schema definitions and renderable components.
type Bridge struct {
	registry   schemaui.ComponentRegistry
	cssFactory *css.Factory
	converters map[schemaui.ComponentType]converterFunc
}

// converterFunc defines the signature for component conversion functions.
type converterFunc func(schemaui.Component) (TemplComponent, error)

// TemplComponent represents a converted Templ component with its props.
type TemplComponent struct {
	Type  string `json:"type"`
	Props any    `json:"props"`
}

// NewBridge creates a new bridge instance with the conversion registry initialized.
func NewBridge() *Bridge {
	b := &Bridge{
		registry:   schemaui.NewRegistry(),
		cssFactory: css.NewFactory("docs/ui/Schema"),
		converters: make(map[schemaui.ComponentType]converterFunc),
	}
	b.registerConverters()
	return b
}

// registerConverters initializes the component converter registry.
func (b *Bridge) registerConverters() {
	b.converters[schemaui.ComponentButton] = b.convertButton
	b.converters[schemaui.ComponentInput] = b.convertInput
	b.converters[schemaui.ComponentTextarea] = b.convertTextarea
	b.converters[schemaui.ComponentSelect] = b.convertSelect
	b.converters[schemaui.ComponentCheckbox] = b.convertCheckbox
	b.converters[schemaui.ComponentRadio] = b.convertRadio
	b.converters[schemaui.ComponentTree] = b.convertTree
}

// ConvertToTempl converts a schema component to its Templ equivalent.
func (b *Bridge) ConvertToTempl(ctx context.Context, component schemaui.Component) (TemplComponent, error) {
	converter, exists := b.converters[component.Type]
	if !exists {
		return TemplComponent{}, fmt.Errorf("unsupported component type: %s", component.Type)
	}

	return converter(component)
}

// ConvertBatch converts multiple schema components to Templ components.
func (b *Bridge) ConvertBatch(ctx context.Context, components []schemaui.Component) ([]TemplComponent, error) {
	result := make([]TemplComponent, 0, len(components))

	for i, component := range components {
		converted, err := b.ConvertToTempl(ctx, component)
		if err != nil {
			return nil, fmt.Errorf("component[%d] (id=%s): %w", i, component.ID, err)
		}
		result = append(result, converted)
	}

	return result, nil
}

// GenerateTemplCode generates Go Templ template code from schema components.
func (b *Bridge) GenerateTemplCode(ctx context.Context, components []schemaui.Component, templateName string) (string, error) {
	converted, err := b.ConvertBatch(ctx, components)
	if err != nil {
		return "", fmt.Errorf("convert batch: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("templ %s() {\n", templateName))

	for _, tc := range converted {
		componentCall := b.formatComponentCall(tc)
		sb.WriteString(fmt.Sprintf("\t%s\n", componentCall))
	}

	sb.WriteString("}\n")
	return sb.String(), nil
}

// formatComponentCall formats a component call for Templ template generation.
func (b *Bridge) formatComponentCall(tc TemplComponent) string {
	// Note: In production, use ast/format for proper code generation
	return fmt.Sprintf("@atoms.%s(...)", tc.Type)
}

// ============================================================================
// BUTTON CONVERTER
// ============================================================================

func (b *Bridge) convertButton(component schemaui.Component) (TemplComponent, error) {
	var config schemaui.ButtonConfig
	if err := unmarshalConfig(component.Config, &config); err != nil {
		return TemplComponent{}, err
	}

	props := atoms.ButtonProps{
		BaseProps: atoms.BaseProps{
			ID:    component.ID,
			Class: b.buildClass(component),
		},
		InteractionProps: atoms.InteractionProps{
			Disabled: component.Disabled,
		},
		AccessibilityProps: atoms.AccessibilityProps{
			AriaLabel: component.AriaLabel,
		},
		AlpinEventHandlers: atoms.AlpinEventHandlers{
			OnClick: component.OnClick,
		},
		Text:         config.Text,
		Icon:         atoms.IconProps{Name: config.Icon, Position: atoms.IconPosition(config.IconPosition)},
		IconPosition: atoms.IconPosition(config.IconPosition),
		Type:         convertButtonType(config.ButtonType),
		Variant:      convertVariant[atoms.ButtonVariant](component.Variant),
		Size:         convertSize[atoms.ButtonSize](component.Size),
		Loading:      config.Loading,
	}

	return TemplComponent{Type: "Button", Props: props}, nil
}

func convertButtonType(bt schemaui.ButtonType) string {
	switch bt {
	case schemaui.ButtonSubmit:
		return "submit"
	case schemaui.ButtonReset:
		return "reset"
	default:
		return "button"
	}
}

func getRadioLayout(direction string) string {
	if direction == "horizontal" {
		return "horizontal"
	}
	return "vertical"
}

func convertInputType(it schemaui.InputType) atoms.InputType {
	switch it {
	case schemaui.InputText:
		return atoms.InputText
	case schemaui.InputPassword:
		return atoms.InputPassword
	case schemaui.InputEmail:
		return atoms.InputEmail
	case schemaui.InputNumber:
		return atoms.InputNumber
	case schemaui.InputTel:
		return atoms.InputTel
	case schemaui.InputURL:
		return atoms.InputURL
	case schemaui.InputSearch:
		return atoms.InputTypeSearch
	case schemaui.InputHidden:
		return atoms.InputText // Map hidden to text since InputHidden doesn't exist
	default:
		return atoms.InputText
	}
}

// ============================================================================
// INPUT CONVERTER
// ============================================================================

func (b *Bridge) convertInput(component schemaui.Component) (TemplComponent, error) {
	var config schemaui.InputConfig
	if err := unmarshalConfig(component.Config, &config); err != nil {
		return TemplComponent{}, err
	}

	props := atoms.InputProps{
		BaseProps: atoms.BaseProps{
			ID:    component.ID,
			Name:  component.Name,
			Class: b.buildClass(component),
		},
		AccessibilityProps: atoms.AccessibilityProps{
			AriaLabel: component.AriaLabel,
		},
		InteractionProps: atoms.InteractionProps{
			Required: component.Required,
			Disabled: component.Disabled,
			ReadOnly: component.ReadOnly,
		},
		PlaceholderProps: atoms.PlaceholderProps{
			Placeholder: config.Placeholder,
		},
		Type:  convertInputType(config.InputType),
		Value: config.Value,
		Size:  convertSize[atoms.InputSize](component.Size),
	}

	// Apply validation rules
	if v := component.Validator; v != nil {
		props.InteractionProps.Required = v.Required
		if v.MaxLength != nil {
			props.MaxLength = *v.MaxLength
		}
		if v.MinLength != nil {
			props.MinLength = *v.MinLength
		}
		props.Pattern = v.Pattern
	}

	return TemplComponent{Type: "Input", Props: props}, nil
}

// ============================================================================
// TEXTAREA CONVERTER
// ============================================================================

func (b *Bridge) convertTextarea(component schemaui.Component) (TemplComponent, error) {
	var config schemaui.TextareaConfig
	if err := unmarshalConfig(component.Config, &config); err != nil {
		return TemplComponent{}, err
	}

	props := atoms.TextareaProps{
		BaseProps: atoms.BaseProps{
			ID:    component.ID,
			Name:  component.Name,
			Class: b.buildClass(component),
		},
		AccessibilityProps: atoms.AccessibilityProps{
			AriaLabel: component.AriaLabel,
		},
		InteractionProps: atoms.InteractionProps{
			Required: component.Required,
			Disabled: component.Disabled,
			ReadOnly: component.ReadOnly,
		},
		PlaceholderProps: atoms.PlaceholderProps{
			Placeholder: config.Placeholder,
		},
		Value:         config.Value,
		Rows:          config.Rows,
		Cols:          config.Cols,
		Resizable:     config.Resize != "none",
		MaxLength:     config.MaxLength,
		ComponentSize: convertSize[atoms.TextareaSize](component.Size),
	}

	return TemplComponent{Type: "Textarea", Props: props}, nil
}

// ============================================================================
// SELECT CONVERTER
// ============================================================================

func (b *Bridge) convertSelect(component schemaui.Component) (TemplComponent, error) {
	var config schemaui.SelectConfig
	if err := unmarshalConfig(component.Config, &config); err != nil {
		return TemplComponent{}, err
	}

	options := make([]atoms.SelectOption, len(config.Options))
	for i, opt := range config.Options {
		options[i] = atoms.SelectOption{
			Value:    opt.Value,
			Label:    opt.Label,
			Disabled: opt.Disabled,
		}
	}

	props := atoms.SelectProps{
		BaseProps: atoms.BaseProps{
			ID:    component.ID,
			Name:  component.Name,
			Class: b.buildClass(component),
		},
		AccessibilityProps: atoms.AccessibilityProps{
			AriaLabel: component.AriaLabel,
		},
		InteractionProps: atoms.InteractionProps{
			Required: component.Required,
			Disabled: component.Disabled,
		},
		PlaceholderProps: atoms.PlaceholderProps{
			Placeholder: config.Placeholder,
		},
		Options:  options,
		Value:    config.Value,
		Multiple: config.Multiple,
		Size:     convertSize[atoms.SelectSize](component.Size),
	}

	return TemplComponent{Type: "Select", Props: props}, nil
}

// ============================================================================
// CHECKBOX CONVERTER
// ============================================================================

func (b *Bridge) convertCheckbox(component schemaui.Component) (TemplComponent, error) {
	var config schemaui.CheckboxConfig
	if err := unmarshalConfig(component.Config, &config); err != nil {
		return TemplComponent{}, err
	}

	label := config.Label
	if label == "" {
		label = component.Label
	}

	props := atoms.CheckboxProps{
		BaseProps: atoms.BaseProps{
			ID:    component.ID,
			Name:  component.Name,
			Class: b.buildClass(component),
		},
		AccessibilityProps: atoms.AccessibilityProps{
			AriaLabel: component.AriaLabel,
		},
		InteractionProps: atoms.InteractionProps{
			Required: component.Required,
			Disabled: component.Disabled,
		},
		LabelProps: atoms.LabelProps{
			Label: label,
		},
		Checked:   config.Checked,
		Value:     fmt.Sprintf("%v", config.Value),
		Size:      convertSize[atoms.CheckboxSize](component.Size),
	}

	return TemplComponent{Type: "Checkbox", Props: props}, nil
}

// ============================================================================
// RADIO CONVERTER
// ============================================================================

func (b *Bridge) convertRadio(component schemaui.Component) (TemplComponent, error) {
	var config schemaui.RadioConfig
	if err := unmarshalConfig(component.Config, &config); err != nil {
		return TemplComponent{}, err
	}

	options := make([]atoms.RadioOption, len(config.Options))
	for i, opt := range config.Options {
		options[i] = atoms.RadioOption{
			Value:    opt.Value,
			Label:    opt.Label,
			Disabled: opt.Disabled,
		}
	}

	props := atoms.RadioGroupProps{
		BaseProps: atoms.BaseProps{
			Name:  component.Name,
			Class: b.buildClass(component),
		},
		InteractionProps: atoms.InteractionProps{
			Required: component.Required,
			Disabled: component.Disabled,
		},
		Options:  options,
		Value:    config.Value,
		Layout:   getRadioLayout(config.Direction),
		Label:    component.Label,
	}

	return TemplComponent{Type: "RadioGroup", Props: props}, nil
}

// ============================================================================
// TREE CONVERTER
// ============================================================================

func (b *Bridge) convertTree(component schemaui.Component) (TemplComponent, error) {
	var config schemaui.TreeConfig
	if err := unmarshalConfig(component.Config, &config); err != nil {
		return TemplComponent{}, err
	}

	// Convert schema nodes to tree nodes
	treeNodes := make([]tree.TreeNode, len(config.Data))
	for i, node := range config.Data {
		treeNodes[i] = convertSchemaNodeToTreeNode(node)
	}

	props := tree.TreeProps{
		BaseProps: atoms.BaseProps{
			ID:    component.ID,
			Class: b.buildClass(component),
		},
		AccessibilityProps: atoms.AccessibilityProps{
			AriaLabel: component.AriaLabel,
		},
		Data: treeNodes,
		Config: tree.TreeConfig{
			Selectable:   config.Selectable,
			MultiSelect:  config.Multiple,
			Checkboxes:   config.Checkable,
			Expandable:   config.Expandable,
			ShowLines:    config.ShowLine,
			ShowIcons:    config.ShowIcon,
			Draggable:    config.Draggable,
			Virtual:      config.VirtualScroll,
			MaxHeight:    config.Height,
			// Set sensible defaults for missing schema fields
			Cascade:      true,
			Accordion:    false,
			ShowRoot:     true,
			ShowTooltips: false,
			Sortable:     false,
			Searchable:   false,
			Filterable:   false,
			LazyLoad:     false,
			Indent:       "1.5rem",
			NodeHeight:   "2.5rem",
			ShowHeaders:  false, // Simple tree doesn't have columns/headers
		},
		Title:       component.Label,
		Description: component.Description,
		Styling: tree.TreeStyling{
			Variant:  convertVariant[atoms.Variant](component.Variant),
			Size:     convertSize[atoms.Size](component.Size),
			Rounded:  true,
			Shadow:   false,
			Bordered: false,
			Striped:  false,
			Hover:    true,
		},
	}

	// Set default expanded nodes if provided
	if len(config.DefaultExpanded) > 0 {
		props.Config.DefaultExpanded = config.DefaultExpanded
	}

	return TemplComponent{Type: "Tree", Props: props}, nil
}

// convertSchemaNodeToTreeNode converts a schema tree node to a tree component node
func convertSchemaNodeToTreeNode(node schemaui.TreeNode) tree.TreeNode {
	// Convert children recursively
	children := make([]tree.TreeNode, len(node.Children))
	for i, child := range node.Children {
		children[i] = convertSchemaNodeToTreeNode(child)
	}

	return tree.TreeNode{
		ID:          node.Key,    // Map Key to ID
		Label:       node.Title,  // Map Title to Label
		Children:    children,
		Disabled:    node.Disabled,
		Selected:    false, // Not available in schema
		HasChildren: len(children) > 0,
		IsLeaf:      node.IsLeaf,
		Icon:        node.Icon,
		// Set defaults for fields not in schema
		Level:       0, // Will be calculated by tree component
		Expanded:    false,
		Hidden:      false,
		Data:        make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}
}


// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// unmarshalConfig unmarshals component config with context.
func unmarshalConfig(data json.RawMessage, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

// buildClass constructs the final CSS class string for a component.
func (b *Bridge) buildClass(component schemaui.Component) string {
	if component.Styles == nil {
		return component.Class
	}

	cssClasses, err := b.cssFactory.GenerateCSS(*component.Styles)
	if err != nil || cssClasses == "" {
		return component.Class
	}

	return joinClasses(component.Class, cssClasses)
}

// joinClasses concatenates CSS class strings, handling empty values.
func joinClasses(classes ...string) string {
	var result strings.Builder
	for i, class := range classes {
		if class == "" {
			continue
		}
		if result.Len() > 0 {
			result.WriteByte(' ')
		}
		result.WriteString(class)
		_ = i // avoid unused variable
	}
	return result.String()
}

// convertSize is a generic size converter for all component types.
func convertSize[T any](size schemaui.Size) T {
	// All size types are aliases for atoms.Size, so we can directly convert
	switch size {
	case schemaui.SizeXS, schemaui.SizeSM:
		return any(atoms.SizeSM).(T)
	case schemaui.SizeLG:
		return any(atoms.SizeLG).(T)
	case schemaui.SizeXL:
		return any(atoms.SizeXL).(T)
	default:
		return any(atoms.SizeMD).(T)
	}
}

// convertVariant converts schema variant to component-specific variant type.
func convertVariant[T any](variant schemaui.Variant) T {
	// For atoms components, all variants are just atoms.Variant
	switch variant {
	case schemaui.VariantPrimary:
		return any(atoms.VariantPrimary).(T)
	case schemaui.VariantSecondary:
		return any(atoms.VariantSecondary).(T)
	case schemaui.VariantSuccess:
		return any(atoms.VariantDefault).(T) // Map to default as success doesn't exist
	case schemaui.VariantDanger:
		return any(atoms.VariantDefault).(T) // Map to default as danger doesn't exist
	case schemaui.VariantWarning:
		return any(atoms.VariantWarning).(T)
	case schemaui.VariantInfo:
		return any(atoms.VariantDefault).(T) // Map to default as info doesn't exist
	case schemaui.VariantLight:
		return any(atoms.VariantLight).(T)
	case schemaui.VariantDark:
		return any(atoms.VariantDark).(T)
	default:
		return any(atoms.VariantPrimary).(T)
	}
}
