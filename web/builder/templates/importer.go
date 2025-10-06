package templates

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/niiniyare/erp/web/builder/core"
)

// Importer handles importing templates from various sources
type Importer struct {
	library  *TemplateLibrary
	registry *core.ComponentRegistry
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	SuccessCount  int             `json:"successCount"`
	ErrorCount    int             `json:"errorCount"`
	WarningCount  int             `json:"warningCount"`
	ImportedIDs   []string        `json:"importedIds"`
	Errors        []ImportError   `json:"errors"`
	Warnings      []ImportWarning `json:"warnings"`
	Summary       string          `json:"summary"`
	Duration      time.Duration   `json:"duration"`
	ConflictCount int             `json:"conflictCount"`
	ResolvedCount int             `json:"resolvedCount"`
}

// ImportError represents an import error
type ImportError struct {
	Source    string `json:"source"`
	Message   string `json:"message"`
	Line      int    `json:"line,omitempty"`
	Component string `json:"component,omitempty"`
	Type      string `json:"type"` // "validation", "parsing", "conflict", "dependency"
}

// ImportWarning represents an import warning
type ImportWarning struct {
	Source  string `json:"source"`
	Message string `json:"message"`
	Type    string `json:"type"` // "deprecated", "compatibility", "performance"
}

// ImportOptions configures import behavior
type ImportOptions struct {
	OverwriteExisting  bool                   `json:"overwriteExisting"`
	SkipDuplicates     bool                   `json:"skipDuplicates"`
	ValidateComponents bool                   `json:"validateComponents"`
	CreateMissing      bool                   `json:"createMissing"`
	ConflictResolution string                 `json:"conflictResolution"` // "skip", "overwrite", "rename", "prompt"
	ComponentMapping   map[string]string      `json:"componentMapping"`   // Map old component types to new ones
	DefaultAuthor      string                 `json:"defaultAuthor"`
	DefaultCategory    string                 `json:"defaultCategory"`
	TagPrefix          string                 `json:"tagPrefix"`
	VariableDefaults   map[string]interface{} `json:"variableDefaults"`
}

// Phase43Schema represents a Phase 4.3 auto-generated schema
type Phase43Schema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"createdAt"`
	Components  []Phase43Component     `json:"components"`
	Layout      Phase43Layout          `json:"layout"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tables      []Phase43Table         `json:"tables,omitempty"`
	Forms       []Phase43Form          `json:"forms,omitempty"`
	Navigation  []Phase43Navigation    `json:"navigation,omitempty"`
}

// Phase43Component represents a component in Phase 4.3 format
type Phase43Component struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Props      map[string]interface{} `json:"props"`
	Children   []Phase43Component     `json:"children,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Events     map[string]interface{} `json:"events,omitempty"`
}

// Phase43Layout represents layout configuration in Phase 4.3 format
type Phase43Layout struct {
	Type      string                 `json:"type"`
	Container string                 `json:"container"`
	Grid      map[string]interface{} `json:"grid,omitempty"`
	Flex      map[string]interface{} `json:"flex,omitempty"`
}

// Phase43Table represents table structure from Phase 4.3
type Phase43Table struct {
	Name    string                   `json:"name"`
	Columns []Phase43TableColumn     `json:"columns"`
	Data    []map[string]interface{} `json:"data,omitempty"`
	Config  map[string]interface{}   `json:"config"`
}

// Phase43TableColumn represents a table column
type Phase43TableColumn struct {
	Key        string      `json:"key"`
	Title      string      `json:"title"`
	Type       string      `json:"type"`
	Sortable   bool        `json:"sortable"`
	Filterable bool        `json:"filterable"`
	Width      string      `json:"width,omitempty"`
	Format     string      `json:"format,omitempty"`
	Default    interface{} `json:"default,omitempty"`
}

// Phase43Form represents form structure from Phase 4.3
type Phase43Form struct {
	Name       string                 `json:"name"`
	Fields     []Phase43FormField     `json:"fields"`
	Validation map[string]interface{} `json:"validation"`
	Actions    []Phase43FormAction    `json:"actions"`
}

// Phase43FormField represents a form field
type Phase43FormField struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Label       string                 `json:"label"`
	Required    bool                   `json:"required"`
	Placeholder string                 `json:"placeholder,omitempty"`
	Options     []Option               `json:"options,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

// Phase43FormAction represents a form action
type Phase43FormAction struct {
	Type   string `json:"type"`
	Label  string `json:"label"`
	Target string `json:"target"`
	Method string `json:"method,omitempty"`
}

// Phase43Navigation represents navigation structure
type Phase43Navigation struct {
	Type   string                 `json:"type"`
	Items  []Phase43NavItem       `json:"items"`
	Config map[string]interface{} `json:"config"`
}

// Phase43NavItem represents a navigation item
type Phase43NavItem struct {
	Label    string           `json:"label"`
	Icon     string           `json:"icon,omitempty"`
	URL      string           `json:"url,omitempty"`
	Children []Phase43NavItem `json:"children,omitempty"`
	Roles    []string         `json:"roles,omitempty"`
}

// TemplFile represents a Templ template file
type TemplFile struct {
	Path       string                 `json:"path"`
	Package    string                 `json:"package"`
	Templates  []TemplTemplate        `json:"templates"`
	Imports    []string               `json:"imports"`
	Components []TemplComponent       `json:"components"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TemplTemplate represents a Templ template
type TemplTemplate struct {
	Name       string                 `json:"name"`
	Parameters []TemplParameter       `json:"parameters"`
	Body       string                 `json:"body"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TemplParameter represents a template parameter
type TemplParameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TemplComponent represents a component extracted from Templ
type TemplComponent struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Props    map[string]string `json:"props"`
	Children []TemplComponent  `json:"children"`
}

// NewImporter creates a new template importer
func NewImporter(library *TemplateLibrary, registry *core.ComponentRegistry) *Importer {
	return &Importer{
		library:  library,
		registry: registry,
	}
}

// ImportFromDirectory imports templates from a directory
func (imp *Importer) ImportFromDirectory(dirPath string, options *ImportOptions) (*ImportResult, error) {
	startTime := time.Now()
	result := &ImportResult{
		ImportedIDs: make([]string, 0),
		Errors:      make([]ImportError, 0),
		Warnings:    make([]ImportWarning, 0),
	}

	// Walk through directory
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Source:  path,
				Message: fmt.Sprintf("Error accessing file: %v", err),
				Type:    "filesystem",
			})
			return nil // Continue walking
		}

		if d.IsDir() {
			return nil
		}

		// Process different file types
		switch {
		case strings.HasSuffix(path, ".json"):
			imp.importJSONFile(path, options, result)
		case strings.HasSuffix(path, ".templ"):
			imp.importTemplFile(path, options, result)
		case strings.HasSuffix(path, ".schema.json"):
			imp.importPhase43Schema(path, options, result)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	result.Duration = time.Since(startTime)
	result.Summary = fmt.Sprintf("Imported %d templates, %d errors, %d warnings",
		result.SuccessCount, result.ErrorCount, result.WarningCount)

	return result, nil
}

// ImportPhase43Schema imports a single Phase 4.3 schema
func (imp *Importer) ImportPhase43Schema(schemaData []byte, options *ImportOptions) (*ImportResult, error) {
	startTime := time.Now()
	result := &ImportResult{
		ImportedIDs: make([]string, 0),
		Errors:      make([]ImportError, 0),
		Warnings:    make([]ImportWarning, 0),
	}

	var phase43Schema Phase43Schema
	if err := json.Unmarshal(schemaData, &phase43Schema); err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  "schema",
			Message: fmt.Sprintf("Failed to parse schema: %v", err),
			Type:    "parsing",
		})
		result.ErrorCount++
		return result, nil
	}

	// Convert Phase 4.3 schema to template
	template, err := imp.convertPhase43Schema(&phase43Schema, options)
	if err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  phase43Schema.Name,
			Message: fmt.Sprintf("Failed to convert schema: %v", err),
			Type:    "conversion",
		})
		result.ErrorCount++
		return result, nil
	}

	// Add template to library
	if err := imp.library.AddTemplate(template); err != nil {
		if options.OverwriteExisting && strings.Contains(err.Error(), "already exists") {
			// Attempt to update existing template
			template.UpdatedAt = time.Now()
			if err := imp.library.AddTemplate(template); err != nil {
				result.Errors = append(result.Errors, ImportError{
					Source:  template.ID,
					Message: fmt.Sprintf("Failed to add template: %v", err),
					Type:    "validation",
				})
				result.ErrorCount++
				return result, nil
			}
		} else {
			result.Errors = append(result.Errors, ImportError{
				Source:  template.ID,
				Message: fmt.Sprintf("Failed to add template: %v", err),
				Type:    "validation",
			})
			result.ErrorCount++
			return result, nil
		}
	}

	result.ImportedIDs = append(result.ImportedIDs, template.ID)
	result.SuccessCount++
	result.Duration = time.Since(startTime)
	result.Summary = fmt.Sprintf("Successfully imported schema: %s", template.Name)

	return result, nil
}

// ImportTemplFile imports templates from a Templ file
func (imp *Importer) ImportTemplFile(filePath string, options *ImportOptions) (*ImportResult, error) {
	startTime := time.Now()
	result := &ImportResult{
		ImportedIDs: make([]string, 0),
		Errors:      make([]ImportError, 0),
		Warnings:    make([]ImportWarning, 0),
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  filePath,
			Message: fmt.Sprintf("Failed to read file: %v", err),
			Type:    "filesystem",
		})
		result.ErrorCount++
		return result, nil
	}

	// Parse Templ file
	templFile, err := imp.parseTemplFile(string(content))
	if err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  filePath,
			Message: fmt.Sprintf("Failed to parse Templ file: %v", err),
			Type:    "parsing",
		})
		result.ErrorCount++
		return result, nil
	}

	// Convert each template
	for _, templTemplate := range templFile.Templates {
		template, err := imp.convertTemplTemplate(&templTemplate, templFile, options)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Source:    filePath,
				Message:   fmt.Sprintf("Failed to convert template %s: %v", templTemplate.Name, err),
				Type:      "conversion",
				Component: templTemplate.Name,
			})
			result.ErrorCount++
			continue
		}

		// Add template to library
		if err := imp.library.AddTemplate(template); err != nil {
			if options.OverwriteExisting {
				template.UpdatedAt = time.Now()
				if err := imp.library.AddTemplate(template); err != nil {
					result.Errors = append(result.Errors, ImportError{
						Source:    filePath,
						Message:   fmt.Sprintf("Failed to add template %s: %v", template.Name, err),
						Type:      "validation",
						Component: templTemplate.Name,
					})
					result.ErrorCount++
					continue
				}
			} else {
				result.Errors = append(result.Errors, ImportError{
					Source:    filePath,
					Message:   fmt.Sprintf("Failed to add template %s: %v", template.Name, err),
					Type:      "validation",
					Component: templTemplate.Name,
				})
				result.ErrorCount++
				continue
			}
		}

		result.ImportedIDs = append(result.ImportedIDs, template.ID)
		result.SuccessCount++
	}

	result.Duration = time.Since(startTime)
	result.Summary = fmt.Sprintf("Imported %d templates from %s", result.SuccessCount, filepath.Base(filePath))

	return result, nil
}

// Helper methods for importing specific file types

func (imp *Importer) importJSONFile(path string, options *ImportOptions, result *ImportResult) {
	content, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  path,
			Message: fmt.Sprintf("Failed to read file: %v", err),
			Type:    "filesystem",
		})
		result.ErrorCount++
		return
	}

	// Try to parse as template
	var template Template
	if err := json.Unmarshal(content, &template); err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  path,
			Message: fmt.Sprintf("Failed to parse JSON template: %v", err),
			Type:    "parsing",
		})
		result.ErrorCount++
		return
	}

	// Validate and add template
	if err := imp.library.AddTemplate(&template); err != nil {
		if options.OverwriteExisting {
			template.UpdatedAt = time.Now()
			if err := imp.library.AddTemplate(&template); err != nil {
				result.Errors = append(result.Errors, ImportError{
					Source:  path,
					Message: fmt.Sprintf("Failed to add template: %v", err),
					Type:    "validation",
				})
				result.ErrorCount++
				return
			}
		} else {
			result.Errors = append(result.Errors, ImportError{
				Source:  path,
				Message: fmt.Sprintf("Failed to add template: %v", err),
				Type:    "validation",
			})
			result.ErrorCount++
			return
		}
	}

	result.ImportedIDs = append(result.ImportedIDs, template.ID)
	result.SuccessCount++
}

func (imp *Importer) importTemplFile(path string, options *ImportOptions, result *ImportResult) {
	templResult, err := imp.ImportTemplFile(path, options)
	if err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  path,
			Message: fmt.Sprintf("Failed to import Templ file: %v", err),
			Type:    "import",
		})
		result.ErrorCount++
		return
	}

	// Merge results
	result.SuccessCount += templResult.SuccessCount
	result.ErrorCount += templResult.ErrorCount
	result.WarningCount += templResult.WarningCount
	result.ImportedIDs = append(result.ImportedIDs, templResult.ImportedIDs...)
	result.Errors = append(result.Errors, templResult.Errors...)
	result.Warnings = append(result.Warnings, templResult.Warnings...)
}

func (imp *Importer) importPhase43Schema(path string, options *ImportOptions, result *ImportResult) {
	content, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  path,
			Message: fmt.Sprintf("Failed to read file: %v", err),
			Type:    "filesystem",
		})
		result.ErrorCount++
		return
	}

	schemaResult, err := imp.ImportPhase43Schema(content, options)
	if err != nil {
		result.Errors = append(result.Errors, ImportError{
			Source:  path,
			Message: fmt.Sprintf("Failed to import Phase 4.3 schema: %v", err),
			Type:    "import",
		})
		result.ErrorCount++
		return
	}

	// Merge results
	result.SuccessCount += schemaResult.SuccessCount
	result.ErrorCount += schemaResult.ErrorCount
	result.WarningCount += schemaResult.WarningCount
	result.ImportedIDs = append(result.ImportedIDs, schemaResult.ImportedIDs...)
	result.Errors = append(result.Errors, schemaResult.Errors...)
	result.Warnings = append(result.Warnings, schemaResult.Warnings...)
}

// Conversion methods

func (imp *Importer) convertPhase43Schema(phase43Schema *Phase43Schema, options *ImportOptions) (*Template, error) {
	// Create template ID
	templateID := strings.ToLower(strings.ReplaceAll(phase43Schema.Name, " ", "-"))
	if options.TagPrefix != "" {
		templateID = options.TagPrefix + "-" + templateID
	}

	// Convert to core schema
	schema := &core.CompositionSchema{
		ID:          templateID + "-schema",
		Name:        phase43Schema.Name,
		Description: phase43Schema.Description,
		Components:  []core.ComponentInstance{},
	}

	// Convert layout
	if phase43Schema.Layout.Type != "" {
		schema.Layout = phase43Schema.Layout.Type

		// Layout-specific configuration would need to be handled differently
		// since schema.Layout is now a string, not a complex object
	}

	// Convert components
	for i, component := range phase43Schema.Components {
		instance, err := imp.convertPhase43Component(&component, i, options)
		if err != nil {
			return nil, fmt.Errorf("failed to convert component %s: %w", component.Name, err)
		}
		schema.Components = append(schema.Components, *instance)
	}

	// Generate template variables from component props
	variables := imp.extractVariablesFromSchema(schema)

	// Create template
	template := &Template{
		ID:             templateID,
		Name:           phase43Schema.Name,
		Description:    phase43Schema.Description,
		Category:       options.DefaultCategory,
		Tags:           []string{"imported", "phase43"},
		Author:         options.DefaultAuthor,
		Version:        phase43Schema.Version,
		CreatedAt:      phase43Schema.CreatedAt,
		UpdatedAt:      time.Now(),
		Schema:         schema,
		Variables:      variables,
		License:        "free",
		IsPublic:       true,
		IsFeatured:     false,
		Complexity:     "intermediate",
		Rating:         0.0,
		UsageCount:     0,
		TargetAudience: []string{"developer"},
		Features:       []string{"responsive", "accessible"},
	}

	// Add category-specific tags
	if options.DefaultCategory != "" {
		template.Tags = append(template.Tags, options.DefaultCategory)
	}

	return template, nil
}

func (imp *Importer) convertPhase43Component(component *Phase43Component, index int, options *ImportOptions) (*core.ComponentInstance, error) {
	// Map component type if mapping exists
	componentType := component.Type
	if mappedType, exists := options.ComponentMapping[component.Type]; exists {
		componentType = mappedType
	}

	// Generate component ID
	componentID := component.Name
	if componentID == "" {
		componentID = fmt.Sprintf("component-%d", index)
	}
	componentID = strings.ToLower(strings.ReplaceAll(componentID, " ", "-"))

	// Create component instance
	instance := &core.ComponentInstance{
		ID:       componentID,
		Type:     componentType,
		Props:    make(map[string]interface{}),
		Position: core.Position{X: 0, Y: float64(index * 100)}, // Simple vertical layout
		Size:     core.Size{Width: 400, Height: 100},
	}

	// Convert properties
	for key, value := range component.Props {
		instance.Props[key] = value
	}

	// Convert events if present
	if len(component.Events) > 0 {
		instance.Events = &core.EventHandlers{
			Handlers: []core.EventHandler{},
		}

		for event := range component.Events {
			instance.Events.Handlers = append(instance.Events.Handlers, core.EventHandler{
				ID:    componentID + "-" + event,
				Event: event,
			})
		}
	}

	return instance, nil
}

func (imp *Importer) parseTemplFile(content string) (*TemplFile, error) {
	templFile := &TemplFile{
		Templates:  make([]TemplTemplate, 0),
		Imports:    make([]string, 0),
		Components: make([]TemplComponent, 0),
		Metadata:   make(map[string]interface{}),
	}

	lines := strings.Split(content, "\n")

	// Extract package
	packageRegex := regexp.MustCompile(`^package\s+(\w+)`)
	for _, line := range lines {
		if match := packageRegex.FindStringSubmatch(line); match != nil {
			templFile.Package = match[1]
			break
		}
	}

	// Extract imports
	importRegex := regexp.MustCompile(`import\s+"([^"]+)"`)
	for _, line := range lines {
		if matches := importRegex.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				templFile.Imports = append(templFile.Imports, match[1])
			}
		}
	}

	// Extract templates
	templateRegex := regexp.MustCompile(`templ\s+(\w+)\s*\(([^)]*)\)\s*{`)
	for i, line := range lines {
		if match := templateRegex.FindStringSubmatch(line); match != nil {
			template := TemplTemplate{
				Name:       match[1],
				Parameters: imp.parseTemplParameters(match[2]),
				Body:       imp.extractTemplateBody(lines, i),
				Metadata:   make(map[string]interface{}),
			}
			templFile.Templates = append(templFile.Templates, template)
		}
	}

	return templFile, nil
}

func (imp *Importer) parseTemplParameters(paramStr string) []TemplParameter {
	params := make([]TemplParameter, 0)

	if paramStr == "" {
		return params
	}

	// Simple parameter parsing - split by comma and extract name/type
	paramParts := strings.Split(paramStr, ",")
	for _, part := range paramParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by space to get name and type
		parts := strings.Fields(part)
		if len(parts) >= 2 {
			params = append(params, TemplParameter{
				Name: parts[0],
				Type: parts[1],
			})
		}
	}

	return params
}

func (imp *Importer) extractTemplateBody(lines []string, startLine int) string {
	body := strings.Builder{}
	braceCount := 0
	inTemplate := false

	for i := startLine; i < len(lines); i++ {
		line := lines[i]

		if !inTemplate {
			if strings.Contains(line, "{") {
				inTemplate = true
				braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			}
			continue
		}

		braceCount += strings.Count(line, "{") - strings.Count(line, "}")
		body.WriteString(line)
		body.WriteString("\n")

		if braceCount <= 0 {
			break
		}
	}

	return body.String()
}

func (imp *Importer) convertTemplTemplate(templTemplate *TemplTemplate, templFile *TemplFile, options *ImportOptions) (*Template, error) {
	// Create template ID
	templateID := strings.ToLower(templTemplate.Name)
	if options.TagPrefix != "" {
		templateID = options.TagPrefix + "-" + templateID
	}

	// Parse template body to extract components
	components, err := imp.extractComponentsFromTemplBody(templTemplate.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract components: %w", err)
	}

	// Convert components to values (not pointers)
	componentValues := make([]core.ComponentInstance, len(components))
	for i, comp := range components {
		componentValues[i] = *comp
	}

	// Create schema
	schema := &core.CompositionSchema{
		ID:          templateID + "-schema",
		Name:        templTemplate.Name,
		Description: fmt.Sprintf("Imported from Templ template: %s", templTemplate.Name),
		Components:  componentValues,
		Layout:      "flex",
	}

	// Extract variables from parameters
	variables := make(map[string]Variable)
	for _, param := range templTemplate.Parameters {
		variables[param.Name] = Variable{
			Name:        param.Name,
			Type:        imp.mapTemplTypeToVariableType(param.Type),
			Default:     options.VariableDefaults[param.Name],
			Description: fmt.Sprintf("Template parameter: %s", param.Name),
			Required:    true,
		}
	}

	// Create template
	template := &Template{
		ID:             templateID,
		Name:           templTemplate.Name,
		Description:    fmt.Sprintf("Imported from Templ template: %s", templTemplate.Name),
		Category:       options.DefaultCategory,
		Tags:           []string{"imported", "templ", templFile.Package},
		Author:         options.DefaultAuthor,
		Version:        "1.0.0",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Schema:         schema,
		Variables:      variables,
		License:        "free",
		IsPublic:       true,
		IsFeatured:     false,
		Complexity:     "intermediate",
		Rating:         0.0,
		UsageCount:     0,
		TargetAudience: []string{"developer"},
		Features:       []string{"templ-based", "server-side"},
	}

	return template, nil
}

func (imp *Importer) extractComponentsFromTemplBody(body string) ([]*core.ComponentInstance, error) {
	components := make([]*core.ComponentInstance, 0)

	// Simple HTML tag extraction for demo
	// In a real implementation, this would be much more sophisticated
	htmlTagRegex := regexp.MustCompile(`<(\w+)([^>]*)>`)
	matches := htmlTagRegex.FindAllStringSubmatch(body, -1)

	for i, match := range matches {
		tagName := match[1]
		attributes := match[2]

		// Skip common HTML tags, focus on component-like elements
		if imp.isCommonHTMLTag(tagName) {
			continue
		}

		componentID := fmt.Sprintf("component-%d", i)
		componentType := fmt.Sprintf("atoms.%s", tagName)

		// Parse attributes to properties
		properties := imp.parseHTMLAttributes(attributes)

		component := &core.ComponentInstance{
			ID:       componentID,
			Type:     componentType,
			Props:    properties,
			Position: core.Position{X: 0, Y: float64(i * 50)},
			Size:     core.Size{Width: 300, Height: 50},
		}

		components = append(components, component)
	}

	return components, nil
}

func (imp *Importer) extractVariablesFromSchema(schema *core.CompositionSchema) map[string]Variable {
	variables := make(map[string]Variable)

	// Extract variables from component properties
	for _, component := range schema.Components {
		for key, value := range component.Props {
			if valueStr, ok := value.(string); ok {
				// Look for template variables like {{varname}}
				varRegex := regexp.MustCompile(`\{\{(\w+)\}\}`)
				matches := varRegex.FindAllStringSubmatch(valueStr, -1)

				for _, match := range matches {
					varName := match[1]
					if _, exists := variables[varName]; !exists {
						variables[varName] = Variable{
							Name:        varName,
							Type:        "string",
							Default:     "",
							Description: fmt.Sprintf("Variable extracted from %s.%s", component.ID, key),
							Required:    false,
						}
					}
				}
			}
		}
	}

	return variables
}

// Utility methods

func (imp *Importer) isCommonHTMLTag(tagName string) bool {
	commonTags := map[string]bool{
		"div": true, "span": true, "p": true, "a": true, "img": true,
		"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
		"ul": true, "ol": true, "li": true, "table": true, "tr": true, "td": true, "th": true,
		"form": true, "input": true, "button": true, "select": true, "option": true,
		"header": true, "footer": true, "nav": true, "main": true, "section": true, "article": true,
	}
	return commonTags[strings.ToLower(tagName)]
}

func (imp *Importer) parseHTMLAttributes(attrStr string) map[string]interface{} {
	attributes := make(map[string]interface{})

	// Simple attribute parsing
	attrRegex := regexp.MustCompile(`(\w+)="([^"]*)"`)
	matches := attrRegex.FindAllStringSubmatch(attrStr, -1)

	for _, match := range matches {
		key := match[1]
		value := match[2]
		attributes[key] = value
	}

	return attributes
}

func (imp *Importer) mapTemplTypeToVariableType(templType string) string {
	typeMapping := map[string]string{
		"string":  "string",
		"int":     "number",
		"float64": "number",
		"bool":    "boolean",
	}

	if mapped, exists := typeMapping[templType]; exists {
		return mapped
	}
	return "string"
}
