# ERP UI Templates: Templ + HTMX + Alpine + Flowbite Implementation Guide

## Architecture Overview

### Template Structure
```
templates/
├── layouts/
│   ├── base.templ           # Main layout with nav/sidebar
│   ├── modal.templ          # Modal container template
│   └── table.templ          # Reusable table layout
├── components/
│   ├── forms/
│   │   ├── input.templ      # Standard input fields
│   │   ├── select.templ     # Dropdown components
│   │   └── validation.templ # Error display
│   ├── tables/
│   │   ├── header.templ     # Table headers with sorting
│   │   ├── row.templ        # Table row template
│   │   └── pagination.templ # Pagination controls
│   └── filters/
│       ├── search.templ     # Search input
│       └── filter-panel.templ # Advanced filters
└── pages/
    ├── list.templ           # List/Index page
    ├── create.templ         # Create form
    ├── edit.templ           # Edit form
    └── view.templ           # Detail view
```

## 1. Base Layout Template

```go
// layouts/base.templ
templ BaseLayout(title string) {
    <!DOCTYPE html>
    <html lang="en" x-data="{ sidebarOpen: false }">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{title} - ERP System</title>
        <script src="https://unpkg.com/htmx.org@1.9.10"></script>
        <script src="https://unpkg.com/alpinejs@3.13.3/dist/cdn.min.js" defer></script>
        <link href="https://cdnjs.cloudflare.com/ajax/libs/flowbite/2.2.0/flowbite.min.css" rel="stylesheet">
    </head>
    <body class="bg-gray-50">
        <!-- Sidebar -->
        @Sidebar()
        
        <!-- Main Content -->
        <div class="p-4 sm:ml-64">
            <div class="p-4 border-2 border-gray-200 border-dashed rounded-lg">
                { children... }
            </div>
        </div>
        
        <!-- Toast Container -->
        <div id="toast-container" class="fixed top-5 right-5 z-50"></div>
        
        <script src="https://cdnjs.cloudflare.com/ajax/libs/flowbite/2.2.0/flowbite.min.js"></script>
    </body>
    </html>
}
```

## 2. List/Index Page Template

```go
// pages/list.templ
templ ListPage(entity string, data []interface{}) {
    @BaseLayout(entity + " Management") {
        <div x-data="listManager()">
            <!-- Header Section -->
            <div class="flex justify-between items-center mb-6">
                <h1 class="text-2xl font-semibold text-gray-900">{entity} Management</h1>
                <button 
                    hx-get={"/"+entity+"/create"} 
                    hx-target="#modal-container"
                    hx-trigger="click"
                    class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg">
                    Add New {entity}
                </button>
            </div>
            
            <!-- Search & Filters -->
            @SearchAndFilters(entity)
            
            <!-- Data Table -->
            <div id="table-container" class="bg-white shadow rounded-lg">
                @DataTable(entity, data)
            </div>
            
            <!-- Modal Container -->
            <div id="modal-container"></div>
        </div>
    }
}

script listManager() {
    return {
        selectedItems: [],
        selectAll: false,
        
        toggleSelectAll() {
            this.selectAll = !this.selectAll;
            const checkboxes = document.querySelectorAll('input[name="selected[]"]');
            checkboxes.forEach(cb => cb.checked = this.selectAll);
            this.updateSelectedItems();
        },
        
        updateSelectedItems() {
            const checkboxes = document.querySelectorAll('input[name="selected[]"]:checked');
            this.selectedItems = Array.from(checkboxes).map(cb => cb.value);
        },
        
        bulkAction(action) {
            if (this.selectedItems.length === 0) {
                alert('Please select items first');
                return;
            }
            // Implement bulk actions
            htmx.ajax('POST', `/${entity}/bulk/${action}`, {
                values: { ids: this.selectedItems }
            });
        }
    }
}
```

## 3. Search and Filters Component

```go
// components/filters/search-and-filters.templ
templ SearchAndFilters(entity string) {
    <div class="bg-white p-4 rounded-lg shadow mb-6" x-data="{ showFilters: false }">
        <!-- Quick Search -->
        <div class="flex gap-4 items-center mb-4">
            <div class="flex-1">
                <input 
                    type="text" 
                    name="search"
                    placeholder="Search {entity}..."
                    hx-get={"/"+entity+"/search"}
                    hx-target="#table-container"
                    hx-trigger="keyup changed delay:300ms"
                    hx-include="[name='filters']"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md">
            </div>
            <button 
                @click="showFilters = !showFilters"
                class="px-4 py-2 text-sm bg-gray-100 rounded-md">
                <span x-show="!showFilters">Show Filters</span>
                <span x-show="showFilters">Hide Filters</span>
            </button>
        </div>
        
        <!-- Advanced Filters -->
        <div x-show="showFilters" x-transition class="border-t pt-4">
            @AdvancedFilters(entity)
        </div>
    </div>
}

templ AdvancedFilters(entity string) {
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <!-- Status Filter -->
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
            <select 
                name="filters[status]"
                hx-get={"/"+entity+"/search"}
                hx-target="#table-container"
                hx-trigger="change"
                hx-include="[name='search'], [name^='filters']"
                class="w-full px-3 py-2 border border-gray-300 rounded-md">
                <option value="">All Statuses</option>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
            </select>
        </div>
        
        <!-- Date Range -->
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Date From</label>
            <input 
                type="date" 
                name="filters[date_from]"
                hx-get={"/"+entity+"/search"}
                hx-target="#table-container"
                hx-trigger="change"
                hx-include="[name='search'], [name^='filters']"
                class="w-full px-3 py-2 border border-gray-300 rounded-md">
        </div>
        
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Date To</label>
            <input 
                type="date" 
                name="filters[date_to]"
                hx-get={"/"+entity+"/search"}
                hx-target="#table-container"
                hx-trigger="change"
                hx-include="[name='search'], [name^='filters']"
                class="w-full px-3 py-2 border border-gray-300 rounded-md">
        </div>
    </div>
}
```

## 4. Data Table Component

```go
// components/tables/data-table.templ
templ DataTable(entity string, data []interface{}) {
    <div class="overflow-x-auto">
        <table class="w-full text-sm text-left text-gray-500">
            <thead class="text-xs text-gray-700 uppercase bg-gray-50">
                @TableHeader(entity)
            </thead>
            <tbody id="table-body">
                for _, item := range data {
                    @TableRow(entity, item)
                }
            </tbody>
        </table>
        
        @Pagination()
    </div>
}

templ TableHeader(entity string) {
    <tr>
        <th class="p-4">
            <input 
                type="checkbox" 
                @change="toggleSelectAll()"
                class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded">
        </th>
        <th class="px-6 py-3 cursor-pointer" 
            hx-get={"/"+entity+"/search?sort=name"}
            hx-target="#table-container"
            hx-include="[name='search'], [name^='filters']">
            Name 
            <i class="fas fa-sort ml-1"></i>
        </th>
        <th class="px-6 py-3">Status</th>
        <th class="px-6 py-3">Created</th>
        <th class="px-6 py-3">Actions</th>
    </tr>
}

templ TableRow(entity string, item interface{}) {
    <tr class="bg-white border-b hover:bg-gray-50" x-data>
        <td class="p-4">
            <input 
                type="checkbox" 
                name="selected[]" 
                value={item.ID}
                @change="updateSelectedItems()"
                class="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded">
        </td>
        <td class="px-6 py-4 font-medium text-gray-900">
            {item.Name}
        </td>
        <td class="px-6 py-4">
            @StatusBadge(item.Status)
        </td>
        <td class="px-6 py-4">
            {item.CreatedAt.Format("Jan 2, 2006")}
        </td>
        <td class="px-6 py-4">
            @ActionButtons(entity, item.ID)
        </td>
    </tr>
}
```

## 5. Form Templates (Create/Edit)

```go
// pages/create.templ
templ CreateForm(entity string) {
    @ModalLayout() {
        <div class="p-6">
            <h2 class="text-xl font-semibold mb-4">Create New {entity}</h2>
            
            <form 
                hx-post={"/"+entity}
                hx-target="#table-container"
                hx-on::after-request="if(event.detail.successful) closeModal()"
                x-data="formHandler()">
                
                @FormFields(entity)
                
                <div class="flex justify-end space-x-3 mt-6">
                    <button 
                        type="button" 
                        @click="closeModal()"
                        class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300">
                        Cancel
                    </button>
                    <button 
                        type="submit"
                        :disabled="submitting"
                        class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">
                        <span x-show="!submitting">Create</span>
                        <span x-show="submitting">Creating...</span>
                    </button>
                </div>
            </form>
        </div>
    }
}

script formHandler() {
    return {
        submitting: false,
        
        init() {
            this.$el.addEventListener('htmx:beforeRequest', () => {
                this.submitting = true;
            });
            
            this.$el.addEventListener('htmx:afterRequest', () => {
                this.submitting = false;
            });
        }
    }
}
```

## 6. Input Components

```go
// components/forms/input.templ
templ TextInput(name, label, value, placeholder string, required bool) {
    <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">
            {label}
            if required {
                <span class="text-red-500">*</span>
            }
        </label>
        <input 
            type="text" 
            name={name}
            value={value}
            placeholder={placeholder}
            if required {
                required
            }
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            x-data
            x-on:blur="htmx.trigger(this, 'validate')">
        <div class="error-message text-red-500 text-sm mt-1 hidden"></div>
    </div>
}

templ SelectInput(name, label string, options []Option, value string, required bool) {
    <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">
            {label}
            if required {
                <span class="text-red-500">*</span>
            }
        </label>
        <select 
            name={name}
            if required {
                required
            }
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option value="">Select {label}</option>
            for _, option := range options {
                <option 
                    value={option.Value}
                    if option.Value == value {
                        selected
                    }>
                    {option.Label}
                </option>
            }
        </select>
    </div>
}
```

## 7. Alpine.js Data Management

```javascript
// Global Alpine stores for state management
document.addEventListener('alpine:init', () => {
    Alpine.store('app', {
        notifications: [],
        
        addNotification(type, message) {
            const id = Date.now();
            this.notifications.push({ id, type, message });
            setTimeout(() => this.removeNotification(id), 5000);
        },
        
        removeNotification(id) {
            this.notifications = this.notifications.filter(n => n.id !== id);
        }
    });
    
    Alpine.store('forms', {
        validation: {},
        
        setFieldError(field, error) {
            this.validation[field] = error;
        },
        
        clearFieldError(field) {
            delete this.validation[field];
        }
    });
});

// Global functions
function closeModal() {
    document.getElementById('modal-container').innerHTML = '';
}

function showNotification(type, message) {
    Alpine.store('app').addNotification(type, message);
}
```

## 8. HTMX Integration Patterns

### Server Response Headers
```go
// In your Go handlers
w.Header().Set("HX-Trigger", "refreshTable")
w.Header().Set("HX-Trigger-After-Swap", "showNotification")
```

### Event Handling
```javascript
// Listen for HTMX events
document.body.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.xhr.getResponseHeader('HX-Trigger')) {
        // Handle custom triggers
        const triggers = JSON.parse(evt.detail.xhr.getResponseHeader('HX-Trigger'));
        if (triggers.showNotification) {
            showNotification(triggers.showNotification.type, triggers.showNotification.message);
        }
    }
});
```

## Key Benefits of This Architecture

1. **Performance**: HTMX handles partial page updates efficiently
2. **Maintainability**: Templ provides type-safe templates
3. **Interactivity**: Alpine.js adds reactivity without complexity
4. **Design System**: Flowbite ensures consistency
5. **Progressive Enhancement**: Works without JavaScript
6. **SEO Friendly**: Server-side rendered content

## Best Practices

1. **Keep Alpine.js Simple**: Use for UI state, not business logic
2. **Leverage HTMX Attributes**: Use `hx-include`, `hx-vals` for context
3. **Template Composition**: Break down into small, reusable components
4. **Error Handling**: Always provide user feedback
5. **Loading States**: Show progress for long operations
6. **Accessibility**: Use proper ARIA attributes and semantic HTML
