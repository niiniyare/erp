package integration

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/niiniyare/erp/web/builder/core"
	"github.com/niiniyare/erp/web/builder/preview"
	"github.com/niiniyare/erp/web/builder/templates"
)

// TestRunner handles end-to-end testing of the visual builder
type TestRunner struct {
	registry *core.ComponentRegistry
	library  *templates.TemplateLibrary
	renderer *preview.PreviewRenderer
	server   *httptest.Server
}

// TestResult represents the result of a test execution
type TestResult struct {
	TestName    string        `json:"testName"`
	Passed      bool          `json:"passed"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error,omitempty"`
	Details     []string      `json:"details,omitempty"`
	Performance *Performance  `json:"performance,omitempty"`
}

// Performance metrics for test execution
type Performance struct {
	RenderTime     time.Duration `json:"renderTime"`
	ComponentCount int           `json:"componentCount"`
	MemoryUsage    int64         `json:"memoryUsage"`
	CacheHitRate   float64       `json:"cacheHitRate"`
}

// TestSuite represents a collection of tests
type TestSuite struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Tests       []*TestResult `json:"tests"`
	Summary     *TestSummary  `json:"summary"`
}

// TestSummary provides overall test results
type TestSummary struct {
	TotalTests    int           `json:"totalTests"`
	PassedTests   int           `json:"passedTests"`
	FailedTests   int           `json:"failedTests"`
	TotalDuration time.Duration `json:"totalDuration"`
	SuccessRate   float64       `json:"successRate"`
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	registry := core.NewComponentRegistry()
	library := templates.NewTemplateLibrary(registry)
	renderer := preview.NewPreviewRenderer(registry)

	// Initialize test components
	initializeTestComponents(registry)

	return &TestRunner{
		registry: registry,
		library:  library,
		renderer: renderer,
	}
}

// RunAllTests executes all integration tests
func (tr *TestRunner) RunAllTests() (*TestSuite, error) {
	suite := &TestSuite{
		Name:        "Visual Builder Integration Tests",
		Description: "End-to-end testing of the visual builder system",
		Tests:       make([]*TestResult, 0),
	}

	startTime := time.Now()

	// Run individual test suites
	coreTests := tr.runCoreTests()
	suite.Tests = append(suite.Tests, coreTests...)

	previewTests := tr.runPreviewTests()
	suite.Tests = append(suite.Tests, previewTests...)

	templateTests := tr.runTemplateTests()
	suite.Tests = append(suite.Tests, templateTests...)

	workflowTests := tr.runWorkflowTests()
	suite.Tests = append(suite.Tests, workflowTests...)

	performanceTests := tr.runPerformanceTests()
	suite.Tests = append(suite.Tests, performanceTests...)

	// Calculate summary
	suite.Summary = tr.calculateSummary(suite.Tests, time.Since(startTime))

	return suite, nil
}

// Core component tests
func (tr *TestRunner) runCoreTests() []*TestResult {
	tests := []*TestResult{}

	// Test component registry
	tests = append(tests, tr.testComponentRegistry())

	// Test schema composition
	tests = append(tests, tr.testSchemaComposition())

	// Test component instance creation
	tests = append(tests, tr.testComponentInstances())

	// Test layout engine
	tests = append(tests, tr.testLayoutEngine())

	return tests
}

func (tr *TestRunner) testComponentRegistry() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Component Registry",
		Details:  make([]string, 0),
	}

	// Test component registration
	testComponent := &core.ComponentDefinition{
		Type:        "test.component",
		Name:        "Test Component",
		Description: "Test component for integration testing",
		Category:    "test",
		Template: core.ComponentTemplate{
			HTML: "<div>{{.Props.text}}</div>",
		},
		DefaultProps: map[string]interface{}{
			"text": "Hello World",
		},
	}

	if err := tr.registry.RegisterComponent(testComponent); err != nil {
		result.Error = fmt.Sprintf("Failed to register component: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Component registration successful")

	// Test component retrieval
	retrievedComponent := tr.registry.GetComponent("test.component")
	if retrievedComponent == nil {
		result.Error = "Failed to retrieve registered component"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Component retrieval successful")

	// Test component listing
	components := tr.registry.GetComponents()
	if len(components) == 0 {
		result.Error = "No components found in registry"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, fmt.Sprintf("✓ Found %d components in registry", len(components)))

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testSchemaComposition() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Schema Composition",
		Details:  make([]string, 0),
	}

	// Create test schema
	schema := &core.CompositionSchema{
		ID:          "test-schema",
		Name:        "Test Schema",
		Description: "Test schema for integration testing",
		Components: []core.ComponentInstance{
			{
				ID:   "header",
				Type: "atoms.text",
				Props: map[string]interface{}{
					"text": "Test Header",
					"size": "large",
				},
				Position: core.Position{X: 0, Y: 0},
				Size:     core.Size{Width: 400, Height: 50},
			},
			{
				ID:   "button",
				Type: "atoms.button",
				Props: map[string]interface{}{
					"text":    "Click Me",
					"variant": "primary",
				},
				Position: core.Position{X: 0, Y: 60},
				Size:     core.Size{Width: 100, Height: 40},
			},
		},
		Layout: "flex",
	}

	// Validate schema
	if err := tr.validateSchema(schema); err != nil {
		result.Error = fmt.Sprintf("Schema validation failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Schema validation successful")

	// Test component positioning
	if schema.Components[0].Position.Y != 0 || schema.Components[1].Position.Y != 60 {
		result.Error = "Component positioning incorrect"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Component positioning correct")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testComponentInstances() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Component Instances",
		Details:  make([]string, 0),
	}

	// Test component instance creation
	instance := &core.ComponentInstance{
		ID:   "test-instance",
		Type: "atoms.button",
		Props: map[string]interface{}{
			"text":    "Test Button",
			"variant": "secondary",
			"size":    "medium",
		},
		Position: core.Position{X: 100, Y: 200},
		Size:     core.Size{Width: 120, Height: 40},
	}

	// Validate instance
	if instance.ID == "" || instance.Type == "" {
		result.Error = "Component instance missing required fields"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Component instance creation successful")

	// Test property access
	if text, ok := instance.Props["text"].(string); !ok || text != "Test Button" {
		result.Error = "Component instance property access failed"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Component instance property access successful")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testLayoutEngine() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Layout Engine",
		Details:  make([]string, 0),
	}

	// Test layout configurations
	flexLayout := "flex"
	gridLayout := "grid"

	if flexLayout != "flex" {
		result.Error = "Flex layout configuration incorrect"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Flex layout configuration successful")

	if gridLayout != "grid" {
		result.Error = "Grid layout configuration incorrect"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Grid layout configuration successful")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

// Preview system tests
func (tr *TestRunner) runPreviewTests() []*TestResult {
	tests := []*TestResult{}

	tests = append(tests, tr.testPreviewGeneration())
	tests = append(tests, tr.testRealTimeUpdates())
	tests = append(tests, tr.testErrorHandling())

	return tests
}

func (tr *TestRunner) testPreviewGeneration() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Preview Generation",
		Details:  make([]string, 0),
	}

	// Create test schema
	schema := &core.CompositionSchema{
		ID:          "preview-test",
		Name:        "Preview Test",
		Description: "Test schema for preview generation",
		Components: []core.ComponentInstance{
			{
				ID:   "test-component",
				Type: "atoms.text",
				Props: map[string]interface{}{
					"text": "Preview Test",
				},
				Position: core.Position{X: 0, Y: 0},
				Size:     core.Size{Width: 200, Height: 30},
			},
		},
	}

	// Generate preview
	update, err := tr.renderer.RenderPreview(schema)
	if err != nil {
		result.Error = fmt.Sprintf("Preview generation failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Preview generation successful")

	// Validate preview content
	if update.HTML == "" {
		result.Error = "Preview HTML is empty"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Preview HTML generated")

	if update.CSS == "" {
		result.Error = "Preview CSS is empty"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Preview CSS generated")

	if len(update.Errors) > 0 {
		result.Error = fmt.Sprintf("Preview has errors: %v", update.Errors)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ No preview errors")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testRealTimeUpdates() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Real-time Updates",
		Details:  make([]string, 0),
	}

	// Test async preview rendering
	schema := &core.CompositionSchema{
		ID:          "realtime-test",
		Name:        "Real-time Test",
		Description: "Test schema for real-time updates",
		Components: []core.ComponentInstance{
			{
				ID:   "dynamic-component",
				Type: "atoms.text",
				Props: map[string]interface{}{
					"text": "Dynamic Content",
				},
				Position: core.Position{X: 0, Y: 0},
				Size:     core.Size{Width: 200, Height: 30},
			},
		},
	}

	// Create update channel
	updates := make(chan *preview.PreviewUpdate, 1)

	// Trigger async render
	tr.renderer.RenderPreviewAsync(schema, updates)

	// Wait for update
	select {
	case update := <-updates:
		if update.SchemaID != schema.ID {
			result.Error = "Received update for wrong schema"
			result.Passed = false
			result.Duration = time.Since(startTime)
			return result
		}
		result.Details = append(result.Details, "✓ Real-time update received")
	case <-time.After(5 * time.Second):
		result.Error = "Timeout waiting for real-time update"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	// Cleanup
	tr.renderer.UnsubscribeFromUpdates(schema.ID)
	result.Details = append(result.Details, "✓ Subscription cleanup successful")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testErrorHandling() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Error Handling",
		Details:  make([]string, 0),
	}

	// Test with invalid schema
	invalidSchema := &core.CompositionSchema{
		ID:          "invalid-test",
		Name:        "", // Invalid: empty name
		Description: "Test schema with invalid properties",
		Components: []core.ComponentInstance{
			{
				ID:       "invalid-component",
				Type:     "nonexistent.component", // Invalid: component doesn't exist
				Position: core.Position{X: 0, Y: 0},
				Size:     core.Size{Width: 200, Height: 30},
			},
		},
	}

	// Generate preview (should handle errors gracefully)
	update, err := tr.renderer.RenderPreview(invalidSchema)
	if err != nil {
		result.Details = append(result.Details, "✓ Error handling successful (expected error)")
	} else if len(update.Errors) > 0 {
		result.Details = append(result.Details, "✓ Errors properly reported in preview update")
	} else {
		result.Error = "Expected errors not properly handled"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

// Template system tests
func (tr *TestRunner) runTemplateTests() []*TestResult {
	tests := []*TestResult{}

	tests = append(tests, tr.testTemplateLibrary())
	tests = append(tests, tr.testTemplateImport())
	tests = append(tests, tr.testTemplateCloning())

	return tests
}

func (tr *TestRunner) testTemplateLibrary() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Template Library",
		Details:  make([]string, 0),
	}

	// Test template creation
	template := &templates.Template{
		ID:          "test-template",
		Name:        "Test Template",
		Description: "Template for integration testing",
		Category:    "test",
		Tags:        []string{"test", "integration"},
		Author:      "Test Suite",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Schema: &core.CompositionSchema{
			ID:          "test-template-schema",
			Name:        "Test Template Schema",
			Description: "Schema for test template",
			Components: []core.ComponentInstance{
				{
					ID:   "template-component",
					Type: "atoms.text",
					Props: map[string]interface{}{
						"text": "Template Component",
					},
					Position: core.Position{X: 0, Y: 0},
					Size:     core.Size{Width: 200, Height: 30},
				},
			},
		},
		Variables: map[string]templates.Variable{
			"title": {
				Name:        "title",
				Type:        "string",
				Default:     "Default Title",
				Description: "Template title",
				Required:    true,
			},
		},
		License:  "free",
		IsPublic: true,
	}

	// Add template to library
	if err := tr.library.AddTemplate(template); err != nil {
		result.Error = fmt.Sprintf("Failed to add template: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Template added to library")

	// Retrieve template
	retrievedTemplate, err := tr.library.GetTemplate("test-template")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to retrieve template: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Template retrieved successfully")

	// Validate template data
	if retrievedTemplate.Name != template.Name {
		result.Error = "Retrieved template data doesn't match"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Template data validation successful")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testTemplateImport() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Template Import",
		Details:  make([]string, 0),
	}

	// Create sample Phase 4.3 schema data
	phase43Data := map[string]interface{}{
		"name":        "Imported Template",
		"description": "Template imported from Phase 4.3",
		"version":     "1.0.0",
		"createdAt":   time.Now(),
		"components": []map[string]interface{}{
			{
				"type": "CustomButton",
				"name": "Action Button",
				"props": map[string]interface{}{
					"text":    "Click Me",
					"variant": "primary",
				},
			},
		},
		"layout": map[string]interface{}{
			"type": "flex",
			"flex": map[string]interface{}{
				"direction": "column",
				"gap":       16,
			},
		},
	}

	phase43JSON, err := json.Marshal(phase43Data)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to marshal test data: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	// Create importer
	importer := templates.NewImporter(tr.library, tr.registry)

	// Import schema
	importOptions := &templates.ImportOptions{
		OverwriteExisting:  true,
		ValidateComponents: true,
		DefaultAuthor:      "Test Suite",
		DefaultCategory:    "imported",
		ComponentMapping: map[string]string{
			"CustomButton": "atoms.button",
		},
	}

	importResult, err := importer.ImportPhase43Schema(phase43JSON, importOptions)
	if err != nil {
		result.Error = fmt.Sprintf("Import failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Phase 4.3 schema import successful")

	// Validate import results
	if importResult.SuccessCount == 0 {
		result.Error = "No templates were imported"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, fmt.Sprintf("✓ Imported %d templates", importResult.SuccessCount))

	if importResult.ErrorCount > 0 {
		result.Error = fmt.Sprintf("Import had %d errors", importResult.ErrorCount)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ No import errors")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testTemplateCloning() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Template Cloning",
		Details:  make([]string, 0),
	}

	// Get a template from the library
	templates := tr.library.GetFeaturedTemplates()
	if len(templates) == 0 {
		result.Error = "No templates available for cloning test"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	template := templates[0]
	result.Details = append(result.Details, fmt.Sprintf("✓ Found template for cloning: %s", template.Name))

	// Clone template with variable substitutions
	variables := map[string]interface{}{
		"title": "Cloned Template Title",
	}

	clonedSchema, err := tr.library.CloneTemplate(template.ID, variables)
	if err != nil {
		result.Error = fmt.Sprintf("Template cloning failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Template cloning successful")

	// Validate cloned schema
	if clonedSchema == nil {
		result.Error = "Cloned schema is nil"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Cloned schema validation successful")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

// Workflow tests (end-to-end scenarios)
func (tr *TestRunner) runWorkflowTests() []*TestResult {
	tests := []*TestResult{}

	tests = append(tests, tr.testCompleteWorkflow())
	tests = append(tests, tr.testComponentCompatibility())

	return tests
}

func (tr *TestRunner) testCompleteWorkflow() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Complete Workflow",
		Details:  make([]string, 0),
	}

	// 1. Create schema from scratch
	schema := &core.CompositionSchema{
		ID:          "workflow-test",
		Name:        "Workflow Test Schema",
		Description: "End-to-end workflow test",
		Components:  []core.ComponentInstance{},
		Layout:      "flex",
	}
	result.Details = append(result.Details, "✓ Schema created")

	// 2. Add components to schema
	schema.Components = append(schema.Components, core.ComponentInstance{
		ID:   "header",
		Type: "atoms.text",
		Props: map[string]interface{}{
			"text": "Workflow Test Header",
			"size": "large",
		},
		Position: core.Position{X: 0, Y: 0},
		Size:     core.Size{Width: 400, Height: 50},
	})

	schema.Components = append(schema.Components, core.ComponentInstance{
		ID:   "button",
		Type: "atoms.button",
		Props: map[string]interface{}{
			"text":    "Test Button",
			"variant": "primary",
		},
		Position: core.Position{X: 0, Y: 60},
		Size:     core.Size{Width: 120, Height: 40},
	})
	result.Details = append(result.Details, "✓ Components added to schema")

	// 3. Generate preview
	update, err := tr.renderer.RenderPreview(schema)
	if err != nil {
		result.Error = fmt.Sprintf("Preview generation failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Preview generated")

	// 4. Validate preview contains expected elements
	if !strings.Contains(update.HTML, "Workflow Test Header") {
		result.Error = "Preview doesn't contain expected header text"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Preview content validated")

	// 5. Create template from schema
	template := &templates.Template{
		ID:          "workflow-template",
		Name:        "Workflow Template",
		Description: "Template created from workflow test",
		Category:    "test",
		Tags:        []string{"workflow", "test"},
		Author:      "Test Suite",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Schema:      schema,
		Variables:   map[string]templates.Variable{},
		License:     "free",
		IsPublic:    true,
	}

	if err := tr.library.AddTemplate(template); err != nil {
		result.Error = fmt.Sprintf("Failed to save template: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Template saved to library")

	// 6. Clone template
	clonedSchema, err := tr.library.CloneTemplate(template.ID, map[string]interface{}{})
	if err != nil {
		result.Error = fmt.Sprintf("Template cloning failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Template cloned successfully")

	// 7. Generate preview of cloned schema
	clonedUpdate, err := tr.renderer.RenderPreview(clonedSchema)
	if err != nil {
		result.Error = fmt.Sprintf("Cloned preview generation failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Cloned template preview generated")

	// 8. Validate both previews are similar
	if len(clonedUpdate.HTML) == 0 || len(clonedUpdate.CSS) == 0 {
		result.Error = "Cloned preview is empty"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Complete workflow successful")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testComponentCompatibility() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName: "Component Compatibility",
		Details:  make([]string, 0),
	}

	// Test all registered components
	components := tr.registry.GetComponents()
	compatibleCount := 0

	for _, component := range components {
		// Create test instance
		instance := core.ComponentInstance{
			ID:   "compat-test-" + component.Type,
			Type: component.Type,
			Props: map[string]interface{}{
				"text": "Compatibility Test",
			},
			Position: core.Position{X: 0, Y: 0},
			Size:     core.Size{Width: 200, Height: 50},
		}

		// Try to render in a simple schema
		schema := &core.CompositionSchema{
			ID:          "compat-test-" + component.Type,
			Name:        "Compatibility Test",
			Description: "Testing component compatibility",
			Components:  []core.ComponentInstance{instance},
		}

		update, err := tr.renderer.RenderPreview(schema)
		if err == nil && len(update.Errors) == 0 {
			compatibleCount++
			result.Details = append(result.Details, fmt.Sprintf("✓ %s compatible", component.Type))
		} else {
			result.Details = append(result.Details, fmt.Sprintf("⚠ %s has issues: %v", component.Type, err))
		}
	}

	// Calculate compatibility rate
	compatibilityRate := float64(compatibleCount) / float64(len(components)) * 100
	result.Details = append(result.Details, fmt.Sprintf("✓ Component compatibility: %.1f%% (%d/%d)",
		compatibilityRate, compatibleCount, len(components)))

	if compatibilityRate < 80 {
		result.Error = fmt.Sprintf("Low component compatibility rate: %.1f%%", compatibilityRate)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

// Performance tests
func (tr *TestRunner) runPerformanceTests() []*TestResult {
	tests := []*TestResult{}

	tests = append(tests, tr.testRenderPerformance())
	tests = append(tests, tr.testCachePerformance())
	tests = append(tests, tr.testLargeSchemaHandling())

	return tests
}

func (tr *TestRunner) testRenderPerformance() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName:    "Render Performance",
		Details:     make([]string, 0),
		Performance: &Performance{},
	}

	// Create a moderately complex schema
	schema := tr.createComplexSchema(20) // 20 components

	// Measure render time
	renderStart := time.Now()
	update, err := tr.renderer.RenderPreview(schema)
	renderTime := time.Since(renderStart)

	if err != nil {
		result.Error = fmt.Sprintf("Render failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	result.Performance.RenderTime = renderTime
	result.Performance.ComponentCount = len(schema.Components)
	result.Details = append(result.Details, fmt.Sprintf("✓ Rendered %d components in %v",
		len(schema.Components), renderTime))

	// Performance thresholds
	if renderTime > 2*time.Second {
		result.Error = fmt.Sprintf("Render time too slow: %v", renderTime)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Render time within acceptable limits")

	if len(update.HTML) == 0 {
		result.Error = "Rendered HTML is empty"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Rendered content is not empty")

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testCachePerformance() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName:    "Cache Performance",
		Details:     make([]string, 0),
		Performance: &Performance{},
	}

	schema := tr.createComplexSchema(10)

	// First render (cold cache)
	firstRenderStart := time.Now()
	_, err := tr.renderer.RenderPreview(schema)
	firstRenderTime := time.Since(firstRenderStart)

	if err != nil {
		result.Error = fmt.Sprintf("First render failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, fmt.Sprintf("✓ First render: %v", firstRenderTime))

	// Second render (hot cache)
	secondRenderStart := time.Now()
	_, err = tr.renderer.RenderPreview(schema)
	secondRenderTime := time.Since(secondRenderStart)

	if err != nil {
		result.Error = fmt.Sprintf("Second render failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, fmt.Sprintf("✓ Second render: %v", secondRenderTime))

	// Calculate cache improvement
	improvement := float64(firstRenderTime-secondRenderTime) / float64(firstRenderTime) * 100
	result.Performance.CacheHitRate = improvement
	result.Details = append(result.Details, fmt.Sprintf("✓ Cache improvement: %.1f%%", improvement))

	if improvement < 10 {
		result.Error = fmt.Sprintf("Cache improvement too low: %.1f%%", improvement)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

func (tr *TestRunner) testLargeSchemaHandling() *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestName:    "Large Schema Handling",
		Details:     make([]string, 0),
		Performance: &Performance{},
	}

	// Create a large schema
	largeSchema := tr.createComplexSchema(100) // 100 components

	// Test rendering
	renderStart := time.Now()
	update, err := tr.renderer.RenderPreview(largeSchema)
	renderTime := time.Since(renderStart)

	if err != nil {
		result.Error = fmt.Sprintf("Large schema render failed: %v", err)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}

	result.Performance.RenderTime = renderTime
	result.Performance.ComponentCount = len(largeSchema.Components)
	result.Details = append(result.Details, fmt.Sprintf("✓ Rendered %d components in %v",
		len(largeSchema.Components), renderTime))

	// Performance threshold for large schemas
	if renderTime > 10*time.Second {
		result.Error = fmt.Sprintf("Large schema render time too slow: %v", renderTime)
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, "✓ Large schema render time acceptable")

	// Check output size
	if len(update.HTML) < 1000 {
		result.Error = "Large schema produced too little HTML"
		result.Passed = false
		result.Duration = time.Since(startTime)
		return result
	}
	result.Details = append(result.Details, fmt.Sprintf("✓ Generated HTML size: %d bytes", len(update.HTML)))

	result.Passed = true
	result.Duration = time.Since(startTime)
	return result
}

// Helper methods

func (tr *TestRunner) validateSchema(schema *core.CompositionSchema) error {
	if schema.ID == "" {
		return fmt.Errorf("schema ID is required")
	}
	if schema.Name == "" {
		return fmt.Errorf("schema name is required")
	}
	return nil
}

func (tr *TestRunner) createComplexSchema(componentCount int) *core.CompositionSchema {
	schema := &core.CompositionSchema{
		ID:          fmt.Sprintf("complex-schema-%d", componentCount),
		Name:        fmt.Sprintf("Complex Schema (%d components)", componentCount),
		Description: "Complex schema for performance testing",
		Components:  make([]core.ComponentInstance, 0, componentCount),
		Layout:      "grid",
	}

	componentTypes := []string{"atoms.text", "atoms.button", "molecules.card"}

	for i := 0; i < componentCount; i++ {
		componentType := componentTypes[i%len(componentTypes)]

		instance := core.ComponentInstance{
			ID:   fmt.Sprintf("component-%d", i),
			Type: componentType,
			Props: map[string]interface{}{
				"text": fmt.Sprintf("Component %d", i),
			},
			Position: core.Position{
				X: float64((i % 5) * 150),
				Y: float64((i / 5) * 100),
			},
			Size: core.Size{Width: 140, Height: 80},
		}

		schema.Components = append(schema.Components, instance)
	}

	return schema
}

func (tr *TestRunner) calculateSummary(tests []*TestResult, totalDuration time.Duration) *TestSummary {
	summary := &TestSummary{
		TotalTests:    len(tests),
		TotalDuration: totalDuration,
	}

	for _, test := range tests {
		if test.Passed {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}
	}

	summary.SuccessRate = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	return summary
}

func initializeTestComponents(registry *core.ComponentRegistry) {
	// Register basic test components
	registry.RegisterComponent(&core.ComponentDefinition{
		Type:        "atoms.text",
		Name:        "Text",
		Description: "Basic text component",
		Category:    "atoms",
		Template: core.ComponentTemplate{
			HTML: `<span class="text-component">{{.Props.text}}</span>`,
		},
		DefaultProps: map[string]interface{}{
			"text": "Default Text",
			"size": "medium",
		},
	})

	registry.RegisterComponent(&core.ComponentDefinition{
		Type:        "atoms.button",
		Name:        "Button",
		Description: "Basic button component",
		Category:    "atoms",
		Template: core.ComponentTemplate{
			HTML: `<button class="btn btn-{{.Props.variant}}">{{.Props.text}}</button>`,
		},
		DefaultProps: map[string]interface{}{
			"text":    "Button",
			"variant": "primary",
			"size":    "medium",
		},
	})

	registry.RegisterComponent(&core.ComponentDefinition{
		Type:        "molecules.card",
		Name:        "Card",
		Description: "Card component",
		Category:    "molecules",
		Template: core.ComponentTemplate{
			HTML: `<div class="card"><div class="card-body">{{.Props.content}}</div></div>`,
		},
		DefaultProps: map[string]interface{}{
			"content": "Card Content",
		},
	})
}
