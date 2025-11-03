import { EventEmitter } from './utils/event-emitter';
import { StateManager } from './utils/state-manager';
import { FilterEngine } from './utils/filter-engine';
import { SortEngine } from './utils/sort-engine';
export class DataTable {
    constructor(options) {
        this.plugins = new Map();
        this.initialized = false;
        this.destroyed = false;
        this.options = this.normalizeOptions(options);
        this.eventEmitter = new EventEmitter();
        this.stateManager = new StateManager();
        this.filterEngine = new FilterEngine(this.options);
        this.sortEngine = new SortEngine(this.options);
        this.state = this.createInitialState();
        this.initialize();
    }
    normalizeOptions(options) {
        return {
            data: options.data || [],
            columns: options.columns,
            dataMode: options.dataMode || 'client',
            serverSide: {
                url: '',
                method: 'GET',
                headers: {},
                params: {},
                dataPath: 'data',
                totalPath: 'total',
                transformer: (response) => ({
                    data: response.data || response,
                    total: response.total || response.data?.length || 0,
                }),
                ...options.serverSide,
            },
            sorting: {
                enabled: true,
                multiSort: false,
                defaultSort: undefined,
                ...options.sorting,
            },
            filtering: {
                enabled: true,
                globalSearch: true,
                debounceMs: 300,
                caseSensitive: false,
                customFilters: {},
                ...options.filtering,
            },
            pagination: {
                enabled: true,
                pageSize: 10,
                pageSizeOptions: [5, 10, 25, 50, 100],
                showFirstLast: true,
                ...options.pagination,
            },
            selection: {
                enabled: false,
                multiple: true,
                mode: 'checkbox',
                selectOnRowClick: false,
                ...options.selection,
            },
            virtualScroll: {
                enabled: false,
                rowHeight: 48,
                overscan: 5,
                ...options.virtualScroll,
            },
            rowId: options.rowId || 'id',
            enableRowExpansion: options.enableRowExpansion || false,
            preserveState: options.preserveState || false,
            stateKey: options.stateKey || 'datatable-state',
            customComparators: options.customComparators || {},
            customFormatters: options.customFormatters || {},
            onInit: options.onInit || (() => { }),
            onDataChange: options.onDataChange || (() => { }),
            onError: options.onError || (() => { }),
        };
    }
    createInitialState() {
        const columns = this.options.columns.map((col) => ({
            visible: true,
            sortable: true,
            filterable: true,
            ...col,
        }));
        return {
            rows: [],
            columns,
            sortState: this.options.sorting.defaultSort || undefined,
            filterState: {
                filters: [],
                globalSearch: '',
            },
            paginationState: {
                pageIndex: 0,
                pageSize: this.options.pagination.pageSize,
                totalRows: 0,
                totalPages: 0,
            },
            selectionState: {
                selectedRowIds: new Set(),
                isAllSelected: false,
                isPartiallySelected: false,
            },
            loadState: {
                loading: false,
                error: null,
            },
            visibleRows: [],
            filteredRows: [],
            sortedRows: [],
            paginatedRows: [],
            visibleColumns: columns.filter((col) => col.visible !== false),
        };
    }
    initialize() {
        if (this.initialized)
            return;
        if (this.options.preserveState) {
            this.restoreState();
        }
        if (this.options.dataMode === 'client' && this.options.data.length > 0) {
            this.setData(this.options.data);
        }
        else if (this.options.dataMode === 'server') {
            this.reload();
        }
        this.initialized = true;
        this.emit('init');
        this.options.onInit();
    }
    setData(data) {
        this.ensureNotDestroyed();
        const rows = data.map((item) => ({
            id: this.getRowId(item),
            data: item,
            selected: false,
            expanded: false,
            disabled: false,
        }));
        this.state.rows = rows;
        this.state.paginationState.totalRows = rows.length;
        this.recompute();
        this.emit('data:change', rows);
        this.emit('data:load', rows);
        this.options.onDataChange(rows);
    }
    getData() {
        return [...this.state.rows];
    }
    addRow(data) {
        this.ensureNotDestroyed();
        const row = {
            id: this.getRowId(data),
            data,
            selected: false,
            expanded: false,
            disabled: false,
        };
        this.state.rows.push(row);
        this.state.paginationState.totalRows = this.state.rows.length;
        this.recompute();
        this.emit('data:change', this.state.rows);
    }
    updateRow(id, data) {
        this.ensureNotDestroyed();
        const row = this.state.rows.find((r) => r.id === id);
        if (!row) {
            console.warn(`DataTable: Row with id ${id} not found`);
            return;
        }
        row.data = { ...row.data, ...data };
        this.recompute();
        this.emit('data:change', this.state.rows);
    }
    deleteRow(id) {
        this.ensureNotDestroyed();
        const index = this.state.rows.findIndex((r) => r.id === id);
        if (index === -1) {
            console.warn(`DataTable: Row with id ${id} not found`);
            return;
        }
        this.state.rows.splice(index, 1);
        this.state.paginationState.totalRows = this.state.rows.length;
        this.state.selectionState.selectedRowIds.delete(id);
        this.recompute();
        this.emit('data:change', this.state.rows);
    }
    clearData() {
        this.ensureNotDestroyed();
        this.state.rows = [];
        this.state.paginationState.totalRows = 0;
        this.state.selectionState.selectedRowIds.clear();
        this.recompute();
        this.emit('data:change', []);
    }
    getColumns() {
        return [...this.state.columns];
    }
    setColumns(columns) {
        this.ensureNotDestroyed();
        this.state.columns = columns.map((col) => ({
            visible: true,
            sortable: true,
            filterable: true,
            ...col,
        }));
        this.state.visibleColumns = this.state.columns.filter((col) => col.visible !== false);
        this.emit('state:change', this.state);
    }
    showColumn(columnId) {
        this.toggleColumnVisibility(columnId, true);
    }
    hideColumn(columnId) {
        this.toggleColumnVisibility(columnId, false);
    }
    toggleColumn(columnId) {
        const column = this.state.columns.find((col) => col.id === columnId);
        if (column) {
            this.toggleColumnVisibility(columnId, !column.visible);
        }
    }
    toggleColumnVisibility(columnId, visible) {
        this.ensureNotDestroyed();
        const column = this.state.columns.find((col) => col.id === columnId);
        if (!column) {
            console.warn(`DataTable: Column ${columnId} not found`);
            return;
        }
        column.visible = visible;
        this.state.visibleColumns = this.state.columns.filter((col) => col.visible !== false);
        this.emit('column:visibility', { columnId, visible });
        this.emit('state:change', this.state);
    }
    sort(columnId, direction) {
        this.ensureNotDestroyed();
        if (!this.options.sorting.enabled) {
            console.warn('DataTable: Sorting is disabled');
            return;
        }
        const column = this.state.columns.find((col) => col.id === columnId);
        if (!column || column.sortable === false) {
            console.warn(`DataTable: Column ${columnId} is not sortable`);
            return;
        }
        let newDirection = direction ?? 'asc';
        if (!direction) {
            const currentSort = this.state.sortState;
            if (currentSort?.columnId === columnId) {
                if (currentSort.direction === 'asc') {
                    newDirection = 'desc';
                }
                else if (currentSort.direction === 'desc') {
                    newDirection = null;
                }
            }
        }
        this.state.sortState = newDirection ? { columnId, direction: newDirection } : undefined;
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('sort:change', this.state.sortState);
        this.emit('state:change', this.state);
    }
    multiSort(sorts) {
        this.ensureNotDestroyed();
        if (!this.options.sorting.enabled || !this.options.sorting.multiSort) {
            console.warn('DataTable: Multi-sort is not enabled');
            return;
        }
        this.state.sortState = { sorts };
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('sort:change', this.state.sortState);
        this.emit('state:change', this.state);
    }
    clearSort() {
        this.state.sortState = undefined;
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('sort:change', this.state.sortState);
        this.emit('state:change', this.state);
    }
    getSortState() {
        return this.state.sortState;
    }
    filter(filters) {
        this.ensureNotDestroyed();
        if (!this.options.filtering.enabled) {
            console.warn('DataTable: Filtering is disabled');
            return;
        }
        this.state.filterState.filters = filters;
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('filter:change', this.state.filterState);
        this.emit('state:change', this.state);
    }
    addFilter(filter) {
        this.ensureNotDestroyed();
        this.state.filterState.filters = this.state.filterState.filters.filter((f) => f.columnId !== filter.columnId);
        this.state.filterState.filters.push(filter);
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('filter:change', this.state.filterState);
        this.emit('state:change', this.state);
    }
    removeFilter(columnId) {
        this.ensureNotDestroyed();
        this.state.filterState.filters = this.state.filterState.filters.filter((f) => f.columnId !== columnId);
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('filter:change', this.state.filterState);
        this.emit('state:change', this.state);
    }
    clearFilters() {
        this.state.filterState.filters = [];
        this.state.filterState.globalSearch = '';
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('filter:change', this.state.filterState);
        this.emit('state:change', this.state);
    }
    search(query) {
        this.ensureNotDestroyed();
        if (!this.options.filtering.enabled || !this.options.filtering.globalSearch) {
            console.warn('DataTable: Global search is not enabled');
            return;
        }
        this.state.filterState.globalSearch = query;
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('search:change', query);
        this.emit('filter:change', this.state.filterState);
        this.emit('state:change', this.state);
    }
    getFilterState() {
        return { ...this.state.filterState };
    }
    goToPage(page) {
        this.ensureNotDestroyed();
        if (!this.options.pagination.enabled) {
            console.warn('DataTable: Pagination is disabled');
            return;
        }
        const totalPages = this.state.paginationState.totalPages;
        const newPage = Math.max(0, Math.min(page, totalPages - 1));
        if (newPage === this.state.paginationState.pageIndex)
            return;
        this.state.paginationState.pageIndex = newPage;
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('page:change', this.state.paginationState);
        this.emit('state:change', this.state);
    }
    nextPage() {
        this.goToPage(this.state.paginationState.pageIndex + 1);
    }
    previousPage() {
        this.goToPage(this.state.paginationState.pageIndex - 1);
    }
    firstPage() {
        this.goToPage(0);
    }
    lastPage() {
        this.goToPage(this.state.paginationState.totalPages - 1);
    }
    setPageSize(size) {
        this.ensureNotDestroyed();
        if (!this.options.pagination.enabled) {
            console.warn('DataTable: Pagination is disabled');
            return;
        }
        this.state.paginationState.pageSize = size;
        this.state.paginationState.pageIndex = 0;
        if (this.options.dataMode === 'client') {
            this.recompute();
        }
        else {
            this.reload();
        }
        this.emit('page:change', this.state.paginationState);
        this.emit('state:change', this.state);
    }
    getPaginationState() {
        return { ...this.state.paginationState };
    }
    selectRow(id) {
        this.ensureNotDestroyed();
        if (!this.options.selection.enabled) {
            console.warn('DataTable: Selection is disabled');
            return;
        }
        if (!this.options.selection.multiple) {
            this.state.selectionState.selectedRowIds.clear();
        }
        const row = this.state.rows.find((r) => r.id === id);
        if (!row || row.disabled)
            return;
        row.selected = true;
        this.state.selectionState.selectedRowIds.add(id);
        this.updateSelectionState();
        this.emit('row:select', { row, selected: true });
        this.emit('selection:change', this.state.selectionState);
        this.emit('state:change', this.state);
    }
    deselectRow(id) {
        this.ensureNotDestroyed();
        const row = this.state.rows.find((r) => r.id === id);
        if (!row)
            return;
        row.selected = false;
        this.state.selectionState.selectedRowIds.delete(id);
        this.updateSelectionState();
        this.emit('row:select', { row, selected: false });
        this.emit('selection:change', this.state.selectionState);
        this.emit('state:change', this.state);
    }
    toggleRowSelection(id) {
        if (this.state.selectionState.selectedRowIds.has(id)) {
            this.deselectRow(id);
        }
        else {
            this.selectRow(id);
        }
    }
    selectAll() {
        this.ensureNotDestroyed();
        if (!this.options.selection.enabled || !this.options.selection.multiple) {
            console.warn('DataTable: Multi-selection is not enabled');
            return;
        }
        this.state.visibleRows.forEach((row) => {
            if (!row.disabled) {
                row.selected = true;
                this.state.selectionState.selectedRowIds.add(row.id);
            }
        });
        this.updateSelectionState();
        this.emit('selection:change', this.state.selectionState);
        this.emit('state:change', this.state);
    }
    deselectAll() {
        this.ensureNotDestroyed();
        this.state.rows.forEach((row) => {
            row.selected = false;
        });
        this.state.selectionState.selectedRowIds.clear();
        this.updateSelectionState();
        this.emit('selection:change', this.state.selectionState);
        this.emit('state:change', this.state);
    }
    getSelectedRows() {
        return this.state.rows.filter((row) => this.state.selectionState.selectedRowIds.has(row.id));
    }
    getSelectionState() {
        return {
            selectedRowIds: new Set(this.state.selectionState.selectedRowIds),
            isAllSelected: this.state.selectionState.isAllSelected,
            isPartiallySelected: this.state.selectionState.isPartiallySelected,
        };
    }
    updateSelectionState() {
        const selectableRows = this.state.visibleRows.filter((row) => !row.disabled);
        const selectedCount = this.state.selectionState.selectedRowIds.size;
        this.state.selectionState.isAllSelected =
            selectableRows.length > 0 && selectedCount === selectableRows.length;
        this.state.selectionState.isPartiallySelected =
            selectedCount > 0 && selectedCount < selectableRows.length;
    }
    getState() {
        return { ...this.state };
    }
    setState(newState) {
        this.ensureNotDestroyed();
        this.state = { ...this.state, ...newState };
        this.emit('state:change', this.state);
    }
    resetState() {
        this.state = this.createInitialState();
        this.clearFilters();
        this.clearSort();
        this.deselectAll();
        this.firstPage();
        this.emit('state:change', this.state);
    }
    saveState() {
        if (!this.options.preserveState)
            return;
        this.stateManager.save(this.options.stateKey, this.state);
    }
    restoreState() {
        const savedState = this.stateManager.load(this.options.stateKey);
        if (savedState) {
            this.state = { ...this.state, ...savedState };
        }
    }
    async reload() {
        if (this.options.dataMode !== 'server') {
            console.warn('DataTable: Server-side mode is not enabled');
            return;
        }
        if (!this.options.serverSide.url) {
            console.error('DataTable: Server URL is not configured');
            return;
        }
        this.state.loadState.loading = true;
        this.state.loadState.error = null;
        this.emit('state:change', this.state);
        try {
            const params = this.buildServerParams();
            const url = new URL(this.options.serverSide.url);
            if (this.options.serverSide.method === 'GET') {
                Object.entries(params).forEach(([key, value]) => {
                    url.searchParams.append(key, String(value));
                });
            }
            const response = await fetch(url.toString(), {
                method: this.options.serverSide.method,
                headers: {
                    'Content-Type': 'application/json',
                    ...this.options.serverSide.headers,
                },
                body: this.options.serverSide.method === 'POST' ? JSON.stringify(params) : undefined,
            });
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const json = await response.json();
            const { data, total } = this.options.serverSide.transformer(json);
            this.setData(data);
            this.state.paginationState.totalRows = total;
            this.state.paginationState.totalPages = Math.ceil(total / this.state.paginationState.pageSize);
            this.state.loadState.lastFetch = new Date();
            this.emit('data:load', this.state.rows);
        }
        catch (error) {
            this.state.loadState.error = error;
            this.emit('data:error', error);
            this.options.onError(error);
        }
        finally {
            this.state.loadState.loading = false;
            this.emit('state:change', this.state);
        }
    }
    buildServerParams() {
        const params = {
            ...this.options.serverSide.params,
        };
        if (this.options.pagination.enabled) {
            params.page = this.state.paginationState.pageIndex;
            params.pageSize = this.state.paginationState.pageSize;
        }
        if (this.state.sortState) {
            if ('sorts' in this.state.sortState) {
                params.sort = this.state.sortState.sorts;
            }
            else {
                params.sortBy = this.state.sortState.columnId;
                params.sortDir = this.state.sortState.direction;
            }
        }
        if (this.state.filterState.filters.length > 0) {
            params.filters = this.state.filterState.filters;
        }
        if (this.state.filterState.globalSearch) {
            params.search = this.state.filterState.globalSearch;
        }
        return params;
    }
    recompute() {
        let rows = [...this.state.rows];
        if (this.options.filtering.enabled) {
            rows = this.filterEngine.filter(rows, this.state.filterState, this.state.columns);
        }
        this.state.filteredRows = rows;
        if (this.options.sorting.enabled && this.state.sortState) {
            rows = this.sortEngine.sort(rows, this.state.sortState, this.state.columns);
        }
        this.state.sortedRows = rows;
        this.state.paginationState.totalRows = rows.length;
        this.state.paginationState.totalPages = Math.ceil(rows.length / this.state.paginationState.pageSize);
        if (this.options.pagination.enabled) {
            const start = this.state.paginationState.pageIndex * this.state.paginationState.pageSize;
            const end = start + this.state.paginationState.pageSize;
            rows = rows.slice(start, end);
        }
        this.state.paginatedRows = rows;
        this.state.visibleRows = rows;
        this.updateSelectionState();
        if (this.options.preserveState) {
            this.saveState();
        }
        this.emit('state:change', this.state);
    }
    on(event, callback) {
        return this.eventEmitter.on(event, callback);
    }
    off(event, callback) {
        this.eventEmitter.off(event, callback);
    }
    emit(event, data) {
        this.eventEmitter.emit(event, data);
    }
    use(plugin) {
        if (this.plugins.has(plugin.name)) {
            console.warn(`DataTable: Plugin ${plugin.name} is already installed`);
            return;
        }
        plugin.install(this);
        this.plugins.set(plugin.name, plugin);
    }
    getRowId(row) {
        if (typeof this.options.rowId === 'function') {
            return this.options.rowId(row);
        }
        return row[this.options.rowId];
    }
    ensureNotDestroyed() {
        if (this.destroyed) {
            throw new Error('DataTable: Instance has been destroyed');
        }
    }
    destroy() {
        if (this.destroyed)
            return;
        this.emit('destroy');
        this.eventEmitter.clear();
        this.plugins.forEach((plugin) => {
            if (plugin.uninstall) {
                plugin.uninstall(this);
            }
        });
        this.plugins.clear();
        this.destroyed = true;
    }
}
export function createDataTable(options) {
    return new DataTable(options);
}
//# sourceMappingURL=core.js.map