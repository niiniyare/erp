export class VirtualScrollPlugin {
    constructor(options) {
        this.name = 'virtualScroll';
        this.version = '1.0.0';
        this.table = null;
        this.scrollContainer = null;
        this.rafId = null;
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
    install(table) {
        this.table = table;
        table.on('data:change', this.handleDataChange.bind(this));
        table.on('state:change', this.handleStateChange.bind(this));
    }
    uninstall(_table) {
        this.cleanup();
    }
    attachToContainer(container) {
        this.scrollContainer = container;
        this.scrollContainer.addEventListener('scroll', this.handleScroll.bind(this));
        this.calculateVisibleRows();
    }
    detachFromContainer() {
        if (this.scrollContainer) {
            this.scrollContainer.removeEventListener('scroll', this.handleScroll.bind(this));
            this.scrollContainer = null;
        }
    }
    getState() {
        return { ...this.state };
    }
    handleScroll(_event) {
        if (!this.scrollContainer)
            return;
        if (this.rafId) {
            cancelAnimationFrame(this.rafId);
        }
        this.rafId = requestAnimationFrame(() => {
            this.state.scrollTop = this.scrollContainer.scrollTop;
            this.calculateVisibleRows();
            if (this.options.onScroll) {
                this.options.onScroll(this.state.scrollTop);
            }
        });
    }
    calculateVisibleRows() {
        if (!this.table)
            return;
        const tableState = this.table.getState();
        const rows = tableState.paginatedRows.length > 0
            ? tableState.paginatedRows
            : tableState.visibleRows;
        const totalRows = rows.length;
        const { rowHeight, containerHeight, overscan = 5 } = this.options;
        this.state.totalHeight = totalRows * rowHeight;
        const scrollTop = this.state.scrollTop;
        const visibleStart = Math.floor(scrollTop / rowHeight);
        const visibleEnd = Math.ceil((scrollTop + containerHeight) / rowHeight);
        const start = Math.max(0, visibleStart - overscan);
        const end = Math.min(totalRows, visibleEnd + overscan);
        this.state.visibleStartIndex = start;
        this.state.visibleEndIndex = end;
        this.state.offsetY = start * rowHeight;
        this.state.virtualRows = rows.slice(start, end);
    }
    handleDataChange(_rows) {
        this.calculateVisibleRows();
    }
    handleStateChange() {
        this.calculateVisibleRows();
    }
    scrollToRow(index) {
        if (!this.scrollContainer)
            return;
        const scrollTop = index * this.options.rowHeight;
        this.scrollContainer.scrollTop = scrollTop;
    }
    scrollToTop() {
        this.scrollToRow(0);
    }
    scrollToBottom() {
        if (!this.table)
            return;
        const tableState = this.table.getState();
        const rows = tableState.paginatedRows.length > 0
            ? tableState.paginatedRows
            : tableState.visibleRows;
        this.scrollToRow(rows.length - 1);
    }
    cleanup() {
        this.detachFromContainer();
        if (this.rafId) {
            cancelAnimationFrame(this.rafId);
            this.rafId = null;
        }
    }
}
export function createVirtualScrollPlugin(options) {
    return new VirtualScrollPlugin(options);
}
export function measureRowHeight(element) {
    const computed = window.getComputedStyle(element);
    return (element.offsetHeight +
        parseFloat(computed.marginTop) +
        parseFloat(computed.marginBottom));
}
export function calculateOptimalOverscan(rowHeight, containerHeight) {
    const visibleRows = Math.ceil(containerHeight / rowHeight);
    return Math.max(3, Math.floor(visibleRows * 0.5));
}
//# sourceMappingURL=virtual-scroll.js.map