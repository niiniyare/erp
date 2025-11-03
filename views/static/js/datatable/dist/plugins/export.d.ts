import type { Plugin, DataTableCore, RowData, ExportOptions } from '../types';
export declare class ExportPlugin<T extends RowData = RowData> implements Plugin<T> {
    name: string;
    version: string;
    private table;
    install(table: DataTableCore<T>): void;
    export(options: ExportOptions): void;
    private toCSV;
    private toJSON;
    private toTXT;
    private getCellValue;
    private escapeCSV;
    private download;
}
export declare function createExportPlugin<T extends RowData = RowData>(): ExportPlugin<T>;
//# sourceMappingURL=export.d.ts.map