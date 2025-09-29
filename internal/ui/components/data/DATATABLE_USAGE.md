# 📊 **ERP DataTable Standardization - Complete Usage Guide**

## 🎯 **Overview**

The standardized DataTable system provides a consistent, powerful, and exciting themed data management experience across all three ERP services (Console, Workspace, Portal) while maintaining proper multi-tenant isolation and ABAC permission enforcement.

## 🚀 **Quick Start**

### Basic Usage

```go
// In your handler
templ UserManagementPage(ctx context.Context, users []UserTableRow) {
    <div class="p-6">
        <h1 class="text-2xl font-bold mb-6">User Management</h1>
        @data.ConsoleUserManagementTable(ctx, users)
    </div>
}
```

### Service-Specific DataTables

```go
// Console (Admin) DataTable - Full permissions and multi-tenant access
@data.ConsoleDataTable(ctx, props)

// Workspace (Business) DataTable - Tenant-scoped with business operations
@data.WorkspaceDataTable(ctx, props) 

// Portal (Customer) DataTable - Read-heavy with limited actions
@data.PortalDataTable(ctx, props)
```

## 🎨 **Exciting Theme Integration**

The DataTable automatically applies service-specific themes:

- **Console**: Blue gradient theme with admin-focused styling
- **Workspace**: Green gradient theme with business operation focus
- **Portal**: Purple gradient theme with customer-friendly appearance

```css
/* Automatically applied classes */
.erp-datatable          /* Modern container with animations */
.erp-datatable-header   /* Gradient header styling */
.erp-datatable-row      /* Hover effects and smooth transitions */
.erp-btn-primary        /* Service-specific button theming */
```

## ⚡ **Enhanced Bulk Actions**

### Standard Bulk Action Patterns

#### Console (Administrative)
```go
BulkActionProps{
    ID:          "bulk_suspend",
    Text:        "Suspend Selected",
    Icon:        "fas fa-user-lock",
    Variant:     "warning",
    Permission:  "user:suspend",        // ABAC permission check
    Confirm:     true,                  // Show confirmation dialog
    ConfirmText: "Are you sure you want to suspend the selected users?",
    Endpoint:    "/console/api/bulk/users/suspend",
    Method:      "POST",
}
```

#### Workspace (Business Operations)
```go
BulkActionProps{
    ID:       "bulk_approve",
    Text:     "Approve Selected",
    Icon:     "fas fa-check",
    Variant:  "primary",
    Permission: "transaction:approve",
    Confirm:  true,
    Endpoint: "/workspace/api/bulk/transactions/approve",
    Method:   "POST",
}
```

#### Portal (Customer Actions)
```go
BulkActionProps{
    ID:       "bulk_download_invoices",
    Text:     "Download Invoices",
    Icon:     "fas fa-download", 
    Variant:  "primary",
    Endpoint: "/portal/api/bulk/orders/download-invoices",
    Method:   "POST",
}
```

## 🔒 **Multi-Tenant Data Isolation**

### Automatic Tenant Context
```go
ServiceContextProps{
    Service:     "workspace",           // Service identification
    TenantID:    getCurrentTenantID(),  // Tenant isolation
    UserID:      getCurrentUserID(),    // User context
    Permissions: getUserPermissions(),  // ABAC permissions
}
```

### Permission-Based Action Visibility
```go
// Actions automatically hidden based on user permissions
if hasPermissionForBulkAction(ctx, action.Permission, serviceContext) {
    @BulkActionButton(action)
}
```

## 📋 **Complete DataTable Configuration**

### Full Feature DataTable
```go
DataTableProps{
    // Core Configuration
    ID:         "my-data-table",
    Selectable: true,                   // Enable row selection
    Searchable: true,                   // Enable search functionality
    Sortable:   true,                   // Enable column sorting
    Pagination: true,                   // Enable pagination
    PageSize:   25,                     // Items per page
    
    // Columns Definition
    Columns: []DataTableColumn{
        {
            Key:      "name",
            Title:    "Full Name",
            Sortable: true,
            Type:     "text",           // text, number, date, currency, badge, action
            Align:    "left",           // left, center, right
            Width:    "300px",          // Optional column width
        },
        {
            Key:      "amount",
            Title:    "Amount",
            Sortable: true,
            Type:     "currency",
            Align:    "right",
            Format:   "USD",            // Currency format
        },
        {
            Key:      "status",
            Title:    "Status",
            Sortable: true,
            Type:     "badge",
            Align:    "center",
        },
    },
    
    // Data
    Rows: convertToDataTableRows(myData),
    
    // Bulk Actions
    BulkActions: []BulkActionProps{
        // Custom bulk actions for this specific table
    },
    
    // Enhanced Features
    ServiceContext:  getServiceContext(),
    PermissionCheck: true,              // Enable ABAC permission checking
    AuditLog:        true,              // Enable audit logging
    Theme:           "console",         // console, workspace, portal
    
    // UI Enhancement
    EmptyMessage: "No records found",
    AriaLabel:    "Data management table",
    Class:        "custom-table-class", // Additional CSS classes
}
```

## 🔧 **Advanced Usage Patterns**

### Custom Column Types
```go
// Badge column with custom styling
DataTableColumn{
    Key:   "priority",
    Title: "Priority", 
    Type:  "badge",
    // Badge content automatically styled based on value
}

// Currency column with automatic formatting
DataTableColumn{
    Key:    "total",
    Title:  "Total Amount",
    Type:   "currency",
    Format: "USD",      // Supports multiple currencies
    Align:  "right",
}

// Date column with consistent formatting
DataTableColumn{
    Key:    "created_at",
    Title:  "Created Date",
    Type:   "date",
    Format: "MM/DD/YYYY", // Custom date format
}
```

### Row-Level Actions
```go
DataTableRow{
    ID: record.ID.String(),
    Data: map[string]interface{}{
        "name":   record.Name,
        "email":  record.Email,
        "status": record.Status,
    },
    Actions: []DataTableAction{
        {
            Text:    "View Details",
            Icon:    "fas fa-eye",
            Variant: "primary",
            Href:    "/console/users/" + record.ID.String(),
        },
        {
            Text:    "Edit",
            Icon:    "fas fa-edit", 
            Variant: "secondary",
            Href:    "/console/users/" + record.ID.String() + "/edit",
        },
        {
            Text:    "Delete",
            Icon:    "fas fa-trash",
            Variant: "danger",
            Action:  "delete_user",  // Custom action handling
        },
    },
}
```

## 🎛️ **Backend Integration**

### Bulk Action Handler
```go
// Example bulk action handler
func (h *ConsoleHandler) HandleBulkUserAction(ctx context.Context, req BulkActionRequest) (*BulkActionResponse, error) {
    // 1. Validate permissions for each selected user
    for _, userID := range req.SelectedIDs {
        if !h.abac.HasPermission(ctx, "user", req.ActionID, &userID) {
            return nil, errors.New("insufficient permissions")
        }
    }
    
    // 2. Execute bulk action with proper tenant isolation
    switch req.ActionID {
    case "bulk_suspend":
        return h.bulkSuspendUsers(ctx, req.SelectedIDs)
    case "bulk_activate":
        return h.bulkActivateUsers(ctx, req.SelectedIDs) 
    case "bulk_delete":
        return h.bulkDeleteUsers(ctx, req.SelectedIDs)
    }
    
    return nil, errors.New("unknown bulk action")
}
```

### HTMX Endpoint Integration
```go
// HTMX-compatible bulk action endpoint
func (h *ConsoleHandler) BulkUserAction(c *fiber.Ctx) error {
    var req struct {
        SelectedIDs []string `json:"selected_ids"`
        ActionID    string   `json:"action_id"`
    }
    
    if err := c.BodyParser(&req); err != nil {
        return err
    }
    
    // Execute bulk action
    result, err := h.HandleBulkUserAction(c.Context(), BulkActionRequest{
        SelectedIDs: req.SelectedIDs,
        ActionID:    req.ActionID,
    })
    
    if err != nil {
        // Return error response for HTMX
        c.Set("HX-Trigger", `{"showNotification": {"type": "error", "message": "` + err.Error() + `"}}`)
        return c.Status(400).SendString("Error processing bulk action")
    }
    
    // Return success response
    c.Set("HX-Trigger", `{"showNotification": {"type": "success", "message": "Bulk action completed successfully"}}`)
    return c.Render("console/users/list", updatedUserList)
}
```

## 🎯 **Migration from Existing Tables**

### Before (Old Table Pattern)
```go
// Old manual table implementation
<table class="min-w-full divide-y divide-gray-200">
    <thead class="bg-gray-50">
        <tr>
            <th>User</th>
            <th>Status</th>
            <th>Actions</th>
        </tr>
    </thead>
    <tbody>
        for _, user := range users {
            <tr>
                <td>{user.Name}</td>
                <td>{user.Status}</td>
                <td>
                    <a href="/edit">Edit</a>
                </td>
            </tr>
        }
    </tbody>
</table>
```

### After (Standardized DataTable)
```go
// New standardized DataTable
@data.ConsoleDataTable(ctx, data.DataTableProps{
    Columns: []data.DataTableColumn{
        {Key: "name", Title: "User", Sortable: true, Type: "text"},
        {Key: "status", Title: "Status", Sortable: true, Type: "badge"},
    },
    Rows: convertUsersToDataTableRows(users),
    BulkActions: data.getStandardConsoleBulkActions(),
})
```

## ✨ **Benefits of Standardized DataTable**

### 🎨 **Visual Consistency**
- ✅ Exciting theme across all services
- ✅ Consistent hover effects and animations
- ✅ Service-specific color schemes
- ✅ Modern gradient styling

### 🚀 **Enhanced Functionality**
- ✅ Advanced search and filtering
- ✅ Intelligent pagination
- ✅ Multi-column sorting
- ✅ Bulk action processing
- ✅ Row-level actions

### 🔒 **Security & Permissions**
- ✅ ABAC permission integration
- ✅ Multi-tenant data isolation
- ✅ Service-specific context
- ✅ Action-level permission checks

### 🛠️ **Developer Experience**
- ✅ Reusable components
- ✅ Type-safe configuration
- ✅ Consistent patterns
- ✅ Easy customization

## 🚀 **Next Steps**

1. **Replace existing tables** with standardized DataTable components
2. **Configure bulk actions** for your specific use cases  
3. **Integrate with ABAC** permission system
4. **Test multi-tenant isolation** in your environment
5. **Customize themes** for your brand requirements

The standardized DataTable system provides a powerful foundation for all data management operations across your ERP system while maintaining the exciting visual theme and ensuring proper security isolation!