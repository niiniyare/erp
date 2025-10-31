## Section 10: Conditional Visibility

### 🎯 Learning Objectives

By the end of this section, you will understand:
- Two strategies: client-side and server-side
- When to use which strategy
- ShowIf expressions (Alpine.js)
- RequirePermission checks (server)
- Runtime flags (FieldRuntime)
- Complex conditions (condition package)

### 10.1 The Two Strategies

**Client-Side (Alpine.js):**
- For simple UI logic
- Field depends on another field's value
- Instant visibility toggle
- No server round-trip
- NOT for security decisions

**Server-Side (Permissions):**
- For security decisions
- Field depends on user role/permission
- Server decides, client respects
- Cannot be bypassed
- Always use for sensitive data

### 10.2 Client-Side Conditional Visibility

#### ShowIf Expression

**Schema Field:**
```go
// ShowIf: Client-side visibility (Alpine.js expression)
// Example: "needs_shipping === true"
// For simple UI logic only
ShowIf string `json:"showIf,omitempty" validate:"js_expression"`
```

**Example:**
```json
{
  "fields": [
    {
      "name": "needs_shipping",
      "type": "checkbox",
      "label": "Ship to different address"
    },
    {
      "name": "shipping_address",
      "type": "textarea",
      "label": "Shipping Address",
      "showIf": "needs_shipping === true"
    }
  ]
}
```

**templ Renders:**
```go
templ FieldWrapper(field *schema.Field, value interface{}) {
    <div 
        class="field"
        if field.ShowIf != "" {
            x-show={field.ShowIf}
        }
    >
        @FieldRenderer(field, value)
    </div>
}
```

**HTML Output:**
```html
<div class="field">
  <label>
    <input type="checkbox" name="needs_shipping" x-model="needs_shipping" />
    Ship to different address
  </label>
</div>

<div class="field" x-show="needs_shipping === true">
  <label for="shipping_address">Shipping Address</label>
  <textarea name="shipping_address"></textarea>
</div>
```

**Behavior:**
- When checkbox unchecked → `shipping_address` hidden
- When checkbox checked → `shipping_address` visible
- Instant (no server call)

#### ShowIf Expression Syntax

Alpine.js expressions support:

**Equality:**
```json
"showIf": "country === 'US'"
"showIf": "age >= 18"
"showIf": "status !== 'cancelled'"
```

**Boolean:**
```json
"showIf": "is_premium"
"showIf": "!is_deleted"
"showIf": "agreed_terms && verified_email"
```

**Multiple Conditions:**
```json
"showIf": "country === 'US' && state === 'CA'"
"showIf": "order_total > 100 || is_premium"
"showIf": "(country === 'US' || country === 'CA') && age >= 18"
```

**Array Checks:**
```json
"showIf": "selected_items.length > 0"
"showIf": "tags.includes('urgent')"
```

**Comparison:**
```json
"showIf": "quantity > 10"
"showIf": "price <= max_price"
"showIf": "end_date > start_date"
```

### 10.3 Server-Side Conditional Visibility

#### RequirePermission

**Schema Field:**
```go
// RequirePermission: Server-side permission check
// Example: "hr.view_salary"
// Backend decides, frontend respects
RequirePermission string `json:"requirePermission,omitempty"`

// Runtime: Runtime flags (populated by enricher)
// Contains: Visible, Editable, Reason
// Set by server, respected by client
Runtime *FieldRuntime `json:"runtime,omitempty"`
```

**Example:**
```json
{
  "name": "employee_salary",
  "type": "currency",
  "label": "Annual Salary",
  "requirePermission": "hr.view_salary"
}
```

**Backend Enrichment:**
```go
func (e *Enricher) Enrich(ctx context.Context, schema *Schema, user *User) (*Schema, error) {
    for i := range schema.Fields {
        field := &schema.Fields[i]
        
        // Check permission requirement
        if field.RequirePermission != "" {
            hasPermission := user.HasPermission(field.RequirePermission)
            
            field.Runtime = &FieldRuntime{
                Visible:  hasPermission,
                Editable: hasPermission,
                Reason:   "permission_required",
            }
        } else {
            // Default: visible and editable (if not readonly)
            field.Runtime = &FieldRuntime{
                Visible:  true,
                Editable: !field.Readonly,
            }
        }
    }
    
    return schema, nil
}
```

**templ Respects Runtime Flags:**
```go
templ FieldWrapper(field *schema.Field, value interface{}) {
    // Only render if visible
    if field.Runtime != nil && field.Runtime.Visible {
        <div class="field">
            @FieldRenderer(field, value)
        </div>
    }
}
```

**Result:**
- User without permission → Field not rendered at all
- User with permission → Field visible and editable
- Cannot be bypassed (server decides)

#### FieldRuntime Struct

```go
type FieldRuntime struct {
    // Visible: Should field be displayed?
    Visible bool `json:"visible"`
    
    // Editable: Can user edit field?
    Editable bool `json:"editable"`
    
    // Reason: Why is field hidden/readonly?
    // Examples: "permission_required", "insufficient_permissions", 
    //           "tenant_restriction", "workflow_state"
    Reason string `json:"reason,omitempty"`
}
```

**Usage:**
```go
field.Runtime = &FieldRuntime{
    Visible:  user.HasPermission("hr.view_salary"),
    Editable: user.HasPermission("hr.edit_salary"),
    Reason:   "insufficient_permissions",
}
```

### 10.4 Complex Conditions (Condition Package)

For complex conditional logic, use the existing Conditional struct:

```go
type Conditional struct {
    Show     *ConditionGroup `json:"show,omitempty"`     // Show field when true
    Hide     *ConditionGroup `json:"hide,omitempty"`     // Hide field when true
    Required *ConditionGroup `json:"required,omitempty"` // Required when true
    Disabled *ConditionGroup `json:"disabled,omitempty"` // Disabled when true
}

type ConditionGroup struct {
    Logic      string      `json:"logic" validate:"oneof=AND OR"`
    Conditions []Condition `json:"conditions" validate:"dive"`
}

type Condition struct {
    Field    string `json:"field" validate:"required"`
    Operator string `json:"operator" validate:"required"`
    Value    any    `json:"value"`
}
```

**Example:**
```json
{
  "name": "discount_code",
  "type": "text",
  "label": "Discount Code",
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {
          "field": "order_total",
          "operator": "greater",
          "value": 100
        },
        {
          "field": "is_premium",
          "operator": "equals",
          "value": true
        }
      ]
    }
  }
}
```

**Meaning:** Show discount code field only if:
- Order total > $100 AND
- User is premium member

**Backend Evaluation:**
```go
func (f *Field) IsVisible(data map[string]any) bool {
    if f.Hidden {
        return false
    }
    
    if f.Conditional != nil && f.Conditional.Show != nil {
        // Evaluate condition group
        evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
        evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
        
        result, err := evaluator.Evaluate(ctx, f.Conditional.Show, evalCtx)
        if err != nil || !result {
            return false
        }
    }
    
    if f.Conditional != nil && f.Conditional.Hide != nil {
        // Evaluate hide condition
        evaluator := condition.NewEvaluator(nil, condition.DefaultEvalOptions())
        evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
        
        result, err := evaluator.Evaluate(ctx, f.Conditional.Hide, evalCtx)
        if err == nil && result {
            return false
        }
    }
    
    return true
}
```

### 10.5 Decision Matrix: Which Strategy?

| Scenario | Strategy | Implementation |
|----------|----------|----------------|
| Checkbox shows field | Client (showIf) | `"showIf": "agreed === true"` |
| Dropdown value shows field | Client (showIf) | `"showIf": "country === 'US'"` |
| Numeric threshold | Client (showIf) | `"showIf": "age >= 18"` |
| Multiple field logic | Client (showIf) | `"showIf": "a && b \|\| c"` |
| User permission | Server (requirePermission) | `"requirePermission": "view_salary"` |
| User role | Server (enricher) | Check role in enricher |
| Tenant isolation | Server (enricher) | Check tenant in enricher |
| Workflow state | Server (enricher) | Check state in enricher |
| Complex business rule | Server (conditional + condition pkg) | Full condition evaluation |

### 10.6 Complete Examples

#### Example 1: Shipping Address (Client-Side)

```json
{
  "fields": [
    {
      "name": "billing_address",
      "type": "textarea",
      "label": "Billing Address",
      "required": true
    },
    {
      "name": "same_as_billing",
      "type": "checkbox",
      "label": "Shipping address same as billing",
      "default": true
    },
    {
      "name": "shipping_address",
      "type": "textarea",
      "label": "Shipping Address",
      "showIf": "same_as_billing === false",
      "required": false
    }
  ]
}
```

#### Example 2: Salary Field (Server-Side Permission)

```json
{
  "fields": [
    {
      "name": "employee_name",
      "type": "text",
      "label": "Employee Name",
      "required": true
    },
    {
      "name": "department",
      "type": "select",
      "label": "Department",
      "required": true
    },
    {
      "name": "salary",
      "type": "currency",
      "label": "Annual Salary",
      "requirePermission": "hr.view_salary"
    },
    {
      "name": "bonus",
      "type": "currency",
      "label": "Annual Bonus",
      "requirePermission": "hr.view_compensation"
    }
  ]
}
```

**Enrichment:**
```go
// Regular user: salary and bonus hidden
user := &User{Permissions: []string{"hr.view_employees"}}
enriched, _ := enricher.Enrich(ctx, schema, user)
// salary.Runtime.Visible = false
// bonus.Runtime.Visible = false

// HR manager: salary and bonus visible
hrManager := &User{Permissions: []string{"hr.view_salary", "hr.view_compensation"}}
enriched, _ := enricher.Enrich(ctx, schema, hrManager)
// salary.Runtime.Visible = true
// bonus.Runtime.Visible = true
```

#### Example 3: Complex Discount Logic (Condition Package)

```json
{
  "fields": [
    {
      "name": "order_total",
      "type": "currency",
      "label": "Order Total",
      "readonly": true
    },
    {
      "name": "customer_tier",
      "type": "select",
      "label": "Customer Tier",
      "options": [
        {"value": "basic", "label": "Basic"},
        {"value": "premium", "label": "Premium"},
        {"value": "vip", "label": "VIP"}
      ]
    },
    {
      "name": "discount_code",
      "type": "text",
      "label": "Discount Code",
      "conditional": {
        "show": {
          "logic": "OR",
          "conditions": [
            {
              "field": "order_total",
              "operator": "greater_or_equal",
              "value": 100
            },
            {
              "field": "customer_tier",
              "operator": "in",
              "value": ["premium", "vip"]
            }
          ]
        }
      }
    }
  ]
}
```

**Meaning:** Show discount code if:
- Order total >= $100, OR
- Customer is premium or VIP

#### Example 4: Multi-Level Conditional

```json
{
  "fields": [
    {
      "name": "employment_type",
      "type": "select",
      "label": "Employment Type",
      "options": [
        {"value": "full_time", "label": "Full Time"},
        {"value": "part_time", "label": "Part Time"},
        {"value": "contractor", "label": "Contractor"}
      ]
    },
    {
      "name": "benefits",
      "type": "multiselect",
      "label": "Benefits",
      "showIf": "employment_type === 'full_time'",
      "options": [
        {"value": "health", "label": "Health Insurance"},
        {"value": "dental", "label": "Dental Insurance"},
        {"value": "401k", "label": "401(k)"}
      ]
    },
    {
      "name": "hourly_rate",
      "type": "currency",
      "label": "Hourly Rate",
      "showIf": "employment_type === 'part_time' || employment_type === 'contractor'",
      "required": true
    },
    {
      "name": "annual_salary",
      "type": "currency",
      "label": "Annual Salary",
      "showIf": "employment_type === 'full_time'",
      "required": true,
      "requirePermission": "hr.view_salary"
    }
  ]
}
```

**Logic:**
- Full time → Shows benefits and salary (if has permission)
- Part time/Contractor → Shows hourly rate
- All show employment type (always visible)

### 10.7 Best Practices

#### Use Client-Side for UX

```json
// ✅ GOOD: Client-side for instant feedback
{
  "name": "other_reason",
  "type": "textarea",
  "label": "Please specify",
  "showIf": "reason === 'other'"
}
```

#### Use Server-Side for Security

```json
// ✅ GOOD: Server-side for sensitive data
{
  "name": "ssn",
  "type": "text",
  "label": "Social Security Number",
  "requirePermission": "hr.view_ssn"
}

// ❌ WRONG: Client-side for sensitive data
{
  "name": "ssn",
  "type": "text",
  "label": "Social Security Number",
  "showIf": "user_role === 'admin'"  // Can be bypassed!
}
```

#### Combine Both Strategies

```json
{
  "fields": [
    {
      "name": "employment_status",
      "type": "select",
      "label": "Employment Status"
    },
    {
      "name": "termination_date",
      "type": "date",
      "label": "Termination Date",
      "showIf": "employment_status === 'terminated'",  // Client: UX
      "requirePermission": "hr.manage_terminations"    // Server: Security
    }
  ]
}
```

**Result:**
- Field only shows when status is "terminated" (instant)
- But only if user has permission (secure)
- Both checks applied

#### Keep Expressions Simple

```json
// ✅ GOOD: Simple, readable
"showIf": "age >= 18"
"showIf": "country === 'US' && state === 'CA'"

// ❌ BAD: Complex, hard to maintain
"showIf": "((age >= 18 && (country === 'US' || country === 'CA')) || (is_premium && verified)) && !is_deleted"
```

**If complex, use Conditional with condition package instead.**

---

