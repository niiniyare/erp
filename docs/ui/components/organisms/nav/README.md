# Nav Component

**FILE PURPOSE**: Navigation organism for application menus, sidebars, and navigation interfaces  
**SCOPE**: Main navigation, hierarchical menus, sidebar navigation, and navigation trees  
**TARGET AUDIENCE**: Developers implementing application navigation, menu systems, and navigation hierarchies

## 📋 Component Overview

Nav provides comprehensive navigation functionality with support for hierarchical menus, icons, badges, search, and multiple display modes. Essential for creating intuitive navigation experiences in ERP systems.

### Schema Reference
- **Primary Schema**: `NavSchema.json`
- **Related Schemas**: `NavItemSchema.json`, `BadgeObject.json`
- **Base Interface**: Navigation organism for menu systems

## Basic Usage

```json
{
    "type": "nav",
    "stacked": true,
    "links": [
        {"label": "Dashboard", "to": "/dashboard", "icon": "dashboard"},
        {"label": "Customers", "to": "/customers", "icon": "users"},
        {"label": "Products", "to": "/products", "icon": "shopping-bag"}
    ]
}
```

## Go Type Definition

```go
type NavProps struct {
    Type                    string              `json:"type"`
    Links                   []interface{}       `json:"links"`               // Navigation items
    
    // Layout and Display
    Stacked                 bool                `json:"stacked"`             // Vertical or horizontal
    Mode                    string              `json:"mode"`                // "float" or "inline"
    Level                   int                 `json:"level"`               // Max display levels
    DefaultOpenLevel        int                 `json:"defaultOpenLevel"`    // Default expansion level
    Collapsed               bool                `json:"collapsed"`           // Collapsed state
    
    // Data Source
    Source                  interface{}         `json:"source"`              // API source
    DeferAPI                interface{}         `json:"deferApi"`            // Lazy loading API
    
    // Styling and Theme
    ThemeColor              string              `json:"themeColor"`          // "light" or "dark"
    IndentSize              int                 `json:"indentSize"`          // Indentation pixels
    ExpandIcon              interface{}         `json:"expandIcon"`          // Custom expand icon
    ExpandPosition          string              `json:"expandPosition"`      // Icon position
    PopupClassName          string              `json:"popupClassName"`      // Submenu styles
    
    // Interaction
    Accordion               bool                `json:"accordion"`           // Accordion expansion
    Draggable               bool                `json:"draggable"`           // Drag and drop
    DragOnSameLevel         bool                `json:"dragOnSameLevel"`     // Same level dragging
    SaveOrderAPI            interface{}         `json:"saveOrderApi"`        // Save order API
    
    // Search
    Searchable              bool                `json:"searchable"`          // Enable search
    SearchConfig            interface{}         `json:"searchConfig"`        // Search configuration
    
    // Actions and Badges
    ItemActions             []interface{}       `json:"itemActions"`         // Item actions
    ItemBadge               interface{}         `json:"itemBadge"`          // Item badges
    Badge                   interface{}         `json:"badge"`              // Global badge
    
    // Navigation Control
    ShowKey                 string              `json:"showKey"`            // Show specific submenu
    Overflow                interface{}         `json:"overflow"`           // Horizontal overflow
}
```

## Navigation Modes

### Vertical Stacked (Sidebar)
```json
{
    "type": "nav",
    "stacked": true,
    "mode": "inline",
    "themeColor": "dark",
    "collapsed": false,
    "accordion": true,
    "links": [
        {
            "label": "Dashboard",
            "to": "/dashboard",
            "icon": "dashboard"
        },
        {
            "label": "Sales",
            "icon": "trending-up",
            "children": [
                {"label": "Leads", "to": "/sales/leads", "icon": "user-plus"},
                {"label": "Opportunities", "to": "/sales/opportunities", "icon": "target"},
                {"label": "Orders", "to": "/sales/orders", "icon": "shopping-cart"}
            ]
        }
    ]
}
```

### Horizontal Navigation
```json
{
    "type": "nav",
    "stacked": false,
    "mode": "float",
    "themeColor": "light",
    "overflow": {
        "enable": true,
        "maxVisibleCount": 6
    },
    "links": [
        {"label": "Home", "to": "/", "icon": "home"},
        {"label": "Products", "to": "/products", "icon": "shopping-bag"},
        {"label": "Orders", "to": "/orders", "icon": "file-text"},
        {"label": "Customers", "to": "/customers", "icon": "users"},
        {"label": "Reports", "to": "/reports", "icon": "bar-chart"}
    ]
}
```

### Breadcrumb Navigation
```json
{
    "type": "nav",
    "stacked": false,
    "mode": "inline",
    "level": 1,
    "links": [
        {"label": "Home", "to": "/", "icon": "home"},
        {"label": "Customers", "to": "/customers"},
        {"label": "Customer Details", "active": true}
    ]
}
```

## Real-World Use Cases

### ERP Main Sidebar Navigation
```json
{
    "type": "nav",
    "stacked": true,
    "mode": "inline",
    "themeColor": "dark",
    "collapsed": false,
    "accordion": true,
    "searchable": true,
    "defaultOpenLevel": 1,
    "searchConfig": {
        "placeholder": "Search menu...",
        "mini": true,
        "clearable": true,
        "searchImediately": true
    },
    "links": [
        {
            "label": "Dashboard",
            "to": "/dashboard",
            "icon": "dashboard",
            "badge": {
                "mode": "dot",
                "className": "bg-green-500"
            }
        },
        {
            "label": "Sales Management",
            "icon": "trending-up",
            "children": [
                {
                    "label": "Leads",
                    "to": "/sales/leads",
                    "icon": "user-plus",
                    "badge": {
                        "mode": "text",
                        "text": "${pending_leads_count}",
                        "className": "bg-blue-500"
                    }
                },
                {
                    "label": "Opportunities",
                    "to": "/sales/opportunities",
                    "icon": "target"
                },
                {
                    "label": "Quotes",
                    "to": "/sales/quotes",
                    "icon": "file-text"
                },
                {
                    "label": "Orders",
                    "to": "/sales/orders",
                    "icon": "shopping-cart",
                    "badge": {
                        "mode": "text",
                        "text": "${new_orders_count}",
                        "className": "bg-orange-500"
                    }
                },
                {
                    "label": "Invoices",
                    "to": "/sales/invoices",
                    "icon": "receipt"
                }
            ]
        },
        {
            "label": "Customer Management",
            "icon": "users",
            "children": [
                {
                    "label": "Customers",
                    "to": "/customers",
                    "icon": "user"
                },
                {
                    "label": "Contacts",
                    "to": "/customers/contacts",
                    "icon": "address-book"
                },
                {
                    "label": "Customer Groups",
                    "to": "/customers/groups",
                    "icon": "users"
                }
            ]
        },
        {
            "label": "Inventory",
            "icon": "package",
            "children": [
                {
                    "label": "Products",
                    "to": "/inventory/products",
                    "icon": "shopping-bag"
                },
                {
                    "label": "Categories",
                    "to": "/inventory/categories",
                    "icon": "folder"
                },
                {
                    "label": "Stock Management",
                    "to": "/inventory/stock",
                    "icon": "warehouse",
                    "badge": {
                        "mode": "text",
                        "text": "${low_stock_count}",
                        "className": "bg-red-500",
                        "visibleOn": "${low_stock_count > 0}"
                    }
                },
                {
                    "label": "Suppliers",
                    "to": "/inventory/suppliers",
                    "icon": "truck"
                }
            ]
        },
        {
            "label": "Financial",
            "icon": "dollar-sign",
            "children": [
                {
                    "label": "Accounts",
                    "to": "/financial/accounts",
                    "icon": "credit-card"
                },
                {
                    "label": "Transactions",
                    "to": "/financial/transactions",
                    "icon": "repeat"
                },
                {
                    "label": "Reports",
                    "to": "/financial/reports",
                    "icon": "bar-chart"
                },
                {
                    "label": "Tax Management",
                    "to": "/financial/tax",
                    "icon": "percent"
                }
            ]
        },
        {
            "label": "Human Resources",
            "icon": "user-check",
            "children": [
                {
                    "label": "Employees",
                    "to": "/hr/employees",
                    "icon": "user"
                },
                {
                    "label": "Departments",
                    "to": "/hr/departments",
                    "icon": "building"
                },
                {
                    "label": "Payroll",
                    "to": "/hr/payroll",
                    "icon": "dollar-sign"
                },
                {
                    "label": "Time Tracking",
                    "to": "/hr/time-tracking",
                    "icon": "clock"
                }
            ]
        },
        {
            "label": "Reports & Analytics",
            "icon": "bar-chart",
            "children": [
                {
                    "label": "Sales Reports",
                    "to": "/reports/sales",
                    "icon": "trending-up"
                },
                {
                    "label": "Financial Reports",
                    "to": "/reports/financial",
                    "icon": "pie-chart"
                },
                {
                    "label": "Inventory Reports",
                    "to": "/reports/inventory",
                    "icon": "package"
                },
                {
                    "label": "Custom Reports",
                    "to": "/reports/custom",
                    "icon": "file-plus"
                }
            ]
        },
        {
            "label": "Settings",
            "icon": "settings",
            "children": [
                {
                    "label": "Company Profile",
                    "to": "/settings/company",
                    "icon": "building"
                },
                {
                    "label": "User Management",
                    "to": "/settings/users",
                    "icon": "users"
                },
                {
                    "label": "System Settings",
                    "to": "/settings/system",
                    "icon": "tool"
                },
                {
                    "label": "Integrations",
                    "to": "/settings/integrations",
                    "icon": "link"
                }
            ]
        }
    ]
}
```

### Dynamic Navigation with API Source
```json
{
    "type": "nav",
    "stacked": true,
    "mode": "inline",
    "source": "/api/navigation/user-menu",
    "deferApi": "/api/navigation/lazy-load",
    "themeColor": "light",
    "accordion": false,
    "searchable": true,
    "draggable": true,
    "saveOrderApi": "put:/api/navigation/save-order",
    "searchConfig": {
        "placeholder": "Search navigation...",
        "matchFunc": "return item.label.toLowerCase().includes(keywords.toLowerCase()) || (item.description && item.description.toLowerCase().includes(keywords.toLowerCase()));",
        "mini": false,
        "enhance": true,
        "clearable": true
    },
    "itemActions": [
        {
            "type": "button",
            "label": "Edit",
            "icon": "edit",
            "actionType": "dialog",
            "dialog": {
                "title": "Edit Menu Item",
                "body": {
                    "type": "form",
                    "api": "put:/api/navigation/items/${id}",
                    "initApi": "/api/navigation/items/${id}",
                    "body": [
                        {"type": "input-text", "name": "label", "label": "Label", "required": true},
                        {"type": "input-text", "name": "to", "label": "URL"},
                        {"type": "input-text", "name": "icon", "label": "Icon"},
                        {"type": "textarea", "name": "description", "label": "Description"},
                        {"type": "switch", "name": "visible", "label": "Visible"}
                    ]
                }
            }
        },
        {
            "type": "button",
            "label": "Delete",
            "icon": "trash",
            "actionType": "ajax",
            "api": "delete:/api/navigation/items/${id}",
            "confirmText": "Are you sure you want to delete this menu item?",
            "level": "danger"
        }
    ]
}
```

### Mobile-Responsive Navigation
```json
{
    "type": "nav",
    "stacked": false,
    "mode": "float",
    "themeColor": "light",
    "overflow": {
        "enable": true,
        "maxVisibleCount": 4,
        "overflowLabel": "More",
        "overflowIcon": "more-horizontal"
    },
    "links": [
        {
            "label": "Dashboard",
            "to": "/dashboard",
            "icon": "home",
            "className": "lg:inline-flex hidden"
        },
        {
            "label": "Sales",
            "icon": "trending-up",
            "children": [
                {"label": "Leads", "to": "/sales/leads"},
                {"label": "Orders", "to": "/sales/orders"},
                {"label": "Invoices", "to": "/sales/invoices"}
            ]
        },
        {
            "label": "Customers",
            "to": "/customers",
            "icon": "users"
        },
        {
            "label": "Products",
            "to": "/products",
            "icon": "package"
        },
        {
            "label": "Reports",
            "to": "/reports",
            "icon": "bar-chart",
            "className": "lg:inline-flex hidden"
        }
    ]
}
```

### Contextual Navigation with Badges
```json
{
    "type": "nav",
    "stacked": true,
    "mode": "inline",
    "themeColor": "dark",
    "showKey": "${current_module}",
    "links": [
        {
            "key": "sales",
            "label": "Sales Process",
            "icon": "trending-up",
            "children": [
                {
                    "label": "Lead Qualification",
                    "to": "/sales/leads/qualify",
                    "badge": {
                        "mode": "text",
                        "text": "${unqualified_leads}",
                        "className": "bg-yellow-500",
                        "visibleOn": "${unqualified_leads > 0}"
                    }
                },
                {
                    "label": "Opportunity Management",
                    "to": "/sales/opportunities",
                    "badge": {
                        "mode": "text",
                        "text": "${hot_opportunities}",
                        "className": "bg-red-500",
                        "visibleOn": "${hot_opportunities > 0}"
                    }
                },
                {
                    "label": "Quote Generation",
                    "to": "/sales/quotes",
                    "badge": {
                        "mode": "text",
                        "text": "${pending_quotes}",
                        "className": "bg-blue-500"
                    }
                },
                {
                    "label": "Order Processing",
                    "to": "/sales/orders",
                    "badge": {
                        "mode": "text",
                        "text": "${new_orders}",
                        "className": "bg-green-500"
                    }
                },
                {
                    "label": "Invoice Management",
                    "to": "/sales/invoices",
                    "badge": {
                        "mode": "text",
                        "text": "${overdue_invoices}",
                        "className": "bg-red-600",
                        "visibleOn": "${overdue_invoices > 0}"
                    }
                }
            ]
        }
    ]
}
```

### Project Navigation with Status Indicators
```json
{
    "type": "nav",
    "stacked": true,
    "mode": "inline",
    "themeColor": "light",
    "accordion": false,
    "defaultOpenLevel": 2,
    "links": [
        {
            "label": "Active Projects",
            "icon": "folder-open",
            "children": [
                {
                    "label": "ERP Implementation",
                    "to": "/projects/erp-impl",
                    "icon": "code",
                    "badge": {
                        "mode": "text",
                        "text": "85%",
                        "className": "bg-green-500"
                    },
                    "children": [
                        {
                            "label": "Backend Development",
                            "to": "/projects/erp-impl/backend",
                            "badge": {"mode": "dot", "className": "bg-green-500"}
                        },
                        {
                            "label": "Frontend Development",
                            "to": "/projects/erp-impl/frontend",
                            "badge": {"mode": "dot", "className": "bg-yellow-500"}
                        },
                        {
                            "label": "Testing & QA",
                            "to": "/projects/erp-impl/testing",
                            "badge": {"mode": "dot", "className": "bg-red-500"}
                        }
                    ]
                },
                {
                    "label": "Customer Portal",
                    "to": "/projects/customer-portal",
                    "icon": "user",
                    "badge": {
                        "mode": "text",
                        "text": "45%",
                        "className": "bg-blue-500"
                    }
                }
            ]
        },
        {
            "label": "Completed Projects",
            "icon": "folder",
            "children": [
                {
                    "label": "Inventory System",
                    "to": "/projects/inventory-system",
                    "icon": "package",
                    "badge": {
                        "mode": "text",
                        "text": "✓",
                        "className": "bg-green-600"
                    }
                }
            ]
        }
    ]
}
```

This component provides essential navigation functionality for ERP systems requiring hierarchical menus, responsive navigation, and dynamic menu management with search, badges, and contextual awareness.