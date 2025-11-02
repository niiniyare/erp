# Layout System

## Layout Types

### Grid Layout

```json
{
  "layout": {
    "type": "grid",
    "columns": 2,
    "gap": "1rem"
  }
}
```

Renders fields in 2-column grid.

### Tabs Layout

```json
{
  "layout": {
    "type": "tabs",
    "tabs": [
      {
        "id": "personal",
        "label": "Personal Info",
        "fields": ["first_name", "last_name", "email"]
      },
      {
        "id": "address",
        "label": "Address",
        "fields": ["street", "city", "zip"]
      }
    ]
  }
}
```

### Steps Layout (Wizard)

```json
{
  "layout": {
    "type": "steps",
    "steps": [
      {
        "id": "step1",
        "title": "Account Details",
        "fields": ["email", "password"]
      },
      {
        "id": "step2",
        "title": "Personal Info",
        "fields": ["first_name", "last_name"]
      }
    ]
  }
}
```

### Sections

```json
{
  "layout": {
    "type": "sections",
    "sections": [
      {
        "id": "basic",
        "title": "Basic Information",
        "fields": ["name", "email"],
        "collapsible": true
      }
    ]
  }
}
```

## Responsive Layout

```json
{
  "layout": {
    "type": "grid",
    "columns": 2,
    "responsive": {
      "tablet": 1,
      "mobile": 1
    }
  }
}
```

[← Back](06-validation.md) | [Next: Conditional Logic →](08-conditional-logic.md)