# User Select Control Component

**FILE PURPOSE**: Specialized user picker control for employee and user selection  
**SCOPE**: User selection, employee assignment, team member picking, and user management interfaces  
**TARGET AUDIENCE**: Developers implementing user assignment, team management, and personnel selection features

## 📋 Component Overview

User Select Control provides specialized functionality for selecting users, employees, and team members with support for search, filtering, multi-selection, and integration with user management systems. Optimized for ERP user assignment scenarios.

### Schema Reference
- **Primary Schema**: `UserSelectControlSchema.json`
- **Related Schemas**: `Option.json`, `BaseApiObject.json`
- **Base Interface**: Form input control for user/employee selection

## Basic Usage

```json
{
    "type": "users-select",
    "name": "assigned_users",
    "label": "Assigned Users",
    "placeholder": "Select users...",
    "source": "/api/users"
}
```

## Go Type Definition

```go
type UserSelectControlProps struct {
    Type            string              `json:"type"`
    Name            string              `json:"name"`
    Label           interface{}         `json:"label"`
    Placeholder     string              `json:"placeholder"`
    Options         interface{}         `json:"options"`
    Source          interface{}         `json:"source"`
    Multiple        bool                `json:"multiple"`
    Clearable       bool                `json:"clearable"`
    SelectFirst     bool                `json:"selectFirst"`
    InitFetch       bool                `json:"initFetch"`
    InitFetchOn     string              `json:"initFetchOn"`
    JoinValues      bool                `json:"joinValues"`
    Delimiter       string              `json:"delimiter"`
    ExtractValue    bool                `json:"extractValue"`
    ValuesNoWrap    bool                `json:"valuesNoWrap"`
    Creatable       bool                `json:"creatable"`
    CreateBtnLabel  string              `json:"createBtnLabel"`
    Editable        bool                `json:"editable"`
    Removable       bool                `json:"removable"`
    DeferField      string              `json:"deferField"`
    DeferApi        interface{}         `json:"deferApi"`
    AddApi          interface{}         `json:"addApi"`
    EditApi         interface{}         `json:"editApi"`
    DeleteApi       interface{}         `json:"deleteApi"`
}
```

## Essential Variants

### Basic User Selection
```json
{
    "type": "users-select",
    "name": "project_manager",
    "label": "Project Manager",
    "placeholder": "Select project manager...",
    "source": "/api/users/managers",
    "clearable": true,
    "required": true
}
```

### Multi-User Selection
```json
{
    "type": "users-select",
    "name": "team_members",
    "label": "Team Members",
    "placeholder": "Select team members...",
    "source": "/api/users/active",
    "multiple": true,
    "clearable": true,
    "extractValue": true,
    "joinValues": false
}
```

### Department-Filtered Users
```json
{
    "type": "users-select",
    "name": "department_users",
    "label": "Department Staff",
    "placeholder": "Select department staff...",
    "source": "/api/users?department=${department_id}",
    "multiple": true,
    "initFetchOn": "${department_id}",
    "clearable": true
}
```

### Role-Based User Selection
```json
{
    "type": "users-select",
    "name": "reviewers",
    "label": "Reviewers",
    "placeholder": "Select reviewers...",
    "source": "/api/users/by-role?role=reviewer",
    "multiple": true,
    "clearable": true,
    "valuesNoWrap": false,
    "extractValue": true
}
```

### Static User List
```json
{
    "type": "users-select",
    "name": "supervisor",
    "label": "Supervisor",
    "placeholder": "Select supervisor...",
    "options": [
        {"label": "John Smith (Manager)", "value": "user_123"},
        {"label": "Sarah Johnson (Senior Manager)", "value": "user_456"},
        {"label": "Mike Davis (Director)", "value": "user_789"}
    ],
    "clearable": true
}
```

## Real-World Use Cases

### Project Assignment
```json
{
    "type": "users-select",
    "name": "project_assignees",
    "label": "Project Assignees",
    "placeholder": "Assign project to users...",
    "source": "/api/users/available-for-projects",
    "multiple": true,
    "clearable": true,
    "extractValue": true,
    "required": true,
    "validations": {
        "minimum": 1
    },
    "validationErrors": {
        "minimum": "At least one user must be assigned to the project"
    },
    "hint": "Select users who will work on this project"
}
```

### Task Assignment
```json
{
    "type": "users-select",
    "name": "task_assignee",
    "label": "Assigned To",
    "placeholder": "Assign task to user...",
    "source": "/api/users/by-skill?skills=${required_skills}",
    "clearable": true,
    "initFetchOn": "${required_skills}",
    "autoFill": {
        "api": "/api/users/${value}/details",
        "fillMapping": {
            "assignee_email": "email",
            "assignee_phone": "phone",
            "assignee_department": "department.name"
        }
    }
}
```

### Approval Workflow
```json
{
    "type": "users-select",
    "name": "approval_chain",
    "label": "Approval Chain",
    "placeholder": "Select approvers in order...",
    "source": "/api/users/approvers",
    "multiple": true,
    "joinValues": false,
    "extractValue": true,
    "clearable": true,
    "required": true,
    "hint": "Select users who will approve this request in order"
}
```

### Meeting Attendees
```json
{
    "type": "users-select",
    "name": "meeting_attendees",
    "label": "Meeting Attendees",
    "placeholder": "Add meeting attendees...",
    "source": "/api/users/active",
    "multiple": true,
    "clearable": true,
    "extractValue": true,
    "valuesNoWrap": true,
    "creatable": false,
    "hint": "Select users to invite to this meeting"
}
```

### Document Reviewers
```json
{
    "type": "users-select",
    "name": "document_reviewers",
    "label": "Document Reviewers",
    "placeholder": "Select reviewers...",
    "source": "/api/users/with-review-permissions",
    "multiple": true,
    "clearable": true,
    "extractValue": true,
    "autoFill": {
        "api": "/api/users/bulk-details",
        "fillMapping": {
            "reviewer_emails": "emails",
            "notification_preferences": "notification_settings"
        }
    }
}
```

### Support Ticket Assignment
```json
{
    "type": "users-select",
    "name": "support_agent",
    "label": "Assign to Support Agent",
    "placeholder": "Select support agent...",
    "source": "/api/users/support-agents?available=true",
    "clearable": true,
    "selectFirst": false,
    "autoFill": {
        "api": "/api/users/${value}/workload",
        "fillMapping": {
            "agent_workload": "current_tickets",
            "agent_expertise": "specializations"
        }
    }
}
```

### Access Control Assignment
```json
{
    "type": "users-select",
    "name": "resource_access_users",
    "label": "Grant Access To",
    "placeholder": "Select users to grant access...",
    "source": "/api/users/eligible-for-resource?resource_id=${resource_id}",
    "multiple": true,
    "initFetchOn": "${resource_id}",
    "clearable": true,
    "extractValue": true,
    "validations": {
        "maximum": 10
    },
    "validationErrors": {
        "maximum": "Cannot grant access to more than 10 users at once"
    }
}
```

### Team Formation
```json
{
    "type": "users-select",
    "name": "team_leads",
    "label": "Team Leads",
    "placeholder": "Select team leads...",
    "source": "/api/users/potential-leads",
    "multiple": true,
    "clearable": true,
    "extractValue": true,
    "editable": true,
    "editApi": "/api/users/update-role",
    "editControls": [
        {
            "type": "select",
            "name": "leadership_level",
            "label": "Leadership Level",
            "options": ["Junior Lead", "Senior Lead", "Principal Lead"]
        }
    ]
}
```

This component provides essential user selection functionality for ERP systems requiring personnel assignment, team management, and user-based workflows.