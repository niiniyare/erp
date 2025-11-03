export class SortEngine {
    constructor(options) {
        this.defaultComparator = (a, b) => {
            if (a === null || a === undefined)
                return b === null || b === undefined ? 0 : -1;
            if (b === null || b === undefined)
                return 1;
            if (typeof a === 'number' && typeof b === 'number') {
                return a - b;
            }
            if (typeof a === 'boolean' && typeof b === 'boolean') {
                return a === b ? 0 : a ? 1 : -1;
            }
            if (this.isDate(a) && this.isDate(b)) {
                return a.getTime() - b.getTime();
            }
            const aNum = this.tryParseNumber(a);
            const bNum = this.tryParseNumber(b);
            if (aNum !== null && bNum !== null) {
                return aNum - bNum;
            }
            const aDate = this.tryParseDate(a);
            const bDate = this.tryParseDate(b);
            if (this.isDate(aDate) && this.isDate(bDate)) {
                return aDate.getTime() - bDate.getTime();
            }
            return String(a).localeCompare(String(b), undefined, {
                numeric: true,
                sensitivity: 'base',
            });
        };
        this.options = options;
    }
    sort(rows, sortState, columns) {
        const sortedRows = [...rows];
        if ('sorts' in sortState) {
            return this.multiSort(sortedRows, sortState.sorts, columns);
        }
        else {
            return this.singleSort(sortedRows, sortState, columns);
        }
    }
    singleSort(rows, sortState, columns) {
        const column = columns.find((col) => col.id === sortState.columnId);
        if (!column || !sortState.direction)
            return rows;
        const comparator = this.getComparator(column);
        const multiplier = sortState.direction === 'asc' ? 1 : -1;
        return rows.sort((a, b) => {
            const aValue = this.getCellValue(a, column);
            const bValue = this.getCellValue(b, column);
            return comparator(aValue, bValue) * multiplier;
        });
    }
    multiSort(rows, sorts, columns) {
        return rows.sort((a, b) => {
            for (const sort of sorts) {
                if (!sort.direction)
                    continue;
                const column = columns.find((col) => col.id === sort.columnId);
                if (!column)
                    continue;
                const comparator = this.getComparator(column);
                const multiplier = sort.direction === 'asc' ? 1 : -1;
                const aValue = this.getCellValue(a, column);
                const bValue = this.getCellValue(b, column);
                const result = comparator(aValue, bValue) * multiplier;
                if (result !== 0)
                    return result;
            }
            return 0;
        });
    }
    getComparator(column) {
        if (column.comparator) {
            return column.comparator;
        }
        if (this.options.customComparators[column.id]) {
            return this.options.customComparators[column.id];
        }
        return this.defaultComparator;
    }
    getCellValue(row, column) {
        if (column.accessor) {
            return column.accessor(row.data);
        }
        return row.data[column.field];
    }
    tryParseNumber(value) {
        if (typeof value === 'number')
            return value;
        if (typeof value === 'string') {
            const cleaned = value.replace(/[,$]/g, '');
            const num = parseFloat(cleaned);
            return isNaN(num) ? null : num;
        }
        return null;
    }
    tryParseDate(value) {
        if (this.isDate(value))
            return value;
        if (typeof value === 'string' || typeof value === 'number') {
            const date = new Date(value);
            return isNaN(date.getTime()) ? null : date;
        }
        return null;
    }
    isDate(value) {
        return value instanceof Date;
    }
}
//# sourceMappingURL=sort-engine.js.map