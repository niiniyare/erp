/**
 * Awo DataTable - Core Implementation
 * Headless DataTable with complete state management
 * 
 * @package @awo/datatable
 * @author Awo ERP
 * @version 2.0.0
 */

import type {
  DataTableCore,
  DataTableOptions,
  DataTableState,
  DataTableEvents,
  Row,
  RowData,
  RowId,
  Column,
  SortState,
  MultiSortState,
  SortDirection,
  Filter,
  FilterState,
  PaginationState,
  SelectionState,
  EventCallback,
  Plugin,
} from './types';

import { EventEmitter } from './utils/event-emitter';
import { StateManager } from './utils/state-manager';
import { FilterEngine } from './utils/filter-engine';
import { SortEngine } from './utils/sort-engine';

/**
 * Main DataTable Class
 * Provides headless table logic without any UI assumptions
 */
export class DataTable<T extends RowData = RowData> implements DataTableCore<T> {
  private options: Required<DataTableOptions<T>>;
  private state: DataTableState<T>;
  private eventEmitter: EventEmitter<Record<string, (...args: any[]) => void>>;
  private stateManager: StateManager<DataTableState<T>>;
  private filterEngine: FilterEngine<T>;
  private sortEngine: SortEngine<T>;
  private plugins: Map<string, Plugin<T>> = new Map();
  private initialized = false;
  private destroyed = false;

  constructor(options: DataTableOptions<T>) {
    this.options = this.normalizeOptions(options);
    this.eventEmitter = new EventEmitter();
    this.stateManager = new StateManager();
    this.filterEngine = new FilterEngine(this.options);
    this.sortEngine = new SortEngine(this.options);

    // Initialize state
    this.state = this.createInitialState();

    // Setup
    this.initialize();
  }

  // ============================================================================
  // Initialization
  // ============================================================================

  private normalizeOptions(options: DataTableOptions<T>): Required<DataTableOptions<T>> {
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
        transformer: (response: any) => ({
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
      onInit: options.onInit || (() => {}),
      onDataChange: options.onDataChange || (() => {}),
      onError: options.onError || (() => {}),
    };
  }

  private createInitialState(): DataTableState<T> {
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

  private initialize(): void {
    if (this.initialized) return;

    // Restore state if enabled
    if (this.options.preserveState) {
      this.restoreState();
    }

    // Load initial data
    if (this.options.dataMode === 'client' && this.options.data.length > 0) {
      this.setData(this.options.data);
    } else if (this.options.dataMode === 'server') {
      this.reload();
    }

    this.initialized = true;
    this.emit('init');
    this.options.onInit();
  }

  // ============================================================================
  // Data Management
  // ============================================================================

  setData(data: T[]): void {
    this.ensureNotDestroyed();

    const rows: Row<T>[] = data.map((item) => ({
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

  getData(): Row<T>[] {
    return [...this.state.rows];
  }

  addRow(data: T): void {
    this.ensureNotDestroyed();

    const row: Row<T> = {
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

  updateRow(id: RowId, data: Partial<T>): void {
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

  deleteRow(id: RowId): void {
    this.ensureNotDestroyed();

    const index = this.state.rows.findIndex((r) => r.id === id);
    if (index === -1) {
      console.warn(`DataTable: Row with id ${id} not found`);
      return;
    }

    this.state.rows.splice(index, 1);
    this.state.paginationState.totalRows = this.state.rows.length;
    
    // Remove from selection if selected
    this.state.selectionState.selectedRowIds.delete(id);
    
    this.recompute();
    this.emit('data:change', this.state.rows);
  }

  clearData(): void {
    this.ensureNotDestroyed();

    this.state.rows = [];
    this.state.paginationState.totalRows = 0;
    this.state.selectionState.selectedRowIds.clear();
    this.recompute();

    this.emit('data:change', []);
  }

  // ============================================================================
  // Columns
  // ============================================================================

  getColumns(): Column<T>[] {
    return [...this.state.columns];
  }

  setColumns(columns: Column<T>[]): void {
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

  showColumn(columnId: string): void {
    this.toggleColumnVisibility(columnId, true);
  }

  hideColumn(columnId: string): void {
    this.toggleColumnVisibility(columnId, false);
  }

  toggleColumn(columnId: string): void {
    const column = this.state.columns.find((col) => col.id === columnId);
    if (column) {
      this.toggleColumnVisibility(columnId, !column.visible);
    }
  }

  private toggleColumnVisibility(columnId: string, visible: boolean): void {
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

  // ============================================================================
  // Sorting
  // ============================================================================

  sort(columnId: string, direction?: SortDirection): void {
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

    // Cycle through: asc -> desc -> null
    let newDirection: SortDirection = direction ?? 'asc';
    if (!direction) {
      const currentSort = this.state.sortState as SortState | null;
      if (currentSort?.columnId === columnId) {
        if (currentSort.direction === 'asc') {
          newDirection = 'desc';
        } else if (currentSort.direction === 'desc') {
          newDirection = null;
        }
      }
    }

    this.state.sortState = newDirection ? { columnId, direction: newDirection } : undefined;
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('sort:change', this.state.sortState);
    this.emit('state:change', this.state);
  }

  multiSort(sorts: SortState[]): void {
    this.ensureNotDestroyed();

    if (!this.options.sorting.enabled || !this.options.sorting.multiSort) {
      console.warn('DataTable: Multi-sort is not enabled');
      return;
    }

    this.state.sortState = { sorts };
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('sort:change', this.state.sortState);
    this.emit('state:change', this.state);
  }

  clearSort(): void {
    this.state.sortState = undefined;
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('sort:change', this.state.sortState);
    this.emit('state:change', this.state);
  }

  getSortState(): SortState | MultiSortState | undefined {
    return this.state.sortState;
  }

  // ============================================================================
  // Filtering
  // ============================================================================

  filter(filters: Filter[]): void {
    this.ensureNotDestroyed();

    if (!this.options.filtering.enabled) {
      console.warn('DataTable: Filtering is disabled');
      return;
    }

    this.state.filterState.filters = filters;
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('filter:change', this.state.filterState);
    this.emit('state:change', this.state);
  }

  addFilter(filter: Filter): void {
    this.ensureNotDestroyed();

    // Remove existing filter for the same column
    this.state.filterState.filters = this.state.filterState.filters.filter(
      (f) => f.columnId !== filter.columnId
    );

    this.state.filterState.filters.push(filter);
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('filter:change', this.state.filterState);
    this.emit('state:change', this.state);
  }

  removeFilter(columnId: string): void {
    this.ensureNotDestroyed();

    this.state.filterState.filters = this.state.filterState.filters.filter(
      (f) => f.columnId !== columnId
    );
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('filter:change', this.state.filterState);
    this.emit('state:change', this.state);
  }

  clearFilters(): void {
    this.state.filterState.filters = [];
    this.state.filterState.globalSearch = '';
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('filter:change', this.state.filterState);
    this.emit('state:change', this.state);
  }

  search(query: string): void {
    this.ensureNotDestroyed();

    if (!this.options.filtering.enabled || !this.options.filtering.globalSearch) {
      console.warn('DataTable: Global search is not enabled');
      return;
    }

    this.state.filterState.globalSearch = query;
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('search:change', query);
    this.emit('filter:change', this.state.filterState);
    this.emit('state:change', this.state);
  }

  getFilterState(): FilterState {
    return { ...this.state.filterState };
  }

  // ============================================================================
  // Pagination
  // ============================================================================

  goToPage(page: number): void {
    this.ensureNotDestroyed();

    if (!this.options.pagination.enabled) {
      console.warn('DataTable: Pagination is disabled');
      return;
    }

    const totalPages = this.state.paginationState.totalPages;
    const newPage = Math.max(0, Math.min(page, totalPages - 1));

    if (newPage === this.state.paginationState.pageIndex) return;

    this.state.paginationState.pageIndex = newPage;
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('page:change', this.state.paginationState);
    this.emit('state:change', this.state);
  }

  nextPage(): void {
    this.goToPage(this.state.paginationState.pageIndex + 1);
  }

  previousPage(): void {
    this.goToPage(this.state.paginationState.pageIndex - 1);
  }

  firstPage(): void {
    this.goToPage(0);
  }

  lastPage(): void {
    this.goToPage(this.state.paginationState.totalPages - 1);
  }

  setPageSize(size: number): void {
    this.ensureNotDestroyed();

    if (!this.options.pagination.enabled) {
      console.warn('DataTable: Pagination is disabled');
      return;
    }

    this.state.paginationState.pageSize = size;
    this.state.paginationState.pageIndex = 0; // Reset to first page
    
    if (this.options.dataMode === 'client') {
      this.recompute();
    } else {
      this.reload();
    }

    this.emit('page:change', this.state.paginationState);
    this.emit('state:change', this.state);
  }

  getPaginationState(): PaginationState {
    return { ...this.state.paginationState };
  }

  // ============================================================================
  // Selection
  // ============================================================================

  selectRow(id: RowId): void {
    this.ensureNotDestroyed();

    if (!this.options.selection.enabled) {
      console.warn('DataTable: Selection is disabled');
      return;
    }

    if (!this.options.selection.multiple) {
      this.state.selectionState.selectedRowIds.clear();
    }

    const row = this.state.rows.find((r) => r.id === id);
    if (!row || row.disabled) return;

    row.selected = true;
    this.state.selectionState.selectedRowIds.add(id);
    this.updateSelectionState();

    this.emit('row:select', { row, selected: true });
    this.emit('selection:change', this.state.selectionState);
    this.emit('state:change', this.state);
  }

  deselectRow(id: RowId): void {
    this.ensureNotDestroyed();

    const row = this.state.rows.find((r) => r.id === id);
    if (!row) return;

    row.selected = false;
    this.state.selectionState.selectedRowIds.delete(id);
    this.updateSelectionState();

    this.emit('row:select', { row, selected: false });
    this.emit('selection:change', this.state.selectionState);
    this.emit('state:change', this.state);
  }

  toggleRowSelection(id: RowId): void {
    if (this.state.selectionState.selectedRowIds.has(id)) {
      this.deselectRow(id);
    } else {
      this.selectRow(id);
    }
  }

  selectAll(): void {
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

  deselectAll(): void {
    this.ensureNotDestroyed();

    this.state.rows.forEach((row) => {
      row.selected = false;
    });

    this.state.selectionState.selectedRowIds.clear();
    this.updateSelectionState();

    this.emit('selection:change', this.state.selectionState);
    this.emit('state:change', this.state);
  }

  getSelectedRows(): Row<T>[] {
    return this.state.rows.filter((row) =>
      this.state.selectionState.selectedRowIds.has(row.id)
    );
  }

  getSelectionState(): SelectionState {
    return {
      selectedRowIds: new Set(this.state.selectionState.selectedRowIds),
      isAllSelected: this.state.selectionState.isAllSelected,
      isPartiallySelected: this.state.selectionState.isPartiallySelected,
    };
  }

  private updateSelectionState(): void {
    const selectableRows = this.state.visibleRows.filter((row) => !row.disabled);
    const selectedCount = this.state.selectionState.selectedRowIds.size;

    this.state.selectionState.isAllSelected =
      selectableRows.length > 0 && selectedCount === selectableRows.length;
    this.state.selectionState.isPartiallySelected =
      selectedCount > 0 && selectedCount < selectableRows.length;
  }

  // ============================================================================
  // State Management
  // ============================================================================

  getState(): DataTableState<T> {
    return { ...this.state };
  }

  setState(newState: Partial<DataTableState<T>>): void {
    this.ensureNotDestroyed();
    this.state = { ...this.state, ...newState };
    this.emit('state:change', this.state);
  }

  resetState(): void {
    this.state = this.createInitialState();
    this.clearFilters();
    this.clearSort();
    this.deselectAll();
    this.firstPage();
    this.emit('state:change', this.state);
  }

  private saveState(): void {
    if (!this.options.preserveState) return;
    this.stateManager.save(this.options.stateKey, this.state);
  }

  private restoreState(): void {
    const savedState = this.stateManager.load<DataTableState<T>>(this.options.stateKey);
    if (savedState) {
      this.state = { ...this.state, ...savedState };
    }
  }

  // ============================================================================
  // Server-side Operations
  // ============================================================================

  async reload(): Promise<void> {
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
      const { data, total } = this.options.serverSide.transformer!(json);

      this.setData(data as T[]);
      this.state.paginationState.totalRows = total;
      this.state.paginationState.totalPages = Math.ceil(total / this.state.paginationState.pageSize);

      this.state.loadState.lastFetch = new Date();
      this.emit('data:load', this.state.rows);
    } catch (error) {
      this.state.loadState.error = error as Error;
      this.emit('data:error', error as Error);
      this.options.onError(error as Error);
    } finally {
      this.state.loadState.loading = false;
      this.emit('state:change', this.state);
    }
  }

  private buildServerParams(): Record<string, any> {
    const params: Record<string, any> = {
      ...this.options.serverSide.params,
    };

    // Pagination
    if (this.options.pagination.enabled) {
      params.page = this.state.paginationState.pageIndex;
      params.pageSize = this.state.paginationState.pageSize;
    }

    // Sorting
    if (this.state.sortState) {
      if ('sorts' in this.state.sortState) {
        params.sort = this.state.sortState.sorts;
      } else {
        params.sortBy = this.state.sortState.columnId;
        params.sortDir = this.state.sortState.direction;
      }
    }

    // Filtering
    if (this.state.filterState.filters.length > 0) {
      params.filters = this.state.filterState.filters;
    }

    // Global search
    if (this.state.filterState.globalSearch) {
      params.search = this.state.filterState.globalSearch;
    }

    return params;
  }

  // ============================================================================
  // Computation Pipeline
  // ============================================================================

  private recompute(): void {
    // 1. Start with all rows
    let rows = [...this.state.rows];

    // 2. Apply filters
    if (this.options.filtering.enabled) {
      rows = this.filterEngine.filter(rows, this.state.filterState, this.state.columns);
    }
    this.state.filteredRows = rows;

    // 3. Apply sorting
    if (this.options.sorting.enabled && this.state.sortState) {
      rows = this.sortEngine.sort(rows, this.state.sortState, this.state.columns);
    }
    this.state.sortedRows = rows;

    // 4. Update pagination metadata
    this.state.paginationState.totalRows = rows.length;
    this.state.paginationState.totalPages = Math.ceil(
      rows.length / this.state.paginationState.pageSize
    );

    // 5. Apply pagination
    if (this.options.pagination.enabled) {
      const start = this.state.paginationState.pageIndex * this.state.paginationState.pageSize;
      const end = start + this.state.paginationState.pageSize;
      rows = rows.slice(start, end);
    }
    this.state.paginatedRows = rows;

    // 6. Final visible rows
    this.state.visibleRows = rows;

    // Update selection state
    this.updateSelectionState();

    // Save state if enabled
    if (this.options.preserveState) {
      this.saveState();
    }

    this.emit('state:change', this.state);
  }

  // ============================================================================
  // Event System
  // ============================================================================

  on<K extends keyof DataTableEvents<T>>(
    event: K,
    callback: EventCallback<Parameters<DataTableEvents<T>[K]>[0]>
  ): () => void {
    return this.eventEmitter.on(event, callback);
  }

  off<K extends keyof DataTableEvents<T>>(
    event: K,
    callback: EventCallback<Parameters<DataTableEvents<T>[K]>[0]>
  ): void {
    this.eventEmitter.off(event, callback);
  }

  emit<K extends keyof DataTableEvents<T>>(
    event: K,
    data?: any
  ): void {
    this.eventEmitter.emit(event as string, data);
  }

  // ============================================================================
  // Plugin System
  // ============================================================================

  use(plugin: Plugin<T>): void {
    if (this.plugins.has(plugin.name)) {
      console.warn(`DataTable: Plugin ${plugin.name} is already installed`);
      return;
    }

    plugin.install(this);
    this.plugins.set(plugin.name, plugin);
  }

  // ============================================================================
  // Utilities
  // ============================================================================

  private getRowId(row: T): RowId {
    if (typeof this.options.rowId === 'function') {
      return this.options.rowId(row);
    }
    return row[this.options.rowId as keyof T] as RowId;
  }

  private ensureNotDestroyed(): void {
    if (this.destroyed) {
      throw new Error('DataTable: Instance has been destroyed');
    }
  }

  // ============================================================================
  // Lifecycle
  // ============================================================================

  destroy(): void {
    if (this.destroyed) return;

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

// Factory function for easier usage
export function createDataTable<T extends RowData = RowData>(
  options: DataTableOptions<T>
): DataTable<T> {
  return new DataTable(options);
}
