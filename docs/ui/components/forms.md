# Form Components Reference

**FILE PURPOSE**: Form components, patterns, and interactive elements with validation
**DEPENDENCIES**: Templ + HTMX + Alpine.js + Flowbite
**TARGET AUDIENCE**: Developers building form interfaces
**RELATED FILES**: `validation-guide.md` (implementation details), `elements.md` (basic components)

## Overview

This reference documents advanced form components and patterns for building interactive, validated forms in ERP applications. These components extend basic elements with validation, real-time feedback, and complex interactions.

**TECHNOLOGY INTEGRATION:**
- **Templ** - Type-safe form rendering
- **HTMX** - Server-side validation and partial updates
- **Alpine.js** - Client-side reactivity and state management
- **Flowbite** - Consistent form styling and validation states

## Form Component Architecture

**VALIDATION LAYERS:**
1. **Client-Side** (Alpine.js) - Immediate feedback, format validation
2. **Real-Time Server** (HTMX) - Business logic, uniqueness checks
3. **Form Submission** (Go Backend) - Final validation, data integrity

**COMPONENT HIERARCHY:**
```
Form Container
├── Validated Input Components
├── Complex Form Elements
├── Multi-step Form Logic
└── Submission Handling
```

---

## Validated Input Components

### ValidatedTextInput

**PURPOSE**: Text input with real-time validation and visual feedback

**PROPS STRUCTURE:**
```go
type ValidatedInputProps struct {
    Name               string  // Form field name
    Label              string  // Display label
    Value              string  // Current value
    Placeholder        string  // Placeholder text
    Required           bool    // Required field indicator
    Error              string  // Server-side error message
    ValidationEndpoint string  // HTMX validation URL
    InputType          string  // "text", "email", "password", "number"
}
```

**USAGE EXAMPLE:**
```go
@ValidatedTextInput(ValidatedInputProps{
    Name: "first_name",
    Label: "First Name",
    Value: form.FirstName,
    Required: true,
    Error: errors.GetFieldError("first_name"),
    ValidationEndpoint: "/validate?field=first_name",
})
```

**VISUAL STATES:**
- **Default**: Gray border, no icon
- **Validating**: Yellow border, spinner icon
- **Valid**: Green border, checkmark icon
- **Invalid**: Red border, error icon

**ALPINE.JS BEHAVIOR:**
```javascript
// Debounced validation (500ms)
@input.debounce.500ms="validateField()"
// Immediate validation on blur
@blur="validateField()"
```

### ValidatedEmailInput

**PURPOSE**: Email input with format validation and uniqueness checking

**ENHANCED FEATURES:**
- Client-side email format validation
- Server-side uniqueness verification
- Domain validation (optional)
- Suggestion for common typos

**USAGE EXAMPLE:**
```go
@ValidatedEmailInput(ValidatedInputProps{
    Name: "email",
    Label: "Email Address",
    Value: form.Email,
    Required: true,
    ValidationEndpoint: "/validate?field=email&check=unique",
})
```

**VALIDATION FLOW:**
1. Format check (immediate)
2. Domain validation (debounced)
3. Uniqueness check (server)
4. Visual feedback update

### ValidatedFileInput

**PURPOSE**: File upload with type, size, and security validation

**PROPS STRUCTURE:**
```go
type ValidatedFileProps struct {
    Name          string   // Form field name
    Label         string   // Display label
    Required      bool     // Required field indicator
    AcceptedTypes []string // Allowed file types
    MaxSize       string   // Maximum file size
    Multiple      bool     // Allow multiple files
}
```

**USAGE EXAMPLE:**
```go
@ValidatedFileInput(ValidatedFileProps{
    Name: "document",
    Label: "Upload Document",
    AcceptedTypes: []string{"pdf", "doc", "docx"},
    MaxSize: "5MB",
    Required: false,
})
```

**CLIENT-SIDE VALIDATION:**
- File type checking
- Size limit enforcement
- Security scanning (basic)
- Progress indication

---

## Advanced Form Elements

### CascadingSelect

**PURPOSE**: Select input that depends on parent field selection

**PROPS STRUCTURE:**
```go
type CascadingSelectProps struct {
    Name              string        // Field name
    Label             string        // Display label
    ParentField       string        // Dependent field name
    ParentLabel       string        // Parent field label
    DependentEndpoint string        // Options loading URL
    Required          bool          // Required field indicator
    Value             string        // Selected value
}
```

**USAGE EXAMPLE:**
```go
// Country selector
@ValidatedSelectInput(ValidatedSelectProps{
    Name: "country",
    Label: "Country",
    Options: countryOptions,
})

// State/Province selector (depends on country)
@CascadingSelectInput(CascadingSelectProps{
    Name: "state",
    Label: "State/Province", 
    ParentField: "country",
    ParentLabel: "country",
    DependentEndpoint: "/api/states",
})
```

**INTERACTION FLOW:**
1. User selects parent field
2. Alpine.js watches for changes
3. AJAX request loads dependent options
4. Select updates with new options
5. Loading states provide feedback

### DynamicFieldGroup

**PURPOSE**: Add/remove field groups dynamically (e.g., multiple phone numbers)

**USAGE EXAMPLE:**
```go
@DynamicFieldGroup(DynamicFieldProps{
    Name: "skills",
    Label: "Skills",
    Fields: form.Skills,
    Template: skillFieldTemplate(),
    MaxItems: 10,
    AddButtonText: "Add Skill",
})
```

**ALPINE.JS FEATURES:**
```javascript
// Add new field group
addField() {
    this.fields.push({ skill: '', level: '', years: '' });
}

// Remove field group
removeField(index) {
    this.fields.splice(index, 1);
}

// Validate all field groups
validateAllFields() {
    return this.fields.every(field => this.validateField(field));
}
```

---

## Form Patterns

### Multi-Step Forms

**PURPOSE**: Break complex forms into manageable steps with validation

**STEP STRUCTURE:**
```go
type FormStep struct {
    Title       string   // Step display title
    Fields      []string // Fields in this step
    Validation  string   // Validation requirements
    Template    templ.Component // Step content
}
```

**USAGE EXAMPLE:**
```go
@MultiStepForm(MultiStepProps{
    CurrentStep: 0,
    Steps: []FormStep{
        {
            Title: "Personal Information",
            Fields: []string{"first_name", "last_name", "email"},
            Template: personalInfoStep(),
        },
        {
            Title: "Employment Details", 
            Fields: []string{"role", "department", "salary"},
            Template: employmentStep(),
        },
        {
            Title: "Review & Submit",
            Fields: []string{},
            Template: reviewStep(),
        },
    },
})
```

**STEP NAVIGATION:**
- Progress indicator with completed steps
- Validation before advancing
- Data persistence between steps
- Review summary on final step

**ALPINE.JS STATE MANAGEMENT:**
```javascript
multiStepFormHandler() {
    return {
        currentStep: 0,
        stepValidation: {},
        
        nextStep() {
            if (this.validateCurrentStep()) {
                this.currentStep++;
            }
        },
        
        previousStep() {
            this.currentStep--;
        }
    }
}
```

### Form Builder Pattern

**PURPOSE**: Dynamically generate forms based on configuration

**CONFIGURATION STRUCTURE:**
```go
type FormConfig struct {
    Title       string      `json:"title"`
    Description string      `json:"description"`
    Fields      []FieldConfig `json:"fields"`
    Validation  ValidationConfig `json:"validation"`
}

type FieldConfig struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Label       string      `json:"label"`
    Required    bool        `json:"required"`
    Options     []Option    `json:"options,omitempty"`
    Validation  []string    `json:"validation,omitempty"`
}
```

**USAGE EXAMPLE:**
```go
@DynamicForm(DynamicFormProps{
    Config: formConfig,
    Values: existingData,
    Errors: validationErrors,
})
```

**BENEFITS:**
- Database-driven form generation
- A/B testing different form layouts
- User-customizable forms
- Rapid prototyping capabilities

---

## Form Validation States

### Visual Feedback System

**VALIDATION STATES:**
```css
/* Default state */
.field-default { border-color: #d1d5db; }

/* Validating state */
.field-validating { 
    border-color: #fbbf24; 
    animation: pulse 2s infinite;
}

/* Valid state */
.field-valid { 
    border-color: #10b981;
    background-image: url('data:image/svg+xml,<svg...checkmark>');
}

/* Invalid state */
.field-invalid { 
    border-color: #ef4444;
    background-image: url('data:image/svg+xml,<svg...error>');
}
```

**ICON INDICATORS:**
- Spinner: Validation in progress
- Checkmark: Field is valid
- Error icon: Validation failed
- Info icon: Additional information needed

**COLOR CODING:**
- **Gray**: Default/neutral state
- **Yellow**: Processing/validating
- **Green**: Valid/success
- **Red**: Invalid/error

### Error Display Patterns

**ERROR HIERARCHY:**
1. **Field-Level**: Individual field validation messages
2. **Section-Level**: Related field group errors
3. **Form-Level**: Cross-field validation issues
4. **System-Level**: Technical or permission errors

**ERROR MESSAGE GUIDELINES:**
```go
// ✅ Good: Specific and actionable
"Please enter a valid email address (example@company.com)"

// ❌ Bad: Generic and unhelpful
"Invalid input"

// ✅ Good: Explains the constraint
"Password must be at least 8 characters with 1 number and 1 symbol"

// ❌ Bad: No guidance for correction
"Password requirements not met"
```

**ERROR RECOVERY ACTIONS:**
- Clear error on field focus
- Provide format examples
- Suggest corrections where possible
- Auto-format when applicable

---

## Accessibility Features

### Screen Reader Support

**ARIA ATTRIBUTES:**
```go
// Proper label association
<label for="email-input">Email Address *</label>
<input 
    id="email-input"
    aria-describedby="email-help email-error"
    aria-required="true"
    aria-invalid="false">
<div id="email-help">We'll never share your email</div>
<div id="email-error" role="alert">Please enter a valid email</div>
```

**KEYBOARD NAVIGATION:**
- Tab order follows logical flow
- Enter submits forms appropriately
- Arrow keys navigate select options
- Escape closes modals/dropdowns

**ERROR ANNOUNCEMENT:**
```go
// Use role="alert" for immediate announcement
<div role="alert" x-show="hasError" x-text="errorMessage"></div>

// Use aria-live for dynamic updates
<div aria-live="polite" x-text="validationStatus"></div>
```

**FOCUS MANAGEMENT:**
- First invalid field receives focus on submission
- Modal forms trap focus appropriately
- Validation errors announced immediately
- Loading states communicated clearly

### Mobile Optimization

**TOUCH-FRIENDLY DESIGN:**
- Minimum 44px touch targets
- Appropriate virtual keyboards
- Thumb-accessible controls
- Responsive error layouts

**INPUT TYPES FOR MOBILE:**
```go
// Email input triggers email keyboard
Type: "email"

// Phone input triggers number pad
Type: "tel"

// Numeric input with decimal support
Type: "number"

// Date input shows date picker
Type: "date"
```

**RESPONSIVE FORM LAYOUTS:**
```go
<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    // Single column on mobile, two columns on desktop
</div>
```

---

## Performance Optimization

### Debouncing and Caching

**DEBOUNCING STRATEGY:**
```javascript
// Different timing for different validations
@input.debounce.100ms="formatField()"    // Immediate formatting
@input.debounce.500ms="validateField()"  // Server validation
@input.debounce.1000ms="checkUnique()"   // Expensive operations
```

**VALIDATION CACHING:**
```go
// Cache validation results
func (v *Validator) CachedEmailValidation(email string) bool {
    if result, found := validationCache.Get("email:" + email); found {
        return result.(bool)
    }
    
    isValid := v.validateEmailFormat(email)
    validationCache.Set("email:" + email, isValid, 5*time.Minute)
    return isValid
}
```

**REQUEST OPTIMIZATION:**
- Batch multiple field validations
- Cancel previous requests on new input
- Minimize payload size
- Use HTTP status codes effectively

### Progressive Enhancement

**GRACEFUL DEGRADATION:**
```go
// Forms work without JavaScript
<form method="POST" action="/users/create">
    @BasicTextInput(BasicInputProps{
        Name: "first_name",
        Required: true,
    })
    <button type="submit">Create User</button>
</form>

// Enhanced with HTMX and Alpine.js
<form 
    hx-post="/users/create"
    hx-target="#form-container"
    x-data="formHandler()">
    @ValidatedTextInput(ValidatedInputProps{
        Name: "first_name",
        ValidationEndpoint: "/validate?field=first_name",
    })
    <button type="submit" :disabled="!isValid">Create User</button>
</form>
```

**ENHANCEMENT LAYERS:**
1. **Base**: HTML forms with server validation
2. **HTMX**: Partial page updates, real-time validation
3. **Alpine.js**: Client-side state, enhanced UX
4. **Advanced**: Real-time collaboration, auto-save

---

## Integration with ERP Features

### Permission-Based Forms

**ROLE-BASED FIELD ACCESS:**
```go
@ValidatedTextInput(ValidatedInputProps{
    Name: "salary",
    Label: "Salary",
    Disabled: !user.CanEditSalary(),
    Visible: user.CanViewSalary(),
})
```

**CONDITIONAL FIELD RENDERING:**
```go
if user.HasRole("manager") {
    @ManagerOnlyFields()
}

if user.HasPermission("edit_sensitive_data") {
    @SensitiveDataFields()
}
```

### Workflow Integration

**STATE-DEPENDENT VALIDATION:**
```go
func (f *OrderForm) ValidateForState(state string) {
    switch state {
    case "draft":
        // Basic validation only
        f.validateRequired()
    case "submitted":
        // Full validation required
        f.validateComplete()
    case "approved":
        // Read-only, no editing
        f.readonly = true
    }
}
```

**APPROVAL WORKFLOWS:**
```go
@WorkflowForm(WorkflowProps{
    CurrentState: order.Status,
    NextStates: getAvailableTransitions(order.Status),
    ApprovalRequired: order.RequiresApproval(),
})
```

---

## Testing Form Components

### Component Testing

**VALIDATION TESTING:**
```go
func TestValidatedInput(t *testing.T) {
    props := ValidatedInputProps{
        Name: "email",
        ValidationEndpoint: "/test/validate",
    }
    
    // Test rendering
    buf := &bytes.Buffer{}
    err := ValidatedTextInput(props).Render(context.Background(), buf)
    require.NoError(t, err)
    
    // Test HTML structure
    html := buf.String()
    assert.Contains(t, html, `name="email"`)
    assert.Contains(t, html, `x-data="fieldValidator()"`)
}
```

**INTERACTION TESTING:**
```javascript
// Test Alpine.js validation logic
describe('Field Validator', () => {
    test('validates email format', async () => {
        const validator = fieldValidator();
        validator.value = 'invalid-email';
        
        await validator.validateField();
        
        expect(validator.hasError).toBe(true);
        expect(validator.errorMessage).toContain('valid email');
    });
});
```

### End-to-End Testing

**USER INTERACTION FLOWS:**
```javascript
// Playwright test example
test('form validation flow', async ({ page }) => {
    await page.goto('/users/create');
    
    // Test invalid input
    await page.fill('[name="email"]', 'invalid-email');
    await page.blur('[name="email"]');
    
    // Verify error display
    await expect(page.locator('.error-message')).toContainText('valid email');
    
    // Test valid input
    await page.fill('[name="email"]', 'user@example.com');
    await page.blur('[name="email"]');
    
    // Verify success state
    await expect(page.locator('.success-icon')).toBeVisible();
});
```

---

## Best Practices

### Component Design

**COMPOSITION OVER INHERITANCE:**
```go
// ✅ Good: Composable components
@FormContainer() {
    @ValidatedTextInput(nameProps)
    @ValidatedEmailInput(emailProps)
    @SubmitButton(buttonProps)
}

// ❌ Bad: Monolithic form component
@UserFormWithEverything(massiveProps)
```

**SINGLE RESPONSIBILITY:**
```go
// ✅ Good: Focused validation logic
func validateEmail(email string) ValidationResult
func checkEmailUniqueness(email string) bool

// ❌ Bad: Mixed concerns
func validateAndSaveUser(userData map[string]interface{}) error
```

**PREDICTABLE STATE MANAGEMENT:**
```javascript
// ✅ Good: Clear state transitions
const ValidationStates = {
    IDLE: 'idle',
    VALIDATING: 'validating', 
    VALID: 'valid',
    INVALID: 'invalid'
};

// ❌ Bad: Ambiguous boolean flags
let isValid, isValidating, hasError, wasValidated;
```

### Security Considerations

**CLIENT-SIDE VALIDATION LIMITATIONS:**
- Never rely solely on client validation
- Sanitize all user inputs on server
- Validate file uploads thoroughly
- Rate limit validation endpoints

**XSS PREVENTION:**
```go
// Escape user content in error messages
ErrorMessage: template.HTMLEscapeString(userInput)

// Use Templ's built-in escaping
{ userInput } // Automatically escaped
```

**CSRF PROTECTION:**
```go
// Include CSRF token in forms
<input type="hidden" name="csrf_token" value={ csrfToken }>
```

## References

**RELATED DOCUMENTATION:**
- [Elements Reference](elements.md) - Basic UI components
- [Validation Guide](../guides/validation-guide.md) - Complete validation implementation
- [HTMX Integration](../patterns/htmx-integration.md) - Server interaction patterns

**EXTERNAL REFERENCES:**

- [templ-llms.md](../templ-llms.md)   - Advanced Templ features (streaming, suspense patterns)
- [flowbite-llms-full.txt](../flowbite-llms-full.txt)  - Complete Flowbite component catalog

**OFFICIAL DOCUMENTATION:**
- [Templ Guide](https://templ.guide) - Go templating language reference
- [Flowbite Components](../flowbite/) - UI component library
- [Alpine.js Documentation](https://alpinejs.dev) - Reactive JavaScript framework
- [TailwindCSS](https://tailwindcss.com/docs) - Utility-first CSS framework
- [Schema](../Schema/) - Ui framework Json Schema 
- [Design](../design/README.md) 
- [HTMX Documentation](https://htmx.org/docs/) - Hypermedia interactions
- [Alpine.js Guide](https://alpinejs.dev) - Reactive JavaScript framework
- [Flowbite Forms](../components/forms.md) - Form component styling
