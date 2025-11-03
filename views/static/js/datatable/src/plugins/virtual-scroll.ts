/**
 * Virtual Scroll Plugin
 * Provides virtual scrolling for large datasets
 */

import type { Plugin, DataTableCore, RowData, Row } from '../types';

export interface VirtualScrollOptions {
  rowHeight: number;
  containerHeight: number;
  overscan?: number;
  onScroll?: (scrollTop: number) => void;
}

export interface VirtualScrollState<T = RowData> {
  scrollTop: number;
  visibleStartIndex: number;
  visibleEndIndex: number;
  virtualRows: Row<T>[];
  totalHeight: number;
  offsetY: number;
}

export class VirtualScrollPlugin<T extends RowData = RowData> implements Plugin<T> {
  name = 'virtualScroll';
  version = '1.0.0';

  private table: DataTableCore<T> | null = null;
  private options: VirtualScrollOptions;
  private state: VirtualScrollState<T>;
  private scrollContainer: HTMLElement | null = null;
  private rafId: number | null = null;

  constructor(options: VirtualScrollOptions) {
    this.options = {
      overscan: 5,
      ...options,
    };

    this.state = {
      scrollTop: 0,
      visibleStartIndex: 0,
      visibleEndIndex: 0,
      virtualRows: [],
      totalHeight: 0,
      offsetY: 0,
    };
  }

  install(table: DataTableCore<T>): void {
    this.table = table;

    // Listen to data changes
    table.on('data:change', this.handleDataChange.bind(this));
    table.on('state:change', this.handleStateChange.bind(this));
  }

  uninstall(_table: DataTableCore<T>): void {
    this.cleanup();
  }

  /**
   * Attach to scroll container
   */
  attachToContainer(container: HTMLElement): void {
    this.scrollContainer = container;
    this.scrollContainer.addEventListener('scroll', this.handleScroll.bind(this));
    this.calculateVisibleRows();
  }

  /**
   * Detach from scroll container
   */
  detachFromContainer(): void {
    if (this.scrollContainer) {
      this.scrollContainer.removeEventListener('scroll', this.handleScroll.bind(this));
      this.scrollContainer = null;
    }
  }

  /**
   * Get virtual scroll state
   */
  getState(): VirtualScrollState<T> {
    return { ...this.state };
  }

  /**
   * Handle scroll event
   */
  private handleScroll(_event: Event): void {
    if (!this.scrollContainer) return;

    // Cancel previous RAF
    if (this.rafId) {
      cancelAnimationFrame(this.rafId);
    }

    // Schedule calculation on next frame
    this.rafId = requestAnimationFrame(() => {
      this.state.scrollTop = this.scrollContainer!.scrollTop;
      this.calculateVisibleRows();

      if (this.options.onScroll) {
        this.options.onScroll(this.state.scrollTop);
      }
    });
  }

  /**
   * Calculate visible rows based on scroll position
   */
  private calculateVisibleRows(): void {
    if (!this.table) return;

    const tableState = this.table.getState();
    const rows = tableState.paginatedRows.length > 0
      ? tableState.paginatedRows
      : tableState.visibleRows;

    const totalRows = rows.length;
    const { rowHeight, containerHeight, overscan = 5 } = this.options;

    // Calculate total height
    this.state.totalHeight = totalRows * rowHeight;

    // Calculate visible range
    const scrollTop = this.state.scrollTop;
    const visibleStart = Math.floor(scrollTop / rowHeight);
    const visibleEnd = Math.ceil((scrollTop + containerHeight) / rowHeight);

    // Add overscan
    const start = Math.max(0, visibleStart - overscan);
    const end = Math.min(totalRows, visibleEnd + overscan);

    this.state.visibleStartIndex = start;
    this.state.visibleEndIndex = end;
    this.state.offsetY = start * rowHeight;

    // Get virtual rows
    this.state.virtualRows = rows.slice(start, end);
  }

  /**
   * Handle data change
   */
  private handleDataChange(_rows: Row<T>[]): void {
    this.calculateVisibleRows();
  }

  /**
   * Handle state change
   */
  private handleStateChange(): void {
    this.calculateVisibleRows();
  }

  /**
   * Scroll to row
   */
  scrollToRow(index: number): void {
    if (!this.scrollContainer) return;

    const scrollTop = index * this.options.rowHeight;
    this.scrollContainer.scrollTop = scrollTop;
  }

  /**
   * Scroll to top
   */
  scrollToTop(): void {
    this.scrollToRow(0);
  }

  /**
   * Scroll to bottom
   */
  scrollToBottom(): void {
    if (!this.table) return;

    const tableState = this.table.getState();
    const rows = tableState.paginatedRows.length > 0
      ? tableState.paginatedRows
      : tableState.visibleRows;

    this.scrollToRow(rows.length - 1);
  }

  /**
   * Cleanup
   */
  private cleanup(): void {
    this.detachFromContainer();

    if (this.rafId) {
      cancelAnimationFrame(this.rafId);
      this.rafId = null;
    }
  }
}

// Factory function
export function createVirtualScrollPlugin<T extends RowData = RowData>(
  options: VirtualScrollOptions
): VirtualScrollPlugin<T> {
  return new VirtualScrollPlugin<T>(options);
}

/**
 * Helper: Calculate row height from element
 */
export function measureRowHeight(element: HTMLElement): number {
  const computed = window.getComputedStyle(element);
  return (
    element.offsetHeight +
    parseFloat(computed.marginTop) +
    parseFloat(computed.marginBottom)
  );
}

/**
 * Helper: Get optimal overscan based on viewport
 */
export function calculateOptimalOverscan(
  rowHeight: number,
  containerHeight: number
): number {
  const visibleRows = Math.ceil(containerHeight / rowHeight);
  return Math.max(3, Math.floor(visibleRows * 0.5));
}
