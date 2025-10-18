# Shared Types Documentation

## Table of Contents

1. [Overview](#overview)
2. [Core Enumerations](#core-enumerations)
3. [Property Structs](#property-structs)
4. [Validation System](#validation-system)
5. [Form State Management](#form-state-management)
6. [Builder Pattern](#builder-pattern)
7. [Responsive Design](#responsive-design)
8. [Theming System](#theming-system)
9. [Utility Functions](#utility-functions)
10. [Best Practices](#best-practices)

---

## Overview

The shared types system provides reusable components for building form inputs and interactive UI elements. It follows these design principles:

- **DRY (Don't Repeat Yourself)**: Define common patterns once
- **Composition**: Components embed shared structs rather than inheriting
- **Type Safety**: Use enums instead of magic strings
- **Consistency**: All form components behave predictably
- **Extensibility**: Easy to add new states, sizes, or validation types

---

## Core Enumerations

### Size

Defines component sizing options used across all visual components.

```go
type Size string

const (
    SizeXS Size = "xs" // Extra small
    SizeSM Size = "sm" // Small
    SizeMD Size = "md" // Medium (default)
    SizeLG Size = "lg" // Large
    SizeXL Size = "xl" // Extra large
)
```

**Helper Functions:**

```go
AllSizes() []Size              // Returns all valid sizes
IsValidSize(s Size) bool       // Validates a size value
```

**Example:**

```go
size := SizeLG
if IsValidSize(size) {
    // Use size
}
```

---

### ValidationState

Defines the validation status of a form field.

```go
type ValidationState string

const (
    StateDefault ValidationState = "default" // Neutral state
    StateSuccess ValidationState = "success" // Valid input
    StateError   ValidationState = "error"   // Invalid input
    StateWarning ValidationState = "warning" // Valid but potentially problematic
    StateInfo    ValidationState = "info"    // Informational feedback
)
```

**Helper Functions:**

```go
AllValidationStates() []ValidationState
IsValidState(s ValidationState) bool
```

---

### LabelPosition

Defines where the label appears relative to the input.

```go
type LabelPosition string

const (
    LabelTop    LabelPosition = "top"    // Above input (default for text inputs)
    LabelBottom LabelPosition = "bottom" // Below input (rare)
    LabelLeft   LabelPosition = "left"   // Left of input (for checkboxes/radios)
    LabelRight  LabelPosition = "right"  // Right of input (default for checkboxes)
    LabelNone   LabelPosition = "none"   // No visible label (aria-label only)
)
```

---

### InputType

Defines HTML input types.

```go
type InputType string

const (
    InputTypeText     InputType = "text"
    InputTypeEmail    InputType = "email"
    InputTypePassword InputType = "password"
    InputTypeNumber   InputType = "number"
    InputTypeTel      InputType = "tel"
    InputTypeURL      InputType = "url"
    InputTypeSearch   InputType = "search"
    InputTypeDate     InputType = "date"
    InputTypeTime     InputType = "time"
    InputTypeDatetime InputType = "datetime-local"
    // ... and more
)
```

---

### Variant

Defines visual style variants for components.

```go
type Variant string

const (
    VariantDefault    Variant = "default"    // Standard styling
    VariantOutlined   Variant = "outlined"   // Outlined/bordered
    VariantFilled     Variant = "filled"     // Filled background
    VariantGhost      Variant = "ghost"      // Minimal styling
    VariantUnderlined Variant = "underlined" // Only bottom border
)
```

---

### ColorScheme

Defines color themes for components.

```go
type ColorScheme string

const (
    ColorDefault   ColorScheme = "default"
    ColorPrimary   ColorScheme = "primary"
    ColorSecondary ColorScheme = "secondary"
    ColorSuccess   ColorScheme = "success"
    ColorDanger    ColorScheme = "danger"
    ColorWarning   ColorScheme = "warning"
    ColorInfo      ColorScheme = "info"
    ColorGray      ColorScheme = "gray"
)
```

---

## Property Structs

### BaseProps

Fundamental HTML attributes shared by all components.

```go
type BaseProps struct {
    ID         string // HTML id attribute
    Name       string // Form field name
    Class      string // Custom CSS classes
    Style      string // Inline styles (use sparingly)
    DataTestID string // For testing frameworks
    TabIndex   int    // Custom tab order
}
```

**Methods:**

```go
HasID() bool                                    // Checks if ID is set
GetClasses(additionalClasses ...string) string // Combines classes
```

**Example:**

```go
base := BaseProps{
    ID:         "email-input",
    Name:       "email",
    DataTestID: "email-field",
}

classes := base.GetClasses("custom-input", "focus:ring-2")
// Result: "custom-input focus:ring-2"
```

---

### AccessibilityProps

ARIA and accessibility-related attributes.

```go
type AccessibilityProps struct {
    AriaLabel       string // Accessible name
    AriaDescribedBy string // ID of describing element
    AriaLabelledBy  string // ID of labeling element
    AriaRequired    bool   // Marks field as required
    AriaInvalid     bool   // Marks field as invalid
    AriaDisabled    bool   // Marks field as disabled
    AriaPlaceholder string // Accessible placeholder
    AriaControls    string // ID of controlled element
    AriaExpanded    *bool  // Expandable state
    AriaHasPopup    string // Popup type
    AriaLive        string // Live region
    AriaHidden      bool   // Hidden from screen readers
    Role            string // ARIA role override
}
```

**Methods:**

```go
GetAriaAttributes() map[string]string // Returns only non-empty attributes
```

**Example:**

```go
a11y := AccessibilityProps{
    AriaLabel:    "Email address",
    AriaRequired: true,
    AriaInvalid:  false,
}

attrs := a11y.GetAriaAttributes()
// Result: {"aria-label": "Email address", "aria-required": "true"}
```

---

### ValidationProps

Handles validation state and feedback messages.

```go
type ValidationProps struct {
    State            ValidationState
    HelpText         string // General guidance
    ErrorText        string // Error message
    SuccessText      string // Success message
    WarningText      string // Warning message
    InfoText         string // Info message
    ShowValidation   bool   // Show validation UI
    ValidateOnBlur   bool   // Trigger on blur
    ValidateOnChange bool   // Trigger on change
}
```

**Methods:**

```go
GetFeedbackMessage() (string, ValidationState) // Returns appropriate message
HasFeedback() bool                             // Checks if any feedback exists
IsValid() bool                                 // Checks if state is valid
IsInvalid() bool                               // Checks if state is invalid
```

**Example:**

```go
validation := ValidationProps{
    State:          StateError,
    ErrorText:      "Email is required",
    ShowValidation: true,
}

message, state := validation.GetFeedbackMessage()
// message: "Email is required", state: StateError

if validation.IsInvalid() {
    // Show error UI
}
```

---

### InteractionProps

Handles user interaction states.

```go
type InteractionProps struct {
    Disabled  bool // Component is disabled
    Required  bool // Field is required
    ReadOnly  bool // Field is read-only
    AutoFocus bool // Auto-focus on page load
}
```

**Methods:**

```go
IsInteractive() bool      // Returns true if not disabled/readonly
ShouldShowRequired() bool // Returns true if required and not disabled
```

**Example:**

```go
interaction := InteractionProps{
    Required: true,
    Disabled: false,
}

if interaction.IsInteractive() {
    // Allow user input
}

if interaction.ShouldShowRequired() {
    // Show required indicator (*)
}
```

---

### AlpineEventHandlers

Alpine.js event handling directives.

```go
type AlpineEventHandlers struct {
    OnChange     string // x-on:change
    OnInput      string // x-on:input
    OnFocus      string // x-on:focus
    OnBlur       string // x-on:blur
    OnClick      string // x-on:click
    OnKeyDown    string // x-on:keydown
    OnKeyUp      string // x-on:keyup
    OnMouseEnter string // x-on:mouseenter
    OnMouseLeave string // x-on:mouseleave
}
```

**Methods:**

```go
GetEventAttributes() map[string]string // Returns x-on: prefixed attributes
HasEventHandlers() bool                // Checks if any handlers defined
```

**Example:**

```go
events := AlpineEventHandlers{
    OnChange: "validateEmail($event.target.value)",
    OnBlur:   "touched = true",
}

attrs := events.GetEventAttributes()
// Result: {
//   "x-on:change": "validateEmail($event.target.value)",
//   "x-on:blur": "touched = true"
// }
```

---

### StyleProps

Visual styling options.

```go
type StyleProps struct {
    Size        Size        // Component size
    Variant     Variant     // Visual variant
    ColorScheme ColorScheme // Color theme
    Rounded     bool        // Rounded corners
    Shadow      bool        // Drop shadow
    FullWidth   bool        // Take full container width
}
```

**Methods:**

```go
GetSizeClass(componentType string) []string // Returns size-specific classes
```

**Example:**

```go
style := StyleProps{
    Size:      SizeLG,
    Variant:   VariantOutlined,
    Rounded:   true,
    FullWidth: true,
}

classes := style.GetSizeClass("input")
// Returns: ["text-base", "py-3", "px-4"]
```

---

### LabelProps

Label-specific properties.

```go
type LabelProps struct {
    Label         string        // Label text
    LabelPosition LabelPosition // Label position
    LabelClass    string        // Custom label classes
    HideLabel     bool          // Visually hide (keep for a11y)
}
```

**Methods:**

```go
HasLabel() bool           // Checks if label text exists
ShouldRenderLabel() bool  // Checks if label should be visible
```

---

### PlaceholderProps

Placeholder-related properties.

```go
type PlaceholderProps struct {
    Placeholder      string // Placeholder text
    FloatingLabel    bool   // Material-style floating label
    PlaceholderClass string // Custom placeholder styling
}
```

---

## Validation System

### Validator Interface

All validators implement this interface:

```go
type Validator interface {
    Validate(value interface{}) error
    ErrorMessage() string
}
```

---

### Built-in Validators

#### RequiredValidator

Ensures a value is present.

```go
validator := RequiredValidator{
    Message: "This field is required", // Optional custom message
}

err := validator.Validate("")
// Returns error: "This field is required"
```

---

#### MinLengthValidator

Ensures minimum string length.

```go
validator := MinLengthValidator{
    MinLength: 8,
    Message:   "Password must be at least 8 characters",
}

err := validator.Validate("short")
// Returns error: "Password must be at least 8 characters"
```

---

#### MaxLengthValidator

Ensures maximum string length.

```go
validator := MaxLengthValidator{
    MaxLength: 100,
    Message:   "Bio must be no more than 100 characters",
}
```

---

#### PatternValidator

Validates against a regex pattern.

```go
validator := PatternValidator{
    Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
    Message: "Invalid email format",
}
```

---

#### RangeValidator

Validates numeric ranges.

```go
validator := RangeValidator{
    Min:     18,
    Max:     99,
    Message: "Age must be between 18 and 99",
}

err := validator.Validate(15.0)
// Returns error: "Age must be between 18 and 99"
```

---

#### CustomValidator

Allows custom validation logic.

```go
validator := CustomValidator{
    Message: "Passwords must match",
    ValidateFunc: func(value interface{}) error {
        password := value.(string)
        if password != confirmPassword {
            return fmt.Errorf("Passwords must match")
        }
        return nil
    },
}
```

---

### ValidationChain

Runs multiple validators in sequence.

```go
chain := ValidationChain{
    Validators: []Validator{
        RequiredValidator{},
        MinLengthValidator{MinLength: 8},
        PatternValidator{Pattern: hasNumberRegex},
    },
}

// Returns first error
err := chain.Validate("short")

// Returns all errors
errors := chain.ValidateAll("short")
```

---

## Form State Management

### FieldState

Represents the state of a single form field.

```go
type FieldState struct {
    Name       string       // Field identifier
    Value      interface{}  // Current value
    IsDirty    bool         // Has been modified
    IsTouched  bool         // Has been focused
    IsValid    bool         // Passes validation
    Errors     []string     // Validation errors
    Validators []Validator  // Field validators
}
```

**Methods:**

```go
Validate()                     // Run all validators
GetFirstError() string         // Get first error or empty string
MarkAsTouched()                // Mark as touched
MarkAsDirty()                  // Mark as dirty
SetValue(value interface{})    // Update value and mark dirty
```

**Example:**

```go
field := &FieldState{
    Name: "email",
    Validators: []Validator{
        RequiredValidator{},
        PatternValidator{Pattern: emailRegex},
    },
}

field.SetValue("invalid-email")
field.Validate()

if !field.IsValid {
    fmt.Println(field.GetFirstError())
}
```

---

### FormState

Manages state for an entire form.

```go
type FormState struct {
    Fields       map[string]*FieldState
    IsSubmitting bool
    IsValid      bool
    SubmitCount  int
}
```

**Methods:**

```go
NewFormState() *FormState                          // Create new form
RegisterField(name string, validators ...Validator) // Add field
SetFieldValue(name string, value interface{})      // Update field
ValidateField(name string) bool                    // Validate one field
ValidateAll() bool                                 // Validate all fields
GetFieldErrors(name string) []string               // Get field errors
GetAllErrors() map[string][]string                 // Get all errors
Reset()                                            // Reset form
```

**Complete Example:**

```go
// Create form
form := NewFormState()

// Register fields with validators
form.RegisterField("email",
    RequiredValidator{Message: "Email is required"},
    PatternValidator{Pattern: emailRegex, Message: "Invalid email"},
)

form.RegisterField("password",
    RequiredValidator{Message: "Password is required"},
    MinLengthValidator{MinLength: 8, Message: "Password too short"},
)

form.RegisterField("age",
    RequiredValidator{},
    RangeValidator{Min: 18, Max: 120, Message: "Invalid age"},
)

// Update field values
form.SetFieldValue("email", "user@example.com")
form.SetFieldValue("password", "securepass123")
form.SetFieldValue("age", 25)

// Validate on submit
if form.ValidateAll() {
    // All fields valid - submit form
    submitForm(form)
} else {
    // Show errors
    errors := form.GetAllErrors()
    for field, fieldErrors := range errors {
        fmt.Printf("%s: %v\n", field, fieldErrors)
    }
}

// Reset form
form.Reset()
```

---

## Builder Pattern

### ValidationPropsBuilder

Fluent API for building ValidationProps.

```go
validation := NewValidationProps().
    WithError("Email is required").
    ShowValidationUI(true).
    Build()
```

**Methods:**

```go
NewValidationProps() *ValidationPropsBuilder
WithState(state ValidationState) *ValidationPropsBuilder
WithError(message string) *ValidationPropsBuilder
WithSuccess(message string) *ValidationPropsBuilder
WithWarning(message string) *ValidationPropsBuilder
WithInfo(message string) *ValidationPropsBuilder
WithHelpText(text string) *ValidationPropsBuilder
ShowValidationUI(show bool) *ValidationPropsBuilder
Build() ValidationProps
```

**Example:**

```go
// Error state
errorValidation := NewValidationProps().
    WithError("Password must contain a number").
    Build()

// Success state
successValidation := NewValidationProps().
    WithSuccess("Email is available!").
    Build()

// Help text only
helpValidation := NewValidationProps().
    WithHelpText("Enter your email address").
    Build()
```

---

## Responsive Design

### ResponsiveValue

Different values at different breakpoints.

```go
type ResponsiveValue struct {
    Base string                 // Default value
    At   map[Breakpoint]string  // Values at breakpoints
}
```

**Breakpoints:**

```go
const (
    BreakpointSM  Breakpoint = "sm"  // 640px
    BreakpointMD  Breakpoint = "md"  // 768px
    BreakpointLG  Breakpoint = "lg"  // 1024px
    BreakpointXL  Breakpoint = "xl"  // 1280px
    BreakpointXXL Breakpoint = "2xl" // 1536px
)
```

**Example:**

```go
width := NewResponsiveValue("w-full").
    AtBreakpoint(BreakpointMD, "w-1/2").
    AtBreakpoint(BreakpointLG, "w-1/3")

classes := width.ToClasses()
// Result: "w-full md:w-1/2 lg:w-1/3"
```

---

### ResponsiveSize

Different sizes at different breakpoints.

```go
size := NewResponsiveSize(SizeSM).
    AtBreakpoint(BreakpointMD, SizeMD).
    AtBreakpoint(BreakpointLG, SizeLG)
```

---

## Theming System

### Theme

Complete color and styling theme.

```go
type Theme struct {
    Name         string
    Primary      ColorScheme
    Secondary    ColorScheme
    Accent       ColorScheme
    Neutral      ColorScheme
    BorderRadius string
    FontFamily   string
}
```

**Pre-configured Themes:**

```go
// Get a theme
theme := GetTheme("ocean")

// Available themes
DefaultThemes = map[string]Theme{
    "default": { /* ... */ },
    "ocean":   { /* ... */ },
    "forest":  { /* ... */ },
}
```

**Example:**

```go
theme := GetTheme("ocean")
primaryColor := theme.Primary    // ColorInfo
borderRadius := theme.BorderRadius // RoundedLG
```

---

## Utility Functions

### Class Management

```go
// Combine classes, removing duplicates
combined := CombineClasses("text-sm", "text-sm hover:text-lg", "text-sm")
// Result: "text-sm hover:text-lg"

// Join class arrays
classes := JoinClasses(
    []string{"border", "rounded"},
    []string{"focus:ring-2"},
)
// Result: "border rounded focus:ring-2"

// Conditional classes
class := ConditionalClass(isActive, "bg-blue-500")
// Returns "bg-blue-500" if isActive is true, "" otherwise

class := ConditionalClasses(isError, "text-red-500", "text-gray-500")
// Returns "text-red-500" if isError is true, "text-gray-500" otherwise
```

---

### ID Generation

```go
// Generate unique ID
id := GenerateID("input") // "input-1"

// Ensure ID exists or generate
id := EnsureID(existingID, "checkbox") // Returns existingID or generates new
```

---

### Color Configuration

```go
// Get state colors
colors := GetStateColors(StateError)
// Returns: StateColorConfig with border, background, text, ring, icon colors

borderClasses := colors.Border     // ["border-red-500", "dark:border-red-500"]
bgClasses := colors.Background     // ["bg-red-50", "dark:bg-gray-700"]
textClasses := colors.Text         // ["text-red-900", "dark:text-red-400"]
ringClasses := colors.Ring         // ["focus:ring-red-500", ...]

// Get feedback message colors
feedbackColors := GetFeedbackColors(StateSuccess)
// Returns: ["text-green-600", "dark:text-green-500"]
```

---

### Accessibility Helpers

```go
// Generate aria-describedby from multiple IDs
ariaDescribedBy := GenerateAriaDescribedBy("help-text", "error-msg", "")
// Result: "help-text error-msg" (empty strings are filtered)

// Get appropriate ARIA role
role := GetAriaRole(InputTypeSearch)
// Result: "searchbox"
```

---

### Border Radius

```go
// Get rounded class based on preferences
roundedClass := GetRoundedClass(true, SizeLG)
// Returns: "rounded-lg"

// Available constants
const (
    RoundedNone = "rounded-none"
    RoundedSM   = "rounded-sm"
    RoundedMD   = "rounded-md"
    RoundedLG   = "rounded-lg"
    RoundedXL   = "rounded-xl"
    RoundedFull = "rounded-full"
)
```

---

### Spacing Constants

```go
const (
    WrapperSpacingDefault = "mb-4"
    WrapperSpacingTight   = "mb-2"
    WrapperSpacingLoose   = "mb-6"
    
    LabelSpacingBottom = "mb-2"
    LabelSpacingRight  = "ms-2"
    LabelSpacingLeft   = "me-2"
    
    FeedbackSpacingTop = "mt-2"
    
    GroupSpacing       = "mb-5"
    GroupItemSpacing   = "space-y-2"
    GroupLegendSpacing = "mb-3"
)
```

---

### Transition Constants

```go
const (
    TransitionColors    = "transition-colors duration-200"
    TransitionAll       = "transition-all duration-200"
    TransitionOpacity   = "transition-opacity duration-200"
    TransitionTransform = "transition-transform duration-200"
)
```

---

## Best Practices

### 1. Use Composition

**Good:**

```go
type InputProps struct {
    BaseProps
    AccessibilityProps
    ValidationProps
    InteractionProps
    StyleProps
    // Input-specific fields
}
```

**Avoid:**

```go
type InputProps struct {
    // Duplicating all fields from shared structs
    ID string
    Name string
    // ... etc
}
```

---

### 2. Leverage Builders

**Good:**

```go
validation := NewValidationProps().
    WithError("Invalid email").
    Build()
```

**Avoid:**

```go
validation := ValidationProps{
    State: StateError,
    ErrorText: "Invalid email",
    ShowValidation: true,
}
```

---

### 3. Use Form State for Multi-Field Forms

**Good:**

```go
form := NewFormState()
form.RegisterField("email", RequiredValidator{}, EmailValidator{})
form.RegisterField("password", RequiredValidator{}, MinLengthValidator{MinLength: 8})

if form.ValidateAll() {
    // Submit
}
```

**Avoid:**

```go
// Manual validation for each field
emailValid := validateEmail(email)
passwordValid := validatePassword(password)
if emailValid && passwordValid {
    // Submit
}
```

---

### 4. Centralize Configuration

**Good:**

```go
colors := GetStateColors(state)
classes := append(classes, colors.Border...)
```

**Avoid:**

```go
// Hard-coding colors everywhere
if state == StateError {
    classes = append(classes, "border-red-500", "dark:border-red-500")
}
```

---

### 5. Use Responsive Values

**Good:**

```go
width := NewResponsiveValue("w-full").
    AtBreakpoint(BreakpointMD, "w-1/2")
classes := width.ToClasses()
```

**Avoid:**

```go
classes := "w-full md:w-1/2" // Hard-coded responsive classes
```

---

### 6. Validate Enums

**Good:**

```go
if IsValidSize(userSize) {
    // Use size
} else {
    // Fallback to default
    size = SizeMD
}
```

---

### 7. Implement Proper Accessibility

**Good:**

```go
a11y := AccessibilityProps{
    AriaLabel:       "Email address",
    AriaDescribedBy: GenerateAriaDescribedBy(helpTextID, errorID),
    AriaRequired:    true,
    AriaInvalid:     hasError,
}
```

---

### 8. Use Validation Chains

**Good:**

```go
chain := ValidationChain{
    Validators: []Validator{
        RequiredValidator{},
        MinLengthValidator{MinLength: 8},
        PatternValidator{Pattern: hasNumberRegex},
    },
}
```

---

### 9. Leverage Helper Methods

**Good:**

```go
if validation.HasFeedback() {
    message, state := validation.GetFeedbackMessage()
    // Display message
}
```

---

### 10. Keep Components Testable

**Good:**

```go
// Test validator independently
validator := MinLengthValidator{MinLength: 8}
err := validator.Validate("short")
assert.NotNil(t, err)
```

---

## Migration Guide

### From Old Checkbox to New Shared System

**Old:**

```go
type CheckboxProps struct {
    ID          string
    Name        string
    Class       string
    Disabled    bool
    Required    bool
    Label       string
    HelpText    string
    ErrorText   string
    Size        CheckboxSize
    State       CheckboxState
    OnChange    string
}
```

**New:**

```go
type CheckboxProps struct {
    BaseProps              // ID, Name, Class
    InteractionProps       // Disabled, Required
    ValidationProps        // HelpText, ErrorText, State
    AlpineEventHandlers    // OnChange
    LabelProps             // Label
    
    // Checkbox-specific
    Size    Size
    Checked bool
}
```

**Benefits:**
- Reusable across all components
- Consistent API
- Built-in validation
- Better type safety
- Rich helper methods

---

## Complete Example: Login Form

```go
// Create form state
loginForm := NewFormState()

// Register email field
loginForm.RegisterField("email",
    RequiredValidator{Message: "Email is required"},
    PatternValidator{
        Pattern: emailRegex,
        Message: "Please enter a valid email",
    },
)

// Register password field
loginForm.RegisterField("password",
    RequiredValidator{Message: "Password is required"},
    MinLengthValidator{
        MinLength: 8,
        Message:   "Password must be at least 8 characters",
    },
)

// Handle form submission
func handleSubmit(email, password string) {
    loginForm.SetFieldValue("email", email)
    loginForm.SetFieldValue("password", password)
    
    if loginForm.ValidateAll() {
        // All fields valid - submit
        submitLogin(email, password)
    } else {
        // Display errors
        if errors := loginForm.GetFieldErrors("email"); len(errors) > 0 {
            showEmailError(errors[0])
        }
        if errors := loginForm.GetFieldErrors("password"); len(errors) > 0 {
            showPasswordError(errors[0])
        }
    }
}
```

---

## FAQ

### Q: Can I use these types with other CSS frameworks?

**A:** Yes! While optimized for Tailwind CSS, you can adapt the class generation functions to output classes for Bootstrap, Bulma, or custom CSS.

### Q: How do I add a custom validator?

**A:** Implement the `Validator` interface:

```go
type MyValidator struct {
    Message string
}

func (v MyValidator) Validate(value interface{}) error {
    // Your validation logic
}

func (v MyValidator) ErrorMessage() string {
    return v.Message
}
```

### Q: Can I use this with HTMX instead of Alpine.js?

**A:** Yes! The event handler strings can be HTMX attributes:

```go
handlers := AlpineEventHandlers{
    OnChange: "hx-post='/validate'",
}
```

### Q: How do I customize color schemes?

**A:** Modify the `stateColorSchemes` map or create custom themes:

```go
customTheme := Theme{
    Name:    "custom",
    Primary: ColorPrimary,
    // ... customize colors
}
```

### Q: Is this production-ready?

**A:** The structure is production-ready. Some validators (like PatternValidator) use simplified logic for demonstration. Replace with proper regex/validation libraries for production use.

---

## Version History

- **v1.0.0** - Initial release with core types and validators
- **v1.1.0** - Added form state management
- **v1.2.0** - Added responsive design helpers
- **v1.3.0** - Added theming system and icon support

---

## Contributing

When adding new shared types:

1. Document all fields with clear comments
2. Provide helper methods for common operations
3. Add validation functions where applicable
4. Include usage examples in this documentation
5. Ensure type safety with enums instead of strings
6. Follow composition over inheritance pattern

---

## License

[Your License Here]

## Support

For issues or questions:
- GitHub Issues: [Your Repo]
- Documentation: [Your Docs Site]
- Email: [Your Email]

---

## Advanced Usage Patterns

### Custom Component Creation

Here's how to create a new component using shared types:

```go
// 1. Define component-specific props
type SelectProps struct {
    // Embed shared structs
    BaseProps
    AccessibilityProps
    ValidationProps
    InteractionProps
    StyleProps
    LabelProps
    AlpineEventHandlers
    
    // Select-specific fields
    Options      []SelectOption
    Value        string
    Multiple     bool
    Searchable   bool
    Clearable    bool
    Placeholder  string
    MaxHeight    string
}

type SelectOption struct {
    Value    string
    Label    string
    Disabled bool
    Group    string
}

// 2. Create constructor with defaults
func NewSelect(id, label string) SelectProps {
    return SelectProps{
        BaseProps: BaseProps{
            ID: id,
        },
        LabelProps: LabelProps{
            Label:         label,
            LabelPosition: LabelTop,
        },
        StyleProps: StyleProps{
            Size:    SizeMD,
            Variant: VariantDefault,
        },
        ValidationProps: ValidationProps{
            State: StateDefault,
        },
        Placeholder: "Select an option...",
        MaxHeight:   "max-h-60",
    }
}

// 3. Add builder methods
func (s SelectProps) WithOptions(options []SelectOption) SelectProps {
    s.Options = options
    return s
}

func (s SelectProps) AsMultiple() SelectProps {
    s.Multiple = true
    return s
}

func (s SelectProps) AsSearchable() SelectProps {
    s.Searchable = true
    return s
}

// 4. Usage
select := NewSelect("country", "Country").
    WithOptions(countryOptions).
    AsSearchable().
    AsRequired()
```

---

### Complex Validation Scenarios

#### Cross-Field Validation

```go
// Password confirmation validator
type PasswordConfirmValidator struct {
    PasswordFieldName string
    Message           string
}

func (v PasswordConfirmValidator) Validate(value interface{}) error {
    // In real implementation, access form state
    confirmPassword := value.(string)
    originalPassword := getFieldValue(v.PasswordFieldName)
    
    if confirmPassword != originalPassword {
        return fmt.Errorf(v.ErrorMessage())
    }
    return nil
}

func (v PasswordConfirmValidator) ErrorMessage() string {
    if v.Message != "" {
        return v.Message
    }
    return "Passwords do not match"
}

// Usage in form
form := NewFormState()
form.RegisterField("password", 
    RequiredValidator{},
    MinLengthValidator{MinLength: 8},
)
form.RegisterField("confirmPassword",
    RequiredValidator{},
    PasswordConfirmValidator{PasswordFieldName: "password"},
)
```

---

#### Async Validation

```go
// Email uniqueness validator (checks against API)
type EmailUniqueValidator struct {
    APIEndpoint string
    Message     string
}

func (v EmailUniqueValidator) Validate(value interface{}) error {
    email := value.(string)
    
    // In production, make actual API call
    exists := checkEmailExists(v.APIEndpoint, email)
    
    if exists {
        return fmt.Errorf(v.ErrorMessage())
    }
    return nil
}

func (v EmailUniqueValidator) ErrorMessage() string {
    if v.Message != "" {
        return v.Message
    }
    return "Email is already registered"
}

// Usage with debouncing in Alpine.js
events := AlpineEventHandlers{
    OnInput: "debounce(() => validateEmailUnique($el.value), 500)",
}
```

---

#### Conditional Validation

```go
// Validator that only runs if condition is met
type ConditionalValidator struct {
    Condition func() bool
    Validator Validator
}

func (v ConditionalValidator) Validate(value interface{}) error {
    if v.Condition() {
        return v.Validator.Validate(value)
    }
    return nil
}

func (v ConditionalValidator) ErrorMessage() string {
    return v.Validator.ErrorMessage()
}

// Usage: Only validate credit card if payment method is "card"
form.RegisterField("creditCard",
    ConditionalValidator{
        Condition: func() bool { 
            return paymentMethod == "card" 
        },
        Validator: PatternValidator{
            Pattern: creditCardRegex,
            Message: "Invalid credit card number",
        },
    },
)
```

---

### Dynamic Form Generation

```go
// Form configuration from JSON/database
type FormConfig struct {
    Fields []FieldConfig
}

type FieldConfig struct {
    Name       string
    Type       string
    Label      string
    Required   bool
    Validators []string
    Options    []string
}

func BuildFormFromConfig(config FormConfig) *FormState {
    form := NewFormState()
    
    for _, fieldConfig := range config.Fields {
        validators := buildValidators(fieldConfig)
        form.RegisterField(fieldConfig.Name, validators...)
    }
    
    return form
}

func buildValidators(config FieldConfig) []Validator {
    var validators []Validator
    
    if config.Required {
        validators = append(validators, RequiredValidator{})
    }
    
    for _, validatorType := range config.Validators {
        switch validatorType {
        case "email":
            validators = append(validators, PatternValidator{
                Pattern: emailRegex,
                Message: "Invalid email",
            })
        case "minLength:8":
            validators = append(validators, MinLengthValidator{
                MinLength: 8,
            })
        // ... more validators
        }
    }
    
    return validators
}

// Usage
jsonConfig := `{
    "fields": [
        {
            "name": "email",
            "type": "email",
            "label": "Email Address",
            "required": true,
            "validators": ["email"]
        },
        {
            "name": "age",
            "type": "number",
            "label": "Age",
            "required": true,
            "validators": ["range:18:120"]
        }
    ]
}`

config := parseFormConfig(jsonConfig)
form := BuildFormFromConfig(config)
```

---

### State Persistence

```go
// Save form state to localStorage (via Alpine.js)
type FormPersistence struct {
    StorageKey string
}

func (fp FormPersistence) SaveState(form *FormState) string {
    // Generate Alpine.js x-data
    data := make(map[string]interface{})
    
    for name, field := range form.Fields {
        data[name] = field.Value
    }
    
    // Return Alpine.js expression
    return fmt.Sprintf(
        `x-data="{ formData: %s }" x-init="$watch('formData', val => localStorage.setItem('%s', JSON.stringify(val)))"`,
        toJSON(data),
        fp.StorageKey,
    )
}

func (fp FormPersistence) LoadState() string {
    return fmt.Sprintf(
        `x-init="formData = JSON.parse(localStorage.getItem('%s') || '{}')"`,
        fp.StorageKey,
    )
}

// Usage in template
persistence := FormPersistence{StorageKey: "registration-form"}
alpineData := persistence.SaveState(form)
```

---

### Multi-Step Forms

```go
type MultiStepForm struct {
    Steps       []FormStep
    CurrentStep int
    FormState   *FormState
}

type FormStep struct {
    Name        string
    Title       string
    Description string
    Fields      []string // Field names in this step
}

func NewMultiStepForm(steps []FormStep) *MultiStepForm {
    return &MultiStepForm{
        Steps:       steps,
        CurrentStep: 0,
        FormState:   NewFormState(),
    }
}

func (msf *MultiStepForm) ValidateCurrentStep() bool {
    currentStep := msf.Steps[msf.CurrentStep]
    
    for _, fieldName := range currentStep.Fields {
        if !msf.FormState.ValidateField(fieldName) {
            return false
        }
    }
    
    return true
}

func (msf *MultiStepForm) NextStep() bool {
    if !msf.ValidateCurrentStep() {
        return false
    }
    
    if msf.CurrentStep < len(msf.Steps)-1 {
        msf.CurrentStep++
        return true
    }
    
    return false
}

func (msf *MultiStepForm) PrevStep() {
    if msf.CurrentStep > 0 {
        msf.CurrentStep--
    }
}

func (msf *MultiStepForm) CanSubmit() bool {
    return msf.CurrentStep == len(msf.Steps)-1 && 
           msf.ValidateCurrentStep()
}

// Usage
steps := []FormStep{
    {
        Name:   "personal",
        Title:  "Personal Information",
        Fields: []string{"firstName", "lastName", "email"},
    },
    {
        Name:   "address",
        Title:  "Address",
        Fields: []string{"street", "city", "zipCode"},
    },
    {
        Name:   "preferences",
        Title:  "Preferences",
        Fields: []string{"newsletter", "notifications"},
    },
}

multiForm := NewMultiStepForm(steps)

// Register all fields
multiForm.FormState.RegisterField("firstName", RequiredValidator{})
multiForm.FormState.RegisterField("lastName", RequiredValidator{})
// ... etc

// Navigation
if multiForm.NextStep() {
    // Move to next step
} else {
    // Show validation errors for current step
}
```

---

### Internationalization (i18n)

```go
// i18n support for validation messages
type i18nValidator struct {
    Validator Validator
    Locale    string
}

var translations = map[string]map[string]string{
    "en": {
        "required":   "This field is required",
        "email":      "Invalid email address",
        "minLength":  "Must be at least %d characters",
    },
    "es": {
        "required":   "Este campo es obligatorio",
        "email":      "Dirección de correo electrónico no válida",
        "minLength":  "Debe tener al menos %d caracteres",
    },
    "fr": {
        "required":   "Ce champ est requis",
        "email":      "Adresse e-mail invalide",
        "minLength":  "Doit contenir au moins %d caractères",
    },
}

func (v i18nValidator) Validate(value interface{}) error {
    return v.Validator.Validate(value)
}

func (v i18nValidator) ErrorMessage() string {
    // Get base message key
    baseMessage := v.Validator.ErrorMessage()
    
    // Look up translation
    if localeMap, ok := translations[v.Locale]; ok {
        if translated, ok := localeMap[baseMessage]; ok {
            return translated
        }
    }
    
    return baseMessage
}

// Usage
form := NewFormState()
form.RegisterField("email",
    i18nValidator{
        Validator: RequiredValidator{Message: "required"},
        Locale:    "es",
    },
    i18nValidator{
        Validator: PatternValidator{Pattern: emailRegex, Message: "email"},
        Locale:    "es",
    },
)
```

---

### Form Analytics & Tracking

```go
type FormAnalytics struct {
    FormID           string
    StartTime        time.Time
    FieldInteractions map[string]int
    ValidationErrors  map[string]int
    SubmitAttempts   int
}

func NewFormAnalytics(formID string) *FormAnalytics {
    return &FormAnalytics{
        FormID:            formID,
        StartTime:         time.Now(),
        FieldInteractions: make(map[string]int),
        ValidationErrors:  make(map[string]int),
    }
}

func (fa *FormAnalytics) TrackFieldInteraction(fieldName string) {
    fa.FieldInteractions[fieldName]++
}

func (fa *FormAnalytics) TrackValidationError(fieldName string) {
    fa.ValidationErrors[fieldName]++
}

func (fa *FormAnalytics) TrackSubmit() {
    fa.SubmitAttempts++
}

func (fa *FormAnalytics) GetTimeToComplete() time.Duration {
    return time.Since(fa.StartTime)
}

func (fa *FormAnalytics) GetReport() map[string]interface{} {
    return map[string]interface{}{
        "formId":           fa.FormID,
        "timeToComplete":   fa.GetTimeToComplete().Seconds(),
        "interactions":     fa.FieldInteractions,
        "validationErrors": fa.ValidationErrors,
        "submitAttempts":   fa.SubmitAttempts,
    }
}

// Usage with Alpine.js
events := AlpineEventHandlers{
    OnFocus: "trackFieldInteraction($el.name)",
    OnBlur:  "trackFieldBlur($el.name)",
    OnInput: "trackFieldInput($el.name)",
}
```

---

### Accessibility Best Practices Implementation

```go
// Helper to ensure complete accessibility
type AccessibilityChecker struct {
    Issues []string
}

func (ac *AccessibilityChecker) CheckComponent(props interface{}) {
    ac.Issues = nil
    
    // Type assert to get accessibility props
    if hasA11y, ok := props.(interface{ GetA11y() AccessibilityProps }); ok {
        a11y := hasA11y.GetA11y()
        
        // Check for label
        if hasLabel, ok := props.(interface{ GetLabel() string }); ok {
            label := hasLabel.GetLabel()
            if label == "" && a11y.AriaLabel == "" && a11y.AriaLabelledBy == "" {
                ac.Issues = append(ac.Issues, 
                    "Component has no accessible label (label, aria-label, or aria-labelledby required)")
            }
        }
        
        // Check required fields
        if hasRequired, ok := props.(interface{ IsRequired() bool }); ok {
            if hasRequired.IsRequired() && !a11y.AriaRequired {
                ac.Issues = append(ac.Issues,
                    "Required field should have aria-required='true'")
            }
        }
        
        // Check error states
        if hasError, ok := props.(interface{ HasError() bool }); ok {
            if hasError.HasError() && !a11y.AriaInvalid {
                ac.Issues = append(ac.Issues,
                    "Invalid field should have aria-invalid='true'")
            }
        }
    }
}

func (ac *AccessibilityChecker) HasIssues() bool {
    return len(ac.Issues) > 0
}

func (ac *AccessibilityChecker) GetIssues() []string {
    return ac.Issues
}

// Usage during development
checker := &AccessibilityChecker{}
checker.CheckComponent(inputProps)
if checker.HasIssues() {
    for _, issue := range checker.GetIssues() {
        log.Println("A11Y WARNING:", issue)
    }
}
```

---

### Performance Optimization Patterns

```go
// Debounced validation for expensive validators
type DebouncedValidator struct {
    Validator Validator
    Delay     time.Duration
    timer     *time.Timer
    lastValue interface{}
    callback  func(error)
}

func NewDebouncedValidator(validator Validator, delay time.Duration) *DebouncedValidator {
    return &DebouncedValidator{
        Validator: validator,
        Delay:     delay,
    }
}

func (dv *DebouncedValidator) Validate(value interface{}) error {
    dv.lastValue = value
    
    if dv.timer != nil {
        dv.timer.Stop()
    }
    
    dv.timer = time.AfterFunc(dv.Delay, func() {
        err := dv.Validator.Validate(dv.lastValue)
        if dv.callback != nil {
            dv.callback(err)
        }
    })
    
    return nil // Immediate return, actual validation happens async
}

func (dv *DebouncedValidator) ErrorMessage() string {
    return dv.Validator.ErrorMessage()
}

// Usage
emailValidator := NewDebouncedValidator(
    EmailUniqueValidator{APIEndpoint: "/api/check-email"},
    500*time.Millisecond,
)

// Alpine.js integration
events := AlpineEventHandlers{
    OnInput: "validateEmailDebounced($el.value)",
}
```

---

### Testing Helpers

```go
// Test utilities for validators
type ValidatorTestCase struct {
    Name        string
    Validator   Validator
    ValidInputs []interface{}
    InvalidInputs []interface{}
}

func RunValidatorTests(t *testing.T, testCases []ValidatorTestCase) {
    for _, tc := range testCases {
        t.Run(tc.Name, func(t *testing.T) {
            // Test valid inputs
            for _, input := range tc.ValidInputs {
                err := tc.Validator.Validate(input)
                if err != nil {
                    t.Errorf("Expected valid input %v to pass, got error: %v", 
                        input, err)
                }
            }
            
            // Test invalid inputs
            for _, input := range tc.InvalidInputs {
                err := tc.Validator.Validate(input)
                if err == nil {
                    t.Errorf("Expected invalid input %v to fail, but it passed", 
                        input)
                }
            }
        })
    }
}

// Usage in tests
func TestValidators(t *testing.T) {
    testCases := []ValidatorTestCase{
        {
            Name:      "Email Validator",
            Validator: PatternValidator{Pattern: emailRegex},
            ValidInputs: []interface{}{
                "user@example.com",
                "name.surname@company.co.uk",
            },
            InvalidInputs: []interface{}{
                "notanemail",
                "@example.com",
                "user@",
            },
        },
        {
            Name:      "MinLength Validator",
            Validator: MinLengthValidator{MinLength: 8},
            ValidInputs: []interface{}{
                "password123",
                "verylongpassword",
            },
            InvalidInputs: []interface{}{
                "short",
                "1234567",
            },
        },
    }
    
    RunValidatorTests(t, testCases)
}

// Mock form state for testing
type MockFormState struct {
    *FormState
    ValidationCalls int
}

func NewMockFormState() *MockFormState {
    return &MockFormState{
        FormState: NewFormState(),
    }
}

func (mfs *MockFormState) ValidateAll() bool {
    mfs.ValidationCalls++
    return mfs.FormState.ValidateAll()
}
```

---

### Integration Examples

#### React Integration (via templ)

```go
// Generate props for React components
func (props InputProps) ToReactProps() string {
    propsMap := map[string]interface{}{
        "id":          props.ID,
        "name":        props.Name,
        "value":       props.Value,
        "disabled":    props.Disabled,
        "required":    props.Required,
        "placeholder": props.Placeholder,
    }
    
    // Add validation props
    if props.ShowValidation {
        propsMap["error"] = props.ErrorText
        propsMap["success"] = props.SuccessText
    }
    
    return toJSON(propsMap)
}

// Usage in template
templ InputWrapper(props InputProps) {
    <div 
        data-react-component="Input"
        data-props={ props.ToReactProps() }
    ></div>
}
```

---

#### HTMX Integration

```go
// Generate HTMX attributes for validation
func (props ValidationProps) ToHTMXAttrs(endpoint string) map[string]string {
    return map[string]string{
        "hx-post":    endpoint,
        "hx-trigger": "change delay:500ms",
        "hx-target":  "#validation-message",
        "hx-swap":    "innerHTML",
    }
}

// Usage
htmxAttrs := validation.ToHTMXAttrs("/api/validate")
```

---

## Performance Considerations

### 1. Validator Caching

```go
// Cache compiled regex patterns
var regexCache = make(map[string]*regexp.Regexp)
var regexCacheMutex sync.RWMutex

func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
    regexCacheMutex.RLock()
    if cached, ok := regexCache[pattern]; ok {
        regexCacheMutex.RUnlock()
        return cached, nil
    }
    regexCacheMutex.RUnlock()
    
    compiled, err := regexp.Compile(pattern)
    if err != nil {
        return nil, err
    }
    
    regexCacheMutex.Lock()
    regexCache[pattern] = compiled
    regexCacheMutex.Unlock()
    
    return compiled, nil
}
```

---

### 2. Lazy Validation

```go
// Only validate touched fields
func (fs *FormState) ValidateTouchedFields() bool {
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
```

---

### 3. Batch Validation Updates

```go
// Batch multiple field updates
func (fs *FormState) SetFieldValues(values map[string]interface{}) {
    for name, value := range values {
        if field, ok := fs.Fields[name]; ok {
            field.Value = value
            field.IsDirty = true
        }
    }
}

// Then validate once
fs.ValidateAll()
```

---

## Troubleshooting

### Common Issues

#### Issue: Validation not triggering

**Problem:**
```go
form.SetFieldValue("email", "test@example.com")
// Validation doesn't run automatically
```

**Solution:**
```go
form.SetFieldValue("email", "test@example.com")
form.ValidateField("email") // Explicitly trigger validation
```

---

#### Issue: Alpine.js events not working

**Problem:**
```go
OnChange: "@change='validate()'" // Wrong syntax
```

**Solution:**
```go
OnChange: "validate()" // Correct - no @change prefix needed
```

---

#### Issue: Circular dependencies in validation

**Problem:**
```go
// Field A depends on Field B, Field B depends on Field A
```

**Solution:**
```go
// Use validation order or break the cycle
type OrderedValidator struct {
    ValidationOrder []string
}
```

---

## Glossary

- **Validator**: Interface that validates a value
- **ValidationState**: Current validation status (error, success, etc.)
- **FieldState**: Complete state of a form field
- **FormState**: Manager for entire form with multiple fields
- **Builder Pattern**: Fluent API for constructing objects
- **Composition**: Combining multiple structs via embedding
- **Breakpoint**: Responsive design screen size threshold
- **ARIA**: Accessible Rich Internet Applications attributes

---

## Quick Reference Card

```go
// Create Form
form := NewFormState()

// Register Field
form.RegisterField("email", RequiredValidator{}, EmailValidator{})

// Update Value
form.SetFieldValue("email", "user@example.com")

// Validate
form.ValidateAll()              // All fields
form.ValidateField("email")     // Single field

// Get Errors
form.GetFieldErrors("email")    // Single field
form.GetAllErrors()             // All fields

// Check State
field := form.Fields["email"]
field.IsDirty                   // Modified?
field.IsTouched                 // Focused?
field.IsValid                   // Valid?

// Builder Pattern
validation := NewValidationProps().
    WithError("Invalid").
    Build()

// Responsive
width := NewResponsiveValue("w-full").
    AtBreakpoint(BreakpointMD, "w-1/2").
    ToClasses()

// Colors
colors := GetStateColors(StateError)
classes := colors.Border // ["border-red-500", ...]

// Accessibility
ariaDescBy := GenerateAriaDescribedBy(helpID, errorID)
```
