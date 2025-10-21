package table

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/niiniyare/erp/internal/shared/format"
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
		filters: {},
		activeFilters: {},
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
				// Select all visible rows on current page
				const startIndex = (this.pageInfo.current - 1) * this.pageInfo.pageSize;
				const endIndex = Math.min(startIndex + this.pageInfo.pageSize, this.pageInfo.total || 0);
				const visibleRows = [];
				
				for (let i = startIndex; i < endIndex; i++) {
					visibleRows.push(i);
				}
				
				this.selectedRows = visibleRows;
			}
		},
		
		clearSelection() {
			this.selectedRows = [];
		},
		
		// Filter methods
		toggleFilters() {
			this.showFilters = !this.showFilters;
		},
		
		setFilter(column, value) {
			if (value === '' || value === null || value === undefined) {
				delete this.activeFilters[column];
			} else {
				this.activeFilters[column] = value;
			}
			
			// Reset to first page when filters change
			this.pageInfo.current = 1;
			
			// Trigger server-side filtering
			this.refreshTable();
		},
		
		clearFilter(column) {
			delete this.activeFilters[column];
			this.pageInfo.current = 1;
			this.refreshTable();
		},
		
		clearAllFilters() {
			this.activeFilters = {};
			this.pageInfo.current = 1;
			this.refreshTable();
		},
		
		hasActiveFilters() {
			return Object.keys(this.activeFilters).length > 0;
		},
		
		// Sorting methods
		sortBy(column) {
			if (this.sortColumn === column) {
				this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
			} else {
				this.sortColumn = column;
				this.sortDirection = 'asc';
			}
			
			// Reset to first page when sorting changes
			this.pageInfo.current = 1;
			
			// Trigger server-side sorting via HTMX
			const tableElement = this.$el.closest('[data-table-id]');
			if (tableElement) {
				const endpoint = tableElement.dataset.endpoint || '';
				const params = this.buildQueryParams();
				
				// Use HTMX to send sorting request
				htmx.ajax('GET', endpoint, {
					target: tableElement.querySelector('.table-body'),
					values: params,
					headers: {
						'HX-Request': 'true',
						'Content-Type': 'application/json'
					}
				});
			}
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
			const tableElement = this.$el.closest('[data-table-id]');
			if (!tableElement) return;
			
			const endpoint = tableElement.dataset.endpoint || '';
			const params = this.buildQueryParams();
			
			// Show loading state
			const tbody = tableElement.querySelector('.table-body');
			if (tbody) {
				tbody.classList.add('opacity-50', 'pointer-events-none');
			}
			
			// Trigger HTMX request with current state
			htmx.ajax('GET', endpoint, {
				target: tbody,
				values: params,
				headers: {
					'HX-Request': 'true',
					'Content-Type': 'application/json'
				}
			}).then(() => {
				// Remove loading state
				if (tbody) {
					tbody.classList.remove('opacity-50', 'pointer-events-none');
				}
			});
		},
		
		buildQueryParams() {
			const params = {
				page: this.pageInfo.current,
				pageSize: this.pageInfo.pageSize,
				sortColumn: this.sortColumn,
				sortDirection: this.sortDirection
			};
			
			// Add filter parameters
			if (Object.keys(this.activeFilters).length > 0) {
				params.filters = this.activeFilters;
				
				// Also add individual filter parameters for backward compatibility
				Object.keys(this.activeFilters).forEach(column => {
					params['filter_' + column] = this.activeFilters[column];
				});
			}
			
			return params;
		},
		
		// Bulk action methods
		handleBulkAction(actionName) {
			if (this.selectedRows.length === 0) {
				// Show notification that no rows are selected
				this.showNotification('Please select at least one row to perform this action.', 'warning');
				return;
			}
			
			const tableElement = this.$el.closest('[data-table-id]');
			if (!tableElement) return;
			
			const bulkEndpoint = tableElement.dataset.bulkEndpoint || '';
			if (!bulkEndpoint) {
				this.showNotification('Bulk actions not configured for this table.', 'error');
				return;
			}
			
			// Confirm destructive actions
			if (['delete', 'archive', 'remove'].includes(actionName.toLowerCase())) {
				if (!confirm("Are you sure you want to " + actionName + " " + this.selectedRows.length + " selected item(s)?")) {
					return;
				}
			}
			
			// Send selected row IDs to server
			const payload = {
				action: actionName,
				rowIds: this.selectedRows,
				selectedCount: this.selectedRows.length
			};
			
			htmx.ajax('POST', bulkEndpoint, {
				values: payload,
				headers: {
					'HX-Request': 'true',
					'Content-Type': 'application/json'
				}
			}).then(() => {
				// Clear selection and refresh table
				this.clearSelection();
				this.refreshTable();
				this.showNotification("Successfully performed " + actionName + " on " + payload.selectedCount + " item(s).", 'success');
			}).catch(() => {
				this.showNotification("Failed to perform " + actionName + ". Please try again.", 'error');
			});
		},
		
		// Notification helper
		showNotification(message, type = 'info') {
			// Dispatch custom event for notification system
			this.$dispatch('show-notification', {
				message: message,
				type: type,
				duration: type === 'error' ? 5000 : 3000
			});
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
func formatCellValue(value any, valueType string, formatter string) string {
	if value == nil {
		return ""
	}

	// Convert to string for basic processing
	strValue := fmt.Sprintf("%v", value)
	strValue = strings.TrimSpace(strValue)

	switch valueType {
	case "number":
		return formatNumber(strValue, formatter)
	case "date":
		return formatDate(strValue, formatter)
	case "datetime":
		return formatDateTime(strValue, formatter)
	case "currency":
		return formatCurrency(strValue, formatter)
	case "percentage":
		return formatPercentage(strValue, formatter)
	case "duration":
		return formatDuration(strValue)
	case "boolean":
		return formatBoolean(strValue)
	default:
		return strValue
	}
}

// formatNumber formats numeric values with proper number formatting
func formatNumber(value, formatter string) string {
	if value == "" {
		return ""
	}

	// Parse numeric value
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value // Return original if parsing fails
	}

	// Apply specific formatter or use default
	switch formatter {
	case "integer":
		return fmt.Sprintf("%.0f", num)
	case "decimal2":
		return fmt.Sprintf("%.2f", num)
	case "decimal4":
		return fmt.Sprintf("%.4f", num)
	case "comma":
		return formatNumberWithCommas(num, 0)
	case "comma-decimal":
		return formatNumberWithCommas(num, 2)
	default:
		// Auto-detect decimal places
		if num == float64(int64(num)) {
			return fmt.Sprintf("%.0f", num)
		}
		return fmt.Sprintf("%.2f", num)
	}
}

// formatNumberWithCommas formats numbers with thousand separators
func formatNumberWithCommas(num float64, decimals int) string {
	// Format with specified decimal places
	formatted := fmt.Sprintf("%."+fmt.Sprintf("%d", decimals)+"f", num)

	// Split integer and decimal parts
	parts := strings.Split(formatted, ".")
	intPart := parts[0]

	// Add commas to integer part
	if len(intPart) > 3 {
		var result strings.Builder
		for i, digit := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(digit)
		}
		intPart = result.String()
	}

	// Reconstruct number
	if decimals > 0 && len(parts) > 1 {
		return intPart + "." + parts[1]
	}
	return intPart
}

// formatDate formats date values using the shared format package
func formatDate(value, formatter string) string {
	if value == "" {
		return ""
	}

	// Try to parse the date using the format package
	formatted, err := format.FormatTimestamp(value, "", format.ISO8601Date, "Local")
	if err != nil {
		return value // Return original if parsing fails
	}

	// Apply specific formatter
	switch formatter {
	case "short":
		if t, err := time.Parse(format.ISO8601Date, formatted); err == nil {
			return t.Format("01/02/06")
		}
	case "long":
		if t, err := time.Parse(format.ISO8601Date, formatted); err == nil {
			return t.Format("January 2, 2006")
		}
	case "iso":
		return formatted
	case "relative":
		if t, err := time.Parse(format.ISO8601Date, formatted); err == nil {
			return format.TimeAgo(t)
		}
	}

	return formatted
}

// formatDateTime formats datetime values using the shared format package
func formatDateTime(value, formatter string) string {
	if value == "" {
		return ""
	}

	// Try to parse the datetime using the format package
	formatted, err := format.FormatTimestamp(value, "", format.DateTime, "Local")
	if err != nil {
		return value // Return original if parsing fails
	}

	// Apply specific formatter
	switch formatter {
	case "short":
		if t, err := time.Parse(format.DateTime, formatted); err == nil {
			return t.Format("01/02/06 3:04 PM")
		}
	case "long":
		if t, err := time.Parse(format.DateTime, formatted); err == nil {
			return t.Format("January 2, 2006 at 3:04 PM")
		}
	case "iso":
		return formatted
	case "relative":
		if t, err := time.Parse(format.DateTime, formatted); err == nil {
			return format.TimeAgo(t)
		}
	case "time-only":
		if t, err := time.Parse(format.DateTime, formatted); err == nil {
			return t.Format("3:04 PM")
		}
	}

	return formatted
}

// formatCurrency formats currency values with proper currency symbols
func formatCurrency(value, formatter string) string {
	if value == "" {
		return ""
	}

	// Parse numeric value
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value // Return original if parsing fails
	}

	// Apply specific currency formatter
	switch formatter {
	case "usd", "dollar":
		return "$" + formatNumberWithCommas(num, 2)
	case "eur", "euro":
		return "€" + formatNumberWithCommas(num, 2)
	case "gbp", "pound":
		return "£" + formatNumberWithCommas(num, 2)
	case "jpy", "yen":
		return "¥" + formatNumberWithCommas(num, 0)
	case "compact":
		return formatCompactCurrency(num)
	default:
		// Default to USD format
		return "$" + formatNumberWithCommas(num, 2)
	}
}

// formatCompactCurrency formats large currency values in compact form (K, M, B)
func formatCompactCurrency(num float64) string {
	abs := num
	if abs < 0 {
		abs = -abs
	}

	var formatted string
	switch {
	case abs >= 1e9:
		formatted = fmt.Sprintf("%.1fB", num/1e9)
	case abs >= 1e6:
		formatted = fmt.Sprintf("%.1fM", num/1e6)
	case abs >= 1e3:
		formatted = fmt.Sprintf("%.1fK", num/1e3)
	default:
		formatted = fmt.Sprintf("%.2f", num)
	}

	return "$" + formatted
}

// formatPercentage formats percentage values
func formatPercentage(value, formatter string) string {
	if value == "" {
		return ""
	}

	// Parse numeric value
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value // Return original if parsing fails
	}

	// Apply specific percentage formatter
	switch formatter {
	case "decimal":
		// Assume value is already in decimal form (0.25 = 25%)
		return fmt.Sprintf("%.1f%%", num*100)
	case "integer":
		// Assume value is already percentage (25 = 25%)
		return fmt.Sprintf("%.0f%%", num)
	default:
		// Auto-detect format
		if num <= 1.0 {
			return fmt.Sprintf("%.1f%%", num*100)
		}
		return fmt.Sprintf("%.1f%%", num)
	}
}

// formatDuration formats duration values using the shared format package
func formatDuration(value string) string {
	if value == "" {
		return ""
	}

	// Try to parse as Go duration
	if duration, err := format.ParseDuration(value); err == nil {
		return format.FormatDuration(duration)
	}

	// Try to parse as seconds
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		duration := time.Duration(seconds) * time.Second
		return format.FormatDuration(duration)
	}

	return value
}

// formatBoolean formats boolean values with user-friendly text
func formatBoolean(value string) string {
	switch strings.ToLower(value) {
	case "true", "1", "yes", "y":
		return "Yes"
	case "false", "0", "no", "n":
		return "No"
	default:
		return value
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
func buildSimpleDatatablesOptions(props DataTableProps) map[string]any {
	options := map[string]any{
		"searchable":    props.Config.Searchable,
		"sortable":      props.Config.Sortable,
		"paging":        props.Config.Paging,
		"perPage":       props.Config.PageLength,
		"info":          props.Config.Info,
		"perPageSelect": props.Config.LengthChange,
		"fixedHeight":   props.Config.FixedHeader,
		"classes": map[string]string{
			"active":     "bg-blue-100 dark:bg-blue-900",
			"disabled":   "opacity-50 cursor-not-allowed",
			"selector":   "w-4 h-4 text-blue-600",
			"loading":    "opacity-50",
			"empty":      "text-center text-gray-500 dark:text-gray-400",
			"info":       "text-sm text-gray-600 dark:text-gray-300",
			"pagination": "flex items-center justify-between",
		},
		"labels": map[string]string{
			"placeholder":  "Search...",
			"perPage":      "Rows per page:",
			"noRows":       "No data available",
			"info":         "Showing {start} to {end} of {rows} entries",
			"infoEmpty":    "Showing 0 to 0 of 0 entries",
			"infoFiltered": "filtered from {max} total entries",
		},
	}

	// Configure scrolling
	if props.Config.ScrollY != "" {
		options["scrollY"] = props.Config.ScrollY
		options["scrollCollapse"] = true
	}

	// Configure column options
	if len(props.Columns) > 0 {
		columns := make([]map[string]any, 0, len(props.Columns))
		for i, col := range props.Columns {
			columnOpts := map[string]any{
				"select":   i,
				"sortable": col.Sortable,
				"type":     mapColumnTypeToSimpleDT(col.Type),
			}

			// Add formatting function if needed
			if col.Type != "text" {
				columnOpts["render"] = generateFormatterJS(col.Type, col.Format)
			}

			// Add column width
			if col.Width != "" {
				columnOpts["width"] = col.Width
			}

			columns = append(columns, columnOpts)
		}
		options["columns"] = columns
	}

	// Server-side processing configuration
	if props.Config.ServerSide {
		options["serverSide"] = true
		options["ajax"] = map[string]any{
			"url":         props.Config.AjaxURL,
			"type":        "POST",
			"data":        "function(d) { return JSON.stringify(d); }",
			"contentType": "application/json; charset=utf-8",
			"dataType":    "json",
		}

		// Disable client-side features when using server-side processing
		options["searching"] = false
		options["ordering"] = false
		options["paging"] = false
	}

	// Page length options
	if props.Config.LengthChange {
		options["perPageSelect"] = []int{10, 25, 50, 100}
		if props.Config.PageLength > 0 {
			found := false
			for _, opt := range options["perPageSelect"].([]int) {
				if opt == props.Config.PageLength {
					found = true
					break
				}
			}
			if !found {
				options["perPageSelect"] = append(options["perPageSelect"].([]int), props.Config.PageLength)
			}
		}
	}

	// Row selection
	if props.Config.RowSelection {
		options["select"] = map[string]any{
			"style":    "multi",
			"selector": "td:first-child",
		}
	}

	return options
}

// mapColumnTypeToSimpleDT maps our column types to simple-datatables types
func mapColumnTypeToSimpleDT(colType string) string {
	switch colType {
	case "number", "currency", "percentage":
		return "number"
	case "date", "datetime":
		return "date"
	case "boolean":
		return "string" // Handle as string with custom rendering
	default:
		return "string"
	}
}

// generateFormatterJS generates JavaScript formatter function for columns
func generateFormatterJS(colType, format string) string {
	switch colType {
	case "currency":
		currency := "USD"
		if format != "" {
			currency = strings.ToUpper(format)
		}
		return fmt.Sprintf(`function(data, type, row) {
			if (type === 'display' || type === 'type') {
				return new Intl.NumberFormat('en-US', {
					style: 'currency',
					currency: '%s'
				}).format(data);
			}
			return data;
		}`, currency)

	case "percentage":
		return `function(data, type, row) {
			if (type === 'display' || type === 'type') {
				return (parseFloat(data) * 100).toFixed(1) + '%';
			}
			return data;
		}`

	case "date":
		dateFormat := "short"
		if format != "" {
			dateFormat = format
		}
		return fmt.Sprintf(`function(data, type, row) {
			if (type === 'display' || type === 'type') {
				const date = new Date(data);
				if (isNaN(date.getTime())) return data;
				
				switch('%s') {
					case 'short':
						return date.toLocaleDateString();
					case 'long':
						return date.toLocaleDateString('en-US', {
							year: 'numeric',
							month: 'long',
							day: 'numeric'
						});
					default:
						return date.toLocaleDateString();
				}
			}
			return data;
		}`, dateFormat)

	case "datetime":
		return `function(data, type, row) {
			if (type === 'display' || type === 'type') {
				const date = new Date(data);
				if (isNaN(date.getTime())) return data;
				return date.toLocaleString();
			}
			return data;
		}`

	case "boolean":
		return `function(data, type, row) {
			if (type === 'display' || type === 'type') {
				return data === true || data === 'true' || data === '1' ? 'Yes' : 'No';
			}
			return data;
		}`

	default:
		return ""
	}
}
