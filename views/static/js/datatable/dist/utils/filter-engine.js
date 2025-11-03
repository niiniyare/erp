export class FilterEngine {
    constructor(options) {
        this.options = options;
    }
    filter(rows, filterState, columns) {
        let filteredRows = rows;
        if (filterState.filters.length > 0) {
            filteredRows = this.applyColumnFilters(filteredRows, filterState.filters, columns);
        }
        if (filterState.globalSearch && filterState.globalSearch.trim()) {
            filteredRows = this.applyGlobalSearch(filteredRows, filterState.globalSearch, columns);
        }
        return filteredRows;
    }
    applyColumnFilters(rows, filters, columns) {
        return rows.filter((row) => {
            return filters.every((filter) => this.evaluateFilter(row, filter, columns));
        });
    }
    applyGlobalSearch(rows, query, columns) {
        const searchableColumns = columns.filter((col) => col.filterable !== false);
        const normalizedQuery = this.options.filtering.caseSensitive
            ? query
            : query.toLowerCase();
        return rows.filter((row) => {
            return searchableColumns.some((column) => {
                const value = this.getCellValue(row, column);
                const normalizedValue = this.normalizeValue(value);
                return normalizedValue.includes(normalizedQuery);
            });
        });
    }
    evaluateFilter(row, filter, columns) {
        if (this.options.filtering.customFilters?.[filter.columnId]) {
            return this.options.filtering.customFilters[filter.columnId](row, filter);
        }
        const column = columns.find((col) => col.id === filter.columnId);
        if (!column)
            return true;
        const cellValue = this.getCellValue(row, column);
        return this.applyOperator(cellValue, filter.operator, filter.value);
    }
    applyOperator(cellValue, operator, filterValue) {
        const normalizedCellValue = this.normalizeValue(cellValue);
        const normalizedFilterValue = Array.isArray(filterValue)
            ? filterValue.map((v) => this.normalizeValue(v))
            : this.normalizeValue(filterValue);
        switch (operator) {
            case 'equals':
                return normalizedCellValue === normalizedFilterValue;
            case 'notEquals':
                return normalizedCellValue !== normalizedFilterValue;
            case 'contains':
                return normalizedCellValue.includes(normalizedFilterValue);
            case 'notContains':
                return !normalizedCellValue.includes(normalizedFilterValue);
            case 'startsWith':
                return normalizedCellValue.startsWith(normalizedFilterValue);
            case 'endsWith':
                return normalizedCellValue.endsWith(normalizedFilterValue);
            case 'isEmpty':
                return cellValue === null || cellValue === undefined || cellValue === '';
            case 'isNotEmpty':
                return cellValue !== null && cellValue !== undefined && cellValue !== '';
            case 'greaterThan':
                return this.compareNumeric(cellValue, filterValue) > 0;
            case 'greaterThanOrEqual':
                return this.compareNumeric(cellValue, filterValue) >= 0;
            case 'lessThan':
                return this.compareNumeric(cellValue, filterValue) < 0;
            case 'lessThanOrEqual':
                return this.compareNumeric(cellValue, filterValue) <= 0;
            case 'between':
                if (!Array.isArray(filterValue) || filterValue.length !== 2)
                    return false;
                const num = this.toNumber(cellValue);
                const min = this.toNumber(filterValue[0]);
                const max = this.toNumber(filterValue[1]);
                return num >= min && num <= max;
            case 'in':
                if (!Array.isArray(filterValue))
                    return false;
                return normalizedFilterValue.includes(normalizedCellValue);
            case 'notIn':
                if (!Array.isArray(filterValue))
                    return false;
                return !normalizedFilterValue.includes(normalizedCellValue);
            default:
                console.warn(`FilterEngine: Unknown operator ${operator}`);
                return true;
        }
    }
    getCellValue(row, column) {
        if (column.accessor) {
            return column.accessor(row.data);
        }
        return row.data[column.field];
    }
    normalizeValue(value) {
        if (value === null || value === undefined) {
            return '';
        }
        const stringValue = String(value);
        return this.options.filtering.caseSensitive
            ? stringValue
            : stringValue.toLowerCase();
    }
    toNumber(value) {
        if (typeof value === 'number')
            return value;
        if (typeof value === 'string') {
            const num = parseFloat(value);
            return isNaN(num) ? 0 : num;
        }
        return 0;
    }
    compareNumeric(a, b) {
        return this.toNumber(a) - this.toNumber(b);
    }
}
export function debounce(func, wait) {
    let timeout = null;
    return function (...args) {
        if (timeout) {
            clearTimeout(timeout);
        }
        timeout = setTimeout(() => {
            func(...args);
        }, wait);
    };
}
//# sourceMappingURL=filter-engine.js.map