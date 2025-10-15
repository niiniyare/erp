package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

// ============================================================================
// BRIDGE CONVERTER SYSTEM
// ============================================================================

// Bridge connects the schema system (@pkg/schema/ui/) with Templ components (@web/components/)
type Bridge struct {
	schemaRegistry schemaui.ComponentRegistry
	cssFactory     *css.Factory
}

// NewBridge creates a new bridge instance with both system integrations
func NewBridge() *Bridge {
	return &Bridge{
		schemaRegistry: schemaui.NewRegistry(),
		cssFactory:     css.NewFactory("docs/ui/Schema"),
	}
}

// ============================================================================
// SCHEMA TO TEMPL CONVERSION
// ============================================================================

// ConvertSchemaToTempl converts a schema component to equivalent Templ component props
func (b *Bridge) ConvertSchemaToTempl(ctx context.Context, schemaComponent schemaui.Component) (TemplComponent, error) {
	switch schemaComponent.Type {
	case schemaui.ComponentButton:
		return b.convertButtonSchemaToTempl(schemaComponent)
	case schemaui.ComponentInput:
		return b.convertInputSchemaToTempl(schemaComponent)
	case schemaui.ComponentTextarea:
		return b.convertTextareaSchemaToTempl(schemaComponent)
	case schemaui.ComponentSelect:
		return b.convertSelectSchemaToTempl(schemaComponent)
	case schemaui.ComponentCheckbox:
		return b.convertCheckboxSchemaToTempl(schemaComponent)
	case schemaui.ComponentRadio:
		return b.convertRadioSchemaToTempl(schemaComponent)
	default:
		return TemplComponent{}, fmt.Errorf("unsupported schema component type: %s", schemaComponent.Type)
	}
}

// ============================================================================
// BUTTON COMPONENT BRIDGE
// ============================================================================

// convertButtonSchemaToTempl converts schema button to Templ ButtonProps
func (b *Bridge) convertButtonSchemaToTempl(schemaComponent schemaui.Component) (TemplComponent, error) {
	// Parse schema button config
	var buttonConfig schemaui.ButtonConfig
	if err := json.Unmarshal(schemaComponent.Config, &buttonConfig); err != nil {
		return TemplComponent{}, fmt.Errorf("failed to parse button config: %w", err)
	}

	// Convert to Templ ButtonProps
	templProps := atoms.ButtonProps{
		Text:        buttonConfig.Text,
		Icon:        buttonConfig.Icon,
		Type:        b.convertButtonType(buttonConfig.ButtonType),
		Disabled:    schemaComponent.Disabled,
		ID:          schemaComponent.ID,
		Class:       schemaComponent.Class,
		AriaLabel:   schemaComponent.AriaLabel,
		OnClick:     schemaComponent.OnClick,
	}

	// Convert variants using bridge mapping
	templProps.Variant = b.convertVariantToButtonVariant(schemaComponent.Variant)
	templProps.Size = b.convertSizeToButtonSize(schemaComponent.Size)

	// Handle icon position
	if buttonConfig.IconPosition != "" {
		templProps.IconPosition = string(buttonConfig.IconPosition)
	}

	// Handle loading state
	templProps.Loading = buttonConfig.Loading

	// Generate CSS if schema has styles
	if schemaComponent.Styles != nil {
		cssClasses, err := b.cssFactory.GenerateCSS(*schemaComponent.Styles)
		if err == nil && cssClasses != "" {
			if templProps.Class != "" {
				templProps.Class += " " + cssClasses
			} else {
				templProps.Class = cssClasses
			}
		}
	}

	return TemplComponent{
		Type:  "Button",
		Props: templProps,
	}, nil
}

// convertVariantToButtonVariant maps schema Variant to Templ ButtonVariant
func (b *Bridge) convertVariantToButtonVariant(variant schemaui.Variant) atoms.ButtonVariant {
	switch variant {
	case schemaui.VariantPrimary:
		return atoms.ButtonPrimary
	case schemaui.VariantSecondary:
		return atoms.ButtonSecondary
	case schemaui.VariantSuccess:
		return atoms.ButtonSuccess
	case schemaui.VariantDanger:
		return atoms.ButtonDanger
	case schemaui.VariantWarning:
		return atoms.ButtonWarning
	case schemaui.VariantInfo:
		return atoms.ButtonInfo
	case schemaui.VariantLight:
		return atoms.ButtonLight
	case schemaui.VariantDark:
		return atoms.ButtonDark
	default:
		return atoms.ButtonPrimary
	}
}

// convertSizeToButtonSize maps schema Size to Templ ButtonSize
func (b *Bridge) convertSizeToButtonSize(size schemaui.Size) atoms.ButtonSize {
	switch size {
	case schemaui.SizeXS:
		return atoms.ButtonSizeXS
	case schemaui.SizeSM:
		return atoms.ButtonSizeSM
	case schemaui.SizeLG:
		return atoms.ButtonSizeLG
	case schemaui.SizeXL:
		return atoms.ButtonSizeXL
	default:
		return atoms.ButtonSizeMD
	}
}

// convertButtonType maps schema ButtonType to string
func (b *Bridge) convertButtonType(buttonType schemaui.ButtonType) string {
	switch buttonType {
	case schemaui.ButtonSubmit:
		return "submit"
	case schemaui.ButtonReset:
		return "reset"
	case schemaui.ButtonButton:
		return "button"
	default:
		return "button"
	}
}

// ============================================================================
// INPUT COMPONENT BRIDGE
// ============================================================================

// convertInputSchemaToTempl converts schema input to Templ InputProps
func (b *Bridge) convertInputSchemaToTempl(schemaComponent schemaui.Component) (TemplComponent, error) {
	// Parse schema input config
	var inputConfig schemaui.InputConfig
	if err := json.Unmarshal(schemaComponent.Config, &inputConfig); err != nil {
		return TemplComponent{}, fmt.Errorf("failed to parse input config: %w", err)
	}

	// Convert to Templ InputProps
	templProps := atoms.InputProps{
		Type:        b.convertInputType(inputConfig.InputType),
		Value:       inputConfig.Value,
		Placeholder: inputConfig.Placeholder,
		Required:    schemaComponent.Required,
		Disabled:    schemaComponent.Disabled,
		ReadOnly:    schemaComponent.ReadOnly,
		ID:          schemaComponent.ID,
		Name:        schemaComponent.Name,
		Class:       schemaComponent.Class,
		AriaLabel:   schemaComponent.AriaLabel,
	}

	// Convert size
	templProps.Size = b.convertSizeToInputSize(schemaComponent.Size)

	// Handle validation
	if schemaComponent.Validator != nil {
		templProps.Required = schemaComponent.Validator.Required
		if schemaComponent.Validator.MaxLength != nil {
			templProps.MaxLength = *schemaComponent.Validator.MaxLength
		}
		if schemaComponent.Validator.MinLength != nil {
			templProps.MinLength = *schemaComponent.Validator.MinLength
		}
		if schemaComponent.Validator.Pattern != "" {
			templProps.Pattern = schemaComponent.Validator.Pattern
		}
	}

	// Generate CSS if schema has styles
	if schemaComponent.Styles != nil {
		cssClasses, err := b.cssFactory.GenerateCSS(*schemaComponent.Styles)
		if err == nil && cssClasses != "" {
			if templProps.Class != "" {
				templProps.Class += " " + cssClasses
			} else {
				templProps.Class = cssClasses
			}
		}
	}

	return TemplComponent{
		Type:  "Input",
		Props: templProps,
	}, nil
}

// convertInputType maps schema InputType to string
func (b *Bridge) convertInputType(inputType schemaui.InputType) string {
	return string(inputType)
}

// convertSizeToInputSize maps schema Size to Templ InputSize
func (b *Bridge) convertSizeToInputSize(size schemaui.Size) atoms.InputSize {
	switch size {
	case schemaui.SizeXS:
		return atoms.InputSizeXS
	case schemaui.SizeSM:
		return atoms.InputSizeSM
	case schemaui.SizeLG:
		return atoms.InputSizeLG
	case schemaui.SizeXL:
		return atoms.InputSizeXL
	default:
		return atoms.InputSizeMD
	}
}

// ============================================================================
// TEXTAREA COMPONENT BRIDGE
// ============================================================================

// convertTextareaSchemaToTempl converts schema textarea to Templ TextareaProps
func (b *Bridge) convertTextareaSchemaToTempl(schemaComponent schemaui.Component) (TemplComponent, error) {
	// Parse schema textarea config
	var textareaConfig schemaui.TextareaConfig
	if err := json.Unmarshal(schemaComponent.Config, &textareaConfig); err != nil {
		return TemplComponent{}, fmt.Errorf("failed to parse textarea config: %w", err)
	}

	// Convert to Templ TextareaProps
	templProps := atoms.TextareaProps{
		Value:       textareaConfig.Value,
		Placeholder: textareaConfig.Placeholder,
		Rows:        textareaConfig.Rows,
		Cols:        textareaConfig.Cols,
		Required:    schemaComponent.Required,
		Disabled:    schemaComponent.Disabled,
		ReadOnly:    schemaComponent.ReadOnly,
		ID:          schemaComponent.ID,
		Name:        schemaComponent.Name,
		Class:       schemaComponent.Class,
		AriaLabel:   schemaComponent.AriaLabel,
	}

	// Convert size
	templProps.Size = b.convertSizeToTextareaSize(schemaComponent.Size)

	// Handle resize setting
	if textareaConfig.Resize != "" {
		templProps.Resize = textareaConfig.Resize
	}

	// Handle max length
	if textareaConfig.MaxLength > 0 {
		templProps.MaxLength = textareaConfig.MaxLength
	}

	// Generate CSS if schema has styles
	if schemaComponent.Styles != nil {
		cssClasses, err := b.cssFactory.GenerateCSS(*schemaComponent.Styles)
		if err == nil && cssClasses != "" {
			if templProps.Class != "" {
				templProps.Class += " " + cssClasses
			} else {
				templProps.Class = cssClasses
			}
		}
	}

	return TemplComponent{
		Type:  "Textarea",
		Props: templProps,
	}, nil
}

// convertSizeToTextareaSize maps schema Size to Templ TextareaSize
func (b *Bridge) convertSizeToTextareaSize(size schemaui.Size) atoms.TextareaSize {
	switch size {
	case schemaui.SizeXS:
		return atoms.TextareaSizeXS
	case schemaui.SizeSM:
		return atoms.TextareaSizeSM
	case schemaui.SizeLG:
		return atoms.TextareaSizeLG
	case schemaui.SizeXL:
		return atoms.TextareaSizeXL
	default:
		return atoms.TextareaSizeMD
	}
}

// ============================================================================
// SELECT COMPONENT BRIDGE
// ============================================================================

// convertSelectSchemaToTempl converts schema select to Templ SelectProps
func (b *Bridge) convertSelectSchemaToTempl(schemaComponent schemaui.Component) (TemplComponent, error) {
	// Parse schema select config
	var selectConfig schemaui.SelectConfig
	if err := json.Unmarshal(schemaComponent.Config, &selectConfig); err != nil {
		return TemplComponent{}, fmt.Errorf("failed to parse select config: %w", err)
	}

	// Convert options
	var templOptions []atoms.SelectOption
	for _, option := range selectConfig.Options {
		templOptions = append(templOptions, atoms.SelectOption{
			Value:    option.Value,
			Label:    option.Label,
			Disabled: option.Disabled,
			Group:    option.Group,
		})
	}

	// Convert to Templ SelectProps
	templProps := atoms.SelectProps{
		Options:     templOptions,
		Value:       selectConfig.Value,
		Multiple:    selectConfig.Multiple,
		Placeholder: selectConfig.Placeholder,
		Required:    schemaComponent.Required,
		Disabled:    schemaComponent.Disabled,
		ID:          schemaComponent.ID,
		Name:        schemaComponent.Name,
		Class:       schemaComponent.Class,
		AriaLabel:   schemaComponent.AriaLabel,
	}

	// Convert size
	templProps.Size = b.convertSizeToSelectSize(schemaComponent.Size)

	// Generate CSS if schema has styles
	if schemaComponent.Styles != nil {
		cssClasses, err := b.cssFactory.GenerateCSS(*schemaComponent.Styles)
		if err == nil && cssClasses != "" {
			if templProps.Class != "" {
				templProps.Class += " " + cssClasses
			} else {
				templProps.Class = cssClasses
			}
		}
	}

	return TemplComponent{
		Type:  "Select",
		Props: templProps,
	}, nil
}

// convertSizeToSelectSize maps schema Size to Templ SelectSize
func (b *Bridge) convertSizeToSelectSize(size schemaui.Size) atoms.SelectSize {
	switch size {
	case schemaui.SizeXS:
		return atoms.SelectSizeXS
	case schemaui.SizeSM:
		return atoms.SelectSizeSM
	case schemaui.SizeLG:
		return atoms.SelectSizeLG
	case schemaui.SizeXL:
		return atoms.SelectSizeXL
	default:
		return atoms.SelectSizeMD
	}
}

// ============================================================================
// CHECKBOX COMPONENT BRIDGE
// ============================================================================

// convertCheckboxSchemaToTempl converts schema checkbox to Templ CheckboxProps
func (b *Bridge) convertCheckboxSchemaToTempl(schemaComponent schemaui.Component) (TemplComponent, error) {
	// Parse schema checkbox config
	var checkboxConfig schemaui.CheckboxConfig
	if err := json.Unmarshal(schemaComponent.Config, &checkboxConfig); err != nil {
		return TemplComponent{}, fmt.Errorf("failed to parse checkbox config: %w", err)
	}

	// Convert to Templ CheckboxProps
	templProps := atoms.CheckboxProps{
		Checked:   checkboxConfig.Checked,
		Value:     checkboxConfig.Value,
		Required:  schemaComponent.Required,
		Disabled:  schemaComponent.Disabled,
		ID:        schemaComponent.ID,
		Name:      schemaComponent.Name,
		Class:     schemaComponent.Class,
		AriaLabel: schemaComponent.AriaLabel,
	}

	// Use label from config or base component
	if checkboxConfig.Label != "" {
		templProps.Label = checkboxConfig.Label
	} else if schemaComponent.Label != "" {
		templProps.Label = schemaComponent.Label
	}

	// Convert size
	templProps.Size = b.convertSizeToCheckboxSize(schemaComponent.Size)

	// Generate CSS if schema has styles
	if schemaComponent.Styles != nil {
		cssClasses, err := b.cssFactory.GenerateCSS(*schemaComponent.Styles)
		if err == nil && cssClasses != "" {
			if templProps.Class != "" {
				templProps.Class += " " + cssClasses
			} else {
				templProps.Class = cssClasses
			}
		}
	}

	return TemplComponent{
		Type:  "Checkbox",
		Props: templProps,
	}, nil
}

// convertSizeToCheckboxSize maps schema Size to Templ CheckboxSize
func (b *Bridge) convertSizeToCheckboxSize(size schemaui.Size) atoms.CheckboxSize {
	switch size {
	case schemaui.SizeXS:
		return atoms.CheckboxSizeXS
	case schemaui.SizeSM:
		return atoms.CheckboxSizeSM
	case schemaui.SizeLG:
		return atoms.CheckboxSizeLG
	case schemaui.SizeXL:
		return atoms.CheckboxSizeXL
	default:
		return atoms.CheckboxSizeMD
	}
}

// ============================================================================
// RADIO COMPONENT BRIDGE
// ============================================================================

// convertRadioSchemaToTempl converts schema radio to Templ RadioProps
func (b *Bridge) convertRadioSchemaToTempl(schemaComponent schemaui.Component) (TemplComponent, error) {
	// Parse schema radio config
	var radioConfig schemaui.RadioConfig
	if err := json.Unmarshal(schemaComponent.Config, &radioConfig); err != nil {
		return TemplComponent{}, fmt.Errorf("failed to parse radio config: %w", err)
	}

	// Convert options
	var templOptions []atoms.RadioOption
	for _, option := range radioConfig.Options {
		templOptions = append(templOptions, atoms.RadioOption{
			Value:    option.Value,
			Label:    option.Label,
			Disabled: option.Disabled,
		})
	}

	// Convert to Templ RadioProps
	templProps := atoms.RadioProps{
		Options:   templOptions,
		Value:     radioConfig.Value,
		Required:  schemaComponent.Required,
		Disabled:  schemaComponent.Disabled,
		ID:        schemaComponent.ID,
		Name:      schemaComponent.Name,
		Class:     schemaComponent.Class,
		AriaLabel: schemaComponent.AriaLabel,
	}

	// Convert direction
	if radioConfig.Direction == "horizontal" {
		templProps.Inline = true
	}

	// Convert size
	templProps.Size = b.convertSizeToRadioSize(schemaComponent.Size)

	// Generate CSS if schema has styles
	if schemaComponent.Styles != nil {
		cssClasses, err := b.cssFactory.GenerateCSS(*schemaComponent.Styles)
		if err == nil && cssClasses != "" {
			if templProps.Class != "" {
				templProps.Class += " " + cssClasses
			} else {
				templProps.Class = cssClasses
			}
		}
	}

	return TemplComponent{
		Type:  "Radio",
		Props: templProps,
	}, nil
}

// convertSizeToRadioSize maps schema Size to Templ RadioSize
func (b *Bridge) convertSizeToRadioSize(size schemaui.Size) atoms.RadioSize {
	switch size {
	case schemaui.SizeXS:
		return atoms.RadioSizeXS
	case schemaui.SizeSM:
		return atoms.RadioSizeSM
	case schemaui.SizeLG:
		return atoms.RadioSizeLG
	case schemaui.SizeXL:
		return atoms.RadioSizeXL
	default:
		return atoms.RadioSizeMD
	}
}

// ============================================================================
// BRIDGE DATA TYPES
// ============================================================================

// TemplComponent represents a converted Templ component
type TemplComponent struct {
	Type  string      `json:"type"`  // Component type name (Button, Input, etc.)
	Props interface{} `json:"props"` // Component-specific props
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// ConvertSchemaComponentsToTempl converts multiple schema components to Templ components
func (b *Bridge) ConvertSchemaComponentsToTempl(ctx context.Context, schemaComponents []schemaui.Component) ([]TemplComponent, error) {
	var templComponents []TemplComponent
	
	for _, schemaComponent := range schemaComponents {
		templComponent, err := b.ConvertSchemaToTempl(ctx, schemaComponent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert component %s: %w", schemaComponent.ID, err)
		}
		templComponents = append(templComponents, templComponent)
	}
	
	return templComponents, nil
}

// GenerateTemplCodeFromSchema generates Go Templ code from schema components
func (b *Bridge) GenerateTemplCodeFromSchema(ctx context.Context, schemaComponents []schemaui.Component, templateName string) (string, error) {
	templComponents, err := b.ConvertSchemaComponentsToTempl(ctx, schemaComponents)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	
	// Generate Templ template header
	builder.WriteString(fmt.Sprintf("templ %s() {\n", templateName))
	
	// Generate component calls
	for _, templComponent := range templComponents {
		switch templComponent.Type {
		case "Button":
			props := templComponent.Props.(atoms.ButtonProps)
			builder.WriteString(fmt.Sprintf("\t@atoms.Button(%+v)\n", props))
		case "Input":
			props := templComponent.Props.(atoms.InputProps)
			builder.WriteString(fmt.Sprintf("\t@atoms.Input(%+v)\n", props))
		case "Textarea":
			props := templComponent.Props.(atoms.TextareaProps)
			builder.WriteString(fmt.Sprintf("\t@atoms.Textarea(%+v)\n", props))
		case "Select":
			props := templComponent.Props.(atoms.SelectProps)
			builder.WriteString(fmt.Sprintf("\t@atoms.Select(%+v)\n", props))
		case "Checkbox":
			props := templComponent.Props.(atoms.CheckboxProps)
			builder.WriteString(fmt.Sprintf("\t@atoms.Checkbox(%+v)\n", props))
		case "Radio":
			props := templComponent.Props.(atoms.RadioProps)
			builder.WriteString(fmt.Sprintf("\t@atoms.Radio(%+v)\n", props))
		}
	}
	
	builder.WriteString("}\n")
	
	return builder.String(), nil
}