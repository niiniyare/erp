# AMIS for Backend Developers
## Complete Low-Code Frontend Framework Guide for Termux

> **Target Audience**: Backend developers who want to build production-ready admin UIs without writing JavaScript/TypeScript  
> **Environment**: Termux (Android terminal) with limited resources  
> **Version**: AMIS 6.13.0 (Latest stable release)  
> **Method**: Pure HTML + JSON configuration (SDK approach)

---

## Table of Contents

### Part 1: Fundamentals
1. [Quick Start](#quick-start)
2. [Understanding AMIS](#understanding-amis)
3. [Termux Setup](#termux-setup)
4. [Project Structure](#project-structure)
5. [Your First Page](#your-first-page)

### Part 2: Core Features
6. [Components Library](#components-library)
7. [Data Flow & APIs](#data-flow--apis)
8. [Event-Action System](#event-action-system)
9. [Data Mapping & Transformations](#data-mapping--transformations)
10. [Expressions & Formulas](#expressions--formulas)

### Part 3: Advanced Topics
11. [Building Real Applications](#building-real-applications)
12. [Advanced Patterns](#advanced-patterns)
13. [Custom Filters & Functions](#custom-filters--functions)
14. [State Management](#state-management)
15. [External JSON Files](#external-json-files)
16. [Performance Optimization](#performance-optimization)

### Part 4: Termux-Specific
17. [Termux Development Workflow](#termux-development-workflow)
18. [Memory & Resource Management](#memory--resource-management)
19. [Troubleshooting](#troubleshooting)

### Part 5: Reference
20. [Component Reference](#component-reference)
21. [API Reference](#api-reference)
22. [Resources & Links](#resources--links)

---

## Quick Start

### What is AMIS?

AMIS is a **JSON-to-UI** framework. You write JSON, it renders HTML. No JavaScript required.

```
Backend Developer → JSON Config → AMIS SDK → Beautiful UI
```

**Why AMIS for Backend Devs?**
- Write UIs like you write config files
- No npm, webpack, React, or build tools
- Just HTML files + JSON
- Perfect for admin panels and CRUD interfaces

**What You Can Build:**
- Admin dashboards
- Database management UIs
- API testing interfaces
- Data visualization tools
- Form-heavy applications

---

## Understanding AMIS

### The Core Concept

Everything in AMIS is a **component** defined by JSON:

```json
{
  "type": "input-text",
  "name": "username",
  "label": "Username"
}
```

This JSON becomes a fully functional input field. That's it.

### How It Works

```
┌─────────────┐
│  Your JSON  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  AMIS SDK   │  (Handles all the React/JS complexity)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  HTML/CSS   │  (Rendered in browser)
└─────────────┘
```

You only touch the JSON. AMIS handles everything else.

---

## Termux Setup

### Prerequisites

```bash
# Update packages
pkg update && pkg upgrade

# Install Node.js (needed only for getting AMIS files)
pkg install nodejs-lts

# Install a simple web server
npm install -g http-server
```

### Getting AMIS SDK

**Option 1: Via npm (Quick)**
```bash
# Create project directory
mkdir ~/amis
cd ~/amis

# Install AMIS
npm init -y
npm install amis

# SDK files are now in: node_modules/amis/sdk/
```

**Option 2: Direct Download (If npm is slow)**
```bash
# Download from GitHub releases
cd ~/amis
wget https://github.com/baidu/amis/releases/latest/download/sdk.tar.gz
tar -xzf sdk.tar.gz
```

### Verify Installation

```bash
ls node_modules/amis/sdk/
# Should see: sdk.css, sdk.js, and other files
```

---

## Project Structure

### Recommended Layout

```
~/amis/
├── sdk/              # AMIS SDK files (copy from node_modules)
│   ├── sdk.css
│   ├── sdk.js
│   └── iconfont.css
├── pages/            # Your HTML pages
│   ├── index.html
│   ├── users.html
│   └── dashboard.html
├── schemas/          # JSON configurations (optional organization)
│   ├── user-form.json
│   └── user-table.json
└── data/             # Mock data for testing (optional)
    └── sample.json
```

### Setup Script

Create `setup.sh` in your project root:

```bash
#!/bin/bash
# Copy SDK files to project
mkdir -p sdk
cp node_modules/amis/sdk/sdk.css sdk/
cp node_modules/amis/sdk/sdk.js sdk/
cp node_modules/amis/sdk/iconfont.css sdk/

# Create directory structure
mkdir -p pages schemas data

echo "Setup complete! SDK copied to ./sdk/"
```

Run it:
```bash
chmod +x setup.sh
./setup.sh
```

---

## Your First Page

### Minimal Template

Create `pages/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AMIS App</title>
  
  <!-- AMIS CSS -->
  <link rel="stylesheet" href="../sdk/sdk.css">
  <link rel="stylesheet" href="../sdk/iconfont.css">
  
  <style>
    /* Full height container */
    html, body, #root {
      height: 100%;
      margin: 0;
      padding: 0;
    }
  </style>
</head>
<body>
  <!-- AMIS renders here -->
  <div id="root"></div>
  
  <!-- AMIS SDK -->
  <script src="../sdk/sdk.js"></script>
  
  <script>
    (function() {
      // Get AMIS
      const amis = amisRequire('amis/embed');
      
      // Your JSON schema
      const schema = {
        type: 'page',
        title: 'Hello AMIS',
        body: 'Welcome to AMIS! This is pure JSON-driven UI.'
      };
      
      // Render
      amis.embed('#root', schema);
    })();
  </script>
</body>
</html>
```

### Test It

```bash
# Start web server
cd ~/amis
http-server -p 8080

# Open in browser (if using Termux on Android)
# Use VNC viewer or install Termux:API for browser access
# Or access from another device: http://<your-phone-ip>:8080
```

---

## Core Components

### 1. Page - The Container

Every AMIS app starts with a page:

```json
{
  "type": "page",
  "title": "My Admin Panel",
  "body": [
    // Your components go here
  ]
}
```

### 2. Forms - Data Input

**Basic Form:**
```json
{
  "type": "form",
  "api": "post:http://your-api.com/save",
  "body": [
    {
      "type": "input-text",
      "name": "username",
      "label": "Username",
      "required": true
    },
    {
      "type": "input-email",
      "name": "email",
      "label": "Email",
      "required": true
    },
    {
      "type": "input-password",
      "name": "password",
      "label": "Password",
      "required": true
    }
  ]
}
```

**Common Form Fields:**

| Type | Purpose | Example |
|------|---------|---------|
| `input-text` | Text input | Username, name, etc. |
| `input-email` | Email validation | Email addresses |
| `input-password` | Password field | Passwords |
| `input-number` | Numbers only | Age, quantity |
| `textarea` | Multi-line text | Description, bio |
| `select` | Dropdown | Country, status |
| `date` | Date picker | Birth date, deadline |
| `switch` | On/off toggle | Active/inactive |
| `checkbox` | Single checkbox | Accept terms |
| `checkboxes` | Multiple checkboxes | Select multiple items |
| `radios` | Radio buttons | Single choice |

### 3. Tables - Data Display

**Basic Table:**
```json
{
  "type": "table",
  "data": {
    "items": [
      {"id": 1, "name": "John", "email": "john@example.com"},
      {"id": 2, "name": "Jane", "email": "jane@example.com"}
    ]
  },
  "columns": [
    {"name": "id", "label": "ID"},
    {"name": "name", "label": "Name"},
    {"name": "email", "label": "Email"}
  ]
}
```

**Table from API:**
```json
{
  "type": "service",
  "api": "http://your-api.com/users",
  "body": {
    "type": "table",
    "columns": [
      {"name": "id", "label": "ID"},
      {"name": "name", "label": "Name"},
      {"name": "email", "label": "Email"}
    ]
  }
}
```

### 4. CRUD - Complete Data Management

CRUD combines table + forms for full data management:

```json
{
  "type": "crud",
  "api": "http://your-api.com/users",
  "columns": [
    {"name": "id", "label": "ID"},
    {"name": "name", "label": "Name"},
    {"name": "email", "label": "Email"},
    {
      "type": "operation",
      "label": "Actions",
      "buttons": [
        {
          "label": "Edit",
          "type": "button",
          "actionType": "dialog",
          "dialog": {
            "title": "Edit User",
            "body": {
              "type": "form",
              "api": "put:http://your-api.com/users/$id",
              "body": [
                {
                  "type": "input-text",
                  "name": "name",
                  "label": "Name"
                },
                {
                  "type": "input-email",
                  "name": "email",
                  "label": "Email"
                }
              ]
            }
          }
        },
        {
          "label": "Delete",
          "type": "button",
          "actionType": "ajax",
          "confirmText": "Sure to delete?",
          "api": "delete:http://your-api.com/users/$id"
        }
      ]
    }
  ]
}
```

### 5. Cards - Visual Display

```json
{
  "type": "cards",
  "source": "${items}",
  "card": {
    "header": {
      "title": "$name",
      "subTitle": "$email"
    },
    "body": [
      {
        "type": "tpl",
        "tpl": "ID: ${id}"
      }
    ],
    "actions": [
      {
        "type": "button",
        "label": "View",
        "actionType": "link",
        "link": "/user/${id}"
      }
    ]
  }
}
```

### 6. Charts - Data Visualization

```json
{
  "type": "chart",
  "api": "http://your-api.com/chart-data",
  "config": {
    "title": {
      "text": "Sales Chart"
    },
    "xAxis": {
      "type": "category",
      "data": ["Mon", "Tue", "Wed", "Thu", "Fri"]
    },
    "yAxis": {
      "type": "value"
    },
    "series": [{
      "data": [120, 200, 150, 80, 70],
      "type": "bar"
    }]
  }
}
```

---

## Building Real Applications

### Example 1: User Management System

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>User Management</title>
  <link rel="stylesheet" href="../sdk/sdk.css">
  <link rel="stylesheet" href="../sdk/iconfont.css">
  <style>
    html, body, #root { height: 100%; margin: 0; }
  </style>
</head>
<body>
  <div id="root"></div>
  <script src="../sdk/sdk.js"></script>
  <script>
    (function() {
      const amis = amisRequire('amis/embed');
      
      const schema = {
        type: 'page',
        title: 'User Management System',
        body: {
          type: 'crud',
          syncLocation: false,
          api: 'http://localhost:3000/api/users',
          
          // Search filters
          filter: {
            title: 'Search Users',
            body: [
              {
                type: 'input-text',
                name: 'keywords',
                label: 'Keywords',
                placeholder: 'Search by name or email'
              },
              {
                type: 'select',
                name: 'status',
                label: 'Status',
                options: [
                  {label: 'All', value: ''},
                  {label: 'Active', value: 'active'},
                  {label: 'Inactive', value: 'inactive'}
                ]
              }
            ]
          },
          
          // Bulk actions
          bulkActions: [
            {
              label: 'Batch Delete',
              actionType: 'ajax',
              api: 'delete:http://localhost:3000/api/users/batch',
              confirmText: 'Delete selected users?'
            }
          ],
          
          // Table columns
          columns: [
            {
              name: 'id',
              label: 'ID',
              sortable: true
            },
            {
              name: 'name',
              label: 'Name',
              sortable: true,
              searchable: true
            },
            {
              name: 'email',
              label: 'Email',
              sortable: true
            },
            {
              name: 'role',
              label: 'Role',
              type: 'mapping',
              map: {
                'admin': '<span class="label label-success">Admin</span>',
                'user': '<span class="label label-default">User</span>'
              }
            },
            {
              name: 'status',
              label: 'Status',
              type: 'status'
            },
            {
              name: 'created_at',
              label: 'Created',
              type: 'date',
              format: 'YYYY-MM-DD'
            },
            {
              type: 'operation',
              label: 'Operations',
              buttons: [
                {
                  label: 'View',
                  type: 'button',
                  level: 'link',
                  actionType: 'drawer',
                  drawer: {
                    title: 'User Details',
                    body: {
                      type: 'form',
                      body: [
                        {
                          type: 'static',
                          name: 'id',
                          label: 'ID'
                        },
                        {
                          type: 'static',
                          name: 'name',
                          label: 'Name'
                        },
                        {
                          type: 'static',
                          name: 'email',
                          label: 'Email'
                        },
                        {
                          type: 'static',
                          name: 'role',
                          label: 'Role'
                        }
                      ]
                    }
                  }
                },
                {
                  label: 'Edit',
                  type: 'button',
                  level: 'link',
                  actionType: 'dialog',
                  dialog: {
                    title: 'Edit User',
                    size: 'lg',
                    body: {
                      type: 'form',
                      api: 'put:http://localhost:3000/api/users/$id',
                      body: [
                        {
                          type: 'input-text',
                          name: 'name',
                          label: 'Name',
                          required: true
                        },
                        {
                          type: 'input-email',
                          name: 'email',
                          label: 'Email',
                          required: true
                        },
                        {
                          type: 'select',
                          name: 'role',
                          label: 'Role',
                          options: [
                            {label: 'Admin', value: 'admin'},
                            {label: 'User', value: 'user'}
                          ]
                        },
                        {
                          type: 'switch',
                          name: 'status',
                          label: 'Active',
                          option: 'Active'
                        }
                      ]
                    }
                  }
                },
                {
                  label: 'Delete',
                  type: 'button',
                  level: 'link',
                  className: 'text-danger',
                  actionType: 'ajax',
                  confirmText: 'Are you sure you want to delete this user?',
                  api: 'delete:http://localhost:3000/api/users/$id'
                }
              ]
            }
          ]
        }
      };
      
      amis.embed('#root', schema);
    })();
  </script>
</body>
</html>
```

### Example 2: Dashboard with Multiple Widgets

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Dashboard</title>
  <link rel="stylesheet" href="../sdk/sdk.css">
  <link rel="stylesheet" href="../sdk/iconfont.css">
  <style>
    html, body, #root { height: 100%; margin: 0; }
  </style>
</head>
<body>
  <div id="root"></div>
  <script src="../sdk/sdk.js"></script>
  <script>
    (function() {
      const amis = amisRequire('amis/embed');
      
      const schema = {
        type: 'page',
        title: 'Dashboard',
        body: [
          // Stats cards
          {
            type: 'grid',
            columns: [
              {
                body: {
                  type: 'card',
                  className: 'bg-primary text-white',
                  header: {
                    title: 'Total Users',
                    subTitle: 'All registered users'
                  },
                  body: {
                    type: 'service',
                    api: 'http://localhost:3000/api/stats/users',
                    body: {
                      type: 'tpl',
                      tpl: '<h1 class="text-white">${count}</h1>'
                    }
                  }
                }
              },
              {
                body: {
                  type: 'card',
                  className: 'bg-success text-white',
                  header: {
                    title: 'Active Sessions',
                    subTitle: 'Current online users'
                  },
                  body: {
                    type: 'service',
                    api: 'http://localhost:3000/api/stats/sessions',
                    body: {
                      type: 'tpl',
                      tpl: '<h1 class="text-white">${count}</h1>'
                    }
                  }
                }
              },
              {
                body: {
                  type: 'card',
                  className: 'bg-warning text-white',
                  header: {
                    title: 'Pending Tasks',
                    subTitle: 'Tasks requiring attention'
                  },
                  body: {
                    type: 'service',
                    api: 'http://localhost:3000/api/stats/tasks',
                    body: {
                      type: 'tpl',
                      tpl: '<h1 class="text-white">${count}</h1>'
                    }
                  }
                }
              },
              {
                body: {
                  type: 'card',
                  className: 'bg-danger text-white',
                  header: {
                    title: 'Errors',
                    subTitle: 'Last 24 hours'
                  },
                  body: {
                    type: 'service',
                    api: 'http://localhost:3000/api/stats/errors',
                    body: {
                      type: 'tpl',
                      tpl: '<h1 class="text-white">${count}</h1>'
                    }
                  }
                }
              }
            ]
          },
          
          // Charts row
          {
            type: 'grid',
            columns: [
              {
                md: 8,
                body: {
                  type: 'panel',
                  title: 'User Activity',
                  body: {
                    type: 'chart',
                    api: 'http://localhost:3000/api/charts/activity',
                    config: {
                      xAxis: {
                        type: 'category',
                        data: '${xData}'
                      },
                      yAxis: {
                        type: 'value'
                      },
                      series: [{
                        data: '${yData}',
                        type: 'line',
                        smooth: true
                      }]
                    }
                  }
                }
              },
              {
                md: 4,
                body: {
                  type: 'panel',
                  title: 'User Distribution',
                  body: {
                    type: 'chart',
                    api: 'http://localhost:3000/api/charts/distribution',
                    config: {
                      series: [{
                        type: 'pie',
                        data: '${data}',
                        radius: '70%'
                      }]
                    }
                  }
                }
              }
            ]
          },
          
          // Recent activity table
          {
            type: 'panel',
            title: 'Recent Activity',
            body: {
              type: 'service',
              api: 'http://localhost:3000/api/activity/recent',
              body: {
                type: 'table',
                columns: [
                  {
                    name: 'user',
                    label: 'User'
                  },
                  {
                    name: 'action',
                    label: 'Action'
                  },
                  {
                    name: 'timestamp',
                    label: 'Time',
                    type: 'datetime'
                  }
                ]
              }
            }
          }
        ]
      };
      
      amis.embed('#root', schema);
    })();
  </script>
</body>
</html>
```

---

## Data Flow & APIs

### How AMIS Calls Your Backend

AMIS automatically handles HTTP requests. You just configure the endpoint:

```json
{
  "type": "crud",
  "api": "http://localhost:3000/api/users"
}
```

This will make:
- `GET /api/users` - List users
- `POST /api/users` - Create user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

### API Configuration Options

**Simple string:**
```json
{
  "api": "http://localhost:3000/api/users"
}
```

**With HTTP method:**
```json
{
  "api": "post:http://localhost:3000/api/users"
}
```

**Advanced configuration:**
```json
{
  "api": {
    "method": "post",
    "url": "http://localhost:3000/api/users",
    "data": {
      "name": "${name}",
      "email": "${email}"
    },
    "headers": {
      "Authorization": "Bearer ${token}"
    }
  }
}
```

### Expected Response Format

AMIS expects this JSON structure:

```json
{
  "status": 0,          // 0 = success, non-zero = error
  "msg": "Success",     // Message to display
  "data": {             // Your data
    "items": [...],     // For lists
    "total": 100        // For pagination
  }
}
```

### Example Backend Response (Node.js)

```javascript
// Success response
res.json({
  status: 0,
  msg: "Users loaded successfully",
  data: {
    items: [
      {id: 1, name: "John", email: "john@example.com"},
      {id: 2, name: "Jane", email: "jane@example.com"}
    ],
    total: 2
  }
});

// Error response
res.json({
  status: 422,
  msg: "Validation failed",
  errors: {
    email: "Invalid email format"
  }
});
```

### Using Variables in API Calls

Access form data or context variables with `${}`:

```json
{
  "type": "button",
  "label": "Delete",
  "api": "delete:http://localhost:3000/api/users/${id}"
}
```

The `${id}` will be replaced with the actual ID from the data context.

---

## Common Patterns

### Pattern 1: Multi-Step Form (Wizard)

```json
{
  "type": "wizard",
  "api": "post:http://localhost:3000/api/register",
  "steps": [
    {
      "title": "Account Info",
      "body": [
        {
          "type": "input-text",
          "name": "username",
          "label": "Username",
          "required": true
        },
        {
          "type": "input-password",
          "name": "password",
          "label": "Password",
          "required": true
        }
      ]
    },
    {
      "title": "Personal Info",
      "body": [
        {
          "type": "input-text",
          "name": "first_name",
          "label": "First Name"
        },
        {
          "type": "input-text",
          "name": "last_name",
          "label": "Last Name"
        }
      ]
    },
    {
      "title": "Confirm",
      "body": [
        {
          "type": "tpl",
          "tpl": "<h3>Review Your Information</h3><p>Username: ${username}</p><p>Name: ${first_name} ${last_name}</p>"
        }
      ]
    }
  ]
}
```

### Pattern 2: Master-Detail View

```json
{
  "type": "page",
  "body": {
    "type": "grid",
    "columns": [
      {
        "md": 4,
        "body": {
          "type": "service",
          "api": "http://localhost:3000/api/categories",
          "body": {
            "type": "list",
            "source": "${items}",
            "listItem": {
              "title": "${name}",
              "actions": [
                {
                  "type": "button",
                  "label": "View",
                  "onEvent": {
                    "click": {
                      "actions": [
                        {
                          "actionType": "reload",
                          "componentId": "detail-view",
                          "data": {
                            "category_id": "${id}"
                          }
                        }
                      ]
                    }
                  }
                }
              ]
            }
          }
        }
      },
      {
        "md": 8,
        "body": {
          "type": "service",
          "id": "detail-view",
          "api": "http://localhost:3000/api/items?category_id=${category_id}",
          "body": {
            "type": "table",
            "columns": [
              {"name": "name", "label": "Name"},
              {"name": "description", "label": "Description"},
              {"name": "price", "label": "Price"}
            ]
          }
        }
      }
    ]
  }
}
```

### Pattern 3: Conditional Fields

Show/hide fields based on other field values:

```json
{
  "type": "form",
  "body": [
    {
      "type": "select",
      "name": "user_type",
      "label": "User Type",
      "options": [
        {"label": "Individual", "value": "individual"},
        {"label": "Business", "value": "business"}
      ]
    },
    {
      "type": "input-text",
      "name": "company_name",
      "label": "Company Name",
      "visibleOn": "this.user_type === 'business'"
    },
    {
      "type": "input-text",
      "name": "tax_id",
      "label": "Tax ID",
      "visibleOn": "this.user_type === 'business'"
    }
  ]
}
```

### Pattern 4: File Upload

```json
{
  "type": "form",
  "api": "post:http://localhost:3000/api/upload",
  "body": [
    {
      "type": "input-file",
      "name": "file",
      "label": "Upload File",
      "accept": ".jpg,.png,.pdf",
      "maxSize": 5242880,
      "receiver": "http://localhost:3000/api/upload/receiver",
      "required": true
    }
  ]
}
```

### Pattern 5: Dynamic Options from API

```json
{
  "type": "form",
  "body": [
    {
      "type": "select",
      "name": "country",
      "label": "Country",
      "source": "http://localhost:3000/api/countries",
      "labelField": "name",
      "valueField": "code"
    },
    {
      "type": "select",
      "name": "city",
      "label": "City",
      "source": "http://localhost:3000/api/cities?country=${country}",
      "labelField": "name",
      "valueField": "id"
    }
  ]
}
```

---

## External JSON Files

### Why Use External JSON Files?

**Benefits:**
1. **Separation of Concerns** - Keep schemas separate from HTML
2. **Reusability** - Use same schema in multiple pages
3. **Version Control** - Better Git diffs and conflict resolution
4. **Team Collaboration** - Developers can work on different schemas
5. **Dynamic Loading** - Load schemas based on user role/permissions
6. **Easy Maintenance** - Update schemas without touching HTML
7. **AI-Friendly** - AI assistants can generate/modify JSON files easily
8. **Storage Efficiency** - Smaller HTML files, better caching

### Project Structure with JSON Files

```
~/amis/
├── sdk/
│   ├── sdk.css
│   ├── sdk.js
│   └── iconfont.css
├── pages/
│   ├── index.html          # Router/launcher page
│   ├── viewer.html         # Generic schema viewer
│   └── template.html       # Reusable template
├── schemas/
│   ├── pages/              # Full page schemas
│   │   ├── dashboard.json
│   │   ├── users.json
│   │   └── settings.json
│   ├── forms/              # Reusable form schemas
│   │   ├── user-form.json
│   │   └── login-form.json
│   ├── tables/             # Reusable table schemas
│   │   ├── user-table.json
│   │   └── product-table.json
│   └── components/         # Reusable components
│       ├── header.json
│       └── sidebar.json
├── utils/
│   └── schema-loader.js    # Schema loading utility
└── data/
    └── mock/               # Mock data for testing
        ├── users.json
        └── products.json
```

### Basic Implementation

**Method 1: Simple Fetch**

Create `pages/viewer.html`:
```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AMIS Page Viewer</title>
  <link rel="stylesheet" href="../sdk/sdk.css">
  <link rel="stylesheet" href="../sdk/iconfont.css">
  <style>
    html, body, #root { height: 100%; margin: 0; padding: 0; }
    .loading { 
      display: flex; 
      justify-content: center; 
      align-items: center; 
      height: 100vh; 
    }
  </style>
</head>
<body>
  <div id="root">
    <div class="loading">Loading...</div>
  </div>
  
  <script src="../sdk/sdk.js"></script>
  <script>
    (function() {
      const amis = amisRequire('amis/embed');
      
      // Get schema file from URL parameter
      const urlParams = new URLSearchParams(window.location.search);
      const schemaFile = urlParams.get('schema') || 'dashboard';
      
      // Load schema
      fetch(`../schemas/pages/${schemaFile}.json`)
        .then(response => {
          if (!response.ok) {
            throw new Error(`Schema not found: ${schemaFile}`);
          }
          return response.json();
        })
        .then(schema => {
          amis.embed('#root', schema);
        })
        .catch(error => {
          document.getElementById('root').innerHTML = `
            <div class="loading">
              <h2>Error: ${error.message}</h2>
              <p>Failed to load schema: ${schemaFile}.json</p>
            </div>
          `;
        });
    })();
  </script>
</body>
</html>
```

**Usage:**
- `viewer.html?schema=dashboard` → Loads `schemas/pages/dashboard.json`
- `viewer.html?schema=users` → Loads `schemas/pages/users.json`
- `viewer.html?schema=settings` → Loads `schemas/pages/settings.json`

### Method 2: Hash-Based Router

Create `pages/index.html` with routing:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>AMIS App</title>
  <link rel="stylesheet" href="../sdk/sdk.css">
  <link rel="stylesheet" href="../sdk/iconfont.css">
  <style>
    html, body, #root { height: 100%; margin: 0; }
  </style>
</head>
<body>
  <div id="root"></div>
  
  <script src="../sdk/sdk.js"></script>
  <script>
    (function() {
      const amis = amisRequire('amis/embed');
      
      const routes = {
        'dashboard': '../schemas/pages/dashboard.json',
        'users': '../schemas/pages/users.json',
        'products': '../schemas/pages/products.json'
      };
      
      function loadPage(route) {
        const schemaPath = routes[route] || routes['dashboard'];
        
        fetch(schemaPath)
          .then(res => res.json())
          .then(schema => amis.embed('#root', schema))
          .catch(error => console.error('Load failed:', error));
      }
      
      window.addEventListener('hashchange', () => {
        loadPage(window.location.hash.slice(1) || 'dashboard');
      });
      
      loadPage(window.location.hash.slice(1) || 'dashboard');
    })();
  </script>
</body>
</html>
```

### Method 3: Schema Composition with $ref

Build modular schemas using references:

**Main schema** (`schemas/pages/dashboard.json`):
```json
{
  "type": "page",
  "title": "Dashboard",
  "body": [
    {
      "$ref": "components/header.json"
    },
    {
      "$ref": "components/stats.json"
    }
  ]
}
```

**Component** (`schemas/components/header.json`):
```json
{
  "type": "panel",
  "title": "Welcome",
  "body": {
    "type": "tpl",
    "tpl": "Hello ${currentUser}!"
  }
}
```

**Loader with $ref resolution:**
```javascript
async function resolveRefs(schema, basePath = '../schemas/') {
  if (typeof schema !== 'object' || schema === null) {
    return schema;
  }
  
  if (Array.isArray(schema)) {
    return Promise.all(schema.map(item => resolveRefs(item, basePath)));
  }
  
  if (schema.$ref) {
    const response = await fetch(basePath + schema.$ref);
    const referenced = await response.json();
    return resolveRefs(referenced, basePath);
  }
  
  const resolved = {};
  for (const [key, value] of Object.entries(schema)) {
    resolved[key] = await resolveRefs(value, basePath);
  }
  
  return resolved;
}

// Usage
const schema = await fetch('../schemas/pages/dashboard.json').then(r => r.json());
const resolved = await resolveRefs(schema);
amis.embed('#root', resolved);
```

### Schema Loader Utility

Create `utils/schema-loader.js`:

```javascript
/**
 * AMIS Schema Loader - Handles loading and caching
 */
class SchemaLoader {
  constructor(baseUrl = '../schemas/') {
    this.baseUrl = baseUrl;
    this.cache = new Map();
  }
  
  async load(path) {
    const fullPath = this.baseUrl + path;
    
    if (this.cache.has(fullPath)) {
      return this.cache.get(fullPath);
    }
    
    const response = await fetch(fullPath);
    if (!response.ok) {
      throw new Error(`Failed to load: ${path}`);
    }
    
    const schema = await response.json();
    this.cache.set(fullPath, schema);
    return schema;
  }
  
  async resolveRefs(schema) {
    if (typeof schema !== 'object' || schema === null) {
      return schema;
    }
    
    if (Array.isArray(schema)) {
      return Promise.all(schema.map(item => this.resolveRefs(item)));
    }
    
    if (schema.$ref) {
      const referenced = await this.load(schema.$ref);
      return this.resolveRefs(referenced);
    }
    
    const resolved = {};
    for (const [key, value] of Object.entries(schema)) {
      resolved[key] = await this.resolveRefs(value);
    }
    
    return resolved;
  }
  
  clearCache() {
    this.cache.clear();
  }
}
```

**Usage:**
```html
<script src="../utils/schema-loader.js"></script>
<script>
  const loader = new SchemaLoader('../schemas/');
  const schema = await loader.load('pages/dashboard.json');
  const resolved = await loader.resolveRefs(schema);
  amis.embed('#root', resolved);
</script>
```

### Role-Based Schema Loading

Load different schemas based on user permissions:

```javascript
const userRole = localStorage.getItem('userRole') || 'user';

const schemaMap = {
  'admin': {
    'dashboard': '../schemas/admin/dashboard.json',
    'users': '../schemas/admin/users.json'
  },
  'user': {
    'dashboard': '../schemas/user/dashboard.json',
    'profile': '../schemas/user/profile.json'
  }
};

function loadByRole(page) {
  const schemas = schemaMap[userRole];
  const schemaPath = schemas[page];
  
  fetch(schemaPath)
    .then(res => res.json())
    .then(schema => {
      schema.data = schema.data || {};
      schema.data.userRole = userRole;
      amis.embed('#root', schema);
    });
}
```

### Best Practices

**1. File Organization:**
```
schemas/
  pages/          # Complete pages
  forms/          # Reusable forms
  tables/         # Reusable tables
  components/     # Small reusable pieces
  templates/      # Template schemas
```

**2. Naming Convention:**
```
kebab-case: user-management.json
            product-list.json
            order-form.json
```

**3. Schema Metadata:**
```json
{
  "_comment": "User Management Page",
  "_version": "1.0.0",
  "_author": "Backend Team",
  "_updated": "2026-01-26",
  
  "type": "page",
  "title": "Users"
}
```

**4. Validation Script:**

Create `schemas/validate.js`:
```javascript
const fs = require('fs');
const path = require('path');

function validateSchema(filePath) {
  try {
    const content = fs.readFileSync(filePath, 'utf8');
    const schema = JSON.parse(content);
    
    if (!schema.type) {
      throw new Error('Missing type field');
    }
    
    console.log(`✓ ${filePath}`);
    return true;
  } catch (error) {
    console.error(`✗ ${filePath}: ${error.message}`);
    return false;
  }
}

function validateDir(dir) {
  let valid = 0, invalid = 0;
  
  fs.readdirSync(dir).forEach(file => {
    const fullPath = path.join(dir, file);
    if (fs.statSync(fullPath).isDirectory()) {
      const result = validateDir(fullPath);
      valid += result.valid;
      invalid += result.invalid;
    } else if (file.endsWith('.json')) {
      validateSchema(fullPath) ? valid++ : invalid++;
    }
  });
  
  return { valid, invalid };
}

const result = validateDir('./schemas');
console.log(`\n${result.valid} valid, ${result.invalid} invalid`);
```

Run: `node schemas/validate.js`

### Termux Helper Scripts

**1. Schema Generator** (`create-schema.sh`):
```bash
#!/bin/bash
SCHEMA_NAME=$1
SCHEMA_FILE="schemas/pages/${SCHEMA_NAME}.json"

cat > $SCHEMA_FILE << EOF
{
  "_comment": "${SCHEMA_NAME} page",
  "_created": "$(date +%Y-%m-%d)",
  "type": "page",
  "title": "${SCHEMA_NAME}",
  "body": {
    "type": "tpl",
    "tpl": "Page content"
  }
}
EOF

echo "Created: $SCHEMA_FILE"
```

Usage: `./create-schema.sh my-page`

**2. Schema Minifier:**
```bash
pkg install jq
jq -c . schemas/pages/large.json > schemas/pages/large.min.json
```

### Complete Example

**Multi-page app** (`pages/app.html`):
```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>AMIS App</title>
  <link rel="stylesheet" href="../sdk/sdk.css">
  <link rel="stylesheet" href="../sdk/iconfont.css">
  <style>
    html, body { height: 100%; margin: 0; }
    #app { display: flex; flex-direction: column; height: 100%; }
    #nav { background: #2c3e50; color: white; padding: 10px; }
    #nav a { color: white; margin: 0 15px; }
    #content { flex: 1; overflow: auto; }
  </style>
</head>
<body>
  <div id="app">
    <div id="nav">
      <a href="#dashboard">Dashboard</a>
      <a href="#users">Users</a>
      <a href="#settings">Settings</a>
    </div>
    <div id="content"></div>
  </div>
  
  <script src="../sdk/sdk.js"></script>
  <script src="../utils/schema-loader.js"></script>
  <script>
    const amis = amisRequire('amis/embed');
    const loader = new SchemaLoader('../schemas/');
    
    async function navigate() {
      const page = window.location.hash.slice(1) || 'dashboard';
      try {
        const schema = await loader.load(`pages/${page}.json`);
        const resolved = await loader.resolveRefs(schema);
        amis.embed('#content', resolved);
      } catch (error) {
        document.getElementById('content').innerHTML = 
          `<div style="padding:20px">Error: ${error.message}</div>`;
      }
    }
    
    window.addEventListener('hashchange', navigate);
    navigate();
  </script>
</body>
</html>
```

This approach gives you complete flexibility to manage your AMIS application through JSON files!

---

## Termux Optimization

### Memory Management

Termux has limited RAM. Follow these practices:

**1. Minimize SDK Files**
Only include what you need:
```bash
# Instead of copying entire sdk folder
cp node_modules/amis/sdk/sdk.css sdk/
cp node_modules/amis/sdk/sdk.js sdk/
# Skip unnecessary files
```

**2. Use Simple HTTP Server**
```bash
# Lightweight option
python -m http.server 8080

# Or Node.js (if already installed)
npx serve -l 8080
```

**3. Pagination for Large Datasets**
Always paginate to avoid loading too much data:
```json
{
  "type": "crud",
  "api": "http://localhost:3000/api/users",
  "perPage": 20,
  "perPageAvailable": [10, 20, 50]
}
```

### Storage Optimization

**1. Keep JSON Schemas in Separate Files**

Instead of embedding large schemas in HTML, load them:

```html
<script>
  (function() {
    const amis = amisRequire('amis/embed');
    
    // Load schema from external file
    fetch('../schemas/user-crud.json')
      .then(res => res.json())
      .then(schema => {
        amis.embed('#root', schema);
      });
  })();
</script>
```

Create `schemas/user-crud.json`:
```json
{
  "type": "page",
  "title": "User Management",
  "body": {
    ...
  }
}
```

**See the [External JSON Files](#external-json-files) section for complete documentation on this approach.**

**2. Compress SDK Files** (Optional)

```bash
# Install terser for JS minification
npm install -g terser

# Minify SDK (only if you're tight on space)
terser sdk/sdk.js -o sdk/sdk.min.js -c -m
# Update your HTML to use sdk.min.js
```

### Development Workflow

**1. Use tmux for Multi-tasking**
```bash
# Install tmux
pkg install tmux

# Start tmux session
tmux new -s dev

# Split screen: Ctrl+B then "
# Switch panes: Ctrl+B then arrow keys
# In one pane: run your backend server
# In other pane: run http-server
```

**2. Auto-reload Setup**

Create a simple bash script `dev.sh`:
```bash
#!/bin/bash
# Kill existing servers
pkill -f http-server
pkill -f "node.*server"

# Start backend (adjust to your backend)
node server.js &

# Start frontend
http-server -p 8080 &

echo "Servers started!"
echo "Frontend: http://localhost:8080"
echo "Backend: http://localhost:3000"
```

**3. Git Integration**
```bash
# Initialize repo
git init
git add .
git commit -m "Initial commit"

# Ignore node_modules
echo "node_modules/" >> .gitignore
echo "sdk/" >> .gitignore  # Regenerate from npm

# Push to GitHub for backup
git remote add origin <your-repo-url>
git push -u origin main
```

### Performance Tips

**1. Lazy Load Charts**
Charts can be heavy. Load them only when needed:
```json
{
  "type": "service",
  "api": "http://localhost:3000/api/chart-data",
  "initFetch": false,
  "body": {
    "type": "chart",
    "config": {...}
  }
}
```

**2. Debounce Search Inputs**
```json
{
  "type": "input-text",
  "name": "search",
  "label": "Search",
  "clearable": true,
  "searchable": true
}
```
AMIS automatically debounces search inputs.

**3. Virtual Scrolling for Long Lists**
Not directly supported, but use pagination instead:
```json
{
  "type": "crud",
  "loadType": "pagination",
  "perPage": 50
}
```

---

## Troubleshooting

### Common Issues

#### Issue 1: AMIS Not Rendering

**Symptoms:** Blank page, no errors

**Solutions:**
1. Check browser console for errors (F12)
2. Verify SDK files are loaded:
   ```html
   <!-- Check paths are correct -->
   <link rel="stylesheet" href="../sdk/sdk.css">
   <script src="../sdk/sdk.js"></script>
   ```
3. Ensure `#root` div exists:
   ```html
   <div id="root"></div>
   ```
4. Validate JSON syntax:
   ```bash
   # Use a JSON validator
   echo '{"type":"page"}' | python -m json.tool
   ```

#### Issue 2: API Calls Failing

**Symptoms:** Data not loading, form submissions failing

**Solutions:**
1. Check CORS on your backend:
   ```javascript
   // Express.js example
   app.use((req, res, next) => {
     res.header('Access-Control-Allow-Origin', '*');
     res.header('Access-Control-Allow-Methods', 'GET,PUT,POST,DELETE');
     res.header('Access-Control-Allow-Headers', 'Content-Type');
     next();
   });
   ```
2. Verify API response format:
   ```json
   {
     "status": 0,
     "msg": "Success",
     "data": {...}
   }
   ```
3. Check browser network tab (F12 → Network)
4. Test API directly with curl:
   ```bash
   curl http://localhost:3000/api/users
   ```

#### Issue 3: Form Not Submitting

**Symptoms:** Submit button does nothing

**Solutions:**
1. Check `api` is configured:
   ```json
   {
     "type": "form",
     "api": "post:http://localhost:3000/api/save"
   }
   ```
2. Verify all required fields are filled
3. Check browser console for validation errors
4. Ensure backend accepts the data format

#### Issue 4: Styles Not Loading

**Symptoms:** Unstyled components, broken layout

**Solutions:**
1. Verify CSS files exist:
   ```bash
   ls -la sdk/sdk.css
   ```
2. Check CSS is loaded in HTML:
   ```html
   <link rel="stylesheet" href="../sdk/sdk.css">
   <link rel="stylesheet" href="../sdk/iconfont.css">
   ```
3. Clear browser cache (Ctrl+Shift+R)
4. Check for CSS path errors in browser console

#### Issue 5: Termux Server Won't Start

**Symptoms:** "Port already in use" error

**Solutions:**
```bash
# Find process using port 8080
lsof -ti:8080

# Kill the process
kill -9 $(lsof -ti:8080)

# Or use a different port
http-server -p 8081
```

#### Issue 6: JSON Schema Too Complex

**Symptoms:** Page loads slowly, browser freezes

**Solutions:**
1. Break schema into smaller components
2. Use lazy loading with `initFetch: false`
3. Implement pagination
4. Load schemas from external files

### Debugging Tips

**1. Enable Debug Mode**
```html
<script>
  (function() {
    const amis = amisRequire('amis/embed');
    
    const schema = {...};
    
    // Add debug theme
    amis.embed('#root', schema, {}, {
      theme: 'cxd'  // or 'antd', 'dark'
    });
  })();
</script>
```

**2. Inspect Data Flow**

Add this to see what data AMIS is working with:
```json
{
  "type": "tpl",
  "tpl": "<pre>${JSON.stringify(this, null, 2)}</pre>"
}
```

**3. Use Browser DevTools**
- Console: See errors and logs
- Network: Check API requests/responses
- Elements: Inspect rendered HTML

**4. Validate JSON Schema**

Create `validate.js`:
```javascript
const fs = require('fs');
const schema = fs.readFileSync('schemas/my-schema.json', 'utf8');

try {
  JSON.parse(schema);
  console.log('✓ Valid JSON');
} catch (e) {
  console.error('✗ Invalid JSON:', e.message);
}
```

Run: `node validate.js`

---

## Reference

### Essential Component Types

| Component | Purpose | Common Props |
|-----------|---------|--------------|
| `page` | Root container | `title`, `body` |
| `form` | Data input | `api`, `body`, `submitText` |
| `crud` | Full CRUD interface | `api`, `columns`, `filter` |
| `table` | Data table | `columns`, `data` |
| `service` | Data fetching | `api`, `body` |
| `dialog` | Modal popup | `title`, `body`, `size` |
| `drawer` | Side panel | `title`, `body`, `position` |
| `wizard` | Multi-step form | `steps`, `api` |
| `chart` | Data visualization | `config`, `api` |
| `card` | Card layout | `header`, `body`, `actions` |
| `grid` | Grid layout | `columns` |
| `panel` | Panel with header | `title`, `body` |
| `tabs` | Tabbed interface | `tabs` |

### Form Field Types

| Type | Purpose |
|------|---------|
| `input-text` | Single-line text |
| `input-email` | Email with validation |
| `input-password` | Password field |
| `input-number` | Numeric input |
| `input-url` | URL with validation |
| `textarea` | Multi-line text |
| `select` | Dropdown select |
| `checkboxes` | Multiple checkboxes |
| `radios` | Radio buttons |
| `switch` | Toggle switch |
| `date` | Date picker |
| `datetime` | Date & time picker |
| `time` | Time picker |
| `input-file` | File upload |
| `input-image` | Image upload |
| `input-rich-text` | WYSIWYG editor |
| `combo` | Composite field |
| `input-tree` | Tree selector |

### Action Types

| actionType | Purpose |
|------------|---------|
| `ajax` | Make API call |
| `dialog` | Open dialog |
| `drawer` | Open drawer |
| `url` | Navigate to URL |
| `link` | Internal link |
| `reload` | Reload component |
| `setValue` | Set form values |
| `toast` | Show notification |
| `copy` | Copy to clipboard |

### Data Expressions

| Expression | Purpose |
|------------|---------|
| `${field}` | Access field value |
| `${field.nested}` | Nested field |
| `${field[0]}` | Array element |
| `${field \| filter}` | Apply filter |

### Common Filters

| Filter | Purpose | Example |
|--------|---------|---------|
| `date` | Format date | `${created_at \| date:YYYY-MM-DD}` |
| `number` | Format number | `${price \| number}` |
| `truncate` | Limit length | `${text \| truncate:100}` |
| `url_encode` | Encode URL | `${query \| url_encode}` |
| `json` | JSON stringify | `${data \| json}` |

### Quick Reference: File Structure

```
~/amis/
├── sdk/
│   ├── sdk.css         # AMIS styles
│   ├── sdk.js          # AMIS runtime
│   └── iconfont.css    # Icon fonts
├── pages/
│   ├── index.html      # Landing page
│   ├── users.html      # User management
│   └── dashboard.html  # Dashboard
├── schemas/
│   ├── user-form.json
│   ├── user-table.json
│   └── dashboard.json
└── data/
    └── mock.json       # Test data
```

### Quick Reference: HTML Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AMIS App</title>
  <link rel="stylesheet" href="../sdk/sdk.css">
  <link rel="stylesheet" href="../sdk/iconfont.css">
  <style>
    html, body, #root {
      height: 100%;
      margin: 0;
      padding: 0;
    }
  </style>
</head>
<body>
  <div id="root"></div>
  <script src="../sdk/sdk.js"></script>
  <script>
    (function() {
      const amis = amisRequire('amis/embed');
      const schema = {
        type: 'page',
        title: 'Page Title',
        body: 'Content here'
      };
      amis.embed('#root', schema);
    })();
  </script>
</body>
</html>
```

### Resources

**Official Documentation:**
- GitHub: https://github.com/baidu/amis
- Docs (Chinese): https://baidu.github.io/amis/
- Examples: https://baidu.github.io/amis/examples/

**Helpful Tools:**
- JSON Validator: https://jsonlint.com/
- AMIS Editor: https://aisuda.github.io/amis-editor-demo/
- ECharts (for charts): https://echarts.apache.org/

**Termux Resources:**
- Termux Wiki: https://wiki.termux.com/
- Package Search: https://packages.termux.dev/

---

## Summary

You've learned how to:

1. ✅ Set up AMIS in Termux without build tools
2. ✅ Create pages using pure JSON configuration
3. ✅ Build forms, tables, and CRUD interfaces
4. ✅ Connect to your backend APIs
5. ✅ Implement common UI patterns
6. ✅ Optimize for Termux's limitations
7. ✅ Debug and troubleshoot issues

### Next Steps

1. **Start Small**: Build a simple CRUD interface for one of your database tables
2. **Expand Gradually**: Add filters, pagination, and charts
3. **Explore Components**: Try different AMIS components from the reference
4. **Integrate Backend**: Connect to your actual API endpoints
5. **Customize**: Adjust styling and behavior to match your needs

### Remember

- AMIS handles all the React/JavaScript complexity
- You only write JSON configurations
- Everything is data-driven
- Focus on your backend APIs returning the right format
- Use the browser console for debugging

Happy building! 🚀
