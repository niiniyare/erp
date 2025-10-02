# TypeScript Hooks API Reference

<!-- LLM-CONTEXT-START -->
**FILE PURPOSE**: Complete API reference for all TypeScript hooks in the ERP system
**SCOPE**: Hook interfaces, methods, events, and usage examples
**TARGET AUDIENCE**: Developers implementing and extending hooks
**RELATED FILES**: `typescript-hooks-integration.md` (patterns), `component-development-guide.md` (workflow)
<!-- LLM-CONTEXT-END -->

## Overview

This reference documents all available TypeScript hooks, their interfaces, methods, and integration patterns with Alpine.js and Templ components.

<!-- LLM-HOOK-CATALOG-START -->
**AVAILABLE HOOKS:**
- [useApp](#useapp-hook) - Global application state and configuration
- [useDataTable](#usedatatable-hook) - Advanced data table functionality
- [useBreadCrumbs](#usebreadcrumbs-hook) - Navigation breadcrumb management
- [useFormValidation](#useformvalidation-hook) - Form validation and submission
- [useModal](#usemodal-hook) - Modal dialog management
- [useNavBar](#usenavbar-hook) - Navigation state management
- [useTreeView](#usetreeview-hook) - Hierarchical data display
<!-- LLM-HOOK-CATALOG-END -->

---

## Global Configuration

### Window Object Extensions

<!-- LLM-GLOBAL-CONFIG-START -->
```typescript
// Global configuration available to all hooks
interface Window {
    APP_CONFIG: {
        api: {
            baseURL: string;
            timeout: number;
            headers: Record<string, string>;
        };
        app: {
            name: string;
            version: string;
            description: string;
        };
        theme: string;
        features: Record<string, boolean>;
        menu: Array<{
            label: string;
            icon: string;
            url: string;
            children: Array<{
                label: string;
                url: string;
                children: any[];
            }>;
        }>;
        defaultTenant: string | null;
        pagination: {
            pageSize: number;
            showSizeChanger: boolean;
            showQuickJumper: boolean;
        };
    };
    
    API: {
        getHeaders: (tenantId?: string | null) => Record<string, string>;
        url: (path: string) => string;
        handleResponse: (response: Response) => Promise<any>;
        fetch: (url: string, options?: RequestInit) => Promise<any>;
    };
    
    // Hook factories
    useDataTable: (config?: any) => any;
    createConsoleDataTable: (config?: any) => any;
    createWorkspaceDataTable: (config?: any) => any;
    createPortalDataTable: (config?: any) => any;
}
```
<!-- LLM-GLOBAL-CONFIG-END -->

---

## useApp Hook

### Overview
Manages global application state, configuration, and cross-component communication.

### Interface

<!-- LLM-USEAPP-INTERFACE-START -->
```typescript
interface AppConfig {
    theme?: 'light' | 'dark' | 'auto';
    language?: string;
    tenantId?: string;
    userId?: string;
    notifications?: boolean;
}

interface AppStore {
    // Configuration
    theme: string;
    language: string;
    tenantId: string | null;
    userId: string | null;
    
    // Application state
    isLoading: boolean;
    notifications: Notification[];
    errors: Error[];
    
    // Feature flags
    features: Record<string, boolean>;
    
    // Methods
    init(): void;
    setTheme(theme: string): void;
    setLanguage(language: string): void;
    addNotification(notification: Notification): void;
    removeNotification(id: string): void;
    addError(error: Error): void;
    clearErrors(): void;
    isFeatureEnabled(feature: string): boolean;
    
    // Computed properties
    readonly isDarkMode: boolean;
    readonly hasNotifications: boolean;
    readonly hasErrors: boolean;
}

function useApp(config?: AppConfig): AppStore;
```
<!-- LLM-USEAPP-INTERFACE-END -->

### Usage Example

<!-- LLM-USEAPP-EXAMPLE-START -->
```go
// Templ component using app hook
templ AppLayout(props AppLayoutProps) {
    <div 
        x-data="useApp({
            theme: @json(props.Theme),
            tenantId: @json(props.TenantID),
        })"
        x-init="init()"
        :class="{ 'dark': isDarkMode }"
    >
        <!-- Theme toggle -->
        <button @click="setTheme(isDarkMode ? 'light' : 'dark')">
            <i :class="isDarkMode ? 'fas fa-sun' : 'fas fa-moon'"></i>
        </button>
        
        <!-- Notifications -->
        <div x-show="hasNotifications" class="notifications">
            <template x-for="notification in notifications" :key="notification.id">
                <div class="notification" x-text="notification.message">
                    <button @click="removeNotification(notification.id)">×</button>
                </div>
            </template>
        </div>
        
        <!-- Main content -->
        <main>
            { children... }
        </main>
    </div>
}
```
<!-- LLM-USEAPP-EXAMPLE-END -->

---

## useDataTable Hook

### Overview
Provides comprehensive data table functionality with search, filtering, pagination, sorting, and bulk actions.

### Interface

<!-- LLM-USEDATATABLE-INTERFACE-START -->
```typescript
interface DataTableConfig {
    rows?: any[];
    searchColumns?: string[];
    defaultSort?: {
        column?: string;
        direction?: 'asc' | 'desc';
    };
    pageSize?: number;
    bulkActions?: BulkAction[];
    permissions?: Record<string, boolean>;
    serviceContext?: string;
    clearSelectionOnSearch?: boolean;
    searchDebounce?: number;
    enableExport?: boolean;
    requestTimeout?: number;
}

interface DataTableStore {
    // State Management
    rows: any[];
    filteredRows: any[];
    paginatedRows: any[];
    selectedRows: string[];
    
    // Search & Filter
    searchQuery: string;
    searchColumns: string[];
    
    // Sorting
    sortColumn: string;
    sortDirection: 'asc' | 'desc';
    
    // Pagination
    currentPage: number;
    pageSize: number;
    totalRows: number;
    totalPages: number;
    
    // Configuration
    bulkActions: BulkAction[];
    permissions: Record<string, boolean>;
    serviceContext: string;
    
    // Loading states
    isLoading: boolean;
    isBulkActionInProgress: boolean;
    
    // Methods
    init(): void;
    destroy(): void;
    handleSearch(): void;
    handleSearchImmediate(): void;
    sortBy(column: string): void;
    nextPage(): void;
    previousPage(): void;
    goToPage(page: number): void;
    toggleRowSelection(rowId: string): void;
    toggleSelectAll(): void;
    clearSelection(): void;
    executeBulkAction(actionId: string): Promise<void>;
    exportData(format: 'csv' | 'json'): void;
    refreshData(): void;
    
    // Computed properties
    readonly selectedCount: number;
    readonly allRowsSelected: boolean;
    readonly someRowsSelected: boolean;
    readonly paginationInfo: PaginationInfo;
    readonly canGoNext: boolean;
    readonly canGoPrevious: boolean;
}

function useDataTable(config?: DataTableConfig): DataTableStore;
```
<!-- LLM-USEDATATABLE-INTERFACE-END -->

### Service-Specific Factories

<!-- LLM-DATATABLE-FACTORIES-START -->
```typescript
// Service-specific data table factories
function createConsoleDataTable(config?: DataTableConfig): DataTableStore;
function createWorkspaceDataTable(config?: DataTableConfig): DataTableStore;
function createPortalDataTable(config?: DataTableConfig): DataTableStore;

// Usage in Templ
templ ConsoleUsersTable(users []User) {
    <div x-data="createConsoleDataTable({
        rows: @json(convertUsersToTableRows(users)),
        bulkActions: @json(getConsoleBulkActions()),
        serviceContext: 'console'
    })">
        <!-- Table implementation -->
    </div>
}
```
<!-- LLM-DATATABLE-FACTORIES-END -->

### Usage Example

<!-- LLM-DATATABLE-EXAMPLE-START -->
```go
templ UsersDataTable(props DataTableProps) {
    <div 
        class="erp-datatable"
        x-data="useDataTable({
            rows: @json(props.Rows),
            searchColumns: ['name', 'email', 'department'],
            pageSize: 20,
            bulkActions: @json(props.BulkActions),
            serviceContext: @json(props.ServiceContext)
        })"
        x-init="init()"
    >
        <!-- Search -->
        <div class="mb-4">
            <input 
                type="text"
                x-model="searchQuery"
                @input="handleSearch()"
                placeholder="Search users..."
                class="block w-full p-2 border border-gray-300 rounded-lg">
        </div>
        
        <!-- Bulk Actions -->
        <div x-show="selectedCount > 0" class="mb-4">
            <template x-for="action in bulkActions" :key="action.id">
                <button 
                    @click="executeBulkAction(action.id)"
                    :disabled="isBulkActionInProgress"
                    class="mr-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
                    <span x-text="action.text"></span>
                    <span x-show="isBulkActionInProgress" class="ml-2">
                        <i class="fas fa-spinner fa-spin"></i>
                    </span>
                </button>
            </template>
        </div>
        
        <!-- Table -->
        <div class="overflow-x-auto">
            <table class="w-full text-sm text-left text-gray-500">
                <thead>
                    <tr>
                        <th class="px-6 py-3">
                            <input 
                                type="checkbox" 
                                @change="toggleSelectAll()"
                                :checked="allRowsSelected"
                                :indeterminate="someRowsSelected && !allRowsSelected">
                        </th>
                        <th class="px-6 py-3 cursor-pointer" @click="sortBy('name')">
                            Name
                            <i x-show="sortColumn === 'name'" 
                               :class="sortDirection === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"></i>
                        </th>
                        <th class="px-6 py-3 cursor-pointer" @click="sortBy('email')">
                            Email
                            <i x-show="sortColumn === 'email'" 
                               :class="sortDirection === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"></i>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <template x-for="row in paginatedRows" :key="row.id">
                        <tr class="bg-white border-b hover:bg-gray-50">
                            <td class="px-6 py-4">
                                <input 
                                    type="checkbox" 
                                    :value="row.id"
                                    x-model="selectedRows">
                            </td>
                            <td class="px-6 py-4" x-text="row.name"></td>
                            <td class="px-6 py-4" x-text="row.email"></td>
                        </tr>
                    </template>
                </tbody>
            </table>
        </div>
        
        <!-- Pagination -->
        <div class="flex justify-between items-center mt-4">
            <div>
                <span x-text="`Showing ${paginationInfo.start} to ${paginationInfo.end} of ${totalRows} entries`"></span>
            </div>
            <div class="flex space-x-2">
                <button 
                    @click="previousPage()" 
                    :disabled="!canGoPrevious"
                    class="px-3 py-2 text-sm bg-blue-500 text-white rounded disabled:opacity-50">
                    Previous
                </button>
                <button 
                    @click="nextPage()" 
                    :disabled="!canGoNext"
                    class="px-3 py-2 text-sm bg-blue-500 text-white rounded disabled:opacity-50">
                    Next
                </button>
            </div>
        </div>
    </div>
}
```
<!-- LLM-DATATABLE-EXAMPLE-END -->

---

## useFormValidation Hook

### Overview
Provides comprehensive form validation with real-time feedback, async validation, and multi-step form support.

### Interface

<!-- LLM-USEFORMVALIDATION-INTERFACE-START -->
```typescript
interface ValidationRule {
    required?: boolean;
    requiredMessage?: string;
    minLength?: number;
    minLengthMessage?: string;
    maxLength?: number;
    maxLengthMessage?: string;
    pattern?: string | RegExp;
    patternMessage?: string;
    email?: boolean;
    emailMessage?: string;
    custom?: (value: any, formData: Record<string, any>) => string | null;
    async?: (value: any, formData: Record<string, any>, signal: AbortSignal) => Promise<{ error?: string } | null>;
}

interface FormValidationConfig {
    initialData?: Record<string, any>;
    rules?: Record<string, ValidationRule>;
    realTimeValidation?: boolean;
    validateOnBlur?: boolean;
    validateOnInput?: boolean;
    debounce?: number;
    csrfTokenSelector?: string;
    toastStore?: string;
}

interface FormValidationStore {
    // State
    formData: Record<string, any>;
    validationErrors: Record<string, string[]>;
    validationRules: Record<string, ValidationRule>;
    
    // Status tracking
    isValid: boolean;
    isDirty: boolean;
    isSubmitting: boolean;
    submitAttempted: boolean;
    
    // Configuration
    realTimeValidation: boolean;
    validateOnBlur: boolean;
    validateOnInput: boolean;
    validationDebounce: number;
    csrfTokenSelector: string;
    toastStoreName: string;
    
    // Methods
    init(): void;
    destroy(): void;
    validateForm(): boolean;
    validateField(field: string, value: any, rules: ValidationRule): string[];
    submitForm(endpoint: string, options?: SubmissionOptions): Promise<any>;
    prepareFormData(): FormData;
    handleFieldInput(field: string, value: any): void;
    handleFieldBlur(field: string): void;
    validateSingleField(field: string): void;
    reset(newInitialData?: Record<string, any>): void;
    
    // Utility methods
    isEmpty(value: any): boolean;
    isValidEmail(email: string): boolean;
    getFieldLabel(field: string): string;
    focusFirstError(): void;
    markDirty(): void;
    
    // Computed properties
    readonly hasErrors: boolean;
    readonly errorCount: number;
    readonly isFormReady: boolean;
}

function useFormValidation(config?: FormValidationConfig): FormValidationStore;
```
<!-- LLM-USEFORMVALIDATION-INTERFACE-END -->

### Usage Example

<!-- LLM-FORMVALIDATION-EXAMPLE-START -->
```go
templ UserForm(props UserFormProps) {
    <form 
        x-data="useFormValidation({
            initialData: @json(props.InitialData),
            rules: {
                name: {
                    required: true,
                    requiredMessage: 'Name is required',
                    minLength: 2,
                    minLengthMessage: 'Name must be at least 2 characters'
                },
                email: {
                    required: true,
                    email: true,
                    emailMessage: 'Please enter a valid email address',
                    async: async (value) => {
                        const response = await fetch('/api/validate-email', {
                            method: 'POST',
                            body: JSON.stringify({ email: value })
                        });
                        const result = await response.json();
                        return result.exists ? { error: 'Email already exists' } : null;
                    }
                },
                password: {
                    required: true,
                    minLength: 8,
                    pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
                    patternMessage: 'Password must contain uppercase, lowercase, and number'
                }
            },
            realTimeValidation: true,
            validateOnBlur: true
        })"
        x-init="init()"
        @submit.prevent="submitForm('/api/users')"
    >
        <!-- Name Field -->
        <div class="mb-4">
            <label for="name" class="block text-sm font-medium text-gray-700">Name</label>
            <input 
                type="text"
                id="name"
                name="name"
                x-model="formData.name"
                @input="handleFieldInput('name', $event.target.value)"
                @blur="handleFieldBlur('name')"
                :class="validationErrors.name ? 'border-red-500' : 'border-gray-300'"
                class="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
            
            <template x-for="error in (validationErrors.name || [])" :key="error">
                <p class="mt-1 text-sm text-red-600" x-text="error"></p>
            </template>
        </div>
        
        <!-- Email Field -->
        <div class="mb-4">
            <label for="email" class="block text-sm font-medium text-gray-700">Email</label>
            <input 
                type="email"
                id="email"
                name="email"
                x-model="formData.email"
                @input="handleFieldInput('email', $event.target.value)"
                @blur="handleFieldBlur('email')"
                :class="validationErrors.email ? 'border-red-500' : 'border-gray-300'"
                class="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
            
            <template x-for="error in (validationErrors.email || [])" :key="error">
                <p class="mt-1 text-sm text-red-600" x-text="error"></p>
            </template>
        </div>
        
        <!-- Password Field -->
        <div class="mb-4">
            <label for="password" class="block text-sm font-medium text-gray-700">Password</label>
            <input 
                type="password"
                id="password"
                name="password"
                x-model="formData.password"
                @input="handleFieldInput('password', $event.target.value)"
                @blur="handleFieldBlur('password')"
                :class="validationErrors.password ? 'border-red-500' : 'border-gray-300'"
                class="mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
            
            <template x-for="error in (validationErrors.password || [])" :key="error">
                <p class="mt-1 text-sm text-red-600" x-text="error"></p>
            </template>
        </div>
        
        <!-- Submit Button -->
        <div class="flex justify-end">
            <button 
                type="submit"
                :disabled="!isValid || isSubmitting"
                class="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed">
                <span x-show="!isSubmitting">Save User</span>
                <span x-show="isSubmitting">
                    <i class="fas fa-spinner fa-spin mr-2"></i>
                    Saving...
                </span>
            </button>
        </div>
        
        <!-- Form Summary -->
        <div x-show="submitAttempted && hasErrors" class="mt-4 p-4 bg-red-50 border border-red-200 rounded-md">
            <h4 class="text-sm font-medium text-red-800">Please fix the following errors:</h4>
            <ul class="mt-2 text-sm text-red-700">
                <template x-for="(errors, field) in validationErrors" :key="field">
                    <template x-for="error in errors" :key="error">
                        <li x-text="error"></li>
                    </template>
                </template>
            </ul>
        </div>
    </form>
}
```
<!-- LLM-FORMVALIDATION-EXAMPLE-END -->

---

## useModal Hook

### Overview
Manages modal dialog state, accessibility, and user interactions.

### Interface

<!-- LLM-USEMODAL-INTERFACE-START -->
```typescript
interface ModalConfig {
    backdrop?: 'static' | 'dismissible';
    keyboard?: boolean;
    focus?: boolean;
    size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
    position?: 'center' | 'top';
    closable?: boolean;
    onShow?: () => void;
    onHide?: () => void;
    onShown?: () => void;
    onHidden?: () => void;
}

interface ModalStore {
    // State
    isOpen: boolean;
    isAnimating: boolean;
    
    // Configuration
    backdrop: 'static' | 'dismissible';
    keyboard: boolean;
    focus: boolean;
    size: string;
    position: string;
    closable: boolean;
    
    // Methods
    init(): void;
    destroy(): void;
    open(): void;
    close(): void;
    toggle(): void;
    
    // Internal methods
    setupEventListeners(): void;
    setupFocusTrap(): void;
    handleBackdropClick(event: Event): void;
    handleKeydown(event: KeyboardEvent): void;
    
    // Computed properties
    readonly isCurrentModal: boolean;
    readonly modalClasses: string;
    readonly backdropClasses: string;
}

function useModal(config?: ModalConfig): ModalStore;
```
<!-- LLM-USEMODAL-INTERFACE-END -->

### Usage Example

<!-- LLM-MODAL-EXAMPLE-START -->
```go
templ UserModal(props UserModalProps) {
    <div 
        x-data="useModal({
            backdrop: 'dismissible',
            keyboard: true,
            focus: true,
            size: 'lg',
            onShow: () => console.log('Modal opened'),
            onHide: () => console.log('Modal closed')
        })"
        x-init="init()"
    >
        <!-- Trigger Button -->
        <button 
            @click="open()"
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
            Open User Modal
        </button>
        
        <!-- Modal -->
        <div 
            x-show="isOpen"
            x-transition:enter="ease-out duration-300"
            x-transition:enter-start="opacity-0"
            x-transition:enter-end="opacity-100"
            x-transition:leave="ease-in duration-200"
            x-transition:leave-start="opacity-100"
            x-transition:leave-end="opacity-0"
            class="fixed inset-0 z-50 overflow-y-auto"
            style="display: none;"
        >
            <!-- Backdrop -->
            <div 
                class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
                @click="handleBackdropClick($event)"
            ></div>
            
            <!-- Modal Dialog -->
            <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
                <div 
                    x-show="isOpen"
                    x-transition:enter="ease-out duration-300"
                    x-transition:enter-start="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    x-transition:enter-end="opacity-100 translate-y-0 sm:scale-100"
                    x-transition:leave="ease-in duration-200"
                    x-transition:leave-start="opacity-100 translate-y-0 sm:scale-100"
                    x-transition:leave-end="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    :class="modalClasses"
                    class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg"
                >
                    <!-- Modal Header -->
                    <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                        <div class="flex items-center justify-between">
                            <h3 class="text-lg font-medium leading-6 text-gray-900">
                                { props.Title }
                            </h3>
                            <button 
                                x-show="closable"
                                @click="close()"
                                class="text-gray-400 hover:text-gray-600">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                    </div>
                    
                    <!-- Modal Content -->
                    <div class="bg-white px-4 pb-4 sm:p-6">
                        { children... }
                    </div>
                    
                    <!-- Modal Footer -->
                    <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
                        <button 
                            @click="close()"
                            class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">
                            Cancel
                        </button>
                        <button 
                            type="button"
                            class="inline-flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-blue-700 sm:ml-3 sm:w-auto sm:text-sm">
                            Save
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
}
```
<!-- LLM-MODAL-EXAMPLE-END -->

---

## Common Events

### Hook Events

<!-- LLM-HOOK-EVENTS-START -->
All hooks emit standardized events that can be listened to across components:

```typescript
// Data Table Events
'datatable:search'         // { query: string }
'datatable:sort'          // { column: string, direction: 'asc' | 'desc' }
'datatable:select'        // { selectedRows: string[] }
'datatable:bulk-action'   // { action: string, rows: string[] }

// Form Events
'form:validate'           // { field: string, isValid: boolean }
'form:submit'            // { data: FormData, isValid: boolean }
'form:error'             // { field?: string, error: string }
'form:reset'             // { }

// Modal Events
'modal:open'             // { modalId: string }
'modal:close'            // { modalId: string }
'modal:backdrop-click'   // { modalId: string }

// Navigation Events
'nav:change'             // { currentPath: string, previousPath: string }
'breadcrumb:update'      // { breadcrumbs: BreadcrumbItem[] }

// Application Events
'app:theme-change'       // { theme: string }
'app:language-change'    // { language: string }
'app:error'              // { error: Error }
'app:notification'       // { notification: Notification }
```

### Event Listening Example

```go
templ PageWithEventHandling() {
    <div 
        @datatable:select="handleTableSelection($event.detail)"
        @form:submit="handleFormSubmit($event.detail)"
        @modal:open="handleModalOpen($event.detail)"
    >
        <!-- Components that emit events -->
        @DataTable(tableProps)
        @UserForm(formProps)
        @Modal(modalProps)
    </div>
    
    <script>
    function handleTableSelection(detail) {
        console.log('Selected rows:', detail.selectedRows);
        // Update other components based on selection
    }
    
    function handleFormSubmit(detail) {
        if (detail.isValid) {
            // Close modal, refresh table, etc.
        }
    }
    
    function handleModalOpen(detail) {
        // Pause auto-refresh, disable keyboard shortcuts, etc.
    }
    </script>
}
```
<!-- LLM-HOOK-EVENTS-END -->

---

## Best Practices

### Performance Optimization

<!-- LLM-PERFORMANCE-BEST-PRACTICES-START -->
```typescript
// 1. Use debouncing for expensive operations
function useOptimizedSearch() {
    let searchTimeout: any;
    
    return {
        search(query: string): void {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                this.performSearch(query);
            }, 300);
        }
    };
}

// 2. Implement virtual scrolling for large datasets
function useVirtualDataTable(config: DataTableConfig) {
    return {
        get visibleRows(): any[] {
            const start = this.virtualScrollTop;
            const end = start + this.virtualViewportSize;
            return this.filteredRows.slice(start, end);
        }
    };
}

// 3. Use computed properties instead of methods
interface OptimizedStore {
    readonly filteredData: any[]; // Computed property
    // Instead of getFilteredData() method
}

// 4. Cleanup resources in destroy methods
function useResourceAwareHook() {
    return {
        destroy(): void {
            // Clear timers
            clearTimeout(this._timer);
            
            // Remove event listeners
            document.removeEventListener('click', this._clickHandler);
            
            // Cancel ongoing requests
            this._abortController?.abort();
        }
    };
}
```
<!-- LLM-PERFORMANCE-BEST-PRACTICES-END -->

### Error Handling

<!-- LLM-ERROR-HANDLING-BEST-PRACTICES-START -->
```typescript
// 1. Comprehensive error state management
interface ErrorHandlingStore {
    hasError: boolean;
    errorMessage: string | null;
    errorCode?: string;
    
    handleError(error: Error): void {
        this.hasError = true;
        this.errorMessage = error.message;
        this.errorCode = (error as any).code;
        
        // Log to console in development
        if (process.env.NODE_ENV === 'development') {
            console.error('Hook Error:', error);
        }
        
        // Emit error event for global handling
        (this as any).$dispatch('hook:error', { error });
    };
    
    clearError(): void {
        this.hasError = false;
        this.errorMessage = null;
        this.errorCode = undefined;
    };
}

// 2. Graceful degradation
function useResilientDataTable(config: DataTableConfig) {
    return {
        async loadData(): Promise<void> {
            try {
                const data = await this.fetchData();
                this.rows = data;
            } catch (error) {
                // Fallback to cached data or empty state
                this.rows = this.getCachedData() || [];
                this.handleError(error);
            }
        }
    };
}

// 3. Retry mechanisms
function useRetryableOperation() {
    return {
        async executeWithRetry(operation: () => Promise<any>, maxRetries = 3): Promise<any> {
            let lastError: Error;
            
            for (let attempt = 1; attempt <= maxRetries; attempt++) {
                try {
                    return await operation();
                } catch (error) {
                    lastError = error as Error;
                    
                    if (attempt < maxRetries) {
                        // Exponential backoff
                        await new Promise(resolve => 
                            setTimeout(resolve, Math.pow(2, attempt) * 1000)
                        );
                    }
                }
            }
            
            throw lastError!;
        }
    };
}
```
<!-- LLM-ERROR-HANDLING-BEST-PRACTICES-END -->

### Type Safety

<!-- LLM-TYPE-SAFETY-BEST-PRACTICES-START -->
```typescript
// 1. Use specific types instead of 'any'
interface SpecificDataRow {
    id: string;
    name: string;
    email: string;
    createdAt: Date;
}

interface TypedDataTableConfig<T = SpecificDataRow> {
    rows?: T[];
    onSelect?: (row: T) => void;
}

// 2. Use union types for constrained values
interface ComponentConfig {
    variant: 'primary' | 'secondary' | 'success' | 'danger';
    size: 'sm' | 'md' | 'lg';
    position: 'top' | 'bottom' | 'left' | 'right';
}

// 3. Use type guards for runtime safety
function isValidUser(obj: any): obj is User {
    return obj && 
           typeof obj.id === 'string' &&
           typeof obj.name === 'string' &&
           typeof obj.email === 'string';
}

// 4. Use generic types for reusability
function createTypedHook<TConfig, TStore>(
    hookFactory: (config: TConfig) => TStore
): (config: TConfig) => TStore {
    return hookFactory;
}
```
<!-- LLM-TYPE-SAFETY-BEST-PRACTICES-END -->

---

## References

<!-- LLM-REFERENCES-START -->
**RELATED DOCUMENTATION:**
- [TypeScript Hooks Integration](../patterns/typescript-hooks-integration.md) - Hook development patterns
- [Component Development Guide](../guides/component-development-guide.md) - Complete development workflow
- [Getting Started](../getting-started.md) - Setup and configuration
- [Architecture](../fundamentals/architecture.md) - System design principles

**EXTERNAL REFERENCES:**
- [Alpine.js Documentation](https://alpinejs.dev/) - Reactive JavaScript framework
- [TypeScript Handbook](https://www.typescriptlang.org/docs/) - TypeScript language features
- [Templ Guide](https://templ.guide/) - Go templating language
- [Flowbite Components](https://flowbite.com/docs/components/) - UI component library

**TESTING RESOURCES:**
- [Jest Documentation](https://jestjs.io/) - JavaScript testing framework
- [Testing Library](https://testing-library.com/) - Simple and complete testing utilities
- [Go Testing](https://golang.org/pkg/testing/) - Go unit testing
<!-- LLM-REFERENCES-END -->

<!-- LLM-METADATA-START -->
**METADATA FOR AI ASSISTANTS:**
- File Type: API Reference Documentation
- Scope: Complete hook interfaces, methods, events, and usage patterns
- Target Audience: Developers implementing and extending TypeScript hooks
- Complexity: Intermediate to Advanced
- Focus: Type-safe hook development and integration
- Dependencies: TypeScript, Alpine.js, Templ, Flowbite, TailwindCSS
- Last Updated: December 2024
<!-- LLM-METADATA-END -->