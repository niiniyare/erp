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
		Icon:         p.Icon,
		IconPosition: stringToPosition(p.IconPosition),
		ButtonType:   stringToButtonType(p.Type),
		Loading:      p.Loading,
	}

	return buildComponent(schemaui.ComponentButton, baseBuilder{
		id:        p.ID,
		label:     p.Text,
		class:     p.Class,
		size:      reverseConvertSize[atoms.ButtonSize](p.Size),
		variant:   reverseConvertVariant(p.Variant),
		disabled:  p.Disabled,
		ariaLabel: p.AriaLabel,
		onClick:   p.OnClick,
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
		Placeholder: p.Placeholder,
		MaxLength:   p.MaxLength,
		MinLength:   p.MinLength,
		Pattern:     p.Pattern,
	}

	validator := buildValidator(p.Required, p.MinLength, p.MaxLength, p.Pattern)

	return buildComponent(schemaui.ComponentInput, baseBuilder{
		id:          p.ID,
		name:        p.Name,
		label:       p.Label,
		placeholder: p.Placeholder,
		class:       p.Class,
		size:        reverseConvertSize[atoms.InputSize](p.Size),
		disabled:    p.Disabled,
		required:    p.Required,
		readOnly:    p.ReadOnly,
		ariaLabel:   p.AriaLabel,
		onChange:    p.OnChange,
		onFocus:     p.OnFocus,
		onBlur:      p.OnBlur,
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
		Placeholder: p.Placeholder,
		Rows:        p.Rows,
		Cols:        p.Cols,
		MaxLength:   p.MaxLength,
		Resize:      fmt.Sprintf("%v", p.Resizable),
	}

	validator := buildValidator(p.Required, 0, p.MaxLength, "")

	return buildComponent(schemaui.ComponentTextarea, baseBuilder{
		id:          p.ID,
		name:        p.Name,
		label:       p.Label,
		placeholder: p.Placeholder,
		class:       p.Class,
		size:        reverseConvertSize[atoms.TextareaSize](p.Size),
		disabled:    p.Disabled,
		required:    p.Required,
		readOnly:    p.ReadOnly,
		ariaLabel:   p.AriaLabel,
		onChange:    p.OnChange,
		onFocus:     p.OnFocus,
		onBlur:      p.OnBlur,
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
		Placeholder: p.Placeholder,
	}

	return buildComponent(schemaui.ComponentSelect, baseBuilder{
		id:          p.ID,
		name:        p.Name,
		label:       p.Label,
		placeholder: p.Placeholder,
		class:       p.Class,
		size:        reverseConvertSize[atoms.SelectSize](p.SelectSize),
		disabled:    p.Disabled,
		required:    p.Required,
		ariaLabel:   p.AriaLabel,
		onChange:    p.OnChange,
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
		Label:   p.Label,
		Checked: p.Checked,
	}

	return buildComponent(schemaui.ComponentCheckbox, baseBuilder{
		id:        p.ID,
		name:      p.Name,
		label:     p.Label,
		class:     p.Class,
		size:      reverseConvertSize[atoms.CheckboxSize](p.Size),
		disabled:  p.Disabled,
		required:  p.Required,
		ariaLabel: p.AriaLabel,
		onChange:  p.OnChange,
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
		name:     p.Name,
		label:    p.Label,
		class:    p.Class,
		size:     reverseConvertSize[atoms.RadioSize](p.Size),
		disabled: p.Disabled,
		required: p.Required,
		onChange: p.OnChange,
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
func reverseConvertSize[T any](size T) schemaui.Size {
	switch any(size).(type) {
	case atoms.ButtonSize:
		return reverseButtonSize(any(size).(atoms.ButtonSize))
	case atoms.InputSize:
		return reverseInputSize(any(size).(atoms.InputSize))
	case atoms.TextareaSize:
		return reverseTextareaSize(any(size).(atoms.TextareaSize))
	case atoms.SelectSize:
		return reverseSelectSize(any(size).(atoms.SelectSize))
	case atoms.CheckboxSize:
		return reverseCheckboxSize(any(size).(atoms.CheckboxSize))
	case atoms.RadioSize:
		return reverseRadioSize(any(size).(atoms.RadioSize))
	default:
		return schemaui.SizeMD
	}
}

func reverseButtonSize(size atoms.ButtonSize) schemaui.Size {
	switch size {
	case atoms.ButtonSizeSM:
		return schemaui.SizeSM
	case atoms.ButtonSizeLG:
		return schemaui.SizeLG
	default:
		return schemaui.SizeMD
	}
}

func reverseInputSize(size atoms.InputSize) schemaui.Size {
	switch size {
	case atoms.InputSizeSM:
		return schemaui.SizeSM
	case atoms.InputSizeLG:
		return schemaui.SizeLG
	default:
		return schemaui.SizeMD
	}
}

func reverseTextareaSize(size atoms.TextareaSize) schemaui.Size {
	switch size {
	case atoms.TextareaSizeSM:
		return schemaui.SizeSM
	case atoms.TextareaSizeLG:
		return schemaui.SizeLG
	default:
		return schemaui.SizeMD
	}
}

func reverseSelectSize(size atoms.SelectSize) schemaui.Size {
	switch size {
	case atoms.SelectSizeSM:
		return schemaui.SizeSM
	case atoms.SelectSizeLG:
		return schemaui.SizeLG
	default:
		return schemaui.SizeMD
	}
}

func reverseCheckboxSize(size atoms.CheckboxSize) schemaui.Size {
	switch size {
	case atoms.CheckboxSizeSM:
		return schemaui.SizeSM
	case atoms.CheckboxSizeLG:
		return schemaui.SizeLG
	default:
		return schemaui.SizeMD
	}
}

func reverseRadioSize(size atoms.RadioSize) schemaui.Size {
	switch size {
	case atoms.RadioSizeSM:
		return schemaui.SizeSM
	case atoms.RadioSizeLG:
		return schemaui.SizeLG
	default:
		return schemaui.SizeMD
	}
}

// ============================================================================
// REVERSE VARIANT CONVERSION
// ============================================================================

func reverseConvertVariant(variant atoms.ButtonVariant) schemaui.Variant {
	switch variant {
	case atoms.ButtonPrimary:
		return schemaui.VariantPrimary
	case atoms.ButtonSecondary:
		return schemaui.VariantSecondary
	case atoms.ButtonSuccess:
		return schemaui.VariantSuccess
	case atoms.ButtonDanger:
		return schemaui.VariantDanger
	case atoms.ButtonWarning:
		return schemaui.VariantWarning
	case atoms.ButtonInfo:
		return schemaui.VariantInfo
	case atoms.ButtonLight:
		return schemaui.VariantLight
	case atoms.ButtonDark:
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
