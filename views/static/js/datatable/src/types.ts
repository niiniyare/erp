/**
 * Awo DataTable - Type Definitions
 * Headless DataTable for any design system
 * 
 * @package @awo/datatable
 * @author Awo ERP
 * @version 2.0.0
 */

// ============================================================================
// Core Types
// ============================================================================

export type RowId = string | number;
export type CellValue = string | number | boolean | null | undefined;
export type RowData = Record<string, CellValue>;

export interface Row<T = RowData> {
  id: RowId;
  data: T;
  selected?: boolean;
  expanded?: boolean;
  disabled?: boolean;
  metadata?: Record<string, any>;
}

export interface Column<T = RowData> {
  id: string;
  field: keyof T | string;
  label?: string;
  sortable?: boolean;
  filterable?: boolean;
  visible?: boolean;
  width?: number | string;
  minWidth?: number;
  maxWidth?: number;
  accessor?: (row: T) => CellValue;
  formatter?: (value: CellValue, row: T) => string | CellValue;
  comparator?: (a: CellValue, b: CellValue) => number;
  metadata?: Record<string, any>;
}

// ============================================================================
// Sorting
// ============================================================================

export type SortDirection = 'asc' | 'desc' | null;

export interface SortState {
  columnId: string;
  direction: SortDirection;
}

export interface MultiSortState {
  sorts: SortState[];
}

// ============================================================================
// Filtering
// ============================================================================

export type FilterOperator = 
  | 'equals'
  | 'notEquals'
  | 'contains'
  | 'notContains'
  | 'startsWith'
  | 'endsWith'
  | 'isEmpty'
  | 'isNotEmpty'
  | 'greaterThan'
  | 'greaterThanOrEqual'
  | 'lessThan'
  | 'lessThanOrEqual'
  | 'between'
  | 'in'
  | 'notIn';

export interface Filter {
  columnId: string;
  operator: FilterOperator;
  value: CellValue | CellValue[];
}

export interface FilterState {
  filters: Filter[];
  globalSearch?: string;
}

export type FilterFunction<T = RowData> = (row: Row<T>, filter: Filter) => boolean;

// ============================================================================
// Pagination
// ============================================================================

export interface PaginationState {
  pageIndex: number;
  pageSize: number;
  totalRows: number;
  totalPages: number;
}

export interface PaginationOptions {
  enabled: boolean;
  pageSize: number;
  pageSizeOptions?: number[];
  showFirstLast?: boolean;
}

// ============================================================================
// Selection
// ============================================================================

export interface SelectionState {
  selectedRowIds: Set<RowId>;
  isAllSelected: boolean;
  isPartiallySelected: boolean;
}

export interface SelectionOptions {
  enabled: boolean;
  multiple?: boolean;
  mode?: 'click' | 'checkbox';
  selectOnRowClick?: boolean;
}

// ============================================================================
// Data Loading
// ============================================================================

export type DataMode = 'client' | 'server';

export interface ServerSideOptions {
  url?: string;
  method?: 'GET' | 'POST';
  headers?: Record<string, string>;
  params?: Record<string, any>;
  dataPath?: string;
  totalPath?: string;
  transformer?: (response: any) => { data: RowData[]; total: number };
}

export interface LoadState {
  loading: boolean;
  error: Error | null;
  lastFetch?: Date;
}

// ============================================================================
// Events
// ============================================================================

export interface DataTableEvents<T = RowData> {
  'init': () => void;
  'data:change': (rows: Row<T>[]) => void;
  'data:load': (rows: Row<T>[]) => void;
  'data:error': (error: Error) => void;
  'sort:change': (sort: SortState | MultiSortState) => void;
  'filter:change': (filters: FilterState) => void;
  'page:change': (pagination: PaginationState) => void;
  'selection:change': (selection: SelectionState) => void;
  'row:select': (row: Row<T>, selected: boolean) => void;
  'row:click': (row: Row<T>, event?: Event) => void;
  'row:expand': (row: Row<T>, expanded: boolean) => void;
  'column:visibility': (columnId: string, visible: boolean) => void;
  'search:change': (query: string) => void;
  'state:change': (state: DataTableState<T>) => void;
  'destroy': () => void;
}

export type EventCallback<T = any> = (data: T) => void;

// ============================================================================
// State Management
// ============================================================================

export interface DataTableState<T = RowData> {
  // Data
  rows: Row<T>[];
  columns: Column<T>[];
  
  // View state
  sortState: SortState | MultiSortState | undefined;
  filterState: FilterState;
  paginationState: PaginationState;
  selectionState: SelectionState;
  loadState: LoadState;
  
  // Computed
  visibleRows: Row<T>[];
  filteredRows: Row<T>[];
  sortedRows: Row<T>[];
  paginatedRows: Row<T>[];
  visibleColumns: Column<T>[];
}

// ============================================================================
// Configuration Options
// ============================================================================

export interface DataTableOptions<T = RowData> {
  // Data
  data?: T[];
  columns: Column<T>[];
  dataMode?: DataMode;
  serverSide?: ServerSideOptions;
  
  // Features
  sorting?: {
    enabled: boolean;
    multiSort?: boolean;
    defaultSort?: SortState | MultiSortState;
  };
  
  filtering?: {
    enabled: boolean;
    globalSearch?: boolean;
    debounceMs?: number;
    caseSensitive?: boolean;
    customFilters?: Record<string, FilterFunction<T>>;
  };
  
  pagination?: PaginationOptions;
  selection?: SelectionOptions;
  
  // Performance
  virtualScroll?: {
    enabled: boolean;
    rowHeight?: number;
    overscan?: number;
  };
  
  // Behavior
  rowId?: keyof T | ((row: T) => RowId);
  enableRowExpansion?: boolean;
  preserveState?: boolean;
  stateKey?: string;
  
  // Customization
  customComparators?: Record<string, (a: CellValue, b: CellValue) => number>;
  customFormatters?: Record<string, (value: CellValue, row: T) => string | CellValue>;
  
  // Callbacks (Legacy - use events instead)
  onInit?: () => void;
  onDataChange?: (rows: Row<T>[]) => void;
  onError?: (error: Error) => void;
}

// ============================================================================
// Plugin System
// ============================================================================

export interface Plugin<T = RowData> {
  name: string;
  version?: string;
  install: (table: DataTableCore<T>) => void;
  uninstall?: (table: DataTableCore<T>) => void;
}

// ============================================================================
// Public API Interface
// ============================================================================

export interface DataTableCore<T = RowData> {
  // Data Management
  setData(data: T[]): void;
  getData(): Row<T>[];
  addRow(data: T): void;
  updateRow(id: RowId, data: Partial<T>): void;
  deleteRow(id: RowId): void;
  clearData(): void;
  
  // Columns
  getColumns(): Column<T>[];
  setColumns(columns: Column<T>[]): void;
  showColumn(columnId: string): void;
  hideColumn(columnId: string): void;
  toggleColumn(columnId: string): void;
  
  // Sorting
  sort(columnId: string, direction?: SortDirection): void;
  multiSort(sorts: SortState[]): void;
  clearSort(): void;
  getSortState(): SortState | MultiSortState | undefined;
  
  // Filtering
  filter(filters: Filter[]): void;
  addFilter(filter: Filter): void;
  removeFilter(columnId: string): void;
  clearFilters(): void;
  search(query: string): void;
  getFilterState(): FilterState;
  
  // Pagination
  goToPage(page: number): void;
  nextPage(): void;
  previousPage(): void;
  firstPage(): void;
  lastPage(): void;
  setPageSize(size: number): void;
  getPaginationState(): PaginationState;
  
  // Selection
  selectRow(id: RowId): void;
  deselectRow(id: RowId): void;
  toggleRowSelection(id: RowId): void;
  selectAll(): void;
  deselectAll(): void;
  getSelectedRows(): Row<T>[];
  getSelectionState(): SelectionState;
  
  // State
  getState(): DataTableState<T>;
  setState(state: Partial<DataTableState<T>>): void;
  resetState(): void;
  
  // Server-side
  reload(): Promise<void>;
  
  // Events
  on<K extends keyof DataTableEvents<T>>(
    event: K,
    callback: EventCallback<Parameters<DataTableEvents<T>[K]>[0]>
  ): () => void;
  
  off<K extends keyof DataTableEvents<T>>(
    event: K,
    callback: EventCallback<Parameters<DataTableEvents<T>[K]>[0]>
  ): void;
  
  emit<K extends keyof DataTableEvents<T>>(
    event: K,
    data?: any
  ): void;
  
  // Lifecycle
  destroy(): void;
  
  // Plugins
  use(plugin: Plugin<T>): void;
}

// ============================================================================
// Export Utilities
// ============================================================================

export type ExportFormat = 'csv' | 'json' | 'xlsx' | 'txt';

export interface ExportOptions {
  format: ExportFormat;
  filename?: string;
  includeHeaders?: boolean;
  selectedOnly?: boolean;
  visibleOnly?: boolean;
  columns?: string[];
}
