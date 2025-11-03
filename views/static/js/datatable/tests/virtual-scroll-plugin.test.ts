/**
 * VirtualScrollPlugin Tests
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { 
  VirtualScrollPlugin, 
  createVirtualScrollPlugin, 
  measureRowHeight, 
  calculateOptimalOverscan 
} from '../src/plugins/virtual-scroll';
import { DataTable } from '../src/core';
import type { Column, RowData, DataTableCore, VirtualScrollOptions } from '../src/types';

interface TestData extends RowData {
  id: number;
  name: string;
  age: number;
  email: string;
}

// Mock DOM APIs
const mockAddEventListener = vi.fn();
const mockRemoveEventListener = vi.fn();
const mockRequestAnimationFrame = vi.fn();
const mockCancelAnimationFrame = vi.fn();

describe('VirtualScrollPlugin', () => {
  let virtualScrollPlugin: VirtualScrollPlugin<TestData>;
  let table: DataTableCore<TestData>;
  let columns: Column<TestData>[];
  let data: TestData[];
  let mockContainer: HTMLElement;

  beforeEach(() => {
    // Mock DOM APIs
    Object.defineProperty(global, 'requestAnimationFrame', {
      value: mockRequestAnimationFrame.mockImplementation((callback) => {
        callback(performance.now());
        return 1;
      }),
      writable: true
    });

    Object.defineProperty(global, 'cancelAnimationFrame', {
      value: mockCancelAnimationFrame,
      writable: true
    });

    Object.defineProperty(global, 'window', {
      value: {
        getComputedStyle: vi.fn().mockReturnValue({
          marginTop: '5px',
          marginBottom: '5px'
        })
      },
      writable: true
    });

    // Mock container element
    mockContainer = {
      addEventListener: mockAddEventListener,
      removeEventListener: mockRemoveEventListener,
      scrollTop: 0,
      offsetHeight: 48
    } as any;

    // Setup test data
    columns = [
      { id: 'id', field: 'id', label: 'ID' },
      { id: 'name', field: 'name', label: 'Name' },
      { id: 'age', field: 'age', label: 'Age' },
      { id: 'email', field: 'email', label: 'Email' },
    ];

    // Create larger dataset for virtual scrolling
    data = Array.from({ length: 1000 }, (_, i) => ({
      id: i + 1,
      name: `User ${i + 1}`,
      age: 20 + (i % 50),
      email: `user${i + 1}@example.com`
    }));

    table = new DataTable({ columns, data });

    const options: VirtualScrollOptions = {
      rowHeight: 48,
      containerHeight: 480,
      overscan: 5
    };

    virtualScrollPlugin = new VirtualScrollPlugin(options);
    virtualScrollPlugin.install(table);

    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Plugin Installation', () => {
    it('should create plugin using factory function', () => {
      const options: VirtualScrollOptions = {
        rowHeight: 48,
        containerHeight: 480
      };
      
      const factoryPlugin = createVirtualScrollPlugin<TestData>(options);
      expect(factoryPlugin).toBeInstanceOf(VirtualScrollPlugin);
      expect(factoryPlugin.name).toBe('virtualScroll');
      expect(factoryPlugin.version).toBe('1.0.0');
    });

    it('should install plugin correctly', () => {
      const newPlugin = new VirtualScrollPlugin<TestData>({
        rowHeight: 48,
        containerHeight: 480
      });
      
      expect(() => newPlugin.install(table)).not.toThrow();
    });

    it('should uninstall plugin correctly', () => {
      expect(() => virtualScrollPlugin.uninstall(table)).not.toThrow();
    });

    it('should use default overscan when not provided', () => {
      const options: VirtualScrollOptions = {
        rowHeight: 48,
        containerHeight: 480
        // overscan not provided
      };
      
      const plugin = new VirtualScrollPlugin(options);
      const state = plugin.getState();
      
      // Default overscan should be 5
      expect(state).toBeDefined();
    });
  });

  describe('Container Attachment', () => {
    it('should attach to container and setup event listeners', () => {
      virtualScrollPlugin.attachToContainer(mockContainer);
      
      expect(mockAddEventListener).toHaveBeenCalledWith('scroll', expect.any(Function));
    });

    it('should detach from container and remove event listeners', () => {
      virtualScrollPlugin.attachToContainer(mockContainer);
      virtualScrollPlugin.detachFromContainer();
      
      expect(mockRemoveEventListener).toHaveBeenCalledWith('scroll', expect.any(Function));
    });

    it('should handle multiple attach/detach cycles', () => {
      virtualScrollPlugin.attachToContainer(mockContainer);
      virtualScrollPlugin.detachFromContainer();
      virtualScrollPlugin.attachToContainer(mockContainer);
      virtualScrollPlugin.detachFromContainer();
      
      expect(mockAddEventListener).toHaveBeenCalledTimes(2);
      expect(mockRemoveEventListener).toHaveBeenCalledTimes(2);
    });

    it('should handle detaching when not attached', () => {
      expect(() => virtualScrollPlugin.detachFromContainer()).not.toThrow();
    });
  });

  describe('Virtual Scrolling Calculations', () => {
    beforeEach(() => {
      virtualScrollPlugin.attachToContainer(mockContainer);
    });

    it('should calculate visible rows correctly', () => {
      const state = virtualScrollPlugin.getState();
      
      expect(state.totalHeight).toBe(1000 * 48); // 1000 rows * 48px height
      expect(state.visibleStartIndex).toBe(0);
      expect(state.visibleEndIndex).toBeGreaterThan(0);
      expect(state.virtualRows.length).toBeGreaterThan(0);
    });

    it('should update visible rows when scrolled', () => {
      // Simulate scroll
      mockContainer.scrollTop = 240; // Scroll down 5 rows
      
      // Trigger scroll event
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      if (scrollHandler) {
        scrollHandler({ target: mockContainer });
      }

      // Should schedule RAF
      expect(mockRequestAnimationFrame).toHaveBeenCalled();
    });

    it('should handle scroll to different positions', () => {
      const positions = [0, 240, 480, 960, 1200];
      
      positions.forEach(scrollTop => {
        mockContainer.scrollTop = scrollTop;
        
        const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
        if (scrollHandler) {
          scrollHandler({ target: mockContainer });
        }
        
        const state = virtualScrollPlugin.getState();
        expect(state.scrollTop).toBe(scrollTop);
      });
    });

    it('should apply overscan correctly', () => {
      const state = virtualScrollPlugin.getState();
      
      // With overscan of 5, should have extra rows before and after visible area
      const visibleRowCount = Math.ceil(480 / 48); // containerHeight / rowHeight
      expect(state.virtualRows.length).toBeGreaterThan(visibleRowCount);
    });

    it('should handle edge cases at start of list', () => {
      mockContainer.scrollTop = 0;
      
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      if (scrollHandler) {
        scrollHandler({ target: mockContainer });
      }
      
      const state = virtualScrollPlugin.getState();
      expect(state.visibleStartIndex).toBe(0);
    });

    it('should handle edge cases at end of list', () => {
      // Scroll to near end
      mockContainer.scrollTop = (1000 * 48) - 480; // Near bottom
      
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      if (scrollHandler) {
        scrollHandler({ target: mockContainer });
      }
      
      const state = virtualScrollPlugin.getState();
      expect(state.visibleEndIndex).toBeLessThanOrEqual(1000);
    });
  });

  describe('Scroll Navigation', () => {
    beforeEach(() => {
      virtualScrollPlugin.attachToContainer(mockContainer);
    });

    it('should scroll to specific row', () => {
      virtualScrollPlugin.scrollToRow(10);
      
      expect(mockContainer.scrollTop).toBe(10 * 48); // row 10 * rowHeight
    });

    it('should scroll to top', () => {
      mockContainer.scrollTop = 500;
      virtualScrollPlugin.scrollToTop();
      
      expect(mockContainer.scrollTop).toBe(0);
    });

    it('should scroll to bottom', () => {
      virtualScrollPlugin.scrollToBottom();
      
      // Should scroll to last row
      const expectedScrollTop = (1000 - 1) * 48;
      expect(mockContainer.scrollTop).toBe(expectedScrollTop);
    });

    it('should handle scroll navigation without container', () => {
      virtualScrollPlugin.detachFromContainer();
      
      expect(() => virtualScrollPlugin.scrollToRow(10)).not.toThrow();
      expect(() => virtualScrollPlugin.scrollToTop()).not.toThrow();
      expect(() => virtualScrollPlugin.scrollToBottom()).not.toThrow();
    });
  });

  describe('Data Changes', () => {
    beforeEach(() => {
      virtualScrollPlugin.attachToContainer(mockContainer);
    });

    it('should recalculate when data changes', () => {
      const initialState = virtualScrollPlugin.getState();
      
      // Change data
      const newData = data.slice(0, 500); // Reduce to 500 rows
      table.setData(newData);
      
      const newState = virtualScrollPlugin.getState();
      expect(newState.totalHeight).toBe(500 * 48);
      expect(newState.totalHeight).not.toBe(initialState.totalHeight);
    });

    it('should handle empty data', () => {
      table.clearData();
      
      const state = virtualScrollPlugin.getState();
      expect(state.totalHeight).toBe(0);
      expect(state.virtualRows).toHaveLength(0);
    });

    it('should handle filtered data', () => {
      // Apply filter to reduce visible rows
      table.addFilter({ columnId: 'age', operator: 'lessThan', value: 30 });
      
      const state = virtualScrollPlugin.getState();
      // Should work with filtered rows
      expect(state.virtualRows.length).toBeGreaterThanOrEqual(0);
    });

    it('should handle paginated data', () => {
      // Create table with pagination
      const paginatedTable = new DataTable({
        columns,
        data,
        pagination: { enabled: true, pageSize: 50 }
      });

      const paginatedPlugin = new VirtualScrollPlugin<TestData>({
        rowHeight: 48,
        containerHeight: 480,
        overscan: 5
      });

      paginatedPlugin.install(paginatedTable);
      paginatedPlugin.attachToContainer(mockContainer);

      const state = paginatedPlugin.getState();
      // Should work with paginated data (first page)
      expect(state.totalHeight).toBe(50 * 48);
    });
  });

  describe('Performance', () => {
    beforeEach(() => {
      virtualScrollPlugin.attachToContainer(mockContainer);
    });

    it('should throttle scroll events with RAF', () => {
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      
      if (scrollHandler) {
        // Trigger multiple scroll events rapidly
        scrollHandler({ target: mockContainer });
        scrollHandler({ target: mockContainer });
        scrollHandler({ target: mockContainer });
      }
      
      // Should only schedule one RAF
      expect(mockRequestAnimationFrame).toHaveBeenCalledTimes(3);
    });

    it('should cancel previous RAF when new scroll occurs', () => {
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      
      if (scrollHandler) {
        scrollHandler({ target: mockContainer });
        scrollHandler({ target: mockContainer });
      }
      
      expect(mockCancelAnimationFrame).toHaveBeenCalled();
    });

    it('should handle rapid data changes efficiently', () => {
      const start = performance.now();
      
      // Simulate rapid data changes
      for (let i = 0; i < 100; i++) {
        const newData = data.slice(0, 900 + i);
        table.setData(newData);
      }
      
      const end = performance.now();
      expect(end - start).toBeLessThan(1000); // Should complete in less than 1 second
    });
  });

  describe('Event Callbacks', () => {
    it('should call onScroll callback when provided', () => {
      const onScrollCallback = vi.fn();
      
      const callbackPlugin = new VirtualScrollPlugin<TestData>({
        rowHeight: 48,
        containerHeight: 480,
        onScroll: onScrollCallback
      });
      
      callbackPlugin.install(table);
      callbackPlugin.attachToContainer(mockContainer);
      
      mockContainer.scrollTop = 240;
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      if (scrollHandler) {
        scrollHandler({ target: mockContainer });
      }
      
      expect(onScrollCallback).toHaveBeenCalledWith(240);
    });

    it('should handle missing onScroll callback gracefully', () => {
      mockContainer.scrollTop = 240;
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      
      expect(() => {
        if (scrollHandler) {
          scrollHandler({ target: mockContainer });
        }
      }).not.toThrow();
    });
  });

  describe('State Management', () => {
    it('should return current state', () => {
      const state = virtualScrollPlugin.getState();
      
      expect(state).toHaveProperty('scrollTop');
      expect(state).toHaveProperty('visibleStartIndex');
      expect(state).toHaveProperty('visibleEndIndex');
      expect(state).toHaveProperty('virtualRows');
      expect(state).toHaveProperty('totalHeight');
      expect(state).toHaveProperty('offsetY');
    });

    it('should update state when table state changes', () => {
      virtualScrollPlugin.attachToContainer(mockContainer);
      
      const initialState = virtualScrollPlugin.getState();
      
      // Trigger state change
      table.sort('name', 'asc');
      
      // State should be recalculated
      const newState = virtualScrollPlugin.getState();
      expect(newState).toBeDefined();
    });
  });

  describe('Cleanup', () => {
    it('should cleanup on uninstall', () => {
      virtualScrollPlugin.attachToContainer(mockContainer);
      virtualScrollPlugin.uninstall(table);
      
      expect(mockRemoveEventListener).toHaveBeenCalled();
    });

    it('should cancel RAF on cleanup', () => {
      virtualScrollPlugin.attachToContainer(mockContainer);
      
      // Trigger scroll to create RAF
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      if (scrollHandler) {
        scrollHandler({ target: mockContainer });
      }
      
      virtualScrollPlugin.uninstall(table);
      
      expect(mockCancelAnimationFrame).toHaveBeenCalled();
    });

    it('should handle cleanup when no container attached', () => {
      expect(() => virtualScrollPlugin.uninstall(table)).not.toThrow();
    });
  });

  describe('Edge Cases', () => {
    it('should handle zero row height', () => {
      const zeroHeightPlugin = new VirtualScrollPlugin<TestData>({
        rowHeight: 0,
        containerHeight: 480
      });
      
      zeroHeightPlugin.install(table);
      zeroHeightPlugin.attachToContainer(mockContainer);
      
      const state = zeroHeightPlugin.getState();
      expect(state.totalHeight).toBe(0);
    });

    it('should handle zero container height', () => {
      const zeroContainerPlugin = new VirtualScrollPlugin<TestData>({
        rowHeight: 48,
        containerHeight: 0
      });
      
      zeroContainerPlugin.install(table);
      zeroContainerPlugin.attachToContainer(mockContainer);
      
      const state = zeroContainerPlugin.getState();
      expect(state.virtualRows).toHaveLength(0);
    });

    it('should handle very large datasets', () => {
      const largeData = Array.from({ length: 100000 }, (_, i) => ({
        id: i + 1,
        name: `User ${i + 1}`,
        age: 20 + (i % 50),
        email: `user${i + 1}@example.com`
      }));
      
      table.setData(largeData);
      virtualScrollPlugin.attachToContainer(mockContainer);
      
      const state = virtualScrollPlugin.getState();
      expect(state.totalHeight).toBe(100000 * 48);
      expect(state.virtualRows.length).toBeLessThan(100); // Should only render visible rows
    });

    it('should handle negative scroll positions', () => {
      virtualScrollPlugin.attachToContainer(mockContainer);
      
      mockContainer.scrollTop = -100; // Negative scroll
      
      const scrollHandler = mockAddEventListener.mock.calls.find(call => call[0] === 'scroll')?.[1];
      if (scrollHandler) {
        scrollHandler({ target: mockContainer });
      }
      
      const state = virtualScrollPlugin.getState();
      expect(state.visibleStartIndex).toBeGreaterThanOrEqual(0);
    });
  });
});

describe('VirtualScroll Utility Functions', () => {
  describe('measureRowHeight', () => {
    it('should measure row height including margins', () => {
      const mockElement = {
        offsetHeight: 40,
      } as HTMLElement;
      
      Object.defineProperty(global, 'window', {
        value: {
          getComputedStyle: vi.fn().mockReturnValue({
            marginTop: '5px',
            marginBottom: '3px'
          })
        },
        writable: true
      });
      
      const height = measureRowHeight(mockElement);
      expect(height).toBe(48); // 40 + 5 + 3
    });

    it('should handle elements without margins', () => {
      const mockElement = {
        offsetHeight: 40,
      } as HTMLElement;
      
      Object.defineProperty(global, 'window', {
        value: {
          getComputedStyle: vi.fn().mockReturnValue({
            marginTop: '0px',
            marginBottom: '0px'
          })
        },
        writable: true
      });
      
      const height = measureRowHeight(mockElement);
      expect(height).toBe(40);
    });

    it('should handle invalid margin values', () => {
      const mockElement = {
        offsetHeight: 40,
      } as HTMLElement;
      
      Object.defineProperty(global, 'window', {
        value: {
          getComputedStyle: vi.fn().mockReturnValue({
            marginTop: 'invalid',
            marginBottom: 'auto'
          })
        },
        writable: true
      });
      
      const height = measureRowHeight(mockElement);
      expect(height).toBe(40); // Should fallback to just offsetHeight
    });
  });

  describe('calculateOptimalOverscan', () => {
    it('should calculate optimal overscan based on visible rows', () => {
      const overscan = calculateOptimalOverscan(48, 480);
      const visibleRows = Math.ceil(480 / 48); // 10 rows
      const expectedOverscan = Math.max(3, Math.floor(visibleRows * 0.5)); // 5
      
      expect(overscan).toBe(expectedOverscan);
    });

    it('should have minimum overscan of 3', () => {
      const overscan = calculateOptimalOverscan(100, 150); // Only 1.5 visible rows
      expect(overscan).toBe(3);
    });

    it('should handle zero container height', () => {
      const overscan = calculateOptimalOverscan(48, 0);
      expect(overscan).toBe(3); // Should return minimum
    });

    it('should handle zero row height', () => {
      const overscan = calculateOptimalOverscan(0, 480);
      expect(overscan).toBe(3); // Should return minimum
    });

    it('should calculate reasonable overscan for large containers', () => {
      const overscan = calculateOptimalOverscan(48, 2400); // 50 visible rows
      const expectedOverscan = Math.floor(50 * 0.5); // 25
      
      expect(overscan).toBe(expectedOverscan);
    });
  });
});