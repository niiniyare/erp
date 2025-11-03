# Schema Structure

## Minimal Schema

```json
{
  "id": "my-form",
  "type": "form",
  "title": "My Form"
}
```

## Complete Schema Structure

```go
type Schema struct {
    // Required
    ID    string `json:"id" validate:"required"`
    Type  string `json:"type" validate:"required"`
    Title string `json:"title" validate:"required"`
    
    // Optional
    Version     string   `json:"version,omitempty"`
    Description string   `json:"description,omitempty"`
    Fields      []Field  `json:"fields,omitempty"`
    Actions     []Action `json:"actions,omitempty"`
    
    // Configuration
    Config *Config `json:"config,omitempty"`
    Layout *Layout `json:"layout,omitempty"`
    HTMX   *HTMX   `json:"htmx,omitempty"`
    
    // Enterprise
    Security   *Security   `json:"security,omitempty"`
    Tenant     *Tenant     `json:"tenant,omitempty"`
    Validation *Validation `json:"validation,omitempty"`
}
```

## Field Struct

```go
type Field struct {
    Name        string          `json:"name" validate:"required"`
    Type        string          `json:"type" validate:"required"`
    Label       string          `json:"label" validate:"required"`
    Required    bool            `json:"required,omitempty"`
    Readonly    bool            `json:"readonly,omitempty"`
    Placeholder string          `json:"placeholder,omitempty"`
    Value       any     `json:"value,omitempty"`
    Options     []Option        `json:"options,omitempty"`
    Validation  *FieldValidation `json:"validation,omitempty"`
    Config      map[string]any  `json:"config,omitempty"`
    ShowIf      string          `json:"showIf,omitempty"`
}
```

## Config Struct

```go
type Config struct {
    Action   string            `json:"action,omitempty"`
    Method   string            `json:"method,omitempty"`
    Encoding string            `json:"encoding,omitempty"`
    Timeout  int               `json:"timeout,omitempty"`
    Headers  map[string]string `json:"headers,omitempty"`
}
```

## Complete Example

```json
{
  "id": "user-registration",
  "type": "form",
  "version": "1.0.0",
  "title": "Create User Account",
  "description": "Register a new user",
  
  "config": {
    "action": "/api/users",
    "method": "POST",
    "encoding": "application/json"
  },
  
  "fields": [
    {
      "name": "email",
      "type": "email",
      "label": "Email Address",
      "required": true,
      "validation": {
        "maxLength": 255
      }
    },
    {
      "name": "password",
      "type": "password",
      "label": "Password",
      "required": true,
      "validation": {
        "minLength": 8
      }
    }
  ],
  
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Create Account"
    }
  ],
  
  "htmx": {
    "enabled": true,
    "post": "/api/users",
    "target": "#result"
  }
}
```

## Multi-Tenant Schema

```json
{
  "id": "invoice-form",
  "type": "form",
  "title": "Create Invoice",
  
  "tenant": {
    "enabled": true,
    "field": "tenant_id",
    "isolation": "strict"
  },
  
  "fields": [
    {
      "name": "tenant_id",
      "type": "hidden",
      "value": "${tenant_id}"
    }
  ]
}
```

[← Back](README.md) | [Next: Field Types →](05-field-types.md)