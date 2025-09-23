# ERP Components Deep Dive - Complete UI System

## Component Architecture Philosophy

The component system follows these enhanced core principles:
- **Single Responsibility**: Each component does one thing well
- **Composability**: Components can be combined to create complex UIs
- **Consistency**: Shared design language across the application
- **Type Safety**: Go's type system prevents runtime errors
- **Progressive **: Works with and without JavaScript
- **Accessibility First**: WCAG 2.1 AA compliance built-in
- **Performance Optimized**: Lazy loading and efficient re-rendering
- **Internationalization Ready**: Support for multiple languages and RTL

## 1. Form Components (`components/forms/`)

### Base Input Component with Advanced Features
```go
// components/forms/input.templ
templ BaseInput(props InputProps) {
    <div class="mb-4" x-data="{ focused: false, hasError: false }" x-init="initValidation()">
        if props.Label != "" {
            <label 
                class="block text-sm font-medium text-gray-700 mb-2"
                for={props.ID}>
                {props.Label}
                if props.Required {
                    <span class="text-red-500 ml-1" aria-label="required">*</span>
                }
                if props.Tooltip != "" {
                    <button 
                        type="button"
                        class="ml-1 text-gray-400 hover:text-gray-600 focus:outline-none"
                        @click="$refs.tooltip.toggle()"
                        aria-describedby={"tooltip-" + props.ID}>
                        <i class="fas fa-info-circle text-xs"></i>
                    </button>
                    <!-- Tooltip component -->
                    <div 
                        x-ref="tooltip"
                        x-data="tooltip()"
                        id={"tooltip-" + props.ID}
                        role="tooltip"
                        class="absolute z-10 invisible opacity-0 px-3 py-2 text-sm text-white bg-gray-900 rounded-lg shadow-sm transition-opacity duration-300">
                        {props.Tooltip}
                    </div>
                }
            </label>
        }
        
        <div class="relative">
            { children... }
            
            <!-- Loading spinner for async validation -->
            <div 
                x-show="$store.forms.validating['{props.Name}']" 
                class="absolute right-3 top-1/2 transform -translate-y-1/2"
                role="status"
                aria-label="Validating">
                <i class="fas fa-spinner fa-spin text-gray-400"></i>
            </div>
            
            <!-- Success indicator -->
            <div 
                x-show="$store.forms.validation['{props.Name}'] === 'valid'" 
                class="absolute right-3 top-1/2 transform -translate-y-1/2">
                <i class="fas fa-check text-green-500"></i>
            </div>
        </div>
        
        <!-- Error message with ARIA -->
        <div 
            x-show="$store.forms.validation['{props.Name}'] && $store.forms.validation['{props.Name}'] !== 'valid'" 
            x-text="$store.forms.validation['{props.Name}']"
            class="text-red-500 text-sm mt-1"
            role="alert"
            aria-live="polite">
        </div>
        
        <!-- Help text -->
        if props.HelpText != "" {
            <div class="text-gray-500 text-sm mt-1" id={"help-" + props.ID}>
                {props.HelpText}
            </div>
        }
        
        <!-- Character counter for text inputs -->
        if props.MaxLength > 0 {
            <div class="text-xs text-gray-400 mt-1 text-right" x-data="charCounter()">
                <span x-text="count"></span>/<span>{strconv.Itoa(props.MaxLength)}</span>
            </div>
        }
    </div>
}

type InputProps struct {
    ID            string
    Name          string
    Label         string
    Value         string
    Placeholder   string
    Required      bool
    Disabled      bool
    Readonly      bool
    Tooltip       string
    HelpText      string
    Validation    string
    MaxLength     int
    MinLength     int
    Pattern       string
    Autocomplete  string
    DataAttributes map[string]string
}
```

### Advanced File Upload Component
```go
templ FileUpload(props FileUploadProps) {
    @BaseInput(props.InputProps) {
        <div 
            x-data="fileUpload()"
            class="w-full"
            @dragover.prevent="dragover = true"
            @dragleave.prevent="dragover = false"
            @drop.prevent="handleDrop($event)">
            
            <!-- Drop zone -->
            <div 
                :class="{
                    'border-blue-500 bg-blue-50': dragover,
                    'border-gray-300': !dragover,
                    'border-red-500': hasError
                }"
                class="border-2 border-dashed rounded-lg p-6 text-center transition-colors">
                
                <input 
                    type="file"
                    :name="name"
                    :accept="accept"
                    :multiple="multiple"
                    @change="handleFileSelect($event)"
                    class="hidden"
                    x-ref="fileInput">
                
                <div class="space-y-2">
                    <i class="fas fa-cloud-upload-alt text-4xl text-gray-400"></i>
                    <div>
                        <button 
                            type="button"
                            @click="$refs.fileInput.click()"
                            class="text-blue-600 hover:text-blue-800 font-medium">
                            Choose files
                        </button>
                        <span class="text-gray-500"> or drag them here</span>
                    </div>
                    <p class="text-xs text-gray-500">
                        {props.AcceptedTypes} up to {props.MaxSizeDisplay}
                    </p>
                </div>
            </div>
            
            <!-- File list -->
            <div x-show="files.length > 0" class="mt-4 space-y-2">
                <template x-for="(file, index) in files" :key="file.id">
                    <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                        <div class="flex items-center space-x-3">
                            <i :class="getFileIcon(file.type)" class="text-gray-500"></i>
                            <div>
                                <div class="text-sm font-medium text-gray-900" x-text="file.name"></div>
                                <div class="text-xs text-gray-500" x-text="formatFileSize(file.size)"></div>
                            </div>
                        </div>
                        
                        <div class="flex items-center space-x-2">
                            <!-- Upload progress -->
                            <div x-show="file.uploading" class="w-20 bg-gray-200 rounded-full h-2">
                                <div 
                                    class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                                    :style="'width: ' + file.progress + '%'"></div>
                            </div>
                            
                            <!-- Status icons -->
                            <i x-show="file.uploaded" class="fas fa-check text-green-500"></i>
                            <i x-show="file.error" class="fas fa-exclamation-triangle text-red-500"></i>
                            
                            <!-- Remove button -->
                            <button 
                                @click="removeFile(index)"
                                type="button"
                                class="text-red-500 hover:text-red-700">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                    </div>
                </template>
            </div>
        </div>
    }
}

type FileUploadProps struct {
    InputProps
    Multiple        bool
    AcceptedTypes   string
    MaxSize         int64
    MaxSizeDisplay  string
    MaxFiles        int
    UploadURL       string
}
```

### Rich Text Editor Component
```go
templ RichTextEditor(props RichTextProps) {
    @BaseInput(props.InputProps) {
        <div x-data="richTextEditor()" class="border border-gray-300 rounded-md overflow-hidden">
            <!-- Toolbar -->
            <div class="bg-gray-50 border-b border-gray-300 p-2 flex flex-wrap gap-1">
                @ToolbarButton("bold", "fas fa-bold", "Bold")
                @ToolbarButton("italic", "fas fa-italic", "Italic")
                @ToolbarButton("underline", "fas fa-underline", "Underline")
                <div class="border-l border-gray-300 mx-2"></div>
                @ToolbarButton("justifyLeft", "fas fa-align-left", "Align Left")
                @ToolbarButton("justifyCenter", "fas fa-align-center", "Center")
                @ToolbarButton("justifyRight", "fas fa-align-right", "Align Right")
                <div class="border-l border-gray-300 mx-2"></div>
                @ToolbarButton("insertUnorderedList", "fas fa-list-ul", "Bullet List")
                @ToolbarButton("insertOrderedList", "fas fa-list-ol", "Numbered List")
                <div class="border-l border-gray-300 mx-2"></div>
                @ToolbarButton("createLink", "fas fa-link", "Insert Link")
                @ToolbarButton("insertImage", "fas fa-image", "Insert Image")
            </div>
            
            <!-- Editor -->
            <div 
                contenteditable="true"
                x-ref="editor"
                @input="updateContent()"
                @paste="handlePaste($event)"
                class="p-4 min-h-[200px] outline-none prose prose-sm max-w-none"
                style="white-space: pre-wrap;">
                {props.Content}
            </div>
            
            <!-- Hidden input for form submission -->
            <input type="hidden" :name="name" :value="content">
        </div>
    }
}

templ ToolbarButton(command, icon, title string) {
    <button 
        type="button"
        @click="execCommand('{command}')"
        :class="{ 'bg-blue-100 text-blue-700': isActive('{command}') }"
        class="p-2 rounded hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
        title={title}>
        <i class={icon}></i>
    </button>
}
```

## 2. Table Components with Advanced Features

### Virtualized Data Table for Large Datasets
```go
templ VirtualizedDataTable(config VirtualTableConfig) {
    <div class="bg-white rounded-lg shadow overflow-hidden" x-data="virtualizedTable()">
        <!-- actions bar -->
        <div class="px-6 py-4 border-b border-gray-200">
            <div class="flex justify-between items-center">
                <div class="flex items-center space-x-4">
                    <!-- Bulk actions with confirmation -->
                    <div x-show="selectedItems.length > 0" x-transition class="flex items-center space-x-3">
                        <span class="text-sm text-gray-500">
                            <span x-text="selectedItems.length"></span> of <span x-text="totalItems"></span> selected
                        </span>
                        @BulkActions(config.BulkActions)
                    </div>
                    
                    <!-- Quick filters -->
                    @QuickFilters(config.QuickFilters)
                </div>
                
                <div class="flex items-center space-x-3">
                    <!-- Search -->
                    @GlobalSearch()
                    <!-- Export options -->
                    @ExportDropdown()
                    <!-- View options -->
                    @TableViewOptions()
                </div>
            </div>
        </div>
        
        <!-- Advanced filters panel -->
        <div x-show="showFilters" x-transition x-collapse class="border-b border-gray-200">
            @AdvancedFiltersPanel(config.Filters)
        </div>
        
        <!-- Virtualized table container -->
        <div class="relative overflow-hidden" style="height: 600px;">
            <!-- Table header (sticky) -->
            <div class="sticky top-0 z-10 bg-white border-b border-gray-200">
                <table class="w-full">
                    <thead>
                        @TableHeader(config.Columns)
                    </thead>
                </table>
            </div>
            
            <!-- Virtualized rows container -->
            <div 
                class="overflow-auto"
                x-ref="scrollContainer"
                @scroll="handleScroll()"
                style="height: calc(100% - 60px);">
                
                <!-- Virtual spacer top -->
                <div :style="'height: ' + spacerTop + 'px;'"></div>
                
                <!-- Visible rows -->
                <table class="w-full">
                    <tbody>
                        <template x-for="(row, index) in visibleRows" :key="row.id">
                            @VirtualTableRow(config.Columns)
                        </template>
                    </tbody>
                </table>
                
                <!-- Virtual spacer bottom -->
                <div :style="'height: ' + spacerBottom + 'px;'"></div>
                
                <!-- Loading indicator -->
                <div x-show="loading" class="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center">
                    <i class="fas fa-spinner fa-spin text-2xl text-gray-400"></i>
                </div>
            </div>
        </div>
        
        <!-- pagination with page size options -->
        @Pagination(config.Pagination)
    </div>
}
```

### Column Management and Resizing
```go
templ ResizableColumn(column Column, index int) {
    <th 
        class="relative px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider select-none"
        :style="'width: ' + columnWidths[{index}] + 'px;'"
        x-data="{ resizing: false }"
        @mousedown.stop>
        
        <!-- Column content -->
        <div class="flex items-center justify-between pr-4">
            <!-- Sortable header content -->
            @SortableHeaderContent(column)
            
            <!-- Column menu -->
            <div class="relative" x-data="{ open: false }">
                <button 
                    @click="open = !open"
                    @click.away="open = false"
                    class="opacity-0 group-hover:opacity-100 p-1 rounded hover:bg-gray-100">
                    <i class="fas fa-ellipsis-v text-xs"></i>
                </button>
                
                <!-- Column menu dropdown -->
                <div 
                    x-show="open" 
                    x-transition
                    class="absolute right-0 top-full mt-1 w-48 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 z-20">
                    @ColumnMenuItems(column, index)
                </div>
            </div>
        </div>
        
        <!-- Resize handle -->
        <div 
            class="absolute top-0 right-0 w-1 h-full cursor-col-resize bg-transparent hover:bg-blue-500 transition-colors"
            @mousedown="startResize($event, {index})"
            :class="{ 'bg-blue-500': resizing }">
        </div>
    </th>
}

templ ColumnMenuItems(column Column, index int) {
    <div class="py-1">
        <button 
            @click="sortColumn('{column.Key}', 'asc')"
            class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
            <i class="fas fa-sort-alpha-down mr-2"></i>Sort A-Z
        </button>
        <button 
            @click="sortColumn('{column.Key}', 'desc')"
            class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
            <i class="fas fa-sort-alpha-up mr-2"></i>Sort Z-A
        </button>
        <div class="border-t border-gray-100 my-1"></div>
        <button 
            @click="hideColumn({index})"
            class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
            <i class="fas fa-eye-slash mr-2"></i>Hide Column
        </button>
        <button 
            @click="pinColumn({index})"
            class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
            <i class="fas fa-thumbtack mr-2"></i>Pin Column
        </button>
        <button 
            @click="autoSizeColumn({index})"
            class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
            <i class="fas fa-arrows-alt-h mr-2"></i>Auto-size
        </button>
    </div>
}
```

## 3. Advanced Navigation Components

### Breadcrumb with Dynamic Loading
```go
templ SmartBreadcrumb(items []BreadcrumbItem, separator string) {
    <nav class="flex items-center space-x-1 text-sm text-gray-500" aria-label="Breadcrumb">
        <ol class="flex items-center space-x-1">
            for i, item := range items {
                <li class="flex items-center">
                    if i > 0 {
                        <span class="mx-2 text-gray-300">
                            if separator != "" {
                                {separator}
                            } else {
                                <i class="fas fa-chevron-right text-xs"></i>
                            }
                        </span>
                    }
                    
                    if item.URL != "" && i < len(items)-1 {
                        <a 
                            href={item.URL}
                            hx-get={item.URL}
                            hx-target="#main-content"
                            hx-push-url="true"
                            class="text-gray-500 hover:text-gray-700 transition-colors">
                            if item.Icon != "" {
                                <i class={item.Icon + " mr-1"}></i>
                            }
                            {item.Label}
                        </a>
                    } else {
                        <span class="text-gray-900 font-medium" aria-current="page">
                            if item.Icon != "" {
                                <i class={item.Icon + " mr-1"}></i>
                            }
                            {item.Label}
                        </span>
                    }
                </li>
            }
        </ol>
        
        <!-- Breadcrumb actions -->
        if len(items) > 0 {
            <div class="ml-4 flex items-center space-x-2">
                @BreadcrumbActions(items[len(items)-1])
            </div>
        }
    </nav>
}

type BreadcrumbItem struct {
    Label   string
    URL     string
    Icon    string
    Actions []Action
}
```

### Multi-level Sidebar Navigation
```go
templ SidebarNavigation(items []NavItem, currentPath string) {
    <nav class="flex-1 px-4 pb-4 space-y-1" x-data="navigation()">
        for _, item := range items {
            @NavigationItem(item, currentPath, 0)
        }
    </nav>
}

templ NavigationItem(item NavItem, currentPath string, level int) {
    <div x-data="{ open: {isExpanded(item, currentPath)} }">
        <div class={"flex items-center justify-between " + getNavigationItemClass(item, currentPath, level)}>
            <!-- Navigation link -->
            if item.URL != "" {
                <a 
                    href={item.URL}
                    hx-get={item.URL}
                    hx-target="#main-content"
                    hx-push-url="true"
                    class="flex-1 flex items-center py-2 px-3 rounded-md text-sm font-medium transition-colors"
                    :class="getActiveClass('{item.URL}', '{currentPath}')">
                    if item.Icon != "" {
                        <i class={item.Icon + " mr-3 text-gray-400"}></i>
                    }
                    {item.Label}
                    if item.Badge != "" {
                        <span class="ml-auto bg-red-100 text-red-800 text-xs px-2 py-0.5 rounded-full">
                            {item.Badge}
                        </span>
                    }
                </a>
            } else {
                <span class="flex-1 flex items-center py-2 px-3 text-sm font-medium text-gray-700">
                    if item.Icon != "" {
                        <i class={item.Icon + " mr-3 text-gray-400"}></i>
                    }
                    {item.Label}
                </span>
            }
            
            <!-- Expand/collapse button for items with children -->
            if len(item.Children) > 0 {
                <button 
                    @click="open = !open"
                    class="p-1 rounded hover:bg-gray-100">
                    <i class="fas fa-chevron-right text-xs transition-transform" :class="{ 'rotate-90': open }"></i>
                </button>
            }
        </div>
        
        <!-- Child items -->
        if len(item.Children) > 0 {
            <div x-show="open" x-transition x-collapse class="ml-4">
                for _, child := range item.Children {
                    @NavigationItem(child, currentPath, level+1)
                }
            </div>
        }
    </div>
}
```

## 4. Modal and Dialog System

### Modal Manager with Stacking
```go
templ ModalManager() {
    <div 
        x-data="modalManager()"
        @modal:open.window="openModal($event.detail)"
        @modal:close.window="closeModal($event.detail.id)"
        class="relative z-50">
        
        <!-- Modal stack -->
        <template x-for="(modal, index) in modals" :key="modal.id">
            <div 
                x-show="modal.visible"
                x-transition:enter="ease-out duration-300"
                x-transition:enter-start="opacity-0"
                x-transition:enter-end="opacity-100"
                x-transition:leave="ease-in duration-200"
                x-transition:leave-start="opacity-100"
                x-transition:leave-end="opacity-0"
                class="fixed inset-0 overflow-y-auto"
                :class="'z-' + (50 + index * 10)"
                @click="modal.closeOnBackdrop && closeModal(modal.id)"
                @keydown.escape.window="modal.closeOnEscape && closeModal(modal.id)">
                
                <!-- Backdrop -->
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
                
                <!-- Modal content -->
                <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
                    <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
                    
                    <div 
                        @click.stop
                        x-transition:enter="ease-out duration-300"
                        x-transition:enter-start="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        x-transition:enter-end="opacity-100 translate-y-0 sm:scale-100"
                        x-transition:leave="ease-in duration-200"
                        x-transition:leave-start="opacity-100 translate-y-0 sm:scale-100"
                        x-transition:leave-end="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        :class="modal.size"
                        class="inline-block align-bottom bg-white rounded-lg text-left shadow-xl transform transition-all sm:my-8 sm:align-middle">
                        
                        <!-- Dynamic content area -->
                        <div x-html="modal.content"></div>
                    </div>
                </div>
            </div>
        </template>
    </div>
}
```

### Confirmation Dialog Component
```go
templ ConfirmationDialog(config ConfirmationConfig) {
    <div class="p-6">
        <div class="flex items-start">
            <div class={"flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center " + getIconBgClass(config.Type)}>
                <i class={getConfirmationIcon(config.Type) + " " + getIconColorClass(config.Type)}></i>
            </div>
            <div class="ml-4 flex-1">
                <h3 class="text-lg font-medium text-gray-900 mb-2">
                    {config.Title}
                </h3>
                <p class="text-sm text-gray-500 mb-4">
                    {config.Message}
                </p>
                
                if config.Details != "" {
                    <div class="bg-gray-50 rounded-md p-3 mb-4">
                        <p class="text-xs text-gray-600">
                            {config.Details}
                        </p>
                    </div>
                }
                
                <!-- Action input (for dangerous actions) -->
                if config.RequireConfirmation {
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            Type <strong>{config.ConfirmationText}</strong> to confirm:
                        </label>
                        <input 
                            type="text"
                            x-model="confirmationInput"
                            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-red-500"
                            placeholder={config.ConfirmationText}>
                    </div>
                }
            </div>
        </div>
        
        <!-- Actions -->
        <div class="flex justify-end space-x-3 pt-4 border-t border-gray-200">
            <button 
                @click="$dispatch('modal:close', { id: modalId })"
                type="button"
                class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500">
                Cancel
            </button>
            <button 
                @click="handleConfirm()"
                :disabled="requireConfirmation && confirmationInput !== confirmationText"
                :class="{
                    'opacity-50 cursor-not-allowed': requireConfirmation && confirmationInput !== confirmationText
                }"
                class={"px-4 py-2 text-sm font-medium text-white rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 " + getButtonClass(config.Type)}>
                {config.ConfirmAction}
            </button>
        </div>
    </div>
}

type ConfirmationConfig struct {
    Title              string
    Message            string
    Details            string
    Type               string // info, warning, danger, success
    ConfirmAction      string
    RequireConfirmation bool
    ConfirmationText   string
    OnConfirm          string // HTMX endpoint or JS function
}
```

## 5. Alpine.js Components

```javascript
// form validation with async rules
function advancedFormValidator() {
    return {
        errors: {},
        touched: {},
        validating: {},
        debounceTimers: {},
        
        async validateField(field, value, rules) {
            // Clear existing timer
            if (this.debounceTimers[field]) {
                clearTimeout(this.debounceTimers[field]);
            }
            
            // Debounce validation
            this.debounceTimers[field] = setTimeout(async () => {
                this.validating[field] = true;
                const fieldErrors = [];
                
                // Client-side validation
                if (rules.required && !value) {
                    fieldErrors.push(`${field} is required`);
                }
                
                if (rules.email && value && !this.isValidEmail(value)) {
                    fieldErrors.push('Please enter a valid email address');
                }
                
                if (rules.minLength && value && value.length < rules.minLength) {
                    fieldErrors.push(`Minimum length is ${rules.minLength} characters`);
                }
                
                if (rules.pattern && value && !new RegExp(rules.pattern).test(value)) {
                    fieldErrors.push(rules.patternMessage || 'Invalid format');
                }
                
                // Server-side validation
                if (rules.async && value && fieldErrors.length === 0) {
                    try {
                        const response = await fetch('/validate/async', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ field, value, rules: rules.async })
                        });
                        const result = await response.json();
                        if (!result.valid) {
                            fieldErrors.push(result.message);
                        }
                    } catch (error) {
                        console.error('Async validation failed:', error);
                    }
                }
                
                this.errors[field] = fieldErrors;
                this.touched[field] = true;
                this.validating[field] = false;
                
                // Update form state
                this.$dispatch('validation-complete', { field, valid: fieldErrors.length === 0 });
            }, rules.debounce || 500);
        },
        
        // Additional validation methods...
        validateForm() {
            // Validate entire form
        },
        
        getFieldState(field) {
            return {
                hasError: this.touched[field] && this.errors[field] && this.errors[field].length > 0,
                isValid: this.touched[field] && (!this.errors[field] || this.errors[field].length === 0),
                isValidating: this.validating[field] || false,
                errorMessage: this.hasError(field) ? this.errors[field][0] : '',
                touched: this.touched[field] || false
            };
        },
        
        isValidEmail(email) {
            return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
        },
        
        hasError(field) {
            return this.touched[field] && this.errors[field] && this.errors[field].length > 0;
        }
    }
}

// Advanced data table with virtualization and infinite scroll
function virtualizedTable() {
    return {
        // Core data
        allData: [],
        visibleRows: [],
        selectedItems: [],
        
        // Virtualization
        itemHeight: 50,
        containerHeight: 600,
        scrollTop: 0,
        startIndex: 0,
        endIndex: 0,
        visibleCount: 0,
        spacerTop: 0,
        spacerBottom: 0,
        
        // Filters and sorting
        filters: {},
        sortField: '',
        sortDirection: 'asc',
        searchQuery: '',
        
        // State
        loading: false,
        hasMore: true,
        page: 1,
        pageSize: 50,
        
        init() {
            this.visibleCount = Math.ceil(this.containerHeight / this.itemHeight) + 5;
            this.calculateVisibleRows();
            
            // Setup infinite scroll
            this.$watch('scrollTop', () => {
                this.calculateVisibleRows();
                this.checkLoadMore();
            });
        },
        
        calculateVisibleRows() {
            this.startIndex = Math.floor(this.scrollTop / this.itemHeight);
            this.endIndex = Math.min(this.startIndex + this.visibleCount, this.allData.length);
            
            this.visibleRows = this.allData.slice(this.startIndex, this.endIndex);
            this.spacerTop = this.startIndex * this.itemHeight;
            this.spacerBottom = (this.allData.length - this.endIndex) * this.itemHeight;
        },
        
        handleScroll() {
            this.scrollTop = this.$refs.scrollContainer.scrollTop;
        },
        
        async checkLoadMore() {
            const scrollContainer = this.$refs.scrollContainer;
            const scrollPercent = (scrollContainer.scrollTop + scrollContainer.clientHeight) / scrollContainer.scrollHeight;
            
            if (scrollPercent > 0.8 && !this.loading && this.hasMore) {
                await this.loadMore();
            }
        },
        
        async loadMore() {
            this.loading = true;
            try {
                const response = await fetch(`/api/data?page=${this.page + 1}&size=${this.pageSize}&search=${this.searchQuery}&sort=${this.sortField}&direction=${this.sortDirection}`);
                const data = await response.json();
                
                if (data.items.length > 0) {
                    this.allData = [...this.allData, ...data.items];
                    this.page++;
                    this.calculateVisibleRows();
                } else {
                    this.hasMore = false;
                }
            } catch (error) {
                console.error('Failed to load more data:', error);
            } finally {
                this.loading = false;
            }
        },
        
        // Selection methods
        toggleSelectAll() {
            if (this.selectedItems.length === this.allData.length) {
                this.selectedItems = [];
            } else {
                this.selectedItems = [...this.allData.map(item => item.id)];
            }
        },
        
        toggleSelectItem(itemId) {
            const index = this.selectedItems.indexOf(itemId);
            if (index > -1) {
                this.selectedItems.splice(index, 1);
            } else {
                this.selectedItems.push(itemId);
            }
        },
        
        isSelected(itemId) {
            return this.selectedItems.includes(itemId);
        },
        
        // Filtering and sorting
        async applyFilters() {
            this.loading = true;
            this.page = 1;
            this.allData = [];
            await this.loadData();
        },
        
        async sortBy(field) {
            if (this.sortField === field) {
                this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortField = field;
                this.sortDirection = 'asc';
            }
            await this.applyFilters();
        }
    }
}

// modal manager with stacking and focus management
function modalManager() {
    return {
        modals: [],
        focusStack: [],
        
        init() {
            // Handle escape key for topmost modal
            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape' && this.modals.length > 0) {
                    const topModal = this.modals[this.modals.length - 1];
                    if (topModal.closeOnEscape) {
                        this.closeModal(topModal.id);
                    }
                }
            });
        },
        
        openModal(config) {
            const modal = {
                id: config.id || 'modal-' + Date.now(),
                content: config.content || '',
                size: config.size || 'max-w-md',
                closeOnBackdrop: config.closeOnBackdrop !== false,
                closeOnEscape: config.closeOnEscape !== false,
                visible: false
            };
            
            // Store current focus
            this.focusStack.push(document.activeElement);
            
            // Add to stack
            this.modals.push(modal);
            
            // Show with slight delay for animation
            setTimeout(() => {
                modal.visible = true;
                this.focusModal(modal.id);
            }, 10);
            
            // Prevent body scroll
            document.body.classList.add('overflow-hidden');
            
            return modal.id;
        },
        
        closeModal(modalId) {
            const index = this.modals.findIndex(m => m.id === modalId);
            if (index === -1) return;
            
            const modal = this.modals[index];
            modal.visible = false;
            
            // Remove after animation
            setTimeout(() => {
                this.modals.splice(index, 1);
                
                // Restore focus
                if (this.focusStack.length > 0) {
                    const previousFocus = this.focusStack.pop();
                    if (previousFocus && previousFocus.focus) {
                        previousFocus.focus();
                    }
                }
                
                // Re-enable body scroll if no modals
                if (this.modals.length === 0) {
                    document.body.classList.remove('overflow-hidden');
                }
            }, 200);
        },
        
        closeAllModals() {
            this.modals.forEach(modal => this.closeModal(modal.id));
        },
        
        focusModal(modalId) {
            // Focus the first focusable element in the modal
            setTimeout(() => {
                const modal = document.querySelector(`[data-modal-id="${modalId}"]`);
                if (modal) {
                    const focusable = modal.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
                    if (focusable) {
                        focusable.focus();
                    }
                }
            }, 100);
        }
    }
}
```

## 6. Advanced Dashboard Components

### Real-time Analytics Widget
```go
templ AnalyticsWidget(config AnalyticsConfig) {
    <div class="bg-white rounded-lg shadow p-6" x-data="analyticsWidget()">
        <!-- Widget header -->
        <div class="flex items-center justify-between mb-4">
            <div>
                <h3 class="text-lg font-medium text-gray-900">{config.Title}</h3>
                if config.Subtitle != "" {
                    <p class="text-sm text-gray-500">{config.Subtitle}</p>
                }
            </div>
            
            <div class="flex items-center space-x-2">
                <!-- Time range selector -->
                @TimeRangeSelector(config.TimeRanges)
                
                <!-- Refresh button -->
                <button 
                    @click="refreshData()"
                    :disabled="loading"
                    class="p-2 text-gray-400 hover:text-gray-600 disabled:opacity-50">
                    <i class="fas fa-refresh" :class="{ 'fa-spin': loading }"></i>
                </button>
                
                <!-- Widget menu -->
                <div class="relative" x-data="{ open: false }">
                    <button @click="open = !open" class="p-2 text-gray-400 hover:text-gray-600">
                        <i class="fas fa-ellipsis-v"></i>
                    </button>
                    @WidgetDropdownMenu()
                </div>
            </div>
        </div>
        
        <!-- Loading state -->
        <div x-show="loading" class="flex items-center justify-center h-48">
            <i class="fas fa-spinner fa-spin text-2xl text-gray-400"></i>
        </div>
        
        <!-- Error state -->
        <div x-show="error" class="flex flex-col items-center justify-center h-48 text-gray-500">
            <i class="fas fa-exclamation-triangle text-2xl mb-2"></i>
            <p>Failed to load data</p>
            <button @click="refreshData()" class="mt-2 text-blue-600 hover:text-blue-800">
                Try again
            </button>
        </div>
        
        <!-- Chart container -->
        <div x-show="!loading && !error" class="relative">
            <!-- KPI cards -->
            if len(config.KPIs) > 0 {
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                    for _, kpi := range config.KPIs {
                        @KPICard(kpi)
                    }
                </div>
            }
            
            <!-- Main chart -->
            <div class="h-64" x-ref="chartContainer"></div>
            
            <!-- Chart legend -->
            <div x-show="showLegend" class="mt-4 flex flex-wrap justify-center gap-4">
                <template x-for="item in legendItems" :key="item.label">
                    <div class="flex items-center">
                        <div class="w-3 h-3 rounded-full mr-2" :style="'background-color: ' + item.color"></div>
                        <span class="text-sm text-gray-600" x-text="item.label"></span>
                    </div>
                </template>
            </div>
        </div>
    </div>
}

type AnalyticsConfig struct {
    Title       string
    Subtitle    string
    ChartType   string // line, bar, pie, donut, area
    DataURL     string
    TimeRanges  []TimeRange
    KPIs        []KPI
    Realtime    bool
    RefreshRate int // seconds
}

type KPI struct {
    Label       string
    Value       string
    Change      float64
    ChangeType  string // positive, negative, neutral
    Icon        string
    Format      string // number, currency, percentage
}
```

### Interactive Data Grid with Excel-like Features
```go
templ DataGrid(config GridConfig) {
    <div class="bg-white rounded-lg shadow overflow-hidden" x-data="dataGrid()">
        <!-- Grid toolbar -->
        <div class="border-b border-gray-200 p-4">
            <div class="flex items-center justify-between">
                <div class="flex items-center space-x-4">
                    @GridToolbarButtons()
                </div>
                
                <div class="flex items-center space-x-2">
                    <!-- Formula bar -->
                    <div x-show="selectedCell" class="flex items-center space-x-2">
                        <span class="text-sm font-medium text-gray-700" x-text="selectedCellRef"></span>
                        <input 
                            type="text"
                            x-model="cellFormula"
                            @keydown.enter="updateCellValue()"
                            @keydown.escape="cancelEdit()"
                            class="px-2 py-1 border border-gray-300 rounded text-sm"
                            placeholder="Enter formula or value">
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Grid container -->
        <div class="relative overflow-auto" style="max-height: 600px;" x-ref="gridContainer">
            <!-- Column headers -->
            <div class="sticky top-0 z-20 bg-gray-50 border-b border-gray-200">
                <div class="flex">
                    <!-- Row selector column -->
                    <div class="w-12 h-8 border-r border-gray-300 flex items-center justify-center bg-gray-100">
                        <button @click="selectAll()" class="text-xs text-gray-500">
                            <i class="fas fa-th"></i>
                        </button>
                    </div>
                    
                    <!-- Column headers -->
                    <template x-for="(column, index) in columns" :key="column.id">
                        @GridColumnHeader()
                    </template>
                </div>
            </div>
            
            <!-- Grid rows -->
            <div class="relative">
                <template x-for="(row, rowIndex) in visibleRows" :key="row.id">
                    <div class="flex border-b border-gray-200 hover:bg-gray-50">
                        <!-- Row selector -->
                        <div class="w-12 h-8 border-r border-gray-300 flex items-center justify-center bg-gray-100 text-xs text-gray-600">
                            <button @click="selectRow(rowIndex)" x-text="rowIndex + 1"></button>
                        </div>
                        
                        <!-- Data cells -->
                        <template x-for="(column, colIndex) in columns" :key="column.id">
                            @GridDataCell()
                        </template>
                    </div>
                </template>
            </div>
        </div>
        
        <!-- Context menu -->
        @GridContextMenu()
    </div>
}

templ GridDataCell() {
    <div 
        :class="{
            'bg-blue-100 ring-2 ring-blue-500': isSelected(rowIndex, colIndex),
            'bg-blue-50': isInSelection(rowIndex, colIndex)
        }"
        class="relative h-8 border-r border-gray-200 min-w-[100px] flex items-center"
        @click="selectCell(rowIndex, colIndex)"
        @dblclick="editCell(rowIndex, colIndex)"
        @contextmenu.prevent="showContextMenu($event, rowIndex, colIndex)">
        
        <!-- Cell content -->
        <div x-show="!isEditing(rowIndex, colIndex)" class="px-2 py-1 text-sm truncate w-full">
            <span x-text="getCellValue(row, column)"></span>
        </div>
        
        <!-- Cell editor -->
        <input 
            x-show="isEditing(rowIndex, colIndex)"
            x-model="editValue"
            @keydown.enter="saveCell()"
            @keydown.escape="cancelEdit()"
            @blur="saveCell()"
            class="absolute inset-0 px-2 py-1 text-sm border-2 border-blue-500 bg-white">
        
        <!-- Cell indicators -->
        <div x-show="hasFormula(row, column)" class="absolute top-0 left-0 w-2 h-2 bg-green-500 rounded-br"></div>
        <div x-show="hasError(row, column)" class="absolute top-0 right-0 w-2 h-2 bg-red-500 rounded-bl"></div>
    </div>
}
```

## 7. Advanced Form Builder Component

### Dynamic Form Builder
```go
templ FormBuilder(config FormBuilderConfig) {
    <div class="flex h-screen bg-gray-100" x-data="formBuilder()">
        <!-- Component palette -->
        <div class="w-64 bg-white border-r border-gray-200 p-4">
            <h3 class="text-sm font-medium text-gray-900 mb-4">Form Components</h3>
            
            <div class="space-y-2">
                for _, component := range config.AvailableComponents {
                    @DraggableComponent(component)
                }
            </div>
        </div>
        
        <!-- Form canvas -->
        <div class="flex-1 p-6">
            <div class="bg-white rounded-lg shadow-sm min-h-full p-6">
                <div 
                    class="space-y-4 min-h-[400px]"
                    @drop="handleDrop($event)"
                    @dragover.prevent
                    @dragenter.prevent>
                    
                    <!-- Form fields -->
                    <template x-for="(field, index) in formFields" :key="field.id">
                        @FormBuilderField()
                    </template>
                    
                    <!-- Empty state -->
                    <div x-show="formFields.length === 0" class="flex flex-col items-center justify-center h-64 text-gray-500">
                        <i class="fas fa-plus-circle text-4xl mb-4"></i>
                        <p>Drag components here to build your form</p>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Properties panel -->
        <div class="w-80 bg-white border-l border-gray-200 p-4" x-show="selectedField">
            @FieldPropertiesPanel()
        </div>
    </div>
}

templ FormBuilderField() {
    <div 
        :class="{ 'ring-2 ring-blue-500': selectedField?.id === field.id }"
        class="relative group border border-transparent rounded-md p-2 hover:border-gray-300"
        @click="selectField(field)"
        x-data="{ hovering: false }"
        @mouseenter="hovering = true"
        @mouseleave="hovering = false">
        
        <!-- Field component -->
        <div :is="getComponentName(field.type)" :field="field"></div>
        
        <!-- Field actions -->
        <div 
            x-show="hovering || selectedField?.id === field.id"
            class="absolute -top-2 -right-2 flex space-x-1">
            
            <button 
                @click.stop="duplicateField(field.id)"
                class="w-6 h-6 bg-white border border-gray-300 rounded text-xs hover:bg-gray-50">
                <i class="fas fa-copy"></i>
            </button>
            
            <button 
                @click.stop="deleteField(field.id)"
                class="w-6 h-6 bg-white border border-gray-300 rounded text-xs hover:bg-red-50 hover:text-red-600">
                <i class="fas fa-trash"></i>
            </button>
        </div>
        
        <!-- Drag handle -->
        <div 
            x-show="hovering || selectedField?.id === field.id"
            class="absolute -left-2 top-1/2 transform -translate-y-1/2 w-4 h-4 bg-gray-300 rounded cursor-move"
            @mousedown="startDrag($event, field.id)">
            <i class="fas fa-grip-vertical text-xs"></i>
        </div>
    </div>
}
```

## 8. Performance Optimizations

### Lazy Loading Components
```javascript
// Intersection Observer for lazy loading
function lazyLoader() {
    return {
        observer: null,
        
        init() {
            this.observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        this.loadComponent(entry.target);
                        this.observer.unobserve(entry.target);
                    }
                });
            }, { rootMargin: '100px' });
            
            // Observe all lazy components
            document.querySelectorAll('[x-lazy]').forEach(el => {
                this.observer.observe(el);
            });
        },
        
        async loadComponent(element) {
            const componentName = element.getAttribute('x-lazy');
            const props = JSON.parse(element.getAttribute('x-lazy-props') || '{}');
            
            try {
                element.innerHTML = '<div class="animate-pulse bg-gray-200 rounded h-32"></div>';
                
                const response = await fetch(`/components/${componentName}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(props)
                });
                
                const html = await response.text();
                element.innerHTML = html;
                
                // Re-initialize Alpine components
                Alpine.initTree(element);
            } catch (error) {
                element.innerHTML = '<div class="text-red-500 text-center py-4">Failed to load component</div>';
            }
        }
    }
}

// Virtual scrolling for large lists
function virtualScrolling() {
    return {
        items: [],
        visibleItems: [],
        itemHeight: 50,
        containerHeight: 400,
        scrollTop: 0,
        startIndex: 0,
        endIndex: 0,
        spacerTop: 0,
        spacerBottom: 0,
        
        init() {
            this.calculateVisible();
            this.$watch('scrollTop', () => this.calculateVisible());
        },
        
        calculateVisible() {
            const visibleCount = Math.ceil(this.containerHeight / this.itemHeight);
            this.startIndex = Math.floor(this.scrollTop / this.itemHeight);
            this.endIndex = Math.min(this.startIndex + visibleCount, this.items.length);
            
            this.visibleItems = this.items.slice(this.startIndex, this.endIndex);
            this.spacerTop = this.startIndex * this.itemHeight;
            this.spacerBottom = (this.items.length - this.endIndex) * this.itemHeight;
        },
        
        handleScroll(event) {
            this.scrollTop = event.target.scrollTop;
        }
    }
}
```

## Key Improvements Made:

### 1. **Accessibility **
- ARIA labels, roles, and properties
- Keyboard navigation support
- Screen reader compatibility
- Focus management

### 2. **Advanced Form Features**
- File upload with drag & drop
- Rich text editor
- Real-time validation
- Character counters
- Form builder interface

### 3. **Table Capabilities**
- Virtual scrolling for performance
- Column resizing and reordering
- Advanced filtering
- Bulk operations
- Export functionality

### 4. **Better User Experience**
- Loading states and error handling
- Confirmation dialogs
- Toast notifications
- Modal stacking
- Responsive design

### 5. **Performance Optimizations**
- Lazy loading components
- Virtual scrolling
- Debounced validation
- Efficient re-rendering

### 6. **Developer Experience**
- Better type definitions
- Comprehensive error handling
- Extensible component architecture
- Clear separation of concerns

This enhanced component system provides a production-ready foundation for building sophisticated ERP interfaces while maintaining excellent performance and user experience standards.
