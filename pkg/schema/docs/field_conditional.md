# Conditional Field Logic - Dual Format Support

## Overview

Field conditionals support **two formats** to balance ease of use with advanced capabilities:

1. **Simple Format** (Recommended) - Easy for backend developers writing JSON
2. **Advanced Format** - Full power of the condition package for complex scenarios

## Simple Format (Recommended)

### Structure

```go
type ConditionGroup struct {
    Logic      string      // "AND" or "OR"
    Conditions []Condition // List of simple conditions
}

type Condition struct {
    Field    string // Field name to check
    Operator string // Comparison operator
    Value    any    // Value to compare against
}
```

### JSON Example

```json
{
  "name": "discount_percentage",
  "type": "number",
  "label": "Discount %",
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {
          "field": "discount_type",
          "operator": "equal",
          "value": "percentage"
        },
        {
          "field": "is_active",
          "operator": "equal",
          "value": true
        }
      ]
    }
  }
}
```

### Supported Operators

| Operator | Aliases | Description | Example |
|----------|---------|-------------|---------|
| `equal` | `equals`, `==`, `eq` | Exact match | `status == "active"` |
| `not_equal` | `not_equals`, `!=`, `neq` | Not equal | `status != "draft"` |
| `greater` | `greater_than`, `>`, `gt` | Greater than | `age > 18` |
| `greater_or_equal` | `>=`, `gte` | Greater or equal | `age >= 18` |
| `less` | `less_than`, `<`, `lt` | Less than | `price < 100` |
| `less_or_equal` | `<=`, `lte` | Less or equal | `price <= 100` |
| `between` | - | In range (inclusive) | `age between [18, 65]` |
| `not_between` | - | Outside range | `age not between [0, 17]` |
| `contains` | - | String contains | `email contains "@"` |
| `not_contains` | - | String doesn't contain | `name not contains "test"` |
| `starts_with` | - | String starts with | `code starts_with "INV"` |
| `ends_with` | - | String ends with | `file ends_with ".pdf"` |
| `in` | `select_any_in` | Value in list | `status in ["active", "pending"]` |
| `not_in` | `select_not_any_in` | Value not in list | `status not in ["deleted"]` |
| `is_empty` | - | Value is null/empty | `notes is_empty` |
| `is_not_empty` | - | Value has content | `email is_not_empty` |
| `match_regexp` | `regex` | Regex pattern match | `phone match_regexp "^\\+1"` |

### Logic Types

- **`AND`** - All conditions must be true
- **`OR`** - At least one condition must be true

### Complete Examples

#### Show Field When Status is Active

```json
{
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {
          "field": "status",
          "operator": "equal",
          "value": "active"
        }
      ]
    }
  }
}
```

#### Hide Field When User is Guest

```json
{
  "conditional": {
    "hide": {
      "logic": "AND",
      "conditions": [
        {
          "field": "user_type",
          "operator": "equal",
          "value": "guest"
        }
      ]
    }
  }
}
```

#### Required When Amount Over 1000

```json
{
  "conditional": {
    "required": {
      "logic": "AND",
      "conditions": [
        {
          "field": "amount",
          "operator": "greater",
          "value": 1000
        }
      ]
    }
  }
}
```

#### Multiple Conditions with OR

```json
{
  "conditional": {
    "show": {
      "logic": "OR",
      "conditions": [
        {
          "field": "is_premium",
          "operator": "equal",
          "value": true
        },
        {
          "field": "is_vip",
          "operator": "equal",
          "value": true
        }
      ]
    }
  }
}
```

#### Complex Multi-Condition

```json
{
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {
          "field": "age",
          "operator": "greater_or_equal",
          "value": 18
        },
        {
          "field": "country",
          "operator": "in",
          "value": ["US", "CA", "UK"]
        },
        {
          "field": "email",
          "operator": "is_not_empty",
          "value": null
        }
      ]
    }
  }
}
```

---

## Advanced Format (For Complex Scenarios)

When you need:
- Nested condition groups (groups within groups)
- Negation (NOT logic)
- Formula-based conditions
- Custom functions

Use the advanced format with full condition package power.

### Structure

```go
// Uses condition package directly
type Conditional struct {
    ShowAdvanced     *condition.ConditionGroup
    HideAdvanced     *condition.ConditionGroup
    RequiredAdvanced *condition.ConditionGroup
    DisabledAdvanced *condition.ConditionGroup
}
```

### JSON Example

```json
{
  "name": "advanced_field",
  "type": "text",
  "conditional": {
    "showAdvanced": {
      "conjunction": "and",
      "rules": [
        {
          "id": "rule-1",
          "left": {
            "type": "field",
            "field": "status"
          },
          "op": "equal",
          "right": "active"
        }
      ],
      "groups": [
        {
          "conjunction": "or",
          "rules": [
            {
              "id": "rule-2",
              "left": {
                "type": "field",
                "field": "premium"
              },
              "op": "equal",
              "right": true
            },
            {
              "id": "rule-3",
              "left": {
                "type": "field",
                "field": "vip"
              },
              "op": "equal",
              "right": true
            }
          ]
        }
      ]
    }
  }
}
```

### With Formulas

```json
{
  "conditional": {
    "showAdvanced": {
      "conjunction": "and",
      "rules": [
        {
          "id": "formula-1",
          "left": {
            "type": "formula",
            "formula": "age >= 18 && age <= 65 && status == 'active'"
          },
          "op": "equal",
          "right": true
        }
      ]
    }
  }
}
```

### With Negation

```json
{
  "conditional": {
    "showAdvanced": {
      "conjunction": "and",
      "not": true,
      "rules": [
        {
          "id": "rule-1",
          "left": {
            "type": "field",
            "field": "deleted"
          },
          "op": "equal",
          "right": true
        }
      ]
    }
  }
}
```

---

## When to Use Which Format

### Use Simple Format When:
- ✅ Basic field comparisons
- ✅ Single level of conditions
- ✅ AND/OR logic is sufficient
- ✅ Backend developers writing JSON
- ✅ 90% of use cases

### Use Advanced Format When:
- ✅ Nested condition groups required
- ✅ Need formula-based conditions
- ✅ Need NOT logic (negation)
- ✅ Need custom functions
- ✅ Complex business rules
- ✅ 10% of use cases

---

## How It Works Internally

The simple format is automatically converted to the advanced format:

```go
// Simple format in JSON
{
  "logic": "AND",
  "conditions": [
    {"field": "status", "operator": "equal", "value": "active"}
  ]
}

// Converted internally to:
condition.NewBuilder(condition.ConjunctionAnd).
    AddRule("status", condition.OpEqual, "active").
    Build()
```

This happens transparently - you don't need to worry about it!

---

## Mixing Simple and Advanced

You can use simple format for some conditions and advanced for others:

```json
{
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {"field": "type", "operator": "equal", "value": "individual"}
      ]
    },
    "requiredAdvanced": {
      "conjunction": "and",
      "rules": [
        {
          "id": "complex-rule",
          "left": {
            "type": "formula",
            "formula": "income > 50000 && creditScore >= 700"
          },
          "op": "equal",
          "right": true
        }
      ]
    }
  }
}
```

---

## Common Patterns

### Pattern 1: Show/Hide Based on Select Value

```json
{
  "name": "shipping_address",
  "type": "group",
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {
          "field": "delivery_method",
          "operator": "equal",
          "value": "shipping"
        }
      ]
    }
  }
}
```

### Pattern 2: Required When Another Field Has Value

```json
{
  "name": "state",
  "type": "select",
  "conditional": {
    "required": {
      "logic": "AND",
      "conditions": [
        {
          "field": "country",
          "operator": "equal",
          "value": "US"
        }
      ]
    }
  }
}
```

### Pattern 3: Multiple Conditions (All Must Be True)

```json
{
  "name": "discount_code",
  "type": "text",
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {
          "field": "is_logged_in",
          "operator": "equal",
          "value": true
        },
        {
          "field": "total_amount",
          "operator": "greater",
          "value": 100
        },
        {
          "field": "first_purchase",
          "operator": "equal",
          "value": false
        }
      ]
    }
  }
}
```

### Pattern 4: Multiple Conditions (Any Can Be True)

```json
{
  "name": "urgent_contact",
  "type": "phone",
  "conditional": {
    "show": {
      "logic": "OR",
      "conditions": [
        {
          "field": "priority",
          "operator": "equal",
          "value": "high"
        },
        {
          "field": "status",
          "operator": "equal",
          "value": "urgent"
        },
        {
          "field": "risk_level",
          "operator": "greater",
          "value": 7
        }
      ]
    }
  }
}
```

### Pattern 5: Disabled Until Prerequisite Completed

```json
{
  "name": "submit_button",
  "type": "button",
  "conditional": {
    "disabled": {
      "logic": "OR",
      "conditions": [
        {
          "field": "terms_accepted",
          "operator": "not_equal",
          "value": true
        },
        {
          "field": "email_verified",
          "operator": "not_equal",
          "value": true
        }
      ]
    }
  }
}
```

---

## Validation

The simple format validates operators and field names:

```go
field.Validate(ctx) // Returns error if:
// - Unknown operator used
// - Field name is empty
// - Logic is not AND or OR
```

---

## Performance

**Simple Format:**
- ✅ Fast conversion to advanced format
- ✅ Cached after first conversion
- ✅ No performance penalty

**Advanced Format:**
- ✅ Direct evaluation (no conversion)
- ✅ Formula compilation cached
- ✅ Optimized for complex rules

Both formats have excellent performance!

---

## Migration Guide

### If You Were Using Old ConditionGroup

**Good news:** Your existing JSON still works! No changes needed.

```json
// This still works exactly as before
{
  "conditional": {
    "show": {
      "logic": "AND",
      "conditions": [
        {"field": "status", "operator": "equal", "value": "active"}
      ]
    }
  }
}
```

### If You Want to Use Advanced Features

Just add the `Advanced` suffix:

```json
{
  "conditional": {
    "showAdvanced": {
      "conjunction": "and",
      "rules": [...]
    }
  }
}
```

---

## Best Practices

1. **Start Simple** - Use simple format for 90% of cases
2. **Upgrade When Needed** - Switch to advanced for complex scenarios
3. **Document Formulas** - Add comments explaining complex formulas
4. **Test Conditions** - Test edge cases in your condition logic
5. **Use Descriptive Operators** - Use `equal` instead of `==` for clarity
6. **Validate Early** - Call `field.Validate()` during schema loading
7. **Handle Errors** - Check errors from `IsVisible()`, `IsRequired()`

---

## Troubleshooting

### "Unknown operator"
**Problem:** Operator string not recognized  
**Solution:** Check supported operators table above

### "Condition evaluator not set"
**Problem:** Evaluator not injected  
**Solution:** Call `field.SetEvaluator(evaluator)`

### "Error evaluating condition"
**Problem:** Field referenced in condition doesn't exist  
**Solution:** Check field name spelling in data

### Condition not working
**Problem:** Logic or operator issue  
**Solution:** Test with simple case first, add conditions incrementally

---

## Examples Repository

See `examples/conditional-fields/` for more examples:
- Basic show/hide
- Complex multi-condition
- Formula-based conditions
- Custom function conditions
- Performance benchmarks

---

## Related Documentation

- [Condition Package Docs](condition-package.md)
- [Field Types Reference](field-types.md)
- [Validation Guide](validation.md)
- [Formula Guide](formulas.md)
