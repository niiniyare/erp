import type { Row, RowData, Column, SortState, MultiSortState, DataTableOptions } from '../types';
export declare class SortEngine<T extends RowData = RowData> {
    private options;
    constructor(options: Pick<Required<DataTableOptions<T>>, 'sorting' | 'customComparators'>);
    sort(rows: Row<T>[], sortState: SortState | MultiSortState, columns: Column<T>[]): Row<T>[];
    private singleSort;
    private multiSort;
    private getComparator;
    private defaultComparator;
    private getCellValue;
    private tryParseNumber;
    private tryParseDate;
    private isDate;
}
//# sourceMappingURL=sort-engine.d.ts.map