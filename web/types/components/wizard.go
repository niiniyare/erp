// Package components - WizardSchema for multi-step wizard forms  
// Based on JSON schema: WizardSchema.json
package components

import (
	"encoding/json"
	"fmt"
)

// WizardMode represents the display orientation of the wizard
type WizardMode string

const (
	WizardModeHorizontal WizardMode = "horizontal" // Steps displayed horizontally (default)
	WizardModeVertical   WizardMode = "vertical"   // Steps displayed vertically
)

// WizardAffixFooter represents when to pin footer buttons
type WizardAffixFooter string

const (
	WizardAffixAlways WizardAffixFooter = "always" // Always pin footer
	WizardAffixNever  WizardAffixFooter = ""       // Never pin footer (default)
)

// WizardStepJumpMode represents how users can navigate between steps
type WizardStepJumpMode string

const (
	StepJumpDisabled WizardStepJumpMode = "disabled" // Cannot jump to steps
	StepJumpEnabled  WizardStepJumpMode = "enabled"  // Can jump to any step
	StepJumpPrevious WizardStepJumpMode = "previous" // Can only jump to previous steps
)

// LoadingConfig represents loading state configuration
type LoadingConfig struct {
	// Whether to show loading indicator
	Show bool `json:"show,omitempty"`
	// Loading icon name or URL
	Icon string `json:"icon,omitempty"`
	// Loading text message
	Text string `json:"text,omitempty"`
}

// WizardStepSchema represents a single step in the wizard
type WizardStepSchema struct {
	// Step identification and display
	// Step title displayed in the step indicator
	Title string `json:"title,omitempty"`
	// Step subtitle for additional context
	SubTitle string `json:"subTitle,omitempty"`
	// Detailed description of the step
	Description string `json:"description,omitempty"`
	// Icon name or URL for the step indicator
	Icon string `json:"icon,omitempty"`
	// Unique value identifier for the step
	Value string `json:"value,omitempty"`

	// Step content configuration
	// Array of form components and layouts for this step
	Body []any `json:"body,omitempty"`
	
	// API Configuration for the step
	// Main API for this step (usually for data submission)
	API *APIConfig `json:"api,omitempty"`
	// API to fetch initial data for this step
	InitAPI *APIConfig `json:"initApi,omitempty"`
	// Async API for background processing
	AsyncAPI *APIConfig `json:"asyncApi,omitempty"`
	// Async API for initial data loading
	InitAsyncAPI *APIConfig `json:"initAsyncApi,omitempty"`
	// API called when jumping to this step
	JumpAPI *APIConfig `json:"jumpApi,omitempty"`

	// Step behavior and validation
	// Expression to determine if step can be jumped to
	JumpableOn string `json:"jumpableOn,omitempty"`
	// Expression to control step visibility
	VisibleOn string `json:"visibleOn,omitempty"`
	// Expression to hide the step
	HiddenOn string `json:"hiddenOn,omitempty"`
	// Expression to disable the step
	DisabledOn string `json:"disabledOn,omitempty"`
	
	// Step validation API
	ValidateAPI *APIConfig `json:"validateApi,omitempty"`
	
	// Step-specific actions (buttons)
	Actions []ActionSchema `json:"actions,omitempty"`
	
	// Dialog configuration for this step
	Dialog *DialogSchema `json:"dialog,omitempty"`

	// Step styling and layout
	// CSS class name for the step
	ClassName string `json:"className,omitempty"`
	// CSS class name for the step body
	BodyClassName string `json:"bodyClassName,omitempty"`
	// Whether to wrap step content in a panel
	WrapWithPanel bool `json:"wrapWithPanel,omitempty"`
	// CSS class name for the panel wrapper
	PanelClassName string `json:"panelClassName,omitempty"`
}

// WizardSchema represents a multi-step form wizard component
// This component guides users through complex forms by breaking them into manageable steps
// with navigation, validation, and progress tracking
// Based on AMis Wizard: https://aisuda.bce.baidu.com/amis/zh-CN/components/wizard
type WizardSchema struct {
	BaseComponentProps

	// Component type identifier (required)
	Type string `json:"type"` // Must be "wizard"

	// Wizard Configuration
	// Display orientation of the wizard steps
	Mode WizardMode `json:"mode,omitempty"`
	
	// Array of wizard steps (required)
	Steps []WizardStepSchema `json:"steps,omitempty"`
	
	// Initial step to display (step index or value)
	StartStep any `json:"startStep,omitempty"` // number or string

	// API Configuration
	// API to fetch initial wizard data
	InitAPI *APIConfig `json:"initApi,omitempty"`
	// Async API for initial data loading with progress
	InitAsyncAPI *APIConfig `json:"initAsyncApi,omitempty"`
	// Main API for final wizard submission
	API *APIConfig `json:"api,omitempty"`
	// Async API for final submission with progress
	AsyncAPI *APIConfig `json:"asyncApi,omitempty"`

	// Async processing configuration
	// Field name indicating async operation completion
	FinishedField string `json:"finishedField,omitempty"`
	// Field name for init async operation completion
	InitFinishedField string `json:"initFinishedField,omitempty"`

	// Form behavior
	// Whether the entire wizard is read-only
	ReadOnly bool `json:"readOnly,omitempty"`
	// Loading state configuration
	LoadingConfig *LoadingConfig `json:"loadingConfig,omitempty"`

	// Navigation and submission
	// Page to redirect to after successful submission
	Redirect string `json:"redirect,omitempty"`
	// Component to reload after submission
	Reload string `json:"reload,omitempty"`
	// Target for form submission
	Target string `json:"target,omitempty"`
	
	// Navigation button configuration
	// Whether previous button is disabled
	ActionPrevDisabled bool `json:"actionPrevDisabled,omitempty"`
	// Whether next button is disabled  
	ActionNextDisabled bool `json:"actionNextDisabled,omitempty"`
	
	// Layout and styling
	// Whether to wrap entire wizard in a panel
	WrapWithPanel bool `json:"wrapWithPanel,omitempty"`
	// Whether to pin footer buttons to bottom
	AffixFooter any `json:"affixFooter,omitempty"` // bool or "always"

	// CSS class names for styling
	// CSS class for steps indicator area
	StepsClassName string `json:"stepsClassName,omitempty"`
	// CSS class for step content body area
	BodyClassName string `json:"bodyClassName,omitempty"`
	// CSS class for individual steps
	StepClassName string `json:"stepClassName,omitempty"`
	// CSS class for footer action area
	FooterClassName string `json:"footerClassName,omitempty"`
	// CSS class for action buttons
	ActionClassName string `json:"actionClassName,omitempty"`

	// Button labels (internationalization support)
	// Text for the finish/submit button
	ActionFinishLabel string `json:"actionFinishLabel,omitempty"`
	// Text for the next step button
	ActionNextLabel string `json:"actionNextLabel,omitempty"`
	// Text for next button when it also saves
	ActionNextSaveLabel string `json:"actionNextSaveLabel,omitempty"`
	// Text for the previous step button
	ActionPrevLabel string `json:"actionPrevLabel,omitempty"`

	// Advanced features
	// Whether to submit all steps data together at the end
	BulkSubmit bool `json:"bulkSubmit,omitempty"`
	// Field name to track individual step completion
	StepFinishedField string `json:"stepFinishedField,omitempty"`
	
	// Form name for data binding and reference
	Name string `json:"name,omitempty"`
	
	// Initial data for the wizard
	Data map[string]any `json:"data,omitempty"`

	// Design-time Configuration
	EditorSetting *EditorSetting `json:"editorSetting,omitempty"`
}

// Factory function to create a basic horizontal wizard
func NewWizard(steps []WizardStepSchema) *WizardSchema {
	return &WizardSchema{
		Type:                "wizard",
		Mode:                WizardModeHorizontal,
		Steps:               steps,
		ActionFinishLabel:   "Complete",
		ActionNextLabel:     "Next",
		ActionPrevLabel:     "Previous",
		WrapWithPanel:       true,
		AffixFooter:         false,
		BulkSubmit:          true,
		ActionPrevDisabled:  false,
		ActionNextDisabled:  false,
	}
}

// Factory function to create a vertical wizard
func NewVerticalWizard(steps []WizardStepSchema) *WizardSchema {
	wizard := NewWizard(steps)
	wizard.Mode = WizardModeVertical
	return wizard
}

// Factory function to create a simple 3-step wizard
func NewThreeStepWizard(step1, step2, step3 WizardStepSchema) *WizardSchema {
	return NewWizard([]WizardStepSchema{step1, step2, step3})
}

// Factory function to create a wizard step
func NewWizardStep(title string, body []any) WizardStepSchema {
	return WizardStepSchema{
		Title:         title,
		Body:          body,
		WrapWithPanel: true,
	}
}

// Factory function to create a form wizard step
func NewFormWizardStep(title string, form *FormSchema) WizardStepSchema {
	return WizardStepSchema{
		Title:         title,
		Body:          []any{form},
		WrapWithPanel: true,
	}
}

// Helper method to add a step to the wizard
func (w *WizardSchema) AddStep(step WizardStepSchema) {
	w.Steps = append(w.Steps, step)
}

// Helper method to set button labels for internationalization
func (w *WizardSchema) SetButtonLabels(next, prev, finish string) {
	w.ActionNextLabel = next
	w.ActionPrevLabel = prev
	w.ActionFinishLabel = finish
}

// Helper method to configure API endpoints
func (w *WizardSchema) SetAPIs(initAPI, submitAPI *APIConfig) {
	w.InitAPI = initAPI
	w.API = submitAPI
}

// Validation function for WizardSchema
func (w *WizardSchema) Validate() error {
	if w.Type != "wizard" {
		return fmt.Errorf("invalid wizard type: %s, must be 'wizard'", w.Type)
	}
	
	if len(w.Steps) == 0 {
		return fmt.Errorf("wizard must have at least one step")
	}

	// Validate mode enum
	if w.Mode != "" &&
		w.Mode != WizardModeHorizontal &&
		w.Mode != WizardModeVertical {
		return fmt.Errorf("invalid mode: %s", w.Mode)
	}

	// Validate each step
	for i, step := range w.Steps {
		if step.Title == "" {
			return fmt.Errorf("step %d must have a title", i)
		}
		if len(step.Body) == 0 {
			return fmt.Errorf("step %d (%s) must have body content", i, step.Title)
		}
	}

	// Validate startStep if specified
	if w.StartStep != nil {
		switch v := w.StartStep.(type) {
		case float64:
			if int(v) < 0 || int(v) >= len(w.Steps) {
				return fmt.Errorf("startStep index %d is out of range (0-%d)", int(v), len(w.Steps)-1)
			}
		case string:
			// Validate that string value exists in steps
			found := false
			for _, step := range w.Steps {
				if step.Value == v {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("startStep value '%s' not found in steps", v)
			}
		default:
			return fmt.Errorf("startStep must be number (index) or string (value)")
		}
	}

	return nil
}

// ToJSON converts WizardSchema to JSON string
func (w *WizardSchema) ToJSON() (string, error) {
	if err := w.Validate(); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal WizardSchema to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON creates WizardSchema from JSON string
func WizardFromJSON(jsonData string) (*WizardSchema, error) {
	var config WizardSchema
	err := json.Unmarshal([]byte(jsonData), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to WizardSchema: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}