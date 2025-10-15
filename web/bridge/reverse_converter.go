package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	schemaui "github.com/niiniyare/erp/pkg/schema/ui"
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/google/uuid"
)

// ============================================================================
// TEMPL TO SCHEMA CONVERSION
// ============================================================================

// ConvertTemplToSchema converts Templ component props to equivalent schema component
func (b *Bridge) ConvertTemplToSchema(ctx context.Context, templComponent TemplComponent) (schemaui.Component, error) {
	switch templComponent.Type {
	case "Button":
		return b.convertTemplButtonToSchema(templComponent.Props.(atoms.ButtonProps))
	case "Input":
		return b.convertTemplInputToSchema(templComponent.Props.(atoms.InputProps))
	case "Textarea":
		return b.convertTemplTextareaToSchema(templComponent.Props.(atoms.TextareaProps))
	case "Select":
		return b.convertTemplSelectToSchema(templComponent.Props.(atoms.SelectProps))
	case "Checkbox":
		return b.convertTemplCheckboxToSchema(templComponent.Props.(atoms.CheckboxProps))
	case "Radio":
		return b.convertTemplRadioToSchema(templComponent.Props.(atoms.RadioProps))
	default:
		return schemaui.Component{}, fmt.Errorf("unsupported Templ component type: %s", templComponent.Type)
	}
}

// ============================================================================
// BUTTON TEMPL TO SCHEMA CONVERSION
// ============================================================================

// convertTemplButtonToSchema converts Templ ButtonProps to schema Component
func (b *Bridge) convertTemplButtonToSchema(props atoms.ButtonProps) (schemaui.Component, error) {
	// Create schema button config
	buttonConfig := schemaui.ButtonConfig{
		Text:         props.Text,
		Icon:         props.Icon,
		IconPosition: b.convertStringToPosition(props.IconPosition),
		ButtonType:   b.convertStringToButtonType(props.Type),
		Loading:      props.Loading,
	}

	// Marshal config to JSON
	configBytes, err := json.Marshal(buttonConfig)
	if err != nil {
		return schemaui.Component{}, fmt.Errorf("failed to marshal button config: %w", err)
	}

	// Create base component
	baseComponent := schemaui.BaseComponent{
		ID:          props.ID,
		Type:        schemaui.ComponentButton,
		Name:        "",
		Label:       props.Text,
		Class:       props.Class,
		Size:        b.convertButtonSizeToSchemaSize(props.Size),
		Variant:     b.convertButtonVariantToSchemaVariant(props.Variant),
		Disabled:    props.Disabled,
		Required:    false,
		AriaLabel:   props.AriaLabel,
		OnClick:     props.OnClick,
		CreatedAt:   &time.Time{},
		UpdatedAt:   &time.Time{},
	}

	// If no ID provided, generate one
	if baseComponent.ID == "" {
		baseComponent.ID = generateComponentID("button")
	}

	// Create full component
	return schemaui.Component{
		BaseComponent: baseComponent,
		Config:        configBytes,
		Children:      []schemaui.Component{},
	}, nil
}

// ============================================================================
// INPUT TEMPL TO SCHEMA CONVERSION
// ============================================================================

// convertTemplInputToSchema converts Templ InputProps to schema Component
func (b *Bridge) convertTemplInputToSchema(props atoms.InputProps) (schemaui.Component, error) {
	// Create schema input config
	inputConfig := schemaui.InputConfig{
		InputType:   b.convertStringToInputType(props.Type),
		Value:       props.Value,
		Placeholder: props.Placeholder,
		MaxLength:   props.MaxLength,
		MinLength:   props.MinLength,
		Pattern:     props.Pattern,
	}

	// Marshal config to JSON
	configBytes, err := json.Marshal(inputConfig)
	if err != nil {
		return schemaui.Component{}, fmt.Errorf("failed to marshal input config: %w", err)
	}

	// Create validator from props
	var validator *schemaui.Validator
	if props.Required || props.MaxLength > 0 || props.MinLength > 0 || props.Pattern != "" {
		validator = &schemaui.Validator{
			Required: props.Required,
			Pattern:  props.Pattern,
		}
		if props.MaxLength > 0 {
			validator.MaxLength = &props.MaxLength
		}
		if props.MinLength > 0 {
			validator.MinLength = &props.MinLength
		}
	}

	// Create base component
	baseComponent := schemaui.BaseComponent{
		ID:          props.ID,
		Type:        schemaui.ComponentInput,
		Name:        props.Name,
		Label:       props.Label,
		Placeholder: props.Placeholder,
		Class:       props.Class,
		Size:        b.convertInputSizeToSchemaSize(props.Size),
		Disabled:    props.Disabled,
		Required:    props.Required,
		ReadOnly:    props.ReadOnly,
		AriaLabel:   props.AriaLabel,
		OnChange:    props.OnChange,
		OnFocus:     props.OnFocus,
		OnBlur:      props.OnBlur,
		CreatedAt:   &time.Time{},
		UpdatedAt:   &time.Time{},
	}

	// If no ID provided, generate one
	if baseComponent.ID == "" {
		baseComponent.ID = generateComponentID("input")
	}

	// Create full component
	return schemaui.Component{
		BaseComponent: baseComponent,
		Config:        configBytes,
		Children:      []schemaui.Component{},
		Validator:     validator,
	}, nil
}

// ============================================================================
// TEXTAREA TEMPL TO SCHEMA CONVERSION
// ============================================================================

// convertTemplTextareaToSchema converts Templ TextareaProps to schema Component
func (b *Bridge) convertTemplTextareaToSchema(props atoms.TextareaProps) (schemaui.Component, error) {
	// Create schema textarea config
	textareaConfig := schemaui.TextareaConfig{
		Value:       props.Value,
		Placeholder: props.Placeholder,
		Rows:        props.Rows,
		Cols:        props.Cols,
		MaxLength:   props.MaxLength,
		Resize:      props.Resize,
	}

	// Marshal config to JSON
	configBytes, err := json.Marshal(textareaConfig)
	if err != nil {
		return schemaui.Component{}, fmt.Errorf("failed to marshal textarea config: %w", err)
	}

	// Create validator from props
	var validator *schemaui.Validator
	if props.Required || props.MaxLength > 0 {
		validator = &schemaui.Validator{
			Required: props.Required,
		}
		if props.MaxLength > 0 {
			validator.MaxLength = &props.MaxLength
		}
	}

	// Create base component
	baseComponent := schemaui.BaseComponent{
		ID:          props.ID,
		Type:        schemaui.ComponentTextarea,
		Name:        props.Name,
		Label:       props.Label,
		Placeholder: props.Placeholder,
		Class:       props.Class,
		Size:        b.convertTextareaSizeToSchemaSize(props.Size),
		Disabled:    props.Disabled,
		Required:    props.Required,
		ReadOnly:    props.ReadOnly,
		AriaLabel:   props.AriaLabel,
		OnChange:    props.OnChange,
		OnFocus:     props.OnFocus,
		OnBlur:      props.OnBlur,
		CreatedAt:   &time.Time{},
		UpdatedAt:   &time.Time{},
	}

	// If no ID provided, generate one
	if baseComponent.ID == "" {
		baseComponent.ID = generateComponentID("textarea")
	}

	// Create full component
	return schemaui.Component{
		BaseComponent: baseComponent,
		Config:        configBytes,
		Children:      []schemaui.Component{},
		Validator:     validator,
	}, nil
}

// ============================================================================
// SELECT TEMPL TO SCHEMA CONVERSION
// ============================================================================

// convertTemplSelectToSchema converts Templ SelectProps to schema Component
func (b *Bridge) convertTemplSelectToSchema(props atoms.SelectProps) (schemaui.Component, error) {
	// Convert options
	var schemaOptions []schemaui.Option
	for _, option := range props.Options {
		schemaOptions = append(schemaOptions, schemaui.Option{
			Value:    option.Value,
			Label:    option.Label,
			Disabled: option.Disabled,
			Group:    option.Group,
		})
	}

	// Create schema select config
	selectConfig := schemaui.SelectConfig{
		Options:     schemaOptions,
		Value:       props.Value,
		Multiple:    props.Multiple,
		Placeholder: props.Placeholder,
	}

	// Marshal config to JSON
	configBytes, err := json.Marshal(selectConfig)
	if err != nil {
		return schemaui.Component{}, fmt.Errorf("failed to marshal select config: %w", err)
	}

	// Create base component
	baseComponent := schemaui.BaseComponent{
		ID:          props.ID,
		Type:        schemaui.ComponentSelect,
		Name:        props.Name,
		Label:       props.Label,
		Placeholder: props.Placeholder,
		Class:       props.Class,
		Size:        b.convertSelectSizeToSchemaSize(props.Size),
		Disabled:    props.Disabled,
		Required:    props.Required,
		AriaLabel:   props.AriaLabel,
		OnChange:    props.OnChange,
		CreatedAt:   &time.Time{},
		UpdatedAt:   &time.Time{},
	}

	// If no ID provided, generate one
	if baseComponent.ID == "" {
		baseComponent.ID = generateComponentID("select")
	}

	// Create full component
	return schemaui.Component{
		BaseComponent: baseComponent,
		Config:        configBytes,
		Children:      []schemaui.Component{},
	}, nil
}

// ============================================================================
// CHECKBOX TEMPL TO SCHEMA CONVERSION
// ============================================================================

// convertTemplCheckboxToSchema converts Templ CheckboxProps to schema Component
func (b *Bridge) convertTemplCheckboxToSchema(props atoms.CheckboxProps) (schemaui.Component, error) {
	// Create schema checkbox config
	checkboxConfig := schemaui.CheckboxConfig{
		Value:   props.Value,
		Label:   props.Label,
		Checked: props.Checked,
	}

	// Marshal config to JSON
	configBytes, err := json.Marshal(checkboxConfig)
	if err != nil {
		return schemaui.Component{}, fmt.Errorf("failed to marshal checkbox config: %w", err)
	}

	// Create base component
	baseComponent := schemaui.BaseComponent{
		ID:        props.ID,
		Type:      schemaui.ComponentCheckbox,
		Name:      props.Name,
		Label:     props.Label,
		Class:     props.Class,
		Size:      b.convertCheckboxSizeToSchemaSize(props.Size),
		Disabled:  props.Disabled,
		Required:  props.Required,
		AriaLabel: props.AriaLabel,
		OnChange:  props.OnChange,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}

	// If no ID provided, generate one
	if baseComponent.ID == "" {
		baseComponent.ID = generateComponentID("checkbox")
	}

	// Create full component
	return schemaui.Component{
		BaseComponent: baseComponent,
		Config:        configBytes,
		Children:      []schemaui.Component{},
	}, nil
}

// ============================================================================
// RADIO TEMPL TO SCHEMA CONVERSION
// ============================================================================

// convertTemplRadioToSchema converts Templ RadioProps to schema Component
func (b *Bridge) convertTemplRadioToSchema(props atoms.RadioProps) (schemaui.Component, error) {
	// Convert options
	var schemaOptions []schemaui.Option
	for _, option := range props.Options {
		schemaOptions = append(schemaOptions, schemaui.Option{
			Value:    option.Value,
			Label:    option.Label,
			Disabled: option.Disabled,
		})
	}

	// Create schema radio config
	radioConfig := schemaui.RadioConfig{
		Options: schemaOptions,
		Value:   props.Value,
	}

	// Set direction based on inline
	if props.Inline {
		radioConfig.Direction = "horizontal"
	} else {
		radioConfig.Direction = "vertical"
	}

	// Marshal config to JSON
	configBytes, err := json.Marshal(radioConfig)
	if err != nil {
		return schemaui.Component{}, fmt.Errorf("failed to marshal radio config: %w", err)
	}

	// Create base component
	baseComponent := schemaui.BaseComponent{
		ID:        props.ID,
		Type:      schemaui.ComponentRadio,
		Name:      props.Name,
		Label:     props.Label,
		Class:     props.Class,
		Size:      b.convertRadioSizeToSchemaSize(props.Size),
		Disabled:  props.Disabled,
		Required:  props.Required,
		AriaLabel: props.AriaLabel,
		OnChange:  props.OnChange,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}

	// If no ID provided, generate one
	if baseComponent.ID == "" {
		baseComponent.ID = generateComponentID("radio")
	}

	// Create full component
	return schemaui.Component{
		BaseComponent: baseComponent,
		Config:        configBytes,
		Children:      []schemaui.Component{},
	}, nil
}

// ============================================================================
// REVERSE MAPPING FUNCTIONS
// ============================================================================

// convertButtonVariantToSchemaVariant maps Templ ButtonVariant to schema Variant
func (b *Bridge) convertButtonVariantToSchemaVariant(variant atoms.ButtonVariant) schemaui.Variant {
	switch variant {
	case atoms.ButtonPrimary:
		return schemaui.VariantPrimary
	case atoms.ButtonSecondary:
		return schemaui.VariantSecondary
	case atoms.ButtonSuccess, atoms.ButtonGreen:
		return schemaui.VariantSuccess
	case atoms.ButtonDanger, atoms.ButtonRed:
		return schemaui.VariantDanger
	case atoms.ButtonWarning, atoms.ButtonYellow:
		return schemaui.VariantWarning
	case atoms.ButtonInfo, atoms.ButtonPurple:
		return schemaui.VariantInfo
	case atoms.ButtonLight:
		return schemaui.VariantLight
	case atoms.ButtonDark:
		return schemaui.VariantDark
	default:
		return schemaui.VariantPrimary
	}
}

// convertButtonSizeToSchemaSize maps Templ ButtonSize to schema Size
func (b *Bridge) convertButtonSizeToSchemaSize(size atoms.ButtonSize) schemaui.Size {
	switch size {
	case atoms.ButtonSizeXS:
		return schemaui.SizeXS
	case atoms.ButtonSizeSM:
		return schemaui.SizeSM
	case atoms.ButtonSizeLG:
		return schemaui.SizeLG
	case atoms.ButtonSizeXL:
		return schemaui.SizeXL
	default:
		return schemaui.SizeMD
	}
}

// convertInputSizeToSchemaSize maps Templ InputSize to schema Size
func (b *Bridge) convertInputSizeToSchemaSize(size atoms.InputSize) schemaui.Size {
	switch size {
	case atoms.InputSizeXS:
		return schemaui.SizeXS
	case atoms.InputSizeSM:
		return schemaui.SizeSM
	case atoms.InputSizeLG:
		return schemaui.SizeLG
	case atoms.InputSizeXL:
		return schemaui.SizeXL
	default:
		return schemaui.SizeMD
	}
}

// convertTextareaSizeToSchemaSize maps Templ TextareaSize to schema Size
func (b *Bridge) convertTextareaSizeToSchemaSize(size atoms.TextareaSize) schemaui.Size {
	switch size {
	case atoms.TextareaSizeXS:
		return schemaui.SizeXS
	case atoms.TextareaSizeSM:
		return schemaui.SizeSM
	case atoms.TextareaSizeLG:
		return schemaui.SizeLG
	case atoms.TextareaSizeXL:
		return schemaui.SizeXL
	default:
		return schemaui.SizeMD
	}
}

// convertSelectSizeToSchemaSize maps Templ SelectSize to schema Size
func (b *Bridge) convertSelectSizeToSchemaSize(size atoms.SelectSize) schemaui.Size {
	switch size {
	case atoms.SelectSizeXS:
		return schemaui.SizeXS
	case atoms.SelectSizeSM:
		return schemaui.SizeSM
	case atoms.SelectSizeLG:
		return schemaui.SizeLG
	case atoms.SelectSizeXL:
		return schemaui.SizeXL
	default:
		return schemaui.SizeMD
	}
}

// convertCheckboxSizeToSchemaSize maps Templ CheckboxSize to schema Size
func (b *Bridge) convertCheckboxSizeToSchemaSize(size atoms.CheckboxSize) schemaui.Size {
	switch size {
	case atoms.CheckboxSizeXS:
		return schemaui.SizeXS
	case atoms.CheckboxSizeSM:
		return schemaui.SizeSM
	case atoms.CheckboxSizeLG:
		return schemaui.SizeLG
	case atoms.CheckboxSizeXL:
		return schemaui.SizeXL
	default:
		return schemaui.SizeMD
	}
}

// convertRadioSizeToSchemaSize maps Templ RadioSize to schema Size
func (b *Bridge) convertRadioSizeToSchemaSize(size atoms.RadioSize) schemaui.Size {
	switch size {
	case atoms.RadioSizeXS:
		return schemaui.SizeXS
	case atoms.RadioSizeSM:
		return schemaui.SizeSM
	case atoms.RadioSizeLG:
		return schemaui.SizeLG
	case atoms.RadioSizeXL:
		return schemaui.SizeXL
	default:
		return schemaui.SizeMD
	}
}

// convertStringToPosition converts string to schema Position
func (b *Bridge) convertStringToPosition(pos string) schemaui.Position {
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

// convertStringToButtonType converts string to schema ButtonType
func (b *Bridge) convertStringToButtonType(buttonType string) schemaui.ButtonType {
	switch buttonType {
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

// convertStringToInputType converts string to schema InputType
func (b *Bridge) convertStringToInputType(inputType string) schemaui.InputType {
	switch inputType {
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
	case "hidden":
		return schemaui.InputHidden
	default:
		return schemaui.InputText
	}
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// generateComponentID generates a unique component ID
func generateComponentID(componentType string) string {
	return fmt.Sprintf("%s-%s", componentType, strings.Replace(uuid.New().String()[:8], "-", "", -1))
}

// ConvertTemplComponentsToSchema converts multiple Templ components to schema components
func (b *Bridge) ConvertTemplComponentsToSchema(ctx context.Context, templComponents []TemplComponent) ([]schemaui.Component, error) {
	var schemaComponents []schemaui.Component
	
	for _, templComponent := range templComponents {
		schemaComponent, err := b.ConvertTemplToSchema(ctx, templComponent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Templ component %s: %w", templComponent.Type, err)
		}
		schemaComponents = append(schemaComponents, schemaComponent)
	}
	
	return schemaComponents, nil
}

// GenerateSchemaFromTemplFile analyzes a Templ file and generates schema definitions
func (b *Bridge) GenerateSchemaFromTemplFile(ctx context.Context, templFilePath string) ([]schemaui.Component, error) {
	// This would require parsing the Templ file and extracting component usage
	// For now, we'll return a placeholder implementation
	// In a full implementation, this would parse the .templ file AST
	
	return nil, fmt.Errorf("Templ file parsing not yet implemented - use ConvertTemplToSchema for individual components")
}