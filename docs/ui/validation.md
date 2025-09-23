# Form Validation System Documentation

## Overview

This documentation covers a comprehensive form validation system designed for ERP applications using **Templ + HTMX + Alpine.js + Flowbite**. The system provides three layers of validation with real-time feedback, maintaining both security and excellent user experience.

## System Architecture

### Validation Layers

1. **Client-Side Validation (Alpine.js)**
   - Immediate user feedback
   - Field format validation  
   - Prevents unnecessary server calls
   - Enhanced user experience

2. **Real-Time Server Validation (HTMX)**
   - Business logic validation
   - Database uniqueness checks
   - Cross-field dependencies
   - Security enforcement

3. **Form Submission Validation (Go Backend)**
   - Final comprehensive validation
   - Data integrity enforcement
   - Error handling and rollback
   - Audit logging

### Component Structure

```
Validation System
├── Backend (Go)
│   ├── Validation Rules Engine
│   ├── Business Logic Validators
│   ├── Database Validators
│   └── Permission-Based Validators
├── Frontend Components (Templ)
│   ├── Input Components with Validation
│   ├── Form Templates
│   ├── Error Display Components
│   └── Success Feedback Components  
├── Client Logic (Alpine.js)
│   ├── Field Validators
│   ├── Form State Management
│   ├── Real-time Feedback
│   └── User Interaction Handlers
└── Styling (Flowbite)
    ├── Validation States
    ├── Error Styling
    ├── Success Indicators
    └── Loading States
```

## Validation Rules Engine

### Core Validation Types

**String Validation:**
- Required field validation
- Length constraints (min/max)
- Pattern matching (regex)
- Character set validation

**Format Validation:**
- Email format validation
- Phone number validation  
- URL validation
- Date/time validation

**Numeric Validation:**
- Range validation (min/max values)
- Decimal precision
- Currency formatting
- Mathematical constraints

**Business Logic Validation:**
- Uniqueness constraints
- Cross-field dependencies
- Role-based validation
- Workflow state validation

### Validation Timing

**Immediate Validation:**
- Field format checks
- Client-side rules
- Pattern matching
- Length validation

**Debounced Validation (500ms):**
- Server-side uniqueness checks
- Complex business rules
- Database lookups
- External API validation

**Submission Validation:**
- Complete form validation
- Transaction integrity
- Final security checks
- Data persistence validation

## Implementation Patterns

### Field Validation Pattern

Each form field follows this validation pattern:

1. **User Input** → Client-side format check
2. **Input Valid** → Debounced server validation  
3. **Server Response** → Visual feedback update
4. **Form Submission** → Complete validation
5. **Success/Error** → User feedback

### Error Handling Hierarchy

1. **Field-Level Errors:** Individual field validation messages
2. **Form-Level Errors:** Cross-field validation issues  
3. **Business Rule Errors:** Domain-specific validation
4. **System Errors:** Technical validation failures

### Visual Feedback System

**Input States:**
- Default: Gray border, no icon
- Validating: Yellow border, spinning icon
- Valid: Green border, checkmark icon
- Invalid: Red border, error icon

**Error Display:**
- Inline error messages below fields
- Color-coded text and borders
- Icon indicators for quick recognition
- Toast notifications for form-level issues

## Component Usage Guide

### Basic Input Validation

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

**Properties:**
- `Name`: Form field name
- `Label`: Display label for the field
- `Value`: Current field value
- `Required`: Whether field is mandatory
- `Error`: Server-side error message
- `ValidationEndpoint`: HTMX validation URL

### Advanced Validation Components

**Email Input with Uniqueness Check:**
```go
@ValidatedEmailInput(ValidatedInputProps{
    Name: "email",
    Label: "Email Address",
    Value: form.Email, 
    Required: true,
    ValidationEndpoint: "/validate?field=email&check=unique",
})
```

**File Upload with Validation:**
```go
@ValidatedFileInput(ValidatedInputProps{
    Name: "document",
    Label: "Upload Document",
    Required: false,
    AcceptedTypes: []string{"pdf", "doc", "docx"},
    MaxSize: "5MB",
})
```

**Dynamic Field Groups:**
```go
@DynamicFieldGroup("Skills", form.Skills, errors)
```

## Best Practices

### Performance Optimization

**Debouncing Strategy:**
- Use 500ms debounce for server validation
- Immediate validation for format checks
- Cache validation results when possible
- Batch multiple field validations

**Network Efficiency:**
- Include related field values in validation requests
- Use HTTP status codes for validation state
- Minimize payload size in responses
- Implement request cancellation

### User Experience Guidelines

**Progressive Enhancement:**
- Forms work without JavaScript
- Graceful degradation for slow connections
- Clear loading states during validation
- Intuitive error recovery

**Accessibility Standards:**
- Screen reader compatible error messages
- Keyboard navigation support
- High contrast error indicators
- ARIA labels and descriptions

**Mobile Optimization:**
- Touch-friendly input fields
- Appropriate virtual keyboards
- Thumb-accessible validation controls
- Responsive error message layout

### Security Considerations

**Client-Side Validation:**
- Never rely solely on client validation
- Validate formats to prevent XSS
- Sanitize display of error messages
- Rate limit validation requests

**Server-Side Validation:**
- Always validate on the server
- Use parameterized queries
- Implement CSRF protection
- Log suspicious validation patterns

## Error Recovery Patterns

### User-Friendly Error Messages

**Instead of:** "Validation error: field required"  
**Use:** "Please enter your first name"

**Instead of:** "Invalid format"  
**Use:** "Please enter a valid email address (example@company.com)"

**Instead of:** "Constraint violation"  
**Use:** "This email address is already registered"

### Error Recovery Actions

**Field-Level Recovery:**
- Clear error on field focus
- Provide format examples
- Suggest corrections
- Auto-format when possible

**Form-Level Recovery:**  
- Highlight first invalid field
- Scroll to error location
- Provide error summary
- Save valid data during correction

## Integration with ERP Systems

### Permission-Based Validation

```go
// Validate based on user role
func (f *UserForm) ValidateWithPermissions(validator *Validator, user *User) {
    // Standard validation
    f.Validate(validator)
    
    // Role-specific validation
    if !user.CanAssignRole(f.Role) {
        validator.AddError("role", "Insufficient permissions", "permission_denied")
    }
}
```

### Workflow State Validation

```go
// Validate based on current workflow state  
func (f *OrderForm) ValidateWorkflowTransition(validator *Validator, currentState string) {
    allowedTransitions := GetWorkflowTransitions(currentState)
    if !contains(allowedTransitions, f.Status) {
        validator.AddError("status", "Invalid status transition", "workflow_error")
    }
}
```

### Audit Integration

```go
// Log validation events for compliance
func (h *Handler) logValidationEvent(field string, valid bool, userID int) {
    h.auditLogger.Log(AuditEvent{
        Type: "field_validation",
        UserID: userID,
        Field: field,
        Valid: valid,
        Timestamp: time.Now(),
    })
}
```

## Testing Strategy

### Unit Testing

**Validation Rules:**
```go
func TestEmailValidation(t *testing.T) {
    validator := NewValidator()
    
    // Test valid email
    validator.Email("email", "user@example.com")
    assert.True(t, validator.IsValid())
    
    // Test invalid email
    validator.Email("email", "invalid-email")
    assert.False(t, validator.IsValid())
}
```

### Integration Testing

**HTMX Validation Endpoints:**
```go
func TestValidateFieldEndpoint(t *testing.T) {
    req := httptest.NewRequest("POST", "/validate?field=email", nil)
    req.Form = url.Values{"email": {"test@example.com"}}
    
    w := httptest.NewRecorder()
    handler.ValidateField(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### End-to-End Testing

**User Interaction Testing:**
- Test form submission flows
- Validate error recovery paths  
- Check accessibility compliance
- Verify mobile responsiveness

## Monitoring and Analytics

### Validation Metrics

**Track Key Metrics:**
- Field validation failure rates
- Most common validation errors
- Time to successful form completion
- User abandonment after validation errors

**Performance Metrics:**
- Validation response times
- Client-side validation coverage
- Server-side validation load
- Error message effectiveness

### Error Analysis

**Common Issues:**
- High validation failure rates indicate UX problems
- Repeated errors suggest unclear requirements
- Slow validation responses hurt user experience
- Inconsistent error messages confuse users

**Optimization Opportunities:**
- Pre-populate fields when possible
- Provide better input examples
- Improve error message clarity
- Optimize validation performance

## Maintenance and Evolution

### Adding New Validation Rules

1. **Define Rule Logic:** Create validation function in rules package
2. **Add Form Integration:** Include rule in form validation method
3. **Update Components:** Add client-side validation if needed
4. **Test Thoroughly:** Unit test rule logic and integration
5. **Document Changes:** Update validation documentation

### Handling Breaking Changes

**Backward Compatibility:**
- Version validation API endpoints
- Maintain legacy error message formats  
- Graceful degradation for old clients
- Migration guides for form updates

**Field Evolution:**
- Support both old and new field names
- Gradually migrate validation rules
- Maintain error message consistency
- Communicate changes to users

This validation system provides a robust foundation for maintaining data quality while delivering excellent user experience in your ERP application.

## 1. Go Backend Validation Structure

### Validation Types and Errors

```go
// validation/types.go
package validation

import (
    "fmt"
    "regexp"
    "strings"
)

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) HasField(field string) bool {
    for _, err := range ve {
        if err.Field == field {
            return true
        }
    }
    return false
}

func (ve ValidationErrors) GetFieldError(field string) string {
    for _, err := range ve {
        if err.Field == field {
            return err.Message
        }
    }
    return ""
}

type Validator struct {
    errors ValidationErrors
}

func NewValidator() *Validator {
    return &Validator{
        errors: make(ValidationErrors, 0),
    }
}

func (v *Validator) AddError(field, message, code string) {
    v.errors = append(v.errors, ValidationError{
        Field:   field,
        Message: message,
        Code:    code,
    })
}

func (v *Validator) IsValid() bool {
    return len(v.errors) == 0
}

func (v *Validator) Errors() ValidationErrors {
    return v.errors
}
```

### Validation Rules

```go
// validation/rules.go
package validation

import (
    "regexp"
    "strings"
    "unicode"
)

// String validation rules
func (v *Validator) Required(field, value string) *Validator {
    if strings.TrimSpace(value) == "" {
        v.AddError(field, "This field is required", "required")
    }
    return v
}

func (v *Validator) MinLength(field, value string, min int) *Validator {
    if len(strings.TrimSpace(value)) < min {
        v.AddError(field, fmt.Sprintf("Must be at least %d characters long", min), "min_length")
    }
    return v
}

func (v *Validator) MaxLength(field, value string, max int) *Validator {
    if len(value) > max {
        v.AddError(field, fmt.Sprintf("Must be no more than %d characters long", max), "max_length")
    }
    return v
}

func (v *Validator) Email(field, value string) *Validator {
    if value != "" {
        emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
        if !emailRegex.MatchString(value) {
            v.AddError(field, "Please enter a valid email address", "invalid_email")
        }
    }
    return v
}

func (v *Validator) Phone(field, value string) *Validator {
    if value != "" {
        phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
        if !phoneRegex.MatchString(strings.ReplaceAll(value, " ", "")) {
            v.AddError(field, "Please enter a valid phone number", "invalid_phone")
        }
    }
    return v
}

func (v *Validator) Numeric(field, value string) *Validator {
    if value != "" {
        numericRegex := regexp.MustCompile(`^\d+(\.\d+)?$`)
        if !numericRegex.MatchString(value) {
            v.AddError(field, "Must be a valid number", "invalid_number")
        }
    }
    return v
}

func (v *Validator) AlphaNumeric(field, value string) *Validator {
    if value != "" {
        for _, char := range value {
            if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
                v.AddError(field, "Must contain only letters and numbers", "invalid_alphanumeric")
                break
            }
        }
    }
    return v
}

func (v *Validator) OneOf(field, value string, options []string) *Validator {
    if value != "" {
        valid := false
        for _, option := range options {
            if value == option {
                valid = true
                break
            }
        }
        if !valid {
            v.AddError(field, "Please select a valid option", "invalid_option")
        }
    }
    return v
}

// Business logic validation
func (v *Validator) UniqueEmail(field, email string, db *Database, excludeID int) *Validator {
    if email != "" {
        exists, err := db.EmailExists(email, excludeID)
        if err != nil {
            v.AddError(field, "Unable to verify email uniqueness", "validation_error")
        } else if exists {
            v.AddError(field, "This email address is already in use", "email_taken")
        }
    }
    return v
}
```

### Form Data Structure

```go
// forms/user.go
package forms

type UserForm struct {
    ID          int    `json:"id"`
    FirstName   string `json:"first_name" form:"first_name"`
    LastName    string `json:"last_name" form:"last_name"`
    Email       string `json:"email" form:"email"`
    Phone       string `json:"phone" form:"phone"`
    Role        string `json:"role" form:"role"`
    Department  string `json:"department" form:"department"`
    Salary      string `json:"salary" form:"salary"`
    StartDate   string `json:"start_date" form:"start_date"`
    Status      string `json:"status" form:"status"`
}

func (f *UserForm) Validate(v *validation.Validator, db *Database) {
    // Required fields
    v.Required("first_name", f.FirstName).
      Required("last_name", f.LastName).
      Required("email", f.Email).
      Required("role", f.Role)

    // String length validation
    v.MinLength("first_name", f.FirstName, 2).
      MaxLength("first_name", f.FirstName, 50).
      MinLength("last_name", f.LastName, 2).
      MaxLength("last_name", f.LastName, 50)

    // Format validation
    v.Email("email", f.Email).
      Phone("phone", f.Phone).
      Numeric("salary", f.Salary)

    // Business rules
    v.OneOf("role", f.Role, []string{"admin", "manager", "employee", "contractor"}).
      OneOf("status", f.Status, []string{"active", "inactive", "pending"})

    // Database validation
    v.UniqueEmail("email", f.Email, db, f.ID)
}
```

## 2. Server-side Handlers

### HTMX Validation Endpoints

```go
// handlers/validation.go
package handlers

import (
    "encoding/json"
    "net/http"
)

// Real-time field validation
func (h *Handler) ValidateField(w http.ResponseWriter, r *http.Request) {
    field := r.URL.Query().Get("field")
    value := r.FormValue(field)
    
    validator := validation.NewValidator()
    
    switch field {
    case "email":
        validator.Required(field, value).Email(field, value)
        if validator.IsValid() {
            validator.UniqueEmail(field, value, h.db, 0)
        }
    case "first_name", "last_name":
        validator.Required(field, value).
                MinLength(field, value, 2).
                MaxLength(field, value, 50)
    case "phone":
        if value != "" {
            validator.Phone(field, value)
        }
    case "salary":
        if value != "" {
            validator.Numeric(field, value)
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    
    if validator.IsValid() {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "valid": true,
            "field": field,
        })
    } else {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "valid": false,
            "field": field,
            "error": validator.Errors()[0].Message,
        })
    }
}

// Form submission with full validation
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var form forms.UserForm
    
    // Parse form data
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }
    
    // Bind form values
    form.FirstName = r.FormValue("first_name")
    form.LastName = r.FormValue("last_name")
    form.Email = r.FormValue("email")
    form.Phone = r.FormValue("phone")
    form.Role = r.FormValue("role")
    form.Department = r.FormValue("department")
    form.Salary = r.FormValue("salary")
    form.StartDate = r.FormValue("start_date")
    form.Status = r.FormValue("status")
    
    // Validate form
    validator := validation.NewValidator()
    form.Validate(validator, h.db)
    
    if !validator.IsValid() {
        // Return form with errors
        w.Header().Set("HX-Reswap", "outerHTML")
        w.WriteHeader(http.StatusUnprocessableEntity)
        
        tmpl := templates.CreateUserForm(form, validator.Errors())
        tmpl.Render(r.Context(), w)
        return
    }
    
    // Save user
    user, err := h.db.CreateUser(form)
    if err != nil {
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }
    
    // Success response
    w.Header().Set("HX-Trigger", `{"userCreated": {"id": `+string(rune(user.ID))+`}, "showNotification": {"type": "success", "message": "User created successfully"}}`)
    w.WriteHeader(http.StatusCreated)
    
    // Return updated table row or redirect
    tmpl := templates.UserTableRow(user)
    tmpl.Render(r.Context(), w)
}
```

## 3. Templ Templates with Validation

### Form Template with Error Handling

```go
// templates/forms/user_form.templ
templ CreateUserForm(form forms.UserForm, errors validation.ValidationErrors) {
    @ModalLayout() {
        <div class="p-6">
            <h2 class="text-xl font-semibold mb-4">Create New User</h2>
            
            <form 
                hx-post="/users"
                hx-target="closest .modal"
                hx-swap="outerHTML"
                x-data="userFormHandler()"
                @submit="handleSubmit">
                
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <!-- First Name -->
                    @ValidatedTextInput(ValidatedInputProps{
                        Name: "first_name",
                        Label: "First Name",
                        Value: form.FirstName,
                        Required: true,
                        Error: errors.GetFieldError("first_name"),
                        ValidationEndpoint: "/validate?field=first_name",
                    })
                    
                    <!-- Last Name -->
                    @ValidatedTextInput(ValidatedInputProps{
                        Name: "last_name", 
                        Label: "Last Name",
                        Value: form.LastName,
                        Required: true,
                        Error: errors.GetFieldError("last_name"),
                        ValidationEndpoint: "/validate?field=last_name",
                    })
                    
                    <!-- Email -->
                    @ValidatedEmailInput(ValidatedInputProps{
                        Name: "email",
                        Label: "Email Address", 
                        Value: form.Email,
                        Required: true,
                        Error: errors.GetFieldError("email"),
                        ValidationEndpoint: "/validate?field=email",
                    })
                    
                    <!-- Phone -->
                    @ValidatedTextInput(ValidatedInputProps{
                        Name: "phone",
                        Label: "Phone Number",
                        Value: form.Phone,
                        Required: false,
                        Error: errors.GetFieldError("phone"),
                        ValidationEndpoint: "/validate?field=phone",
                    })
                    
                    <!-- Role -->
                    @ValidatedSelectInput(ValidatedSelectProps{
                        Name: "role",
                        Label: "Role",
                        Value: form.Role,
                        Required: true,
                        Error: errors.GetFieldError("role"),
                        Options: []SelectOption{
                            {Value: "admin", Label: "Administrator"},
                            {Value: "manager", Label: "Manager"},
                            {Value: "employee", Label: "Employee"},
                            {Value: "contractor", Label: "Contractor"},
                        },
                    })
                    
                    <!-- Salary -->
                    @ValidatedTextInput(ValidatedInputProps{
                        Name: "salary",
                        Label: "Salary",
                        Value: form.Salary,
                        Required: false,
                        Error: errors.GetFieldError("salary"),
                        ValidationEndpoint: "/validate?field=salary",
                        Placeholder: "50000.00",
                    })
                </div>
                
                <!-- Form Actions -->
                <div class="flex justify-end space-x-3 mt-6">
                    <button 
                        type="button" 
                        @click="closeModal()"
                        class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300">
                        Cancel
                    </button>
                    <button 
                        type="submit"
                        :disabled="!isValid || submitting"
                        :class="{'bg-blue-600 hover:bg-blue-700': isValid && !submitting, 'bg-gray-400 cursor-not-allowed': !isValid || submitting}"
                        class="px-4 py-2 text-white rounded-md transition-colors">
                        <span x-show="!submitting">Create User</span>
                        <span x-show="submitting" class="flex items-center">
                            <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            Creating...
                        </span>
                    </button>
                </div>
            </form>
        </div>
    }
}
```

### Validated Input Components

```go
// templates/components/validated_inputs.templ
type ValidatedInputProps struct {
    Name               string
    Label              string
    Value              string
    Placeholder        string
    Required           bool
    Error              string
    ValidationEndpoint string
    InputType          string
}

templ ValidatedTextInput(props ValidatedInputProps) {
    <div class="mb-4" x-data="fieldValidator()">
        <label class="block text-sm font-medium text-gray-700 mb-2">
            {props.Label}
            if props.Required {
                <span class="text-red-500">*</span>
            }
        </label>
        
        <div class="relative">
            <input 
                type={templ.SafeString(props.InputType)}
                if props.InputType == "" {
                    type="text"
                }
                name={props.Name}
                value={props.Value}
                placeholder={props.Placeholder}
                if props.Required {
                    required
                }
                class="w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 transition-colors"
                :class="{
                    'border-gray-300 focus:ring-blue-500 focus:border-transparent': !hasError && !validating,
                    'border-red-300 focus:ring-red-500': hasError,
                    'border-green-300 focus:ring-green-500': isValid && !hasError,
                    'border-yellow-300': validating
                }"
                x-model="value"
                @input.debounce.500ms="validateField()"
                @blur="validateField()">
                
            <!-- Loading Spinner -->
            <div x-show="validating" class="absolute inset-y-0 right-0 flex items-center pr-3">
                <svg class="animate-spin h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            </div>
            
            <!-- Success Icon -->
            <div x-show="isValid && !hasError && !validating" class="absolute inset-y-0 right-0 flex items-center pr-3">
                <svg class="h-4 w-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path>
                </svg>
            </div>
            
            <!-- Error Icon -->
            <div x-show="hasError" class="absolute inset-y-0 right-0 flex items-center pr-3">
                <svg class="h-4 w-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
                </svg>
            </div>
        </div>
        
        <!-- Error Message -->
        <div x-show="hasError" x-transition class="text-red-500 text-sm mt-1" x-text="errorMessage"></div>
        
        <!-- Server-side Error (from initial load) -->
        if props.Error != "" {
            <div class="text-red-500 text-sm mt-1">{props.Error}</div>
        }
    </div>
    
    <script>
        function fieldValidator() {
            return {
                value: '{props.Value}',
                isValid: false,
                hasError: {props.Error != ""},
                errorMessage: '{props.Error}',
                validating: false,
                
                async validateField() {
                    if (!this.value.trim() && !{props.Required}) {
                        this.hasError = false;
                        this.isValid = false;
                        return;
                    }
                    
                    this.validating = true;
                    
                    try {
                        const formData = new FormData();
                        formData.append('{props.Name}', this.value);
                        
                        const response = await fetch('{props.ValidationEndpoint}', {
                            method: 'POST',
                            body: formData,
                        });
                        
                        const result = await response.json();
                        
                        this.hasError = !result.valid;
                        this.isValid = result.valid;
                        this.errorMessage = result.error || '';
                        
                        // Update parent form validity
                        this.$dispatch('field-validated', {
                            field: '{props.Name}',
                            valid: result.valid
                        });
                        
                    } catch (error) {
                        console.error('Validation error:', error);
                        this.hasError = true;
                        this.errorMessage = 'Validation failed. Please try again.';
                    } finally {
                        this.validating = false;
                    }
                }
            }
        }
    </script>
}
```

## 4. Alpine.js Form Handler

```javascript
// Form handler with validation state management
function userFormHandler() {
    return {
        submitting: false,
        fieldValidation: {},
        isValid: false,
        
        init() {
            // Listen for field validation events
            this.$el.addEventListener('field-validated', (event) => {
                this.fieldValidation[event.detail.field] = event.detail.valid;
                this.updateFormValidity();
            });
            
            // HTMX event listeners
            this.$el.addEventListener('htmx:beforeRequest', () => {
                this.submitting = true;
            });
            
            this.$el.addEventListener('htmx:afterRequest', (event) => {
                this.submitting = false;
                
                if (event.detail.successful) {
                    // Form submitted successfully
                    this.$dispatch('user-created');
                    closeModal();
                } else {
                    // Handle validation errors from server
                    this.handleServerErrors(event.detail.xhr);
                }
            });
        },
        
        updateFormValidity() {
            // Check if all required fields are valid
            const requiredFields = ['first_name', 'last_name', 'email', 'role'];
            const optionalFields = ['phone', 'salary'];
            
            let valid = true;
            
            // Check required fields
            for (const field of requiredFields) {
                if (this.fieldValidation[field] !== true) {
                    valid = false;
                    break;
                }
            }
            
            // Check optional fields (if they have values, they must be valid)
            for (const field of optionalFields) {
                const input = this.$el.querySelector(`[name="${field}"]`);
                if (input && input.value.trim() !== '' && this.fieldValidation[field] !== true) {
                    valid = false;
                    break;
                }
            }
            
            this.isValid = valid;
        },
        
        handleSubmit(event) {
            // Additional client-side validation before submission
            if (!this.isValid) {
                event.preventDefault();
                this.showValidationSummary();
            }
        },
        
        handleServerErrors(xhr) {
            try {
                const response = JSON.parse(xhr.responseText);
                if (response.errors) {
                    // Display server validation errors
                    response.errors.forEach(error => {
                        const field = this.$el.querySelector(`[name="${error.field}"]`);
                        if (field) {
                            // Update field error state
                            const fieldComponent = Alpine.closestDataStack(field);
                            if (fieldComponent) {
                                fieldComponent.hasError = true;
                                fieldComponent.errorMessage = error.message;
                            }
                        }
                    });
                }
            } catch (e) {
                console.error('Error parsing server response:', e);
            }
        },
        
        showValidationSummary() {
            // Show a summary of validation errors
            const invalidFields = Object.entries(this.fieldValidation)
                .filter(([field, valid]) => !valid)
                .map(([field]) => field);
                
            if (invalidFields.length > 0) {
                Alpine.store('app').addNotification('error', 
                    'Please correct the errors in the following fields: ' + 
                    invalidFields.join(', ')
                );
            }
        }
    }
}
```

## 5. Real-time Validation Features

### Dependent Field Validation

```javascript
// For fields that depend on other fields
function dependentFieldValidator(dependsOn, validationEndpoint) {
    return {
        ...fieldValidator(),
        dependentField: dependsOn,
        
        init() {
            // Watch for changes in dependent field
            this.$watch(`$store.form.${this.dependentField}`, () => {
                if (this.value) {
                    this.validateField();
                }
            });
        },
        
        async validateField() {
            const dependentValue = this.$store.form[this.dependentField];
            
            if (!dependentValue) {
                this.hasError = true;
                this.errorMessage = `Please select ${this.dependentField} first`;
                return;
            }
            
            // Include dependent field value in validation
            const formData = new FormData();
            formData.append(this.fieldName, this.value);
            formData.append(this.dependentField, dependentValue);
            
            // Continue with normal validation...
        }
    }
}
```

### File Upload Validation

```go
// File upload validation component
templ ValidatedFileInput(props ValidatedInputProps) {
    <div class="mb-4" x-data="fileValidator()">
        <label class="block text-sm font-medium text-gray-700 mb-2">
            {props.Label}
            if props.Required {
                <span class="text-red-500">*</span>
            }
        </label>
        
        <div class="relative">
            <input 
                type="file"
                name={props.Name}
                if props.Required {
                    required
                }
                accept=".jpg,.jpeg,.png,.pdf,.doc,.docx"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                @change="validateFile($event)">
        </div>
        
        <!-- File validation messages -->
        <div x-show="hasError" x-transition class="text-red-500 text-sm mt-1" x-text="errorMessage"></div>
        <div x-show="fileInfo" class="text-sm text-gray-600 mt-1" x-text="fileInfo"></div>
    </div>
    
    <script>
        function fileValidator() {
            return {
                hasError: false,
                errorMessage: '',
                fileInfo: '',
                
                validateFile(event) {
                    const file = event.target.files[0];
                    if (!file) return;
                    
                    // File size validation (5MB limit)
                    const maxSize = 5 * 1024 * 1024;
                    if (file.size > maxSize) {
                        this.hasError = true;
                        this.errorMessage = 'File size must be less than 5MB';
                        return;
                    }
                    
                    // File type validation
                    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'application/pdf', 
                                        'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];
                    if (!allowedTypes.includes(file.type)) {
                        this.hasError = true;
                        this.errorMessage = 'Please select a valid file type (JPG, PNG, PDF, DOC, DOCX)';
                        return;
                    }
                    
                    // Success
                    this.hasError = false;
                    this.fileInfo = `Selected: ${file.name} (${(file.size / 1024).toFixed(1)} KB)`;
                }
            }
        }
    </script>
}
```

### **Cascading Select Validation**

```go
// templates/components/cascading_select.templ
templ CascadingSelectInput(props CascadingSelectProps) {
    <div class="mb-4" x-data="cascadingSelect('{props.DependentEndpoint}', '{props.ParentField}')">
        <label class="block text-sm font-medium text-gray-700 mb-2">
            {props.Label}
            if props.Required {
                <span class="text-red-500">*</span>
            }
        </label>
        
        <div class="relative">
            <select 
                name={props.Name}
                if props.Required {
                    required
                }
                :disabled="loading || options.length === 0"
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
                x-model="selectedValue">
                
                <option value="">
                    <span x-show="loading">Loading...</span>
                    <span x-show="!loading && options.length === 0">Select {props.ParentLabel} first</span>
                    <span x-show="!loading && options.length > 0">Choose {props.Label}</span>
                </option>
                
                <template x-for="option in options" :key="option.value">
                    <option :value="option.value" x-text="option.label"></option>
                </template>
            </select>
            
            <!-- Loading spinner -->
            <div x-show="loading" class="absolute inset-y-0 right-0 flex items-center pr-3">
                <svg class="animate-spin h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            </div>
        </div>
        
        <script>
            function cascadingSelect(endpoint, parentField) {
                return {
                    options: [],
                    loading: false,
                    selectedValue: '',
                    
                    init() {
                        // Watch parent field changes
                        this.$watch(`$store.form.${parentField}`, (value) => {
                            this.loadOptions(value);
                        });
                    },
                    
                    async loadOptions(parentValue) {
                        if (!parentValue) {
                            this.options = [];
                            this.selectedValue = '';
                            return;
                        }
                        
                        this.loading = true;
                        this.selectedValue = '';
                        
                        try {
                            const response = await fetch(`${endpoint}?parent=${encodeURIComponent(parentValue)}`);
                            if (response.ok) {
                                this.options = await response.json();
                            } else {
                                throw new Error('Failed to load options');
                            }
                        } catch (error) {
                            console.error('Error loading options:', error);
                            this.options = [];
                            Alpine.store('app').addNotification('error', 'Failed to load options');
                        } finally {
                            this.loading = false;
                        }
                    }
                }
            }
        </script>
    </div>
}
```

### **Multi-step Form with Validation**

```go
// templates/forms/multi_step_form.templ
templ MultiStepForm(currentStep int, form interface{}, errors validation.ValidationErrors) {
    <div x-data="multiStepFormHandler({currentStep})" class="max-w-4xl mx-auto">
        <!-- Progress Bar -->
        <div class="mb-8">
            <div class="flex items-center justify-between">
                <template x-for="(step, index) in steps" :key="index">
                    <div class="flex items-center">
                        <div 
                            :class="{
                                'bg-blue-600 text-white': index < currentStep,
                                'bg-blue-600 text-white': index === currentStep,
                                'bg-gray-200 text-gray-400': index > currentStep
                            }"
                            class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium">
                            <span x-show="index < currentStep">✓</span>
                            <span x-show="index >= currentStep" x-text="index + 1"></span>
                        </div>
                        <span x-show="index < steps.length - 1" class="ml-2 mr-2 h-0.5 bg-gray-200 flex-1"></span>
                    </div>
                </template>
            </div>
            <div class="mt-2 text-center">
                <span class="text-sm text-gray-600" x-text="steps[currentStep]?.title"></span>
            </div>
        </div>
        
        <!-- Form Steps -->
        <form hx-post="/users/create" hx-target="this" x-show="currentStep === 0">
            <div class="bg-white p-6 rounded-lg shadow">
                <h3 class="text-lg font-medium mb-4">Personal Information</h3>
                <!-- Step 1 fields -->
                @ValidatedTextInput(ValidatedInputProps{...})
                @ValidatedEmailInput(ValidatedInputProps{...})
                
                <div class="flex justify-end mt-6">
                    <button type="button" @click="nextStep()" class="bg-blue-600 text-white px-4 py-2 rounded-md">
                        Next Step
                    </button>
                </div>
            </div>
        </form>
        
        <form hx-post="/users/create" hx-target="this" x-show="currentStep === 1">
            <div class="bg-white p-6 rounded-lg shadow">
                <h3 class="text-lg font-medium mb-4">Employment Details</h3>
                <!-- Step 2 fields -->
                
                <div class="flex justify-between mt-6">
                    <button type="button" @click="previousStep()" class="bg-gray-300 text-gray-700 px-4 py-2 rounded-md">
                        Previous
                    </button>
                    <button type="button" @click="nextStep()" class="bg-blue-600 text-white px-4 py-2 rounded-md">
                        Next Step
                    </button>
                </div>
            </div>
        </form>
        
        <form hx-post="/users/create" hx-target="this" x-show="currentStep === 2">
            <div class="bg-white p-6 rounded-lg shadow">
                <h3 class="text-lg font-medium mb-4">Review & Submit</h3>
                <!-- Review step with summary -->
                
                <div class="flex justify-between mt-6">
                    <button type="button" @click="previousStep()" class="bg-gray-300 text-gray-700 px-4 py-2 rounded-md">
                        Previous
                    </button>
                    <button type="submit" class="bg-green-600 text-white px-4 py-2 rounded-md">
                        Create User
                    </button>
                </div>
            </div>
        </form>
    </div>
    
    <script>
        function multiStepFormHandler(initialStep = 0) {
            return {
                currentStep: initialStep,
                steps: [
                    { title: 'Personal Information', fields: ['first_name', 'last_name', 'email'] },
                    { title: 'Employment Details', fields: ['role', 'department', 'salary'] },
                    { title: 'Review & Submit', fields: [] }
                ],
                stepValidation: {},
                
                nextStep() {
                    if (this.validateCurrentStep()) {
                        this.currentStep = Math.min(this.currentStep + 1, this.steps.length - 1);
                    }
                },
                
                previousStep() {
                    this.currentStep = Math.max(this.currentStep - 1, 0);
                },
                
                validateCurrentStep() {
                    const currentStepFields = this.steps[this.currentStep].fields;
                    let isValid = true;
                    
                    currentStepFields.forEach(field => {
                        const input = document.querySelector(`[name="${field}"]`);
                        if (input && !input.checkValidity()) {
                            input.reportValidity();
                            isValid = false;
                        }
                    });
                    
                    return isValid;
                }
            }
        }
    </script>
}
```

### **Bulk Operations with Validation**

```go
// handlers/bulk_operations.go
func (h *Handler) BulkUpdateUsers(w http.ResponseWriter, r *http.Request) {
    var request struct {
        IDs     []int                  `json:"ids"`
        Updates map[string]interface{} `json:"updates"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Validate bulk update data
    validator := validation.NewValidator()
    
    for field, value := range request.Updates {
        switch field {
        case "status":
            if str, ok := value.(string); ok {
                validator.OneOf(field, str, []string{"active", "inactive", "pending"})
            }
        case "department":
            if str, ok := value.(string); ok {
                validator.Required(field, str)
            }
        }
    }
    
    if !validator.IsValid() {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "errors": validator.Errors(),
        })
        return
    }
    
    // Perform bulk update
    updatedUsers, err := h.db.BulkUpdateUsers(request.IDs, request.Updates)
    if err != nil {
        http.Error(w, "Bulk update failed", http.StatusInternalServerError)
        return
    }
    
    // Return updated table
    w.Header().Set("HX-Trigger", `{"bulkUpdateComplete": true, "showNotification": {"type": "success", "message": "Users updated successfully"}}`)
    tmpl := templates.UserTableRows(updatedUsers)
    tmpl.Render(r.Context(), w)
}
```

### **Custom Validation Messages**

```go
// validation/messages.go
package validation

var ValidationMessages = map[string]map[string]string{
    "en": {
        "required":           "This field is required",
        "min_length":         "Must be at least %d characters long",
        "max_length":         "Must be no more than %d characters long",
        "invalid_email":      "Please enter a valid email address",
        "invalid_phone":      "Please enter a valid phone number",
        "email_taken":        "This email address is already in use",
        "password_mismatch":  "Passwords do not match",
        "invalid_date":       "Please enter a valid date",
        "future_date_only":   "Date must be in the future",
        "business_hours":     "Time must be within business hours (9 AM - 6 PM)",
    },
    "es": {
        "required":           "Este campo es obligatorio",
        "invalid_email":      "Por favor ingrese un email válido",
        // ... more translations
    },
}

func (v *Validator) AddLocalizedError(field, code, lang string, args ...interface{}) {
    messages, exists := ValidationMessages[lang]
    if !exists {
        messages = ValidationMessages["en"] // fallback to English
    }
    
    template, exists := messages[code]
    if !exists {
        template = "Validation error"
    }
    
    message := fmt.Sprintf(template, args...)
    v.AddError(field, message, code)
}
```

## **Production Considerations:**

### **Rate Limiting for Validation Endpoints:**
```go
func (h *Handler) rateLimitedValidation(next http.HandlerFunc) http.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(10), 20) // 10 requests per second, burst of 20
    
    return func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too many validation requests", http.StatusTooManyRequests)
            return
        }
        next(w, r)
    }
}
```

### **Caching Validation Results:**
```go
// Cache validation results for expensive operations
func (v *Validator) UniqueEmailCached(field, email string, cache *Cache, db *Database, excludeID int) *Validator {
    if email == "" {
        return v
    }
    
    cacheKey := fmt.Sprintf("email_exists:%s:%d", email, excludeID)
    
    if exists, found := cache.Get(cacheKey); found {
        if exists.(bool) {
            v.AddError(field, "This email address is already in use", "email_taken")
        }
        return v
    }
    
    exists, err := db.EmailExists(email, excludeID)
    if err != nil {
        v.AddError(field, "Unable to verify email uniqueness", "validation_error")
    } else {
        cache.Set(cacheKey, exists, 5*time.Minute) // Cache for 5 minutes
        if exists {
            v.AddError(field, "This email address is already in use", "email_taken")
        }
    }
    
    return v
}
```

## Key Benefits

1. **Real-time Validation**: Users get immediate feedback as they type
2. **Server-side Security**: All validation is also enforced on the server
3. **Progressive Enhancement**: Forms work even without JavaScript
4. **User Experience**: Clear visual indicators and helpful error messages
5. **Reusable Components**: Validation logic can be applied to any form field
6. **Type Safety**: Templ ensures type-safe template rendering
7. **Scalable Architecture**: Easy to extend with new validation rules
8. **Internationalization**: Support for multiple languages
9. **Performance Optimized**: Caching and rate limiting for production use
10. **Advanced UX**: Multi-step forms, cascading selects, and bulk operations

This system provides comprehensive form validation that enhances user experience while maintaining security and data integrity across complex ERP workflows.
