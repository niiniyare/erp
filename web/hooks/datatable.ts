/**
 * DataTable AlpineJS Hooks (TypeScript)
 * 
 * Extracted from internal/ui/components/data/datatable.templ
 * Provides reusable DataTable functionality for all ERP services
 * 
 * @version 2.0.0
 * @backward-compatible with v1.x
 */

/// <reference path="./types.d.ts" />

// Import global Alpine.js hook functions from window  
// These are already defined in types.d.ts with 'any' types

// Types
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

interface BulkAction {
    id: string;
    text: string;
    icon?: string;
    variant?: string;
    permission?: string;
    endpoint: string;
    method?: string;
    confirm?: boolean;
    confirmText?: string;
    refreshOnSuccess?: boolean;
}

interface PaginationInfo {
    start: number;
    end: number;
    total: number;
}

interface ExportOptions {
    format: 'csv' | 'json';
    rowCount: number;
}

interface RequestOptions {
    method?: string;
    data?: Record<string, any>;
    timeout?: number;
}

interface ApiResponse {
    success: boolean;
    message?: string;
    data?: any;
}

interface DataTableStore {
    // State Management
    rows: any[];
    filteredRows: any[];
    paginatedRows: any[];
    selectedRows: string[];
    _selectedRowsSet: Set<string> | null;
    
    // Search & Filter
    searchQuery: string;
    searchColumns: string[];
    _searchDebounceTimer: any;
    
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
    clearSelectionOnSearch: boolean;
    searchDebounce: number;
    enableExport: boolean;
    requestTimeout: number;
    
    // Loading states
    isLoading: boolean;
    isBulkActionInProgress: boolean;
    
    // Computed Properties
    readonly allSelected: boolean;
    readonly someSelected: boolean;
    readonly selectedCount: number;
    readonly hasNextPage: boolean;
    readonly hasPrevPage: boolean;
    readonly isEmpty: boolean;
    readonly paginationInfo: PaginationInfo;
    readonly selectedRowsSet: Set<string>;
    
    // Methods
    init(): void;
    handleSearch(): void;
    handleSearchImmediate(): void;
    sort(column: string): void;
    updatePagination(): void;
    updateTotals(): void;
    nextPage(): void;
    prevPage(): void;
    goToPage(page: number | string): void;
    setPageSize(size: number | string): void;
    toggleSelectAll(): void;
    toggleRowSelection(rowId: string): void;
    isRowSelected(rowId: string): boolean;
    clearSelection(): void;
    selectRows(rowIds: string[]): void;
    executeBulkAction(actionId: string): Promise<void>;
    exportData(format?: 'csv' | 'json'): Promise<void>;
    exportToCSV(data: any[]): void;
    exportToJSON(data: any[]): void;
    downloadFile(content: string, filename: string, mimeType: string): void;
    getNestedValue(obj: any, path: string): any;
    hasPermission(permission: string): boolean;
    makeHTMXRequest(url: string, options?: RequestOptions): Promise<ApiResponse>;
    showSuccess(message: string): void;
    showError(message: string): void;
    showConfirmation(message: string): Promise<boolean>;
    refreshData(): void;
    logDebug(message: string, data?: Record<string, any>): void;
    logError(message: string, error: any): void;
    
    // Backward compatibility methods
    add(rowId: string): void;
    delete(rowId: string): void;
    has(rowId: string): boolean;
    clear(): void;
}

interface Alpine {
    store: (name: string) => any;
}

// Export to make this a module
export {};

/**
 * Core DataTable Hook
 * Manages state, search, sorting, pagination, and bulk actions
 */
function useDataTable(config: DataTableConfig = {}): DataTableStore {
    return {
        // State Management
        rows: config.rows || [],
        filteredRows: [],
        paginatedRows: [],
        selectedRows: [],
        _selectedRowsSet: null,
        
        // Search & Filter
        searchQuery: '',
        searchColumns: config.searchColumns || [],
        _searchDebounceTimer: null,
        
        // Sorting
        sortColumn: config.defaultSort?.column || '',
        sortDirection: config.defaultSort?.direction || 'asc',
        
        // Pagination
        currentPage: 1,
        pageSize: config.pageSize || 10,
        totalRows: 0,
        totalPages: 0,
        
        // Configuration
        bulkActions: config.bulkActions || [],
        permissions: config.permissions || {},
        serviceContext: config.serviceContext || 'console',
        clearSelectionOnSearch: config.clearSelectionOnSearch !== false,
        searchDebounce: config.searchDebounce || 300,
        enableExport: config.enableExport || false,
        requestTimeout: config.requestTimeout || 30000,
        
        // Loading states
        isLoading: false,
        isBulkActionInProgress: false,
        
        // Computed Properties
        get allSelected(): boolean {
            return this.filteredRows.length > 0 && 
                   this.filteredRows.every((row: any) => this.isRowSelected(row?.id));
        },
        
        get someSelected(): boolean {
            return this.selectedRows.length > 0 && 
                   this.selectedRows.length < this.filteredRows.length;
        },
        
        get selectedCount(): number {
            return this.selectedRows.length;
        },
        
        get hasNextPage(): boolean {
            return this.currentPage < this.totalPages;
        },
        
        get hasPrevPage(): boolean {
            return this.currentPage > 1;
        },
        
        get isEmpty(): boolean {
            return this.filteredRows.length === 0;
        },
        
        get paginationInfo(): PaginationInfo {
            const start = this.isEmpty ? 0 : (this.currentPage - 1) * this.pageSize + 1;
            const end = Math.min(this.currentPage * this.pageSize, this.totalRows);
            return { start, end, total: this.totalRows };
        },
        
        get selectedRowsSet(): Set<string> {
            if (!this._selectedRowsSet) {
                this._selectedRowsSet = new Set(this.selectedRows);
            }
            return this._selectedRowsSet;
        },
        
        // Initialization
        init(): void {
            this.filteredRows = [...this.rows];
            this.updatePagination();
            this.updateTotals();
            
            // Watch for external data updates
            if (typeof (this as any).$watch === 'function') {
                (this as any).$watch('rows', () => {
                    this.handleSearchImmediate();
                });
                
                // Sync selectedRows array with internal Set
                (this as any).$watch('selectedRows', () => {
                    this._selectedRowsSet = new Set(this.selectedRows);
                });
            }
            
            // Log initialization
            this.logDebug('DataTable initialized', {
                rowCount: this.rows.length,
                serviceContext: this.serviceContext
            });
        },
        
        // Search Functionality (Debounced)
        handleSearch(): void {
            if (this._searchDebounceTimer) {
                clearTimeout(this._searchDebounceTimer);
            }
            
            this._searchDebounceTimer = setTimeout(() => {
                this.handleSearchImmediate();
            }, this.searchDebounce);
        },
        
        // Immediate search (for internal use)
        handleSearchImmediate(): void {
            try {
                if (!this.searchQuery.trim()) {
                    this.filteredRows = [...this.rows];
                } else {
                    const query = this.searchQuery.toLowerCase().trim();
                    this.filteredRows = this.rows.filter((row: any) => {
                        return this.searchColumns.some(column => {
                            const value = this.getNestedValue(row, column);
                            return String(value ?? '').toLowerCase().includes(query);
                        });
                    });
                }
                
                this.currentPage = 1; // Reset to first page
                
                // Clear selections if configured
                if (this.clearSelectionOnSearch) {
                    this.clearSelection();
                }
                
                this.updatePagination();
                this.updateTotals();
                
                // Emit search event
                if (typeof (this as any).$dispatch === 'function') {
                    (this as any).$dispatch('datatable-search', {
                        query: this.searchQuery,
                        resultCount: this.filteredRows.length
                    });
                }
            } catch (error) {
                this.logError('Search failed', error);
                this.showError('Search failed. Please try again.');
            }
        },
        
        // Sorting Functionality
        sort(column: string): void {
            if (!column) return;
            
            try {
                if (this.sortColumn === column) {
                    this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
                } else {
                    this.sortColumn = column;
                    this.sortDirection = 'asc';
                }
                
                this.filteredRows.sort((a: any, b: any) => {
                    const aVal = this.getNestedValue(a, column);
                    const bVal = this.getNestedValue(b, column);
                    
                    // Handle null/undefined values
                    if (aVal == null && bVal == null) return 0;
                    if (aVal == null) return 1;
                    if (bVal == null) return -1;
                    
                    // Type-aware comparison
                    let comparison = 0;
                    if (typeof aVal === 'number' && typeof bVal === 'number') {
                        comparison = aVal - bVal;
                    } else if (aVal instanceof Date && bVal instanceof Date) {
                        comparison = aVal.getTime() - bVal.getTime();
                    } else {
                        comparison = String(aVal).localeCompare(String(bVal));
                    }
                    
                    return this.sortDirection === 'asc' ? comparison : -comparison;
                });
                
                this.updatePagination();
                
                // Emit sort event
                if (typeof (this as any).$dispatch === 'function') {
                    (this as any).$dispatch('datatable-sort', {
                        column: this.sortColumn,
                        direction: this.sortDirection
                    });
                }
            } catch (error) {
                this.logError('Sort failed', error);
                this.showError('Sorting failed. Please try again.');
            }
        },
        
        // Pagination
        updatePagination(): void {
            this.totalRows = this.filteredRows.length;
            this.totalPages = Math.ceil(this.totalRows / this.pageSize) || 1;
            
            // Ensure current page is valid
            if (this.currentPage > this.totalPages) {
                this.currentPage = this.totalPages;
            }
            
            const start = (this.currentPage - 1) * this.pageSize;
            const end = start + this.pageSize;
            this.paginatedRows = this.filteredRows.slice(start, end);
        },
        
        updateTotals(): void {
            this.totalRows = this.filteredRows.length;
            this.totalPages = Math.ceil(this.totalRows / this.pageSize) || 1;
        },
        
        nextPage(): void {
            if (this.hasNextPage) {
                this.currentPage++;
                this.updatePagination();
                if (typeof (this as any).$dispatch === 'function') {
                    (this as any).$dispatch('datatable-page-change', { page: this.currentPage });
                }
            }
        },
        
        prevPage(): void {
            if (this.hasPrevPage) {
                this.currentPage--;
                this.updatePagination();
                if (typeof (this as any).$dispatch === 'function') {
                    (this as any).$dispatch('datatable-page-change', { page: this.currentPage });
                }
            }
        },
        
        goToPage(page: number | string): void {
            const pageNum = parseInt(String(page), 10);
            if (pageNum >= 1 && pageNum <= this.totalPages) {
                this.currentPage = pageNum;
                this.updatePagination();
                if (typeof (this as any).$dispatch === 'function') {
                    (this as any).$dispatch('datatable-page-change', { page: this.currentPage });
                }
            }
        },
        
        setPageSize(size: number | string): void {
            const newSize = parseInt(String(size), 10);
            if (newSize > 0) {
                this.pageSize = newSize;
                this.currentPage = 1;
                this.updatePagination();
            }
        },
        
        // Selection Management
        toggleSelectAll(): void {
            if (this.allSelected) {
                this.selectedRows = [];
            } else {
                this.selectedRows = this.filteredRows
                    .map((row: any) => row?.id)
                    .filter((id: any) => id != null);
            }
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('datatable-selection-change', {
                    selectedCount: this.selectedRows.length
                });
            }
        },
        
        toggleRowSelection(rowId: string): void {
            if (rowId == null) return;
            
            const index = this.selectedRows.indexOf(rowId);
            if (index > -1) {
                this.selectedRows.splice(index, 1);
            } else {
                this.selectedRows.push(rowId);
            }
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('datatable-selection-change', {
                    selectedCount: this.selectedRows.length
                });
            }
        },
        
        isRowSelected(rowId: string): boolean {
            if (rowId == null) return false;
            return this.selectedRows.includes(rowId);
        },
        
        clearSelection(): void {
            this.selectedRows = [];
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('datatable-selection-change', { selectedCount: 0 });
            }
        },
        
        selectRows(rowIds: string[]): void {
            if (!Array.isArray(rowIds)) return;
            this.selectedRows = [...new Set([...this.selectedRows, ...rowIds])];
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('datatable-selection-change', {
                    selectedCount: this.selectedRows.length
                });
            }
        },
        
        // Bulk Actions
        async executeBulkAction(actionId: string): Promise<void> {
            const action = this.bulkActions.find(a => a.id === actionId);
            if (!action || this.selectedRows.length === 0) {
                if (this.selectedRows.length === 0) {
                    this.showError('Please select at least one item');
                }
                return;
            }
            
            // Prevent concurrent bulk actions
            if (this.isBulkActionInProgress) {
                this.showError('Another action is in progress');
                return;
            }
            
            // Permission check
            if (action.permission && !this.hasPermission(action.permission)) {
                this.showError('You do not have permission to perform this action');
                return;
            }
            
            // Confirmation dialog
            if (action.confirm) {
                const confirmed = await this.showConfirmation(
                    action.confirmText || 
                    `Are you sure you want to ${action.text.toLowerCase()} ${this.selectedRows.length} item(s)?`
                );
                if (!confirmed) return;
            }
            
            this.isBulkActionInProgress = true;
            
            try {
                const selectedIds = [...this.selectedRows];
                const response = await this.makeHTMXRequest(action.endpoint, {
                    method: action.method || 'POST',
                    data: {
                        action: actionId,
                        ids: selectedIds,
                        service: this.serviceContext
                    },
                    timeout: this.requestTimeout
                });
                
                if (response?.success) {
                    this.clearSelection();
                    this.showSuccess(response.message || `${action.text} completed successfully`);
                    
                    // Refresh data if needed
                    if (action.refreshOnSuccess !== false) {
                        this.refreshData();
                    }
                    
                    // Emit success event
                    if (typeof (this as any).$dispatch === 'function') {
                        (this as any).$dispatch('datatable-bulk-action-success', {
                            action: actionId,
                            count: selectedIds.length
                        });
                    }
                } else {
                    this.showError(response?.message || `Failed to ${action.text.toLowerCase()}`);
                }
                
            } catch (error) {
                this.logError('Bulk action failed', error);
                this.showError(`Failed to ${action.text.toLowerCase()}: ${(error as Error).message}`);
            } finally {
                this.isBulkActionInProgress = false;
            }
        },
        
        // Export Functionality
        async exportData(format: 'csv' | 'json' = 'csv'): Promise<void> {
            if (!this.enableExport) {
                this.showError('Export is not enabled');
                return;
            }
            
            try {
                const dataToExport = this.selectedRows.length > 0
                    ? this.filteredRows.filter((row: any) => this.isRowSelected(row?.id))
                    : this.filteredRows;
                
                if (dataToExport.length === 0) {
                    this.showError('No data to export');
                    return;
                }
                
                if (format === 'csv') {
                    this.exportToCSV(dataToExport);
                } else if (format === 'json') {
                    this.exportToJSON(dataToExport);
                }
                
                if (typeof (this as any).$dispatch === 'function') {
                    (this as any).$dispatch('datatable-export', {
                        format,
                        rowCount: dataToExport.length
                    });
                }
            } catch (error) {
                this.logError('Export failed', error);
                this.showError('Export failed. Please try again.');
            }
        },
        
        exportToCSV(data: any[]): void {
            if (!data.length) return;
            
            const headers = this.searchColumns.length > 0
                ? this.searchColumns
                : Object.keys(data[0]);
            
            const csvContent = [
                headers.join(','),
                ...data.map((row: any) => 
                    headers.map(header => {
                        const value = this.getNestedValue(row, header);
                        const escaped = String(value ?? '').replace(/"/g, '""');
                        return `"${escaped}"`;
                    }).join(',')
                )
            ].join('\n');
            
            this.downloadFile(csvContent, `export-${Date.now()}.csv`, 'text/csv');
        },
        
        exportToJSON(data: any[]): void {
            const jsonContent = JSON.stringify(data, null, 2);
            this.downloadFile(jsonContent, `export-${Date.now()}.json`, 'application/json');
        },
        
        downloadFile(content: string, filename: string, mimeType: string): void {
            const blob = new Blob([content], { type: mimeType });
            const url = URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(url);
        },
        
        // Utility Methods
        getNestedValue(obj: any, path: string): any {
            if (!obj || !path) return null;
            return path.split('.').reduce((current, key) => current?.[key], obj);
        },
        
        hasPermission(permission: string): boolean {
            // TODO: Integration with global permission system
            return (window as any).Alpine?.store?.('app')?.hasPermission?.(permission) ?? true;
        },
        
        async makeHTMXRequest(url: string, options: RequestOptions = {}): Promise<ApiResponse> {
            if (!url) {
                throw new Error('URL is required');
            }
            
            return new Promise((resolve, reject) => {
                const controller = new AbortController();
                const timeoutId = setTimeout(() => {
                    controller.abort();
                    reject(new Error('Request timeout'));
                }, options.timeout || this.requestTimeout);
                
                const formData = new FormData();
                Object.entries(options.data || {}).forEach(([key, value]: [string, any]) => {
                    if (Array.isArray(value)) {
                        value.forEach(v => formData.append(`${key}[]`, v));
                    } else if (value != null) {
                        formData.append(key, value);
                    }
                });
                
                fetch(url, {
                    method: options.method || 'POST',
                    body: formData,
                    headers: {
                        'X-Requested-With': 'XMLHttpRequest',
                        'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || ''
                    },
                    signal: controller.signal
                })
                .then(async response => {
                    clearTimeout(timeoutId);
                    
                    if (!response.ok) {
                        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                    }
                    
                    const contentType = response.headers.get('content-type');
                    if (contentType?.includes('application/json')) {
                        return response.json();
                    } else {
                        const text = await response.text();
                        return { success: true, data: text };
                    }
                })
                .then(resolve)
                .catch(error => {
                    clearTimeout(timeoutId);
                    reject(error);
                });
            });
        },
        
        showSuccess(message: string): void {
            (window as any).Alpine?.store?.('toast')?.success?.(message);
            console.info('[DataTable] Success:', message);
        },
        
        showError(message: string): void {
            (window as any).Alpine?.store?.('toast')?.error?.(message);
            console.error('[DataTable] Error:', message);
        },
        
        async showConfirmation(message: string): Promise<boolean> {
            // confirmation with custom modal support
            if ((window as any).Alpine?.store?.('modal')?.confirm) {
                return await (window as any).Alpine.store('modal').confirm(message);
            }
            return window.confirm(message);
        },
        
        refreshData(): void {
            // Trigger HTMX refresh or emit event for parent to handle
            if (typeof (this as any).$dispatch === 'function') {
                (this as any).$dispatch('datatable-refresh', { context: this.serviceContext });
            }
        },
        
        // Logging utilities
        logDebug(message: string, data: Record<string, any> = {}): void {
            if ((window as any).Alpine?.store?.('app')?.debug) {
                console.log(`[DataTable:${this.serviceContext}]`, message, data);
            }
        },
        
        logError(message: string, error: any): void {
            console.error(`[DataTable:${this.serviceContext}]`, message, error);
        },
        
        // Backward compatibility methods
        add(rowId: string): void {
            console.warn('DataTable.add() is deprecated. Use toggleRowSelection() instead.');
            if (!this.isRowSelected(rowId)) {
                this.toggleRowSelection(rowId);
            }
        },
        
        delete(rowId: string): void {
            console.warn('DataTable.delete() is deprecated. Use toggleRowSelection() instead.');
            if (this.isRowSelected(rowId)) {
                this.toggleRowSelection(rowId);
            }
        },
        
        has(rowId: string): boolean {
            console.warn('DataTable.has() is deprecated. Use isRowSelected() instead.');
            return this.isRowSelected(rowId);
        },
        
        clear(): void {
            console.warn('DataTable.clear() is deprecated. Use clearSelection() instead.');
            this.clearSelection();
        }
    };
}

/**
 * Service-specific DataTable configurations
 */
const DataTableConfigs = {
    console: {
        searchColumns: ['name', 'email', 'status', 'created_at'],
        enableExport: true,
        bulkActions: [
            {
                id: 'activate',
                text: 'Activate',
                icon: 'fas fa-check',
                variant: 'success',
                permission: 'tenant:activate',
                endpoint: '/console/bulk/activate',
                method: 'POST'
            },
            {
                id: 'suspend',
                text: 'Suspend',
                icon: 'fas fa-pause',
                variant: 'warning',
                permission: 'tenant:suspend',
                confirm: true,
                confirmText: 'Are you sure you want to suspend these tenants?',
                endpoint: '/console/bulk/suspend',
                method: 'POST'
            },
            {
                id: 'delete',
                text: 'Delete',
                icon: 'fas fa-trash',
                variant: 'danger',
                permission: 'tenant:delete',
                confirm: true,
                confirmText: 'This action cannot be undone. Delete selected tenants?',
                endpoint: '/console/bulk/delete',
                method: 'DELETE'
            }
        ]
    },
    
    workspace: {
        searchColumns: ['name', 'email', 'department', 'role'],
        enableExport: true,
        bulkActions: [
            {
                id: 'invite',
                text: 'Send Invites',
                icon: 'fas fa-envelope',
                variant: 'primary',
                permission: 'user:invite',
                endpoint: '/workspace/users/bulk',
                method: 'POST'
            },
            {
                id: 'deactivate',
                text: 'Deactivate',
                icon: 'fas fa-user-slash',
                variant: 'warning',
                permission: 'user:deactivate',
                confirm: true,
                confirmText: 'Deactivate selected users?',
                endpoint: '/workspace/users/bulk',
                method: 'POST'
            },
            {
                id: 'change-role',
                text: 'Change Role',
                icon: 'fas fa-user-tag',
                variant: 'secondary',
                permission: 'user:manage',
                confirm: true,
                confirmText: 'Change role for selected users?',
                endpoint: '/workspace/users/bulk',
                method: 'POST'
            },
            {
                id: 'export',
                text: 'Export Data',
                icon: 'fas fa-download',
                variant: 'secondary',
                permission: 'user:export',
                endpoint: '/workspace/users/export',
                method: 'GET'
            }
        ]
    },
    
    portal: {
        searchColumns: ['number', 'amount', 'status', 'due_date', 'description'],
        enableExport: true,
        bulkActions: [
            {
                id: 'download',
                text: 'Download PDFs',
                icon: 'fas fa-download',
                variant: 'secondary',
                permission: 'client:read',
                endpoint: '/portal/invoices/bulk',
                method: 'POST'
            },
            {
                id: 'pay',
                text: 'Pay Selected',
                icon: 'fas fa-credit-card',
                variant: 'success',
                permission: 'client:read',
                confirm: true,
                confirmText: 'Proceed with payment for selected invoices?',
                endpoint: '/portal/invoices/bulk',
                method: 'POST'
            },
            {
                id: 'email',
                text: 'Email Support',
                icon: 'fas fa-envelope',
                variant: 'primary',
                permission: 'client:read',
                confirm: true,
                confirmText: 'Contact support about selected invoices?',
                endpoint: '/portal/invoices/bulk',
                method: 'POST'
            }
        ]
    }
};

/**
 * Convenience factories for service-specific DataTables
 */
(window as any).createConsoleDataTable = (config: DataTableConfig = {}): DataTableStore => {
    return useDataTable({
        ...DataTableConfigs.console,
        ...config,
        serviceContext: 'console'
    });
};

(window as any).createWorkspaceDataTable = (config: DataTableConfig = {}): DataTableStore => {
    return useDataTable({
        ...DataTableConfigs.workspace,
        ...config,
        serviceContext: 'workspace'
    });
};

(window as any).createPortalDataTable = (config: DataTableConfig = {}): DataTableStore => {
    return useDataTable({
        ...DataTableConfigs.portal,
        ...config,
        serviceContext: 'portal'
    });
};

// Export for use in components
(window as any).useDataTable = useDataTable;
(window as any).DataTableConfigs = DataTableConfigs;

// Version info
(window as any).DataTableVersion = '2.0.0';