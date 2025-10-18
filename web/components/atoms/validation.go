package atoms

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
)

// ============================================================================
// COMMON REGEX PATTERNS
// ============================================================================

// Common regex patterns for validation.
const (
	EmailRegex          = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	URLRegex            = `^https?://[^\s/$.?#].[^\s]*$`
	PhoneRegex          = `^\+?[1-9]\d{1,14}$` // E.164 format
	ZipCodeUSRegex      = `^\d{5}(-\d{4})?$`
	StrongPasswordRegex = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$`
	AlphanumericRegex   = `^[a-zA-Z0-9]+$`
	AlphaRegex          = `^[a-zA-Z]+$`
	NumericRegex        = `^\d+$`
	HexColorRegex       = `^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`
	IPv4Regex           = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`
	CreditCardRegex     = `^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13})$`
)

// ============================================================================
// REGEX CACHE (Thread-Safe)
// ============================================================================

var (
	regexCache      = make(map[string]*regexp.Regexp)
	regexCacheMutex sync.RWMutex
)

// getCompiledRegex retrieves a compiled regex from cache or compiles and caches it.
// Thread-safe implementation using RWMutex.
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	// Fast path: read lock
	regexCacheMutex.RLock()
	if cached, ok := regexCache[pattern]; ok {
		regexCacheMutex.RUnlock()
		return cached, nil
	}
	regexCacheMutex.RUnlock()

	// Slow path: compile and cache
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	regexCacheMutex.Lock()
	regexCache[pattern] = compiled
	regexCacheMutex.Unlock()

	return compiled, nil
}

// ClearRegexCache clears the regex cache (useful for testing).
func ClearRegexCache() {
	regexCacheMutex.Lock()
	defer regexCacheMutex.Unlock()
	regexCache = make(map[string]*regexp.Regexp)
}

// ============================================================================
// VALIDATOR INTERFACE
// ============================================================================

// Validator defines the interface for all validators.
// All validators must implement these two methods.
type Validator interface {
	Validate(value interface{}) error
	ErrorMessage() string
}

// ============================================================================
// REQUIRED VALIDATOR
// ============================================================================

// RequiredValidator ensures a value is present.
type RequiredValidator struct {
	Message string // Custom error message
}

// Validate checks if the value is not empty.
func (v RequiredValidator) Validate(value interface{}) error {
	switch val := value.(type) {
	case string:
		if len(val) == 0 || len(val) == len(val)-len(val) {
			// Check for whitespace-only strings
			trimmed := ""
			for _, r := range val {
				if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
					trimmed += string(r)
				}
			}
			if trimmed == "" {
				return errors.New(v.ErrorMessage())
			}
		}
	case nil:
		return errors.New(v.ErrorMessage())
	case []interface{}:
		if len(val) == 0 {
			return errors.New(v.ErrorMessage())
		}
	case map[string]interface{}:
		if len(val) == 0 {
			return errors.New(v.ErrorMessage())
		}
	}
	return nil
}

// ErrorMessage returns the error message.
func (v RequiredValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return "This field is required"
}

// ============================================================================
// MIN LENGTH VALIDATOR
// ============================================================================

// MinLengthValidator ensures minimum string length.
type MinLengthValidator struct {
	MinLength int
	Message   string
}

// Validate checks if the string meets minimum length.
func (v MinLengthValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	if len(str) < v.MinLength {
		return errors.New(v.ErrorMessage())
	}
	return nil
}

// ErrorMessage returns the error message.
func (v MinLengthValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return fmt.Sprintf("Must be at least %d characters", v.MinLength)
}

// ============================================================================
// MAX LENGTH VALIDATOR
// ============================================================================

// MaxLengthValidator ensures maximum string length.
type MaxLengthValidator struct {
	MaxLength int
	Message   string
}

// Validate checks if the string is within maximum length.
func (v MaxLengthValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	if len(str) > v.MaxLength {
		return errors.New(v.ErrorMessage())
	}
	return nil
}

// ErrorMessage returns the error message.
func (v MaxLengthValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return fmt.Sprintf("Must be no more than %d characters", v.MaxLength)
}

// ============================================================================
// PATTERN VALIDATOR
// ============================================================================

// PatternValidator validates against a regex pattern.
type PatternValidator struct {
	Pattern string
	Message string
}

// Validate checks if the value matches the pattern.
func (v PatternValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	regex, err := getCompiledRegex(v.Pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	if !regex.MatchString(str) {
		return errors.New(v.ErrorMessage())
	}
	return nil
}

// ErrorMessage returns the error message.
func (v PatternValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return "Invalid format"
}

// ============================================================================
// EMAIL VALIDATOR
// ============================================================================

// EmailValidator validates email addresses.
type EmailValidator struct {
	Message string
}

// Validate checks if the value is a valid email.
func (v EmailValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	regex, err := getCompiledRegex(EmailRegex)
	if err != nil {
		return err
	}

	if !regex.MatchString(str) {
		return errors.New(v.ErrorMessage())
	}
	return nil
}

// ErrorMessage returns the error message.
func (v EmailValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return "Please enter a valid email address"
}

// ============================================================================
// URL VALIDATOR
// ============================================================================

// URLValidator validates URLs.
type URLValidator struct {
	Message string
}

// Validate checks if the value is a valid URL.
func (v URLValidator) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	regex, err := getCompiledRegex(URLRegex)
	if err != nil {
		return err
	}

	if !regex.MatchString(str) {
		return errors.New(v.ErrorMessage())
	}
	return nil
}

// ErrorMessage returns the error message.
func (v URLValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return "Please enter a valid URL"
}

// ============================================================================
// RANGE VALIDATOR
// ============================================================================

// RangeValidator validates numeric ranges.
type RangeValidator struct {
	Min     float64
	Max     float64
	Message string
}

// Validate checks if the value is within the specified range.
func (v RangeValidator) Validate(value interface{}) error {
	var num float64

	switch val := value.(type) {
	case float64:
		num = val
	case float32:
		num = float64(val)
	case int:
		num = float64(val)
	case int64:
		num = float64(val)
	case int32:
		num = float64(val)
	case string:
		parsed, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("value must be numeric")
		}
		num = parsed
	default:
		return fmt.Errorf("value must be numeric")
	}

	if num < v.Min || num > v.Max {
		return errors.New(v.ErrorMessage())
	}
	return nil
}

// ErrorMessage returns the error message.
func (v RangeValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return fmt.Sprintf("Must be between %.2f and %.2f", v.Min, v.Max)
}

// ============================================================================
// CUSTOM VALIDATOR
// ============================================================================

// CustomValidator allows custom validation logic.
type CustomValidator struct {
	ValidateFunc func(value interface{}) error
	Message      string
}

// Validate runs the custom validation function.
func (v CustomValidator) Validate(value interface{}) error {
	if v.ValidateFunc == nil {
		return nil
	}
	return v.ValidateFunc(value)
}

// ErrorMessage returns the error message.
func (v CustomValidator) ErrorMessage() string {
	if v.Message != "" {
		return v.Message
	}
	return "Validation failed"
}

// ============================================================================
// VALIDATION CHAIN
// ============================================================================

// ValidationChain runs multiple validators in sequence.
type ValidationChain struct {
	Validators []Validator
}

// Validate runs all validators and returns the first error encountered.
func (vc ValidationChain) Validate(value interface{}) error {
	for _, validator := range vc.Validators {
		if err := validator.Validate(value); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAll runs all validators and returns all errors.
func (vc ValidationChain) ValidateAll(value interface{}) []error {
	var errors []error
	for _, validator := range vc.Validators {
		if err := validator.Validate(value); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// ErrorMessage returns a combined error message.
func (vc ValidationChain) ErrorMessage() string {
	if len(vc.Validators) > 0 {
		return vc.Validators[0].ErrorMessage()
	}
	return "Validation failed"
}

// ============================================================================
// FIELD STATE
// ============================================================================

// FieldState represents the state of a single form field.
type FieldState struct {
	Name       string      // Field identifier
	Value      interface{} // Current value
	IsDirty    bool        // Has been modified
	IsTouched  bool        // Has been focused
	IsValid    bool        // Passes validation
	Errors     []string    // Validation errors
	Validators []Validator // Field validators
}

// NewFieldState creates a new field state.
func NewFieldState(name string, validators ...Validator) *FieldState {
	return &FieldState{
		Name:       name,
		Validators: validators,
		IsValid:    true,
		Errors:     []string{},
	}
}

// Validate runs all validators for this field.
func (fs *FieldState) Validate() {
	fs.Errors = nil
	fs.IsValid = true

	for _, validator := range fs.Validators {
		if err := validator.Validate(fs.Value); err != nil {
			fs.Errors = append(fs.Errors, err.Error())
			fs.IsValid = false
		}
	}
}

// GetFirstError returns the first error message or empty string.
func (fs *FieldState) GetFirstError() string {
	if len(fs.Errors) > 0 {
		return fs.Errors[0]
	}
	return ""
}

// MarkAsTouched marks the field as touched (focused at least once).
func (fs *FieldState) MarkAsTouched() {
	fs.IsTouched = true
}

// MarkAsDirty marks the field as dirty (value changed).
func (fs *FieldState) MarkAsDirty() {
	fs.IsDirty = true
}

// SetValue updates the field value and marks as dirty.
func (fs *FieldState) SetValue(value interface{}) {
	fs.Value = value
	fs.MarkAsDirty()
}

// Reset resets the field to initial state.
func (fs *FieldState) Reset() {
	fs.Value = nil
	fs.IsDirty = false
	fs.IsTouched = false
	fs.IsValid = true
	fs.Errors = nil
}

// ============================================================================
// FORM STATE
// ============================================================================

// FormState manages state for an entire form.
type FormState struct {
	Fields       map[string]*FieldState
	IsSubmitting bool
	IsValid      bool
	SubmitCount  int
	mu           sync.RWMutex // Protect concurrent access
}

// NewFormState creates a new form state manager.
func NewFormState() *FormState {
	return &FormState{
		Fields:  make(map[string]*FieldState),
		IsValid: true,
	}
}

// RegisterField adds a field to the form.
func (fs *FormState) RegisterField(name string, validators ...Validator) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.Fields[name] = NewFieldState(name, validators...)
}

// SetFieldValue updates a field's value.
func (fs *FormState) SetFieldValue(name string, value interface{}) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if field, ok := fs.Fields[name]; ok {
		field.SetValue(value)
	}
}

// ValidateField validates a specific field.
func (fs *FormState) ValidateField(name string) bool {
	fs.mu.RLock()
	field, ok := fs.Fields[name]
	fs.mu.RUnlock()

	if !ok {
		return false
	}

	field.Validate()
	return field.IsValid
}

// ValidateAll validates all fields in the form.
func (fs *FormState) ValidateAll() bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.IsValid = true

	for _, field := range fs.Fields {
		field.Validate()
		if !field.IsValid {
			fs.IsValid = false
		}
	}

	return fs.IsValid
}

// ValidateTouchedFields validates only fields that have been touched.
func (fs *FormState) ValidateTouchedFields() bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	allValid := true

	for _, field := range fs.Fields {
		if field.IsTouched {
			field.Validate()
			if !field.IsValid {
				allValid = false
			}
		}
	}

	return allValid
}

// GetFieldErrors returns all errors for a specific field.
func (fs *FormState) GetFieldErrors(name string) []string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	if field, ok := fs.Fields[name]; ok {
		return field.Errors
	}
	return nil
}

// GetAllErrors returns a map of all field errors.
func (fs *FormState) GetAllErrors() map[string][]string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	errors := make(map[string][]string)

	for name, field := range fs.Fields {
		if len(field.Errors) > 0 {
			errors[name] = field.Errors
		}
	}

	return errors
}

// Reset resets the form to initial state.
func (fs *FormState) Reset() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, field := range fs.Fields {
		field.Reset()
	}
	fs.IsSubmitting = false
	fs.IsValid = true
	fs.SubmitCount = 0
}

// ToJSON serializes form state to JSON.
func (fs *FormState) ToJSON() ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	type exportField struct {
		Value     interface{} `json:"value"`
		IsDirty   bool        `json:"isDirty"`
		IsTouched bool        `json:"isTouched"`
		IsValid   bool        `json:"isValid"`
		Errors    []string    `json:"errors,omitempty"`
	}

	export := make(map[string]exportField)
	for name, field := range fs.Fields {
		export[name] = exportField{
			Value:     field.Value,
			IsDirty:   field.IsDirty,
			IsTouched: field.IsTouched,
			IsValid:   field.IsValid,
			Errors:    field.Errors,
		}
	}

	return ToJSON(export)
}

// FromJSON deserializes form state from JSON.
func (fs *FormState) FromJSON(data []byte) error {
	type importField struct {
		Value     interface{} `json:"value"`
		IsDirty   bool        `json:"isDirty"`
		IsTouched bool        `json:"isTouched"`
	}

	imported := make(map[string]importField)
	if err := FromJSON(data, &imported); err != nil {
		return err
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	for name, field := range imported {
		if formField, exists := fs.Fields[name]; exists {
			formField.Value = field.Value
			formField.IsDirty = field.IsDirty
			formField.IsTouched = field.IsTouched
			formField.Validate() // Re-validate after loading
		}
	}

	return nil
}

// ============================================================================
// VALIDATION PROPS BUILDER
// ============================================================================

// ValidationPropsBuilder provides fluent API for building ValidationProps.
type ValidationPropsBuilder struct {
	props ValidationProps
}

// NewValidationProps creates a new ValidationPropsBuilder with defaults.
func NewValidationProps() *ValidationPropsBuilder {
	return &ValidationPropsBuilder{
		props: ValidationProps{
			State: StateDefault,
		},
	}
}

// WithState sets the validation state.
func (b *ValidationPropsBuilder) WithState(state ValidationState) *ValidationPropsBuilder {
	b.props.State = state
	return b
}

// WithError sets error state and message.
func (b *ValidationPropsBuilder) WithError(message string) *ValidationPropsBuilder {
	b.props.State = StateError
	b.props.ErrorText = message
	b.props.ShowValidation = true
	return b
}

// WithSuccess sets success state and message.
func (b *ValidationPropsBuilder) WithSuccess(message string) *ValidationPropsBuilder {
	b.props.State = StateSuccess
	b.props.SuccessText = message
	b.props.ShowValidation = true
	return b
}

// WithWarning sets warning state and message.
func (b *ValidationPropsBuilder) WithWarning(message string) *ValidationPropsBuilder {
	b.props.State = StateWarning
	b.props.WarningText = message
	b.props.ShowValidation = true
	return b
}

// WithInfo sets info state and message.
func (b *ValidationPropsBuilder) WithInfo(message string) *ValidationPropsBuilder {
	b.props.State = StateInfo
	b.props.InfoText = message
	b.props.ShowValidation = true
	return b
}

// WithHelpText sets general help text.
func (b *ValidationPropsBuilder) WithHelpText(text string) *ValidationPropsBuilder {
	b.props.HelpText = text
	return b
}

// ShowValidationUI enables validation UI display.
func (b *ValidationPropsBuilder) ShowValidationUI(show bool) *ValidationPropsBuilder {
	b.props.ShowValidation = show
	return b
}

// Build returns the configured ValidationProps.
func (b *ValidationPropsBuilder) Build() ValidationProps {
	return b.props
}
