package table

import (
	"fmt"
	"strings"
)

// getDataTableWrapperClasses returns CSS classes for the table wrapper
func getDataTableWrapperClasses(props DataTableProps) string {
	classes := []string{"data-table-wrapper", "w-full"}

	if props.Class != "" {
		classes = append(classes, props.Class)
	}

	return strings.Join(classes, " ")
}

// getDataTableClasses returns CSS classes for the main table element
func getDataTableClasses(props DataTableProps) string {
	classes := []string{
		"w-full",
		"text-sm",
		"text-left",
		"text-gray-500",
		"dark:text-gray-400",
	}

	if props.Striped {
		classes = append(classes, "stripe")
	}

	if props.Bordered {
		classes = append(classes, "border", "border-gray-200", "dark:border-gray-700")
	}

	if props.Hover {
		classes = append(classes, "hover")
	}

	if props.Compact {
		classes = append(classes, "compact")
	}

	return strings.Join(classes, " ")
}

// getTableHeaderClasses returns CSS classes for table header cells
func getTableHeaderClasses(column DataTableColumn, sortable bool) string {
	classes := []string{"px-6", "py-3", "text-left", "text-xs", "font-medium", "text-gray-500", "uppercase", "tracking-wider"}

	if sortable && column.Sortable {
		classes = append(classes,
			"cursor-pointer",
			"select-none",
			"hover:bg-gray-100",
			"dark:hover:bg-gray-600",
			"transition-colors",
			"duration-150",
		)
	}

	if column.Class != "" {
		classes = append(classes, column.Class)
	}

	return strings.Join(classes, " ")
}

// getTableCellClasses returns CSS classes for table data cells
func getTableCellClasses(column DataTableColumn) string {
	classes := []string{"px-6", "py-4", "whitespace-nowrap"}

	switch column.Type {
	case "number":
		classes = append(classes, "text-right", "font-mono")
	case "date":
		classes = append(classes, "text-gray-900", "dark:text-white")
	case "badge":
		classes = append(classes, "text-center")
	case "status":
		classes = append(classes, "text-gray-900", "dark:text-white")
	default:
		classes = append(classes, "font-medium", "text-gray-900", "dark:text-white")
	}

	if column.Class != "" {
		classes = append(classes, column.Class)
	}

	return strings.Join(classes, " ")
}

// getDataTableAlpineData returns the Alpine.js data object for the table
func getDataTableAlpineData(props DataTableProps) string {
	pageLength := props.Config.PageLength
	if pageLength == 0 {
		pageLength = 25
	}

	return fmt.Sprintf(`{
		selectedRows: [],
		showFilters: false,
		sortColumn: '',
		sortDirection: 'asc',
		pageInfo: {
			current: 1,
			total: 0,
			start: 0,
			end: 0,
			totalPages: 0,
			pageSize: %d
		},
		paginationPages: [],
		
		// Row selection methods
		toggleRow(index) {
			const pos = this.selectedRows.indexOf(index);
			if (pos > -1) {
				this.selectedRows.splice(pos, 1);
			} else {
				this.selectedRows.push(index);
			}
		},
		
		toggleSelectAll() {
			if (this.selectedRows.length > 0) {
				this.selectedRows = [];
			} else {
				// TODO: Select all visible rows based on current page
				// FIXME: Implement proper select all logic with pagination
				this.selectedRows = Array.from({length: 10}, (_, i) => i); // Placeholder
			}
		},
		
		clearSelection() {
			this.selectedRows = [];
		},
		
		// Filter methods
		toggleFilters() {
			this.showFilters = !this.showFilters;
		},
		
		// Sorting methods
		sortBy(column) {
			if (this.sortColumn === column) {
				this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
			} else {
				this.sortColumn = column;
				this.sortDirection = 'asc';
			}
			
			// TODO: Trigger HTMX request with sort parameters
			// FIXME: Implement server-side sorting integration
			this.refreshTable();
		},
		
		// Pagination methods
		goToPage(page) {
			if (page >= 1 && page <= this.pageInfo.totalPages) {
				this.pageInfo.current = page;
				this.refreshTable();
			}
		},
		
		changePageSize(newSize) {
			this.pageInfo.pageSize = parseInt(newSize);
			this.pageInfo.current = 1; // Reset to first page
			this.refreshTable();
		},
		
		// Data refresh methods
		refreshTable() {
			// TODO: Implement table refresh with current state
			// FIXME: Build query parameters and trigger HTMX request
			const params = this.buildQueryParams();
			// htmx.trigger('#table-id', 'refreshTable', {params});
		},
		
		buildQueryParams() {
			return {
				page: this.pageInfo.current,
				pageSize: this.pageInfo.pageSize,
				sortColumn: this.sortColumn,
				sortDirection: this.sortDirection,
				// TODO: Add filter parameters
			};
		},
		
		// Bulk action methods
		handleBulkAction(actionName) {
			if (this.selectedRows.length === 0) {
				// TODO: Show notification that no rows are selected
				return;
			}
			
			// TODO: Implement bulk action handling
			// FIXME: Send selected row IDs to server
			console.log('Bulk action:', actionName, 'on rows:', this.selectedRows);
		},
		
		// Update pagination info (called from server responses)
		updatePaginationInfo(info) {
			this.pageInfo = { ...this.pageInfo, ...info };
			this.updatePaginationPages();
		},
		
		updatePaginationPages() {
			const current = this.pageInfo.current;
			const total = this.pageInfo.totalPages;
			const delta = 2; // Show 2 pages before and after current
			
			let pages = [];
			const start = Math.max(1, current - delta);
			const end = Math.min(total, current + delta);
			
			// Always show first page
			if (start > 1) {
				pages.push(1);
				if (start > 2) pages.push('...');
			}
			
			// Show pages around current
			for (let i = start; i <= end; i++) {
				pages.push(i);
			}
			
			// Always show last page
			if (end < total) {
				if (end < total - 1) pages.push('...');
				pages.push(total);
			}
			
			this.paginationPages = pages;
		}
	}`, pageLength)
}

// formatCellValue formats a cell value based on its type and formatter
// NOTE: This function needs integration with @internal/shared/format/
func formatCellValue(value interface{}, valueType string, formatter string) string {
	if value == nil {
		return ""
	}

	// Convert to string for basic processing
	strValue := fmt.Sprintf("%v", value)
	strValue = strings.TrimSpace(strValue)

	// TODO: Implement proper formatting using @internal/shared/format/
	// FIXME: Add support for custom formatters and type-specific formatting
	switch valueType {
	case "number":
		// TODO: Use @internal/shared/format/ for number formatting
		return strValue
	case "date":
		// TODO: Use @internal/shared/format/ for date formatting
		return strValue
	case "datetime":
		// TODO: Use @internal/shared/format/ for datetime formatting
		return strValue
	case "currency":
		// TODO: Use @internal/shared/format/ for currency formatting
		return strValue
	case "percentage":
		// TODO: Use @internal/shared/format/ for percentage formatting
		return strValue
	default:
		return strValue
	}
}

// buildDataTableConfig builds a default configuration
func buildDataTableConfig(config DataTableConfig) DataTableConfig {
	// Set defaults
	if config.PageLength == 0 {
		config.PageLength = 25
	}

	return config
}

// buildSimpleDatatablesOptions builds options for simple-datatables library
// TODO: Implement integration with simple-datatables
// FIXME: Map our config to simple-datatables options format
func buildSimpleDatatablesOptions(props DataTableProps) map[string]interface{} {
	options := map[string]interface{}{
		"searchable":    props.Config.Searchable,
		"sortable":      props.Config.Sortable,
		"paging":        props.Config.Paging,
		"perPage":       props.Config.PageLength,
		"info":          props.Config.Info,
		"perPageSelect": props.Config.LengthChange,
		"fixedHeight":   props.Config.FixedHeader,
	}

	if props.Config.ScrollY != "" {
		options["scrollY"] = props.Config.ScrollY
	}

	// TODO: Add server-side processing options
	if props.Config.ServerSide {
		// FIXME: Implement server-side processing configuration
	}

	return options
}
