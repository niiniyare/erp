# Tag Control Component

**FILE PURPOSE**: Tag/chip input control for multi-value selection and labeling  
**SCOPE**: Tag creation, selection, batch input, dropdown mode, and value management  
**TARGET AUDIENCE**: Developers implementing tagging systems, multi-select inputs, and categorization interfaces

## 📋 Component Overview

Tag Control provides interactive tag/chip input functionality with support for creating, selecting, and managing multiple values through tags. Features include dropdown mode, batch input, custom options, and advanced tag management.

### Schema Reference
- **Primary Schema**: `TagControlSchema.json`
- **Related Schemas**: `Option.json`, `TooltipWrapperSchema.json`
- **Base Interface**: Form input control for tag-based multi-value selection

## Basic Usage

```json
{
    "type": "input-tag",
    "name": "tags",
    "label": "Tags",
    "placeholder": "Add tags...",
    "enableBatchAdd": true
}
```

## Go Type Definition

```go
type TagControlProps struct {
    Type                string              `json:"type"`
    Name                string              `json:"name"`
    Label               interface{}         `json:"label"`
    Placeholder         string              `json:"placeholder"`
    Options             interface{}         `json:"options"`
    Source              interface{}         `json:"source"`
    Multiple            bool                `json:"multiple"`
    Clearable           bool                `json:"clearable"`
    Dropdown            bool                `json:"dropdown"`
    EnableBatchAdd      bool                `json:"enableBatchAdd"`
    Separator           string              `json:"separator"`
    Max                 int                 `json:"max"`
    MaxTagLength        int                 `json:"maxTagLength"`
    MaxTagCount         int                 `json:"maxTagCount"`
    OverflowTagPopover  interface{}         `json:"overflowTagPopover"`
    Creatable           bool                `json:"creatable"`
    CreateBtnLabel      string              `json:"createBtnLabel"`
    JoinValues          bool                `json:"joinValues"`
    Delimiter           string              `json:"delimiter"`
    ExtractValue        bool                `json:"extractValue"`
    OptionsTip          string              `json:"optionsTip"`
}
```

## Essential Variants

### Basic Tag Input
```json
{
    "type": "input-tag",
    "name": "product_tags",
    "label": "Product Tags",
    "placeholder": "Add product tags...",
    "enableBatchAdd": true,
    "separator": ",",
    "clearable": true
}
```

### Predefined Options Dropdown
```json
{
    "type": "input-tag",
    "name": "categories",
    "label": "Categories",
    "placeholder": "Select categories...",
    "dropdown": true,
    "multiple": true,
    "options": [
        {"label": "Electronics", "value": "electronics"},
        {"label": "Clothing", "value": "clothing"},
        {"label": "Home & Garden", "value": "home-garden"},
        {"label": "Sports", "value": "sports"},
        {"label": "Books", "value": "books"}
    ],
    "clearable": true
}
```

### Limited Tag Input
```json
{
    "type": "input-tag",
    "name": "skills",
    "label": "Top Skills",
    "placeholder": "Add up to 5 skills...",
    "enableBatchAdd": true,
    "max": 5,
    "maxTagLength": 20,
    "maxTagCount": 3,
    "overflowTagPopover": {
        "title": "All Skills",
        "placement": "top"
    }
}
```

### API-Driven Tags
```json
{
    "type": "input-tag",
    "name": "departments",
    "label": "Departments",
    "placeholder": "Select departments...",
    "dropdown": true,
    "source": "/api/departments",
    "creatable": true,
    "createBtnLabel": "Add New Department",
    "clearable": true
}
```

### Batch Input Mode
```json
{
    "type": "input-tag",
    "name": "keywords",
    "label": "Keywords",
    "placeholder": "Enter keywords separated by commas or semicolons",
    "enableBatchAdd": true,
    "separator": ",;",
    "max": 10,
    "maxTagLength": 15,
    "optionsTip": "Use commas or semicolons to separate multiple keywords"
}
```

## Real-World Use Cases

### Employee Skills Management
```json
{
    "type": "input-tag",
    "name": "employee_skills",
    "label": "Technical Skills",
    "placeholder": "Add technical skills...",
    "dropdown": true,
    "source": "/api/skills/technical",
    "creatable": true,
    "createBtnLabel": "Add New Skill",
    "enableBatchAdd": true,
    "separator": ",",
    "max": 15,
    "maxTagLength": 30,
    "maxTagCount": 5,
    "overflowTagPopover": {
        "title": "All Technical Skills",
        "placement": "bottom"
    },
    "joinValues": true,
    "extractValue": true
}
```

### Project Labels
```json
{
    "type": "input-tag",
    "name": "project_labels",
    "label": "Project Labels",
    "placeholder": "Add project labels...",
    "enableBatchAdd": true,
    "separator": ",",
    "options": [
        {"label": "High Priority", "value": "high-priority"},
        {"label": "Backend", "value": "backend"},
        {"label": "Frontend", "value": "frontend"},
        {"label": "Database", "value": "database"},
        {"label": "API", "value": "api"},
        {"label": "UI/UX", "value": "ui-ux"},
        {"label": "Testing", "value": "testing"},
        {"label": "Documentation", "value": "documentation"}
    ],
    "creatable": true,
    "max": 8,
    "clearable": true
}
```

### Product Attributes
```json
{
    "type": "input-tag",
    "name": "product_attributes",
    "label": "Product Attributes",
    "placeholder": "Add product attributes...",
    "dropdown": true,
    "source": "/api/product-attributes",
    "enableBatchAdd": true,
    "separator": ",|",
    "maxTagLength": 25,
    "optionsTip": "Select from existing attributes or create new ones",
    "creatable": true,
    "createBtnLabel": "Create Attribute",
    "validations": {
        "minLength": 1
    },
    "validationErrors": {
        "minLength": "At least one attribute is required"
    }
}
```

### Customer Tags
```json
{
    "type": "input-tag",
    "name": "customer_tags",
    "label": "Customer Tags",
    "placeholder": "Tag this customer...",
    "options": [
        {"label": "VIP", "value": "vip"},
        {"label": "Returning Customer", "value": "returning"},
        {"label": "Bulk Buyer", "value": "bulk"},
        {"label": "Corporate", "value": "corporate"},
        {"label": "Individual", "value": "individual"},
        {"label": "International", "value": "international"}
    ],
    "creatable": true,
    "enableBatchAdd": true,
    "max": 6,
    "clearable": true,
    "multiple": true
}
```

### Document Keywords
```json
{
    "type": "input-tag",
    "name": "document_keywords",
    "label": "Document Keywords",
    "placeholder": "Enter keywords for better searchability...",
    "enableBatchAdd": true,
    "separator": ",;",
    "max": 20,
    "maxTagLength": 20,
    "maxTagCount": 8,
    "overflowTagPopover": {
        "title": "All Keywords",
        "placement": "right"
    },
    "hint": "Use relevant keywords to improve document discoverability"
}
```

### Event Topics
```json
{
    "type": "input-tag",
    "name": "event_topics",
    "label": "Event Topics",
    "placeholder": "Select or add topics...",
    "dropdown": true,
    "source": "/api/event-topics",
    "creatable": true,
    "createBtnLabel": "Add Topic",
    "enableBatchAdd": true,
    "separator": ",",
    "max": 10,
    "joinValues": false,
    "extractValue": true,
    "required": true,
    "validations": {
        "minimum": 1
    },
    "validationErrors": {
        "minimum": "Please select at least one topic"
    }
}
```

This component provides essential tag-based input functionality for ERP systems requiring multi-value categorization, labeling, and selection scenarios.