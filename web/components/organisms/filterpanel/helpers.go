package filterpanel

//nolint:unused // Helper functions may be used by templates or future features

import (
	"fmt"
	"strings"

	"github.com/niiniyare/erp/web/components/atoms"
)

// getFilterPanelClasses returns CSS classes for the filter panel
func getFilterPanelClasses(props FilterPanelProps) string {
	baseClasses := "bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg"

	var classes []string
	classes = append(classes, baseClasses)

	// Variant styling
	switch props.Variant {
	case "minimal":
		classes = append(classes, "border-0 bg-transparent")
	case "card":
		classes = append(classes, "shadow-md")
	default:
		classes = append(classes, "shadow-sm")
	}

	// Sticky positioning
	if props.Sticky {
		switch props.Position {
		case "top":
			classes = append(classes, "sticky top-0 z-30")
		case "left":
			classes = append(classes, "sticky left-0")
		case "right":
			classes = append(classes, "sticky right-0")
		default:
			classes = append(classes, "sticky top-0 z-30")
		}
	}

	// Width
	if props.Width != "" {
		classes = append(classes, props.Width)
	}

	// Custom classes
	if props.Class != "" {
		classes = append(classes, props.Class)
	}

	return strings.Join(classes, " ")
}

// getFilterGroupClasses returns CSS classes for filter groups
func getFilterGroupClasses(group FilterGroup) string {
	classes := []string{"border-b border-gray-200 dark:border-gray-700 last:border-b-0"}

	if group.Class != "" {
		classes = append(classes, group.Class)
	}

	return strings.Join(classes, " ")
}

// getFilterFieldClasses returns CSS classes for filter fields
func getFilterFieldClasses(field FilterField, columns int) string {
	var classes []string

	// Width based on field-specific setting or group columns
	if field.Width != "" {
		classes = append(classes, field.Width)
	} else if columns > 1 {
		switch columns {
		case 2:
			classes = append(classes, "w-1/2")
		case 3:
			classes = append(classes, "w-1/3")
		case 4:
			classes = append(classes, "w-1/4")
		default:
			classes = append(classes, "w-full")
		}
	} else {
		classes = append(classes, "w-full")
	}

	// Custom classes
	if field.Class != "" {
		classes = append(classes, field.Class)
	}

	return strings.Join(classes, " ")
}

// getFilterLayoutClasses returns grid classes for field layout
func getFilterLayoutClasses(columns int, spacing string) string {
	var classes []string

	if columns > 1 {
		classes = append(classes, "grid")
		switch columns {
		case 2:
			classes = append(classes, "grid-cols-1 md:grid-cols-2")
		case 3:
			classes = append(classes, "grid-cols-1 md:grid-cols-2 lg:grid-cols-3")
		case 4:
			classes = append(classes, "grid-cols-1 md:grid-cols-2 lg:grid-cols-4")
		default:
			classes = append(classes, fmt.Sprintf("grid-cols-%d", columns))
		}
	}

	// Spacing
	if spacing != "" {
		classes = append(classes, spacing)
	} else {
		classes = append(classes, "gap-4")
	}

	return strings.Join(classes, " ")
}

// getOperatorOptions returns available operators for a filter type
func getOperatorOptions(filterType FilterType) []FilterOption {
	switch filterType {
	case FilterTypeText, FilterTypeSearch:
		return []FilterOption{
			{Value: string(OperatorContains), Label: "Contains", Selected: true},
			{Value: string(OperatorEquals), Label: "Equals"},
			{Value: string(OperatorStartsWith), Label: "Starts with"},
			{Value: string(OperatorEndsWith), Label: "Ends with"},
			{Value: string(OperatorNotEquals), Label: "Not equals"},
		}
	case FilterTypeNumber:
		return []FilterOption{
			{Value: string(OperatorEquals), Label: "Equals", Selected: true},
			{Value: string(OperatorGreaterThan), Label: "Greater than"},
			{Value: string(OperatorLessThan), Label: "Less than"},
			{Value: string(OperatorGreaterEqual), Label: "Greater or equal"},
			{Value: string(OperatorLessEqual), Label: "Less or equal"},
			{Value: string(OperatorBetween), Label: "Between"},
		}
	case FilterTypeDate:
		return []FilterOption{
			{Value: string(OperatorEquals), Label: "On date", Selected: true},
			{Value: string(OperatorGreaterThan), Label: "After"},
			{Value: string(OperatorLessThan), Label: "Before"},
			{Value: string(OperatorBetween), Label: "Between"},
		}
	case FilterTypeDateRange:
		return []FilterOption{
			{Value: string(OperatorBetween), Label: "Between", Selected: true},
		}
	case FilterTypeSelect, FilterTypeCheckbox, FilterTypeRadio:
		return []FilterOption{
			{Value: string(OperatorIn), Label: "Is any of", Selected: true},
			{Value: string(OperatorNotIn), Label: "Is none of"},
		}
	default:
		return []FilterOption{
			{Value: string(OperatorEquals), Label: "Equals", Selected: true},
		}
	}
}

// generateFilterPanelAlpineData creates Alpine.js data object
func generateFilterPanelAlpineData(props FilterPanelProps) string {
	return fmt.Sprintf(`{
		collapsed: %t,
		filters: {},
		presets: %s,
		selectedPreset: null,
		resultCount: 0,
		loading: false,
		
		init() {
			this.initializeFilters();
			if (%t) {
				this.loadState();
			}
		},
		
		initializeFilters() {
			// Initialize filter values from props
			this.filters = this.getDefaultValues();
		},
		
		getDefaultValues() {
			const defaults = {};
			// Set default values for all fields
			return defaults;
		},
		
		toggleCollapsed() {
			this.collapsed = !this.collapsed;
		},
		
		applyFilters() {
			this.loading = true;
			if (%t) {
				this.saveState();
			}
			// HTMX will handle the actual submission
		},
		
		resetFilters() {
			this.filters = this.getDefaultValues();
			this.selectedPreset = null;
			if (%t) {
				this.applyFilters();
			}
		},
		
		clearAllFilters() {
			this.filters = {};
			this.selectedPreset = null;
			if (%t) {
				this.applyFilters();
			}
		},
		
		loadPreset(preset) {
			this.filters = { ...preset.values };
			this.selectedPreset = preset.id;
			if (%t) {
				this.applyFilters();
			}
		},
		
		savePreset(name, description) {
			// Implementation for saving presets
			const preset = {
				id: Date.now().toString(),
				name: name,
				description: description,
				values: { ...this.filters }
			};
			this.presets.push(preset);
		},
		
		deletePreset(presetId) {
			this.presets = this.presets.filter(p => p.id !== presetId);
			if (this.selectedPreset === presetId) {
				this.selectedPreset = null;
			}
		},
		
		onFieldChange(fieldName, value) {
			this.filters[fieldName] = value;
			if (%t) {
				this.debounceApply();
			}
		},
		
		debounceApply() {
			clearTimeout(this.applyTimeout);
			this.applyTimeout = setTimeout(() => {
				this.applyFilters();
			}, 500);
		},
		
		saveState() {
			if (typeof localStorage !== 'undefined') {
				localStorage.setItem('filterState_%s', JSON.stringify({
					filters: this.filters,
					collapsed: this.collapsed,
					selectedPreset: this.selectedPreset
				}));
			}
		},
		
		loadState() {
			if (typeof localStorage !== 'undefined') {
				const saved = localStorage.getItem('filterState_%s');
				if (saved) {
					const state = JSON.parse(saved);
					this.filters = state.filters || {};
					this.collapsed = state.collapsed || false;
					this.selectedPreset = state.selectedPreset || null;
				}
			}
		},
		
		updateResultCount(count) {
			this.resultCount = count;
			this.loading = false;
		},
		
		isFieldVisible(field) {
			if (!field.showWhen) return true;
			// Simple condition evaluation - can be expanded
			return true;
		}
	}`,
		props.Collapsed,
		"[]", // JSON.stringify(props.Presets) - simplified for now
		props.SaveState,
		props.SaveState,
		props.AutoApply,
		props.AutoApply,
		props.AutoApply,
		props.AutoApply,
		props.ID,
		props.ID,
	)
}

// getQuickFilterClasses returns classes for quick filter layout
func getQuickFilterClasses(props QuickFilterProps) string {
	var classes []string

	if props.Horizontal {
		classes = append(classes, "flex flex-wrap items-end")
		if props.Compact {
			classes = append(classes, "gap-2")
		} else {
			classes = append(classes, "gap-4")
		}
	} else {
		classes = append(classes, "space-y-4")
	}

	return strings.Join(classes, " ")
}

// getPresetButtonVariant returns button variant for preset state
func getPresetButtonVariant(isSelected bool) atoms.ButtonVariant {
	if isSelected {
		return atoms.ButtonPrimary
	}
	return atoms.ButtonLight
}

// formatFieldValue formats a filter field value for display
func formatFieldValue(value any, fieldType FilterType) string {
	if value == nil {
		return ""
	}

	switch fieldType {
	case FilterTypeDate:
		// Format date value
		return fmt.Sprintf("%v", value)
	case FilterTypeDateRange:
		// Format date range
		return fmt.Sprintf("%v", value)
	case FilterTypeNumber:
		// Format number
		return fmt.Sprintf("%v", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// getFilterTypeInputProps returns input props for different filter types
func getFilterTypeInputProps(field FilterField) map[string]any {
	props := make(map[string]any)

	props["id"] = field.ID
	props["name"] = field.Name
	props["placeholder"] = field.Placeholder
	props["required"] = field.Required

	switch field.Type {
	case FilterTypeNumber:
		props["type"] = "number"
		if field.Min != nil {
			props["min"] = field.Min
		}
		if field.Max != nil {
			props["max"] = field.Max
		}
	case FilterTypeDate:
		props["type"] = "date"
		if field.Min != nil {
			props["min"] = field.Min
		}
		if field.Max != nil {
			props["max"] = field.Max
		}
	case FilterTypeSearch:
		props["type"] = "search"
	default:
		props["type"] = "text"
	}

	if field.Pattern != "" {
		props["pattern"] = field.Pattern
	}

	return props
}

// shouldShowOperatorSelect determines if operator selection should be shown
func shouldShowOperatorSelect(field FilterField) bool {
	operators := getOperatorOptions(field.Type)
	return len(operators) > 1
}

// getDefaultFilterState returns default filter panel state
func getDefaultFilterState() map[string]any {
	return map[string]any{
		"collapsed":      false,
		"loading":        false,
		"resultCount":    0,
		"selectedPreset": nil,
		"filters":        map[string]any{},
	}
}
