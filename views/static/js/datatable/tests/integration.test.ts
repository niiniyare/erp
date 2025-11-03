/**
 * Integration Tests
 * Tests how all DataTable components work together
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { DataTable, createDataTable } from '../src/core';
import { ExportPlugin } from '../src/plugins/export';
import { VirtualScrollPlugin } from '../src/plugins/virtual-scroll';
import type { Column, RowData, DataTableOptions } from '../src/types';

interface UserData extends RowData {
  id: number;
  firstName: string;
  lastName: string;
  email: string;
  age: number;
  department: string;
  salary: number;
  isActive: boolean;
  joinDate: string;
  tags: string[];
}

describe('DataTable Integration Tests', () => {
  let table: DataTable<UserData>;
  let columns: Column<UserData>[];
  let data: UserData[];

  beforeEach(() => {
    columns = [
      { id: 'id', field: 'id', label: 'ID' },
      { 
        id: 'fullName', 
        field: 'firstName', 
        label: 'Full Name',
        accessor: (row: UserData) => `${row.firstName} ${row.lastName}`
      },
      { id: 'email', field: 'email', label: 'Email' },
      { id: 'age', field: 'age', label: 'Age' },
      { id: 'department', field: 'department', label: 'Department' },
      { 
        id: 'salary', 
        field: 'salary', 
        label: 'Salary',
        formatter: (value: any) => `$${value?.toLocaleString()}`
      },
      { id: 'isActive', field: 'isActive', label: 'Active' },
      { id: 'joinDate', field: 'joinDate', label: 'Join Date' },
    ];

    data = [
      {
        id: 1,
        firstName: 'John',
        lastName: 'Doe',
        email: 'john.doe@company.com',
        age: 30,
        department: 'Engineering',
        salary: 75000,
        isActive: true,
        joinDate: '2023-01-15',
        tags: ['senior', 'frontend']
      },
      {
        id: 2,
        firstName: 'Jane',
        lastName: 'Smith',
        email: 'jane.smith@company.com',
        age: 28,
        department: 'Design',
        salary: 65000,
        isActive: true,
        joinDate: '2023-02-20',
        tags: ['ui', 'ux']
      },
      {
        id: 3,
        firstName: 'Bob',
        lastName: 'Johnson',
        email: 'bob.johnson@company.com',
        age: 35,
        department: 'Engineering',
        salary: 85000,
        isActive: false,
        joinDate: '2022-12-10',
        tags: ['backend', 'senior']
      },
      {
        id: 4,
        firstName: 'Alice',
        lastName: 'Brown',
        email: 'alice.brown@company.com',
        age: 26,
        department: 'Marketing',
        salary: 55000,
        isActive: true,
        joinDate: '2023-03-05',
        tags: ['social', 'content']
      },
      {
        id: 5,
        firstName: 'Charlie',
        lastName: 'Wilson',
        email: 'charlie.wilson@company.com',
        age: 42,
        department: 'Management',
        salary: 95000,
        isActive: true,
        joinDate: '2022-11-01',
        tags: ['leadership', 'strategy']
      }
    ];
  });

  describe('Complete Feature Integration', () => {
    it('should integrate all features together', () => {
      const options: DataTableOptions<UserData> = {
        columns,
        data,
        sorting: { enabled: true, multiSort: true },
        filtering: { enabled: true, globalSearch: true },
        pagination: { enabled: true, pageSize: 3 },
        selection: { enabled: true, multiple: true },
        preserveState: true,
        stateKey: 'integration-test'
      };

      table = new DataTable(options);

      // Test sorting
      table.sort('age', 'asc');
      let state = table.getState();
      expect(state.visibleRows[0].data.age).toBe(26); // Alice

      // Test filtering
      table.addFilter({ columnId: 'department', operator: 'equals', value: 'Engineering' });
      state = table.getState();
      expect(state.visibleRows.every(row => row.data.department === 'Engineering')).toBe(true);

      // Test global search
      table.search('john');
      state = table.getState();
      expect(state.visibleRows.length).toBeGreaterThan(0);
      expect(state.visibleRows.some(row => 
        row.data.firstName.toLowerCase().includes('john') || 
        row.data.lastName.toLowerCase().includes('john')
      )).toBe(true);

      // Test selection
      table.selectRow(1);
      table.selectRow(3);
      const selectedRows = table.getSelectedRows();
      expect(selectedRows).toHaveLength(2);

      // Test pagination with filtered data
      table.clearFilters();
      table.search('');
      table.setPageSize(2);
      state = table.getState();
      expect(state.visibleRows).toHaveLength(2);
      expect(state.paginationState.totalPages).toBe(3);
    });

    it('should handle complex multi-sort with filtering', () => {
      table = new DataTable({
        columns,
        data,
        sorting: { enabled: true, multiSort: true },
        filtering: { enabled: true }
      });

      // Apply filter first
      table.addFilter({ columnId: 'isActive', operator: 'equals', value: true });

      // Then multi-sort
      table.multiSort([
        { columnId: 'department', direction: 'asc' },
        { columnId: 'age', direction: 'desc' }
      ]);

      const state = table.getState();
      
      // Should only show active users, sorted by department then age desc
      expect(state.visibleRows.every(row => row.data.isActive)).toBe(true);
      
      // Check sorting within departments
      const engineeringUsers = state.visibleRows.filter(row => row.data.department === 'Engineering');
      if (engineeringUsers.length > 1) {
        expect(engineeringUsers[0].data.age).toBeGreaterThan(engineeringUsers[1].data.age);
      }
    });

    it('should maintain selection across data operations', () => {
      table = new DataTable({
        columns,
        data,
        selection: { enabled: true, multiple: true }
      });

      // Select some rows
      table.selectRow(1);
      table.selectRow(3);
      expect(table.getSelectedRows()).toHaveLength(2);

      // Sort data
      table.sort('age', 'desc');
      expect(table.getSelectedRows()).toHaveLength(2); // Selection should persist

      // Filter data
      table.addFilter({ columnId: 'isActive', operator: 'equals', value: true });
      const selectedAfterFilter = table.getSelectedRows();
      
      // Should only show selected rows that match filter
      expect(selectedAfterFilter.every(row => row.data.isActive)).toBe(true);
    });

    it('should handle state changes with event propagation', () => {
      const events: string[] = [];
      
      table = new DataTable({ columns, data });

      table.on('sort:change', () => events.push('sort'));
      table.on('filter:change', () => events.push('filter'));
      table.on('page:change', () => events.push('page'));
      table.on('selection:change', () => events.push('selection'));
      table.on('state:change', () => events.push('state'));

      // Perform operations
      table.sort('age', 'asc');
      table.addFilter({ columnId: 'department', operator: 'equals', value: 'Engineering' });
      table.goToPage(1);
      table.selectRow(1);

      // Each operation should trigger appropriate events
      expect(events).toContain('sort');
      expect(events).toContain('filter');
      expect(events).toContain('state');
      
      // state:change should be called multiple times
      const stateChangeCount = events.filter(e => e === 'state').length;
      expect(stateChangeCount).toBeGreaterThan(3);
    });
  });

  describe('Plugin Integration', () => {
    it('should integrate with export plugin', () => {
      // Mock DOM for export functionality
      const mockClick = vi.fn();
      const mockCreateElement = vi.fn().mockReturnValue({
        href: '',
        download: '',
        click: mockClick
      });

      Object.defineProperty(global, 'document', {
        value: { createElement: mockCreateElement },
        writable: true
      });

      Object.defineProperty(global, 'URL', {
        value: {
          createObjectURL: vi.fn().mockReturnValue('mock-url'),
          revokeObjectURL: vi.fn()
        },
        writable: true
      });

      Object.defineProperty(global, 'Blob', {
        value: vi.fn().mockImplementation(function(content: any[], options: any) {
          this.content = content;
          this.options = options;
        }),
        writable: true
      });

      table = new DataTable({
        columns,
        data,
        selection: { enabled: true, multiple: true }
      });

      const exportPlugin = new ExportPlugin<UserData>();
      table.use(exportPlugin);

      // Apply some filters and selections
      table.addFilter({ columnId: 'department', operator: 'equals', value: 'Engineering' });
      table.selectRow(1);
      table.selectRow(3);

      // Export selected data
      exportPlugin.export({
        format: 'csv',
        selectedOnly: true,
        filename: 'selected-users'
      });

      expect(mockClick).toHaveBeenCalled();
    });

    it('should integrate with virtual scroll plugin', () => {
      // Create large dataset
      const largeData = Array.from({ length: 10000 }, (_, i) => ({
        id: i + 1,
        firstName: `User${i + 1}`,
        lastName: `Last${i + 1}`,
        email: `user${i + 1}@company.com`,
        age: 20 + (i % 50),
        department: ['Engineering', 'Design', 'Marketing'][i % 3],
        salary: 50000 + (i * 1000),
        isActive: i % 2 === 0,
        joinDate: `2023-${String((i % 12) + 1).padStart(2, '0')}-01`,
        tags: [`tag${i % 5}`]
      }));

      table = new DataTable({
        columns,
        data: largeData,
        virtualScroll: { enabled: true, rowHeight: 48 }
      });

      const virtualScrollPlugin = new VirtualScrollPlugin<UserData>({
        rowHeight: 48,
        containerHeight: 480,
        overscan: 5
      });

      table.use(virtualScrollPlugin);

      // Mock container
      const mockContainer = {
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        scrollTop: 0
      } as any;

      virtualScrollPlugin.attachToContainer(mockContainer);

      // Apply filter - should work with virtual scrolling
      table.addFilter({ columnId: 'department', operator: 'equals', value: 'Engineering' });

      const state = virtualScrollPlugin.getState();
      expect(state.virtualRows.length).toBeLessThan(100); // Should only render visible rows
      expect(state.totalHeight).toBeGreaterThan(0);
    });

    it('should handle multiple plugins together', () => {
      // Mock DOM APIs
      const mockClick = vi.fn();
      Object.defineProperty(global, 'document', {
        value: {
          createElement: vi.fn().mockReturnValue({
            href: '',
            download: '',
            click: mockClick
          })
        },
        writable: true
      });

      Object.defineProperty(global, 'URL', {
        value: {
          createObjectURL: vi.fn().mockReturnValue('mock-url'),
          revokeObjectURL: vi.fn()
        },
        writable: true
      });

      Object.defineProperty(global, 'Blob', {
        value: vi.fn().mockImplementation(function(content: any[], options: any) {
          this.content = content;
          this.options = options;
        }),
        writable: true
      });

      table = new DataTable({
        columns,
        data,
        selection: { enabled: true, multiple: true }
      });

      const exportPlugin = new ExportPlugin<UserData>();
      const virtualScrollPlugin = new VirtualScrollPlugin<UserData>({
        rowHeight: 48,
        containerHeight: 480
      });

      // Install both plugins
      table.use(exportPlugin);
      table.use(virtualScrollPlugin);

      // Both should work together
      table.selectRow(1);
      table.sort('age', 'desc');

      exportPlugin.export({ format: 'json', selectedOnly: true });
      expect(mockClick).toHaveBeenCalled();

      const vsState = virtualScrollPlugin.getState();
      expect(vsState).toBeDefined();
    });
  });

  describe('Complex Data Operations', () => {
    it('should handle cascading data transformations', () => {
      table = new DataTable({
        columns,
        data,
        sorting: { enabled: true },
        filtering: { enabled: true },
        pagination: { enabled: true, pageSize: 2 }
      });

      // Start with all data
      expect(table.getState().rows).toHaveLength(5);

      // Apply filter (should affect pagination)
      table.addFilter({ columnId: 'department', operator: 'equals', value: 'Engineering' });
      let state = table.getState();
      expect(state.filteredRows.length).toBeLessThan(5);
      expect(state.paginationState.totalPages).toBeLessThanOrEqual(2);

      // Apply sort (should maintain filter)
      table.sort('salary', 'desc');
      state = table.getState();
      expect(state.visibleRows.every(row => row.data.department === 'Engineering')).toBe(true);
      
      // Check if sorted correctly within filtered data
      if (state.visibleRows.length > 1) {
        expect(state.visibleRows[0].data.salary).toBeGreaterThanOrEqual(
          state.visibleRows[1].data.salary
        );
      }

      // Change page (should maintain filter and sort)
      if (state.paginationState.totalPages > 1) {
        table.nextPage();
        state = table.getState();
        expect(state.visibleRows.every(row => row.data.department === 'Engineering')).toBe(true);
      }
    });

    it('should handle data updates with active features', () => {
      table = new DataTable({
        columns,
        data,
        sorting: { enabled: true },
        filtering: { enabled: true },
        selection: { enabled: true, multiple: true }
      });

      // Setup initial state
      table.sort('age', 'asc');
      table.addFilter({ columnId: 'isActive', operator: 'equals', value: true });
      table.selectRow(1);
      table.selectRow(2);

      // Add new row
      const newUser: UserData = {
        id: 6,
        firstName: 'New',
        lastName: 'User',
        email: 'new.user@company.com',
        age: 29,
        department: 'Engineering',
        salary: 70000,
        isActive: true,
        joinDate: '2023-04-01',
        tags: ['junior']
      };

      table.addRow(newUser);

      const state = table.getState();
      
      // New row should be included in sorted, filtered results
      const newRowInVisible = state.visibleRows.find(row => row.id === 6);
      expect(newRowInVisible).toBeDefined();
      expect(newRowInVisible?.data.isActive).toBe(true);

      // Should maintain sort order
      const ages = state.visibleRows.map(row => row.data.age);
      for (let i = 1; i < ages.length; i++) {
        expect(ages[i]).toBeGreaterThanOrEqual(ages[i - 1]);
      }

      // Previous selections should be maintained
      expect(table.getSelectedRows().length).toBeGreaterThanOrEqual(2);
    });

    it('should handle edge cases in combined operations', () => {
      table = new DataTable({
        columns,
        data,
        sorting: { enabled: true },
        filtering: { enabled: true },
        pagination: { enabled: true, pageSize: 10 },
        selection: { enabled: true, multiple: true }
      });

      // Select all rows
      table.selectAll();
      expect(table.getSelectedRows()).toHaveLength(5);

      // Apply filter that excludes some selected rows
      table.addFilter({ columnId: 'age', operator: 'lessThan', value: 30 });
      const state = table.getState();
      
      // Should only show rows that match filter
      expect(state.visibleRows.every(row => row.data.age < 30)).toBe(true);
      
      // Selection should be updated to only include visible rows
      const selectedAfterFilter = table.getSelectedRows();
      expect(selectedAfterFilter.every(row => row.data.age < 30)).toBe(true);

      // Clear filter and check if selection is restored correctly
      table.clearFilters();
      const finalState = table.getState();
      expect(finalState.visibleRows).toHaveLength(5);
    });
  });

  describe('Performance Integration', () => {
    it('should handle large datasets with all features enabled', () => {
      const largeData = Array.from({ length: 10000 }, (_, i) => ({
        id: i + 1,
        firstName: `First${i}`,
        lastName: `Last${i}`,
        email: `user${i}@company.com`,
        age: 20 + (i % 50),
        department: ['Engineering', 'Design', 'Marketing', 'Sales'][i % 4],
        salary: 40000 + (i * 5),
        isActive: i % 3 !== 0,
        joinDate: `2023-${String((i % 12) + 1).padStart(2, '0')}-${String((i % 28) + 1).padStart(2, '0')}`,
        tags: [`tag${i % 10}`]
      }));

      const start = performance.now();

      table = new DataTable({
        columns,
        data: largeData,
        sorting: { enabled: true, multiSort: true },
        filtering: { enabled: true, globalSearch: true },
        pagination: { enabled: true, pageSize: 50 },
        selection: { enabled: true, multiple: true }
      });

      // Perform operations
      table.sort('salary', 'desc');
      table.addFilter({ columnId: 'department', operator: 'equals', value: 'Engineering' });
      table.search('1000');
      table.selectRow(1);

      const end = performance.now();

      // Should complete in reasonable time
      expect(end - start).toBeLessThan(1000);

      // Should produce correct results
      const state = table.getState();
      expect(state.visibleRows.length).toBeGreaterThan(0);
      expect(state.visibleRows.every(row => row.data.department === 'Engineering')).toBe(true);
    });

    it('should handle rapid state changes efficiently', () => {
      table = new DataTable({
        columns,
        data,
        sorting: { enabled: true },
        filtering: { enabled: true },
        pagination: { enabled: true, pageSize: 3 }
      });

      const start = performance.now();

      // Perform many rapid operations
      for (let i = 0; i < 100; i++) {
        table.sort('age', i % 2 === 0 ? 'asc' : 'desc');
        table.goToPage(i % 2);
        if (i % 10 === 0) {
          table.addFilter({ columnId: 'isActive', operator: 'equals', value: i % 20 === 0 });
        }
      }

      const end = performance.now();

      expect(end - start).toBeLessThan(1000);
      expect(table.getState().visibleRows.length).toBeGreaterThanOrEqual(0);
    });
  });

  describe('Error Handling Integration', () => {
    it('should handle errors gracefully across components', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

      table = new DataTable({
        columns,
        data,
        sorting: { enabled: true },
        filtering: { enabled: true }
      });

      // Try invalid operations
      table.sort('nonexistent', 'asc');
      table.addFilter({ columnId: 'nonexistent', operator: 'equals', value: 'test' });
      table.updateRow(999, { firstName: 'Updated' });
      table.deleteRow(999);

      // Table should still function
      const state = table.getState();
      expect(state.rows).toHaveLength(5);
      expect(state.visibleRows).toHaveLength(5);

      consoleSpy.mockRestore();
    });

    it('should recover from plugin errors', () => {
      table = new DataTable({ columns, data });

      // Create plugin that throws error
      const errorPlugin = {
        name: 'error-plugin',
        install: () => {
          throw new Error('Plugin installation failed');
        }
      };

      expect(() => table.use(errorPlugin as any)).toThrow();

      // Table should still work normally
      table.sort('age', 'asc');
      const state = table.getState();
      expect(state.visibleRows[0].data.age).toBe(26);
    });
  });

  describe('Factory Function Integration', () => {
    it('should work with factory function and all features', () => {
      const factoryTable = createDataTable({
        columns,
        data,
        sorting: { enabled: true, multiSort: true },
        filtering: { enabled: true, globalSearch: true },
        pagination: { enabled: true, pageSize: 2 },
        selection: { enabled: true, multiple: true },
        onInit: vi.fn(),
        onDataChange: vi.fn()
      });

      expect(factoryTable).toBeInstanceOf(DataTable);

      // Test all features work
      factoryTable.sort('age', 'desc');
      factoryTable.addFilter({ columnId: 'isActive', operator: 'equals', value: true });
      factoryTable.selectRow(1);
      factoryTable.nextPage();

      const state = factoryTable.getState();
      expect(state.visibleRows.every(row => row.data.isActive)).toBe(true);
      expect(state.paginationState.pageIndex).toBe(1);
      expect(factoryTable.getSelectedRows()).toHaveLength(1);
    });
  });
});