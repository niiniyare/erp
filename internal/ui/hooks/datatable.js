/**
 * DataTable AlpineJS Hooks
 * 
 * Extracted from internal/ui/components/data/datatable.templ
 * Provides reusable DataTable functionality for all ERP services
 */

/**
 * Core DataTable Hook
 * Manages state, search, sorting, pagination, and bulk actions
 */
function useDataTable(config = {}) {
    return {
        // State Management
        rows: config.rows || [],
        filteredRows: [],
        paginatedRows: [],
        selectedRows: new Set(),
        
        // Search & Filter
        searchQuery: '',
        searchColumns: config.searchColumns || [],
        
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
        
        // Computed Properties
        get allSelected() {
            return this.filteredRows.length > 0 && 
                   this.filteredRows.every(row => this.selectedRows.has(row.id));
        },
        
        get someSelected() {
            return this.selectedRows.size > 0 && 
                   this.selectedRows.size < this.filteredRows.length;
        },
        
        get selectedCount() {
            return this.selectedRows.size;
        },
        
        get hasNextPage() {
            return this.currentPage < this.totalPages;
        },
        
        get hasPrevPage() {
            return this.currentPage > 1;
        },
        
        // Initialization
        init() {
            this.filteredRows = [...this.rows];
            this.updatePagination();
            this.updateTotals();
            
            // Watch for external data updates
            this.$watch('rows', () => {
                this.handleSearch();
            });
        },
        
        // Search Functionality
        handleSearch() {
            if (!this.searchQuery.trim()) {
                this.filteredRows = [...this.rows];
            } else {
                this.filteredRows = this.rows.filter(row => {
                    return this.searchColumns.some(column => {
                        const value = this.getNestedValue(row, column);
                        return String(value || '').toLowerCase()
                            .includes(this.searchQuery.toLowerCase());
                    });
                });
            }
            
            this.currentPage = 1; // Reset to first page
            this.selectedRows.clear(); // Clear selections
            this.updatePagination();
            this.updateTotals();
        },
        
        // Sorting Functionality
        sort(column) {
            if (this.sortColumn === column) {
                this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortColumn = column;
                this.sortDirection = 'asc';
            }
            
            this.filteredRows.sort((a, b) => {
                const aVal = this.getNestedValue(a, column);
                const bVal = this.getNestedValue(b, column);
                
                if (aVal === bVal) return 0;
                
                const comparison = aVal < bVal ? -1 : 1;
                return this.sortDirection === 'asc' ? comparison : -comparison;
            });
            
            this.updatePagination();
        },
        
        // Pagination
        updatePagination() {
            this.totalRows = this.filteredRows.length;
            this.totalPages = Math.ceil(this.totalRows / this.pageSize);
            
            const start = (this.currentPage - 1) * this.pageSize;
            const end = start + this.pageSize;
            this.paginatedRows = this.filteredRows.slice(start, end);
        },
        
        updateTotals() {
            this.totalRows = this.filteredRows.length;
            this.totalPages = Math.ceil(this.totalRows / this.pageSize);
        },
        
        nextPage() {
            if (this.hasNextPage) {
                this.currentPage++;
                this.updatePagination();
            }
        },
        
        prevPage() {
            if (this.hasPrevPage) {
                this.currentPage--;
                this.updatePagination();
            }
        },
        
        goToPage(page) {
            if (page >= 1 && page <= this.totalPages) {
                this.currentPage = page;
                this.updatePagination();
            }
        },
        
        // Selection Management
        toggleSelectAll() {
            if (this.allSelected) {
                this.filteredRows.forEach(row => this.selectedRows.delete(row.id));
            } else {
                this.filteredRows.forEach(row => this.selectedRows.add(row.id));
            }
        },
        
        toggleRowSelection(rowId) {
            if (this.selectedRows.has(rowId)) {
                this.selectedRows.delete(rowId);
            } else {
                this.selectedRows.add(rowId);
            }
        },
        
        isRowSelected(rowId) {
            return this.selectedRows.has(rowId);
        },
        
        clearSelection() {
            this.selectedRows.clear();
        },
        
        // Bulk Actions
        async executeBulkAction(actionId) {
            const action = this.bulkActions.find(a => a.id === actionId);
            if (!action || this.selectedRows.size === 0) return;
            
            // Permission check
            if (action.permission && !this.hasPermission(action.permission)) {
                this.showError('Permission denied for this action');
                return;
            }
            
            // Confirmation dialog
            if (action.confirm) {
                const confirmed = await this.showConfirmation(
                    action.confirmText || `Are you sure you want to ${action.text.toLowerCase()} ${this.selectedRows.size} item(s)?`
                );
                if (!confirmed) return;
            }
            
            try {
                const selectedIds = Array.from(this.selectedRows);
                const response = await this.makeHTMXRequest(action.endpoint, {
                    method: action.method || 'POST',
                    data: {
                        action: actionId,
                        ids: selectedIds,
                        service: this.serviceContext
                    }
                });
                
                if (response.success) {
                    this.clearSelection();
                    this.showSuccess(response.message || `${action.text} completed successfully`);
                    
                    // Refresh data if needed
                    if (action.refreshOnSuccess !== false) {
                        this.refreshData();
                    }
                } else {
                    this.showError(response.message || `Failed to ${action.text.toLowerCase()}`);
                }
                
            } catch (error) {
                console.error('Bulk action failed:', error);
                this.showError(`Failed to ${action.text.toLowerCase()}: ${error.message}`);
            }
        },
        
        // Utility Methods
        getNestedValue(obj, path) {
            return path.split('.').reduce((current, key) => current?.[key], obj);
        },
        
        hasPermission(permission) {
            // Integration with global permission system
            return window.Alpine?.store?.('app')?.hasPermission?.(permission) ?? true;
        },
        
        async makeHTMXRequest(url, options) {
            // Integration with HTMX for server requests
            return new Promise((resolve, reject) => {
                const formData = new FormData();
                Object.entries(options.data || {}).forEach(([key, value]) => {
                    if (Array.isArray(value)) {
                        value.forEach(v => formData.append(`${key}[]`, v));
                    } else {
                        formData.append(key, value);
                    }
                });
                
                fetch(url, {
                    method: options.method || 'POST',
                    body: formData,
                    headers: {
                        'X-Requested-With': 'XMLHttpRequest',
                        'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.content
                    }
                })
                .then(response => response.json())
                .then(resolve)
                .catch(reject);
            });
        },
        
        showSuccess(message) {
            window.Alpine?.store?.('toast')?.success?.(message);
        },
        
        showError(message) {
            window.Alpine?.store?.('toast')?.error?.(message);
        },
        
        async showConfirmation(message) {
            return window.confirm(message); // Simple implementation, can be enhanced
        },
        
        refreshData() {
            // Trigger HTMX refresh or emit event for parent to handle
            this.$dispatch('datatable-refresh', { context: this.serviceContext });
        }
    };
}

/**
 * Service-specific DataTable configurations
 */
const DataTableConfigs = {
    console: {
        searchColumns: ['name', 'email', 'status', 'created_at'],
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
        bulkActions: [
            {
                id: 'invite',
                text: 'Send Invites',
                icon: 'fas fa-envelope',
                variant: 'primary',
                permission: 'user:invite',
                endpoint: '/workspace/bulk/invite',
                method: 'POST'
            },
            {
                id: 'deactivate',
                text: 'Deactivate',
                icon: 'fas fa-user-slash',
                variant: 'warning',
                permission: 'user:deactivate',
                confirm: true,
                endpoint: '/workspace/bulk/deactivate',
                method: 'POST'
            }
        ]
    },
    
    portal: {
        searchColumns: ['invoice_number', 'amount', 'status', 'due_date'],
        bulkActions: [
            {
                id: 'download',
                text: 'Download PDF',
                icon: 'fas fa-download',
                variant: 'secondary',
                permission: 'invoice:download',
                endpoint: '/portal/bulk/download',
                method: 'POST'
            },
            {
                id: 'pay',
                text: 'Mark as Paid',
                icon: 'fas fa-credit-card',
                variant: 'success',
                permission: 'invoice:pay',
                confirm: true,
                endpoint: '/portal/bulk/pay',
                method: 'POST'
            }
        ]
    }
};

/**
 * Convenience factories for service-specific DataTables
 */
window.createConsoleDataTable = (config = {}) => {
    return useDataTable({
        ...DataTableConfigs.console,
        ...config,
        serviceContext: 'console'
    });
};

window.createWorkspaceDataTable = (config = {}) => {
    return useDataTable({
        ...DataTableConfigs.workspace,
        ...config,
        serviceContext: 'workspace'
    });
};

window.createPortalDataTable = (config = {}) => {
    return useDataTable({
        ...DataTableConfigs.portal,
        ...config,
        serviceContext: 'portal'
    });
};

// Export for use in components
window.useDataTable = useDataTable;
window.DataTableConfigs = DataTableConfigs;