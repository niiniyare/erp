package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Schema represents a JSON schema with its metadata
type Schema struct {
	Name         string     `json:"name"`
	Path         string     `json:"path"`
	Category     string     `json:"category"`
	Subcategory  string     `json:"subcategory"`
	Meta         SchemaMeta `json:"meta"`
	Dependencies []string   `json:"dependencies"`
	Size         int64      `json:"size"`
	LastModified time.Time  `json:"lastModified"`
}

// SchemaMeta represents the metadata structure
type SchemaMeta struct {
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	Category     string           `json:"category"`
	Subcategory  string           `json:"subcategory"`
	Dependencies []string         `json:"dependencies"`
	Examples     []string         `json:"examples"`
	Version      string           `json:"version"`
	Deprecated   bool             `json:"deprecated"`
	Tags         []string         `json:"tags"`
	Framework    FrameworkSupport `json:"framework"`
	Testing      TestingInfo      `json:"testing"`
}

type FrameworkSupport struct {
	Templ  TemplSupport  `json:"templ"`
	HTMX   HTMXSupport   `json:"htmx"`
	Alpine AlpineSupport `json:"alpine"`
}

type TemplSupport struct {
	ComponentFile  string `json:"componentFile"`
	PropsInterface string `json:"propsInterface"`
}

type HTMXSupport struct {
	Supported  bool     `json:"supported"`
	Attributes []string `json:"attributes"`
}

type AlpineSupport struct {
	Directives []string `json:"directives"`
}

type TestingInfo struct {
	UnitTests        string  `json:"unitTests"`
	IntegrationTests string  `json:"integrationTests"`
	VisualTests      string  `json:"visualTests"`
	TestCoverage     float64 `json:"testCoverage"`
}

// SchemaRegistry holds all discovered schemas
type SchemaRegistry struct {
	Schemas       []Schema            `json:"schemas"`
	Categories    map[string]int      `json:"categories"`
	Subcategories map[string]int      `json:"subcategories"`
	TotalSchemas  int                 `json:"totalSchemas"`
	LastUpdated   time.Time           `json:"lastUpdated"`
	Dependencies  map[string][]string `json:"dependencies"`
}

func main() {
	fmt.Println("🔍 Starting schema discovery...")

	registry := &SchemaRegistry{
		Schemas:       []Schema{},
		Categories:    make(map[string]int),
		Subcategories: make(map[string]int),
		Dependencies:  make(map[string][]string),
		LastUpdated:   time.Now(),
	}

	// Walk through all schema files
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip metadata files and directories
		if d.IsDir() || !strings.HasSuffix(path, ".json") ||
			strings.HasSuffix(path, ".meta.json") ||
			strings.Contains(path, "schema_metadata_template") {
			return nil
		}

		schema, err := processSchema(path)
		if err != nil {
			fmt.Printf("❌ Error processing %s: %v\n", path, err)
			return nil
		}

		registry.Schemas = append(registry.Schemas, *schema)
		registry.Categories[schema.Category]++
		registry.Subcategories[schema.Subcategory]++

		if len(schema.Dependencies) > 0 {
			registry.Dependencies[schema.Name] = schema.Dependencies
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	registry.TotalSchemas = len(registry.Schemas)

	// Sort schemas by category then name
	sort.Slice(registry.Schemas, func(i, j int) bool {
		if registry.Schemas[i].Category == registry.Schemas[j].Category {
			return registry.Schemas[i].Name < registry.Schemas[j].Name
		}
		return registry.Schemas[i].Category < registry.Schemas[j].Category
	})

	// Generate outputs
	generateRegistryJSON(registry)
	generateMarkdownReport(registry)
	generateDependencyGraph(registry)

	fmt.Printf("✅ Schema discovery complete! Found %d schemas\n", registry.TotalSchemas)
	printSummary(registry)
}

func processSchema(path string) (*Schema, error) {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSuffix(filepath.Base(path), ".json")
	category, subcategory := inferCategoryFromPath(path)

	schema := &Schema{
		Name:         name,
		Path:         path,
		Category:     category,
		Subcategory:  subcategory,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		Dependencies: []string{},
	}

	// Try to load metadata if it exists
	metaPath := strings.TrimSuffix(path, ".json") + ".meta.json"
	if _, err := os.Stat(metaPath); err == nil {
		metaData, err := os.ReadFile(metaPath)
		if err == nil {
			var metaContainer struct {
				Meta SchemaMeta `json:"meta"`
			}
			if json.Unmarshal(metaData, &metaContainer) == nil {
				schema.Meta = metaContainer.Meta
				schema.Dependencies = metaContainer.Meta.Dependencies
			}
		}
	}

	// If no metadata, create basic info
	if schema.Meta.Title == "" {
		schema.Meta = SchemaMeta{
			Title:       generateTitle(name),
			Description: generateDescription(name, category),
			Category:    category,
			Subcategory: subcategory,
			Version:     "1.0.0",
			Deprecated:  false,
			Tags:        []string{strings.ToLower(subcategory), strings.ToLower(name)},
		}
	}

	return schema, nil
}

func inferCategoryFromPath(path string) (string, string) {
	if strings.Contains(path, "core/layout") {
		return "core", "layout"
	} else if strings.Contains(path, "core/typography") {
		return "core", "typography"
	} else if strings.Contains(path, "core/color") {
		return "core", "color"
	} else if strings.Contains(path, "core/spacing") {
		return "core", "spacing"
	} else if strings.Contains(path, "core/datatypes") {
		return "core", "datatypes"
	} else if strings.Contains(path, "core/compatibility") {
		return "core", "compatibility"
	} else if strings.Contains(path, "components/atoms") {
		return "atoms", "forms"
	} else if strings.Contains(path, "components/molecules") {
		return "molecules", "forms"
	} else if strings.Contains(path, "components/organisms") {
		return "organisms", "forms"
	} else if strings.Contains(path, "components/templates") {
		return "templates", "layouts"
	} else if strings.Contains(path, "interactions/forms") {
		return "interactions", "forms"
	} else if strings.Contains(path, "interactions/navigation") {
		return "interactions", "navigation"
	} else if strings.Contains(path, "interactions/data") {
		return "interactions", "data"
	}
	return "utility", "helpers"
}

func generateTitle(name string) string {
	// Remove Schema suffix and add spaces before capitals
	name = strings.TrimSuffix(name, "Schema")
	return addSpaces(name)
}

func generateDescription(name, category string) string {
	switch category {
	case "atoms":
		return fmt.Sprintf("A fundamental UI element: %s", name)
	case "molecules":
		return fmt.Sprintf("A composite UI component: %s", name)
	case "organisms":
		return fmt.Sprintf("A complex UI component: %s", name)
	case "templates":
		return fmt.Sprintf("A page-level template: %s", name)
	case "interactions":
		return fmt.Sprintf("Interactive behavior schema: %s", name)
	case "core":
		return fmt.Sprintf("Core CSS property definition: %s", name)
	default:
		return fmt.Sprintf("Schema definition for %s", name)
	}
}

func addSpaces(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return result.String()
}

func generateRegistryJSON(registry *SchemaRegistry) {
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		fmt.Printf("Error generating registry JSON: %v\n", err)
		return
	}

	err = os.WriteFile("schema_registry.json", data, 0o644)
	if err != nil {
		fmt.Printf("Error writing registry JSON: %v\n", err)
		return
	}

	fmt.Println("📋 Generated schema_registry.json")
}

func generateMarkdownReport(registry *SchemaRegistry) {
	var md strings.Builder

	md.WriteString("# Schema Discovery Report\n\n")
	md.WriteString(fmt.Sprintf("**Generated:** %s  \n", registry.LastUpdated.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Total Schemas:** %d\n\n", registry.TotalSchemas))

	// Category breakdown
	md.WriteString("## Category Breakdown\n\n")
	md.WriteString("| Category | Count | Percentage |\n")
	md.WriteString("|----------|-------|------------|\n")

	for category, count := range registry.Categories {
		percentage := float64(count) / float64(registry.TotalSchemas) * 100
		md.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", category, count, percentage))
	}

	md.WriteString("\n## Schemas by Category\n\n")

	currentCategory := ""
	for _, schema := range registry.Schemas {
		if schema.Category != currentCategory {
			currentCategory = schema.Category
			md.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(currentCategory)))
		}

		deprecated := ""
		if schema.Meta.Deprecated {
			deprecated = " ⚠️ *Deprecated*"
		}

		md.WriteString(fmt.Sprintf("- **%s**%s - %s\n", schema.Meta.Title, deprecated, schema.Meta.Description))
		if len(schema.Meta.Tags) > 0 {
			md.WriteString(fmt.Sprintf("  - *Tags:* %s\n", strings.Join(schema.Meta.Tags, ", ")))
		}
	}

	err := os.WriteFile("SCHEMA_DISCOVERY_REPORT.md", []byte(md.String()), 0o644)
	if err != nil {
		fmt.Printf("Error writing markdown report: %v\n", err)
		return
	}

	fmt.Println("📊 Generated SCHEMA_DISCOVERY_REPORT.md")
}

func generateDependencyGraph(registry *SchemaRegistry) {
	var dot strings.Builder

	dot.WriteString("digraph SchemaRegistry {\n")
	dot.WriteString("  rankdir=TB;\n")
	dot.WriteString("  node [shape=box, style=filled];\n\n")

	// Color nodes by category
	colors := map[string]string{
		"core":         "#FFE5B4", // Light orange
		"atoms":        "#B4D4FF", // Light blue
		"molecules":    "#B4FFB4", // Light green
		"organisms":    "#FFB4B4", // Light red
		"templates":    "#E5B4FF", // Light purple
		"interactions": "#FFD4B4", // Light peach
		"utility":      "#D4D4D4", // Light gray
	}

	// Add nodes
	for _, schema := range registry.Schemas {
		color, exists := colors[schema.Category]
		if !exists {
			color = "#FFFFFF"
		}

		label := strings.ReplaceAll(schema.Name, "\"", "\\\"")
		dot.WriteString(fmt.Sprintf("  \"%s\" [fillcolor=\"%s\", label=\"%s\"];\n",
			schema.Name, color, label))
	}

	dot.WriteString("\n")

	// Add dependencies
	for schema, deps := range registry.Dependencies {
		for _, dep := range deps {
			dot.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", schema, dep))
		}
	}

	dot.WriteString("}\n")

	err := os.WriteFile("schema_dependencies.dot", []byte(dot.String()), 0o644)
	if err != nil {
		fmt.Printf("Error writing dependency graph: %v\n", err)
		return
	}

	fmt.Println("🔗 Generated schema_dependencies.dot (use Graphviz to render)")
}

func printSummary(registry *SchemaRegistry) {
	fmt.Println("\n📊 Discovery Summary:")
	fmt.Printf("   • Total Schemas: %d\n", registry.TotalSchemas)
	fmt.Printf("   • Categories: %d\n", len(registry.Categories))
	fmt.Printf("   • Subcategories: %d\n", len(registry.Subcategories))
	fmt.Printf("   • With Dependencies: %d\n", len(registry.Dependencies))

	fmt.Println("\n📂 Category Breakdown:")
	for category, count := range registry.Categories {
		percentage := float64(count) / float64(registry.TotalSchemas) * 100
		fmt.Printf("   • %-12s: %3d schemas (%.1f%%)\n", category, count, percentage)
	}
}
