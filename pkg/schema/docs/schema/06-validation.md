# Validation System

## Two-Layer Validation

1. **Client-side (HTML5)** - Instant feedback
2. **Server-side (Go)** - Security enforcement

## HTML5 Validation Rules

```json
{
  "name": "email",
  "type": "email",
  "required": true,
  "validation": {
    "minLength": 5,
    "maxLength": 255,
    "pattern": "^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$",
    "messages": {
      "required": "Email is required",
      "pattern": "Invalid email format"
    }
  }
}
```

### Available Rules

| Rule | Type | Description |
|------|------|-------------|
| `required` | boolean | Field must have value |
| `minLength` | int | Minimum string length |
| `maxLength` | int | Maximum string length |
| `min` | number | Minimum numeric value |
| `max` | number | Maximum numeric value |
| `pattern` | string | Regex pattern |
| `step` | number | Numeric step value |

## Server-Side Validation

### Uniqueness Check

```json
{
  "validation": {
    "server": {
      "unique": true,
      "uniqueScope": ["tenant_id"]
    }
  }
}
```

Generates query:
```sql
SELECT EXISTS(
  SELECT 1 FROM users 
  WHERE email = $1 
  AND tenant_id = $2
  AND id != $3
)
```

### Custom Validation

```go
validator.AddCustomRule("business_email", func(value any) error {
    email := value.(string)
    if !strings.HasSuffix(email, "@company.com") {
        return errors.New("Must use company email")
    }
    return nil
})
```

### Cross-Field Validation

```json
{
  "validation": {
    "rules": [
      {
        "id": "end-after-start",
        "description": "End date must be after start date",
        "condition": {
          "left": {"type": "field", "field": "end_date"},
          "op": "greater",
          "right": {"type": "field", "field": "start_date"}
        },
        "message": "End date must be after start date"
      }
    ]
  }
}
```

## Validation in Go

```go
// In handler
validator := validate.NewValidator(db)

errors := validator.ValidateData(ctx, schema, data)
if len(errors) > 0 {
    return c.Status(400).JSON(fiber.Map{
        "errors": errors,
    })
}
```

## Error Response Format

```json
{
  "errors": [
    {
      "field": "email",
      "type": "required",
      "message": "Email is required"
    },
    {
      "field": "password",
      "type": "minLength",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

[← Back](05-field-types.md) | [Next: Layout →](07-layout.md)