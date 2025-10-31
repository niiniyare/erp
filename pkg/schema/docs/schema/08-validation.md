## Section 8: Validation System

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Complete validation philosophy (HTML5 + Server)
- FieldValidation struct in detail
- HTML5-compatible validation rules
- Server-side validation (uniqueness, business rules)
- Custom validation messages
- Integration with condition package

### 8.1 Validation Philosophy

**Two-Layer Validation Approach:**

```
Layer 1: HTML5 Validation (Client-Side)
├─ Instant feedback (no server round-trip)
├─ Native browser validation
├─ Cannot be trusted (can be bypassed)
└─ User experience enhancement

Layer 2: Server Validation (Server-Side)
├─ ALWAYS re-checks HTML5 rules
├─ Adds uniqueness checks (database)
├─ Adds business rules (condition package)
├─ Custom validators (Go functions)
└─ Source of truth (secure)
```

**Critical Principle:** Never trust client-side validation. Always validate on server.

### 8.2 FieldValidation Struct

```go
type FieldValidation struct {
    // ═══════════════════════════════════════════════════════
    // HTML5-COMPATIBLE VALIDATION
    // These become HTML attributes
    // ═══════════════════════════════════════════════════════
    
    // String validation
    MinLength *int   `json:"minLength,omitempty"` // Minimum string length
    MaxLength *int   `json:"maxLength,omitempty"` // Maximum string length
    Pattern   string `json:"pattern,omitempty"`   // Regex pattern
    Format    string `json:"format,omitempty"`    // Format validator (email, url, uuid)
    
    // Number validation
    Min          *float64 `json:"min,omitempty"`          // Minimum value
    Max          *float64 `json:"max,omitempty"`          // Maximum value
    Step         *float64 `json:"step,omitempty"`         // Value increment
    Integer      bool     `json:"integer,omitempty"`      // Must be integer
    Positive     bool     `json:"positive,omitempty"`     // Must be positive
    Negative     bool     `json:"negative,omitempty"`     // Must be negative
    MultipleOf   *float64 `json:"multipleOf,omitempty"`   // Must be multiple of
    ExclusiveMin bool     `json:"exclusiveMin,omitempty"` // Min is exclusive (>)
    ExclusiveMax bool     `json:"exclusiveMax,omitempty"` // Max is exclusive (<)
    
    // Array validation
    MinItems    *int `json:"minItems,omitempty"`    // Min array length
    MaxItems    *int `json:"maxItems,omitempty"`    // Max array length
    UniqueItems bool `json:"uniqueItems,omitempty"` // Items must be unique
    
    // Custom error messages
    Messages Messages `json:"messages,omitempty"` // Custom error messages
    
    // ═══════════════════════════════════════════════════════
    // SERVER-ONLY VALIDATION
    // These cannot be checked client-side
    // ═══════════════════════════════════════════════════════
    
    Server *ServerValidation `json:"server,omitempty"`
    
    // Custom validation function (registered in Go)
    Custom string `json:"custom,omitempty"`
}

// Messages holds custom validation error messages
type Messages struct {
    Required  string `json:"required,omitempty"`
    MinLength string `json:"minLength,omitempty"`
    MaxLength string `json:"maxLength,omitempty"`
    Pattern   string `json:"pattern,omitempty"`
    Min       string `json:"min,omitempty"`
    Max       string `json:"max,omitempty"`
    Custom    string `json:"custom,omitempty"`
}

// ServerValidation defines server-only validation
type ServerValidation struct {
    // Database uniqueness
    Unique       bool     `json:"unique,omitempty"`
    UniqueWith   []string `json:"uniqueWith,omitempty"` // Composite unique
    
    // Business rules (uses condition package)
    BusinessRules []BusinessRule `json:"businessRules,omitempty"`
    
    // Custom validator function name
    Custom string `json:"custom,omitempty"`
}

// BusinessRule defines a business validation rule
type BusinessRule struct {
    ID          string                   `json:"id"`
    Description string                   `json:"description"`
    Condition   condition.ConditionGroup `json:"condition"`
    Message     string                   `json:"message"`
}
```

### 8.3 HTML5-Compatible Validation

These rules become HTML attributes and are checked by the browser:

#### String Validation

**MinLength:**
```json
{
  "validation": {
    "minLength": 5
  }
}
```

**Renders as:**
```html
<input type="text" minlength="5" />
```

**Browser behavior:** Prevents submission if length < 5.

**MaxLength:**
```json
{
  "validation": {
    "maxLength": 100
  }
}
```

**Renders as:**
```html
<input type="text" maxlength="100" />
```

**Browser behavior:** Prevents typing beyond 100 characters.

**Pattern:**
```json
{
  "validation": {
    "pattern": "^[A-Z0-9]+$",
    "messages": {
      "pattern": "Only uppercase letters and numbers allowed"
    }
  }
}
```

**Renders as:**
```html
<input type="text" pattern="^[A-Z0-9]+$" />
```

**Browser behavior:** Validates against regex pattern.

**Format:**
```json
{
  "type": "email",
  "validation": {
    "format": "email"
  }
}
```

**Note:** `format` is not an HTML attribute, but field type affects validation.

**Common formats:**
- `email` - Email validation
- `url` - URL validation
- `uuid` - UUID format
- `phone` - Phone number (with mask)
- `postal_code` - Postal code

#### Number Validation

**Min/Max:**
```json
{
  "type": "number",
  "validation": {
    "min": 18,
    "max": 120
  }
}
```

**Renders as:**
```html
<input type="number" min="18" max="120" />
```

**Step:**
```json
{
  "type": "number",
  "validation": {
    "step": 0.01
  }
}
```

**Renders as:**
```html
<input type="number" step="0.01" />
```

**Integer:**
```json
{
  "type": "number",
  "validation": {
    "integer": true
  }
}
```

**Renders as:**
```html
<input type="number" step="1" />
```

**Exclusive Min/Max:**
```json
{
  "validation": {
    "min": 0,
    "exclusiveMin": true  // Value must be > 0 (not >= 0)
  }
}
```

**Note:** HTML5 doesn't support exclusive, so this is server-only.

#### Array Validation

**MinItems/MaxItems:**
```json
{
  "type": "multiselect",
  "validation": {
    "minItems": 1,
    "maxItems": 5
  }
}
```

**UniqueItems:**
```json
{
  "type": "tags",
  "validation": {
    "uniqueItems": true
  }
}
```

### 8.4 Custom Error Messages

Override default browser messages:

```json
{
  "name": "email",
  "type": "email",
  "required": true,
  "validation": {
    "maxLength": 255,
    "messages": {
      "required": "We need your email to contact you",
      "pattern": "Please enter a valid email address like user@example.com",
      "maxLength": "Email is too long (max 255 characters)"
    }
  }
}
```

**How it works:**

1. Browser validates with HTML5 attributes
2. On invalid event, JavaScript sets custom message:

```javascript
// Alpine.js handles this
input.addEventListener('invalid', (e) => {
  e.preventDefault();
  const field = schema.fields.find(f => f.name === input.name);
  const customMessage = field.validation.messages[e.target.validity];
  if (customMessage) {
    e.target.setCustomValidity(customMessage);
  }
});
```

### 8.5 Server-Side Validation

**Critical:** Server ALWAYS re-validates everything, even HTML5 rules.

#### Uniqueness Validation

**Single field uniqueness:**
```json
{
  "name": "email",
  "type": "email",
  "validation": {
    "server": {
      "unique": true
    }
  }
}
```

**Server checks:**
```go
func (v *Validator) ValidateData(ctx context.Context, schema *Schema, data map[string]any) []ValidationError {
    var errors []ValidationError
    
    for _, field := range schema.Fields {
        if field.Validation != nil && field.Validation.Server != nil {
            sv := field.Validation.Server
            
            // Check uniqueness
            if sv.Unique {
                exists, err := v.db.Exists(ctx, field.Name, data[field.Name])
                if err != nil {
                    return []ValidationError{{Field: field.Name, Message: "Database error"}}
                }
                if exists {
                    errors = append(errors, ValidationError{
                        Field:   field.Name,
                        Message: fmt.Sprintf("%s is already in use", field.Label),
                    })
                }
            }
        }
    }
    
    return errors
}
```

**Composite uniqueness:**
```json
{
  "name": "email",
  "type": "email",
  "validation": {
    "server": {
      "unique": true,
      "uniqueWith": ["tenant_id"]
    }
  }
}
```

**Meaning:** Email must be unique within the same tenant.

#### Business Rules Validation

Business rules use the condition package for complex logic:

```json
{
  "name": "discount_code",
  "type": "text",
  "validation": {
    "server": {
      "businessRules": [
        {
          "id": "discount-requires-minimum",
          "description": "Discount codes require minimum order of $100",
          "condition": {
            "conjunction": "and",
            "rules": [
              {
                "left": {"type": "field", "field": "order_total"},
                "op": "greater",
                "right": 100
              }
            ]
          },
          "message": "Discount codes require a minimum order of $100"
        },
        {
          "id": "discount-premium-only",
          "description": "Only premium members can use discount codes",
          "condition": {
            "conjunction": "and",
            "rules": [
              {
                "left": {"type": "field", "field": "is_premium"},
                "op": "equals",
                "right": true
              }
            ]
          },
          "message": "Only premium members can use discount codes"
        }
      ]
    }
  }
}
```

**Server validates:**
```go
func (v *Validator) ValidateBusinessRules(ctx context.Context, field *Field, data map[string]any) []ValidationError {
    var errors []ValidationError
    
    if field.Validation == nil || field.Validation.Server == nil {
        return errors
    }
    
    for _, rule := range field.Validation.Server.BusinessRules {
        evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
        evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
        
        result, err := evaluator.Evaluate(ctx, &rule.Condition, evalCtx)
        if err != nil || !result {
            errors = append(errors, ValidationError{
                Field:   field.Name,
                Message: rule.Message,
                Code:    rule.ID,
            })
        }
    }
    
    return errors
}
```

#### Custom Validators

Register custom validation functions:

**Schema:**
```json
{
  "name": "username",
  "type": "text",
  "validation": {
    "custom": "valid_username"
  }
}
```

**Go code:**
```go
// Register custom validator
registry := validate.NewValidationRegistry()
registry.Register("valid_username", func(ctx context.Context, value any, params map[string]any) error {
    username, ok := value.(string)
    if !ok {
        return errors.New("username must be string")
    }
    
    // Custom validation logic
    if strings.Contains(username, "admin") {
        return errors.New("username cannot contain 'admin'")
    }
    
    // Check against reserved words
    reserved := []string{"root", "system", "administrator"}
    for _, word := range reserved {
        if strings.EqualFold(username, word) {
            return errors.New("this username is reserved")
        }
    }
    
    return nil
})

// Use validator
validator := validate.NewValidator(db)
validator.SetRegistry(registry)
errors := validator.ValidateData(ctx, schema, data)
```

### 8.6 Complete Validation Flow

**Client-Side (Browser):**
```
1. User fills field
   ↓
2. Browser validates HTML5 attributes (instant)
   ↓
3. If invalid: Show error, prevent submission
   ↓
4. If valid: Allow submission
```

**Server-Side:**
```
1. Receive form data
   ↓
2. Load schema
   ↓
3. Re-validate HTML5 rules (don't trust client)
   ↓
4. Validate uniqueness (database check)
   ↓
5. Evaluate business rules (condition package)
   ↓
6. Run custom validators (Go functions)
   ↓
7. If any errors: Return error list
   ↓
8. If all valid: Process data
```

**Example Handler:**
```go
func HandleSubmit(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse form data
    r.ParseForm()
    data := formToMap(r.Form)
    
    // Load schema
    schema, err := registry.Get(ctx, "user-form")
    if err != nil {
        http.Error(w, "Schema not found", 404)
        return
    }
    
    // Validate
    validator := validate.NewValidator(db)
    errors := validator.ValidateData(ctx, schema, data)
    
    if len(errors) > 0 {
        // Return errors to client
        views.ValidationErrors(errors).Render(ctx, w)
        return
    }
    
    // Save data
    id, err := saveUser(ctx, data)
    if err != nil {
        http.Error(w, "Save failed", 500)
        return
    }
    
    // Return success
    views.SuccessMessage("User created successfully").Render(ctx, w)
}
```

### 8.7 Validation Examples

#### Example 1: Simple Text Field

```json
{
  "name": "username",
  "type": "text",
  "label": "Username",
  "required": true,
  "validation": {
    "minLength": 3,
    "maxLength": 20,
    "pattern": "^[a-zA-Z0-9_]+$",
    "messages": {
      "required": "Username is required",
      "minLength": "Username must be at least 3 characters",
      "maxLength": "Username cannot exceed 20 characters",
      "pattern": "Username can only contain letters, numbers, and underscores"
    },
    "server": {
      "unique": true
    }
  }
}
```

#### Example 2: Email with Uniqueness

```json
{
  "name": "email",
  "type": "email",
  "label": "Email Address",
  "required": true,
  "validation": {
    "maxLength": 255,
    "messages": {
      "required": "Email address is required",
      "pattern": "Please enter a valid email address"
    },
    "server": {
      "unique": true,
      "uniqueWith": ["tenant_id"]
    }
  }
}
```

#### Example 3: Password with Complex Rules

```json
{
  "name": "password",
  "type": "password",
  "label": "Password",
  "required": true,
  "validation": {
    "minLength": 8,
    "maxLength": 128,
    "pattern": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]+$",
    "messages": {
      "required": "Password is required",
      "minLength": "Password must be at least 8 characters",
      "pattern": "Password must contain: uppercase, lowercase, number, and special character"
    }
  }
}
```

#### Example 4: Number with Range

```json
{
  "name": "age",
  "type": "number",
  "label": "Age",
  "required": true,
  "validation": {
    "min": 18,
    "max": 120,
    "integer": true,
    "messages": {
      "required": "Age is required",
      "min": "You must be at least 18 years old",
      "max": "Please enter a valid age"
    }
  }
}
```

#### Example 5: Business Rule Validation

```json
{
  "name": "end_date",
  "type": "date",
  "label": "End Date",
  "required": true,
  "validation": {
    "server": {
      "businessRules": [
        {
          "id": "end-after-start",
          "description": "End date must be after start date",
          "condition": {
            "conjunction": "and",
            "rules": [
              {
                "left": {"type": "field", "field": "end_date"},
                "op": "greater",
                "right": {"type": "field", "field": "start_date"}
              }
            ]
          },
          "message": "End date must be after start date"
        }
      ]
    }
  }
}
```

#### Example 6: Multi-Select with Limits

```json
{
  "name": "skills",
  "type": "multiselect",
  "label": "Skills",
  "required": true,
  "options": [
    {"value": "go", "label": "Go"},
    {"value": "js", "label": "JavaScript"},
    {"value": "python", "label": "Python"}
  ],
  "validation": {
    "minItems": 1,
    "maxItems": 5,
    "messages": {
      "required": "Please select at least one skill",
      "minItems": "Please select at least one skill",
      "maxItems": "You can select up to 5 skills"
    }
  }
}
```

### 8.8 Common Pitfalls

#### Pitfall 1: Trusting Client Validation

```go
// ❌ WRONG: Only client-side validation
<input type="email" required />

// ✅ RIGHT: Client + Server validation
<input type="email" required />
// Plus server-side:
validator.ValidateData(ctx, schema, data)
```

#### Pitfall 2: Not Re-Validating HTML5 Rules

```go
// ❌ WRONG: Only checking server-specific rules
if field.Validation.Server != nil {
    // Check uniqueness
}
// Missing: Re-check minLength, maxLength, pattern, etc.

// ✅ RIGHT: Check everything
validator.ValidateField(ctx, field, value, allData)
// This checks HTML5 rules + server rules
```

#### Pitfall 3: Missing Error Messages

```json
// ❌ BAD: No custom messages
{
  "validation": {
    "pattern": "^[A-Z0-9]+$"
  }
}
// User sees: "Please match the requested format" (unclear)

// ✅ GOOD: Custom messages
{
  "validation": {
    "pattern": "^[A-Z0-9]+$",
    "messages": {
      "pattern": "Only uppercase letters and numbers are allowed"
    }
  }
}
```

#### Pitfall 4: Exclusive Min/Max in HTML5

```json
// ❌ WRONG: HTML5 doesn't support exclusive
{
  "validation": {
    "min": 0,
    "exclusiveMin": true  // Will not work in browser
  }
}

// ✅ RIGHT: Use server-side for exclusive
{
  "validation": {
    "min": 0,  // HTML5 checks >= 0
    "server": {
      "businessRules": [{
        "condition": {
          "rules": [{"left": {"type": "field", "field": "amount"}, "op": "greater", "right": 0}]
        },
        "message": "Amount must be greater than 0"
      }]
    }
  }
}
```



# 📘 Schema Engine Specification v1.0 - Part 3

**Layout, Conditional Logic, Security & Workflow**

---

**Table of Contents - Part 3**
- Section 9: Layout System
- Section 10: Conditional Visibility
- Section 11: Security & Permissions
- Section 12: Events & Actions (HTMX Integration)
- Section 13: Workflow System

---

