/**
 * Filter Engine
 * Handles all filtering operations for DataTable
 */

import type {
  Row,
  RowData,
  Column,
  Filter,
  FilterState,
  FilterOperator,
  CellValue,
  DataTableOptions,
} from '../types';

export class FilterEngine<T extends RowData = RowData> {
  private options: Pick<Required<DataTableOptions<T>>, 'filtering'>;

  constructor(options: Pick<Required<DataTableOptions<T>>, 'filtering'>) {
    this.options = options;
  }

  /**
   * Apply all filters to rows
   */
  filter(rows: Row<T>[], filterState: FilterState, columns: Column<T>[]): Row<T>[] {
    let filteredRows = rows;

    // Apply column filters
    if (filterState.filters.length > 0) {
      filteredRows = this.applyColumnFilters(filteredRows, filterState.filters, columns);
    }

    // Apply global search
    if (filterState.globalSearch && filterState.globalSearch.trim()) {
      filteredRows = this.applyGlobalSearch(filteredRows, filterState.globalSearch, columns);
    }

    return filteredRows;
  }

  /**
   * Apply individual column filters
   */
  private applyColumnFilters(rows: Row<T>[], filters: Filter[], columns: Column<T>[]): Row<T>[] {
    return rows.filter((row) => {
      return filters.every((filter) => this.evaluateFilter(row, filter, columns));
    });
  }

  /**
   * Apply global search across all searchable columns
   */
  private applyGlobalSearch(rows: Row<T>[], query: string, columns: Column<T>[]): Row<T>[] {
    const searchableColumns = columns.filter((col) => col.filterable !== false);
    
    // If no searchable columns, return all rows (no filtering applied)
    if (searchableColumns.length === 0) {
      return rows;
    }
    
    const normalizedQuery = this.options.filtering.caseSensitive
      ? query
      : query.toLowerCase();

    return rows.filter((row) => {
      return searchableColumns.some((column) => {
        const value = this.getCellValue(row, column);
        const normalizedValue = this.normalizeValue(value);

        return normalizedValue.includes(normalizedQuery);
      });
    });
  }

  /**
   * Evaluate a single filter against a row
   */
  private evaluateFilter(row: Row<T>, filter: Filter, columns: Column<T>[]): boolean {
    // Check for custom filter function
    if (this.options.filtering.customFilters?.[filter.columnId]) {
      return this.options.filtering.customFilters[filter.columnId]!(row, filter);
    }

    const column = columns.find((col) => col.id === filter.columnId);
    if (!column) return true;

    const cellValue = this.getCellValue(row, column);

    return this.applyOperator(cellValue, filter.operator, filter.value);
  }

  /**
   * Apply filter operator
   */
  private applyOperator(
    cellValue: CellValue,
    operator: FilterOperator,
    filterValue: CellValue | CellValue[]
  ): boolean {
    const normalizedCellValue = this.normalizeValue(cellValue);
    const normalizedFilterValue = Array.isArray(filterValue)
      ? filterValue.map((v) => this.normalizeValue(v))
      : this.normalizeValue(filterValue);

    switch (operator) {
      case 'equals':
        return normalizedCellValue === normalizedFilterValue;

      case 'notEquals':
        return normalizedCellValue !== normalizedFilterValue;

      case 'contains':
        return normalizedCellValue.includes(normalizedFilterValue as string);

      case 'notContains':
        return !normalizedCellValue.includes(normalizedFilterValue as string);

      case 'startsWith':
        return normalizedCellValue.startsWith(normalizedFilterValue as string);

      case 'endsWith':
        return normalizedCellValue.endsWith(normalizedFilterValue as string);

      case 'isEmpty':
        return cellValue === null || cellValue === undefined || cellValue === '';

      case 'isNotEmpty':
        return cellValue !== null && cellValue !== undefined && cellValue !== '';

      case 'greaterThan':
        return this.compareNumeric(cellValue, filterValue as CellValue) > 0;

      case 'greaterThanOrEqual':
        return this.compareNumeric(cellValue, filterValue as CellValue) >= 0;

      case 'lessThan':
        return this.compareNumeric(cellValue, filterValue as CellValue) < 0;

      case 'lessThanOrEqual':
        return this.compareNumeric(cellValue, filterValue as CellValue) <= 0;

      case 'between':
        if (!Array.isArray(filterValue) || filterValue.length !== 2) return false;
        const num = this.toNumber(cellValue);
        const min = this.toNumber(filterValue[0]);
        const max = this.toNumber(filterValue[1]);
        return num >= min && num <= max;

      case 'in':
        if (!Array.isArray(filterValue)) return false;
        return (normalizedFilterValue as string[]).includes(normalizedCellValue);

      case 'notIn':
        if (!Array.isArray(filterValue)) return false;
        return !(normalizedFilterValue as string[]).includes(normalizedCellValue);

      default:
        console.warn(`FilterEngine: Unknown operator ${operator}`);
        return true;
    }
  }

  /**
   * Get cell value from row using column definition
   */
  private getCellValue(row: Row<T>, column: Column<T>): CellValue {
    // Use custom accessor if provided
    if (column.accessor) {
      return column.accessor(row.data);
    }

    // Use field
    return row.data[column.field as keyof T] as CellValue;
  }

  /**
   * Normalize value for comparison
   */
  private normalizeValue(value: CellValue): string {
    if (value === null || value === undefined) {
      return '';
    }

    const stringValue = String(value);
    return this.options.filtering.caseSensitive
      ? stringValue
      : stringValue.toLowerCase();
  }

  /**
   * Convert value to number for numeric comparisons
   */
  private toNumber(value: CellValue): number {
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
      const num = parseFloat(value);
      return isNaN(num) ? 0 : num;
    }
    return 0;
  }

  /**
   * Compare numeric values
   */
  private compareNumeric(a: CellValue, b: CellValue): number {
    return this.toNumber(a) - this.toNumber(b);
  }
}

/**
 * Debounce utility for search
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;

  return function (...args: Parameters<T>) {
    if (timeout) {
      clearTimeout(timeout);
    }

    timeout = setTimeout(() => {
      func(...args);
    }, wait);
  };
}
