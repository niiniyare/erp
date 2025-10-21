package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/pkg/schema/ui/css"
	"github.com/niiniyare/erp/web/components/atoms"
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
		Text:         config.Text,
		Icon:         config.Icon,
		IconPosition: string(config.IconPosition),
		Type:         convertButtonType(config.ButtonType),
		Variant:      convertVariant[atoms.ButtonVariant](component.Variant),
		Size:         convertSize[atoms.ButtonSize](component.Size),
		Loading:      config.Loading,
		Disabled:     component.Disabled,
		ID:           component.ID,
		Class:        b.buildClass(component),
		AriaLabel:    component.AriaLabel,
		OnClick:      component.OnClick,
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
		return atoms.InputSearch
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
		Type:        convertInputType(config.InputType),
		Value:       config.Value,
		Placeholder: config.Placeholder,
		Size:        convertSize[atoms.InputSize](component.Size),
		Required:    component.Required,
		Disabled:    component.Disabled,
		ReadOnly:    component.ReadOnly,
		ID:          component.ID,
		Name:        component.Name,
		Class:       b.buildClass(component),
		AriaLabel:   component.AriaLabel,
	}

	// Apply validation rules
	if v := component.Validator; v != nil {
		props.Required = v.Required
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
		Value:       config.Value,
		Placeholder: config.Placeholder,
		Rows:        config.Rows,
		Cols:        config.Cols,
		Resizable:   config.Resize != "none",
		MaxLength:   config.MaxLength,
		Size:        convertSize[atoms.TextareaSize](component.Size),
		Required:    component.Required,
		Disabled:    component.Disabled,
		ReadOnly:    component.ReadOnly,
		ID:          component.ID,
		Name:        component.Name,
		Class:       b.buildClass(component),
		AriaLabel:   component.AriaLabel,
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
		Options:     options,
		Value:       config.Value,
		Multiple:    config.Multiple,
		Placeholder: config.Placeholder,
		SelectSize:  convertSize[atoms.SelectSize](component.Size),
		Required:    component.Required,
		Disabled:    component.Disabled,
		ID:          component.ID,
		Name:        component.Name,
		Class:       b.buildClass(component),
		AriaLabel:   component.AriaLabel,
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
		Label:     label,
		Checked:   config.Checked,
		Value:     fmt.Sprintf("%v", config.Value),
		Size:      convertSize[atoms.CheckboxSize](component.Size),
		Required:  component.Required,
		Disabled:  component.Disabled,
		ID:        component.ID,
		Name:      component.Name,
		Class:     b.buildClass(component),
		AriaLabel: component.AriaLabel,
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
		Options:  options,
		Value:    config.Value,
		Layout:   getRadioLayout(config.Direction),
		Size:     convertSize[atoms.RadioSize](component.Size),
		Required: component.Required,
		Disabled: component.Disabled,
		Name:     component.Name,
		Class:    b.buildClass(component),
		Label:    component.Label,
	}

	return TemplComponent{Type: "RadioGroup", Props: props}, nil
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
	// Size constants mapping
	var (
		sm, md, lg T
	)

	// Use type assertion to get proper size constants
	switch any(sm).(type) {
	case atoms.ButtonSize:
		sm, md, lg = any(atoms.ButtonSizeSM).(T), any(atoms.ButtonSizeMD).(T), any(atoms.ButtonSizeLG).(T)
	case atoms.InputSize:
		sm, md, lg = any(atoms.InputSizeSM).(T), any(atoms.InputSizeMD).(T), any(atoms.InputSizeLG).(T)
	case atoms.TextareaSize:
		sm, md, lg = any(atoms.TextareaSizeSM).(T), any(atoms.TextareaSizeMD).(T), any(atoms.TextareaSizeLG).(T)
	case atoms.SelectSize:
		sm, md, lg = any(atoms.SelectSizeSM).(T), any(atoms.SelectSizeMD).(T), any(atoms.SelectSizeLG).(T)
	case atoms.CheckboxSize:
		sm, md, lg = any(atoms.CheckboxSizeSM).(T), any(atoms.CheckboxSizeMD).(T), any(atoms.CheckboxSizeLG).(T)
	case atoms.RadioSize:
		sm, md, lg = any(atoms.RadioSizeSM).(T), any(atoms.RadioSizeMD).(T), any(atoms.RadioSizeLG).(T)
	}

	switch size {
	case schemaui.SizeXS, schemaui.SizeSM:
		return sm
	case schemaui.SizeLG:
		return lg
	case schemaui.SizeXL:
		return lg // Map XL to LG since XL doesn't exist
	default:
		return md
	}
}

// convertVariant converts schema variant to component-specific variant type.
func convertVariant[T any](variant schemaui.Variant) T {
	var primary, secondary, success, danger, warning, info, light, dark T

	// Use type assertion to get proper variant constants
	switch any(primary).(type) {
	case atoms.ButtonVariant:
		primary = any(atoms.ButtonPrimary).(T)
		secondary = any(atoms.ButtonSecondary).(T)
		success = any(atoms.ButtonSuccess).(T)
		danger = any(atoms.ButtonDanger).(T)
		warning = any(atoms.ButtonWarning).(T)
		info = any(atoms.ButtonInfo).(T)
		light = any(atoms.ButtonLight).(T)
		dark = any(atoms.ButtonDark).(T)
	}

	switch variant {
	case schemaui.VariantPrimary:
		return primary
	case schemaui.VariantSecondary:
		return secondary
	case schemaui.VariantSuccess:
		return success
	case schemaui.VariantDanger:
		return danger
	case schemaui.VariantWarning:
		return warning
	case schemaui.VariantInfo:
		return info
	case schemaui.VariantLight:
		return light
	case schemaui.VariantDark:
		return dark
	default:
		return primary
	}
}
