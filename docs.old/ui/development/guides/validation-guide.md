# Complete Validation System Implementation Guide

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Complete implementation guide for form validation system
**SCOPE**: Backend validation, HTMX endpoints, Alpine.js integration, production considerations
**TARGET AUDIENCE**: Backend developers, full-stack developers
**RELATED FILES**: `components/forms.md` (form components), `patterns/htmx-integration.md` (server patterns)
<!-- LLM-CONTEXT-END -->

## System Architecture Overview

This guide provides complete implementation details for a three-layer validation system designed for ERP applications using **Go + Templ + HTMX + Alpine.js + Flowbite**.

<!-- LLM-VALIDATION-LAYERS-START -->
**VALIDATION ARCHITECTURE:**
```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENT LAYER (Alpine.js)                │
│  • Immediate format validation                             │
│  • User experience enhancement                             │
│  • Input formatting and masking                           │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                 REAL-TIME LAYER (HTMX)                     │
│  • Debounced server validation                             │
│  • Business rule checking                                  │
│  • Database uniqueness verification                        │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                SUBMISSION LAYER (Go Backend)                │
│  • Comprehensive validation                                │
│  • Transaction integrity                                   │
│  • Security enforcement                                    │
│  • Audit logging                                          │
└─────────────────────────────────────────────────────────────┘
```
<!-- LLM-VALIDATION-LAYERS-END -->

---

## Go Backend Implementation

### Core Validation Types

<!-- LLM-GO-VALIDATION-TYPES-START -->
```go
// validation/types.go
package validation

import (
    "fmt"
    "regexp"
    "strings"
    "time"
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

func (ve ValidationErrors) GetFieldErrors(field string) []string {
    var errors []string
    for _, err := range ve {
        if err.Field == field {
            errors = append(errors, err.Message)
        }
    }
    return errors
}

type Validator struct {
    errors ValidationErrors
    lang   string // For internationalization
}

func NewValidator() *Validator {
    return &Validator{
        errors: make(ValidationErrors, 0),
        lang:   "en",
    }
}

func NewValidatorWithLang(lang string) *Validator {
    return &Validator{
        errors: make(ValidationErrors, 0),
        lang:   lang,
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

func (v *Validator) ErrorCount() int {
    return len(v.errors)
}

func (v *Validator) HasErrorsForField(field string) bool {
    return v.errors.HasField(field)
}
```
<!-- LLM-GO-VALIDATION-TYPES-END -->

### Comprehensive Validation Rules

<!-- LLM-VALIDATION-RULES-START -->
```go
// validation/rules.go
package validation

import (
    "net/url"
    "regexp"
    "strconv"
    "strings"
    "time"
    "unicode"
)

// String validation rules
func (v *Validator) Required(field, value string) *Validator {
    if strings.TrimSpace(value) == "" {
        v.AddError(field, v.getMessage("required"), "required")
    }
    return v
}

func (v *Validator) MinLength(field, value string, min int) *Validator {
    if len(strings.TrimSpace(value)) < min {
        v.AddError(field, v.getMessagef("min_length", min), "min_length")
    }
    return v
}

func (v *Validator) MaxLength(field, value string, max int) *Validator {
    if len(value) > max {
        v.AddError(field, v.getMessagef("max_length", max), "max_length")
    }
    return v
}

func (v *Validator) LengthRange(field, value string, min, max int) *Validator {
    length := len(strings.TrimSpace(value))
    if length < min || length > max {
        v.AddError(field, v.getMessagef("length_range", min, max), "length_range")
    }
    return v
}

// Format validation rules
func (v *Validator) Email(field, value string) *Validator {
    if value != "" {
        emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
        if !emailRegex.MatchString(value) {
            v.AddError(field, v.getMessage("invalid_email"), "invalid_email")
        }
    }
    return v
}

func (v *Validator) Phone(field, value string) *Validator {
    if value != "" {
        // Remove spaces, dashes, parentheses for validation
        cleaned := regexp.MustCompile(`[\s\-\(\)]+`).ReplaceAllString(value, "")
        phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
        if !phoneRegex.MatchString(cleaned) {
            v.AddError(field, v.getMessage("invalid_phone"), "invalid_phone")
        }
    }
    return v
}

func (v *Validator) URL(field, value string) *Validator {
    if value != "" {
        if _, err := url.ParseRequestURI(value); err != nil {
            v.AddError(field, v.getMessage("invalid_url"), "invalid_url")
        }
    }
    return v
}

func (v *Validator) Numeric(field, value string) *Validator {
    if value != "" {
        if _, err := strconv.ParseFloat(value, 64); err != nil {
            v.AddError(field, v.getMessage("invalid_number"), "invalid_number")
        }
    }
    return v
}

func (v *Validator) Integer(field, value string) *Validator {
    if value != "" {
        if _, err := strconv.Atoi(value); err != nil {
            v.AddError(field, v.getMessage("invalid_integer"), "invalid_integer")
        }
    }
    return v
}

func (v *Validator) Date(field, value string) *Validator {
    if value != "" {
        formats := []string{
            "2006-01-02",
            "01/02/2006",
            "02-01-2006",
            "2006-01-02 15:04:05",
        }
        
        valid := false
        for _, format := range formats {
            if _, err := time.Parse(format, value); err == nil {
                valid = true
                break
            }
        }
        
        if !valid {
            v.AddError(field, v.getMessage("invalid_date"), "invalid_date")
        }
    }
    return v
}

// Numeric range validation
func (v *Validator) MinValue(field, value string, min float64) *Validator {
    if value != "" {
        if val, err := strconv.ParseFloat(value, 64); err == nil {
            if val < min {
                v.AddError(field, v.getMessagef("min_value", min), "min_value")
            }
        }
    }
    return v
}

func (v *Validator) MaxValue(field, value string, max float64) *Validator {
    if value != "" {
        if val, err := strconv.ParseFloat(value, 64); err == nil {
            if val > max {
                v.AddError(field, v.getMessagef("max_value", max), "max_value")
            }
        }
    }
    return v
}

func (v *Validator) Range(field, value string, min, max float64) *Validator {
    if value != "" {
        if val, err := strconv.ParseFloat(value, 64); err == nil {
            if val < min || val > max {
                v.AddError(field, v.getMessagef("value_range", min, max), "value_range")
            }
        }
    }
    return v
}

// Pattern validation
func (v *Validator) Pattern(field, value, pattern, message string) *Validator {
    if value != "" {
        if matched, err := regexp.MatchString(pattern, value); err != nil || !matched {
            v.AddError(field, message, "pattern_mismatch")
        }
    }
    return v
}

func (v *Validator) AlphaNumeric(field, value string) *Validator {
    if value != "" {
        for _, char := range value {
            if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
                v.AddError(field, v.getMessage("invalid_alphanumeric"), "invalid_alphanumeric")
                break
            }
        }
    }
    return v
}

func (v *Validator) Alpha(field, value string) *Validator {
    if value != "" {
        for _, char := range value {
            if !unicode.IsLetter(char) && !unicode.IsSpace(char) {
                v.AddError(field, v.getMessage("invalid_alpha"), "invalid_alpha")
                break
            }
        }
    }
    return v
}

// Choice validation
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
            v.AddError(field, v.getMessage("invalid_option"), "invalid_option")
        }
    }
    return v
}

// Password validation
func (v *Validator) Password(field, value string, minLength int, requireSymbol, requireNumber bool) *Validator {
    if value == "" {
        return v
    }
    
    if len(value) < minLength {
        v.AddError(field, v.getMessagef("password_min_length", minLength), "password_weak")
    }
    
    if requireNumber {
        hasNumber := regexp.MustCompile(`\d`).MatchString(value)
        if !hasNumber {
            v.AddError(field, v.getMessage("password_need_number"), "password_weak")
        }
    }
    
    if requireSymbol {
        hasSymbol := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(value)
        if !hasSymbol {
            v.AddError(field, v.getMessage("password_need_symbol"), "password_weak")
        }
    }
    
    return v
}

func (v *Validator) PasswordConfirmation(field, password, confirmation string) *Validator {
    if password != confirmation {
        v.AddError(field, v.getMessage("password_mismatch"), "password_mismatch")
    }
    return v
}

// Cross-field validation
func (v *Validator) FieldsMatch(field1, value1, field2, value2 string) *Validator {
    if value1 != value2 {
        v.AddError(field2, v.getMessagef("fields_must_match", field1), "fields_mismatch")
    }
    return v
}

func (v *Validator) DateAfter(field, dateValue, afterField, afterValue string) *Validator {
    if dateValue != "" && afterValue != "" {
        date1, err1 := time.Parse("2006-01-02", dateValue)
        date2, err2 := time.Parse("2006-01-02", afterValue)
        
        if err1 == nil && err2 == nil && !date1.After(date2) {
            v.AddError(field, v.getMessagef("date_after", afterField), "date_comparison")
        }
    }
    return v
}

func (v *Validator) DateBefore(field, dateValue, beforeField, beforeValue string) *Validator {
    if dateValue != "" && beforeValue != "" {
        date1, err1 := time.Parse("2006-01-02", dateValue)
        date2, err2 := time.Parse("2006-01-02", beforeValue)
        
        if err1 == nil && err2 == nil && !date1.Before(date2) {
            v.AddError(field, v.getMessagef("date_before", beforeField), "date_comparison")
        }
    }
    return v
}
```
<!-- LLM-VALIDATION-RULES-END -->

### Database Validation

<!-- LLM-DATABASE-VALIDATION-START -->
```go
// validation/database.go
package validation

import (
    "context"
    "database/sql"
    "fmt"
)

type DatabaseValidator struct {
    db *sql.DB
}

func NewDatabaseValidator(db *sql.DB) *DatabaseValidator {
    return &DatabaseValidator{db: db}
}

// Uniqueness validation
func (v *Validator) UniqueEmail(field, email string, dbValidator *DatabaseValidator, excludeID int) *Validator {
    if email != "" {
        exists, err := dbValidator.EmailExists(context.Background(), email, excludeID)
        if err != nil {
            v.AddError(field, v.getMessage("validation_error"), "validation_error")
        } else if exists {
            v.AddError(field, v.getMessage("email_taken"), "email_taken")
        }
    }
    return v
}

func (v *Validator) UniqueUsername(field, username string, dbValidator *DatabaseValidator, excludeID int) *Validator {
    if username != "" {
        exists, err := dbValidator.UsernameExists(context.Background(), username, excludeID)
        if err != nil {
            v.AddError(field, v.getMessage("validation_error"), "validation_error")
        } else if exists {
            v.AddError(field, v.getMessage("username_taken"), "username_taken")
        }
    }
    return v
}

func (v *Validator) RecordExists(field, tableName, idField, id string, dbValidator *DatabaseValidator) *Validator {
    if id != "" {
        exists, err := dbValidator.RecordExists(context.Background(), tableName, idField, id)
        if err != nil {
            v.AddError(field, v.getMessage("validation_error"), "validation_error")
        } else if !exists {
            v.AddError(field, v.getMessagef("record_not_found", tableName), "record_not_found")
        }
    }
    return v
}

// Database validation methods
func (dv *DatabaseValidator) EmailExists(ctx context.Context, email string, excludeID int) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`
    
    var exists bool
    err := dv.db.QueryRowContext(ctx, query, email, excludeID).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("error checking email uniqueness: %w", err)
    }
    
    return exists, nil
}

func (dv *DatabaseValidator) UsernameExists(ctx context.Context, username string, excludeID int) (bool, error) {
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2)`
    
    var exists bool
    err := dv.db.QueryRowContext(ctx, query, username, excludeID).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("error checking username uniqueness: %w", err)
    }
    
    return exists, nil
}

func (dv *DatabaseValidator) RecordExists(ctx context.Context, tableName, idField, id string) (bool, error) {
    query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)`, tableName, idField)
    
    var exists bool
    err := dv.db.QueryRowContext(ctx, query, id).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("error checking record existence: %w", err)
    }
    
    return exists, nil
}

// Business rule validation
func (v *Validator) ValidateBusinessRule(field string, rule func() (bool, string)) *Validator {
    if valid, message := rule(); !valid {
        v.AddError(field, message, "business_rule")
    }
    return v
}

// Example business rules
func (dv *DatabaseValidator) CanAssignRole(ctx context.Context, userID int, roleID int) (bool, string) {
    // Check if user has permission to assign this role
    query := `
        SELECT EXISTS(
            SELECT 1 FROM user_permissions up
            JOIN role_permissions rp ON up.permission_id = rp.permission_id
            WHERE up.user_id = $1 AND rp.role_id = $2 AND rp.permission_name = 'assign_role'
        )`
    
    var canAssign bool
    err := dv.db.QueryRowContext(ctx, query, userID, roleID).Scan(&canAssign)
    if err != nil {
        return false, "Unable to verify permissions"
    }
    
    if !canAssign {
        return false, "You don't have permission to assign this role"
    }
    
    return true, ""
}

func (dv *DatabaseValidator) ValidateWorkflowTransition(ctx context.Context, recordID int, fromState, toState string) (bool, string) {
    // Check if state transition is allowed
    query := `
        SELECT EXISTS(
            SELECT 1 FROM workflow_transitions 
            WHERE from_state = $1 AND to_state = $2 AND active = true
        )`
    
    var allowed bool
    err := dv.db.QueryRowContext(ctx, query, fromState, toState).Scan(&allowed)
    if err != nil {
        return false, "Unable to verify workflow transition"
    }
    
    if !allowed {
        return false, fmt.Sprintf("Transition from %s to %s is not allowed", fromState, toState)
    }
    
    return true, ""
}
```
<!-- LLM-DATABASE-VALIDATION-END -->

### Internationalization Support

<!-- LLM-I18N-VALIDATION-START -->
```go
// validation/messages.go
package validation

import "fmt"

var ValidationMessages = map[string]map[string]string{
    "en": {
        "required":             "This field is required",
        "min_length":           "Must be at least %d characters long",
        "max_length":           "Must be no more than %d characters long",
        "length_range":         "Must be between %d and %d characters long",
        "invalid_email":        "Please enter a valid email address",
        "invalid_phone":        "Please enter a valid phone number",
        "invalid_url":          "Please enter a valid URL",
        "invalid_number":       "Must be a valid number",
        "invalid_integer":      "Must be a valid whole number",
        "invalid_date":         "Please enter a valid date",
        "invalid_alphanumeric": "Must contain only letters and numbers",
        "invalid_alpha":        "Must contain only letters",
        "invalid_option":       "Please select a valid option",
        "min_value":            "Must be at least %g",
        "max_value":            "Must be no more than %g",
        "value_range":          "Must be between %g and %g",
        "email_taken":          "This email address is already in use",
        "username_taken":       "This username is already in use",
        "password_min_length":  "Password must be at least %d characters long",
        "password_need_number": "Password must contain at least one number",
        "password_need_symbol": "Password must contain at least one symbol",
        "password_mismatch":    "Passwords do not match",
        "fields_must_match":    "Must match %s",
        "date_after":           "Must be after %s",
        "date_before":          "Must be before %s",
        "record_not_found":     "%s not found",
        "validation_error":     "Validation failed. Please try again.",
    },
    "es": {
        "required":        "Este campo es obligatorio",
        "invalid_email":   "Por favor ingrese un email válido",
        "email_taken":     "Este email ya está en uso",
        // ... more translations
    },
    "fr": {
        "required":        "Ce champ est requis",
        "invalid_email":   "Veuillez saisir une adresse email valide",
        "email_taken":     "Cette adresse email est déjà utilisée",
        // ... more translations
    },
}

func (v *Validator) getMessage(key string) string {
    if messages, exists := ValidationMessages[v.lang]; exists {
        if message, exists := messages[key]; exists {
            return message
        }
    }
    
    // Fallback to English
    if messages, exists := ValidationMessages["en"]; exists {
        if message, exists := messages[key]; exists {
            return message
        }
    }
    
    return "Validation error"
}

func (v *Validator) getMessagef(key string, args ...interface{}) string {
    template := v.getMessage(key)
    return fmt.Sprintf(template, args...)
}

func (v *Validator) SetLanguage(lang string) {
    v.lang = lang
}
```
<!-- LLM-I18N-VALIDATION-END -->

---

## HTMX Validation Endpoints

### Real-time Validation Handler

<!-- LLM-HTMX-HANDLERS-START -->
```go
// handlers/validation.go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "time"
    
    "your-app/validation"
    "your-app/forms"
)

type ValidationHandler struct {
    validator *validation.DatabaseValidator
    cache     *ValidationCache
    rateLimiter *RateLimiter
}

func NewValidationHandler(db *sql.DB) *ValidationHandler {
    return &ValidationHandler{
        validator: validation.NewDatabaseValidator(db),
        cache:     NewValidationCache(),
        rateLimiter: NewRateLimiter(),
    }
}

// Real-time field validation endpoint
func (h *ValidationHandler) ValidateField(w http.ResponseWriter, r *http.Request) {
    // Rate limiting
    clientIP := getClientIP(r)
    if !h.rateLimiter.Allow(clientIP) {
        http.Error(w, "Too many requests", http.StatusTooManyRequests)
        return
    }
    
    field := r.URL.Query().Get("field")
    value := r.FormValue(field)
    checkType := r.URL.Query().Get("check") // unique, exists, etc.
    excludeID := r.URL.Query().Get("exclude_id")
    
    validator := validation.NewValidator()
    
    // Apply validation rules based on field name
    switch field {
    case "email":
        validator.Required(field, value).Email(field, value)
        if validator.IsValid() && checkType == "unique" {
            excludeIDInt := 0
            if excludeID != "" {
                excludeIDInt, _ = strconv.Atoi(excludeID)
            }
            validator.UniqueEmail(field, value, h.validator, excludeIDInt)
        }
        
    case "first_name", "last_name":
        validator.Required(field, value).
                MinLength(field, value, 2).
                MaxLength(field, value, 50).
                Alpha(field, value)
                
    case "username":
        validator.Required(field, value).
                MinLength(field, value, 3).
                MaxLength(field, value, 30).
                AlphaNumeric(field, value)
        if validator.IsValid() && checkType == "unique" {
            excludeIDInt := 0
            if excludeID != "" {
                excludeIDInt, _ = strconv.Atoi(excludeID)
            }
            validator.UniqueUsername(field, value, h.validator, excludeIDInt)
        }
        
    case "phone":
        if value != "" {
            validator.Phone(field, value)
        }
        
    case "password":
        if value != "" {
            validator.Password(field, value, 8, true, true)
        }
        
    case "password_confirmation":
        password := r.FormValue("password")
        validator.PasswordConfirmation(field, password, value)
        
    case "salary":
        if value != "" {
            validator.Numeric(field, value).MinValue(field, value, 0)
        }
        
    case "birth_date":
        if value != "" {
            validator.Date(field, value)
            // Additional business rule: must be at least 18 years old
            if validator.IsValid() {
                if birthDate, err := time.Parse("2006-01-02", value); err == nil {
                    eighteenYearsAgo := time.Now().AddDate(-18, 0, 0)
                    if birthDate.After(eighteenYearsAgo) {
                        validator.AddError(field, "Must be at least 18 years old", "age_requirement")
                    }
                }
            }
        }
        
    case "department_id":
        if value != "" {
            validator.RecordExists(field, "departments", "id", value, h.validator)
        }
        
    default:
        // Generic validation for unknown fields
        if value != "" {
            validator.MaxLength(field, value, 255)
        }
    }
    
    // Return JSON response
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
            "valid":  false,
            "field":  field,
            "error":  validator.Errors()[0].Message,
            "errors": validator.Errors(),
        })
    }
}

// Batch validation for multiple fields
func (h *ValidationHandler) ValidateFields(w http.ResponseWriter, r *http.Request) {
    var request struct {
        Fields map[string]string `json:"fields"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    validator := validation.NewValidator()
    results := make(map[string]interface{})
    
    for field, value := range request.Fields {
        fieldValidator := validation.NewValidator()
        h.validateSingleField(fieldValidator, field, value, r)
        
        results[field] = map[string]interface{}{
            "valid":  fieldValidator.IsValid(),
            "errors": fieldValidator.Errors(),
        }
        
        // Add to main validator for overall result
        for _, err := range fieldValidator.Errors() {
            validator.AddError(err.Field, err.Message, err.Code)
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "valid":   validator.IsValid(),
        "results": results,
    })
}

func (h *ValidationHandler) validateSingleField(validator *validation.Validator, field, value string, r *http.Request) {
    // Reuse validation logic from ValidateField
    // This is extracted to avoid code duplication
    switch field {
    case "email":
        validator.Required(field, value).Email(field, value)
        // Add uniqueness check if needed
    case "first_name", "last_name":
        validator.Required(field, value).MinLength(field, value, 2).MaxLength(field, value, 50)
    // ... other field validations
    }
}

// Helper functions
func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header first
    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        // Take the first IP in the list
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }
    
    // Check X-Real-IP header
    xri := r.Header.Get("X-Real-IP")
    if xri != "" {
        return xri
    }
    
    // Fall back to RemoteAddr
    ip := r.RemoteAddr
    if colon := strings.LastIndex(ip, ":"); colon != -1 {
        ip = ip[:colon]
    }
    return ip
}
```
<!-- LLM-HTMX-HANDLERS-END -->

### Form Submission Handlers

<!-- LLM-FORM-HANDLERS-START -->
```go
// handlers/forms.go
package handlers

import (
    "context"
    "net/http"
    "strconv"
    
    "your-app/forms"
    "your-app/templates"
    "your-app/validation"
)

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
    dbValidator := validation.NewDatabaseValidator(h.db)
    form.Validate(validator, dbValidator)
    
    // Additional business rule validation
    h.validateBusinessRules(&form, validator, r.Context())
    
    if !validator.IsValid() {
        // Return form with errors
        w.Header().Set("HX-Reswap", "outerHTML")
        w.WriteHeader(http.StatusUnprocessableEntity)
        
        tmpl := templates.CreateUserForm(form, validator.Errors())
        tmpl.Render(r.Context(), w)
        return
    }
    
    // Begin transaction
    tx, err := h.db.BeginTx(r.Context(), nil)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()
    
    // Create user
    user, err := h.userService.CreateUser(r.Context(), tx, form)
    if err != nil {
        // Log error and return user-friendly message
        h.logger.Error("Failed to create user", "error", err, "form", form)
        http.Error(w, "Failed to create user", http.StatusInternalServerError)
        return
    }
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        h.logger.Error("Failed to commit user creation", "error", err, "user_id", user.ID)
        http.Error(w, "Failed to save user", http.StatusInternalServerError)
        return
    }
    
    // Log audit event
    h.auditLogger.LogUserCreation(r.Context(), user.ID, getCurrentUser(r).ID)
    
    // Success response
    w.Header().Set("HX-Trigger", `{
        "userCreated": {"id": `+strconv.Itoa(user.ID)+`}, 
        "showNotification": {
            "type": "success", 
            "message": "User created successfully"
        },
        "closeModal": true
    }`)
    w.WriteHeader(http.StatusCreated)
    
    // Return updated table row or redirect
    tmpl := templates.UserTableRow(user)
    tmpl.Render(r.Context(), w)
}

func (h *Handler) validateBusinessRules(form *forms.UserForm, validator *validation.Validator, ctx context.Context) {
    currentUser := getCurrentUser(ctx)
    
    // Role assignment validation
    if form.Role != "" {
        validator.ValidateBusinessRule("role", func() (bool, string) {
            return h.userService.CanAssignRole(ctx, currentUser.ID, form.Role)
        })
    }
    
    // Department access validation
    if form.Department != "" {
        validator.ValidateBusinessRule("department", func() (bool, string) {
            return h.userService.CanAccessDepartment(ctx, currentUser.ID, form.Department)
        })
    }
    
    // Salary range validation based on role
    if form.Salary != "" && form.Role != "" {
        validator.ValidateBusinessRule("salary", func() (bool, string) {
            return h.userService.ValidateSalaryForRole(ctx, form.Role, form.Salary)
        })
    }
}

// Update user with partial validation
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    userID, err := strconv.Atoi(r.URL.Query().Get("id"))
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // Get existing user
    existingUser, err := h.userService.GetUser(r.Context(), userID)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    var form forms.UserForm
    
    // Parse only provided fields
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }
    
    // Start with existing values
    form = forms.UserFormFromUser(existingUser)
    
    // Update only provided fields
    if firstName := r.FormValue("first_name"); firstName != "" {
        form.FirstName = firstName
    }
    if lastName := r.FormValue("last_name"); lastName != "" {
        form.LastName = lastName
    }
    if email := r.FormValue("email"); email != "" {
        form.Email = email
    }
    // ... other fields
    
    // Validate with exclusion for current user
    validator := validation.NewValidator()
    dbValidator := validation.NewDatabaseValidator(h.db)
    form.ValidateForUpdate(validator, dbValidator, userID)
    
    if !validator.IsValid() {
        w.Header().Set("HX-Reswap", "outerHTML")
        w.WriteHeader(http.StatusUnprocessableEntity)
        
        tmpl := templates.EditUserForm(form, validator.Errors())
        tmpl.Render(r.Context(), w)
        return
    }
    
    // Update user
    updatedUser, err := h.userService.UpdateUser(r.Context(), userID, form)
    if err != nil {
        http.Error(w, "Failed to update user", http.StatusInternalServerError)
        return
    }
    
    // Success response
    w.Header().Set("HX-Trigger", `{
        "userUpdated": {"id": `+strconv.Itoa(updatedUser.ID)+`},
        "showNotification": {
            "type": "success",
            "message": "User updated successfully"
        }
    }`)
    
    tmpl := templates.UserTableRow(updatedUser)
    tmpl.Render(r.Context(), w)
}
```
<!-- LLM-FORM-HANDLERS-END -->

---

## Alpine.js Client Integration

### Field Validator Component

<!-- LLM-ALPINE-VALIDATOR-START -->
```javascript
// static/js/field-validator.js

// Base field validator
function fieldValidator(endpoint, fieldName, options = {}) {
    return {
        value: options.initialValue || '',
        isValid: false,
        hasError: options.initialError !== '',
        errorMessage: options.initialError || '',
        validating: false,
        touched: false,
        
        // Validation state
        validationTimeout: null,
        validationRequest: null,
        
        init() {
            // Set initial state
            this.value = this.$el.value || this.value;
            
            // Watch for external value changes
            this.$watch('value', (newValue) => {
                if (this.touched) {
                    this.scheduleValidation();
                }
            });
            
            // Mark as touched on first interaction
            this.$el.addEventListener('focus', () => {
                this.touched = true;
            });
            
            // Clear errors on focus
            this.$el.addEventListener('focus', () => {
                if (this.hasError) {
                    this.clearError();
                }
            });
        },
        
        scheduleValidation() {
            // Clear existing timeout
            if (this.validationTimeout) {
                clearTimeout(this.validationTimeout);
            }
            
            // Cancel previous request
            if (this.validationRequest) {
                this.validationRequest.abort();
            }
            
            // Schedule new validation
            this.validationTimeout = setTimeout(() => {
                this.validateField();
            }, options.debounceMs || 500);
        },
        
        async validateField() {
            if (!this.touched || !this.value.trim()) {
                this.isValid = false;
                this.hasError = false;
                this.errorMessage = '';
                return;
            }
            
            this.validating = true;
            
            try {
                // Create form data
                const formData = new FormData();
                formData.append(fieldName, this.value);
                
                // Include related field values for cross-field validation
                if (options.relatedFields) {
                    options.relatedFields.forEach(relatedField => {
                        const relatedInput = document.querySelector(`[name="${relatedField}"]`);
                        if (relatedInput) {
                            formData.append(relatedField, relatedInput.value);
                        }
                    });
                }
                
                // Create abort controller for this request
                const controller = new AbortController();
                this.validationRequest = controller;
                
                const response = await fetch(endpoint, {
                    method: 'POST',
                    body: formData,
                    signal: controller.signal,
                    headers: {
                        'X-Requested-With': 'XMLHttpRequest',
                    }
                });
                
                const result = await response.json();
                
                this.hasError = !result.valid;
                this.isValid = result.valid;
                this.errorMessage = result.error || '';
                
                // Dispatch validation event for form-level handling
                this.$dispatch('field-validated', {
                    field: fieldName,
                    valid: result.valid,
                    error: result.error,
                    value: this.value
                });
                
            } catch (error) {
                if (error.name !== 'AbortError') {
                    console.error('Validation error:', error);
                    this.hasError = true;
                    this.errorMessage = 'Validation failed. Please try again.';
                }
            } finally {
                this.validating = false;
                this.validationRequest = null;
            }
        },
        
        clearError() {
            this.hasError = false;
            this.errorMessage = '';
        },
        
        // Manual validation trigger
        validate() {
            this.touched = true;
            return this.validateField();
        },
        
        // Get current validation state
        getValidationState() {
            return {
                field: fieldName,
                value: this.value,
                valid: this.isValid,
                error: this.errorMessage,
                validating: this.validating
            };
        }
    };
}

// Specialized validators
function emailValidator(endpoint, options = {}) {
    return {
        ...fieldValidator(endpoint, 'email', options),
        
        // Client-side email format validation
        validateFormat() {
            const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
            if (this.value && !emailRegex.test(this.value)) {
                this.hasError = true;
                this.errorMessage = 'Please enter a valid email address';
                this.isValid = false;
                return false;
            }
            return true;
        },
        
        async validateField() {
            // First check format
            if (!this.validateFormat()) {
                this.validating = false;
                return;
            }
            
            // Then check server-side validation
            return await this.__proto__.validateField.call(this);
        }
    };
}

function passwordValidator(endpoint, options = {}) {
    return {
        ...fieldValidator(endpoint, 'password', options),
        
        strength: 'weak',
        requirements: {
            minLength: options.minLength || 8,
            hasNumber: options.requireNumber !== false,
            hasSymbol: options.requireSymbol !== false,
            hasUpper: options.requireUpper !== false,
            hasLower: options.requireLower !== false,
        },
        
        init() {
            this.__proto__.init.call(this);
            
            // Watch for password changes to update strength
            this.$watch('value', (newValue) => {
                this.updateStrength();
            });
        },
        
        updateStrength() {
            if (!this.value) {
                this.strength = 'none';
                return;
            }
            
            let score = 0;
            
            // Check requirements
            if (this.value.length >= this.requirements.minLength) score++;
            if (this.requirements.hasNumber && /\d/.test(this.value)) score++;
            if (this.requirements.hasSymbol && /[!@#$%^&*(),.?":{}|<>]/.test(this.value)) score++;
            if (this.requirements.hasUpper && /[A-Z]/.test(this.value)) score++;
            if (this.requirements.hasLower && /[a-z]/.test(this.value)) score++;
            
            // Determine strength
            if (score < 2) this.strength = 'weak';
            else if (score < 4) this.strength = 'medium';
            else this.strength = 'strong';
        },
        
        getStrengthColor() {
            switch (this.strength) {
                case 'weak': return 'text-red-500';
                case 'medium': return 'text-yellow-500';
                case 'strong': return 'text-green-500';
                default: return 'text-gray-400';
            }
        },
        
        getRequirementsMet() {
            return {
                minLength: this.value.length >= this.requirements.minLength,
                hasNumber: !this.requirements.hasNumber || /\d/.test(this.value),
                hasSymbol: !this.requirements.hasSymbol || /[!@#$%^&*(),.?":{}|<>]/.test(this.value),
                hasUpper: !this.requirements.hasUpper || /[A-Z]/.test(this.value),
                hasLower: !this.requirements.hasLower || /[a-z]/.test(this.value),
            };
        }
    };
}

function confirmationValidator(originalFieldName, options = {}) {
    return {
        value: '',
        hasError: false,
        errorMessage: '',
        touched: false,
        
        init() {
            this.value = this.$el.value || '';
            
            // Watch both fields
            this.$watch('value', () => {
                if (this.touched) {
                    this.validateConfirmation();
                }
            });
            
            // Watch original field
            const originalField = document.querySelector(`[name="${originalFieldName}"]`);
            if (originalField) {
                originalField.addEventListener('input', () => {
                    if (this.touched) {
                        this.validateConfirmation();
                    }
                });
            }
            
            this.$el.addEventListener('focus', () => {
                this.touched = true;
            });
        },
        
        validateConfirmation() {
            const originalField = document.querySelector(`[name="${originalFieldName}"]`);
            const originalValue = originalField ? originalField.value : '';
            
            if (this.value !== originalValue) {
                this.hasError = true;
                this.errorMessage = options.message || `Must match ${originalFieldName}`;
            } else {
                this.hasError = false;
                this.errorMessage = '';
            }
            
            this.$dispatch('field-validated', {
                field: this.$el.name,
                valid: !this.hasError,
                error: this.errorMessage,
                value: this.value
            });
        }
    };
}
```
<!-- LLM-ALPINE-VALIDATOR-END -->

### Form Manager Component

<!-- LLM-FORM-MANAGER-START -->
```javascript
// static/js/form-manager.js

function formManager(options = {}) {
    return {
        // Form state
        submitting: false,
        fieldValidation: {},
        touchedFields: new Set(),
        isValid: false,
        
        // Validation summary
        validationSummary: {
            totalFields: 0,
            validFields: 0,
            invalidFields: 0,
            errors: []
        },
        
        init() {
            // Get all validatable fields
            this.initializeFields();
            
            // Listen for field validation events
            this.$el.addEventListener('field-validated', (event) => {
                this.handleFieldValidation(event.detail);
            });
            
            // HTMX event listeners
            this.setupHTMXEventListeners();
            
            // Form submission handling
            this.$el.addEventListener('submit', (event) => {
                this.handleSubmit(event);
            });
        },
        
        initializeFields() {
            const fields = this.$el.querySelectorAll('[name]');
            fields.forEach(field => {
                const fieldName = field.name;
                this.fieldValidation[fieldName] = {
                    valid: false,
                    required: field.hasAttribute('required'),
                    touched: false,
                    error: ''
                };
            });
            
            this.updateValidationSummary();
        },
        
        handleFieldValidation(detail) {
            const { field, valid, error, value } = detail;
            
            // Update field validation state
            this.fieldValidation[field] = {
                ...this.fieldValidation[field],
                valid: valid,
                error: error || '',
                touched: true,
                value: value
            };
            
            this.touchedFields.add(field);
            this.updateFormValidity();
            this.updateValidationSummary();
            
            // Store validation state for persistence
            this.storeValidationState();
        },
        
        updateFormValidity() {
            // Check if all required fields are valid
            let allValid = true;
            
            for (const [fieldName, fieldState] of Object.entries(this.fieldValidation)) {
                if (fieldState.required) {
                    // Required field must be valid
                    if (!fieldState.valid) {
                        allValid = false;
                        break;
                    }
                } else {
                    // Optional field must be valid if it has a value
                    const input = this.$el.querySelector(`[name="${fieldName}"]`);
                    if (input && input.value.trim() !== '' && !fieldState.valid) {
                        allValid = false;
                        break;
                    }
                }
            }
            
            this.isValid = allValid;
        },
        
        updateValidationSummary() {
            const fields = Object.values(this.fieldValidation);
            const errors = [];
            
            this.validationSummary = {
                totalFields: fields.length,
                validFields: fields.filter(f => f.valid).length,
                invalidFields: fields.filter(f => !f.valid && f.touched).length,
                errors: fields.filter(f => f.error).map(f => f.error)
            };
        },
        
        setupHTMXEventListeners() {
            // Before request
            this.$el.addEventListener('htmx:beforeRequest', (event) => {
                this.submitting = true;
                this.clearServerErrors();
            });
            
            // After request
            this.$el.addEventListener('htmx:afterRequest', (event) => {
                this.submitting = false;
                
                if (event.detail.successful) {
                    this.handleSuccessfulSubmission(event);
                } else {
                    this.handleFailedSubmission(event);
                }
            });
            
            // Request error
            this.$el.addEventListener('htmx:responseError', (event) => {
                this.submitting = false;
                this.handleRequestError(event);
            });
        },
        
        handleSubmit(event) {
            // Validate all fields before submission
            this.validateAllFields();
            
            if (!this.isValid) {
                event.preventDefault();
                this.showValidationSummary();
                this.focusFirstInvalidField();
            }
        },
        
        async validateAllFields() {
            const promises = [];
            
            // Trigger validation for all fields
            this.$el.querySelectorAll('[x-data*="fieldValidator"]').forEach(field => {
                const alpineData = Alpine.$data(field);
                if (alpineData && typeof alpineData.validate === 'function') {
                    promises.push(alpineData.validate());
                }
            });
            
            // Wait for all validations to complete
            await Promise.all(promises);
        },
        
        focusFirstInvalidField() {
            for (const [fieldName, fieldState] of Object.entries(this.fieldValidation)) {
                if (!fieldState.valid && fieldState.required) {
                    const field = this.$el.querySelector(`[name="${fieldName}"]`);
                    if (field) {
                        field.focus();
                        field.scrollIntoView({ behavior: 'smooth', block: 'center' });
                        break;
                    }
                }
            }
        },
        
        handleSuccessfulSubmission(event) {
            // Reset form state
            this.resetValidationState();
            
            // Handle success notifications
            this.$dispatch('form-submitted-successfully', {
                response: event.detail.xhr.response
            });
        },
        
        handleFailedSubmission(event) {
            try {
                const response = JSON.parse(event.detail.xhr.response);
                
                if (response.errors) {
                    this.displayServerErrors(response.errors);
                }
                
                if (response.message) {
                    this.showErrorNotification(response.message);
                }
            } catch (e) {
                console.error('Error parsing server response:', e);
                this.showErrorNotification('An unexpected error occurred. Please try again.');
            }
        },
        
        handleRequestError(event) {
            const status = event.detail.xhr.status;
            let message = 'An error occurred. Please try again.';
            
            switch (status) {
                case 422:
                    message = 'Please correct the errors and try again.';
                    break;
                case 429:
                    message = 'Too many requests. Please wait a moment and try again.';
                    break;
                case 500:
                    message = 'Server error. Please try again later.';
                    break;
            }
            
            this.showErrorNotification(message);
        },
        
        displayServerErrors(errors) {
            errors.forEach(error => {
                const field = this.$el.querySelector(`[name="${error.field}"]`);
                if (field) {
                    // Update Alpine component if it exists
                    const alpineData = Alpine.$data(field);
                    if (alpineData) {
                        alpineData.hasError = true;
                        alpineData.errorMessage = error.message;
                    }
                    
                    // Update internal state
                    this.fieldValidation[error.field] = {
                        ...this.fieldValidation[error.field],
                        valid: false,
                        error: error.message
                    };
                }
            });
            
            this.updateFormValidity();
        },
        
        clearServerErrors() {
            // Clear all server-side errors
            Object.keys(this.fieldValidation).forEach(fieldName => {
                this.fieldValidation[fieldName].error = '';
                
                const field = this.$el.querySelector(`[name="${fieldName}"]`);
                if (field) {
                    const alpineData = Alpine.$data(field);
                    if (alpineData && alpineData.clearError) {
                        alpineData.clearError();
                    }
                }
            });
        },
        
        resetValidationState() {
            this.fieldValidation = {};
            this.touchedFields.clear();
            this.isValid = false;
            this.initializeFields();
        },
        
        showValidationSummary() {
            const invalidFields = Object.entries(this.fieldValidation)
                .filter(([_, state]) => !state.valid && state.required)
                .map(([fieldName, _]) => fieldName);
                
            if (invalidFields.length > 0) {
                this.showErrorNotification(
                    `Please correct the errors in: ${invalidFields.join(', ')}`
                );
            }
        },
        
        showErrorNotification(message) {
            // Use your notification system
            if (window.showNotification) {
                window.showNotification('error', message);
            } else {
                alert(message); // Fallback
            }
        },
        
        // Persistence helpers
        storeValidationState() {
            if (options.persistState) {
                const state = {
                    fieldValidation: this.fieldValidation,
                    touchedFields: Array.from(this.touchedFields)
                };
                sessionStorage.setItem(`form_state_${this.$el.id}`, JSON.stringify(state));
            }
        },
        
        restoreValidationState() {
            if (options.persistState && this.$el.id) {
                const stored = sessionStorage.getItem(`form_state_${this.$el.id}`);
                if (stored) {
                    try {
                        const state = JSON.parse(stored);
                        this.fieldValidation = state.fieldValidation || {};
                        this.touchedFields = new Set(state.touchedFields || []);
                        this.updateFormValidity();
                    } catch (e) {
                        console.warn('Failed to restore form state:', e);
                    }
                }
            }
        },
        
        // Public API methods
        getFormData() {
            const formData = new FormData(this.$el);
            const data = {};
            for (const [key, value] of formData.entries()) {
                data[key] = value;
            }
            return data;
        },
        
        setFieldValue(fieldName, value) {
            const field = this.$el.querySelector(`[name="${fieldName}"]`);
            if (field) {
                field.value = value;
                field.dispatchEvent(new Event('input', { bubbles: true }));
            }
        },
        
        triggerFieldValidation(fieldName) {
            const field = this.$el.querySelector(`[name="${fieldName}"]`);
            if (field) {
                const alpineData = Alpine.$data(field);
                if (alpineData && typeof alpineData.validate === 'function') {
                    return alpineData.validate();
                }
            }
        }
    };
}
```
<!-- LLM-FORM-MANAGER-END -->

---

## Production Considerations

### Rate Limiting and Caching

<!-- LLM-PRODUCTION-OPTIMIZATIONS-START -->
```go
// utils/rate_limiter.go
package utils

import (
    "sync"
    "time"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *RateLimiter) Allow(identifier string) bool {
    rl.mu.RLock()
    limiter, exists := rl.limiters[identifier]
    rl.mu.RUnlock()
    
    if !exists {
        rl.mu.Lock()
        // Double-check after acquiring write lock
        if limiter, exists = rl.limiters[identifier]; !exists {
            limiter = rate.NewLimiter(rl.rate, rl.burst)
            rl.limiters[identifier] = limiter
        }
        rl.mu.Unlock()
    }
    
    return limiter.Allow()
}

// Cleanup old limiters periodically
func (rl *RateLimiter) Cleanup() {
    ticker := time.NewTicker(time.Hour)
    for range ticker.C {
        rl.mu.Lock()
        for identifier, limiter := range rl.limiters {
            // Remove limiters that haven't been used recently
            if limiter.TokensAt(time.Now()) == float64(rl.burst) {
                delete(rl.limiters, identifier)
            }
        }
        rl.mu.Unlock()
    }
}

// validation/cache.go
package validation

import (
    "crypto/md5"
    "fmt"
    "sync"
    "time"
)

type ValidationCache struct {
    cache map[string]CacheEntry
    mu    sync.RWMutex
}

type CacheEntry struct {
    Value     interface{}
    ExpiresAt time.Time
}

func NewValidationCache() *ValidationCache {
    vc := &ValidationCache{
        cache: make(map[string]CacheEntry),
    }
    
    // Start cleanup goroutine
    go vc.cleanup()
    
    return vc
}

func (vc *ValidationCache) Set(key string, value interface{}, ttl time.Duration) {
    vc.mu.Lock()
    defer vc.mu.Unlock()
    
    vc.cache[key] = CacheEntry{
        Value:     value,
        ExpiresAt: time.Now().Add(ttl),
    }
}

func (vc *ValidationCache) Get(key string) (interface{}, bool) {
    vc.mu.RLock()
    defer vc.mu.RUnlock()
    
    entry, exists := vc.cache[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    
    return entry.Value, true
}

func (vc *ValidationCache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        vc.mu.Lock()
        now := time.Now()
        for key, entry := range vc.cache {
            if now.After(entry.ExpiresAt) {
                delete(vc.cache, key)
            }
        }
        vc.mu.Unlock()
    }
}

// Generate cache key for validation
func GenerateValidationCacheKey(field, value string, additionalParams ...string) string {
    data := fmt.Sprintf("%s:%s:%v", field, value, additionalParams)
    return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// Enhanced validator with caching
func (v *Validator) UniqueEmailCached(field, email string, cache *ValidationCache, db *DatabaseValidator, excludeID int) *Validator {
    if email == "" {
        return v
    }
    
    cacheKey := GenerateValidationCacheKey("email_unique", email, fmt.Sprintf("%d", excludeID))
    
    if exists, found := cache.Get(cacheKey); found {
        if exists.(bool) {
            v.AddError(field, v.getMessage("email_taken"), "email_taken")
        }
        return v
    }
    
    exists, err := db.EmailExists(context.Background(), email, excludeID)
    if err != nil {
        v.AddError(field, v.getMessage("validation_error"), "validation_error")
    } else {
        // Cache result for 5 minutes
        cache.Set(cacheKey, exists, 5*time.Minute)
        if exists {
            v.AddError(field, v.getMessage("email_taken"), "email_taken")
        }
    }
    
    return v
}
```
<!-- LLM-PRODUCTION-OPTIMIZATIONS-END -->

### Monitoring and Analytics

<!-- LLM-MONITORING-START -->
```go
// monitoring/validation_metrics.go
package monitoring

import (
    "context"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    validationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "validation_duration_seconds",
            Help: "Time spent on field validation",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
        },
        []string{"field", "validation_type"},
    )
    
    validationErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "validation_errors_total",
            Help: "Total number of validation errors",
        },
        []string{"field", "error_code"},
    )
    
    validationCacheHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "validation_cache_hits_total",
            Help: "Total number of validation cache hits",
        },
        []string{"field"},
    )
    
    validationCacheMisses = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "validation_cache_misses_total",
            Help: "Total number of validation cache misses",
        },
        []string{"field"},
    )
)

type ValidationMetrics struct {
    fieldValidationCount map[string]int
    errorsByField       map[string]map[string]int
    validationTimes     map[string][]time.Duration
}

func NewValidationMetrics() *ValidationMetrics {
    return &ValidationMetrics{
        fieldValidationCount: make(map[string]int),
        errorsByField:       make(map[string]map[string]int),
        validationTimes:     make(map[string][]time.Duration),
    }
}

func (vm *ValidationMetrics) RecordValidation(field, validationType string, duration time.Duration, errors []ValidationError) {
    // Prometheus metrics
    validationDuration.WithLabelValues(field, validationType).Observe(duration.Seconds())
    
    for _, err := range errors {
        validationErrors.WithLabelValues(err.Field, err.Code).Inc()
    }
    
    // Internal metrics
    vm.fieldValidationCount[field]++
    vm.validationTimes[field] = append(vm.validationTimes[field], duration)
    
    if vm.errorsByField[field] == nil {
        vm.errorsByField[field] = make(map[string]int)
    }
    
    for _, err := range errors {
        vm.errorsByField[field][err.Code]++
    }
}

func (vm *ValidationMetrics) RecordCacheHit(field string) {
    validationCacheHits.WithLabelValues(field).Inc()
}

func (vm *ValidationMetrics) RecordCacheMiss(field string) {
    validationCacheMisses.WithLabelValues(field).Inc()
}

// Enhanced validator with metrics
func (v *Validator) ValidateWithMetrics(field, value string, metrics *ValidationMetrics, validationFunc func() error) *Validator {
    start := time.Now()
    
    err := validationFunc()
    duration := time.Since(start)
    
    errors := []ValidationError{}
    if err != nil {
        // Extract validation errors
        if ve, ok := err.(*ValidationErrors); ok {
            errors = *ve
        }
    }
    
    metrics.RecordValidation(field, "server", duration, errors)
    
    return v
}
```
<!-- LLM-MONITORING-END -->

### Security Hardening

<!-- LLM-SECURITY-START -->
```go
// security/validation_security.go
package security

import (
    "context"
    "crypto/subtle"
    "html"
    "net/http"
    "regexp"
    "strings"
    "time"
)

type SecurityValidator struct {
    rateLimiter    *RateLimiter
    csrfValidator  *CSRFValidator
    contentFilter  *ContentFilter
}

func NewSecurityValidator() *SecurityValidator {
    return &SecurityValidator{
        rateLimiter:   NewRateLimiter(rate.Limit(10), 20), // 10 req/sec, burst of 20
        csrfValidator: NewCSRFValidator(),
        contentFilter: NewContentFilter(),
    }
}

// CSRF Protection
type CSRFValidator struct {
    tokenSecret []byte
}

func NewCSRFValidator() *CSRFValidator {
    return &CSRFValidator{
        tokenSecret: []byte("your-csrf-secret-key"), // Use environment variable
    }
}

func (cv *CSRFValidator) ValidateToken(r *http.Request) bool {
    // Extract token from form or header
    token := r.FormValue("csrf_token")
    if token == "" {
        token = r.Header.Get("X-CSRF-Token")
    }
    
    // Extract expected token from session/cookie
    expectedToken := getCSRFTokenFromSession(r)
    
    return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}

// Content filtering for XSS prevention
type ContentFilter struct {
    scriptPattern *regexp.Regexp
    eventPattern  *regexp.Regexp
    urlPattern    *regexp.Regexp
}

func NewContentFilter() *ContentFilter {
    return &ContentFilter{
        scriptPattern: regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
        eventPattern:  regexp.MustCompile(`(?i)on\w+\s*=`),
        urlPattern:    regexp.MustCompile(`(?i)javascript:`),
    }
}

func (cf *ContentFilter) SanitizeInput(input string) string {
    // HTML escape
    sanitized := html.EscapeString(input)
    
    // Remove script tags
    sanitized = cf.scriptPattern.ReplaceAllString(sanitized, "")
    
    // Remove event handlers
    sanitized = cf.eventPattern.ReplaceAllString(sanitized, "")
    
    // Remove javascript: URLs
    sanitized = cf.urlPattern.ReplaceAllString(sanitized, "")
    
    return sanitized
}

func (cf *ContentFilter) IsPotentiallyMalicious(input string) bool {
    // Check for common XSS patterns
    patterns := []string{
        "<script",
        "javascript:",
        "onload=",
        "onerror=",
        "onclick=",
        "eval(",
        "expression(",
    }
    
    lowerInput := strings.ToLower(input)
    for _, pattern := range patterns {
        if strings.Contains(lowerInput, pattern) {
            return true
        }
    }
    
    return false
}

// Enhanced validator with security checks
func (v *Validator) SecureValidation(field, value string, secValidator *SecurityValidator) *Validator {
    // Content filtering
    if secValidator.contentFilter.IsPotentiallyMalicious(value) {
        v.AddError(field, "Input contains potentially malicious content", "security_violation")
        return v
    }
    
    // Length limits for DoS prevention
    if len(value) > 10000 { // Configurable limit
        v.AddError(field, "Input too long", "input_too_long")
        return v
    }
    
    return v
}

// Audit logging for security events
type SecurityAuditLogger struct {
    logger Logger
}

func (sal *SecurityAuditLogger) LogValidationSecurityEvent(ctx context.Context, event SecurityEvent) {
    sal.logger.With(
        "event_type", "validation_security",
        "user_id", getUserIDFromContext(ctx),
        "ip_address", getIPFromContext(ctx),
        "field", event.Field,
        "violation_type", event.ViolationType,
        "input_sample", event.InputSample[:min(len(event.InputSample), 100)], // Limit logged input
        "timestamp", time.Now(),
    ).Warn("Validation security event detected")
}

type SecurityEvent struct {
    Field         string
    ViolationType string
    InputSample   string
    UserAgent     string
    IPAddress     string
}

// Input validation middleware
func ValidationSecurityMiddleware(secValidator *SecurityValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Rate limiting
            clientIP := getClientIP(r)
            if !secValidator.rateLimiter.Allow(clientIP) {
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }
            
            // CSRF validation for state-changing requests
            if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
                if !secValidator.csrfValidator.ValidateToken(r) {
                    http.Error(w, "CSRF token validation failed", http.StatusForbidden)
                    return
                }
            }
            
            // Content-Length limits
            if r.ContentLength > 1024*1024 { // 1MB limit
                http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```
<!-- LLM-SECURITY-END -->

## Testing Strategy

### Unit Testing

<!-- LLM-UNIT-TESTS-START -->
```go
// validation/validation_test.go
package validation_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "your-app/validation"
)

func TestValidator_Email(t *testing.T) {
    tests := []struct {
        name     string
        email    string
        wantErr  bool
        errCode  string
    }{
        {
            name:    "valid email",
            email:   "user@example.com",
            wantErr: false,
        },
        {
            name:    "valid email with subdomain",
            email:   "user@mail.example.com",
            wantErr: false,
        },
        {
            name:    "invalid email - no @",
            email:   "userexample.com",
            wantErr: true,
            errCode: "invalid_email",
        },
        {
            name:    "invalid email - no domain",
            email:   "user@",
            wantErr: true,
            errCode: "invalid_email",
        },
        {
            name:    "invalid email - no TLD",
            email:   "user@example",
            wantErr: true,
            errCode: "invalid_email",
        },
        {
            name:    "empty email (should pass - not required)",
            email:   "",
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := validation.NewValidator()
            validator.Email("email", tt.email)
            
            if tt.wantErr {
                assert.False(t, validator.IsValid())
                assert.True(t, validator.HasErrorsForField("email"))
                
                errors := validator.Errors()
                require.Len(t, errors, 1)
                assert.Equal(t, tt.errCode, errors[0].Code)
            } else {
                assert.True(t, validator.IsValid())
                assert.False(t, validator.HasErrorsForField("email"))
            }
        })
    }
}

func TestValidator_ChainedValidation(t *testing.T) {
    validator := validation.NewValidator()
    
    // Test chained validation
    validator.Required("name", "").
            Email("email", "invalid-email").
            MinLength("password", "123", 8)
    
    assert.False(t, validator.IsValid())
    assert.Equal(t, 3, validator.ErrorCount())
    
    errors := validator.Errors()
    fieldErrors := make(map[string]string)
    for _, err := range errors {
        fieldErrors[err.Field] = err.Code
    }
    
    assert.Equal(t, "required", fieldErrors["name"])
    assert.Equal(t, "invalid_email", fieldErrors["email"])
    assert.Equal(t, "min_length", fieldErrors["password"])
}

func TestValidator_BusinessRules(t *testing.T) {
    validator := validation.NewValidator()
    
    // Test business rule validation
    validator.ValidateBusinessRule("role", func() (bool, string) {
        return false, "Cannot assign admin role"
    })
    
    assert.False(t, validator.IsValid())
    errors := validator.Errors()
    require.Len(t, errors, 1)
    assert.Equal(t, "role", errors[0].Field)
    assert.Equal(t, "Cannot assign admin role", errors[0].Message)
    assert.Equal(t, "business_rule", errors[0].Code)
}

// Benchmark validation performance
func BenchmarkValidator_Email(b *testing.B) {
    emails := []string{
        "user@example.com",
        "test.email@domain.co.uk",
        "invalid-email",
        "another@test.org",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        validator := validation.NewValidator()
        email := emails[i%len(emails)]
        validator.Email("email", email)
    }
}
```
<!-- LLM-UNIT-TESTS-END -->

### Integration Testing

<!-- LLM-INTEGRATION-TESTS-START -->
```go
// handlers/validation_test.go
package handlers_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestValidationHandler_ValidateField(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer teardownTestDB(t, db)
    
    handler := NewValidationHandler(db)
    
    tests := []struct {
        name         string
        field        string
        value        string
        expectedCode int
        expectedValid bool
    }{
        {
            name:         "valid email",
            field:        "email",
            value:        "test@example.com",
            expectedCode: http.StatusOK,
            expectedValid: true,
        },
        {
            name:         "invalid email",
            field:        "email", 
            value:        "invalid-email",
            expectedCode: http.StatusUnprocessableEntity,
            expectedValid: false,
        },
        {
            name:         "empty required field",
            field:        "first_name",
            value:        "",
            expectedCode: http.StatusUnprocessableEntity,
            expectedValid: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create request
            data := url.Values{}
            data.Set(tt.field, tt.value)
            
            req := httptest.NewRequest("POST", "/validate?field="+tt.field, strings.NewReader(data.Encode()))
            req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
            
            // Record response
            w := httptest.NewRecorder()
            handler.ValidateField(w, req)
            
            // Assert response
            assert.Equal(t, tt.expectedCode, w.Code)
            
            var response map[string]interface{}
            err := json.Unmarshal(w.Body.Bytes(), &response)
            require.NoError(t, err)
            
            assert.Equal(t, tt.expectedValid, response["valid"])
            assert.Equal(t, tt.field, response["field"])
            
            if !tt.expectedValid {
                assert.NotEmpty(t, response["error"])
            }
        })
    }
}

func TestValidationHandler_RateLimiting(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)
    
    handler := NewValidationHandler(db)
    
    // Make requests up to rate limit
    for i := 0; i < 25; i++ { // Burst limit is 20
        data := url.Values{}
        data.Set("email", "test@example.com")
        
        req := httptest.NewRequest("POST", "/validate?field=email", strings.NewReader(data.Encode()))
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        req.Header.Set("X-Real-IP", "127.0.0.1")
        
        w := httptest.NewRecorder()
        handler.ValidateField(w, req)
        
        if i < 20 {
            assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i)
        } else {
            assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be rate limited", i)
        }
    }
}

func setupTestDB(t *testing.T) *sql.DB {
    // Setup test database with sample data
    db, err := sql.Open("postgres", "postgres://test:test@localhost/test_db?sslmode=disable")
    require.NoError(t, err)
    
    // Create test tables
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            email VARCHAR(255) UNIQUE,
            username VARCHAR(100) UNIQUE
        )
    `)
    require.NoError(t, err)
    
    // Insert test data
    _, err = db.Exec("INSERT INTO users (email, username) VALUES ('existing@example.com', 'existinguser')")
    require.NoError(t, err)
    
    return db
}

func teardownTestDB(t *testing.T, db *sql.DB) {
    _, err := db.Exec("DROP TABLE IF EXISTS users")
    require.NoError(t, err)
    db.Close()
}
```
<!-- LLM-INTEGRATION-TESTS-END -->

## Best Practices Summary

<!-- LLM-BEST-PRACTICES-SUMMARY-START -->
**VALIDATION SYSTEM BEST PRACTICES:**

### 1. Security First
- Always validate on server-side
- Sanitize user inputs to prevent XSS
- Implement rate limiting for validation endpoints
- Use CSRF protection for state-changing requests
- Log security violations for monitoring

### 2. Performance Optimization
- Cache expensive validation results
- Use debouncing for real-time validation
- Implement request cancellation in JavaScript
- Batch multiple field validations when possible
- Monitor validation performance with metrics

### 3. User Experience
- Provide immediate feedback for format errors
- Use clear, actionable error messages
- Support multiple languages for error messages
- Implement progressive enhancement (works without JS)
- Focus first invalid field on form submission

### 4. Maintainability
- Use consistent validation patterns across fields
- Extract business rules into reusable functions
- Document validation rules and business logic
- Use type-safe validation with Go structs
- Test validation logic thoroughly

### 5. Scalability
- Design for horizontal scaling
- Use distributed caching for validation results
- Implement circuit breakers for external validations
- Monitor validation performance and errors
- Plan for internationalization from the start
<!-- LLM-BEST-PRACTICES-SUMMARY-END -->

## References

<!-- LLM-REFERENCES-START -->
**RELATED DOCUMENTATION:**
- [Form Components](../components/forms.md) - Form component usage and patterns
- [HTMX Integration](../patterns/htmx-integration.md) - Server interaction patterns
- [Elements Reference](../components/elements.md) - Basic UI components

**EXTERNAL REFERENCES:**
- `templ-llms.md` - Advanced Templ features and patterns
- `flowbite-llms-full.txt` - Complete Flowbite styling reference

**OFFICIAL DOCUMENTATION:**
- [Go Validation Packages](https://github.com/go-playground/validator) - Alternative validation library
- [HTMX Documentation](https://htmx.org/docs/) - Server interaction patterns
- [Alpine.js Guide](https://alpinejs.dev) - Client-side reactivity
- [OWASP Input Validation](https://owasp.org/www-project-cheat-sheets/cheatsheets/Input_Validation_Cheat_Sheet.html) - Security guidelines
<!-- LLM-REFERENCES-END -->

<!-- LLM-METADATA-START -->
**METADATA FOR AI ASSISTANTS:**
- File Type: Implementation Guide
- Scope: Complete validation system implementation
- Technologies: Go + Templ + HTMX + Alpine.js + Flowbite
- Complexity: Advanced
- Use Case: Production-ready ERP validation system
- Security Level: High (includes security hardening)
<!-- LLM-METADATA-END -->