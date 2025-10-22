# Root Template

**FILE PURPOSE**: Application root template that serves as the top-level container for entire ERP applications  
**SCOPE**: Application initialization, global settings, theme configuration, and top-level routing  
**TARGET AUDIENCE**: Developers setting up new ERP applications and configuring application-wide settings

## 📋 Component Overview

Root serves as the foundation template for ERP applications, providing application-level configuration, theme management, and global context initialization. It acts as the entry point that wraps all other components and manages application-wide state.

### Schema Reference
- **Primary Schema**: `RootSchema.json` (references `PageSchema`)
- **Base Interface**: Application container template
- **Architecture Role**: Top-level application wrapper

## Basic Usage

```json
{
    "type": "page",
    "title": "AWO ERP System",
    "initApi": "/api/app/initialize",
    "body": [
        {
            "type": "nav",
            "links": [
                {"label": "Dashboard", "to": "/dashboard"},
                {"label": "Customers", "to": "/customers"},
                {"label": "Orders", "to": "/orders"}
            ]
        },
        {
            "type": "service",
            "api": "/api/dashboard/stats",
            "body": [
                {"$ref": "#/definitions/DashboardContent"}
            ]
        }
    ]
}
```

## Go Type Definition

```go
// RootSchema directly references PageSchema
type RootApplicationProps struct {
    Type                    string              `json:"type"`
    Title                   string              `json:"title"`
    SubTitle               string              `json:"subTitle"`
    Body                   []interface{}       `json:"body"`
    InitApi                string              `json:"initApi"`
    Definitions            map[string]interface{} `json:"definitions"`
    CSS                    map[string]interface{} `json:"css"`
    Data                   map[string]interface{} `json:"data"`
}
```

## Application Architecture Patterns

### Single-Page Application (SPA) Structure
```json
{
    "type": "page",
    "title": "AWO Enterprise ERP",
    "initApi": "/api/app/bootstrap",
    
    "css": {
        ":root": {
            "--primary-color": "#3b82f6",
            "--secondary-color": "#6b7280",
            "--success-color": "#10b981",
            "--warning-color": "#f59e0b",
            "--error-color": "#ef4444"
        }
    },
    
    "data": {
        "app_name": "AWO ERP",
        "version": "2.0.0",
        "environment": "production"
    },
    
    "body": [
        {
            "type": "nav",
            "className": "bg-white shadow-sm border-b",
            "links": [
                {
                    "label": "Dashboard",
                    "icon": "home",
                    "to": "/dashboard",
                    "activeOn": "${page.pathname === '/dashboard'}"
                },
                {
                    "label": "Sales",
                    "icon": "trending-up",
                    "children": [
                        {"label": "Customers", "to": "/customers"},
                        {"label": "Leads", "to": "/leads"},
                        {"label": "Quotes", "to": "/quotes"},
                        {"label": "Orders", "to": "/orders"}
                    ]
                },
                {
                    "label": "Finance",
                    "icon": "dollar-sign",
                    "children": [
                        {"label": "Invoices", "to": "/invoices"},
                        {"label": "Payments", "to": "/payments"},
                        {"label": "Reports", "to": "/financial-reports"}
                    ]
                }
            ]
        },
        {
            "type": "switch-container",
            "className": "app-content flex-1",
            "items": [
                {
                    "test": "${page.pathname === '/dashboard'}",
                    "schema": {"$ref": "#/definitions/DashboardPage"}
                },
                {
                    "test": "${page.pathname.startsWith('/customers')}",
                    "schema": {"$ref": "#/definitions/CustomerPages"}
                },
                {
                    "test": "${page.pathname.startsWith('/orders')}",
                    "schema": {"$ref": "#/definitions/OrderPages"}
                }
            ]
        }
    ],
    
    "definitions": {
        "DashboardPage": {
            "type": "page",
            "title": "Dashboard",
            "body": [
                {
                    "type": "service",
                    "api": "/api/dashboard/overview",
                    "body": [
                        {"$ref": "#/definitions/DashboardStats"},
                        {"$ref": "#/definitions/RecentActivity"}
                    ]
                }
            ]
        },
        
        "CustomerPages": {
            "type": "switch-container",
            "items": [
                {
                    "test": "${page.pathname === '/customers'}",
                    "schema": {"$ref": "#/definitions/CustomerListPage"}
                },
                {
                    "test": "${page.pathname.match(/\\/customers\\/\\d+/)}",
                    "schema": {"$ref": "#/definitions/CustomerDetailPage"}
                }
            ]
        }
    }
}
```

### Multi-Tenant Application Structure
```json
{
    "type": "page",
    "title": "${tenant.company_name} - AWO ERP",
    "initApi": "/api/tenant/${tenant_id}/bootstrap",
    
    "css": {
        ":root": {
            "--primary-color": "${tenant.theme.primary_color}",
            "--logo-url": "url('${tenant.logo_url}')"
        },
        ".tenant-branding": {
            "background-color": "${tenant.theme.primary_color}",
            "color": "${tenant.theme.text_color}"
        }
    },
    
    "data": {
        "tenant_id": "${tenant_id}",
        "user_permissions": [],
        "app_features": []
    },
    
    "body": [
        {
            "type": "div",
            "className": "tenant-header tenant-branding p-4",
            "body": [
                {
                    "type": "div",
                    "className": "flex items-center justify-between",
                    "body": [
                        {
                            "type": "image",
                            "src": "${tenant.logo_url}",
                            "alt": "${tenant.company_name} Logo",
                            "className": "h-8"
                        },
                        {
                            "type": "dropdown-button",
                            "label": "${user.name}",
                            "className": "user-menu",
                            "buttons": [
                                {"type": "button", "label": "Profile", "actionType": "link", "link": "/profile"},
                                {"type": "button", "label": "Settings", "actionType": "link", "link": "/settings"},
                                {"type": "divider"},
                                {"type": "button", "label": "Logout", "actionType": "ajax", "api": "post:/api/auth/logout"}
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "type": "nav",
            "className": "main-navigation",
            "visibleOn": "${user.permissions.length > 0}",
            "links": [
                {
                    "label": "Dashboard",
                    "to": "/dashboard",
                    "visibleOn": "${user.permissions.includes('dashboard.view')}"
                },
                {
                    "label": "Customer Management",
                    "children": [
                        {"label": "Customers", "to": "/customers", "visibleOn": "${user.permissions.includes('customers.view')}"},
                        {"label": "Leads", "to": "/leads", "visibleOn": "${user.permissions.includes('leads.view')}"}
                    ]
                }
            ]
        },
        {
            "type": "service",
            "api": "/api/tenant/${tenant_id}/route-handler",
            "data": {
                "pathname": "${page.pathname}",
                "query": "${page.query}"
            },
            "body": [
                {
                    "type": "switch-container",
                    "items": [
                        {
                            "test": "${route.component === 'dashboard'}",
                            "schema": {"$ref": "#/definitions/TenantDashboard"}
                        },
                        {
                            "test": "${route.component === 'customers'}",
                            "schema": {"$ref": "#/definitions/CustomerManagement"}
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Progressive Web App (PWA) Structure
```json
{
    "type": "page",
    "title": "AWO ERP Mobile",
    "className": "mobile-app",
    "useMobileUI": true,
    
    "css": {
        ".mobile-app": {
            "height": "100vh",
            "display": "flex",
            "flex-direction": "column"
        },
        ".mobile-header": {
            "position": "sticky",
            "top": "0",
            "z-index": "50"
        },
        ".mobile-content": {
            "flex": "1",
            "overflow-y": "auto"
        },
        ".mobile-nav": {
            "position": "fixed",
            "bottom": "0",
            "left": "0",
            "right": "0"
        }
    },
    
    "pullRefresh": {
        "disabled": false,
        "pullingText": "Pull to refresh...",
        "loosingText": "Release to refresh..."
    },
    
    "body": [
        {
            "type": "div",
            "className": "mobile-header bg-blue-600 text-white p-4",
            "body": [
                {
                    "type": "div",
                    "className": "flex items-center justify-between",
                    "body": [
                        {
                            "type": "text",
                            "text": "AWO ERP",
                            "className": "text-lg font-bold"
                        },
                        {
                            "type": "button",
                            "icon": "bell",
                            "level": "link",
                            "className": "text-white",
                            "actionType": "dialog"
                        }
                    ]
                }
            ]
        },
        {
            "type": "div",
            "className": "mobile-content",
            "body": [
                {
                    "type": "switch-container",
                    "items": [
                        {
                            "test": "${page.pathname === '/mobile/dashboard'}",
                            "schema": {"$ref": "#/definitions/MobileDashboard"}
                        },
                        {
                            "test": "${page.pathname === '/mobile/customers'}",
                            "schema": {"$ref": "#/definitions/MobileCustomers"}
                        }
                    ]
                }
            ]
        },
        {
            "type": "nav",
            "className": "mobile-nav bg-white border-t",
            "mode": "horizontal",
            "links": [
                {"label": "Home", "icon": "home", "to": "/mobile/dashboard"},
                {"label": "Customers", "icon": "users", "to": "/mobile/customers"},
                {"label": "Orders", "icon": "shopping-cart", "to": "/mobile/orders"},
                {"label": "More", "icon": "menu", "to": "/mobile/menu"}
            ]
        }
    ]
}
```

## Real-World Application Examples

### Manufacturing ERP Root
```json
{
    "type": "page",
    "title": "AWO Manufacturing ERP",
    "initApi": "/api/manufacturing/bootstrap",
    
    "data": {
        "modules": ["production", "inventory", "quality", "maintenance"],
        "shift_info": {},
        "production_alerts": []
    },
    
    "css": {
        ":root": {
            "--manufacturing-primary": "#2563eb",
            "--production-green": "#16a34a",
            "--alert-red": "#dc2626",
            "--warning-yellow": "#ca8a04"
        }
    },
    
    "body": [
        {
            "type": "div",
            "className": "manufacturing-header bg-blue-600 text-white p-4",
            "body": [
                {
                    "type": "grid",
                    "columns": [
                        {
                            "md": 4,
                            "body": [
                                {
                                    "type": "text",
                                    "text": "AWO Manufacturing",
                                    "className": "text-xl font-bold"
                                }
                            ]
                        },
                        {
                            "md": 4,
                            "body": [
                                {
                                    "type": "service",
                                    "api": "/api/manufacturing/shift-status",
                                    "interval": 30000,
                                    "body": [
                                        {
                                            "type": "text",
                                            "text": "Shift: ${current_shift} | Line Status: ${active_lines}/${total_lines}",
                                            "className": "text-center"
                                        }
                                    ]
                                }
                            ]
                        },
                        {
                            "md": 4,
                            "body": [
                                {
                                    "type": "service",
                                    "api": "/api/manufacturing/alerts-count",
                                    "interval": 10000,
                                    "body": [
                                        {
                                            "type": "badge",
                                            "text": "${critical_alerts} Critical Alerts",
                                            "level": "${critical_alerts > 0 ? 'danger' : 'success'}",
                                            "className": "float-right"
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "type": "nav",
            "className": "manufacturing-nav",
            "links": [
                {
                    "label": "Production Dashboard",
                    "icon": "activity",
                    "to": "/production/dashboard"
                },
                {
                    "label": "Work Orders",
                    "icon": "clipboard",
                    "to": "/production/work-orders",
                    "badge": {
                        "mode": "text",
                        "text": "${pending_work_orders}",
                        "className": "bg-orange-500"
                    }
                },
                {
                    "label": "Inventory",
                    "icon": "package",
                    "children": [
                        {"label": "Raw Materials", "to": "/inventory/raw-materials"},
                        {"label": "WIP", "to": "/inventory/work-in-progress"},
                        {"label": "Finished Goods", "to": "/inventory/finished-goods"}
                    ]
                },
                {
                    "label": "Quality Control",
                    "icon": "check-circle",
                    "to": "/quality/inspections"
                },
                {
                    "label": "Maintenance",
                    "icon": "tool",
                    "to": "/maintenance/schedule"
                }
            ]
        },
        {
            "type": "switch-container",
            "className": "manufacturing-content",
            "items": [
                {
                    "test": "${page.pathname.startsWith('/production')}",
                    "schema": {"$ref": "#/definitions/ProductionModule"}
                },
                {
                    "test": "${page.pathname.startsWith('/inventory')}",
                    "schema": {"$ref": "#/definitions/InventoryModule"}
                },
                {
                    "test": "${page.pathname.startsWith('/quality')}",
                    "schema": {"$ref": "#/definitions/QualityModule"}
                }
            ]
        }
    ],
    
    "definitions": {
        "ProductionModule": {
            "type": "page",
            "title": "Production Management",
            "body": [
                {
                    "type": "service",
                    "api": "/api/production/real-time-data",
                    "ws": "wss://api.awo-erp.com/production/live",
                    "interval": 5000,
                    "body": [
                        {"$ref": "#/definitions/ProductionDashboard"}
                    ]
                }
            ]
        }
    }
}
```

### Financial Services ERP Root
```json
{
    "type": "page",
    "title": "AWO Financial ERP",
    "initApi": "/api/financial/bootstrap",
    
    "css": {
        ":root": {
            "--financial-primary": "#059669",
            "--financial-secondary": "#0d9488",
            "--profit-green": "#16a34a",
            "--loss-red": "#dc2626"
        },
        ".financial-header": {
            "background": "linear-gradient(135deg, #059669 0%, #0d9488 100%)"
        }
    },
    
    "body": [
        {
            "type": "div",
            "className": "financial-header text-white p-6",
            "body": [
                {
                    "type": "grid",
                    "columns": [
                        {
                            "md": 6,
                            "body": [
                                {
                                    "type": "text",
                                    "text": "AWO Financial Services",
                                    "className": "text-2xl font-bold mb-2"
                                },
                                {
                                    "type": "service",
                                    "api": "/api/financial/market-status",
                                    "interval": 60000,
                                    "body": [
                                        {
                                            "type": "text",
                                            "text": "Market: ${market_status} | Last Update: ${last_update | fromNow}",
                                            "className": "text-sm opacity-90"
                                        }
                                    ]
                                }
                            ]
                        },
                        {
                            "md": 6,
                            "body": [
                                {
                                    "type": "service",
                                    "api": "/api/financial/portfolio-summary",
                                    "interval": 300000,
                                    "body": [
                                        {
                                            "type": "grid",
                                            "columns": [
                                                {
                                                    "md": 4,
                                                    "body": [
                                                        {
                                                            "type": "stats",
                                                            "title": "Portfolio Value",
                                                            "value": "$${total_value | number}",
                                                            "className": "text-center text-white"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "md": 4,
                                                    "body": [
                                                        {
                                                            "type": "stats",
                                                            "title": "Daily P&L",
                                                            "value": "${daily_pnl >= 0 ? '+' : ''}$${daily_pnl | number}",
                                                            "className": "text-center ${daily_pnl >= 0 ? 'text-green-200' : 'text-red-200'}"
                                                        }
                                                    ]
                                                },
                                                {
                                                    "md": 4,
                                                    "body": [
                                                        {
                                                            "type": "stats",
                                                            "title": "Active Positions",
                                                            "value": "${active_positions}",
                                                            "className": "text-center text-white"
                                                        }
                                                    ]
                                                }
                                            ]
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            ]
        },
        {
            "type": "tabs",
            "className": "financial-tabs",
            "tabs": [
                {
                    "title": "Dashboard",
                    "icon": "pie-chart",
                    "body": [
                        {"$ref": "#/definitions/FinancialDashboard"}
                    ]
                },
                {
                    "title": "Portfolio",
                    "icon": "briefcase",
                    "body": [
                        {"$ref": "#/definitions/PortfolioManagement"}
                    ]
                },
                {
                    "title": "Trading",
                    "icon": "trending-up",
                    "body": [
                        {"$ref": "#/definitions/TradingInterface"}
                    ]
                },
                {
                    "title": "Risk Management",
                    "icon": "shield",
                    "body": [
                        {"$ref": "#/definitions/RiskDashboard"}
                    ]
                },
                {
                    "title": "Reports",
                    "icon": "file-text",
                    "body": [
                        {"$ref": "#/definitions/FinancialReports"}
                    ]
                }
            ]
        }
    ]
}
```

## Application Configuration Patterns

### Theme and Branding Configuration
```json
{
    "css": {
        ":root": {
            "--primary-color": "${theme.primary_color || '#3b82f6'}",
            "--secondary-color": "${theme.secondary_color || '#6b7280'}",
            "--success-color": "${theme.success_color || '#10b981'}",
            "--warning-color": "${theme.warning_color || '#f59e0b'}",
            "--error-color": "${theme.error_color || '#ef4444'}",
            "--border-radius": "${theme.border_radius || '0.375rem'}",
            "--font-family": "${theme.font_family || 'Inter, sans-serif'}"
        },
        ".custom-brand": {
            "background-color": "var(--primary-color)",
            "color": "white"
        }
    }
}
```

### Global Data and State Management
```json
{
    "data": {
        "user": {
            "id": null,
            "name": "",
            "email": "",
            "permissions": [],
            "preferences": {}
        },
        "tenant": {
            "id": null,
            "name": "",
            "settings": {},
            "features": []
        },
        "app": {
            "version": "2.0.0",
            "environment": "production",
            "features": [],
            "maintenance_mode": false
        }
    }
}
```

### Real-time Updates Configuration
```json
{
    "interval": 30000,
    "silentPolling": true,
    "stopAutoRefreshWhen": "${app.maintenance_mode === true}",
    "ws": "wss://api.awo-erp.com/realtime/${tenant_id}",
    "onEvent": {
        "websocket:message": {
            "actions": [
                {
                    "actionType": "setValue",
                    "componentId": "global_notifications",
                    "value": "${event.data.notifications}"
                }
            ]
        }
    }
}
```

This Root template serves as the foundation for building comprehensive ERP applications with proper architecture, theming, multi-tenancy support, and real-time capabilities.