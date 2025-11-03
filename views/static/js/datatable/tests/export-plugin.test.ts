/**
 * ExportPlugin Tests
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { ExportPlugin, createExportPlugin } from '../src/plugins/export';
import { DataTable } from '../src/core';
import type { Column, RowData, DataTableCore, ExportOptions } from '../src/types';

interface TestData extends RowData {
  id: number;
  name: string;
  age: number;
  email: string;
  active: boolean;
  score: number | null;
  joinDate: string;
}

// Mock DOM APIs
const mockCreateElement = vi.fn();
const mockClick = vi.fn();
const mockCreateObjectURL = vi.fn();
const mockRevokeObjectURL = vi.fn();

describe('ExportPlugin', () => {
  let exportPlugin: ExportPlugin<TestData>;
  let table: DataTableCore<TestData>;
  let columns: Column<TestData>[];
  let data: TestData[];
  let MockBlob: any;

  beforeEach(() => {
    // Mock DOM APIs
    Object.defineProperty(global, 'document', {
      value: {
        createElement: mockCreateElement.mockReturnValue({
          href: '',
          download: '',
          click: mockClick
        })
      },
      writable: true
    });

    Object.defineProperty(global, 'URL', {
      value: {
        createObjectURL: mockCreateObjectURL.mockReturnValue('mock-url'),
        revokeObjectURL: mockRevokeObjectURL
      },
      writable: true
    });

    MockBlob = vi.fn().mockImplementation(function(content: any[], options: any) {
      this.content = content;
      this.options = options;
    });
    
    Object.defineProperty(global, 'Blob', {
      value: MockBlob,
      writable: true
    });

    // Setup test data
    columns = [
      { id: 'id', field: 'id', label: 'ID' },
      { id: 'name', field: 'name', label: 'Name' },
      { id: 'age', field: 'age', label: 'Age' },
      { id: 'email', field: 'email', label: 'Email' },
      { id: 'active', field: 'active', label: 'Active' },
      { id: 'score', field: 'score', label: 'Score' },
      { id: 'joinDate', field: 'joinDate', label: 'Join Date' },
    ];

    data = [
      { id: 1, name: 'John Doe', age: 30, email: 'john@example.com', active: true, score: 85, joinDate: '2023-01-15' },
      { id: 2, name: 'Jane Smith', age: 25, email: 'jane@example.com', active: false, score: 92, joinDate: '2023-02-20' },
      { id: 3, name: 'Bob Johnson', age: 35, email: 'bob@company.org', active: true, score: null, joinDate: '2022-12-10' },
      { id: 4, name: 'Alice Brown', age: 28, email: 'alice@example.com', active: false, score: 78, joinDate: '2023-03-05' },
      { id: 5, name: 'Charlie Wilson', age: 42, email: 'charlie@example.com', active: true, score: 90, joinDate: '2022-11-01' },
    ];

    table = new DataTable({ columns, data });
    exportPlugin = new ExportPlugin();
    exportPlugin.install(table);

    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.clearAllTimers();
  });

  describe('Plugin Installation', () => {
    it('should create plugin using factory function', () => {
      const factoryPlugin = createExportPlugin<TestData>();
      expect(factoryPlugin).toBeInstanceOf(ExportPlugin);
      expect(factoryPlugin.name).toBe('export');
      expect(factoryPlugin.version).toBe('1.0.0');
    });

    it('should install plugin correctly', () => {
      const newPlugin = new ExportPlugin<TestData>();
      expect(() => newPlugin.install(table)).not.toThrow();
    });

    it('should warn when exporting without table instance', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const uninstalledPlugin = new ExportPlugin<TestData>();
      
      uninstalledPlugin.export({ format: 'csv' });
      
      expect(consoleSpy).toHaveBeenCalledWith('ExportPlugin: Table instance not available');
      consoleSpy.mockRestore();
    });
  });

  describe('CSV Export', () => {
    it('should export to CSV with headers', () => {
      const options: ExportOptions = {
        format: 'csv',
        includeHeaders: true
      };

      exportPlugin.export(options);

      expect(mockCreateElement).toHaveBeenCalledWith('a');
      expect(mockCreateObjectURL).toHaveBeenCalled();
      expect(mockClick).toHaveBeenCalled();

      // Check blob content
      const blobCall = mockCreateObjectURL.mock.calls[0];
      expect(blobCall).toBeDefined();
    });

    it('should export to CSV without headers', () => {
      const options: ExportOptions = {
        format: 'csv',
        includeHeaders: false
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should handle CSV escaping correctly', () => {
      // Add data with special characters that need escaping
      const specialData = [
        { id: 1, name: 'John "Doe"', age: 30, email: 'john@example.com', active: true, score: 85, joinDate: '2023-01-15' },
        { id: 2, name: 'Jane,Smith', age: 25, email: 'jane@example.com', active: false, score: 92, joinDate: '2023-02-20' },
        { id: 3, name: 'Bob\nJohnson', age: 35, email: 'bob@company.org', active: true, score: null, joinDate: '2022-12-10' },
      ];

      table.setData(specialData);

      const options: ExportOptions = {
        format: 'csv',
        includeHeaders: true
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should export selected rows only when selectedOnly is true', () => {
      // Enable selection and select some rows
      const selectableTable = new DataTable({
        columns,
        data,
        selection: { enabled: true, multiple: true }
      });
      
      const selectablePlugin = new ExportPlugin<TestData>();
      selectablePlugin.install(selectableTable);

      selectableTable.selectRow(1);
      selectableTable.selectRow(3);

      const options: ExportOptions = {
        format: 'csv',
        selectedOnly: true
      };

      selectablePlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should export visible rows only when visibleOnly is true', () => {
      // Apply filter to make only some rows visible
      table.addFilter({ columnId: 'active', operator: 'equals', value: true });

      const options: ExportOptions = {
        format: 'csv',
        visibleOnly: true
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should export specific columns only', () => {
      const options: ExportOptions = {
        format: 'csv',
        columns: ['name', 'email']
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should use custom filename', () => {
      const options: ExportOptions = {
        format: 'csv',
        filename: 'custom-export'
      };

      exportPlugin.export(options);

      const linkElement = mockCreateElement.mock.results[0].value;
      expect(linkElement.download).toBe('custom-export.csv');
    });
  });

  describe('JSON Export', () => {
    it('should export to JSON format', () => {
      const options: ExportOptions = {
        format: 'json',
        filename: 'test-export'
      };

      exportPlugin.export(options);

      expect(mockCreateElement).toHaveBeenCalledWith('a');
      expect(mockCreateObjectURL).toHaveBeenCalled();
      expect(mockClick).toHaveBeenCalled();

      const linkElement = mockCreateElement.mock.results[0].value;
      expect(linkElement.download).toBe('test-export.json');
    });

    it('should export selected rows to JSON', () => {
      const selectableTable = new DataTable({
        columns,
        data,
        selection: { enabled: true, multiple: true }
      });
      
      const selectablePlugin = new ExportPlugin<TestData>();
      selectablePlugin.install(selectableTable);

      selectableTable.selectRow(1);
      selectableTable.selectRow(2);

      const options: ExportOptions = {
        format: 'json',
        selectedOnly: true
      };

      selectablePlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should export specific columns to JSON', () => {
      const options: ExportOptions = {
        format: 'json',
        columns: ['id', 'name', 'age']
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });
  });

  describe('TXT Export', () => {
    it('should export to TXT format with headers', () => {
      const options: ExportOptions = {
        format: 'txt',
        includeHeaders: true
      };

      exportPlugin.export(options);

      expect(mockCreateElement).toHaveBeenCalledWith('a');
      expect(mockCreateObjectURL).toHaveBeenCalled();
      expect(mockClick).toHaveBeenCalled();

      const linkElement = mockCreateElement.mock.results[0].value;
      expect(linkElement.download).toContain('.txt');
    });

    it('should export to TXT format without headers', () => {
      const options: ExportOptions = {
        format: 'txt',
        includeHeaders: false
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should handle empty data for TXT export', () => {
      table.clearData();

      const options: ExportOptions = {
        format: 'txt',
        includeHeaders: true
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });
  });

  describe('XLSX Export', () => {
    it('should warn about XLSX export requiring additional library', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      const options: ExportOptions = {
        format: 'xlsx'
      };

      exportPlugin.export(options);

      expect(consoleSpy).toHaveBeenCalledWith('ExportPlugin: XLSX export requires additional library');
      expect(mockClick).not.toHaveBeenCalled();

      consoleSpy.mockRestore();
    });
  });

  describe('Column Handling', () => {
    it('should handle columns with custom accessors', () => {
      const customColumns: Column<TestData>[] = [
        ...columns,
        {
          id: 'fullName',
          field: 'name',
          label: 'Full Name',
          accessor: (row: TestData) => `${row.name} (ID: ${row.id})`
        }
      ];

      table.setColumns(customColumns);

      const options: ExportOptions = {
        format: 'csv',
        columns: ['fullName', 'email']
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should handle columns with custom formatters', () => {
      const formattedColumns: Column<TestData>[] = [
        ...columns.slice(0, 3),
        {
          id: 'formattedScore',
          field: 'score',
          label: 'Formatted Score',
          formatter: (value: any) => value ? `${value}%` : 'N/A'
        }
      ];

      table.setColumns(formattedColumns);

      const options: ExportOptions = {
        format: 'json',
        columns: ['name', 'formattedScore']
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should handle null and undefined values correctly', () => {
      const dataWithNulls = [
        { id: 1, name: 'John', age: 30, email: null as any, active: true, score: undefined as any, joinDate: '2023-01-15' },
        { id: 2, name: null as any, age: 25, email: 'jane@example.com', active: false, score: 92, joinDate: null as any },
      ];

      table.setData(dataWithNulls);

      const options: ExportOptions = {
        format: 'csv'
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });
  });

  describe('Error Handling', () => {
    it('should handle unsupported export formats', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      const options: ExportOptions = {
        format: 'pdf' as any // Unsupported format
      };

      exportPlugin.export(options);

      expect(consoleSpy).toHaveBeenCalledWith('ExportPlugin: Unsupported format pdf');
      expect(mockClick).not.toHaveBeenCalled();

      consoleSpy.mockRestore();
    });

    it('should handle missing column gracefully', () => {
      const options: ExportOptions = {
        format: 'csv',
        columns: ['name', 'nonexistent', 'email']
      };

      exportPlugin.export(options);

      // Should still export, just skip the non-existent column
      expect(mockClick).toHaveBeenCalled();
    });

    it('should handle DOM API errors gracefully', () => {
      // Mock createElement to throw an error
      mockCreateElement.mockImplementationOnce(() => {
        throw new Error('DOM error');
      });

      const options: ExportOptions = {
        format: 'csv'
      };

      expect(() => exportPlugin.export(options)).toThrow('DOM error');
    });
  });

  describe('File Cleanup', () => {
    it('should revoke object URL after download', () => {
      vi.useFakeTimers();

      const options: ExportOptions = {
        format: 'csv'
      };

      exportPlugin.export(options);

      expect(mockRevokeObjectURL).not.toHaveBeenCalled();

      // Fast-forward past the cleanup timeout
      vi.advanceTimersByTime(150);

      expect(mockRevokeObjectURL).toHaveBeenCalledWith('mock-url');

      vi.useRealTimers();
    });
  });

  describe('Data Processing', () => {
    it('should handle large datasets efficiently', () => {
      // Create large dataset
      const largeData = Array.from({ length: 10000 }, (_, i) => ({
        id: i,
        name: `User ${i}`,
        age: 20 + (i % 50),
        email: `user${i}@example.com`,
        active: i % 2 === 0,
        score: i % 100,
        joinDate: `2023-${String((i % 12) + 1).padStart(2, '0')}-01`
      }));

      table.setData(largeData);

      const start = performance.now();
      
      const options: ExportOptions = {
        format: 'csv'
      };

      exportPlugin.export(options);

      const end = performance.now();

      expect(end - start).toBeLessThan(1000); // Should complete in less than 1 second
      expect(mockClick).toHaveBeenCalled();
    });

    it('should handle empty datasets', () => {
      table.clearData();

      const options: ExportOptions = {
        format: 'csv',
        includeHeaders: true
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should handle datasets with only headers', () => {
      table.clearData();

      const options: ExportOptions = {
        format: 'txt',
        includeHeaders: true
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });
  });

  describe('MIME Types and Extensions', () => {
    it('should use correct MIME type for CSV', () => {
      const options: ExportOptions = {
        format: 'csv'
      };

      exportPlugin.export(options);

      expect(MockBlob).toHaveBeenCalledWith(
        expect.any(Array),
        { type: 'text/csv' }
      );
    });

    it('should use correct MIME type for JSON', () => {
      const options: ExportOptions = {
        format: 'json'
      };

      exportPlugin.export(options);

      expect(MockBlob).toHaveBeenCalledWith(
        expect.any(Array),
        { type: 'application/json' }
      );
    });

    it('should use correct MIME type for TXT', () => {
      const options: ExportOptions = {
        format: 'txt'
      };

      exportPlugin.export(options);

      expect(MockBlob).toHaveBeenCalledWith(
        expect.any(Array),
        { type: 'text/plain' }
      );
    });

    it('should generate default filename with timestamp', () => {
      const options: ExportOptions = {
        format: 'csv'
      };

      exportPlugin.export(options);

      const linkElement = mockCreateElement.mock.results[0].value;
      expect(linkElement.download).toMatch(/^export-\d+\.csv$/);
    });
  });

  describe('Integration with DataTable Features', () => {
    it('should export filtered data when visibleOnly is true', () => {
      // Apply multiple filters
      table.addFilter({ columnId: 'active', operator: 'equals', value: true });
      table.addFilter({ columnId: 'age', operator: 'greaterThan', value: 30 });

      const options: ExportOptions = {
        format: 'json',
        visibleOnly: true
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should export paginated data correctly', () => {
      // Create table with pagination
      const paginatedTable = new DataTable({
        columns,
        data,
        pagination: { enabled: true, pageSize: 2 }
      });

      const paginatedPlugin = new ExportPlugin<TestData>();
      paginatedPlugin.install(paginatedTable);

      paginatedTable.nextPage();

      const options: ExportOptions = {
        format: 'csv',
        visibleOnly: true
      };

      paginatedPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });

    it('should export sorted data in correct order', () => {
      table.sort('age', 'desc');

      const options: ExportOptions = {
        format: 'csv',
        visibleOnly: true
      };

      exportPlugin.export(options);

      expect(mockClick).toHaveBeenCalled();
    });
  });
});