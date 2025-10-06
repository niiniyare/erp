package generator

import (
	"fmt"
	"strings"
	"time"
)

// UISchema represents the generated UI schema for a struct
type UISchema struct {
	Entity      string `json:"entity"`
	Package     string `json:"package"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`

	// Schema components
	ListSchema   *PageSchema   `json:"listSchema"`
	CreateSchema *PageSchema   `json:"createSchema"`
	EditSchema   *PageSchema   `json:"editSchema"`
	DetailSchema *PageSchema   `json:"detailSchema"`
	FilterSchema *FilterSchema `json:"filterSchema"`

	// Entity metadata
	Metadata EntityMetadata `json:"metadata"`

	// Generated at timestamp
	GeneratedAt time.Time `json:"generatedAt"`
}

// FilterSchema defines filtering capabilities
type FilterSchema struct {
	ID       string                  `json:"id"`
	Title    string                  `json:"title"`
	Fields   []FilterFieldDefinition `json:"fields"`
	Presets  []FilterPreset          `json:"presets"`
	Advanced bool                    `json:"advanced"`
}

// FilterFieldDefinition defines a filterable field
type FilterFieldDefinition struct {
	Field       string         `json:"field"`
	Type        string         `json:"type"` // "text", "select", "date-range", "number-range"
	Label       string         `json:"label"`
	Options     []FilterOption `json:"options,omitempty"`
	Multiple    bool           `json:"multiple,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
}

// FilterOption represents a filter option
type FilterOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// FilterPreset defines a predefined filter combination
type FilterPreset struct {
	ID      string            `json:"id"`
	Label   string            `json:"label"`
	Icon    string            `json:"icon,omitempty"`
	Filters map[string]string `json:"filters"`
}

// EntityMetadata contains metadata about the entity
type EntityMetadata struct {
	IsCRUD         bool     `json:"isCrud"`
	IsReadOnly     bool     `json:"isReadOnly"`
	HasWorkflow    bool     `json:"hasWorkflow"`
	HasValidation  bool     `json:"hasValidation"`
	HasAuditTrail  bool     `json:"hasAuditTrail"`
	HasSoftDelete  bool     `json:"hasSoftDelete"`
	HasTenantScope bool     `json:"hasTenantScope"`
	HasHierarchy   bool     `json:"hasHierarchy"`
	PrimaryKey     string   `json:"primaryKey"`
	Relationships  []string `json:"relationships"`
}

// PageSchema represents a complete page schema (imported from existing types)
type PageSchema struct {
	ID          string                `json:"id"`
	Version     string                `json:"version,omitempty"`
	Layout      string                `json:"layout"` // "app", "auth", "minimal", "base"
	Title       string                `json:"title"`
	Description string                `json:"description,omitempty"`
	Components  []ComponentDefinition `json:"components"`
	DataSources []DataSource          `json:"dataSources,omitempty"`
	Actions     []ActionDefinition    `json:"actions,omitempty"`
	Permissions *PermissionRules      `json:"permissions,omitempty"`
	Meta        map[string]any        `json:"meta,omitempty"`
	CreatedAt   time.Time             `json:"createdAt,omitempty"`
	UpdatedAt   time.Time             `json:"updatedAt,omitempty"`
}

// ComponentDefinition and related types (simplified for generator)
type ComponentDefinition struct {
	ID          string                `json:"id"`
	Type        string                `json:"type"`
	Props       map[string]any        `json:"props"`
	Layout      *LayoutRules          `json:"layout,omitempty"`
	Children    []ComponentDefinition `json:"children,omitempty"`
	DataSource  string                `json:"dataSource,omitempty"`
	Permissions *PermissionRules      `json:"permissions,omitempty"`
	Conditions  []RenderCondition     `json:"conditions,omitempty"`
}

type LayoutRules struct {
	Container  string            `json:"container,omitempty"`
	Grid       *GridLayout       `json:"grid,omitempty"`
	Classes    []string          `json:"classes,omitempty"`
	Responsive map[string]string `json:"responsive,omitempty"`
}

type GridLayout struct {
	Columns    string `json:"columns"`
	Gap        string `json:"gap,omitempty"`
	ColumnSpan int    `json:"columnSpan,omitempty"`
}

type DataSource struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Endpoint string            `json:"endpoint,omitempty"`
	Method   string            `json:"method,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

type ActionDefinition struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Label       string            `json:"label,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	Target      string            `json:"target,omitempty"`
	Method      string            `json:"method,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
}

type PermissionRules struct {
	RequiredRoles       []string          `json:"requiredRoles,omitempty"`
	RequiredPermissions []string          `json:"requiredPermissions,omitempty"`
	Conditions          []PermissionCheck `json:"conditions,omitempty"`
}

type PermissionCheck struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Source   string `json:"source,omitempty"`
}

type RenderCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
	Source   string `json:"source,omitempty"`
}

// SchemaGenerator generates UI schemas from struct information
type SchemaGenerator struct {
	patternMatcher *PatternMatcher
	config         GeneratorConfig
}

// GeneratorConfig contains configuration for schema generation
type GeneratorConfig struct {
	DefaultLayout     string            `json:"defaultLayout"`
	EnablePermissions bool              `json:"enablePermissions"`
	EnableAuditTrail  bool              `json:"enableAuditTrail"`
	EnableSoftDelete  bool              `json:"enableSoftDelete"`
	BaseURL           string            `json:"baseUrl"`
	APIPrefix         string            `json:"apiPrefix"`
	CustomComponents  map[string]string `json:"customComponents"`
	TablePageSize     int               `json:"tablePageSize"`
	EnableFiltering   bool              `json:"enableFiltering"`
	EnableSorting     bool              `json:"enableSorting"`
	EnableBulkActions bool              `json:"enableBulkActions"`
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(config GeneratorConfig) *SchemaGenerator {
	// Set defaults
	if config.DefaultLayout == "" {
		config.DefaultLayout = "app"
	}
	if config.APIPrefix == "" {
		config.APIPrefix = "/api"
	}
	if config.TablePageSize == 0 {
		config.TablePageSize = 20
	}

	return &SchemaGenerator{
		patternMatcher: NewPatternMatcher(),
		config:         config,
	}
}

// GenerateUISchema creates a complete UI schema from struct information
func (sg *SchemaGenerator) GenerateUISchema(structInfo StructInfo) (*UISchema, error) {
	if structInfo.Name == "" {
		return nil, fmt.Errorf("struct name cannot be empty")
	}

	// Generate entity metadata
	metadata := sg.generateEntityMetadata(structInfo)

	// Create base schema
	schema := &UISchema{
		Entity:      structInfo.Name,
		Package:     structInfo.PackageName,
		Title:       sg.generateEntityTitle(structInfo.Name),
		Description: sg.generateEntityDescription(structInfo),
		Version:     "1.0.0",
		Metadata:    metadata,
		GeneratedAt: time.Now(),
	}

	// Generate different view schemas
	if metadata.IsCRUD {
		var err error

		// List view schema
		schema.ListSchema, err = sg.generateListSchema(structInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate list schema: %w", err)
		}

		// Create form schema
		if !metadata.IsReadOnly {
			schema.CreateSchema, err = sg.generateCreateSchema(structInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to generate create schema: %w", err)
			}

			// Edit form schema
			schema.EditSchema, err = sg.generateEditSchema(structInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to generate edit schema: %w", err)
			}
		}

		// Detail view schema
		schema.DetailSchema, err = sg.generateDetailSchema(structInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to generate detail schema: %w", err)
		}

		// Filter schema
		if sg.config.EnableFiltering {
			schema.FilterSchema = sg.generateFilterSchema(structInfo)
		}
	}

	return schema, nil
}

// generateEntityMetadata creates metadata from struct information
func (sg *SchemaGenerator) generateEntityMetadata(structInfo StructInfo) EntityMetadata {
	var relationships []string
	for _, field := range structInfo.Fields {
		if field.IsRelationship && field.TargetEntity != "" {
			relationships = append(relationships, field.TargetEntity)
		}
	}

	return EntityMetadata{
		IsCRUD:         structInfo.IsCRUDEntity,
		IsReadOnly:     structInfo.IsReadOnly,
		HasWorkflow:    structInfo.HasWorkflow,
		HasValidation:  structInfo.HasValidation,
		HasAuditTrail:  structInfo.HasAuditTrail,
		HasSoftDelete:  structInfo.HasSoftDelete,
		HasTenantScope: structInfo.HasTenantScope,
		HasHierarchy:   structInfo.HasHierarchy,
		PrimaryKey:     structInfo.PrimaryKeyField,
		Relationships:  relationships,
	}
}

// generateListSchema creates a list/table view schema
func (sg *SchemaGenerator) generateListSchema(structInfo StructInfo) (*PageSchema, error) {
	entityName := strings.ToLower(structInfo.Name)
	entityPlural := sg.pluralize(entityName)

	schema := &PageSchema{
		ID:          fmt.Sprintf("%s-list", entityName),
		Layout:      sg.config.DefaultLayout,
		Title:       sg.generateEntityTitle(structInfo.Name) + " List",
		Description: fmt.Sprintf("List and manage %s", entityPlural),
		Components:  []ComponentDefinition{},
		DataSources: []DataSource{},
		Actions:     []ActionDefinition{},
	}

	// Add data source for list data
	schema.DataSources = append(schema.DataSources, DataSource{
		ID:       fmt.Sprintf("%s-list-data", entityName),
		Type:     "api",
		Endpoint: fmt.Sprintf("%s/%s", sg.config.APIPrefix, entityPlural),
		Method:   "GET",
	})

	// Generate table component
	tableComponent := sg.generateTableComponent(structInfo)
	schema.Components = append(schema.Components, tableComponent)

	// Add create action if not read-only
	if !structInfo.IsReadOnly {
		createAction := ActionDefinition{
			ID:     fmt.Sprintf("create-%s", entityName),
			Type:   "navigate",
			Label:  fmt.Sprintf("Create %s", sg.generateEntityTitle(structInfo.Name)),
			Icon:   "plus",
			Target: fmt.Sprintf("/%s/create", entityPlural),
		}

		if sg.config.EnablePermissions {
			createAction.Permissions = []string{fmt.Sprintf("%s:create", entityName)}
		}

		schema.Actions = append(schema.Actions, createAction)
	}

	return schema, nil
}

// generateCreateSchema creates a form schema for creating new entities
func (sg *SchemaGenerator) generateCreateSchema(structInfo StructInfo) (*PageSchema, error) {
	entityName := strings.ToLower(structInfo.Name)

	schema := &PageSchema{
		ID:          fmt.Sprintf("%s-create", entityName),
		Layout:      sg.config.DefaultLayout,
		Title:       fmt.Sprintf("Create %s", sg.generateEntityTitle(structInfo.Name)),
		Description: fmt.Sprintf("Create a new %s", entityName),
		Components:  []ComponentDefinition{},
		Actions:     []ActionDefinition{},
	}

	// Generate form component
	formComponent := sg.generateFormComponent(structInfo, "create")
	schema.Components = append(schema.Components, formComponent)

	// Add submit action
	submitAction := ActionDefinition{
		ID:     fmt.Sprintf("submit-create-%s", entityName),
		Type:   "submit",
		Label:  "Create",
		Icon:   "save",
		Target: fmt.Sprintf("%s/%s", sg.config.APIPrefix, sg.pluralize(entityName)),
		Method: "POST",
	}

	if sg.config.EnablePermissions {
		submitAction.Permissions = []string{fmt.Sprintf("%s:create", entityName)}
	}

	schema.Actions = append(schema.Actions, submitAction)

	return schema, nil
}

// generateEditSchema creates a form schema for editing entities
func (sg *SchemaGenerator) generateEditSchema(structInfo StructInfo) (*PageSchema, error) {
	entityName := strings.ToLower(structInfo.Name)

	schema := &PageSchema{
		ID:          fmt.Sprintf("%s-edit", entityName),
		Layout:      sg.config.DefaultLayout,
		Title:       fmt.Sprintf("Edit %s", sg.generateEntityTitle(structInfo.Name)),
		Description: fmt.Sprintf("Edit %s details", entityName),
		Components:  []ComponentDefinition{},
		DataSources: []DataSource{},
		Actions:     []ActionDefinition{},
	}

	// Add data source for existing entity data
	schema.DataSources = append(schema.DataSources, DataSource{
		ID:       fmt.Sprintf("%s-edit-data", entityName),
		Type:     "api",
		Endpoint: fmt.Sprintf("%s/%s/{id}", sg.config.APIPrefix, sg.pluralize(entityName)),
		Method:   "GET",
	})

	// Generate form component
	formComponent := sg.generateFormComponent(structInfo, "edit")
	formComponent.DataSource = fmt.Sprintf("%s-edit-data", entityName)
	schema.Components = append(schema.Components, formComponent)

	// Add update action
	updateAction := ActionDefinition{
		ID:     fmt.Sprintf("submit-edit-%s", entityName),
		Type:   "submit",
		Label:  "Update",
		Icon:   "save",
		Target: fmt.Sprintf("%s/%s/{id}", sg.config.APIPrefix, sg.pluralize(entityName)),
		Method: "PUT",
	}

	if sg.config.EnablePermissions {
		updateAction.Permissions = []string{fmt.Sprintf("%s:update", entityName)}
	}

	schema.Actions = append(schema.Actions, updateAction)

	return schema, nil
}

// generateDetailSchema creates a detail view schema
func (sg *SchemaGenerator) generateDetailSchema(structInfo StructInfo) (*PageSchema, error) {
	entityName := strings.ToLower(structInfo.Name)

	schema := &PageSchema{
		ID:          fmt.Sprintf("%s-detail", entityName),
		Layout:      sg.config.DefaultLayout,
		Title:       fmt.Sprintf("%s Details", sg.generateEntityTitle(structInfo.Name)),
		Description: fmt.Sprintf("View %s details", entityName),
		Components:  []ComponentDefinition{},
		DataSources: []DataSource{},
		Actions:     []ActionDefinition{},
	}

	// Add data source
	schema.DataSources = append(schema.DataSources, DataSource{
		ID:       fmt.Sprintf("%s-detail-data", entityName),
		Type:     "api",
		Endpoint: fmt.Sprintf("%s/%s/{id}", sg.config.APIPrefix, sg.pluralize(entityName)),
		Method:   "GET",
	})

	// Generate detail view component
	detailComponent := sg.generateDetailComponent(structInfo)
	detailComponent.DataSource = fmt.Sprintf("%s-detail-data", entityName)
	schema.Components = append(schema.Components, detailComponent)

	// Add edit action if not read-only
	if !structInfo.IsReadOnly {
		editAction := ActionDefinition{
			ID:     fmt.Sprintf("edit-%s", entityName),
			Type:   "navigate",
			Label:  "Edit",
			Icon:   "edit",
			Target: fmt.Sprintf("/%s/{id}/edit", sg.pluralize(entityName)),
		}

		if sg.config.EnablePermissions {
			editAction.Permissions = []string{fmt.Sprintf("%s:update", entityName)}
		}

		schema.Actions = append(schema.Actions, editAction)
	}

	return schema, nil
}

// generateFilterSchema creates a filter schema for the entity
func (sg *SchemaGenerator) generateFilterSchema(structInfo StructInfo) *FilterSchema {
	entityName := strings.ToLower(structInfo.Name)

	filterSchema := &FilterSchema{
		ID:       fmt.Sprintf("%s-filters", entityName),
		Title:    "Filter " + sg.generateEntityTitle(structInfo.Name),
		Fields:   []FilterFieldDefinition{},
		Presets:  []FilterPreset{},
		Advanced: false,
	}

	// Generate filter fields from struct fields
	for _, field := range structInfo.Fields {
		if sg.shouldIncludeInFilter(field) {
			filterField := sg.generateFilterField(field)
			filterSchema.Fields = append(filterSchema.Fields, filterField)
		}
	}

	// Add common presets
	if structInfo.HasStatus {
		filterSchema.Presets = append(filterSchema.Presets, FilterPreset{
			ID:    "active",
			Label: "Active",
			Icon:  "check-circle",
			Filters: map[string]string{
				"status": "active",
			},
		})
	}

	if structInfo.HasCreatedAt {
		filterSchema.Presets = append(filterSchema.Presets, FilterPreset{
			ID:    "recent",
			Label: "Recent",
			Icon:  "clock",
			Filters: map[string]string{
				"created_at": "last_7_days",
			},
		})
	}

	return filterSchema
}

// Helper methods

// generateTableComponent creates a table component definition
func (sg *SchemaGenerator) generateTableComponent(structInfo StructInfo) ComponentDefinition {
	entityName := strings.ToLower(structInfo.Name)

	// Generate columns from struct fields
	columns := []map[string]any{}
	for _, field := range structInfo.Fields {
		if sg.shouldIncludeInTable(field) {
			column := map[string]any{
				"field":    strings.ToLower(field.Name),
				"label":    sg.generateFieldLabel(field),
				"type":     sg.mapFieldTypeToColumnType(field),
				"sortable": sg.isFieldSortable(field),
			}

			if field.IsRelationship {
				column["component"] = "organisms.relationship-cell"
				column["props"] = map[string]any{
					"entity": field.TargetEntity,
				}
			}

			columns = append(columns, column)
		}
	}

	props := map[string]any{
		"title":      sg.generateEntityTitle(structInfo.Name) + " List",
		"columns":    columns,
		"dataSource": fmt.Sprintf("%s-list-data", entityName),
		"pageSize":   sg.config.TablePageSize,
		"sortable":   sg.config.EnableSorting,
		"filterable": sg.config.EnableFiltering,
		"searchable": true,
	}

	if sg.config.EnableBulkActions {
		props["bulkActions"] = []map[string]any{
			{
				"id":    "delete",
				"label": "Delete Selected",
				"icon":  "trash",
				"type":  "danger",
			},
		}
	}

	return ComponentDefinition{
		ID:    fmt.Sprintf("%s-table", entityName),
		Type:  "organisms.table",
		Props: props,
		Layout: &LayoutRules{
			Container: "grid",
			Grid: &GridLayout{
				Columns: "1fr",
				Gap:     "1.5rem",
			},
		},
	}
}

// generateFormComponent creates a form component definition
func (sg *SchemaGenerator) generateFormComponent(structInfo StructInfo, mode string) ComponentDefinition {
	entityName := strings.ToLower(structInfo.Name)

	// Generate form fields
	fields := []map[string]any{}
	for _, field := range structInfo.Fields {
		if sg.shouldIncludeInForm(field, mode) {
			formField := sg.generateFormField(field)
			fields = append(fields, formField)
		}
	}

	props := map[string]any{
		"title":  fmt.Sprintf("%s %s", strings.Title(mode), sg.generateEntityTitle(structInfo.Name)),
		"fields": fields,
		"layout": "vertical",
	}

	if structInfo.HasValidation {
		props["validation"] = true
		props["validateOnChange"] = true
	}

	return ComponentDefinition{
		ID:    fmt.Sprintf("%s-%s-form", entityName, mode),
		Type:  "organisms.form",
		Props: props,
		Layout: &LayoutRules{
			Container: "grid",
			Grid: &GridLayout{
				Columns: "1fr",
				Gap:     "1.5rem",
			},
		},
	}
}

// generateDetailComponent creates a detail view component
func (sg *SchemaGenerator) generateDetailComponent(structInfo StructInfo) ComponentDefinition {
	entityName := strings.ToLower(structInfo.Name)

	// Generate detail sections
	sections := []map[string]any{}

	// Basic information section
	basicFields := []map[string]any{}
	for _, field := range structInfo.Fields {
		if sg.shouldIncludeInDetail(field) && !field.IsRelationship {
			detailField := map[string]any{
				"field": strings.ToLower(field.Name),
				"label": sg.generateFieldLabel(field),
				"type":  sg.mapFieldTypeToDisplayType(field),
			}
			basicFields = append(basicFields, detailField)
		}
	}

	if len(basicFields) > 0 {
		sections = append(sections, map[string]any{
			"title":  "Basic Information",
			"fields": basicFields,
		})
	}

	// Relationships section
	relationshipFields := []map[string]any{}
	for _, field := range structInfo.Fields {
		if field.IsRelationship {
			relationField := map[string]any{
				"field":  strings.ToLower(field.Name),
				"label":  sg.generateFieldLabel(field),
				"type":   "relationship",
				"entity": field.TargetEntity,
			}
			relationshipFields = append(relationshipFields, relationField)
		}
	}

	if len(relationshipFields) > 0 {
		sections = append(sections, map[string]any{
			"title":  "Related Data",
			"fields": relationshipFields,
		})
	}

	props := map[string]any{
		"title":    sg.generateEntityTitle(structInfo.Name) + " Details",
		"sections": sections,
		"layout":   "sections",
	}

	if structInfo.HasAuditTrail {
		props["showAuditTrail"] = true
	}

	return ComponentDefinition{
		ID:    fmt.Sprintf("%s-detail", entityName),
		Type:  "organisms.detail-view",
		Props: props,
		Layout: &LayoutRules{
			Container: "grid",
			Grid: &GridLayout{
				Columns: "1fr",
				Gap:     "2rem",
			},
		},
	}
}

// Additional helper methods for field processing, validation, and type mapping
// These would continue with the same level of detail...

// TODO: Complete implementation of helper methods:
// - generateFormField(field FieldInfo) map[string]any
// - generateFilterField(field FieldInfo) FilterFieldDefinition
// - shouldIncludeInTable(field FieldInfo) bool
// - shouldIncludeInForm(field FieldInfo, mode string) bool
// - shouldIncludeInDetail(field FieldInfo) bool
// - shouldIncludeInFilter(field FieldInfo) bool
// - mapFieldTypeToColumnType(field FieldInfo) string
// - mapFieldTypeToDisplayType(field FieldInfo) string
// - generateFieldLabel(field FieldInfo) string
// - isFieldSortable(field FieldInfo) bool
// - generateEntityTitle(name string) string
// - generateEntityDescription(structInfo StructInfo) string
// - pluralize(singular string) string

// Placeholder implementations
func (sg *SchemaGenerator) generateFormField(field FieldInfo) map[string]any {
	// TODO: Complete implementation
	return map[string]any{
		"field":       strings.ToLower(field.Name),
		"label":       sg.generateFieldLabel(field),
		"type":        field.Component,
		"required":    field.Required,
		"placeholder": field.Placeholder,
	}
}

func (sg *SchemaGenerator) generateFilterField(field FieldInfo) FilterFieldDefinition {
	// TODO: Complete implementation
	return FilterFieldDefinition{
		Field:       strings.ToLower(field.Name),
		Type:        sg.mapFieldTypeToFilterType(field),
		Label:       sg.generateFieldLabel(field),
		Placeholder: fmt.Sprintf("Filter by %s", sg.generateFieldLabel(field)),
	}
}

func (sg *SchemaGenerator) shouldIncludeInTable(field FieldInfo) bool {
	return !field.Hidden && field.Name != "ID" && !strings.Contains(strings.ToLower(field.Name), "password")
}

func (sg *SchemaGenerator) shouldIncludeInForm(field FieldInfo, mode string) bool {
	if field.Hidden {
		return false
	}

	// Skip auto-generated fields in create mode
	if mode == "create" {
		skipFields := []string{"id", "createdat", "updatedat", "deletedat"}
		for _, skip := range skipFields {
			if strings.ToLower(field.Name) == skip {
				return false
			}
		}
	}

	return true
}

func (sg *SchemaGenerator) shouldIncludeInDetail(field FieldInfo) bool {
	return !field.Hidden && !strings.Contains(strings.ToLower(field.Name), "password")
}

func (sg *SchemaGenerator) shouldIncludeInFilter(field FieldInfo) bool {
	return sg.isFieldFilterable(field) && !field.Hidden
}

func (sg *SchemaGenerator) isFieldFilterable(field FieldInfo) bool {
	return field.IsStringType || field.IsBoolType || field.IsTimeType ||
		field.IsRelationship || len(field.Options) > 0
}

func (sg *SchemaGenerator) mapFieldTypeToColumnType(field FieldInfo) string {
	// TODO: Complete mapping
	if field.IsTimeType {
		return "datetime"
	}
	if field.IsBoolType {
		return "boolean"
	}
	if field.IsNumberType {
		return "number"
	}
	return "text"
}

func (sg *SchemaGenerator) mapFieldTypeToDisplayType(field FieldInfo) string {
	return sg.mapFieldTypeToColumnType(field)
}

func (sg *SchemaGenerator) mapFieldTypeToFilterType(field FieldInfo) string {
	if field.IsTimeType {
		return "date-range"
	}
	if field.IsBoolType {
		return "select"
	}
	if field.IsNumberType {
		return "number-range"
	}
	if len(field.Options) > 0 {
		return "select"
	}
	return "text"
}

func (sg *SchemaGenerator) generateFieldLabel(field FieldInfo) string {
	if field.Label != "" {
		return field.Label
	}
	return sg.toTitleCase(field.Name)
}

func (sg *SchemaGenerator) isFieldSortable(field FieldInfo) bool {
	return !field.IsSlice && (field.IsStringType || field.IsNumberType || field.IsTimeType)
}

func (sg *SchemaGenerator) generateEntityTitle(name string) string {
	return sg.toTitleCase(name)
}

func (sg *SchemaGenerator) generateEntityDescription(structInfo StructInfo) string {
	return fmt.Sprintf("Manage %s entities with full CRUD capabilities", strings.ToLower(structInfo.Name))
}

func (sg *SchemaGenerator) pluralize(singular string) string {
	// Simple pluralization - in production, use a proper library
	if strings.HasSuffix(singular, "y") {
		return strings.TrimSuffix(singular, "y") + "ies"
	}
	if strings.HasSuffix(singular, "s") {
		return singular + "es"
	}
	return singular + "s"
}

func (sg *SchemaGenerator) toTitleCase(s string) string {
	// Convert camelCase/PascalCase to Title Case
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		if i == 0 {
			result.WriteRune(r)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// NOTE: This implementation provides a solid foundation for schema generation
// TODO: Enhanced field type mapping, validation rule generation, and relationship handling
// FIXME: Pluralization needs proper implementation using inflection library
