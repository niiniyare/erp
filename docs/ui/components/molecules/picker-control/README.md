# Picker Control Component

**FILE PURPOSE**: Generic picker control for custom data selection interfaces  
**SCOPE**: Custom pickers, data selection, entity picking, and specialized selection workflows  
**TARGET AUDIENCE**: Developers implementing custom selection interfaces, entity pickers, and specialized data selection features

## 📋 Component Overview

Picker Control provides a generic, configurable picker interface for selecting custom data types with support for search, filtering, pagination, multi-selection, and custom rendering. Essential for specialized selection scenarios in ERP systems.

### Schema Reference
- **Primary Schema**: `PickerControlSchema.json`
- **Related Schemas**: `Option.json`, `BaseApiObject.json`, `PickerCondition.json`
- **Base Interface**: Form input control for custom data selection

## Basic Usage

```json
{
    "type": "picker",
    "name": "selected_items",
    "label": "Select Items",
    "placeholder": "Choose items...",
    "source": "/api/picker/data"
}
```

## Go Type Definition

```go
type PickerControlProps struct {
    Type               string              `json:"type"`
    Name               string              `json:"name"`
    Label              interface{}         `json:"label"`
    Placeholder        string              `json:"placeholder"`
    Options            interface{}         `json:"options"`
    Source             interface{}         `json:"source"`
    Multiple           bool                `json:"multiple"`
    Clearable          bool                `json:"clearable"`
    SearchEnable       bool                `json:"searchEnable"`
    SearchPrompt       string              `json:"searchPrompt"`
    SearchPlaceholder  string              `json:"searchPlaceholder"`
    SearchApi          interface{}         `json:"searchApi"`
    Embed              bool                `json:"embed"`
    ModalMode          string              `json:"modalMode"`        // "dialog", "drawer"
    ModalTitle         string              `json:"modalTitle"`
    Size               string              `json:"size"`             // Modal size
    PickerSchema       interface{}         `json:"pickerSchema"`     // Custom schema
    ValueField         string              `json:"valueField"`
    LabelField         string              `json:"labelField"`
    InitAutoFill       bool                `json:"initAutoFill"`
    AutoFill           interface{}         `json:"autoFill"`
    Conditions         interface{}         `json:"conditions"`       // Filter conditions
    OrderBy            string              `json:"orderBy"`          // Sort order
    OrderDir           string              `json:"orderDir"`         // Sort direction
    PerPage            int                 `json:"perPage"`          // Pagination
    HeaderToolbar      interface{}         `json:"headerToolbar"`    // Toolbar config
    FooterToolbar      interface{}         `json:"footerToolbar"`    // Footer config
}
```

## Essential Variants

### Basic Entity Picker
```json
{
    "type": "picker",
    "name": "selected_entity",
    "label": "Select Entity",
    "placeholder": "Choose entity...",
    "source": "/api/entities",
    "valueField": "id",
    "labelField": "name",
    "searchEnable": true,
    "clearable": true
}
```

### Multi-Selection Picker
```json
{
    "type": "picker",
    "name": "selected_items",
    "label": "Select Multiple Items",
    "placeholder": "Choose items...",
    "source": "/api/items",
    "multiple": true,
    "valueField": "id",
    "labelField": "title",
    "searchEnable": true,
    "clearable": true
}
```

### Modal Picker
```json
{
    "type": "picker",
    "name": "selected_record",
    "label": "Select Record",
    "placeholder": "Browse records...",
    "source": "/api/records",
    "modalMode": "dialog",
    "modalTitle": "Select Record",
    "size": "lg",
    "searchEnable": true,
    "perPage": 20
}
```

### Embedded Picker
```json
{
    "type": "picker",
    "name": "inline_selection",
    "label": "Select Option",
    "source": "/api/options",
    "embed": true,
    "searchEnable": true,
    "perPage": 10
}
```

## Real-World Use Cases

### Customer Picker
```json
{
    "type": "picker",
    "name": "customer_id",
    "label": "Select Customer",
    "placeholder": "Search and select customer...",
    "source": "/api/customers",
    "valueField": "id",
    "labelField": "company_name",
    "searchEnable": true,
    "searchPlaceholder": "Search by company name, email, or phone...",
    "modalMode": "dialog",
    "modalTitle": "Select Customer",
    "size": "lg",
    "clearable": true,
    "required": true,
    "pickerSchema": {
        "type": "crud",
        "syncLocation": false,
        "api": "/api/customers",
        "filter": {
            "body": [
                {"type": "input-text", "name": "company_name", "label": "Company", "placeholder": "Search by company name"},
                {"type": "input-text", "name": "email", "label": "Email", "placeholder": "Search by email"},
                {"type": "select", "name": "status", "label": "Status", "options": ["active", "inactive", "pending"]}
            ]
        },
        "columns": [
            {"name": "company_name", "label": "Company Name", "sortable": true},
            {"name": "contact_person", "label": "Contact Person"},
            {"name": "email", "label": "Email"},
            {"name": "phone", "label": "Phone"},
            {"name": "status", "label": "Status", "type": "status"}
        ]
    },
    "autoFill": {
        "api": "/api/customers/${value}",
        "fillMapping": {
            "customer_name": "company_name",
            "customer_email": "email",
            "customer_phone": "phone",
            "billing_address": "address"
        }
    }
}
```

### Product Picker
```json
{
    "type": "picker",
    "name": "product_ids",
    "label": "Select Products",
    "placeholder": "Add products to order...",
    "source": "/api/products",
    "multiple": true,
    "valueField": "id",
    "labelField": "name",
    "searchEnable": true,
    "searchPlaceholder": "Search products by name, SKU, or category...",
    "modalMode": "drawer",
    "modalTitle": "Product Catalog",
    "size": "lg",
    "clearable": true,
    "conditions": {
        "status": "active",
        "in_stock": true
    },
    "pickerSchema": {
        "type": "cards",
        "api": "/api/products",
        "filter": {
            "body": [
                {"type": "input-text", "name": "name", "label": "Product Name"},
                {"type": "input-text", "name": "sku", "label": "SKU"},
                {"type": "select", "name": "category_id", "label": "Category", "source": "/api/categories"},
                {"type": "range", "name": "price_range", "label": "Price Range", "min": 0, "max": 10000}
            ]
        },
        "card": {
            "header": {"title": "${name}", "subTitle": "SKU: ${sku}"},
            "body": [
                {"type": "image", "src": "${image_url}", "defaultImage": "/images/product-placeholder.png"},
                {"type": "text", "text": "${description}"},
                {"type": "text", "text": "Price: $${price}", "className": "font-bold"},
                {"type": "text", "text": "Stock: ${stock_quantity}", "visibleOn": "${stock_quantity > 0}"}
            ]
        }
    }
}
```

### Employee Picker
```json
{
    "type": "picker",
    "name": "assigned_employees",
    "label": "Assign Employees",
    "placeholder": "Select employees for this task...",
    "source": "/api/employees",
    "multiple": true,
    "valueField": "id",
    "labelField": "full_name",
    "searchEnable": true,
    "modalMode": "dialog",
    "modalTitle": "Select Employees",
    "size": "lg",
    "conditions": {
        "status": "active",
        "department_id": "${department_id}"
    },
    "pickerSchema": {
        "type": "table",
        "api": "/api/employees",
        "columns": [
            {"name": "avatar", "label": "", "type": "image", "width": "60px"},
            {"name": "full_name", "label": "Name", "sortable": true},
            {"name": "department", "label": "Department"},
            {"name": "position", "label": "Position"},
            {"name": "skills", "label": "Skills", "type": "tags"},
            {"name": "availability", "label": "Availability", "type": "status"}
        ]
    }
}
```

### Vendor Picker
```json
{
    "type": "picker",
    "name": "vendor_id",
    "label": "Select Vendor",
    "placeholder": "Choose vendor for this purchase...",
    "source": "/api/vendors",
    "valueField": "id",
    "labelField": "company_name",
    "searchEnable": true,
    "modalMode": "dialog",
    "modalTitle": "Vendor Directory",
    "size": "lg",
    "clearable": true,
    "required": true,
    "conditions": {
        "status": "approved",
        "category": "${procurement_category}"
    },
    "pickerSchema": {
        "type": "crud",
        "api": "/api/vendors",
        "filter": {
            "body": [
                {"type": "input-text", "name": "company_name", "label": "Company Name"},
                {"type": "select", "name": "category", "label": "Category", "source": "/api/vendor-categories"},
                {"type": "select", "name": "location", "label": "Location", "source": "/api/locations"},
                {"type": "range", "name": "rating", "label": "Rating", "min": 1, "max": 5}
            ]
        },
        "columns": [
            {"name": "company_name", "label": "Company", "sortable": true},
            {"name": "category", "label": "Category"},
            {"name": "location", "label": "Location"},
            {"name": "rating", "label": "Rating", "type": "rating"},
            {"name": "last_order_date", "label": "Last Order", "type": "date"},
            {"name": "total_orders", "label": "Total Orders"}
        ]
    },
    "autoFill": {
        "api": "/api/vendors/${value}/details",
        "fillMapping": {
            "vendor_name": "company_name",
            "vendor_contact": "primary_contact",
            "vendor_email": "email",
            "payment_terms": "default_payment_terms"
        }
    }
}
```

### Project Picker
```json
{
    "type": "picker",
    "name": "project_id",
    "label": "Select Project",
    "placeholder": "Choose project...",
    "source": "/api/projects",
    "valueField": "id",
    "labelField": "name",
    "searchEnable": true,
    "modalMode": "dialog",
    "modalTitle": "Project Selection",
    "size": "lg",
    "conditions": {
        "status": ["active", "planning"],
        "assigned_to": "${current_user_id}"
    },
    "pickerSchema": {
        "type": "cards",
        "api": "/api/projects",
        "card": {
            "header": {
                "title": "${name}",
                "subTitle": "${client_name}",
                "avatar": {"type": "icon", "icon": "project", "className": "bg-blue-500"}
            },
            "body": [
                {"type": "text", "text": "${description}"},
                {"type": "progress", "value": "${completion_percentage}", "label": "Progress"},
                {"type": "text", "text": "Due: ${due_date}", "className": "text-sm text-gray-600"},
                {"type": "tags", "value": "${technologies}", "label": "Technologies"}
            ],
            "actions": [
                {"type": "button", "label": "View Details", "actionType": "link", "link": "/projects/${id}"}
            ]
        }
    }
}
```

### Asset Picker
```json
{
    "type": "picker",
    "name": "asset_ids",
    "label": "Select Assets",
    "placeholder": "Choose assets for maintenance...",
    "source": "/api/assets",
    "multiple": true,
    "valueField": "id",
    "labelField": "asset_name",
    "searchEnable": true,
    "modalMode": "dialog",
    "modalTitle": "Asset Selection",
    "size": "lg",
    "conditions": {
        "status": "active",
        "location_id": "${facility_id}"
    },
    "pickerSchema": {
        "type": "table",
        "api": "/api/assets",
        "filter": {
            "body": [
                {"type": "input-text", "name": "asset_name", "label": "Asset Name"},
                {"type": "select", "name": "category", "label": "Category", "source": "/api/asset-categories"},
                {"type": "select", "name": "location", "label": "Location", "source": "/api/locations"},
                {"type": "select", "name": "condition", "label": "Condition", "options": ["excellent", "good", "fair", "poor"]}
            ]
        },
        "columns": [
            {"name": "asset_tag", "label": "Asset Tag", "sortable": true},
            {"name": "asset_name", "label": "Name", "sortable": true},
            {"name": "category", "label": "Category"},
            {"name": "location", "label": "Location"},
            {"name": "condition", "label": "Condition", "type": "status"},
            {"name": "last_maintenance", "label": "Last Maintenance", "type": "date"},
            {"name": "next_maintenance", "label": "Next Due", "type": "date"}
        ]
    }
}
```

### Document Picker
```json
{
    "type": "picker",
    "name": "related_documents",
    "label": "Related Documents",
    "placeholder": "Link related documents...",
    "source": "/api/documents",
    "multiple": true,
    "valueField": "id",
    "labelField": "title",
    "searchEnable": true,
    "modalMode": "drawer",
    "modalTitle": "Document Library",
    "size": "lg",
    "pickerSchema": {
        "type": "list",
        "api": "/api/documents",
        "filter": {
            "body": [
                {"type": "input-text", "name": "title", "label": "Document Title"},
                {"type": "select", "name": "document_type", "label": "Type", "source": "/api/document-types"},
                {"type": "select", "name": "department", "label": "Department", "source": "/api/departments"},
                {"type": "date", "name": "created_date", "label": "Created Date"}
            ]
        },
        "listItem": {
            "title": "${title}",
            "subTitle": "${document_type} • Created: ${created_date}",
            "avatar": {"type": "icon", "icon": "${file_icon}"},
            "desc": "${description}",
            "actions": [
                {"type": "button", "label": "Preview", "actionType": "dialog", "dialog": {"title": "Document Preview", "body": {"type": "iframe", "src": "${preview_url}"}}}
            ]
        }
    }
}
```

### Location Picker
```json
{
    "type": "picker",
    "name": "delivery_location",
    "label": "Delivery Location",
    "placeholder": "Select delivery location...",
    "source": "/api/locations",
    "valueField": "id",
    "labelField": "full_address",
    "searchEnable": true,
    "modalMode": "dialog",
    "modalTitle": "Location Selection",
    "size": "lg",
    "pickerSchema": {
        "type": "cards",
        "api": "/api/locations",
        "card": {
            "header": {
                "title": "${name}",
                "subTitle": "${type}",
                "avatar": {"type": "icon", "icon": "location", "className": "bg-green-500"}
            },
            "body": [
                {"type": "text", "text": "${full_address}"},
                {"type": "text", "text": "Contact: ${contact_person}"},
                {"type": "text", "text": "Phone: ${phone}"},
                {"type": "text", "text": "Hours: ${operating_hours}"}
            ]
        }
    },
    "autoFill": {
        "api": "/api/locations/${value}/details",
        "fillMapping": {
            "delivery_address": "full_address",
            "delivery_contact": "contact_person",
            "delivery_phone": "phone",
            "delivery_instructions": "special_instructions"
        }
    }
}
```

This component provides essential picker functionality for ERP systems requiring custom data selection, entity picking, and specialized selection workflows with rich interfaces.