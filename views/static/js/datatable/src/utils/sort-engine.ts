/**
 * Sort Engine
 * Handles all sorting operations for DataTable
 */

import type {
  Row,
  RowData,
  Column,
  SortState,
  MultiSortState,
  CellValue,
  DataTableOptions,
} from '../types';

export class SortEngine<T extends RowData = RowData> {
  private options: Pick<Required<DataTableOptions<T>>, 'sorting' | 'customComparators'>;

  constructor(options: Pick<Required<DataTableOptions<T>>, 'sorting' | 'customComparators'>) {
    this.options = options;
  }

  /**
   * Sort rows based on sort state
   */
  sort(
    rows: Row<T>[],
    sortState: SortState | MultiSortState,
    columns: Column<T>[]
  ): Row<T>[] {
    // Don't mutate original array
    const sortedRows = [...rows];

    if ('sorts' in sortState) {
      // Multi-sort
      return this.multiSort(sortedRows, sortState.sorts, columns);
    } else {
      // Single sort
      return this.singleSort(sortedRows, sortState, columns);
    }
  }

  /**
   * Single column sort
   */
  private singleSort(rows: Row<T>[], sortState: SortState, columns: Column<T>[]): Row<T>[] {
    const column = columns.find((col) => col.id === sortState.columnId);
    if (!column || !sortState.direction) return rows;

    const comparator = this.getComparator(column);
    const multiplier = sortState.direction === 'asc' ? 1 : -1;

    return rows.sort((a, b) => {
      const aValue = this.getCellValue(a, column);
      const bValue = this.getCellValue(b, column);
      return comparator(aValue, bValue) * multiplier;
    });
  }

  /**
   * Multi-column sort
   */
  private multiSort(rows: Row<T>[], sorts: SortState[], columns: Column<T>[]): Row<T>[] {
    return rows.sort((a, b) => {
      for (const sort of sorts) {
        if (!sort.direction) continue;

        const column = columns.find((col) => col.id === sort.columnId);
        if (!column) continue;

        const comparator = this.getComparator(column);
        const multiplier = sort.direction === 'asc' ? 1 : -1;

        const aValue = this.getCellValue(a, column);
        const bValue = this.getCellValue(b, column);
        const result = comparator(aValue, bValue) * multiplier;

        // If not equal, return result; otherwise continue to next sort
        if (result !== 0) return result;
      }
      return 0;
    });
  }

  /**
   * Get comparator function for column
   */
  private getComparator(column: Column<T>): (a: CellValue, b: CellValue) => number {
    // Use custom comparator if provided
    if (column.comparator) {
      return column.comparator;
    }

    // Use global custom comparator if available
    if (this.options.customComparators[column.id]) {
      return this.options.customComparators[column.id]!;
    }

    // Default comparator
    return this.defaultComparator;
  }

  /**
   * Default comparison logic
   */
  private defaultComparator = (a: CellValue, b: CellValue): number => {
    // Handle null/undefined
    if (a === null || a === undefined) return b === null || b === undefined ? 0 : -1;
    if (b === null || b === undefined) return 1;

    // Handle numbers
    if (typeof a === 'number' && typeof b === 'number') {
      return a - b;
    }

    // Handle booleans
    if (typeof a === 'boolean' && typeof b === 'boolean') {
      return a === b ? 0 : a ? 1 : -1;
    }

    // Handle dates
    if (this.isDate(a) && this.isDate(b)) {
      return (a as Date).getTime() - (b as Date).getTime();
    }

    // Try to parse as numbers
    const aNum = this.tryParseNumber(a);
    const bNum = this.tryParseNumber(b);
    if (aNum !== null && bNum !== null) {
      return aNum - bNum;
    }

    // Try to parse as dates
    const aDate = this.tryParseDate(a);
    const bDate = this.tryParseDate(b);
    if (this.isDate(aDate) && this.isDate(bDate)) {
      return (aDate as Date).getTime() - (bDate as Date).getTime();
    }

    // Default: string comparison
    return String(a).localeCompare(String(b), undefined, {
      numeric: true,
      sensitivity: 'base',
    });
  };

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
   * Try to parse value as number
   */
  private tryParseNumber(value: CellValue): number | null {
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
      // Don't parse date-like strings as numbers
      if (/^\d{4}-\d{2}-\d{2}/.test(value) || /^\d{1,2}\/\d{1,2}\/\d{4}/.test(value) || /^\d{1,2}-\d{1,2}-\d{4}/.test(value)) {
        return null;
      }
      
      // Remove common number formatting
      const cleaned = value.replace(/[,$]/g, '');
      const num = parseFloat(cleaned);
      
      // Only return if the entire string was consumed (i.e., it's actually a number)
      if (isNaN(num) || cleaned !== String(num)) {
        return null;
      }
      
      return num;
    }
    return null;
  }

  /**
   * Try to parse value as date
   */
  private tryParseDate(value: CellValue): Date | null {
    if (this.isDate(value)) return value;
    if (typeof value === 'string' || typeof value === 'number') {
      // Handle common date formats more specifically
      if (typeof value === 'string') {
        // Check for ISO date format (YYYY-MM-DD, YYYY-MM-DDTHH:mm:ss, etc.)
        if (/^\d{4}-\d{2}-\d{2}/.test(value)) {
          const date = new Date(value);
          return isNaN(date.getTime()) ? null : date;
        }
        
        // Check for other common formats
        if (/^\d{1,2}\/\d{1,2}\/\d{4}/.test(value) || /^\d{1,2}-\d{1,2}-\d{4}/.test(value)) {
          const date = new Date(value);
          return isNaN(date.getTime()) ? null : date;
        }
      }
      
      const date = new Date(value);
      return isNaN(date.getTime()) ? null : date;
    }
    return null;
  }

  /**
   * Check if value is a Date
   */
  private isDate(value: unknown): value is Date {
    return value instanceof Date;
  }
}
