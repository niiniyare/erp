import type { Row, RowData, Column, FilterState, DataTableOptions } from '../types';
export declare class FilterEngine<T extends RowData = RowData> {
    private options;
    constructor(options: Pick<Required<DataTableOptions<T>>, 'filtering'>);
    filter(rows: Row<T>[], filterState: FilterState, columns: Column<T>[]): Row<T>[];
    private applyColumnFilters;
    private applyGlobalSearch;
    private evaluateFilter;
    private applyOperator;
    private getCellValue;
    private normalizeValue;
    private toNumber;
    private compareNumeric;
}
export declare function debounce<T extends (...args: any[]) => any>(func: T, wait: number): (...args: Parameters<T>) => void;
//# sourceMappingURL=filter-engine.d.ts.map