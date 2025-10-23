package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/web/components/atoms"
)

// ============================================================================
// TEMPL TO SCHEMA CONVERSION
// ============================================================================

// reverseConverter defines the interface for Templ to Schema conversion.
type reverseConverter func(any) (schemaui.Component, error)

// reverseConverters holds the registered reverse converters.
var reverseConverters = map[string]reverseConverter{
	"Button":   convertButtonToSchema,
	"Input":    convertInputToSchema,
	"Textarea": convertTextareaToSchema,
	"Select":   convertSelectToSchema,
	"Checkbox": convertCheckboxToSchema,
	"Radio":    convertRadioToSchema,
}

// ConvertTemplToSchema converts a Templ component to its schema equivalent.
func (b *Bridge) ConvertTemplToSchema(ctx context.Context, component TemplComponent) (schemaui.Component, error) {
	converter, exists := reverseConverters[component.Type]
	if !exists {
		return schemaui.Component{}, fmt.Errorf("unsupported Templ component: %s", component.Type)
	}

	return converter(component.Props)
}

// ConvertBatchToSchema converts multiple Templ components to schema components.
func (b *Bridge) ConvertBatchToSchema(ctx context.Context, components []TemplComponent) ([]schemaui.Component, error) {
	result := make([]schemaui.Component, 0, len(components))

	for i, component := range components {
		converted, err := b.ConvertTemplToSchema(ctx, component)
		if err != nil {
			return nil, fmt.Errorf("component[%d] (type=%s): %w", i, component.Type, err)
		}
		result = append(result, converted)
	}

	return result, nil
}

// ============================================================================
// BUTTON CONVERTER
// ============================================================================

func convertButtonToSchema(props any) (schemaui.Component, error) {
	p, ok := props.(atoms.ButtonProps)
	if !ok {
		return schemaui.Component{}, fmt.Errorf("invalid props type for Button")
	}

	config := schemaui.ButtonConfig{
		Text:         p.Text,
		Icon:         p.Icon.Name,
		IconPosition: stringToPosition(string(p.IconPosition)),
		ButtonType:   stringToButtonType(p.Type),
		Loading:      p.Loading,
	}

	return buildComponent(schemaui.ComponentButton, baseBuilder{
		id:        p.BaseProps.ID,
		label:     p.Text,
		class:     p.BaseProps.Class,
		size:      reverseConvertSize[atoms.ButtonSize](p.Size),
		variant:   reverseConvertVariant(p.Variant),
		disabled:  p.InteractionProps.Disabled,
		ariaLabel: p.AccessibilityProps.AriaLabel,
		onClick:   p.AlpinEventHandlers.OnClick,
	}, config)
}

// ============================================================================
// INPUT CONVERTER
// ============================================================================

func convertInputToSchema(props any) (schemaui.Component, error) {
	p, ok := props.(atoms.InputProps)
	if !ok {
		return schemaui.Component{}, fmt.Errorf("invalid props type for Input")
	}

	config := schemaui.InputConfig{
		InputType:   stringToInputType(string(p.Type)),
		Value:       p.Value,
		Placeholder: p.PlaceholderProps.Placeholder,
		MaxLength:   p.MaxLength,
		MinLength:   p.MinLength,
		Pattern:     p.Pattern,
	}

	validator := buildValidator(p.InteractionProps.Required, p.MinLength, p.MaxLength, p.Pattern)

	return buildComponent(schemaui.ComponentInput, baseBuilder{
		id:          p.BaseProps.ID,
		name:        p.BaseProps.Name,
		label:       "", // Input doesn't have direct label field
		placeholder: p.PlaceholderProps.Placeholder,
		class:       p.BaseProps.Class,
		size:        reverseConvertSize[atoms.InputSize](p.Size),
		disabled:    p.InteractionProps.Disabled,
		required:    p.InteractionProps.Required,
		readOnly:    p.InteractionProps.ReadOnly,
		ariaLabel:   p.AccessibilityProps.AriaLabel,
		onChange:    p.AlpinEventHandlers.OnChange,
		onFocus:     p.AlpinEventHandlers.OnFocus,
		onBlur:      p.AlpinEventHandlers.OnBlur,
		validator:   validator,
	}, config)
}

// ============================================================================
// TEXTAREA CONVERTER
// ============================================================================

func convertTextareaToSchema(props any) (schemaui.Component, error) {
	p, ok := props.(atoms.TextareaProps)
	if !ok {
		return schemaui.Component{}, fmt.Errorf("invalid props type for Textarea")
	}

	config := schemaui.TextareaConfig{
		Value:       p.Value,
		Placeholder: p.PlaceholderProps.Placeholder,
		Rows:        p.Rows,
		Cols:        p.Cols,
		MaxLength:   p.MaxLength,
		Resize:      fmt.Sprintf("%v", p.Resizable),
	}

	validator := buildValidator(p.InteractionProps.Required, 0, p.MaxLength, "")

	return buildComponent(schemaui.ComponentTextarea, baseBuilder{
		id:          p.BaseProps.ID,
		name:        p.BaseProps.Name,
		label:       "", // Textarea doesn't have direct label field
		placeholder: p.PlaceholderProps.Placeholder,
		class:       p.BaseProps.Class,
		size:        reverseConvertSize[atoms.TextareaSize](p.ComponentSize),
		disabled:    p.InteractionProps.Disabled,
		required:    p.InteractionProps.Required,
		readOnly:    p.InteractionProps.ReadOnly,
		ariaLabel:   p.AccessibilityProps.AriaLabel,
		onChange:    p.AlpinEventHandlers.OnChange,
		onFocus:     p.AlpinEventHandlers.OnFocus,
		onBlur:      p.AlpinEventHandlers.OnBlur,
		validator:   validator,
	}, config)
}

// ============================================================================
// SELECT CONVERTER
// ============================================================================

func convertSelectToSchema(props any) (schemaui.Component, error) {
	p, ok := props.(atoms.SelectProps)
	if !ok {
		return schemaui.Component{}, fmt.Errorf("invalid props type for Select")
	}

	options := make([]schemaui.Option, len(p.Options))
	for i, opt := range p.Options {
		options[i] = schemaui.Option{
			Value:    opt.Value,
			Label:    opt.Label,
			Disabled: opt.Disabled,
		}
	}

	config := schemaui.SelectConfig{
		Options:     options,
		Value:       p.Value,
		Multiple:    p.Multiple,
		Placeholder: p.PlaceholderProps.Placeholder,
	}

	return buildComponent(schemaui.ComponentSelect, baseBuilder{
		id:          p.BaseProps.ID,
		name:        p.BaseProps.Name,
		label:       "", // Select doesn't have direct label field
		placeholder: p.PlaceholderProps.Placeholder,
		class:       p.BaseProps.Class,
		size:        reverseConvertSize[atoms.SelectSize](p.Size),
		disabled:    p.InteractionProps.Disabled,
		required:    p.InteractionProps.Required,
		ariaLabel:   p.AccessibilityProps.AriaLabel,
		onChange:    p.AlpinEventHandlers.OnChange,
	}, config)
}

// ============================================================================
// CHECKBOX CONVERTER
// ============================================================================

func convertCheckboxToSchema(props any) (schemaui.Component, error) {
	p, ok := props.(atoms.CheckboxProps)
	if !ok {
		return schemaui.Component{}, fmt.Errorf("invalid props type for Checkbox")
	}

	config := schemaui.CheckboxConfig{
		Value:   p.Value == "true",
		Label:   getLabelString(p.LabelProps.Label),
		Checked: p.Checked,
	}

	return buildComponent(schemaui.ComponentCheckbox, baseBuilder{
		id:        p.BaseProps.ID,
		name:      p.BaseProps.Name,
		label:     getLabelString(p.LabelProps.Label),
		class:     p.BaseProps.Class,
		size:      reverseConvertSize[atoms.CheckboxSize](p.Size),
		disabled:  p.InteractionProps.Disabled,
		required:  p.InteractionProps.Required,
		ariaLabel: p.AccessibilityProps.AriaLabel,
		onChange:  p.AlpinEventHandlers.OnChange,
	}, config)
}

// ============================================================================
// RADIO CONVERTER
// ============================================================================

func convertRadioToSchema(props any) (schemaui.Component, error) {
	p, ok := props.(atoms.RadioGroupProps)
	if !ok {
		return schemaui.Component{}, fmt.Errorf("invalid props type for RadioGroup")
	}

	options := make([]schemaui.Option, len(p.Options))
	for i, opt := range p.Options {
		options[i] = schemaui.Option{
			Value:    opt.Value,
			Label:    opt.Label,
			Disabled: opt.Disabled,
		}
	}

	direction := p.Layout
	if direction == "" {
		direction = "vertical"
	}

	config := schemaui.RadioConfig{
		Options:   options,
		Value:     p.Value,
		Direction: direction,
	}

	return buildComponent(schemaui.ComponentRadio, baseBuilder{
		name:     p.BaseProps.Name,
		label:    p.Label,
		class:    p.BaseProps.Class,
		size:     reverseConvertSize[atoms.RadioSize](atoms.Size("md")), // Default size since RadioGroupProps doesn't have Size
		disabled: p.InteractionProps.Disabled,
		required: p.InteractionProps.Required,
		onChange: "", // RadioGroupProps doesn't have AlpinEventHandlers
	}, config)
}

// ============================================================================
// COMPONENT BUILDER
// ============================================================================

// baseBuilder holds common component properties for building schema components.
type baseBuilder struct {
	id          string
	name        string
	label       string
	placeholder string
	class       string
	size        schemaui.Size
	variant     schemaui.Variant
	disabled    bool
	required    bool
	readOnly    bool
	ariaLabel   string
	onClick     string
	onChange    string
	onFocus     string
	onBlur      string
	validator   *schemaui.Validator
}

// buildComponent constructs a schema component from base properties and config.
func buildComponent(typ schemaui.ComponentType, base baseBuilder, config any) (schemaui.Component, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return schemaui.Component{}, fmt.Errorf("marshal config: %w", err)
	}

	id := base.id
	if id == "" {
		id = generateID(string(typ))
	}

	now := time.Now()
	return schemaui.Component{
		BaseComponent: schemaui.BaseComponent{
			ID:          id,
			Type:        typ,
			Name:        base.name,
			Label:       base.label,
			Placeholder: base.placeholder,
			Class:       base.class,
			Size:        base.size,
			Variant:     base.variant,
			Disabled:    base.disabled,
			Required:    base.required,
			ReadOnly:    base.readOnly,
			AriaLabel:   base.ariaLabel,
			OnClick:     base.onClick,
			OnChange:    base.onChange,
			OnFocus:     base.onFocus,
			OnBlur:      base.onBlur,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		},
		Config:    configBytes,
		Children:  []schemaui.Component{},
		Validator: base.validator,
	}, nil
}

// buildValidator creates a validator if any validation rules are present.
func buildValidator(required bool, minLength, maxLength int, pattern string) *schemaui.Validator {
	if !required && minLength == 0 && maxLength == 0 && pattern == "" {
		return nil
	}

	v := &schemaui.Validator{
		Required: required,
		Pattern:  pattern,
	}

	if minLength > 0 {
		v.MinLength = &minLength
	}
	if maxLength > 0 {
		v.MaxLength = &maxLength
	}

	return v
}

// ============================================================================
// REVERSE SIZE CONVERSION
// ============================================================================

// reverseConvertSize converts component-specific size types to schema size.
// Since all size types are aliases for atoms.Size, we can convert them directly
func reverseConvertSize[T any](size T) schemaui.Size {
	// Convert to atoms.Size since all size types are aliases
	atomsSize := atoms.Size(fmt.Sprintf("%v", size))
	return reverseAtomSize(atomsSize)
}

func reverseAtomSize(size atoms.Size) schemaui.Size {
	switch size {
	case atoms.SizeXS, atoms.SizeSM:
		return schemaui.SizeSM
	case atoms.SizeLG:
		return schemaui.SizeLG
	case atoms.SizeXL:
		return schemaui.SizeXL
	default:
		return schemaui.SizeMD
	}
}


// getLabelString safely converts interface{} label to string
func getLabelString(label interface{}) string {
	if label == nil {
		return ""
	}
	if str, ok := label.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", label)
}


// ============================================================================
// REVERSE VARIANT CONVERSION
// ============================================================================

func reverseConvertVariant(variant atoms.ButtonVariant) schemaui.Variant {
	switch variant {
	case atoms.VariantPrimary:
		return schemaui.VariantPrimary
	case atoms.VariantSecondary:
		return schemaui.VariantSecondary
	case atoms.VariantWarning:
		return schemaui.VariantWarning
	case atoms.VariantLight:
		return schemaui.VariantLight
	case atoms.VariantDark:
		return schemaui.VariantDark
	default:
		return schemaui.VariantPrimary
	}
}

// ============================================================================
// STRING TO ENUM CONVERSIONS
// ============================================================================

func stringToPosition(pos string) schemaui.Position {
	switch pos {
	case "left":
		return schemaui.PositionLeft
	case "right":
		return schemaui.PositionRight
	case "top":
		return schemaui.PositionTop
	case "bottom":
		return schemaui.PositionBottom
	case "center":
		return schemaui.PositionCenter
	default:
		return schemaui.PositionLeft
	}
}

func stringToButtonType(typ string) schemaui.ButtonType {
	switch typ {
	case "submit":
		return schemaui.ButtonSubmit
	case "reset":
		return schemaui.ButtonReset
	case "button":
		return schemaui.ButtonButton
	default:
		return schemaui.ButtonButton
	}
}

func stringToInputType(typ string) schemaui.InputType {
	switch typ {
	case "text":
		return schemaui.InputText
	case "email":
		return schemaui.InputEmail
	case "password":
		return schemaui.InputPassword
	case "number":
		return schemaui.InputNumber
	case "tel":
		return schemaui.InputTel
	case "url":
		return schemaui.InputURL
	case "search":
		return schemaui.InputSearch
	case "date":
		return schemaui.InputDate
	case "time":
		return schemaui.InputTime
	case "datetime-local":
		return schemaui.InputDateTimeLocal
	case "file":
		return schemaui.InputFile
	case "hidden":
		return schemaui.InputHidden
	default:
		return schemaui.InputText
	}
}

// ============================================================================
// UTILITIES
// ============================================================================

// generateID creates a unique component identifier.
func generateID(typ string) string {
	id := uuid.New().String()[:8]
	id = strings.ReplaceAll(id, "-", "")
	return fmt.Sprintf("%s-%s", typ, id)
}

// GenerateSchemaFromTemplFile analyzes a Templ file and generates schema definitions.
// Note: This requires AST parsing of .templ files and is a placeholder for future implementation.
func (b *Bridge) GenerateSchemaFromTemplFile(ctx context.Context, templFilePath string) ([]schemaui.Component, error) {
	return nil, fmt.Errorf("Templ file parsing not implemented: use ConvertTemplToSchema for individual components")
}
