# TypeScript Hooks Integration with Alpine.js and Templ

**FILE PURPOSE**: Complete integration guide for TypeScript hooks with Alpine.js, Templ, and Flowbite components
**SCOPE**: Hook architecture, Alpine.js patterns, TypeScript integration, component development workflow
**TARGET AUDIENCE**: Full-stack developers building ERP interfaces with type-safe client-side interactions
**RELATED FILES**: `fundamentals/architecture.md` (system design), `components/elements.md` (basic components), `getting-started.md` (setup)

## Overview

This guide documents the integration of TypeScript hooks with Alpine.js for type-safe, reactive client-side behavior in Templ-based ERP interfaces. The hooks system provides reusable, composable functionality while maintaining the server-first architecture principles.

**INTEGRATION PRINCIPLES:**
- **Type Safety First** - TypeScript interfaces for all hook configurations and state
- **Server-First Architecture** - Hooks enhance server-rendered Templ components
- **Progressive Enhancement** - Works without JavaScript, enhanced with Alpine.js
- **Composition Over Inheritance** - Reusable hooks for complex UI behaviors
- **Performance Optimized** - Minimal JavaScript footprint with efficient state management

---

## Architecture Overview

### Hook Integration Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    DEVELOPMENT FLOW                        │
│                                                             │
│  1. Define TypeScript Hook                                  │
│     ├── Interface definitions                               │
│     ├── Hook implementation                                 │
│     └── Export to window global                             │
│                          │                                  │
│                          ▼                                  │
│  2. Templ Component Integration                             │
│     ├── x-data="hookName(config)"                          │
│     ├── x-init="init()"                                     │
│     └── Alpine.js directives                                │
│                          │                                  │
│                          ▼                                  │
│  3. Server-Side Rendering                                   │
│     ├── Go handlers pass data                               │
│     ├── Templ generates HTML                                │
│     └── Alpine.js hydrates client-side                     │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack Integration

**TECHNOLOGY LAYERS:**

#### 1. TypeScript Layer (`/web/hooks/`)
- **Hook definitions** with comprehensive type interfaces
- **Global window exports** for Templ component access
- **Type-safe configuration** objects and state management
- **Reusable business logic** abstracted from UI concerns

#### 2. Alpine.js Layer (Client-Side)
- **Reactive state management** with TypeScript backing
- **Event handling** and DOM manipulation
- **Component lifecycle** management (init, destroy)
- **Store-based** cross-component communication

#### 3. Templ Layer (Server-Side)
- **Type-safe HTML generation** with Go structs
- **Server-side data** passed to client-side hooks
- **Progressive enhancement** with Alpine.js directives
- **HTMX integration** for server interactions

#### 4. Flowbite + TailwindCSS Layer
- **Consistent styling** across all components
- **Responsive design** patterns and utilities
- **Interactive elements** (modals, dropdowns, forms)
- **Design system** compliance and theming

---

## Hook Development Patterns

### 1. Hook File Structure

**STANDARD HOOK FILE STRUCTURE:**

```typescript
/**
 * Feature Hook (TypeScript)
 * 
 * Extracted from internal/ui/components/feature/*.templ
 * Provides reusable Feature functionality for all ERP services
 * 
 * @version 2.0.0
 */

/// <reference path="./types.d.ts" />

// Types and Interfaces
interface FeatureConfig {
    // Configuration properties
    option1?: string;
    option2?: boolean;
    onEvent?: (data: any) => void;
}

interface FeatureStore {
    // State properties
    data: any[];
    loading: boolean;
    error: string | null;
    
    // Methods
    init(): void;
    handleAction(): void;
    destroy(): void;
    
    // Computed properties
    readonly isValid: boolean;
    readonly totalCount: number;
}

// Hook Implementation
function useFeature(config: FeatureConfig = {}): FeatureStore {
    return {
        // State initialization
        data: config.data || [],
        loading: false,
        error: null,
        
        // Lifecycle methods
        init(): void {
            // Setup watchers, event listeners, etc.
            if (typeof (this as any).$watch === 'function') {
                (this as any).$watch('data', () => {
                    this.handleDataChange();
                });
            }
        },
        
        // Business logic methods
        handleAction(): void {
            // Implementation
        },
        
        // Cleanup
        destroy(): void {
            // Cleanup listeners, timers, etc.
        },
        
        // Computed properties
        get isValid(): boolean {
            return this.data.length > 0 && !this.error;
        },
        
        get totalCount(): number {
            return this.data.length;
        }
    };
}

// Export to global scope
(window as any).useFeature = useFeature;

// Export for type checking
export { useFeature, FeatureConfig, FeatureStore };
```

### 2. Component Integration Patterns

**TEMPL COMPONENT WITH TYPESCRIPT HOOK:**

```go
// Go Templ Component
package components

type DataTableProps struct {
    ID          string              `json:"id"`
    Columns     []DataTableColumn   `json:"columns"`
    Rows        []DataTableRow      `json:"rows"`
    Selectable  bool                `json:"selectable"`
    Searchable  bool                `json:"searchable"`
    PageSize    int                 `json:"page_size"`
}

templ DataTable(props DataTableProps) {
    <div 
        id={ props.ID }
        class="erp-datatable"
        x-data={ "useDataTable(" + toJSON(props) + ")" }
        x-init="init()"
    >
        <!-- Search Input -->
        if props.Searchable {
            <div class="mb-4">
                <input 
                    type="text" 
                    x-model="searchQuery"
                    @input="handleSearch()"
                    placeholder="Search..."
                    class="block w-full p-2 border border-gray-300 rounded-lg">
            </div>
        }
        
        <!-- Data Table -->
        <div class="overflow-x-auto">
            <table class="w-full text-sm text-left text-gray-500">
                <thead class="text-xs text-gray-700 uppercase bg-gray-50">
                    <tr>
                        if props.Selectable {
                            <th class="px-6 py-3">
                                <input 
                                    type="checkbox" 
                                    @change="toggleSelectAll()"
                                    :checked="allRowsSelected">
                            </th>
                        }
                        for _, column := range props.Columns {
                            <th class="px-6 py-3" @click={ "sortBy('" + column.Key + "')" }>
                                { column.Title }
                                <span x-show={ "sortColumn === '" + column.Key + "'" }>
                                    <i x-show="sortDirection === 'asc'" class="fas fa-sort-up"></i>
                                    <i x-show="sortDirection === 'desc'" class="fas fa-sort-down"></i>
                                </span>
                            </th>
                        }
                    </tr>
                </thead>
                <tbody>
                    <template x-for="row in paginatedRows" :key="row.id">
                        <tr class="bg-white border-b hover:bg-gray-50">
                            if props.Selectable {
                                <td class="px-6 py-4">
                                    <input 
                                        type="checkbox" 
                                        :value="row.id"
                                        x-model="selectedRows"
                                        @change="updateSelection()">
                                </td>
                            }
                            for _, column := range props.Columns {
                                <td class="px-6 py-4" x-text={ "row." + column.Key }></td>
                            }
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
                    :disabled="currentPage === 1"
                    class="px-3 py-2 text-sm bg-blue-500 text-white rounded disabled:opacity-50">
                    Previous
                </button>
                <button 
                    @click="nextPage()" 
                    :disabled="currentPage === totalPages"
                    class="px-3 py-2 text-sm bg-blue-500 text-white rounded disabled:opacity-50">
                    Next
                </button>
            </div>
        </div>
    </div>
}

// Helper function to convert props to JSON
func toJSON(v interface{}) string {
    b, _ := json.Marshal(v)
    return string(b)
}
```

### 3. Advanced Hook Patterns

**ADVANCED HOOK COMPOSITION PATTERNS:**

#### A. Multi-Hook Composition
```typescript
// Combining multiple hooks for complex functionality
function useAdvancedDataTable(config: DataTableConfig = {}): AdvancedDataTableStore {
    // Compose multiple hooks
    const dataTable = useDataTable(config);
    const formValidation = useFormValidation({
        rules: config.validationRules
    });
    const modal = useModal({
        backdrop: 'static'
    });
    
    return {
        // Merge functionality from multiple hooks
        ...dataTable,
        ...formValidation,
        ...modal,
        
        // Override or extend methods
        init(): void {
            // Initialize all composed hooks
            dataTable.init();
            formValidation.init();
            modal.init();
            
            // Add custom initialization
            this.setupAdvancedFeatures();
        },
        
        // Custom methods that use composed functionality
        editRow(rowId: string): void {
            const row = this.getRowById(rowId);
            this.setFormData(row);
            this.openModal();
        },
        
        saveRow(): Promise<void> {
            if (this.validateForm()) {
                const data = this.getFormData();
                await this.updateRow(data);
                this.closeModal();
                this.refreshData();
            }
        }
    };
}
```

#### B. Plugin Architecture
```typescript
// Plugin system for extensible hooks
interface HookPlugin {
    name: string;
    install(hook: any): void;
    uninstall?(hook: any): void;
}

class DataTablePluginManager {
    private plugins: Map<string, HookPlugin> = new Map();
    
    register(plugin: HookPlugin): void {
        this.plugins.set(plugin.name, plugin);
    }
    
    apply(hook: any): void {
        this.plugins.forEach(plugin => {
            plugin.install(hook);
        });
    }
}

// Example plugin: Export functionality
const exportPlugin: HookPlugin = {
    name: 'export',
    install(hook: DataTableStore) {
        hook.exportToCSV = function() {
            const csv = this.rows.map(row => 
                Object.values(row).join(',')
            ).join('\n');
            
            const blob = new Blob([csv], { type: 'text/csv' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'data.csv';
            a.click();
        };
    }
};
```

#### C. State Synchronization
```typescript
// Cross-component state synchronization
class StateManager {
    private stores: Map<string, any> = new Map();
    private subscribers: Map<string, Function[]> = new Map();
    
    register(name: string, store: any): void {
        this.stores.set(name, store);
        this.setupWatchers(name, store);
    }
    
    private setupWatchers(name: string, store: any): void {
        // Watch for changes and notify subscribers
        if (typeof store.$watch === 'function') {
            Object.keys(store).forEach(key => {
                if (typeof store[key] !== 'function') {
                    store.$watch(key, (newValue: any) => {
                        this.notify(`${name}.${key}`, newValue);
                    });
                }
            });
        }
    }
    
    subscribe(path: string, callback: Function): void {
        if (!this.subscribers.has(path)) {
            this.subscribers.set(path, []);
        }
        this.subscribers.get(path)!.push(callback);
    }
    
    private notify(path: string, value: any): void {
        const callbacks = this.subscribers.get(path) || [];
        callbacks.forEach(callback => callback(value));
    }
}

// Global state manager instance
const stateManager = new StateManager();
(window as any).stateManager = stateManager;
```

---

## Component Development Workflow

### 1. Creating New Components

**STEP-BY-STEP COMPONENT CREATION:**

#### Step 1: Define TypeScript Hook
```bash
# Create new hook file
touch web/hooks/newFeature.ts
```

```typescript
// web/hooks/newFeature.ts
interface NewFeatureConfig {
    option1?: string;
    option2?: boolean;
}

interface NewFeatureStore {
    data: any[];
    init(): void;
    handleAction(): void;
}

function useNewFeature(config: NewFeatureConfig = {}): NewFeatureStore {
    return {
        data: [],
        
        init(): void {
            console.log('NewFeature initialized');
        },
        
        handleAction(): void {
            // Implementation
        }
    };
}

// Export to global scope
(window as any).useNewFeature = useNewFeature;
```

#### Step 2: Create Templ Component
```bash
# Create component directory and file
mkdir -p web/components/feature
touch web/components/feature/newFeature.templ
```

```go
// web/components/feature/newFeature.templ
package feature

type NewFeatureProps struct {
    ID      string `json:"id"`
    Title   string `json:"title"`
    Options map[string]interface{} `json:"options"`
}

templ NewFeature(props NewFeatureProps) {
    <div 
        id={ props.ID }
        class="new-feature-component"
        x-data={ "useNewFeature(" + toJSON(props.Options) + ")" }
        x-init="init()"
    >
        <h3 class="text-lg font-semibold mb-4">{ props.Title }</h3>
        
        <!-- Component content -->
        <div class="feature-content">
            <button 
                @click="handleAction()"
                class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
                Action Button
            </button>
        </div>
    </div>
}
```

#### Step 3: Generate and Test
```bash
# Generate Templ code
templ generate

# Test TypeScript compilation
tsc --noEmit

# Start development server
go run cmd/server/main.go
```

### 2. File Organization

**RECOMMENDED FILE STRUCTURE:**

```
project/
├── web/
│   ├── hooks/                     # TypeScript hooks
│   │   ├── types.d.ts            # Global type definitions
│   │   ├── app.ts                # Application-level hooks
│   │   ├── datatable.ts          # Data table functionality
│   │   ├── forms.ts              # Form validation and handling
│   │   ├── modal.ts              # Modal management
│   │   ├── navbar.ts             # Navigation state
│   │   └── treeView.ts           # Hierarchical data display
│   │
│   ├── components/               # Templ components
│   │   ├── core/                 # Basic elements (atoms)
│   │   │   ├── button.templ
│   │   │   ├── input.templ
│   │   │   └── ...
│   │   ├── data/                 # Data components (molecules)
│   │   │   ├── datatable.templ
│   │   │   └── ...
│   │   ├── forms/                # Form components
│   │   │   ├── userForm.templ
│   │   │   └── ...
│   │   └── layout/               # Layout components (organisms)
│   │       ├── navbar.templ
│   │       ├── sidebar.templ
│   │       └── ...
│   │
│   └── styles/                   # Custom CSS (optional)
│       ├── components.css
│       └── themes.css
│
├── docs/ui/                      # Documentation
│   ├── patterns/
│   │   └── typescript-hooks-integration.md  # This file
│   └── ...
│
└── tsconfig.json                 # TypeScript configuration
```

**FILE NAMING CONVENTIONS:**
- **Hooks**: `camelCase.ts` (e.g., `dataTable.ts`, `formValidation.ts`)
- **Components**: `camelCase.templ` (e.g., `dataTable.templ`, `userForm.templ`)
- **Types**: `PascalCase` interfaces (e.g., `DataTableConfig`, `FormValidationStore`)
- **Functions**: `camelCase` with `use` prefix (e.g., `useDataTable`, `useFormValidation`)

### 3. Development Best Practices

**DEVELOPMENT BEST PRACTICES:**

#### A. Type Safety
```typescript
// Always define comprehensive interfaces
interface ComponentConfig {
    // Use specific types instead of 'any'
    data: Array<{ id: string; name: string; }>;
    
    // Use union types for constrained values
    variant: 'primary' | 'secondary' | 'success' | 'danger';
    
    // Use optional properties with defaults
    size?: 'sm' | 'md' | 'lg';
    
    // Use function types for callbacks
    onSelect?: (item: { id: string; name: string; }) => void;
}

// Use type guards for runtime safety
function isValidConfig(config: any): config is ComponentConfig {
    return config && 
           Array.isArray(config.data) &&
           ['primary', 'secondary', 'success', 'danger'].includes(config.variant);
}
```

#### B. Performance Optimization
```typescript
// Debounce expensive operations
function useOptimizedSearch() {
    let searchTimeout: any;
    
    return {
        search(query: string): void {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                this.performSearch(query);
            }, 300); // 300ms debounce
        }
    };
}

// Use computed properties instead of methods when possible
interface DataStore {
    readonly filteredData: any[]; // Computed property
    
    // Instead of getFilteredData() method
}
```

#### C. Error Handling
```typescript
interface ErrorState {
    hasError: boolean;
    errorMessage: string | null;
    errorCode?: string;
}

function useErrorHandling(): ErrorState & {
    handleError: (error: Error) => void;
    clearError: () => void;
} {
    return {
        hasError: false,
        errorMessage: null,
        errorCode: undefined,
        
        handleError(error: Error): void {
            this.hasError = true;
            this.errorMessage = error.message;
            this.errorCode = (error as any).code;
            
            // Log to console in development
            if (process.env.NODE_ENV === 'development') {
                console.error('Component Error:', error);
            }
        },
        
        clearError(): void {
            this.hasError = false;
            this.errorMessage = null;
            this.errorCode = undefined;
        }
    };
}
```

#### D. Testing Considerations
```typescript
// Make hooks testable by exposing test utilities
interface TestableHook {
    // Expose internal state for testing
    _getInternalState(): any;
    
    // Provide reset method for tests
    _reset(): void;
    
    // Mock-friendly methods
    _setMockData(data: any): void;
}

// Example testable hook
function useTestableDataTable(config: DataTableConfig): DataTableStore & TestableHook {
    const store = useDataTable(config);
    
    return {
        ...store,
        
        _getInternalState() {
            return {
                filteredRows: this.filteredRows,
                selectedRows: this.selectedRows,
                currentPage: this.currentPage
            };
        },
        
        _reset() {
            this.selectedRows = [];
            this.currentPage = 1;
            this.searchQuery = '';
        },
        
        _setMockData(data: any[]) {
            this.rows = data;
            this.updateFiltering();
        }
    };
}
```

---

## Integration with Existing Systems

### 1. Server-Side Integration

**GO HANDLER INTEGRATION:**

```go
// Handler that provides data to TypeScript hooks
func (h *Handler) GetUsersPage(w http.ResponseWriter, r *http.Request) {
    // Get data from database
    users, err := h.userService.GetUsers(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Prepare data for client-side hook
    tableConfig := DataTableConfig{
        Rows: convertUsersToTableRows(users),
        Columns: []DataTableColumn{
            {Key: "name", Title: "Name", Sortable: true},
            {Key: "email", Title: "Email", Sortable: true},
            {Key: "role", Title: "Role", Sortable: false},
        },
        Selectable: true,
        Searchable: true,
        PageSize: 20,
    }
    
    // Render page with table component
    component := pages.UsersPage(UsersPageProps{
        Title: "User Management",
        TableConfig: tableConfig,
    })
    
    component.Render(r.Context(), w)
}

// HTMX endpoint for table updates
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    
    // Perform server-side search
    users, err := h.userService.SearchUsers(r.Context(), query)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    
    // Return partial table update
    rows := convertUsersToTableRows(users)
    component := components.DataTableRows(rows)
    component.Render(r.Context(), w)
}
```

### 2. HTMX Integration

**HTMX + TYPESCRIPT HOOKS COORDINATION:**

```typescript
// Hook that coordinates with HTMX requests
function useHTMXDataTable(config: DataTableConfig): DataTableStore {
    const store = useDataTable(config);
    
    return {
        ...store,
        
        init(): void {
            // Call parent init
            store.init();
            
            // Setup HTMX event listeners
            this.setupHTMXListeners();
        },
        
        setupHTMXListeners(): void {
            const element = (this as any).$el;
            
            // Before HTMX request
            element.addEventListener('htmx:beforeRequest', (event: any) => {
                this.isLoading = true;
                (this as any).$dispatch('table-loading', { loading: true });
            });
            
            // After HTMX request
            element.addEventListener('htmx:afterRequest', (event: any) => {
                this.isLoading = false;
                (this as any).$dispatch('table-loading', { loading: false });
                
                // Handle errors
                if (event.detail.xhr.status >= 400) {
                    this.handleError(new Error('Server error: ' + event.detail.xhr.status));
                }
            });
            
            // After content swap
            element.addEventListener('htmx:afterSwap', (event: any) => {
                // Reinitialize any new components
                this.reinitializeComponents();
            });
        },
        
        // Enhanced search that works with HTMX
        handleSearch(): void {
            const searchElement = (this as any).$el.querySelector('[hx-get]');
            if (searchElement) {
                // Trigger HTMX request with search parameter
                searchElement.setAttribute('hx-vals', JSON.stringify({
                    q: this.searchQuery
                }));
                (window as any).htmx.trigger(searchElement, 'htmx:trigger');
            } else {
                // Fallback to client-side search
                store.handleSearch();
            }
        }
    };
}
```

### 3. State Persistence

**STATE PERSISTENCE PATTERNS:**

```typescript
// Hook with automatic state persistence
function usePersistentDataTable(config: DataTableConfig & {
    persistKey?: string;
}): DataTableStore {
    const persistKey = config.persistKey || 'datatable-state';
    
    return {
        ...useDataTable(config),
        
        init(): void {
            // Load persisted state
            this.loadPersistedState();
            
            // Setup auto-save watchers
            this.setupPersistence();
        },
        
        loadPersistedState(): void {
            try {
                const saved = localStorage.getItem(persistKey);
                if (saved) {
                    const state = JSON.parse(saved);
                    this.currentPage = state.currentPage || 1;
                    this.pageSize = state.pageSize || this.pageSize;
                    this.sortColumn = state.sortColumn || '';
                    this.sortDirection = state.sortDirection || 'asc';
                    this.searchQuery = state.searchQuery || '';
                }
            } catch (error) {
                console.warn('Failed to load persisted state:', error);
            }
        },
        
        setupPersistence(): void {
            // Debounced save function
            const saveState = this.debounce(() => {
                try {
                    const state = {
                        currentPage: this.currentPage,
                        pageSize: this.pageSize,
                        sortColumn: this.sortColumn,
                        sortDirection: this.sortDirection,
                        searchQuery: this.searchQuery
                    };
                    localStorage.setItem(persistKey, JSON.stringify(state));
                } catch (error) {
                    console.warn('Failed to persist state:', error);
                }
            }, 500);
            
            // Watch for changes
            if (typeof (this as any).$watch === 'function') {
                const watchProps = ['currentPage', 'pageSize', 'sortColumn', 'sortDirection', 'searchQuery'];
                watchProps.forEach(prop => {
                    (this as any).$watch(prop, saveState);
                });
            }
        },
        
        debounce(func: Function, wait: number): Function {
            let timeout: any;
            return function(...args: any[]) {
                clearTimeout(timeout);
                timeout = setTimeout(() => func.apply(this, args), wait);
            };
        }
    };
}
```

---

## Testing Strategies

### 1. TypeScript Hook Testing

**UNIT TESTING APPROACH:**

```typescript
// test/hooks/dataTable.test.ts
describe('useDataTable Hook', () => {
    let hookInstance: DataTableStore;
    
    beforeEach(() => {
        // Create fresh hook instance for each test
        hookInstance = useDataTable({
            rows: [
                { id: '1', name: 'John', email: 'john@example.com' },
                { id: '2', name: 'Jane', email: 'jane@example.com' }
            ],
            columns: [
                { key: 'name', title: 'Name', sortable: true },
                { key: 'email', title: 'Email', sortable: true }
            ]
        });
        
        // Mock Alpine.js methods
        (hookInstance as any).$watch = jest.fn();
        (hookInstance as any).$dispatch = jest.fn();
    });
    
    test('should initialize with correct default state', () => {
        expect(hookInstance.currentPage).toBe(1);
        expect(hookInstance.selectedRows).toEqual([]);
        expect(hookInstance.searchQuery).toBe('');
    });
    
    test('should handle search correctly', () => {
        hookInstance.searchQuery = 'John';
        hookInstance.handleSearch();
        
        expect(hookInstance.filteredRows).toHaveLength(1);
        expect(hookInstance.filteredRows[0].name).toBe('John');
    });
    
    test('should handle row selection', () => {
        hookInstance.toggleRowSelection('1');
        
        expect(hookInstance.selectedRows).toContain('1');
        expect(hookInstance.isRowSelected('1')).toBe(true);
    });
    
    test('should handle pagination', () => {
        hookInstance.pageSize = 1;
        hookInstance.updatePagination();
        
        expect(hookInstance.totalPages).toBe(2);
        expect(hookInstance.paginatedRows).toHaveLength(1);
    });
});
```

**INTEGRATION TESTING:**

```typescript
// test/integration/components.test.ts
describe('Component Integration', () => {
    let container: HTMLElement;
    
    beforeEach(() => {
        container = document.createElement('div');
        document.body.appendChild(container);
    });
    
    afterEach(() => {
        document.body.removeChild(container);
    });
    
    test('should integrate TypeScript hook with Alpine.js', async () => {
        // Create component HTML
        container.innerHTML = `
            <div x-data="useDataTable({
                rows: [
                    { id: '1', name: 'Test User' }
                ]
            })" x-init="init()">
                <span x-text="rows.length"></span>
                <button @click="handleSearch()">Search</button>
            </div>
        `;
        
        // Initialize Alpine.js
        (window as any).Alpine.start();
        
        // Wait for initialization
        await new Promise(resolve => setTimeout(resolve, 100));
        
        // Check initial state
        const span = container.querySelector('span');
        expect(span?.textContent).toBe('1');
        
        // Test interaction
        const button = container.querySelector('button');
        button?.click();
        
        // Verify hook method was called
        // (This would require more sophisticated mocking in real tests)
    });
});
```

### 2. Component Testing

**TEMPL COMPONENT TESTING:**

```go
// test/components/datatable_test.go
package components_test

import (
    "context"
    "strings"
    "testing"
    
    "your-project/web/components/data"
)

func TestDataTableComponent(t *testing.T) {
    tests := []struct {
        name     string
        props    data.DataTableProps
        contains []string
    }{
        {
            name: "renders basic table",
            props: data.DataTableProps{
                ID: "test-table",
                Columns: []data.DataTableColumn{
                    {Key: "name", Title: "Name"},
                },
                Rows: []data.DataTableRow{
                    {ID: "1", Data: map[string]interface{}{"name": "Test"}},
                },
            },
            contains: []string{
                `id="test-table"`,
                `x-data="useDataTable(`,
                `x-init="init()"`,
                "Name",
            },
        },
        {
            name: "renders searchable table",
            props: data.DataTableProps{
                ID:         "searchable-table",
                Searchable: true,
                Columns:    []data.DataTableColumn{{Key: "name", Title: "Name"}},
            },
            contains: []string{
                `placeholder="Search..."`,
                `@input="handleSearch()"`,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf strings.Builder
            err := data.DataTable(tt.props).Render(context.Background(), &buf)
            
            if err != nil {
                t.Fatalf("Failed to render component: %v", err)
            }
            
            html := buf.String()
            for _, expected := range tt.contains {
                if !strings.Contains(html, expected) {
                    t.Errorf("Expected HTML to contain %q, but it didn't. Got: %s", expected, html)
                }
            }
        })
    }
}
```

---

## Performance Optimization

### 1. Bundle Optimization

**TYPESCRIPT COMPILATION OPTIMIZATION:**

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM"],
    "outDir": "./dist",
    "rootDir": "./web/hooks",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["web/hooks/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

**BUILD OPTIMIZATION:**

```bash
# Development build with source maps
tsc --sourcemap

# Production build with minification
tsc && uglifyjs dist/**/*.js -c -m -o dist/hooks.min.js

# Or use modern bundler
npx esbuild web/hooks/*.ts --bundle --minify --outdir=dist
```

### 2. Runtime Optimization

**PERFORMANCE OPTIMIZATION TECHNIQUES:**

```typescript
// Lazy loading for expensive hooks
function useLazyDataTable(config: DataTableConfig): Promise<DataTableStore> {
    return new Promise((resolve) => {
        // Load heavy dependencies only when needed
        import('./heavyDataTableFeatures').then((features) => {
            const baseStore = useDataTable(config);
            const enhancedStore = {
                ...baseStore,
                ...features.advancedFeatures(baseStore)
            };
            resolve(enhancedStore);
        });
    });
}

// Virtual scrolling for large datasets
function useVirtualizedDataTable(config: DataTableConfig): DataTableStore {
    return {
        ...useDataTable(config),
        
        // Override pagination for virtual scrolling
        get paginatedRows(): any[] {
            const startIndex = this.virtualScrollTop;
            const endIndex = Math.min(
                startIndex + this.virtualViewportSize,
                this.filteredRows.length
            );
            return this.filteredRows.slice(startIndex, endIndex);
        },
        
        // Add virtual scroll properties
        virtualScrollTop: 0,
        virtualViewportSize: 50,
        virtualRowHeight: 40,
        
        // Handle scroll events efficiently
        handleVirtualScroll(scrollTop: number): void {
            const newIndex = Math.floor(scrollTop / this.virtualRowHeight);
            if (newIndex !== this.virtualScrollTop) {
                this.virtualScrollTop = newIndex;
            }
        }
    };
}

// Memoization for expensive computations
class MemoizedComputation {
    private cache = new Map<string, any>();
    
    memoize<T>(key: string, fn: () => T): T {
        if (this.cache.has(key)) {
            return this.cache.get(key);
        }
        
        const result = fn();
        this.cache.set(key, result);
        return result;
    }
    
    invalidate(key: string): void {
        this.cache.delete(key);
    }
    
    clear(): void {
        this.cache.clear();
    }
}

// Usage in hooks
function useOptimizedDataTable(config: DataTableConfig): DataTableStore {
    const memo = new MemoizedComputation();
    
    return {
        ...useDataTable(config),
        
        // Memoized filtering
        get filteredRows(): any[] {
            const key = `filtered-${this.searchQuery}-${JSON.stringify(this.filters)}`;
            return memo.memoize(key, () => {
                return this.rows.filter(row => {
                    // Expensive filtering logic
                    return this.matchesSearch(row) && this.matchesFilters(row);
                });
            });
        },
        
        // Invalidate cache when data changes
        updateData(newRows: any[]): void {
            this.rows = newRows;
            memo.clear(); // Clear all cached computations
        }
    };
}
```

---

## Migration and Upgrade Strategies

### 1. JavaScript to TypeScript Migration

**MIGRATION ROADMAP:**

#### Phase 1: Add TypeScript Support
```bash
# Add TypeScript configuration
npm install -D typescript @types/node

# Create tsconfig.json
cat > tsconfig.json << 'EOF'
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM"],
    "strict": true,
    "allowJs": true,
    "checkJs": false
  },
  "include": ["web/hooks/**/*"]
}
EOF

# Create types file
touch web/hooks/types.d.ts
```

#### Phase 2: Convert JavaScript Files
```typescript
// Before: JavaScript hook
function useDataTable(config) {
    return {
        rows: config.rows || [],
        
        init() {
            this.updateData();
        },
        
        updateData() {
            // Implementation
        }
    };
}

// After: TypeScript hook with interfaces
interface DataTableConfig {
    rows?: any[];
    pageSize?: number;
}

interface DataTableStore {
    rows: any[];
    init(): void;
    updateData(): void;
}

function useDataTable(config: DataTableConfig = {}): DataTableStore {
    return {
        rows: config.rows || [],
        
        init(): void {
            this.updateData();
        },
        
        updateData(): void {
            // Implementation with type safety
        }
    };
}
```

#### Phase 3: Add Advanced Types
```typescript
// Add generic types for better type safety
interface DataRow {
    id: string;
    [key: string]: any;
}

interface DataTableConfig<T extends DataRow = DataRow> {
    rows?: T[];
    pageSize?: number;
    onSelect?: (row: T) => void;
}

interface DataTableStore<T extends DataRow = DataRow> {
    rows: T[];
    selectedRows: string[];
    
    selectRow(row: T): void;
    getSelectedRows(): T[];
}

function useDataTable<T extends DataRow = DataRow>(
    config: DataTableConfig<T> = {}
): DataTableStore<T> {
    // Implementation with generic type safety
}
```

### 2. Version Management

**VERSION MANAGEMENT STRATEGY:**

```typescript
// Version-aware hook system
interface HookVersion {
    major: number;
    minor: number;
    patch: number;
}

class VersionManager {
    private currentVersion: HookVersion = { major: 2, minor: 0, patch: 0 };
    
    isCompatible(requiredVersion: HookVersion): boolean {
        // Semantic versioning compatibility check
        return this.currentVersion.major === requiredVersion.major &&
               this.currentVersion.minor >= requiredVersion.minor;
    }
    
    getDeprecationWarning(hookName: string): string | null {
        const deprecated = {
            'useOldDataTable': 'Use useDataTable instead. Will be removed in v3.0.0',
            'useFormValidation': 'Use useAdvancedForm instead. Will be removed in v3.0.0'
        };
        
        return deprecated[hookName] || null;
    }
}

// Version-aware hook factory
function createVersionedHook<T>(
    hookName: string,
    hookFunction: () => T,
    requiredVersion: HookVersion
): () => T {
    return function() {
        const versionManager = new VersionManager();
        
        // Check compatibility
        if (!versionManager.isCompatible(requiredVersion)) {
            console.error(`Hook ${hookName} requires version ${requiredVersion.major}.${requiredVersion.minor}.${requiredVersion.patch} or higher`);
            throw new Error('Incompatible hook version');
        }
        
        // Show deprecation warning
        const warning = versionManager.getDeprecationWarning(hookName);
        if (warning) {
            console.warn(`DEPRECATED: ${hookName} - ${warning}`);
        }
        
        return hookFunction();
    };
}

// Usage
const useDataTableV2 = createVersionedHook(
    'useDataTable',
    () => useDataTable(),
    { major: 2, minor: 0, patch: 0 }
);
```

---

## References

**RELATED DOCUMENTATION:**
- [Getting Started](../getting-started.md) - Initial setup and configuration
- [Architecture](../fundamentals/architecture.md) - System design principles
- [Components](../components/elements.md) - Basic component library
- [Forms](../components/forms.md) - Form components and validation
- [HTMX Integration](./htmx-integration.md) - Server interaction patterns

**EXTERNAL REFERENCES:**
- [TypeScript Handbook](https://www.typescriptlang.org/docs/) - TypeScript language features
- [Alpine.js Guide](https://alpinejs.dev/start-here) - Reactive JavaScript framework
- [Templ Documentation](https://templ.guide/) - Go templating language
- [Flowbite Components](https://flowbite.com/docs/components/) - UI component library
- [TailwindCSS](https://tailwindcss.com/docs) - Utility-first CSS framework

**COMMUNITY RESOURCES:**
- [Alpine.js Examples](https://alpinejs.dev/start-here) - Community patterns and examples
- [TypeScript Best Practices](https://typescript-eslint.io/rules/) - Code quality guidelines
- [Web Components Standards](https://developer.mozilla.org/en-US/docs/Web/Web_Components) - Modern web standards

**METADATA FOR AI ASSISTANTS:**
- File Type: Integration Documentation
- Scope: TypeScript hooks, Alpine.js, Templ components, development workflow
- Target Audience: Full-stack developers building ERP interfaces
- Complexity: Intermediate to Advanced
- Focus: Type-safe client-side interactions with server-rendered components
- Dependencies: TypeScript, Alpine.js, Templ, Flowbite, TailwindCSS
- Last Updated: December 2024
