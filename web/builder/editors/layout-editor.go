package editors

import (
	"fmt"

	"github.com/niiniyare/erp/web/builder/core"
)

// LayoutEditor handles grid/flex layout configuration
type LayoutEditor struct {
	component *core.ComponentInstance
}

// LayoutConfiguration represents layout settings
type LayoutConfiguration struct {
	Type       string         `json:"type"` // "grid", "flex", "absolute"
	Properties map[string]any `json:"properties"`
	Responsive map[string]any `json:"responsive,omitempty"`
}

// GridLayoutConfig for CSS Grid layouts
type GridLayoutConfig struct {
	Columns        string `json:"columns"`                  // e.g., "1fr 2fr", "repeat(3, 1fr)"
	Rows           string `json:"rows,omitempty"`           // e.g., "auto 1fr"
	Gap            string `json:"gap"`                      // e.g., "1rem", "10px 20px"
	AlignItems     string `json:"alignItems,omitempty"`     // "start", "end", "center", "stretch"
	JustifyItems   string `json:"justifyItems,omitempty"`   // "start", "end", "center", "stretch"
	AlignContent   string `json:"alignContent,omitempty"`   // "start", "end", "center", "stretch", "space-between"
	JustifyContent string `json:"justifyContent,omitempty"` // "start", "end", "center", "stretch", "space-between"
}

// FlexLayoutConfig for CSS Flexbox layouts
type FlexLayoutConfig struct {
	Direction      string `json:"direction"`      // "row", "column"
	Wrap           string `json:"wrap"`           // "nowrap", "wrap", "wrap-reverse"
	AlignItems     string `json:"alignItems"`     // "flex-start", "flex-end", "center", "stretch"
	JustifyContent string `json:"justifyContent"` // "flex-start", "flex-end", "center", "space-between", "space-around"
	Gap            string `json:"gap,omitempty"`  // "1rem"
}

// NewLayoutEditor creates a new layout editor instance
func NewLayoutEditor(component *core.ComponentInstance) *LayoutEditor {
	return &LayoutEditor{
		component: component,
	}
}

// GetCurrentLayout returns the current layout configuration
func (le *LayoutEditor) GetCurrentLayout() *LayoutConfiguration {
	if le.component.Layout == nil {
		return &LayoutConfiguration{
			Type:       "flex",
			Properties: map[string]any{},
		}
	}

	layoutType := le.detectLayoutType()
	properties := le.extractLayoutProperties(layoutType)

	return &LayoutConfiguration{
		Type:       layoutType,
		Properties: properties,
		Responsive: le.extractResponsiveProperties(),
	}
}

// UpdateLayout applies new layout configuration
func (le *LayoutEditor) UpdateLayout(config *LayoutConfiguration) error {
	if le.component.Layout == nil {
		le.component.Layout = &core.LayoutRules{}
	}

	switch config.Type {
	case "grid":
		return le.applyGridLayout(config.Properties)
	case "flex":
		return le.applyFlexLayout(config.Properties)
	case "absolute":
		return le.applyAbsoluteLayout(config.Properties)
	default:
		return fmt.Errorf("unsupported layout type: %s", config.Type)
	}
}

// GetLayoutPresets returns common layout presets
func (le *LayoutEditor) GetLayoutPresets() []LayoutPreset {
	return []LayoutPreset{
		{
			ID:          "two-column",
			Name:        "Two Column",
			Description: "Two equal columns",
			Type:        "grid",
			Properties: map[string]any{
				"columns": "1fr 1fr",
				"gap":     "1rem",
			},
		},
		{
			ID:          "three-column",
			Name:        "Three Column",
			Description: "Three equal columns",
			Type:        "grid",
			Properties: map[string]any{
				"columns": "repeat(3, 1fr)",
				"gap":     "1rem",
			},
		},
		{
			ID:          "sidebar-main",
			Name:        "Sidebar + Main",
			Description: "Sidebar with main content area",
			Type:        "grid",
			Properties: map[string]any{
				"columns": "250px 1fr",
				"gap":     "1rem",
			},
		},
		{
			ID:          "header-content-footer",
			Name:        "Header/Content/Footer",
			Description: "Header, content, and footer layout",
			Type:        "grid",
			Properties: map[string]any{
				"rows": "auto 1fr auto",
				"gap":  "1rem",
			},
		},
		{
			ID:          "flex-row",
			Name:        "Horizontal Flex",
			Description: "Horizontal flexbox layout",
			Type:        "flex",
			Properties: map[string]any{
				"direction":      "row",
				"alignItems":     "center",
				"justifyContent": "space-between",
				"gap":            "1rem",
			},
		},
		{
			ID:          "flex-column",
			Name:        "Vertical Flex",
			Description: "Vertical flexbox layout",
			Type:        "flex",
			Properties: map[string]any{
				"direction":      "column",
				"alignItems":     "stretch",
				"justifyContent": "flex-start",
				"gap":            "1rem",
			},
		},
	}
}

// LayoutPreset represents a predefined layout configuration
type LayoutPreset struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Properties  map[string]any `json:"properties"`
	Preview     string         `json:"preview,omitempty"`
}

// Helper methods

func (le *LayoutEditor) detectLayoutType() string {
	if le.component.Layout.Grid != nil {
		return "grid"
	}

	if container := le.component.Layout.Container; container != "" {
		if container == "grid" {
			return "grid"
		} else if container == "flex" {
			return "flex"
		}
	}

	// Default to flex
	return "flex"
}

func (le *LayoutEditor) extractLayoutProperties(layoutType string) map[string]any {
	properties := make(map[string]any)

	switch layoutType {
	case "grid":
		if le.component.Layout.Grid != nil {
			grid := le.component.Layout.Grid
			properties["columns"] = grid.Columns
			properties["gap"] = grid.Gap
			if grid.ColumnSpan > 0 {
				properties["columnSpan"] = grid.ColumnSpan
			}
		}
	case "flex":
		// Extract flex properties from classes or styles
		// This would parse CSS classes to determine flex properties
		properties["direction"] = "row"
		properties["alignItems"] = "stretch"
		properties["justifyContent"] = "flex-start"
	}

	return properties
}

func (le *LayoutEditor) extractResponsiveProperties() map[string]any {
	responsive := make(map[string]any)

	if le.component.Layout.Responsive != nil {
		for breakpoint, value := range le.component.Layout.Responsive {
			responsive[breakpoint] = value
		}
	}

	return responsive
}

func (le *LayoutEditor) applyGridLayout(properties map[string]any) error {
	if le.component.Layout.Grid == nil {
		le.component.Layout.Grid = &core.GridLayout{}
	}

	grid := le.component.Layout.Grid

	if columns, ok := properties["columns"].(string); ok {
		grid.Columns = columns
	}

	if gap, ok := properties["gap"].(string); ok {
		grid.Gap = gap
	}

	if columnSpan, ok := properties["columnSpan"].(float64); ok {
		grid.ColumnSpan = int(columnSpan)
	}

	le.component.Layout.Container = "grid"

	return nil
}

func (le *LayoutEditor) applyFlexLayout(properties map[string]any) error {
	le.component.Layout.Container = "flex"

	// Convert flex properties to CSS classes
	classes := []string{}

	if direction, ok := properties["direction"].(string); ok {
		if direction == "column" {
			classes = append(classes, "flex-col")
		} else {
			classes = append(classes, "flex-row")
		}
	}

	if alignItems, ok := properties["alignItems"].(string); ok {
		switch alignItems {
		case "flex-start":
			classes = append(classes, "items-start")
		case "flex-end":
			classes = append(classes, "items-end")
		case "center":
			classes = append(classes, "items-center")
		case "stretch":
			classes = append(classes, "items-stretch")
		}
	}

	if justifyContent, ok := properties["justifyContent"].(string); ok {
		switch justifyContent {
		case "flex-start":
			classes = append(classes, "justify-start")
		case "flex-end":
			classes = append(classes, "justify-end")
		case "center":
			classes = append(classes, "justify-center")
		case "space-between":
			classes = append(classes, "justify-between")
		case "space-around":
			classes = append(classes, "justify-around")
		}
	}

	if gap, ok := properties["gap"].(string); ok {
		// Convert gap to spacing classes
		switch gap {
		case "0.5rem":
			classes = append(classes, "gap-2")
		case "1rem":
			classes = append(classes, "gap-4")
		case "1.5rem":
			classes = append(classes, "gap-6")
		case "2rem":
			classes = append(classes, "gap-8")
		}
	}

	le.component.Layout.Classes = classes

	return nil
}

func (le *LayoutEditor) applyAbsoluteLayout(properties map[string]any) error {
	le.component.Layout.Container = "absolute"

	// Absolute positioning uses component position directly
	// No additional layout configuration needed

	return nil
}

// GetLayoutCSS generates CSS for the current layout
func (le *LayoutEditor) GetLayoutCSS() string {
	config := le.GetCurrentLayout()

	switch config.Type {
	case "grid":
		return le.generateGridCSS(config.Properties)
	case "flex":
		return le.generateFlexCSS(config.Properties)
	case "absolute":
		return le.generateAbsoluteCSS(config.Properties)
	default:
		return ""
	}
}

func (le *LayoutEditor) generateGridCSS(properties map[string]any) string {
	css := "display: grid;\n"

	if columns, ok := properties["columns"].(string); ok {
		css += fmt.Sprintf("grid-template-columns: %s;\n", columns)
	}

	if rows, ok := properties["rows"].(string); ok {
		css += fmt.Sprintf("grid-template-rows: %s;\n", rows)
	}

	if gap, ok := properties["gap"].(string); ok {
		css += fmt.Sprintf("gap: %s;\n", gap)
	}

	if alignItems, ok := properties["alignItems"].(string); ok {
		css += fmt.Sprintf("align-items: %s;\n", alignItems)
	}

	if justifyItems, ok := properties["justifyItems"].(string); ok {
		css += fmt.Sprintf("justify-items: %s;\n", justifyItems)
	}

	return css
}

func (le *LayoutEditor) generateFlexCSS(properties map[string]any) string {
	css := "display: flex;\n"

	if direction, ok := properties["direction"].(string); ok {
		css += fmt.Sprintf("flex-direction: %s;\n", direction)
	}

	if wrap, ok := properties["wrap"].(string); ok {
		css += fmt.Sprintf("flex-wrap: %s;\n", wrap)
	}

	if alignItems, ok := properties["alignItems"].(string); ok {
		css += fmt.Sprintf("align-items: %s;\n", alignItems)
	}

	if justifyContent, ok := properties["justifyContent"].(string); ok {
		css += fmt.Sprintf("justify-content: %s;\n", justifyContent)
	}

	if gap, ok := properties["gap"].(string); ok {
		css += fmt.Sprintf("gap: %s;\n", gap)
	}

	return css
}

func (le *LayoutEditor) generateAbsoluteCSS(properties map[string]any) string {
	return "position: relative;\n"
}
