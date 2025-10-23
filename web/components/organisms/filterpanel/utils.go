package filterpanel

import (
	"fmt"

	"github.com/niiniyare/erp/web/components/atoms"
)

// getInputTypeForFilter returns the appropriate input type for a filter
func getInputTypeForFilter(filterType FilterType) atoms.InputType {
	switch filterType {
	case FilterTypeNumber:
		return atoms.InputTypeNumber
	case FilterTypeDate:
		return atoms.InputTypeDate
	case FilterTypeSearch:
		return atoms.InputTypeSearch
	default:
		return atoms.InputTypeText
	}
}

// convertToSelectOptions converts FilterOption to atoms.SelectOption
func convertToSelectOptions(options []FilterOption) []atoms.SelectOption {
	selectOptions := make([]atoms.SelectOption, len(options))
	for i, option := range options {
		selectOptions[i] = atoms.SelectOption{
			Value:    option.Value,
			Label:    option.Label,
			Selected: option.Selected,
			Disabled: option.Disabled,
		}
	}
	return selectOptions
}

// createTextFilter creates a simple text filter field
func createTextFilter(name, label, placeholder string) FilterField {
	return FilterField{
		Name:        name,
		Label:       label,
		Type:        FilterTypeText,
		Placeholder: placeholder,
		Operator:    OperatorContains,
	}
}

// createSelectFilter creates a select filter field with options
func createSelectFilter(name, label string, options []FilterOption, multiple bool) FilterField {
	return FilterField{
		Name:     name,
		Label:    label,
		Type:     FilterTypeSelect,
		Options:  options,
		Multiple: multiple,
		Operator: OperatorIn,
	}
}

// createDateRangeFilter creates a date range filter field
func createDateRangeFilter(name, label string) FilterField {
	return FilterField{
		Name:     name,
		Label:    label,
		Type:     FilterTypeDateRange,
		Operator: OperatorBetween,
	}
}

// createNumberFilter creates a number filter field
func createNumberFilter(name, label string, min, max any) FilterField {
	return FilterField{
		Name:     name,
		Label:    label,
		Type:     FilterTypeNumber,
		Min:      min,
		Max:      max,
		Operator: OperatorEquals,
	}
}

// createSearchFilter creates a search filter with autocomplete
func createSearchFilter(name, label, placeholder, hxGet string) FilterField {
	return FilterField{
		Name:        name,
		Label:       label,
		Type:        FilterTypeSearch,
		Placeholder: placeholder,
		HxGet:       hxGet,
		HxTrigger:   "keyup changed delay:300ms",
		Operator:    OperatorContains,
	}
}

// createCheckboxFilter creates a checkbox group filter
func createCheckboxFilter(name, label string, options []FilterOption) FilterField {
	return FilterField{
		Name:     name,
		Label:    label,
		Type:     FilterTypeCheckbox,
		Options:  options,
		Multiple: true,
		Operator: OperatorIn,
	}
}

// createBasicFilterGroup creates a basic filter group
func createBasicFilterGroup(title string, fields []FilterField, collapsible bool) FilterGroup {
	return FilterGroup{
		Title:       title,
		Fields:      fields,
		Collapsible: collapsible,
		Collapsed:   false,
		Columns:     1,
	}
}

// createResponsiveFilterGroup creates a responsive multi-column filter group
func createResponsiveFilterGroup(title string, fields []FilterField, columns int) FilterGroup {
	return FilterGroup{
		Title:       title,
		Fields:      fields,
		Collapsible: true,
		Collapsed:   false,
		Columns:     columns,
		Spacing:     "gap-4",
	}
}

// getUserManagementFilters creates standard user management filters
func getUserManagementFilters() FilterPanelProps {
	return FilterPanelProps{
		ID:          "user-filters",
		Title:       "Filter Users",
		Collapsible: true,
		ShowApply:   true,
		ShowReset:   true,
		ShowCount:   true,
		AutoApply:   false,
		HxPost:      "/api/users/filter",
		HxTarget:    "#user-table",
		HxSwap:      "innerHTML",
		Groups: []FilterGroup{
			{
				Title:       "Basic Information",
				Collapsible: true,
				Columns:     2,
				Fields: []FilterField{
					createTextFilter("name", "Name", "Enter user name"),
					createTextFilter("email", "Email", "Enter email address"),
					createSelectFilter("status", "Status", []FilterOption{
						{Value: "", Label: "All Statuses", Selected: true},
						{Value: "active", Label: "Active"},
						{Value: "inactive", Label: "Inactive"},
						{Value: "pending", Label: "Pending"},
					}, false),
					createSelectFilter("role", "Role", []FilterOption{
						{Value: "", Label: "All Roles", Selected: true},
						{Value: "admin", Label: "Administrator"},
						{Value: "user", Label: "User"},
						{Value: "manager", Label: "Manager"},
					}, true),
				},
			},
			{
				Title:       "Date Filters",
				Collapsible: true,
				Collapsed:   true,
				Columns:     2,
				Fields: []FilterField{
					createDateRangeFilter("created_date", "Created Date"),
					createDateRangeFilter("last_login", "Last Login"),
				},
			},
		},
	}
}

// getProjectFilters creates project management filters
func getProjectFilters() FilterPanelProps {
	return FilterPanelProps{
		ID:          "project-filters",
		Title:       "Filter Projects",
		Collapsible: true,
		ShowApply:   true,
		ShowReset:   true,
		ShowPresets: true,
		AutoApply:   true,
		HxPost:      "/api/projects/filter",
		HxTarget:    "#project-list",
		Fields: []FilterField{
			createSearchFilter("search", "Search Projects", "Search by name or description", "/api/projects/search"),
			createSelectFilter("status", "Status", []FilterOption{
				{Value: "active", Label: "Active", Selected: true},
				{Value: "completed", Label: "Completed"},
				{Value: "archived", Label: "Archived"},
			}, true),
		},
		Groups: []FilterGroup{
			{
				Title:       "Project Details",
				Collapsible: true,
				Columns:     3,
				Fields: []FilterField{
					createSelectFilter("priority", "Priority", []FilterOption{
						{Value: "low", Label: "Low"},
						{Value: "medium", Label: "Medium"},
						{Value: "high", Label: "High"},
						{Value: "urgent", Label: "Urgent"},
					}, true),
					createDateRangeFilter("due_date", "Due Date"),
					createNumberFilter("budget", "Budget Range", 0, 1000000),
				},
			},
		},
	}
}

// getFinancialFilters creates financial data filters
func getFinancialFilters() FilterPanelProps {
	return FilterPanelProps{
		ID:          "financial-filters",
		Title:       "Financial Filters",
		Collapsible: true,
		ShowApply:   true,
		ShowReset:   true,
		ShowSave:    true,
		ShowPresets: true,
		ShowCount:   true,
		AutoApply:   false,
		SaveState:   true,
		HxPost:      "/api/finance/transactions/filter",
		HxTarget:    "#transaction-table",
		Groups: []FilterGroup{
			{
				Title:       "Transaction Details",
				Collapsible: false,
				Columns:     2,
				Fields: []FilterField{
					createDateRangeFilter("date_range", "Date Range"),
					createSelectFilter("account", "Account", []FilterOption{
						{Value: "", Label: "All Accounts", Selected: true},
						{Value: "1000", Label: "Cash"},
						{Value: "1100", Label: "Accounts Receivable"},
						{Value: "2000", Label: "Accounts Payable"},
					}, true),
					createSelectFilter("type", "Transaction Type", []FilterOption{
						{Value: "", Label: "All Types", Selected: true},
						{Value: "debit", Label: "Debit"},
						{Value: "credit", Label: "Credit"},
					}, false),
					createNumberFilter("amount", "Amount Range", 0, nil),
				},
			},
			{
				Title:       "Categories & Tags",
				Collapsible: true,
				Collapsed:   true,
				Columns:     1,
				Fields: []FilterField{
					createCheckboxFilter("categories", "Categories", []FilterOption{
						{Value: "income", Label: "Income"},
						{Value: "expense", Label: "Expense"},
						{Value: "transfer", Label: "Transfer"},
						{Value: "adjustment", Label: "Adjustment"},
					}),
				},
			},
		},
	}
}

// createQuickSearchFilter creates a simplified quick filter
func createQuickSearchFilter(hxTarget, hxPost string) QuickFilterProps {
	return QuickFilterProps{
		Horizontal: true,
		Compact:    false,
		HxTarget:   hxTarget,
		HxPost:     hxPost,
		Fields: []FilterField{
			{
				Name:        "search",
				Type:        FilterTypeSearch,
				Placeholder: "Search...",
				Width:       "flex-1",
			},
			{
				Name: "status",
				Type: FilterTypeSelect,
				Options: []FilterOption{
					{Value: "", Label: "All Status", Selected: true},
					{Value: "active", Label: "Active"},
					{Value: "inactive", Label: "Inactive"},
				},
				Width: "w-48",
			},
		},
	}
}

// validateFilterField validates a filter field configuration
func validateFilterField(field FilterField) error {
	if field.Name == "" {
		return fmt.Errorf("filter field name is required")
	}

	if field.Label == "" {
		return fmt.Errorf("filter field label is required")
	}

	if field.Type == FilterTypeSelect || field.Type == FilterTypeRadio || field.Type == FilterTypeCheckbox {
		if len(field.Options) == 0 {
			return fmt.Errorf("filter field %s requires options for type %s", field.Name, field.Type)
		}
	}

	return nil
}

// getFilterFieldJavaScript returns JavaScript helper functions for filter fields
func getFilterFieldJavaScript() string {
	return `
// Helper function to get checkbox values
function getCheckboxValues(name) {
	const checkboxes = document.querySelectorAll('input[name="' + name + '[]"]:checked');
	return Array.from(checkboxes).map(cb => cb.value);
}

// Helper function to format date values
function formatDateValue(date) {
	if (!date) return '';
	const d = new Date(date);
	return d.toISOString().split('T')[0];
}

// Helper function to validate date ranges
function validateDateRange(startDate, endDate) {
	if (!startDate || !endDate) return true;
	return new Date(startDate) <= new Date(endDate);
}

// Helper function to debounce filter applications
function debounceFilter(func, wait) {
	let timeout;
	return function executedFunction(...args) {
		const later = () => {
			clearTimeout(timeout);
			func(...args);
		};
		clearTimeout(timeout);
		timeout = setTimeout(later, wait);
	};
}
`
}
