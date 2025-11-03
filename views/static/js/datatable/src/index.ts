/**
 * Awo DataTable
 * Headless DataTable for any design system
 * 
 * @package @awo/datatable
 * @author Awo ERP
 * @version 2.0.0
 * @license MIT
 */

// Core exports
export { DataTable, createDataTable } from './core';

// Type exports
export type {
  // Core types
  RowId,
  CellValue,
  RowData,
  Row,
  Column,
  
  // Sorting
  SortDirection,
  SortState,
  MultiSortState,
  
  // Filtering
  FilterOperator,
  Filter,
  FilterState,
  FilterFunction,
  
  // Pagination
  PaginationState,
  PaginationOptions,
  
  // Selection
  SelectionState,
  SelectionOptions,
  
  // Server-side
  DataMode,
  ServerSideOptions,
  LoadState,
  
  // Events
  DataTableEvents,
  EventCallback,
  
  // State
  DataTableState,
  DataTableOptions,
  
  // Plugin
  Plugin,
  
  // Export
  ExportFormat,
  ExportOptions,
  
  // Public API
  DataTableCore,
} from './types';

// Plugin exports
export { ExportPlugin, createExportPlugin } from './plugins/export';
export {
  VirtualScrollPlugin,
  createVirtualScrollPlugin,
  measureRowHeight,
  calculateOptimalOverscan,
} from './plugins/virtual-scroll';
export type { VirtualScrollOptions, VirtualScrollState } from './plugins/virtual-scroll';

// Utility exports
export { EventEmitter } from './utils/event-emitter';
export { StateManager } from './utils/state-manager';
export { FilterEngine, debounce } from './utils/filter-engine';
export { SortEngine } from './utils/sort-engine';

/**
 * Version
 */
export const VERSION = '2.0.0';

// Re-import for default export
import { DataTable, createDataTable } from './core';

/**
 * Default export
 */
export default {
  DataTable,
  createDataTable,
  VERSION,
};
