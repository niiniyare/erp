package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/niiniyare/erp/pkg/schema/ui/css"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	factory := css.NewFactory("") // Empty schema dir for this demo

	switch command {
	case "categories":
		listCategories(factory)
	case "properties":
		if len(os.Args) < 3 {
			fmt.Println("Usage: css-categories properties <category-name>")
			return
		}
		listPropertiesInCategory(factory, os.Args[2])
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Usage: css-categories search <query>")
			return
		}
		searchProperties(factory, os.Args[2])
	case "categorize":
		if len(os.Args) < 3 {
			fmt.Println("Usage: css-categories categorize <property-name>")
			return
		}
		categorizeProperty(factory, os.Args[2])
	case "related":
		if len(os.Args) < 3 {
			fmt.Println("Usage: css-categories related <property-name>")
			return
		}
		findRelatedProperties(factory, os.Args[2])
	case "stats":
		showCategoryStats(factory)
	case "export":
		exportCategories(factory)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("CSS Property Categorization Tool")
	fmt.Println("")
	fmt.Println("Usage: css-categories <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  categories                    - List all CSS property categories")
	fmt.Println("  properties <category>         - List properties in a specific category")
	fmt.Println("  search <query>               - Search for properties across categories")
	fmt.Println("  categorize <property>        - Show which category a property belongs to")
	fmt.Println("  related <property>           - Find properties related to a given property")
	fmt.Println("  stats                        - Show statistics for all categories")
	fmt.Println("  export                       - Export all categories as JSON")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  css-categories categories")
	fmt.Println("  css-categories properties Layout")
	fmt.Println("  css-categories search margin")
	fmt.Println("  css-categories categorize display")
	fmt.Println("  css-categories related font-size")
	fmt.Println("  css-categories stats")
}

func listCategories(factory *css.Factory) {
	categories := factory.GetAllCategories()
	
	fmt.Printf("📋 CSS Property Categories (%d total)\n", len(categories))
	fmt.Println(strings.Repeat("=", 60))
	
	for _, category := range categories {
		icon := "📁"
		if category.Icon != "" {
			// Map some common icons for better display
			switch category.Icon {
			case "layout":
				icon = "🏗️ "
			case "flex":
				icon = "🔄"
			case "grid":
				icon = "📊"
			case "spacing":
				icon = "📏"
			case "type":
				icon = "🔤"
			case "palette":
				icon = "🎨"
			case "image":
				icon = "🖼️ "
			case "square":
				icon = "⬜"
			case "wand":
				icon = "✨"
			case "rotate":
				icon = "🔄"
			case "clock":
				icon = "⏰"
			case "play":
				icon = "▶️ "
			}
		}
		
		fmt.Printf("%s %s (%d properties)\n", icon, category.Name, len(category.Properties))
		fmt.Printf("   %s\n", category.Description)
		if category.Color != "" {
			fmt.Printf("   Color: %s\n", category.Color)
		}
		fmt.Println()
	}
}

func listPropertiesInCategory(factory *css.Factory, categoryName string) {
	properties := factory.GetPropertiesByCategory(categoryName)
	if properties == nil {
		fmt.Printf("❌ Category '%s' not found\n", categoryName)
		return
	}

	category, exists := factory.GetPropertyRegistry().GetCategory(categoryName)
	if !exists {
		fmt.Printf("❌ Category '%s' not found\n", categoryName)
		return
	}

	fmt.Printf("📦 %s Properties (%d total)\n", categoryName, len(properties))
	fmt.Printf("📝 %s\n", category.Description)
	fmt.Println(strings.Repeat("=", 50))
	
	// Group properties in columns for better display
	const colWidth = 25
	const cols = 3
	
	for i := 0; i < len(properties); i += cols {
		for j := 0; j < cols && i+j < len(properties); j++ {
			prop := properties[i+j]
			fmt.Printf("%-*s", colWidth, prop)
		}
		fmt.Println()
	}
}

func searchProperties(factory *css.Factory, query string) {
	results := factory.SearchProperties(query)
	if len(results) == 0 {
		fmt.Printf("🔍 No properties found matching '%s'\n", query)
		return
	}

	fmt.Printf("🔍 Search Results for '%s'\n", query)
	fmt.Println(strings.Repeat("=", 40))
	
	for categoryName, properties := range results {
		fmt.Printf("\n📂 %s (%d matches):\n", categoryName, len(properties))
		for _, prop := range properties {
			fmt.Printf("  • %s\n", prop)
		}
	}
}

func categorizeProperty(factory *css.Factory, propertyName string) {
	category := factory.GetPropertyCategory(propertyName)
	
	fmt.Printf("🏷️  Property Categorization\n")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Printf("Property: %s\n", propertyName)
	fmt.Printf("Category: %s\n", category)
	
	if category != "Other" {
		categoryInfo, exists := factory.GetPropertyRegistry().GetCategory(category)
		if exists {
			fmt.Printf("Description: %s\n", categoryInfo.Description)
			fmt.Printf("Total properties in category: %d\n", len(categoryInfo.Properties))
		}
	}
}

func findRelatedProperties(factory *css.Factory, propertyName string) {
	related := factory.GetRelatedProperties(propertyName)
	category := factory.GetPropertyCategory(propertyName)
	
	fmt.Printf("🔗 Related Properties for '%s'\n", propertyName)
	fmt.Printf("📂 Category: %s\n", category)
	fmt.Println(strings.Repeat("=", 40))
	
	if len(related) == 0 {
		fmt.Println("No related properties found")
		return
	}
	
	// Show up to 20 related properties
	maxShow := 20
	if len(related) > maxShow {
		related = related[:maxShow]
	}
	
	for i, prop := range related {
		fmt.Printf("%2d. %s\n", i+1, prop)
	}
	
	if len(factory.GetRelatedProperties(propertyName)) > maxShow {
		remaining := len(factory.GetRelatedProperties(propertyName)) - maxShow
		fmt.Printf("... and %d more properties\n", remaining)
	}
}

func showCategoryStats(factory *css.Factory) {
	stats := factory.GetCategoryStats()
	if len(stats) == 0 {
		fmt.Println("❌ No category statistics available")
		return
	}

	fmt.Printf("📊 CSS Category Statistics (%d categories)\n", len(stats))
	fmt.Println(strings.Repeat("=", 60))
	
	// Convert to slice and sort by property count (descending)
	type statEntry struct {
		name  string
		stats css.CategoryStats
	}
	
	var entries []statEntry
	for name, stat := range stats {
		entries = append(entries, statEntry{name: name, stats: stat})
	}
	
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].stats.PropertyCount > entries[j].stats.PropertyCount
	})
	
	fmt.Printf("%-20s %s %s\n", "Category", "Properties", "Description")
	fmt.Println(strings.Repeat("-", 60))
	
	totalProperties := 0
	for _, entry := range entries {
		fmt.Printf("%-20s %10d  %s\n", 
			entry.name, 
			entry.stats.PropertyCount,
			truncateString(entry.stats.Description, 30))
		totalProperties += entry.stats.PropertyCount
	}
	
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-20s %10d\n", "TOTAL", totalProperties)
}

func exportCategories(factory *css.Factory) {
	categories := factory.GetAllCategories()
	stats := factory.GetCategoryStats()
	
	exportData := map[string]any{
		"metadata": map[string]any{
			"total_categories": len(categories),
			"generated_by":     "CSS Categories Tool",
			"version":          "1.0",
		},
		"categories": categories,
		"statistics": stats,
	}
	
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error exporting categories: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}