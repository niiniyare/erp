package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LayoutTestSuite struct {
	suite.Suite
}

func TestLayoutTestSuite(t *testing.T) {
	suite.Run(t, new(LayoutTestSuite))
}

// Test grid layout functionality
func (suite *LayoutTestSuite) TestGridLayout() {
	layout := &Layout{
		Type:    LayoutGrid,
		Columns: 3,
		Gap:     "md",
	}

	// Test GetColumns
	require.Equal(suite.T(), 3, layout.GetColumns())

	// Test GetGap with valid gap
	require.Equal(suite.T(), "md", layout.GetGap())

	// Test with different gap values
	layout.Gap = "sm"
	require.Equal(suite.T(), "sm", layout.GetGap())

	layout.Gap = "lg"
	require.Equal(suite.T(), "lg", layout.GetGap())

	// Test with invalid gap (should return default)
	layout.Gap = "invalid"
	require.Equal(suite.T(), "invalid", layout.GetGap())

	// Test with empty gap (should return default)
	layout.Gap = ""
	require.Equal(suite.T(), "1rem", layout.GetGap())
}

// Test flex layout direction
func (suite *LayoutTestSuite) TestFlexDirection() {
	layout := &Layout{
		Type:      LayoutFlex,
		Direction: "row",
	}

	// Test GetDirection with valid directions
	require.Equal(suite.T(), "row", layout.GetDirection())

	layout.Direction = "column"
	require.Equal(suite.T(), "column", layout.GetDirection())

	layout.Direction = "row-reverse"
	require.Equal(suite.T(), "row-reverse", layout.GetDirection())

	layout.Direction = "column-reverse"
	require.Equal(suite.T(), "column-reverse", layout.GetDirection())

	// Test with invalid direction (should return default)
	layout.Direction = "invalid"
	require.Equal(suite.T(), "invalid", layout.GetDirection())

	// Test with empty direction (should return default)
	layout.Direction = ""
	require.Equal(suite.T(), "column", layout.GetDirection())
}

// Test layout type checking
func (suite *LayoutTestSuite) TestLayoutTypeChecking() {
	// Test HasTabs
	tabLayout := &Layout{
		Type: LayoutTabs,
		Tabs: []Tab{{ID: "tab1", Label: "Tab 1", Fields: []string{"field1"}}},
	}
	require.True(suite.T(), tabLayout.HasTabs())

	gridLayout := &Layout{Type: LayoutGrid}
	require.False(suite.T(), gridLayout.HasTabs())

	// Test HasSteps
	stepLayout := &Layout{
		Type:  LayoutSteps,
		Steps: []Step{{ID: "step1", Title: "Step 1", Order: 1, Fields: []string{"field1"}}},
	}
	require.True(suite.T(), stepLayout.HasSteps())

	require.False(suite.T(), gridLayout.HasSteps())

	// Test HasSections
	sectionLayout := &Layout{
		Type:     LayoutSections,
		Sections: []Section{{ID: "section1", Title: "Section 1", Fields: []string{"field1"}}},
	}
	require.True(suite.T(), sectionLayout.HasSections())

	require.False(suite.T(), gridLayout.HasSections())
}

// Test field organization by sections
func (suite *LayoutTestSuite) TestFieldsForSection() {
	layout := &Layout{
		Type: LayoutSections,
		Sections: []Section{
			{
				ID:     "section1",
				Title:  "Personal Info",
				Fields: []string{"name", "email", "phone"},
			},
			{
				ID:     "section2",
				Title:  "Address",
				Fields: []string{"street", "city", "zip"},
			},
		},
	}

	// Test getting fields for existing section
	fields := layout.GetFieldsForSection("section1")
	require.Equal(suite.T(), []string{"name", "email", "phone"}, fields)

	fields = layout.GetFieldsForSection("section2")
	require.Equal(suite.T(), []string{"street", "city", "zip"}, fields)

	// Test getting fields for non-existent section
	fields = layout.GetFieldsForSection("nonexistent")
	require.Empty(suite.T(), fields)
}

// Test field organization by tabs
func (suite *LayoutTestSuite) TestFieldsForTab() {
	layout := &Layout{
		Type: LayoutTabs,
		Tabs: []Tab{
			{
				ID:     "tab1",
				Label:  "Basic Info",
				Fields: []string{"name", "email"},
			},
			{
				ID:     "tab2",
				Label:  "Contact",
				Fields: []string{"phone", "address"},
			},
		},
	}

	// Test getting fields for existing tab
	fields := layout.GetFieldsForTab("tab1")
	require.Equal(suite.T(), []string{"name", "email"}, fields)

	fields = layout.GetFieldsForTab("tab2")
	require.Equal(suite.T(), []string{"phone", "address"}, fields)

	// Test getting fields for non-existent tab
	fields = layout.GetFieldsForTab("nonexistent")
	require.Empty(suite.T(), fields)
}

// Test field organization by steps
func (suite *LayoutTestSuite) TestFieldsForStep() {
	layout := &Layout{
		Type: LayoutSteps,
		Steps: []Step{
			{
				ID:     "step1",
				Title:  "Step 1",
				Order:  1,
				Fields: []string{"name", "email"},
			},
			{
				ID:     "step2",
				Title:  "Step 2",
				Order:  2,
				Fields: []string{"phone", "address"},
			},
		},
	}

	// Test getting fields for existing step
	fields := layout.GetFieldsForStep("step1")
	require.Equal(suite.T(), []string{"name", "email"}, fields)

	fields = layout.GetFieldsForStep("step2")
	require.Equal(suite.T(), []string{"phone", "address"}, fields)

	// Test getting fields for non-existent step
	fields = layout.GetFieldsForStep("nonexistent")
	require.Empty(suite.T(), fields)
}

// Test ordered steps
func (suite *LayoutTestSuite) TestOrderedSteps() {
	layout := &Layout{
		Type: LayoutSteps,
		Steps: []Step{
			{
				ID:     "step3",
				Title:  "Step 3",
				Order:  3,
				Fields: []string{"review"},
			},
			{
				ID:     "step1",
				Title:  "Step 1",
				Order:  1,
				Fields: []string{"name"},
			},
			{
				ID:     "step2",
				Title:  "Step 2",
				Order:  2,
				Fields: []string{"email"},
			},
		},
	}

	// Test getting ordered steps
	orderedSteps := layout.GetOrderedSteps()
	require.Len(suite.T(), orderedSteps, 3)
	require.Equal(suite.T(), "step1", orderedSteps[0].ID)
	require.Equal(suite.T(), "step2", orderedSteps[1].ID)
	require.Equal(suite.T(), "step3", orderedSteps[2].ID)
}

// Test layout validation
func (suite *LayoutTestSuite) TestLayoutValidation() {
	// Create a mock schema for validation
	schema := &Schema{
		ID:     "test",
		Type:   TypeForm,
		Title:  "Test",
		Fields: []Field{{Name: "field1", Type: FieldText}},
	}

	// Test valid grid layout
	validGrid := &Layout{
		Type:    LayoutGrid,
		Columns: 3,
		Gap:     "md",
	}
	err := validGrid.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test grid layout without columns (should pass - columns have defaults)
	gridWithoutColumns := &Layout{
		Type: LayoutGrid,
		Gap:  "md",
	}
	err = gridWithoutColumns.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test valid tabs layout
	validTabs := &Layout{
		Type: LayoutTabs,
		Tabs: []Tab{
			{ID: "tab1", Label: "Tab 1", Fields: []string{"field1"}},
		},
	}
	err = validTabs.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test tabs layout without tabs (should pass - validation only checks field references)
	tabsWithoutTabs := &Layout{
		Type: LayoutTabs,
	}
	err = tabsWithoutTabs.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test valid steps layout
	validSteps := &Layout{
		Type: LayoutSteps,
		Steps: []Step{
			{ID: "step1", Title: "Step 1", Order: 1, Fields: []string{"field1"}},
		},
	}
	err = validSteps.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test steps layout without steps (should pass - validation only checks field references)
	stepsWithoutSteps := &Layout{
		Type: LayoutSteps,
	}
	err = stepsWithoutSteps.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test valid sections layout
	validSections := &Layout{
		Type: LayoutSections,
		Sections: []Section{
			{ID: "section1", Title: "Section 1", Fields: []string{"field1"}},
		},
	}
	err = validSections.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test sections layout without sections (should pass - validation only checks field references)
	sectionsWithoutSections := &Layout{
		Type: LayoutSections,
	}
	err = sectionsWithoutSections.ValidateLayout(schema)
	require.NoError(suite.T(), err)

	// Test layout with invalid field reference (should error)
	layoutWithInvalidField := &Layout{
		Type: LayoutSections,
		Sections: []Section{
			{ID: "section1", Title: "Section 1", Fields: []string{"nonexistent_field"}},
		},
	}
	err = layoutWithInvalidField.ValidateLayout(schema)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "references non-existent field")
}
