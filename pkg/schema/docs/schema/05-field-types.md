# Field Types

## Basic Text Types

| Type | HTML Input | Description |
|------|-----------|-------------|
| `text` | `<input type="text">` | Single-line text |
| `email` | `<input type="email">` | Email with validation |
| `password` | `<input type="password">` | Masked password |
| `number` | `<input type="number">` | Numeric input |
| `phone` | `<input type="tel">` | Phone number |
| `url` | `<input type="url">` | URL with validation |
| `hidden` | `<input type="hidden">` | Hidden field |

## Date/Time Types

| Type | HTML Input | Description |
|------|-----------|-------------|
| `date` | `<input type="date">` | Date picker |
| `time` | `<input type="time">` | Time picker |
| `datetime` | `<input type="datetime-local">` | Date + time |
| `daterange` | Two date inputs | Start and end dates |

## Text Content

| Type | HTML Element | Description |
|------|-------------|-------------|
| `textarea` | `<textarea>` | Multi-line text |
| `richtext` | `<div contenteditable>` | WYSIWYG editor |
| `code` | `<textarea>` + syntax highlighting | Code editor |

## Selection Types

| Type | HTML Element | Description |
|------|-------------|-------------|
| `select` | `<select>` | Dropdown list |
| `multiselect` | `<select multiple>` | Multiple selection |
| `radio` | `<input type="radio">` | Radio buttons |
| `checkbox` | `<input type="checkbox">` | Checkboxes |
| `switch` | `<input type="checkbox">` styled | Toggle switch |

## File Types

| Type | HTML Input | Description |
|------|-----------|-------------|
| `file` | `<input type="file">` | Generic file upload |
| `image` | `<input type="file" accept="image/*">` | Image upload |

## Specialized Types

| Type | Description |
|------|-------------|
| `currency` | Money input with formatting |
| `tags` | Tag input (comma-separated) |
| `location` | Address/location picker |
| `relation` | Foreign key relationship |
| `autocomplete` | Autocomplete search |
| `repeatable` | Repeatable field groups |
| `color` | Color picker |
| `rating` | Star rating |
| `slider` | Range slider |

## Usage Examples

### Text Field
```json
{
  "name": "username",
  "type": "text",
  "label": "Username",
  "required": true,
  "validation": {
    "minLength": 3,
    "maxLength": 50
  }
}
```

### Select Field
```json
{
  "name": "country",
  "type": "select",
  "label": "Country",
  "required": true,
  "options": [
    {"value": "us", "label": "United States"},
    {"value": "uk", "label": "United Kingdom"}
  ]
}
```

### Date Field
```json
{
  "name": "birth_date",
  "type": "date",
  "label": "Date of Birth",
  "required": true
}
```

### Textarea
```json
{
  "name": "description",
  "type": "textarea",
  "label": "Description",
  "config": {
    "rows": 5
  }
}
```

### File Upload
```json
{
  "name": "avatar",
  "type": "image",
  "label": "Profile Picture",
  "validation": {
    "maxSize": 5242880,
    "accept": ["image/jpeg", "image/png"]
  }
}
```

### Relation Field
```json
{
  "name": "customer_id",
  "type": "relation",
  "label": "Customer",
  "config": {
    "targetSchema": "customers",
    "displayField": "name",
    "searchFields": ["name", "email"]
  }
}
```

### Repeatable Field
```json
{
  "name": "line_items",
  "type": "repeatable",
  "label": "Invoice Items",
  "fields": [
    {
      "name": "product",
      "type": "text",
      "label": "Product"
    },
    {
      "name": "quantity",
      "type": "number",
      "label": "Qty"
    },
    {
      "name": "price",
      "type": "currency",
      "label": "Price"
    }
  ]
}
```

[← Back](04-schema-structure.md) | [Next: Validation →](06-validation.md)