# Textarea Control Component

**FILE PURPOSE**: Multi-line text input control for longer text content and descriptions  
**SCOPE**: All textarea variants, character limits, auto-sizing, and validation patterns  
**TARGET AUDIENCE**: Developers implementing text input areas, forms, and content editing interfaces

## 📋 Component Overview

Textarea Control provides multi-line text input functionality with features like character counting, auto-resizing, validation, and advanced text formatting options. Essential for capturing longer user input in ERP forms.

### Schema Reference
- **Primary Schema**: `TextareaControlSchema.json`
- **Related Schemas**: `FormHorizontal.json`, `SchemaValidation.json`
- **Base Interface**: Form input control for multi-line text

## Basic Usage

```json
{
    "type": "textarea",
    "name": "description",
    "label": "Description",
    "placeholder": "Enter detailed description..."
}
```

## Go Type Definition

```go
type TextareaControlProps struct {
    Type            string              `json:"type"`
    Name            string              `json:"name"`
    Label           interface{}         `json:"label"`
    Placeholder     string              `json:"placeholder"`
    Value           string              `json:"value"`
    Required        bool                `json:"required"`
    Disabled        bool                `json:"disabled"`
    ReadOnly        bool                `json:"readOnly"`
    Rows            int                 `json:"rows"`
    MinRows         int                 `json:"minRows"`
    MaxRows         int                 `json:"maxRows"`
    MaxLength       int                 `json:"maxLength"`
    ShowCounter     bool                `json:"showCounter"`
    Clearable       bool                `json:"clearable"`
    BorderMode      string              `json:"borderMode"`     // "full", "half", "none"
    Size            string              `json:"size"`           // "xs", "sm", "md", "lg", "full"
    AutoFill        interface{}         `json:"autoFill"`
}
```

## Essential Variants

### Basic Textarea
```json
{
    "type": "textarea",
    "name": "comments",
    "label": "Comments",
    "placeholder": "Enter your comments...",
    "rows": 4,
    "required": true
}
```

### Character Limited Textarea
```json
{
    "type": "textarea",
    "name": "description",
    "label": "Product Description",
    "placeholder": "Describe the product features...",
    "maxLength": 500,
    "showCounter": true,
    "rows": 6,
    "clearable": true
}
```

### Auto-Resizing Textarea
```json
{
    "type": "textarea",
    "name": "notes",
    "label": "Meeting Notes",
    "placeholder": "Enter meeting notes...",
    "minRows": 3,
    "maxRows": 10,
    "clearable": true
}
```

### Large Content Textarea
```json
{
    "type": "textarea",
    "name": "content",
    "label": "Article Content",
    "placeholder": "Write your article content...",
    "rows": 12,
    "maxLength": 5000,
    "showCounter": true,
    "size": "lg",
    "borderMode": "full"
}
```

### Compact Textarea
```json
{
    "type": "textarea",
    "name": "quick_note",
    "label": "Quick Note",
    "placeholder": "Add a quick note...",
    "rows": 2,
    "size": "sm",
    "borderMode": "half",
    "clearable": true
}
```

## Real-World Use Cases

### Employee Feedback Form
```json
{
    "type": "textarea",
    "name": "feedback",
    "label": "Employee Feedback",
    "placeholder": "Please provide your feedback on performance, goals, and development areas...",
    "rows": 8,
    "maxLength": 2000,
    "showCounter": true,
    "required": true,
    "validations": {
        "minLength": 50,
        "maxLength": 2000
    },
    "validationErrors": {
        "minLength": "Feedback must be at least 50 characters",
        "maxLength": "Feedback cannot exceed 2000 characters"
    }
}
```

### Project Description
```json
{
    "type": "textarea",
    "name": "project_description",
    "label": "Project Description",
    "placeholder": "Provide a detailed description of the project scope, objectives, and deliverables...",
    "minRows": 5,
    "maxRows": 15,
    "maxLength": 3000,
    "showCounter": true,
    "required": true,
    "size": "lg"
}
```

### Support Ticket Details
```json
{
    "type": "textarea",
    "name": "issue_description",
    "label": "Issue Description",
    "placeholder": "Please describe the issue in detail, including steps to reproduce...",
    "rows": 6,
    "maxLength": 1500,
    "showCounter": true,
    "required": true,
    "validations": {
        "minLength": 20
    },
    "hint": "Include as much detail as possible to help us resolve your issue quickly"
}
```

### Product Review
```json
{
    "type": "textarea",
    "name": "review_text",
    "label": "Review",
    "placeholder": "Share your experience with this product...",
    "rows": 5,
    "maxLength": 1000,
    "showCounter": true,
    "clearable": true,
    "borderMode": "full"
}
```

### Invoice Notes
```json
{
    "type": "textarea",
    "name": "invoice_notes",
    "label": "Additional Notes",
    "placeholder": "Add any special instructions or notes for this invoice...",
    "rows": 3,
    "maxLength": 500,
    "showCounter": true,
    "clearable": true,
    "size": "md"
}
```

### Document Summary
```json
{
    "type": "textarea",
    "name": "document_summary",
    "label": "Document Summary",
    "placeholder": "Provide a brief summary of the document content...",
    "minRows": 4,
    "maxRows": 8,
    "maxLength": 800,
    "showCounter": true,
    "required": true
}
```

This component provides essential multi-line text input functionality for ERP forms and data entry scenarios requiring longer text content.