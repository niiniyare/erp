/**
 * Core DataTable Tests
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { DataTable, createDataTable } from '../src/core';
import type { Column, RowData } from '../src/types';

interface TestData extends RowData {
  id: number;
  name: string;
  age: number;
  email: string;
  active: boolean;
}

describe('DataTable Core', () => {
  let table: DataTable<TestData>;
  let columns: Column<TestData>[];
  let data: TestData[];

  beforeEach(() => {
    columns = [
      { id: 'id', field: 'id', label: 'ID' },
      { id: 'name', field: 'name', label: 'Name' },
      { id: 'age', field: 'age', label: 'Age' },
      { id: 'email', field: 'email', label: 'Email' },
      { id: 'active', field: 'active', label: 'Active' },
    ];

    data = [
      { id: 1, name: 'John Doe', age: 30, email: 'john@example.com', active: true },
      { id: 2, name: 'Jane Smith', age: 25, email: 'jane@example.com', active: false },
      { id: 3, name: 'Bob Johnson', age: 35, email: 'bob@example.com', active: true },
      { id: 4, name: 'Alice Brown', age: 28, email: 'alice@example.com', active: false },
      { id: 5, name: 'Charlie Wilson', age: 42, email: 'charlie@example.com', active: true },
    ];

    table = new DataTable({
      columns,
      data,
    });
  });

  describe('Factory Function', () => {
    it('should create DataTable instance using factory function', () => {
      const factoryTable = createDataTable({ columns, data });
      expect(factoryTable).toBeInstanceOf(DataTable);
    });
  });

  describe('Initialization', () => {
    it('should initialize with provided data and columns', () => {
      const state = table.getState();
      expect(state.rows).toHaveLength(5);
      expect(state.columns).toHaveLength(5);
      expect(state.visibleRows).toHaveLength(5);
    });

    it('should initialize with empty data', () => {
      const emptyTable = new DataTable({ columns });
      const state = emptyTable.getState();
      expect(state.rows).toHaveLength(0);
      expect(state.columns).toHaveLength(5);
    });

    it('should call onInit callback', () => {
      const onInit = vi.fn();
      new DataTable({ columns, data, onInit });
      expect(onInit).toHaveBeenCalled();
    });

    it('should emit init event', () => {
      const initHandler = vi.fn();
      const newTable = new DataTable({ columns, data: [] });
      newTable.on('init', initHandler);
      // The init event is emitted during construction, so we need to create a new table after registering the handler
      const anotherTable = new DataTable({ columns, data });
      anotherTable.on('init', initHandler);
      // Since init is called during construction, let's test by creating a table with data
      expect(initHandler).toHaveBeenCalled();
    });
  });

  describe('Data Management', () => {
    it('should set data correctly', () => {
      const newData = [
        { id: 6, name: 'David Lee', age: 33, email: 'david@example.com', active: true },
      ];
      table.setData(newData);
      const state = table.getState();
      expect(state.rows).toHaveLength(1);
      expect(state.rows[0].data.name).toBe('David Lee');
    });

    it('should get data correctly', () => {
      const rows = table.getData();
      expect(rows).toHaveLength(5);
      expect(rows[0].data.name).toBe('John Doe');
    });

    it('should add row correctly', () => {
      const newRow = { id: 6, name: 'David Lee', age: 33, email: 'david@example.com', active: true };
      table.addRow(newRow);
      const state = table.getState();
      expect(state.rows).toHaveLength(6);
      expect(state.rows[5].data.name).toBe('David Lee');
    });

    it('should update row correctly', () => {
      table.updateRow(1, { name: 'John Updated' });
      const state = table.getState();
      const updatedRow = state.rows.find(r => r.id === 1);
      expect(updatedRow?.data.name).toBe('John Updated');
      expect(updatedRow?.data.age).toBe(30); // Should preserve other fields
    });

    it('should delete row correctly', () => {
      table.deleteRow(1);
      const state = table.getState();
      expect(state.rows).toHaveLength(4);
      expect(state.rows.find(r => r.id === 1)).toBeUndefined();
    });

    it('should clear data correctly', () => {
      table.clearData();
      const state = table.getState();
      expect(state.rows).toHaveLength(0);
      expect(state.paginationState.totalRows).toBe(0);
    });

    it('should emit data:change event on data operations', () => {
      const dataChangeHandler = vi.fn();
      table.on('data:change', dataChangeHandler);
      
      table.addRow({ id: 6, name: 'David Lee', age: 33, email: 'david@example.com', active: true });
      expect(dataChangeHandler).toHaveBeenCalled();
    });
  });

  describe('Column Management', () => {
    it('should get columns correctly', () => {
      const cols = table.getColumns();
      expect(cols).toHaveLength(5);
      expect(cols[0].id).toBe('id');
    });

    it('should set columns correctly', () => {
      const newColumns = [
        { id: 'name', field: 'name', label: 'Full Name' },
        { id: 'email', field: 'email', label: 'Email Address' },
      ];
      table.setColumns(newColumns);
      const cols = table.getColumns();
      expect(cols).toHaveLength(2);
      expect(cols[0].label).toBe('Full Name');
    });

    it('should show/hide columns correctly', () => {
      table.hideColumn('age');
      let state = table.getState();
      expect(state.visibleColumns).toHaveLength(4);
      expect(state.visibleColumns.find(c => c.id === 'age')).toBeUndefined();

      table.showColumn('age');
      state = table.getState();
      expect(state.visibleColumns).toHaveLength(5);
      expect(state.visibleColumns.find(c => c.id === 'age')).toBeDefined();
    });

    it('should toggle column visibility', () => {
      table.toggleColumn('age');
      let state = table.getState();
      expect(state.visibleColumns).toHaveLength(4);

      table.toggleColumn('age');
      state = table.getState();
      expect(state.visibleColumns).toHaveLength(5);
    });

    it('should emit column:visibility event', () => {
      const visibilityHandler = vi.fn();
      table.on('column:visibility', visibilityHandler);
      
      table.hideColumn('age');
      expect(visibilityHandler).toHaveBeenCalledWith({ columnId: 'age', visible: false });
    });
  });

  describe('Sorting', () => {
    it('should sort by single column ascending', () => {
      table.sort('age', 'asc');
      const state = table.getState();
      expect(state.visibleRows[0].data.age).toBe(25); // Jane Smith
      expect(state.visibleRows[4].data.age).toBe(42); // Charlie Wilson
    });

    it('should sort by single column descending', () => {
      table.sort('age', 'desc');
      const state = table.getState();
      expect(state.visibleRows[0].data.age).toBe(42); // Charlie Wilson
      expect(state.visibleRows[4].data.age).toBe(25); // Jane Smith
    });

    it('should clear sort', () => {
      table.sort('age', 'asc');
      table.clearSort();
      const sortState = table.getSortState();
      expect(sortState).toBeUndefined();
    });

    it('should handle multi-sort', () => {
      // First create a table with multi-sort enabled
      const multiSortTable = new DataTable({
        columns,
        data,
        sorting: { enabled: true, multiSort: true }
      });
      
      multiSortTable.multiSort([
        { columnId: 'active', direction: 'desc' },
        { columnId: 'age', direction: 'asc' }
      ]);
      
      const state = multiSortTable.getState();
      // Should sort by active first (true before false), then by age ascending
      expect(state.visibleRows[0].data.active).toBe(true);
      expect(state.visibleRows[0].data.age).toBe(30); // John (active, youngest)
    });

    it('should emit sort:change event', () => {
      const sortHandler = vi.fn();
      table.on('sort:change', sortHandler);
      
      table.sort('age', 'asc');
      expect(sortHandler).toHaveBeenCalledWith({ columnId: 'age', direction: 'asc' });
    });

    it('should cycle through sort directions', () => {
      // First sort - asc
      table.sort('age');
      let sortState = table.getSortState();
      expect(sortState).toEqual({ columnId: 'age', direction: 'asc' });

      // Second sort - desc
      table.sort('age');
      sortState = table.getSortState();
      expect(sortState).toEqual({ columnId: 'age', direction: 'desc' });

      // Third sort - clear
      table.sort('age');
      sortState = table.getSortState();
      expect(sortState).toBeUndefined();
    });
  });

  describe('Filtering', () => {
    it('should filter by single column', () => {
      table.addFilter({ columnId: 'active', operator: 'equals', value: true });
      const state = table.getState();
      expect(state.visibleRows).toHaveLength(3); // John, Bob, Charlie
      expect(state.visibleRows.every(row => row.data.active)).toBe(true);
    });

    it('should filter by multiple columns', () => {
      table.filter([
        { columnId: 'active', operator: 'equals', value: true },
        { columnId: 'age', operator: 'greaterThan', value: 30 }
      ]);
      const state = table.getState();
      expect(state.visibleRows).toHaveLength(2); // Bob, Charlie
    });

    it('should perform global search', () => {
      table.search('john');
      const state = table.getState();
      expect(state.visibleRows).toHaveLength(2); // John Doe, Bob Johnson
    });

    it('should remove filter', () => {
      table.addFilter({ columnId: 'active', operator: 'equals', value: true });
      table.removeFilter('active');
      const state = table.getState();
      expect(state.visibleRows).toHaveLength(5); // All rows visible again
    });

    it('should clear all filters', () => {
      table.addFilter({ columnId: 'active', operator: 'equals', value: true });
      table.search('john');
      table.clearFilters();
      const state = table.getState();
      expect(state.visibleRows).toHaveLength(5);
      expect(state.filterState.filters).toHaveLength(0);
      expect(state.filterState.globalSearch).toBe('');
    });

    it('should emit filter:change event', () => {
      const filterHandler = vi.fn();
      table.on('filter:change', filterHandler);
      
      table.addFilter({ columnId: 'active', operator: 'equals', value: true });
      expect(filterHandler).toHaveBeenCalled();
    });

    it('should emit search:change event', () => {
      const searchHandler = vi.fn();
      table.on('search:change', searchHandler);
      
      table.search('john');
      expect(searchHandler).toHaveBeenCalledWith('john');
    });
  });

  describe('Pagination', () => {
    beforeEach(() => {
      // Create table with pagination enabled
      table = new DataTable({
        columns,
        data,
        pagination: { enabled: true, pageSize: 2 }
      });
    });

    it('should paginate correctly', () => {
      const state = table.getState();
      expect(state.visibleRows).toHaveLength(2);
      expect(state.paginationState.totalPages).toBe(3);
      expect(state.paginationState.pageIndex).toBe(0);
    });

    it('should go to next page', () => {
      table.nextPage();
      const state = table.getState();
      expect(state.paginationState.pageIndex).toBe(1);
      expect(state.visibleRows[0].data.id).toBe(3); // Bob Johnson
    });

    it('should go to previous page', () => {
      table.goToPage(1);
      table.previousPage();
      const state = table.getState();
      expect(state.paginationState.pageIndex).toBe(0);
    });

    it('should go to first/last page', () => {
      table.lastPage();
      let state = table.getState();
      expect(state.paginationState.pageIndex).toBe(2);

      table.firstPage();
      state = table.getState();
      expect(state.paginationState.pageIndex).toBe(0);
    });

    it('should change page size', () => {
      table.setPageSize(3);
      const state = table.getState();
      expect(state.paginationState.pageSize).toBe(3);
      expect(state.visibleRows).toHaveLength(3);
      expect(state.paginationState.totalPages).toBe(2);
    });

    it('should emit page:change event', () => {
      const pageHandler = vi.fn();
      table.on('page:change', pageHandler);
      
      table.nextPage();
      expect(pageHandler).toHaveBeenCalled();
    });
  });

  describe('Selection', () => {
    beforeEach(() => {
      // Create table with selection enabled
      table = new DataTable({
        columns,
        data,
        selection: { enabled: true, multiple: true }
      });
    });

    it('should select single row', () => {
      table.selectRow(1);
      const state = table.getState();
      expect(state.selectionState.selectedRowIds.has(1)).toBe(true);
      expect(state.rows.find(r => r.id === 1)?.selected).toBe(true);
    });

    it('should deselect row', () => {
      table.selectRow(1);
      table.deselectRow(1);
      const state = table.getState();
      expect(state.selectionState.selectedRowIds.has(1)).toBe(false);
      expect(state.rows.find(r => r.id === 1)?.selected).toBe(false);
    });

    it('should toggle row selection', () => {
      table.toggleRowSelection(1);
      let state = table.getState();
      expect(state.selectionState.selectedRowIds.has(1)).toBe(true);

      table.toggleRowSelection(1);
      state = table.getState();
      expect(state.selectionState.selectedRowIds.has(1)).toBe(false);
    });

    it('should select all rows', () => {
      table.selectAll();
      const state = table.getState();
      expect(state.selectionState.selectedRowIds.size).toBe(5);
      expect(state.selectionState.isAllSelected).toBe(true);
    });

    it('should deselect all rows', () => {
      table.selectAll();
      table.deselectAll();
      const state = table.getState();
      expect(state.selectionState.selectedRowIds.size).toBe(0);
      expect(state.selectionState.isAllSelected).toBe(false);
    });

    it('should get selected rows', () => {
      table.selectRow(1);
      table.selectRow(3);
      const selectedRows = table.getSelectedRows();
      expect(selectedRows).toHaveLength(2);
      expect(selectedRows.map(r => r.id)).toEqual([1, 3]);
    });

    it('should emit selection events', () => {
      const selectionHandler = vi.fn();
      const rowSelectHandler = vi.fn();
      table.on('selection:change', selectionHandler);
      table.on('row:select', rowSelectHandler);
      
      table.selectRow(1);
      expect(selectionHandler).toHaveBeenCalled();
      expect(rowSelectHandler).toHaveBeenCalledWith({ row: expect.any(Object), selected: true });
    });

    it('should handle partial selection state', () => {
      table.selectRow(1);
      table.selectRow(2);
      const state = table.getState();
      expect(state.selectionState.isPartiallySelected).toBe(true);
      expect(state.selectionState.isAllSelected).toBe(false);
    });
  });

  describe('State Management', () => {
    it('should get current state', () => {
      const state = table.getState();
      expect(state).toHaveProperty('rows');
      expect(state).toHaveProperty('columns');
      expect(state).toHaveProperty('sortState');
      expect(state).toHaveProperty('filterState');
      expect(state).toHaveProperty('paginationState');
      expect(state).toHaveProperty('selectionState');
    });

    it('should set partial state', () => {
      const newFilterState = {
        filters: [{ columnId: 'active', operator: 'equals' as const, value: true }],
        globalSearch: 'test'
      };
      table.setState({ filterState: newFilterState });
      const state = table.getState();
      expect(state.filterState).toEqual(newFilterState);
    });

    it('should reset state', () => {
      table.sort('age', 'asc');
      table.addFilter({ columnId: 'active', operator: 'equals', value: true });
      table.selectRow(1);
      
      table.resetState();
      const state = table.getState();
      expect(state.sortState).toBeUndefined();
      expect(state.filterState.filters).toHaveLength(0);
      expect(state.selectionState.selectedRowIds.size).toBe(0);
    });

    it('should emit state:change event', () => {
      const stateHandler = vi.fn();
      table.on('state:change', stateHandler);
      
      table.sort('age', 'asc');
      expect(stateHandler).toHaveBeenCalled();
    });
  });

  describe('Event System', () => {
    it('should register and unregister event listeners', () => {
      const handler = vi.fn();
      const unsubscribe = table.on('data:change', handler);
      
      table.addRow({ id: 6, name: 'Test', age: 25, email: 'test@example.com', active: true });
      expect(handler).toHaveBeenCalled();
      
      handler.mockClear();
      unsubscribe();
      
      table.addRow({ id: 7, name: 'Test2', age: 26, email: 'test2@example.com', active: true });
      expect(handler).not.toHaveBeenCalled();
    });

    it('should handle multiple listeners for same event', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      table.on('data:change', handler1);
      table.on('data:change', handler2);
      
      table.addRow({ id: 6, name: 'Test', age: 25, email: 'test@example.com', active: true });
      
      expect(handler1).toHaveBeenCalled();
      expect(handler2).toHaveBeenCalled();
    });
  });

  describe('Lifecycle', () => {
    it('should destroy table and cleanup', () => {
      const destroyHandler = vi.fn();
      table.on('destroy', destroyHandler);
      
      table.destroy();
      expect(destroyHandler).toHaveBeenCalled();
      
      // Should throw when trying to use destroyed table
      expect(() => table.addRow({ id: 6, name: 'Test', age: 25, email: 'test@example.com', active: true }))
        .toThrow('DataTable: Instance has been destroyed');
    });

    it('should prevent double destruction', () => {
      const destroyHandler = vi.fn();
      table.on('destroy', destroyHandler);
      
      table.destroy();
      table.destroy(); // Second call should be safe
      
      expect(destroyHandler).toHaveBeenCalledTimes(1);
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty data gracefully', () => {
      const emptyTable = new DataTable({ columns, data: [] });
      const state = emptyTable.getState();
      expect(state.visibleRows).toHaveLength(0);
      expect(state.paginationState.totalRows).toBe(0);
    });

    it('should handle invalid row operations gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      table.updateRow(999, { name: 'Non-existent' });
      table.deleteRow(999);
      
      expect(consoleSpy).toHaveBeenCalledTimes(2);
      consoleSpy.mockRestore();
    });

    it('should handle invalid column operations gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      table.hideColumn('non-existent');
      table.sort('non-existent', 'asc');
      
      expect(consoleSpy).toHaveBeenCalled();
      consoleSpy.mockRestore();
    });

    it('should handle disabled features gracefully', () => {
      const disabledTable = new DataTable({
        columns,
        data,
        sorting: { enabled: false },
        filtering: { enabled: false },
        pagination: { enabled: false },
        selection: { enabled: false }
      });

      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      disabledTable.sort('age', 'asc');
      disabledTable.search('john');
      disabledTable.goToPage(1);
      disabledTable.selectRow(1);
      
      expect(consoleSpy).toHaveBeenCalledTimes(4);
      consoleSpy.mockRestore();
    });
  });
});