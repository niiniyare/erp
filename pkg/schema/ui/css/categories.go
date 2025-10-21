package css

import (
	"fmt"
	"sort"
	"strings"
)

// PropertyCategory represents a functional group of CSS properties
type PropertyCategory struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Properties  []string `json:"properties"`
	Icon        string   `json:"icon,omitempty"`     // For UI representation
	Color       string   `json:"color,omitempty"`    // For UI theming
	Priority    int      `json:"priority,omitempty"` // Display order
}

// CategoryManager manages CSS property categorization
type CategoryManager struct {
	categories         map[string]*PropertyCategory
	propertyToCategory map[string]string
}

// NewCategoryManager creates a new CSS property category manager
func NewCategoryManager() *CategoryManager {
	manager := &CategoryManager{
		categories:         make(map[string]*PropertyCategory),
		propertyToCategory: make(map[string]string),
	}

	manager.initializeDefaultCategories()
	return manager
}

// initializeDefaultCategories sets up the default CSS property categories
func (cm *CategoryManager) initializeDefaultCategories() {
	categories := []*PropertyCategory{
		{
			Name:        "Layout",
			Description: "Properties that control element positioning, sizing, and layout behavior",
			Properties: []string{
				"display", "position", "top", "bottom", "left", "right",
				"width", "height", "max-width", "max-height", "min-width", "min-height",
				"float", "clear", "overflow", "overflow-x", "overflow-y",
				"visibility", "z-index", "clip", "clip-path",
			},
			Icon:     "layout",
			Color:    "#3b82f6", // Blue
			Priority: 1,
		},
		{
			Name:        "Flexbox",
			Description: "Properties for flexbox layout system",
			Properties: []string{
				"flex", "flex-direction", "flex-wrap", "flex-flow",
				"flex-grow", "flex-shrink", "flex-basis",
				"justify-content", "align-items", "align-content", "align-self",
				"order", "gap", "row-gap", "column-gap",
			},
			Icon:     "flex",
			Color:    "#8b5cf6", // Purple
			Priority: 2,
		},
		{
			Name:        "Grid",
			Description: "Properties for CSS Grid layout system",
			Properties: []string{
				"grid", "grid-template", "grid-template-areas", "grid-template-rows", "grid-template-columns",
				"grid-auto-rows", "grid-auto-columns", "grid-auto-flow",
				"grid-area", "grid-row", "grid-column", "grid-row-start", "grid-row-end",
				"grid-column-start", "grid-column-end", "grid-gap", "grid-row-gap", "grid-column-gap",
				"justify-items", "place-items", "place-content", "place-self",
			},
			Icon:     "grid",
			Color:    "#10b981", // Green
			Priority: 3,
		},
		{
			Name:        "Spacing",
			Description: "Properties that control margins, padding, and spacing",
			Properties: []string{
				"margin", "margin-top", "margin-bottom", "margin-left", "margin-right",
				"margin-block", "margin-block-start", "margin-block-end",
				"margin-inline", "margin-inline-start", "margin-inline-end",
				"padding", "padding-top", "padding-bottom", "padding-left", "padding-right",
				"padding-block", "padding-block-start", "padding-block-end",
				"padding-inline", "padding-inline-start", "padding-inline-end",
			},
			Icon:     "spacing",
			Color:    "#f59e0b", // Amber
			Priority: 4,
		},
		{
			Name:        "Typography",
			Description: "Properties that control text appearance and layout",
			Properties: []string{
				"font", "font-family", "font-size", "font-weight", "font-style", "font-variant",
				"font-stretch", "font-optical-sizing", "font-kerning", "font-feature-settings",
				"line-height", "letter-spacing", "word-spacing", "text-align", "text-align-last",
				"text-decoration", "text-decoration-line", "text-decoration-style", "text-decoration-color",
				"text-decoration-thickness", "text-underline-offset", "text-underline-position",
				"text-transform", "text-indent", "text-justify", "text-orientation", "text-overflow",
				"text-shadow", "text-rendering", "white-space", "word-break", "word-wrap",
				"hyphens", "tab-size", "direction", "writing-mode",
			},
			Icon:     "type",
			Color:    "#ef4444", // Red
			Priority: 5,
		},
		{
			Name:        "Colors",
			Description: "Properties that control colors and color-related effects",
			Properties: []string{
				"color", "opacity", "color-scheme", "forced-color-adjust",
			},
			Icon:     "palette",
			Color:    "#ec4899", // Pink
			Priority: 6,
		},
		{
			Name:        "Background",
			Description: "Properties that control background appearance",
			Properties: []string{
				"background", "background-color", "background-image", "background-position",
				"background-position-x", "background-position-y", "background-size",
				"background-repeat", "background-attachment", "background-origin",
				"background-clip", "background-blend-mode",
			},
			Icon:     "image",
			Color:    "#06b6d4", // Cyan
			Priority: 7,
		},
		{
			Name:        "Borders",
			Description: "Properties that control borders and outlines",
			Properties: []string{
				"border", "border-width", "border-style", "border-color",
				"border-top", "border-top-width", "border-top-style", "border-top-color",
				"border-bottom", "border-bottom-width", "border-bottom-style", "border-bottom-color",
				"border-left", "border-left-width", "border-left-style", "border-left-color",
				"border-right", "border-right-width", "border-right-style", "border-right-color",
				"border-radius", "border-top-left-radius", "border-top-right-radius",
				"border-bottom-left-radius", "border-bottom-right-radius",
				"border-image", "border-image-source", "border-image-slice", "border-image-width",
				"border-image-outset", "border-image-repeat",
				"outline", "outline-width", "outline-style", "outline-color", "outline-offset",
				"box-decoration-break",
			},
			Icon:     "square",
			Color:    "#84cc16", // Lime
			Priority: 8,
		},
		{
			Name:        "Visual Effects",
			Description: "Properties that create visual effects and transformations",
			Properties: []string{
				"box-shadow", "filter", "backdrop-filter", "mix-blend-mode", "isolation",
				"mask", "mask-image", "mask-mode", "mask-repeat", "mask-position",
				"mask-clip", "mask-origin", "mask-size", "mask-composite",
				"clip-path", "shape-outside", "shape-margin", "shape-image-threshold",
			},
			Icon:     "wand",
			Color:    "#a855f7", // Violet
			Priority: 9,
		},
		{
			Name:        "Transforms",
			Description: "Properties that control 2D and 3D transformations",
			Properties: []string{
				"transform", "transform-origin", "transform-style", "transform-box",
				"perspective", "perspective-origin", "backface-visibility",
			},
			Icon:     "rotate",
			Color:    "#f97316", // Orange
			Priority: 10,
		},
		{
			Name:        "Transitions",
			Description: "Properties that control animated transitions",
			Properties: []string{
				"transition", "transition-property", "transition-duration",
				"transition-timing-function", "transition-delay",
			},
			Icon:     "clock",
			Color:    "#06b6d4", // Cyan
			Priority: 11,
		},
		{
			Name:        "Animations",
			Description: "Properties that control CSS animations",
			Properties: []string{
				"animation", "animation-name", "animation-duration", "animation-timing-function",
				"animation-delay", "animation-iteration-count", "animation-direction",
				"animation-fill-mode", "animation-play-state", "animation-timeline",
				"animation-range", "animation-range-start", "animation-range-end",
			},
			Icon:     "play",
			Color:    "#8b5cf6", // Purple
			Priority: 12,
		},
		{
			Name:        "Lists",
			Description: "Properties specific to list elements",
			Properties: []string{
				"list-style", "list-style-type", "list-style-position", "list-style-image",
				"counter-reset", "counter-increment", "counter-set",
			},
			Icon:     "list",
			Color:    "#64748b", // Slate
			Priority: 13,
		},
		{
			Name:        "Tables",
			Description: "Properties specific to table elements",
			Properties: []string{
				"table-layout", "border-collapse", "border-spacing", "caption-side",
				"empty-cells", "vertical-align",
			},
			Icon:     "table",
			Color:    "#64748b", // Slate
			Priority: 14,
		},
		{
			Name:        "User Interface",
			Description: "Properties that affect user interaction and interface",
			Properties: []string{
				"cursor", "pointer-events", "user-select", "resize", "scroll-behavior",
				"scroll-margin", "scroll-padding", "scroll-snap-type", "scroll-snap-align",
				"scroll-snap-stop", "overscroll-behavior", "overscroll-behavior-x", "overscroll-behavior-y",
				"touch-action", "caret-color", "accent-color",
			},
			Icon:     "cursor",
			Color:    "#6366f1", // Indigo
			Priority: 15,
		},
		{
			Name:        "Print",
			Description: "Properties specific to print media",
			Properties: []string{
				"break-before", "break-after", "break-inside", "page-break-before",
				"page-break-after", "page-break-inside", "orphans", "widows",
			},
			Icon:     "printer",
			Color:    "#64748b", // Slate
			Priority: 16,
		},
		{
			Name:        "Custom Properties",
			Description: "CSS custom properties (variables) and vendor-specific properties",
			Properties: []string{
				"--*", // Pattern for custom properties
			},
			Icon:     "variable",
			Color:    "#64748b", // Slate
			Priority: 17,
		},
		{
			Name:        "Experimental",
			Description: "Experimental and cutting-edge CSS properties",
			Properties: []string{
				"container", "container-type", "container-name",
				"contain", "contain-intrinsic-size", "content-visibility",
				"anchor-name", "position-anchor", "position-area",
			},
			Icon:     "flask",
			Color:    "#64748b", // Slate
			Priority: 18,
		},
	}

	// Add categories and build property-to-category mapping
	for _, category := range categories {
		cm.categories[category.Name] = category
		for _, property := range category.Properties {
			cm.propertyToCategory[property] = category.Name
		}
	}
}

// GetCategory returns a category by name
func (cm *CategoryManager) GetCategory(name string) (*PropertyCategory, bool) {
	category, exists := cm.categories[name]
	return category, exists
}

// GetAllCategories returns all categories sorted by priority
func (cm *CategoryManager) GetAllCategories() []*PropertyCategory {
	categories := make([]*PropertyCategory, 0, len(cm.categories))
	for _, category := range cm.categories {
		categories = append(categories, category)
	}

	// Sort by priority
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Priority < categories[j].Priority
	})

	return categories
}

// GetCategoryNames returns all category names sorted by priority
func (cm *CategoryManager) GetCategoryNames() []string {
	categories := cm.GetAllCategories()
	names := make([]string, len(categories))
	for i, category := range categories {
		names[i] = category.Name
	}
	return names
}

// GetPropertyCategory returns the category name for a CSS property
func (cm *CategoryManager) GetPropertyCategory(propertyName string) string {
	// Normalize property name
	normalized := strings.ToLower(strings.TrimSpace(propertyName))

	// Check for exact match first
	if category, exists := cm.propertyToCategory[normalized]; exists {
		return category
	}

	// Check for custom property pattern
	if strings.HasPrefix(normalized, "--") {
		return "Custom Properties"
	}

	// Check for vendor prefixes - these should always be categorized as Experimental
	if strings.HasPrefix(normalized, "-webkit-") ||
		strings.HasPrefix(normalized, "-moz-") ||
		strings.HasPrefix(normalized, "-ms-") ||
		strings.HasPrefix(normalized, "-o-") {
		return "Experimental"
	}

	// Return unknown category for unrecognized properties
	return "Other"
}

// GetPropertiesByCategory returns all properties in a specific category
func (cm *CategoryManager) GetPropertiesByCategory(categoryName string) []string {
	if category, exists := cm.categories[categoryName]; exists {
		// Return a copy to prevent modification
		properties := make([]string, len(category.Properties))
		copy(properties, category.Properties)
		sort.Strings(properties)
		return properties
	}
	return nil
}

// GetCategorizedProperties returns all properties organized by category
func (cm *CategoryManager) GetCategorizedProperties() map[string][]string {
	result := make(map[string][]string)
	for name, category := range cm.categories {
		properties := make([]string, len(category.Properties))
		copy(properties, category.Properties)
		sort.Strings(properties)
		result[name] = properties
	}
	return result
}

// AddCustomCategory adds a new custom category
func (cm *CategoryManager) AddCustomCategory(category *PropertyCategory) {
	if category == nil || category.Name == "" {
		return
	}

	cm.categories[category.Name] = category
	for _, property := range category.Properties {
		cm.propertyToCategory[property] = category.Name
	}
}

// AddPropertyToCategory adds a property to an existing category
func (cm *CategoryManager) AddPropertyToCategory(propertyName, categoryName string) bool {
	if category, exists := cm.categories[categoryName]; exists {
		// Check if property is already in the category
		for _, prop := range category.Properties {
			if prop == propertyName {
				return true // Already exists
			}
		}

		// Add the property
		category.Properties = append(category.Properties, propertyName)
		cm.propertyToCategory[propertyName] = categoryName
		return true
	}
	return false
}

// GetCategoryStats returns statistics about each category
func (cm *CategoryManager) GetCategoryStats() map[string]CategoryStats {
	stats := make(map[string]CategoryStats)

	for name, category := range cm.categories {
		stats[name] = CategoryStats{
			Name:          name,
			Description:   category.Description,
			PropertyCount: len(category.Properties),
			Priority:      category.Priority,
			Icon:          category.Icon,
			Color:         category.Color,
		}
	}

	return stats
}

// CategoryStats provides statistical information about a category
type CategoryStats struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	PropertyCount int    `json:"property_count"`
	Priority      int    `json:"priority"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
}

// SearchProperties searches for properties across all categories
func (cm *CategoryManager) SearchProperties(query string) map[string][]string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	results := make(map[string][]string)

	for categoryName, category := range cm.categories {
		var matches []string
		for _, property := range category.Properties {
			if strings.Contains(strings.ToLower(property), query) {
				matches = append(matches, property)
			}
		}
		if len(matches) > 0 {
			sort.Strings(matches)
			results[categoryName] = matches
		}
	}

	return results
}

// GetRelatedProperties returns properties related to a given property
func (cm *CategoryManager) GetRelatedProperties(propertyName string) []string {
	categoryName := cm.GetPropertyCategory(propertyName)
	if categoryName == "Other" {
		return nil
	}

	properties := cm.GetPropertiesByCategory(categoryName)
	if properties == nil {
		return nil
	}

	// Remove the original property from the results
	var related []string
	normalized := strings.ToLower(propertyName)
	for _, prop := range properties {
		if strings.ToLower(prop) != normalized {
			related = append(related, prop)
		}
	}

	return related
}

// ValidateCategoryStructure validates that all categories have required fields
func (cm *CategoryManager) ValidateCategoryStructure() []string {
	var issues []string

	for name, category := range cm.categories {
		if category.Name != name {
			issues = append(issues, fmt.Sprintf("Category %s has mismatched name field", name))
		}

		if category.Description == "" {
			issues = append(issues, fmt.Sprintf("Category %s missing description", name))
		}

		if len(category.Properties) == 0 {
			issues = append(issues, fmt.Sprintf("Category %s has no properties", name))
		}

		if category.Priority == 0 {
			issues = append(issues, fmt.Sprintf("Category %s has no priority set", name))
		}
	}

	return issues
}

// ExportCategoriesJSON exports all categories as JSON-serializable data
func (cm *CategoryManager) ExportCategoriesJSON() map[string]any {
	return map[string]any{
		"categories":       cm.GetAllCategories(),
		"stats":            cm.GetCategoryStats(),
		"total_categories": len(cm.categories),
		"total_properties": len(cm.propertyToCategory),
	}
}
