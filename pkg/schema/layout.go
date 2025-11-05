package schema

import (
	"context"
	"fmt"
	"sort"

	"github.com/niiniyare/erp/pkg/condition"
)

// Layout defines the visual structure and arrangement of form elements
type Layout struct {
	Type LayoutType `json:"type" validate:"required" example:"grid"` // Layout algorithm

	// Grid/Flex configuration
	Columns    int    `json:"columns,omitempty" validate:"min=1,max=24" example:"3"` // Number of columns
	Gap        string `json:"gap,omitempty" validate:"css_size" example:"1rem"`      // Space between items
	Direction  string `json:"direction,omitempty" validate:"oneof=row column row-reverse column-reverse" example:"column"`
	Wrap       bool   `json:"wrap,omitempty"`       // Allow wrapping
	Responsive bool   `json:"responsive,omitempty"` // Enable responsive behavior

	// Complex layouts
	Sections []Section `json:"sections,omitempty" validate:"dive"` // Logical sections
	Groups   []Group   `json:"groups,omitempty" validate:"dive"`   // Field groups
	Tabs     []Tab     `json:"tabs,omitempty" validate:"dive"`     // Tabbed interface
	Steps    []Step    `json:"steps,omitempty" validate:"dive"`    // Multi-step wizard

	// Responsive breakpoints
	Breakpoints *Breakpoints `json:"breakpoints,omitempty"` // Screen size configurations

	// Theme
	Theme *LayoutTheme `json:"theme,omitempty"` // Layout-specific theming

	// Internal (not serialized)
	evaluator *condition.Evaluator `json:"-"` // Condition evaluator
}

// LayoutType defines the layout algorithm
type LayoutType string

const (
	LayoutGrid     LayoutType = "grid"     // CSS Grid layout
	LayoutFlex     LayoutType = "flex"     // Flexbox layout
	LayoutTabs     LayoutType = "tabs"     // Tabbed interface
	LayoutSteps    LayoutType = "steps"    // Multi-step wizard
	LayoutSections LayoutType = "sections" // Divided into sections
	LayoutGroups   LayoutType = "groups"   // Field grouping
)

// Section represents a logical grouping of fields with a title
type Section struct {
	ID          string                    `json:"id" validate:"required"`
	Title       string                    `json:"title,omitempty"`
	Description string                    `json:"description,omitempty"`
	Icon        string                    `json:"icon,omitempty" validate:"icon_name"`
	Fields      []string                  `json:"fields" validate:"required,dive,fieldname"` // Field names
	Collapsible bool                      `json:"collapsible,omitempty"`                     // Can be collapsed
	Collapsed   bool                      `json:"collapsed,omitempty"`                       // Initially collapsed
	Columns     int                       `json:"columns,omitempty" validate:"min=1,max=12"` // Section columns
	Order       int                       `json:"order,omitempty"`                           // Display order
	Conditional *Conditional              `json:"conditional,omitempty"`                     // Show/hide conditions (legacy)
	Condition   *condition.ConditionGroup `json:"condition,omitempty"`                       // Advanced conditions
	Style       *Style                    `json:"style,omitempty"`                           // Custom styling
	Theme       *SectionTheme             `json:"theme,omitempty"`                           // Section-specific theme
}

// Group represents a visual grouping of fields (like a fieldset)
type Group struct {
	ID          string                    `json:"id" validate:"required"`
	Label       string                    `json:"label,omitempty"`
	Description string                    `json:"description,omitempty"`
	Fields      []string                  `json:"fields" validate:"required,dive,fieldname"` // Field names
	Border      bool                      `json:"border,omitempty"`                          // Show border
	Columns     int                       `json:"columns,omitempty" validate:"min=1,max=12"` // Group columns
	Order       int                       `json:"order,omitempty"`                           // Display order
	Conditional *Conditional              `json:"conditional,omitempty"`                     // Show/hide conditions (legacy)
	Condition   *condition.ConditionGroup `json:"condition,omitempty"`                       // Advanced conditions
	Style       *Style                    `json:"style,omitempty"`                           // Custom styling
	Theme       *GroupTheme               `json:"theme,omitempty"`                           // Group-specific theme
}

// Tab represents a tab in a tabbed interface
type Tab struct {
	ID          string                    `json:"id" validate:"required"`
	Label       string                    `json:"label" validate:"required"`
	Icon        string                    `json:"icon,omitempty" validate:"icon_name"`
	Description string                    `json:"description,omitempty"`
	Fields      []string                  `json:"fields" validate:"required,dive,fieldname"` // Field names in tab
	Badge       string                    `json:"badge,omitempty"`                           // Badge text
	Disabled    bool                      `json:"disabled,omitempty"`                        // Tab disabled
	Order       int                       `json:"order,omitempty"`                           // Tab order
	Conditional *Conditional              `json:"conditional,omitempty"`                     // Show/hide conditions (legacy)
	Condition   *condition.ConditionGroup `json:"condition,omitempty"`                       // Advanced conditions
	Style       *Style                    `json:"style,omitempty"`                           // Custom styling
	Theme       *TabTheme                 `json:"theme,omitempty"`                           // Tab-specific theme
}

// Step represents a step in a multi-step wizard
type Step struct {
	ID          string                    `json:"id" validate:"required"`
	Title       string                    `json:"title" validate:"required"`
	Description string                    `json:"description,omitempty"`
	Icon        string                    `json:"icon,omitempty" validate:"icon_name"`
	Fields      []string                  `json:"fields" validate:"required,dive,fieldname"` // Fields in this step
	Order       int                       `json:"order" validate:"min=0"`                    // Step number
	Skippable   bool                      `json:"skippable,omitempty"`                       // Can skip this step
	Validation  bool                      `json:"validation,omitempty"`                      // Validate before next
	Conditional *Conditional              `json:"conditional,omitempty"`                     // Show/hide conditions (legacy)
	Condition   *condition.ConditionGroup `json:"condition,omitempty"`                       // Advanced conditions
	Style       *Style                    `json:"style,omitempty"`                           // Custom styling
	Theme       *StepTheme                `json:"theme,omitempty"`                           // Step-specific theme
}

// Breakpoints defines responsive behavior at different screen sizes
type Breakpoints struct {
	Mobile  *BreakpointConfig            `json:"mobile,omitempty"`  // < 640px
	Tablet  *BreakpointConfig            `json:"tablet,omitempty"`  // 640px - 1024px
	Desktop *BreakpointConfig            `json:"desktop,omitempty"` // > 1024px
	Custom  map[string]*BreakpointConfig `json:"custom,omitempty"`  // Custom breakpoints
}

// BreakpointConfig defines layout at specific breakpoint
type BreakpointConfig struct {
	Columns    int      `json:"columns,omitempty" validate:"min=1,max=24"` // Columns at this size
	Gap        string   `json:"gap,omitempty" validate:"css_size"`         // Gap at this size
	Direction  string   `json:"direction,omitempty" validate:"oneof=row column"`
	HideFields []string `json:"hideFields,omitempty"` // Fields to hide at this size
	ShowFields []string `json:"showFields,omitempty"` // Fields to show only at this size
}

// Theme types for layout components
type LayoutTheme struct {
	Colors       map[string]string `json:"colors,omitempty"`
	Spacing      map[string]string `json:"spacing,omitempty"`
	BorderRadius string            `json:"borderRadius,omitempty"`
	CustomCSS    string            `json:"customCSS,omitempty"`
}

type SectionTheme struct {
	Background   string            `json:"background,omitempty"`
	Border       string            `json:"border,omitempty"`
	Padding      string            `json:"padding,omitempty"`
	BorderRadius string            `json:"borderRadius,omitempty"`
	Colors       map[string]string `json:"colors,omitempty"`
	CustomCSS    string            `json:"customCSS,omitempty"`
}

type GroupTheme struct {
	Background   string            `json:"background,omitempty"`
	Border       string            `json:"border,omitempty"`
	Padding      string            `json:"padding,omitempty"`
	BorderRadius string            `json:"borderRadius,omitempty"`
	Colors       map[string]string `json:"colors,omitempty"`
	CustomCSS    string            `json:"customCSS,omitempty"`
}

type TabTheme struct {
	ActiveBackground   string            `json:"activeBackground,omitempty"`
	InactiveBackground string            `json:"inactiveBackground,omitempty"`
	ActiveColor        string            `json:"activeColor,omitempty"`
	InactiveColor      string            `json:"inactiveColor,omitempty"`
	BorderColor        string            `json:"borderColor,omitempty"`
	Colors             map[string]string `json:"colors,omitempty"`
	CustomCSS          string            `json:"customCSS,omitempty"`
}

type StepTheme struct {
	ActiveColor    string            `json:"activeColor,omitempty"`
	CompletedColor string            `json:"completedColor,omitempty"`
	InactiveColor  string            `json:"inactiveColor,omitempty"`
	ConnectorColor string            `json:"connectorColor,omitempty"`
	Colors         map[string]string `json:"colors,omitempty"`
	CustomCSS      string            `json:"customCSS,omitempty"`
}

// SetEvaluator sets the condition evaluator for the layout
func (l *Layout) SetEvaluator(evaluator *condition.Evaluator) {
	l.evaluator = evaluator
}

// GetEvaluator returns the layout's condition evaluator
func (l *Layout) GetEvaluator() *condition.Evaluator {
	return l.evaluator
}

// GetColumns returns the appropriate column count for the layout
func (l *Layout) GetColumns() int {
	if l.Columns > 0 {
		return l.Columns
	}
	// Default columns by type
	switch l.Type {
	case LayoutGrid:
		return 2
	case LayoutFlex:
		return 1
	default:
		return 1
	}
}

// GetGap returns the gap value with fallback
func (l *Layout) GetGap() string {
	if l.Gap != "" {
		return l.Gap
	}
	return GetSpacing("md") // Use design token
}

// GetDirection returns the direction with fallback
func (l *Layout) GetDirection() string {
	if l.Direction != "" {
		return l.Direction
	}
	return "column" // Default direction
}

// HasTabs checks if layout uses tabs
func (l *Layout) HasTabs() bool {
	return l.Type == LayoutTabs && len(l.Tabs) > 0
}

// HasSteps checks if layout is a multi-step wizard
func (l *Layout) HasSteps() bool {
	return l.Type == LayoutSteps && len(l.Steps) > 0
}

// HasSections checks if layout has sections
func (l *Layout) HasSections() bool {
	return (l.Type == LayoutSections && len(l.Sections) > 0) || len(l.Sections) > 0
}

// HasGroups checks if layout has groups
func (l *Layout) HasGroups() bool {
	return (l.Type == LayoutGroups && len(l.Groups) > 0) || len(l.Groups) > 0
}

// GetFieldsForSection returns field names for a specific section
func (l *Layout) GetFieldsForSection(sectionID string) []string {
	for _, section := range l.Sections {
		if section.ID == sectionID {
			return section.Fields
		}
	}
	return []string{}
}

// GetFieldsForTab returns field names for a specific tab
func (l *Layout) GetFieldsForTab(tabID string) []string {
	for _, tab := range l.Tabs {
		if tab.ID == tabID {
			return tab.Fields
		}
	}
	return []string{}
}

// GetFieldsForStep returns field names for a specific step
func (l *Layout) GetFieldsForStep(stepID string) []string {
	for _, step := range l.Steps {
		if step.ID == stepID {
			return step.Fields
		}
	}
	return []string{}
}

// GetFieldsForGroup returns field names for a specific group
func (l *Layout) GetFieldsForGroup(groupID string) []string {
	for _, group := range l.Groups {
		if group.ID == groupID {
			return group.Fields
		}
	}
	return []string{}
}

// GetSection returns a section by ID
func (l *Layout) GetSection(sectionID string) (*Section, bool) {
	for i := range l.Sections {
		if l.Sections[i].ID == sectionID {
			return &l.Sections[i], true
		}
	}
	return nil, false
}

// GetTab returns a tab by ID
func (l *Layout) GetTab(tabID string) (*Tab, bool) {
	for i := range l.Tabs {
		if l.Tabs[i].ID == tabID {
			return &l.Tabs[i], true
		}
	}
	return nil, false
}

// GetStep returns a step by ID
func (l *Layout) GetStep(stepID string) (*Step, bool) {
	for i := range l.Steps {
		if l.Steps[i].ID == stepID {
			return &l.Steps[i], true
		}
	}
	return nil, false
}

// GetGroup returns a group by ID
func (l *Layout) GetGroup(groupID string) (*Group, bool) {
	for i := range l.Groups {
		if l.Groups[i].ID == groupID {
			return &l.Groups[i], true
		}
	}
	return nil, false
}

// GetOrderedSteps returns steps sorted by order
func (l *Layout) GetOrderedSteps() []Step {
	steps := make([]Step, len(l.Steps))
	copy(steps, l.Steps)

	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Order < steps[j].Order
	})

	return steps
}

// GetOrderedSections returns sections sorted by order
func (l *Layout) GetOrderedSections() []Section {
	sections := make([]Section, len(l.Sections))
	copy(sections, l.Sections)

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Order < sections[j].Order
	})

	return sections
}

// GetOrderedGroups returns groups sorted by order
func (l *Layout) GetOrderedGroups() []Group {
	groups := make([]Group, len(l.Groups))
	copy(groups, l.Groups)

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Order < groups[j].Order
	})

	return groups
}

// GetOrderedTabs returns tabs sorted by order
func (l *Layout) GetOrderedTabs() []Tab {
	tabs := make([]Tab, len(l.Tabs))
	copy(tabs, l.Tabs)

	sort.Slice(tabs, func(i, j int) bool {
		return tabs[i].Order < tabs[j].Order
	})

	return tabs
}

// GetVisibleSections returns sections visible for given data context
func (l *Layout) GetVisibleSections(ctx context.Context, data map[string]any) []Section {
	visible := make([]Section, 0)
	for _, section := range l.Sections {
		if isVisible, _ := l.isSectionVisible(ctx, &section, data); isVisible {
			visible = append(visible, section)
		}
	}
	return visible
}

// GetVisibleTabs returns tabs visible for given data context
func (l *Layout) GetVisibleTabs(ctx context.Context, data map[string]any) []Tab {
	visible := make([]Tab, 0)
	for _, tab := range l.Tabs {
		if isVisible, _ := l.isTabVisible(ctx, &tab, data); isVisible && !tab.Disabled {
			visible = append(visible, tab)
		}
	}
	return visible
}

// GetVisibleSteps returns steps visible for given data context
func (l *Layout) GetVisibleSteps(ctx context.Context, data map[string]any) []Step {
	visible := make([]Step, 0)
	for _, step := range l.Steps {
		if isVisible, _ := l.isStepVisible(ctx, &step, data); isVisible {
			visible = append(visible, step)
		}
	}
	return visible
}

// isSectionVisible checks if section is visible
func (l *Layout) isSectionVisible(ctx context.Context, section *Section, data map[string]any) (bool, error) {
	if section.Condition != nil && l.evaluator != nil {
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		return l.evaluator.Evaluate(ctx, section.Condition, evalCtx)
	}
	return true, nil
}

// isTabVisible checks if tab is visible
func (l *Layout) isTabVisible(ctx context.Context, tab *Tab, data map[string]any) (bool, error) {
	if tab.Condition != nil && l.evaluator != nil {
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		return l.evaluator.Evaluate(ctx, tab.Condition, evalCtx)
	}
	return true, nil
}

// isStepVisible checks if step is visible
func (l *Layout) isStepVisible(ctx context.Context, step *Step, data map[string]any) (bool, error) {
	if step.Condition != nil && l.evaluator != nil {
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		return l.evaluator.Evaluate(ctx, step.Condition, evalCtx)
	}
	return true, nil
}

// isGroupVisible checks if group is visible
func (l *Layout) isGroupVisible(ctx context.Context, group *Group, data map[string]any) (bool, error) {
	if group.Condition != nil && l.evaluator != nil {
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		return l.evaluator.Evaluate(ctx, group.Condition, evalCtx)
	}
	return true, nil
}

// ValidateLayout checks if layout configuration is valid
func (l *Layout) ValidateLayout(schema *Schema) error {
	collector := NewErrorCollector()

	// Collect all field names used in layout
	usedFields := make(map[string]bool)

	// Check sections
	for i, section := range l.Sections {
		if section.ID == "" {
			collector.AddValidationError(
				fmt.Sprintf("Sections[%d]", i),
				"missing_id",
				"section ID is required",
			)
		}
		for _, fieldName := range section.Fields {
			if !schema.HasField(fieldName) {
				collector.AddValidationError(
					"Section."+section.ID,
					"invalid_field_reference",
					"references non-existent field: "+fieldName,
				)
			}
			usedFields[fieldName] = true
		}
	}

	// Check tabs
	for i, tab := range l.Tabs {
		if tab.ID == "" {
			collector.AddValidationError(
				fmt.Sprintf("Tabs[%d]", i),
				"missing_id",
				"tab ID is required",
			)
		}
		if tab.Label == "" {
			collector.AddValidationError(
				"Tab."+tab.ID,
				"missing_label",
				"tab label is required",
			)
		}
		for _, fieldName := range tab.Fields {
			if !schema.HasField(fieldName) {
				collector.AddValidationError(
					"Tab."+tab.ID,
					"invalid_field_reference",
					"references non-existent field: "+fieldName,
				)
			}
			usedFields[fieldName] = true
		}
	}

	// Check steps
	for i, step := range l.Steps {
		if step.ID == "" {
			collector.AddValidationError(
				fmt.Sprintf("Steps[%d]", i),
				"missing_id",
				"step ID is required",
			)
		}
		if step.Title == "" {
			collector.AddValidationError(
				"Step."+step.ID,
				"missing_title",
				"step title is required",
			)
		}
		for _, fieldName := range step.Fields {
			if !schema.HasField(fieldName) {
				collector.AddValidationError(
					"Step."+step.ID,
					"invalid_field_reference",
					"references non-existent field: "+fieldName,
				)
			}
			usedFields[fieldName] = true
		}
	}

	// Check groups
	for i, group := range l.Groups {
		if group.ID == "" {
			collector.AddValidationError(
				fmt.Sprintf("Groups[%d]", i),
				"missing_id",
				"group ID is required",
			)
		}
		for _, fieldName := range group.Fields {
			if !schema.HasField(fieldName) {
				collector.AddValidationError(
					"Group."+group.ID,
					"invalid_field_reference",
					"references non-existent field: "+fieldName,
				)
			}
			usedFields[fieldName] = true
		}
	}

	// Check for duplicate IDs
	l.checkDuplicateIDs(collector)

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// checkDuplicateIDs validates that all IDs are unique
func (l *Layout) checkDuplicateIDs(collector *ErrorCollector) {
	ids := make(map[string]bool)

	for _, section := range l.Sections {
		if ids[section.ID] {
			collector.AddValidationError(
				"Section."+section.ID,
				"duplicate_id",
				"duplicate section ID: "+section.ID,
			)
		}
		ids[section.ID] = true
	}

	for _, tab := range l.Tabs {
		if ids[tab.ID] {
			collector.AddValidationError(
				"Tab."+tab.ID,
				"duplicate_id",
				"duplicate tab ID: "+tab.ID,
			)
		}
		ids[tab.ID] = true
	}

	for _, step := range l.Steps {
		if ids[step.ID] {
			collector.AddValidationError(
				"Step."+step.ID,
				"duplicate_id",
				"duplicate step ID: "+step.ID,
			)
		}
		ids[step.ID] = true
	}

	for _, group := range l.Groups {
		if ids[group.ID] {
			collector.AddValidationError(
				"Group."+group.ID,
				"duplicate_id",
				"duplicate group ID: "+group.ID,
			)
		}
		ids[group.ID] = true
	}
}

// GetAllFields returns all field names referenced in the layout
func (l *Layout) GetAllFields() []string {
	fieldMap := make(map[string]bool)

	for _, section := range l.Sections {
		for _, field := range section.Fields {
			fieldMap[field] = true
		}
	}

	for _, tab := range l.Tabs {
		for _, field := range tab.Fields {
			fieldMap[field] = true
		}
	}

	for _, step := range l.Steps {
		for _, field := range step.Fields {
			fieldMap[field] = true
		}
	}

	for _, group := range l.Groups {
		for _, field := range group.Fields {
			fieldMap[field] = true
		}
	}

	fields := make([]string, 0, len(fieldMap))
	for field := range fieldMap {
		fields = append(fields, field)
	}

	return fields
}

// Clone creates a copy of the layout with evaluator preservation
func (l *Layout) Clone() *Layout {
	clone := *l
	clone.evaluator = l.evaluator
	return &clone
}

// ApplyTheme applies theme settings to the layout
func (l *Layout) ApplyTheme(theme *Theme) {
	if theme == nil {
		return
	}

	// Apply theme overrides
	if l.Theme == nil {
		l.Theme = &LayoutTheme{}
	}

	// Theme will be applied by renderer
}

// GetBreakpointColumns returns columns for a specific breakpoint
func (l *Layout) GetBreakpointColumns(breakpoint string) int {
	if l.Breakpoints == nil {
		return l.GetColumns()
	}

	var config *BreakpointConfig
	switch breakpoint {
	case "mobile":
		config = l.Breakpoints.Mobile
	case "tablet":
		config = l.Breakpoints.Tablet
	case "desktop":
		config = l.Breakpoints.Desktop
	default:
		if l.Breakpoints.Custom != nil {
			config = l.Breakpoints.Custom[breakpoint]
		}
	}

	if config != nil && config.Columns > 0 {
		return config.Columns
	}

	return l.GetColumns()
}

// GetBreakpointGap returns gap for a specific breakpoint
func (l *Layout) GetBreakpointGap(breakpoint string) string {
	if l.Breakpoints == nil {
		return l.GetGap()
	}

	var config *BreakpointConfig
	switch breakpoint {
	case "mobile":
		config = l.Breakpoints.Mobile
	case "tablet":
		config = l.Breakpoints.Tablet
	case "desktop":
		config = l.Breakpoints.Desktop
	default:
		if l.Breakpoints.Custom != nil {
			config = l.Breakpoints.Custom[breakpoint]
		}
	}

	if config != nil && config.Gap != "" {
		return config.Gap
	}

	return l.GetGap()
}
