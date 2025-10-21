# Steps Component

**FILE PURPOSE**: Step indicator organism for multi-step processes and workflow progress tracking  
**SCOPE**: Wizards, multi-step forms, process tracking, and workflow visualization  
**TARGET AUDIENCE**: Developers implementing multi-step processes, onboarding flows, and progress indicators

## 📋 Component Overview

Steps provides comprehensive step indicator functionality with support for different display modes, status indicators, and progress tracking. Essential for guiding users through complex multi-step processes in ERP systems.

### Schema Reference
- **Primary Schema**: `StepsSchema.json`
- **Related Schemas**: `StepSchema.json`, `StepStatus.json`
- **Base Interface**: Step indicator organism for process navigation

## Basic Usage

```json
{
    "type": "steps",
    "value": 1,
    "steps": [
        {"title": "Personal Info", "description": "Enter your details"},
        {"title": "Company Info", "description": "Company details"},
        {"title": "Review", "description": "Review and submit"}
    ]
}
```

## Go Type Definition

```go
type StepsProps struct {
    Type                    string              `json:"type"`
    Steps                   []interface{}       `json:"steps"`               // Step definitions
    
    // Data Source
    Source                  string              `json:"source"`              // API or data mapping
    Name                    string              `json:"name"`                // Variable mapping
    
    // Current State
    Value                   interface{}         `json:"value"`               // Current step index/value
    Status                  interface{}         `json:"status"`              // Step status mapping
    
    // Display Options
    Mode                    string              `json:"mode"`                // "horizontal" or "vertical"
    LabelPlacement          string              `json:"labelPlacement"`      // "horizontal" or "vertical"
    ProgressDot             bool                `json:"progressDot"`         // Dot style indicators
    IconPosition            bool                `json:"iconPosition"`        // Icon position control
}
```

## Display Modes

### Horizontal Steps (Default)
```json
{
    "type": "steps",
    "mode": "horizontal",
    "value": 2,
    "steps": [
        {
            "title": "Customer Info",
            "description": "Basic customer information"
        },
        {
            "title": "Address Details", 
            "description": "Billing and shipping address"
        },
        {
            "title": "Payment Method",
            "description": "Payment and billing setup"
        },
        {
            "title": "Confirmation",
            "description": "Review and confirm"
        }
    ]
}
```

### Vertical Steps
```json
{
    "type": "steps",
    "mode": "vertical",
    "value": 1,
    "labelPlacement": "horizontal",
    "steps": [
        {
            "title": "Requirements Gathering",
            "description": "Define project scope and requirements",
            "icon": "clipboard-list"
        },
        {
            "title": "System Design",
            "description": "Create technical architecture",
            "icon": "sitemap"
        },
        {
            "title": "Development",
            "description": "Implementation and coding",
            "icon": "code"
        },
        {
            "title": "Testing",
            "description": "Quality assurance and testing",
            "icon": "bug"
        },
        {
            "title": "Deployment",
            "description": "Production deployment",
            "icon": "rocket"
        }
    ]
}
```

### Progress Dot Style
```json
{
    "type": "steps",
    "progressDot": true,
    "value": 3,
    "steps": [
        {"title": "Start"},
        {"title": "Processing"},
        {"title": "Review"},
        {"title": "Complete"}
    ]
}
```

## Real-World Use Cases

### Employee Onboarding Process
```json
{
    "type": "steps",
    "mode": "horizontal",
    "value": "${current_step}",
    "status": {
        "0": "${personal_info_status}",
        "1": "${employment_status}", 
        "2": "${documents_status}",
        "3": "${system_access_status}",
        "4": "${completion_status}"
    },
    "steps": [
        {
            "title": "Personal Information",
            "description": "Basic personal details and contact information",
            "icon": "user",
            "subTitle": "Required fields: Name, Email, Phone"
        },
        {
            "title": "Employment Details",
            "description": "Job position, department, and start date",
            "icon": "briefcase",
            "subTitle": "Position, Department, Manager"
        },
        {
            "title": "Documentation",
            "description": "Upload required documents and agreements",
            "icon": "file-text",
            "subTitle": "ID, Tax forms, Contracts"
        },
        {
            "title": "System Access",
            "description": "Create accounts and assign permissions",
            "icon": "key",
            "subTitle": "Email, Software access, Security"
        },
        {
            "title": "Completion",
            "description": "Review information and complete setup",
            "icon": "check-circle",
            "subTitle": "Final review and activation"
        }
    ]
}
```

### Order Processing Workflow
```json
{
    "type": "steps",
    "mode": "horizontal",
    "value": "${order_step}",
    "status": {
        "order_received": "finish",
        "payment_processing": "${payment_status}",
        "inventory_check": "${inventory_status}",
        "fulfillment": "${fulfillment_status}",
        "shipping": "${shipping_status}",
        "delivered": "${delivery_status}"
    },
    "source": "/api/orders/${order_id}/steps",
    "steps": [
        {
            "title": "Order Received",
            "description": "Order placed and received",
            "icon": "shopping-cart",
            "subTitle": "Order #${order_number}"
        },
        {
            "title": "Payment Processing",
            "description": "Payment verification and processing",
            "icon": "credit-card",
            "subTitle": "$${order_total}"
        },
        {
            "title": "Inventory Check",
            "description": "Stock verification and allocation",
            "icon": "package",
            "subTitle": "${item_count} items"
        },
        {
            "title": "Fulfillment",
            "description": "Order picking and packaging",
            "icon": "box",
            "subTitle": "Warehouse processing"
        },
        {
            "title": "Shipping",
            "description": "Package shipped to customer",
            "icon": "truck",
            "subTitle": "Tracking: ${tracking_number}"
        },
        {
            "title": "Delivered",
            "description": "Order delivered to customer",
            "icon": "check-circle",
            "subTitle": "Delivery confirmation"
        }
    ]
}
```

### Project Implementation Timeline
```json
{
    "type": "steps",
    "mode": "vertical",
    "labelPlacement": "horizontal",
    "value": "${current_phase}",
    "status": {
        "planning": "finish",
        "design": "finish", 
        "development": "process",
        "testing": "wait",
        "deployment": "wait",
        "maintenance": "wait"
    },
    "steps": [
        {
            "title": "Project Planning",
            "description": "Requirements analysis and project planning",
            "icon": "clipboard-list",
            "subTitle": "Duration: 2 weeks",
            "body": [
                {"type": "text", "text": "✓ Requirements gathering completed"},
                {"type": "text", "text": "✓ Resource allocation finalized"},
                {"type": "text", "text": "✓ Timeline established"}
            ]
        },
        {
            "title": "System Design",
            "description": "Technical architecture and UI/UX design", 
            "icon": "sitemap",
            "subTitle": "Duration: 3 weeks",
            "body": [
                {"type": "text", "text": "✓ Database schema designed"},
                {"type": "text", "text": "✓ API architecture planned"},
                {"type": "text", "text": "✓ UI mockups approved"}
            ]
        },
        {
            "title": "Development Phase",
            "description": "Backend and frontend development",
            "icon": "code",
            "subTitle": "Duration: 8 weeks",
            "body": [
                {"type": "progress", "value": 65, "showLabel": true},
                {"type": "text", "text": "Backend API: 85% complete"},
                {"type": "text", "text": "Frontend UI: 45% complete"},
                {"type": "text", "text": "Integration: 30% complete"}
            ]
        },
        {
            "title": "Testing & QA",
            "description": "Comprehensive testing and quality assurance",
            "icon": "bug",
            "subTitle": "Duration: 2 weeks",
            "body": [
                {"type": "text", "text": "⏳ Unit testing planned"},
                {"type": "text", "text": "⏳ Integration testing planned"},
                {"type": "text", "text": "⏳ User acceptance testing planned"}
            ]
        },
        {
            "title": "Deployment",
            "description": "Production deployment and go-live",
            "icon": "rocket",
            "subTitle": "Duration: 1 week",
            "body": [
                {"type": "text", "text": "⏳ Staging environment setup"},
                {"type": "text", "text": "⏳ Production deployment"},
                {"type": "text", "text": "⏳ Go-live support"}
            ]
        },
        {
            "title": "Maintenance & Support",
            "description": "Ongoing maintenance and user support",
            "icon": "life-buoy",
            "subTitle": "Ongoing",
            "body": [
                {"type": "text", "text": "⏳ User training"},
                {"type": "text", "text": "⏳ Bug fixes and updates"},
                {"type": "text", "text": "⏳ Performance monitoring"}
            ]
        }
    ]
}
```

### Customer Onboarding Journey
```json
{
    "type": "steps",
    "mode": "horizontal",
    "progressDot": true,
    "value": "${onboarding_step}",
    "status": "${step_statuses}",
    "steps": [
        {
            "title": "Welcome",
            "description": "Account creation and welcome",
            "icon": "user-plus",
            "body": [
                {"type": "text", "text": "Account created successfully"},
                {"type": "text", "text": "Welcome email sent"},
                {"type": "text", "text": "Initial setup completed"}
            ]
        },
        {
            "title": "Profile Setup",
            "description": "Complete company profile information",
            "icon": "building",
            "body": [
                {"type": "form", "body": [
                    {"type": "input-text", "name": "company_name", "label": "Company Name", "required": true},
                    {"type": "select", "name": "industry", "label": "Industry", "source": "/api/industries"},
                    {"type": "textarea", "name": "description", "label": "Company Description"}
                ]}
            ]
        },
        {
            "title": "Team Invitation",
            "description": "Invite team members to join",
            "icon": "users",
            "body": [
                {"type": "combo", "name": "team_members", "label": "Team Members", "multiple": true, "items": [
                    {"type": "input-email", "name": "email", "label": "Email", "required": true},
                    {"type": "select", "name": "role", "label": "Role", "options": ["Admin", "User", "Viewer"]}
                ]}
            ]
        },
        {
            "title": "Integration Setup",
            "description": "Connect external services",
            "icon": "link",
            "body": [
                {"type": "checkboxes", "name": "integrations", "label": "Available Integrations", "options": [
                    {"label": "Google Workspace", "value": "google", "description": "Email and calendar integration"},
                    {"label": "Slack", "value": "slack", "description": "Team communication"},
                    {"label": "Accounting Software", "value": "accounting", "description": "Financial data sync"}
                ]}
            ]
        },
        {
            "title": "Tour & Training",
            "description": "Product tour and training resources",
            "icon": "graduation-cap",
            "body": [
                {"type": "button", "label": "Start Product Tour", "actionType": "dialog", "level": "primary"},
                {"type": "button", "label": "Download User Guide", "actionType": "download", "api": "/api/resources/user-guide.pdf"},
                {"type": "button", "label": "Schedule Training Session", "actionType": "link", "link": "/training/schedule"}
            ]
        },
        {
            "title": "Ready to Go!",
            "description": "Onboarding complete, start using the system",
            "icon": "check-circle",
            "body": [
                {"type": "alert", "level": "success", "body": "Congratulations! Your account is fully set up and ready to use."},
                {"type": "button", "label": "Go to Dashboard", "actionType": "link", "link": "/dashboard", "level": "primary"},
                {"type": "button", "label": "Contact Support", "actionType": "link", "link": "/support"}
            ]
        }
    ]
}
```

### Invoice Processing Workflow
```json
{
    "type": "steps",
    "mode": "horizontal",
    "value": "${invoice_step}",
    "status": {
        "created": "finish",
        "sent": "${sent_status}",
        "viewed": "${viewed_status}",
        "paid": "${payment_status}",
        "completed": "${completion_status}"
    },
    "steps": [
        {
            "title": "Invoice Created",
            "description": "Invoice generated and ready",
            "icon": "file-plus",
            "subTitle": "Invoice #${invoice_number}",
            "body": [
                {"type": "static", "name": "created_date", "label": "Created", "tpl": "${created_date | date:\"MMM DD, YYYY\"}"},
                {"type": "static", "name": "total_amount", "label": "Amount", "tpl": "$${total_amount}"},
                {"type": "static", "name": "created_by", "label": "Created By"}
            ]
        },
        {
            "title": "Sent to Customer",
            "description": "Invoice delivered to customer",
            "icon": "send",
            "subTitle": "Email: ${customer_email}",
            "body": [
                {"type": "static", "name": "sent_date", "label": "Sent Date", "tpl": "${sent_date | date:\"MMM DD, YYYY HH:mm\"}"},
                {"type": "static", "name": "delivery_method", "label": "Method"},
                {"type": "button", "label": "Resend Invoice", "actionType": "ajax", "api": "post:/api/invoices/${id}/resend", "level": "secondary", "size": "sm"}
            ]
        },
        {
            "title": "Customer Viewed",
            "description": "Customer opened the invoice",
            "icon": "eye",
            "subTitle": "Viewed ${view_count} times",
            "body": [
                {"type": "static", "name": "first_viewed", "label": "First Viewed", "tpl": "${first_viewed | date:\"MMM DD, YYYY HH:mm\"}"},
                {"type": "static", "name": "last_viewed", "label": "Last Viewed", "tpl": "${last_viewed | date:\"MMM DD, YYYY HH:mm\"}"},
                {"type": "static", "name": "view_count", "label": "Total Views"}
            ]
        },
        {
            "title": "Payment Received",
            "description": "Invoice payment processed",
            "icon": "dollar-sign",
            "subTitle": "Payment: $${paid_amount}",
            "body": [
                {"type": "static", "name": "payment_date", "label": "Payment Date", "tpl": "${payment_date | date:\"MMM DD, YYYY\"}"},
                {"type": "static", "name": "payment_method", "label": "Payment Method"},
                {"type": "static", "name": "transaction_id", "label": "Transaction ID"}
            ]
        },
        {
            "title": "Completed",
            "description": "Invoice fully processed",
            "icon": "check-circle",
            "subTitle": "Status: ${status}",
            "body": [
                {"type": "static", "name": "completed_date", "label": "Completed", "tpl": "${completed_date | date:\"MMM DD, YYYY\"}"},
                {"type": "button", "label": "Download Receipt", "actionType": "download", "api": "/api/invoices/${id}/receipt.pdf", "level": "primary", "size": "sm"},
                {"type": "button", "label": "View Details", "actionType": "link", "link": "/invoices/${id}", "level": "secondary", "size": "sm"}
            ]
        }
    ]
}
```

### Product Launch Process
```json
{
    "type": "steps",
    "mode": "vertical",
    "labelPlacement": "horizontal",
    "value": "${launch_phase}",
    "status": "${phase_statuses}",
    "steps": [
        {
            "title": "Market Research",
            "description": "Analyze market conditions and competition",
            "icon": "search",
            "subTitle": "Research Phase",
            "body": [
                {"type": "progress", "value": 100, "className": "bg-green-500"},
                {"type": "text", "text": "✓ Competitor analysis completed"},
                {"type": "text", "text": "✓ Target market identified"},
                {"type": "text", "text": "✓ Pricing strategy defined"}
            ]
        },
        {
            "title": "Product Development",
            "description": "Design and develop the product",
            "icon": "cog",
            "subTitle": "Development Phase",
            "body": [
                {"type": "progress", "value": 80, "className": "bg-blue-500"},
                {"type": "text", "text": "✓ Product specification finalized"},
                {"type": "text", "text": "✓ Prototype development: 80%"},
                {"type": "text", "text": "⏳ Testing and refinement ongoing"}
            ]
        },
        {
            "title": "Marketing Campaign",
            "description": "Prepare marketing materials and campaigns",
            "icon": "megaphone",
            "subTitle": "Marketing Phase",
            "body": [
                {"type": "progress", "value": 60, "className": "bg-yellow-500"},
                {"type": "text", "text": "✓ Brand identity created"},
                {"type": "text", "text": "⏳ Website development: 60%"},
                {"type": "text", "text": "⏳ Social media strategy in progress"}
            ]
        },
        {
            "title": "Pre-Launch Testing",
            "description": "Beta testing and feedback collection",
            "icon": "users",
            "subTitle": "Testing Phase",
            "body": [
                {"type": "progress", "value": 30, "className": "bg-orange-500"},
                {"type": "text", "text": "⏳ Beta user recruitment"},
                {"type": "text", "text": "⏳ Testing protocol preparation"},
                {"type": "text", "text": "⏳ Feedback collection system setup"}
            ]
        },
        {
            "title": "Launch Event",
            "description": "Official product launch and announcement",
            "icon": "rocket",
            "subTitle": "Launch Phase",
            "body": [
                {"type": "text", "text": "⏳ Event planning"},
                {"type": "text", "text": "⏳ Press release preparation"},
                {"type": "text", "text": "⏳ Launch day coordination"}
            ]
        },
        {
            "title": "Post-Launch Monitoring",
            "description": "Monitor performance and gather feedback",
            "icon": "bar-chart",
            "subTitle": "Monitoring Phase",
            "body": [
                {"type": "text", "text": "⏳ Performance tracking setup"},
                {"type": "text", "text": "⏳ Customer feedback analysis"},
                {"type": "text", "text": "⏳ Continuous improvement planning"}
            ]
        }
    ]
}
```

This component provides essential step indicator functionality for ERP systems requiring multi-step process guidance, workflow visualization, and progress tracking with clear visual indicators and detailed status information.