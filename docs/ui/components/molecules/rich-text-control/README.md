# Rich Text Control Component

**FILE PURPOSE**: Rich text editor control for content creation and editing  
**SCOPE**: Text editing with formatting, media insertion, and content management  
**TARGET AUDIENCE**: Developers implementing content management and text editing features

## 📋 Component Overview

Rich text editor component for content creation with formatting tools, media insertion, and HTML output.

### Schema Reference
- **Primary Schema**: `RichTextControlSchema.json`
- **Base Interface**: Form input control for formatted text content

## Basic Usage

```json
{
    "type": "input-rich-text",
    "name": "content",
    "label": "Content",
    "vendor": "tinymce"
}
```

## Go Type Definition

```go
type RichTextControlProps struct {
    Type            string      `json:"type"`
    Name            string      `json:"name"`
    Label           interface{} `json:"label"`
    Vendor          string      `json:"vendor"`           // "froala" or "tinymce"
    Receiver        interface{} `json:"receiver"`         // Image upload API
    VideoReceiver   interface{} `json:"videoReceiver"`    // Video upload API
    FileField       string      `json:"fileField"`        // Upload field name
    BorderMode      string      `json:"borderMode"`       // "full", "half", "none"
    Options         interface{} `json:"options"`          // Editor configuration
    Required        bool        `json:"required"`
    Placeholder     string      `json:"placeholder"`
    ReadOnly        bool        `json:"readOnly"`
}
```

## Essential Variants

### Basic Editor
```json
{
    "type": "input-rich-text",
    "name": "description",
    "label": "Description",
    "vendor": "tinymce",
    "required": true
}
```

### Advanced Editor with Media
```json
{
    "type": "input-rich-text",
    "name": "article_content",
    "label": "Article Content",
    "vendor": "tinymce",
    "receiver": "/api/upload/image",
    "videoReceiver": "/api/upload/video",
    "fileField": "file",
    "options": {
        "height": 500,
        "plugins": "image media link lists",
        "toolbar": "bold italic underline | link image media | bullist numlist"
    }
}
```

## Real-World Use Cases

- Content management systems
- Employee handbooks and documentation
- Product descriptions
- Email templates
- News and announcements
- Training materials

This component provides essential rich text editing capabilities for ERP content management needs.