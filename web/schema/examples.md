# 🚀 Schema Engine Implementation Examples

**Version:** 1.0.0  
**Status:** Production Ready  
**Last Updated:** 2025-10-29

This document provides comprehensive examples of how to implement the Schema Engine with design system integration.

---

## Table of Contents

1. [Basic Form Examples](#1-basic-form-examples)
2. [Design Token Integration](#2-design-token-integration)
3. [Advanced Field Types](#3-advanced-field-types)
4. [Enterprise Features](#4-enterprise-features)
5. [HTMX Integration](#5-htmx-integration)
6. [Alpine.js Enhancement](#6-alpinejs-enhancement)
7. [Multi-Tenant Scenarios](#7-multi-tenant-scenarios)
8. [Workflow Integration](#8-workflow-integration)
9. [Accessibility Examples](#9-accessibility-examples)
10. [Complete Applications](#10-complete-applications)
11. [End User UI Examples](#11-end-user-ui-examples)

---

## 1. 📝 Basic Form Examples

### Simple Contact Form

**JSON Schema:**
```json
{
  "id": "contact-form",
  "type": "form",
  "title": "Contact Us",
  "description": "Send us a message and we'll get back to you",
  "config": {
    "action": "/api/contact",
    "method": "POST"
  },
  "fields": [
    {
      "name": "name",
      "type": "text",
      "label": "Full Name",
      "placeholder": "Enter your full name",
      "required": true,
      "validation": {
        "required": true,
        "minLength": 2,
        "maxLength": 100
      }
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email Address",
      "placeholder": "user@example.com",
      "required": true,
      "validation": {
        "required": true,
        "email": true
      }
    },
    {
      "name": "message",
      "type": "textarea",
      "label": "Message",
      "placeholder": "Tell us what you need help with",
      "required": true,
      "validation": {
        "required": true,
        "minLength": 10,
        "maxLength": 1000
      }
    }
  ],
  "actions": [
    {
      "id": "submit",
      "type": "submit",
      "text": "Send Message",
      "variant": "primary"
    },
    {
      "id": "reset",
      "type": "reset",
      "text": "Clear Form",
      "variant": "outline"
    }
  ]
}
```

**Go Implementation:**
```go
// Using the FormSchema builder
func CreateContactForm() *FormSchema {
    return NewFormSchema("contact-form", "Contact Us").
        AddTextField("name", "Full Name", "Enter your full name", true).
        AddEmailField("email", "Email Address", true).
        AddTextareaField("message", "Message", true, IntPtr(10), IntPtr(1000)).
        AddSubmitAction("Send Message", "primary").
        AddResetAction("Clear Form").
        SetFormAction("/api/contact", "POST")
}

// Usage
form := CreateContactForm()
result, err := form.ValidateFormData(ctx, submittedData)
if err != nil {
    // Handle validation errors
}
html, err := form.Render(ctx, data)
```

---

## 2. 🎨 Design Token Integration

### Form with Custom Styling

**JSON Schema with Design Tokens:**
```json
{
  "id": "styled-form",
  "type": "form",
  "title": "Styled Registration Form",
  "theme": {
    "primary": "blue",
    "mode": "light"
  },
  "tokens": {
    "form": {
      "background": "background.subtle",
      "padding": "spacing.6",
      "borderRadius": "radius.lg"
    }
  },
  "fields": [
    {
      "name": "username",
      "type": "text",
      "label": "Username",
      "style": {
        "backgroundToken": "input.background",
        "colorToken": "input.text",
        "borderToken": "input.border",
        "focusBorderToken": "input.focus.border"
      },
      "layout": {
        "colSpan": 2,
        "class": "field-emphasized"
      }
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email",
      "style": {
        "backgroundToken": "input.background",
        "colorToken": "input.text"
      }
    }
  ],
  "actions": [
    {
      "id": "register",
      "type": "submit",
      "text": "Create Account",
      "style": {
        "backgroundToken": "button.primary.background",
        "colorToken": "button.primary.text",
        "hoverBackgroundToken": "button.primary.hover.background"
      }
    }
  ],
  "layout": {
    "type": "grid",
    "columns": 2,
    "gap": "1rem",
    "responsive": true
  }
}
```

**Generated CSS:**
```css
.schema-form[data-form="styled-form"] {
  background: var(--background-subtle);
  padding: var(--spacing-6);
  border-radius: var(--radius-lg);
}

.field-input[data-field="username"] {
  background: var(--input-background);
  color: var(--input-text);
  border: 1px solid var(--input-border);
  grid-column: span 2;
}

.field-input[data-field="username"]:focus {
  border-color: var(--input-focus-border);
}

.action-button[data-action="register"] {
  background: var(--button-primary-background);
  color: var(--button-primary-text);
}

.action-button[data-action="register"]:hover {
  background: var(--button-primary-hover-background);
}
```

---

## 3. 🔧 Advanced Field Types

### Complex Field Configuration

**JSON Schema:**
```json
{
  "id": "advanced-form",
  "type": "form",
  "title": "Advanced Product Form",
  "fields": [
    {
      "name": "category",
      "type": "select",
      "label": "Product Category",
      "required": true,
      "dataSource": {
        "type": "api",
        "url": "/api/categories",
        "method": "GET",
        "valueField": "id",
        "labelField": "name",
        "searchField": "name",
        "cache": true,
        "cacheTTL": 300
      },
      "conditional": {
        "showIf": "product_type === 'physical'"
      }
    },
    {
      "name": "price",
      "type": "currency",
      "label": "Price",
      "required": true,
      "validation": {
        "min": 0.01,
        "max": 999999.99
      },
      "mask": {
        "pattern": "$999,999.99",
        "type": "currency"
      },
      "transform": {
        "input": "value => parseFloat(value.replace(/[$,]/g, ''))",
        "output": "value => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value)"
      }
    },
    {
      "name": "availability",
      "type": "daterange",
      "label": "Available Dates",
      "validation": {
        "required": true,
        "custom": ["future_dates_only"]
      },
      "events": {
        "onChange": "handleAvailabilityChange",
        "debounce": 300
      }
    },
    {
      "name": "features",
      "type": "multiselect",
      "label": "Product Features",
      "options": [
        {
          "value": "wireless",
          "label": "Wireless Connectivity",
          "icon": "wifi",
          "group": "connectivity"
        },
        {
          "value": "waterproof",
          "label": "Waterproof",
          "icon": "shield",
          "group": "durability"
        }
      ],
      "layout": {
        "display": "chips",
        "searchable": true,
        "groupBy": "group"
      }
    },
    {
      "name": "description",
      "type": "richtext",
      "label": "Product Description",
      "validation": {
        "minLength": 50,
        "maxLength": 2000
      },
      "layout": {
        "colSpan": 2,
        "height": "200px"
      }
    }
  ],
  "layout": {
    "type": "grid",
    "columns": 2,
    "sections": [
      {
        "id": "basic-info",
        "title": "Basic Information",
        "fields": ["category", "price"],
        "icon": "info",
        "collapsible": false
      },
      {
        "id": "details",
        "title": "Product Details",
        "fields": ["availability", "features", "description"],
        "icon": "settings",
        "collapsible": true
      }
    ]
  }
}
```

---

## 4. 🏢 Enterprise Features

### Multi-Tenant Form with Security

**JSON Schema:**
```json
{
  "id": "customer-data-form",
  "type": "form",
  "title": "Customer Data Management",
  "tenant": {
    "enabled": true,
    "field": "tenant_id",
    "isolation": "strict",
    "inherit": false
  },
  "security": {
    "csrf": {
      "enabled": true,
      "tokenField": "_token"
    },
    "rateLimit": {
      "enabled": true,
      "maxRequests": 10,
      "windowSeconds": 60
    },
    "encryption": {
      "enabled": true,
      "encryptedFields": ["ssn", "credit_card"],
      "algorithm": "AES-256-GCM"
    }
  },
  "workflow": {
    "enabled": true,
    "id": "customer-approval-workflow",
    "approvals": {
      "required": true,
      "type": "sequential",
      "approvers": ["manager-uuid", "compliance-uuid"],
      "minVotes": 1
    },
    "notifications": [
      {
        "id": "approval-request",
        "event": "workflow.approval_required",
        "recipients": ["manager@company.com"],
        "channels": ["email", "slack"],
        "template": "approval-request-template"
      }
    ]
  },
  "fields": [
    {
      "name": "customer_id",
      "type": "text",
      "label": "Customer ID",
      "required": true,
      "permissions": {
        "view": {
          "operator": "AND",
          "conditions": [
            {"field": "user.role", "operator": "IN", "value": ["admin", "manager"]}
          ]
        },
        "edit": {
          "operator": "AND",
          "conditions": [
            {"field": "user.role", "operator": "EQUALS", "value": "admin"}
          ]
        }
      }
    },
    {
      "name": "ssn",
      "type": "text",
      "label": "Social Security Number",
      "mask": {
        "pattern": "XXX-XX-XXXX",
        "type": "custom"
      },
      "permissions": {
        "view": {
          "operator": "AND",
          "conditions": [
            {"field": "user.permissions", "operator": "CONTAINS", "value": "view_pii"}
          ]
        }
      },
      "validation": {
        "pattern": "^\\d{3}-\\d{2}-\\d{4}$",
        "async": {
          "url": "/api/validate/ssn",
          "method": "POST",
          "debounce": 500
        }
      }
    }
  ],
  "i18n": {
    "enabled": true,
    "defaultLocale": "en-US",
    "supportedLocales": ["en-US", "es-ES", "fr-FR"],
    "translations": {
      "customer_id.label": {
        "en-US": "Customer ID",
        "es-ES": "ID del Cliente",
        "fr-FR": "ID Client"
      }
    }
  }
}
```

---

## 5. ⚡ HTMX Integration

### Dynamic Form with Real-Time Updates

**JSON Schema:**
```json
{
  "id": "dynamic-search-form",
  "type": "form",
  "title": "Product Search",
  "htmx": {
    "enabled": true,
    "boost": true
  },
  "fields": [
    {
      "name": "search_query",
      "type": "search",
      "label": "Search Products",
      "placeholder": "Type to search...",
      "htmx": {
        "get": "/api/search/suggestions",
        "trigger": "keyup changed delay:300ms",
        "target": "#search-suggestions",
        "swap": "innerHTML",
        "include": "this",
        "headers": {
          "X-Search-Context": "products"
        }
      },
      "events": {
        "onChange": "handleSearchQuery",
        "debounce": 300
      }
    },
    {
      "name": "category_filter",
      "type": "select",
      "label": "Category",
      "htmx": {
        "post": "/api/search/filter",
        "trigger": "change",
        "target": "#search-results",
        "swap": "outerHTML",
        "include": "closest form"
      },
      "dataSource": {
        "type": "api",
        "url": "/api/categories",
        "dependsOn": ["search_query"]
      }
    }
  ],
  "actions": [
    {
      "id": "search",
      "type": "submit",
      "text": "Search",
      "htmx": {
        "post": "/api/search",
        "target": "#search-results",
        "swap": "innerHTML",
        "indicator": "#search-loading"
      }
    }
  ]
}
```

**Generated HTML:**
```html
<form id="dynamic-search-form" hx-boost="true">
  <div class="field-group">
    <label for="search_query">Search Products</label>
    <input 
      type="search" 
      id="search_query"
      name="search_query"
      placeholder="Type to search..."
      hx-get="/api/search/suggestions"
      hx-trigger="keyup changed delay:300ms"
      hx-target="#search-suggestions"
      hx-swap="innerHTML"
      hx-include="this"
      hx-headers='{"X-Search-Context": "products"}'
    >
    <div id="search-suggestions"></div>
  </div>
  
  <div class="field-group">
    <label for="category_filter">Category</label>
    <select 
      id="category_filter"
      name="category_filter"
      hx-post="/api/search/filter"
      hx-trigger="change"
      hx-target="#search-results"
      hx-swap="outerHTML"
      hx-include="closest form"
    >
      <!-- Options loaded dynamically -->
    </select>
  </div>
  
  <button 
    type="submit"
    hx-post="/api/search"
    hx-target="#search-results"
    hx-swap="innerHTML"
    hx-indicator="#search-loading"
  >
    Search
  </button>
</form>

<div id="search-loading" class="htmx-indicator">
  Searching...
</div>

<div id="search-results">
  <!-- Results loaded here -->
</div>
```

---

## 6. 🎮 Alpine.js Enhancement

### Interactive Form with Client-Side Logic

**JSON Schema:**
```json
{
  "id": "interactive-calculator",
  "type": "form",
  "title": "Loan Calculator",
  "alpine": {
    "enabled": true,
    "xData": "{ loan: { amount: 0, rate: 0, term: 0 }, payment: 0, calculatePayment() { this.payment = (this.loan.amount * (this.loan.rate/100/12)) / (1 - Math.pow(1 + (this.loan.rate/100/12), -this.loan.term)); } }"
  },
  "fields": [
    {
      "name": "loan_amount",
      "type": "currency",
      "label": "Loan Amount",
      "required": true,
      "alpine": {
        "xModel": "loan.amount",
        "xOn": {
          "input": "calculatePayment()"
        }
      },
      "validation": {
        "min": 1000,
        "max": 1000000
      }
    },
    {
      "name": "interest_rate",
      "type": "number",
      "label": "Interest Rate (%)",
      "required": true,
      "alpine": {
        "xModel": "loan.rate",
        "xOn": {
          "input": "calculatePayment()"
        }
      },
      "validation": {
        "min": 0.1,
        "max": 30
      }
    },
    {
      "name": "loan_term",
      "type": "number",
      "label": "Loan Term (months)",
      "required": true,
      "alpine": {
        "xModel": "loan.term",
        "xOn": {
          "input": "calculatePayment()"
        }
      },
      "validation": {
        "min": 6,
        "max": 360
      }
    },
    {
      "name": "monthly_payment",
      "type": "display",
      "label": "Monthly Payment",
      "alpine": {
        "xText": "new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(payment || 0)",
        "xShow": "payment > 0"
      },
      "style": {
        "fontSize": "1.5rem",
        "fontWeight": "bold",
        "colorToken": "color.primary"
      }
    }
  ],
  "events": {
    "onMount": "calculatePayment()"
  }
}
```

**Generated HTML:**
```html
<form 
  id="interactive-calculator"
  x-data="{
    loan: { amount: 0, rate: 0, term: 0 },
    payment: 0,
    calculatePayment() {
      this.payment = (this.loan.amount * (this.loan.rate/100/12)) / 
                     (1 - Math.pow(1 + (this.loan.rate/100/12), -this.loan.term));
    }
  }"
  x-init="calculatePayment()"
>
  <div class="field-group">
    <label for="loan_amount">Loan Amount</label>
    <input 
      type="text"
      id="loan_amount"
      name="loan_amount"
      x-model="loan.amount"
      x-on:input="calculatePayment()"
      placeholder="$0.00"
    >
  </div>
  
  <div class="field-group">
    <label for="interest_rate">Interest Rate (%)</label>
    <input 
      type="number"
      id="interest_rate"
      name="interest_rate"
      x-model="loan.rate"
      x-on:input="calculatePayment()"
      step="0.1"
      min="0.1"
      max="30"
    >
  </div>
  
  <div class="field-group">
    <label for="loan_term">Loan Term (months)</label>
    <input 
      type="number"
      id="loan_term"
      name="loan_term"
      x-model="loan.term"
      x-on:input="calculatePayment()"
      min="6"
      max="360"
    >
  </div>
  
  <div class="field-group" x-show="payment > 0">
    <label>Monthly Payment</label>
    <div 
      class="payment-display"
      x-text="new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(payment || 0)"
      style="font-size: 1.5rem; font-weight: bold; color: var(--color-primary);"
    ></div>
  </div>
</form>
```

---

## 7. 🔐 Multi-Tenant Scenarios

### Tenant-Aware Data Filtering

**Go Implementation:**
```go
func CreateTenantAwareForm(tenantID string, userRole string) *FormSchema {
    form := NewFormSchema("customer-management", "Customer Management")
    
    // Configure tenant isolation
    form.Tenant = &Tenant{
        Enabled:   true,
        Field:     "tenant_id",
        Isolation: "strict",
    }
    
    // Add fields based on tenant and role
    form.AddTextField("customer_name", "Customer Name", "", true)
    
    if userRole == "admin" {
        // Admin can see all customer data
        form.AddTextField("internal_notes", "Internal Notes", "", false)
        form.AddSelectField("status", "Status", true, []Option{
            {Value: "active", Label: "Active"},
            {Value: "suspended", Label: "Suspended"},
            {Value: "closed", Label: "Closed"},
        })
    }
    
    // Add conditional permissions
    for i := range form.Fields {
        field := &form.Fields[i]
        if field.Name == "internal_notes" {
            field.Permissions = &FieldPermissions{
                View: &condition.ConditionGroup{
                    Operator: "AND",
                    Conditions: []condition.Condition{
                        {
                            Field:    "user.role",
                            Operator: "EQUALS",
                            Value:    "admin",
                        },
                        {
                            Field:    "user.tenant_id",
                            Operator: "EQUALS",
                            Value:    tenantID,
                        },
                    },
                },
            }
        }
    }
    
    return form
}

// Usage with context
func HandleFormRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Extract tenant and user info from context
    tenantID := middleware.GetTenantID(ctx)
    userRole := middleware.GetUserRole(ctx)
    
    // Create tenant-specific form
    form := CreateTenantAwareForm(tenantID, userRole)
    
    // Set runtime context
    form.Context = &Context{
        TenantID:    tenantID,
        UserID:      middleware.GetUserID(ctx),
        Permissions: middleware.GetUserPermissions(ctx),
        Environment: "production",
    }
    
    // Validate and render
    if err := form.ValidateForm(ctx); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    html, err := form.Render(ctx, nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
}
```

---

## 8. 🔄 Workflow Integration

### Approval Workflow Form

**JSON Schema:**
```json
{
  "id": "expense-approval-form",
  "type": "form",
  "title": "Expense Report Submission",
  "workflow": {
    "enabled": true,
    "id": "expense-approval-workflow",
    "stage": "draft",
    "status": "pending",
    "actions": [
      {
        "id": "submit-for-approval",
        "label": "Submit for Approval",
        "type": "submit",
        "button": {
          "id": "submit-approval",
          "type": "submit",
          "text": "Submit for Approval",
          "variant": "primary"
        },
        "conditions": {
          "operator": "AND",
          "conditions": [
            {"field": "form.valid", "operator": "EQUALS", "value": true},
            {"field": "total_amount", "operator": "GREATER_THAN", "value": 0}
          ]
        },
        "confirm": {
          "title": "Submit Expense Report",
          "message": "Are you sure you want to submit this expense report for approval?",
          "confirmText": "Submit",
          "type": "info"
        }
      },
      {
        "id": "save-draft",
        "label": "Save as Draft",
        "type": "custom",
        "button": {
          "id": "save-draft",
          "type": "button",
          "text": "Save Draft",
          "variant": "outline"
        }
      }
    ],
    "transitions": [
      {
        "from": "draft",
        "to": "pending_approval",
        "action": "submit-for-approval",
        "onSuccess": "showSuccessMessage('Expense report submitted for approval')"
      }
    ],
    "approvals": {
      "required": true,
      "type": "sequential",
      "approvers": ["manager-uuid"],
      "escalation": {
        "after": 7200,
        "recipients": ["director-uuid"],
        "action": "notify-admin"
      }
    },
    "notifications": [
      {
        "id": "approval-request",
        "event": "workflow.submitted",
        "recipients": ["manager@company.com"],
        "channels": ["email"],
        "template": "expense-approval-request"
      },
      {
        "id": "approval-granted",
        "event": "workflow.approved",
        "recipients": ["{{ submitter.email }}"],
        "channels": ["email", "push"],
        "template": "expense-approved"
      }
    ]
  },
  "fields": [
    {
      "name": "expense_date",
      "type": "date",
      "label": "Expense Date",
      "required": true,
      "validation": {
        "required": true,
        "custom": ["not_future_date"]
      }
    },
    {
      "name": "category",
      "type": "select",
      "label": "Expense Category",
      "required": true,
      "options": [
        {"value": "travel", "label": "Travel"},
        {"value": "meals", "label": "Meals & Entertainment"},
        {"value": "supplies", "label": "Office Supplies"},
        {"value": "software", "label": "Software & Tools"}
      ]
    },
    {
      "name": "amount",
      "type": "currency",
      "label": "Amount",
      "required": true,
      "validation": {
        "required": true,
        "min": 0.01,
        "max": 5000
      },
      "events": {
        "onChange": "updateTotalAmount",
        "debounce": 300
      }
    },
    {
      "name": "description",
      "type": "textarea",
      "label": "Description",
      "placeholder": "Provide details about this expense",
      "required": true,
      "validation": {
        "required": true,
        "minLength": 10,
        "maxLength": 500
      }
    },
    {
      "name": "receipt",
      "type": "file",
      "label": "Receipt",
      "validation": {
        "fileTypes": ["image/*", "application/pdf"],
        "maxSize": 5242880
      },
      "conditional": {
        "showIf": "amount >= 25"
      }
    }
  ],
  "validation": {
    "crossFieldRules": [
      {
        "id": "receipt-required-for-large-expenses",
        "message": "Receipt is required for expenses over $25",
        "fields": ["amount", "receipt"],
        "conditions": {
          "operator": "OR",
          "conditions": [
            {"field": "amount", "operator": "LESS_THAN", "value": 25},
            {"field": "receipt", "operator": "NOT_EMPTY"}
          ]
        }
      }
    ]
  }
}
```

---

## 9. ♿ Accessibility Examples

### WCAG 2.1 Compliant Form

**Go Implementation:**
```go
func CreateAccessibleForm() *FormSchema {
    form := NewFormSchema("accessible-form", "Accessible Registration Form")
    
    // Enable accessibility features
    form.Config.Headers = map[string]string{
        "X-Accessibility-Mode": "enhanced",
    }
    
    // Add fields with accessibility attributes
    emailField := Field{
        Name:        "email",
        Type:        FieldEmail,
        Label:       "Email Address",
        Placeholder: "Enter your email address",
        Required:    true,
        Help:        "We'll use this to send you important updates",
        
        // Accessibility attributes
        Layout: &FieldLayout{
            Class: "field-required",
        },
        
        // ARIA attributes will be automatically added:
        // aria-required="true"
        // aria-describedby="email-help email-error"
        // aria-invalid="false" (updated on validation)
        
        Validation: &FieldValidation{
            Required: true,
            Email:    true,
            Messages: map[string]string{
                "required":      "Email address is required",
                "invalid_email": "Please enter a valid email address",
            },
        },
    }
    
    passwordField := Field{
        Name:        "password",
        Type:        FieldPassword,
        Label:       "Password",
        Required:    true,
        Help:        "Must be at least 8 characters with uppercase, lowercase, and numbers",
        
        Validation: &FieldValidation{
            Required:  true,
            MinLength: IntPtr(8),
            Pattern:   "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)[a-zA-Z\\d@$!%*?&]{8,}$",
            Messages: map[string]string{
                "required":   "Password is required",
                "min_length": "Password must be at least 8 characters",
                "pattern":    "Password must contain uppercase, lowercase, and numbers",
            },
        },
        
        // Progressive enhancement for password strength
        Events: &FieldEvents{
            OnInput:  "checkPasswordStrength",
            Debounce: 300,
        },
    }
    
    form.AddField(emailField)
    form.AddField(passwordField)
    
    // Add submit action with proper labeling
    submitAction := Action{
        ID:      "register",
        Type:    ActionSubmit,
        Text:    "Create Account",
        Variant: "primary",
        
        // Will generate:
        // <button type="submit" aria-describedby="register-help">
        //   Create Account
        // </button>
        // <div id="register-help" class="sr-only">
        //   Submits the registration form
        // </div>
    }
    
    form.AddAction(submitAction)
    
    return form
}
```

**Generated Accessible HTML:**
```html
<form 
  id="accessible-form" 
  role="form"
  aria-labelledby="form-title"
  novalidate
>
  <h1 id="form-title">Accessible Registration Form</h1>
  
  <div class="field-group field-required" role="group">
    <label for="email" class="field-label">
      Email Address
      <span class="required-indicator" aria-label="required">*</span>
    </label>
    
    <input 
      type="email"
      id="email"
      name="email"
      class="field-input"
      placeholder="Enter your email address"
      aria-required="true"
      aria-describedby="email-help email-error"
      aria-invalid="false"
      autocomplete="email"
    >
    
    <div id="email-help" class="field-help">
      We'll use this to send you important updates
    </div>
    
    <div id="email-error" class="field-error" aria-live="polite"></div>
  </div>
  
  <div class="field-group field-required" role="group">
    <label for="password" class="field-label">
      Password
      <span class="required-indicator" aria-label="required">*</span>
    </label>
    
    <input 
      type="password"
      id="password"
      name="password"
      class="field-input"
      aria-required="true"
      aria-describedby="password-help password-error password-strength"
      aria-invalid="false"
      autocomplete="new-password"
    >
    
    <div id="password-help" class="field-help">
      Must be at least 8 characters with uppercase, lowercase, and numbers
    </div>
    
    <div id="password-strength" class="password-strength" aria-live="polite">
      <!-- Updated dynamically -->
    </div>
    
    <div id="password-error" class="field-error" aria-live="polite"></div>
  </div>
  
  <button 
    type="submit"
    class="action-button action-button--primary"
    aria-describedby="register-help"
  >
    Create Account
  </button>
  
  <div id="register-help" class="sr-only">
    Submits the registration form
  </div>
</form>

<!-- Live region for form-level messages -->
<div id="form-messages" class="sr-only" aria-live="polite" aria-atomic="true"></div>
```

---

## 10. 🚀 Complete Applications

### Full ERP Customer Management Module

This example shows how multiple forms work together in a complete application:

**Customer List Schema:**
```json
{
  "id": "customer-list",
  "type": "list",
  "title": "Customer Management",
  "config": {
    "dataSource": "/api/customers",
    "pageSize": 25,
    "searchable": true,
    "sortable": true
  },
  "fields": [
    {
      "name": "name",
      "type": "text",
      "label": "Customer Name",
      "sortable": true,
      "searchable": true
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email",
      "sortable": true
    },
    {
      "name": "status",
      "type": "badge",
      "label": "Status",
      "options": [
        {"value": "active", "label": "Active", "color": "green"},
        {"value": "inactive", "label": "Inactive", "color": "gray"}
      ]
    },
    {
      "name": "created_at",
      "type": "date",
      "label": "Created",
      "sortable": true,
      "format": "MMM DD, YYYY"
    }
  ],
  "actions": [
    {
      "id": "create",
      "type": "button",
      "text": "Add Customer",
      "variant": "primary",
      "config": {
        "url": "/customers/new",
        "method": "GET"
      }
    }
  ],
  "rowActions": [
    {
      "id": "view",
      "type": "link",
      "text": "View",
      "icon": "eye",
      "config": {
        "url": "/customers/{{id}}"
      }
    },
    {
      "id": "edit",
      "type": "link", 
      "text": "Edit",
      "icon": "edit",
      "config": {
        "url": "/customers/{{id}}/edit"
      }
    }
  ]
}
```

**Customer Create/Edit Form:**
```json
{
  "id": "customer-form",
  "type": "form",
  "title": "{{#if editing}}Edit Customer{{else}}Create Customer{{/if}}",
  "config": {
    "action": "{{#if editing}}/api/customers/{{id}}{{else}}/api/customers{{/if}}",
    "method": "{{#if editing}}PUT{{else}}POST{{/if}}"
  },
  "layout": {
    "type": "sections",
    "sections": [
      {
        "id": "basic-info",
        "title": "Basic Information",
        "fields": ["company_name", "contact_name", "email", "phone"],
        "icon": "user"
      },
      {
        "id": "address",
        "title": "Address Information", 
        "fields": ["street", "city", "state", "zip", "country"],
        "icon": "map-pin"
      },
      {
        "id": "business",
        "title": "Business Details",
        "fields": ["industry", "company_size", "annual_revenue"],
        "icon": "briefcase"
      }
    ]
  },
  "fields": [
    {
      "name": "company_name",
      "type": "text",
      "label": "Company Name",
      "required": true,
      "validation": {
        "required": true,
        "minLength": 2,
        "maxLength": 100
      }
    },
    {
      "name": "contact_name",
      "type": "text",
      "label": "Primary Contact",
      "required": true
    },
    {
      "name": "email",
      "type": "email",
      "label": "Email Address",
      "required": true,
      "validation": {
        "required": true,
        "email": true,
        "async": {
          "url": "/api/validate/email",
          "method": "POST",
          "debounce": 500
        }
      }
    },
    {
      "name": "phone",
      "type": "phone",
      "label": "Phone Number",
      "mask": {
        "pattern": "(999) 999-9999",
        "type": "phone"
      }
    },
    {
      "name": "industry",
      "type": "select",
      "label": "Industry",
      "dataSource": {
        "type": "api",
        "url": "/api/industries",
        "cache": true,
        "cacheTTL": 3600
      }
    },
    {
      "name": "company_size",
      "type": "select",
      "label": "Company Size",
      "options": [
        {"value": "1-10", "label": "1-10 employees"},
        {"value": "11-50", "label": "11-50 employees"},
        {"value": "51-200", "label": "51-200 employees"},
        {"value": "201-500", "label": "201-500 employees"},
        {"value": "500+", "label": "500+ employees"}
      ]
    }
  ],
  "actions": [
    {
      "id": "save",
      "type": "submit",
      "text": "{{#if editing}}Update Customer{{else}}Create Customer{{/if}}",
      "variant": "primary"
    },
    {
      "id": "cancel",
      "type": "button",
      "text": "Cancel",
      "variant": "outline",
      "config": {
        "url": "/customers",
        "method": "GET"
      }
    }
  ],
  "validation": {
    "mode": "onChange",
    "showErrors": "touched"
  }
}
```

**Go Application Integration:**
```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/gofiber/fiber/v2"
    "github.com/your-org/erp/web/schema"
)

func main() {
    app := fiber.New()
    
    // Initialize schema system
    renderer := NewTemplRenderer()
    validator := schema.NewValidator()
    
    // Customer list page
    app.Get("/customers", func(c *fiber.Ctx) error {
        listSchema, err := LoadSchemaFromFile("schemas/customer-list.json")
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        
        // Load customer data
        customers, err := customerService.GetAll(c.Context())
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        
        html, err := renderer.RenderList(c.Context(), listSchema, customers)
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        
        return c.Type("html").SendString(html)
    })
    
    // Customer create form
    app.Get("/customers/new", func(c *fiber.Ctx) error {
        formSchema := schema.NewFormSchema("customer-form", "Create Customer").
            AddTextField("company_name", "Company Name", "", true).
            AddTextField("contact_name", "Primary Contact", "", true).
            AddEmailField("email", "Email Address", true).
            AddTextField("phone", "Phone Number", "", false).
            AddSubmitAction("Create Customer", "primary").
            SetFormAction("/api/customers", "POST")
        
        html, err := formSchema.Render(c.Context(), nil)
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        
        return c.Type("html").SendString(html)
    })
    
    // Customer edit form
    app.Get("/customers/:id/edit", func(c *fiber.Ctx) error {
        customerID := c.Params("id")
        
        customer, err := customerService.GetByID(c.Context(), customerID)
        if err != nil {
            return c.Status(404).SendString("Customer not found")
        }
        
        formSchema := CreateCustomerEditForm(customerID)
        
        // Pre-populate with existing data
        data := map[string]interface{}{
            "company_name": customer.CompanyName,
            "contact_name": customer.ContactName,
            "email":        customer.Email,
            "phone":        customer.Phone,
        }
        
        html, err := formSchema.Render(c.Context(), data)
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        
        return c.Type("html").SendString(html)
    })
    
    // API endpoints for form submissions
    app.Post("/api/customers", func(c *fiber.Ctx) error {
        var formData map[string]interface{}
        if err := c.BodyParser(&formData); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid form data"})
        }
        
        // Validate using schema
        formSchema := CreateCustomerForm()
        result, err := formSchema.ValidateFormData(c.Context(), formData)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }
        
        if !result.Valid {
            return c.Status(400).JSON(fiber.Map{
                "errors": result.FieldErrors,
            })
        }
        
        // Create customer
        customer, err := customerService.Create(c.Context(), formData)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }
        
        return c.JSON(customer)
    })
    
    app.Listen(":3000")
}

func CreateCustomerForm() *schema.FormSchema {
    return schema.NewFormSchema("customer-form", "Customer Information").
        AddTextField("company_name", "Company Name", "", true).
        AddTextField("contact_name", "Primary Contact", "", true).
        AddEmailField("email", "Email Address", true).
        AddTextField("phone", "Phone Number", "", false).
        SetFormAction("/api/customers", "POST").
        SetValidationMode(schema.ValidationOnChange)
}

func CreateCustomerEditForm(customerID string) *schema.FormSchema {
    form := CreateCustomerForm()
    form.Config.Action = "/api/customers/" + customerID
    form.Config.Method = "PUT"
    return form
}
```

---

## 📚 **Best Practices Summary**

### 1. **Schema Design**
- Keep schemas focused and single-purpose
- Use descriptive IDs and names
- Include helpful descriptions and examples
- Leverage validation for data integrity

### 2. **Design Token Integration**
- Always use tokens over hardcoded values
- Create semantic tokens for component-specific styling
- Ensure contrast ratios meet accessibility standards
- Test token combinations across themes

### 3. **Accessibility**
- Use semantic HTML structure
- Include proper ARIA attributes
- Provide clear error messages
- Test with screen readers

### 4. **Performance**
- Cache schema definitions where possible
- Use lazy loading for dynamic data sources
- Minimize HTMX requests with proper debouncing
- Optimize Alpine.js expressions

### 5. **Security**
- Enable CSRF protection for all forms
- Implement proper field-level permissions
- Encrypt sensitive field data
- Validate all inputs on both client and server

### 6. **Enterprise Features**
- Configure multi-tenancy from the start
- Implement comprehensive audit logging
- Use workflow approvals for sensitive operations
- Support internationalization early

---

## 11. 👤 End User UI Examples

This section shows what the actual UI looks like to end users - the beautiful, accessible interfaces that our Schema Engine generates.

### Customer Registration Form

**What the user sees:**

```json
{
  "userInterface": {
    "type": "form",
    "appearance": {
      "title": "Create Your Account",
      "subtitle": "Join thousands of satisfied customers",
      "layout": "centered_card",
      "theme": "modern_blue",
      "responsive": true
    },
    "visualElements": {
      "header": {
        "logo": "/assets/logo.svg",
        "background": "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
        "textColor": "white"
      },
      "form": {
        "width": "480px",
        "maxWidth": "90vw",
        "padding": "2rem",
        "borderRadius": "12px",
        "boxShadow": "0 20px 40px rgba(0,0,0,0.1)",
        "background": "white"
      }
    },
    "fields": [
      {
        "section": "Personal Information",
        "sectionIcon": "👤",
        "fields": [
          {
            "type": "text",
            "label": "Full Name",
            "placeholder": "Enter your full name",
            "icon": "user",
            "appearance": {
              "width": "100%",
              "height": "48px",
              "borderRadius": "8px",
              "border": "2px solid #e2e8f0",
              "focusBorder": "2px solid #3b82f6",
              "fontSize": "16px",
              "padding": "12px 16px"
            },
            "states": {
              "default": {
                "backgroundColor": "#ffffff",
                "borderColor": "#e2e8f0"
              },
              "focus": {
                "borderColor": "#3b82f6",
                "boxShadow": "0 0 0 3px rgba(59, 130, 246, 0.1)"
              },
              "error": {
                "borderColor": "#ef4444",
                "backgroundColor": "#fef2f2"
              },
              "success": {
                "borderColor": "#10b981",
                "backgroundColor": "#f0fdf4"
              }
            },
            "validation": {
              "required": true,
              "errorMessage": "Please enter your full name",
              "liveValidation": true
            }
          },
          {
            "type": "email",
            "label": "Email Address", 
            "placeholder": "user@company.com",
            "icon": "mail",
            "appearance": {
              "width": "100%",
              "height": "48px",
              "autocomplete": "email"
            },
            "validation": {
              "required": true,
              "emailFormat": true,
              "asyncCheck": {
                "endpoint": "/api/check-email",
                "message": "Checking availability...",
                "debounce": 500
              }
            }
          }
        ]
      },
      {
        "section": "Account Security",
        "sectionIcon": "🔒",
        "fields": [
          {
            "type": "password",
            "label": "Password",
            "placeholder": "Create a strong password",
            "icon": "lock",
            "features": {
              "showHideToggle": true,
              "strengthMeter": {
                "enabled": true,
                "position": "below",
                "levels": ["Weak", "Fair", "Good", "Strong", "Very Strong"],
                "colors": ["#ef4444", "#f97316", "#eab308", "#10b981", "#059669"]
              }
            },
            "validation": {
              "required": true,
              "minLength": 8,
              "requirements": [
                "At least 8 characters",
                "One uppercase letter",
                "One lowercase letter", 
                "One number",
                "One special character"
              ]
            }
          },
          {
            "type": "password",
            "label": "Confirm Password",
            "placeholder": "Confirm your password",
            "icon": "lock",
            "validation": {
              "required": true,
              "mustMatch": "password",
              "errorMessage": "Passwords must match"
            }
          }
        ]
      }
    ],
    "actions": [
      {
        "type": "submit",
        "text": "Create Account",
        "appearance": {
          "variant": "primary",
          "width": "100%",
          "height": "48px",
          "borderRadius": "8px",
          "fontSize": "16px",
          "fontWeight": "600",
          "backgroundColor": "#3b82f6",
          "color": "white",
          "hoverBackgroundColor": "#2563eb",
          "disabledOpacity": 0.6
        },
        "states": {
          "loading": {
            "text": "Creating Account...",
            "spinner": true,
            "disabled": true
          },
          "success": {
            "text": "Account Created! ✓",
            "backgroundColor": "#10b981"
          }
        }
      },
      {
        "type": "link",
        "text": "Already have an account? Sign in",
        "appearance": {
          "variant": "text",
          "color": "#3b82f6",
          "textAlign": "center",
          "marginTop": "1rem"
        },
        "action": "/login"
      }
    ],
    "footer": {
      "text": "By creating an account, you agree to our Terms of Service and Privacy Policy",
      "appearance": {
        "fontSize": "14px",
        "color": "#6b7280",
        "textAlign": "center",
        "marginTop": "1.5rem"
      }
    }
  }
}
```

### E-commerce Product Search

**What the user sees:**

```json
{
  "userInterface": {
    "type": "search_interface",
    "appearance": {
      "layout": "sidebar_main",
      "responsive": true,
      "theme": "modern_commerce"
    },
    "components": {
      "header": {
        "searchBar": {
          "placeholder": "Search for products...",
          "width": "600px",
          "height": "52px",
          "borderRadius": "26px",
          "features": {
            "autocomplete": true,
            "voiceSearch": true,
            "recentSearches": true,
            "suggestions": {
              "maxItems": 8,
              "categories": ["Products", "Brands", "Categories"],
              "highlightMatches": true
            }
          },
          "appearance": {
            "border": "2px solid #e5e7eb",
            "focusBorder": "2px solid #3b82f6",
            "boxShadow": "0 4px 6px rgba(0,0,0,0.05)",
            "fontSize": "16px",
            "padding": "0 24px"
          }
        }
      },
      "sidebar": {
        "title": "Filters",
        "width": "280px",
        "filters": [
          {
            "type": "category",
            "title": "Category",
            "appearance": {
              "collapsible": true,
              "expanded": true,
              "maxHeight": "200px",
              "scrollable": true
            },
            "options": [
              {
                "label": "Electronics",
                "count": 1247,
                "selected": false,
                "children": [
                  {"label": "Smartphones", "count": 234},
                  {"label": "Laptops", "count": 156},
                  {"label": "Headphones", "count": 89}
                ]
              },
              {
                "label": "Clothing",
                "count": 2341,
                "selected": false
              }
            ]
          },
          {
            "type": "price_range",
            "title": "Price Range",
            "appearance": {
              "sliderType": "dual_handle",
              "showValues": true,
              "currency": "USD"
            },
            "values": {
              "min": 0,
              "max": 2000,
              "currentMin": 50,
              "currentMax": 500,
              "step": 10
            },
            "display": {
              "format": "${value}",
              "inputs": true,
              "histogram": false
            }
          },
          {
            "type": "brand",
            "title": "Brand",
            "appearance": {
              "searchable": true,
              "showCount": true,
              "maxVisible": 8,
              "showMore": true
            },
            "options": [
              {"label": "Apple", "count": 145, "selected": false},
              {"label": "Samsung", "count": 132, "selected": true},
              {"label": "Sony", "count": 98, "selected": false}
            ]
          },
          {
            "type": "rating",
            "title": "Customer Rating",
            "appearance": {
              "style": "stars_and_up",
              "showCount": true
            },
            "options": [
              {"stars": 4, "label": "4 stars & up", "count": 456},
              {"stars": 3, "label": "3 stars & up", "count": 789}
            ]
          }
        ]
      },
      "mainContent": {
        "toolbar": {
          "resultsCount": {
            "text": "Showing 1-24 of 1,247 results",
            "appearance": {
              "fontSize": "14px",
              "color": "#6b7280"
            }
          },
          "sortOptions": {
            "label": "Sort by:",
            "selected": "relevance",
            "options": [
              {"value": "relevance", "label": "Best Match"},
              {"value": "price_low", "label": "Price: Low to High"},
              {"value": "price_high", "label": "Price: High to Low"},
              {"value": "rating", "label": "Customer Rating"},
              {"value": "newest", "label": "Newest First"}
            ]
          },
          "viewToggle": {
            "selected": "grid",
            "options": [
              {"value": "grid", "icon": "grid", "label": "Grid View"},
              {"value": "list", "icon": "list", "label": "List View"}
            ]
          }
        },
        "productGrid": {
          "layout": "grid",
          "columns": {
            "desktop": 4,
            "tablet": 3,
            "mobile": 2
          },
          "gap": "1.5rem",
          "productCard": {
            "appearance": {
              "borderRadius": "12px",
              "border": "1px solid #f3f4f6",
              "padding": "1rem",
              "hoverShadow": "0 8px 25px rgba(0,0,0,0.15)",
              "transition": "all 0.2s ease"
            },
            "elements": {
              "image": {
                "aspectRatio": "1:1",
                "borderRadius": "8px",
                "objectFit": "cover",
                "lazyLoad": true
              },
              "title": {
                "fontSize": "16px",
                "fontWeight": "600",
                "maxLines": 2,
                "color": "#1f2937"
              },
              "rating": {
                "style": "stars",
                "showCount": true,
                "size": "small"
              },
              "price": {
                "fontSize": "18px",
                "fontWeight": "700",
                "color": "#059669",
                "originalPrice": {
                  "show": true,
                  "strikethrough": true,
                  "color": "#9ca3af"
                }
              },
              "badge": {
                "position": "top_left",
                "types": ["sale", "new", "bestseller"],
                "colors": {
                  "sale": "#ef4444",
                  "new": "#3b82f6",
                  "bestseller": "#f59e0b"
                }
              }
            },
            "actions": {
              "quickView": {
                "enabled": true,
                "trigger": "hover",
                "position": "overlay"
              },
              "addToCart": {
                "text": "Add to Cart",
                "appearance": {
                  "variant": "primary",
                  "fullWidth": true
                }
              },
              "wishlist": {
                "icon": "heart",
                "position": "top_right"
              }
            }
          }
        },
        "pagination": {
          "type": "infinite_scroll",
          "fallback": "traditional",
          "appearance": {
            "loadMoreButton": {
              "text": "Load More Products",
              "variant": "outline"
            }
          }
        }
      }
    },
    "mobileExperience": {
      "filterDrawer": {
        "trigger": "Filter & Sort",
        "position": "bottom",
        "height": "80vh",
        "backdrop": true
      },
      "searchOverlay": {
        "fullScreen": true,
        "recentSearches": true,
        "popularSearches": true
      }
    }
  }
}
```

### Dashboard Analytics Interface

**What the user sees:**

```json
{
  "userInterface": {
    "type": "dashboard",
    "appearance": {
      "layout": "sidebar_main",
      "theme": "professional_dark",
      "responsive": true
    },
    "components": {
      "sidebar": {
        "width": "280px",
        "appearance": {
          "backgroundColor": "#1f2937",
          "borderRight": "1px solid #374151"
        },
        "navigation": [
          {
            "label": "Overview",
            "icon": "dashboard",
            "active": true,
            "badge": null
          },
          {
            "label": "Analytics",
            "icon": "chart-bar",
            "active": false,
            "badge": "new"
          },
          {
            "label": "Customers",
            "icon": "users",
            "active": false,
            "badge": "247"
          }
        ]
      },
      "header": {
        "height": "64px",
        "appearance": {
          "backgroundColor": "white",
          "borderBottom": "1px solid #e5e7eb",
          "padding": "0 2rem"
        },
        "elements": {
          "title": "Analytics Dashboard",
          "breadcrumbs": ["Home", "Analytics", "Overview"],
          "dateRange": {
            "type": "picker",
            "selected": "Last 30 days",
            "options": ["Today", "Yesterday", "Last 7 days", "Last 30 days", "Custom"]
          },
          "userMenu": {
            "avatar": "/avatars/john-doe.jpg",
            "name": "John Doe",
            "role": "Admin"
          }
        }
      },
      "mainContent": {
        "layout": "grid",
        "gap": "1.5rem",
        "padding": "2rem",
        "widgets": [
          {
            "type": "metric_card",
            "title": "Total Revenue",
            "value": "$142,850",
            "change": "+12.5%",
            "changeType": "positive",
            "icon": "dollar-sign",
            "appearance": {
              "size": "large",
              "backgroundColor": "white",
              "borderRadius": "12px",
              "padding": "1.5rem",
              "boxShadow": "0 4px 6px rgba(0,0,0,0.05)"
            },
            "chart": {
              "type": "sparkline",
              "data": [120, 135, 142, 138, 155, 162, 143],
              "color": "#10b981"
            }
          },
          {
            "type": "metric_card",
            "title": "New Customers",
            "value": "2,847",
            "change": "+8.2%",
            "changeType": "positive",
            "icon": "users"
          },
          {
            "type": "metric_card",
            "title": "Orders",
            "value": "1,429",
            "change": "-2.1%",
            "changeType": "negative",
            "icon": "shopping-cart"
          },
          {
            "type": "metric_card",
            "title": "Conversion Rate",
            "value": "3.24%",
            "change": "+0.8%",
            "changeType": "positive",
            "icon": "trending-up"
          },
          {
            "type": "chart",
            "title": "Revenue Trend",
            "subtitle": "Monthly revenue for the past 12 months",
            "size": "large",
            "chart": {
              "type": "line",
              "height": "300px",
              "data": {
                "labels": ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"],
                "datasets": [
                  {
                    "label": "Revenue",
                    "data": [65000, 59000, 80000, 81000, 56000, 85000, 98000, 89000, 76000, 94000, 108000, 142850],
                    "borderColor": "#3b82f6",
                    "backgroundColor": "rgba(59, 130, 246, 0.1)"
                  }
                ]
              },
              "options": {
                "responsive": true,
                "maintainAspectRatio": false,
                "plugins": {
                  "legend": {"display": false},
                  "tooltip": {
                    "mode": "index",
                    "intersect": false
                  }
                }
              }
            }
          },
          {
            "type": "chart",
            "title": "Traffic Sources",
            "subtitle": "Website traffic by source",
            "size": "medium",
            "chart": {
              "type": "doughnut",
              "height": "250px",
              "data": {
                "labels": ["Organic Search", "Direct", "Social Media", "Email", "Referral"],
                "datasets": [{
                  "data": [45, 25, 15, 10, 5],
                  "backgroundColor": ["#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6"]
                }]
              },
              "centerText": {
                "title": "Total Visits",
                "value": "45,692"
              }
            }
          },
          {
            "type": "table",
            "title": "Top Products",
            "subtitle": "Best selling products this month",
            "size": "medium",
            "table": {
              "headers": ["Product", "Sales", "Revenue", "Change"],
              "rows": [
                {
                  "product": {
                    "name": "Wireless Headphones",
                    "image": "/products/headphones.jpg",
                    "sku": "WH-001"
                  },
                  "sales": "1,247",
                  "revenue": "$31,175",
                  "change": {
                    "value": "+15.3%",
                    "type": "positive"
                  }
                },
                {
                  "product": {
                    "name": "Smart Watch",
                    "image": "/products/watch.jpg",
                    "sku": "SW-002"
                  },
                  "sales": "892",
                  "revenue": "$26,760",
                  "change": {
                    "value": "+8.7%",
                    "type": "positive"
                  }
                }
              ],
              "pagination": {
                "enabled": true,
                "pageSize": 10
              }
            }
          },
          {
            "type": "activity_feed",
            "title": "Recent Activity",
            "subtitle": "Latest system activities",
            "size": "medium",
            "activities": [
              {
                "type": "order",
                "title": "New order received",
                "description": "Order #12847 from John Smith",
                "timestamp": "2 minutes ago",
                "icon": "shopping-bag",
                "iconColor": "#10b981"
              },
              {
                "type": "user",
                "title": "New customer registered",
                "description": "Sarah Johnson joined as Premium member",
                "timestamp": "15 minutes ago",
                "icon": "user-plus",
                "iconColor": "#3b82f6"
              },
              {
                "type": "payment",
                "title": "Payment processed",
                "description": "$2,450 payment confirmed",
                "timestamp": "1 hour ago",
                "icon": "credit-card",
                "iconColor": "#059669"
              }
            ],
            "showMore": {
              "enabled": true,
              "text": "View all activities"
            }
          }
        ]
      }
    },
    "mobileExperience": {
      "sidebar": {
        "collapsed": true,
        "overlay": true
      },
      "widgets": {
        "stackedLayout": true,
        "scrollable": true
      },
      "charts": {
        "responsiveHeight": true,
        "touchOptimized": true
      }
    },
    "interactions": {
      "realTimeUpdates": {
        "enabled": true,
        "interval": 30000,
        "indicator": "pulse"
      },
      "filtering": {
        "dateRange": true,
        "quickFilters": ["Today", "This Week", "This Month"],
        "globalSearch": true
      },
      "export": {
        "formats": ["PDF", "Excel", "CSV"],
        "customizable": true
      }
    }
  }
}
```

### Invoice Creation Form

**What the user sees:**

```json
{
  "userInterface": {
    "type": "invoice_form",
    "appearance": {
      "layout": "document_editor",
      "theme": "business_professional",
      "responsive": true
    },
    "components": {
      "header": {
        "title": "Create New Invoice",
        "actions": [
          {
            "text": "Save Draft",
            "variant": "outline",
            "icon": "save"
          },
          {
            "text": "Preview",
            "variant": "secondary",
            "icon": "eye"
          },
          {
            "text": "Send Invoice",
            "variant": "primary",
            "icon": "send"
          }
        ]
      },
      "formSections": [
        {
          "title": "Invoice Details",
          "appearance": {
            "layout": "two_column",
            "backgroundColor": "#f9fafb",
            "padding": "1.5rem",
            "borderRadius": "8px"
          },
          "fields": [
            {
              "type": "auto_number",
              "label": "Invoice Number",
              "value": "INV-2025-001234",
              "appearance": {
                "readonly": true,
                "backgroundColor": "#f3f4f6",
                "prefix": "INV-",
                "autoGenerate": true
              }
            },
            {
              "type": "date",
              "label": "Invoice Date",
              "value": "2025-10-29",
              "appearance": {
                "calendar": true,
                "format": "MM/DD/YYYY"
              }
            },
            {
              "type": "date",
              "label": "Due Date",
              "placeholder": "Select due date",
              "appearance": {
                "calendar": true,
                "minDate": "today",
                "suggestedDates": [
                  {"label": "Net 15", "days": 15},
                  {"label": "Net 30", "days": 30},
                  {"label": "Net 45", "days": 45}
                ]
              }
            },
            {
              "type": "select",
              "label": "Payment Terms",
              "options": [
                {"value": "net_15", "label": "Net 15 - Payment due in 15 days"},
                {"value": "net_30", "label": "Net 30 - Payment due in 30 days"},
                {"value": "due_on_receipt", "label": "Due on Receipt"}
              ]
            }
          ]
        },
        {
          "title": "Customer Information",
          "appearance": {
            "collapsible": false,
            "border": "1px solid #e5e7eb",
            "borderRadius": "8px",
            "padding": "1.5rem"
          },
          "fields": [
            {
              "type": "customer_search",
              "label": "Select Customer",
              "appearance": {
                "searchable": true,
                "placeholder": "Search customers...",
                "createNew": {
                  "enabled": true,
                  "text": "Add New Customer"
                }
              },
              "features": {
                "autocomplete": true,
                "recentCustomers": true,
                "customerPreview": {
                  "showAddress": true,
                  "showBalance": true,
                  "showLastInvoice": true
                }
              }
            },
            {
              "type": "address_display",
              "label": "Billing Address",
              "appearance": {
                "editable": true,
                "multiLine": true,
                "copyButton": true
              },
              "value": {
                "company": "Acme Corporation",
                "street": "123 Business Ave",
                "city": "New York",
                "state": "NY",
                "zip": "10001",
                "country": "United States"
              }
            }
          ]
        },
        {
          "title": "Line Items",
          "appearance": {
            "fullWidth": true,
            "tableStyle": true
          },
          "component": {
            "type": "line_items_table",
            "appearance": {
              "striped": true,
              "sortable": false,
              "addButton": {
                "text": "+ Add Line Item",
                "position": "top_right"
              }
            },
            "columns": [
              {
                "field": "description",
                "title": "Description",
                "width": "40%",
                "type": "textarea",
                "appearance": {
                  "autoExpand": true,
                  "placeholder": "Enter item description..."
                }
              },
              {
                "field": "quantity",
                "title": "Qty",
                "width": "10%",
                "type": "number",
                "appearance": {
                  "min": 0,
                  "step": 1,
                  "textAlign": "center"
                }
              },
              {
                "field": "rate",
                "title": "Rate",
                "width": "15%",
                "type": "currency",
                "appearance": {
                  "currency": "USD",
                  "textAlign": "right"
                }
              },
              {
                "field": "amount",
                "title": "Amount",
                "width": "15%",
                "type": "currency_display",
                "appearance": {
                  "readonly": true,
                  "calculated": true,
                  "formula": "quantity * rate",
                  "textAlign": "right",
                  "fontWeight": "bold"
                }
              },
              {
                "field": "actions",
                "title": "",
                "width": "10%",
                "type": "row_actions",
                "actions": [
                  {"icon": "copy", "tooltip": "Duplicate"},
                  {"icon": "trash", "tooltip": "Delete", "color": "red"}
                ]
              }
            ],
            "footer": {
              "summary": {
                "subtotal": {
                  "label": "Subtotal",
                  "value": "$2,480.00",
                  "appearance": {"fontWeight": "normal"}
                },
                "tax": {
                  "label": "Tax (8.5%)",
                  "value": "$210.80",
                  "editable": true
                },
                "discount": {
                  "label": "Discount",
                  "value": "-$50.00",
                  "optional": true,
                  "types": ["percentage", "fixed_amount"]
                },
                "total": {
                  "label": "Total",
                  "value": "$2,640.80",
                  "appearance": {
                    "fontSize": "1.25rem",
                    "fontWeight": "bold",
                    "color": "#059669"
                  }
                }
              }
            }
          }
        },
        {
          "title": "Additional Information",
          "appearance": {
            "collapsible": true,
            "collapsed": false
          },
          "fields": [
            {
              "type": "textarea",
              "label": "Notes",
              "placeholder": "Add any additional notes for the customer...",
              "appearance": {
                "height": "100px",
                "maxLength": 500,
                "characterCounter": true
              }
            },
            {
              "type": "textarea",
              "label": "Terms & Conditions",
              "placeholder": "Enter payment terms and conditions...",
              "appearance": {
                "height": "80px",
                "templates": [
                  "Standard payment terms",
                  "Extended payment terms",
                  "Custom terms"
                ]
              }
            },
            {
              "type": "file_upload",
              "label": "Attachments",
              "appearance": {
                "multiple": true,
                "dragDrop": true,
                "allowedTypes": ["pdf", "doc", "docx", "jpg", "png"],
                "maxSize": "10MB",
                "preview": true
              }
            }
          ]
        }
      ],
      "sidebar": {
        "width": "300px",
        "appearance": {
          "backgroundColor": "#f9fafb",
          "padding": "1.5rem"
        },
        "widgets": [
          {
            "type": "invoice_preview",
            "title": "Preview",
            "appearance": {
              "miniature": true,
              "clickToExpand": true,
              "autoUpdate": true
            }
          },
          {
            "type": "payment_options",
            "title": "Payment Methods",
            "options": [
              {
                "type": "credit_card",
                "label": "Credit Card",
                "enabled": true,
                "fees": "2.9% + $0.30"
              },
              {
                "type": "bank_transfer",
                "label": "Bank Transfer",
                "enabled": true,
                "fees": "Free"
              },
              {
                "type": "paypal",
                "label": "PayPal",
                "enabled": false
              }
            ]
          },
          {
            "type": "delivery_options",
            "title": "Delivery Options",
            "options": [
              {
                "type": "email",
                "label": "Send via Email",
                "selected": true,
                "schedule": {
                  "enabled": true,
                  "options": ["Now", "Later", "Recurring"]
                }
              },
              {
                "type": "print",
                "label": "Print & Mail",
                "selected": false,
                "cost": "$1.50"
              }
            ]
          }
        ]
      },
      "bottomActions": {
        "alignment": "space_between",
        "leftActions": [
          {
            "text": "Cancel",
            "variant": "ghost"
          },
          {
            "text": "Save as Template",
            "variant": "outline",
            "icon": "bookmark"
          }
        ],
        "rightActions": [
          {
            "text": "Save Draft",
            "variant": "secondary"
          },
          {
            "text": "Send Invoice",
            "variant": "primary",
            "icon": "send",
            "dropdown": [
              "Send Now",
              "Schedule Send",
              "Save & Send Later"
            ]
          }
        ]
      }
    },
    "features": {
      "autoSave": {
        "enabled": true,
        "interval": 30000,
        "indicator": "Draft saved at 2:45 PM"
      },
      "validation": {
        "realTime": true,
        "blockSubmit": true,
        "highlightErrors": true
      },
      "calculations": {
        "autoUpdate": true,
        "precision": 2,
        "rounding": "half_up"
      },
      "templates": {
        "saveAsTemplate": true,
        "loadFromTemplate": true,
        "recentTemplates": 5
      }
    },
    "mobileExperience": {
      "layout": "stacked",
      "sidebar": "bottom_sheet",
      "lineItems": {
        "cardView": true,
        "swipeActions": true
      },
      "navigation": {
        "tabBar": true,
        "sections": ["Details", "Customer", "Items", "Preview"]
      }
    }
  }
}
```

### Customer Support Ticket Interface

**What the user sees:**

```json
{
  "userInterface": {
    "type": "ticket_interface",
    "appearance": {
      "layout": "three_panel",
      "theme": "support_professional",
      "responsive": true
    },
    "components": {
      "leftPanel": {
        "width": "320px",
        "title": "Tickets",
        "appearance": {
          "backgroundColor": "#f8fafc",
          "borderRight": "1px solid #e2e8f0"
        },
        "components": [
          {
            "type": "search_filter",
            "appearance": {
              "searchPlaceholder": "Search tickets...",
              "quickFilters": [
                {"label": "Open", "count": 23, "active": true},
                {"label": "Pending", "count": 8, "active": false},
                {"label": "Resolved", "count": 156, "active": false}
              ]
            }
          },
          {
            "type": "ticket_list",
            "appearance": {
              "itemHeight": "80px",
              "virtualScroll": true,
              "groupBy": "date"
            },
            "tickets": [
              {
                "id": "#T-2024-1247",
                "subject": "Login issues with mobile app",
                "customer": {
                  "name": "Sarah Johnson",
                  "avatar": "/avatars/sarah-j.jpg",
                  "tier": "Premium"
                },
                "priority": {
                  "level": "high",
                  "color": "#ef4444",
                  "label": "High"
                },
                "status": "open",
                "lastUpdate": "2 hours ago",
                "unread": true,
                "appearance": {
                  "selected": true,
                  "hoverEffect": true
                }
              },
              {
                "id": "#T-2024-1246",
                "subject": "Billing question about invoice",
                "customer": {
                  "name": "Michael Chen",
                  "avatar": "/avatars/michael-c.jpg",
                  "tier": "Business"
                },
                "priority": {
                  "level": "medium",
                  "color": "#f59e0b",
                  "label": "Medium"
                },
                "status": "pending",
                "lastUpdate": "1 day ago",
                "unread": false
              }
            ]
          }
        ]
      },
      "centerPanel": {
        "title": "Ticket Details",
        "appearance": {
          "backgroundColor": "white",
          "padding": "1.5rem"
        },
        "components": [
          {
            "type": "ticket_header",
            "ticket": {
              "id": "#T-2024-1247",
              "subject": "Login issues with mobile app",
              "status": {
                "current": "open",
                "options": ["open", "pending", "resolved", "closed"],
                "appearance": {
                  "dropdown": true,
                  "colors": {
                    "open": "#3b82f6",
                    "pending": "#f59e0b",
                    "resolved": "#10b981",
                    "closed": "#6b7280"
                  }
                }
              },
              "priority": {
                "current": "high",
                "options": ["low", "medium", "high", "urgent"],
                "editable": true
              },
              "assignee": {
                "current": "John Doe",
                "avatar": "/avatars/john-doe.jpg",
                "changeable": true,
                "options": "team_members"
              },
              "created": "2 hours ago",
              "lastUpdated": "15 minutes ago"
            }
          },
          {
            "type": "conversation_thread",
            "appearance": {
              "maxHeight": "600px",
              "scrollable": true,
              "messageSpacing": "1rem"
            },
            "messages": [
              {
                "type": "customer_message",
                "author": {
                  "name": "Sarah Johnson",
                  "avatar": "/avatars/sarah-j.jpg",
                  "role": "Customer"
                },
                "timestamp": "2 hours ago",
                "content": "Hi, I'm having trouble logging into the mobile app. It keeps saying 'Invalid credentials' even though I'm sure my password is correct. I can log in fine on the website.",
                "appearance": {
                  "backgroundColor": "#f1f5f9",
                  "alignment": "left",
                  "borderRadius": "12px 12px 12px 4px"
                },
                "attachments": [
                  {
                    "type": "image",
                    "name": "error-screenshot.jpg",
                    "size": "245 KB",
                    "thumbnail": "/uploads/thumbnails/error-screenshot.jpg"
                  }
                ]
              },
              {
                "type": "agent_message",
                "author": {
                  "name": "John Doe",
                  "avatar": "/avatars/john-doe.jpg",
                  "role": "Support Agent"
                },
                "timestamp": "1 hour ago",
                "content": "Hi Sarah, thank you for reaching out. I can see you're a Premium customer and I'd be happy to help resolve this login issue.\n\nLet's try a few troubleshooting steps:\n1. Force close the app completely and restart it\n2. Clear the app cache (Settings > Apps > [App Name] > Storage > Clear Cache)\n3. Try logging out of the website and back in to ensure your session is fresh\n\nCould you please try these steps and let me know if the issue persists?",
                "appearance": {
                  "backgroundColor": "#dbeafe",
                  "alignment": "right",
                  "borderRadius": "12px 12px 4px 12px"
                }
              },
              {
                "type": "customer_message",
                "author": {
                  "name": "Sarah Johnson",
                  "avatar": "/avatars/sarah-j.jpg",
                  "role": "Customer"
                },
                "timestamp": "15 minutes ago",
                "content": "I tried all those steps but I'm still getting the same error. The website login works perfectly fine, but the mobile app just won't accept my credentials.",
                "appearance": {
                  "backgroundColor": "#f1f5f9",
                  "alignment": "left"
                }
              }
            ]
          },
          {
            "type": "message_composer",
            "appearance": {
              "position": "bottom",
              "backgroundColor": "white",
              "border": "1px solid #e2e8f0",
              "borderRadius": "8px",
              "padding": "1rem"
            },
            "features": {
              "richText": {
                "enabled": true,
                "toolbar": ["bold", "italic", "link", "bullet_list", "code"]
              },
              "attachments": {
                "enabled": true,
                "dragDrop": true,
                "maxSize": "25MB",
                "allowedTypes": ["image", "document", "video"]
              },
              "templates": {
                "enabled": true,
                "quick_responses": [
                  "Thank you for contacting support...",
                  "I understand your frustration...",
                  "Let me escalate this to our technical team..."
                ]
              },
              "mentions": {
                "enabled": true,
                "trigger": "@",
                "options": "team_members"
              }
            },
            "actions": [
              {
                "text": "Send",
                "variant": "primary",
                "shortcut": "Ctrl+Enter"
              },
              {
                "text": "Send & Close",
                "variant": "secondary"
              },
              {
                "icon": "paperclip",
                "tooltip": "Attach file"
              },
              {
                "icon": "template",
                "tooltip": "Use template"
              }
            ]
          }
        ]
      },
      "rightPanel": {
        "width": "300px",
        "title": "Customer Information",
        "appearance": {
          "backgroundColor": "#f8fafc",
          "borderLeft": "1px solid #e2e8f0",
          "padding": "1.5rem"
        },
        "widgets": [
          {
            "type": "customer_profile",
            "customer": {
              "name": "Sarah Johnson",
              "email": "sarah.johnson@company.com",
              "phone": "+1 (555) 123-4567",
              "avatar": "/avatars/sarah-j.jpg",
              "tier": {
                "name": "Premium",
                "color": "#8b5cf6",
                "benefits": ["Priority Support", "Advanced Features"]
              },
              "accountSince": "March 2023",
              "location": "New York, NY"
            }
          },
          {
            "type": "ticket_history",
            "title": "Recent Tickets",
            "tickets": [
              {
                "id": "#T-2024-1156",
                "subject": "Feature request: Dark mode",
                "status": "resolved",
                "date": "1 week ago"
              },
              {
                "id": "#T-2024-1089",
                "subject": "Payment processing issue",
                "status": "resolved",
                "date": "2 weeks ago"
              }
            ],
            "actions": {
              "viewAll": {
                "enabled": true,
                "text": "View all tickets"
              }
            }
          },
          {
            "type": "account_summary",
            "title": "Account Summary",
            "data": {
              "plan": "Premium Monthly",
              "status": "Active",
              "nextBilling": "Nov 15, 2025",
              "totalSpent": "$2,847",
              "satisfactionScore": {
                "score": 4.8,
                "outOf": 5,
                "basedOn": "12 interactions"
              }
            }
          },
          {
            "type": "internal_notes",
            "title": "Internal Notes",
            "appearance": {
              "textarea": true,
              "height": "120px",
              "placeholder": "Add private notes about this customer..."
            },
            "features": {
              "autoSave": true,
              "visibility": "agents_only"
            }
          }
        ]
      }
    },
    "globalFeatures": {
      "notifications": {
        "newTickets": true,
        "mentions": true,
        "escalations": true,
        "position": "top_right"
      },
      "keyboard_shortcuts": {
        "enabled": true,
        "shortcuts": [
          {"key": "R", "action": "Reply"},
          {"key": "A", "action": "Assign to me"},
          {"key": "C", "action": "Close ticket"},
          {"key": "E", "action": "Escalate"}
        ]
      },
      "collaboration": {
        "realTimeEditing": true,
        "typingIndicators": true,
        "presenceAwareness": true
      }
    },
    "mobileExperience": {
      "layout": "single_panel",
      "navigation": {
        "tabBar": ["Tickets", "Conversation", "Customer"],
        "backButton": true
      },
      "swipeGestures": {
        "leftSwipe": "Mark as resolved",
        "rightSwipe": "Assign to me"
      },
      "voiceNotes": {
        "enabled": true,
        "maxDuration": "5 minutes"
      }
    }
  }
}
```

This comprehensive guide provides the foundation for implementing any type of form or UI component using the Schema Engine with full design system integration.