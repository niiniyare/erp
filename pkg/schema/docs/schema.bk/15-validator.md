## Section 15: Validator Implementation

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Validator architecture
- Field validation implementation
- Cross-field validation
- Server-side validation
- Custom validators
- Error formatting

### 15.1 Validator Interface

```go
// Validator validates schemas and data
type Validator interface {
    // ValidateSchema validates schema structure
    ValidateSchema(ctx context.Context, schema *Schema) []ValidationError
    
    // ValidateData validates data against schema
    ValidateData(ctx context.Context, schema *Schema, data map[string]any) []ValidationError
    
    // ValidateField validates single field value
    ValidateField(ctx context.Context, field *Field, value any, allData map[string]any) []ValidationError
}

// ValidationError represents a validation error
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code,omitempty"`
    Value   any    `json:"value,omitempty"`
}
```

### 15.2 Validator Implementation

```go
// validator implements the Validator interface
type validator struct {
    db       *sql.DB
    registry ValidationRegistry
}

// NewValidator creates a new validator
func NewValidator(db *sql.DB) Validator {
    return &validator{
        db:       db,
        registry: NewValidationRegistry(),
    }
}

// ValidateSchema validates schema structure
func (v *validator) ValidateSchema(ctx context.Context, schema *Schema) []ValidationError {
    var errors []ValidationError
    
    // Use go-playground/validator for struct validation
    validate := validator.New()
    if err := validate.Struct(schema); err != nil {
        for _, err := range err.(validator.ValidationErrors) {
            errors = append(errors, ValidationError{
                Field:   err.Field(),
                Message: formatValidationError(err),
                Code:    "struct_validation",
            })
        }
    }
    
    // Check for duplicate field names
    fieldNames := make(map[string]bool)
    for _, field := range schema.Fields {
        if fieldNames[field.Name] {
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: fmt.Sprintf("Duplicate field name: %s", field.Name),
                Code:    "duplicate_field",
            })
        }
        fieldNames[field.Name] = true
    }
    
    // Validate field dependencies exist
    for _, field := range schema.Fields {
        for _, dep := range field.Dependencies {
            if !fieldNames[dep] {
                errors = append(errors, ValidationError{
                    Field:   field.Name,
                    Message: fmt.Sprintf("Dependency not found: %s", dep),
                    Code:    "missing_dependency",
                })
            }
        }
    }
    
    // Check for circular dependencies
    if err := schema.DetectCircularDependencies(); err != nil {
        errors = append(errors, ValidationError{
            Field:   "",
            Message: err.Error(),
            Code:    "circular_dependency",
        })
    }
    
    return errors
}

// ValidateData validates data against schema
func (v *validator) ValidateData(ctx context.Context, schema *Schema, data map[string]any) []ValidationError {
    var errors []ValidationError
    
    // Validate each field
    for _, field := range schema.Fields {
        value := data[field.Name]
        fieldErrors := v.ValidateField(ctx, &field, value, data)
        errors = append(errors, fieldErrors...)
    }
    
    // Cross-field validation
    if schema.Validation != nil {
        crossFieldErrors := v.validateCrossField(ctx, schema.Validation, data)
        errors = append(errors, crossFieldErrors...)
    }
    
    return errors
}

// ValidateField validates single field value
func (v *validator) ValidateField(ctx context.Context, field *Field, value any, allData map[string]any) []ValidationError {
    var errors []ValidationError
    
    // Check required
    if field.Required && isEmpty(value) {
        msg := field.Label + " is required"
        if field.Validation != nil && field.Validation.Messages.Required != "" {
            msg = field.Validation.Messages.Required
        }
        errors = append(errors, ValidationError{
            Field:   field.Name,
            Message: msg,
            Code:    "required",
        })
        return errors // Stop if required and empty
    }
    
    // Skip validation if empty and not required
    if isEmpty(value) {
        return errors
    }
    
    // Type-specific validation
    if field.Validation != nil {
        errors = append(errors, v.validateFieldValue(field, value)...)
        
        // Server-side validation
        if field.Validation.Server != nil {
            serverErrors := v.validateServer(ctx, field, value, allData)
            errors = append(errors, serverErrors...)
        }
        
        // Custom validation
        if field.Validation.Custom != "" {
            customErrors := v.validateCustom(ctx, field, value, allData)
            errors = append(errors, customErrors...)
        }
    }
    
    return errors
}

// validateFieldValue validates field value against validation rules
func (v *validator) validateFieldValue(field *Field, value any) []ValidationError {
    var errors []ValidationError
    val := field.Validation
    
    // String validation
    if str, ok := value.(string); ok {
        if val.MinLength != nil && len(str) < *val.MinLength {
            msg := fmt.Sprintf("%s must be at least %d characters", field.Label, *val.MinLength)
            if val.Messages.MinLength != "" {
                msg = val.Messages.MinLength
            }
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "min_length",
                Value:   value,
            })
        }
        
        if val.MaxLength != nil && len(str) > *val.MaxLength {
            msg := fmt.Sprintf("%s must be at most %d characters", field.Label, *val.MaxLength)
            if val.Messages.MaxLength != "" {
                msg = val.Messages.MaxLength
            }
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "max_length",
                Value:   value,
            })
        }
        
        if val.Pattern != "" {
            re, err := regexp.Compile(val.Pattern)
            if err == nil && !re.MatchString(str) {
                msg := fmt.Sprintf("%s format is invalid", field.Label)
                if val.Messages.Pattern != "" {
                    msg = val.Messages.Pattern
                }
                errors = append(errors, ValidationError{
                    Field:   field.Name,
                    Message: msg,
                    Code:    "pattern",
                    Value:   value,
                })
            }
        }
    }
    
    // Number validation
    if num, ok := toFloat64(value); ok {
        if val.Min != nil && num < *val.Min {
            exclusive := val.ExclusiveMin
            op := "at least"
            if exclusive {
                op = "greater than"
            }
            msg := fmt.Sprintf("%s must be %s %v", field.Label, op, *val.Min)
            if val.Messages.Min != "" {
                msg = val.Messages.Min
            }
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "min",
                Value:   value,
            })
        }
        
        if val.Max != nil && num > *val.Max {
            exclusive := val.ExclusiveMax
            op := "at most"
            if exclusive {
                op = "less than"
            }
            msg := fmt.Sprintf("%s must be %s %v", field.Label, op, *val.Max)
            if val.Messages.Max != "" {
                msg = val.Messages.Max
            }
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "max",
                Value:   value,
            })
        }
        
        if val.Integer && num != float64(int64(num)) {
            msg := fmt.Sprintf("%s must be a whole number", field.Label)
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "integer",
                Value:   value,
            })
        }
        
        if val.Positive && num <= 0 {
            msg := fmt.Sprintf("%s must be positive", field.Label)
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "positive",
                Value:   value,
            })
        }
        
        if val.Negative && num >= 0 {
            msg := fmt.Sprintf("%s must be negative", field.Label)
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "negative",
                Value:   value,
            })
        }
        
        if val.MultipleOf != nil && math.Mod(num, *val.MultipleOf) != 0 {
            msg := fmt.Sprintf("%s must be a multiple of %v", field.Label, *val.MultipleOf)
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "multiple_of",
                Value:   value,
            })
        }
    }
    
    // Array validation
    if arr, ok := value.([]any); ok {
        if val.MinItems != nil && len(arr) < *val.MinItems {
            msg := fmt.Sprintf("%s must have at least %d items", field.Label, *val.MinItems)
            if val.Messages.MinLength != "" {
                msg = val.Messages.MinLength
            }
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "min_items",
                Value:   value,
            })
        }
        
        if val.MaxItems != nil && len(arr) > *val.MaxItems {
            msg := fmt.Sprintf("%s must have at most %d items", field.Label, *val.MaxItems)
            if val.Messages.MaxLength != "" {
                msg = val.Messages.MaxLength
            }
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "max_items",
                Value:   value,
            })
        }
        
        if val.UniqueItems {
            seen := make(map[string]bool)
            for _, item := range arr {
                key := fmt.Sprintf("%v", item)
                if seen[key] {
                    msg := fmt.Sprintf("%s must have unique items", field.Label)
                    errors = append(errors, ValidationError{
                        Field:   field.Name,
                        Message: msg,
                        Code:    "unique_items",
                        Value:   value,
                    })
                    break
                }
                seen[key] = true
            }
        }
    }
    
    return errors
}

// validateServer performs server-side validation
func (v *validator) validateServer(ctx context.Context, field *Field, value any, allData map[string]any) []ValidationError {
    var errors []ValidationError
    server := field.Validation.Server
    
    // Check uniqueness
    if server.Unique {
        exists, err := v.checkUnique(ctx, field, value, server.UniqueWith, allData)
        if err != nil {
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: "Failed to check uniqueness",
                Code:    "server_error",
            })
        } else if exists {
            msg := fmt.Sprintf("%s is already in use", field.Label)
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: msg,
                Code:    "unique",
                Value:   value,
            })
        }
    }
    
    // Check business rules
    for _, rule := range server.BusinessRules {
        evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
        evalCtx := condition.NewEvalContext(allData, condition.DefaultEvalOptions())
        
        result, err := evaluator.Evaluate(ctx, &rule.Condition, evalCtx)
        if err != nil || !result {
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: rule.Message,
                Code:    rule.ID,
            })
        }
    }
    
    // Custom server validation
    if server.Custom != "" {
        fn := v.registry.Get(server.Custom)
        if fn != nil {
            if err := fn(ctx, value, allData); err != nil {
                msg := err.Error()
                if field.Validation.Messages.Custom != "" {
                    msg = field.Validation.Messages.Custom
                }
                errors = append(errors, ValidationError{
                    Field:   field.Name,
                    Message: msg,
                    Code:    "custom",
                    Value:   value,
                })
            }
        }
    }
    
    return errors
}

// checkUnique checks if value is unique in database
func (v *validator) checkUnique(ctx context.Context, field *Field, value any, uniqueWith []string, allData map[string]any) (bool, error) {
    // This is a simplified example - you'd need table name from somewhere
    tableName := "users" // Would come from schema metadata
    
    query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1", tableName, field.Name)
    args := []interface{}{value}
    argCount := 2
    
    // Add composite unique constraints
    for _, withField := range uniqueWith {
        query += fmt.Sprintf(" AND %s = $%d", withField, argCount)
        args = append(args, allData[withField])
        argCount++
    }
    
    query += ")"
    
    var exists bool
    err := v.db.QueryRowContext(ctx, query, args...).Scan(&exists)
    return exists, err
}

// validateCustom runs custom validation function
func (v *validator) validateCustom(ctx context.Context, field *Field, value any, allData map[string]any) []ValidationError {
    fn := v.registry.Get(field.Validation.Custom)
    if fn == nil {
        return nil
    }
    
    if err := fn(ctx, value, allData); err != nil {
        msg := err.Error()
        if field.Validation.Messages.Custom != "" {
            msg = field.Validation.Messages.Custom
        }
        return []ValidationError{{
            Field:   field.Name,
            Message: msg,
            Code:    "custom",
            Value:   value,
        }}
    }
    
    return nil
}

// validateCrossField validates cross-field rules
func (v *validator) validateCrossField(ctx context.Context, validation *Validation, data map[string]any) []ValidationError {
    var errors []ValidationError
    
    for _, rule := range validation.Rules {
        evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
        evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
        
        result, err := evaluator.Evaluate(ctx, &rule.Condition, evalCtx)
        if err != nil || !result {
            errors = append(errors, ValidationError{
                Field:   "", // Cross-field error
                Message: rule.Message,
                Code:    rule.ID,
            })
        }
    }
    
    return errors
}

// Helper functions
func isEmpty(value any) bool {
    if value == nil {
        return true
    }
    
    switch v := value.(type) {
    case string:
        return v == ""
    case []any:
        return len(v) == 0
    case map[string]any:
        return len(v) == 0
    default:
        return false
    }
}

func toFloat64(value any) (float64, bool) {
    switch v := value.(type) {
    case float64:
        return v, true
    case float32:
        return float64(v), true
    case int:
        return float64(v), true
    case int64:
        return float64(v), true
    case int32:
        return float64(v), true
    default:
        return 0, false
    }
}

func formatValidationError(err validator.FieldError) string {
    switch err.Tag() {
    case "required":
        return fmt.Sprintf("%s is required", err.Field())
    case "min":
        return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
    case "max":
        return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
    case "email":
        return fmt.Sprintf("%s must be a valid email", err.Field())
    default:
        return fmt.Sprintf("%s is invalid", err.Field())
    }
}
```

### 15.3 Custom Validator Registry

```go
// ValidationRegistry stores custom validation functions
type ValidationRegistry struct {
    mu         sync.RWMutex
    validators map[string]ValidationFunc
}

// ValidationFunc is a custom validation function
type ValidationFunc func(ctx context.Context, value any, allData map[string]any) error

// NewValidationRegistry creates a new registry
func NewValidationRegistry() ValidationRegistry {
    return ValidationRegistry{
        validators: make(map[string]ValidationFunc),
    }
}

// Register adds a custom validator
func (r *ValidationRegistry) Register(name string, fn ValidationFunc) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.validators[name] = fn
}

// Get retrieves a validator
func (r *ValidationRegistry) Get(name string) ValidationFunc {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.validators[name]
}

// Example custom validators
func init() {
    registry := NewValidationRegistry()
    
    // Username validator
    registry.Register("valid_username", func(ctx context.Context, value any, data map[string]any) error {
        username, ok := value.(string)
        if !ok {
            return errors.New("username must be a string")
        }
        
        // Check against reserved words
        reserved := []string{"admin", "root", "system", "administrator"}
        for _, word := range reserved {
            if strings.EqualFold(username, word) {
                return errors.New("this username is reserved")
            }
        }
        
        // Check for profanity (simplified)
        if strings.Contains(strings.ToLower(username), "badword") {
            return errors.New("username contains inappropriate content")
        }
        
        return nil
    })
    
    // Age validator
    registry.Register("valid_age", func(ctx context.Context, value any, data map[string]any) error {
        age, ok := value.(float64)
        if !ok {
            return errors.New("age must be a number")
        }
        
        if age < 0 || age > 150 {
            return errors.New("age must be between 0 and 150")
        }
        
        return nil
    })
    
    // Date range validator
    registry.Register("valid_date_range", func(ctx context.Context, value any, data map[string]any) error {
        endDate, ok := value.(time.Time)
        if !ok {
            return errors.New("invalid date format")
        }
        
        startDate, ok := data["start_date"].(time.Time)
        if !ok {
            return errors.New("start_date is required")
        }
        
        if endDate.Before(startDate) {
            return errors.New("end date must be after start date")
        }
        
        return nil
    })
}
```

### 15.4 Validator Usage Examples

#### Example 1: Validate Form Submission

```go
func HandleSubmit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse form
    r.ParseForm()
    data := formToMap(r.Form)
    
    // Load schema
    schema, _ := registry.Get(ctx, "user-form")
    
    // Validate
    validator := NewValidator(db)
    errors := validator.ValidateData(ctx, schema, data)
    
    if len(errors) > 0 {
        // Return validation errors
        views.ValidationErrors(errors).Render(ctx, w)
        return
    }
    
    // Save data
    id, _ := repo.Save(ctx, data)
    
    views.SuccessMessage("User created successfully").Render(ctx, w)
}
```

#### Example 2: Validate Single Field (AJAX)

```go
func HandleFieldValidation(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    fieldName := r.URL.Query().Get("field")
    value := r.FormValue(fieldName)
    
    // Load schema
    schema, _ := registry.Get(ctx, "user-form")
    
    // Get field
    field, ok := schema.GetField(fieldName)
    if !ok {
        http.Error(w, "Field not found", 404)
        return
    }
    
    // Validate
    validator := NewValidator(db)
    errors := validator.ValidateField(ctx, field, value, nil)
    
    if len(errors) > 0 {
        // Return error
        w.Write([]byte(errors[0].Message))
        return
    }
    
    // Return success
    w.Write([]byte("✓"))
}
```

---

