export class ExportPlugin {
    constructor() {
        this.name = 'export';
        this.version = '1.0.0';
        this.table = null;
    }
    install(table) {
        this.table = table;
    }
    export(options) {
        if (!this.table) {
            console.error('ExportPlugin: Table instance not available');
            return;
        }
        const state = this.table.getState();
        let rows = state.rows;
        if (options.selectedOnly) {
            rows = this.table.getSelectedRows();
        }
        else if (options.visibleOnly) {
            rows = state.visibleRows;
        }
        const columns = options.columns
            ? state.columns.filter((col) => options.columns.includes(col.id))
            : state.visibleColumns;
        let content;
        let mimeType;
        let extension;
        switch (options.format) {
            case 'csv':
                content = this.toCSV(rows, columns, options.includeHeaders ?? true);
                mimeType = 'text/csv';
                extension = 'csv';
                break;
            case 'json':
                content = this.toJSON(rows, columns);
                mimeType = 'application/json';
                extension = 'json';
                break;
            case 'txt':
                content = this.toTXT(rows, columns, options.includeHeaders ?? true);
                mimeType = 'text/plain';
                extension = 'txt';
                break;
            case 'xlsx':
                console.warn('ExportPlugin: XLSX export requires additional library');
                return;
            default:
                console.error(`ExportPlugin: Unsupported format ${options.format}`);
                return;
        }
        this.download(content, options.filename || `export-${Date.now()}`, mimeType, extension);
    }
    toCSV(rows, columns, includeHeaders) {
        const lines = [];
        if (includeHeaders) {
            const headers = columns.map((col) => this.escapeCSV(col.label || col.id));
            lines.push(headers.join(','));
        }
        rows.forEach((row) => {
            const values = columns.map((col) => {
                const value = this.getCellValue(row, col);
                return this.escapeCSV(value);
            });
            lines.push(values.join(','));
        });
        return lines.join('\n');
    }
    toJSON(rows, columns) {
        const data = rows.map((row) => {
            const obj = {};
            columns.forEach((col) => {
                obj[col.label || col.id] = this.getCellValue(row, col);
            });
            return obj;
        });
        return JSON.stringify(data, null, 2);
    }
    toTXT(rows, columns, includeHeaders) {
        const lines = [];
        const widths = columns.map((col) => {
            const headerWidth = (col.label || col.id).length;
            const maxDataWidth = rows.length > 0
                ? Math.max(...rows.map((row) => String(this.getCellValue(row, col)).length), 0)
                : 0;
            return Math.max(headerWidth, maxDataWidth);
        });
        if (includeHeaders) {
            const headers = columns
                .map((col, i) => (col.label || col.id).padEnd(widths[i] || 0))
                .join(' | ');
            lines.push(headers);
            lines.push(widths.map((w) => '-'.repeat(w || 0)).join('-+-'));
        }
        rows.forEach((row) => {
            const values = columns
                .map((col, i) => String(this.getCellValue(row, col)).padEnd(widths[i] || 0))
                .join(' | ');
            lines.push(values);
        });
        return lines.join('\n');
    }
    getCellValue(row, column) {
        if (column.accessor) {
            return column.accessor(row.data);
        }
        if (column.formatter) {
            return column.formatter(row.data[column.field], row.data);
        }
        return row.data[column.field];
    }
    escapeCSV(value) {
        if (value === null || value === undefined) {
            return '';
        }
        const str = String(value);
        if (str.includes(',') || str.includes('"') || str.includes('\n')) {
            return `"${str.replace(/"/g, '""')}"`;
        }
        return str;
    }
    download(content, filename, mimeType, extension) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${filename}.${extension}`;
        link.click();
        setTimeout(() => URL.revokeObjectURL(url), 100);
    }
}
export function createExportPlugin() {
    return new ExportPlugin();
}
//# sourceMappingURL=export.js.map