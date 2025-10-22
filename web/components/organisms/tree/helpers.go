package tree

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/niiniyare/erp/web/components/atoms"
)

// ============================================================================
// TREE STYLING HELPER FUNCTIONS
// ============================================================================

// getTreeClasses generates CSS classes for the tree container
func getTreeClasses(props TreeProps) string {
	classes := []string{
		"tree-container",
		"relative",
	}

	// Add variant classes
	switch props.Styling.Variant {
	case atoms.VariantDefault:
		classes = append(classes, "bg-white", "dark:bg-gray-900")
	case atoms.VariantLight:
		classes = append(classes, "bg-gray-50", "dark:bg-gray-800")
	case atoms.VariantDark:
		classes = append(classes, "bg-gray-800", "text-white")
	}

	// Add size classes
	switch props.Styling.Size {
	case atoms.SizeXS:
		classes = append(classes, "text-xs")
	case atoms.SizeSM:
		classes = append(classes, "text-sm")
	case atoms.SizeMD:
		classes = append(classes, "text-base")
	case atoms.SizeLG:
		classes = append(classes, "text-lg")
	case atoms.SizeXL:
		classes = append(classes, "text-xl")
	}

	// Add conditional styling
	if props.Styling.Rounded {
		classes = append(classes, "rounded-lg")
	}
	if props.Styling.Shadow {
		classes = append(classes, "shadow-md")
	}
	if props.Styling.Bordered {
		classes = append(classes, "border", "border-gray-200", "dark:border-gray-700")
	}

	// Add max height if specified
	if props.Config.MaxHeight != "" {
		classes = append(classes, "overflow-auto")
	}

	// Add custom class if specified
	if props.Class != "" {
		classes = append(classes, props.Class)
	}

	return strings.Join(classes, " ")
}

// getTreeNodeClasses generates CSS classes for individual tree nodes
func getTreeNodeClasses(node TreeNode, props TreeProps, level int) string {
	classes := []string{
		"tree-node",
		"relative",
		"group",
	}

	// For multi-column view, use different layout
	if len(props.Columns) > 0 {
		classes = append(classes, "grid")
	} else {
		classes = append(classes, "flex", "items-center")
	}

	// Add hover effects if enabled
	if props.Styling.Hover {
		classes = append(classes, "hover:bg-gray-50", "dark:hover:bg-gray-800")
	}

	// Add striped background if enabled
	if props.Styling.Striped && level%2 == 1 {
		classes = append(classes, "bg-gray-25", "dark:bg-gray-850")
	}

	// Add selection state classes
	if node.Selected {
		classes = append(classes, "bg-blue-50", "dark:bg-blue-900/20", "border-l-2", "border-blue-500")
	}

	// Add disabled state classes
	if node.Disabled {
		classes = append(classes, "opacity-50", "cursor-not-allowed")
	} else {
		classes = append(classes, "cursor-pointer")
	}

	// Add height based on configuration
	nodeHeight := props.Config.NodeHeight
	if nodeHeight == "" {
		nodeHeight = "2.5rem"
	}

	switch nodeHeight {
	case "2rem":
		classes = append(classes, "h-8")
	case "2.5rem":
		classes = append(classes, "h-10")
	case "3rem":
		classes = append(classes, "h-12")
	default:
		classes = append(classes, "h-10")
	}

	// Add parent node styling if has aggregation
	if node.HasChildren && hasAggregatedColumns(props.Columns) {
		classes = append(classes, "font-semibold", "bg-gray-50/50", "dark:bg-gray-800/50")
	}

	return strings.Join(classes, " ")
}

// hasAggregatedColumns checks if any column has aggregation enabled
func hasAggregatedColumns(columns []TreeColumn) bool {
	for _, col := range columns {
		if col.Aggregation != nil && col.Aggregation.Enabled {
			return true
		}
	}
	return false
}

// getTreeNodeIndentStyle generates inline style for node indentation
func getTreeNodeIndentStyle(level int, indent string) string {
	if indent == "" {
		indent = "1.5rem"
	}

	// Calculate padding based on level and indent
	return fmt.Sprintf("padding-left: calc(%d * %s)", level, indent)
}

// getExpandIconClasses generates CSS classes for expand/collapse icons
func getExpandIconClasses(node TreeNode, expanded bool) string {
	classes := []string{
		"tree-expand-icon",
		"flex-shrink-0",
		"w-4",
		"h-4",
		"mr-2",
		"text-gray-400",
		"transition-transform",
		"duration-200",
		"cursor-pointer",
	}

	// Add rotation for expanded state
	if expanded {
		classes = append(classes, "transform", "rotate-90")
	}

	// Hide icon if node has no children
	if !node.HasChildren {
		classes = append(classes, "invisible")
	} else {
		classes = append(classes, "hover:text-gray-600", "dark:hover:text-gray-300")
	}

	return strings.Join(classes, " ")
}

// getNodeIconClasses generates CSS classes for node icons
func getNodeIconClasses(node TreeNode) string {
	classes := []string{
		"tree-node-icon",
		"flex-shrink-0",
		"w-4",
		"h-4",
		"mr-2",
	}

	// Add icon color based on node type
	if node.IsLeaf {
		classes = append(classes, "text-blue-500")
	} else {
		classes = append(classes, "text-yellow-500")
	}

	return strings.Join(classes, " ")
}

// getCheckboxClasses generates CSS classes for selection checkboxes
func getCheckboxClasses() string {
	return strings.Join([]string{
		"tree-checkbox",
		"flex-shrink-0",
		"mr-2",
		"w-4",
		"h-4",
		"rounded",
		"border",
		"border-gray-300",
		"dark:border-gray-600",
		"focus:ring-blue-500",
		"cursor-pointer",
	}, " ")
}

// getNodeLabelClasses generates CSS classes for node labels
func getNodeLabelClasses(node TreeNode) string {
	classes := []string{
		"tree-node-label",
		"flex-1",
		"truncate",
		"select-none",
	}

	// Add text color based on state
	if node.Disabled {
		classes = append(classes, "text-gray-400")
	} else {
		classes = append(classes, "text-gray-900", "dark:text-white")
	}

	// Add font weight for selected nodes
	if node.Selected {
		classes = append(classes, "font-medium")
	}

	return strings.Join(classes, " ")
}

// getNodeActionsClasses generates CSS classes for node action buttons
func getNodeActionsClasses(visibleOn string) string {
	classes := []string{
		"tree-node-actions",
		"flex",
		"items-center",
		"gap-1",
		"ml-2",
	}

	// Add visibility classes based on visibleOn setting
	switch visibleOn {
	case "hover":
		classes = append(classes, "opacity-0", "group-hover:opacity-100", "transition-opacity")
	case "selected":
		classes = append(classes, "opacity-0", "group-[.selected]:opacity-100")
	case "always":
		// Always visible, no additional classes needed
	default:
		// Default to hover
		classes = append(classes, "opacity-0", "group-hover:opacity-100", "transition-opacity")
	}

	return strings.Join(classes, " ")
}

// getBadgeClasses generates CSS classes for node badges
func getBadgeClasses() string {
	return strings.Join([]string{
		"tree-node-badge",
		"inline-flex",
		"items-center",
		"px-2",
		"py-0.5",
		"ml-2",
		"text-xs",
		"font-medium",
		"bg-gray-100",
		"text-gray-800",
		"rounded-full",
		"dark:bg-gray-700",
		"dark:text-gray-300",
	}, " ")
}

// getAggregationBadgeClasses generates classes for aggregated value indicators
func getAggregationBadgeClasses() string {
	return strings.Join([]string{
		"inline-flex",
		"items-center",
		"px-2",
		"py-0.5",
		"ml-2",
		"text-xs",
		"font-semibold",
		"bg-blue-100",
		"text-blue-800",
		"rounded",
		"dark:bg-blue-900",
		"dark:text-blue-200",
	}, " ")
}

// ============================================================================
// COLUMN STYLING HELPERS
// ============================================================================

// getColumnGridStyle generates the CSS grid template for columns
func getColumnGridStyle(columns []TreeColumn) string {
	if len(columns) == 0 {
		return ""
	}

	widths := make([]string, len(columns))
	for i, col := range columns {
		if col.Width != "" {
			widths[i] = col.Width
		} else {
			widths[i] = "1fr"
		}
	}

	return fmt.Sprintf("grid-template-columns: %s", strings.Join(widths, " "))
}

// getColumnCellClasses generates CSS classes for column cells
func getColumnCellClasses(column TreeColumn, isFirst bool) string {
	classes := []string{
		"tree-cell",
		"px-3",
		"py-2",
		"truncate",
	}

	// Add alignment
	switch column.Align {
	case "center":
		classes = append(classes, "text-center")
	case "right":
		classes = append(classes, "text-right")
	default:
		classes = append(classes, "text-left")
	}

	// First column needs flex for tree structure
	if isFirst {
		classes = append(classes, "flex", "items-center")
	}

	// Add column type specific classes
	switch column.Type {
	case "currency", "number":
		classes = append(classes, "font-mono", "tabular-nums")
	}

	// Add custom class if specified
	if column.ClassName != "" {
		classes = append(classes, column.ClassName)
	}

	return strings.Join(classes, " ")
}

// getColumnHeaderClasses generates CSS classes for column headers
func getColumnHeaderClasses(column TreeColumn) string {
	classes := []string{
		"tree-header-cell",
		"px-3",
		"py-2",
		"font-medium",
		"text-sm",
		"text-gray-700",
		"dark:text-gray-300",
		"border-b",
		"border-gray-200",
		"dark:border-gray-700",
		"bg-gray-50",
		"dark:bg-gray-800",
	}

	// Add alignment
	switch column.Align {
	case "center":
		classes = append(classes, "text-center")
	case "right":
		classes = append(classes, "text-right")
	default:
		classes = append(classes, "text-left")
	}

	// Add sortable indicator
	if column.Sortable {
		classes = append(classes, "cursor-pointer", "hover:bg-gray-100", "dark:hover:bg-gray-700")
	}

	// Add custom header class if specified
	if column.HeaderClassName != "" {
		classes = append(classes, column.HeaderClassName)
	}

	return strings.Join(classes, " ")
}

// ============================================================================
// VALUE FORMATTING HELPERS
// ============================================================================

// formatColumnValue formats a value based on column type and format
func formatColumnValue(value interface{}, column TreeColumn) string {
	if value == nil {
		return "-"
	}

	switch column.Type {
	case "currency":
		if v, ok := toFloat64(value); ok {
			if column.Format != "" {
				return formatCurrency(v, column.Format)
			}
			return fmt.Sprintf("$%.2f", v)
		}
	case "number":
		if v, ok := toFloat64(value); ok {
			if column.Format != "" {
				return formatNumber(v, column.Format)
			}
			return fmt.Sprintf("%.2f", v)
		}
	case "percentage":
		if v, ok := toFloat64(value); ok {
			return fmt.Sprintf("%.1f%%", v*100)
		}
	case "date":
		// Handle date formatting
		return fmt.Sprintf("%v", value)
	}

	return fmt.Sprintf("%v", value)
}

// Helper to convert interface{} to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// formatCurrency formats currency with thousand separators
func formatCurrency(value float64, format string) string {
	// Simple implementation - could be enhanced with format parsing
	return fmt.Sprintf("$%s", formatNumberWithCommas(value))
}

// formatNumber formats numbers with thousand separators
func formatNumber(value float64, format string) string {
	return formatNumberWithCommas(value)
}

// formatNumberWithCommas adds thousand separators to numbers
func formatNumberWithCommas(value float64) string {
	// Handle negative numbers
	negative := value < 0
	if negative {
		value = -value
	}

	parts := strings.Split(fmt.Sprintf("%.2f", value), ".")
	intPart := parts[0]
	decPart := parts[1]

	// Add commas to integer part
	var result strings.Builder
	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}

	formatted := result.String() + "." + decPart
	if negative {
		return "-" + formatted
	}
	return formatted
}

// ============================================================================
// TREE STATE MANAGEMENT HELPERS
// ============================================================================

// getAlpineTreeData generates the Alpine.js data object for tree state management
func getAlpineTreeData(props TreeProps) string {
	alpineData := map[string]interface{}{
		"expanded":   getInitialExpandedState(props),
		"selected":   getInitialSelectedState(props),
		"searchTerm": "",
		"showSearch": props.Config.Searchable,
		"loading":    props.Loading.Show,
		"nodes":      props.Data,
		"columns":    props.Columns,
	}

	return buildAlpineObject(alpineData)
}

// buildAlpineObject properly builds Alpine.js object with methods
func buildAlpineObject(data map[string]interface{}) string {
	var parts []string

	// Add data properties
	for key, value := range data {
		switch v := value.(type) {
		case bool:
			parts = append(parts, fmt.Sprintf("%s: %t", key, v))
		case string:
			parts = append(parts, fmt.Sprintf("%s: '%s'", key, escapeString(v)))
		case []string:
			parts = append(parts, fmt.Sprintf("%s: %s", key, formatStringArray(v)))
		default:
			if jsonStr := toJSON(value); jsonStr != "null" {
				parts = append(parts, fmt.Sprintf("%s: %s", key, jsonStr))
			} else {
				parts = append(parts, fmt.Sprintf("%s: null", key))
			}
		}
	}

	// Add Alpine.js methods for tree operations
	methods := []string{
		// Toggle node expansion
		`toggleNode(nodeId) {
			const idx = this.expanded.indexOf(nodeId);
			if (idx > -1) {
				this.expanded.splice(idx, 1);
			} else {
				this.expanded.push(nodeId);
			}
		}`,

		// Check if node is expanded
		`isExpanded(nodeId) {
			return this.expanded.includes(nodeId);
		}`,

		// Toggle node selection
		`toggleSelection(nodeId, checked) {
			if (checked === undefined) {
				const idx = this.selected.indexOf(nodeId);
				if (idx > -1) {
					this.selected.splice(idx, 1);
				} else {
					this.selected.push(nodeId);
				}
			} else if (checked) {
				if (!this.selected.includes(nodeId)) {
					this.selected.push(nodeId);
				}
			} else {
				const idx = this.selected.indexOf(nodeId);
				if (idx > -1) {
					this.selected.splice(idx, 1);
				}
			}
		}`,

		// Check if node is selected
		`isSelected(nodeId) {
			return this.selected.includes(nodeId);
		}`,

		// Find node by ID
		`findNode(nodeId, nodes = this.nodes) {
			for (let node of nodes) {
				if (node.id === nodeId) return node;
				if (node.children && node.children.length > 0) {
					const found = this.findNode(nodeId, node.children);
					if (found) return found;
				}
			}
			return null;
		}`,

		// Get column value
		`getColumnValue(node, field) {
			if (!node.data) return '';
			return node.data[field] ?? '';
		}`,

		// Get aggregated value for a column
		`getAggregatedValue(node, column) {
			if (!column.aggregation || !column.aggregation.enabled || !node.hasChildren) {
				return this.getColumnValue(node, column.field);
			}
			return this.calculateAggregation(node, column);
		}`,

		// Calculate aggregation
		`calculateAggregation(node, column) {
			const method = column.aggregation.method;
			const excludeParent = column.aggregation.excludeParentValue;
			let values = [];
			
			if (!excludeParent) {
				const val = this.getColumnValue(node, column.field);
				if (val !== null && val !== '' && !isNaN(val)) {
					values.push(parseFloat(val));
				}
			}
			
			this.collectChildValues(node, column.field, values);
			
			if (values.length === 0) return 0;
			
			switch(method) {
				case 'sum': return values.reduce((a, b) => a + b, 0);
				case 'average': return values.reduce((a, b) => a + b, 0) / values.length;
				case 'count': return values.length;
				case 'min': return Math.min(...values);
				case 'max': return Math.max(...values);
				default: return values.reduce((a, b) => a + b, 0);
			}
		}`,

		// Collect child values recursively
		`collectChildValues(node, field, values) {
			if (!node.children || node.children.length === 0) return;
			node.children.forEach(child => {
				if (child.hasChildren) {
					this.collectChildValues(child, field, values);
				} else {
					const val = this.getColumnValue(child, field);
					if (val !== null && val !== '' && !isNaN(val)) {
						values.push(parseFloat(val));
					}
				}
			});
		}`,

		// Format value based on column type
		`formatValue(value, column) {
			if (value === null || value === undefined || value === '') return '-';
			
			switch(column.type) {
				case 'currency':
					return new Intl.NumberFormat('en-US', {
						style: 'currency',
						currency: 'USD'
					}).format(value);
				case 'number':
					return new Intl.NumberFormat('en-US').format(value);
				case 'percentage':
					return (value * 100).toFixed(1) + '%';
				case 'date':
					return new Date(value).toLocaleDateString();
				default:
					return value.toString();
			}
		}`,

		// Filter nodes based on search term
		`filterNodes() {
			if (!this.searchTerm || this.searchTerm.length === 0) {
				return;
			}
			// Search implementation would go here
		}`,
	}

	// Combine properties and methods
	allParts := append(parts, methods...)

	return fmt.Sprintf("{\n    %s\n}", strings.Join(allParts, ",\n    "))
}

// getInitialExpandedState returns the initial expanded state for nodes
func getInitialExpandedState(props TreeProps) []string {
	var expanded []string

	if props.Config.ExpandAll {
		// Expand all nodes that have children
		expanded = getAllExpandableNodeIDs(props.Data)
	} else {
		// Use default expanded nodes
		expanded = props.Config.DefaultExpanded
	}

	return expanded
}

// getInitialSelectedState returns the initial selected state for nodes
func getInitialSelectedState(props TreeProps) []string {
	return props.Config.DefaultSelected
}

// getAllExpandableNodeIDs recursively collects all node IDs that have children
func getAllExpandableNodeIDs(nodes []TreeNode) []string {
	var ids []string

	for _, node := range nodes {
		if node.HasChildren {
			ids = append(ids, node.ID)
			// Recursively get child node IDs
			ids = append(ids, getAllExpandableNodeIDs(node.Children)...)
		}
	}

	return ids
}

// formatStringArray formats a string array for Alpine.js
func formatStringArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}

	quoted := make([]string, len(arr))
	for i, s := range arr {
		quoted[i] = fmt.Sprintf("'%s'", escapeString(s))
	}

	return fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
}

// escapeString escapes special characters for JavaScript strings
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// toJSON converts a value to JSON string
func toJSON(v interface{}) string {
	if v == nil {
		return "null"
	}
	if data, err := json.Marshal(v); err == nil {
		return string(data)
	}
	return "null"
}

// ============================================================================
// ACCESSIBILITY HELPERS
// ============================================================================

// getTreeAccessibilityAttrs generates accessibility attributes for the tree
func getTreeAccessibilityAttrs(props TreeProps) map[string]string {
	attrs := map[string]string{
		"role":       "tree",
		"aria-label": getTreeAriaLabel(props),
	}

	if props.AriaDescribedBy != "" {
		attrs["aria-describedby"] = props.AriaDescribedBy
	}

	if props.Config.MultiSelect {
		attrs["aria-multiselectable"] = "true"
	}

	return attrs
}

// getNodeAccessibilityAttrs generates accessibility attributes for individual nodes
func getNodeAccessibilityAttrs(node TreeNode, props TreeProps) map[string]string {
	attrs := map[string]string{
		"role":       "treeitem",
		"aria-label": node.Label,
		"aria-level": fmt.Sprintf("%d", node.Level+1),
	}

	if node.Selected {
		attrs["aria-selected"] = "true"
	}

	if node.HasChildren {
		attrs["aria-expanded"] = fmt.Sprintf("%t", node.Expanded)
	}

	if node.Disabled {
		attrs["aria-disabled"] = "true"
	}

	return attrs
}

// getTreeAriaLabel generates an appropriate ARIA label for the tree
func getTreeAriaLabel(props TreeProps) string {
	if props.AriaLabel != "" {
		return props.AriaLabel
	}

	if props.Title != "" {
		return props.Title
	}

	return "Tree navigation"
}

// ============================================================================
// HTMX INTEGRATION HELPERS
// ============================================================================

// getHTMXAttrs generates HTMX attributes for tree interactions
func getHTMXAttrs(props TreeProps) map[string]string {
	attrs := make(map[string]string)

	if props.HxGet != "" {
		attrs["hx-get"] = props.HxGet
	}
	if props.HxPost != "" {
		attrs["hx-post"] = props.HxPost
	}
	if props.HxTarget != "" {
		attrs["hx-target"] = props.HxTarget
	}
	if props.HxSwap != "" {
		attrs["hx-swap"] = props.HxSwap
	}
	if props.HxTrigger != "" {
		attrs["hx-trigger"] = props.HxTrigger
	}

	return attrs
}

// getLazyLoadAttrs generates HTMX attributes for lazy loading nodes
func getLazyLoadAttrs(node TreeNode, props TreeProps) map[string]string {
	attrs := make(map[string]string)

	if props.Config.LazyLoad && props.LazyLoadUrl != "" && node.HasChildren && !node.Expanded {
		attrs["hx-get"] = fmt.Sprintf("%s?nodeId=%s", props.LazyLoadUrl, node.ID)
		attrs["hx-target"] = fmt.Sprintf("#node-%s-children", node.ID)
		attrs["hx-swap"] = "innerHTML"
		attrs["hx-trigger"] = "click"
	}

	return attrs
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// joinClasses joins CSS class strings, filtering out empty strings
func joinClasses(classes ...string) string {
	var nonEmpty []string
	for _, class := range classes {
		if class != "" {
			nonEmpty = append(nonEmpty, class)
		}
	}
	return strings.Join(nonEmpty, " ")
}

// escapeID ensures the ID is safe for use in HTML and CSS selectors
func escapeID(id string) string {
	// Replace any characters that might cause issues in CSS selectors
	id = strings.ReplaceAll(id, ":", "\\:")
	id = strings.ReplaceAll(id, ".", "\\.")
	return id
}

