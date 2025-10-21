# Icon Picker Control Component

**FILE PURPOSE**: Icon selection control for interface customization and visual configuration  
**SCOPE**: Icon picking, visual symbols, interface customization, and brand/theme configuration  
**TARGET AUDIENCE**: Developers implementing customizable interfaces, icon selection, and visual configuration features

## 📋 Component Overview

Icon Picker Control provides a specialized interface for selecting icons from icon libraries with support for search, categories, custom icons, and preview functionality. Essential for customizable ERP interfaces and visual configuration.

### Schema Reference
- **Primary Schema**: `IconPickerControlSchema.json`
- **Related Schemas**: `Option.json`, `BaseApiObject.json`
- **Base Interface**: Form input control for icon selection

## Basic Usage

```json
{
    "type": "icon-picker",
    "name": "selected_icon",
    "label": "Select Icon",
    "placeholder": "Choose an icon..."
}
```

## Go Type Definition

```go
type IconPickerControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    Value              string              `json:"value"`
    IconSize           string              `json:"iconSize"`         // Icon display size
    IconClassName      string              `json:"iconClassName"`    // Icon CSS classes
    NoResult           string              `json:"noResult"`         // No results message
    Clearable          bool                `json:"clearable"`
    Disabled           bool                `json:"disabled"`
    ReadOnly           bool                `json:"readOnly"`
    Required           bool                `json:"required"`
    SearchEnable       bool                `json:"searchEnable"`
    SearchKeyword      string              `json:"searchKeyword"`
    Categories         interface{}         `json:"categories"`       // Icon categories
    Icons              interface{}         `json:"icons"`            // Custom icon set
    CustomIcons        interface{}         `json:"customIcons"`      // User custom icons
    DefaultIcon        string              `json:"defaultIcon"`      // Default selection
    ShowPreview        bool                `json:"showPreview"`      // Preview mode
    PreviewSize        string              `json:"previewSize"`      // Preview icon size
    AllowUpload        bool                `json:"allowUpload"`      // Custom upload
    UploadApi          interface{}         `json:"uploadApi"`        // Upload endpoint
    MaxUploadSize      int                 `json:"maxUploadSize"`    // File size limit
    AcceptedFormats    []string            `json:"acceptedFormats"`  // File formats
}
```

## Essential Variants

### Basic Icon Picker
```json
{
    "type": "icon-picker",
    "name": "menu_icon",
    "label": "Menu Icon",
    "placeholder": "Select menu icon...",
    "iconSize": "md",
    "clearable": true,
    "searchEnable": true
}
```

### Categorized Icon Picker
```json
{
    "type": "icon-picker",
    "name": "feature_icon",
    "label": "Feature Icon",
    "placeholder": "Choose feature icon...",
    "iconSize": "lg",
    "searchEnable": true,
    "categories": [
        {"label": "Business", "value": "business"},
        {"label": "Technology", "value": "technology"},
        {"label": "Communication", "value": "communication"},
        {"label": "Finance", "value": "finance"}
    ],
    "clearable": true
}
```

### Custom Icon Upload
```json
{
    "type": "icon-picker",
    "name": "brand_icon",
    "label": "Brand Icon",
    "placeholder": "Select or upload brand icon...",
    "allowUpload": true,
    "uploadApi": "/api/icons/upload",
    "maxUploadSize": 1048576,
    "acceptedFormats": ["svg", "png", "jpg"],
    "showPreview": true,
    "previewSize": "xl"
}
```

### Icon Library Picker
```json
{
    "type": "icon-picker",
    "name": "dashboard_icon",
    "label": "Dashboard Icon",
    "placeholder": "Select dashboard icon...",
    "icons": "/api/icon-library",
    "searchEnable": true,
    "iconSize": "lg",
    "showPreview": true
}
```

## Real-World Use Cases

### Menu Item Configuration
```json
{
    "type": "icon-picker",
    "name": "menu_item_icon",
    "label": "Menu Icon",
    "placeholder": "Select icon for menu item...",
    "iconSize": "md",
    "searchEnable": true,
    "categories": [
        {"label": "Navigation", "value": "navigation"},
        {"label": "Actions", "value": "actions"},
        {"label": "Content", "value": "content"},
        {"label": "Settings", "value": "settings"}
    ],
    "icons": [
        {"category": "navigation", "icons": ["home", "dashboard", "menu", "back", "forward"]},
        {"category": "actions", "icons": ["add", "edit", "delete", "save", "cancel"]},
        {"category": "content", "icons": ["document", "image", "video", "file", "folder"]},
        {"category": "settings", "icons": ["gear", "cog", "preferences", "tools", "admin"]}
    ],
    "defaultIcon": "home",
    "clearable": true,
    "required": true,
    "hint": "Choose an icon that represents this menu item"
}
```

### Dashboard Widget Icon
```json
{
    "type": "icon-picker",
    "name": "widget_icon",
    "label": "Widget Icon",
    "placeholder": "Select widget icon...",
    "iconSize": "lg",
    "searchEnable": true,
    "showPreview": true,
    "previewSize": "xl",
    "categories": [
        {"label": "Analytics", "value": "analytics"},
        {"label": "Finance", "value": "finance"},
        {"label": "Sales", "value": "sales"},
        {"label": "Operations", "value": "operations"}
    ],
    "icons": "/api/widget-icons",
    "defaultIcon": "chart-bar",
    "clearable": true,
    "required": true,
    "autoFill": {
        "api": "/api/icons/${value}/metadata",
        "fillMapping": {
            "icon_name": "name",
            "icon_description": "description",
            "icon_keywords": "tags"
        }
    }
}
```

### Status Indicator Icon
```json
{
    "type": "icon-picker",
    "name": "status_icon",
    "label": "Status Icon",
    "placeholder": "Select status indicator icon...",
    "iconSize": "sm",
    "searchEnable": true,
    "categories": [
        {"label": "Success", "value": "success"},
        {"label": "Warning", "value": "warning"},
        {"label": "Error", "value": "error"},
        {"label": "Info", "value": "info"}
    ],
    "icons": [
        {"category": "success", "icons": ["check", "check-circle", "thumbs-up", "star", "heart"]},
        {"category": "warning", "icons": ["warning", "exclamation", "alert-triangle", "clock", "pause"]},
        {"category": "error", "icons": ["x", "x-circle", "alert-circle", "thumbs-down", "ban"]},
        {"category": "info", "icons": ["info", "info-circle", "question", "lightbulb", "message"]}
    ],
    "iconClassName": "status-icon",
    "clearable": true
}
```

### Department/Team Icon
```json
{
    "type": "icon-picker",
    "name": "department_icon",
    "label": "Department Icon",
    "placeholder": "Select department icon...",
    "iconSize": "lg",
    "searchEnable": true,
    "showPreview": true,
    "categories": [
        {"label": "Management", "value": "management"},
        {"label": "Operations", "value": "operations"},
        {"label": "Technology", "value": "technology"},
        {"label": "Support", "value": "support"}
    ],
    "allowUpload": true,
    "uploadApi": "/api/departments/icons/upload",
    "maxUploadSize": 512000,
    "acceptedFormats": ["svg", "png"],
    "customIcons": "/api/departments/${department_id}/custom-icons",
    "autoFill": {
        "api": "/api/icons/${value}/usage-stats",
        "fillMapping": {
            "icon_popularity": "usage_count",
            "icon_last_used": "last_used_date"
        }
    }
}
```

### Project Category Icon
```json
{
    "type": "icon-picker",
    "name": "project_category_icon",
    "label": "Project Category Icon",
    "placeholder": "Select category icon...",
    "iconSize": "md",
    "searchEnable": true,
    "categories": [
        {"label": "Development", "value": "development"},
        {"label": "Design", "value": "design"},
        {"label": "Marketing", "value": "marketing"},
        {"label": "Research", "value": "research"}
    ],
    "icons": [
        {"category": "development", "icons": ["code", "terminal", "git-branch", "database", "server"]},
        {"category": "design", "icons": ["palette", "brush", "image", "layers", "grid"]},
        {"category": "marketing", "icons": ["megaphone", "trending-up", "users", "mail", "target"]},
        {"category": "research", "icons": ["search", "microscope", "book", "lightbulb", "flask"]}
    ],
    "showPreview": true,
    "clearable": true,
    "validations": {
        "isRequired": true
    },
    "validationErrors": {
        "isRequired": "Please select an icon for this project category"
    }
}
```

### Notification Type Icon
```json
{
    "type": "icon-picker",
    "name": "notification_icon",
    "label": "Notification Icon",
    "placeholder": "Select notification icon...",
    "iconSize": "sm",
    "searchEnable": true,
    "categories": [
        {"label": "System", "value": "system"},
        {"label": "User", "value": "user"},
        {"label": "Business", "value": "business"},
        {"label": "Alert", "value": "alert"}
    ],
    "icons": "/api/notification-icons",
    "iconClassName": "notification-icon",
    "defaultIcon": "bell",
    "clearable": true,
    "hint": "Icon will appear in notification messages"
}
```

### Custom Field Icon
```json
{
    "type": "icon-picker",
    "name": "custom_field_icon",
    "label": "Field Icon",
    "placeholder": "Select field icon...",
    "iconSize": "sm",
    "searchEnable": true,
    "categories": [
        {"label": "Text", "value": "text"},
        {"label": "Number", "value": "number"},
        {"label": "Date", "value": "date"},
        {"label": "Selection", "value": "selection"}
    ],
    "icons": [
        {"category": "text", "icons": ["type", "file-text", "edit", "message-square", "align-left"]},
        {"category": "number", "icons": ["hash", "calculator", "trending-up", "percent", "dollar-sign"]},
        {"category": "date", "icons": ["calendar", "clock", "watch", "sunrise", "sunset"]},
        {"category": "selection", "icons": ["check-square", "list", "toggle-left", "circle", "square"]}
    ],
    "clearable": true
}
```

### Workflow Step Icon
```json
{
    "type": "icon-picker",
    "name": "workflow_step_icon",
    "label": "Step Icon",
    "placeholder": "Select workflow step icon...",
    "iconSize": "md",
    "searchEnable": true,
    "categories": [
        {"label": "Start/End", "value": "endpoints"},
        {"label": "Process", "value": "process"},
        {"label": "Decision", "value": "decision"},
        {"label": "Action", "value": "action"}
    ],
    "icons": [
        {"category": "endpoints", "icons": ["play", "stop", "pause", "flag", "target"]},
        {"category": "process", "icons": ["cog", "settings", "tool", "refresh", "repeat"]},
        {"category": "decision", "icons": ["help-circle", "git-branch", "shuffle", "filter", "sort"]},
        {"category": "action", "icons": ["zap", "send", "save", "download", "upload"]}
    ],
    "showPreview": true,
    "previewSize": "lg",
    "clearable": true
}
```

### Report Type Icon
```json
{
    "type": "icon-picker",
    "name": "report_type_icon",
    "label": "Report Icon",
    "placeholder": "Select report type icon...",
    "iconSize": "lg",
    "searchEnable": true,
    "categories": [
        {"label": "Financial", "value": "financial"},
        {"label": "Analytics", "value": "analytics"},
        {"label": "Performance", "value": "performance"},
        {"label": "Compliance", "value": "compliance"}
    ],
    "icons": "/api/report-icons",
    "showPreview": true,
    "previewSize": "xl",
    "allowUpload": true,
    "uploadApi": "/api/reports/icons/upload",
    "maxUploadSize": 1048576,
    "acceptedFormats": ["svg", "png"],
    "autoFill": {
        "api": "/api/icons/${value}/report-metadata",
        "fillMapping": {
            "icon_category": "category",
            "icon_usage": "common_usage",
            "icon_color_scheme": "recommended_colors"
        }
    }
}
```

### Integration Service Icon
```json
{
    "type": "icon-picker",
    "name": "service_icon",
    "label": "Service Icon",
    "placeholder": "Select integration service icon...",
    "iconSize": "lg",
    "searchEnable": true,
    "customIcons": "/api/integrations/service-icons",
    "allowUpload": true,
    "uploadApi": "/api/integrations/icons/upload",
    "maxUploadSize": 2097152,
    "acceptedFormats": ["svg", "png", "jpg"],
    "showPreview": true,
    "previewSize": "xl",
    "categories": [
        {"label": "Cloud Services", "value": "cloud"},
        {"label": "Databases", "value": "database"},
        {"label": "APIs", "value": "api"},
        {"label": "Communication", "value": "communication"}
    ],
    "hint": "Upload your service logo or select from available icons"
}
```

This component provides essential icon selection functionality for ERP systems requiring visual customization, interface personalization, and brand/theme configuration capabilities.