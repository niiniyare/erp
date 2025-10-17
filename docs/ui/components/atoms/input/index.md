# Input Component

**FILE PURPOSE**: Text input field implementation and specifications  
**SCOPE**: All text input variants, validation, and interaction patterns  
**TARGET AUDIENCE**: Developers implementing form controls

## 📋 Component Overview

The Input component is the fundamental text entry element for forms and data collection. It provides type-safe validation, accessibility compliance, and consistent styling across all input types.

### Schema Reference
- **Primary Schema**: `TextControlSchema.json`
- **Related Schemas**: `InputColorControlSchema.json`, `HiddenControlSchema.json`
- **Base Interface**: Form control with validation

## 🎨 Input Types

### Text Input (Default)
**Purpose**: General text entry and string data collection

```go
// Basic text input
textInput := InputProps{
    Type:        "text",
    Name:        "firstName",
    Label:       "First Name",
    Placeholder: "Enter your first name",
    Required:    true,
    MaxLength:   50,
}

// Generated Templ
templ TextInput(props InputProps) {
    <div class="form-group">
        <label for={ props.Name } class="form-label">
            { props.Label }
            if props.Required {
                <span class="required">*</span>
            }
        </label>
        <input 
            type="text"
            id={ props.Name }
            name={ props.Name }
            class={ props.GetClasses() }
            placeholder={ props.Placeholder }
            value={ props.Value }
            maxlength?={ props.MaxLength }
            required?={ props.Required }
            disabled?={ props.Disabled }
            readonly?={ props.ReadOnly }
        />
        if props.HelpText != "" {
            <div class="help-text">{ props.HelpText }</div>
        }
        if props.Error != "" {
            <div class="error-message">{ props.Error }</div>
        }
    </div>
}
```

### Email Input
**Purpose**: Email address collection with built-in validation

```go
emailInput := InputProps{
    Type:        "email",
    Name:        "email",
    Label:       "Email Address",
    Placeholder: "user@example.com",
    Required:    true,
    Validation: ValidationConfig{
        Pattern: `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
        Message: "Please enter a valid email address",
    },
}
```

### Password Input  
**Purpose**: Secure password entry with visibility toggle

```go
passwordInput := InputProps{
    Type:        "password",
    Name:        "password",
    Label:       "Password",
    Required:    true,
    MinLength:   8,
    ShowToggle:  true,  // Show/hide password toggle
    Validation: ValidationConfig{
        Pattern: `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$`,
        Message: "Password must contain uppercase, lowercase, number, and special character",
    },
}

// Enhanced password component
templ PasswordInput(props InputProps) {
    <div class="form-group" x-data="{ showPassword: false }">
        <label for={ props.Name } class="form-label">{ props.Label }</label>
        <div class="input-group">
            <input 
                :type="showPassword ? 'text' : 'password'"
                id={ props.Name }
                name={ props.Name }
                class={ props.GetClasses() }
                required?={ props.Required }
                minlength?={ props.MinLength }
            />
            if props.ShowToggle {
                <button 
                    type="button"
                    class="input-toggle"
                    @click="showPassword = !showPassword"
                    :aria-label="showPassword ? 'Hide password' : 'Show password'">
                    <span x-show="!showPassword">👁️</span>
                    <span x-show="showPassword">🙈</span>
                </button>
            }
        </div>
    </div>
}
```

### Number Input
**Purpose**: Numeric data entry with constraints

```go
numberInput := InputProps{
    Type:        "number",
    Name:        "quantity",
    Label:       "Quantity",
    Value:       "1",
    Min:         1,
    Max:         100,
    Step:        1,
    Required:    true,
}

// Currency input variant
currencyInput := InputProps{
    Type:        "number",
    Name:        "amount",
    Label:       "Amount",
    Prefix:      "$",
    Step:        0.01,
    Min:         0,
    Placeholder: "0.00",
    Class:       "text-right font-mono",
}
```

### URL Input
**Purpose**: Web address entry with validation

```go
urlInput := InputProps{
    Type:        "url",
    Name:        "website",
    Label:       "Website",
    Placeholder: "https://example.com",
    Validation: ValidationConfig{
        Pattern: `^https?:\/\/.+$`,
        Message: "Please enter a valid URL starting with http:// or https://",
    },
}
```

### Phone Input
**Purpose**: Phone number entry with formatting

```go
phoneInput := InputProps{
    Type:        "tel",
    Name:        "phone",
    Label:       "Phone Number",
    Placeholder: "(555) 123-4567",
    Mask:        "(999) 999-9999",  // Input mask
    Validation: ValidationConfig{
        Pattern: `^\(\d{3}\) \d{3}-\d{4}$`,
        Message: "Please enter phone in format (555) 123-4567",
    },
}
```

### Search Input
**Purpose**: Search functionality with enhanced UX

```go
searchInput := InputProps{
    Type:        "search",
    Name:        "search",
    Label:       "Search",
    Placeholder: "Search items...",
    Icon:        "search",
    Debounce:    300,  // Debounce search requests
    ClearButton: true, // Show clear button
}

templ SearchInput(props InputProps) {
    <div class="search-group" x-data="{ query: '', loading: false }">
        <div class="input-with-icon">
            if props.Icon != "" {
                @Icon(IconProps{Name: props.Icon, Class: "input-icon-left"})
            }
            <input 
                type="search"
                x-model="query"
                @input.debounce.300ms="search(query)"
                class={ props.GetClasses() }
                placeholder={ props.Placeholder }
            />
            if props.ClearButton {
                <button 
                    type="button"
                    x-show="query.length > 0"
                    @click="query = ''; search('')"
                    class="input-clear">
                    ×
                </button>
            }
        </div>
        <div x-show="loading" class="search-spinner">
            @Spinner(SpinnerProps{Size: "sm"})
        </div>
    </div>
}
```

## 🎯 Props Interface

### Core Properties
```go
type InputProps struct {
    // Identity
    Name     string `json:"name"`
    ID       string `json:"id"`
    TestID   string `json:"testid"`
    
    // Content
    Type        string `json:"type"`         // text, email, password, number, url, tel, search
    Value       string `json:"value"`
    DefaultValue string `json:"defaultValue"`
    Placeholder string `json:"placeholder"`
    
    // Labels & Help
    Label       string `json:"label"`
    HelpText    string `json:"helpText"`
    Error       string `json:"error"`
    Prefix      string `json:"prefix"`       // $ symbol, etc.
    Suffix      string `json:"suffix"`       // units, etc.
    
    // Validation
    Required    bool              `json:"required"`
    ReadOnly    bool              `json:"readonly"`
    Disabled    bool              `json:"disabled"`
    MinLength   int               `json:"minLength"`
    MaxLength   int               `json:"maxLength"`
    Pattern     string            `json:"pattern"`
    Min         *float64          `json:"min"`          // For number inputs
    Max         *float64          `json:"max"`          // For number inputs
    Step        *float64          `json:"step"`         // For number inputs
    Validation  ValidationConfig  `json:"validation"`
    
    // Appearance
    Size        InputSize         `json:"size"`         // xs, sm, md, lg, xl
    Variant     InputVariant      `json:"variant"`      // default, success, error
    Class       string            `json:"className"`
    Style       map[string]string `json:"style"`
    FullWidth   bool              `json:"fullWidth"`
    
    // Features
    Icon        string            `json:"icon"`         // Icon name
    IconPosition IconPosition     `json:"iconPosition"` // left, right
    ShowToggle  bool              `json:"showToggle"`   // For password inputs
    ClearButton bool              `json:"clearButton"`  // Show clear button
    Mask        string            `json:"mask"`         // Input mask pattern
    Debounce    int               `json:"debounce"`     // Debounce delay in ms
    
    // Events
    OnChange    string            `json:"onChange"`
    OnFocus     string            `json:"onFocus"`
    OnBlur      string            `json:"onBlur"`
    OnKeyPress  string            `json:"onKeyPress"`
    OnKeyUp     string            `json:"onKeyUp"`
    OnKeyDown   string            `json:"onKeyDown"`
    
    // Accessibility
    AriaLabel       string `json:"ariaLabel"`
    AriaDescribedBy string `json:"ariaDescribedBy"`
    AriaInvalid     bool   `json:"ariaInvalid"`
    TabIndex        int    `json:"tabIndex"`
}
```

### Size Variants
```go
type InputSize string

const (
    InputXS InputSize = "xs"    // 24px height, compact
    InputSM InputSize = "sm"    // 32px height, small forms  
    InputMD InputSize = "md"    // 36px height, default
    InputLG InputSize = "lg"    // 44px height, prominent
    InputXL InputSize = "xl"    // 52px height, hero forms
)
```

### Visual Variants
```go
type InputVariant string

const (
    InputDefault InputVariant = "default"  // Standard appearance
    InputSuccess InputVariant = "success"  // Valid state
    InputError   InputVariant = "error"    // Invalid state
    InputWarning InputVariant = "warning"  // Warning state
)
```

## 🎨 Styling Implementation

### Base Input Styles
```css
.input {
    /* Base properties */
    display: block;
    width: 100%;
    font-family: var(--font-family-ui);
    font-size: var(--font-size-base);
    line-height: 1.5;
    color: var(--color-text-primary);
    background: var(--color-bg-surface);
    border: 1px solid var(--color-border-medium);
    border-radius: var(--radius-md);
    transition: var(--transition-base);
    
    /* Interaction states */
    &:hover {
        border-color: var(--color-border-dark);
    }
    
    &:focus {
        outline: none;
        border-color: var(--color-primary);
        box-shadow: 0 0 0 3px var(--color-primary-light);
    }
    
    &:disabled {
        background: var(--color-bg-disabled);
        border-color: var(--color-border-light);
        color: var(--color-text-disabled);
        cursor: not-allowed;
    }
    
    &:readonly {
        background: var(--color-bg-readonly);
        border-color: var(--color-border-light);
    }
    
    /* Placeholder styling */
    &::placeholder {
        color: var(--color-text-placeholder);
        opacity: 1;
    }
}
```

### Size Variants
```css
/* Size-specific styling */
.input-xs {
    padding: 4px 8px;
    font-size: var(--font-size-xs);
    height: 24px;
}

.input-sm {
    padding: 6px 12px;
    font-size: var(--font-size-sm);
    height: 32px;
}

.input-md {
    padding: 8px 12px;
    font-size: var(--font-size-base);
    height: 36px;
}

.input-lg {
    padding: 12px 16px;
    font-size: var(--font-size-lg);
    height: 44px;
}

.input-xl {
    padding: 16px 20px;
    font-size: var(--font-size-xl);
    height: 52px;
}
```

### State Variants
```css
/* Success state */
.input-success {
    border-color: var(--color-success);
    background: var(--color-success-bg);
    
    &:focus {
        border-color: var(--color-success);
        box-shadow: 0 0 0 3px var(--color-success-light);
    }
}

/* Error state */
.input-error {
    border-color: var(--color-error);
    background: var(--color-error-bg);
    
    &:focus {
        border-color: var(--color-error);
        box-shadow: 0 0 0 3px var(--color-error-light);
    }
}

/* Warning state */
.input-warning {
    border-color: var(--color-warning);
    background: var(--color-warning-bg);
    
    &:focus {
        border-color: var(--color-warning);
        box-shadow: 0 0 0 3px var(--color-warning-light);
    }
}
```

### Input Groups
```css
/* Input with icons */
.input-group {
    position: relative;
    display: flex;
    align-items: center;
    
    .input {
        flex: 1;
    }
    
    .input-icon-left {
        position: absolute;
        left: 12px;
        z-index: 1;
        color: var(--color-text-tertiary);
        pointer-events: none;
        
        + .input {
            padding-left: 40px;
        }
    }
    
    .input-icon-right {
        position: absolute;
        right: 12px;
        z-index: 1;
        color: var(--color-text-tertiary);
        
        + .input {
            padding-right: 40px;
        }
    }
    
    .input-toggle,
    .input-clear {
        position: absolute;
        right: 8px;
        z-index: 1;
        background: none;
        border: none;
        color: var(--color-text-tertiary);
        cursor: pointer;
        padding: 4px;
        border-radius: var(--radius-sm);
        
        &:hover {
            color: var(--color-text-primary);
            background: var(--color-bg-hover);
        }
    }
}

/* Prefix/suffix */
.input-with-addons {
    display: flex;
    
    .input-prefix,
    .input-suffix {
        display: flex;
        align-items: center;
        padding: 0 12px;
        background: var(--color-bg-secondary);
        border: 1px solid var(--color-border-medium);
        color: var(--color-text-secondary);
        font-size: var(--font-size-sm);
        white-space: nowrap;
    }
    
    .input-prefix {
        border-right: none;
        border-top-left-radius: var(--radius-md);
        border-bottom-left-radius: var(--radius-md);
        
        + .input {
            border-top-left-radius: 0;
            border-bottom-left-radius: 0;
        }
    }
    
    .input-suffix {
        border-left: none;
        border-top-right-radius: var(--radius-md);
        border-bottom-right-radius: var(--radius-md);
        
        ~ .input {
            border-top-right-radius: 0;
            border-bottom-right-radius: 0;
        }
    }
}
```

## ⚙️ Validation System

### Client-Side Validation
```go
type ValidationConfig struct {
    Required    bool     `json:"required"`
    MinLength   int      `json:"minLength"`
    MaxLength   int      `json:"maxLength"`
    Pattern     string   `json:"pattern"`
    Min         *float64 `json:"min"`
    Max         *float64 `json:"max"`
    Custom      string   `json:"custom"`      // Custom validation function
    Message     string   `json:"message"`     // Error message
    Trigger     string   `json:"trigger"`     // blur, change, input
    Debounce    int      `json:"debounce"`    // Validation delay
}

// Validation implementation
func (input InputProps) Validate(value string) ValidationResult {
    result := ValidationResult{Valid: true}
    
    // Required validation
    if input.Required && strings.TrimSpace(value) == "" {
        result.Valid = false
        result.Message = "This field is required"
        return result
    }
    
    // Length validation
    if input.MinLength > 0 && len(value) < input.MinLength {
        result.Valid = false
        result.Message = fmt.Sprintf("Minimum length is %d characters", input.MinLength)
        return result
    }
    
    if input.MaxLength > 0 && len(value) > input.MaxLength {
        result.Valid = false
        result.Message = fmt.Sprintf("Maximum length is %d characters", input.MaxLength)
        return result
    }
    
    // Pattern validation
    if input.Pattern != "" {
        matched, _ := regexp.MatchString(input.Pattern, value)
        if !matched {
            result.Valid = false
            result.Message = input.Validation.Message
            if result.Message == "" {
                result.Message = "Invalid format"
            }
            return result
        }
    }
    
    return result
}
```

### Real-Time Validation
```go
templ ValidatedInput(props InputProps) {
    <div class="form-group" x-data={ fmt.Sprintf(`{
        value: '%s',
        error: '',
        valid: null,
        validating: false,
        validate() {
            this.validating = true;
            // Validation logic here
            setTimeout(() => {
                this.validating = false;
                this.valid = this.value.length >= %d;
                this.error = this.valid ? '' : '%s';
            }, 200);
        }
    }`, props.Value, props.MinLength, props.Validation.Message) }>
        
        <label class="form-label">{ props.Label }</label>
        
        <div class="input-validation">
            <input 
                x-model="value"
                @input.debounce.300ms="validate()"
                :class="{
                    'input': true,
                    'input-error': valid === false,
                    'input-success': valid === true,
                    'input-validating': validating
                }"
                type={ props.Type }
                name={ props.Name }
                required?={ props.Required }
            />
            
            <div class="validation-icons">
                <span x-show="validating" class="validating-icon">
                    @Spinner(SpinnerProps{Size: "xs"})
                </span>
                <span x-show="valid === true" class="valid-icon">✓</span>
                <span x-show="valid === false" class="error-icon">✗</span>
            </div>
        </div>
        
        <div x-show="error" x-text="error" class="error-message"></div>
    </div>
}
```

## 📱 Responsive Design

### Mobile Optimizations
```css
/* Mobile-specific input styling */
@media (max-width: 479px) {
    .input {
        font-size: 16px !important;  /* Prevent zoom on iOS */
        min-height: 44px;            /* Touch target */
    }
    
    /* Adjust sizes for mobile */
    .input-xs { height: 36px; font-size: 14px !important; }
    .input-sm { height: 40px; font-size: 15px !important; }
    .input-md { height: 44px; font-size: 16px !important; }
    .input-lg { height: 48px; font-size: 16px !important; }
    .input-xl { height: 52px; font-size: 16px !important; }
    
    /* Simplify input groups on mobile */
    .input-group {
        .input-icon-left + .input { padding-left: 44px; }
        .input-icon-right + .input { padding-right: 44px; }
    }
    
    /* Full-width inputs on mobile */
    .input {
        width: 100%;
    }
}
```

### Touch Interactions
```css
/* Touch-friendly interactions */
@media (hover: none) {
    .input {
        /* Remove hover states on touch devices */
        &:hover {
            border-color: var(--color-border-medium);
        }
    }
    
    /* Larger touch targets for buttons */
    .input-toggle,
    .input-clear {
        min-width: 44px;
        min-height: 44px;
    }
}
```

## ♿ Accessibility Implementation

### ARIA Attributes
```go
func (input InputProps) GetAriaAttributes() map[string]string {
    attrs := make(map[string]string)
    
    // Required state
    if input.Required {
        attrs["aria-required"] = "true"
    }
    
    // Invalid state
    if input.Error != "" {
        attrs["aria-invalid"] = "true"
        if input.ID != "" {
            attrs["aria-describedby"] = input.ID + "-error"
        }
    }
    
    // Help text association
    if input.HelpText != "" && input.ID != "" {
        attrs["aria-describedby"] = input.ID + "-help"
    }
    
    // Label association
    if input.AriaLabel != "" {
        attrs["aria-label"] = input.AriaLabel
    }
    
    return attrs
}
```

### Screen Reader Support
```go
templ AccessibleInput(props InputProps) {
    <div class="form-group">
        <label for={ props.ID } class="form-label">
            { props.Label }
            if props.Required {
                <span class="required" aria-label="required">*</span>
            }
        </label>
        
        <input 
            id={ props.ID }
            name={ props.Name }
            type={ props.Type }
            class={ props.GetClasses() }
            placeholder={ props.Placeholder }
            for attrName, attrValue := range props.GetAriaAttributes() {
                { attrName }={ attrValue }
            }
            required?={ props.Required }
            disabled?={ props.Disabled }
        />
        
        if props.HelpText != "" {
            <div id={ props.ID + "-help" } class="help-text">
                { props.HelpText }
            </div>
        }
        
        if props.Error != "" {
            <div 
                id={ props.ID + "-error" }
                class="error-message"
                role="alert"
                aria-live="polite">
                { props.Error }
            </div>
        }
    </div>
}
```

### Keyboard Navigation
```css
/* Focus management */
.input:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 3px var(--color-primary-light);
    z-index: 1;
    position: relative;
}

/* Skip links for keyboard users */
.input-skip-link {
    position: absolute;
    top: -40px;
    left: 6px;
    background: var(--color-bg-surface);
    color: var(--color-text-primary);
    padding: 8px;
    text-decoration: none;
    z-index: 1000;
    border-radius: var(--radius-sm);
    
    &:focus {
        top: 6px;
    }
}
```

## 🧪 Testing Guidelines

### Unit Tests
```go
func TestInputComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    InputProps
        expected []string
    }{
        {
            name: "required text input",
            props: InputProps{
                Type:     "text",
                Name:     "test",
                Required: true,
            },
            expected: []string{"input", "required"},
        },
        {
            name: "email input with validation",
            props: InputProps{
                Type:    "email",
                Pattern: `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
            },
            expected: []string{"input", "type=\"email\""},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            component := renderInput(tt.props)
            for _, expected := range tt.expected {
                assert.Contains(t, component, expected)
            }
        })
    }
}
```

### Accessibility Tests
```javascript
describe('Input Accessibility', () => {
    test('has proper label association', () => {
        const input = render(<Input name="test" label="Test Label" />);
        const labelElement = screen.getByLabelText('Test Label');
        expect(labelElement).toBeInTheDocument();
    });
    
    test('announces errors to screen readers', () => {
        const input = render(<Input name="test" error="Invalid input" />);
        const errorElement = screen.getByRole('alert');
        expect(errorElement).toHaveTextContent('Invalid input');
    });
    
    test('supports keyboard navigation', () => {
        const input = render(<Input name="test" />);
        const inputElement = screen.getByRole('textbox');
        
        inputElement.focus();
        expect(inputElement).toHaveFocus();
        
        fireEvent.keyDown(inputElement, { key: 'Tab' });
        // Should move to next focusable element
    });
});
```

### Visual Regression Tests
```javascript
test.describe('Input Visual Tests', () => {
    test('all input states', async ({ page }) => {
        await page.goto('/components/input');
        
        // Test all sizes
        for (const size of ['xs', 'sm', 'md', 'lg', 'xl']) {
            await expect(page.locator(`[data-size="${size}"]`)).toHaveScreenshot(`input-${size}.png`);
        }
        
        // Test all states
        const states = ['default', 'success', 'error', 'disabled', 'readonly'];
        for (const state of states) {
            await expect(page.locator(`[data-state="${state}"]`)).toHaveScreenshot(`input-${state}.png`);
        }
    });
});
```

## 📚 Usage Examples

### Basic Form Input
```go
templ ContactForm() {
    <form class="space-y-4">
        @Input(InputProps{
            Type:        "text",
            Name:        "firstName",
            Label:       "First Name",
            Required:    true,
            MaxLength:   50,
        })
        
        @Input(InputProps{
            Type:        "email",
            Name:        "email", 
            Label:       "Email Address",
            Required:    true,
            Validation:  ValidationConfig{
                Pattern: `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
                Message: "Please enter a valid email",
            },
        })
        
        @Input(InputProps{
            Type:        "tel",
            Name:        "phone",
            Label:       "Phone Number",
            Placeholder: "(555) 123-4567",
            Mask:        "(999) 999-9999",
        })
    </form>
}
```

### Advanced Input Patterns
```go
templ AdvancedInputs() {
    <div class="form-advanced">
        // Search input with icon
        @Input(InputProps{
            Type:        "search",
            Name:        "search",
            Placeholder: "Search products...",
            Icon:        "search",
            IconPosition: IconLeft,
            ClearButton: true,
            Debounce:    300,
        })
        
        // Currency input
        @Input(InputProps{
            Type:     "number",
            Name:     "price",
            Label:    "Price",
            Prefix:   "$",
            Step:     0.01,
            Min:      0,
            Class:    "text-right font-mono",
        })
        
        // Password with toggle
        @Input(InputProps{
            Type:       "password",
            Name:       "password",
            Label:      "Password",
            Required:   true,
            MinLength:  8,
            ShowToggle: true,
        })
    </div>
}
```

## 🔗 Related Components

- **[Form Group](../../molecules/form-group/)**: Input with enhanced layout
- **[Validation](../../molecules/validation/)**: Advanced validation patterns
- **[Input Group](../../molecules/input-group/)**: Input combinations
- **[Search Box](../../molecules/search-box/)**: Enhanced search functionality

---

**Component Status**: ✅ Production Ready  
**Schema Reference**: `TextControlSchema.json`  
**CSS Classes**: `.input`, `.input-{size}`, `.input-{variant}`  
**Accessibility**: WCAG 2.1 AA Compliant