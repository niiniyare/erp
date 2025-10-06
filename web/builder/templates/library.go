package templates

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/niiniyare/erp/web/builder/core"
)

// TemplateLibrary manages template storage and retrieval
type TemplateLibrary struct {
	templates map[string]*Template
	mu        sync.RWMutex
	registry  *core.ComponentRegistry
}

// Template represents a reusable UI template
type Template struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Category    string                  `json:"category"`
	Tags        []string                `json:"tags"`
	Author      string                  `json:"author"`
	Version     string                  `json:"version"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
	Schema      *core.CompositionSchema `json:"schema"`
	Thumbnail   string                  `json:"thumbnail"`
	Screenshots []string                `json:"screenshots"`
	UsageCount  int                     `json:"usageCount"`
	Rating      float64                 `json:"rating"`
	Complexity  string                  `json:"complexity"` // "simple", "intermediate", "advanced"

	// Template metadata
	TargetAudience []string            `json:"targetAudience"` // "developer", "designer", "business-user"
	Industries     []string            `json:"industries"`     // "healthcare", "finance", "ecommerce", etc.
	Features       []string            `json:"features"`       // "responsive", "accessible", "interactive"
	Dependencies   []string            `json:"dependencies"`   // Required component types
	Variables      map[string]Variable `json:"variables"`      // Customizable variables

	// License and usage
	License    string `json:"license"` // "free", "premium", "enterprise"
	IsPublic   bool   `json:"isPublic"`
	IsFeatured bool   `json:"isFeatured"`
}

// Variable represents a customizable template variable
type Variable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "string", "number", "color", "boolean", "select"
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Options     []Option    `json:"options,omitempty"` // For select type
	Required    bool        `json:"required"`
	Validation  string      `json:"validation,omitempty"` // Regex pattern
}

// Option represents a select option for variables
type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// TemplateCategory represents a template category
type TemplateCategory struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Order       int      `json:"order"`
	ParentID    string   `json:"parentId,omitempty"`
	Children    []string `json:"children,omitempty"`
}

// SearchFilter represents search and filter criteria
type SearchFilter struct {
	Query      string   `json:"query"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	Industries []string `json:"industries"`
	Complexity []string `json:"complexity"`
	License    []string `json:"license"`
	Author     string   `json:"author"`
	MinRating  float64  `json:"minRating"`
	Featured   *bool    `json:"featured"`
	SortBy     string   `json:"sortBy"`    // "name", "created", "updated", "rating", "usage"
	SortOrder  string   `json:"sortOrder"` // "asc", "desc"
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

// SearchResult represents search results
type SearchResult struct {
	Templates []*Template `json:"templates"`
	Total     int         `json:"total"`
	Page      int         `json:"page"`
	PerPage   int         `json:"perPage"`
	HasMore   bool        `json:"hasMore"`
}

// NewTemplateLibrary creates a new template library
func NewTemplateLibrary(registry *core.ComponentRegistry) *TemplateLibrary {
	library := &TemplateLibrary{
		templates: make(map[string]*Template),
		registry:  registry,
	}

	// Initialize with built-in templates
	library.initializeBuiltInTemplates()

	return library
}

// AddTemplate adds a template to the library
func (tl *TemplateLibrary) AddTemplate(template *Template) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	// Validate template schema
	if err := tl.validateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Set timestamps
	now := time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	template.UpdatedAt = now

	tl.templates[template.ID] = template
	return nil
}

// GetTemplate retrieves a template by ID
func (tl *TemplateLibrary) GetTemplate(id string) (*Template, error) {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	template, exists := tl.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	return template, nil
}

// SearchTemplates searches templates based on filter criteria
func (tl *TemplateLibrary) SearchTemplates(filter *SearchFilter) (*SearchResult, error) {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	var matchedTemplates []*Template

	// Filter templates
	for _, template := range tl.templates {
		if tl.matchesFilter(template, filter) {
			matchedTemplates = append(matchedTemplates, template)
		}
	}

	// Sort results
	tl.sortTemplates(matchedTemplates, filter.SortBy, filter.SortOrder)

	// Paginate
	total := len(matchedTemplates)
	start := filter.Offset
	end := start + filter.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	pagedTemplates := matchedTemplates[start:end]

	return &SearchResult{
		Templates: pagedTemplates,
		Total:     total,
		Page:      (filter.Offset / filter.Limit) + 1,
		PerPage:   filter.Limit,
		HasMore:   end < total,
	}, nil
}

// GetCategories returns all template categories
func (tl *TemplateLibrary) GetCategories() []*TemplateCategory {
	return []*TemplateCategory{
		{ID: "dashboard", Name: "Dashboard", Description: "Complete dashboard layouts", Icon: "📊", Order: 1},
		{ID: "forms", Name: "Forms", Description: "Form layouts and components", Icon: "📝", Order: 2},
		{ID: "tables", Name: "Tables", Description: "Data table layouts", Icon: "📋", Order: 3},
		{ID: "navigation", Name: "Navigation", Description: "Navigation patterns", Icon: "🧭", Order: 4},
		{ID: "auth", Name: "Authentication", Description: "Login and signup pages", Icon: "🔐", Order: 5},
		{ID: "ecommerce", Name: "E-commerce", Description: "Shopping and product pages", Icon: "🛒", Order: 6},
		{ID: "admin", Name: "Admin", Description: "Admin panel layouts", Icon: "⚙️", Order: 7},
		{ID: "landing", Name: "Landing Pages", Description: "Marketing and landing pages", Icon: "🎯", Order: 8},
		{ID: "crm", Name: "CRM", Description: "Customer relationship management", Icon: "👥", Order: 9},
		{ID: "finance", Name: "Finance", Description: "Financial and accounting layouts", Icon: "💰", Order: 10},
		{ID: "healthcare", Name: "Healthcare", Description: "Medical and healthcare interfaces", Icon: "🏥", Order: 11},
		{ID: "education", Name: "Education", Description: "Learning management systems", Icon: "🎓", Order: 12},
	}
}

// GetFeaturedTemplates returns featured templates
func (tl *TemplateLibrary) GetFeaturedTemplates() []*Template {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	var featured []*Template
	for _, template := range tl.templates {
		if template.IsFeatured {
			featured = append(featured, template)
		}
	}

	// Sort by rating and usage
	sort.Slice(featured, func(i, j int) bool {
		if featured[i].Rating != featured[j].Rating {
			return featured[i].Rating > featured[j].Rating
		}
		return featured[i].UsageCount > featured[j].UsageCount
	})

	return featured
}

// GetPopularTemplates returns popular templates based on usage
func (tl *TemplateLibrary) GetPopularTemplates(limit int) []*Template {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	var templates []*Template
	for _, template := range tl.templates {
		templates = append(templates, template)
	}

	// Sort by usage count
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].UsageCount > templates[j].UsageCount
	})

	if limit > 0 && len(templates) > limit {
		templates = templates[:limit]
	}

	return templates
}

// GetRecentTemplates returns recently added templates
func (tl *TemplateLibrary) GetRecentTemplates(limit int) []*Template {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	var templates []*Template
	for _, template := range tl.templates {
		templates = append(templates, template)
	}

	// Sort by creation date
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].CreatedAt.After(templates[j].CreatedAt)
	})

	if limit > 0 && len(templates) > limit {
		templates = templates[:limit]
	}

	return templates
}

// CloneTemplate creates a copy of a template with customizations
func (tl *TemplateLibrary) CloneTemplate(templateID string, variables map[string]interface{}) (*core.CompositionSchema, error) {
	template, err := tl.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Clone the schema
	schemaData, err := json.Marshal(template.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize template schema: %w", err)
	}

	var clonedSchema core.CompositionSchema
	if err := json.Unmarshal(schemaData, &clonedSchema); err != nil {
		return nil, fmt.Errorf("failed to deserialize template schema: %w", err)
	}

	// Apply variable substitutions
	if err := tl.applyVariables(&clonedSchema, template.Variables, variables); err != nil {
		return nil, fmt.Errorf("failed to apply variables: %w", err)
	}

	// Update usage count
	tl.incrementUsage(templateID)

	return &clonedSchema, nil
}

// Helper methods

func (tl *TemplateLibrary) validateTemplate(template *Template) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Schema == nil {
		return fmt.Errorf("template schema is required")
	}

	// Validate that all required components exist in registry
	for _, component := range template.Schema.Components {
		if !tl.registry.HasComponent(component.Type) {
			return fmt.Errorf("unknown component type: %s", component.Type)
		}
	}

	return nil
}

func (tl *TemplateLibrary) matchesFilter(template *Template, filter *SearchFilter) bool {
	// Query search
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)
		if !strings.Contains(strings.ToLower(template.Name), query) &&
			!strings.Contains(strings.ToLower(template.Description), query) &&
			!tl.containsAnyString(template.Tags, query) {
			return false
		}
	}

	// Category filter
	if len(filter.Categories) > 0 && !tl.containsString(filter.Categories, template.Category) {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 && !tl.hasAnyTag(template.Tags, filter.Tags) {
		return false
	}

	// Industry filter
	if len(filter.Industries) > 0 && !tl.hasAnyTag(template.Industries, filter.Industries) {
		return false
	}

	// Complexity filter
	if len(filter.Complexity) > 0 && !tl.containsString(filter.Complexity, template.Complexity) {
		return false
	}

	// License filter
	if len(filter.License) > 0 && !tl.containsString(filter.License, template.License) {
		return false
	}

	// Author filter
	if filter.Author != "" && !strings.Contains(strings.ToLower(template.Author), strings.ToLower(filter.Author)) {
		return false
	}

	// Rating filter
	if filter.MinRating > 0 && template.Rating < filter.MinRating {
		return false
	}

	// Featured filter
	if filter.Featured != nil && template.IsFeatured != *filter.Featured {
		return false
	}

	return true
}

func (tl *TemplateLibrary) sortTemplates(templates []*Template, sortBy, sortOrder string) {
	ascending := sortOrder != "desc"

	sort.Slice(templates, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "name":
			less = templates[i].Name < templates[j].Name
		case "created":
			less = templates[i].CreatedAt.Before(templates[j].CreatedAt)
		case "updated":
			less = templates[i].UpdatedAt.Before(templates[j].UpdatedAt)
		case "rating":
			less = templates[i].Rating < templates[j].Rating
		case "usage":
			less = templates[i].UsageCount < templates[j].UsageCount
		default:
			less = templates[i].Name < templates[j].Name
		}

		if ascending {
			return less
		}
		return !less
	})
}

func (tl *TemplateLibrary) applyVariables(schema *core.CompositionSchema, templateVars map[string]Variable, userVars map[string]interface{}) error {
	// Serialize schema to JSON for string replacement
	schemaData, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	schemaString := string(schemaData)

	// Apply each variable
	for varName, varDef := range templateVars {
		placeholder := fmt.Sprintf("{{%s}}", varName)

		var value interface{}
		if userValue, exists := userVars[varName]; exists {
			value = userValue
		} else {
			value = varDef.Default
		}

		// Convert value to string
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case bool:
			valueStr = fmt.Sprintf("%t", v)
		case float64:
			valueStr = fmt.Sprintf("%.2f", v)
		case int:
			valueStr = fmt.Sprintf("%d", v)
		default:
			valueBytes, _ := json.Marshal(v)
			valueStr = string(valueBytes)
		}

		schemaString = strings.ReplaceAll(schemaString, placeholder, valueStr)
	}

	// Parse back to schema
	return json.Unmarshal([]byte(schemaString), schema)
}

func (tl *TemplateLibrary) incrementUsage(templateID string) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if template, exists := tl.templates[templateID]; exists {
		template.UsageCount++
		template.UpdatedAt = time.Now()
	}
}

// Utility functions

func (tl *TemplateLibrary) containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (tl *TemplateLibrary) containsAnyString(slice []string, query string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), query) {
			return true
		}
	}
	return false
}

func (tl *TemplateLibrary) hasAnyTag(templateTags, filterTags []string) bool {
	for _, filterTag := range filterTags {
		for _, templateTag := range templateTags {
			if templateTag == filterTag {
				return true
			}
		}
	}
	return false
}

// Initialize built-in templates
func (tl *TemplateLibrary) initializeBuiltInTemplates() {
	// Add sample templates
	tl.addBuiltInTemplate("dashboard-basic", "Basic Dashboard", "dashboard", []string{"dashboard", "admin", "overview"})
	tl.addBuiltInTemplate("form-contact", "Contact Form", "forms", []string{"form", "contact", "lead"})
	tl.addBuiltInTemplate("table-users", "User Management Table", "tables", []string{"table", "users", "admin"})
	tl.addBuiltInTemplate("nav-sidebar", "Sidebar Navigation", "navigation", []string{"navigation", "sidebar", "menu"})
	tl.addBuiltInTemplate("auth-login", "Login Page", "auth", []string{"login", "authentication", "security"})
	tl.addBuiltInTemplate("ecommerce-product", "Product Catalog", "ecommerce", []string{"product", "catalog", "shopping"})
}

func (tl *TemplateLibrary) addBuiltInTemplate(id, name, category string, tags []string) {
	template := &Template{
		ID:          id,
		Name:        name,
		Description: fmt.Sprintf("Built-in %s template", name),
		Category:    category,
		Tags:        tags,
		Author:      "System",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		License:     "free",
		IsPublic:    true,
		IsFeatured:  true,
		Complexity:  "simple",
		Rating:      4.5,
		UsageCount:  0,
		Schema:      tl.createBuiltInSchema(id),
		Variables:   tl.createBuiltInVariables(id),
	}

	tl.templates[id] = template
}

func (tl *TemplateLibrary) createBuiltInSchema(templateID string) *core.CompositionSchema {
	// Create basic schema based on template ID
	schema := &core.CompositionSchema{
		ID:          templateID + "-schema",
		Name:        "Generated Schema",
		Description: "Auto-generated schema for " + templateID,
		Components:  []core.ComponentInstance{},
		Layout:      "flex",
	}

	// Add sample components based on template type
	switch {
	case strings.Contains(templateID, "dashboard"):
		schema.Components = append(schema.Components,
			core.ComponentInstance{
				ID:       "header",
				Type:     "molecules.header",
				Position: core.Position{X: 0, Y: 0},
				Size:     core.Size{Width: 1200, Height: 80},
				Props: map[string]interface{}{
					"title":    "{{title}}",
					"subtitle": "{{subtitle}}",
				},
			},
			core.ComponentInstance{
				ID:       "metrics",
				Type:     "organisms.metrics-grid",
				Position: core.Position{X: 0, Y: 100},
				Size:     core.Size{Width: 1200, Height: 200},
				Props: map[string]interface{}{
					"columns": 4,
					"gap":     16,
				},
			},
		)
	case strings.Contains(templateID, "form"):
		schema.Components = append(schema.Components,
			core.ComponentInstance{
				ID:       "form",
				Type:     "molecules.form",
				Position: core.Position{X: 0, Y: 0},
				Size:     core.Size{Width: 600, Height: 400},
				Props: map[string]interface{}{
					"title":  "{{formTitle}}",
					"method": "POST",
					"action": "/contact",
				},
			},
		)
	case strings.Contains(templateID, "table"):
		schema.Components = append(schema.Components,
			core.ComponentInstance{
				ID:       "table",
				Type:     "organisms.table",
				Position: core.Position{X: 0, Y: 0},
				Size:     core.Size{Width: 1200, Height: 600},
				Props: map[string]interface{}{
					"title":      "{{tableTitle}}",
					"pagination": true,
					"search":     true,
				},
			},
		)
	}

	return schema
}

func (tl *TemplateLibrary) createBuiltInVariables(templateID string) map[string]Variable {
	variables := make(map[string]Variable)

	switch {
	case strings.Contains(templateID, "dashboard"):
		variables["title"] = Variable{
			Name:        "title",
			Type:        "string",
			Default:     "Dashboard",
			Description: "Main dashboard title",
			Required:    true,
		}
		variables["subtitle"] = Variable{
			Name:        "subtitle",
			Type:        "string",
			Default:     "Welcome to your dashboard",
			Description: "Dashboard subtitle",
			Required:    false,
		}
	case strings.Contains(templateID, "form"):
		variables["formTitle"] = Variable{
			Name:        "formTitle",
			Type:        "string",
			Default:     "Contact Us",
			Description: "Form title",
			Required:    true,
		}
	case strings.Contains(templateID, "table"):
		variables["tableTitle"] = Variable{
			Name:        "tableTitle",
			Type:        "string",
			Default:     "Data Table",
			Description: "Table title",
			Required:    true,
		}
	}

	return variables
}
