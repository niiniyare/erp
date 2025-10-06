package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
	"github.com/niiniyare/erp/web/schemas"
)

// PatternRenderer handles rendering of unified UI patterns
type PatternRenderer struct {
	registry       *ComponentRegistry
	schemaRenderer *SchemaRenderer
	patterns       map[string]*PatternDefinition
}

// PatternDefinition represents a complete UI pattern
type PatternDefinition struct {
	ID          string                                  `json:"id"`
	Title       string                                  `json:"title"`
	Description string                                  `json:"description"`
	Schema      *schemas.ComponentDefinition            `json:"schema"`
	Variants    map[string]*schemas.ComponentDefinition `json:"variants,omitempty"`
	Examples    []PatternExample                        `json:"examples,omitempty"`
}

// PatternExample shows how to use a pattern
type PatternExample struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Data        map[string]any `json:"data"`
	Props       map[string]any `json:"props"`
}

// NewPatternRenderer creates a new pattern renderer
func NewPatternRenderer(registry *ComponentRegistry, schemaRenderer *SchemaRenderer) *PatternRenderer {
	return &PatternRenderer{
		registry:       registry,
		schemaRenderer: schemaRenderer,
		patterns:       make(map[string]*PatternDefinition),
	}
}

// LoadPatterns loads pattern definitions from JSON files
func (pr *PatternRenderer) LoadPatterns(patternFiles map[string]string) error {
	for category, content := range patternFiles {
		var patternFile struct {
			Patterns map[string]*PatternDefinition `json:"patterns"`
		}

		if err := json.Unmarshal([]byte(content), &patternFile); err != nil {
			return fmt.Errorf("failed to parse pattern file %s: %w", category, err)
		}

		for patternID, pattern := range patternFile.Patterns {
			pattern.ID = patternID
			pr.patterns[patternID] = pattern
		}
	}

	return nil
}

// RenderPattern renders a specific pattern with the given data
func (pr *PatternRenderer) RenderPattern(patternID string, data map[string]any, ctx context.Context) (templ.Component, error) {
	pattern, exists := pr.patterns[patternID]
	if !exists {
		return nil, fmt.Errorf("pattern '%s' not found", patternID)
	}

	// Clone the schema to avoid modifying the original
	schema := pr.cloneComponentDefinition(pattern.Schema)

	// Inject data into the schema props
	if schema.Props == nil {
		schema.Props = make(map[string]any)
	}

	// Merge data into props
	for key, value := range data {
		schema.Props[key] = value
	}

	// Process the schema with data interpolation
	pr.interpolateData(schema, data)

	// Render using the schema renderer
	return pr.schemaRenderer.RenderComponent(schema, ctx)
}

// RenderEnhancedDataTable renders an enhanced data table with all features
func (pr *PatternRenderer) RenderEnhancedDataTable(config DataTableConfig, ctx context.Context) (templ.Component, error) {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		// Start table container
		if _, err := w.Write([]byte(`<div class="enhanced-data-table" data-component="enhanced-table">`)); err != nil {
			return err
		}

		// Render batch operations toolbar if enabled
		if config.Features.BatchOperations.Enabled {
			if err := pr.renderBatchToolbar(w, config.Features.BatchOperations); err != nil {
				return err
			}
		}

		// Render filter panel if enabled
		if config.Features.Filtering.Enabled {
			if err := pr.renderFilterPanel(w, config.Features.Filtering); err != nil {
				return err
			}
		}

		// Start table wrapper
		if _, err := w.Write([]byte(`<div class="table-wrapper overflow-auto">`)); err != nil {
			return err
		}

		// Render table header
		if err := pr.renderTableHeader(w, config); err != nil {
			return err
		}

		// Render table body
		if err := pr.renderTableBody(w, config); err != nil {
			return err
		}

		// End table wrapper
		if _, err := w.Write([]byte(`</div>`)); err != nil {
			return err
		}

		// Render pagination if enabled
		if config.Features.Pagination.Enabled {
			if err := pr.renderPagination(w, config.Features.Pagination); err != nil {
				return err
			}
		}

		// End table container
		_, err := w.Write([]byte(`</div>`))
		return err
	}), nil
}

// DataTableConfig represents configuration for enhanced data tables
type DataTableConfig struct {
	Title        string             `json:"title"`
	Subtitle     string             `json:"subtitle"`
	Data         []map[string]any   `json:"data"`
	Columns      []ColumnDefinition `json:"columns"`
	Features     TableFeatures      `json:"features"`
	Styling      TableStyling       `json:"styling"`
	EmptyState   EmptyStateConfig   `json:"emptyState"`
	LoadingState LoadingStateConfig `json:"loadingState"`
}

// ColumnDefinition defines a table column
type ColumnDefinition struct {
	Key        string         `json:"key"`
	Label      string         `json:"label"`
	Type       string         `json:"type"`
	Width      string         `json:"width,omitempty"`
	Sortable   bool           `json:"sortable"`
	Searchable bool           `json:"searchable"`
	Filterable bool           `json:"filterable"`
	Primary    bool           `json:"primary,omitempty"`
	Sticky     bool           `json:"sticky,omitempty"`
	Props      map[string]any `json:"props,omitempty"`
}

// TableFeatures defines enabled table features
type TableFeatures struct {
	Selection           SelectionConfig     `json:"selection"`
	Favorites           FavoritesConfig     `json:"favorites"`
	Sorting             SortingConfig       `json:"sorting"`
	Filtering           FilteringConfig     `json:"filtering"`
	Search              SearchConfig        `json:"search"`
	Pagination          PaginationConfig    `json:"pagination"`
	BatchOperations     BatchOpsConfig      `json:"batchOperations"`
	ColumnCustomization ColumnCustomConfig  `json:"columnCustomization"`
	VirtualScrolling    VirtualScrollConfig `json:"virtualScrolling"`
	RealTimeUpdates     RealTimeConfig      `json:"realTimeUpdates"`
}

// SelectionConfig for row selection
type SelectionConfig struct {
	Enabled           bool   `json:"enabled"`
	Type              string `json:"type"` // "single" or "multiple"
	PreserveSelection bool   `json:"preserveSelection"`
}

// FavoritesConfig for favorite toggles
type FavoritesConfig struct {
	Enabled      bool `json:"enabled"`
	PersistState bool `json:"persistState"`
}

// SortingConfig for column sorting
type SortingConfig struct {
	Enabled     bool             `json:"enabled"`
	MultiColumn bool             `json:"multiColumn"`
	DefaultSort []SortDefinition `json:"defaultSort"`
}

// SortDefinition defines default sorting
type SortDefinition struct {
	Column    string `json:"column"`
	Direction string `json:"direction"`
}

// FilteringConfig for data filtering
type FilteringConfig struct {
	Enabled         bool             `json:"enabled"`
	QuickFilters    []QuickFilter    `json:"quickFilters"`
	AdvancedFilters []AdvancedFilter `json:"advancedFilters"`
}

// QuickFilter for one-click filtering
type QuickFilter struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Icon   string `json:"icon,omitempty"`
	Filter string `json:"filter"`
}

// AdvancedFilter for complex filtering
type AdvancedFilter struct {
	Field   string         `json:"field"`
	Type    string         `json:"type"`
	Options any            `json:"options,omitempty"`
	Props   map[string]any `json:"props,omitempty"`
}

// SearchConfig for table search
type SearchConfig struct {
	Enabled     bool           `json:"enabled"`
	Placeholder string         `json:"placeholder"`
	Fields      []string       `json:"fields"`
	Keyboard    KeyboardConfig `json:"keyboard"`
}

// KeyboardConfig for keyboard shortcuts
type KeyboardConfig struct {
	Shortcut     string `json:"shortcut"`
	GlobalSearch bool   `json:"globalSearch"`
}

// PaginationConfig for table pagination
type PaginationConfig struct {
	Enabled         bool  `json:"enabled"`
	PageSize        int   `json:"pageSize"`
	PageSizeOptions []int `json:"pageSizeOptions"`
	ShowInfo        bool  `json:"showInfo"`
	ShowJumper      bool  `json:"showJumper"`
}

// BatchOpsConfig for batch operations
type BatchOpsConfig struct {
	Enabled  bool          `json:"enabled"`
	Position string        `json:"position"`
	Actions  []BatchAction `json:"actions"`
}

// BatchAction defines a batch operation
type BatchAction struct {
	ID          string           `json:"id"`
	Label       string           `json:"label"`
	Icon        string           `json:"icon"`
	Variant     string           `json:"variant,omitempty"`
	Modal       string           `json:"modal,omitempty"`
	Dropdown    []DropdownOption `json:"dropdown,omitempty"`
	Action      string           `json:"action,omitempty"`
	Confirm     *ConfirmConfig   `json:"confirm,omitempty"`
	Permissions []string         `json:"permissions,omitempty"`
}

// DropdownOption for dropdown actions
type DropdownOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ConfirmConfig for confirmation dialogs
type ConfirmConfig struct {
	Title       string `json:"title"`
	Message     string `json:"message"`
	ConfirmText string `json:"confirmText"`
}

// ColumnCustomConfig for column customization
type ColumnCustomConfig struct {
	Enabled     bool           `json:"enabled"`
	Hideable    bool           `json:"hideable"`
	Resizable   bool           `json:"resizable"`
	Reorderable bool           `json:"reorderable"`
	Presets     []ColumnPreset `json:"presets"`
}

// ColumnPreset defines a column preset
type ColumnPreset struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Columns []string `json:"columns"`
}

// VirtualScrollConfig for virtual scrolling
type VirtualScrollConfig struct {
	Enabled    bool `json:"enabled"`
	Threshold  int  `json:"threshold"`
	ItemHeight int  `json:"itemHeight"`
}

// RealTimeConfig for real-time updates
type RealTimeConfig struct {
	Enabled        bool   `json:"enabled"`
	Websocket      string `json:"websocket"`
	UpdateStrategy string `json:"updateStrategy"`
}

// TableStyling defines table appearance
type TableStyling struct {
	Density      string            `json:"density"`
	Striped      bool              `json:"striped"`
	Bordered     bool              `json:"bordered"`
	Hover        bool              `json:"hover"`
	StickyHeader bool              `json:"stickyHeader"`
	CompactMode  bool              `json:"compactMode"`
	Theme        map[string]string `json:"theme"`
}

// EmptyStateConfig for empty table state
type EmptyStateConfig struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	Actions     []ActionButton `json:"actions"`
}

// LoadingStateConfig for loading state
type LoadingStateConfig struct {
	Type        string `json:"type"`
	Rows        int    `json:"rows"`
	ShowShimmer bool   `json:"showShimmer"`
}

// ActionButton represents an action button
type ActionButton struct {
	Text        string   `json:"text"`
	Action      string   `json:"action"`
	Variant     string   `json:"variant"`
	Permissions []string `json:"permissions,omitempty"`
}

// Helper methods for rendering table components

func (pr *PatternRenderer) renderBatchToolbar(w io.Writer, config BatchOpsConfig) error {
	if _, err := w.Write([]byte(`<div class="batch-operations-toolbar hidden" data-selected-count="0">`)); err != nil {
		return err
	}

	// Selection info
	if _, err := w.Write([]byte(`<div class="selection-info"><span class="selected-count">0</span> items selected</div>`)); err != nil {
		return err
	}

	// Batch actions
	if _, err := w.Write([]byte(`<div class="batch-actions">`)); err != nil {
		return err
	}

	for _, action := range config.Actions {
		actionHTML := fmt.Sprintf(`<button class="btn btn-%s" data-action="%s">`, action.Variant, action.Action)
		if action.Icon != "" {
			actionHTML += fmt.Sprintf(`<i class="icon-%s"></i>`, action.Icon)
		}
		actionHTML += action.Label + `</button>`

		if _, err := w.Write([]byte(actionHTML)); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(`</div>`)); err != nil {
		return err
	}

	_, err := w.Write([]byte(`</div>`))
	return err
}

func (pr *PatternRenderer) renderFilterPanel(w io.Writer, config FilteringConfig) error {
	if _, err := w.Write([]byte(`<div class="filter-panel">`)); err != nil {
		return err
	}

	// Quick filters
	if len(config.QuickFilters) > 0 {
		if _, err := w.Write([]byte(`<div class="quick-filters">`)); err != nil {
			return err
		}

		for _, filter := range config.QuickFilters {
			filterHTML := fmt.Sprintf(`<button class="filter-toggle" data-filter="%s">`, filter.Filter)
			if filter.Icon != "" {
				filterHTML += fmt.Sprintf(`<i class="icon-%s"></i>`, filter.Icon)
			}
			filterHTML += filter.Label + `</button>`

			if _, err := w.Write([]byte(filterHTML)); err != nil {
				return err
			}
		}

		if _, err := w.Write([]byte(`</div>`)); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte(`</div>`))
	return err
}

func (pr *PatternRenderer) renderTableHeader(w io.Writer, config DataTableConfig) error {
	if _, err := w.Write([]byte(`<table class="enhanced-table"><thead><tr>`)); err != nil {
		return err
	}

	for _, column := range config.Columns {
		thClass := "table-header"
		if column.Sortable {
			thClass += " sortable"
		}
		if column.Sticky {
			thClass += " sticky"
		}

		thHTML := fmt.Sprintf(`<th class="%s" data-column="%s"`, thClass, column.Key)
		if column.Width != "" {
			thHTML += fmt.Sprintf(` style="width: %s"`, column.Width)
		}
		thHTML += ">"

		// Column content
		thHTML += column.Label

		if column.Sortable {
			thHTML += `<span class="sort-indicator"></span>`
		}

		thHTML += "</th>"

		if _, err := w.Write([]byte(thHTML)); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte(`</tr></thead>`))
	return err
}

func (pr *PatternRenderer) renderTableBody(w io.Writer, config DataTableConfig) error {
	if _, err := w.Write([]byte(`<tbody>`)); err != nil {
		return err
	}

	if len(config.Data) == 0 {
		// Render empty state
		if err := pr.renderEmptyState(w, config.EmptyState, len(config.Columns)); err != nil {
			return err
		}
	} else {
		// Render data rows
		for i, row := range config.Data {
			if err := pr.renderTableRow(w, row, config.Columns, i); err != nil {
				return err
			}
		}
	}

	_, err := w.Write([]byte(`</tbody></table>`))
	return err
}

func (pr *PatternRenderer) renderTableRow(w io.Writer, row map[string]any, columns []ColumnDefinition, index int) error {
	rowHTML := fmt.Sprintf(`<tr class="table-row" data-row-index="%d">`, index)
	if _, err := w.Write([]byte(rowHTML)); err != nil {
		return err
	}

	for _, column := range columns {
		cellValue := row[column.Key]
		cellHTML := fmt.Sprintf(`<td class="table-cell table-cell-%s">`, column.Type)

		// Render cell content based on type
		switch column.Type {
		case "checkbox":
			cellHTML += `<input type="checkbox" class="row-select">`
		case "favorite-toggle":
			cellHTML += `<button class="favorite-toggle"><i class="icon-star"></i></button>`
		case "avatar":
			if avatar, ok := cellValue.(map[string]any); ok {
				cellHTML += pr.renderAvatar(avatar)
			}
		case "status-badge":
			if status, ok := cellValue.(string); ok {
				cellHTML += pr.renderStatusBadge(status, column.Props)
			}
		case "progress-indicator":
			if progress, ok := cellValue.(float64); ok {
				cellHTML += pr.renderProgressIndicator(progress, column.Props)
			}
		default:
			if cellValue != nil {
				cellHTML += fmt.Sprintf("%v", cellValue)
			}
		}

		cellHTML += "</td>"

		if _, err := w.Write([]byte(cellHTML)); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte(`</tr>`))
	return err
}

func (pr *PatternRenderer) renderEmptyState(w io.Writer, config EmptyStateConfig, colSpan int) error {
	emptyHTML := fmt.Sprintf(`<tr><td colspan="%d" class="empty-state">`, colSpan)
	emptyHTML += `<div class="empty-state-content">`

	if config.Icon != "" {
		emptyHTML += fmt.Sprintf(`<i class="icon-%s empty-state-icon"></i>`, config.Icon)
	}

	emptyHTML += fmt.Sprintf(`<h3>%s</h3>`, config.Title)
	emptyHTML += fmt.Sprintf(`<p>%s</p>`, config.Description)

	if len(config.Actions) > 0 {
		emptyHTML += `<div class="empty-state-actions">`
		for _, action := range config.Actions {
			emptyHTML += fmt.Sprintf(`<button class="btn btn-%s" data-action="%s">%s</button>`,
				action.Variant, action.Action, action.Text)
		}
		emptyHTML += `</div>`
	}

	emptyHTML += `</div></td></tr>`

	_, err := w.Write([]byte(emptyHTML))
	return err
}

func (pr *PatternRenderer) renderPagination(w io.Writer, config PaginationConfig) error {
	paginationHTML := `<div class="pagination-container">`

	if config.ShowInfo {
		paginationHTML += `<div class="pagination-info">Showing 1 to 25 of 100 results</div>`
	}

	paginationHTML += `<div class="pagination-controls">`
	paginationHTML += `<button class="btn btn-ghost" disabled><i class="icon-chevron-left"></i></button>`
	paginationHTML += `<span class="pagination-pages">Page 1 of 4</span>`
	paginationHTML += `<button class="btn btn-ghost"><i class="icon-chevron-right"></i></button>`
	paginationHTML += `</div>`

	if config.ShowJumper {
		paginationHTML += `<div class="page-jumper">`
		paginationHTML += `<input type="number" placeholder="Go to page" class="page-input">`
		paginationHTML += `</div>`
	}

	paginationHTML += `</div>`

	_, err := w.Write([]byte(paginationHTML))
	return err
}

// Helper methods for rendering cell content

func (pr *PatternRenderer) renderAvatar(avatar map[string]any) string {
	src, _ := avatar["src"].(string)
	name, _ := avatar["name"].(string)
	size, _ := avatar["size"].(string)

	avatarHTML := fmt.Sprintf(`<div class="avatar avatar-%s">`, size)
	if src != "" {
		avatarHTML += fmt.Sprintf(`<img src="%s" alt="%s">`, src, name)
	} else if name != "" {
		// Generate initials
		parts := strings.Fields(name)
		initials := ""
		for i, part := range parts {
			if i < 2 && len(part) > 0 {
				initials += string(part[0])
			}
		}
		avatarHTML += fmt.Sprintf(`<span class="avatar-initials">%s</span>`, strings.ToUpper(initials))
	}
	avatarHTML += `</div>`

	return avatarHTML
}

func (pr *PatternRenderer) renderStatusBadge(status string, props map[string]any) string {
	statusConfig, _ := props["statusConfig"].(map[string]any)
	if statusConfig == nil {
		return fmt.Sprintf(`<span class="badge">%s</span>`, status)
	}

	config, exists := statusConfig[status].(map[string]any)
	if !exists {
		return fmt.Sprintf(`<span class="badge">%s</span>`, status)
	}

	label, _ := config["label"].(string)
	color, _ := config["color"].(string)
	icon, _ := config["icon"].(string)

	badgeHTML := fmt.Sprintf(`<span class="badge badge-%s">`, color)
	if icon != "" {
		badgeHTML += fmt.Sprintf(`<i class="icon-%s"></i>`, icon)
	}
	if label != "" {
		badgeHTML += label
	} else {
		badgeHTML += status
	}
	badgeHTML += `</span>`

	return badgeHTML
}

func (pr *PatternRenderer) renderProgressIndicator(progress float64, props map[string]any) string {
	showPercentage, _ := props["showPercentage"].(bool)
	showBar, _ := props["showBar"].(bool)

	progressHTML := `<div class="progress-indicator">`

	if showBar {
		progressHTML += fmt.Sprintf(`<div class="progress-bar"><div class="progress-fill" style="width: %.1f%%"></div></div>`, progress)
	}

	if showPercentage {
		progressHTML += fmt.Sprintf(`<span class="progress-text">%.1f%%</span>`, progress)
	}

	progressHTML += `</div>`

	return progressHTML
}

// Helper methods

func (pr *PatternRenderer) cloneComponentDefinition(def *schemas.ComponentDefinition) *schemas.ComponentDefinition {
	clone := &schemas.ComponentDefinition{
		ID:          def.ID,
		Type:        def.Type,
		DataSource:  def.DataSource,
		Layout:      def.Layout,
		Conditions:  def.Conditions,
		Permissions: def.Permissions,
	}

	if def.Props != nil {
		clone.Props = make(map[string]any)
		for k, v := range def.Props {
			clone.Props[k] = v
		}
	}

	if len(def.Children) > 0 {
		clone.Children = make([]schemas.ComponentDefinition, len(def.Children))
		for i, child := range def.Children {
			clone.Children[i] = *pr.cloneComponentDefinition(&child)
		}
	}

	return clone
}

func (pr *PatternRenderer) interpolateData(def *schemas.ComponentDefinition, data map[string]any) {
	if def.Props != nil {
		pr.interpolateMap(def.Props, data)
	}

	for i := range def.Children {
		pr.interpolateData(&def.Children[i], data)
	}
}

func (pr *PatternRenderer) interpolateMap(m map[string]any, data map[string]any) {
	for key, value := range m {
		switch v := value.(type) {
		case string:
			m[key] = pr.interpolateString(v, data)
		case map[string]any:
			pr.interpolateMap(v, data)
		case []any:
			pr.interpolateSlice(v, data)
		}
	}
}

func (pr *PatternRenderer) interpolateSlice(s []any, data map[string]any) {
	for i, value := range s {
		switch v := value.(type) {
		case string:
			s[i] = pr.interpolateString(v, data)
		case map[string]any:
			pr.interpolateMap(v, data)
		case []any:
			pr.interpolateSlice(v, data)
		}
	}
}

func (pr *PatternRenderer) interpolateString(template string, data map[string]any) string {
	// Simple template interpolation for {key} patterns
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}
	return result
}
