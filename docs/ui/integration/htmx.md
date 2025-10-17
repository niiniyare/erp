# HTMX Integration Patterns

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Server interaction patterns and HTMX implementation strategies for ERP interfaces
**SCOPE**: HTMX attributes, server responses, event handling, progressive enhancement
**TARGET AUDIENCE**: Full-stack developers implementing server-driven UI interactions
**RELATED FILES**: `components/forms.md` (form patterns), `fundamentals/architecture.md` (system design)
<!-- LLM-CONTEXT-END -->

## Overview

This guide documents proven patterns for integrating HTMX with Templ templates to create responsive, server-driven ERP interfaces. These patterns enable rich user interactions while maintaining the simplicity of server-side rendering.

<!-- LLM-HTMX-PRINCIPLES-START -->
**HTMX INTEGRATION PRINCIPLES:**
- **Server-First Architecture** - Business logic remains on the server
- **Progressive Enhancement** - Core functionality works without JavaScript
- **Declarative Interactions** - UI behavior defined in HTML attributes
- **Efficient Updates** - Only update the parts of the page that change
- **SEO Friendly** - Server-rendered content for search engines
<!-- LLM-HTMX-PRINCIPLES-END -->

---

## Core HTMX Patterns

### Request Patterns

<!-- LLM-REQUEST-PATTERNS-START -->
**FUNDAMENTAL REQUEST PATTERNS:**

#### 1. Basic Request Pattern
```html
<!-- Simple GET request -->
<button 
    hx-get="/api/users/123"
    hx-target="#user-details"
    hx-swap="innerHTML">
    Load User Details
</button>

<!-- POST with form data -->
<form 
    hx-post="/api/users"
    hx-target="#user-list"
    hx-swap="afterbegin">
    <input name="name" required>
    <button type="submit">Create User</button>
</form>
```

#### 2. Conditional Requests
```html
<!-- Request with conditions -->
<button 
    hx-get="/api/reports/generate"
    hx-target="#report-container"
    hx-trigger="click"
    hx-confirm="This will take several minutes. Continue?"
    hx-indicator="#spinner">
    Generate Report
</button>

<!-- Authenticated requests -->
<div 
    hx-get="/api/sensitive-data"
    hx-target="this"
    hx-headers='{"Authorization": "Bearer ${token}"}'>
    Loading secure content...
</div>
```

#### 3. Parameterized Requests
```html
<!-- Include additional form data -->
<select 
    name="department"
    hx-get="/api/users/filter"
    hx-target="#user-list"
    hx-include="[name='role'], [name='status']"
    hx-trigger="change">
    <option value="">All Departments</option>
    <option value="engineering">Engineering</option>
    <option value="sales">Sales</option>
</select>

<!-- Dynamic URL construction -->
<button 
    hx-get="/api/users/${userId}/permissions"
    hx-target="#permissions-panel"
    hx-vals='{"include_inherited": true}'>
    Load Permissions
</button>
```
<!-- LLM-REQUEST-PATTERNS-END -->

### Response Handling Patterns

<!-- LLM-RESPONSE-PATTERNS-START -->
**SERVER RESPONSE PATTERNS:**

#### 1. Content Replacement
```go
// Go handler returning partial HTML
func (h *Handler) GetUserList(w http.ResponseWriter, r *http.Request) {
    users := h.userService.GetUsers(r.Context())
    
    // Return just the table rows
    tmpl := templates.UserTableRows(users)
    tmpl.Render(r.Context(), w)
}
```

```html
<!-- Different swap strategies -->
<div 
    hx-get="/api/users"
    hx-target="#user-table tbody"
    hx-swap="innerHTML">          <!-- Replace content -->
    
<div 
    hx-post="/api/users"
    hx-target="#user-list"
    hx-swap="afterbegin">         <!-- Prepend new item -->

<div 
    hx-delete="/api/users/123"
    hx-target="closest tr"
    hx-swap="outerHTML">          <!-- Remove entire row -->
```

#### 2. Multiple Target Updates
```go
// Server response with multiple triggers
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.userService.CreateUser(r.Context(), form)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Trigger multiple UI updates
    w.Header().Set("HX-Trigger", `{
        "userCreated": {"id": `+strconv.Itoa(user.ID)+`},
        "updateCounter": {"count": `+strconv.Itoa(h.userService.GetCount())+`},
        "showNotification": {
            "type": "success",
            "message": "User created successfully"
        }
    }`)
    
    // Return the new table row
    tmpl := templates.UserTableRow(user)
    tmpl.Render(r.Context(), w)
}
```

#### 3. Error Handling
```go
// Validation error response
func (h *Handler) ValidateForm(w http.ResponseWriter, r *http.Request) {
    if !validator.IsValid() {
        // Return form with errors
        w.Header().Set("HX-Reswap", "outerHTML")
        w.WriteHeader(http.StatusUnprocessableEntity)
        
        tmpl := templates.FormWithErrors(form, validator.Errors())
        tmpl.Render(r.Context(), w)
        return
    }
    
    // Success response
    w.Header().Set("HX-Trigger", "formValid")
    w.WriteHeader(http.StatusOK)
}
```
<!-- LLM-RESPONSE-PATTERNS-END -->

---

## UI Interaction Patterns

### Search and Filtering

<!-- LLM-SEARCH-PATTERNS-START -->
**SEARCH AND FILTER PATTERNS:**

#### 1. Real-time Search
```go
// Templ template with debounced search
templ SearchableUserList() {
    <div class="search-container">
        <input 
            type="text"
            name="search"
            placeholder="Search users..."
            hx-get="/api/users/search"
            hx-target="#user-results"
            hx-trigger="keyup changed delay:300ms"
            hx-indicator="#search-spinner"
            class="w-full px-3 py-2 border rounded-md">
        
        <div id="search-spinner" class="htmx-indicator">
            <svg class="animate-spin h-5 w-5">...</svg>
        </div>
    </div>
    
    <div id="user-results">
        @UserList(users)
    </div>
}
```

#### 2. Advanced Filtering
```go
templ FilterPanel(entity string) {
    <form class="filter-panel" x-data="{ expanded: false }">
        <!-- Quick Filters -->
        <div class="flex gap-4 mb-4">
            <select 
                name="status"
                hx-get={"/api/"+entity+"/filter"}
                hx-target="#results"
                hx-include="[name^='filter']"
                hx-trigger="change">
                <option value="">All Status</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
            </select>
            
            <button 
                type="button"
                @click="expanded = !expanded"
                class="px-3 py-2 text-sm bg-gray-100">
                <span x-text="expanded ? 'Hide' : 'Show'"></span> Advanced
            </button>
        </div>
        
        <!-- Advanced Filters -->
        <div x-show="expanded" x-transition class="grid grid-cols-3 gap-4">
            <input 
                type="date"
                name="filter[date_from]"
                hx-get={"/api/"+entity+"/filter"}
                hx-target="#results"
                hx-include="[name^='filter'], [name='status']"
                hx-trigger="change">
            
            <input 
                type="date"
                name="filter[date_to]"
                hx-get={"/api/"+entity+"/filter"}
                hx-target="#results"
                hx-include="[name^='filter'], [name='status']"
                hx-trigger="change">
                
            <select 
                name="filter[department]"
                hx-get={"/api/"+entity+"/filter"}
                hx-target="#results"
                hx-include="[name^='filter'], [name='status']"
                hx-trigger="change">
                <option value="">All Departments</option>
                <option value="engineering">Engineering</option>
                <option value="sales">Sales</option>
            </select>
        </div>
    </form>
}
```

#### 3. Server-side Filter Handler
```go
func (h *Handler) FilterUsers(w http.ResponseWriter, r *http.Request) {
    // Parse filter parameters
    filters := FilterParams{
        Search:     r.FormValue("search"),
        Status:     r.FormValue("status"),
        DateFrom:   r.FormValue("filter[date_from]"),
        DateTo:     r.FormValue("filter[date_to]"),
        Department: r.FormValue("filter[department]"),
    }
    
    // Apply filters
    users, total := h.userService.FilterUsers(r.Context(), filters)
    
    // Return filtered results
    w.Header().Set("HX-Trigger", fmt.Sprintf(`{"updateCount": {"total": %d}}`, total))
    tmpl := templates.UserTable(users)
    tmpl.Render(r.Context(), w)
}
```
<!-- LLM-SEARCH-PATTERNS-END -->

### Form Interactions

<!-- LLM-FORM-PATTERNS-START -->
**FORM INTERACTION PATTERNS:**

#### 1. Modal Forms
```go
templ CreateUserModal() {
    <div class="modal-backdrop" x-data="{ open: true }" x-show="open">
        <div class="modal-content" @click.away="open = false">
            <form 
                hx-post="/api/users"
                hx-target="#user-list"
                hx-swap="afterbegin"
                hx-on::after-request="if(event.detail.successful) $el.closest('.modal-backdrop').remove()">
                
                <h2 class="text-xl font-semibold mb-4">Create New User</h2>
                
                @FormField("name", "Full Name", "", true)
                @FormField("email", "Email Address", "", true)
                @SelectField("role", "Role", roleOptions, "")
                
                <div class="flex justify-end gap-3 mt-6">
                    <button 
                        type="button"
                        @click="open = false"
                        class="px-4 py-2 bg-gray-200 rounded">
                        Cancel
                    </button>
                    <button 
                        type="submit"
                        class="px-4 py-2 bg-blue-600 text-white rounded">
                        Create User
                    </button>
                </div>
            </form>
        </div>
    </div>
}
```

#### 2. Inline Editing
```go
templ EditableField(fieldName, value string, editable bool) {
    <div class="editable-field" x-data="{editing: false, value: '"+value+"'}">
        <div x-show="!editing" class="flex items-center gap-2">
            <span x-text="value"></span>
            if editable {
                <button 
                    @click="editing = true"
                    class="text-blue-600 hover:text-blue-800">
                    <svg class="w-4 h-4">...</svg>
                </button>
            }
        </div>
        
        <div x-show="editing" class="flex items-center gap-2">
            <input 
                type="text"
                x-model="value"
                @keydown.enter="$refs.saveBtn.click()"
                @keydown.escape="editing = false; value = '"+value+"'"
                class="px-2 py-1 border rounded">
            
            <button 
                x-ref="saveBtn"
                hx-put={"/api/users/"+fieldName}
                hx-vals="js:{value: value}"
                hx-on::after-request="editing = false"
                class="px-2 py-1 bg-green-600 text-white rounded">
                Save
            </button>
            
            <button 
                @click="editing = false; value = '"+value+"'"
                class="px-2 py-1 bg-gray-400 text-white rounded">
                Cancel
            </button>
        </div>
    </div>
}
```

#### 3. Multi-step Forms
```go
templ MultiStepForm(currentStep int) {
    <div class="multi-step-form" x-data="{step: "+strconv.Itoa(currentStep)+"}">
        <!-- Progress Indicator -->
        <div class="flex justify-between mb-8">
            <div class="flex items-center">
                <div 
                    :class="step >= 1 ? 'bg-blue-600 text-white' : 'bg-gray-200'"
                    class="w-8 h-8 rounded-full flex items-center justify-center">
                    1
                </div>
                <div class="w-24 h-0.5 bg-gray-200 mx-2"></div>
                <div 
                    :class="step >= 2 ? 'bg-blue-600 text-white' : 'bg-gray-200'"
                    class="w-8 h-8 rounded-full flex items-center justify-center">
                    2
                </div>
                <div class="w-24 h-0.5 bg-gray-200 mx-2"></div>
                <div 
                    :class="step >= 3 ? 'bg-blue-600 text-white' : 'bg-gray-200'"
                    class="w-8 h-8 rounded-full flex items-center justify-center">
                    3
                </div>
            </div>
        </div>
        
        <!-- Step Content -->
        <div id="step-content">
            if currentStep == 1 {
                @PersonalInfoStep()
            } else if currentStep == 2 {
                @ContactInfoStep()
            } else if currentStep == 3 {
                @ReviewStep()
            }
        </div>
        
        <!-- Navigation -->
        <div class="flex justify-between mt-8">
            if currentStep > 1 {
                <button 
                    hx-get={"/api/forms/step/" + strconv.Itoa(currentStep-1)}
                    hx-target="#step-content"
                    hx-swap="innerHTML"
                    @click="step = step - 1"
                    class="px-4 py-2 bg-gray-200 rounded">
                    Previous
                </button>
            }
            
            if currentStep < 3 {
                <button 
                    hx-post={"/api/forms/validate-step/" + strconv.Itoa(currentStep)}
                    hx-target="#step-content"
                    hx-swap="innerHTML"
                    hx-on::after-request="if(event.detail.successful) step = step + 1"
                    class="px-4 py-2 bg-blue-600 text-white rounded">
                    Next
                </button>
            } else {
                <button 
                    hx-post="/api/users/create"
                    hx-target="#form-container"
                    hx-swap="outerHTML"
                    class="px-4 py-2 bg-green-600 text-white rounded">
                    Complete
                </button>
            }
        </div>
    </div>
}
```
<!-- LLM-FORM-PATTERNS-END -->

### Data Table Patterns

<!-- LLM-TABLE-PATTERNS-START -->
**DATA TABLE INTERACTION PATTERNS:**

#### 1. Sortable Tables
```go
templ SortableTable(entity string, data []interface{}, sortField, sortDir string) {
    <table class="w-full">
        <thead>
            <tr>
                @SortableHeader("name", "Name", entity, sortField, sortDir)
                @SortableHeader("email", "Email", entity, sortField, sortDir)
                @SortableHeader("created_at", "Created", entity, sortField, sortDir)
                <th>Actions</th>
            </tr>
        </thead>
        <tbody id="table-body">
            for _, item := range data {
                @TableRow(item)
            }
        </tbody>
    </table>
}

templ SortableHeader(field, label, entity, currentSort, currentDir string) {
    <th class="px-6 py-3 text-left">
        <button 
            hx-get={"/api/"+entity+"?sort="+field+"&dir="+getNextSortDir(field, currentSort, currentDir)}
            hx-target="#table-container"
            hx-include="[name='search'], [name^='filter']"
            class="flex items-center gap-1 hover:text-blue-600">
            
            {label}
            
            if currentSort == field {
                if currentDir == "asc" {
                    <svg class="w-4 h-4">↑</svg>
                } else {
                    <svg class="w-4 h-4">↓</svg>
                }
            } else {
                <svg class="w-4 h-4 text-gray-400">↕</svg>
            }
        </button>
    </th>
}
```

#### 2. Bulk Actions
```go
templ BulkActionToolbar() {
    <div class="bulk-actions" x-data="bulkActions()" x-show="selectedCount > 0">
        <div class="flex items-center gap-4 p-4 bg-blue-50 border border-blue-200 rounded">
            <span x-text="`${selectedCount} items selected`"></span>
            
            <button 
                hx-post="/api/users/bulk/activate"
                hx-vals="js:{ids: getSelectedIds()}"
                hx-target="#table-container"
                hx-confirm="Activate selected users?"
                class="px-3 py-1 bg-green-600 text-white rounded">
                Activate
            </button>
            
            <button 
                hx-post="/api/users/bulk/deactivate"
                hx-vals="js:{ids: getSelectedIds()}"
                hx-target="#table-container"
                hx-confirm="Deactivate selected users?"
                class="px-3 py-1 bg-yellow-600 text-white rounded">
                Deactivate
            </button>
            
            <button 
                hx-delete="/api/users/bulk"
                hx-vals="js:{ids: getSelectedIds()}"
                hx-target="#table-container"
                hx-confirm="Delete selected users? This cannot be undone."
                class="px-3 py-1 bg-red-600 text-white rounded">
                Delete
            </button>
        </div>
    </div>
}
```

#### 3. Pagination
```go
templ Pagination(currentPage, totalPages int, entity string) {
    <div class="pagination flex items-center justify-between">
        <div class="info">
            Showing page {strconv.Itoa(currentPage)} of {strconv.Itoa(totalPages)}
        </div>
        
        <div class="controls flex gap-2">
            if currentPage > 1 {
                <button 
                    hx-get={"/api/"+entity+"?page=1"}
                    hx-target="#table-container"
                    hx-include="[name='search'], [name^='filter'], [name='sort']"
                    class="px-3 py-2 border rounded">
                    First
                </button>
                
                <button 
                    hx-get={"/api/"+entity+"?page="+strconv.Itoa(currentPage-1)}
                    hx-target="#table-container"
                    hx-include="[name='search'], [name^='filter'], [name='sort']"
                    class="px-3 py-2 border rounded">
                    Previous
                </button>
            }
            
            <!-- Page numbers -->
            for i := max(1, currentPage-2); i <= min(totalPages, currentPage+2); i++ {
                <button 
                    hx-get={"/api/"+entity+"?page="+strconv.Itoa(i)}
                    hx-target="#table-container"
                    hx-include="[name='search'], [name^='filter'], [name='sort']"
                    class={getPaginationButtonClass(i, currentPage)}>
                    {strconv.Itoa(i)}
                </button>
            }
            
            if currentPage < totalPages {
                <button 
                    hx-get={"/api/"+entity+"?page="+strconv.Itoa(currentPage+1)}
                    hx-target="#table-container"
                    hx-include="[name='search'], [name^='filter'], [name='sort']"
                    class="px-3 py-2 border rounded">
                    Next
                </button>
                
                <button 
                    hx-get={"/api/"+entity+"?page="+strconv.Itoa(totalPages)}
                    hx-target="#table-container"
                    hx-include="[name='search'], [name^='filter'], [name='sort']"
                    class="px-3 py-2 border rounded">
                    Last
                </button>
            }
        </div>
    </div>
}
```
<!-- LLM-TABLE-PATTERNS-END -->

---

## Advanced HTMX Patterns

### Real-time Updates

<!-- LLM-REALTIME-PATTERNS-START -->
**REAL-TIME INTERACTION PATTERNS:**

#### 1. Server-Sent Events (SSE)
```go
// Go SSE handler
func (h *Handler) EventStream(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // Send initial connection
    fmt.Fprintf(w, "data: connected\n\n")
    w.(http.Flusher).Flush()
    
    // Listen for events
    for {
        select {
        case event := <-h.eventChannel:
            data, _ := json.Marshal(event)
            fmt.Fprintf(w, "data: %s\n\n", data)
            w.(http.Flusher).Flush()
        case <-r.Context().Done():
            return
        }
    }
}
```

```html
<!-- HTML with SSE -->
<div 
    hx-ext="sse"
    sse-connect="/api/events"
    sse-swap="message"
    hx-target="#notifications">
    <div id="notifications"></div>
</div>
```

#### 2. WebSocket Integration
```html
<!-- WebSocket with HTMX -->
<div 
    hx-ext="ws"
    ws-connect="/api/websocket">
    
    <form ws-send>
        <input name="message" placeholder="Type a message...">
        <button type="submit">Send</button>
    </form>
    
    <div id="chat-messages"></div>
</div>
```

#### 3. Polling for Updates
```html
<!-- Automatic polling -->
<div 
    hx-get="/api/dashboard/stats"
    hx-trigger="every 30s"
    hx-target="this"
    hx-swap="innerHTML">
    <div class="stats-grid">
        <!-- Initial stats content -->
    </div>
</div>

<!-- Conditional polling -->
<div 
    hx-get="/api/job/status"
    hx-trigger="every 2s [job-status!='completed']"
    hx-target="#job-progress">
    <div id="job-progress">Processing...</div>
</div>
```
<!-- LLM-REALTIME-PATTERNS-END -->

### Error Handling and Recovery

<!-- LLM-ERROR-PATTERNS-START -->
**ERROR HANDLING PATTERNS:**

#### 1. Graceful Error Responses
```go
func (h *Handler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
    switch {
    case errors.Is(err, ErrNotFound):
        w.Header().Set("HX-Trigger", `{"showNotification": {"type": "error", "message": "Resource not found"}}`)
        w.WriteHeader(http.StatusNotFound)
        
        tmpl := templates.NotFoundMessage()
        tmpl.Render(r.Context(), w)
        
    case errors.Is(err, ErrValidation):
        w.Header().Set("HX-Reswap", "outerHTML")
        w.WriteHeader(http.StatusUnprocessableEntity)
        
        tmpl := templates.ValidationError(err.Error())
        tmpl.Render(r.Context(), w)
        
    case errors.Is(err, ErrUnauthorized):
        w.Header().Set("HX-Redirect", "/login")
        w.WriteHeader(http.StatusUnauthorized)
        
    default:
        w.Header().Set("HX-Trigger", `{"showNotification": {"type": "error", "message": "An unexpected error occurred"}}`)
        w.WriteHeader(http.StatusInternalServerError)
        
        tmpl := templates.ErrorMessage("Something went wrong. Please try again.")
        tmpl.Render(r.Context(), w)
    }
}
```

#### 2. Client-side Error Handling
```javascript
// Global HTMX error handling
document.body.addEventListener('htmx:responseError', function(evt) {
    const status = evt.detail.xhr.status;
    
    switch (status) {
        case 401:
            // Redirect to login
            window.location.href = '/login';
            break;
            
        case 403:
            showNotification('error', 'You do not have permission to perform this action');
            break;
            
        case 404:
            showNotification('error', 'The requested resource was not found');
            break;
            
        case 422:
            // Validation errors are handled by server response
            break;
            
        case 500:
            showNotification('error', 'Server error. Please try again later');
            break;
            
        default:
            showNotification('error', 'An unexpected error occurred');
    }
});

// Network error handling
document.body.addEventListener('htmx:sendError', function(evt) {
    showNotification('error', 'Network error. Please check your connection');
});

// Timeout handling
document.body.addEventListener('htmx:timeout', function(evt) {
    showNotification('warning', 'Request timed out. Please try again');
});
```

#### 3. Retry Mechanisms
```html
<!-- Automatic retry on failure -->
<div 
    hx-get="/api/data"
    hx-trigger="load, retry from:body"
    hx-target="this"
    hx-on::htmx:response-error="
        setTimeout(() => {
            if (this.retryCount < 3) {
                this.retryCount = (this.retryCount || 0) + 1;
                htmx.trigger(document.body, 'retry');
            }
        }, 1000 * Math.pow(2, this.retryCount || 0))
    ">
    Loading...
</div>

<!-- Manual retry button -->
<div id="error-container" style="display: none;">
    <p>Failed to load content</p>
    <button 
        hx-get="/api/data"
        hx-target="#content"
        hx-on::after-request="document.getElementById('error-container').style.display = 'none'">
        Retry
    </button>
</div>
```
<!-- LLM-ERROR-PATTERNS-END -->

---

## Performance Optimization Patterns

### Request Optimization

<!-- LLM-PERFORMANCE-PATTERNS-START -->
**PERFORMANCE OPTIMIZATION STRATEGIES:**

#### 1. Request Debouncing
```html
<!-- Debounced search -->
<input 
    type="text"
    hx-get="/api/search"
    hx-target="#results"
    hx-trigger="keyup changed delay:500ms"
    hx-indicator="#search-spinner">

<!-- Throttled updates -->
<input 
    type="range"
    hx-post="/api/settings/volume"
    hx-trigger="input throttle:100ms"
    hx-vals="js:{volume: this.value}">
```

#### 2. Request Cancellation
```javascript
// Cancel previous request on new input
document.body.addEventListener('htmx:beforeRequest', function(evt) {
    const target = evt.target;
    
    // Cancel previous request if it's still pending
    if (target.pendingRequest) {
        target.pendingRequest.abort();
    }
    
    // Store current request for potential cancellation
    target.pendingRequest = evt.detail.xhr;
});

document.body.addEventListener('htmx:afterRequest', function(evt) {
    // Clear the pending request
    evt.target.pendingRequest = null;
});
```

#### 3. Intelligent Caching
```go
// Server-side caching with ETags
func (h *Handler) GetUserList(w http.ResponseWriter, r *http.Request) {
    // Generate ETag based on data version
    etag := h.userService.GetDataETag()
    w.Header().Set("ETag", etag)
    
    // Check if client has current version
    if r.Header.Get("If-None-Match") == etag {
        w.WriteHeader(http.StatusNotModified)
        return
    }
    
    // Return updated data
    users := h.userService.GetUsers(r.Context())
    tmpl := templates.UserList(users)
    tmpl.Render(r.Context(), w)
}
```

```html
<!-- Client-side cache control -->
<div 
    hx-get="/api/users"
    hx-target="this"
    hx-swap="innerHTML"
    hx-headers='{"Cache-Control": "max-age=300"}'>
    <!-- User list will be cached for 5 minutes -->
</div>
```
<!-- LLM-PERFORMANCE-PATTERNS-END -->

### Response Optimization

<!-- LLM-RESPONSE-OPTIMIZATION-START -->
**RESPONSE OPTIMIZATION TECHNIQUES:**

#### 1. Partial Updates
```go
// Return only changed data
func (h *Handler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    newStatus := r.FormValue("status")
    
    user, err := h.userService.UpdateStatus(userID, newStatus)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Return only the status badge, not the entire row
    tmpl := templates.StatusBadge(user.Status)
    tmpl.Render(r.Context(), w)
}
```

#### 2. Lazy Loading
```html
<!-- Lazy load content when it comes into view -->
<div 
    hx-get="/api/heavy-content"
    hx-trigger="intersect once"
    hx-target="this"
    hx-swap="innerHTML">
    <div class="placeholder">
        <div class="animate-pulse bg-gray-200 h-32"></div>
        <p>Loading content...</p>
    </div>
</div>

<!-- Lazy load images -->
<div 
    hx-get="/api/user/avatar"
    hx-trigger="intersect once"
    hx-target="this"
    hx-vals="js:{userId: this.dataset.userId}"
    data-user-id="123">
    <div class="w-16 h-16 bg-gray-200 rounded-full"></div>
</div>
```

#### 3. Progressive Enhancement
```html
<!-- Basic form that works without JavaScript -->
<form method="POST" action="/api/users">
    <input name="name" required>
    <button type="submit">Create User</button>
</form>

<!-- Enhanced with HTMX -->
<form 
    method="POST" 
    action="/api/users"
    hx-post="/api/users"
    hx-target="#user-list"
    hx-swap="afterbegin"
    hx-on::after-request="if(event.detail.successful) this.reset()">
    
    <input name="name" required>
    <button type="submit">Create User</button>
</form>
```
<!-- LLM-RESPONSE-OPTIMIZATION-END -->

---

## Security Considerations

### CSRF Protection

<!-- LLM-SECURITY-PATTERNS-START -->
**SECURITY IMPLEMENTATION PATTERNS:**

#### 1. CSRF Token Integration
```go
// Middleware to inject CSRF token
func CSRFMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := generateCSRFToken(r)
        
        // Add token to context
        ctx := context.WithValue(r.Context(), "csrf_token", token)
        
        // Set cookie
        http.SetCookie(w, &http.Cookie{
            Name:     "csrf_token",
            Value:    token,
            HttpOnly: true,
            Secure:   true,
            SameSite: http.SameSiteStrictMode,
        })
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

```html
<!-- Include CSRF token in HTMX requests -->
<form 
    hx-post="/api/users"
    hx-headers='{"X-CSRF-Token": document.cookie.match(/csrf_token=([^;]+)/)[1]}'>
    
<!-- Or use meta tag approach -->
<meta name="csrf-token" content="{{ .CSRFToken }}">
<form 
    hx-post="/api/users"
    hx-headers='{"X-CSRF-Token": document.querySelector("meta[name=csrf-token]").content}'>
```

#### 2. Input Sanitization
```go
// Server-side input validation and sanitization
func (h *Handler) sanitizeInput(input string) string {
    // Remove potentially dangerous HTML
    policy := bluemonday.UGCPolicy()
    cleaned := policy.Sanitize(input)
    
    // Additional custom sanitization
    cleaned = strings.TrimSpace(cleaned)
    cleaned = html.EscapeString(cleaned)
    
    return cleaned
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    name := h.sanitizeInput(r.FormValue("name"))
    email := h.sanitizeInput(r.FormValue("email"))
    
    // Validate inputs
    if name == "" || email == "" {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    
    // Process request...
}
```

#### 3. Rate Limiting
```go
// Rate limiting middleware
func RateLimitMiddleware(limiter *rate.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                w.Header().Set("HX-Trigger", `{"showNotification": {"type": "warning", "message": "Too many requests. Please slow down."}}`)
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```
<!-- LLM-SECURITY-PATTERNS-END -->

---

## Testing HTMX Interactions

### Integration Testing

<!-- LLM-TESTING-PATTERNS-START -->
**TESTING STRATEGIES FOR HTMX:**

#### 1. Server-side Testing
```go
func TestUserListEndpoint(t *testing.T) {
    // Setup test server
    server := setupTestServer()
    defer server.Close()
    
    // Make HTMX request
    req := httptest.NewRequest("GET", "/api/users", nil)
    req.Header.Set("HX-Request", "true")
    
    w := httptest.NewRecorder()
    server.ServeHTTP(w, req)
    
    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "user-table")
    
    // Check HTMX-specific headers
    assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
}

func TestFormValidation(t *testing.T) {
    server := setupTestServer()
    defer server.Close()
    
    // Submit invalid form
    form := url.Values{}
    form.Set("name", "") // Missing required field
    
    req := httptest.NewRequest("POST", "/api/users", strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("HX-Request", "true")
    
    w := httptest.NewRecorder()
    server.ServeHTTP(w, req)
    
    // Should return validation error
    assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
    assert.Contains(t, w.Body.String(), "error")
    assert.Equal(t, "outerHTML", w.Header().Get("HX-Reswap"))
}
```

#### 2. End-to-end Testing
```javascript
// Playwright test for HTMX interactions
test('user creation flow', async ({ page }) => {
    await page.goto('/users');
    
    // Click create button
    await page.click('[hx-get="/users/create"]');
    
    // Wait for modal to appear
    await page.waitForSelector('.modal');
    
    // Fill form
    await page.fill('input[name="name"]', 'John Doe');
    await page.fill('input[name="email"]', 'john@example.com');
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Wait for table update
    await page.waitForSelector('#user-list tr:has-text("John Doe")');
    
    // Verify modal closed
    await expect(page.locator('.modal')).not.toBeVisible();
});

test('search functionality', async ({ page }) => {
    await page.goto('/users');
    
    // Type in search box
    await page.fill('input[name="search"]', 'john');
    
    // Wait for debounced request
    await page.waitForTimeout(500);
    
    // Verify filtered results
    await expect(page.locator('#user-table tbody tr')).toHaveCount(1);
    await expect(page.locator('#user-table')).toContainText('john');
});
```

#### 3. Mock Server Testing
```javascript
// Mock server for testing HTMX requests
const mockServer = setupServer(
    rest.get('/api/users', (req, res, ctx) => {
        return res(
            ctx.set('Content-Type', 'text/html'),
            ctx.body('<tr><td>Mocked User</td></tr>')
        );
    }),
    
    rest.post('/api/users', (req, res, ctx) => {
        return res(
            ctx.set('HX-Trigger', '{"userCreated": true}'),
            ctx.body('<tr><td>New User</td></tr>')
        );
    })
);

beforeAll(() => mockServer.listen());
afterEach(() => mockServer.resetHandlers());
afterAll(() => mockServer.close());
```
<!-- LLM-TESTING-PATTERNS-END -->

## Best Practices Summary

<!-- LLM-BEST-PRACTICES-START -->
**HTMX INTEGRATION BEST PRACTICES:**

### 1. Progressive Enhancement
- **Always provide fallback** functionality without JavaScript
- **Enhance existing forms** rather than building HTMX-only interfaces
- **Test without HTMX** to ensure core functionality works

### 2. Server Response Design
- **Return semantic HTML** that matches your design system
- **Use appropriate HTTP status codes** for different scenarios
- **Include HTMX headers** for coordinated client-side behavior
- **Keep responses minimal** - only return what changed

### 3. Error Handling
- **Provide meaningful error messages** to users
- **Use proper HTTP status codes** for different error types
- **Include retry mechanisms** for transient failures
- **Log errors server-side** for debugging

### 4. Performance Optimization
- **Debounce rapid requests** like search and filters
- **Cache responses** when appropriate
- **Use lazy loading** for expensive content
- **Minimize response payloads** by returning only changed data

### 5. Security Considerations
- **Validate all inputs** server-side
- **Include CSRF protection** for state-changing requests
- **Sanitize HTML output** to prevent XSS
- **Rate limit requests** to prevent abuse

### 6. Testing Strategy
- **Test server endpoints** independently
- **Mock HTMX headers** in integration tests
- **Use end-to-end tests** for complex workflows
- **Test error scenarios** and edge cases
<!-- LLM-BEST-PRACTICES-END -->

## References

<!-- LLM-REFERENCES-START -->
**RELATED DOCUMENTATION:**
- [Form Components](../components/forms.md) - Form component patterns and validation
- [Architecture Fundamentals](../fundamentals/architecture.md) - System design principles
- [Validation Guide](../guides/validation-guide.md) - Server-side validation implementation

**EXTERNAL REFERENCES:**
- `templ-llms.md` - Advanced Templ templating features
- `flowbite-llms-full.txt` - Complete Flowbite component styling reference

**OFFICIAL DOCUMENTATION:**
- [HTMX Documentation](https://htmx.org/docs/) - Complete HTMX reference and guides
- [HTMX Examples](https://htmx.org/examples/) - Common interaction patterns
- [Templ Guide](https://templ.guide) - Go templating language documentation
- [HTTP Status Codes](https://httpstatuses.com/) - HTTP response code reference
<!-- LLM-REFERENCES-END -->

<!-- LLM-METADATA-START -->
**METADATA FOR AI ASSISTANTS:**
- File Type: Implementation Patterns Guide
- Scope: HTMX integration patterns and server interaction strategies
- Target Audience: Full-stack developers
- Complexity: Intermediate to Advanced
- Focus: Server-driven UI interactions with progressive enhancement
- Dependencies: HTMX + Templ + Go backend
<!-- LLM-METADATA-END -->