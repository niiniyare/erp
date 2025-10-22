package tree

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ============================================================================
// TREE DATA MANIPULATION UTILITIES
// ============================================================================

// FlattenTree converts a hierarchical tree structure to a flat list
func FlattenTree(nodes []TreeNode) []TreeNode {
	var flat []TreeNode

	for _, node := range nodes {
		flat = append(flat, node)
		if len(node.Children) > 0 {
			flat = append(flat, FlattenTree(node.Children)...)
		}
	}

	return flat
}

// BuildTreeFromFlat constructs a hierarchical tree from a flat list of nodes
func BuildTreeFromFlat(nodes []TreeNode) []TreeNode {
	nodeMap := make(map[string]*TreeNode)
	var roots []TreeNode

	// Create a map of all nodes
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
		nodes[i].Children = []TreeNode{} // Initialize children slice
	}

	// Build the tree structure
	for i := range nodes {
		node := &nodes[i]
		if node.ParentID == "" {
			// Root node
			roots = append(roots, *node)
		} else {
			// Child node
			if parent, exists := nodeMap[node.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
				parent.HasChildren = true
				parent.IsLeaf = false
			}
		}
	}

	return roots
}

// SearchTree searches for nodes matching the given term in specified fields
func SearchTree(nodes []TreeNode, searchTerm string, fields []string) []TreeNode {
	if searchTerm == "" {
		return nodes
	}

	var matches []TreeNode
	searchTerm = strings.ToLower(searchTerm)

	for _, node := range nodes {
		if nodeMatches(node, searchTerm, fields) {
			matches = append(matches, node)
		} else if len(node.Children) > 0 {
			// Recursively search children
			childMatches := SearchTree(node.Children, searchTerm, fields)
			if len(childMatches) > 0 {
				// Include parent if children match
				nodeWithMatches := node
				nodeWithMatches.Children = childMatches
				matches = append(matches, nodeWithMatches)
			}
		}
	}

	return matches
}

// nodeMatches checks if a node matches the search term in any of the specified fields
func nodeMatches(node TreeNode, searchTerm string, fields []string) bool {
	for _, field := range fields {
		var value string

		switch field {
		case "label":
			value = node.Label
		case "description":
			value = node.Description
		case "badge":
			value = node.Badge
		default:
			// Check data fields
			if node.Data != nil {
				if val, exists := node.Data[field]; exists {
					if strVal, ok := val.(string); ok {
						value = strVal
					} else {
						value = fmt.Sprintf("%v", val)
					}
				}
			}
			// Check metadata
			if node.Metadata != nil {
				if val, exists := node.Metadata[field]; exists {
					if strVal, ok := val.(string); ok {
						value = strVal
					}
				}
			}
		}

		if strings.Contains(strings.ToLower(value), searchTerm) {
			return true
		}
	}

	return false
}

// FilterTree filters nodes based on a predicate function
func FilterTree(nodes []TreeNode, predicate func(TreeNode) bool) []TreeNode {
	var filtered []TreeNode

	for _, node := range nodes {
		if predicate(node) {
			filtered = append(filtered, node)
		} else if len(node.Children) > 0 {
			// Recursively filter children
			childFiltered := FilterTree(node.Children, predicate)
			if len(childFiltered) > 0 {
				// Include parent if children pass filter
				nodeWithFiltered := node
				nodeWithFiltered.Children = childFiltered
				filtered = append(filtered, nodeWithFiltered)
			}
		}
	}

	return filtered
}

// ExpandPath expands all nodes in the path to a specific node
func ExpandPath(nodes []TreeNode, targetID string) []TreeNode {
	return expandPathRecursive(nodes, targetID, []string{})
}

func expandPathRecursive(nodes []TreeNode, targetID string, path []string) []TreeNode {
	for i, node := range nodes {
		currentPath := append(path, node.ID)

		if node.ID == targetID {
			// Found target, expand all nodes in path
			nodes[i].Expanded = true
			return nodes
		}

		if len(node.Children) > 0 {
			updatedChildren := expandPathRecursive(node.Children, targetID, currentPath)
			nodes[i].Children = updatedChildren

			// Check if target was found in children
			if containsExpandedNode(updatedChildren) {
				nodes[i].Expanded = true
			}
		}
	}

	return nodes
}

// containsExpandedNode checks if any node in the tree is expanded
func containsExpandedNode(nodes []TreeNode) bool {
	for _, node := range nodes {
		if node.Expanded {
			return true
		}
		if len(node.Children) > 0 && containsExpandedNode(node.Children) {
			return true
		}
	}
	return false
}

// GetNodeByID finds a node by its ID in the tree
func GetNodeByID(nodes []TreeNode, id string) *TreeNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if len(nodes[i].Children) > 0 {
			if found := GetNodeByID(nodes[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

// GetNodePath returns the path from root to a specific node
func GetNodePath(nodes []TreeNode, targetID string) []string {
	path := findNodePath(nodes, targetID, []string{})
	return path
}

func findNodePath(nodes []TreeNode, targetID string, currentPath []string) []string {
	for _, node := range nodes {
		nodePath := append(currentPath, node.ID)

		if node.ID == targetID {
			return nodePath
		}

		if len(node.Children) > 0 {
			if path := findNodePath(node.Children, targetID, nodePath); len(path) > 0 {
				return path
			}
		}
	}

	return nil
}

// ============================================================================
// TREE AGGREGATION UTILITIES
// ============================================================================

// CalculateAggregations calculates aggregated values for all parent nodes
func CalculateAggregations(nodes []TreeNode, columns []TreeColumn) []TreeNode {
	var result []TreeNode

	for _, node := range nodes {
		if len(node.Children) > 0 {
			// Recursively calculate children first
			node.Children = CalculateAggregations(node.Children, columns)

			// Calculate aggregations for this node
			for _, column := range columns {
				if column.Aggregation != nil && column.Aggregation.Enabled {
					aggValue := calculateNodeAggregation(node, column)

					// Store aggregated value in metadata
					if node.Metadata == nil {
						node.Metadata = make(map[string]interface{})
					}
					node.Metadata[fmt.Sprintf("%s_aggregated", column.Field)] = aggValue
				}
			}
		}

		result = append(result, node)
	}

	return result
}

// calculateNodeAggregation calculates aggregated value for a single node
func calculateNodeAggregation(node TreeNode, column TreeColumn) interface{} {
	if !node.HasChildren {
		return node.Data[column.Field]
	}

	method := column.Aggregation.Method
	excludeParent := column.Aggregation.ExcludeParentValue
	field := column.Field

	values := []float64{}

	// Include parent's own value if configured
	if !excludeParent {
		if val, ok := toFloat64(node.Data[field]); ok {
			values = append(values, val)
		}
	}

	// Collect child values recursively
	collectNodeValues(node.Children, field, &values)

	if len(values) == 0 {
		return 0.0
	}

	// Apply aggregation method
	switch method {
	case "sum":
		return sumFloats(values)
	case "average":
		return averageFloats(values)
	case "count":
		return len(values)
	case "min":
		return minFloat(values)
	case "max":
		return maxFloat(values)
	default:
		return sumFloats(values)
	}
}

// collectNodeValues recursively collects values from children
func collectNodeValues(nodes []TreeNode, field string, values *[]float64) {
	for _, node := range nodes {
		if node.HasChildren {
			// Recursively collect from children
			collectNodeValues(node.Children, field, values)
		} else {
			// Leaf node - get its value
			if val, ok := toFloat64(node.Data[field]); ok {
				*values = append(*values, val)
			}
		}
	}
}

// Math helper functions
func sumFloats(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}

func averageFloats(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return sumFloats(values) / float64(len(values))
}

func minFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func maxFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// GetAggregatedValue retrieves the aggregated value for a node and column
func GetAggregatedValue(node TreeNode, columnField string) interface{} {
	if node.Metadata == nil {
		return node.Data[columnField]
	}

	aggKey := fmt.Sprintf("%s_aggregated", columnField)
	if aggVal, exists := node.Metadata[aggKey]; exists {
		return aggVal
	}

	return node.Data[columnField]
}

// ============================================================================
// CHART OF ACCOUNTS UTILITIES
// ============================================================================

// AccountNode represents a chart of accounts node with financial data
type AccountNode struct {
	TreeNode
	AccountCode string  `json:"accountCode"`
	AccountType string  `json:"accountType"` // asset, liability, equity, revenue, expense
	Balance     float64 `json:"balance"`
	DebitCredit string  `json:"debitCredit"` // debit, credit
	Currency    string  `json:"currency"`
}

// BuildChartOfAccounts converts AccountNodes to TreeNodes with proper aggregation
func BuildChartOfAccounts(accounts []AccountNode) []TreeNode {
	nodes := make([]TreeNode, len(accounts))

	for i, account := range accounts {
		node := TreeNode{
			ID:          account.AccountCode,
			Label:       account.Label,
			ParentID:    account.ParentID,
			Level:       account.Level,
			HasChildren: account.HasChildren,
			IsLeaf:      account.IsLeaf,
			Data: map[string]interface{}{
				"accountCode": account.AccountCode,
				"accountName": account.Label,
				"accountType": account.AccountType,
				"balance":     account.Balance,
				"debitCredit": account.DebitCredit,
				"currency":    account.Currency,
			},
			Metadata: make(map[string]interface{}),
		}
		nodes[i] = node
	}

	// Build tree structure
	tree := BuildTreeFromFlat(nodes)

	// Calculate aggregated balances
	columns := []TreeColumn{
		{
			Field: "balance",
			Aggregation: &TreeAggregation{
				Enabled:            true,
				Method:             "sum",
				ExcludeParentValue: true,
			},
		},
	}

	return CalculateAggregations(tree, columns)
}

// CalculateAccountBalance calculates the balance for an account group
func CalculateAccountBalance(node TreeNode) float64 {
	if !node.HasChildren {
		if bal, ok := toFloat64(node.Data["balance"]); ok {
			return bal
		}
		return 0
	}

	total := 0.0
	for _, child := range node.Children {
		total += CalculateAccountBalance(child)
	}

	return total
}

// FilterAccountsByType filters accounts by type (asset, liability, etc.)
func FilterAccountsByType(nodes []TreeNode, accountType string) []TreeNode {
	return FilterTree(nodes, func(node TreeNode) bool {
		if nodeType, ok := node.Data["accountType"].(string); ok {
			return nodeType == accountType
		}
		return false
	})
}

// GetAccountHierarchyPath returns the full path string for an account
func GetAccountHierarchyPath(nodes []TreeNode, accountCode string) string {
	path := GetNodePath(nodes, accountCode)
	if len(path) == 0 {
		return ""
	}

	labels := make([]string, len(path))
	for i, id := range path {
		if node := GetNodeByID(nodes, id); node != nil {
			labels[i] = node.Label
		}
	}

	return strings.Join(labels, " > ")
}

// ============================================================================
// TREE VALIDATION UTILITIES
// ============================================================================

// ValidateTree checks the tree structure for common issues
func ValidateTree(nodes []TreeNode) []string {
	var issues []string
	seen := make(map[string]bool)

	issues = append(issues, validateTreeRecursive(nodes, seen, 0)...)

	return issues
}

func validateTreeRecursive(nodes []TreeNode, seen map[string]bool, level int) []string {
	var issues []string

	for _, node := range nodes {
		// Check for duplicate IDs
		if seen[node.ID] {
			issues = append(issues, fmt.Sprintf("Duplicate node ID: %s", node.ID))
		}
		seen[node.ID] = true

		// Check for empty required fields
		if node.ID == "" {
			issues = append(issues, "Node with empty ID found")
		}
		if node.Label == "" {
			issues = append(issues, fmt.Sprintf("Node %s has empty label", node.ID))
		}

		// Check level consistency
		if node.Level != 0 && node.Level != level {
			issues = append(issues, fmt.Sprintf("Node %s has inconsistent level: expected %d, got %d", node.ID, level, node.Level))
		}

		// Check HasChildren consistency
		hasChildren := len(node.Children) > 0
		if node.HasChildren != hasChildren {
			issues = append(issues, fmt.Sprintf("Node %s HasChildren flag inconsistent with actual children", node.ID))
		}

		// Check IsLeaf consistency
		if node.IsLeaf && hasChildren {
			issues = append(issues, fmt.Sprintf("Node %s marked as leaf but has children", node.ID))
		}

		// Recursively validate children
		if len(node.Children) > 0 {
			issues = append(issues, validateTreeRecursive(node.Children, seen, level+1)...)
		}
	}

	return issues
}

// ============================================================================
// TREE SERIALIZATION UTILITIES
// ============================================================================

// TreeToJSON converts a tree structure to JSON
func TreeToJSON(nodes []TreeNode) (string, error) {
	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// TreeFromJSON creates a tree structure from JSON
func TreeFromJSON(jsonData string) ([]TreeNode, error) {
	var nodes []TreeNode
	err := json.Unmarshal([]byte(jsonData), &nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// ============================================================================
// TREE EXPORT/IMPORT UTILITIES
// ============================================================================

// ExportToCSV exports tree data to CSV format
func ExportToCSV(nodes []TreeNode, columns []TreeColumn, includeAggregates bool) (string, error) {
	flat := FlattenTree(nodes)

	var lines []string

	// Header row
	headers := []string{"Level", "Path"}
	for _, col := range columns {
		headers = append(headers, col.Label)
	}
	lines = append(lines, strings.Join(headers, ","))

	// Data rows
	for _, node := range flat {
		row := []string{
			fmt.Sprintf("%d", node.Level),
			getNodePathString(nodes, node.ID),
		}

		for _, col := range columns {
			var value interface{}
			if includeAggregates && node.HasChildren {
				value = GetAggregatedValue(node, col.Field)
			} else {
				value = node.Data[col.Field]
			}
			row = append(row, escapeCSV(fmt.Sprintf("%v", value)))
		}

		lines = append(lines, strings.Join(row, ","))
	}

	return strings.Join(lines, "\n"), nil
}

// escapeCSV escapes values for CSV format
func escapeCSV(value string) string {
	if strings.ContainsAny(value, ",\"\n") {
		value = strings.ReplaceAll(value, "\"", "\"\"")
		return fmt.Sprintf("\"%s\"", value)
	}
	return value
}

// getNodePathString gets the path string for a node
func getNodePathString(nodes []TreeNode, nodeID string) string {
	path := GetNodePath(nodes, nodeID)
	if len(path) == 0 {
		return ""
	}

	labels := make([]string, len(path))
	for i, id := range path {
		if node := GetNodeByID(nodes, id); node != nil {
			labels[i] = node.Label
		}
	}

	return strings.Join(labels, " > ")
}

// ExportToJSON exports tree data to JSON format
func ExportToJSON(nodes []TreeNode, includeAggregates bool) (string, error) {
	if includeAggregates {
		// Include aggregated values in export
		enriched := enrichWithAggregates(nodes)
		return TreeToJSON(enriched)
	}
	return TreeToJSON(nodes)
}

func enrichWithAggregates(nodes []TreeNode) []TreeNode {
	var result []TreeNode

	for _, node := range nodes {
		if node.Metadata != nil {
			// Move aggregated values to main Data
			for key, value := range node.Metadata {
				if strings.HasSuffix(key, "_aggregated") {
					if node.Data == nil {
						node.Data = make(map[string]interface{})
					}
					node.Data[key] = value
				}
			}
		}

		if len(node.Children) > 0 {
			node.Children = enrichWithAggregates(node.Children)
		}

		result = append(result, node)
	}

	return result
}

// ============================================================================
// TREE STATISTICS UTILITIES
// ============================================================================

// TreeStats contains statistics about a tree structure
type TreeStats struct {
	TotalNodes    int     `json:"totalNodes"`
	MaxDepth      int     `json:"maxDepth"`
	LeafNodes     int     `json:"leafNodes"`
	BranchNodes   int     `json:"branchNodes"`
	SelectedNodes int     `json:"selectedNodes"`
	ExpandedNodes int     `json:"expandedNodes"`
	AvgChildren   float64 `json:"avgChildren"`
}

// GetTreeStats calculates statistics for a tree
func GetTreeStats(nodes []TreeNode) TreeStats {
	stats := TreeStats{}

	calculateStatsRecursive(nodes, &stats, 0)

	// Calculate average children per branch node
	if stats.BranchNodes > 0 {
		totalChildren := stats.TotalNodes - len(nodes) // Total nodes minus root nodes
		stats.AvgChildren = float64(totalChildren) / float64(stats.BranchNodes)
	}

	return stats
}

func calculateStatsRecursive(nodes []TreeNode, stats *TreeStats, level int) {
	for _, node := range nodes {
		stats.TotalNodes++

		if level > stats.MaxDepth {
			stats.MaxDepth = level
		}

		if node.Selected {
			stats.SelectedNodes++
		}

		if node.Expanded {
			stats.ExpandedNodes++
		}

		if len(node.Children) == 0 {
			stats.LeafNodes++
		} else {
			stats.BranchNodes++
			calculateStatsRecursive(node.Children, stats, level+1)
		}
	}
}

// ============================================================================
// TREE TRANSFORMATION UTILITIES
// ============================================================================

// TransformTree applies a transformation function to all nodes in the tree
func TransformTree(nodes []TreeNode, transform func(TreeNode) TreeNode) []TreeNode {
	var transformed []TreeNode

	for _, node := range nodes {
		transformedNode := transform(node)

		if len(node.Children) > 0 {
			transformedNode.Children = TransformTree(node.Children, transform)
		}

		transformed = append(transformed, transformedNode)
	}

	return transformed
}

// NormalizeTree ensures all tree nodes have consistent metadata
func NormalizeTree(nodes []TreeNode) []TreeNode {
	return normalizeTreeRecursive(nodes, "", 0, []string{})
}

func normalizeTreeRecursive(nodes []TreeNode, parentID string, level int, parentPath []string) []TreeNode {
	var normalized []TreeNode

	for _, node := range nodes {
		// Set parent ID and level
		node.ParentID = parentID
		node.Level = level

		// Set HasChildren and IsLeaf flags
		node.HasChildren = len(node.Children) > 0
		node.IsLeaf = !node.HasChildren

		// Calculate path
		node.Path = append(parentPath, node.ID)

		// Initialize maps if nil
		if node.Data == nil {
			node.Data = make(map[string]interface{})
		}
		if node.Metadata == nil {
			node.Metadata = make(map[string]interface{})
		}

		// Recursively normalize children
		if len(node.Children) > 0 {
			node.Children = normalizeTreeRecursive(node.Children, node.ID, level+1, node.Path)
		}

		normalized = append(normalized, node)
	}

	return normalized
}

// ============================================================================
// TREE SORTING UTILITIES
// ============================================================================

// SortTree sorts all nodes in the tree by a given comparison function
func SortTree(nodes []TreeNode, less func(a, b TreeNode) bool) []TreeNode {
	return sortTreeRecursive(nodes, less)
}

func sortTreeRecursive(nodes []TreeNode, less func(a, b TreeNode) bool) []TreeNode {
	// Sort current level using bubble sort (simple implementation)
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if less(nodes[j], nodes[i]) {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}

	// Recursively sort children
	for i := range nodes {
		if len(nodes[i].Children) > 0 {
			nodes[i].Children = sortTreeRecursive(nodes[i].Children, less)
		}
	}

	return nodes
}

// SortByLabel sorts tree nodes alphabetically by label
func SortByLabel(nodes []TreeNode) []TreeNode {
	return SortTree(nodes, func(a, b TreeNode) bool {
		return strings.ToLower(a.Label) < strings.ToLower(b.Label)
	})
}

// SortByType sorts tree nodes by type (folders first, then files)
func SortByType(nodes []TreeNode) []TreeNode {
	return SortTree(nodes, func(a, b TreeNode) bool {
		// Folders (HasChildren) come before files (IsLeaf)
		if a.HasChildren && b.IsLeaf {
			return true
		}
		if a.IsLeaf && b.HasChildren {
			return false
		}
		// If same type, sort by label
		return strings.ToLower(a.Label) < strings.ToLower(b.Label)
	})
}

// SortByColumn sorts tree nodes by a specific column value
func SortByColumn(nodes []TreeNode, columnField string, ascending bool) []TreeNode {
	return SortTree(nodes, func(a, b TreeNode) bool {
		aVal := a.Data[columnField]
		bVal := b.Data[columnField]

		// Try numeric comparison first
		if aFloat, aOk := toFloat64(aVal); aOk {
			if bFloat, bOk := toFloat64(bVal); bOk {
				if ascending {
					return aFloat < bFloat
				}
				return aFloat > bFloat
			}
		}

		// Fall back to string comparison
		aStr := fmt.Sprintf("%v", aVal)
		bStr := fmt.Sprintf("%v", bVal)

		if ascending {
			return strings.ToLower(aStr) < strings.ToLower(bStr)
		}
		return strings.ToLower(aStr) > strings.ToLower(bStr)
	})
}

// ============================================================================
// TREE PERFORMANCE UTILITIES
// ============================================================================

// PaginateTree returns a subset of nodes for virtual scrolling
func PaginateTree(nodes []TreeNode, start, count int) []TreeNode {
	flat := FlattenTree(nodes)

	end := start + count
	if end > len(flat) {
		end = len(flat)
	}

	if start >= len(flat) {
		return []TreeNode{}
	}

	return flat[start:end]
}

// GetVisibleNodes returns only expanded/visible nodes for rendering
func GetVisibleNodes(nodes []TreeNode) []TreeNode {
	var visible []TreeNode

	for _, node := range nodes {
		visible = append(visible, node)

		if node.Expanded && len(node.Children) > 0 {
			visible = append(visible, GetVisibleNodes(node.Children)...)
		}
	}

	return visible
}

// TreeState caches the current tree state (expanded, selected nodes)
type TreeState struct {
	ExpandedNodes  []string `json:"expanded"`
	SelectedNodes  []string `json:"selected"`
	ScrollPosition int      `json:"scrollPosition"`
}

// GetTreeState extracts the current state from a tree
func GetTreeState(nodes []TreeNode) TreeState {
	flat := FlattenTree(nodes)

	state := TreeState{
		ExpandedNodes: []string{},
		SelectedNodes: []string{},
	}

	for _, node := range flat {
		if node.Expanded {
			state.ExpandedNodes = append(state.ExpandedNodes, node.ID)
		}
		if node.Selected {
			state.SelectedNodes = append(state.SelectedNodes, node.ID)
		}
	}

	return state
}

// ApplyTreeState applies a saved state to a tree
func ApplyTreeState(nodes []TreeNode, state TreeState) []TreeNode {
	expandedMap := make(map[string]bool)
	for _, id := range state.ExpandedNodes {
		expandedMap[id] = true
	}

	selectedMap := make(map[string]bool)
	for _, id := range state.SelectedNodes {
		selectedMap[id] = true
	}

	return applyStateRecursive(nodes, expandedMap, selectedMap)
}

func applyStateRecursive(nodes []TreeNode, expanded, selected map[string]bool) []TreeNode {
	var result []TreeNode

	for _, node := range nodes {
		node.Expanded = expanded[node.ID]
		node.Selected = selected[node.ID]

		if len(node.Children) > 0 {
			node.Children = applyStateRecursive(node.Children, expanded, selected)
		}

		result = append(result, node)
	}

	return result
}

// ============================================================================
// TREE CLONING UTILITIES
// ============================================================================

// CloneTree creates a deep copy of a tree
func CloneTree(nodes []TreeNode) []TreeNode {
	var cloned []TreeNode

	for _, node := range nodes {
		clonedNode := cloneNode(node)
		if len(node.Children) > 0 {
			clonedNode.Children = CloneTree(node.Children)
		}
		cloned = append(cloned, clonedNode)
	}

	return cloned
}

// cloneNode creates a deep copy of a single node
func cloneNode(node TreeNode) TreeNode {
	cloned := node

	// Deep copy Data map
	if node.Data != nil {
		cloned.Data = make(map[string]interface{})
		for k, v := range node.Data {
			cloned.Data[k] = v
		}
	}

	// Deep copy Metadata map
	if node.Metadata != nil {
		cloned.Metadata = make(map[string]interface{})
		for k, v := range node.Metadata {
			cloned.Metadata[k] = v
		}
	}

	// Deep copy Path slice
	if node.Path != nil {
		cloned.Path = make([]string, len(node.Path))
		copy(cloned.Path, node.Path)
	}

	// Don't copy children here, handled by CloneTree
	cloned.Children = nil

	return cloned
}

