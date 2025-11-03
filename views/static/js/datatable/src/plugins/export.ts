/**
 * Export Plugin
 * Provides data export functionality in multiple formats
 */

import type { Plugin, DataTableCore, RowData, ExportOptions } from '../types';

export class ExportPlugin<T extends RowData = RowData> implements Plugin<T> {
  name = 'export';
  version = '1.0.0';

  private table: DataTableCore<T> | null = null;

  install(table: DataTableCore<T>): void {
    this.table = table;
  }

  /**
   * Export table data
   */
  export(options: ExportOptions): void {
    if (!this.table) {
      console.error('ExportPlugin: Table instance not available');
      return;
    }

    const state = this.table.getState();
    let rows = state.rows;

    // Filter rows based on options
    if (options.selectedOnly) {
      rows = this.table.getSelectedRows();
    } else if (options.visibleOnly) {
      rows = state.visibleRows;
    }

    // Filter columns
    const columns = options.columns
      ? state.columns.filter((col) => options.columns!.includes(col.id))
      : state.visibleColumns;

    // Generate export based on format
    let content: string;
    let mimeType: string;
    let extension: string;

    switch (options.format) {
      case 'csv':
        content = this.toCSV(rows, columns, options.includeHeaders ?? true);
        mimeType = 'text/csv';
        extension = 'csv';
        break;

      case 'json':
        content = this.toJSON(rows, columns);
        mimeType = 'application/json';
        extension = 'json';
        break;

      case 'txt':
        content = this.toTXT(rows, columns, options.includeHeaders ?? true);
        mimeType = 'text/plain';
        extension = 'txt';
        break;

      case 'xlsx':
        console.warn('ExportPlugin: XLSX export requires additional library');
        return;

      default:
        console.error(`ExportPlugin: Unsupported format ${options.format}`);
        return;
    }

    // Download file
    this.download(
      content,
      options.filename || `export-${Date.now()}`,
      mimeType,
      extension
    );
  }

  /**
   * Convert to CSV format
   */
  private toCSV(rows: any[], columns: any[], includeHeaders: boolean): string {
    const lines: string[] = [];

    // Add headers
    if (includeHeaders) {
      const headers = columns.map((col) => this.escapeCSV(col.label || col.id));
      lines.push(headers.join(','));
    }

    // Add data rows
    rows.forEach((row) => {
      const values = columns.map((col) => {
        const value = this.getCellValue(row, col);
        return this.escapeCSV(value);
      });
      lines.push(values.join(','));
    });

    return lines.join('\n');
  }

  /**
   * Convert to JSON format
   */
  private toJSON(rows: any[], columns: any[]): string {
    const data = rows.map((row) => {
      const obj: Record<string, any> = {};
      columns.forEach((col) => {
        obj[col.label || col.id] = this.getCellValue(row, col);
      });
      return obj;
    });

    return JSON.stringify(data, null, 2);
  }

  /**
   * Convert to plain text format
   */
  private toTXT(rows: any[], columns: any[], includeHeaders: boolean): string {
    const lines: string[] = [];

    // Calculate column widths
    const widths = columns.map((col) => {
      const headerWidth = (col.label || col.id).length;
      const maxDataWidth = rows.length > 0 
        ? Math.max(...rows.map((row) => String(this.getCellValue(row, col)).length), 0)
        : 0;
      return Math.max(headerWidth, maxDataWidth);
    });

    // Add headers
    if (includeHeaders) {
      const headers = columns
        .map((col, i) => (col.label || col.id).padEnd(widths[i] || 0))
        .join(' | ');
      lines.push(headers);
      lines.push(widths.map((w) => '-'.repeat(w || 0)).join('-+-'));
    }

    // Add data rows
    rows.forEach((row) => {
      const values = columns
        .map((col, i) => String(this.getCellValue(row, col)).padEnd(widths[i] || 0))
        .join(' | ');
      lines.push(values);
    });

    return lines.join('\n');
  }

  /**
   * Get cell value
   */
  private getCellValue(row: any, column: any): any {
    if (column.accessor) {
      return column.accessor(row.data);
    }
    if (column.formatter) {
      return column.formatter(row.data[column.field], row.data);
    }
    return row.data[column.field];
  }

  /**
   * Escape CSV value
   */
  private escapeCSV(value: any): string {
    if (value === null || value === undefined) {
      return '';
    }

    const str = String(value);

    // Escape if contains comma, quote, or newline
    if (str.includes(',') || str.includes('"') || str.includes('\n')) {
      return `"${str.replace(/"/g, '""')}"`;
    }

    return str;
  }

  /**
   * Download content as file
   */
  private download(
    content: string,
    filename: string,
    mimeType: string,
    extension: string
  ): void {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');

    link.href = url;
    link.download = `${filename}.${extension}`;
    link.click();

    // Cleanup
    setTimeout(() => URL.revokeObjectURL(url), 100);
  }
}

// Factory function
export function createExportPlugin<T extends RowData = RowData>(): ExportPlugin<T> {
  return new ExportPlugin<T>();
}
