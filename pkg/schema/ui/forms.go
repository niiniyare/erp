package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Input component configuration
type InputConfig struct {
	InputType   InputType `json:"input_type" validate:"required"`
	Value       string    `json:"value,omitempty"`
	Placeholder string    `json:"placeholder,omitempty"`
	MaxLength   int       `json:"max_length,omitempty"`
	MinLength   int       `json:"min_length,omitempty"`
	Pattern     string    `json:"pattern,omitempty"`
	AutoFocus   bool      `json:"auto_focus,omitempty"`
	SpellCheck  bool      `json:"spell_check,omitempty"`
}

// InputType is already defined in types.go - no need to duplicate

// Textarea component configuration
type TextareaConfig struct {
	Value       string `json:"value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Rows        int    `json:"rows,omitempty"`
	Cols        int    `json:"cols,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
	Resize      string `json:"resize,omitempty"` // none, both, horizontal, vertical
	Wrap        string `json:"wrap,omitempty"`   // soft, hard
}

// Select component configuration
type SelectConfig struct {
	Options      []Option `json:"options" validate:"required"`
	Value        string   `json:"value,omitempty"`
	Multiple     bool     `json:"multiple,omitempty"`
	Searchable   bool     `json:"searchable,omitempty"`
	Clearable    bool     `json:"clearable,omitempty"`
	Placeholder  string   `json:"placeholder,omitempty"`
	LoadingText  string   `json:"loading_text,omitempty"`
	NoOptionsText string  `json:"no_options_text,omitempty"`
	DataSource   string   `json:"data_source,omitempty"` // API endpoint for dynamic options
}

// Option is already defined in types.go - no need to duplicate

// Checkbox component configuration
type CheckboxConfig struct {
	Value   bool   `json:"value,omitempty"`
	Label   string `json:"label,omitempty"`
	Checked bool   `json:"checked,omitempty"`
}

// Radio component configuration
type RadioConfig struct {
	Options   []Option `json:"options" validate:"required"` // Option is defined in types.go
	Value     string   `json:"value,omitempty"`
	Direction string   `json:"direction,omitempty"` // horizontal, vertical
}

// Button component configuration
type ButtonConfig struct {
	Text         string     `json:"text,omitempty"`
	Icon         string     `json:"icon,omitempty"`
	IconPosition Position   `json:"icon_position,omitempty"`
	ButtonType   ButtonType `json:"button_type,omitempty"`
	Action       string     `json:"action,omitempty"`
	Href         string     `json:"href,omitempty"`
	Target       string     `json:"target,omitempty"`
	Download     bool       `json:"download,omitempty"`
	Loading      bool       `json:"loading,omitempty"`
}

// ButtonType is already defined in types.go - no need to duplicate

// DatePicker component configuration
type DatePickerConfig struct {
	Value       string     `json:"value,omitempty"`
	Format      string     `json:"format,omitempty"`
	MinDate     string     `json:"min_date,omitempty"`
	MaxDate     string     `json:"max_date,omitempty"`
	ShowTime    bool       `json:"show_time,omitempty"`
	TimeFormat  string     `json:"time_format,omitempty"`
	Placeholder string     `json:"placeholder,omitempty"`
	Mode        DateMode   `json:"mode,omitempty"`
	FirstDayOfWeek int     `json:"first_day_of_week,omitempty"`
}

// DateMode represents different date picker modes
type DateMode string

const (
	DateModeDate     DateMode = "date"
	DateModeDateTime DateMode = "datetime"
	DateModeTime     DateMode = "time"
	DateModeMonth    DateMode = "month"
	DateModeYear     DateMode = "year"
	DateModeRange    DateMode = "range"
)

// FileUpload component configuration
type FileUploadConfig struct {
	Accept       []string `json:"accept,omitempty"`
	Multiple     bool     `json:"multiple,omitempty"`
	MaxSize      int64    `json:"max_size,omitempty"`
	MaxFiles     int      `json:"max_files,omitempty"`
	UploadURL    string   `json:"upload_url,omitempty"`
	PreviewType  string   `json:"preview_type,omitempty"` // image, list, grid
	DragDrop     bool     `json:"drag_drop,omitempty"`
	ShowProgress bool     `json:"show_progress,omitempty"`
	AutoUpload   bool     `json:"auto_upload,omitempty"`
}

// Form component configuration
type FormConfig struct {
	Method      string            `json:"method,omitempty"`
	Action      string            `json:"action,omitempty"`
	Enctype     string            `json:"enctype,omitempty"`
	NoValidate  bool              `json:"no_validate,omitempty"`
	AutoComplete string           `json:"autocomplete,omitempty"`
	Layout      FormLayout        `json:"layout,omitempty"`
	LabelWidth  string            `json:"label_width,omitempty"`
	Spacing     string            `json:"spacing,omitempty"`
	Validation  *FormValidation   `json:"validation,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// FormLayout represents form layout options
type FormLayout string

const (
	FormLayoutVertical   FormLayout = "vertical"
	FormLayoutHorizontal FormLayout = "horizontal"
	FormLayoutInline     FormLayout = "inline"
)

// FormValidation represents form-level validation configuration
type FormValidation struct {
	ValidateOnSubmit bool              `json:"validate_on_submit,omitempty"`
	ValidateOnChange bool              `json:"validate_on_change,omitempty"`
	ValidateOnBlur   bool              `json:"validate_on_blur,omitempty"`
	ShowErrors       bool              `json:"show_errors,omitempty"`
	ErrorPosition    Position          `json:"error_position,omitempty"`
	CustomRules      map[string]string `json:"custom_rules,omitempty"`
}

// Form component factories
type InputFactory struct{}
type TextareaFactory struct{}
type SelectFactory struct{}
type CheckboxFactory struct{}
type RadioFactory struct{}
type ButtonFactory struct{}
type DatePickerFactory struct{}
type FileUploadFactory struct{}
type FormFactory struct{}

// Input factory implementation
func (f *InputFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var inputConfig InputConfig
	if err := mapToStruct(config, &inputConfig); err != nil {
		return Component{}, fmt.Errorf("invalid input config: %w", err)
	}

	component := NewComponent(ComponentInput, generateID())
	return component.WithConfig(inputConfig).Build(), nil
}

func (f *InputFactory) Validate(ctx context.Context, component Component) error {
	var inputConfig InputConfig
	if err := json.Unmarshal(component.Config, &inputConfig); err != nil {
		return fmt.Errorf("invalid input config: %w", err)
	}

	if inputConfig.InputType == "" {
		return fmt.Errorf("input type is required")
	}

	return nil
}

func (f *InputFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentInput,
		Title:       "Input Field",
		Description: "A text input field for user data entry",
		Properties: map[string]Property{
			"input_type": {
				Type:        "string",
				Description: "The type of input field",
				Enum:        []string{"text", "email", "password", "number", "tel", "url", "search", "hidden"},
				Default:     "text",
			},
			"value": {
				Type:        "string",
				Description: "The current value of the input",
			},
			"placeholder": {
				Type:        "string",
				Description: "Placeholder text shown when input is empty",
			},
			"max_length": {
				Type:        "integer",
				Description: "Maximum number of characters allowed",
				Min:         new(float64),
			},
			"pattern": {
				Type:        "string",
				Description: "Regular expression pattern for validation",
			},
		},
		Required: []string{"input_type"},
		Examples: []map[string]any{
			{
				"input_type":  "email",
				"placeholder": "Enter your email",
				"required":    true,
			},
		},
	}
}

// Select factory implementation
func (f *SelectFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var selectConfig SelectConfig
	if err := mapToStruct(config, &selectConfig); err != nil {
		return Component{}, fmt.Errorf("invalid select config: %w", err)
	}

	component := NewComponent(ComponentSelect, generateID())
	return component.WithConfig(selectConfig).Build(), nil
}

func (f *SelectFactory) Validate(ctx context.Context, component Component) error {
	var selectConfig SelectConfig
	if err := json.Unmarshal(component.Config, &selectConfig); err != nil {
		return fmt.Errorf("invalid select config: %w", err)
	}

	if len(selectConfig.Options) == 0 && selectConfig.DataSource == "" {
		return fmt.Errorf("select must have options or data source")
	}

	return nil
}

func (f *SelectFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentSelect,
		Title:       "Select Dropdown",
		Description: "A dropdown selection component",
		Properties: map[string]Property{
			"options": {
				Type:        "array",
				Description: "Static options for the select",
			},
			"multiple": {
				Type:        "boolean",
				Description: "Allow multiple selections",
				Default:     false,
			},
			"searchable": {
				Type:        "boolean",
				Description: "Enable search within options",
				Default:     false,
			},
			"data_source": {
				Type:        "string",
				Description: "API endpoint for dynamic options",
			},
		},
		Examples: []map[string]any{
			{
				"options": []Option{
					{Value: "option1", Label: "Option 1"},
					{Value: "option2", Label: "Option 2"},
				},
				"searchable": true,
			},
		},
	}
}

// Button factory implementation
func (f *ButtonFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var buttonConfig ButtonConfig
	if err := mapToStruct(config, &buttonConfig); err != nil {
		return Component{}, fmt.Errorf("invalid button config: %w", err)
	}

	component := NewComponent(ComponentButton, generateID())
	return component.WithConfig(buttonConfig).Build(), nil
}

func (f *ButtonFactory) Validate(ctx context.Context, component Component) error {
	var buttonConfig ButtonConfig
	if err := json.Unmarshal(component.Config, &buttonConfig); err != nil {
		return fmt.Errorf("invalid button config: %w", err)
	}

	if buttonConfig.Text == "" && buttonConfig.Icon == "" {
		return fmt.Errorf("button must have text or icon")
	}

	return nil
}

func (f *ButtonFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentButton,
		Title:       "Button",
		Description: "An interactive button component",
		Properties: map[string]Property{
			"text": {
				Type:        "string",
				Description: "Button text content",
			},
			"icon": {
				Type:        "string",
				Description: "Icon identifier",
			},
			"button_type": {
				Type:        "string",
				Description: "HTML button type",
				Enum:        []string{"button", "submit", "reset"},
				Default:     "button",
			},
			"action": {
				Type:        "string",
				Description: "Action to perform when clicked",
			},
		},
		Examples: []map[string]any{
			{
				"text":        "Submit",
				"button_type": "submit",
				"variant":     "primary",
			},
		},
	}
}

// Utility functions
func mapToStruct(m map[string]any, target any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func generateID() string {
	// Simple ID generation - in production use UUID or similar
	return fmt.Sprintf("component_%d", time.Now().UnixNano())
}

// Textarea factory implementation
func (f *TextareaFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var textareaConfig TextareaConfig
	if err := mapToStruct(config, &textareaConfig); err != nil {
		return Component{}, fmt.Errorf("invalid textarea config: %w", err)
	}
	component := NewComponent(ComponentTextarea, generateID())
	return component.WithConfig(textareaConfig).Build(), nil
}

func (f *TextareaFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *TextareaFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentTextarea,
		Title:       "Textarea",
		Description: "A multi-line text input field",
		Properties: map[string]Property{
			"rows": {Type: "integer", Description: "Number of rows", Default: 4},
			"cols": {Type: "integer", Description: "Number of columns", Default: 40},
		},
	}
}

// Checkbox factory implementation
func (f *CheckboxFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var checkboxConfig CheckboxConfig
	if err := mapToStruct(config, &checkboxConfig); err != nil {
		return Component{}, fmt.Errorf("invalid checkbox config: %w", err)
	}
	component := NewComponent(ComponentCheckbox, generateID())
	return component.WithConfig(checkboxConfig).Build(), nil
}

func (f *CheckboxFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *CheckboxFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentCheckbox,
		Title:       "Checkbox",
		Description: "A checkbox input field",
	}
}

// Radio factory implementation
func (f *RadioFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var radioConfig RadioConfig
	if err := mapToStruct(config, &radioConfig); err != nil {
		return Component{}, fmt.Errorf("invalid radio config: %w", err)
	}
	component := NewComponent(ComponentRadio, generateID())
	return component.WithConfig(radioConfig).Build(), nil
}

func (f *RadioFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *RadioFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentRadio,
		Title:       "Radio Button Group",
		Description: "A group of radio button options",
	}
}

// DatePicker factory implementation
func (f *DatePickerFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var datePickerConfig DatePickerConfig
	if err := mapToStruct(config, &datePickerConfig); err != nil {
		return Component{}, fmt.Errorf("invalid date picker config: %w", err)
	}
	component := NewComponent(ComponentDatePicker, generateID())
	return component.WithConfig(datePickerConfig).Build(), nil
}

func (f *DatePickerFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *DatePickerFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentDatePicker,
		Title:       "Date Picker",
		Description: "A date selection component",
	}
}

// FileUpload factory implementation
func (f *FileUploadFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var fileUploadConfig FileUploadConfig
	if err := mapToStruct(config, &fileUploadConfig); err != nil {
		return Component{}, fmt.Errorf("invalid file upload config: %w", err)
	}
	component := NewComponent(ComponentFileUpload, generateID())
	return component.WithConfig(fileUploadConfig).Build(), nil
}

func (f *FileUploadFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *FileUploadFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentFileUpload,
		Title:       "File Upload",
		Description: "A file upload component",
	}
}

// Form factory implementation
func (f *FormFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var formConfig FormConfig
	if err := mapToStruct(config, &formConfig); err != nil {
		return Component{}, fmt.Errorf("invalid form config: %w", err)
	}
	component := NewComponent(ComponentForm, generateID())
	return component.WithConfig(formConfig).Build(), nil
}

func (f *FormFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *FormFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentForm,
		Title:       "Form",
		Description: "A form container component",
	}
}