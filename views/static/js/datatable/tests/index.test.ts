/**
 * Main Test Index
 * Ensures all components are exported and tested
 */

import { describe, it, expect } from 'vitest';
import * as DataTableModule from '../src/index';

describe('DataTable Module Exports', () => {
  it('should export core DataTable components', () => {
    expect(DataTableModule.DataTable).toBeDefined();
    expect(DataTableModule.createDataTable).toBeDefined();
    expect(DataTableModule.VERSION).toBeDefined();
  });

  it('should export plugins', () => {
    expect(DataTableModule.ExportPlugin).toBeDefined();
    expect(DataTableModule.createExportPlugin).toBeDefined();
    expect(DataTableModule.VirtualScrollPlugin).toBeDefined();
    expect(DataTableModule.createVirtualScrollPlugin).toBeDefined();
  });

  it('should export utilities', () => {
    expect(DataTableModule.EventEmitter).toBeDefined();
    expect(DataTableModule.StateManager).toBeDefined();
    expect(DataTableModule.FilterEngine).toBeDefined();
    expect(DataTableModule.SortEngine).toBeDefined();
    expect(DataTableModule.debounce).toBeDefined();
  });

  it('should export helper functions', () => {
    expect(DataTableModule.measureRowHeight).toBeDefined();
    expect(DataTableModule.calculateOptimalOverscan).toBeDefined();
  });

  it('should have correct version', () => {
    expect(DataTableModule.VERSION).toBe('2.0.0');
  });

  it('should export default object with main components', () => {
    const defaultExport = DataTableModule.default;
    expect(defaultExport).toHaveProperty('DataTable');
    expect(defaultExport).toHaveProperty('createDataTable');
    expect(defaultExport).toHaveProperty('VERSION');
  });
});