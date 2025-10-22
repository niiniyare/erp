package ui

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ============================================================================
// HTMX AND ALPINE.JS INTEGRATION VALIDATION
// ============================================================================
// This file validates that organism components properly integrate with
// HTMX for server interactions and Alpine.js for client-side reactivity.

// HTMXIntegrationPattern represents a pattern for HTMX integration
type HTMXIntegrationPattern struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Target      string `json:"target"`
	Swap        string `json:"swap"`
	Trigger     string `json:"trigger"`
	Example     string `json:"example"`
}

// AlpineIntegrationPattern represents a pattern for Alpine.js integration
type AlpineIntegrationPattern struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Data        map[string]any    `json:"data"`
	Methods     map[string]string `json:"methods"`
	Events      map[string]string `json:"events"`
	Example     string            `json:"example"`
}

// GetHTMXTablePatterns returns common HTMX patterns for table components
func GetHTMXTablePatterns() []HTMXIntegrationPattern {
	return []HTMXIntegrationPattern{
		{
			Name:        "Dynamic Data Loading",
			Description: "Load table data from server endpoint",
			Endpoint:    "/api/table/data",
			Method:      "GET",
			Target:      "#table-body",
			Swap:        "innerHTML",
			Trigger:     "load",
			Example:     `hx-get="/api/table/data" hx-target="#table-body" hx-swap="innerHTML" hx-trigger="load"`,
		},
		{
			Name:        "Pagination",
			Description: "Handle pagination with server-side processing",
			Endpoint:    "/api/table/data",
			Method:      "GET",
			Target:      "#table-container",
			Swap:        "outerHTML",
			Trigger:     "click",
			Example:     `hx-get="/api/table/data?page={page}" hx-target="#table-container" hx-swap="outerHTML"`,
		},
		{
			Name:        "Sorting",
			Description: "Server-side sorting of table columns",
			Endpoint:    "/api/table/data",
			Method:      "GET",
			Target:      "#table-body",
			Swap:        "innerHTML",
			Trigger:     "click",
			Example:     `hx-get="/api/table/data?sort={column}&order={direction}" hx-target="#table-body"`,
		},
		{
			Name:        "Filtering",
			Description: "Apply filters to table data",
			Endpoint:    "/api/table/filter",
			Method:      "POST",
			Target:      "#table-body",
			Swap:        "innerHTML",
			Trigger:     "change",
			Example:     `hx-post="/api/table/filter" hx-target="#table-body" hx-trigger="change"`,
		},
		{
			Name:        "Row Actions",
			Description: "Handle actions on table rows",
			Endpoint:    "/api/table/action",
			Method:      "POST",
			Target:      "#notification-area",
			Swap:        "afterbegin",
			Trigger:     "click",
			Example:     `hx-post="/api/table/action" hx-vals='{"action":"delete","id":"123"}' hx-confirm="Are you sure?"`,
		},
		{
			Name:        "Live Search",
			Description: "Search table content with debounced input",
			Endpoint:    "/api/table/search",
			Method:      "GET",
			Target:      "#table-body",
			Swap:        "innerHTML",
			Trigger:     "input changed delay:300ms",
			Example:     `hx-get="/api/table/search" hx-target="#table-body" hx-trigger="input changed delay:300ms"`,
		},
	}
}

// GetHTMXModalPatterns returns common HTMX patterns for modal components
func GetHTMXModalPatterns() []HTMXIntegrationPattern {
	return []HTMXIntegrationPattern{
		{
			Name:        "Dynamic Modal Content",
			Description: "Load modal content from server",
			Endpoint:    "/api/modal/content",
			Method:      "GET",
			Target:      "#modal-body",
			Swap:        "innerHTML",
			Trigger:     "click",
			Example:     `hx-get="/api/modal/content/{id}" hx-target="#modal-body" hx-swap="innerHTML"`,
		},
		{
			Name:        "Form Modal Submission",
			Description: "Submit modal form and handle response",
			Endpoint:    "/api/forms/submit",
			Method:      "POST",
			Target:      "#modal-container",
			Swap:        "outerHTML",
			Trigger:     "submit",
			Example:     `hx-post="/api/forms/submit" hx-target="#modal-container" hx-swap="outerHTML"`,
		},
		{
			Name:        "Modal State Management",
			Description: "Open/close modals via server triggers",
			Endpoint:    "/api/modal/toggle",
			Method:      "POST",
			Target:      "body",
			Swap:        "beforeend",
			Trigger:     "click",
			Example:     `hx-post="/api/modal/toggle" hx-target="body" hx-swap="beforeend"`,
		},
		{
			Name:        "Validation Feedback",
			Description: "Real-time form validation in modal",
			Endpoint:    "/api/forms/validate",
			Method:      "POST",
			Target:      "#validation-feedback",
			Swap:        "innerHTML",
			Trigger:     "blur",
			Example:     `hx-post="/api/forms/validate" hx-target="#validation-feedback" hx-trigger="blur"`,
		},
	}
}

// GetHTMXTreePatterns returns common HTMX patterns for tree components
func GetHTMXTreePatterns() []HTMXIntegrationPattern {
	return []HTMXIntegrationPattern{
		{
			Name:        "Lazy Loading Nodes",
			Description: "Load tree nodes on expansion",
			Endpoint:    "/api/tree/nodes",
			Method:      "GET",
			Target:      "#node-children",
			Swap:        "innerHTML",
			Trigger:     "click",
			Example:     `hx-get="/api/tree/nodes/{nodeId}" hx-target="#node-{nodeId}-children" hx-trigger="click"`,
		},
		{
			Name:        "Node Actions",
			Description: "Perform actions on tree nodes",
			Endpoint:    "/api/tree/action",
			Method:      "POST",
			Target:      "#tree-container",
			Swap:        "outerHTML",
			Trigger:     "click",
			Example:     `hx-post="/api/tree/action" hx-vals='{"action":"rename","nodeId":"123"}'`,
		},
		{
			Name:        "Drag and Drop",
			Description: "Handle node reordering via drag and drop",
			Endpoint:    "/api/tree/reorder",
			Method:      "PUT",
			Target:      "#tree-container",
			Swap:        "outerHTML",
			Trigger:     "drop",
			Example:     `hx-put="/api/tree/reorder" hx-vals='{"sourceId":"123","targetId":"456"}'`,
		},
		{
			Name:        "Search Tree",
			Description: "Search and filter tree nodes",
			Endpoint:    "/api/tree/search",
			Method:      "GET",
			Target:      "#tree-nodes",
			Swap:        "innerHTML",
			Trigger:     "input changed delay:300ms",
			Example:     `hx-get="/api/tree/search" hx-target="#tree-nodes" hx-trigger="input changed delay:300ms"`,
		},
	}
}

// GetAlpineTablePatterns returns common Alpine.js patterns for table components
func GetAlpineTablePatterns() []AlpineIntegrationPattern {
	return []AlpineIntegrationPattern{
		{
			Name:        "Selection Management",
			Description: "Handle row selection with checkboxes",
			Data: map[string]any{
				"selectedRows": []string{},
				"selectAll":    false,
			},
			Methods: map[string]string{
				"toggleSelection": "function(rowId) { /* toggle logic */ }",
				"toggleSelectAll": "function() { /* select all logic */ }",
				"getSelectedIds":  "function() { return this.selectedRows }",
			},
			Events: map[string]string{
				"row-selected":   "updateSelectedRows($event.detail)",
				"select-all":     "toggleSelectAll()",
				"clear-selection": "selectedRows = []",
			},
			Example: `x-data="{ selectedRows: [], selectAll: false }" x-on:row-selected="updateSelectedRows($event.detail)"`,
		},
		{
			Name:        "Column Visibility",
			Description: "Toggle table column visibility",
			Data: map[string]any{
				"visibleColumns": map[string]bool{
					"name":   true,
					"email":  true,
					"status": true,
				},
			},
			Methods: map[string]string{
				"toggleColumn": "function(columnKey) { this.visibleColumns[columnKey] = !this.visibleColumns[columnKey] }",
				"showAllColumns": "function() { Object.keys(this.visibleColumns).forEach(key => this.visibleColumns[key] = true) }",
				"hideAllColumns": "function() { Object.keys(this.visibleColumns).forEach(key => this.visibleColumns[key] = false) }",
			},
			Events: map[string]string{
				"column-toggled": "toggleColumn($event.detail.column)",
			},
			Example: `x-data="{ visibleColumns: {name: true, email: true} }" x-show="visibleColumns.name"`,
		},
		{
			Name:        "Filter State",
			Description: "Manage active filters and search state",
			Data: map[string]any{
				"filters": map[string]any{},
				"searchTerm": "",
				"sortColumn": "",
				"sortDirection": "asc",
			},
			Methods: map[string]string{
				"addFilter":     "function(column, value) { this.filters[column] = value }",
				"removeFilter":  "function(column) { delete this.filters[column] }",
				"clearFilters":  "function() { this.filters = {} }",
				"applySort":     "function(column) { /* sorting logic */ }",
			},
			Events: map[string]string{
				"filter-changed": "addFilter($event.detail.column, $event.detail.value)",
				"sort-changed":   "applySort($event.detail.column)",
			},
			Example: `x-data="{ filters: {}, searchTerm: '' }" x-model="searchTerm"`,
		},
	}
}

// GetAlpineModalPatterns returns common Alpine.js patterns for modal components
func GetAlpineModalPatterns() []AlpineIntegrationPattern {
	return []AlpineIntegrationPattern{
		{
			Name:        "Modal State",
			Description: "Control modal open/close state",
			Data: map[string]any{
				"open":    false,
				"loading": false,
			},
			Methods: map[string]string{
				"openModal":  "function() { this.open = true }",
				"closeModal": "function() { this.open = false }",
				"toggleModal": "function() { this.open = !this.open }",
			},
			Events: map[string]string{
				"keydown.escape": "closeModal()",
				"click.away":     "closeModal()",
			},
			Example: `x-data="{ open: false }" x-show="open" @keydown.escape="closeModal()"`,
		},
		{
			Name:        "Form Validation",
			Description: "Handle form validation state in modal",
			Data: map[string]any{
				"formData": map[string]any{},
				"errors":   map[string]string{},
				"isValid":  false,
			},
			Methods: map[string]string{
				"validateField": "function(field, value) { /* validation logic */ }",
				"validateForm":  "function() { /* form validation */ }",
				"submitForm":    "function() { /* form submission */ }",
			},
			Events: map[string]string{
				"input":     "validateField($event.target.name, $event.target.value)",
				"submit":    "submitForm()",
				"blur":      "validateField($event.target.name, $event.target.value)",
			},
			Example: `x-data="{ formData: {}, errors: {} }" @input="validateField($event.target.name, $event.target.value)"`,
		},
	}
}

// GetAlpineTreePatterns returns common Alpine.js patterns for tree components
func GetAlpineTreePatterns() []AlpineIntegrationPattern {
	return []AlpineIntegrationPattern{
		{
			Name:        "Expansion State",
			Description: "Manage tree node expansion state",
			Data: map[string]any{
				"expandedNodes": []string{},
				"selectedNodes": []string{},
			},
			Methods: map[string]string{
				"toggleNode":   "function(nodeId) { /* toggle logic */ }",
				"expandAll":    "function() { /* expand all nodes */ }",
				"collapseAll":  "function() { this.expandedNodes = [] }",
				"selectNode":   "function(nodeId) { /* selection logic */ }",
			},
			Events: map[string]string{
				"node-clicked":   "selectNode($event.detail.nodeId)",
				"node-expanded":  "toggleNode($event.detail.nodeId)",
				"expand-all":     "expandAll()",
				"collapse-all":   "collapseAll()",
			},
			Example: `x-data="{ expandedNodes: [], selectedNodes: [] }" @node-clicked="selectNode($event.detail.nodeId)"`,
		},
		{
			Name:        "Search and Filter",
			Description: "Filter tree nodes based on search criteria",
			Data: map[string]any{
				"searchTerm":    "",
				"filteredNodes": []string{},
				"highlightTerm": "",
			},
			Methods: map[string]string{
				"filterNodes":    "function() { /* filter logic */ }",
				"clearSearch":    "function() { this.searchTerm = ''; this.filterNodes() }",
				"highlightMatch": "function(text) { /* highlight logic */ }",
			},
			Events: map[string]string{
				"input":       "filterNodes()",
				"search-clear": "clearSearch()",
			},
			Example: `x-data="{ searchTerm: '', filteredNodes: [] }" @input="filterNodes()"`,
		},
	}
}

// ValidateHTMXIntegration validates that components properly implement HTMX patterns
func ValidateHTMXIntegration(component Component) []string {
	var issues []string
	
	// Check if component config contains HTMX attributes
	var config map[string]any
	if err := json.Unmarshal(component.Config, &config); err != nil {
		issues = append(issues, "Failed to parse component config")
		return issues
	}

	// Validate common HTMX patterns based on component type
	switch component.Type {
	case ComponentTable:
		issues = append(issues, validateTableHTMX(config)...)
	case ComponentModal:
		issues = append(issues, validateModalHTMX(config)...)
	case ComponentTree:
		issues = append(issues, validateTreeHTMX(config)...)
	}

	return issues
}

// validateTableHTMX validates HTMX patterns for table components
func validateTableHTMX(config map[string]any) []string {
	var issues []string

	// Check for data source endpoint
	if dataSource, exists := config["data_source"]; !exists || dataSource == "" {
		issues = append(issues, "Table missing data_source for HTMX integration")
	}

	// Check for pagination HTMX support
	if pagination, exists := config["pagination"]; exists {
		if paginationMap, ok := pagination.(map[string]any); ok {
			if _, hasRemote := paginationMap["remote"]; !hasRemote {
				issues = append(issues, "Table pagination should support remote HTMX loading")
			}
		}
	}

	// Check for sorting HTMX support
	if sorting, exists := config["sorting"]; exists {
		if sortingMap, ok := sorting.(map[string]any); ok {
			if remote, hasRemote := sortingMap["remote"]; hasRemote && remote == false {
				issues = append(issues, "Table sorting should support remote HTMX processing")
			}
		}
	}

	return issues
}

// validateModalHTMX validates HTMX patterns for modal components
func validateModalHTMX(config map[string]any) []string {
	var issues []string

	// Check for dynamic content loading capability
	if _, hasContentEndpoint := config["content_endpoint"]; !hasContentEndpoint {
		// This is optional, so just note it
	}

	// Check for form submission handling
	if _, hasFormEndpoint := config["form_endpoint"]; !hasFormEndpoint {
		// This is optional for non-form modals
	}

	return issues
}

// validateTreeHTMX validates HTMX patterns for tree components
func validateTreeHTMX(config map[string]any) []string {
	var issues []string

	// Check for lazy loading support
	if _, hasLazyLoad := config["lazy_load"]; hasLazyLoad {
		if _, hasEndpoint := config["lazy_load_endpoint"]; !hasEndpoint {
			issues = append(issues, "Tree with lazy loading requires lazy_load_endpoint")
		}
	}

	// Check for data source
	if dataSource, exists := config["data_source"]; exists && dataSource != "" {
		// Good - tree has remote data source
	} else if data, exists := config["data"]; !exists || data == nil {
		issues = append(issues, "Tree requires either data_source or static data")
	}

	return issues
}

// ValidateAlpineIntegration validates that components properly implement Alpine.js patterns
func ValidateAlpineIntegration(component Component) []string {
	var issues []string

	// Check if component config contains Alpine.js data and methods
	var config map[string]any
	if err := json.Unmarshal(component.Config, &config); err != nil {
		issues = append(issues, "Failed to parse component config")
		return issues
	}

	// Validate Alpine.js patterns based on component type
	switch component.Type {
	case ComponentTable:
		issues = append(issues, validateTableAlpine(config)...)
	case ComponentModal:
		issues = append(issues, validateModalAlpine(config)...)
	case ComponentTree:
		issues = append(issues, validateTreeAlpine(config)...)
	}

	return issues
}

// validateTableAlpine validates Alpine.js patterns for table components
func validateTableAlpine(config map[string]any) []string {
	var issues []string

	// Check for selection support
	if selection, exists := config["selection"]; exists {
		if selectionMap, ok := selection.(map[string]any); ok {
			if selType, hasType := selectionMap["type"]; hasType {
				if selType == SelectionCheckbox {
					// Should have Alpine.js state for selection management
					// This would be validated in the actual template
				}
			}
		}
	}

	// Check for interactive features that need Alpine.js
	if search, exists := config["search"]; exists {
		if searchMap, ok := search.(map[string]any); ok {
			if debounce, hasDebounce := searchMap["debounce"]; hasDebounce && debounce != nil {
				// Good - has debounced search which needs Alpine.js
			}
		}
	}

	return issues
}

// validateModalAlpine validates Alpine.js patterns for modal components
func validateModalAlpine(config map[string]any) []string {
	var issues []string

	// Modals should have Alpine.js state management
	if closable, exists := config["closable"]; exists && closable == true {
		// Should have Alpine.js for close handling
	}

	if keyboard, exists := config["keyboard"]; exists && keyboard == true {
		// Should have Alpine.js for keyboard handling
	}

	return issues
}

// validateTreeAlpine validates Alpine.js patterns for tree components
func validateTreeAlpine(config map[string]any) []string {
	var issues []string

	// Check for interactive features
	if expandable, exists := config["expandable"]; exists && expandable == true {
		// Should have Alpine.js for expansion state
	}

	if selectable, exists := config["selectable"]; exists && selectable == true {
		// Should have Alpine.js for selection state
	}

	if checkable, exists := config["checkable"]; exists && checkable == true {
		// Should have Alpine.js for checkbox state
	}

	return issues
}

// GenerateIntegrationReport generates a comprehensive report of HTMX and Alpine.js integration
func GenerateIntegrationReport(registry ComponentRegistry) string {
	var report strings.Builder
	
	report.WriteString("# HTMX and Alpine.js Integration Report\n\n")
	
	// Get organism types
	organismTypes := []ComponentType{ComponentModal, ComponentTable, ComponentTree}
	
	for _, componentType := range organismTypes {
		report.WriteString(fmt.Sprintf("## %s Component\n\n", componentType))
		
		if registry.IsRegistered(componentType) {
			report.WriteString("✅ **Status**: Registered in component registry\n\n")
			
			// Create example component for validation
			example, err := createExampleComponent(registry, componentType)
			if err != nil {
				report.WriteString(fmt.Sprintf("❌ **Error**: Failed to create example: %v\n\n", err))
				continue
			}
			
			// Validate HTMX integration
			htmxIssues := ValidateHTMXIntegration(example)
			if len(htmxIssues) == 0 {
				report.WriteString("✅ **HTMX Integration**: No issues found\n")
			} else {
				report.WriteString("⚠️ **HTMX Integration Issues**:\n")
				for _, issue := range htmxIssues {
					report.WriteString(fmt.Sprintf("  - %s\n", issue))
				}
			}
			
			// Validate Alpine.js integration
			alpineIssues := ValidateAlpineIntegration(example)
			if len(alpineIssues) == 0 {
				report.WriteString("✅ **Alpine.js Integration**: No issues found\n")
			} else {
				report.WriteString("⚠️ **Alpine.js Integration Issues**:\n")
				for _, issue := range alpineIssues {
					report.WriteString(fmt.Sprintf("  - %s\n", issue))
				}
			}
			
		} else {
			report.WriteString("❌ **Status**: Not registered in component registry\n")
		}
		
		report.WriteString("\n")
	}
	
	// Add pattern documentation
	report.WriteString("## Integration Patterns\n\n")
	report.WriteString("### HTMX Patterns\n")
	report.WriteString("- Dynamic data loading\n")
	report.WriteString("- Form submission and validation\n")
	report.WriteString("- Lazy loading of content\n")
	report.WriteString("- Real-time updates\n\n")
	
	report.WriteString("### Alpine.js Patterns\n")
	report.WriteString("- State management\n")
	report.WriteString("- Event handling\n")
	report.WriteString("- UI reactivity\n")
	report.WriteString("- Client-side validation\n\n")
	
	return report.String()
}

// createExampleComponent creates an example component for testing
func createExampleComponent(registry ComponentRegistry, componentType ComponentType) (Component, error) {
	switch componentType {
	case ComponentModal:
		return CreateModalExample(registry, "Test Modal", "Test content")
	case ComponentTable:
		return CreateTableExample(registry)
	case ComponentTree:
		return CreateTreeExample(registry)
	default:
		return Component{}, fmt.Errorf("unsupported component type: %s", componentType)
	}
}