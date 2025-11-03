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
export declare class VirtualScrollPlugin<T extends RowData = RowData> implements Plugin<T> {
    name: string;
    version: string;
    private table;
    private options;
    private state;
    private scrollContainer;
    private rafId;
    constructor(options: VirtualScrollOptions);
    install(table: DataTableCore<T>): void;
    uninstall(_table: DataTableCore<T>): void;
    attachToContainer(container: HTMLElement): void;
    detachFromContainer(): void;
    getState(): VirtualScrollState<T>;
    private handleScroll;
    private calculateVisibleRows;
    private handleDataChange;
    private handleStateChange;
    scrollToRow(index: number): void;
    scrollToTop(): void;
    scrollToBottom(): void;
    private cleanup;
}
export declare function createVirtualScrollPlugin<T extends RowData = RowData>(options: VirtualScrollOptions): VirtualScrollPlugin<T>;
export declare function measureRowHeight(element: HTMLElement): number;
export declare function calculateOptimalOverscan(rowHeight: number, containerHeight: number): number;
//# sourceMappingURL=virtual-scroll.d.ts.map