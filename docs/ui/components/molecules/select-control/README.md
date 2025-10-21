# Select Control Component

**FILE PURPOSE**: Dropdown selection control for single and multi-select scenarios  
**SCOPE**: All select types, data sources, search functionality, and advanced selection modes  
**TARGET AUDIENCE**: Developers implementing dropdown selections, multi-select controls, and data-driven options

## 📋 Component Overview

Select Control provides comprehensive dropdown selection functionality with support for static options, API-driven data, search capabilities, multi-selection, and advanced selection modes including table, tree, and associated selections.

### Schema Reference
- **Primary Schema**: `SelectControlSchema.json`
- **Related Schemas**: `Option.json`, `BaseApiObject.json`, `FormHorizontal.json`
- **Base Interface**: Form input control for option selection

## Basic Usage

```json
{
    "type": "select",
    "name": "category",
    "label": "Category",
    "options": [
        {"label": "Option 1", "value": "opt1"},
        {"label": "Option 2", "value": "opt2"}
    ]
}
```

## Go Type Definition

```go
type SelectControlProps struct {
    Type            string              `json:"type"`
    Name            string              `json:"name"`
    Label           interface{}         `json:"label"`
    Options         interface{}         `json:"options"`
    Source          interface{}         `json:"source"`
    Multiple        bool                `json:"multiple"`
    Searchable      bool                `json:"searchable"`
    Clearable       bool                `json:"clearable"`
    Placeholder     string              `json:"placeholder"`
    Required        bool                `json:"required"`
    Disabled        bool                `json:"disabled"`
    JoinValues      bool                `json:"joinValues"`
    ExtractValue    bool                `json:"extractValue"`
    Delimiter       string              `json:"delimiter"`
    SelectMode      string              `json:"selectMode"`     // "table", "group", "tree", "chained", "associated"
    BorderMode      string              `json:"borderMode"`     // "full", "half", "none"
    CheckAll        bool                `json:"checkAll"`
    MaxTagCount     int                 `json:"maxTagCount"`
    AutoComplete    interface{}         `json:"autoComplete"`
    SearchApi       interface{}         `json:"searchApi"`
}
```

## Essential Variants

### Basic Dropdown
```json
{
    "type": "select",
    "name": "status",
    "label": "Status",
    "placeholder": "Select status...",
    "options": [
        {"label": "Active", "value": "active"},
        {"label": "Inactive", "value": "inactive"},
        {"label": "Pending", "value": "pending"}
    ],
    "clearable": true
}
```

### Multi-Select
```json
{
    "type": "select",
    "name": "skills",
    "label": "Skills",
    "placeholder": "Select skills...",
    "multiple": true,
    "checkAll": true,
    "extractValue": true,
    "options": [
        {"label": "JavaScript", "value": "js"},
        {"label": "Python", "value": "python"},
        {"label": "Go", "value": "go"},
        {"label": "React", "value": "react"}
    ]
}
```

### API-Driven Select
```json
{
    "type": "select",
    "name": "department",
    "label": "Department",
    "placeholder": "Choose department...",
    "source": "/api/departments",
    "searchable": true,
    "clearable": true
}
```

### Searchable with AutoComplete
```json
{
    "type": "select",
    "name": "employee",
    "label": "Employee",
    "placeholder": "Search employees...",
    "searchable": true,
    "autoComplete": "/api/employees/search?q=${term}",
    "clearable": true
}
```

### Advanced Table Mode
```json
{
    "type": "select",
    "name": "products",
    "label": "Products",
    "multiple": true,
    "selectMode": "table",
    "source": "/api/products",
    "columns": [
        {"name": "name", "label": "Product Name"},
        {"name": "sku", "label": "SKU"},
        {"name": "price", "label": "Price"}
    ],
    "searchable": true,
    "checkAll": true
}
```

## Real-World Use Cases

### Employee Selection
```json
{
    "type": "select",
    "name": "assigned_to",
    "label": "Assign To",
    "placeholder": "Select employee...",
    "source": "/api/employees/active",
    "searchable": true,
    "clearable": true,
    "autoComplete": "/api/employees/search?q=${term}",
    "menuTpl": "${name} (${department})"
}
```

### Department Hierarchy
```json
{
    "type": "select",
    "name": "department_id",
    "label": "Department",
    "selectMode": "tree",
    "source": "/api/departments/tree",
    "searchable": true,
    "clearable": true
}
```

### Category Selection
```json
{
    "type": "select",
    "name": "categories",
    "label": "Categories",
    "multiple": true,
    "selectMode": "group",
    "source": "/api/categories/grouped",
    "checkAll": true,
    "maxTagCount": 3,
    "extractValue": true
}
```

### Country and State Selection
```json
{
    "type": "select",
    "name": "location",
    "label": "Location",
    "selectMode": "chained",
    "source": "/api/locations/chained",
    "searchable": true
}
```

This component provides essential dropdown selection functionality for ERP forms and data entry scenarios.