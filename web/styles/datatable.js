// DataTable Alpine.js component functions
function dataTable(config) {
    return {
        // State
        loading: false,
        searchTerm: '',
        selectedRows: [],
        allSelected: false,
        activeFilter: 'all',
        sortColumn: '',
        sortDirection: 'asc',
        rows: [],
        
        // Config
        searchUrl: config.searchUrl || '',
        filterUrl: config.filterUrl || '',
        sortUrl: config.sortUrl || '',
        
        // Initialize component
        initTable() {
            // Load initial data if URLs are provided
            this.loadInitialData();
            
            // Set up keyboard shortcuts
            this.setupKeyboardShortcuts();
        },
        
        // Load initial table data
        loadInitialData() {
            // Implementation depends on whether server-side processing is used
            // For client-side tables, rows are already loaded from the template
            // For server-side tables, we would fetch data here
        },
        
        // Search functionality
        performSearch() {
            if (!this.searchUrl) {
                // Client-side search
                this.filterRowsLocally();
                return;
            }
            
            // Server-side search
            this.loading = true;
            
            htmx.ajax('GET', this.searchUrl, {
                values: { 
                    search: this.searchTerm,
                    filter: this.activeFilter,
                    sort: this.sortColumn,
                    direction: this.sortDirection
                },
                target: '#table-body',
                swap: 'innerHTML'
            }).then(() => {
                this.loading = false;
            }).catch(() => {
                this.loading = false;
                Alpine.store('app').addNotification('error', 'Failed to search data');
            });
        },
        
        // Client-side filtering (fallback when no server-side search)
        filterRowsLocally() {
            // This would filter the rows array based on searchTerm
            // Implementation depends on how data is structured
        },
        
        // Filter functionality
        applyFilter() {
            if (!this.filterUrl) {
                this.filterRowsLocally();
                return;
            }
            
            this.loading = true;
            
            htmx.ajax('GET', this.filterUrl, {
                values: { 
                    filter: this.activeFilter,
                    search: this.searchTerm,
                    sort: this.sortColumn,
                    direction: this.sortDirection
                },
                target: '#table-body',
                swap: 'innerHTML'
            }).then(() => {
                this.loading = false;
            }).catch(() => {
                this.loading = false;
                Alpine.store('app').addNotification('error', 'Failed to filter data');
            });
        },
        
        // Get active filter label
        getActiveFilterLabel() {
            const filters = {
                'all': 'All Items',
                'active': 'Active',
                'inactive': 'Inactive',
                'recent': 'Recent',
                'archived': 'Archived'
            };
            return filters[this.activeFilter] || 'Filter';
        },
        
        // Sorting functionality
        toggleSort(column) {
            if (this.sortColumn === column) {
                this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortColumn = column;
                this.sortDirection = 'asc';
            }
            
            if (!this.sortUrl) {
                this.sortRowsLocally();
                return;
            }
            
            // Server-side sorting
            this.loading = true;
            
            htmx.ajax('GET', this.sortUrl, {
                values: { 
                    sort: this.sortColumn,
                    direction: this.sortDirection,
                    search: this.searchTerm,
                    filter: this.activeFilter
                },
                target: '#table-body',
                swap: 'innerHTML'
            }).then(() => {
                this.loading = false;
            }).catch(() => {
                this.loading = false;
                Alpine.store('app').addNotification('error', 'Failed to sort data');
            });
        },
        
        // Client-side sorting (fallback)
        sortRowsLocally() {
            // Implementation for local sorting
        },
        
        // Selection functionality
        toggleAll() {
            if (this.allSelected) {
                // Select all visible rows
                const checkboxes = document.querySelectorAll('input[type="checkbox"][id^="checkbox-table-"]');
                checkboxes.forEach(checkbox => {
                    checkbox.checked = true;
                    const rowId = this.getRowIdFromCheckbox(checkbox);
                    if (rowId && !this.selectedRows.includes(rowId)) {
                        this.selectedRows.push(rowId);
                    }
                });
            } else {
                // Deselect all
                this.selectedRows = [];
                const checkboxes = document.querySelectorAll('input[type="checkbox"][id^="checkbox-table-"]');
                checkboxes.forEach(checkbox => {
                    checkbox.checked = false;
                });
            }
        },
        
        // Update selected rows when individual checkbox is changed
        updateSelectedRows(rowId, selected) {
            if (selected) {
                if (!this.selectedRows.includes(rowId)) {
                    this.selectedRows.push(rowId);
                }
            } else {
                this.selectedRows = this.selectedRows.filter(id => id !== rowId);
                this.allSelected = false;
            }
            
            // Update "select all" checkbox state
            const totalCheckboxes = document.querySelectorAll('input[type="checkbox"][id^="checkbox-table-"]').length;
            this.allSelected = this.selectedRows.length === totalCheckboxes && totalCheckboxes > 0;
        },
        
        // Get row ID from checkbox element
        getRowIdFromCheckbox(checkbox) {
            const row = checkbox.closest('tr');
            return row ? row.getAttribute('data-row-id') : null;
        },
        
        // Bulk actions
        executeBulkAction(action) {
            if (this.selectedRows.length === 0) {
                Alpine.store('app').addNotification('warning', 'Please select items first');
                return;
            }
            
            // Handle different bulk actions
            switch (action) {
                case 'delete':
                    this.confirmBulkDelete();
                    break;
                case 'export':
                    this.exportSelected();
                    break;
                case 'archive':
                    this.archiveSelected();
                    break;
                default:
                    // Custom bulk action
                    this.executeCustomBulkAction(action);
            }
        },
        
        // Confirm bulk delete
        confirmBulkDelete() {
            if (confirm(`Are you sure you want to delete ${this.selectedRows.length} items?`)) {
                this.performBulkAction('delete');
            }
        },
        
        // Export selected items
        exportSelected() {
            const params = new URLSearchParams();
            this.selectedRows.forEach(id => params.append('ids[]', id));
            
            // Create download link
            const downloadUrl = `/export?${params.toString()}`;
            window.open(downloadUrl, '_blank');
        },
        
        // Archive selected items
        archiveSelected() {
            this.performBulkAction('archive');
        },
        
        // Execute custom bulk action
        executeCustomBulkAction(action) {
            this.performBulkAction(action);
        },
        
        // Perform bulk action via HTMX
        performBulkAction(action) {
            this.loading = true;
            
            htmx.ajax('POST', `/bulk-actions/${action}`, {
                values: { 
                    ids: this.selectedRows,
                    action: action
                },
                headers: {
                    'X-CSRF-Token': document.querySelector('meta[name=csrf-token]').content
                },
                target: '#table-body',
                swap: 'innerHTML'
            }).then(() => {
                this.loading = false;
                this.selectedRows = [];
                this.allSelected = false;
                Alpine.store('app').addNotification('success', `Bulk ${action} completed successfully`);
            }).catch(() => {
                this.loading = false;
                Alpine.store('app').addNotification('error', `Failed to perform bulk ${action}`);
            });
        },
        
        // Row actions
        executeRowAction(action, rowId) {
            switch (action) {
                case 'delete':
                    this.confirmRowDelete(rowId);
                    break;
                case 'duplicate':
                    this.duplicateRow(rowId);
                    break;
                default:
                    this.performRowAction(action, rowId);
            }
        },
        
        // Confirm row deletion
        confirmRowDelete(rowId) {
            if (confirm('Are you sure you want to delete this item?')) {
                this.performRowAction('delete', rowId);
            }
        },
        
        // Duplicate row
        duplicateRow(rowId) {
            this.performRowAction('duplicate', rowId);
        },
        
        // Perform row action via HTMX
        performRowAction(action, rowId) {
            this.loading = true;
            
            htmx.ajax('POST', `/row-actions/${action}`, {
                values: { 
                    id: rowId,
                    action: action
                },
                headers: {
                    'X-CSRF-Token': document.querySelector('meta[name=csrf-token]').content
                },
                target: '#table-body',
                swap: 'innerHTML'
            }).then(() => {
                this.loading = false;
                Alpine.store('app').addNotification('success', `${action} completed successfully`);
            }).catch(() => {
                this.loading = false;
                Alpine.store('app').addNotification('error', `Failed to perform ${action}`);
            });
        },
        
        // Open action modal
        openActionModal(modalId, rowId) {
            // Load modal content with row data
            htmx.ajax('GET', `/modals/${modalId}`, {
                values: { id: rowId },
                target: '#modal-container',
                swap: 'innerHTML'
            });
        },
        
        // Pagination
        goToPage(page) {
            this.loading = true;
            
            const currentUrl = new URL(window.location);
            currentUrl.searchParams.set('page', page);
            
            htmx.ajax('GET', currentUrl.toString(), {
                target: '#table-body',
                swap: 'innerHTML'
            }).then(() => {
                this.loading = false;
                
                // Update URL without page reload
                window.history.pushState({}, '', currentUrl.toString());
            }).catch(() => {
                this.loading = false;
                Alpine.store('app').addNotification('error', 'Failed to load page');
            });
        },
        
        // Keyboard shortcuts
        setupKeyboardShortcuts() {
            document.addEventListener('keydown', (e) => {
                // Only handle shortcuts when not in an input field
                if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
                    return;
                }
                
                switch (e.key) {
                    case 'f':
                        if (e.ctrlKey || e.metaKey) {
                            e.preventDefault();
                            document.getElementById('table-search')?.focus();
                        }
                        break;
                    case 'a':
                        if (e.ctrlKey || e.metaKey) {
                            e.preventDefault();
                            this.allSelected = !this.allSelected;
                            this.toggleAll();
                        }
                        break;
                    case 'Escape':
                        // Clear selection
                        this.selectedRows = [];
                        this.allSelected = false;
                        document.querySelectorAll('input[type="checkbox"]').forEach(cb => cb.checked = false);
                        break;
                }
            });
        }
    }
}

// Global functions for external use
window.refreshTable = function() {
    // Refresh the current table
    window.location.reload();
};

window.clearTableFilters = function() {
    // Reset all filters
    const tableComponent = document.querySelector('[x-data]').__x?.$data;
    if (tableComponent) {
        tableComponent.searchTerm = '';
        tableComponent.activeFilter = 'all';
        tableComponent.sortColumn = '';
        tableComponent.sortDirection = 'asc';
        tableComponent.performSearch();
    }
};