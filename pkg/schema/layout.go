package schema

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
	ID          string       `json:"id" validate:"required"`
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Icon        string       `json:"icon,omitempty" validate:"icon_name"`
	Fields      []string     `json:"fields" validate:"required,dive,fieldname"` // Field names
	Collapsible bool         `json:"collapsible,omitempty"`                     // Can be collapsed
	Collapsed   bool         `json:"collapsed,omitempty"`                       // Initially collapsed
	Columns     int          `json:"columns,omitempty" validate:"min=1,max=12"` // Section columns
	Order       int          `json:"order,omitempty"`                           // Display order
	Conditional *Conditional `json:"conditional,omitempty"`                     // Show/hide conditions
	Style       *Style       `json:"style,omitempty"`                           // Custom styling
}

// Group represents a visual grouping of fields (like a fieldset)
type Group struct {
	ID          string       `json:"id" validate:"required"`
	Label       string       `json:"label,omitempty"`
	Description string       `json:"description,omitempty"`
	Fields      []string     `json:"fields" validate:"required,dive,fieldname"` // Field names
	Border      bool         `json:"border,omitempty"`                          // Show border
	Columns     int          `json:"columns,omitempty" validate:"min=1,max=12"` // Group columns
	Order       int          `json:"order,omitempty"`                           // Display order
	Conditional *Conditional `json:"conditional,omitempty"`                     // Show/hide conditions
	Style       *Style       `json:"style,omitempty"`                           // Custom styling
}

// Tab represents a tab in a tabbed interface
type Tab struct {
	ID          string       `json:"id" validate:"required"`
	Label       string       `json:"label" validate:"required"`
	Icon        string       `json:"icon,omitempty" validate:"icon_name"`
	Description string       `json:"description,omitempty"`
	Fields      []string     `json:"fields" validate:"required,dive,fieldname"` // Field names in tab
	Badge       string       `json:"badge,omitempty"`                           // Badge text
	Disabled    bool         `json:"disabled,omitempty"`                        // Tab disabled
	Order       int          `json:"order,omitempty"`                           // Tab order
	Conditional *Conditional `json:"conditional,omitempty"`                     // Show/hide conditions
	Style       *Style       `json:"style,omitempty"`                           // Custom styling
}

// Step represents a step in a multi-step wizard
type Step struct {
	ID          string       `json:"id" validate:"required"`
	Title       string       `json:"title" validate:"required"`
	Description string       `json:"description,omitempty"`
	Icon        string       `json:"icon,omitempty" validate:"icon_name"`
	Fields      []string     `json:"fields" validate:"required,dive,fieldname"` // Fields in this step
	Order       int          `json:"order" validate:"min=0"`                    // Step number
	Skippable   bool         `json:"skippable,omitempty"`                       // Can skip this step
	Validation  bool         `json:"validation,omitempty"`                      // Validate before next
	Conditional *Conditional `json:"conditional,omitempty"`                     // Show/hide conditions
	Style       *Style       `json:"style,omitempty"`                           // Custom styling
}

// Breakpoints defines responsive behavior at different screen sizes
type Breakpoints struct {
	Mobile  *BreakpointConfig `json:"mobile,omitempty"`  // < 640px
	Tablet  *BreakpointConfig `json:"tablet,omitempty"`  // 640px - 1024px
	Desktop *BreakpointConfig `json:"desktop,omitempty"` // > 1024px
}

// BreakpointConfig defines layout at specific breakpoint
type BreakpointConfig struct {
	Columns    int      `json:"columns,omitempty" validate:"min=1,max=24"` // Columns at this size
	Gap        string   `json:"gap,omitempty" validate:"css_size"`         // Gap at this size
	Direction  string   `json:"direction,omitempty" validate:"oneof=row column"`
	HideFields []string `json:"hideFields,omitempty"` // Fields to hide at this size
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
	return "1rem" // Default gap
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

// GetOrderedSteps returns steps sorted by order
func (l *Layout) GetOrderedSteps() []Step {
	steps := make([]Step, len(l.Steps))
	copy(steps, l.Steps)
	// Simple bubble sort by order
	for i := 0; i < len(steps)-1; i++ {
		for j := 0; j < len(steps)-i-1; j++ {
			if steps[j].Order > steps[j+1].Order {
				steps[j], steps[j+1] = steps[j+1], steps[j]
			}
		}
	}
	return steps
}

// ValidateLayout checks if layout configuration is valid
func (l *Layout) ValidateLayout(schema *Schema) error {
	collector := NewErrorCollector()

	// Collect all field names used in layout
	usedFields := make(map[string]bool)

	// Check sections
	for _, section := range l.Sections {
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
	for _, tab := range l.Tabs {
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
	for _, step := range l.Steps {
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
	for _, group := range l.Groups {
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

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}
